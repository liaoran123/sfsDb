package web

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Metrics 性能指标
type Metrics struct {
	Timestamp      time.Time         `json:"timestamp"`
	RequestCount   int64            `json:"request_count"`
	ErrorCount     int64            `json:"error_count"`
	AvgResponseTime float64          `json:"avg_response_time"`
	MaxResponseTime float64          `json:"max_response_time"`
	MinResponseTime float64          `json:"min_response_time"`
	MemoryUsage    map[string]uint64 `json:"memory_usage"`
	CPUUsage       float64          `json:"cpu_usage"`
	DiskUsage      map[string]uint64 `json:"disk_usage"`
	NetworkStats   map[string]uint64 `json:"network_stats"`
}

// Alert 告警信息
type Alert struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Metrics   Metrics   `json:"metrics"`
}

// AlertThreshold 告警阈值
type AlertThreshold struct {
	Metric    string  `json:"metric"`
	Operator  string  `json:"operator"` // >, <, >=, <=, ==
	Threshold float64 `json:"threshold"`
	Level     string  `json:"level"` // info, warning, error, critical
}

// MetricsManager 指标管理器
type MetricsManager struct {
	metrics        []Metrics
	alerts         []Alert
	thresholds     []AlertThreshold
	requestCounter int64
	errorCounter   int64
	responseTimes  []float64
	mutex          sync.RWMutex
}

// NewMetricsManager 创建指标管理器
func NewMetricsManager() *MetricsManager {
	return &MetricsManager{
		metrics:        make([]Metrics, 0),
		alerts:         make([]Alert, 0),
		thresholds:     make([]AlertThreshold, 0),
		responseTimes:  make([]float64, 0),
	}
}

// RecordRequest 记录请求
func (mm *MetricsManager) RecordRequest(duration float64, isError bool) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	mm.requestCounter++
	mm.responseTimes = append(mm.responseTimes, duration)

	if len(mm.responseTimes) > 1000 {
		// 保留最近1000个响应时间
		mm.responseTimes = mm.responseTimes[len(mm.responseTimes)-1000:]
	}

	if isError {
		mm.errorCounter++
	}
}

// CollectMetrics 采集指标
func (mm *MetricsManager) CollectMetrics() Metrics {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	// 计算响应时间统计
	var totalResponseTime float64
	var maxResponseTime float64
	var minResponseTime float64

	if len(mm.responseTimes) > 0 {
		minResponseTime = mm.responseTimes[0]
		for _, t := range mm.responseTimes {
			totalResponseTime += t
			if t > maxResponseTime {
				maxResponseTime = t
			}
			if t < minResponseTime {
				minResponseTime = t
			}
		}
	}

	avgResponseTime := 0.0
	if len(mm.responseTimes) > 0 {
		avgResponseTime = totalResponseTime / float64(len(mm.responseTimes))
	}

	// 创建指标
	metrics := Metrics{
		Timestamp:       time.Now(),
		RequestCount:    mm.requestCounter,
		ErrorCount:      mm.errorCounter,
		AvgResponseTime: avgResponseTime,
		MaxResponseTime: maxResponseTime,
		MinResponseTime: minResponseTime,
		MemoryUsage:     getMemoryUsage(),
		CPUUsage:        getCPUUsage(),
		DiskUsage:       getDiskUsage(),
		NetworkStats:    getNetworkStats(),
	}

	// 存储指标
	mm.metrics = append(mm.metrics, metrics)
	if len(mm.metrics) > 1008 { // 保留7天的指标，每小时一个
		mm.metrics = mm.metrics[len(mm.metrics)-1008:]
	}

	// 检查告警
	mm.checkAlerts(metrics)

	return metrics
}

// GetMetrics 获取指标
func (mm *MetricsManager) GetMetrics(since time.Time) []Metrics {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	var result []Metrics
	for _, m := range mm.metrics {
		if m.Timestamp.After(since) {
			result = append(result, m)
		}
	}

	return result
}

// GetAlerts 获取告警
func (mm *MetricsManager) GetAlerts(since time.Time) []Alert {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	var result []Alert
	for _, a := range mm.alerts {
		if a.Timestamp.After(since) {
			result = append(result, a)
		}
	}

	return result
}

// SetThresholds 设置告警阈值
func (mm *MetricsManager) SetThresholds(thresholds []AlertThreshold) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	mm.thresholds = thresholds
}

// GetThresholds 获取告警阈值
func (mm *MetricsManager) GetThresholds() []AlertThreshold {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	return mm.thresholds
}

// checkAlerts 检查告警
func (mm *MetricsManager) checkAlerts(metrics Metrics) {
	for _, threshold := range mm.thresholds {
		var value float64

		// 获取指标值
		switch threshold.Metric {
		case "request_count":
			value = float64(metrics.RequestCount)
		case "error_count":
			value = float64(metrics.ErrorCount)
		case "avg_response_time":
			value = metrics.AvgResponseTime
		case "max_response_time":
			value = metrics.MaxResponseTime
		case "memory_usage":
			// 使用内存使用总量
			value = float64(metrics.MemoryUsage["total"])
		case "cpu_usage":
			value = metrics.CPUUsage
		default:
			continue
		}

		// 检查阈值
		var triggered bool
		switch threshold.Operator {
		case ">":
			triggered = value > threshold.Threshold
		case "<":
			triggered = value < threshold.Threshold
		case ">=":
			triggered = value >= threshold.Threshold
		case "<=":
			triggered = value <= threshold.Threshold
		case "==":
			triggered = value == threshold.Threshold
		default:
			continue
		}

		if triggered {
			// 创建告警
			alert := Alert{
				ID:        time.Now().Format("20060102150405"),
				Timestamp: time.Now(),
				Level:     threshold.Level,
				Message:   threshold.Metric + " " + threshold.Operator + " " + fmt.Sprintf("%f", threshold.Threshold),
				Metrics:   metrics,
			}

			mm.alerts = append(mm.alerts, alert)
			if len(mm.alerts) > 100 {
				mm.alerts = mm.alerts[len(mm.alerts)-100:]
			}
		}
	}
}

// 模拟获取内存使用情况
func getMemoryUsage() map[string]uint64 {
	// 实际实现中，应该使用系统API获取真实的内存使用情况
	return map[string]uint64{
		"total":       8 * 1024 * 1024 * 1024, // 8GB
		"used":        4 * 1024 * 1024 * 1024, // 4GB
		"free":        4 * 1024 * 1024 * 1024, // 4GB
		"cached":      1 * 1024 * 1024 * 1024, // 1GB
		"buffers":     512 * 1024 * 1024,      // 512MB
		"swap_total":  2 * 1024 * 1024 * 1024, // 2GB
		"swap_used":   512 * 1024 * 1024,      // 512MB
		"swap_free":   1536 * 1024 * 1024,     // 1.5GB
	}
}

// 模拟获取CPU使用情况
func getCPUUsage() float64 {
	// 实际实现中，应该使用系统API获取真实的CPU使用情况
	return 45.5 // 45.5%
}

// 模拟获取磁盘使用情况
func getDiskUsage() map[string]uint64 {
	// 实际实现中，应该使用系统API获取真实的磁盘使用情况
	return map[string]uint64{
		"total": 500 * 1024 * 1024 * 1024, // 500GB
		"used":  200 * 1024 * 1024 * 1024, // 200GB
		"free":  300 * 1024 * 1024 * 1024, // 300GB
	}
}

// 模拟获取网络统计信息
func getNetworkStats() map[string]uint64 {
	// 实际实现中，应该使用系统API获取真实的网络统计信息
	return map[string]uint64{
		"bytes_sent":     1024 * 1024 * 1024, // 1GB
		"bytes_recv":     2048 * 1024 * 1024, // 2GB
		"packets_sent":   1000000,
		"packets_recv":   2000000,
		"errin":          100,
		"errout":         50,
		"dropin":         10,
		"dropout":        5,
	}
}

// ResponseTimeMiddleware 响应时间中间件
func ResponseTimeMiddleware(metricsManager *MetricsManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		end := time.Now()
		latency := end.Sub(start).Seconds() * 1000 // 转换为毫秒

		// 检查是否有错误
		hasError := len(c.Errors) > 0

		// 记录请求
		metricsManager.RecordRequest(latency, hasError)
	}
}

// handleGetMetrics 处理获取指标请求
func (s *Server) handleGetMetrics(c *gin.Context) {
	// 解析时间范围
	sinceStr := c.DefaultQuery("since", "1h")
	since, err := time.ParseDuration(sinceStr)
	if err != nil {
		since = 1 * time.Hour
	}

	// 获取指标
	metrics := s.metricsManager.GetMetrics(time.Now().Add(-since))

	s.sendSuccess(c, gin.H{
		"metrics":  metrics,
		"count":    len(metrics),
		"since":    time.Now().Add(-since).Format(time.RFC3339),
		"until":    time.Now().Format(time.RFC3339),
		"message":  "Metrics retrieved successfully",
	})
}

// handleGetAlerts 处理获取告警请求
func (s *Server) handleGetAlerts(c *gin.Context) {
	// 解析时间范围
	sinceStr := c.DefaultQuery("since", "1h")
	since, err := time.ParseDuration(sinceStr)
	if err != nil {
		since = 1 * time.Hour
	}

	// 获取告警
	alerts := s.metricsManager.GetAlerts(time.Now().Add(-since))

	s.sendSuccess(c, gin.H{
		"alerts":  alerts,
		"count":   len(alerts),
		"since":   time.Now().Add(-since).Format(time.RFC3339),
		"until":   time.Now().Format(time.RFC3339),
		"message": "Alerts retrieved successfully",
	})
}

// handleSetThresholds 处理设置告警阈值请求
func (s *Server) handleSetThresholds(c *gin.Context) {
	var thresholds []AlertThreshold
	if err := c.ShouldBindJSON(&thresholds); err != nil {
		s.sendError(c, http.StatusBadRequest, "invalid request", "Invalid request format", err.Error())
		return
	}

	// 设置告警阈值
	s.metricsManager.SetThresholds(thresholds)

	s.sendSuccess(c, gin.H{
		"thresholds": thresholds,
		"message":    "Thresholds set successfully",
	})
}

// handleGetThresholds 处理获取告警阈值请求
func (s *Server) handleGetThresholds(c *gin.Context) {
	// 获取告警阈值
	thresholds := s.metricsManager.GetThresholds()

	s.sendSuccess(c, gin.H{
		"thresholds": thresholds,
		"message":    "Thresholds retrieved successfully",
	})
}

// handleCollectMetrics 处理采集指标请求
func (s *Server) handleCollectMetrics(c *gin.Context) {
	// 采集指标
	metrics := s.metricsManager.CollectMetrics()

	s.sendSuccess(c, gin.H{
		"metrics": metrics,
		"message": "Metrics collected successfully",
	})
}
