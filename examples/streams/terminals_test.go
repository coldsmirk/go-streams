package streams_examples

import (
	"fmt"
	"slices"
	"strings"

	streams "github.com/coldsmirk/go-streams"
)

func Example_terminals_queries() {
	s := streams.Of(1, 2, 3, 4, 5)
	fmt.Println(s.Count())
	fmt.Println(s.First().Get(), s.Last().Get())
	fmt.Println(s.FindFirst(func(v int) bool { return v > 3 }).Get())
	fmt.Println(s.AnyMatch(func(v int) bool { return v%2 == 0 }))
	fmt.Println(s.AllMatch(func(v int) bool { return v > 0 }))
	fmt.Println(s.NoneMatch(func(v int) bool { return v < 0 }))
	// Output:
	// 5
	// 1 5
	// 4
	// true
	// true
	// true
}

func Example_terminals_reduce_map_set_group() {
	s := streams.Of(1, 2, 3)
	fmt.Println(s.Reduce(0, func(a, b int) int { return a + b }))
	fmt.Println(streams.FoldTo(streams.Of("a", "bb"), 0, func(acc int, v string) int { return acc + len(v) }))
	m := streams.ToMap(streams.Of("a", "bb"), func(s string) int { return len(s) }, func(s string) string { return s })
	keys := slices.Collect(streams.FromMap(m).Keys().Seq())
	slices.Sort(keys)
	fmt.Println(keys)
	set := streams.ToSet(streams.Of(1, 2, 2))
	fmt.Println(len(set))
	g := streams.GroupBy(streams.Of("a", "aa"), func(s string) int { return len(s) })
	fmt.Println(len(g[1]), len(g[2]))
	fmt.Println(streams.Joining(streams.Of("a", "b"), ","))
	fmt.Println(streams.JoiningWithPrefixSuffix(streams.Of("a", "b"), ",", "[", "]"))
	// Output:
	// 6
	// 3
	// [1 2]
	// 2
	// 1 1
	// a,b
	// [a,b]
}

func Example_terminals_index_and_more() {
	s := streams.Of(10, 20, 30)
	fmt.Println(s.At(1).Get(), s.Nth(2).Get())
	fmt.Println(streams.Contains(streams.Of("a", "b"), "b"))
	fmt.Println(streams.Associate(streams.Of("a", "bb"), func(s string) (string, int) { return s, len(s) })["bb"])
	fmt.Println(streams.AssociateBy(streams.Of("a", "bb"), func(s string) int { return len(s) })[2])
	fmt.Println(streams.IndexBy(streams.Of("a", "bb"), func(s string) string { return strings.ToUpper(s) })["BB"]) // value for key "BB"
	fmt.Println(streams.CountBy(streams.Of("a", "aa"), func(s string) int { return len(s) })[2])
	fmt.Println(streams.Frequencies(streams.Of(1, 1, 2))[1])
	// Output:
	// 20 30
	// true
	// 2
	// bb
	// bb
	// 1
	// 2
}
