package tuples_examples

import (
	"fmt"

	streams "github.com/coldsmirk/go-streams"
)

func Example_tuplesAdvanced() {
	p := streams.NewPair(1, "x").Swap()
	t := streams.NewTriple(1, "y", true).ToPair()
	q := streams.NewQuad(1, 2, 3, 4).ToTriple()
	fmt.Println(p)
	fmt.Println(t)
	fmt.Println(q)
	// Output:
	// {x 1}
	// {1 y}
	// {1 2 3}
}

