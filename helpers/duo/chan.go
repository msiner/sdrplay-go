// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package duo

import (
	"errors"
	"fmt"
)

// SynchroMsg is a type for storing or transferring the parameters
// of a SynchroCbFn callback. Specifically, sending via a chan.
type SynchroMsg struct {
	Xia    []int16
	Xqa    []int16
	Xib    []int16
	Xqb    []int16
	Reset  bool
	MsgNum uint64
}

// SynchroChan is a type that provides a stream callback handler that
// sends a message for each callback call that allows users to handle
// sample data asynchronously to the C callback thread. This frees the
// callback thread quickly to minimize dropped samples inside the API.
//
// SynchroChan is designed to be used with the Synchro type in this
// package. For asynchronous handling of individual streams, see the
// helpers/callback.StreamChan type instead.
type SynchroChan struct {
	C      <-chan SynchroMsg
	c      chan<- SynchroMsg
	done   chan struct{}
	msgNum uint64
}

// NewSynchroChan creates a SynchroChan. The depth parameters specifies
// the depth of the channel. Since the callback happens on a thread from
// the C library, it should not and will not block on channel send if the
// channel is full or there is not a receiver waiting. It will simply drop
// the message. A depth of 0 will drop any message if the receiver is not
// ready when it is available. Any depth greater than zero will buffer that
// number of messages and allow the receiver to receive them asynchronously.
func NewSynchroChan(depth uint) *SynchroChan {
	cbChan := make(chan SynchroMsg, depth)
	return &SynchroChan{
		C:    cbChan,
		c:    cbChan,
		done: make(chan struct{}, 1),
	}
}

// Close implements io.Closer. It will stop any more messages from being sent
// on the C chan. However, the C chan will not be closed until the next
// call to Callback.
func (s *SynchroChan) Close() error {
	select {
	case <-s.done:
		return errors.New("already closed")
	default:
		close(s.done)
		return nil
	}
}

// Callback is a bound implementation of api.SynchroCbFn. It can be
// passed to the API as the stream callback or used directly. Valid calls
// to call back will generate a message on the C chan.
func (s *SynchroChan) Callback(xia, xqa, xib, xqb []int16, reset bool) {
	select {
	case <-s.done:
		if s.c != nil {
			close(s.c)
			s.c = nil
		}
		return
	default:
	}

	numSamples := len(xia)
	if numSamples != len(xqa) || numSamples != len(xib) || numSamples != len(xqb) {
		panic(fmt.Sprintf("mismatched buffer lengths; %d, %d, %d, %d", len(xia), len(xqa), len(xib), len(xqb)))
	}

	// Create one big buffer for a single allocation and
	// to optimize locality.
	buf := make([]int16, numSamples*4)
	pay := SynchroMsg{
		Reset:  reset,
		MsgNum: s.msgNum,
		Xia:    buf[:numSamples],
		Xqa:    buf[numSamples : numSamples*2],
		Xib:    buf[numSamples*2 : numSamples*3],
		Xqb:    buf[numSamples*3:],
	}

	s.msgNum++

	copy(pay.Xia, xia)
	copy(pay.Xqa, xqa)
	copy(pay.Xib, xib)
	copy(pay.Xqb, xqb)

	select {
	case s.c <- pay:
	default:
		// Receiver not ready or channel full; drop the payload.
	}
}
