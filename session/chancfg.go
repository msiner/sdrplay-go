// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package session

import (
	"errors"
	"fmt"

	"github.com/msiner/sdrplay-go/api"
)

// ChanConfigFn is a function type for configuring a single channel. The
// function type is designed to be used during device configuration and
// likely by a DevConfigFn. In most cases, such a function should only
// need to modify the given RxChannelParamsT. The DeviceParamsT is
// provided primarily as information about the configuration, but some
// configuration options that are provided in the channel params for one
// device type, may be in the device params for another device
// (e.g. RfNotchEnable). The function returns a non-nil error if an
// incompatible or impossible configuration is detected or requested.
type ChanConfigFn func(d *api.DeviceT, p *api.DeviceParamsT, c *api.RxChannelParamsT) error

// NoopChanConfig is a channel configuration function that returns nil
// without checking or modifying the params. It can be used as a noop
// or placeholder for another ChanConfigFn function.
func NoopChanConfig(d *api.DeviceT, p *api.DeviceParamsT, c *api.RxChannelParamsT) error {
	return nil
}

// WithSingleChannelConfig creates a DevConfigFn that applies the given
// ChanConfigFn function list to the selected device and channel. The function
// returns an error if the device is an RSPduo with neither or both channels
// selected. Otherwise, it applies the channel configuration to RxChannelA.
func WithSingleChannelConfig(fns ...ChanConfigFn) DevConfigFn {
	return func(d *api.DeviceT, p *api.DeviceParamsT) error {
		apply := func(c *api.RxChannelParamsT) error {
			for _, fn := range fns {
				if err := fn(d, p, c); err != nil {
					return err
				}
			}
			return nil
		}

		switch d.HWVer {
		case api.RSPduo_ID:
			switch d.Tuner {
			case api.Tuner_A:
				return apply(p.RxChannelA)
			case api.Tuner_B:
				return apply(p.RxChannelA)
			case api.Tuner_Both:
				return errors.New("attempting single channel config in dual-tuner mode")
			default:
				return errors.New("no tuner selected")
			}
		default:
			return apply(p.RxChannelA)
		}
	}
}

// WithDuoChannelAConfig creates a DevConfigFn that applies the given
// ChanConfigFn function list to the selected device and channel if the
// device is an RSPduo and tuner A or both tuners are selected. The
// function will return an error if the device is an RSPduo with only
// tuner B selected. If the device is not an RSPduo, the function does
// nothing.
func WithDuoChannelAConfig(fns ...ChanConfigFn) DevConfigFn {
	return func(d *api.DeviceT, p *api.DeviceParamsT) error {
		apply := func(c *api.RxChannelParamsT) error {
			for _, fn := range fns {
				if err := fn(d, p, c); err != nil {
					return err
				}
			}
			return nil
		}

		switch d.HWVer {
		case api.RSPduo_ID:
			switch d.Tuner {
			case api.Tuner_A:
				return apply(p.RxChannelA)
			case api.Tuner_B:
				return errors.New("attempting channel A config when tuner B is selected")
			case api.Tuner_Both:
				return apply(p.RxChannelA)
			default:
				return errors.New("no tuner selected")
			}
		default:
			return nil
		}
	}
}

// WithDuoChannelBConfig creates a DevConfigFn that applies the given
// ChanConfigFn function list to the selected device and channel if the
// device is an RSPduo and tuner B or both tuners are selected. The
// function will return an error if the device is an RSPduo with only
// tuner A selected. If the device is not an RSPduo, the function does
// nothing.
func WithDuoChannelBConfig(fns ...ChanConfigFn) DevConfigFn {
	return func(d *api.DeviceT, p *api.DeviceParamsT) error {
		apply := func(c *api.RxChannelParamsT) error {
			for _, fn := range fns {
				if err := fn(d, p, c); err != nil {
					return err
				}
			}
			return nil
		}

		switch d.HWVer {
		case api.RSPduo_ID:
			switch d.Tuner {
			case api.Tuner_A:
				return errors.New("attempting channel B config when tuner A is selected")
			case api.Tuner_B:
				return apply(p.RxChannelB)
			case api.Tuner_Both:
				return apply(p.RxChannelB)
			default:
				return errors.New("no tuner selected")
			}
		default:
			return nil
		}
	}
}

// SetTuneFreq configures the tune frequency in the given device params.
func SetTuneFreq(d *api.DeviceT, p *api.DeviceParamsT, c *api.RxChannelParamsT, freq float64) error {
	if c == nil {
		return errors.New("cannot configure nil channel")
	}
	if freq < 1e3 || freq > 2e9 {
		return fmt.Errorf("invalid tune frequency; got %f Hz, want 1kHz<=Freq<=2GHz", freq)
	}
	c.TunerParams.RfFreq.RfHz = freq
	return nil
}

// WithTuneFreq creates a function that configures the center RF tuning frequency.
func WithTuneFreq(freq float64) ChanConfigFn {
	return func(d *api.DeviceT, p *api.DeviceParamsT, c *api.RxChannelParamsT) error {
		return SetTuneFreq(d, p, c, freq)
	}
}

// SetGainReduction configures the gain reduction for the specified channel.
func SetGainReduction(d *api.DeviceT, p *api.DeviceParamsT, c *api.RxChannelParamsT, grdb int32) error {
	if c == nil {
		return errors.New("cannot configure nil channel")
	}

	minGr := api.NORMAL_MIN_GR
	switch {
	case grdb > 59:
		grdb = 59
	case grdb < 20:
		minGr = api.EXTENDED_MIN_GR
	}

	c.TunerParams.Gain.GRdB = grdb
	c.TunerParams.Gain.MinGr = minGr

	return nil
}

// WithGainReduction creates a function that configures the gain reduction.
func WithGainReduction(grdb int32) ChanConfigFn {
	return func(d *api.DeviceT, p *api.DeviceParamsT, c *api.RxChannelParamsT) error {
		return SetGainReduction(d, p, c, grdb)
	}
}

// SetBandwidth sets the IF bandwidth. It does not do any checking for
// validity
func SetBandwidth(d *api.DeviceT, p *api.DeviceParamsT, c *api.RxChannelParamsT, bw api.Bw_MHzT) error {
	if c == nil {
		return errors.New("cannot configure nil channel")
	}

	switch bw {
	case api.BW_0_200, api.BW_0_300, api.BW_0_600, api.BW_1_536, api.BW_5_000, api.BW_6_000, api.BW_7_000, api.BW_8_000:
		// good
	default:
		return fmt.Errorf("invalid bandwidth; got %v", bw)
	}
	c.TunerParams.BwType = bw
	return nil
}

// WithBandwidth creates a function that sets the IF bandwidth. The
// WithZeroIF or WithLowIF are also being used. Those functions will
// automatically select the ideal bandwidth. WithBandwidth should only
// be used to override that value.
func WithBandwidth(bw api.Bw_MHzT) ChanConfigFn {
	return func(d *api.DeviceT, p *api.DeviceParamsT, c *api.RxChannelParamsT) error {
		return SetBandwidth(d, p, c, bw)
	}
}

// SetAGC configures the AGC for the specified channel.
func SetAGC(d *api.DeviceT, p *api.DeviceParamsT, c *api.RxChannelParamsT, en api.AgcControlT, set int32) error {
	// Just use the SDRuno defaults for now.
	const (
		attack      = 500
		decay       = 500
		decayDelay  = 200
		decayThresh = 5
	)

	if c == nil {
		return errors.New("cannot configure nil channel")
	}
	if set > 0 {
		return fmt.Errorf("invalid AGC set point; got %d dBFS, want <= 0", set)
	}

	c.CtrlParams.Agc.Enable = en
	c.CtrlParams.Agc.SetPoint_dBfs = set
	c.CtrlParams.Agc.Attack_ms = attack
	c.CtrlParams.Agc.Decay_ms = decay
	c.CtrlParams.Agc.Decay_delay_ms = decayDelay
	c.CtrlParams.Agc.Decay_threshold_dB = decayThresh

	return nil
}

// WithAGC creates a function that configures the AGC as specified.
// This simplified function uses the default values from SDRuno for
// attack, decay, decay delay, and decay threshold.
func WithAGC(en api.AgcControlT, set int32) ChanConfigFn {
	return func(d *api.DeviceT, p *api.DeviceParamsT, c *api.RxChannelParamsT) error {
		return SetAGC(d, p, c, en, set)
	}
}

// SetRfNotchEnabled enables or disables the filter labeled as "RfNotch" in
// the C API. The RSP1A, RSP2, RSPduo, and RSPdx devices have an
// "RfNotch". The individual filter responses may vary, but this
// generally refers to a combined FM/MW notch filter. See the device
// datasheets for details. The function has no effect on RSP1 devices.
func SetRfNotchEnabled(d *api.DeviceT, p *api.DeviceParamsT, c *api.RxChannelParamsT, en bool) error {
	if c == nil {
		return errors.New("cannot configure nil channel")
	}

	var val uint8
	if en {
		val = 1
	}
	switch d.HWVer {
	case api.RSP1_ID:
		// not available
	case api.RSP1A_ID:
		p.DevParams.Rsp1aParams.RfNotchEnable = val
	case api.RSP2_ID:
		c.Rsp2TunerParams.RfNotchEnable = val
	case api.RSPduo_ID:
		c.RspDuoTunerParams.RfNotchEnable = val
	case api.RSPdx_ID:
		p.DevParams.RspDxParams.RfNotchEnable = val
	}
	return nil
}

// WithRfNotchEnabled creates a function that uses SetRfNotchEnabled to enable
// or disable RF notch filter.
func WithRfNotchEnabled(en bool) ChanConfigFn {
	return func(d *api.DeviceT, p *api.DeviceParamsT, c *api.RxChannelParamsT) error {
		return SetRfNotchEnabled(d, p, c, en)
	}
}

// SetRfDabNotchEnabled cenables or disables the filter labeled as
// "RfDabNotch" in the C API. The RSP1A, RSPduo, and RSPdx devices have
// an "RfDabNotch". The individual filter responses may vary, but this
// refers to a digital audio broadcasting (DAB) notch filter. See the
// device datasheets for details. The function has no effect on RSP1
// and RSP2 devices.
func SetRfDabNotchEnabled(d *api.DeviceT, p *api.DeviceParamsT, c *api.RxChannelParamsT, en bool) error {
	if c == nil {
		return errors.New("cannot configure nil channel")
	}

	var val uint8
	if en {
		val = 1
	}
	switch d.HWVer {
	case api.RSP1_ID:
		// not available
	case api.RSP1A_ID:
		p.DevParams.Rsp1aParams.RfDabNotchEnable = val
	case api.RSP2_ID:
		// not available
	case api.RSPduo_ID:
		c.RspDuoTunerParams.RfDabNotchEnable = val
	case api.RSPdx_ID:
		p.DevParams.RspDxParams.RfDabNotchEnable = val
	}
	return nil
}

// WithRfDabNotchEnabled creates a function that uses the
// SetRfDabNotchEnabled function to enable or disable the DAB notch filter.
func WithRfDabNotchEnabled(en bool) ChanConfigFn {
	return func(d *api.DeviceT, p *api.DeviceParamsT, c *api.RxChannelParamsT) error {
		return SetRfDabNotchEnabled(d, p, c, en)
	}
}

// Logger is compatible with standard library and logrus.
type Logger interface {
	Printf(format string, v ...interface{})
}

// WithLogChannelParams creates a function that prints channel configuration
// values to the specified Logger.
func WithLogChannelParams(prefix string, lg Logger) ChanConfigFn {
	return func(d *api.DeviceT, p *api.DeviceParamsT, c *api.RxChannelParamsT) error {
		rate, err := GetEffectiveSampleRate(d, p, c)
		if err != nil {
			return err
		}
		pct, err := GetLNAPercent(d, p, c)
		if err != nil {
			return err
		}
		if prefix != "" {
			prefix += " "
		}
		lg.Printf("%sIFMode=%v", prefix, c.TunerParams.IfType)
		lg.Printf("%sRFFrequency=%vHz\n", prefix, c.TunerParams.RfFreq.RfHz)
		lg.Printf("%sADCSampleRate=%vHz\n", prefix, p.DevParams.FsFreq.FsHz)
		lg.Printf("%sEffectiveSampleRate=%vHz\n", prefix, rate)
		lg.Printf("%sIFFilterBandwidth=%vHz\n", prefix, c.TunerParams.BwType.Hz())
		lg.Printf("%sAGCControl=%v\n", prefix, c.CtrlParams.Agc.Enable)
		lg.Printf("%sLNAState=%v\n", prefix, c.TunerParams.Gain.LNAstate)
		lg.Printf("%sLNAPercent=%d%%\n", prefix, int(pct*100))
		return nil
	}
}
