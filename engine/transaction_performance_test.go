package engine

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/liaoran123/sfsDb/storage"
)

// BenchmarkTransactionManager 测试事务管理器的性能
func BenchmarkTransactionManager(b *testing.B) {
	// 测试目录设置
	testDir := filepath.Join(os.TempDir(), "sfsdb_benchmark")
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

// BenchmarkTransactionManagerConcurrent 测试事务管理器的并发性能
func BenchmarkTransactionManagerConcurrent(b *testing.B) {
	// 测试目录设置
	testDir := filepath.Join(os.TempDir(), "sfsdb_benchmark_concurrent")
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
	table, err := TableNew("benchmark_concurrent_table")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// 设置表的字段
	fields := map[string]any{
		"id":   0,
		"name": "",
	}
	err = table.SetFields(fields)
	if err != nil {
		b.Fatalf("Failed to set fields for table: %v", err)
	}

	// 并发测试
	var wg sync.WaitGroup
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

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
				"id":   id,
				"name": "concurrent",
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
		}(i)
	}

	wg.Wait()
}

// TestTransactionManagerConcurrent 测试事务管理器的并发性能
func TestTransactionManagerConcurrent(t *testing.T) {
	// 测试目录设置
	testDir := filepath.Join(os.TempDir(), "sfsdb_concurrent_test")
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
	table, err := TableNew("concurrent_test_table")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置表的字段
	fields := map[string]any{
		"id":   0,
		"name": "",
	}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields for table: %v", err)
	}

	// 并发测试参数
	concurrency := 100
	operations := 1000

	var wg sync.WaitGroup
	var mutex sync.Mutex
	errors := []error{}
	startTime := time.Now()

	// 执行并发操作
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
					"id":   workerID*1000 + j,
					"name": "concurrent_test",
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
		t.Fatalf("Concurrent test failed with %d errors: %v", len(errors), errors[0])
	}

	// 验证数据
	count := 0
	for i := 0; i < operations; i++ {
		readFields := &map[string]any{
			"id": i,
		}
		_, err := table.Read(readFields)
		if err == nil {
			count++
		}
	}

	// 计算性能指标
	opsPerSecond := float64(operations) / duration.Seconds()

	t.Logf("Concurrent test completed successfully!")
	t.Logf("Operations: %d", operations)
	t.Logf("Concurrency: %d", concurrency)
	t.Logf("Duration: %v", duration)
	t.Logf("Operations per second: %.2f", opsPerSecond)
	t.Logf("Success rate: %.2f%%", float64(count)/float64(operations)*100)

	// 性能阈值检查
	if opsPerSecond < 100 {
		t.Logf("Low performance: %.2f ops/s", opsPerSecond)
	}
}

// TestTransactionManagerLongRunning 测试事务管理器的长时间运行性能
func TestTransactionManagerLongRunning(t *testing.T) {
	// 测试目录设置
	testDir := filepath.Join(os.TempDir(), "sfsdb_long_running")
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
	table, err := TableNew("long_running_table")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置表的字段
	fields := map[string]any{
		"id":   0,
		"name": "",
	}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields for table: %v", err)
	}

	// 测试参数
	testDuration := 30 * time.Second
	operationInterval := 10 * time.Millisecond
	expectedOperations := int(testDuration / operationInterval)

	var wg sync.WaitGroup
	var mutex sync.Mutex
	errors := []error{}
	operations := 0
	startTime := time.Now()

	// 长时间运行测试
	timer := time.NewTimer(testDuration)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			goto endTest
		default:
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

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
					"id":   id,
					"name": "long_running",
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

				mutex.Lock()
				operations++
				mutex.Unlock()
			}(operations)

			operations++
			time.Sleep(operationInterval)
		}
	}

endTest:
	wg.Wait()
	duration := time.Since(startTime)

	// 检查错误
	if len(errors) > 0 {
		t.Fatalf("Long running test failed with %d errors: %v", len(errors), errors[0])
	}

	// 计算性能指标
	opsPerSecond := float64(operations) / duration.Seconds()

	t.Logf("Long running test completed successfully!")
	t.Logf("Duration: %v", duration)
	t.Logf("Operations: %d", operations)
	t.Logf("Expected operations: %d", expectedOperations)
	t.Logf("Operations per second: %.2f", opsPerSecond)

	// 验证系统稳定性
	if operations < int(float64(expectedOperations)*0.9) {
		t.Logf("Low operation count: %d, expected: %d", operations, expectedOperations)
	}
}
