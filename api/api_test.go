// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package api

import "testing"

func TestBytesToString(t *testing.T) {
	t.Parallel()

	specs := []struct {
		b []byte
		s string
	}{
		{nil, ""},
		{[]byte{}, ""},
		{[]byte(""), ""},
		{[]byte{'a'}, "a"},
		{[]byte{'a', 0, 'b'}, "a"},
		{[]byte{0, 'a', 0, 'b'}, ""},
		{[]byte("Hello, 世界"), "Hello, 世界"},
		{append(append([]byte("世"), 'a', 0), []byte("界")...), "世a"},
	}

	for _, spec := range specs {
		got := bytesToString(spec.b)
		if got != spec.s {
			t.Errorf("bad conversion: got %s, want %s", got, spec.s)
		}
	}
}

func BenchmarkBytesToString(b *testing.B) {
	buf := make([]byte, 1024)
	for i := range buf {
		buf[i] = 'a'
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		s := bytesToString(buf)
		if len(s) != 1024 {
			b.Fatalf("wrong string length: got %d, want 1024", len(s))
		}
	}
}
