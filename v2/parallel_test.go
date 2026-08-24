package streams

import (
	"fmt"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestParallelMapPreservesOrder(t *testing.T) {
	// Sleeping longer for earlier elements would expose any reordering.
	got := Range(0, 20).ParallelMap(func(i int) int {
		time.Sleep(time.Duration(20-i) * time.Millisecond)
		return i * 2
	}, WithConcurrency(8)).Collect()

	want := Range(0, 20).Map(func(i int) int { return i * 2 }).Collect()
	if !slices.Equal(got, want) {
		t.Errorf("ParallelMap = %v, want %v", got, want)
	}
}

func TestParallelMapUnorderedKeepsEveryElement(t *testing.T) {
	got := Range(0, 50).ParallelMap(func(i int) int { return i * 2 },
		WithConcurrency(8), Unordered()).Collect()
	slices.Sort(got)
	want := Range(0, 50).Map(func(i int) int { return i * 2 }).Collect()
	if !slices.Equal(got, want) {
		t.Errorf("unordered lost or duplicated elements: got %d, want %d", len(got), len(want))
	}
}

// The bound must hold exactly, for every entry point and both orderings. The
// earlier version of this test asserted peak <= limit+1, which encoded the
// overshoot the ordered paths had rather than catching it, and it passed even
// when WithConcurrency was ignored entirely.
func TestParallelRespectsConcurrencyLimitExactly(t *testing.T) {
	for _, limit := range []int{1, 2, 4, 8} {
		for _, unordered := range []bool{false, true} {
			opts := []ParallelOption{WithConcurrency(limit)}
			name := "ordered"
			if unordered {
				opts, name = append(opts, Unordered()), "unordered"
			}
			for _, op := range []string{"Map", "Filter", "ForEach"} {
				if unordered && op == "ForEach" {
					continue // ForEach has no ordering mode
				}
				t.Run(fmt.Sprintf("%s/%s/%d", op, name, limit), func(t *testing.T) {
					var inFlight, peak atomic.Int64
					work := func(i int) int {
						n := inFlight.Add(1)
						for {
							old := peak.Load()
							if n <= old || peak.CompareAndSwap(old, n) {
								break
							}
						}
						time.Sleep(time.Millisecond)
						inFlight.Add(-1)
						return i
					}
					switch op {
					case "Map":
						Range(0, 60).ParallelMap(work, opts...).Collect()
					case "Filter":
						Range(0, 60).ParallelFilter(func(i int) bool { work(i); return true }, opts...).Collect()
					case "ForEach":
						Range(0, 60).ParallelForEach(func(i int) { work(i) }, opts...)
					}
					if got := peak.Load(); got != int64(limit) {
						t.Errorf("peak concurrency = %d, want exactly %d", got, limit)
					}
				})
			}
		}
	}
}

func TestParallelMapStopsEarlyWithoutLeaking(t *testing.T) {
	before := countGoroutines()
	var started atomic.Int64
	got := Range(0, 10_000).ParallelMap(func(i int) int {
		started.Add(1)
		return i
	}, WithConcurrency(4)).Take(5).Collect()

	if !slices.Equal(got, []int{0, 1, 2, 3, 4}) {
		t.Errorf("Take after ParallelMap = %v", got)
	}
	if n := started.Load(); n > 200 {
		t.Errorf("early exit still processed %d elements", n)
	}
	// give the abandoned workers a moment to unwind
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if countGoroutines() <= before+2 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Errorf("goroutines leaked: before=%d after=%d", before, countGoroutines())
}

func TestParallelFilter(t *testing.T) {
	got := Range(0, 30).ParallelFilter(func(i int) bool { return i%3 == 0 },
		WithConcurrency(4)).Collect()
	want := Range(0, 30).Filter(func(i int) bool { return i%3 == 0 }).Collect()
	if !slices.Equal(got, want) {
		t.Errorf("ParallelFilter = %v, want %v", got, want)
	}

	unordered := Range(0, 30).ParallelFilter(func(i int) bool { return i%3 == 0 },
		WithConcurrency(4), Unordered()).Collect()
	slices.Sort(unordered)
	if !slices.Equal(unordered, want) {
		t.Errorf("unordered ParallelFilter = %v, want %v", unordered, want)
	}
}

func TestParallelForEach(t *testing.T) {
	var mu sync.Mutex
	seen := map[int]bool{}
	Range(0, 100).ParallelForEach(func(i int) {
		mu.Lock()
		defer mu.Unlock()
		seen[i] = true
	}, WithConcurrency(8))
	if len(seen) != 100 {
		t.Errorf("ParallelForEach visited %d elements, want 100", len(seen))
	}
}

func TestParallelDefaultsAndDegenerateConcurrency(t *testing.T) {
	// concurrency below one is clamped rather than deadlocking
	got := Of(1, 2, 3).ParallelMap(func(i int) int { return i }, WithConcurrency(0)).Collect()
	if !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("WithConcurrency(0) = %v", got)
	}
	// no options at all
	if got := Of(1, 2).ParallelMap(func(i int) int { return i * 3 }).Collect(); !slices.Equal(got, []int{3, 6}) {
		t.Errorf("default options = %v", got)
	}
	if got := Empty[int]().ParallelMap(func(i int) int { return i }).Collect(); len(got) != 0 {
		t.Errorf("empty = %v", got)
	}
}
