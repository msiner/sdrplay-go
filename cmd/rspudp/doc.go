// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

/*
rspudp is a command-line utility that interleaves, packetizes, and transmits
the a single stream from any RSP device.

	Usage: rspudp [FLAGS] <tuneHz>

	rspudp connects to an available RSP device, configures it, and
	sends the received samples to the specified UDP target. The samples
	are interleaved and framed in the payload as:
	[I1,Q1,I2,Q2,...,IN,QN]

	Arguments:
	tuneHz
			Tuner RF frequency in Hz is a mandatory argument. It can
			be specified with k, K, m, M, g, or G suffix to indicate the
			value is in kHz, MHz, or GHz respectively (e.g. 1.42G).

	Flags:
	-agcctl string
			disable|enable|5|50|100: AGC Control
			Disable or enable AGC with the specified loop bandwidth. (default "enable")
	-agcset int
			dBFS: AGC Set Point
			AGC set point in dBFS. (default -30)
	-big
			Write samples with big-endian byte order
	-dec uint
			1|2|4|8|16|32: Decimation factor
			Sets the decimation factor. This will reduce the effective sample rate.
			The analog bandwidth will be adjusted automatically to use the best fit
			as the effective sample rate decreases. (default 1)
	-duotuner string
			a|1|b|2|either: RSPDuo Tuner Selection
			Select which RSPDuo tuner to use if the selected device is an RSPduo. If
			"either" is specified, tuner A will be used if available. Otherwise, tuner
			B will be used if available. If the selected device is not an RSPduo, this
			option will have no effect. (default "either")
	-dxant string
			a|b|c: RSPdx Antenna
			Select RSPdx antenna input. (default "a")
	-fs string
			FsHz: Sample Rate
			Sample rate between 2 MHz and 10 MHz specified in Hz. Can be specified
			with k, K, m, M, g, or G suffix to indicate the value is in kHz, MHz,
			or GHz respectively (e.g. 2.1M is equal to 2100000) (default "6M")
	-hiz
			Enable High-Z Port
			If using an RSP2 or RSPduo, enable the High-Z port.
	-lif
			Use low-IF mode. In low-IF mode, the effective sample rate, before decimation
			is 2 MHz. When -lif is specified, the -fs option cannot be used to configure
			the sample rate.
	-lna string
			0-27|0%-100%: LNA State or Percent
			Sets the LNA level. Without a % suffix, is an LNA state where 0 provides
			the least RF gain reduction. The maximum number of valid states depends
			on device type, antenna input, and band. With a % suffix, the LNA gain
			as a percent of the maximum where 0% is the minimum amount of gain and
			100% is the maximum amount of gain. Specifying as a percent allows
			automatic determination of LNA state based on the dependent variables. (default "50%")
	-pay uint
			UDP payload size in bytes. This must be small enough to fit in
			the network MTU with IP and UDP headers. It must also be a multiple
			of the 4 byte frame size. (default 1400)
	-remote string
			Target host address or name and UDP port (default "127.0.0.1:1234")
	-rsp2ant string
			a|b: RSP2 Antenna
			Select RSP2 antenna input. (default "a")
	-seq
			Insert a 64-bit sequence number at the beginning of each packet.
			This will use 8 bytes of the specified payload size.
	-serials string
			serialA,serialB,...: Device Serial Numbers
			Provide a comma-separated list of one or more device serial numbers
			to select from. If a device with one of the provided serial numbers
			is not found, no device will be selected. The value "any" matches
			any serial number. (default "any")
	-usb string
			isoch|bulk: USB Transfer Mode
			Select to configure the device in either isochronous or bulk mode. (default "isoch")
	-warm uint
			seconds: Warmup Time
			Run the radio for the specified number of seconds to warm up and
			stabilize performance before capture. It also avoids sample drops
			typically encountered when the stream is first starting. During
			the warmup period, samples are discarded. The maximum value allowed
			is 60 seconds. (default 2)
*/
package main
