package streams

import "runtime"

func countGoroutines() int {
	runtime.GC()
	return runtime.NumGoroutine()
}
