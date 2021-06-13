// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

/*
Package api provides direct access to the C API.

Functions from the C API are provided by implementations of the
API interface instead of being exported at the package level. See
the documentation for the API interface type in this package for
function signatures and implementation details. Specifically, any
deviation from the behavior or function signature of the C API is
documented there.

See the SDRplay API documentation for more information about the
types and functions. The names of types, constants, and functions
are carried over from the C API, but without the "sdrplay_api_" prefix.
*/
package api
