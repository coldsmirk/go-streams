// Package source reads sequences from readers and files, and writes them back
// out.
//
// A source that can fail returns the standard iter.Seq2[T, error] rather than a
// [streams.Stream], matching the way the streams package handles fallible work.
// The sequence ends at the first error, so a consumer may range over it
// directly or hand it to [streams.Try]:
//
//	for line, err := range source.LinesFile("access.log") { ... }
//
//	rows, err := streams.Try(source.CSVFile("data.csv"))
//
// A source that cannot fail, such as [StringLines], returns a [streams.Stream]
// and chains without an error slot.
//
// A file source opens the file when iteration begins and closes it when
// iteration ends, including when the consumer stops early, so the caller has
// nothing to close. A file that cannot be opened is reported as a single pair
// of the zero value and the error.
package source
