// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package api

import (
	"testing"
	"unsafe"
)

func BenchmarkHandleCallback(b *testing.B) {
	numSamples := 1008
	xi := make([]int16, numSamples)
	xq := make([]int16, numSamples)
	params := StreamCbParamsT{}

	cb := func(_, _ []int16, _ *StreamCbParamsT, _ bool) {}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		handleCallback(
			cb,
			unsafe.Pointer(&xi[0]),
			unsafe.Pointer(&xq[0]),
			unsafe.Pointer(&params),
			uintptr(numSamples),
			0,
		)
	}
}

func TestHandleCallback(t *testing.T) {
	t.Parallel()

	numSamples := 1008
	xi := make([]int16, numSamples)
	xq := make([]int16, numSamples)
	params := StreamCbParamsT{FirstSampleNum: 42}

	for i := range xi {
		xi[i] = int16(i)
		xq[i] = int16(-i)
	}

	cb := func(cxi, cxq []int16, cparams *StreamCbParamsT, reset bool) {
		if len(cxi) != numSamples {
			t.Errorf("xi had wrong number of samples; got %d, want %d", len(cxi), numSamples)
		}
		if len(cxq) != numSamples {
			t.Errorf("xq had wrong number of samples; got %d, want %d", len(cxq), numSamples)
		}
		if cparams.FirstSampleNum != params.FirstSampleNum {
			t.Errorf(
				"params had wrong first sample num; got %d, want %d",
				cparams.FirstSampleNum,
				params.FirstSampleNum,
			)
		}
		if !reset {
			t.Errorf("wrong value passed as reset; got false, want true")
		}

		for i := range xi {
			if xi[i] != cxi[i] {
				t.Errorf("wrong data in xi at %d; got %d, want %d", i, cxi[i], xi[i])
				break
			}
			if xq[i] != cxq[i] {
				t.Errorf("wrong data in xq at %d; got %d, want %d", i, cxq[i], xq[i])
				break
			}
		}
	}

	handleCallback(
		cb,
		unsafe.Pointer(&xi[0]),
		unsafe.Pointer(&xq[0]),
		unsafe.Pointer(&params),
		uintptr(numSamples),
		1,
	)

	called := false
	cb = func(cxi, cxq []int16, cparams *StreamCbParamsT, reset bool) {
		called = true
		if cxi != nil {
			t.Errorf("expected nil xi; got len of %d", len(cxi))
		}
		if cxq != nil {
			t.Errorf("expected nil xq; got len of %d", len(cxq))
		}
		if cparams != nil {
			t.Errorf("expected nil cxi; got len of %d", len(cxi))
		}
		if reset {
			t.Errorf("wrong value passed as reset; got true, want false")
		}
	}

	handleCallback(cb, nil, nil, nil, 0, 0)

	if !called {
		t.Error("callback not called on zero samples")
	}

	called = false
	handleCallback(cb, nil, nil, nil, uintptr(numSamples), 0)

	if !called {
		t.Error("callback not called on zero samples and positive numSamples")
	}
}
