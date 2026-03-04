package streams

import (
	"iter"

	collections "github.com/coldsmirk/go-collections"
)

// =============================================================================
// Stream Constructors from go-collections
// =============================================================================

// FromSet creates a Stream from a collections.Set.
func FromSet[T any](set collections.Set[T]) Stream[T] {
	return From(set.Seq())
}

// FromSortedSet creates a Stream from a collections.SortedSet in ascending order.
func FromSortedSet[T any](set collections.SortedSet[T]) Stream[T] {
	return From(set.Seq())
}

// FromSortedSetDescending creates a Stream from a collections.SortedSet in descending order.
func FromSortedSetDescending[T any](set collections.SortedSet[T]) Stream[T] {
	return From(set.Reversed())
}

// FromList creates a Stream from a collections.List.
func FromList[T any](list collections.List[T]) Stream[T] {
	return From(list.Seq())
}

// FromMapC creates a Stream2 from a collections.Map.
// The "C" suffix distinguishes it from FromMap which takes a Go map.
func FromMapC[K, V any](m collections.Map[K, V]) Stream2[K, V] {
	return From2(m.Seq())
}

// FromSortedMapC creates a Stream2 from a collections.SortedMap in ascending key order.
func FromSortedMapC[K, V any](m collections.SortedMap[K, V]) Stream2[K, V] {
	return From2(m.Seq())
}

// FromSortedMapCDescending creates a Stream2 from a collections.SortedMap in descending key order.
func FromSortedMapCDescending[K, V any](m collections.SortedMap[K, V]) Stream2[K, V] {
	return From2(m.Reversed())
}

// FromQueue creates a Stream from a collections.Queue (FIFO order).
func FromQueue[T any](q collections.Queue[T]) Stream[T] {
	return From(q.Seq())
}

// FromStack creates a Stream from a collections.Stack (LIFO order).
func FromStack[T any](s collections.Stack[T]) Stream[T] {
	return From(s.Seq())
}

// FromDeque creates a Stream from a collections.Deque (front to back).
func FromDeque[T any](d collections.Deque[T]) Stream[T] {
	return From(d.Seq())
}

// FromDequeReversed creates a Stream from a collections.Deque (back to front).
func FromDequeReversed[T any](d collections.Deque[T]) Stream[T] {
	return From(d.Reversed())
}

// FromPriorityQueue creates a Stream from a collections.PriorityQueue.
// Elements are yielded in heap order (not priority-sorted order).
// Use FromPriorityQueueSorted for priority-sorted iteration.
func FromPriorityQueue[T any](pq collections.PriorityQueue[T]) Stream[T] {
	return From(pq.Seq())
}

// FromPriorityQueueSorted creates a Stream from a collections.PriorityQueue.
// Elements are yielded in priority order (sorted).
// Note: This collects all elements first.
func FromPriorityQueueSorted[T any](pq collections.PriorityQueue[T]) Stream[T] {
	return FromSlice(pq.ToSortedSlice())
}

// =============================================================================
// Terminal Operations returning go-collections types
// =============================================================================

// ToHashSet collects stream elements into a collections.Set[T].
func ToHashSet[T comparable](s Stream[T]) collections.Set[T] {
	set := collections.NewHashSet[T]()
	set.AddSeq(s.Seq())
	return set
}

// ToTreeSet collects stream elements into a collections.SortedSet[T].
// Elements are maintained in sorted order according to the comparator.
func ToTreeSet[T any](s Stream[T], cmp collections.Comparator[T]) collections.SortedSet[T] {
	set := collections.NewTreeSet(cmp)
	set.AddSeq(s.Seq())
	return set
}

// ToArrayList collects stream elements into a collections.List[T].
func ToArrayList[T any](s Stream[T]) collections.List[T] {
	list := collections.NewArrayList[T]()
	list.AddSeq(s.Seq())
	return list
}

// ToLinkedList collects stream elements into a collections.List[T] (linked list implementation).
func ToLinkedList[T any](s Stream[T]) collections.List[T] {
	list := collections.NewLinkedList[T]()
	list.AddSeq(s.Seq())
	return list
}

// ToHashMapC collects stream elements into a collections.Map[K, V].
// The "C" suffix distinguishes it from ToMap which returns a Go map.
func ToHashMapC[T any, K comparable, V any](s Stream[T], keyFn func(T) K, valFn func(T) V) collections.Map[K, V] {
	m := collections.NewHashMap[K, V]()
	for v := range s.Seq() {
		m.Put(keyFn(v), valFn(v))
	}
	return m
}

// ToTreeMapC collects stream elements into a collections.SortedMap[K, V].
// Keys are maintained in sorted order according to the comparator.
func ToTreeMapC[T any, K any, V any](s Stream[T], keyFn func(T) K, valFn func(T) V, keyCmp collections.Comparator[K]) collections.SortedMap[K, V] {
	m := collections.NewTreeMap[K, V](keyCmp)
	for v := range s.Seq() {
		m.Put(keyFn(v), valFn(v))
	}
	return m
}

// ToHashMap2C converts a Stream2[K, V] into a collections.Map[K, V].
func ToHashMap2C[K comparable, V any](s Stream2[K, V]) collections.Map[K, V] {
	m := collections.NewHashMap[K, V]()
	m.PutSeq(s.Seq2())
	return m
}

// ToTreeMap2C converts a Stream2[K, V] into a collections.SortedMap[K, V].
func ToTreeMap2C[K any, V any](s Stream2[K, V], keyCmp collections.Comparator[K]) collections.SortedMap[K, V] {
	m := collections.NewTreeMap[K, V](keyCmp)
	for k, v := range s.Seq2() {
		m.Put(k, v)
	}
	return m
}

// =============================================================================
// Collectors returning go-collections types
// =============================================================================

// ToHashSetCollector returns a Collector that accumulates elements into a collections.Set.
func ToHashSetCollector[T comparable]() Collector[T, collections.Set[T], collections.Set[T]] {
	return Collector[T, collections.Set[T], collections.Set[T]]{
		Supplier:    func() collections.Set[T] { return collections.NewHashSet[T]() },
		Accumulator: func(acc collections.Set[T], v T) collections.Set[T] { acc.Add(v); return acc },
		Finisher:    func(acc collections.Set[T]) collections.Set[T] { return acc },
	}
}

// ToTreeSetCollector returns a Collector that accumulates elements into a collections.SortedSet.
func ToTreeSetCollector[T any](cmp collections.Comparator[T]) Collector[T, collections.SortedSet[T], collections.SortedSet[T]] {
	return Collector[T, collections.SortedSet[T], collections.SortedSet[T]]{
		Supplier:    func() collections.SortedSet[T] { return collections.NewTreeSet(cmp) },
		Accumulator: func(acc collections.SortedSet[T], v T) collections.SortedSet[T] { acc.Add(v); return acc },
		Finisher:    func(acc collections.SortedSet[T]) collections.SortedSet[T] { return acc },
	}
}

// ToArrayListCollector returns a Collector that accumulates elements into a collections.List.
func ToArrayListCollector[T any]() Collector[T, collections.List[T], collections.List[T]] {
	return Collector[T, collections.List[T], collections.List[T]]{
		Supplier:    func() collections.List[T] { return collections.NewArrayList[T]() },
		Accumulator: func(acc collections.List[T], v T) collections.List[T] { acc.Add(v); return acc },
		Finisher:    func(acc collections.List[T]) collections.List[T] { return acc },
	}
}

// ToHashMapCollector returns a Collector that accumulates elements into a collections.Map.
func ToHashMapCollector[T any, K comparable, V any](keyFn func(T) K, valFn func(T) V) Collector[T, collections.Map[K, V], collections.Map[K, V]] {
	return Collector[T, collections.Map[K, V], collections.Map[K, V]]{
		Supplier: func() collections.Map[K, V] { return collections.NewHashMap[K, V]() },
		Accumulator: func(acc collections.Map[K, V], v T) collections.Map[K, V] {
			acc.Put(keyFn(v), valFn(v))
			return acc
		},
		Finisher: func(acc collections.Map[K, V]) collections.Map[K, V] { return acc },
	}
}

// ToTreeMapCollector returns a Collector that accumulates elements into a collections.SortedMap.
func ToTreeMapCollector[T any, K any, V any](keyFn func(T) K, valFn func(T) V, keyCmp collections.Comparator[K]) Collector[T, collections.SortedMap[K, V], collections.SortedMap[K, V]] {
	return Collector[T, collections.SortedMap[K, V], collections.SortedMap[K, V]]{
		Supplier: func() collections.SortedMap[K, V] { return collections.NewTreeMap[K, V](keyCmp) },
		Accumulator: func(acc collections.SortedMap[K, V], v T) collections.SortedMap[K, V] {
			acc.Put(keyFn(v), valFn(v))
			return acc
		},
		Finisher: func(acc collections.SortedMap[K, V]) collections.SortedMap[K, V] { return acc },
	}
}

// =============================================================================
// GroupBy variants returning go-collections types
// =============================================================================

// GroupByToHashMap groups elements by key into a collections.Map[K, []T].
func GroupByToHashMap[T any, K comparable](s Stream[T], keyFn func(T) K) collections.Map[K, []T] {
	m := collections.NewHashMap[K, []T]()
	for v := range s.Seq() {
		k := keyFn(v)
		if existing, ok := m.Get(k); ok {
			m.Put(k, append(existing, v))
		} else {
			m.Put(k, []T{v})
		}
	}
	return m
}

// GroupByToTreeMap groups elements by key into a collections.SortedMap[K, []T].
func GroupByToTreeMap[T any, K any](s Stream[T], keyFn func(T) K, keyCmp collections.Comparator[K]) collections.SortedMap[K, []T] {
	m := collections.NewTreeMap[K, []T](keyCmp)
	for v := range s.Seq() {
		k := keyFn(v)
		if existing, ok := m.Get(k); ok {
			m.Put(k, append(existing, v))
		} else {
			m.Put(k, []T{v})
		}
	}
	return m
}

// GroupValuesToHashMap groups Stream2 values by key into a collections.Map[K, []V].
func GroupValuesToHashMap[K comparable, V any](s Stream2[K, V]) collections.Map[K, []V] {
	m := collections.NewHashMap[K, []V]()
	for k, v := range s.Seq2() {
		if existing, ok := m.Get(k); ok {
			m.Put(k, append(existing, v))
		} else {
			m.Put(k, []V{v})
		}
	}
	return m
}

// =============================================================================
// Frequency/Histogram variants returning go-collections types
// =============================================================================

// FrequencyToHashMap returns element frequencies as a collections.Map[T, int].
func FrequencyToHashMap[T comparable](s Stream[T]) collections.Map[T, int] {
	m := collections.NewHashMap[T, int]()
	for v := range s.Seq() {
		if cnt, ok := m.Get(v); ok {
			m.Put(v, cnt+1)
		} else {
			m.Put(v, 1)
		}
	}
	return m
}

// HistogramToHashMap buckets elements by key into a collections.Map[K, []T].
func HistogramToHashMap[T any, K comparable](s Stream[T], keyFn func(T) K) collections.Map[K, []T] {
	return GroupByToHashMap(s, keyFn)
}

// =============================================================================
// Stream creation from iter.Seq with go-collections helpers
// =============================================================================

// CollectToSet is a convenience function that collects an iter.Seq into a collections.Set.
func CollectToSet[T comparable](seq iter.Seq[T]) collections.Set[T] {
	set := collections.NewHashSet[T]()
	set.AddSeq(seq)
	return set
}

// CollectToList is a convenience function that collects an iter.Seq into a collections.List.
func CollectToList[T any](seq iter.Seq[T]) collections.List[T] {
	list := collections.NewArrayList[T]()
	list.AddSeq(seq)
	return list
}

// CollectToMap is a convenience function that collects an iter.Seq2 into a collections.Map.
func CollectToMap[K comparable, V any](seq iter.Seq2[K, V]) collections.Map[K, V] {
	m := collections.NewHashMap[K, V]()
	m.PutSeq(seq)
	return m
}
