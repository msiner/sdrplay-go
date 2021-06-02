// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package session

import (
	"errors"
	"fmt"

	"github.com/msiner/sdrplay-go/api"
)

// DevConfigFn is a function type for configuring a DeviceParamsT. At the
// point that a DevConfigFn is called, the DeviceT has already been
// selected. So the function should not make any changes to the DeviceT.
// It is provided so the function can configure the DeviceParamsT corrrectly
// for the selected hardware, mode, tuner, and input. The function returns
// a non-nil error if an incompatible or impossible configuration is
// detected or requested.
type DevConfigFn func(d *api.DeviceT, p *api.DeviceParamsT) error

// NoopDevConfig is a device configuration function that returns nil
// without checking or modifying the params. It can be used as a noop
// or placeholder for another DevConfigFn function.
func NoopDevConfig(d *api.DeviceT, p *api.DeviceParamsT) error {
	return nil
}

// WithTransferMode creates a function that sets the USB transfer mode
// as specified (isochronous or bulk).
func WithTransferMode(mode api.TransferModeT) DevConfigFn {
	return func(d *api.DeviceT, p *api.DeviceParamsT) error {
		p.DevParams.Mode = mode
		return nil
	}
}

// WithRefClockOutput creates a function that enables or disables the
// reference clock output on an RSP2 or RSPduo device. It will have no
// effect on other device types.
func WithRefClockOutput(enabled bool) DevConfigFn {
	return func(d *api.DeviceT, p *api.DeviceParamsT) error {
		var val uint8
		if enabled {
			val = 1
		}
		switch d.HWVer {
		case api.RSP1_ID:
			// not available
		case api.RSP1A_ID:
			// not available
		case api.RSP2_ID:
			p.DevParams.Rsp2Params.ExtRefOutputEn = val
		case api.RSPduo_ID:
			p.DevParams.RspDuoParams.ExtRefOutputEn = int32(val)
		case api.RSPdx_ID:
			// not available
		}
		return nil
	}
}

// GetHighZPortEnabled returns true if the High-Z port is enabled or false
// if it is not enabled or if the device does not have a High-Z port.
func GetHighZPortEnabled(d *api.DeviceT, p *api.DeviceParamsT) bool {
	switch d.HWVer {
	case api.RSP2_ID:
		return p.RxChannelA.Rsp2TunerParams.AmPortSel == api.Rsp2_AMPORT_1
	case api.RSPduo_ID:
		switch d.Tuner {
		case api.Tuner_A, api.Tuner_Both:
			return p.RxChannelA.RspDuoTunerParams.Tuner1AmPortSel == api.RspDuo_AMPORT_1
		default:
			return false
		}
	default:
		return false
	}
}

// SetHighZPortEnabled configures the devices High-Z port to enabled or
// disabled as specified by the en argument. If the device does not have
// a High-Z port, this function does nothing. If the device is an RSPduo
// and tuner A or both tuners are selected, it will configure tuner A to
// use the High-Z port. If the device is an RSPduo, but only tuner B is
// selected, it returns an error.
func SetHighZPortEnabled(d *api.DeviceT, p *api.DeviceParamsT, en bool) error {
	switch d.HWVer {
	case api.RSP1_ID:
		// not available
	case api.RSP1A_ID:
		// not available
	case api.RSP2_ID:
		port := api.Rsp2_AMPORT_2
		if en {
			port = api.Rsp2_AMPORT_1
		}
		p.RxChannelA.Rsp2TunerParams.AmPortSel = port
	case api.RSPduo_ID:
		port := api.RspDuo_AMPORT_2
		if en {
			port = api.RspDuo_AMPORT_1
		}
		switch d.Tuner {
		case api.Tuner_A, api.Tuner_Both:
			p.RxChannelA.RspDuoTunerParams.Tuner1AmPortSel = port
		case api.Tuner_B:
			if port == api.RspDuo_AMPORT_1 {
				return errors.New("cannot select High-Z port for tuner B")
			}
		}
	case api.RSPdx_ID:
		// not available
	}
	return nil
}

// WithHighZPortEnabled creates a function that enables or disables
// use of the high-z port on the RSP2 and RSPduo. It will return an
// error if an RSPduo is being used, but only tuner B is selected.
// The function has no effect on other devices.
func WithHighZPortEnabled(en bool) DevConfigFn {
	return func(d *api.DeviceT, p *api.DeviceParamsT) error {
		return SetHighZPortEnabled(d, p, en)
	}
}

// WithDxAntennaSelect creates a function that configures the antenna
// used on RSPdx devices. The function has no effect on other devices.
func WithDxAntennaSelect(ant api.RspDx_AntennaSelectT) DevConfigFn {
	return func(d *api.DeviceT, p *api.DeviceParamsT) error {
		switch ant {
		case api.RspDx_ANTENNA_A, api.RspDx_ANTENNA_B, api.RspDx_ANTENNA_C:
			// good
		default:
			return fmt.Errorf("invalid RSPdx antenna selection; got %v", ant)
		}
		if d.HWVer == api.RSPdx_ID {
			p.DevParams.RspDxParams.AntennaSel = ant
		}
		return nil
	}
}

// WithRsp2AntennaSelect creates a function that configures the antenna
// used on RSP2 devices. The function has no effect on other devices.
func WithRsp2AntennaSelect(ant api.Rsp2_AntennaSelectT) DevConfigFn {
	return func(d *api.DeviceT, p *api.DeviceParamsT) error {
		switch ant {
		case api.Rsp2_ANTENNA_A, api.Rsp2_ANTENNA_B:
			// good
		default:
			return fmt.Errorf("invalid RSP2 antenna selection; got %v", ant)
		}
		if d.HWVer == api.RSP2_ID {
			p.RxChannelA.Rsp2TunerParams.AntennaSel = ant
		}
		return nil
	}
}
