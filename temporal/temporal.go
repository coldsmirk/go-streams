package temporal

import (
	"context"
	"iter"
	"sync"
	"time"

	streams "github.com/coldsmirk/go-streams/v2"
)

// Timeout returns a sequence of the elements of s, each paired with a nil
// error, that is cut short by context.DeadlineExceeded if s has not been
// exhausted within d.
//
// d bounds the whole iteration, not the wait for any one element: it is
// measured from the moment iteration begins and it covers the time the consumer
// spends on each element as well as the time s spends producing them. The
// deadline is only noticed between elements, so a consumer that blocks in its
// own loop body overruns it by however long it blocks there.
//
// On expiry the sequence yields one final pair of the zero value of T and
// context.DeadlineExceeded, and ends. If ctx is done first, that pair carries
// ctx.Err() instead. A sequence that finishes in time yields no error at all,
// so the error slot is non-nil at most once, in the last pair. Timeout panics
// if d is not positive.
func Timeout[T any](ctx context.Context, s streams.Stream[T], d time.Duration) iter.Seq2[T, error] {
	if d <= 0 {
		panic("temporal: Timeout called with d <= 0")
	}
	return func(yield func(T, error) bool) {
		elems, stop := pump(s)
		defer stop()

		timer := time.NewTimer(d)
		defer timer.Stop()

		var zero T
		for {
			select {
			case <-ctx.Done():
				yield(zero, ctx.Err())
				return
			case <-timer.C:
				yield(zero, context.DeadlineExceeded)
				return
			case v, ok := <-elems:
				if !ok {
					return
				}
				if canceled(ctx) {
					yield(zero, ctx.Err())
					return
				}
				if !yield(v, nil) {
					return
				}
			}
		}
	}
}

// Interval returns an infinite Stream of the times at which a tick of period d
// fires. The first tick is one d after the iteration begins, not immediately,
// and the Stream ends only when ctx is done or the consumer stops it.
//
// As with time.Ticker, a tick that fires while the consumer is still working on
// the previous one is dropped rather than queued, so the emitted times may skip
// a period but never fall behind the clock. Interval panics if d is not
// positive.
func Interval(ctx context.Context, d time.Duration) streams.Stream[time.Time] {
	if d <= 0 {
		panic("temporal: Interval called with d <= 0")
	}
	return func(yield func(time.Time) bool) {
		ticker := time.NewTicker(d)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case t := <-ticker.C:
				if canceled(ctx) {
					return
				}
				if !yield(t) {
					return
				}
			}
		}
	}
}

// Stamp returns a Stream2 pairing each element of s with the time it passed
// through, which is the time it was pulled from s rather than the time it was
// consumed downstream. Stamp does not wait on the clock, so it takes no context
// and an empty Stream stays empty.
//
// The times are read from a monotonic clock in the order the elements are
// pulled, so they never decrease, though two elements pulled in close
// succession may share one.
func Stamp[T any](s streams.Stream[T]) streams.Stream2[time.Time, T] {
	return s.KeyBy(func(T) time.Time { return time.Now() })
}

// --- internals ---

// canceled reports whether ctx is done. The operators call it before every
// emission, the end-of-source flushes included, because a done context and a
// ready element or timer are both live select cases and select breaks the tie
// at random: without the recheck an operator could emit after cancellation.
// The recheck is what turns the package promise — done means nothing further
// is emitted — from a probabilistic outcome into a deterministic one.
// (Throttle, Delay and RateLimit get the same recheck through wait, which
// every one of their emissions passes through.)
func canceled(ctx context.Context) bool { return ctx.Err() != nil }

// recv waits for the next element of elems or for ctx to be done, whichever
// happens first. It reports false when ctx is done or elems is closed.
//
// The clock operators cannot read the source with a plain range: that parks the
// consumer's own goroutine inside the source, where ctx.Done is never observed,
// so a source that goes quiet would never let the operator return.
func recv[T any](ctx context.Context, elems <-chan T) (T, bool) {
	select {
	case <-ctx.Done():
		var zero T
		return zero, false
	case v, ok := <-elems:
		return v, ok
	}
}

// pump copies the elements of s into an unbuffered channel so that an operator
// can select on them alongside a timer, and returns a stop function the caller
// must defer. The channel is closed once s is exhausted.
//
// Calling stop releases the goroutine only while it is handing an element over.
// Once the operator takes that element the goroutine re-enters s to fetch the
// next one, and from inside s the close is invisible: nothing is abandoned and
// ch is never closed. A source that goes quiet therefore keeps the goroutine
// parked until it yields again or ends. That cannot be fixed here, because Go
// offers no way to interrupt a goroutine blocked inside a caller-supplied
// iterator; iter.Pull is worse, since its stop waits on the parked coroutine
// and would hang the caller instead. See the package documentation.
func pump[T any](s streams.Stream[T]) (<-chan T, func()) {
	ch := make(chan T)
	done := make(chan struct{})
	go func() {
		defer close(ch)
		for v := range s {
			select {
			case ch <- v:
			case <-done:
				return
			}
		}
	}()
	return ch, sync.OnceFunc(func() { close(done) })
}

// wait blocks for d and reports whether it elapsed with ctx still live. A d
// that is not positive waits not at all and only reports on ctx, which keeps
// the callers a single check away from stopping the moment ctx is done.
func wait(ctx context.Context, d time.Duration) bool {
	if d <= 0 {
		return ctx.Err() == nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
