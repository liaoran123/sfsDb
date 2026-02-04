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

// GetKeyChangeStats 获取键值变化统计
// 返回:
//   KeyChangeStats: 键值变化统计信息

func (mm *MonitorManager) GetKeyChangeStats() KeyChangeStats {
	putCounters, deleteCounters := monitor.GetAllCounters()
	return KeyChangeStats{
		PutCounters:    putCounters,
		DeleteCounters: deleteCounters,
	}
}
