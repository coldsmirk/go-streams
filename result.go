package streams

import (
	"errors"
	"fmt"
)

// Result represents a value that may be either a success (Ok) or a failure (Err).
// It's useful for error propagation in stream pipelines.
type Result[T any] struct {
	value T
	err   error
}

// Ok creates a successful Result containing the given value.
func Ok[T any](value T) Result[T] {
	return Result[T]{value: value}
}

// Err creates a failed Result containing the given error.
func Err[T any](err error) Result[T] {
	return Result[T]{err: err}
}

// ErrMsg creates a failed Result with an error message.
func ErrMsg[T any](msg string) Result[T] {
	return Result[T]{err: errors.New(msg)}
}

// IsOk returns true if the Result is successful.
func (r Result[T]) IsOk() bool {
	return r.err == nil
}

// IsErr returns true if the Result is a failure.
func (r Result[T]) IsErr() bool {
	return r.err != nil
}

// Unwrap returns the value if Ok, or panics if Err.
func (r Result[T]) Unwrap() T {
	if r.err != nil {
		panic(fmt.Sprintf("called Unwrap on Err: %v", r.err))
	}
	return r.value
}

// UnwrapOr returns the value if Ok, or the default value if Err.
func (r Result[T]) UnwrapOr(defaultVal T) T {
	if r.err != nil {
		return defaultVal
	}
	return r.value
}

// UnwrapOrElse returns the value if Ok, or calls the function if Err.
func (r Result[T]) UnwrapOrElse(fn func(error) T) T {
	if r.err != nil {
		return fn(r.err)
	}
	return r.value
}

// UnwrapErr returns the error if Err, or panics if Ok.
func (r Result[T]) UnwrapErr() error {
	if r.err == nil {
		panic("called UnwrapErr on Ok")
	}
	return r.err
}

// Error returns the error (or nil if Ok).
func (r Result[T]) Error() error {
	return r.err
}

// Value returns the value (zero value if Err).
func (r Result[T]) Value() T {
	return r.value
}

// Get returns both value and error.
func (r Result[T]) Get() (T, error) {
	return r.value, r.err
}

// ToOptional converts Result to Optional, discarding the error.
func (r Result[T]) ToOptional() Optional[T] {
	if r.err != nil {
		return None[T]()
	}
	return Some(r.value)
}

// Map transforms the value if Ok, passes through Err unchanged.
func (r Result[T]) Map(fn func(T) T) Result[T] {
	if r.err != nil {
		return r
	}
	return Ok(fn(r.value))
}

// MapErr transforms the error if Err, passes through Ok unchanged.
func (r Result[T]) MapErr(fn func(error) error) Result[T] {
	if r.err == nil {
		return r
	}
	return Err[T](fn(r.err))
}

// And returns the other Result if this is Ok, otherwise returns this Err.
func (r Result[T]) And(other Result[T]) Result[T] {
	if r.err != nil {
		return r
	}
	return other
}

// Or returns this Result if Ok, otherwise returns the other Result.
func (r Result[T]) Or(other Result[T]) Result[T] {
	if r.err == nil {
		return r
	}
	return other
}

// --- Result Stream Operations ---

// MapResultTo transforms Result[T] to Result[U] using the given function.
func MapResultTo[T, U any](r Result[T], fn func(T) U) Result[U] {
	if r.err != nil {
		return Err[U](r.err)
	}
	return Ok(fn(r.value))
}

// FlatMapResult transforms Result[T] to Result[U], allowing the function to fail.
func FlatMapResult[T, U any](r Result[T], fn func(T) Result[U]) Result[U] {
	if r.err != nil {
		return Err[U](r.err)
	}
	return fn(r.value)
}

// --- Error-Aware Stream Operations ---

// MapErrTo transforms each element using a function that may return an error.
// The resulting stream contains Result values.
func MapErrTo[T, U any](s Stream[T], fn func(T) (U, error)) Stream[Result[U]] {
	return Stream[Result[U]]{
		seq: func(yield func(Result[U]) bool) {
			for v := range s.seq {
				result, err := fn(v)
				if err != nil {
					if !yield(Err[U](err)) {
						return
					}
				} else {
					if !yield(Ok(result)) {
						return
					}
				}
			}
		},
	}
}

// FilterErr filters elements using a predicate that may return an error.
// Elements that pass the predicate are wrapped in Ok, errors are wrapped in Err.
func FilterErr[T any](s Stream[T], pred func(T) (bool, error)) Stream[Result[T]] {
	return Stream[Result[T]]{
		seq: func(yield func(Result[T]) bool) {
			for v := range s.seq {
				ok, err := pred(v)
				if err != nil {
					if !yield(Err[T](err)) {
						return
					}
				} else if ok {
					if !yield(Ok(v)) {
						return
					}
				}
			}
		},
	}
}

// FlatMapErr maps each element to a stream using a function that may return an error.
func FlatMapErr[T, U any](s Stream[T], fn func(T) (Stream[U], error)) Stream[Result[U]] {
	return Stream[Result[U]]{
		seq: func(yield func(Result[U]) bool) {
			for v := range s.seq {
				inner, err := fn(v)
				if err != nil {
					if !yield(Err[U](err)) {
						return
					}
					continue
				}
				for u := range inner.seq {
					if !yield(Ok(u)) {
						return
					}
				}
			}
		},
	}
}

// --- Result Stream Collectors ---

// CollectResults collects a stream of Results into a slice and error.
// Returns the first error encountered, or nil if all succeeded.
func CollectResults[T any](s Stream[Result[T]]) ([]T, error) {
	var results []T
	for r := range s.seq {
		if r.IsErr() {
			return results, r.err
		}
		results = append(results, r.value)
	}
	return results, nil
}

// CollectResultsAll collects all Results, continuing even after errors.
// Returns all successful values and all errors encountered.
func CollectResultsAll[T any](s Stream[Result[T]]) ([]T, []error) {
	var results []T
	var errs []error
	for r := range s.seq {
		if r.IsErr() {
			errs = append(errs, r.err)
		} else {
			results = append(results, r.value)
		}
	}
	return results, errs
}

// PartitionResults separates a stream of Results into successes and failures.
func PartitionResults[T any](s Stream[Result[T]]) ([]T, []error) {
	return CollectResultsAll(s)
}

// FilterOk filters a stream of Results to only include successful values.
func FilterOk[T any](s Stream[Result[T]]) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			for r := range s.seq {
				if r.IsOk() {
					if !yield(r.value) {
						return
					}
				}
			}
		},
	}
}

// FilterErrs filters a stream of Results to only include errors.
func FilterErrs[T any](s Stream[Result[T]]) Stream[error] {
	return Stream[error]{
		seq: func(yield func(error) bool) {
			for r := range s.seq {
				if r.IsErr() {
					if !yield(r.err) {
						return
					}
				}
			}
		},
	}
}

// UnwrapResults unwraps all Results, panicking on the first error.
func UnwrapResults[T any](s Stream[Result[T]]) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			for r := range s.seq {
				if !yield(r.Unwrap()) {
					return
				}
			}
		},
	}
}

// UnwrapOrDefault unwraps Results, using a default value for errors.
func UnwrapOrDefault[T any](s Stream[Result[T]], defaultVal T) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			for r := range s.seq {
				if !yield(r.UnwrapOr(defaultVal)) {
					return
				}
			}
		},
	}
}

// TakeUntilErr takes elements until the first error is encountered.
// The error is not yielded; use CollectResults if you need the error.
func TakeUntilErr[T any](s Stream[Result[T]]) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			for r := range s.seq {
				if r.IsErr() {
					return
				}
				if !yield(r.value) {
					return
				}
			}
		},
	}
}

// FromResults creates a Stream of Results from variadic Results.
func FromResults[T any](results ...Result[T]) Stream[Result[T]] {
	return FromSlice(results)
}

// TryCollect attempts to collect a stream, wrapping any panic as an error.
// This is useful when the stream's source might panic.
func TryCollect[T any](s Stream[T]) Result[[]T] {
	var results []T
	var panicErr error

	func() {
		defer func() {
			if r := recover(); r != nil {
				if err, ok := r.(error); ok {
					panicErr = err
				} else {
					panicErr = fmt.Errorf("panic: %v", r)
				}
			}
		}()
		results = s.Collect()
	}()

	if panicErr != nil {
		return Err[[]T](panicErr)
	}
	return Ok(results)
}
