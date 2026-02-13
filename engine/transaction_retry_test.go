package engine

import (
	"testing"
	"time"
)

// TestTransactionRetry 测试事务重试机制
func TestTransactionRetry(t *testing.T) {
	// 创建测试表
	table, err := TableNew("test_retry_table")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置字段
	fields := map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 测试1: 正常事务（不需要重试）
	t.Run("NormalTransaction", func(t *testing.T) {
		// 开始事务
		tx, err := table.Begin()
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}

		// 插入记录
		record := map[string]any{
			"id":   1,
			"name": "Alice",
			"age":  20,
		}
		_, err = tx.Insert(&record)
		if err != nil {
			t.Fatalf("Failed to insert record: %v", err)
		}

		// 提交事务
		err = tx.Commit()
		if err != nil {
			t.Fatalf("Failed to commit transaction: %v", err)
		}

		// 验证记录
		// 使用事务读取验证
		txRead, err := table.Begin()
		if err != nil {
			t.Fatalf("Failed to begin read transaction: %v", err)
		}
		result, err := txRead.Read(&map[string]any{"id": 1})
		if err != nil {
			t.Fatalf("Failed to read record: %v", err)
		}
		txRead.Commit()
		if len(result) == 0 {
			t.Fatalf("Record not found after commit")
		}
	})

	// 测试2: 自定义事务选项（增加重试次数）
	t.Run("CustomRetryOptions", func(t *testing.T) {
		// 开始事务，使用自定义选项
		options := DefaultTransactionOptions()
		options.MaxRetries = 5
		options.InitialRetryDelay = 5 * time.Millisecond
		options.RetryBackoffFactor = 1.5

		tx, err := table.BeginWithOptions(options)
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}

		// 插入记录
		record := map[string]any{
			"id":   2,
			"name": "Bob",
			"age":  25,
		}
		_, err = tx.Insert(&record)
		if err != nil {
			t.Fatalf("Failed to insert record: %v", err)
		}

		// 提交事务
		err = tx.Commit()
		if err != nil {
			t.Fatalf("Failed to commit transaction: %v", err)
		}

		// 验证记录
		txRead, err := table.Begin()
		if err != nil {
			t.Fatalf("Failed to begin read transaction: %v", err)
		}
		result, err := txRead.Read(&map[string]any{"id": 2})
		if err != nil {
			t.Fatalf("Failed to read record: %v", err)
		}
		txRead.Commit()
		if len(result) == 0 {
			t.Fatalf("Record not found after commit")
		}
	})

	// 测试3: 禁用重试
	t.Run("NoRetry", func(t *testing.T) {
		// 开始事务，禁用重试
		options := DefaultTransactionOptions()
		options.MaxRetries = 0

		tx, err := table.BeginWithOptions(options)
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}

		// 插入记录
		record := map[string]any{
			"id":   3,
			"name": "Charlie",
			"age":  30,
		}
		_, err = tx.Insert(&record)
		if err != nil {
			t.Fatalf("Failed to insert record: %v", err)
		}

		// 提交事务
		err = tx.Commit()
		if err != nil {
			t.Fatalf("Failed to commit transaction: %v", err)
		}

		// 验证记录
		txRead, err := table.Begin()
		if err != nil {
			t.Fatalf("Failed to begin read transaction: %v", err)
		}
		result, err := txRead.Read(&map[string]any{"id": 3})
		if err != nil {
			t.Fatalf("Failed to read record: %v", err)
		}
		txRead.Commit()
		if len(result) == 0 {
			t.Fatalf("Record not found after commit")
		}
	})

	// 测试4: 批量操作的事务重试
	t.Run("BatchTransactionRetry", func(t *testing.T) {
		// 开始事务
		tx, err := table.Begin()
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}

		// 批量插入多条记录
		records := []map[string]any{
			{"id": 4, "name": "David", "age": 35},
			{"id": 5, "name": "Eve", "age": 40},
			{"id": 6, "name": "Frank", "age": 45},
		}

		for _, record := range records {
			_, err := tx.Insert(&record)
			if err != nil {
				t.Fatalf("Failed to insert record: %v", err)
			}
		}

		// 提交事务
		err = tx.Commit()
		if err != nil {
			t.Fatalf("Failed to commit transaction: %v", err)
		}

		// 验证记录
		for _, record := range records {
			txRead, err := table.Begin()
			if err != nil {
				t.Fatalf("Failed to begin read transaction: %v", err)
			}
			result, err := txRead.Read(&map[string]any{"id": record["id"]})
			if err != nil {
				t.Fatalf("Failed to read record %v: %v", record["id"], err)
			}
			txRead.Commit()
			if len(result) == 0 {
				t.Fatalf("Record %v not found after commit", record["id"])
			}
		}
	})
}
