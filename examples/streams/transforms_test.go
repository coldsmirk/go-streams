package streams_examples

import (
	"fmt"
	"iter"
	"strings"

	streams "github.com/coldsmirk/go-streams"
)

func Example_mapTo_flatMap_zip() {
	src := streams.FromSlice([]string{"go", "stream"})
	upper := streams.MapTo(src, strings.ToUpper).Collect()
	runes := streams.FlatMap(src, func(s string) streams.Stream[rune] { return streams.FromRunes(s) }).Collect()
	zip := streams.Zip(streams.Of("a", "b"), streams.Of(1, 2, 3)).Collect()
	fmt.Println(upper)
	fmt.Println(len(runes) > 0, len(zip)) // zip ends with shorter input
	// Output:
	// [GO STREAM]
	// true 2
}

func Example_window_chunk_scan_flatten() {
	w := streams.Window(streams.Range(1, 5), 3).Collect()
	ws := streams.WindowWithStep(streams.Range(1, 6), 3, 2, true).Collect()
	ch := streams.Chunk(streams.Of(1, 2, 3, 4, 5), 2).Collect()
	sc := streams.Scan(streams.Of(1, 2, 3), 0, func(a, b int) int { return a + b }).Collect()
	fl := streams.Flatten(streams.Of([]int{1, 2}, []int{3})).Collect()
	fmt.Println(w)
	fmt.Println(ws)
	fmt.Println(ch)
	fmt.Println(sc)
	fmt.Println(fl)
	// Output:
	// [[1 2 3] [2 3 4]]
	// [[1 2 3] [3 4 5] [5]]
	// [[1 2] [3 4] [5]]
	// [1 3 6]
	// [1 2 3]
}

func Example_interleave_pairwise_triples_flatmapSeq() {
	inter := streams.Interleave(streams.Of(1, 3, 5), streams.Of(2, 4, 6)).Collect()
	pairs := streams.Pairwise(streams.Of(1, 2, 3)).Collect()
	tris := streams.Triples(streams.Of(1, 2, 3, 4)).Collect()
	// FlatMapSeq using iter.Seq
	seq := func(n int) iter.Seq[int] {
		return func(yield func(int) bool) {
			for i := range n {
				if !yield(i) {
					return
				}
			}
		}
	}
	flatSeq := streams.FlatMapSeq(streams.Of(2, 3), seq).Collect()
	fmt.Println(inter)
	fmt.Println(pairs)
	fmt.Println(tris)
	fmt.Println(flatSeq)
	// Output:
	// [1 2 3 4 5 6]
	// [{1 2} {2 3}]
	// [{1 2 3} {2 3 4}]
	// [0 1 0 1 2]
}
