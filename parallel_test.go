package streams

import (
	"fmt"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// The expected value below, and in the two tests after it, is what the serial
// operation produces: what these pin is that adding concurrency changes
// nothing. A regression in Range or Map itself moves both sides at once and so
// is invisible here on purpose — TestConstructors and TestIntermediateOps hold
// those to literal values.

func TestParallelMapPreservesOrder(t *testing.T) {
	// Sleeping longer for earlier elements would expose any reordering.
	got := Range(0, 20).ParallelMap(func(i int) int {
		time.Sleep(time.Duration(20-i) * time.Millisecond)
		return i * 2
	}, WithConcurrency(8)).Collect()

	want := Range(0, 20).Map(func(i int) int { return i * 2 }).Collect()
	assert.Equal(t, want, got, "ParallelMap")
}

func TestParallelMapUnorderedKeepsEveryElement(t *testing.T) {
	got := Range(0, 50).ParallelMap(func(i int) int { return i * 2 },
		WithConcurrency(8), Unordered()).Collect()
	slices.Sort(got)
	want := Range(0, 50).Map(func(i int) int { return i * 2 }).Collect()
	assert.Equal(t, want, got, "unordered lost or duplicated elements")
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
					assert.Equal(t, int64(limit), peak.Load(), "peak concurrency")
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

	assert.Equal(t, []int{0, 1, 2, 3, 4}, got, "Take after ParallelMap")
	assert.LessOrEqual(t, started.Load(), int64(200), "elements processed after the early exit")

	// give the abandoned workers a moment to unwind
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if countGoroutines() <= before+2 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	// Fail outright rather than re-testing the count: the loop above has already
	// established it, and a fresh reading could flip it to a pass.
	assert.Failf(t, "goroutines leaked", "before=%d after=%d", before, countGoroutines())
}

func TestParallelFilter(t *testing.T) {
	want := Range(0, 30).Filter(func(i int) bool { return i%3 == 0 }).Collect()

	got := Range(0, 30).ParallelFilter(func(i int) bool { return i%3 == 0 },
		WithConcurrency(4)).Collect()
	assert.Equal(t, want, got, "ParallelFilter")

	unordered := Range(0, 30).ParallelFilter(func(i int) bool { return i%3 == 0 },
		WithConcurrency(4), Unordered()).Collect()
	slices.Sort(unordered)
	assert.Equal(t, want, unordered, "unordered ParallelFilter")
}

func TestParallelForEach(t *testing.T) {
	var mu sync.Mutex
	seen := map[int]bool{}
	Range(0, 100).ParallelForEach(func(i int) {
		mu.Lock()
		defer mu.Unlock()
		seen[i] = true
	}, WithConcurrency(8))
	assert.Len(t, seen, 100, "elements ParallelForEach visited")
}

func TestParallelDefaultsAndDegenerateConcurrency(t *testing.T) {
	// concurrency below one is clamped rather than deadlocking
	assert.Equal(t, []int{1, 2, 3},
		Of(1, 2, 3).ParallelMap(func(i int) int { return i }, WithConcurrency(0)).Collect(),
		"WithConcurrency(0)")
	// no options at all
	assert.Equal(t, []int{3, 6},
		Of(1, 2).ParallelMap(func(i int) int { return i * 3 }).Collect(), "default options")
	assert.Empty(t, Empty[int]().ParallelMap(func(i int) int { return i }).Collect(), "empty")
}
