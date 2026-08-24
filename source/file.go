package source

import (
	"io"
	"iter"
	"os"
)

// readFile returns the sequence seq produces over the contents of the file at
// path. The file is opened when iteration begins and closed when it ends,
// including when the consumer stops early. A file that cannot be opened is
// reported as the single pair of the zero value and the error.
// File opens the file at path, hands it to parse, and yields what parse yields,
// closing the file when the iteration ends however it ends. A file that cannot
// be opened yields exactly one pair, of the zero value and the error.
//
// It is the general form of [LinesFile], [CSVFile], [TSVFile] and
// [RecordsFile], and the way to read a file in any other format:
//
//	for row, err := range source.File(path, func(r io.Reader) iter.Seq2[[]string, error] {
//		return source.Delimited(r, ';')
//	}) { ... }
func File[T any](path string, parse func(io.Reader) iter.Seq2[T, error]) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		f, err := os.Open(path)
		if err != nil {
			var zero T
			yield(zero, err)
			return
		}
		defer func() { _ = f.Close() }()
		for v, err := range parse(f) {
			if !yield(v, err) {
				return
			}
		}
	}
}

// writeFile creates the file at path, truncating an existing one, and passes it
// to write. The file is closed before writeFile returns; if write succeeded but
// closing failed, the close error is returned.
func writeFile(path string, write func(io.Writer) error) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := write(f); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}
