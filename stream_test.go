package streams

import (
	"context"
	"iter"
	"maps"
	"slices"
	"strings"
	"testing"
	"time"
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

func eq[T comparable](t *testing.T, name string, got, want []T) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Errorf("%s = %v, want %v", name, got, want)
	}
}

// --- constructors ---

func TestConstructors(t *testing.T) {
	eq(t, "Of", Of(1, 2, 3).Collect(), []int{1, 2, 3})
	eq(t, "Of(slice...)", Of([]int{4, 5}...).Collect(), []int{4, 5})
	eq(t, "Empty", Empty[int]().Collect(), nil)
	eq(t, "Range", Range(2, 5).Collect(), []int{2, 3, 4})
	eq(t, "Range empty", Range(5, 2).Collect(), nil)
	eq(t, "Repeat", Repeat("x", 3).Collect(), []string{"x", "x", "x"})
	eq(t, "Repeat 0", Repeat("x", 0).Collect(), nil)
	eq(t, "Repeat infinite", Repeat("x", -1).Take(2).Collect(), []string{"x", "x"})
	eq(t, "Iterate", Iterate(1, func(i int) int { return i * 2 }).Take(4).Collect(), []int{1, 2, 4, 8})

	n := 0
	eq(t, "Generate", Generate(func() int { n++; return n }).Take(3).Collect(), []int{1, 2, 3})

	ch := make(chan int, 3)
	ch <- 7
	ch <- 8
	close(ch)
	eq(t, "Chan", Chan(ch).Collect(), []int{7, 8})
}

func TestFromInfersElementType(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	// From exists for inference: Stream[string](maps.Keys(m)) would need the type spelled out.
	got := From(maps.Keys(m)).SortFunc(strings.Compare).Collect()
	eq(t, "From", got, []string{"a", "b"})
}

// --- intermediate operations ---

func TestIntermediateOps(t *testing.T) {
	s := func() Stream[int] { return Of(1, 2, 3, 4, 5) }

	eq(t, "Filter", s().Filter(func(i int) bool { return i%2 == 1 }).Collect(), []int{1, 3, 5})
	eq(t, "Map", s().Map(func(i int) int { return i * 2 }).Collect(), []int{2, 4, 6, 8, 10})
	eq(t, "Take", s().Take(2).Collect(), []int{1, 2})
	eq(t, "Take 0", s().Take(0).Collect(), nil)
	eq(t, "Take over", s().Take(99).Collect(), []int{1, 2, 3, 4, 5})
	eq(t, "Drop", s().Drop(3).Collect(), []int{4, 5})
	eq(t, "Drop over", s().Drop(99).Collect(), nil)
	eq(t, "TakeWhile", s().TakeWhile(func(i int) bool { return i < 3 }).Collect(), []int{1, 2})
	eq(t, "DropWhile", s().DropWhile(func(i int) bool { return i < 3 }).Collect(), []int{3, 4, 5})
	eq(t, "Reverse", s().Reverse().Collect(), []int{5, 4, 3, 2, 1})
	eq(t, "Scan", s().Scan(0, func(a, v int) int { return a + v }).Collect(), []int{1, 3, 6, 10, 15})

	// DropWhile must not resume dropping after the predicate first fails.
	eq(t, "DropWhile resumes", Of(1, 2, 1, 5).DropWhile(func(i int) bool { return i < 2 }).Collect(),
		[]int{2, 1, 5})

	seen := []int{}
	out := s().Peek(func(i int) { seen = append(seen, i) }).Take(2).Collect()
	eq(t, "Peek passthrough", out, []int{1, 2})
	eq(t, "Peek observed", seen, []int{1, 2})
}

func TestMapChangesElementType(t *testing.T) {
	got := Of(users...).
		Filter(func(u user) bool { return u.Age > 40 }).
		Map(func(u user) string { return u.Name }).
		Map(strings.ToUpper).
		Map(func(s string) int { return len(s) }).
		Collect()
	eq(t, "three type changes", got, []int{5, 3, 3, 5})
}

func TestFlatMap(t *testing.T) {
	got := Of("a b", "c").FlatMap(func(s string) Stream[string] {
		return Of(strings.Fields(s)...)
	}).Collect()
	eq(t, "FlatMap", got, []string{"a", "b", "c"})
	eq(t, "FlatMap empty", Of("").FlatMap(func(string) Stream[string] {
		return Empty[string]()
	}).Collect(), nil)
}

func TestSortAndCompact(t *testing.T) {
	byLen := func(a, b string) int { return len(a) - len(b) }
	eq(t, "SortFunc", Of("ccc", "a", "bb").SortFunc(byLen).Collect(), []string{"a", "bb", "ccc"})
	// stability: equal-length elements keep their input order
	eq(t, "SortStableFunc", Of("bb", "aa", "c").SortStableFunc(byLen).Collect(),
		[]string{"c", "bb", "aa"})
	// CompactFunc removes only adjacent duplicates, matching slices.CompactFunc
	eq(t, "CompactFunc", Of(1, 1, 2, 2, 1).CompactFunc(func(a, b int) bool { return a == b }).Collect(),
		[]int{1, 2, 1})
	eq(t, "DistinctBy", Of(1, 1, 2, 2, 1).DistinctBy(func(i int) int { return i }).Collect(),
		[]int{1, 2})
	eq(t, "DistinctBy key", Of("aa", "ab", "b").DistinctBy(func(s string) byte { return s[0] }).Collect(),
		[]string{"aa", "b"})
}

func TestZipAndEnumerate(t *testing.T) {
	// Zip yields a Stream2, so the package needs no Pair type.
	m := maps.Collect(iter.Seq2[string, int](Of("a", "b", "c").Zip(Of(1, 2, 3))))
	if m["b"] != 2 || len(m) != 3 {
		t.Errorf("Zip = %v", m)
	}
	// zipping stops at the shorter side
	eq(t, "Zip short", Of(1, 2, 3).Zip(Of("x")).Keys().Collect(), []int{1})
	eq(t, "ZipWith", Of(1, 2, 3).ZipWith(Of(10, 20, 30), func(a, b int) int { return a + b }).Collect(),
		[]int{11, 22, 33})
	eq(t, "Enumerate", Of("x", "y").Enumerate().Keys().Collect(), []int{0, 1})
}

// --- terminal operations ---

func TestTerminalOps(t *testing.T) {
	s := func() Stream[int] { return Of(3, 1, 4, 1, 5) }

	eq(t, "Collect", s().Collect(), []int{3, 1, 4, 1, 5})
	if got := s().Count(); got != 5 {
		t.Errorf("Count = %d", got)
	}
	if got := s().Fold(0, func(a, v int) int { return a + v }); got != 14 {
		t.Errorf("Fold = %d", got)
	}
	// Fold accumulates into a different type
	if got := s().Fold("", func(a string, _ int) string { return a + "." }); got != "....." {
		t.Errorf("Fold to string = %q", got)
	}
	if got, ok := s().Reduce(func(a, b int) int { return a + b }); !ok || got != 14 {
		t.Errorf("Reduce = %d, %v", got, ok)
	}
	if _, ok := Empty[int]().Reduce(func(a, _ int) int { return a }); ok {
		t.Error("Reduce on empty must report false")
	}

	if v, ok := s().First(); !ok || v != 3 {
		t.Errorf("First = %d, %v", v, ok)
	}
	if v, ok := s().Last(); !ok || v != 5 {
		t.Errorf("Last = %d, %v", v, ok)
	}
	if v, ok := s().Find(func(i int) bool { return i > 3 }); !ok || v != 4 {
		t.Errorf("Find = %d, %v", v, ok)
	}
	if _, ok := s().Find(func(i int) bool { return i > 99 }); ok {
		t.Error("Find must report false when nothing matches")
	}
	if !s().Any(func(i int) bool { return i == 4 }) {
		t.Error("Any")
	}
	if s().All(func(i int) bool { return i > 2 }) {
		t.Error("All must be false")
	}
	if !Empty[int]().All(func(_ int) bool { return false }) {
		t.Error("All must be true for an empty Stream")
	}

	cmpInt := func(a, b int) int { return a - b }
	if v, ok := s().MinFunc(cmpInt); !ok || v != 1 {
		t.Errorf("MinFunc = %d", v)
	}
	if v, ok := s().MaxFunc(cmpInt); !ok || v != 5 {
		t.Errorf("MaxFunc = %d", v)
	}
	if _, ok := Empty[int]().MaxFunc(cmpInt); ok {
		t.Error("MaxFunc on empty must report false")
	}
}

func TestEmptyReportsCommaOkNotOptional(t *testing.T) {
	if _, ok := Empty[int]().First(); ok {
		t.Error("First")
	}
	if _, ok := Empty[int]().Last(); ok {
		t.Error("Last")
	}
}

func TestGroupingTerminals(t *testing.T) {
	byCity := Of(users...).GroupBy(func(u user) string { return u.City })
	if len(byCity["NYC"]) != 3 || len(byCity["London"]) != 1 {
		t.Errorf("GroupBy = %v", byCity)
	}
	// GroupBy keeps encounter order within a group
	names := Of(users...).GroupBy(func(u user) string { return u.City })["NYC"]
	eq(t, "GroupBy order", []string{names[0].Name, names[1].Name, names[2].Name},
		[]string{"Rob", "Ken", "Grace"})

	idx := Of(users...).IndexBy(func(u user) string { return u.City })
	if idx["NYC"].Name != "Grace" {
		t.Errorf("IndexBy must keep the last match, got %v", idx["NYC"])
	}
	ages := Of(users...).ToMap(func(u user) (string, int) { return u.Name, u.Age })
	if ages["Ada"] != 36 || len(ages) != 5 {
		t.Errorf("ToMap = %v", ages)
	}
	yes, no := Of(1, 2, 3, 4).Partition(func(i int) bool { return i%2 == 0 })
	eq(t, "Partition yes", yes, []int{2, 4})
	eq(t, "Partition no", no, []int{1, 3})
}

// --- laziness ---

func TestLazinessAndShortCircuit(t *testing.T) {
	count := func(fn func(s Stream[int])) int {
		n := 0
		src := Range(0, 1000).Peek(func(int) { n++ })
		fn(src)
		return n
	}

	if got := count(func(s Stream[int]) { s.Take(3).Collect() }); got != 3 {
		t.Errorf("Take consumed %d elements, want 3", got)
	}
	if got := count(func(s Stream[int]) { s.First() }); got != 1 {
		t.Errorf("First consumed %d elements, want 1", got)
	}
	if got := count(func(s Stream[int]) { s.Any(func(i int) bool { return i == 2 }) }); got != 3 {
		t.Errorf("Any consumed %d elements, want 3", got)
	}
	if got := count(func(s Stream[int]) { s.All(func(i int) bool { return i < 2 }) }); got != 3 {
		t.Errorf("All consumed %d elements, want 3", got)
	}
	if got := count(func(s Stream[int]) { s.TakeWhile(func(i int) bool { return i < 5 }).Collect() }); got != 6 {
		t.Errorf("TakeWhile consumed %d elements, want 6", got)
	}
	if got := count(func(s Stream[int]) {
		s.Map(func(i int) int { return i }).Filter(func(i int) bool { return i > 1 }).First()
	}); got != 3 {
		t.Errorf("Map+Filter+First consumed %d elements, want 3", got)
	}
	// an infinite source must stay usable behind a bounded operation
	if got := Iterate(1, func(i int) int { return i + 1 }).Take(4).Collect(); len(got) != 4 {
		t.Errorf("infinite source = %v", got)
	}
}

// --- standard library interoperation ---

func TestStdlibInterop(t *testing.T) {
	s := Of(3, 1, 2)
	// a Stream is an iter.Seq: one conversion, no wrapper call
	eq(t, "slices.Sorted", slices.Sorted(iter.Seq[int](s)), []int{1, 2, 3})
	// and the reverse
	eq(t, "from slices.Values", From(slices.Values([]int{7, 8})).Collect(), []int{7, 8})
	// range directly over a Stream
	total := 0
	for v := range s {
		total += v
	}
	if total != 6 {
		t.Errorf("range over Stream = %d", total)
	}
	// range over a Stream2
	keys := 0
	for k := range Of("a", "b").Enumerate() {
		keys += k
	}
	if keys != 1 {
		t.Errorf("range over Stream2 = %d", keys)
	}
}

func TestChanContext(t *testing.T) {
	t.Run("ends when the channel closes", func(t *testing.T) {
		ch := make(chan int, 3)
		ch <- 1
		ch <- 2
		close(ch)
		eq(t, "values", ChanContext(context.Background(), ch).Collect(), []int{1, 2})
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
			eq(t, "values before cancel", got, []int{1})
		case <-time.After(2 * time.Second):
			t.Fatal("ChanContext did not end within 2s of cancellation")
		}
	})

	t.Run("honours early stop", func(t *testing.T) {
		ch := make(chan int, 3)
		ch <- 1
		ch <- 2
		close(ch)
		eq(t, "values", ChanContext(context.Background(), ch).Take(1).Collect(), []int{1})
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
		t.Errorf("goroutines stranded: before=%d after=%d", before, countGoroutines())
	})
}
