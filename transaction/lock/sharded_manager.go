package lock

import (
	"hash/fnv"
	"time"
)

// ShardedManager 分片锁管理器
type ShardedManager struct {
	shards     []*Manager          // 锁管理器分片
	shardCount int                 // 分片数量
	hashFunc   func([]byte) uint32 // 哈希函数
}

// hashBytes 计算字节切片的哈希值
func hashBytes(data []byte) uint32 {
	h := fnv.New32()
	h.Write(data)
	return h.Sum32()
}

// NewShardedManager 创建新的分片锁管理器
func NewShardedManager(shardCount int) *ShardedManager {
	if shardCount <= 0 {
		shardCount = 16 // 默认分片数
	}

	shards := make([]*Manager, shardCount)
	for i := range shards {
		shards[i] = NewManager()
	}

	return &ShardedManager{
		shards:     shards,
		shardCount: shardCount,
		hashFunc:   hashBytes,
	}
}

// getShardIndex 根据键获取分片索引
func (sm *ShardedManager) getShardIndex(key string) int {
	hash := sm.hashFunc([]byte(key))
	return int(hash) % sm.shardCount
}

// AcquireLock 获取锁
func (sm *ShardedManager) AcquireLock(req LockRequest) error {
	shardIndex := sm.getShardIndex(req.Key)
	return sm.shards[shardIndex].AcquireLock(req)
}

// ReleaseLock 释放锁
func (sm *ShardedManager) ReleaseLock(key, requestor string) error {
	shardIndex := sm.getShardIndex(key)
	return sm.shards[shardIndex].ReleaseLock(key, requestor)
}

// ReleaseAllLocks 释放请求者的所有锁
func (sm *ShardedManager) ReleaseAllLocks(requestor string) error {
	for _, shard := range sm.shards {
		if err := shard.ReleaseAllLocks(requestor); err != nil {
			return err
		}
	}
	return nil
}

// GetLockStatus 获取锁状态
func (sm *ShardedManager) GetLockStatus(key string) (*LockStatus, error) {
	shardIndex := sm.getShardIndex(key)
	return sm.shards[shardIndex].GetLockStatus(key)
}

// GetAllLocks 获取所有锁
func (sm *ShardedManager) GetAllLocks() map[string]*LockStatus {
	allLocks := make(map[string]*LockStatus)
	for _, shard := range sm.shards {
		shardLocks := shard.GetAllLocks()
		for key, lock := range shardLocks {
			allLocks[key] = lock
		}
	}
	return allLocks
}

// GetStats 获取锁管理器统计信息
func (sm *ShardedManager) GetStats() map[string]interface{} {
	allStats := make(map[string]interface{})
	allStats["shardCount"] = sm.shardCount

	shardStats := make([]map[string]interface{}, sm.shardCount)
	for i, shard := range sm.shards {
		shardStats[i] = shard.GetStats()
	}
	allStats["shards"] = shardStats

	return allStats
}

// ClearExpiredLocks 清理过期锁
func (sm *ShardedManager) ClearExpiredLocks(expiration time.Duration) int {
	totalCleared := 0
	for _, shard := range sm.shards {
		cleared := shard.ClearExpiredLocks(expiration)
		totalCleared += cleared
	}
	return totalCleared
}

// IsLocked 检查键是否被锁定
func (sm *ShardedManager) IsLocked(key string) bool {
	shardIndex := sm.getShardIndex(key)
	return sm.shards[shardIndex].IsLocked(key)
}

// GetLockHolders 获取锁的持有者
func (sm *ShardedManager) GetLockHolders(key string) []string {
	shardIndex := sm.getShardIndex(key)
	return sm.shards[shardIndex].GetLockHolders(key)
}
