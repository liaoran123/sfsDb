package transactionLockANT

import "time"

// 事务隔离级别常量
const (
	// ReadUncommitted 读未提交：允许读取未提交的数据，可能导致脏读、不可重复读、幻读
	ReadUncommitted = "READ_UNCOMMITTED"
	// ReadCommitted 读已提交：只能读取已提交的数据，避免脏读，但可能导致不可重复读、幻读
	ReadCommitted = "READ_COMMITTED"
	// RepeatableRead 可重复读：确保同一事务中多次读取同一数据时结果一致，避免脏读、不可重复读，但可能导致幻读
	RepeatableRead = "REPEATABLE_READ"
	// Serializable 可序列化：最高隔离级别，完全避免脏读、不可重复读、幻读
	Serializable = "SERIALIZABLE"
)

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
}

// DefaultTransactionOptions 默认事务选项
func DefaultTransactionOptions() *TransactionOptions {
	return &TransactionOptions{
		IsolationLevel:     RepeatableRead,
		AllowNested:        false,
		Timeout:            0,
		MaxRetries:         3,
		InitialRetryDelay:  10 * time.Millisecond,
		RetryBackoffFactor: 2.0,
	}
}
