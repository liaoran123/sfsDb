package audit

import (
	"fmt"
	"time"
)

// AuditEvent 审计事件
type AuditEvent struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	EventType string    `json:"event_type"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	Details   string    `json:"details"`
	Success   bool      `json:"success"`
	IPAddress string    `json:"ip_address"`
}

// AuditLogger 审计日志记录器接口
type AuditLogger interface {
	// LogEvent 记录审计事件
	LogEvent(event AuditEvent)

	// GetEvents 获取审计事件
	GetEvents(filter map[string]string, limit, offset int) ([]AuditEvent, error)

	// GetEventByID 根据ID获取审计事件
	GetEventByID(eventID string) (*AuditEvent, error)

	// SearchEvents 搜索审计事件
	SearchEvents(query string, limit, offset int) ([]AuditEvent, error)
}

// AuditConfig 审计配置
type AuditConfig struct {
	Enabled      bool   `json:"enabled"`
	LogFile      string `json:"log_file"`
	MaxEvents    int    `json:"max_events"`
	EventTypes   []string `json:"event_types"`
}

// ConsoleAuditLogger 控制台审计日志记录器
type ConsoleAuditLogger struct {
	config AuditConfig
	events []AuditEvent
}

// NewConsoleAuditLogger 创建控制台审计日志记录器
func NewConsoleAuditLogger(config AuditConfig) *ConsoleAuditLogger {
	return &ConsoleAuditLogger{
		config: config,
		events: make([]AuditEvent, 0),
	}
}

// LogEvent 记录审计事件
func (l *ConsoleAuditLogger) LogEvent(event AuditEvent) {
	// 控制台输出
	fmt.Printf("[%s] [%s] %s %s %s %v\n",
		event.Timestamp.Format("2006-01-02 15:04:05"),
		event.EventType,
		event.UserID,
		event.Action,
		event.Resource,
		event.Success,
	)

	// 存储事件
	l.events = append(l.events, event)

	// 限制事件数量
	if len(l.events) > l.config.MaxEvents {
		l.events = l.events[len(l.events)-l.config.MaxEvents:]
	}
}

// GetEvents 获取审计事件
func (l *ConsoleAuditLogger) GetEvents(filter map[string]string, limit, offset int) ([]AuditEvent, error) {
	// 简单实现，实际生产环境可能需要更复杂的过滤逻辑
	result := make([]AuditEvent, 0)

	for _, event := range l.events {
		// 应用过滤
		match := true
		for key, value := range filter {
			switch key {
			case "event_type":
				if event.EventType != value {
					match = false
				}
			case "user_id":
				if event.UserID != value {
					match = false
				}
			case "action":
				if event.Action != value {
					match = false
				}
			case "resource":
				if event.Resource != value {
					match = false
				}
			case "success":
				if fmt.Sprintf("%v", event.Success) != value {
					match = false
				}
			}
		}

		if match {
			result = append(result, event)
		}
	}

	// 应用分页
	if offset >= len(result) {
		return []AuditEvent{}, nil
	}

	end := offset + limit
	if end > len(result) {
		end = len(result)
	}

	return result[offset:end], nil
}

// GetEventByID 根据ID获取审计事件
func (l *ConsoleAuditLogger) GetEventByID(eventID string) (*AuditEvent, error) {
	for _, event := range l.events {
		if event.ID == eventID {
			return &event, nil
		}
	}

	return nil, fmt.Errorf("event not found")
}

// SearchEvents 搜索审计事件
func (l *ConsoleAuditLogger) SearchEvents(query string, limit, offset int) ([]AuditEvent, error) {
	// 简单实现，实际生产环境可能需要更复杂的搜索逻辑
	result := make([]AuditEvent, 0)

	for _, event := range l.events {
		// 简单的字符串包含搜索
		if contains(event.Details, query) || contains(event.Action, query) || contains(event.Resource, query) {
			result = append(result, event)
		}
	}

	// 应用分页
	if offset >= len(result) {
		return []AuditEvent{}, nil
	}

	end := offset + limit
	if end > len(result) {
		end = len(result)
	}

	return result[offset:end], nil
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// GenerateEventID 生成审计事件ID
func GenerateEventID() string {
	return fmt.Sprintf("audit_%d_%d", time.Now().Unix(), time.Now().UnixNano()%1000)
}

// NewAuditEvent 创建审计事件
func NewAuditEvent(eventType, userID, username, action, resource, details string, success bool, ipAddress string) AuditEvent {
	return AuditEvent{
		ID:        GenerateEventID(),
		Timestamp: time.Now(),
		EventType: eventType,
		UserID:    userID,
		Username:  username,
		Action:    action,
		Resource:  resource,
		Details:   details,
		Success:   success,
		IPAddress: ipAddress,
	}
}
