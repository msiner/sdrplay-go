// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package api

//go:generate go run golang.org/x/tools/cmd/stringer -type Bw_MHzT,If_kHzT,LoModeT,MinGainReductionT,TunerSelectT -output tuner_string.go

const MAX_BB_GR = 59

// Tuner parameter enums
type Bw_MHzT int32

const (
	BW_Undefined Bw_MHzT = 0
	BW_0_200     Bw_MHzT = 200
	BW_0_300     Bw_MHzT = 300
	BW_0_600     Bw_MHzT = 600
	BW_1_536     Bw_MHzT = 1536
	BW_5_000     Bw_MHzT = 5000
	BW_6_000     Bw_MHzT = 6000
	BW_7_000     Bw_MHzT = 7000
	BW_8_000     Bw_MHzT = 8000
)

func (v Bw_MHzT) Hz() float64 {
	if v < 0 {
		return 0
	}
	return float64(v) * 1000
}

type If_kHzT int32

const (
	IF_Undefined If_kHzT = -1
	IF_Zero      If_kHzT = 0
	IF_0_450     If_kHzT = 450
	IF_1_620     If_kHzT = 1620
	IF_2_048     If_kHzT = 2048
)

type LoModeT int32

const (
	LO_Undefined LoModeT = 0
	LO_Auto      LoModeT = 1
	LO_120MHz    LoModeT = 2
	LO_144MHz    LoModeT = 3
	LO_168MHz    LoModeT = 4
)

type MinGainReductionT int32

const (
	EXTENDED_MIN_GR MinGainReductionT = 0
	NORMAL_MIN_GR   MinGainReductionT = 20
)

type TunerSelectT int32

const (
	Tuner_Neither TunerSelectT = 0
	Tuner_A       TunerSelectT = 1
	Tuner_B       TunerSelectT = 2
	Tuner_Both    TunerSelectT = 3
)

type GainValuesT struct {
	Curr float32
	Max  float32
	Min  float32
}

type GainT struct {
	GRdB       int32             // default: 50
	LNAstate   uint8             // default: 0
	SyncUpdate uint8             // default: 0
	MinGr      MinGainReductionT // default: NORMAL_MIN_GR
	GainVals   GainValuesT       // output parameter
}

type RfFreqT struct {
	RfHz       float64 // default: 200000000.0
	SyncUpdate uint8   // default: 0
}

type DcOffsetTunerT struct {
	DcCal           uint8 // default: 3 (Periodic mode)
	SpeedUp         uint8 // default: 0 (No speedup)
	TrackTime       int32 // default: 1    (=> time in uSec = (72 * 3 * trackTime) / 24e6       = 9uSec)
	RefreshRateTime int32 // default: 2048 (=> time in uSec = (72 * 3 * refreshRateTime) / 24e6 = 18432uSec)
}

type TunerParamsT struct {
	BwType        Bw_MHzT // default: BW_0_200
	IfType        If_kHzT // default: IF_Zero
	LoMode        LoModeT // default: LO_Auto
	Gain          GainT
	RfFreq        RfFreqT
	DcOffsetTuner DcOffsetTunerT
}
