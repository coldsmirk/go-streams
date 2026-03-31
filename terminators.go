package streams

import (
	"slices"
	"strings"
)

// --- Terminal Operations (consume Stream, eager) ---

// ForEach executes the action on each element.
func (s Stream[T]) ForEach(action func(T)) {
	for v := range s.seq {
		action(v)
	}
}

// ForEachIndexed executes the action on each element with its index.
func (s Stream[T]) ForEachIndexed(action func(int, T)) {
	idx := 0
	for v := range s.seq {
		action(idx, v)
		idx++
	}
}

// ForEachErr executes the action on each element, returning the first error encountered.
// If the action returns an error, iteration stops immediately and the error is returned.
func (s Stream[T]) ForEachErr(action func(T) error) error {
	for v := range s.seq {
		if err := action(v); err != nil {
			return err
		}
	}
	return nil
}

// ForEachIndexedErr executes the action on each element with its index, returning the first error encountered.
// If the action returns an error, iteration stops immediately and the error is returned.
func (s Stream[T]) ForEachIndexedErr(action func(int, T) error) error {
	idx := 0
	for v := range s.seq {
		if err := action(idx, v); err != nil {
			return err
		}
		idx++
	}
	return nil
}

// Collect gathers all elements into a slice.
func (s Stream[T]) Collect() []T {
	return slices.Collect(s.seq)
}

// Reduce combines all elements using the given function.
// Returns identity if the stream is empty.
func (s Stream[T]) Reduce(identity T, fn func(T, T) T) T {
	result := identity
	for v := range s.seq {
		result = fn(result, v)
	}
	return result
}

// ReduceOptional combines all elements using the given function.
// Returns None if the stream is empty.
func (s Stream[T]) ReduceOptional(fn func(T, T) T) Optional[T] {
	var result T
	first := true
	for v := range s.seq {
		if first {
			result = v
			first = false
		} else {
			result = fn(result, v)
		}
	}
	if first {
		return None[T]()
	}
	return Some(result)
}

// Fold is an alias for Reduce.
func (s Stream[T]) Fold(identity T, fn func(T, T) T) T {
	return s.Reduce(identity, fn)
}

// Count returns the number of elements in the stream.
func (s Stream[T]) Count() int {
	count := 0
	for range s.seq {
		count++
	}
	return count
}

// First returns the first element as an Optional.
func (s Stream[T]) First() Optional[T] {
	for v := range s.seq {
		return Some(v)
	}
	return None[T]()
}

// Last returns the last element as an Optional.
func (s Stream[T]) Last() Optional[T] {
	var last T
	found := false
	for v := range s.seq {
		last = v
		found = true
	}
	if found {
		return Some(last)
	}
	return None[T]()
}

// FindFirst returns the first element that matches the predicate.
func (s Stream[T]) FindFirst(pred func(T) bool) Optional[T] {
	for v := range s.seq {
		if pred(v) {
			return Some(v)
		}
	}
	return None[T]()
}

// FindLast returns the last element that matches the predicate.
func (s Stream[T]) FindLast(pred func(T) bool) Optional[T] {
	var last T
	found := false
	for v := range s.seq {
		if pred(v) {
			last = v
			found = true
		}
	}
	if found {
		return Some(last)
	}
	return None[T]()
}

// AnyMatch returns true if any element matches the predicate.
func (s Stream[T]) AnyMatch(pred func(T) bool) bool {
	for v := range s.seq {
		if pred(v) {
			return true
		}
	}
	return false
}

// AllMatch returns true if all elements match the predicate.
// Returns true for an empty stream.
func (s Stream[T]) AllMatch(pred func(T) bool) bool {
	for v := range s.seq {
		if !pred(v) {
			return false
		}
	}
	return true
}

// NoneMatch returns true if no elements match the predicate.
// Returns true for an empty stream.
func (s Stream[T]) NoneMatch(pred func(T) bool) bool {
	for v := range s.seq {
		if pred(v) {
			return false
		}
	}
	return true
}

// Contains returns true if the stream contains the target element.
// Elements must be comparable.
func Contains[T comparable](s Stream[T], target T) bool {
	for v := range s.seq {
		if v == target {
			return true
		}
	}
	return false
}

// Min returns the minimum element using the comparison function.
func (s Stream[T]) Min(cmp func(T, T) int) Optional[T] {
	var min T
	first := true
	for v := range s.seq {
		if first {
			min = v
			first = false
		} else if cmp(v, min) < 0 {
			min = v
		}
	}
	if first {
		return None[T]()
	}
	return Some(min)
}

// Max returns the maximum element using the comparison function.
func (s Stream[T]) Max(cmp func(T, T) int) Optional[T] {
	var max T
	first := true
	for v := range s.seq {
		if first {
			max = v
			first = false
		} else if cmp(v, max) > 0 {
			max = v
		}
	}
	if first {
		return None[T]()
	}
	return Some(max)
}

// --- Free Functions for terminal operations with type transformation ---

// FoldTo reduces Stream[T] to type R with an identity value.
func FoldTo[T, R any](s Stream[T], identity R, fn func(R, T) R) R {
	result := identity
	for v := range s.seq {
		result = fn(result, v)
	}
	return result
}

// ToMap collects Stream into a map using a function that returns key-value pairs.
func ToMap[T any, K comparable, V any](s Stream[T], fn func(T) (K, V)) map[K]V {
	result := make(map[K]V)
	for v := range s.seq {
		k, val := fn(v)
		result[k] = val
	}
	return result
}

// ToSet collects Stream into a set (map with struct{} values).
func ToSet[T comparable](s Stream[T]) map[T]struct{} {
	result := make(map[T]struct{})
	for v := range s.seq {
		result[v] = struct{}{}
	}
	return result
}

// GroupBy groups elements by a key function.
func GroupBy[T any, K comparable](s Stream[T], keyFn func(T) K) map[K][]T {
	result := make(map[K][]T)
	for v := range s.seq {
		k := keyFn(v)
		result[k] = append(result[k], v)
	}
	return result
}

// GroupByTo groups elements by a function that returns key-value pairs, collecting values into slices.
func GroupByTo[T any, K comparable, V any](s Stream[T], fn func(T) (K, V)) map[K][]V {
	result := make(map[K][]V)
	for v := range s.seq {
		k, val := fn(v)
		result[k] = append(result[k], val)
	}
	return result
}

// PartitionBy splits elements into two groups based on a predicate.
// Returns (matching, notMatching).
func PartitionBy[T any](s Stream[T], pred func(T) bool) ([]T, []T) {
	var matching, notMatching []T
	for v := range s.seq {
		if pred(v) {
			matching = append(matching, v)
		} else {
			notMatching = append(notMatching, v)
		}
	}
	return matching, notMatching
}

// Joining concatenates string elements with a separator.
// Uses strings.Builder for O(n) performance.
func Joining(s Stream[string], sep string) string {
	var b strings.Builder
	first := true
	for v := range s.seq {
		if !first {
			b.WriteString(sep)
		}
		b.WriteString(v)
		first = false
	}
	return b.String()
}

// JoiningWithPrefixSuffix concatenates string elements with separator, prefix, and suffix.
func JoiningWithPrefixSuffix(s Stream[string], sep, prefix, suffix string) string {
	return prefix + Joining(s, sep) + suffix
}

// AssociateBy creates a map using element as value and a key function.
func AssociateBy[T any, K comparable](s Stream[T], keyFn func(T) K) map[K]T {
	result := make(map[K]T)
	for v := range s.seq {
		result[keyFn(v)] = v
	}
	return result
}

// IndexBy is an alias for AssociateBy.
func IndexBy[T any, K comparable](s Stream[T], keyFn func(T) K) map[K]T {
	return AssociateBy(s, keyFn)
}

// CountBy counts elements by a key function.
func CountBy[T any, K comparable](s Stream[T], keyFn func(T) K) map[K]int {
	result := make(map[K]int)
	for v := range s.seq {
		result[keyFn(v)]++
	}
	return result
}

// Frequencies counts occurrences of each element.
func Frequencies[T comparable](s Stream[T]) map[T]int {
	result := make(map[T]int)
	for v := range s.seq {
		result[v]++
	}
	return result
}

// At returns the element at the specified index, or None if out of bounds.
func (s Stream[T]) At(index int) Optional[T] {
	if index < 0 {
		return None[T]()
	}
	current := 0
	for v := range s.seq {
		if current == index {
			return Some(v)
		}
		current++
	}
	return None[T]()
}

// Nth returns the element at the specified index (0-based).
// Alias for At.
func (s Stream[T]) Nth(index int) Optional[T] {
	return s.At(index)
}

// Single returns the only element if there is exactly one, otherwise None.
func (s Stream[T]) Single() Optional[T] {
	var result T
	count := 0
	for v := range s.seq {
		result = v
		count++
		if count > 1 {
			return None[T]()
		}
	}
	if count == 1 {
		return Some(result)
	}
	return None[T]()
}

// IsEmpty returns true if the stream has no elements.
func (s Stream[T]) IsEmpty() bool {
	for range s.seq {
		return false
	}
	return true
}

// IsNotEmpty returns true if the stream has at least one element.
func (s Stream[T]) IsNotEmpty() bool {
	return !s.IsEmpty()
}
