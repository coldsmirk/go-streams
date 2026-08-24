package streams_test

import (
	"cmp"
	"fmt"
	"iter"
	"maps"
	"slices"
	"strconv"
	"strings"

	streams "github.com/coldsmirk/go-streams/v2"
)

// Counting and ranking. Frequency builds the histogram, Pairs puts the map back
// into a pipeline, and SortFunc with Take ranks it. Comparing the word after the
// count keeps ties in a defined order.
func Example_wordFrequency() {
	const text = "the quick brown fox jumps over the lazy dog the fox"

	type count struct {
		word string
		n    int
	}

	top := streams.Pairs(streams.Frequency(streams.Of(strings.Fields(text)...))).
		Collapse(func(w string, n int) count { return count{w, n} }).
		SortFunc(func(a, b count) int {
			if c := cmp.Compare(b.n, a.n); c != 0 {
				return c
			}
			return cmp.Compare(a.word, b.word)
		}).
		Take(3).
		Collect()

	for _, c := range top {
		fmt.Println(c.word, c.n)
	}
	// Output:
	// the 3
	// fox 2
	// brown 1
}

// Infinite sources are bounded by the pipeline, not by the source. Nothing past
// the bound is ever produced.
func Example_infiniteStream() {
	powers := streams.Iterate(1, func(n int) int { return n * 2 }).
		TakeWhile(func(n int) bool { return n < 100 }).
		Collect()
	fmt.Println(powers)

	a, b := 0, 1
	fib := streams.Generate(func() int {
		a, b = b, a+b
		return a
	}).Take(8).Collect()
	fmt.Println(fib)
	// Output:
	// [1 2 4 8 16 32 64]
	// [1 1 2 3 5 8 13 21]
}

// Fallible work travels as iter.Seq2[T, error]. Try collects eagerly and returns
// the first error; Ok keeps the pipeline lazy and reports the failure once it
// has drained.
func Example_errorHandling() {
	nums, err := streams.Try(streams.TryMap(streams.Of("1", "2", "3"), strconv.Atoi))
	fmt.Println(nums, err)

	_, err = streams.Try(streams.TryMap(streams.Of("1", "x", "3"), strconv.Atoi))
	fmt.Println(err)

	parsed, failed := streams.Ok(streams.TryMap(streams.Of("4", "5", "boom"), strconv.Atoi))
	fmt.Println(streams.Sum(parsed), failed() != nil)
	// Output:
	// [1 2 3] <nil>
	// strconv.Atoi: parsing "x": invalid syntax
	// 9 true
}

// A Stream is an iter.Seq, so it crosses into and out of the standard library
// with a conversion rather than an adapter.
func Example_stdlibInterop() {
	s := streams.Of(3, 1, 2)
	fmt.Println(slices.Sorted(iter.Seq[int](s)))

	m := map[string]int{"b": 2, "a": 1}
	fmt.Println(streams.From(maps.Keys(m)).SortFunc(cmp.Compare).Collect())

	for i, name := range streams.Of("x", "y").Enumerate() {
		fmt.Printf("%d:%s ", i, name)
	}
	fmt.Println()
	// Output:
	// [1 2 3]
	// [a b]
	// 0:x 1:y
}

// GroupBy returns a plain map, which Pairs feeds back into a pipeline when the
// per-group work is itself a pipeline.
func Example_groupAndAggregate() {
	byCity := streams.Of(users...).GroupBy(func(u user) string { return u.City })

	for _, city := range slices.Sorted(maps.Keys(byCity)) {
		ages := streams.Of(byCity[city]...).Map(func(u user) int { return u.Age })
		avg, _ := streams.Average(ages)
		fmt.Printf("%-8s n=%d avg=%.1f\n", city, len(byCity[city]), avg)
	}
	// Output:
	// London   n=1 avg=36.0
	// NYC      n=2 avg=70.5
	// Portland n=1 avg=54.0
}

// A Stream is single-pass, so each statistic needs one of its own. Where several
// are wanted, build them from a constructor rather than reusing a Stream.
func Example_statistics() {
	ages := func() streams.Stream[int] {
		return streams.Of(users...).Map(func(u user) int { return u.Age })
	}

	lo, _ := streams.Min(ages())
	hi, _ := streams.Max(ages())
	avg, _ := streams.Average(ages())

	fmt.Printf("n=%d sum=%d range=%d..%d avg=%.2f\n",
		ages().Count(), streams.Sum(ages()), lo, hi, avg)
	// Output: n=4 sum=231 range=36..81 avg=57.75
}
