package collections_test

import (
	"fmt"
	"strings"

	coll "github.com/coldsmirk/go-collections"
	streams "github.com/coldsmirk/go-streams/v2"
	"github.com/coldsmirk/go-streams/v2/collections"
)

// The bridge holds only the two conversions; the work in between is ordinary
// Stream work.
func Example() {
	names := coll.NewArrayListFrom("ada", "linus", "rob")

	upper := collections.ToTreeSet(
		collections.FromList(names).
			Filter(func(s string) bool { return s != "rob" }).
			Map(strings.ToUpper),
		strings.Compare,
	)

	fmt.Println(upper.ToSlice())
	// Output: [ADA LINUS]
}

// A Stream2 is an iter.Seq2, so the pairs of a map may be ranged over directly.
func ExampleFromSortedMap() {
	m := coll.NewTreeMapOrdered[string, int]()
	m.Put("b", 2)
	m.Put("a", 1)
	m.Put("c", 3)

	for k, v := range collections.FromSortedMap(m) {
		fmt.Printf("%s=%d ", k, v)
	}
	fmt.Println()
	// Output: a=1 b=2 c=3
}

// A Stream2 collects into a map without an intermediate pair type.
func ExampleToHashMap() {
	m := collections.ToHashMap(streams.Of("ada", "linus").Enumerate().Swap())

	fmt.Println(m.Size())
	fmt.Println(m.GetOrDefault("linus", -1))
	// Output:
	// 2
	// 1
}
