// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package session

import (
	"errors"
	"fmt"

	"github.com/msiner/sdrplay-go/api"
)

func GetMaxLNAState(d *api.DeviceT, p *api.DeviceParamsT, c *api.RxChannelParamsT) uint8 {
	if c == nil {
		return 0
	}
	freq := c.TunerParams.RfFreq.RfHz
	switch d.HWVer {
	case api.RSP1_ID:
		return 3
	case api.RSP1A_ID:
		switch {
		case freq <= 60e6:
			return 6
		case freq <= 1000e6:
			return 9
		default:
			return 8
		}
	case api.RSP2_ID:
		switch {
		case freq <= 60e6:
			switch p.RxChannelA.Rsp2TunerParams.AmPortSel {
			case api.Rsp2_AMPORT_1:
				return 4
			default:
				return 8
			}
		case freq <= 420e6:
			return 8
		default:
			return 5
		}
	case api.RSPduo_ID:
		switch {
		case freq <= 60e6:
			switch p.RxChannelA.RspDuoTunerParams.Tuner1AmPortSel {
			case api.RspDuo_AMPORT_1:
				return 4
			default:
				return 6
			}
		case freq <= 420e6:
			return 9
		default:
			return 8
		}
	case api.RSPdx_ID:
		switch {
		case freq <= 2e6:
			switch p.DevParams.RspDxParams.HdrEnable {
			case 0:
				return 18
			default:
				return 21
			}
		case freq <= 12e6:
			return 18
		case freq <= 60e6:
			return 19
		case freq <= 250e6:
			return 26
		case freq <= 420e6:
			return 27
		case freq <= 1000e6:
			return 20
		default:
			return 18
		}
	default:
		return 0
	}
}

func SetLNAState(d *api.DeviceT, p *api.DeviceParamsT, c *api.RxChannelParamsT, val uint8) error {
	max := GetMaxLNAState(d, p, p.RxChannelA)
	if val > max {
		return fmt.Errorf("invalid LNA state; got %d, want <= %d", val, max)
	}
	c.TunerParams.Gain.LNAstate = uint8(val)
	return nil
}

// WithLNAState creates a function that sets the LNA state to the provided
// value.
func WithLNAState(val uint8) ChanConfigFn {
	return func(d *api.DeviceT, p *api.DeviceParamsT, c *api.RxChannelParamsT) error {
		return SetLNAState(d, p, c, val)
	}
}

func GetLNAPercent(d *api.DeviceT, p *api.DeviceParamsT, c *api.RxChannelParamsT) (float64, error) {
	if c == nil {
		return 0, errors.New("cannot inspect nil channel")
	}
	max := GetMaxLNAState(d, p, p.RxChannelA)
	val := c.TunerParams.Gain.LNAstate
	if val > max {
		return 0, fmt.Errorf("invalid LNA state; got %d, want <= %d", val, max)
	}
	return float64(max-val) / float64(max), nil
}

func SetLNAPercent(d *api.DeviceT, p *api.DeviceParamsT, c *api.RxChannelParamsT, pct float64) error {
	if c == nil {
		return errors.New("cannot configure nil channel")
	}
	if pct < 0 || pct > 1 {
		return fmt.Errorf("invalid LNA percent; got %f, want 0 <= pct <= 1", pct)
	}
	max := GetMaxLNAState(d, p, p.RxChannelA)
	val := max - uint8(float64(max)*pct)
	c.TunerParams.Gain.LNAstate = val
	return nil
}

func WithLNAPercent(pct float64) ChanConfigFn {
	return func(d *api.DeviceT, p *api.DeviceParamsT, c *api.RxChannelParamsT) error {
		return SetLNAPercent(d, p, c, pct)
	}
}
