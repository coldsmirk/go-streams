package numeric_examples

import (
	"fmt"

	streams "github.com/coldsmirk/go-streams"
)

func Example_absFloat_negative_runningProduct() {
	// AbsFloat for floating point numbers
	fmt.Println(streams.AbsFloat(streams.Of(-1.5, 2.5, -3.5)).Collect())
	// Negative filters only negative numbers
	fmt.Println(streams.Negative(streams.Of(-3, -1, 0, 2, 5)).Collect())
	// RunningProduct: cumulative product
	fmt.Println(streams.RunningProduct(streams.Of(1, 2, 3, 4)).Collect())
	// Output:
	// [1.5 2.5 3.5]
	// [-3 -1]
	// [1 2 6 24]
}
