package source

import (
	"bufio"
	"io"
	"iter"
	"strings"

	streams "github.com/coldsmirk/go-streams/v2"
)

// Lines returns a sequence of the lines read from r. A line ends at a newline,
// which is not included, and a carriage return immediately before it is
// dropped. The last line need not be terminated. An empty r yields nothing.
//
// Reading uses a [bufio.Scanner], so a line longer than
// [bufio.MaxScanTokenSize] fails with [bufio.ErrTooLong]. A read failure ends
// the sequence with a final pair of the empty string and the error.
func Lines(r io.Reader) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			if !yield(scanner.Text(), nil) {
				return
			}
		}
		if err := scanner.Err(); err != nil {
			yield("", err)
		}
	}
}

// LinesFile returns a sequence of the lines of the file at path, split as in
// [Lines]. The file is opened when iteration begins and closed when it ends,
// including when the consumer stops early, so the caller has nothing to close.
// A file that cannot be opened is reported as the single pair of the empty
// string and the error.
func LinesFile(path string) iter.Seq2[string, error] { return File(path, Lines) }

// StringLines returns a Stream of the lines of s, split as in [Lines]. An empty
// string yields nothing. Splitting a string cannot fail, so StringLines has no
// error slot and chains as an ordinary Stream.
func StringLines(s string) streams.Stream[string] {
	return func(yield func(string) bool) {
		for line := range strings.Lines(s) {
			line = strings.TrimSuffix(line, "\n")
			if !yield(strings.TrimSuffix(line, "\r")) {
				return
			}
		}
	}
}

// Bytes returns a Stream over the bytes of b. A nil or empty slice yields
// nothing.
func Bytes(b []byte) streams.Stream[byte] { return streams.Of(b...) }

// Runes returns a Stream over the runes of s. As when ranging over a string,
// each byte of an invalid UTF-8 encoding yields utf8.RuneError. An empty
// string yields nothing.
func Runes(s string) streams.Stream[rune] {
	return func(yield func(rune) bool) {
		for _, r := range s {
			if !yield(r) {
				return
			}
		}
	}
}
