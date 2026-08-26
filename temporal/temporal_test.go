package temporal

import (
	"context"
	"runtime"
	"slices"
	"testing"
	"time"

	streams "github.com/coldsmirk/go-streams/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The tests below are written to survive a loaded machine: a slow scheduler may
// only make an operator emit later than it meant to, never earlier. Timing is
// therefore asserted with lower bounds, and the few upper bounds use a duration
// far longer than the operator needs, so that only a stall of several hundred
// milliseconds could break one. Where an exact grouping is asserted, the burst
// that forms it is fed through an unbuffered channel and the next burst waits
// for the operator to emit, so the assertion turns on the operator rather than
// on the clock.

// paced returns an infinite Stream of 0, 1, 2, ... one element every d. It
// stops at the next element after the consumer does.
func paced(d time.Duration) streams.Stream[int] {
	return func(yield func(int) bool) {
		for i := 0; ; i++ {
			time.Sleep(d)
			if !yield(i) {
				return
			}
		}
	}
}

func countGoroutines() int {
	runtime.GC()
	return runtime.NumGoroutine()
}

func increasing(t *testing.T, name string, got []int) {
	t.Helper()
	assert.IsIncreasingf(t, got, "%s = %v, want strictly increasing", name, got)
}

func flatten(windows [][]int) []int {
	var out []int
	for _, w := range windows {
		out = append(out, w...)
	}
	return out
}

func TestThrottleEmitsEveryElementPaced(t *testing.T) {
	const interval = 30 * time.Millisecond

	start := time.Now()
	got := Throttle(t.Context(), streams.Of(1, 2, 3, 4), interval).Collect()
	elapsed := time.Since(start)

	want := []int{1, 2, 3, 4}
	assert.Equal(t, want, got, "Throttle")
	// The first element is free, the other three each wait an interval.
	atLeast := 3*interval - 10*time.Millisecond
	assert.GreaterOrEqualf(t, elapsed, atLeast, "Throttle took %v, want at least %v", elapsed, atLeast)
}

func TestThrottleEmitsTheFirstElementImmediately(t *testing.T) {
	// The interval is far longer than the bound, so only a stall of most of a
	// second could make this fail.
	start := time.Now()
	got, ok := Throttle(t.Context(), streams.Of(7), time.Second).First()
	elapsed := time.Since(start)

	assert.True(t, ok, "Throttle first")
	assert.Equal(t, 7, got, "Throttle first")
	assert.LessOrEqualf(t, elapsed, 500*time.Millisecond, "Throttle held the first element for %v", elapsed)
}

func TestDelayShiftsTheWholeStream(t *testing.T) {
	// Three elements produced at once are delayed together, not one after
	// another, so the whole run takes about one d. The upper bound is what
	// asserts the overlap — a serial wait would take at least 3*d — and it
	// leaves a full d of slack for a loaded machine.
	const d = 400 * time.Millisecond

	start := time.Now()
	got := Delay(t.Context(), streams.Of(1, 2, 3), d).Collect()
	elapsed := time.Since(start)

	want := []int{1, 2, 3}
	assert.Equal(t, want, got, "Delay")
	atLeast := d - 10*time.Millisecond
	assert.GreaterOrEqualf(t, elapsed, atLeast, "Delay took %v, want at least %v", elapsed, atLeast)
	atMost := 2 * d
	assert.Lessf(t, elapsed, atMost, "Delay took %v, want under %v: the waits must overlap", elapsed, atMost)
}

func TestDelayPreservesSpacing(t *testing.T) {
	// The second element arrives gap after the first and must be emitted about
	// gap after it too. A serial wait would stretch the spacing to d, and
	// holding both elements to emit together would collapse it to nothing, so
	// the bounds separate the shift from both failure shapes. The channel is
	// unbuffered so each send synchronises with the operator's receive: the
	// sleep cannot start before the first element is taken, and a slow start
	// cannot leave both elements buffered to arrive with no gap at all.
	const (
		gap = 250 * time.Millisecond
		d   = 500 * time.Millisecond
	)
	ch := make(chan int)
	go func() {
		ch <- 1
		time.Sleep(gap)
		ch <- 2
		close(ch)
	}()

	start := time.Now()
	var emitted []time.Duration
	for range Delay(t.Context(), streams.Chan(ch), d) {
		emitted = append(emitted, time.Since(start))
	}

	require.Lenf(t, emitted, 2, "Delay emitted %d elements, want 2", len(emitted))
	spacing := emitted[1] - emitted[0]
	assert.GreaterOrEqualf(t, spacing, 100*time.Millisecond,
		"emissions %v apart, want about the arrival gap of %v", spacing, gap)
	assert.Lessf(t, spacing, 400*time.Millisecond,
		"emissions %v apart, want about the arrival gap of %v: the waits must overlap", spacing, gap)
}

// The end of the Stream is shifted like the elements: a source that emits and
// then lingers before finishing keeps the delayed Stream open for d past the
// moment it finished, not merely d past its last element.
func TestDelayShiftsTheEndOfTheStream(t *testing.T) {
	const (
		d      = 200 * time.Millisecond
		linger = 400 * time.Millisecond
	)
	src := streams.Stream[int](func(yield func(int) bool) {
		if !yield(1) {
			return
		}
		time.Sleep(linger)
	})

	start := time.Now()
	got := Delay(t.Context(), src, d).Collect()
	elapsed := time.Since(start)

	assert.Equal(t, []int{1}, got, "Delay")
	atLeast := linger + d - 10*time.Millisecond
	assert.GreaterOrEqualf(t, elapsed, atLeast,
		"Delay ended after %v, want at least %v: the end must be shifted too", elapsed, atLeast)
}

func TestRateLimitPacesBeyondTheInitialBurst(t *testing.T) {
	const (
		n        = 2
		per      = 100 * time.Millisecond
		emission = per / n
	)

	start := time.Now()
	got := RateLimit(t.Context(), streams.Range(0, 6), n, per).Collect()
	elapsed := time.Since(start)

	want := []int{0, 1, 2, 3, 4, 5}
	assert.Equal(t, want, got, "RateLimit")
	// Two elements ride the initial burst; the other four are paced.
	atLeast := 3 * emission
	assert.GreaterOrEqualf(t, elapsed, atLeast, "RateLimit took %v, want at least %v", elapsed, atLeast)
}

func TestRateLimitEmitsTheInitialBurstImmediately(t *testing.T) {
	start := time.Now()
	got := RateLimit(t.Context(), streams.Of(1, 2, 3), 3, 3*time.Second).Collect()
	elapsed := time.Since(start)

	want := []int{1, 2, 3}
	assert.Equal(t, want, got, "RateLimit")
	assert.LessOrEqualf(t, elapsed, time.Second, "RateLimit held the initial burst for %v", elapsed)
}

// An n above per's nanosecond count would truncate the emission interval to
// zero and stop the virtual clock; the clamp to one nanosecond keeps the
// limiter live. A rate that high is unreachable, so the elements pass at once
// — the test pins that they pass at all.
func TestRateLimitSurvivesASubNanosecondEmissionInterval(t *testing.T) {
	start := time.Now()
	got := RateLimit(t.Context(), streams.Range(0, 4), 100, 50*time.Nanosecond).Collect()
	elapsed := time.Since(start)

	assert.Equal(t, []int{0, 1, 2, 3}, got, "RateLimit")
	assert.LessOrEqualf(t, elapsed, time.Second, "RateLimit stalled for %v on a truncated emission interval", elapsed)
}

// The clamp is only observable through the derived interval — nanosecond
// pacing cannot be told apart from no pacing by timing — so the derivation is
// asserted directly, which is what catches the clamp being dropped.
func TestEmissionIntervalClampsToANanosecond(t *testing.T) {
	assert.Equal(t, time.Nanosecond, emissionInterval(100, 50*time.Nanosecond), "n above per's nanosecond count")
	assert.Equal(t, 50*time.Millisecond, emissionInterval(2, 100*time.Millisecond), "even division")
	assert.Equal(t, 33333333*time.Nanosecond, emissionInterval(3, 100*time.Millisecond), "uneven division rounds down")
}

func TestDebounceCoalescesABurst(t *testing.T) {
	// The source is exhausted long before the quiet period elapses, so the
	// burst collapses to its last element.
	got := Debounce(t.Context(), streams.Of(1, 2, 3), time.Second).Collect()
	want := []int{3}
	assert.Equal(t, want, got, "Debounce")
}

func TestDebounceEmitsAfterTheQuietPeriod(t *testing.T) {
	ch := make(chan int)
	emitted := make(chan struct{})
	go func() {
		ch <- 1
		ch <- 2
		ch <- 3
		<-emitted // the quiet period has expired and 3 has been emitted
		ch <- 4
		close(ch)
	}()

	var got []int
	for v := range Debounce(t.Context(), streams.Chan(ch), 200*time.Millisecond) {
		got = append(got, v)
		if len(got) == 1 {
			close(emitted)
		}
	}
	// 3 comes from the quiet period, 4 from the end of the source.
	want := []int{3, 4}
	assert.Equal(t, want, got, "Debounce")
}

func TestSampleTakesTheMostRecentElement(t *testing.T) {
	ch := make(chan int)
	sampled := make(chan struct{})
	go func() {
		ch <- 1
		ch <- 2
		ch <- 3
		<-sampled // 3 has been sampled
		ch <- 4
		ch <- 5
		close(ch)
	}()

	var got []int
	for v := range Sample(t.Context(), streams.Chan(ch), 200*time.Millisecond) {
		got = append(got, v)
		if len(got) == 1 {
			close(sampled)
		}
	}

	require.GreaterOrEqualf(t, len(got), 2, "Sample = %v, want two or three elements", got)
	require.LessOrEqualf(t, len(got), 3, "Sample = %v, want two or three elements", got)
	assert.Equal(t, 3, got[0], "Sample first is the most recent of the burst")
	last := got[len(got)-1]
	assert.Equal(t, 5, last, "Sample last is the element left when the source ended")
	increasing(t, "Sample", got)
}

func TestSampleDropsElementsBetweenIntervals(t *testing.T) {
	const interval = 100 * time.Millisecond

	got := Sample(t.Context(), paced(5*time.Millisecond), interval).Take(3).Collect()

	require.Lenf(t, got, 3, "Sample = %v, want three samples", got)
	increasing(t, "Sample", got)
	// The source counts up from zero, so the last sample's value is the number
	// of elements produced before it. Sampling three of at least six means the
	// rest were dropped.
	assert.GreaterOrEqualf(t, got[2], 5, "Sample = %v, want the third sample to skip past the earlier elements", got)
}

// An interval in which no element arrived produces nothing rather than
// repeating the previous element. The other Sample tests keep the source
// faster than the interval, so none of them ever leaves one empty.
func TestSampleSkipsIntervalsWithNoElement(t *testing.T) {
	const interval = 50 * time.Millisecond
	ctx, cancel := context.WithTimeout(t.Context(), 8*interval)
	defer cancel()

	// One element, then silence: the first interval samples it and the rest
	// hold nothing.
	got := Sample(ctx, quiet(), interval).Collect()

	assert.Equal(t, []int{1}, got, "Sample over a source that goes quiet")
}

func TestTumblingGroupsByWindow(t *testing.T) {
	ch := make(chan int)
	cut := make(chan struct{})
	go func() {
		ch <- 1
		ch <- 2
		<-cut // the first window has closed
		ch <- 3
		close(ch)
	}()

	var got [][]int
	for w := range Tumbling(t.Context(), streams.Chan(ch), 200*time.Millisecond) {
		got = append(got, w)
		if len(got) == 1 {
			close(cut)
		}
	}

	want := [][]int{{1, 2}, {3}}
	assert.Equal(t, want, got, "Tumbling")
}

func TestTumblingFlushesTheOpenWindow(t *testing.T) {
	// The source is exhausted well inside the first window.
	got := Tumbling(t.Context(), streams.Of(1, 2, 3), time.Second).Collect()
	want := [][]int{{1, 2, 3}}
	assert.Equal(t, want, got, "Tumbling")
}

// A window in which no element arrived is skipped rather than emitted empty.
// Every other Tumbling test synchronises its source with the cut, so none of
// them ever lets a window pass with nothing in it.
func TestTumblingSkipsEmptyWindows(t *testing.T) {
	const size = 50 * time.Millisecond
	ctx, cancel := context.WithTimeout(t.Context(), 8*size)
	defer cancel()

	// One element, then silence for the remaining windows.
	got := Tumbling(ctx, quiet(), size).Collect()

	require.Lenf(t, got, 1, "Tumbling over a source that goes quiet = %v, want only the window holding the element", got)
	assert.Equal(t, []int{1}, got[0], "Tumbling window")
}

func TestSlidingOverlapsAndExpires(t *testing.T) {
	const (
		size  = 300 * time.Millisecond
		every = 50 * time.Millisecond
	)

	ch := make(chan int)
	go func() {
		ch <- 1
		time.Sleep(700 * time.Millisecond) // more than size, so 1 expires
		ch <- 2
		close(ch)
	}()

	got := Sliding(t.Context(), streams.Chan(ch), size, every).Collect()

	require.GreaterOrEqualf(t, len(got), 2, "Sliding = %v, want at least two windows", got)
	held := 0
	for _, w := range got {
		if !assert.NotEmptyf(t, w, "Sliding = %v, want no empty window", got) {
			break
		}
		if slices.Contains(w, 1) {
			held++
		}
	}
	assert.GreaterOrEqualf(t, held, 2, "Sliding held 1 in %d windows, want it to overlap at least two", held)
	last, want := got[len(got)-1], []int{2}
	assert.Equal(t, want, last, "Sliding last, with 1 expired")
}

func TestSessionGroupsByGap(t *testing.T) {
	ch := make(chan int)
	closed := make(chan struct{})
	go func() {
		ch <- 1
		ch <- 2
		<-closed // the gap expired and the first session was emitted
		ch <- 3
		ch <- 4
		close(ch)
	}()

	var got [][]int
	for s := range Session(t.Context(), streams.Chan(ch), 200*time.Millisecond) {
		got = append(got, s)
		if len(got) == 1 {
			close(closed)
		}
	}

	want := [][]int{{1, 2}, {3, 4}}
	assert.Equal(t, want, got, "Session")
	order := flatten(got)
	assert.Equal(t, []int{1, 2, 3, 4}, order, "Session order is arrival order")
}

func TestSessionFlushesTheOpenSession(t *testing.T) {
	got := Session(t.Context(), streams.Of(1, 2), time.Second).Collect()
	want := [][]int{{1, 2}}
	assert.Equal(t, want, got, "Session")
}

func TestTimeoutPassesAStreamThatFinishesInTime(t *testing.T) {
	got, err := streams.Try(Timeout(t.Context(), streams.Of(1, 2, 3), 5*time.Second))
	require.NoErrorf(t, err, "Timeout err = %v, want nil", err)
	assert.Equal(t, []int{1, 2, 3}, got, "Timeout")
}

func TestTimeoutBoundsTheWholeIteration(t *testing.T) {
	const d = 200 * time.Millisecond

	// The source never pauses for longer than 20ms, so only a deadline on the
	// iteration as a whole can end this.
	start := time.Now()
	_, err := streams.Try(Timeout(t.Context(), paced(20*time.Millisecond), d))
	elapsed := time.Since(start)

	require.ErrorIs(t, err, context.DeadlineExceeded, "Timeout err")
	atLeast := d - 10*time.Millisecond
	assert.GreaterOrEqualf(t, elapsed, atLeast, "Timeout fired after %v, want at least %v", elapsed, atLeast)
	// An upper bound as well, or a deadline off by an order of magnitude would
	// satisfy the lower one just as happily. It is loose enough that only a
	// mistake in the duration itself, not a slow scheduler, can reach it.
	atMost := 5 * d
	assert.LessOrEqualf(t, elapsed, atMost, "Timeout fired after %v, want at most %v", elapsed, atMost)
}

func TestTimeoutReportsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	_, err := streams.Try(Timeout(ctx, paced(20*time.Millisecond), time.Hour))
	assert.ErrorIs(t, err, context.Canceled, "Timeout err")
}

func TestIntervalTicks(t *testing.T) {
	const d = 30 * time.Millisecond

	start := time.Now()
	got := Interval(t.Context(), d).Take(3).Collect()
	elapsed := time.Since(start)

	require.Lenf(t, got, 3, "Interval = %v, want three ticks", got)
	// The first tick is one period in, not immediate.
	assert.Falsef(t, got[0].Before(start.Add(d)),
		"Interval first tick came %v after the start, want at least %v", got[0].Sub(start), d)
	for i := 1; i < len(got); i++ {
		if !assert.Truef(t, got[i].After(got[i-1]), "Interval ticks = %v, want increasing times", got) {
			break
		}
	}
	atLeast := 3*d - 10*time.Millisecond
	assert.GreaterOrEqualf(t, elapsed, atLeast, "Interval took %v for three ticks, want at least %v", elapsed, atLeast)
}

func TestStampPairsEachElementWithItsTime(t *testing.T) {
	start := time.Now()

	var values []string
	var times []time.Time
	for at, v := range Stamp(streams.Of("a", "b", "c")) {
		values = append(values, v)
		times = append(times, at)
	}
	end := time.Now()

	want := []string{"a", "b", "c"}
	assert.Equal(t, want, values, "Stamp values")
	require.Lenf(t, times, 3, "Stamp produced %d pairs, want 3", len(times))
	assert.Falsef(t, times[0].Before(start) || times[2].After(end),
		"Stamp times %v fall outside the iteration", times)
	for i := 1; i < len(times); i++ {
		if !assert.Falsef(t, times[i].Before(times[i-1]), "Stamp times = %v, want non-decreasing", times) {
			break
		}
	}
}

// The times must follow the source rather than being read once. Every
// assertion above holds just as well for a single timestamp shared by all
// three elements, because Of produces them too fast to tell apart; a source
// that pauses between elements is what separates the two.
func TestStampFollowsTheSourceThroughTime(t *testing.T) {
	const gap = 20 * time.Millisecond

	var times []time.Time
	for at := range Stamp(paced(gap)).Take(3) {
		times = append(times, at)
	}

	require.Lenf(t, times, 3, "Stamp produced %d pairs, want 3", len(times))
	for i := 1; i < len(times); i++ {
		assert.GreaterOrEqualf(t, times[i].Sub(times[i-1]), gap/2,
			"Stamp times = %v, want each one later than the last", times)
	}
}

func TestEmptySourceYieldsNothing(t *testing.T) {
	ctx := t.Context()
	empty := streams.Empty[int]

	assert.Zero(t, Throttle(ctx, empty(), time.Second).Count(), "Throttle over an empty Stream yielded elements")
	assert.Zero(t, Debounce(ctx, empty(), time.Second).Count(), "Debounce over an empty Stream yielded elements")
	assert.Zero(t, Sample(ctx, empty(), time.Second).Count(), "Sample over an empty Stream yielded elements")
	assert.Zero(t, Delay(ctx, empty(), time.Second).Count(), "Delay over an empty Stream yielded elements")
	assert.Zero(t, RateLimit(ctx, empty(), 1, time.Second).Count(), "RateLimit over an empty Stream yielded elements")
	assert.Zero(t, Tumbling(ctx, empty(), time.Second).Count(), "Tumbling over an empty Stream yielded windows")
	assert.Zero(t, Sliding(ctx, empty(), time.Second, time.Second).Count(), "Sliding over an empty Stream yielded windows")
	assert.Zero(t, Session(ctx, empty(), time.Second).Count(), "Session over an empty Stream yielded sessions")
	assert.Zero(t, Stamp(empty()).Count(), "Stamp over an empty Stream yielded pairs")
	got, err := streams.Try(Timeout(ctx, empty(), time.Second))
	// Try over an empty Stream returns a nil slice, so this asks for emptiness
	// rather than equality with a particular empty slice.
	assert.Emptyf(t, got, "Timeout over an empty Stream = %v, %v, want nothing", got, err)
	assert.NoError(t, err, "Timeout over an empty Stream")
}

func TestNonPositiveArgumentsPanic(t *testing.T) {
	ctx := context.Background()
	src := streams.Of(1, 2, 3)

	cases := []struct {
		name string
		call func()
	}{
		{"Throttle", func() { Throttle(ctx, src, 0) }},
		{"Debounce", func() { Debounce(ctx, src, 0) }},
		{"Sample", func() { Sample(ctx, src, -time.Second) }},
		{"Delay", func() { Delay(ctx, src, 0) }},
		{"RateLimit n", func() { RateLimit(ctx, src, 0, time.Second) }},
		{"RateLimit per", func() { RateLimit(ctx, src, 1, 0) }},
		{"Tumbling", func() { Tumbling(ctx, src, 0) }},
		{"Sliding size", func() { Sliding(ctx, src, 0, time.Second) }},
		{"Sliding every", func() { Sliding(ctx, src, time.Second, 0) }},
		{"Session", func() { Session(ctx, src, 0) }},
		{"Timeout", func() { Timeout(ctx, src, 0) }},
		{"Interval", func() { Interval(ctx, 0) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Panicsf(t, tc.call, "%s did not panic", tc.name)
		})
	}
}

// A drainer reads one operator, breaking out of the iteration as soon as stop
// reports true after an element.
type drainer struct {
	name  string
	drain func(ctx context.Context, s streams.Stream[int], stop func() bool)
}

// unit is the period every drained operator is built around.
const unit = 20 * time.Millisecond

// drainers holds one entry per operator that waits on the clock.
func drainers() []drainer {
	return []drainer{
		{"Throttle", func(ctx context.Context, s streams.Stream[int], stop func() bool) {
			for range Throttle(ctx, s, unit) {
				if stop() {
					break
				}
			}
		}},
		{"Debounce", func(ctx context.Context, s streams.Stream[int], stop func() bool) {
			for range Debounce(ctx, s, unit) {
				if stop() {
					break
				}
			}
		}},
		{"Sample", func(ctx context.Context, s streams.Stream[int], stop func() bool) {
			for range Sample(ctx, s, unit) {
				if stop() {
					break
				}
			}
		}},
		{"Delay", func(ctx context.Context, s streams.Stream[int], stop func() bool) {
			for range Delay(ctx, s, unit) {
				if stop() {
					break
				}
			}
		}},
		{"RateLimit", func(ctx context.Context, s streams.Stream[int], stop func() bool) {
			for range RateLimit(ctx, s, 1, unit) {
				if stop() {
					break
				}
			}
		}},
		{"Tumbling", func(ctx context.Context, s streams.Stream[int], stop func() bool) {
			for range Tumbling(ctx, s, unit) {
				if stop() {
					break
				}
			}
		}},
		{"Sliding", func(ctx context.Context, s streams.Stream[int], stop func() bool) {
			for range Sliding(ctx, s, 2*unit, unit) {
				if stop() {
					break
				}
			}
		}},
		{"Session", func(ctx context.Context, s streams.Stream[int], stop func() bool) {
			for range Session(ctx, s, unit) {
				if stop() {
					break
				}
			}
		}},
		{"Timeout", func(ctx context.Context, s streams.Stream[int], stop func() bool) {
			for range Timeout(ctx, s, time.Hour) {
				if stop() {
					break
				}
			}
		}},
		{"Interval", func(ctx context.Context, _ streams.Stream[int], stop func() bool) {
			for range Interval(ctx, unit) {
				if stop() {
					break
				}
			}
		}},
	}
}

func keepReading() bool { return false }

func stopAtOnce() bool { return true }

// end runs drain in its own goroutine and fails the test unless it returns
// within a span far longer than any operator needs to notice it is over.
func end(t *testing.T, name string, drain func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		drain()
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		require.Failf(t, "operator did not end", "%s did not end", name)
	}
}

// cancelSoon cancels ctx once the operator under test is running rather than
// starting. Nothing is asserted about how long that takes.
func cancelSoon(cancel context.CancelFunc) {
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
}

func TestCancellationEndsEveryOperator(t *testing.T) {
	for _, tc := range drainers() {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			cancelSoon(cancel)
			end(t, tc.name, func() { tc.drain(ctx, paced(time.Millisecond), keepReading) })
		})
	}
}

func TestOperatorsLeaveNoGoroutineBehind(t *testing.T) {
	before := countGoroutines()

	// Both ways out of an operator have to release the goroutine reading the
	// source: the consumer stopping, and the context being cancelled. For the
	// first the source is slower than unit, so that every operator has
	// something to emit and the consumer gets its chance to stop.
	for _, tc := range drainers() {
		end(t, tc.name, func() { tc.drain(context.Background(), paced(3*unit), stopAtOnce) })

		ctx, cancel := context.WithCancel(context.Background())
		cancelSoon(cancel)
		end(t, tc.name, func() { tc.drain(ctx, paced(time.Millisecond), keepReading) })
		cancel()
	}

	// The source and the goroutine reading it unwind on their own schedule.
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if countGoroutines() <= before+2 {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	assert.Failf(t, "goroutines leaked", "before=%d after=%d", before, countGoroutines())
}

// quiet yields one element and then blocks forever. It is the state a live feed
// is in whenever nothing is happening, and it is what the paced source used by
// the other cancellation tests cannot reproduce: paced keeps waking the reader,
// which hides whether the operator itself observes the context.
func quiet() streams.Stream[int] {
	ch := make(chan int, 1)
	ch <- 1
	return streams.Chan(ch)
}

// Cancelling the context must end the operator even when the source has gone
// quiet. Throttle, Delay and RateLimit read the source with a plain range until
// this was fixed, so their only context check ran after an element arrived and
// they hung for the life of the process.
func TestCancellationEndsEveryOperatorWithAQuietSource(t *testing.T) {
	drain := map[string]func(ctx context.Context){
		"Throttle": func(ctx context.Context) {
			for range Throttle(ctx, quiet(), 10*time.Millisecond) {
			}
		},
		"Delay": func(ctx context.Context) {
			for range Delay(ctx, quiet(), 10*time.Millisecond) {
			}
		},
		"RateLimit": func(ctx context.Context) {
			for range RateLimit(ctx, quiet(), 5, 10*time.Millisecond) {
			}
		},
		"Debounce": func(ctx context.Context) {
			for range Debounce(ctx, quiet(), 10*time.Millisecond) {
			}
		},
		"Sample": func(ctx context.Context) {
			for range Sample(ctx, quiet(), 10*time.Millisecond) {
			}
		},
		"Tumbling": func(ctx context.Context) {
			for range Tumbling(ctx, quiet(), 10*time.Millisecond) {
			}
		},
		"Sliding": func(ctx context.Context) {
			for range Sliding(ctx, quiet(), 20*time.Millisecond, 10*time.Millisecond) {
			}
		},
		"Session": func(ctx context.Context) {
			for range Session(ctx, quiet(), 10*time.Millisecond) {
			}
		},
		"Interval": func(ctx context.Context) {
			for range Interval(ctx, 10*time.Millisecond) {
			}
		},
	}
	for name, run := range drain {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan struct{})
			go func() { defer close(done); run(ctx) }()
			time.Sleep(50 * time.Millisecond)
			cancel()
			select {
			case <-done:
			case <-time.After(2 * time.Second):
				require.Fail(t, "operator did not return within 2s of the context being cancelled")
			}
		})
	}
}

// burst returns a source whose elements are already buffered and whose channel
// is closed, so an operator's receive case is ready the moment its select runs.
// Together with a context cancelled before iteration begins, that forces the
// select tie the operators must break in cancellation's favour.
func burst() streams.Stream[int] {
	ch := make(chan int, 3)
	ch <- 1
	ch <- 2
	ch <- 3
	close(ch)
	return streams.Chan(ch)
}

// Once the context is done nothing further is emitted — the promise doc.go
// makes for every operator, end-of-source flushes included. It needs its own
// test because a done context and a ready element or timer are both live
// select cases, where the winner is random: each operator is driven many times
// so a regression cannot hide behind a lucky draw. The nanosecond unit makes
// the timers part of the tie — a longer one would never be ready before the
// operator returns, leaving the timer branches to fire only in production.
// Timeout, the documented exception, has the test after this one.
func TestOperatorsEmitNothingOnceCancelled(t *testing.T) {
	const unit = time.Nanosecond
	count := map[string]func(ctx context.Context) int{
		"Throttle":  func(ctx context.Context) int { return Throttle(ctx, burst(), unit).Count() },
		"Debounce":  func(ctx context.Context) int { return Debounce(ctx, burst(), unit).Count() },
		"Sample":    func(ctx context.Context) int { return Sample(ctx, burst(), unit).Count() },
		"Delay":     func(ctx context.Context) int { return Delay(ctx, burst(), unit).Count() },
		"RateLimit": func(ctx context.Context) int { return RateLimit(ctx, burst(), 5, unit).Count() },
		"Tumbling":  func(ctx context.Context) int { return Tumbling(ctx, burst(), unit).Count() },
		"Sliding":   func(ctx context.Context) int { return Sliding(ctx, burst(), 2*unit, unit).Count() },
		"Session":   func(ctx context.Context) int { return Session(ctx, burst(), unit).Count() },
		"Interval":  func(ctx context.Context) int { return Interval(ctx, unit).Count() },
	}
	for name, run := range count {
		t.Run(name, func(t *testing.T) {
			for range 50 {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				assert.Zerof(t, run(ctx), "%s emitted after cancellation", name)
			}
		})
	}
}

// Timeout is the exception doc.go names: cancellation still emits, but only
// the one final pair carrying ctx.Err(), never an element. The nanosecond
// deadline puts the expired timer into the tie as well, and cancellation must
// beat it too.
func TestTimeoutReportsCancellationWithoutEmittingElements(t *testing.T) {
	for range 50 {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		values := 0
		var errs []error
		for _, err := range Timeout(ctx, burst(), time.Nanosecond) {
			if err != nil {
				errs = append(errs, err)
			} else {
				values++
			}
		}
		assert.Zero(t, values, "Timeout emitted elements after cancellation")
		require.Len(t, errs, 1, "Timeout errors")
		assert.ErrorIs(t, errs[0], context.Canceled, "Timeout error")
	}
}

// A done context still reports even when the source is already exhausted:
// whichever the select observes first — the done context or the closed source
// — the final pair carries ctx.Err(), so the outcome does not depend on the
// draw.
func TestTimeoutReportsCancellationOnAnExhaustedSource(t *testing.T) {
	for range 50 {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		var errs []error
		for _, err := range Timeout(ctx, streams.Empty[int](), time.Second) {
			errs = append(errs, err)
		}
		require.Len(t, errs, 1, "Timeout pairs")
		assert.ErrorIs(t, errs[0], context.Canceled, "Timeout error")
	}
}

// Cancelling with elements still in flight races the done context against the
// next ready element, and both draws must suppress the element and report
// ctx.Err(). The sleep parks the source's next element in the handoff, so the
// tie is real.
func TestTimeoutReportsCancellationBetweenElements(t *testing.T) {
	for range 50 {
		ctx, cancel := context.WithCancel(context.Background())
		var got []int
		var errs []error
		for v, err := range Timeout(ctx, burst(), time.Second) {
			if err != nil {
				errs = append(errs, err)
				continue
			}
			got = append(got, v)
			cancel()
			time.Sleep(time.Millisecond)
		}
		cancel() // a second call is a no-op; vet wants it called on every path
		assert.Equal(t, []int{1}, got, "Timeout values")
		require.Len(t, errs, 1, "Timeout errors")
		assert.ErrorIs(t, errs[0], context.Canceled, "Timeout error")
	}
}

// Cancelling after the last element races the done context against the
// source's close, and both draws must report ctx.Err(). The sleep gives the
// source time to finish, so the close is ready and the tie is real.
func TestTimeoutReportsCancellationAfterTheLastElement(t *testing.T) {
	for range 50 {
		ctx, cancel := context.WithCancel(context.Background())
		var got []int
		var errs []error
		for v, err := range Timeout(ctx, streams.Of(7), time.Second) {
			if err != nil {
				errs = append(errs, err)
				continue
			}
			got = append(got, v)
			time.Sleep(time.Millisecond)
			cancel()
		}
		cancel() // a second call is a no-op; vet wants it called on every path
		assert.Equal(t, []int{7}, got, "Timeout values")
		require.Len(t, errs, 1, "Timeout errors")
		assert.ErrorIs(t, errs[0], context.Canceled, "Timeout error")
	}
}

// A deadline that expires while the consumer is busy must win the tie with the
// next ready element: the doc promises the deadline is noticed between
// elements, and noticing must not depend on the draw. The sleep lets the
// deadline fire and the next element park in the handoff, so both are ready.
// A stalled start can make the deadline beat even the first element — also
// correct — so the values are asserted as at most that element, never as
// exactly it; the failure this test exists to catch is a second one.
func TestTimeoutNoticesTheDeadlineBeforeTheNextElement(t *testing.T) {
	const d = 10 * time.Millisecond
	for range 25 {
		var got []int
		var errs []error
		for v, err := range Timeout(t.Context(), burst(), d) {
			if err != nil {
				errs = append(errs, err)
				continue
			}
			got = append(got, v)
			time.Sleep(25 * time.Millisecond)
		}
		assert.LessOrEqualf(t, len(got), 1, "Timeout emitted %v: an element after the deadline", got)
		require.Len(t, errs, 1, "Timeout errors")
		assert.ErrorIs(t, errs[0], context.DeadlineExceeded, "Timeout error")
	}
}

// A deadline that expires before the source's close is observed must be
// reported: a clean end wins only while the deadline has not fired. The
// consumer outspends d on the only element, so the close and the expired
// timer reach the select together. A stalled start can make the deadline beat
// the element itself — also correct, and also not a clean end — so the values
// are asserted as at most that element; the failure this test exists to catch
// is a missing error.
func TestTimeoutNoticesTheDeadlineBeforeTheClose(t *testing.T) {
	const d = 10 * time.Millisecond
	for range 25 {
		var got []int
		var errs []error
		for v, err := range Timeout(t.Context(), streams.Of(7), d) {
			if err != nil {
				errs = append(errs, err)
				continue
			}
			got = append(got, v)
			time.Sleep(25 * time.Millisecond)
		}
		assert.LessOrEqualf(t, len(got), 1, "Timeout emitted %v, want at most the one element", got)
		require.Len(t, errs, 1, "Timeout errors")
		assert.ErrorIs(t, errs[0], context.DeadlineExceeded, "Timeout error")
	}
}

// wait must refuse a done context even when its timer has already fired: both
// are live select cases and the tie is broken at random, so without a recheck
// Throttle and RateLimit could emit one element after cancellation. The
// nanosecond timer has always expired by the time the select runs, which makes
// every pass a tie.
func TestWaitRefusesADoneContextWithAnExpiredTimer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for range 100 {
		assert.False(t, wait(ctx, time.Nanosecond), "wait with a done context")
	}
}

// span must leave the expired elements unreachable, not merely outside the
// returned slice: Sliding keeps its held slice for as long as the Stream runs,
// so whatever the backing array still references is whatever the GC cannot
// take. The test inspects the array directly, which is exactly the surface a
// consumer never sees.
func TestSpanReleasesExpiredElements(t *testing.T) {
	old := time.Now().Add(-time.Minute)
	a, b, c := new(int), new(int), new(int)
	held := []stamped[*int]{{at: old, value: a}, {at: old, value: b}, {at: time.Now(), value: c}}

	kept, window := span(held, time.Second)
	require.Len(t, kept, 1, "span kept")
	assert.Equal(t, []*int{c}, window, "span window")
	assert.Nil(t, held[1].value, "a vacated slot still references its element")
	assert.Nil(t, held[2].value, "a vacated slot still references its element")

	kept, window = span(kept, time.Nanosecond)
	assert.Nil(t, kept, "a batch that expired in full must drop its backing array")
	assert.Nil(t, window, "window of a fully expired batch")
}

// A stopped operator leaves nothing behind but its documented parked reader:
// one goroutine per iteration, stuck inside the quiet source, is expected —
// doc.go explains why it cannot be reclaimed — and anything beyond that is an
// operator goroutine that leaked. Timer hygiene is outside what a goroutine
// count can see, because a runtime timer has no goroutine of its own and a
// missing Stop would not move the count; the deferred Stops are a code
// convention, not a promise this test could pin.
func TestStoppedOperatorsLeakOnlyTheParkedReader(t *testing.T) {
	base := countGoroutines()
	for range 30 {
		ctx, cancel := context.WithCancel(context.Background())
		for range Tumbling(ctx, quiet(), 5*time.Millisecond) {
			break
		}
		cancel()
	}
	// One parked reader per iteration is expected and documented; anything
	// beyond that means an operator goroutine leaked.
	limit := base + 30 + 5
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if countGoroutines() <= limit {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	assert.Failf(t, "operator goroutines leaked",
		"goroutines = %d, want at most %d (30 documented parked readers plus slack)",
		countGoroutines(), limit)
}

// The reader goroutine an operator uses is released only where the source hands
// control back. streams.Chan over a quiet channel never does, so the reader
// stays parked; streams.ChanContext selects on the same context the operator
// was given, so cancelling reclaims it. This test pins both halves, because the
// difference is the whole reason ChanContext exists.
func TestReaderGoroutineIsReclaimedWithAContextAwareSource(t *testing.T) {
	const runs = 30

	// A channel that carries one element and then stays quiet forever. Nothing
	// closes it: this is a live feed that has simply gone idle.
	start := func() (chan int, context.Context, context.CancelFunc) {
		ch := make(chan int, 1)
		ch <- 1
		ctx, cancel := context.WithCancel(context.Background())
		return ch, ctx, cancel
	}

	t.Run("ChanContext releases it", func(t *testing.T) {
		before := countGoroutines()
		for range runs {
			ch, ctx, cancel := start()
			for range Sample(ctx, streams.ChanContext(ctx, ch), 5*time.Millisecond) {
				break
			}
			cancel()
		}
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			if countGoroutines() <= before+2 {
				return
			}
			time.Sleep(20 * time.Millisecond)
		}
		assert.Failf(t, "goroutines stranded with a context-aware source", "before=%d after=%d",
			before, countGoroutines())
	})

	t.Run("Chan over a quiet channel parks it, as documented", func(t *testing.T) {
		before := countGoroutines()
		chans := make([]chan int, 0, runs)
		for range runs {
			ch, ctx, cancel := start()
			for range Sample(ctx, streams.Chan(ch), 5*time.Millisecond) {
				break
			}
			cancel()
			chans = append(chans, ch)
		}
		time.Sleep(200 * time.Millisecond)
		got := countGoroutines()
		assert.GreaterOrEqualf(t, got, before+runs,
			"expected %d readers parked in the quiet source, saw %d extra;"+
				" if this now passes, Chan has become interruptible and the"+
				" documentation in doc.go and ChanContext must be revisited",
			runs, got-before)
		// Closing the channel is the other documented way out; take it so the
		// test leaves nothing behind.
		for _, ch := range chans {
			close(ch)
		}
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			if countGoroutines() <= before+2 {
				return
			}
			time.Sleep(20 * time.Millisecond)
		}
		assert.Failf(t, "closing the channel did not release the readers", "before=%d after=%d",
			before, countGoroutines())
	})
}
