package join

import (
	"fmt"
	"slices"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	streams "github.com/coldsmirk/go-streams/v2"
)

// srcA and srcB are the two sides of the joins below. The key "x" carries two
// rows on each side, so every join has a duplicate key on both sides to work
// with; "y" is only in a and "z" is only in b.
func srcA() streams.Stream2[string, int] {
	return streams.Of("x", "y", "x").Zip(streams.Of(1, 2, 3))
}

func srcB() streams.Stream2[string, string] {
	return streams.Of("x", "z", "x").Zip(streams.Of("A", "B", "C"))
}

// disjoint shares no key with srcA.
func disjoint() streams.Stream2[string, string] {
	return streams.Of("p", "q").Zip(streams.Of("P", "Q"))
}

// show renders one result of any of the joins; an absent side is written "-".
// Every join formats through it, so their results are directly comparable.
func show(k string, l int, hasLeft bool, r string, hasRight bool) string {
	left, right := "-", "-"
	if hasLeft {
		left = strconv.Itoa(l)
	}
	if hasRight {
		right = r
	}
	return k + left + right
}

func showInner(k string, l int, r string) string { return show(k, l, true, r, true) }

func showLeft(k string, l int, r string, matched bool) string { return show(k, l, true, r, matched) }

func showRight(k string, l int, matched bool, r string) string { return show(k, l, matched, r, true) }

func showGroup(k string, l []int, r []string) string { return fmt.Sprintf("%s%v%v", k, l, r) }

func showPair(k string, v int) string { return k + strconv.Itoa(v) }

func TestInner(t *testing.T) {
	// a duplicate key on both sides yields the cartesian product for that key,
	// in the encounter order of a and then of b
	assert.Equal(t, []string{"x1A", "x1C", "x3A", "x3C"},
		Inner(srcA(), srcB(), showInner).Collect(), "Inner")
	assert.Empty(t, Inner(srcA(), disjoint(), showInner).Collect(), "Inner without a match")
	assert.Empty(t, Inner(streams.Empty2[string, int](), srcB(), showInner).Collect(), "Inner with an empty a")
	assert.Empty(t, Inner(srcA(), streams.Empty2[string, string](), showInner).Collect(), "Inner with an empty b")
}

func TestLeft(t *testing.T) {
	assert.Equal(t, []string{"x1A", "x1C", "y2-", "x3A", "x3C"},
		Left(srcA(), srcB(), showLeft).Collect(), "Left")
	assert.Equal(t, []string{"x1-", "y2-", "x3-"},
		Left(srcA(), disjoint(), showLeft).Collect(), "Left without a match")
	assert.Equal(t, []string{"x1-", "y2-", "x3-"},
		Left(srcA(), streams.Empty2[string, string](), showLeft).Collect(), "Left with an empty b")
	assert.Empty(t, Left(streams.Empty2[string, int](), srcB(), showLeft).Collect(), "Left with an empty a")
}

func TestRight(t *testing.T) {
	// the matched rows follow a, the unmatched rows of b follow at the end
	assert.Equal(t, []string{"x1A", "x1C", "x3A", "x3C", "z-B"},
		Right(srcA(), srcB(), showRight).Collect(), "Right")
	assert.Equal(t, []string{"p-P", "q-Q"},
		Right(srcA(), disjoint(), showRight).Collect(), "Right without a match")
	assert.Equal(t, []string{"x-A", "z-B", "x-C"},
		Right(streams.Empty2[string, int](), srcB(), showRight).Collect(), "Right with an empty a")
	assert.Empty(t, Right(srcA(), streams.Empty2[string, string](), showRight).Collect(), "Right with an empty b")
}

func TestFull(t *testing.T) {
	assert.Equal(t, []string{"x1A", "x1C", "y2-", "x3A", "x3C", "z-B"},
		Full(srcA(), srcB(), show).Collect(), "Full")
	assert.Equal(t, []string{"x1-", "y2-", "x3-", "p-P", "q-Q"},
		Full(srcA(), disjoint(), show).Collect(), "Full without a match")
	assert.Equal(t, []string{"x-A", "z-B", "x-C"},
		Full(streams.Empty2[string, int](), srcB(), show).Collect(), "Full with an empty a")
	assert.Equal(t, []string{"x1-", "y2-", "x3-"},
		Full(srcA(), streams.Empty2[string, string](), show).Collect(), "Full with an empty b")
	assert.Empty(t, Full(streams.Empty2[string, int](), streams.Empty2[string, string](), show).Collect(),
		"Full with two empty sides")
}

func TestOuterJoinsReplayBInEncounterOrder(t *testing.T) {
	// the keys interleave, so replaying the unmatched rows grouped by key would
	// reorder them
	b := func() streams.Stream2[string, string] {
		return streams.Of("p", "q", "p").Zip(streams.Of("P1", "Q", "P2"))
	}
	assert.Equal(t, []string{"p-P1", "q-Q", "p-P2"},
		Right(streams.Empty2[string, int](), b(), showRight).Collect(), "Right")
	assert.Equal(t, []string{"p-P1", "q-Q", "p-P2"},
		Full(streams.Empty2[string, int](), b(), show).Collect(), "Full")
}

func TestGroup(t *testing.T) {
	// keys in the order a first saw them, then the keys only b carries
	assert.Equal(t, []string{"x[1 3][A C]", "y[2][]", "z[][B]"},
		Group(srcA(), srcB(), showGroup).Collect(), "Group")
	assert.Equal(t, []string{"x[1 3][]", "y[2][]", "p[][P]", "q[][Q]"},
		Group(srcA(), disjoint(), showGroup).Collect(), "Group without a match")
	assert.Equal(t, []string{"x[][A C]", "z[][B]"},
		Group(streams.Empty2[string, int](), srcB(), showGroup).Collect(), "Group with an empty a")
	assert.Equal(t, []string{"x[1 3][]", "y[2][]"},
		Group(srcA(), streams.Empty2[string, string](), showGroup).Collect(), "Group with an empty b")
	assert.Empty(t, Group(streams.Empty2[string, int](), streams.Empty2[string, string](), showGroup).Collect(),
		"Group with two empty sides")
}

func TestGroupPassesNilForAKeyOnlyOneSideCarries(t *testing.T) {
	sides := Group(srcA(), srcB(), func(k string, l []int, r []string) string {
		switch k {
		case "y":
			assert.Nilf(t, r, "Group(%q) right", k)
		case "z":
			assert.Nilf(t, l, "Group(%q) left", k)
		}
		return k
	})
	assert.Equal(t, []string{"x", "y", "z"}, sides.Collect(), "Group keys")
}

func TestSemiAndAnti(t *testing.T) {
	// a row of a is admitted once however many rows of b carry its key
	assert.Equal(t, []string{"x1", "x3"}, Semi(srcA(), srcB()).Collapse(showPair).Collect(), "Semi")
	assert.Equal(t, []string{"y2"}, Anti(srcA(), srcB()).Collapse(showPair).Collect(), "Anti")

	assert.Empty(t, Semi(srcA(), disjoint()).Collapse(showPair).Collect(), "Semi without a match")
	assert.Equal(t, []string{"x1", "y2", "x3"},
		Anti(srcA(), disjoint()).Collapse(showPair).Collect(), "Anti without a match")

	empty := streams.Empty2[string, string]
	assert.Empty(t, Semi(srcA(), empty()).Collapse(showPair).Collect(), "Semi with an empty b")
	assert.Equal(t, []string{"x1", "y2", "x3"},
		Anti(srcA(), empty()).Collapse(showPair).Collect(), "Anti with an empty b")
	assert.Empty(t, Semi(streams.Empty2[string, int](), srcB()).Collapse(showPair).Collect(), "Semi with an empty a")
	assert.Empty(t, Anti(streams.Empty2[string, int](), srcB()).Collapse(showPair).Collect(), "Anti with an empty a")
}

// The derived-key joins: "cherry" has no match, and the key "a" is carried
// twice on each side.
func onSrcA() streams.Stream[string] { return streams.Of("apple", "banana", "avocado", "cherry") }

func onSrcB() streams.Stream[string] { return streams.Of("ant", "bee", "auk") }

func initial(s string) string { return s[:1] }

// Deriving keys with Stream.KeyBy is how two unkeyed streams reach any of the
// seven joins. Before KeyBy the package shipped On/SemiOn/AntiOn and nothing
// else, so a left, right, full or grouped join over unkeyed streams had no path
// through this package at all.
func TestUnkeyedStreamsJoinViaKeyBy(t *testing.T) {
	pair := func(a, b string) string { return a + "-" + b }

	t.Run("Inner", func(t *testing.T) {
		assert.Equal(t, []string{"apple-ant", "apple-auk", "banana-bee", "avocado-ant", "avocado-auk"},
			Inner(onSrcA().KeyBy(initial), onSrcB().KeyBy(initial),
				func(_ string, a, b string) string { return pair(a, b) }).Collect(), "inner")
		assert.Empty(t, Inner(streams.Of("cherry").KeyBy(initial), onSrcB().KeyBy(initial),
			func(_ string, a, b string) string { return pair(a, b) }).Collect(), "no match")
		assert.Empty(t, Inner(onSrcA().KeyBy(initial), streams.Empty[string]().KeyBy(initial),
			func(_ string, a, b string) string { return pair(a, b) }).Collect(), "empty b")
		assert.Empty(t, Inner(streams.Empty[string]().KeyBy(initial), onSrcB().KeyBy(initial),
			func(_ string, a, b string) string { return pair(a, b) }).Collect(), "empty a")
	})

	// These four had no derived-key form before KeyBy.
	t.Run("Left", func(t *testing.T) {
		assert.Equal(t,
			[]string{"apple-ant", "apple-auk", "banana-bee", "avocado-ant", "avocado-auk", "cherry-none"},
			Left(onSrcA().KeyBy(initial), onSrcB().KeyBy(initial),
				func(_ string, a, b string, ok bool) string {
					if !ok {
						return a + "-none"
					}
					return pair(a, b)
				}).Collect(), "left")
	})

	t.Run("Right", func(t *testing.T) {
		got := Right(streams.Of("cherry").KeyBy(initial), onSrcB().KeyBy(initial),
			func(_ string, a string, ok bool, b string) string {
				if !ok {
					return "none-" + b
				}
				return pair(a, b)
			}).Collect()
		assert.Equal(t, []string{"none-ant", "none-bee", "none-auk"}, got, "right")
	})

	t.Run("Full", func(t *testing.T) {
		got := Full(streams.Of("apple").KeyBy(initial), streams.Of("bee").KeyBy(initial),
			func(k string, a string, _ bool, b string, _ bool) string {
				return k + ":" + a + "/" + b
			}).Collect()
		assert.Equal(t, []string{"a:apple/", "b:/bee"}, got, "full")
	})

	t.Run("Group", func(t *testing.T) {
		got := Group(onSrcA().KeyBy(initial), onSrcB().KeyBy(initial),
			func(k string, a, b []string) string {
				return k + ":" + strconv.Itoa(len(a)) + "/" + strconv.Itoa(len(b))
			}).Collect()
		slices.Sort(got)
		assert.Equal(t, []string{"a:2/2", "b:1/1", "c:1/0"}, got, "group")
	})

	t.Run("Semi and Anti", func(t *testing.T) {
		assert.Equal(t, []string{"apple", "banana", "avocado"},
			Semi(onSrcA().KeyBy(initial), onSrcB().KeyBy(initial)).Values().Collect(), "semi")
		assert.Equal(t, []string{"cherry"},
			Anti(onSrcA().KeyBy(initial), onSrcB().KeyBy(initial)).Values().Collect(), "anti")
		assert.Equal(t, []string{"apple", "banana", "avocado", "cherry"},
			Anti(onSrcA().KeyBy(initial), streams.Empty[string]().KeyBy(initial)).Values().Collect(),
			"anti with an empty b")
	})
}

func TestJoinsStreamTheLeftSide(t *testing.T) {
	infinite := func() streams.Stream2[string, int] {
		return streams.Repeat("x", -1).Zip(streams.Iterate(1, func(i int) int { return i + 1 }))
	}
	assert.Equal(t, []string{"x1A", "x1C", "x2A"},
		Inner(infinite(), srcB(), showInner).Take(3).Collect(), "Inner")
	assert.Equal(t, []string{"x1A", "x1C", "x2A"},
		Left(infinite(), srcB(), showLeft).Take(3).Collect(), "Left")
	assert.Equal(t, []string{"x1", "x2"},
		Semi(infinite(), srcB()).Take(2).Collapse(showPair).Collect(), "Semi")
	assert.Equal(t, []string{"x1", "x2"},
		Anti(infinite(), disjoint()).Take(2).Collapse(showPair).Collect(), "Anti")

	// KeyBy is lazy, so an infinite unkeyed stream still streams.
	unkeyed := func() streams.Stream2[string, string] {
		return streams.Repeat("apple", -1).KeyBy(initial)
	}
	assert.Equal(t, []string{"ant", "auk", "ant"},
		Inner(unkeyed(), onSrcB().KeyBy(initial),
			func(_ string, _, b string) string { return b }).Take(3).Collect(),
		"Inner over KeyBy")
	assert.Equal(t, []string{"apple", "apple"},
		Semi(unkeyed(), onSrcB().KeyBy(initial)).Take(2).Values().Collect(),
		"Semi over KeyBy")
	assert.Equal(t, []string{"apple", "apple"},
		Anti(unkeyed(), streams.Empty[string]().KeyBy(initial)).Take(2).Values().Collect(),
		"Anti over KeyBy")
}

// The buffered side is drained once, not once per row of the streamed side.
func TestBIsDrainedOnce(t *testing.T) {
	drained := 0
	b := func() streams.Stream2[string, string] {
		return func(yield func(string, string) bool) {
			drained++
			for k, v := range srcB() {
				if !yield(k, v) {
					return
				}
			}
		}
	}
	for _, tc := range []struct {
		name string
		run  func(streams.Stream2[string, string])
	}{
		{"Inner", func(s streams.Stream2[string, string]) { Inner(srcA(), s, showInner).ForEach(func(string) {}) }},
		{"Left", func(s streams.Stream2[string, string]) { Left(srcA(), s, showLeft).ForEach(func(string) {}) }},
		{"Right", func(s streams.Stream2[string, string]) { Right(srcA(), s, showRight).ForEach(func(string) {}) }},
		{"Full", func(s streams.Stream2[string, string]) { Full(srcA(), s, show).ForEach(func(string) {}) }},
		{"Group", func(s streams.Stream2[string, string]) { Group(srcA(), s, showGroup).ForEach(func(string) {}) }},
		{"Semi", func(s streams.Stream2[string, string]) { Semi(srcA(), s).ForEach(func(string, int) {}) }},
		{"Anti", func(s streams.Stream2[string, string]) { Anti(srcA(), s).ForEach(func(string, int) {}) }},
	} {
		drained = 0
		tc.run(b())
		assert.Equalf(t, 1, drained, "%s drained b", tc.name)
	}
}

// Nothing is consumed until the result is iterated.
func TestJoinsAreLazy(t *testing.T) {
	touched := 0
	a := streams.Of("x", "y").Peek(func(string) { touched++ }).Zip(streams.Of(1, 2))
	s := Inner(a, srcB(), showInner)
	assert.Equal(t, 0, touched, "Inner consumed rows of a before iteration")
	s.ForEach(func(string) {})
	assert.Equal(t, 2, touched, "Inner consumed rows of a")
}
