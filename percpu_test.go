package percpu

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
)

func TestPercpu(t *testing.T) {
	var wg sync.WaitGroup
	p := NewIntVar()
	ng := 1000

	wg.Add(ng)
	for i := 0; i < ng; i++ {
		go func() {
			p.Add(1)
			wg.Done()
		}()
	}

	wg.Wait()
	t.Log(p.Value())
	if ng != p.Value() {
		t.Fatalf("total should be:%d, but actually is %d\n", ng, p.Value())
	}
}

func TestPercpuParallel(t *testing.T) {
	var wg sync.WaitGroup
	p := NewIntVar()
	np := runtime.GOMAXPROCS(0)
	loop := int(100000000)

	wg.Add(np)
	for i := 0; i < np; i++ {
		go func(index int) {
			t.Logf("go index:%d, pid:%d\n", index, GetPid())
			for j := 0; j < loop; j++ {
				p.Add(1)
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
	t.Log(p.Value())
	if np*loop != p.Value() {
		t.Fatalf("total should be:%d, but actually is %d\n", np*loop, p.Value())
	}
}

func BenchmarkAtomic(b *testing.B) {
	var s int64
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		atomic.AddInt64(&s, 1)
	}
}

func BenchmarkPercpu(b *testing.B) {
	p := NewIntVar()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		p.Add(1)
	}
}

func BenchmarkAtomicParallel(b *testing.B) {
	var s int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			atomic.AddInt64(&s, 1)
		}
	})
}
func BenchmarkPercpuParallel(b *testing.B) {
	p := NewIntVar()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			p.Add(1)
		}
	})
}
func BenchmarkPercpuSeqParallel(b *testing.B) {
	p := NewIntSeqVar()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			p.Add(1)
		}
	})
}
