package streams

import (
	"iter"
	"slices"
)

// Collect returns a slice of the elements.
func (s Stream[T]) Collect() []T { return slices.Collect(iter.Seq[T](s)) }

// ForEach calls fn for each element.
func (s Stream[T]) ForEach(fn func(T)) {
	for v := range s {
		fn(v)
	}
}

// Count returns the number of elements.
func (s Stream[T]) Count() int {
	n := 0
	for range s {
		n++
	}
	return n
}

// Fold accumulates the elements into a single value of any type, starting from
// init. Unlike [Stream.Reduce], the accumulator need not be the element type.
func (s Stream[T]) Fold[A any](init A, fn func(A, T) A) A {
	acc := init
	for v := range s {
		acc = fn(acc, v)
	}
	return acc
}

// Reduce combines the elements using fn, returning false if the Stream is
// empty.
func (s Stream[T]) Reduce(fn func(a, b T) T) (T, bool) {
	var acc T
	empty := true
	for v := range s {
		if empty {
			acc, empty = v, false
			continue
		}
		acc = fn(acc, v)
	}
	return acc, !empty
}

// First returns the first element, or false if the Stream is empty. It stops
// after the first element.
func (s Stream[T]) First() (T, bool) {
	for v := range s {
		return v, true
	}
	var zero T
	return zero, false
}

// Last returns the last element, or false if the Stream is empty.
func (s Stream[T]) Last() (T, bool) {
	var last T
	empty := true
	for v := range s {
		last, empty = v, false
	}
	return last, !empty
}

// Find returns the first element for which pred reports true, or false if none
// does. It stops at the first match.
func (s Stream[T]) Find(pred func(T) bool) (T, bool) {
	for v := range s {
		if pred(v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}

// Any reports whether pred is true for at least one element. It stops at the
// first match.
func (s Stream[T]) Any(pred func(T) bool) bool {
	for v := range s {
		if pred(v) {
			return true
		}
	}
	return false
}

// All reports whether pred is true for every element. It stops at the first
// element that fails, and is true for an empty Stream.
func (s Stream[T]) All(pred func(T) bool) bool {
	for v := range s {
		if !pred(v) {
			return false
		}
	}
	return true
}

// MinFunc returns the minimal element as ordered by compare, or false if the
// Stream is empty. If several elements are minimal, it returns the first.
func (s Stream[T]) MinFunc(compare func(a, b T) int) (T, bool) {
	var best T
	empty := true
	for v := range s {
		if empty || compare(v, best) < 0 {
			best, empty = v, false
		}
	}
	return best, !empty
}

// MaxFunc returns the maximal element as ordered by compare, or false if the
// Stream is empty. If several elements are maximal, it returns the first.
func (s Stream[T]) MaxFunc(compare func(a, b T) int) (T, bool) {
	var best T
	empty := true
	for v := range s {
		if empty || compare(v, best) > 0 {
			best, empty = v, false
		}
	}
	return best, !empty
}

// GroupBy returns a map from each key to the elements that produced it, in
// encounter order.
func (s Stream[T]) GroupBy[K comparable](key func(T) K) map[K][]T {
	out := make(map[K][]T)
	for v := range s {
		k := key(v)
		out[k] = append(out[k], v)
	}
	return out
}

// IndexBy returns a map from each key to the last element that produced it.
func (s Stream[T]) IndexBy[K comparable](key func(T) K) map[K]T {
	out := make(map[K]T)
	for v := range s {
		out[key(v)] = v
	}
	return out
}

// ToMap returns a map built from the key-value pairs fn derives from each
// element. Later pairs overwrite earlier ones with the same key.
func (s Stream[T]) ToMap[K comparable, V any](fn func(T) (K, V)) map[K]V {
	out := make(map[K]V)
	for v := range s {
		k, val := fn(v)
		out[k] = val
	}
	return out
}

// Partition returns the elements for which pred reports true and false.
func (s Stream[T]) Partition(pred func(T) bool) (yes, no []T) {
	for v := range s {
		if pred(v) {
			yes = append(yes, v)
		} else {
			no = append(no, v)
		}
	}
	return yes, no
}
