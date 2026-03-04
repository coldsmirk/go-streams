package specialized_examples

import (
	"fmt"

	streams "github.com/coldsmirk/go-streams"
)

func Example_mergeSortedNHeap() {
	s := streams.MergeSortedNHeap(
		func(a, b int) int { return a - b },
		streams.Of(1, 4, 7),
		streams.Of(2, 5, 8),
		streams.Of(3, 6, 9),
	).Collect()
	fmt.Println(s)
	// Output:
	// [1 2 3 4 5 6 7 8 9]
}

