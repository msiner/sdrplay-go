// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package callback

import "math"

// ConvertToFloat32Fn is a function type that returns a slice with the
// provided sample scalars converted to float32.
type ConvertToFloat32Fn func(x []int16) []float32

// NewConvertToFloat32Fn creates a new ConvertToFloat32Fn.
//
// The numBits argument determines the scaling factor. When numBits is 16 or
// greater, the scaling will convert all values in the int16 domain to
// floating point values in the [-1,1] range. Any smaller number assumes a
// smaller potential domain, but does not enforce it. That is, an input
// sample that uses bits outside the specified range, will produce a value
// outside of the [-1,1] range.
//
// The function uses an internal persistent buffer to minimize allocations.
// The returned slice is a slice of that internal buffer and should not be
// modified or stored.
func NewConvertToFloat32Fn(numBits uint) ConvertToFloat32Fn {
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

// ConvertToComplex64Fn is a function type that returns a slice with the
// provided indepedent signal component sample scalars converted to complex64.
// The xi slice contains the real component and the xq slice contains the
// imaginary component. The lengths of xi and xq should be equal. The length
// of the resulting slice is twice the length of the shortest of the lengths
// of xi and xq.
type ConvertToComplex64Fn func(xi, xq []int16) []complex64

// NewConvertToComplex64Fn creates a new ConvertToComplex64Fn.
//
// The numBits argument determines the scaling factor. When numBits is 16 or
// greater, the scaling will convert all values in the int16 domain to
// floating point values in the [-1,1] range. Any smaller number assumes a
// smaller potential domain, but does not enforce it. That is, an input
// sample that uses bits outside the specified range, will produce a value
// outside of the [-1,1] range.
//
// The function uses an internal persistent buffer to minimize allocations.
// The returned slice is a slice of that internal buffer and should not be
// modified or stored.
func NewConvertToComplex64Fn(numBits uint) ConvertToComplex64Fn {
	if numBits > 16 {
		numBits = 16
	}
	maxMag := float32(math.Pow(2, float64(numBits-1)))
	buf := make([]complex64, 2048)
	return func(xi, xq []int16) []complex64 {
		minLen := len(xi)
		if len(xq) < minLen {
			minLen = len(xq)
		}
		if len(buf) < minLen {
			next := len(buf) * 2
			if next < minLen {
				next = minLen
			}
			buf = make([]complex64, next)
		}
		switch len(xi) >= len(xq) {
		case true:
			for i := range xq {
				buf[i] = complex(
					float32(xi[i])/maxMag,
					float32(xq[i])/maxMag,
				)
			}
		default:
			for i := range xi {
				buf[i] = complex(
					float32(xi[i])/maxMag,
					float32(xq[i])/maxMag,
				)
			}
		}
		return buf[:minLen]
	}
}
