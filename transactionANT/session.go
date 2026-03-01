package transactionANT

import (
	"fmt"
	"sync"
	"time"
)

// Session 会话结构体
type Session struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"userId"`
	CreatedAt time.Time              `json:"createdAt"`
	ExpiresAt time.Time              `json:"expiresAt"`
	Data      map[string]interface{} `json:"data"`
}

// SessionManager 会话管理器
type SessionManager struct {
	sessions map[string]*Session
	mutex    sync.RWMutex
	timeout  time.Duration
}

// NewSessionManager 创建会话管理器
func NewSessionManager(timeout time.Duration) *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
		timeout:  timeout,
	}
}

// GlobalSessionManager 全局会话管理器
var GlobalSessionManager *SessionManager

// InitSessionManager 初始化会话管理器
func InitSessionManager(timeout time.Duration) error {
	GlobalSessionManager = NewSessionManager(timeout)
	// 启动会话清理器
	go GlobalSessionManager.startCleanup()
	return nil
}

// GetSessionManager 获取会话管理器
func GetSessionManager() *SessionManager {
	return GlobalSessionManager
}

// startCleanup 启动会话清理器
func (sm *SessionManager) startCleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		<-ticker.C
		sm.cleanupExpiredSessions()
	}
}

// cleanupExpiredSessions 清理过期会话
func (sm *SessionManager) cleanupExpiredSessions() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	now := time.Now()
	for id, session := range sm.sessions {
		if now.After(session.ExpiresAt) {
			delete(sm.sessions, id)
		}
	}
}

// CreateSession 创建会话
func (sm *SessionManager) CreateSession(userID string) (*Session, error) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 生成会话ID
	sessionID := fmt.Sprintf("session:%d:%s", time.Now().UnixNano(), userID)

	// 计算过期时间
	expiresAt := time.Now().Add(sm.timeout)

	// 创建会话
	session := &Session{
		ID:        sessionID,
		UserID:    userID,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
		Data:      make(map[string]interface{}),
	}

	// 存储会话
	sm.sessions[sessionID] = session

	return session, nil
}

// GetSession 获取会话
func (sm *SessionManager) GetSession(sessionID string) (*Session, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	// 检查会话是否过期
	if time.Now().After(session.ExpiresAt) {
		return nil, fmt.Errorf("session expired")
	}

	return session, nil
}

// RefreshSession 刷新会话
func (sm *SessionManager) RefreshSession(sessionID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	// 更新过期时间
	session.ExpiresAt = time.Now().Add(sm.timeout)

	return nil
}

// DeleteSession 删除会话
func (sm *SessionManager) DeleteSession(sessionID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	_, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	delete(sm.sessions, sessionID)
	return nil
}

// ValidateSession 验证会话
func (sm *SessionManager) ValidateSession(sessionID string) (bool, error) {
	session, err := sm.GetSession(sessionID)
	if err != nil {
		return false, err
	}

	// 检查用户是否存在
	acm := GetAccessControlManager()
	_, err = acm.GetUser(session.UserID)
	if err != nil {
		return false, fmt.Errorf("user not found")
	}

	return true, nil
}

// GetUserIDFromSession 从会话获取用户ID
func (sm *SessionManager) GetUserIDFromSession(sessionID string) (string, error) {
	session, err := sm.GetSession(sessionID)
	if err != nil {
		return "", err
	}

	return session.UserID, nil
}
