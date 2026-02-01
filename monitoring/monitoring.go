package monitoring

import (
	"time"

	"github.com/liaoran123/sfsDb/monitoring/alerts"
	"github.com/liaoran123/sfsDb/monitoring/exporters"
	"github.com/liaoran123/sfsDb/monitoring/metrics"
)

// Monitoring 监控系统接口
type Monitoring interface {
	// 指标相关
	Metrics() metrics.Metrics

	// 告警相关
	AlertManager() *alerts.AlertManager

	// 导出相关
	PrometheusExporter() *exporters.PrometheusExporter

	// 启动监控
	Start() error

	// 停止监控
	Stop() error
}

// Monitor 监控系统实现
type Monitor struct {
	metrics            metrics.Metrics
	alertManager       *alerts.AlertManager
	prometheusExporter *exporters.PrometheusExporter
}

// NewMonitor 创建监控系统
func NewMonitor() Monitoring {
	metrics := metrics.NewDefaultMetrics()
	alertManager := alerts.NewAlertManager()
	prometheusExporter := exporters.NewPrometheusExporter(metrics)

	// 添加默认告警处理器
	alertManager.AddHandler(alerts.NewConsoleAlertHandler())

	return &Monitor{
		metrics:            metrics,
		alertManager:       alertManager,
		prometheusExporter: prometheusExporter,
	}
}

// Metrics 获取指标收集器
func (m *Monitor) Metrics() metrics.Metrics {
	return m.metrics
}

// AlertManager 获取告警管理器
func (m *Monitor) AlertManager() *alerts.AlertManager {
	return m.alertManager
}

// PrometheusExporter 获取Prometheus导出器
func (m *Monitor) PrometheusExporter() *exporters.PrometheusExporter {
	return m.prometheusExporter
}

// Start 启动监控系统
func (m *Monitor) Start() error {
	// 启动Prometheus指标导出
	go func() {
		m.prometheusExporter.StartHTTP(":9090")
	}()

	// 启动告警检查
	go func() {
		for {
			m.alertManager.CheckAlerts()
			// 每5秒检查一次告警
			sleep(5)
		}
	}()

	return nil
}

// Stop 停止监控系统
func (m *Monitor) Stop() error {
	// 停止逻辑
	return nil
}

// sleep 辅助函数
func sleep(seconds int) {
	time.Sleep(time.Duration(seconds) * time.Second)
}
