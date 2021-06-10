// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package callback_test

import (
	"fmt"

	"github.com/msiner/sdrplay-go/api"
	"github.com/msiner/sdrplay-go/helpers/callback"
)

func ExampleDropDetectFn() {
	// Simple mock callback parameters
	const numSamples = 1000
	params := api.StreamCbParamsT{
		NumSamples:     numSamples,
		FirstSampleNum: 0,
	}

	detect := callback.NewDropDetectFn()

	// On the first call or after a reset is signaled, the detector
	// does not have any prior knowledge to use. So it will report
	// zero drops.
	drops := detect(&params, false)
	fmt.Println(drops)

	// If FirstSampleNum increases by the number of samples in the last
	// update, the detector reports zero drops.
	params.FirstSampleNum += params.NumSamples
	drops = detect(&params, false)
	fmt.Println(drops)

	// If FirstSampleNum increases by more than the number of samples
	// in the last update, then the detector assumes the difference
	// in the expected FirstSampleNum and the actual FirstSampleNum
	// to be the number of dropped samples.
	params.FirstSampleNum += params.NumSamples * 2
	drops = detect(&params, false)
	fmt.Println(drops)

	// Output:
	// 0
	// 0
	// 1000
}
