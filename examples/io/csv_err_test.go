package io_examples

import (
	"fmt"
	"strings"

	streams "github.com/coldsmirk/go-streams"
)

func Example_csvErr() {
	r := strings.NewReader("k,v\nA,1\nB\nC,3\n")
	ok, errs := streams.CollectResultsAll(streams.FromCSVWithHeaderErr(r))
	fmt.Println(len(ok), len(errs)) // rows parsed OK vs errors
	// Output:
	// 2 1
}

