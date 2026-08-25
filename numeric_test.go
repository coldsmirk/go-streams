package streams

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNumericOperations tests numeric stream functions.
func TestNumericOperations(t *testing.T) {
	t.Parallel()
	t.Run("Sum", func(t *testing.T) {
		t.Parallel()
		result := Sum(Of(1, 2, 3, 4, 5))
		assert.Equal(t, 15, result, "Sum should add all elements")

		emptyResult := Sum(Empty[int]())
		assert.Equal(t, 0, emptyResult, "Sum of empty should be 0")
	})

	t.Run("SumFloat", func(t *testing.T) {
		t.Parallel()
		result := Sum(Of(1.5, 2.5, 3.0))
		assert.Equal(t, 7.0, result, "Sum should work with floats")
	})

	t.Run("Average", func(t *testing.T) {
		t.Parallel()
		result := Average(Of(1, 2, 3, 4, 5))
		assert.True(t, result.IsPresent(), "Average should return Some for non-empty")
		assert.Equal(t, 3.0, result.Get(), "Average should compute average")

		emptyResult := Average(Empty[int]())
		assert.True(t, emptyResult.IsEmpty(), "Average of empty should be None")
	})

	t.Run("MinValue", func(t *testing.T) {
		t.Parallel()
		result := MinValue(Of(3, 1, 4, 1, 5))
		assert.True(t, result.IsPresent(), "MinValue should return Some for non-empty")
		assert.Equal(t, 1, result.Get(), "MinValue should find minimum")

		emptyResult := MinValue(Empty[int]())
		assert.True(t, emptyResult.IsEmpty(), "MinValue of empty should be None")
	})

	t.Run("MaxValue", func(t *testing.T) {
		t.Parallel()
		result := MaxValue(Of(3, 1, 4, 1, 5))
		assert.True(t, result.IsPresent(), "MaxValue should return Some for non-empty")
		assert.Equal(t, 5, result.Get(), "MaxValue should find maximum")

		emptyResult := MaxValue(Empty[int]())
		assert.True(t, emptyResult.IsEmpty(), "MaxValue of empty should be None")
	})

	t.Run("MinMax", func(t *testing.T) {
		t.Parallel()
		result := MinMax(Of(3, 1, 4, 1, 5))
		assert.True(t, result.IsPresent(), "MinMax should return Some for non-empty")
		assert.Equal(t, 1, result.Get().First, "MinMax should find minimum")
		assert.Equal(t, 5, result.Get().Second, "MinMax should find maximum")

		emptyResult := MinMax(Empty[int]())
		assert.True(t, emptyResult.IsEmpty(), "MinMax of empty should be None")
	})

	t.Run("Product", func(t *testing.T) {
		t.Parallel()
		result := Product(Of(1, 2, 3, 4, 5))
		assert.Equal(t, 120, result, "Product should multiply all elements")

		emptyResult := Product(Empty[int]())
		assert.Equal(t, 1, emptyResult, "Product of empty should be 1")
	})

	t.Run("SumBy", func(t *testing.T) {
		t.Parallel()
		type Item struct {
			Value int
		}
		items := []Item{{Value: 10}, {Value: 20}, {Value: 30}}

		result := SumBy(FromSlice(items), func(i Item) int {
			return i.Value
		})
		assert.Equal(t, 60, result, "SumBy should sum extracted values")
	})

	t.Run("AverageBy", func(t *testing.T) {
		t.Parallel()
		type Item struct {
			Value int
		}
		items := []Item{{Value: 10}, {Value: 20}, {Value: 30}}

		result := AverageBy(FromSlice(items), func(i Item) int {
			return i.Value
		})
		assert.True(t, result.IsPresent(), "AverageBy should return Some for non-empty")
		assert.Equal(t, 20.0, result.Get(), "AverageBy should compute average of extracted values")

		emptyResult := AverageBy(Empty[Item](), func(i Item) int {
			return i.Value
		})
		assert.True(t, emptyResult.IsEmpty(), "AverageBy of empty should be None")
	})

	t.Run("MinBy", func(t *testing.T) {
		t.Parallel()
		type Person struct {
			Name string
			Age  int
		}
		people := []Person{
			{Name: "Alice", Age: 30},
			{Name: "Bob", Age: 25},
			{Name: "Charlie", Age: 35},
		}

		result := MinBy(FromSlice(people), func(p Person) int { return p.Age })
		assert.True(t, result.IsPresent(), "MinBy should return Some for non-empty")
		assert.Equal(t, "Bob", result.Get().Name, "MinBy should find element with minimum key")

		emptyResult := MinBy(Empty[Person](), func(p Person) int { return p.Age })
		assert.True(t, emptyResult.IsEmpty(), "MinBy of empty should be None")
	})

	t.Run("MaxBy", func(t *testing.T) {
		t.Parallel()
		type Person struct {
			Name string
			Age  int
		}
		people := []Person{
			{Name: "Alice", Age: 30},
			{Name: "Bob", Age: 25},
			{Name: "Charlie", Age: 35},
		}

		result := MaxBy(FromSlice(people), func(p Person) int { return p.Age })
		assert.True(t, result.IsPresent(), "MaxBy should return Some for non-empty")
		assert.Equal(t, "Charlie", result.Get().Name, "MaxBy should find element with maximum key")

		emptyResult := MaxBy(Empty[Person](), func(p Person) int { return p.Age })
		assert.True(t, emptyResult.IsEmpty(), "MaxBy of empty should be None")
	})

	t.Run("GetStatistics", func(t *testing.T) {
		t.Parallel()
		result := GetStatistics(Of(1, 2, 3, 4, 5))
		assert.True(t, result.IsPresent(), "GetStatistics should return Some for non-empty")

		stats := result.Get()
		assert.Equal(t, 5, stats.Count, "Statistics should count elements")
		assert.Equal(t, 15, stats.Sum, "Statistics should sum elements")
		assert.Equal(t, 1, stats.Min, "Statistics should find minimum")
		assert.Equal(t, 5, stats.Max, "Statistics should find maximum")
		assert.Equal(t, 3.0, stats.Average, "Statistics should compute average")

		emptyResult := GetStatistics(Empty[int]())
		assert.True(t, emptyResult.IsEmpty(), "GetStatistics of empty should be None")

		// Test descending order to cover min/max update branches
		result2 := GetStatistics(Of(5, 3, 1, 4, 2))
		assert.True(t, result2.IsPresent(), "GetStatistics should return Some for unordered input")
		stats2 := result2.Get()
		assert.Equal(t, 1, stats2.Min, "GetStatistics should find minimum from unordered input")
		assert.Equal(t, 5, stats2.Max, "GetStatistics should find maximum from unordered input")
	})

	t.Run("RunningSum", func(t *testing.T) {
		t.Parallel()
		result := RunningSum(Of(1, 2, 3, 4, 5)).Collect()
		assert.Equal(t, []int{1, 3, 6, 10, 15}, result, "RunningSum should compute cumulative sums")
	})

	t.Run("RunningProduct", func(t *testing.T) {
		t.Parallel()
		result := RunningProduct(Of(1, 2, 3, 4)).Collect()
		assert.Equal(t, []int{1, 2, 6, 24}, result, "RunningProduct should compute cumulative products")
	})

	t.Run("Differences", func(t *testing.T) {
		t.Parallel()
		result := Differences(Of(1, 3, 6, 10, 15)).Collect()
		assert.Equal(t, []int{2, 3, 4, 5}, result, "Differences should compute differences between consecutive elements")
	})

	t.Run("Clamp", func(t *testing.T) {
		t.Parallel()
		result := Clamp(Of(1, 5, 10, 15, 20), 5, 15).Collect()
		assert.Equal(t, []int{5, 5, 10, 15, 15}, result, "Clamp should clamp values to range")
	})

	t.Run("Abs", func(t *testing.T) {
		t.Parallel()
		result := Abs(Of(-3, -1, 0, 1, 3)).Collect()
		assert.Equal(t, []int{3, 1, 0, 1, 3}, result, "Abs should compute absolute values")
	})

	t.Run("AbsFloat", func(t *testing.T) {
		t.Parallel()
		result := AbsFloat(Of(-3.5, -1.0, 0.0, 1.5, 3.0)).Collect()
		assert.Equal(t, []float64{3.5, 1.0, 0.0, 1.5, 3.0}, result, "AbsFloat should compute absolute values for floats")
	})

	t.Run("Scale", func(t *testing.T) {
		t.Parallel()
		result := Scale(Of(1, 2, 3), 10).Collect()
		assert.Equal(t, []int{10, 20, 30}, result, "Scale should multiply by factor")
	})

	t.Run("ScaleEmpty", func(t *testing.T) {
		t.Parallel()
		result := Scale(Empty[int](), 10).Collect()
		assert.Empty(t, result, "Scale on empty stream should be empty")
	})

	t.Run("ScaleZero", func(t *testing.T) {
		t.Parallel()
		result := Scale(Of(1, 2, 3), 0).Collect()
		assert.Equal(t, []int{0, 0, 0}, result, "Scale by 0 should produce zeros")
	})

	t.Run("ScaleEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := Scale(Of(1, 2, 3, 4, 5), 2).Limit(3).Collect()
		assert.Equal(t, []int{2, 4, 6}, result, "Scale with Limit should stop early")
	})

	t.Run("Offset", func(t *testing.T) {
		t.Parallel()
		result := Offset(Of(1, 2, 3), 100).Collect()
		assert.Equal(t, []int{101, 102, 103}, result, "Offset should add offset")
	})

	t.Run("OffsetEmpty", func(t *testing.T) {
		t.Parallel()
		result := Offset(Empty[int](), 100).Collect()
		assert.Empty(t, result, "Offset on empty stream should be empty")
	})

	t.Run("OffsetZero", func(t *testing.T) {
		t.Parallel()
		result := Offset(Of(1, 2, 3), 0).Collect()
		assert.Equal(t, []int{1, 2, 3}, result, "Offset by 0 should not change values")
	})

	t.Run("OffsetEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := Offset(Of(1, 2, 3, 4, 5), 10).Limit(3).Collect()
		assert.Equal(t, []int{11, 12, 13}, result, "Offset with Limit should stop early")
	})

	t.Run("Positive", func(t *testing.T) {
		t.Parallel()
		result := Positive(Of(-2, -1, 0, 1, 2)).Collect()
		assert.Equal(t, []int{1, 2}, result, "Positive should filter to positive values")
	})

	t.Run("Negative", func(t *testing.T) {
		t.Parallel()
		result := Negative(Of(-2, -1, 0, 1, 2)).Collect()
		assert.Equal(t, []int{-2, -1}, result, "Negative should filter to negative values")
	})

	t.Run("NonZero", func(t *testing.T) {
		t.Parallel()
		result := NonZero(Of(-2, -1, 0, 1, 2)).Collect()
		assert.Equal(t, []int{-2, -1, 1, 2}, result, "NonZero should filter out zeros")
	})

	t.Run("RunningSumEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := RunningSum(Of(1, 2, 3, 4, 5)).Limit(3).Collect()
		assert.Equal(t, []int{1, 3, 6}, result, "RunningSum with Limit should stop early")
	})

	t.Run("RunningProductEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := RunningProduct(Of(1, 2, 3, 4, 5)).Limit(3).Collect()
		assert.Equal(t, []int{1, 2, 6}, result, "RunningProduct with Limit should stop early")
	})

	t.Run("DifferencesEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := Differences(Of(1, 3, 6, 10, 15)).Limit(2).Collect()
		assert.Equal(t, []int{2, 3}, result, "Differences with Limit should stop early")
	})

	t.Run("ClampEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := Clamp(Of(1, 5, 10, 15, 20), 5, 15).Limit(3).Collect()
		assert.Equal(t, []int{5, 5, 10}, result, "Clamp with Limit should stop early")
	})

	t.Run("AbsEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := Abs(Of(-3, -1, 0, 1, 3)).Limit(3).Collect()
		assert.Equal(t, []int{3, 1, 0}, result, "Abs with Limit should stop early")
	})

	t.Run("AbsFloatEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := AbsFloat(Of(-3.5, -1.0, 0.0, 1.5, 3.0)).Limit(3).Collect()
		assert.Equal(t, []float64{3.5, 1.0, 0.0}, result, "AbsFloat with Limit should stop early")
	})

	t.Run("MinByEarlyTermination", func(t *testing.T) {
		t.Parallel()
		type Person struct {
			Name string
			Age  int
		}
		// MinBy consumes all elements, but we can test that it works with filter
		result := MinBy(Of(Person{"A", 30}, Person{"B", 25}).Filter(func(_ Person) bool { return true }), func(p Person) int { return p.Age })
		assert.True(t, result.IsPresent(), "MinBy should return Some for non-empty stream")
		assert.Equal(t, "B", result.Get().Name, "MinBy should return the smallest key element")
	})

	t.Run("MaxByEarlyTermination", func(t *testing.T) {
		t.Parallel()
		type Person struct {
			Name string
			Age  int
		}
		result := MaxBy(Of(Person{"A", 30}, Person{"B", 25}).Filter(func(_ Person) bool { return true }), func(p Person) int { return p.Age })
		assert.True(t, result.IsPresent(), "MaxBy should return Some for non-empty stream")
		assert.Equal(t, "A", result.Get().Name, "MaxBy should return the largest key element")
	})
}
