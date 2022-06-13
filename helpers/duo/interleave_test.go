// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package duo

import (
	"math"
	"math/rand"
	"testing"
)

func TestInterleave(t *testing.T) {
	t.Parallel()

	const (
		maxSamples      = 10000
		scalarsPerFrame = 4
	)
	xia := make([]int16, maxSamples)
	xqa := make([]int16, maxSamples)
	xib := make([]int16, maxSamples)
	xqb := make([]int16, maxSamples)
	for i := range xia {
		xia[i] = int16(rand.Int31n(math.MaxInt16))
		xqa[i] = int16(rand.Int31n(math.MaxInt16))
		xib[i] = int16(rand.Int31n(math.MaxInt16))
		xqb[i] = int16(rand.Int31n(math.MaxInt16))
	}
	inter := NewInterleaveFn()
	for i := 0; i < 100; i++ {
		numSamples := rand.Intn(maxSamples)
		x := inter(xia[:numSamples], xqa[:numSamples], xib[:numSamples], xqb[:numSamples])
		for j := range xia[:numSamples] {
			curr := scalarsPerFrame * j
			if x[curr] != xia[j] {
				t.Errorf("wrong value for xia[%d]", j)
			}
			if x[curr+1] != xqa[j] {
				t.Errorf("wrong value for xqa[%d]", j)
			}
			if x[curr+2] != xib[j] {
				t.Errorf("wrong value for xia[%d]", j)
			}
			if x[curr+3] != xqb[j] {
				t.Errorf("wrong value for xqb[%d]", j)
			}
		}
	}

	x := inter(xia[:1], xqa, xib, xqb)
	if len(x) != scalarsPerFrame {
		t.Errorf("wrong length on unbalanced interleave: got %d, want %d", len(x), scalarsPerFrame)
	}
	x = inter(xia, xqa[:2], xib, xqb)
	if len(x) != scalarsPerFrame*2 {
		t.Errorf("wrong length on unbalanced interleave: got %d, want %d", len(x), scalarsPerFrame)
	}
	x = inter(xia, xqa, xib[:3], xqb)
	if len(x) != scalarsPerFrame*3 {
		t.Errorf("wrong length on unbalanced interleave: got %d, want %d", len(x), scalarsPerFrame)
	}
	x = inter(xia, xqa, xib, xqb[:4])
	if len(x) != scalarsPerFrame*4 {
		t.Errorf("wrong length on unbalanced interleave: got %d, want %d", len(x), scalarsPerFrame)
	}
}

func BenchmarkInterleave(b *testing.B) {
	xia := make([]int16, 2048)
	xqa := make([]int16, 2048)
	xib := make([]int16, 2048)
	xqb := make([]int16, 2048)
	inter := NewInterleaveFn()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		inter(xia, xqa, xib, xqb)
	}
}
