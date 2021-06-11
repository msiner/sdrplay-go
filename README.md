# sdrplay-go

sdrplay-go is a Go programming language wrapper of the SDRplay API that can be
used to access SDRplay RSP receivers. It implements direct access to the C API
as well as a higher-level declarative API that takes advantage of features in
the Go programming language. On Windows, Cgo is not required.

## Build Configuration

### Linux
On Linux, building sdrplay-go applications requires Go, the SDRplay API, and
a Cgo-compatible C compiler. The Go environment must be configured with CGO_ENABLED=1.

### Windows
On Windows, building sdrplay-go applications requires Go and the SDRplay API.
If CGO_ENABLED=0, sdrplay-go defaults to using runtime DLL loading instead of
typical dynamic linking. This allows sdrplay-go applications to be compiled without
using Cgo or requiring a C compiler. In this mode, sdrplay-go will adaptively
find the sdrplay_api.dll in either the current working directory, the same
directory as the executable, a directory in the Path environment variable, or
in the normal SDRplay API install location `C:\Program Files\SDRplay\API`.

If CGO_ENABLED=1, sdrplay-go will build using Cgo and typical dynamic linking. To
build with runtime DLL loading while keeping CGO_ENABLED=1, declare the "dll" tag
during build (e.g. `go build -tags dll ./...`). An application built using Cgo
must be able to resolve the dynamic link by finding sdrplay_api.dll in either the
current working directory, the same directory as the executable, a directory in
the Path environment variable

## Quickstart

This section describes how to create a new Go application that uses
sdrplay-go.

1. Create a new directory for you application and initialize it as
a Go module.
```sh
mkdir sdrapp
cd sdrapp
go mod init local/sdrapp
```

2. Create and open new file at ```./main.go```

3. Add the following Go code to the file. This main function will
select and configure a single RSP device, receive a single callback
of samples, and then exit. 
```go
package main

import (
	"context"
	"log"

	"github.com/msiner/sdrplay-go/api"
	"github.com/msiner/sdrplay-go/session"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := session.Run(
		ctx,
		session.WithSelector(
			session.WithDuoModeSingle(),
			session.WithDuoTunerEither(),
		),
		session.WithDeviceConfig(
			session.WithSingleChannelConfig(
				session.WithTuneFreq(1e9),
				session.WithAGC(api.AGC_CTRL_EN, -20),
			),
		),
		session.WithStreamACallback(
			func(xi, xq []int16, params *api.StreamCbParamsT, reset bool) {
				select {
				case <-ctx.Done():
				default:
					log.Printf("Got %d samples\n", params.NumSamples)
					cancel()
				}
			},
		),
	)
	if err != nil {
		log.Println(err)
	}
}
```

4. Compile and run the new application.
```sh
go run ./main.go
```
```
2021/06/07 20:51:20 Run
2021/06/07 20:51:21 Got 1008 samples
2021/06/07 20:51:22 context canceled
```

## Utilities

This module also provides some command-line utilities in the cmd package.
These utilities are designed to be useful tools on their own as well as
examples of how to use the module.

#### rspdetect
The rspdetect command simply opens the API, requests a list of available
devices, and then prints the information for each device.

#### rspwav
The rspwav command opens an available RSP device and configures a single
stream. Samples are recorded to a stereo WAV file where the I component
is left and the Q component is right.

#### rspudp
The rspudp command opens an available RSP device and configures a single
stream. Samples are packetized into UDP payloads and sent to the configured
destination address and port. It also supports an optional 64-bit packet
counter at the start of each payload. Sample components are interleaved
(e.g. I1,Q1,I2,Q2,...,IN,QN).

#### duowav
The duowav command is similar to rspwav, but it uses an RSPduo device
configured in dual-tune mode with both tuners configured identically.
Components and samples are interleaved
(e.g. I1a,Q1a,I1b,Q1b,...,INa,QNa,INb,QNb). In the WAV multi-channel
audo context, this maps to Ia, Qa, Ib, and Qb as front-left, front-right,
front-center, and low-frequency respectively.

#### duoudp
The duoudp command is similar to rspudp, but it uses and RSPduo device
configured in dual-tuner mode with both tuners configured identically.
Components and samples are interleaved
(e.g. I1a,Q1a,I1b,Q1b,...,INa,QNa,INb,QNb).

#### duocorr
The duocorr command is a rudimentary tool included for testing purposes.
It uses an RSPduo device configured in dual-tuner mode with both tuners
configured identically. For a specified duration, it calculates a
cross-correlation peak offset between channel A and channel B using the
magnitude of the samples.

## License

Use of this source code is governed by a MIT-style license that can be
found in the LICENSE file.
