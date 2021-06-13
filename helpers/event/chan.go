// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package event

import (
	"errors"

	"github.com/msiner/sdrplay-go/api"
)

// Msg is a type for storing or transferring the parameters
// of a EventCallbackT. Specifically, sending via a chan.
type Msg struct {
	EventID api.EventT
	Tuner   api.TunerSelectT
	Params  api.EventParamsT
	MsgNum  uint32
}

// Chan is a type that provides a stream callback handler that
// sends a message for each callback call that allows users to handle
// events asynchronously to the C callback thread.
type Chan struct {
	C      <-chan Msg
	c      chan<- Msg
	done   chan struct{}
	msgNum uint32
}

// NewChan creates an Chan. The depth parameters specifies
// the depth of the channel. Since the callback happens on a thread from
// the C library, it should not and will not block on channel send if the
// channel is full or there is not a receiver waiting. It will simply drop
// the message. A depth of 0 will drop any message if the receiver is not
// ready when it is available. Any depth greater than zero will buffer that
// number of messages and allow the receiver to receive them asynchronously.
func NewChan(depth uint) *Chan {
	cbChan := make(chan Msg, depth)
	return &Chan{
		C:    cbChan,
		c:    cbChan,
		done: make(chan struct{}, 1),
	}
}

// Close implements io.Closer. It will stop any more messages from being sent
// on the C chan. However, the C chan will not be closed until the next
// call to Callback.
func (e *Chan) Close() error {
	select {
	case <-e.done:
		return errors.New("already closed")
	default:
		close(e.done)
		return nil
	}
}

// Callback is a bound implementation of api.EventCallbackT. It can be
// passed to the API as the event callback or used directly. Valid calls
// to call back will generate a message on the C chan.
func (e *Chan) Callback(eventID api.EventT, tuner api.TunerSelectT, params *api.EventParamsT) {
	select {
	case <-e.done:
		if e.c != nil {
			close(e.c)
			e.c = nil
		}
		return
	default:
	}

	pay := Msg{
		EventID: eventID,
		Tuner:   tuner,
		MsgNum:  e.msgNum,
	}

	e.msgNum++

	if params != nil {
		pay.Params = *params
	}

	select {
	case e.c <- pay:
	default:
	}
}
