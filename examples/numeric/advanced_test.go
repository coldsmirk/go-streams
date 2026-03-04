package numeric_examples

import (
	"fmt"

	streams "github.com/coldsmirk/go-streams"
)

func Example_numericAdvanced() {
	s := streams.Of(-3, -1, 0, 2, 5)
	fmt.Println(streams.MinValue(s).Get(), streams.MaxValue(streams.Of(1, 9, 2)).Get())
	fmt.Println(streams.Clamp(streams.Of(-2, 3, 10), 0, 5).Collect())
	fmt.Println(streams.Abs(streams.Of(-1, 2, -3)).Collect())
	fmt.Println(streams.Scale(streams.Of(1, 2), 10).Collect())
	fmt.Println(streams.Offset(streams.Of(1, 2), 5).Collect())
	fmt.Println(streams.Positive(streams.Of(-1, 0, 2)).Collect())
	fmt.Println(streams.NonZero(streams.Of(0, 1)).Collect())
	stats := streams.GetStatistics(streams.Of(1, 2, 3)).Get()
	fmt.Println(stats.Count, stats.Sum, int(stats.Average))
	// Output:
	// -3 9
	// [0 3 5]
	// [1 2 3]
	// [10 20]
	// [6 7]
	// [2]
	// [1]
	// 3 6 2
}
