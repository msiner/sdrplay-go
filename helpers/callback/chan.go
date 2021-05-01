// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package callback

import (
	"errors"

	"github.com/msiner/sdrplay-go/api"
)

// StreamMsg is a type for storing or transferring the parameters
// of a StreamCallbackT. Specifically, sending via a chan.
//
// A StreamCallbackT called from the C API does not own the sample
// buffers. If a callback stores the parameters in a StreamMsg,
// it must first copy the xi and xq sample buffers. Note that, since
// Xi and Xq are slices, the StreamMsg type itself does not have memory
// allocated for the samples.
type StreamMsg struct {
	Xi     []int16
	Xq     []int16
	Params api.StreamCbParamsT
	Reset  bool
	MsgNum uint32
}

// StreamChan is a type that provides a stream callback handler that
// sends a message for each callback call that allows users to handle
// sample data asynchronously to the C callback thread. This frees the
// callback thread quickly to minimize dropped samples inside the API.
type StreamChan struct {
	C      <-chan StreamMsg
	c      chan<- StreamMsg
	done   chan struct{}
	msgNum uint32
}

// NewStreamChan creates a StreamChan. The depth parameters specifies
// the depth of the channel. Since the callback happens on a thread from
// the C library, it should not and will not block on channel send if the
// channel is full or there is not a receiver waiting. It will simply drop
// the message. A depth of 0 will drop any message if the receiver is not
// ready when it is available. Any depth greater than zero will buffer that
// number of messages and allow the receiver to receive them asynchronously.
func NewStreamChan(depth uint) *StreamChan {
	cbChan := make(chan StreamMsg, depth)
	return &StreamChan{
		C:    cbChan,
		c:    cbChan,
		done: make(chan struct{}, 1),
	}
}

// Close implements io.Closer. It will stop any more messages from being sent
// on the C chan. However, the C chan will not be closed until the next
// call to Callback.
func (s *StreamChan) Close() error {
	select {
	case <-s.done:
		return errors.New("already closed")
	default:
		close(s.done)
		return nil
	}
}

// Callback is a bound implementation of api.StreamCallbackT. It can be
// passed to the API as the stream callback or used directly. Valid calls
// to call back will generate a message on the C chan.
func (s *StreamChan) Callback(xi, xq []int16, params *api.StreamCbParamsT, reset bool) {
	select {
	case <-s.done:
		if s.c != nil {
			close(s.c)
			s.c = nil
		}
		return
	default:
	}

	// Create one big buffer for a single allocation and
	// to optimize locality.
	buf := make([]int16, params.NumSamples*2)
	pay := StreamMsg{
		Reset:  reset,
		MsgNum: s.msgNum,
		Xi:     buf[:int(params.NumSamples)],
		Xq:     buf[int(params.NumSamples):],
	}

	s.msgNum++

	copy(pay.Xi, xi)
	copy(pay.Xq, xq)

	if params != nil {
		pay.Params = *params
	}

	select {
	case s.c <- pay:
	default:
	}
}
