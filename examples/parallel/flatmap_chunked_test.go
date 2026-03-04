package parallel_examples

import (
	"fmt"

	streams "github.com/coldsmirk/go-streams"
)

// Example: ParallelFlatMap with ordered chunked reordering to bound memory.
func Example_parallelFlatMap_chunked() {
	out := streams.ParallelFlatMap(
		streams.Range(1, 6),
		func(n int) streams.Stream[int] {
			// produce {n, n+100} to visualize order
			return streams.Of(n, n+100)
		},
		streams.WithConcurrency(3),
		streams.WithChunkSize(2),
	).Collect()
	fmt.Println(out)
	// Output:
	// [1 101 2 102 3 103 4 104 5 105]
}

