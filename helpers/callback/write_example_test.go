package callback_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"

	"github.com/msiner/sdrplay-go/helpers/callback"
)

func ExampleWriteFn() {
	// Declare and store ByteOrder to use to create the WriteFn and
	// for use reading back the written samples for verification.
	order := binary.BigEndian
	write := callback.NewWriteFn(order)

	x := []int16{1, 2, 3, 4, 5, 6, 7, 8}
	fmt.Printf("Original Samples: %v\n", x)

	// Destination io.Writer
	buf := bytes.NewBuffer(nil)

	n, err := write(buf, x)
	fmt.Printf("Num Bytes Written: %d\n", n)
	if err != nil {
		log.Fatal(err)
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

func ExampleFloat32WriteFn() {
	// Declare and store ByteOrder to use to create the WriteFn and
	// for use reading back the written samples for verification.
	order := binary.BigEndian
	write := callback.NewFloat32WriteFn(order)

	x := []float32{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8}
	fmt.Printf("Original Samples: %v\n", x)

	// Destination io.Writer
	buf := bytes.NewBuffer(nil)

	n, err := write(buf, x)
	fmt.Printf("Num Bytes Written: %d\n", n)
	if err != nil {
		log.Fatal(err)
	}

	// Create a buffer big enough to read back all of the written samples.
	res := make([]float32, buf.Len()/4)

	// Now use buf as an io.Reader for readback verification.
	if err := binary.Read(buf, order, &res); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Written Samples: %v\n", res)

	// Output:
	// Original Samples: [0.1 0.2 0.3 0.4 0.5 0.6 0.7 0.8]
	// Num Bytes Written: 32
	// Written Samples: [0.1 0.2 0.3 0.4 0.5 0.6 0.7 0.8]
}

func ExampleComplex64WriteFn() {
	// Declare and store ByteOrder to use to create the WriteFn and
	// for use reading back the written samples for verification.
	order := binary.BigEndian
	write := callback.NewComplex64WriteFn(order)

	x := []complex64{complex(0.1, 0.2), complex(0.3, 0.4), complex(0.5, 0.6), complex(0.7, 0.8)}
	fmt.Printf("Original Samples: %v\n", x)

	// Destination io.Writer
	buf := bytes.NewBuffer(nil)

	n, err := write(buf, x)
	fmt.Printf("Num Bytes Written: %d\n", n)
	if err != nil {
		log.Fatal(err)
	}

	// Create a buffer big enough to read back all of the written samples.
	res := make([]complex64, buf.Len()/8)

	// Now use buf as an io.Reader for readback verification.
	if err := binary.Read(buf, order, &res); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Written Samples: %v\n", res)

	// Output:
	// Original Samples: [(0.1+0.2i) (0.3+0.4i) (0.5+0.6i) (0.7+0.8i)]
	// Num Bytes Written: 32
	// Written Samples: [(0.1+0.2i) (0.3+0.4i) (0.5+0.6i) (0.7+0.8i)]
}
