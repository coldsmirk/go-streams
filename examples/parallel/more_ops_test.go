package parallel_examples

import (
	"fmt"
	"slices"
	"sync/atomic"

	streams "github.com/coldsmirk/go-streams"
)

func Example_parallelFilter_prefetch() {
	// Parallel filter evens, then Prefetch before collect to overlap producer/consumer.
	evens := streams.ParallelFilter(
		streams.Range(1, 10),
		func(n int) bool { return n%2 == 0 },
		streams.WithConcurrency(3),
	).Collect()
	// Demonstrate Prefetch
	pref := streams.Prefetch(streams.Of(1, 2, 3, 4), 2).Collect()
	fmt.Println(evens)
	fmt.Println(pref)
	// Output:
	// [2 4 6 8]
	// [1 2 3 4]
}

func Example_parallelForEach_and_collect() {
	var sum int64
	streams.ParallelForEach(
		streams.Range(1, 6),
		func(n int) { atomic.AddInt64(&sum, int64(n)) },
		streams.WithConcurrency(2),
	)
	collected := streams.ParallelCollect(
		streams.Range(5, 10),
		streams.WithConcurrency(3),
		streams.WithOrdered(false),
	)
	slices.Sort(collected)
	fmt.Println(sum)
	fmt.Println(collected)
	// Output:
	// 15
	// [5 6 7 8 9]
}

