package engine

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/liaoran123/sfsDb/storage"
)

// TestTransactionWithUpdatePerformance 测试包含更新操作的事务性能
func TestTransactionWithUpdatePerformance(t *testing.T) {
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
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDb.Close()

	// 创建表结构
	table, err := TableNew("benchmark_update_table")
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

	// 初始化测试数据
	for i := 0; i < 1000; i++ {
		batch := testDb.GetBatch()
		tm := NewTransactionManager(batch)
		tx, _ := tm.AddTable(table)
		insertFields := &map[string]any{
			"id":   i,
			"name": "initial",
			"value": i * 10,
		}
		tx.Insert(insertFields)
		tm.Commit()
	}

	// 测试参数
	operations := 10000
	concurrency := 100

	var wg sync.WaitGroup
	var mutex sync.Mutex
	errors := []error{}
	successCount := 0
	startTime := time.Now()

	// 执行并发更新操作
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

				// 更新测试数据
			updateFields := &map[string]any{
				"id":   (workerID*100 + j) % 1000,
				"name": "updated",
				"value": (workerID*100 + j) * 20,
			}
			err = tx.Update(updateFields)
			if err != nil {
				// 乐观锁冲突是正常的，继续下一次测试
				continue
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
				successCount++
				mutex.Unlock()
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	// 检查错误
	if len(errors) > 0 {
		t.Fatalf("Transaction with update test failed with %d errors: %v", len(errors), errors[0])
	}

	// 计算性能指标
	opsPerSecond := float64(successCount) / duration.Seconds()

	t.Logf("Transaction with update test completed successfully!")
	t.Logf("Total operations: %d", operations)
	t.Logf("Success operations: %d", successCount)
	t.Logf("Concurrency: %d", concurrency)
	t.Logf("Duration: %v", duration)
	t.Logf("Operations per second: %.2f", opsPerSecond)
}

// TestTransactionWithDeletePerformance 测试包含删除操作的事务性能
func TestTransactionWithDeletePerformance(t *testing.T) {
	// 测试目录设置
	testDir := filepath.Join(os.TempDir(), "sfsdb_benchmark_delete")
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
	table, err := TableNew("benchmark_delete_table")
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
	initialRecords := 10000
	deleteOperations := 5000
	concurrency := 100

	// 初始化测试数据
	for i := 0; i < initialRecords; i++ {
		batch := testDb.GetBatch()
		tm := NewTransactionManager(batch)
		tx, _ := tm.AddTable(table)
		insertFields := &map[string]any{
			"id":   i,
			"name": "initial",
			"value": i * 10,
		}
		tx.Insert(insertFields)
		tm.Commit()
	}

	var wg sync.WaitGroup
	var mutex sync.Mutex
	errors := []error{}
	successCount := 0
	startTime := time.Now()

	// 执行并发删除操作
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < deleteOperations/concurrency; j++ {
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

				// 删除测试数据
				deleteFields := &map[string]any{
					"id": (workerID*50 + j) % initialRecords,
				}
				err = tx.Delete(deleteFields)
				if err != nil {
					// 记录不存在是正常的，继续下一次测试
					continue
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
				successCount++
				mutex.Unlock()
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	// 检查错误
	if len(errors) > 0 {
		t.Fatalf("Transaction with delete test failed with %d errors: %v", len(errors), errors[0])
	}

	// 计算性能指标
	opsPerSecond := float64(successCount) / duration.Seconds()

	t.Logf("Transaction with delete test completed successfully!")
	t.Logf("Initial records: %d", initialRecords)
	t.Logf("Delete operations: %d", deleteOperations)
	t.Logf("Success operations: %d", successCount)
	t.Logf("Concurrency: %d", concurrency)
	t.Logf("Duration: %v", duration)
	t.Logf("Operations per second: %.2f", opsPerSecond)
}

// TestTransactionMixedOperationsPerformance 测试包含混合操作的事务性能
func TestTransactionMixedOperationsPerformance(t *testing.T) {
	// 测试目录设置
	testDir := filepath.Join(os.TempDir(), "sfsdb_benchmark_mixed_operations")
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
	table, err := TableNew("benchmark_mixed_table")
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

	// 初始化测试数据
	for i := 0; i < 1000; i++ {
		batch := testDb.GetBatch()
		tm := NewTransactionManager(batch)
		tx, _ := tm.AddTable(table)
		insertFields := &map[string]any{
			"id":   i,
			"name": "initial",
			"value": i * 10,
		}
		tx.Insert(insertFields)
		tm.Commit()
	}

	// 测试参数
	operations := 10000
	concurrency := 100

	var wg sync.WaitGroup
	var mutex sync.Mutex
	errors := []error{}
	successCount := 0
	insertCount := 0
	updateCount := 0
	deleteCount := 0
	readCount := 0
	startTime := time.Now()

	// 执行并发混合操作
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

				// 根据操作类型执行不同操作
				opType := (workerID*100 + j) % 4
				success := false

				switch opType {
				case 0: // 插入操作
					insertFields := &map[string]any{
						"id":   1000 + workerID*100 + j,
						"name": "inserted",
						"value": (workerID*100 + j) * 10,
					}
					_, err := tx.Insert(insertFields)
					if err == nil {
						success = true
						mutex.Lock()
						insertCount++
						mutex.Unlock()
					}

				case 1: // 更新操作
					updateFields := &map[string]any{
						"id":   (workerID*100 + j) % 1000,
						"name": "updated",
						"value": (workerID*100 + j) * 20,
					}
					err := tx.Update(updateFields)
					if err == nil {
						success = true
						mutex.Lock()
						updateCount++
						mutex.Unlock()
					}

				case 2: // 删除操作
					deleteFields := &map[string]any{
						"id": (workerID*100 + j) % 1000,
					}
					err := tx.Delete(deleteFields)
					if err == nil {
						success = true
						mutex.Lock()
						deleteCount++
						mutex.Unlock()
					}

				case 3: // 读取操作
					readFields := &map[string]any{
						"id": (workerID*100 + j) % 1000,
					}
					_, err := tx.Read(readFields)
					if err == nil {
						success = true
						mutex.Lock()
						readCount++
						mutex.Unlock()
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

				if success {
					mutex.Lock()
					successCount++
					mutex.Unlock()
				}
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	// 检查错误
	if len(errors) > 0 {
		t.Fatalf("Transaction with mixed operations test failed with %d errors: %v", len(errors), errors[0])
	}

	// 计算性能指标
	opsPerSecond := float64(successCount) / duration.Seconds()

	t.Logf("Transaction with mixed operations test completed successfully!")
	t.Logf("Total operations: %d", operations)
	t.Logf("Success operations: %d", successCount)
	t.Logf("Insert operations: %d", insertCount)
	t.Logf("Update operations: %d", updateCount)
	t.Logf("Delete operations: %d", deleteCount)
	t.Logf("Read operations: %d", readCount)
	t.Logf("Concurrency: %d", concurrency)
	t.Logf("Duration: %v", duration)
	t.Logf("Operations per second: %.2f", opsPerSecond)
}
