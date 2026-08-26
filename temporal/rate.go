package temporal

import (
	"context"
	"time"

	streams "github.com/coldsmirk/go-streams/v2"
)

// Throttle returns a Stream of every element of s, spacing the emissions at
// least interval apart. The first element is emitted as soon as it arrives;
// each one after it is held until interval has passed since the previous
// emission. An element that arrives more than interval after the last emission
// is not held at all.
//
// Nothing is dropped, so Throttle slows a fast source down rather than thinning
// it out; use [Sample] to keep only the most recent element of each interval.
// The Stream ends as soon as ctx is done, without emitting the element being
// held. Throttle panics if interval is not positive.
func Throttle[T any](ctx context.Context, s streams.Stream[T], interval time.Duration) streams.Stream[T] {
	if interval <= 0 {
		panic("temporal: Throttle called with interval <= 0")
	}
	return func(yield func(T) bool) {
		elems, stop := pump(s)
		defer stop()
		// The zero time is far in the past, so the first element waits for a
		// negative duration, which is to say not at all.
		var next time.Time
		for {
			v, ok := recv(ctx, elems)
			if !ok {
				return
			}
			if !wait(ctx, time.Until(next)) {
				return
			}
			next = time.Now().Add(interval)
			if !yield(v) {
				return
			}
		}
	}
}

// Debounce returns a Stream of the elements of s that are followed by quiet of
// inactivity, which coalesces a burst of rapid elements into its last one. Each
// element restarts the quiet period and replaces the element waiting on it, so
// a source that never pauses for quiet produces nothing.
//
// When s is exhausted the element waiting on the quiet period, if any, is
// emitted at once rather than after the remainder of the period. An empty
// Stream yields nothing. The Stream ends as soon as ctx is done, discarding any
// waiting element. Debounce panics if quiet is not positive.
func Debounce[T any](ctx context.Context, s streams.Stream[T], quiet time.Duration) streams.Stream[T] {
	if quiet <= 0 {
		panic("temporal: Debounce called with quiet <= 0")
	}
	return func(yield func(T) bool) {
		elems, stop := pump(s)
		defer stop()

		timer := time.NewTimer(quiet)
		defer timer.Stop()
		timer.Stop() // nothing is waiting yet

		var pending T
		waiting := false
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-elems:
				if !ok {
					if waiting && !canceled(ctx) {
						yield(pending)
					}
					return
				}
				pending, waiting = v, true
				// Resetting without draining timer.C relies on the Go 1.23
				// timer semantics, under which a stale expiry is never
				// delivered after Reset.
				timer.Reset(quiet)
			case <-timer.C:
				if canceled(ctx) {
					return
				}
				// The timer runs only while an element is waiting, so a firing
				// always has one to emit.
				waiting = false
				if !yield(pending) {
					return
				}
			}
		}
	}
}

// Sample returns a Stream of the most recent element of s at each interval,
// dropping the elements that arrive in between. The first emission is one
// interval after the iteration begins, not immediately, and an interval in
// which no element arrived produces nothing rather than repeating the previous
// element.
//
// When s is exhausted the most recent element, if it has not already been
// emitted, is emitted before the Stream ends, so the last element of a source
// shorter than interval is not lost. An empty Stream yields nothing. The Stream
// ends as soon as ctx is done, discarding any unsampled element. Sample panics
// if interval is not positive.
func Sample[T any](ctx context.Context, s streams.Stream[T], interval time.Duration) streams.Stream[T] {
	if interval <= 0 {
		panic("temporal: Sample called with interval <= 0")
	}
	return func(yield func(T) bool) {
		elems, stop := pump(s)
		defer stop()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		var latest T
		unsampled := false
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-elems:
				if !ok {
					if unsampled && !canceled(ctx) {
						yield(latest)
					}
					return
				}
				latest, unsampled = v, true
			case <-ticker.C:
				if canceled(ctx) {
					return
				}
				if !unsampled {
					continue
				}
				unsampled = false
				if !yield(latest) {
					return
				}
			}
		}
	}
}

// Delay returns a Stream of the elements of s, each emitted d after it was
// received. The waits overlap rather than queue behind one another, so the
// whole Stream is shifted by d: a burst arrives as a burst, d later, and the
// spacing between consecutive emissions matches the spacing between their
// arrivals. The elements inside the shift are buffered, so at any moment Delay
// holds whatever s produced in the last d.
//
// The Stream ends d after s is exhausted, once the last held element has been
// emitted, and as soon as ctx is done, discarding the elements still held.
// Delay panics if d is not positive.
func Delay[T any](ctx context.Context, s streams.Stream[T], d time.Duration) streams.Stream[T] {
	if d <= 0 {
		panic("temporal: Delay called with d <= 0")
	}
	return func(yield func(T) bool) {
		elems, stop := pump(s)
		defer stop()

		timer := time.NewTimer(d)
		defer timer.Stop()
		timer.Stop() // nothing is due yet

		// held is the queue of elements inside the shift, in arrival order, so
		// the element due next is always held[0] and the timer only ever needs
		// to track the head. The timer is armed exactly while held is not
		// empty: armed when the queue starts, re-armed after each emission.
		var held []stamped[T]
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-elems:
				if !ok {
					if len(held) == 0 {
						return
					}
					// Disarm this case and drain held on the timer alone: the
					// end of the Stream is shifted by d like everything else.
					elems = nil
					continue
				}
				held = append(held, stamped[T]{at: time.Now(), value: v})
				if len(held) == 1 {
					timer.Reset(d)
				}
			case <-timer.C:
				if canceled(ctx) {
					return
				}
				v := held[0].value
				held[0] = stamped[T]{} // the queue walks forward; release the slot for GC
				held = held[1:]
				if !yield(v) {
					return
				}
				if len(held) > 0 {
					// A head already overdue — the consumer was slow — gives a
					// non-positive reset, which fires the timer at once.
					timer.Reset(time.Until(held[0].at.Add(d)))
				} else if elems == nil {
					return
				}
			}
		}
	}
}

// RateLimit returns a Stream of every element of s, paced to an average of n
// elements per per. It is a token bucket that starts full: the first n elements
// are emitted as fast as they arrive, and once that budget is spent emissions
// are held to one every per/n. An idle stretch refills the bucket, up to n
// tokens, so a burst that follows one is emitted at once.
//
// Nothing is dropped; elements are only delayed. The Stream ends as soon as ctx
// is done, without emitting the element being held. RateLimit panics if n or
// per is not positive.
func RateLimit[T any](ctx context.Context, s streams.Stream[T], n int, per time.Duration) streams.Stream[T] {
	if n < 1 {
		panic("temporal: RateLimit called with n < 1")
	}
	if per <= 0 {
		panic("temporal: RateLimit called with per <= 0")
	}
	return func(yield func(T) bool) {
		// A virtual scheduling clock: emission is the interval one token buys,
		// tat the time the bucket next runs dry, and burst how far ahead of the
		// present tat may run while n tokens remain.
		emission := per / time.Duration(n)
		burst := time.Duration(n-1) * emission
		elems, stop := pump(s)
		defer stop()
		var tat time.Time
		for {
			v, ok := recv(ctx, elems)
			if !ok {
				return
			}
			now := time.Now()
			if tat.Before(now) {
				tat = now
			}
			if !wait(ctx, tat.Sub(now)-burst) {
				return
			}
			tat = tat.Add(emission)
			if !yield(v) {
				return
			}
		}
	}
}
