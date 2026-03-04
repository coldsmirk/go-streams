package stream2_examples

import (
	"fmt"

	streams "github.com/coldsmirk/go-streams"
)

func Example_stream2_advanced() {
	s2 := streams.PairsOf(
		streams.NewPair("k1", 1),
		streams.NewPair("k1", 2),
		streams.NewPair("k2", 2),
	)
	// ReduceByKey and DistinctKeys/Values
	sum := streams.ReduceByKey(s2, func(a, b int) int { return a + b })
	fmt.Println(sum["k1"], sum["k2"])
	fmt.Println(streams.DistinctKeys(s2).CollectPairs())
	fmt.Println(streams.DistinctValues(s2).CollectPairs())
	// MapPairs and SwapKeyValue
	swapped := streams.SwapKeyValue(streams.MapPairs(s2, func(k string, v int) (string, string) {
		return k + "!", fmt.Sprintf("%d", v)
	}))
	fmt.Println(len(swapped.CollectPairs()))
	// Output:
	// 3 2
	// [{k1 1} {k2 2}]
	// [{k1 1} {k1 2}]
	// 3
}
