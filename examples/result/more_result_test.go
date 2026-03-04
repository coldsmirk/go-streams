package result_examples

import (
	"errors"
	"fmt"

	streams "github.com/coldsmirk/go-streams"
)

func Example_result_map_and_filters() {
	r := streams.Ok(1)
	r2 := streams.MapResultTo(r, func(v int) string { return fmt.Sprintf("v=%d", v) })
	r3 := streams.FlatMapResult(r, func(v int) streams.Result[int] { return streams.Ok(v * 10) })
	fmt.Println(r2.IsOk(), r3.Unwrap())

	// FilterErr / FilterOk / FilterErrs
	in := streams.FromResults(streams.Ok(1), streams.Err[int](errors.New("e1")), streams.Ok(2), streams.Err[int](errors.New("e2")))
	okOnly := streams.FilterOk(in).Collect()
	errOnly := streams.FilterErrs(in).Collect()
	fmt.Println(okOnly, len(errOnly))
	// Partition and unwrap variants
	vals, errs := streams.PartitionResults(in)
	fmt.Println(len(vals), len(errs))
	unwrap := streams.UnwrapResults(streams.FromResults(streams.Ok(1), streams.Ok(2))).Collect()
	take := streams.TakeUntilErr(streams.FromResults(streams.Ok(1), streams.Err[int](errors.New("x")), streams.Ok(2))).Collect()
	fmt.Println(unwrap, take)
	// TryCollect: trap panic
	rc := streams.TryCollect(streams.Of(1, 2))
	fmt.Println(rc.IsOk())
	// Output:
	// true 10
	// [1 2] 2
	// 2 2
	// [1 2] [1]
	// true
}

