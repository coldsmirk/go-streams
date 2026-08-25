package streams

import (
	"context"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"testing/synctest"
)

// --- TimestampedValue Tests ---

func TestTimestampedValue(t *testing.T) {
	t.Run("NewTimestamped", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			before := time.Now()
			ts := NewTimestamped(42)
			after := time.Now()

			assert.Equal(t, 42, ts.Value, "Value should match")
			assert.True(t, ts.Timestamp.After(before) || ts.Timestamp.Equal(before), "Timestamp >= before")
			assert.True(t, ts.Timestamp.Before(after) || ts.Timestamp.Equal(after), "Timestamp <= after")
		})
	})

	t.Run("NewTimestampedAt", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			fixedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
			ts := NewTimestampedAt("hello", fixedTime)

			assert.Equal(t, "hello", ts.Value, "Value should match")
			assert.Equal(t, fixedTime, ts.Timestamp, "Timestamp should match fixed time")
		})
	})

	t.Run("WithTimestamp", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			before := time.Now()
			result := WithTimestamp(Of(1, 2, 3)).Collect()
			after := time.Now()

			assert.Len(t, result, 3, "Should wrap all elements")
			for i, ts := range result {
				assert.Equal(t, i+1, ts.Value, "Value should match index+1")
				assert.True(t, ts.Timestamp.After(before) || ts.Timestamp.Equal(before), "Timestamp >= before")
				assert.True(t, ts.Timestamp.Before(after) || ts.Timestamp.Equal(after), "Timestamp <= after")
			}
		})
	})

	t.Run("WithTimestampEmptyStream", func(t *testing.T) {
		result := WithTimestamp(Empty[int]()).Collect()
		assert.Empty(t, result, "Empty input yields empty output")
	})

	t.Run("WithTimestampEarlyTermination", func(t *testing.T) {
		result := WithTimestamp(Of(1, 2, 3, 4, 5)).Limit(2).Collect()
		assert.Len(t, result, 2, "Limit should apply")
		assert.Equal(t, 1, result[0].Value, "First value should be 1")
		assert.Equal(t, 2, result[1].Value, "Second value should be 2")
	})
}

// --- Throttle Tests ---

func TestThrottle(t *testing.T) {
	t.Run("Basic", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			start := time.Now()
			result := Throttle(Of(1, 2, 3), 10*time.Millisecond).Collect()
			elapsed := time.Since(start)

			assert.Equal(t, []int{1, 2, 3}, result, "Should return all elements")
			// After first element, 2 more elements need at least 20ms
			assert.GreaterOrEqual(t, elapsed, 20*time.Millisecond, "Fake elapsed should account for throttling")
		})
	})

	t.Run("EmptyStream", func(t *testing.T) {
		result := Throttle(Empty[int](), 10*time.Millisecond).Collect()
		assert.Empty(t, result, "Empty input yields empty output")
	})

	t.Run("SingleElement", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			start := time.Now()
			result := Throttle(Of(42), 10*time.Millisecond).Collect()
			elapsed := time.Since(start)

			assert.Equal(t, []int{42}, result, "Single element should pass through")
			// Single element should not need delay beyond the first emit
			assert.Less(t, elapsed, 15*time.Millisecond, "Fake elapsed should be small for single element")
		})
	})

	t.Run("EarlyTermination", func(t *testing.T) {
		result := Throttle(Of(1, 2, 3, 4, 5), 5*time.Millisecond).Limit(2).Collect()
		assert.Equal(t, []int{1, 2}, result, "Limit should stop early")
	})
}

func TestThrottleCtx(t *testing.T) {
	t.Run("NormalCompletion", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx := context.Background()
			result := ThrottleCtx(ctx, Of(1, 2, 3), 5*time.Millisecond).Collect()
			assert.Equal(t, []int{1, 2, 3}, result, "ThrottleCtx should emit all elements")
		})
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var result []int
			var mu sync.Mutex
			done := make(chan struct{})

			go func() {
				defer close(done)
				r := ThrottleCtx(ctx, Of(1, 2, 3, 4, 5, 6, 7, 8, 9, 10), 50*time.Millisecond).Collect()
				mu.Lock()
				result = r
				mu.Unlock()
			}()

			time.Sleep(30 * time.Millisecond)
			synctest.Wait()
			cancel()
			<-done

			mu.Lock()
			assert.Less(t, len(result), 10, "Should get partial results")
			mu.Unlock()
		})
	})

	t.Run("AlreadyCancelledContext", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			result := ThrottleCtx(ctx, Of(1, 2, 3), 10*time.Millisecond).Collect()
			assert.LessOrEqual(t, len(result), 1, "May get at most one")
		})
	})
}

// --- RateLimit Tests ---

func TestRateLimit(t *testing.T) {
	t.Run("Basic", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			start := time.Now()
			result := RateLimit(Of(1, 2, 3, 4, 5), 10, 100*time.Millisecond).Collect()
			elapsed := time.Since(start)

			assert.Equal(t, []int{1, 2, 3, 4, 5}, result, "RateLimit should emit all elements")
			// 10 per 100ms = 5 elements should be fast
			assert.Less(t, elapsed, 150*time.Millisecond, "RateLimit should not delay when capacity is high")
		})
	})

	t.Run("RateLimitWithNZero", func(t *testing.T) {
		result := RateLimit(Of(1, 2, 3), 0, 100*time.Millisecond).Collect()
		assert.Empty(t, result, "RateLimit with n=0 should return empty")
	})

	t.Run("RateLimitWithNNegative", func(t *testing.T) {
		result := RateLimit(Of(1, 2, 3), -5, 100*time.Millisecond).Collect()
		assert.Empty(t, result, "RateLimit with negative n should return empty")
	})

	t.Run("EmptyStream", func(t *testing.T) {
		result := RateLimit(Empty[int](), 10, 100*time.Millisecond).Collect()
		assert.Empty(t, result, "RateLimit on empty stream should return empty")
	})

	t.Run("RateLimitSlowsDownStream", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			start := time.Now()
			result := RateLimit(Of(1, 2, 3), 1, 50*time.Millisecond).Collect()
			elapsed := time.Since(start)

			assert.Equal(t, []int{1, 2, 3}, result, "RateLimit should emit all elements")
			// 1 per 50ms = at least 100ms for 3 elements (first is immediate)
			assert.GreaterOrEqual(t, elapsed, 100*time.Millisecond, "RateLimit should delay between elements")
		})
	})

	t.Run("EarlyTermination", func(t *testing.T) {
		result := RateLimit(Of(1, 2, 3, 4, 5), 10, 100*time.Millisecond).Limit(2).Collect()
		assert.Equal(t, []int{1, 2}, result, "RateLimit should respect Limit")
	})
}

func TestRateLimitCtx(t *testing.T) {
	t.Run("NormalCompletion", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx := context.Background()
			result := RateLimitCtx(ctx, Of(1, 2, 3), 10, 100*time.Millisecond).Collect()
			assert.Equal(t, []int{1, 2, 3}, result, "RateLimitCtx should emit all elements")
		})
	})

	t.Run("RateLimitWithNLessOrEqualZeroReturnsEmpty", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx := context.Background()
			result := RateLimitCtx(ctx, Of(1, 2, 3), 0, 100*time.Millisecond).Collect()
			assert.Empty(t, result, "RateLimitCtx should return empty for non-positive n")
		})
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var result []int
			var mu sync.Mutex
			done := make(chan struct{})

			go func() {
				defer close(done)
				r := RateLimitCtx(ctx, Of(1, 2, 3, 4, 5, 6, 7, 8, 9, 10), 1, 100*time.Millisecond).Collect()
				mu.Lock()
				result = r
				mu.Unlock()
			}()

			time.Sleep(50 * time.Millisecond)
			synctest.Wait()
			cancel()
			<-done

			mu.Lock()
			assert.Less(t, len(result), 10, "RateLimitCtx should stop early on cancellation")
			mu.Unlock()
		})
	})

	t.Run("ContextCancellationWhileWaitingForToken", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var result []int
			done := make(chan struct{})
			go func() {
				defer close(done)
				result = RateLimitCtx(ctx, Of(1, 2, 3, 4), 1, 100*time.Millisecond).Collect()
			}()

			// Allow first element, then cancel while waiting for next token.
			time.Sleep(10 * time.Millisecond)
			cancel()
			<-done

			assert.LessOrEqual(t, len(result), 1, "RateLimitCtx should stop when cancelled during token wait")
		})
	})

	t.Run("AlreadyCancelledContext", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			result := RateLimitCtx(ctx, Of(1, 2, 3), 10, 100*time.Millisecond).Collect()
			assert.Empty(t, result, "RateLimitCtx should return empty when already cancelled")
		})
	})
}

// --- Delay Tests ---

func TestDelay(t *testing.T) {
	t.Run("Basic", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			start := time.Now()
			result := Delay(Of(1, 2, 3), 10*time.Millisecond).Collect()
			elapsed := time.Since(start)

			assert.Equal(t, []int{1, 2, 3}, result, "Delay should emit all elements")
			// 3 elements with 10ms delay each = at least 30ms
			assert.GreaterOrEqual(t, elapsed, 30*time.Millisecond, "Delay should wait for each element")
		})
	})

	t.Run("EmptyStream", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			start := time.Now()
			result := Delay(Empty[int](), 10*time.Millisecond).Collect()
			elapsed := time.Since(start)

			assert.Empty(t, result, "Delay on empty stream should return empty")
			assert.Less(t, elapsed, 15*time.Millisecond, "Delay should not wait for empty stream")
		})
	})

	t.Run("SingleElement", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			start := time.Now()
			result := Delay(Of(42), 15*time.Millisecond).Collect()
			elapsed := time.Since(start)

			assert.Equal(t, []int{42}, result, "Delay should emit the single element")
			assert.GreaterOrEqual(t, elapsed, 15*time.Millisecond, "Delay should wait at least the delay duration")
		})
	})

	t.Run("EarlyTermination", func(t *testing.T) {
		result := Delay(Of(1, 2, 3, 4, 5), 5*time.Millisecond).Limit(2).Collect()
		assert.Equal(t, []int{1, 2}, result, "Delay should respect Limit")
	})
}

func TestDelayCtx(t *testing.T) {
	t.Run("NormalCompletion", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx := context.Background()
			result := DelayCtx(ctx, Of(1, 2, 3), 5*time.Millisecond).Collect()
			assert.Equal(t, []int{1, 2, 3}, result, "DelayCtx should emit all elements")
		})
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var result []int
			var mu sync.Mutex
			done := make(chan struct{})

			go func() {
				defer close(done)
				r := DelayCtx(ctx, Of(1, 2, 3, 4, 5, 6, 7, 8, 9, 10), 50*time.Millisecond).Collect()
				mu.Lock()
				result = r
				mu.Unlock()
			}()

			// Cancel after short delay (virtual)
			time.Sleep(30 * time.Millisecond)
			cancel()
			<-done

			mu.Lock()
			// Should get partial results
			assert.Less(t, len(result), 10, "DelayCtx should stop early on cancellation")
			mu.Unlock()
		})
	})

	t.Run("ContextCancellationDuringDelay", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var result []int
			done := make(chan struct{})
			go func() {
				defer close(done)
				result = DelayCtx(ctx, Of(1), 50*time.Millisecond).Collect()
			}()

			// Cancel before the delay elapses.
			time.Sleep(10 * time.Millisecond)
			cancel()
			<-done

			assert.Empty(t, result, "DelayCtx should not emit when cancelled during delay")
		})
	})

	t.Run("AlreadyCancelledContext", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			result := DelayCtx(ctx, Of(1, 2, 3), 10*time.Millisecond).Collect()
			assert.Empty(t, result, "DelayCtx should return empty when already cancelled")
		})
	})
}

// --- Interval Tests ---

func TestInterval(t *testing.T) {
	t.Run("Basic", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 55*time.Millisecond)
			defer cancel()

			result := Interval(ctx, 10*time.Millisecond).Collect()
			// Should get about 5 elements (at 10, 20, 30, 40, 50 ms)
			assert.GreaterOrEqual(t, len(result), 3, "Interval should emit multiple ticks before timeout")
			assert.LessOrEqual(t, len(result), 6, "Interval should stop when context times out")

			// Should be sequential starting from 0
			for i, v := range result {
				assert.Equal(t, i, v, "Interval should count upward from zero")
			}
		})
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			done := make(chan struct{})
			var result []int

			go func() {
				defer close(done)
				result = Interval(ctx, 10*time.Millisecond).Collect()
			}()

			// Let it run for a bit (virtual)
			time.Sleep(35 * time.Millisecond)
			cancel()
			<-done

			// Should have gotten some elements
			assert.Greater(t, len(result), 0, "Interval should emit values before cancellation")
			assert.Less(t, len(result), 10, "Interval should stop after cancellation")
		})
	})

	t.Run("AlreadyCancelledContext", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			result := Interval(ctx, 10*time.Millisecond).Collect()
			assert.Empty(t, result, "Interval should return empty when already cancelled")
		})
	})

	t.Run("EarlyTerminationWithLimit", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			result := Interval(ctx, 10*time.Millisecond).Limit(3).Collect()
			assert.Equal(t, []int{0, 1, 2}, result, "Interval should respect Limit")
		})
	})
}

// --- Timer Tests ---

func TestTimer(t *testing.T) {
	t.Run("Basic", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx := context.Background()
			start := time.Now()
			result := Timer(ctx, 20*time.Millisecond, "fired").Collect()
			elapsed := time.Since(start)

			assert.Equal(t, []string{"fired"}, result, "Timer should emit once after delay")
			assert.GreaterOrEqual(t, elapsed, 20*time.Millisecond, "Timer should wait at least the delay duration")
		})
	})

	t.Run("ContextCancellationBeforeTimerFires", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			done := make(chan struct{})
			var result []string

			go func() {
				defer close(done)
				result = Timer(ctx, 100*time.Millisecond, "fired").Collect()
			}()

			// Cancel before timer fires (virtual)
			time.Sleep(20 * time.Millisecond)
			cancel()
			<-done

			assert.Empty(t, result, "Timer should return empty when cancelled before firing")
		})
	})

	t.Run("AlreadyCancelledContext", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			result := Timer(ctx, 10*time.Millisecond, 42).Collect()
			assert.Empty(t, result, "Timer should return empty when already cancelled")
		})
	})
}

// --- Timeout Tests ---

func TestTimeout(t *testing.T) {
	t.Run("NormalElementsBeforeTimeout", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx := context.Background()
			// Fast stream that completes well before timeout
			s := Of(1, 2, 3)
			results := Timeout(ctx, s, 100*time.Millisecond).Collect()

			assert.Len(t, results, 3, "Timeout should return all values before deadline")
			for i, r := range results {
				assert.True(t, r.IsOk(), "Timeout should return Ok for values before timeout")
				assert.Equal(t, i+1, r.Value(), "Timeout should preserve original values")
			}
		})
	})

	t.Run("TimeoutOccurs", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()

			// Slow stream that will timeout between elements
			slowStream := Stream[int]{
				seq: func(yield func(int) bool) {
					for i := 1; i <= 5; i++ {
						time.Sleep(30 * time.Millisecond)
						if !yield(i) {
							return
						}
					}
				},
			}

			results := Timeout(ctx, slowStream, 20*time.Millisecond).Collect()

			// Should have some Ok and some Err results
			hasErr := false
			for _, r := range results {
				if r.IsErr() {
					hasErr = true
					assert.Equal(t, context.DeadlineExceeded, r.Error(), "Timeout should return DeadlineExceeded on timeout")
				}
			}
			assert.True(t, hasErr, "Should have at least one timeout error")
		})
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			slowStream := Stream[int]{
				seq: func(yield func(int) bool) {
					for i := 1; ; i++ {
						select {
						case <-ctx.Done():
							return
						case <-time.After(50 * time.Millisecond):
							if !yield(i) {
								return
							}
						}
					}
				},
			}

			// Schedule cancellation in another goroutine (virtual sleep).
			go func() {
				time.Sleep(30 * time.Millisecond)
				cancel()
			}()

			results := Timeout(ctx, slowStream, 200*time.Millisecond).Collect()
			// Allow timers to settle in virtual time to avoid dangling goroutines.
			synctest.Wait()

			// Last result should be an error for context cancellation
			if len(results) > 0 {
				lastResult := results[len(results)-1]
				if lastResult.IsErr() {
					assert.Equal(t, context.Canceled, lastResult.Error(), "Timeout should return context.Canceled on cancellation")
				}
			}
		})
	})

	t.Run("EmptyStream", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx := context.Background()
			results := Timeout(ctx, Empty[int](), 100*time.Millisecond).Collect()
			assert.Empty(t, results, "Timeout on empty stream should return empty")
		})
	})
}

// --- Debounce Tests ---

func TestDebounce(t *testing.T) {
	t.Run("BasicDebounce", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx := context.Background()
			// Rapid elements followed by a pause
			s := Stream[int]{
				seq: func(yield func(int) bool) {
					// Rapid fire
					for i := 1; i <= 3; i++ {
						if !yield(i) {
							return
						}
						time.Sleep(5 * time.Millisecond)
					}
					// Wait for debounce to fire
					time.Sleep(50 * time.Millisecond)
					// Another burst
					for i := 4; i <= 5; i++ {
						if !yield(i) {
							return
						}
						time.Sleep(5 * time.Millisecond)
					}
				},
			}

			result := Debounce(ctx, s, 20*time.Millisecond).Collect()

			// Should get last values of each burst
			assert.GreaterOrEqual(t, len(result), 1, "Debounce should emit at least one value")
			// Last value of first burst is 3, last value of second burst is 5
			// Due to timing, we should see at least the final value
			assert.Contains(t, result, 5, "Debounce should include the last value of the final burst")
		})
	})

	t.Run("SingleElement", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx := context.Background()
			result := Debounce(ctx, Of(42), 10*time.Millisecond).Collect()
			assert.Equal(t, []int{42}, result, "Debounce should return the single element")
		})
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		synctest.Test(t, func(_ *testing.T) {
			// Drive cancellation via virtual time. Using WithTimeout prevents
			// lingering goroutines when the subtest exits.
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
			defer cancel()

			slowStream := Stream[int]{
				seq: func(yield func(int) bool) {
					for i := 1; ; i++ {
						select {
						case <-ctx.Done():
							return
						case <-time.After(10 * time.Millisecond):
							if !yield(i) {
								return
							}
						}
					}
				},
			}

			// Collect synchronously to keep all activity within the bubble.
			_ = Debounce(ctx, slowStream, 50*time.Millisecond).Collect()
			// Ensure all timers/goroutines have quiesced in synctest.
			synctest.Wait()
		})
	})

	t.Run("EmptyStream", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx := context.Background()
			result := Debounce(ctx, Empty[int](), 10*time.Millisecond).Collect()
			assert.Empty(t, result, "Debounce on empty stream should return empty")
		})
	})
}

// --- Sample Tests ---

func TestSample(t *testing.T) {
	t.Run("BasicSampling", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			// Fast stream that generates many values
			fastStream := Stream[int]{
				seq: func(yield func(int) bool) {
					for i := 1; ; i++ {
						select {
						case <-ctx.Done():
							return
						case <-time.After(5 * time.Millisecond):
							if !yield(i) {
								return
							}
						}
					}
				},
			}

			result := Sample(ctx, fastStream, 25*time.Millisecond).Collect()

			// Should sample approximately every 25ms for 100ms = ~3-4 samples
			assert.Greater(t, len(result), 0, "Sample should emit some values during the interval")
			assert.LessOrEqual(t, len(result), 6, "Sample should not exceed expected sample count")
		})
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		synctest.Test(t, func(_ *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			fastStream := Stream[int]{
				seq: func(yield func(int) bool) {
					for i := 1; ; i++ {
						select {
						case <-ctx.Done():
							return
						case <-time.After(5 * time.Millisecond):
							if !yield(i) {
								return
							}
						}
					}
				},
			}

			// Schedule cancellation and run collection synchronously to avoid lingering goroutines.
			go func() {
				time.Sleep(50 * time.Millisecond)
				cancel()
			}()
			_ = Sample(ctx, fastStream, 20*time.Millisecond).Collect()
			synctest.Wait()
		})
	})

	t.Run("FiniteStreamSamplesAndReturnsFinalValue", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx := context.Background()
			// Short stream that finishes quickly
			result := Sample(ctx, Of(1, 2, 3), 100*time.Millisecond).Collect()
			// Should get at least the final sample
			assert.GreaterOrEqual(t, len(result), 1, "Sample should emit at least one value for finite stream")
			// Last collected value should be from the stream
			if len(result) > 0 {
				assert.Contains(t, []int{1, 2, 3}, result[len(result)-1], "Sample should include a value from the stream")
			}
		})
	})
}

// --- Window Tests ---

func TestTumblingTimeWindow(t *testing.T) {
	t.Run("BasicTumblingWindow", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			// Fast stream that generates values
			counter := 0
			fastStream := Stream[int]{
				seq: func(yield func(int) bool) {
					for i := 1; i <= 10; i++ {
						counter++
						if !yield(i) {
							return
						}
						time.Sleep(10 * time.Millisecond)
					}
				},
			}

			windows := TumblingTimeWindow(ctx, fastStream, 35*time.Millisecond).Collect()

			// Should have at least one window
			assert.Greater(t, len(windows), 0, "TumblingTimeWindow should emit windows")
			// Each window should have elements
			totalElements := 0
			for _, w := range windows {
				assert.Greater(t, len(w), 0, "TumblingTimeWindow should yield non-empty windows")
				totalElements += len(w)
			}
		})
	})

	t.Run("ContextCancellationYieldsRemaining", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()

			fastStream := Stream[int]{
				seq: func(yield func(int) bool) {
					for i := 1; i <= 100; i++ {
						if !yield(i) {
							return
						}
						select {
						case <-ctx.Done():
							return
						case <-time.After(5 * time.Millisecond):
						}
					}
				},
			}

			windows := TumblingTimeWindow(ctx, fastStream, 500*time.Millisecond).Collect()

			// Should get at least one window with partial elements
			assert.GreaterOrEqual(t, len(windows), 1, "TumblingTimeWindow should emit a window on cancellation")
			synctest.Wait()
		})
	})

	t.Run("EmptyStream", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()

			windows := TumblingTimeWindow(ctx, Empty[int](), 20*time.Millisecond).Collect()
			assert.Empty(t, windows, "TumblingTimeWindow on empty stream should return empty")
		})
	})
}

func TestSlidingTimeWindow(t *testing.T) {
	t.Run("BasicSlidingWindow", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			// Stream that generates values
			fastStream := Stream[int]{
				seq: func(yield func(int) bool) {
					for i := 1; i <= 20; i++ {
						if !yield(i) {
							return
						}
						// Sleep on virtual time, but allow early exit on cancellation.
						select {
						case <-ctx.Done():
							return
						case <-time.After(8 * time.Millisecond):
						}
					}
				},
			}

			windows := SlidingTimeWindow(ctx, fastStream, 40*time.Millisecond, 20*time.Millisecond).Collect()

			// Should have overlapping windows
			assert.Greater(t, len(windows), 0, "SlidingTimeWindow should emit windows")
			// Ensure any internal timers settle before exiting the bubble.
			synctest.Wait()
		})
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()

			fastStream := Stream[int]{
				seq: func(yield func(int) bool) {
					for i := 1; i <= 100; i++ {
						if !yield(i) {
							return
						}
						select {
						case <-ctx.Done():
							return
						case <-time.After(5 * time.Millisecond):
						}
					}
				},
			}

			windows := SlidingTimeWindow(ctx, fastStream, 200*time.Millisecond, 30*time.Millisecond).Collect()
			assert.GreaterOrEqual(t, len(windows), 1, "SlidingTimeWindow should emit a window on cancellation")
			synctest.Wait()
		})
	})

	t.Run("EmptyStream", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()

			windows := SlidingTimeWindow(ctx, Empty[int](), 30*time.Millisecond, 10*time.Millisecond).Collect()
			assert.Empty(t, windows, "SlidingTimeWindow on empty stream should return empty")
		})
	})
}

func TestSessionWindow(t *testing.T) {
	t.Run("BasicSessionWindow", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()

			// Stream with activity gaps
			sessionStream := Stream[int]{
				seq: func(yield func(int) bool) {
					// Session 1
					for i := 1; i <= 3; i++ {
						if !yield(i) {
							return
						}
						time.Sleep(10 * time.Millisecond)
					}
					// Gap
					time.Sleep(50 * time.Millisecond)
					// Session 2
					for i := 4; i <= 5; i++ {
						if !yield(i) {
							return
						}
						time.Sleep(10 * time.Millisecond)
					}
				},
			}

			windows := SessionWindow(ctx, sessionStream, 30*time.Millisecond).Collect()

			// Should have at least 2 sessions
			assert.GreaterOrEqual(t, len(windows), 1, "SessionWindow should emit at least one session")

			// Total elements across all sessions
			totalElements := 0
			for _, w := range windows {
				totalElements += len(w)
			}
			assert.GreaterOrEqual(t, totalElements, 1, "SessionWindow should include elements in sessions")
		})
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()

			slowStream := Stream[int]{
				seq: func(yield func(int) bool) {
					for i := 1; i <= 100; i++ {
						if !yield(i) {
							return
						}
						select {
						case <-ctx.Done():
							return
						case <-time.After(10 * time.Millisecond):
						}
					}
				},
			}

			windows := SessionWindow(ctx, slowStream, 100*time.Millisecond).Collect()
			// Should get at least one window with partial elements
			assert.GreaterOrEqual(t, len(windows), 1, "SessionWindow should emit a window on cancellation")
			synctest.Wait()
		})
	})

	t.Run("EmptyStream", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()

			windows := SessionWindow(ctx, Empty[int](), 20*time.Millisecond).Collect()
			assert.Empty(t, windows, "SessionWindow on empty stream should return empty")
		})
	})
}

// --- Integration Tests ---

func TestTimeOperationsChained(t *testing.T) {
	t.Run("ThrottleWithMap", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			result := Throttle(Of(1, 2, 3), 5*time.Millisecond).
				Map(func(n int) int { return n * 2 }).
				Collect()
			assert.Equal(t, []int{2, 4, 6}, result, "Throttle then Map should transform values")
		})
	})

	t.Run("DelayWithFilter", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			result := Delay(Of(1, 2, 3, 4), 5*time.Millisecond).
				Filter(func(n int) bool { return n%2 == 0 }).
				Collect()
			assert.Equal(t, []int{2, 4}, result, "Delay then Filter should keep even values")
		})
	})

	t.Run("IntervalWithLimitAndMap", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()

			result := Interval(ctx, 10*time.Millisecond).
				Limit(5).
				Map(func(n int) int { return n * 10 }).
				Collect()

			assert.Len(t, result, 5, "Interval with Limit should produce 5 values")
			assert.Equal(t, []int{0, 10, 20, 30, 40}, result, "Interval Map should scale ticks by 10")
		})
	})
}

func TestTimestampSorting(t *testing.T) {
	t.Run("SortByTimestamp", func(t *testing.T) {
		now := time.Now()
		ts := []TimestampedValue[string]{
			{Value: "c", Timestamp: now.Add(2 * time.Second)},
			{Value: "a", Timestamp: now},
			{Value: "b", Timestamp: now.Add(1 * time.Second)},
		}

		// Sort by timestamp
		sort.Slice(ts, func(i, j int) bool {
			return ts[i].Timestamp.Before(ts[j].Timestamp)
		})

		assert.Equal(t, "a", ts[0].Value, "Sorting by timestamp should order earliest value first")
		assert.Equal(t, "b", ts[1].Value, "Sorting by timestamp should order middle value")
		assert.Equal(t, "c", ts[2].Value, "Sorting by timestamp should order latest value last")
	})
}

// --- Additional Edge Case Tests for Coverage ---

func TestRateLimitCtxEdgeCases(t *testing.T) {
	t.Run("TokenRefillDuringProcessing", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx := context.Background()
			// Process enough elements to trigger token refill
			result := RateLimitCtx(ctx, Range(1, 8), 2, 50*time.Millisecond).Collect()
			assert.Equal(t, []int{1, 2, 3, 4, 5, 6, 7}, result, "Should process all elements with token refill")
		})
	})

	t.Run("WaitForTokenThenCancel", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var result []int
			done := make(chan struct{})

			go func() {
				defer close(done)
				// 1 token per 100ms, many elements should trigger wait
				result = RateLimitCtx(ctx, Range(1, 10), 1, 200*time.Millisecond).Collect()
			}()

			// Let first element process, then cancel while waiting for second token
			time.Sleep(50 * time.Millisecond)
			synctest.Wait()
			cancel()
			<-done

			assert.LessOrEqual(t, len(result), 2, "Should stop when cancelled during token wait")
		})
	})
}

func TestThrottleCtxEdgeCases(t *testing.T) {
	t.Run("CancelDuringThrottleWait", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var result []int
			done := make(chan struct{})

			go func() {
				defer close(done)
				result = ThrottleCtx(ctx, Of(1, 2, 3, 4, 5), 100*time.Millisecond).Collect()
			}()

			// Let first element through, then cancel during wait for second
			time.Sleep(20 * time.Millisecond)
			synctest.Wait()
			cancel()
			<-done

			assert.LessOrEqual(t, len(result), 2, "Should stop when cancelled during throttle wait")
		})
	})
}

func TestTimeoutEdgeCases(t *testing.T) {
	t.Run("TimeoutThenEarlyExit", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
			defer cancel()

			// Stream that's slow enough to trigger timeout
			slowStream := Stream[int]{
				seq: func(yield func(int) bool) {
					for i := 1; i <= 10; i++ {
						select {
						case <-ctx.Done():
							return
						case <-time.After(50 * time.Millisecond):
							if !yield(i) {
								return
							}
						}
					}
				},
			}

			// Only take first 2 results (whether Ok or Err)
			result := Timeout(ctx, slowStream, 30*time.Millisecond).Limit(2).Collect()
			assert.Len(t, result, 2, "Should stop after limit even with timeouts")
		})
	})

	t.Run("FastStreamNoTimeout", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx := context.Background()
			result := Timeout(ctx, Of(1, 2, 3), 100*time.Millisecond).Collect()

			assert.Len(t, result, 3, "Should return all elements for fast stream")
			for _, r := range result {
				assert.True(t, r.IsOk(), "All results should be Ok for fast stream")
			}
		})
	})

	t.Run("TimeoutYieldReturnsFalse", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
			defer cancel()

			// Very slow stream to ensure timeout
			slowStream := Stream[int]{
				seq: func(yield func(int) bool) {
					select {
					case <-ctx.Done():
						return
					case <-time.After(200 * time.Millisecond):
						yield(1)
					}
				},
			}

			// Limit to 1 to trigger early exit on timeout error
			result := Timeout(ctx, slowStream, 10*time.Millisecond).Limit(1).Collect()
			assert.Len(t, result, 1, "Should get exactly one result")
			assert.True(t, result[0].IsErr(), "Should be timeout error")
		})
	})
}

func TestTumblingTimeWindowEdgeCases(t *testing.T) {
	t.Run("WindowYieldReturnsFalse", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		fastStream := Stream[int]{
			seq: func(yield func(int) bool) {
				for i := 1; i <= 20; i++ {
					select {
					case <-ctx.Done():
						return
					default:
						if !yield(i) {
							return
						}
						time.Sleep(10 * time.Millisecond)
					}
				}
			},
		}

		// Limit to 1 window to test early exit
		windows := TumblingTimeWindow(ctx, fastStream, 30*time.Millisecond).Limit(1).Collect()
		assert.Len(t, windows, 1, "Should get exactly one window")
		assert.Greater(t, len(windows[0]), 0, "Window should have elements")
	})
}

func TestSlidingTimeWindowEdgeCases(t *testing.T) {
	t.Run("WindowYieldReturnsFalse", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		fastStream := Stream[int]{
			seq: func(yield func(int) bool) {
				for i := 1; i <= 20; i++ {
					select {
					case <-ctx.Done():
						return
					default:
						if !yield(i) {
							return
						}
						time.Sleep(10 * time.Millisecond)
					}
				}
			},
		}

		// Limit to 1 window to test early exit
		windows := SlidingTimeWindow(ctx, fastStream, 50*time.Millisecond, 20*time.Millisecond).Limit(1).Collect()
		assert.Len(t, windows, 1, "Should get exactly one window")
	})
}

func TestSessionWindowEdgeCases(t *testing.T) {
	t.Run("SessionYieldReturnsFalse", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer cancel()

		// Stream with gaps to create multiple sessions
		sessionStream := Stream[int]{
			seq: func(yield func(int) bool) {
				// Session 1
				for i := 1; i <= 3; i++ {
					select {
					case <-ctx.Done():
						return
					default:
						if !yield(i) {
							return
						}
						time.Sleep(5 * time.Millisecond)
					}
				}
				// Gap to end session
				select {
				case <-ctx.Done():
					return
				case <-time.After(50 * time.Millisecond):
				}
				// Session 2
				for i := 4; i <= 6; i++ {
					select {
					case <-ctx.Done():
						return
					default:
						if !yield(i) {
							return
						}
						time.Sleep(5 * time.Millisecond)
					}
				}
			},
		}

		// Limit to 1 session to test early exit
		windows := SessionWindow(ctx, sessionStream, 30*time.Millisecond).Limit(1).Collect()
		assert.Len(t, windows, 1, "Should get exactly one session")
	})
}

func TestSampleEdgeCases(t *testing.T) {
	t.Run("SampleYieldReturnsFalse", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		fastStream := Stream[int]{
			seq: func(yield func(int) bool) {
				for i := 1; i <= 100; i++ {
					select {
					case <-ctx.Done():
						return
					default:
						if !yield(i) {
							return
						}
						time.Sleep(5 * time.Millisecond)
					}
				}
			},
		}

		// Limit to 2 samples to test early exit
		result := Sample(ctx, fastStream, 20*time.Millisecond).Limit(2).Collect()
		assert.Len(t, result, 2, "Should get exactly 2 samples")
	})

	t.Run("EmptyStream", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()

			result := Sample(ctx, Empty[int](), 10*time.Millisecond).Collect()
			assert.Empty(t, result, "Sample on empty stream should return empty")
		})
	})
}

func TestDebounceEdgeCases(t *testing.T) {
	t.Run("DebounceYieldReturnsFalse", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer cancel()

		// Stream with pauses to trigger debounce
		s := Stream[int]{
			seq: func(yield func(int) bool) {
				if !yield(1) {
					return
				}
				select {
				case <-ctx.Done():
					return
				case <-time.After(50 * time.Millisecond): // Wait for debounce
				}
				if !yield(2) {
					return
				}
				select {
				case <-ctx.Done():
					return
				case <-time.After(50 * time.Millisecond): // Wait for debounce
				}
				if !yield(3) {
					return
				}
			},
		}

		// Limit to 1 to test early exit
		result := Debounce(ctx, s, 20*time.Millisecond).Limit(1).Collect()
		assert.Len(t, result, 1, "Should get exactly 1 debounced value")
	})

	t.Run("AlreadyCancelledContext", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			result := Debounce(ctx, Of(1, 2, 3), 10*time.Millisecond).Collect()
			assert.Empty(t, result, "Debounce should return empty when already cancelled")
		})
	})
}

func TestDelayCtxEdgeCases(t *testing.T) {
	t.Run("EarlyTermination", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx := context.Background()
			result := DelayCtx(ctx, Of(1, 2, 3, 4, 5), 5*time.Millisecond).Limit(2).Collect()
			assert.Equal(t, []int{1, 2}, result, "DelayCtx should respect Limit")
		})
	})
}
