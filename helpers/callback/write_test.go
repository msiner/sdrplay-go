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

func testByteOrders(f func(order binary.ByteOrder)) {
	type CustomOrder struct {
		binary.ByteOrder
	}
	f(binary.LittleEndian)
	f(binary.BigEndian)
	f(CustomOrder{binary.BigEndian})
}

func TestWrite(t *testing.T) {
	t.Parallel()

	testByteOrders(func(order binary.ByteOrder) {
		write := NewWrite(order)

		for i := 0; i < 100; i++ {
			samples := make([]int16, rand.Int31n(100000))
			for j := range samples {
				samples[j] = int16(rand.Int())
			}
			buf := bytes.NewBuffer(nil)
			binary.Write(buf, order, samples)
			want := buf.Bytes()
			buf.Reset()

			n, err := write(buf, samples)
			if err != nil {
				t.Fatal(err)
			}
			if n != len(want) {
				t.Fatalf("wrong number of bytes from write; got %d, want %d", n, len(want))
			}
			got := buf.Bytes()
			if !bytes.Equal(got, want) {
				t.Errorf("wrong bytes after write; got %v, want %v", got, want)
			}
		}
	})
}

func TestFloat32Write(t *testing.T) {
	t.Parallel()

	testByteOrders(func(order binary.ByteOrder) {
		write := NewFloat32Write(order)

		for i := 0; i < 100; i++ {
			samples := make([]float32, rand.Int31n(100000))
			for j := range samples {
				samples[j] = rand.Float32()
			}
			buf := bytes.NewBuffer(nil)
			binary.Write(buf, order, samples)
			want := buf.Bytes()
			buf.Reset()

			n, err := write(buf, samples)
			if err != nil {
				t.Fatal(err)
			}
			if n != len(want) {
				t.Fatalf("wrong number of bytes from write; got %d, want %d", n, len(want))
			}
			got := buf.Bytes()
			if !bytes.Equal(got, want) {
				t.Errorf("wrong bytes after write; got %v, want %v", got, want)
			}
		}
	})
}

func TestComplex64Write(t *testing.T) {
	t.Parallel()

	testByteOrders(func(order binary.ByteOrder) {
		write := NewComplex64Write(order)

		for i := 0; i < 100; i++ {
			samples := make([]complex64, rand.Int31n(100000))
			for j := range samples {
				samples[j] = complex(rand.Float32(), rand.Float32())
			}
			buf := bytes.NewBuffer(nil)
			binary.Write(buf, order, samples)
			want := buf.Bytes()
			buf.Reset()

			n, err := write(buf, samples)
			if err != nil {
				t.Fatal(err)
			}
			if n != len(want) {
				t.Fatalf("wrong number of bytes from write; got %d, want %d", n, len(want))
			}
			got := buf.Bytes()
			if !bytes.Equal(got, want) {
				t.Errorf("wrong bytes after write; got %v, want %v", got, want)
			}
		}
	})
}

func BenchmarkWrite(b *testing.B) {
	x := make([]int16, 2048)
	write := NewWrite(binary.LittleEndian)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		write(ioutil.Discard, x)
	}
}
