// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func ParseDuration(s string) (time.Duration, error) {
	s = strings.ToLower(s)
	mult := time.Second
	switch {
	case strings.HasSuffix(s, "s"):
		s = strings.TrimSuffix(s, "s")
	case strings.HasSuffix(s, "m"):
		s = strings.TrimSuffix(s, "m")
		mult = time.Minute
	case strings.HasSuffix(s, "h"):
		s = strings.TrimSuffix(s, "h")
		mult = time.Hour
	}
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	if val < 0 {
		return 0, fmt.Errorf("invalid negative duration value; got %d", val)
	}
	return time.Duration(val) * mult, nil
}
