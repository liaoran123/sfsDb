package transaction

import (
	"testing"
	"time"

	"github.com/liaoran123/sfsDb/transaction/batch"
	"github.com/liaoran123/sfsDb/transaction/lock"
)

// BenchmarkLockOperations 测试锁操作性能
func BenchmarkLockOperations(b *testing.B) {
	// 初始化锁管理器
	lockManager := lock.NewManager()

	b.ResetTimer()

	// 并发测试锁获取和释放
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// 尝试获取锁
			err := lockManager.AcquireLock(lock.LockRequest{
				Key:       "test_key",
				Type:      lock.ReadLock,
				Mode:      lock.NonBlocking,
				Timeout:   0,
				Requestor: "benchmark",
			})
			if err == nil {
				// 释放锁
				lockManager.ReleaseLock("test_key", "benchmark")
			}
		}
	})
}

// BenchmarkWriteLockOperations 测试写锁操作性能
func BenchmarkWriteLockOperations(b *testing.B) {
	// 初始化锁管理器
	lockManager := lock.NewManager()

	b.ResetTimer()

	// 并发测试写锁获取和释放
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// 尝试获取写锁
			err := lockManager.AcquireLock(lock.LockRequest{
				Key:       "test_key",
				Type:      lock.WriteLock,
				Mode:      lock.NonBlocking,
				Timeout:   0,
				Requestor: "benchmark",
			})
			if err == nil {
				// 释放锁
				lockManager.ReleaseLock("test_key", "benchmark")
			}
		}
	})
}

// BenchmarkBatchOperations 测试批量操作性能
func BenchmarkBatchOperations(b *testing.B) {
	// 初始化批量操作优化器
	batchOptimizer := batch.NewOptimizer()

	b.ResetTimer()

	// 测试批量操作统计
	for i := 0; i < b.N; i++ {
		// 记录操作
		batchOptimizer.RecordOperation(10, 1*time.Millisecond)
	}
}

// BenchmarkLockStats 测试锁统计信息获取性能
func BenchmarkLockStats(b *testing.B) {
	// 初始化锁管理器
	lockManager := lock.NewManager()

	// 预先进行一些锁操作
	for i := 0; i < 1000; i++ {
		err := lockManager.AcquireLock(lock.LockRequest{
			Key:       "test_key",
			Type:      lock.ReadLock,
			Mode:      lock.NonBlocking,
			Timeout:   0,
			Requestor: "benchmark",
		})
		if err == nil {
			lockManager.ReleaseLock("test_key", "benchmark")
		}
	}

	b.ResetTimer()

	// 测试获取统计信息
	for i := 0; i < b.N; i++ {
		lockManager.GetStats()
	}
}
