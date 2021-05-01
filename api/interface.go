// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package api

// API is an interface wrapper around the entire sdrplay_api. There are two
// reasons for defining this interface.
// 1. With 2 separate implementations (cgo and dll), this interface is used
//    to verify the implementations match.
// 2. Allows code that uses the API to be tested indepent of an
//    implementation or using a mock of the API.
type API interface {
	// Open is a wrapper around sdrplay_api_Open(). Begins API access.
	// Returns an error if called more than once before Close() is called.
	Open() error

	// Close is a wrapper around sdrplay_api_Close(). Closes API access.
	// Returns an error if Open() has not been successfully called or if
	// Close() has already been called since the last Open() call.
	Close() error

	// ApiVersion is a wrapper around sdrplay_api_ApiVersion(). Unlike
	// the C function, ApiVersion simply returns the version number float.
	ApiVersion() (float32, error)

	// LockDeviceApi is a wrapper around sdrplay_api_LockDeviceApi().
	LockDeviceApi() error

	// UnlockDeviceApi is a wrapper around sdrplay_api_UnlockDeviceApi().
	UnlockDeviceApi() error

	// GetDevices is a wrapper around sdrplay_api_GetDevices().
	// Unlike the C function, GetDevices does all necessary allocation.
	GetDevices() ([]*DeviceT, error)

	// SelectDevice is a wrapper around sdrplay_api_SelectDevice().
	SelectDevice(dev *DeviceT) error

	// ReleaseDevice is a wrapper around sdrplay_api_ReleaseDevice().
	ReleaseDevice(dev *DeviceT) error

	// GetLastError is a wrapper around sdrplay_api_GetLastError(). Unlike
	// the C funcion, it simply returns an error info struct that is a copy
	// of a pointer to an API-internal statically allocated struct.
	GetLastError(dev *DeviceT) ErrorInfoT

	// DisableHeartbeat is a wrapper around sdrplay_api_DisableHeartbeat().
	DisableHeartbeat() error

	// DebugEnable is a wrapper around sdrplay_api_DebugEnable().
	DebugEnable(dev Handle, enable DbgLvlT) error

	// LoadDeviceParams is the first half of the wrapper around
	// sdrplay_api_GetDeviceParams(). That C function returns a pointer to
	// an API-internal struct and modifications can be made to that
	// struct without any commit step. To keep the memory between C and
	// Go separate, this function returns a copy of that struct.
	// Modifications can be made to that copy and then committed with
	// StoreDeviceParams().
	LoadDeviceParams(dev Handle) (*DeviceParamsT, error)

	// StoreDeviceParams is the second half of the wrapper around
	// sdrplay_api_GetDeviceParams(). That C function returns a pointer to
	// an API-internal struct and modifications can be made to that
	// struct without any commit step. To keep the memory between C and
	// Go separate, LoadDeviceParams is used to get a copy of that struct.
	// Modifications can be made to that copy and then committed with
	// this function.
	StoreDeviceParams(dev Handle, params *DeviceParamsT) error

	// Init is a wrapper around sdrplay_api_Init(). Unlike the C version,
	// there is no cbContext argument. In Go, this is unneccessary because
	// callbacks can be closures or methods that have bound state.
	Init(dev Handle, callbacks CallbackFnsT) error

	// Uninit is a wrapper around sdrplay_api_Uninit().
	Uninit(dev Handle) error

	// Update is a wrapper around sdrplay_api_Update().
	Update(dev Handle, tuner TunerSelectT, reasonForUpdate ReasonForUpdateT, reasonForUpdateExt1 ReasonForUpdateExtension1T) error

	// SwapRspDuoActiveTuner is a wrapper around
	// sdrplay_api_SwapRspDuoActiveTuner().
	SwapRspDuoActiveTuner(dev Handle, currentTuner *TunerSelectT, tuner1AmPortSel RspDuo_AmPortSelectT) error

	// SwapRspDuoDualTunerModeSampleRate is a wrapper around
	// sdrplay_api_SwapRspDuoDualTunerModeSampleRate().
	SwapRspDuoDualTunerModeSampleRate(dev Handle, currentSampleRate *float64) error
}
