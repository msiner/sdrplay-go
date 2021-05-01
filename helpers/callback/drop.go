// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package callback

import (
	"math"

	"github.com/msiner/sdrplay-go/api"
)

// NewDropDetect creates a function ath uses the specified subset of
// stream callback parameters to detect and quantify dropped samples.
// Specifically, it uses the FirstSampleNum and NumSamples fields in
// the callback params. To work, it must be called every callback to
// keep the internal state valid. When reset is true, it resets all
// internal state and reports 0 drops.
func NewDropDetect() func(params *api.StreamCbParamsT, reset bool) uint32 {
	var (
		valid         bool
		lastSampleNum uint32
	)
	return func(params *api.StreamCbParamsT, reset bool) uint32 {
		first := params.FirstSampleNum
		last := lastSampleNum
		numSamples := params.NumSamples

		// Compute the new last sample number while compensating
		// for a uint32 wrap.
		var newLast uint32
		toWrap := math.MaxUint32 - first
		switch toWrap < numSamples {
		case true:
			// FirstSampleNum + NumSamples would wrap
			newLast = numSamples - toWrap - 1
		default:
			newLast = first + numSamples
		}
		lastSampleNum = newLast

		// Either the device was reset or we are in an invalid state and
		// we can't determine if a drop has occurred.
		if reset || !valid {
			valid = true
			return 0
		}

		if last == first {
			return 0
		}

		// Compute packet count gap while compensating for an apparent wrap.
		switch first < last {
		case true:
			// possible wrap
			return (math.MaxUint32 - last) + first
		default:
			// possible normal increase
			return first - last
		}
	}
}
