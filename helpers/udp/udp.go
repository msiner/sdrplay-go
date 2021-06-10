// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package udp

import (
	"encoding/binary"
	"fmt"
	"io"
)

// UDPPacketWriteFn is a function type that writes the provided samples to
// a packet buffer and, when the packet is filled, writes the packet to the
// provided io.Writer. It returns the number of bytes written to the packet
// buffer. The function will only write to the io.Writer when a full packet
// payload is ready. Depending on the state of the buffer and the size of
// the input slice, it may write zero, one, or multiple packets to the
// io.Writer.
//
// payloadLen is the number of bytes in each payload.
//
// scalarsPerFrame is the number of scalars per frame. That is, interleaved
// complex components from a single channel is 2 scalars per frame. A single
// component buffer, I or Q, would be 1 scalar per frame.
//
// seqHeader is true to insert a 64-bit unsigned sequence number at the
// beginning of each packet.
//
// bigEndian is true to encode all data, headers and scalars, as big-endian
// and false for little-endian.
type PacketWriteFn func(out io.Writer, x []int16) (int, error)

// NewUDPPacketWrite creates a new UDPPacketWriteFn.
func NewPacketWriteFn(payloadLen, scalarsPerFrame uint, seqHeader bool, order binary.ByteOrder) (PacketWriteFn, error) {
	const (
		sizeofScalar = 2
		sizeofHeader = 8
	)
	dataBytes := payloadLen
	if seqHeader {
		dataBytes -= sizeofHeader
	}
	if dataBytes%scalarsPerFrame != 0 {
		return nil, fmt.Errorf(
			"frames will not fit evenly in payload; payloadLen=%d seqHeader=%v scalarsPerFrame=%d",
			payloadLen, seqHeader, scalarsPerFrame,
		)
	}

	var (
		seq uint64
		buf = make([]byte, int(payloadLen))
		bi  int
	)

	if seqHeader {
		// buf is already zero-initialized, so seq zero is already encoded.
		// Just move the buf index to skip it.
		seq++
		bi = sizeofHeader
	}
	write := func(out io.Writer, x []int16) (int, error) {
		if len(x)%int(scalarsPerFrame) != 0 {
			return 0, fmt.Errorf("invalid number of scalars; got %d, want multiple of %d", len(x), scalarsPerFrame)
		}
		var total int
		switch order {
		case binary.BigEndian:
			for i := range x {
				binary.BigEndian.PutUint16(buf[bi:], uint16(x[i]))
				total += sizeofScalar
				bi += sizeofScalar
				if bi == int(payloadLen) {
					_, err := out.Write(buf)
					if err != nil {
						return total, err
					}
					bi = 0
					if seqHeader {
						binary.BigEndian.PutUint64(buf, seq)
						bi += sizeofHeader
						seq++
					}
				}
			}
		case binary.LittleEndian:
			for i := range x {
				binary.LittleEndian.PutUint16(buf[bi:], uint16(x[i]))
				total += sizeofScalar
				bi += sizeofScalar
				if bi == int(payloadLen) {
					_, err := out.Write(buf)
					if err != nil {
						return total, err
					}
					bi = 0
					if seqHeader {
						binary.LittleEndian.PutUint64(buf, seq)
						bi += sizeofHeader
						seq++
					}
				}
			}
		default:
			for i := range x {
				order.PutUint16(buf[bi:], uint16(x[i]))
				total += sizeofScalar
				bi += sizeofScalar
				if bi == int(payloadLen) {
					_, err := out.Write(buf)
					if err != nil {
						return total, err
					}
					bi = 0
					if seqHeader {
						order.PutUint64(buf, seq)
						bi += sizeofHeader
						seq++
					}
				}
			}
		}
		return total, nil
	}

	return write, nil
}
