package monitor

import (
	"github.com/liaoran123/sfsDb/monitor"
)

// KeyChangeStats 键值变化统计信息
type KeyChangeStats struct {
	PutCounters    map[string]int64 // put操作的计数器数据
	DeleteCounters map[string]int64 // delete操作的计数器数据
}

// MonitorManager 监控管理器
type MonitorManager struct{}

// NewMonitorManager 创建监控管理器
// 返回:
//   *MonitorManager: 监控管理器实例

func NewMonitorManager() *MonitorManager {
	return &MonitorManager{}
}

// GetKeyChangeStats 获取键值变化统计信息
// 返回:
//   map[int]*monitor.Keys: 键值变化统计信息

func (mm *MonitorManager) GetKeyChangeStats() map[int]*monitor.Keys {
	return monitor.GlobalKeysMap.GetAll()
}

// GetIndexStats 获取索引统计信息
// 返回:
//   map[int]*monitor.IndexStats: 索引统计信息映射

func (mm *MonitorManager) GetIndexStats() map[int]*monitor.IndexStats {
	return monitor.GIndexStatsMap.GetAll()
}

// GetTransactionStats 获取事务统计信息
// 返回:
//
//	map[uint64]*monitor.TransactionStats: 事务统计信息映射
func (mm *MonitorManager) GetTransactionStats() map[uint64]*monitor.TransactionStats {
	return monitor.GTransactionStatsMap.GetAll()
}
