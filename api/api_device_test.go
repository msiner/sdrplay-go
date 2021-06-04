// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//+build devicetest
package api

import (
	"log"
	"testing"
	"time"
)

func TestWithDevice(t *testing.T) {
	impl := GetAPI()

	err := impl.Open()
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		// This is overkill, but makes rspdetect useful for detecting
		// this type of error that might otherwise go unnoticed.
		if err := impl.Close(); err != nil {
			t.Errorf("error on close: %v", err)
		}
	}()

	selectDevice := func() *DeviceT {
		if err := impl.LockDeviceApi(); err != nil {
			t.Fatalf("failed to lock API: %v", impl.GetLastError(nil))
		}
		defer impl.UnlockDeviceApi()

		ver, err := impl.ApiVersion()
		if err != nil {
			log.Fatal(err)
		}
		t.Logf("API Version %v\n", ver)

		devs, err := impl.GetDevices()
		if err != nil {
			t.Fatalf("error on GetDevices; %v", err)
		}

		if len(devs) == 0 {
			t.Fatal("no devices found")
		}

		dev := devs[0]
		switch dev.HWVer {
		case RSPduo_ID:
			t.Logf("%v,%v,%v,%v,%v\n", dev.HWVer, dev.SerNo, dev.Tuner, dev.RspDuoMode, dev.RspDuoSampleFreq)
			if dev.RspDuoMode&RspDuoMode_Dual_Tuner == 0 {
				t.Fatalf("test requires exclusive access to RSPduo")
			}
			dev.RspDuoMode = RspDuoMode_Primary
			dev.RspDuoSampleFreq = 6e6
			dev.Tuner = Tuner_A
		default:
			t.Logf("%v,%v\n", dev.HWVer, dev.SerNo)
		}

		if err := impl.SelectDevice(dev); err != nil {
			t.Fatalf("error on SelectDevice; %v", err)
		}
		return dev
	}

	dev := selectDevice()
	defer impl.ReleaseDevice(dev)

	gotData := make(chan bool, 1)
	cbFuncs := CallbackFnsT{
		StreamACbFn: func(xi, xq []int16, params *StreamCbParamsT, reset bool) {
			select {
			case gotData <- true:
			default:
				// Don't block the callback thread
			}
		},
		EventCbFn: func(eventId EventT, tuner TunerSelectT, params *EventParamsT) {
			t.Log(eventId, tuner)
		},
	}
	if err := impl.Init(dev.Dev, cbFuncs); err != nil {
		t.Fatalf("init failed: %v", impl.GetLastError(dev))
	}
	defer impl.Uninit(dev.Dev)

	tmr := time.NewTimer(2 * time.Second)
	<-tmr.C
	select {
	case <-gotData:
	default:
		t.Error("did not receive stream callback")
	}
}
