package collections

import (
	"iter"
	"testing"

	coll "github.com/coldsmirk/go-collections"
	streams "github.com/coldsmirk/go-streams/v2"
	"github.com/stretchr/testify/assert"
)

// The iter contract says yield panics if it is called after returning false, so
// every From function must stop as soon as the consumer does. These tests break
// out of each Stream after one element, which panics if the sequence underneath
// keeps yielding.

func breakAfterOne[T any](t *testing.T, name string, s streams.Stream[T]) {
	t.Helper()
	n := 0
	// A panic ends the check here, as the original recover-based harness did:
	// the count below describes a run that completed.
	if !assert.NotPanicsf(t, func() {
		for range iter.Seq[T](s) {
			n++
			break
		}
	}, "%s: yielding after the consumer stopped", name) {
		return
	}
	assert.Equalf(t, 1, n, "%s: elements consumed before the break", name)
}

func breakAfterOne2[K, V any](t *testing.T, name string, s streams.Stream2[K, V]) {
	t.Helper()
	n := 0
	if !assert.NotPanicsf(t, func() {
		for range iter.Seq2[K, V](s) {
			n++
			break
		}
	}, "%s: yielding after the consumer stopped", name) {
		return
	}
	assert.Equalf(t, 1, n, "%s: pairs consumed before the break", name)
}

func TestFromFunctionsHonourEarlyStop(t *testing.T) {
	breakAfterOne(t, "FromSet", FromSet(coll.NewHashSetFrom(1, 2, 3)))
	breakAfterOne(t, "FromSortedSet", FromSortedSet(coll.NewTreeSetFrom(byInt, 1, 2, 3)))
	breakAfterOne(t, "FromList array", FromList(coll.NewArrayListFrom(1, 2, 3)))
	breakAfterOne(t, "FromList linked", FromList(coll.NewLinkedListFrom(1, 2, 3)))
	breakAfterOne(t, "FromQueue", FromQueue(coll.NewArrayQueueFrom(1, 2, 3)))
	breakAfterOne(t, "FromStack", FromStack(coll.NewArrayStackFrom(1, 2, 3)))
	breakAfterOne(t, "FromDeque", FromDeque(coll.NewArrayDequeFrom(1, 2, 3)))
	breakAfterOne(t, "FromPriorityQueue", FromPriorityQueue(coll.NewPriorityQueueFrom(byInt, 1, 2, 3)))

	pairs := map[string]int{"a": 1, "b": 2, "c": 3}
	breakAfterOne2(t, "FromMap", FromMap(coll.NewHashMapFrom(pairs)))
	breakAfterOne2(t, "FromSortedMap", FromSortedMap(coll.NewTreeMapFrom(byString, pairs)))
}
