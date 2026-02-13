package engine

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/liaoran123/sfsDb/storage"
)

// BenchmarkTransactionComparison 比较不同事务操作的性能
func BenchmarkTransactionComparison(b *testing.B) {
	// 测试目录设置
	testDir := filepath.Join(os.TempDir(), "sfsdb_benchmark_comparison")
	testDbPath := filepath.Join(testDir, "test_db")

	// 清理测试目录
	defer func() {
		os.RemoveAll(testDir)
	}()

	// 创建测试数据库
	testDb, err := storage.OpenDefaultDb(testDbPath)
	if err != nil {
		b.Fatalf("Failed to open test database: %v", err)
	}
	defer testDb.Close()

	// 创建表结构
	table, err := TableNew("benchmark_table")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// 设置表的字段
	fields := map[string]any{
		"id":   0,
		"name": "",
		"value": 0,
	}
	err = table.SetFields(fields)
	if err != nil {
		b.Fatalf("Failed to set fields for table: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 创建 batch
		batch := testDb.GetBatch()
		if batch == nil {
			b.Fatalf("Failed to create batch")
		}

		// 创建事务管理器
		tm := NewTransactionManager(batch)

		// 添加表并执行操作
		tx, err := tm.AddTable(table)
		if err != nil {
			b.Fatalf("Failed to add table: %v", err)
		}

		// 插入测试数据
		insertFields := &map[string]any{
			"id":   i,
			"name": "benchmark",
			"value": i * 10,
		}
		_, err = tx.Insert(insertFields)
		if err != nil {
			b.Fatalf("Failed to insert data: %v", err)
		}

		// 提交事务
		err = tm.Commit()
		if err != nil {
			b.Fatalf("Failed to commit transaction: %v", err)
		}
	}
}

// BenchmarkTransactionUpdate 测试事务中更新操作的性能
func BenchmarkTransactionUpdate(b *testing.B) {
	// 测试目录设置
	testDir := filepath.Join(os.TempDir(), "sfsdb_benchmark_update")
	testDbPath := filepath.Join(testDir, "test_db")

	// 清理测试目录
	defer func() {
		os.RemoveAll(testDir)
	}()

	// 创建测试数据库
	testDb, err := storage.OpenDefaultDb(testDbPath)
	if err != nil {
		b.Fatalf("Failed to open test database: %v", err)
	}
	defer testDb.Close()

	// 创建表结构
	table, err := TableNew("benchmark_update_table")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// 设置表的字段
	fields := map[string]any{
		"id":   0,
		"name": "",
		"value": 0,
		"v":    "", // 版本号字段，用于乐观锁
	}
	err = table.SetFields(fields)
	if err != nil {
		b.Fatalf("Failed to set fields for table: %v", err)
	}

	// 初始化测试数据
	for i := 0; i < 100; i++ {
		batch := testDb.GetBatch()
		tm := NewTransactionManager(batch)
		tx, _ := tm.AddTable(table)
		insertFields := &map[string]any{
			"id":   i,
			"name": "initial",
			"value": i * 10,
			"v":    "1",
		}
		tx.Insert(insertFields)
		tm.Commit()
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 创建 batch
		batch := testDb.GetBatch()
		if batch == nil {
			b.Fatalf("Failed to create batch")
		}

		// 创建事务管理器
		tm := NewTransactionManager(batch)

		// 添加表并执行操作
		tx, err := tm.AddTable(table)
		if err != nil {
			b.Fatalf("Failed to add table: %v", err)
		}

		// 更新测试数据
		updateFields := &map[string]any{
			"id":   i % 100,
			"name": "updated",
			"value": i * 20,
			"v":    "1", // 版本号
		}
		err = tx.Update(updateFields)
		if err != nil {
			// 乐观锁冲突是正常的，继续下一次测试
			continue
		}

		// 提交事务
		err = tm.Commit()
		if err != nil {
			b.Fatalf("Failed to commit transaction: %v", err)
		}
	}
}

// BenchmarkTransactionRead 测试事务中读取操作的性能
func BenchmarkTransactionRead(b *testing.B) {
	// 测试目录设置
	testDir := filepath.Join(os.TempDir(), "sfsdb_benchmark_read")
	testDbPath := filepath.Join(testDir, "test_db")

	// 清理测试目录
	defer func() {
		os.RemoveAll(testDir)
	}()

	// 创建测试数据库
	testDb, err := storage.OpenDefaultDb(testDbPath)
	if err != nil {
		b.Fatalf("Failed to open test database: %v", err)
	}
	defer testDb.Close()

	// 创建表结构
	table, err := TableNew("benchmark_read_table")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// 设置表的字段
	fields := map[string]any{
		"id":   0,
		"name": "",
		"value": 0,
	}
	err = table.SetFields(fields)
	if err != nil {
		b.Fatalf("Failed to set fields for table: %v", err)
	}

	// 初始化测试数据
	for i := 0; i < 100; i++ {
		batch := testDb.GetBatch()
		tm := NewTransactionManager(batch)
		tx, _ := tm.AddTable(table)
		insertFields := &map[string]any{
			"id":   i,
			"name": "data",
			"value": i * 10,
		}
		tx.Insert(insertFields)
		tm.Commit()
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 创建 batch
		batch := testDb.GetBatch()
		if batch == nil {
			b.Fatalf("Failed to create batch")
		}

		// 创建事务管理器
		tm := NewTransactionManager(batch)

		// 添加表并执行操作
		tx, err := tm.AddTable(table)
		if err != nil {
			b.Fatalf("Failed to add table: %v", err)
		}

		// 读取测试数据
		readFields := &map[string]any{
			"id": i % 100,
		}
		_, err = tx.Read(readFields)
		if err != nil {
			b.Fatalf("Failed to read data: %v", err)
		}

		// 提交事务
		err = tm.Commit()
		if err != nil {
			b.Fatalf("Failed to commit transaction: %v", err)
		}
	}
}

// BenchmarkTransactionMixed 测试事务中混合操作的性能
func BenchmarkTransactionMixed(b *testing.B) {
	// 测试目录设置
	testDir := filepath.Join(os.TempDir(), "sfsdb_benchmark_mixed")
	testDbPath := filepath.Join(testDir, "test_db")

	// 清理测试目录
	defer func() {
		os.RemoveAll(testDir)
	}()

	// 创建测试数据库
	testDb, err := storage.OpenDefaultDb(testDbPath)
	if err != nil {
		b.Fatalf("Failed to open test database: %v", err)
	}
	defer testDb.Close()

	// 创建表结构
	table, err := TableNew("benchmark_mixed_table")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// 设置表的字段
	fields := map[string]any{
		"id":   0,
		"name": "",
		"value": 0,
	}
	err = table.SetFields(fields)
	if err != nil {
		b.Fatalf("Failed to set fields for table: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 创建 batch
		batch := testDb.GetBatch()
		if batch == nil {
			b.Fatalf("Failed to create batch")
		}

		// 创建事务管理器
		tm := NewTransactionManager(batch)

		// 添加表并执行操作
		tx, err := tm.AddTable(table)
		if err != nil {
			b.Fatalf("Failed to add table: %v", err)
		}

		// 插入测试数据
		insertFields := &map[string]any{
			"id":   i,
			"name": "mixed",
			"value": i * 10,
		}
		_, err = tx.Insert(insertFields)
		if err != nil {
			b.Fatalf("Failed to insert data: %v", err)
		}

		// 读取测试数据
		readFields := &map[string]any{
			"id": i,
		}
		_, err = tx.Read(readFields)
		if err != nil {
			b.Fatalf("Failed to read data: %v", err)
		}

		// 提交事务
		err = tm.Commit()
		if err != nil {
			b.Fatalf("Failed to commit transaction: %v", err)
		}
	}
}

// TestTransactionPerformanceComparison 测试并比较不同事务操作的性能
func TestTransactionPerformanceComparison(t *testing.T) {
	// 测试目录设置
	testDir := filepath.Join(os.TempDir(), "sfsdb_performance_comparison")
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
	table, err := TableNew("performance_comparison_table")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置表的字段
	fields := map[string]any{
		"id":   0,
		"name": "",
		"value": 0,
	}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields for table: %v", err)
	}

	// 测试参数
	operations := 10000
	concurrency := 100

	// 测试1: 基本事务性能
	t.Run("BasicTransaction", func(t *testing.T) {
		var wg sync.WaitGroup
		var mutex sync.Mutex
		errors := []error{}
		startTime := time.Now()

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()

				for j := 0; j < operations/concurrency; j++ {
					// 创建 batch
					batch := testDb.GetBatch()
					if batch == nil {
						mutex.Lock()
						errors = append(errors, ErrTableNotExist)
						mutex.Unlock()
						return
					}

					// 创建事务管理器
					tm := NewTransactionManager(batch)

					// 添加表并执行操作
					tx, err := tm.AddTable(table)
					if err != nil {
						mutex.Lock()
						errors = append(errors, err)
						mutex.Unlock()
						return
					}

					// 插入测试数据
					insertFields := &map[string]any{
						"id":   workerID*1000 + j + operations, // 避免ID冲突
						"name": "basic",
						"value": (workerID*1000 + j) * 10,
					}
					_, err = tx.Insert(insertFields)
					if err != nil {
						mutex.Lock()
						errors = append(errors, err)
						mutex.Unlock()
						return
					}

					// 提交事务
					err = tm.Commit()
					if err != nil {
						mutex.Lock()
						errors = append(errors, err)
						mutex.Unlock()
						return
					}
				}
			}(i)
		}

		wg.Wait()
		duration := time.Since(startTime)

		// 检查错误
		if len(errors) > 0 {
			t.Fatalf("Basic transaction test failed with %d errors: %v", len(errors), errors[0])
		}

		// 计算性能指标
		opsPerSecond := float64(operations) / duration.Seconds()

		t.Logf("Basic transaction test completed successfully!")
		t.Logf("Operations: %d", operations)
		t.Logf("Concurrency: %d", concurrency)
		t.Logf("Duration: %v", duration)
		t.Logf("Operations per second: %.2f", opsPerSecond)
	})

	// 测试2: 事务中包含读取操作
	t.Run("TransactionWithRead", func(t *testing.T) {
		// 先插入一些测试数据
		for i := 0; i < 100; i++ {
			batch := testDb.GetBatch()
			tm := NewTransactionManager(batch)
			tx, _ := tm.AddTable(table)
			insertFields := &map[string]any{
				"id":   i + operations*2, // 避免ID冲突
				"name": "read_test",
				"value": i * 10,
			}
			tx.Insert(insertFields)
			tm.Commit()
		}

		var wg sync.WaitGroup
		var mutex sync.Mutex
		errors := []error{}
		startTime := time.Now()

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()

				for j := 0; j < operations/concurrency; j++ {
					// 创建 batch
					batch := testDb.GetBatch()
					if batch == nil {
						mutex.Lock()
						errors = append(errors, ErrTableNotExist)
						mutex.Unlock()
						return
					}

					// 创建事务管理器
					tm := NewTransactionManager(batch)

					// 添加表并执行操作
					tx, err := tm.AddTable(table)
					if err != nil {
						mutex.Lock()
						errors = append(errors, err)
						mutex.Unlock()
						return
					}

					// 读取测试数据
					readFields := &map[string]any{
						"id": (workerID*1000 + j) % 100 + operations*2,
					}
					_, err = tx.Read(readFields)
					if err != nil {
						mutex.Lock()
						errors = append(errors, err)
						mutex.Unlock()
						return
					}

					// 提交事务
					err = tm.Commit()
					if err != nil {
						mutex.Lock()
						errors = append(errors, err)
						mutex.Unlock()
						return
					}
				}
			}(i)
		}

		wg.Wait()
		duration := time.Since(startTime)

		// 检查错误
		if len(errors) > 0 {
			t.Fatalf("Transaction with read test failed with %d errors: %v", len(errors), errors[0])
		}

		// 计算性能指标
		opsPerSecond := float64(operations) / duration.Seconds()

		t.Logf("Transaction with read test completed successfully!")
		t.Logf("Operations: %d", operations)
		t.Logf("Concurrency: %d", concurrency)
		t.Logf("Duration: %v", duration)
		t.Logf("Operations per second: %.2f", opsPerSecond)
	})

	// 测试3: 批量事务操作
	t.Run("BatchTransaction", func(t *testing.T) {
		var wg sync.WaitGroup
		var mutex sync.Mutex
		errors := []error{}
		startTime := time.Now()
		batchSize := 10

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()

				for j := 0; j < (operations/concurrency)/batchSize; j++ {
					// 创建 batch
					batch := testDb.GetBatch()
					if batch == nil {
						mutex.Lock()
						errors = append(errors, ErrTableNotExist)
						mutex.Unlock()
						return
					}

					// 创建事务管理器
					tm := NewTransactionManager(batch)

					// 添加表并执行操作
					tx, err := tm.AddTable(table)
					if err != nil {
						mutex.Lock()
						errors = append(errors, err)
						mutex.Unlock()
						return
					}

					// 批量插入测试数据
					for k := 0; k < batchSize; k++ {
						insertFields := &map[string]any{
							"id":   workerID*1000 + j*batchSize + k + operations*3, // 避免ID冲突
							"name": "batch",
							"value": (workerID*1000 + j*batchSize + k) * 10,
						}
						_, err := tx.Insert(insertFields)
						if err != nil {
							mutex.Lock()
							errors = append(errors, err)
							mutex.Unlock()
							return
						}
					}

					// 提交事务
					err = tm.Commit()
					if err != nil {
						mutex.Lock()
						errors = append(errors, err)
						mutex.Unlock()
						return
					}
				}
			}(i)
		}

		wg.Wait()
		duration := time.Since(startTime)

		// 检查错误
		if len(errors) > 0 {
			t.Fatalf("Batch transaction test failed with %d errors: %v", len(errors), errors[0])
		}

		// 计算性能指标
		opsPerSecond := float64(operations) / duration.Seconds()

		t.Logf("Batch transaction test completed successfully!")
		t.Logf("Operations: %d", operations)
		t.Logf("Concurrency: %d", concurrency)
		t.Logf("Batch size: %d", batchSize)
		t.Logf("Duration: %v", duration)
		t.Logf("Operations per second: %.2f", opsPerSecond)
	})
}
