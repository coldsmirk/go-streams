package streams

import (
	"cmp"
	"errors"
	"iter"
	"math"
	"slices"
	"testing"
)

func TestConstrainedFunctionsPairWithFuncMethods(t *testing.T) {
	// package function: element type is constrained, like slices.Max
	if v, ok := Max(Of(3, 9, 4)); !ok || v != 9 {
		t.Errorf("Max = %d, %v", v, ok)
	}
	if v, ok := Min(Of(3, 9, 4)); !ok || v != 3 {
		t.Errorf("Min = %d, %v", v, ok)
	}
	if _, ok := Max(Empty[int]()); ok {
		t.Error("Max on empty must report false")
	}
	// method: no constraint, caller supplies the comparison
	oldest, ok := Of(users...).MaxFunc(func(a, b user) int { return cmp.Compare(a.Age, b.Age) })
	if !ok || oldest.Name != "Ken" {
		t.Errorf("MaxFunc = %v", oldest)
	}

	eq(t, "Sort", Sort(Of(3, 1, 2)).Collect(), []int{1, 2, 3})
	eq(t, "Sort strings", Sort(Of("b", "a")).Collect(), []string{"a", "b"})
	eq(t, "Compact", Compact(Of(1, 1, 2, 2, 1)).Collect(), []int{1, 2, 1})
	eq(t, "Distinct", Distinct(Of(1, 1, 2, 2, 1)).Collect(), []int{1, 2})

	if !Contains(Of(1, 2, 3), 2) || Contains(Of(1, 2, 3), 9) {
		t.Error("Contains")
	}
	f := Frequency(Of("a", "b", "a"))
	if f["a"] != 2 || f["b"] != 1 {
		t.Errorf("Frequency = %v", f)
	}
}

func TestNumericAggregates(t *testing.T) {
	if got := Sum(Of(1, 2, 3)); got != 6 {
		t.Errorf("Sum = %d", got)
	}
	if got := Sum(Empty[int]()); got != 0 {
		t.Errorf("Sum of empty = %d", got)
	}
	if got := Sum(Of(1.5, 2.5)); got != 4.0 {
		t.Errorf("Sum float = %v", got)
	}
	if got := Product(Of(2, 3, 4)); got != 24 {
		t.Errorf("Product = %d", got)
	}
	if got := Product(Empty[int]()); got != 1 {
		t.Errorf("Product of empty = %d, want the multiplicative identity", got)
	}
	if got, ok := Average(Of(1, 2, 3, 4)); !ok || got != 2.5 {
		t.Errorf("Average = %v, %v", got, ok)
	}
	if _, ok := Average(Empty[int]()); ok {
		t.Error("Average on empty must report false")
	}
	// a defined type with a numeric underlying type satisfies Numeric
	type celsius float64
	if got := Sum(Of[celsius](1, 2)); got != 3 {
		t.Errorf("Sum of defined type = %v", got)
	}
}

func TestRegrouping(t *testing.T) {
	eq(t, "Chunk sizes", Chunk(Of(1, 2, 3, 4, 5), 2).Map(func(c []int) int { return len(c) }).Collect(),
		[]int{2, 2, 1})
	first, _ := Chunk(Of(1, 2, 3), 2).First()
	eq(t, "Chunk content", first, []int{1, 2})
	if got := Chunk(Empty[int](), 2).Collect(); len(got) != 0 {
		t.Errorf("Chunk of empty = %v", got)
	}
	eq(t, "Chunk exact", Chunk(Of(1, 2), 2).Map(func(c []int) int { return len(c) }).Collect(), []int{2})

	// Window slides one element at a time and yields nothing if the input is short
	windows := Window(Of(1, 2, 3, 4), 2).Collect()
	if len(windows) != 3 {
		t.Fatalf("Window count = %d, want 3", len(windows))
	}
	eq(t, "Window[0]", windows[0], []int{1, 2})
	eq(t, "Window[2]", windows[2], []int{3, 4})
	if got := Window(Of(1), 2).Collect(); len(got) != 0 {
		t.Errorf("Window shorter than n = %v", got)
	}
	// each window must be an independent slice, not a view of a shared buffer
	if slices.Equal(windows[0], windows[1]) {
		t.Error("Window reused its buffer across yields")
	}

	eq(t, "Flatten", Flatten(Of(Of(1, 2), Of(3), Empty[int]())).Collect(), []int{1, 2, 3})
}

func TestChunkAndWindowRejectNonPositive(t *testing.T) {
	for _, tc := range []struct {
		name string
		call func()
	}{
		{"Chunk", func() { Chunk(Of(1), 0) }},
		{"Window", func() { Window(Of(1), 0) }},
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s(n=0) must panic", tc.name)
				}
			}()
			tc.call()
		}()
	}
}

func TestCombiningStreams(t *testing.T) {
	eq(t, "Concat", Concat(Of(1, 2), Of(3), Empty[int]()).Collect(), []int{1, 2, 3})
	eq(t, "Concat none", Concat[int]().Collect(), nil)
	eq(t, "Interleave", Interleave(Of(1, 3), Of(2, 4)).Collect(), []int{1, 2, 3, 4})
	eq(t, "Interleave uneven", Interleave(Of(1), Of(2, 4, 6)).Collect(), []int{1, 2, 4, 6})

	byInt := func(a, b int) int { return a - b }
	eq(t, "Merge", Merge(byInt, Of(1, 4, 7), Of(2, 5), Of(3, 6)).Collect(),
		[]int{1, 2, 3, 4, 5, 6, 7})
	eq(t, "Merge one", Merge(byInt, Of(3, 1)).Collect(), []int{3, 1})
	eq(t, "Merge none", Merge(byInt).Collect(), nil)
	eq(t, "Merge with empty", Merge(byInt, Empty[int](), Of(1, 2)).Collect(), []int{1, 2})

	eq(t, "Cycle", Cycle(Of(1, 2)).Take(5).Collect(), []int{1, 2, 1, 2, 1})
	eq(t, "Cycle empty", Cycle(Empty[int]()).Take(3).Collect(), nil)
}

func TestFallibleSequences(t *testing.T) {
	parse := func(s string) (int, error) {
		if s == "bad" {
			return 0, errBad
		}
		return len(s), nil
	}
	got, err := Try(TryMap(Of("aa", "bbb"), parse))
	if err != nil {
		t.Fatalf("Try returned %v", err)
	}
	eq(t, "Try values", got, []int{2, 3})

	partial, err := Try(TryMap(Of("aa", "bad", "cc"), parse))
	if !errors.Is(err, errBad) {
		t.Fatalf("Try err = %v, want errBad", err)
	}
	eq(t, "Try stops at the error", partial, []int{2})

	// The Try call above only compiles because TryMap returns exactly
	// iter.Seq2[int, error]; being a plain stdlib type, it also ranges directly.
	var seen int
	for range TryMap(Of("a", "bad"), parse) {
		seen++
	}
	if seen != 2 {
		t.Errorf("ranged over %d pairs, want 2", seen)
	}
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
		{name: "infinities", in: []float64{math.Inf(-1), 0, math.Inf(1), nan},
			min: nan, max: math.Inf(1)},
		{name: "empty", in: nil, empty: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gotMin, okMin := Min(Of(tc.in...))
			gotMax, okMax := Max(Of(tc.in...))
			if tc.empty {
				if okMin || okMax {
					t.Fatalf("empty must report false, got %v %v", okMin, okMax)
				}
				return
			}
			// Compare against cmp.Compare, the ordering the delegating
			// implementation used, so any drift shows up here.
			wantMin, _ := Of(tc.in...).MinFunc(cmp.Compare[float64])
			wantMax, _ := Of(tc.in...).MaxFunc(cmp.Compare[float64])
			if !sameFloat(gotMin, wantMin) || !sameFloat(gotMin, tc.min) {
				t.Errorf("Min = %v, want %v (MinFunc says %v)", gotMin, tc.min, wantMin)
			}
			if !sameFloat(gotMax, wantMax) || !sameFloat(gotMax, tc.max) {
				t.Errorf("Max = %v, want %v (MaxFunc says %v)", gotMax, tc.max, wantMax)
			}
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
		if len(got) != len(want) {
			t.Fatalf("Sort(%v) length = %d, want %d", in, len(got), len(want))
		}
		for i := range got {
			if !sameFloat(got[i], want[i]) {
				t.Errorf("Sort(%v) = %v, want %v", in, got, want)
				break
			}
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
		eq(t, "values", s.Collect(), []int{0, 1, 2})
		if err() != nil {
			t.Errorf("err = %v, want nil", err())
		}
	})

	t.Run("stops at the first error and reports it", func(t *testing.T) {
		seq, _ := fallible(10, 2)
		s, err := Ok(seq)
		eq(t, "values before the error", s.Collect(), []int{0, 1})
		if !errors.Is(err(), errBad) {
			t.Errorf("err = %v, want errBad", err())
		}
	})

	t.Run("reads only what the consumer asks for", func(t *testing.T) {
		seq, read := fallible(1_000_000, -1)
		s, err := Ok(seq)
		got := s.Filter(func(i int) bool { return i%2 == 0 }).Take(3).Collect()
		eq(t, "values", got, []int{0, 2, 4})
		if *read != 5 {
			t.Errorf("consumed %d elements of the source, want 5 -- Ok is buffering", *read)
		}
		if err() != nil {
			t.Errorf("err = %v, want nil", err())
		}
	})

	t.Run("an empty source is not an error", func(t *testing.T) {
		s, err := Ok(func(_ func(int, error) bool) {})
		if got := s.Collect(); len(got) != 0 {
			t.Errorf("values = %v", got)
		}
		if err() != nil {
			t.Errorf("err = %v, want nil", err())
		}
	})
}

func TestKeyBy(t *testing.T) {
	s := Of("apple", "banana", "avocado")
	initial := func(v string) string { return v[:1] }

	eq(t, "keys", s.KeyBy(initial).Keys().Collect(), []string{"a", "b", "a"})
	eq(t, "values", s.KeyBy(initial).Values().Collect(),
		[]string{"apple", "banana", "avocado"})
	eq(t, "empty", Empty[string]().KeyBy(initial).Keys().Collect(), nil)

	// The key type need not be comparable at this point; only the consumer
	// that groups by it needs that.
	type box struct{ n []int }
	if got := Of(1, 2).KeyBy(func(i int) box { return box{[]int{i}} }).Count(); got != 2 {
		t.Errorf("count = %d", got)
	}

	// Lazy: an infinite source stays usable.
	consumed := 0
	got := Range(0, 1_000_000).
		Peek(func(int) { consumed++ }).
		KeyBy(func(i int) int { return i % 3 }).
		Take(2).
		Values().
		Collect()
	eq(t, "lazy", got, []int{0, 1})
	if consumed != 2 {
		t.Errorf("consumed %d, want 2", consumed)
	}
}
