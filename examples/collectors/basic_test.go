package collectors_examples

import (
	"fmt"

	streams "github.com/coldsmirk/go-streams"
)

// Example: TopK convenience backed by collectors.
func Example_topK() {
	top2 := streams.TopK(streams.Of(5, 1, 4, 3, 2), 2, func(a, b int) bool { return a < b })
	fmt.Println(top2)
	// Output:
	// [5 4]
}

