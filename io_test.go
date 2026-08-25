package streams

import (
	"bufio"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromReaderLines(t *testing.T) {
	t.Parallel()
	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		reader := strings.NewReader("line1\nline2\nline3")
		lines := FromReaderLines(reader).Collect()
		assert.Equal(t, []string{"line1", "line2", "line3"}, lines, "FromReaderLines should split lines")
	})

	t.Run("Empty", func(t *testing.T) {
		t.Parallel()
		reader := strings.NewReader("")
		lines := FromReaderLines(reader).Collect()
		assert.Len(t, lines, 0, "FromReaderLines empty should return zero lines")
	})

	t.Run("SingleLine", func(t *testing.T) {
		t.Parallel()
		reader := strings.NewReader("single")
		lines := FromReaderLines(reader).Collect()
		assert.Equal(t, []string{"single"}, lines, "FromReaderLines single should return one line")
	})
}

func TestFromScanner(t *testing.T) {
	t.Parallel()
	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		scanner := bufio.NewScanner(strings.NewReader("a\nb\nc"))
		lines := FromScanner(scanner).Collect()
		assert.Equal(t, []string{"a", "b", "c"}, lines, "FromScanner should scan all lines")
	})

	t.Run("EarlyTermination", func(t *testing.T) {
		t.Parallel()
		scanner := bufio.NewScanner(strings.NewReader("1\n2\n3\n4\n5"))
		lines := FromScanner(scanner).Limit(2).Collect()
		assert.Equal(t, []string{"1", "2"}, lines, "FromScanner should respect Limit")
	})
}

func TestFromScannerErr(t *testing.T) {
	t.Parallel()
	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		scanner := bufio.NewScanner(strings.NewReader("a\nb"))
		results := FromScannerErr(scanner).Collect()
		assert.Len(t, results, 2, "FromScannerErr should return one result per line")
		assert.True(t, results[0].IsOk(), "FromScannerErr should return Ok for valid lines")
		assert.Equal(t, "a", results[0].Value(), "FromScannerErr should preserve line content")
	})
}

func TestFromStringLines(t *testing.T) {
	t.Parallel()
	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		lines := FromStringLines("line1\nline2\nline3").Collect()
		assert.Equal(t, []string{"line1", "line2", "line3"}, lines, "FromStringLines should split lines")
	})

	t.Run("Empty", func(t *testing.T) {
		t.Parallel()
		lines := FromStringLines("").Collect()
		assert.Len(t, lines, 0, "FromStringLines on empty should return zero lines")
	})

	t.Run("WithEmptyLines", func(t *testing.T) {
		t.Parallel()
		lines := FromStringLines("a\n\nb").Collect()
		assert.Equal(t, []string{"a", "", "b"}, lines, "FromStringLines should preserve empty lines")
	})
}

func TestFromRunes(t *testing.T) {
	t.Parallel()
	t.Run("ASCII", func(t *testing.T) {
		t.Parallel()
		runes := FromRunes("hello").Collect()
		assert.Equal(t, []rune{'h', 'e', 'l', 'l', 'o'}, runes, "FromRunes should split ASCII into runes")
	})

	t.Run("Unicode", func(t *testing.T) {
		t.Parallel()
		runes := FromRunes("你好").Collect()
		assert.Equal(t, []rune{'你', '好'}, runes, "FromRunes should split Unicode into runes")
	})

	t.Run("Empty", func(t *testing.T) {
		t.Parallel()
		runes := FromRunes("").Collect()
		assert.Len(t, runes, 0, "FromRunes on empty should return zero runes")
	})

	t.Run("EarlyTermination", func(t *testing.T) {
		t.Parallel()
		runes := FromRunes("hello").Limit(3).Collect()
		assert.Equal(t, []rune{'h', 'e', 'l'}, runes, "FromRunes should respect Limit")
	})
}

func TestFromBytes(t *testing.T) {
	t.Parallel()
	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		bytes := FromBytes([]byte{1, 2, 3}).Collect()
		assert.Equal(t, []byte{1, 2, 3}, bytes, "FromBytes should return all bytes")
	})

	t.Run("Empty", func(t *testing.T) {
		t.Parallel()
		bytes := FromBytes([]byte{}).Collect()
		assert.Len(t, bytes, 0, "FromBytes on empty should return zero bytes")
	})

	t.Run("EarlyTermination", func(t *testing.T) {
		t.Parallel()
		bytes := FromBytes([]byte{1, 2, 3, 4, 5}).Limit(3).Collect()
		assert.Equal(t, []byte{1, 2, 3}, bytes, "FromBytes should respect Limit")
	})
}

func TestFromCSV(t *testing.T) {
	t.Parallel()
	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		csv := "a,b,c\n1,2,3\n4,5,6"
		records := FromCSV(strings.NewReader(csv)).Collect()
		assert.Len(t, records, 3, "FromCSV should return header and data rows")
		assert.Equal(t, []string{"a", "b", "c"}, records[0], "FromCSV should parse header row")
		assert.Equal(t, []string{"1", "2", "3"}, records[1], "FromCSV should parse first data row")
		assert.Equal(t, []string{"4", "5", "6"}, records[2], "FromCSV should parse second data row")
	})

	t.Run("Empty", func(t *testing.T) {
		t.Parallel()
		records := FromCSV(strings.NewReader("")).Collect()
		assert.Len(t, records, 0, "FromCSV on empty should return zero records")
	})

	t.Run("EarlyTermination", func(t *testing.T) {
		t.Parallel()
		csv := "a,b\n1,2\n3,4\n5,6"
		records := FromCSV(strings.NewReader(csv)).Limit(2).Collect()
		assert.Len(t, records, 2, "FromCSV should respect Limit")
	})
}

func TestFromCSVErr(t *testing.T) {
	t.Parallel()
	t.Run("AllValid", func(t *testing.T) {
		t.Parallel()
		csv := "a,b\n1,2"
		results := FromCSVErr(strings.NewReader(csv)).Collect()
		assert.Len(t, results, 2, "FromCSVErr should return one result per row")
		assert.True(t, results[0].IsOk(), "FromCSVErr should return Ok for valid header row")
		assert.True(t, results[1].IsOk(), "FromCSVErr should return Ok for valid data row")
	})

	t.Run("EarlyTermination", func(t *testing.T) {
		t.Parallel()
		csv := "a,b\n1,2\n3,4\n5,6"
		results := FromCSVErr(strings.NewReader(csv)).Limit(2).Collect()
		assert.Len(t, results, 2, "FromCSVErr should respect Limit")
	})
}

func TestFromTSV(t *testing.T) {
	t.Parallel()
	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		tsv := "a\tb\tc\n1\t2\t3"
		records := FromTSV(strings.NewReader(tsv)).Collect()
		assert.Len(t, records, 2, "FromTSV should return header and data rows")
		assert.Equal(t, []string{"a", "b", "c"}, records[0], "FromTSV should parse header row")
	})

	t.Run("Empty", func(t *testing.T) {
		t.Parallel()
		records := FromTSV(strings.NewReader("")).Collect()
		assert.Len(t, records, 0, "FromTSV on empty should return zero records")
	})
}

func TestFromTSVErr(t *testing.T) {
	t.Parallel()
	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		tsv := "a\tb\n1\t2"
		results := FromTSVErr(strings.NewReader(tsv)).Collect()
		assert.Len(t, results, 2, "FromTSVErr should return one result per row")
		assert.True(t, results[0].IsOk(), "FromTSVErr should return Ok for valid header row")
	})
}

func TestFromCSVWithHeader(t *testing.T) {
	t.Parallel()
	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		csv := "name,age\nAlice,30\nBob,25"
		records := FromCSVWithHeader(strings.NewReader(csv)).Collect()
		assert.Len(t, records, 2, "FromCSVWithHeader should return data rows only")
		assert.Equal(t, "Alice", records[0]["name"], "FromCSVWithHeader should parse first row name")
		assert.Equal(t, "30", records[0]["age"], "FromCSVWithHeader should parse first row age")
		assert.Equal(t, "Bob", records[1]["name"], "FromCSVWithHeader should parse second row name")
	})

	t.Run("Empty", func(t *testing.T) {
		t.Parallel()
		records := FromCSVWithHeader(strings.NewReader("")).Collect()
		assert.Len(t, records, 0, "FromCSVWithHeader on empty should return zero records")
	})

	t.Run("HeaderOnly", func(t *testing.T) {
		t.Parallel()
		csv := "name,age"
		records := FromCSVWithHeader(strings.NewReader(csv)).Collect()
		assert.Len(t, records, 0, "FromCSVWithHeader with header only should return zero records")
	})
}

func TestFromCSVWithHeaderErr(t *testing.T) {
	t.Parallel()
	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		csv := "name,age\nAlice,30"
		results := FromCSVWithHeaderErr(strings.NewReader(csv)).Collect()
		assert.Len(t, results, 1, "FromCSVWithHeaderErr should return one result per data row")
		assert.True(t, results[0].IsOk(), "FromCSVWithHeaderErr should return Ok for valid row")
		assert.Equal(t, "Alice", results[0].Value()["name"], "FromCSVWithHeaderErr should parse name field")
	})

	t.Run("Empty", func(t *testing.T) {
		t.Parallel()
		results := FromCSVWithHeaderErr(strings.NewReader("")).Collect()
		assert.Len(t, results, 1, "FromCSVWithHeaderErr should return a header error result") // Header error
		assert.True(t, results[0].IsErr(), "FromCSVWithHeaderErr should return Err for invalid header")
	})
}

func TestCSVRecord(t *testing.T) {
	t.Parallel()
	t.Run("Get", func(t *testing.T) {
		t.Parallel()
		record := CSVRecord{"name": "Alice", "age": "30"}
		assert.Equal(t, "Alice", record.Get("name"), "CSVRecord.Get should return existing field")
		assert.Equal(t, "", record.Get("missing"), "CSVRecord.Get should return empty string for missing field")
	})

	t.Run("GetOr", func(t *testing.T) {
		t.Parallel()
		record := CSVRecord{"name": "Alice"}
		assert.Equal(t, "Alice", record.GetOr("name", "default"), "CSVRecord.GetOr should return existing field")
		assert.Equal(t, "default", record.GetOr("missing", "default"), "CSVRecord.GetOr should return default for missing field")
	})
}

func TestFileLineStream(t *testing.T) {
	t.Parallel()
	t.Run("FromFileLines", func(t *testing.T) {
		t.Parallel()
		path := createTempFile(t, "test.txt", "line1\nline2\nline3")
		stream, err := FromFileLines(path)
		require.NoError(t, err, "FromFileLines should open existing file")
		defer func() { _ = stream.Close() }()

		lines := stream.Collect()
		assert.Equal(t, []string{"line1", "line2", "line3"}, lines, "FromFileLines should read all lines")
	})

	t.Run("FromFileLinesNotExists", func(t *testing.T) {
		t.Parallel()
		_, err := FromFileLines("/nonexistent/file.txt")
		assert.Error(t, err, "FromFileLines should return error for missing file")
	})

	t.Run("MustFromFileLinesPanics", func(t *testing.T) {
		t.Parallel()
		assert.Panics(t, func() {
			MustFromFileLines("/nonexistent/file.txt")
		}, "MustFromFileLines should panic for missing file")
	})
}

func TestToWriter(t *testing.T) {
	t.Parallel()
	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		var buf strings.Builder
		err := ToWriter(Of(1, 2, 3), &buf, func(n int) string {
			return string(rune('a' + n - 1))
		})
		assert.NoError(t, err, "ToWriter should not error on valid input")
		assert.Equal(t, "a\nb\nc\n", buf.String(), "ToWriter should write each element with newline")
	})
}

// errorWriter is a writer that always returns an error.
type errorWriter struct {
	failAfter int
	count     int
}

func (e *errorWriter) Write(p []byte) (n int, err error) {
	e.count++
	if e.failAfter < 0 || e.count > e.failAfter {
		return 0, errors.New("write error")
	}
	return len(p), nil
}

func TestToFileError(t *testing.T) {
	t.Parallel()
	t.Run("CreateError", func(t *testing.T) {
		t.Parallel()
		err := ToFile(Of("a", "b"), "/nonexistent/dir/file.txt", func(s string) string { return s })
		assert.Error(t, err, "ToFile should return error when file creation fails")
	})
}

// Note: ToCSV error testing is challenging because csv.Writer uses internal buffering
// and errors may not propagate immediately. The error handling code in ToCSV is
// correct but difficult to test without a custom csv.Writer implementation.

func TestToCSVFileError(t *testing.T) {
	t.Parallel()
	t.Run("CreateError", func(t *testing.T) {
		t.Parallel()
		err := ToCSVFile(Of([]string{"a", "b"}), "/nonexistent/dir/file.csv")
		assert.Error(t, err, "ToCSVFile should return error when file creation fails")
	})
}

func TestToFile(t *testing.T) {
	t.Parallel()
	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		path := createTempFile(t, "out.txt", "")
		err := ToFile(Of("a", "b", "c"), path, func(s string) string { return s })
		assert.NoError(t, err, "ToFile should write without error")

		content, err := os.ReadFile(path)
		require.NoError(t, err, "ToFile should create readable output file")
		assert.Equal(t, "a\nb\nc\n", string(content), "ToFile should write lines with newlines")
	})
}

func TestToCSV(t *testing.T) {
	t.Parallel()
	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		var buf strings.Builder
		err := ToCSV(Of([]string{"a", "b"}, []string{"1", "2"}), &buf)
		assert.NoError(t, err, "ToCSV should not error on valid input")
		assert.Equal(t, "a,b\n1,2\n", buf.String(), "ToCSV should write CSV rows")
	})
}

func TestToCSVFile(t *testing.T) {
	t.Parallel()
	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		path := createTempFile(t, "data.csv", "")
		err := ToCSVFile(Of([]string{"a", "b"}, []string{"1", "2"}), path)
		assert.NoError(t, err, "ToCSVFile should not error on valid input")

		content, err := os.ReadFile(path)
		require.NoError(t, err, "ToCSVFile should create readable output file")
		assert.Equal(t, "a,b\n1,2\n", string(content), "ToCSVFile should write CSV rows")
	})
}

func TestFromCSVFile(t *testing.T) {
	t.Parallel()
	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		path := createTempFile(t, "in.csv", "a,b\n1,2")
		stream, err := FromCSVFile(path)
		require.NoError(t, err, "FromCSVFile should open existing file")
		defer func() { _ = stream.Close() }()

		records := stream.Collect()
		assert.Len(t, records, 2, "FromCSVFile should read all rows")
	})

	t.Run("NotExists", func(t *testing.T) {
		t.Parallel()
		_, err := FromCSVFile("/nonexistent/file.csv")
		assert.Error(t, err, "FromCSVFile should return error for missing file")
	})
}

// TestFromReaderLinesErr tests FromReaderLinesErr function.
func TestFromReaderLinesErr(t *testing.T) {
	t.Parallel()
	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		reader := strings.NewReader("line1\nline2\nline3")
		results := FromReaderLinesErr(reader).Collect()
		assert.Len(t, results, 3, "FromReaderLinesErr should return one result per line")
		assert.True(t, results[0].IsOk(), "FromReaderLinesErr should return Ok for valid lines")
		assert.Equal(t, "line1", results[0].Value(), "FromReaderLinesErr should preserve line content")
		assert.True(t, results[1].IsOk(), "FromReaderLinesErr should return Ok for valid lines")
		assert.Equal(t, "line2", results[1].Value(), "FromReaderLinesErr should preserve line content")
	})

	t.Run("Empty", func(t *testing.T) {
		t.Parallel()
		reader := strings.NewReader("")
		results := FromReaderLinesErr(reader).Collect()
		assert.Len(t, results, 0, "FromReaderLinesErr on empty should return zero results")
	})

	t.Run("EarlyTermination", func(t *testing.T) {
		t.Parallel()
		reader := strings.NewReader("a\nb\nc\nd\ne")
		results := FromReaderLinesErr(reader).Limit(2).Collect()
		assert.Len(t, results, 2, "FromReaderLinesErr should respect Limit")
	})
}

// TestFromScannerErrEarlyBreak tests early termination in FromScannerErr.
func TestFromScannerErrEarlyBreak(t *testing.T) {
	t.Parallel()
	t.Run("EarlyTermination", func(t *testing.T) {
		t.Parallel()
		scanner := bufio.NewScanner(strings.NewReader("1\n2\n3\n4\n5"))
		results := FromScannerErr(scanner).Limit(2).Collect()
		assert.Len(t, results, 2, "FromScannerErr should respect Limit")
		assert.True(t, results[0].IsOk(), "FromScannerErr should return Ok for valid lines")
		assert.Equal(t, "1", results[0].Value(), "FromScannerErr should preserve line content")
	})
}

// TestFromTSVErrMore tests more cases for FromTSVErr.
func TestFromTSVErrMore(t *testing.T) {
	t.Parallel()
	t.Run("Empty", func(t *testing.T) {
		t.Parallel()
		results := FromTSVErr(strings.NewReader("")).Collect()
		assert.Len(t, results, 0, "FromTSVErr on empty should return zero results")
	})

	t.Run("EarlyTermination", func(t *testing.T) {
		t.Parallel()
		tsv := "a\tb\n1\t2\n3\t4\n5\t6"
		results := FromTSVErr(strings.NewReader(tsv)).Limit(2).Collect()
		assert.Len(t, results, 2, "FromTSVErr should respect Limit")
	})

	t.Run("MultipleRecords", func(t *testing.T) {
		t.Parallel()
		tsv := "name\tage\nAlice\t30\nBob\t25\nCharlie\t35"
		results := FromTSVErr(strings.NewReader(tsv)).Collect()
		assert.Len(t, results, 4, "FromTSVErr should return header and data rows")
		for _, r := range results {
			assert.True(t, r.IsOk(), "FromTSVErr should return Ok for valid rows")
		}
	})
}

// TestToWriterError tests ToWriter with error conditions.
func TestToWriterError(t *testing.T) {
	t.Parallel()
	t.Run("Empty", func(t *testing.T) {
		t.Parallel()
		var buf strings.Builder
		err := ToWriter(Empty[int](), &buf, func(n int) string {
			return string(rune('a' + n - 1))
		})
		assert.NoError(t, err, "ToWriter should not error on empty input")
		assert.Equal(t, "", buf.String(), "ToWriter on empty should produce empty output")
	})

	t.Run("EarlyTerminationInFormat", func(t *testing.T) {
		t.Parallel()
		var buf strings.Builder
		err := ToWriter(Of(1, 2, 3), &buf, func(n int) string {
			return string(rune('0' + n))
		})
		assert.NoError(t, err, "ToWriter should not error when format returns strings")
		assert.Contains(t, buf.String(), "1", "ToWriter output should include formatted elements")
	})

	t.Run("WriteStringError", func(t *testing.T) {
		t.Parallel()
		w := &errorWriter{failAfter: 0}
		err := ToWriter(Of(1, 2, 3), w, func(n int) string {
			return string(rune('a' + n - 1))
		})
		assert.Error(t, err, "ToWriter should return error when WriteString fails")
	})

	t.Run("WriteNewlineError", func(t *testing.T) {
		t.Parallel()
		// bufio.Writer has a default buffer of 4096 bytes.
		// Write 4096 bytes (fills buffer exactly).
		// When we write \n, bufio flushes the 4096 bytes first (1st write).
		// If 1st write fails, WriteString("\n") returns error -> covers io.go:376.
		largeString := strings.Repeat("x", 4096)
		w := &errorWriter{failAfter: 0} // Fail on 1st write (which happens inside WriteString("\n"))
		err := ToWriter(Of(largeString), w, func(s string) string { return s })
		assert.Error(t, err, "ToWriter should return error when writing newline fails")
	})
}

// TestFileLineStreamClose tests FileLineStream.Close with nil file.
func TestFileLineStreamClose(t *testing.T) {
	t.Parallel()
	t.Run("NilFile", func(t *testing.T) {
		t.Parallel()
		fls := &FileLineStream{file: nil}
		err := fls.Close()
		assert.NoError(t, err, "FileLineStream.Close should succeed with nil file")
	})
}

// TestCSVStreamClose tests CSVStream.Close with nil closer.
func TestCSVStreamClose(t *testing.T) {
	t.Parallel()
	t.Run("NilCloser", func(t *testing.T) {
		t.Parallel()
		cs := &CSVStream{closer: nil}
		err := cs.Close()
		assert.NoError(t, err, "CSVStream.Close should succeed with nil closer")
	})
}

// TestMustFromFileLinesSuccess tests MustFromFileLines with valid file.
func TestMustFromFileLinesSuccess(t *testing.T) {
	t.Parallel()
	path := createTempFile(t, "must.txt", "line1\nline2")
	// Should not panic
	fls := MustFromFileLines(path)
	defer func() { _ = fls.Close() }()

	lines := fls.Collect()
	assert.Equal(t, []string{"line1", "line2"}, lines, "MustFromFileLines should read all lines")
}

// TestFromScannerErrWithError tests FromScannerErr with scanner error.
func TestFromScannerErrWithError(t *testing.T) {
	t.Parallel()
	t.Run("ScannerError", func(t *testing.T) {
		t.Parallel()
		// Create a scanner with a very small buffer that will error on long lines
		longLine := strings.Repeat("x", 1024*1024) // 1MB line
		scanner := bufio.NewScanner(strings.NewReader(longLine))
		scanner.Buffer(make([]byte, 100), 100) // Very small max buffer

		results := FromScannerErr(scanner).Collect()
		// Should have an error result
		hasError := false
		for _, r := range results {
			if r.IsErr() {
				hasError = true
				break
			}
		}
		assert.True(t, hasError, "Should have an error result when scanner fails")
	})
}

// TestFromCSVWithError tests FromCSV with malformed CSV.
func TestFromCSVWithError(t *testing.T) {
	t.Parallel()
	t.Run("MalformedCSV", func(t *testing.T) {
		t.Parallel()
		// CSV with wrong number of fields
		malformedCSV := "a,b,c\n1,2\n3,4,5"
		records := FromCSV(strings.NewReader(malformedCSV)).Collect()
		// Should stop at malformed record
		assert.Less(t, len(records), 3, "Should stop when encountering malformed CSV")
	})

	t.Run("MalformedQuotes", func(t *testing.T) {
		t.Parallel()
		// CSV with unclosed quote
		malformedCSV := "a,b\n\"unclosed,2"
		records := FromCSV(strings.NewReader(malformedCSV)).Collect()
		assert.LessOrEqual(t, len(records), 1, "Should stop when encountering malformed CSV")
	})
}

// TestFromCSVErrWithError tests FromCSVErr with malformed CSV.
func TestFromCSVErrWithError(t *testing.T) {
	t.Parallel()
	t.Run("MalformedCSV", func(t *testing.T) {
		t.Parallel()
		// CSV with wrong number of fields
		malformedCSV := "a,b,c\n1,2\n3,4,5"
		results := FromCSVErr(strings.NewReader(malformedCSV)).Collect()
		// Should have error result for malformed row
		hasError := false
		for _, r := range results {
			if r.IsErr() {
				hasError = true
				break
			}
		}
		assert.True(t, hasError, "Should yield error for malformed CSV")
	})

	t.Run("MalformedCSVEarlyTermination", func(t *testing.T) {
		t.Parallel()
		// CSV with error in the middle
		malformedCSV := "a,b\n1,2\n\"unclosed\n3,4\n5,6"
		results := FromCSVErr(strings.NewReader(malformedCSV)).Limit(3).Collect()
		assert.Len(t, results, 3, "FromCSVErr should respect Limit")
	})

	t.Run("ContinueAfterError", func(t *testing.T) {
		t.Parallel()
		// CSV with wrong field count (error) then valid row
		malformedCSV := "a,b,c\n1,2\n3,4,5"
		results := FromCSVErr(strings.NewReader(malformedCSV)).Collect()
		// First row valid, second row error, third row valid
		assert.GreaterOrEqual(t, len(results), 2, "FromCSVErr should return header and at least one data result")
		assert.True(t, results[0].IsOk(), "FromCSVErr should return Ok for header") // header
		// Check if there's an error followed by more results
		foundErrorThenOk := false
		for i := 0; i < len(results)-1; i++ {
			if results[i].IsErr() && i+1 < len(results) && results[i+1].IsOk() {
				foundErrorThenOk = true
				break
			}
		}
		assert.True(t, foundErrorThenOk, "Should continue after error and yield more results")
	})

	t.Run("StopOnError", func(t *testing.T) {
		t.Parallel()
		// Test early termination when consumer returns false on error
		// This covers io.go:188-190
		malformedCSV := "a,b\n\"unclosed\n3,4"
		count := 0
		FromCSVErr(strings.NewReader(malformedCSV)).seq(func(r Result[[]string]) bool {
			count++
			if r.IsErr() {
				return false // Stop on first error
			}
			return true
		})
		// Should have 1 valid row then stopped on error
		assert.GreaterOrEqual(t, count, 1, "FromCSVErr should yield at least one row before stopping on error")
	})
}

// TestFromCSVFileEarlyTermination tests FromCSVFile with early termination.
func TestFromCSVFileEarlyTermination(t *testing.T) {
	t.Parallel()
	t.Run("EarlyTermination", func(t *testing.T) {
		t.Parallel()
		tmpFile, err := os.CreateTemp("", "test*.csv")
		require.NoError(t, err, "CreateTemp should succeed")
		t.Cleanup(func() { _ = os.Remove(tmpFile.Name()) })

		_, err = tmpFile.WriteString("a,b\n1,2\n3,4\n5,6\n7,8")
		require.NoError(t, err, "Temp CSV should be written successfully")
		require.NoError(t, tmpFile.Close(), "Temp CSV file should close successfully")

		stream, err := FromCSVFile(tmpFile.Name())
		require.NoError(t, err, "FromCSVFile should open temp file")
		defer func() { _ = stream.Close() }()

		records := stream.Limit(2).Collect()
		assert.Len(t, records, 2, "FromCSVFile should respect Limit")
	})

	t.Run("MalformedCSV", func(t *testing.T) {
		t.Parallel()
		tmpFile, err := os.CreateTemp("", "test*.csv")
		require.NoError(t, err, "CreateTemp should succeed")
		t.Cleanup(func() { _ = os.Remove(tmpFile.Name()) })

		_, err = tmpFile.WriteString("a,b,c\n1,2\n3,4,5")
		require.NoError(t, err, "Temp CSV should be written successfully")
		require.NoError(t, tmpFile.Close(), "Temp CSV file should close successfully")

		stream, err := FromCSVFile(tmpFile.Name())
		require.NoError(t, err, "FromCSVFile should open temp file")
		defer func() { _ = stream.Close() }()

		records := stream.Collect()
		assert.Less(t, len(records), 3, "Should stop on malformed row")
	})
}

// TestFromTSVWithError tests FromTSV with malformed TSV.
func TestFromTSVWithError(t *testing.T) {
	t.Parallel()
	t.Run("MalformedTSV", func(t *testing.T) {
		t.Parallel()
		// TSV with wrong number of fields
		malformedTSV := "a\tb\tc\n1\t2\n3\t4\t5"
		records := FromTSV(strings.NewReader(malformedTSV)).Collect()
		assert.Less(t, len(records), 3, "Should stop on malformed TSV")
	})

	t.Run("EarlyTermination", func(t *testing.T) {
		t.Parallel()
		tsv := "a\tb\n1\t2\n3\t4\n5\t6\n7\t8"
		records := FromTSV(strings.NewReader(tsv)).Limit(2).Collect()
		assert.Len(t, records, 2, "FromTSV should respect Limit")
	})
}

// TestFromTSVErrWithError tests FromTSVErr with malformed TSV.
func TestFromTSVErrWithError(t *testing.T) {
	t.Parallel()
	t.Run("MalformedTSV", func(t *testing.T) {
		t.Parallel()
		// TSV with wrong number of fields
		malformedTSV := "a\tb\tc\n1\t2\n3\t4\t5"
		results := FromTSVErr(strings.NewReader(malformedTSV)).Collect()
		hasError := false
		for _, r := range results {
			if r.IsErr() {
				hasError = true
				break
			}
		}
		assert.True(t, hasError, "Should yield error for malformed TSV")
	})

	t.Run("ContinueAfterError", func(t *testing.T) {
		t.Parallel()
		// TSV with wrong field count (error) then valid row
		malformedTSV := "a\tb\tc\n1\t2\n3\t4\t5"
		results := FromTSVErr(strings.NewReader(malformedTSV)).Collect()
		assert.GreaterOrEqual(t, len(results), 2, "FromTSVErr should return header and at least one data result")
		// Check if there's an error followed by more results
		foundErrorThenOk := false
		for i := 0; i < len(results)-1; i++ {
			if results[i].IsErr() && i+1 < len(results) && results[i+1].IsOk() {
				foundErrorThenOk = true
				break
			}
		}
		assert.True(t, foundErrorThenOk, "Should continue after error and yield more results")
	})

	t.Run("StopOnError", func(t *testing.T) {
		t.Parallel()
		// Test early termination when consumer returns false on error
		// This covers io.go:264-266
		malformedTSV := "a\tb\tc\n1\t2\n3\t4\t5"
		count := 0
		FromTSVErr(strings.NewReader(malformedTSV)).seq(func(r Result[[]string]) bool {
			count++
			if r.IsErr() {
				return false // Stop on first error
			}
			return true
		})
		assert.GreaterOrEqual(t, count, 1, "FromTSVErr should yield at least one row before stopping on error")
	})
}

// TestFromCSVWithHeaderError tests FromCSVWithHeader error cases.
func TestFromCSVWithHeaderError(t *testing.T) {
	t.Parallel()
	t.Run("ReadError", func(t *testing.T) {
		t.Parallel()
		// Header with no data rows but with parse error
		malformedCSV := "name,age\n\"unclosed"
		records := FromCSVWithHeader(strings.NewReader(malformedCSV)).Collect()
		// Should handle error gracefully
		assert.LessOrEqual(t, len(records), 1, "FromCSVWithHeader should return at most header on parse error")
	})

	t.Run("EarlyTermination", func(t *testing.T) {
		t.Parallel()
		csv := "name,age\nAlice,30\nBob,25\nCharlie,35"
		records := FromCSVWithHeader(strings.NewReader(csv)).Limit(2).Collect()
		assert.Len(t, records, 2, "FromCSVWithHeader should respect Limit")
	})
}

// TestFromCSVWithHeaderErrError tests FromCSVWithHeaderErr error cases.
func TestFromCSVWithHeaderErrError(t *testing.T) {
	t.Parallel()
	t.Run("DataError", func(t *testing.T) {
		t.Parallel()
		// Valid header, then malformed data row
		csv := "name,age,city\nAlice,30\nBob,25,NYC"
		results := FromCSVWithHeaderErr(strings.NewReader(csv)).Collect()
		hasError := false
		for _, r := range results {
			if r.IsErr() {
				hasError = true
				break
			}
		}
		assert.True(t, hasError, "Should yield error for malformed data row")
	})

	t.Run("EarlyTermination", func(t *testing.T) {
		t.Parallel()
		csv := "name,age\nAlice,30\nBob,25\nCharlie,35"
		results := FromCSVWithHeaderErr(strings.NewReader(csv)).Limit(2).Collect()
		assert.Len(t, results, 2, "FromCSVWithHeaderErr should respect Limit")
	})

	t.Run("ContinueAfterError", func(t *testing.T) {
		t.Parallel()
		// Valid header, error row, then valid row
		csv := "name,age,city\nAlice,30\nBob,25,NYC"
		results := FromCSVWithHeaderErr(strings.NewReader(csv)).Collect()
		// Should have header, error, then valid
		foundErrorThenOk := false
		for i := 0; i < len(results)-1; i++ {
			if results[i].IsErr() && i+1 < len(results) && results[i+1].IsOk() {
				foundErrorThenOk = true
				break
			}
		}
		assert.True(t, foundErrorThenOk, "Should continue after error and yield more results")
	})

	t.Run("StopOnError", func(t *testing.T) {
		t.Parallel()
		// Test early termination when consumer returns false on error
		// This covers io.go:346-348
		csv := "name,age,city\nAlice,30\nBob,25,NYC"
		count := 0
		FromCSVWithHeaderErr(strings.NewReader(csv)).seq(func(r Result[CSVRecord]) bool {
			count++
			if r.IsErr() {
				return false // Stop on first error
			}
			return true
		})
		assert.GreaterOrEqual(t, count, 1, "FromCSVWithHeaderErr should yield at least one row before stopping on error")
	})
}

// TestToWriterErrorAfterFirst tests ToWriter with error after first write.
// Note: bufio.Writer has a default buffer of 4096 bytes. To trigger immediate
// errors, we need to write strings larger than the buffer size.
func TestToWriterErrorAfterFirst(t *testing.T) {
	t.Parallel()
	largeString := strings.Repeat("x", 5000) // Exceeds 4096 byte buffer

	t.Run("ErrorOnLargeWrite", func(t *testing.T) {
		t.Parallel()
		w := &errorWriter{failAfter: 0} // Fail on first write
		err := ToWriter(Of(1, 2, 3), w, func(_ int) string {
			return largeString
		})
		assert.Error(t, err, "ToWriter should return error when writing large string fails")
	})

	t.Run("ErrorOnFlush", func(t *testing.T) {
		t.Parallel()
		// Small writes get buffered, error happens on Flush
		w := &errorWriter{failAfter: 0}
		err := ToWriter(Of(1, 2, 3), w, func(n int) string {
			return string(rune('a' + n - 1))
		})
		assert.Error(t, err, "ToWriter should return error when Flush fails")
	})
}

// TestToCSVError tests ToCSV with error conditions.
// Note: csv.Writer uses internal buffering. Errors may surface via csvWriter.Error()
// at the end rather than immediately during Write().
func TestToCSVError(t *testing.T) {
	t.Parallel()
	largeField := strings.Repeat("x", 5000) // Large field to exceed buffer

	t.Run("WriteErrorLargeData", func(t *testing.T) {
		t.Parallel()
		w := &errorWriter{failAfter: 0} // Fail on first write
		err := ToCSV(Of([]string{largeField, largeField}), w)
		assert.Error(t, err, "ToCSV should return error when writing large data fails")
	})

	t.Run("Empty", func(t *testing.T) {
		t.Parallel()
		var buf strings.Builder
		err := ToCSV(Empty[[]string](), &buf)
		assert.NoError(t, err, "ToCSV should not error on empty input")
		assert.Equal(t, "", buf.String(), "ToCSV on empty should produce empty output")
	})
}

// TestToWriterEarlyTermination tests ToWriter when format function can determine early termination.
func TestToWriterEarlyTermination(t *testing.T) {
	t.Parallel()
	t.Run("WriteAfterSuccess", func(t *testing.T) {
		t.Parallel()
		var buf strings.Builder
		err := ToWriter(Of(1, 2, 3, 4, 5), &buf, func(n int) string {
			return string(rune('0' + n))
		})
		assert.NoError(t, err, "ToWriter should not error on valid input")
		assert.Equal(t, "1\n2\n3\n4\n5\n", buf.String(), "ToWriter should write all formatted lines")
	})
}

// TestMustFromCSVFile tests MustFromCSVFile function.
func TestMustFromCSVFile(t *testing.T) {
	t.Parallel()
	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		path := createTempFile(t, "must.csv", "a,b\n1,2\n3,4")
		cs := MustFromCSVFile(path)
		defer func() { _ = cs.Close() }()

		records := cs.Collect()
		assert.Len(t, records, 3, "MustFromCSVFile should read all rows")
	})

	t.Run("Panics", func(t *testing.T) {
		t.Parallel()
		assert.Panics(t, func() {
			MustFromCSVFile("/nonexistent/file.csv")
		}, "MustFromCSVFile should panic for missing file")
	})
}

// TestFromTSVFile tests FromTSVFile function.
func TestFromTSVFile(t *testing.T) {
	t.Parallel()
	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		path := createTempFile(t, "test.tsv", "a\tb\n1\t2\n3\t4")
		stream, err := FromTSVFile(path)
		require.NoError(t, err, "FromTSVFile should open existing file")
		defer func() { _ = stream.Close() }()

		records := stream.Collect()
		assert.Len(t, records, 3, "FromTSVFile should read all rows")
		assert.Equal(t, []string{"a", "b"}, records[0], "FromTSVFile should parse header row")
	})

	t.Run("NotExists", func(t *testing.T) {
		t.Parallel()
		_, err := FromTSVFile("/nonexistent/file.tsv")
		assert.Error(t, err, "FromTSVFile should return error for missing file")
	})

	t.Run("EarlyTermination", func(t *testing.T) {
		t.Parallel()
		path := createTempFile(t, "test.tsv", "a\tb\n1\t2\n3\t4\n5\t6")
		stream, err := FromTSVFile(path)
		require.NoError(t, err, "FromTSVFile should open existing file")
		defer func() { _ = stream.Close() }()

		records := stream.Limit(2).Collect()
		assert.Len(t, records, 2, "FromTSVFile should respect Limit")
	})

	t.Run("MalformedTSV", func(t *testing.T) {
		t.Parallel()
		path := createTempFile(t, "test.tsv", "a\tb\tc\n1\t2\n3\t4\t5")
		stream, err := FromTSVFile(path)
		require.NoError(t, err, "FromTSVFile should open existing file")
		defer func() { _ = stream.Close() }()

		records := stream.Collect()
		assert.Less(t, len(records), 3, "FromTSVFile should stop on malformed row")
	})
}

// TestMustFromTSVFile tests MustFromTSVFile function.
func TestMustFromTSVFile(t *testing.T) {
	t.Parallel()
	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		path := createTempFile(t, "must.tsv", "a\tb\n1\t2\n3\t4")
		cs := MustFromTSVFile(path)
		defer func() { _ = cs.Close() }()

		records := cs.Collect()
		assert.Len(t, records, 3, "MustFromTSVFile should read all rows")
	})

	t.Run("Panics", func(t *testing.T) {
		t.Parallel()
		assert.Panics(t, func() {
			MustFromTSVFile("/nonexistent/file.tsv")
		}, "MustFromTSVFile should panic for missing file")
	})
}

// TestUsing tests the Using function for resource management.
func TestUsing(t *testing.T) {
	t.Parallel()
	t.Run("FileLineStream", func(t *testing.T) {
		t.Parallel()
		path := createTempFile(t, "using.txt", "line1\nline2\nline3")

		count := Using(MustFromFileLines(path), func(s *FileLineStream) int {
			return s.Count()
		})
		assert.Equal(t, 3, count, "Using should return block result")
	})

	t.Run("CSVStream", func(t *testing.T) {
		t.Parallel()
		path := createTempFile(t, "using.csv", "a,b\n1,2\n3,4")

		records := Using(MustFromCSVFile(path), func(s *CSVStream) [][]string {
			return s.Collect()
		})
		assert.Len(t, records, 3, "Using should return collected records")
	})

	t.Run("TSVStream", func(t *testing.T) {
		t.Parallel()
		path := createTempFile(t, "using.tsv", "a\tb\n1\t2")

		records := Using(MustFromTSVFile(path), func(s *CSVStream) [][]string {
			return s.Collect()
		})
		assert.Len(t, records, 2, "Using should return collected TSV records")
	})

	t.Run("WithFilter", func(t *testing.T) {
		t.Parallel()
		path := createTempFile(t, "using_filter.txt", "apple\nbanana\napricot\ncherry")

		lines := Using(MustFromFileLines(path), func(s *FileLineStream) []string {
			return s.Filter(func(line string) bool {
				return len(line) > 0 && line[0] == 'a'
			}).Collect()
		})
		assert.Equal(t, []string{"apple", "apricot"}, lines, "Using should work with filter")
	})

	t.Run("WithMap", func(t *testing.T) {
		t.Parallel()
		path := createTempFile(t, "using_map.txt", "1\n2\n3")

		sum := Using(MustFromFileLines(path), func(s *FileLineStream) int {
			total := 0
			s.ForEach(func(line string) {
				if n := len(line); n > 0 {
					total += int(line[0] - '0')
				}
			})
			return total
		})
		assert.Equal(t, 6, sum, "Using should work with ForEach")
	})
}
