package stream2_examples

import (
	"fmt"
	"iter"

	streams "github.com/coldsmirk/go-streams"
)

func Example_stream2_iter_and_basicOps() {
	// From2 + Seq2 interop
	var seq2 iter.Seq2[string, int] = func(yield func(string, int) bool) {
		if !yield("a", 1) {
			return
		}
		if !yield("b", 2) {
			return
		}
	}
	s2 := streams.From2(seq2)
	fmt.Println(s2.Count())

	// Filter/MapKeys/MapValues/Limit/Skip/Peek/TakeWhile/DropWhile
	var peekSum int
	out := s2.
		Filter(func(k string, v int) bool { return v > 0 }).
		MapKeys(func(k string) string { return k + "!" }).
		MapValues(func(v int) int { return v * 10 }).
		Peek(func(k string, v int) { peekSum += v }).
		TakeWhile(func(k string, v int) bool { return v <= 20 }).
		ToPairs().Collect()
	fmt.Println(len(out), peekSum)

	// Reduce and First
	r := s2.Reduce(streams.NewPair("", 0), func(acc streams.Pair[string, int], k string, v int) streams.Pair[string, int] {
		acc.Second += v
		return acc
	})
	fmt.Println(r.Second, s2.First().IsPresent())
	// Output:
	// 2
	// 2 30
	// 3 true
}

func Example_stream2_mapTo_and_swap() {
	s2 := streams.PairsOf(
		streams.NewPair("k", 1),
		streams.NewPair("k", 2),
	)
	// MapKeysTo, MapValuesTo, MapPairs and SwapKeyValue
	mk := streams.MapKeysTo(s2, func(k string) int { return len(k) })
	mv := streams.MapValuesTo(s2, func(v int) string { return fmt.Sprintf("%d", v) })
	mp := streams.MapPairs(s2, func(k string, v int) (int, string) { return len(k) + v, "x" })
	sw := streams.SwapKeyValue(s2)
	fmt.Println(mk.Count(), mv.Count(), mp.Count(), sw.Count())
	// Output:
	// 2 2 2 2
}
