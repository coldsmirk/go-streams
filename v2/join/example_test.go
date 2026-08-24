package join_test

import (
	"fmt"

	streams "github.com/coldsmirk/go-streams/v2"
	"github.com/coldsmirk/go-streams/v2/join"
)

// A join takes a combiner and returns a Stream of what it produces, so the
// package declares no result type. A key carried by several rows on both sides
// yields the cartesian product for that key.
func ExampleInner() {
	people := streams.Of("ada", "rob").Zip(streams.Of(36, 60))
	languages := streams.Of("ada", "ada", "rob").Zip(streams.Of("Go", "Lisp", "C"))

	got := join.Inner(people, languages, func(name string, age int, lang string) string {
		return fmt.Sprintf("%s(%d):%s", name, age, lang)
	}).Collect()
	fmt.Println(got)
	// Output: [ada(36):Go ada(36):Lisp rob(60):C]
}

// An outer join hands the combiner a presence flag beside the value it
// describes. Where the flag is false the value is the zero value of its type.
func ExampleLeft() {
	people := streams.Of("ada", "ken").Zip(streams.Of(36, 81))
	languages := streams.Of("ada").Zip(streams.Of("Go"))

	got := join.Left(people, languages, func(name string, age int, lang string, matched bool) string {
		if !matched {
			lang = "none"
		}
		return name + ":" + lang
	}).Collect()
	fmt.Println(got)
	// Output: [ada:Go ken:none]
}

// Full streams the left side and replays the rows of the right side it did not
// match afterwards, so "ken" comes last however early the right side carries
// it.
func ExampleFull() {
	ages := streams.Of("ada").Zip(streams.Of(36))
	languages := streams.Of("ken", "ada").Zip(streams.Of("C", "Go"))

	join.Full(ages, languages,
		func(name string, age int, hasAge bool, lang string, hasLang bool) string {
			row := name
			if hasAge {
				row += fmt.Sprintf(" %d", age)
			}
			if hasLang {
				row += " " + lang
			}
			return row
		}).ForEach(func(row string) { fmt.Println(row) })
	// Output:
	// ada 36 Go
	// ken C
}

// Group pairs each key with the rows carrying it on either side. A key that
// only one side carries is combined with a nil slice for the other.
func ExampleGroup() {
	people := streams.Of("NYC", "London", "NYC").Zip(streams.Of("Rob", "Ada", "Ken"))
	offices := streams.Of("NYC", "Berlin").Zip(streams.Of(8, 3))

	join.Group(people, offices, func(city string, names []string, floors []int) string {
		return fmt.Sprintf("%s %v %v", city, names, floors)
	}).ForEach(func(row string) { fmt.Println(row) })
	// Output:
	// NYC [Rob Ken] [8]
	// London [Ada] []
	// Berlin [] [3]
}

// Semi and Anti filter the left side by whether the right side carries its key
// and yield no value of the right side, so their result is still a Stream2.
func ExampleAnti() {
	stock := streams.Of("apple", "pear", "plum").Zip(streams.Of(3, 6, 7))
	recalled := streams.Of("pear", "fig").Zip(streams.Of("mould", "unknown"))

	for name, n := range join.Anti(stock, recalled) {
		fmt.Println(name, n)
	}
	// Output:
	// apple 3
	// plum 7
}

// Two unkeyed streams join by deriving their keys with Stream.KeyBy first.
// Every join in this package accepts them the same way, so a left or full join
// over unkeyed streams needs nothing the package does not already have.
func Example_derivedKeys() {
	type sale struct {
		city  string
		total int
	}
	cities := streams.Of("London", "NYC", "Oslo")
	sales := streams.Of(sale{"NYC", 20}, sale{"London", 5}, sale{"NYC", 7})

	byCity := func(s sale) string { return s.city }
	itself := func(c string) string { return c }

	inner := join.Inner(cities.KeyBy(itself), sales.KeyBy(byCity),
		func(c string, _ string, s sale) string { return fmt.Sprintf("%s:%d", c, s.total) },
	).Collect()
	fmt.Println(inner)

	// Oslo made no sales, and a left join keeps it.
	left := join.Left(cities.KeyBy(itself), sales.KeyBy(byCity),
		func(c string, _ string, s sale, sold bool) string {
			if !sold {
				return c + ":none"
			}
			return fmt.Sprintf("%s:%d", c, s.total)
		},
	).Collect()
	fmt.Println(left)
	// Output:
	// [London:5 NYC:20 NYC:7]
	// [London:5 NYC:20 NYC:7 Oslo:none]
}
