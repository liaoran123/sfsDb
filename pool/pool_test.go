package pool

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestPool 测试Pool的基本功能
func TestPool(t *testing.T) {
	// 创建工作池
	pool := NewPool()
	defer pool.Stop()

	// 测试任务数量
	const taskCount = 100
	var wg sync.WaitGroup
	results := make([]int, taskCount)

	// 提交任务
	for i := 0; i < taskCount; i++ {
		wg.Add(1)
		index := i
		pool.Submit(func() {
			defer wg.Done()
			// 简单的任务：睡眠1毫秒，然后设置结果
			time.Sleep(time.Millisecond)
			results[index] = index
		})
	}

	// 等待所有任务完成
	wg.Wait()

	// 验证结果
	for i := 0; i < taskCount; i++ {
		if results[i] != i {
			t.Errorf("任务%d执行失败，结果为%d，期望为%d", i, results[i], i)
		}
	}
}

// TestPoolConcurrency 测试Pool在不同并发数下的性能
func TestPoolConcurrency(t *testing.T) {
	// 测试不同并发数下的性能
	concurrencyLevels := []int{1, 2, 4, 8, 16, 32, 64, 128}

	for _, concurrency := range concurrencyLevels {
		t.Run(fmt.Sprintf("Concurrency_%d", concurrency), func(t *testing.T) {
			// 创建自定义并发数的工作池
			pool := NewPoolWithConcurrency(concurrency)
			defer pool.Stop()

			// 测试并发性能
			const taskCount = 1000
			var wg sync.WaitGroup
			start := time.Now()

			// 提交任务
			for i := 0; i < taskCount; i++ {
				wg.Add(1)
				pool.Submit(func() {
					defer wg.Done()
					// 简单的计算任务
					for j := 0; j < 100000; j++ {
						_ = j * j
					}
				})
			}

			// 等待所有任务完成
			wg.Wait()
			duration := time.Since(start)
			t.Logf("执行%d个任务，并发数%d，耗时%v", taskCount, concurrency, duration)
		})
	}
}

// TestPoolWithIO 测试Pool在IO密集型任务下的性能
func TestPoolWithIO(t *testing.T) {
	// 测试不同并发数下的IO密集型任务性能
	concurrencyLevels := []int{1, 2, 4, 8, 16, 32, 64, 128}

	for _, concurrency := range concurrencyLevels {
		t.Run(fmt.Sprintf("IO_Concurrency_%d", concurrency), func(t *testing.T) {
			// 创建自定义并发数的工作池
			pool := NewPoolWithConcurrency(concurrency)
			defer pool.Stop()

			// 测试IO密集型任务性能
			const taskCount = 100
			var wg sync.WaitGroup
			start := time.Now()

			// 提交任务
			for i := 0; i < taskCount; i++ {
				wg.Add(1)
				pool.Submit(func() {
					defer wg.Done()
					// 模拟IO密集型任务
					time.Sleep(10 * time.Millisecond)
				})
			}

			// 等待所有任务完成
			wg.Wait()
			duration := time.Since(start)
			t.Logf("执行%d个IO密集型任务，并发数%d，耗时%v", taskCount, concurrency, duration)
		})
	}
}

// NewPoolWithConcurrency 创建指定并发数的工作池
func NewPoolWithConcurrency(concurrency int) *Pool {
	p := &Pool{
		work: make(chan func(), concurrency),
	}
	for range concurrency {
		p.wg.Add(1)
		go p.worker()
	}
	return p
}
