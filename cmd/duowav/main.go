// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
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
	"github.com/msiner/sdrplay-go/helpers/wav"
	"github.com/msiner/sdrplay-go/session"
)

func duowav() error {
	flags := flag.NewFlagSet("duowav", flag.ExitOnError)
	flags.Usage = func() {
		fmt.Fprintln(flags.Output(), strings.TrimSpace(`
Usage: duowav [FLAGS] <tuneHz> <fileBytes>

duowav connects to an available RSPduo device, configures it, and writes
the received samples to a file in WAV format. The WAV file has four
channels (interpreted by audio software as front left, front right, front
center, and low frequency) corresponding to tuner A I, tuner A Q,
tuner B I, and tuner B Q respectively where I and Q are the real and
imaginary components respectively. The samples are stored
as 16-bit signed integers in little-endian format with the components
interleaved (e.g. AI1,AQ1,BI1,BQ1,...,AIn,AQn,BIn,BQn).

Arguments:
  tuneHz
	Tuner RF frequency in Hz is a mandatory argument. It can
	be specified with k, K, m, M, g, or G suffix to indicate the
	value is in kHz, MHz, or GHz respectively (e.g. 1.42G).
  fileBytes
	Maximum output file size in bytes. It can be specified with
	k, K, m, M, g, or G suffix to indicate the value is in KiB, MiB,
	or GiB respectively (e.g. 10M)
	NOTE: WAV files cannot exceed 4 GiB.

Flags:
`,
		))
		flags.PrintDefaults()
	}
	outOpt := flags.String("out", "duo.wav", "Write WAV file to specified path.")
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
	floatOpt := flags.Bool("float", false, "Write samples in floating-point format")
	bigOpt := flags.Bool("big", false, "Write samples with big-endian byte order")

	flags.Parse(os.Args[1:])

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
	numBytes, err := parse.ParseSize(flags.Arg(1))
	if err != nil {
		return err
	}
	// Limitation of standard WAV header format.
	if numBytes > 4*1024*1024*1024 {
		return fmt.Errorf("invalid file size; got %d bytes, but WAV has a maximum of 4 GiB", numBytes)
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

	var order binary.ByteOrder = binary.LittleEndian
	if *bigOpt {
		order = binary.BigEndian
	}

	// Size of WAV sample, which is one component of an IQ sample.
	bytesPerSample := uint8(2) // sizeof(int16)
	if *floatOpt {
		bytesPerSample = uint8(4) // sizeof(float32)
	}

	// Setup buffered file output.
	fout, err := os.Create(*outOpt)
	if err != nil {
		return err
	}
	defer fout.Close()
	out := bufio.NewWriterSize(fout, 1024*1024)

	// Write the initial WAV header with 0 samples.
	var totalBytes uint64
	finalFs := uint32(session.LowIFSampleRate / float64(dec))
	head := wav.NewHeader(finalFs, 4, bytesPerSample, *floatOpt, *bigOpt, 0)
	if err := binary.Write(out, order, head); err != nil {
		return err
	}
	totalBytes += uint64(binary.Size(head))

	// Before rspwav exits, seek back to the beginning and
	// update the WAV header with the correct number of samples and
	// flush the buffered writer.
	defer func() {
		dataBytes := totalBytes - uint64(binary.Size(head))
		numFrames := uint32(dataBytes / uint64(bytesPerSample) / 4)
		lg.Printf("update WAV header; dataBytes=%d dataFrames=%d", dataBytes, numFrames)
		head.Update(numFrames)
		out.Flush()
		_, err = fout.Seek(0, io.SeekStart)
		if err != nil {
			lg.Printf("failed to seek back to header; %v", err)
		}
		if err := binary.Write(fout, order, head); err != nil {
			lg.Printf("failed to update header; %v", err)
		}
	}()

	// Setup callback and control state.
	interleave := duo.NewInterleave()
	toFloats := callback.NewConvertToFloat32(16)
	writeInts := callback.NewWrite(*bigOpt)
	writeFloats := callback.NewFloat32Write(*bigOpt)
	detectDropsA := callback.NewDropDetect()
	detectDropsB := callback.NewDropDetect()

	var isWarm uint32
	go func() {
		time.Sleep(warm)
		lg.Println("warm-up complete")
		atomic.StoreUint32(&isWarm, 1)
	}()

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

	evtChan := event.NewEventChan(10)
	defer evtChan.Close()

	synchro := duo.NewSynchro(
		10000,
		func(xia, xqa, xib, xqb []int16, reset bool) {
			select {
			case <-ctx.Done():
				return
			default:
			}

			x := interleave(xia, xqa, xib, xqb)

			var (
				n   int
				err error
			)
			switch *floatOpt {
			case true:
				n, err = writeFloats(out, toFloats(x))
			default:
				n, err = writeInts(out, x)
			}

			totalBytes += uint64(n)
			switch {
			case err != nil:
				lg.Printf("write failed, cancel; %v\n", err)
				cancel()
			case totalBytes > numBytes:
				cancel()
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
	if err := session.Run(ctx, sess); err != nil {
		return fmt.Errorf("error during session run; %v", err)
	}

	lg.Println("clean exit")

	return nil
}

func main() {
	err := duowav()
	if err != nil {
		log.Fatal(err)
	}
}
