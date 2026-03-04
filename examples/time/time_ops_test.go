package time_examples

import (
	"context"
	"time"

	streams "github.com/coldsmirk/go-streams"
)

// NOTE: Time-based examples are illustrative and intentionally do not assert Output
// to avoid flakiness across slow CI environments.

func Example_debounce() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	src := streams.Of(1, 2, 3)
	_ = streams.Debounce(ctx, src, 5*time.Millisecond).Collect()
}

func Example_sample() {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()
	src := streams.Range(0, 100)
	_ = streams.Sample(ctx, src, 5*time.Millisecond).Collect()
}

