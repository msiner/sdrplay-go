// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// +build ppc64 mips mips64 s390x

package wav

import "encoding/binary"

const IsBigEndian = true

var NativeOrder = binary.BigEndian
