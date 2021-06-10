// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package callback_test

import (
	"fmt"

	"github.com/msiner/sdrplay-go/api"
	"github.com/msiner/sdrplay-go/helpers/callback"
)

func ExampleStreamChan() {
	// Simple mock samples and callback parameters
	xi := []int16{1, 3, 5, 7}
	xq := []int16{2, 4, 6, 8}
	params := api.StreamCbParamsT{
		NumSamples: 4,
	}

	// Depth of 1 to guarantee our test callback won't drop.
	sc := callback.NewStreamChan(1)

	// Callback would normally happen in C API processing thread.
	// Just use a goroutine here for demonstration.
	go func() {
		sc.Callback(xi, xq, &params, false)
	}()

	// Wait for the message generated by the StreamChan callback.
	msg := <-sc.C

	fmt.Println(msg.Params.NumSamples)
	fmt.Println(msg.Xi)
	fmt.Println(msg.Xq)
	// Output:
	// 4
	// [1 3 5 7]
	// [2 4 6 8]
}
