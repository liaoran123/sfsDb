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
	// 事务总数
	TotalCount atomic.Int64 `json:"totalCount"`
	// 成功事务数
	SuccessCount atomic.Int64 `json:"successCount"`
	// 失败事务数
	FailCount atomic.Int64 `json:"failCount"`
	// 成功事务耗时总和
	SuccessDuration time.Duration `json:"successDuration"`
	// 失败事务耗时总和
	FailDuration time.Duration `json:"failDuration"`
	// 成功事务耗时平均值
	SuccessAvgDuration time.Duration `json:"successAvgDuration"`
	// 失败事务耗时平均值
	FailAvgDuration time.Duration `json:"failAvgDuration"`
	// 最大事务耗时
	MaxDuration time.Duration `json:"maxDuration"`
	// 最小事务耗时
	MinDuration time.Duration `json:"minDuration"`
}
