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

## Quick Start

This section describes how to download the sdrplay-go repo and build the included
example tools. After executing the following steps correctly, a copy of the
sdrplay-go repository is available as an environment to build and modify Go
applications added to the sdrplay-go module. The steps do not cover the use case
of creating a new module that requires sdrplay-go.

1. Download and install the SDRPlay API for your platform.
https://www.sdrplay.com/start-here/

2. Clone the repository and move into the directory.
```sh
git clone https://github.com/msiner/sdrplay-go.git
cd sdrplay-go
```

3. On Windows, copy the SDRPlay API DLL into the repo directory.
This step is not required on Linux.
```powershell
cp "C:\Program Files\SDRplay\API\x64\sdrplay_api.dll"
```

4. Run all unit tests to verify a compatible build environment.
```sh
go test ./...
```

5. Build the rspdetect tool.
```sh
go build ./cmd/rspdetect
```

6. Connect an RSP device.

7. Run the rspdetect tool.
The example output below shows a single connected RSPduo device.
```sh
./rspdetect.exe
RSPduo_ID,5555555555,Tuner_Both,Single|Dual|Primary,0
```

8. Build the remaining example tools.
```sh
go build ./cmd/rspwav
go build ./cmd/rspudp
go build ./cmd/duowav
go build ./cmd/duoudp
```

## Create a New Application

This section describes how to create a new Go application that uses
sdrplay-go.

1. Create a new directory for you application and initialize it as
a Go module.
```sh
mkdir sdrapp
cd sdrapp
go mod init local/sdrapp
```

2. Create and open new file at ```./cmd/app```

3. Add the following Go code to the file.
```go

```