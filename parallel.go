package streams

import (
	"runtime"
	"sync"
)

// A ParallelOption configures a parallel operation.
type ParallelOption func(*parallelConfig)

type parallelConfig struct {
	concurrency int
	ordered     bool
}

func newParallelConfig(opts []ParallelOption) parallelConfig {
	cfg := parallelConfig{concurrency: runtime.GOMAXPROCS(0), ordered: true}
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.concurrency < 1 {
		cfg.concurrency = 1
	}
	return cfg
}

// WithConcurrency sets how many elements may be processed at once. The default
// is runtime.GOMAXPROCS(0). Values below one are treated as one.
func WithConcurrency(n int) ParallelOption {
	return func(c *parallelConfig) { c.concurrency = n }
}

// Unordered lets results be emitted as they finish rather than in the order of
// the input. It is faster when the work per element varies.
func Unordered() ParallelOption {
	return func(c *parallelConfig) { c.ordered = false }
}

// ParallelMap is [Stream.Map] with fn applied concurrently. Results keep their
// input order unless [Unordered] is given. fn must be safe for concurrent use.
//
// Concurrency is not free: each element costs a goroutine and a channel
// handoff, which is on the order of a microsecond. Below roughly five
// microseconds of work per element this is slower than [Stream.Map], and for
// trivial work it is slower by orders of magnitude. Measure before reaching
// for it.
//
// ParallelMap reads s on a goroutine of its own. Stopping the returned Stream
// early releases that goroutine at the next point where s yields or ends, so a
// source that blocks indefinitely, such as [Chan] over a quiet channel, keeps
// it parked until then. Use [ChanContext] or another source that ends on
// cancellation where bounded shutdown matters; the documentation of
// [github.com/coldsmirk/go-streams/v2/temporal] describes the underlying
// constraint.
func (s Stream[T]) ParallelMap[R any](fn func(T) R, opts ...ParallelOption) Stream[R] {
	return parallelRun(s, func(v T) (R, bool) { return fn(v), true }, newParallelConfig(opts))
}

// ParallelFilter is [Stream.Filter] with pred evaluated concurrently. Elements
// keep their input order unless [Unordered] is given. pred must be safe for
// concurrent use. The notes on [Stream.ParallelMap] about when concurrency
// pays and about the goroutine that reads s apply here too.
func (s Stream[T]) ParallelFilter(pred func(T) bool, opts ...ParallelOption) Stream[T] {
	return parallelRun(s, func(v T) (T, bool) { return v, pred(v) }, newParallelConfig(opts))
}

// outcome carries what a worker produced together with whether it should be
// emitted, so that one engine serves both mapping and filtering.
type outcome[R any] struct {
	value R
	keep  bool
}

// parallelRun applies work to the elements of s concurrently. work returns the
// value to emit and whether to emit it.
//
// Each element gets its own goroutine. A fixed pool of workers would cost far
// less — measured at a tenth of the allocations and roughly half the time on
// trivial work — but it was tried and rejected, because it partitions the input
// and so cannot move a slow element off a busy worker. Concurrency only pays
// here above roughly five microseconds of work per element, and work that heavy
// is usually work of uneven duration: an HTTP call, a query, a decode. In that
// regime a pool measured 20% slower, while the per-element allocation it saves
// is a fraction of a percent of the work. The cheap design is cheap in the
// regime this operator should not be used in, so it is not the one to have.
//
// It is a free function rather than a method because a method would have to
// name outcome in terms of its receiver, making Stream[T] instantiate
// Stream[outcome[T]] without end — an instantiation cycle the compiler
// rejects. As a free function the filtering path instantiates
// parallelRun[T, T], whose type arguments do not grow.
func parallelRun[T, R any](s Stream[T], work func(T) (R, bool), cfg parallelConfig) Stream[R] {
	if cfg.ordered {
		return parallelOrdered(s, work, cfg.concurrency)
	}
	return parallelUnordered(s, work, cfg.concurrency)
}

func parallelOrdered[T, R any](s Stream[T], work func(T) (R, bool), concurrency int) Stream[R] {
	return func(yield func(R) bool) {
		done := make(chan struct{})
		defer close(done)

		// The semaphore is what bounds the work, not the buffer: the consumer
		// frees a buffer cell as soon as it takes a slot, then waits on that
		// slot while its worker is still running, which would let the producer
		// start one more than concurrency.
		results := make(chan chan outcome[R], concurrency)
		sem := make(chan struct{}, concurrency)

		go func() {
			defer close(results)
			for v := range s {
				select {
				case sem <- struct{}{}:
				case <-done:
					return
				}
				slot := make(chan outcome[R], 1)
				select {
				case results <- slot:
				case <-done:
					<-sem
					return
				}
				go func() {
					defer func() { <-sem }()
					r, keep := work(v)
					slot <- outcome[R]{value: r, keep: keep}
				}()
			}
		}()

		for slot := range results {
			if o := <-slot; o.keep && !yield(o.value) {
				return
			}
		}
	}
}

func parallelUnordered[T, R any](s Stream[T], work func(T) (R, bool), concurrency int) Stream[R] {
	return func(yield func(R) bool) {
		done := make(chan struct{})
		defer close(done)

		results := make(chan R)
		var wg sync.WaitGroup
		sem := make(chan struct{}, concurrency)

		go func() {
			// LIFO: every worker has finished before results is closed, even
			// when the consumer stops early and this loop returns below.
			defer close(results)
			defer wg.Wait()
			for v := range s {
				select {
				case sem <- struct{}{}:
				case <-done:
					return
				}
				wg.Go(func() {
					defer func() { <-sem }()
					r, keep := work(v)
					if !keep {
						return
					}
					select {
					case results <- r:
					case <-done:
					}
				})
			}
		}()

		for r := range results {
			if !yield(r) {
				return
			}
		}
	}
}

// ParallelForEach calls fn for each element concurrently and returns once every
// call has completed. fn must be safe for concurrent use. [Unordered] has no
// effect here, since there are no results to order.
func (s Stream[T]) ParallelForEach(fn func(T), opts ...ParallelOption) {
	cfg := newParallelConfig(opts)
	var wg sync.WaitGroup
	sem := make(chan struct{}, cfg.concurrency)
	for v := range s {
		sem <- struct{}{}
		wg.Go(func() {
			defer func() { <-sem }()
			fn(v)
		})
	}
	wg.Wait()
}
