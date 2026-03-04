package streams_examples

import (
	"fmt"

	streams "github.com/coldsmirk/go-streams"
)

func Example_stream_skip_step_dropWhile() {
	s := streams.Range(1, 10)
	fmt.Println(s.Skip(3).Collect())
	fmt.Println(streams.Range(1, 10).Step(2).Collect())
	fmt.Println(streams.Of(1, 2, 5, 3, 4).DropWhile(func(n int) bool { return n < 3 }).Collect())
	// Output:
	// [4 5 6 7 8 9]
	// [1 3 5 7 9]
	// [5 3 4]
}

func Example_stream_forEachIndexed() {
	var pairs []string
	streams.Of("a", "b", "c").ForEachIndexed(func(i int, v string) {
		pairs = append(pairs, fmt.Sprintf("%d:%s", i, v))
	})
	fmt.Println(pairs)
	// Output:
	// [0:a 1:b 2:c]
}

func Example_stream_isEmpty_isNotEmpty_single() {
	fmt.Println(streams.Empty[int]().IsEmpty(), streams.Of(1).IsEmpty())
	fmt.Println(streams.Empty[int]().IsNotEmpty(), streams.Of(1).IsNotEmpty())
	fmt.Println(streams.Of(42).Single().Get(), streams.Of(1, 2).Single().IsPresent())
	// Output:
	// true false
	// false true
	// 42 false
}

func Example_stream_reduceOptional_fold() {
	ro := streams.Of(1, 2, 3).ReduceOptional(func(a, b int) int { return a + b })
	empty := streams.Empty[int]().ReduceOptional(func(a, b int) int { return a + b })
	fold := streams.Of(1, 2, 3).Fold(10, func(a, b int) int { return a + b })
	fmt.Println(ro.Get(), empty.IsPresent())
	fmt.Println(fold)
	// Output:
	// 6 false
	// 16
}

func Example_stream_findLast_min_max() {
	s := streams.Of(3, 1, 4, 1, 5)
	fmt.Println(s.FindLast(func(n int) bool { return n > 3 }).Get())
	cmp := func(a, b int) int { return a - b }
	fmt.Println(s.Min(cmp).Get(), s.Max(cmp).Get())
	// Output:
	// 5
	// 1 5
}

func Example_stream_sortedStable() {
	// SortedStable preserves relative order of equal elements
	type item struct {
		key   int
		value string
	}
	items := streams.Of(
		item{1, "a"}, item{2, "b"}, item{1, "c"}, item{2, "d"},
	)
	sorted := items.SortedStable(func(a, b item) int { return a.key - b.key }).Collect()
	fmt.Println(sorted[0].value, sorted[1].value, sorted[2].value, sorted[3].value)
	// Output:
	// a c b d
}
