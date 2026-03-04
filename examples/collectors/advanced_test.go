package collectors_examples

import (
	"fmt"
	"strings"

	streams "github.com/coldsmirk/go-streams"
)

func Example_collectorsAdvanced() {
	// Partitioning
	part := streams.CollectTo(streams.Of(1, 2, 3, 4), streams.PartitioningByCollector(func(v int) bool { return v%2 == 0 }))
	fmt.Println(len(part[true]), len(part[false]))

	// ToMap with merge
	m := streams.CollectTo(
		streams.Of("a", "aa", "b"),
		streams.ToMapCollectorMerging(func(s string) int { return len(s) }, func(s string) string { return s }, func(a, b string) string { return a + "|" + b }),
	)
	fmt.Println(len(m), strings.Contains(m[1], "b"))

	// Teeing average: sum and count then merge
	sumC := streams.SummingCollector[int]()
	cntC := streams.CountingCollector[int]()
	avg := streams.CollectTo(streams.Of(1, 2, 3, 4), streams.TeeingCollector(sumC, cntC, func(sum, cnt int) float64 {
		return float64(sum) / float64(cnt)
	}))
	fmt.Printf("%.1f\n", avg)
	// Output:
	// 2 2
	// 2 true
	// 2.5
}

