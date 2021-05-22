// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package duo

import (
	"log"
	"math/rand"
	"strings"
	"testing"
)

// TestSynchroBasic tests the basic operation of Synchro.
func TestSynchroBasic(t *testing.T) {
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
					t.Fatalf("wrong value at %d in xia; got %d, want %d", i, xia[i], want)
				}
				want++
				if xqa[i] != want {
					t.Fatalf("wrong value at %d in xqa; got %d, want %d", i, xqa[i], want)
				}
				want++
				if xib[i] != want {
					t.Fatalf("wrong value at %d in xib; got %d, want %d", i, xib[i], want)
				}
				want++
				if xqb[i] != want {
					t.Fatalf("wrong value at %d in xqb; got %d, want %d", i, xqb[i], want)
				}
				want++
			}
		},
		func(evt SynchroEvent, msg string) {
			lastEvent = &evt
			lastMsg = msg
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
	f.UpdateStreamA(xia, xqa, nil, true)
	f.UpdateStreamB(xib, xqb, nil, true)
	f.UpdateStreamA(xia, xqa, nil, false)
	f.UpdateStreamB(xib, xqb, nil, false)
	if numResets != 1 {
		t.Errorf("wrong number of resets; got %d, want 1", numResets)
	}
	if numCallbacks != 2 {
		t.Errorf("wrong number of callbacks; got %d, want 2", numCallbacks)
	}
	if totalFrames != (numSamples * numCallbacks) {
		t.Errorf("wrong total number of frames; got %d, want %d", totalFrames, numSamples*numCallbacks)
	}

	lastEvent = nil
	f.UpdateStreamB(xib, xqb, nil, false)
	if lastEvent == nil {
		t.Fatal("expected out of sync event; got none")
	}
	switch *lastEvent {
	case SynchroOutOfSync:
		want := "A has not been handled"
		if !strings.Contains(lastMsg, want) {
			t.Fatalf("wrong event message; got '%s', want '%s'", lastMsg, want)
		}
	default:
		t.Fatalf("wrong event type; got %v, want %v", lastEvent, SynchroOutOfSync)
	}

	lastEvent = nil
	f.UpdateStreamB(xib, xqb, nil, false)
	if lastEvent != nil {
		t.Fatalf("new event when already out of sync; got %v, want none", lastEvent)
	}

	lastEvent = nil
	f.UpdateStreamA(xia, xqa, nil, false)
	f.UpdateStreamB(xib, xqb, nil, false)
	if lastEvent == nil {
		t.Fatal("expected sync event; got none")
	}
	switch *lastEvent {
	case SynchroSync:
		// good
	default:
		t.Fatalf("wrong event type; got %v, want %v", lastEvent, SynchroSync)
	}

	lastEvent = nil
	f.UpdateStreamA(xia, xqa, nil, false)
	f.UpdateStreamB(xib[:10], xqb[:10], nil, false)
	if lastEvent == nil {
		t.Fatal("expected out of sync event; got none")
	}
	switch *lastEvent {
	case SynchroOutOfSync:
		want := "numSamplesB=10"
		if !strings.Contains(lastMsg, want) {
			t.Fatalf("wrong event message; got '%s', want '%s'", lastMsg, want)
		}
	default:
		t.Fatalf("wrong event type; got %v, want %v", lastEvent, SynchroOutOfSync)
	}

	f.UpdateStreamA(xia, xqa, nil, false)
	f.UpdateStreamB(xib, xqb, nil, false)

	lastEvent = nil
	f.UpdateStreamA(xia, xqa, nil, false)
	f.UpdateStreamA(xia, xqa, nil, false)
	if lastEvent == nil {
		t.Fatal("expected out of sync event; got none")
	}
	switch *lastEvent {
	case SynchroOutOfSync:
		want := "B has not been handled"
		if !strings.Contains(lastMsg, want) {
			t.Fatalf("wrong event message; got '%s', want '%s'", lastMsg, want)
		}
	default:
		t.Fatalf("wrong event type; got %v, want %v", lastEvent, SynchroOutOfSync)
	}
}

// TestSynchroLong tests the operation of Synchro when it
// is forced to wrap around its internal buffer.
func TestSynchroLong(t *testing.T) {
	t.Parallel()

	for i := 0; i < 1000; i++ {
		numSamples := rand.Intn(2000) + 2
		modVal := int16(rand.Intn(100)) + 2
		f := NewSynchro(
			numSamples,
			func(xia, xqa, xib, xqb []int16, reset bool) {
				var want int16
				for i := range xia {
					if xia[i] != want {
						t.Fatalf("wrong value at %d in xia; got %d, want %d", i, xia[i], want)
					}
					want = (want + 1) % modVal
					if xqa[i] != want {
						t.Fatalf("wrong value at %d in xqa; got %d, want %d", i, xqa[i], want)
					}
					want = (want + 1) % modVal
					if xib[i] != want {
						t.Fatalf("wrong value at %d in xib; got %d, want %d", i, xib[i], want)
					}
					want = (want + 1) % modVal
					if xqb[i] != want {
						t.Fatalf("wrong value at %d in xqb; got %d, want %d", i, xqb[i], want)
					}
					want = (want + 1) % modVal
				}
			},
			func(evt SynchroEvent, msg string) {
				log.Printf("%s: %s\n", evt, msg)
			},
		)
		xia := make([]int16, numSamples)
		xqa := make([]int16, numSamples)
		xib := make([]int16, numSamples)
		xqb := make([]int16, numSamples)
		var val int16
		for i := 0; i < numSamples; i++ {
			xia[i] = val
			val = (val + 1) % modVal
			xqa[i] = val
			val = (val + 1) % modVal
			xib[i] = val
			val = (val + 1) % modVal
			xqb[i] = val
			val = (val + 1) % modVal
		}
		f.UpdateStreamA(xia, xqa, nil, true)
		f.UpdateStreamB(xib, xqb, nil, true)
		for i := 0; i < 100; i++ {
			f.UpdateStreamA(xia, xqa, nil, false)
			f.UpdateStreamB(xib, xqb, nil, false)
		}
	}
}

func BenchmarkSynchro(b *testing.B) {
	const numSamples = 1000
	f := NewSynchro(
		numSamples,
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
		f.UpdateStreamA(xia, xqa, nil, true)
		f.UpdateStreamB(xib, xqb, nil, true)
	}
}
