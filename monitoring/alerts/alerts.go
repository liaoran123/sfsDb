package alerts

import (
	"fmt"
	"time"
)

// AlertLevel 告警级别
type AlertLevel string

const (
	// AlertLevelInfo 信息级告警
	AlertLevelInfo AlertLevel = "INFO"
	// AlertLevelWarning 警告级告警
	AlertLevelWarning AlertLevel = "WARNING"
	// AlertLevelError 错误级告警
	AlertLevelError AlertLevel = "ERROR"
	// AlertLevelCritical 严重级告警
	AlertLevelCritical AlertLevel = "CRITICAL"
)

// Alert 告警信息
type Alert struct {
	ID        string      `json:"id"`
	Level     AlertLevel  `json:"level"`
	Message   string      `json:"message"`
	Timestamp time.Time   `json:"timestamp"`
	Labels    map[string]string `json:"labels"`
}

// AlertRule 告警规则
type AlertRule struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Level       AlertLevel  `json:"level"`
	Condition   func() bool `json:"-"` // 告警条件函数
	Labels      map[string]string `json:"labels"`
}

// AlertHandler 告警处理器接口
type AlertHandler interface {
	HandleAlert(alert Alert)
}

// AlertManager 告警管理器
type AlertManager struct {
	rules    []AlertRule
	handlers []AlertHandler
	alerts   map[string]Alert
}

// NewAlertManager 创建告警管理器
func NewAlertManager() *AlertManager {
	return &AlertManager{
		rules:    make([]AlertRule, 0),
		handlers: make([]AlertHandler, 0),
		alerts:   make(map[string]Alert),
	}
}

// AddRule 添加告警规则
func (am *AlertManager) AddRule(rule AlertRule) {
	am.rules = append(am.rules, rule)
}

// AddHandler 添加告警处理器
func (am *AlertManager) AddHandler(handler AlertHandler) {
	am.handlers = append(am.handlers, handler)
}

// CheckAlerts 检查告警规则
func (am *AlertManager) CheckAlerts() {
	for _, rule := range am.rules {
		if rule.Condition() {
			alert := Alert{
				ID:        fmt.Sprintf("%s_%d", rule.Name, time.Now().Unix()),
				Level:     rule.Level,
				Message:   rule.Description,
				Timestamp: time.Now(),
				Labels:    rule.Labels,
			}

			// 检查告警是否已存在
			if _, exists := am.alerts[alert.ID]; !exists {
				am.alerts[alert.ID] = alert
				am.dispatchAlert(alert)
			}
		}
	}
}

// dispatchAlert 分发告警
func (am *AlertManager) dispatchAlert(alert Alert) {
	for _, handler := range am.handlers {
		handler.HandleAlert(alert)
	}
}

// GetActiveAlerts 获取活跃告警
func (am *AlertManager) GetActiveAlerts() []Alert {
	alerts := make([]Alert, 0, len(am.alerts))
	for _, alert := range am.alerts {
		alerts = append(alerts, alert)
	}
	return alerts
}

// ClearAlert 清除告警
func (am *AlertManager) ClearAlert(alertID string) {
	delete(am.alerts, alertID)
}

// ClearAllAlerts 清除所有告警
func (am *AlertManager) ClearAllAlerts() {
	am.alerts = make(map[string]Alert)
}

// ConsoleAlertHandler 控制台告警处理器
type ConsoleAlertHandler struct {}

// NewConsoleAlertHandler 创建控制台告警处理器
func NewConsoleAlertHandler() *ConsoleAlertHandler {
	return &ConsoleAlertHandler{}
}

// HandleAlert 处理告警
func (h *ConsoleAlertHandler) HandleAlert(alert Alert) {
	fmt.Printf("[%s] [%s] %s\n", alert.Timestamp.Format("2006-01-02 15:04:05"), alert.Level, alert.Message)
	if len(alert.Labels) > 0 {
		fmt.Println("Labels:")
		for key, value := range alert.Labels {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}
}
