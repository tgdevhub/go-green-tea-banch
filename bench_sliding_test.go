package testGC

import (
	"runtime"
	"testing"
)

func BenchmarkSlidingWindow(b *testing.B) {
	b.ReportAllocs()
	window := make([]*Item, 10_000)
	idx := 0
	for i := 0; i < b.N; i++ {
		window[idx%len(window)] = &Item{ID: i, Value: "sliding"}
		idx++
	}
	runtime.KeepAlive(window)
}

func BenchmarkLargeHeapPressure(b *testing.B) {
	live := make([][]*Item, 1000)
	for i := range live {
		live[i] = make([]*Item, 100)
		for j := range live[i] {
			live[i][j] = &Item{ID: j, Value: "resident"}
		}
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			tmp := make([]*Item, 500)
			for j := range tmp {
				tmp[j] = &Item{ID: j, Value: "pressure"}
			}
			_ = tmp
		}
	})
	runtime.KeepAlive(live)
}
