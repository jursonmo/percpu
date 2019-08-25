package percpu

import (
	"runtime"
	"sync/atomic"
	_ "unsafe"
)

//go:linkname procPin runtime.procPin
func procPin() int

//go:linkname procUnpin runtime.procUnpin
func procUnpin()

func GetPid() int {
	pid := procPin()
	procUnpin()
	return pid
}

type perCpuInt struct {
	vs []int
}

func NewIntVar() *perCpuInt {
	n := runtime.GOMAXPROCS(0)
	return &perCpuInt{make([]int, n)}
}

func (p *perCpuInt) Add(v int) {
	pid := GetPid()
	p.vs[pid] += v
}

func (p *perCpuInt) Dec(v int) {
	pid := GetPid()
	p.vs[pid] -= v
}

func (p *perCpuInt) Value() (s int) {

	for _, v := range p.vs {
		s += v
	}

	return
}

/*
type perCpuUint struct {
	vs []uint
}

func NewUintVar() *perCpuUint {
	n := runtime.GOMAXPROCS(0)
	return &perCpuUint{make([]uint, n)}
}
*/

//with seq, make sum()
type intSeq struct {
	v   int
	seq int32
}
type perCpuIntSeq struct {
	vs []intSeq
}

func NewIntSeqVar() *perCpuIntSeq {
	n := runtime.GOMAXPROCS(0)
	return &perCpuIntSeq{make([]intSeq, n)}
}

func (p *perCpuIntSeq) Add(v int) {
	pid := GetPid()
	p.vs[pid].v += v
	p.vs[pid].seq++
}

func (p *perCpuIntSeq) Dec(v int) {
	pid := GetPid()
	p.vs[pid].v -= v
	p.vs[pid].seq++
}

func (p *perCpuIntSeq) Value() (s int) {
retry:
	for i, v := range p.vs {
		s += v.v
		if v.seq == atomic.LoadInt32(&p.vs[i].seq) {
			s = 0
			goto retry
		}
	}

	return
}
