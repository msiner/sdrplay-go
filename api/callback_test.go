// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package api

import (
	"reflect"
	"testing"
)

// TestEventParams is a sanity check for the EventParamsT type that
// is implementing a C union in Go.
func TestEventParams(t *testing.T) {
	t.Parallel()

	p := EventParamsT{}
	gain := p.GainParams()
	pwr := p.PowerOverloadParams()
	duo := p.RspDuoModeParams()

	pt := reflect.TypeOf(p)
	typs := []reflect.Type{
		reflect.TypeOf(*gain),
		reflect.TypeOf(*pwr),
		reflect.TypeOf(*duo),
	}
	var (
		maxSize  uintptr
		maxAlign int
	)
	for _, typ := range typs {
		if typ.Size() > maxSize {
			maxSize = typ.Size()
		}
		if typ.Align() > maxAlign {
			maxAlign = typ.Align()
		}
	}
	if maxSize != pt.Size() {
		t.Errorf("union members larger than overlay; got %d, want %d", maxSize, pt.Size())
	}
	if maxAlign != pt.Align() {
		t.Errorf("union members have different alignment overlay; got %d, want %d", maxAlign, pt.Align())
	}

	gain.GRdB = 1
	gain.LnaGRdB = 2
	gain.CurrGain = 3

	if pwr.PowerOverloadChangeType != 1 {
		t.Errorf("wrong value in PowerOverloadParams; got %d, want 1", pwr.PowerOverloadChangeType)
	}

	if duo.ModeChangeType != 1 {
		t.Errorf("wrong value in RspDuoModeParams; got %d, want 1", duo.ModeChangeType)
	}

	duo.ModeChangeType = 42

	if gain.GRdB != 42 {
		t.Errorf("wrong value in GainParamsT; got %d, want 42", gain.GRdB)
	}
}
