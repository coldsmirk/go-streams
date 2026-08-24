# go-streams

[![Go Reference](https://pkg.go.dev/badge/github.com/coldsmirk/go-streams/v2.svg)](https://pkg.go.dev/github.com/coldsmirk/go-streams/v2)
[![Go Report Card](https://goreportcard.com/badge/github.com/coldsmirk/go-streams/v2)](https://goreportcard.com/report/github.com/coldsmirk/go-streams/v2)
[![Build Status](https://github.com/coldsmirk/go-streams/actions/workflows/test.yml/badge.svg)](https://github.com/coldsmirk/go-streams/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/coldsmirk/go-streams/branch/main/graph/badge.svg)](https://codecov.io/gh/coldsmirk/go-streams)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Lazy, composable sequences for Go 1.27, built directly on `iter.Seq`.

```sh
go get github.com/coldsmirk/go-streams/v2
```

```go
import streams "github.com/coldsmirk/go-streams/v2"

names := streams.Of(users...).
    Filter(func(u User) bool { return u.Active }).
    Map(func(u User) string { return u.Name }).
    Map(strings.ToUpper).
    Collect()
```

That chain changes the element type twice without leaving the chain. In v1 it
could not: Go had no type parameters on methods, so every type-changing
operation had to be a free function and the pipeline turned inside out. Go 1.27
added generic methods, and v2 is the redesign that follows from them.

Migrating from v1: see [MIGRATION.md](MIGRATION.md). v1 remains available at
`github.com/coldsmirk/go-streams` and is maintained on the
[`v1` branch](https://github.com/coldsmirk/go-streams/tree/v1).

## Design

**A Stream is an iterator, not a wrapper around one.**

```go
type Stream[T any] iter.Seq[T]
type Stream2[K, V any] iter.Seq2[K, V]
```

Because the underlying type is the standard library's, a Stream converts both
ways at no cost and can be ranged over directly:

```go
for v := range s { ... }                     // no adapter needed
sorted := slices.Sorted(iter.Seq[int](s))    // out to the standard library
s := streams.From(maps.Keys(m))              // in from it
```

**Methods and functions split where the compiler makes them split.**

A method is used when the operation needs nothing from the element type beyond
`any`, and the result type is either unchanged or determined by a function the
caller supplies. Everything else is a package function — for the same three
reasons `slices.Sort` and `slices.Chunk` are functions:

| Reason | Examples |
| --- | --- |
| Needs a constraint on the element type | `Sort`, `Max`, `Distinct`, `Sum` |
| Regroups the sequence | `Chunk`, `Window` |
| Destructures the element type | `Flatten` |
| Combines several streams | `Concat`, `Merge`, `Interleave` |

As in the standard library, each constrained function is paired with an
unconstrained `Func` method:

```go
top, ok := streams.Max(numbers)                  // T is cmp.Ordered
oldest, ok := people.MaxFunc(func(a, b Person) int {
    return cmp.Compare(a.Age, b.Age)             // T is anything
})
```

**No Optional, no Result, no Pair.**

Optional results use the comma-ok form Go already uses everywhere else.
Fallible work uses the standard `iter.Seq2[T, error]` — `slices`, `maps` and
`iter` have no error abstraction, and neither does this package. Pairs are
`Stream2`, and three or more values travel through a combiner function:

```go
v, ok := s.First()
lines, err := streams.Try(streams.TryMap(s, parse))
pairs := names.Zip(ages)                              // Stream2[string, int]
sums  := a.ZipWith(b, func(x, y int) int { return x + y })
```

## Core API

Full documentation and runnable examples:
[pkg.go.dev](https://pkg.go.dev/github.com/coldsmirk/go-streams/v2), or
`go doc -all github.com/coldsmirk/go-streams/v2`.

**Constructing** — `Of`, `From`, `From2`, `Pairs`, `Chan`, `ChanContext`,
`Range`, `Repeat`, `Iterate`, `Generate`, `Empty`, `Empty2`

**Transforming** — `Filter`, `Map`, `FlatMap`, `Scan`, `DistinctBy`, `Take`,
`Drop`, `TakeWhile`, `DropWhile`, `SortFunc`, `SortStableFunc`, `CompactFunc`,
`Reverse`, `Peek`, `KeyBy`, `Zip`, `ZipWith`, `Enumerate`

**Consuming** — `Collect`, `ForEach`, `Count`, `Fold`, `Reduce`, `First`,
`Last`, `Find`, `Any`, `All`, `MinFunc`, `MaxFunc`, `GroupBy`, `IndexBy`,
`ToMap`, `Partition`

**Package functions** — `Sort`, `Min`, `Max`, `Compact`, `Distinct`,
`Contains`, `Frequency`, `Sum`, `Product`, `Average`, `Chunk`, `Window`,
`Flatten`, `Concat`, `Merge`, `Interleave`, `Cycle`, `TryMap`, `Ok`, `Try`

**Stream2** — `Keys`, `Values`, `Filter`, `MapKeys`, `MapValues`, `Collapse`,
`Swap`, `Take`, `ForEach`, `Count`, `Fold`

**Parallel** — `ParallelMap`, `ParallelFilter`, `ParallelForEach`, configured
with `WithConcurrency` and `Unordered`. Results keep their input order unless
`Unordered` is given.

```go
squares := streams.Range(1, 1000).
    ParallelMap(expensive, streams.WithConcurrency(8)).
    Collect()
```

## Subpackages

The core package imports nothing outside the standard library. Everything with a
heavier dependency lives in its own package. Fallible sources return the standard
`iter.Seq2[T, error]`; `streams.Try` collects one, stopping at the first error.

### `.../v2/source` — readers, files, CSV

```go
// streams the file; use streams.Try instead only when you want it all in a slice
lines, readErr := streams.Ok(source.LinesFile("access.log"))
n := lines.Filter(isError).Count()
if err := readErr(); err != nil {
    return err
}

for rec, err := range source.Records(r) {   // first row is the header
    if err != nil { return err }
    use(rec["email"])
}

err := source.WriteCSVFile("out.csv", rows)
```

`Lines`, `LinesFile`, `StringLines`, `Bytes`, `Runes`, `CSV`, `CSVFile`, `TSV`,
`TSVFile`, `Delimited`, `Records`, `RecordsFile`, `Keyed`, `File`, `WriteLines`,
`WriteFile`, `WriteCSV`, `WriteCSVFile`, and the `Record` type.

File sources open and close the file themselves — unlike v1, there is nothing for
the caller to `Close`. An error ends the sequence.

The three concerns compose rather than multiply: `Delimited` picks the
delimiter, `Keyed` applies a header row, and `File` handles opening. A
tab-separated file with a header is `Keyed(TSVFile(path))`, and a format this
package does not name is `File(path, parse)`.

### `.../v2/temporal` — time-based operators

```go
recent := temporal.Throttle(ctx, events, 100*time.Millisecond)
batches := temporal.Tumbling(ctx, events, time.Second)
```

`Throttle`, `Debounce`, `Sample`, `Delay`, `RateLimit`, `Tumbling`, `Sliding`,
`Session`, `Timeout`, `Interval`, `Stamp`.

Every operator that waits on the clock takes `ctx` first and stops promptly when
it is done; `Stamp` never waits, so it takes none. `Stamp` returns
`Stream2[time.Time, T]`, so there is no timestamp wrapper type.

These operators read the source on a goroutine of their own, and Go cannot
interrupt a goroutine parked inside a caller-supplied iterator. So feed them a
source that ends when your context does, and shutdown is bounded:

```go
events := streams.ChanContext(ctx, ch)        // not streams.Chan(ch)
recent := temporal.Throttle(ctx, events, 100*time.Millisecond)
```

With `streams.Chan` over a channel that goes quiet and is never closed, the
reader stays parked until the channel produces or closes. Timers are always
released either way. The package documentation states this precisely.

### `.../v2/join` — relational joins

```go
rows := join.Inner(orders, customers, func(id OrderID, o Order, c Customer) Row {
    return Row{Order: o, Customer: c}
})

// two unkeyed streams: derive the keys first, and every join accepts them
rows = join.Left(orders.KeyBy(byCustomer), customers.KeyBy(byID), combine)
```

`Inner`, `Left`, `Right`, `Full`, `Group`, `Semi`, `Anti`.

Every join takes a combiner and returns what it produces, so there are no
`JoinResult` types. Outer joins pass a presence flag alongside the possibly-zero
value. Each doc comment states which side is buffered.

### `.../v2/collections` — go-collections bridge

```go
set := collections.ToHashSet(streams.Of(ids...))   // coll.Set[int]
back := collections.FromSet(set).Collect()

sorted := collections.ToTreeSet(streams.Of(ids...), cmp.Compare)
top, _ := collections.FromSortedSet(sorted).First()
```

`FromSet`, `FromSortedSet`, `FromList`, `FromQueue`, `FromStack`, `FromDeque`,
`FromPriorityQueue`, `FromMap`, `FromSortedMap`, and `ToHashSet`, `ToTreeSet`,
`ToArrayList`, `ToLinkedList`, `ToHashMap`, `ToTreeMap`.

Every `From*` iterates the live collection lazily rather than copying it.

## Semantics worth knowing

**Errors live at the edges.** The core has no error type, matching `slices`,
`maps` and `iter`. A fallible source returns the standard
`iter.Seq2[T, error]`; `streams.Ok` turns one into a Stream plus an error
accessor, in the shape of `bufio.Scanner`, and `streams.Try` collects one into
a slice when buffering is fine.

**Streams are single-pass.** A terminal operation consumes the underlying
iterator. Traverse a Stream once; build a new one to traverse again.

**Laziness is real, and short-circuiting works.** `First`, `Any`, `All`, `Find`
and `Take` stop the source as soon as they can:

```go
// evaluates the mapping three times, not a thousand
v, _ := streams.Range(0, 1000).Map(expensive).Filter(pred).First()
```

**The zero Stream is not valid.** Use `Empty[T]()`. A nil iterator panics when
ranged over, exactly as a nil `iter.Seq` does.

**Some operations must buffer.** `Sort`, `SortFunc`, `SortStableFunc`,
`Reverse` and `Cycle` read the whole sequence, so they cannot be used on an
infinite stream. Everything else streams. Each doc comment says which.

**Early termination is honoured everywhere.** The `iter` contract panics if
`yield` is called after it returns false; every operation in this package stops
instead, and every one of them has a test that proves it.

## Requirements

Go 1.27 or later. Generic methods are load-bearing; the package will not build
on an earlier toolchain.

## Contributing

Issues and pull requests are welcome. `task check` runs the formatter, `go vet`,
the linter and the tests — the same set CI runs, plus the `modernize` analyzer.

Two conventions this package holds to:

- **Every operation gets an early-termination test.** The `iter` contract panics
  if `yield` is called after it returns false, so stopping correctly is a
  correctness requirement, not a nicety.
- **Examples are runnable.** They belong in `example_test.go` as `Example`
  functions with an `// Output:` block, not in this file, so that the compiler
  and CI check them.

## Acknowledgments

The shape of the API follows the standard library's `slices`, `maps` and `iter`
packages. The operator vocabulary owes to Java's Stream and Rust's Iterator.

## License

MIT. See [LICENSE](LICENSE).
