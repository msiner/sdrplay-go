// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"math"
	"sort"
)

type IntStats struct {
	Count  int
	Mean   float64
	StdDev float64
	Median int
	Min    int
	Max    int
}

type IntAccumulator struct {
	vals  []int
	sum   float64
	stats IntStats
}

func (a *IntAccumulator) Add(x int) {
	switch a.stats.Count {
	case 0:
		a.stats.Min = x
		a.stats.Max = x
	default:
		if x > a.stats.Max {
			a.stats.Max = x
		}
		if x < a.stats.Min {
			a.stats.Min = x
		}
	}
	a.stats.Count++
	a.sum += float64(x)
	a.vals = append(a.vals, x)
}

func (a *IntAccumulator) Analyze() IntStats {
	sort.Ints(a.vals)
	a.stats.Median = a.vals[len(a.vals)/2]
	a.stats.Mean = a.sum / float64(a.stats.Count)
	var (
		tmp      float64
		variance float64
	)
	for _, x := range a.vals {
		tmp = float64(x) - a.stats.Mean
		variance += tmp * tmp
	}
	a.stats.StdDev = math.Sqrt(variance / float64(a.stats.Count))
	return a.stats
}

func (a *IntAccumulator) Reset() {
	a.stats.Count = 0
	a.sum = 0
	a.vals = a.vals[:0]
}

type FloatStats struct {
	Count  int
	Mean   float64
	StdDev float64
	Median float64
	Min    float64
	Max    float64
}

type FloatAccumulator struct {
	vals  []float64
	sum   float64
	stats FloatStats
}

func (a *FloatAccumulator) Add(x float64) {
	switch a.stats.Count {
	case 0:
		a.stats.Min = x
		a.stats.Max = x
	default:
		if x > a.stats.Max {
			a.stats.Max = x
		}
		if x < a.stats.Min {
			a.stats.Min = x
		}
	}
	a.stats.Count++
	a.sum += float64(x)
	a.vals = append(a.vals, x)
}

func (a *FloatAccumulator) Analyze() FloatStats {
	sort.Float64s(a.vals)
	a.stats.Median = a.vals[len(a.vals)/2]
	a.stats.Mean = a.sum / float64(a.stats.Count)
	var (
		tmp      float64
		variance float64
	)
	for _, x := range a.vals {
		tmp = float64(x) - a.stats.Mean
		variance += tmp * tmp
	}
	a.stats.StdDev = math.Sqrt(variance / float64(a.stats.Count))
	return a.stats
}

func (a *FloatAccumulator) Reset() {
	a.stats.Count = 0
	a.sum = 0
	a.vals = a.vals[:0]
}
