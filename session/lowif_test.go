// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package session

import (
	"testing"

	"github.com/msiner/sdrplay-go/api"
)

func TestWithLowIF(t *testing.T) {
	t.Parallel()

	specs := []struct {
		strat LowIFStrategy
		dec   uint8
		fs    float64
		bw    api.Bw_MHzT
		inter api.If_kHzT
	}{
		{LowIFMaxBits, 1, 6e6, api.BW_1_536, api.IF_1_620},
		{LowIFMaxFs, 1, 8e6, api.BW_1_536, api.IF_2_048},
		{LowIFMinFs, 1, 6e6, api.BW_1_536, api.IF_1_620},
		{LowIFMaxBits, 2, 6e6, api.BW_0_600, api.IF_1_620},
		{LowIFMaxFs, 2, 6e6, api.BW_0_600, api.IF_1_620},
		{LowIFMinFs, 2, 2e6, api.BW_0_600, api.IF_0_450},
		{LowIFMaxBits, 4, 6e6, api.BW_0_300, api.IF_1_620},
		{LowIFMaxFs, 4, 6e6, api.BW_0_300, api.IF_1_620},
		{LowIFMinFs, 4, 2e6, api.BW_0_300, api.IF_0_450},
		{LowIFMaxBits, 8, 6e6, api.BW_0_200, api.IF_1_620},
		{LowIFMaxFs, 8, 6e6, api.BW_0_200, api.IF_1_620},
		{LowIFMinFs, 8, 2e6, api.BW_0_200, api.IF_0_450},
		{LowIFMaxBits, 16, 6e6, api.BW_0_200, api.IF_1_620},
		{LowIFMaxFs, 16, 6e6, api.BW_0_200, api.IF_1_620},
		{LowIFMinFs, 16, 2e6, api.BW_0_200, api.IF_0_450},
		{LowIFMaxBits, 32, 6e6, api.BW_0_200, api.IF_1_620},
		{LowIFMaxFs, 32, 6e6, api.BW_0_200, api.IF_1_620},
		{LowIFMinFs, 32, 2e6, api.BW_0_200, api.IF_0_450},
	}

	for i, spec := range specs {
		dev := &api.DeviceT{}
		params := &api.DeviceParamsT{DevParams: &api.DevParamsT{}}
		rx := &api.RxChannelParamsT{}
		cfg := WithLowIF(spec.strat, spec.dec)
		err := cfg(dev, params, rx)
		switch {
		case err != nil:
			t.Errorf("%d: unexpected failure: %v", i, err)
		default:
			fs := params.DevParams.FsFreq.FsHz
			if fs != spec.fs {
				t.Errorf("%d: wrong sample rate: got %v, want %v", i, fs, spec.fs)
			}
			bw := rx.TunerParams.BwType
			if bw != spec.bw {
				t.Errorf("%d: wrong bandwidth: got %v, want %v", i, bw, spec.bw)
			}
			inter := rx.TunerParams.IfType
			if inter != spec.inter {
				t.Errorf("%d: wrong IF: got %v, want %v", i, inter, spec.inter)
			}
		}
	}

	for i, spec := range specs {
		dev := &api.DeviceT{
			HWVer:            api.RSPduo_ID,
			RspDuoSampleFreq: 8e6,
			RspDuoMode:       api.RspDuoMode_Secondary,
		}
		params := &api.DeviceParamsT{DevParams: &api.DevParamsT{}}
		rx := &api.RxChannelParamsT{}
		cfg := WithLowIF(spec.strat, spec.dec)
		err := cfg(dev, params, rx)
		switch {
		case err != nil:
			t.Errorf("%d: unexpected failure: %v", i, err)
		default:
			// As a secondary, it should not configure the Fs.
			fs := params.DevParams.FsFreq.FsHz
			if fs != 0 {
				t.Errorf("%d: wrong sample rate: got %v, want %v", i, fs, 0)
			}
			bw := rx.TunerParams.BwType
			if bw != api.BW_1_536 {
				t.Errorf("%d: wrong bandwidth: got %v, want %v", i, bw, api.BW_1_536)
			}
			inter := rx.TunerParams.IfType
			if inter != api.IF_2_048 {
				t.Errorf("%d: wrong IF: got %v, want %v", i, inter, api.IF_2_048)
			}
		}
	}

	for i, spec := range specs {
		dev := &api.DeviceT{
			HWVer:            api.RSPduo_ID,
			RspDuoSampleFreq: 8e6,
			RspDuoMode:       api.RspDuoMode_Primary,
		}
		params := &api.DeviceParamsT{DevParams: &api.DevParamsT{}}
		rx := &api.RxChannelParamsT{}
		cfg := WithLowIF(spec.strat, spec.dec)
		err := cfg(dev, params, rx)
		switch {
		case err != nil:
			t.Errorf("%d: unexpected failure: %v", i, err)
		default:
			// As a primary, stuck with Fs from DeviceT selection.
			fs := params.DevParams.FsFreq.FsHz
			if fs != 8e6 {
				t.Errorf("%d: wrong sample rate: got %v, want %v", i, fs, 8e6)
			}
			bw := rx.TunerParams.BwType
			if bw != api.BW_1_536 {
				t.Errorf("%d: wrong bandwidth: got %v, want %v", i, bw, api.BW_1_536)
			}
			inter := rx.TunerParams.IfType
			if inter != api.IF_2_048 {
				t.Errorf("%d: wrong IF: got %v, want %v", i, inter, api.IF_2_048)
			}
		}
	}

	for i, spec := range specs {
		dev := &api.DeviceT{
			HWVer:            api.RSPduo_ID,
			RspDuoSampleFreq: 8e6,
			RspDuoMode:       api.RspDuoMode_Dual_Tuner,
		}
		params := &api.DeviceParamsT{DevParams: &api.DevParamsT{}}
		rx := &api.RxChannelParamsT{}
		cfg := WithLowIF(spec.strat, spec.dec)
		err := cfg(dev, params, rx)
		switch {
		case err != nil:
			t.Errorf("%d: unexpected failure: %v", i, err)
		default:
			// As primary in dual tuner mode, stuck with Fs from DeviceT selection.
			fs := params.DevParams.FsFreq.FsHz
			if fs != 8e6 {
				t.Errorf("%d: wrong sample rate: got %v, want %v", i, fs, 8e6)
			}
			bw := rx.TunerParams.BwType
			if bw != api.BW_1_536 {
				t.Errorf("%d: wrong bandwidth: got %v, want %v", i, bw, api.BW_1_536)
			}
			inter := rx.TunerParams.IfType
			if inter != api.IF_2_048 {
				t.Errorf("%d: wrong IF: got %v, want %v", i, inter, api.IF_2_048)
			}
		}
	}

	for i := 0; i < 40; i++ {
		dev := &api.DeviceT{}
		params := &api.DeviceParamsT{DevParams: &api.DevParamsT{}}
		rx := &api.RxChannelParamsT{}
		cfg := WithLowIF(LowIFMaxBits, uint8(i))
		err := cfg(dev, params, rx)
		switch i {
		case 1, 2, 4, 8, 16, 32:
			if err != nil {
				t.Errorf("%d: unexpected error: %v", i, err)
			}
		default:
			if err == nil {
				t.Errorf("%d: unexpected success", i)
			}
		}
	}
}
