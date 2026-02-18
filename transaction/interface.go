package transaction

import (
	"fmt"
	"time"

	"github.com/liaoran123/sfsDb/storage"
	"github.com/liaoran123/sfsDb/util"
)

// 事务隔离级别常量
const (
	// ReadUncommitted 读未提交：允许读取未提交的数据，可能导致脏读、不可重复读、幻读
	ReadUncommitted = "READ_UNCOMMITTED"
	// ReadCommitted 读已提交：只能读取已提交的数据，避免脏读，但可能导致不可重复读、幻读
	ReadCommitted = "READ_COMMITTED"
	// RepeatableRead 可重复读：确保同一事务中多次读取同一数据时结果一致，避免脏读、不可重复读，但可能导致幻读
	RepeatableRead = "REPEATABLE_READ"
	// Serializable 可序列化：最高隔离级别，完全避免脏读、不可重复读、幻读，但性能最低
	Serializable = "SERIALIZABLE"
)

// 错误类型常量
const (
	// ErrorTypePermission 权限错误
	ErrorTypePermission = "PERMISSION_ERROR"
	// ErrorTypeLock 锁错误
	ErrorTypeLock = "LOCK_ERROR"
	// ErrorTypeTransaction 事务错误
	ErrorTypeTransaction = "TRANSACTION_ERROR"
	// ErrorTypeStorage 存储错误
	ErrorTypeStorage = "STORAGE_ERROR"
	// ErrorTypeEncryption 加密错误
	ErrorTypeEncryption = "ENCRYPTION_ERROR"
	// ErrorTypeTimeout 超时错误
	ErrorTypeTimeout = "TIMEOUT_ERROR"
	// ErrorTypeValidation 验证错误
	ErrorTypeValidation = "VALIDATION_ERROR"
	// ErrorTypeRecovery 恢复错误
	ErrorTypeRecovery = "RECOVERY_ERROR"
)

// TransactionError 事务错误结构体
type TransactionError struct {
	Type        string `json:"type"`
	Message     string `json:"message"`
	Operation   string `json:"operation"`
	Table       string `json:"table"`
	TxID        uint64 `json:"txID"`
	IsRetryable bool   `json:"isRetryable"`
	Cause       error  `json:"cause,omitempty"`
}

// Error 实现error接口
func (e *TransactionError) Error() string {
	return fmt.Sprintf("%s: %s (operation: %s, table: %s, txID: %d)", e.Type, e.Message, e.Operation, e.Table, e.TxID)
}

// Unwrap 实现errors.Unwrap接口
func (e *TransactionError) Unwrap() error {
	return e.Cause
}

// TransactionOptions 事务选项结构体
type TransactionOptions struct {
	// 隔离级别
	IsolationLevel string `json:"isolationLevel"`
	// 是否启用嵌套事务
	AllowNested bool `json:"allowNested"`
	// 事务超时时间
	Timeout time.Duration `json:"timeout"`
	// 最大重试次数
	MaxRetries int `json:"maxRetries"`
	// 初始重试延迟
	InitialRetryDelay time.Duration `json:"initialRetryDelay"`
	// 重试退避因子（sleep时间乘法因子）
	RetryBackoffFactor float64 `json:"retryBackoffFactor"`
	// 事务权限
	Permissions []string `json:"permissions"`
	// 是否启用数据加密
	EnableEncryption bool `json:"enableEncryption"`
	// 加密密钥ID
	EncryptionKeyID string `json:"encryptionKeyID"`
	// 是否启用事务日志
	EnableTransactionLog bool `json:"enableTransactionLog"`
	// 事务日志路径
	TransactionLogPath string `json:"transactionLogPath"`
	// 事务日志备份间隔
	TransactionLogBackupInterval time.Duration `json:"transactionLogBackupInterval"`
}

// DefaultTransactionOptions 默认事务选项
func DefaultTransactionOptions() *TransactionOptions {
	return &TransactionOptions{
		IsolationLevel:               RepeatableRead,
		AllowNested:                  false,
		Timeout:                      0,
		MaxRetries:                   3,
		InitialRetryDelay:            10 * time.Millisecond,
		RetryBackoffFactor:           2.0,
		Permissions:                  []string{},
		EnableEncryption:             false,
		EncryptionKeyID:              "",
		EnableTransactionLog:         true,
		TransactionLogPath:           "./transaction_logs",
		TransactionLogBackupInterval: 1 * time.Hour,
	}
}

// Transaction 定义事务接口
type Transaction interface {
	// Insert 在事务中插入记录
	Insert(fields *map[string]any) (int, error)
	// Update 在事务中更新记录
	Update(fields *map[string]any) error
	// Delete 在事务中删除记录
	Delete(fields *map[string]any) error
	// Search 在事务中搜索记录（支持读一致性）
	Search(fields *map[string]any, ops ...util.ComparisonOperator) (interface{}, error)
	// SearchRange 在事务中进行区间搜索（支持读一致性）
	SearchRange(funIter storage.FunIter, fieldname string, Start, Limit any) (interface{}, error)
	// Read 在事务中读取单条记录（支持读一致性）
	Read(fields *map[string]any) ([]byte, error)
	// Commit 提交事务
	Commit() error
	// Rollback 回滚事务
	Rollback() error
	// BeginNested 创建一个嵌套事务
	BeginNested() (Transaction, error)
	// GetOptions 获取事务选项
	GetOptions() *TransactionOptions
	// GetTxID 获取事务ID
	GetTxID() uint64
	// CheckPermission 检查事务是否有指定权限
	CheckPermission(permission string) bool
	// EncryptData 加密数据
	EncryptData(data []byte) ([]byte, error)
	// DecryptData 解密数据
	DecryptData(data []byte) ([]byte, error)
	// BackupTransactionLog 备份事务日志
	BackupTransactionLog() error
	// RecoverFromTransactionLog 从事务日志恢复
	RecoverFromTransactionLog(logPath string) error
	// GetTransactionLogPath 获取事务日志路径
	GetTransactionLogPath() string
}

// TableAccessor 表访问接口，用于事务操作表
type TableAccessor interface {
	// GetName 获取表名
	GetName() string
	// GetID 获取表ID
	GetID() int
	// FieldsToBytes 将字段转换为字节数组
	FieldsToBytes(fields *map[string]any) *map[string][]byte
	// FormatRecord 格式化记录
	FormatRecord(fieldsBytes *map[string][]byte) []byte
	// Insert 插入记录
	Insert(fields *map[string]any, batch storage.Batch) (int, error)
	// Update 更新记录
	Update(fields *map[string]any, batch storage.Batch) error
	// Delete 删除记录
	Delete(fields *map[string]any, batch storage.Batch) error
	// Searchs 搜索记录
	Searchs(funIter storage.FunIter, fields *map[string]any, ops ...util.ComparisonOperator) (interface{}, error)
	// SearchRange 区间搜索
	SearchRange(funIter storage.FunIter, fieldname string, Start, Limit any) (interface{}, error)
	// GetPrimaryKey 获取主键
	GetPrimaryKey() interface{}
	// generateLockKey 生成锁键
	GenerateLockKey(fields *map[string]any) string
	// GetKVStore 获取KV存储
	GetKVStore() storage.Store
}
