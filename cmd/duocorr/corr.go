package main

import (
	"math"

	"github.com/msiner/sdrplay-go/helpers/callback"
	"github.com/msiner/sdrplay-go/helpers/duo"
)

// Correlation holds the results of a single correlation pass.
type Correlation struct {
	Peak    int
	PeakRaw float64
	ZScore  float64
	Raw     []float64
}

// CorrelationFn is a function type that processes a SynchroMsg
// by performing a correlation on the magnitude of the samples.
type CorrelationFn func(msg *duo.SynchroMsg) Correlation

// NewCorrelationFn creates a new CorrelationFn that maintains
// its state in internal captured buffers.
func NewCorrelationFn(cbSamples int) CorrelationFn {
	const offset = 20
	const width = offset*2 + 1
	amag := make([]float64, cbSamples)
	bmag := make([]float64, cbSamples)
	corr := make([]float64, width)
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

		currOffset := -offset
		awin := amag[offset : len(amag)-offset]
		for i := range corr {
			corr[i] = 0
			bwin := bmag[offset+currOffset : offset+currOffset+len(awin)]
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

		pos := maxi - (width / 2)

		res := Correlation{
			Peak:    pos,
			PeakRaw: max,
			ZScore:  (max - mean) / stats.StdDev,
			Raw:     make([]float64, width),
		}
		copy(res.Raw, corr)
		return res
	}
}
