// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package parse

import (
	"testing"

	"github.com/msiner/sdrplay-go/api"
)

func TestLNAFlag(t *testing.T) {
	state, percent, err := LNAFlag("")
	if state != nil || percent != nil || err != nil {
		t.Errorf("non-nil result on empty string: got %v, %v, %v", state, percent, err)
	}

	specs := []struct {
		arg       string
		valid     bool
		isPercent bool
		want      uint8
	}{
		{"100%", true, true, 100},
		{"12", true, false, 12},
		{"two", false, false, 0},
		{"%", false, false, 0},
		{"101%", false, false, 0},
		{"256", false, false, 0},
		{"28", false, false, 0},
		{"27", true, false, 27},
		{"0", true, false, 0},
	}

	for i, spec := range specs {
		gotInt, gotFloat, err := LNAFlag(spec.arg)
		switch {
		case !spec.valid && err == nil:
			t.Errorf("%d: unexpected success", i)
		case !spec.valid && err != nil:
			// expected error
		case spec.valid && err != nil:
			t.Errorf("%d: unexpected error: %v", i, err)
		case spec.isPercent && gotInt != nil:
			t.Errorf("%d: got %v, want percent", i, *gotInt)
		case !spec.isPercent && gotFloat != nil:
			t.Errorf("%d: got %v, want int", i, *gotFloat)
		case spec.isPercent:
			got := uint8(*gotFloat * 100)
			if got != spec.want {
				t.Errorf("%d: wrong value: got %v, want %v", i, got, spec.want)
			}
		default:
			if *gotInt != spec.want {
				t.Errorf("%d: wrong value: got %v, want %v", i, *gotInt, spec.want)
			}
		}
	}
}

func TestSerialsFlag(t *testing.T) {
	const longest = "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz0123456789ab"
	var longestSerial api.SerialNumber
	if len(longest) != len(longestSerial) {
		t.Fatalf("incorrect assumption of serial length: got %v, want %v", len(longestSerial), len(longest))
	}
	for i := 0; i < len(longest); i++ {
		longestSerial[i] = longest[i]
	}
	tooLong := longest + "z"

	specs := []struct {
		arg   string
		valid bool
		want  []api.SerialNumber
	}{
		{"abc", true, []api.SerialNumber{
			{'a', 'b', 'c'},
		}},
		{"abcdefg", true, []api.SerialNumber{
			{'a', 'b', 'c', 'd', 'e', 'f', 'g'},
		}},

		{"ab123fg", true, []api.SerialNumber{
			{'a', 'b', '1', '2', '3', 'f', 'g'},
		}},
		{"abc,defg", true, []api.SerialNumber{
			{'a', 'b', 'c'},
			{'d', 'e', 'f', 'g'},
		}},
		{"abc,defg,123", true, []api.SerialNumber{
			{'a', 'b', 'c'},
			{'d', 'e', 'f', 'g'},
			{'1', '2', '3'},
		}},
		{"abc#", false, nil},
		{"ab\u0000c", false, nil},
		{"世界", false, nil},
		{"#abc", false, nil},
		{longest, true, []api.SerialNumber{longestSerial}},
		{tooLong, false, nil},
	}

	for i, spec := range specs {
		got, err := SerialsFlag(spec.arg)
		switch {
		case !spec.valid && err == nil:
			t.Errorf("%d: unexpected success", i)
		case !spec.valid && err != nil:
			// expected error
		case spec.valid && err != nil:
			t.Errorf("%d: unexpected error: %v", i, err)
		case len(got) != len(spec.want):
			t.Errorf("%d: wrong number of values: got %v, want %v", i, len(got), len(spec.want))
		default:
			for j := range got {
				if got[j] != spec.want[j] {
					t.Errorf("%d: wrong value at %d: got %v, want %v", i, j, got[j], spec.want[j])
				}
			}
		}
	}
}
