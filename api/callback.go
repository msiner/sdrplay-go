// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package api

//go:generate go run golang.org/x/tools/cmd/stringer -type PowerOverloadCbEventIdT,RspDuoModeCbEventIdT,EventT -output callback_string.go

// Event callback enums
type PowerOverloadCbEventIdT int32

const (
	Overload_Detected  PowerOverloadCbEventIdT = 0
	Overload_Corrected PowerOverloadCbEventIdT = 1
)

type RspDuoModeCbEventIdT int32

const (
	PrimaryInitialised      RspDuoModeCbEventIdT = 0
	SecondaryAttached       RspDuoModeCbEventIdT = 1
	SecondaryDetached       RspDuoModeCbEventIdT = 2
	SecondaryInitialised    RspDuoModeCbEventIdT = 3
	SecondaryUninitialised  RspDuoModeCbEventIdT = 4
	PrimaryDllDisappeared   RspDuoModeCbEventIdT = 5
	SecondaryDllDisappeared RspDuoModeCbEventIdT = 6
)

type EventT int32

const (
	GainChange          EventT = 0
	PowerOverloadChange EventT = 1
	DeviceRemoved       EventT = 2
	RspDuoModeChange    EventT = 3
)

// Event callback parameter structs
type GainCbParamT struct {
	GRdB     uint32
	LnaGRdB  uint32
	CurrGain float64
}

type PowerOverloadCbParamT struct {
	PowerOverloadChangeType PowerOverloadCbEventIdT
}

type RspDuoModeCbParamT struct {
	ModeChangeType RspDuoModeCbEventIdT
}

// EventParamsT represents a union type of the same name in the C API.
// Recreating an actual union in Go tends to be ugly and unsafe.
// Instead, the Go version of EventParamsT holds a full copy of each
// of the union members.
type EventParamsT struct {
	GainParams          GainCbParamT
	PowerOverloadParams PowerOverloadCbParamT
	RspDuoModeParams    RspDuoModeCbParamT
}

// Stream callback parameter structs
type StreamCbParamsT struct {
	FirstSampleNum uint32
	GrChanged      int32
	RfChanged      int32
	FsChanged      int32
	NumSamples     uint32
}

// Callback function prototypes. Unlike the C versions,
// there is no cbContext argument. In Go, this is unneccessary because
// callbacks can be closures or functions bound to types with state.
// Regardless of the actual implementation, callback implementations
// should assume any slice and pointer types in the callback arguments
// are owned by the C API. Data should be explicitly copied if it
// needs to be stored or otherwise escape the scope of the callback
// function.
type (
	StreamCallbackT func(xi, xq []int16, params *StreamCbParamsT, reset bool)
	EventCallbackT  func(eventId EventT, tuner TunerSelectT, params *EventParamsT)
)

// Callback channels struct
type CallbackFnsT struct {
	StreamACbFn StreamCallbackT
	StreamBCbFn StreamCallbackT
	EventCbFn   EventCallbackT
}
