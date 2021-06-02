// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// +build cgo,!dll

package api

/*
#cgo !windows LDFLAGS: -lsdrplay_api
#cgo windows CFLAGS: -I "C:/Program Files/SDRplay/API/inc"
#cgo windows,amd64 LDFLAGS: -L "C:/Program Files/SDRplay/API/x64" -lsdrplay_api
#cgo windows,386 LDFLAGS: -L "C:/Program Files/SDRplay/API/x86" -lsdrplay_api


#include <stdint.h>
#include <sdrplay_api.h>

extern sdrplay_api_ErrT wrapper_api_Init(HANDLE dev);
*/
import "C"

import (
	"errors"
	"sync"
	"unsafe"
)

// Handle is a session handle provided by the C API.
type Handle unsafe.Pointer

var apiMutex sync.Mutex

type Impl struct{}

// Verify that Impl implements API.
var _ API = Impl{}

func GetAPI() Impl {
	return Impl{}
}

// Open implements API.
func (Impl) Open() error {
	apiMutex.Lock()
	defer apiMutex.Unlock()

	e := C.sdrplay_api_Open()
	et := ErrT(e)
	if et != Success {
		return et
	}

	return nil
}

// Close implements API.
func (Impl) Close() error {
	apiMutex.Lock()
	defer apiMutex.Unlock()

	e := C.sdrplay_api_Close()
	et := ErrT(e)
	if et != Success {
		return et
	}

	return nil
}

// ApiVersion implements SDRPlayAPI.
func (Impl) ApiVersion() (float32, error) {
	apiMutex.Lock()
	defer apiMutex.Unlock()

	var ver C.float
	e := C.sdrplay_api_ApiVersion(&ver)
	et := ErrT(e)
	if et != Success {
		return 0, et
	}

	return float32(ver), nil
}

// LockDeviceApi implements SDRPlayAPI.
func (Impl) LockDeviceApi() error {
	apiMutex.Lock()
	defer apiMutex.Unlock()

	e := C.sdrplay_api_LockDeviceApi()
	et := ErrT(e)
	if et != Success {
		return et
	}

	return nil
}

// UnlockDeviceApi implements SDRPlayAPI.
func (Impl) UnlockDeviceApi() error {
	apiMutex.Lock()
	defer apiMutex.Unlock()

	e := C.sdrplay_api_UnlockDeviceApi()
	et := ErrT(e)
	if et != Success {
		return et
	}

	return nil
}

// GetDevices implements SDRPlayAPI. Unlike the C function,
// sdrplay_api_GetDevices, GetDevices will do all necessary
// allocation.
func (Impl) GetDevices() ([]*DeviceT, error) {
	apiMutex.Lock()
	defer apiMutex.Unlock()

	devs := make([]C.sdrplay_api_DeviceT, MAX_DEVICES)
	var numDevs C.uint
	e := C.sdrplay_api_GetDevices(&devs[0], &numDevs, C.uint(len(devs)))

	et := ErrT(e)
	if et != Success {
		return nil, et
	}

	res := make([]*DeviceT, numDevs)
	for i := range res {
		cpy := *(*DeviceT)(unsafe.Pointer(&devs[i]))
		res[i] = &cpy
	}

	return res, nil
}

// SelectDevice implements SDRPlayAPI.
func (Impl) SelectDevice(dev *DeviceT) error {
	apiMutex.Lock()
	defer apiMutex.Unlock()

	e := C.sdrplay_api_SelectDevice((*C.sdrplay_api_DeviceT)(unsafe.Pointer(dev)))

	et := ErrT(e)
	if et != Success {
		return et
	}

	return nil
}

// ReleaseDevice implements SDRPlayAPI.
func (Impl) ReleaseDevice(dev *DeviceT) error {
	apiMutex.Lock()
	defer apiMutex.Unlock()

	e := C.sdrplay_api_ReleaseDevice((*C.sdrplay_api_DeviceT)(unsafe.Pointer(dev)))

	et := ErrT(e)
	if et != Success {
		return et
	}

	return nil
}

// GetLastError implements SDRPlayAPI.
func (Impl) GetLastError(dev *DeviceT) ErrorInfoT {
	apiMutex.Lock()
	defer apiMutex.Unlock()

	e := C.sdrplay_api_GetLastError((*C.sdrplay_api_DeviceT)(unsafe.Pointer(dev)))
	if e == nil {
		return ErrorInfoT{}
	}
	return *(*ErrorInfoT)(unsafe.Pointer(e))
}

// DisableHeartbeat implements SDRPlayAPI.
func (Impl) DisableHeartbeat() error {
	apiMutex.Lock()
	defer apiMutex.Unlock()

	e := C.sdrplay_api_DisableHeartbeat()
	et := ErrT(e)
	if et != Success {
		return et
	}

	return nil
}

// DebugEnable implements SDRPlayAPI.
func (Impl) DebugEnable(dev Handle, enable DbgLvlT) error {
	apiMutex.Lock()
	defer apiMutex.Unlock()

	e := C.sdrplay_api_DebugEnable(
		C.HANDLE(dev),
		C.sdrplay_api_DbgLvl_t(enable),
	)
	et := ErrT(e)
	if et != Success {
		return et
	}

	return nil
}

// LoadDeviceParams implements SDRPlayAPI.
func (Impl) LoadDeviceParams(dev Handle) (*DeviceParamsT, error) {
	apiMutex.Lock()
	defer apiMutex.Unlock()

	res := DeviceParamsT{}

	var ptr *C.sdrplay_api_DeviceParamsT
	e := C.sdrplay_api_GetDeviceParams(
		C.HANDLE(dev),
		&ptr,
	)

	et := ErrT(e)
	if et != Success {
		return &res, et
	}

	if ptr == nil {
		return &res, errors.New("got nil pointer from GetDeviceParams")
	}

	params := *(*DeviceParamsT)(unsafe.Pointer(ptr))
	if params.DevParams != nil {
		// Need to create a copy first because the compiler will
		// optimize away the seemingly redundant &*foo.
		cpy := *params.DevParams
		res.DevParams = &cpy
	}
	if params.RxChannelA != nil {
		// Need to create a copy first because the compiler will
		// optimize away the seemingly redundant &*foo.
		cpy := *params.RxChannelA
		res.RxChannelA = &cpy
	}
	if params.RxChannelB != nil {
		// Need to create a copy first because the compiler will
		// optimize away the seemingly redundant &*foo.
		cpy := *params.RxChannelB
		res.RxChannelB = &cpy
	}

	return &res, nil
}

// StoreDeviceParams implements SDRPlayAPI.
func (Impl) StoreDeviceParams(dev Handle, newParams *DeviceParamsT) error {
	apiMutex.Lock()
	defer apiMutex.Unlock()

	var ptr *C.sdrplay_api_DeviceParamsT
	e := C.sdrplay_api_GetDeviceParams(
		C.HANDLE(dev),
		&ptr,
	)

	et := ErrT(e)
	if et != Success {
		return et
	}

	if ptr == nil {
		return errors.New("got nil pointer from GetDeviceParams")
	}

	params := (*DeviceParamsT)(unsafe.Pointer(ptr))
	if params.DevParams != nil && newParams.DevParams != nil {
		*params.DevParams = *newParams.DevParams
	}
	if params.RxChannelA != nil && newParams.RxChannelA != nil {
		*params.RxChannelA = *newParams.RxChannelA
	}
	if params.RxChannelB != nil && newParams.RxChannelB != nil {
		*params.RxChannelB = *newParams.RxChannelB
	}

	return nil
}

// Init implements SDRPlayAPI.
func (Impl) Init(dev Handle, callbacks CallbackFnsT) error {
	apiMutex.Lock()
	defer apiMutex.Unlock()

	registerCallbacks(dev, callbacks)

	e := C.wrapper_api_Init(C.HANDLE(dev))
	et := ErrT(e)
	if et != Success {
		unregisterCallbacks(dev)
		return et
	}

	return nil
}

// Uninit implements SDRPlayAPI.
func (Impl) Uninit(dev Handle) error {
	apiMutex.Lock()
	defer apiMutex.Unlock()

	e := C.sdrplay_api_Uninit(C.HANDLE(dev))

	et := ErrT(e)
	if et != Success {
		return et
	}

	return nil
}

// Update implements SDRPlayAPI.
func (Impl) Update(dev Handle, tuner TunerSelectT, reasonForUpdate ReasonForUpdateT, reasonForUpdateExt1 ReasonForUpdateExtension1T) error {
	apiMutex.Lock()
	defer apiMutex.Unlock()

	e := C.sdrplay_api_Update(
		C.HANDLE(dev),
		C.sdrplay_api_TunerSelectT(tuner),
		C.sdrplay_api_ReasonForUpdateT(reasonForUpdate),
		C.sdrplay_api_ReasonForUpdateExtension1T(reasonForUpdateExt1),
	)

	et := ErrT(e)
	if et != Success {
		return et
	}

	return nil
}

// SwapRspDuoActiveTuner implements SDRPlayAPI.
func (Impl) SwapRspDuoActiveTuner(dev Handle, currentTuner *TunerSelectT, tuner1AmPortSel RspDuo_AmPortSelectT) error {
	apiMutex.Lock()
	defer apiMutex.Unlock()

	e := C.sdrplay_api_SwapRspDuoActiveTuner(
		C.HANDLE(dev),
		(*C.sdrplay_api_TunerSelectT)(unsafe.Pointer(currentTuner)),
		C.sdrplay_api_RspDuo_AmPortSelectT(tuner1AmPortSel),
	)

	et := ErrT(e)
	if et != Success {
		return et
	}

	return nil
}

// SwapRspDuoDualTunerModeSampleRate implements SDRPlayAPI.
func (Impl) SwapRspDuoDualTunerModeSampleRate(dev Handle, currentSampleRate *float64) error {
	apiMutex.Lock()
	defer apiMutex.Unlock()

	e := C.sdrplay_api_SwapRspDuoDualTunerModeSampleRate(
		C.HANDLE(dev),
		(*C.double)(unsafe.Pointer(currentSampleRate)),
	)

	et := ErrT(e)
	if et != Success {
		return et
	}

	return nil
}
