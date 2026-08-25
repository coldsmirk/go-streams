package streams

import (
	"iter"
	"maps"
	"strconv"
	"testing"
)

func TestStream2Ops(t *testing.T) {
	s := func() Stream2[string, int] {
		return Of("a", "b", "c").Zip(Of(1, 2, 3)).Swap().Swap()
	}

	eq(t, "Keys", s().Keys().Collect(), []string{"a", "b", "c"})
	eq(t, "Values", s().Values().Collect(), []int{1, 2, 3})
	eq(t, "Filter", s().Filter(func(_ string, v int) bool { return v > 1 }).Keys().Collect(),
		[]string{"b", "c"})
	eq(t, "MapKeys", s().MapKeys(func(k string) int { return len(k) }).Keys().Collect(),
		[]int{1, 1, 1})
	eq(t, "MapValues", s().MapValues(strconv.Itoa).Values().Collect(),
		[]string{"1", "2", "3"})
	eq(t, "Collapse", s().Collapse(func(k string, v int) string { return k + strconv.Itoa(v) }).Collect(),
		[]string{"a1", "b2", "c3"})
	eq(t, "Swap", s().Swap().Keys().Collect(), []int{1, 2, 3})
	eq(t, "Take", s().Take(2).Keys().Collect(), []string{"a", "b"})
	eq(t, "Take 0", s().Take(0).Keys().Collect(), nil)
	eq(t, "Drop", s().Drop(1).Keys().Collect(), []string{"b", "c"})
	eq(t, "Drop all", s().Drop(5).Keys().Collect(), nil)

	if got := s().Count(); got != 3 {
		t.Errorf("Count = %d", got)
	}
	if got := s().Fold(0, func(a int, _ string, v int) int { return a + v }); got != 6 {
		t.Errorf("Fold = %d", got)
	}
	visited := 0
	s().ForEach(func(string, int) { visited++ })
	if visited != 3 {
		t.Errorf("ForEach visited %d", visited)
	}

	if k, v, ok := s().First(); !ok || k != "a" || v != 1 {
		t.Errorf("First = %v, %v, %v", k, v, ok)
	}
	if k, v, ok := s().Last(); !ok || k != "c" || v != 3 {
		t.Errorf("Last = %v, %v, %v", k, v, ok)
	}
	if k, v, ok := Of("a").Zip(Of(1)).Last(); !ok || k != "a" || v != 1 {
		t.Errorf("Last on a single pair = %v, %v, %v", k, v, ok)
	}
	if k, v, ok := s().Find(func(_ string, v int) bool { return v > 1 }); !ok || k != "b" || v != 2 {
		t.Errorf("Find = %v, %v, %v", k, v, ok)
	}
	if k, v, ok := s().Find(func(string, int) bool { return false }); ok || k != "" || v != 0 {
		t.Errorf("Find with no match = %v, %v, %v, want the zero pair and not ok", k, v, ok)
	}
	if !s().Any(func(_ string, v int) bool { return v == 3 }) {
		t.Error("Any with a match = false")
	}
	if s().Any(func(string, int) bool { return false }) {
		t.Error("Any with no match = true")
	}
	if !s().All(func(_ string, v int) bool { return v > 0 }) {
		t.Error("All with every pair passing = false")
	}
	if s().All(func(_ string, v int) bool { return v > 1 }) {
		t.Error("All with a failing pair = true")
	}

	eq(t, "Empty2", Empty2[string, int]().Keys().Collect(), nil)
}

func TestEmpty2ReportsCommaOkNotOptional(t *testing.T) {
	if k, v, ok := Empty2[string, int]().First(); ok || k != "" || v != 0 {
		t.Errorf("First = %v, %v, %v", k, v, ok)
	}
	if k, v, ok := Empty2[string, int]().Last(); ok || k != "" || v != 0 {
		t.Errorf("Last = %v, %v, %v", k, v, ok)
	}
	if k, v, ok := Empty2[string, int]().Find(func(string, int) bool { return true }); ok || k != "" || v != 0 {
		t.Errorf("Find = %v, %v, %v", k, v, ok)
	}
	if Empty2[string, int]().Any(func(string, int) bool { return true }) {
		t.Error("Any on an empty Stream2 = true")
	}
	if !Empty2[string, int]().All(func(string, int) bool { return false }) {
		t.Error("All on an empty Stream2 = false")
	}
}

func TestPairsAndMapInterop(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	// Pairs accepts any map type whose underlying type is map[K]V
	type scores map[string]int
	if got := Pairs(scores(m)).Count(); got != 3 {
		t.Errorf("Pairs on a defined map type = %d", got)
	}
	// round-trip through the standard library
	doubled := Pairs(m).MapValues(func(v int) int { return v * 2 })
	back := maps.Collect(iter.Seq2[string, int](doubled))
	if back["b"] != 4 || len(back) != 3 {
		t.Errorf("round trip = %v", back)
	}
	// CollectMap performs the same conversion without naming the types
	if got := CollectMap(Pairs(m).MapValues(func(v int) int { return v * 2 })); got["c"] != 6 || len(got) != 3 {
		t.Errorf("CollectMap = %v", got)
	}
	// From2 infers K and V
	if got := From2(maps.All(m)).Count(); got != 3 {
		t.Errorf("From2 = %d", got)
	}
}

func TestStream2ShortCircuits(t *testing.T) {
	count := func(fn func(s Stream2[int, int])) int {
		n := 0
		src := Range(0, 1000).Peek(func(int) { n++ }).Enumerate()
		fn(src)
		return n
	}

	if got := count(func(s Stream2[int, int]) { s.Take(3).ForEach(func(int, int) {}) }); got != 3 {
		t.Errorf("Take consumed %d pairs, want 3", got)
	}
	if got := count(func(s Stream2[int, int]) { s.Collapse(func(_, v int) int { return v }).First() }); got != 1 {
		t.Errorf("Collapse+First consumed %d pairs, want 1", got)
	}
	if got := count(func(s Stream2[int, int]) { s.First() }); got != 1 {
		t.Errorf("First consumed %d pairs, want 1", got)
	}
	if got := count(func(s Stream2[int, int]) { s.Find(func(_, v int) bool { return v == 2 }) }); got != 3 {
		t.Errorf("Find consumed %d pairs, want 3", got)
	}
	if got := count(func(s Stream2[int, int]) { s.Any(func(_, v int) bool { return v == 2 }) }); got != 3 {
		t.Errorf("Any consumed %d pairs, want 3", got)
	}
	if got := count(func(s Stream2[int, int]) { s.All(func(_, v int) bool { return v < 2 }) }); got != 3 {
		t.Errorf("All consumed %d pairs, want 3", got)
	}

	// an infinite source must stay usable behind a short-circuiting terminal
	inf := Iterate(1, func(v int) int { return v + 1 }).Enumerate()
	if k, v, ok := inf.Find(func(_, v int) bool { return v == 5 }); !ok || k != 4 || v != 5 {
		t.Errorf("Find on an infinite source = %v, %v, %v", k, v, ok)
	}
}
