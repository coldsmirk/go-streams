package streams

import (
	"iter"
	"maps"
)

// Stream2 is a lazy sequence of paired values, conventionally key-value. Its
// underlying type is iter.Seq2[K, V].
type Stream2[K, V any] iter.Seq2[K, V]

// From2 returns seq as a Stream2. Like [From], it exists to infer K and V.
func From2[K, V any](seq iter.Seq2[K, V]) Stream2[K, V] { return Stream2[K, V](seq) }

// Pairs returns a Stream2 over the key-value pairs of m, in unspecified order.
// An operation that depends on encounter order, such as [Stream2.First] or
// [Stream2.Take], therefore gives an arbitrary result over the Stream2 it
// returns.
func Pairs[M ~map[K]V, K comparable, V any](m M) Stream2[K, V] {
	return Stream2[K, V](maps.All(m))
}

// Empty2 returns a Stream2 with no elements.
func Empty2[K, V any]() Stream2[K, V] { return func(func(K, V) bool) {} }

// Keys returns a Stream of the first element of each pair.
func (s Stream2[K, V]) Keys() Stream[K] {
	return func(yield func(K) bool) {
		for k := range s {
			if !yield(k) {
				return
			}
		}
	}
}

// Values returns a Stream of the second element of each pair.
func (s Stream2[K, V]) Values() Stream[V] {
	return func(yield func(V) bool) {
		for _, v := range s {
			if !yield(v) {
				return
			}
		}
	}
}

// Filter returns a Stream2 of the pairs for which keep reports true.
func (s Stream2[K, V]) Filter(keep func(K, V) bool) Stream2[K, V] {
	return func(yield func(K, V) bool) {
		for k, v := range s {
			if keep(k, v) && !yield(k, v) {
				return
			}
		}
	}
}

// MapKeys returns a Stream2 with fn applied to each key.
func (s Stream2[K, V]) MapKeys[J any](fn func(K) J) Stream2[J, V] {
	return func(yield func(J, V) bool) {
		for k, v := range s {
			if !yield(fn(k), v) {
				return
			}
		}
	}
}

// MapValues returns a Stream2 with fn applied to each value.
func (s Stream2[K, V]) MapValues[W any](fn func(V) W) Stream2[K, W] {
	return func(yield func(K, W) bool) {
		for k, v := range s {
			if !yield(k, fn(v)) {
				return
			}
		}
	}
}

// Collapse returns a Stream of fn applied to each pair. It is the way back from
// a Stream2 to a Stream.
func (s Stream2[K, V]) Collapse[R any](fn func(K, V) R) Stream[R] {
	return func(yield func(R) bool) {
		for k, v := range s {
			if !yield(fn(k, v)) {
				return
			}
		}
	}
}

// Swap returns a Stream2 with the elements of each pair exchanged.
func (s Stream2[K, V]) Swap() Stream2[V, K] {
	return func(yield func(V, K) bool) {
		for k, v := range s {
			if !yield(v, k) {
				return
			}
		}
	}
}

// Take returns a Stream2 of at most the first n pairs.
func (s Stream2[K, V]) Take(n int) Stream2[K, V] {
	return func(yield func(K, V) bool) {
		if n <= 0 {
			return
		}
		taken := 0
		for k, v := range s {
			if !yield(k, v) {
				return
			}
			if taken++; taken >= n {
				return
			}
		}
	}
}

// Drop returns a Stream2 omitting the first n pairs.
func (s Stream2[K, V]) Drop(n int) Stream2[K, V] {
	return func(yield func(K, V) bool) {
		dropped := 0
		for k, v := range s {
			if dropped < n {
				dropped++
				continue
			}
			if !yield(k, v) {
				return
			}
		}
	}
}

// --- terminal operations ---

// ForEach calls fn for each pair.
func (s Stream2[K, V]) ForEach(fn func(K, V)) {
	for k, v := range s {
		fn(k, v)
	}
}

// Count returns the number of pairs.
func (s Stream2[K, V]) Count() int {
	n := 0
	for range s {
		n++
	}
	return n
}

// Fold accumulates the pairs into a single value, starting from init.
func (s Stream2[K, V]) Fold[A any](init A, fn func(A, K, V) A) A {
	acc := init
	for k, v := range s {
		acc = fn(acc, k, v)
	}
	return acc
}

// First returns the first pair, or false if the Stream2 is empty. It stops
// after the first pair.
func (s Stream2[K, V]) First() (K, V, bool) {
	for k, v := range s {
		return k, v, true
	}
	var zeroK K
	var zeroV V
	return zeroK, zeroV, false
}

// Last returns the last pair, or false if the Stream2 is empty.
func (s Stream2[K, V]) Last() (K, V, bool) {
	var lastK K
	var lastV V
	empty := true
	for k, v := range s {
		lastK, lastV, empty = k, v, false
	}
	return lastK, lastV, !empty
}

// Find returns the first pair for which pred reports true, or false if none
// does. It stops at the first match.
func (s Stream2[K, V]) Find(pred func(K, V) bool) (K, V, bool) {
	for k, v := range s {
		if pred(k, v) {
			return k, v, true
		}
	}
	var zeroK K
	var zeroV V
	return zeroK, zeroV, false
}

// Any reports whether pred is true for at least one pair. It stops at the
// first match.
func (s Stream2[K, V]) Any(pred func(K, V) bool) bool {
	for k, v := range s {
		if pred(k, v) {
			return true
		}
	}
	return false
}

// All reports whether pred is true for every pair. It stops at the first pair
// that fails, and is true for an empty Stream2.
func (s Stream2[K, V]) All(pred func(K, V) bool) bool {
	for k, v := range s {
		if !pred(k, v) {
			return false
		}
	}
	return true
}

// CollectMap returns a map of the pairs of s. A later pair overwrites an
// earlier one with the same key. It is the counterpart of [Pairs].
func CollectMap[K comparable, V any](s Stream2[K, V]) map[K]V {
	return maps.Collect(iter.Seq2[K, V](s))
}
