package context_examples

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	streams "github.com/coldsmirk/go-streams"
)

func Example_withContext_and_ctxConstructors() {
	ctx := context.Background()
	fmt.Println(streams.WithContext(ctx, streams.Of(1, 2, 3)).Count())
	fmt.Println(streams.WithContext2(ctx, streams.PairsOf(streams.NewPair("k", 1))).Count())
	fmt.Println(streams.GenerateCtx(ctx, func() int { return 1 }).Limit(2).Collect())
	fmt.Println(streams.IterateCtx(ctx, 1, func(n int) int { return n + 1 }).Limit(2).Collect())
	fmt.Println(streams.RangeCtx(ctx, 1, 3).Collect())
	ch := make(chan int, 2)
	ch <- 7
	ch <- 8
	close(ch)
	fmt.Println(streams.FromChannelCtx(ctx, ch).Collect())
	fmt.Println(streams.FromReaderLinesCtx(ctx, strings.NewReader("x\ny")).Collect())
	// Output:
	// 3
	// 1
	// [1 1]
	// [1 2]
	// [1 2]
	// [7 8]
	// [x y]
}

func Example_ctxTerminals_and_intermediate() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Terminals
	vals, err := streams.CollectCtx(ctx, streams.Of(1, 2, 3))
	fmt.Println(len(vals), err == nil)
	var sum int
	_ = streams.ForEachCtx(ctx, streams.Of(1, 2, 3), func(n int) { sum += n })
	fmt.Println(sum)
	acc, _ := streams.ReduceCtx(ctx, streams.Of(1, 2, 3), 0, func(a, b int) int { return a + b })
	fmt.Println(acc)
	f, _ := streams.FindFirstCtx(ctx, streams.Of(1, 2, 3), func(n int) bool { return n > 1 })
	fmt.Println(f.Get())
	am, _ := streams.AnyMatchCtx(ctx, streams.Of(1, 2), func(n int) bool { return n == 2 })
	al, _ := streams.AllMatchCtx(ctx, streams.Of(2, 4), func(n int) bool { return n%2 == 0 })
	cn, _ := streams.CountCtx(ctx, streams.Of("a", "b"))
	fmt.Println(am, al, cn)
	// Intermediate
	fil := streams.FilterCtx(ctx, streams.Of(1, 2, 3), func(n int) bool { return n > 1 }).Collect()
	mp := streams.MapCtx(ctx, streams.Of(1, 2), func(n int) int { return n * 2 }).Collect()
	mpt := streams.MapToCtx(ctx, streams.Of(1, 2), func(n int) string { return fmt.Sprint(n) }).Collect()
	fmt.Println(fil, mp, mpt)
	// Output:
	// 3 true
	// 6
	// 6
	// 2
	// true true 2
	// [2 3] [2 4] [1 2]
}

func Example_fromReaderLinesCtx_scanner() {
	sc := bufio.NewScanner(strings.NewReader("a\nb"))
	fmt.Println(len(streams.FromScannerErr(sc).Collect()))
	// Output:
	// 2
}
