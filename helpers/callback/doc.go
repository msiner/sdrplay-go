// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

/*
Package callback provides types and functions for handling sample data
from a stream callback.

To avoid extra allocations and complexity in the callack path, many of
the types in this package are designed with internal buffers. This allows,
for example, the InterleaveFn function type to simply accept two sample
buffers and return a single interleaved buffer. The returning function
still owns the buffer, but this design allows functions to be chained
together and the user does not need to worry about providing an adequately
sized buffer as an input argument.

	write := NewWriteFn(binary.LittleEndian)
	interleave := NewInterleaveFn()
	...
	n, err := write(out, interleave(xi, xq))

The buffers returned by such functions must not be stored, reused, or
otherwise escape the callback. Copy samples out of the buffer in they
need be used later.
*/
package callback
