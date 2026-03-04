package context_examples

import (
	"context"
	"fmt"

	streams "github.com/coldsmirk/go-streams"
)

func Example_collectCtx() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	counter := 0
	src := streams.Generate(func() int {
		counter++
		if counter == 3 {
			cancel()
		}
		return counter
	}).Limit(10)
	vals, err := streams.CollectCtx(ctx, src)
	fmt.Println(len(vals) <= 4, err != nil)
	// Output:
	// true true
}

