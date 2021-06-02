package session

import (
	"strings"
	"testing"

	"github.com/msiner/sdrplay-go/api"
)

func TestWithZeroIF(t *testing.T) {
	t.Parallel()

	specs := []struct {
		fs    float64
		dec   uint8
		valid bool
		err   string
		bw    api.Bw_MHzT
	}{
		{0, 1, false, "invalid sample rate", 0},
		{100, 1, false, "invalid sample rate", 0},
		{1000, 1, false, "invalid sample rate", 0},
		{10000, 1, false, "invalid sample rate", 0},
		{100000, 1, false, "invalid sample rate", 0},
		{1000000, 1, false, "invalid sample rate", 0},
		{2000000, 3, false, "invalid decimation", 0},
		{2000000, 5, false, "invalid decimation", 0},
		{2000000, 6, false, "invalid decimation", 0},
		{2000000, 10, false, "invalid decimation", 0},
		{2000000, 12, false, "invalid decimation", 0},
		{2000000, 14, false, "invalid decimation", 0},
		{2000000, 1, true, "", api.BW_1_536},
		{2000000, 2, true, "", api.BW_0_600},
		{2000000, 4, true, "", api.BW_0_300},
		{2000000, 8, true, "", api.BW_0_200},
		{2000000, 16, true, "", api.BW_0_200},
		{2000000, 32, true, "", api.BW_0_200},
		{3072000, 1, true, "", api.BW_1_536},
		{3072000, 2, true, "", api.BW_0_600},
		{3072000.1, 2, true, "", api.BW_1_536},
		{4000000, 1, true, "", api.BW_1_536},
		{4000000, 2, true, "", api.BW_1_536},
		{4000000, 4, true, "", api.BW_0_600},
		{4000000, 8, true, "", api.BW_0_300},
		{4000000, 16, true, "", api.BW_0_200},
		{4000000, 32, true, "", api.BW_0_200},
		{5000000, 1, true, "", api.BW_1_536},
		{5000000.1, 1, true, "", api.BW_5_000},
		{5000000, 2, true, "", api.BW_1_536},
		{5000000, 4, true, "", api.BW_0_600},
		{5000000, 8, true, "", api.BW_0_600},
		{5000000, 16, true, "", api.BW_0_300},
		{5000000, 32, true, "", api.BW_0_200},
		{6000000, 1, true, "", api.BW_5_000},
		{6000000.1, 1, true, "", api.BW_6_000},
		{6000000, 2, true, "", api.BW_1_536},
		{6000000, 4, true, "", api.BW_0_600},
		{6000000, 8, true, "", api.BW_0_600},
		{6000000, 16, true, "", api.BW_0_300},
		{6000000, 32, true, "", api.BW_0_200},
		{7000000, 1, true, "", api.BW_6_000},
		{7000000.1, 1, true, "", api.BW_7_000},
		{7000000, 2, true, "", api.BW_1_536},
		{7000000, 4, true, "", api.BW_1_536},
		{7000000, 8, true, "", api.BW_0_600},
		{7000000, 16, true, "", api.BW_0_300},
		{7000000, 32, true, "", api.BW_0_200},
		{8000000, 1, true, "", api.BW_7_000},
		{8000000.1, 1, true, "", api.BW_8_000},
		{8000000, 2, true, "", api.BW_1_536},
		{8000000, 4, true, "", api.BW_1_536},
		{8000000, 8, true, "", api.BW_0_600},
		{8000000, 16, true, "", api.BW_0_300},
		{8000000, 32, true, "", api.BW_0_200},
		{9000000, 1, true, "", api.BW_8_000},
		{9000000.1, 1, true, "", api.BW_8_000},
		{9000000, 2, true, "", api.BW_1_536},
		{9000000, 4, true, "", api.BW_1_536},
		{9000000, 8, true, "", api.BW_0_600},
		{9000000, 16, true, "", api.BW_0_300},
		{9000000, 32, true, "", api.BW_0_200},
		{10000000, 1, true, "", api.BW_8_000},
		{10000000.1, 1, false, "invalid sample rate", 0},
		{10000000, 2, true, "", api.BW_1_536},
		{10000000, 4, true, "", api.BW_1_536},
		{10000000, 8, true, "", api.BW_0_600},
		{10000000, 16, true, "", api.BW_0_600},
		{10000000, 32, true, "", api.BW_0_300},
	}

	for i, spec := range specs {
		dev := &api.DeviceT{}
		params := &api.DeviceParamsT{DevParams: &api.DevParamsT{}}
		rx := &api.RxChannelParamsT{}
		cfg := WithZeroIF(spec.fs, spec.dec)
		err := cfg(dev, params, rx)
		switch {
		case spec.valid && err != nil:
			t.Errorf("%d: unexpected failure; %v", i, err)
		case !spec.valid && err == nil:
			t.Errorf("%d: unexpected success", i)
		case !spec.valid && err != nil:
			txt := err.Error()
			if !strings.Contains(txt, spec.err) {
				t.Errorf("%d: wrong error message; got '%s', want '%s'", i, txt, spec.err)
			}
		default:
			bw := rx.TunerParams.BwType
			if bw != spec.bw {
				t.Errorf("%d: wrong bandwidth; got %v, want %v", i, bw, spec.bw)
			}
		}
	}
}
