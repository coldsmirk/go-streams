package collections

import (
	"cmp"
	"slices"
	"strings"
	"testing"

	coll "github.com/coldsmirk/go-collections"
	streams "github.com/coldsmirk/go-streams/v2"
)

var (
	byInt    = coll.CompareFunc[int]()
	byString = coll.CompareFunc[string]()
)

func eq[T comparable](t *testing.T, name string, got, want []T) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Errorf("%s = %v, want %v", name, got, want)
	}
}

func TestFromCollectionsKeepTheirIterationOrder(t *testing.T) {
	// The hash types iterate in unspecified order, so compare after sorting.
	eq(t, "FromSet", streams.Sort(FromSet(coll.NewHashSetFrom(3, 1, 2))).Collect(), []int{1, 2, 3})
	eq(t, "FromSortedSet", FromSortedSet(coll.NewTreeSetFrom(byInt, 3, 1, 2)).Collect(), []int{1, 2, 3})

	eq(t, "FromList array", FromList(coll.NewArrayListFrom(3, 1, 2)).Collect(), []int{3, 1, 2})
	eq(t, "FromList linked", FromList(coll.NewLinkedListFrom(3, 1, 2)).Collect(), []int{3, 1, 2})

	eq(t, "FromQueue", FromQueue(coll.NewArrayQueueFrom(1, 2, 3)).Collect(), []int{1, 2, 3})
	eq(t, "FromStack", FromStack(coll.NewArrayStackFrom(1, 2, 3)).Collect(), []int{3, 2, 1})
	eq(t, "FromDeque", FromDeque(coll.NewArrayDequeFrom(1, 2, 3)).Collect(), []int{1, 2, 3})

	m := coll.NewHashMapFrom(map[string]int{"a": 1, "b": 2})
	eq(t, "FromMap keys", streams.Sort(FromMap(m).Keys()).Collect(), []string{"a", "b"})
	eq(t, "FromMap values", streams.Sort(FromMap(m).Values()).Collect(), []int{1, 2})

	sm := coll.NewTreeMapFrom(byString, map[string]int{"c": 3, "a": 1, "b": 2})
	eq(t, "FromSortedMap keys", FromSortedMap(sm).Keys().Collect(), []string{"a", "b", "c"})
	eq(t, "FromSortedMap values", FromSortedMap(sm).Values().Collect(), []int{1, 2, 3})
}

func TestFromPriorityQueueYieldsHeapOrder(t *testing.T) {
	q := coll.NewPriorityQueueFrom(byInt, 3, 1, 2)

	// Only the head is ordered; the rest is the internal heap layout.
	head, ok := FromPriorityQueue(q).First()
	want, _ := q.Peek()
	if !ok || head != want {
		t.Errorf("FromPriorityQueue head = %d, %v, want %d", head, ok, want)
	}
	eq(t, "FromPriorityQueue sorted", streams.Sort(FromPriorityQueue(q)).Collect(), []int{1, 2, 3})

	// Iterating must not consume the queue.
	if q.Size() != 3 {
		t.Errorf("FromPriorityQueue drained the queue: size = %d, want 3", q.Size())
	}
}

func TestFromEmptyCollections(t *testing.T) {
	eq(t, "FromSet", FromSet(coll.NewHashSet[int]()).Collect(), nil)
	eq(t, "FromSortedSet", FromSortedSet(coll.NewTreeSet(byInt)).Collect(), nil)
	eq(t, "FromList array", FromList(coll.NewArrayList[int]()).Collect(), nil)
	eq(t, "FromList linked", FromList(coll.NewLinkedList[int]()).Collect(), nil)
	eq(t, "FromQueue", FromQueue(coll.NewArrayQueue[int]()).Collect(), nil)
	eq(t, "FromStack", FromStack(coll.NewArrayStack[int]()).Collect(), nil)
	eq(t, "FromDeque", FromDeque(coll.NewArrayDeque[int]()).Collect(), nil)
	eq(t, "FromPriorityQueue", FromPriorityQueue(coll.NewPriorityQueue(byInt)).Collect(), nil)

	if got := FromMap(coll.NewHashMap[string, int]()).Count(); got != 0 {
		t.Errorf("FromMap on an empty map yielded %d pairs", got)
	}
	if got := FromSortedMap(coll.NewTreeMap[string, int](byString)).Count(); got != 0 {
		t.Errorf("FromSortedMap on an empty map yielded %d pairs", got)
	}
}

func TestToCollections(t *testing.T) {
	set := ToHashSet(streams.Of(1, 2, 2, 3))
	if set.Size() != 3 || !set.ContainsAll(1, 2, 3) {
		t.Errorf("ToHashSet = %v, want the three distinct elements", set)
	}
	eq(t, "ToTreeSet", ToTreeSet(streams.Of(3, 1, 2, 1), byInt).ToSlice(), []int{1, 2, 3})

	eq(t, "ToArrayList", ToArrayList(streams.Of(3, 1, 2)).ToSlice(), []int{3, 1, 2})
	eq(t, "ToLinkedList", ToLinkedList(streams.Of(3, 1, 2)).ToSlice(), []int{3, 1, 2})
	eq(t, "ToArrayList keeps duplicates", ToArrayList(streams.Of(1, 1)).ToSlice(), []int{1, 1})

	// A later pair overwrites an earlier one with the same key.
	m := ToHashMap(streams.Of("a", "b", "a").Zip(streams.Of(1, 2, 3)))
	if m.Size() != 2 {
		t.Errorf("ToHashMap size = %d, want 2", m.Size())
	}
	if v, _ := m.Get("a"); v != 3 {
		t.Errorf("ToHashMap a = %d, want the later pair 3", v)
	}

	sm := ToTreeMap(streams.Of("b", "a", "b").Zip(streams.Of(1, 2, 3)), byString)
	eq(t, "ToTreeMap keys", sm.Keys(), []string{"a", "b"})
	if v, _ := sm.Get("b"); v != 3 {
		t.Errorf("ToTreeMap b = %d, want the later pair 3", v)
	}
}

func TestToEmptyCollections(t *testing.T) {
	if got := ToHashSet(streams.Empty[int]()); !got.IsEmpty() {
		t.Errorf("ToHashSet of an empty Stream = %v", got)
	}
	if got := ToTreeSet(streams.Empty[int](), byInt); !got.IsEmpty() {
		t.Errorf("ToTreeSet of an empty Stream = %v", got)
	}
	if got := ToArrayList(streams.Empty[int]()); !got.IsEmpty() {
		t.Errorf("ToArrayList of an empty Stream = %v", got)
	}
	if got := ToLinkedList(streams.Empty[int]()); !got.IsEmpty() {
		t.Errorf("ToLinkedList of an empty Stream = %v", got)
	}
	if got := ToHashMap(streams.Empty2[string, int]()); !got.IsEmpty() {
		t.Errorf("ToHashMap of an empty Stream2 = %v", got)
	}
	if got := ToTreeMap(streams.Empty2[string, int](), byString); !got.IsEmpty() {
		t.Errorf("ToTreeMap of an empty Stream2 = %v", got)
	}
}

func TestRoundTripThroughAStream(t *testing.T) {
	list := coll.NewArrayListFrom(3, 1, 2)
	eq(t, "list", ToArrayList(FromList(list)).ToSlice(), list.ToSlice())
	eq(t, "list to linked", ToLinkedList(FromList(list)).ToSlice(), list.ToSlice())

	set := coll.NewHashSetFrom(1, 2, 3)
	if back := ToHashSet(FromSet(set)); !back.Equals(set) {
		t.Errorf("hash set round trip = %v, want %v", back, set)
	}
	eq(t, "tree set", ToTreeSet(FromSortedSet(coll.NewTreeSetFrom(byInt, 3, 1, 2)), byInt).ToSlice(),
		[]int{1, 2, 3})

	m := coll.NewHashMapFrom(map[string]int{"a": 1, "b": 2})
	if back := ToHashMap(FromMap(m)); !back.Equals(m, coll.EqualFunc[int]()) {
		t.Errorf("hash map round trip = %v, want %v", back, m)
	}
	back := ToTreeMap(FromSortedMap(coll.NewTreeMapFrom(byString, map[string]int{"b": 2, "a": 1})), byString)
	eq(t, "tree map keys", back.Keys(), []string{"a", "b"})
	eq(t, "tree map values", back.Values(), []int{1, 2})

	// The queue-like types have no To counterpart; their order survives in a list.
	eq(t, "queue", ToArrayList(FromQueue(coll.NewArrayQueueFrom(1, 2, 3))).ToSlice(), []int{1, 2, 3})
	eq(t, "stack", ToArrayList(FromStack(coll.NewArrayStackFrom(1, 2, 3))).ToSlice(), []int{3, 2, 1})
	eq(t, "deque", ToArrayList(FromDeque(coll.NewArrayDequeFrom(1, 2, 3))).ToSlice(), []int{1, 2, 3})
	eq(t, "priority queue",
		ToTreeSet(FromPriorityQueue(coll.NewPriorityQueueFrom(byInt, 3, 1, 2)), byInt).ToSlice(),
		[]int{1, 2, 3})
}

func TestSortedTypesUseTheirComparator(t *testing.T) {
	desc := coll.Comparator[int](func(a, b int) int { return cmp.Compare(b, a) })

	eq(t, "FromSortedSet", FromSortedSet(coll.NewTreeSetFrom(desc, 1, 3, 2)).Collect(), []int{3, 2, 1})
	eq(t, "FromSortedMap",
		FromSortedMap(coll.NewTreeMapFrom(desc, map[int]string{1: "a", 2: "b", 3: "c"})).Keys().Collect(),
		[]int{3, 2, 1})
	eq(t, "ToTreeSet", ToTreeSet(streams.Of(1, 3, 2), desc).ToSlice(), []int{3, 2, 1})
	eq(t, "ToTreeMap", ToTreeMap(streams.Of(1, 3, 2).Zip(streams.Of("a", "b", "c")), desc).Keys(),
		[]int{3, 2, 1})
}

func TestFromReadsTheCollectionLazily(t *testing.T) {
	// The Stream pulls elements one at a time rather than materialising the
	// collection up front.
	consumed := 0
	FromList(coll.NewArrayListFrom(1, 2, 3, 4, 5)).
		Peek(func(int) { consumed++ }).
		Take(2).
		ForEach(func(int) {})
	if consumed != 2 {
		t.Errorf("consumed %d elements for a Take(2), want 2", consumed)
	}

	// The Stream iterates the collection itself, so an element added before the
	// Stream is consumed is still seen.
	l := coll.NewLinkedListFrom(1, 2)
	s := FromList(l)
	l.Add(3)
	eq(t, "live view", s.Collect(), []int{1, 2, 3})
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
			eq(t, "live view", s.Collect(), []int{1, 2, 3})
		})
	}
}

func TestFromQueueIsLive(t *testing.T) {
	q := coll.NewArrayQueueFrom(1, 2)
	s := FromQueue(q)
	q.Enqueue(3)
	eq(t, "live view", s.Collect(), []int{1, 2, 3})
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
	if len(got) != 2 || got[0].N != 1 || got[1].N != 3 {
		t.Errorf("ToTreeSet = %v, want the first of each group: [{a 1} {b 3}]", got)
	}
}

// ToTreeMap keeps the last, by the same convention as assigning to a Go map.
// The two are deliberately different; pin both so neither drifts.
func TestToTreeMapKeepsTheLastOfEachEqualKey(t *testing.T) {
	m := ToTreeMap(streams.Of("a", "a").Zip(streams.Of(1, 2)),
		strings.Compare)
	if v, _ := m.Get("a"); v != 2 {
		t.Errorf("ToTreeMap kept %d, want the last value 2", v)
	}
}
