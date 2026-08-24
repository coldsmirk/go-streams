// Package join provides relational joins over keyed streams.
//
// A join takes a combiner function and returns a [streams.Stream] of whatever
// that function produces, in the manner of [streams.Stream.ZipWith], so the
// package declares no result type of its own. An outer join hands the combiner
// a presence flag beside the value it describes; where the flag is false the
// value is the zero value of its type.
//
// Every join buffers exactly one side, b, and streams the other, a, so a may be
// infinite. [Right] and [Full] are the exceptions that still need a to end,
// because the rows of b that went unmatched are known only once a is exhausted,
// and [Group] buffers both sides because it cannot emit a key before it has
// seen every row carrying it. Each doc comment states what its join holds in
// memory.
//
// Where a key carries several rows on both sides, a join yields their cartesian
// product for that key.
//
// Two unkeyed streams join by deriving their keys first, which every join
// accepts uniformly:
//
//	join.Left(orders.KeyBy(byCustomer), customers.KeyBy(byID),
//		func(id CustomerID, o Order, c Customer, known bool) Row { ... })
//
// Semi and Anti return a Stream2; call Values on the result to get the rows of
// a back without their keys.
package join

import (
	streams "github.com/coldsmirk/go-streams/v2"
)

// Inner returns a Stream of combine applied to every pair of rows of a and b
// that share a key. A row with no match on the other side is dropped, so an
// empty side on either end yields an empty Stream.
//
// Inner buffers b for the duration of the iteration and streams a, which may
// therefore be infinite. Pairs appear in the encounter order of a, and for each
// row of a its matches appear in the encounter order of b, so a key carrying
// several rows on both sides yields the cartesian product for that key.
func Inner[K comparable, A, B, R any](a streams.Stream2[K, A], b streams.Stream2[K, B],
	combine func(key K, left A, right B) R) streams.Stream[R] {
	return func(yield func(R) bool) {
		right := buffer(b)
		for k, l := range a {
			r, seen := right.at[k]
			if !seen {
				continue
			}
			for i := r.first; i >= 0; i = right.next[i] {
				if !yield(combine(k, l, right.values[i])) {
					return
				}
			}
		}
	}
}

// Left returns a Stream of combine applied to every row of a, matched or not.
// The hasRight argument reports whether right holds a row of b; when it is
// false, right is the zero value of B. An empty b therefore yields one
// unmatched result per row of a.
//
// Left buffers b for the duration of the iteration and streams a, which may
// therefore be infinite. Results appear in the encounter order of a, and for
// each row of a its matches appear in the encounter order of b, so a key
// carrying several rows on both sides yields the cartesian product for that
// key.
func Left[K comparable, A, B, R any](a streams.Stream2[K, A], b streams.Stream2[K, B],
	combine func(key K, left A, right B, hasRight bool) R) streams.Stream[R] {
	return func(yield func(R) bool) {
		right := buffer(b)
		for k, l := range a {
			r, seen := right.at[k]
			if !seen {
				var zero B
				if !yield(combine(k, l, zero, false)) {
					return
				}
				continue
			}
			for i := r.first; i >= 0; i = right.next[i] {
				if !yield(combine(k, l, right.values[i], true)) {
					return
				}
			}
		}
	}
}

// Right returns a Stream of combine applied to every row of b, matched or not.
// The hasLeft argument reports whether left holds a row of a; when it is false,
// left is the zero value of A. An empty a therefore yields one unmatched result
// per row of b.
//
// Right buffers b for the duration of the iteration and streams a, as the other
// joins do. The matched rows are consequently emitted as a is traversed, in the
// encounter order of a rather than of b, and the rows of b that no row of a
// matched follow once a is exhausted, in the encounter order of b. Right does
// not terminate on an infinite a. A key carrying several rows on both sides
// yields the cartesian product for that key.
func Right[K comparable, A, B, R any](a streams.Stream2[K, A], b streams.Stream2[K, B],
	combine func(key K, left A, hasLeft bool, right B) R) streams.Stream[R] {
	return func(yield func(R) bool) {
		right := buffer(b)
		matched := make(map[K]struct{})
		for k, l := range a {
			r, seen := right.at[k]
			if !seen {
				continue
			}
			matched[k] = struct{}{}
			for i := r.first; i >= 0; i = right.next[i] {
				if !yield(combine(k, l, true, right.values[i])) {
					return
				}
			}
		}
		var zero A
		for i, k := range right.keys {
			if _, ok := matched[k]; ok {
				continue
			}
			if !yield(combine(k, zero, false, right.values[i])) {
				return
			}
		}
	}
}

// Full returns a Stream of combine applied to every row of a and every row of
// b, matched where they share a key. The hasLeft and hasRight arguments report
// whether left and right hold a row; the value beside a false flag is the zero
// value of its type, and at least one flag is always true.
//
// Full buffers b entirely for the duration of the iteration and streams a. The
// rows of a are emitted as they arrive, each with its matches in b or with
// hasRight false, and the rows of b that no row of a matched follow once a is
// exhausted, in the encounter order of b. Full does not terminate on an
// infinite a. A key carrying several rows on both sides yields the cartesian
// product for that key.
func Full[K comparable, A, B, R any](a streams.Stream2[K, A], b streams.Stream2[K, B],
	combine func(key K, left A, hasLeft bool, right B, hasRight bool) R) streams.Stream[R] {
	return func(yield func(R) bool) {
		right := buffer(b)
		matched := make(map[K]struct{})
		for k, l := range a {
			r, seen := right.at[k]
			if !seen {
				var zero B
				if !yield(combine(k, l, true, zero, false)) {
					return
				}
				continue
			}
			matched[k] = struct{}{}
			for i := r.first; i >= 0; i = right.next[i] {
				if !yield(combine(k, l, true, right.values[i], true)) {
					return
				}
			}
		}
		var zero A
		for i, k := range right.keys {
			if _, ok := matched[k]; ok {
				continue
			}
			if !yield(combine(k, zero, false, right.values[i], true)) {
				return
			}
		}
	}
}

// Group returns a Stream of combine applied once per distinct key of a or b,
// with the rows carrying that key on each side, in encounter order. A key that
// only one side carries is combined with a nil slice for the other, so the two
// slices are never both empty.
//
// Grouping cannot emit a key before both sides are complete, so Group is the
// one join here that buffers a as well as b, and neither may be infinite. Keys
// appear in the order a first encountered them, followed by the keys only b
// carries, in the order b first encountered them. The slices are freshly
// allocated for each result.
func Group[K comparable, A, B, R any](a streams.Stream2[K, A], b streams.Stream2[K, B],
	combine func(key K, left []A, right []B) R) streams.Stream[R] {
	return func(yield func(R) bool) {
		left, right := buffer(a), buffer(b)
		for i, k := range left.keys {
			if left.at[k].first != i {
				continue // not this key's first row; it was emitted already
			}
			if !yield(combine(k, left.rows(k), right.rows(k))) {
				return
			}
		}
		for i, k := range right.keys {
			if right.at[k].first != i {
				continue // not this key's first row
			}
			if left.has(k) {
				continue // emitted alongside the rows of a
			}
			if !yield(combine(k, nil, right.rows(k))) {
				return
			}
		}
	}
}

// Semi returns a Stream2 of the rows of a whose key appears in b, in the
// encounter order of a. Unlike [Inner] it yields no value of b, and it yields
// each row of a at most once however many rows of b carry the key. An empty b
// yields an empty Stream2.
//
// Semi buffers the keys of b for the duration of the iteration and streams a,
// which may therefore be infinite.
func Semi[K comparable, A, B any](a streams.Stream2[K, A], b streams.Stream2[K, B]) streams.Stream2[K, A] {
	return func(yield func(K, A) bool) {
		match := keys(b)
		for k, v := range a {
			if _, ok := match[k]; ok && !yield(k, v) {
				return
			}
		}
	}
}

// Anti returns a Stream2 of the rows of a whose key does not appear in b, in
// the encounter order of a. It is the complement of [Semi]: an empty b admits
// every row of a.
//
// Anti buffers the keys of b for the duration of the iteration and streams a,
// which may therefore be infinite.
func Anti[K comparable, A, B any](a streams.Stream2[K, A], b streams.Stream2[K, B]) streams.Stream2[K, A] {
	return func(yield func(K, A) bool) {
		match := keys(b)
		for k, v := range a {
			if _, ok := match[k]; !ok && !yield(k, v) {
				return
			}
		}
	}
}

// run locates the rows carrying one key: where the first is, where the last is
// so that another can be appended after it, and how many there are so that
// rows can size its result exactly.
type run struct{ first, last, n int }

// side is one side of a join held in memory: the key and the value of every
// row, in the encounter order of the source, and the rows carrying each key.
//
// The rows of a key form a chain through next rather than a slice per key. A
// map[K][]int is the obvious shape and costs one allocation for every distinct
// key, which for a join over mostly-unique keys is one allocation per row.
type side[K comparable, V any] struct {
	keys   []K
	values []V
	at     map[K]run
	next   []int // the row after this one carrying the same key, or -1
}

// buffer drains s into a side.
func buffer[K comparable, V any](s streams.Stream2[K, V]) side[K, V] {
	buf := side[K, V]{at: make(map[K]run)}
	for k, v := range s {
		i := len(buf.values)
		buf.keys = append(buf.keys, k)
		buf.values = append(buf.values, v)
		buf.next = append(buf.next, -1)
		if r, seen := buf.at[k]; seen {
			buf.next[r.last] = i
			buf.at[k] = run{first: r.first, last: i, n: r.n + 1}
		} else {
			buf.at[k] = run{first: i, last: i, n: 1}
		}
	}
	return buf
}

// has reports whether any row carries k.
func (buf side[K, V]) has(k K) bool {
	_, seen := buf.at[k]
	return seen
}

// rows returns a new slice of the values carrying k, in encounter order, or nil
// if no row carries it.
func (buf side[K, V]) rows(k K) []V {
	r, seen := buf.at[k]
	if !seen {
		return nil
	}
	out := make([]V, 0, r.n)
	for i := r.first; i >= 0; i = buf.next[i] {
		out = append(out, buf.values[i])
	}
	return out
}

// keys drains s into the set of keys it carries.
func keys[K comparable, V any](s streams.Stream2[K, V]) map[K]struct{} {
	set := make(map[K]struct{})
	for k := range s {
		set[k] = struct{}{}
	}
	return set
}
