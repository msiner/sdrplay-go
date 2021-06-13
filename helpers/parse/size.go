// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package parse

import (
	"strconv"
	"strings"
)

// SizeInBytes is a helper function to parse a size in bytes
// specified as a command-line argument. For convenience, valid
// arguments can have a suffix of k, K, m, M, g, or G to indicate
// the value is in KiB, MiB, or GiB respectively (e.g. 10M). Any
// text before such a prefix must represent a valid unsigned integer
// value as parsed by strconv.ParseUint(). The return value is the
// parsed size in bytes.
func SizeInBytes(arg string) (uint64, error) {
	var mult uint64 = 1
	arg = strings.ToLower(arg)
	switch {
	case arg == "":
		// do nothing
	case strings.HasSuffix(arg, "k"):
		mult = 1024
		arg = strings.TrimSuffix(arg, "k")
	case strings.HasSuffix(arg, "m"):
		mult = 1024 * 1024
		arg = strings.TrimSuffix(arg, "m")
	case strings.HasSuffix(arg, "g"):
		mult = 1024 * 1024 * 1024
		arg = strings.TrimSuffix(arg, "g")
	}
	size, err := strconv.ParseUint(arg, 10, 64)
	if err != nil {
		return 0, err
	}
	return size * mult, nil
}
