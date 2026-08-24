package streams

import (
	"math/rand/v2"
	"slices"
	"testing"
)

// Merge maintains a hand-written k-way heap. Compare it against a full sort of
// the concatenated inputs across many randomly shaped cases.
func TestMergeAgainstSort(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 2))
	cmpInt := func(a, b int) int { return a - b }

	for trial := range 300 {
		k := 1 + r.IntN(6)
		inputs := make([]Stream[int], 0, k)
		var all []int
		for range k {
			n := r.IntN(8)
			part := make([]int, n)
			for i := range part {
				part[i] = r.IntN(20)
			}
			slices.Sort(part)
			all = append(all, part...)
			inputs = append(inputs, Of(part...))
		}
		slices.Sort(all)

		got := Merge(cmpInt, inputs...).Collect()
		if !slices.Equal(got, all) {
			t.Fatalf("trial %d: Merge = %v, want %v", trial, got, all)
		}
	}
}

func TestMergeStopsEarlyOnEveryShape(t *testing.T) {
	cmpInt := func(a, b int) int { return a - b }
	// Taking one element must not drain the inputs or panic.
	for _, inputs := range [][]Stream[int]{
		{Of(1, 2, 3)},
		{Of(1), Of(2), Of(3)},
		{Empty[int](), Of(1, 2)},
		{Of(1, 2), Empty[int]()},
	} {
		got := Merge(cmpInt, inputs...).Take(1).Collect()
		if len(got) != 1 {
			t.Errorf("Take(1) after Merge = %v", got)
		}
	}
}

// Merge must stream, not buffer: an infinite input is fine as long as the
// consumer stops.
func TestMergeStreamsInfiniteInputs(t *testing.T) {
	cmpInt := func(a, b int) int { return a - b }
	evens := Iterate(0, func(i int) int { return i + 2 })
	odds := Iterate(1, func(i int) int { return i + 2 })
	got := Merge(cmpInt, evens, odds).Take(6).Collect()
	if !slices.Equal(got, []int{0, 1, 2, 3, 4, 5}) {
		t.Errorf("Merge of infinite inputs = %v", got)
	}
}

func TestWindowSharesOverlappingElements(t *testing.T) {
	// Windows are cut from a shared array, so the elements two windows have in
	// common are one element, not two copies. This is the documented contract;
	// it is what makes a sliding window cost no allocation per element.
	windows := Window(Of(1, 2, 3, 4, 5), 3).Collect()
	if len(windows) != 3 {
		t.Fatalf("got %d windows, want 3", len(windows))
	}
	eq(t, "windows", windows[0], []int{1, 2, 3})
	eq(t, "windows", windows[1], []int{2, 3, 4})
	eq(t, "windows", windows[2], []int{3, 4, 5})

	// Element 3 sits at index 2 of the first window, 1 of the second and 0 of
	// the third. Writing through one is visible in the others.
	windows[0][2] = 99
	if windows[1][1] != 99 || windows[2][0] != 99 {
		t.Errorf("overlapping elements are not shared: %v", windows)
	}

	// Where independence is wanted, the caller asks for it.
	windows = Window(Of(1, 2, 3, 4, 5), 3).Map(slices.Clone).Collect()
	windows[0][2] = 99
	if windows[1][1] != 3 {
		t.Errorf("Map(slices.Clone) must give independent windows: %v", windows)
	}
}

// A window must stay correct after the arena it was cut from has been retired,
// which happens once the backing array fills. With a small n that takes a few
// hundred elements, so this walks well past the first refill.
func TestWindowSurvivesArenaRefills(t *testing.T) {
	const n, count = 3, 500
	windows := Window(Range(0, count), n).Collect()
	if len(windows) != count-n+1 {
		t.Fatalf("got %d windows, want %d", len(windows), count-n+1)
	}
	// Every window, including the earliest, still holds what it held when it
	// was yielded -- no refill overwrote it.
	for i, w := range windows {
		for j := range n {
			if w[j] != i+j {
				t.Fatalf("window %d = %v, want [%d..%d]", i, w, i, i+n-1)
			}
		}
	}
}

func TestChunkDoesNotAliasItsBuffer(t *testing.T) {
	chunks := Chunk(Of(1, 2, 3, 4), 2).Collect()
	chunks[0][0] = 99
	if chunks[1][0] != 3 {
		t.Errorf("chunks share a buffer: %v", chunks)
	}
}

func TestCycleStoppedDuringFirstPass(t *testing.T) {
	if got := Cycle(Of(1, 2, 3)).Take(2).Collect(); !slices.Equal(got, []int{1, 2}) {
		t.Errorf("Cycle stopped in the first pass = %v", got)
	}
	if got := Cycle(Of(1, 2, 3)).Take(3).Collect(); !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("Cycle stopped at the pass boundary = %v", got)
	}
	if got := Cycle(Of(1, 2, 3)).Take(4).Collect(); !slices.Equal(got, []int{1, 2, 3, 1}) {
		t.Errorf("Cycle across the boundary = %v", got)
	}
}
