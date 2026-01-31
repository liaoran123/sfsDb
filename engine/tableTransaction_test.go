package engine

import (
	"fmt"
	"testing"

	"github.com/liaoran123/sfsDb/util"
)

// TestTableTransaction_BasicOperations 测试基本事务操作
func TestTableTransaction_BasicOperations(t *testing.T) {
	// 创建测试表
	table, err := createTestTable()
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// 开始事务
	tx, err := table.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// 插入记录
	fields1 := &map[string]any{
		"name": "Test1",
		"age":  20,
	}
	id, err := tx.Insert(fields1)
	if err != nil {
		t.Fatalf("Failed to insert in transaction: %v", err)
	}

	// 验证可以读取自己的写操作
	readFields := &map[string]any{"id": id}
	record, err := tx.Read(readFields)
	if err != nil {
		t.Fatalf("Failed to read in transaction: %v", err)
	}
	if len(record) == 0 {
		t.Fatalf("Expected to read inserted record, but got nothing")
	}

	// 更新记录
	updateFields := &map[string]any{
		"id":   id,
		"name": "UpdatedTest1",
		"age":  21,
	}
	err = tx.Update(updateFields)
	if err != nil {
		t.Fatalf("Failed to update in transaction: %v", err)
	}

	// 验证更新
	updatedRecord, err := tx.Read(readFields)
	if err != nil {
		t.Fatalf("Failed to read updated record: %v", err)
	}
	if len(updatedRecord) == 0 {
		t.Fatalf("Expected to read updated record, but got nothing")
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// 验证提交后可以读取到数据
	finalRecord, err := table.Read(readFields)
	if err != nil {
		t.Fatalf("Failed to read after commit: %v", err)
	}
	if len(finalRecord) == 0 {
		t.Fatalf("Expected to read committed record, but got nothing")
	}
}

// TestTableTransaction_Rollback 测试事务回滚
func TestTableTransaction_Rollback(t *testing.T) {
	// 创建测试表
	table, err := createTestTable()
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// 开始事务
	tx, err := table.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// 插入记录
	fields := &map[string]any{
		"name": "TestRollback",
		"age":  30,
	}
	id, err := tx.Insert(fields)
	if err != nil {
		t.Fatalf("Failed to insert in transaction: %v", err)
	}

	// 验证事务内可以读取
	readFields := &map[string]any{"id": id}
	_, err = tx.Read(readFields)
	if err != nil {
		t.Fatalf("Failed to read in transaction: %v", err)
	}

	// 回滚事务
	err = tx.Rollback()
	if err != nil {
		t.Fatalf("Failed to rollback transaction: %v", err)
	}

	// 验证回滚后读取不到数据
	record, err := table.Read(readFields)
	if err == nil && record != nil {
		t.Fatalf("Expected no record when reading rolled back record, but got one")
	}
}

// TestTableTransaction_UpdateCacheOnlyRecord 测试更新只存在于缓存中的记录
func TestTableTransaction_UpdateCacheOnlyRecord(t *testing.T) {
	// 创建测试表
	table, err := createTestTable()
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// 开始事务
	tx, err := table.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// 插入记录
	fields := &map[string]any{
		"name": "CacheOnly",
		"age":  25,
	}
	id, err := tx.Insert(fields)
	if err != nil {
		t.Fatalf("Failed to insert in transaction: %v", err)
	}

	// 更新记录（此时记录只存在于缓存中）
	updateFields := &map[string]any{
		"id":   id,
		"name": "UpdatedCacheOnly",
		"age":  26,
	}
	err = tx.Update(updateFields)
	if err != nil {
		t.Fatalf("Failed to update cache-only record: %v", err)
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// 验证提交后可以读取到更新后的数据
	readFields := &map[string]any{"id": id}
	record, err := table.Read(readFields)
	if err != nil {
		t.Fatalf("Failed to read after commit: %v", err)
	}
	if len(record) == 0 {
		t.Fatalf("Expected to read committed record, but got nothing")
	}
}

// TestTableTransaction_Search 测试事务中的搜索操作
func TestTableTransaction_Search(t *testing.T) {
	// 创建测试表
	table, err := createTestTable()
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// 开始事务
	tx, err := table.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// 插入多条记录
	for i := 0; i < 5; i++ {
		fields := &map[string]any{
			"name": fmt.Sprintf("Test%d", i),
			"age":  i + 18,
		}
		_, err := tx.Insert(fields)
		if err != nil {
			t.Fatalf("Failed to insert in transaction: %v", err)
		}
	}

	// 注意：由于没有为name字段创建索引，这里使用id字段进行搜索
	// 在实际使用中，应该为需要搜索的字段创建索引
	searchFields := &map[string]any{"id": 1}
	iter := tx.Search(searchFields, util.Like)
	if iter == nil {
		t.Fatalf("Failed to create search iterator")
	}

	// 验证可以遍历结果
	count := 0
	for iter.Next() {
		count++
	}
	if count == 0 {
		t.Fatalf("Expected to find search results, but got none")
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}
}

// TestTableTransaction_CommittedTransaction 测试事务已提交后尝试操作
func TestTableTransaction_CommittedTransaction(t *testing.T) {
	// 创建测试表
	table, err := createTestTable()
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

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

	// 尝试在已提交的事务中插入
	fields := &map[string]any{
		"name": "Test",
		"age":  20,
	}
	_, err = tx.Insert(fields)
	if err == nil {
		t.Fatalf("Expected error when inserting in committed transaction, but got none")
	}

	// 尝试在已提交的事务中读取
	readFields := &map[string]any{"id": 1}
	_, err = tx.Read(readFields)
	if err == nil {
		t.Fatalf("Expected error when reading in committed transaction, but got none")
	}
}

// TestTableTransaction_ConcurrentTransactions 测试并发事务
func TestTableTransaction_ConcurrentTransactions(t *testing.T) {
	// 创建测试表
	table, err := createTestTable()
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// 启动多个并发事务
	const concurrentCount = 3
	errChan := make(chan error, concurrentCount)

	for i := 0; i < concurrentCount; i++ {
		go func(id int) {
			tx, err := table.Begin()
			if err != nil {
				errChan <- fmt.Errorf("Failed to begin transaction %d: %v", id, err)
				return
			}

			// 插入记录
			fields := &map[string]any{
				"name": fmt.Sprintf("Concurrent%d", id),
				"age":  id + 20,
			}
			_, err = tx.Insert(fields)
			if err != nil {
				errChan <- fmt.Errorf("Failed to insert in transaction %d: %v", id, err)
				tx.Rollback()
				return
			}

			// 提交事务
			err = tx.Commit()
			if err != nil {
				errChan <- fmt.Errorf("Failed to commit transaction %d: %v", id, err)
				return
			}

			errChan <- nil
		}(i)
	}

	// 收集错误
	errors := []error{}
	for i := 0; i < concurrentCount; i++ {
		if err := <-errChan; err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		for _, err := range errors {
			t.Errorf("Concurrent transaction error: %v", err)
		}
		t.Fatalf("Got %d errors in concurrent transactions", len(errors))
	}
}

// TestTableTransaction_Delete 测试事务中的删除操作
func TestTableTransaction_Delete(t *testing.T) {
	// 创建测试表
	table, err := createTestTable()
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// 先插入一条记录
	fields := &map[string]any{
		"name": "ToDelete",
		"age":  30,
	}
	id, err := table.Insert(fields)
	if err != nil {
		t.Fatalf("Failed to insert record: %v", err)
	}

	// 开始事务
	tx, err := table.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// 删除记录
	deleteFields := &map[string]any{"id": id}
	err = tx.Delete(deleteFields)
	if err != nil {
		t.Fatalf("Failed to delete in transaction: %v", err)
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// 验证提交后读取不到
	record, err := table.Read(deleteFields)
	if err == nil && record != nil {
		t.Fatalf("Expected no record when reading deleted record after commit, but got one")
	}
}

// createTestTable 创建测试表
func createTestTable() (*Table, error) {
	// 创建表
	table, err := TableNew("test_transaction_table")
	if err != nil {
		return nil, err
	}

	// 设置字段
	fields := map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	}
	err = table.SetFields(fields)
	if err != nil {
		return nil, err
	}

	return table, nil
}
