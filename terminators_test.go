package streams

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestTerminators tests terminal operations on Stream.
func TestTerminators(t *testing.T) {
	t.Parallel()
	t.Run("ForEach", func(t *testing.T) {
		t.Parallel()
		var collected []int
		Of(1, 2, 3).ForEach(func(n int) {
			collected = append(collected, n)
		})
		assert.Equal(t, []int{1, 2, 3}, collected, "ForEach should visit all elements")
	})

	t.Run("ForEachIndexed", func(t *testing.T) {
		t.Parallel()
		var (
			indices []int
			values  []string
		)
		Of("a", "b", "c").ForEachIndexed(func(i int, s string) {
			indices = append(indices, i)
			values = append(values, s)
		})
		assert.Equal(t, []int{0, 1, 2}, indices, "ForEachIndexed should provide indices")
		assert.Equal(t, []string{"a", "b", "c"}, values, "ForEachIndexed should provide values")
	})

	t.Run("ForEachErr", func(t *testing.T) {
		t.Parallel()

		t.Run("Success", func(t *testing.T) {
			t.Parallel()
			var collected []int
			err := Of(1, 2, 3).ForEachErr(func(n int) error {
				collected = append(collected, n*2)
				return nil
			})
			assert.NoError(t, err, "ForEachErr should not return error on success")
			assert.Equal(t, []int{2, 4, 6}, collected, "ForEachErr should visit all elements")
		})

		t.Run("Error", func(t *testing.T) {
			t.Parallel()
			var collected []int
			expectedErr := assert.AnError
			err := Of(1, 2, 3, 4).ForEachErr(func(n int) error {
				if n == 3 {
					return expectedErr
				}
				collected = append(collected, n)
				return nil
			})
			assert.ErrorIs(t, err, expectedErr, "ForEachErr should return the error")
			assert.Equal(t, []int{1, 2}, collected, "ForEachErr should stop at first error")
		})

		t.Run("Empty", func(t *testing.T) {
			t.Parallel()
			err := Empty[int]().ForEachErr(func(n int) error {
				return assert.AnError
			})
			assert.NoError(t, err, "ForEachErr on empty stream should return nil")
		})
	})

	t.Run("ForEachIndexedErr", func(t *testing.T) {
		t.Parallel()

		t.Run("Success", func(t *testing.T) {
			t.Parallel()
			var indices []int
			var values []string
			err := Of("a", "b", "c").ForEachIndexedErr(func(i int, s string) error {
				indices = append(indices, i)
				values = append(values, s)
				return nil
			})
			assert.NoError(t, err, "ForEachIndexedErr should not return error on success")
			assert.Equal(t, []int{0, 1, 2}, indices, "ForEachIndexedErr should provide indices")
			assert.Equal(t, []string{"a", "b", "c"}, values, "ForEachIndexedErr should provide values")
		})

		t.Run("Error", func(t *testing.T) {
			t.Parallel()
			var indices []int
			expectedErr := assert.AnError
			err := Of("a", "b", "c", "d").ForEachIndexedErr(func(i int, s string) error {
				if i == 2 {
					return expectedErr
				}
				indices = append(indices, i)
				return nil
			})
			assert.ErrorIs(t, err, expectedErr, "ForEachIndexedErr should return the error")
			assert.Equal(t, []int{0, 1}, indices, "ForEachIndexedErr should stop at first error")
		})

		t.Run("Empty", func(t *testing.T) {
			t.Parallel()
			err := Empty[string]().ForEachIndexedErr(func(i int, s string) error {
				return assert.AnError
			})
			assert.NoError(t, err, "ForEachIndexedErr on empty stream should return nil")
		})
	})

	t.Run("Collect", func(t *testing.T) {
		t.Parallel()
		result := Of(1, 2, 3).Collect()
		assert.Equal(t, []int{1, 2, 3}, result, "Collect should gather all elements")
	})

	t.Run("Reduce", func(t *testing.T) {
		t.Parallel()
		result := Of(1, 2, 3, 4).Reduce(0, func(acc, n int) int {
			return acc + n
		})
		assert.Equal(t, 10, result, "Reduce should combine all elements")

		emptyResult := Empty[int]().Reduce(100, func(acc, n int) int {
			return acc + n
		})
		assert.Equal(t, 100, emptyResult, "Reduce on empty stream should return identity")
	})

	t.Run("ReduceOptional", func(t *testing.T) {
		t.Parallel()
		result := Of(1, 2, 3).ReduceOptional(func(a, b int) int {
			return a + b
		})
		assert.True(t, result.IsPresent(), "ReduceOptional should return Some for non-empty")
		assert.Equal(t, 6, result.Get(), "ReduceOptional should combine elements")

		emptyResult := Empty[int]().ReduceOptional(func(a, b int) int {
			return a + b
		})
		assert.True(t, emptyResult.IsEmpty(), "ReduceOptional on empty stream should return None")
	})

	t.Run("Fold", func(t *testing.T) {
		t.Parallel()
		result := Of("a", "b", "c").Fold("", func(acc, s string) string {
			return acc + s
		})
		assert.Equal(t, "abc", result, "Fold should be alias for Reduce")
	})

	t.Run("Count", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, 5, Of(1, 2, 3, 4, 5).Count(), "Count should return number of elements")
		assert.Equal(t, 0, Empty[int]().Count(), "Count of empty stream should be 0")
	})

	t.Run("First", func(t *testing.T) {
		t.Parallel()
		first := Of(1, 2, 3).First()
		assert.True(t, first.IsPresent(), "First should return Some for non-empty")
		assert.Equal(t, 1, first.Get(), "First should return first element")

		emptyFirst := Empty[int]().First()
		assert.True(t, emptyFirst.IsEmpty(), "First on empty should return None")
	})

	t.Run("Last", func(t *testing.T) {
		t.Parallel()
		last := Of(1, 2, 3).Last()
		assert.True(t, last.IsPresent(), "Last should return Some for non-empty")
		assert.Equal(t, 3, last.Get(), "Last should return last element")

		emptyLast := Empty[int]().Last()
		assert.True(t, emptyLast.IsEmpty(), "Last on empty should return None")
	})

	t.Run("FindFirst", func(t *testing.T) {
		t.Parallel()
		found := Of(1, 2, 3, 4, 5).FindFirst(func(n int) bool { return n > 3 })
		assert.True(t, found.IsPresent(), "FindFirst should find matching element")
		assert.Equal(t, 4, found.Get(), "FindFirst should return first match")

		notFound := Of(1, 2, 3).FindFirst(func(n int) bool { return n > 10 })
		assert.True(t, notFound.IsEmpty(), "FindFirst should return None when no match")
	})

	t.Run("FindLast", func(t *testing.T) {
		t.Parallel()
		found := Of(1, 2, 3, 4, 5).FindLast(func(n int) bool { return n > 3 })
		assert.True(t, found.IsPresent(), "FindLast should find matching element")
		assert.Equal(t, 5, found.Get(), "FindLast should return last match")

		notFound := Of(1, 2, 3).FindLast(func(n int) bool { return n > 10 })
		assert.True(t, notFound.IsEmpty(), "FindLast should return None when no match")
	})

	t.Run("AnyMatch", func(t *testing.T) {
		t.Parallel()
		assert.True(t, Of(1, 2, 3).AnyMatch(func(n int) bool { return n == 2 }), "AnyMatch should find match")
		assert.False(t, Of(1, 2, 3).AnyMatch(func(n int) bool { return n == 5 }), "AnyMatch should return false when no match")
		assert.False(t, Empty[int]().AnyMatch(func(n int) bool { return true }), "AnyMatch on empty should be false")
	})

	t.Run("AllMatch", func(t *testing.T) {
		t.Parallel()
		assert.True(t, Of(2, 4, 6).AllMatch(func(n int) bool { return n%2 == 0 }), "AllMatch should be true when all match")
		assert.False(t, Of(2, 3, 6).AllMatch(func(n int) bool { return n%2 == 0 }), "AllMatch should be false when one doesn't match")
		assert.True(t, Empty[int]().AllMatch(func(n int) bool { return false }), "AllMatch on empty should be true")
	})

	t.Run("NoneMatch", func(t *testing.T) {
		t.Parallel()
		assert.True(t, Of(1, 3, 5).NoneMatch(func(n int) bool { return n%2 == 0 }), "NoneMatch should be true when none match")
		assert.False(t, Of(1, 2, 3).NoneMatch(func(n int) bool { return n == 2 }), "NoneMatch should be false when one matches")
		assert.True(t, Empty[int]().NoneMatch(func(n int) bool { return true }), "NoneMatch on empty should be true")
	})

	t.Run("Contains", func(t *testing.T) {
		t.Parallel()
		assert.True(t, Contains(Of(1, 2, 3), 2), "Contains should find element")
		assert.False(t, Contains(Of(1, 2, 3), 5), "Contains should return false for missing element")
	})

	t.Run("Min", func(t *testing.T) {
		t.Parallel()
		min := Of(3, 1, 4, 1, 5).Min(func(a, b int) int { return a - b })
		assert.True(t, min.IsPresent(), "Min should return Some for non-empty")
		assert.Equal(t, 1, min.Get(), "Min should return minimum element")

		emptyMin := Empty[int]().Min(func(a, b int) int { return a - b })
		assert.True(t, emptyMin.IsEmpty(), "Min on empty should return None")
	})

	t.Run("Max", func(t *testing.T) {
		t.Parallel()
		max := Of(3, 1, 4, 1, 5).Max(func(a, b int) int { return a - b })
		assert.True(t, max.IsPresent(), "Max should return Some for non-empty")
		assert.Equal(t, 5, max.Get(), "Max should return maximum element")

		emptyMax := Empty[int]().Max(func(a, b int) int { return a - b })
		assert.True(t, emptyMax.IsEmpty(), "Max on empty should return None")
	})

	t.Run("At", func(t *testing.T) {
		t.Parallel()
		s := Of(10, 20, 30, 40, 50)

		assert.Equal(t, 10, s.At(0).Get(), "At(0) should return first element")
		assert.Equal(t, 30, s.At(2).Get(), "At(2) should return third element")
		assert.True(t, Of(1, 2, 3).At(10).IsEmpty(), "At with out of bounds should return None")
		assert.True(t, Of(1, 2, 3).At(-1).IsEmpty(), "At with negative index should return None")
	})

	t.Run("Nth", func(t *testing.T) {
		t.Parallel()
		s := Of(10, 20, 30, 40, 50)

		assert.Equal(t, 10, s.Nth(0).Get(), "Nth(0) should return first element")
		assert.Equal(t, 30, s.Nth(2).Get(), "Nth(2) should return third element")
		assert.Equal(t, 50, s.Nth(4).Get(), "Nth(4) should return last element")
		assert.True(t, Of(1, 2, 3).Nth(10).IsEmpty(), "Nth with out of bounds should return None")
		assert.True(t, Of(1, 2, 3).Nth(-1).IsEmpty(), "Nth with negative index should return None")
		assert.True(t, Empty[int]().Nth(0).IsEmpty(), "Nth on empty stream should return None")
	})

	t.Run("Single", func(t *testing.T) {
		t.Parallel()
		single := Of(42).Single()
		assert.True(t, single.IsPresent(), "Single should return Some for single element")
		assert.Equal(t, 42, single.Get(), "Single should return the element")

		multiple := Of(1, 2).Single()
		assert.True(t, multiple.IsEmpty(), "Single should return None for multiple elements")

		empty := Empty[int]().Single()
		assert.True(t, empty.IsEmpty(), "Single should return None for empty stream")
	})

	t.Run("IsEmpty", func(t *testing.T) {
		t.Parallel()
		assert.True(t, Empty[int]().IsEmpty(), "IsEmpty should be true for empty stream")
		assert.False(t, Of(1).IsEmpty(), "IsEmpty should be false for non-empty stream")
	})

	t.Run("IsNotEmpty", func(t *testing.T) {
		t.Parallel()
		assert.False(t, Empty[int]().IsNotEmpty(), "IsNotEmpty should be false for empty stream")
		assert.True(t, Of(1).IsNotEmpty(), "IsNotEmpty should be true for non-empty stream")
	})
}

// TestFoldTo tests FoldTo function.
func TestFoldTo(t *testing.T) {
	t.Parallel()
	result := FoldTo(Of(1, 2, 3), "", func(acc string, n int) string {
		return acc + string(rune('a'+n-1))
	})
	assert.Equal(t, "abc", result, "FoldTo should transform to different type")
}

// TestToMap tests ToMap function.
func TestToMap(t *testing.T) {
	t.Parallel()
	type Person struct {
		Id   int
		Name string
	}
	people := []Person{
		{Id: 1, Name: "Alice"},
		{Id: 2, Name: "Bob"},
	}

	result := ToMap(FromSlice(people), func(p Person) (int, string) {
		return p.Id, p.Name
	})

	assert.Equal(t, "Alice", result[1], "ToMap should create map with key function")
	assert.Equal(t, "Bob", result[2], "ToMap should create map with value function")
}

// TestToSet tests ToSet function.
func TestToSet(t *testing.T) {
	t.Parallel()
	result := ToSet(Of(1, 2, 2, 3, 3, 3))
	assert.Len(t, result, 3, "ToSet should create set without duplicates")
	_, ok := result[1]
	assert.True(t, ok, "ToSet should contain element")
}

// TestGroupBy tests GroupBy function.
func TestGroupBy(t *testing.T) {
	t.Parallel()
	result := GroupBy(Of(1, 2, 3, 4, 5, 6), func(n int) string {
		if n%2 == 0 {
			return "even"
		}
		return "odd"
	})

	assert.Equal(t, []int{1, 3, 5}, result["odd"], "GroupBy should group odd numbers")
	assert.Equal(t, []int{2, 4, 6}, result["even"], "GroupBy should group even numbers")
}

// TestGroupByTo tests GroupByTo function.
func TestGroupByTo(t *testing.T) {
	t.Parallel()
	result := GroupByTo(Of(1, 2, 3, 4), func(n int) (string, int) {
		if n%2 == 0 {
			return "even", n * 10
		}
		return "odd", n * 10
	})

	assert.Equal(t, []int{10, 30}, result["odd"], "GroupByTo should transform values")
	assert.Equal(t, []int{20, 40}, result["even"], "GroupByTo should transform values")
}

// TestPartitionBy tests PartitionBy function.
func TestPartitionBy(t *testing.T) {
	t.Parallel()
	matching, notMatching := PartitionBy(Of(1, 2, 3, 4, 5), func(n int) bool {
		return n > 3
	})

	assert.Equal(t, []int{4, 5}, matching, "PartitionBy should return matching elements")
	assert.Equal(t, []int{1, 2, 3}, notMatching, "PartitionBy should return non-matching elements")
}

// TestJoining tests Joining functions.
func TestJoining(t *testing.T) {
	t.Parallel()
	t.Run("Joining", func(t *testing.T) {
		t.Parallel()
		result := Joining(Of("a", "b", "c"), ", ")
		assert.Equal(t, "a, b, c", result, "Joining should concatenate with separator")
	})

	t.Run("JoiningEmpty", func(t *testing.T) {
		t.Parallel()
		result := Joining(Empty[string](), ", ")
		assert.Equal(t, "", result, "Joining empty stream should return empty string")
	})

	t.Run("JoiningWithPrefixSuffix", func(t *testing.T) {
		t.Parallel()
		result := JoiningWithPrefixSuffix(Of("a", "b", "c"), ", ", "[", "]")
		assert.Equal(t, "[a, b, c]", result, "JoiningWithPrefixSuffix should add prefix and suffix")
	})
}

// TestAssociate tests Associate functions.
func TestAssociate(t *testing.T) {
	t.Parallel()
	t.Run("ToMap", func(t *testing.T) {
		t.Parallel()
		result := ToMap(Of("abc", "de", "f"), func(s string) (int, string) {
			return len(s), s
		})

		assert.Equal(t, "abc", result[3], "ToMap should map by function result")
		assert.Equal(t, "de", result[2], "ToMap should map by function result")
		assert.Equal(t, "f", result[1], "ToMap should map by function result")
	})

	t.Run("AssociateBy", func(t *testing.T) {
		t.Parallel()
		result := AssociateBy(Of("abc", "de", "f"), func(s string) int {
			return len(s)
		})

		assert.Equal(t, "abc", result[3], "AssociateBy should map by key function")
	})

	t.Run("IndexBy", func(t *testing.T) {
		t.Parallel()
		type Person struct {
			Id   int
			Name string
		}
		people := []Person{
			{Id: 1, Name: "Alice"},
			{Id: 2, Name: "Bob"},
		}

		result := IndexBy(FromSlice(people), func(p Person) int {
			return p.Id
		})

		assert.Equal(t, "Alice", result[1].Name, "IndexBy should index by key")
	})
}

// TestCountBy tests CountBy function.
func TestCountBy(t *testing.T) {
	t.Parallel()
	result := CountBy(Of("apple", "apricot", "banana", "blueberry"), func(s string) rune {
		return rune(s[0])
	})

	assert.Equal(t, 2, result['a'], "CountBy should count by key")
	assert.Equal(t, 2, result['b'], "CountBy should count by key")
}

// TestFrequencies tests Frequencies function.
func TestFrequencies(t *testing.T) {
	t.Parallel()
	result := Frequencies(Of("a", "b", "a", "c", "a", "b"))

	assert.Equal(t, 3, result["a"], "Frequencies should count occurrences")
	assert.Equal(t, 2, result["b"], "Frequencies should count occurrences")
	assert.Equal(t, 1, result["c"], "Frequencies should count occurrences")
}
