# go-streams

[![Go Reference](https://pkg.go.dev/badge/github.com/coldsmirk/go-streams.svg)](https://pkg.go.dev/github.com/coldsmirk/go-streams)
[![Go Report Card](https://goreportcard.com/badge/github.com/coldsmirk/go-streams)](https://goreportcard.com/report/github.com/coldsmirk/go-streams)
[![Build Status](https://github.com/coldsmirk/go-streams/actions/workflows/test.yml/badge.svg)](https://github.com/coldsmirk/go-streams/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/coldsmirk/go-streams/branch/main/graph/badge.svg)](https://codecov.io/gh/coldsmirk/go-streams)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A lazy, type-safe stream processing library for Go 1.25+, built on `iter.Seq` and `iter.Seq2`.

## Features

- **Zero-allocation priority**: Lazy evaluation avoids intermediate slice allocations
- **Type-safe**: Full generics support, no `interface{}` or reflection
- **Idiomatic Go**: Supports both method chaining and function composition
- **Composable**: Every operation is independent and freely combinable
- **Standard library compatible**: Built on `iter.Seq`/`iter.Seq2` for seamless integration with `slices`, `maps` packages
- **Parallel processing**: ParallelMap/Filter with configurable concurrency and ordering
- **Context-aware**: Full support for context cancellation and timeouts
- **Error handling**: Result type for explicit error propagation in pipelines
- **IO support**: Built-in constructors for files, readers, CSV/TSV
- **Time-based operations**: Windowing, throttling, rate limiting, debouncing

## Quick API Index

- Constructors & Interop: Streams, Ranges, Generators → see Streams: Constructors and Interop
- Stream[T] (lazy): Filter, Map, Peek, Limit/Skip, Take/DropWhile, Step, TakeLast/DropLast, Intersperse
- Type‑changing: MapTo, FlatMap, Flatten, Zip/Zip3/ZipWithIndex, Distinct*, Window/Chunk, Interleave, Pairwise/Triples
- Specialized: MergeSorted*, Cartesian/Cross/Combinations/Permutations
- Terminals: Collect, Reduce/Fold, Count/First/Last/Find*, Any/All/NoneMatch, Min/Max, At/Nth, Single, IsEmpty
- Parallel: ParallelMap/Filter/FlatMap/Reduce/ForEach/Collect, Prefetch, options WithConcurrency/Ordered/BufferSize/ChunkSize
- Context‑Aware: WithContext/WithContext2, Generate/Iterate/Range/FromChannel/FromReaderLines Ctx variants, Collect/ForEach/Reduce Ctx variants, Parallel*Ctx
- Resource Management: Using (try-with-resources)
- IO: FromReaderLines/Scanner/String/Bytes/Runes, FromCSV/TSV/WithHeader (+Err), ToWriter/ToFile/ToCSV(+File)
- Time‑Based: WithTimestamp, Tumbling/Sliding/Session windows, Throttle/RateLimit/Debounce/Sample/Delay/Timeout, Interval/Timer
- Stream2: Keys/Values/ToPairs/Reduce/DistinctKeys/Values, MapKeys/Values/Pairs, ReduceByKey/GroupValues/ToMap2
- Joins: Inner/Left/Right/Full, LeftJoinWith/RightJoinWith, CoGroup, JoinBy/LeftJoinBy, Semi/Anti (and *By)
- Numeric/Stats: Sum/Average/Min/Max/MinMax/Product/RunningSum/Differences/etc, GetStatistics
- Collectors: ToSlice/Set, Grouping/Partitioning/ToMap, Mapping/Filtering/FlatMapping/Teeing, TopK/BottomK/Quantile/Histogram + helpers
- Result pipeline: Ok/Err, MapErrTo/FilterErr/FlatMapErr, CollectResults*, FilterOk/Errs, Unwrap*, TakeUntilErr, FromResults, TryCollect
- Optional: Some/None + Map/Filter/Zip, conversions
- Tuples: Pair/Triple/Quad, Unzip, helpers

## Table of Contents

- [Features](#features)
- [Quick API Index](#quick-api-index)
- [Requirements](#requirements)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Core Concepts](#core-concepts)
- [Examples](#examples)
- [Design Philosophy](#design-philosophy)
- [API Reference](#api-reference)
  - [Parameter Glossary](#parameter-glossary)
  - [Streams: Constructors and Interop](#streams-constructors-and-interop)
  - [Stream[T]: Intermediate Methods](#streamt-intermediate-lazy-methods)
  - [Free Transformations (type-changing)](#streamt-free-transformations-type-changing-lazy)
  - [Specialized Combinators and Merges](#specialized-combinators-and-merge-lazy-unless-noted)
  - [Terminal Operations](#terminal-operations-eager)
  - [Parallel Processing](#parallel-processing)
  - [Context-Aware APIs](#context-aware-apis)
  - [IO: Lines/CSV/TSV/Writers](#io-lines-bytes-csvtsv)
  - [Time-Based Operators](#time-based-operators)
  - [Stream2[K,V]](#stream2kv-key-value-streams)
  - [Joins](#join-family-on-stream2-and-on-streamt-with-keys)
  - [Numeric and Statistics](#numeric-and-statistics)
  - [Collectors](#collectors-composable-accumulators)
  - [Result[T] Pipeline](#resultt-pipeline-error-aware-streams)
  - [Optional[T]](#optionalt)
  - [Tuples and Helpers](#tuples-and-helpers)
  - [go-collections Integration](#go-collections-integration)
  - [Practical Guidance](#practical-guidance)
- [Contributing](#contributing)
- [License](#license)
- [Acknowledgments](#acknowledgments)

## Requirements

- Go 1.25 or later

## Installation

```bash
go get github.com/coldsmirk/go-streams
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/coldsmirk/go-streams"
)

func main() {
    // Basic filtering and mapping
    result := streams.Of(1, 2, 3, 4, 5).
        Filter(func(n int) bool { return n%2 == 0 }).
        Map(func(n int) int { return n * 2 }).
        Collect()

    fmt.Println(result) // [4 8]
}
```

## Core Concepts

### Stream Types

| Type | Description |
|------|-------------|
| `Stream[T]` | A lazy sequence of elements of type T |
| `Stream2[K, V]` | A lazy sequence of key-value pairs |
| `Optional[T]` | A value that may or may not exist |
| `Collector[T, A, R]` | Strategy for accumulating elements into a result |

### Constructors

```go
// From values
streams.Of(1, 2, 3, 4, 5)

// From slice
streams.FromSlice([]string{"a", "b", "c"})

// From iter.Seq (stdlib interop)
streams.From(slices.Values(mySlice))

// From map
streams.FromMap(map[string]int{"a": 1, "b": 2})

// From channel
streams.FromChannel(ch)

// Numeric ranges
streams.Range(1, 10)       // [1, 10)
streams.RangeClosed(1, 10) // [1, 10]

// Infinite streams (use with Limit or TakeWhile)
streams.Generate(func() int { return rand.Int() })
streams.Iterate(1, func(n int) int { return n * 2 })
streams.Cycle(1, 2, 3)
streams.RepeatForever("x")

// IO constructors
streams.FromReaderLines(reader)        // Stream lines from io.Reader
streams.FromScanner(scanner)           // Stream from bufio.Scanner
streams.FromFileLines("path.txt")      // Stream lines from file
streams.FromCSV(reader)                // Stream CSV records
streams.FromCSVWithHeader(reader)      // Stream CSV as maps
streams.FromTSV(reader)                // Stream TSV records
streams.FromStringLines("a\nb\nc")     // Stream lines from string
streams.FromRunes("hello")             // Stream runes
streams.FromBytes([]byte{1, 2, 3})     // Stream bytes

// Context-aware constructors
streams.GenerateCtx(ctx, supplier)     // Cancellable Generate
streams.FromChannelCtx(ctx, ch)        // Cancellable FromChannel
streams.Interval(ctx, interval)        // Emit integers at intervals
streams.Timer(ctx, duration, value)    // Emit single value after delay
```

### Intermediate Operations (Lazy)

```go
s := streams.Of(1, 2, 3, 4, 5)

s.Filter(pred)           // Keep elements matching predicate
s.Map(fn)                // Transform elements (same type)
s.Limit(n)               // Take first n elements
s.Skip(n)                // Skip first n elements
s.TakeWhile(pred)        // Take while predicate is true
s.DropWhile(pred)        // Drop while predicate is true
s.TakeLast(n)            // Take last n elements (buffered)
s.DropLast(n)            // Drop last n elements (buffered)
s.Step(n)                // Take every nth element (sampling)
s.Peek(action)           // Execute action on each element
s.Intersperse(sep)       // Insert separator between elements

// Type-changing transformations (free functions)
streams.MapTo(s, fn)     // Transform to different type
streams.FlatMap(s, fn)   // Map and flatten
streams.Distinct(s)      // Remove duplicates
streams.DistinctBy(s, keyFn)
streams.DistinctUntilChanged(s)    // Remove consecutive duplicates
streams.DistinctUntilChangedBy(s, eq) // With custom equality
streams.Zip(s1, s2)      // Combine two streams into Pairs
streams.Zip3(s1, s2, s3) // Combine three streams into Triples
streams.ZipWithIndex(s)  // Add indices
streams.Chunk(s, size)   // Split into fixed-size chunks
streams.Window(s, size)  // Sliding windows (step=1)
streams.WindowWithStep(s, size, step, allowPartial) // Configurable sliding windows
streams.Pairwise(s)      // Consecutive pairs: [(a,b), (b,c), ...]
streams.Triples(s)       // Consecutive triples
streams.Interleave(s1, s2)
streams.Flatten(s)       // Flatten Stream[[]T] to Stream[T]
streams.Scan(s, init, fn) // Running accumulation (generalized RunningSum)

// Specialized operations
streams.MergeSorted(s1, s2, cmp)    // Merge two sorted streams
streams.MergeSortedN(cmp, s1, s2, s3...) // Merge multiple sorted streams (pairwise)
streams.MergeSortedNHeap(cmp, s1, s2...) // Merge multiple sorted streams (heap-based, O(n log k))
streams.ZipLongest(s1, s2)          // Zip with Optionals for missing values
streams.ZipLongestWith(s1, s2, def1, def2) // Zip with default values
streams.Cartesian(s1, s2)           // Cartesian product → Pairs
streams.CartesianSelf(s)            // Self Cartesian product
streams.CrossProduct(s1, s2, s3...) // N-way Cartesian product → []T
streams.Combinations(s, k)          // k-combinations
streams.Permutations(s)             // All permutations

// Context-aware operations
streams.WithContext(ctx, s)         // Add context cancellation
streams.ThrottleCtx(ctx, s, interval)
streams.DelayCtx(ctx, s, duration)
streams.RateLimitCtx(ctx, s, n, per)

// Time-based operations
streams.Throttle(s, interval)       // Min time between elements
streams.Delay(s, duration)          // Delay each element
streams.RateLimit(s, n, per)        // Token bucket rate limiting
streams.Debounce(ctx, s, quiet)     // Emit after quiet period
streams.Sample(ctx, s, interval)    // Sample at intervals
streams.WithTimestamp(s)            // Add timestamps to elements
streams.TumblingTimeWindow(ctx, s, size)   // Fixed time windows
streams.SlidingTimeWindow(ctx, s, size, slide) // Overlapping time windows
streams.SessionWindow(ctx, s, gap)  // Session-based windows
streams.Timeout(ctx, s, timeout)    // Emit Result with timeout errors

// Eager operations (collect all elements into memory first)
s.Sorted(cmp)            // Sort elements ⚠️ eager
s.SortedStable(cmp)      // Stable sort ⚠️ eager
s.Reverse()              // Reverse order ⚠️ eager
streams.SortedBy(s, keyFn)       // Sort by key ⚠️ eager
streams.SortedStableBy(s, keyFn) // Stable sort by key ⚠️ eager

// Parallel operations
streams.ParallelMap(s, fn, opts...)    // Parallel map with options
streams.ParallelFilter(s, pred, opts...) // Parallel filter
streams.ParallelFlatMap(s, fn, opts...)  // Parallel flatMap
streams.ParallelReduce(s, identity, op)  // Parallel reduce
streams.Prefetch(s, n)                   // Prefetch n elements ahead

// Parallel options
streams.WithConcurrency(n)    // Set worker count
streams.WithOrdered(true)     // Preserve input order
streams.WithBufferSize(n)     // Set buffer size
streams.WithChunkSize(n)      // Set chunk size for memory-bounded ordered processing
```

### Terminal Operations (Eager)

```go
s := streams.Of(1, 2, 3, 4, 5)

s.Collect()              // []T
s.ForEach(action)        // Execute action on each
s.ForEachErr(action)     // Execute action with error handling (stops on first error)
s.Reduce(identity, fn)   // Combine elements
s.Count()                // Number of elements
s.First()                // Optional[T]
s.Last()                 // Optional[T]
s.FindFirst(pred)        // First matching element
s.AnyMatch(pred)         // Any match?
s.AllMatch(pred)         // All match?
s.NoneMatch(pred)        // None match?
s.Min(cmp)               // Minimum element
s.Max(cmp)               // Maximum element
s.IsEmpty()              // Is stream empty?
s.Seq()                  // Convert back to iter.Seq

// Free functions
streams.ToMap(s, keyFn, valFn)
streams.ToSet(s)
streams.GroupBy(s, keyFn)
streams.PartitionBy(s, pred)
streams.Joining(s, separator)
streams.Contains(s, target)

// Context-aware terminal operations
streams.CollectCtx(ctx, s)                    // Collect with cancellation
streams.ForEachCtx(ctx, s, action)            // ForEach with cancellation (action: func(T) or func(T) error)
streams.ReduceCtx(ctx, s, identity, reducer)  // Reduce with cancellation (reducer: func(T,T) T or func(T,T) (T, error))

// IO output
streams.ToWriter(s, writer, format)  // Write to io.Writer
streams.ToCSV(s, writer)             // Write as CSV
```

### Stream2 Operations (Key-Value Pairs)

```go
s2 := streams.FromMap(map[string]int{"a": 1, "b": 2})

// Transformations
s2.Filter(pred)          // Filter by (k, v) predicate
s2.MapKeys(fn)           // Transform keys
s2.MapValues(fn)         // Transform values
s2.Limit(n)              // Take first n pairs
s2.Skip(n)               // Skip first n pairs

// Type-changing transformations (free functions)
streams.MapKeysTo(s2, fn)    // Transform key type
streams.MapValuesTo(s2, fn)  // Transform value type
streams.MapPairs(s2, fn)     // Transform both types
streams.SwapKeyValue(s2)     // Swap keys and values
streams.DistinctKeys(s2)     // Keep first occurrence of each key
streams.DistinctValues(s2)   // Keep first occurrence of each value

// Terminal operations
s2.Keys()                // Stream[K] of keys
s2.Values()              // Stream[V] of values
s2.ToPairs()             // Stream[Pair[K, V]]
s2.CollectPairs()        // []Pair[K, V]
s2.ForEach(action)       // Execute action on each pair
s2.Count()               // Number of pairs
streams.ToMap2(s2)       // map[K]V (requires comparable K)
streams.ReduceByKey(s2, merge)    // Reduce values by key → map[K]V
streams.GroupValues(s2)           // Group values by key → map[K][]V
```

### Numeric Operations

```go
s := streams.Of(1, 2, 3, 4, 5)

streams.Sum(s)           // Sum all elements
streams.Average(s)       // Average (returns Optional)
streams.MinValue(s)      // Minimum (returns Optional)
streams.MaxValue(s)      // Maximum (returns Optional)
streams.MinMax(s)        // Both min and max
streams.Product(s)       // Multiply all elements
streams.GetStatistics(s) // Count, Sum, Min, Max, Average

// Transformations
streams.RunningSum(s)    // Cumulative sums
streams.Differences(s)   // Differences between consecutive elements
streams.Scale(s, factor) // Multiply by factor
streams.Offset(s, delta) // Add offset
streams.Clamp(s, min, max)
streams.Abs(s)           // Absolute values
streams.Positive(s)      // Filter positive values
streams.Negative(s)      // Filter negative values
```

### Collectors

```go
// Basic collectors
streams.CollectTo(s, streams.ToSliceCollector[int]())
streams.CollectTo(s, streams.ToSetCollector[int]())
streams.CollectTo(s, streams.JoiningCollector(", "))

// Aggregating collectors
streams.CollectTo(s, streams.CountingCollector[int]())
streams.CollectTo(s, streams.SummingCollector[int]())
streams.CollectTo(s, streams.AveragingCollector[int]())
streams.CollectTo(s, streams.MaxByCollector[int](cmp))
streams.CollectTo(s, streams.MinByCollector[int](cmp))

// Grouping collectors
streams.CollectTo(s, streams.GroupingByCollector(keyFn))
streams.CollectTo(s, streams.PartitioningByCollector(pred))
streams.CollectTo(s, streams.ToMapCollector(keyFn, valFn))

// Composite collectors
streams.MappingCollector(mapper, downstream)
streams.FilteringCollector(pred, downstream)
streams.FlatMappingCollector(mapper, downstream)
streams.TeeingCollector(c1, c2, merger)

// TopK and statistical collectors
streams.TopKCollector(k, less)           // Find k largest elements
streams.BottomKCollector(k, less)        // Find k smallest elements
streams.QuantileCollector(q, less)       // Compute quantile (0.0-1.0)
streams.FrequencyCollector[T]()          // Count occurrences → map[T]int
streams.HistogramCollector(keyFn)        // Group into buckets

// Convenience functions
streams.TopK(s, k, less)                 // []T - k largest
streams.BottomK(s, k, less)              // []T - k smallest
streams.Median(s, less)                  // Optional[T]
streams.Quantile(s, q, less)             // Optional[T]
streams.Percentile(s, p, less)           // Optional[T] (p in 0-100)
streams.Frequency(s)                     // map[T]int
streams.MostCommon(s, n)                 // []Pair[T, int] - n most common
```

### Optional

```go
opt := streams.Some(42)
opt := streams.None[int]()

opt.IsPresent()          // true/false
opt.IsEmpty()            // true/false
opt.Get()                // Value (panics if empty)
opt.GetOrElse(default)   // Value or default
opt.GetOrZero()          // Value or zero value
opt.IfPresent(action)    // Execute if present
opt.Filter(pred)         // Filter by predicate
opt.Map(fn)              // Transform value
opt.OrElse(other)        // This or other Optional
opt.ToSlice()            // []T (empty or single element)
opt.ToStream()           // Stream[T]

// Type-changing transformations
streams.OptionalMap(opt, fn)
streams.OptionalFlatMap(opt, fn)
streams.OptionalZip(opt1, opt2)
```

### Tuple Types

```go
// Pair
p := streams.NewPair(1, "hello")
p.First                  // 1
p.Second                 // "hello"
p.Swap()                 // Pair["hello", 1]
first, second := p.Unpack()

// Triple
t := streams.NewTriple(1, "hello", 3.14)
t.ToPair()               // Drop third element

// Quad
q := streams.NewQuad(1, "hello", 3.14, true)
q.ToTriple()             // Drop fourth element
```

### Result Type (Error Handling)

```go
// Creating Results
r := streams.Ok(42)                    // Success result
r := streams.Err[int](err)             // Error result

// Checking state
r.IsOk()                               // true if success
r.IsErr()                              // true if error
r.Value()                              // Get value (zero if error)
r.Error()                              // Get error (nil if success)
r.Unwrap()                             // Get value (panics if error)
r.UnwrapOr(default)                    // Get value or default
r.ToOptional()                         // Convert to Optional

// Error-aware stream operations
streams.MapErrTo(s, fn)                // Map with error return
streams.FilterErr(s, pred)             // Filter with error return
streams.FlatMapErr(s, fn)              // FlatMap with error return

// Working with Result streams
streams.CollectResults(s)              // ([]T, error) - collect until first error
streams.FilterOk(s)                    // Stream[T] - keep only Ok values
streams.FromResults(r1, r2, r3...)     // Create stream from Results
```

### Join Operations

```go
// Create key-value streams for joining
s1 := streams.PairsOf(streams.NewPair("a", 1), streams.NewPair("b", 2))
s2 := streams.PairsOf(streams.NewPair("a", "x"), streams.NewPair("c", "y"))

// SQL-style joins on Stream2
streams.InnerJoin(s1, s2)              // Only matching keys
streams.LeftJoin(s1, s2)               // All from left, matched from right
streams.RightJoin(s1, s2)              // All from right, matched from left
streams.FullJoin(s1, s2)               // All from both sides

// Joins with default values
streams.LeftJoinWith(s1, s2, defaultV2)
streams.RightJoinWith(s1, s2, defaultV1)

// Semi and Anti joins
streams.SemiJoin(s1, s2)               // Left keys that exist in right
streams.AntiJoin(s1, s2)               // Left keys that don't exist in right

// CoGroup - group all values by key
streams.CoGroup(s1, s2)                // Stream[CoGrouped[K, V1, V2]]

// Join on Stream[T] with key extractors
streams.JoinBy(s1, s2, keyFn1, keyFn2)
streams.LeftJoinBy(s1, s2, keyFn1, keyFn2)
streams.SemiJoinBy(s1, s2, keyFn1, keyFn2)
streams.AntiJoinBy(s1, s2, keyFn1, keyFn2)
```

## Examples

### Filter and Transform Users

```go
type User struct {
    Name   string
    Age    int
    Active bool
}

users := []User{
    {Name: "Alice", Age: 30, Active: true},
    {Name: "Bob", Age: 25, Active: false},
    {Name: "Charlie", Age: 35, Active: true},
}

// Get names of active users over 25
names := streams.MapTo(
    streams.FromSlice(users).
        Filter(func(u User) bool { return u.Active && u.Age > 25 }),
    func(u User) string { return u.Name },
).Collect()
// ["Alice", "Charlie"]
```

### Group and Count

```go
words := []string{"apple", "apricot", "banana", "blueberry", "cherry"}

// Group by first letter
grouped := streams.GroupBy(
    streams.FromSlice(words),
    func(s string) rune { return rune(s[0]) },
)
// {'a': ["apple", "apricot"], 'b': ["banana", "blueberry"], 'c': ["cherry"]}

// Count by first letter
counts := streams.CountBy(
    streams.FromSlice(words),
    func(s string) rune { return rune(s[0]) },
)
// {'a': 2, 'b': 2, 'c': 1}
```

### Working with Infinite Streams

```go
// First 10 Fibonacci numbers
fib := streams.Generate(func() func() int {
    a, b := 0, 1
    return func() int {
        a, b = b, a+b
        return a
    }
}()).Limit(10).Collect()
// [1, 1, 2, 3, 5, 8, 13, 21, 34, 55]

// Powers of 2
powers := streams.Iterate(1, func(n int) int { return n * 2 }).
    Limit(10).
    Collect()
// [1, 2, 4, 8, 16, 32, 64, 128, 256, 512]
```

### Parallel Processing

```go
// Parallel map with ordered output
result := streams.ParallelMap(
    streams.Range(1, 1000),
    func(n int) int {
        // CPU-intensive operation
        return heavyComputation(n)
    },
    streams.WithConcurrency(4),
    streams.WithOrdered(true),
).Collect()

// Ordered (chunked) mode to bound memory when preserving order
resultChunked := streams.ParallelMap(
    streams.Range(1, 100000),
    func(n int) int { return n * n },
    streams.WithConcurrency(8),
    streams.WithOrdered(true),
    streams.WithChunkSize(8*4), // e.g., 4x concurrency as a starting point
).Collect()

// Parallel filter
evens := streams.ParallelFilter(
    streams.Range(1, 10000),
    func(n int) bool { return n%2 == 0 },
).Collect()

// Ordered (chunked) filter
evensChunked := streams.ParallelFilter(
    streams.Range(1, 200000),
    func(n int) bool { return n%2 == 0 },
    streams.WithOrdered(true),
    streams.WithChunkSize(8*4),
).Collect()

// Parallel flatMap with chunked reordering for bounded memory
result := streams.ParallelFlatMap(
    streams.Range(1, 1000),
    func(n int) streams.Stream[int] {
        return streams.Of(n*2, n*2+1)
    },
    streams.WithConcurrency(4),
    streams.WithChunkSize(100), // Process 100 elements per chunk
).Collect()
```

> **Note**: When using `WithOrdered(true)` (the default), results are buffered to preserve order, which may increase memory usage. For `ParallelFlatMap` with ordered mode, each sub-stream is fully collected into memory before yielding.
>
> **Memory considerations for ParallelFlatMap:**
> - **Streaming mode (default)**: May buffer all out-of-order results globally. Memory = O(pending results × avg sub-stream size)
> - **Chunked mode (`WithChunkSize(n)`)**: Bounds memory to n sub-streams at a time. Memory = O(n × avg sub-stream size)
> - `WithChunkSize(1)` provides minimum memory but lowest parallelism utilization
> - **Tuning guide**: Balance chunk size based on average sub-stream length and available memory. Start with 2-4× concurrency level, then adjust based on profiling.
> - If individual sub-streams are very large, consider: `WithOrdered(false)`, splitting sub-streams with `Chunk()`, or using sequential `FlatMap`
>
> The same trade-off applies to `ParallelMap` and `ParallelFilter` in ordered mode. Use `WithChunkSize(n)` to bound memory when strict ordering is required and workload is skewed.

### Error Handling with Result

```go
// Process items that may fail
results := streams.MapErrTo(
    streams.Of("1", "abc", "3"),
    func(s string) (int, error) {
        return strconv.Atoi(s)
    },
).Collect()

// Collect until first error
values, err := streams.CollectResults(
    streams.FromResults(
        streams.Ok(1),
        streams.Ok(2),
        streams.Err[int](errors.New("failed")),
    ),
)
// values = [1, 2], err = "failed"
```

### CSV Processing

```go
// Read CSV and filter rows
file, _ := os.Open("data.csv")
defer file.Close()

activeUsers := streams.FromCSVWithHeader(file).
    Filter(func(row streams.CSVRecord) bool {
        return row["status"] == "active"
    }).Collect()
```

### Rate Limiting

```go
// Process at most 10 requests per second
streams.RateLimit(
    streams.FromChannel(requests),
    10,
    time.Second,
).ForEach(processRequest)
```

### TopK Analysis

```go
// Find top 5 most common words
words := streams.FromStringLines(text)
topWords := streams.MostCommon(
    streams.FlatMap(words, func(line string) streams.Stream[string] {
        return streams.FromSlice(strings.Fields(line))
    }),
    5,
)
// []Pair[string, int] - word and count
```

### Join Operations Example

```go
// Join users with their orders
users := streams.PairsOf(
    streams.NewPair("u1", User{Name: "Alice"}),
    streams.NewPair("u2", User{Name: "Bob"}),
)
orders := streams.PairsOf(
    streams.NewPair("u1", Order{Amount: 100}),
    streams.NewPair("u1", Order{Amount: 200}),
)

// Inner join: only users with orders
for r := range streams.InnerJoin(users, orders).Seq() {
    fmt.Printf("%s ordered $%d\n", r.Left.Name, r.Right.Amount)
}
```

### Using for-range (stdlib interop)

```go
// Convert to iter.Seq and use with for-range
for name := range streams.MapTo(
    streams.FromSlice(users).Filter(func(u User) bool { return u.Active }),
    func(u User) string { return u.Name },
).Seq() {
    fmt.Println(name)
}
```

### Statistics

```go
numbers := streams.Of(10, 20, 30, 40, 50)

stats := streams.GetStatistics(numbers)
if stats.IsPresent() {
    s := stats.Get()
    fmt.Printf("Count: %d\n", s.Count)   // 5
    fmt.Printf("Sum: %d\n", s.Sum)       // 150
    fmt.Printf("Min: %d\n", s.Min)       // 10
    fmt.Printf("Max: %d\n", s.Max)       // 50
    fmt.Printf("Avg: %.1f\n", s.Average) // 30.0
}
```

## Design Philosophy

### Method Chaining vs Free Functions

- **Methods** are used for operations that don't change the element type (e.g., `Filter`, `Map`, `Limit`)
- **Free functions** are used for operations that change types (e.g., `MapTo`, `FlatMap`, `Zip`) due to Go generics limitations

### Why Free Functions for Some Operations?

Go's type system doesn't allow methods to return types with different type parameters than the receiver. For example:

```go
// This is NOT allowed in Go:
// func (s Stream[T]) MapTo(fn func(T) U) Stream[U]

// So we use a free function instead:
func MapTo[T, U any](s Stream[T], fn func(T) U) Stream[U]
```

### Lazy Evaluation

All intermediate operations are lazy - they don't execute until a terminal operation is called:

```go
// Nothing happens yet - just building the pipeline
pipeline := streams.Range(1, 1000000).
    Filter(func(n int) bool { return n%2 == 0 }).
    Map(func(n int) int { return n * 2 }).
    Limit(5)

// Now it executes, but only processes 10 elements (not 1 million)
result := pipeline.Collect() // [4, 8, 12, 16, 20]
```

### Boundary Conventions

Understanding how operations behave with edge-case inputs:

| Operation | n ≤ 0 Behavior | Notes |
|-----------|----------------|-------|
| `Limit(n)` | Returns empty stream | `Limit(0)` → `[]` |
| `Skip(n)` | Returns original stream | `Skip(0)` → all elements |
| `TakeLast(n)` | Returns empty stream | `TakeLast(0)` → `[]` |
| `DropLast(n)` | Returns original stream | `DropLast(0)` → all elements |
| `Step(n)` | Returns original stream | `Step(1)` or `Step(0)` → all elements |
| `Chunk(s, n)` | Returns empty stream | Invalid chunk size |
| `Window(s, n)` | Returns empty stream | Invalid window size |
| `WindowWithStep(s, size, step, _)` | Returns empty stream | If size ≤ 0 or step ≤ 0 |

**Window vs Chunk behavior:**
- `Window(s, 3)` with `[1,2,3,4,5]` → `[[1,2,3], [2,3,4], [3,4,5]]` (overlapping)
- `Chunk(s, 3)` with `[1,2,3,4,5]` → `[[1,2,3], [4,5]]` (non-overlapping)
- `WindowWithStep(s, 3, 2, false)` with `[1,2,3,4,5]` → `[[1,2,3], [3,4,5]]` (step=2)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Inspired by Java Stream API and Rust Iterator
- Built on Go 1.23's `iter.Seq` and `iter.Seq2`

## API Reference

This section documents all public APIs in one place. It complements the quick tour above with precise behavior, laziness/ eagerness notes, complexity hints, ordering, and cancellation/parallel considerations.

Conventions used below:
- Lazy: operation produces output on demand; memory is bounded by the operator, not by total input
- Eager: operation collects the whole input first (O(n) memory)
- Ordering: whether original order is preserved
- Cancellation: whether a `context.Context` can short‑circuit production/consumption

### Parameter Glossary

- `fn`: transformation function applied to each element.
- `pred`: predicate function returning `bool`.
- `cmp(a, b T) int`: comparison; negative if a<b, zero if a==b, positive if a>b.
- `keyFn`: extracts a comparable key.
- `less(a, b T) bool`: strict weak ordering for TopK/BottomK; returns true if a<b.
- `identity`: identity (initial) value for reduce/fold.
- `op` / `merge`: associative binary operation (e.g., sum, min, custom combine).
- `size`, `step`: window/chunk sizes and slide step.
- `n`: count or concurrency depending on context (see function docs).
- `ctx`: context for cancellation/timeouts.
- `interval`, `duration`, `per`, `gap`, `quiet`: time durations for time-based ops.
- `defaultT`, `defaultU`: default values for missing pairs in zip-longest with defaults.
- `K`, `V`, `T`, `U`, `A`, `R`: generic type parameters.
- Parallel options: `WithConcurrency`, `WithOrdered`, `WithBufferSize`, `WithChunkSize`.

### Streams: Constructors and Interop

```go
// Interop with stdlib iter
func From[T any](seq iter.Seq[T]) Stream[T]
func (s Stream[T]) Seq() iter.Seq[T]

// Values, slices, maps, channels
func Of[T any](values ...T) Stream[T]
func FromSlice[T any](s []T) Stream[T]
func FromMap[K comparable, V any](m map[K]V) Stream2[K,V]
func FromChannel[T any](ch <-chan T) Stream[T]

// Generators and ranges
func Generate[T any](supplier func() T) Stream[T]           // infinite, use Limit/TakeWhile
func Iterate[T any](seed T, f func(T) T) Stream[T]           // infinite
func Range(start, end int) Stream[int]                       // [start, end)
func RangeClosed(start, end int) Stream[int]                 // [start, end]

// Composition and helpers
func Concat[T any](streams ...Stream[T]) Stream[T]
func Using[T interface{ Close() error }, R any](resource T, fn func(T) R) R // try-with-resources
func Empty[T any]() Stream[T]
func Repeat[T any](value T, n int) Stream[T]
func RepeatForever[T any](value T) Stream[T]                  // infinite
func Cycle[T any](values ...T) Stream[T]                      // infinite if values not empty
```

Notes:
- All constructors are lazy except those explicitly returning collected results.
- `FromMap` returns `Stream2[K,V]` for key/value workflows.
- Infinite sources must be bounded with `Limit`/`TakeWhile` or consumed carefully.

Examples:
```go
// Of / FromSlice / Seq
nums := streams.Of(1,2,3).Collect()                 // [1 2 3]
names := streams.FromSlice([]string{"a","b"}).Seq() // iter.Seq[string]

// Range / RangeClosed
r1 := streams.Range(3, 6).Collect()        // [3 4 5]
r2 := streams.RangeClosed(3, 6).Collect()  // [3 4 5 6]

// Iterate
pow2 := streams.Iterate(1, func(n int) int { return n*2 }).Limit(5).Collect() // [1 2 4 8 16]

// Concat
joined := streams.Concat(streams.Of(1,2), streams.Of(3)).Collect() // [1 2 3]

// FromMap -> Stream2
kv := streams.FromMap(map[string]int{"a":1,"b":2}).CollectPairs()  // []Pair[string,int]

// FromChannel
ch := make(chan int, 2); ch<-10; ch<-20; close(ch)
fromCh := streams.FromChannel(ch).Collect() // [10 20]
```

### Stream[T]: Intermediate (Lazy) Methods

```go
// Filtering / mapping / tapping
func (s Stream[T]) Filter(pred func(T) bool) Stream[T]
func (s Stream[T]) Map(fn func(T) T) Stream[T]
func (s Stream[T]) Peek(action func(T)) Stream[T]

// Slicing
func (s Stream[T]) Limit(n int) Stream[T]                     // n<=0 -> empty
func (s Stream[T]) Skip(n int) Stream[T]                      // n<=0 -> original
func (s Stream[T]) Step(n int) Stream[T]                      // n<=1 -> original
func (s Stream[T]) TakeWhile(pred func(T) bool) Stream[T]
func (s Stream[T]) DropWhile(pred func(T) bool) Stream[T]
func (s Stream[T]) TakeLast(n int) Stream[T]                  // buffered; yields only at end
func (s Stream[T]) DropLast(n int) Stream[T]                  // buffered; yields while reading

// Ordering
func (s Stream[T]) Sorted(cmp func(a,b T) int) Stream[T]      // eager
func (s Stream[T]) SortedStable(cmp func(a,b T) int) Stream[T]// eager
func (s Stream[T]) Reverse() Stream[T]                        // eager
func (s Stream[T]) Intersperse(sep T) Stream[T]
```

Notes:
- `TakeLast`/`DropLast` use ring buffers of size n; O(n) memory, one pass.
- `Sorted*`/`Reverse` are eager by design (O(n) memory).

Examples:
```go
src := streams.Of(1,2,3,4,5)

// Filter / Map / Peek
evensTimes10 := src.Filter(func(x int) bool { return x%2==0 }).
  Map(func(x int) int { return x*10 }).
  Peek(func(x int){ fmt.Println("peek", x) }).
  Collect() // prints and returns [20 40]

// Slicing
first3 := src.Limit(3).Collect()         // [1 2 3]
skip2 := src.Skip(2).Collect()           // [3 4 5]
every2 := src.Step(2).Collect()          // [1 3 5]
upto3  := src.TakeWhile(func(x int) bool { return x<=3 }).Collect() // [1 2 3]
drop3  := src.DropWhile(func(x int) bool { return x<=3 }).Collect() // [4 5]

// Tail ops (buffered)
tail2 := src.TakeLast(2).Collect()       // [4 5]
dropLast2 := src.DropLast(2).Collect()   // [1 2 3]

// Sorting (eager)
desc := src.Sorted(func(a,b int) int { return b-a }).Collect() // [5 4 3 2 1]
stable := src.SortedStable(func(a,b int) int { return a-b }).Collect()

// Intersperse
withDots := streams.Of("a","b","c").Intersperse(".").Collect() // ["a" "." "b" "." "c"]
```

### Stream[T]: Free Transformations (Type‑Changing, Lazy)

```go
// Map / flatMap including iter.Seq interop
func MapTo[T,U any](s Stream[T], fn func(T) U) Stream[U]
func FlatMap[T,U any](s Stream[T], fn func(T) Stream[U]) Stream[U]
func FlatMapSeq[T,U any](s Stream[T], fn func(T) iter.Seq[U]) Stream[U]
func Flatten[T any](s Stream[[]T]) Stream[T]
func FlattenSeq[T any](s Stream[iter.Seq[T]]) Stream[T]

// Zipping and indexing
func Zip[T,U any](s1 Stream[T], s2 Stream[U]) Stream[Pair[T,U]]
func Zip3[A,B,C any](s1 Stream[A], s2 Stream[B], s3 Stream[C]) Stream[Triple[A,B,C]]
func ZipWithIndex[T any](s Stream[T]) Stream2[int,T]
func ZipLongest[T,U any](s1 Stream[T], s2 Stream[U]) Stream[Pair[Optional[T], Optional[U]]]
func ZipLongestWith[T,U any](s1 Stream[T], s2 Stream[U], defT T, defU U) Stream[Pair[T,U]]

// Distinct / dedupe
func Distinct[T comparable](s Stream[T]) Stream[T]
func DistinctBy[T any, K comparable](s Stream[T], keyFn func(T) K) Stream[T]
func DistinctUntilChanged[T comparable](s Stream[T]) Stream[T]
func DistinctUntilChangedBy[T any](s Stream[T], eq func(a,b T) bool) Stream[T]

// Windows and chunks
func Window[T any](s Stream[T], size int) Stream[[]T]         // sliding, step=1
func WindowWithStep[T any](s Stream[T], size, step int, allowPartial bool) Stream[[]T]
func Chunk[T any](s Stream[T], size int) Stream[[]T]          // non-overlapping

// Interleave and neighbors
func Interleave[T any](s1, s2 Stream[T]) Stream[T]
func Pairwise[T any](s Stream[T]) Stream[Pair[T,T]]
func Triples[T any](s Stream[T]) Stream[Triple[T,T,T]]

// Sorting by key (eager)
func SortedBy[T any, K cmp.Ordered](s Stream[T], keyFn func(T) K) Stream[T]            // eager
func SortedStableBy[T any, K cmp.Ordered](s Stream[T], keyFn func(T) K) Stream[T]      // eager
```

Behavior notes:
- `Window` copies each yielded window; safe to retain; memory O(size).
- `WindowWithStep`: step<size → overlapping; step==size → chunks; step>size → gaps. Optional trailing partial window.
- `Zip` ends with the shorter input; `ZipLongest*` continues until both end.

Examples:
```go
// MapTo / FlatMap
words := streams.FromSlice([]string{"go streams"})
wordLens := streams.MapTo(words, func(s string) int { return len(s) }).Collect() // [10]
chars := streams.FlatMap(words, func(s string) streams.Stream[rune] { return streams.FromRunes(s) }).Collect()

// Distinct / DistinctBy / DistinctUntilChanged
distinct := streams.Distinct(streams.Of(1,1,2,2,3)).Collect()                 // [1 2 3]
distinctByLen := streams.DistinctBy(streams.Of("a","b","aa"), func(s string) int { return len(s) }).Collect() // ["a" "aa"]
noStutter := streams.DistinctUntilChanged(streams.Of(1,1,2,1,1)).Collect()    // [1 2 1]

// Zip / Zip3 / ZipWithIndex
z := streams.Zip(streams.Of("a","b"), streams.Of(1,2,3)).Collect() // [("a",1) ("b",2)]
z3 := streams.Zip3(streams.Of(1), streams.Of(2), streams.Of(3)).Collect()
zidx := streams.ZipWithIndex(streams.Of("x","y")).ToPairs().Collect() // [(0,"x") (1,"y")]

// Windows and chunks
w := streams.Window(streams.Of(1,2,3,4), 3).Collect()                 // [[1 2 3] [2 3 4]]
ws := streams.WindowWithStep(streams.Of(1,2,3,4,5), 3, 2, false).Collect() // [[1 2 3] [3 4 5]]
chunks := streams.Chunk(streams.Of(1,2,3,4,5), 2).Collect()           // [[1 2] [3 4] [5]]

// Flatten / FlattenSeq
flat := streams.Flatten(streams.Of([]int{1,2}, []int{3})).Collect()   // [1 2 3]

// Interleave / Pairwise / Triples
inter := streams.Interleave(streams.Of(1,3,5), streams.Of(2,4,6)).Collect() // [1 2 3 4 5 6]
pairs := streams.Pairwise(streams.Of(1,2,3)).Collect()               // [(1,2) (2,3)]
tri   := streams.Triples(streams.Of(1,2,3,4)).Collect()              // [(1,2,3) (2,3,4)]
```

### Specialized Combinators and Merge (Lazy unless noted)

```go
// Sorted merge
func MergeSorted[T any](s1, s2 Stream[T], cmp func(a,b T) int) Stream[T]
func MergeSortedN[T any](cmp func(a,b T) int, streams ...Stream[T]) Stream[T]          // pairwise, O(n*k)
func MergeSortedNHeap[T any](cmp func(a,b T) int, streams ...Stream[T]) Stream[T]      // O(n log k)

// Products and combinatorics (collects)
func Cartesian[T,U any](s1 Stream[T], s2 Stream[U]) Stream[Pair[T,U]]                  // collects s2
func CartesianSelf[T any](s Stream[T]) Stream[Pair[T,T]]                               // collects s
func CrossProduct[T any](streams ...Stream[T]) Stream[[]T]                             // collects all
func Combinations[T any](s Stream[T], k int) Stream[[]T]                               // collects s
func Permutations[T any](s Stream[T]) Stream[[]T]                                      // collects s
```

Notes:
- Product/combination families collect entire inputs needed for recombination; consider input sizes.

Examples:
```go
// MergeSorted*
ms := streams.MergeSorted(streams.Of(1,3,5), streams.Of(2,4,6), func(a,b int) int { return a-b }).Collect() // [1 2 3 4 5 6]

// ZipLongest* defaults
zl := streams.ZipLongest(streams.Of(1), streams.Of(10,20)).Collect() // [(Some(1),Some(10)) (None,Some(20))]
zld := streams.ZipLongestWith(streams.Of(1), streams.Of(10,20), 0, 0).Collect() // [(1,10) (0,20)]

// Cartesian / Cross / Combinatorics
cart := streams.Cartesian(streams.Of(1,2), streams.Of("a")).Collect() // [(1,"a") (2,"a")]
self := streams.CartesianSelf(streams.Of(1,2)).Collect()               // [(1,1) (1,2) (2,1) (2,2)]
cross := streams.CrossProduct(streams.Of(1,2), streams.Of(3), streams.Of(4,5)).Collect()
comb := streams.Combinations(streams.Of(1,2,3), 2).Collect()          // [[1 2] [1 3] [2 3]]
perm := streams.Permutations(streams.Of(1,2,3)).Collect()             // [[1 2 3] ...]
```

### Terminal Operations (Eager)

```go
// Consumption
func (s Stream[T]) ForEach(action func(T))
func (s Stream[T]) ForEachIndexed(action func(int, T))
func (s Stream[T]) ForEachErr(action func(T) error) error
func (s Stream[T]) ForEachIndexedErr(action func(int, T) error) error
func (s Stream[T]) Collect() []T
func (s Stream[T]) Reduce(identity T, fn func(T,T) T) T
func (s Stream[T]) ReduceOptional(fn func(T,T) T) Optional[T]
func (s Stream[T]) Fold(identity T, fn func(T,T) T) T
func FoldTo[T,R any](s Stream[T], identity R, fn func(R,T) R) R

// Queries
func (s Stream[T]) Count() int
func (s Stream[T]) First() Optional[T]
func (s Stream[T]) Last() Optional[T]
func (s Stream[T]) FindFirst(pred func(T) bool) Optional[T]
func (s Stream[T]) FindLast(pred func(T) bool) Optional[T]
func (s Stream[T]) AnyMatch(pred func(T) bool) bool
func (s Stream[T]) AllMatch(pred func(T) bool) bool
func (s Stream[T]) NoneMatch(pred func(T) bool) bool
func (s Stream[T]) Min(cmp func(T,T) int) Optional[T]
func (s Stream[T]) Max(cmp func(T,T) int) Optional[T]
func Contains[T comparable](s Stream[T], target T) bool
func (s Stream[T]) At(index int) Optional[T]
func (s Stream[T]) Nth(index int) Optional[T]
func (s Stream[T]) Single() Optional[T]
func (s Stream[T]) IsEmpty() bool
func (s Stream[T]) IsNotEmpty() bool

// Collect into maps/sets and groupings
func ToMap[T any, K comparable, V any](s Stream[T], keyFn func(T) K, valFn func(T) V) map[K]V
func ToSet[T comparable](s Stream[T]) map[T]struct{}
func GroupBy[T any, K comparable](s Stream[T], keyFn func(T) K) map[K][]T
func GroupByTo[T any, K comparable, V any](s Stream[T], keyFn func(T) K, valFn func(T) V) map[K][]V
func PartitionBy[T any](s Stream[T], pred func(T) bool) ([]T, []T)
func Joining(s Stream[string], sep string) string
func JoiningWithPrefixSuffix(s Stream[string], sep, prefix, suffix string) string
func Associate[T any, K comparable, V any](s Stream[T], fn func(T) (K,V)) map[K]V
func AssociateBy[T any, K comparable](s Stream[T], keyFn func(T) K) map[K]T
func IndexBy[T any, K comparable](s Stream[T], keyFn func(T) K) map[K]T
func CountBy[T any, K comparable](s Stream[T], keyFn func(T) K) map[K]int
func Frequencies[T comparable](s Stream[T]) map[T]int
```

Notes:
- All terminal operations consume the stream. Reuse requires reconstructing the stream.

Examples:
```go
s := streams.Of(1,2,3,4,5)

// ForEach / ForEachIndexed
s.ForEach(func(v int){ fmt.Println(v) })
s.ForEachIndexed(func(i,v int){ fmt.Println(i, v) })

// ForEachErr / ForEachIndexedErr (error-aware iteration)
err := s.ForEachErr(func(v int) error {
    if v == 3 { return errors.New("error at 3") }
    fmt.Println(v)
    return nil
}) // Stops at first error

err = s.ForEachIndexedErr(func(i, v int) error {
    return processItem(i, v) // Process with error handling
})

// Reduce / ReduceOptional / Fold / FoldTo
sum := s.Reduce(0, func(a,b int) int { return a+b }) // 15
optSum := s.ReduceOptional(func(a,b int) int { return a+b }) // Some(15)
foldToLen := streams.FoldTo(streams.Of("a","bb"), 0, func(acc int, v string) int { return acc+len(v) }) // 3

// Queries
n := s.Count()                        // 5
first := s.First()                    // Some(1)
last := s.Last()                      // Some(5)
f2 := s.FindFirst(func(v int) bool { return v>2 }) // Some(3)
anyEven := s.AnyMatch(func(v int) bool { return v%2==0 }) // true
min := s.Min(func(a,b int) int { return a-b }).Get() // 1
at2 := s.At(2)                        // Some(3)
single := streams.Of(42).Single()     // Some(42)

// Collectors-like helpers
mp := streams.ToMap(streams.Of("a","bb"), func(s string) int { return len(s) }, func(s string) string { return s }) // map[1:"a",2:"bb"]
set := streams.ToSet(streams.Of(1,2,2)) // map[1:{} 2:{}]
g := streams.GroupBy(streams.Of("a","aa"), func(s string) int { return len(s) }) // map[1:["a"] 2:["aa"]]
joined := streams.Joining(streams.Of("a","b"), ",") // "a,b"
```

### Parallel Processing

Configuration:
```go
type ParallelOption func(*ParallelConfig)
func WithConcurrency(n int) ParallelOption             // default: GOMAXPROCS
func WithOrdered(ordered bool) ParallelOption          // default: true
func WithBufferSize(size int) ParallelOption           // default: 2*GOMAXPROCS
func WithChunkSize(size int) ParallelOption            // default: 0 (disabled)
```

Operators:
```go
func ParallelMap[T,U any](s Stream[T], fn func(T) U, opts ...ParallelOption) Stream[U]
func ParallelFilter[T any](s Stream[T], pred func(T) bool, opts ...ParallelOption) Stream[T]
func ParallelFlatMap[T,U any](s Stream[T], fn func(T) Stream[U], opts ...ParallelOption) Stream[U]
func ParallelMapCtx[T,U any](ctx context.Context, s Stream[T], fn func(context.Context, T) U, opts ...ParallelOption) Stream[U]
func ParallelFilterCtx[T any](ctx context.Context, s Stream[T], pred func(context.Context, T) bool, opts ...ParallelOption) Stream[T]
func ParallelFlatMapCtx[T,U any](ctx context.Context, s Stream[T], fn func(context.Context, T) Stream[U], opts ...ParallelOption) Stream[U]
func Prefetch[T any](s Stream[T], n int) Stream[T]                   // decouple producer/consumer

// Terminals
func ParallelForEach[T any](s Stream[T], action func(T), opts ...ParallelOption)
func ParallelForEachCtx[T any](ctx context.Context, s Stream[T], action func(context.Context, T), opts ...ParallelOption) error
func ParallelReduce[T any](s Stream[T], identity T, op func(T,T) T, opts ...ParallelOption) T
func ParallelCollect[T any](s Stream[T], opts ...ParallelOption) []T // order not guaranteed
```

Behavior and tuning:
- Ordered vs unordered:
  - Ordered (default) preserves input order; out‑of‑order results are buffered until they can be yielded.
  - Unordered (`WithOrdered(false)`) yields ASAP; no reordering buffer.
- ParallelFlatMap ordered mode:
  - Sub‑streams are collected to preserve order (bounded per sub‑stream, not globally).
  - Streaming mode (default): may buffer many out‑of‑order sub‑results; use when sub‑streams are small/medium.
  - Chunked reordering (`WithChunkSize(n)`): processes inputs in chunks of size n with a semaphore; bounds memory to O(n × avg sub‑stream size). `n=1` minimizes memory but lowers utilization.
- Early termination: downstream stop triggers cooperative cancellation and draining; goroutines are not leaked.
- Start tuning with `WithConcurrency(GOMAXPROCS)` and `WithChunkSize(2-4× concurrency)` for ordered flatMap, then profile.

Examples:
```go
// ParallelMap (ordered by default)
res := streams.ParallelMap(streams.Range(1,8), func(x int) int { return x*x }).Collect()

// Unordered for throughput
res2 := streams.ParallelMap(streams.Range(1,8), func(x int) int { return x*x }, streams.WithOrdered(false)).Collect()

// ParallelFilter
evens := streams.ParallelFilter(streams.Range(1,10), func(x int) bool { return x%2==0 }).Collect()

// ParallelFlatMap with chunked reordering
pfm := streams.ParallelFlatMap(
  streams.Range(1,6),
  func(n int) streams.Stream[int] { return streams.Of(n, n) }, // duplicate
  streams.WithConcurrency(4),
  streams.WithChunkSize(3),
).Collect()

// Prefetch to overlap producer/consumer
pref := streams.Prefetch(streams.Range(1,5), 2).Collect()

// ParallelReduce
sum := streams.ParallelReduce(streams.Range(1,1000), 0, func(a,b int) int { return a+b })
```

### Context‑Aware APIs

Wrappers and constructors:
```go
func WithContext[T any](ctx context.Context, s Stream[T]) Stream[T]
func WithContext2[K,V any](ctx context.Context, s Stream2[K,V]) Stream2[K,V]
func GenerateCtx[T any](ctx context.Context, supplier func() T) Stream[T]
func IterateCtx[T any](ctx context.Context, seed T, fn func(T) T) Stream[T]
func RangeCtx(ctx context.Context, start, end int) Stream[int]
func FromChannelCtx[T any](ctx context.Context, ch <-chan T) Stream[T]
func FromReaderLinesCtx(ctx context.Context, r io.Reader) Stream[string]
```

Intermediate and terminals with ctx:
```go
func FilterCtx[T any](ctx context.Context, s Stream[T], pred func(T) bool) Stream[T]
func MapCtx[T any](ctx context.Context, s Stream[T], fn func(T) T) Stream[T]
func MapToCtx[T,U any](ctx context.Context, s Stream[T], fn func(T) U) Stream[U]

func CollectCtx[T any](ctx context.Context, s Stream[T]) ([]T, error)
func ForEachCtx[T any, A ~func(T) | ~func(T) error](ctx context.Context, s Stream[T], action A) error
func ReduceCtx[T any, F ~func(T, T) T | ~func(T, T) (T, error)](ctx context.Context, s Stream[T], identity T, fn F) (T, error)
func FindFirstCtx[T any](ctx context.Context, s Stream[T], pred func(T) bool) (Optional[T], error)
func AnyMatchCtx[T any](ctx context.Context, s Stream[T], pred func(T) bool) (bool, error)
func AllMatchCtx[T any](ctx context.Context, s Stream[T], pred func(T) bool) (bool, error)
func CountCtx[T any](ctx context.Context, s Stream[T]) (int, error)
```

Notes:
- On cancellation, ctx variants return the partial result plus `ctx.Err()` where applicable.
- `WithContext*` stops emission promptly when `ctx.Done()` fires.
- `ForEachCtx` action can be `func(T)` or `func(T) error` - stops on first error
- `ReduceCtx` reducer can be `func(T, T) T` or `func(T, T) (T, error)` - stops on first error

Examples:
```go
ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
defer cancel()

// Cancellable source
_ = streams.FromChannelCtx(ctx, make(chan int)) // stops on ctx.Done()

// Cancellable map/filter and terminals
xs, err := streams.CollectCtx(ctx, streams.Range(1,1_000_000))
_ = err // context deadline exceeded (if it fired)

// ForEachCtx with error handling
err = streams.ForEachCtx(ctx, s, func(v int) error {
    return process(v) // stops on first error
})

// ReduceCtx with error handling
sum, err := streams.ReduceCtx(ctx, s, 0, func(a, b int) (int, error) {
    if b == 0 { return a, errors.New("division by zero") }
    return a / b, nil
})
```

### IO: Lines, Bytes, CSV/TSV

Constructors:
```go
// Lines and text
func FromReaderLines(r io.Reader) Stream[string]
func FromScanner(scanner *bufio.Scanner) Stream[string]
func FromScannerErr(scanner *bufio.Scanner) Stream[Result[string]]
func FromReaderLinesErr(r io.Reader) Stream[Result[string]]

// Files
type FileLineStream struct { Stream[string] /* ... */ }
func FromFileLines(path string) (*FileLineStream, error)       // remember to Close()
func MustFromFileLines(path string) *FileLineStream            // panics on open error
func (f *FileLineStream) Close() error

// Bytes and runes
func FromStringLines(s string) Stream[string]
func FromBytes(data []byte) Stream[byte]
func FromRunes(s string) Stream[rune]

// CSV / TSV
func FromCSV(r io.Reader) Stream[[]string]
func FromCSVErr(r io.Reader) Stream[Result[[]string]]
func FromCSVFile(path string) (*CSVStream, error)               // remember to Close()
func FromTSV(r io.Reader) Stream[[]string]
func FromTSVErr(r io.Reader) Stream[Result[[]string]]
func FromCSVWithHeader(r io.Reader) Stream[CSVRecord]
func FromCSVWithHeaderErr(r io.Reader) Stream[Result[CSVRecord]]

// Writers
func ToWriter[T any](s Stream[T], w io.Writer, format func(T) string) error
func ToFile[T any](s Stream[T], path string, format func(T) string) error
func ToCSV(s Stream[[]string], w io.Writer) error
func ToCSVFile(s Stream[[]string], path string) error
```

Types:
```go
type CSVStream struct { Stream[[]string] /* ... */ }
func (c *CSVStream) Close() error

type CSVRecord map[string]string
func (r CSVRecord) Get(field string) string
func (r CSVRecord) GetOr(field, defaultVal string) string
```

Notes:
- Non‑Err variants stop on the first parse error (fail‑fast).
- Err variants emit `Result[T]` so pipelines can handle or skip bad rows.
- `FromFileLines`/`FromCSVFile` return closers; always call `Close()` (use `defer`).

Examples:
```go
// Lines from string
lines := streams.FromStringLines("a\nb\nc").Collect() // ["a" "b" "c"]

// CSV without header
r := strings.NewReader("x,y\n1,2\n3,4\n")
rows := streams.FromCSV(r).Collect() // [["x","y"],["1","2"],["3","4"]]

// CSV with header (records as map)
r2 := strings.NewReader("k,v\nA,1\nB,2\n")
recs := streams.FromCSVWithHeader(r2).Collect() // []CSVRecord
firstV := recs[0].Get("v")                      // "1"

// Writers
var sb strings.Builder
_ = streams.ToWriter(streams.Of(1,2,3), &sb, func(v int) string { return strconv.Itoa(v) })
```

### Time‑Based Operators

```go
// Timestamp decoration
type TimestampedValue[T any] struct{ Value T; Timestamp time.Time }
func WithTimestamp[T any](s Stream[T]) Stream[TimestampedValue[T]]

// Windows
func TumblingTimeWindow[T any](ctx context.Context, s Stream[T], windowSize time.Duration) Stream[[]T]
func SlidingTimeWindow[T any](ctx context.Context, s Stream[T], windowSize, slide time.Duration) Stream[[]T]
func SessionWindow[T any](ctx context.Context, s Stream[T], gap time.Duration) Stream[[]T]

// Rates and delays
func Throttle[T any](s Stream[T], interval time.Duration) Stream[T]
func ThrottleCtx[T any](ctx context.Context, s Stream[T], interval time.Duration) Stream[T]
func RateLimit[T any](s Stream[T], n int, per time.Duration) Stream[T]              // token bucket
func RateLimitCtx[T any](ctx context.Context, s Stream[T], n int, per time.Duration) Stream[T]
func Debounce[T any](ctx context.Context, s Stream[T], quiet time.Duration) Stream[T]
func Sample[T any](ctx context.Context, s Stream[T], interval time.Duration) Stream[T]
func Delay[T any](s Stream[T], d time.Duration) Stream[T]
func DelayCtx[T any](ctx context.Context, s Stream[T], d time.Duration) Stream[T]
func Timeout[T any](ctx context.Context, s Stream[T], d time.Duration) Stream[Result[T]]

// Interval and one‑shot
func Interval(ctx context.Context, interval time.Duration) Stream[int]              // 0,1,2,...
func Timer[T any](ctx context.Context, duration time.Duration, value T) Stream[T]   // single value
```

Behavior notes:
- Windows use wall‑clock arrival time. Tumbling emits non‑overlapping buckets; Sliding emits at `slide` cadence and keeps elements within the last `windowSize`; Session splits when no arrival within `gap`.
- Debounce emits the last value after a quiet period; Sample emits the latest value on each tick.
- Timeout yields `Err(context.DeadlineExceeded)` if no element arrives in `d`; resets on each element.
- All ctx operators drain timers safely (stop+drain) to avoid spurious wakeups.

Examples:
```go
ctx, cancel := context.WithCancel(context.Background()); defer cancel()

// Interval
ticks := streams.Interval(ctx, 10*time.Millisecond).Limit(3).Collect() // [0 1 2]

// Throttle / RateLimit
slow := streams.Throttle(streams.Range(1,5), 5*time.Millisecond).Collect()
rl   := streams.RateLimit(streams.Range(1,5), 2, 5*time.Millisecond).Collect()

// Debounce
deb := streams.Debounce(ctx, streams.Of(1,2,3), 1*time.Millisecond).Collect() // emits last value
```

### Stream2[K,V] (Key‑Value Streams)

Constructors and interop:
```go
func From2[K,V any](seq iter.Seq2[K,V]) Stream2[K,V]
func (s Stream2[K,V]) Seq2() iter.Seq2[K,V]
func PairsOf[K,V any](pairs ...Pair[K,V]) Stream2[K,V]
func Empty2[K,V any]() Stream2[K,V]
```

Intermediate:
```go
func (s Stream2[K,V]) Filter(pred func(K,V) bool) Stream2[K,V]
func (s Stream2[K,V]) MapKeys(fn func(K) K) Stream2[K,V]
func (s Stream2[K,V]) MapValues(fn func(V) V) Stream2[K,V]
func (s Stream2[K,V]) Limit(n int) Stream2[K,V]
func (s Stream2[K,V]) Skip(n int) Stream2[K,V]
func (s Stream2[K,V]) Peek(action func(K,V)) Stream2[K,V]
func (s Stream2[K,V]) TakeWhile(pred func(K,V) bool) Stream2[K,V]
func (s Stream2[K,V]) DropWhile(pred func(K,V) bool) Stream2[K,V]
func (s Stream2[K,V]) ParallelMapValues(fn func(V) V, opts ...ParallelOption) Stream2[K,V]
func (s Stream2[K,V]) ParallelFilter(pred func(K,V) bool, opts ...ParallelOption) Stream2[K,V]
```

Terminals and transformations:
```go
func (s Stream2[K,V]) Keys() Stream[K]
func (s Stream2[K,V]) Values() Stream[V]
func (s Stream2[K,V]) ToPairs() Stream[Pair[K,V]]
func (s Stream2[K,V]) ForEach(action func(K,V))
func (s Stream2[K,V]) Count() int
func (s Stream2[K,V]) AnyMatch(pred func(K,V) bool) bool
func (s Stream2[K,V]) AllMatch(pred func(K,V) bool) bool
func (s Stream2[K,V]) NoneMatch(pred func(K,V) bool) bool
func (s Stream2[K,V]) First() Optional[Pair[K,V]]
func (s Stream2[K,V]) CollectPairs() []Pair[K,V]
func (s Stream2[K,V]) Reduce(identity Pair[K,V], fn func(Pair[K,V], K, V) Pair[K,V]) Pair[K,V]

// Free transformations
func MapKeysTo[K,V,K2 any](s Stream2[K,V], fn func(K) K2) Stream2[K2,V]
func MapValuesTo[K,V,V2 any](s Stream2[K,V], fn func(V) V2) Stream2[K,V2]
func MapPairs[K,V,K2,V2 any](s Stream2[K,V], fn func(K,V) (K2,V2)) Stream2[K2,V2]
func SwapKeyValue[K,V any](s Stream2[K,V]) Stream2[V,K]
func ToMap2[K comparable, V any](s Stream2[K,V]) map[K]V
func ReduceByKey[K comparable, V any](s Stream2[K,V], merge func(V,V) V) map[K]V
func ReduceByKeyWithInit[K comparable, V, R any](s Stream2[K,V], init func() R, merge func(R,V) R) map[K]R
func GroupValues[K comparable, V any](s Stream2[K,V]) map[K][]V
func DistinctKeys[K comparable, V any](s Stream2[K,V]) Stream2[K,V]
func DistinctValues[K any, V comparable](s Stream2[K,V]) Stream2[K,V]
```

Examples:
```go
m := map[string]int{"a":1, "b":2, "b":3}
s2 := streams.FromMap(m)

onlyKeys := s2.Keys().Collect()                    // ["a","b"]
values := s2.Values().Collect()
distinctKeys := streams.DistinctKeys(s2).ToPairs().Collect()
grouped := streams.GroupValues(streams.PairsOf(streams.NewPair("k",1), streams.NewPair("k",2))) // {"k":[1,2]}
byKeys := streams.ReduceByKey(streams.PairsOf(streams.NewPair("k",1), streams.NewPair("k",2)), func(a,b int) int { return a+b }) // {"k":3}
```

### Join Family (on Stream2 and on Stream[T] with keys)

```go
// Stream2 joins (collect into lookup maps internally)
type JoinResult[K,V1,V2 any] struct { Key K; Left V1; Right V2 }
type JoinResultOptional[K,V1,V2 any] struct { Key K; Left Optional[V1]; Right Optional[V2] }

func InnerJoin[K comparable, V1, V2 any](s1 Stream2[K,V1], s2 Stream2[K,V2]) Stream[JoinResult[K,V1,V2]]
func LeftJoin[K comparable, V1, V2 any](s1 Stream2[K,V1], s2 Stream2[K,V2]) Stream[JoinResultOptional[K,V1,V2]]
func RightJoin[K comparable, V1, V2 any](s1 Stream2[K,V1], s2 Stream2[K,V2]) Stream[JoinResultOptional[K,V1,V2]]
func FullJoin[K comparable, V1, V2 any](s1 Stream2[K,V1], s2 Stream2[K,V2]) Stream[JoinResultOptional[K,V1,V2]]

// Defaults
func LeftJoinWith[K comparable, V1, V2 any](s1 Stream2[K,V1], s2 Stream2[K,V2], defaultV2 V2) Stream[JoinResult[K,V1,V2]]
func RightJoinWith[K comparable, V1, V2 any](s1 Stream2[K,V1], s2 Stream2[K,V2], defaultV1 V1) Stream[JoinResult[K,V1,V2]]

// CoGroup
type CoGrouped[K,V1,V2 any] struct { Key K; Left []V1; Right []V2 }
func CoGroup[K comparable, V1, V2 any](s1 Stream2[K,V1], s2 Stream2[K,V2]) Stream[CoGrouped[K,V1,V2]]

// Stream[T] join by keys
func JoinBy[T,U,K comparable](s1 Stream[T], s2 Stream[U], keyT func(T) K, keyU func(U) K) Stream[Pair[T,U]]
func LeftJoinBy[T,U any, K comparable](s1 Stream[T], s2 Stream[U], keyT func(T) K, keyU func(U) K) Stream[Pair[T, Optional[U]]]
func SemiJoin[K comparable, V1, V2 any](s1 Stream2[K,V1], s2 Stream2[K,V2]) Stream2[K,V1]
func AntiJoin[K comparable, V1, V2 any](s1 Stream2[K,V1], s2 Stream2[K,V2]) Stream2[K,V1]
func SemiJoinBy[T,U any, K comparable](s1 Stream[T], s2 Stream[U], keyT func(T) K, keyU func(U) K) Stream[T]
func AntiJoinBy[T,U any, K comparable](s1 Stream[T], s2 Stream[U], keyT func(T) K, keyU func(U) K) Stream[T]
```

Notes:
- Joins build in‑memory lookups (maps) of one/both inputs; ensure inputs are bounded.

Examples:
```go
left  := streams.PairsOf(streams.NewPair("a", 1), streams.NewPair("b", 2))
right := streams.PairsOf(streams.NewPair("a", "x"))
ij := streams.InnerJoin(left, right).Collect()  // [ {Key:"a", Left:1, Right:"x"} ]
lj := streams.LeftJoin(left, right).Collect()   // includes "b" with Right=None

// Join by key extractors on Stream[T]
users := streams.Of(struct{ID string}{"u1"}, struct{ID string}{"u2"})
orders := streams.Of(struct{UID string}{"u1"})
joined := streams.JoinBy(users, orders, func(u struct{ID string}) string { return u.ID }, func(o struct{UID string}) string { return o.UID }).Collect()
```

### Numeric and Statistics

```go
// Aggregations on Numeric (ints/floats)
func Sum[T Numeric](s Stream[T]) T
func Average[T Numeric](s Stream[T]) Optional[float64]
func MinValue[T cmp.Ordered](s Stream[T]) Optional[T]
func MaxValue[T cmp.Ordered](s Stream[T]) Optional[T]
func MinMax[T cmp.Ordered](s Stream[T]) Optional[Pair[T,T]]
func Product[T Numeric](s Stream[T]) T
func SumBy[T any, N Numeric](s Stream[T], fn func(T) N) N
func AverageBy[T any, N Numeric](s Stream[T], fn func(T) N) Optional[float64]
func MinBy[T any, K cmp.Ordered](s Stream[T], fn func(T) K) Optional[T]
func MaxBy[T any, K cmp.Ordered](s Stream[T], fn func(T) K) Optional[T]

// Running/transform
func RunningSum[T Numeric](s Stream[T]) Stream[T]
func RunningProduct[T Numeric](s Stream[T]) Stream[T]
func Differences[T Numeric](s Stream[T]) Stream[T]
func Clamp[T cmp.Ordered](s Stream[T], minVal, maxVal T) Stream[T]
func Abs[T Signed](s Stream[T]) Stream[T]
func AbsFloat[T Float](s Stream[T]) Stream[T]
func Scale[T Numeric](s Stream[T], factor T) Stream[T]
func Offset[T Numeric](s Stream[T], offset T) Stream[T]
func Positive[T Numeric](s Stream[T]) Stream[T]
func Negative[T Signed](s Stream[T]) Stream[T]
func NonZero[T Numeric](s Stream[T]) Stream[T]

// Statistics struct
type Statistics[T Numeric] struct {
    Count int
    Sum   T
    Min   T
    Max   T
    Average float64
}
func GetStatistics[T Numeric](s Stream[T]) Optional[Statistics[T]]
```

Examples:
```go
nums := streams.Of(10,20,30)
sum := streams.Sum(nums) // 60
avg := streams.Average(nums).Get() // 20.0
stats := streams.GetStatistics(nums).Get()
run := streams.RunningSum(nums).Collect() // [10 30 60]
diff := streams.Differences(streams.Of(1,4,9)).Collect() // [3 5]
```

### Collectors (Composable Accumulators)

```go
// Core
type Collector[T, A, R any] struct { /* builder for accumulation */ }
func CollectTo[T, A, R any](s Stream[T], c Collector[T,A,R]) R
func ToSliceCollector[T any]() Collector[T, []T, []T]
func ToSetCollector[T comparable]() Collector[T, map[T]struct{}, map[T]struct{}]
func JoiningCollector(sep string) Collector[string, *strings.Builder, string]
func JoiningCollectorFull(sep, prefix, suffix string) Collector[string, *strings.Builder, string]
func CountingCollector[T any]() Collector[T, *countingState, int]
func SummingCollector[T Numeric]() Collector[T, *summingState[T], T]
func AveragingCollector[T Numeric]() Collector[T, *averagingState, Optional[float64]]
func MaxByCollector[T any](cmp func(T,T) int) Collector[T, *maxState[T], Optional[T]]
func MinByCollector[T any](cmp func(T,T) int) Collector[T, *minState[T], Optional[T]]

// Grouping and maps
func GroupingByCollector[T any, K comparable](keyFn func(T) K) Collector[T, map[K][]T, map[K][]T]
func PartitioningByCollector[T any](pred func(T) bool) Collector[T, *partitionState[T], map[bool][]T]
func ToMapCollector[T any, K comparable, V any](keyFn func(T) K, valFn func(T) V) Collector[T, map[K]V, map[K]V]
func ToMapCollectorMerging[T any, K comparable, V any](keyFn func(T) K, valFn func(T) V, merge func(V,V) V) Collector[T, map[K]V, map[K]V]

// Composition
func MappingCollector[T,U,A,R any](mapper func(T) U, downstream Collector[U,A,R]) Collector[T,A,R]
func FilteringCollector[T,A,R any](pred func(T) bool, downstream Collector[T,A,R]) Collector[T,A,R]
func FlatMappingCollector[T,U,A,R any](mapper func(T) Stream[U], downstream Collector[U,A,R]) Collector[T,A,R]
func TeeingCollector[T,A1,R1,A2,R2,R any](c1 Collector[T,A1,R1], c2 Collector[T,A2,R2], merge func(R1,R2) R) Collector[T, *teeingState[T,A1,A2], R]

// Ranking and statistics
func TopKCollector[T any](k int, less func(T,T) bool) Collector[T, *topKState[T], []T]
func BottomKCollector[T any](k int, less func(T,T) bool) Collector[T, *bottomKState[T], []T]
func QuantileCollector[T any](q float64, less func(T,T) bool) Collector[T, *quantileState[T], Optional[T]]

// Convenience helpers backed by collectors
func TopK[T any](s Stream[T], k int, less func(T,T) bool) []T
func BottomK[T any](s Stream[T], k int, less func(T,T) bool) []T
func Quantile[T any](s Stream[T], q float64, less func(T,T) bool) Optional[T]
func Median[T any](s Stream[T], less func(T,T) bool) Optional[T]
func Percentile[T any](s Stream[T], p float64, less func(T,T) bool) Optional[T] // p in [0,100]
func FrequencyCollector[T comparable]() Collector[T, map[T]int, map[T]int]
func Frequency[T comparable](s Stream[T]) map[T]int
func MostCommon[T comparable](s Stream[T], n int) []Pair[T,int]
func HistogramCollector[T any, K comparable](keyFn func(T) K) Collector[T, *histogramState[T,K], map[K][]T]
```

Examples:
```go
// CollectTo with core collectors
xs := streams.Of(1,2,2,3)
slice := streams.CollectTo(xs, streams.ToSliceCollector[int]())         // []int
set   := streams.CollectTo(xs, streams.ToSetCollector[int]())           // map[int]struct{}
cnt   := streams.CollectTo(xs, streams.CountingCollector[int]())        // 4
max   := streams.CollectTo(xs, streams.MaxByCollector[int](func(a,b int) int { return a-b })).Get()

// Grouping and composition
grp := streams.CollectTo(streams.Of("a","bb","c"), streams.GroupingByCollector(func(s string) int { return len(s) })) // map[int][]string
mappedAndGrouped := streams.CollectTo(streams.Of("a","bb"), streams.MappingCollector(func(s string) string { return strings.ToUpper(s) }, streams.GroupingByCollector(func(s string) int { return len(s) })))

// TopK convenience
top2 := streams.TopK(streams.Of(5,1,4,3,2), 2, func(a,b int) bool { return a<b }) // [5 4]
median := streams.Median(streams.Of(1,3,2), func(a,b int) bool { return a<b }).Get() // 2
freq := streams.Frequency(streams.Of("a","b","a")) // map[string]int{"a":2,"b":1}
```

### Result[T] Pipeline (Error‑Aware Streams)

```go
// Result primitives
type Result[T any] struct { /* Ok(value) or Err(error) */ }
func Ok[T any](value T) Result[T]
func Err[T any](err error) Result[T]
func ErrMsg[T any](msg string) Result[T]
func (r Result[T]) IsOk() bool
func (r Result[T]) IsErr() bool
func (r Result[T]) Unwrap() T                 // panics on Err
func (r Result[T]) UnwrapOr(defaultVal T) T
func (r Result[T]) UnwrapOrElse(fn func(error) T) T
func (r Result[T]) UnwrapErr() error
func (r Result[T]) Error() error
func (r Result[T]) Value() T                  // zero on Err
func (r Result[T]) Get() (T, error)
func (r Result[T]) ToOptional() Optional[T]
func (r Result[T]) Map(fn func(T) T) Result[T]
func (r Result[T]) MapErr(fn func(error) error) Result[T]
func (r Result[T]) And(other Result[T]) Result[T]
func (r Result[T]) Or(other Result[T]) Result[T]
func MapResultTo[T,U any](r Result[T], fn func(T) U) Result[U]
func FlatMapResult[T,U any](r Result[T], fn func(T) Result[U]) Result[U]

// Error‑aware stream ops
func MapErrTo[T,U any](s Stream[T], fn func(T) (U, error)) Stream[Result[U]]
func FilterErr[T any](s Stream[T], pred func(T) (bool, error)) Stream[Result[T]]
func FlatMapErr[T,U any](s Stream[T], fn func(T) (Stream[U], error)) Stream[Result[U]]

// Collectors over Result streams
func CollectResults[T any](s Stream[Result[T]]) ([]T, error)        // stops at first Err
func CollectResultsAll[T any](s Stream[Result[T]]) ([]T, []error)   // collects all
func PartitionResults[T any](s Stream[Result[T]]) ([]T, []error)
func FilterOk[T any](s Stream[Result[T]]) Stream[T]
func FilterErrs[T any](s Stream[Result[T]]) Stream[error]
func UnwrapResults[T any](s Stream[Result[T]]) Stream[T]             // panics on Err
func UnwrapOrDefault[T any](s Stream[Result[T]], defaultVal T) Stream[T]
func TakeUntilErr[T any](s Stream[Result[T]]) Stream[T]
func FromResults[T any](results ...Result[T]) Stream[Result[T]]
func TryCollect[T any](s Stream[T]) Result[[]T]                      // trap panics
```

Examples:
```go
// Constructing Results
ok := streams.Ok(42)
er := streams.Err[int](errors.New("boom"))

// Error-aware map
parsed := streams.MapErrTo(streams.Of("1","x","2"), strconv.Atoi).Collect()
vals, err := streams.CollectResults(parsed) // vals: [1], err: from "x"

// FilterOk
oks := streams.FilterOk(parsed).Collect() // keep only successful ints
```

### Optional[T]

```go
func Some[T any](value T) Optional[T]
func None[T any]() Optional[T]
func OptionalOf[T any](ptr *T) Optional[T]
func OptionalFromCondition[T any](condition bool, value T) Optional[T]
func (o Optional[T]) IsPresent() bool
func (o Optional[T]) IsEmpty() bool
func (o Optional[T]) Get() T                    // panics if empty
func (o Optional[T]) GetOrElse(defaultVal T) T
func (o Optional[T]) GetOrElseGet(supplier func() T) T
func (o Optional[T]) GetOrZero() T
func (o Optional[T]) IfPresent(action func(T))
func (o Optional[T]) IfPresentOrElse(action func(T), emptyAction func())
func (o Optional[T]) Filter(pred func(T) bool) Optional[T]
func (o Optional[T]) Map(fn func(T) T) Optional[T]
func (o Optional[T]) OrElse(other Optional[T]) Optional[T]
func (o Optional[T]) OrElseGet(supplier func() Optional[T]) Optional[T]
func (o Optional[T]) ToSlice() []T
func (o Optional[T]) ToPointer() *T
func (o Optional[T]) ToStream() Stream[T]
func OptionalMap[T,U any](o Optional[T], fn func(T) U) Optional[U]
func OptionalFlatMap[T,U any](o Optional[T], fn func(T) Optional[U]) Optional[U]
func OptionalZip[T,U any](o1 Optional[T], o2 Optional[U]) Optional[Pair[T,U]]
func OptionalEquals[T comparable](o1, o2 Optional[T]) bool
```

Examples:
```go
o := streams.Some(10)
_ = o.GetOrElse(0)                     // 10
o2 := streams.OptionalMap(o, func(v int) string { return strconv.Itoa(v) }) // Some("10")
both := streams.OptionalZip(streams.Some(1), streams.Some("a")).Get()       // Pair(1,"a")
```

### Tuples and Helpers

```go
// Pair, Triple, Quad
type Pair[T,U any] struct { First T; Second U }
func NewPair[T,U any](first T, second U) Pair[T,U]
func (p Pair[T,U]) Swap() Pair[U,T]
func (p Pair[T,U]) MapFirst(fn func(T) T) Pair[T,U]
func (p Pair[T,U]) MapSecond(fn func(U) U) Pair[T,U]
func (p Pair[T,U]) Unpack() (T,U)

type Triple[A,B,C any] struct { First A; Second B; Third C }
func NewTriple[A,B,C any](a A, b B, c C) Triple[A,B,C]
func (t Triple[A,B,C]) ToPair() Pair[A,B]
func (t Triple[A,B,C]) Unpack() (A,B,C)
func (t Triple[A,B,C]) MapFirst(fn func(A) A) Triple[A,B,C]
func (t Triple[A,B,C]) MapSecond(fn func(B) B) Triple[A,B,C]
func (t Triple[A,B,C]) MapThird(fn func(C) C) Triple[A,B,C]

type Quad[A,B,C,D any] struct { First A; Second B; Third C; Fourth D }
func NewQuad[A,B,C,D any](a A, b B, c C, d D) Quad[A,B,C,D]
func (q Quad[A,B,C,D]) ToTriple() Triple[A,B,C]
func (q Quad[A,B,C,D]) ToPair() Pair[A,B]
func (q Quad[A,B,C,D]) Unpack() (A,B,C,D)

// Other helpers
func Unzip[T,U any](s Stream[Pair[T,U]]) ([]T, []U)
```

Examples:
```go
p := streams.NewPair(1,"x")
a,b := p.Unpack()
t := streams.NewTriple(1,"x",true)
xs, ys := streams.Unzip(streams.Of(streams.NewPair(1,"a"), streams.NewPair(2,"b")))
```

### go-collections Integration

go-streams provides seamless integration with [go-collections](https://github.com/coldsmirk/go-collections), a comprehensive collections library offering Set, List, Map, Queue, Stack, and more.

#### Constructors from go-collections

```go
// From Set/SortedSet
streams.FromSet(hashSet)                    // Stream from Set
streams.FromSortedSet(treeSet)              // Stream in ascending order
streams.FromSortedSetDescending(treeSet)    // Stream in descending order

// From List
streams.FromList(arrayList)                 // Stream from List

// From Map/SortedMap
streams.FromMapC(hashMap)                   // Stream2 from collections.Map
streams.FromSortedMapC(treeMap)             // Stream2 in ascending key order
streams.FromSortedMapCDescending(treeMap)   // Stream2 in descending key order

// From Queue/Stack/Deque
streams.FromQueue(queue)                    // FIFO order
streams.FromStack(stack)                    // LIFO order
streams.FromDeque(deque)                    // Front to back
streams.FromPriorityQueue(pq)               // Heap order
streams.FromPriorityQueueSorted(pq)         // Priority order (collects first)
```

#### Terminal Operations returning go-collections

```go
// Collect into Set
set := streams.ToHashSet(s)                                      // collections.Set[T]
sortedSet := streams.ToTreeSet(s, cmp.Compare[int])              // collections.SortedSet[T]

// Collect into List
list := streams.ToArrayList(s)                                   // collections.List[T]
list := streams.ToLinkedList(s)                                  // LinkedList implementation

// Collect into Map
m := streams.ToHashMapC(s, keyFn, valFn)                         // collections.Map[K,V]
m := streams.ToTreeMapC(s, keyFn, valFn, keyCmp)                 // collections.SortedMap[K,V]
m := streams.ToHashMap2C(stream2)                                // From Stream2 to Map

// GroupBy into collections.Map
m := streams.GroupByToHashMap(s, keyFn)                          // collections.Map[K,[]T]
m := streams.GroupByToTreeMap(s, keyFn, keyCmp)                  // collections.SortedMap[K,[]T]

// Frequency into collections.Map
freq := streams.FrequencyToHashMap(s)                            // collections.Map[T,int]
```

#### Collectors returning go-collections

```go
// Collector for Set
streams.CollectTo(s, streams.ToHashSetCollector[int]())          // collections.Set
streams.CollectTo(s, streams.ToTreeSetCollector(cmp))            // collections.SortedSet

// Collector for List
streams.CollectTo(s, streams.ToArrayListCollector[T]())          // collections.List

// Collector for Map
streams.CollectTo(s, streams.ToHashMapCollector(keyFn, valFn))   // collections.Map
streams.CollectTo(s, streams.ToTreeMapCollector(keyFn, valFn, keyCmp)) // collections.SortedMap
```

#### Set Operations Example

```go
// Collect into HashSet and use set algebra
set1 := streams.ToHashSet(streams.Of(1, 2, 3, 4))
set2 := streams.ToHashSet(streams.Of(3, 4, 5, 6))

union := set1.Union(set2)           // {1, 2, 3, 4, 5, 6}
inter := set1.Intersection(set2)    // {3, 4}
diff := set1.Difference(set2)       // {1, 2}
symDiff := set1.SymmetricDifference(set2) // {1, 2, 5, 6}

// Check relations
set1.IsSubsetOf(set2)      // false
set1.IsDisjoint(set2)      // false
set1.Equals(set2)          // false
```

### Practical Guidance

- Prefer lazy operators to keep memory bounded; be mindful of eager ones noted above.
- For very large sub‑streams with `ParallelFlatMap` and ordered output, prefer `WithChunkSize` or `WithOrdered(false)` to bound memory.
- When joining or doing Cartesian/combinatorics, inputs are collected; validate sizes or pre‑filter.
- Use `ctx` variants in long‑running or IO/timer pipelines to support cancellation and timeouts cleanly.

---

## License

MIT, see [LICENSE](LICENSE).
