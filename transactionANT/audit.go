package transactionANT

import (
	"fmt"
	"sync"
	"time"

	"github.com/liaoran123/sfsDb/storage"
)

// 审计操作类型
const (
	AuditActionCreate     = "CREATE"
	AuditActionRead       = "READ"
	AuditActionUpdate     = "UPDATE"
	AuditActionDelete     = "DELETE"
	AuditActionCommit     = "COMMIT"
	AuditActionRollback   = "ROLLBACK"
	AuditActionLogin      = "LOGIN"
	AuditActionLogout     = "LOGOUT"
	AuditActionPermission = "PERMISSION"
)

// 审计资源类型
const (
	AuditResourceTable       = "TABLE"
	AuditResourceField       = "FIELD"
	AuditResourceSystem      = "SYSTEM"
	AuditResourceTransaction = "TRANSACTION"
	AuditResourceSession     = "SESSION"
)

// 审计状态
const (
	AuditStatusSuccess = "SUCCESS"
	AuditStatusFailed  = "FAILED"
)

// AuditLog 审计日志记录
type AuditLog struct {
	ID         string    `json:"id"`
	Action     string    `json:"action"`     // 操作类型
	Resource   string    `json:"resource"`   // 资源类型
	ResourceID string    `json:"resourceId"` // 具体资源ID
	UserID     string    `json:"userId"`
	Timestamp  time.Time `json:"timestamp"`
	Details    string    `json:"details"` // 操作详情
	Status     string    `json:"status"`  // 操作状态
	Hash       string    `json:"hash"`    // 日志哈希，用于防篡改
}

// AuditManager 审计日志管理器
type AuditManager struct {
	store      storage.Store
	encryption *EncryptionManager
	mutex      sync.RWMutex
	enabled    bool
}

// NewAuditManager 创建审计日志管理器
func NewAuditManager(store storage.Store) *AuditManager {
	return &AuditManager{
		store:   store,
		enabled: false,
	}
}

// GlobalAuditManager 全局审计日志管理器
var GlobalAuditManager *AuditManager

// InitAudit 初始化审计日志管理器
func InitAudit(store storage.Store) error {
	GlobalAuditManager = NewAuditManager(store)
	GlobalAuditManager.enabled = true
	return nil
}

// GetAuditManager 获取审计日志管理器
func GetAuditManager() *AuditManager {
	return GlobalAuditManager
}

// IsAuditEnabled 检查审计日志是否启用
func IsAuditEnabled() bool {
	if GlobalAuditManager == nil {
		return false
	}
	return GlobalAuditManager.enabled
}

// EnableAudit 启用审计日志
func EnableAudit() {
	if GlobalAuditManager != nil {
		GlobalAuditManager.enabled = true
	}
}

// DisableAudit 禁用审计日志
func DisableAudit() {
	if GlobalAuditManager != nil {
		GlobalAuditManager.enabled = false
	}
}

// Log 记录审计日志
func (am *AuditManager) Log(log *AuditLog) error {
	if !am.enabled || am.store == nil {
		return nil
	}

	am.mutex.Lock()
	defer am.mutex.Unlock()

	// 生成日志ID
	if log.ID == "" {
		log.ID = fmt.Sprintf("audit:%d:%s", time.Now().UnixNano(), log.UserID)
	}

	// 设置时间戳
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}

	// 生成日志哈希（简化实现，实际应该使用更安全的哈希算法）
	log.Hash = fmt.Sprintf("%d", time.Now().UnixNano())

	// 序列化日志
	// 这里简化实现，实际应该使用JSON或其他格式
	logKey := []byte(fmt.Sprintf("audit:%s", log.ID))
	logValue := []byte(fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%s",
		log.Action, log.Resource, log.ResourceID, log.UserID,
		log.Timestamp.Format(time.RFC3339), log.Details, log.Status, log.Hash))

	// 加密日志（如果启用了加密）
	if am.encryption != nil {
		encryptedValue, err := am.encryption.Encrypt(logValue)
		if err == nil {
			logValue = encryptedValue
		}
	}

	// 存储日志
	return am.store.Put(logKey, logValue)
}

// LogWithFields 便捷函数：记录审计日志
func LogAudit(action, resource, resourceID, userID, details, status string) error {
	if !IsAuditEnabled() {
		return nil
	}

	auditLog := &AuditLog{
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		UserID:     userID,
		Details:    details,
		Status:     status,
	}

	return GetAuditManager().Log(auditLog)
}

// QueryLogs 查询审计日志
func (am *AuditManager) QueryLogs(startTime, endTime time.Time, userID, action, resource string) ([]*AuditLog, error) {
	if !am.enabled || am.store == nil {
		return []*AuditLog{}, nil
	}

	am.mutex.RLock()
	defer am.mutex.RUnlock()

	// 简化实现，实际应该使用更复杂的查询逻辑
	// 这里只是返回空结果
	return []*AuditLog{}, nil
}

// GetAuditLog 获取单个审计日志
func (am *AuditManager) GetAuditLog(logID string) (*AuditLog, error) {
	if !am.enabled || am.store == nil {
		return nil, fmt.Errorf("audit not enabled")
	}

	am.mutex.RLock()
	defer am.mutex.RUnlock()

	// 简化实现，实际应该从存储中读取
	return nil, fmt.Errorf("not implemented")
}

// SetEncryptionManager 设置加密管理器
func (am *AuditManager) SetEncryptionManager(encryption *EncryptionManager) {
	am.encryption = encryption
}
