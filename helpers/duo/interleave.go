// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package duo

// InterleaveFn is a function type that returns a slice with the
// provided component signal sample scalars interleaved into a single
// slice. The xia slice contains the real component of stream A, the
// xqa slice contains the imaginary component of stream A, the xib
// slice contains the real component of stream B, the xqa slice contains
// the imaginary component of stream B. The lengths of xia, xqa, xib,
// and xqb should be equal. The length of the resulting slice is four
// times the length of the shortest of the lengths of all provided slices.
type InterleaveFn func(xia, xqa, xib, xqb []int16) []int16

// NewInterleaveFn creates a new InterleaveFn.
//
// The function uses an internal persistent buffer to minimize allocations.
// The returned slice is a slice of that internal buffer and should not be
// modified or stored.
func NewInterleaveFn() InterleaveFn {
	const scalarsPerFrame = 4
	buf := make([]int16, 10*1024)
	return func(xia, xqa, xib, xqb []int16) []int16 {
		minLen := len(xia)
		if len(xqa) < minLen {
			minLen = len(xqa)
		}
		if len(xib) < minLen {
			minLen = len(xib)
		}
		if len(xqb) < minLen {
			minLen = len(xqb)
		}
		numScalars := minLen * scalarsPerFrame
		if len(buf) < numScalars {
			next := len(buf) * 2
			if next < numScalars {
				next = numScalars
			}
			buf = make([]int16, next)
		}
		var (
			bi    int
			frame []int16
		)
		for i := 0; i < minLen; i++ {
			// The extra work here is made up for by allowing
			// the compiler to eliminate some bounds checks because
			// it knows the length of frame and all indexes into
			// frame are constants.
			frame = buf[bi : bi+scalarsPerFrame]
			frame[0] = xia[i]
			frame[1] = xqa[i]
			frame[2] = xib[i]
			frame[3] = xqb[i]
			bi += scalarsPerFrame
		}
		return buf[:numScalars]
	}
}
