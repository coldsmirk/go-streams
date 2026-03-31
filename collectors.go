package streams

import (
	"cmp"
	"slices"
	"strings"
)

// Collector defines how to accumulate elements into a result.
// T is the element type, A is the accumulator type, R is the result type.
type Collector[T, A, R any] struct {
	// Supplier creates a new accumulator.
	Supplier func() A
	// Accumulator adds an element to the accumulator.
	Accumulator func(A, T) A
	// Finisher transforms the accumulator to the final result.
	Finisher func(A) R
}

// CollectTo collects stream elements using the given Collector.
func CollectTo[T, A, R any](s Stream[T], c Collector[T, A, R]) R {
	acc := c.Supplier()
	for v := range s.seq {
		acc = c.Accumulator(acc, v)
	}
	return c.Finisher(acc)
}

// --- Pre-built Collectors ---

// ToSliceCollector returns a Collector that accumulates into a slice.
func ToSliceCollector[T any]() Collector[T, []T, []T] {
	return Collector[T, []T, []T]{
		Supplier:    func() []T { return make([]T, 0) },
		Accumulator: func(acc []T, v T) []T { return append(acc, v) },
		Finisher:    func(acc []T) []T { return acc },
	}
}

// ToSetCollector returns a Collector that accumulates into a set.
func ToSetCollector[T comparable]() Collector[T, map[T]struct{}, map[T]struct{}] {
	return Collector[T, map[T]struct{}, map[T]struct{}]{
		Supplier: func() map[T]struct{} { return make(map[T]struct{}) },
		Accumulator: func(acc map[T]struct{}, v T) map[T]struct{} {
			acc[v] = struct{}{}
			return acc
		},
		Finisher: func(acc map[T]struct{}) map[T]struct{} { return acc },
	}
}

// JoiningCollector returns a Collector that joins strings with a separator.
func JoiningCollector(sep string) Collector[string, *strings.Builder, string] {
	return Collector[string, *strings.Builder, string]{
		Supplier: func() *strings.Builder {
			return &strings.Builder{}
		},
		Accumulator: func(sb *strings.Builder, s string) *strings.Builder {
			if sb.Len() > 0 {
				sb.WriteString(sep)
			}
			sb.WriteString(s)
			return sb
		},
		Finisher: func(sb *strings.Builder) string {
			return sb.String()
		},
	}
}

// JoiningCollectorFull returns a Collector that joins strings with separator, prefix, and suffix.
func JoiningCollectorFull(sep, prefix, suffix string) Collector[string, *strings.Builder, string] {
	return Collector[string, *strings.Builder, string]{
		Supplier: func() *strings.Builder {
			sb := &strings.Builder{}
			sb.WriteString(prefix)
			return sb
		},
		Accumulator: func(sb *strings.Builder, s string) *strings.Builder {
			if sb.Len() > len(prefix) {
				sb.WriteString(sep)
			}
			sb.WriteString(s)
			return sb
		},
		Finisher: func(sb *strings.Builder) string {
			sb.WriteString(suffix)
			return sb.String()
		},
	}
}

// countingState holds state for the Counting collector.
type countingState struct{ count int }

// CountingCollector returns a Collector that counts elements.
func CountingCollector[T any]() Collector[T, *countingState, int] {
	return Collector[T, *countingState, int]{
		Supplier: func() *countingState { return &countingState{} },
		Accumulator: func(cs *countingState, _ T) *countingState {
			cs.count++
			return cs
		},
		Finisher: func(cs *countingState) int { return cs.count },
	}
}

// summingState holds state for numeric summing collectors.
type summingState[T Numeric] struct{ sum T }

// SummingCollector returns a Collector that sums numeric elements.
func SummingCollector[T Numeric]() Collector[T, *summingState[T], T] {
	return Collector[T, *summingState[T], T]{
		Supplier: func() *summingState[T] { return &summingState[T]{} },
		Accumulator: func(ss *summingState[T], v T) *summingState[T] {
			ss.sum += v
			return ss
		},
		Finisher: func(ss *summingState[T]) T { return ss.sum },
	}
}

// averagingState holds state for averaging collectors.
type averagingState struct {
	sum   float64
	count int
}

// AveragingCollector returns a Collector that computes the average of numeric elements.
func AveragingCollector[T Numeric]() Collector[T, *averagingState, Optional[float64]] {
	return Collector[T, *averagingState, Optional[float64]]{
		Supplier: func() *averagingState { return &averagingState{} },
		Accumulator: func(as *averagingState, v T) *averagingState {
			as.sum += float64(v)
			as.count++
			return as
		},
		Finisher: func(as *averagingState) Optional[float64] {
			if as.count == 0 {
				return None[float64]()
			}
			return Some(as.sum / float64(as.count))
		},
	}
}

// maxState holds state for max collectors.
type maxState[T any] struct {
	max   T
	found bool
}

// minState holds state for min collectors.
type minState[T any] struct {
	min   T
	found bool
}

// MaxByCollector returns a Collector that finds the maximum element.
func MaxByCollector[T any](cmp func(T, T) int) Collector[T, *maxState[T], Optional[T]] {
	return Collector[T, *maxState[T], Optional[T]]{
		Supplier: func() *maxState[T] { return &maxState[T]{} },
		Accumulator: func(ms *maxState[T], v T) *maxState[T] {
			if !ms.found || cmp(v, ms.max) > 0 {
				ms.max = v
				ms.found = true
			}
			return ms
		},
		Finisher: func(ms *maxState[T]) Optional[T] {
			if ms.found {
				return Some(ms.max)
			}
			return None[T]()
		},
	}
}

// MinByCollector returns a Collector that finds the minimum element.
func MinByCollector[T any](cmp func(T, T) int) Collector[T, *minState[T], Optional[T]] {
	return Collector[T, *minState[T], Optional[T]]{
		Supplier: func() *minState[T] { return &minState[T]{} },
		Accumulator: func(ms *minState[T], v T) *minState[T] {
			if !ms.found || cmp(v, ms.min) < 0 {
				ms.min = v
				ms.found = true
			}
			return ms
		},
		Finisher: func(ms *minState[T]) Optional[T] {
			if ms.found {
				return Some(ms.min)
			}
			return None[T]()
		},
	}
}

// GroupingByCollector returns a Collector that groups elements by a key function.
func GroupingByCollector[T any, K comparable](keyFn func(T) K) Collector[T, map[K][]T, map[K][]T] {
	return Collector[T, map[K][]T, map[K][]T]{
		Supplier: func() map[K][]T { return make(map[K][]T) },
		Accumulator: func(m map[K][]T, v T) map[K][]T {
			k := keyFn(v)
			m[k] = append(m[k], v)
			return m
		},
		Finisher: func(m map[K][]T) map[K][]T { return m },
	}
}

// partitionState holds state for partitioning collectors.
type partitionState[T any] struct {
	trueGroup, falseGroup []T
}

// PartitioningByCollector returns a Collector that partitions elements by a predicate.
func PartitioningByCollector[T any](pred func(T) bool) Collector[T, *partitionState[T], map[bool][]T] {
	return Collector[T, *partitionState[T], map[bool][]T]{
		Supplier: func() *partitionState[T] { return &partitionState[T]{} },
		Accumulator: func(ps *partitionState[T], v T) *partitionState[T] {
			if pred(v) {
				ps.trueGroup = append(ps.trueGroup, v)
			} else {
				ps.falseGroup = append(ps.falseGroup, v)
			}
			return ps
		},
		Finisher: func(ps *partitionState[T]) map[bool][]T {
			return map[bool][]T{
				true:  ps.trueGroup,
				false: ps.falseGroup,
			}
		},
	}
}

// ToMapCollector returns a Collector that creates a map from elements.
func ToMapCollector[T any, K comparable, V any](fn func(T) (K, V)) Collector[T, map[K]V, map[K]V] {
	return Collector[T, map[K]V, map[K]V]{
		Supplier: func() map[K]V { return make(map[K]V) },
		Accumulator: func(m map[K]V, v T) map[K]V {
			k, val := fn(v)
			m[k] = val
			return m
		},
		Finisher: func(m map[K]V) map[K]V { return m },
	}
}

// ToMapCollectorMerging returns a Collector that creates a map with a merge function for duplicate keys.
func ToMapCollectorMerging[T any, K comparable, V any](
	fn func(T) (K, V),
	merge func(V, V) V,
) Collector[T, map[K]V, map[K]V] {
	return Collector[T, map[K]V, map[K]V]{
		Supplier: func() map[K]V { return make(map[K]V) },
		Accumulator: func(m map[K]V, v T) map[K]V {
			k, newVal := fn(v)
			if existing, ok := m[k]; ok {
				m[k] = merge(existing, newVal)
			} else {
				m[k] = newVal
			}
			return m
		},
		Finisher: func(m map[K]V) map[K]V { return m },
	}
}

// firstState holds state for first element collector.
type firstState[T any] struct {
	value T
	found bool
}

// FirstCollector returns a Collector that returns the first element.
func FirstCollector[T any]() Collector[T, *firstState[T], Optional[T]] {
	return Collector[T, *firstState[T], Optional[T]]{
		Supplier: func() *firstState[T] { return &firstState[T]{} },
		Accumulator: func(fs *firstState[T], v T) *firstState[T] {
			if !fs.found {
				fs.value = v
				fs.found = true
			}
			return fs
		},
		Finisher: func(fs *firstState[T]) Optional[T] {
			if fs.found {
				return Some(fs.value)
			}
			return None[T]()
		},
	}
}

// lastState holds state for last element collector.
type lastState[T any] struct {
	value T
	found bool
}

// LastCollector returns a Collector that returns the last element.
func LastCollector[T any]() Collector[T, *lastState[T], Optional[T]] {
	return Collector[T, *lastState[T], Optional[T]]{
		Supplier: func() *lastState[T] { return &lastState[T]{} },
		Accumulator: func(ls *lastState[T], v T) *lastState[T] {
			ls.value = v
			ls.found = true
			return ls
		},
		Finisher: func(ls *lastState[T]) Optional[T] {
			if ls.found {
				return Some(ls.value)
			}
			return None[T]()
		},
	}
}

// ReducingCollector returns a Collector that reduces elements using an identity and function.
func ReducingCollector[T any](identity T, fn func(T, T) T) Collector[T, *T, T] {
	return Collector[T, *T, T]{
		Supplier: func() *T {
			v := identity
			return &v
		},
		Accumulator: func(acc *T, v T) *T {
			*acc = fn(*acc, v)
			return acc
		},
		Finisher: func(acc *T) T { return *acc },
	}
}

// MappingCollector applies a transformation before collecting.
func MappingCollector[T, U, A, R any](mapper func(T) U, downstream Collector[U, A, R]) Collector[T, A, R] {
	return Collector[T, A, R]{
		Supplier: downstream.Supplier,
		Accumulator: func(acc A, v T) A {
			return downstream.Accumulator(acc, mapper(v))
		},
		Finisher: downstream.Finisher,
	}
}

// FilteringCollector filters elements before collecting.
func FilteringCollector[T, A, R any](pred func(T) bool, downstream Collector[T, A, R]) Collector[T, A, R] {
	return Collector[T, A, R]{
		Supplier: downstream.Supplier,
		Accumulator: func(acc A, v T) A {
			if pred(v) {
				return downstream.Accumulator(acc, v)
			}
			return acc
		},
		Finisher: downstream.Finisher,
	}
}

// FlatMappingCollector flat-maps elements before collecting.
func FlatMappingCollector[T, U, A, R any](mapper func(T) Stream[U], downstream Collector[U, A, R]) Collector[T, A, R] {
	return Collector[T, A, R]{
		Supplier: downstream.Supplier,
		Accumulator: func(acc A, v T) A {
			for u := range mapper(v).seq {
				acc = downstream.Accumulator(acc, u)
			}
			return acc
		},
		Finisher: downstream.Finisher,
	}
}

// teeingState holds state for teeing collectors.
type teeingState[A1, A2 any] struct {
	acc1 A1
	acc2 A2
}

// TeeingCollector combines the results of two collectors.
func TeeingCollector[T, A1, R1, A2, R2, R any](
	c1 Collector[T, A1, R1],
	c2 Collector[T, A2, R2],
	merger func(R1, R2) R,
) Collector[T, *teeingState[A1, A2], R] {
	return Collector[T, *teeingState[A1, A2], R]{
		Supplier: func() *teeingState[A1, A2] {
			return &teeingState[A1, A2]{
				acc1: c1.Supplier(),
				acc2: c2.Supplier(),
			}
		},
		Accumulator: func(ts *teeingState[A1, A2], v T) *teeingState[A1, A2] {
			ts.acc1 = c1.Accumulator(ts.acc1, v)
			ts.acc2 = c2.Accumulator(ts.acc2, v)
			return ts
		},
		Finisher: func(ts *teeingState[A1, A2]) R {
			return merger(c1.Finisher(ts.acc1), c2.Finisher(ts.acc2))
		},
	}
}

// --- TopK and BottomK Collectors ---

// topKState holds state for TopK collector using a min-heap.
type topKState[T any] struct {
	heap []T
	k    int
	less func(T, T) bool
}

// bottomKState holds state for BottomK collector using a max-heap.
type bottomKState[T any] struct {
	heap []T
	k    int
	less func(T, T) bool
}

// heapify maintains the min-heap property.
func (s *topKState[T]) heapifyDown(i int) {
	n := len(s.heap)
	for {
		smallest := i
		left := 2*i + 1
		right := 2*i + 2

		if left < n && s.less(s.heap[left], s.heap[smallest]) {
			smallest = left
		}
		if right < n && s.less(s.heap[right], s.heap[smallest]) {
			smallest = right
		}
		if smallest == i {
			break
		}
		s.heap[i], s.heap[smallest] = s.heap[smallest], s.heap[i]
		i = smallest
	}
}

func (s *topKState[T]) heapifyUp(i int) {
	for i > 0 {
		parent := (i - 1) / 2
		if !s.less(s.heap[i], s.heap[parent]) {
			break
		}
		s.heap[i], s.heap[parent] = s.heap[parent], s.heap[i]
		i = parent
	}
}

// TopKCollector returns a Collector that finds the k largest elements.
// Uses a min-heap to maintain O(n log k) complexity.
// The less function should return true if a < b.
func TopKCollector[T any](k int, less func(T, T) bool) Collector[T, *topKState[T], []T] {
	return Collector[T, *topKState[T], []T]{
		Supplier: func() *topKState[T] {
			return &topKState[T]{
				heap: make([]T, 0, k),
				k:    k,
				less: less,
			}
		},
		Accumulator: func(s *topKState[T], v T) *topKState[T] {
			if len(s.heap) < s.k {
				// Heap not full, just add
				s.heap = append(s.heap, v)
				s.heapifyUp(len(s.heap) - 1)
			} else if !s.less(v, s.heap[0]) {
				// v is larger than min in heap, replace
				s.heap[0] = v
				s.heapifyDown(0)
			}
			return s
		},
		Finisher: func(s *topKState[T]) []T {
			// Sort result in descending order (largest first) using O(n log n) algorithm
			result := make([]T, len(s.heap))
			copy(result, s.heap)
			slices.SortFunc(result, func(a, b T) int {
				if s.less(a, b) {
					return 1 // a < b, so in descending order, b comes first
				}
				if s.less(b, a) {
					return -1 // b < a, so in descending order, a comes first
				}
				return 0
			})
			return result
		},
	}
}

func (s *bottomKState[T]) heapifyDown(i int) {
	n := len(s.heap)
	for {
		largest := i
		left := 2*i + 1
		right := 2*i + 2

		// Max-heap: parent should be larger than children
		if left < n && s.less(s.heap[largest], s.heap[left]) {
			largest = left
		}
		if right < n && s.less(s.heap[largest], s.heap[right]) {
			largest = right
		}
		if largest == i {
			break
		}
		s.heap[i], s.heap[largest] = s.heap[largest], s.heap[i]
		i = largest
	}
}

func (s *bottomKState[T]) heapifyUp(i int) {
	for i > 0 {
		parent := (i - 1) / 2
		if !s.less(s.heap[parent], s.heap[i]) {
			break
		}
		s.heap[i], s.heap[parent] = s.heap[parent], s.heap[i]
		i = parent
	}
}

// BottomKCollector returns a Collector that finds the k smallest elements.
// Uses a max-heap to maintain O(n log k) complexity.
func BottomKCollector[T any](k int, less func(T, T) bool) Collector[T, *bottomKState[T], []T] {
	return Collector[T, *bottomKState[T], []T]{
		Supplier: func() *bottomKState[T] {
			return &bottomKState[T]{
				heap: make([]T, 0, k),
				k:    k,
				less: less,
			}
		},
		Accumulator: func(s *bottomKState[T], v T) *bottomKState[T] {
			if len(s.heap) < s.k {
				s.heap = append(s.heap, v)
				s.heapifyUp(len(s.heap) - 1)
			} else if s.less(v, s.heap[0]) {
				// v is smaller than max in heap, replace
				s.heap[0] = v
				s.heapifyDown(0)
			}
			return s
		},
		Finisher: func(s *bottomKState[T]) []T {
			result := make([]T, len(s.heap))
			copy(result, s.heap)
			slices.SortFunc(result, func(a, b T) int {
				if s.less(a, b) {
					return -1 // a < b, so a comes first
				}
				if s.less(b, a) {
					return 1 // b < a, so b comes first
				}
				return 0
			})
			return result
		},
	}
}

// --- Convenience Functions for TopK ---

// TopK returns the k largest elements from a stream.
func TopK[T any](s Stream[T], k int, less func(T, T) bool) []T {
	return CollectTo(s, TopKCollector(k, less))
}

// BottomK returns the k smallest elements from a stream.
func BottomK[T any](s Stream[T], k int, less func(T, T) bool) []T {
	return CollectTo(s, BottomKCollector(k, less))
}

// --- Quantile Collectors ---

// quantileState holds all elements for quantile calculation.
type quantileState[T any] struct {
	elements []T
	less     func(T, T) bool
}

// QuantileCollector returns a Collector that computes a quantile.
// The quantile q should be between 0 and 1 (e.g., 0.5 for median).
// Note: This collector stores all elements in memory.
func QuantileCollector[T any](q float64, less func(T, T) bool) Collector[T, *quantileState[T], Optional[T]] {
	return Collector[T, *quantileState[T], Optional[T]]{
		Supplier: func() *quantileState[T] {
			return &quantileState[T]{
				elements: make([]T, 0),
				less:     less,
			}
		},
		Accumulator: func(s *quantileState[T], v T) *quantileState[T] {
			s.elements = append(s.elements, v)
			return s
		},
		Finisher: func(s *quantileState[T]) Optional[T] {
			n := len(s.elements)
			if n == 0 {
				return None[T]()
			}

			// Sort elements using O(n log n) algorithm
			sorted := make([]T, n)
			copy(sorted, s.elements)
			slices.SortFunc(sorted, func(a, b T) int {
				if s.less(a, b) {
					return -1
				}
				if s.less(b, a) {
					return 1
				}
				return 0
			})

			// Calculate index
			idx := int(q * float64(n-1))
			idx = max(idx, 0)
			idx = min(idx, n-1)

			return Some(sorted[idx])
		},
	}
}

// Quantile returns the q-th quantile from a stream.
func Quantile[T any](s Stream[T], q float64, less func(T, T) bool) Optional[T] {
	return CollectTo(s, QuantileCollector(q, less))
}

// Median returns the median (0.5 quantile) from a stream.
func Median[T any](s Stream[T], less func(T, T) bool) Optional[T] {
	return Quantile(s, 0.5, less)
}

// Percentile returns the p-th percentile (p in 0-100) from a stream.
func Percentile[T any](s Stream[T], p float64, less func(T, T) bool) Optional[T] {
	return Quantile(s, p/100.0, less)
}

// --- Frequency Collectors ---

// FrequencyCollector returns a Collector that counts occurrences of each element.
func FrequencyCollector[T comparable]() Collector[T, map[T]int, map[T]int] {
	return Collector[T, map[T]int, map[T]int]{
		Supplier: func() map[T]int { return make(map[T]int) },
		Accumulator: func(m map[T]int, v T) map[T]int {
			m[v]++
			return m
		},
		Finisher: func(m map[T]int) map[T]int { return m },
	}
}

// Frequency returns a map of element frequencies.
func Frequency[T comparable](s Stream[T]) map[T]int {
	return CollectTo(s, FrequencyCollector[T]())
}

// MostCommon returns the n most common elements with their counts.
func MostCommon[T comparable](s Stream[T], n int) []Pair[T, int] {
	freq := Frequency(s)

	// Convert to pairs
	pairs := make([]Pair[T, int], 0, len(freq))
	for k, v := range freq {
		pairs = append(pairs, NewPair(k, v))
	}

	// Sort by count descending using O(n log n) algorithm
	slices.SortFunc(pairs, func(a, b Pair[T, int]) int {
		return cmp.Compare(b.Second, a.Second) // Descending order
	})

	if n > len(pairs) {
		n = len(pairs)
	}
	return pairs[:n]
}

// --- Histogram Collector ---

// histogramState holds state for histogram building.
type histogramState[T any, K comparable] struct {
	buckets map[K][]T
	keyFn   func(T) K
}

// HistogramCollector groups elements into buckets based on a key function.
func HistogramCollector[T any, K comparable](keyFn func(T) K) Collector[T, *histogramState[T, K], map[K][]T] {
	return Collector[T, *histogramState[T, K], map[K][]T]{
		Supplier: func() *histogramState[T, K] {
			return &histogramState[T, K]{
				buckets: make(map[K][]T),
				keyFn:   keyFn,
			}
		},
		Accumulator: func(s *histogramState[T, K], v T) *histogramState[T, K] {
			k := s.keyFn(v)
			s.buckets[k] = append(s.buckets[k], v)
			return s
		},
		Finisher: func(s *histogramState[T, K]) map[K][]T {
			return s.buckets
		},
	}
}
