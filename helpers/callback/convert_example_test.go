// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package callback_test

import (
	"fmt"
	"math"

	"github.com/msiner/sdrplay-go/helpers/callback"
)

func ExampleConvertToFloat32Fn() {
	const (
		max   = math.MaxInt16
		phalf = math.MaxInt16 / 2
		nhalf = math.MinInt16 / 2
		min   = math.MinInt16
	)
	convert := callback.NewConvertToFloat32Fn(16)

	ints := []int16{0, phalf, max, phalf, 0, nhalf, min, nhalf}
	floats := convert(ints)
	fmt.Println(floats)
	// Output:
	// [0 0.49996948 0.9999695 0.49996948 0 -0.5 -1 -0.5]
}

func ExampleConvertToComplex64Fn() {
	const (
		max   = math.MaxInt16
		phalf = math.MaxInt16 / 2
		nhalf = math.MinInt16 / 2
		min   = math.MinInt16
	)
	convert := callback.NewConvertToComplex64Fn(16)

	xi := []int16{0, phalf, max, phalf}
	xq := []int16{nhalf, min, nhalf, 0}
	cx := convert(xi, xq)
	fmt.Println(cx)
	// Output:
	// [(0-0.5i) (0.49996948-1i) (0.9999695-0.5i) (0.49996948+0i)]
}
