// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package callback

import (
	"testing"
)

func BenchmarkConvertToFloat32(b *testing.B) {
	x := make([]int16, 4096)
	conv := NewConvertToFloat32(14)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		conv(x)
	}
}

func BenchmarkConvertToComplex64(b *testing.B) {
	xi := make([]int16, 2048)
	xq := make([]int16, 2048)
	conv := NewConvertToComplex64(14)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		conv(xi, xq)
	}
}
