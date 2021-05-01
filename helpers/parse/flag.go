// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package parse

import (
	"flag"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/msiner/sdrplay-go/api"
)

type FlagSet interface {
	Var(value flag.Value, name string, usage string)
}

const LNAFlagHelp = `0-27|0%-100%: LNA State or Percent
Sets the LNA level. Without a % suffix, is an LNA state where 0 provides
the least RF gain reduction. The maximum number of valid states depends
on device type, antenna input, and band. With a % suffix, the LNA gain
as a percent of the maximum where 0% is the minimum amount of gain and
100% is the maximum amount of gain. Specifying as a percent allows
automatic determination of LNA state based on the dependent variables.`

func CheckLNAFlag(arg string) (*uint8, *float64, error) {
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

const LNAPercentFlagHelp = `0-100: LNA Gain Percent
Sets the LNA gain as a percent of the maximum where 0 is the minimum amount
of gain and 100 is the maximum amount of gain.`

func CheckLNAPercentFlag(val uint) (float64, error) {
	if val > 100 {
		return 0, fmt.Errorf("invalid LNA percent; got %d, want 0-100", val)
	}
	return float64(val) / 100, nil
}

const DecFlagHelp = `1|2|4|8|16|32: Decimation factor
Sets the decimation factor. This will reduce the effective sample rate.
The analog bandwidth will be adjusted automatically to use the best fit
as the effective sample rate decreases.`

func CheckDecFlag(val uint) (uint8, error) {
	switch val {
	case 1, 2, 4, 8, 16, 32:
		return uint8(val), nil
	default:
		return 0, fmt.Errorf("invalid decimation factor; got %d, want 1|2|4|8|16|32", val)
	}
}

const FsFlagHelp = `FsHz: Sample Rate
Sample rate between 2 MHz and 10 MHz specified in Hz. Can be specified
with k, K, m, M, g, or G suffix to indicate the value is in kHz, MHz,
or GHz respectively (e.g. 2.1M is equal to 2100000)`

func ParseFsFlag(arg string) (float64, error) {
	return ParseSampleRate(arg)
}

const WarmFlagHelp = `seconds: Warmup Time
Run the radio for the specified number of seconds to warm up and
stabilize performance before capture. It also avoids sample drops
typically encountered when the stream is first starting. During
the warmup period, samples are discarded. The maximum value allowed
is 60 seconds.`

func ParseWarmFlag(val uint) (time.Duration, error) {
	const MAX_WARM = 60 * time.Second

	warm := time.Duration(val) * time.Second
	if warm > MAX_WARM {
		return 0, fmt.Errorf("invalid warmup duration; got %v, want <= %v", warm, MAX_WARM)
	}

	return warm, nil
}

const AGCCtlFlagHelp = `disable|enable|5|50|100: AGC Control
Disable or enable AGC with the specified loop bandwidth.`

func ParseAGCCtlFlag(arg string) (api.AgcControlT, error) {
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

const AGCSetFlagHelp = `dBFS: AGC Set Point
AGC set point in dBFS.`

func ParseAGCSetFlag(val int) (int32, error) {
	if val > 0 {
		return 0, fmt.Errorf("invalid AGC set point; got %d dBFS, want <= 0", val)
	}
	return int32(val), nil
}

type DuoTunerFlag string

const (
	DuoTunerFlagEither DuoTunerFlag = "either"
	DuoTunerFlagA      DuoTunerFlag = "a"
	DuoTunerFlagB      DuoTunerFlag = "b"
	DuoTunerFlagHelp                = `a|1|b|2|either: RSPDuo Tuner Selection
Select which RSPDuo tuner to use if the selected device is an RSPduo. If
"either" is specified, tuner A will be used if available. Otherwise, tuner
B will be used if available. If the selected device is not an RSPduo, this
option will have no effect.`
)

func ParseDuoTunerFlag(arg string) (DuoTunerFlag, error) {
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

const SerialsFlagHelp = `serialA,serialB,...: Device Serial Numbers
Provide a comma-separated list of one or more device serial numbers
to select from. If a device with one of the provided serial numbers
is not found, no device will be selected. The value "any" matches
any serial number.`

func ParseSerialsFlag(arg string) ([]api.SerialNumber, error) {
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

const USBFlagHelp = `isoch|bulk: USB Transfer Mode
Select to configure the device in either isochronous or bulk mode.`

func ParseUSBFlag(arg string) (api.TransferModeT, error) {
	switch strings.ToLower(arg) {
	case "isoch":
		return api.ISOCH, nil
	case "bulk":
		return api.BULK, nil
	default:
		return 0, fmt.Errorf("invalid usb mode; got %s, want isoch|bulk", arg)
	}
}

const HiZFlagHelp = `Enable High-Z Port
If using an RSP2 or RSPduo, enable the High-Z port.`

const DxAntFlagHelp = `a|b|c: RSPdx Antenna
Select RSPdx antenna input.`

func ParseDxAntFlag(arg string) (api.RspDx_AntennaSelectT, error) {
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

const Rsp2AntFlagHelp = `a|b: RSP2 Antenna
Select RSP2 antenna input.`

func ParseRsp2AntFlag(arg string) (api.Rsp2_AntennaSelectT, error) {
	switch strings.ToLower(arg) {
	case "a":
		return api.Rsp2_ANTENNA_A, nil
	case "b":
		return api.Rsp2_ANTENNA_B, nil
	default:
		return 0, fmt.Errorf("invalid RSP2 antenna; got %s, want a|b", arg)
	}
}
