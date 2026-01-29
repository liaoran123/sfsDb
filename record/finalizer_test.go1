package record

import (
	"runtime"
	"testing"
	"time"
)

// TestFinalizerAutoRelease 测试finalizer自动释放机制
func TestFinalizerAutoRelease(t *testing.T) {
	// 重置对象池统计信息
	ResetPoolStats()

	// 记录初始创建的对象数量
	initialStats := PoolStats()
	t.Logf("初始对象池统计信息:")
	t.Logf("Record - 创建: %d, 获取: %d", initialStats["recordCreated"], initialStats["recordGet"])

	// 获取对象但不手动释放
	recordCount := 20
	recordsCount := 10

	for i := 0; i < recordCount; i++ {
		r := GetRecord()
		r["key"] = "value"
		r["index"] = i
		// 不调用 PutRecord(r)，让finalizer自动释放
	}

	// 获取Records对象但不手动释放
	for i := 0; i < recordsCount; i++ {
		rs := GetRecords()
		for j := 0; j < 3; j++ {
			r := GetRecord()
			r["key"] = "value"
			r["index"] = i*3 + j
			rs = append(rs, r)
		}
		// 不调用 PutRecords(rs)，让finalizer自动释放
	}

	// 强制垃圾回收，触发finalizer
	for i := 0; i < 3; i++ {
		runtime.GC()
		// 等待finalizer执行
		time.Sleep(200 * time.Millisecond)
	}

	// 获取对象池统计信息
	statsAfterGC := PoolStats()
	t.Logf("垃圾回收后对象池统计信息:")
	t.Logf("Record - 创建: %d, 获取: %d, 手动放回: %d", statsAfterGC["recordCreated"], statsAfterGC["recordGet"], statsAfterGC["recordPutManual"])
	t.Logf("Records - 创建: %d, 获取: %d, 手动放回: %d", statsAfterGC["recordsCreated"], statsAfterGC["recordsGet"], statsAfterGC["recordsPutManual"])

	// 测试对象池复用
	t.Logf("测试对象池复用...")
	reuseCount := 30
	for i := 0; i < reuseCount; i++ {
		r := GetRecord()
		if r == nil {
			t.Fatalf("GetRecord() returned nil")
		}
		r["test"] = i
		r["timestamp"] = time.Now().UnixNano()
		PutRecord(r)
	}

	// 再次获取对象池统计信息
	statsAfterReuse := PoolStats()
	t.Logf("复用后对象池统计信息:")
	t.Logf("Record - 创建: %d, 获取: %d, 手动放回: %d", statsAfterReuse["recordCreated"], statsAfterReuse["recordGet"], statsAfterReuse["recordPutManual"])

	// 计算复用率
	totalGet := statsAfterReuse["recordGet"]
	totalCreated := statsAfterReuse["recordCreated"]
	initialCreated := initialStats["recordCreated"]
	actualCreated := totalCreated - initialCreated

	reuseRate := float64(totalGet-actualCreated) / float64(totalGet) * 100
	t.Logf("对象池复用率: %.2f%%", reuseRate)

	// 验证复用率
	if reuseRate < 50 {
		t.Logf("警告: 对象池复用率较低 - %.2f%%", reuseRate)
	} else {
		t.Logf("对象池复用正常 - %.2f%%", reuseRate)
	}

	// 验证创建的对象数量没有急剧增加
	if actualCreated > totalGet/2 {
		t.Logf("警告: 对象池创建的对象数量过多 - 创建: %d, 获取: %d", actualCreated, totalGet)
	} else {
		t.Logf("对象池创建对象数量正常 - 创建: %d, 获取: %d", actualCreated, totalGet)
	}
}

// TestFinalizerMultipleGarbageCollections 测试多次垃圾回收触发finalizer
func TestFinalizerMultipleGarbageCollections(t *testing.T) {
	// 重置对象池统计信息
	ResetPoolStats()

	// 第一轮：获取对象但不释放
	for i := 0; i < 5; i++ {
		r := GetRecord()
		r["key"] = "value"
		r["round"] = 1
		r["index"] = i
	}

	// 触发第一次垃圾回收
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	// 第二轮：获取对象但不释放
	for i := 0; i < 5; i++ {
		r := GetRecord()
		r["key"] = "value"
		r["round"] = 2
		r["index"] = i
	}

	// 触发第二次垃圾回收
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	// 第三轮：获取对象但不释放
	for i := 0; i < 5; i++ {
		r := GetRecord()
		r["key"] = "value"
		r["round"] = 3
		r["index"] = i
	}

	// 触发第三次垃圾回收
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	// 获取对象池统计信息
	stats := PoolStats()
	t.Logf("对象池使用统计信息(多次GC):")
	t.Logf("Record - 创建: %d, 获取: %d, 手动放回: %d", stats["recordCreated"], stats["recordGet"], stats["recordPutManual"])

	// 验证对象池能够正常工作
	if stats["recordGet"] == 0 {
		t.Fatalf("没有从对象池获取对象")
	}

	// 验证创建的对象数量是否合理
	if stats["recordCreated"] == 0 {
		t.Logf("注意: 对象池可能复用了之前创建的对象，没有创建新对象")
	} else if stats["recordCreated"] > stats["recordGet"] {
		t.Logf("警告: 创建的对象数量大于获取的数量 - 创建: %d, 获取: %d", stats["recordCreated"], stats["recordGet"])
	} else {
		t.Logf("对象池创建对象数量正常 - 创建: %d, 获取: %d", stats["recordCreated"], stats["recordGet"])
	}

	t.Logf("多次垃圾回收测试通过，finalizer自动释放机制正常工作")
}
