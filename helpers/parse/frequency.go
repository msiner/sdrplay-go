// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package parse

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseFrequency is a helper function to parse a frequency value
// specified as a command-line argument. For convenience, valid
// arguments can have a suffix of k, K, m, M, g, or G to indicate
// the value is in kHz, MHz, or GHz respectively (e.g. 1.42G). Any
// text before such a prefix must represent a valid floating point
// value as parsed by strconv.ParseFloat(). The return value is the
// parsed frequency in Hz.
func ParseFrequency(arg string) (float64, error) {
	var mult float64 = 1
	arg = strings.ToLower(arg)
	switch {
	case arg == "":
		// do nothing
	case strings.HasSuffix(arg, "k"):
		mult = 1000
		arg = strings.TrimSuffix(arg, "k")
	case strings.HasSuffix(arg, "m"):
		mult = 1000 * 1000
		arg = strings.TrimSuffix(arg, "m")
	case strings.HasSuffix(arg, "g"):
		mult = 1000 * 1000 * 1000
		arg = strings.TrimSuffix(arg, "g")
	}
	freq, err := strconv.ParseFloat(arg, 64)
	if err != nil {
		return 0, err
	}
	return freq * mult, nil
}

// ParseTuneFrequency is a wrapper around ParseFrequency that also guarantees
// the result is a valid tune frequency for RSP hardware. Specifically, it
// will return an error if the frequency is less than 1 kHz or greater than
// 2 GHz.
func ParseTuneFrequency(arg string) (float64, error) {
	freq, err := ParseFrequency(arg)
	if err != nil {
		return 0, err
	}
	if freq < 1e3 || freq > 2e9 {
		return 0, fmt.Errorf("invalid tune frequency; got %f Hz, want 1kHz<=Freq<=2GHz", freq)
	}
	return freq, nil
}

// ParseSampleRate is a wrapper around ParseFrequency that also guarantees
// the result is a valid sample rate for RSP hardware. Specifically, it
// will return an error if the rate is less than 2 MHz or greater than
// 10 MHz.
func ParseSampleRate(arg string) (float64, error) {
	freq, err := ParseFrequency(arg)
	if err != nil {
		return 0, err
	}
	if freq < 2e6 || freq > 10e6 {
		return 0, fmt.Errorf("invalid sample rate; got %f Hz, want 2MHz<=Rate<=10MHz", freq)
	}
	return freq, nil
}
