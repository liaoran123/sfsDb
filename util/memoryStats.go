package util

import "runtime"

// MemoryStats 内存使用统计信息结构体
type MemoryStats struct {
	Alloc      uint64 // 当前分配的内存大小（字节）
	TotalAlloc uint64 // 累计分配的内存大小（字节）
	Sys        uint64 // 从系统获取的内存大小（字节）
	NumGC      uint32 // GC 次数
}

// GetMemoryStats 获取当前内存使用情况
func GetMemoryStats() MemoryStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return MemoryStats{
		Alloc:      m.Alloc,
		TotalAlloc: m.TotalAlloc,
		Sys:        m.Sys,
		NumGC:      m.NumGC,
	}
}

//使用方法

/*
// TestCachePerformance 测试缓存系统性能
func TestCachePerformance() {


	// 获取初始内存状态
	initialMem := GetMemoryStats()
	fmt.Printf("初始内存状态: Alloc=%.2fMB, TotalAlloc=%.2fMB, Sys=%.2fMB, NumGC=%d\n",
		float64(initialMem.Alloc)/1024/1024,
		float64(initialMem.TotalAlloc)/1024/1024,
		float64(initialMem.Sys)/1024/1024,
		initialMem.NumGC)

	// 模拟大量缓存操作
	const iterations = 10000
	const concurrentGoroutines = 10

	var wg sync.WaitGroup
	wg.Add(concurrentGoroutines)

	startTime := time.Now()

	for i := 0; i < concurrentGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < iterations; j++ {
				// 测试索引匹配缓存
				fields := []string{fmt.Sprintf("field%d", j%100), fmt.Sprintf("field%d", (j+1)%100)}
				// 注意：这里需要一个实际的 Table 实例，这里只是示例
				// 实际使用时，应该传入一个有效的 Table 实例
				// t.MatchIndexCached(fields)

				// 测试字段转换缓存
				fieldsMap := map[string]any{
					fmt.Sprintf("key%d", j%100):     fmt.Sprintf("value%d", j),
					fmt.Sprintf("key%d", (j+1)%100): j,
				}
				// 注意：这里需要一个实际的 Table 实例，这里只是示例
				// 实际使用时，应该传入一个有效的 Table 实例
				// t.FieldsToBytesNilCached(&fieldsMap)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	// 获取最终内存状态
	finalMem := GetMemoryStats()
	fmt.Printf("最终内存状态: Alloc=%.2fMB, TotalAlloc=%.2fMB, Sys=%.2fMB, NumGC=%d\n",
		float64(finalMem.Alloc)/1024/1024,
		float64(finalMem.TotalAlloc)/1024/1024,
		float64(finalMem.Sys)/1024/1024,
		finalMem.NumGC)

	// 计算内存变化
	memIncrease := float64(finalMem.Alloc-initialMem.Alloc) / 1024 / 1024
	gcIncrease := finalMem.NumGC - initialMem.NumGC


	fmt.Printf("性能测试结果:\n")
	fmt.Printf("执行时间: %v\n", duration)
	fmt.Printf("内存增加: %.2fMB\n", memIncrease)
	fmt.Printf("GC 次数增加: %d\n", gcIncrease)

}

*/
