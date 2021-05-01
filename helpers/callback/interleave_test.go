// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package callback

import (
	"testing"
)

func BenchmarkInterleave(b *testing.B) {
	xi := make([]int16, 2048)
	xq := make([]int16, 2048)
	inter := NewInterleave()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		inter(xi, xq)
	}
}
