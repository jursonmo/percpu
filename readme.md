
#### golang 版本的per_cpu

1. golang 的调度模式MPG，因为没有线程M执行任务时，优先必须绑定一个P, 所以P 就是并行的关键, 通过runtime.GOMAXPROCS(0)可以获取，一般等于cpu核数，代码有多少个任务同时进行

2. 通过sync.Pool源码学习可以看到，为了减少锁的竞争，sync.Pool的底层是由runtime.GOMAXPROCS(0)队列组成，每个任务 在Get()对象时，从当前procces_id(模式MPG中的P的id) 对应的private 里去对象时不需要加锁的,提高性能.

3. 做过内核开发的同学都知道，内核可以有per_cpu 变量，也是同样的原理，当前cpu操作其对应的变量，不需要加锁。
这个方法大量用到网卡统计信息时：
```c
	for_each_possible_cpu(cpu) {
		unsigned int start;
		const struct mwb_cpu_dev_stats *bstats = per_cpu_ptr(stats, cpu);
		do {
    			start = u64_stats_fetch_begin(&bstats->syncp);
    			memcpy(&tmp, bstats, sizeof(tmp));
		} while (u64_stats_fetch_retry(&bstats->syncp, start));
		sum.tx_bytes += tmp.tx_bytes;
		sum.rx_bytes += tmp.rx_bytes;
	}
```
这里可以注意到,有用到顺序计数功能。为了保证读写一致性。

4. 我想实现一个golang 版本的per_cpu 就可以利用上面的知识了：
 * 优先要能获取到procces_id，这时关键，runtime 里的procPin()方法可以获取到pid,但是它是非导出方法
 * //go:linkname 可以间接使用runtime 里的procPin()方法

 5. 统计信息一般用atomic来做，可以对比下per_cpu 性能
 ```
 $ go test -bench .
goos: windows
goarch: amd64
pkg: percpu
BenchmarkAtomic-4               300000000                5.48 ns/op
BenchmarkPercpu-4               200000000                6.03 ns/op
BenchmarkAtomicParallel-4       100000000               17.1 ns/op
BenchmarkPercpuParallel-4       200000000                9.28 ns/op
BenchmarkPercpuSeqParallel-4    200000000                9.72 ns/op

$ go test -bench . -cpu=1,2,4
goos: windows
goarch: amd64
pkg: percpu
BenchmarkAtomic                 300000000                5.54 ns/op
BenchmarkAtomic-2               300000000                5.56 ns/op
BenchmarkAtomic-4               300000000                5.51 ns/op
BenchmarkPercpu                 200000000                5.98 ns/op
BenchmarkPercpu-2               200000000                6.02 ns/op
BenchmarkPercpu-4               200000000                6.15 ns/op
BenchmarkAtomicParallel         200000000                8.12 ns/op
BenchmarkAtomicParallel-2       100000000               11.0 ns/op
BenchmarkAtomicParallel-4       100000000               15.7 ns/op
BenchmarkPercpuParallel         200000000                6.87 ns/op
BenchmarkPercpuParallel-2       200000000                8.56 ns/op
BenchmarkPercpuParallel-4       200000000                9.29 ns/op
BenchmarkPercpuSeqParallel      200000000                8.40 ns/op
BenchmarkPercpuSeqParallel-2    200000000                9.61 ns/op
BenchmarkPercpuSeqParallel-4    100000000               10.7 ns/op
PASS
ok      percpu  41.514s

```
当cpu可以看到atomic 性能随着cpu核数的递增而下降, 但是Percpu 方式几乎没有变化

参数说明:
```
-cpu 1,2,4
    Specify a list of GOMAXPROCS values for which the tests or
    benchmarks should be executed. The default is the current value of GOMAXPROCS.
    ```
   
