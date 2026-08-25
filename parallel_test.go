package streams

import (
	"context"
	"runtime"
	"sort"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// testCtx returns a background context for tests.
func testCtx() context.Context {
	return context.Background()
}

// --- ParallelConfig Tests ---

func TestDefaultParallelConfig(t *testing.T) {
	t.Parallel()
	cfg := DefaultParallelConfig()
	assert.Equal(t, runtime.NumCPU(), cfg.Concurrency, "Default Concurrency should be NumCPU")
	assert.True(t, cfg.Ordered, "Default Ordered should be true")
	assert.Equal(t, runtime.NumCPU()*2, cfg.BufferSize, "Default BufferSize should be 2*NumCPU")
	assert.Equal(t, 0, cfg.ChunkSize, "Default ChunkSize 0 means streaming")
}

func TestWithConcurrency(t *testing.T) {
	t.Parallel()
	t.Run("PositiveValue", func(t *testing.T) {
		t.Parallel()
		cfg := DefaultParallelConfig()
		WithConcurrency(4)(&cfg)
		assert.Equal(t, 4, cfg.Concurrency, "WithConcurrency should set positive values")
	})

	t.Run("ZeroValueIgnored", func(t *testing.T) {
		t.Parallel()
		cfg := DefaultParallelConfig()
		original := cfg.Concurrency
		WithConcurrency(0)(&cfg)
		assert.Equal(t, original, cfg.Concurrency, "WithConcurrency(0) should be ignored")
	})

	t.Run("NegativeValueIgnored", func(t *testing.T) {
		t.Parallel()
		cfg := DefaultParallelConfig()
		original := cfg.Concurrency
		WithConcurrency(-1)(&cfg)
		assert.Equal(t, original, cfg.Concurrency, "WithConcurrency(<0) should be ignored")
	})
}

func TestWithOrdered(t *testing.T) {
	t.Parallel()
	t.Run("SetTrue", func(t *testing.T) {
		t.Parallel()
		cfg := DefaultParallelConfig()
		cfg.Ordered = false
		WithOrdered(true)(&cfg)
		assert.True(t, cfg.Ordered, "WithOrdered(true) should set ordered mode")
	})

	t.Run("SetFalse", func(t *testing.T) {
		t.Parallel()
		cfg := DefaultParallelConfig()
		WithOrdered(false)(&cfg)
		assert.False(t, cfg.Ordered, "WithOrdered(false) should set unordered mode")
	})
}

func TestWithBufferSize(t *testing.T) {
	t.Parallel()
	t.Run("PositiveValue", func(t *testing.T) {
		t.Parallel()
		cfg := DefaultParallelConfig()
		WithBufferSize(10)(&cfg)
		assert.Equal(t, 10, cfg.BufferSize, "WithBufferSize should set positive values")
	})

	t.Run("ZeroValueIgnored", func(t *testing.T) {
		t.Parallel()
		cfg := DefaultParallelConfig()
		original := cfg.BufferSize
		WithBufferSize(0)(&cfg)
		assert.Equal(t, original, cfg.BufferSize, "WithBufferSize(0) should be ignored")
	})

	t.Run("NegativeValueIgnored", func(t *testing.T) {
		t.Parallel()
		cfg := DefaultParallelConfig()
		original := cfg.BufferSize
		WithBufferSize(-1)(&cfg)
		assert.Equal(t, original, cfg.BufferSize, "WithBufferSize(<0) should be ignored")
	})
}

func TestWithChunkSize(t *testing.T) {
	t.Parallel()
	t.Run("PositiveValue", func(t *testing.T) {
		t.Parallel()
		cfg := DefaultParallelConfig()
		WithChunkSize(5)(&cfg)
		assert.Equal(t, 5, cfg.ChunkSize, "WithChunkSize should set positive values")
	})

	t.Run("ZeroValueSetsStreamingMode", func(t *testing.T) {
		t.Parallel()
		cfg := DefaultParallelConfig()
		cfg.ChunkSize = 10
		WithChunkSize(0)(&cfg)
		assert.Equal(t, 0, cfg.ChunkSize, "WithChunkSize(0) should switch to streaming mode")
	})

	t.Run("NegativeValueIgnored", func(t *testing.T) {
		t.Parallel()
		cfg := DefaultParallelConfig()
		cfg.ChunkSize = 5
		WithChunkSize(-1)(&cfg)
		assert.Equal(t, 5, cfg.ChunkSize, "WithChunkSize(<0) should be ignored")
	})
}

// --- ParallelMap Tests ---

func TestParallelMap(t *testing.T) {
	t.Parallel()
	t.Run("OrderedBasic", func(t *testing.T) {
		t.Parallel()
		result := ParallelMap(Of(1, 2, 3, 4, 5), func(n int) int {
			return n * 2
		}, WithConcurrency(2), WithOrdered(true)).Collect()
		assert.Equal(t, []int{2, 4, 6, 8, 10}, result, "ParallelMap ordered should preserve order and map *2")
	})

	t.Run("OrderedEmptyStream", func(t *testing.T) {
		t.Parallel()
		result := ParallelMap(Empty[int](), func(n int) int {
			return n * 2
		}, WithOrdered(true)).Collect()
		assert.Empty(t, result, "ParallelMap ordered on empty should be empty")
	})

	t.Run("OrderedSingleElement", func(t *testing.T) {
		t.Parallel()
		result := ParallelMap(Of(42), func(n int) int {
			return n * 2
		}, WithOrdered(true)).Collect()
		assert.Equal(t, []int{84}, result, "ParallelMap ordered single should map")
	})

	t.Run("OrderedEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := ParallelMap(Of(1, 2, 3, 4, 5, 6, 7, 8, 9, 10), func(n int) int {
			return n * 2
		}, WithConcurrency(2), WithOrdered(true)).Limit(3).Collect()
		assert.Equal(t, []int{2, 4, 6}, result, "ParallelMap ordered should respect Limit for early termination")
	})

	t.Run("OrderedChunkedBasic", func(t *testing.T) {
		t.Parallel()
		result := ParallelMap(Of(1, 2, 3, 4, 5), func(n int) int {
			return n * 2
		}, WithConcurrency(2), WithOrdered(true), WithChunkSize(2)).Collect()
		assert.Equal(t, []int{2, 4, 6, 8, 10}, result, "ParallelMap ordered chunked should preserve order")
	})

	t.Run("OrderedChunkedEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := ParallelMap(Of(1, 2, 3, 4, 5), func(n int) int {
			return n * 2
		}, WithConcurrency(2), WithOrdered(true), WithChunkSize(2)).Limit(3).Collect()
		assert.Equal(t, []int{2, 4, 6}, result, "ParallelMap ordered chunked should respect Limit")
	})

	t.Run("OrderedChunkedChunkSizeLargerThanInput", func(t *testing.T) {
		t.Parallel()
		result := ParallelMap(Of(1, 2), func(n int) int {
			return n * 2
		}, WithOrdered(true), WithChunkSize(10)).Collect()
		assert.Equal(t, []int{2, 4}, result, "ParallelMap ordered chunked should handle large chunk size")
	})

	t.Run("UnorderedBasic", func(t *testing.T) {
		t.Parallel()
		result := ParallelMap(Of(1, 2, 3, 4, 5), func(n int) int {
			return n * 2
		}, WithConcurrency(2), WithOrdered(false)).Collect()

		assert.Len(t, result, 5, "ParallelMap unordered should produce all elements")
		sort.Ints(result)
		assert.Equal(t, []int{2, 4, 6, 8, 10}, result, "ParallelMap unordered values should match set")
	})

	t.Run("UnorderedEmptyStream", func(t *testing.T) {
		t.Parallel()
		result := ParallelMap(Empty[int](), func(n int) int {
			return n * 2
		}, WithOrdered(false)).Collect()
		assert.Empty(t, result, "ParallelMap unordered on empty should be empty")
	})

	t.Run("UnorderedEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := ParallelMap(Of(1, 2, 3, 4, 5, 6, 7, 8, 9, 10), func(n int) int {
			return n * 2
		}, WithConcurrency(2), WithOrdered(false)).Limit(3).Collect()
		assert.Len(t, result, 3, "ParallelMap unordered should stop after Limit elements")
	})

	t.Run("TypeTransformation", func(t *testing.T) {
		t.Parallel()
		result := ParallelMap(Of(1, 2, 3), func(n int) string {
			return string(rune('a' + n - 1))
		}, WithOrdered(true)).Collect()
		assert.Equal(t, []string{"a", "b", "c"}, result, "ParallelMap should support type transformation")
	})

	t.Run("LargeConcurrency", func(t *testing.T) {
		t.Parallel()
		result := ParallelMap(Of(1, 2, 3), func(n int) int {
			return n * 2
		}, WithConcurrency(100), WithOrdered(true)).Collect()
		assert.Equal(t, []int{2, 4, 6}, result, "ParallelMap with large concurrency should still be correct")
	})

	t.Run("DefaultOptionsUsesOrderedMode", func(t *testing.T) {
		t.Parallel()
		result := ParallelMap(Of(1, 2, 3, 4, 5), func(n int) int {
			return n * 2
		}).Collect()
		assert.Equal(t, []int{2, 4, 6, 8, 10}, result, "ParallelMap default should be ordered")
	})
}

// --- ParallelFilter Tests ---

func TestParallelFilter(t *testing.T) {
	t.Parallel()
	t.Run("OrderedBasic", func(t *testing.T) {
		t.Parallel()
		result := ParallelFilter(Of(1, 2, 3, 4, 5, 6), func(n int) bool {
			return n%2 == 0
		}, WithConcurrency(2), WithOrdered(true)).Collect()
		assert.Equal(t, []int{2, 4, 6}, result, "ParallelFilter ordered should preserve even numbers and order")
	})

	t.Run("OrderedEmptyStream", func(t *testing.T) {
		t.Parallel()
		result := ParallelFilter(Empty[int](), func(_ int) bool {
			return true
		}, WithOrdered(true)).Collect()
		assert.Empty(t, result, "ParallelFilter ordered on empty should be empty")
	})

	t.Run("OrderedSingleElementPasses", func(t *testing.T) {
		t.Parallel()
		result := ParallelFilter(Of(42), func(_ int) bool {
			return true
		}, WithOrdered(true)).Collect()
		assert.Equal(t, []int{42}, result, "ParallelFilter ordered should keep element if predicate true")
	})

	t.Run("OrderedSingleElementFails", func(t *testing.T) {
		t.Parallel()
		result := ParallelFilter(Of(42), func(_ int) bool {
			return false
		}, WithOrdered(true)).Collect()
		assert.Empty(t, result, "ParallelFilter ordered should drop element if predicate false")
	})

	t.Run("OrderedAllFilteredOut", func(t *testing.T) {
		t.Parallel()
		result := ParallelFilter(Of(1, 3, 5, 7), func(n int) bool {
			return n%2 == 0
		}, WithOrdered(true)).Collect()
		assert.Empty(t, result, "ParallelFilter ordered should drop all when none match")
	})

	t.Run("OrderedEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := ParallelFilter(Of(1, 2, 3, 4, 5, 6, 7, 8, 9, 10), func(n int) bool {
			return n%2 == 0
		}, WithConcurrency(2), WithOrdered(true)).Limit(2).Collect()
		assert.Equal(t, []int{2, 4}, result, "ParallelFilter ordered should respect Limit")
	})

	t.Run("OrderedChunkedBasic", func(t *testing.T) {
		t.Parallel()
		result := ParallelFilter(Of(1, 2, 3, 4, 5, 6), func(n int) bool {
			return n%2 == 0
		}, WithConcurrency(2), WithOrdered(true), WithChunkSize(2)).Collect()
		assert.Equal(t, []int{2, 4, 6}, result, "ParallelFilter ordered chunked should preserve order")
	})

	t.Run("OrderedChunkedEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := ParallelFilter(Of(1, 2, 3, 4, 5, 6), func(n int) bool {
			return n%2 == 0
		}, WithConcurrency(2), WithOrdered(true), WithChunkSize(2)).Limit(2).Collect()
		assert.Equal(t, []int{2, 4}, result, "ParallelFilter ordered chunked should respect Limit")
	})

	t.Run("OrderedChunkedChunkSizeLargerThanInput", func(t *testing.T) {
		t.Parallel()
		result := ParallelFilter(Of(1, 2), func(n int) bool {
			return n%2 == 0
		}, WithOrdered(true), WithChunkSize(10)).Collect()
		assert.Equal(t, []int{2}, result, "ParallelFilter ordered chunked should handle large chunk size")
	})

	t.Run("UnorderedBasic", func(t *testing.T) {
		t.Parallel()
		result := ParallelFilter(Of(1, 2, 3, 4, 5, 6), func(n int) bool {
			return n%2 == 0
		}, WithConcurrency(2), WithOrdered(false)).Collect()

		assert.Len(t, result, 3, "ParallelFilter unordered should produce all matches")
		sort.Ints(result)
		assert.Equal(t, []int{2, 4, 6}, result, "ParallelFilter unordered values should match set")
	})

	t.Run("UnorderedEmptyStream", func(t *testing.T) {
		t.Parallel()
		result := ParallelFilter(Empty[int](), func(_ int) bool {
			return true
		}, WithOrdered(false)).Collect()
		assert.Empty(t, result, "ParallelFilter unordered on empty should be empty")
	})

	t.Run("UnorderedEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := ParallelFilter(Of(1, 2, 3, 4, 5, 6, 7, 8, 9, 10), func(n int) bool {
			return n%2 == 0
		}, WithConcurrency(2), WithOrdered(false)).Limit(2).Collect()
		assert.Len(t, result, 2, "ParallelFilter unordered should stop after Limit elements")
	})

	t.Run("DefaultOptionsUsesOrderedMode", func(t *testing.T) {
		t.Parallel()
		result := ParallelFilter(Of(1, 2, 3, 4, 5, 6), func(n int) bool {
			return n%2 == 0
		}).Collect()
		assert.Equal(t, []int{2, 4, 6}, result, "ParallelFilter default should be ordered")
	})
}

// --- ParallelFlatMap Tests ---

func TestParallelFlatMap(t *testing.T) {
	t.Parallel()
	t.Run("OrderedStreamingBasic", func(t *testing.T) {
		t.Parallel()
		result := ParallelFlatMap(Of(1, 2, 3), func(n int) Stream[int] {
			return Of(n*10, n*10+1)
		}, WithConcurrency(2), WithOrdered(true)).Collect()
		assert.Equal(t, []int{10, 11, 20, 21, 30, 31}, result, "ParallelFlatMap ordered streaming should preserve order")
	})

	t.Run("OrderedStreamingEmptyStream", func(t *testing.T) {
		t.Parallel()
		result := ParallelFlatMap(Empty[int](), func(n int) Stream[int] {
			return Of(n, n)
		}, WithOrdered(true)).Collect()
		assert.Empty(t, result, "ParallelFlatMap ordered streaming on empty should be empty")
	})

	t.Run("OrderedStreamingEmptySubStreams", func(t *testing.T) {
		t.Parallel()
		result := ParallelFlatMap(Of(1, 2, 3), func(_ int) Stream[int] {
			return Empty[int]()
		}, WithOrdered(true)).Collect()
		assert.Empty(t, result, "ParallelFlatMap ordered streaming should drop empty sub-streams")
	})

	t.Run("OrderedStreamingEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := ParallelFlatMap(Of(1, 2, 3, 4, 5), func(n int) Stream[int] {
			return Of(n*10, n*10+1)
		}, WithConcurrency(2), WithOrdered(true)).Limit(5).Collect()
		assert.Equal(t, []int{10, 11, 20, 21, 30}, result, "ParallelFlatMap ordered streaming should respect Limit")
	})

	t.Run("OrderedChunkedBasic", func(t *testing.T) {
		t.Parallel()
		result := ParallelFlatMap(Of(1, 2, 3, 4, 5, 6), func(n int) Stream[int] {
			return Of(n*10, n*10+1)
		}, WithConcurrency(2), WithOrdered(true), WithChunkSize(2)).Collect()
		assert.Equal(t, []int{10, 11, 20, 21, 30, 31, 40, 41, 50, 51, 60, 61}, result, "ParallelFlatMap ordered chunked should preserve order")
	})

	t.Run("OrderedChunkedEmptyStream", func(t *testing.T) {
		t.Parallel()
		result := ParallelFlatMap(Empty[int](), func(n int) Stream[int] {
			return Of(n, n)
		}, WithOrdered(true), WithChunkSize(2)).Collect()
		assert.Empty(t, result, "ParallelFlatMap ordered chunked on empty should be empty")
	})

	t.Run("OrderedChunkedEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := ParallelFlatMap(Of(1, 2, 3, 4, 5, 6), func(n int) Stream[int] {
			return Of(n*10, n*10+1)
		}, WithConcurrency(2), WithOrdered(true), WithChunkSize(2)).Limit(5).Collect()
		assert.Equal(t, []int{10, 11, 20, 21, 30}, result, "ParallelFlatMap ordered chunked should respect Limit")
	})

	t.Run("OrderedChunkedChunkSizeLargerThanInput", func(t *testing.T) {
		t.Parallel()
		result := ParallelFlatMap(Of(1, 2), func(n int) Stream[int] {
			return Of(n*10, n*10+1)
		}, WithOrdered(true), WithChunkSize(100)).Collect()
		assert.Equal(t, []int{10, 11, 20, 21}, result, "ParallelFlatMap ordered chunked should handle large chunk size")
	})

	t.Run("UnorderedBasic", func(t *testing.T) {
		t.Parallel()
		result := ParallelFlatMap(Of(1, 2, 3), func(n int) Stream[int] {
			return Of(n*10, n*10+1)
		}, WithConcurrency(2), WithOrdered(false)).Collect()

		assert.Len(t, result, 6, "ParallelFlatMap unordered should produce all elements")
		sort.Ints(result)
		assert.Equal(t, []int{10, 11, 20, 21, 30, 31}, result, "ParallelFlatMap unordered values should match set")
	})

	t.Run("UnorderedEmptyStream", func(t *testing.T) {
		t.Parallel()
		result := ParallelFlatMap(Empty[int](), func(n int) Stream[int] {
			return Of(n, n)
		}, WithOrdered(false)).Collect()
		assert.Empty(t, result, "ParallelFlatMap unordered on empty should be empty")
	})

	t.Run("UnorderedEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := ParallelFlatMap(Of(1, 2, 3, 4, 5), func(n int) Stream[int] {
			return Of(n*10, n*10+1)
		}, WithConcurrency(2), WithOrdered(false)).Limit(4).Collect()
		assert.Len(t, result, 4, "ParallelFlatMap unordered should stop after Limit elements")
	})

	t.Run("VariableLengthSubStreams", func(t *testing.T) {
		t.Parallel()
		streamingResult := ParallelFlatMap(Of(3, 1, 4, 2), func(n int) Stream[int] {
			return Range(0, n)
		}, WithConcurrency(4), WithOrdered(true)).Collect()

		chunkedResult := ParallelFlatMap(Of(3, 1, 4, 2), func(n int) Stream[int] {
			return Range(0, n)
		}, WithConcurrency(4), WithOrdered(true), WithChunkSize(2)).Collect()

		expected := []int{0, 1, 2, 0, 0, 1, 2, 3, 0, 1}
		assert.Equal(t, expected, streamingResult, "ParallelFlatMap ordered streaming should match expected")
		assert.Equal(t, expected, chunkedResult, "ParallelFlatMap ordered chunked should match expected")
	})

	t.Run("GrowingSubStreams", func(t *testing.T) {
		t.Parallel()
		streamingResult := ParallelFlatMap(Range(1, 6), func(n int) Stream[int] {
			return RangeClosed(0, n)
		}, WithConcurrency(4), WithOrdered(true)).Collect()

		chunkedResult := ParallelFlatMap(Range(1, 6), func(n int) Stream[int] {
			return RangeClosed(0, n)
		}, WithConcurrency(4), WithOrdered(true), WithChunkSize(2)).Collect()

		expected := []int{0, 1, 0, 1, 2, 0, 1, 2, 3, 0, 1, 2, 3, 4, 0, 1, 2, 3, 4, 5}
		assert.Equal(t, expected, streamingResult, "ParallelFlatMap growing streaming should match expected")
		assert.Equal(t, expected, chunkedResult, "ParallelFlatMap growing chunked should match expected")
	})

	t.Run("DefaultOptionsUsesOrderedStreaming", func(t *testing.T) {
		t.Parallel()
		result := ParallelFlatMap(Of(1, 2, 3), func(n int) Stream[int] {
			return Of(n*10, n*10+1)
		}).Collect()
		assert.Equal(t, []int{10, 11, 20, 21, 30, 31}, result, "ParallelFlatMap default should be ordered streaming")
	})
}

// --- Prefetch Tests ---

func TestPrefetch(t *testing.T) {
	t.Parallel()
	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		result := Prefetch(Of(1, 2, 3, 4, 5), 3).Collect()
		assert.Equal(t, []int{1, 2, 3, 4, 5}, result, "Prefetch should not change element order/content")
	})

	t.Run("EmptyStream", func(t *testing.T) {
		t.Parallel()
		result := Prefetch(Empty[int](), 3).Collect()
		assert.Empty(t, result, "Prefetch on empty should be empty")
	})

	t.Run("SingleElement", func(t *testing.T) {
		t.Parallel()
		result := Prefetch(Of(42), 3).Collect()
		assert.Equal(t, []int{42}, result, "Prefetch single element should pass through")
	})

	t.Run("BufferSizeZeroDefaultsTo1", func(t *testing.T) {
		t.Parallel()
		result := Prefetch(Of(1, 2, 3), 0).Collect()
		assert.Equal(t, []int{1, 2, 3}, result, "Prefetch(0) should behave like buffer size 1")
	})

	t.Run("BufferSizeNegativeDefaultsTo1", func(t *testing.T) {
		t.Parallel()
		result := Prefetch(Of(1, 2, 3), -5).Collect()
		assert.Equal(t, []int{1, 2, 3}, result, "Prefetch(<0) should behave like buffer size 1")
	})

	t.Run("LargeBufferSize", func(t *testing.T) {
		t.Parallel()
		result := Prefetch(Of(1, 2, 3), 1000).Collect()
		assert.Equal(t, []int{1, 2, 3}, result, "Prefetch large buffer should preserve content")
	})

	t.Run("EarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := Prefetch(Of(1, 2, 3, 4, 5, 6, 7, 8, 9, 10), 5).Limit(3).Collect()
		assert.Equal(t, []int{1, 2, 3}, result, "Prefetch should respect Limit")
	})

	t.Run("WithTransformation", func(t *testing.T) {
		t.Parallel()
		result := Prefetch(Of(1, 2, 3, 4, 5), 2).
			Map(func(n int) int { return n * 2 }).
			Collect()
		assert.Equal(t, []int{2, 4, 6, 8, 10}, result, "Prefetch combined with Map should produce mapped sequence")
	})
}

// --- ParallelForEach Tests ---

func TestParallelForEach(t *testing.T) {
	t.Parallel()
	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		var sum atomic.Int64
		ParallelForEach(Of(1, 2, 3, 4, 5), func(n int) {
			sum.Add(int64(n))
		}, WithConcurrency(2))
		assert.Equal(t, int64(15), sum.Load(), "ParallelForEach should visit all elements exactly once")
	})

	t.Run("EmptyStream", func(t *testing.T) {
		t.Parallel()
		var count atomic.Int64
		ParallelForEach(Empty[int](), func(_ int) {
			count.Add(1)
		}, WithConcurrency(2))
		assert.Equal(t, int64(0), count.Load(), "ParallelForEach on empty should not invoke callback")
	})

	t.Run("SingleElement", func(t *testing.T) {
		t.Parallel()
		var count atomic.Int64
		ParallelForEach(Of(42), func(_ int) {
			count.Add(1)
		}, WithConcurrency(2))
		assert.Equal(t, int64(1), count.Load(), "ParallelForEach single element should invoke once")
	})

	t.Run("AllElementsProcessed", func(t *testing.T) {
		t.Parallel()
		seen := make(map[int]bool)
		var mu = &struct {
			seen map[int]bool
		}{seen: seen}
		var lock atomic.Bool

		ParallelForEach(Of(1, 2, 3, 4, 5), func(n int) {
			for !lock.CompareAndSwap(false, true) {
			}
			mu.seen[n] = true
			lock.Store(false)
		}, WithConcurrency(2))

		assert.Len(t, seen, 5, "ParallelForEach should visit exactly 5 unique elements")
		for i := 1; i <= 5; i++ {
			assert.True(t, seen[i], "Element %d should be processed", i)
		}
	})

	t.Run("LargeConcurrency", func(t *testing.T) {
		t.Parallel()
		var sum atomic.Int64
		ParallelForEach(Of(1, 2, 3), func(n int) {
			sum.Add(int64(n))
		}, WithConcurrency(100))
		assert.Equal(t, int64(6), sum.Load(), "ParallelForEach with large concurrency should still process all elements")
	})

	t.Run("DefaultOptions", func(t *testing.T) {
		t.Parallel()
		var sum atomic.Int64
		ParallelForEach(Of(1, 2, 3, 4, 5), func(n int) {
			sum.Add(int64(n))
		})
		assert.Equal(t, int64(15), sum.Load(), "ParallelForEach default options should process all elements")
	})
}

// --- ParallelReduce Tests ---

func TestParallelReduce(t *testing.T) {
	t.Parallel()
	t.Run("Sum", func(t *testing.T) {
		t.Parallel()
		result := ParallelReduce(Of(1, 2, 3, 4, 5), 0, func(a, b int) int {
			return a + b
		}, WithConcurrency(2))
		assert.Equal(t, 15, result, "ParallelReduce sum should equal 15")
	})

	t.Run("EmptyStreamReturnsIdentity", func(t *testing.T) {
		t.Parallel()
		result := ParallelReduce(Empty[int](), 100, func(a, b int) int {
			return a + b
		}, WithConcurrency(2))
		assert.Equal(t, 100, result, "ParallelReduce empty should return identity")
	})

	t.Run("SingleElement", func(t *testing.T) {
		t.Parallel()
		result := ParallelReduce(Of(42), 0, func(a, b int) int {
			return a + b
		}, WithConcurrency(2))
		assert.Equal(t, 42, result, "ParallelReduce single element should return that value")
	})

	t.Run("Product", func(t *testing.T) {
		t.Parallel()
		result := ParallelReduce(Of(1, 2, 3, 4, 5), 1, func(a, b int) int {
			return a * b
		}, WithConcurrency(2))
		assert.Equal(t, 120, result, "ParallelReduce product should equal 120")
	})

	t.Run("Max", func(t *testing.T) {
		t.Parallel()
		result := ParallelReduce(Of(3, 1, 4, 1, 5, 9, 2, 6), 0, func(a, b int) int {
			if a > b {
				return a
			}
			return b
		}, WithConcurrency(2))
		assert.Equal(t, 9, result, "ParallelReduce max should equal 9")
	})

	t.Run("SequentialFallbackForSmallInput", func(t *testing.T) {
		t.Parallel()
		result := ParallelReduce(Of(1, 2), 0, func(a, b int) int {
			return a + b
		}, WithConcurrency(100))
		assert.Equal(t, 3, result, "ParallelReduce should fall back to sequential for small input")
	})

	t.Run("LargeInput", func(t *testing.T) {
		t.Parallel()
		input := Range(1, 1001).Collect()
		result := ParallelReduce(FromSlice(input), 0, func(a, b int) int {
			return a + b
		}, WithConcurrency(4))
		assert.Equal(t, 500500, result, "ParallelReduce sum 1..1000 should be 500500") // Sum of 1 to 1000
	})

	t.Run("DefaultOptions", func(t *testing.T) {
		t.Parallel()
		result := ParallelReduce(Of(1, 2, 3, 4, 5), 0, func(a, b int) int {
			return a + b
		})
		assert.Equal(t, 15, result, "ParallelReduce default options should work")
	})
}

// --- ParallelCollect Tests ---

func TestParallelCollect(t *testing.T) {
	t.Parallel()
	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		result := ParallelCollect(Of(1, 2, 3, 4, 5), WithConcurrency(2))
		assert.Len(t, result, 5, "ParallelCollect should collect all elements")
		sort.Ints(result)
		assert.Equal(t, []int{1, 2, 3, 4, 5}, result, "ParallelCollect values should match input set")
	})

	t.Run("EmptyStream", func(t *testing.T) {
		t.Parallel()
		result := ParallelCollect(Empty[int](), WithConcurrency(2))
		assert.Empty(t, result, "ParallelCollect on empty should be empty")
	})

	t.Run("SingleElement", func(t *testing.T) {
		t.Parallel()
		result := ParallelCollect(Of(42), WithConcurrency(2))
		assert.Equal(t, []int{42}, result, "ParallelCollect single element should collect one value")
	})

	t.Run("LargeInput", func(t *testing.T) {
		t.Parallel()
		input := Range(1, 101).Collect()
		result := ParallelCollect(FromSlice(input), WithConcurrency(4))
		assert.Len(t, result, 100, "ParallelCollect should collect 100 elements")
		sort.Ints(result)
		assert.Equal(t, input, result, "ParallelCollect values should match input after sort")
	})

	t.Run("DefaultOptions", func(t *testing.T) {
		t.Parallel()
		result := ParallelCollect(Of(1, 2, 3))
		assert.Len(t, result, 3, "ParallelCollect default options should collect all")
	})
}

// --- Integration Tests ---

func TestParallelChainOperations(t *testing.T) {
	t.Parallel()
	t.Run("ParallelMapThenParallelFilter", func(t *testing.T) {
		t.Parallel()
		result := ParallelFilter(
			ParallelMap(Of(1, 2, 3, 4, 5), func(n int) int {
				return n * 2
			}, WithOrdered(true)),
			func(n int) bool {
				return n > 4
			},
			WithOrdered(true),
		).Collect()
		assert.Equal(t, []int{6, 8, 10}, result, "ParallelMap then Filter ordered should compose")
	})

	t.Run("PrefetchThenParallelMap", func(t *testing.T) {
		t.Parallel()
		result := ParallelMap(
			Prefetch(Of(1, 2, 3, 4, 5), 3),
			func(n int) int { return n * 2 },
			WithOrdered(true),
		).Collect()
		assert.Equal(t, []int{2, 4, 6, 8, 10}, result, "Prefetch then ParallelMap ordered should compose")
	})

	t.Run("ParallelFlatMapThenReduce", func(t *testing.T) {
		t.Parallel()
		result := ParallelReduce(
			ParallelFlatMap(Of(1, 2, 3), func(n int) Stream[int] {
				return Of(n, n)
			}, WithOrdered(true)),
			0,
			func(a, b int) int { return a + b },
		)
		assert.Equal(t, 12, result, "ParallelFlatMap then Reduce should sum to 12") // (1+1) + (2+2) + (3+3)
	})
}

func TestParallelEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("OrderedStreamingEarlyExitMidChunk", func(t *testing.T) {
		t.Parallel()
		// Test early termination in the middle of processing to cover done channel branches
		result := ParallelMap(Range(1, 100), func(n int) int {
			return n * 2
		}, WithConcurrency(4), WithOrdered(true)).Limit(5).Collect()
		assert.Equal(t, []int{2, 4, 6, 8, 10}, result, "Should stop early with correct values")
	})

	t.Run("UnorderedEarlyExitMidProcessing", func(t *testing.T) {
		t.Parallel()
		// Test early termination for unordered parallel map
		result := ParallelMap(Range(1, 100), func(n int) int {
			return n * 2
		}, WithConcurrency(4), WithOrdered(false)).Limit(5).Collect()
		assert.Len(t, result, 5, "Should stop early with 5 elements")
	})

	t.Run("ChunkedEarlyExitDuringChunkProcessing", func(t *testing.T) {
		t.Parallel()
		// Test early termination while a chunk is still being processed
		result := ParallelMap(Range(1, 100), func(n int) int {
			return n * 2
		}, WithConcurrency(4), WithOrdered(true), WithChunkSize(10)).Limit(3).Collect()
		assert.Equal(t, []int{2, 4, 6}, result, "Should stop early with correct values in chunked mode")
	})

	t.Run("FilterOrderedStreamingEarlyExit", func(t *testing.T) {
		t.Parallel()
		result := ParallelFilter(Range(1, 100), func(n int) bool {
			return n%2 == 0
		}, WithConcurrency(4), WithOrdered(true)).Limit(3).Collect()
		assert.Equal(t, []int{2, 4, 6}, result, "Ordered filter should stop early correctly")
	})

	t.Run("FilterUnorderedEarlyExit", func(t *testing.T) {
		t.Parallel()
		result := ParallelFilter(Range(1, 100), func(n int) bool {
			return n%2 == 0
		}, WithConcurrency(4), WithOrdered(false)).Limit(3).Collect()
		assert.Len(t, result, 3, "Unordered filter should stop early with 3 elements")
	})

	t.Run("FilterChunkedEarlyExit", func(t *testing.T) {
		t.Parallel()
		result := ParallelFilter(Range(1, 100), func(n int) bool {
			return n%2 == 0
		}, WithConcurrency(4), WithOrdered(true), WithChunkSize(10)).Limit(3).Collect()
		assert.Equal(t, []int{2, 4, 6}, result, "Chunked filter should stop early correctly")
	})

	t.Run("FlatMapOrderedStreamingEarlyExit", func(t *testing.T) {
		t.Parallel()
		result := ParallelFlatMap(Range(1, 50), func(n int) Stream[int] {
			return Of(n*10, n*10+1, n*10+2)
		}, WithConcurrency(4), WithOrdered(true)).Limit(5).Collect()
		assert.Equal(t, []int{10, 11, 12, 20, 21}, result, "Ordered flatmap streaming should stop early correctly")
	})

	t.Run("FlatMapUnorderedEarlyExit", func(t *testing.T) {
		t.Parallel()
		result := ParallelFlatMap(Range(1, 50), func(n int) Stream[int] {
			return Of(n*10, n*10+1, n*10+2)
		}, WithConcurrency(4), WithOrdered(false)).Limit(5).Collect()
		assert.Len(t, result, 5, "Unordered flatmap should stop early with 5 elements")
	})

	t.Run("FlatMapChunkedEarlyExit", func(t *testing.T) {
		t.Parallel()
		result := ParallelFlatMap(Range(1, 50), func(n int) Stream[int] {
			return Of(n*10, n*10+1, n*10+2)
		}, WithConcurrency(4), WithOrdered(true), WithChunkSize(5)).Limit(5).Collect()
		assert.Equal(t, []int{10, 11, 12, 20, 21}, result, "Chunked flatmap should stop early correctly")
	})

	t.Run("PrefetchEarlyExitDuringBuffer", func(t *testing.T) {
		t.Parallel()
		result := Prefetch(Range(1, 100), 10).Limit(3).Collect()
		assert.Equal(t, []int{1, 2, 3}, result, "Prefetch should stop early correctly")
	})

	t.Run("VerySmallBufferSize", func(t *testing.T) {
		t.Parallel()
		result := ParallelMap(Of(1, 2, 3, 4, 5), func(n int) int {
			return n * 2
		}, WithConcurrency(2), WithBufferSize(1), WithOrdered(true)).Collect()
		assert.Equal(t, []int{2, 4, 6, 8, 10}, result, "Small buffer should not affect correctness")
	})

	t.Run("ConcurrencyEqualsInputSize", func(t *testing.T) {
		t.Parallel()
		result := ParallelMap(Of(1, 2, 3), func(n int) int {
			return n * 2
		}, WithConcurrency(3), WithOrdered(true)).Collect()
		assert.Equal(t, []int{2, 4, 6}, result, "Concurrency equal to input size should work")
	})

	t.Run("ConcurrencyExceedsInputSize", func(t *testing.T) {
		t.Parallel()
		result := ParallelMap(Of(1, 2, 3), func(n int) int {
			return n * 2
		}, WithConcurrency(10), WithOrdered(true)).Collect()
		assert.Equal(t, []int{2, 4, 6}, result, "Concurrency larger than input size should work")
	})

	t.Run("ChunkSizeOf1", func(t *testing.T) {
		t.Parallel()
		result := ParallelFlatMap(Of(1, 2, 3), func(n int) Stream[int] {
			return Of(n*10, n*10+1)
		}, WithConcurrency(2), WithOrdered(true), WithChunkSize(1)).Collect()
		assert.Equal(t, []int{10, 11, 20, 21, 30, 31}, result, "ChunkSize 1 should still preserve order")
	})
}

// --- ParallelForEachCtx Tests ---

func TestParallelForEachCtx(t *testing.T) {
	t.Parallel()

	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		var sum atomic.Int64
		err := ParallelForEachCtx(testCtx(), Of(1, 2, 3, 4, 5), func(_ context.Context, n int) {
			sum.Add(int64(n))
		}, WithConcurrency(2))
		assert.NoError(t, err, "ParallelForEachCtx should not error")
		assert.Equal(t, int64(15), sum.Load(), "ParallelForEachCtx should process all elements")
	})

	t.Run("Empty", func(t *testing.T) {
		t.Parallel()
		var count atomic.Int64
		err := ParallelForEachCtx(testCtx(), Empty[int](), func(_ context.Context, _ int) {
			count.Add(1)
		})
		assert.NoError(t, err, "ParallelForEachCtx on empty should not error")
		assert.Equal(t, int64(0), count.Load(), "ParallelForEachCtx on empty should not call action")
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		var count atomic.Int64

		// Cancel immediately
		cancel()

		err := ParallelForEachCtx(ctx, Range(1, 1000), func(_ context.Context, _ int) {
			count.Add(1)
		}, WithConcurrency(4))

		assert.Error(t, err, "ParallelForEachCtx should return error on cancelled context")
		assert.Equal(t, context.Canceled, err, "Error should be context.Canceled")
	})

	t.Run("ContextTimeout", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		var count atomic.Int64
		err := ParallelForEachCtx(ctx, Range(1, 10000), func(_ context.Context, _ int) {
			time.Sleep(10 * time.Millisecond) // Slow operation
			count.Add(1)
		}, WithConcurrency(2))

		assert.Error(t, err, "ParallelForEachCtx should return error on timeout")
	})

	t.Run("LargeInput", func(t *testing.T) {
		t.Parallel()
		var sum atomic.Int64
		err := ParallelForEachCtx(testCtx(), Range(1, 101), func(_ context.Context, n int) {
			sum.Add(int64(n))
		}, WithConcurrency(4))
		assert.NoError(t, err, "ParallelForEachCtx should not error on large input")
		assert.Equal(t, int64(5050), sum.Load(), "ParallelForEachCtx should sum 1-100 correctly")
	})
}

// --- ParallelMapCtx Tests ---

func TestParallelMapCtx(t *testing.T) {
	t.Parallel()

	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		result := ParallelMapCtx(testCtx(), Of(1, 2, 3), func(_ context.Context, n int) int {
			return n * 2
		}, WithOrdered(true)).Collect()
		assert.Equal(t, []int{2, 4, 6}, result, "ParallelMapCtx should transform elements")
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		result := ParallelMapCtx(ctx, Range(1, 1000), func(_ context.Context, n int) int {
			return n * 2
		}, WithConcurrency(4), WithOrdered(true)).Collect()

		// Should return early due to cancellation
		assert.Less(t, len(result), 1000, "ParallelMapCtx should stop early on cancellation")
	})

	t.Run("Unordered", func(t *testing.T) {
		t.Parallel()
		result := ParallelMapCtx(testCtx(), Of(1, 2, 3, 4, 5), func(_ context.Context, n int) int {
			return n * 2
		}, WithOrdered(false), WithConcurrency(2)).Collect()
		assert.Len(t, result, 5, "ParallelMapCtx unordered should return all elements")
	})

	t.Run("Chunked", func(t *testing.T) {
		t.Parallel()
		result := ParallelMapCtx(testCtx(), Of(1, 2, 3, 4, 5), func(_ context.Context, n int) int {
			return n * 2
		}, WithOrdered(true), WithChunkSize(2)).Collect()
		assert.Equal(t, []int{2, 4, 6, 8, 10}, result, "ParallelMapCtx chunked should preserve order")
	})
}

// --- ParallelFilterCtx Tests ---

func TestParallelFilterCtx(t *testing.T) {
	t.Parallel()

	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		result := ParallelFilterCtx(testCtx(), Of(1, 2, 3, 4, 5), func(_ context.Context, n int) bool {
			return n%2 == 0
		}, WithOrdered(true)).Collect()
		assert.Equal(t, []int{2, 4}, result, "ParallelFilterCtx should filter elements")
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		result := ParallelFilterCtx(ctx, Range(1, 1000), func(_ context.Context, n int) bool {
			return n%2 == 0
		}, WithConcurrency(4), WithOrdered(true)).Collect()

		// Should return early due to cancellation
		assert.Less(t, len(result), 500, "ParallelFilterCtx should stop early on cancellation")
	})

	t.Run("Unordered", func(t *testing.T) {
		t.Parallel()
		result := ParallelFilterCtx(testCtx(), Of(1, 2, 3, 4, 5, 6), func(_ context.Context, n int) bool {
			return n%2 == 0
		}, WithOrdered(false), WithConcurrency(2)).Collect()
		assert.Len(t, result, 3, "ParallelFilterCtx unordered should return filtered elements")
	})

	t.Run("Chunked", func(t *testing.T) {
		t.Parallel()
		result := ParallelFilterCtx(testCtx(), Of(1, 2, 3, 4, 5, 6), func(_ context.Context, n int) bool {
			return n%2 == 0
		}, WithOrdered(true), WithChunkSize(2)).Collect()
		assert.Equal(t, []int{2, 4, 6}, result, "ParallelFilterCtx chunked should preserve order")
	})
}

// --- ParallelFlatMapCtx Tests ---

func TestParallelFlatMapCtx(t *testing.T) {
	t.Parallel()

	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		result := ParallelFlatMapCtx(testCtx(), Of(1, 2, 3), func(_ context.Context, n int) Stream[int] {
			return Of(n, n*10)
		}, WithOrdered(true)).Collect()
		assert.Equal(t, []int{1, 10, 2, 20, 3, 30}, result, "ParallelFlatMapCtx should flatten elements")
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		result := ParallelFlatMapCtx(ctx, Range(1, 100), func(_ context.Context, n int) Stream[int] {
			return Of(n, n*10)
		}, WithConcurrency(4), WithOrdered(true)).Collect()

		// Should return early due to cancellation
		assert.Less(t, len(result), 200, "ParallelFlatMapCtx should stop early on cancellation")
	})

	t.Run("Unordered", func(t *testing.T) {
		t.Parallel()
		result := ParallelFlatMapCtx(testCtx(), Of(1, 2, 3), func(_ context.Context, n int) Stream[int] {
			return Of(n, n*10)
		}, WithOrdered(false), WithConcurrency(2)).Collect()
		assert.Len(t, result, 6, "ParallelFlatMapCtx unordered should return all elements")
	})

	t.Run("Chunked", func(t *testing.T) {
		t.Parallel()
		result := ParallelFlatMapCtx(testCtx(), Of(1, 2, 3), func(_ context.Context, n int) Stream[int] {
			return Of(n, n*10)
		}, WithOrdered(true), WithChunkSize(2)).Collect()
		assert.Equal(t, []int{1, 10, 2, 20, 3, 30}, result, "ParallelFlatMapCtx chunked should preserve order")
	})
}
