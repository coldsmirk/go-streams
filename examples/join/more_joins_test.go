package join_examples

import (
	"fmt"

	streams "github.com/coldsmirk/go-streams"
)

func Example_more_joins() {
	left := streams.PairsOf(
		streams.NewPair("a", 1),
		streams.NewPair("b", 2),
	)
	right := streams.PairsOf(
		streams.NewPair("a", "x"),
		streams.NewPair("c", "y"),
	)

	rj := streams.RightJoin(left, right).Collect()
	fj := streams.FullJoin(left, right).Collect()
	ljw := streams.LeftJoinWith(left, right, "NA").Collect()
	rjw := streams.RightJoinWith(left, right, -1).Collect()

	// JoinBy variants
	type U struct{ ID string }
	type O struct{ UID string }
	users := streams.Of(U{"u1"}, U{"u2"})
	orders := streams.Of(O{"u1"})
	ljb := streams.LeftJoinBy(users, orders, func(u U) string { return u.ID }, func(o O) string { return o.UID }).Collect()
	semi := streams.SemiJoin(streams.PairsOf(streams.NewPair("k", 1)), streams.PairsOf(streams.NewPair("k", 2))).CollectPairs()
	anti := streams.AntiJoin(streams.PairsOf(streams.NewPair("k", 1)), streams.PairsOf(streams.NewPair("z", 2))).CollectPairs()
	sb := streams.SemiJoinBy(streams.Of(1, 2, 3), streams.Of(2, 4), func(v int) int { return v }, func(v int) int { return v }).Collect()
	ab := streams.AntiJoinBy(streams.Of(1, 2, 3), streams.Of(2, 4), func(v int) int { return v }, func(v int) int { return v }).Collect()

	fmt.Println(len(rj), len(fj), len(ljw), len(rjw))
	fmt.Println(len(ljb), len(semi), len(anti), sb, ab)
	// Output:
	// 2 3 2 2
	// 2 1 1 [2] [1 3]
}

