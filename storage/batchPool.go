package storage

import (
	"sync"
	"sync/atomic"

	"github.com/syndtr/goleveldb/leveldb"
)

// 批处理对象池配置
const (
	// 最大批处理对象缓存数量
	MaxBatchPoolSize = 1000
	// 批处理对象大小阈值（字节），超过此值的批处理对象不会被缓存
	MaxBatchSize = 1024 * 1024 // 1MB
)

// batchPool 包装sync.Pool，添加大小限制和统计功能
type batchPool struct {
	pool        sync.Pool
	currentSize uint64 // 当前池中对象数量，使用原子操作管理
}

// Get 从池中获取批处理对象
func (p *batchPool) Get() *leveldb.Batch {
	batch := p.pool.Get().(*leveldb.Batch)
	if batch == nil {
		// 安全保障：如果从池中获取到 nil，创建一个新的批处理对象
		return new(leveldb.Batch)
	}
	// 确保返回的是干净的批处理对象
	batch.Reset()
	return batch
}

// Put 将批处理对象放回池中
func (p *batchPool) Put(batch *leveldb.Batch) {
	// 检查批处理对象是否为 nil
	if batch == nil {
		// 批处理对象为 nil，直接返回
		return
	}

	// 检查批处理对象大小
	if batch.Len() > MaxBatchSize {
		// 批处理对象大小超过阈值，直接丢弃，让垃圾回收器处理这个对象
		batch.Reset()
		return
	}

	// 检查池大小（使用原子操作）
	if atomic.LoadUint64(&p.currentSize) >= MaxBatchPoolSize {
		// 池大小超过限制，直接返回，让垃圾回收器处理这个对象
		batch.Reset()
		return
	}

	// 原子递增计数器
	if atomic.AddUint64(&p.currentSize, 1) > MaxBatchPoolSize {
		// 如果递增后超过限制，立即递减
		atomic.AddUint64(&p.currentSize, ^uint64(0))
		batch.Reset()
		return
	}

	// 重置批处理对象
	batch.Reset()

	// 将对象放回池中
	p.pool.Put(batch)
}

// 批处理对象池
var LdbBatchPool = &batchPool{
	pool: sync.Pool{
		New: func() any {
			return new(leveldb.Batch)
		},
	},
}

// GetPoolSize 获取当前批处理对象池的大小
func (p *batchPool) GetPoolSize() uint64 {
	return atomic.LoadUint64(&p.currentSize)
}

// GetLdbBatchPoolSize 获取全局批处理对象池的大小
func GetLdbBatchPoolSize() uint64 {
	return LdbBatchPool.GetPoolSize()
}
