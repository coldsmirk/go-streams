package join

import (
	"testing"

	streams "github.com/coldsmirk/go-streams/v2"
)

const benchN = 1000

type row struct {
	ID   int
	Name string
}

// keyed input with a controllable number of rows per key.
func rows(n, perKey int) streams.Stream2[int, row] {
	xs := make([]row, n)
	for i := range xs {
		xs[i] = row{ID: i / perKey, Name: "r"}
	}
	return func(yield func(int, row) bool) {
		for _, r := range xs {
			if !yield(r.ID, r) {
				return
			}
		}
	}
}

var sink int

func benchJoin(b *testing.B, perKey int, run func(a, c streams.Stream2[int, row]) int) {
	b.ReportAllocs()
	for b.Loop() {
		sink = run(rows(benchN, perKey), rows(benchN, perKey))
	}
}

func BenchmarkInner(b *testing.B) {
	benchJoin(b, 1, func(a, c streams.Stream2[int, row]) int {
		return Inner(a, c, func(k int, _, _ row) int { return k }).Count()
	})
}

func BenchmarkInnerFanOut(b *testing.B) {
	benchJoin(b, 100, func(a, c streams.Stream2[int, row]) int {
		return Inner(a, c, func(k int, _, _ row) int { return k }).Count()
	})
}

func BenchmarkLeft(b *testing.B) {
	benchJoin(b, 1, func(a, c streams.Stream2[int, row]) int {
		return Left(a, c, func(k int, _, _ row, _ bool) int { return k }).Count()
	})
}

func BenchmarkRight(b *testing.B) {
	benchJoin(b, 1, func(a, c streams.Stream2[int, row]) int {
		return Right(a, c, func(k int, _ row, _ bool, _ row) int { return k }).Count()
	})
}

func BenchmarkFull(b *testing.B) {
	benchJoin(b, 1, func(a, c streams.Stream2[int, row]) int {
		return Full(a, c, func(k int, _ row, _ bool, _ row, _ bool) int { return k }).Count()
	})
}

func BenchmarkGroup(b *testing.B) {
	benchJoin(b, 4, func(a, c streams.Stream2[int, row]) int {
		return Group(a, c, func(_ int, l, r []row) int { return len(l) + len(r) }).Count()
	})
}

func BenchmarkSemi(b *testing.B) {
	benchJoin(b, 1, func(a, c streams.Stream2[int, row]) int {
		return Semi(a, c).Count()
	})
}
