// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package session

import (
	"errors"

	"github.com/msiner/sdrplay-go/api"
)

func GetEffectiveSampleRate(d *api.DeviceT, p *api.DeviceParamsT, c *api.RxChannelParamsT) (float64, error) {
	if c == nil {
		return 0, errors.New("cannot inspect nil channel")
	}
	fsHz := p.DevParams.FsFreq.FsHz
	dec := c.CtrlParams.Decimation.DecimationFactor
	if dec == 0 {
		return 0, errors.New("invalid decimation factor of zero")
	}
	ifType := c.TunerParams.IfType
	if c.TunerParams.IfType != api.IF_Zero {
		down := false
		bwType := c.TunerParams.BwType
		// See API documentation for sdrplay_api_Init(). Downconversion is only
		// enabled if at least one of the following criteria is met.
		switch {
		case d.HWVer == api.RSPduo_ID && d.RspDuoMode == api.RspDuoMode_Dual_Tuner:
			down = true
		case d.HWVer == api.RSPduo_ID && d.RspDuoMode == api.RspDuoMode_Primary:
			down = true
		case d.HWVer == api.RSPduo_ID && d.RspDuoMode == api.RspDuoMode_Secondary:
			down = true
		case fsHz == 8192000 && bwType == api.BW_1_536 && ifType == api.IF_2_048:
			down = true
		case fsHz == 8000000 && bwType == api.BW_1_536 && ifType == api.IF_2_048:
			down = true
		case fsHz == 8000000 && bwType == api.BW_5_000 && ifType == api.IF_2_048:
			down = true
		case fsHz == 2000000 && bwType <= api.BW_0_300 && ifType == api.IF_0_450:
			down = true
		case fsHz == 2000000 && bwType == api.BW_0_600 && ifType == api.IF_0_450:
			down = true
		case fsHz == 6000000 && bwType <= api.BW_1_536 && ifType == api.IF_1_620:
			down = true
		}
		if down {
			fsHz = 2e6
		}
	}
	return fsHz / float64(dec), nil
}
