package numeric_examples

import (
	"fmt"

	streams "github.com/coldsmirk/go-streams"
)

func Example_numericRunning() {
	nums := streams.Of(1, 2, 3, 5)
	fmt.Println(streams.RunningSum(nums).Collect())
	fmt.Println(streams.Differences(streams.Of(1, 4, 9)).Collect())
	// Output:
	// [1 3 6 11]
	// [3 5]
}

