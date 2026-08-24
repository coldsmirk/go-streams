// Package streams provides lazy, composable sequences.
//
// A [Stream] is defined as [iter.Seq], so it converts to and from a standard
// library iterator at no cost and may be ranged over directly:
//
//	for v := range s { ... }
//
// Operations split along the same line the standard library draws in [slices].
// A method is used when the operation needs nothing from the element type
// beyond any, and the result type is either unchanged or determined by a
// function the caller supplies:
//
//	names := streams.Of(users...).
//		Filter(func(u User) bool { return u.Active }).
//		Map(func(u User) string { return u.Name }).
//		Collect()
//
// A package-level function is used when the operation constrains the element
// type, regroups the sequence, or destructures the element type. As in
// [slices], a constrained function is paired with an unconstrained Func method:
// [Max] requires cmp.Ordered, while [Stream.MaxFunc] takes a comparison
// function and works for any element type.
//
// The zero Stream is not valid; use [Empty] for an empty sequence.
//
// Streams are lazy and single-pass. Each terminal operation consumes the
// underlying iterator, so a Stream should be traversed once. Construct a new
// Stream to iterate again.
//
// This package has no error abstraction, matching [slices], [maps] and [iter].
// Fallible sequences use the standard iter.Seq2[T, error]; see [TryMap] and
// [Try] for the bridge.
package streams
