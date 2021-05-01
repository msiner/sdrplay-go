// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package callback

import "math"

// NewConvertToFloat32 creates a function that returns a slice with the
// provided scalars converted to float32. The numBits determines the
// scaling factor.
func NewConvertToFloat32(numBits uint) func(x []int16) []float32 {
	if numBits > 16 {
		numBits = 16
	}
	maxMag := float32(math.Pow(2, float64(numBits-1)))
	buf := make([]float32, 4096)
	return func(x []int16) []float32 {
		if len(buf) < len(x) {
			next := len(buf) * 2
			if next < len(x) {
				next = len(x)
			}
			buf = make([]float32, next)
		}
		for i := range x {
			buf[i] = float32(x[i]) / maxMag
		}
		return buf[:len(x)]
	}
}

// NewConvertToComplex64 creates a function that returns a slice with the
// provided samples converted to complex64. The numBits determines the
// scaling factor.
func NewConvertToComplex64(numBits uint) func(xi, xq []int16) []complex64 {
	if numBits > 16 {
		numBits = 16
	}
	maxMag := float32(math.Pow(2, float64(numBits-1)))
	buf := make([]complex64, 2048)
	return func(xi, xq []int16) []complex64 {
		if len(buf) < len(xi) {
			next := len(buf) * 2
			if next < len(xi) {
				next = len(xi)
			}
			buf = make([]complex64, next)
		}
		for i := range xi {
			buf[i] = complex(
				float32(xi[i])/maxMag,
				float32(xq[i])/maxMag,
			)
		}
		return buf[:len(xi)]
	}
}
