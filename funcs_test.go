package streams

import (
	"cmp"
	"iter"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConstrainedFunctionsPairWithFuncMethods(t *testing.T) {
	// package function: element type is constrained, like slices.Max
	maxV, ok := Max(Of(3, 9, 4))
	assert.True(t, ok, "Max")
	assert.Equal(t, 9, maxV, "Max")

	minV, ok := Min(Of(3, 9, 4))
	assert.True(t, ok, "Min")
	assert.Equal(t, 3, minV, "Min")

	_, ok = Max(Empty[int]())
	assert.False(t, ok, "Max on empty must report false")

	// method: no constraint, caller supplies the comparison
	oldest, ok := Of(users...).MaxFunc(func(a, b user) int { return cmp.Compare(a.Age, b.Age) })
	assert.True(t, ok, "MaxFunc")
	assert.Equal(t, "Ken", oldest.Name, "MaxFunc")

	assert.Equal(t, []int{1, 2, 3}, Sort(Of(3, 1, 2)).Collect(), "Sort")
	assert.Equal(t, []string{"a", "b"}, Sort(Of("b", "a")).Collect(), "Sort strings")
	assert.Equal(t, []int{1, 2, 1}, Compact(Of(1, 1, 2, 2, 1)).Collect(), "Compact")
	assert.Equal(t, []int{1, 2}, Distinct(Of(1, 1, 2, 2, 1)).Collect(), "Distinct")

	assert.True(t, Contains(Of(1, 2, 3), 2), "Contains a present element")
	assert.False(t, Contains(Of(1, 2, 3), 9), "Contains an absent element")
	assert.Equal(t, map[string]int{"a": 2, "b": 1}, Frequency(Of("a", "b", "a")), "Frequency")
}

func TestNumericAggregates(t *testing.T) {
	assert.Equal(t, 6, Sum(Of(1, 2, 3)), "Sum")
	assert.Equal(t, 0, Sum(Empty[int]()), "Sum of empty")
	assert.Equal(t, 4.0, Sum(Of(1.5, 2.5)), "Sum float")
	assert.Equal(t, 24, Product(Of(2, 3, 4)), "Product")
	assert.Equal(t, 1, Product(Empty[int]()), "Product of empty is the multiplicative identity")

	avg, ok := Average(Of(1, 2, 3, 4))
	assert.True(t, ok, "Average")
	assert.Equal(t, 2.5, avg, "Average")

	_, ok = Average(Empty[int]())
	assert.False(t, ok, "Average on empty must report false")

	// a defined type with a numeric underlying type satisfies Numeric
	type celsius float64
	assert.Equal(t, celsius(3), Sum(Of[celsius](1, 2)), "Sum of a defined type")
}

func TestRegrouping(t *testing.T) {
	assert.Equal(t, []int{2, 2, 1},
		Chunk(Of(1, 2, 3, 4, 5), 2).Map(func(c []int) int { return len(c) }).Collect(), "Chunk sizes")
	first, _ := Chunk(Of(1, 2, 3), 2).First()
	assert.Equal(t, []int{1, 2}, first, "Chunk content")
	assert.Empty(t, Chunk(Empty[int](), 2).Collect(), "Chunk of empty")
	assert.Equal(t, []int{2},
		Chunk(Of(1, 2), 2).Map(func(c []int) int { return len(c) }).Collect(), "Chunk exact")

	// Window slides one element at a time and yields nothing if the input is short
	windows := Window(Of(1, 2, 3, 4), 2).Collect()
	require.Len(t, windows, 3, "Window count")
	assert.Equal(t, []int{1, 2}, windows[0], "Window[0]")
	assert.Equal(t, []int{3, 4}, windows[2], "Window[2]")
	assert.Empty(t, Window(Of(1), 2).Collect(), "Window shorter than n")
	// each window must be an independent slice, not a view of a shared buffer
	assert.NotEqual(t, windows[0], windows[1], "Window reused its buffer across yields")

	assert.Equal(t, []int{1, 2, 3}, Flatten(Of(Of(1, 2), Of(3), Empty[int]())).Collect(), "Flatten")
}

func TestChunkAndWindowRejectNonPositive(t *testing.T) {
	for _, tc := range []struct {
		name string
		call func()
	}{
		{"Chunk", func() { Chunk(Of(1), 0) }},
		{"Window", func() { Window(Of(1), 0) }},
	} {
		assert.Panicsf(t, tc.call, "%s(n=0) must panic", tc.name)
	}
}

func TestCombiningStreams(t *testing.T) {
	assert.Equal(t, []int{1, 2, 3}, Concat(Of(1, 2), Of(3), Empty[int]()).Collect(), "Concat")
	assert.Empty(t, Concat[int]().Collect(), "Concat none")
	assert.Equal(t, []int{1, 2, 3, 4}, Interleave(Of(1, 3), Of(2, 4)).Collect(), "Interleave")
	assert.Equal(t, []int{1, 2, 4, 6}, Interleave(Of(1), Of(2, 4, 6)).Collect(), "Interleave uneven")

	byInt := func(a, b int) int { return a - b }
	assert.Equal(t, []int{1, 2, 3, 4, 5, 6, 7},
		Merge(byInt, Of(1, 4, 7), Of(2, 5), Of(3, 6)).Collect(), "Merge")
	assert.Equal(t, []int{3, 1}, Merge(byInt, Of(3, 1)).Collect(), "Merge one")
	assert.Empty(t, Merge(byInt).Collect(), "Merge none")
	assert.Equal(t, []int{1, 2}, Merge(byInt, Empty[int](), Of(1, 2)).Collect(), "Merge with empty")

	assert.Equal(t, []int{1, 2, 1, 2, 1}, Cycle(Of(1, 2)).Take(5).Collect(), "Cycle")
	assert.Empty(t, Cycle(Empty[int]()).Take(3).Collect(), "Cycle empty")
}

func TestFallibleSequences(t *testing.T) {
	parse := func(s string) (int, error) {
		if s == "bad" {
			return 0, errBad
		}
		return len(s), nil
	}
	got, err := Try(TryMap(Of("aa", "bbb"), parse))
	require.NoError(t, err, "Try")
	assert.Equal(t, []int{2, 3}, got, "Try values")

	partial, err := Try(TryMap(Of("aa", "bad", "cc"), parse))
	require.ErrorIs(t, err, errBad, "Try")
	assert.Equal(t, []int{2}, partial, "Try stops at the error")

	// The Try call above only compiles because TryMap returns exactly
	// iter.Seq2[int, error]; being a plain stdlib type, it also ranges directly.
	var seen int
	for range TryMap(Of("a", "bad"), parse) {
		seen++
	}
	assert.Equal(t, 2, seen, "pairs ranged over")
}

// Min and Max no longer delegate to their Func twins, so their NaN ordering is
// now this code's responsibility rather than cmp.Compare's. cmp.Less orders NaN
// below every non-NaN, which is what cmp.Compare did.
func TestMinMaxNaNOrdering(t *testing.T) {
	nan := math.NaN()
	for _, tc := range []struct {
		name  string
		in    []float64
		min   float64
		max   float64
		empty bool
	}{
		{name: "no NaN", in: []float64{2, 1, 3}, min: 1, max: 3},
		{name: "NaN first", in: []float64{nan, 1, 2}, min: nan, max: 2},
		{name: "NaN middle", in: []float64{1, nan, 2}, min: nan, max: 2},
		{name: "NaN last", in: []float64{2, 1, nan}, min: nan, max: 2},
		{name: "all NaN", in: []float64{nan, nan}, min: nan, max: nan},
		{
			name: "infinities", in: []float64{math.Inf(-1), 0, math.Inf(1), nan},
			min: nan, max: math.Inf(1),
		},
		{name: "empty", in: nil, empty: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gotMin, okMin := Min(Of(tc.in...))
			gotMax, okMax := Max(Of(tc.in...))
			if tc.empty {
				require.False(t, okMin, "empty must report false")
				require.False(t, okMax, "empty must report false")
				return
			}
			// Compare against cmp.Compare, the ordering the delegating
			// implementation used, so any drift shows up here.
			wantMin, _ := Of(tc.in...).MinFunc(cmp.Compare[float64])
			wantMax, _ := Of(tc.in...).MaxFunc(cmp.Compare[float64])
			// NaN is never equal to itself, so assert.Equal cannot express this.
			assert.Truef(t, sameFloat(gotMin, wantMin) && sameFloat(gotMin, tc.min),
				"Min = %v, want %v (MinFunc says %v)", gotMin, tc.min, wantMin)
			assert.Truef(t, sameFloat(gotMax, wantMax) && sameFloat(gotMax, tc.max),
				"Max = %v, want %v (MaxFunc says %v)", gotMax, tc.max, wantMax)
		})
	}
}

func sameFloat(a, b float64) bool {
	return a == b || (math.IsNaN(a) && math.IsNaN(b))
}

// Sort now calls slices.Sorted rather than SortFunc(cmp.Compare); pin that the
// ordering, including NaN placement, did not move.
func TestSortMatchesSortFunc(t *testing.T) {
	nan := math.NaN()
	for _, in := range [][]float64{
		{3, 1, 2},
		{nan, 1, 2},
		{1, nan, 2},
		{math.Inf(1), nan, math.Inf(-1), 0},
		{},
	} {
		got := Sort(Of(in...)).Collect()
		want := Of(in...).SortFunc(cmp.Compare[float64]).Collect()
		require.Lenf(t, got, len(want), "Sort(%v) length", in)
		for i := range got {
			// NaN is never equal to itself, so assert.Equal cannot express this.
			assert.Truef(t, sameFloat(got[i], want[i]), "Sort(%v) = %v, want %v", in, got, want)
		}
	}
}

// Ok is the lazy counterpart of Try. Try holds every value in a slice, so a
// fallible source larger than memory has no path into a pipeline through it;
// Ok gives one. These tests pin that it really is lazy, since that is its
// entire reason to exist.
func TestOk(t *testing.T) {
	fallible := func(n int, failAt int) (iter.Seq2[int, error], *int) {
		read := 0
		return func(yield func(int, error) bool) {
			for i := range n {
				read++
				if i == failAt {
					yield(0, errBad)
					return
				}
				if !yield(i, nil) {
					return
				}
			}
		}, &read
	}

	t.Run("passes the values through", func(t *testing.T) {
		seq, _ := fallible(3, -1)
		s, err := Ok(seq)
		assert.Equal(t, []int{0, 1, 2}, s.Collect(), "values")
		assert.NoError(t, err())
	})

	t.Run("stops at the first error and reports it", func(t *testing.T) {
		seq, _ := fallible(10, 2)
		s, err := Ok(seq)
		assert.Equal(t, []int{0, 1}, s.Collect(), "values before the error")
		assert.ErrorIs(t, err(), errBad)
	})

	t.Run("reads only what the consumer asks for", func(t *testing.T) {
		seq, read := fallible(1_000_000, -1)
		s, err := Ok(seq)
		got := s.Filter(func(i int) bool { return i%2 == 0 }).Take(3).Collect()
		assert.Equal(t, []int{0, 2, 4}, got, "values")
		assert.Equal(t, 5, *read, "elements of the source consumed -- more means Ok is buffering")
		assert.NoError(t, err())
	})

	t.Run("an empty source is not an error", func(t *testing.T) {
		s, err := Ok(func(_ func(int, error) bool) {})
		assert.Empty(t, s.Collect(), "values")
		assert.NoError(t, err())
	})
}

func TestKeyBy(t *testing.T) {
	s := Of("apple", "banana", "avocado")
	initial := func(v string) string { return v[:1] }

	assert.Equal(t, []string{"a", "b", "a"}, s.KeyBy(initial).Keys().Collect(), "keys")
	assert.Equal(t, []string{"apple", "banana", "avocado"},
		s.KeyBy(initial).Values().Collect(), "values")
	assert.Empty(t, Empty[string]().KeyBy(initial).Keys().Collect(), "empty")

	// The key type need not be comparable at this point; only the consumer
	// that groups by it needs that.
	type box struct{ n []int }
	assert.Equal(t, 2, Of(1, 2).KeyBy(func(i int) box { return box{[]int{i}} }).Count(),
		"a key type that is not comparable")

	// Lazy: an infinite source stays usable.
	consumed := 0
	got := Range(0, 1_000_000).
		Peek(func(int) { consumed++ }).
		KeyBy(func(i int) int { return i % 3 }).
		Take(2).
		Values().
		Collect()
	assert.Equal(t, []int{0, 1}, got, "lazy")
	assert.Equal(t, 2, consumed, "elements consumed")
}
