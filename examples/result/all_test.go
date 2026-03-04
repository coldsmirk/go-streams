package result_examples

import (
	"errors"
	"fmt"

	streams "github.com/coldsmirk/go-streams"
)

func Example_collectResultsAll() {
	in := streams.FromResults(
		streams.Ok(1),
		streams.Err[int](errors.New("bad")),
		streams.Ok(2),
	)
	vals, errs := streams.CollectResultsAll(in)
	fmt.Println(vals, len(errs))
	// Output:
	// [1 2] 1
}

