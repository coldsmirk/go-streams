package parallel_examples

import (
	"fmt"
	"slices"

	streams "github.com/coldsmirk/go-streams"
)

func Example_parallel_unordered_and_collect_reduce() {
	// Unordered map for throughput, then sort for deterministic presentation.
	out := streams.ParallelMap(
		streams.Range(1, 6),
		func(n int) int { return n * n },
		streams.WithOrdered(false),
	).Collect()
	slices.Sort(out)
	fmt.Println(out)

	// ParallelReduce with associative op
	sum := streams.ParallelReduce(streams.Range(1, 6), 0, func(a, b int) int { return a + b })
	fmt.Println(sum)
	// Output:
	// [1 4 9 16 25]
	// 15
}

