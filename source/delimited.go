package source

import (
	"encoding/csv"
	"errors"
	"io"
	"iter"
	"slices"
)

// Delimited returns a sequence of the records read from r, with fields
// separated by delim. Parsing follows [encoding/csv]: fields may be quoted,
// blank lines are skipped, and every record must hold as many fields as the
// first. An empty r yields nothing.
//
// The sequence ends at the first error, reported as a final pair of a nil
// record and the error. A record with the wrong number of fields is a
// [csv.ParseError] wrapping [csv.ErrFieldCount]; a delim that [encoding/csv]
// does not accept as a field separator, such as a quote or a newline, fails on
// the first read.
func Delimited(r io.Reader, delim rune) iter.Seq2[[]string, error] {
	return delimited(r, delim, false)
}

// delimited is Delimited with control over ReuseRecord. Reuse is only safe when
// the caller consumes each row before asking for the next and never retains it,
// which is true of Records and false of every exported entry point.
func delimited(r io.Reader, delim rune, reuse bool) iter.Seq2[[]string, error] {
	return func(yield func([]string, error) bool) {
		reader := csv.NewReader(r)
		reader.Comma = delim
		reader.ReuseRecord = reuse
		for {
			record, err := reader.Read()
			if errors.Is(err, io.EOF) {
				return
			}
			if err != nil {
				yield(nil, err)
				return
			}
			if !yield(record, nil) {
				return
			}
		}
	}
}

// CSV returns a sequence of the comma-separated records read from r. It is
// [Delimited] with a comma.
func CSV(r io.Reader) iter.Seq2[[]string, error] { return Delimited(r, ',') }

// CSVFile returns a sequence of the comma-separated records of the file at
// path. The file is opened when iteration begins and closed when it ends,
// including when the consumer stops early, so the caller has nothing to close.
// A file that cannot be opened is reported as the single pair of a nil record
// and the error.
func CSVFile(path string) iter.Seq2[[]string, error] { return File(path, CSV) }

// TSV returns a sequence of the tab-separated records read from r. It is
// [Delimited] with a tab.
func TSV(r io.Reader) iter.Seq2[[]string, error] { return Delimited(r, '\t') }

// TSVFile returns a sequence of the tab-separated records of the file at path,
// opened and closed as in [CSVFile].
func TSVFile(path string) iter.Seq2[[]string, error] { return File(path, TSV) }

// Record maps each column name of a delimited file to the value that column
// holds in one row.
type Record map[string]string

// Keyed keys the rows of seq by the names in its first row. The header row is
// not itself yielded, so a seq that is empty, or that holds only a header,
// yields nothing. Where the header repeats a name, the last column with that
// name wins.
//
// The sequence ends at the first error from seq, reported as a final pair of a
// nil Record and the error. Header-keying is independent of the delimiter, so
// it composes with any row source:
//
//	source.Keyed(source.TSVFile(path))
//	source.Keyed(source.File(path, func(r io.Reader) iter.Seq2[[]string, error] {
//		return source.Delimited(r, ';')
//	}))
//
// Keyed copies each row into a Record and retains nothing of it, so a seq that
// reuses its backing slice between rows is safe to pass.
func Keyed(seq iter.Seq2[[]string, error]) iter.Seq2[Record, error] {
	return func(yield func(Record, error) bool) {
		var header []string
		for row, err := range seq {
			if err != nil {
				yield(nil, err)
				return
			}
			if header == nil {
				// Cloned because it outlives the row it came from, which a
				// reusing source is free to overwrite.
				header = slices.Clone(row)
				continue
			}
			record := make(Record, len(header))
			for i, name := range header {
				record[name] = row[i]
			}
			if !yield(record, nil) {
				return
			}
		}
	}
}

// Records returns a sequence of the comma-separated rows read from r, keyed by
// the names in the first row. It is [Keyed] over [CSV]; use Keyed directly for
// any other delimiter.
//
// The header row is not itself yielded, so an r that is empty, or that holds
// only a header, yields nothing. Where the header repeats a name, the last
// column with that name wins. The sequence ends at the first error, reported as
// a final pair of a nil Record and the error. A row whose field count differs
// from the header is a [csv.ParseError] wrapping [csv.ErrFieldCount].
func Records(r io.Reader) iter.Seq2[Record, error] {
	// The reusing reader is safe here because Keyed never retains a row.
	return Keyed(delimited(r, ',', true))
}

// RecordsFile returns a sequence of the header-keyed rows of the file at path,
// parsed as in [Records] and opened and closed as in [CSVFile].
func RecordsFile(path string) iter.Seq2[Record, error] { return File(path, Records) }
