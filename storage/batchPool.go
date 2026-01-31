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

// 批处理对象池统计信息
type BatchPoolStats struct {
	TotalGets      uint64 // 总获取次数
	TotalPuts      uint64 // 总放回次数
	TotalCreates   uint64 // 总创建次数
	TotalDrops     uint64 // 总丢弃次数（超过大小限制）
	CurrentSize    uint64 // 当前池中对象数量
	MaxSizeReached uint64 // 达到最大容量的次数
}

// 批处理对象池统计信息
var batchPoolStats BatchPoolStats

// EnableBatchPoolTracing 是否启用批处理池跟踪
var EnableBatchPoolTracing bool

// BatchPoolTrace 批处理池跟踪函数类型
type BatchPoolTrace func(operation string, batch *leveldb.Batch, size int, stats BatchPoolStats)

// BatchPoolTracer 批处理池跟踪函数
var BatchPoolTracer BatchPoolTrace

// batchPool 包装sync.Pool，添加大小限制和统计功能
type batchPool struct {
	pool        sync.Pool
	currentSize uint64
	mu          sync.Mutex
}

// Get 从池中获取批处理对象
func (p *batchPool) Get() *leveldb.Batch {
	atomic.AddUint64(&batchPoolStats.TotalGets, 1)
	batch := p.pool.Get().(*leveldb.Batch)

	// 跟踪获取操作
	if EnableBatchPoolTracing && BatchPoolTracer != nil {
		stats := GetBatchPoolStats()
		BatchPoolTracer("get", batch, batch.Len(), stats)
	}

	return batch
}

// Put 将批处理对象放回池中
func (p *batchPool) Put(batch *leveldb.Batch) {
	atomic.AddUint64(&batchPoolStats.TotalPuts, 1)
	batchSize := batch.Len()

	// 检查批处理对象大小
	if batch.Len() > MaxBatchSize {
		atomic.AddUint64(&batchPoolStats.TotalDrops, 1)

		// 跟踪丢弃操作（大小超过限制）
		if EnableBatchPoolTracing && BatchPoolTracer != nil {
			stats := GetBatchPoolStats()
			BatchPoolTracer("drop_size", batch, batchSize, stats)
		}

		return
	}

	// 检查池大小
	p.mu.Lock()
	if p.currentSize >= MaxBatchPoolSize {
		p.mu.Unlock()
		atomic.AddUint64(&batchPoolStats.TotalDrops, 1)
		atomic.AddUint64(&batchPoolStats.MaxSizeReached, 1)

		// 跟踪丢弃操作（池已满）
		if EnableBatchPoolTracing && BatchPoolTracer != nil {
			stats := GetBatchPoolStats()
			BatchPoolTracer("drop_full", batch, batchSize, stats)
		}

		return
	}

	p.currentSize++
	p.mu.Unlock()

	// 重置批处理对象
	batch.Reset()

	// 将对象放回池中
	p.pool.Put(batch)

	// 更新统计信息
	atomic.StoreUint64(&batchPoolStats.CurrentSize, p.currentSize)

	// 跟踪放回操作
	if EnableBatchPoolTracing && BatchPoolTracer != nil {
		stats := GetBatchPoolStats()
		BatchPoolTracer("put", batch, 0, stats) // 重置后大小为0
	}
}

// 批处理对象池
var LdbBatchPool = &batchPool{
	pool: sync.Pool{
		New: func() any {
			atomic.AddUint64(&batchPoolStats.TotalCreates, 1)
			return new(leveldb.Batch)
		},
	},
}

// GetBatchPoolStats 获取批处理对象池统计信息
func GetBatchPoolStats() BatchPoolStats {
	return batchPoolStats
}
