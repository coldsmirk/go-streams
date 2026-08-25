package time_examples

import (
	"context"
	"time"

	streams "github.com/coldsmirk/go-streams"
)

// These examples print nothing: the timing-dependent values they produce
// would be flaky to assert. The empty Output block still has go test run
// them and check that they finish without printing.

func Example_debounce() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	src := streams.Of(1, 2, 3)
	_ = streams.Debounce(ctx, src, 5*time.Millisecond).Collect()
	// Output:
}

func Example_sample() {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()
	src := streams.Range(0, 100)
	_ = streams.Sample(ctx, src, 5*time.Millisecond).Collect()
	// Output:
}
