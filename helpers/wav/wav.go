// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package wav

import "fmt"

type RiffChunk struct {
	ChunkId   [4]byte
	ChunkSize uint32
	Format    [4]byte
}

type FmtChunk struct {
	ChunkId       [4]byte
	ChunkSize     uint32
	AudioFormat   uint16
	NumChannels   uint16
	SampleRate    uint32
	ByteRate      uint32
	BlockAlign    uint16
	BitsPerSample uint16
	ExtSize       uint16 // required to support floating-point
}

type FactChunk struct {
	ChunkId      [4]byte
	ChunkSize    uint32
	SampleLength uint32
}

type DataChunk struct {
	ChunkId   [4]byte
	ChunkSize uint32
	// samples follow this chunk
}

type Header struct {
	Riff RiffChunk
	Fmt  FmtChunk
	Fact FactChunk // required for floating-point
	Data DataChunk
	// samples follow the header data chunk
}

type SampleFormat uint16

const (
	LPCM              SampleFormat = 1
	IEEEFloatingPoint SampleFormat = 3
)

// NewHeader creates and initializes a new WAV Header struct. The header
// can be written using encoding/binary.Write().
//
// If bigEndian is true, then the "RIFX" chunk ID is used instead of "RIFF".
// This option should only be used if all samples will be written using
// big-endian byte-order (e.g. binary.BigEndian). For the best compatibility
// little-endian encoding should be preferred, but big-endian encoding
// may be advantageous if fast writes on a big-endian CPU are required.
//
// The numFrames parameter should be the number of frames written or
// to be written. A value of zero can be used initially and the header
// struct can be updated with the correct size later with the Update method.
func NewHeader(
	sampleRate uint32, numChannels uint16, bytesPerSample uint8,
	format SampleFormat, bigEndian bool, numFrames uint32,
) (*Header, error) {
	head := Header{}
	dataBytes := numFrames * (uint32(bytesPerSample) * uint32(numChannels))

	// RIFF header
	switch bigEndian {
	case true:
		head.Riff.ChunkId = [4]byte{'R', 'I', 'F', 'X'}
	default:
		head.Riff.ChunkId = [4]byte{'R', 'I', 'F', 'F'}
	}
	head.Riff.ChunkSize = 4 + dataBytes
	head.Riff.Format = [4]byte{'W', 'A', 'V', 'E'}

	// fmt header
	head.Fmt.ChunkId = [4]byte{'f', 'm', 't', ' '}
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
	head.Fact.ChunkId = [4]byte{'f', 'a', 'c', 't'}
	head.Fact.ChunkSize = 4
	head.Fact.SampleLength = numFrames

	// data header
	head.Data.ChunkId = [4]byte{'d', 'a', 't', 'a'}
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
