package source_test

import (
	"fmt"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"

	streams "github.com/coldsmirk/go-streams/v2"
	"github.com/coldsmirk/go-streams/v2/source"
)

// A report end to end: read a delimited source by column name, aggregate it,
// and write the result back out. Ok keeps the read lazy and defers the error
// until the pipeline has drained, so the aggregation stays a single expression.
func Example_report() {
	const orders = `region,units
west,10
east,7
west,5
north,3
`

	rows, readErr := streams.Ok(source.Records(strings.NewReader(orders)))

	totals := rows.Fold(map[string]int{}, func(acc map[string]int, r source.Record) map[string]int {
		n, _ := strconv.Atoi(r["units"])
		acc[r["region"]] += n
		return acc
	})
	if err := readErr(); err != nil {
		fmt.Println("read:", err)
		return
	}

	// Sort the regions so the report is stable; map order is not.
	report := streams.Of(slices.Sorted(maps.Keys(totals))...).
		Map(func(region string) []string {
			return []string{region, strconv.Itoa(totals[region])}
		})

	header := streams.Of([]string{"region", "total"})
	if err := source.WriteCSV(os.Stdout, streams.Concat(header, report)); err != nil {
		fmt.Println("write:", err)
	}
	// Output:
	// region,total
	// east,7
	// north,3
	// west,15
}
