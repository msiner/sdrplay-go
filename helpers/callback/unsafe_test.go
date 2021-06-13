// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package callback

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
	"math/rand"
	"testing"
)

func TestFastWrite(t *testing.T) {
	t.Parallel()

	b := bytes.NewBuffer(nil)

	n, err := FastWrite(b, nil)
	if err != nil {
		t.Fatalf("unexpected error; %v", err)
	}
	if n != 0 {
		t.Fatalf("wrong number of bytes written; got %d, want 0", n)
	}

	n, err = FastWrite(b, []int16{})
	if err != nil {
		t.Fatalf("unexpected error; %v", err)
	}
	if n != 0 {
		t.Fatalf("wrong number of bytes written; got %d, want 0", n)
	}

	x := make([]int16, 2048)
	xs := binary.Size(x)
	for i := range x {
		x[i] = int16(rand.Int())
	}
	n, err = FastWrite(b, x)
	if err != nil {
		t.Fatalf("unexpected error; %v", err)
	}
	if n != xs {
		t.Fatalf("wrong number of bytes written; got %d, want %d", n, xs)
	}
	buf := b.Bytes()
	var order binary.ByteOrder = binary.LittleEndian
	if x[0] == int16(binary.BigEndian.Uint16(buf)) {
		order = binary.BigEndian
	}
	bi := 0
	for i := range x {
		val := int16(order.Uint16(buf[bi:]))
		if val != x[i] {
			t.Fatalf("wrong value at %d; got %d, want %d", i, val, x[i])
		}
		bi += 2
	}
}

func TestFastWriteKeepAlive(t *testing.T) {
	t.Parallel()

	b := bytes.NewBuffer(nil)
	n, err := FastWrite(b, []int16{0x4242, 0x4343})
	if err != nil {
		t.Fatalf("unexpected error; %v", err)
	}
	if n != 4 {
		t.Fatalf("wrong number of bytes written; got %d, want 4", n)
	}
	buf := b.Bytes()
	if buf[0] != 0x42 || buf[1] != 0x42 {
		t.Errorf("wrong bytes in first word; got %v, want 0x4242", buf[:2])
	}
	if buf[2] != 0x43 || buf[3] != 0x43 {
		t.Errorf("wrong bytes in second word; got %v, want 0x4343", buf[2:])
	}
}

func TestFastWriteBig(t *testing.T) {
	const (
		numScalars = 10 * 1024 * 1024
		numBytes   = numScalars * 2
	)
	x := make([]int16, numScalars)
	x[numScalars-1] = 0x2323
	b := bytes.NewBuffer(nil)
	n, err := FastWrite(b, x)
	if err != nil {
		t.Fatalf("unexpected error; %v", err)
	}
	if n != numBytes {
		t.Fatalf("wrong number of bytes written; got %d, want %d", n, numBytes)
	}
	buf := b.Bytes()
	if buf[0] != 0 || buf[1] != 0 {
		t.Errorf("wrong bytes in first word; got %v, want 0x0000", buf[:2])
	}
	if buf[numBytes-2] != 0x23 || buf[numBytes-1] != 0x23 {
		t.Errorf("wrong bytes in last word; got %v, want 0x2323", buf[numBytes-2:])
	}
}

func BenchmarkFastWrite(b *testing.B) {
	x := make([]int16, 2048)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, _ = FastWrite(ioutil.Discard, x)
	}
}
