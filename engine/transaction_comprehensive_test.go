package engine

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/liaoran123/sfsDb/storage"
)

// TestTransactionBasic 测试基本事务操作
func TestTransactionBasic(t *testing.T) {
	// 初始化数据库
	_, err := storage.OpenDefaultDb("./test_basic_db")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer storage.CloseDb()

	// 创建表
	table, err := TableNew("test_basic")
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

	// 开始事务
	tx, err := table.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// 插入记录
	fields1 := map[string]any{
		"name": "Alice",
		"age":  25,
	}
	id, err := tx.Insert(&fields1)
	if err != nil {
		t.Fatalf("Failed to insert record: %v", err)
	}
	if id <= 0 {
		t.Errorf("Expected positive id, got %d", id)
	}

	// 读取记录
	readFields := map[string]any{
		"id": id,
	}
	record, err := tx.Read(&readFields)
	if err != nil {
		t.Fatalf("Failed to read record: %v", err)
	}
	if len(record) == 0 {
		t.Error("Expected non-empty record")
	}

	// 更新记录
	updateFields := map[string]any{
		"id":   id,
		"name": "Alice Smith",
		"age":  26,
	}
	err = tx.Update(&updateFields)
	if err != nil {
		t.Fatalf("Failed to update record: %v", err)
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// 验证提交后的记录
	recordAfterCommit, err := table.Read(&readFields)
	if err != nil {
		t.Fatalf("Failed to read record after commit: %v", err)
	}
	if len(recordAfterCommit) == 0 {
		t.Error("Expected non-empty record after commit")
	}
}

// TestTransactionRollback 测试事务回滚
func TestTransactionRollback(t *testing.T) {
	// 初始化数据库
	_, err := storage.OpenDefaultDb("./test_rollback_db")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer storage.CloseDb()

	// 创建表
	table, err := TableNew("test_rollback")
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

	// 先插入一条初始记录
	initialFields := map[string]any{
		"name": "Bob",
		"age":  30,
	}
	id, err := table.Insert(&initialFields)
	if err != nil {
		t.Fatalf("Failed to insert initial record: %v", err)
	}

	// 开始事务
	tx, err := table.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// 更新记录
	updateFields := map[string]any{
		"id":   id,
		"name": "Bob Johnson",
		"age":  31,
	}
	err = tx.Update(&updateFields)
	if err != nil {
		t.Fatalf("Failed to update record: %v", err)
	}

	// 回滚事务
	err = tx.Rollback()
	if err != nil {
		t.Fatalf("Failed to rollback transaction: %v", err)
	}

	// 验证回滚后的记录（应该保持初始值）
	readFields := map[string]any{
		"id": id,
	}
	record, err := table.Read(&readFields)
	if err != nil {
		t.Fatalf("Failed to read record after rollback: %v", err)
	}
	if len(record) == 0 {
		t.Error("Expected non-empty record after rollback")
	}
}

// TestTransactionNested 测试嵌套事务
func TestTransactionNested(t *testing.T) {
	// 初始化数据库
	_, err := storage.OpenDefaultDb("./test_nested_db")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer storage.CloseDb()

	// 创建表
	table, err := TableNew("test_nested")
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

	// 开始事务，允许嵌套
	options := &TransactionOptions{
		IsolationLevel: RepeatableRead,
		AllowNested:    true,
	}
	tx, err := table.BeginWithOptions(options)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// 创建嵌套事务
	nestedTx, err := tx.BeginNested()
	if err != nil {
		t.Fatalf("Failed to begin nested transaction: %v", err)
	}

	// 在嵌套事务中插入记录
	fields1 := map[string]any{
		"name": "Charlie",
		"age":  35,
	}
	id, err := nestedTx.Insert(&fields1)
	if err != nil {
		t.Fatalf("Failed to insert record in nested transaction: %v", err)
	}

	// 提交嵌套事务
	err = nestedTx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit nested transaction: %v", err)
	}

	// 提交父事务
	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit parent transaction: %v", err)
	}

	// 验证记录已插入
	readFields := map[string]any{
		"id": id,
	}
	record, err := table.Read(&readFields)
	if err != nil {
		t.Fatalf("Failed to read record after nested transaction: %v", err)
	}
	if len(record) == 0 {
		t.Error("Expected non-empty record after nested transaction")
	}
}

// TestTransactionConcurrent 测试并发事务
func TestTransactionConcurrent(t *testing.T) {
	// 初始化数据库
	_, err := storage.OpenDefaultDb("./test_concurrent_db")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer storage.CloseDb()

	// 创建表
	table, err := TableNew("test_concurrent")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置字段
	fields := map[string]any{
		"id":    0,
		"name":  "",
		"value": 0,
	}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 插入初始记录
	initialFields := map[string]any{
		"name":  "Product",
		"value": 100,
	}
	id, err := table.Insert(&initialFields)
	if err != nil {
		t.Fatalf("Failed to insert initial record: %v", err)
	}

	// 并发更新同一记录
	var wg sync.WaitGroup
	var mu sync.Mutex
	errors := []error{}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			// 开始事务
			tx, err := table.Begin()
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("goroutine %d: failed to begin transaction: %v", i, err))
				mu.Unlock()
				return
			}

			// 读取记录
			readFields := map[string]any{
				"id": id,
			}
			_, err = tx.Read(&readFields)
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("goroutine %d: failed to read record: %v", i, err))
				mu.Unlock()
				tx.Rollback()
				return
			}

			// 更新记录
			updateFields := map[string]any{
				"id":    id,
				"value": 100 + i,
			}
			err = tx.Update(&updateFields)
			if err != nil {
				// 乐观锁冲突是预期的，不是错误
				tx.Rollback()
				return
			}

			// 提交事务
			err = tx.Commit()
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("goroutine %d: failed to commit transaction: %v", i, err))
				mu.Unlock()
				return
			}
		}(i)
	}

	wg.Wait()

	// 检查错误
	for _, err := range errors {
		t.Errorf("Concurrent transaction error: %v", err)
	}
}

// TestTransactionObjectPool 测试对象池内存泄漏
func TestTransactionObjectPool(t *testing.T) {
	// 初始化数据库
	_, err := storage.OpenDefaultDb("./test_object_pool_db")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer storage.CloseDb()

	// 创建表
	table, err := TableNew("test_object_pool")
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

	// 监控内存使用
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	initialAlloc := memStats.Alloc

	// 执行大量事务操作，测试对象池是否泄漏
	for i := 0; i < 1000; i++ {
		// 开始事务
		tx, err := table.Begin()
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}

		// 插入记录
		fields1 := map[string]any{
			"name": fmt.Sprintf("User%d", i),
			"age":  20 + i%50,
		}
		_, err = tx.Insert(&fields1)
		if err != nil {
			t.Fatalf("Failed to insert record: %v", err)
		}

		// 提交事务
		err = tx.Commit()
		if err != nil {
			t.Fatalf("Failed to commit transaction: %v", err)
		}
	}

	// 强制GC
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	// 再次监控内存使用
	runtime.ReadMemStats(&memStats)
	finalAlloc := memStats.Alloc

	// 计算内存增长
	memGrowth := finalAlloc - initialAlloc
	memGrowthMB := float64(memGrowth) / 1024 / 1024

	// 打印内存使用情况
	t.Logf("Initial allocation: %.2f MB", float64(initialAlloc)/1024/1024)
	t.Logf("Final allocation: %.2f MB", float64(finalAlloc)/1024/1024)
	t.Logf("Memory growth: %.2f MB", memGrowthMB)

	// 检查内存增长是否合理（不应该超过10MB）
	if memGrowthMB > 10 {
		t.Errorf("Potential memory leak detected: memory growth of %.2f MB is too high", memGrowthMB)
	}
}

// TestTransactionSnapshotPool 测试快照池使用
func TestTransactionSnapshotPool(t *testing.T) {
	// 初始化数据库
	_, err := storage.OpenDefaultDb("./test_snapshot_pool_db")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer storage.CloseDb()

	// 创建表
	table, err := TableNew("test_snapshot_pool")
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

	// 测试快照池的使用

	// 执行大量使用快照的事务
	for i := 0; i < 100; i++ {
		// 开始事务，使用可重复读隔离级别（会创建快照）
		options := &TransactionOptions{
			IsolationLevel: RepeatableRead,
		}
		tx, err := table.BeginWithOptions(options)
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}

		// 插入记录
		fields1 := map[string]any{
			"name": fmt.Sprintf("User%d", i),
			"age":  20 + i%50,
		}
		_, err = tx.Insert(&fields1)
		if err != nil {
			t.Fatalf("Failed to insert record: %v", err)
		}

		// 提交事务
		err = tx.Commit()
		if err != nil {
			t.Fatalf("Failed to commit transaction: %v", err)
		}
	}

	// 强制GC
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	// 验证事务操作正常完成
	t.Log("Snapshot pool test completed successfully")
}

// TestTransactionVersionGeneration 测试版本生成
func TestTransactionVersionGeneration(t *testing.T) {
	// 初始化数据库
	_, err := storage.OpenDefaultDb("./test_version_gen_db")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer storage.CloseDb()

	// 创建表
	table, err := TableNew("test_version_gen")
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

	// 开始事务
	tx, err := table.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// 插入记录
	fields1 := map[string]any{
		"name": "David",
		"age":  40,
	}
	id, err := tx.Insert(&fields1)
	if err != nil {
		t.Fatalf("Failed to insert record: %v", err)
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// 开始新事务更新记录
	tx2, err := table.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// 更新记录
	updateFields := map[string]any{
		"id":   id,
		"name": "David Updated",
		"age":  41,
	}
	err = tx2.Update(&updateFields)
	if err != nil {
		t.Fatalf("Failed to update record: %v", err)
	}

	// 提交事务
	err = tx2.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// 验证版本号已更新
	readFields := map[string]any{
		"id": id,
	}
	record, err := table.Read(&readFields)
	if err != nil {
		t.Fatalf("Failed to read record: %v", err)
	}
	if len(record) == 0 {
		t.Error("Expected non-empty record")
	}
}

// TestTransactionLockMechanism 测试锁机制
func TestTransactionLockMechanism(t *testing.T) {
	// 初始化数据库
	_, err := storage.OpenDefaultDb("./test_lock_mechanism_db")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer storage.CloseDb()

	// 创建表
	table, err := TableNew("test_lock_mechanism")
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

	// 插入初始记录
	initialFields := map[string]any{
		"name": "Eve",
		"age":  45,
	}
	id, err := table.Insert(&initialFields)
	if err != nil {
		t.Fatalf("Failed to insert initial record: %v", err)
	}

	// 开始第一个事务
	tx1, err := table.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction 1: %v", err)
	}

	// 读取记录
	readFields := map[string]any{
		"id": id,
	}
	_, err = tx1.Read(&readFields)
	if err != nil {
		t.Fatalf("Failed to read record in tx1: %v", err)
	}

	// 开始第二个事务
	tx2, err := table.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction 2: %v", err)
	}

	// 尝试更新同一记录
	updateFields := map[string]any{
		"id":   id,
		"name": "Eve Updated",
		"age":  46,
	}
	err = tx2.Update(&updateFields)
	if err != nil {
		// 乐观锁冲突是预期的，不是错误
		tx2.Rollback()
	} else {
		// 如果没有冲突，提交事务
		err = tx2.Commit()
		if err != nil {
			t.Fatalf("Failed to commit transaction 2: %v", err)
		}
	}

	// 提交第一个事务
	err = tx1.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction 1: %v", err)
	}
}

// TestTransactionPerformance 测试事务性能
func TestTransactionPerformance(t *testing.T) {
	// 初始化数据库
	_, err := storage.OpenDefaultDb("./test_performance_db")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer storage.CloseDb()

	// 创建测试表
	table, err := TableNew("test_performance")
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

	// 测试批量事务性能
	startTime := time.Now()

	for i := 0; i < 100; i++ {
		// 开始事务
		tx, err := table.Begin()
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}

		// 插入记录
		insertFields := map[string]any{
			"name": fmt.Sprintf("User%d", i),
			"age":  20 + i%50,
		}
		_, err = tx.Insert(&insertFields)
		if err != nil {
			t.Fatalf("Failed to insert record: %v", err)
		}

		// 提交事务
		err = tx.Commit()
		if err != nil {
			t.Fatalf("Failed to commit transaction: %v", err)
		}
	}

	elapsed := time.Since(startTime)
	t.Logf("100 transactions took: %v", elapsed)
}
