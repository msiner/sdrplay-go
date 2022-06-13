// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package wav

import (
	"bytes"
	"encoding/binary"
	"strings"
	"testing"
)

func TestHeader(t *testing.T) {
	h, err := NewHeader(20000, 1, 2, LPCM, binary.LittleEndian, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	w := bytes.NewBuffer(nil)
	if err := binary.Write(w, binary.LittleEndian, h); err != nil {
		t.Fatalf("failed to write header: %v", err)
	}
	b := w.Bytes()
	sizeOfHeader := binary.Size(h)
	if len(b) != sizeOfHeader {
		t.Fatalf("wrong number of header bytes written: got %d, want %d", len(b), sizeOfHeader)
	}
	magic := string(b[:4])
	if magic != "RIFF" {
		t.Fatalf("wrong magic number in header bytes: got %s, want RIFF", magic)
	}
	if h.Data.ChunkSize != 0 {
		t.Fatalf("wrong data chunk size: got %d, want 0", h.Data.ChunkSize)
	}
	h.Update(1)
	if h.Data.ChunkSize != 2 {
		t.Fatalf("wrong data chunk size: got %d, want 2", h.Data.ChunkSize)
	}
	h.Update(2)
	if h.Data.ChunkSize != 4 {
		t.Fatalf("wrong data chunk size: got %d, want 4", h.Data.ChunkSize)
	}

	hb, err := NewHeader(20000, 1, 2, LPCM, binary.BigEndian, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	w.Reset()
	if err := binary.Write(w, binary.BigEndian, hb); err != nil {
		t.Fatalf("failed to write header: %v", err)
	}
	bb := w.Bytes()
	if len(bb) != sizeOfHeader {
		t.Fatalf("wrong number of header bytes written: got %d, want %d", len(bb), sizeOfHeader)
	}
	magicb := string(bb[:4])
	if magicb != "RIFX" {
		t.Fatalf("wrong magic number in header bytes: got %s, want RIFX", magicb)
	}

	hf, err := NewHeader(20000, 1, 4, IEEEFloatingPoint, binary.BigEndian, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	w.Reset()
	if err := binary.Write(w, binary.LittleEndian, hf); err != nil {
		t.Fatalf("failed to write header: %v", err)
	}
	if w.Len() != sizeOfHeader {
		t.Fatalf("wrong number of header bytes written: got %d, want %d", w.Len(), sizeOfHeader)
	}
	hf.Update(1)
	if hf.Data.ChunkSize != 4 {
		t.Fatalf("wrong data chunk size: got %d, want 4", hf.Data.ChunkSize)
	}
	hf.Update(2)
	if hf.Data.ChunkSize != 8 {
		t.Fatalf("wrong data chunk size: got %d, want 8", hf.Data.ChunkSize)
	}

	_, err = NewHeader(20000, 1, 5, LPCM, binary.BigEndian, 0)
	if err == nil {
		t.Fatal("unexpected success on invalid bytes per sample")
	}
	if !strings.Contains(err.Error(), "bytes per sample") {
		t.Errorf("wrong error message: got '%s', want 'bytes per sample'", err.Error())
	}

	_, err = NewHeader(20000, 1, 2, IEEEFloatingPoint, binary.BigEndian, 0)
	if err == nil {
		t.Fatal("unexpected success on invalid bytes per sample")
	}
	if !strings.Contains(err.Error(), "bytes per sample") {
		t.Errorf("wrong error message: got '%s', want 'bytes per sample'", err.Error())
	}

	_, err = NewHeader(20000, 1, 2, 0, binary.BigEndian, 0)
	if err == nil {
		t.Fatal("unexpected success on invalid sample format")
	}
	if !strings.Contains(err.Error(), "sample format") {
		t.Errorf("wrong error message: got '%s', want 'sample format'", err.Error())
	}
}
