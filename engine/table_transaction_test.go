package engine

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestTableTransaction 测试 TableTransaction 功能
func TestTableTransaction(t *testing.T) {
	// 使用唯一表名，避免测试数据累积
	tableName := fmt.Sprintf("test_transaction_%d", time.Now().UnixNano())

	// 创建表
	table, err := TableNew(tableName)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{
		"id":    0,
		"name":  "",
		"age":   0,
		"email": "",
	}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键索引
	pk, err := DefaultPrimaryKeyNew("pk_id")
	if err != nil {
		t.Fatalf("Failed to create primary key: %v", err)
	}
	pk.AddFields("id")
	err = table.CreateIndex(pk)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// 测试1: 基本事务操作
	t.Run("BasicTransactionOperations", func(t *testing.T) {
		// 开始事务
		tx, err := table.Begin()
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}

		// 插入记录
		insertRecord := map[string]any{"id": 1, "name": "张三", "age": 25, "email": "zhangsan@example.com"}
		id, err := tx.Insert(&insertRecord)
		if err != nil {
			t.Fatalf("Failed to insert record in transaction: %v", err)
		}
		if id != 1 {
			t.Errorf("Expected id=1, got %d", id)
		}

		// 读取刚插入的记录（测试读自己的写）
		readRecord := map[string]any{"id": 1}
		recordData, err := tx.Read(&readRecord)
		if err != nil {
			t.Fatalf("Failed to read record in transaction: %v", err)
		}
		if len(recordData) == 0 {
			t.Errorf("Expected record data, got empty")
		}

		// 更新记录
		updateRecord := map[string]any{"id": 1, "age": 26}
		err = tx.Update(&updateRecord)
		if err != nil {
			t.Fatalf("Failed to update record in transaction: %v", err)
		}

		// 提交事务
		err = tx.Commit()
		if err != nil {
			t.Fatalf("Failed to commit transaction: %v", err)
		}

		// 验证记录已被更新
		iter := table.Search(&readRecord)
		if iter == nil {
			t.Fatalf("Failed to search record after transaction")
		}
		defer iter.Release()

		records := iter.GetRecords(true)
		if len(records) != 1 {
			t.Fatalf("Expected 1 record, got %d", len(records))
		}

		if records[0]["age"] != 26 {
			t.Errorf("Expected age=26, got %v", records[0]["age"])
		}
	})

	// 测试2: 事务回滚
	t.Run("TransactionRollback", func(t *testing.T) {
		// 开始事务
		tx, err := table.Begin()
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}

		// 插入记录
		insertRecord := map[string]any{"id": 2, "name": "李四", "age": 30, "email": "lisi@example.com"}
		_, err = tx.Insert(&insertRecord)
		if err != nil {
			t.Fatalf("Failed to insert record in transaction: %v", err)
		}

		// 回滚事务
		err = tx.Rollback()
		if err != nil {
			t.Fatalf("Failed to rollback transaction: %v", err)
		}

		// 验证记录未被插入
		readRecord := map[string]any{"id": 2}
		iter := table.Search(&readRecord)
		if iter == nil {
			t.Fatalf("Failed to search record after rollback")
		}
		defer iter.Release()

		records := iter.GetRecords(true)
		if len(records) != 0 {
			t.Errorf("Expected 0 records after rollback, got %d", len(records))
		}
	})

	// 测试3: 事务中的删除操作
	t.Run("TransactionDelete", func(t *testing.T) {
		// 先插入一条记录
		insertRecord := map[string]any{"id": 3, "name": "王五", "age": 35, "email": "wangwu@example.com"}
		_, err := table.Insert(&insertRecord)
		if err != nil {
			t.Fatalf("Failed to insert record before transaction: %v", err)
		}

		// 开始事务
		tx, err := table.Begin()
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}

		// 删除记录
		deleteRecord := map[string]any{"id": 3}
		err = tx.Delete(&deleteRecord)
		if err != nil {
			t.Fatalf("Failed to delete record in transaction: %v", err)
		}

		// 提交事务
		err = tx.Commit()
		if err != nil {
			t.Fatalf("Failed to commit transaction: %v", err)
		}

		// 验证记录已被删除
		iter := table.Search(&deleteRecord)
		if iter == nil {
			t.Fatalf("Failed to search record after delete")
		}
		defer iter.Release()

		records := iter.GetRecords(true)
		if len(records) != 0 {
			t.Errorf("Expected 0 records after delete, got %d", len(records))
		}
	})

	// 测试4: 事务中的搜索操作
	t.Run("TransactionSearch", func(t *testing.T) {
		// 先插入几条记录
		records := []map[string]any{
			{"id": 4, "name": "赵六", "age": 40, "email": "zhaoliu@example.com"},
			{"id": 5, "name": "钱七", "age": 45, "email": "qianqi@example.com"},
			{"id": 6, "name": "孙八", "age": 50, "email": "sunba@example.com"},
		}

		for _, record := range records {
			_, err := table.Insert(&record)
			if err != nil {
				t.Fatalf("Failed to insert record before transaction: %v", err)
			}
		}

		// 开始事务
		tx, err := table.Begin()
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}

		// 搜索记录（使用id字段，因为它有主键索引）
		searchRecord := map[string]any{"id": 4}
		iter := tx.Search(&searchRecord)
		if iter == nil {
			t.Fatalf("Failed to search records in transaction")
		}
		defer iter.Release()

		searchResults := iter.GetRecords(true)
		if len(searchResults) == 0 {
			t.Errorf("Expected search results, got empty")
		}

		// 提交事务
		err = tx.Commit()
		if err != nil {
			t.Fatalf("Failed to commit transaction: %v", err)
		}
	})

	// 测试5: 事务的重复提交和回滚
	t.Run("TransactionDuplicateOperations", func(t *testing.T) {
		// 开始事务
		tx, err := table.Begin()
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}

		// 提交事务
		err = tx.Commit()
		if err != nil {
			t.Fatalf("Failed to commit transaction: %v", err)
		}

		// 尝试再次提交，应该失败
		err = tx.Commit()
		if err == nil {
			t.Errorf("Expected error when committing already committed transaction, got nil")
		}

		// 尝试回滚，应该失败
		err = tx.Rollback()
		if err == nil {
			t.Errorf("Expected error when rolling back already committed transaction, got nil")
		}
	})
}

// TestTableTransaction_Concurrent 测试并发事务
func TestTableTransaction_Concurrent(t *testing.T) {
	// 使用唯一表名，避免测试数据累积
	tableName := fmt.Sprintf("test_transaction_concurrent_%d", time.Now().UnixNano())

	// 创建表
	table, err := TableNew(tableName)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键索引
	pk, err := DefaultPrimaryKeyNew("pk_id")
	if err != nil {
		t.Fatalf("Failed to create primary key: %v", err)
	}
	pk.AddFields("id")
	err = table.CreateIndex(pk)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// 并发插入测试
	const concurrentCount = 5
	var wg sync.WaitGroup
	errCh := make(chan error, concurrentCount)

	for i := 0; i < concurrentCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// 开始事务
			tx, err := table.Begin()
			if err != nil {
				errCh <- fmt.Errorf("Failed to begin transaction: %v", err)
				return
			}

			// 插入记录
			insertRecord := map[string]any{"id": id + 1, "name": fmt.Sprintf("User%d", id+1), "age": 20 + id}
			_, err = tx.Insert(&insertRecord)
			if err != nil {
				tx.Rollback()
				errCh <- fmt.Errorf("Failed to insert record in transaction: %v", err)
				return
			}

			// 提交事务
			err = tx.Commit()
			if err != nil {
				errCh <- fmt.Errorf("Failed to commit transaction: %v", err)
				return
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	// 检查错误
	for err := range errCh {
		if err != nil {
			t.Fatalf("Concurrent transaction error: %v", err)
		}
	}

	// 验证所有记录都已插入
	iter := table.Search(&map[string]any{"id": nil})
	if iter == nil {
		t.Fatalf("Failed to search records after concurrent transactions")
	}
	defer iter.Release()

	records := iter.GetRecords(true)
	if len(records) != concurrentCount {
		t.Errorf("Expected %d records after concurrent transactions, got %d", concurrentCount, len(records))
	}
}

// TestTableTransaction_ReadOnly 测试只读事务
func TestTableTransaction_ReadOnly(t *testing.T) {
	// 使用唯一表名，避免测试数据累积
	tableName := fmt.Sprintf("test_transaction_readonly_%d", time.Now().UnixNano())

	// 创建表
	table, err := TableNew(tableName)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键索引
	pk, err := DefaultPrimaryKeyNew("pk_id")
	if err != nil {
		t.Fatalf("Failed to create primary key: %v", err)
	}
	pk.AddFields("id")
	err = table.CreateIndex(pk)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// 插入测试数据
	insertRecord := map[string]any{"id": 1, "name": "张三", "age": 25}
	_, err = table.Insert(&insertRecord)
	if err != nil {
		t.Fatalf("Failed to insert record before transaction: %v", err)
	}

	// 开始事务
	tx, err := table.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// 读取记录
	readRecord := map[string]any{"id": 1}
	recordData, err := tx.Read(&readRecord)
	if err != nil {
		t.Fatalf("Failed to read record in transaction: %v", err)
	}
	if len(recordData) == 0 {
		t.Errorf("Expected record data, got empty")
	}

	// 搜索记录
	iter := tx.Search(&readRecord)
	if iter == nil {
		t.Fatalf("Failed to search record in transaction")
	}
	defer iter.Release()

	records := iter.GetRecords(true)
	if len(records) != 1 {
		t.Errorf("Expected 1 record, got %d", len(records))
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}
}
