package streams_test

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strconv"

	"github.com/coldsmirk/go-streams"
)

type user struct {
	Name   string
	Age    int
	Active bool
}

var users = []user{
	{"Alice", 30, true},
	{"Bob", 25, false},
	{"Charlie", 35, true},
}

// Operations that leave the element type alone chain as methods.
func Example() {
	evens := streams.Of(1, 2, 3, 4, 5).
		Filter(func(n int) bool { return n%2 == 0 }).
		Map(func(n int) int { return n * 2 }).
		Collect()
	fmt.Println(evens)
	// Output: [4 8]
}

// Changing the element type needs a free function: Go 1.25 has no type
// parameters on methods, so MapTo cannot be a method on Stream[T].
func Example_typeChanging() {
	names := streams.MapTo(
		streams.FromSlice(users).Filter(func(u user) bool { return u.Active && u.Age > 25 }),
		func(u user) string { return u.Name },
	).Collect()
	fmt.Println(names)
	// Output: [Alice Charlie]
}

// GroupBy and CountBy answer the same question at two levels of detail.
func Example_groupAndCount() {
	words := []string{"apple", "apricot", "banana", "blueberry", "cherry"}
	initial := func(s string) string { return s[:1] }

	grouped := streams.GroupBy(streams.FromSlice(words), initial)
	counts := streams.CountBy(streams.FromSlice(words), initial)

	for _, k := range slices.Sorted(maps.Keys(grouped)) {
		fmt.Printf("%s %v (%d)\n", k, grouped[k], counts[k])
	}
	// Output:
	// a [apple apricot] (2)
	// b [banana blueberry] (2)
	// c [cherry] (1)
}

// Infinite sources are bounded by the pipeline, not by the source.
func Example_infiniteStream() {
	powers := streams.Iterate(1, func(n int) int { return n * 2 }).Limit(8).Collect()
	fmt.Println(powers)

	a, b := 0, 1
	fib := streams.Generate(func() int {
		a, b = b, a+b
		return a
	}).Limit(8).Collect()
	fmt.Println(fib)
	// Output:
	// [1 2 4 8 16 32 64 128]
	// [1 1 2 3 5 8 13 21]
}

// Fallible work travels as Result[T]; CollectResults stops at the first error.
func Example_errorHandling() {
	parse := func(s string) streams.Result[int] {
		n, err := strconv.Atoi(s)
		if err != nil {
			return streams.Err[int](err)
		}
		return streams.Ok(n)
	}

	good, err := streams.CollectResults(streams.MapTo(streams.Of("1", "2", "3"), parse))
	fmt.Println(good, err)

	_, err = streams.CollectResults(streams.MapTo(streams.Of("1", "x", "3"), parse))
	fmt.Println(err)
	// Output:
	// [1 2 3] <nil>
	// strconv.Atoi: parsing "x": invalid syntax
}

// A Stream carries an iter.Seq, which Seq hands back to the standard library.
func Example_stdlibInterop() {
	fmt.Println(slices.Sorted(streams.Of(3, 1, 2).Seq()))

	for v := range streams.Range(0, 3).Seq() {
		fmt.Print(v, " ")
	}
	fmt.Println()
	// Output:
	// [1 2 3]
	// 0 1 2
}

// GetStatistics makes one pass and returns every basic statistic together.
func Example_statistics() {
	ages := streams.MapTo(streams.FromSlice(users), func(u user) int { return u.Age })

	s := streams.GetStatistics(ages).Get()
	fmt.Printf("n=%d sum=%d range=%d..%d avg=%.2f\n",
		s.Count, s.Sum, s.Min, s.Max, s.Average)
	// Output: n=3 sum=90 range=25..35 avg=30.00
}

// TopK ranks without sorting the whole stream; less reports a < b.
func Example_topK() {
	oldest := streams.TopK(streams.FromSlice(users), 2,
		func(a, b user) bool { return a.Age < b.Age })

	for _, u := range oldest {
		fmt.Println(u.Name, u.Age)
	}
	// Output:
	// Charlie 35
	// Alice 30
}

// Joins run over Stream2, which FromMap builds from a keyed collection.
func Example_join() {
	units := streams.FromMap(map[string]int{"west": 15, "east": 7})
	leads := streams.FromMap(map[string]string{"west": "Alice", "east": "Bob"})

	rows := streams.InnerJoin(units, leads).Collect()
	slices.SortFunc(rows, func(x, y streams.JoinResult[string, int, string]) int {
		return cmp.Compare(x.Key, y.Key)
	})
	for _, r := range rows {
		fmt.Printf("%s %s %d\n", r.Key, r.Right, r.Left)
	}
	// Output:
	// east Bob 7
	// west Alice 15
}

// Optional carries "maybe absent" out of a terminal without a sentinel value.
func Example_optional() {
	missing := streams.FromSlice(users).
		FindFirst(func(u user) bool { return u.Age > 100 })

	fmt.Println(missing.IsPresent(), missing.GetOrElse(user{Name: "nobody"}).Name)
	// Output: false nobody
}
