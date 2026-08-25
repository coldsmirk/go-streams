package temporal

import (
	"iter"
	"testing"
	"time"

	streams "github.com/coldsmirk/go-streams/v2"
	"github.com/stretchr/testify/assert"
)

// The iter contract says yield panics if it is called after returning false.
// Every operation that forwards elements must therefore honour a false result
// and stop. These tests break out of each operation after one element, which
// panics if the early-return path is missing.
//
// The operators driven by a timer are given a source that keeps producing, so
// that the element the consumer breaks on comes from the timer rather than from
// the end of the source, which is the path a return would otherwise skip.

func breakAfterOne[T any](t *testing.T, name string, s streams.Stream[T]) {
	t.Helper()
	n := 0
	ok := assert.NotPanicsf(t, func() {
		for range iter.Seq[T](s) {
			n++
			break
		}
	}, "%s: yielding after the consumer stopped", name)
	if !ok {
		return
	}
	assert.Equalf(t, 1, n, "%s: consumed %d elements before the break, want 1", name, n)
}

func breakAfterOne2[K, V any](t *testing.T, name string, seq iter.Seq2[K, V]) {
	t.Helper()
	n := 0
	ok := assert.NotPanicsf(t, func() {
		for range seq {
			n++
			break
		}
	}, "%s: yielding after the consumer stopped", name)
	if !ok {
		return
	}
	assert.Equalf(t, 1, n, "%s: consumed %d pairs before the break, want 1", name, n)
}

func TestOperatorsHonourEarlyStop(t *testing.T) {
	const unit = 20 * time.Millisecond
	ctx := t.Context()
	fast := func() streams.Stream[int] { return paced(unit / 4) }
	slow := func() streams.Stream[int] { return paced(3 * unit) }
	src := func() streams.Stream[int] { return streams.Of(1, 2, 3, 4, 5) }

	breakAfterOne(t, "Throttle", Throttle(ctx, src(), unit))
	breakAfterOne(t, "Debounce", Debounce(ctx, slow(), unit))
	breakAfterOne(t, "Sample", Sample(ctx, fast(), unit))
	breakAfterOne(t, "Delay", Delay(ctx, src(), unit))
	breakAfterOne(t, "RateLimit", RateLimit(ctx, src(), 1, unit))
	breakAfterOne(t, "Tumbling", Tumbling(ctx, fast(), unit))
	breakAfterOne(t, "Sliding", Sliding(ctx, fast(), 2*unit, unit))
	breakAfterOne(t, "Session", Session(ctx, slow(), unit))
	breakAfterOne(t, "Interval", Interval(ctx, unit))

	breakAfterOne2(t, "Timeout", Timeout(ctx, src(), time.Hour))
	breakAfterOne2(t, "Stamp", iter.Seq2[time.Time, int](Stamp(src())))
}
