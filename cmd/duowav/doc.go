// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

/*
duowav is a command-line utility that records to a 4-channel WAV file
the components of the two streams of an RSPduo in dual-tuner mode.

	Usage: duowav [FLAGS] <tuneHz> <fileBytes>

	duowav connects to an available RSPduo device, configures it, and writes
	the received samples to a file in WAV format. The WAV file has four
	channels (interpreted by audio software as front left, front right, front
	center, and low frequency) corresponding to tuner A I, tuner A Q,
	tuner B I, and tuner B Q respectively where I and Q are the real and
	imaginary components respectively. The samples are stored
	as 16-bit signed integers in little-endian format with the components
	interleaved (e.g. AI1,AQ1,BI1,BQ1,...,AIn,AQn,BIn,BQn).

	Arguments:
	tuneHz
			Tuner RF frequency in Hz is a mandatory argument. It can
			be specified with k, K, m, M, g, or G suffix to indicate the
			value is in kHz, MHz, or GHz respectively (e.g. 1.42G).
	fileBytes
			Maximum output file size in bytes. It can be specified with
			k, K, m, M, g, or G suffix to indicate the value is in KiB, MiB,
			or GiB respectively (e.g. 10M)
			NOTE: WAV files cannot exceed 4 GiB.

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
	-float
			Write samples in floating-point format
	-hiz
			Enable High-Z Port
			If using an RSP2 or RSPduo, enable the High-Z port.
	-lna string
			0-27|0%-100%: LNA State or Percent
			Sets the LNA level. Without a % suffix, is an LNA state where 0 provides
			the least RF gain reduction. The maximum number of valid states depends
			on device type, antenna input, and band. With a % suffix, the LNA gain
			as a percent of the maximum where 0% is the minimum amount of gain and
			100% is the maximum amount of gain. Specifying as a percent allows
			automatic determination of LNA state based on the dependent variables. (default "50%")
	-maxfs
			Use the maximum 8MHz sample rate.
			This will deliver 12-bit ADC resolution, but with slightly better
			anti-aliasing performance at the widest bandwidth. This mode is
			only available at 1.536 MHz analog bandwidth. The default mode is
			to use a 6 MHz ADC sample rate. That mode delivers 14-bit ADC
			resolution, but with slightly inferior anti-aliasing performance
			at the widest bandwidth. The default mode is also compatible with
			analog bandwidths of 1.536 MHz, 600 kHz, 300 kHz, and 200 kHz.
			6 MHz operation should result in a slightly lower CPU load.
	-out string
			Write WAV file to specified path. (default "duo.wav")
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
