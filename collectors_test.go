package streams

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCollectors tests Collector operations.
func TestCollectors(t *testing.T) {
	t.Parallel()
	t.Run("ToSliceCollector", func(t *testing.T) {
		t.Parallel()
		result := CollectTo(Of(1, 2, 3), ToSliceCollector[int]())
		assert.Equal(t, []int{1, 2, 3}, result, "ToSliceCollector should collect to slice")
	})

	t.Run("ToSetCollector", func(t *testing.T) {
		t.Parallel()
		result := CollectTo(Of(1, 2, 2, 3, 3), ToSetCollector[int]())
		assert.Len(t, result, 3, "ToSetCollector should create set without duplicates")
		_, ok := result[1]
		assert.True(t, ok, "Set should contain element")
	})

	t.Run("JoiningCollector", func(t *testing.T) {
		t.Parallel()
		result := CollectTo(Of("a", "b", "c"), JoiningCollector(", "))
		assert.Equal(t, "a, b, c", result, "JoiningCollector should join with separator")
	})

	t.Run("JoiningCollectorEmpty", func(t *testing.T) {
		t.Parallel()
		result := CollectTo(Empty[string](), JoiningCollector(", "))
		assert.Equal(t, "", result, "JoiningCollector on empty should return empty string")
	})

	t.Run("JoiningCollectorFull", func(t *testing.T) {
		t.Parallel()
		result := CollectTo(Of("a", "b", "c"), JoiningCollectorFull(", ", "[", "]"))
		assert.Equal(t, "[a, b, c]", result, "JoiningCollectorFull should add prefix and suffix")
	})

	t.Run("CountingCollector", func(t *testing.T) {
		t.Parallel()
		result := CollectTo(Of(1, 2, 3, 4, 5), CountingCollector[int]())
		assert.Equal(t, 5, result, "CountingCollector should count elements")
	})

	t.Run("SummingCollector", func(t *testing.T) {
		t.Parallel()
		result := CollectTo(Of(1, 2, 3, 4, 5), SummingCollector[int]())
		assert.Equal(t, 15, result, "SummingCollector should sum elements")
	})

	t.Run("AveragingCollector", func(t *testing.T) {
		t.Parallel()
		result := CollectTo(Of(1, 2, 3, 4, 5), AveragingCollector[int]())
		assert.True(t, result.IsPresent(), "AveragingCollector should return Some for non-empty")
		assert.Equal(t, 3.0, result.Get(), "AveragingCollector should compute average")

		emptyResult := CollectTo(Empty[int](), AveragingCollector[int]())
		assert.True(t, emptyResult.IsEmpty(), "AveragingCollector on empty should return None")
	})

	t.Run("MaxByCollector", func(t *testing.T) {
		t.Parallel()
		result := CollectTo(Of(3, 1, 4, 1, 5), MaxByCollector(func(a, b int) int { return a - b }))
		assert.True(t, result.IsPresent(), "MaxByCollector should return Some for non-empty")
		assert.Equal(t, 5, result.Get(), "MaxByCollector should find maximum")

		emptyResult := CollectTo(Empty[int](), MaxByCollector(func(a, b int) int { return a - b }))
		assert.True(t, emptyResult.IsEmpty(), "MaxByCollector on empty should return None")
	})

	t.Run("MinByCollector", func(t *testing.T) {
		t.Parallel()
		result := CollectTo(Of(3, 1, 4, 1, 5), MinByCollector(func(a, b int) int { return a - b }))
		assert.True(t, result.IsPresent(), "MinByCollector should return Some for non-empty")
		assert.Equal(t, 1, result.Get(), "MinByCollector should find minimum")
	})

	t.Run("GroupingByCollector", func(t *testing.T) {
		t.Parallel()
		result := CollectTo(Of(1, 2, 3, 4, 5, 6), GroupingByCollector(func(n int) string {
			if n%2 == 0 {
				return "even"
			}
			return "odd"
		}))
		assert.Equal(t, []int{1, 3, 5}, result["odd"], "GroupingByCollector should group odd")
		assert.Equal(t, []int{2, 4, 6}, result["even"], "GroupingByCollector should group even")
	})

	t.Run("PartitioningByCollector", func(t *testing.T) {
		t.Parallel()
		result := CollectTo(Of(1, 2, 3, 4, 5), PartitioningByCollector(func(n int) bool { return n > 3 }))
		assert.Equal(t, []int{4, 5}, result[true], "PartitioningByCollector should partition matching")
		assert.Equal(t, []int{1, 2, 3}, result[false], "PartitioningByCollector should partition non-matching")
	})

	t.Run("ToMapCollector", func(t *testing.T) {
		t.Parallel()
		type Person struct {
			Id   int
			Name string
		}
		people := []Person{
			{Id: 1, Name: "Alice"},
			{Id: 2, Name: "Bob"},
		}

		result := CollectTo(FromSlice(people), ToMapCollector(
			func(p Person) (int, string) { return p.Id, p.Name },
		))

		assert.Equal(t, "Alice", result[1], "ToMapCollector should create map")
		assert.Equal(t, "Bob", result[2], "ToMapCollector should create map")
	})

	t.Run("ToMapCollectorMerging", func(t *testing.T) {
		t.Parallel()
		type Item struct {
			Key   string
			Value int
		}
		items := []Item{
			{Key: "a", Value: 1},
			{Key: "a", Value: 2},
			{Key: "b", Value: 3},
		}

		result := CollectTo(FromSlice(items), ToMapCollectorMerging(
			func(i Item) (string, int) { return i.Key, i.Value },
			func(v1, v2 int) int { return v1 + v2 },
		))

		assert.Equal(t, 3, result["a"], "ToMapCollectorMerging should merge values")
		assert.Equal(t, 3, result["b"], "ToMapCollectorMerging should handle non-duplicates")
	})

	t.Run("FirstCollector", func(t *testing.T) {
		t.Parallel()
		result := CollectTo(Of(1, 2, 3), FirstCollector[int]())
		assert.True(t, result.IsPresent(), "FirstCollector should return Some for non-empty")
		assert.Equal(t, 1, result.Get(), "FirstCollector should return first element")

		emptyResult := CollectTo(Empty[int](), FirstCollector[int]())
		assert.True(t, emptyResult.IsEmpty(), "FirstCollector on empty should return None")
	})

	t.Run("LastCollector", func(t *testing.T) {
		t.Parallel()
		result := CollectTo(Of(1, 2, 3), LastCollector[int]())
		assert.True(t, result.IsPresent(), "LastCollector should return Some for non-empty")
		assert.Equal(t, 3, result.Get(), "LastCollector should return last element")
	})

	t.Run("ReducingCollector", func(t *testing.T) {
		t.Parallel()
		result := CollectTo(Of(1, 2, 3, 4), ReducingCollector(0, func(a, b int) int { return a + b }))
		assert.Equal(t, 10, result, "ReducingCollector should reduce elements")
	})
}

// TestCompositeCollectors tests composite collector operations.
func TestCompositeCollectors(t *testing.T) {
	t.Parallel()
	t.Run("MappingCollector", func(t *testing.T) {
		t.Parallel()
		result := CollectTo(Of(1, 2, 3), MappingCollector(
			func(n int) string { return string(rune('a' + n - 1)) },
			JoiningCollector(", "),
		))
		assert.Equal(t, "a, b, c", result, "MappingCollector should transform before collecting")
	})

	t.Run("FilteringCollector", func(t *testing.T) {
		t.Parallel()
		result := CollectTo(Of(1, 2, 3, 4, 5), FilteringCollector(
			func(n int) bool { return n%2 == 0 },
			ToSliceCollector[int](),
		))
		assert.Equal(t, []int{2, 4}, result, "FilteringCollector should filter before collecting")
	})

	t.Run("FlatMappingCollector", func(t *testing.T) {
		t.Parallel()
		result := CollectTo(Of(1, 2, 3), FlatMappingCollector(
			func(n int) Stream[int] { return Of(n, n*10) },
			ToSliceCollector[int](),
		))
		assert.Equal(t, []int{1, 10, 2, 20, 3, 30}, result, "FlatMappingCollector should flatten before collecting")
	})

	t.Run("TeeingCollector", func(t *testing.T) {
		t.Parallel()
		result := CollectTo(Of(1, 2, 3, 4, 5), TeeingCollector(
			SummingCollector[int](),
			CountingCollector[int](),
			func(sum int, count int) float64 {
				return float64(sum) / float64(count)
			},
		))
		assert.Equal(t, 3.0, result, "TeeingCollector should combine two collectors")
	})
}

// TestPercentile tests Percentile function.
func TestPercentile(t *testing.T) {
	t.Parallel()
	less := func(a, b int) bool { return a < b }

	t.Run("Percentile75", func(t *testing.T) {
		t.Parallel()
		result := Percentile(Of(1, 2, 3, 4, 5, 6, 7, 8, 9, 10), 75, less)
		assert.True(t, result.IsPresent(), "Percentile should return Some for non-empty")
		// 75th percentile of 1-10 should be around 7-8
		assert.True(t, result.Get() >= 7, "75th percentile should be >= 7")
	})

	t.Run("Percentile50", func(t *testing.T) {
		t.Parallel()
		result := Percentile(Of(1, 2, 3, 4, 5), 50, less)
		assert.True(t, result.IsPresent(), "Percentile 50 should return Some")
		assert.Equal(t, 3, result.Get(), "50th percentile (median) of 1-5 should be 3")
	})

	t.Run("Percentile0", func(t *testing.T) {
		t.Parallel()
		result := Percentile(Of(1, 2, 3, 4, 5), 0, less)
		assert.True(t, result.IsPresent(), "Percentile 0 should return Some")
		assert.Equal(t, 1, result.Get(), "0th percentile should be minimum")
	})

	t.Run("Percentile100", func(t *testing.T) {
		t.Parallel()
		result := Percentile(Of(1, 2, 3, 4, 5), 100, less)
		assert.True(t, result.IsPresent(), "Percentile 100 should return Some")
		assert.Equal(t, 5, result.Get(), "100th percentile should be maximum")
	})

	t.Run("PercentileEmpty", func(t *testing.T) {
		t.Parallel()
		result := Percentile(Empty[int](), 50, less)
		assert.True(t, result.IsEmpty(), "Percentile on empty should return None")
	})
}

// TestHistogramCollector tests HistogramCollector function.
func TestHistogramCollector(t *testing.T) {
	t.Parallel()
	t.Run("HistogramByLength", func(t *testing.T) {
		t.Parallel()
		result := CollectTo(
			Of("a", "bb", "ccc", "dd", "eee"),
			HistogramCollector(func(s string) int { return len(s) }),
		)
		assert.Len(t, result[1], 1, "Bucket for length 1 should have 1 element")
		assert.Len(t, result[2], 2, "Bucket for length 2 should have 2 elements")
		assert.Len(t, result[3], 2, "Bucket for length 3 should have 2 elements")
		assert.ElementsMatch(t, []string{"a"}, result[1], "Bucket for length 1 should contain the single-letter item")
		assert.ElementsMatch(t, []string{"bb", "dd"}, result[2], "Bucket for length 2 should contain the 2-letter items")
		assert.ElementsMatch(t, []string{"ccc", "eee"}, result[3], "Bucket for length 3 should contain the 3-letter items")
	})

	t.Run("HistogramByFirstChar", func(t *testing.T) {
		t.Parallel()
		result := CollectTo(
			Of("apple", "apricot", "banana", "cherry"),
			HistogramCollector(func(s string) byte { return s[0] }),
		)
		assert.Len(t, result['a'], 2, "Bucket for 'a' should have 2 elements")
		assert.Len(t, result['b'], 1, "Bucket for 'b' should have 1 element")
		assert.Len(t, result['c'], 1, "Bucket for 'c' should have 1 element")
	})

	t.Run("HistogramEmpty", func(t *testing.T) {
		t.Parallel()
		result := CollectTo(
			Empty[string](),
			HistogramCollector(func(s string) int { return len(s) }),
		)
		assert.Empty(t, result, "Histogram of empty stream should be empty map")
	})
}

// TestMinByCollectorEmpty tests MinByCollector on empty stream.
func TestMinByCollectorEmpty(t *testing.T) {
	t.Parallel()
	emptyResult := CollectTo(Empty[int](), MinByCollector(func(a, b int) int { return a - b }))
	assert.True(t, emptyResult.IsEmpty(), "MinByCollector on empty should return None")
}

// TestLastCollectorEmpty tests LastCollector on empty stream.
func TestLastCollectorEmpty(t *testing.T) {
	t.Parallel()
	emptyResult := CollectTo(Empty[int](), LastCollector[int]())
	assert.True(t, emptyResult.IsEmpty(), "LastCollector on empty should return None")
}

// TestTopKBottomK tests TopK and BottomK collectors.
func TestTopKBottomK(t *testing.T) {
	t.Parallel()
	less := func(a, b int) bool { return a < b }

	// Table-driven to reduce duplication while keeping descriptive subtest names.
	type tc struct {
		name     string
		input    []int
		k        int
		useTop   bool // if false use BottomK
		expected []int
	}
	cases := []tc{
		{
			name:     "TopKBasic",
			input:    []int{5, 2, 8, 1, 9, 3, 7, 4, 6},
			k:        3,
			useTop:   true,
			expected: []int{9, 8, 7},
		},
		{
			name:     "TopKLargerK",
			input:    []int{3, 1, 2},
			k:        10,
			useTop:   true,
			expected: []int{3, 2, 1},
		},
		{
			name:     "TopKEmpty",
			input:    []int{},
			k:        3,
			useTop:   true,
			expected: []int{},
		},
		{
			name:     "BottomKBasic",
			input:    []int{5, 2, 8, 1, 9, 3, 7, 4, 6},
			k:        3,
			useTop:   false,
			expected: []int{1, 2, 3},
		},
		{
			name:     "BottomKLargerK",
			input:    []int{3, 1, 2},
			k:        10,
			useTop:   false,
			expected: []int{1, 2, 3},
		},
		{
			name:     "BottomKEmpty",
			input:    []int{},
			k:        3,
			useTop:   false,
			expected: []int{},
		},
		{
			name:     "BottomKWithDuplicates",
			input:    []int{5, 2, 5, 2, 1, 3, 1},
			k:        4,
			useTop:   false,
			expected: []int{1, 1, 2, 2},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			var s Stream[int]
			if len(c.input) == 0 {
				s = Empty[int]()
			} else {
				s = FromSlice(c.input)
			}
			var out []int
			if c.useTop {
				out = CollectTo(s, TopKCollector(c.k, less))
			} else {
				out = CollectTo(s, BottomKCollector(c.k, less))
			}
			assert.Equal(t, c.expected, out, "Collector %s expected output mismatch", c.name)
		})
	}
}

// TestQuantileCollectorUnsorted tests QuantileCollector with unsorted input.
func TestQuantileCollectorUnsorted(t *testing.T) {
	t.Parallel()
	less := func(a, b int) bool { return a < b }

	t.Run("QuantileUnsorted", func(t *testing.T) {
		t.Parallel()
		// Use unsorted input to trigger the less(a, b) branch in sorting
		result := Quantile(Of(9, 3, 7, 1, 5, 8, 2, 6, 4), 0.5, less)
		assert.True(t, result.IsPresent(), "Quantile should return Some")
		assert.Equal(t, 5, result.Get(), "Median of 1-9 unsorted should be 5")
	})

	t.Run("QuantileReverseSorted", func(t *testing.T) {
		t.Parallel()
		// Reverse sorted to force sorting comparisons
		result := Quantile(Of(10, 9, 8, 7, 6, 5, 4, 3, 2, 1), 0.75, less)
		assert.True(t, result.IsPresent(), "Quantile should return Some")
		assert.True(t, result.Get() >= 7, "75th percentile should be >= 7")
	})

	t.Run("Median", func(t *testing.T) {
		t.Parallel()
		// Test Median function with unsorted input
		result := Median(Of(5, 1, 3, 2, 4), less)
		assert.True(t, result.IsPresent(), "Median should return Some")
		assert.Equal(t, 3, result.Get(), "Median of 1-5 should be 3")
	})

	t.Run("PercentileUnsorted", func(t *testing.T) {
		t.Parallel()
		// Test Percentile with unsorted input to trigger sorting branches
		result := Percentile(Of(10, 9, 8, 7, 6, 5, 4, 3, 2, 1), 25, less)
		assert.True(t, result.IsPresent(), "Percentile should return Some")
		// 25th percentile of sorted 1-10 should be around 2-3
		assert.True(t, result.Get() >= 2 && result.Get() <= 3, "25th percentile should be 2 or 3")
	})
}
