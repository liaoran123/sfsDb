package management

import (
	"time"

	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/management/backup"
	"github.com/liaoran123/sfsDb/management/config"
	"github.com/liaoran123/sfsDb/management/index"
	"github.com/liaoran123/sfsDb/management/monitor"
	"github.com/liaoran123/sfsDb/management/stats"
	"github.com/liaoran123/sfsDb/management/status"
	"github.com/liaoran123/sfsDb/management/system"
	"github.com/liaoran123/sfsDb/storage"
)

// Manager 数据库管理器
// 作为管理工具库的入口点，提供对各个管理模块的访问

type Manager struct {
	store storage.Store
	table *engine.Table // 可选的表实例，用于深度集成
}

// NewManager 创建数据库管理器
// 参数:
//   store: 存储实例，用于访问数据库
// 返回:
//   *Manager: 数据库管理器实例

func NewManager(store storage.Store) *Manager {
	return &Manager{
		store: store,
		table: nil,
	}
}

// NewManagerWithTable 创建带表实例的数据库管理器
// 参数:
//   store: 存储实例，用于访问数据库
//   table: 表实例，用于深度集成，获取更详细的表和索引信息
// 返回:
//   *Manager: 数据库管理器实例

func NewManagerWithTable(store storage.Store, table *engine.Table) *Manager {
	return &Manager{
		store: store,
		table: table,
	}
}

// GetStatus 获取数据库状态
// 返回:
//   status.StatusInfo: 数据库状态信息
//   error: 错误信息

func (m *Manager) GetStatus() (status.StatusInfo, error) {
	statusMgr := status.NewStatusManager(m.store)
	return statusMgr.GetStatus()
}

// IndexManager 获取索引管理器
// 返回:
//   *index.IndexManager: 索引管理器实例

func (m *Manager) IndexManager() *index.IndexManager {
	return index.NewIndexManager(m.store, m.table)
}

// StatsManager 获取性能统计管理器
// 返回:
//   *stats.StatsManager: 性能统计管理器实例

func (m *Manager) StatsManager() *stats.StatsManager {
	return stats.NewStatsManager(m.store)
}

// BackupManager 获取备份管理器
// 返回:
//   *backup.BackupManager: 备份管理器实例

func (m *Manager) BackupManager() *backup.BackupManager {
	return backup.NewBackupManager(m.store)
}

// ConfigManager 获取配置管理器
// 返回:
//   *config.ConfigManager: 配置管理器实例

func (m *Manager) ConfigManager() *config.ConfigManager {
	return config.NewConfigManager(m.store)
}

// Monitor 获取监控器
// 参数:
//   interval: 监控间隔
//   thresholds: 监控阈值
// 返回:
//   *Monitor: 监控器实例

func (m *Manager) Monitor(interval time.Duration, thresholds Thresholds) *Monitor {
	return NewMonitor(m, interval, thresholds)
}

// SystemManager 获取系统信息管理器
// 返回:
//   *system.SystemManager: 系统信息管理器实例

func (m *Manager) SystemManager() *system.SystemManager {
	return system.NewSystemManager(m.store)
}

// MonitorManager 获取监控管理器
// 返回:
//   *monitor.MonitorManager: 监控管理器实例

func (m *Manager) MonitorManager() *monitor.MonitorManager {
	return monitor.NewMonitorManager()
}

// GetTable 获取表实例
// 参数:
//   tableName: 表名
// 返回:
//   interface{}: 表实例

func (m *Manager) GetTable(tableName string) interface{} {
	// 尝试根据表名创建表实例
	// 注意：这里使用engine.TableNew创建表实例
	// 实际使用中，可能需要先检查表是否存在
	table, err := engine.TableNew(tableName)
	if err != nil {
		// 如果创建失败，返回nil
		return nil
	}
	return table
}

// Store 获取存储实例
// 返回:
//   storage.Store: 存储实例

func (m *Manager) Store() storage.Store {
	return m.store
}
