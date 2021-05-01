// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package session

import (
	"fmt"

	"github.com/msiner/sdrplay-go/api"
)

const LowIFSampleRate = 2e6

// LowIFStrategy is a type used to communicate the priorities of the user
// when determining the best combination of sample rate, IF frequency, and
// analog bandwidth in low IF mode.
type LowIFStrategy int

const (
	// LowIFMaxBits is the default strategy that maximizes the effective
	// number of bits in the sample. In practice, this results in always
	// choosing an ADC sample rate of 6 MHz.
	LowIFMaxBits LowIFStrategy = iota

	// LowIFMaxFs is a strategy that always chooses the highest ADC sample
	// rate that is also compatible with the final effective sample rate of
	// the down-converted stream. In practice, this results in choosing
	// 8 MHz when the final effective sample rate is greater than 1.536 MHz
	// and choosing 6 MHz in all other cases. At 8 MHz, the effective number
	// of bits is reduced to 12 for RSP1A, RSPduo, and RSPdx hardware.
	LowIFMaxFs

	// LowIFMinFs is a strategy that chooses the lowest ADC sample rate
	// when the final effective sample rate of the the down-converted stream
	// In practice, this results in choosing the 2 MHz when the final
	// rate is less than 1.536 MHz and choosing 6 MHz in all other cases.
	// The possible benefit of this strategy is to reduce demand on the USB
	// bus and the CPU by reducing the data rate.
	LowIFMinFs
)

// GetBestLowIFSettings calculates the correct and appropriate settings for
// sample rate, analog bandwidth, and IF frequency in low-IF mode given the
// desired strategy and decimation factor. It does not apply any settings to
// the given device pointer. It needs the device pointer due mainly to
// accomodate RSPduo hardware in dual-tuner mode.
//
// It will select the largest available analog bandwidth that is less than
// the final effective sample rate of the stream after down-conversion and
// decimation.
func GetBestLowIFSettings(d *api.DeviceT, strat LowIFStrategy, dec uint8) (float64, api.Bw_MHzT, api.If_kHzT, error) {
	var fsRes float64
	bwRes := api.BW_Undefined
	ifRes := api.IF_Undefined

	switch dec {
	case 1, 2, 4, 8, 16, 32:
		// good
	default:
		return fsRes, bwRes, ifRes, fmt.Errorf("invalid decimation; got %d, want 1|2|4|8|16|32", dec)
	}

	// Maximum final sample rate, after down-conversion, is 2 MHz.
	finalFs := 2e6 / float64(dec)

	// First, determine the ADC sample rate based on hardware and strategy.
	switch {
	case d.HWVer == api.RSPduo_ID && d.RspDuoMode != api.RspDuoMode_Single_Tuner:
		// For the RSPduo in primary or dual-tuner modes, the sample rate is
		// requested in the call to sdrplay_api_SelectDevice(). In secondary
		// mode, the sample rate is pre-determined by the primary.
		fsRes = d.RspDuoSampleFreq
	case strat == LowIFMaxFs:
		switch finalFs > 1.536e6 {
		case true:
			fsRes = 8e6
		default:
			fsRes = 6e6
		}
	case strat == LowIFMinFs:
		switch finalFs > 1.536e6 {
		case true:
			fsRes = 6e6
		default:
			fsRes = 2e6
		}
	default:
		fsRes = 6e6
	}

	// Finally, determine the best best IF frequency and bandwidth for
	// the final sample rate that is compatible with the ADC sample
	// rate. See documentation for sdrplay_api_Init() for more details.
	switch fsRes {
	case 8e6:
		ifRes = api.IF_2_048
		bwRes = api.BW_1_536
	case 2e6:
		ifRes = api.IF_0_450
		switch {
		case finalFs > 600e3:
			bwRes = api.BW_0_600
		case finalFs > 300e3:
			bwRes = api.BW_0_300
		default:
			bwRes = api.BW_0_200
		}
	default:
		ifRes = api.IF_1_620
		switch {
		case finalFs > 1.536e6:
			bwRes = api.BW_1_536
		case finalFs > 600e3:
			bwRes = api.BW_0_600
		case finalFs > 300e3:
			bwRes = api.BW_0_300
		default:
			bwRes = api.BW_0_200
		}
	}

	return fsRes, bwRes, ifRes, nil
}

func SetLowIF(d *api.DeviceT, p *api.DeviceParamsT, c *api.RxChannelParamsT, strat LowIFStrategy, dec uint8) error {
	fsBest, bwBest, ifBest, err := GetBestLowIFSettings(d, strat, dec)
	if err != nil {
		return err
	}

	var decEnable uint8
	if dec > 1 {
		decEnable = 1
	}

	c.TunerParams.BwType = bwBest
	c.TunerParams.IfType = ifBest
	c.CtrlParams.Decimation.Enable = decEnable
	c.CtrlParams.Decimation.DecimationFactor = uint8(dec)

	switch d.HWVer {
	case api.RSPduo_ID:
		if d.RspDuoMode != api.RspDuoMode_Secondary {
			p.DevParams.FsFreq.FsHz = fsBest
		}
	default:
		p.DevParams.FsFreq.FsHz = fsBest
	}
	return nil
}

// WithLowIF creates a function to configure the selected device to
// operate in low-IF mode. It will use GetBestLowIFSettings with the
// given arguments to determine the optimal settings.
func WithLowIF(strat LowIFStrategy, dec uint8) ChanConfigFn {
	return func(d *api.DeviceT, p *api.DeviceParamsT, c *api.RxChannelParamsT) error {
		return SetLowIF(d, p, c, strat, dec)
	}
}
