package source

import (
	"iter"
	"os"
	"strings"
	"testing"

	streams "github.com/coldsmirk/go-streams/v2"
)

// The iter contract says yield panics if it is called after returning false.
// Every source that forwards elements must therefore honour a false result and
// stop. These tests break out of each source after one element, which panics if
// the early-return path is missing.

func breakAfterOne[T any](t *testing.T, name string, s streams.Stream[T]) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("%s: yielding after the consumer stopped: %v", name, r)
		}
	}()
	n := 0
	for range iter.Seq[T](s) {
		n++
		break
	}
	if n != 1 {
		t.Errorf("%s: consumed %d elements before the break, want 1", name, n)
	}
}

func breakAfterOne2[T any](t *testing.T, name string, seq iter.Seq2[T, error]) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("%s: yielding after the consumer stopped: %v", name, r)
		}
	}()
	n := 0
	for range seq {
		n++
		break
	}
	if n != 1 {
		t.Errorf("%s: consumed %d pairs before the break, want 1", name, n)
	}
}

func TestTextSourcesHonourEarlyStop(t *testing.T) {
	const text = "alpha\nbeta\ngamma\n"
	path := writeTemp(t, "lines.txt", text)

	breakAfterOne2(t, "Lines", Lines(strings.NewReader(text)))
	breakAfterOne2(t, "LinesFile", LinesFile(path))
	breakAfterOne(t, "StringLines", StringLines(text))
	breakAfterOne(t, "Bytes", Bytes([]byte("abc")))
	breakAfterOne(t, "Runes", Runes("abc"))
}

func TestDelimitedSourcesHonourEarlyStop(t *testing.T) {
	const comma = "a,b\n1,2\n3,4\n"
	const tab = "a\tb\n1\t2\n3\t4\n"
	csvPath := writeTemp(t, "data.csv", comma)
	tsvPath := writeTemp(t, "data.tsv", tab)

	breakAfterOne2(t, "Delimited", Delimited(strings.NewReader("a;b\n1;2\n"), ';'))
	breakAfterOne2(t, "CSV", CSV(strings.NewReader(comma)))
	breakAfterOne2(t, "CSVFile", CSVFile(csvPath))
	breakAfterOne2(t, "TSV", TSV(strings.NewReader(tab)))
	breakAfterOne2(t, "TSVFile", TSVFile(tsvPath))
	breakAfterOne2(t, "Records", Records(strings.NewReader(comma)))
	breakAfterOne2(t, "RecordsFile", RecordsFile(csvPath))
}

// A file source must close its file however the iteration ends, so a caller
// never has to. Leaking one descriptor per pass would show as a rising count.
func TestFileSourcesCloseWhateverEndsTheIteration(t *testing.T) {
	path := writeTemp(t, "data.csv", "a,b\n1,2\n3,4\n")
	pass := func() {
		for range LinesFile(path) {
			break
		}
		for range CSVFile(path) {
			break
		}
		for range TSVFile(path) {
			break
		}
		for range RecordsFile(path) {
			break
		}
		for range LinesFile(path) { // drained to the end rather than stopped early
		}
	}

	pass() // warm up, so any descriptor opened once is already open
	before := openDescriptors(t)
	for range 10 {
		pass()
	}
	if after := openDescriptors(t); after != before {
		t.Errorf("open descriptors went from %d to %d over 10 passes, want no change", before, after)
	}
}

// openDescriptors returns the number of descriptors this process holds open.
func openDescriptors(t *testing.T) int {
	t.Helper()
	entries, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		t.Skipf("counting open descriptors is unsupported here: %v", err)
	}
	return len(entries)
}
