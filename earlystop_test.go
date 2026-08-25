package streams

import (
	"iter"
	"testing"

	"github.com/stretchr/testify/assert"
)

// The iter contract says yield panics if it is called after returning false.
// Every operation that forwards elements must therefore honour a false result
// and stop. These tests break out of each operation after one element, which
// panics if the early-return path is missing.

func breakAfterOne[T any](t *testing.T, name string, s Stream[T]) {
	t.Helper()
	n := 0
	// A panic ends the check here, as the original recover-based harness did:
	// the count below describes a run that completed.
	if !assert.NotPanicsf(t, func() {
		for range iter.Seq[T](s) {
			n++
			break
		}
	}, "%s: yielding after the consumer stopped", name) {
		return
	}
	assert.Equalf(t, 1, n, "%s: elements consumed before the break", name)
}

func breakAfterOne2[K, V any](t *testing.T, name string, s Stream2[K, V]) {
	t.Helper()
	n := 0
	if !assert.NotPanicsf(t, func() {
		for range iter.Seq2[K, V](s) {
			n++
			break
		}
	}, "%s: yielding after the consumer stopped", name) {
		return
	}
	assert.Equalf(t, 1, n, "%s: pairs consumed before the break", name)
}

func TestIntermediateOpsHonourEarlyStop(t *testing.T) {
	src := func() Stream[int] { return Of(1, 2, 3, 4, 5) }
	long := func() Stream[int] { return Range(0, 100) }

	breakAfterOne(t, "Filter", src().Filter(func(int) bool { return true }))
	breakAfterOne(t, "Map", src().Map(func(i int) int { return i }))
	breakAfterOne(t, "FlatMap", src().FlatMap(func(i int) Stream[int] { return Of(i, i) }))
	breakAfterOne(t, "Scan", src().Scan(0, func(a, v int) int { return a + v }))
	breakAfterOne(t, "DistinctBy", src().DistinctBy(func(i int) int { return i }))
	breakAfterOne(t, "Take", src().Take(3))
	breakAfterOne(t, "Drop", src().Drop(1))
	breakAfterOne(t, "TakeWhile", src().TakeWhile(func(int) bool { return true }))
	breakAfterOne(t, "DropWhile", src().DropWhile(func(i int) bool { return i < 2 }))
	breakAfterOne(t, "SortFunc", src().SortFunc(func(a, b int) int { return a - b }))
	breakAfterOne(t, "SortStableFunc", src().SortStableFunc(func(a, b int) int { return a - b }))
	breakAfterOne(t, "CompactFunc", src().CompactFunc(func(a, b int) bool { return a == b }))
	breakAfterOne(t, "Reverse", src().Reverse())
	breakAfterOne(t, "Peek", src().Peek(func(int) {}))
	breakAfterOne(t, "ZipWith", src().ZipWith(src(), func(a, b int) int { return a + b }))
	breakAfterOne(t, "ParallelMap", long().ParallelMap(func(i int) int { return i }, WithConcurrency(4)))
	breakAfterOne(t, "ParallelMap unordered",
		long().ParallelMap(func(i int) int { return i }, WithConcurrency(4), Unordered()))
	breakAfterOne(t, "ParallelFilter",
		long().ParallelFilter(func(int) bool { return true }, WithConcurrency(4)))
	breakAfterOne(t, "ParallelFilter unordered",
		long().ParallelFilter(func(int) bool { return true }, WithConcurrency(4), Unordered()))

	breakAfterOne2(t, "Zip", src().Zip(src()))
	breakAfterOne2(t, "Enumerate", src().Enumerate())
}

func TestPackageFunctionsHonourEarlyStop(t *testing.T) {
	src := func() Stream[int] { return Of(1, 2, 3, 4, 5) }

	breakAfterOne(t, "Chunk", Chunk(src(), 2))
	breakAfterOne(t, "Window", Window(src(), 2))
	breakAfterOne(t, "Flatten", Flatten(Of(src(), src())))
	breakAfterOne(t, "Concat", Concat(src(), src()))
	breakAfterOne(t, "Interleave", Interleave(src(), src()))
	breakAfterOne(t, "Merge", Merge(func(a, b int) int { return a - b }, src(), src()))
	breakAfterOne(t, "Cycle", Cycle(src()))
	breakAfterOne(t, "Sort", Sort(src()))
	breakAfterOne(t, "Compact", Compact(src()))
	breakAfterOne(t, "Distinct", Distinct(src()))
	breakAfterOne(t, "Repeat", Repeat(1, 5))
	breakAfterOne(t, "Range", Range(0, 5))
	breakAfterOne(t, "Iterate", Iterate(1, func(i int) int { return i + 1 }))
	breakAfterOne(t, "Generate", Generate(func() int { return 1 }))

	ch := make(chan int, 5)
	for i := range 5 {
		ch <- i
	}
	close(ch)
	breakAfterOne(t, "Chan", Chan(ch))
}

func TestStream2OpsHonourEarlyStop(t *testing.T) {
	src := func() Stream2[int, int] { return Of(1, 2, 3, 4, 5).Enumerate() }

	breakAfterOne2(t, "Stream2.Filter", src().Filter(func(int, int) bool { return true }))
	breakAfterOne2(t, "Stream2.MapKeys", src().MapKeys(func(k int) int { return k }))
	breakAfterOne2(t, "Stream2.MapValues", src().MapValues(func(v int) int { return v }))
	breakAfterOne2(t, "Stream2.Swap", src().Swap())
	breakAfterOne2(t, "Stream2.Take", src().Take(3))
	breakAfterOne2(t, "Stream2.Drop", src().Drop(1))
	breakAfterOne2(t, "Pairs", Pairs(map[int]int{1: 1, 2: 2, 3: 3}))
	breakAfterOne(t, "Stream2.Keys", src().Keys())
	breakAfterOne(t, "Stream2.Values", src().Values())
	breakAfterOne(t, "Stream2.Collapse", src().Collapse(func(_, v int) int { return v }))
}

func TestTryMapHonoursEarlyStop(t *testing.T) {
	n := 0
	if !assert.NotPanics(t, func() {
		for range TryMap(Of(1, 2, 3), func(i int) (int, error) { return i, nil }) {
			n++
			break
		}
	}, "TryMap yielded after the consumer stopped") {
		return
	}
	assert.Equal(t, 1, n, "elements TryMap consumed before the break")
}
