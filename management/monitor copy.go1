package management

import (
	"fmt"
	"log"
	"time"

	"github.com/liaoran123/sfsDb/management/status"
)

// Thresholds 监控阈值
type Thresholds struct {
	MemoryUsage float64 // 内存使用阈值（MB）
	GCCount     int     // GC 次数阈值
	// 其他阈值可以根据需要扩展
}

// AlertNotifier 告警通知器接口
type AlertNotifier interface {
	Notify(message string) error
}

// LogNotifier 日志通知器
type LogNotifier struct{}

// Notify 通过日志通知告警
func (ln *LogNotifier) Notify(message string) error {
	log.Printf("[ALERT] %s\n", message)
	return nil
}

// Monitor 监控器
type Monitor struct {
	manager    *Manager
	interval   time.Duration
	thresholds Thresholds
	notifier   AlertNotifier
	running    bool
	// 其他监控相关字段
}

// NewMonitor 创建监控器
// 参数:
//   manager: 管理器实例
//   interval: 监控间隔
//   thresholds: 监控阈值
// 返回:
//   *Monitor: 监控器实例

func NewMonitor(manager *Manager, interval time.Duration, thresholds Thresholds) *Monitor {
	return &Monitor{
		manager:    manager,
		interval:   interval,
		thresholds: thresholds,
		notifier:   &LogNotifier{},
		running:    false,
	}
}

// SetNotifier 设置告警通知器
// 参数:
//   notifier: 告警通知器

func (m *Monitor) SetNotifier(notifier AlertNotifier) {
	m.notifier = notifier
}

// Start 启动监控
// 返回:
//   error: 错误信息

func (m *Monitor) Start() error {
	if m.running {
		return fmt.Errorf("监控已经在运行")
	}

	m.running = true

	// 启动监控协程
	go m.monitorLoop()

	log.Println("监控已启动")
	return nil
}

// Stop 停止监控

func (m *Monitor) Stop() {
	m.running = false
	log.Println("监控已停止")
}

// monitorLoop 监控循环

func (m *Monitor) monitorLoop() {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for m.running {
		select {
		case <-ticker.C:
			m.checkStatus()
		}
	}
}

// checkStatus 检查状态并触发告警

func (m *Monitor) checkStatus() {
	// 获取数据库状态
	statusInfo, err := m.manager.GetStatus()
	if err != nil {
		log.Printf("获取状态失败: %v\n", err)
		return
	}

	// 检查内存使用
	memoryUsage := float64(statusInfo.Memory.Alloc) / 1024 / 1024
	if memoryUsage > m.thresholds.MemoryUsage {
		alertMsg := fmt.Sprintf("内存使用超过阈值: %.2f MB > %.2f MB", memoryUsage, m.thresholds.MemoryUsage)
		m.notifier.Notify(alertMsg)
	}

	// 检查 GC 次数
	gcCount := int(statusInfo.Memory.NumGC)
	if gcCount > m.thresholds.GCCount {
		alertMsg := fmt.Sprintf("GC 次数超过阈值: %d > %d", gcCount, m.thresholds.GCCount)
		m.notifier.Notify(alertMsg)
	}

	// 其他检查可以根据需要扩展
}

// GetCurrentStatus 获取当前状态
// 返回:
//   status.StatusInfo: 状态信息
//   error: 错误信息

func (m *Monitor) GetCurrentStatus() (status.StatusInfo, error) {
	return m.manager.GetStatus()
}
