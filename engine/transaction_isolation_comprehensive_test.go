package engine

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestTransactionIsolationLevelsComprehensive 测试所有事务隔离级别
func TestTransactionIsolationLevelsComprehensive(t *testing.T) {
	// 准备测试表
	table, err := TableNew("test_isolation_levels")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{"id": 0, "name": "", "value": 0}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键
	pk, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		t.Fatalf("Failed to create primary key: %v", err)
	}
	pk.AddFields("id")
	table.CreateIndex(pk)

	// 测试数据
	testData := []map[string]any{
		{"id": 1, "name": "test1", "value": 100},
		{"id": 2, "name": "test2", "value": 200},
		{"id": 3, "name": "test3", "value": 300},
	}

	// 插入测试数据
	for _, data := range testData {
		_, err := table.Insert(&data)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// 测试不同隔离级别
	isolationLevels := []string{
		ReadUncommitted,
		ReadCommitted,
		RepeatableRead,
		Serializable,
	}

	for _, level := range isolationLevels {
		t.Run(level, func(t *testing.T) {
			// 测试基本功能
			t.Run("BasicFunctionality", func(t *testing.T) {
				testIsolationLevelBasic(t, table, level)
			})

			// 测试并发场景
			t.Run("ConcurrentAccess", func(t *testing.T) {
				testIsolationLevelConcurrent(t, table, level)
			})

			// 测试边界情况
			t.Run("EdgeCases", func(t *testing.T) {
				testIsolationLevelEdgeCases(t, table, level)
			})
		})
	}
}

// testIsolationLevelBasic 测试隔离级别的基本功能
func testIsolationLevelBasic(t *testing.T, table *Table, isolationLevel string) {
	// 准备事务选项
	options := DefaultTransactionOptions()
	options.IsolationLevel = isolationLevel

	// 开始事务
	tx, err := table.BeginWithOptions(options)
	if err != nil {
		t.Fatalf("Failed to begin transaction with %s isolation: %v", isolationLevel, err)
	}

	// 测试读取
	readFields := map[string]any{"id": 1}
	records, err := tx.Search(&readFields)
	if err != nil {
		t.Fatalf("Failed to search in transaction: %v", err)
	}

	// 验证读取结果
	resultRecords := records.GetRecords(true)
	if len(resultRecords) == 0 {
		t.Fatalf("No records found for id=1")
	}

	// 测试更新
	updateFields := map[string]any{"id": 1, "value": 150}
	err = tx.Update(&updateFields)
	if err != nil {
		t.Fatalf("Failed to update in transaction: %v", err)
	}

	// 测试插入
	insertFields := map[string]any{"id": 4, "name": "test4", "value": 400}
	_, err = tx.Insert(&insertFields)
	if err != nil {
		t.Fatalf("Failed to insert in transaction: %v", err)
	}

	// 测试删除
	deleteFields := map[string]any{"id": 3}
	err = tx.Delete(&deleteFields)
	if err != nil {
		t.Fatalf("Failed to delete in transaction: %v", err)
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// 验证事务结果
	// 验证更新
	verifyFields := map[string]any{"id": 1}
	verifyIter, _ := table.Search(&verifyFields)
	verifyRecords := verifyIter.GetRecords(true)
	if verifyRecords[0]["value"] != 150 {
		t.Errorf("Update failed, expected value=150, got %v", verifyRecords[0]["value"])
	}

	// 验证插入
	insertVerifyFields := map[string]any{"id": 4}
	insertVerifyIter, _ := table.Search(&insertVerifyFields)
	insertVerifyRecords := insertVerifyIter.GetRecords(true)
	if len(insertVerifyRecords) == 0 {
		t.Error("Insert failed, record not found")
	}

	// 验证删除
	deleteVerifyFields := map[string]any{"id": 3}
	deleteVerifyIter, _ := table.Search(&deleteVerifyFields)
	deleteVerifyRecords := deleteVerifyIter.GetRecords(true)
	if len(deleteVerifyRecords) > 0 {
		t.Error("Delete failed, record still exists")
	}
}

// testIsolationLevelConcurrent 测试隔离级别的并发场景
func testIsolationLevelConcurrent(t *testing.T, table *Table, isolationLevel string) {
	// 重置测试数据
	resetTestData(t, table)

	// 准备事务选项
	options := DefaultTransactionOptions()
	options.IsolationLevel = isolationLevel

	// 测试并发读取
	t.Run("ConcurrentReads", func(t *testing.T) {
		var wg sync.WaitGroup
		readCount := 10

		wg.Add(readCount)
		for i := 0; i < readCount; i++ {
			go func(id int) {
				defer wg.Done()

				// 开始事务
				tx, err := table.BeginWithOptions(options)
				if err != nil {
					t.Errorf("Failed to begin transaction: %v", err)
					return
				}

				// 读取数据
				readFields := map[string]any{"id": 1}
				records, err := tx.Search(&readFields)
				if err != nil {
					t.Errorf("Failed to search: %v", err)
					return
				}

				// 验证读取结果
				resultRecords := records.GetRecords(true)
				if len(resultRecords) == 0 {
					t.Error("No records found")
					return
				}

				// 提交事务
				err = tx.Commit()
				if err != nil {
					t.Errorf("Failed to commit transaction: %v", err)
					return
				}
			}(i)
		}

		wg.Wait()
	})

	// 测试读写并发
	t.Run("ReadWriteConcurrency", func(t *testing.T) {
		var wg sync.WaitGroup

		// 写入事务
		wg.Add(1)
		go func() {
			defer wg.Done()

			// 开始事务
			tx, err := table.BeginWithOptions(options)
			if err != nil {
				t.Errorf("Failed to begin write transaction: %v", err)
				return
			}

			// 更新数据
			updateFields := map[string]any{"id": 1, "value": 999}
			err = tx.Update(&updateFields)
			if err != nil {
				t.Errorf("Failed to update: %v", err)
				return
			}

			// 模拟长时间操作
			time.Sleep(100 * time.Millisecond)

			// 提交事务
			err = tx.Commit()
			if err != nil {
				t.Errorf("Failed to commit write transaction: %v", err)
				return
			}
		}()

		// 读取事务
		wg.Add(1)
		go func() {
			defer wg.Done()

			// 稍微延迟，确保写入事务先开始
			time.Sleep(50 * time.Millisecond)

			// 开始事务
			tx, err := table.BeginWithOptions(options)
			if err != nil {
				t.Errorf("Failed to begin read transaction: %v", err)
				return
			}

			// 读取数据
			readFields := map[string]any{"id": 1}
			records, err := tx.Search(&readFields)
			if err != nil {
				t.Errorf("Failed to search: %v", err)
				return
			}

			// 验证读取结果
			resultRecords := records.GetRecords(true)
			if len(resultRecords) == 0 {
				t.Error("No records found")
				return
			}

			// 提交事务
			err = tx.Commit()
			if err != nil {
				t.Errorf("Failed to commit read transaction: %v", err)
				return
			}
		}()

		wg.Wait()
	})
}

// testIsolationLevelEdgeCases 测试隔离级别的边界情况
func testIsolationLevelEdgeCases(t *testing.T, table *Table, isolationLevel string) {
	// 重置测试数据
	resetTestData(t, table)

	// 准备事务选项
	options := DefaultTransactionOptions()
	options.IsolationLevel = isolationLevel

	// 测试空事务
	t.Run("EmptyTransaction", func(t *testing.T) {
		// 开始事务
		tx, err := table.BeginWithOptions(options)
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}

		// 直接提交
		err = tx.Commit()
		if err != nil {
			t.Fatalf("Failed to commit empty transaction: %v", err)
		}
	})

	// 测试只读事务
	t.Run("ReadOnlyTransaction", func(t *testing.T) {
		// 开始事务
		tx, err := table.BeginWithOptions(options)
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}

		// 读取数据
		readFields := map[string]any{"id": 1}
		records, err := tx.Search(&readFields)
		if err != nil {
			t.Fatalf("Failed to search: %v", err)
		}

		// 验证读取结果
		resultRecords := records.GetRecords(true)
		if len(resultRecords) == 0 {
			t.Fatalf("No records found")
		}

		// 提交事务
		err = tx.Commit()
		if err != nil {
			t.Fatalf("Failed to commit read-only transaction: %v", err)
		}
	})

	// 测试事务回滚
	t.Run("TransactionRollback", func(t *testing.T) {
		// 开始事务
		tx, err := table.BeginWithOptions(options)
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}

		// 更新数据
		updateFields := map[string]any{"id": 1, "value": 500}
		err = tx.Update(&updateFields)
		if err != nil {
			t.Fatalf("Failed to update: %v", err)
		}

		// 回滚事务
		err = tx.Rollback()
		if err != nil {
			t.Fatalf("Failed to rollback transaction: %v", err)
		}

		// 验证数据未被修改
		verifyFields := map[string]any{"id": 1}
		verifyIter, _ := table.Search(&verifyFields)
		verifyRecords := verifyIter.GetRecords(true)
		if verifyRecords[0]["value"] != 100 {
			t.Errorf("Rollback failed, expected value=100, got %v", verifyRecords[0]["value"])
		}
	})
}

// TestTransactionIsolationLevelPerformance 测试事务隔离级别的性能
func TestTransactionIsolationLevelPerformance(t *testing.T) {
	// 准备测试表
	table, err := TableNew("test_isolation_performance")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{"id": 0, "name": "", "value": 0}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键
	pk, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		t.Fatalf("Failed to create primary key: %v", err)
	}
	pk.AddFields("id")
	table.CreateIndex(pk)

	// 测试不同隔离级别的性能
	isolationLevels := []string{
		ReadUncommitted,
		ReadCommitted,
		RepeatableRead,
		Serializable,
	}

	for _, level := range isolationLevels {
		t.Run(fmt.Sprintf("%s_Performance", level), func(t *testing.T) {
			// 准备事务选项
			options := DefaultTransactionOptions()
			options.IsolationLevel = level

			// 测试参数
			transactionCount := 100
			operationsPerTransaction := 5

			// 开始计时
			startTime := time.Now()

			// 执行事务
			for i := 0; i < transactionCount; i++ {
				// 开始事务
				tx, err := table.BeginWithOptions(options)
				if err != nil {
					t.Errorf("Failed to begin transaction: %v", err)
					continue
				}

				// 执行多个操作
				for j := 0; j < operationsPerTransaction; j++ {
					// 插入数据
					insertFields := map[string]any{
						"id":    i*100 + j,
						"name":  fmt.Sprintf("test_%d_%d", i, j),
						"value": i*100 + j,
					}
					_, err := tx.Insert(&insertFields)
					if err != nil {
						t.Errorf("Failed to insert: %v", err)
						break
					}
				}

				// 提交事务
				err = tx.Commit()
				if err != nil {
					t.Errorf("Failed to commit transaction: %v", err)
					continue
				}
			}

			// 计算性能
			duration := time.Since(startTime)
			totalOperations := transactionCount * operationsPerTransaction
			operationsPerSecond := float64(totalOperations) / duration.Seconds()

			t.Logf("%s isolation level: %d transactions, %d operations, took %v, %.2f operations/second",
				level, transactionCount, totalOperations, duration, operationsPerSecond)
		})
	}
}

// TestTransactionIsolationLevelConsistency 测试事务隔离级别的一致性
func TestTransactionIsolationLevelConsistency(t *testing.T) {
	// 准备测试表
	table, err := TableNew("test_isolation_consistency")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{"id": 0, "name": "", "balance": 1000}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键
	pk, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		t.Fatalf("Failed to create primary key: %v", err)
	}
	pk.AddFields("id")
	table.CreateIndex(pk)

	// 插入测试数据
	testData := map[string]any{"id": 1, "name": "account", "balance": 1000}
	_, err = table.Insert(&testData)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// 测试不同隔离级别
	isolationLevels := []string{
		ReadUncommitted,
		ReadCommitted,
		RepeatableRead,
		Serializable,
	}

	for _, level := range isolationLevels {
		t.Run(fmt.Sprintf("%s_Consistency", level), func(t *testing.T) {
			// 重置数据
			resetData := map[string]any{"id": 1, "balance": 1000}
			err := table.Update(&resetData)
			if err != nil {
				t.Fatalf("Failed to reset data: %v", err)
			}

			// 准备事务选项
			options := DefaultTransactionOptions()
			options.IsolationLevel = level

			// 并发执行多个更新操作
			var wg sync.WaitGroup
			transactionCount := 10

			for i := 0; i < transactionCount; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()

					// 开始事务
					tx, err := table.BeginWithOptions(options)
					if err != nil {
						t.Errorf("Failed to begin transaction: %v", err)
						return
					}

					// 读取当前余额
					readFields := map[string]any{"id": 1}
					records, err := tx.Search(&readFields)
					if err != nil {
						t.Errorf("Failed to search: %v", err)
						return
					}

					// 获取当前余额
					resultRecords := records.GetRecords(true)
					if len(resultRecords) == 0 {
						t.Error("No records found")
						return
					}

					currentBalance := resultRecords[0]["balance"].(int)

					// 模拟业务逻辑：扣除10元
					newBalance := currentBalance - 10

					// 更新余额
					updateFields := map[string]any{"id": 1, "balance": newBalance}
					err = tx.Update(&updateFields)
					if err != nil {
						t.Errorf("Failed to update: %v", err)
						return
					}

					// 提交事务
					err = tx.Commit()
					if err != nil {
						t.Errorf("Failed to commit transaction: %v", err)
						return
					}
				}(i)
			}

			wg.Wait()

			// 验证最终余额
			verifyFields := map[string]any{"id": 1}
			verifyIter, _ := table.Search(&verifyFields)
			verifyRecords := verifyIter.GetRecords(true)
			finalBalance := verifyRecords[0]["balance"].(int)
			expectedBalance := 1000 - transactionCount*10

			if finalBalance != expectedBalance {
				t.Errorf("Consistency test failed for %s isolation level: expected balance=%d, got %d",
					level, expectedBalance, finalBalance)
			} else {
				t.Logf("Consistency test passed for %s isolation level: balance=%d", level, finalBalance)
			}
		})
	}
}

// resetTestData 重置测试数据
func resetTestData(t *testing.T, table *Table) {
	// 删除现有数据
	deleteFields := map[string]any{"id": 1}
	err := table.Delete(&deleteFields)
	if err != nil {
		t.Logf("Failed to delete existing data: %v", err)
	}

	deleteFields["id"] = 2
	err = table.Delete(&deleteFields)
	if err != nil {
		t.Logf("Failed to delete existing data: %v", err)
	}

	deleteFields["id"] = 3
	err = table.Delete(&deleteFields)
	if err != nil {
		t.Logf("Failed to delete existing data: %v", err)
	}

	// 插入测试数据
	testData := []map[string]any{
		{"id": 1, "name": "test1", "value": 100},
		{"id": 2, "name": "test2", "value": 200},
		{"id": 3, "name": "test3", "value": 300},
	}

	for _, data := range testData {
		_, err := table.Insert(&data)
		if err != nil {
			t.Logf("Failed to insert test data: %v", err)
		}
	}
}
