// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package api

//go:generate go run golang.org/x/tools/cmd/stringer -type TransferModeT -output dev_string.go

// Dev parameter enums
type TransferModeT int32

const (
	ISOCH TransferModeT = 0
	BULK  TransferModeT = 1
)

// Dev parameter structs
type FsFreqT struct {
	FsHz       float64 // default: 2000000.0
	SyncUpdate uint8   // default: 0
	ReCal      uint8   // default: 0
}

type SyncUpdateT struct {
	SampleNum uint32 // default: 0
	Period    uint32 // default: 0
}

type ResetFlagsT struct {
	ResetGainUpdate uint8 // default: 0
	ResetRfUpdate   uint8 // default: 0
	ResetFsUpdate   uint8 // default: 0
}

type DevParamsT struct {
	Ppm           float64 // default: 0.0
	FsFreq        FsFreqT
	SyncUpdate    SyncUpdateT
	ResetFlags    ResetFlagsT
	Mode          TransferModeT // default: ISOCH
	SamplesPerPkt uint32        // default: 0 (output param)
	Rsp1aParams   Rsp1aParamsT
	Rsp2Params    Rsp2ParamsT
	RspDuoParams  RspDuoParamsT
	RspDxParams   RspDxParamsT
}
