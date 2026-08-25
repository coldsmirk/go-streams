package collections

import (
	"cmp"
	"strings"
	"testing"

	coll "github.com/coldsmirk/go-collections"
	streams "github.com/coldsmirk/go-streams/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	byInt    = coll.CompareFunc[int]()
	byString = coll.CompareFunc[string]()
)

func TestFromCollectionsKeepTheirIterationOrder(t *testing.T) {
	// The hash types iterate in unspecified order, so compare after sorting.
	assert.Equal(t, []int{1, 2, 3}, streams.Sort(FromSet(coll.NewHashSetFrom(3, 1, 2))).Collect(), "FromSet")
	assert.Equal(t, []int{1, 2, 3}, FromSortedSet(coll.NewTreeSetFrom(byInt, 3, 1, 2)).Collect(), "FromSortedSet")

	assert.Equal(t, []int{3, 1, 2}, FromList(coll.NewArrayListFrom(3, 1, 2)).Collect(), "FromList array")
	assert.Equal(t, []int{3, 1, 2}, FromList(coll.NewLinkedListFrom(3, 1, 2)).Collect(), "FromList linked")

	assert.Equal(t, []int{1, 2, 3}, FromQueue(coll.NewArrayQueueFrom(1, 2, 3)).Collect(), "FromQueue")
	assert.Equal(t, []int{3, 2, 1}, FromStack(coll.NewArrayStackFrom(1, 2, 3)).Collect(), "FromStack")
	assert.Equal(t, []int{1, 2, 3}, FromDeque(coll.NewArrayDequeFrom(1, 2, 3)).Collect(), "FromDeque")

	m := coll.NewHashMapFrom(map[string]int{"a": 1, "b": 2})
	assert.Equal(t, []string{"a", "b"}, streams.Sort(FromMap(m).Keys()).Collect(), "FromMap keys")
	assert.Equal(t, []int{1, 2}, streams.Sort(FromMap(m).Values()).Collect(), "FromMap values")

	sm := coll.NewTreeMapFrom(byString, map[string]int{"c": 3, "a": 1, "b": 2})
	assert.Equal(t, []string{"a", "b", "c"}, FromSortedMap(sm).Keys().Collect(), "FromSortedMap keys")
	assert.Equal(t, []int{1, 2, 3}, FromSortedMap(sm).Values().Collect(), "FromSortedMap values")
}

func TestFromPriorityQueueYieldsHeapOrder(t *testing.T) {
	q := coll.NewPriorityQueueFrom(byInt, 3, 1, 2)

	// Only the head is ordered; the rest is the internal heap layout.
	head, ok := FromPriorityQueue(q).First()
	want, _ := q.Peek()
	assert.True(t, ok, "FromPriorityQueue head")
	assert.Equal(t, want, head, "FromPriorityQueue head")
	assert.Equal(t, []int{1, 2, 3}, streams.Sort(FromPriorityQueue(q)).Collect(), "FromPriorityQueue sorted")

	// Iterating must not consume the queue.
	assert.Equalf(t, 3, q.Size(), "FromPriorityQueue drained the queue: size = %d, want 3", q.Size())
}

func TestFromEmptyCollections(t *testing.T) {
	assert.Empty(t, FromSet(coll.NewHashSet[int]()).Collect(), "FromSet")
	assert.Empty(t, FromSortedSet(coll.NewTreeSet(byInt)).Collect(), "FromSortedSet")
	assert.Empty(t, FromList(coll.NewArrayList[int]()).Collect(), "FromList array")
	assert.Empty(t, FromList(coll.NewLinkedList[int]()).Collect(), "FromList linked")
	assert.Empty(t, FromQueue(coll.NewArrayQueue[int]()).Collect(), "FromQueue")
	assert.Empty(t, FromStack(coll.NewArrayStack[int]()).Collect(), "FromStack")
	assert.Empty(t, FromDeque(coll.NewArrayDeque[int]()).Collect(), "FromDeque")
	assert.Empty(t, FromPriorityQueue(coll.NewPriorityQueue(byInt)).Collect(), "FromPriorityQueue")

	gotMap := FromMap(coll.NewHashMap[string, int]()).Count()
	assert.Zerof(t, gotMap, "FromMap on an empty map yielded %d pairs", gotMap)
	gotSortedMap := FromSortedMap(coll.NewTreeMap[string, int](byString)).Count()
	assert.Zerof(t, gotSortedMap, "FromSortedMap on an empty map yielded %d pairs", gotSortedMap)
}

func TestToCollections(t *testing.T) {
	set := ToHashSet(streams.Of(1, 2, 2, 3))
	assert.Equal(t, 3, set.Size(), "ToHashSet size")
	assert.Truef(t, set.ContainsAll(1, 2, 3), "ToHashSet = %v, want the three distinct elements", set)
	assert.Equal(t, []int{1, 2, 3}, ToTreeSet(streams.Of(3, 1, 2, 1), byInt).ToSlice(), "ToTreeSet")

	assert.Equal(t, []int{3, 1, 2}, ToArrayList(streams.Of(3, 1, 2)).ToSlice(), "ToArrayList")
	assert.Equal(t, []int{3, 1, 2}, ToLinkedList(streams.Of(3, 1, 2)).ToSlice(), "ToLinkedList")
	assert.Equal(t, []int{1, 1}, ToArrayList(streams.Of(1, 1)).ToSlice(), "ToArrayList keeps duplicates")

	// A later pair overwrites an earlier one with the same key.
	m := ToHashMap(streams.Of("a", "b", "a").Zip(streams.Of(1, 2, 3)))
	assert.Equalf(t, 2, m.Size(), "ToHashMap size = %d, want 2", m.Size())
	v, _ := m.Get("a")
	assert.Equalf(t, 3, v, "ToHashMap a = %d, want the later pair 3", v)

	sm := ToTreeMap(streams.Of("b", "a", "b").Zip(streams.Of(1, 2, 3)), byString)
	assert.Equal(t, []string{"a", "b"}, sm.Keys(), "ToTreeMap keys")
	v, _ = sm.Get("b")
	assert.Equalf(t, 3, v, "ToTreeMap b = %d, want the later pair 3", v)
}

func TestToEmptyCollections(t *testing.T) {
	hashSet := ToHashSet(streams.Empty[int]())
	assert.Truef(t, hashSet.IsEmpty(), "ToHashSet of an empty Stream = %v", hashSet)
	treeSet := ToTreeSet(streams.Empty[int](), byInt)
	assert.Truef(t, treeSet.IsEmpty(), "ToTreeSet of an empty Stream = %v", treeSet)
	arrayList := ToArrayList(streams.Empty[int]())
	assert.Truef(t, arrayList.IsEmpty(), "ToArrayList of an empty Stream = %v", arrayList)
	linkedList := ToLinkedList(streams.Empty[int]())
	assert.Truef(t, linkedList.IsEmpty(), "ToLinkedList of an empty Stream = %v", linkedList)
	hashMap := ToHashMap(streams.Empty2[string, int]())
	assert.Truef(t, hashMap.IsEmpty(), "ToHashMap of an empty Stream2 = %v", hashMap)
	treeMap := ToTreeMap(streams.Empty2[string, int](), byString)
	assert.Truef(t, treeMap.IsEmpty(), "ToTreeMap of an empty Stream2 = %v", treeMap)
}

func TestRoundTripThroughAStream(t *testing.T) {
	list := coll.NewArrayListFrom(3, 1, 2)
	assert.Equal(t, list.ToSlice(), ToArrayList(FromList(list)).ToSlice(), "list")
	assert.Equal(t, list.ToSlice(), ToLinkedList(FromList(list)).ToSlice(), "list to linked")

	set := coll.NewHashSetFrom(1, 2, 3)
	backSet := ToHashSet(FromSet(set))
	assert.Truef(t, backSet.Equals(set), "hash set round trip = %v, want %v", backSet, set)
	assert.Equal(t, []int{1, 2, 3},
		ToTreeSet(FromSortedSet(coll.NewTreeSetFrom(byInt, 3, 1, 2)), byInt).ToSlice(), "tree set")

	m := coll.NewHashMapFrom(map[string]int{"a": 1, "b": 2})
	backMap := ToHashMap(FromMap(m))
	assert.Truef(t, backMap.Equals(m, coll.EqualFunc[int]()), "hash map round trip = %v, want %v", backMap, m)
	back := ToTreeMap(FromSortedMap(coll.NewTreeMapFrom(byString, map[string]int{"b": 2, "a": 1})), byString)
	assert.Equal(t, []string{"a", "b"}, back.Keys(), "tree map keys")
	assert.Equal(t, []int{1, 2}, back.Values(), "tree map values")

	// The queue-like types have no To counterpart; their order survives in a list.
	assert.Equal(t, []int{1, 2, 3}, ToArrayList(FromQueue(coll.NewArrayQueueFrom(1, 2, 3))).ToSlice(), "queue")
	assert.Equal(t, []int{3, 2, 1}, ToArrayList(FromStack(coll.NewArrayStackFrom(1, 2, 3))).ToSlice(), "stack")
	assert.Equal(t, []int{1, 2, 3}, ToArrayList(FromDeque(coll.NewArrayDequeFrom(1, 2, 3))).ToSlice(), "deque")
	assert.Equal(t, []int{1, 2, 3},
		ToTreeSet(FromPriorityQueue(coll.NewPriorityQueueFrom(byInt, 3, 1, 2)), byInt).ToSlice(),
		"priority queue")
}

func TestSortedTypesUseTheirComparator(t *testing.T) {
	desc := coll.Comparator[int](func(a, b int) int { return cmp.Compare(b, a) })

	assert.Equal(t, []int{3, 2, 1}, FromSortedSet(coll.NewTreeSetFrom(desc, 1, 3, 2)).Collect(), "FromSortedSet")
	assert.Equal(t, []int{3, 2, 1},
		FromSortedMap(coll.NewTreeMapFrom(desc, map[int]string{1: "a", 2: "b", 3: "c"})).Keys().Collect(),
		"FromSortedMap")
	assert.Equal(t, []int{3, 2, 1}, ToTreeSet(streams.Of(1, 3, 2), desc).ToSlice(), "ToTreeSet")
	assert.Equal(t, []int{3, 2, 1}, ToTreeMap(streams.Of(1, 3, 2).Zip(streams.Of("a", "b", "c")), desc).Keys(),
		"ToTreeMap")
}

func TestFromReadsTheCollectionLazily(t *testing.T) {
	// The Stream pulls elements one at a time rather than materialising the
	// collection up front.
	consumed := 0
	FromList(coll.NewArrayListFrom(1, 2, 3, 4, 5)).
		Peek(func(int) { consumed++ }).
		Take(2).
		ForEach(func(int) {})
	assert.Equalf(t, 2, consumed, "consumed %d elements for a Take(2), want 2", consumed)

	// The Stream iterates the collection itself, so an element added before the
	// Stream is consumed is still seen.
	l := coll.NewLinkedListFrom(1, 2)
	s := FromList(l)
	l.Add(3)
	assert.Equal(t, []int{1, 2, 3}, s.Collect(), "live view")
}

// The liveness contract has to hold for every implementation behind the
// interface, not just the ones that re-walk a linked structure. The earlier
// test used only NewLinkedListFrom, which happens to be live either way; the
// slice-backed implementations were snapshotted at construction and it passed
// anyway.
func TestFromListIsLiveForEveryImplementation(t *testing.T) {
	for name, make := range map[string]func() coll.List[int]{
		"arrayList":     func() coll.List[int] { return coll.NewArrayListFrom(1, 2) },
		"linkedList":    func() coll.List[int] { return coll.NewLinkedListFrom(1, 2) },
		"cowList":       func() coll.List[int] { return coll.NewCOWListFrom(1, 2) },
		"segmentedList": func() coll.List[int] { return coll.NewSegmentedListFrom(1, 2) },
	} {
		t.Run(name, func(t *testing.T) {
			l := make()
			s := FromList(l)
			l.Add(3)
			assert.Equal(t, []int{1, 2, 3}, s.Collect(), "live view")
		})
	}
}

func TestFromQueueIsLive(t *testing.T) {
	q := coll.NewArrayQueueFrom(1, 2)
	s := FromQueue(q)
	q.Enqueue(3)
	assert.Equal(t, []int{1, 2, 3}, s.Collect(), "live view")
}

type keyed struct {
	Key string
	N   int
}

func byKey(a, b keyed) int {
	switch {
	case a.Key < b.Key:
		return -1
	case a.Key > b.Key:
		return 1
	}
	return 0
}

// The tie-break among comparator-equal elements is only observable for a type
// whose comparator ignores part of the value, which is exactly what ToTreeSet's
// comparator parameter is for. Every earlier test used ints, where equal means
// identical, so the documented behaviour was untested.
func TestToTreeSetKeepsTheFirstOfEachEqualGroup(t *testing.T) {
	set := ToTreeSet(streams.Of(
		keyed{"a", 1}, keyed{"a", 2}, keyed{"b", 3}, keyed{"b", 4},
	), byKey)
	got := set.ToSlice()
	require.Lenf(t, got, 2, "ToTreeSet = %v, want the first of each group: [{a 1} {b 3}]", got)
	assert.Equal(t, 1, got[0].N, "ToTreeSet keeps the first of the a group")
	assert.Equal(t, 3, got[1].N, "ToTreeSet keeps the first of the b group")
}

// ToTreeMap keeps the last, by the same convention as assigning to a Go map.
// The two are deliberately different; pin both so neither drifts.
func TestToTreeMapKeepsTheLastOfEachEqualKey(t *testing.T) {
	m := ToTreeMap(streams.Of("a", "a").Zip(streams.Of(1, 2)),
		strings.Compare)
	v, _ := m.Get("a")
	assert.Equalf(t, 2, v, "ToTreeMap kept %d, want the last value 2", v)
}
