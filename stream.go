package streams

import (
	"context"
	"iter"
	"slices"
)

// Stream is a lazy sequence of values. Its underlying type is iter.Seq[T], so a
// Stream may be converted to and from iter.Seq at no cost and ranged over
// directly.
type Stream[T any] iter.Seq[T]

// From returns seq as a Stream. It is a conversion that infers T, which is its
// only purpose: Stream[T](seq) requires naming T, From(seq) does not.
func From[T any](seq iter.Seq[T]) Stream[T] { return Stream[T](seq) }

// Of returns a Stream over values. Pass a slice with s... to stream it.
func Of[T any](values ...T) Stream[T] { return Stream[T](slices.Values(values)) }

// Empty returns a Stream with no elements.
func Empty[T any]() Stream[T] { return func(func(T) bool) {} }

// Chan returns a Stream over the values received from ch, ending when ch is
// closed. Stopping the iteration early leaves any unreceived values in ch.
//
// A Stream is pulled, so the only way out of a receive that never completes is
// for ch to produce or close. An operator that reads the Stream on a goroutine
// of its own, as those in [github.com/coldsmirk/go-streams/v2/temporal] do,
// therefore cannot release that goroutine while ch stays quiet. Use
// [ChanContext] for a pipeline that must shut down on cancellation alone.
func Chan[T any](ch <-chan T) Stream[T] {
	return func(yield func(T) bool) {
		for v := range ch {
			if !yield(v) {
				return
			}
		}
	}
}

// ChanContext returns a Stream over the values received from ch, ending when ch
// is closed or ctx is done, whichever happens first. Stopping the iteration
// early leaves any unreceived values in ch.
//
// Prefer it to [Chan] for a long-lived source. Because the receive and the
// cancellation are one select, a goroutine parked in this Stream is always
// released by cancelling ctx — including the goroutine a temporal operator uses
// to read its source, which [Chan] leaves parked until ch produces or closes.
// Pass the same context to the source and to the operator reading it.
func ChanContext[T any](ctx context.Context, ch <-chan T) Stream[T] {
	return func(yield func(T) bool) {
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-ch:
				// A buffered value and a done context are both live cases, and
				// select breaks that tie at random. Cancellation must win it,
				// so ctx is rechecked before the value is emitted; the value
				// itself is discarded, like anything else cancellation cuts off.
				if !ok || ctx.Err() != nil || !yield(v) {
					return
				}
			}
		}
	}
}

// Range returns a Stream over the integers in [start, end).
func Range(start, end int) Stream[int] {
	return func(yield func(int) bool) {
		for i := start; i < end; i++ {
			if !yield(i) {
				return
			}
		}
	}
}

// Repeat returns a Stream containing value n times. If n is negative, the
// Stream is infinite.
func Repeat[T any](value T, n int) Stream[T] {
	return func(yield func(T) bool) {
		for i := 0; n < 0 || i < n; i++ {
			if !yield(value) {
				return
			}
		}
	}
}

// Iterate returns an infinite Stream of seed, next(seed), next(next(seed)), ...
func Iterate[T any](seed T, next func(T) T) Stream[T] {
	return func(yield func(T) bool) {
		for v := seed; ; v = next(v) {
			if !yield(v) {
				return
			}
		}
	}
}

// Generate returns an infinite Stream of values produced by fn.
func Generate[T any](fn func() T) Stream[T] {
	return func(yield func(T) bool) {
		for {
			if !yield(fn()) {
				return
			}
		}
	}
}

// --- intermediate operations ---

// Filter returns a Stream of the elements for which keep reports true.
func (s Stream[T]) Filter(keep func(T) bool) Stream[T] {
	return func(yield func(T) bool) {
		for v := range s {
			if keep(v) && !yield(v) {
				return
			}
		}
	}
}

// Map returns a Stream of the results of applying fn to each element.
func (s Stream[T]) Map[R any](fn func(T) R) Stream[R] {
	return func(yield func(R) bool) {
		for v := range s {
			if !yield(fn(v)) {
				return
			}
		}
	}
}

// FlatMap returns a Stream of the elements of every Stream produced by fn.
func (s Stream[T]) FlatMap[R any](fn func(T) Stream[R]) Stream[R] {
	return func(yield func(R) bool) {
		for v := range s {
			for r := range fn(v) {
				if !yield(r) {
					return
				}
			}
		}
	}
}

// Scan returns a Stream of the successive accumulated values, starting with the
// result of combining init with the first element. It is Fold that yields every
// intermediate accumulator rather than only the last.
func (s Stream[T]) Scan[A any](init A, fn func(A, T) A) Stream[A] {
	return func(yield func(A) bool) {
		acc := init
		for v := range s {
			acc = fn(acc, v)
			if !yield(acc) {
				return
			}
		}
	}
}

// DistinctBy returns a Stream omitting elements whose key has already been
// seen. Keys are held for the duration of the iteration.
func (s Stream[T]) DistinctBy[K comparable](key func(T) K) Stream[T] {
	return func(yield func(T) bool) {
		seen := make(map[K]struct{})
		for v := range s {
			k := key(v)
			if _, dup := seen[k]; dup {
				continue
			}
			seen[k] = struct{}{}
			if !yield(v) {
				return
			}
		}
	}
}

// Take returns a Stream of at most the first n elements.
func (s Stream[T]) Take(n int) Stream[T] {
	return func(yield func(T) bool) {
		if n <= 0 {
			return
		}
		taken := 0
		for v := range s {
			if !yield(v) {
				return
			}
			if taken++; taken >= n {
				return
			}
		}
	}
}

// Drop returns a Stream omitting the first n elements.
func (s Stream[T]) Drop(n int) Stream[T] {
	return func(yield func(T) bool) {
		dropped := 0
		for v := range s {
			if dropped < n {
				dropped++
				continue
			}
			if !yield(v) {
				return
			}
		}
	}
}

// TakeWhile returns a Stream of the leading elements for which pred reports
// true, stopping at the first element for which it reports false.
func (s Stream[T]) TakeWhile(pred func(T) bool) Stream[T] {
	return func(yield func(T) bool) {
		for v := range s {
			if !pred(v) || !yield(v) {
				return
			}
		}
	}
}

// DropWhile returns a Stream omitting the leading elements for which pred
// reports true.
func (s Stream[T]) DropWhile(pred func(T) bool) Stream[T] {
	return func(yield func(T) bool) {
		dropping := true
		for v := range s {
			if dropping {
				if pred(v) {
					continue
				}
				dropping = false
			}
			if !yield(v) {
				return
			}
		}
	}
}

// SortFunc returns a Stream of the elements sorted by compare. It buffers the
// whole sequence, so it is not usable on an infinite Stream.
func (s Stream[T]) SortFunc(compare func(a, b T) int) Stream[T] {
	return func(yield func(T) bool) {
		for _, v := range slices.SortedFunc(iter.Seq[T](s), compare) {
			if !yield(v) {
				return
			}
		}
	}
}

// SortStableFunc is like [Stream.SortFunc] but keeps equal elements in their
// original order.
func (s Stream[T]) SortStableFunc(compare func(a, b T) int) Stream[T] {
	return func(yield func(T) bool) {
		for _, v := range slices.SortedStableFunc(iter.Seq[T](s), compare) {
			if !yield(v) {
				return
			}
		}
	}
}

// CompactFunc returns a Stream omitting each element that eq reports equal to
// the element before it. Like slices.CompactFunc, it removes only adjacent
// duplicates; use [Distinct] or [Stream.DistinctBy] to remove all of them.
func (s Stream[T]) CompactFunc(eq func(a, b T) bool) Stream[T] {
	return func(yield func(T) bool) {
		var prev T
		first := true
		for v := range s {
			if !first && eq(prev, v) {
				continue
			}
			prev, first = v, false
			if !yield(v) {
				return
			}
		}
	}
}

// Reverse returns a Stream of the elements in reverse order. It buffers the
// whole sequence, so it is not usable on an infinite Stream.
func (s Stream[T]) Reverse() Stream[T] {
	return func(yield func(T) bool) {
		for _, v := range slices.Backward(s.Collect()) {
			if !yield(v) {
				return
			}
		}
	}
}

// Peek returns a Stream that calls fn for each element as it passes through.
// It is intended for tracing a pipeline, not for mutating it.
func (s Stream[T]) Peek(fn func(T)) Stream[T] {
	return func(yield func(T) bool) {
		for v := range s {
			fn(v)
			if !yield(v) {
				return
			}
		}
	}
}

// Zip returns a Stream2 pairing each element of s with the element of o at the
// same position, ending when either is exhausted.
func (s Stream[T]) Zip[U any](o Stream[U]) Stream2[T, U] {
	return func(yield func(T, U) bool) {
		next, stop := iter.Pull(iter.Seq[U](o))
		defer stop()
		for a := range s {
			b, ok := next()
			if !ok || !yield(a, b) {
				return
			}
		}
	}
}

// ZipWith returns a Stream of fn applied to the elements of s and o at the same
// position, ending when either is exhausted. It is [Stream.Zip] followed by
// [Stream2.Collapse]; neither materialises a pair, since a Stream2 passes its
// two values to yield separately.
func (s Stream[T]) ZipWith[U, R any](o Stream[U], fn func(T, U) R) Stream[R] {
	return s.Zip(o).Collapse(fn)
}

// KeyBy returns a Stream2 pairing each element with the key that key derives
// from it. It is the general way from a Stream to a Stream2, as
// [Stream2.Collapse] is the general way back; [Stream.Enumerate] and
// [Stream.Zip] are the two pairings a key function cannot express.
func (s Stream[T]) KeyBy[K any](key func(T) K) Stream2[K, T] {
	return func(yield func(K, T) bool) {
		for v := range s {
			if !yield(key(v), v) {
				return
			}
		}
	}
}

// Enumerate returns a Stream2 pairing each element with its zero-based index.
func (s Stream[T]) Enumerate() Stream2[int, T] {
	return func(yield func(int, T) bool) {
		i := 0
		for v := range s {
			if !yield(i, v) {
				return
			}
			i++
		}
	}
}
