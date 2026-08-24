package temporal_test

import (
	"context"
	"fmt"
	"time"

	streams "github.com/coldsmirk/go-streams/v2"
	"github.com/coldsmirk/go-streams/v2/temporal"
)

// Throttle paces a Stream without losing anything: the first element goes
// straight through and the rest follow an interval apart.
func ExampleThrottle() {
	got := temporal.Throttle(context.Background(), streams.Of(1, 2, 3),
		10*time.Millisecond).Collect()
	fmt.Println(got)
	// Output: [1 2 3]
}

// Debounce keeps only the last element of a burst. Here the whole burst arrives
// well inside the quiet period, so it collapses to one element.
func ExampleDebounce() {
	edits := streams.Of("a", "ab", "abc")
	got := temporal.Debounce(context.Background(), edits, time.Second).Collect()
	fmt.Println(got)
	// Output: [abc]
}

// A window is cut when its size elapses, or when the source ends. This source
// ends within the first second, so its elements arrive as a single window.
func ExampleTumbling() {
	got := temporal.Tumbling(context.Background(), streams.Of(1, 2, 3), time.Second).Collect()
	fmt.Println(got)
	// Output: [[1 2 3]]
}

// The package has no error type: Timeout returns the standard
// iter.Seq2[T, error], with the deadline reported in the last pair.
func ExampleTimeout() {
	// A channel nobody sends on stands in for a source that never finishes.
	ch := make(chan int)
	defer close(ch)

	for v, err := range temporal.Timeout(context.Background(), streams.Chan(ch),
		50*time.Millisecond) {
		fmt.Println(v, err)
	}
	// Output: 0 context deadline exceeded
}

// Stamp yields a Stream2, so the package needs no timestamped value type.
func ExampleStamp() {
	start := time.Now()

	stamped := true
	for at, v := range temporal.Stamp(streams.Of("a", "b")) {
		if at.Before(start) {
			stamped = false
		}
		fmt.Print(v, " ")
	}
	fmt.Println(stamped)
	// Output: a b true
}

// Interval is a source rather than an operator, and it is infinite: something
// downstream has to end it.
func ExampleInterval() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var previous time.Time
	ordered := true
	for tick := range temporal.Interval(ctx, 10*time.Millisecond).Take(3) {
		if tick.Before(previous) {
			ordered = false
		}
		previous = tick
	}
	fmt.Println("ticks in order:", ordered)
	// Output: ticks in order: true
}
