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
using Cgo or requiring a C compiler.

If CGO_ENABLED=1, sdrplay-go will build using Cgo and typical dynamic linking. To
build with runtime DLL loading while keeping CGO_ENABLED=1, declare the "dll" tag
during build (e.g. ```go build -tags dll ./...```).

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

## License

Copyright 2021 Mark Siner. All rights reserved. Use of this source
code is governed by a MIT-style license that can be found in the LICENSE file.
