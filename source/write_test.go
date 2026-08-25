package source

import (
	"bytes"
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	streams "github.com/coldsmirk/go-streams/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, WriteLines(&buf, streams.Of(1, 2, 3), strconv.Itoa), "WriteLines error")
	assert.Equal(t, "1\n2\n3\n", buf.String(), "WriteLines output")
}

func TestWriteLinesEmpty(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, WriteLines(&buf, streams.Empty[int](), strconv.Itoa), "WriteLines error")
	assert.Empty(t, buf.String(), "WriteLines wrote for an empty Stream, want nothing")
}

func TestWriteLinesFlushError(t *testing.T) {
	err := WriteLines(failWriter{}, streams.Of(1, 2, 3), strconv.Itoa)
	assert.ErrorIs(t, err, errWrite, "WriteLines error")
}

func TestWriteLinesStopsAtFirstError(t *testing.T) {
	formatted := 0
	err := WriteLines(failWriter{}, streams.Range(0, 100), func(int) string {
		formatted++
		return oversized
	})
	require.ErrorIs(t, err, errWrite, "WriteLines error")
	assert.Equal(t, 1, formatted, "WriteLines kept formatting after the write failed, want 1 element")
}

func TestWriteFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.txt")
	err := WriteFile(path, streams.Of("alpha", "beta"), func(s string) string { return s })
	require.NoError(t, err, "WriteFile error")

	// Reading the file back proves it was flushed and closed.
	got, err := collect2(LinesFile(path))
	require.NoError(t, err, "LinesFile error")
	assert.Equal(t, []string{"alpha", "beta"}, got, "WriteFile then LinesFile")
}

func TestWriteFileTruncates(t *testing.T) {
	path := writeTemp(t, "out.txt", "old\ncontent\nhere\n")
	err := WriteFile(path, streams.Of("new"), func(s string) string { return s })
	require.NoError(t, err, "WriteFile error")

	got, err := collect2(LinesFile(path))
	require.NoError(t, err, "LinesFile error")
	assert.Equal(t, []string{"new"}, got, "WriteFile over an existing file")
}

func TestWriteFileUncreatable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "absent", "out.txt")
	err := WriteFile(path, streams.Of("alpha"), func(s string) string { return s })
	assert.Error(t, err, "WriteFile into a missing directory = nil, want an error")
}

func TestWriteCSV(t *testing.T) {
	var buf bytes.Buffer
	s := streams.Of([]string{"a", "b"}, []string{"x,y", "z"})
	require.NoError(t, WriteCSV(&buf, s), "WriteCSV error")
	assert.Equal(t, "a,b\n\"x,y\",z\n", buf.String(), "WriteCSV output")
}

func TestWriteCSVEmpty(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, WriteCSV(&buf, streams.Empty[[]string]()), "WriteCSV error")
	assert.Empty(t, buf.String(), "WriteCSV wrote for an empty Stream, want nothing")
}

func TestWriteCSVFlushError(t *testing.T) {
	err := WriteCSV(failWriter{}, streams.Of([]string{"a", "b"}))
	assert.ErrorIs(t, err, errWrite, "WriteCSV error")
}

func TestWriteCSVStopsAtFirstError(t *testing.T) {
	pulled := 0
	s := streams.Range(0, 100).Map(func(int) []string {
		pulled++
		return []string{oversized}
	})
	require.ErrorIs(t, WriteCSV(failWriter{}, s), errWrite, "WriteCSV error")
	assert.Equal(t, 1, pulled, "WriteCSV kept consuming after the write failed, want 1 record")
}

func TestWriteCSVFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.csv")
	s := streams.Of([]string{"name", "city"}, []string{"Ada", "London"})
	require.NoError(t, WriteCSVFile(path, s), "WriteCSVFile error")

	got, err := collect2(RecordsFile(path))
	require.NoError(t, err, "RecordsFile error")
	require.Lenf(t, got, 1, "WriteCSVFile then RecordsFile = %v, want one Ada/London record", got)
	assert.Equalf(t, "Ada", got[0]["name"], "WriteCSVFile then RecordsFile = %v", got)
	assert.Equalf(t, "London", got[0]["city"], "WriteCSVFile then RecordsFile = %v", got)
}

func TestWriteCSVFileUncreatable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "absent", "out.csv")
	assert.Error(t, WriteCSVFile(path, streams.Of([]string{"a"})),
		"WriteCSVFile into a missing directory = nil, want an error")
}
