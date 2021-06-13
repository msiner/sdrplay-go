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
	"github.com/msiner/sdrplay-go/helpers/parse"
	"github.com/msiner/sdrplay-go/helpers/wav"
	"github.com/msiner/sdrplay-go/session"
)

func rspwav() error {
	flags := flag.NewFlagSet("rspwav", flag.ExitOnError)
	flags.Usage = func() {
		fmt.Fprintln(flags.Output(), strings.TrimSpace(`
Usage: rspwav [FLAGS] <tuneHz> <fileBytes>

rspwav connects to an available RSP device, configures it, and writes
the received samples to a file in WAV format. The WAV file has two
channels (interpreted by audio software as stereo left and right) in
which the first (left) channel is the real or I component and the second
(right) channel is the imaginary or Q component. The samples are stored
as 16-bit signed integers in little-endian format with the components
interleaved (e.g. I1,Q1,I2,Q2,...,In,Qn).

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
	outOpt := flags.String("out", "rsp.wav", "Write WAV file to specified path.")
	lifOpt := flags.Bool("lif", false, strings.TrimSpace(`
Use low-IF mode. In low-IF mode, the effective sample rate, before decimation
is 2 MHz. When -lif is specified, the -fs option cannot be used to configure
the sample rate.`,
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
	floatOpt := flags.Bool("float", false, "Write samples in floating-point format")
	bigOpt := flags.Bool("big", false, "Write samples with big-endian byte order")

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

	freq, err := parse.TuneFrequency(flags.Arg(0))
	if err != nil {
		return err
	}
	numBytes, err := parse.SizeInBytes(flags.Arg(1))
	if err != nil {
		return err
	}
	// Limitation of standard WAV header format.
	if numBytes > 4*1024*1024*1024 {
		return fmt.Errorf("invalid file size; got %d bytes, but WAV has a maximum of 4 GiB", numBytes)
	}

	fs, err := parse.FsFlag(*fsOpt)
	if err != nil {
		return err
	}

	dec, err := parse.DecFlag(*decOpt)
	if err != nil {
		return err
	}

	warm, err := parse.WarmFlag(*warmOpt)
	if err != nil {
		return err
	}

	agcCtl, err := parse.AGCCtlFlag(*agcCtlOpt)
	if err != nil {
		return err
	}

	agcSet, err := parse.AGCSetFlag(*agcSetOpt)
	if err != nil {
		return err
	}

	usb, err := parse.USBFlag(*usbOpt)
	if err != nil {
		return err
	}

	serials, err := parse.SerialsFlag(*serialsOpt)
	if err != nil {
		return err
	}

	duoTuner, err := parse.DuoTunerFlag(*duoTunerOpt)
	if err != nil {
		return err
	}

	dxAnt, err := parse.DxAntFlag(*dxAntOpt)
	if err != nil {
		return err
	}

	rsp2Ant, err := parse.Rsp2AntFlag(*rsp2AntOpt)
	if err != nil {
		return err
	}

	lnaState, lnaPct, err := parse.LNAFlag(*lnaOpt)
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
		serialsFilter = session.NoopDevFilter
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

	var order binary.ByteOrder = binary.LittleEndian
	if *bigOpt {
		order = binary.BigEndian
	}

	// Size of WAV sample, which is one component of an IQ sample.
	var bytesPerSample uint8 = uint8(2) // sizeof(int16)
	sampleFormat := wav.LPCM
	if *floatOpt {
		bytesPerSample = uint8(4) // sizeof(float32)
		sampleFormat = wav.IEEEFloatingPoint
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
	finalFs := uint32(fs / float64(dec))
	if *lifOpt {
		finalFs = uint32(session.LowIFSampleRate / float64(dec))
	}

	head, err := wav.NewHeader(finalFs, 2, bytesPerSample, sampleFormat, order, 0)
	if err != nil {
		return err
	}

	if err := binary.Write(out, order, head); err != nil {
		return err
	}
	totalBytes += uint64(binary.Size(head))

	// Before rspwav exits, seek back to the beginning and
	// update the WAV header with the correct number of samples and
	// flush the buffered writer.
	defer func() {
		dataBytes := totalBytes - uint64(binary.Size(head))
		numFrames := uint32(dataBytes / uint64(bytesPerSample) / 2)
		log.Printf("update WAV header; dataBytes=%d dataFrames=%d", dataBytes, numFrames)
		head.Update(numFrames)
		out.Flush()
		_, err = fout.Seek(0, io.SeekStart)
		if err != nil {
			log.Printf("failed to seek back to header; %v", err)
		}
		if err := binary.Write(fout, order, head); err != nil {
			log.Printf("failed to update header; %v", err)
		}
	}()

	// Setup callback and control state.
	interleave := callback.NewInterleaveFn()
	toFloats := callback.NewConvertToFloat32Fn(16)
	writeInts := callback.NewWriteFn(order)
	writeFloats := callback.NewFloat32WriteFn(order)
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

	err = session.Run(
		ctx,
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
				log.Printf("dropped %d samples: %d\n", d, totalBytes)
			}
			var (
				n   int
				err error
			)
			switch *floatOpt {
			case true:
				n, err = writeFloats(out, toFloats(interleave(xi, xq)))
			default:
				n, err = writeInts(out, interleave(xi, xq))
			}
			totalBytes += uint64(n)
			switch {
			case err != nil:
				log.Printf("write failed, cancel; %v\n", err)
				cancel()
			case totalBytes > numBytes:
				cancel()
			}
		}),
		session.WithEventCallback(func(eventId api.EventT, tuner api.TunerSelectT, params *api.EventParamsT) {
			switch eventId {
			case api.GainChange:
				p := params.GainParams
				log.Printf(
					"EventID=%v Tuner=%v GRdB=%d LNAGRdB=%d SystemGain=%.02f\n",
					eventId, tuner, p.GRdB, p.LnaGRdB, p.CurrGain,
				)
			case api.PowerOverloadChange:
				p := params.PowerOverloadParams
				log.Printf("EventID=%v Tuner=%v Type=%v\n", eventId, tuner, p.PowerOverloadChangeType)
			case api.DeviceRemoved:
				log.Printf("EventID=%v Tuner=%v\n", eventId, tuner)
			case api.RspDuoModeChange:
				p := params.RspDuoModeParams
				log.Printf("EventID=%v Tuner=%v Type=%v\n", eventId, tuner, p.ModeChangeType)
			default:
				log.Printf("EventID=%v Tuner=%v\n", eventId, tuner)
			}
		}),
	)
	switch err {
	case nil, context.Canceled:
		log.Println("clean exit")
	default:
		return fmt.Errorf("error during session run; %v", err)
	}

	return nil
}

func main() {
	err := rspwav()
	if err != nil {
		log.Fatal(err)
	}
}
