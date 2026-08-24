# go-streams

[![Go Reference](https://pkg.go.dev/badge/github.com/coldsmirk/go-streams.svg)](https://pkg.go.dev/github.com/coldsmirk/go-streams)
[![Go Report Card](https://goreportcard.com/badge/github.com/coldsmirk/go-streams)](https://goreportcard.com/report/github.com/coldsmirk/go-streams)
[![Build Status](https://github.com/coldsmirk/go-streams/actions/workflows/test.yml/badge.svg?branch=v1)](https://github.com/coldsmirk/go-streams/actions/workflows/test.yml?query=branch%3Av1)
[![codecov](https://codecov.io/gh/coldsmirk/go-streams/branch/v1/graph/badge.svg)](https://codecov.io/gh/coldsmirk/go-streams/branch/v1)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A lazy, type-safe stream processing library for Go 1.25+, built on `iter.Seq`
and `iter.Seq2`.

> **This is the v1 maintenance branch.** v1 remains available at
> `github.com/coldsmirk/go-streams` and receives fixes only. Active development
> happens in v2, which requires Go 1.27 and uses generic methods so that
> type-changing operations stay in the chain — see the
> [main branch](https://github.com/coldsmirk/go-streams) and the
> [migration guide](https://github.com/coldsmirk/go-streams/blob/main/MIGRATION.md).

```sh
go get github.com/coldsmirk/go-streams
```

```go
import "github.com/coldsmirk/go-streams"

evens := streams.Of(1, 2, 3, 4, 5).
    Filter(func(n int) bool { return n%2 == 0 }).
    Map(func(n int) int { return n * 2 }).
    Collect()                                    // [4 8]

// Changing the element type leaves the chain — see Design below.
names := streams.MapTo(
    streams.Of(users...).Filter(func(u User) bool { return u.Active }),
    func(u User) string { return u.Name },
).Collect()
```

## Design

**A Stream is a lazy sequence over an iterator.**

| Type | What it is |
| --- | --- |
| `Stream[T]` | a lazy sequence of `T` |
| `Stream2[K, V]` | a lazy sequence of key-value pairs |
| `Optional[T]` | a value that may be absent |
| `Result[T]` | a value or an error |
| `Collector[T, A, R]` | a strategy for accumulating elements into a result |

**Methods where the element type survives, functions where it does not.**

Go 1.25 has no type parameters on methods, so a method cannot introduce a new
type parameter and return a `Stream` of it:

```go
// not expressible in Go 1.25:
// func (s Stream[T]) MapTo[U any](fn func(T) U) Stream[U]

// so the operation is a free function:
func MapTo[T, U any](s Stream[T], fn func(T) U) Stream[U]
```

That split runs through the whole API. `Filter`, `Map` and `Peek` are methods
because `T` survives; `MapTo`, `FlatMap`, `Zip` and `Chunk` are functions
because it does not. Go 1.27 added generic methods, which is what v2 is built
on — there the split is gone.

**Laziness is real.**

```go
// builds a pipeline; nothing has run yet
pipeline := streams.Range(1, 1000000).
    Filter(func(n int) bool { return n%2 == 0 }).
    Map(func(n int) int { return n * 2 }).
    Limit(5)

result := pipeline.Collect()   // touches 10 elements, not a million
```

## Core API

Full documentation and runnable examples:
[pkg.go.dev](https://pkg.go.dev/github.com/coldsmirk/go-streams), or
`go doc -all github.com/coldsmirk/go-streams`.

**Constructing** — `From`, `Of`, `FromSlice`, `FromMap`, `FromChannel`,
`Generate`, `Iterate`, `Range`, `RangeClosed`, `Repeat`, `RepeatForever`,
`Cycle`, `Concat`, `Empty`; for pairs `From2`, `PairsOf`, `Empty2`

**`Stream[T]` methods** — `Seq`, `Filter`, `Map`, `Peek`, `Limit`, `Skip`,
`TakeWhile`, `DropWhile`, `Sorted`, `SortedStable`, `Reverse`, `Step`,
`TakeLast`, `DropLast`, `Intersperse`

**Type-changing functions** — `MapTo`, `FlatMap`, `FlatMapSeq`, `Flatten`,
`FlattenSeq`, `Scan`, `Zip`, `ZipWithIndex`, `ZipLongest`, `ZipLongestWith`,
`Zip3`, `Unzip`, `Distinct`, `DistinctBy`, `DistinctUntilChanged`,
`DistinctUntilChangedBy`, `SortedBy`, `SortedStableBy`, `Chunk`, `Window`,
`WindowWithStep`, `Pairwise`, `Triples`, `Interleave`

**Specialized** — `MergeSorted`, `MergeSortedN`, `MergeSortedNHeap`,
`Cartesian`, `CartesianSelf`, `CrossProduct`, `Combinations`, `Permutations`

**Terminals** — `Collect`, `ForEach`, `ForEachIndexed`, `ForEachErr`,
`ForEachIndexedErr`, `Reduce`, `ReduceOptional`, `Fold`, `FoldTo`, `Count`,
`First`, `Last`, `FindFirst`, `FindLast`, `AnyMatch`, `AllMatch`, `NoneMatch`,
`Min`, `Max`, `At`, `Nth`, `Single`, `IsEmpty`, `IsNotEmpty`, `Contains`,
`ToMap`, `ToSet`, `GroupBy`, `GroupByTo`, `PartitionBy`, `AssociateBy`,
`IndexBy`, `CountBy`, `Frequencies`, `Joining`, `JoiningWithPrefixSuffix`

**`Stream2[K, V]`** — methods `Seq2`, `Filter`, `MapKeys`, `MapValues`,
`Limit`, `Skip`, `Peek`, `TakeWhile`, `DropWhile`, `Keys`, `Values`, `ToPairs`,
`First`, `ForEach`, `Count`, `AnyMatch`, `AllMatch`, `NoneMatch`,
`CollectPairs`, `Reduce`, `ParallelMapValues`, `ParallelFilter`; functions
`MapKeysTo`, `MapValuesTo`, `MapPairs`, `MapPairsTo`, `MapToPairs`,
`SwapKeyValue`, `ToMap2`, `ReduceByKey`, `ReduceByKeyWithInit`, `GroupValues`,
`DistinctKeys`, `DistinctValues`

**Parallel** — `ParallelMap`, `ParallelFilter`, `ParallelFlatMap`,
`ParallelForEach`, `ParallelReduce`, `ParallelCollect`, `Prefetch`, configured
with `WithConcurrency`, `WithOrdered`, `WithBufferSize`, `WithChunkSize`
(`ParallelOption`, `ParallelConfig`, `DefaultParallelConfig`)

**Context-aware** — `WithContext`, `WithContext2`, and the `Ctx` variants
`GenerateCtx`, `IterateCtx`, `RangeCtx`, `FromChannelCtx`,
`FromReaderLinesCtx`, `FilterCtx`, `MapCtx`, `MapToCtx`, `CollectCtx`,
`ForEachCtx`, `ReduceCtx`, `FindFirstCtx`, `AnyMatchCtx`, `AllMatchCtx`,
`CountCtx`, `ParallelMapCtx`, `ParallelFilterCtx`, `ParallelFlatMapCtx`,
`ParallelForEachCtx`, plus `ContextError`

**`Optional[T]`** — `Some`, `None`, `OptionalOf`, `OptionalFromCondition`,
`OptionalMap`, `OptionalFlatMap`, `OptionalZip`, `OptionalEquals`; methods
`IsPresent`, `IsEmpty`, `Get`, `GetOrElse`, `GetOrElseGet`, `GetOrZero`,
`IfPresent`, `IfPresentOrElse`, `Filter`, `Map`, `OrElse`, `OrElseGet`,
`ToSlice`, `ToPointer`, `ToStream`, `String`

**`Result[T]`** — `Ok`, `Err`, `ErrMsg`, `MapResultTo`, `FlatMapResult`,
`MapErrTo`, `FilterErr`, `FlatMapErr`, `CollectResults`, `CollectResultsAll`,
`PartitionResults`, `FilterOk`, `FilterErrs`, `UnwrapResults`,
`UnwrapOrDefault`, `TakeUntilErr`, `FromResults`, `TryCollect`; methods
`IsOk`, `IsErr`, `Unwrap`, `UnwrapOr`, `UnwrapOrElse`, `UnwrapErr`, `Value`,
`Get`, `Error`, `ToOptional`, `Map`, `MapErr`, `And`, `Or`

**Tuples** — `Pair`/`NewPair`, `Triple`/`NewTriple`, `Quad`/`NewQuad`, with
`Swap`, `Unpack`, `MapFirst`, `MapSecond`, `MapThird`, `ToPair`, `ToTriple`

**Numeric and statistics** — `Sum`, `Average`, `MinValue`, `MaxValue`,
`MinMax`, `Product`, `SumBy`, `AverageBy`, `MinBy`, `MaxBy`, `GetStatistics`
(`Statistics`), `RunningSum`, `RunningProduct`, `Differences`, `Clamp`, `Abs`,
`AbsFloat`, `Scale`, `Offset`, `Positive`, `Negative`, `NonZero`, over the
constraints `Numeric`, `Signed`, `Unsigned`, `Integer`, `Float`

**Collectors** — `CollectTo` with `ToSliceCollector`, `ToSetCollector`,
`JoiningCollector`, `JoiningCollectorFull`, `CountingCollector`,
`SummingCollector`, `AveragingCollector`, `MaxByCollector`, `MinByCollector`,
`GroupingByCollector`, `PartitioningByCollector`, `ToMapCollector`,
`ToMapCollectorMerging`, `FirstCollector`, `LastCollector`, `ReducingCollector`,
`MappingCollector`, `FilteringCollector`, `FlatMappingCollector`,
`TeeingCollector`, `TopKCollector`, `BottomKCollector`, `QuantileCollector`,
`FrequencyCollector`, `HistogramCollector`; shorthands `TopK`, `BottomK`,
`Quantile`, `Median`, `Percentile`, `Frequency`, `MostCommon`

**Joins** — `InnerJoin`, `LeftJoin`, `RightJoin`, `FullJoin`, `LeftJoinWith`,
`RightJoinWith`, `CoGroup`, `JoinBy`, `LeftJoinBy`, `SemiJoin`, `AntiJoin`,
`SemiJoinBy`, `AntiJoinBy`, producing `JoinResult`, `JoinResultOptional`,
`CoGrouped`

**IO** — `FromReaderLines`, `FromScanner`, `FromStringLines`, `FromBytes`,
`FromRunes`, `FromCSV`, `FromTSV`, `FromCSVFile`, `FromTSVFile`,
`FromCSVWithHeader`, `FromFileLines`; the error-returning forms
`FromReaderLinesErr`, `FromScannerErr`, `FromCSVErr`, `FromTSVErr`,
`FromCSVWithHeaderErr`; the panicking forms `MustFromFileLines`,
`MustFromCSVFile`, `MustFromTSVFile`; the writers `ToWriter`, `ToFile`,
`ToCSV`, `ToCSVFile`; and the types `FileLineStream` and `CSVStream` (each with
`Close`) and `CSVRecord` (with `Get`, `GetOr`)

**Time-based** — `WithTimestamp` (`TimestampedValue`, `NewTimestamped`,
`NewTimestampedAt`), `TumblingTimeWindow`, `SlidingTimeWindow`, `SessionWindow`,
`Throttle`, `RateLimit`, `Debounce`, `Sample`, `Delay`, `Timeout`, `Interval`,
`Timer`, with `ThrottleCtx`, `RateLimitCtx`, `DelayCtx`

**go-collections bridge** — `FromSet`, `FromSortedSet`,
`FromSortedSetDescending`, `FromList`, `FromMapC`, `FromSortedMapC`,
`FromSortedMapCDescending`, `FromQueue`, `FromStack`, `FromDeque`,
`FromDequeReversed`, `FromPriorityQueue`, `FromPriorityQueueSorted`, and
`ToHashSet`, `ToTreeSet`, `ToArrayList`, `ToLinkedList`, `ToHashMapC`,
`ToTreeMapC`, `ToHashMap2C`, `ToTreeMap2C`; the collector forms
`ToHashSetCollector`, `ToTreeSetCollector`, `ToArrayListCollector`,
`ToHashMapCollector`, `ToTreeMapCollector`, `CollectToSet`, `CollectToList`,
`CollectToMap`; and the grouping forms `GroupByToHashMap`, `GroupByToTreeMap`,
`GroupValuesToHashMap`, `FrequencyToHashMap`, `HistogramToHashMap`

**Resource management** — `Using`

## Semantics worth knowing

**Streams are single-pass.** A terminal operation consumes the underlying
iterator. Traverse a Stream once; build a new one to traverse again.

**File sources own a handle and must be closed.** `FromFileLines` returns a
`*FileLineStream`; `Using` closes it defer-style, even if the body panics.

```go
lines := streams.Using(stream, func(s *streams.FileLineStream) []string {
    return s.Filter(isError).Collect()
})
```

**Some operations must buffer.** `Sorted`, `SortedStable`, `Reverse`,
`TakeLast`, `DropLast` and `Cycle` read the whole sequence, so they cannot be
used on an infinite stream. Everything else streams.

**Boundary conventions.** Where a count is out of range, an operation either
yields nothing or passes the stream through unchanged — never panics:

| Operation | `n <= 0` | Note |
| --- | --- | --- |
| `Limit(n)` | empty stream | `Limit(0)` produces `[]` |
| `Skip(n)` | original stream | `Skip(0)` keeps every element |
| `TakeLast(n)` | empty stream | |
| `DropLast(n)` | original stream | |
| `Step(n)` | original stream | `Step(0)` and `Step(1)` keep every element |
| `Chunk(s, n)` | empty stream | |
| `Window(s, n)` | empty stream | |
| `WindowWithStep(s, size, step, _)` | empty stream | if `size <= 0` or `step <= 0` |

**Window overlaps, Chunk does not.** Over `[1,2,3,4,5]`:

| Call | Result |
| --- | --- |
| `Window(s, 3)` | `[[1,2,3], [2,3,4], [3,4,5]]` |
| `Chunk(s, 3)` | `[[1,2,3], [4,5]]` |
| `WindowWithStep(s, 3, 2, false)` | `[[1,2,3], [3,4,5]]` |

## Requirements

Go 1.25 or later.

## Contributing

v1 is in maintenance: it takes bug fixes and security fixes, not new features.
New work goes to v2 on the
[main branch](https://github.com/coldsmirk/go-streams).

Fixes are welcome as pull requests against the `v1` branch. `task check` runs
the formatter, `go vet`, the linter and the tests.

## Acknowledgments

- Inspired by Java Stream API and Rust Iterator
- Built on the standard library's `iter.Seq` and `iter.Seq2`

## License

MIT. See [LICENSE](LICENSE).
