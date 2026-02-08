package engine

import (
	"sync"
)

// Snapshot 事务快照
// 用于记录事务开始时的数据库状态，支持多版本并发控制
// 注意：当前实现是一个简化版本，实际应用中可能需要更复杂的实现
type Snapshot struct {
	id        uint64 // 快照ID
	timestamp int64  // 快照创建时间戳
	// 可以添加更多字段，如：
	// - 表级快照信息
	// - 索引快照信息
	// - 事务隔离级别
}

// SnapshotPool 快照池
// 用于管理和复用Snapshot实例，减少内存分配和GC压力
type SnapshotPool struct {
	pool sync.Pool
}

// NewSnapshotPool 创建一个新的快照池
func NewSnapshotPool() *SnapshotPool {
	return &SnapshotPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &Snapshot{}
			},
		},
	}
}

// Get 从池中获取一个快照实例
func (sp *SnapshotPool) Get() *Snapshot {
	return sp.pool.Get().(*Snapshot)
}

// Put 将快照实例放回池中
func (sp *SnapshotPool) Put(snapshot *Snapshot) {
	// 重置快照状态，以便重用
	snapshot.id = 0
	snapshot.timestamp = 0
	// 重置其他字段...

	sp.pool.Put(snapshot)
}

// 全局快照池实例
var GlobalSnapshotPool = NewSnapshotPool()
