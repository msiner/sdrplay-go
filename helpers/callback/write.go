// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package callback

import (
	"encoding/binary"
	"io"
	"math"
)

// NewWrite creates a function that writes the provided samples to the
// provided io.Writer. The function is roughly equivalent to binary.Write()
// except for some application specific optimizations. The function uses a
// persistent buffer to avoid allocations.
func NewWrite(order binary.ByteOrder) func(out io.Writer, x []int16) (int, error) {
	const sizeOfScalar = 2
	buf := make([]byte, 4096)
	return func(out io.Writer, x []int16) (int, error) {
		numBytes := len(x) * sizeOfScalar
		if len(buf) < numBytes {
			next := len(buf) * sizeOfScalar
			if next < numBytes {
				next = numBytes
			}
			buf = make([]byte, next)
		}
		switch order {
		case binary.LittleEndian:
			bi := 0
			for i := range x {
				binary.LittleEndian.PutUint16(buf[bi:], uint16(x[i]))
				bi += sizeOfScalar
			}
		case binary.BigEndian:
			bi := 0
			for i := range x {
				binary.BigEndian.PutUint16(buf[bi:], uint16(x[i]))
				bi += sizeOfScalar
			}
		default:
			bi := 0
			for i := range x {
				order.PutUint16(buf[bi:], uint16(x[i]))
				bi += sizeOfScalar
			}
		}
		return out.Write(buf[:numBytes])
	}
}

// NewFloat32Write creates a function that writes the provided samples to the
// provided io.Writer. The function is roughly equivalent to binary.Write()
// except for some application specific optimizations. The function uses a
// persistent buffer to avoid allocations.
func NewFloat32Write(order binary.ByteOrder) func(out io.Writer, x []float32) (int, error) {
	const sizeOfScalar = 4
	buf := make([]byte, 4096)
	return func(out io.Writer, x []float32) (int, error) {
		numBytes := len(x) * sizeOfScalar
		if len(buf) < numBytes {
			next := len(buf) * 2
			if next < numBytes {
				next = numBytes
			}
			buf = make([]byte, next)
		}
		switch order {
		case binary.LittleEndian:
			bi := 0
			for i := range x {
				binary.LittleEndian.PutUint32(buf[bi:], math.Float32bits(x[i]))
				bi += sizeOfScalar
			}
		case binary.BigEndian:
			bi := 0
			for i := range x {
				binary.BigEndian.PutUint32(buf[bi:], math.Float32bits(x[i]))
				bi += sizeOfScalar
			}
		default:
			bi := 0
			for i := range x {
				order.PutUint32(buf[bi:], math.Float32bits(x[i]))
				bi += sizeOfScalar
			}
		}
		return out.Write(buf[:numBytes])
	}
}

// NewComplex64Write creates a function that writes the provided samples to the
// provided io.Writer. The function is roughly equivalent to binary.Write()
// except for some application specific optimizations. The function uses a
// persistent buffer to avoid allocations. The big parameter is true for
// big-endian and false for little-endian. It is specified that way instead
// of accepting a ByteOrder interface to allow inlining direct access to
// binary.BigEndian and binary.LittleEndian.
func NewComplex64Write(order binary.ByteOrder) func(out io.Writer, x []complex64) (int, error) {
	const (
		sizeOfScalar    = 4
		scalarsPerFrame = 2
		sizeOfFrame     = sizeOfScalar * scalarsPerFrame
	)
	buf := make([]byte, 4096)
	return func(out io.Writer, x []complex64) (int, error) {
		numBytes := len(x) * sizeOfFrame
		if len(buf) < numBytes {
			next := len(buf) * 2
			if next < numBytes {
				next = numBytes
			}
			buf = make([]byte, next)
		}
		switch order {
		case binary.BigEndian:
			bi := 0
			for i := range x {
				binary.BigEndian.PutUint32(buf[bi:], math.Float32bits(real(x[i])))
				binary.BigEndian.PutUint32(buf[bi+sizeOfScalar:], math.Float32bits(imag(x[i])))
				bi += sizeOfFrame
			}
		case binary.LittleEndian:
			bi := 0
			for i := range x {
				binary.LittleEndian.PutUint32(buf[bi:], math.Float32bits(real(x[i])))
				binary.LittleEndian.PutUint32(buf[bi+sizeOfScalar:], math.Float32bits(imag(x[i])))
				bi += sizeOfFrame
			}
		default:
			bi := 0
			for i := range x {
				order.PutUint32(buf[bi:], math.Float32bits(real(x[i])))
				order.PutUint32(buf[bi+sizeOfScalar:], math.Float32bits(imag(x[i])))
				bi += sizeOfFrame
			}
		}
		return out.Write(buf[:numBytes])
	}
}
