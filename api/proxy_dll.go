// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
// +build windows,!cgo windows,dll

package api

import (
	"reflect"
	"unsafe"
)

// streamACallback is the global callback handler/proxy that routes
// stream A callbacks from sdrplay_api to a user-defined Go
// StreamCallbackT function.
func streamACallback(xi, xq, params unsafe.Pointer, numSamples, reset uintptr, cbContext unsafe.Pointer) uintptr {
	cbMutex.Lock()
	cb := streamACbFn
	cbMutex.Unlock()
	if cb == nil {
		return 0
	}

	// The following code is duplicated from handleCallback(), DO NOT EDIT THIS COPY
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

	return 0
}

// streamBCallback is the global callback handler/proxy that routes
// stream B callbacks from sdrplay_api to a user-defined Go
// StreamCallbackT function.
func streamBCallback(xi, xq, params unsafe.Pointer, numSamples, reset uintptr, cbContext unsafe.Pointer) uintptr {
	cbMutex.Lock()
	cb := streamBCbFn
	cbMutex.Unlock()
	if cb == nil {
		return 0
	}

	// The following code is duplicated from handleCallback(), DO NOT EDIT THIS COPY
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

	return 0
}

// eventCallback is the global callback handler/proxy that routes
// event callbacks from sdrplay_api to a user-defined Go
// EventCallbackT function.
func eventCallback(eventId, tuner uintptr, params, cbContext unsafe.Pointer) uintptr {
	cbMutex.Lock()
	cb := eventCbFn
	cbMutex.Unlock()
	if cb == nil {
		return 0
	}

	var evParams *EventParamsT
	if params != nil {
		evParams = (*EventParamsT)(unsafe.Pointer(params))
	}

	cb(EventT(eventId), TunerSelectT(tuner), evParams)

	return 0
}
