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
	"github.com/msiner/sdrplay-go/helpers/duo"
	"github.com/msiner/sdrplay-go/helpers/event"
	"github.com/msiner/sdrplay-go/helpers/parse"
	"github.com/msiner/sdrplay-go/helpers/udp"
	"github.com/msiner/sdrplay-go/session"
)

func duoudp() error {
	flags := flag.NewFlagSet("duoudp", flag.ExitOnError)
	flags.Usage = func() {
		fmt.Fprintln(flags.Output(), strings.TrimSpace(`
Usage: duoudp [FLAGS] <tuneHz>

duoudp connects to an available RSPduo device, configures it, and
sends the received samples to the specified UDP target. By default,
duoudp will configure both A and B channels identically. The samples
are then interleaved and framed in the payload as:
[Ia1,Qa1,Ib1,Qb1,Ia2,Qa2,Ib2,Qb2,...,IaN,QaN,IbN,QbN]

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

	lg := log.New(os.Stdout, "", log.LstdFlags)

	freq, err := parse.ParseTuneFrequency(flags.Arg(0))
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
	lnaCfg := session.WithNoopChanConfig()
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

	addr, err := net.ResolveUDPAddr("udp", *remoteOpt)
	if err != nil {
		return err
	}

	conn, err := net.DialUDP(addr.Network(), nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	lg.Printf("UDP initialized: local=%v remote=%v", conn.LocalAddr(), conn.RemoteAddr())

	if *payOpt%4 != 0 {
		return fmt.Errorf("payload size must be multiple of frame size; got %d", *payOpt)
	}

	lg.Printf("Payload Size: %d B", *payOpt)

	// Setup callback and control state.
	write, err := udp.NewUDPPacketWrite(*payOpt, 4, *seqOpt, *bigOpt)
	if err != nil {
		return err
	}
	interleave := duo.NewInterleave()
	detectDropsA := callback.NewDropDetectFn()
	detectDropsB := callback.NewDropDetectFn()

	// Use a simple atomic variable to signal warm-up complete
	var isWarm uint32
	go func() {
		time.Sleep(warm)
		lg.Println("warm-up complete")
		atomic.StoreUint32(&isWarm, 1)
	}()

	// Context and interrupt signal handling
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

	// Event forwarding
	evtChan := event.NewEventChan(10)
	defer evtChan.Close()

	// Duo channel synchronization and callback.
	synchro := duo.NewSynchro(
		10000,
		func(xia, xqa, xib, xqb []int16, reset bool) {
			select {
			case <-ctx.Done():
				return
			default:
			}

			// At this point, we have 4 synchronized components
			// with the same slice length.
			x := interleave(xia, xqa, xib, xqb)
			if _, err := write(conn, x); err != nil {
				lg.Println(err)
				cancel()
				return
			}
		},
		func(evt duo.SynchroEvent, msg string) {
			lg.Printf("%s: %s\n", evt, msg)
		},
	)

	sess, err := session.NewSession(
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
				lg.Printf("stream A: dropped %d samples %v\n", d, reset)
				reset = true
			}
			synchro.UpdateStreamA(xi, xq, params, reset)
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
				lg.Printf("stream B: dropped %d samples %v\n", d, reset)
				reset = true
			}
			synchro.UpdateStreamB(xi, xq, params, reset)
		}),
		session.WithEventCallback(evtChan.Callback),
		session.WithControlLoop(func(_ context.Context, d *api.DeviceT, a api.API) error {
			for {
				select {
				case <-ctx.Done():
					return nil
				case evt := <-evtChan.C:
					event.LogEventMsg(evt, lg)
					event.HandlePowerOverloadChangeMsg(d, a, evt, lg, true)
				}
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
		lg.Println("clean exit")
	default:
		return fmt.Errorf("error during session run; %v", err)
	}

	return nil
}

func main() {
	err := duoudp()
	if err != nil {
		log.Fatal(err)
	}
}
