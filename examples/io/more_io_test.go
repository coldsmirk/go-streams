package io_examples

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	streams "github.com/coldsmirk/go-streams"
)

func Example_scanner_and_readerErr() {
	// FromScanner
	sc := bufio.NewScanner(strings.NewReader("a\nb"))
	fmt.Println(streams.FromScanner(sc).Collect())
	// FromReaderLinesErr (no error case)
	ok := streams.FromReaderLinesErr(strings.NewReader("x\ny")).Collect()
	fmt.Println(len(ok))
	// FromBytes / FromRunes
	// Convert []byte to []int to avoid staticcheck complaint about printing []byte directly
	b := streams.FromBytes([]byte{1, 2}).Collect()
	ints := make([]int, len(b))
	for i, v := range b {
		ints[i] = int(v)
	}
	fmt.Println(ints)
	fmt.Println(streams.FromRunes("go").Collect())
	// Output:
	// [a b]
	// 2
	// [1 2]
	// [103 111]
}

func Example_files_and_csv_tsv() {
	// temp text file
	tmp, _ := os.CreateTemp("", "gs-lines-*.txt")
	defer func() { _ = os.Remove(tmp.Name()) }()
	if _, err := tmp.WriteString("L1\nL2\n"); err != nil {
		panic(err)
	}
	if err := tmp.Close(); err != nil {
		panic(err)
	}
	fl, _ := streams.FromFileLines(tmp.Name())
	defer func() { _ = fl.Close() }()
	fmt.Println(len(fl.Collect()))

	// ToFile
	tmp2, _ := os.CreateTemp("", "gs-out-*.txt")
	if err := tmp2.Close(); err != nil {
		panic(err)
	}
	defer func() { _ = os.Remove(tmp2.Name()) }()
	_ = streams.ToFile(streams.Of(1, 2), tmp2.Name(), func(v int) string { return fmt.Sprintf("%d", v) })

	// CSV (no header) / TSV
	csv := streams.FromCSV(strings.NewReader("a,b\nc,d\n")).Collect()
	tsv := streams.FromTSV(strings.NewReader("a\tb\nc\td\n")).Collect()
	fmt.Println(len(csv), len(tsv))

	// CSV with header, plus ToCSVFile
	tmp3, _ := os.CreateTemp("", "gs-csv-*.csv")
	defer func() { _ = os.Remove(tmp3.Name()) }()
	if err := tmp3.Close(); err != nil {
		panic(err)
	}
	_ = streams.ToCSVFile(streams.Of([]string{"h1", "h2"}, []string{"v1", "v2"}), tmp3.Name())
	rec := streams.FromCSVWithHeader(strings.NewReader("k,v\nA,1\n")).Collect()
	fmt.Println(rec[0].Get("v"))
	// Output:
	// 2
	// 2 2
	// 1
}
