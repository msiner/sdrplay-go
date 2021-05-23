// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package callback

import (
	"encoding/binary"
	"io"
	"math"
)

// WriteFn is a function type that writes the provided samples to the
// specified io.Writer. It returns the number of bytes written and a non-nil
// error if an error is encountered during write.
type WriteFn func(out io.Writer, x []int16) (int, error)

// NewWriteFn creates a new WriteFn that writes samples using the provided
// ByteOrder. The function uses an internal persistent buffer to avoid
// allocations. In comparison, calling encoding/binary on a slice will
// allocate a new []byte on each call.
func NewWriteFn(order binary.ByteOrder) WriteFn {
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

// Float32WriteFn is a function type that writes the provided samples to the
// specified io.Writer. It returns the number of bytes written and a non-nil
// error if an error is encountered during write.
type Float32WriteFn func(out io.Writer, x []float32) (int, error)

// NewFloat32WriteFn creates a new Float32WriteFn that writes samples using
// the provided ByteOrder. The function uses an internal persistent buffer
// to avoid allocations. In comparison, calling encoding/binary on a slice
// will allocate a new []byte on each call.
func NewFloat32WriteFn(order binary.ByteOrder) Float32WriteFn {
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

// Complex64WriteFn is a function type that writes the provided samples to the
// specified io.Writer. It returns the number of bytes written and a non-nil
// error if an error is encountered during write.
type Complex64WriteFn func(out io.Writer, x []complex64) (int, error)

// NewComplex64WriteFn creates a new Complex64WriteFn that writes samples using
// the provided ByteOrder. The function uses an internal persistent buffer
// to avoid allocations. In comparison, calling encoding/binary on a slice
// will allocate a new []byte on each call.
func NewComplex64WriteFn(order binary.ByteOrder) Complex64WriteFn {
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
