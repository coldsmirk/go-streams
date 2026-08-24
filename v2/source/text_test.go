package source

import (
	"errors"
	"io"
	"iter"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"testing/iotest"
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
			if err != nil {
				t.Fatalf("Lines(%q) error = %v, want nil", tt.in, err)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("Lines(%q) = %q, want %q", tt.in, got, tt.want)
			}

			// StringLines must split a string exactly as Lines splits a reader.
			if s := StringLines(tt.in).Collect(); !slices.Equal(s, tt.want) {
				t.Errorf("StringLines(%q) = %q, want %q", tt.in, s, tt.want)
			}
		})
	}
}

func TestLinesReadError(t *testing.T) {
	got, err := collect2(Lines(iotest.ErrReader(errRead)))
	if !errors.Is(err, errRead) {
		t.Fatalf("Lines error = %v, want %v", err, errRead)
	}
	if len(got) != 0 {
		t.Errorf("Lines yielded %q before the error, want none", got)
	}
}

func TestLinesReadErrorAfterValues(t *testing.T) {
	r := io.MultiReader(strings.NewReader("alpha\nbeta\n"), iotest.ErrReader(errRead))
	got, err := collect2(Lines(r))
	if !errors.Is(err, errRead) {
		t.Fatalf("Lines error = %v, want %v", err, errRead)
	}
	if want := []string{"alpha", "beta"}; !slices.Equal(got, want) {
		t.Errorf("Lines = %q before the error, want %q", got, want)
	}
}

func TestLinesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lines.txt")
	if err := os.WriteFile(path, []byte("alpha\nbeta\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := collect2(LinesFile(path))
	if err != nil {
		t.Fatalf("LinesFile error = %v, want nil", err)
	}
	if want := []string{"alpha", "beta"}; !slices.Equal(got, want) {
		t.Errorf("LinesFile = %q, want %q", got, want)
	}
}

func TestLinesFileEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.txt")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := collect2(LinesFile(path))
	if err != nil || len(got) != 0 {
		t.Errorf("LinesFile on an empty file = %q, %v, want no lines and no error", got, err)
	}
}

func TestLinesFileMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "absent.txt")

	pairs := 0
	var last error
	for line, err := range LinesFile(path) {
		pairs++
		last = err
		if line != "" {
			t.Errorf("LinesFile on a missing file yielded line %q, want the empty string", line)
		}
	}
	if pairs != 1 {
		t.Fatalf("LinesFile on a missing file yielded %d pairs, want 1", pairs)
	}
	if !errors.Is(last, os.ErrNotExist) {
		t.Errorf("LinesFile error = %v, want one matching os.ErrNotExist", last)
	}
}

func TestBytes(t *testing.T) {
	if got := Bytes([]byte("abc")).Collect(); !slices.Equal(got, []byte("abc")) {
		t.Errorf("Bytes = %v, want %v", got, []byte("abc"))
	}
	if got := Bytes(nil).Count(); got != 0 {
		t.Errorf("Bytes(nil) yielded %d bytes, want 0", got)
	}
}

func TestRunes(t *testing.T) {
	got := Runes("héllo").Collect()
	if want := []rune("héllo"); !slices.Equal(got, want) {
		t.Errorf("Runes = %q, want %q", got, want)
	}
	if got := Runes("").Count(); got != 0 {
		t.Errorf("Runes(\"\") yielded %d runes, want 0", got)
	}
}
