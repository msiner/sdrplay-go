// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package session

import (
	"errors"
	"fmt"

	"github.com/msiner/sdrplay-go/api"
)

// GetBestZeroIFBW calculates  the largest available analog bandwidth that
// is less than the final effective sample rate of the stream after
// decimation.
func GetBestZeroIFBW(fs float64, dec uint8) (api.Bw_MHzT, error) {
	if fs < 2e6 || fs > 10e6 {
		return api.BW_Undefined, fmt.Errorf("invalid sample rate; got %f Hz, want 2MHz<=Fs<=10MHz", fs)
	}
	switch dec {
	case 1, 2, 4, 8, 16, 32:
		// good
	default:
		return api.BW_Undefined, fmt.Errorf("invalid decimation; got %d, want 1|2|4|8|16|32", dec)
	}
	finalFs := fs / float64(dec)
	switch {
	case finalFs > 8e6:
		return api.BW_8_000, nil
	case finalFs > 7e6:
		return api.BW_7_000, nil
	case finalFs > 6e6:
		return api.BW_6_000, nil
	case finalFs > 5e6:
		return api.BW_5_000, nil
	case finalFs > 1.536e6:
		return api.BW_1_536, nil
	case finalFs > 600e3:
		return api.BW_0_600, nil
	case finalFs > 300e3:
		return api.BW_0_300, nil
	default:
		return api.BW_0_200, nil
	}
}

func SetZeroIF(d *api.DeviceT, p *api.DeviceParamsT, c *api.RxChannelParamsT, fs float64, dec uint8) error {
	bw, err := GetBestZeroIFBW(fs, dec)
	if err != nil {
		return err
	}
	p.DevParams.FsFreq.FsHz = fs
	var decEnable uint8
	if dec > 1 {
		decEnable = 1
	}

	c.TunerParams.BwType = bw
	c.TunerParams.IfType = api.IF_Zero
	c.CtrlParams.Decimation.Enable = decEnable
	c.CtrlParams.Decimation.DecimationFactor = uint8(dec)

	if d.HWVer == api.RSPduo_ID {
		if d.RspDuoMode != api.RspDuoMode_Single_Tuner {
			return fmt.Errorf("zero IF only valid for RSPduo single tuner mode; got %v", d.RspDuoMode)
		}
		if d.Tuner == api.Tuner_Both {
			return errors.New("zero IF not valid for dual tuner mode")
		}
	}
	return nil
}

// WithZeroIF creates a function to configure the selected device to
// operate in zero-IF mode. It will use GetBestZeroIFBW with the
// given arguments to determine the optimal analog bandwidth.
func WithZeroIF(fs float64, dec uint8) ChanConfigFn {
	return func(d *api.DeviceT, p *api.DeviceParamsT, c *api.RxChannelParamsT) error {
		return SetZeroIF(d, p, c, fs, dec)
	}
}
