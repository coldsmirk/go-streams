package result_examples

import (
	"fmt"
	"strconv"

	streams "github.com/coldsmirk/go-streams"
)

// Example: MapErrTo and CollectResults.
func Example_resultPipeline() {
	parsed := streams.MapErrTo(streams.Of("1", "x", "2"), strconv.Atoi)
	vals, err := streams.CollectResults(parsed)
	fmt.Println(vals, err != nil)
	// Output:
	// [1] true
}

