package temporal

import (
	"context"
	"errors"
	"runtime"
	"slices"
	"testing"
	"time"

	streams "github.com/coldsmirk/go-streams/v2"
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
	for i := 1; i < len(got); i++ {
		if got[i] <= got[i-1] {
			t.Errorf("%s = %v, want strictly increasing", name, got)
			return
		}
	}
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

	if want := []int{1, 2, 3, 4}; !slices.Equal(got, want) {
		t.Errorf("Throttle = %v, want %v", got, want)
	}
	// The first element is free, the other three each wait an interval.
	if atLeast := 3*interval - 10*time.Millisecond; elapsed < atLeast {
		t.Errorf("Throttle took %v, want at least %v", elapsed, atLeast)
	}
}

func TestThrottleEmitsTheFirstElementImmediately(t *testing.T) {
	// The interval is far longer than the bound, so only a stall of most of a
	// second could make this fail.
	start := time.Now()
	got, ok := Throttle(t.Context(), streams.Of(7), time.Second).First()
	elapsed := time.Since(start)

	if !ok || got != 7 {
		t.Errorf("Throttle first = %v, %v, want 7, true", got, ok)
	}
	if elapsed > 500*time.Millisecond {
		t.Errorf("Throttle held the first element for %v", elapsed)
	}
}

func TestDelayShiftsEveryElement(t *testing.T) {
	const d = 30 * time.Millisecond

	start := time.Now()
	got := Delay(t.Context(), streams.Of(1, 2, 3), d).Collect()
	elapsed := time.Since(start)

	if want := []int{1, 2, 3}; !slices.Equal(got, want) {
		t.Errorf("Delay = %v, want %v", got, want)
	}
	if atLeast := 3*d - 10*time.Millisecond; elapsed < atLeast {
		t.Errorf("Delay took %v, want at least %v", elapsed, atLeast)
	}
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

	if want := []int{0, 1, 2, 3, 4, 5}; !slices.Equal(got, want) {
		t.Errorf("RateLimit = %v, want %v", got, want)
	}
	// Two elements ride the initial burst; the other four are paced.
	if atLeast := 3 * emission; elapsed < atLeast {
		t.Errorf("RateLimit took %v, want at least %v", elapsed, atLeast)
	}
}

func TestRateLimitEmitsTheInitialBurstImmediately(t *testing.T) {
	start := time.Now()
	got := RateLimit(t.Context(), streams.Of(1, 2, 3), 3, 3*time.Second).Collect()
	elapsed := time.Since(start)

	if want := []int{1, 2, 3}; !slices.Equal(got, want) {
		t.Errorf("RateLimit = %v, want %v", got, want)
	}
	if elapsed > time.Second {
		t.Errorf("RateLimit held the initial burst for %v", elapsed)
	}
}

func TestDebounceCoalescesABurst(t *testing.T) {
	// The source is exhausted long before the quiet period elapses, so the
	// burst collapses to its last element.
	got := Debounce(t.Context(), streams.Of(1, 2, 3), time.Second).Collect()
	if want := []int{3}; !slices.Equal(got, want) {
		t.Errorf("Debounce = %v, want %v", got, want)
	}
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
	if want := []int{3, 4}; !slices.Equal(got, want) {
		t.Errorf("Debounce = %v, want %v", got, want)
	}
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

	if len(got) < 2 || len(got) > 3 {
		t.Fatalf("Sample = %v, want two or three elements", got)
	}
	if got[0] != 3 {
		t.Errorf("Sample first = %d, want 3, the most recent of the burst", got[0])
	}
	if last := got[len(got)-1]; last != 5 {
		t.Errorf("Sample last = %d, want 5, the element left when the source ended", last)
	}
	increasing(t, "Sample", got)
}

func TestSampleDropsElementsBetweenIntervals(t *testing.T) {
	const interval = 100 * time.Millisecond

	got := Sample(t.Context(), paced(5*time.Millisecond), interval).Take(3).Collect()

	if len(got) != 3 {
		t.Fatalf("Sample = %v, want three samples", got)
	}
	increasing(t, "Sample", got)
	// The source counts up from zero, so the last sample's value is the number
	// of elements produced before it. Sampling three of at least six means the
	// rest were dropped.
	if got[2] < 5 {
		t.Errorf("Sample = %v, want the third sample to skip past the earlier elements", got)
	}
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
	if !slices.EqualFunc(got, want, slices.Equal) {
		t.Errorf("Tumbling = %v, want %v", got, want)
	}
}

func TestTumblingFlushesTheOpenWindow(t *testing.T) {
	// The source is exhausted well inside the first window.
	got := Tumbling(t.Context(), streams.Of(1, 2, 3), time.Second).Collect()
	want := [][]int{{1, 2, 3}}
	if !slices.EqualFunc(got, want, slices.Equal) {
		t.Errorf("Tumbling = %v, want %v", got, want)
	}
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

	if len(got) < 2 {
		t.Fatalf("Sliding = %v, want at least two windows", got)
	}
	held := 0
	for _, w := range got {
		if len(w) == 0 {
			t.Errorf("Sliding = %v, want no empty window", got)
			break
		}
		if slices.Contains(w, 1) {
			held++
		}
	}
	if held < 2 {
		t.Errorf("Sliding held 1 in %d windows, want it to overlap at least two", held)
	}
	if last, want := got[len(got)-1], []int{2}; !slices.Equal(last, want) {
		t.Errorf("Sliding last = %v, want %v, with 1 expired", last, want)
	}
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
	if !slices.EqualFunc(got, want, slices.Equal) {
		t.Errorf("Session = %v, want %v", got, want)
	}
	if order := flatten(got); !slices.Equal(order, []int{1, 2, 3, 4}) {
		t.Errorf("Session order = %v, want arrival order", order)
	}
}

func TestSessionFlushesTheOpenSession(t *testing.T) {
	got := Session(t.Context(), streams.Of(1, 2), time.Second).Collect()
	want := [][]int{{1, 2}}
	if !slices.EqualFunc(got, want, slices.Equal) {
		t.Errorf("Session = %v, want %v", got, want)
	}
}

func TestTimeoutPassesAStreamThatFinishesInTime(t *testing.T) {
	got, err := streams.Try(Timeout(t.Context(), streams.Of(1, 2, 3), 5*time.Second))
	if err != nil {
		t.Errorf("Timeout err = %v, want nil", err)
	}
	if want := []int{1, 2, 3}; !slices.Equal(got, want) {
		t.Errorf("Timeout = %v, want %v", got, want)
	}
}

func TestTimeoutBoundsTheWholeIteration(t *testing.T) {
	const d = 200 * time.Millisecond

	// The source never pauses for longer than 20ms, so only a deadline on the
	// iteration as a whole can end this.
	start := time.Now()
	_, err := streams.Try(Timeout(t.Context(), paced(20*time.Millisecond), d))
	elapsed := time.Since(start)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Timeout err = %v, want %v", err, context.DeadlineExceeded)
	}
	if atLeast := d - 10*time.Millisecond; elapsed < atLeast {
		t.Errorf("Timeout fired after %v, want at least %v", elapsed, atLeast)
	}
}

func TestTimeoutReportsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	_, err := streams.Try(Timeout(ctx, paced(20*time.Millisecond), time.Hour))
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Timeout err = %v, want %v", err, context.Canceled)
	}
}

func TestIntervalTicks(t *testing.T) {
	const d = 30 * time.Millisecond

	start := time.Now()
	got := Interval(t.Context(), d).Take(3).Collect()
	elapsed := time.Since(start)

	if len(got) != 3 {
		t.Fatalf("Interval = %v, want three ticks", got)
	}
	// The first tick is one period in, not immediate.
	if got[0].Before(start.Add(d)) {
		t.Errorf("Interval first tick came %v after the start, want at least %v", got[0].Sub(start), d)
	}
	for i := 1; i < len(got); i++ {
		if !got[i].After(got[i-1]) {
			t.Errorf("Interval ticks = %v, want increasing times", got)
			break
		}
	}
	if atLeast := 3*d - 10*time.Millisecond; elapsed < atLeast {
		t.Errorf("Interval took %v for three ticks, want at least %v", elapsed, atLeast)
	}
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

	if want := []string{"a", "b", "c"}; !slices.Equal(values, want) {
		t.Errorf("Stamp values = %v, want %v", values, want)
	}
	if len(times) != 3 {
		t.Fatalf("Stamp produced %d pairs, want 3", len(times))
	}
	if times[0].Before(start) || times[2].After(end) {
		t.Errorf("Stamp times %v fall outside the iteration", times)
	}
	for i := 1; i < len(times); i++ {
		if times[i].Before(times[i-1]) {
			t.Errorf("Stamp times = %v, want non-decreasing", times)
			break
		}
	}
}

func TestEmptySourceYieldsNothing(t *testing.T) {
	ctx := t.Context()
	empty := streams.Empty[int]

	if got := Throttle(ctx, empty(), time.Second).Count(); got != 0 {
		t.Errorf("Throttle over an empty Stream yielded %d elements", got)
	}
	if got := Debounce(ctx, empty(), time.Second).Count(); got != 0 {
		t.Errorf("Debounce over an empty Stream yielded %d elements", got)
	}
	if got := Sample(ctx, empty(), time.Second).Count(); got != 0 {
		t.Errorf("Sample over an empty Stream yielded %d elements", got)
	}
	if got := Delay(ctx, empty(), time.Second).Count(); got != 0 {
		t.Errorf("Delay over an empty Stream yielded %d elements", got)
	}
	if got := RateLimit(ctx, empty(), 1, time.Second).Count(); got != 0 {
		t.Errorf("RateLimit over an empty Stream yielded %d elements", got)
	}
	if got := Tumbling(ctx, empty(), time.Second).Count(); got != 0 {
		t.Errorf("Tumbling over an empty Stream yielded %d windows", got)
	}
	if got := Sliding(ctx, empty(), time.Second, time.Second).Count(); got != 0 {
		t.Errorf("Sliding over an empty Stream yielded %d windows", got)
	}
	if got := Session(ctx, empty(), time.Second).Count(); got != 0 {
		t.Errorf("Session over an empty Stream yielded %d sessions", got)
	}
	if got := Stamp(empty()).Count(); got != 0 {
		t.Errorf("Stamp over an empty Stream yielded %d pairs", got)
	}
	got, err := streams.Try(Timeout(ctx, empty(), time.Second))
	if len(got) != 0 || err != nil {
		t.Errorf("Timeout over an empty Stream = %v, %v, want nothing", got, err)
	}
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
			defer func() {
				if recover() == nil {
					t.Errorf("%s did not panic", tc.name)
				}
			}()
			tc.call()
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
		t.Fatalf("%s did not end", name)
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
	t.Errorf("goroutines leaked: before=%d after=%d", before, countGoroutines())
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
				t.Fatal("operator did not return within 2s of the context being cancelled")
			}
		})
	}
}

// Timers must be released even for a source that never yields again. The reader
// goroutine is a weaker guarantee and is documented as such in doc.go: a source
// that stays quiet keeps it parked, and Go offers no way to reclaim it. This
// test pins the part that is guaranteed, so a regression in timer handling is
// caught without asserting a promise the package does not make.
func TestOperatorsReleaseTheirTimersWithAQuietSource(t *testing.T) {
	base := countGoroutines()
	for range 30 {
		ctx, cancel := context.WithCancel(context.Background())
		for range Tumbling(ctx, quiet(), 5*time.Millisecond) {
			break
		}
		cancel()
	}
	// One parked reader per iteration is expected and documented; anything
	// beyond that means a timer goroutine or an operator goroutine also leaked.
	limit := base + 30 + 5
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if countGoroutines() <= limit {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Errorf("goroutines = %d, want at most %d (30 documented parked readers plus slack)",
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
		t.Errorf("goroutines stranded with a context-aware source: before=%d after=%d",
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
		if got := countGoroutines(); got < before+runs {
			t.Errorf("expected %d readers parked in the quiet source, saw %d extra;"+
				" if this now passes, Chan has become interruptible and the"+
				" documentation in doc.go and ChanContext must be revisited",
				runs, got-before)
		}
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
		t.Errorf("closing the channel did not release the readers: before=%d after=%d",
			before, countGoroutines())
	})
}
