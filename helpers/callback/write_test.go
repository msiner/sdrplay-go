// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package callback

import (
	"io/ioutil"
	"testing"
)

func BenchmarkWrite(b *testing.B) {
	x := make([]int16, 2048)
	write := NewWrite(false)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		write(ioutil.Discard, x)
	}
}
