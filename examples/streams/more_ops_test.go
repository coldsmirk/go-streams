package streams_examples

import (
	"fmt"
	"iter"

	streams "github.com/coldsmirk/go-streams"
)

func Example_zip3_and_repeatForever() {
	z3 := streams.Zip3(streams.Of(1, 2), streams.Of("a", "b"), streams.Of(true, false)).Collect()
	fmt.Println(z3[0], z3[1])
	// RepeatForever is infinite, so limit it
	rf := streams.RepeatForever("x").Limit(3).Collect()
	fmt.Println(rf)
	// Output:
	// {1 a true} {2 b false}
	// [x x x]
}

func Example_flattenSeq() {
	seqOf := func(vals ...int) iter.Seq[int] {
		return func(yield func(int) bool) {
			for _, v := range vals {
				if !yield(v) {
					return
				}
			}
		}
	}
	src := streams.Of(seqOf(1, 2), seqOf(3, 4))
	flat := streams.FlattenSeq(src).Collect()
	fmt.Println(flat)
	// Output:
	// [1 2 3 4]
}

func Example_filterErr() {
	s := streams.Of(1, 2, 3, 4, 5)
	// FilterErr returns Result for each element
	results := streams.FilterErr(s, func(n int) (bool, error) {
		if n == 3 {
			return false, fmt.Errorf("skip 3")
		}
		return n%2 == 0, nil
	}).Collect()
	okCount := 0
	errCount := 0
	for _, r := range results {
		if r.IsOk() {
			okCount++
		} else {
			errCount++
		}
	}
	fmt.Println(okCount, errCount)
	// Output:
	// 2 1
}
