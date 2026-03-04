package join_examples

import (
	"fmt"

	streams "github.com/coldsmirk/go-streams"
)

func Example_joinBy() {
	type U struct{ ID string }
	type O struct{ UID string; Amt int }
	users := streams.Of(U{"u1"}, U{"u2"})
	orders := streams.Of(O{"u1", 10}, O{"u1", 20})
	joined := streams.JoinBy(users, orders, func(u U) string { return u.ID }, func(o O) string { return o.UID }).Collect()
	// Print number of joined rows and first pair IDs
	fmt.Println(len(joined), joined[0].First.ID, joined[0].Second.UID)
	// Output:
	// 2 u1 u1
}

