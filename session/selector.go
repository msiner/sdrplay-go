// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package session

import (
	"github.com/msiner/sdrplay-go/api"
)

// DevSelectFn is a function that selects a single device out of
// a list of devices or returns nil if none are suitable. This type
// is not meant to be implemented directly. Instead it is implemented
// internally by providing a list of DevFilterFn filters to
// WithSelector().
type DevSelectFn func(devs []*api.DeviceT) *api.DeviceT

// DevFilterFn is a function that selects a subset of devices
// out of a list of devices or returns nil or an empty slice
// if none are suitable. It is used by WithSelector() to create
// a DevSelectFn that calls the filters and then selects the
// first remaining device.
type DevFilterFn func(devs []*api.DeviceT) []*api.DeviceT

// WithAnyDevice creates a filter function that accepts
// any device. It can be used as a noop or placeholder for
// another function.
func WithNoopDevFilter() DevFilterFn {
	return func(devs []*api.DeviceT) []*api.DeviceT {
		return devs
	}
}

// WithSerials creates a device filter function that keeps only
// devices with matching one of the serial numbers provided.
//
// Examples:
// Select an exact serial number:
// 		WithSerials(api.ParseSerialNumber("1905015632"))
func WithSerials(vals ...api.SerialNumber) DevFilterFn {
	return func(devs []*api.DeviceT) []*api.DeviceT {
		var res []*api.DeviceT
		for _, dev := range devs {
			for _, val := range vals {
				if dev.SerNo == val {
					res = append(res, dev)
				}
			}
		}
		return res
	}
}

// WithModels creates a device filter function that keeps only
// devices with hardware versions matching the hardware version
// IDs provided.
//
// Examples:
// Get only RSPdx and RSP1A hardware:
// 		WithModels(api.RSPdx_ID, api.RS1A_ID)
// Get only RSPduo hardware:
// 		WithModels(api.RSPduo_ID)
func WithModels(vals ...api.HWVersion) DevFilterFn {
	return func(devs []*api.DeviceT) []*api.DeviceT {
		var res []*api.DeviceT
		for _, dev := range devs {
			for _, val := range vals {
				if dev.HWVer == val {
					res = append(res, dev)
				}
			}
		}
		return res
	}
}

// WithRSP1 is a wrapper around WithModels() to select only RSP1 hardware.
func WithRSP1() DevFilterFn {
	return WithModels(api.RSP1_ID)
}

// WithRSP1A is a wrapper around WithModels() to select only RSP1A hardware.
func WithRSP1A() DevFilterFn {
	return WithModels(api.RSP1A_ID)
}

// WithRSP2 is a wrapper around WithModels() to select only RSP2 hardware.
func WithRSP2() DevFilterFn {
	return WithModels(api.RSP2_ID)
}

// WithRSPduo is a wrapper around WithModels() to select only RSPduo hardware.
func WithRSPduo() DevFilterFn {
	return WithModels(api.RSPduo_ID)
}

// WithRSPdx is a wrapper around WithModels() to select only RSPdx hardware.
func WithRSPdx() DevFilterFn {
	return WithModels(api.RSPdx_ID)
}

// WithDuoTunerA creates a device filter function that filters out any
// RSPduo hardware that does not have tuner A available. That is, the
// DeviceT.Tuner field is either Tuner_A or Tuner_Both. If it is
// Tuner_Both, the function will set it to Tuner_A.
//
// Note that the only devices the filter function will remove are RSPduo
// hardware without Tuner_A available. Non-RSPduo hardware will pass
// unaffected. To filter on hardware use WithModel() or one of the
// hardware-specific wrapper functions.
func WithDuoTunerA() DevFilterFn {
	return func(devs []*api.DeviceT) []*api.DeviceT {
		var res []*api.DeviceT
		for _, dev := range devs {
			if dev.HWVer != api.RSPduo_ID {
				res = append(res, dev)
				continue
			}
			switch dev.Tuner {
			case api.Tuner_A:
				res = append(res, dev)
			case api.Tuner_Both:
				dev.Tuner = api.Tuner_A
				res = append(res, dev)
			}
		}
		return res
	}
}

// WithDuoTunerB creates a device filter function that filters out any
// RSPduo hardware that does not have tuner B available. That is, the
// DeviceT.Tuner field is either Tuner_B or Tuner_Both. If it is
// Tuner_Both, the function will set it to Tuner_B.
//
// Note that the only devices the filter function will remove are RSPduo
// hardware without Tuner_B available. Non-RSPduo hardware will pass
// unaffected. To filter on hardware use WithModel() or one of the
// hardware-specific wrapper functions.
func WithDuoTunerB() DevFilterFn {
	return func(devs []*api.DeviceT) []*api.DeviceT {
		var res []*api.DeviceT
		for _, dev := range devs {
			if dev.HWVer != api.RSPduo_ID {
				res = append(res, dev)
				continue
			}
			switch dev.Tuner {
			case api.Tuner_B:
				res = append(res, dev)
			case api.Tuner_Both:
				dev.Tuner = api.Tuner_B
				res = append(res, dev)
			}
		}
		return res
	}
}

// WithDuoTunerBoth creates a device filter function that filters out any
// RSPduo hardware that does not have both tuners available. That is, the
// DeviceT.Tuner field must be Tuner_Both.
//
// Note that the only devices the filter function will remove are RSPduo
// hardware without both tuners available. Non-RSPduo hardware will pass
// unaffected. To filter on hardware use WithModel() or one of the
// hardware-specific wrapper functions.
func WithDuoTunerBoth() DevFilterFn {
	return func(devs []*api.DeviceT) []*api.DeviceT {
		var res []*api.DeviceT
		for _, dev := range devs {
			if dev.HWVer != api.RSPduo_ID {
				res = append(res, dev)
				continue
			}
			if dev.Tuner == api.Tuner_Both {
				res = append(res, dev)
			}
		}
		return res
	}
}

// WithDuoTunerEither creates a device filter function that selects either
// available tuner on RSPduo hardware. That is, it will set the
// DeviceT.Tuner field to either Tuner_A or Tuner_B. If it is already set
// Tuner_Both, the function will set it to Tuner_A. Otherwise, it will
// leave it set to either Tuner_A or Tuner_B.
//
// Note that the filter function will not remove any hardware. If an
// RSPduo device is available, it will have at least one tuner available.
func WithDuoTunerEither() DevFilterFn {
	return func(devs []*api.DeviceT) []*api.DeviceT {
		var res []*api.DeviceT
		for _, dev := range devs {
			if dev.HWVer != api.RSPduo_ID {
				res = append(res, dev)
				continue
			}
			switch dev.Tuner {
			case api.Tuner_A, api.Tuner_B:
				res = append(res, dev)
			case api.Tuner_Both:
				dev.Tuner = api.Tuner_A
				res = append(res, dev)
			}
		}
		return res
	}
}

// WithDuoModeSingle creates a device filter function that filters out any
// RSPduo hardware that is not available for single-tuner mode. That is,
// if the DeviceT.RspDuoMode field has the RspDuoMode_Single_Tuner flag
// set, it will clear all of the other flags in order to select
// single-tuner mode.
//
// Note that the only devices the filter function will remove are RSPduo
// hardware without single-tuner mode available. Non-RSPduo hardware will
// pass unaffected. To filter on hardware use WithModel() or one of the
// hardware-specific wrapper functions.
func WithDuoModeSingle() DevFilterFn {
	return func(devs []*api.DeviceT) []*api.DeviceT {
		var res []*api.DeviceT
		for _, dev := range devs {
			if dev.HWVer != api.RSPduo_ID {
				res = append(res, dev)
				continue
			}
			if dev.RspDuoMode&api.RspDuoMode_Single_Tuner != 0 {
				dev.RspDuoMode = api.RspDuoMode_Single_Tuner
				res = append(res, dev)
			}
		}
		return res
	}
}

// WithDuoModeDual creates a device filter function that filters out any
// RSPduo hardware that is not available for dual-tuner mode. That is,
// if the DeviceT.RspDuoMode field has the RspDuoMode_Dual_Tuner flag
// set, it will clear all of the other flags in order to select
// dual-tuner mode.
//
// Note that the only devices the filter function will remove are RSPduo
// hardware without dual-tuner mode available. Non-RSPduo hardware will
// pass unaffected. To filter on hardware use WithModel() or one of the
// hardware-specific wrapper functions.
func WithDuoModeDual(maxFs bool) DevFilterFn {
	return func(devs []*api.DeviceT) []*api.DeviceT {
		var res []*api.DeviceT
		for _, dev := range devs {
			if dev.HWVer != api.RSPduo_ID {
				res = append(res, dev)
				continue
			}
			if dev.RspDuoMode&api.RspDuoMode_Dual_Tuner != 0 {
				dev.RspDuoMode = api.RspDuoMode_Dual_Tuner
				switch maxFs {
				case true:
					dev.RspDuoSampleFreq = 8e6
				default:
					dev.RspDuoSampleFreq = 6e6
				}
				res = append(res, dev)
			}
		}
		return res
	}
}

// WithDuoModeShared creates a device filter function that filters out any
// RSPduo hardware that is not available for shared mode as the primary or
// secondary. That is, if the DeviceT.RspDuoMode field has either the
// RspDuoMode_Primary or RspDuoMode_Secondary flag set. If primary is
// available it will set the field to RspDuoMode_Primary to select shared
// mode as primary. Otherwise, it will set the field to RspDuoMode_Secondary
// to select shared mode as secondary.
//
// Note that the only devices the filter function will remove are RSPduo
// hardware without shared, primary or secondary mode available. Non-RSPduo
// hardware will pass unaffected. To filter on hardware use WithModel() or
// one of the hardware-specific wrapper functions.
func WithDuoModeShared(maxFs bool) DevFilterFn {
	return func(devs []*api.DeviceT) []*api.DeviceT {
		var res []*api.DeviceT
		for _, dev := range devs {
			if dev.HWVer != api.RSPduo_ID {
				res = append(res, dev)
				continue
			}
			switch {
			case dev.RspDuoMode&api.RspDuoMode_Primary != 0:
				dev.RspDuoMode = api.RspDuoMode_Primary
				switch maxFs {
				case true:
					dev.RspDuoSampleFreq = 8e6
				default:
					dev.RspDuoSampleFreq = 6e6
				}
				res = append(res, dev)
			case dev.RspDuoMode&api.RspDuoMode_Secondary != 0:
				dev.RspDuoMode = api.RspDuoMode_Secondary
				switch maxFs {
				case true:
					if dev.RspDuoSampleFreq == 8e6 {
						res = append(res, dev)
					}
				default:
					if dev.RspDuoSampleFreq == 6e6 {
						res = append(res, dev)
					}
				}
			}
		}
		return res
	}
}

// WithDuoModePrimary creates a device filter function that filters out any
// RSPduo hardware that is not available for shared mode as the primary.
// That is, if the DeviceT.RspDuoMode field has the RspDuoMode_Primary flag
// set, it will clear all of the other flags in order to select
// shared mode as primary.
//
// Note that the only devices the filter function will remove are RSPduo
// hardware without shared, primary mode available. Non-RSPduo hardware will
// pass unaffected. To filter on hardware use WithModel() or one of the
// hardware-specific wrapper functions.
func WithDuoModePrimary(maxFs bool) DevFilterFn {
	return func(devs []*api.DeviceT) []*api.DeviceT {
		var res []*api.DeviceT
		for _, dev := range devs {
			if dev.HWVer != api.RSPduo_ID {
				res = append(res, dev)
				continue
			}
			if dev.RspDuoMode&api.RspDuoMode_Primary != 0 {
				dev.RspDuoMode = api.RspDuoMode_Primary
				switch maxFs {
				case true:
					dev.RspDuoSampleFreq = 8e6
				default:
					dev.RspDuoSampleFreq = 6e6
				}
				res = append(res, dev)
			}
		}
		return res
	}
}

// WithDuoModeSecondary creates a device filter function that filters out any
// RSPduo hardware that is not available for shared mode as the secondary.
// That is, if the DeviceT.RspDuoMode field has the RspDuoMode_Secondary flag
// set, it will clear all of the other flags in order to select
// shared mode as secondary.
//
// Note that the only devices the filter function will remove are RSPduo
// hardware without shared, secondary mode available. Non-RSPduo hardware will
// pass unaffected. To filter on hardware use WithModel() or one of the
// hardware-specific wrapper functions.
func WithDuoModeSecondary(maxFs bool) DevFilterFn {
	return func(devs []*api.DeviceT) []*api.DeviceT {
		var res []*api.DeviceT
		for _, dev := range devs {
			if dev.HWVer != api.RSPduo_ID {
				res = append(res, dev)
				continue
			}
			if dev.RspDuoMode&api.RspDuoMode_Secondary != 0 {
				dev.RspDuoMode = api.RspDuoMode_Secondary
				switch maxFs {
				case true:
					if dev.RspDuoSampleFreq == 8e6 {
						res = append(res, dev)
					}
				default:
					if dev.RspDuoSampleFreq == 6e6 {
						res = append(res, dev)
					}
				}
			}
		}
		return res
	}
}
