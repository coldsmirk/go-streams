package tuples_examples

import (
	"fmt"

	streams "github.com/coldsmirk/go-streams"
)

func Example_tuple_mappers_and_unzip() {
	p := streams.NewPair(1, 2).MapFirst(func(v int) int { return v * 10 }).MapSecond(func(v int) int { return v + 1 })
	t := streams.NewTriple(1, 2, 3).MapFirst(func(v int) int { return v + 1 }).MapSecond(func(v int) int { return v + 2 }).MapThird(func(v int) int { return v + 3 })
	q := streams.NewQuad(1, 2, 3, 4).ToPair()
	a, b := p.Unpack()
	xs, ys := streams.Unzip(streams.Of(streams.NewPair("a", 1), streams.NewPair("b", 2)))
	fmt.Println(a, b)
	fmt.Println(t)
	fmt.Println(q)
	fmt.Println(xs, ys)
	// Output:
	// 10 3
	// {2 4 6}
	// {1 2}
	// [a b] [1 2]
}

