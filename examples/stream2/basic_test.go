package stream2_examples

import (
	"fmt"

	streams "github.com/coldsmirk/go-streams"
)

// Example: Stream2 keys and values from explicit pairs (deterministic order).
func Example_stream2() {
	s2 := streams.PairsOf(
		streams.NewPair("a", 1),
		streams.NewPair("b", 2),
	)
	fmt.Println(s2.Keys().Collect())
	fmt.Println(s2.Values().Collect())
	// Output:
	// [a b]
	// [1 2]
}

