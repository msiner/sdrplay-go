// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package callback

// InterleaveFn is a function type that returns a slice with the
// provided component signal sample scalars interleaved into a single
// slice. The xi slice contains the real component and the xq slice
// contains the imaginary component. The lengths of xi and xq should be
// equal. The length of the resulting slice is twice the length of the
// shortest of the lengths of xi and xq.
type InterleaveFn func(xi, xq []int16) []int16

// NewInterleaveFn creates a new InterleaveFn.
//
// The function uses an internal persistent buffer to minimize allocations.
// The returned slice is a slice of that internal buffer and should not be
// modified or stored.
func NewInterleaveFn() InterleaveFn {
	const scalarsPerFrame = 2
	buf := make([]int16, 4096)
	return func(xi, xq []int16) []int16 {
		minLen := len(xi)
		if len(xq) < minLen {
			minLen = len(xq)
		}
		numScalars := minLen * scalarsPerFrame
		if len(buf) < numScalars {
			next := len(buf) * 2
			if next < numScalars {
				next = numScalars
			}
			buf = make([]int16, next)
		}
		switch len(xi) >= len(xq) {
		case true:
			var bi int
			for i := range xq {
				buf[bi] = xi[i]
				buf[bi+1] = xq[i]
				bi += scalarsPerFrame
			}
		default:
			var bi int
			for i := range xi {
				buf[bi] = xi[i]
				buf[bi+1] = xq[i]
				bi += scalarsPerFrame
			}
		}
		return buf[:numScalars]
	}
}
