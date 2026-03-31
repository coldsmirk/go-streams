package streams

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestStream2 tests Stream2 operations.
func TestStream2(t *testing.T) {
	t.Parallel()
	t.Run("From2", func(t *testing.T) {
		t.Parallel()
		m := map[string]int{"a": 1, "b": 2}
		s := FromMap(m)
		count := s.Count()
		assert.Equal(t, 2, count, "From2 via FromMap should create Stream2")
	})

	t.Run("PairsOf", func(t *testing.T) {
		t.Parallel()
		pairs := PairsOf(
			NewPair("a", 1),
			NewPair("b", 2),
		)
		result := pairs.CollectPairs()
		assert.Len(t, result, 2, "PairsOf should create Stream2 from pairs")
	})

	t.Run("Empty2", func(t *testing.T) {
		t.Parallel()
		result := Empty2[string, int]().CollectPairs()
		assert.Empty(t, result, "Empty2 should create empty Stream2")
	})

	t.Run("Filter", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(
			NewPair("a", 1),
			NewPair("b", 2),
			NewPair("c", 3),
		)

		result := s.Filter(func(k string, v int) bool {
			return v > 1
		}).CollectPairs()

		assert.Len(t, result, 2, "Filter should keep matching pairs")
	})

	t.Run("MapKeys", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(NewPair("a", 1), NewPair("b", 2))
		result := s.MapKeys(func(k string) string {
			return k + k
		}).CollectPairs()

		assert.Equal(t, "aa", result[0].First, "MapKeys should transform keys")
		assert.Equal(t, "bb", result[1].First, "MapKeys should transform keys")
	})

	t.Run("MapValues", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(NewPair("a", 1), NewPair("b", 2))
		result := s.MapValues(func(v int) int {
			return v * 10
		}).CollectPairs()

		assert.Equal(t, 10, result[0].Second, "MapValues should transform values")
		assert.Equal(t, 20, result[1].Second, "MapValues should transform values")
	})

	t.Run("Limit", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(
			NewPair("a", 1),
			NewPair("b", 2),
			NewPair("c", 3),
		)
		result := s.Limit(2).CollectPairs()
		assert.Len(t, result, 2, "Limit should take first n pairs")
	})

	t.Run("Skip", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(
			NewPair("a", 1),
			NewPair("b", 2),
			NewPair("c", 3),
		)
		result := s.Skip(1).CollectPairs()
		assert.Len(t, result, 2, "Skip should skip first n pairs")
	})

	t.Run("Keys", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(NewPair("a", 1), NewPair("b", 2))
		keys := s.Keys().Collect()
		assert.Equal(t, []string{"a", "b"}, keys, "Keys should extract keys")
	})

	t.Run("Values", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(NewPair("a", 1), NewPair("b", 2))
		values := s.Values().Collect()
		assert.Equal(t, []int{1, 2}, values, "Values should extract values")
	})

	t.Run("ToPairs", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(NewPair("a", 1), NewPair("b", 2))
		pairs := s.ToPairs().Collect()
		assert.Len(t, pairs, 2, "ToPairs should convert to Stream of Pairs")
	})

	t.Run("ForEach", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(NewPair("a", 1), NewPair("b", 2))
		var keys []string
		var values []int
		s.ForEach(func(k string, v int) {
			keys = append(keys, k)
			values = append(values, v)
		})
		assert.Equal(t, []string{"a", "b"}, keys, "ForEach should visit all keys")
		assert.Equal(t, []int{1, 2}, values, "ForEach should visit all values")
	})

	t.Run("Count", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(NewPair("a", 1), NewPair("b", 2), NewPair("c", 3))
		assert.Equal(t, 3, s.Count(), "Count should return number of pairs")
	})

	t.Run("AnyMatch", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(NewPair("a", 1), NewPair("b", 2))
		assert.True(t, s.AnyMatch(func(k string, v int) bool { return v == 2 }), "AnyMatch(v==2) should return true")
		assert.False(t, s.AnyMatch(func(k string, v int) bool { return v == 5 }), "AnyMatch(v==5) should return false when no pair matches")
	})

	t.Run("AllMatch", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(NewPair("a", 2), NewPair("b", 4))
		assert.True(t, s.AllMatch(func(k string, v int) bool { return v%2 == 0 }), "AllMatch(v%2==0) should be true for values [2,4]")
	})

	t.Run("NoneMatch", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(NewPair("a", 1), NewPair("b", 3))
		assert.True(t, s.NoneMatch(func(k string, v int) bool { return v%2 == 0 }), "NoneMatch(v%2==0) should be true for values [1,3]")
	})

	t.Run("First", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(NewPair("a", 1), NewPair("b", 2))
		first := s.First()
		assert.True(t, first.IsPresent(), "First should return Some for non-empty")
		assert.Equal(t, "a", first.Get().First, "First should return first pair")
	})

	t.Run("FirstEmpty", func(t *testing.T) {
		t.Parallel()
		s := Empty2[string, int]()
		first := s.First()
		assert.True(t, first.IsEmpty(), "First should return None for empty stream")
	})

	t.Run("TakeWhile", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(
			NewPair("a", 1),
			NewPair("b", 2),
			NewPair("c", 3),
		)
		result := s.TakeWhile(func(k string, v int) bool { return v < 3 }).CollectPairs()
		assert.Len(t, result, 2, "TakeWhile should take while predicate is true")
	})

	t.Run("DropWhile", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(
			NewPair("a", 1),
			NewPair("b", 2),
			NewPair("c", 3),
		)
		result := s.DropWhile(func(k string, v int) bool { return v < 2 }).CollectPairs()
		assert.Len(t, result, 2, "DropWhile should drop while predicate is true")
	})

	t.Run("Peek", func(t *testing.T) {
		t.Parallel()
		var peeked []string
		s := PairsOf(NewPair("a", 1), NewPair("b", 2))
		result := s.Peek(func(k string, v int) {
			peeked = append(peeked, k)
		}).CollectPairs()

		assert.Len(t, result, 2, "Peek should not modify stream")
		assert.Equal(t, []string{"a", "b"}, peeked, "Peek should execute action")
	})

	t.Run("ToMap2", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(NewPair("a", 1), NewPair("b", 2))
		m := ToMap2(s)
		assert.Equal(t, 1, m["a"], "ToMap2 should create map")
		assert.Equal(t, 2, m["b"], "ToMap2 should create map")
	})

	t.Run("Reduce", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(NewPair("a", 1), NewPair("b", 2), NewPair("c", 3))
		result := s.Reduce(NewPair("", 0), func(acc Pair[string, int], k string, v int) Pair[string, int] {
			return NewPair(acc.First+k, acc.Second+v)
		})
		assert.Equal(t, "abc", result.First, "Reduce should combine keys")
		assert.Equal(t, 6, result.Second, "Reduce should combine values")
	})

	t.Run("Seq2", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(NewPair("a", 1), NewPair("b", 2))
		seq := s.Seq2()

		var keys []string
		for k := range seq {
			keys = append(keys, k)
		}
		assert.Equal(t, []string{"a", "b"}, keys, "Seq2 should return underlying iter.Seq2")
	})
}

// TestStream2TypeTransformations tests Stream2 type transformation functions.
func TestStream2TypeTransformations(t *testing.T) {
	t.Parallel()
	t.Run("MapKeysTo", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(NewPair("a", 1), NewPair("bb", 2))
		result := MapKeysTo(s, func(k string) int {
			return len(k)
		}).CollectPairs()

		assert.Equal(t, 1, result[0].First, "MapKeysTo should transform key type")
		assert.Equal(t, 2, result[1].First, "MapKeysTo should transform key type")
	})

	t.Run("MapValuesTo", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(NewPair("a", 1), NewPair("b", 2))
		result := MapValuesTo(s, func(v int) string {
			return string(rune('a' + v - 1))
		}).CollectPairs()

		assert.Equal(t, "a", result[0].Second, "MapValuesTo should transform value type")
		assert.Equal(t, "b", result[1].Second, "MapValuesTo should transform value type")
	})

	t.Run("MapPairs", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(NewPair("a", 1), NewPair("b", 2))
		result := MapPairs(s, func(k string, v int) (int, string) {
			return v, k
		}).CollectPairs()

		assert.Equal(t, 1, result[0].First, "MapPairs should transform both types")
		assert.Equal(t, "a", result[0].Second, "MapPairs should transform both types")
	})

	t.Run("MapPairsTo", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(NewPair("a", 1), NewPair("bb", 2), NewPair("ccc", 3))
		result := MapPairsTo(s, func(k string, v int) string {
			return k + "=" + strconv.Itoa(v)
		}).Collect()

		assert.Equal(t, []string{"a=1", "bb=2", "ccc=3"}, result, "MapPairsTo should combine key and value into single element")
	})

	t.Run("MapPairsToEmpty", func(t *testing.T) {
		t.Parallel()
		result := MapPairsTo(Empty2[string, int](), func(k string, v int) string {
			return k
		}).Collect()

		assert.Empty(t, result, "MapPairsTo on empty stream should return empty slice")
	})

	t.Run("MapToPairs", func(t *testing.T) {
		t.Parallel()
		s := Of("hello", "go", "streams")
		result := MapToPairs(s, func(s string) (string, int) {
			return s, len(s)
		}).CollectPairs()

		assert.Equal(t, []Pair[string, int]{
			NewPair("hello", 5),
			NewPair("go", 2),
			NewPair("streams", 7),
		}, result, "MapToPairs should map each element to a key-value pair")
	})

	t.Run("MapToPairsEmpty", func(t *testing.T) {
		t.Parallel()
		result := MapToPairs(Empty[string](), func(s string) (string, int) {
			return s, len(s)
		}).CollectPairs()

		assert.Empty(t, result, "MapToPairs on empty stream should return empty")
	})

	t.Run("MapToPairsChained", func(t *testing.T) {
		t.Parallel()
		result := MapToPairs(Of("a", "bb", "ccc"), func(s string) (int, string) {
			return len(s), s
		}).Filter(func(k int, v string) bool {
			return k > 1
		}).CollectPairs()

		assert.Equal(t, []Pair[int, string]{
			NewPair(2, "bb"),
			NewPair(3, "ccc"),
		}, result, "MapToPairs should chain with Stream2 operations")
	})

	t.Run("MapToPairsEarlyTermination", func(t *testing.T) {
		t.Parallel()
		count := 0
		result := MapToPairs(Of(1, 2, 3, 4, 5), func(v int) (int, int) {
			count++
			return v, v * v
		}).Limit(2).CollectPairs()

		assert.Len(t, result, 2, "MapToPairs should support early termination")
		assert.LessOrEqual(t, count, 3, "MapToPairs should not process all elements beyond limit")
	})

	t.Run("SwapKeyValue", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(NewPair("a", 1), NewPair("b", 2))
		result := SwapKeyValue(s).CollectPairs()

		assert.Equal(t, 1, result[0].First, "SwapKeyValue should swap key and value")
		assert.Equal(t, "a", result[0].Second, "SwapKeyValue should swap key and value")
	})
}

// TestStream2NewOperations tests newly added Stream2 operations.
func TestStream2NewOperations(t *testing.T) {
	t.Parallel()
	t.Run("ReduceByKey", func(t *testing.T) {
		t.Parallel()
		// Sum values by key
		s := PairsOf(
			NewPair("a", 1), NewPair("b", 2), NewPair("a", 3), NewPair("b", 4), NewPair("a", 5),
		)
		result := ReduceByKey(s, func(a, b int) int { return a + b })
		assert.Equal(t, 9, result["a"], "ReduceByKey should sum values for key 'a'")
		assert.Equal(t, 6, result["b"], "ReduceByKey should sum values for key 'b'")

		// Single key
		s2 := PairsOf(NewPair("x", 10))
		result2 := ReduceByKey(s2, func(a, b int) int { return a + b })
		assert.Equal(t, 10, result2["x"], "ReduceByKey with single value should return that value")

		// Empty stream
		s3 := Empty2[string, int]()
		result3 := ReduceByKey(s3, func(a, b int) int { return a + b })
		assert.Empty(t, result3, "ReduceByKey on empty stream should return empty map")
	})

	t.Run("ReduceByKeyWithInit", func(t *testing.T) {
		t.Parallel()
		// Count by key
		s := PairsOf(
			NewPair("a", "x"), NewPair("b", "y"), NewPair("a", "z"),
		)
		result := ReduceByKeyWithInit(s, func() int { return 0 }, func(count int, _ string) int { return count + 1 })
		assert.Equal(t, 2, result["a"], "ReduceByKeyWithInit should count 2 for key 'a'")
		assert.Equal(t, 1, result["b"], "ReduceByKeyWithInit should count 1 for key 'b'")
	})

	t.Run("GroupValues", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(
			NewPair("a", 1), NewPair("b", 2), NewPair("a", 3), NewPair("b", 4),
		)
		result := GroupValues(s)
		assert.Equal(t, []int{1, 3}, result["a"], "GroupValues should group values for key 'a'")
		assert.Equal(t, []int{2, 4}, result["b"], "GroupValues should group values for key 'b'")

		// Empty stream
		s2 := Empty2[string, int]()
		result2 := GroupValues(s2)
		assert.Empty(t, result2, "GroupValues on empty stream should return empty map")
	})

	t.Run("DistinctKeys", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(
			NewPair("a", 1), NewPair("b", 2), NewPair("a", 3), NewPair("c", 4), NewPair("b", 5),
		)
		result := DistinctKeys(s).CollectPairs()
		assert.Len(t, result, 3, "DistinctKeys should keep only first occurrence of each key")
		assert.Equal(t, "a", result[0].First, "DistinctKeys should keep first key in stream order")
		assert.Equal(t, 1, result[0].Second, "DistinctKeys should keep first value for key 'a'")
		assert.Equal(t, "b", result[1].First, "DistinctKeys should keep second distinct key in order")
		assert.Equal(t, 2, result[1].Second, "DistinctKeys should keep first value for key 'b'")
	})

	t.Run("DistinctValues", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(
			NewPair("a", 1), NewPair("b", 2), NewPair("c", 1), NewPair("d", 3), NewPair("e", 2),
		)
		result := DistinctValues(s).CollectPairs()
		assert.Len(t, result, 3, "DistinctValues should keep only first occurrence of each value")
		assert.Equal(t, 1, result[0].Second, "DistinctValues result[0].Second should be 1")
		assert.Equal(t, 2, result[1].Second, "DistinctValues result[1].Second should be 2")
		assert.Equal(t, 3, result[2].Second, "DistinctValues result[2].Second should be 3")
	})

	t.Run("KeysEarlyTermination", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(NewPair("a", 1), NewPair("b", 2), NewPair("c", 3))
		keys := s.Keys().Limit(2).Collect()
		assert.Equal(t, []string{"a", "b"}, keys, "Keys().Limit(2) should yield [\"a\",\"b\"]")
	})

	t.Run("ValuesEarlyTermination", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(NewPair("a", 1), NewPair("b", 2), NewPair("c", 3))
		values := s.Values().Limit(2).Collect()
		assert.Equal(t, []int{1, 2}, values, "Values().Limit(2) should yield [1,2]")
	})

	t.Run("ToPairsEarlyTermination", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(NewPair("a", 1), NewPair("b", 2), NewPair("c", 3))
		pairs := s.ToPairs().Limit(2).Collect()
		assert.Len(t, pairs, 2, "ToPairs().Limit(2) should return exactly 2 pairs")
	})

	t.Run("LimitEarlyTermination", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(NewPair("a", 1), NewPair("b", 2), NewPair("c", 3))
		result := s.Limit(2).Limit(1).CollectPairs()
		assert.Len(t, result, 1, "Nested Limit should stop at smallest")
	})

	t.Run("SkipEarlyTermination", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(NewPair("a", 1), NewPair("b", 2), NewPair("c", 3))
		result := s.Skip(1).Limit(1).CollectPairs()
		assert.Len(t, result, 1, "Skip(1) then Limit(1) should return 1 pair")
		assert.Equal(t, "b", result[0].First, "Skip(1) should drop the first key")
	})

	t.Run("PeekEarlyTermination", func(t *testing.T) {
		t.Parallel()
		var peeked []string
		s := PairsOf(NewPair("a", 1), NewPair("b", 2), NewPair("c", 3))
		result := s.Peek(func(k string, v int) {
			peeked = append(peeked, k)
		}).Limit(2).CollectPairs()
		assert.Len(t, result, 2, "Peek with Limit should stop after collecting 2")
		// Peek sees elements before yield, so it may see one extra element
		assert.GreaterOrEqual(t, len(peeked), 2, "Peek should see at least 2 elements")
	})

	t.Run("TakeWhileEarlyTermination", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(NewPair("a", 1), NewPair("b", 2), NewPair("c", 3))
		result := s.TakeWhile(func(k string, v int) bool { return v < 3 }).Limit(1).CollectPairs()
		assert.Len(t, result, 1, "TakeWhile(v<3) then Limit(1) should return 1 pair")
	})

	t.Run("DropWhileEarlyTermination", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(NewPair("a", 1), NewPair("b", 2), NewPair("c", 3))
		result := s.DropWhile(func(k string, v int) bool { return v < 2 }).Limit(1).CollectPairs()
		assert.Len(t, result, 1, "DropWhile(v<2) then Limit(1) should return 1 pair")
		assert.Equal(t, "b", result[0].First, "DropWhile should drop until predicate fails")
	})

	t.Run("AllMatchEarlyFalse", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(NewPair("a", 1), NewPair("b", 2), NewPair("c", 3))
		result := s.AllMatch(func(k string, v int) bool { return v < 2 })
		assert.False(t, result, "AllMatch should return false early when condition fails")
	})

	t.Run("NoneMatchEarlyFalse", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(NewPair("a", 1), NewPair("b", 2), NewPair("c", 3))
		result := s.NoneMatch(func(k string, v int) bool { return v == 2 })
		assert.False(t, result, "NoneMatch should return false early when match found")
	})

	t.Run("MapKeysToEarlyTermination", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(NewPair("a", 1), NewPair("bb", 2), NewPair("ccc", 3))
		result := MapKeysTo(s, func(k string) int { return len(k) }).Limit(2).CollectPairs()
		assert.Len(t, result, 2, "MapKeysTo(len(k)) then Limit(2) should return 2 pairs")
	})

	t.Run("MapValuesToEarlyTermination", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(NewPair("a", 1), NewPair("b", 2), NewPair("c", 3))
		result := MapValuesTo(s, func(v int) string { return string(rune('a' + v - 1)) }).Limit(2).CollectPairs()
		assert.Len(t, result, 2, "MapValuesTo(rune map) then Limit(2) should return 2 pairs")
	})

	t.Run("MapPairsEarlyTermination", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(NewPair("a", 1), NewPair("b", 2), NewPair("c", 3))
		result := MapPairs(s, func(k string, v int) (int, string) { return v, k }).Limit(2).CollectPairs()
		assert.Len(t, result, 2, "MapPairs(swap types) then Limit(2) should return 2 pairs")
	})

	t.Run("SwapKeyValueEarlyTermination", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(NewPair("a", 1), NewPair("b", 2), NewPair("c", 3))
		result := SwapKeyValue(s).Limit(2).CollectPairs()
		assert.Len(t, result, 2, "SwapKeyValue then Limit(2) should return 2 pairs")
	})

	t.Run("DistinctKeysEarlyTermination", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(NewPair("a", 1), NewPair("b", 2), NewPair("a", 3), NewPair("c", 4))
		result := DistinctKeys(s).Limit(2).CollectPairs()
		assert.Len(t, result, 2, "DistinctKeys then Limit(2) should return 2 pairs with unique keys")
	})

	t.Run("DistinctValuesEarlyTermination", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(NewPair("a", 1), NewPair("b", 2), NewPair("c", 1), NewPair("d", 3))
		result := DistinctValues(s).Limit(2).CollectPairs()
		assert.Len(t, result, 2, "DistinctValues then Limit(2) should return 2 pairs with unique values")
	})

	t.Run("FilterEarlyTermination", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(
			NewPair("a", 2),
			NewPair("b", 4),
			NewPair("c", 6),
			NewPair("d", 8),
		)
		result := s.Filter(func(k string, v int) bool {
			return v > 0
		}).Limit(2).CollectPairs()
		assert.Len(t, result, 2, "Filter(v>0) then Limit(2) should return 2 pairs")
	})

	t.Run("MapKeysEarlyTermination", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(NewPair("a", 1), NewPair("b", 2), NewPair("c", 3))
		result := s.MapKeys(func(k string) string {
			return k + k
		}).Limit(2).CollectPairs()
		assert.Len(t, result, 2, "MapKeys(k+k) then Limit(2) should return 2 pairs")
	})

	t.Run("MapValuesEarlyTermination", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(NewPair("a", 1), NewPair("b", 2), NewPair("c", 3))
		result := s.MapValues(func(v int) int {
			return v * 10
		}).Limit(2).CollectPairs()
		assert.Len(t, result, 2, "MapValues(*10) then Limit(2) should return 2 pairs")
	})
}

// TestStream2ParallelMapValues tests Stream2.ParallelMapValues.
func TestStream2ParallelMapValues(t *testing.T) {
	t.Parallel()

	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(
			NewPair("a", 1),
			NewPair("b", 2),
			NewPair("c", 3),
		)
		result := s.ParallelMapValues(func(v int) int {
			return v * 10
		}, WithOrdered(true)).CollectPairs()

		assert.Len(t, result, 3, "ParallelMapValues should return all pairs")
		assert.Equal(t, 10, result[0].Second, "ParallelMapValues should transform first value")
		assert.Equal(t, 20, result[1].Second, "ParallelMapValues should transform second value")
		assert.Equal(t, 30, result[2].Second, "ParallelMapValues should transform third value")
	})

	t.Run("Empty", func(t *testing.T) {
		t.Parallel()
		result := Empty2[string, int]().ParallelMapValues(func(v int) int {
			return v * 10
		}).CollectPairs()
		assert.Empty(t, result, "ParallelMapValues on empty should return empty")
	})

	t.Run("EarlyTermination", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(
			NewPair("a", 1),
			NewPair("b", 2),
			NewPair("c", 3),
			NewPair("d", 4),
			NewPair("e", 5),
		)
		result := s.ParallelMapValues(func(v int) int {
			return v * 10
		}, WithOrdered(true)).Limit(2).CollectPairs()

		assert.Len(t, result, 2, "ParallelMapValues should respect Limit")
	})

	t.Run("Unordered", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(
			NewPair("a", 1),
			NewPair("b", 2),
			NewPair("c", 3),
		)
		result := s.ParallelMapValues(func(v int) int {
			return v * 10
		}, WithOrdered(false)).CollectPairs()

		assert.Len(t, result, 3, "ParallelMapValues unordered should return all pairs")
	})

	t.Run("LargeInput", func(t *testing.T) {
		t.Parallel()
		pairs := make([]Pair[int, int], 100)
		for i := range pairs {
			pairs[i] = NewPair(i, i)
		}
		s := PairsOf(pairs...)
		result := s.ParallelMapValues(func(v int) int {
			return v * 2
		}, WithConcurrency(4), WithOrdered(true)).CollectPairs()

		assert.Len(t, result, 100, "ParallelMapValues should handle large input")
		for i, p := range result {
			assert.Equal(t, i*2, p.Second, "ParallelMapValues should transform all values correctly")
		}
	})
}

// TestStream2ParallelFilter tests Stream2.ParallelFilter.
func TestStream2ParallelFilter(t *testing.T) {
	t.Parallel()

	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(
			NewPair("a", 1),
			NewPair("b", 2),
			NewPair("c", 3),
			NewPair("d", 4),
		)
		result := s.ParallelFilter(func(k string, v int) bool {
			return v%2 == 0
		}, WithOrdered(true)).CollectPairs()

		assert.Len(t, result, 2, "ParallelFilter should filter pairs")
		assert.Equal(t, "b", result[0].First, "ParallelFilter should keep matching pairs")
		assert.Equal(t, "d", result[1].First, "ParallelFilter should keep matching pairs")
	})

	t.Run("Empty", func(t *testing.T) {
		t.Parallel()
		result := Empty2[string, int]().ParallelFilter(func(k string, v int) bool {
			return true
		}).CollectPairs()
		assert.Empty(t, result, "ParallelFilter on empty should return empty")
	})

	t.Run("AllMatch", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(
			NewPair("a", 2),
			NewPair("b", 4),
			NewPair("c", 6),
		)
		result := s.ParallelFilter(func(k string, v int) bool {
			return v%2 == 0
		}, WithOrdered(true)).CollectPairs()

		assert.Len(t, result, 3, "ParallelFilter should keep all when all match")
	})

	t.Run("NoneMatch", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(
			NewPair("a", 1),
			NewPair("b", 3),
			NewPair("c", 5),
		)
		result := s.ParallelFilter(func(k string, v int) bool {
			return v%2 == 0
		}, WithOrdered(true)).CollectPairs()

		assert.Empty(t, result, "ParallelFilter should return empty when none match")
	})

	t.Run("EarlyTermination", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(
			NewPair("a", 2),
			NewPair("b", 4),
			NewPair("c", 6),
			NewPair("d", 8),
			NewPair("e", 10),
		)
		result := s.ParallelFilter(func(k string, v int) bool {
			return v%2 == 0
		}, WithOrdered(true)).Limit(2).CollectPairs()

		assert.Len(t, result, 2, "ParallelFilter should respect Limit")
	})

	t.Run("Unordered", func(t *testing.T) {
		t.Parallel()
		s := PairsOf(
			NewPair("a", 1),
			NewPair("b", 2),
			NewPair("c", 3),
			NewPair("d", 4),
		)
		result := s.ParallelFilter(func(k string, v int) bool {
			return v%2 == 0
		}, WithOrdered(false)).CollectPairs()

		assert.Len(t, result, 2, "ParallelFilter unordered should filter correctly")
	})

	t.Run("LargeInput", func(t *testing.T) {
		t.Parallel()
		pairs := make([]Pair[int, int], 100)
		for i := range pairs {
			pairs[i] = NewPair(i, i)
		}
		s := PairsOf(pairs...)
		result := s.ParallelFilter(func(k int, v int) bool {
			return v%2 == 0
		}, WithConcurrency(4), WithOrdered(true)).CollectPairs()

		assert.Len(t, result, 50, "ParallelFilter should filter large input correctly")
	})
}
