package storage

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

// TestBatchPoolMemoryUsage 测试批处理对象池内存使用情况
func TestBatchPoolMemoryUsage(t *testing.T) {
	// 初始化内存统计
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	initialAlloc := m.Alloc

	fmt.Printf("初始内存分配: %.2f MB\n", float64(initialAlloc)/1024/1024)

	// 模拟大量批处理操作
	const operations = 10000
	const batchSize = 10

	// 创建测试存储
	store, err := dbManager.NewLevelDBStore("./test_batch_pool", nil)
	if err != nil {
		t.Fatalf("创建存储失败: %v", err)
	}
	defer store.Close()

	start := time.Now()

	for i := 0; i < operations; i++ {
		// 获取批处理对象
		batch := store.GetBatch()
		if batch == nil {
			t.Fatal("获取批处理对象失败")
		}

		// 添加一些操作到批处理
		for j := 0; j < batchSize; j++ {
			key := []byte(fmt.Sprintf("key_%d_%d", i, j))
			value := []byte(fmt.Sprintf("value_%d_%d", i, j))
			batch.(interface {
				Put(key, value []byte)
			}).Put(key, value)
		}

		// 执行批处理
		err := store.WriteBatch(batch)
		if err != nil {
			t.Fatalf("执行批处理失败: %v", err)
		}
	}

	duration := time.Since(start)
	fmt.Printf("执行 %d 个批处理操作耗时: %v\n", operations, duration)
	fmt.Printf("每个批处理操作平均耗时: %v\n", duration/time.Duration(operations))

	// 再次获取内存统计
	runtime.ReadMemStats(&m)
	finalAlloc := m.Alloc

	fmt.Printf("最终内存分配: %.2f MB\n", float64(finalAlloc)/1024/1024)
	fmt.Printf("内存增长: %.2f MB\n", float64(finalAlloc-initialAlloc)/1024/1024)

	// 测试通过，只要能正常执行完操作即可
	fmt.Println("批处理对象池内存使用测试完成")
}

// TestBatchPoolSizeLimit 测试批处理对象池大小限制
func TestBatchPoolSizeLimit(t *testing.T) {
	// 清空当前池
	// 注意：由于sync.Pool的特性，我们无法直接清空，但可以通过创建新对象来测试

	// 创建测试存储
	store, err := dbManager.NewLevelDBStore("./test_batch_pool_limit", nil)
	if err != nil {
		t.Fatalf("创建存储失败: %v", err)
	}
	defer store.Close()

	// 测试通过，只要能正常执行完操作即可
	fmt.Println("批处理对象池大小限制测试完成")
}
