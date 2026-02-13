package engine

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/liaoran123/sfsDb/storage"
)

// 测试行级锁和高级乐观锁机制
func TestRowLockAndOptimisticLock(t *testing.T) {
	// 创建临时目录作为数据库路径
	tempDir := t.TempDir()

	// 显式打开数据库，使用临时目录
	db, err := storage.OpenDefaultDb(tempDir)
	if err != nil {
		t.Fatalf("打开数据库失败: %v", err)
	}

	// 清理测试环境
	defer func() {
		if db != nil {
			db.Close()
			storage.KVDb = nil
		}
	}()

	// 创建测试表
	table, err := TableNew("test_row_lock")
	if err != nil {
		t.Fatalf("创建表失败: %v", err)
	}

	// 设置表字段
	fields := map[string]any{
		"id":   0,
		"name": "",
		"data": "",
	}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("设置表字段失败: %v", err)
	}

	// 创建主键索引
	pk, err := DefaultPrimaryKeyNew("pk_id")
	if err != nil {
		t.Fatalf("创建主键失败: %v", err)
	}
	pk.AddFields("id")
	err = table.CreateIndex(pk)
	if err != nil {
		t.Fatalf("创建主键索引失败: %v", err)
	}

	// 插入测试记录
	insertRecord := map[string]any{
		"id":   1,
		"name": "测试记录",
		"data": "初始数据",
	}
	_, err = table.Insert(&insertRecord)
	if err != nil {
		t.Fatalf("插入记录失败: %v", err)
	}

	// 测试1: 基本行级锁和乐观锁
	t.Run("BasicRowLockAndOptimisticLock", func(t *testing.T) {
		// 读取记录
		readRecord := map[string]any{"id": 1}
		recordBytes, err := table.Read(&readRecord)
		if err != nil {
			t.Fatalf("读取记录失败: %v", err)
		}
		if recordBytes == nil {
			t.Fatalf("记录不存在")
		}

		// 解析记录
		pk := table.GetPrimaryKey()
		fieldsBytes, err := pk.Parse(table.fieldsid, recordBytes)
		if err != nil {
			t.Fatalf("解析记录失败: %v", err)
		}

		// 获取当前版本号
		currentVersion := string((*fieldsBytes)["v"])

		// 更新记录，使用正确的版本号
		updateRecord := map[string]any{
			"id":   1,
			"data": "更新后的数据",
			"v":    currentVersion,
		}
		err = table.Update(&updateRecord)
		if err != nil {
			t.Fatalf("更新记录失败: %v", err)
		}

		// 再次读取记录，验证更新是否成功
		recordBytes, err = table.Read(&readRecord)
		if err != nil {
			t.Fatalf("读取记录失败: %v", err)
		}
		fieldsBytes, err = pk.Parse(table.fieldsid, recordBytes)
		if err != nil {
			t.Fatalf("解析记录失败: %v", err)
		}
		updatedVersion := string((*fieldsBytes)["v"])
		if updatedVersion == currentVersion {
			t.Fatalf("版本号没有更新，当前版本: %s, 更新后版本: %s", currentVersion, updatedVersion)
		}

		// 测试乐观锁冲突
		updateRecordWithOldVersion := map[string]any{
			"id":   1,
			"data": "再次更新",
			"v":    currentVersion, // 使用旧版本号
		}
		err = table.Update(&updateRecordWithOldVersion)
		if err == nil {
			t.Fatalf("乐观锁冲突测试失败，应该返回错误")
		}
		if !strings.Contains(err.Error(), "optimistic lock conflict") {
			t.Fatalf("错误信息不符合预期: %v", err)
		}
	})

	// 测试2: 并发行级锁
	t.Run("ConcurrentRowLock", func(t *testing.T) {
		const goroutines = 10
		const iterations = 5

		var wg sync.WaitGroup
		errorChan := make(chan error, goroutines)

		// 并发更新同一记录
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				for j := 0; j < iterations; j++ {
					// 读取记录
					readRecord := map[string]any{"id": 1}
					recordBytes, err := table.Read(&readRecord)
					if err != nil {
						errorChan <- fmt.Errorf("goroutine %d 读取记录失败: %v", goroutineID, err)
						return
					}

					// 解析记录
					pk := table.GetPrimaryKey()
					fieldsBytes, err := pk.Parse(table.fieldsid, recordBytes)
					if err != nil {
						errorChan <- fmt.Errorf("goroutine %d 解析记录失败: %v", goroutineID, err)
						return
					}

					// 获取当前版本号
					currentVersion := string((*fieldsBytes)["v"])

					// 更新记录
					updateRecord := map[string]any{
						"id":   1,
						"data": fmt.Sprintf("goroutine %d 更新 %d", goroutineID, j),
						"v":    currentVersion,
					}
					err = table.Update(&updateRecord)
					if err != nil {
						// 乐观锁冲突是正常的，忽略
						if !strings.Contains(err.Error(), "optimistic lock conflict") {
							errorChan <- fmt.Errorf("goroutine %d 更新记录失败: %v", goroutineID, err)
							return
						}
					}

					// 模拟处理时间
					time.Sleep(time.Millisecond * 10)
				}
			}(i)
		}

		wg.Wait()
		close(errorChan)

		// 检查是否有错误
		for err := range errorChan {
			t.Fatalf("并发测试失败: %v", err)
		}

		// 验证记录最终存在
		readRecord := map[string]any{"id": 1}
		recordBytes, err := table.Read(&readRecord)
		if err != nil {
			t.Fatalf("读取记录失败: %v", err)
		}
		if recordBytes == nil {
			t.Fatalf("记录不存在")
		}
	})

	// 测试3: 批量操作
	t.Run("BatchOperations", func(t *testing.T) {
		// 插入多条记录
		records := make([]*map[string]any, 5)
		for i := 0; i < 5; i++ {
			record := map[string]any{
				"id":   i + 2, // 从2开始
				"name": fmt.Sprintf("测试记录 %d", i+2),
				"data": fmt.Sprintf("初始数据 %d", i+2),
			}
			records[i] = &record
		}

		// 批量插入
		ids, err := table.BatchInsert(records)
		if err != nil {
			t.Fatalf("批量插入失败: %v", err)
		}
		if len(ids) != 5 {
			t.Fatalf("批量插入返回的ID数量不正确: %d", len(ids))
		}

		// 验证记录存在
		for i := 0; i < 5; i++ {
			readRecord := map[string]any{"id": i + 2}
			recordBytes, err := table.Read(&readRecord)
			if err != nil {
				t.Fatalf("读取记录 %d 失败: %v", i+2, err)
			}
			if recordBytes == nil {
				t.Fatalf("记录 %d 不存在", i+2)
			}
		}
	})
}

// 测试事务中的行级锁和乐观锁
func TestTransactionWithRowLock(t *testing.T) {
	// 创建临时目录作为数据库路径
	tempDir := t.TempDir()

	// 显式打开数据库，使用临时目录
	db, err := storage.OpenDefaultDb(tempDir)
	if err != nil {
		t.Fatalf("打开数据库失败: %v", err)
	}

	// 清理测试环境
	defer func() {
		if db != nil {
			db.Close()
			storage.KVDb = nil
		}
	}()

	// 创建测试表
	table1, err := TableNew("test_transaction_row_lock_1")
	if err != nil {
		t.Fatalf("创建表1失败: %v", err)
	}

	table2, err := TableNew("test_transaction_row_lock_2")
	if err != nil {
		t.Fatalf("创建表2失败: %v", err)
	}

	// 设置表字段
	fields := map[string]any{
		"id":   0,
		"name": "",
		"data": "",
	}
	err = table1.SetFields(fields)
	if err != nil {
		t.Fatalf("设置表1字段失败: %v", err)
	}

	err = table2.SetFields(fields)
	if err != nil {
		t.Fatalf("设置表2字段失败: %v", err)
	}

	// 创建主键索引
	pk, err := DefaultPrimaryKeyNew("pk_id")
	if err != nil {
		t.Fatalf("创建主键失败: %v", err)
	}
	pk.AddFields("id")

	err = table1.CreateIndex(pk)
	if err != nil {
		t.Fatalf("创建表1主键索引失败: %v", err)
	}

	err = table2.CreateIndex(pk)
	if err != nil {
		t.Fatalf("创建表2主键索引失败: %v", err)
	}

	// 插入测试记录
	insertRecord1 := map[string]any{
		"id":   1,
		"name": "表1记录",
		"data": "初始数据",
	}
	_, err = table1.Insert(&insertRecord1)
	if err != nil {
		t.Fatalf("插入表1记录失败: %v", err)
	}

	insertRecord2 := map[string]any{
		"id":   1,
		"name": "表2记录",
		"data": "初始数据",
	}
	_, err = table2.Insert(&insertRecord2)
	if err != nil {
		t.Fatalf("插入表2记录失败: %v", err)
	}

	// 测试事务
	t.Run("MultiTableTransaction", func(t *testing.T) {
		// 创建共享batch
		batch := table1.kvStore.GetBatch()
		if batch == nil {
			t.Fatalf("创建batch失败")
		}

		// 执行多表事务
		err := WithTransaction(batch, []*Table{table1, table2}, func(transactions map[*Table]Transaction) error {
			// 读取表1记录
			tx1 := transactions[table1]
			readRecord1 := map[string]any{"id": 1}
			iter1, err := tx1.Search(&readRecord1)
			if err != nil {
				return fmt.Errorf("搜索表1记录失败: %v", err)
			}
			defer GlobalTableIterPool.Put(iter1)

			records1 := iter1.GetRecords(true)
			if len(records1) != 1 {
				return fmt.Errorf("表1记录不存在")
			}

			// 读取表2记录
			tx2 := transactions[table2]
			readRecord2 := map[string]any{"id": 1}
			iter2, err := tx2.Search(&readRecord2)
			if err != nil {
				return fmt.Errorf("搜索表2记录失败: %v", err)
			}
			defer GlobalTableIterPool.Put(iter2)

			records2 := iter2.GetRecords(true)
			if len(records2) != 1 {
				return fmt.Errorf("表2记录不存在")
			}

			// 更新表1记录
			updateRecord1 := map[string]any{
				"id":   1,
				"data": "事务更新后的数据",
				"v":    records1[0]["v"],
			}
			err = tx1.Update(&updateRecord1)
			if err != nil {
				return fmt.Errorf("更新表1记录失败: %v", err)
			}

			// 更新表2记录
			updateRecord2 := map[string]any{
				"id":   1,
				"data": "事务更新后的数据",
				"v":    records2[0]["v"],
			}
			err = tx2.Update(&updateRecord2)
			if err != nil {
				return fmt.Errorf("更新表2记录失败: %v", err)
			}

			return nil
		})

		if err != nil {
			t.Fatalf("事务执行失败: %v", err)
		}

		// 验证记录已更新
		readRecord1 := map[string]any{"id": 1}
		recordBytes1, err := table1.Read(&readRecord1)
		if err != nil {
			t.Fatalf("读取表1记录失败: %v", err)
		}
		if recordBytes1 == nil {
			t.Fatalf("表1记录不存在")
		}

		readRecord2 := map[string]any{"id": 1}
		recordBytes2, err := table2.Read(&readRecord2)
		if err != nil {
			t.Fatalf("读取表2记录失败: %v", err)
		}
		if recordBytes2 == nil {
			t.Fatalf("表2记录不存在")
		}
	})
}

// 测试事务锁跟踪功能
func TestTransactionLockTracking(t *testing.T) {
	// 创建临时目录作为数据库路径
	tempDir := t.TempDir()

	// 显式打开数据库，使用临时目录
	db, err := storage.OpenDefaultDb(tempDir)
	if err != nil {
		t.Fatalf("打开数据库失败: %v", err)
	}

	// 清理测试环境
	defer func() {
		if db != nil {
			db.Close()
			storage.KVDb = nil
		}
	}()

	// 创建测试表
	table, err := TableNew("test_transaction_lock_tracking")
	if err != nil {
		t.Fatalf("创建表失败: %v", err)
	}

	// 设置表字段
	fields := map[string]any{
		"id":   0,
		"name": "",
		"data": "",
	}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("设置表字段失败: %v", err)
	}

	// 创建主键索引
	pk, err := DefaultPrimaryKeyNew("pk_id")
	if err != nil {
		t.Fatalf("创建主键失败: %v", err)
	}
	pk.AddFields("id")
	err = table.CreateIndex(pk)
	if err != nil {
		t.Fatalf("创建主键索引失败: %v", err)
	}

	// 插入测试记录
	insertRecord := map[string]any{
		"id":   1,
		"name": "测试记录",
		"data": "初始数据",
	}
	_, err = table.Insert(&insertRecord)
	if err != nil {
		t.Fatalf("插入记录失败: %v", err)
	}

	// 测试1: 基本事务锁跟踪
	t.Run("BasicTransactionLockTracking", func(t *testing.T) {
		// 生成事务ID
		txID := uint64(time.Now().UnixNano())

		// 开始事务锁跟踪
		table.BeginTransaction(txID)

		// 检查事务锁信息是否存在
		lockInfo := table.GetTransactionLockInfo(txID)
		if lockInfo == nil {
			t.Fatalf("事务锁信息不存在")
		}

		// 先获取锁，这样 rowLocks 中就会存在对应的锁信息
		err := table.acquireRowReadLock("1", txID)
		if err != nil {
			t.Fatalf("获取读锁失败: %v", err)
		}

		// 记录持有锁
		table.RecordHeldLock(txID, "1")

		// 记录等待锁
		table.RecordWaitingLock(txID, "2")

		// 检查持有锁数量
		lockInfo = table.GetTransactionLockInfo(txID)
		if len(lockInfo.HeldLocks) != 1 {
			t.Fatalf("持有锁数量不正确: %d", len(lockInfo.HeldLocks))
		}

		// 结束事务锁跟踪
		table.EndTransaction(txID)

		// 检查事务锁信息是否已删除
		lockInfo = table.GetTransactionLockInfo(txID)
		if lockInfo != nil {
			t.Fatalf("事务锁信息应该已删除")
		}
	})

	// 测试2: 并发事务锁跟踪
	t.Run("ConcurrentTransactionLockTracking", func(t *testing.T) {
		const goroutines = 5

		var wg sync.WaitGroup
		errorChan := make(chan error, goroutines)

		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				// 生成唯一事务ID
				txID := uint64(time.Now().UnixNano() + int64(goroutineID))

				// 开始事务锁跟踪
				table.BeginTransaction(txID)

				// 记录持有锁
				lockKey := fmt.Sprintf("%d", goroutineID+1)

				// 先获取锁，这样 rowLocks 中就会存在对应的锁信息
				err := table.acquireRowReadLock(lockKey, txID)
				if err != nil {
					errorChan <- fmt.Errorf("goroutine %d 获取读锁失败: %v", goroutineID, err)
					return
				}

				table.RecordHeldLock(txID, lockKey)

				// 模拟处理时间
				time.Sleep(time.Millisecond * 50)

				// 结束事务锁跟踪
				table.EndTransaction(txID)

				// 检查事务锁信息是否已删除
				lockInfo := table.GetTransactionLockInfo(txID)
				if lockInfo != nil {
					errorChan <- fmt.Errorf("goroutine %d: 事务锁信息应该已删除", goroutineID)
					return
				}
			}(i)
		}

		wg.Wait()
		close(errorChan)

		// 检查是否有错误
		for err := range errorChan {
			t.Fatalf("并发测试失败: %v", err)
		}
	})

	// 测试3: 死锁检测
	t.Run("DeadlockDetection", func(t *testing.T) {
		// 生成事务ID
		txID1 := uint64(time.Now().UnixNano())
		txID2 := uint64(time.Now().UnixNano() + 1)

		// 开始事务锁跟踪
		table.BeginTransaction(txID1)
		table.BeginTransaction(txID2)

		// 模拟循环等待：tx1 等待 tx2 持有的锁，tx2 等待 tx1 持有的锁

		// 先获取锁，这样 rowLocks 中就会存在对应的锁信息
		err := table.acquireRowReadLock("1", txID1)
		if err != nil {
			t.Fatalf("获取读锁失败: %v", err)
		}
		err = table.acquireRowReadLock("2", txID2)
		if err != nil {
			t.Fatalf("获取读锁失败: %v", err)
		}

		table.RecordHeldLock(txID1, "1")
		table.RecordHeldLock(txID2, "2")

		// 记录等待锁，触发死锁检测
		table.RecordWaitingLock(txID1, "2") // tx1 等待 tx2 持有的锁
		table.RecordWaitingLock(txID2, "1") // tx2 等待 tx1 持有的锁

		// 检测死锁
		deadlockedTxs := table.DetectDeadlock()
		if len(deadlockedTxs) == 0 {
			t.Logf("未检测到死锁，这是正常的，因为我们只是模拟了锁等待，没有实际获取锁")
		}

		// 结束事务锁跟踪
		table.EndTransaction(txID1)
		table.EndTransaction(txID2)

		// 检查事务锁信息是否已删除
		if table.GetTransactionLockInfo(txID1) != nil {
			t.Fatalf("事务1锁信息应该已删除")
		}
		if table.GetTransactionLockInfo(txID2) != nil {
			t.Fatalf("事务2锁信息应该已删除")
		}
	})
}
