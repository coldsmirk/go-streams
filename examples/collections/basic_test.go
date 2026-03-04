package collections_examples

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/coldsmirk/go-collections"
	"github.com/coldsmirk/go-streams"
)

func Example_fromSet() {
	// Create a HashSet
	set := collections.NewHashSetFrom(3, 1, 4, 1, 5, 9, 2)

	// Create Stream from Set and process
	result := streams.FromSet(set).
		Filter(func(n int) bool { return n > 2 }).
		Collect()
	slices.Sort(result)
	fmt.Println(result)
	// Output:
	// [3 4 5 9]
}

func Example_fromSortedSet() {
	// Create a TreeSet (sorted)
	set := collections.NewTreeSetFrom(cmp.Compare[int], 3, 1, 4, 1, 5)

	// Stream iterates in sorted order
	result := streams.FromSortedSet(set).Collect()
	fmt.Println(result)

	// Descending order
	desc := streams.FromSortedSetDescending(set).Collect()
	fmt.Println(desc)
	// Output:
	// [1 3 4 5]
	// [5 4 3 1]
}

func Example_fromList() {
	// Create an ArrayList
	list := collections.NewArrayListFrom("a", "b", "c", "d")

	// Create Stream from List
	result := streams.FromList(list).
		Map(func(s string) string { return s + "!" }).
		Collect()
	fmt.Println(result)
	// Output:
	// [a! b! c! d!]
}

func Example_fromMapC() {
	// Create a HashMap
	m := collections.NewHashMap[string, int]()
	m.Put("a", 1)
	m.Put("b", 2)
	m.Put("c", 3)

	// Create Stream2 from collections.Map
	keys := streams.FromMapC(m).Keys().Collect()
	slices.Sort(keys)
	fmt.Println(keys)
	// Output:
	// [a b c]
}

func Example_toHashSet() {
	// Collect Stream into a HashSet
	set := streams.ToHashSet(streams.Of(1, 2, 2, 3, 3, 3))
	fmt.Println(set.Size(), set.Contains(2), set.Contains(5))
	// Output:
	// 3 true false
}

func Example_toTreeSet() {
	// Collect Stream into a TreeSet (sorted)
	set := streams.ToTreeSet(streams.Of(3, 1, 4, 1, 5), cmp.Compare[int])
	first, _ := set.First()
	last, _ := set.Last()
	fmt.Println(set.Size(), first, last)
	// Output:
	// 4 1 5
}

func Example_toArrayList() {
	// Collect Stream into an ArrayList
	list := streams.ToArrayList(streams.Of("x", "y", "z"))
	first, _ := list.First()
	last, _ := list.Last()
	fmt.Println(list.Size(), first, last)
	// Output:
	// 3 x z
}

func Example_toHashMapC() {
	// Collect Stream into a HashMap
	m := streams.ToHashMapC(
		streams.Of("apple", "banana", "cherry"),
		func(s string) string { return string(s[0]) }, // first char as key
		func(s string) int { return len(s) },          // length as value
	)
	aVal, _ := m.Get("a")
	bVal, _ := m.Get("b")
	fmt.Println(m.Size(), aVal, bVal)
	// Output:
	// 3 5 6
}

func Example_toHashMap2C() {
	// Convert Stream2 to collections.Map
	s2 := streams.PairsOf(
		streams.NewPair("x", 10),
		streams.NewPair("y", 20),
	)
	m := streams.ToHashMap2C(s2)
	xVal, _ := m.Get("x")
	yVal, _ := m.Get("y")
	fmt.Println(m.Size(), xVal, yVal)
	// Output:
	// 2 10 20
}

func Example_collectors_hashSet() {
	// Use ToHashSetCollector with CollectTo
	set := streams.CollectTo(
		streams.Of(1, 2, 2, 3),
		streams.ToHashSetCollector[int](),
	)
	fmt.Println(set.Size(), set.Contains(2))
	// Output:
	// 3 true
}

func Example_collectors_treeSet() {
	// Use ToTreeSetCollector with CollectTo
	set := streams.CollectTo(
		streams.Of(5, 2, 8, 1),
		streams.ToTreeSetCollector(cmp.Compare[int]),
	)
	first, _ := set.First()
	fmt.Println(set.Size(), first)
	// Output:
	// 4 1
}

func Example_collectors_arrayList() {
	// Use ToArrayListCollector with CollectTo
	list := streams.CollectTo(
		streams.Of("a", "b", "c"),
		streams.ToArrayListCollector[string](),
	)
	fmt.Println(list.Size())
	// Output:
	// 3
}

func Example_collectors_hashMap() {
	// Use ToHashMapCollector with CollectTo
	m := streams.CollectTo(
		streams.Of("a", "bb", "ccc"),
		streams.ToHashMapCollector(
			func(s string) int { return len(s) },
			func(s string) string { return s },
		),
	)
	oneVal, _ := m.Get(1)
	twoVal, _ := m.Get(2)
	threeVal, _ := m.Get(3)
	fmt.Println(m.Size(), oneVal, twoVal, threeVal)
	// Output:
	// 3 a bb ccc
}

func Example_groupByToHashMap() {
	// Group elements into a collections.Map
	m := streams.GroupByToHashMap(
		streams.Of("apple", "apricot", "banana", "cherry"),
		func(s string) byte { return s[0] },
	)
	aGroup, _ := m.Get('a')
	bGroup, _ := m.Get('b')
	fmt.Println(len(aGroup), len(bGroup))
	// Output:
	// 2 1
}

func Example_frequencyToHashMap() {
	// Get frequency counts as collections.Map
	freq := streams.FrequencyToHashMap(streams.Of("a", "b", "a", "c", "a", "b"))
	aCount, _ := freq.Get("a")
	bCount, _ := freq.Get("b")
	cCount, _ := freq.Get("c")
	fmt.Println(aCount, bCount, cCount)
	// Output:
	// 3 2 1
}

func Example_setOperations() {
	// Demonstrate set operations after collecting
	set1 := streams.ToHashSet(streams.Of(1, 2, 3, 4))
	set2 := streams.ToHashSet(streams.Of(3, 4, 5, 6))

	// Union
	union := set1.Union(set2)
	fmt.Println("union size:", union.Size())

	// Intersection
	inter := set1.Intersection(set2)
	result := inter.ToSlice()
	slices.Sort(result)
	fmt.Println("intersection:", result)

	// Difference
	diff := set1.Difference(set2)
	diffResult := diff.ToSlice()
	slices.Sort(diffResult)
	fmt.Println("difference:", diffResult)
	// Output:
	// union size: 6
	// intersection: [3 4]
	// difference: [1 2]
}
