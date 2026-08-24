package streams

import (
	"bufio"
	"context"
	"io"
)

// --- Context-Aware Stream Wrapper ---

// WithContext wraps a Stream to respect context cancellation.
// When the context is cancelled, the stream will stop yielding elements.
func WithContext[T any](ctx context.Context, s Stream[T]) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			for v := range s.seq {
				select {
				case <-ctx.Done():
					return
				default:
					if !yield(v) {
						return
					}
				}
			}
		},
	}
}

// WithContext2 wraps a Stream2 to respect context cancellation.
func WithContext2[K, V any](ctx context.Context, s Stream2[K, V]) Stream2[K, V] {
	return Stream2[K, V]{
		seq: func(yield func(K, V) bool) {
			for k, v := range s.seq {
				select {
				case <-ctx.Done():
					return
				default:
					if !yield(k, v) {
						return
					}
				}
			}
		},
	}
}

// --- Context-Aware Constructors ---

// GenerateCtx creates an infinite Stream using a supplier function with context support.
func GenerateCtx[T any](ctx context.Context, supplier func() T) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					if !yield(supplier()) {
						return
					}
				}
			}
		},
	}
}

// IterateCtx creates an infinite Stream with context support.
func IterateCtx[T any](ctx context.Context, seed T, fn func(T) T) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			current := seed
			for {
				select {
				case <-ctx.Done():
					return
				default:
					if !yield(current) {
						return
					}
					current = fn(current)
				}
			}
		},
	}
}

// RangeCtx creates a Stream of integers [start, end) with context support.
func RangeCtx(ctx context.Context, start, end int) Stream[int] {
	return Stream[int]{
		seq: func(yield func(int) bool) {
			for i := start; i < end; i++ {
				select {
				case <-ctx.Done():
					return
				default:
					if !yield(i) {
						return
					}
				}
			}
		},
	}
}

// FromChannelCtx creates a Stream from a channel with context support.
func FromChannelCtx[T any](ctx context.Context, ch <-chan T) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			for {
				// Prioritize context cancellation if both ctx and channel are ready.
				select {
				case <-ctx.Done():
					return
				default:
				}
				// Then perform a blocking receive that still exits on ctx.
				select {
				case <-ctx.Done():
					return
				case v, ok := <-ch:
					if !ok {
						return
					}
					if !yield(v) {
						return
					}
				}
			}
		},
	}
}

// FromReaderLinesCtx creates a Stream of lines from an io.Reader with context support.
func FromReaderLinesCtx(ctx context.Context, r io.Reader) Stream[string] {
	scanner := bufio.NewScanner(r)
	return Stream[string]{
		seq: func(yield func(string) bool) {
			for scanner.Scan() {
				select {
				case <-ctx.Done():
					return
				default:
					if !yield(scanner.Text()) {
						return
					}
				}
			}
		},
	}
}

// --- Context-Aware Terminal Operations ---

// CollectCtx collects all elements into a slice with context support.
// Returns the elements collected so far and the context error if cancelled.
func CollectCtx[T any](ctx context.Context, s Stream[T]) ([]T, error) {
	var result []T
	for v := range s.seq {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
			result = append(result, v)
		}
	}
	return result, nil
}

// ForEachCtx executes an action on each element with context support.
// The action can be either func(T) or func(T) error.
// If action is func(T) error and returns an error, iteration stops and the error is returned.
// Returns the context error if cancelled, or the first error from action.
//
// Examples:
//
//	err := ForEachCtx(ctx, s, func(v int) { fmt.Println(v) })
//	err := ForEachCtx(ctx, s, func(v int) error { return process(v) })
func ForEachCtx[T any, A ~func(T) | ~func(T) error](ctx context.Context, s Stream[T], action A) error {
	for v := range s.seq {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			switch fn := any(action).(type) {
			case func(T):
				fn(v)
			case func(T) error:
				if err := fn(v); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// ReduceCtx reduces the stream with context support.
// The reducer function can be either func(T, T) T or func(T, T) (T, error).
// If reducer is func(T, T) (T, error) and returns an error, reduction stops and the error is returned.
// Returns the accumulated result and context error if cancelled, or the first error from reducer.
//
// Examples:
//
//	result, err := ReduceCtx(ctx, s, 0, func(a, b int) int { return a + b })
//	result, err := ReduceCtx(ctx, s, 0, func(a, b int) (int, error) { return compute(a, b) })
func ReduceCtx[T any, F ~func(T, T) T | ~func(T, T) (T, error)](ctx context.Context, s Stream[T], identity T, fn F) (T, error) {
	result := identity
	for v := range s.seq {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
			switch reducer := any(fn).(type) {
			case func(T, T) T:
				result = reducer(result, v)
			case func(T, T) (T, error):
				var err error
				result, err = reducer(result, v)
				if err != nil {
					return result, err
				}
			}
		}
	}
	return result, nil
}

// FindFirstCtx finds the first element matching the predicate with context support.
func FindFirstCtx[T any](ctx context.Context, s Stream[T], pred func(T) bool) (Optional[T], error) {
	for v := range s.seq {
		select {
		case <-ctx.Done():
			return None[T](), ctx.Err()
		default:
			if pred(v) {
				return Some(v), nil
			}
		}
	}
	return None[T](), nil
}

// AnyMatchCtx checks if any element matches the predicate with context support.
func AnyMatchCtx[T any](ctx context.Context, s Stream[T], pred func(T) bool) (bool, error) {
	for v := range s.seq {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
			if pred(v) {
				return true, nil
			}
		}
	}
	return false, nil
}

// AllMatchCtx checks if all elements match the predicate with context support.
func AllMatchCtx[T any](ctx context.Context, s Stream[T], pred func(T) bool) (bool, error) {
	for v := range s.seq {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
			if !pred(v) {
				return false, nil
			}
		}
	}
	return true, nil
}

// CountCtx counts elements with context support.
func CountCtx[T any](ctx context.Context, s Stream[T]) (int, error) {
	count := 0
	for range s.seq {
		select {
		case <-ctx.Done():
			return count, ctx.Err()
		default:
			count++
		}
	}
	return count, nil
}

// --- Context-Aware Intermediate Operations ---

// FilterCtx returns a Stream that filters with context support.
func FilterCtx[T any](ctx context.Context, s Stream[T], pred func(T) bool) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			for v := range s.seq {
				select {
				case <-ctx.Done():
					return
				default:
					if pred(v) && !yield(v) {
						return
					}
				}
			}
		},
	}
}

// MapCtx returns a Stream that maps with context support.
func MapCtx[T any](ctx context.Context, s Stream[T], fn func(T) T) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			for v := range s.seq {
				select {
				case <-ctx.Done():
					return
				default:
					if !yield(fn(v)) {
						return
					}
				}
			}
		},
	}
}

// MapToCtx transforms Stream[T] to Stream[U] with context support.
func MapToCtx[T, U any](ctx context.Context, s Stream[T], fn func(T) U) Stream[U] {
	return Stream[U]{
		seq: func(yield func(U) bool) {
			for v := range s.seq {
				select {
				case <-ctx.Done():
					return
				default:
					if !yield(fn(v)) {
						return
					}
				}
			}
		},
	}
}

// --- Error Context ---

// ContextError represents an error that occurred during context-aware operations.
type ContextError struct {
	Err     error
	Partial bool // true if some results were collected before the error
}

func (e *ContextError) Error() string {
	return e.Err.Error()
}

func (e *ContextError) Unwrap() error {
	return e.Err
}
