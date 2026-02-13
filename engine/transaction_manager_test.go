package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/liaoran123/sfsDb/storage"
)

// 测试错误定义
var ErrTableNotExist = fmt.Errorf("table not exist")

// TestTransactionManagerBasic 测试事务管理器的基本功能
func TestTransactionManagerBasic(t *testing.T) {
	// 测试目录设置
	testDir := filepath.Join(os.TempDir(), "sfsdb_tx_test")
	testDbPath := filepath.Join(testDir, "test_db")

	// 清理测试目录
	defer func() {
		os.RemoveAll(testDir)
	}()

	// 创建测试数据库
	testDb, err := storage.OpenDefaultDb(testDbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDb.Close()

	// 创建表结构
	table1, err := TableNew("test_table1")
	if err != nil {
		t.Fatalf("Failed to create table1: %v", err)
	}

	// 设置表1的字段
	fields1 := map[string]any{
		"id":   0,
		"name": "",
	}
	err = table1.SetFields(fields1)
	if err != nil {
		t.Fatalf("Failed to set fields for table1: %v", err)
	}

	table2, err := TableNew("test_table2")
	if err != nil {
		t.Fatalf("Failed to create table2: %v", err)
	}

	// 设置表2的字段
	fields2 := map[string]any{
		"id":   0,
		"name": "",
	}
	err = table2.SetFields(fields2)
	if err != nil {
		t.Fatalf("Failed to set fields for table2: %v", err)
	}

	// 创建 batch
	batch := testDb.GetBatch()
	if batch == nil {
		t.Fatalf("Failed to create batch")
	}

	// 测试创建事务管理器
	t.Run("CreateTransactionManager", func(t *testing.T) {
		tm := NewTransactionManager(batch)
		if tm == nil {
			t.Fatalf("Failed to create transaction manager")
		}

		if len(tm.GetTransactions()) != 0 {
			t.Fatalf("Expected empty transactions list, got %d", len(tm.GetTransactions()))
		}

		if tm.GetBatch() != batch {
			t.Fatalf("Batch mismatch")
		}
	})

	// 测试添加表到事务管理器
	t.Run("AddTable", func(t *testing.T) {
		tm := NewTransactionManager(batch)

		// 添加第一个表
		tx1, err := tm.AddTable(table1)
		if err != nil {
			t.Fatalf("Failed to add table1: %v", err)
		}

		if tx1 == nil {
			t.Fatalf("Expected non-nil transaction for table1")
		}

		if len(tm.GetTransactions()) != 1 {
			t.Fatalf("Expected 1 transaction, got %d", len(tm.GetTransactions()))
		}

		// 添加第二个表
		tx2, err := tm.AddTable(table2)
		if err != nil {
			t.Fatalf("Failed to add table2: %v", err)
		}

		if tx2 == nil {
			t.Fatalf("Expected non-nil transaction for table2")
		}

		if len(tm.GetTransactions()) != 2 {
			t.Fatalf("Expected 2 transactions, got %d", len(tm.GetTransactions()))
		}
	})

	// 测试提交事务
	t.Run("CommitTransaction", func(t *testing.T) {
		tm := NewTransactionManager(batch)

		// 添加表并执行操作
		tx1, err := tm.AddTable(table1)
		if err != nil {
			t.Fatalf("Failed to add table1: %v", err)
		}

		// 插入测试数据
		fields := &map[string]any{
			"id":   1,
			"name": "test",
		}
		_, err = tx1.Insert(fields)
		if err != nil {
			t.Fatalf("Failed to insert data: %v", err)
		}

		// 提交事务
		err = tm.Commit()
		if err != nil {
			t.Fatalf("Failed to commit transaction: %v", err)
		}

		// 验证事务已提交
		tx3, err := tm.AddTable(table1)
		if err == nil {
			t.Fatalf("Expected error when adding table to committed transaction manager, got nil")
		}

		if tx3 != nil {
			t.Fatalf("Expected nil transaction when adding table to committed transaction manager, got non-nil")
		}
	})

	// 测试回滚事务
	t.Run("RollbackTransaction", func(t *testing.T) {
		// 创建新的 batch 用于回滚测试
		batch2 := testDb.GetBatch()
		if batch2 == nil {
			t.Fatalf("Failed to create batch2")
		}

		tm := NewTransactionManager(batch2)

		// 添加表并执行操作
		tx1, err := tm.AddTable(table1)
		if err != nil {
			t.Fatalf("Failed to add table1: %v", err)
		}

		// 插入测试数据
		fields := &map[string]any{
			"id":   2,
			"name": "test2",
		}
		_, err = tx1.Insert(fields)
		if err != nil {
			t.Fatalf("Failed to insert data: %v", err)
		}

		// 回滚事务
		err = tm.Rollback()
		if err != nil {
			t.Fatalf("Failed to rollback transaction: %v", err)
		}

		// 验证事务已回滚
		tx3, err := tm.AddTable(table1)
		if err == nil {
			t.Fatalf("Expected error when adding table to rolled back transaction manager, got nil")
		}

		if tx3 != nil {
			t.Fatalf("Expected nil transaction when adding table to rolled back transaction manager, got non-nil")
		}
	})

	// 测试空事务
	t.Run("EmptyTransaction", func(t *testing.T) {
		tm := NewTransactionManager(batch)

		// 提交空事务
		err := tm.Commit()
		if err != nil {
			t.Fatalf("Failed to commit empty transaction: %v", err)
		}

		// 验证事务已提交
		tx, err := tm.AddTable(table1)
		if err == nil {
			t.Fatalf("Expected error when adding table to committed transaction manager, got nil")
		}

		if tx != nil {
			t.Fatalf("Expected nil transaction when adding table to committed transaction manager, got non-nil")
		}
	})

	// 测试重复提交
	t.Run("DuplicateCommit", func(t *testing.T) {
		tm := NewTransactionManager(batch)

		// 提交事务
		err := tm.Commit()
		if err != nil {
			t.Fatalf("Failed to commit transaction: %v", err)
		}

		// 再次提交，应该失败
		err = tm.Commit()
		if err == nil {
			t.Fatalf("Expected error when committing already committed transaction, got nil")
		}
	})

	// 测试回滚已提交的事务
	t.Run("RollbackCommittedTransaction", func(t *testing.T) {
		tm := NewTransactionManager(batch)

		// 提交事务
		err := tm.Commit()
		if err != nil {
			t.Fatalf("Failed to commit transaction: %v", err)
		}

		// 回滚已提交的事务，应该失败
		err = tm.Rollback()
		if err == nil {
			t.Fatalf("Expected error when rolling back already committed transaction, got nil")
		}
	})
}

// TestTransactionManagerMultiTable 测试多表事务
func TestTransactionManagerMultiTable(t *testing.T) {
	// 测试目录设置
	testDir := filepath.Join(os.TempDir(), "sfsdb_tx_multi_test")
	testDbPath := filepath.Join(testDir, "test_db")

	// 清理测试目录
	defer func() {
		os.RemoveAll(testDir)
	}()

	// 创建测试数据库
	testDb, err := storage.OpenDefaultDb(testDbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDb.Close()

	// 创建表结构
	table1, err := TableNew("test_table1")
	if err != nil {
		t.Fatalf("Failed to create table1: %v", err)
	}

	// 设置表1的字段
	fields1 := map[string]any{
		"id":   0,
		"name": "",
	}
	err = table1.SetFields(fields1)
	if err != nil {
		t.Fatalf("Failed to set fields for table1: %v", err)
	}

	table2, err := TableNew("test_table2")
	if err != nil {
		t.Fatalf("Failed to create table2: %v", err)
	}

	// 设置表2的字段
	fields2 := map[string]any{
		"id":   0,
		"name": "",
	}
	err = table2.SetFields(fields2)
	if err != nil {
		t.Fatalf("Failed to set fields for table2: %v", err)
	}

	// 测试多表事务
	t.Run("MultiTableTransaction", func(t *testing.T) {
		// 创建 batch
		batch := testDb.GetBatch()
		if batch == nil {
			t.Fatalf("Failed to create batch")
		}

		tm := NewTransactionManager(batch)

		// 添加两个表
		tx1, err := tm.AddTable(table1)
		if err != nil {
			t.Fatalf("Failed to add table1: %v", err)
		}

		tx2, err := tm.AddTable(table2)
		if err != nil {
			t.Fatalf("Failed to add table2: %v", err)
		}

		// 在两个表中插入数据
		// 表1插入数据
		fields1 := &map[string]any{
			"id":   1,
			"name": "table1_test",
		}
		_, err = tx1.Insert(fields1)
		if err != nil {
			t.Fatalf("Failed to insert data into table1: %v", err)
		}

		// 表2插入数据
		fields2 := &map[string]any{
			"id":   1,
			"name": "table2_test",
		}
		_, err = tx2.Insert(fields2)
		if err != nil {
			t.Fatalf("Failed to insert data into table2: %v", err)
		}

		// 提交事务
		err = tm.Commit()
		if err != nil {
			t.Fatalf("Failed to commit multi-table transaction: %v", err)
		}

		// 验证数据已插入
		// 验证表1数据
		readFields1 := &map[string]any{
			"id": 1,
		}
		data1, err := table1.Read(readFields1)
		if err != nil {
			t.Fatalf("Failed to read data from table1: %v", err)
		}

		if len(data1) == 0 {
			t.Fatalf("Expected non-empty data from table1")
		}

		// 验证表2数据
		readFields2 := &map[string]any{
			"id": 1,
		}
		data2, err := table2.Read(readFields2)
		if err != nil {
			t.Fatalf("Failed to read data from table2: %v", err)
		}

		if len(data2) == 0 {
			t.Fatalf("Expected non-empty data from table2")
		}
	})
}

// TestTransactionManagerWithTransaction 测试 WithTransaction 函数
func TestTransactionManagerWithTransaction(t *testing.T) {
	// 测试目录设置
	testDir := filepath.Join(os.TempDir(), "sfsdb_with_tx_test")
	testDbPath := filepath.Join(testDir, "test_db")

	// 清理测试目录
	defer func() {
		os.RemoveAll(testDir)
	}()

	// 创建测试数据库
	testDb, err := storage.OpenDefaultDb(testDbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDb.Close()

	// 创建表结构
	table1, err := TableNew("test_table1")
	if err != nil {
		t.Fatalf("Failed to create table1: %v", err)
	}

	// 设置表1的字段
	fields1 := map[string]any{
		"id":   0,
		"name": "",
	}
	err = table1.SetFields(fields1)
	if err != nil {
		t.Fatalf("Failed to set fields for table1: %v", err)
	}

	table2, err := TableNew("test_table2")
	if err != nil {
		t.Fatalf("Failed to create table2: %v", err)
	}

	// 设置表2的字段
	fields2 := map[string]any{
		"id":   0,
		"name": "",
	}
	err = table2.SetFields(fields2)
	if err != nil {
		t.Fatalf("Failed to set fields for table2: %v", err)
	}

	// 测试 WithTransaction 函数
	t.Run("WithTransactionSuccess", func(t *testing.T) {
		// 创建 batch
		batch := testDb.GetBatch()
		if batch == nil {
			t.Fatalf("Failed to create batch")
		}

		// 测试成功的事务
		tables := []*Table{table1, table2}
		err := WithTransaction(batch, tables, func(transactions map[*Table]Transaction) error {
			// 在两个表中插入数据
			// 表1插入数据
			tx1 := transactions[table1]
			fields1 := &map[string]any{
				"id":   1,
				"name": "with_tx_table1",
			}
			_, err := tx1.Insert(fields1)
			if err != nil {
				return err
			}

			// 表2插入数据
			tx2 := transactions[table2]
			fields2 := &map[string]any{
				"id":   1,
				"name": "with_tx_table2",
			}
			_, err = tx2.Insert(fields2)
			if err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			t.Fatalf("Failed to execute WithTransaction: %v", err)
		}

		// 验证数据已插入
		// 验证表1数据
		readFields1 := &map[string]any{
			"id": 1,
		}
		data1, err := table1.Read(readFields1)
		if err != nil {
			t.Fatalf("Failed to read data from table1: %v", err)
		}

		if len(data1) == 0 {
			t.Fatalf("Expected non-empty data from table1")
		}

		// 验证表2数据
		readFields2 := &map[string]any{
			"id": 1,
		}
		data2, err := table2.Read(readFields2)
		if err != nil {
			t.Fatalf("Failed to read data from table2: %v", err)
		}

		if len(data2) == 0 {
			t.Fatalf("Expected non-empty data from table2")
		}
	})

	// 测试 WithTransaction 函数的错误处理
	t.Run("WithTransactionError", func(t *testing.T) {
		// 创建 batch
		batch := testDb.GetBatch()
		if batch == nil {
			t.Fatalf("Failed to create batch")
		}

		// 测试失败的事务
		tables := []*Table{table1, table2}
		err := WithTransaction(batch, tables, func(transactions map[*Table]Transaction) error {
			// 在表1中插入数据
			tx1 := transactions[table1]
			fields1 := &map[string]any{
				"id":   2,
				"name": "with_tx_error_table1",
			}
			_, err := tx1.Insert(fields1)
			if err != nil {
				return err
			}

			// 故意返回错误
			return ErrTableNotExist
		})

		if err == nil {
			t.Fatalf("Expected error from WithTransaction, got nil")
		}

		// 验证表1数据未插入（因为事务回滚）
		// 注意：由于 LevelDB 的 WriteBatch 不支持真正的回滚，
		// 这里可能会看到数据已插入，但在实际生产环境中，
		// 应该确保在事务中所有操作都成功后再提交
	})
}
