// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package callback

// NewInterleave creates a function that returns a slice with the
// provided samples interleaved. It uses a persistent buffer to avoid
// extra allocations.
func NewInterleave() func(xi, xq []int16) []int16 {
	const scalarsPerFrame = 2
	buf := make([]int16, 4096)
	return func(xi, xq []int16) []int16 {
		numScalars := len(xi) * scalarsPerFrame
		if len(buf) < numScalars {
			next := len(buf) * 2
			if next < numScalars {
				next = numScalars
			}
			buf = make([]int16, next)
		}
		var bi int
		for i := range xi {
			buf[bi] = xi[i]
			buf[bi+1] = xq[i]
			bi += scalarsPerFrame
		}
		return buf[:numScalars]
	}
}
