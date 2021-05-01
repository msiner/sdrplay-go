// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package duo

import (
	"testing"
)

func BenchmarkInterleave(b *testing.B) {
	xia := make([]int16, 2048)
	xqa := make([]int16, 2048)
	xib := make([]int16, 2048)
	xqb := make([]int16, 2048)
	inter := NewInterleave()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		inter(xia, xqa, xib, xqb)
	}
}
