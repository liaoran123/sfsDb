package engine

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/liaoran123/sfsDb/storage"
)

// TestOptimisticLock 测试乐观锁的基本功能
func TestOptimisticLock(t *testing.T) {
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
	table, err := TableNew("test_optimistic_lock")
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

	// 测试1: 基本乐观锁功能
	t.Run("BasicOptimisticLock", func(t *testing.T) {
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
		t.Logf("当前版本号: %s", currentVersion)

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
		t.Logf("更新后版本号: %s", updatedVersion)

		if updatedVersion == currentVersion {
			t.Fatalf("版本号没有更新，当前版本: %s, 更新后版本: %s", currentVersion, updatedVersion)
		}
	})

	// 测试2: 乐观锁冲突
	t.Run("OptimisticLockConflict", func(t *testing.T) {
		// 读取记录
		readRecord := map[string]any{"id": 1}
		recordBytes, err := table.Read(&readRecord)
		if err != nil {
			t.Fatalf("读取记录失败: %v", err)
		}

		// 解析记录
		pk := table.GetPrimaryKey()
		fieldsBytes, err := pk.Parse(table.fieldsid, recordBytes)
		if err != nil {
			t.Fatalf("解析记录失败: %v", err)
		}

		// 获取当前版本号
		currentVersion := string((*fieldsBytes)["v"])

		// 第一次更新，使用正确的版本号
		updateRecord1 := map[string]any{
			"id":   1,
			"data": "第一次更新",
			"v":    currentVersion,
		}
		err = table.Update(&updateRecord1)
		if err != nil {
			t.Fatalf("第一次更新记录失败: %v", err)
		}

		// 第二次更新，使用旧版本号，应该失败
		updateRecord2 := map[string]any{
			"id":   1,
			"data": "第二次更新",
			"v":    currentVersion, // 使用旧版本号
		}
		err = table.Update(&updateRecord2)
		if err == nil {
			t.Fatalf("乐观锁冲突测试失败，应该返回错误")
		}
		if !strings.Contains(err.Error(), "optimistic lock conflict") {
			t.Fatalf("错误信息不符合预期: %v", err)
		}
		t.Logf("乐观锁冲突测试成功，错误信息: %v", err)
	})

	// 测试3: 并发场景下的乐观锁
	t.Run("ConcurrentOptimisticLock", func(t *testing.T) {
		const goroutines = 5
		const iterations = 3

		var wg sync.WaitGroup
		var successCount int32
		var failCount int32

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
						t.Errorf("goroutine %d 读取记录失败: %v", goroutineID, err)
						continue
					}

					// 解析记录
					pk := table.GetPrimaryKey()
					fieldsBytes, err := pk.Parse(table.fieldsid, recordBytes)
					if err != nil {
						t.Errorf("goroutine %d 解析记录失败: %v", goroutineID, err)
						continue
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
							t.Errorf("goroutine %d 更新记录失败: %v", goroutineID, err)
						}
						atomic.AddInt32(&failCount, 1)
					} else {
						atomic.AddInt32(&successCount, 1)
						t.Logf("goroutine %d 更新成功", goroutineID)
					}

					// 模拟处理时间
					time.Sleep(time.Millisecond * 5)
				}
			}(i)
		}

		wg.Wait()

		t.Logf("并发测试完成，成功次数: %d, 失败次数: %d", successCount, failCount)
		if successCount == 0 {
			t.Fatalf("并发测试失败，没有成功的更新")
		}
	})

	// 测试4: 批量操作中的乐观锁
	t.Run("BatchOperationOptimisticLock", func(t *testing.T) {
		// 插入多条测试记录
		records := make([]*map[string]any, 3)
		for i := 0; i < 3; i++ {
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
		if len(ids) != 3 {
			t.Fatalf("批量插入返回的ID数量不正确: %d", len(ids))
		}

		// 批量更新（使用乐观锁）
		batch := table.kvStore.GetBatch()
		for i := 0; i < 3; i++ {
			// 读取记录
			readRecord := map[string]any{"id": i + 2}
			recordBytes, err := table.Read(&readRecord)
			if err != nil {
				t.Fatalf("读取记录 %d 失败: %v", i+2, err)
			}

			// 解析记录
			pk := table.GetPrimaryKey()
			fieldsBytes, err := pk.Parse(table.fieldsid, recordBytes)
			if err != nil {
				t.Fatalf("解析记录 %d 失败: %v", i+2, err)
			}

			// 获取当前版本号
			currentVersion := string((*fieldsBytes)["v"])

			// 更新记录
			updateRecord := map[string]any{
				"id":   i + 2,
				"data": fmt.Sprintf("批量更新后的数据 %d", i+2),
				"v":    currentVersion,
			}
			err = table.Update(&updateRecord, batch)
			if err != nil {
				t.Fatalf("更新记录 %d 失败: %v", i+2, err)
			}
		}

		// 提交批量操作
		err = table.kvStore.WriteBatch(batch)
		if err != nil {
			t.Fatalf("提交批量操作失败: %v", err)
		}

		// 验证批量更新是否成功
		for i := 0; i < 3; i++ {
			readRecord := map[string]any{"id": i + 2}
			recordBytes, err := table.Read(&readRecord)
			if err != nil {
				t.Fatalf("读取记录 %d 失败: %v", i+2, err)
			}
			if recordBytes == nil {
				t.Fatalf("记录 %d 不存在", i+2)
			}

			// 解析记录
			pk := table.GetPrimaryKey()
			fieldsBytes, err := pk.Parse(table.fieldsid, recordBytes)
			if err != nil {
				t.Fatalf("解析记录 %d 失败: %v", i+2, err)
			}

			// 验证数据是否更新
			data := string((*fieldsBytes)["data"])
			expectedData := fmt.Sprintf("批量更新后的数据 %d", i+2)
			if data != expectedData {
				t.Fatalf("记录 %d 数据更新失败，期望: %s, 实际: %s", i+2, expectedData, data)
			}
			t.Logf("记录 %d 批量更新成功，数据: %s", i+2, data)
		}
	})
}

// TestOptimisticLockInTransaction 测试事务中的乐观锁
func TestOptimisticLockInTransaction(t *testing.T) {
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
	table, err := TableNew("test_optimistic_lock_transaction")
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
		"name": "事务测试记录",
		"data": "初始数据",
	}
	_, err = table.Insert(&insertRecord)
	if err != nil {
		t.Fatalf("插入记录失败: %v", err)
	}

	// 测试事务中的乐观锁
	t.Run("TransactionOptimisticLock", func(t *testing.T) {
		// 开始事务
		tx, err := table.Begin()
		if err != nil {
			t.Fatalf("开始事务失败: %v", err)
		}

		// 读取记录
		readRecord := map[string]any{"id": 1}
		recordBytes, err := tx.Read(&readRecord)
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
		t.Logf("当前版本号: %s", currentVersion)

		// 更新记录，使用正确的版本号
		updateRecord := map[string]any{
			"id":   1,
			"data": "事务中更新后的数据",
			"v":    currentVersion,
		}
		err = tx.Update(&updateRecord)
		if err != nil {
			t.Fatalf("更新记录失败: %v", err)
		}

		// 提交事务
		err = tx.Commit()
		if err != nil {
			t.Fatalf("提交事务失败: %v", err)
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
		t.Logf("更新后版本号: %s", updatedVersion)

		if updatedVersion == currentVersion {
			t.Fatalf("版本号没有更新，当前版本: %s, 更新后版本: %s", currentVersion, updatedVersion)
		}

		// 验证数据是否更新
		data := string((*fieldsBytes)["data"])
		if data != "事务中更新后的数据" {
			t.Fatalf("数据更新失败，期望: 事务中更新后的数据, 实际: %s", data)
		}
		t.Logf("事务中乐观锁测试成功，数据: %s", data)
	})
}
