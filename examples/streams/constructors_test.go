package streams_examples

import (
	"fmt"
	"slices"

	streams "github.com/coldsmirk/go-streams"
)

func Example_constructors_basic() {
	// Of / FromSlice
	fmt.Println(streams.Of(1, 2, 3).Collect())
	fmt.Println(streams.FromSlice([]string{"a", "b"}).Collect())
	// Range / RangeClosed
	fmt.Println(streams.Range(3, 6).Collect())
	fmt.Println(streams.RangeClosed(3, 6).Collect())
	// Concat
	fmt.Println(streams.Concat(streams.Of(1, 2), streams.Of(3)).Collect())
	// Empty
	fmt.Println(streams.Empty[int]().Collect())
	// Output:
	// [1 2 3]
	// [a b]
	// [3 4 5]
	// [3 4 5 6]
	// [1 2 3]
	// []
}

func Example_constructors_advanced() {
	// Generate / Iterate (bounded with Limit)
	g := streams.Generate(func() int { return 1 }).Limit(3).Collect()
	it := streams.Iterate(1, func(n int) int { return n * 2 }).Limit(4).Collect()
	// Repeat / RepeatForever / Cycle (bounded)
	rep := streams.Repeat("x", 3).Collect()
	cyc := streams.Cycle(1, 2).Limit(5).Collect()
	// From map via ToPairs for deterministic order
	mp := map[string]int{"a": 1, "b": 2}
	keys := slices.Collect(streams.FromMap(mp).Keys().Seq())
	slices.Sort(keys)
	fmt.Println(g, it)
	fmt.Println(rep, cyc)
	fmt.Println(keys)
	// Output:
	// [1 1 1] [1 2 4 8]
	// [x x x] [1 2 1 2 1]
	// [a b]
}

