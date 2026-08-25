package streams_test

import (
	"cmp"
	"errors"
	"fmt"
	"iter"
	"maps"
	"slices"
	"strings"

	streams "github.com/coldsmirk/go-streams/v2"
)

type user struct {
	Name string
	Age  int
	City string
}

var users = []user{
	{"Ada", 36, "London"},
	{"Linus", 54, "Portland"},
	{"Rob", 60, "NYC"},
	{"Ken", 81, "NYC"},
}

// A pipeline stays a single chain even where the element type changes.
func Example() {
	lengths := streams.Of(users...).
		Filter(func(u user) bool { return u.Age > 40 }).
		Map(func(u user) string { return u.Name }).
		Map(strings.ToUpper).
		Map(func(s string) int { return len(s) }).
		Collect()

	fmt.Println(lengths)
	// Output: [5 3 3]
}

// A Stream is an iter.Seq, so it may be ranged over directly.
func ExampleStream() {
	for v := range streams.Range(0, 3) {
		fmt.Print(v, " ")
	}
	fmt.Println()
	// Output: 0 1 2
}

// Map is a generic method, so the result element type follows the function.
func ExampleStream_Map() {
	names := streams.Of(1, 2, 3).Map(func(i int) string {
		return strings.Repeat("*", i)
	}).Collect()
	fmt.Println(names)
	// Output: [* ** ***]
}

// Fold accumulates into a type unrelated to the element type.
func ExampleStream_Fold() {
	total := streams.Of(users...).Fold(0, func(sum int, u user) int {
		return sum + u.Age
	})
	fmt.Println(total)
	// Output: 231
}

// MaxFunc works for any element type; Max is its constrained counterpart, in
// the same way slices.MaxFunc pairs with slices.Max.
func ExampleStream_MaxFunc() {
	oldest, ok := streams.Of(users...).MaxFunc(func(a, b user) int {
		return cmp.Compare(a.Age, b.Age)
	})
	fmt.Println(oldest.Name, ok)

	largest, ok := streams.Max(streams.Of(3, 9, 4))
	fmt.Println(largest, ok)
	// Output:
	// Ken true
	// 9 true
}

// Zip yields a Stream2, so the package needs no pair type.
func ExampleStream_Zip() {
	paired := streams.Of("a", "b").Zip(streams.Of(1, 2))
	for k, v := range paired {
		fmt.Printf("%s=%d ", k, v)
	}
	fmt.Println()
	// Output: a=1 b=2
}

// ZipWith fuses the pairing with the mapping and allocates nothing in between.
func ExampleStream_ZipWith() {
	sums := streams.Of(1, 2, 3).ZipWith(streams.Of(10, 20, 30),
		func(a, b int) int { return a + b }).Collect()
	fmt.Println(sums)
	// Output: [11 22 33]
}

func ExampleStream_GroupBy() {
	byCity := streams.Of(users...).GroupBy(func(u user) string { return u.City })
	for _, city := range slices.Sorted(maps.Keys(byCity)) {
		fmt.Printf("%s:%d ", city, len(byCity[city]))
	}
	fmt.Println()
	// Output: London:1 NYC:2 Portland:1
}

// Chunk regroups the sequence, so it is a package function, mirroring
// slices.Chunk. The chain resumes off its result.
func ExampleChunk() {
	sizes := streams.Chunk(streams.Range(0, 5), 2).
		Map(func(c []int) int { return len(c) }).
		Collect()
	fmt.Println(sizes)
	// Output: [2 2 1]
}

// Streams convert to and from standard library iterators at no cost.
func ExampleFrom() {
	m := map[string]int{"b": 2, "a": 1, "c": 3}

	// in: maps.Keys returns an iter.Seq, From infers the element type
	names := streams.From(maps.Keys(m)).
		Filter(func(s string) bool { return s != "b" }).
		SortFunc(strings.Compare).
		Collect()
	fmt.Println(names)

	// out: a Stream converts straight back to an iter.Seq
	fmt.Println(slices.Sorted(iter.Seq[int](streams.Pairs(m).Values())))
	// Output:
	// [a c]
	// [1 2 3]
}

// The package has no error type. Fallible work uses the standard
// iter.Seq2[T, error].
func ExampleTryMap() {
	parse := func(s string) (int, error) {
		if s == "" {
			return 0, errors.New("empty field")
		}
		return len(s), nil
	}

	got, err := streams.Try(streams.TryMap(streams.Of("ab", "cde"), parse))
	fmt.Println(got, err)

	got, err = streams.Try(streams.TryMap(streams.Of("ab", "", "cde"), parse))
	fmt.Println(got, err)
	// Output:
	// [2 3] <nil>
	// [2] empty field
}

// ParallelMap keeps the input order unless Unordered is given.
func ExampleStream_ParallelMap() {
	squares := streams.Range(1, 6).
		ParallelMap(func(i int) int { return i * i }, streams.WithConcurrency(4)).
		Collect()
	fmt.Println(squares)
	// Output: [1 4 9 16 25]
}
