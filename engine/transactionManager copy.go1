package engine

import (
	"fmt"

	"github.com/liaoran123/sfsDb/storage"
)

// NewTransactionManager 创建一个新的事务管理器
// batch: 共享的batch对象，用于所有事务
func NewTransactionManager(batch storage.Batch) *TransactionManager {
	return &TransactionManager{
		transactions: make([]Transaction, 0),
		batch:        batch,
		committed:    false,
	}
}

// AddTable 添加一个表到事务管理器，并返回对应的事务
// table: 要添加的表
func (tm *TransactionManager) AddTable(table *Table) (Transaction, error) {
	if tm.committed {
		return nil, fmt.Errorf("transaction manager already committed")
	}

	// 为表创建使用共享batch的事务
	tx, err := table.BeginWithBatch(tm.batch)
	if err != nil {
		return nil, err
	}

	// 将事务添加到管理列表
	tm.transactions = append(tm.transactions, tx)
	return tx, nil
}

// Commit 提交所有事务
// 注意：只需要提交第一个事务，因为所有事务共享同一个batch
func (tm *TransactionManager) Commit() error {
	if tm.committed {
		return fmt.Errorf("transaction already committed")
	}

	if len(tm.transactions) == 0 {
		// 没有事务需要提交
		tm.committed = true
		return nil
	}

	// 提交第一个事务（所有事务共享同一个batch，只需要提交一次）
	err := tm.transactions[0].Commit()
	if err != nil {
		return err
	}

	// 释放其他事务的资源
	for i := 1; i < len(tm.transactions); i++ {
		tm.transactions[i].Rollback()
	}

	tm.committed = true
	return nil
}

// Rollback 回滚所有事务
func (tm *TransactionManager) Rollback() error {
	if tm.committed {
		return fmt.Errorf("transaction already committed")
	}

	// 回滚所有事务
	for _, tx := range tm.transactions {
		tx.Rollback()
	}

	tm.committed = true
	return nil
}

// GetBatch 获取事务管理器使用的batch
func (tm *TransactionManager) GetBatch() storage.Batch {
	return tm.batch
}

// GetTransactions 获取所有事务
func (tm *TransactionManager) GetTransactions() []Transaction {
	return tm.transactions
}

// WithTransaction 执行多表事务的便捷函数
// batch: 共享的batch对象
// tables: 要参与事务的表列表
// fn: 事务回调函数，接收每个表对应的事务
func WithTransaction(batch storage.Batch, tables []*Table, fn func(transactions map[*Table]Transaction) error) error {
	// 创建事务管理器
	tm := NewTransactionManager(batch)
	transactions := make(map[*Table]Transaction)

	// 为每个表创建事务
	for _, table := range tables {
		tx, err := tm.AddTable(table)
		if err != nil {
			// 出错时回滚所有事务
			tm.Rollback()
			return err
		}
		transactions[table] = tx
	}

	// 执行回调函数
	err := fn(transactions)
	if err != nil {
		// 回调函数出错，回滚所有事务
		tm.Rollback()
		return err
	}

	// 回调函数成功，提交所有事务
	return tm.Commit()
}
