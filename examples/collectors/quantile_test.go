package collectors_examples

import (
	"fmt"

	streams "github.com/coldsmirk/go-streams"
)

func Example_quantile() {
	q := streams.Quantile(streams.Of(1, 2, 3, 4, 5), 0.5, func(a, b int) bool { return a < b })
	fmt.Println(q.IsPresent(), q.Get())
	// Output:
	// true 3
}

