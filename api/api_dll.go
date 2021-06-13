// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// +build windows,!cgo windows,dll

package api

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Handle is a session handle provided by the C API. Using the
// windows package DLL loading, we are generally stuck with
// uintptr as a representation for data passed to and returned
// from functions.
type Handle uintptr

// Proc is an interface that is implemented by both windows.LazyProc
// and windows.Proc. The adaptive DLL loading implemented in the
// init() function means that we might be using either implementation
// at run-time.
type Proc interface {
	Call(a ...uintptr) (r1, r2 uintptr, lastErr error)
}

// References to C API functions are loaded in init().
var (
	sdrplay_api_Open                              Proc
	sdrplay_api_Close                             Proc
	sdrplay_api_ApiVersion                        Proc
	sdrplay_api_LockDeviceApi                     Proc
	sdrplay_api_UnlockDeviceApi                   Proc
	sdrplay_api_GetDevices                        Proc
	sdrplay_api_SelectDevice                      Proc
	sdrplay_api_ReleaseDevice                     Proc
	sdrplay_api_GetLastError                      Proc
	sdrplay_api_DisableHeartbeat                  Proc
	sdrplay_api_DebugEnable                       Proc
	sdrplay_api_GetDeviceParams                   Proc
	sdrplay_api_Init                              Proc
	sdrplay_api_Uninit                            Proc
	sdrplay_api_Update                            Proc
	sdrplay_api_SwapRspDuoActiveTuner             Proc
	sdrplay_api_SwapRspDuoDualTunerModeSampleRate Proc
)

var apiMutex sync.Mutex

// Impl is an implementation of API.
type Impl struct{}

// Verify that Impl implements API.
var _ API = Impl{}

// Wrap the static callback handler functions so they can
// be called from code inside the DLL.
var (
	streamACallbackWin = windows.NewCallback(streamACallback)
	streamBCallbackWin = windows.NewCallback(streamBCallback)
	eventCallbackWin   = windows.NewCallback(eventCallback)
)

// GetAPI returns an implementation of the API interface.
func GetAPI() Impl {
	return Impl{}
}

// Open implements API.
func (Impl) Open() error {
	apiMutex.Lock()
	defer apiMutex.Unlock()

	e, _, _ := sdrplay_api_Open.Call()
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

	e, _, _ := sdrplay_api_Close.Call()
	et := ErrT(e)
	if et != Success {
		return et
	}

	return nil
}

// ApiVersion implements API.
func (Impl) ApiVersion() (float32, error) {
	apiMutex.Lock()
	defer apiMutex.Unlock()

	var v float32
	e, _, _ := sdrplay_api_ApiVersion.Call(
		uintptr(unsafe.Pointer(&v)),
	)
	et := ErrT(e)
	if et != Success {
		return 0, et
	}
	return v, nil
}

// LockDeviceApi implements API.
func (Impl) LockDeviceApi() error {
	apiMutex.Lock()
	defer apiMutex.Unlock()

	e, _, _ := sdrplay_api_LockDeviceApi.Call()
	et := ErrT(e)
	if et != Success {
		return et
	}
	return nil
}

// UnlockDeviceApi implements API.
func (Impl) UnlockDeviceApi() error {
	apiMutex.Lock()
	defer apiMutex.Unlock()

	e, _, _ := sdrplay_api_UnlockDeviceApi.Call()
	et := ErrT(e)
	if et != Success {
		return et
	}
	return nil
}

// GetDevices implements API. Unlike the C function,
// sdrplay_api_GetDevices, GetDevices will do all necessary
// allocation.
func (Impl) GetDevices() ([]*DeviceT, error) {
	apiMutex.Lock()
	defer apiMutex.Unlock()

	devs := make([]DeviceT, MAX_DEVICES)
	var numDevs uint32
	e, _, _ := sdrplay_api_GetDevices.Call(
		uintptr(unsafe.Pointer(&devs[0])),
		uintptr(unsafe.Pointer(&numDevs)),
		uintptr(uint32(len(devs))),
	)

	et := ErrT(e)
	if et != Success {
		return nil, et
	}

	res := make([]*DeviceT, numDevs)
	for i := range res {
		cpy := devs[i]
		res[i] = &cpy
	}

	return res, nil
}

// SelectDevice implements API.
func (Impl) SelectDevice(dev *DeviceT) error {
	apiMutex.Lock()
	defer apiMutex.Unlock()

	e, _, _ := sdrplay_api_SelectDevice.Call(
		uintptr(unsafe.Pointer(dev)),
	)
	et := ErrT(e)
	if et != Success {
		return et
	}
	return nil
}

// ReleaseDevice implements API.
func (Impl) ReleaseDevice(dev *DeviceT) error {
	apiMutex.Lock()
	defer apiMutex.Unlock()

	e, _, _ := sdrplay_api_ReleaseDevice.Call(
		uintptr(unsafe.Pointer(dev)),
	)
	et := ErrT(e)
	if et != Success {
		return et
	}
	return nil
}

// GetLastError implements API.
func (Impl) GetLastError(dev *DeviceT) ErrorInfoT {
	apiMutex.Lock()
	defer apiMutex.Unlock()

	e, _, _ := sdrplay_api_GetLastError.Call(
		uintptr(unsafe.Pointer(dev)),
	)
	if e == 0 {
		return ErrorInfoT{}
	}
	// go vet doesn't like this, but it is the only way we can do
	// this since sdrplay_api_GetLastError actually returns a
	// pointer, but Proc returns uintptr.
	return *(*ErrorInfoT)(unsafe.Pointer(e))
}

// DisableHeartbeat implements API.
func (Impl) DisableHeartbeat() error {
	apiMutex.Lock()
	defer apiMutex.Unlock()

	e, _, _ := sdrplay_api_DisableHeartbeat.Call()
	et := ErrT(e)
	if et != Success {
		return et
	}
	return nil
}

// DebugEnable implements API.
func (Impl) DebugEnable(dev Handle, enable DbgLvlT) error {
	apiMutex.Lock()
	defer apiMutex.Unlock()

	e, _, _ := sdrplay_api_DebugEnable.Call(
		uintptr(dev),
		uintptr(enable),
	)
	et := ErrT(e)
	if et != Success {
		return et
	}
	return nil
}

// LoadDeviceParams implements API.
func (Impl) LoadDeviceParams(dev Handle) (*DeviceParamsT, error) {
	apiMutex.Lock()
	defer apiMutex.Unlock()

	res := DeviceParamsT{}

	var ptr unsafe.Pointer
	e, _, _ := sdrplay_api_GetDeviceParams.Call(
		uintptr(dev),
		uintptr(unsafe.Pointer(&ptr)),
	)
	et := ErrT(e)
	if et != Success {
		return nil, et
	}
	if ptr == nil {
		return nil, errors.New("got nil pointer from GetDeviceParams")
	}

	params := *(*DeviceParamsT)(ptr)
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

// StoreDeviceParams implements API.
func (Impl) StoreDeviceParams(dev Handle, newParams *DeviceParamsT) error {
	apiMutex.Lock()
	defer apiMutex.Unlock()

	var ptr unsafe.Pointer
	e, _, _ := sdrplay_api_GetDeviceParams.Call(
		uintptr(dev),
		uintptr(unsafe.Pointer(&ptr)),
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

// Init implements API.
func (Impl) Init(dev Handle, callbacks CallbackFnsT) error {
	apiMutex.Lock()
	defer apiMutex.Unlock()

	registerCallbacks(dev, callbacks)

	// Create an anonymous version of the CallbackFnsT struct.
	cCallbacks := struct {
		cbA uintptr
		cbB uintptr
		cbE uintptr
	}{
		streamACallbackWin,
		streamBCallbackWin,
		eventCallbackWin,
	}
	e, _, _ := sdrplay_api_Init.Call(
		uintptr(dev),
		uintptr(unsafe.Pointer(&cCallbacks)),
		uintptr(dev),
	)

	et := ErrT(e)
	if et != Success {
		unregisterCallbacks(dev)
		return et
	}

	return nil
}

// Uninit implements API.
func (Impl) Uninit(dev Handle) error {
	apiMutex.Lock()
	defer apiMutex.Unlock()

	unregisterCallbacks(dev)

	e, _, _ := sdrplay_api_Uninit.Call(uintptr(dev))

	et := ErrT(e)
	if et != Success {
		return et
	}

	return nil
}

// Update implements API.
func (Impl) Update(dev Handle, tuner TunerSelectT, reasonForUpdate ReasonForUpdateT, reasonForUpdateExt1 ReasonForUpdateExtension1T) error {
	apiMutex.Lock()
	defer apiMutex.Unlock()

	e, _, _ := sdrplay_api_Update.Call(
		uintptr(dev),
		uintptr(tuner),
		uintptr(reasonForUpdate),
		uintptr(reasonForUpdateExt1),
	)

	et := ErrT(e)
	if et != Success {
		return et
	}

	return nil
}

// SwapRspDuoActiveTuner implements API.
func (Impl) SwapRspDuoActiveTuner(dev Handle, currentTuner *TunerSelectT, tuner1AmPortSel RspDuo_AmPortSelectT) error {
	apiMutex.Lock()
	defer apiMutex.Unlock()

	e, _, _ := sdrplay_api_SwapRspDuoActiveTuner.Call(
		uintptr(dev),
		uintptr(unsafe.Pointer(currentTuner)),
		uintptr(tuner1AmPortSel),
	)

	et := ErrT(e)
	if et != Success {
		return et
	}

	return nil
}

// SwapRspDuoDualTunerModeSampleRate implements API.
func (Impl) SwapRspDuoDualTunerModeSampleRate(dev Handle, currentSampleRate *float64) error {
	apiMutex.Lock()
	defer apiMutex.Unlock()

	e, _, _ := sdrplay_api_SwapRspDuoDualTunerModeSampleRate.Call(
		uintptr(dev),
		uintptr(unsafe.Pointer(currentSampleRate)),
	)

	et := ErrT(e)
	if et != Success {
		return et
	}

	return nil
}

func init() {
	// First try to load the DLL by name using standard WIndows
	// library path resolution. This should find sdrplay_api.dll if
	// it is in one of the following locations.
	//
	// 	1. Current working directory.
	//	2. Same directory as executable.
	//	3. A directory included in the Path environment variable.
	//
	// Unlike Linux, where libraries are grouped into common directories
	// like /usr/lib or /usr/local/lib. It is less likely that a Windows user
	// has modified their environment variables to include random library
	// directories in "C:\Program Files".
	lazy := windows.NewLazyDLL("sdrplay_api")
	newProc := func(name string) Proc {
		return lazy.NewProc(name)
	}
	if err := lazy.Load(); err != nil {
		// We couldn't find sdrplay_api.dll through normal path resolution.
		// As a backup plan, we will try to load from an absolute path that
		// points to the standard location for the SDRPlay API installer to
		// install sdrplay_api.dll. DLL_PATH is different for 32-bit and
		// 64-bit architectures.
		//
		// For security reasons, the LoadDLL function should only be used
		// with absolute paths.
		if !filepath.IsAbs(DLL_PATH) {
			panic(fmt.Sprintf("DLL_PATH is not absolute, refusing to load DLL; %s", DLL_PATH))
		}
		direct, err := windows.LoadDLL(DLL_PATH)
		if err != nil {
			// There is nothing else we can do at this point.
			panic(fmt.Sprintf("sdrplay_api.dll not found in Path or %s", DLL_PATH))
		}
		newProc = func(name string) Proc {
			// Ignore the error because that produces the same behavior
			// as if we were using the lazy loaded DLL. Specifically,
			// we will panic when a missing Proc is used. The Swap*
			// functions don't appear in older API versions, so this
			// is also a somewhat graceful way of being backwards
			// compatible.
			proc, _ := direct.FindProc(name)
			return proc
		}
	}

	sdrplay_api_Open = newProc("sdrplay_api_Open")
	sdrplay_api_Close = newProc("sdrplay_api_Open")
	sdrplay_api_ApiVersion = newProc("sdrplay_api_ApiVersion")
	sdrplay_api_LockDeviceApi = newProc("sdrplay_api_LockDeviceApi")
	sdrplay_api_UnlockDeviceApi = newProc("sdrplay_api_UnlockDeviceApi")
	sdrplay_api_GetDevices = newProc("sdrplay_api_GetDevices")
	sdrplay_api_SelectDevice = newProc("sdrplay_api_SelectDevice")
	sdrplay_api_ReleaseDevice = newProc("sdrplay_api_ReleaseDevice")
	sdrplay_api_GetLastError = newProc("sdrplay_api_GetLastError")
	sdrplay_api_DisableHeartbeat = newProc("sdrplay_api_DisableHeartbeat")
	sdrplay_api_DebugEnable = newProc("sdrplay_api_DebugEnable")
	sdrplay_api_GetDeviceParams = newProc("sdrplay_api_GetDeviceParams")
	sdrplay_api_Init = newProc("sdrplay_api_Init")
	sdrplay_api_Uninit = newProc("sdrplay_api_Uninit")
	sdrplay_api_Update = newProc("sdrplay_api_Update")
	sdrplay_api_SwapRspDuoActiveTuner = newProc("sdrplay_api_SwapRspDuoActiveTuner")
	sdrplay_api_SwapRspDuoDualTunerModeSampleRate = newProc("sdrplay_api_SwapRspDuoDualTunerModeSampleRate")
}
