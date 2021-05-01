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

type Handler func(d *api.DeviceT, a api.API, e api.EventT, t api.TunerSelectT, p *api.EventParamsT) error

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

func GetRelatedParamsMsg(d *api.DeviceT, a api.API, evt EventMsg) (*api.DeviceParamsT, []*api.RxChannelParamsT, error) {
	return GetRelatedParams(d, a, evt.EventId, evt.Tuner, &evt.Params)
}

func HandlePowerOverloadChange(d *api.DeviceT, a api.API, e api.EventT, t api.TunerSelectT, p *api.EventParamsT, lg Logger, adjust bool) error {
	if e != api.PowerOverloadChange {
		return nil
	}

	powParams := p.PowerOverloadParams()

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

func HandlePowerOverloadChangeMsg(d *api.DeviceT, a api.API, evt EventMsg, lg Logger, adjust bool) error {
	return HandlePowerOverloadChange(d, a, evt.EventId, evt.Tuner, &evt.Params, lg, adjust)
}

func LogEvent(e api.EventT, t api.TunerSelectT, p *api.EventParamsT, lg Logger) {
	if lg == nil {
		return
	}
	switch e {
	case api.GainChange:
		gainParams := p.GainParams()
		lg.Printf(
			"Event ID=%v Tuner=%v GRdB=%d LNAGRdB=%d SystemGain=%.02f\n",
			e, t, gainParams.GRdB, gainParams.LnaGRdB, gainParams.CurrGain,
		)
	case api.PowerOverloadChange:
		powParams := p.PowerOverloadParams()
		lg.Printf("Event ID=%v Tuner=%v Type=%v\n", e, t, powParams.PowerOverloadChangeType)
	case api.DeviceRemoved:
		lg.Printf("Event ID=%v Tuner=%v\n", e, t)
	case api.RspDuoModeChange:
		duoParams := p.RspDuoModeParams()
		lg.Printf("Event ID=%v Tuner=%v Type=%v\n", e, t, duoParams.ModeChangeType)
	default:
		lg.Printf("Event ID=%v Tuner=%v\n", e, t)
	}
}

func LogEventMsg(evt EventMsg, lg Logger) {
	LogEvent(evt.EventId, evt.Tuner, &evt.Params, lg)
}
