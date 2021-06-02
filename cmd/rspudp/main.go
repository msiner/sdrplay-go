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
	"net"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"time"

	"github.com/msiner/sdrplay-go/api"
	"github.com/msiner/sdrplay-go/helpers/callback"
	"github.com/msiner/sdrplay-go/helpers/parse"
	"github.com/msiner/sdrplay-go/helpers/udp"
	"github.com/msiner/sdrplay-go/session"
)

func rspudp() error {
	flags := flag.NewFlagSet("rspwav", flag.ExitOnError)
	flags.Usage = func() {
		fmt.Fprintln(flags.Output(), strings.TrimSpace(`
Usage: rspudp [FLAGS] <tuneHz>

rspwav connects to an available RSP device, configures it, and sends
the received samples to the specified UDP target.

Arguments:
  tuneHz
	Tuner RF frequency in Hz is a mandatory argument. It can
	be specified with k, K, m, M, g, or G suffix to indicate the
	value is in kHz, MHz, or GHz respectively (e.g. 1.42G).

Flags:
`,
		))
		flags.PrintDefaults()
	}
	lifOpt := flags.Bool("lif", false, strings.TrimSpace(`
Use low-IF mode. In low-IF mode, the effective sample rate, before decimation
is 2 MHz. When -lif is specified, the -fs option cannot be used to configure
the sample rate.`,
	))
	remoteOpt := flags.String("remote", "127.0.0.1:1234", strings.TrimSpace(`
Target host address or name and UDP port`,
	))
	payOpt := flags.Uint("pay", 1400, strings.TrimSpace(`
UDP payload size in bytes. This must be small enough to fit in
the network MTU with IP and UDP headers. It must also be a multiple
of the 4 byte frame size.`,
	))
	seqOpt := flags.Bool("seq", false, strings.TrimSpace(`
Insert a 64-bit sequence number at the beginning of each packet.
This will use 8 bytes of the specified payload size.`,
	))
	lnaOpt := flags.String("lna", "50%", parse.LNAFlagHelp)
	fsOpt := flags.String("fs", "6M", parse.FsFlagHelp)
	decOpt := flags.Uint("dec", 1, parse.DecFlagHelp)
	warmOpt := flags.Uint("warm", 2, parse.WarmFlagHelp)
	agcCtlOpt := flags.String("agcctl", "enable", parse.AGCCtlFlagHelp)
	agcSetOpt := flags.Int("agcset", -30, parse.AGCSetFlagHelp)
	duoTunerOpt := flags.String("duotuner", "either", parse.DuoTunerFlagHelp)
	serialsOpt := flags.String("serials", "any", parse.SerialsFlagHelp)
	usbOpt := flags.String("usb", "isoch", parse.USBFlagHelp)
	hizOpt := flags.Bool("hiz", false, parse.HiZFlagHelp)
	dxAntOpt := flags.String("dxant", "a", parse.DxAntFlagHelp)
	rsp2AntOpt := flags.String("rsp2ant", "a", parse.Rsp2AntFlagHelp)
	bigOpt := flags.Bool("big", false, "Write samples with big-endian byte order")

	flags.Parse(os.Args[1:])

	switch flags.NArg() {
	case 0:
		flags.Usage()
		return errors.New("missing tune frequency")
	case 1:
		// good
	default:
		flags.Usage()
		return errors.New("too many arguments")
	}

	freq, err := parse.ParseTuneFrequency(flags.Arg(0))
	if err != nil {
		return err
	}

	fs, err := parse.ParseFsFlag(*fsOpt)
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

	serials, err := parse.ParseSerialsFlag(*serialsOpt)
	if err != nil {
		return err
	}

	duoTuner, err := parse.ParseDuoTunerFlag(*duoTunerOpt)
	if err != nil {
		return err
	}

	dxAnt, err := parse.ParseDxAntFlag(*dxAntOpt)
	if err != nil {
		return err
	}

	rsp2Ant, err := parse.ParseRsp2AntFlag(*rsp2AntOpt)
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

	// Check the options and get the correct function to configure
	// the IF mode as zero or low.
	var ifModeCfg session.ChanConfigFn
	switch *lifOpt {
	case true:
		ifModeCfg = session.WithLowIF(session.LowIFMaxBits, dec)
	default:
		ifModeCfg = session.WithZeroIF(fs, dec)
	}

	var serialsFilter session.DevFilterFn
	switch serials {
	case nil:
		serialsFilter = session.WithNoopDevFilter()
	default:
		serialsFilter = session.WithSerials(serials...)
	}

	var duoTunerFilter session.DevFilterFn
	switch duoTuner {
	case parse.DuoTunerFlagA:
		duoTunerFilter = session.WithDuoTunerA()
	case parse.DuoTunerFlagB:
		duoTunerFilter = session.WithDuoTunerB()
	default:
		duoTunerFilter = session.WithDuoTunerEither()
	}

	addr, err := net.ResolveUDPAddr("udp", *remoteOpt)
	if err != nil {
		return err
	}

	conn, err := net.DialUDP(addr.Network(), nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	log.Printf("UDP initialized: local=%v remote=%v", conn.LocalAddr(), conn.RemoteAddr())

	if *payOpt%4 != 0 {
		return fmt.Errorf("payload size must be multiple of frame size; got %d", *payOpt)
	}

	log.Printf("Payload Size: %d B", *payOpt)

	// Setup callback and control state.
	write, err := udp.NewPacketWriteFn(*payOpt, 2, *seqOpt, *bigOpt)
	if err != nil {
		return err
	}
	interleave := callback.NewInterleaveFn()
	detectDrops := callback.NewDropDetectFn()
	var isWarm uint32
	go func() {
		time.Sleep(warm)
		log.Println("warm-up complete")
		atomic.StoreUint32(&isWarm, 1)
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)
		v, ok := <-sig
		if ok {
			log.Printf("signal; got %v", v)
			cancel()
		}
	}()

	sess, err := session.NewSession(
		session.WithSelector(
			serialsFilter,
			duoTunerFilter,
			session.WithDuoModeSingle(),
		),
		session.WithDebug(true),
		session.WithDeviceConfig(
			func(d *api.DeviceT, p *api.DeviceParamsT) error {
				switch d.HWVer {
				case api.RSPduo_ID:
					log.Printf("Device: %v,%v,%v,%v,%v\n", d.HWVer, d.SerNo, d.Tuner, d.RspDuoMode, d.RspDuoSampleFreq)
				default:
					log.Printf("Device: %v,%v\n", d.HWVer, d.SerNo)
				}
				return nil
			},
			session.WithTransferMode(usb),
			session.WithHighZPortEnabled(*hizOpt),
			session.WithDxAntennaSelect(dxAnt),
			session.WithRsp2AntennaSelect(rsp2Ant),
			session.WithSingleChannelConfig(
				ifModeCfg,
				session.WithTuneFreq(freq),
				session.WithAGC(agcCtl, agcSet),
				lnaCfg,
				func(d *api.DeviceT, p *api.DeviceParamsT, c *api.RxChannelParamsT) error {
					rate, err := session.GetEffectiveSampleRate(d, p, c)
					if err != nil {
						return err
					}
					pct, err := session.GetLNAPercent(d, p, c)
					if err != nil {
						return err
					}
					log.Printf("IF Mode: %v", c.TunerParams.IfType)
					log.Printf("RF Frequency: %v Hz\n", c.TunerParams.RfFreq.RfHz)
					log.Printf("ADC Sample Rate: %v Hz\n", p.DevParams.FsFreq.FsHz)
					log.Printf("Effective Sample Rate: %v Hz\n", rate)
					log.Printf("IF Filter Bandwidth: %v Hz\n", c.TunerParams.BwType.Hz())
					log.Printf("AGC Control: %v\n", c.CtrlParams.Agc.Enable)
					log.Printf("LNA State: %v\n", c.TunerParams.Gain.LNAstate)
					log.Printf("LNA Percent: %d%%\n", int(pct*100))
					return nil
				},
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

			d := detectDrops(params, reset)
			if d != 0 {
				log.Printf("dropped %d samples\n", d)
			}

			_, err = write(conn, interleave(xi, xq))
			if err != nil {
				log.Println(err)
				cancel()
				return
			}
		}),
		session.WithEventCallback(func(eventId api.EventT, tuner api.TunerSelectT, params *api.EventParamsT) {
			switch eventId {
			case api.GainChange:
				p := params.GainParams()
				log.Printf(
					"EventID=%v Tuner=%v GRdB=%d LNAGRdB=%d SystemGain=%.02f\n",
					eventId, tuner, p.GRdB, p.LnaGRdB, p.CurrGain,
				)
			case api.PowerOverloadChange:
				p := params.PowerOverloadParams()
				log.Printf("EventID=%v Tuner=%v Type=%v\n", eventId, tuner, p.PowerOverloadChangeType)
			case api.DeviceRemoved:
				log.Printf("EventID=%v Tuner=%v\n", eventId, tuner)
			case api.RspDuoModeChange:
				p := params.RspDuoModeParams()
				log.Printf("EventID=%v Tuner=%v Type=%v\n", eventId, tuner, p.ModeChangeType)
			default:
				log.Printf("EventID=%v Tuner=%v\n", eventId, tuner)
			}
		}),
	)
	if err != nil {
		return fmt.Errorf("error creating session; %v", err)
	}

	// Blocks until error or the context is canceled.
	err = session.Run(ctx, sess)
	switch err {
	case nil, context.Canceled:
		log.Println("clean exit")
	default:
		return fmt.Errorf("error during session run; %v", err)
	}

	return nil
}

func main() {
	err := rspudp()
	if err != nil {
		log.Fatal(err)
	}
}
