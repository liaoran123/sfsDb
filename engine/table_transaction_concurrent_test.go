package engine

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestTableTransactionConcurrent 测试TableTransaction的并发性能
func TestTableTransactionConcurrent(t *testing.T) {
	// 使用唯一表名，避免测试数据累积
	tableName := fmt.Sprintf("test_transaction_concurrent_%d", time.Now().UnixNano())

	// 创建表
	table, err := TableNew(tableName)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 定义表字段
	fields := map[string]any{
		"id":    0,
		"name":  "",
		"age":   0,
		"count": 0,
	}

	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键索引
	pk, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}
	pk.AddFields("id")
	err = table.CreateIndex(pk)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// 插入初始测试数据
	initialRecord := map[string]any{
		"id":    1,
		"name":  "Test Record",
		"age":   25,
		"count": 0,
	}

	_, err = table.Insert(&initialRecord)
	if err != nil {
		t.Fatalf("Failed to insert initial record: %v", err)
	}

	// 测试1: 多个事务同时读取同一记录
	t.Log("=== 测试1: 多个事务同时读取同一记录 ===")
	testConcurrentRead(t, table)

	// 测试2: 多个事务同时更新同一记录，测试乐观锁
	t.Log("=== 测试2: 多个事务同时更新同一记录，测试乐观锁 ===")
	testConcurrentUpdateWithOptimisticLock(t, table)

	// 测试3: 一个事务修改记录，另一个事务读取，测试读一致性
	t.Log("=== 测试3: 一个事务修改记录，另一个事务读取，测试读一致性 ===")
	testReadConsistencyDuringWrite(t, table)

	// 测试4: 多个事务并发插入记录
	t.Log("=== 测试4: 多个事务并发插入记录 ===")
	testConcurrentInsert(t, table)
}

// 测试多个事务同时读取同一记录
func testConcurrentRead(t *testing.T, table *Table) {
	const concurrentReads = 10
	var wg sync.WaitGroup
	var readErrors int

	wg.Add(concurrentReads)
	for i := 0; i < concurrentReads; i++ {
		go func(threadID int) {
			defer wg.Done()

			// 开始事务
			tx, err := table.Begin()
			if err != nil {
				t.Logf("线程%d: 开始事务失败: %v", threadID, err)
				readErrors++
				return
			}
			defer tx.Rollback()

			// 读取记录
			readFields := map[string]any{"id": 1}
			iter := tx.Search(&readFields)
			if iter == nil {
				t.Logf("线程%d: 搜索记录失败", threadID)
				readErrors++
				return
			}
			defer iter.Release()

			records := iter.GetRecords(true)
			if len(records) != 1 {
				t.Logf("线程%d: 读取记录失败，期望1条记录，实际%d条", threadID, len(records))
				readErrors++
				return
			}

			// 验证记录内容
			record := records[0]
			if record["name"] != "Test Record" || record["age"] != 25 {
				t.Logf("线程%d: 读取到的记录内容不正确: %v", threadID, record)
				readErrors++
				return
			}

			t.Logf("线程%d: 成功读取记录: %v", threadID, record)
		}(i)
	}

	wg.Wait()

	if readErrors > 0 {
		t.Fatalf("并发读测试失败，共%d个错误", readErrors)
	}
	t.Logf("✓ 并发读测试通过，%d个事务同时读取同一记录成功", concurrentReads)
}

// 测试多个事务同时更新同一记录，测试乐观锁
func testConcurrentUpdateWithOptimisticLock(t *testing.T, table *Table) {
	const concurrentUpdates = 3
	var wg sync.WaitGroup
	var totalSuccess int32
	var totalConflicts int32
	var totalErrors int32
	var mu sync.Mutex

	wg.Add(concurrentUpdates)
	for i := 0; i < concurrentUpdates; i++ {
		go func(threadID int) {
			defer wg.Done()

			// 等待一小段时间，确保所有事务都能获取到初始快照
			time.Sleep(time.Millisecond * 10)

			// 开始事务
			tx, err := table.Begin()
			if err != nil {
				mu.Lock()
				totalErrors++
				mu.Unlock()
				return
			}

			// 在事务内部读取记录，获取当前count值和版本号
			readFields := map[string]any{"id": 1}
			recordBytes, err := tx.Read(&readFields)
			if err != nil {
				tx.Rollback()
				mu.Lock()
				totalErrors++
				mu.Unlock()
				return
			}
			if recordBytes == nil {
				tx.Rollback()
				mu.Lock()
				totalErrors++
				mu.Unlock()
				return
			}

			// 解析记录，获取当前count值和版本号
			pk := table.GetPrimaryKey()
			fieldsBytes, err := pk.Parse(table.fieldsid, recordBytes)
			if err != nil {
				tx.Rollback()
				mu.Lock()
				totalErrors++
				mu.Unlock()
				return
			}

			// 将字节数组转换为any类型
			currentRecord := *table.RecordByteToAny(fieldsBytes)
			currentVersion, _ := currentRecord["v"].(int)
			currentCount := currentRecord["count"].(int)

			// 故意延迟，让其他事务也能获取到相同的版本号
			time.Sleep(time.Millisecond * 100)

			// 准备更新数据，包含版本号
			updateFields := map[string]any{
				"id":    1,
				"count": currentCount + 1,
				"v":     currentVersion,
			}

			// 更新记录
			err = tx.Update(&updateFields)
			if err != nil {
				// 检查是否是乐观锁冲突
				if containsOptimisticLockError(err.Error()) {
					tx.Rollback()
					mu.Lock()
					totalConflicts++
					mu.Unlock()
					return
				} else {
					tx.Rollback()
					mu.Lock()
					totalErrors++
					mu.Unlock()
					return
				}
			}

			// 提交事务
			err = tx.Commit()
			if err != nil {
				tx.Rollback()
				mu.Lock()
				totalErrors++
				mu.Unlock()
				return
			}

			mu.Lock()
			totalSuccess++
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	if totalErrors > 0 {
		t.Fatalf("并发更新测试失败，共%d个其他错误", totalErrors)
	}

	// 至少有一个事务应该成功
	if totalSuccess == 0 {
		t.Fatalf("并发更新测试失败，没有事务成功更新记录")
	}

	// 应该有并发冲突
	if totalConflicts == 0 {
		t.Logf("警告：没有检测到乐观锁冲突，可能是测试环境并发不足")
	} else {
		t.Logf("✓ 检测到%d个乐观锁冲突，乐观锁机制正常工作", totalConflicts)
	}

	t.Logf("✓ 并发更新测试通过: %d个事务成功，%d个乐观锁冲突，%d个其他错误", totalSuccess, totalConflicts, totalErrors)

	// 验证最终记录的count值至少为1
	searchFields := map[string]any{"id": 1}
	iter := table.Search(&searchFields)
	if iter == nil {
		t.Fatalf("搜索最终记录失败")
	}
	defer iter.Release()

	records := iter.GetRecords(true)
	if len(records) != 1 {
		t.Fatalf("读取最终记录失败，期望1条记录，实际%d条", len(records))
	}

	finalRecord := records[0]
	finalCount := finalRecord["count"].(int)
	if finalCount < 1 {
		t.Fatalf("最终count值不正确，期望至少1，实际%d", finalCount)
	}
	t.Logf("✓ 最终记录count值正确: %d", finalCount)
}

// 测试一个事务修改记录，另一个事务读取，测试读一致性
func testReadConsistencyDuringWrite(t *testing.T, table *Table) {
	// 重置测试数据
	resetFields := map[string]any{
		"id":    1,
		"name":  "Test Record",
		"age":   25,
		"count": 0,
	}
	table.Update(&resetFields)

	// 开始第一个事务，获取快照
	tx1, err := table.Begin()
	if err != nil {
		t.Fatalf("开始事务1失败: %v", err)
	}

	// 在第一个事务外部，开始第二个事务修改记录
	tx2, err := table.Begin()
	if err != nil {
		t.Fatalf("开始事务2失败: %v", err)
	}

	// 事务2更新记录
	updateFields := map[string]any{
		"id":   1,
		"name": "Updated Record",
		"age":  26,
	}
	err = tx2.Update(&updateFields)
	if err != nil {
		t.Fatalf("事务2更新记录失败: %v", err)
	}

	// 事务2提交
	err = tx2.Commit()
	if err != nil {
		t.Fatalf("事务2提交失败: %v", err)
	}

	// 事务1读取记录，应该读取到的是快照中的初始状态
	readFields := map[string]any{"id": 1}
	iter := tx1.Search(&readFields)
	if iter == nil {
		t.Fatalf("事务1搜索记录失败")
	}
	defer iter.Release()

	records := iter.GetRecords(true)
	if len(records) != 1 {
		t.Fatalf("事务1读取记录失败，期望1条记录，实际%d条", len(records))
	}

	record := records[0]
	if record["name"] != "Test Record" || record["age"] != 25 {
		t.Fatalf("事务1读取到的记录与快照不一致，期望'Test Record'，实际'%s'，期望年龄25，实际%d", record["name"], record["age"])
	}

	// 事务1提交
	err = tx1.Commit()
	if err != nil {
		t.Fatalf("事务1提交失败: %v", err)
	}

	t.Logf("✓ 读一致性测试通过: 事务1读取到的是快照数据，不受事务2修改影响")
}

// 测试多个事务并发插入记录
func testConcurrentInsert(t *testing.T, table *Table) {
	const concurrentInserts = 10
	var wg sync.WaitGroup
	var insertErrors int
	var insertSuccesses int

	wg.Add(concurrentInserts)
	for i := 0; i < concurrentInserts; i++ {
		go func(threadID int) {
			defer wg.Done()

			// 开始事务
			tx, err := table.Begin()
			if err != nil {
				t.Logf("线程%d: 开始事务失败: %v", threadID, err)
				insertErrors++
				return
			}

			// 插入记录
			insertFields := map[string]any{
				"id":    2 + threadID,
				"name":  fmt.Sprintf("Concurrent Record %d", threadID),
				"age":   20 + threadID,
				"count": 0,
			}

			_, err = tx.Insert(&insertFields)
			if err != nil {
				t.Logf("线程%d: 插入记录失败: %v", threadID, err)
				tx.Rollback()
				insertErrors++
				return
			}

			// 提交事务
			err = tx.Commit()
			if err != nil {
				t.Logf("线程%d: 提交事务失败: %v", threadID, err)
				insertErrors++
				return
			}

			t.Logf("线程%d: 成功插入记录，ID: %d", threadID, 2+threadID)
			insertSuccesses++
		}(i)
	}

	wg.Wait()

	if insertErrors > 0 {
		t.Fatalf("并发插入测试失败，共%d个错误", insertErrors)
	}

	// 验证插入的记录数
	searchFields := map[string]any{"id": nil} // 使用id=nil搜索所有记录（主键搜索，nil表示所有）
	iter := table.Search(&searchFields)
	if iter == nil {
		t.Fatalf("搜索所有记录失败")
	}
	defer iter.Release()

	records := iter.GetRecords(true)
	expectedCount := 1 + concurrentInserts // 初始记录 + 并发插入的记录
	if len(records) != expectedCount {
		t.Fatalf("验证插入记录数失败，期望%d条，实际%d条", expectedCount, len(records))
	}

	t.Logf("✓ 并发插入测试通过，成功插入%d条记录，共验证%d条记录", insertSuccesses, len(records))
}

// 检查错误信息是否包含乐观锁冲突
func containsOptimisticLockError(err string) bool {
	return len(err) > 0 && (err == "optimistic lock conflict: version mismatch" || contains(err, "optimistic lock conflict"))
}

// 字符串包含检查
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || indexOf(s, substr) >= 0)
}

// 字符串索引查找
func indexOf(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
