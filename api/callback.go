// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package api

//go:generate go run golang.org/x/tools/cmd/stringer -type PowerOverloadCbEventIdT,RspDuoModeCbEventIdT,EventT -output callback_string.go

import (
	"unsafe"
)

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

// Event parameters overlay. It is hard to correctly represent a
// C union in Go. Of the 3 union members, the largest size is GainParamsT
// at 16 bytes. The strictest alignment is 8 bytes due to the 64-bit
// CurrGain parameters in GainCbParamT. Therefore, using [2]uint64 as
// the base type to represent the data gives us the correct size and
// alignment.
type EventParamsT [2]uint64

func (p *EventParamsT) GainParams() *GainCbParamT {
	return (*GainCbParamT)(unsafe.Pointer(p))
}

func (p *EventParamsT) PowerOverloadParams() *PowerOverloadCbParamT {
	return (*PowerOverloadCbParamT)(unsafe.Pointer(p))
}

func (p *EventParamsT) RspDuoModeParams() *RspDuoModeCbParamT {
	return (*RspDuoModeCbParamT)(unsafe.Pointer(p))
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
// callbacks can be closures or methods that have bound state.
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
