// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/msiner/sdrplay-go/api"
)

func main() {
	flags := flag.NewFlagSet("rspdetect", flag.ExitOnError)
	showVer := flags.Bool("ver", false, "Print API version and exit")
	flags.Usage = func() {
		fmt.Fprintln(flags.Output(), strings.TrimSpace(`
Usage: rspdetect [FLAGS]

rspdetect prints the list of available RSP devices. For all device types,
the device type and serial number are printed. For RSPduo devices,
the available tuners, modes, and current sample rate are also printed.

Flags:
`,
		))
		flags.PrintDefaults()
	}

	// Using ExitOnError
	_ = flags.Parse(os.Args[1:])

	if flags.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "too many arguments provided")
		flags.Usage()
		os.Exit(1)
	}

	lib := api.GetAPI()

	err := lib.Open()
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		// This is overkill, but makes rspdetect useful for detecting
		// this type of error that might otherwise go unnoticed.
		if err := lib.Close(); err != nil {
			log.Fatalf("error on close: %v", err)
		}
	}()

	switch *showVer {
	case true:
		ver, err := lib.ApiVersion()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(ver)
	default:
		devs, err := lib.GetDevices()
		if err != nil {
			log.Fatal(err)
		}

		for _, dev := range devs {
			switch dev.HWVer {
			case api.RSPduo_ID:
				fmt.Printf("%v,%v,%v,%v,%v\n", dev.HWVer, dev.SerNo, dev.Tuner, dev.RspDuoMode, dev.RspDuoSampleFreq)
			default:
				fmt.Printf("%v,%v\n", dev.HWVer, dev.SerNo)
			}
		}
	}
}
