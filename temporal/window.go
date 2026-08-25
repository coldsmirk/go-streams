package temporal

import (
	"context"
	"time"

	streams "github.com/coldsmirk/go-streams/v2"
)

// Tumbling returns a Stream of the elements of s grouped into consecutive,
// non-overlapping windows of size. The windows are cut on a fixed schedule that
// starts when the iteration does, not when the first element arrives, and each
// window holds the elements received since the previous cut, in arrival order.
//
// A window in which no element arrived is skipped rather than emitted empty, so
// an idle source produces nothing instead of a run of empty slices. When s is
// exhausted the window still open, if it is not empty, is emitted before the
// Stream ends; a source that finishes within the first window therefore yields
// exactly one. An empty Stream yields nothing. The Stream ends as soon as ctx
// is done, discarding the open window. Tumbling panics if size is not positive.
func Tumbling[T any](ctx context.Context, s streams.Stream[T], size time.Duration) streams.Stream[[]T] {
	if size <= 0 {
		panic("temporal: Tumbling called with size <= 0")
	}
	return func(yield func([]T) bool) {
		elems, stop := pump(s)
		defer stop()

		ticker := time.NewTicker(size)
		defer ticker.Stop()

		var window []T
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-elems:
				if !ok {
					if len(window) > 0 {
						yield(window)
					}
					return
				}
				window = append(window, v)
			case <-ticker.C:
				if len(window) == 0 {
					continue
				}
				closed := window
				window = nil
				if !yield(closed) {
					return
				}
			}
		}
	}
}

// Sliding returns a Stream of overlapping windows of the elements of s: one
// window every, holding the elements received within the last size, in arrival
// order. An element appears in every window whose span covers it, so it is
// emitted about size/every times when every is the shorter of the two; every
// need not divide size, and an every longer than size leaves the elements
// between windows in no window at all.
//
// A window with nothing in its span is skipped rather than emitted empty. When
// s is exhausted a final window of the elements still within size of that
// moment is emitted, if there are any. An empty Stream yields nothing. The
// Stream ends as soon as ctx is done, discarding the elements it holds. Sliding
// panics if size or every is not positive.
func Sliding[T any](ctx context.Context, s streams.Stream[T], size, every time.Duration) streams.Stream[[]T] {
	if size <= 0 {
		panic("temporal: Sliding called with size <= 0")
	}
	if every <= 0 {
		panic("temporal: Sliding called with every <= 0")
	}
	return func(yield func([]T) bool) {
		elems, stop := pump(s)
		defer stop()

		ticker := time.NewTicker(every)
		defer ticker.Stop()

		var held []stamped[T]
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-elems:
				if !ok {
					if _, window := span(held, size); window != nil {
						yield(window)
					}
					return
				}
				held = append(held, stamped[T]{at: time.Now(), value: v})
			case <-ticker.C:
				kept, window := span(held, size)
				held = kept
				if window == nil {
					continue
				}
				if !yield(window) {
					return
				}
			}
		}
	}
}

// Session returns a Stream of the elements of s grouped into sessions, in
// arrival order: a session collects elements until gap passes with none
// arriving, and the next element after that starts a new one. The first session
// starts with the first element rather than with the iteration, and a session
// is emitted when its gap expires, not when the next session begins.
//
// When s is exhausted the open session, if any, is emitted at once rather than
// after the remainder of its gap. An empty Stream yields nothing. The Stream
// ends as soon as ctx is done, discarding the open session. Session panics if
// gap is not positive.
func Session[T any](ctx context.Context, s streams.Stream[T], gap time.Duration) streams.Stream[[]T] {
	if gap <= 0 {
		panic("temporal: Session called with gap <= 0")
	}
	return func(yield func([]T) bool) {
		elems, stop := pump(s)
		defer stop()

		timer := time.NewTimer(gap)
		defer timer.Stop()
		timer.Stop() // no session is open yet

		var session []T
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-elems:
				if !ok {
					if len(session) > 0 {
						yield(session)
					}
					return
				}
				session = append(session, v)
				// Resetting without draining timer.C relies on the Go 1.23
				// timer semantics, under which a stale expiry is never
				// delivered after Reset.
				timer.Reset(gap)
			case <-timer.C:
				// The timer runs only while a session is open, so a firing
				// always closes one that has elements in it.
				closed := session
				session = nil
				if !yield(closed) {
					return
				}
			}
		}
	}
}

// stamped is an element paired with the time [Sliding] received it.
type stamped[T any] struct {
	at    time.Time
	value T
}

// span drops the elements of held that are older than size and returns what is
// left along with their values, or a nil window if nothing is left. The
// elements are in arrival order, so the expired ones are a prefix.
func span[T any](held []stamped[T], size time.Duration) (kept []stamped[T], window []T) {
	cutoff := time.Now().Add(-size)
	expired := 0
	for expired < len(held) && !held[expired].at.After(cutoff) {
		expired++
	}
	held = append(held[:0], held[expired:]...)
	if len(held) == 0 {
		return held, nil
	}
	window = make([]T, len(held))
	for i, e := range held {
		window[i] = e.value
	}
	return held, window
}
