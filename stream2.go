package streams

import (
	"iter"
	"maps"
	"slices"
)

// Stream2 is a lazy sequence of key-value pairs.
// It wraps iter.Seq2[K, V] and provides fluent functional programming operations.
type Stream2[K, V any] struct {
	seq iter.Seq2[K, V]
}

// From2 creates a Stream2 from an iter.Seq2.
// This provides interoperability with the standard library.
func From2[K, V any](seq iter.Seq2[K, V]) Stream2[K, V] {
	return Stream2[K, V]{seq: seq}
}

// Seq2 returns the underlying iter.Seq2 for stdlib interop.
func (s Stream2[K, V]) Seq2() iter.Seq2[K, V] {
	return s.seq
}

// --- Constructors ---

// PairsOf creates a Stream2 from variadic Pair values.
func PairsOf[K, V any](pairs ...Pair[K, V]) Stream2[K, V] {
	return Stream2[K, V]{
		seq: func(yield func(K, V) bool) {
			for _, p := range pairs {
				if !yield(p.First, p.Second) {
					return
				}
			}
		},
	}
}

// Empty2 returns an empty Stream2.
func Empty2[K, V any]() Stream2[K, V] {
	return Stream2[K, V]{
		seq: func(yield func(K, V) bool) {},
	}
}

// --- Intermediate Operations ---

// Filter returns a Stream2 containing only pairs that match the predicate.
func (s Stream2[K, V]) Filter(pred func(K, V) bool) Stream2[K, V] {
	return Stream2[K, V]{
		seq: func(yield func(K, V) bool) {
			for k, v := range s.seq {
				if pred(k, v) && !yield(k, v) {
					return
				}
			}
		},
	}
}

// MapKeys transforms the keys using the given function.
func (s Stream2[K, V]) MapKeys(fn func(K) K) Stream2[K, V] {
	return Stream2[K, V]{
		seq: func(yield func(K, V) bool) {
			for k, v := range s.seq {
				if !yield(fn(k), v) {
					return
				}
			}
		},
	}
}

// MapValues transforms the values using the given function.
func (s Stream2[K, V]) MapValues(fn func(V) V) Stream2[K, V] {
	return Stream2[K, V]{
		seq: func(yield func(K, V) bool) {
			for k, v := range s.seq {
				if !yield(k, fn(v)) {
					return
				}
			}
		},
	}
}

// Limit returns a Stream2 containing at most n pairs.
func (s Stream2[K, V]) Limit(n int) Stream2[K, V] {
	return Stream2[K, V]{
		seq: func(yield func(K, V) bool) {
			count := 0
			for k, v := range s.seq {
				if count >= n {
					return
				}
				if !yield(k, v) {
					return
				}
				count++
			}
		},
	}
}

// Skip returns a Stream2 that skips the first n pairs.
func (s Stream2[K, V]) Skip(n int) Stream2[K, V] {
	return Stream2[K, V]{
		seq: func(yield func(K, V) bool) {
			count := 0
			for k, v := range s.seq {
				if count < n {
					count++
					continue
				}
				if !yield(k, v) {
					return
				}
			}
		},
	}
}

// Peek performs the given action on each pair as it passes through.
func (s Stream2[K, V]) Peek(action func(K, V)) Stream2[K, V] {
	return Stream2[K, V]{
		seq: func(yield func(K, V) bool) {
			for k, v := range s.seq {
				action(k, v)
				if !yield(k, v) {
					return
				}
			}
		},
	}
}

// TakeWhile returns a Stream2 that yields pairs while the predicate is true.
func (s Stream2[K, V]) TakeWhile(pred func(K, V) bool) Stream2[K, V] {
	return Stream2[K, V]{
		seq: func(yield func(K, V) bool) {
			for k, v := range s.seq {
				if !pred(k, v) {
					return
				}
				if !yield(k, v) {
					return
				}
			}
		},
	}
}

// DropWhile returns a Stream2 that skips pairs while the predicate is true.
func (s Stream2[K, V]) DropWhile(pred func(K, V) bool) Stream2[K, V] {
	return Stream2[K, V]{
		seq: func(yield func(K, V) bool) {
			dropping := true
			for k, v := range s.seq {
				if dropping {
					if pred(k, v) {
						continue
					}
					dropping = false
				}
				if !yield(k, v) {
					return
				}
			}
		},
	}
}

// --- Terminal Operations ---

// Keys returns a Stream containing only the keys.
func (s Stream2[K, V]) Keys() Stream[K] {
	return Stream[K]{
		seq: func(yield func(K) bool) {
			for k := range s.seq {
				if !yield(k) {
					return
				}
			}
		},
	}
}

// Values returns a Stream containing only the values.
func (s Stream2[K, V]) Values() Stream[V] {
	return Stream[V]{
		seq: func(yield func(V) bool) {
			for _, v := range s.seq {
				if !yield(v) {
					return
				}
			}
		},
	}
}

// ToPairs returns a Stream of Pair[K, V].
func (s Stream2[K, V]) ToPairs() Stream[Pair[K, V]] {
	return Stream[Pair[K, V]]{
		seq: func(yield func(Pair[K, V]) bool) {
			for k, v := range s.seq {
				if !yield(Pair[K, V]{First: k, Second: v}) {
					return
				}
			}
		},
	}
}

// ForEach executes the action on each key-value pair.
func (s Stream2[K, V]) ForEach(action func(K, V)) {
	for k, v := range s.seq {
		action(k, v)
	}
}

// Count returns the number of pairs in the stream.
func (s Stream2[K, V]) Count() int {
	count := 0
	for range s.seq {
		count++
	}
	return count
}

// AnyMatch returns true if any pair matches the predicate.
func (s Stream2[K, V]) AnyMatch(pred func(K, V) bool) bool {
	for k, v := range s.seq {
		if pred(k, v) {
			return true
		}
	}
	return false
}

// AllMatch returns true if all pairs match the predicate.
// Returns true for an empty stream.
func (s Stream2[K, V]) AllMatch(pred func(K, V) bool) bool {
	for k, v := range s.seq {
		if !pred(k, v) {
			return false
		}
	}
	return true
}

// NoneMatch returns true if no pairs match the predicate.
// Returns true for an empty stream.
func (s Stream2[K, V]) NoneMatch(pred func(K, V) bool) bool {
	for k, v := range s.seq {
		if pred(k, v) {
			return false
		}
	}
	return true
}

// First returns the first pair as an Optional.
func (s Stream2[K, V]) First() Optional[Pair[K, V]] {
	for k, v := range s.seq {
		return Some(Pair[K, V]{First: k, Second: v})
	}
	return None[Pair[K, V]]()
}

// CollectPairs collects all pairs into a slice of Pairs.
func (s Stream2[K, V]) CollectPairs() []Pair[K, V] {
	return slices.Collect(s.ToPairs().Seq())
}

// Reduce combines all pairs into a single value.
func (s Stream2[K, V]) Reduce(identity Pair[K, V], fn func(Pair[K, V], K, V) Pair[K, V]) Pair[K, V] {
	result := identity
	for k, v := range s.seq {
		result = fn(result, k, v)
	}
	return result
}

// --- Free Functions for Stream2 type transformation ---

// MapKeysTo transforms Stream2[K, V] to Stream2[K2, V].
func MapKeysTo[K, V, K2 any](s Stream2[K, V], fn func(K) K2) Stream2[K2, V] {
	return Stream2[K2, V]{
		seq: func(yield func(K2, V) bool) {
			for k, v := range s.seq {
				if !yield(fn(k), v) {
					return
				}
			}
		},
	}
}

// MapValuesTo transforms Stream2[K, V] to Stream2[K, V2].
func MapValuesTo[K, V, V2 any](s Stream2[K, V], fn func(V) V2) Stream2[K, V2] {
	return Stream2[K, V2]{
		seq: func(yield func(K, V2) bool) {
			for k, v := range s.seq {
				if !yield(k, fn(v)) {
					return
				}
			}
		},
	}
}

// ParallelMapValues transforms the values using the given function in parallel.
func (s Stream2[K, V]) ParallelMapValues(fn func(V) V, opts ...ParallelOption) Stream2[K, V] {
	return Stream2[K, V]{
		seq: func(yield func(K, V) bool) {
			mapped := ParallelMap(s.ToPairs(), func(p Pair[K, V]) Pair[K, V] {
				return Pair[K, V]{First: p.First, Second: fn(p.Second)}
			}, opts...)

			for p := range mapped.Seq() {
				if !yield(p.First, p.Second) {
					return
				}
			}
		},
	}
}

// ParallelFilter returns a Stream2 containing only pairs that match the predicate in parallel.
func (s Stream2[K, V]) ParallelFilter(pred func(K, V) bool, opts ...ParallelOption) Stream2[K, V] {
	return Stream2[K, V]{
		seq: func(yield func(K, V) bool) {
			filtered := ParallelFilter(s.ToPairs(), func(p Pair[K, V]) bool {
				return pred(p.First, p.Second)
			}, opts...)

			for p := range filtered.Seq() {
				if !yield(p.First, p.Second) {
					return
				}
			}
		},
	}
}

// MapPairs transforms Stream2[K, V] to Stream2[K2, V2].
func MapPairs[K, V, K2, V2 any](s Stream2[K, V], fn func(K, V) (K2, V2)) Stream2[K2, V2] {
	return Stream2[K2, V2]{
		seq: func(yield func(K2, V2) bool) {
			for k, v := range s.seq {
				k2, v2 := fn(k, v)
				if !yield(k2, v2) {
					return
				}
			}
		},
	}
}

// MapPairsTo transforms each key-value pair of a Stream2 into a single value, returning a Stream[R].
func MapPairsTo[K, V, R any](s Stream2[K, V], fn func(K, V) R) Stream[R] {
	return Stream[R]{
		seq: func(yield func(R) bool) {
			for k, v := range s.seq {
				if !yield(fn(k, v)) {
					return
				}
			}
		},
	}
}

// SwapKeyValue swaps keys and values in a Stream2.
func SwapKeyValue[K, V any](s Stream2[K, V]) Stream2[V, K] {
	return Stream2[V, K]{
		seq: func(yield func(V, K) bool) {
			for k, v := range s.seq {
				if !yield(v, k) {
					return
				}
			}
		},
	}
}

// ToMap2 collects Stream2 into a map. Keys must be comparable.
func ToMap2[K comparable, V any](s Stream2[K, V]) map[K]V {
	return maps.Collect(s.seq)
}

// ReduceByKey groups values by key and reduces each group using the merge function.
// Returns a map where each key maps to the reduced value of all values with that key.
func ReduceByKey[K comparable, V any](s Stream2[K, V], merge func(V, V) V) map[K]V {
	result := make(map[K]V)
	for k, v := range s.seq {
		if cur, ok := result[k]; ok {
			result[k] = merge(cur, v)
		} else {
			result[k] = v
		}
	}
	return result
}

// ReduceByKeyWithInit groups values by key and reduces each group using the merge function.
// Uses init as the initial value for each key's reduction.
func ReduceByKeyWithInit[K comparable, V, R any](s Stream2[K, V], init func() R, merge func(R, V) R) map[K]R {
	result := make(map[K]R)
	for k, v := range s.seq {
		if _, ok := result[k]; !ok {
			result[k] = init()
		}
		result[k] = merge(result[k], v)
	}
	return result
}

// GroupValues groups all values by their keys into slices.
// Returns a map where each key maps to a slice of all values with that key.
func GroupValues[K comparable, V any](s Stream2[K, V]) map[K][]V {
	result := make(map[K][]V)
	for k, v := range s.seq {
		result[k] = append(result[k], v)
	}
	return result
}

// DistinctKeys returns a Stream2 with duplicate keys removed.
// Only the first occurrence of each key is kept.
func DistinctKeys[K comparable, V any](s Stream2[K, V]) Stream2[K, V] {
	return Stream2[K, V]{
		seq: func(yield func(K, V) bool) {
			seen := make(map[K]struct{})
			for k, v := range s.seq {
				if _, exists := seen[k]; !exists {
					seen[k] = struct{}{}
					if !yield(k, v) {
						return
					}
				}
			}
		},
	}
}

// DistinctValues returns a Stream2 with duplicate values removed.
// Only the first occurrence of each value is kept.
func DistinctValues[K any, V comparable](s Stream2[K, V]) Stream2[K, V] {
	return Stream2[K, V]{
		seq: func(yield func(K, V) bool) {
			seen := make(map[V]struct{})
			for k, v := range s.seq {
				if _, exists := seen[v]; !exists {
					seen[v] = struct{}{}
					if !yield(k, v) {
						return
					}
				}
			}
		},
	}
}
