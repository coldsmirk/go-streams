package parallel_examples

import (
	"fmt"

	streams "github.com/coldsmirk/go-streams"
)

// Example: parallel map with default ordered output.
func Example_parallelMap() {
	squares := streams.ParallelMap(
		streams.Range(1, 6),
		func(n int) int { return n * n },
		streams.WithConcurrency(2),
	).Collect()
	fmt.Println(squares)
	// Output:
	// [1 4 9 16 25]
}

