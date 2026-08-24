# Migrating from go-streams v1

v2 is a deliberate break. Nothing is deprecated and nothing is aliased, because a
compatibility layer would preserve the very shapes v2 exists to remove.

The two libraries can be imported side by side while you migrate:

```go
import (
    v1 "github.com/coldsmirk/go-streams"
    streams "github.com/coldsmirk/go-streams/v2"
)
```

## What changed, and why

### Stream is now `iter.Seq`

```go
// v1
type Stream[T any] struct{ seq iter.Seq[T] }

// v2
type Stream[T any] iter.Seq[T]
```

A Stream converts to and from a standard library iterator with a plain type
conversion, and can be ranged over directly. `From`, `Seq()`, `From2` and
`Seq2()` are gone as concepts; `From` and `From2` survive only to infer the
element type, which a conversion cannot do.

| v1 | v2 |
| --- | --- |
| `streams.From(seq)` | `streams.From(seq)`, or `streams.Stream[T](seq)` |
| `s.Seq()` | `iter.Seq[T](s)` |
| `streams.FromSlice(xs)` | `streams.Of(xs...)` |
| `streams.FromMap(m)` | `streams.Pairs(m)` |
| `streams.FromChannel(ch)` | `streams.Chan(ch)` |
| — | `for v := range s` now works |

### Type-changing operations became methods

Go 1.27 allows type parameters on methods, so the free functions that existed
only to change the element type are now methods and the chain no longer breaks.

| v1 | v2 |
| --- | --- |
| `streams.MapTo(s, fn)` | `s.Map(fn)` |
| `streams.FlatMap(s, fn)` | `s.FlatMap(fn)` |
| `streams.FoldTo(s, init, fn)` | `s.Fold(init, fn)` |
| `streams.Scan(s, init, fn)` | `s.Scan(init, fn)` |
| `streams.GroupBy(s, key)` | `s.GroupBy(key)` |
| `streams.ToMap(s, fn)` | `s.ToMap(fn)` |
| `streams.AssociateBy(s, key)` / `streams.IndexBy(s, key)` | `s.IndexBy(key)` |
| `streams.DistinctBy(s, key)` | `s.DistinctBy(key)` |
| `streams.MinBy(s, key)` / `streams.MaxBy(s, key)` | `s.MinFunc(cmp)` / `s.MaxFunc(cmp)` |
| `streams.ParallelMap(s, fn, opts...)` | `s.ParallelMap(fn, opts...)` |
| `streams.MapValuesTo(s2, fn)` | `s2.MapValues(fn)` |
| `streams.MapKeysTo(s2, fn)` | `s2.MapKeys(fn)` |
| `streams.MapPairsTo(s2, fn)` | `s2.Collapse(fn)` |
| `streams.OptionalMap(o, fn)` | see *Optional is gone* below |

`s.Map(fn)` in v1 could only map `T` to `T`. In v2 the result type follows `fn`.

### Some operations stayed functions, on purpose

A method cannot constrain its receiver's type parameter, cannot return a type
that wraps it, and cannot destructure it. These are compiler rules, not
oversights, so the operations below remain package functions — exactly as
`slices.Sort` and `slices.Chunk` are functions rather than methods.

| Reason | Examples |
| --- | --- |
| Needs a constraint on the element type | `Sort`, `Min`, `Max`, `Distinct`, `Compact`, `Contains`, `Frequency`, `Sum`, `Product`, `Average` |
| Regroups the sequence | `Chunk`, `Window` |
| Destructures the element type | `Flatten` |
| Combines several streams | `Concat`, `Merge`, `Interleave` |

Each constrained function is paired with an unconstrained `Func` method, as in
the standard library: use `streams.Max(s)` when `T` is ordered, and
`s.MaxFunc(cmp)` when it is not.

### Optional is gone

Return values use the comma-ok form Go already uses for map lookups and type
assertions.

| v1 | v2 |
| --- | --- |
| `opt := s.First(); if opt.IsPresent() { use(opt.Get()) }` | `v, ok := s.First(); if ok { use(v) }` |
| `opt.GetOrElse(d)` | `v, ok := s.First(); if !ok { v = d }` |
| `opt.GetOrZero()` | `v, _ := s.First()` |
| `streams.OptionalMap(o, fn)` | map before the terminal call: `s.Map(fn).First()` |

### Result is gone; there is no error type

`slices`, `maps` and `iter` have no error handling at all, and v2 follows them.
Fallible work uses the standard `iter.Seq2[T, error]`.

| v1 | v2 |
| --- | --- |
| `streams.MapErrTo(s, fn)` | `streams.TryMap(s, fn)` — returns `iter.Seq2[R, error]` |
| `streams.CollectResults(rs)` | `streams.Try(seq)`, or `streams.Ok(seq)` to stream it |
| `streams.FilterOk(rs)` | `for v, err := range seq { if err == nil { ... } }` |
| `r.IsOk()` / `r.Unwrap()` | `v, err := ...; if err != nil { ... }` |

### Pair, Triple and Quad are gone

`Zip` yields a `Stream2`, and where three or more values must travel together
you pass a combiner function instead of receiving a tuple.

| v1 | v2 |
| --- | --- |
| `streams.Zip(a, b)` → `Stream[Pair[T,U]]` | `a.Zip(b)` → `Stream2[T,U]` |
| `streams.MapTo(streams.Zip(a,b), f)` | `a.ZipWith(b, f)` — no intermediate pair |
| `streams.ZipWithIndex(s)` | `s.Enumerate()` |
| `streams.SwapKeyValue(s2)` | `s2.Swap()` |
| `join.InnerJoin(...)` → `Stream[JoinResult[...]]` | `join.Inner(a, b, combine)` → `Stream[R]` |
| `streams.JoinBy(a, b, keyA, keyB)` | `join.Inner(a.KeyBy(keyA), b.KeyBy(keyB), combine)` |
| `streams.SemiJoinBy` / `AntiJoinBy` | `join.Semi` / `join.Anti` over `KeyBy`, then `.Values()` |

### Collector is gone

The `Collector` type and its 33 constructors were a port of a Java idiom. With
generic methods each one is a single call.

| v1 | v2 |
| --- | --- |
| `streams.CollectTo(s, streams.ToSliceCollector[T]())` | `s.Collect()` |
| `streams.CollectTo(s, streams.GroupingByCollector(key))` | `s.GroupBy(key)` |
| `streams.CollectTo(s, streams.SummingCollector[T]())` | `streams.Sum(s)` |
| `streams.CollectTo(s, streams.CountingCollector[T]())` | `s.Count()` |
| `streams.CollectTo(s, streams.JoiningCollector(sep))` | `strings.Join(s.Collect(), sep)` |
| `streams.PartitioningByCollector(pred)` | `s.Partition(pred)` |

### Renamed for internal consistency

| v1 | v2 | Reason |
| --- | --- | --- |
| `Limit` / `Skip` | `Take` / `Drop` | pairs with the `TakeWhile` / `DropWhile` that both versions have |
| `Distinct` | `Compact` (adjacent) and `Distinct` (global) | `Compact` now matches `slices.Compact` exactly |
| `DistinctUntilChanged` | `Compact` | same operation, standard library name |
| `DistinctUntilChangedBy` | `s.CompactFunc(eq)` | matches `slices.CompactFunc` |
| `SortedBy(s, key)` | `s.SortFunc(cmp)` | Go sorts by comparison; use `cmp.Compare(key(a), key(b))` |
| `Frequencies` | `Frequency` | the two were duplicates of each other |
| `MergeSortedN` / `MergeSortedNHeap` | `Merge` | the two were duplicate implementations |

### Moved to subpackages

Split by dependency weight, so the core package pulls in nothing beyond the
standard library.

| v1 file | v2 package |
| --- | --- |
| `io.go` | `github.com/coldsmirk/go-streams/v2/source` |
| `time.go` | `github.com/coldsmirk/go-streams/v2/temporal` |
| `join.go` | `github.com/coldsmirk/go-streams/v2/join` |
| `collections.go` | `github.com/coldsmirk/go-streams/v2/collections` |

### Removed without replacement

These were either one-line wrappers or outside the scope of a stream library.

- `Positive`, `Negative`, `NonZero`, `Abs`, `AbsFloat`, `Scale`, `Offset`, `Clamp` —
  write the `Filter` or `Map` directly; it is shorter and clearer at the call site.
- `Combinations`, `Permutations`, `CrossProduct`, `Cartesian`, `CartesianSelf` —
  combinatorics, not stream processing.
- `Zip3`, `ZipLongest`, `ZipLongestWith`, `Triples`, `Pairwise` — the tuple types
  they returned no longer exist. Use `ZipWith` with a combiner, or `Window(s, 2)`.
- The 16 `*Ctx` function variants — `WithContext` was always the single entry point;
  the rest duplicated it. Context now belongs to the `temporal` package operators,
  which take `ctx` as their first parameter.
- `Statistics`, `GetStatistics` — compute what you need with `Fold`.
- `TryCollect` — its purpose was converting a panic from the source into an
  error, and v2 has no panic-to-error conversion anywhere. Recover at the call
  site if you need it. Note that `Try(TryMap(s, fn))` is *not* a replacement: it
  forwards errors your function returns and lets a panic escape.

## A worked example

```go
// v1
result := streams.MapTo(
    streams.Distinct(streams.FromSlice(users)),
    func(u User) string { return u.Name },
)
opt := streams.MinBy(result, func(s string) int { return len(s) })
if opt.IsPresent() {
    use(opt.Get())
}

// v2
name, ok := streams.Of(users...).
    DistinctBy(func(u User) string { return u.ID }).
    Map(func(u User) string { return u.Name }).
    MinFunc(func(a, b string) int { return cmp.Compare(len(a), len(b)) })
if ok {
    use(name)
}
```
