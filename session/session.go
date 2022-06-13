// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package session

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/msiner/sdrplay-go/api"
)

// ConfigFn is implemented by a function that can take a Session
// and perform some configuration or return a non-nil error if
// a problem with the configuration is detected.
type ConfigFn func(o *Session) error

// ControlFn is implemented by a function that is responsible for
// run-time control after Init() has been called. Using the provided
// DeviceT and API, the device can be reconfigured as necessary.
//
// The function should implement some form of loop, sleep, or wait
// and not return until the device is no longer required. When the
// function returns, Uninit() will be called, Close() will be called,
// and the session will end.
type ControlFn func(ctx context.Context, d *api.DeviceT, a api.API) error

// Session is a type for storing/configuring a single session of
// access to an RSP device. The members can be set directly or
// by calling NewSession with the desired options declared using the
// WithXYZ() functions that return a ConfigFn (e.g. WithSelector).
type Session struct {
	Impl        api.API
	Selector    DevSelectFn
	DebugEn     bool
	DevCfg      DevConfigFn
	StreamACbFn api.StreamCallbackT
	StreamBCbFn api.StreamCallbackT
	EventCbFn   api.EventCallbackT
	Control     ControlFn
}

// NewSession creates a new Session and calls each given ConfigFn with
// it as the argument and then returns the configured Session. It returns
// a non-nil error immediately if any of the ConfigFn functions returns
// a non-nil error. It will call the ConfigFn functions in the order
// they are provided as arguments.
func NewSession(fns ...ConfigFn) (*Session, error) {
	opts := &Session{}
	for _, fn := range fns {
		if err := fn(opts); err != nil {
			return nil, err
		}
	}
	return opts, nil
}

// WithImplementation creates a function that sets the Impl member to the
// specified implementation of api.API. This is not necessary, as Run()
// will call api.GetInstance() if Impl is nil. This is available for testing
// via dependency injection.
func WithImplementation(impl api.API) ConfigFn {
	return func(o *Session) error {
		if o.Impl != nil {
			return errors.New("implementation already set")
		}
		o.Impl = impl
		return nil
	}
}

// WithSelector creates a ConfigFn that configures the Session with
// a DevSelectFn that applies all of the provided DevFilterFn functions
// to filter out or reorder available devices. The DevSelectFn will
// then select, if any are available, the first device in the filtered
// list of devices.
func WithSelector(filters ...DevFilterFn) ConfigFn {
	return func(o *Session) error {
		if o.Selector != nil {
			return errors.New("select function already set")
		}
		o.Selector = func(devs []*api.DeviceT) *api.DeviceT {
			for _, filt := range filters {
				devs = filt(devs)
			}
			if len(devs) == 0 {
				return nil
			}
			return devs[0]
		}
		return nil
	}
}

// WithDebug creates a ConfigFn that configures the Session to
// have debug output enabled or disabled, as specified.
func WithDebug(en bool) ConfigFn {
	return func(o *Session) error {
		o.DebugEn = en
		return nil
	}
}

// WithDeviceConfig creates a ConfigFn that configures the Session
// with a single DevConfigFn that applies all of the provided DevConfigFn
// functions in the order they are provided.
func WithDeviceConfig(fns ...DevConfigFn) ConfigFn {
	return func(o *Session) error {
		if o.DevCfg != nil {
			return errors.New("device config function already set")
		}
		o.DevCfg = func(d *api.DeviceT, p *api.DeviceParamsT) error {
			for _, fn := range fns {
				if err := fn(d, p); err != nil {
					return err
				}
			}
			return nil
		}
		return nil
	}
}

// WithStreamACallback creates a ConfigFn that configures the
// Session to use the provided StreamCallbackT function as
// the stream callback for stream A.
func WithStreamACallback(fn api.StreamCallbackT) ConfigFn {
	return func(o *Session) error {
		if o.StreamACbFn != nil {
			return errors.New("channel A callback function already set")
		}
		o.StreamACbFn = fn
		return nil
	}
}

// WithStreamBCallback creates a ConfigFn that configures the
// Session to use the provided StreamCallbackT function as
// the stream callback for stream B.
func WithStreamBCallback(fn api.StreamCallbackT) ConfigFn {
	return func(o *Session) error {
		if o.StreamBCbFn != nil {
			return errors.New("channel B callback function already set")
		}
		o.StreamBCbFn = fn
		return nil
	}
}

// WithEventCallback creates a ConfigFn that configures the
// Session to use the provided EventCallbackT function.
func WithEventCallback(fn api.EventCallbackT) ConfigFn {
	return func(o *Session) error {
		if o.EventCbFn != nil {
			return errors.New("event callback function already set")
		}
		o.EventCbFn = fn
		return nil
	}
}

// WithControlLoop configures the provided function as the control
// loop. This function will be called after Init(). When the function
// returns, Uninit() will be called and Run() will exit. A control loop
// function is not necessary. If a control loop function is not
// provided, Run will wait on the Context until it is canceled.
func WithControlLoop(fn ControlFn) ConfigFn {
	return func(o *Session) error {
		if o.Control != nil {
			return errors.New("control loop function already set")
		}
		o.Control = fn
		return nil
	}
}

// Run runs the configured Session. The provided Context is passed to the
// control loop function if one is provided. If no control loop has been
// provided, Run will wait on the ctx.Done() channel. Therfore, this
// function will block until an error is encountered, the control loop
// exits, and/or the Context is canceled.
func (s *Session) Run(ctx context.Context) error {
	impl := s.Impl
	if impl == nil {
		impl = api.GetAPI()
	}

	err := impl.Open()
	if err != nil {
		switch err.(type) {
		case api.ErrT:
			return fmt.Errorf("failed to open API: %v", impl.GetLastError(nil))
		default:
			return fmt.Errorf("failed to open API: %v", err)
		}
	}
	defer impl.Close()

	// Encapsulate device selection in a separate function so we can
	// do LockDeviceApi and then defer UnlockDeviceApi to cleanup
	// after device selection regardless of success or failure.
	selectDevice := func() (*api.DeviceT, error) {
		if err := impl.LockDeviceApi(); err != nil {
			return nil, fmt.Errorf("failed to lock API: %v", impl.GetLastError(nil))
		}
		defer func() {
			if err := impl.UnlockDeviceApi(); err != nil {
				fmt.Fprintf(os.Stderr, "UnlockDeviceApi failed: %v", err)
			}
		}()

		devs, err := impl.GetDevices()
		if err != nil {
			return nil, fmt.Errorf("failed to get device list: %v", impl.GetLastError(nil))
		}

		if len(devs) == 0 {
			return nil, errors.New("no RSP devices found")
		}

		res := devs[0]
		if s.Selector != nil {
			res := s.Selector(devs)
			if res == nil {
				var parts []string
				for _, dev := range devs {
					switch dev.HWVer {
					case api.RSPduo_ID:
						parts = append(parts, fmt.Sprintf("(%v,%v,%v,%v,%v)", dev.HWVer, dev.SerNo, dev.Tuner, dev.RspDuoMode, dev.RspDuoSampleFreq))
					default:
						parts = append(parts, fmt.Sprintf("(%v,%v)", dev.HWVer, dev.SerNo))
					}
				}
				return nil, fmt.Errorf("no matching devices selected from: %v", parts)
			}
		}

		if err := impl.SelectDevice(res); err != nil {
			return nil, fmt.Errorf("device selection failed: %v", impl.GetLastError(nil))
		}

		return res, nil
	}

	dev, err := selectDevice()
	if err != nil {
		return err
	}
	defer func() {
		if err := impl.ReleaseDevice(dev); err != nil {
			fmt.Fprintf(os.Stderr, "ReleaseDevice failed: %v", err)
		}
	}()

	if s.DebugEn {
		if err := impl.DebugEnable(dev.Dev, api.DbgLvl_Message); err != nil {
			return fmt.Errorf("debug enable failed: %v", impl.GetLastError(dev))
		}
	}

	params, err := impl.LoadDeviceParams(dev.Dev)
	if err != nil {
		return fmt.Errorf("failed to load device params: %v", impl.GetLastError(dev))
	}

	if s.DevCfg != nil {
		if err := s.DevCfg(dev, params); err != nil {
			return err
		}
	}

	if err := impl.StoreDeviceParams(dev.Dev, params); err != nil {
		return fmt.Errorf("failed to store device params: %v", impl.GetLastError(dev))
	}

	cbFuncs := api.CallbackFnsT{
		StreamACbFn: s.StreamACbFn,
		StreamBCbFn: s.StreamBCbFn,
		EventCbFn:   s.EventCbFn,
	}
	if err := impl.Init(dev.Dev, cbFuncs); err != nil {
		return fmt.Errorf("init failed: %v", impl.GetLastError(dev))
	}
	defer func() {
		if err := impl.Uninit(dev.Dev); err != nil {
			fmt.Fprintf(os.Stderr, "Uninit failed: %v", err)
		}
	}()

	switch s.Control {
	case nil:
		// No control loop provided, just wait on the context.
		<-ctx.Done()
		return ctx.Err()
	default:
		return s.Control(ctx, dev, impl)
	}
}

// Run is a simplified wrapper around calling NewSession, checking for
// and error, and then calling Session.Run.
func Run(ctx context.Context, fns ...ConfigFn) error {
	s, err := NewSession(fns...)
	if err != nil {
		return err
	}
	return s.Run(ctx)
}
