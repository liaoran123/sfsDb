package monitor

import (
	"runtime"
	"sync"
)

// Pool 进程池
type Pool struct {
	wg   sync.WaitGroup
	work chan func()
	size int // 工作协程数量
}

// NewPool 创建并返回一个拥有最佳并发数的 Pool
// 默认使用 CPU 核心数作为并发数，适合 CPU 密集型任务
func NewPool() *Pool {
	return NewPoolWithSize(runtime.NumCPU())
}

// NewPoolWithSize 创建并返回一个指定大小的 Pool
// 参数 size: 工作协程数量
// - CPU密集型任务: 建议设置为 runtime.NumCPU() 或 runtime.NumCPU()+1
// - IO密集型任务: 建议设置为 runtime.NumCPU()*2 至 runtime.NumCPU()*10
// - 网络IO密集型: 建议设置为 runtime.NumCPU()*5 至 runtime.NumCPU()*20
func NewPoolWithSize(size int) *Pool {
	if size <= 0 {
		size = runtime.NumCPU()
	}
	p := &Pool{
		work: make(chan func(), size),
		size: size,
	}
	for i := 0; i < size; i++ {
		p.wg.Add(1)
		go p.worker()
	}
	return p
}

// NewPoolForCPU 创建适合 CPU 密集型任务的 Pool
// 并发数: runtime.NumCPU()
func NewPoolForCPU() *Pool {
	return NewPoolWithSize(runtime.NumCPU())
}

// NewPoolForIO 创建适合 IO 密集型任务的 Pool
// 并发数: runtime.NumCPU() * 4
func NewPoolForIO() *Pool {
	return NewPoolWithSize(runtime.NumCPU() * 4)
}

// NewPoolForNetwork 创建适合网络 IO 密集型任务的 Pool
// 并发数: runtime.NumCPU() * 10
func NewPoolForNetwork() *Pool {
	return NewPoolWithSize(runtime.NumCPU() * 10)
}

// worker 持续从 work 通道中取出任务并执行
func (p *Pool) worker() {
	defer p.wg.Done()
	for task := range p.work {
		task()
	}
}

// Submit 向 Pool 提交一个任务
func (p *Pool) Submit(task func()) {
	p.work <- task
}

// Stop 关闭 work 通道并等待所有 worker 退出
func (p *Pool) Stop() {
	close(p.work)
	p.wg.Wait()
}

// Size 返回当前 Pool 的工作协程数量
func (p *Pool) Size() int {
	return p.size
}
