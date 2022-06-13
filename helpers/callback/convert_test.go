// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package callback

import (
	"math/rand"
	"testing"
)

func TestConvertToFloat32(t *testing.T) {
	t.Parallel()

	convert := NewConvertToFloat32Fn(17)
	for i := 0; i < 100; i++ {
		samples := make([]int16, rand.Int31n(20000))
		for j := range samples {
			samples[j] = int16(j)
		}
		floats := convert(samples)
		if len(floats) != len(samples) {
			t.Fatalf("float slice has wrong length: got %d, want %d", len(floats), len(samples))
		}
		var last float32
		for j := range samples {
			curr := floats[j]
			if j != 0 {
				if curr <= last {
					t.Fatalf("floats are not monotonically increasing: curr=%f, last=%f", curr, last)
				}
			}
			last = curr
			if curr < -1 || curr > 1 {
				t.Fatalf("value outside of range: got %f, want [-1,1]", curr)
			}
		}
	}
}

func TestConvertToComplex64(t *testing.T) {
	t.Parallel()

	convert := NewConvertToComplex64Fn(17)
	for i := 0; i < 100; i++ {
		xi := make([]int16, rand.Int31n(20000))
		xq := make([]int16, len(xi))
		for j := range xi {
			xi[j] = int16(j)
			xq[j] = -int16(j)
		}
		cx := convert(xi, xq)
		if len(cx) != len(xi) {
			t.Fatalf("float slice has wrong length: got %d, want %d", len(cx), len(xi))
		}
		var last complex64
		for j := range cx {
			curr := cx[j]
			if j != 0 {
				if real(curr) <= real(last) {
					t.Fatalf("floats are not monotonically increasing: curr=%f, last=%f", curr, last)
				}
			}
			last = curr
			if real(curr) < -1 || real(curr) > 1 {
				t.Fatalf("real value outside of range: got %f, want [-1,1]", real(curr))
			}
			if imag(curr) < -1 || imag(curr) > 1 {
				t.Fatalf("imag value outside of range: got %f, want [-1,1]", imag(curr))
			}
			if (real(curr) + imag(curr)) > 0.00001 {
				t.Fatalf("imag not negative of real: %v", curr)
			}
		}

		x := convert(xi, xq[:1])
		if len(x) != 1 {
			t.Errorf("wrong length on unbalanced convert: got %d, want 1", len(x))
		}
		x = convert(xi[:2], xq)
		if len(x) != 2 {
			t.Errorf("wrong length on unbalanced convert: got %d, want 2", len(x))
		}
	}
}

func BenchmarkConvertToFloat32(b *testing.B) {
	x := make([]int16, 4096)
	conv := NewConvertToFloat32Fn(14)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		conv(x)
	}
}

func BenchmarkConvertToComplex64(b *testing.B) {
	xi := make([]int16, 2048)
	xq := make([]int16, 2048)
	conv := NewConvertToComplex64Fn(14)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		conv(xi, xq)
	}
}
