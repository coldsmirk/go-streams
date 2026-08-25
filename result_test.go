package streams

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResultBasic(t *testing.T) {
	t.Parallel()
	t.Run("Ok", func(t *testing.T) {
		t.Parallel()
		r := Ok(42)
		assert.True(t, r.IsOk(), "Ok result should report IsOk")
		assert.False(t, r.IsErr(), "Ok result should not report IsErr")
		assert.Equal(t, 42, r.Unwrap(), "Ok.Unwrap should return the value")
		assert.Equal(t, 42, r.Value(), "Ok.Value should return the value")
		assert.Nil(t, r.Error(), "Ok result should have nil error")
	})

	t.Run("Err", func(t *testing.T) {
		t.Parallel()
		err := errors.New("test error")
		r := Err[int](err)
		assert.False(t, r.IsOk(), "Err result should not report IsOk")
		assert.True(t, r.IsErr(), "Err result should report IsErr")
		assert.Equal(t, err, r.Error(), "Err should preserve the original error")
		assert.Equal(t, 0, r.Value(), "Err.Value should return the zero value") // zero value for error result
	})

	t.Run("ErrMsg", func(t *testing.T) {
		t.Parallel()
		r := ErrMsg[int]("test error message")
		assert.True(t, r.IsErr(), "ErrMsg should return an Err result")
		assert.Equal(t, "test error message", r.Error().Error(), "ErrMsg should wrap the provided message")
	})

	t.Run("UnwrapOr", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, 42, Ok(42).UnwrapOr(0), "UnwrapOr should return value for Ok")
		assert.Equal(t, 0, Err[int](errors.New("err")).UnwrapOr(0), "UnwrapOr should return default for Err")
	})

	t.Run("UnwrapOrElse", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, 42, Ok(42).UnwrapOrElse(func(_ error) int { return 0 }), "UnwrapOrElse should return value for Ok")
		assert.Equal(t, 99, Err[int](errors.New("err")).UnwrapOrElse(func(_ error) int { return 99 }), "UnwrapOrElse should call fallback for Err")
	})

	t.Run("UnwrapPanics", func(t *testing.T) {
		t.Parallel()
		assert.Panics(t, func() {
			Err[int](errors.New("err")).Unwrap()
		}, "Unwrap should panic on Err")
	})

	t.Run("UnwrapErr", func(t *testing.T) {
		t.Parallel()
		err := errors.New("test error")
		assert.Equal(t, err, Err[int](err).UnwrapErr(), "UnwrapErr should return the error")
	})

	t.Run("UnwrapErrPanics", func(t *testing.T) {
		t.Parallel()
		assert.Panics(t, func() {
			_ = Ok(42).UnwrapErr()
		}, "UnwrapErr should panic on Ok")
	})

	t.Run("Get", func(t *testing.T) {
		t.Parallel()
		v, err := Ok(42).Get()
		assert.Equal(t, 42, v, "Get on Ok should return value")
		assert.NoError(t, err, "Get on Ok should not error")

		v, err = Err[int](errors.New("test")).Get()
		assert.Equal(t, 0, v, "Get on Err should return zero value")
		assert.Error(t, err, "Get on Err should return error")
	})

	t.Run("ToOptional", func(t *testing.T) {
		t.Parallel()
		assert.True(t, Ok(42).ToOptional().IsPresent(), "Ok.ToOptional should return Some")
		assert.False(t, Err[int](errors.New("err")).ToOptional().IsPresent(), "Err.ToOptional should return None")
	})

	t.Run("Map", func(t *testing.T) {
		t.Parallel()
		okResult := Ok(5).Map(func(n int) int { return n * 2 })
		assert.True(t, okResult.IsOk(), "Map should return Ok for Ok input")
		assert.Equal(t, 10, okResult.Value(), "Map should transform the value")

		errResult := Err[int](errors.New("err")).Map(func(n int) int { return n * 2 })
		assert.True(t, errResult.IsErr(), "Map should keep Err without calling mapper")
	})

	t.Run("MapErr", func(t *testing.T) {
		t.Parallel()
		okResult := Ok(5).MapErr(func(_ error) error { return errors.New("wrapped") })
		assert.True(t, okResult.IsOk(), "MapErr should keep Ok result")
		assert.Equal(t, 5, okResult.Value(), "MapErr should keep the Ok value")

		errResult := Err[int](errors.New("original")).MapErr(func(e error) error {
			return errors.New("wrapped: " + e.Error())
		})
		assert.True(t, errResult.IsErr(), "MapErr should return Err when input is Err")
		assert.Equal(t, "wrapped: original", errResult.Error().Error(), "MapErr should transform the error")
	})

	t.Run("And", func(t *testing.T) {
		t.Parallel()
		ok1 := Ok(1)
		ok2 := Ok(2)
		err1 := Err[int](errors.New("err1"))

		assert.Equal(t, 2, ok1.And(ok2).Value(), "And should return second Ok when first is Ok")
		assert.True(t, err1.And(ok2).IsErr(), "And should short-circuit on first Err")
		assert.True(t, ok1.And(err1).IsErr(), "And should return Err when second is Err")
	})

	t.Run("Or", func(t *testing.T) {
		t.Parallel()
		ok1 := Ok(1)
		ok2 := Ok(2)
		err1 := Err[int](errors.New("err1"))

		assert.Equal(t, 1, ok1.Or(ok2).Value(), "Or should return first Ok")
		assert.Equal(t, 2, err1.Or(ok2).Value(), "Or should return second Ok when first is Err")
		assert.True(t, err1.Or(err1).IsErr(), "Or should return Err when both are Err")
	})
}

func TestResultTransformations(t *testing.T) {
	t.Parallel()
	t.Run("MapResultTo", func(t *testing.T) {
		t.Parallel()
		okResult := MapResultTo(Ok(5), func(n int) string {
			return strings.Repeat("x", n)
		})
		assert.True(t, okResult.IsOk(), "MapResultTo should return Ok for Ok input")
		assert.Equal(t, "xxxxx", okResult.Value(), "MapResultTo should map value to string")

		errResult := MapResultTo(Err[int](errors.New("err")), func(n int) string {
			return strings.Repeat("x", n)
		})
		assert.True(t, errResult.IsErr(), "MapResultTo should return Err for Err input")
	})

	t.Run("FlatMapResult", func(t *testing.T) {
		t.Parallel()
		okResult := FlatMapResult(Ok(5), func(n int) Result[string] {
			if n > 0 {
				return Ok(strings.Repeat("x", n))
			}
			return Err[string](errors.New("negative"))
		})
		assert.True(t, okResult.IsOk(), "FlatMapResult should return Ok for Ok input")
		assert.Equal(t, "xxxxx", okResult.Value(), "FlatMapResult should map the Ok value")

		errInner := FlatMapResult(Ok(-1), func(n int) Result[string] {
			if n > 0 {
				return Ok(strings.Repeat("x", n))
			}
			return Err[string](errors.New("negative"))
		})
		assert.True(t, errInner.IsErr(), "FlatMapResult should return Err when mapper returns Err")

		errOuter := FlatMapResult(Err[int](errors.New("outer")), func(_ int) Result[string] {
			return Ok("never called")
		})
		assert.True(t, errOuter.IsErr(), "FlatMapResult should return Err when input is Err")
	})
}

func TestResultStreamOperations(t *testing.T) {
	t.Parallel()
	t.Run("MapErrTo", func(t *testing.T) {
		t.Parallel()
		results := MapErrTo(Of(1, 2, 3), func(n int) (string, error) {
			if n == 2 {
				return "", errors.New("error on 2")
			}
			return strings.Repeat("x", n), nil
		}).Collect()

		assert.Len(t, results, 3, "MapErrTo should yield one result per input")
		assert.True(t, results[0].IsOk(), "First should be Ok")
		assert.Equal(t, "x", results[0].Value(), "First value should be 'x'")
		assert.True(t, results[1].IsErr(), "Second should be Err")
		assert.True(t, results[2].IsOk(), "Third should be Ok")
		assert.Equal(t, "xxx", results[2].Value(), "Third value should be 'xxx'")
	})

	t.Run("MapErrToEmpty", func(t *testing.T) {
		t.Parallel()
		results := MapErrTo(Empty[int](), func(n int) (string, error) {
			return strings.Repeat("x", n), nil
		}).Collect()
		assert.Empty(t, results, "MapErrTo on empty stream should be empty")
	})

	t.Run("MapErrToEarlyTermination", func(t *testing.T) {
		t.Parallel()
		results := MapErrTo(Of(1, 2, 3, 4, 5), func(n int) (string, error) {
			return strings.Repeat("x", n), nil
		}).Limit(2).Collect()
		assert.Len(t, results, 2, "MapErrTo with Limit should stop early")
	})

	t.Run("MapErrToEarlyTerminationOnError", func(t *testing.T) {
		t.Parallel()
		// Test early termination when yield returns false during error path
		results := MapErrTo(Of(1, 2, 3, 4, 5), func(_ int) (string, error) {
			// Return error for all elements to test error path termination
			return "", errors.New("error")
		}).Limit(1).Collect()
		assert.Len(t, results, 1, "MapErrTo should stop early on error path")
		assert.True(t, results[0].IsErr(), "MapErrTo should return Err on error path")
	})

	t.Run("FilterErr", func(t *testing.T) {
		t.Parallel()
		// FilterErr only yields elements that pass the predicate (or have errors)
		// Elements where predicate returns false are NOT yielded
		results := FilterErr(Of(1, 2, 3, 4, 5), func(n int) (bool, error) {
			if n == 3 {
				return false, errors.New("error on 3")
			}
			return n%2 == 0, nil
		}).Collect()

		// Input: 1(odd), 2(even), 3(error), 4(even), 5(odd)
		// Output: Ok(2), Err, Ok(4) = 3 results
		assert.Len(t, results, 3, "FilterErr should yield matching Ok values and Errs")
		assert.True(t, results[0].IsOk(), "FilterErr should keep first matching Ok")
		assert.Equal(t, 2, results[0].Value(), "FilterErr should yield value 2 first")
		assert.True(t, results[1].IsErr(), "FilterErr should yield Err for error case")
		assert.True(t, results[2].IsOk(), "FilterErr should keep later matching Ok")
		assert.Equal(t, 4, results[2].Value(), "FilterErr should yield value 4 last")
	})

	t.Run("FilterErrEarlyTerminationOnError", func(t *testing.T) {
		t.Parallel()
		// Test early termination when yield returns false during error path
		results := FilterErr(Of(1, 2, 3, 4, 5), func(_ int) (bool, error) {
			return false, errors.New("error")
		}).Limit(1).Collect()
		assert.Len(t, results, 1, "FilterErr should stop early on error path")
		assert.True(t, results[0].IsErr(), "FilterErr should yield Err on error path")
	})

	t.Run("FilterErrEarlyTerminationOnOk", func(t *testing.T) {
		t.Parallel()
		// Test early termination when yield returns false during ok path
		results := FilterErr(Of(1, 2, 3, 4, 5), func(_ int) (bool, error) {
			return true, nil // All pass
		}).Limit(2).Collect()
		assert.Len(t, results, 2, "FilterErr should stop early on ok path")
		assert.True(t, results[0].IsOk(), "FilterErr should yield Ok results when predicate passes")
		assert.True(t, results[1].IsOk(), "FilterErr should yield Ok results when predicate passes")
	})

	t.Run("FlatMapErr", func(t *testing.T) {
		t.Parallel()
		results := FlatMapErr(Of(1, 2, 3), func(n int) (Stream[int], error) {
			if n == 2 {
				return Empty[int](), errors.New("error on 2")
			}
			return Of(n*10, n*10+1), nil
		}).Collect()

		// Input 1 produces [10, 11], input 2 produces error, input 3 produces [30, 31]
		// So: Ok(10), Ok(11), Err, Ok(30), Ok(31) = 5 results
		assert.Len(t, results, 5, "FlatMapErr should yield all Ok values plus Err")
		assert.True(t, results[0].IsOk(), "FlatMapErr should yield Ok values from first input")
		assert.Equal(t, 10, results[0].Value(), "FlatMapErr should yield 10 from first input")
		assert.True(t, results[1].IsOk(), "FlatMapErr should yield Ok values from first input")
		assert.Equal(t, 11, results[1].Value(), "FlatMapErr should yield 11 from first input")
		assert.True(t, results[2].IsErr(), "FlatMapErr should yield Err for error input")
		assert.True(t, results[3].IsOk(), "FlatMapErr should yield Ok values from later input")
		assert.Equal(t, 30, results[3].Value(), "FlatMapErr should yield 30 from third input")
	})

	t.Run("FlatMapErrEarlyTerminationOnError", func(t *testing.T) {
		t.Parallel()
		// Test early termination when yield returns false during error path
		results := FlatMapErr(Of(1, 2, 3), func(_ int) (Stream[int], error) {
			return Empty[int](), errors.New("error")
		}).Limit(1).Collect()
		assert.Len(t, results, 1, "FlatMapErr should stop early on error path")
		assert.True(t, results[0].IsErr(), "FlatMapErr should yield Err on error path")
	})

	t.Run("FlatMapErrEarlyTerminationInInnerStream", func(t *testing.T) {
		t.Parallel()
		// Test early termination when yield returns false during inner stream iteration
		results := FlatMapErr(Of(1, 2, 3), func(n int) (Stream[int], error) {
			return Of(n*10, n*10+1, n*10+2), nil // Each produces 3 elements
		}).Limit(2).Collect()
		assert.Len(t, results, 2, "FlatMapErr should stop early in inner stream")
		assert.True(t, results[0].IsOk(), "FlatMapErr should yield Ok results before limit")
		assert.True(t, results[1].IsOk(), "FlatMapErr should yield Ok results before limit")
	})
}

func TestResultCollectors(t *testing.T) {
	t.Parallel()
	t.Run("CollectResultsAllOk", func(t *testing.T) {
		t.Parallel()
		vals, err := CollectResults(FromResults(Ok(1), Ok(2), Ok(3)))
		assert.NoError(t, err, "CollectResultsAllOk should not return error")
		assert.Equal(t, []int{1, 2, 3}, vals, "CollectResultsAllOk should return all values")
	})

	t.Run("CollectResultsWithError", func(t *testing.T) {
		t.Parallel()
		vals, err := CollectResults(FromResults(Ok(1), Ok(2), Err[int](errors.New("err")), Ok(4)))
		assert.Equal(t, []int{1, 2}, vals, "CollectResultsWithError should collect Ok values before error")
		assert.Error(t, err, "CollectResultsWithError should return error")
	})

	t.Run("CollectResultsErrorAtStart", func(t *testing.T) {
		t.Parallel()
		vals, err := CollectResults(FromResults(Err[int](errors.New("err")), Ok(2)))
		assert.Empty(t, vals, "CollectResultsErrorAtStart should return no values")
		assert.Error(t, err, "CollectResultsErrorAtStart should return error")
	})

	t.Run("CollectResultsAll", func(t *testing.T) {
		t.Parallel()
		vals, errs := CollectResultsAll(FromResults(
			Ok(1), Err[int](errors.New("err1")), Ok(3), Err[int](errors.New("err2")),
		))
		assert.Equal(t, []int{1, 3}, vals, "CollectResultsAll should contain only Ok values")
		assert.Len(t, errs, 2, "CollectResultsAll should contain all errors")
	})

	t.Run("PartitionResults", func(t *testing.T) {
		t.Parallel()
		oks, errs := PartitionResults(FromResults(Ok(1), Err[int](errors.New("err")), Ok(3)))
		assert.Equal(t, []int{1, 3}, oks, "PartitionResults should split Ok values")
		assert.Len(t, errs, 1, "PartitionResults should split Err values")
	})

	t.Run("FilterOk", func(t *testing.T) {
		t.Parallel()
		results := FromResults(Ok(1), Err[int](errors.New("err")), Ok(3))
		vals := FilterOk(results).Collect()
		assert.Equal(t, []int{1, 3}, vals, "FilterOk should keep only Ok values")
	})

	t.Run("FilterOkEarlyTermination", func(t *testing.T) {
		t.Parallel()
		results := FromResults(Ok(1), Ok(2), Ok(3), Ok(4), Ok(5))
		vals := FilterOk(results).Limit(2).Collect()
		assert.Len(t, vals, 2, "FilterOk should stop early when Limit is applied")
		assert.Equal(t, []int{1, 2}, vals, "FilterOk should preserve order under Limit")
	})

	t.Run("FilterErrs", func(t *testing.T) {
		t.Parallel()
		results := FromResults(Ok(1), Err[int](errors.New("err1")), Ok(3), Err[int](errors.New("err2")))
		errs := FilterErrs(results).Collect()
		assert.Len(t, errs, 2, "FilterErrs should collect only errors")
		assert.Equal(t, "err1", errs[0].Error(), "First error should match")
		assert.Equal(t, "err2", errs[1].Error(), "Second error should match")
	})

	t.Run("FilterErrsEarlyTermination", func(t *testing.T) {
		t.Parallel()
		results := FromResults(Err[int](errors.New("err1")), Err[int](errors.New("err2")), Err[int](errors.New("err3")))
		errs := FilterErrs(results).Limit(2).Collect()
		assert.Len(t, errs, 2, "FilterErrs should stop early when Limit is applied")
	})

	t.Run("UnwrapResults", func(t *testing.T) {
		t.Parallel()
		vals := UnwrapResults(FromResults(Ok(1), Ok(2), Ok(3))).Collect()
		assert.Equal(t, []int{1, 2, 3}, vals, "UnwrapResults should extract Ok values")
	})

	t.Run("UnwrapResultsEarlyTermination", func(t *testing.T) {
		t.Parallel()
		vals := UnwrapResults(FromResults(Ok(1), Ok(2), Ok(3), Ok(4), Ok(5))).Limit(2).Collect()
		assert.Len(t, vals, 2, "UnwrapResults should stop early when Limit is applied")
		assert.Equal(t, []int{1, 2}, vals, "UnwrapResults should preserve order under Limit")
	})

	t.Run("UnwrapResultsEmpty", func(t *testing.T) {
		t.Parallel()
		vals := UnwrapResults(Empty[Result[int]]()).Collect()
		assert.Empty(t, vals, "UnwrapResults on empty stream should be empty")
	})

	t.Run("UnwrapResultsPanics", func(t *testing.T) {
		t.Parallel()
		assert.Panics(t, func() {
			UnwrapResults(FromResults(Ok(1), Err[int](errors.New("err")))).Collect()
		}, "UnwrapResults should panic on Err")
	})

	t.Run("CollectResultsEmpty", func(t *testing.T) {
		t.Parallel()
		vals, err := CollectResults(Empty[Result[int]]())
		assert.NoError(t, err, "CollectResults on empty stream should not error")
		assert.Empty(t, vals, "CollectResults on empty stream should return empty slice")
	})

	t.Run("CollectResultsAllEmpty", func(t *testing.T) {
		t.Parallel()
		vals, errs := CollectResultsAll(Empty[Result[int]]())
		assert.Empty(t, vals, "CollectResultsAll on empty stream should return empty values")
		assert.Empty(t, errs, "CollectResultsAll on empty stream should return empty errors")
	})

	t.Run("PartitionResultsEmpty", func(t *testing.T) {
		t.Parallel()
		oks, errs := PartitionResults(Empty[Result[int]]())
		assert.Empty(t, oks, "PartitionResults on empty stream should return empty oks")
		assert.Empty(t, errs, "PartitionResults on empty stream should return empty errs")
	})

	t.Run("UnwrapOrDefault", func(t *testing.T) {
		t.Parallel()
		vals := UnwrapOrDefault(FromResults(Ok(1), Err[int](errors.New("err")), Ok(3)), 0).Collect()
		assert.Equal(t, []int{1, 0, 3}, vals, "UnwrapOrDefault should replace Err with default")
	})

	t.Run("UnwrapOrDefaultEarlyTermination", func(t *testing.T) {
		t.Parallel()
		vals := UnwrapOrDefault(FromResults(Ok(1), Ok(2), Ok(3), Ok(4), Ok(5)), 0).Limit(2).Collect()
		assert.Len(t, vals, 2, "UnwrapOrDefault should stop early when Limit is applied")
		assert.Equal(t, []int{1, 2}, vals, "UnwrapOrDefault should preserve order under Limit")
	})

	t.Run("TakeUntilErr", func(t *testing.T) {
		t.Parallel()
		vals := TakeUntilErr(FromResults(Ok(1), Ok(2), Err[int](errors.New("err")), Ok(4))).Collect()
		assert.Equal(t, []int{1, 2}, vals, "TakeUntilErr should stop before first error")
	})

	t.Run("TakeUntilErrEarlyTermination", func(t *testing.T) {
		t.Parallel()
		vals := TakeUntilErr(FromResults(Ok(1), Ok(2), Ok(3), Ok(4), Ok(5))).Limit(2).Collect()
		assert.Len(t, vals, 2, "TakeUntilErr should stop early when Limit is applied")
		assert.Equal(t, []int{1, 2}, vals, "TakeUntilErr should preserve order under Limit")
	})

	t.Run("FromResults", func(t *testing.T) {
		t.Parallel()
		results := FromResults(Ok(1), Ok(2), Err[int](errors.New("err"))).Collect()
		assert.Len(t, results, 3, "FromResults should collect all inputs")
		assert.True(t, results[0].IsOk(), "First should be Ok")
		assert.True(t, results[1].IsOk(), "Second should be Ok")
		assert.True(t, results[2].IsErr(), "Third should be Err")
	})

	t.Run("TryCollectSuccess", func(t *testing.T) {
		t.Parallel()
		result := TryCollect(Of(1, 2, 3))
		assert.True(t, result.IsOk(), "TryCollect success should return Ok")
		assert.Equal(t, []int{1, 2, 3}, result.Value(), "TryCollect should return collected values")
	})

	t.Run("TryCollectPanic", func(t *testing.T) {
		t.Parallel()
		panicStream := Stream[int]{
			seq: func(yield func(int) bool) {
				yield(1)
				panic("test panic")
			},
		}
		result := TryCollect(panicStream)
		assert.True(t, result.IsErr(), "TryCollect panic should return Err")
		assert.Contains(t, result.Error().Error(), "test panic", "Error should contain panic message")
	})

	t.Run("TryCollectPanicWithError", func(t *testing.T) {
		t.Parallel()
		panicStream := Stream[int]{
			seq: func(yield func(int) bool) {
				yield(1)
				panic(errors.New("test error panic"))
			},
		}
		result := TryCollect(panicStream)
		assert.True(t, result.IsErr(), "TryCollect panic(error) should return Err")
		assert.Equal(t, "test error panic", result.Error().Error(), "Error should match panic error")
	})
}
