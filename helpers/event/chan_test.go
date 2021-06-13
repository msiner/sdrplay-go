// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package event

import (
	"strings"
	"testing"
	"time"

	"github.com/msiner/sdrplay-go/api"
)

func TestEventChanSync(t *testing.T) {
	ec := NewChan(0)
	defer ec.Close()

	ec.Callback(api.DeviceRemoved, api.Tuner_A, nil)

	select {
	case msg := <-ec.C:
		t.Fatalf("unexpected message; %v", msg)
	default:
		// Good, message should have been dropped
	}

	ec.Callback(api.DeviceRemoved, api.Tuner_A, nil)

	select {
	case msg := <-ec.C:
		t.Fatalf("unexpected message; %v", msg)
	default:
		// Good, message should have been dropped
	}

	wait := make(chan struct{})
	params := &api.EventParamsT{}
	params.GainParams.CurrGain = 42
	go func() {
		<-wait
		// Yield to ensure the reader gets to select.
		time.Sleep(1 * time.Millisecond)
		ec.Callback(api.PowerOverloadChange, api.Tuner_A, params)
	}()
	tmr := time.NewTimer(1 * time.Second)
	defer tmr.Stop()

	wait <- struct{}{}
	select {
	case msg := <-ec.C:
		// MsgNum 0 and 1 were dropped earlier in the test.
		if msg.MsgNum != 2 {
			t.Errorf("wrong MsgNum; got %d, want 2", msg.MsgNum)
		}
		// Ensure it is not one of the earlier DeviceRemoved messages
		if msg.EventID != api.PowerOverloadChange {
			t.Errorf("wrong EventID; got %v, want %v", msg.EventID, api.PowerOverloadChange)
		}
		if msg.Params.GainParams.CurrGain != params.GainParams.CurrGain {
			t.Fatalf("wrong params; got %v, want %v", msg.Params.GainParams, msg.Params.GainParams)
		}
	case <-tmr.C:
		t.Fatal("synchronous event read failed")
	}

	if err := ec.Close(); err != nil {
		t.Errorf("unexpected error on close; %v", err)
	}

	err := ec.Close()
	if err == nil {
		t.Fatalf("no error on double close")
	}
	if !strings.Contains(err.Error(), "already closed") {
		t.Errorf("wrong error message; got '%s', want 'already closed'", err.Error())
	}

	select {
	case msg, ok := <-ec.C:
		t.Errorf("got unexpected message; %v", msg)
		if !ok {
			t.Error("chan unexpectedly closed")
		}
	default:
		// Expected on empty, but open chan
	}

	ec.Callback(api.DeviceRemoved, api.Tuner_A, nil)

	select {
	case _, ok := <-ec.C:
		if ok {
			t.Error("chan not closed")
		}
	default:
		t.Errorf("default select case when chan is supped to be closed")
	}
}

func TestEventChanAsync(t *testing.T) {
	ec := NewChan(2)
	defer ec.Close()

	params := &api.EventParamsT{}
	params.GainParams.CurrGain = 42
	ec.Callback(api.DeviceRemoved, api.Tuner_A, params)
	ec.Callback(api.RspDuoModeChange, api.Tuner_Both, params)
	ec.Callback(api.PowerOverloadChange, api.Tuner_B, params)

	select {
	case msg := <-ec.C:
		if msg.MsgNum != 0 {
			t.Errorf("wrong MsgNum; got %d, want 0", msg.MsgNum)
		}
		if msg.EventID != api.DeviceRemoved {
			t.Errorf("wrong EventID; got %v, want %v", msg.EventID, api.DeviceRemoved)
		}
		if msg.Params.GainParams.CurrGain != params.GainParams.CurrGain {
			t.Fatalf("wrong params; got %v, want %v", msg.Params.GainParams, msg.Params.GainParams)
		}
	default:
		t.Fatalf("no async message")
	}

	select {
	case msg := <-ec.C:
		if msg.MsgNum != 1 {
			t.Errorf("wrong MsgNum; got %d, want 1", msg.MsgNum)
		}
		if msg.EventID != api.RspDuoModeChange {
			t.Errorf("wrong EventID; got %v, want %v", msg.EventID, api.RspDuoModeChange)
		}
		if msg.Params.GainParams.CurrGain != params.GainParams.CurrGain {
			t.Fatalf("wrong params; got %v, want %v", msg.Params.GainParams, msg.Params.GainParams)
		}
	default:
		t.Fatalf("no async message")
	}

	// Channel should be drained by this point.
	select {
	case msg := <-ec.C:
		t.Fatalf("unexpected message; %v", msg)
	default:
		// Good, message should have been dropped
	}

	ec.Callback(api.PowerOverloadChange, api.Tuner_B, nil)
	ec.Callback(api.RspDuoModeChange, api.Tuner_Both, nil)
	ec.Callback(api.GainChange, api.Tuner_A, nil)

	select {
	case msg := <-ec.C:
		// MsgNum 2 should have been dropped
		if msg.MsgNum != 3 {
			t.Errorf("wrong MsgNum; got %d, want 3", msg.MsgNum)
		}
		if msg.EventID != api.PowerOverloadChange {
			t.Errorf("wrong EventID; got %v, want %v", msg.EventID, api.PowerOverloadChange)
		}
	default:
		t.Fatalf("no async message")
	}

	select {
	case msg := <-ec.C:
		if msg.MsgNum != 4 {
			t.Errorf("wrong MsgNum; got %d, want 4", msg.MsgNum)
		}
		if msg.EventID != api.RspDuoModeChange {
			t.Errorf("wrong EventID; got %v, want %v", msg.EventID, api.RspDuoModeChange)
		}
	default:
		t.Fatalf("no async message")
	}

	// Channel should be drained by this point.
	select {
	case msg := <-ec.C:
		t.Fatalf("unexpected message; %v", msg)
	default:
		// Good, message should have been dropped
	}

	if err := ec.Close(); err != nil {
		t.Errorf("unexpected error on close; %v", err)
	}

	err := ec.Close()
	if err == nil {
		t.Fatalf("no error on double close")
	}
	if !strings.Contains(err.Error(), "already closed") {
		t.Errorf("wrong error message; got '%s', want 'already closed'", err.Error())
	}
}
