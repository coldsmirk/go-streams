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
	eq(t, "Empty2", Empty2[string, int]().Keys().Collect(), nil)
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
	consumed := 0
	src := Range(0, 1000).Peek(func(int) { consumed++ }).Enumerate()
	src.Take(3).ForEach(func(int, int) {})
	if consumed != 3 {
		t.Errorf("Stream2.Take consumed %d, want 3", consumed)
	}

	consumed = 0
	Range(0, 1000).Peek(func(int) { consumed++ }).Enumerate().
		Collapse(func(_, v int) int { return v }).First()
	if consumed != 1 {
		t.Errorf("Collapse+First consumed %d, want 1", consumed)
	}
}
