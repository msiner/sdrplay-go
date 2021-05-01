// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package api

const (
	RSPIA_NUM_LNA_STATES       = 10
	RSPIA_NUM_LNA_STATES_AM    = 7
	RSPIA_NUM_LNA_STATES_LBAND = 9
)

// RSP1A parameter enums

// RSP1A parameter structs
type Rsp1aParamsT struct {
	RfNotchEnable    uint8 // default: 0
	RfDabNotchEnable uint8 // default: 0
}

type Rsp1aTunerParamsT struct {
	BiasTEnable uint8 // default: 0
}
