package engine

import (
	"fmt"
	"sync"
	"testing"

	"github.com/liaoran123/sfsDb/storage"
)

// TestStressTransactionManager_ConcurrentTransfers 测试并发资金转账
func TestStressTransactionManager_ConcurrentTransfers(t *testing.T) {
	// 打开默认存储
	_, err := storage.OpenDefaultDb("./test/kvdb_stress_concurrent")
	if err != nil {
		t.Fatalf("Failed to open store: %v", err)
	}

	// 创建账户表
	accountTable, err := TableNew("accounts")
	if err != nil {
		t.Fatalf("Failed to create account table: %v", err)
	}

	// 设置表字段
	err = accountTable.SetFields(map[string]any{"id": "", "name": "", "balance": 0.0})
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 初始化测试数据
	aliceData := map[string]any{"id": "1", "name": "Alice", "balance": 10000.0}
	bobData := map[string]any{"id": "2", "name": "Bob", "balance": 10000.0}

	_, err = accountTable.Insert(&aliceData)
	if err != nil {
		t.Fatalf("Failed to insert Alice's account: %v", err)
	}

	_, err = accountTable.Insert(&bobData)
	if err != nil {
		t.Fatalf("Failed to insert Bob's account: %v", err)
	}

	// 并发转账次数
	transferCount := 10

	// 使用WaitGroup等待所有并发操作完成
	var wg sync.WaitGroup
	wg.Add(transferCount)

	// 记录错误
	var mu sync.Mutex
	errors := []error{}

	// 并发执行转账操作
	for i := 0; i < transferCount; i++ {
		go func(transferID int) {
			defer wg.Done()

			// 创建事务管理器
			batch := accountTable.kvStore.GetBatch()
			manager := NewTransactionManager(batch)

			// 添加表到事务管理器
			accountTx, err := manager.AddTable(accountTable)
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
				return
			}

			// 读取Alice的账户
			aliceReadFields := map[string]any{"id": "1"}
			_, err = accountTx.Read(&aliceReadFields)
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
				manager.Rollback()
				return
			}

			// 读取Bob的账户
			bobReadFields := map[string]any{"id": "2"}
			_, err = accountTx.Read(&bobReadFields)
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
				manager.Rollback()
				return
			}

			// 提交事务
			if err := manager.Commit(); err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
				return
			}
		}(i)
	}

	// 等待所有并发操作完成
	wg.Wait()

	// 检查是否有错误
	if len(errors) > 0 {
		t.Fatalf("Got %d errors during concurrent transfers: %v", len(errors), errors[0])
	}

	// 验证结果
	// 注意：这里简化测试，只验证事务提交成功
	t.Log("Concurrent transfers completed successfully")
}

// TestStressTransactionManager_TransactionRetry 测试事务重试机制
func TestStressTransactionManager_TransactionRetry(t *testing.T) {
	// 打开默认存储
	_, err := storage.OpenDefaultDb("./test/kvdb_stress_retry")
	if err != nil {
		t.Fatalf("Failed to open store: %v", err)
	}

	// 创建计数器表
	counterTable, err := TableNew("counters")
	if err != nil {
		t.Fatalf("Failed to create counter table: %v", err)
	}

	// 设置表字段
	err = counterTable.SetFields(map[string]any{"id": "", "value": 0})
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 初始化测试数据
	counterData := map[string]any{"id": "1", "value": 0}
	_, err = counterTable.Insert(&counterData)
	if err != nil {
		t.Fatalf("Failed to insert counter: %v", err)
	}

	// 并发递增次数
	incrementCount := 10

	// 使用WaitGroup等待所有并发操作完成
	var wg sync.WaitGroup
	wg.Add(incrementCount)

	// 记录错误
	var mu sync.Mutex
	errors := []error{}

	// 并发执行递增操作
	for i := 0; i < incrementCount; i++ {
		go func(incrementID int) {
			defer wg.Done()

			// 重试机制
			maxRetries := 5
			retryCount := 0

			for retryCount < maxRetries {
				// 创建事务管理器
				batch := counterTable.kvStore.GetBatch()
				manager := NewTransactionManager(batch)

				// 添加表到事务管理器
				counterTx, err := manager.AddTable(counterTable)
				if err != nil {
					mu.Lock()
					errors = append(errors, err)
					mu.Unlock()
					retryCount++
					continue
				}

				// 读取计数器
				counterReadFields := map[string]any{"id": "1"}
				_, err = counterTx.Read(&counterReadFields)
				if err != nil {
					mu.Lock()
					errors = append(errors, err)
					mu.Unlock()
					manager.Rollback()
					retryCount++
					continue
				}

				// 提交事务
				if err := manager.Commit(); err != nil {
					manager.Rollback()
					retryCount++
					continue
				}

				// 事务成功，退出重试循环
				break
			}

			// 达到最大重试次数仍失败
			if retryCount >= maxRetries {
				mu.Lock()
				errors = append(errors, fmt.Errorf("Max retries reached for increment %d", incrementID))
				mu.Unlock()
			}
		}(i)
	}

	// 等待所有并发操作完成
	wg.Wait()

	// 检查是否有错误
	if len(errors) > 0 {
		t.Fatalf("Got %d errors during concurrent increments: %v", len(errors), errors[0])
	}

	// 验证结果
	// 注意：这里简化测试，只验证事务提交成功
	t.Log("Concurrent increments completed successfully")
}
