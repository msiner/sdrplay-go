// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package api

//go:generate go run golang.org/x/tools/cmd/stringer -type HWVersion,ErrT,ReasonForUpdateT,ReasonForUpdateExtension1T,DbgLvlT -output api_string.go

const (
	MAX_DEVICES           = 16
	MAX_TUNERS_PER_DEVICE = 2

	MAX_SER_NO_LEN  = 64
	MAX_ROOT_NM_LEN = 32
)

type HWVersion uint8

const (
	RSP1_ID   HWVersion = 1
	RSP1A_ID  HWVersion = 255
	RSP2_ID   HWVersion = 2
	RSPduo_ID HWVersion = 3
	RSPdx_ID  HWVersion = 4
)

type ErrT int32

const (
	Success               ErrT = 0
	Fail                  ErrT = 1
	InvalidParam          ErrT = 2
	OutOfRange            ErrT = 3
	GainUpdateError       ErrT = 4
	RfUpdateError         ErrT = 5
	FsUpdateError         ErrT = 6
	HwError               ErrT = 7
	AliasingError         ErrT = 8
	AlreadyInitialised    ErrT = 9
	NotInitialised        ErrT = 10
	NotEnabled            ErrT = 11
	HwVerError            ErrT = 12
	OutOfMemError         ErrT = 13
	ServiceNotResponding  ErrT = 14
	StartPending          ErrT = 15
	StopPending           ErrT = 16
	InvalidMode           ErrT = 17
	FailedVerification1   ErrT = 18
	FailedVerification2   ErrT = 19
	FailedVerification3   ErrT = 20
	FailedVerification4   ErrT = 21
	FailedVerification5   ErrT = 22
	FailedVerification6   ErrT = 23
	InvalidServiceVersion ErrT = 24
)

func (e ErrT) Error() string {
	return e.String()
}

type ReasonForUpdateT uint32

const (
	Update_None ReasonForUpdateT = 0x00000000

	// Reasons for master only mode
	Update_Dev_Fs         ReasonForUpdateT = 0x00000001
	Update_Dev_Ppm        ReasonForUpdateT = 0x00000002
	Update_Dev_SyncUpdate ReasonForUpdateT = 0x00000004
	Update_Dev_ResetFlags ReasonForUpdateT = 0x00000008

	Update_Rsp1a_BiasTControl      ReasonForUpdateT = 0x00000010
	Update_Rsp1a_RfNotchControl    ReasonForUpdateT = 0x00000020
	Update_Rsp1a_RfDabNotchControl ReasonForUpdateT = 0x00000040

	Update_Rsp2_BiasTControl   ReasonForUpdateT = 0x00000080
	Update_Rsp2_AmPortSelect   ReasonForUpdateT = 0x00000100
	Update_Rsp2_AntennaControl ReasonForUpdateT = 0x00000200
	Update_Rsp2_RfNotchControl ReasonForUpdateT = 0x00000400
	Update_Rsp2_ExtRefControl  ReasonForUpdateT = 0x00000800

	Update_RspDuo_ExtRefControl ReasonForUpdateT = 0x00001000

	Update_Master_Spare_1 ReasonForUpdateT = 0x00002000
	Update_Master_Spare_2 ReasonForUpdateT = 0x00004000

	// Reasons for master and slave mode
	// Note: Update_Tuner_Gr MUST be the first value defined in this section!
	Update_Tuner_Gr       ReasonForUpdateT = 0x00008000
	Update_Tuner_GrLimits ReasonForUpdateT = 0x00010000
	Update_Tuner_Frf      ReasonForUpdateT = 0x00020000
	Update_Tuner_BwType   ReasonForUpdateT = 0x00040000
	Update_Tuner_IfType   ReasonForUpdateT = 0x00080000
	Update_Tuner_DcOffset ReasonForUpdateT = 0x00100000
	Update_Tuner_LoMode   ReasonForUpdateT = 0x00200000

	Update_Ctrl_DCoffsetIQimbalance ReasonForUpdateT = 0x00400000
	Update_Ctrl_Decimation          ReasonForUpdateT = 0x00800000
	Update_Ctrl_Agc                 ReasonForUpdateT = 0x01000000
	Update_Ctrl_AdsbMode            ReasonForUpdateT = 0x02000000
	Update_Ctrl_OverloadMsgAck      ReasonForUpdateT = 0x04000000

	Update_RspDuo_BiasTControl         ReasonForUpdateT = 0x08000000
	Update_RspDuo_AmPortSelect         ReasonForUpdateT = 0x10000000
	Update_RspDuo_Tuner1AmNotchControl ReasonForUpdateT = 0x20000000
	Update_RspDuo_RfNotchControl       ReasonForUpdateT = 0x40000000
	Update_RspDuo_RfDabNotchControl    ReasonForUpdateT = 0x80000000
)

type ReasonForUpdateExtension1T int32

const (
	Update_Ext1_None ReasonForUpdateExtension1T = 0x00000000

	// Reasons for master only mode
	Update_RspDx_HdrEnable         ReasonForUpdateExtension1T = 0x00000001
	Update_RspDx_BiasTControl      ReasonForUpdateExtension1T = 0x00000002
	Update_RspDx_AntennaControl    ReasonForUpdateExtension1T = 0x00000004
	Update_RspDx_RfNotchControl    ReasonForUpdateExtension1T = 0x00000008
	Update_RspDx_RfDabNotchControl ReasonForUpdateExtension1T = 0x00000010
	Update_RspDx_HdrBw             ReasonForUpdateExtension1T = 0x00000020

	// Reasons for master and slave mode
)

type DbgLvlT int32

const (
	DbgLvl_Disable DbgLvlT = 0
	DbgLvl_Verbose DbgLvlT = 1
	DbgLvl_Warning DbgLvlT = 2
	DbgLvl_Error   DbgLvlT = 3
	DbgLvl_Message DbgLvlT = 4
)

// bytesToString converts a slice of bytes into a string. The bytes
// are interpreted as a zero terminated C string buffer.
func bytesToString(buf []byte) string {
	end := len(buf)
	for i := 0; i < end; i++ {
		if buf[i] == 0 {
			end = i
			break
		}
	}
	return string(buf[:end])
}

type SerialNumber [MAX_SER_NO_LEN]byte

func ParseSerialNumber(s string) SerialNumber {
	var res SerialNumber
	copy(res[:], []byte(s))
	return res
}

func (s SerialNumber) String() string {
	return bytesToString(s[:])
}

func (s SerialNumber) GoString() string {
	return bytesToString(s[:])
}

type Message256 [256]byte

func (m Message256) String() string {
	return bytesToString(m[:])
}

type Message1024 [1024]byte

func (m Message1024) String() string {
	return bytesToString(m[:])
}

type DeviceT struct {
	SerNo            SerialNumber
	HWVer            HWVersion
	Tuner            TunerSelectT
	RspDuoMode       RspDuoModeT
	RspDuoSampleFreq float64
	Dev              Handle
}

type DeviceParamsT struct {
	DevParams  *DevParamsT
	RxChannelA *RxChannelParamsT
	RxChannelB *RxChannelParamsT
}

type ErrorInfoT struct {
	File     Message256
	Function Message256
	Line     int32
	Message  Message1024
}
