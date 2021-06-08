# sdrplay-go

sdrplay-go is a wrapper of the SDRPlay API for the Go programming language.
It implements direct access to the C API as well as a higher-level declarative
API that takes advantage of features in the Go programming language.

On Windows, build sdrplay-go applications requires Go and the SDRPlay API.
If CGO_ENABLED=0, sdrplay-go defaults to using runtime DLL loading instead of
typical dynamic linking. This allows sdrplay-go applications to be compiled without
using Cgo or requiring a C compiler. If CGO_ENABLED=1, sdrplay-go will build
using Cgo and typical dynamic linking.

On Linux, building sdrplay-go applications requires Go, the SDRPlay API, and
a Cgo-compatible C compiler. The Go environment must be configured with CGO_ENABLED=1.

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
