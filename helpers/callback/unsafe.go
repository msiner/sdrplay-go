// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package callback

import (
	"io"
	"reflect"
	"runtime"
	"unsafe"
)

// FastWrite writes the scalar sample data in x to out. It is
// "fast" because it uses the unsafe and reflect packages to create
// a []byte that points to the same data as x and has the appropriate
// length and capacity. This allows it to avoid copying data. It,
// therefore, writes the int16 scalars out in native byte-order. To
// write int16 scalars without unsafe casts or with configurable
// byte-order, see NewWrite().
//
//go:nosplit
func FastWrite(out io.Writer, x []int16) (int, error) {
	const sizeOfScalar = 2
	if len(x) == 0 {
		return 0, nil
	}
	numBytes := len(x) * sizeOfScalar
	var xb []byte
	hx := (*reflect.SliceHeader)(unsafe.Pointer(&xb))
	hx.Cap = numBytes
	hx.Len = numBytes
	hx.Data = uintptr(unsafe.Pointer(&x[0]))
	n, err := out.Write(xb)
	runtime.KeepAlive(x)
	return n, err
}
