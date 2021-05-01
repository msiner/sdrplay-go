// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// +build cgo,!dll

#include "_cgo_export.h"

#include <stdint.h>
#include <sdrplay_api.h>

void cgoStreamACallback(
	short *xi, short *xq, sdrplay_api_StreamCbParamsT *params,
	unsigned int numSamples, unsigned int reset, void *cbContext
) {
	streamACallback((GoUintptr)xi, (GoUintptr)xq, (GoUintptr)params, (GoUint32)numSamples, (GoUint32)reset, (GoUintptr)cbContext);
}

void cgoStreamBCallback(
	short *xi, short *xq, sdrplay_api_StreamCbParamsT *params,
	unsigned int numSamples, unsigned int reset, void *cbContext
) {
	streamBCallback((GoUintptr)xi, (GoUintptr)xq, (GoUintptr)params, (GoUint32)numSamples, (GoUint32)reset, (GoUintptr)cbContext);
}

void cgoEventCallback(
	sdrplay_api_EventT eventId, sdrplay_api_TunerSelectT tuner,
	sdrplay_api_EventParamsT *params, void *cbContext
) {
	eventCallback((GoInt32)eventId, (GoInt32)tuner, (GoUintptr)params, (GoUintptr)cbContext);
}

sdrplay_api_ErrT wrapper_api_Init(HANDLE dev) {
    sdrplay_api_CallbackFnsT cbFuncs;
	cbFuncs.StreamACbFn = cgoStreamACallback;
	cbFuncs.StreamBCbFn = cgoStreamBCallback;
	cbFuncs.EventCbFn = cgoEventCallback;
    return sdrplay_api_Init(dev, &cbFuncs, (void*)dev);
}