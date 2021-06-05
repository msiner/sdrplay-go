//+build devicetest

package session

import (
	"context"
	"testing"
	"time"

	"github.com/msiner/sdrplay-go/api"
)

func TestWithDevice(t *testing.T) {
	gotData := make(chan bool, 1)

	err := Run(
		context.Background(),
		WithSelector(
			WithDuoModeSingle(),
			WithDuoTunerA(),
		),
		WithDeviceConfig(
			WithTransferMode(api.ISOCH),
			WithSingleChannelConfig(
				WithTuneFreq(1.001e9),
				WithAGC(api.AGC_CTRL_EN, -30),
				WithZeroIF(4e6, 2),
			),
		),
		WithStreamACallback(func(xi, xq []int16, params *api.StreamCbParamsT, reset bool) {
			select {
			case gotData <- true:
			default:
				// Don't block the callback thread
			}
		}),
		WithEventCallback(func(eventId api.EventT, tuner api.TunerSelectT, params *api.EventParamsT) {
			t.Log(eventId, tuner)
		}),
		WithControlLoop(func(ctx context.Context, d *api.DeviceT, a api.API) error {
			tmr := time.NewTimer(2 * time.Second)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-tmr.C:
			}
			select {
			case <-gotData:
			default:
				t.Error("did not receive stream callback")
			}
			return nil
		}),
	)
	if err != nil {
		t.Errorf("error returned from Run; %v", err)
	}
}
