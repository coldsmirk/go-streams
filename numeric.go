package streams

import "cmp"

// Numeric is a constraint that includes all numeric types.
type Numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

// Signed is a constraint for signed integer types.
type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// Unsigned is a constraint for unsigned integer types.
type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// Integer is a constraint for all integer types.
type Integer interface {
	Signed | Unsigned
}

// Float is a constraint for floating-point types.
type Float interface {
	~float32 | ~float64
}

// --- Numeric Stream Functions ---

// Sum returns the sum of all numeric elements.
func Sum[T Numeric](s Stream[T]) T {
	var sum T
	for v := range s.seq {
		sum += v
	}
	return sum
}

// Average returns the average of all numeric elements.
// Returns None for an empty stream.
func Average[T Numeric](s Stream[T]) Optional[float64] {
	var sum float64
	count := 0
	for v := range s.seq {
		sum += float64(v)
		count++
	}
	if count == 0 {
		return None[float64]()
	}
	return Some(sum / float64(count))
}

// MinValue returns the minimum value from a stream of ordered elements.
// Returns None for an empty stream.
func MinValue[T cmp.Ordered](s Stream[T]) Optional[T] {
	var smallest T
	first := true
	for v := range s.seq {
		if first {
			smallest = v
			first = false
		} else if v < smallest {
			smallest = v
		}
	}
	if first {
		return None[T]()
	}
	return Some(smallest)
}

// MaxValue returns the maximum value from a stream of ordered elements.
// Returns None for an empty stream.
func MaxValue[T cmp.Ordered](s Stream[T]) Optional[T] {
	var largest T
	first := true
	for v := range s.seq {
		if first {
			largest = v
			first = false
		} else if v > largest {
			largest = v
		}
	}
	if first {
		return None[T]()
	}
	return Some(largest)
}

// MinMax returns both the minimum and maximum values.
// Returns None for an empty stream.
func MinMax[T cmp.Ordered](s Stream[T]) Optional[Pair[T, T]] {
	var smallest, largest T
	first := true
	for v := range s.seq {
		if first {
			smallest = v
			largest = v
			first = false
		} else {
			if v < smallest {
				smallest = v
			}
			if v > largest {
				largest = v
			}
		}
	}
	if first {
		return None[Pair[T, T]]()
	}
	return Some(Pair[T, T]{First: smallest, Second: largest})
}

// Product returns the product of all numeric elements.
// Returns 1 for an empty stream.
func Product[T Numeric](s Stream[T]) T {
	var product T = 1
	for v := range s.seq {
		product *= v
	}
	return product
}

// SumBy sums the results of applying a function to each element.
func SumBy[T any, N Numeric](s Stream[T], fn func(T) N) N {
	var sum N
	for v := range s.seq {
		sum += fn(v)
	}
	return sum
}

// AverageBy computes the average of the results of applying a function to each element.
func AverageBy[T any, N Numeric](s Stream[T], fn func(T) N) Optional[float64] {
	var sum float64
	count := 0
	for v := range s.seq {
		sum += float64(fn(v))
		count++
	}
	if count == 0 {
		return None[float64]()
	}
	return Some(sum / float64(count))
}

// MinBy returns the element that produces the minimum value when the function is applied.
func MinBy[T any, K cmp.Ordered](s Stream[T], fn func(T) K) Optional[T] {
	var minElem T
	var minKey K
	first := true
	for v := range s.seq {
		key := fn(v)
		if first {
			minElem = v
			minKey = key
			first = false
		} else if key < minKey {
			minElem = v
			minKey = key
		}
	}
	if first {
		return None[T]()
	}
	return Some(minElem)
}

// MaxBy returns the element that produces the maximum value when the function is applied.
func MaxBy[T any, K cmp.Ordered](s Stream[T], fn func(T) K) Optional[T] {
	var maxElem T
	var maxKey K
	first := true
	for v := range s.seq {
		key := fn(v)
		if first {
			maxElem = v
			maxKey = key
			first = false
		} else if key > maxKey {
			maxElem = v
			maxKey = key
		}
	}
	if first {
		return None[T]()
	}
	return Some(maxElem)
}

// Statistics holds basic statistics about a numeric stream.
type Statistics[T Numeric] struct {
	Count   int
	Sum     T
	Min     T
	Max     T
	Average float64
}

// GetStatistics computes basic statistics for a numeric stream.
// Returns None for an empty stream.
func GetStatistics[T Numeric](s Stream[T]) Optional[Statistics[T]] {
	var sum, smallest, largest T
	count := 0
	first := true

	for v := range s.seq {
		sum += v
		count++
		if first {
			smallest = v
			largest = v
			first = false
		} else {
			if v < smallest {
				smallest = v
			}
			if v > largest {
				largest = v
			}
		}
	}

	if count == 0 {
		return None[Statistics[T]]()
	}

	return Some(Statistics[T]{
		Count:   count,
		Sum:     sum,
		Min:     smallest,
		Max:     largest,
		Average: float64(sum) / float64(count),
	})
}

// RunningSum returns a Stream of cumulative sums.
func RunningSum[T Numeric](s Stream[T]) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			var sum T
			for v := range s.seq {
				sum += v
				if !yield(sum) {
					return
				}
			}
		},
	}
}

// RunningProduct returns a Stream of cumulative products.
func RunningProduct[T Numeric](s Stream[T]) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			var product T = 1
			for v := range s.seq {
				product *= v
				if !yield(product) {
					return
				}
			}
		},
	}
}

// Differences returns a Stream of differences between consecutive elements.
// The first difference is between the second and first elements.
func Differences[T Numeric](s Stream[T]) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			var prev T
			first := true
			for v := range s.seq {
				if first {
					prev = v
					first = false
					continue
				}
				if !yield(v - prev) {
					return
				}
				prev = v
			}
		},
	}
}

// Clamp returns a Stream where each element is clamped to [min, max].
func Clamp[T cmp.Ordered](s Stream[T], minVal, maxVal T) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			for v := range s.seq {
				clamped := v
				if clamped < minVal {
					clamped = minVal
				} else if clamped > maxVal {
					clamped = maxVal
				}
				if !yield(clamped) {
					return
				}
			}
		},
	}
}

// Abs returns a Stream of absolute values (for signed numeric types).
func Abs[T Signed](s Stream[T]) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			for v := range s.seq {
				if v < 0 {
					v = -v
				}
				if !yield(v) {
					return
				}
			}
		},
	}
}

// AbsFloat returns a Stream of absolute values for floating-point types.
func AbsFloat[T Float](s Stream[T]) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			for v := range s.seq {
				if v < 0 {
					v = -v
				}
				if !yield(v) {
					return
				}
			}
		},
	}
}

// Scale multiplies each element by a scalar.
func Scale[T Numeric](s Stream[T], factor T) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			for v := range s.seq {
				if !yield(v * factor) {
					return
				}
			}
		},
	}
}

// Offset adds an offset to each element.
func Offset[T Numeric](s Stream[T], offset T) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			for v := range s.seq {
				if !yield(v + offset) {
					return
				}
			}
		},
	}
}

// Positive filters to only positive values (> 0).
func Positive[T Numeric](s Stream[T]) Stream[T] {
	var zero T
	return s.Filter(func(v T) bool { return v > zero })
}

// Negative filters to only negative values (< 0).
func Negative[T Signed](s Stream[T]) Stream[T] {
	var zero T
	return s.Filter(func(v T) bool { return v < zero })
}

// NonZero filters out zero values.
func NonZero[T Numeric](s Stream[T]) Stream[T] {
	var zero T
	return s.Filter(func(v T) bool { return v != zero })
}
