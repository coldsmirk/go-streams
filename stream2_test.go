package streams

import (
	"iter"
	"maps"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStream2Ops(t *testing.T) {
	s := func() Stream2[string, int] {
		return Of("a", "b", "c").Zip(Of(1, 2, 3)).Swap().Swap()
	}

	assert.Equal(t, []string{"a", "b", "c"}, s().Keys().Collect(), "Keys")
	assert.Equal(t, []int{1, 2, 3}, s().Values().Collect(), "Values")
	assert.Equal(t, []string{"b", "c"},
		s().Filter(func(_ string, v int) bool { return v > 1 }).Keys().Collect(), "Filter")
	assert.Equal(t, []int{1, 1, 1},
		s().MapKeys(func(k string) int { return len(k) }).Keys().Collect(), "MapKeys")
	assert.Equal(t, []string{"1", "2", "3"},
		s().MapValues(strconv.Itoa).Values().Collect(), "MapValues")
	assert.Equal(t, []string{"a1", "b2", "c3"},
		s().Collapse(func(k string, v int) string { return k + strconv.Itoa(v) }).Collect(), "Collapse")
	assert.Equal(t, []int{1, 2, 3}, s().Swap().Keys().Collect(), "Swap")
	assert.Equal(t, []string{"a", "b"}, s().Take(2).Keys().Collect(), "Take")
	assert.Empty(t, s().Take(0).Keys().Collect(), "Take 0")
	assert.Equal(t, []string{"b", "c"}, s().Drop(1).Keys().Collect(), "Drop")
	assert.Empty(t, s().Drop(5).Keys().Collect(), "Drop all")

	assert.Equal(t, 3, s().Count(), "Count")
	assert.Equal(t, 6, s().Fold(0, func(a int, _ string, v int) int { return a + v }), "Fold")

	visited := 0
	s().ForEach(func(string, int) { visited++ })
	assert.Equal(t, 3, visited, "pairs ForEach visited")

	k, v, ok := s().First()
	assert.True(t, ok, "First")
	assert.Equal(t, "a", k, "First key")
	assert.Equal(t, 1, v, "First value")

	k, v, ok = s().Last()
	assert.True(t, ok, "Last")
	assert.Equal(t, "c", k, "Last key")
	assert.Equal(t, 3, v, "Last value")

	k, v, ok = Of("a").Zip(Of(1)).Last()
	assert.True(t, ok, "Last on a single pair")
	assert.Equal(t, "a", k, "Last key on a single pair")
	assert.Equal(t, 1, v, "Last value on a single pair")

	k, v, ok = s().Find(func(_ string, v int) bool { return v > 1 })
	assert.True(t, ok, "Find")
	assert.Equal(t, "b", k, "Find key")
	assert.Equal(t, 2, v, "Find value")

	k, v, ok = s().Find(func(string, int) bool { return false })
	assert.False(t, ok, "Find with no match")
	assert.Zero(t, k, "Find with no match returns the zero key")
	assert.Zero(t, v, "Find with no match returns the zero value")

	assert.True(t, s().Any(func(_ string, v int) bool { return v == 3 }), "Any with a match")
	assert.False(t, s().Any(func(string, int) bool { return false }), "Any with no match")
	assert.True(t, s().All(func(_ string, v int) bool { return v > 0 }), "All with every pair passing")
	assert.False(t, s().All(func(_ string, v int) bool { return v > 1 }), "All with a failing pair")

	assert.Empty(t, Empty2[string, int]().Keys().Collect(), "Empty2")
}

func TestEmpty2ReportsCommaOkNotOptional(t *testing.T) {
	k, v, ok := Empty2[string, int]().First()
	assert.False(t, ok, "First")
	assert.Zero(t, k, "First key")
	assert.Zero(t, v, "First value")

	k, v, ok = Empty2[string, int]().Last()
	assert.False(t, ok, "Last")
	assert.Zero(t, k, "Last key")
	assert.Zero(t, v, "Last value")

	k, v, ok = Empty2[string, int]().Find(func(string, int) bool { return true })
	assert.False(t, ok, "Find")
	assert.Zero(t, k, "Find key")
	assert.Zero(t, v, "Find value")

	assert.False(t, Empty2[string, int]().Any(func(string, int) bool { return true }),
		"Any on an empty Stream2")
	assert.True(t, Empty2[string, int]().All(func(string, int) bool { return false }),
		"All on an empty Stream2")
}

func TestPairsAndMapInterop(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	// Pairs accepts any map type whose underlying type is map[K]V
	type scores map[string]int
	assert.Equal(t, 3, Pairs(scores(m)).Count(), "Pairs on a defined map type")

	// round-trip through the standard library
	doubled := Pairs(m).MapValues(func(v int) int { return v * 2 })
	back := maps.Collect(iter.Seq2[string, int](doubled))
	assert.Equal(t, map[string]int{"a": 2, "b": 4, "c": 6}, back, "round trip")

	// CollectMap performs the same conversion without naming the types
	assert.Equal(t, map[string]int{"a": 2, "b": 4, "c": 6},
		CollectMap(Pairs(m).MapValues(func(v int) int { return v * 2 })), "CollectMap")

	// From2 infers K and V
	assert.Equal(t, 3, From2(maps.All(m)).Count(), "From2")
}

func TestStream2ShortCircuits(t *testing.T) {
	count := func(fn func(s Stream2[int, int])) int {
		n := 0
		src := Range(0, 1000).Peek(func(int) { n++ }).Enumerate()
		fn(src)
		return n
	}

	assert.Equal(t, 3, count(func(s Stream2[int, int]) {
		s.Take(3).ForEach(func(int, int) {})
	}), "pairs Take consumed")
	assert.Equal(t, 1, count(func(s Stream2[int, int]) {
		s.Collapse(func(_, v int) int { return v }).First()
	}), "pairs Collapse+First consumed")
	assert.Equal(t, 1, count(func(s Stream2[int, int]) { s.First() }), "pairs First consumed")
	assert.Equal(t, 3, count(func(s Stream2[int, int]) {
		s.Find(func(_, v int) bool { return v == 2 })
	}), "pairs Find consumed")
	assert.Equal(t, 3, count(func(s Stream2[int, int]) {
		s.Any(func(_, v int) bool { return v == 2 })
	}), "pairs Any consumed")
	assert.Equal(t, 3, count(func(s Stream2[int, int]) {
		s.All(func(_, v int) bool { return v < 2 })
	}), "pairs All consumed")

	// an infinite source must stay usable behind a short-circuiting terminal
	inf := Iterate(1, func(v int) int { return v + 1 }).Enumerate()
	k, v, ok := inf.Find(func(_, v int) bool { return v == 5 })
	assert.True(t, ok, "Find on an infinite source")
	assert.Equal(t, 4, k, "Find key on an infinite source")
	assert.Equal(t, 5, v, "Find value on an infinite source")
}
