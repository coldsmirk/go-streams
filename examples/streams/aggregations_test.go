package streams_examples

import (
	"fmt"

	streams "github.com/coldsmirk/go-streams"
)

func Example_sum_and_product() {
	fmt.Println(streams.Sum(streams.Of(1, 2, 3, 4)))
	fmt.Println(streams.SumBy(streams.Of("a", "bb", "ccc"), func(s string) int { return len(s) }))
	fmt.Println(streams.Product(streams.Of(1, 2, 3, 4)))
	// Output:
	// 10
	// 6
	// 24
}

func Example_average_and_averageBy() {
	avg := streams.Average(streams.Of(1, 2, 3, 4, 5))
	avgBy := streams.AverageBy(streams.Of("a", "bb", "ccc"), func(s string) int { return len(s) })
	fmt.Printf("%.1f\n", avg.Get())
	fmt.Printf("%.1f\n", avgBy.Get())
	// Output:
	// 3.0
	// 2.0
}

func Example_minBy_maxBy_minMax() {
	type item struct {
		name string
		val  int
	}
	items := streams.Of(item{"a", 10}, item{"b", 5}, item{"c", 20})
	minByVal := streams.MinBy(items, func(it item) int { return it.val })
	maxByVal := streams.MaxBy(items, func(it item) int { return it.val })
	fmt.Println(minByVal.Get().name, maxByVal.Get().name)

	mm := streams.MinMax(streams.Of(3, 1, 4, 1, 5))
	fmt.Println(mm.Get().First, mm.Get().Second)
	// Output:
	// b c
	// 1 5
}

func Example_median_percentile() {
	less := func(a, b int) bool { return a < b }
	median := streams.Median(streams.Of(5, 1, 9, 3, 7), less)
	p75 := streams.Percentile(streams.Of(1, 2, 3, 4, 5, 6, 7, 8, 9, 10), 0.75, less)
	fmt.Println(median.Get())
	fmt.Println(p75.IsPresent())
	// Output:
	// 5
	// true
}

func Example_contains_and_joining() {
	fmt.Println(streams.Contains(streams.Of(1, 2, 3), 2))
	fmt.Println(streams.Contains(streams.Of(1, 2, 3), 5))
	fmt.Println(streams.Joining(streams.Of("a", "b", "c"), "-"))
	fmt.Println(streams.JoiningWithPrefixSuffix(streams.Of("1", "2"), ",", "[", "]"))
	// Output:
	// true
	// false
	// a-b-c
	// [1,2]
}

func Example_associate_and_associateBy() {
	m := streams.Associate(streams.Of("a", "bb", "ccc"), func(s string) (string, int) { return s, len(s) })
	fmt.Println(m["bb"])
	mb := streams.AssociateBy(streams.Of("a", "bb"), func(s string) int { return len(s) })
	fmt.Println(mb[2])
	// Output:
	// 2
	// bb
}

func Example_groupByTo_indexBy_countBy() {
	g := streams.GroupByTo(streams.Of("a", "bb", "ccc"), func(s string) int { return len(s) }, func(s string) string { return s + "!" })
	fmt.Println(g[1][0], g[2][0])
	idx := streams.IndexBy(streams.Of("a", "bb"), func(s string) int { return len(s) })
	fmt.Println(idx[1], idx[2])
	cnt := streams.CountBy(streams.Of("a", "bb", "ccc", "dd"), func(s string) int { return len(s) })
	fmt.Println(cnt[1], cnt[2], cnt[3])
	// Output:
	// a! bb!
	// a bb
	// 1 2 1
}

func Example_partitionBy_mostCommon() {
	evens, odds := streams.PartitionBy(streams.Of(1, 2, 3, 4, 5), func(n int) bool { return n%2 == 0 })
	fmt.Println(evens, odds)
	mc := streams.MostCommon(streams.Of("a", "b", "a", "c", "a", "b"), 2)
	fmt.Println(mc[0].First, mc[0].Second)
	// Output:
	// [2 4] [1 3 5]
	// a 3
}

func Example_toMap_toSet_foldTo() {
	m := streams.ToMap(streams.Of("a", "bb"), func(s string) int { return len(s) }, func(s string) string { return s })
	fmt.Println(m[1], m[2])
	set := streams.ToSet(streams.Of(1, 2, 2, 3))
	_, ok1 := set[1]
	_, ok5 := set[5]
	fmt.Println(len(set), ok1, ok5)
	ft := streams.FoldTo(streams.Of("a", "bb", "ccc"), 0, func(acc int, s string) int { return acc + len(s) })
	fmt.Println(ft)
	// Output:
	// a bb
	// 3 true false
	// 6
}

func Example_topK_bottomK_frequency() {
	less := func(a, b int) bool { return a < b }
	// TopK returns the k largest elements
	top := streams.TopK(streams.Of(3, 1, 4, 1, 5, 9, 2, 6), 3, less)
	fmt.Println(top)
	// BottomK returns the k smallest elements
	bottom := streams.BottomK(streams.Of(3, 1, 4, 1, 5, 9, 2, 6), 3, less)
	fmt.Println(bottom)
	// Frequency returns a map of element frequencies
	freq := streams.Frequency(streams.Of("a", "b", "a", "c", "a", "b"))
	fmt.Println(freq["a"], freq["b"], freq["c"])
	// Output:
	// [9 6 5]
	// [1 1 2]
	// 3 2 1
}
