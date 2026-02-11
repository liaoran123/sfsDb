package monitor

import (
	"sync/atomic"
	"time"
)

type TransactionStats struct {
	// 事务ID
	TxID uint64 `json:"txID"`
	// 表名
	TableName string `json:"tableName"`
	// 隔离级别
	IsolationLevel string `json:"isolationLevel"`
	// 是否提交
	IsCommitted bool `json:"isCommitted"` //不是提交则是回滚
	// 事务总数
	TotalCount atomic.Int64 `json:"totalCount"`
	// 成功事务数
	SuccessCount atomic.Int64 `json:"successCount"` // 成功事务数，提交事务数
	// 失败事务数
	FailCount atomic.Int64 `json:"failCount"` // 失败事务数，回滚事务数
	// 成功事务耗时总和
	SuccessDuration time.Duration `json:"successDuration"` // 成功事务耗时总和，提交
	// 失败事务耗时总和
	FailDuration time.Duration `json:"failDuration"` // 失败事务耗时总和
	// 成功事务耗时平均值
	SuccessAvgDuration time.Duration `json:"successAvgDuration"` // 成功事务耗时平均值，提交事务耗时平均值
	// 失败事务耗时平均值
	FailAvgDuration time.Duration `json:"failAvgDuration"` // 失败事务耗时平均值
	// 最大事务耗时
	MaxDuration time.Duration `json:"maxDuration"` // 最大事务耗时，提交事务耗时
	// 最小事务耗时
	MinDuration time.Duration `json:"minDuration"` // 最小事务耗时，提交事务耗时
}
