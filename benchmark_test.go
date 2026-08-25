package streams

import (
	"testing"
)

// sink consumes a stream without allocating result slice.
// This measures the algorithm itself without collection overhead.
func sink[T any](s Stream[T]) int {
	count := 0
	for range s.seq {
		count++
	}
	return count
}

// BenchmarkTakeLast benchmarks the TakeLast operation with ring buffer.
func BenchmarkTakeLast(b *testing.B) {
	input := Range(1, 10001).Collect() // 10000 elements

	b.Run("TakeLast_10_Collect", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = FromSlice(input).TakeLast(10).Collect()
		}
	})

	b.Run("TakeLast_10_Count", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = sink(FromSlice(input).TakeLast(10))
		}
	})

	b.Run("TakeLast_1000_Collect", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = FromSlice(input).TakeLast(1000).Collect()
		}
	})

	b.Run("TakeLast_1000_Count", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = sink(FromSlice(input).TakeLast(1000))
		}
	})
}

// BenchmarkDropLast benchmarks the DropLast operation with ring buffer.
func BenchmarkDropLast(b *testing.B) {
	input := Range(1, 10001).Collect() // 10000 elements

	b.Run("DropLast_10_Collect", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = FromSlice(input).DropLast(10).Collect()
		}
	})

	b.Run("DropLast_10_Count", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = sink(FromSlice(input).DropLast(10))
		}
	})

	b.Run("DropLast_1000_Collect", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = FromSlice(input).DropLast(1000).Collect()
		}
	})

	b.Run("DropLast_1000_Count", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = sink(FromSlice(input).DropLast(1000))
		}
	})
}

// BenchmarkWindowWithStep benchmarks the WindowWithStep operation.
func BenchmarkWindowWithStep(b *testing.B) {
	input := Range(1, 10001).Collect() // 10000 elements

	b.Run("WindowWithStep_Size3_Step1_Collect", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = WindowWithStep(FromSlice(input), 3, 1, false).Collect()
		}
	})

	b.Run("WindowWithStep_Size3_Step1_Count", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = sink(WindowWithStep(FromSlice(input), 3, 1, false))
		}
	})

	b.Run("WindowWithStep_Size100_Step100_Collect", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = WindowWithStep(FromSlice(input), 100, 100, false).Collect()
		}
	})

	b.Run("WindowWithStep_Size100_Step100_Count", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = sink(WindowWithStep(FromSlice(input), 100, 100, false))
		}
	})
}

// BenchmarkScan benchmarks the Scan operation.
func BenchmarkScan(b *testing.B) {
	input := Range(1, 10001).Collect() // 10000 elements

	b.Run("Scan_Collect", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = Scan(FromSlice(input), 0, func(acc, v int) int { return acc + v }).Collect()
		}
	})

	b.Run("Scan_Count", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = sink(Scan(FromSlice(input), 0, func(acc, v int) int { return acc + v }))
		}
	})
}

// BenchmarkDistinctUntilChanged benchmarks the DistinctUntilChanged operation.
func BenchmarkDistinctUntilChanged(b *testing.B) {
	// Create input with many consecutive duplicates
	input := make([]int, 10000)
	for i := range input {
		input[i] = i / 10 // Groups of 10 identical values
	}

	b.Run("DistinctUntilChanged_Collect", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = DistinctUntilChanged(FromSlice(input)).Collect()
		}
	})

	b.Run("DistinctUntilChanged_Count", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = sink(DistinctUntilChanged(FromSlice(input)))
		}
	})
}

// BenchmarkStep benchmarks the Step operation.
func BenchmarkStep(b *testing.B) {
	input := Range(1, 10001).Collect() // 10000 elements

	b.Run("Step_10_Collect", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = FromSlice(input).Step(10).Collect()
		}
	})

	b.Run("Step_10_Count", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = sink(FromSlice(input).Step(10))
		}
	})
}

// BenchmarkFlatten benchmarks the Flatten operation.
func BenchmarkFlatten(b *testing.B) {
	// Create 100 slices of 100 elements each
	input := make([][]int, 100)
	for i := range input {
		input[i] = make([]int, 100)
		for j := range input[i] {
			input[i][j] = i*100 + j
		}
	}

	b.Run("Flatten_Collect", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = Flatten(FromSlice(input)).Collect()
		}
	})

	b.Run("Flatten_Count", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = sink(Flatten(FromSlice(input)))
		}
	})
}

// BenchmarkZip3 benchmarks the Zip3 operation.
func BenchmarkZip3(b *testing.B) {
	input1 := Range(1, 1001).Collect()
	input2 := Range(1001, 2001).Collect()
	input3 := Range(2001, 3001).Collect()

	b.Run("Zip3_Collect", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = Zip3(FromSlice(input1), FromSlice(input2), FromSlice(input3)).Collect()
		}
	})

	b.Run("Zip3_Count", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = sink(Zip3(FromSlice(input1), FromSlice(input2), FromSlice(input3)))
		}
	})
}

// BenchmarkReduceByKey benchmarks the ReduceByKey operation.
func BenchmarkReduceByKey(b *testing.B) {
	// Create 10000 key-value pairs with 100 unique keys
	pairs := make([]Pair[int, int], 10000)
	for i := range pairs {
		pairs[i] = NewPair(i%100, i)
	}

	b.Run("ReduceByKey_10000_100keys", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			s := PairsOf(pairs...)
			_ = ReduceByKey(s, func(a, b int) int { return a + b })
		}
	})
}

// BenchmarkParallelFlatMap compares streaming vs chunked modes.
// Use these benchmarks to tune ChunkSize based on your workload.
func BenchmarkParallelFlatMap(b *testing.B) {
	// Simulate workload: 100 inputs, each producing sub-stream of 100 elements
	input := Range(0, 100).Collect()
	subStreamSize := 100

	b.Run("Streaming_100x100", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = ParallelFlatMap(FromSlice(input), func(_ int) Stream[int] {
				return Range(0, subStreamSize)
			}, WithConcurrency(4), WithOrdered(true)).Collect()
		}
	})

	b.Run("Chunked_100x100_Chunk10", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = ParallelFlatMap(FromSlice(input), func(_ int) Stream[int] {
				return Range(0, subStreamSize)
			}, WithConcurrency(4), WithOrdered(true), WithChunkSize(10)).Collect()
		}
	})

	b.Run("Chunked_100x100_Chunk25", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = ParallelFlatMap(FromSlice(input), func(_ int) Stream[int] {
				return Range(0, subStreamSize)
			}, WithConcurrency(4), WithOrdered(true), WithChunkSize(25)).Collect()
		}
	})

	b.Run("Chunked_100x100_Chunk50", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = ParallelFlatMap(FromSlice(input), func(_ int) Stream[int] {
				return Range(0, subStreamSize)
			}, WithConcurrency(4), WithOrdered(true), WithChunkSize(50)).Collect()
		}
	})

	// Large sub-streams scenario
	largeSubStreamSize := 1000

	b.Run("Streaming_50x1000", func(b *testing.B) {
		b.ReportAllocs()
		smallInput := Range(0, 50).Collect()
		for b.Loop() {
			_ = ParallelFlatMap(FromSlice(smallInput), func(_ int) Stream[int] {
				return Range(0, largeSubStreamSize)
			}, WithConcurrency(4), WithOrdered(true)).Collect()
		}
	})

	b.Run("Chunked_50x1000_Chunk10", func(b *testing.B) {
		b.ReportAllocs()
		smallInput := Range(0, 50).Collect()
		for b.Loop() {
			_ = ParallelFlatMap(FromSlice(smallInput), func(_ int) Stream[int] {
				return Range(0, largeSubStreamSize)
			}, WithConcurrency(4), WithOrdered(true), WithChunkSize(10)).Collect()
		}
	})
}
