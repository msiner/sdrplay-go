package callback_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"

	"github.com/msiner/sdrplay-go/helpers/callback"
)

func ExampleFastWrite() {
	x := []int16{1, 2, 3, 4, 5, 6, 7, 8}
	fmt.Printf("Original Samples: %v\n", x)

	// Destination io.Writer
	buf := bytes.NewBuffer(nil)

	// FastWrite doesn't need or use a buffer, so it doesn't need to
	// be created like WriteFn. It can be called directly.
	n, err := callback.FastWrite(buf, x)
	fmt.Printf("Num Bytes Written: %d\n", n)
	if err != nil {
		log.Fatal(err)
	}

	// Since FastWrite uses the native byte order, we need to determine
	// which order to use when doing readback with the encoding/binary
	// package.
	b := buf.Bytes()
	var order binary.ByteOrder = binary.LittleEndian
	// We know first sample is 0x0001. In big-endian, first byte will
	// be most-significant and, therefore, zero.
	if b[0] == 0 {
		order = binary.BigEndian
	}

	// Create a buffer big enough to read back all of the written samples.
	res := make([]int16, buf.Len()/2)

	// Now use buf as an io.Reader for readback verification.
	if err := binary.Read(buf, order, &res); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Written Samples: %v\n", res)

	// Output:
	// Original Samples: [1 2 3 4 5 6 7 8]
	// Num Bytes Written: 16
	// Written Samples: [1 2 3 4 5 6 7 8]
}
