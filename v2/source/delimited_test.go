package source

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"iter"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"testing/iotest"

	streams "github.com/coldsmirk/go-streams/v2"
)

func equalRecords(a, b [][]string) bool {
	return slices.EqualFunc(a, b, slices.Equal[[]string, string])
}

// writeTemp writes content to a new file in a temporary directory and returns
// its path.
func writeTemp(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestCSV(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want [][]string
	}{
		{"empty", "", nil},
		{"rows", "a,b\n1,2\n", [][]string{{"a", "b"}, {"1", "2"}}},
		{"unterminated", "a,b\n1,2", [][]string{{"a", "b"}, {"1", "2"}}},
		{"quoted", "\"a,x\",b\n1,2\n", [][]string{{"a,x", "b"}, {"1", "2"}}},
		{"blank lines skipped", "a,b\n\n1,2\n", [][]string{{"a", "b"}, {"1", "2"}}},
		{"empty fields", ",\n", [][]string{{"", ""}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := collect2(CSV(strings.NewReader(tt.in)))
			if err != nil {
				t.Fatalf("CSV(%q) error = %v, want nil", tt.in, err)
			}
			if !equalRecords(got, tt.want) {
				t.Errorf("CSV(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestTSV(t *testing.T) {
	got, err := collect2(TSV(strings.NewReader("a\tb\n1\t2\n")))
	if err != nil {
		t.Fatalf("TSV error = %v, want nil", err)
	}
	if want := [][]string{{"a", "b"}, {"1", "2"}}; !equalRecords(got, want) {
		t.Errorf("TSV = %q, want %q", got, want)
	}
}

func TestDelimited(t *testing.T) {
	got, err := collect2(Delimited(strings.NewReader("a;b\n1;2\n"), ';'))
	if err != nil {
		t.Fatalf("Delimited error = %v, want nil", err)
	}
	if want := [][]string{{"a", "b"}, {"1", "2"}}; !equalRecords(got, want) {
		t.Errorf("Delimited = %q, want %q", got, want)
	}
}

func TestDelimitedInvalidDelimiter(t *testing.T) {
	got, err := collect2(Delimited(strings.NewReader("a,b\n"), '"'))
	if err == nil {
		t.Fatalf("Delimited with a quote delimiter = %q, want an error", got)
	}
	if len(got) != 0 {
		t.Errorf("Delimited yielded %q before the error, want none", got)
	}
}

func TestCSVMalformedRow(t *testing.T) {
	got, err := collect2(CSV(strings.NewReader("a,b\n1,2,3\n4,5\n")))
	if !errors.Is(err, csv.ErrFieldCount) {
		t.Fatalf("CSV error = %v, want one matching csv.ErrFieldCount", err)
	}
	if want := [][]string{{"a", "b"}}; !equalRecords(got, want) {
		t.Errorf("CSV = %q before the error, want %q", got, want)
	}
}

func TestCSVReadError(t *testing.T) {
	got, err := collect2(CSV(iotest.ErrReader(errRead)))
	if !errors.Is(err, errRead) {
		t.Fatalf("CSV error = %v, want %v", err, errRead)
	}
	if len(got) != 0 {
		t.Errorf("CSV yielded %q before the error, want none", got)
	}
}

func TestCSVFile(t *testing.T) {
	got, err := collect2(CSVFile(writeTemp(t, "data.csv", "a,b\n1,2\n")))
	if err != nil {
		t.Fatalf("CSVFile error = %v, want nil", err)
	}
	if want := [][]string{{"a", "b"}, {"1", "2"}}; !equalRecords(got, want) {
		t.Errorf("CSVFile = %q, want %q", got, want)
	}
}

func TestTSVFile(t *testing.T) {
	got, err := collect2(TSVFile(writeTemp(t, "data.tsv", "a\tb\n1\t2\n")))
	if err != nil {
		t.Fatalf("TSVFile error = %v, want nil", err)
	}
	if want := [][]string{{"a", "b"}, {"1", "2"}}; !equalRecords(got, want) {
		t.Errorf("TSVFile = %q, want %q", got, want)
	}
}

func TestCSVFileMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "absent.csv")

	pairs := 0
	var last error
	for record, err := range CSVFile(path) {
		pairs++
		last = err
		if record != nil {
			t.Errorf("CSVFile on a missing file yielded record %q, want nil", record)
		}
	}
	if pairs != 1 {
		t.Fatalf("CSVFile on a missing file yielded %d pairs, want 1", pairs)
	}
	if !errors.Is(last, os.ErrNotExist) {
		t.Errorf("CSVFile error = %v, want one matching os.ErrNotExist", last)
	}
}

func TestRecords(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []Record
	}{
		{"empty", "", nil},
		{"header only", "name,city\n", nil},
		{"rows", "name,city\nAda,London\nKen,NYC\n", []Record{
			{"name": "Ada", "city": "London"},
			{"name": "Ken", "city": "NYC"},
		}},
		{"duplicate header name", "k,k\n1,2\n", []Record{{"k": "2"}}},
		{"empty values", "a,b\n,\n", []Record{{"a": "", "b": ""}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := collect2(Records(strings.NewReader(tt.in)))
			if err != nil {
				t.Fatalf("Records(%q) error = %v, want nil", tt.in, err)
			}
			if !slices.EqualFunc(got, tt.want, maps.Equal[Record, Record]) {
				t.Errorf("Records(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestRecordsMalformedRow(t *testing.T) {
	got, err := collect2(Records(strings.NewReader("a,b\n1,2\n3,4,5\n")))
	if !errors.Is(err, csv.ErrFieldCount) {
		t.Fatalf("Records error = %v, want one matching csv.ErrFieldCount", err)
	}
	want := []Record{{"a": "1", "b": "2"}}
	if !slices.EqualFunc(got, want, maps.Equal[Record, Record]) {
		t.Errorf("Records = %v before the error, want %v", got, want)
	}
}

func TestRecordsReadError(t *testing.T) {
	got, err := collect2(Records(iotest.ErrReader(errRead)))
	if !errors.Is(err, errRead) {
		t.Fatalf("Records error = %v, want %v", err, errRead)
	}
	if len(got) != 0 {
		t.Errorf("Records yielded %v before the error, want none", got)
	}
}

func TestRecordsFile(t *testing.T) {
	got, err := collect2(RecordsFile(writeTemp(t, "data.csv", "name,city\nAda,London\n")))
	if err != nil {
		t.Fatalf("RecordsFile error = %v, want nil", err)
	}
	want := []Record{{"name": "Ada", "city": "London"}}
	if !slices.EqualFunc(got, want, maps.Equal[Record, Record]) {
		t.Errorf("RecordsFile = %v, want %v", got, want)
	}
}

func TestRecordsFileMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "absent.csv")

	pairs := 0
	var last error
	for record, err := range RecordsFile(path) {
		pairs++
		last = err
		if record != nil {
			t.Errorf("RecordsFile on a missing file yielded record %v, want nil", record)
		}
	}
	if pairs != 1 {
		t.Fatalf("RecordsFile on a missing file yielded %d pairs, want 1", pairs)
	}
	if !errors.Is(last, os.ErrNotExist) {
		t.Errorf("RecordsFile error = %v, want one matching os.ErrNotExist", last)
	}
}

// Header-keying used to be reachable only from comma-separated input, because
// Records built its own CSV reader. Keyed makes it a transformation over any
// row source, so every delimiter — including one this package does not name —
// composes with it.
func TestKeyedComposesWithAnyDelimiter(t *testing.T) {
	t.Run("TSV with a header", func(t *testing.T) {
		in := "name\tage\nada\t36\nlinus\t54\n"
		got, err := streams.Try(Keyed(TSV(strings.NewReader(in))))
		if err != nil {
			t.Fatalf("err = %v", err)
		}
		if len(got) != 2 || got[0]["name"] != "ada" || got[1]["age"] != "54" {
			t.Errorf("records = %v", got)
		}
	})

	t.Run("semicolons with a header", func(t *testing.T) {
		in := "name;age\nada;36\n"
		got, err := streams.Try(Keyed(Delimited(strings.NewReader(in), ';')))
		if err != nil {
			t.Fatalf("err = %v", err)
		}
		if len(got) != 1 || got[0]["age"] != "36" {
			t.Errorf("records = %v", got)
		}
	})

	t.Run("Records is Keyed over CSV", func(t *testing.T) {
		in := "name,age\nada,36\n"
		viaRecords, err1 := streams.Try(Records(strings.NewReader(in)))
		viaKeyed, err2 := streams.Try(Keyed(CSV(strings.NewReader(in))))
		if err1 != nil || err2 != nil {
			t.Fatalf("errs = %v %v", err1, err2)
		}
		if len(viaRecords) != len(viaKeyed) || viaRecords[0]["name"] != viaKeyed[0]["name"] {
			t.Errorf("Records = %v, Keyed(CSV) = %v", viaRecords, viaKeyed)
		}
	})

	t.Run("header only, and empty", func(t *testing.T) {
		for _, in := range []string{"name,age\n", ""} {
			got, err := streams.Try(Keyed(CSV(strings.NewReader(in))))
			if err != nil || len(got) != 0 {
				t.Errorf("Keyed(%q) = %v, %v", in, got, err)
			}
		}
	})
}

// File is the general form the *File twins are built on. Exporting it is what
// lets a caller read a file in a format this package does not name, without
// re-implementing the open/close/error-ordering contract.
func TestFileReadsAnyFormat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.psv")
	if err := os.WriteFile(path, []byte("name|age\nada|36\nlinus|54\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	rows, err := streams.Try(File(path, func(r io.Reader) iter.Seq2[[]string, error] {
		return Delimited(r, '|')
	}))
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if len(rows) != 3 || rows[1][0] != "ada" {
		t.Errorf("rows = %v", rows)
	}

	// Composing both new pieces: a pipe-separated file with a header.
	recs, err := streams.Try(Keyed(File(path, func(r io.Reader) iter.Seq2[[]string, error] {
		return Delimited(r, '|')
	})))
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if len(recs) != 2 || recs[0]["name"] != "ada" || recs[1]["age"] != "54" {
		t.Errorf("records = %v", recs)
	}

	// A missing file still yields exactly one pair, as every *File twin does.
	n := 0
	for _, err := range File(filepath.Join(dir, "nope"), Lines) {
		n++
		if err == nil {
			t.Error("expected an error pair")
		}
	}
	if n != 1 {
		t.Errorf("missing file yielded %d pairs, want 1", n)
	}
}

// The lazy bridge end to end: a file feeds a pipeline without being buffered.
func TestOkStreamsAFileWithoutBuffering(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "big.txt")
	var sb strings.Builder
	for i := range 100_000 {
		fmt.Fprintf(&sb, "line-%d\n", i)
	}
	if err := os.WriteFile(path, []byte(sb.String()), 0o644); err != nil {
		t.Fatal(err)
	}

	lines, readErr := streams.Ok(LinesFile(path))
	got := lines.Filter(func(s string) bool { return strings.HasSuffix(s, "7") }).Take(3).Collect()
	if err := readErr(); err != nil {
		t.Fatalf("err = %v", err)
	}
	if len(got) != 3 || got[0] != "line-7" {
		t.Errorf("lines = %v", got)
	}
}
