package callback_test

import (
	"fmt"

	"github.com/msiner/sdrplay-go/helpers/callback"
)

func ExampleInterleaveFn() {
	interleave := callback.NewInterleaveFn()
	xi := []int16{1, 3, 5, 7}
	xq := []int16{2, 4, 6, 8}
	x := interleave(xi, xq)
	fmt.Println(x)
	// Output:
	// [1 2 3 4 5 6 7 8]
}
