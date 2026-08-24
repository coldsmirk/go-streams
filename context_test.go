package streams

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"testing/synctest"
)

func TestWithContext(t *testing.T) {
	t.Parallel()
	t.Run("NormalCompletion", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		result := WithContext(ctx, Of(1, 2, 3, 4, 5)).Collect()
		assert.Equal(t, []int{1, 2, 3, 4, 5}, result, "WithContext should pass through all elements")
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		count := 0
		s := Generate(func() int {
			count++
			if count == 3 {
				cancel()
			}
			return count
		})
		result := WithContext(ctx, s).Limit(10).Collect()
		assert.LessOrEqual(t, len(result), 4, "WithContext should stop when context is cancelled") // May get 3 or 4 elements depending on timing
	})

	t.Run("AlreadyCancelledContext", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		result := WithContext(ctx, Of(1, 2, 3)).Collect()
		assert.Empty(t, result, "WithContext on already-cancelled context should yield empty")
	})

	t.Run("EarlyTerminationByConsumer", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		result := WithContext(ctx, Of(1, 2, 3, 4, 5)).Limit(2).Collect()
		assert.Equal(t, []int{1, 2}, result, "WithContext should respect consumer early termination")
	})
}

func TestWithContext2(t *testing.T) {
	t.Parallel()
	t.Run("NormalCompletion", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		s2 := FromMap(map[string]int{"a": 1, "b": 2})
		result := WithContext2(ctx, s2).CollectPairs()
		assert.Len(t, result, 2, "WithContext2 should collect all pairs")
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		s2 := FromMap(map[string]int{"a": 1, "b": 2, "c": 3})
		result := WithContext2(ctx, s2).CollectPairs()
		assert.Empty(t, result, "WithContext2 on cancelled context should yield empty")
	})

	t.Run("EarlyTerminationByConsumer", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		s2 := FromMap(map[string]int{"a": 1, "b": 2, "c": 3})
		result := WithContext2(ctx, s2).Limit(1).CollectPairs()
		assert.Len(t, result, 1, "WithContext2 should respect Limit")
	})
}

func TestGenerateCtx(t *testing.T) {
	t.Parallel()
	t.Run("NormalGenerationWithLimit", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		counter := 0
		result := GenerateCtx(ctx, func() int {
			counter++
			return counter
		}).Limit(5).Collect()
		assert.Equal(t, []int{1, 2, 3, 4, 5}, result, "GenerateCtx should generate increasing integers")
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		counter := 0
		s := GenerateCtx(ctx, func() int {
			counter++
			if counter >= 3 {
				cancel()
			}
			return counter
		})
		result := s.Limit(10).Collect()
		assert.LessOrEqual(t, len(result), 4, "GenerateCtx should stop when context cancelled")
	})

	t.Run("AlreadyCancelled", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		result := GenerateCtx(ctx, func() int { return 1 }).Limit(5).Collect()
		assert.Empty(t, result, "GenerateCtx with already-cancelled context should yield empty")
	})

	t.Run("EarlyTerminationByConsumer", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		result := GenerateCtx(ctx, func() int { return 42 }).Limit(3).Collect()
		assert.Equal(t, []int{42, 42, 42}, result, "GenerateCtx should respect Limit")
	})
}

func TestIterateCtx(t *testing.T) {
	t.Parallel()
	t.Run("NormalIterationWithLimit", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		result := IterateCtx(ctx, 1, func(n int) int { return n * 2 }).Limit(5).Collect()
		assert.Equal(t, []int{1, 2, 4, 8, 16}, result, "IterateCtx should apply function iteratively")
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		count := 0
		result := IterateCtx(ctx, 1, func(n int) int {
			count++
			if count >= 3 {
				cancel()
			}
			return n + 1
		}).Limit(10).Collect()
		assert.LessOrEqual(t, len(result), 5, "IterateCtx should stop when context cancelled")
	})

	t.Run("AlreadyCancelled", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		result := IterateCtx(ctx, 1, func(n int) int { return n + 1 }).Limit(5).Collect()
		assert.Empty(t, result, "IterateCtx with already-cancelled context should be empty")
	})
}

func TestRangeCtx(t *testing.T) {
	t.Parallel()
	t.Run("NormalRange", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		result := RangeCtx(ctx, 1, 6).Collect()
		assert.Equal(t, []int{1, 2, 3, 4, 5}, result, "RangeCtx should produce [1..5]")
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		count := 0
		var result []int
		for v := range RangeCtx(ctx, 1, 100).seq {
			result = append(result, v)
			count++
			if count >= 3 {
				cancel()
			}
		}
		assert.LessOrEqual(t, len(result), 4, "RangeCtx should stop when context cancelled")
	})

	t.Run("AlreadyCancelled", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		result := RangeCtx(ctx, 1, 10).Collect()
		assert.Empty(t, result, "RangeCtx with already-cancelled context should be empty")
	})

	t.Run("EmptyRange", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		result := RangeCtx(ctx, 5, 5).Collect()
		assert.Empty(t, result, "RangeCtx with start==end should be empty")
	})
}

func TestFromChannelCtx(t *testing.T) {
	t.Parallel()
	t.Run("NormalChannelRead", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		ch := make(chan int, 5)
		for i := 1; i <= 5; i++ {
			ch <- i
		}
		close(ch)
		result := FromChannelCtx(ctx, ch).Collect()
		assert.Equal(t, []int{1, 2, 3, 4, 5}, result, "FromChannelCtx should drain buffered channel")
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		ch := make(chan int, 10)
		for i := 1; i <= 10; i++ {
			ch <- i
		}

		var result []int
		for v := range FromChannelCtx(ctx, ch).seq {
			result = append(result, v)
			if len(result) >= 3 {
				cancel()
			}
		}
		assert.LessOrEqual(t, len(result), 4, "FromChannelCtx should stop when context cancelled")
	})

	t.Run("AlreadyCancelled", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			// Use an unbuffered channel that blocks - cancellation should prevent any reads
			ch := make(chan int)
			go func() {
				// Non-blocking attempt to send; immediately close to avoid lingering timers/goroutines in synctest.
				select {
				case ch <- 1:
				default:
				}
				close(ch)
			}()
			result := FromChannelCtx(ctx, ch).Collect()
			// With cancellation, should get no elements (cancelled before first read)
			assert.Empty(t, result, "AlreadyCancelled should yield empty result")
		})
	})

	t.Run("EmptyChannel", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		ch := make(chan int)
		close(ch)
		result := FromChannelCtx(ctx, ch).Collect()
		assert.Empty(t, result, "EmptyChannel should yield empty result")
	})

	t.Run("ContextCancellationWhileWaitingForChannel", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			// This test covers the case where context is cancelled while blocked on channel receive
			ctx, cancel := context.WithCancel(context.Background())
			ch := make(chan int, 2)
			ch <- 1
			ch <- 2
			// Don't close channel - we want to test cancellation while waiting

			var result []int
			done := make(chan struct{})
			go func() {
				for v := range FromChannelCtx(ctx, ch).seq {
					result = append(result, v)
				}
				close(done)
			}()

			// Wait a bit to ensure we're blocked waiting for channel
			time.Sleep(50 * time.Millisecond)
			cancel()
			<-done

			// Should have received the 2 values that were in buffer
			assert.Equal(t, []int{1, 2}, result, "Should drain buffered values before cancellation")
		})
	})
}

func TestFromReaderLinesCtx(t *testing.T) {
	t.Parallel()
	t.Run("NormalRead", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		reader := strings.NewReader("line1\nline2\nline3")
		result := FromReaderLinesCtx(ctx, reader).Collect()
		assert.Equal(t, []string{"line1", "line2", "line3"}, result, "FromReaderLinesCtx should split lines")
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		reader := strings.NewReader("line1\nline2\nline3\nline4\nline5")
		var result []string
		for v := range FromReaderLinesCtx(ctx, reader).seq {
			result = append(result, v)
			if len(result) >= 2 {
				cancel()
			}
		}
		assert.LessOrEqual(t, len(result), 3, "FromReaderLinesCtx should stop when context cancelled")
	})

	t.Run("AlreadyCancelled", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		reader := strings.NewReader("line1\nline2")
		result := FromReaderLinesCtx(ctx, reader).Collect()
		assert.Empty(t, result, "FromReaderLinesCtx with already-cancelled context should be empty")
	})

	t.Run("EmptyReader", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		reader := strings.NewReader("")
		result := FromReaderLinesCtx(ctx, reader).Collect()
		assert.Empty(t, result, "FromReaderLinesCtx empty input should be empty")
	})
}

func TestCollectCtx(t *testing.T) {
	t.Parallel()
	t.Run("NormalCollection", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		result, err := CollectCtx(ctx, Of(1, 2, 3, 4, 5))
		assert.NoError(t, err, "CollectCtx should not error on normal stream")
		assert.Equal(t, []int{1, 2, 3, 4, 5}, result, "CollectCtx should collect all elements")
	})

	t.Run("ContextCancellationDuringCollection", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		counter := 0
		s := Generate(func() int {
			counter++
			if counter == 3 {
				cancel()
			}
			return counter
		}).Limit(10)

		result, err := CollectCtx(ctx, s)
		assert.Error(t, err, "CollectCtx should return context error when cancelled")
		assert.Equal(t, context.Canceled, err, "CollectCtx should return context.Canceled")
		assert.NotEmpty(t, result, "CollectCtx should return partial results when cancelled")
	})

	t.Run("AlreadyCancelled", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		result, err := CollectCtx(ctx, Of(1, 2, 3))
		assert.Error(t, err, "CollectCtx should error on already-cancelled context")
		assert.Empty(t, result, "CollectCtx should return no results for already-cancelled context")
	})

	t.Run("EmptyStream", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		result, err := CollectCtx(ctx, Empty[int]())
		assert.NoError(t, err, "CollectCtx empty should not error")
		assert.Empty(t, result, "CollectCtx empty should return empty slice")
	})
}

func TestForEachCtx(t *testing.T) {
	t.Parallel()
	t.Run("NormalIteration", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		var sum int
		err := ForEachCtx(ctx, Of(1, 2, 3, 4, 5), func(n int) { sum += n })
		assert.NoError(t, err, "ForEachCtx should not error on normal iteration")
		assert.Equal(t, 15, sum, "ForEachCtx should aggregate all elements")
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		var sum atomic.Int32
		counter := 0
		s := Generate(func() int {
			counter++
			return counter
		}).Limit(10)

		err := ForEachCtx(ctx, s, func(n int) {
			sum.Add(int32(n))
			if n >= 3 {
				cancel()
			}
		})
		assert.Error(t, err, "ForEachCtx should return error on cancellation")
		assert.Equal(t, context.Canceled, err, "ForEachCtx error should be context.Canceled")
	})

	t.Run("AlreadyCancelled", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		var count int
		err := ForEachCtx(ctx, Of(1, 2, 3), func(n int) { count++ })
		assert.Error(t, err, "ForEachCtx should error on already-cancelled context")
		assert.Equal(t, 0, count, "ForEachCtx should not execute callback on cancelled context")
	})

	t.Run("WithErrorFunc", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		var collected []int
		expectedErr := errors.New("test error")

		err := ForEachCtx(ctx, Of(1, 2, 3, 4, 5), func(n int) error {
			if n == 3 {
				return expectedErr
			}
			collected = append(collected, n)
			return nil
		})

		assert.Error(t, err, "ForEachCtx should return error from action")
		assert.Equal(t, expectedErr, err, "ForEachCtx should return the exact error")
		assert.Equal(t, []int{1, 2}, collected, "ForEachCtx should stop at first error")
	})

	t.Run("WithErrorFuncSuccess", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		var sum int

		err := ForEachCtx(ctx, Of(1, 2, 3, 4, 5), func(n int) error {
			sum += n
			return nil
		})

		assert.NoError(t, err, "ForEachCtx should not error on successful iteration")
		assert.Equal(t, 15, sum, "ForEachCtx should process all elements")
	})

	t.Run("WithErrorFuncAndContextCancellation", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		var collected []int

		err := ForEachCtx(ctx, Of(1, 2, 3, 4, 5), func(n int) error {
			collected = append(collected, n)
			if n == 2 {
				cancel()
			}
			return nil
		})

		assert.Error(t, err, "ForEachCtx should return context error")
		assert.Equal(t, context.Canceled, err, "ForEachCtx should return context.Canceled")
	})
}

func TestReduceCtx(t *testing.T) {
	t.Parallel()
	t.Run("NormalReduction", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		result, err := ReduceCtx(ctx, Of(1, 2, 3, 4, 5), 0, func(a, b int) int { return a + b })
		assert.NoError(t, err, "ReduceCtx should not error on normal reduction")
		assert.Equal(t, 15, result, "ReduceCtx should reduce to sum=15")
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		counter := 0
		s := Generate(func() int {
			counter++
			if counter >= 3 {
				cancel()
			}
			return counter
		}).Limit(10)

		result, err := ReduceCtx(ctx, s, 0, func(a, b int) int { return a + b })
		assert.Error(t, err, "ReduceCtx should error when cancelled")
		assert.Greater(t, result, 0, "ReduceCtx should return partial result when cancelled") // Partial result
	})

	t.Run("AlreadyCancelled", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		result, err := ReduceCtx(ctx, Of(1, 2, 3), 0, func(a, b int) int { return a + b })
		assert.Error(t, err, "ReduceCtx should error on already-cancelled context")
		assert.Equal(t, 0, result, "ReduceCtx should return identity for already-cancelled context")
	})

	t.Run("EmptyStream", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		result, err := ReduceCtx(ctx, Empty[int](), 100, func(a, b int) int { return a + b })
		assert.NoError(t, err, "ReduceCtx empty should not error")
		assert.Equal(t, 100, result, "ReduceCtx empty should return identity")
	})

	t.Run("WithErrorFunc", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		expectedErr := errors.New("reduction error")

		result, err := ReduceCtx(ctx, Of(1, 2, 3, 4, 5), 0, func(a, b int) (int, error) {
			if b == 3 {
				return a, expectedErr
			}
			return a + b, nil
		})

		assert.Error(t, err, "ReduceCtx should return error from reducer")
		assert.Equal(t, expectedErr, err, "ReduceCtx should return the exact error")
		assert.Equal(t, 3, result, "ReduceCtx should return partial result (0+1+2=3)")
	})

	t.Run("WithErrorFuncSuccess", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		result, err := ReduceCtx(ctx, Of(1, 2, 3, 4, 5), 0, func(a, b int) (int, error) {
			return a + b, nil
		})

		assert.NoError(t, err, "ReduceCtx should not error on successful reduction")
		assert.Equal(t, 15, result, "ReduceCtx should reduce to sum=15")
	})

	t.Run("WithErrorFuncAndContextCancellation", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		counter := 0

		s := Generate(func() int {
			counter++
			if counter == 3 {
				cancel()
			}
			return counter
		}).Limit(10)

		result, err := ReduceCtx(ctx, s, 0, func(a, b int) (int, error) {
			return a + b, nil
		})

		assert.Error(t, err, "ReduceCtx should return context error")
		assert.Equal(t, context.Canceled, err, "ReduceCtx should return context.Canceled")
		assert.Greater(t, result, 0, "ReduceCtx should return partial result")
	})

	t.Run("WithErrorFuncReturningError", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		expectedErr := errors.New("division by zero")

		result, err := ReduceCtx(ctx, Of(10, 5, 0, 2), 100, func(a, b int) (int, error) {
			if b == 0 {
				return a, expectedErr
			}
			return a / b, nil
		})

		assert.Error(t, err, "ReduceCtx should return error")
		assert.Equal(t, expectedErr, err, "ReduceCtx should return division error")
		assert.Equal(t, 2, result, "ReduceCtx should return partial result (100/10/5=2)")
	})
}

func TestFindFirstCtx(t *testing.T) {
	t.Parallel()
	t.Run("FoundElement", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		result, err := FindFirstCtx(ctx, Of(1, 2, 3, 4, 5), func(n int) bool { return n > 3 })
		assert.NoError(t, err, "FindFirstCtx should not error when found")
		assert.True(t, result.IsPresent(), "FindFirstCtx should return present result")
		assert.Equal(t, 4, result.Get(), "FindFirstCtx should find first >3")
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		result, err := FindFirstCtx(ctx, Of(1, 2, 3), func(n int) bool { return n > 10 })
		assert.NoError(t, err, "FindFirstCtx should not error when not found")
		assert.False(t, result.IsPresent(), "FindFirstCtx should be empty when not found")
	})

	t.Run("ContextCancellationBeforeFound", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		counter := 0
		s := Generate(func() int {
			counter++
			if counter >= 3 {
				cancel()
			}
			return counter
		}).Limit(10)

		result, err := FindFirstCtx(ctx, s, func(n int) bool { return n > 100 })
		assert.Error(t, err, "FindFirstCtx should error when cancelled")
		assert.False(t, result.IsPresent(), "FindFirstCtx should be empty on cancellation")
	})

	t.Run("AlreadyCancelled", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		result, err := FindFirstCtx(ctx, Of(1, 2, 3), func(n int) bool { return true })
		assert.Error(t, err, "FindFirstCtx should error on already-cancelled context")
		assert.False(t, result.IsPresent(), "FindFirstCtx should be empty on already-cancelled context")
	})
}

func TestAnyMatchCtx(t *testing.T) {
	t.Parallel()
	t.Run("MatchFound", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		result, err := AnyMatchCtx(ctx, Of(1, 2, 3, 4, 5), func(n int) bool { return n > 3 })
		assert.NoError(t, err, "AnyMatchCtx should not error on match")
		assert.True(t, result, "AnyMatchCtx should return true when any match")
	})

	t.Run("NoMatch", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		result, err := AnyMatchCtx(ctx, Of(1, 2, 3), func(n int) bool { return n > 10 })
		assert.NoError(t, err, "AnyMatchCtx should not error on no match")
		assert.False(t, result, "AnyMatchCtx should return false when no match")
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		counter := 0
		s := Generate(func() int {
			counter++
			if counter >= 3 {
				cancel()
			}
			return counter
		}).Limit(10)

		result, err := AnyMatchCtx(ctx, s, func(n int) bool { return n > 100 })
		assert.Error(t, err, "AnyMatchCtx should error when cancelled")
		assert.False(t, result, "AnyMatchCtx should return false on cancellation")
	})

	t.Run("AlreadyCancelled", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		result, err := AnyMatchCtx(ctx, Of(1, 2, 3), func(n int) bool { return true })
		assert.Error(t, err, "AnyMatchCtx should error on already-cancelled context")
		assert.False(t, result, "AnyMatchCtx should return false on already-cancelled context")
	})
}

func TestAllMatchCtx(t *testing.T) {
	t.Parallel()
	t.Run("AllMatch", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		result, err := AllMatchCtx(ctx, Of(2, 4, 6, 8), func(n int) bool { return n%2 == 0 })
		assert.NoError(t, err, "AllMatchCtx should not error when all match")
		assert.True(t, result, "AllMatchCtx should return true when all match")
	})

	t.Run("NotAllMatch", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		result, err := AllMatchCtx(ctx, Of(2, 3, 4), func(n int) bool { return n%2 == 0 })
		assert.NoError(t, err, "AllMatchCtx should not error when some don't match")
		assert.False(t, result, "AllMatchCtx should return false when not all match")
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		counter := 0
		s := Generate(func() int {
			counter++
			if counter >= 3 {
				cancel()
			}
			return counter * 2 // All even
		}).Limit(10)

		result, err := AllMatchCtx(ctx, s, func(n int) bool { return n%2 == 0 })
		assert.Error(t, err, "AllMatchCtx should error on cancellation")
		assert.False(t, result, "AllMatchCtx should return false on cancellation")
	})

	t.Run("AlreadyCancelled", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		result, err := AllMatchCtx(ctx, Of(2, 4, 6), func(n int) bool { return n%2 == 0 })
		assert.Error(t, err, "AllMatchCtx should error on already-cancelled context")
		assert.False(t, result, "AllMatchCtx should return false on already-cancelled context")
	})

	t.Run("EmptyStream", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		result, err := AllMatchCtx(ctx, Empty[int](), func(n int) bool { return false })
		assert.NoError(t, err, "AllMatchCtx empty should not error")
		assert.True(t, result, "AllMatchCtx empty should be true (vacuous truth)")
	})
}

func TestCountCtx(t *testing.T) {
	t.Parallel()
	t.Run("NormalCount", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		result, err := CountCtx(ctx, Of(1, 2, 3, 4, 5))
		assert.NoError(t, err, "CountCtx should not error on normal stream")
		assert.Equal(t, 5, result, "CountCtx should count all elements")
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		counter := 0
		s := Generate(func() int {
			counter++
			if counter >= 3 {
				cancel()
			}
			return counter
		}).Limit(10)

		result, err := CountCtx(ctx, s)
		assert.Error(t, err, "CountCtx should error on cancellation")
		assert.Greater(t, result, 0, "CountCtx should return partial count on cancellation") // Partial count
	})

	t.Run("AlreadyCancelled", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		result, err := CountCtx(ctx, Of(1, 2, 3))
		assert.Error(t, err, "CountCtx should error on already-cancelled context")
		assert.Equal(t, 0, result, "CountCtx should return 0 on already-cancelled context")
	})

	t.Run("EmptyStream", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		result, err := CountCtx(ctx, Empty[int]())
		assert.NoError(t, err, "CountCtx empty should not error")
		assert.Equal(t, 0, result, "CountCtx empty should return 0")
	})
}

func TestFilterCtx(t *testing.T) {
	t.Parallel()
	t.Run("NormalFilter", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		result := FilterCtx(ctx, Of(1, 2, 3, 4, 5), func(n int) bool { return n%2 == 0 }).Collect()
		assert.Equal(t, []int{2, 4}, result, "FilterCtx should keep only even numbers")
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		count := 0
		var result []int
		for v := range FilterCtx(ctx, Range(1, 100), func(n int) bool {
			count++
			if count >= 5 {
				cancel()
			}
			return n%2 == 0
		}).seq {
			result = append(result, v)
		}
		assert.LessOrEqual(t, len(result), 3, "FilterCtx should stop when context cancelled")
	})

	t.Run("AlreadyCancelled", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		result := FilterCtx(ctx, Of(1, 2, 3, 4, 5), func(n int) bool { return true }).Collect()
		assert.Empty(t, result, "FilterCtx with already-cancelled context should be empty")
	})

	t.Run("FilterAllOut", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		result := FilterCtx(ctx, Of(1, 3, 5, 7), func(n int) bool { return n%2 == 0 }).Collect()
		assert.Empty(t, result, "FilterCtx should drop all when predicate false for all")
	})
}

func TestMapCtx(t *testing.T) {
	t.Parallel()
	t.Run("NormalMap", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		result := MapCtx(ctx, Of(1, 2, 3), func(n int) int { return n * 2 }).Collect()
		assert.Equal(t, []int{2, 4, 6}, result, "MapCtx should map *2")
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		count := 0
		var result []int
		for v := range MapCtx(ctx, Range(1, 100), func(n int) int {
			count++
			if count >= 3 {
				cancel()
			}
			return n * 2
		}).seq {
			result = append(result, v)
		}
		assert.LessOrEqual(t, len(result), 4, "MapCtx should stop when context cancelled")
	})

	t.Run("AlreadyCancelled", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		result := MapCtx(ctx, Of(1, 2, 3), func(n int) int { return n * 2 }).Collect()
		assert.Empty(t, result, "MapCtx with already-cancelled context should be empty")
	})
}

func TestMapToCtx(t *testing.T) {
	t.Parallel()
	t.Run("NormalMapTo", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		result := MapToCtx(ctx, Of(1, 2, 3), func(n int) string {
			return strings.Repeat("x", n)
		}).Collect()
		assert.Equal(t, []string{"x", "xx", "xxx"}, result, "MapToCtx should map to strings by Repeat")
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		count := 0
		var result []string
		for v := range MapToCtx(ctx, Range(1, 100), func(n int) string {
			count++
			if count >= 3 {
				cancel()
			}
			return strings.Repeat("x", n)
		}).seq {
			result = append(result, v)
		}
		assert.LessOrEqual(t, len(result), 4, "MapToCtx should stop when context cancelled")
	})

	t.Run("AlreadyCancelled", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		result := MapToCtx(ctx, Of(1, 2, 3), func(n int) string { return "x" }).Collect()
		assert.Empty(t, result, "MapToCtx with already-cancelled context should be empty")
	})
}

func TestContextError(t *testing.T) {
	t.Parallel()
	t.Run("ErrorMessage", func(t *testing.T) {
		t.Parallel()
		err := &ContextError{
			Err:     context.Canceled,
			Partial: true,
		}
		assert.Equal(t, "context canceled", err.Error(), "ContextError should wrap context error message")
		assert.True(t, err.Partial, "ContextError partial flag should be true")
	})

	t.Run("Unwrap", func(t *testing.T) {
		t.Parallel()
		err := &ContextError{
			Err:     context.DeadlineExceeded,
			Partial: false,
		}
		assert.Equal(t, context.DeadlineExceeded, err.Unwrap(), "Unwrap should return wrapped error")
		assert.False(t, err.Partial, "Partial should be false")
	})
}

// TestContextEarlyTermination tests early termination by consumer for context functions.
func TestContextEarlyTermination(t *testing.T) {
	t.Parallel()
	t.Run("RangeCtxEarlyTermination", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		result := RangeCtx(ctx, 1, 100).Limit(3).Collect()
		assert.Equal(t, []int{1, 2, 3}, result, "RangeCtx with Limit should stop early")
	})

	t.Run("FromChannelCtxEarlyTermination", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		ch := make(chan int, 10)
		for i := 1; i <= 10; i++ {
			ch <- i
		}
		close(ch)
		result := FromChannelCtx(ctx, ch).Limit(3).Collect()
		assert.Equal(t, []int{1, 2, 3}, result, "FromChannelCtx with Limit should stop early")
	})

	t.Run("FromReaderLinesCtxEarlyTermination", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		reader := strings.NewReader("line1\nline2\nline3\nline4\nline5")
		result := FromReaderLinesCtx(ctx, reader).Limit(2).Collect()
		assert.Equal(t, []string{"line1", "line2"}, result, "FromReaderLinesCtx with Limit should stop early")
	})

	t.Run("FilterCtxEarlyTermination", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		result := FilterCtx(ctx, Of(1, 2, 3, 4, 5, 6, 7, 8), func(n int) bool { return n%2 == 0 }).Limit(2).Collect()
		assert.Equal(t, []int{2, 4}, result, "FilterCtx with Limit should stop early")
	})

	t.Run("MapCtxEarlyTermination", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		result := MapCtx(ctx, Of(1, 2, 3, 4, 5), func(n int) int { return n * 2 }).Limit(3).Collect()
		assert.Equal(t, []int{2, 4, 6}, result, "MapCtx with Limit should stop early")
	})

	t.Run("MapToCtxEarlyTermination", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		result := MapToCtx(ctx, Of(1, 2, 3, 4, 5), func(n int) string {
			return strings.Repeat("x", n)
		}).Limit(3).Collect()
		assert.Equal(t, []string{"x", "xx", "xxx"}, result, "MapToCtx with Limit should stop early")
	})
}

func TestContextTimeout(t *testing.T) {
	t.Run("TimeoutContext", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()

			counter := 0
			s := GenerateCtx(ctx, func() int {
				time.Sleep(20 * time.Millisecond)
				counter++
				return counter
			})

			result := s.Limit(100).Collect()
			// Should get only a few elements before timeout
			assert.Less(t, len(result), 10, "Virtual timeout should cut the stream early")
		})
	})
}

func TestContextWithValue(t *testing.T) {
	t.Parallel()
	t.Run("ContextWithValuePassesThrough", func(t *testing.T) {
		t.Parallel()
		type key string
		ctx := context.WithValue(context.Background(), key("test"), "value")

		result := WithContext(ctx, Of(1, 2, 3)).Collect()
		assert.Equal(t, []int{1, 2, 3}, result, "WithContext should preserve values when context carries data")
	})
}
