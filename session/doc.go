// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

/*
Package session implements a high-level API on top of and as an
alternative to the low-level the api package that is mostly a one-to-one
mapping of the C API. The API provided in this package is designed
using a functional options pattern to wrap common configuration tasks
in composable functions for a highly declarative API.
*/
package session
