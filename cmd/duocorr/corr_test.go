// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"math"
	"math/rand"
	"testing"

	"github.com/msiner/sdrplay-go/helpers/duo"
)

func TestNewCorrelationFn(t *testing.T) {
	const (
		numSamples = 4096
		numIters   = 100
	)

	corr := NewCorrelationFn(numSamples)
	msg := &duo.SynchroMsg{
		Xia:   make([]int16, numSamples),
		Xqa:   make([]int16, numSamples),
		Xib:   make([]int16, numSamples),
		Xqb:   make([]int16, numSamples),
		Reset: true,
	}

	fillRand := func(dst []int16) {
		for i := range dst {
			dst[i] = int16(rand.Intn(2*math.MaxInt16) - math.MaxInt16)
		}
	}

	shiftLeft := func(slc []int16) {
		last := slc[0]
		for i := len(slc) - 1; i >= 0; i-- {
			x := slc[i]
			slc[i] = last
			last = x
		}
	}

	shiftRight := func(slc []int16) {
		last := slc[len(slc)-1]
		for i, x := range slc {
			slc[i] = last
			last = x
		}
	}

	for i := 0; i < numIters; i++ {
		fillRand(msg.Xia)
		fillRand(msg.Xqa)
		copy(msg.Xib, msg.Xia)
		copy(msg.Xqb, msg.Xqa)
		resA := corr(msg)
		if resA.ZScore < 4 {
			t.Errorf("z-score is too small for identical random data: got %v, want > 4", resA.ZScore)
		}
		fillRand(msg.Xib)
		fillRand(msg.Xqb)
		resB := corr(msg)
		if resB.ZScore > resB.ZScore {
			t.Errorf("z-score is too large for different random data: got %v, want < %v", resB.ZScore, resA.ZScore)
		}
	}

	fillRand(msg.Xia)
	fillRand(msg.Xqa)

	copy(msg.Xib, msg.Xia)
	copy(msg.Xqb, msg.Xqa)
	for i := 0; i < corrMaxOffset; i++ {
		res := corr(msg)
		if res.Peak != -i {
			t.Errorf("wrong peak offset after shift left %d: got %d, want %d", i, res.Peak, -i)
		}
		shiftLeft(msg.Xib)
		shiftLeft(msg.Xqb)
	}
	res := corr(msg)
	if res.Peak != -corrMaxOffset {
		t.Errorf("wrong peak offset after shift left %d: got %d, want %d", -corrMaxOffset, res.Peak, -corrMaxOffset)
	}

	copy(msg.Xib, msg.Xia)
	copy(msg.Xqb, msg.Xqa)
	for i := 0; i < corrMaxOffset; i++ {
		res := corr(msg)
		if res.Peak != i {
			t.Errorf("wrong peak offset after shift right %d: got %d, want %d", i, res.Peak, i)
		}
		shiftRight(msg.Xib)
		shiftRight(msg.Xqb)
	}
	res = corr(msg)
	if res.Peak != corrMaxOffset {
		t.Errorf("wrong peak offset after shift left %d: got %d, want %d", corrMaxOffset, res.Peak, corrMaxOffset)
	}
}

func BenchmarkNewCorrelationFn(b *testing.B) {
	const numSamples = 4096

	corr := NewCorrelationFn(numSamples)
	msg := &duo.SynchroMsg{
		Xia:   make([]int16, numSamples),
		Xqa:   make([]int16, numSamples),
		Xib:   make([]int16, numSamples),
		Xqb:   make([]int16, numSamples),
		Reset: true,
	}

	fillRand := func(dst []int16) {
		for i := range dst {
			dst[i] = int16(rand.Intn(2*math.MaxInt16) - math.MaxInt16)
		}
	}

	fillRand(msg.Xia)
	fillRand(msg.Xqa)

	copy(msg.Xib, msg.Xia)
	copy(msg.Xqb, msg.Xqa)

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		res := corr(msg)
		if res.ZScore < 4 {
			b.Fatalf("z-score is too small for identical random data: got %v, want > 4", res.ZScore)
		}
	}
}
