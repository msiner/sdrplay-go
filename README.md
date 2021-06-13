# sdrplay-go

[![Go Reference](https://pkg.go.dev/badge/github.com/msiner/sdrplay-go.svg)](https://pkg.go.dev/github.com/msiner/sdrplay-go)

sdrplay-go is a Go programming language wrapper of the SDRplay service based API
(version 3.0 and greater) that can be used to access SDRplay RSP receivers. It
implements direct access to the C API as well as a higher-level declarative API
that takes advantage of features in the Go programming language. On Windows, Cgo
is not required.

## Build Configuration

### Linux
On Linux, building sdrplay-go applications requires Go, the SDRplay API, and
a Cgo-compatible C compiler. The Go environment must be configured with `CGO_ENABLED=1`.

### Windows
On Windows, building sdrplay-go applications requires Go and the SDRplay API.
If `CGO_ENABLED=0`, sdrplay-go defaults to using runtime DLL loading instead of
typical dynamic linking. This allows sdrplay-go applications to be compiled without
using Cgo or requiring a C compiler. In this mode, sdrplay-go will adaptively
find the sdrplay_api.dll in either the current working directory, the same
directory as the executable, a directory in the Path environment variable, or
in the normal SDRplay API install location `C:\Program Files\SDRplay\API`.

If `CGO_ENABLED=1`, sdrplay-go will build using Cgo and typical dynamic linking. To
build with adaptive runtime DLL loading while keeping `CGO_ENABLED=1`, declare the
"dll" tag during build (e.g. `go build -tags dll ./...`). An application built
using Cgo must be able to resolve the dynamic link by finding sdrplay_api.dll in
either the current working directory, the same directory as the executable, a directory
in the `Path` environment variable

## Quickstart

This section describes how to create a new Go application that uses
sdrplay-go.

1. Create a new directory for you application and initialize it as
a Go module.
```sh
mkdir sdrapp
cd sdrapp
go mod init my.local/sdrapp
```

2. Create and open a new file called `main.go`.

3. Add the following Go code to the file. This main function will
select and configure a single RSP device, receive a single stream
callback, print the number of samples received, and then exit. 
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

## Device Support

This library has been tested with the following devices. If you verify
either correct or incorrect operation with a device not listed here,
please open an issue to share your experience.

- SDRplay RSPduo

## Legal

Use of this source code is governed by a MIT-style license that can be
found in the LICENSE file.

Use of the SDRplay API library is governed by a separate license distributed
with the SDRplay API.

SDRplay is the trading name of SDRplay Limited a company registered in England # 0903524.

This source code is not a product of SDRplay Limited and the author has no
affiliation with SDRplay Limited.
