package streams

import (
	"cmp"
	"testing"

	collections "github.com/coldsmirk/go-collections"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Tests for Stream Constructors from go-collections
// =============================================================================

func TestFromSet(t *testing.T) {
	t.Parallel()
	set := collections.NewHashSetFrom(1, 2, 3)
	result := FromSet(set).Collect()
	assert.Len(t, result, 3, "FromSet should collect all elements")
	assert.ElementsMatch(t, []int{1, 2, 3}, result, "FromSet should preserve set contents (unordered)")
}

func TestFromSortedSet(t *testing.T) {
	t.Parallel()
	set := collections.NewTreeSetFrom(cmp.Compare[int], 3, 1, 2)

	// Ascending
	asc := FromSortedSet(set).Collect()
	assert.Equal(t, []int{1, 2, 3}, asc, "FromSortedSet should iterate ascending")

	// Descending
	desc := FromSortedSetDescending(set).Collect()
	assert.Equal(t, []int{3, 2, 1}, desc, "FromSortedSetDescending should iterate descending")
}

func TestFromList(t *testing.T) {
	t.Parallel()
	list := collections.NewArrayListFrom("a", "b", "c")
	result := FromList(list).Collect()
	assert.Equal(t, []string{"a", "b", "c"}, result, "FromList should iterate in order")
}

func TestFromMapC(t *testing.T) {
	t.Parallel()
	m := collections.NewHashMap[string, int]()
	m.Put("a", 1)
	m.Put("b", 2)

	pairs := FromMapC(m).CollectPairs()
	assert.Len(t, pairs, 2, "FromMapC should produce two pairs")
}

func TestFromSortedMapC(t *testing.T) {
	t.Parallel()
	m := collections.NewTreeMap[string, int](cmp.Compare[string])
	m.Put("c", 3)
	m.Put("a", 1)
	m.Put("b", 2)

	// Ascending
	keys := FromSortedMapC(m).Keys().Collect()
	assert.Equal(t, []string{"a", "b", "c"}, keys, "FromSortedMapC should iterate ascending keys")

	// Descending
	keysDesc := FromSortedMapCDescending(m).Keys().Collect()
	assert.Equal(t, []string{"c", "b", "a"}, keysDesc, "FromSortedMapCDescending should iterate descending keys")
}

func TestFromQueue(t *testing.T) {
	t.Parallel()
	q := collections.NewArrayQueue[int]()
	q.Enqueue(1)
	q.Enqueue(2)
	q.Enqueue(3)

	result := FromQueue(q).Collect()
	assert.Equal(t, []int{1, 2, 3}, result, "FromQueue should iterate FIFO")
}

func TestFromStack(t *testing.T) {
	t.Parallel()
	s := collections.NewArrayStack[int]()
	s.Push(1)
	s.Push(2)
	s.Push(3)

	result := FromStack(s).Collect()
	// Stack iterates LIFO
	assert.Equal(t, []int{3, 2, 1}, result, "FromStack should iterate LIFO")
}

func TestFromDeque(t *testing.T) {
	t.Parallel()
	d := collections.NewArrayDeque[int]()
	d.PushBack(1)
	d.PushBack(2)
	d.PushBack(3)

	// Front to back
	result := FromDeque(d).Collect()
	assert.Equal(t, []int{1, 2, 3}, result, "FromDeque should iterate front-to-back")

	// Back to front
	reversed := FromDequeReversed(d).Collect()
	assert.Equal(t, []int{3, 2, 1}, reversed, "FromDequeReversed should iterate back-to-front")
}

func TestFromPriorityQueue(t *testing.T) {
	t.Parallel()
	pq := collections.NewPriorityQueue(cmp.Compare[int])
	pq.Push(3)
	pq.Push(1)
	pq.Push(2)

	// Heap order (not sorted)
	heapOrder := FromPriorityQueue(pq).Collect()
	assert.Len(t, heapOrder, 3, "FromPriorityQueue should produce three elements")
	assert.ElementsMatch(t, []int{1, 2, 3}, heapOrder, "FromPriorityQueue contents should match")

	// Sorted order (collects first)
	sorted := FromPriorityQueueSorted(pq).Collect()
	assert.Equal(t, []int{1, 2, 3}, sorted, "FromPriorityQueueSorted should return sorted slice")
}

// =============================================================================
// Tests for Terminal Operations returning go-collections
// =============================================================================

func TestToHashSet(t *testing.T) {
	t.Parallel()
	set := ToHashSet(Of(1, 2, 2, 3, 3, 3))
	assert.Equal(t, 3, set.Size(), "ToHashSet should de-duplicate")
	assert.True(t, set.Contains(1), "HashSet should contain 1")
	assert.True(t, set.Contains(2), "HashSet should contain 2")
	assert.True(t, set.Contains(3), "HashSet should contain 3")
	assert.False(t, set.Contains(4), "HashSet should not contain 4")
}

func TestToTreeSet(t *testing.T) {
	t.Parallel()
	set := ToTreeSet(Of(3, 1, 2), cmp.Compare[int])
	assert.Equal(t, 3, set.Size(), "ToTreeSet should contain all elements")

	first, ok := set.First()
	assert.True(t, ok, "TreeSet.First should succeed")
	assert.Equal(t, 1, first, "TreeSet.First should be min")

	last, ok := set.Last()
	assert.True(t, ok, "TreeSet.Last should succeed")
	assert.Equal(t, 3, last, "TreeSet.Last should be max")
}

func TestToArrayList(t *testing.T) {
	t.Parallel()
	list := ToArrayList(Of("a", "b", "c"))
	assert.Equal(t, 3, list.Size(), "ToArrayList should contain all elements")

	first, ok := list.First()
	assert.True(t, ok, "ArrayList.First should succeed")
	assert.Equal(t, "a", first, "ArrayList.First should be head")

	last, ok := list.Last()
	assert.True(t, ok, "ArrayList.Last should succeed")
	assert.Equal(t, "c", last, "ArrayList.Last should be tail")
}

func TestToLinkedList(t *testing.T) {
	t.Parallel()
	list := ToLinkedList(Of(1, 2, 3))
	assert.Equal(t, 3, list.Size(), "ToLinkedList should contain all elements")
}

func TestToHashMapC(t *testing.T) {
	t.Parallel()
	m := ToHashMapC(
		Of("a", "bb", "ccc"),
		func(s string) int { return len(s) },
		func(s string) string { return s },
	)
	assert.Equal(t, 3, m.Size(), "ToHashMapC should map all elements")

	v1, ok := m.Get(1)
	assert.True(t, ok, "HashMapC should contain key=1")
	assert.Equal(t, "a", v1, "HashMapC value for len=1 should be 'a'")

	v2, ok := m.Get(2)
	assert.True(t, ok, "HashMapC should contain key=2")
	assert.Equal(t, "bb", v2, "HashMapC value for len=2 should be 'bb'")
}

func TestToTreeMapC(t *testing.T) {
	t.Parallel()
	m := ToTreeMapC(
		Of("a", "bb", "ccc"),
		func(s string) int { return len(s) },
		func(s string) string { return s },
		cmp.Compare[int],
	)
	assert.Equal(t, 3, m.Size(), "ToTreeMapC should map all elements")

	firstKey, ok := m.FirstKey()
	assert.True(t, ok, "TreeMap.FirstKey should succeed")
	assert.Equal(t, 1, firstKey, "TreeMap.FirstKey should be min")
}

func TestToHashMap2C(t *testing.T) {
	t.Parallel()
	s2 := PairsOf(
		NewPair("a", 1),
		NewPair("b", 2),
	)
	m := ToHashMap2C(s2)
	assert.Equal(t, 2, m.Size(), "ToHashMap2C should map all pairs")

	v, ok := m.Get("a")
	assert.True(t, ok, "HashMap2C should contain key 'a'")
	assert.Equal(t, 1, v, "HashMap2C value for 'a' should be 1")
}

func TestToTreeMap2C(t *testing.T) {
	t.Parallel()
	s2 := PairsOf(
		NewPair("c", 3),
		NewPair("a", 1),
		NewPair("b", 2),
	)
	m := ToTreeMap2C(s2, cmp.Compare[string])
	assert.Equal(t, 3, m.Size(), "ToTreeMap2C should map all pairs")

	firstKey, ok := m.FirstKey()
	assert.True(t, ok, "TreeMap2C.FirstKey should succeed")
	assert.Equal(t, "a", firstKey, "TreeMap2C.FirstKey should be min")
}

// =============================================================================
// Tests for GroupBy and Frequency variants
// =============================================================================

func TestGroupByToHashMap(t *testing.T) {
	t.Parallel()
	m := GroupByToHashMap(
		Of("apple", "apricot", "banana"),
		func(s string) byte { return s[0] },
	)

	aGroup, ok := m.Get('a')
	assert.True(t, ok, "HashMap from GroupBy should contain group 'a'")
	assert.Len(t, aGroup, 2, "Group 'a' should have 2 elements")

	bGroup, ok := m.Get('b')
	assert.True(t, ok, "HashMap from GroupBy should contain group 'b'")
	assert.Len(t, bGroup, 1, "Group 'b' should have 1 element")
}

func TestGroupByToTreeMap(t *testing.T) {
	t.Parallel()
	m := GroupByToTreeMap(
		Of("apple", "apricot", "banana", "cherry"),
		func(s string) byte { return s[0] },
		cmp.Compare[byte],
	)

	firstKey, ok := m.FirstKey()
	assert.True(t, ok, "TreeMap.FirstKey should succeed")
	assert.Equal(t, byte('a'), firstKey, "TreeMap.FirstKey should be 'a'")

	// Verify grouping with multiple items per key
	aGroup, ok := m.Get('a')
	assert.True(t, ok, "Group 'a' should exist")
	assert.Len(t, aGroup, 2, "Group 'a' should have 2 elements") // apple, apricot
	assert.ElementsMatch(t, []string{"apple", "apricot"}, aGroup, "Group 'a' should contain apple and apricot")
}

func TestHistogramToHashMap(t *testing.T) {
	t.Parallel()
	// HistogramToHashMap is an alias for GroupByToHashMap
	m := HistogramToHashMap(
		Of("a", "bb", "ccc", "dd"),
		func(s string) int { return len(s) },
	)

	len1, ok := m.Get(1)
	assert.True(t, ok, "Bucket len=1 should exist")
	assert.Equal(t, []string{"a"}, len1, "Bucket len=1 should contain 'a'")

	len2, ok := m.Get(2)
	assert.True(t, ok, "Bucket len=2 should exist")
	assert.Len(t, len2, 2, "Bucket len=2 should contain 2 elements")
	assert.ElementsMatch(t, []string{"bb", "dd"}, len2, "Bucket len=2 should contain the 2-letter strings")

	len3, ok := m.Get(3)
	assert.True(t, ok, "Bucket len=3 should exist")
	assert.Equal(t, []string{"ccc"}, len3, "Bucket len=3 should contain 'ccc'")
}

func TestGroupValuesToHashMap(t *testing.T) {
	t.Parallel()
	s2 := PairsOf(
		NewPair("k", 1),
		NewPair("k", 2),
		NewPair("z", 3),
	)
	m := GroupValuesToHashMap(s2)

	kVals, ok := m.Get("k")
	assert.True(t, ok, "Key 'k' should exist")
	assert.Equal(t, []int{1, 2}, kVals, "Values for 'k' should match input")

	zVals, ok := m.Get("z")
	assert.True(t, ok, "Key 'z' should exist")
	assert.Equal(t, []int{3}, zVals, "Values for 'z' should match input")
}

func TestFrequencyToHashMap(t *testing.T) {
	t.Parallel()
	freq := FrequencyToHashMap(Of("a", "b", "a", "c", "a", "b"))

	aCount, ok := freq.Get("a")
	assert.True(t, ok, "Frequency for 'a' should exist")
	assert.Equal(t, 3, aCount, "Frequency for 'a' should be 3")

	bCount, ok := freq.Get("b")
	assert.True(t, ok, "Frequency for 'b' should exist")
	assert.Equal(t, 2, bCount, "Frequency for 'b' should be 2")

	cCount, ok := freq.Get("c")
	assert.True(t, ok, "Frequency for 'c' should exist")
	assert.Equal(t, 1, cCount, "Frequency for 'c' should be 1")
}

// =============================================================================
// Tests for Collectors
// =============================================================================

func TestToHashSetCollector(t *testing.T) {
	t.Parallel()
	set := CollectTo(Of(1, 2, 2, 3), ToHashSetCollector[int]())
	assert.Equal(t, 3, set.Size(), "ToHashSetCollector should de-duplicate")
	assert.True(t, set.Contains(2), "HashSet from collector should contain 2")
}

func TestToTreeSetCollector(t *testing.T) {
	t.Parallel()
	set := CollectTo(Of(3, 1, 2), ToTreeSetCollector(cmp.Compare[int]))
	first, ok := set.First()
	assert.True(t, ok, "TreeSet.First should return a value")
	assert.Equal(t, 1, first, "TreeSet should order elements and return the smallest")
}

func TestToArrayListCollector(t *testing.T) {
	t.Parallel()
	list := CollectTo(Of("a", "b", "c"), ToArrayListCollector[string]())
	assert.Equal(t, 3, list.Size(), "ToArrayListCollector should collect all elements")
}

func TestToHashMapCollector(t *testing.T) {
	t.Parallel()
	m := CollectTo(
		Of("a", "bb"),
		ToHashMapCollector(
			func(s string) int { return len(s) },
			func(s string) string { return s },
		),
	)
	assert.Equal(t, 2, m.Size(), "ToHashMapCollector should collect all mappings")
}

func TestToTreeMapCollector(t *testing.T) {
	t.Parallel()
	m := CollectTo(
		Of("a", "bb"),
		ToTreeMapCollector(
			func(s string) int { return len(s) },
			func(s string) string { return s },
			cmp.Compare[int],
		),
	)
	firstKey, ok := m.FirstKey()
	assert.True(t, ok, "TreeMap from collector should have FirstKey")
	assert.Equal(t, 1, firstKey, "FirstKey should be 1")
}

// =============================================================================
// Tests for convenience functions
// =============================================================================

func TestCollectToSet(t *testing.T) {
	t.Parallel()
	set := CollectToSet(Of(1, 2, 2, 3).Seq())
	assert.Equal(t, 3, set.Size(), "CollectToSet should de-duplicate")
}

func TestCollectToList(t *testing.T) {
	t.Parallel()
	list := CollectToList(Of("a", "b").Seq())
	assert.Equal(t, 2, list.Size(), "CollectToList should collect all elements")
}

func TestCollectToMap(t *testing.T) {
	t.Parallel()
	s2 := PairsOf(NewPair("a", 1), NewPair("b", 2))
	m := CollectToMap(s2.Seq2())
	assert.Equal(t, 2, m.Size(), "CollectToMap should collect all pairs")
}

// =============================================================================
// Tests for Set Operations (via ToHashSet)
// =============================================================================

func TestSetOperations(t *testing.T) {
	t.Parallel()
	set1 := ToHashSet(Of(1, 2, 3, 4))
	set2 := ToHashSet(Of(3, 4, 5, 6))

	// Union
	union := set1.Union(set2)
	assert.Equal(t, 6, union.Size(), "Union size should be 6")

	// Intersection
	inter := set1.Intersection(set2)
	assert.Equal(t, 2, inter.Size(), "Intersection size should be 2")
	assert.True(t, inter.Contains(3), "Intersection should contain 3")
	assert.True(t, inter.Contains(4), "Intersection should contain 4")

	// Difference
	diff := set1.Difference(set2)
	assert.Equal(t, 2, diff.Size(), "Difference size should be 2")
	assert.True(t, diff.Contains(1), "Difference should contain 1")
	assert.True(t, diff.Contains(2), "Difference should contain 2")

	// SymmetricDifference
	symDiff := set1.SymmetricDifference(set2)
	assert.Equal(t, 4, symDiff.Size(), "SymmetricDifference size should be 4")

	// Relations
	assert.False(t, set1.IsSubsetOf(set2), "set1 should not be subset of set2")
	assert.False(t, set1.IsDisjoint(set2), "set1 and set2 should not be disjoint")
	assert.False(t, set1.Equals(set2), "set1 and set2 should not be equal")
}
