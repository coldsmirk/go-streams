package parallel_examples

import (
	"fmt"

	streams "github.com/coldsmirk/go-streams"
)

func Example_prefetch() {
	out := streams.Prefetch(streams.Range(1, 5), 2).Collect()
	fmt.Println(out)
	// Output:
	// [1 2 3 4]
}

