package optional_examples

import (
	"fmt"

	streams "github.com/coldsmirk/go-streams"
)

func Example_optional_more() {
	v := 7
	o := streams.OptionalOf(&v)
	o2 := streams.OptionalFromCondition(true, 9)
	var nilPtr *int
	o3 := streams.OptionalOf(nilPtr)
	fmt.Println(o.Get(), o2.Get(), o3.IsEmpty())
	// IfPresent / IfPresentOrElse / GetOrElseGet
	sum := 0
	o.IfPresent(func(x int) { sum += x })
	o3.IfPresentOrElse(func(_ int) { sum += 100 }, func() { sum += 1 })
	def := streams.Some(0).GetOrElseGet(func() int { return 42 })
	fmt.Println(sum, def)
	// ToSlice / ToPointer / ToStream / OptionalEquals
	fmt.Println(len(o.ToSlice()), o.ToPointer() != nil, o.ToStream().Count(), streams.OptionalEquals(o, o))
	// Output:
	// 7 9 true
	// 8 0
	// 1 true 1 true
}

