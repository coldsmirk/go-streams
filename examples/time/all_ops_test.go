package time_examples

import (
	"context"
	"time"

	streams "github.com/coldsmirk/go-streams"
)

// The examples below demonstrate usage patterns and compile,
// but intentionally do not assert Output to avoid flakiness on slow CI.

func Example_windows_and_timestamps() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_ = streams.WithTimestamp(streams.Of(1, 2, 3)).Collect()
	_ = streams.TumblingTimeWindow(ctx, streams.Range(0, 5), 5*time.Millisecond).Collect()
	_ = streams.SlidingTimeWindow(ctx, streams.Range(0, 5), 10*time.Millisecond, 5*time.Millisecond).Collect()
	_ = streams.SessionWindow(ctx, streams.Range(0, 3), 2*time.Millisecond).Collect()
}

func Example_rates_delays_timeouts() {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()
	_ = streams.Throttle(streams.Range(1, 5), 2*time.Millisecond).Collect()
	_ = streams.ThrottleCtx(ctx, streams.Range(1, 5), 2*time.Millisecond).Collect()
	_ = streams.RateLimit(streams.Range(1, 10), 3, 6*time.Millisecond).Collect()
	_ = streams.RateLimitCtx(ctx, streams.Range(1, 10), 3, 6*time.Millisecond).Collect()
	_ = streams.Delay(streams.Of(1, 2), 1*time.Millisecond).Collect()
	_ = streams.DelayCtx(ctx, streams.Of(1, 2), 1*time.Millisecond).Collect()
	_ = streams.Timeout(ctx, streams.Range(1, 3), 1*time.Millisecond).Collect()
}

func Example_interval_and_timer() {
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Millisecond)
	defer cancel()
	_ = streams.Interval(ctx, 3*time.Millisecond).Limit(3).Collect()
	_ = streams.Timer(ctx, 1*time.Millisecond, 42).Collect()
}

