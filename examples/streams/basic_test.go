package streams_examples

import (
	"fmt"

	streams "github.com/coldsmirk/go-streams"
)

// Example: basic filtering and mapping.
func Example_basic() {
	result := streams.Of(1, 2, 3, 4, 5).
		Filter(func(n int) bool { return n%2 == 0 }).
		Map(func(n int) int { return n * 2 }).
		Collect()
	fmt.Println(result)
	// Output:
	// [4 8]
}

