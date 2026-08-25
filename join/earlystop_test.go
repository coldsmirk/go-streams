package join

import (
	"iter"
	"testing"

	streams "github.com/coldsmirk/go-streams/v2"
	"github.com/stretchr/testify/assert"
)

// The iter contract says yield panics if it is called after returning false, so
// every operation that forwards elements must honour a false result and stop.
// These tests break out of each join after n elements, which panics if the
// early-return path is missing. n also selects the phase to break in: Right and
// Full emit the unmatched rows of b after the streamed side is exhausted, and
// that phase has to stop as well.

func breakAfter[T any](t *testing.T, name string, n int, s streams.Stream[T]) {
	t.Helper()
	seen := 0
	// A panic ends the check here, as the original recover-based harness did:
	// the count below describes a run that completed.
	if !assert.NotPanicsf(t, func() {
		for range iter.Seq[T](s) {
			if seen++; seen == n {
				break
			}
		}
	}, "%s: yielding after the consumer stopped", name) {
		return
	}
	assert.Equalf(t, n, seen, "%s: elements consumed before the break", name)
}

func breakAfter2[K, V any](t *testing.T, name string, n int, s streams.Stream2[K, V]) {
	t.Helper()
	seen := 0
	if !assert.NotPanicsf(t, func() {
		for range iter.Seq2[K, V](s) {
			if seen++; seen == n {
				break
			}
		}
	}, "%s: yielding after the consumer stopped", name) {
		return
	}
	assert.Equalf(t, n, seen, "%s: pairs consumed before the break", name)
}

func TestJoinsHonourEarlyStop(t *testing.T) {
	pair := func(a, b string) string { return a + "-" + b }

	breakAfter(t, "Inner", 1, Inner(srcA(), srcB(), showInner))
	breakAfter(t, "Left", 1, Left(srcA(), srcB(), showLeft))
	breakAfter(t, "Right", 1, Right(srcA(), srcB(), showRight))
	breakAfter(t, "Full", 1, Full(srcA(), srcB(), show))
	breakAfter(t, "Group", 1, Group(srcA(), srcB(), showGroup))
	breakAfter(t, "Inner over KeyBy", 1, Inner(onSrcA().KeyBy(initial), onSrcB().KeyBy(initial),
		func(_ string, a, b string) string { return pair(a, b) }))

	breakAfter2(t, "Semi", 1, Semi(srcA(), srcB()))
	breakAfter2(t, "Anti", 1, Anti(srcA(), disjoint()))

	// the third result of each is the row of a that b does not match, so
	// breaking there exercises the unmatched branch of the outer joins
	breakAfter(t, "Left at an unmatched row of a", 3, Left(srcA(), srcB(), showLeft))
	breakAfter(t, "Full at an unmatched row of a", 3, Full(srcA(), srcB(), show))
}

// Breaking part way through a join is not enough for Right and Full: each has a
// second phase that replays the unmatched rows of b, and a break in it, or on
// the last element before it, must stop the join just the same.
func TestOuterJoinsHonourEarlyStopWhileReplayingB(t *testing.T) {
	none := streams.Empty2[string, int]

	// with no rows in a, every result comes from the replay
	breakAfter(t, "Right replaying b", 1, Right(none(), srcB(), showRight))
	breakAfter(t, "Full replaying b", 1, Full(none(), srcB(), show))
	breakAfter(t, "Right replaying b, later", 2, Right(none(), srcB(), showRight))
	breakAfter(t, "Full replaying b, later", 2, Full(none(), srcB(), show))

	// Right yields four results from a and Full five, so breaking there lands
	// on the last element before the replay begins
	breakAfter(t, "Right at the last streamed element", 4, Right(srcA(), srcB(), showRight))
	breakAfter(t, "Full at the last streamed element", 5, Full(srcA(), srcB(), show))
	breakAfter(t, "Right in the replay", 5, Right(srcA(), srcB(), showRight))
	breakAfter(t, "Full in the replay", 6, Full(srcA(), srcB(), show))
}

// Group emits one result per key, so an early stop has to be honoured across
// the switch from the keys of a to the keys only b carries.
func TestGroupHonoursEarlyStopAcrossBothSides(t *testing.T) {
	breakAfter(t, "Group at the last key of a", 2, Group(srcA(), srcB(), showGroup))
	breakAfter(t, "Group at a key only b carries", 3, Group(srcA(), srcB(), showGroup))
	breakAfter(t, "Group with an empty a", 1,
		Group(streams.Empty2[string, int](), srcB(), showGroup))
}
