package percpu

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestPercpu(t *testing.T) {
	p := NewIntVar()

	for i := 0; i < 1000; i++ {
		go func() {
			p.Add(1)
		}()
	}

	time.Sleep(time.Second * 3)
	t.Log(p.Value())
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
