// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package duo

//go:generate go run golang.org/x/tools/cmd/stringer -type SynchroEvent -output synchro_string.go

import (
	"fmt"

	"github.com/msiner/sdrplay-go/api"
)

type SynchroEvent int

const (
	SynchroMessage SynchroEvent = iota
	SynchroReset
	SynchroSync
	SynchroOutOfSync
)

// SynchroCbFn is the function type for user-provided stream
// callback functions provided to Synchro. It provides samples for
// all four channels in separate slices.
type SynchroCbFn func(xia, xqa, xib, xqb []int16, reset bool)

// SynchroEventCbFn is the function type for user-provided event
// callback functions provided to Synchro.
type SynchroEventCbFn func(evt SynchroEvent, msg string)

// Synchro is a helper type that provides the ability to synchronize
// samples from both streams into a single callback. It allows the
// user to provide a single callback that handles time-aligned samples
// from all four channels.
//
// Synchronization Assumptions:
// 1. Streams A and B have the same effective sample rate.
// 2. Stream A and B callbacks must be called in a consistently
//    alternating serial order.
// 3. Stream A and B callbacks receive the same number of
//    samples in successive callbacks.
// 4. Samples provided to a stream B callback match the same
//    time span as the samples provided to the last stream A
//    callback.
type Synchro struct {
	cbScalars   int
	cb          SynchroCbFn
	evtCb       SynchroEventCbFn
	xia         []int16
	xqa         []int16
	xib         []int16
	xqb         []int16
	numSamplesA int
	numSamplesB int
	rxIdx       int
	txIdx       int
	reset       bool
	sync        bool
}

// NewSynchro creates a new Synchro.
//
// cbSamples specifies the exact number of scalars per channel to provide to
// the user-provided cb callback.
//
// cb is the callback function that is called when cbSamples number of
// samples are available for all channels.
//
// evtCb is the callback function that is called when a SynchroEvent
// occurs. It may be nil to ignore events.
func NewSynchro(cbSamples int, cb SynchroCbFn, evtCb SynchroEventCbFn) *Synchro {
	bufSize := 10 * cbSamples
	for bufSize < 10*1024 {
		bufSize *= 2
	}
	buf := make([]int16, 4*bufSize)
	return &Synchro{
		cbScalars: cbSamples,
		cb:        cb,
		evtCb:     evtCb,
		xia:       buf[:bufSize],
		xqa:       buf[bufSize : 2*bufSize],
		xib:       buf[2*bufSize : 3*bufSize],
		xqb:       buf[3*bufSize:],
		reset:     true,
	}
}

// Reset resets the state of Synchro. It should only be called from
// either the stream or event callback function.
func (f *Synchro) Reset() {
	f.doEvent(SynchroReset, "synchro reset")
	f.numSamplesA = 0
	f.numSamplesB = 0
	f.rxIdx = 0
	f.txIdx = 0
	f.reset = true
	f.sync = false
}

// doCallback creates sub-slices of the internal buffers and
// uses them to call the stream callback function. It advances
// the internal transfer index.
func (f *Synchro) doCallback() {
	end := f.txIdx + f.cbScalars
	xia := f.xia[f.txIdx:end]
	xqa := f.xqa[f.txIdx:end]
	xib := f.xib[f.txIdx:end]
	xqb := f.xqb[f.txIdx:end]
	f.txIdx = (f.txIdx + f.cbScalars) % len(f.xia)
	reset := f.reset
	f.reset = false
	if f.cb == nil {
		return
	}
	f.cb(xia, xqa, xib, xqb, reset)
}

// doEvent calls the user event callback, if it is not nil,
// with the provided arguments. It mostly exists to centralize
// the nil check.
func (f *Synchro) doEvent(evt SynchroEvent, msg string) {
	if f.evtCb != nil {
		f.evtCb(evt, msg)
	}
}

// UpdateStreamA is an implementation of StreamCallbackT bound to a
// Synchro instance. It can be provided directly as the stream callback
// or called inside a user implementation of StreamCallbackT. It updates
// the internal stream A buffers with the provided samples after verifying
// synchronization with stream B. If reset is true, the internal state
// of the Synchro is reset and an event is issued.
//
// The user-provided stream callback is not called from UpdateStreamA.
// It is only called from UpdateStreamB.
func (f *Synchro) StreamACallback(xi, xq []int16, params *api.StreamCbParamsT, reset bool) {
	if reset {
		f.Reset()
	}
	switch {
	case len(xi) != len(xq):
		if f.sync {
			f.doEvent(SynchroOutOfSync, fmt.Sprintf("len(xia)=%d len(xqa)=%d", len(xi), len(xq)))
		}
		return
	case f.numSamplesA != 0:
		if f.sync {
			f.doEvent(SynchroOutOfSync, "stream B has not been handled")
		}
		return
	}
	f.reset = f.reset || reset

	f.numSamplesA = len(xi)
	idx := f.rxIdx
	mod := len(f.xia)

	switch idx+len(xi) < mod {
	case true:
		copy(f.xia[idx:], xi)
		copy(f.xqa[idx:], xq)
	default:
		copy(f.xia[idx:], xi)
		n := copy(f.xqa[idx:], xq)
		copy(f.xia, xi[n:])
		copy(f.xqa, xq[n:])
	}
}

// UpdateStreamB is an implementation of StreamCallbackT bound to a
// Synchro instance. It can be provided directly as the stream callback
// or called inside a user implementation of StreamCallbackT. It updates
// the internal stream B buffers with the provided samples after verifying
// synchronization with stream A. If reset is true, an internal flag is
// set, but a Reset event will not be issued until the next call to
// UpdateStreamA.
//
// The user-provided stream callback is only called from UpdateStreamB.
func (f *Synchro) StreamBCallback(xi, xq []int16, params *api.StreamCbParamsT, reset bool) {
	switch {
	case len(xi) != len(xq):
		if f.sync {
			f.doEvent(SynchroOutOfSync, fmt.Sprintf("len(xib)=%d len(xqb)=%d", len(xi), len(xq)))
		}
		return
	case f.numSamplesA == 0:
		if f.sync {
			f.sync = false
			f.doEvent(SynchroOutOfSync, "stream A has not been handled")
		}
		return
	case f.numSamplesA != len(xi):
		if f.sync {
			f.sync = false
			f.doEvent(SynchroOutOfSync, fmt.Sprintf("numSamplesA=%d numSamplesB=%d", f.numSamplesA, len(xi)))
		}
		return
	}
	f.reset = f.reset || reset

	if !f.sync {
		f.doEvent(SynchroSync, fmt.Sprintf("synchronized; numSamples=%d", len(xi)))
	}
	f.sync = true

	f.numSamplesB = len(xi)
	idx := f.rxIdx
	rem := len(xi)
	for rem > 0 {
		lastCb := idx - (idx % f.cbScalars)
		nextCb := lastCb + f.cbScalars
		toNext := nextCb - idx
		bufIdx := len(xi) - rem
		switch toNext > rem {
		case true:
			copy(f.xib[idx:], xi[bufIdx:])
			copy(f.xqb[idx:], xq[bufIdx:])
			idx += rem
			rem = 0
		default:
			copy(f.xib[idx:nextCb], xi[bufIdx:])
			copy(f.xqb[idx:nextCb], xq[bufIdx:])
			idx = nextCb
			rem -= toNext
			f.doCallback()
		}
		idx %= len(f.xia)
	}

	f.rxIdx = idx
	// clear to indicate to A that B has been handled
	f.numSamplesA = 0
}
