package source_test

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"strings"

	streams "github.com/coldsmirk/go-streams/v2"
	"github.com/coldsmirk/go-streams/v2/source"
)

// A fallible source is a plain iter.Seq2, so it may be ranged over directly.
func ExampleLines() {
	for line, err := range source.Lines(strings.NewReader("alpha\nbeta\n")) {
		if err != nil {
			fmt.Println("read:", err)
			return
		}
		fmt.Println(line)
	}
	// Output:
	// alpha
	// beta
}

// Splitting a string cannot fail, so StringLines has no error slot and chains
// as an ordinary Stream.
func ExampleStringLines() {
	lengths := source.StringLines("a\nbb\nccc\n").
		Map(func(s string) int { return len(s) }).
		Collect()
	fmt.Println(lengths)
	// Output: [1 2 3]
}

// streams.Try collects a fallible source, stopping at the first error.
func ExampleCSV() {
	rows, err := streams.Try(source.CSV(strings.NewReader("name,city\nAda,London\n")))
	fmt.Println(rows, err)

	_, err = streams.Try(source.CSV(strings.NewReader("name,city\nAda,London,UK\n")))
	fmt.Println(errors.Is(err, csv.ErrFieldCount))
	// Output:
	// [[name city] [Ada London]] <nil>
	// true
}

// Records keys each row by the names in the header row, which is not itself
// yielded.
func ExampleRecords() {
	in := "name,city\nAda,London\nKen,NYC\n"
	for record, err := range source.Records(strings.NewReader(in)) {
		if err != nil {
			fmt.Println("parse:", err)
			return
		}
		fmt.Printf("%s lives in %s\n", record["name"], record["city"])
	}
	// Output:
	// Ada lives in London
	// Ken lives in NYC
}

// WriteCSV quotes the fields that need it.
func ExampleWriteCSV() {
	var buf bytes.Buffer
	rows := streams.Of(
		[]string{"name", "city"},
		[]string{"Ada", "London, UK"},
	)
	if err := source.WriteCSV(&buf, rows); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Print(buf.String())
	// Output:
	// name,city
	// Ada,"London, UK"
}
