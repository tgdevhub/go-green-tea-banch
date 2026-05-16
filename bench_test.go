package testGC

import (
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"testing"
)

type Item struct {
	ID    int
	Value string
}

type Request struct {
	ID      int
	Path    string
	Headers map[string]string
	Body    []byte
}

type Node struct {
	Value int
	Next  *Node
}

func BenchmarkSmallAllocs(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		items := make([]*Item, 100)
		for j := range items {
			items[j] = &Item{ID: j, Value: "data"}
		}
		_ = items
	}
}

func BenchmarkMixedSizeAllocs(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		small := make([]byte, 16)
		medium := make([]byte, 256)
		large := make([]byte, 4096)
		_, _, _ = small, medium, large
	}
}

func BenchmarkHTTPRequestLike(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := &Request{
			ID:   i,
			Path: "/api/v1/users/" + strconv.Itoa(i),
			Headers: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer token",
				"User-Agent":    "bench/1.0",
				"Accept":        "*/*",
			},
			Body: []byte(`{"name":"test","value":42}`),
		}
		_ = req
	}
}

func BenchmarkConcurrentAllocs(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			items := make([]*Item, 50)
			for j := range items {
				items[j] = &Item{ID: j, Value: "x"}
			}
			_ = items
		}
	})
}

func BenchmarkPointerHeavy(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var head *Node
		for j := 0; j < 200; j++ {
			head = &Node{Value: j, Next: head}
		}
		_ = head
	}
}

func BenchmarkStringBuilding(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s := ""
		for j := 0; j < 50; j++ {
			s += "chunk-" + strconv.Itoa(j) + ";"
		}
		_ = s
	}
}

func BenchmarkLongLivedPlusGarbage(b *testing.B) {
	base := make([]*Item, 10_000)
	for i := range base {
		base[i] = &Item{ID: i, Value: "persistent"}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		tmp := make([]*Item, 100)
		for j := range tmp {
			tmp[j] = &Item{ID: j, Value: "transient"}
		}
		_ = tmp
	}

	runtime.KeepAlive(base)
}

func BenchmarkWithGCStats(b *testing.B) {
	b.ReportAllocs()

	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)

	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		for w := 0; w < 4; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				items := make([]*Item, 200)
				for j := range items {
					items[j] = &Item{ID: j, Value: "data"}
				}
				_ = items
			}()
		}
		wg.Wait()
	}

	runtime.ReadMemStats(&after)

	gcCount := after.NumGC - before.NumGC
	pauseTotalNs := after.PauseTotalNs - before.PauseTotalNs

	b.ReportMetric(float64(gcCount), "gc-cycles")
	if gcCount > 0 {
		b.ReportMetric(float64(pauseTotalNs)/float64(gcCount), "ns/gc-pause")
	}
}

var sink any

func init() {
	sink = fmt.Sprintf("%d", runtime.NumCPU())
}
