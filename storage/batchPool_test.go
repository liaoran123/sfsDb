package storage

import (
	"sync"
	"testing"

	"github.com/syndtr/goleveldb/leveldb"
)

// TestBatchPool_BasicOperations 测试批处理池的基本操作
func TestBatchPool_BasicOperations(t *testing.T) {
	// 测试获取批处理对象
	batch := LdbBatchPool.Get()
	if batch == nil {
		t.Error("Failed to get batch from pool")
	}

	// 测试放回批处理对象
	LdbBatchPool.Put(batch)

	// 再次获取批处理对象，应该能够成功
	batch2 := LdbBatchPool.Get()
	if batch2 == nil {
		t.Error("Failed to get batch from pool again")
	}

	LdbBatchPool.Put(batch2)
}

// TestBatchPool_SizeLimit 测试批处理池的大小限制
func TestBatchPool_SizeLimit(t *testing.T) {
	// 创建多个批处理对象，超过池大小限制
	var batches []*leveldb.Batch
	for i := 0; i < MaxBatchPoolSize*2; i++ {
		batch := LdbBatchPool.Get()
		if batch == nil {
			t.Error("Failed to get batch from pool")
		}
		batches = append(batches, batch)
	}

	// 放回所有批处理对象
	for _, batch := range batches {
		LdbBatchPool.Put(batch)
	}

	// 测试通过，只要能正常执行完操作即可
}

// TestBatchPool_ConcurrentOperations 测试批处理池的并发操作
func TestBatchPool_ConcurrentOperations(t *testing.T) {
	var wg sync.WaitGroup
	concurrentCount := 100
	operationCount := 100

	wg.Add(concurrentCount)

	for i := 0; i < concurrentCount; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < operationCount; j++ {
				batch := LdbBatchPool.Get()
				if batch != nil {
					LdbBatchPool.Put(batch)
				}
			}
		}()
	}

	wg.Wait()

	// 测试通过，只要能正常执行完并发操作即可
}

// TestBatchPool_SizeThreshold 测试批处理对象大小阈值
func TestBatchPool_SizeThreshold(t *testing.T) {
	// 创建一个大的批处理对象
	largeBatch := LdbBatchPool.Get()
	if largeBatch == nil {
		t.Error("Failed to get batch from pool")
	}

	// 向批处理对象中添加大量操作，使其大小超过阈值
	// 每个操作使用更大的键值对，确保总大小超过 1MB
	for i := 0; i < 100000; i++ {
		// 使用更大的键和值
		key := make([]byte, 100)   // 100字节键
		value := make([]byte, 100) // 100字节值
		for j := 0; j < 100; j++ {
			key[j] = byte((i + j) % 255)
			value[j] = byte((i + j) % 255)
		}
		largeBatch.Put(key, value)
	}

	// 放回大的批处理对象，应该被丢弃
	LdbBatchPool.Put(largeBatch)

	// 测试通过，只要能正常执行完操作即可
}

// BenchmarkBatchPool_Operations 基准测试批处理池操作性能
func BenchmarkBatchPool_Operations(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		batch := LdbBatchPool.Get()
		if batch != nil {
			LdbBatchPool.Put(batch)
		}
	}
}

// BenchmarkBatchPool_Concurrent 基准测试批处理池的并发性能
func BenchmarkBatchPool_Concurrent(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			batch := LdbBatchPool.Get()
			if batch != nil {
				LdbBatchPool.Put(batch)
			}
		}
	})
}
