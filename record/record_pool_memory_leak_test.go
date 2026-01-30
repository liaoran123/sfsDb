package record

import (
	"runtime"
	"testing"
	"time"
)

// TestPutRecordsMemoryLeak 测试PutRecords函数是否存在内存泄漏
func TestPutRecordsMemoryLeak(t *testing.T) {
	// 运行多次循环，创建和销毁大量Records对象
	iterations := 10000
	batchSize := 100

	// 运行前获取内存使用情况
	var memStatsBefore runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)

	// 开始测试
	startTime := time.Now()
	for i := 0; i < iterations; i++ {
		// 创建一个Records对象并添加多个Record
		rs := GetRecords()
		for j := 0; j < batchSize; j++ {
			r := GetRecord()
			r["id"] = j
			r["name"] = "test"
			r["age"] = 25
			rs = append(rs, r)
		}
		// 放回对象池
		PutRecords(rs)

		// 每1000次迭代打印一次进度
		if (i+1)%1000 == 0 {
			t.Logf("完成 %d/%d 次迭代，耗时: %v", i+1, iterations, time.Since(startTime))
		}

		// 每1000次迭代进行一次垃圾回收
		if (i+1)%1000 == 0 {
			runtime.GC()
		}
	}

	// 运行后获取内存使用情况
	var memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsAfter)

	// 打印内存使用情况
	t.Logf("内存使用情况:")
	t.Logf("运行前分配的内存: %d MB", memStatsBefore.Alloc/1024/1024)
	t.Logf("运行后分配的内存: %d MB", memStatsAfter.Alloc/1024/1024)
	t.Logf("内存增长: %d MB", (memStatsAfter.Alloc-memStatsBefore.Alloc)/1024/1024)
	t.Logf("运行前GC次数: %d", memStatsBefore.NumGC)
	t.Logf("运行后GC次数: %d", memStatsAfter.NumGC)

	// 检查内存增长是否合理
	// 对于10000次迭代，每次100个Record，内存增长应该很小
	maxAllowedGrowth := uint64(50) // 允许最多增长50MB
	actualGrowth := (memStatsAfter.Alloc - memStatsBefore.Alloc) / 1024 / 1024
	if actualGrowth > maxAllowedGrowth {
		t.Errorf("内存增长过大，可能存在内存泄漏: %d MB > %d MB", actualGrowth, maxAllowedGrowth)
	} else {
		t.Logf("内存增长在合理范围内: %d MB", actualGrowth)
	}

	t.Logf("测试完成，总耗时: %v", time.Since(startTime))
}
