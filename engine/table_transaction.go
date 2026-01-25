package engine

import (
	"fmt"

	"github.com/liaoran123/sfsDb/storage"
)

// Transaction 定义事务接口
type Transaction interface {
	// Insert 在事务中插入记录
	Insert(fields *map[string]any) (int, error)
	// Update 在事务中更新记录
	Update(fields *map[string]any) error
	// Delete 在事务中删除记录
	Delete(fields *map[string]any) error
	// Commit 提交事务
	Commit() error
	// Rollback 回滚事务
	Rollback() error
}

// TableTransaction 实现Transaction接口的具体结构体
type TableTransaction struct {
	table     *Table       // 关联的表
	batch     storage.Batch // 事务使用的batch
	committed bool         // 是否已提交
}

// Begin 创建一个新的事务
func (t *Table) Begin() (Transaction, error) {
	batch := t.kvStore.GetBatch()
	if batch == nil {
		return nil, fmt.Errorf("failed to create batch for transaction")
	}
	return &TableTransaction{
		table:     t,
		batch:     batch,
		committed: false,
	}, nil
}

// Insert 在事务中插入记录
func (tx *TableTransaction) Insert(fields *map[string]any) (int, error) {
	if tx.committed {
		return 0, fmt.Errorf("transaction already committed")
	}
	return tx.table.Insert(fields, tx.batch)
}

// Update 在事务中更新记录
func (tx *TableTransaction) Update(fields *map[string]any) error {
	if tx.committed {
		return fmt.Errorf("transaction already committed")
	}
	return tx.table.Update(fields, tx.batch)
}

// Delete 在事务中删除记录
func (tx *TableTransaction) Delete(fields *map[string]any) error {
	if tx.committed {
		return fmt.Errorf("transaction already committed")
	}
	return tx.table.Delete(fields, tx.batch)
}

// Commit 提交事务
func (tx *TableTransaction) Commit() error {
	if tx.committed {
		return fmt.Errorf("transaction already committed")
	}
	err := tx.table.kvStore.WriteBatch(tx.batch)
	if err == nil {
		tx.committed = true
	}
	return err
}

// Rollback 回滚事务
// 注意：LevelDB的WriteBatch不支持真正的回滚
// 此方法仅标记事务已结束，防止重复提交
func (tx *TableTransaction) Rollback() error {
	if tx.committed {
		return fmt.Errorf("transaction already committed")
	}
	tx.committed = true
	return nil
}