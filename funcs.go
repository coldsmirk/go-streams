package streams

import (
	"cmp"
	"iter"
	"slices"
)

// Numeric is a constraint for the types that support the arithmetic used by
// [Sum], [Product] and [Average]. Complex types are excluded: they are not
// ordered and an average over them is rarely meaningful.
type Numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

// --- constrained element types: the counterparts to the Func methods ---

// Sort returns a Stream of the elements in ascending order. It buffers the whole
// sequence, so it is not usable on an infinite Stream. For an element type that
// is not ordered, use [Stream.SortFunc].
func Sort[T cmp.Ordered](s Stream[T]) Stream[T] {
	return func(yield func(T) bool) {
		// slices.Sorted, not SortFunc(cmp.Compare): a comparator turns every
		// comparison into an indirect call, which costs several times the
		// direct one the constraint already permits.
		for _, v := range slices.Sorted(iter.Seq[T](s)) {
			if !yield(v) {
				return
			}
		}
	}
}

// Min returns the smallest element, or false if the Stream is empty.
func Min[T cmp.Ordered](s Stream[T]) (T, bool) {
	var best T
	empty := true
	for v := range s {
		// cmp.Less, not <, so that NaN orders the same way cmp.Compare does.
		if empty || cmp.Less(v, best) {
			best, empty = v, false
		}
	}
	return best, !empty
}

// Max returns the largest element, or false if the Stream is empty.
func Max[T cmp.Ordered](s Stream[T]) (T, bool) {
	var best T
	empty := true
	for v := range s {
		if empty || cmp.Less(best, v) {
			best, empty = v, false
		}
	}
	return best, !empty
}

// Compact returns a Stream omitting each element equal to the one before it.
// Like slices.Compact, it removes only adjacent duplicates.
func Compact[T comparable](s Stream[T]) Stream[T] {
	return s.CompactFunc(func(a, b T) bool { return a == b })
}

// Distinct returns a Stream omitting every element that has already appeared.
// Elements are held for the duration of the iteration.
func Distinct[T comparable](s Stream[T]) Stream[T] {
	return s.DistinctBy(func(v T) T { return v })
}

// Contains reports whether any element equals target. It stops at the first
// match.
func Contains[T comparable](s Stream[T], target T) bool {
	for v := range s {
		if v == target {
			return true
		}
	}
	return false
}

// Frequency returns a map from each distinct element to the number of times it
// appears.
func Frequency[T comparable](s Stream[T]) map[T]int {
	out := make(map[T]int)
	for v := range s {
		out[v]++
	}
	return out
}

// Sum returns the sum of the elements, or the zero value for an empty Stream.
func Sum[T Numeric](s Stream[T]) T {
	var total T
	for v := range s {
		total += v
	}
	return total
}

// Product returns the product of the elements, or one for an empty Stream.
func Product[T Numeric](s Stream[T]) T {
	var product T = 1
	for v := range s {
		product *= v
	}
	return product
}

// Average returns the arithmetic mean of the elements, or false if the Stream
// is empty.
func Average[T Numeric](s Stream[T]) (float64, bool) {
	var total float64
	n := 0
	for v := range s {
		total += float64(v)
		n++
	}
	if n == 0 {
		return 0, false
	}
	return total / float64(n), true
}

// --- regrouping: the result element type wraps T, so these cannot be methods ---

// Chunk returns a Stream of consecutive, non-overlapping slices of up to n
// elements. The final slice is short if the Stream does not divide evenly.
// Chunk panics if n is not positive.
func Chunk[T any](s Stream[T], n int) Stream[[]T] {
	if n < 1 {
		panic("streams: Chunk called with n < 1")
	}
	return func(yield func([]T) bool) {
		buf := make([]T, 0, n)
		for v := range s {
			if buf = append(buf, v); len(buf) == n {
				if !yield(buf) {
					return
				}
				buf = make([]T, 0, n)
			}
		}
		if len(buf) > 0 {
			yield(buf)
		}
	}
}

// Window returns a Stream of overlapping slices of n consecutive elements,
// advancing one element at a time. It yields nothing if the Stream is shorter
// than n. Window panics if n is not positive.
//
// The windows are cut from a shared backing array rather than allocated one at
// a time, which is what keeps a sliding window from costing an allocation and a
// copy per element. Two windows therefore share the elements they overlap on:
// reading them is unaffected, but writing through one window is visible in its
// neighbours, and holding on to one keeps a block of elements alive. No window
// is ever overwritten once yielded. Where independent slices are wanted, ask
// for them:
//
//	streams.Window(s, 16).Map(slices.Clone)
//
// [Chunk] does allocate per chunk, because its chunks do not overlap and
// sharing an array measurably costs more there than it saves.
func Window[T any](s Stream[T], n int) Stream[[]T] {
	if n < 1 {
		panic("streams: Window called with n < 1")
	}
	return func(yield func([]T) bool) {
		// Room for windowsPerArena windows before a fresh array is needed.
		const windowsPerArena = 64
		buf := make([]T, 0, n+windowsPerArena)
		for v := range s {
			if len(buf) == cap(buf) {
				// The old array stays alive for as long as the windows cut
				// from it do; carry the tail forward so the next window is
				// still complete.
				fresh := make([]T, n-1, n+windowsPerArena)
				copy(fresh, buf[len(buf)-(n-1):])
				buf = fresh
			}
			buf = append(buf, v)
			if len(buf) < n {
				continue
			}
			// Capped so that appending to a window cannot reach into the
			// elements the next one will occupy.
			if !yield(buf[len(buf)-n : len(buf) : len(buf)]) {
				return
			}
		}
	}
}

// Flatten returns a Stream of the elements of every Stream in s. It destructures
// the element type, so it cannot be a method.
func Flatten[T any](s Stream[Stream[T]]) Stream[T] {
	return s.FlatMap(func(inner Stream[T]) Stream[T] { return inner })
}

// --- combining several streams ---

// Concat returns a Stream of the elements of each Stream in turn.
func Concat[T any](ss ...Stream[T]) Stream[T] {
	return func(yield func(T) bool) {
		for _, s := range ss {
			for v := range s {
				if !yield(v) {
					return
				}
			}
		}
	}
}

// Interleave returns a Stream alternating between the elements of a and b. When
// one is exhausted the remainder of the other follows.
func Interleave[T any](a, b Stream[T]) Stream[T] {
	return func(yield func(T) bool) {
		// Only b is pulled. Ranging a costs nothing, while iter.Pull runs a
		// coroutine switch per element, so pulling both sides doubles the
		// dominant cost of this operator for no reason.
		next, stop := iter.Pull(iter.Seq[T](b))
		defer stop()
		exhausted := false
		for va := range a {
			if !yield(va) {
				return
			}
			if exhausted {
				continue
			}
			vb, ok := next()
			if !ok {
				exhausted = true
				continue
			}
			if !yield(vb) {
				return
			}
		}
		for !exhausted {
			vb, ok := next()
			if !ok || !yield(vb) {
				return
			}
		}
	}
}

// Merge returns a Stream of the elements of every Stream in ss in the order
// given by compare. Each input must already be sorted by compare; if it is not,
// the output is not sorted either.
func Merge[T any](compare func(a, b T) int, ss ...Stream[T]) Stream[T] {
	return func(yield func(T) bool) {
		h := make(mergeHeap[T], 0, len(ss))
		for _, s := range ss {
			next, stop := iter.Pull(iter.Seq[T](s))
			defer stop()
			if v, ok := next(); ok {
				h = append(h, mergeSource[T]{value: v, next: next})
			}
		}
		h.init(compare)
		for len(h) > 0 {
			if !yield(h[0].value) {
				return
			}
			if v, ok := h[0].next(); ok {
				h[0].value = v
				h.down(0, compare)
			} else {
				last := len(h) - 1
				h[0] = h[last]
				h = h[:last]
				h.down(0, compare)
			}
		}
	}
}

// Cycle returns an infinite Stream repeating the elements of s. It buffers the
// elements on the first pass, so s must be finite.
func Cycle[T any](s Stream[T]) Stream[T] {
	return func(yield func(T) bool) {
		var buf []T
		for v := range s {
			buf = append(buf, v)
			if !yield(v) {
				return
			}
		}
		if len(buf) == 0 {
			return
		}
		for {
			for _, v := range buf {
				if !yield(v) {
					return
				}
			}
		}
	}
}

// --- fallible sequences ---

// TryMap applies fn to each element and returns the results paired with the
// error fn reported. The result is a plain iter.Seq2, so it can be ranged over
// directly:
//
//	for v, err := range streams.TryMap(s, parse) { ... }
func TryMap[T, R any](s Stream[T], fn func(T) (R, error)) iter.Seq2[R, error] {
	return func(yield func(R, error) bool) {
		for v := range s {
			if !yield(fn(v)) {
				return
			}
		}
	}
}

// Ok returns a Stream of the values of seq up to its first error, together with
// a function reporting that error. Consume the Stream, then call the function,
// the way bufio.Scanner pairs Scan with Err:
//
//	lines, readErr := streams.Ok(source.LinesFile(path))
//	n := lines.Filter(nonBlank).Count()
//	if err := readErr(); err != nil {
//		return err
//	}
//
// It is the lazy counterpart of [Try]. Try holds every value in a slice, so it
// cannot read a source larger than memory; Ok lets a fallible source feed a
// pipeline one element at a time. The Stream ends at the first error, and the
// function reports nil until one is reached, so it is only meaningful once the
// Stream has been consumed. Like every Stream it is single-pass, and like
// bufio.Scanner it is not safe to consume from one goroutine while calling the
// function from another.
func Ok[T any](seq iter.Seq2[T, error]) (Stream[T], func() error) {
	var failure error
	s := Stream[T](func(yield func(T) bool) {
		for v, err := range seq {
			if err != nil {
				failure = err
				return
			}
			if !yield(v) {
				return
			}
		}
	})
	return s, func() error { return failure }
}

// Try collects the values of seq, stopping at and returning the first error.
func Try[T any](seq iter.Seq2[T, error]) ([]T, error) {
	var out []T
	for v, err := range seq {
		if err != nil {
			return out, err
		}
		out = append(out, v)
	}
	return out, nil
}

// --- k-way merge heap ---

type mergeSource[T any] struct {
	value T
	next  func() (T, bool)
}

type mergeHeap[T any] []mergeSource[T]

func (h mergeHeap[T]) init(compare func(a, b T) int) {
	for i := len(h)/2 - 1; i >= 0; i-- {
		h.down(i, compare)
	}
}

func (h mergeHeap[T]) down(i int, compare func(a, b T) int) {
	for {
		smallest := i
		if l := 2*i + 1; l < len(h) && compare(h[l].value, h[smallest].value) < 0 {
			smallest = l
		}
		if r := 2*i + 2; r < len(h) && compare(h[r].value, h[smallest].value) < 0 {
			smallest = r
		}
		if smallest == i {
			return
		}
		h[i], h[smallest] = h[smallest], h[i]
		i = smallest
	}
}
