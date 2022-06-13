// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package callback

import (
	"math"
	"testing"

	"github.com/msiner/sdrplay-go/api"
)

func TestDropDetect(t *testing.T) {
	t.Parallel()

	const numSamples = 1234

	var n uint32
	params := &api.StreamCbParamsT{
		FirstSampleNum: 5678,
		NumSamples:     numSamples,
	}

	detect := NewDropDetectFn()

	// Doesn't matter what the numbers are, always zero on reset.
	n = detect(params, true)
	if n != 0 {
		t.Errorf("drops detected on reset: got %d, want 0", n)
	}

	// Do a simple, non-wrapping increase
	params.FirstSampleNum += numSamples
	n = detect(params, false)
	if n != 0 {
		t.Errorf("drops detected normal increment: got %d, want 0", n)
	}

	// Omit 1 sample in the increase
	params.FirstSampleNum += numSamples + 1
	n = detect(params, false)
	if n != 1 {
		t.Errorf("wrong number of droops: got %d, want 1", n)
	}

	// Reset to a state where we are 1 sample away from the count wrapping.
	params.FirstSampleNum = math.MaxUint32 - 1
	n = detect(params, true)
	if n != 0 {
		t.Errorf("drops detected on reset: got %d, want 0", n)
	}

	// Wrap the count.
	params.FirstSampleNum += numSamples
	n = detect(params, false)
	if n != 0 {
		t.Errorf("drops detected wrapped increment: got %d, want 0", n)
	}

	// Reset to a state where we are two updates away from wrapping.
	params.FirstSampleNum = math.MaxUint32 - numSamples - 1
	n = detect(params, true)
	if n != 0 {
		t.Errorf("drops detected on reset: got %d, want 0", n)
	}

	// Normal increase.
	params.FirstSampleNum += numSamples
	n = detect(params, false)
	if n != 0 {
		t.Errorf("drops detected normal increment: got %d, want 0", n)
	}

	// Short the count by 1 and drop the maximum number of samples.
	params.FirstSampleNum += numSamples - 1
	n = detect(params, false)
	if n != (math.MaxUint32 - 1) {
		t.Errorf("wrong number of drops: got %d, want %d", n, uint32(math.MaxUint32-1))
	}
}
