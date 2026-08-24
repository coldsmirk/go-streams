package source

import (
	"bytes"
	"errors"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	streams "github.com/coldsmirk/go-streams/v2"
)

var errWrite = errors.New("source: test write failure")

// failWriter rejects every write.
type failWriter struct{}

func (failWriter) Write([]byte) (int, error) { return 0, errWrite }

// oversized is longer than the buffer bufio.Writer and csv.Writer use, so
// writing it reaches the underlying writer immediately rather than being held
// until the final flush.
var oversized = strings.Repeat("x", 8192)

func TestWriteLines(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteLines(&buf, streams.Of(1, 2, 3), strconv.Itoa); err != nil {
		t.Fatalf("WriteLines error = %v, want nil", err)
	}
	if got, want := buf.String(), "1\n2\n3\n"; got != want {
		t.Errorf("WriteLines wrote %q, want %q", got, want)
	}
}

func TestWriteLinesEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteLines(&buf, streams.Empty[int](), strconv.Itoa); err != nil {
		t.Fatalf("WriteLines error = %v, want nil", err)
	}
	if buf.Len() != 0 {
		t.Errorf("WriteLines wrote %q for an empty Stream, want nothing", buf.String())
	}
}

func TestWriteLinesFlushError(t *testing.T) {
	err := WriteLines(failWriter{}, streams.Of(1, 2, 3), strconv.Itoa)
	if !errors.Is(err, errWrite) {
		t.Errorf("WriteLines error = %v, want %v", err, errWrite)
	}
}

func TestWriteLinesStopsAtFirstError(t *testing.T) {
	formatted := 0
	err := WriteLines(failWriter{}, streams.Range(0, 100), func(int) string {
		formatted++
		return oversized
	})
	if !errors.Is(err, errWrite) {
		t.Fatalf("WriteLines error = %v, want %v", err, errWrite)
	}
	if formatted != 1 {
		t.Errorf("WriteLines formatted %d elements after the write failed, want 1", formatted)
	}
}

func TestWriteFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.txt")
	if err := WriteFile(path, streams.Of("alpha", "beta"), func(s string) string { return s }); err != nil {
		t.Fatalf("WriteFile error = %v, want nil", err)
	}

	// Reading the file back proves it was flushed and closed.
	got, err := collect2(LinesFile(path))
	if err != nil {
		t.Fatalf("LinesFile error = %v, want nil", err)
	}
	if want := []string{"alpha", "beta"}; !slices.Equal(got, want) {
		t.Errorf("WriteFile then LinesFile = %q, want %q", got, want)
	}
}

func TestWriteFileTruncates(t *testing.T) {
	path := writeTemp(t, "out.txt", "old\ncontent\nhere\n")
	if err := WriteFile(path, streams.Of("new"), func(s string) string { return s }); err != nil {
		t.Fatalf("WriteFile error = %v, want nil", err)
	}
	got, err := collect2(LinesFile(path))
	if err != nil {
		t.Fatalf("LinesFile error = %v, want nil", err)
	}
	if want := []string{"new"}; !slices.Equal(got, want) {
		t.Errorf("WriteFile over an existing file left %q, want %q", got, want)
	}
}

func TestWriteFileUncreatable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "absent", "out.txt")
	if err := WriteFile(path, streams.Of("alpha"), func(s string) string { return s }); err == nil {
		t.Error("WriteFile into a missing directory = nil, want an error")
	}
}

func TestWriteCSV(t *testing.T) {
	var buf bytes.Buffer
	s := streams.Of([]string{"a", "b"}, []string{"x,y", "z"})
	if err := WriteCSV(&buf, s); err != nil {
		t.Fatalf("WriteCSV error = %v, want nil", err)
	}
	if got, want := buf.String(), "a,b\n\"x,y\",z\n"; got != want {
		t.Errorf("WriteCSV wrote %q, want %q", got, want)
	}
}

func TestWriteCSVEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteCSV(&buf, streams.Empty[[]string]()); err != nil {
		t.Fatalf("WriteCSV error = %v, want nil", err)
	}
	if buf.Len() != 0 {
		t.Errorf("WriteCSV wrote %q for an empty Stream, want nothing", buf.String())
	}
}

func TestWriteCSVFlushError(t *testing.T) {
	err := WriteCSV(failWriter{}, streams.Of([]string{"a", "b"}))
	if !errors.Is(err, errWrite) {
		t.Errorf("WriteCSV error = %v, want %v", err, errWrite)
	}
}

func TestWriteCSVStopsAtFirstError(t *testing.T) {
	pulled := 0
	s := streams.Range(0, 100).Map(func(int) []string {
		pulled++
		return []string{oversized}
	})
	if err := WriteCSV(failWriter{}, s); !errors.Is(err, errWrite) {
		t.Fatalf("WriteCSV error = %v, want %v", err, errWrite)
	}
	if pulled != 1 {
		t.Errorf("WriteCSV consumed %d records after the write failed, want 1", pulled)
	}
}

func TestWriteCSVFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.csv")
	s := streams.Of([]string{"name", "city"}, []string{"Ada", "London"})
	if err := WriteCSVFile(path, s); err != nil {
		t.Fatalf("WriteCSVFile error = %v, want nil", err)
	}

	got, err := collect2(RecordsFile(path))
	if err != nil {
		t.Fatalf("RecordsFile error = %v, want nil", err)
	}
	if len(got) != 1 || got[0]["name"] != "Ada" || got[0]["city"] != "London" {
		t.Errorf("WriteCSVFile then RecordsFile = %v, want one Ada/London record", got)
	}
}

func TestWriteCSVFileUncreatable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "absent", "out.csv")
	if err := WriteCSVFile(path, streams.Of([]string{"a"})); err == nil {
		t.Error("WriteCSVFile into a missing directory = nil, want an error")
	}
}
