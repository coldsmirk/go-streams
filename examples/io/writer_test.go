package io_examples

import (
	"bytes"
	"fmt"
	"os"

	streams "github.com/coldsmirk/go-streams"
)

func Example_toWriter() {
	var buf bytes.Buffer
	err := streams.ToWriter(streams.Of(1, 2, 3), &buf, func(n int) string {
		return fmt.Sprintf("%d\n", n)
	})
	fmt.Println(err == nil, len(buf.String()) > 0)
	// Output:
	// true true
}

func Example_toCSV() {
	var buf bytes.Buffer
	err := streams.ToCSV(streams.Of([]string{"a", "1"}, []string{"b", "2"}), &buf)
	fmt.Println(err == nil, len(buf.String()) > 0)
	// Output:
	// true true
}

func Example_csvFile_and_mustFromFileLines() {
	// Create temp file for FromCSVFile test
	tmp, _ := os.CreateTemp("", "gs-csv-*.csv")
	defer func() { _ = os.Remove(tmp.Name()) }()
	if _, err := tmp.WriteString("name,age\nAlice,30\nBob,25\n"); err != nil {
		panic(err)
	}
	if err := tmp.Close(); err != nil {
		panic(err)
	}

	csvStream, err := streams.FromCSVFile(tmp.Name())
	if err == nil {
		defer func() { _ = csvStream.Close() }()
		records := csvStream.Collect()
		// First row is header, second is data
		fmt.Println(len(records), records[0][0], records[1][0])
	}

	// MustFromFileLines panics on error, but works for valid files
	tmp2, _ := os.CreateTemp("", "gs-lines-*.txt")
	defer func() { _ = os.Remove(tmp2.Name()) }()
	if _, err := tmp2.WriteString("line1\nline2\n"); err != nil {
		panic(err)
	}
	if err := tmp2.Close(); err != nil {
		panic(err)
	}

	fls := streams.MustFromFileLines(tmp2.Name())
	defer func() { _ = fls.Close() }()
	lines := fls.Collect()
	fmt.Println(len(lines))
	// Output:
	// 3 name Alice
	// 2
}
