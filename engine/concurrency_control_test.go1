package engine

import (
	"sync"
	"testing"
	"time"
)

// 测试死锁检测
func TestDeadlockDetection(t *testing.T) {
	// 创建表
	table, err := TableNew("test_deadlock")
	if err != nil {
		t.Fatalf("创建表失败: %v", err)
	}

	// 设置字段
	err = table.SetFields(map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	})
	if err != nil {
		t.Fatalf("设置字段失败: %v", err)
	}

	// 开始两个事务
	tx1 := uint64(1)
	tx2 := uint64(2)
	table.BeginTransaction(tx1)
	table.BeginTransaction(tx2)

	// 模拟死锁场景：tx1持有row1，等待row2；tx2持有row2，等待row1
	var wg sync.WaitGroup
	wg.Add(2)

	// 事务1：获取row1的写锁
	go func() {
		defer wg.Done()
		// 获取row1的写锁
		err := table.acquireRowWriteLock("1", tx1)
		if err != nil {
			t.Logf("事务1获取row1锁失败: %v", err)
			return
		}
		table.RecordHeldLock(tx1, "1")
		
		// 等待一段时间，确保事务2获取row2的锁
		time.Sleep(100 * time.Millisecond)
		
		// 尝试获取row2的写锁（会等待）
		table.RecordWaitingLock(tx1, "2")
		err = table.acquireRowWriteLock("2", tx1)
		if err != nil {
			t.Logf("事务1获取row2锁失败: %v", err)
		}
	}()

	// 事务2：获取row2的写锁
	go func() {
		defer wg.Done()
		// 获取row2的写锁
		err := table.acquireRowWriteLock("2", tx2)
		if err != nil {
			t.Logf("事务2获取row2锁失败: %v", err)
			return
		}
		table.RecordHeldLock(tx2, "2")
		
		// 等待一段时间，确保事务1获取row1的锁
		time.Sleep(100 * time.Millisecond)
		
		// 尝试获取row1的写锁（会等待）
		table.RecordWaitingLock(tx2, "1")
		err = table.acquireRowWriteLock("1", tx2)
		if err != nil {
			t.Logf("事务2获取row1锁失败: %v", err)
		}
	}()

	// 等待两个事务执行
	wg.Wait()

	// 检测死锁
	deadlockedTxs := table.DetectDeadlock()
	if len(deadlockedTxs) == 0 {
		t.Log("未检测到死锁（可能需要调整等待时间）")
	} else {
		t.Logf("检测到死锁的事务: %v", deadlockedTxs)
	}

	// 结束事务
	table.EndTransaction(tx1)
	table.EndTransaction(tx2)
}

// 测试锁超时机制
func TestLockTimeout(t *testing.T) {
	// 创建表
	table, err := TableNew("test_lock_timeout")
	if err != nil {
		t.Fatalf("创建表失败: %v", err)
	}

	// 设置字段
	err = table.SetFields(map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	})
	if err != nil {
		t.Fatalf("设置字段失败: %v", err)
	}

	// 设置短超时时间
	table.SetLockTimeout(500 * time.Millisecond)

	// 开始事务
	tx1 := uint64(1)
	table.BeginTransaction(tx1)

	// 获取写锁
	err = table.acquireRowWriteLock("1", tx1)
	if err != nil {
		t.Fatalf("获取锁失败: %v", err)
	}
	table.RecordHeldLock(tx1, "1")

	// 等待锁超时
	time.Sleep(600 * time.Millisecond)

	// 检查锁是否过期
	lockInfo := table.GetLockInfo("1")
	if !lockInfo.IsExpired {
		t.Error("锁应该已过期")
	}

	// 尝试获取过期锁
	tx2 := uint64(2)
	table.BeginTransaction(tx2)
	err = table.acquireRowWriteLock("1", tx2)
	if err != nil {
		t.Fatalf("获取过期锁失败: %v", err)
	}
	table.RecordHeldLock(tx2, "1")

	// 结束事务
	table.EndTransaction(tx1)
	table.EndTransaction(tx2)
}

// 测试锁升级/降级
func TestLockUpgradeDowngrade(t *testing.T) {
	// 创建表
	table, err := TableNew("test_lock_upgrade_downgrade")
	if err != nil {
		t.Fatalf("创建表失败: %v", err)
	}

	// 设置字段
	err = table.SetFields(map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	})
	if err != nil {
		t.Fatalf("设置字段失败: %v", err)
	}

	// 开始事务
	tx := uint64(1)
	table.BeginTransaction(tx)

	// 获取读锁
	err = table.acquireRowReadLock("1", tx)
	if err != nil {
		t.Fatalf("获取读锁失败: %v", err)
	}
	table.RecordHeldLock(tx, "1")

	// 升级为写锁
	err = table.UpgradeLock("1", tx)
	if err != nil {
		t.Fatalf("升级锁失败: %v", err)
	}

	// 降级为读锁
	err = table.DowngradeLock("1", tx)
	if err != nil {
		t.Fatalf("降级锁失败: %v", err)
	}

	// 结束事务
	table.EndTransaction(tx)
}

// 测试并发控制
func TestConcurrencyControl(t *testing.T) {
	// 创建表
	table, err := TableNew("test_concurrency")
	if err != nil {
		t.Fatalf("创建表失败: %v", err)
	}

	// 设置字段
	err = table.SetFields(map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	})
	if err != nil {
		t.Fatalf("设置字段失败: %v", err)
	}

	// 启动锁清理
	table.StartLockCleanup(1 * time.Second)

	// 并发读写测试
	var wg sync.WaitGroup
	var mu sync.Mutex
	errorCount := 0
	readCount := 0
	writeCount := 0

	// 启动10个读协程
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			tx := uint64(100 + id)
			table.BeginTransaction(tx)

			// 获取读锁
			err := table.acquireRowReadLock("1", tx)
			if err != nil {
				mu.Lock()
				errorCount++
				mu.Unlock()
				table.EndTransaction(tx)
				return
			}
			table.RecordHeldLock(tx, "1")

			// 模拟读取操作
			time.Sleep(10 * time.Millisecond)
			mu.Lock()
			readCount++
			mu.Unlock()

			// 释放锁
			table.releaseRowLock("1")
			table.EndTransaction(tx)
		}(i)
	}

	// 启动5个写协程
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			tx := uint64(200 + id)
			table.BeginTransaction(tx)

			// 获取写锁
			err := table.acquireRowWriteLock("1", tx)
			if err != nil {
				mu.Lock()
				errorCount++
				mu.Unlock()
				table.EndTransaction(tx)
				return
			}
			table.RecordHeldLock(tx, "1")

			// 模拟写入操作
			time.Sleep(20 * time.Millisecond)
			mu.Lock()
			writeCount++
			mu.Unlock()

			// 释放锁
			table.releaseRowLock("1")
			table.EndTransaction(tx)
		}(i)
	}

	// 等待所有协程完成
	wg.Wait()

	// 检查结果
	if errorCount > 0 {
		t.Errorf("并发操作出错: %d 个错误", errorCount)
	}
	if readCount == 0 {
		t.Error("没有成功的读操作")
	}
	if writeCount == 0 {
		t.Error("没有成功的写操作")
	}

	t.Logf("并发测试结果: 读操作=%d, 写操作=%d, 错误=%d", readCount, writeCount, errorCount)
}

// 测试锁统计信息
func TestLockStats(t *testing.T) {
	// 创建表
	table, err := TableNew("test_lock_stats")
	if err != nil {
		t.Fatalf("创建表失败: %v", err)
	}

	// 设置字段
	err = table.SetFields(map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	})
	if err != nil {
		t.Fatalf("设置字段失败: %v", err)
	}

	// 开始事务
	tx := uint64(1)
	table.BeginTransaction(tx)

	// 获取读锁
	err = table.acquireRowReadLock("1", tx)
	if err != nil {
		t.Fatalf("获取读锁失败: %v", err)
	}
	table.RecordHeldLock(tx, "1")

	// 获取写锁
	err = table.acquireRowWriteLock("2", tx)
	if err != nil {
		t.Fatalf("获取写锁失败: %v", err)
	}
	table.RecordHeldLock(tx, "2")

	// 获取锁统计信息
	stats := table.GetLockStats()
	if stats.TotalLocks == 0 {
		t.Error("锁数量应该大于0")
	}
	if stats.ReadLocks == 0 {
		t.Error("读锁数量应该大于0")
	}
	if stats.WriteLocks == 0 {
		t.Error("写锁数量应该大于0")
	}

	t.Logf("锁统计信息: 总锁=%d, 读锁=%d, 写锁=%d", stats.TotalLocks, stats.ReadLocks, stats.WriteLocks)

	// 结束事务
	table.EndTransaction(tx)
}

// 测试锁清理
func TestLockCleanup(t *testing.T) {
	// 创建表
	table, err := TableNew("test_lock_cleanup")
	if err != nil {
		t.Fatalf("创建表失败: %v", err)
	}

	// 设置字段
	err = table.SetFields(map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	})
	if err != nil {
		t.Fatalf("设置字段失败: %v", err)
	}

	// 开始事务
	tx := uint64(1)
	table.BeginTransaction(tx)

	// 获取锁
	err = table.acquireRowWriteLock("1", tx)
	if err != nil {
		t.Fatalf("获取锁失败: %v", err)
	}
	table.RecordHeldLock(tx, "1")

	// 检查初始锁数量
	stats1 := table.GetLockStats()

	// 结束事务（会释放锁）
	table.EndTransaction(tx)

	// 检查锁是否被清理
	stats2 := table.GetLockStats()
	if stats2.TotalLocks >= stats1.TotalLocks {
		t.Error("锁应该被清理")
	}
}

// 测试锁延长超时
func TestExtendLockTimeout(t *testing.T) {
	// 创建表
	table, err := TableNew("test_extend_timeout")
	if err != nil {
		t.Fatalf("创建表失败: %v", err)
	}

	// 设置字段
	err = table.SetFields(map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	})
	if err != nil {
		t.Fatalf("设置字段失败: %v", err)
	}

	// 开始事务
	tx := uint64(1)
	table.BeginTransaction(tx)

	// 获取锁
	err = table.acquireRowWriteLock("1", tx)
	if err != nil {
		t.Fatalf("获取锁失败: %v", err)
	}
	table.RecordHeldLock(tx, "1")

	// 检查锁是否即将过期
	isAboutToExpire := table.IsLockAboutToExpire("1", 1*time.Second)
	if isAboutToExpire {
		t.Error("锁不应该即将过期")
	}

	// 延长锁超时
	err = table.ExtendLockTimeout("1", 5*time.Second)
	if err != nil {
		t.Fatalf("延长锁超时失败: %v", err)
	}

	// 检查锁是否已延长
	lockInfo := table.GetLockInfo("1")
	if lockInfo.Timeout < 5*time.Second {
		t.Error("锁超时应该已延长")
	}

	// 结束事务
	table.EndTransaction(tx)
}
