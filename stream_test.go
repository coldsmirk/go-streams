package streams

import (
	"iter"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestStream tests Stream constructors and basic operations.
func TestStream(t *testing.T) {
	t.Parallel()
	t.Run("Of", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			name     string
			input    []int
			expected []int
		}{
			{"SingleElement", []int{1}, []int{1}},
			{"MultipleElements", []int{1, 2, 3, 4, 5}, []int{1, 2, 3, 4, 5}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				result := Of(tt.input...).Collect()
				assert.Equal(t, tt.expected, result, "Of should create stream from values")
			})
		}

		// Test empty stream separately
		t.Run("EmptyStream", func(t *testing.T) {
			t.Parallel()
			result := Of[int]().Collect()
			assert.Empty(t, result, "Of with no values should create empty stream")
		})
	})

	t.Run("FromSlice", func(t *testing.T) {
		t.Parallel()
		slice := []string{"a", "b", "c"}
		result := FromSlice(slice).Collect()
		assert.Equal(t, slice, result, "FromSlice should create stream from slice")
	})

	t.Run("From", func(t *testing.T) {
		t.Parallel()
		slice := []int{1, 2, 3}
		seq := slices.Values(slice)
		result := From(seq).Collect()
		assert.Equal(t, slice, result, "From should wrap iter.Seq")
	})

	t.Run("FromChannel", func(t *testing.T) {
		t.Parallel()
		ch := make(chan int, 3)
		ch <- 1
		ch <- 2
		ch <- 3
		close(ch)

		result := FromChannel(ch).Collect()
		assert.Equal(t, []int{1, 2, 3}, result, "FromChannel should create stream from channel")
	})

	t.Run("Range", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			name     string
			start    int
			end      int
			expected []int
		}{
			{"NormalRange", 1, 5, []int{1, 2, 3, 4}},
			{"ZeroToThree", 0, 3, []int{0, 1, 2}},
			{"NegativeStart", -2, 2, []int{-2, -1, 0, 1}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				result := Range(tt.start, tt.end).Collect()
				assert.Equal(t, tt.expected, result, "Range should generate [start, end)")
			})
		}

		// Test empty ranges separately
		t.Run("EmptyRange", func(t *testing.T) {
			t.Parallel()
			result := Range(5, 5).Collect()
			assert.Empty(t, result, "Range with start==end should be empty")
		})

		t.Run("NegativeRange", func(t *testing.T) {
			t.Parallel()
			result := Range(5, 1).Collect()
			assert.Empty(t, result, "Range with start>end should be empty")
		})
	})

	t.Run("RangeClosed", func(t *testing.T) {
		t.Parallel()
		result := RangeClosed(1, 5).Collect()
		assert.Equal(t, []int{1, 2, 3, 4, 5}, result, "RangeClosed should generate [start, end]")
	})

	t.Run("Empty", func(t *testing.T) {
		t.Parallel()
		result := Empty[int]().Collect()
		assert.Empty(t, result, "Empty should create empty stream")
	})

	t.Run("Repeat", func(t *testing.T) {
		t.Parallel()
		result := Repeat("x", 3).Collect()
		assert.Equal(t, []string{"x", "x", "x"}, result, "Repeat should repeat value n times")

		emptyResult := Repeat("x", 0).Collect()
		assert.Empty(t, emptyResult, "Repeat with n=0 should be empty")
	})

	t.Run("Generate", func(t *testing.T) {
		t.Parallel()
		counter := 0
		result := Generate(func() int {
			counter++
			return counter
		}).Limit(5).Collect()
		assert.Equal(t, []int{1, 2, 3, 4, 5}, result, "Generate should produce infinite stream")
	})

	t.Run("Iterate", func(t *testing.T) {
		t.Parallel()
		result := Iterate(1, func(n int) int { return n * 2 }).Limit(5).Collect()
		assert.Equal(t, []int{1, 2, 4, 8, 16}, result, "Iterate should apply function repeatedly")
	})

	t.Run("Concat", func(t *testing.T) {
		t.Parallel()
		s1 := Of(1, 2)
		s2 := Of(3, 4)
		s3 := Of(5)
		result := Concat(s1, s2, s3).Collect()
		assert.Equal(t, []int{1, 2, 3, 4, 5}, result, "Concat should join streams")
	})

	t.Run("Cycle", func(t *testing.T) {
		t.Parallel()
		result := Cycle(1, 2, 3).Limit(7).Collect()
		assert.Equal(t, []int{1, 2, 3, 1, 2, 3, 1}, result, "Cycle should repeat values")

		emptyResult := Cycle[int]().Limit(5).Collect()
		assert.Empty(t, emptyResult, "Cycle with no values should be empty")
	})

	t.Run("FromMap", func(t *testing.T) {
		t.Parallel()
		m := map[string]int{"a": 1, "b": 2}
		count := FromMap(m).Count()
		assert.Equal(t, 2, count, "FromMap should create Stream2 from map")
	})

	t.Run("Seq", func(t *testing.T) {
		t.Parallel()
		s := Of(1, 2, 3)
		seq := s.Seq()

		var result []int
		for v := range seq {
			result = append(result, v)
		}
		assert.Equal(t, []int{1, 2, 3}, result, "Seq should return underlying iter.Seq")
	})
}

// TestIntermediateOperations tests intermediate operations on Stream.
func TestIntermediateOperations(t *testing.T) {
	t.Parallel()
	t.Run("Filter", func(t *testing.T) {
		t.Parallel()
		result := Of(1, 2, 3, 4, 5).
			Filter(func(n int) bool { return n%2 == 0 }).
			Collect()
		assert.Equal(t, []int{2, 4}, result, "Filter should keep matching elements")
	})

	t.Run("Map", func(t *testing.T) {
		t.Parallel()
		result := Of(1, 2, 3).
			Map(func(n int) int { return n * 2 }).
			Collect()
		assert.Equal(t, []int{2, 4, 6}, result, "Map should transform elements")
	})

	t.Run("Peek", func(t *testing.T) {
		t.Parallel()
		var peeked []int
		result := Of(1, 2, 3).
			Peek(func(n int) { peeked = append(peeked, n) }).
			Collect()
		assert.Equal(t, []int{1, 2, 3}, result, "Peek should not modify stream")
		assert.Equal(t, []int{1, 2, 3}, peeked, "Peek should execute action")
	})

	t.Run("Limit", func(t *testing.T) {
		t.Parallel()
		result := Of(1, 2, 3, 4, 5).Limit(3).Collect()
		assert.Equal(t, []int{1, 2, 3}, result, "Limit should take first n elements")

		emptyResult := Of(1, 2, 3).Limit(0).Collect()
		assert.Empty(t, emptyResult, "Limit(0) should be empty")
	})

	t.Run("Skip", func(t *testing.T) {
		t.Parallel()
		result := Of(1, 2, 3, 4, 5).Skip(2).Collect()
		assert.Equal(t, []int{3, 4, 5}, result, "Skip should skip first n elements")

		fullResult := Of(1, 2, 3).Skip(0).Collect()
		assert.Equal(t, []int{1, 2, 3}, fullResult, "Skip(0) should return all elements")
	})

	t.Run("TakeWhile", func(t *testing.T) {
		t.Parallel()
		result := Of(1, 2, 3, 4, 1, 2).
			TakeWhile(func(n int) bool { return n < 4 }).
			Collect()
		assert.Equal(t, []int{1, 2, 3}, result, "TakeWhile should take while predicate is true")
	})

	t.Run("DropWhile", func(t *testing.T) {
		t.Parallel()
		result := Of(1, 2, 3, 4, 1, 2).
			DropWhile(func(n int) bool { return n < 3 }).
			Collect()
		assert.Equal(t, []int{3, 4, 1, 2}, result, "DropWhile should drop while predicate is true")
	})

	t.Run("Sorted", func(t *testing.T) {
		t.Parallel()
		result := Of(3, 1, 4, 1, 5).
			Sorted(func(a, b int) int { return a - b }).
			Collect()
		assert.Equal(t, []int{1, 1, 3, 4, 5}, result, "Sorted should sort elements")
	})

	t.Run("Reverse", func(t *testing.T) {
		t.Parallel()
		result := Of(1, 2, 3, 4, 5).Reverse().Collect()
		assert.Equal(t, []int{5, 4, 3, 2, 1}, result, "Reverse should reverse order")
	})

	t.Run("Chunk", func(t *testing.T) {
		t.Parallel()
		result := Chunk(Of(1, 2, 3, 4, 5), 2).Collect()
		assert.Equal(t, [][]int{{1, 2}, {3, 4}, {5}}, result, "Chunk should group elements")

		emptyResult := Chunk(Of(1, 2, 3), 0).Collect()
		assert.Empty(t, emptyResult, "Chunk(0) should be empty")
	})

	// Additional early termination tests
	t.Run("PeekEarlyTermination", func(t *testing.T) {
		t.Parallel()
		var peeked []int
		result := Of(1, 2, 3, 4, 5).
			Peek(func(n int) { peeked = append(peeked, n) }).
			Limit(2).
			Collect()
		assert.Equal(t, []int{1, 2}, result, "Peek then Limit(2) should collect first two values [1 2]")
		// Peek sees the element before yield returns false, so it may see one more
		assert.True(t, len(peeked) >= 2 && len(peeked) <= 3, "Peek should see at least limited elements")
	})

	t.Run("SkipEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := Of(1, 2, 3, 4, 5).Skip(2).Limit(2).Collect()
		assert.Equal(t, []int{3, 4}, result, "Skip(2) then Limit(2) should yield [3 4]")
	})

	t.Run("TakeWhileEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := Of(1, 2, 3, 4, 5).TakeWhile(func(n int) bool { return n < 10 }).Limit(2).Collect()
		assert.Equal(t, []int{1, 2}, result, "TakeWhile(n<10) then Limit(2) should yield [1 2]")
	})

	t.Run("DropWhileEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := Of(1, 2, 3, 4, 5).DropWhile(func(n int) bool { return n < 2 }).Limit(2).Collect()
		assert.Equal(t, []int{2, 3}, result, "DropWhile(n<2) then Limit(2) should yield [2 3]")
	})

	t.Run("SortedEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := Of(5, 4, 3, 2, 1).Sorted(func(a, b int) int { return a - b }).Limit(2).Collect()
		assert.Equal(t, []int{1, 2}, result, "Sorted(asc) then Limit(2) should yield [1 2]")
	})

	t.Run("ReverseEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := Of(1, 2, 3, 4, 5).Reverse().Limit(2).Collect()
		assert.Equal(t, []int{5, 4}, result, "Reverse then Limit(2) should yield [5 4]")
	})

	t.Run("ChunkEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := Chunk(Of(1, 2, 3, 4, 5, 6), 2).Limit(2).Collect()
		assert.Equal(t, [][]int{{1, 2}, {3, 4}}, result, "Chunk(size=2) then Limit(2) should yield [[1 2] [3 4]]")
	})
}

// TestMapToAndFlatMap tests type-changing transformations.
func TestMapToAndFlatMap(t *testing.T) {
	t.Parallel()
	t.Run("MapTo", func(t *testing.T) {
		t.Parallel()
		type Person struct {
			Name string
			Age  int
		}
		people := []Person{
			{Name: "Alice", Age: 30},
			{Name: "Bob", Age: 25},
		}

		names := MapTo(FromSlice(people), func(p Person) string {
			return p.Name
		}).Collect()

		assert.Equal(t, []string{"Alice", "Bob"}, names, "MapTo should change element type")
	})

	t.Run("FlatMap", func(t *testing.T) {
		t.Parallel()
		result := FlatMap(Of(1, 2, 3), func(n int) Stream[int] {
			return Of(n, n*10)
		}).Collect()
		assert.Equal(t, []int{1, 10, 2, 20, 3, 30}, result, "FlatMap should flatten results")
	})

	t.Run("FlatMapSeq", func(t *testing.T) {
		t.Parallel()
		result := FlatMapSeq(Of("ab", "cd"), func(s string) iter.Seq[rune] {
			return func(yield func(rune) bool) {
				for _, r := range s {
					if !yield(r) {
						return
					}
				}
			}
		}).Collect()
		assert.Equal(t, []rune{'a', 'b', 'c', 'd'}, result, "FlatMapSeq over \"ab\",\"cd\" should yield runes [a b c d]")
	})

	// Early termination tests
	t.Run("MapToEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := MapTo(Of(1, 2, 3, 4, 5), func(n int) string {
			return string(rune('a' + n - 1))
		}).Limit(2).Collect()
		assert.Equal(t, []string{"a", "b"}, result, "MapTo int->letter then Limit(2) should yield [\"a\" \"b\"]")
	})

	t.Run("FlatMapEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := FlatMap(Of(1, 2, 3), func(n int) Stream[int] {
			return Of(n*10, n*10+1)
		}).Limit(3).Collect()
		assert.Equal(t, []int{10, 11, 20}, result, "FlatMap n->[n*10,n*10+1] then Limit(3) should yield [10 11 20]")
	})

	t.Run("FlatMapSeqEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := FlatMapSeq(Of("abc", "def"), func(s string) iter.Seq[rune] {
			return func(yield func(rune) bool) {
				for _, r := range s {
					if !yield(r) {
						return
					}
				}
			}
		}).Limit(4).Collect()
		assert.Equal(t, []rune{'a', 'b', 'c', 'd'}, result, "FlatMapSeq over \"abc\",\"def\" then Limit(4) should yield [a b c d]")
	})
}

// TestZipOperations tests zip-related operations.
func TestZipOperations(t *testing.T) {
	t.Parallel()
	t.Run("Zip", func(t *testing.T) {
		t.Parallel()
		s1 := Of(1, 2, 3)
		s2 := Of("a", "b", "c")
		result := Zip(s1, s2).Collect()

		expected := []Pair[int, string]{
			{First: 1, Second: "a"},
			{First: 2, Second: "b"},
			{First: 3, Second: "c"},
		}
		assert.Equal(t, expected, result, "Zip should combine streams")
	})

	t.Run("ZipUnequalLength", func(t *testing.T) {
		t.Parallel()
		s1 := Of(1, 2, 3, 4)
		s2 := Of("a", "b")
		result := Zip(s1, s2).Collect()

		assert.Len(t, result, 2, "Zip should stop at shorter stream")
	})

	t.Run("ZipWithIndex", func(t *testing.T) {
		t.Parallel()
		result := ZipWithIndex(Of("a", "b", "c")).CollectPairs()

		expected := []Pair[int, string]{
			{First: 0, Second: "a"},
			{First: 1, Second: "b"},
			{First: 2, Second: "c"},
		}
		assert.Equal(t, expected, result, "ZipWithIndex should add indices")
	})

	// Early termination tests
	t.Run("ZipEarlyTermination", func(t *testing.T) {
		t.Parallel()
		s1 := Of(1, 2, 3, 4, 5)
		s2 := Of("a", "b", "c", "d", "e")
		result := Zip(s1, s2).Limit(2).Collect()
		assert.Len(t, result, 2, "Zip then Limit(2) should return exactly 2 pairs")
	})

	t.Run("ZipWithIndexEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := ZipWithIndex(Of("a", "b", "c", "d", "e")).Limit(2).CollectPairs()
		assert.Len(t, result, 2, "ZipWithIndex then Limit(2) should return first 2 indexed pairs")
	})

	t.Run("Unzip", func(t *testing.T) {
		t.Parallel()
		pairs := Of(
			NewPair(1, "a"),
			NewPair(2, "b"),
			NewPair(3, "c"),
		)
		firsts, seconds := Unzip(pairs)
		assert.Equal(t, []int{1, 2, 3}, firsts, "Unzip should extract first elements")
		assert.Equal(t, []string{"a", "b", "c"}, seconds, "Unzip should extract second elements")
	})

	t.Run("Zip3", func(t *testing.T) {
		t.Parallel()
		s1 := Of(1, 2)
		s2 := Of("a", "b")
		s3 := Of(1.0, 2.0)
		result := Zip3(s1, s2, s3).Collect()

		assert.Len(t, result, 2, "Zip3 should combine three streams")
		assert.Equal(t, 1, result[0].First, "First element should match")
		assert.Equal(t, "a", result[0].Second, "Second element should match")
		assert.Equal(t, 1.0, result[0].Third, "Third element should match")
	})
}

// TestDistinctOperations tests distinct-related operations.
func TestDistinctOperations(t *testing.T) {
	t.Parallel()
	t.Run("Distinct", func(t *testing.T) {
		t.Parallel()
		result := Distinct(Of(1, 2, 2, 3, 1, 3)).Collect()
		assert.Equal(t, []int{1, 2, 3}, result, "Distinct should remove duplicates")
	})

	t.Run("DistinctBy", func(t *testing.T) {
		t.Parallel()
		type Person struct {
			Name string
			Age  int
		}
		people := []Person{
			{Name: "Alice", Age: 30},
			{Name: "Bob", Age: 30},
			{Name: "Charlie", Age: 25},
		}

		result := DistinctBy(FromSlice(people), func(p Person) int {
			return p.Age
		}).Collect()

		assert.Len(t, result, 2, "DistinctBy should remove duplicates by key")
	})
}

// TestSortingOperations tests sorting-related operations.
func TestSortingOperations(t *testing.T) {
	t.Parallel()
	t.Run("SortedBy", func(t *testing.T) {
		t.Parallel()
		type Person struct {
			Name string
			Age  int
		}
		people := []Person{
			{Name: "Charlie", Age: 35},
			{Name: "Alice", Age: 30},
			{Name: "Bob", Age: 25},
		}

		result := SortedBy(FromSlice(people), func(p Person) int {
			return p.Age
		}).Collect()

		assert.Equal(t, "Bob", result[0].Name, "SortedBy should sort by key")
		assert.Equal(t, "Alice", result[1].Name, "SortedBy should sort by key")
		assert.Equal(t, "Charlie", result[2].Name, "SortedBy should sort by key")
	})

	t.Run("SortedByEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := SortedBy(Of(5, 3, 1, 4, 2), func(n int) int { return n }).Limit(2).Collect()
		assert.Equal(t, []int{1, 2}, result, "SortedBy(identity) then Limit(2) should yield [1 2]")
	})
}

// TestWindowAndInterleave tests window and interleave operations.
func TestWindowAndInterleave(t *testing.T) {
	t.Parallel()
	t.Run("Window", func(t *testing.T) {
		t.Parallel()
		result := Window(Of(1, 2, 3, 4, 5), 3).Collect()
		expected := [][]int{{1, 2, 3}, {2, 3, 4}, {3, 4, 5}}
		assert.Equal(t, expected, result, "Window should create sliding windows")
	})

	t.Run("WindowSmallerThanSize", func(t *testing.T) {
		t.Parallel()
		result := Window(Of(1, 2), 5).Collect()
		assert.Empty(t, result, "Window should be empty if size > stream length")
	})

	t.Run("Interleave", func(t *testing.T) {
		t.Parallel()
		s1 := Of(1, 3, 5)
		s2 := Of(2, 4)
		result := Interleave(s1, s2).Collect()
		assert.Equal(t, []int{1, 2, 3, 4, 5}, result, "Interleave should alternate elements")
	})

	// Early termination tests
	t.Run("WindowEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := Window(Of(1, 2, 3, 4, 5, 6, 7), 3).Limit(2).Collect()
		assert.Len(t, result, 2, "Window(size=3) then Limit(2) should return first 2 windows")
	})

	t.Run("InterleaveEarlyTermination", func(t *testing.T) {
		t.Parallel()
		s1 := Of(1, 3, 5, 7, 9)
		s2 := Of(2, 4, 6, 8, 10)
		result := Interleave(s1, s2).Limit(4).Collect()
		assert.Len(t, result, 4, "Interleave then Limit(4) should return exactly 4 elements")
	})

	t.Run("InterleaveEarlyTerminationS1Side", func(t *testing.T) {
		t.Parallel()
		s1 := Of(1, 3, 5, 7, 9, 11, 13)
		s2 := Of(2, 4, 6)
		result := Interleave(s1, s2).Limit(5).Collect()
		assert.Len(t, result, 5, "Interleave(s2 shorter) then Limit(5) should return 5 elements")
	})
}

// TestEarlyTermination tests that streams handle early termination correctly.
func TestEarlyTermination(t *testing.T) {
	t.Parallel()
	t.Run("FilterWithLimit", func(t *testing.T) {
		t.Parallel()
		// Filter should stop processing once limit is reached
		result := Of(1, 2, 3, 4, 5, 6, 7, 8, 9, 10).
			Filter(func(n int) bool { return n%2 == 0 }).
			Limit(2).
			Collect()
		assert.Equal(t, []int{2, 4}, result, "Filter evens then Limit(2) should yield [2 4]")
	})

	t.Run("MapWithLimit", func(t *testing.T) {
		t.Parallel()
		result := Of(1, 2, 3, 4, 5).
			Map(func(n int) int { return n * 2 }).
			Limit(2).
			Collect()
		assert.Equal(t, []int{2, 4}, result, "Map(*2) then Limit(2) should yield [2 4]")
	})

	t.Run("GenerateWithLimit", func(t *testing.T) {
		t.Parallel()
		counter := 0
		result := Generate(func() int {
			counter++
			return counter
		}).Limit(3).Collect()

		assert.Equal(t, []int{1, 2, 3}, result, "Generate(counter) then Limit(3) should yield [1 2 3]")
	})

	t.Run("InfiniteIterateWithTakeWhile", func(t *testing.T) {
		t.Parallel()
		result := Iterate(1, func(n int) int { return n + 1 }).
			TakeWhile(func(n int) bool { return n <= 5 }).
			Collect()
		assert.Equal(t, []int{1, 2, 3, 4, 5}, result, "Iterate(+1) then TakeWhile(n<=5) should yield [1 2 3 4 5]")
	})

	t.Run("ChainedOperationsWithLimit", func(t *testing.T) {
		t.Parallel()
		result := Range(1, 100).
			Filter(func(n int) bool { return n%3 == 0 }).
			Map(func(n int) int { return n * 2 }).
			Limit(3).
			Collect()
		assert.Equal(t, []int{6, 12, 18}, result, "Range(1,100)->Filter(n%3==0)->Map(*2) then Limit(3) should yield [6 12 18]")
	})

	t.Run("FindFirstStopsEarly", func(t *testing.T) {
		t.Parallel()
		// FindFirst should stop as soon as it finds a match
		result := Range(1, 1000000).
			FindFirst(func(n int) bool { return n > 100 })
		assert.True(t, result.IsPresent(), "FindFirst should find element")
		assert.Equal(t, 101, result.Get(), "FindFirst should return first match")
	})

	t.Run("AnyMatchStopsEarly", func(t *testing.T) {
		t.Parallel()
		result := Range(1, 1000000).
			AnyMatch(func(n int) bool { return n == 50 })
		assert.True(t, result, "AnyMatch should find match early")
	})

	// Additional early termination tests for previously uncovered code paths
	t.Run("FromChannelWithLimit", func(t *testing.T) {
		t.Parallel()
		ch := make(chan int, 10)
		for i := 1; i <= 10; i++ {
			ch <- i
		}
		close(ch)

		result := FromChannel(ch).Limit(3).Collect()
		assert.Equal(t, []int{1, 2, 3}, result, "FromChannel(1..10) then Limit(3) should yield [1 2 3]")
	})

	t.Run("RangeClosedWithLimit", func(t *testing.T) {
		t.Parallel()
		result := RangeClosed(1, 100).Limit(3).Collect()
		assert.Equal(t, []int{1, 2, 3}, result, "RangeClosed(1,100) then Limit(3) should yield [1 2 3]")
	})

	t.Run("ConcatWithLimit", func(t *testing.T) {
		t.Parallel()
		s1 := Of(1, 2, 3)
		s2 := Of(4, 5, 6)
		result := Concat(s1, s2).Limit(2).Collect()
		assert.Equal(t, []int{1, 2}, result, "Concat([1 2 3],[4 5 6]) then Limit(2) should yield [1 2]")
	})

	t.Run("RepeatWithLimit", func(t *testing.T) {
		t.Parallel()
		result := Repeat("x", 100).Limit(3).Collect()
		assert.Equal(t, []string{"x", "x", "x"}, result, "Repeat(\"x\",100) then Limit(3) should yield [x x x]")
	})

	t.Run("RepeatForeverWithLimit", func(t *testing.T) {
		t.Parallel()
		result := RepeatForever("x").Limit(3).Collect()
		assert.Equal(t, []string{"x", "x", "x"}, result, "RepeatForever(\"x\") then Limit(3) should yield [x x x]")
	})
}

// TestNewStreamOperations tests newly added stream operations.
func TestNewStreamOperations(t *testing.T) {
	t.Parallel()
	t.Run("Scan", func(t *testing.T) {
		t.Parallel()
		// Running sum using Scan
		result := Scan(Of(1, 2, 3, 4, 5), 0, func(acc, v int) int { return acc + v }).Collect()
		assert.Equal(t, []int{1, 3, 6, 10, 15}, result, "Scan should produce running totals")

		// Running product
		result2 := Scan(Of(1, 2, 3, 4), 1, func(acc, v int) int { return acc * v }).Collect()
		assert.Equal(t, []int{1, 2, 6, 24}, result2, "Scan(product) over [1 2 3 4] should yield [1 2 6 24]")

		// Empty stream
		result3 := Scan(Empty[int](), 0, func(acc, v int) int { return acc + v }).Collect()
		assert.Empty(t, result3, "Scan on empty stream should be empty")
	})

	t.Run("Step", func(t *testing.T) {
		t.Parallel()
		// Every 2nd element
		result := Of(1, 2, 3, 4, 5, 6, 7, 8, 9, 10).Step(2).Collect()
		assert.Equal(t, []int{1, 3, 5, 7, 9}, result, "Step(2) should return every 2nd element")

		// Every 3rd element
		result2 := Of(1, 2, 3, 4, 5, 6, 7, 8, 9).Step(3).Collect()
		assert.Equal(t, []int{1, 4, 7}, result2, "Step(3) should return every 3rd element")

		// Step 1 should return all elements
		result3 := Of(1, 2, 3).Step(1).Collect()
		assert.Equal(t, []int{1, 2, 3}, result3, "Step(1) should return all elements")

		// Step 0 should return all elements
		result4 := Of(1, 2, 3).Step(0).Collect()
		assert.Equal(t, []int{1, 2, 3}, result4, "Step(0) should return all elements")
	})

	t.Run("DistinctUntilChanged", func(t *testing.T) {
		t.Parallel()
		result := DistinctUntilChanged(Of(1, 1, 2, 2, 2, 3, 1, 1)).Collect()
		assert.Equal(t, []int{1, 2, 3, 1}, result, "DistinctUntilChanged should remove consecutive duplicates")

		// All same
		result2 := DistinctUntilChanged(Of(1, 1, 1, 1)).Collect()
		assert.Equal(t, []int{1}, result2, "DistinctUntilChanged should collapse all same values to one")

		// All different
		result3 := DistinctUntilChanged(Of(1, 2, 3, 4)).Collect()
		assert.Equal(t, []int{1, 2, 3, 4}, result3, "DistinctUntilChanged should keep all different values")

		// Empty stream
		result4 := DistinctUntilChanged(Empty[int]()).Collect()
		assert.Empty(t, result4, "DistinctUntilChanged on empty stream should be empty")
	})

	t.Run("DistinctUntilChangedBy", func(t *testing.T) {
		t.Parallel()
		type item struct {
			id   int
			name string
		}
		items := Of(
			item{1, "a"}, item{1, "b"}, item{2, "c"}, item{2, "d"}, item{1, "e"},
		)
		result := DistinctUntilChangedBy(items, func(a, b item) bool { return a.id == b.id }).Collect()
		assert.Len(t, result, 3, "DistinctUntilChangedBy should remove consecutive duplicates by key")
		assert.Equal(t, 1, result[0].id, "DistinctUntilChangedBy result[0].id should be 1")
		assert.Equal(t, 2, result[1].id, "DistinctUntilChangedBy result[1].id should be 2")
		assert.Equal(t, 1, result[2].id, "DistinctUntilChangedBy result[2].id should be 1")
	})

	t.Run("TakeLast", func(t *testing.T) {
		t.Parallel()
		result := Of(1, 2, 3, 4, 5).TakeLast(3).Collect()
		assert.Equal(t, []int{3, 4, 5}, result, "TakeLast should return last 3 elements")

		// TakeLast more than available
		result2 := Of(1, 2).TakeLast(5).Collect()
		assert.Equal(t, []int{1, 2}, result2, "TakeLast should return all if n > length")

		// TakeLast 0
		result3 := Of(1, 2, 3).TakeLast(0).Collect()
		assert.Empty(t, result3, "TakeLast(0) should return empty")

		// TakeLast negative
		result4 := Of(1, 2, 3).TakeLast(-1).Collect()
		assert.Empty(t, result4, "TakeLast(-1) should return empty")

		// Empty stream
		result5 := Empty[int]().TakeLast(3).Collect()
		assert.Empty(t, result5, "TakeLast on empty stream should be empty")
	})

	t.Run("DropLast", func(t *testing.T) {
		t.Parallel()
		result := Of(1, 2, 3, 4, 5).DropLast(2).Collect()
		assert.Equal(t, []int{1, 2, 3}, result, "DropLast should remove last 2 elements")

		// DropLast more than available
		result2 := Of(1, 2).DropLast(5).Collect()
		assert.Empty(t, result2, "DropLast should return empty if n >= length")

		// DropLast 0
		result3 := Of(1, 2, 3).DropLast(0).Collect()
		assert.Equal(t, []int{1, 2, 3}, result3, "DropLast(0) should return all elements")
	})

	t.Run("WindowWithStep", func(t *testing.T) {
		t.Parallel()
		// Step 1 (sliding window)
		result := WindowWithStep(Of(1, 2, 3, 4, 5), 3, 1, false).Collect()
		assert.Equal(t, [][]int{{1, 2, 3}, {2, 3, 4}, {3, 4, 5}}, result, "WindowWithStep should create sliding windows")

		// Step 2
		result2 := WindowWithStep(Of(1, 2, 3, 4, 5, 6), 2, 2, false).Collect()
		assert.Equal(t, [][]int{{1, 2}, {3, 4}, {5, 6}}, result2, "WindowWithStep with step=size should create chunks")

		// With partial window allowed
		result3 := WindowWithStep(Of(1, 2, 3, 4, 5), 3, 2, true).Collect()
		assert.Equal(t, [][]int{{1, 2, 3}, {3, 4, 5}, {5}}, result3, "WindowWithStep with allowPartial should include partial")

		// Without partial window
		result4 := WindowWithStep(Of(1, 2, 3, 4, 5), 3, 2, false).Collect()
		assert.Equal(t, [][]int{{1, 2, 3}, {3, 4, 5}}, result4, "WindowWithStep without allowPartial should exclude partial")
	})

	t.Run("Pairwise", func(t *testing.T) {
		t.Parallel()
		result := Pairwise(Of(1, 2, 3, 4)).Collect()
		assert.Len(t, result, 3, "Pairwise should return n-1 pairs")
		assert.Equal(t, Pair[int, int]{1, 2}, result[0], "Pairwise result[0] should be (1,2)")
		assert.Equal(t, Pair[int, int]{2, 3}, result[1], "Pairwise result[1] should be (2,3)")
		assert.Equal(t, Pair[int, int]{3, 4}, result[2], "Pairwise result[2] should be (3,4)")

		// Single element
		result2 := Pairwise(Of(1)).Collect()
		assert.Empty(t, result2, "Pairwise with single element should be empty")

		// Empty stream
		result3 := Pairwise(Empty[int]()).Collect()
		assert.Empty(t, result3, "Pairwise on empty stream should be empty")
	})

	t.Run("Triples", func(t *testing.T) {
		t.Parallel()
		result := Triples(Of(1, 2, 3, 4, 5)).Collect()
		assert.Len(t, result, 3, "Triples should return n-2 triples")
		assert.Equal(t, Triple[int, int, int]{1, 2, 3}, result[0], "Triples result[0] should be (1,2,3)")
		assert.Equal(t, Triple[int, int, int]{2, 3, 4}, result[1], "Triples result[1] should be (2,3,4)")
		assert.Equal(t, Triple[int, int, int]{3, 4, 5}, result[2], "Triples result[2] should be (3,4,5)")
	})

	t.Run("SortedStable", func(t *testing.T) {
		t.Parallel()
		type item struct {
			key   int
			order int // original order
		}
		items := []item{{1, 1}, {2, 2}, {1, 3}, {2, 4}, {1, 5}}
		result := Of(items...).SortedStable(func(a, b item) int {
			return a.key - b.key
		}).Collect()

		// Items with key=1 should maintain their relative order
		key1Items := []item{}
		for _, it := range result {
			if it.key == 1 {
				key1Items = append(key1Items, it)
			}
		}
		assert.Equal(t, []item{{1, 1}, {1, 3}, {1, 5}}, key1Items, "SortedStable should maintain relative order")
	})

	t.Run("Flatten", func(t *testing.T) {
		t.Parallel()
		nested := Of([]int{1, 2}, []int{3, 4, 5}, []int{6})
		result := Flatten(nested).Collect()
		assert.Equal(t, []int{1, 2, 3, 4, 5, 6}, result, "Flatten should flatten nested slices")

		// Empty inner slices
		nested2 := Of([]int{1}, []int{}, []int{2, 3})
		result2 := Flatten(nested2).Collect()
		assert.Equal(t, []int{1, 2, 3}, result2, "Flatten should handle empty inner slices")
	})

	t.Run("Intersperse", func(t *testing.T) {
		t.Parallel()
		result := Of(1, 2, 3).Intersperse(0).Collect()
		assert.Equal(t, []int{1, 0, 2, 0, 3}, result, "Intersperse should insert separator between elements")

		// Single element
		result2 := Of(1).Intersperse(0).Collect()
		assert.Equal(t, []int{1}, result2, "Intersperse with single element should not add separator")

		// Empty stream
		result3 := Empty[int]().Intersperse(0).Collect()
		assert.Empty(t, result3, "Intersperse on empty stream should be empty")
	})
}

// TestEdgeCases tests boundary conditions and edge cases.
func TestEdgeCases(t *testing.T) {
	t.Parallel()
	t.Run("TakeLastLargeN", func(t *testing.T) {
		t.Parallel()
		// n much larger than input
		result := Of(1, 2, 3).TakeLast(1000).Collect()
		assert.Equal(t, []int{1, 2, 3}, result, "TakeLast with n > len should return all elements")

		// Large input with small n
		input := Range(1, 10001).Collect() // 1 to 10000
		result2 := FromSlice(input).TakeLast(5).Collect()
		assert.Equal(t, []int{9996, 9997, 9998, 9999, 10000}, result2, "TakeLast should correctly return last 5 of large input")
	})

	t.Run("DropLastLargeN", func(t *testing.T) {
		t.Parallel()
		// n much larger than input
		result := Of(1, 2, 3).DropLast(1000).Collect()
		assert.Empty(t, result, "DropLast with n > len should return empty")

		// Large input with small n
		input := Range(1, 101).Collect() // 1 to 100
		result2 := FromSlice(input).DropLast(3).Collect()
		assert.Len(t, result2, 97, "DropLast should drop last 3 elements")
		assert.Equal(t, 1, result2[0], "First element should be 1")
		assert.Equal(t, 97, result2[96], "Last element should be 97")
	})

	t.Run("TakeLastRingBufferCorrectness", func(t *testing.T) {
		t.Parallel()
		// Test that ring buffer correctly wraps around
		result := Range(1, 11).TakeLast(3).Collect() // 1-10, take last 3
		assert.Equal(t, []int{8, 9, 10}, result, "Ring buffer should correctly track last elements")

		// Edge case: n equals input length
		result2 := Of(1, 2, 3, 4, 5).TakeLast(5).Collect()
		assert.Equal(t, []int{1, 2, 3, 4, 5}, result2, "TakeLast with n == len should return all")
	})

	t.Run("DistinctUntilChangedByAlwaysTrue", func(t *testing.T) {
		t.Parallel()
		// eq always returns true: only first element should be yielded
		result := DistinctUntilChangedBy(Of(1, 2, 3, 4, 5), func(_, _ int) bool { return true }).Collect()
		assert.Equal(t, []int{1}, result, "DistinctUntilChangedBy with always-true eq should yield only first")
	})

	t.Run("DistinctUntilChangedByAlwaysFalse", func(t *testing.T) {
		t.Parallel()
		// eq always returns false: all elements should be yielded
		result := DistinctUntilChangedBy(Of(1, 1, 1, 1, 1), func(_, _ int) bool { return false }).Collect()
		assert.Equal(t, []int{1, 1, 1, 1, 1}, result, "DistinctUntilChangedBy with always-false eq should yield all")
	})

	t.Run("WindowWithStepStepGreaterThanSize", func(t *testing.T) {
		t.Parallel()
		// Step > Size: windows should not overlap, with gaps
		result := WindowWithStep(Of(1, 2, 3, 4, 5, 6, 7, 8, 9, 10), 2, 4, false).Collect()
		assert.Equal(t, [][]int{{1, 2}, {5, 6}, {9, 10}}, result, "Windows with step>size should skip elements")

		// Step > Size with allowPartial
		result2 := WindowWithStep(Of(1, 2, 3, 4, 5, 6, 7), 2, 3, true).Collect()
		assert.Equal(t, [][]int{{1, 2}, {4, 5}, {7}}, result2, "allowPartial should yield partial window at end")
	})

	t.Run("WindowWithStepStepEqualsSizeNoPartial", func(t *testing.T) {
		t.Parallel()
		// Step == Size: non-overlapping chunks
		result := WindowWithStep(Of(1, 2, 3, 4, 5), 2, 2, false).Collect()
		assert.Equal(t, [][]int{{1, 2}, {3, 4}}, result, "Step==Size without partial should act like Chunk")
	})

	t.Run("LimitSkipEdgeCases", func(t *testing.T) {
		t.Parallel()
		// Limit(0) returns empty
		result := Of(1, 2, 3).Limit(0).Collect()
		assert.Empty(t, result, "Limit(0) should return empty")

		// Skip(0) returns all
		result2 := Of(1, 2, 3).Skip(0).Collect()
		assert.Equal(t, []int{1, 2, 3}, result2, "Skip(0) should return all elements")

		// Negative values
		result3 := Of(1, 2, 3).Limit(-1).Collect()
		assert.Empty(t, result3, "Limit(-1) should return empty")
	})

	t.Run("StepEdgeCases", func(t *testing.T) {
		t.Parallel()
		// Step(1) returns all
		result := Of(1, 2, 3, 4, 5).Step(1).Collect()
		assert.Equal(t, []int{1, 2, 3, 4, 5}, result, "Step(1) should return all")

		// Step(0) returns all
		result2 := Of(1, 2, 3).Step(0).Collect()
		assert.Equal(t, []int{1, 2, 3}, result2, "Step(0) should return all")

		// Step(-1) returns all
		result3 := Of(1, 2, 3).Step(-1).Collect()
		assert.Equal(t, []int{1, 2, 3}, result3, "Step(-1) should return all")

		// Step larger than input
		result4 := Of(1, 2, 3).Step(10).Collect()
		assert.Equal(t, []int{1}, result4, "Step(10) on 3 elements should return only first")
	})
}

// TestRepeatForever tests RepeatForever function.
func TestRepeatForever(t *testing.T) {
	t.Parallel()
	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		result := RepeatForever("x").Limit(5).Collect()
		assert.Equal(t, []string{"x", "x", "x", "x", "x"}, result, "RepeatForever(\"x\") Limit(5) should yield five \"x\" values")
	})

	t.Run("WithInt", func(t *testing.T) {
		t.Parallel()
		result := RepeatForever(42).Limit(3).Collect()
		assert.Equal(t, []int{42, 42, 42}, result, "RepeatForever(42) Limit(3) should yield [42 42 42]")
	})

	t.Run("EarlyTermination", func(t *testing.T) {
		t.Parallel()
		// Test that early termination works
		count := 0
		RepeatForever(1).Limit(10).ForEach(func(_ int) {
			count++
		})
		assert.Equal(t, 10, count, "Limit(10) should invoke action exactly 10 times")
	})
}

// TestSortedStableBy tests SortedStableBy function.
func TestSortedStableBy(t *testing.T) {
	t.Parallel()
	type Item struct {
		Name  string
		Order int
	}

	t.Run("StableSort", func(t *testing.T) {
		t.Parallel()
		items := []Item{
			{Name: "a", Order: 2},
			{Name: "b", Order: 1},
			{Name: "c", Order: 2},
			{Name: "d", Order: 1},
		}
		result := SortedStableBy(FromSlice(items), func(i Item) int { return i.Order }).Collect()

		// Items with same Order should maintain relative order
		assert.Equal(t, 1, result[0].Order, "First two items should have Order==1")
		assert.Equal(t, 1, result[1].Order, "First two items should have Order==1")
		assert.Equal(t, "b", result[0].Name, "Stable sort should keep 'b' before 'd' for equal keys") // b came before d
		assert.Equal(t, "d", result[1].Name, "Stable sort should keep 'd' after 'b' for equal keys")

		assert.Equal(t, 2, result[2].Order, "Last two items should have Order==2")
		assert.Equal(t, 2, result[3].Order, "Last two items should have Order==2")
		assert.Equal(t, "a", result[2].Name, "Stable sort should keep 'a' before 'c' for equal keys") // a came before c
		assert.Equal(t, "c", result[3].Name, "Stable sort should keep 'c' after 'a' for equal keys")
	})

	t.Run("Empty", func(t *testing.T) {
		t.Parallel()
		result := SortedStableBy(Empty[int](), func(i int) int { return i }).Collect()
		assert.Empty(t, result, "SortedStableBy on empty stream should return empty")
	})

	t.Run("EarlyTermination", func(t *testing.T) {
		t.Parallel()
		items := []Item{
			{Name: "a", Order: 3},
			{Name: "b", Order: 1},
			{Name: "c", Order: 2},
		}
		result := SortedStableBy(FromSlice(items), func(i Item) int { return i.Order }).Limit(2).Collect()
		assert.Len(t, result, 2, "SortedStableBy then Limit(2) should return 2 items")
		assert.Equal(t, "b", result[0].Name, "SortedStableBy should order by key before limiting")
		assert.Equal(t, "c", result[1].Name, "SortedStableBy should order by key before limiting")
	})
}

// TestFlattenSeq tests FlattenSeq function.
func TestFlattenSeq(t *testing.T) {
	t.Parallel()
	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		seq1 := Of(1, 2).Seq()
		seq2 := Of(3, 4).Seq()
		seq3 := Of(5, 6).Seq()

		result := FlattenSeq(Of(seq1, seq2, seq3)).Collect()
		assert.Equal(t, []int{1, 2, 3, 4, 5, 6}, result, "FlattenSeq should concatenate all inner sequences")
	})

	t.Run("Empty", func(t *testing.T) {
		t.Parallel()
		result := FlattenSeq(Empty[iter.Seq[int]]()).Collect()
		assert.Empty(t, result, "FlattenSeq on empty outer sequence should be empty")
	})

	t.Run("WithEmptyInner", func(t *testing.T) {
		t.Parallel()
		seq1 := Of(1, 2).Seq()
		seq2 := Empty[int]().Seq()
		seq3 := Of(3, 4).Seq()

		result := FlattenSeq(Of(seq1, seq2, seq3)).Collect()
		assert.Equal(t, []int{1, 2, 3, 4}, result, "FlattenSeq should skip empty inner sequences")
	})

	t.Run("EarlyTermination", func(t *testing.T) {
		t.Parallel()
		seq1 := Of(1, 2, 3).Seq()
		seq2 := Of(4, 5, 6).Seq()

		result := FlattenSeq(Of(seq1, seq2)).Limit(4).Collect()
		assert.Equal(t, []int{1, 2, 3, 4}, result, "FlattenSeq then Limit(4) should yield [1 2 3 4]")
	})
}

// TestCartesianSelf tests CartesianSelf function.
func TestCartesianSelf(t *testing.T) {
	t.Parallel()
	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		result := CartesianSelf(Of(1, 2)).Collect()
		expected := []Pair[int, int]{
			{First: 1, Second: 1},
			{First: 1, Second: 2},
			{First: 2, Second: 1},
			{First: 2, Second: 2},
		}
		assert.Equal(t, expected, result, "CartesianSelf should produce all pairs of the same stream")
	})

	t.Run("SingleElement", func(t *testing.T) {
		t.Parallel()
		result := CartesianSelf(Of("x")).Collect()
		assert.Len(t, result, 1, "CartesianSelf with single element should produce 1 pair")
		assert.Equal(t, "x", result[0].First, "CartesianSelf should use the element as First")
		assert.Equal(t, "x", result[0].Second, "CartesianSelf should use the element as Second")
	})

	t.Run("Empty", func(t *testing.T) {
		t.Parallel()
		result := CartesianSelf(Empty[int]()).Collect()
		assert.Empty(t, result, "CartesianSelf on empty stream should be empty")
	})

	t.Run("EarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := CartesianSelf(Of(1, 2, 3)).Limit(5).Collect()
		assert.Len(t, result, 5, "CartesianSelf then Limit(5) should return 5 pairs")
	})
}

// TestEarlyTerminationAdditional tests additional early termination paths.
func TestEarlyTerminationAdditional(t *testing.T) {
	t.Parallel()
	t.Run("DistinctEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := Distinct(Of(1, 2, 2, 3, 3, 4)).Limit(2).Collect()
		assert.Equal(t, []int{1, 2}, result, "Distinct then Limit(2) should yield [1 2]")
	})

	t.Run("DistinctByEarlyTermination", func(t *testing.T) {
		t.Parallel()
		type Item struct {
			Key   int
			Value string
		}
		items := []Item{{1, "a"}, {1, "b"}, {2, "c"}, {2, "d"}, {3, "e"}}
		result := DistinctBy(FromSlice(items), func(i Item) int { return i.Key }).Limit(2).Collect()
		assert.Len(t, result, 2, "DistinctBy(key) then Limit(2) should return 2 items")
	})

	t.Run("ScanEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := Scan(Of(1, 2, 3, 4, 5), 0, func(acc, v int) int { return acc + v }).Limit(3).Collect()
		assert.Equal(t, []int{1, 3, 6}, result, "Scan(sum) then Limit(3) should yield [1 3 6]")
	})

	t.Run("StepEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := Of(1, 2, 3, 4, 5, 6, 7, 8, 9, 10).Step(2).Limit(3).Collect()
		assert.Equal(t, []int{1, 3, 5}, result, "Step(2) then Limit(3) should yield [1 3 5]")
	})

	t.Run("DistinctUntilChangedEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := DistinctUntilChanged(Of(1, 1, 2, 2, 3, 3, 4, 4)).Limit(2).Collect()
		assert.Equal(t, []int{1, 2}, result, "DistinctUntilChanged then Limit(2) should yield [1 2]")
	})

	t.Run("DistinctUntilChangedByEarlyTermination", func(t *testing.T) {
		t.Parallel()
		type Item struct {
			Key   int
			Value string
		}
		items := []Item{{1, "a"}, {1, "b"}, {2, "c"}, {2, "d"}, {3, "e"}}
		result := DistinctUntilChangedBy(FromSlice(items), func(a, b Item) bool { return a.Key == b.Key }).Limit(2).Collect()
		assert.Len(t, result, 2, "DistinctUntilChangedBy(key) then Limit(2) should return 2 items")
	})

	t.Run("TakeLastEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := Of(1, 2, 3, 4, 5).TakeLast(3).Limit(2).Collect()
		assert.Equal(t, []int{3, 4}, result, "TakeLast(3) then Limit(2) should yield [3 4]")
	})

	t.Run("DropLastEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := Of(1, 2, 3, 4, 5).DropLast(2).Limit(2).Collect()
		assert.Equal(t, []int{1, 2}, result, "DropLast(2) then Limit(2) should yield [1 2]")
	})

	t.Run("WindowWithStepInvalidParams", func(t *testing.T) {
		t.Parallel()
		result := WindowWithStep(Of(1, 2, 3, 4, 5), 0, 1, false).Collect()
		assert.Empty(t, result, "WindowWithStep with size=0 should be empty")

		result2 := WindowWithStep(Of(1, 2, 3, 4, 5), 2, 0, false).Collect()
		assert.Empty(t, result2, "WindowWithStep with step=0 should be empty")
	})

	t.Run("WindowWithStepEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := WindowWithStep(Of(1, 2, 3, 4, 5, 6, 7, 8), 2, 1, false).Limit(3).Collect()
		assert.Len(t, result, 3, "WindowWithStep(size=2,step=1) then Limit(3) should return 3 windows")
	})

	t.Run("PairwiseEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := Pairwise(Of(1, 2, 3, 4, 5)).Limit(2).Collect()
		assert.Len(t, result, 2, "Pairwise then Limit(2) should return 2 pairs")
	})

	t.Run("TriplesEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := Triples(Of(1, 2, 3, 4, 5, 6)).Limit(2).Collect()
		assert.Len(t, result, 2, "Triples then Limit(2) should return 2 triples")
	})

	t.Run("SortedStableEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := Of(5, 4, 3, 2, 1).SortedStable(func(a, b int) int { return a - b }).Limit(2).Collect()
		assert.Equal(t, []int{1, 2}, result, "SortedStable(asc) then Limit(2) should yield [1 2]")
	})

	t.Run("FlattenEarlyTermination", func(t *testing.T) {
		t.Parallel()
		nested := Of([]int{1, 2}, []int{3, 4}, []int{5, 6})
		result := Flatten(nested).Limit(3).Collect()
		assert.Equal(t, []int{1, 2, 3}, result, "Flatten then Limit(3) should yield [1 2 3]")
	})

	t.Run("IntersperseEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := Of(1, 2, 3, 4, 5).Intersperse(0).Limit(5).Collect()
		assert.Equal(t, []int{1, 0, 2, 0, 3}, result, "Intersperse(0) then Limit(5) should yield [1 0 2 0 3]")
	})

	t.Run("IntersperseEarlyTerminationOnSeparator", func(t *testing.T) {
		t.Parallel()
		result := Of(1, 2, 3, 4, 5).Intersperse(0).Limit(4).Collect()
		assert.Equal(t, []int{1, 0, 2, 0}, result, "Intersperse(0) then Limit(4) should yield [1 0 2 0]")
	})
}

// TestMergeSortedOperations tests MergeSorted operations.
func TestMergeSortedOperations(t *testing.T) {
	t.Parallel()
	t.Run("MergeSortedBasic", func(t *testing.T) {
		t.Parallel()
		s1 := Of(1, 3, 5)
		s2 := Of(2, 4, 6)
		result := MergeSorted(s1, s2, func(a, b int) int { return a - b }).Collect()
		assert.Equal(t, []int{1, 2, 3, 4, 5, 6}, result, "MergeSorted should merge two sorted streams")
	})

	t.Run("MergeSortedDrainS1", func(t *testing.T) {
		t.Parallel()
		s1 := Of(1, 3, 5, 7, 9)
		s2 := Of(2, 4)
		result := MergeSorted(s1, s2, func(a, b int) int { return a - b }).Collect()
		assert.Equal(t, []int{1, 2, 3, 4, 5, 7, 9}, result, "MergeSorted should continue with remaining elements from longer s1")
	})

	t.Run("MergeSortedDrainS2", func(t *testing.T) {
		t.Parallel()
		s1 := Of(1, 3)
		s2 := Of(2, 4, 6, 8, 10)
		result := MergeSorted(s1, s2, func(a, b int) int { return a - b }).Collect()
		assert.Equal(t, []int{1, 2, 3, 4, 6, 8, 10}, result, "MergeSorted should continue with remaining elements from longer s2")
	})

	t.Run("MergeSortedDrainS1EarlyTermination", func(t *testing.T) {
		t.Parallel()
		s1 := Of(1, 3, 5, 7, 9)
		s2 := Of(2)
		result := MergeSorted(s1, s2, func(a, b int) int { return a - b }).Limit(4).Collect()
		assert.Len(t, result, 4, "MergeSorted(s1 longer) then Limit(4) should return 4 elements")
	})

	t.Run("MergeSortedDrainS2EarlyTermination", func(t *testing.T) {
		t.Parallel()
		s1 := Of(1)
		s2 := Of(2, 4, 6, 8, 10)
		result := MergeSorted(s1, s2, func(a, b int) int { return a - b }).Limit(3).Collect()
		assert.Len(t, result, 3, "MergeSorted(s2 longer) then Limit(3) should return 3 elements")
	})

	t.Run("MergeSortedNEmpty", func(t *testing.T) {
		t.Parallel()
		result := MergeSortedN(func(a, b int) int { return a - b }).Collect()
		assert.Empty(t, result, "MergeSortedN with no streams should be empty")
	})

	t.Run("MergeSortedNSingle", func(t *testing.T) {
		t.Parallel()
		result := MergeSortedN(func(a, b int) int { return a - b }, Of(1, 2, 3)).Collect()
		assert.Equal(t, []int{1, 2, 3}, result, "MergeSortedN with single stream should return all elements")
	})

	t.Run("MergeSortedNMultiple", func(t *testing.T) {
		t.Parallel()
		result := MergeSortedN(func(a, b int) int { return a - b }, Of(1, 4), Of(2, 5), Of(3, 6)).Collect()
		assert.Equal(t, []int{1, 2, 3, 4, 5, 6}, result, "MergeSortedN should merge multiple sorted streams")
	})

	t.Run("MergeSortedNHeapEmpty", func(t *testing.T) {
		t.Parallel()
		result := MergeSortedNHeap(func(a, b int) int { return a - b }).Collect()
		assert.Empty(t, result, "MergeSortedNHeap with no streams should be empty")
	})

	t.Run("MergeSortedNHeapSingle", func(t *testing.T) {
		t.Parallel()
		result := MergeSortedNHeap(func(a, b int) int { return a - b }, Of(1, 2, 3)).Collect()
		assert.Equal(t, []int{1, 2, 3}, result, "MergeSortedNHeap with single stream should return all elements")
	})

	t.Run("MergeSortedNHeapMultiple", func(t *testing.T) {
		t.Parallel()
		result := MergeSortedNHeap(func(a, b int) int { return a - b }, Of(1, 4, 7), Of(2, 5, 8), Of(3, 6, 9)).Collect()
		assert.Equal(t, []int{1, 2, 3, 4, 5, 6, 7, 8, 9}, result, "MergeSortedNHeap should merge and sort across streams")
	})

	t.Run("MergeSortedNHeapEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := MergeSortedNHeap(func(a, b int) int { return a - b }, Of(1, 4, 7), Of(2, 5, 8), Of(3, 6, 9)).Limit(5).Collect()
		assert.Len(t, result, 5, "MergeSortedNHeap then Limit(5) should return first 5 elements")
	})

	t.Run("MergeSortedNHeapManyStreams", func(t *testing.T) {
		t.Parallel()
		// Test with many streams to exercise heap operations
		streams := make([]Stream[int], 10)
		for i := range 10 {
			streams[i] = Of(i, i+10, i+20)
		}
		result := MergeSortedNHeap(func(a, b int) int { return a - b }, streams...).Collect()
		assert.Len(t, result, 30, "MergeSortedNHeap with 10 streams of 3 elements should produce 30 outputs")
		// Verify sorted
		for i := 1; i < len(result); i++ {
			assert.LessOrEqual(t, result[i-1], result[i], "MergeSortedNHeap result should be non-decreasing")
		}
	})
}

// TestZipLongestOperations tests ZipLongest operations.
func TestZipLongestOperations(t *testing.T) {
	t.Parallel()
	t.Run("ZipLongestSameLength", func(t *testing.T) {
		t.Parallel()
		result := ZipLongest(Of(1, 2, 3), Of("a", "b", "c")).Collect()
		assert.Len(t, result, 3, "ZipLongest with same-length inputs should produce 3 pairs")
		assert.Equal(t, 1, result[0].First.Get(), "ZipLongest first pair should have First=1")
		assert.Equal(t, "a", result[0].Second.Get(), "ZipLongest first pair should have Second=\"a\"")
	})

	t.Run("ZipLongestFirstLonger", func(t *testing.T) {
		t.Parallel()
		result := ZipLongest(Of(1, 2, 3, 4, 5), Of("a", "b")).Collect()
		assert.Len(t, result, 5, "ZipLongest where first is longer should produce 5 pairs")
		assert.True(t, result[2].First.IsPresent(), "ZipLongest third pair should have First present when first stream longer")
		assert.True(t, result[2].Second.IsEmpty(), "ZipLongest third pair should have Second empty when second stream exhausted")
	})

	t.Run("ZipLongestSecondLonger", func(t *testing.T) {
		t.Parallel()
		result := ZipLongest(Of(1, 2), Of("a", "b", "c", "d", "e")).Collect()
		assert.Len(t, result, 5, "ZipLongest where second is longer should produce 5 pairs")
		assert.True(t, result[2].First.IsEmpty(), "ZipLongest third pair should have First empty when first stream exhausted")
		assert.True(t, result[2].Second.IsPresent(), "ZipLongest third pair should have Second present when second stream longer")
	})

	t.Run("ZipLongestEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := ZipLongest(Of(1, 2, 3, 4, 5), Of("a", "b", "c", "d", "e")).Limit(2).Collect()
		assert.Len(t, result, 2, "ZipLongest then Limit(2) should return 2 pairs")
	})

	t.Run("ZipLongestEarlyTerminationDrainFirst", func(t *testing.T) {
		t.Parallel()
		result := ZipLongest(Of(1, 2, 3, 4, 5), Of("a")).Limit(3).Collect()
		assert.Len(t, result, 3, "ZipLongest(first longer) then Limit(3) should return 3 pairs")
	})

	t.Run("ZipLongestEarlyTerminationDrainSecond", func(t *testing.T) {
		t.Parallel()
		result := ZipLongest(Of(1), Of("a", "b", "c", "d", "e")).Limit(3).Collect()
		assert.Len(t, result, 3, "ZipLongest(second longer) then Limit(3) should return 3 pairs")
	})

	t.Run("ZipLongestWithSameLength", func(t *testing.T) {
		t.Parallel()
		result := ZipLongestWith(Of(1, 2, 3), Of("a", "b", "c"), 0, "").Collect()
		assert.Len(t, result, 3, "ZipLongestWith with same-length inputs should produce 3 pairs")
		assert.Equal(t, 1, result[0].First, "ZipLongestWith first pair should have First=1")
		assert.Equal(t, "a", result[0].Second, "ZipLongestWith first pair should have Second=\"a\"")
	})

	t.Run("ZipLongestWithFirstLonger", func(t *testing.T) {
		t.Parallel()
		result := ZipLongestWith(Of(1, 2, 3, 4, 5), Of("a", "b"), 0, "default").Collect()
		assert.Len(t, result, 5, "ZipLongestWith where first longer should produce 5 pairs using default for missing second")
		assert.Equal(t, 3, result[2].First, "ZipLongestWith third pair should have First=3")
		assert.Equal(t, "default", result[2].Second, "Should use default value when second exhausted")
	})

	t.Run("ZipLongestWithSecondLonger", func(t *testing.T) {
		t.Parallel()
		result := ZipLongestWith(Of(1, 2), Of("a", "b", "c", "d", "e"), -1, "").Collect()
		assert.Len(t, result, 5, "ZipLongestWith where second longer should produce 5 pairs using default for missing first")
		assert.Equal(t, -1, result[2].First, "Should use default value when first exhausted")
		assert.Equal(t, "c", result[2].Second, "ZipLongestWith should keep second stream values in order")
	})

	t.Run("ZipLongestWithEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := ZipLongestWith(Of(1, 2, 3, 4, 5), Of("a", "b", "c", "d", "e"), 0, "").Limit(2).Collect()
		assert.Len(t, result, 2, "ZipLongestWith then Limit(2) should return 2 pairs")
	})

	t.Run("ZipLongestWithEarlyTerminationDrainFirst", func(t *testing.T) {
		t.Parallel()
		result := ZipLongestWith(Of(1, 2, 3, 4, 5), Of("a"), 0, "x").Limit(3).Collect()
		assert.Len(t, result, 3, "ZipLongestWith(first longer) then Limit(3) should return 3 pairs")
	})

	t.Run("ZipLongestWithEarlyTerminationDrainSecond", func(t *testing.T) {
		t.Parallel()
		result := ZipLongestWith(Of(1), Of("a", "b", "c", "d", "e"), 0, "").Limit(3).Collect()
		assert.Len(t, result, 3, "ZipLongestWith(second longer) then Limit(3) should return 3 pairs")
	})
}

// TestCartesianOperations tests Cartesian product operations.
func TestCartesianOperations(t *testing.T) {
	t.Parallel()
	t.Run("CartesianBasic", func(t *testing.T) {
		t.Parallel()
		result := Cartesian(Of(1, 2), Of("a", "b")).Collect()
		expected := []Pair[int, string]{
			{First: 1, Second: "a"},
			{First: 1, Second: "b"},
			{First: 2, Second: "a"},
			{First: 2, Second: "b"},
		}
		assert.Equal(t, expected, result, "Cartesian should produce the full cross product")
	})

	t.Run("CartesianEmpty", func(t *testing.T) {
		t.Parallel()
		result := Cartesian(Empty[int](), Of("a", "b")).Collect()
		assert.Empty(t, result, "Cartesian with empty first stream should be empty")

		result2 := Cartesian(Of(1, 2), Empty[string]()).Collect()
		assert.Empty(t, result2, "Cartesian with empty second stream should be empty")
	})

	t.Run("CartesianEarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := Cartesian(Of(1, 2, 3), Of("a", "b", "c")).Limit(4).Collect()
		assert.Len(t, result, 4, "Cartesian then Limit(4) should return 4 pairs")
	})

	t.Run("CartesianEarlyTerminationInnerLoop", func(t *testing.T) {
		t.Parallel()
		result := Cartesian(Of(1, 2, 3), Of("a", "b", "c", "d", "e")).Limit(3).Collect()
		assert.Len(t, result, 3, "Cartesian with Limit should stop in inner loop")
	})
}

// TestCrossProduct tests CrossProduct function.
func TestCrossProduct(t *testing.T) {
	t.Parallel()
	t.Run("TwoStreams", func(t *testing.T) {
		t.Parallel()
		result := CrossProduct(Of(1, 2), Of(3, 4)).Collect()
		expected := [][]int{
			{1, 3}, {1, 4},
			{2, 3}, {2, 4},
		}
		assert.Equal(t, expected, result, "CrossProduct with 2 streams should produce 4 combinations")
	})

	t.Run("ThreeStreams", func(t *testing.T) {
		t.Parallel()
		result := CrossProduct(Of(1), Of(2), Of(3)).Collect()
		assert.Len(t, result, 1, "CrossProduct with 3 singletons should produce 1 combination")
		assert.Equal(t, []int{1, 2, 3}, result[0], "The only combination should be [1 2 3]")
	})

	t.Run("SingleStream", func(t *testing.T) {
		t.Parallel()
		result := CrossProduct(Of(1, 2, 3)).Collect()
		expected := [][]int{{1}, {2}, {3}}
		assert.Equal(t, expected, result, "CrossProduct with single stream should wrap each element")
	})

	t.Run("Empty", func(t *testing.T) {
		t.Parallel()
		result := CrossProduct[int]().Collect()
		assert.Empty(t, result, "CrossProduct with no streams should be empty")
	})

	t.Run("WithEmptyStream", func(t *testing.T) {
		t.Parallel()
		result := CrossProduct(Of(1, 2), Empty[int]()).Collect()
		assert.Empty(t, result, "CrossProduct with an empty inner stream should be empty")
	})

	t.Run("EarlyTermination", func(t *testing.T) {
		t.Parallel()
		result := CrossProduct(Of(1, 2), Of(3, 4), Of(5, 6)).Limit(3).Collect()
		assert.Len(t, result, 3, "CrossProduct then Limit(3) should return 3 combinations")
	})
}

// TestWindowZeroSize tests Window with zero or negative size.
func TestWindowZeroSize(t *testing.T) {
	t.Parallel()
	t.Run("ZeroSize", func(t *testing.T) {
		t.Parallel()
		result := Window(Of(1, 2, 3, 4, 5), 0).Collect()
		assert.Empty(t, result, "Window with size 0 should return empty stream")
	})

	t.Run("NegativeSize", func(t *testing.T) {
		t.Parallel()
		result := Window(Of(1, 2, 3, 4, 5), -1).Collect()
		assert.Empty(t, result, "Window with negative size should return empty stream")
	})
}

// TestPermutationsOperations tests Permutations function.
func TestPermutationsOperations(t *testing.T) {
	t.Parallel()
	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		result := Permutations(Of(1, 2, 3)).Collect()
		// 3! = 6 permutations
		assert.Len(t, result, 6, "Permutations of 3 elements should be 6")
	})

	t.Run("Empty", func(t *testing.T) {
		t.Parallel()
		result := Permutations(Empty[int]()).Collect()
		assert.Empty(t, result, "Permutations of empty stream should be empty")
	})

	t.Run("Single", func(t *testing.T) {
		t.Parallel()
		result := Permutations(Of(1)).Collect()
		assert.Len(t, result, 1, "Permutations of 1 element should be 1")
		assert.Equal(t, []int{1}, result[0], "Only permutation of [1] should be [1]")
	})

	t.Run("EarlyTerminationFirstPermutation", func(t *testing.T) {
		t.Parallel()
		// This tests the early termination after the first yield (lines 1329-1331)
		result := Permutations(Of(1, 2, 3)).Limit(1).Collect()
		assert.Len(t, result, 1, "Permutations with Limit(1) should stop after first permutation")
		assert.Equal(t, []int{1, 2, 3}, result[0], "First permutation should be original order")
	})

	t.Run("EarlyTerminationPartialPermutations", func(t *testing.T) {
		t.Parallel()
		result := Permutations(Of(1, 2, 3)).Limit(3).Collect()
		assert.Len(t, result, 3, "Permutations with Limit(3) should return 3 permutations")
	})

	t.Run("TwoElements", func(t *testing.T) {
		t.Parallel()
		result := Permutations(Of("a", "b")).Collect()
		assert.Len(t, result, 2, "Permutations of 2 elements should be 2")
		// Should have [a,b] and [b,a]
		assert.Contains(t, result, []string{"a", "b"}, "Permutations should include [a, b]")
		assert.Contains(t, result, []string{"b", "a"}, "Permutations should include [b, a]")
	})
}

// TestMergeSortedNHeapAdvanced tests advanced MergeSortedNHeap operations.
func TestMergeSortedNHeapAdvanced(t *testing.T) {
	t.Parallel()
	t.Run("HeapUpOperation", func(t *testing.T) {
		t.Parallel()
		// Create streams where smaller values come from later streams
		// This forces the heap to bubble up elements during Push
		s1 := Of(10, 20, 30)
		s2 := Of(1, 2, 3)
		s3 := Of(5, 15, 25)

		result := MergeSortedNHeap(func(a, b int) int { return a - b }, s1, s2, s3).Collect()
		expected := []int{1, 2, 3, 5, 10, 15, 20, 25, 30}
		assert.Equal(t, expected, result, "MergeSortedNHeap should correctly sort elements requiring heap up")
	})

	t.Run("EarlyTerminationAfterFirstYield", func(t *testing.T) {
		t.Parallel()
		// Test early termination right after the first element is yielded
		s1 := Of(1, 4, 7)
		s2 := Of(2, 5, 8)
		s3 := Of(3, 6, 9)

		result := MergeSortedNHeap(func(a, b int) int { return a - b }, s1, s2, s3).Limit(1).Collect()
		assert.Len(t, result, 1, "MergeSortedNHeap with Limit(1) should stop after first element")
		assert.Equal(t, 1, result[0], "First element should be the smallest across all streams")
	})

	t.Run("EmptyStreamsInMix", func(t *testing.T) {
		t.Parallel()
		// Mix of empty and non-empty streams
		s1 := Empty[int]()
		s2 := Of(1, 3, 5)
		s3 := Empty[int]()
		s4 := Of(2, 4, 6)

		result := MergeSortedNHeap(func(a, b int) int { return a - b }, s1, s2, s3, s4).Collect()
		assert.Equal(t, []int{1, 2, 3, 4, 5, 6}, result, "MergeSortedNHeap should merge non-empty streams and skip empties")
	})

	t.Run("DescendingOrder", func(t *testing.T) {
		t.Parallel()
		// Test with descending order comparator
		s1 := Of(5, 3, 1)
		s2 := Of(6, 4, 2)

		result := MergeSortedNHeap(func(a, b int) int { return b - a }, s1, s2).Collect()
		expected := []int{6, 5, 4, 3, 2, 1}
		assert.Equal(t, expected, result, "MergeSortedNHeap should work with descending order")
	})

	t.Run("SingleElementStreams", func(t *testing.T) {
		t.Parallel()
		// Many single-element streams to exercise heap operations
		result := MergeSortedNHeap(func(a, b int) int { return a - b },
			Of(5), Of(1), Of(9), Of(3), Of(7), Of(2), Of(8), Of(4), Of(6),
		).Collect()
		expected := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
		assert.Equal(t, expected, result, "MergeSortedNHeap with many single-element streams should yield sorted 1..9")
	})

	t.Run("DuplicateValues", func(t *testing.T) {
		t.Parallel()
		// Streams with duplicate values
		s1 := Of(1, 3, 3, 5)
		s2 := Of(2, 3, 4, 5)

		result := MergeSortedNHeap(func(a, b int) int { return a - b }, s1, s2).Collect()
		expected := []int{1, 2, 3, 3, 3, 4, 5, 5}
		assert.Equal(t, expected, result, "MergeSortedNHeap should retain duplicates and sort overall")
	})
}

func TestMergeHeapPop(t *testing.T) {
	t.Parallel()
	t.Run("Empty", func(t *testing.T) {
		t.Parallel()
		h := &mergeHeap[int]{cmp: func(a, b int) int { return a - b }}
		assert.Nil(t, h.Pop(), "Pop on empty heap should return nil")
	})

	t.Run("SingleElement", func(t *testing.T) {
		t.Parallel()
		h := &mergeHeap[int]{cmp: func(a, b int) int { return a - b }}
		h.Push(&mergeHeapItem[int]{value: 42})
		item := h.Pop()
		if assert.NotNil(t, item, "Pop should return the only item") {
			assert.Equal(t, 42, item.value, "Pop should return the pushed value")
		}
		assert.Nil(t, h.Pop(), "Pop should return nil after removing last item")
	})
}

// TestEarlyTerminationDirectSeq tests early termination by directly calling seq with a yield that returns false.
func TestEarlyTerminationDirectSeq(t *testing.T) {
	t.Parallel()
	t.Run("MergeSortedNHeapDirectEarlyTermination", func(t *testing.T) {
		t.Parallel()
		// Test MergeSortedNHeap early termination by directly using seq with yield returning false
		s1 := Of(1, 4, 7)
		s2 := Of(2, 5, 8)
		s3 := Of(3, 6, 9)

		merged := MergeSortedNHeap(func(a, b int) int { return a - b }, s1, s2, s3)
		count := 0
		merged.seq(func(_ int) bool {
			count++
			return false // Stop immediately after first element
		})
		assert.Equal(t, 1, count, "Should stop after first element when yield returns false")
	})

	t.Run("PermutationsDirectEarlyTerminationFirstPermutation", func(t *testing.T) {
		t.Parallel()
		// Test Permutations early termination on first permutation
		stream := Permutations(Of(1, 2, 3))
		count := 0
		stream.seq(func(_ []int) bool {
			count++
			return false // Stop immediately after first permutation
		})
		assert.Equal(t, 1, count, "Should stop after first permutation when yield returns false")
	})

	t.Run("PermutationsDirectEarlyTerminationMiddlePermutation", func(t *testing.T) {
		t.Parallel()
		// Test Permutations early termination in middle of generation
		stream := Permutations(Of(1, 2, 3))
		count := 0
		stream.seq(func(_ []int) bool {
			count++
			return count < 3 // Stop after 3 permutations
		})
		assert.Equal(t, 3, count, "Should stop after 3 permutations when yield returns false")
	})

	t.Run("MergeSortedNHeapHeapPathEarlyTermination", func(t *testing.T) {
		t.Parallel()
		// Use more than 4 streams to trigger the heap-based path
		s1 := Of(1, 11, 21)
		s2 := Of(2, 12, 22)
		s3 := Of(3, 13, 23)
		s4 := Of(4, 14, 24)
		s5 := Of(5, 15, 25)

		merged := MergeSortedNHeap(func(a, b int) int { return a - b }, s1, s2, s3, s4, s5)
		count := 0
		merged.seq(func(_ int) bool {
			count++
			return false // Stop immediately after first element
		})
		assert.Equal(t, 1, count, "Heap path should stop after first element when yield returns false")
	})

	t.Run("MergeSortedNHeapHeapPathPartialConsumption", func(t *testing.T) {
		t.Parallel()
		// Use more than 4 streams to trigger the heap-based path
		s1 := Of(1, 11, 21)
		s2 := Of(2, 12, 22)
		s3 := Of(3, 13, 23)
		s4 := Of(4, 14, 24)
		s5 := Of(5, 15, 25)

		result := MergeSortedNHeap(func(a, b int) int { return a - b }, s1, s2, s3, s4, s5).Limit(7).Collect()
		expected := []int{1, 2, 3, 4, 5, 11, 12}
		assert.Equal(t, expected, result, "Heap path should correctly return first 7 elements")
	})
}
