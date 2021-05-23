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

type SynchroCbFn func(xia, xqa, xib, xqb []int16, reset bool)
type SynchroEventCbFn func(evt SynchroEvent, msg string)

// Synchro is a helper type that provides the ability to synchronize
// samples from both streams into a single callback.
type Synchro struct {
	cbFrames    int
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

func (f *Synchro) Reset() {
	f.doEvent(SynchroReset, "synchro reset")
	f.numSamplesA = 0
	f.numSamplesB = 0
	f.rxIdx = 0
	f.txIdx = 0
	f.reset = true
	f.sync = false
}

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

func (f *Synchro) doEvent(evt SynchroEvent, msg string) {
	if f.evtCb != nil {
		f.evtCb(evt, msg)
	}
}

func (f *Synchro) UpdateStreamA(xi, xq []int16, params *api.StreamCbParamsT, reset bool) {
	switch {
	case reset:
		f.Reset()
	case len(xi) != len(xq):
		if f.sync {
			f.doEvent(SynchroOutOfSync, fmt.Sprintf("len(xia)=%d len(xqa)=%d", len(xi), len(xq)))
		}
		return
	case f.numSamplesA != 0 || f.numSamplesB == 0:
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

func (f *Synchro) UpdateStreamB(xi, xq []int16, params *api.StreamCbParamsT, reset bool) {
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
	mod := len(f.xia)

	switch idx+len(xi) < mod {
	case true:
		copy(f.xib[idx:], xi)
		copy(f.xqb[idx:], xq)
		for range xi {
			idx++
			if (idx % f.cbScalars) == 0 {
				f.doCallback()
			}
		}
	default:
		copy(f.xib[idx:], xi)
		n := copy(f.xqb[idx:], xq)
		for idx < mod {
			idx++
			if (idx % f.cbScalars) == 0 {
				f.doCallback()
			}
		}
		copy(f.xib, xi[n:])
		n = copy(f.xqb, xq[n:])
		idx = 0
		for idx < n {
			idx++
			if (idx % f.cbScalars) == 0 {
				f.doCallback()
			}
		}
	}

	f.rxIdx = idx
	// clear to indicate to A that B has been handled
	f.numSamplesA = 0
}
