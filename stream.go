package streams

import (
	"cmp"
	"iter"
	"slices"
)

// Stream is a lazy sequence of elements.
// It wraps iter.Seq[T] and provides fluent functional programming operations.
type Stream[T any] struct {
	seq iter.Seq[T]
}

// From creates a Stream from an iter.Seq.
// This provides interoperability with the standard library.
func From[T any](seq iter.Seq[T]) Stream[T] {
	return Stream[T]{seq: seq}
}

// Of creates a Stream from variadic values.
func Of[T any](values ...T) Stream[T] {
	return Stream[T]{seq: slices.Values(values)}
}

// FromSlice creates a Stream from a slice.
func FromSlice[T any](s []T) Stream[T] {
	return Stream[T]{seq: slices.Values(s)}
}

// FromMap creates a Stream2 from a map.
func FromMap[K comparable, V any](m map[K]V) Stream2[K, V] {
	return Stream2[K, V]{
		seq: func(yield func(K, V) bool) {
			for k, v := range m {
				if !yield(k, v) {
					return
				}
			}
		},
	}
}

// FromChannel creates a Stream from a receive-only channel.
// The stream will consume all values from the channel until it's closed.
func FromChannel[T any](ch <-chan T) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			for v := range ch {
				if !yield(v) {
					return
				}
			}
		},
	}
}

// Generate creates an infinite Stream using a supplier function.
// Each call to the supplier generates the next element.
// Be sure to use Limit() or TakeWhile() to bound the stream.
func Generate[T any](supplier func() T) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			for {
				if !yield(supplier()) {
					return
				}
			}
		},
	}
}

// Iterate creates an infinite Stream: seed, f(seed), f(f(seed)), ...
// Be sure to use Limit() or TakeWhile() to bound the stream.
func Iterate[T any](seed T, fn func(T) T) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			current := seed
			for {
				if !yield(current) {
					return
				}
				current = fn(current)
			}
		},
	}
}

// Range creates a Stream of integers [start, end).
// Returns an empty stream if start >= end.
func Range(start, end int) Stream[int] {
	return Stream[int]{
		seq: func(yield func(int) bool) {
			for i := start; i < end; i++ {
				if !yield(i) {
					return
				}
			}
		},
	}
}

// RangeClosed creates a Stream of integers [start, end].
// Returns an empty stream if start > end.
func RangeClosed(start, end int) Stream[int] {
	return Stream[int]{
		seq: func(yield func(int) bool) {
			for i := start; i <= end; i++ {
				if !yield(i) {
					return
				}
			}
		},
	}
}

// Concat concatenates multiple Streams into one.
// Elements are produced in order from each stream.
func Concat[T any](streams ...Stream[T]) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			for _, s := range streams {
				for v := range s.seq {
					if !yield(v) {
						return
					}
				}
			}
		},
	}
}

// Empty returns an empty Stream.
func Empty[T any]() Stream[T] {
	return Stream[T]{
		seq: func(_ func(T) bool) {},
	}
}

// Repeat creates a Stream that repeats the given value n times.
// If n <= 0, returns an empty stream.
func Repeat[T any](value T, n int) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			for range n {
				if !yield(value) {
					return
				}
			}
		},
	}
}

// RepeatForever creates an infinite Stream that repeatedly yields the given value.
// Be sure to use Limit() or TakeWhile() to bound the stream.
func RepeatForever[T any](value T) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			for {
				if !yield(value) {
					return
				}
			}
		},
	}
}

// Cycle creates an infinite Stream that cycles through the given values.
// Returns an empty stream if no values are provided.
// Be sure to use Limit() or TakeWhile() to bound the stream.
func Cycle[T any](values ...T) Stream[T] {
	if len(values) == 0 {
		return Empty[T]()
	}
	return Stream[T]{
		seq: func(yield func(T) bool) {
			for {
				for _, v := range values {
					if !yield(v) {
						return
					}
				}
			}
		},
	}
}

// Seq returns the underlying iter.Seq for stdlib interop.
// This is the escape hatch to use the stream with for-range loops
// and other iter.Seq-based APIs.
func (s Stream[T]) Seq() iter.Seq[T] {
	return s.seq
}

// --- Intermediate Operations (return Stream, lazy) ---

// Filter returns a Stream containing only elements that match the predicate.
func (s Stream[T]) Filter(pred func(T) bool) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			for v := range s.seq {
				if pred(v) && !yield(v) {
					return
				}
			}
		},
	}
}

// Map transforms each element using the given function.
// For type-changing transformations, use the MapTo function instead.
func (s Stream[T]) Map(fn func(T) T) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			for v := range s.seq {
				if !yield(fn(v)) {
					return
				}
			}
		},
	}
}

// Peek performs the given action on each element as it passes through.
// Useful for debugging or side effects.
func (s Stream[T]) Peek(action func(T)) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			for v := range s.seq {
				action(v)
				if !yield(v) {
					return
				}
			}
		},
	}
}

// Limit returns a Stream containing at most n elements.
func (s Stream[T]) Limit(n int) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			count := 0
			for v := range s.seq {
				if count >= n {
					return
				}
				if !yield(v) {
					return
				}
				count++
			}
		},
	}
}

// Skip returns a Stream that skips the first n elements.
func (s Stream[T]) Skip(n int) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			count := 0
			for v := range s.seq {
				if count < n {
					count++
					continue
				}
				if !yield(v) {
					return
				}
			}
		},
	}
}

// TakeWhile returns a Stream that yields elements while the predicate is true.
// Stops producing elements as soon as the predicate returns false.
func (s Stream[T]) TakeWhile(pred func(T) bool) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			for v := range s.seq {
				if !pred(v) {
					return
				}
				if !yield(v) {
					return
				}
			}
		},
	}
}

// DropWhile returns a Stream that skips elements while the predicate is true.
// Once the predicate returns false, all remaining elements are yielded.
func (s Stream[T]) DropWhile(pred func(T) bool) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			dropping := true
			for v := range s.seq {
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
		},
	}
}

// Sorted returns a Stream with elements sorted using the given comparison function.
// The comparison function should return:
//   - negative if a < b
//   - zero if a == b
//   - positive if a > b
//
// Note: This is an eager operation that collects all elements into memory.
func (s Stream[T]) Sorted(cmp func(a, b T) int) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			collected := slices.Collect(s.seq)
			slices.SortFunc(collected, cmp)
			for _, v := range collected {
				if !yield(v) {
					return
				}
			}
		},
	}
}

// Reverse returns a Stream with elements in reverse order.
// Note: This is an eager operation that collects all elements into memory.
func (s Stream[T]) Reverse() Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			collected := slices.Collect(s.seq)
			for _, v := range slices.Backward(collected) {
				if !yield(v) {
					return
				}
			}
		},
	}
}

// Chunk returns a Stream of slices, each containing up to size elements.
// The last chunk may contain fewer elements.
// Note: This is a free function due to Go generics limitation with method return types.
func Chunk[T any](s Stream[T], size int) Stream[[]T] {
	if size <= 0 {
		return Empty[[]T]()
	}
	return Stream[[]T]{
		seq: func(yield func([]T) bool) {
			chunk := make([]T, 0, size)
			for v := range s.seq {
				chunk = append(chunk, v)
				if len(chunk) == size {
					if !yield(chunk) {
						return
					}
					chunk = make([]T, 0, size)
				}
			}
			if len(chunk) > 0 {
				yield(chunk)
			}
		},
	}
}

// --- Free Functions for type transformation ---

// MapTo transforms Stream[T] to Stream[U].
// Use this when the transformation changes the element type.
func MapTo[T, U any](s Stream[T], fn func(T) U) Stream[U] {
	return Stream[U]{
		seq: func(yield func(U) bool) {
			for v := range s.seq {
				if !yield(fn(v)) {
					return
				}
			}
		},
	}
}

// FlatMap maps each element to a Stream and flattens the result.
func FlatMap[T, U any](s Stream[T], fn func(T) Stream[U]) Stream[U] {
	return Stream[U]{
		seq: func(yield func(U) bool) {
			for v := range s.seq {
				for u := range fn(v).seq {
					if !yield(u) {
						return
					}
				}
			}
		},
	}
}

// FlatMapSeq maps each element to an iter.Seq and flattens the result.
func FlatMapSeq[T, U any](s Stream[T], fn func(T) iter.Seq[U]) Stream[U] {
	return Stream[U]{
		seq: func(yield func(U) bool) {
			for v := range s.seq {
				for u := range fn(v) {
					if !yield(u) {
						return
					}
				}
			}
		},
	}
}

// Zip combines two Streams into a Stream of Pairs.
// The resulting stream ends when either input stream ends.
func Zip[T, U any](s1 Stream[T], s2 Stream[U]) Stream[Pair[T, U]] {
	return Stream[Pair[T, U]]{
		seq: func(yield func(Pair[T, U]) bool) {
			next1, stop1 := iter.Pull(s1.seq)
			defer stop1()
			next2, stop2 := iter.Pull(s2.seq)
			defer stop2()

			for {
				v1, ok1 := next1()
				v2, ok2 := next2()
				if !ok1 || !ok2 {
					return
				}
				if !yield(Pair[T, U]{First: v1, Second: v2}) {
					return
				}
			}
		},
	}
}

// ZipWithIndex adds an index to each element.
// Returns a Stream2[int, T] where the key is the index.
func ZipWithIndex[T any](s Stream[T]) Stream2[int, T] {
	return Stream2[int, T]{
		seq: func(yield func(int, T) bool) {
			idx := 0
			for v := range s.seq {
				if !yield(idx, v) {
					return
				}
				idx++
			}
		},
	}
}

// Distinct returns a Stream with duplicate elements removed.
// Elements must be comparable.
func Distinct[T comparable](s Stream[T]) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			seen := make(map[T]struct{})
			for v := range s.seq {
				if _, exists := seen[v]; !exists {
					seen[v] = struct{}{}
					if !yield(v) {
						return
					}
				}
			}
		},
	}
}

// DistinctBy returns a Stream with duplicates removed based on a key function.
// Elements are considered duplicates if they produce the same key.
func DistinctBy[T any, K comparable](s Stream[T], keyFn func(T) K) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			seen := make(map[K]struct{})
			for v := range s.seq {
				key := keyFn(v)
				if _, exists := seen[key]; !exists {
					seen[key] = struct{}{}
					if !yield(v) {
						return
					}
				}
			}
		},
	}
}

// SortedBy returns a Stream sorted by a key extracted from each element.
// Note: This is an eager operation that collects all elements into memory.
func SortedBy[T any, K cmp.Ordered](s Stream[T], keyFn func(T) K) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			collected := slices.Collect(s.seq)
			slices.SortFunc(collected, func(a, b T) int {
				return cmp.Compare(keyFn(a), keyFn(b))
			})
			for _, v := range collected {
				if !yield(v) {
					return
				}
			}
		},
	}
}

// Window returns a Stream of sliding windows of size n.
// Each window is a slice containing exactly n elements.
// The last few elements that don't form a complete window are not yielded.
// Note: This is a free function due to Go generics limitation with method return types.
func Window[T any](s Stream[T], size int) Stream[[]T] {
	if size <= 0 {
		return Empty[[]T]()
	}
	return Stream[[]T]{
		seq: func(yield func([]T) bool) {
			window := make([]T, 0, size)
			for v := range s.seq {
				window = append(window, v)
				if len(window) == size {
					// Make a copy to yield
					result := make([]T, size)
					copy(result, window)
					if !yield(result) {
						return
					}
					// Slide the window
					window = window[1:]
				}
			}
		},
	}
}

// Interleave combines two streams by alternating their elements.
// Elements are taken one at a time from each stream.
// When one stream is exhausted, remaining elements from the other stream are included.
func Interleave[T any](s1, s2 Stream[T]) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			next1, stop1 := iter.Pull(s1.seq)
			defer stop1()
			next2, stop2 := iter.Pull(s2.seq)
			defer stop2()

			for {
				v1, ok1 := next1()
				if ok1 {
					if !yield(v1) {
						return
					}
				}
				v2, ok2 := next2()
				if ok2 {
					if !yield(v2) {
						return
					}
				}
				if !ok1 && !ok2 {
					return
				}
			}
		},
	}
}

// Scan applies an accumulator function over the stream and yields each intermediate result.
// This is a generalized version of RunningSum/RunningProduct that works with any accumulator.
// The first yielded value is fn(init, first_element).
func Scan[T any, A any](s Stream[T], init A, fn func(A, T) A) Stream[A] {
	return Stream[A]{
		seq: func(yield func(A) bool) {
			acc := init
			for v := range s.seq {
				acc = fn(acc, v)
				if !yield(acc) {
					return
				}
			}
		},
	}
}

// Step returns a Stream that yields every nth element (starting from the first).
// If n <= 1, returns the original stream unchanged.
func (s Stream[T]) Step(n int) Stream[T] {
	if n <= 1 {
		return s
	}
	return Stream[T]{
		seq: func(yield func(T) bool) {
			i := 0
			for v := range s.seq {
				if i%n == 0 {
					if !yield(v) {
						return
					}
				}
				i++
			}
		},
	}
}

// DistinctUntilChanged returns a Stream that removes consecutive duplicate elements.
// Only adjacent duplicates are removed; the same value appearing later is kept.
// Elements must be comparable.
func DistinctUntilChanged[T comparable](s Stream[T]) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			var (
				prev  T
				first = true
			)
			for v := range s.seq {
				if first || prev != v {
					if !yield(v) {
						return
					}
					prev, first = v, false
				}
			}
		},
	}
}

// DistinctUntilChangedBy returns a Stream that removes consecutive elements
// that produce the same key. Uses the provided equality function for comparison.
func DistinctUntilChangedBy[T any](s Stream[T], eq func(a, b T) bool) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			var (
				prev  T
				first = true
			)
			for v := range s.seq {
				if first || !eq(prev, v) {
					if !yield(v) {
						return
					}
					prev, first = v, false
				}
			}
		},
	}
}

// TakeLast returns a Stream containing the last n elements.
// Uses a ring buffer for O(L) time complexity where L is the input length.
// If n <= 0, returns an empty stream.
// Note: The stream must be fully consumed before any elements are yielded.
func (s Stream[T]) TakeLast(n int) Stream[T] {
	if n <= 0 {
		return Empty[T]()
	}
	return Stream[T]{
		seq: func(yield func(T) bool) {
			buf := make([]T, n)
			head := 0  // write position
			count := 0 // number of elements seen
			for v := range s.seq {
				buf[head] = v
				head = (head + 1) % n
				count++
			}
			// Calculate actual elements to yield
			size := min(count, n)
			// Start position: if count < n, start from 0; otherwise start from head
			start := 0
			if count >= n {
				start = head // head points to oldest element after wraparound
			}
			for i := range size {
				if !yield(buf[(start+i)%n]) {
					return
				}
			}
		},
	}
}

// DropLast returns a Stream with the last n elements removed.
// Uses a ring buffer for O(L) time complexity where L is the input length.
// If n <= 0, returns the original stream unchanged.
func (s Stream[T]) DropLast(n int) Stream[T] {
	if n <= 0 {
		return s
	}
	return Stream[T]{
		seq: func(yield func(T) bool) {
			buf := make([]T, n)
			head := 0  // write position
			count := 0 // number of elements buffered
			for v := range s.seq {
				if count >= n {
					// Buffer is full, yield the oldest element before overwriting
					if !yield(buf[head]) {
						return
					}
				} else {
					count++
				}
				buf[head] = v
				head = (head + 1) % n
			}
			// Remaining elements in buffer are dropped
		},
	}
}

// WindowWithStep returns a Stream of sliding windows with configurable step size.
//
// Parameters:
//   - size: the number of elements in each window
//   - step: how many elements to advance between windows
//   - allowPartial: if true, yields partial windows at the end; if false, only full windows
//
// Behavior:
//   - step < size: overlapping windows (sliding window)
//   - step == size: non-overlapping chunks (same as Chunk)
//   - step > size: windows with gaps (elements between windows are skipped)
//
// Each yielded window is an independent copy (safe to retain).
// Note: This is a free function due to Go generics limitation with method return types.
func WindowWithStep[T any](s Stream[T], size, step int, allowPartial bool) Stream[[]T] {
	if size <= 0 || step <= 0 {
		return Empty[[]T]()
	}
	return Stream[[]T]{
		seq: func(yield func([]T) bool) {
			win := make([]T, 0, size)
			skip := 0 // number of elements to skip before starting next window
			for v := range s.seq {
				if skip > 0 {
					skip--
					continue
				}
				win = append(win, v)
				if len(win) == size {
					// Make a copy to yield
					cp := make([]T, size)
					copy(cp, win)
					if !yield(cp) {
						return
					}
					// Slide by step
					if step >= size {
						// Need to skip (step - size) elements before next window
						skip = step - size
						win = win[:0]
					} else {
						win = win[step:]
					}
				}
			}
			// Handle partial window at the end
			if allowPartial && len(win) > 0 {
				cp := make([]T, len(win))
				copy(cp, win)
				yield(cp)
			}
		},
	}
}

// Pairwise returns a Stream of consecutive pairs (sliding window of size 2).
// For input [a, b, c, d], yields [(a,b), (b,c), (c,d)].
func Pairwise[T any](s Stream[T]) Stream[Pair[T, T]] {
	return Stream[Pair[T, T]]{
		seq: func(yield func(Pair[T, T]) bool) {
			var prev T
			first := true
			for v := range s.seq {
				if !first {
					if !yield(Pair[T, T]{First: prev, Second: v}) {
						return
					}
				}
				prev, first = v, false
			}
		},
	}
}

// Triples returns a Stream of consecutive triples (sliding window of size 3).
// For input [a, b, c, d, e], yields [(a,b,c), (b,c,d), (c,d,e)].
func Triples[T any](s Stream[T]) Stream[Triple[T, T, T]] {
	return Stream[Triple[T, T, T]]{
		seq: func(yield func(Triple[T, T, T]) bool) {
			var p1, p2 T
			count := 0
			for v := range s.seq {
				if count >= 2 {
					if !yield(Triple[T, T, T]{First: p1, Second: p2, Third: v}) {
						return
					}
				}
				p1, p2 = p2, v
				count++
			}
		},
	}
}

// SortedStable returns a Stream with elements sorted using the given comparison function.
// Unlike Sorted, this maintains the relative order of equal elements (stable sort).
// Note: This is an eager operation that collects all elements into memory.
func (s Stream[T]) SortedStable(cmp func(a, b T) int) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			collected := slices.Collect(s.seq)
			slices.SortStableFunc(collected, cmp)
			for _, v := range collected {
				if !yield(v) {
					return
				}
			}
		},
	}
}

// SortedStableBy returns a Stream sorted by a key extracted from each element.
// Unlike SortedBy, this maintains the relative order of elements with equal keys (stable sort).
// Note: This is an eager operation that collects all elements into memory.
func SortedStableBy[T any, K cmp.Ordered](s Stream[T], keyFn func(T) K) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			collected := slices.Collect(s.seq)
			slices.SortStableFunc(collected, func(a, b T) int {
				return cmp.Compare(keyFn(a), keyFn(b))
			})
			for _, v := range collected {
				if !yield(v) {
					return
				}
			}
		},
	}
}

// Flatten flattens a Stream of slices into a single Stream.
func Flatten[T any](s Stream[[]T]) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			for slice := range s.seq {
				for _, v := range slice {
					if !yield(v) {
						return
					}
				}
			}
		},
	}
}

// FlattenSeq flattens a Stream of iter.Seq into a single Stream.
func FlattenSeq[T any](s Stream[iter.Seq[T]]) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			for innerSeq := range s.seq {
				for v := range innerSeq {
					if !yield(v) {
						return
					}
				}
			}
		},
	}
}

// Intersperse inserts a separator element between each element of the stream.
// For input [a, b, c] with separator x, yields [a, x, b, x, c].
func (s Stream[T]) Intersperse(sep T) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			first := true
			for v := range s.seq {
				if !first {
					if !yield(sep) {
						return
					}
				}
				if !yield(v) {
					return
				}
				first = false
			}
		},
	}
}

// --- Specialized Operations ---

// MergeSorted merges two sorted streams into one sorted stream.
// Both input streams must be sorted according to the same comparison function.
// The comparison function should return negative if a < b, zero if a == b, positive if a > b.
func MergeSorted[T any](s1, s2 Stream[T], cmp func(a, b T) int) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			next1, stop1 := iter.Pull(s1.seq)
			defer stop1()
			next2, stop2 := iter.Pull(s2.seq)
			defer stop2()

			v1, ok1 := next1()
			v2, ok2 := next2()

			for ok1 && ok2 {
				if cmp(v1, v2) <= 0 {
					if !yield(v1) {
						return
					}
					v1, ok1 = next1()
				} else {
					if !yield(v2) {
						return
					}
					v2, ok2 = next2()
				}
			}

			// Drain remaining from s1
			for ok1 {
				if !yield(v1) {
					return
				}
				v1, ok1 = next1()
			}

			// Drain remaining from s2
			for ok2 {
				if !yield(v2) {
					return
				}
				v2, ok2 = next2()
			}
		},
	}
}

// MergeSortedN merges multiple sorted streams into one sorted stream.
// All input streams must be sorted according to the same comparison function.
//
// Complexity: Uses pairwise merge with O(n * k) comparisons where n is total elements
// and k is number of streams. For large k with many streams, use MergeSortedNHeap
// for O(n log k) complexity.
func MergeSortedN[T any](cmp func(a, b T) int, streams ...Stream[T]) Stream[T] {
	if len(streams) == 0 {
		return Empty[T]()
	}
	if len(streams) == 1 {
		return streams[0]
	}

	// Merge pairwise
	result := streams[0]
	for i := 1; i < len(streams); i++ {
		result = MergeSorted(result, streams[i], cmp)
	}
	return result
}

// mergeHeapItem holds a value and its source stream iterator for heap-based merge.
type mergeHeapItem[T any] struct {
	value T
	next  func() (T, bool)
	stop  func()
}

// mergeHeap implements a min-heap for k-way merge.
type mergeHeap[T any] struct {
	items []*mergeHeapItem[T]
	cmp   func(a, b T) int
}

func (h *mergeHeap[T]) Len() int { return len(h.items) }

func (h *mergeHeap[T]) Less(i, j int) bool {
	return h.cmp(h.items[i].value, h.items[j].value) < 0
}

func (h *mergeHeap[T]) Swap(i, j int) {
	h.items[i], h.items[j] = h.items[j], h.items[i]
}

func (h *mergeHeap[T]) Push(x *mergeHeapItem[T]) {
	h.items = append(h.items, x)
	h.up(len(h.items) - 1)
}

func (h *mergeHeap[T]) Pop() *mergeHeapItem[T] {
	if len(h.items) == 0 {
		return nil
	}
	item := h.items[0]
	n := len(h.items) - 1
	h.items[0] = h.items[n]
	h.items = h.items[:n]
	if n > 0 {
		h.down(0)
	}
	return item
}

func (h *mergeHeap[T]) up(i int) {
	for i > 0 {
		parent := (i - 1) / 2
		if !h.Less(i, parent) {
			break
		}
		h.Swap(i, parent)
		i = parent
	}
}

func (h *mergeHeap[T]) down(i int) {
	n := len(h.items)
	for {
		left := 2*i + 1
		if left >= n {
			break
		}
		smallest := left
		if right := left + 1; right < n && h.Less(right, left) {
			smallest = right
		}
		if !h.Less(smallest, i) {
			break
		}
		h.Swap(i, smallest)
		i = smallest
	}
}

// MergeSortedNHeap merges multiple sorted streams using a heap-based k-way merge.
// All input streams must be sorted according to the same comparison function.
//
// Complexity: O(n log k) where n is total elements and k is number of streams.
// Preferred over MergeSortedN when k is large (e.g., k > 8).
func MergeSortedNHeap[T any](cmp func(a, b T) int, streams ...Stream[T]) Stream[T] {
	if len(streams) == 0 {
		return Empty[T]()
	}
	if len(streams) == 1 {
		return streams[0]
	}
	// For small k, pairwise merge may be faster due to lower overhead
	if len(streams) <= 4 {
		return MergeSortedN(cmp, streams...)
	}

	return Stream[T]{
		seq: func(yield func(T) bool) {
			heap := &mergeHeap[T]{
				items: make([]*mergeHeapItem[T], 0, len(streams)),
				cmp:   cmp,
			}

			// Initialize heap with first element from each stream
			stopFuncs := make([]func(), 0, len(streams))
			for _, s := range streams {
				next, stop := iter.Pull(s.seq)
				stopFuncs = append(stopFuncs, stop)
				if v, ok := next(); ok {
					heap.Push(&mergeHeapItem[T]{value: v, next: next, stop: stop})
				}
			}

			// Ensure all streams are stopped on exit
			defer func() {
				for _, stop := range stopFuncs {
					stop()
				}
			}()

			// Extract minimum and refill from same stream
			for heap.Len() > 0 {
				item := heap.Pop()
				if !yield(item.value) {
					return
				}
				// Try to get next value from same stream
				if v, ok := item.next(); ok {
					item.value = v
					heap.Push(item)
				}
			}
		},
	}
}

// ZipLongest combines two streams, continuing until both are exhausted.
// Missing elements are represented as None in the Optional.
func ZipLongest[T, U any](s1 Stream[T], s2 Stream[U]) Stream[Pair[Optional[T], Optional[U]]] {
	return Stream[Pair[Optional[T], Optional[U]]]{
		seq: func(yield func(Pair[Optional[T], Optional[U]]) bool) {
			next1, stop1 := iter.Pull(s1.seq)
			defer stop1()
			next2, stop2 := iter.Pull(s2.seq)
			defer stop2()

			for {
				v1, ok1 := next1()
				v2, ok2 := next2()

				if !ok1 && !ok2 {
					return
				}

				var (
					opt1 Optional[T]
					opt2 Optional[U]
				)
				if ok1 {
					opt1 = Some(v1)
				} else {
					opt1 = None[T]()
				}
				if ok2 {
					opt2 = Some(v2)
				} else {
					opt2 = None[U]()
				}

				if !yield(Pair[Optional[T], Optional[U]]{First: opt1, Second: opt2}) {
					return
				}
			}
		},
	}
}

// ZipLongestWith combines two streams with default values for missing elements.
func ZipLongestWith[T, U any](s1 Stream[T], s2 Stream[U], defaultT T, defaultU U) Stream[Pair[T, U]] {
	return Stream[Pair[T, U]]{
		seq: func(yield func(Pair[T, U]) bool) {
			next1, stop1 := iter.Pull(s1.seq)
			defer stop1()
			next2, stop2 := iter.Pull(s2.seq)
			defer stop2()

			for {
				v1, ok1 := next1()
				v2, ok2 := next2()

				if !ok1 && !ok2 {
					return
				}

				if !ok1 {
					v1 = defaultT
				}
				if !ok2 {
					v2 = defaultU
				}

				if !yield(Pair[T, U]{First: v1, Second: v2}) {
					return
				}
			}
		},
	}
}

// Cartesian returns the Cartesian product of two streams.
// Note: The second stream is collected into memory as it needs to be iterated multiple times.
func Cartesian[T, U any](s1 Stream[T], s2 Stream[U]) Stream[Pair[T, U]] {
	return Stream[Pair[T, U]]{
		seq: func(yield func(Pair[T, U]) bool) {
			// Collect s2 since we need to iterate it multiple times
			s2Collected := s2.Collect()
			if len(s2Collected) == 0 {
				return
			}

			for v1 := range s1.seq {
				for _, v2 := range s2Collected {
					if !yield(Pair[T, U]{First: v1, Second: v2}) {
						return
					}
				}
			}
		},
	}
}

// CartesianSelf returns the Cartesian product of a stream with itself.
// Note: The stream is collected into memory.
func CartesianSelf[T any](s Stream[T]) Stream[Pair[T, T]] {
	return Stream[Pair[T, T]]{
		seq: func(yield func(Pair[T, T]) bool) {
			collected := s.Collect()
			for _, v1 := range collected {
				for _, v2 := range collected {
					if !yield(Pair[T, T]{First: v1, Second: v2}) {
						return
					}
				}
			}
		},
	}
}

// CrossProduct returns the Cartesian product of multiple streams of the same type.
// Note: All streams are collected into memory.
func CrossProduct[T any](streams ...Stream[T]) Stream[[]T] {
	if len(streams) == 0 {
		return Empty[[]T]()
	}
	if len(streams) == 1 {
		return MapTo(streams[0], func(v T) []T { return []T{v} })
	}

	return Stream[[]T]{
		seq: func(yield func([]T) bool) {
			// Collect all streams except the first
			collected := make([][]T, len(streams))
			for i, s := range streams {
				collected[i] = s.Collect()
				if len(collected[i]) == 0 {
					return // Empty product
				}
			}

			// Generate all combinations
			indices := make([]int, len(collected))
			for {
				// Build current combination
				combo := make([]T, len(collected))
				for i, idx := range indices {
					combo[i] = collected[i][idx]
				}

				if !yield(combo) {
					return
				}

				// Increment indices (like counting)
				i := len(indices) - 1
				for i >= 0 {
					indices[i]++
					if indices[i] < len(collected[i]) {
						break
					}
					indices[i] = 0
					i--
				}
				if i < 0 {
					return // All combinations exhausted
				}
			}
		},
	}
}

// Combinations returns all k-combinations of elements from the stream.
// Note: The stream is collected into memory.
func Combinations[T any](s Stream[T], k int) Stream[[]T] {
	if k <= 0 {
		return Empty[[]T]()
	}

	return Stream[[]T]{
		seq: func(yield func([]T) bool) {
			elements := s.Collect()
			n := len(elements)
			if k > n {
				return
			}

			// Generate combinations using indices
			indices := make([]int, k)
			for i := range indices {
				indices[i] = i
			}

			for {
				// Build current combination
				combo := make([]T, k)
				for i, idx := range indices {
					combo[i] = elements[idx]
				}

				if !yield(combo) {
					return
				}

				// Find rightmost index that can be incremented
				i := k - 1
				for i >= 0 && indices[i] == n-k+i {
					i--
				}
				if i < 0 {
					return
				}

				// Increment and reset indices to the right
				indices[i]++
				for j := i + 1; j < k; j++ {
					indices[j] = indices[j-1] + 1
				}
			}
		},
	}
}

// Permutations returns all permutations of elements from the stream.
// Note: The stream is collected into memory.
func Permutations[T any](s Stream[T]) Stream[[]T] {
	return Stream[[]T]{
		seq: func(yield func([]T) bool) {
			elements := s.Collect()
			n := len(elements)
			if n == 0 {
				return
			}

			// Generate permutations using Heap's algorithm
			c := make([]int, n)
			perm := make([]T, n)
			copy(perm, elements)

			// Yield initial permutation
			result := make([]T, n)
			copy(result, perm)
			if !yield(result) {
				return
			}

			i := 0
			for i < n {
				if c[i] < i {
					if i%2 == 0 {
						perm[0], perm[i] = perm[i], perm[0]
					} else {
						perm[c[i]], perm[i] = perm[i], perm[c[i]]
					}

					result := make([]T, n)
					copy(result, perm)
					if !yield(result) {
						return
					}

					c[i]++
					i = 0
				} else {
					c[i] = 0
					i++
				}
			}
		},
	}
}

// Using ensures that the resource is closed after the function completes.
// This is similar to try-with-resources in Java or using in C#.
//
// The resource must implement the Close() error interface.
// The Close method is called defer-style, ensuring it runs even if fn panics.
func Using[T interface{ Close() error }, R any](resource T, fn func(T) R) R {
	defer func() { _ = resource.Close() }()
	return fn(resource)
}
