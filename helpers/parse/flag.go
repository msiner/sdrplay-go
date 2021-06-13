// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package parse

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/msiner/sdrplay-go/api"
)

// LNAFlagHelp contains a flag help message for a flag that accepts an
// LNA configuration and has a value that is parsed by LNAFlag.
const LNAFlagHelp = `0-27|0%-100%: LNA State or Percent
Sets the LNA level. Without a % suffix, is an LNA state where 0 provides
the least RF gain reduction. The maximum number of valid states depends
on device type, antenna input, and band. With a % suffix, the LNA gain
as a percent of the maximum where 0% is the minimum amount of gain and
100% is the maximum amount of gain. Specifying as a percent allows
automatic determination of LNA state based on the dependent variables.`

// LNAFlag parses and validates an LNA configuration flag.
func LNAFlag(arg string) (*uint8, *float64, error) {
	if arg == "" {
		return nil, nil, nil
	}
	switch strings.HasSuffix(arg, "%") {
	case true:
		arg = strings.TrimSuffix(arg, "%")
		val, err := strconv.ParseFloat(arg, 64)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid LNA percent; %v", err)
		}
		if val > 100 {
			return nil, nil, fmt.Errorf("invalid LNA percent; got %f, want 0-100", val)
		}
		val = val / 100
		return nil, &val, nil
	default:
		val, err := strconv.ParseUint(arg, 10, 8)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid LNA state; %v", err)
		}
		const maxLNAState = 27
		if val > maxLNAState {
			return nil, nil, fmt.Errorf("invalid LNA state; got %d, want 0-%d", val, maxLNAState)
		}
		res := uint8(val)
		return &res, nil, nil
	}
}

// DecFlagHelp contains a flag help message for a flag that accepts a
// decimation factor and has a value that is checked by DecFlag.
const DecFlagHelp = `1|2|4|8|16|32: Decimation factor
Sets the decimation factor. This will reduce the effective sample rate.
The analog bandwidth will be adjusted automatically to use the best fit
as the effective sample rate decreases.`

// DecFlag validates and converts a decimation factor.
func DecFlag(val uint) (uint8, error) {
	switch val {
	case 1, 2, 4, 8, 16, 32:
		return uint8(val), nil
	default:
		return 0, fmt.Errorf("invalid decimation factor; got %d, want 1|2|4|8|16|32", val)
	}
}

// FsFlagHelp contains a flag help message for a flag that accepts a
// sample rate and has a value that is parsed by FsFlag.
const FsFlagHelp = `FsHz: Sample Rate
Sample rate between 2 MHz and 10 MHz specified in Hz. Can be specified
with k, K, m, M, g, or G suffix to indicate the value is in kHz, MHz,
or GHz respectively (e.g. 2.1M is equal to 2100000)`

// FsFlag parses and validates a sample rate and returns it in Hz.
func FsFlag(arg string) (float64, error) {
	return SampleRate(arg)
}

// WarmFlagHelp contains a flag help message for a flag that accepts a
// warm-up duration and has a value that is parsed and validated by
// WarmFlag.
const WarmFlagHelp = `seconds: Warmup Time
Run the radio for the specified number of seconds to warm up and
stabilize performance before capture. It also avoids sample drops
typically encountered when the stream is first starting. During
the warmup period, samples are discarded. The maximum value allowed
is 60 seconds.`

// WarmFlag parses and validates a warm-up duration.
func WarmFlag(val uint) (time.Duration, error) {
	const MaxWarm = 60 * time.Second

	warm := time.Duration(val) * time.Second
	if warm > MaxWarm {
		return 0, fmt.Errorf("invalid warmup duration; got %v, want <= %v", warm, MaxWarm)
	}

	return warm, nil
}

// AGCCtlFlagHelp contains a flag help message for a flag that accepts an
// AGC control configuration and has a value that is parsed and validated by
// AGCCtlFlag.
const AGCCtlFlagHelp = `disable|enable|5|50|100: AGC Control
Disable or enable AGC with the specified loop bandwidth.`

// AGCCtlFlag parses and validates an AGC control configuration.
func AGCCtlFlag(arg string) (api.AgcControlT, error) {
	switch arg {
	case "disable":
		return api.AGC_DISABLE, nil
	case "5":
		return api.AGC_5HZ, nil
	case "50":
		return api.AGC_50HZ, nil
	case "100":
		return api.AGC_100HZ, nil
	case "enable":
		return api.AGC_CTRL_EN, nil
	default:
		return 0, fmt.Errorf("invalid AGC control value; got %s, want disable|enable|5|50|100", arg)
	}
}

// AGCSetFlagHelp contains a flag help message for a flag that accepts an
// AGC set point and has a value that is parsed and validated by
// AGCSetFlag.
const AGCSetFlagHelp = `dBFS: AGC Set Point
AGC set point in dBFS.`

// AGCSetFlag parses and validates an AGC set point.
func AGCSetFlag(val int) (int32, error) {
	if val > 0 {
		return 0, fmt.Errorf("invalid AGC set point; got %d dBFS, want <= 0", val)
	}
	return int32(val), nil
}

// DuoTunerFlagHelp contains a flag help message for a flag that accepts an
// RSPduo tuner specifier and has a value that is parsed and validated by
// DuoTunerFlag.
const DuoTunerFlagHelp = `a|1|b|2|either: RSPDuo Tuner Selection
Select which RSPDuo tuner to use if the selected device is an RSPduo. If
"either" is specified, tuner A will be used if available. Otherwise, tuner
B will be used if available. If the selected device is not an RSPduo, this
option will have no effect.`

// DuoTunerSelect is an enum type returned from DuoTunerFlag to convey
// the parsed tuner selection. This type has been created because there
// is not a single type or value from the C API that conveys the meaning
// needed by DuoTunerFlag.
type DuoTunerSelect string

const (
	// DuoTunerFlagEither specifies that either tuner can be selected.
	DuoTunerFlagEither DuoTunerSelect = "either"
	// DuoTunerFlagA specifies that only tuner A or 1 should be selected.
	DuoTunerFlagA DuoTunerSelect = "a"
	// DuoTunerFlagB specifies that only tuner B or 2 should be selected.
	DuoTunerFlagB DuoTunerSelect = "b"
)

// DuoTunerFlag parses and validates an RSPduo tuner selection. Valid
// values are "either", "a", "1", "b", or "2".
func DuoTunerFlag(arg string) (DuoTunerSelect, error) {
	switch arg {
	case "either":
		return DuoTunerFlagEither, nil
	case "a", "1":
		return DuoTunerFlagA, nil
	case "b", "2":
		return DuoTunerFlagB, nil
	default:
		return "", fmt.Errorf("invalid duotuner value; got %s, want a|1|b|2|either", arg)
	}
}

// SerialsFlagHelp contains a flag help message for a flag that accepts a
// comma-separated list of device serial numbers and has a value that is
// parsed and validated by SerialsFlag.
const SerialsFlagHelp = `serialA,serialB,...: Device Serial Numbers
Provide a comma-separated list of one or more device serial numbers
to select from. If a device with one of the provided serial numbers
is not found, no device will be selected. The value "any" matches
any serial number.`

// SerialsFlag parses and validates a comma-separated list of device
// serial numbers.
func SerialsFlag(arg string) ([]api.SerialNumber, error) {
	var serials []api.SerialNumber
	if arg == "any" {
		return serials, nil
	}
	// There is no defined spec for serial number format, but it seems
	// like are only alpha-numeric characters. No spaces or special chars.
	rx := regexp.MustCompile(`^[[:alnum:]]+?`)
	parts := strings.Split(arg, ",")
	if len(parts) == 0 {
		return nil, fmt.Errorf("no serials specified; got %s", arg)
	}
	for _, part := range parts {
		if !rx.MatchString(part) {
			return nil, fmt.Errorf("invalid serial; got %s", part)
		}
		serials = append(serials, api.ParseSerialNumber(part))
	}
	return serials, nil
}

// USBFlagHelp contains a flag help message for a flag that accepts a
// USB transfer mode and has a value that is parsed and validated by
// USBFlag.
const USBFlagHelp = `isoch|bulk: USB Transfer Mode
Select to configure the device in either isochronous or bulk mode.`

// USBFlag parses and validates a USB transfer mode specification.
func USBFlag(arg string) (api.TransferModeT, error) {
	switch strings.ToLower(arg) {
	case "isoch":
		return api.ISOCH, nil
	case "bulk":
		return api.BULK, nil
	default:
		return 0, fmt.Errorf("invalid usb mode; got %s, want isoch|bulk", arg)
	}
}

// HiZFlagHelp contains a flag help message for a boolean flag that
// requests use of a High-Z port when true.
const HiZFlagHelp = `Enable High-Z Port
If using an RSP2 or RSPduo, enable the High-Z port.`

// DxAntFlagHelp contains a flag help message for a flag that accepts an
// RSPdx antenna selection and has a value that is parsed and validated by
// DxAntFlag.
const DxAntFlagHelp = `a|b|c: RSPdx Antenna
Select RSPdx antenna input.`

// DxAntFlag parses and validates an RSPdx antenna selection. Valid
// values are "a", "b", or "c".
func DxAntFlag(arg string) (api.RspDx_AntennaSelectT, error) {
	switch strings.ToLower(arg) {
	case "a":
		return api.RspDx_ANTENNA_A, nil
	case "b":
		return api.RspDx_ANTENNA_B, nil
	case "c":
		return api.RspDx_ANTENNA_C, nil
	default:
		return 0, fmt.Errorf("invalid RSPdx antenna; got %s, want a|b|c", arg)
	}
}

// Rsp2AntFlagHelp contains a flag help message for a flag that accepts an
// RSP2 antenna selection and has a value that is parsed and validated by
// Rsp2AntFlag.
const Rsp2AntFlagHelp = `a|b: RSP2 Antenna
Select RSP2 antenna input.`

// Rsp2AntFlag parses and validates an RSP2 antenna selection. Valid
// values are "a" or "b".
func Rsp2AntFlag(arg string) (api.Rsp2_AntennaSelectT, error) {
	switch strings.ToLower(arg) {
	case "a":
		return api.Rsp2_ANTENNA_A, nil
	case "b":
		return api.Rsp2_ANTENNA_B, nil
	default:
		return 0, fmt.Errorf("invalid RSP2 antenna; got %s, want a|b", arg)
	}
}
