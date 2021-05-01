// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package parse

import (
	"testing"
)

func TestParseSize(t *testing.T) {
	t.Parallel()

	const (
		KIL = 1 << 10
		MEG = 1 << 20
		GIG = 1 << 30
	)

	specs := []struct {
		txt   string
		want  uint64
		valid bool
	}{
		{"0", 0, true},
		{"0G", 0, true},
		{"1e6", 0, false},
		{"1m", MEG, true},
		{"1M", MEG, true},
		{"1000000", 1000000, true},
		{"-1", 0, false},
		{"1J", 0, false},
		{"1.123456M", 0, false},
		{"1123.456k", 0, false},
		{"3", 3, true},
		{"3k", 3 * KIL, true},
		{"3K", 3 * KIL, true},
		{"3m", 3 * MEG, true},
		{"3M", 3 * MEG, true},
		{"3g", 3 * GIG, true},
		{"3G", 3 * GIG, true},
		{"3t", 0, false},
		{"3T", 0, false},
	}

	for _, spec := range specs {
		val, err := ParseSize(spec.txt)
		switch {
		case spec.valid && err != nil:
			t.Errorf("parse error on valid input; %v", err)
		case !spec.valid && err != nil:
			// good
		case !spec.valid && err == nil:
			t.Errorf("no error on invalid input %v", spec.txt)
		case val != spec.want:
			t.Errorf("wrong parse result; got %v, want %v", val, spec.want)
		default:
			// good
		}

	}
}
