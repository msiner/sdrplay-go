// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"math"

	"github.com/msiner/sdrplay-go/helpers/callback"
	"github.com/msiner/sdrplay-go/helpers/duo"
)

const (
	corrMaxOffset = 20
	corrWidth     = corrMaxOffset*2 + 1
)

// Correlation holds the results of a single correlation pass.
type Correlation struct {
	Peak    int
	PeakRaw float64
	ZScore  float64
	Stats   FloatStats
	Raw     []float64
}

// CorrelationFn is a function type that processes a SynchroMsg
// by performing a correlation on the magnitude of the samples.
type CorrelationFn func(msg *duo.SynchroMsg) Correlation

// NewCorrelationFn creates a new CorrelationFn that maintains
// its state in internal captured buffers. The correlation code
// is hand-written in order to avoid a module dependency on a
// third-party DSP library. This is a naive brute-force
// cross-correlation implementation.
func NewCorrelationFn(cbSamples int) CorrelationFn {
	amag := make([]float64, cbSamples)
	bmag := make([]float64, cbSamples)
	corr := make([]float64, corrWidth)
	aToCx := callback.NewConvertToComplex64Fn(16)
	bToCx := callback.NewConvertToComplex64Fn(16)
	acc := FloatAccumulator{}
	return func(msg *duo.SynchroMsg) Correlation {
		acc.Reset()

		cxa := aToCx(msg.Xia, msg.Xqa)
		cxb := bToCx(msg.Xib, msg.Xqb)

		if cap(amag) < len(cxa) {
			amag = make([]float64, len(cxa))
			bmag = make([]float64, len(cxa))
		}
		amag = amag[:len(cxa)]
		bmag = bmag[:len(cxa)]

		for i, a := range cxa {
			b := cxb[i]
			amag[i] = math.Hypot(float64(real(a)), float64(imag(a)))
			bmag[i] = math.Hypot(float64(real(b)), float64(imag(b)))
		}

		currOffset := -corrMaxOffset
		awin := amag[corrMaxOffset : len(amag)-corrMaxOffset]
		for i := range corr {
			corr[i] = 0
			bwin := bmag[corrMaxOffset+currOffset : corrMaxOffset+currOffset+len(awin)]
			for j, a := range awin {
				corr[i] += a * bwin[j]
			}
			currOffset++
		}

		var (
			max  float64
			maxi int
			sum  float64
		)
		for i, x := range corr {
			acc.Add(x)
			sum += x
			if x > max {
				max = x
				maxi = i
			}
		}

		mean := sum / float64(len(corr))
		stats := acc.Analyze()

		pos := maxi - (corrWidth / 2)

		res := Correlation{
			Peak:    pos,
			PeakRaw: max,
			ZScore:  (max - mean) / stats.StdDev,
			Stats:   stats,
			Raw:     make([]float64, corrWidth),
		}
		copy(res.Raw, corr)
		return res
	}
}
