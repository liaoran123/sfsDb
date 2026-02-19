package management

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/management/backup"
	"github.com/liaoran123/sfsDb/management/config"
	"github.com/liaoran123/sfsDb/management/importexport"
	"github.com/liaoran123/sfsDb/management/monitor"

	"github.com/liaoran123/sfsDb/management/status"
	"github.com/liaoran123/sfsDb/management/system"
	"github.com/liaoran123/sfsDb/storage"
)

// Manager 数据库管理器
// 作为管理工具库的入口点，提供对各个管理模块的访问

type Manager struct {
	store storage.Store
	table *engine.Table // 可选的表实例，用于深度集成
	//statsManager *stats.StatsManager
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
		//statsManager: stats.NewStatsManager(store),
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
		//statsManager: stats.NewStatsManager(store),
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

// StatsManager 获取性能统计管理器
// 返回:
//   *stats.StatsManager: 性能统计管理器实例

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

// ImportExportManager 获取导入导出管理器
// 返回:
//   *importexport.ImportExportManager: 导入导出管理器实例

func (m *Manager) ImportExportManager() *importexport.ImportExportManager {
	return importexport.NewImportExportManager(m.store)
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
//
//	storage.Store: 存储实例

func (m *Manager) Store() storage.Store {
	return m.store
}

// HTTPClientManager HTTP客户端管理器
// 用于通过HTTP API访问远程数据库管理功能
type HTTPClientManager struct {
	baseURL string
	client  *http.Client
}

// NewHTTPClientManager 创建HTTP客户端管理器
// 参数:
//
//	baseURL: API服务器基础URL
//
// 返回:
//
//	*HTTPClientManager: HTTP客户端管理器实例

func NewHTTPClientManager(baseURL string) *HTTPClientManager {
	return &HTTPClientManager{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

// GetStatus 获取数据库状态
// 返回:
//
//	status.StatusInfo: 数据库状态信息
//	error: 错误信息

func (m *HTTPClientManager) GetStatus() (status.StatusInfo, error) {
	var statusInfo status.StatusInfo
	resp, err := m.client.Get(m.baseURL + "/api/status")
	if err != nil {
		return statusInfo, err
	}
	defer resp.Body.Close()

	var response struct {
		Memory  map[string]interface{}
		Storage map[string]interface{}
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return statusInfo, err
	}

	if memory, ok := response.Memory["Alloc"].(float64); ok {
		statusInfo.Memory.Alloc = uint64(memory)
	}
	if memory, ok := response.Memory["TotalAlloc"].(float64); ok {
		statusInfo.Memory.TotalAlloc = uint64(memory)
	}
	if memory, ok := response.Memory["Sys"].(float64); ok {
		statusInfo.Memory.Sys = uint64(memory)
	}
	if memory, ok := response.Memory["NumGC"].(float64); ok {
		statusInfo.Memory.NumGC = uint32(memory)
	}
	if storage, ok := response.Storage["StoreType"].(string); ok {
		statusInfo.Storage.StoreType = storage
	}

	return statusInfo, nil
}

// BackupManager 获取备份管理器
// 返回:
//
//	*backup.BackupManager: 备份管理器实例

func (m *HTTPClientManager) BackupManager() *backup.BackupManager {
	return backup.NewBackupManager(nil)
}

// ConfigManager 获取配置管理器
// 返回:
//
//	*config.ConfigManager: 配置管理器实例

func (m *HTTPClientManager) ConfigManager() *config.ConfigManager {
	return config.NewConfigManager(nil)
}

// SystemManager 获取系统信息管理器
// 返回:
//
//	*system.SystemManager: 系统信息管理器实例

func (m *HTTPClientManager) SystemManager() *system.SystemManager {
	return system.NewSystemManager(nil)
}

// MonitorManager 获取监控管理器
// 返回:
//
//	*monitor.MonitorManager: 监控管理器实例

func (m *HTTPClientManager) MonitorManager() *monitor.MonitorManager {
	return monitor.NewMonitorManager()
}

// Monitor 获取监控器
// 参数:
//
//	interval: 监控间隔
//	thresholds: 监控阈值
//
// 返回:
//
//	*Monitor: 监控器实例

func (m *HTTPClientManager) Monitor(interval time.Duration, thresholds Thresholds) *Monitor {
	return NewMonitor(nil, interval, thresholds)
}

// GetTable 获取表实例
// 参数:
//
//	tableName: 表名
//
// 返回:
//
//	interface{}: 表实例

func (m *HTTPClientManager) GetTable(tableName string) interface{} {
	return nil
}

// Store 获取存储实例
// 返回:
//
//	storage.Store: 存储实例

func (m *HTTPClientManager) Store() storage.Store {
	return nil
}
