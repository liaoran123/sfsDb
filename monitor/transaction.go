package monitor

import (
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"
)

func init() {
	GTransactionStatsMap = NewTransactionStatsMap()
}

var GTransactionStatsMap *TransactionStatsMap

type TransactionStatsMap struct {
	Data sync.Map // 使用 sync.Map 替代 map + 互斥锁
}

func NewTransactionStatsMap() *TransactionStatsMap {
	return &TransactionStatsMap{}
}

func (t *TransactionStatsMap) SetTime(txID uint64, duration time.Duration, tableName string, isolationLevel string, isCommitted bool) {
	value, ok := t.Data.Load(txID)
	if !ok {
		// 使用 LoadOrStore 避免竞态条件
		newValue := NewTransactionStats(txID, tableName, isolationLevel)
		newValue.IsCommitted = isCommitted
		value, ok = t.Data.LoadOrStore(txID, newValue)
		if !ok {
			value = newValue
		}
	}

	// 获取 TransactionStats 实例
	stats := value.(*TransactionStats)

	// 更新统计信息
	stats.AddDuration(duration)
}

// SetTimeAsync 异步记录事务耗时
func (t *TransactionStatsMap) SetTimeAsync(txID uint64, duration time.Duration, tableName string, isolationLevel string, isCommitted bool) {
	// 提交任务到全局 Pool
	globalPool.Submit(func() {
		t.SetTime(txID, duration, tableName, isolationLevel, isCommitted)
	})
}

// SetCount 记录事务操作计数和冲突计数
func (t *TransactionStatsMap) SetCount(txID uint64, operationCount, conflictCount int, tableName string, isolationLevel string, isCommitted bool) {
	value, ok := t.Data.Load(txID)
	if !ok {
		// 使用 LoadOrStore 避免竞态条件
		newValue := NewTransactionStats(txID, tableName, isolationLevel)
		newValue.IsCommitted = isCommitted
		value, ok = t.Data.LoadOrStore(txID, newValue)
		if !ok {
			value = newValue
		}
	}

	// 获取 TransactionStats 实例
	stats := value.(*TransactionStats)

	// 更新统计信息
	stats.TotalCount.Add(int64(operationCount))
}

// SetCountAsync 异步记录事务操作计数和冲突计数
func (t *TransactionStatsMap) SetCountAsync(txID uint64, operationCount, conflictCount int, tableName string, isolationLevel string, isCommitted bool) {
	// 提交任务到全局 Pool
	globalPool.Submit(func() {
		t.SetCount(txID, operationCount, conflictCount, tableName, isolationLevel, isCommitted)
	})
}

// GetAll 返回所有数据的普通 map 副本
func (t *TransactionStatsMap) GetAll() map[uint64]*TransactionStats {
	result := make(map[uint64]*TransactionStats)
	t.Data.Range(func(key, value interface{}) bool {
		result[key.(uint64)] = value.(*TransactionStats)
		return true
	})
	return result
}

type TransactionStats struct {
	// 事务ID
	TxID uint64 `json:"txID"`
	// 表名
	TableName string `json:"tableName"`
	// 隔离级别
	IsolationLevel string `json:"isolationLevel"`
	// 是否提交
	IsCommitted bool `json:"isCommitted"` //不是提交则是回滚
	// 事务总数
	TotalCount atomic.Int64 `json:"totalCount"`
	// 事务耗时总和（纳秒）
	Duration atomic.Int64 `json:"duration"` // 事务耗时总和，提交事务耗时
	// 事务耗时平均值
	AvgDuration time.Duration `json:"avgDuration"` // 事务耗时平均值，提交事务耗时平均值
	// 最大事务耗时（纳秒）
	MaxDuration atomic.Int64 `json:"maxDuration"` // 最大事务耗时，提交事务耗时
	// 最小事务耗时（纳秒）
	MinDuration atomic.Int64 `json:"minDuration"` // 最小事务耗时，提交事务耗时
}

func NewTransactionStats(txID uint64, tableName string, isolationLevel string) *TransactionStats {
	stats := &TransactionStats{
		TxID:           txID,
		TableName:      tableName,
		IsolationLevel: isolationLevel,
		IsCommitted:    true,
	}
	// 初始化最小时间为一个较大的值（纳秒）
	stats.MinDuration.Store(int64(time.Hour * 1))
	return stats
}

func (t *TransactionStats) AddDuration(duration time.Duration) {
	// 原子更新事务计数
	t.TotalCount.Add(1)

	// 原子更新总时间
	oldTotal := t.Duration.Load()
	newTotal := oldTotal + int64(duration)
	t.Duration.Store(newTotal)

	// 计算平均时间
	count := t.TotalCount.Load()
	if count > 0 {
		t.AvgDuration = time.Duration(newTotal / count)
	}

	// 原子更新最大时间
	oldMax := t.MaxDuration.Load()
	if int64(duration) > oldMax {
		t.MaxDuration.Store(int64(duration))
	}

	// 原子更新最小时间
	oldMin := t.MinDuration.Load()
	if int64(duration) < oldMin && duration > 0 {
		t.MinDuration.Store(int64(duration))
	}
}
func (t *TransactionStats) GetDuration() time.Duration {
	return time.Duration(t.Duration.Load())
}
func (t *TransactionStats) GetAvgDuration() time.Duration {
	if t.TotalCount.Load() == 0 {
		return 0
	}
	return t.AvgDuration
}
func (t *TransactionStats) GetMaxDuration() time.Duration {
	if t.TotalCount.Load() == 0 {
		return 0
	}
	return time.Duration(t.MaxDuration.Load())
}
func (t *TransactionStats) GetMinDuration() time.Duration {
	if t.TotalCount.Load() == 0 {
		return 0
	}
	return time.Duration(t.MinDuration.Load())
}
func (t *TransactionStats) GetTotalCount() int64 {
	return t.TotalCount.Load()
}

// MarshalJSON 自定义JSON序列化方法
func (t *TransactionStats) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"txID":           t.TxID,
		"tableName":      t.TableName,
		"isolationLevel": t.IsolationLevel,
		"isCommitted":    t.IsCommitted,
		"totalCount":     t.TotalCount.Load(),
		"duration":       time.Duration(t.Duration.Load()),
		"avgDuration":    t.AvgDuration,
		"maxDuration":    time.Duration(t.MaxDuration.Load()),
		"minDuration":    time.Duration(t.MinDuration.Load()),
	})
}
