// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"time"

	"github.com/msiner/sdrplay-go/api"
	"github.com/msiner/sdrplay-go/helpers/callback"
	"github.com/msiner/sdrplay-go/helpers/duo"
	"github.com/msiner/sdrplay-go/helpers/event"
	"github.com/msiner/sdrplay-go/helpers/parse"
	"github.com/msiner/sdrplay-go/session"
)

func duowav() error {
	flags := flag.NewFlagSet("duocorr", flag.ExitOnError)
	flags.Usage = func() {
		fmt.Fprintln(flags.Output(), strings.TrimSpace(`
Usage: duocorr [FLAGS] <tuneHz> <duration>

duocorr connects to an available RSPduo device, configures it
in dual-tuner mode with both tuners having the same configuration.
It then continuously calculates a correlation between channel A and
channel B for the specified time duration.

Correlation results will be printed to standard out. The "-inter"
flag determines the interval for printing results. The results
reported at the end of each interval period reflect all measurements
taken during that period.

Since the dual tuners of the RSPduo are not phase-coherent, duocorr
performs a correlation of the magnitudes of samples. This is reported
as the offset of the correlation peak in the number of samples plus
or minus from the center. The z-score of the correlation peak is also
computed and reported.

duocorr is a very rudimentary tool included for testing purposes. It
can serve the practical application of measuring or detecting a time
offset in an antenna and cabling setup that might make dual-antenna
diversity processing less reliable. Also, duocorr is not optimized.
This was intentional to be able to include this tool in the repository
without introducing a module dependency on a third-party DSP library.

Arguments:
  tuneHz
	Tuner RF frequency in Hz is a mandatory argument. It can
	be specified with k, K, m, M, g, or G suffix to indicate the
	value is in kHz, MHz, or GHz respectively (e.g. 1.42G).
  duration
	Amount of time to process correlations before exiting. It can be
	specified with a s, S, m, M, h, or H suffix to indicate the value
	is in seconds, minutes, or hours respectively (e.g. 10s). Without
	a suffix provided, the value is interpreted as seconds.

Flags:
`,
		))
		flags.PrintDefaults()
	}
	lnaOpt := flags.String("lna", "50%", parse.LNAFlagHelp)
	decOpt := flags.Uint("dec", 1, parse.DecFlagHelp)
	warmOpt := flags.Uint("warm", 2, parse.WarmFlagHelp)
	agcCtlOpt := flags.String("agcctl", "enable", parse.AGCCtlFlagHelp)
	agcSetOpt := flags.Int("agcset", -30, parse.AGCSetFlagHelp)
	serialsOpt := flags.String("serials", "any", parse.SerialsFlagHelp)
	usbOpt := flags.String("usb", "isoch", parse.USBFlagHelp)
	hizOpt := flags.Bool("hiz", false, parse.HiZFlagHelp)
	maxFsOpt := flags.Bool("maxfs", false, strings.TrimSpace(`
Use the maximum 8MHz sample rate.
This will deliver 12-bit ADC resolution, but with slightly better
anti-aliasing performance at the widest bandwidth. This mode is
only available at 1.536 MHz analog bandwidth. The default mode is
to use a 6 MHz ADC sample rate. That mode delivers 14-bit ADC
resolution, but with slightly inferior anti-aliasing performance
at the widest bandwidth. The default mode is also compatible with
analog bandwidths of 1.536 MHz, 600 kHz, 300 kHz, and 200 kHz.
6 MHz operation should result in a slightly lower CPU load.`,
	))
	interOpt := flags.String("inter", "1s", "measurement reporting interval")

	// Using ExitOnError
	_ = flags.Parse(os.Args[1:])

	switch flags.NArg() {
	case 0:
		flags.Usage()
		return errors.New("missing frequency and number of bytes")
	case 1:
		flags.Usage()
		return errors.New("missing number of bytes")
	case 2:
		// good
	default:
		flags.Usage()
		return errors.New("too many arguments")
	}

	lg := log.New(os.Stdout, "", log.LstdFlags)

	freq, err := parse.ParseTuneFrequency(flags.Arg(0))
	if err != nil {
		return err
	}
	dur, err := ParseDuration(flags.Arg(1))
	if err != nil {
		return err
	}

	dec, err := parse.CheckDecFlag(*decOpt)
	if err != nil {
		return err
	}

	warm, err := parse.ParseWarmFlag(*warmOpt)
	if err != nil {
		return err
	}

	agcCtl, err := parse.ParseAGCCtlFlag(*agcCtlOpt)
	if err != nil {
		return err
	}

	agcSet, err := parse.ParseAGCSetFlag(*agcSetOpt)
	if err != nil {
		return err
	}

	usb, err := parse.ParseUSBFlag(*usbOpt)
	if err != nil {
		return err
	}

	lnaState, lnaPct, err := parse.CheckLNAFlag(*lnaOpt)
	lnaCfg := session.NoopChanConfig
	switch {
	case err != nil:
		return err
	case lnaState != nil:
		lnaCfg = session.WithLNAState(*lnaState)
	case lnaPct != nil:
		lnaCfg = session.WithLNAPercent(*lnaPct)
	}

	serials, err := parse.ParseSerialsFlag(*serialsOpt)
	if err != nil {
		return err
	}

	var serialsFilter session.DevFilterFn
	switch serials {
	case nil:
		serialsFilter = session.WithNoopDevFilter()
	default:
		serialsFilter = session.WithSerials(serials...)
	}

	inter, err := ParseDuration(*interOpt)
	if err != nil {
		return err
	}

	// Setup callback and control state.
	const cbSamples = 10000
	corr := NewCorrelationFn(cbSamples)
	peak := IntAccumulator{}
	zscore := FloatAccumulator{}
	detectDropsA := callback.NewDropDetectFn()
	detectDropsB := callback.NewDropDetectFn()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)
		v, ok := <-sig
		if ok {
			lg.Printf("signal; got %v", v)
			cancel()
		}
	}()

	var isWarm uint32
	interTkr := time.NewTicker(inter)
	go func() {
		time.Sleep(warm)
		lg.Println("warm-up complete")
		interTkr.Reset(inter)
		atomic.StoreUint32(&isWarm, 1)
		tmr := time.NewTimer(dur)
		select {
		case <-ctx.Done():
		case <-tmr.C:
			cancel()
		}
	}()

	evtChan := event.NewEventChan(10)
	defer evtChan.Close()

	syncChan := duo.NewSynchroChan(1000)

	synchro := duo.NewSynchro(
		cbSamples,
		syncChan.Callback,
		func(evt duo.SynchroEvent, msg string) {
			lg.Printf("%s: %s\n", evt, msg)
		},
	)

	var (
		msgNum      uint64
		msgNumValid bool
	)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-syncChan.C:
				switch msgNumValid {
				case true:
					diff := msg.MsgNum - msgNum
					if diff != 1 {
						lg.Printf("dropped %d messages from synchro", diff)
					}
				default:
					msgNumValid = true
				}
				msgNum = msg.MsgNum

				res := corr(&msg)
				peak.Add(res.Peak)
				zscore.Add(res.ZScore)

				select {
				case <-interTkr.C:
					pstats := peak.Analyze()
					peak.Reset()
					zstats := zscore.Analyze()
					zscore.Reset()
					lg.Printf("Report: measurements=%d\n", pstats.Count)
					lg.Printf(
						"Report: peak_offset(mean=%0.2f stddev=%0.2f median=%d min=%d max=%d)\n",
						pstats.Mean, pstats.StdDev, pstats.Median, pstats.Min, pstats.Max,
					)
					lg.Printf(
						"Report: peak_zscore(mean=%0.2f stddev=%0.2f median=%0.2f min=%0.2f max=%0.2f)\n",
						zstats.Mean, zstats.StdDev, zstats.Median, zstats.Min, zstats.Max,
					)
				default:
				}
			}
		}
	}()

	err = session.Run(
		ctx,
		session.WithSelector(
			serialsFilter,
			session.WithRSPduo(),
			session.WithDuoTunerBoth(),
			session.WithDuoModeDual(*maxFsOpt),
		),
		session.WithDebug(true),
		session.WithDeviceConfig(
			func(d *api.DeviceT, p *api.DeviceParamsT) error {
				lg.Printf("Device: %v,%v,%v,%v,%v\n", d.HWVer, d.SerNo, d.Tuner, d.RspDuoMode, d.RspDuoSampleFreq)
				return nil
			},
			session.WithTransferMode(usb),
			session.WithHighZPortEnabled(*hizOpt),
			session.WithDuoChannelAConfig(
				session.WithLowIF(session.LowIFMaxBits, dec),
				session.WithTuneFreq(freq),
				lnaCfg,
				session.WithAGC(agcCtl, agcSet),
				session.WithLogChannelParams("Tuner_A", lg),
			),
			session.WithDuoChannelBConfig(
				session.WithLowIF(session.LowIFMaxBits, dec),
				session.WithTuneFreq(freq),
				lnaCfg,
				session.WithAGC(agcCtl, agcSet),
				session.WithLogChannelParams("Tuner_B", lg),
			),
		),
		session.WithStreamACallback(func(xi, xq []int16, params *api.StreamCbParamsT, reset bool) {
			select {
			case <-ctx.Done():
				return
			default:
			}

			if atomic.LoadUint32(&isWarm) == 0 {
				return
			}

			d := detectDropsA(params, reset)
			if d != 0 {
				lg.Printf(
					"stream A: dropped %d samples; firstSampleNum=%d numSamples=%d reset=%v\n",
					d, params.FirstSampleNum, params.NumSamples, reset,
				)
				reset = true
			}
			synchro.StreamACallback(xi, xq, params, reset)
		}),
		session.WithStreamBCallback(func(xi, xq []int16, params *api.StreamCbParamsT, reset bool) {
			select {
			case <-ctx.Done():
				return
			default:
			}

			if atomic.LoadUint32(&isWarm) == 0 {
				return
			}

			d := detectDropsB(params, reset)
			if d != 0 {
				lg.Printf(
					"stream B: dropped %d samples; firstSampleNum=%d numSamples=%d reset=%v\n",
					d, params.FirstSampleNum, params.NumSamples, reset,
				)
				reset = true
			}
			synchro.StreamBCallback(xi, xq, params, reset)
		}),
		session.WithEventCallback(evtChan.Callback),
		session.WithControlLoop(func(_ context.Context, d *api.DeviceT, a api.API) error {
			for {
				select {
				case <-ctx.Done():
					return nil
				case evt := <-evtChan.C:
					event.LogEventMsg(evt, lg)
					err := event.HandlePowerOverloadChangeMsg(d, a, evt, lg, true)
					if err != nil {
						lg.Printf("failed to handle power overload event; %v", err)
						cancel()
					}
				}
			}
		}),
	)
	switch err {
	case nil, context.Canceled:
		lg.Println("clean exit")
	default:
		return fmt.Errorf("error during session run; %v", err)
	}

	return nil
}

func main() {
	err := duowav()
	if err != nil {
		log.Fatal(err)
	}
}
