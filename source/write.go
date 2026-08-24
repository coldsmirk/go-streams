package source

import (
	"bufio"
	"encoding/csv"
	"io"

	streams "github.com/coldsmirk/go-streams/v2"
)

// WriteLines writes each element of s to w on a line of its own, formatted by
// format. A newline is appended after every element, so format should not end
// with one. WriteLines returns the first write error and stops, leaving the
// rest of s unconsumed. An empty Stream writes nothing and returns nil.
func WriteLines[T any](w io.Writer, s streams.Stream[T], format func(T) string) error {
	bw := bufio.NewWriter(w)
	for v := range s {
		if _, err := bw.WriteString(format(v)); err != nil {
			return err
		}
		if err := bw.WriteByte('\n'); err != nil {
			return err
		}
	}
	return bw.Flush()
}

// WriteFile writes s to the file at path as in [WriteLines], creating the file
// or truncating an existing one. The file is closed before WriteFile returns;
// if writing succeeded but closing failed, the close error is returned.
func WriteFile[T any](path string, s streams.Stream[T], format func(T) string) error {
	return writeFile(path, func(w io.Writer) error { return WriteLines(w, s, format) })
}

// WriteCSV writes each record of s to w as one comma-separated row, quoting the
// fields that [encoding/csv] requires to be quoted. WriteCSV returns the first
// write error and stops, leaving the rest of s unconsumed. An empty Stream
// writes nothing and returns nil.
func WriteCSV(w io.Writer, s streams.Stream[[]string]) error {
	writer := csv.NewWriter(w)
	for record := range s {
		if err := writer.Write(record); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

// WriteCSVFile writes s to the file at path as in [WriteCSV], creating the file
// or truncating an existing one. The file is closed before WriteCSVFile
// returns; if writing succeeded but closing failed, the close error is
// returned.
func WriteCSVFile(path string, s streams.Stream[[]string]) error {
	return writeFile(path, func(w io.Writer) error { return WriteCSV(w, s) })
}
