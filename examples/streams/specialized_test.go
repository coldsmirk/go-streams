package streams_examples

import (
	"fmt"
	"slices"

	streams "github.com/coldsmirk/go-streams"
)

func Example_intersperse_and_takelast_droplast() {
	fmt.Println(streams.Of("a", "b", "c").Intersperse("-").Collect())
	fmt.Println(streams.Of(1, 2, 3, 4, 5).TakeLast(2).Collect())
	fmt.Println(streams.Of(1, 2, 3, 4, 5).DropLast(2).Collect())
	// Output:
	// [a - b - c]
	// [4 5]
	// [1 2 3]
}

func Example_zipWithIndex_and_zipLongest() {
	zipIdx := streams.ZipWithIndex(streams.Of("x", "y")).ToPairs().Collect()
	zl := streams.ZipLongest(streams.Of(1), streams.Of(10, 20)).Collect()
	zld := streams.ZipLongestWith(streams.Of(1), streams.Of(10, 20), 0, 0).Collect()
	fmt.Println(zipIdx)
	fmt.Println(len(zl), zl[0].First.IsPresent(), zl[1].First.IsPresent())
	fmt.Println(zld)
	// Output:
	// [{0 x} {1 y}]
	// 2 true false
	// [{1 10} {0 20}]
}

func Example_mergeSorted_and_mergeSortedN() {
	ms := streams.MergeSorted(
		streams.Of(1, 3, 5),
		streams.Of(2, 4, 6),
		func(a, b int) int { return a - b },
	).Collect()
	msn := streams.MergeSortedN(
		func(a, b int) int { return a - b },
		streams.Of(1, 4, 7),
		streams.Of(2, 5, 8),
		streams.Of(3, 6, 9),
	).Collect()
	fmt.Println(ms)
	fmt.Println(msn)
	// Output:
	// [1 2 3 4 5 6]
	// [1 2 3 4 5 6 7 8 9]
}

func Example_cartesian_and_combinatorics() {
	cart := streams.Cartesian(streams.Of(1, 2), streams.Of("a")).Collect()
	self := streams.CartesianSelf(streams.Of(1, 2)).Collect()
	cross := streams.CrossProduct(streams.Of(1, 2), streams.Of(3), streams.Of(4, 5)).Collect()
	comb := streams.Combinations(streams.Of(1, 2, 3), 2).Collect()
	perm := streams.Permutations(streams.Of(1, 2, 3)).Collect()
	// Normalize for deterministic presentation where needed
	slices.SortFunc(cross, func(a, b []int) int {
		for i := range a {
			if a[i] != b[i] {
				return a[i] - b[i]
			}
		}
		return 0
	})
	fmt.Println(cart)
	fmt.Println(self)
	fmt.Println(len(cross))
	fmt.Println(comb)
	fmt.Println(len(perm))
	// Output:
	// [{1 a} {2 a}]
	// [{1 1} {1 2} {2 1} {2 2}]
	// 4
	// [[1 2] [1 3] [2 3]]
	// 6
}

