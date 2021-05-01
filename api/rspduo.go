// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package api

import (
	"fmt"
	"strings"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type RspDuo_AmPortSelectT -output rspduo_string.go

const (
	RSPDUO_NUM_LNA_STATES        = 10
	RSPDUO_NUM_LNA_STATES_AMPORT = 5
	RSPDUO_NUM_LNA_STATES_AM     = 7
	RSPDUO_NUM_LNA_STATES_LBAND  = 9
)

type RspDuoModeT int32

const (
	RspDuoMode_Unknown      RspDuoModeT = 0
	RspDuoMode_Single_Tuner RspDuoModeT = 1
	RspDuoMode_Dual_Tuner   RspDuoModeT = 2
	RspDuoMode_Primary      RspDuoModeT = 4
	RspDuoMode_Secondary    RspDuoModeT = 8
)

func (m RspDuoModeT) String() string {
	if m == RspDuoMode_Unknown {
		return "Unknown"
	}

	var words []string
	if m&RspDuoMode_Single_Tuner != 0 {
		words = append(words, "Single")
	}
	if m&RspDuoMode_Dual_Tuner != 0 {
		words = append(words, "Dual")
	}
	if m&RspDuoMode_Primary != 0 {
		words = append(words, "Primary")
	}
	if m&RspDuoMode_Secondary != 0 {
		words = append(words, "Secondary")
	}
	if len(words) == 0 {
		return fmt.Sprintf("InvalidMode(%d)", m)
	}
	return strings.Join(words, "|")
}

type RspDuo_AmPortSelectT int32

const (
	RspDuo_AMPORT_1 RspDuo_AmPortSelectT = 1
	RspDuo_AMPORT_2 RspDuo_AmPortSelectT = 0
)

// RSPduo parameter structs
type RspDuoParamsT struct {
	ExtRefOutputEn int32 // default: 0
}

type RspDuoTunerParamsT struct {
	BiasTEnable         uint8                // default: 0
	Tuner1AmPortSel     RspDuo_AmPortSelectT // default: sdrplay_api_RspDuo_AMPORT_2
	Tuner1AmNotchEnable uint8                // default: 0
	RfNotchEnable       uint8                // default: 0
	RfDabNotchEnable    uint8                // default: 0
}
