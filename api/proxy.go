// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package api

import (
	"reflect"
	"sync"
	"unsafe"
)

var (
	cbMutex     sync.Mutex
	streamACbFn StreamCallbackT
	streamBCbFn StreamCallbackT
	eventCbFn   EventCallbackT
)

// registerCallbacks swaps in the provided callback functions in
// a thread-safe way.
func registerCallbacks(dev Handle, callbacks CallbackFnsT) {
	cbMutex.Lock()
	streamACbFn = callbacks.StreamACbFn
	streamBCbFn = callbacks.StreamBCbFn
	eventCbFn = callbacks.EventCbFn
	cbMutex.Unlock()
}

// unregisterCallbacks sets all callbacks to nil in a thread-safe way.
func unregisterCallbacks(dev Handle) {
	cbMutex.Lock()
	streamACbFn = nil
	streamBCbFn = nil
	eventCbFn = nil
	cbMutex.Unlock()
}

// handleCallback contains the code to translate the sdrplay_api callback
// arguments containing pointers, to a call to a Go StreamCallbackT
// function. Currently, this function is not actually used by the stream
// handlers because it is too complex for the Go compiler to inline.
// So the code is duplicated in each handler. This function remains as
// the standard to compare to and an easy way to benchmark and test
// modifications in a generic way.
func handleCallback(cb StreamCallbackT, xi, xq, params unsafe.Pointer, numSamples, reset uintptr) {
	var (
		sxi      []int16
		sxq      []int16
		cbParams *StreamCbParamsT
	)

	if xi != nil && xq != nil && numSamples > 0 {
		hxi := (*reflect.SliceHeader)(unsafe.Pointer(&sxi))
		hxi.Cap = int(numSamples)
		hxi.Len = int(numSamples)
		hxi.Data = (uintptr)(xi)

		hxq := (*reflect.SliceHeader)(unsafe.Pointer(&sxq))
		hxq.Cap = int(numSamples)
		hxq.Len = int(numSamples)
		hxq.Data = (uintptr)(xq)
	}

	if params != nil {
		cbParams = (*StreamCbParamsT)(params)
	}

	cb(sxi, sxq, cbParams, reset != 0)
}
