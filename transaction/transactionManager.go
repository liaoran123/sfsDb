package transaction

import (
	"fmt"

	"github.com/liaoran123/sfsDb/storage"
)

// TransactionManager 管理多表事务
// 用于简化多表事务的创建、提交和回滚操作
// 所有表共享同一个batch，确保原子性
type TransactionManager struct {
	transactions []Transaction // 管理的事务列表
	batch        storage.Batch // 共享的batch
	committed    bool          // 是否已提交
}

// NewTransactionManager 创建一个新的事务管理器
// batch: 共享的batch对象，用于所有事务
func NewTransactionManager(batch storage.Batch) *TransactionManager {
	return &TransactionManager{
		transactions: make([]Transaction, 0),
		batch:        batch,
		committed:    false,
	}
}

// AddTableAccessor 添加一个表访问器到事务管理器，并返回对应的事务
// tableAccessor: 要添加的表访问器
func (tm *TransactionManager) AddTableAccessor(tableAccessor TableAccessor) (Transaction, error) {
	if tm.committed {
		return nil, fmt.Errorf("transaction manager already committed")
	}
	tx, err := NewTableTransaction(tableAccessor, tm.batch, DefaultTransactionOptions())
	if err != nil {
		return nil, err
	}
	// 将事务添加到管理列表
	tm.transactions = append(tm.transactions, tx)
	return tx, nil
}

// AddTableAccessorWithOptions 添加一个表访问器到事务管理器，并使用指定的事务选项
// tableAccessor: 要添加的表访问器
// options: 事务选项
func (tm *TransactionManager) AddTableAccessorWithOptions(tableAccessor TableAccessor, options *TransactionOptions) (Transaction, error) {
	if tm.committed {
		return nil, fmt.Errorf("transaction manager already committed")
	}
	tx, err := NewTableTransaction(tableAccessor, tm.batch, options)
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
// tableAccessors: 要参与事务的表访问器列表
// fn: 事务回调函数，接收每个表访问器对应的事务
func WithTransaction(batch storage.Batch, tableAccessors []TableAccessor, fn func(transactions map[TableAccessor]Transaction) error) error {
	// 创建事务管理器
	tm := NewTransactionManager(batch)
	transactions := make(map[TableAccessor]Transaction)

	// 为每个表访问器创建事务
	for _, tableAccessor := range tableAccessors {
		tx, err := tm.AddTableAccessor(tableAccessor)
		if err != nil {
			// 出错时回滚所有事务
			tm.Rollback()
			return err
		}
		transactions[tableAccessor] = tx
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

// WithTransactionWithOptions 执行多表事务的便捷函数，使用指定的事务选项
// batch: 共享的batch对象
// tableAccessors: 要参与事务的表访问器列表
// options: 事务选项
// fn: 事务回调函数，接收每个表访问器对应的事务
func WithTransactionWithOptions(batch storage.Batch, tableAccessors []TableAccessor, options *TransactionOptions, fn func(transactions map[TableAccessor]Transaction) error) error {
	// 创建事务管理器
	tm := NewTransactionManager(batch)
	transactions := make(map[TableAccessor]Transaction)

	// 为每个表访问器创建事务
	for _, tableAccessor := range tableAccessors {
		tx, err := tm.AddTableAccessorWithOptions(tableAccessor, options)
		if err != nil {
			// 出错时回滚所有事务
			tm.Rollback()
			return err
		}
		transactions[tableAccessor] = tx
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
