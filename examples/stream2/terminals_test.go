package stream2_examples

import (
	"fmt"
	"slices"

	streams "github.com/coldsmirk/go-streams"
)

func Example_stream2_terminals() {
	s2 := streams.PairsOf(
		streams.NewPair("a", 1),
		streams.NewPair("b", 2),
	)
	// ToMap2 converts Stream2 to map
	m := streams.ToMap2(s2)
	keys := slices.Collect(streams.FromMap(m).Keys().Seq())
	slices.Sort(keys)
	fmt.Println(keys, m["a"], m["b"])
	// GroupValues groups values by key
	s3 := streams.PairsOf(
		streams.NewPair("k", 1),
		streams.NewPair("k", 2),
		streams.NewPair("z", 3),
	)
	gv := streams.GroupValues(s3)
	fmt.Println(gv["k"], gv["z"])
	// ReduceByKeyWithInit uses custom init
	rbk := streams.ReduceByKeyWithInit(s3, func() []int { return nil }, func(acc []int, v int) []int { return append(acc, v) })
	fmt.Println(len(rbk["k"]), len(rbk["z"]))
	// Output:
	// [a b] 1 2
	// [1 2] [3]
	// 2 1
}

func Example_stream2_skip_limit() {
	s2 := streams.PairsOf(
		streams.NewPair("a", 1),
		streams.NewPair("b", 2),
		streams.NewPair("c", 3),
		streams.NewPair("d", 4),
	)
	// Skip and Limit work on Stream2
	skipped := s2.Skip(1).Limit(2).CollectPairs()
	fmt.Println(len(skipped), skipped[0].First, skipped[1].First)
	// Output:
	// 2 b c
}

func Example_stream2_dropWhile() {
	s2 := streams.PairsOf(
		streams.NewPair("a", 1),
		streams.NewPair("b", 2),
		streams.NewPair("a", 3),
	)
	dropped := s2.DropWhile(func(k string, v int) bool { return v < 2 }).CollectPairs()
	fmt.Println(len(dropped), dropped[0].Second)
	// Output:
	// 2 2
}
