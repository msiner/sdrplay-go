// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package duo

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

// TestSynchroBasic tests the basic operation of Synchro.
func TestSynchro(t *testing.T) {
	t.Parallel()

	const numSamples = 1000
	totalFrames := 0
	numCallbacks := 0
	numResets := 0
	var (
		lastEvent *SynchroEvent
		lastMsg   string
	)
	f := NewSynchro(
		numSamples,
		func(xia, xqa, xib, xqb []int16, reset bool) {
			numCallbacks++
			if reset {
				numResets++
			}
			totalFrames += len(xia)
			var want int16
			for i := range xia {
				if xia[i] != want {
					t.Fatalf("wrong value at %d in xia: got %d, want %d", i, xia[i], want)
				}
				want++
				if xqa[i] != want {
					t.Fatalf("wrong value at %d in xqa: got %d, want %d", i, xqa[i], want)
				}
				want++
				if xib[i] != want {
					t.Fatalf("wrong value at %d in xib: got %d, want %d", i, xib[i], want)
				}
				want++
				if xqb[i] != want {
					t.Fatalf("wrong value at %d in xqb: got %d, want %d", i, xqb[i], want)
				}
				want++
			}
		},
		func(evt SynchroEvent, msg string) {
			lastEvent = &evt
			lastMsg = msg
			fmt.Println(evt, msg)
		},
	)
	xia := make([]int16, numSamples)
	xqa := make([]int16, numSamples)
	xib := make([]int16, numSamples)
	xqb := make([]int16, numSamples)
	val := 0
	for i := 0; i < numSamples; i++ {
		xia[i] = int16(val)
		xqa[i] = int16(val + 1)
		xib[i] = int16(val + 2)
		xqb[i] = int16(val + 3)
		val += 4
	}
	f.StreamACallback(xia, xqa, nil, true)
	f.StreamBCallback(xib, xqb, nil, true)
	f.StreamACallback(xia, xqa, nil, false)
	f.StreamBCallback(xib, xqb, nil, false)
	if numResets != 1 {
		t.Errorf("wrong number of resets: got %d, want 1", numResets)
	}
	if numCallbacks != 2 {
		t.Errorf("wrong number of callbacks: got %d, want 2", numCallbacks)
	}
	if totalFrames != (numSamples * numCallbacks) {
		t.Errorf("wrong total number of frames: got %d, want %d", totalFrames, numSamples*numCallbacks)
	}

	lastEvent = nil
	f.StreamBCallback(xib, xqb, nil, false)
	if lastEvent == nil {
		t.Fatal("expected out of sync event: got none")
	}
	switch *lastEvent {
	case SynchroOutOfSync:
		want := "A has not been handled"
		if !strings.Contains(lastMsg, want) {
			t.Fatalf("wrong event message: got '%s', want '%s'", lastMsg, want)
		}
	default:
		t.Fatalf("wrong event type: got %v, want %v", lastEvent, SynchroOutOfSync)
	}

	lastEvent = nil
	f.StreamBCallback(xib, xqb, nil, false)
	if lastEvent != nil {
		t.Fatalf("new event when already out of sync: got %v, want none", lastEvent)
	}

	lastEvent = nil
	f.StreamACallback(xia, xqa, nil, false)
	f.StreamBCallback(xib, xqb, nil, false)
	if lastEvent == nil {
		t.Fatal("expected sync event: got none")
	}
	switch *lastEvent {
	case SynchroSync:
		// good
	default:
		t.Fatalf("wrong event type: got %v, want %v", lastEvent, SynchroSync)
	}

	lastEvent = nil
	f.StreamACallback(xia, xqa, nil, false)
	f.StreamBCallback(xib[:10], xqb[:10], nil, false)
	if lastEvent == nil {
		t.Fatal("expected out of sync event: got none")
	}
	switch *lastEvent {
	case SynchroOutOfSync:
		want := "numSamplesB=10"
		if !strings.Contains(lastMsg, want) {
			t.Fatalf("wrong event message: got '%s', want '%s'", lastMsg, want)
		}
	default:
		t.Fatalf("wrong event type: got %v, want %v", lastEvent, SynchroOutOfSync)
	}

	f.StreamACallback(xia, xqa, nil, false)
	f.StreamBCallback(xib, xqb, nil, false)

	lastEvent = nil
	f.StreamACallback(xia, xqa, nil, false)
	f.StreamACallback(xia, xqa, nil, false)
	if lastEvent == nil {
		t.Fatal("expected out of sync event: got none")
	}
	switch *lastEvent {
	case SynchroOutOfSync:
		want := "B has not been handled"
		if !strings.Contains(lastMsg, want) {
			t.Fatalf("wrong event message: got '%s', want '%s'", lastMsg, want)
		}
	default:
		t.Fatalf("wrong event type: got %v, want %v", lastEvent, SynchroOutOfSync)
	}

	f.StreamACallback(xia, xqa, nil, false)
	f.StreamBCallback(xib, xqb, nil, false)

	lastEvent = nil
	f.StreamACallback(xia[:10], xqa, nil, false)
	if lastEvent == nil {
		t.Fatal("expected out of sync event: got none")
	}
	switch *lastEvent {
	case SynchroOutOfSync:
		want := "len(xia)=10"
		if !strings.Contains(lastMsg, want) {
			t.Fatalf("wrong event message: got '%s', want '%s'", lastMsg, want)
		}
	default:
		t.Fatalf("wrong event type: got %v, want %v", lastEvent, SynchroOutOfSync)
	}

	f.StreamACallback(xia, xqa, nil, false)
	f.StreamBCallback(xib, xqb, nil, false)

	lastEvent = nil
	f.StreamBCallback(xib[:10], xqb, nil, false)
	if lastEvent == nil {
		t.Fatal("expected out of sync event: got none")
	}
	switch *lastEvent {
	case SynchroOutOfSync:
		want := "len(xib)=10"
		if !strings.Contains(lastMsg, want) {
			t.Fatalf("wrong event message: got '%s', want '%s'", lastMsg, want)
		}
	default:
		t.Fatalf("wrong event type: got %v, want %v", lastEvent, SynchroOutOfSync)
	}
}

// TestSynchroLong tests the operation of Synchro when it
// is forced to wrap around its internal buffer.
func TestSynchroLong(t *testing.T) {
	for i := 0; i < 1000; i++ {
		var (
			wantia int16 = int16(rand.Int())
			wantqa int16 = int16(rand.Int())
			wantib int16 = int16(rand.Int())
			wantqb int16 = int16(rand.Int())
		)
		bufSize := rand.Intn(1000) + 2
		f := NewSynchro(
			bufSize,
			func(xia, xqa, xib, xqb []int16, reset bool) {
				for i := range xia {
					if xia[i] != wantia {
						t.Fatalf("wrong value at %d in xia: got %d, want %d", i, xia[i], wantia)
					}
					wantia++
					if xqa[i] != wantqa {
						t.Fatalf("wrong value at %d in xqa: got %d, want %d", i, xqa[i], wantqa)
					}
					wantqa++
					if xib[i] != wantib {
						t.Fatalf("wrong value at %d in xib: got %d, want %d", i, xib[i], wantib)
					}
					wantib++
					if xqb[i] != wantqb {
						t.Fatalf("wrong value at %d in xqb: got %d, want %d", i, xqb[i], wantqb)
					}
					wantqb++
				}
			},
			nil,
		)
		var (
			total  int
			target int   = bufSize * 42
			valia  int16 = wantia
			valqa  int16 = wantqa
			valib  int16 = wantib
			valqb  int16 = wantqb
		)
		const maxSamples = 2000
		xia := make([]int16, maxSamples)
		xqa := make([]int16, maxSamples)
		xib := make([]int16, maxSamples)
		xqb := make([]int16, maxSamples)
		update := func() int {
			n := rand.Intn(maxSamples)
			for i := 0; i < n; i++ {
				xia[i] = valia
				valia++
				xqa[i] = valqa
				valqa++
				xib[i] = valib
				valib++
				xqb[i] = valqb
				valqb++
			}
			total += n
			return n
		}
		n := update()
		f.StreamACallback(xia[:n], xqa[:n], nil, true)
		f.StreamBCallback(xib[:n], xqb[:n], nil, true)
		for total < target {
			n = update()
			f.StreamACallback(xia[:n], xqa[:n], nil, false)
			f.StreamBCallback(xib[:n], xqb[:n], nil, false)
		}
	}
}

func BenchmarkSynchro(b *testing.B) {
	const (
		cbSamples  = 4000
		numSamples = 1000
	)
	f := NewSynchro(
		cbSamples,
		func(xia, xqa, xib, xqb []int16, reset bool) {
		},
		func(evt SynchroEvent, msg string) {
		},
	)
	xia := make([]int16, numSamples)
	xqa := make([]int16, numSamples)
	xib := make([]int16, numSamples)
	xqb := make([]int16, numSamples)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		f.StreamACallback(xia, xqa, nil, true)
		f.StreamBCallback(xib, xqb, nil, true)
	}
}
