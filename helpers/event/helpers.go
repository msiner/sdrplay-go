// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package event

import (
	"fmt"

	"github.com/msiner/sdrplay-go/api"
	"github.com/msiner/sdrplay-go/session"
)

// Logger is compatible with standard library and logrus.
type Logger interface {
	Printf(format string, v ...interface{})
}

// GetRelatedParams is a helper function to select and extract the
// device/channel parameters related to the event information.
func GetRelatedParams(d *api.DeviceT, a api.API, e api.EventT, t api.TunerSelectT, p *api.EventParamsT) (*api.DeviceParamsT, []*api.RxChannelParamsT, error) {
	devParams, err := a.LoadDeviceParams(d.Dev)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get devparams; %v: %v", err, a.GetLastError(d))
	}

	var rxParams []*api.RxChannelParamsT
	switch d.HWVer {
	case api.RSPduo_ID:
		switch t {
		case api.Tuner_A:
			rxParams = append(rxParams, devParams.RxChannelA)
		case api.Tuner_B:
			rxParams = append(rxParams, devParams.RxChannelB)
		case api.Tuner_Both:
			rxParams = append(rxParams, devParams.RxChannelA, devParams.RxChannelB)
		}
	default:
		rxParams = append(rxParams, devParams.RxChannelA)
	}

	return devParams, rxParams, nil
}

// GetRelatedParamsMsg is a wrapper around GetRelatedParams for easy
// use with EventChan.
func GetRelatedParamsMsg(d *api.DeviceT, a api.API, evt Msg) (*api.DeviceParamsT, []*api.RxChannelParamsT, error) {
	return GetRelatedParams(d, a, evt.EventID, evt.Tuner, &evt.Params)
}

// HandlePowerOverloadChange is a helper function to automatically handle
// power overload events by reducing gain, if possible, on the affected
// channel.
func HandlePowerOverloadChange(d *api.DeviceT, a api.API, e api.EventT, t api.TunerSelectT, p *api.EventParamsT, lg Logger, adjust bool) error {
	if e != api.PowerOverloadChange {
		return nil
	}

	powParams := p.PowerOverloadParams

	if lg != nil {
		lg.Printf("Acknowledge ID=%v Tuner=%v Type=%v", e, t, powParams.PowerOverloadChangeType)
	}
	if err := a.Update(d.Dev, t, api.Update_Ctrl_OverloadMsgAck, api.Update_Ext1_None); err != nil {
		return err
	}

	if powParams.PowerOverloadChangeType != api.Overload_Detected {
		return nil
	}

	devParams, rxParams, err := GetRelatedParams(d, a, e, t, p)
	if err != nil {
		return err
	}

	for _, rx := range rxParams {
		currState := rx.TunerParams.Gain.LNAstate
		maxState := session.GetMaxLNAState(d, devParams, rx)
		if currState >= maxState {
			continue
		}
		if lg != nil {
			lg.Printf("Adjust LNA Tuner=%v CurrState=%d NewState=%d", t, currState, currState+1)
		}
		rx.TunerParams.Gain.LNAstate++
	}
	if err := a.StoreDeviceParams(d.Dev, devParams); err != nil {
		return err
	}
	if err := a.Update(d.Dev, t, api.Update_Tuner_Gr, api.Update_Ext1_None); err != nil {
		return err
	}
	return nil
}

// HandlePowerOverloadChangeMsg is a wrapper around HandlePowerOverloadChange
// for easy use with EventChan.
func HandlePowerOverloadChangeMsg(d *api.DeviceT, a api.API, evt Msg, lg Logger, adjust bool) error {
	return HandlePowerOverloadChange(d, a, evt.EventID, evt.Tuner, &evt.Params, lg, adjust)
}

// LogEvent is a helper function for logging a basic message representing
// the provided event.
func LogEvent(e api.EventT, t api.TunerSelectT, params *api.EventParamsT, lg Logger) {
	if lg == nil {
		return
	}
	switch e {
	case api.GainChange:
		p := params.GainParams
		lg.Printf(
			"Event ID=%v Tuner=%v GRdB=%d LNAGRdB=%d SystemGain=%.02f\n",
			e, t, p.GRdB, p.LnaGRdB, p.CurrGain,
		)
	case api.PowerOverloadChange:
		p := params.PowerOverloadParams
		lg.Printf("Event ID=%v Tuner=%v Type=%v\n", e, t, p.PowerOverloadChangeType)
	case api.DeviceRemoved:
		lg.Printf("Event ID=%v Tuner=%v\n", e, t)
	case api.RspDuoModeChange:
		p := params.RspDuoModeParams
		lg.Printf("Event ID=%v Tuner=%v Type=%v\n", e, t, p.ModeChangeType)
	default:
		lg.Printf("Event ID=%v Tuner=%v\n", e, t)
	}
}

// LogMsg is a wrapper around LogEvent for use with EventChan.
func LogMsg(evt Msg, lg Logger) {
	LogEvent(evt.EventID, evt.Tuner, &evt.Params, lg)
}
