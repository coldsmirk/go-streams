package source

import (
	"encoding/csv"
	"fmt"
	"io"
	"iter"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/iotest"

	streams "github.com/coldsmirk/go-streams/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeTemp writes content to a new file in a temporary directory and returns
// its path.
func writeTemp(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
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
			require.NoErrorf(t, err, "CSV(%q)", tt.in)
			assert.Equalf(t, tt.want, got, "CSV(%q)", tt.in)
		})
	}
}

func TestTSV(t *testing.T) {
	got, err := collect2(TSV(strings.NewReader("a\tb\n1\t2\n")))
	require.NoError(t, err, "TSV error")
	assert.Equal(t, [][]string{{"a", "b"}, {"1", "2"}}, got, "TSV")
}

func TestDelimited(t *testing.T) {
	got, err := collect2(Delimited(strings.NewReader("a;b\n1;2\n"), ';'))
	require.NoError(t, err, "Delimited error")
	assert.Equal(t, [][]string{{"a", "b"}, {"1", "2"}}, got, "Delimited")
}

func TestDelimitedInvalidDelimiter(t *testing.T) {
	got, err := collect2(Delimited(strings.NewReader("a,b\n"), '"'))
	require.Errorf(t, err, "Delimited with a quote delimiter = %q, want an error", got)
	assert.Empty(t, got, "Delimited yielded records before the error, want none")
}

func TestCSVMalformedRow(t *testing.T) {
	got, err := collect2(CSV(strings.NewReader("a,b\n1,2,3\n4,5\n")))
	require.ErrorIs(t, err, csv.ErrFieldCount, "CSV error")
	assert.Equal(t, [][]string{{"a", "b"}}, got, "CSV before the error")
}

func TestCSVReadError(t *testing.T) {
	got, err := collect2(CSV(iotest.ErrReader(errRead)))
	require.ErrorIs(t, err, errRead, "CSV error")
	assert.Empty(t, got, "CSV yielded records before the error, want none")
}

func TestCSVFile(t *testing.T) {
	got, err := collect2(CSVFile(writeTemp(t, "data.csv", "a,b\n1,2\n")))
	require.NoError(t, err, "CSVFile error")
	assert.Equal(t, [][]string{{"a", "b"}, {"1", "2"}}, got, "CSVFile")
}

func TestTSVFile(t *testing.T) {
	got, err := collect2(TSVFile(writeTemp(t, "data.tsv", "a\tb\n1\t2\n")))
	require.NoError(t, err, "TSVFile error")
	assert.Equal(t, [][]string{{"a", "b"}, {"1", "2"}}, got, "TSVFile")
}

func TestCSVFileMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "absent.csv")

	pairs := 0
	var last error
	for record, err := range CSVFile(path) {
		pairs++
		last = err
		assert.Nil(t, record, "CSVFile on a missing file yielded a record, want nil")
	}
	require.Equal(t, 1, pairs, "CSVFile on a missing file yielded the wrong number of pairs, want 1")
	assert.ErrorIs(t, last, os.ErrNotExist, "CSVFile error")
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
			require.NoErrorf(t, err, "Records(%q)", tt.in)
			assert.Equalf(t, tt.want, got, "Records(%q)", tt.in)
		})
	}
}

func TestRecordsMalformedRow(t *testing.T) {
	got, err := collect2(Records(strings.NewReader("a,b\n1,2\n3,4,5\n")))
	require.ErrorIs(t, err, csv.ErrFieldCount, "Records error")
	assert.Equal(t, []Record{{"a": "1", "b": "2"}}, got, "Records before the error")
}

func TestRecordsReadError(t *testing.T) {
	got, err := collect2(Records(iotest.ErrReader(errRead)))
	require.ErrorIs(t, err, errRead, "Records error")
	assert.Empty(t, got, "Records yielded records before the error, want none")
}

func TestRecordsFile(t *testing.T) {
	got, err := collect2(RecordsFile(writeTemp(t, "data.csv", "name,city\nAda,London\n")))
	require.NoError(t, err, "RecordsFile error")
	assert.Equal(t, []Record{{"name": "Ada", "city": "London"}}, got, "RecordsFile")
}

func TestRecordsFileMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "absent.csv")

	pairs := 0
	var last error
	for record, err := range RecordsFile(path) {
		pairs++
		last = err
		assert.Nil(t, record, "RecordsFile on a missing file yielded a record, want nil")
	}
	require.Equal(t, 1, pairs, "RecordsFile on a missing file yielded the wrong number of pairs, want 1")
	assert.ErrorIs(t, last, os.ErrNotExist, "RecordsFile error")
}

// Header-keying used to be reachable only from comma-separated input, because
// Records built its own CSV reader. Keyed makes it a transformation over any
// row source, so every delimiter — including one this package does not name —
// composes with it.
func TestKeyedComposesWithAnyDelimiter(t *testing.T) {
	t.Run("TSV with a header", func(t *testing.T) {
		in := "name\tage\nada\t36\nlinus\t54\n"
		got, err := streams.Try(Keyed(TSV(strings.NewReader(in))))
		require.NoError(t, err)
		require.Lenf(t, got, 2, "records = %v", got)
		assert.Equalf(t, "ada", got[0]["name"], "records = %v", got)
		assert.Equalf(t, "54", got[1]["age"], "records = %v", got)
	})

	t.Run("semicolons with a header", func(t *testing.T) {
		in := "name;age\nada;36\n"
		got, err := streams.Try(Keyed(Delimited(strings.NewReader(in), ';')))
		require.NoError(t, err)
		require.Lenf(t, got, 1, "records = %v", got)
		assert.Equalf(t, "36", got[0]["age"], "records = %v", got)
	})

	t.Run("Records is Keyed over CSV", func(t *testing.T) {
		in := "name,age\nada,36\n"
		viaRecords, err1 := streams.Try(Records(strings.NewReader(in)))
		viaKeyed, err2 := streams.Try(Keyed(CSV(strings.NewReader(in))))
		require.NoError(t, err1)
		require.NoError(t, err2)
		require.Lenf(t, viaKeyed, len(viaRecords), "Records = %v, Keyed(CSV) = %v", viaRecords, viaKeyed)
		assert.Equalf(t, viaRecords[0]["name"], viaKeyed[0]["name"], "Records = %v, Keyed(CSV) = %v", viaRecords, viaKeyed)
	})

	t.Run("header only, and empty", func(t *testing.T) {
		for _, in := range []string{"name,age\n", ""} {
			got, err := streams.Try(Keyed(CSV(strings.NewReader(in))))
			assert.NoErrorf(t, err, "Keyed(%q)", in)
			assert.Emptyf(t, got, "Keyed(%q) yielded records, want none", in)
		}
	})
}

// Keyed accepts any row source, so it must enforce the field-count contract
// itself; encoding/csv enforces it only for the sources built on Delimited.
func TestKeyedRejectsRaggedRows(t *testing.T) {
	rows := func(rows ...[]string) iter.Seq2[[]string, error] {
		return func(yield func([]string, error) bool) {
			for _, r := range rows {
				if !yield(r, nil) {
					return
				}
			}
		}
	}

	t.Run("short row", func(t *testing.T) {
		got, err := collect2(Keyed(rows([]string{"a", "b"}, []string{"1"})))
		require.ErrorIs(t, err, ErrFieldCount, "Keyed error")
		assert.Empty(t, got, "Keyed yielded records before the error, want none")
	})

	t.Run("long row", func(t *testing.T) {
		got, err := collect2(Keyed(rows([]string{"a", "b"}, []string{"1", "2", "3"})))
		require.ErrorIs(t, err, ErrFieldCount, "Keyed error")
		assert.Empty(t, got, "Keyed yielded records before the error, want none")
	})

	t.Run("conforming rows precede the error", func(t *testing.T) {
		got, err := collect2(Keyed(rows([]string{"a"}, []string{"1"}, []string{"2", "x"})))
		require.ErrorIs(t, err, ErrFieldCount, "Keyed error")
		assert.Equal(t, []Record{{"a": "1"}}, got, "Keyed before the error")
	})
}

// File is the general form the *File twins are built on. Exporting it is what
// lets a caller read a file in a format this package does not name, without
// re-implementing the open/close/error-ordering contract.
func TestFileReadsAnyFormat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.psv")
	require.NoError(t, os.WriteFile(path, []byte("name|age\nada|36\nlinus|54\n"), 0o644))

	rows, err := streams.Try(File(path, func(r io.Reader) iter.Seq2[[]string, error] {
		return Delimited(r, '|')
	}))
	require.NoError(t, err)
	require.Lenf(t, rows, 3, "rows = %v", rows)
	assert.Equalf(t, "ada", rows[1][0], "rows = %v", rows)

	// Composing both new pieces: a pipe-separated file with a header.
	recs, err := streams.Try(Keyed(File(path, func(r io.Reader) iter.Seq2[[]string, error] {
		return Delimited(r, '|')
	})))
	require.NoError(t, err)
	require.Lenf(t, recs, 2, "records = %v", recs)
	assert.Equalf(t, "ada", recs[0]["name"], "records = %v", recs)
	assert.Equalf(t, "54", recs[1]["age"], "records = %v", recs)

	// A missing file still yields exactly one pair, as every *File twin does.
	n := 0
	for _, err := range File(filepath.Join(dir, "nope"), Lines) {
		n++
		assert.Error(t, err, "expected an error pair")
	}
	assert.Equal(t, 1, n, "missing file yielded the wrong number of pairs, want 1")
}

// The lazy bridge end to end: a file feeds a pipeline without being buffered.
func TestOkStreamsAFileWithoutBuffering(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "big.txt")
	var sb strings.Builder
	for i := range 100_000 {
		fmt.Fprintf(&sb, "line-%d\n", i)
	}
	require.NoError(t, os.WriteFile(path, []byte(sb.String()), 0o644))

	lines, readErr := streams.Ok(LinesFile(path))
	got := lines.Filter(func(s string) bool { return strings.HasSuffix(s, "7") }).Take(3).Collect()
	require.NoError(t, readErr())
	require.Lenf(t, got, 3, "lines = %v", got)
	assert.Equalf(t, "line-7", got[0], "lines = %v", got)
}
