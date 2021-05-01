// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package duo

// NewInterleave creates a function that returns a slice with the
// provided samples interleaved. It uses a persistent buffer to avoid
// extra allocations.
func NewInterleave() func(xia, xqa, xib, xqb []int16) []int16 {
	const scalarsPerFrame = 4
	buf := make([]int16, 10*1024)
	return func(xia, xqa, xib, xqb []int16) []int16 {
		numScalars := len(xia) * scalarsPerFrame
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
		for i := range xia {
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
