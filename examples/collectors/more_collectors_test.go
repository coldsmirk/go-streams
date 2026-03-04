package collectors_examples

import (
	"fmt"
	"strings"

	streams "github.com/coldsmirk/go-streams"
)

func Example_collectors_core_and_orderStats() {
	// Core collectors
	first := streams.CollectTo(streams.Of(2, 1), streams.FirstCollector[int]()).Get()
	last := streams.CollectTo(streams.Of(2, 1), streams.LastCollector[int]()).Get()
	reduce := streams.CollectTo(streams.Of(1, 2, 3), streams.ReducingCollector(0, func(a, b int) int { return a + b }))
	joinFull := streams.CollectTo(streams.Of("a", "b"), streams.JoiningCollectorFull(",", "[", "]"))
	count := streams.CollectTo(streams.Of(1, 2, 3), streams.CountingCollector[int]())
	sum := streams.CollectTo(streams.Of(1, 2, 3), streams.SummingCollector[int]())
	avg := streams.CollectTo(streams.Of(1, 2, 3), streams.AveragingCollector[int]()).Get()
	max := streams.CollectTo(streams.Of(1, 5, 3), streams.MaxByCollector[int](func(a, b int) int { return a - b })).Get()
	min := streams.CollectTo(streams.Of(1, 5, 3), streams.MinByCollector[int](func(a, b int) int { return a - b })).Get()
	fmt.Println(first, last, reduce)
	fmt.Println(joinFull)
	fmt.Println(count, sum, int(avg), max, min)
	// Order/ranking collectors
	top2 := streams.CollectTo(streams.Of(5, 1, 4, 3, 2), streams.TopKCollector(2, func(a, b int) bool { return a < b }))
	bot2 := streams.CollectTo(streams.Of(5, 1, 4, 3, 2), streams.BottomKCollector(2, func(a, b int) bool { return a < b }))
	q50 := streams.CollectTo(streams.Of(1, 2, 3, 4, 5), streams.QuantileCollector(0.5, func(a, b int) bool { return a < b })).Get()
	freq := streams.CollectTo(streams.Of("a", "b", "a"), streams.FrequencyCollector[string]())
	hist := streams.CollectTo(streams.Of("a", "bb", "ccc"), streams.HistogramCollector(func(s string) int { return len(s) }))
	fmt.Println(len(top2), len(bot2), q50)
	fmt.Println(freq["a"], len(hist[1]), len(hist[2]), len(hist[3]))
	// Output:
	// 2 1 6
	// [a,b]
	// 3 6 2 5 1
	// 2 2 3
	// 2 1 1 1
}

func Example_collectors_mapping_filtering_flatMapping() {
	base := streams.Of("a", "bb", "ccc")
	mapped := streams.CollectTo(base, streams.MappingCollector(func(s string) int { return len(s) }, streams.ToSliceCollector[int]()))
	filtered := streams.CollectTo(base, streams.FilteringCollector(func(s string) bool { return len(s) > 1 }, streams.JoiningCollector("|")))
	flatMapped := streams.CollectTo(base, streams.FlatMappingCollector(func(s string) streams.Stream[string] { return streams.FromSlice(strings.Split(s, "")) }, streams.JoiningCollector(",")))
	fmt.Println(mapped)
	fmt.Println(filtered)
	fmt.Println(len(flatMapped) > 0)
	// Output:
	// [1 2 3]
	// bb|ccc
	// true
}
