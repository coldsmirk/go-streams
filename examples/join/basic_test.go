package join_examples

import (
	"fmt"

	streams "github.com/coldsmirk/go-streams"
)

// Example: InnerJoin only keeps matching keys.
func Example_innerJoin() {
	left := streams.PairsOf(
		streams.NewPair("a", 1),
		streams.NewPair("b", 2),
		streams.NewPair("c", 3),
	)
	right := streams.PairsOf(
		streams.NewPair("a", "x"),
		streams.NewPair("c", "z"),
	)
	res := streams.InnerJoin(left, right).Collect()
	fmt.Println(len(res), res[0].Key, res[0].Left, res[0].Right)
	// Output:
	// 2 a 1 x
}

// Example: LeftJoin with Optional semantics.
func Example_leftJoin() {
	left := streams.PairsOf(
		streams.NewPair("a", 1),
		streams.NewPair("b", 2),
	)
	right := streams.PairsOf(
		streams.NewPair("a", "x"),
	)
	res := streams.LeftJoin(left, right).Collect()
	// Print presence of right side for two results
	fmt.Println(len(res), res[0].Right.IsPresent(), res[1].Right.IsPresent())
	// Output:
	// 2 true false
}

