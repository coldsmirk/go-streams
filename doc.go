// Package streams provides lazy, type-safe stream processing built on
// [iter.Seq] and [iter.Seq2].
//
// A [Stream] wraps an iterator and evaluates lazily: intermediate operations
// build a pipeline, and nothing runs until a terminal operation consumes it.
//
//	names := streams.MapTo(
//		streams.Of(users...).Filter(func(u User) bool { return u.Active }),
//		func(u User) string { return u.Name },
//	).Collect()
//
// Operations that leave the element type alone are methods; operations that
// change it are package-level functions. Go 1.25 has no type parameters on
// methods, so a method cannot introduce a new type parameter and return a
// Stream of it — which is why [Stream.Map] is a method and [MapTo] is not.
//
// Streams are single-pass. A terminal operation consumes the underlying
// iterator, so a Stream should be traversed once; build a new one to traverse
// again. Sources that hold a file handle, such as [FromFileLines], are closed
// with [Using].
//
// Optional results are carried by [Optional] and fallible ones by [Result].
//
// This is v1, which is in maintenance and takes fixes only. Go 1.27 added
// generic methods, and v2 is the redesign that follows from them; see
// https://github.com/coldsmirk/go-streams for the current version.
package streams
