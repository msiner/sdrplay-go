// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

/*
rspdetect is a command-line utility that searches for available RSP
devices and prints a list of available devices.

	Usage: rspdetect [FLAGS]

	rspdetect prints the list of available RSP devices. For all device types,
	the device type and serial number are printed. For RSPduo devices,
	the available tuners, modes, and current sample rate are also printed.

	Flags:
  	-ver
        	Print API version and exit
*/
package main
