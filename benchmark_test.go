package streams

import (
	"cmp"
	"iter"
	"math/rand/v2"
	"slices"
	"testing"
)

// The input size for every benchmark. Large enough that per-element cost
// dominates the one-off cost of building the pipeline.
const benchN = 1000

func benchInput() []int {
	xs := make([]int, benchN)
	for i := range xs {
		xs[i] = i
	}
	return xs
}

var (
	sinkInt   int
	sinkSlice []int
	sinkMap   map[int][]int
)

// --- the cost of the abstraction ---
//
// Each of these has a hand-written loop counterpart, so the difference is
// exactly what a Stream costs over writing the loop out.

func BenchmarkLoop_MapFilterSum(b *testing.B) {
	xs := benchInput()
	b.ReportAllocs()
	for b.Loop() {
		total := 0
		for _, v := range xs {
			w := v * 2
			if w%3 == 0 {
				total += w
			}
		}
		sinkInt = total
	}
}

func BenchmarkStream_MapFilterSum(b *testing.B) {
	xs := benchInput()
	b.ReportAllocs()
	for b.Loop() {
		sinkInt = Of(xs...).
			Map(func(v int) int { return v * 2 }).
			Filter(func(v int) bool { return v%3 == 0 }).
			Fold(0, func(a, v int) int { return a + v })
	}
}

func BenchmarkLoop_Collect(b *testing.B) {
	xs := benchInput()
	b.ReportAllocs()
	for b.Loop() {
		out := make([]int, 0, len(xs))
		out = append(out, xs...)
		sinkSlice = out
	}
}

func BenchmarkStream_Collect(b *testing.B) {
	xs := benchInput()
	b.ReportAllocs()
	for b.Loop() {
		sinkSlice = Of(xs...).Collect()
	}
}

// --- single operators ---

func BenchmarkMap(b *testing.B) {
	xs := benchInput()
	b.ReportAllocs()
	for b.Loop() {
		sinkSlice = Of(xs...).Map(func(v int) int { return v * 2 }).Collect()
	}
}

func BenchmarkMapToOtherType(b *testing.B) {
	xs := benchInput()
	b.ReportAllocs()
	for b.Loop() {
		n := 0
		for range Of(xs...).Map(func(v int) int64 { return int64(v) }) {
			n++
		}
		sinkInt = n
	}
}

func BenchmarkFilter(b *testing.B) {
	xs := benchInput()
	b.ReportAllocs()
	for b.Loop() {
		sinkSlice = Of(xs...).Filter(func(v int) bool { return v%2 == 0 }).Collect()
	}
}

func BenchmarkFold(b *testing.B) {
	xs := benchInput()
	b.ReportAllocs()
	for b.Loop() {
		sinkInt = Of(xs...).Fold(0, func(a, v int) int { return a + v })
	}
}

func BenchmarkTake(b *testing.B) {
	xs := benchInput()
	b.ReportAllocs()
	for b.Loop() {
		sinkSlice = Of(xs...).Take(benchN / 2).Collect()
	}
}

func BenchmarkScan(b *testing.B) {
	xs := benchInput()
	b.ReportAllocs()
	for b.Loop() {
		sinkSlice = Of(xs...).Scan(0, func(a, v int) int { return a + v }).Collect()
	}
}

// --- operators that buffer or build a map ---

func BenchmarkSort(b *testing.B) {
	// Not a reversed run: that is pdqsort's best case and hides most of the
	// difference between a direct comparison and a comparator call.
	xs := benchInput()
	r := rand.New(rand.NewPCG(1, 2))
	r.Shuffle(len(xs), func(i, j int) { xs[i], xs[j] = xs[j], xs[i] })
	b.ReportAllocs()
	for b.Loop() {
		sinkSlice = Sort(Of(xs...)).Collect()
	}
}

func BenchmarkDistinct(b *testing.B) {
	xs := benchInput()
	b.ReportAllocs()
	for b.Loop() {
		sinkSlice = Distinct(Of(xs...)).Collect()
	}
}

func BenchmarkGroupBy(b *testing.B) {
	xs := benchInput()
	b.ReportAllocs()
	for b.Loop() {
		sinkMap = Of(xs...).GroupBy(func(v int) int { return v % 10 })
	}
}

func BenchmarkReverse(b *testing.B) {
	xs := benchInput()
	b.ReportAllocs()
	for b.Loop() {
		sinkSlice = Of(xs...).Reverse().Collect()
	}
}

// --- regrouping ---

func BenchmarkChunk(b *testing.B) {
	xs := benchInput()
	b.ReportAllocs()
	for b.Loop() {
		n := 0
		for c := range Chunk(Of(xs...), 16) {
			n += len(c)
		}
		sinkInt = n
	}
}

func BenchmarkWindow(b *testing.B) {
	xs := benchInput()
	b.ReportAllocs()
	for b.Loop() {
		n := 0
		for w := range Window(Of(xs...), 16) {
			n += len(w)
		}
		sinkInt = n
	}
}

// --- combining ---

func BenchmarkZip(b *testing.B) {
	xs := benchInput()
	b.ReportAllocs()
	for b.Loop() {
		n := 0
		for range Of(xs...).Zip(Of(xs...)) {
			n++
		}
		sinkInt = n
	}
}

func BenchmarkZipWith(b *testing.B) {
	xs := benchInput()
	b.ReportAllocs()
	for b.Loop() {
		sinkInt = Of(xs...).
			ZipWith(Of(xs...), func(a, c int) int { return a + c }).
			Fold(0, func(a, v int) int { return a + v })
	}
}

func BenchmarkMerge(b *testing.B) {
	half := benchN / 2
	a := make([]int, half)
	c := make([]int, half)
	for i := range half {
		a[i], c[i] = 2*i, 2*i+1
	}
	b.ReportAllocs()
	for b.Loop() {
		sinkSlice = Merge(cmp.Compare[int], Of(a...), Of(c...)).Collect()
	}
}

func BenchmarkConcat(b *testing.B) {
	xs := benchInput()
	b.ReportAllocs()
	for b.Loop() {
		sinkSlice = Concat(Of(xs...), Of(xs...)).Collect()
	}
}

// --- Stream2 ---

func BenchmarkStream2MapValues(b *testing.B) {
	xs := benchInput()
	b.ReportAllocs()
	for b.Loop() {
		sinkInt = Of(xs...).Enumerate().
			MapValues(func(v int) int { return v * 2 }).
			Fold(0, func(a, _, v int) int { return a + v })
	}
}

// --- parallel ---
//
// The work per element is deliberately tiny, so these measure the machinery's
// overhead rather than the speed-up. A real workload amortises it.

func BenchmarkParallelMapOrdered(b *testing.B) {
	xs := benchInput()
	b.ReportAllocs()
	for b.Loop() {
		sinkSlice = Of(xs...).ParallelMap(func(v int) int { return v * 2 },
			WithConcurrency(4)).Collect()
	}
}

func BenchmarkParallelMapUnordered(b *testing.B) {
	xs := benchInput()
	b.ReportAllocs()
	for b.Loop() {
		sinkSlice = Of(xs...).ParallelMap(func(v int) int { return v * 2 },
			WithConcurrency(4), Unordered()).Collect()
	}
}

// --- interoperation cost ---
//
// A Stream is an iter.Seq, so crossing to the standard library should be free.

func BenchmarkStdlibCollect(b *testing.B) {
	xs := benchInput()
	b.ReportAllocs()
	for b.Loop() {
		sinkSlice = slices.Collect(iter.Seq[int](Of(xs...)))
	}
}
