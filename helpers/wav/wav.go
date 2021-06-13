// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package wav

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

// RiffChunk represents the first chunk of a WAV file.
type RiffChunk struct {
	ChunkID   [4]byte
	ChunkSize uint32
	Format    [4]byte
}

// FmtChunk represents a fmt chunk in a header file.
type FmtChunk struct {
	ChunkID       [4]byte
	ChunkSize     uint32
	AudioFormat   uint16
	NumChannels   uint16
	SampleRate    uint32
	ByteRate      uint32
	BlockAlign    uint16
	BitsPerSample uint16
	ExtSize       uint16 // required to support floating-point
}

// FactChunk represents a WAV file "fact" chunk. This chunk is
// required for WAV files using a non-PCM format. Since we support
// IEEE floating point sample data, we just include it in every
// header.
type FactChunk struct {
	ChunkID      [4]byte
	ChunkSize    uint32
	SampleLength uint32
}

// DataChunk represets the WAV file "data" chunk. All samples that
// follow the "data" chunk header fields in DataChunk are also part
// of the "data" chunk. The samples are not represented as a field
// because the structs in this package are designed to be serialize
// via encoding/binary to create a binary block of data that can
// be inserted before sample data.
type DataChunk struct {
	ChunkID   [4]byte
	ChunkSize uint32
	// samples follow this chunk
}

// Header represents a WAV header that is the collection of all chunks
// before "data" and the header or meta-data portions of the "data"
// chunk. When serialized with encoding/binary, it creates a valid
// beginning to a WAV file.
type Header struct {
	Riff RiffChunk
	Fmt  FmtChunk
	Fact FactChunk // required for floating-point
	Data DataChunk
	// samples follow the header data chunk
}

// SampleFormat is an enum type that represents the two sample formats
// supported by this package: L-PCM and IEEE floating-point.
type SampleFormat uint16

const (
	// LPCM is used to specify a linear PCM sample format.
	LPCM SampleFormat = 1
	// IEEEFloatingPoint is used to specify a 32-bit or 64-bit
	// IEEE floating-point sample format.
	IEEEFloatingPoint SampleFormat = 3
)

// NewHeader creates and initializes a new WAV Header struct. The header
// can be written using encoding/binary.Write().
//
// The order argument should be one of binary.BigEndian or
// binary.LittleEndian. If it is BigEndian then the "RIFX" chunk ID is
// used instead of "RIFF". This option should only be used if all samples
// will be written using big-endian byte-order (e.g. binary.BigEndian).
// For the best compatibility with WAV-supporting software, little-endian
// encoding should be preferred, but big-endian encoding may be
// advantageous if fast writes on a big-endian CPU are required.
//
// The numFrames parameter should be the number of frames written or
// to be written. A value of zero can be used initially and the header
// struct can be updated with the correct size later with the Update method.
func NewHeader(
	sampleRate uint32, numChannels uint16, bytesPerSample uint8,
	format SampleFormat, order binary.ByteOrder, numFrames uint32,
) (*Header, error) {
	head := Header{}
	dataBytes := numFrames * (uint32(bytesPerSample) * uint32(numChannels))

	isBig := false
	switch order {
	case binary.BigEndian:
		isBig = true
	case binary.LittleEndian:
		isBig = false
	default:
		var bb, bl, bo [4]byte
		val := uint32(0x01234567)
		binary.BigEndian.PutUint32(bb[:], val)
		binary.LittleEndian.PutUint32(bl[:], val)
		order.PutUint32(bo[:], val)
		switch {
		case bytes.Equal(bb[:], bo[:]):
			isBig = true
		case bytes.Equal(bl[:], bo[:]):
			isBig = false
		default:
			return nil, errors.New("provided ByteOrder must be big or little endian")
		}
	}
	// RIFF header
	switch isBig {
	case true:
		head.Riff.ChunkID = [4]byte{'R', 'I', 'F', 'X'}
	default:
		head.Riff.ChunkID = [4]byte{'R', 'I', 'F', 'F'}
	}
	head.Riff.ChunkSize = 4 + dataBytes
	head.Riff.Format = [4]byte{'W', 'A', 'V', 'E'}

	// fmt header
	head.Fmt.ChunkID = [4]byte{'f', 'm', 't', ' '}
	head.Fmt.ChunkSize = 18
	switch format {
	case IEEEFloatingPoint:
		head.Fmt.AudioFormat = uint16(IEEEFloatingPoint)
		switch bytesPerSample {
		case 4, 8:
			// Good
		default:
			return nil, fmt.Errorf("invalid bytes per sample for floating point format; got %d, want 4 or 8", bytesPerSample)
		}
	case LPCM:
		head.Fmt.AudioFormat = uint16(LPCM)
		switch bytesPerSample {
		case 1, 2, 3, 4:
			// Good
		default:
			return nil, fmt.Errorf("invalid bytes per sample for PCM format; got %d, want 1, 2, 3, or 4", bytesPerSample)
		}
	default:
		return nil, fmt.Errorf("invalid sample format; got %d, want LPCM or IEEEFloatingPoint", format)
	}
	head.Fmt.NumChannels = numChannels
	head.Fmt.SampleRate = sampleRate
	head.Fmt.ByteRate = sampleRate * uint32(numChannels) * uint32(bytesPerSample)
	head.Fmt.BlockAlign = numChannels * uint16(bytesPerSample)
	head.Fmt.BitsPerSample = uint16(bytesPerSample * 8)
	head.Fmt.ExtSize = 0 // required for floating-point

	// fact header
	head.Fact.ChunkID = [4]byte{'f', 'a', 'c', 't'}
	head.Fact.ChunkSize = 4
	head.Fact.SampleLength = numFrames

	// data header
	head.Data.ChunkID = [4]byte{'d', 'a', 't', 'a'}
	head.Data.ChunkSize = dataBytes

	return &head, nil
}

// Update sets all of the data size dependent fields in the
// header struct with a new value reflecting a new total number
// of frames. Note that updates do not accumulate.
func (h *Header) Update(numFrames uint32) {
	bytesPerFrame := h.Fmt.BitsPerSample / 8 * h.Fmt.NumChannels
	numBytes := uint32(bytesPerFrame) * numFrames
	h.Riff.ChunkSize = 4 + numBytes
	h.Fact.SampleLength = numFrames
	h.Data.ChunkSize = numBytes
}
