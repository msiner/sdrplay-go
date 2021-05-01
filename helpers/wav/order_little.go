// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// +build arm arm64 386 amd64 ppc64le mipsle mips64le riscv64 wasm

package wav

import "encoding/binary"

const IsBigEndian = false

var NativeOrder = binary.LittleEndian
