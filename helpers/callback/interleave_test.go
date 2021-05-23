// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package callback

import (
	"math"
	"math/rand"
	"testing"
)

func TestInterleave(t *testing.T) {
	t.Parallel()

	const maxSamples = 10000
	xi := make([]int16, maxSamples)
	xq := make([]int16, maxSamples)
	for i := range xi {
		xi[i] = int16(rand.Int31n(math.MaxInt16))
		xq[i] = int16(rand.Int31n(math.MaxInt16))
	}
	inter := NewInterleaveFn()
	for i := 0; i < 100; i++ {
		numSamples := rand.Intn(maxSamples)
		x := inter(xi[:numSamples], xq[:numSamples])
		for j := range xi[:numSamples] {
			curr := 2 * j
			if x[curr] != xi[j] {
				t.Errorf("wrong value for xi[%d]", j)
			}
			if x[curr+1] != xq[j] {
				t.Errorf("wrong value for xq[%d]", j)
			}
		}
	}

	x := inter(xi, xq[:1])
	if len(x) != 2 {
		t.Errorf("wrong length on unbalanced interleave; got %d, want 2", len(x))
	}
	x = inter(xi[:1], xq)
	if len(x) != 2 {
		t.Errorf("wrong length on unbalanced interleave; got %d, want 2", len(x))
	}
}

func BenchmarkInterleave(b *testing.B) {
	const maxSamples = 2048
	xi := make([]int16, maxSamples)
	xq := make([]int16, maxSamples)
	inter := NewInterleaveFn()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		inter(xi, xq)
	}
}
