// Package temporal provides time-based stream operators.
//
// The operators shape a streams.Stream against the clock rather than against
// element positions: they pace, drop, delay, or group elements by the time at
// which those elements pass through. An element's time is the moment the
// operator receives it from the source, which is later than the moment the
// source produced it whenever the consumer is slow.
//
// Every operator that waits on the clock takes a context.Context as its first
// parameter and ends as soon as that context is done, emitting nothing further:
// a pending element, a partly filled window and an open session are all
// discarded. Timeout is the one exception, because it is the one operator with
// an error slot to report the cause through.
//
// Every operator reads the source in a separate goroutine, so that waiting for
// the next element and watching the clock can be one select. Timers are always
// released: each is stopped by a defer in the operator's own goroutine.
//
// Whether that reader goroutine is released on cancellation is decided by the
// source, not by these operators. Go cannot interrupt a goroutine blocked inside
// a caller-supplied iterator, so an operator can only reclaim its reader at a
// point where the source hands control back.
//
// Give them a source that ends when the context does and shutdown is bounded:
// streams.ChanContext is the one to reach for, and any source that selects on
// ctx.Done behaves the same way. Pass the same context to the source and to the
// operator.
//
// A source that does neither -- streams.Chan over a channel that goes quiet and
// is never closed is the case to watch for -- leaves the reader parked inside
// it, holding one element, until the channel produces or closes. That is the
// normal state of a live feed idle at the moment its consumer stops, so it is
// worth getting right.
//
// Every duration argument must be positive, as must the element count of
// RateLimit; a value that is not panics, as streams.Chunk does for a
// non-positive chunk size. The panic happens when the operator is called, not
// when the resulting Stream is iterated.
//
// This package has no error abstraction. Timeout, the only fallible operation,
// returns the standard iter.Seq2[T, error].
package temporal
