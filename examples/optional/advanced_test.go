package optional_examples

import (
	"fmt"
	"strconv"

	streams "github.com/coldsmirk/go-streams"
)

func Example_optionalAdvanced() {
	o1 := streams.Some(10)
	o2 := streams.None[int]()
	fmt.Println(o1.Map(func(v int) int { return v * 2 }).Get())
	fmt.Println(o2.GetOrElse(42))
	fmt.Println(streams.OptionalZip(streams.Some(1), streams.Some("a")).Get())
	fmt.Println(streams.OptionalFlatMap(streams.Some("3"), func(s string) streams.Optional[int] {
		if v, err := strconv.Atoi(s); err == nil {
			return streams.Some(v)
		}
		return streams.None[int]()
	}).Get())
	// Output:
	// 20
	// 42
	// {1 a}
	// 3
}

