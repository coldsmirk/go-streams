package streams_examples

import (
	"fmt"

	streams "github.com/coldsmirk/go-streams"
)

func Example_distinct_and_untilChanged() {
	d1 := streams.Distinct(streams.Of(1, 1, 2, 2, 3)).Collect()
	d2 := streams.DistinctBy(streams.Of("a", "b", "aa"), func(s string) int { return len(s) }).Collect()
	d3 := streams.DistinctUntilChanged(streams.Of(1, 1, 2, 1, 1)).Collect()
	d4 := streams.DistinctUntilChangedBy(streams.Of("a", "A", "a"), func(a, b string) bool { return len(a) == len(b) }).Collect()
	fmt.Println(d1)
	fmt.Println(d2)
	fmt.Println(d3)
	fmt.Println(d4)
	// Output:
	// [1 2 3]
	// [a aa]
	// [1 2 1]
	// [a]
}

func Example_sorting_and_reverse() {
	s := streams.Of(3, 1, 2)
	sorted := s.Sorted(func(a, b int) int { return a - b }).Collect()
	reversed := s.Reverse().Collect()
	sb := streams.SortedBy(streams.Of("bbb", "a", "cc"), func(s string) int { return len(s) }).Collect()
	ssb := streams.SortedStableBy(streams.Of("b", "aa", "cc", "a"), func(s string) int { return len(s) }).Collect()
	fmt.Println(sorted)
	fmt.Println(reversed)
	fmt.Println(sb)
	fmt.Println(ssb)
	// Output:
	// [1 2 3]
	// [2 1 3]
	// [a cc bbb]
	// [b a aa cc]
}
