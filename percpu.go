package percpu

import (
	"runtime"
	"sync/atomic"
	"unsafe"
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

//with seq and pad, make sum()
type intSeq struct {
	intSeqInternal
	pad [128 - unsafe.Sizeof(intSeqInternal{})%128]byte
}

type intSeqInternal struct {
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

/*
// save mode: procPin()-->m.locks++, means  g don't schedule
func (p *perCpuIntSeq) Add(v int) {
	pid := procPin()
	p.vs[pid].v += v //should violatile:  use atomic, P's id does't mean run on the same cpu
	p.vs[pid].seq++
	procUnpin()
}

func (p *perCpuIntSeq) Dec(v int) {
	pid := procPin()
	p.vs[pid].v -= v //should violatile:  use atomic, P's id  does't mean run on the same cpu
	p.vs[pid].seq++
	procUnpin()
}
*/
//i think that g never be scheduled between "p.vs[pid].v += v" and "p.vs[pid].seq++"
func (p *perCpuIntSeq) Add(v int) {
	pid := GetPid()
	p.vs[pid].v += v //should violatile:  use atomic, P's id does't mean run on the same cpu
	p.vs[pid].seq++
}

func (p *perCpuIntSeq) Dec(v int) {
	pid := GetPid()
	p.vs[pid].v -= v //should violatile:  use atomic, P's id  does't mean run on the same cpu
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
