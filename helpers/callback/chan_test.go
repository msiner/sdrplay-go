// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package callback

import (
	"testing"

	"github.com/msiner/sdrplay-go/api"
)

func TestStreamChan(t *testing.T) {
	const numSamples = 1008
	xi := make([]int16, numSamples)
	xq := make([]int16, numSamples)
	params := api.StreamCbParamsT{
		NumSamples: numSamples,
	}

	sc := NewStreamChan(1)

	select {
	case <-sc.C:
		t.Fatal("unexpected message available on chan")
	default:
		// good
	}

	sc.Callback(xi, xq, &params, false)
	select {
	case msg, ok := <-sc.C:
		if !ok {
			t.Fatal("stream message channel not ok")
		}
		if len(msg.Xi) != numSamples {
			t.Fatalf("msg payload has wrong length; got %d, want %d", len(msg.Xi), numSamples)
		}
	default:
		t.Fatal("no message available on chan")
	}

	sc.Callback(xi, xq, &params, false)
	select {
	case _, ok := <-sc.C:
		if !ok {
			t.Fatal("stream message channel not ok")
		}
	default:
		t.Fatal("no message available on chan")
	}

	sc.Close()

	sc.Callback(xi, xq, &params, false)
	select {
	case _, ok := <-sc.C:
		if ok {
			t.Fatal("stream message channel ok after close")
		}
	default:
		t.Fatal("chan not closed")
	}
}

func BenchmarkStreamChan(b *testing.B) {
	const numSamples = 1008
	xi := make([]int16, numSamples)
	xq := make([]int16, numSamples)
	params := api.StreamCbParamsT{
		NumSamples: numSamples,
	}

	sc := NewStreamChan(uint(b.N + 1))

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		sc.Callback(xi, xq, &params, false)
	}
}
