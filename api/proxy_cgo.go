// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// +build cgo,!dll

package api

import "C"

import (
	"reflect"
	"unsafe"
)

// streamACallback is the global callback handler/proxy that routes
// stream A callbacks from a cgo global callback to a user-defined Go
// StreamCallbackT function.
//
//export streamACallback
func streamACallback(xi, xq, params uintptr, numSamples, reset uint32, cbContext uintptr) {
	cbMutex.Lock()
	cb := streamACbFn
	cbMutex.Unlock()
	if cb == nil {
		return
	}

	// The following code is duplicated from handleCallback(), DO NOT EDIT THIS COPY
	var (
		sxi      []int16
		sxq      []int16
		cbParams *StreamCbParamsT
	)

	if xi != 0 && xq != 0 && numSamples > 0 {
		hxi := (*reflect.SliceHeader)(unsafe.Pointer(&sxi))
		hxi.Cap = int(numSamples)
		hxi.Len = int(numSamples)
		hxi.Data = xi

		hxq := (*reflect.SliceHeader)(unsafe.Pointer(&sxq))
		hxq.Cap = int(numSamples)
		hxq.Len = int(numSamples)
		hxq.Data = xq
	}

	if params != 0 {
		cbParams = (*StreamCbParamsT)(unsafe.Pointer(params))
	}

	cb(sxi, sxq, cbParams, reset != 0)
}

// streamBCallback is the global callback handler/proxy that routes
// stream B callbacks from a cgo global callback to a user-defined Go
// StreamCallbackT function.
//
//export streamBCallback
func streamBCallback(xi, xq, params uintptr, numSamples, reset uint32, cbContext uintptr) {
	cbMutex.Lock()
	cb := streamBCbFn
	cbMutex.Unlock()
	if cb == nil {
		return
	}

	// The following code is duplicated from handleCallback(), DO NOT EDIT THIS COPY
	var (
		sxi      []int16
		sxq      []int16
		cbParams *StreamCbParamsT
	)

	if xi != 0 && xq != 0 && numSamples > 0 {
		hxi := (*reflect.SliceHeader)(unsafe.Pointer(&sxi))
		hxi.Cap = int(numSamples)
		hxi.Len = int(numSamples)
		hxi.Data = xi

		hxq := (*reflect.SliceHeader)(unsafe.Pointer(&sxq))
		hxq.Cap = int(numSamples)
		hxq.Len = int(numSamples)
		hxq.Data = xq
	}

	if params != 0 {
		cbParams = (*StreamCbParamsT)(unsafe.Pointer(params))
	}

	cb(sxi, sxq, cbParams, reset != 0)
}

// eventCallback is the global callback handler/proxy that routes
// event callbacks from a cgo global callback to a user-defined Go
// EventCallbackT function.
//
//export eventCallback
func eventCallback(eventId, tuner int32, params, cbContext uintptr) {
	cbMutex.Lock()
	cb := eventCbFn
	cbMutex.Unlock()
	if cb == nil {
		return
	}

	var evParams *EventParamsT
	if params != 0 {
		evParams = (*EventParamsT)(unsafe.Pointer(params))
	}

	cb(EventT(eventId), TunerSelectT(tuner), evParams)
}
