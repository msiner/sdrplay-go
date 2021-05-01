// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package api

//go:generate go run golang.org/x/tools/cmd/stringer -type Rsp2_AntennaSelectT,Rsp2_AmPortSelectT -output rsp2_string.go

const (
	RSPII_NUM_LNA_STATES        = 9
	RSPII_NUM_LNA_STATES_AMPORT = 5
	RSPII_NUM_LNA_STATES_420MHZ = 6
)

// RSP2 parameter enums
type Rsp2_AntennaSelectT int32

const (
	Rsp2_ANTENNA_A Rsp2_AntennaSelectT = 5
	Rsp2_ANTENNA_B Rsp2_AntennaSelectT = 6
)

type Rsp2_AmPortSelectT int32

const (
	Rsp2_AMPORT_1 Rsp2_AmPortSelectT = 1
	Rsp2_AMPORT_2 Rsp2_AmPortSelectT = 0
)

// RSP2 parameter structs
type Rsp2ParamsT struct {
	ExtRefOutputEn uint8 // default: 0
}

type Rsp2TunerParamsT struct {
	BiasTEnable   uint8               // default: 0
	AmPortSel     Rsp2_AmPortSelectT  // default: sdrplay_api_Rsp2_AMPORT_2
	AntennaSel    Rsp2_AntennaSelectT // default: sdrplay_api_Rsp2_ANTENNA_A
	RfNotchEnable uint8               // default: 0
}
