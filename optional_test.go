package streams

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestOptional tests Optional operations.
func TestOptional(t *testing.T) {
	t.Parallel()
	t.Run("Some", func(t *testing.T) {
		t.Parallel()
		opt := Some(42)
		assert.True(t, opt.IsPresent(), "Some should be present")
		assert.False(t, opt.IsEmpty(), "Some should not be empty")
		assert.Equal(t, 42, opt.Get(), "Some should contain value")
	})

	t.Run("None", func(t *testing.T) {
		t.Parallel()
		opt := None[int]()
		assert.False(t, opt.IsPresent(), "None should not be present")
		assert.True(t, opt.IsEmpty(), "None should be empty")
	})

	t.Run("GetPanicsOnNone", func(t *testing.T) {
		t.Parallel()
		opt := None[int]()
		assert.Panics(t, func() { opt.Get() }, "Get on None should panic")
	})

	t.Run("GetOrElse", func(t *testing.T) {
		t.Parallel()
		some := Some(42)
		assert.Equal(t, 42, some.GetOrElse(0), "GetOrElse on Some should return value")

		none := None[int]()
		assert.Equal(t, 100, none.GetOrElse(100), "GetOrElse on None should return default")
	})

	t.Run("GetOrElseGet", func(t *testing.T) {
		t.Parallel()
		some := Some(42)
		assert.Equal(t, 42, some.GetOrElseGet(func() int { return 0 }), "GetOrElseGet on Some should return value")

		callCount := 0
		none := None[int]()
		result := none.GetOrElseGet(func() int {
			callCount++
			return 100
		})
		assert.Equal(t, 100, result, "GetOrElseGet on None should call supplier")
		assert.Equal(t, 1, callCount, "GetOrElseGet should call supplier only once")
	})

	t.Run("GetOrZero", func(t *testing.T) {
		t.Parallel()
		some := Some(42)
		assert.Equal(t, 42, some.GetOrZero(), "GetOrZero on Some should return value")

		none := None[int]()
		assert.Equal(t, 0, none.GetOrZero(), "GetOrZero on None should return zero value")
	})

	t.Run("IfPresent", func(t *testing.T) {
		t.Parallel()
		var called bool
		Some(42).IfPresent(func(_ int) { called = true })
		assert.True(t, called, "IfPresent should call action on Some")

		called = false
		None[int]().IfPresent(func(_ int) { called = true })
		assert.False(t, called, "IfPresent should not call action on None")
	})

	t.Run("IfPresentOrElse", func(t *testing.T) {
		t.Parallel()
		var presentCalled, emptyCalled bool

		Some(42).IfPresentOrElse(
			func(_ int) { presentCalled = true },
			func() { emptyCalled = true },
		)
		assert.True(t, presentCalled, "IfPresentOrElse should call present action on Some")
		assert.False(t, emptyCalled, "IfPresentOrElse should not call empty action on Some")

		presentCalled = false
		emptyCalled = false
		None[int]().IfPresentOrElse(
			func(_ int) { presentCalled = true },
			func() { emptyCalled = true },
		)
		assert.False(t, presentCalled, "IfPresentOrElse should not call present action on None")
		assert.True(t, emptyCalled, "IfPresentOrElse should call empty action on None")
	})

	t.Run("Filter", func(t *testing.T) {
		t.Parallel()
		some := Some(42)
		filtered := some.Filter(func(v int) bool { return v > 40 })
		assert.True(t, filtered.IsPresent(), "Filter should keep matching value")

		notFiltered := some.Filter(func(v int) bool { return v > 50 })
		assert.True(t, notFiltered.IsEmpty(), "Filter should remove non-matching value")

		none := None[int]()
		noneFiltered := none.Filter(func(_ int) bool { return true })
		assert.True(t, noneFiltered.IsEmpty(), "Filter on None should remain None")
	})

	t.Run("Map", func(t *testing.T) {
		t.Parallel()
		some := Some(42)
		mapped := some.Map(func(v int) int { return v * 2 })
		assert.True(t, mapped.IsPresent(), "Map should preserve presence")
		assert.Equal(t, 84, mapped.Get(), "Map should transform value")

		none := None[int]()
		noneMapped := none.Map(func(v int) int { return v * 2 })
		assert.True(t, noneMapped.IsEmpty(), "Map on None should remain None")
	})

	t.Run("OrElse", func(t *testing.T) {
		t.Parallel()
		some := Some(42)
		result := some.OrElse(Some(100))
		assert.Equal(t, 42, result.Get(), "OrElse should return first if present")

		none := None[int]()
		resultNone := none.OrElse(Some(100))
		assert.Equal(t, 100, resultNone.Get(), "OrElse should return other if first is None")
	})

	t.Run("OrElseGet", func(t *testing.T) {
		t.Parallel()
		callCount := 0
		supplier := func() Optional[int] {
			callCount++
			return Some(100)
		}

		some := Some(42)
		result := some.OrElseGet(supplier)
		assert.Equal(t, 42, result.Get(), "OrElseGet should return first if present")
		assert.Equal(t, 0, callCount, "OrElseGet should not call supplier if present")

		none := None[int]()
		resultNone := none.OrElseGet(supplier)
		assert.Equal(t, 100, resultNone.Get(), "OrElseGet should call supplier if None")
		assert.Equal(t, 1, callCount, "OrElseGet should call supplier once")
	})

	t.Run("ToSlice", func(t *testing.T) {
		t.Parallel()
		some := Some(42)
		assert.Equal(t, []int{42}, some.ToSlice(), "ToSlice on Some should return single-element slice")

		none := None[int]()
		assert.Nil(t, none.ToSlice(), "ToSlice on None should return nil")
	})

	t.Run("ToPointer", func(t *testing.T) {
		t.Parallel()
		some := Some(42)
		ptr := some.ToPointer()
		assert.NotNil(t, ptr, "ToPointer on Some should return non-nil pointer")
		assert.Equal(t, 42, *ptr, "ToPointer should point to value")

		none := None[int]()
		assert.Nil(t, none.ToPointer(), "ToPointer on None should return nil")
	})

	t.Run("ToStream", func(t *testing.T) {
		t.Parallel()
		some := Some(42)
		result := some.ToStream().Collect()
		assert.Equal(t, []int{42}, result, "ToStream on Some should return single-element stream")

		none := None[int]()
		resultNone := none.ToStream().Collect()
		assert.Empty(t, resultNone, "ToStream on None should return empty stream")
	})

	t.Run("String", func(t *testing.T) {
		t.Parallel()
		some := Some(42)
		assert.Equal(t, "Some(...)", some.String(), "String on Some should return Some(...)")

		none := None[int]()
		assert.Equal(t, "None", none.String(), "String on None should return None")
	})
}

// TestOptionalOf tests OptionalOf constructor.
func TestOptionalOf(t *testing.T) {
	t.Parallel()
	t.Run("FromNonNilPointer", func(t *testing.T) {
		t.Parallel()
		val := 42
		opt := OptionalOf(&val)
		assert.True(t, opt.IsPresent(), "OptionalOf non-nil should be present")
		assert.Equal(t, 42, opt.Get(), "OptionalOf should contain pointed value")
	})

	t.Run("FromNilPointer", func(t *testing.T) {
		t.Parallel()
		var ptr *int
		opt := OptionalOf(ptr)
		assert.True(t, opt.IsEmpty(), "OptionalOf nil should be empty")
	})
}

// TestOptionalFromCondition tests OptionalFromCondition constructor.
func TestOptionalFromCondition(t *testing.T) {
	t.Parallel()
	t.Run("ConditionTrue", func(t *testing.T) {
		t.Parallel()
		opt := OptionalFromCondition(true, 42)
		assert.True(t, opt.IsPresent(), "OptionalFromCondition(true) should be present")
		assert.Equal(t, 42, opt.Get(), "OptionalFromCondition should contain value")
	})

	t.Run("ConditionFalse", func(t *testing.T) {
		t.Parallel()
		opt := OptionalFromCondition(false, 42)
		assert.True(t, opt.IsEmpty(), "OptionalFromCondition(false) should be empty")
	})
}

// TestOptionalMap tests OptionalMap function.
func TestOptionalMap(t *testing.T) {
	t.Parallel()
	t.Run("MapToString", func(t *testing.T) {
		t.Parallel()
		some := Some(42)
		mapped := OptionalMap(some, func(_ int) string {
			return "value"
		})
		assert.True(t, mapped.IsPresent(), "OptionalMap should preserve presence")
		assert.Equal(t, "value", mapped.Get(), "OptionalMap should change type")
	})

	t.Run("MapNone", func(t *testing.T) {
		t.Parallel()
		none := None[int]()
		mapped := OptionalMap(none, func(_ int) string {
			return "value"
		})
		assert.True(t, mapped.IsEmpty(), "OptionalMap on None should remain None")
	})
}

// TestOptionalFlatMap tests OptionalFlatMap function.
func TestOptionalFlatMap(t *testing.T) {
	t.Parallel()
	t.Run("FlatMapSome", func(t *testing.T) {
		t.Parallel()
		some := Some(42)
		result := OptionalFlatMap(some, func(_ int) Optional[string] {
			return Some("value")
		})
		assert.True(t, result.IsPresent(), "OptionalFlatMap Some->Some should be present")
		assert.Equal(t, "value", result.Get(), "OptionalFlatMap should transform")
	})

	t.Run("FlatMapSomeToNone", func(t *testing.T) {
		t.Parallel()
		some := Some(42)
		result := OptionalFlatMap(some, func(_ int) Optional[string] {
			return None[string]()
		})
		assert.True(t, result.IsEmpty(), "OptionalFlatMap Some->None should be empty")
	})

	t.Run("FlatMapNone", func(t *testing.T) {
		t.Parallel()
		none := None[int]()
		result := OptionalFlatMap(none, func(_ int) Optional[string] {
			return Some("value")
		})
		assert.True(t, result.IsEmpty(), "OptionalFlatMap None should remain None")
	})
}

// TestOptionalZip tests OptionalZip function.
func TestOptionalZip(t *testing.T) {
	t.Parallel()
	t.Run("ZipBothPresent", func(t *testing.T) {
		t.Parallel()
		o1 := Some(42)
		o2 := Some("hello")
		result := OptionalZip(o1, o2)
		assert.True(t, result.IsPresent(), "OptionalZip both present should be present")
		assert.Equal(t, 42, result.Get().First, "OptionalZip should combine values")
		assert.Equal(t, "hello", result.Get().Second, "OptionalZip should combine values")
	})

	t.Run("ZipFirstNone", func(t *testing.T) {
		t.Parallel()
		o1 := None[int]()
		o2 := Some("hello")
		result := OptionalZip(o1, o2)
		assert.True(t, result.IsEmpty(), "OptionalZip with first None should be empty")
	})

	t.Run("ZipSecondNone", func(t *testing.T) {
		t.Parallel()
		o1 := Some(42)
		o2 := None[string]()
		result := OptionalZip(o1, o2)
		assert.True(t, result.IsEmpty(), "OptionalZip with second None should be empty")
	})

	t.Run("ZipBothNone", func(t *testing.T) {
		t.Parallel()
		o1 := None[int]()
		o2 := None[string]()
		result := OptionalZip(o1, o2)
		assert.True(t, result.IsEmpty(), "OptionalZip both None should be empty")
	})
}

// TestOptionalEquals tests OptionalEquals function.
func TestOptionalEquals(t *testing.T) {
	t.Parallel()
	t.Run("BothSomeEqual", func(t *testing.T) {
		t.Parallel()
		assert.True(t, OptionalEquals(Some(42), Some(42)), "Equal Some values should be equal")
	})

	t.Run("BothSomeNotEqual", func(t *testing.T) {
		t.Parallel()
		assert.False(t, OptionalEquals(Some(42), Some(43)), "Different Some values should not be equal")
	})

	t.Run("BothNone", func(t *testing.T) {
		t.Parallel()
		assert.True(t, OptionalEquals(None[int](), None[int]()), "Both None should be equal")
	})

	t.Run("SomeAndNone", func(t *testing.T) {
		t.Parallel()
		assert.False(t, OptionalEquals(Some(42), None[int]()), "Some and None should not be equal")
	})
}
