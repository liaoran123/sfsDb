package monitor

import (
	"sync"
	"testing"
	"time"
)

// TestSettimeConcurrent 测试Settime方法的并发安全性
func TestSettimeConcurrent(t *testing.T) {
	// 创建一个新的IndexStatsMap实例
	statsMap := NewIndexStatsMap()

	// 测试的goroutine数量
	goroutineCount := 1000

	// 使用WaitGroup等待所有goroutine完成
	var wg sync.WaitGroup
	wg.Add(goroutineCount)

	// 启动多个goroutine同时调用Settime方法
	for i := 0; i < goroutineCount; i++ {
		go func(index int) {
			defer wg.Done()

			// 生成不同的indexKey
			indexKey := index % 10 // 使用0-9的indexKey

			// 调用Settime方法
			statsMap.Settime(indexKey, time.Millisecond, "testTable", "testIndex", "testSearchType")
		}(i)
	}

	// 等待所有goroutine完成
	wg.Wait()

	// 检查结果
	t.Logf("测试完成，共处理 %d 个请求，创建了 %d 个IndexStats实例", goroutineCount, len(statsMap.GetAll()))

	// 验证所有IndexStats实例的Count字段是否正确
	for key, stats := range statsMap.GetAll() {
		count := stats.Count.Load()
		t.Logf("IndexKey: %d, Count: %d", key, count)
	}
}
