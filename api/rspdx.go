// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package api

//go:generate go run golang.org/x/tools/cmd/stringer -type RspDx_AntennaSelectT,RspDx_HdrModeBwT -output rspdx_string.go

const (
	RSPDX_NUM_LNA_STATES               = 28 // Number of LNA states in all bands (except where defined differently below)
	RSPDX_NUM_LNA_STATES_AMPORT2_0_12  = 19 // Number of LNA states when using AM Port 2 between 0 and 12MHz
	RSPDX_NUM_LNA_STATES_AMPORT2_12_60 = 20 // Number of LNA states when using AM Port 2 between 12 and 60MHz
	RSPDX_NUM_LNA_STATES_VHF_BAND3     = 27 // Number of LNA states in VHF and Band3
	RSPDX_NUM_LNA_STATES_420MHZ        = 21 // Number of LNA states in 420MHz band
	RSPDX_NUM_LNA_STATES_LBAND         = 19 // Number of LNA states in L-band
	RSPDX_NUM_LNA_STATES_DX            = 26 // Number of LNA states in DX path
)

// RSPdx parameter enums
type RspDx_AntennaSelectT int32

const (
	RspDx_ANTENNA_A RspDx_AntennaSelectT = 0
	RspDx_ANTENNA_B RspDx_AntennaSelectT = 1
	RspDx_ANTENNA_C RspDx_AntennaSelectT = 2
)

type RspDx_HdrModeBwT int32

const (
	RspDx_HDRMODE_BW_0_200 RspDx_HdrModeBwT = 0
	RspDx_HDRMODE_BW_0_500 RspDx_HdrModeBwT = 1
	RspDx_HDRMODE_BW_1_200 RspDx_HdrModeBwT = 2
	RspDx_HDRMODE_BW_1_700 RspDx_HdrModeBwT = 3
)

// RSPdx parameter structs
type RspDxParamsT struct {
	HdrEnable        uint8                // default: 0
	BiasTEnable      uint8                // default: 0
	AntennaSel       RspDx_AntennaSelectT // default: RspDx_ANTENNA_A
	RfNotchEnable    uint8                // default: 0
	RfDabNotchEnable uint8                // default: 0
}

type RspDxTunerParamsT struct {
	HdrBw RspDx_HdrModeBwT // default: RspDx_HDRMODE_BW_1_700
}
