package streams_examples

import (
	"fmt"

	streams "github.com/coldsmirk/go-streams"
)

func Example_seq_forRange() {
	sum := 0
	for v := range streams.Of(1, 2, 3).Seq() {
		sum += v
	}
	fmt.Println(sum)
	// Output:
	// 6
}

