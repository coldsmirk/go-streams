// Package collections bridges [streams] and the container types of
// github.com/coldsmirk/go-collections.
//
// The From functions present a collection as a Stream and the To functions
// drain a Stream into a collection. Everything in between is an ordinary
// Stream operation, so the bridge holds conversions and nothing else:
//
//	upper := collections.ToTreeSet(
//		collections.FromList(names).Map(strings.ToUpper),
//		strings.Compare,
//	)
//
// A From function iterates the collection it is given rather than copying it,
// so the Stream reads the collection as it is consumed. Modifying a collection
// while a Stream over it is being consumed is unsafe, except for the concurrent
// implementations, whose iteration is documented to yield a best-effort
// snapshot.
package collections

import (
	"iter"

	coll "github.com/coldsmirk/go-collections"
	streams "github.com/coldsmirk/go-streams/v2"
)

// --- collections into streams ---

// Each From function defers the call to the collection's Seq method into the
// returned iterator rather than calling it here. That is load-bearing, not
// ceremony: several go-collections implementations capture their backing slice
// when Seq is called, so calling it eagerly would snapshot the collection at
// construction and break the contract above. Do not rewrite these as
// streams.From(x.Seq()).

// FromSet returns a Stream over the elements of s in unspecified order.
// Modifying s while the Stream is being consumed is unsafe.
func FromSet[T any](s coll.Set[T]) streams.Stream[T] {
	return func(yield func(T) bool) { s.Seq()(yield) }
}

// FromSortedSet returns a Stream over the elements of s in ascending order as
// defined by the comparator of s. Modifying s while the Stream is being
// consumed is unsafe.
func FromSortedSet[T any](s coll.SortedSet[T]) streams.Stream[T] {
	return func(yield func(T) bool) { s.Seq()(yield) }
}

// FromList returns a Stream over the elements of l in index order, from the
// first element to the last. Modifying l while the Stream is being consumed is
// unsafe.
func FromList[T any](l coll.List[T]) streams.Stream[T] {
	return func(yield func(T) bool) { l.Seq()(yield) }
}

// FromQueue returns a Stream over the elements of q from front to back, the
// order in which they would be dequeued. The elements stay in q. Modifying q
// while the Stream is being consumed is unsafe.
func FromQueue[T any](q coll.Queue[T]) streams.Stream[T] {
	return func(yield func(T) bool) { q.Seq()(yield) }
}

// FromStack returns a Stream over the elements of s from top to bottom, the
// order in which they would be popped. The elements stay in s. Modifying s
// while the Stream is being consumed is unsafe.
func FromStack[T any](s coll.Stack[T]) streams.Stream[T] {
	return func(yield func(T) bool) { s.Seq()(yield) }
}

// FromDeque returns a Stream over the elements of d from front to back. The
// elements stay in d. Modifying d while the Stream is being consumed is unsafe.
func FromDeque[T any](d coll.Deque[T]) streams.Stream[T] {
	return func(yield func(T) bool) { d.Seq()(yield) }
}

// FromPriorityQueue returns a Stream over the elements of q in the internal
// heap order, which is not priority order: only the first element is the
// highest-priority one. Sort the Stream with the comparator of q to obtain
// priority order. The elements stay in q. Modifying q while the Stream is being
// consumed is unsafe.
func FromPriorityQueue[T any](q coll.PriorityQueue[T]) streams.Stream[T] {
	return func(yield func(T) bool) { q.Seq()(yield) }
}

// FromMap returns a Stream2 over the entries of m in unspecified order.
// Modifying m while the Stream2 is being consumed is unsafe.
func FromMap[K, V any](m coll.Map[K, V]) streams.Stream2[K, V] {
	return func(yield func(K, V) bool) { m.Seq()(yield) }
}

// FromSortedMap returns a Stream2 over the entries of m in ascending key order
// as defined by the key comparator of m. Modifying m while the Stream2 is being
// consumed is unsafe.
func FromSortedMap[K, V any](m coll.SortedMap[K, V]) streams.Stream2[K, V] {
	return func(yield func(K, V) bool) { m.Seq()(yield) }
}

// --- streams into collections ---

// ToHashSet returns a hash set of the elements of s, keeping one of each
// duplicate. The set iterates in unspecified order.
func ToHashSet[T comparable](s streams.Stream[T]) coll.Set[T] {
	set := coll.NewHashSet[T]()
	set.AddSeq(iter.Seq[T](s))
	return set
}

// ToTreeSet returns a tree set of the elements of s ordered by cmp, keeping the
// first of each group of elements that cmp reports equal.
//
// It costs two tree traversals per element rather than one, because the
// underlying set's Add overwrites an element its comparator reports equal
// before reporting that it did. Use [ToHashSet] where ordering is not needed.
func ToTreeSet[T any](s streams.Stream[T], cmp coll.Comparator[T]) coll.SortedSet[T] {
	set := coll.NewTreeSet(cmp)
	// Not AddSeq: the underlying tree overwrites an item its comparator reports
	// equal, which would keep the last of each group rather than the first.
	for v := range s {
		if !set.Contains(v) {
			set.Add(v)
		}
	}
	return set
}

// ToArrayList returns a slice-backed list of the elements of s, in encounter
// order.
func ToArrayList[T any](s streams.Stream[T]) coll.List[T] {
	list := coll.NewArrayList[T]()
	list.AddSeq(iter.Seq[T](s))
	return list
}

// ToLinkedList returns a doubly linked list of the elements of s, in encounter
// order.
func ToLinkedList[T any](s streams.Stream[T]) coll.List[T] {
	list := coll.NewLinkedList[T]()
	list.AddSeq(iter.Seq[T](s))
	return list
}

// ToHashMap returns a hash map of the pairs of s. A later pair overwrites an
// earlier one with the same key. The map iterates in unspecified order.
func ToHashMap[K comparable, V any](s streams.Stream2[K, V]) coll.Map[K, V] {
	m := coll.NewHashMap[K, V]()
	m.PutSeq(iter.Seq2[K, V](s))
	return m
}

// ToTreeMap returns a tree map of the pairs of s with keys ordered by cmp. A
// later pair overwrites an earlier one whose key cmp reports equal.
func ToTreeMap[K, V any](s streams.Stream2[K, V], cmp coll.Comparator[K]) coll.SortedMap[K, V] {
	m := coll.NewTreeMap[K, V](cmp)
	m.PutSeq(iter.Seq2[K, V](s))
	return m
}
