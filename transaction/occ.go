package transaction

import (
	"fmt"
	"time"

	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/transaction/lock"
)

// OCCSupport 乐观并发控制支持
type OCCSupport struct {
	versionMap map[string]uint64 // 记录每个键的版本号
}

// NewOCCSupport 创建新的乐观并发控制支持
func NewOCCSupport() *OCCSupport {
	return &OCCSupport{
		versionMap: make(map[string]uint64),
	}
}

// GetVersion 获取键的版本号
func (occ *OCCSupport) GetVersion(key string) uint64 {
	if version, exists := occ.versionMap[key]; exists {
		return version
	}
	return 0
}

// IncrementVersion 递增键的版本号
func (occ *OCCSupport) IncrementVersion(key string) uint64 {
	version := occ.GetVersion(key)
	version++
	occ.versionMap[key] = version
	return version
}

// CheckConflict 检查冲突
func (occ *OCCSupport) CheckConflict(key string, expectedVersion uint64) bool {
	currentVersion := occ.GetVersion(key)
	return currentVersion != expectedVersion
}

// TableTransactionWithOCC 带乐观并发控制的表事务
type TableTransactionWithOCC struct {
	*TableTransaction
	occ *OCCSupport
}

// NewTableTransactionWithOCC 创建带乐观并发控制的表事务
func NewTableTransactionWithOCC(table *engine.Table) (*TableTransactionWithOCC, error) {
	tx, err := NewTableTransaction(table)
	if err != nil {
		return nil, err
	}

	return &TableTransactionWithOCC{
		TableTransaction: tx,
		occ:              NewOCCSupport(),
	}, nil
}

// InsertWithOCC 使用乐观并发控制插入记录
func (tx *TableTransactionWithOCC) InsertWithOCC(fields *map[string]interface{}) (int, error) {
	// 预处理：计算缓存键
	cacheKey := tx.GetCacheKey(fields)

	// 获取当前版本
	currentVersion := tx.occ.GetVersion(cacheKey)

	// 检查冲突
	if tx.occ.CheckConflict(cacheKey, currentVersion) {
		return 0, fmt.Errorf("conflict detected, please retry")
	}

	// 执行插入操作
	id, err := tx.Insert(fields)
	if err != nil {
		return 0, err
	}

	// 递增版本号
	tx.occ.IncrementVersion(cacheKey)

	return id, nil
}

// UpdateWithOCC 使用乐观并发控制更新记录
func (tx *TableTransactionWithOCC) UpdateWithOCC(fields *map[string]interface{}) error {
	// 预处理：计算缓存键
	cacheKey := tx.GetCacheKey(fields)

	// 获取当前版本
	currentVersion := tx.occ.GetVersion(cacheKey)

	// 检查冲突
	if tx.occ.CheckConflict(cacheKey, currentVersion) {
		return fmt.Errorf("conflict detected, please retry")
	}

	// 执行更新操作
	err := tx.Update(fields)
	if err != nil {
		return err
	}

	// 递增版本号
	tx.occ.IncrementVersion(cacheKey)

	return nil
}

// DeleteWithOCC 使用乐观并发控制删除记录
func (tx *TableTransactionWithOCC) DeleteWithOCC(fields *map[string]interface{}) error {
	// 预处理：计算缓存键
	cacheKey := tx.GetCacheKey(fields)

	// 获取当前版本
	currentVersion := tx.occ.GetVersion(cacheKey)

	// 检查冲突
	if tx.occ.CheckConflict(cacheKey, currentVersion) {
		return fmt.Errorf("conflict detected, please retry")
	}

	// 执行删除操作
	err := tx.Delete(fields)
	if err != nil {
		return err
	}

	// 递增版本号
	tx.occ.IncrementVersion(cacheKey)

	return nil
}

// BatchOperationWithOCC 批量操作的乐观并发控制
type BatchOperationWithOCC struct {
	tx         *TableTransaction
	operations []BatchOperation
	occ        *OCCSupport
}

// BatchOperation 批量操作
type BatchOperation struct {
	Type     string                  // 操作类型：insert, update, delete
	Fields   *map[string]interface{} // 操作字段
	CacheKey string                  // 缓存键
	Version  uint64                  // 预期版本号
}

// NewBatchOperationWithOCC 创建新的批量操作
func NewBatchOperationWithOCC(tx *TableTransaction) *BatchOperationWithOCC {
	return &BatchOperationWithOCC{
		tx:         tx,
		operations: make([]BatchOperation, 0),
		occ:        NewOCCSupport(),
	}
}

// AddInsert 添加插入操作
func (batch *BatchOperationWithOCC) AddInsert(fields *map[string]interface{}) {
	cacheKey := batch.tx.GetCacheKey(fields)
	version := batch.occ.GetVersion(cacheKey)

	batch.operations = append(batch.operations, BatchOperation{
		Type:     "insert",
		Fields:   fields,
		CacheKey: cacheKey,
		Version:  version,
	})
}

// AddUpdate 添加更新操作
func (batch *BatchOperationWithOCC) AddUpdate(fields *map[string]interface{}) {
	cacheKey := batch.tx.GetCacheKey(fields)
	version := batch.occ.GetVersion(cacheKey)

	batch.operations = append(batch.operations, BatchOperation{
		Type:     "update",
		Fields:   fields,
		CacheKey: cacheKey,
		Version:  version,
	})
}

// AddDelete 添加删除操作
func (batch *BatchOperationWithOCC) AddDelete(fields *map[string]interface{}) {
	cacheKey := batch.tx.GetCacheKey(fields)
	version := batch.occ.GetVersion(cacheKey)

	batch.operations = append(batch.operations, BatchOperation{
		Type:     "delete",
		Fields:   fields,
		CacheKey: cacheKey,
		Version:  version,
	})
}

// Execute 执行批量操作
func (batch *BatchOperationWithOCC) Execute() error {
	// 1. 检查冲突
	for _, op := range batch.operations {
		if batch.occ.CheckConflict(op.CacheKey, op.Version) {
			return fmt.Errorf("conflict detected for key %s, please retry", op.CacheKey)
		}
	}

	// 2. 获取所有锁
	for _, op := range batch.operations {
		err := GlobalLockManager.AcquireLock(lock.LockRequest{
			Key:       op.CacheKey,
			Type:      lock.WriteLock,
			Mode:      lock.Blocking,
			Timeout:   5 * time.Second,
			Requestor: fmt.Sprintf("tx_%d", batch.tx.TxID),
		})
		if err != nil {
			// 释放已获取的锁
			for _, prevOp := range batch.operations {
				if prevOp.CacheKey == op.CacheKey {
					break
				}
				GlobalLockManager.ReleaseLock(prevOp.CacheKey, fmt.Sprintf("tx_%d", batch.tx.TxID))
			}
			return fmt.Errorf("failed to acquire lock: %v", err)
		}
	}

	// 3. 执行所有操作
	for _, op := range batch.operations {
		var err error
		switch op.Type {
		case "insert":
			_, err = batch.tx.Insert(op.Fields)
		case "update":
			err = batch.tx.Update(op.Fields)
		case "delete":
			err = batch.tx.Delete(op.Fields)
		}
		if err != nil {
			// 释放所有锁
			for _, op := range batch.operations {
				GlobalLockManager.ReleaseLock(op.CacheKey, fmt.Sprintf("tx_%d", batch.tx.TxID))
			}
			return fmt.Errorf("failed to execute operation: %v", err)
		}

		// 递增版本号
		batch.occ.IncrementVersion(op.CacheKey)
	}

	// 4. 释放所有锁
	for _, op := range batch.operations {
		GlobalLockManager.ReleaseLock(op.CacheKey, fmt.Sprintf("tx_%d", batch.tx.TxID))
	}

	return nil
}
