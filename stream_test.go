package streams

import (
	"context"
	"iter"
	"maps"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type user struct {
	Name string
	Age  int
	City string
}

var users = []user{
	{"Ada", 36, "London"},
	{"Linus", 54, "Portland"},
	{"Rob", 60, "NYC"},
	{"Ken", 81, "NYC"},
	{"Grace", 45, "NYC"},
}

// --- constructors ---

func TestConstructors(t *testing.T) {
	assert.Equal(t, []int{1, 2, 3}, Of(1, 2, 3).Collect(), "Of")
	assert.Equal(t, []int{4, 5}, Of([]int{4, 5}...).Collect(), "Of(slice...)")
	assert.Empty(t, Empty[int]().Collect(), "Empty")
	assert.Equal(t, []int{2, 3, 4}, Range(2, 5).Collect(), "Range")
	assert.Empty(t, Range(5, 2).Collect(), "Range empty")
	assert.Equal(t, []string{"x", "x", "x"}, Repeat("x", 3).Collect(), "Repeat")
	assert.Empty(t, Repeat("x", 0).Collect(), "Repeat 0")
	assert.Equal(t, []string{"x", "x"}, Repeat("x", -1).Take(2).Collect(), "Repeat infinite")
	assert.Equal(t, []int{1, 2, 4, 8},
		Iterate(1, func(i int) int { return i * 2 }).Take(4).Collect(), "Iterate")

	n := 0
	assert.Equal(t, []int{1, 2, 3},
		Generate(func() int { n++; return n }).Take(3).Collect(), "Generate")

	ch := make(chan int, 3)
	ch <- 7
	ch <- 8
	close(ch)
	assert.Equal(t, []int{7, 8}, Chan(ch).Collect(), "Chan")
}

func TestFromInfersElementType(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	// From exists for inference: Stream[string](maps.Keys(m)) would need the type spelled out.
	got := From(maps.Keys(m)).SortFunc(strings.Compare).Collect()
	assert.Equal(t, []string{"a", "b"}, got, "From")
}

// --- intermediate operations ---

func TestIntermediateOps(t *testing.T) {
	s := func() Stream[int] { return Of(1, 2, 3, 4, 5) }

	assert.Equal(t, []int{1, 3, 5}, s().Filter(func(i int) bool { return i%2 == 1 }).Collect(), "Filter")
	assert.Equal(t, []int{2, 4, 6, 8, 10}, s().Map(func(i int) int { return i * 2 }).Collect(), "Map")
	assert.Equal(t, []int{1, 2}, s().Take(2).Collect(), "Take")
	assert.Empty(t, s().Take(0).Collect(), "Take 0")
	assert.Equal(t, []int{1, 2, 3, 4, 5}, s().Take(99).Collect(), "Take over")
	assert.Equal(t, []int{4, 5}, s().Drop(3).Collect(), "Drop")
	assert.Empty(t, s().Drop(99).Collect(), "Drop over")
	assert.Equal(t, []int{1, 2}, s().TakeWhile(func(i int) bool { return i < 3 }).Collect(), "TakeWhile")
	assert.Equal(t, []int{3, 4, 5}, s().DropWhile(func(i int) bool { return i < 3 }).Collect(), "DropWhile")
	assert.Equal(t, []int{5, 4, 3, 2, 1}, s().Reverse().Collect(), "Reverse")
	assert.Equal(t, []int{1, 3, 6, 10, 15}, s().Scan(0, func(a, v int) int { return a + v }).Collect(), "Scan")

	// DropWhile must not resume dropping after the predicate first fails.
	assert.Equal(t, []int{2, 1, 5},
		Of(1, 2, 1, 5).DropWhile(func(i int) bool { return i < 2 }).Collect(), "DropWhile resumes")

	seen := []int{}
	out := s().Peek(func(i int) { seen = append(seen, i) }).Take(2).Collect()
	assert.Equal(t, []int{1, 2}, out, "Peek passthrough")
	assert.Equal(t, []int{1, 2}, seen, "Peek observed")
}

func TestMapChangesElementType(t *testing.T) {
	got := Of(users...).
		Filter(func(u user) bool { return u.Age > 40 }).
		Map(func(u user) string { return u.Name }).
		Map(strings.ToUpper).
		Map(func(s string) int { return len(s) }).
		Collect()
	assert.Equal(t, []int{5, 3, 3, 5}, got, "three type changes")
}

func TestFlatMap(t *testing.T) {
	got := Of("a b", "c").FlatMap(func(s string) Stream[string] {
		return Of(strings.Fields(s)...)
	}).Collect()
	assert.Equal(t, []string{"a", "b", "c"}, got, "FlatMap")
	assert.Empty(t, Of("").FlatMap(func(string) Stream[string] {
		return Empty[string]()
	}).Collect(), "FlatMap empty")
}

func TestSortAndCompact(t *testing.T) {
	byLen := func(a, b string) int { return len(a) - len(b) }
	assert.Equal(t, []string{"a", "bb", "ccc"}, Of("ccc", "a", "bb").SortFunc(byLen).Collect(), "SortFunc")
	// stability: equal-length elements keep their input order
	assert.Equal(t, []string{"c", "bb", "aa"},
		Of("bb", "aa", "c").SortStableFunc(byLen).Collect(), "SortStableFunc")
	// CompactFunc removes only adjacent duplicates, matching slices.CompactFunc
	assert.Equal(t, []int{1, 2, 1},
		Of(1, 1, 2, 2, 1).CompactFunc(func(a, b int) bool { return a == b }).Collect(), "CompactFunc")
	assert.Equal(t, []int{1, 2},
		Of(1, 1, 2, 2, 1).DistinctBy(func(i int) int { return i }).Collect(), "DistinctBy")
	assert.Equal(t, []string{"aa", "b"},
		Of("aa", "ab", "b").DistinctBy(func(s string) byte { return s[0] }).Collect(), "DistinctBy key")
}

func TestZipAndEnumerate(t *testing.T) {
	// Zip yields a Stream2, so the package needs no Pair type.
	m := maps.Collect(iter.Seq2[string, int](Of("a", "b", "c").Zip(Of(1, 2, 3))))
	assert.Equal(t, map[string]int{"a": 1, "b": 2, "c": 3}, m, "Zip")
	// zipping stops at the shorter side
	assert.Equal(t, []int{1}, Of(1, 2, 3).Zip(Of("x")).Keys().Collect(), "Zip short")
	assert.Equal(t, []int{11, 22, 33},
		Of(1, 2, 3).ZipWith(Of(10, 20, 30), func(a, b int) int { return a + b }).Collect(), "ZipWith")
	assert.Equal(t, []int{0, 1}, Of("x", "y").Enumerate().Keys().Collect(), "Enumerate")
}

// --- terminal operations ---

func TestTerminalOps(t *testing.T) {
	s := func() Stream[int] { return Of(3, 1, 4, 1, 5) }

	assert.Equal(t, []int{3, 1, 4, 1, 5}, s().Collect(), "Collect")
	assert.Equal(t, 5, s().Count(), "Count")
	assert.Equal(t, 14, s().Fold(0, func(a, v int) int { return a + v }), "Fold")
	// Fold accumulates into a different type
	assert.Equal(t, ".....", s().Fold("", func(a string, _ int) string { return a + "." }), "Fold to string")

	sum, ok := s().Reduce(func(a, b int) int { return a + b })
	assert.True(t, ok, "Reduce")
	assert.Equal(t, 14, sum, "Reduce")
	_, ok = Empty[int]().Reduce(func(a, _ int) int { return a })
	assert.False(t, ok, "Reduce on empty must report false")

	v, ok := s().First()
	assert.True(t, ok, "First")
	assert.Equal(t, 3, v, "First")

	v, ok = s().Last()
	assert.True(t, ok, "Last")
	assert.Equal(t, 5, v, "Last")

	v, ok = s().Find(func(i int) bool { return i > 3 })
	assert.True(t, ok, "Find")
	assert.Equal(t, 4, v, "Find")

	_, ok = s().Find(func(i int) bool { return i > 99 })
	assert.False(t, ok, "Find must report false when nothing matches")

	assert.True(t, s().Any(func(i int) bool { return i == 4 }), "Any")
	assert.False(t, s().All(func(i int) bool { return i > 2 }), "All must be false")
	assert.True(t, Empty[int]().All(func(_ int) bool { return false }),
		"All must be true for an empty Stream")

	cmpInt := func(a, b int) int { return a - b }
	v, ok = s().MinFunc(cmpInt)
	assert.True(t, ok, "MinFunc")
	assert.Equal(t, 1, v, "MinFunc")

	v, ok = s().MaxFunc(cmpInt)
	assert.True(t, ok, "MaxFunc")
	assert.Equal(t, 5, v, "MaxFunc")

	_, ok = Empty[int]().MaxFunc(cmpInt)
	assert.False(t, ok, "MaxFunc on empty must report false")
}

func TestEmptyReportsCommaOkNotOptional(t *testing.T) {
	_, ok := Empty[int]().First()
	assert.False(t, ok, "First")
	_, ok = Empty[int]().Last()
	assert.False(t, ok, "Last")
}

func TestGroupingTerminals(t *testing.T) {
	byCity := Of(users...).GroupBy(func(u user) string { return u.City })
	assert.Len(t, byCity["NYC"], 3, "GroupBy NYC")
	assert.Len(t, byCity["London"], 1, "GroupBy London")
	// GroupBy keeps encounter order within a group
	names := Of(users...).GroupBy(func(u user) string { return u.City })["NYC"]
	assert.Equal(t, []string{"Rob", "Ken", "Grace"},
		[]string{names[0].Name, names[1].Name, names[2].Name}, "GroupBy order")

	idx := Of(users...).IndexBy(func(u user) string { return u.City })
	assert.Equal(t, "Grace", idx["NYC"].Name, "IndexBy must keep the last match")

	ages := Of(users...).ToMap(func(u user) (string, int) { return u.Name, u.Age })
	assert.Len(t, ages, 5, "ToMap")
	assert.Equal(t, 36, ages["Ada"], "ToMap")

	yes, no := Of(1, 2, 3, 4).Partition(func(i int) bool { return i%2 == 0 })
	assert.Equal(t, []int{2, 4}, yes, "Partition yes")
	assert.Equal(t, []int{1, 3}, no, "Partition no")

	yes, no = Empty[int]().Partition(func(int) bool { return true })
	assert.Empty(t, yes, "Partition yes over an empty Stream")
	assert.Empty(t, no, "Partition no over an empty Stream")
}

// --- laziness ---

func TestLazinessAndShortCircuit(t *testing.T) {
	count := func(fn func(s Stream[int])) int {
		n := 0
		src := Range(0, 1000).Peek(func(int) { n++ })
		fn(src)
		return n
	}

	assert.Equal(t, 3, count(func(s Stream[int]) { s.Take(3).Collect() }), "elements Take consumed")
	assert.Equal(t, 1, count(func(s Stream[int]) { s.First() }), "elements First consumed")
	assert.Equal(t, 3, count(func(s Stream[int]) {
		s.Any(func(i int) bool { return i == 2 })
	}), "elements Any consumed")
	assert.Equal(t, 3, count(func(s Stream[int]) {
		s.All(func(i int) bool { return i < 2 })
	}), "elements All consumed")
	assert.Equal(t, 6, count(func(s Stream[int]) {
		s.TakeWhile(func(i int) bool { return i < 5 }).Collect()
	}), "elements TakeWhile consumed")
	assert.Equal(t, 3, count(func(s Stream[int]) {
		s.Map(func(i int) int { return i }).Filter(func(i int) bool { return i > 1 }).First()
	}), "elements Map+Filter+First consumed")

	// an infinite source must stay usable behind a bounded operation
	assert.Len(t, Iterate(1, func(i int) int { return i + 1 }).Take(4).Collect(), 4, "infinite source")
}

// --- standard library interoperation ---

func TestStdlibInterop(t *testing.T) {
	s := Of(3, 1, 2)
	// a Stream is an iter.Seq: one conversion, no wrapper call
	assert.Equal(t, []int{1, 2, 3}, slices.Sorted(iter.Seq[int](s)), "slices.Sorted")
	// and the reverse
	assert.Equal(t, []int{7, 8}, From(slices.Values([]int{7, 8})).Collect(), "from slices.Values")
	// range directly over a Stream
	total := 0
	for v := range s {
		total += v
	}
	assert.Equal(t, 6, total, "range over Stream")
	// range over a Stream2
	keys := 0
	for k := range Of("a", "b").Enumerate() {
		keys += k
	}
	assert.Equal(t, 1, keys, "range over Stream2")
}

func TestChanContext(t *testing.T) {
	t.Run("ends when the channel closes", func(t *testing.T) {
		ch := make(chan int, 3)
		ch <- 1
		ch <- 2
		close(ch)
		assert.Equal(t, []int{1, 2}, ChanContext(context.Background(), ch).Collect())
	})

	t.Run("ends when the context is done", func(t *testing.T) {
		ch := make(chan int, 1)
		ch <- 1
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan []int, 1)
		go func() { done <- ChanContext(ctx, ch).Collect() }()
		time.Sleep(30 * time.Millisecond)
		cancel()
		select {
		case got := <-done:
			assert.Equal(t, []int{1}, got, "values before cancel")
		case <-time.After(2 * time.Second):
			require.Fail(t, "ChanContext did not end within 2s of cancellation")
		}
	})

	t.Run("emits nothing once the context is done", func(t *testing.T) {
		// A done context and a buffered value are both live select cases, and
		// the tie between them is broken at random, so a single pass could get
		// lucky. Many passes make a regression practically certain to surface.
		for range 100 {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			ch := make(chan int, 3)
			ch <- 1
			ch <- 2
			ch <- 3
			assert.Empty(t, ChanContext(ctx, ch).Collect(), "values after cancel")
		}
	})

	t.Run("honours early stop", func(t *testing.T) {
		ch := make(chan int, 3)
		ch <- 1
		ch <- 2
		close(ch)
		assert.Equal(t, []int{1}, ChanContext(context.Background(), ch).Take(1).Collect())
	})

	// The reason ChanContext exists: a goroutine parked in it is reclaimable,
	// which a goroutine parked in Chan over a quiet channel is not.
	t.Run("a parked reader is released by cancelling", func(t *testing.T) {
		before := countGoroutines()
		for range 30 {
			ch := make(chan int) // never written to, never closed
			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				for range ChanContext(ctx, ch) {
				}
			}()
			time.Sleep(time.Millisecond)
			cancel()
		}
		deadline := time.Now().Add(3 * time.Second)
		for time.Now().Before(deadline) {
			if countGoroutines() <= before+2 {
				return
			}
			time.Sleep(20 * time.Millisecond)
		}
		// Fail outright rather than re-testing the count: the loop above has
		// already established it, and a fresh reading could flip it to a pass.
		assert.Failf(t, "goroutines stranded", "before=%d after=%d", before, countGoroutines())
	})
}
