// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package api

//go:generate go run golang.org/x/tools/cmd/stringer -type AgcControlT,AdsbModeT -output control_string.go

// Control parameter enums
type AgcControlT int32

const (
	AGC_DISABLE AgcControlT = 0
	AGC_100HZ   AgcControlT = 1
	AGC_50HZ    AgcControlT = 2
	AGC_5HZ     AgcControlT = 3
	AGC_CTRL_EN AgcControlT = 4
)

type AdsbModeT int32

const (
	ADSB_DECIMATION                  AdsbModeT = 0
	ADSB_NO_DECIMATION_LOWPASS       AdsbModeT = 1
	ADSB_NO_DECIMATION_BANDPASS_2MHZ AdsbModeT = 2
	ADSB_NO_DECIMATION_BANDPASS_3MHZ AdsbModeT = 3
)

// Control parameter structs
type DcOffsetT struct {
	DCenable uint8 // default: 1
	IQenable uint8 // default: 1
}

type DecimationT struct {
	Enable           uint8 // default: 0
	DecimationFactor uint8 // default: 1
	WideBandSignal   uint8 // default: 0
}

type AgcT struct {
	Enable             AgcControlT // default: AGC_50HZ
	SetPoint_dBfs      int32       // default: -60
	Attack_ms          uint16      // default: 0
	Decay_ms           uint16      // default: 0
	Decay_delay_ms     uint16      // default: 0
	Decay_threshold_dB uint16      // default: 0
	SyncUpdate         int32       // default: 0
}

type ControlParamsT struct {
	DcOffset   DcOffsetT
	Decimation DecimationT
	Agc        AgcT
	AdsbMode   AdsbModeT //default: ADSB_DECIMATION
}
