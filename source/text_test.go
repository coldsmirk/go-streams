package source

import (
	"errors"
	"io"
	"iter"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errRead = errors.New("source: test read failure")

// collect2 drains seq, returning the values yielded before the first error and
// that error, if any.
func collect2[T any](seq iter.Seq2[T, error]) ([]T, error) {
	var got []T
	for v, err := range seq {
		if err != nil {
			return got, err
		}
		got = append(got, v)
	}
	return got, nil
}

func TestLines(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{"empty", "", nil},
		{"one line", "alpha", []string{"alpha"}},
		{"terminated", "alpha\nbeta\n", []string{"alpha", "beta"}},
		{"unterminated", "alpha\nbeta", []string{"alpha", "beta"}},
		{"crlf", "alpha\r\nbeta\r\n", []string{"alpha", "beta"}},
		{"trailing cr", "alpha\r", []string{"alpha"}},
		{"blank lines", "\n\nalpha\n", []string{"", "", "alpha"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := collect2(Lines(strings.NewReader(tt.in)))
			require.NoErrorf(t, err, "Lines(%q)", tt.in)
			assert.Equalf(t, tt.want, got, "Lines(%q)", tt.in)

			// StringLines must split a string exactly as Lines splits a reader.
			assert.Equalf(t, tt.want, StringLines(tt.in).Collect(), "StringLines(%q)", tt.in)
		})
	}
}

func TestLinesReadError(t *testing.T) {
	got, err := collect2(Lines(iotest.ErrReader(errRead)))
	require.ErrorIs(t, err, errRead, "Lines error")
	assert.Empty(t, got, "Lines yielded lines before the error, want none")
}

func TestLinesReadErrorAfterValues(t *testing.T) {
	r := io.MultiReader(strings.NewReader("alpha\nbeta\n"), iotest.ErrReader(errRead))
	got, err := collect2(Lines(r))
	require.ErrorIs(t, err, errRead, "Lines error")
	assert.Equal(t, []string{"alpha", "beta"}, got, "Lines before the error")
}

func TestLinesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lines.txt")
	require.NoError(t, os.WriteFile(path, []byte("alpha\nbeta\n"), 0o600))
	got, err := collect2(LinesFile(path))
	require.NoError(t, err, "LinesFile error")
	assert.Equal(t, []string{"alpha", "beta"}, got, "LinesFile")
}

func TestLinesFileEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.txt")
	require.NoError(t, os.WriteFile(path, nil, 0o600))
	got, err := collect2(LinesFile(path))
	require.NoError(t, err, "LinesFile on an empty file, want no error")
	assert.Empty(t, got, "LinesFile on an empty file, want no lines")
}

func TestLinesFileMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "absent.txt")

	pairs := 0
	var last error
	for line, err := range LinesFile(path) {
		pairs++
		last = err
		assert.Empty(t, line, "LinesFile on a missing file yielded a line, want the empty string")
	}
	require.Equal(t, 1, pairs, "LinesFile on a missing file yielded the wrong number of pairs, want 1")
	assert.ErrorIs(t, last, os.ErrNotExist, "LinesFile error")
}

func TestBytes(t *testing.T) {
	assert.Equal(t, []byte("abc"), Bytes([]byte("abc")).Collect(), "Bytes")
	assert.Zero(t, Bytes(nil).Count(), "Bytes(nil) yielded bytes, want 0")
}

func TestRunes(t *testing.T) {
	assert.Equal(t, []rune("héllo"), Runes("héllo").Collect(), "Runes")
	assert.Zero(t, Runes("").Count(), `Runes("") yielded runes, want 0`)
}
