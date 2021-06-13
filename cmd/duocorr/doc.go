// Copyright 2021 Mark Siner. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

/*
duocorr is a command-line utility that measures the cross-correlation of
the two streams of an RSPduo in dual-tuner mode.

	Usage: duocorr [FLAGS] <tuneHz> <duration>

	duocorr connects to an available RSPduo device, configures it
	in dual-tuner mode with both tuners having the same configuration.
	It then continuously calculates a correlation between channel A and
	channel B for the specified time duration.

	Correlation results will be printed to standard out. The "-inter"
	flag determines the interval for printing results. The results
	reported at the end of each interval period reflect all measurements
	taken during that period.

	Since the dual tuners of the RSPduo are not phase-coherent, duocorr
	performs a correlation of the magnitudes of samples. This is reported
	as the offset of the correlation peak in the number of samples plus
	or minus from the center. The z-score of the correlation peak is also
	computed and reported.

	duocorr is a very rudimentary tool included for testing purposes. It
	can serve the practical application of measuring or detecting a time
	offset in an antenna and cabling setup that might make dual-antenna
	diversity processing less reliable. Also, duocorr is not optimized.
	This was intentional to be able to include this tool in the repository
	without introducing a module dependency on a third-party DSP library.

	Arguments:
	tuneHz
			Tuner RF frequency in Hz is a mandatory argument. It can
			be specified with k, K, m, M, g, or G suffix to indicate the
			value is in kHz, MHz, or GHz respectively (e.g. 1.42G).
	duration
			Amount of time to process correlations before exiting. It can be
			specified with a s, S, m, M, h, or H suffix to indicate the value
			is in seconds, minutes, or hours respectively (e.g. 10s). Without
			a suffix provided, the value is interpreted as seconds.

	Flags:
	-agcctl string
			disable|enable|5|50|100: AGC Control
			Disable or enable AGC with the specified loop bandwidth. (default "enable")
	-agcset int
			dBFS: AGC Set Point
			AGC set point in dBFS. (default -30)
	-dec uint
			1|2|4|8|16|32: Decimation factor
			Sets the decimation factor. This will reduce the effective sample rate.
			The analog bandwidth will be adjusted automatically to use the best fit
			as the effective sample rate decreases. (default 1)
	-hiz
			Enable High-Z Port
			If using an RSP2 or RSPduo, enable the High-Z port.
	-inter string
			measurement reporting interval (default "1s")
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
