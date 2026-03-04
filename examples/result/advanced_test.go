package result_examples

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	streams "github.com/coldsmirk/go-streams"
)

func Example_resultAdvanced() {
	// FlatMapErr across lines
	lines := streams.Of("1 2", "x 3")
	split := func(s string) streams.Stream[string] { return streams.FromSlice(strings.Fields(s)) }
	out := streams.FlatMapErr(lines, func(s string) (streams.Stream[int], error) {
		// parse each token in line; if any fails -> error for the line
		vals := streams.MapErrTo(split(s), strconv.Atoi)
		vs, err := streams.CollectResults(vals)
		if err != nil {
			return streams.Empty[int](), err
		}
		return streams.FromSlice(vs), nil
	})
	okVals, errs := streams.CollectResultsAll(out)
	fmt.Println(okVals, len(errs))

	// Unwrap and defaults
	rs := streams.FromResults(streams.Ok(1), streams.Err[int](errors.New("e")), streams.Ok(2))
	fmt.Println(streams.UnwrapOrDefault(rs, 0).Collect())
	// Output:
	// [1 2] 1
	// [1 0 2]
}

