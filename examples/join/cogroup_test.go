package join_examples

import (
	"fmt"
	"slices"

	streams "github.com/coldsmirk/go-streams"
)

func Example_coGroup() {
	left := streams.PairsOf(
		streams.NewPair("a", 1),
		streams.NewPair("b", 2),
		streams.NewPair("a", 3),
	)
	right := streams.PairsOf(
		streams.NewPair("a", "x"),
		streams.NewPair("c", "y"),
	)
	cg := streams.CoGroup(left, right).Collect()
	// Normalize order for deterministic print
	slices.SortFunc(cg, func(a, b streams.CoGrouped[string, int, string]) int {
		if a.Key < b.Key {
			return -1
		}
		if a.Key > b.Key {
			return 1
		}
		return 0
	})
	// Print sizes per key
	for _, g := range cg {
		fmt.Printf("%s:%d|%d ", g.Key, len(g.Left), len(g.Right))
	}
	fmt.Println()
	// Output:
	// a:2|1 b:1|0 c:0|1 
}

