// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package parse

import (
	"math"
	"testing"
)

func TestParseFrequency(t *testing.T) {
	t.Parallel()

	close := func(a, b float64) bool {
		return math.Abs(a-b) <= 1e-9
	}

	specs := []struct {
		txt   string
		want  float64
		valid bool
	}{
		{"0", 0, true},
		{"0G", 0, true},
		{"1e6", 1e6, true},
		{"1m", 1e6, true},
		{"1M", 1e6, true},
		{"1000000", 1e6, true},
		{"1000000.0000000", 1e6, true},
		{"-1k", -1e3, true},
		{"1J", 0, false},
		{"1.123456M", 1123456, true},
		{"1123.456k", 1123456, true},
		{"abc", 0, false},
		{"M", 0, false},
		{"3.3", 3.3, true},
		{"3.3k", 3.3e3, true},
		{"3.3K", 3.3e3, true},
		{"3.3m", 3.3e6, true},
		{"3.3M", 3.3e6, true},
		{"3.3g", 3.3e9, true},
		{"3.3G", 3.3e9, true},
		{"3.3t", 0, false},
		{"3.3T", 0, false},
		{".00001", 0.00001, true},
		{"0.00001", 0.00001, true},
		{"1e-5", 0.00001, true},
	}

	for _, spec := range specs {
		val, err := Frequency(spec.txt)
		switch {
		case spec.valid && err != nil:
			t.Errorf("parse error on valid input; %v", err)
		case !spec.valid && err != nil:
			// good
		case !spec.valid && err == nil:
			t.Errorf("no error on invalid input %v", spec.txt)
		case !close(val, spec.want):
			t.Errorf("wrong parse result; got %v, want %v", val, spec.want)
		default:
			// good
		}

	}
}
