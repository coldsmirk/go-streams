package join

import (
	"fmt"
	"slices"
	"strconv"
	"testing"

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

func eq[T comparable](t *testing.T, name string, got, want []T) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Errorf("%s = %v, want %v", name, got, want)
	}
}

func TestInner(t *testing.T) {
	// a duplicate key on both sides yields the cartesian product for that key,
	// in the encounter order of a and then of b
	eq(t, "Inner", Inner(srcA(), srcB(), showInner).Collect(),
		[]string{"x1A", "x1C", "x3A", "x3C"})
	eq(t, "Inner without a match", Inner(srcA(), disjoint(), showInner).Collect(), nil)
	eq(t, "Inner with an empty a", Inner(streams.Empty2[string, int](), srcB(), showInner).Collect(), nil)
	eq(t, "Inner with an empty b", Inner(srcA(), streams.Empty2[string, string](), showInner).Collect(), nil)
}

func TestLeft(t *testing.T) {
	eq(t, "Left", Left(srcA(), srcB(), showLeft).Collect(),
		[]string{"x1A", "x1C", "y2-", "x3A", "x3C"})
	eq(t, "Left without a match", Left(srcA(), disjoint(), showLeft).Collect(),
		[]string{"x1-", "y2-", "x3-"})
	eq(t, "Left with an empty b", Left(srcA(), streams.Empty2[string, string](), showLeft).Collect(),
		[]string{"x1-", "y2-", "x3-"})
	eq(t, "Left with an empty a", Left(streams.Empty2[string, int](), srcB(), showLeft).Collect(), nil)
}

func TestRight(t *testing.T) {
	// the matched rows follow a, the unmatched rows of b follow at the end
	eq(t, "Right", Right(srcA(), srcB(), showRight).Collect(),
		[]string{"x1A", "x1C", "x3A", "x3C", "z-B"})
	eq(t, "Right without a match", Right(srcA(), disjoint(), showRight).Collect(),
		[]string{"p-P", "q-Q"})
	eq(t, "Right with an empty a", Right(streams.Empty2[string, int](), srcB(), showRight).Collect(),
		[]string{"x-A", "z-B", "x-C"})
	eq(t, "Right with an empty b", Right(srcA(), streams.Empty2[string, string](), showRight).Collect(), nil)
}

func TestFull(t *testing.T) {
	eq(t, "Full", Full(srcA(), srcB(), show).Collect(),
		[]string{"x1A", "x1C", "y2-", "x3A", "x3C", "z-B"})
	eq(t, "Full without a match", Full(srcA(), disjoint(), show).Collect(),
		[]string{"x1-", "y2-", "x3-", "p-P", "q-Q"})
	eq(t, "Full with an empty a", Full(streams.Empty2[string, int](), srcB(), show).Collect(),
		[]string{"x-A", "z-B", "x-C"})
	eq(t, "Full with an empty b", Full(srcA(), streams.Empty2[string, string](), show).Collect(),
		[]string{"x1-", "y2-", "x3-"})
	eq(t, "Full with two empty sides",
		Full(streams.Empty2[string, int](), streams.Empty2[string, string](), show).Collect(), nil)
}

func TestOuterJoinsReplayBInEncounterOrder(t *testing.T) {
	// the keys interleave, so replaying the unmatched rows grouped by key would
	// reorder them
	b := func() streams.Stream2[string, string] {
		return streams.Of("p", "q", "p").Zip(streams.Of("P1", "Q", "P2"))
	}
	eq(t, "Right", Right(streams.Empty2[string, int](), b(), showRight).Collect(),
		[]string{"p-P1", "q-Q", "p-P2"})
	eq(t, "Full", Full(streams.Empty2[string, int](), b(), show).Collect(),
		[]string{"p-P1", "q-Q", "p-P2"})
}

func TestGroup(t *testing.T) {
	// keys in the order a first saw them, then the keys only b carries
	eq(t, "Group", Group(srcA(), srcB(), showGroup).Collect(),
		[]string{"x[1 3][A C]", "y[2][]", "z[][B]"})
	eq(t, "Group without a match", Group(srcA(), disjoint(), showGroup).Collect(),
		[]string{"x[1 3][]", "y[2][]", "p[][P]", "q[][Q]"})
	eq(t, "Group with an empty a", Group(streams.Empty2[string, int](), srcB(), showGroup).Collect(),
		[]string{"x[][A C]", "z[][B]"})
	eq(t, "Group with an empty b", Group(srcA(), streams.Empty2[string, string](), showGroup).Collect(),
		[]string{"x[1 3][]", "y[2][]"})
	eq(t, "Group with two empty sides",
		Group(streams.Empty2[string, int](), streams.Empty2[string, string](), showGroup).Collect(), nil)
}

func TestGroupPassesNilForAKeyOnlyOneSideCarries(t *testing.T) {
	sides := Group(srcA(), srcB(), func(k string, l []int, r []string) string {
		switch k {
		case "y":
			if r != nil {
				t.Errorf("Group(%q) right = %v, want nil", k, r)
			}
		case "z":
			if l != nil {
				t.Errorf("Group(%q) left = %v, want nil", k, l)
			}
		}
		return k
	})
	eq(t, "Group keys", sides.Collect(), []string{"x", "y", "z"})
}

func TestSemiAndAnti(t *testing.T) {
	// a row of a is admitted once however many rows of b carry its key
	eq(t, "Semi", Semi(srcA(), srcB()).Collapse(showPair).Collect(), []string{"x1", "x3"})
	eq(t, "Anti", Anti(srcA(), srcB()).Collapse(showPair).Collect(), []string{"y2"})

	eq(t, "Semi without a match", Semi(srcA(), disjoint()).Collapse(showPair).Collect(), nil)
	eq(t, "Anti without a match", Anti(srcA(), disjoint()).Collapse(showPair).Collect(),
		[]string{"x1", "y2", "x3"})

	empty := streams.Empty2[string, string]
	eq(t, "Semi with an empty b", Semi(srcA(), empty()).Collapse(showPair).Collect(), nil)
	eq(t, "Anti with an empty b", Anti(srcA(), empty()).Collapse(showPair).Collect(),
		[]string{"x1", "y2", "x3"})
	eq(t, "Semi with an empty a", Semi(streams.Empty2[string, int](), srcB()).Collapse(showPair).Collect(), nil)
	eq(t, "Anti with an empty a", Anti(streams.Empty2[string, int](), srcB()).Collapse(showPair).Collect(), nil)
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
		eq(t, "inner", Inner(onSrcA().KeyBy(initial), onSrcB().KeyBy(initial),
			func(_ string, a, b string) string { return pair(a, b) }).Collect(),
			[]string{"apple-ant", "apple-auk", "banana-bee", "avocado-ant", "avocado-auk"})
		eq(t, "no match", Inner(streams.Of("cherry").KeyBy(initial), onSrcB().KeyBy(initial),
			func(_ string, a, b string) string { return pair(a, b) }).Collect(), nil)
		eq(t, "empty b", Inner(onSrcA().KeyBy(initial), streams.Empty[string]().KeyBy(initial),
			func(_ string, a, b string) string { return pair(a, b) }).Collect(), nil)
		eq(t, "empty a", Inner(streams.Empty[string]().KeyBy(initial), onSrcB().KeyBy(initial),
			func(_ string, a, b string) string { return pair(a, b) }).Collect(), nil)
	})

	// These four had no derived-key form before KeyBy.
	t.Run("Left", func(t *testing.T) {
		eq(t, "left", Left(onSrcA().KeyBy(initial), onSrcB().KeyBy(initial),
			func(_ string, a, b string, ok bool) string {
				if !ok {
					return a + "-none"
				}
				return pair(a, b)
			}).Collect(),
			[]string{"apple-ant", "apple-auk", "banana-bee", "avocado-ant", "avocado-auk", "cherry-none"})
	})

	t.Run("Right", func(t *testing.T) {
		got := Right(streams.Of("cherry").KeyBy(initial), onSrcB().KeyBy(initial),
			func(_ string, a string, ok bool, b string) string {
				if !ok {
					return "none-" + b
				}
				return pair(a, b)
			}).Collect()
		eq(t, "right", got, []string{"none-ant", "none-bee", "none-auk"})
	})

	t.Run("Full", func(t *testing.T) {
		got := Full(streams.Of("apple").KeyBy(initial), streams.Of("bee").KeyBy(initial),
			func(k string, a string, ha bool, b string, hb bool) string {
				return k + ":" + a + "/" + b
			}).Collect()
		eq(t, "full", got, []string{"a:apple/", "b:/bee"})
	})

	t.Run("Group", func(t *testing.T) {
		got := Group(onSrcA().KeyBy(initial), onSrcB().KeyBy(initial),
			func(k string, a, b []string) string {
				return k + ":" + strconv.Itoa(len(a)) + "/" + strconv.Itoa(len(b))
			}).Collect()
		slices.Sort(got)
		eq(t, "group", got, []string{"a:2/2", "b:1/1", "c:1/0"})
	})

	t.Run("Semi and Anti", func(t *testing.T) {
		eq(t, "semi", Semi(onSrcA().KeyBy(initial), onSrcB().KeyBy(initial)).Values().Collect(),
			[]string{"apple", "banana", "avocado"})
		eq(t, "anti", Anti(onSrcA().KeyBy(initial), onSrcB().KeyBy(initial)).Values().Collect(),
			[]string{"cherry"})
		eq(t, "anti with an empty b",
			Anti(onSrcA().KeyBy(initial), streams.Empty[string]().KeyBy(initial)).Values().Collect(),
			[]string{"apple", "banana", "avocado", "cherry"})
	})
}

func TestJoinsStreamTheLeftSide(t *testing.T) {
	infinite := func() streams.Stream2[string, int] {
		return streams.Repeat("x", -1).Zip(streams.Iterate(1, func(i int) int { return i + 1 }))
	}
	eq(t, "Inner", Inner(infinite(), srcB(), showInner).Take(3).Collect(),
		[]string{"x1A", "x1C", "x2A"})
	eq(t, "Left", Left(infinite(), srcB(), showLeft).Take(3).Collect(),
		[]string{"x1A", "x1C", "x2A"})
	eq(t, "Semi", Semi(infinite(), srcB()).Take(2).Collapse(showPair).Collect(),
		[]string{"x1", "x2"})
	eq(t, "Anti", Anti(infinite(), disjoint()).Take(2).Collapse(showPair).Collect(),
		[]string{"x1", "x2"})

	// KeyBy is lazy, so an infinite unkeyed stream still streams.
	unkeyed := func() streams.Stream2[string, string] {
		return streams.Repeat("apple", -1).KeyBy(initial)
	}
	eq(t, "Inner over KeyBy",
		Inner(unkeyed(), onSrcB().KeyBy(initial),
			func(_ string, a, b string) string { return b }).Take(3).Collect(),
		[]string{"ant", "auk", "ant"})
	eq(t, "Semi over KeyBy",
		Semi(unkeyed(), onSrcB().KeyBy(initial)).Take(2).Values().Collect(),
		[]string{"apple", "apple"})
	eq(t, "Anti over KeyBy",
		Anti(unkeyed(), streams.Empty[string]().KeyBy(initial)).Take(2).Values().Collect(),
		[]string{"apple", "apple"})
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
		if drained != 1 {
			t.Errorf("%s drained b %d times, want 1", tc.name, drained)
		}
	}
}

// Nothing is consumed until the result is iterated.
func TestJoinsAreLazy(t *testing.T) {
	touched := 0
	a := streams.Of("x", "y").Peek(func(string) { touched++ }).Zip(streams.Of(1, 2))
	s := Inner(a, srcB(), showInner)
	if touched != 0 {
		t.Errorf("Inner consumed %d rows of a before iteration", touched)
	}
	s.ForEach(func(string) {})
	if touched != 2 {
		t.Errorf("Inner consumed %d rows of a, want 2", touched)
	}
}
