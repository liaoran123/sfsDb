package record

import (
	"runtime"
	"testing"
	"time"
)

// TestGlobalPoolStats 测试全局对象池统计信息
func TestGlobalPoolStats(t *testing.T) {
	// 重置全局统计信息
	ResetGlobalPoolStats()
	
	// 记录初始全局统计信息
	initialStats := GlobalPoolStats()
	t.Logf("初始全局统计信息:")
	t.Logf("Record - 创建: %d, 获取: %d, 放回: %d", initialStats["globalRecordCreated"], initialStats["globalRecordGet"], initialStats["globalRecordPut"])
	t.Logf("Records - 创建: %d, 获取: %d, 放回: %d", initialStats["globalRecordsCreated"], initialStats["globalRecordsGet"], initialStats["globalRecordsPut"])
	
	// 执行一些对象池操作
	for i := 0; i < 10; i++ {
		r := GetRecord()
		r["key"] = "value"
		PutRecord(r)
	}
	
	for i := 0; i < 5; i++ {
		rs := GetRecords()
		for j := 0; j < 3; j++ {
			r := GetRecord()
			r["key"] = "value"
			rs = append(rs, r)
		}
		PutRecords(rs)
	}
	
	// 执行一些不手动释放的操作，测试finalizer
	for i := 0; i < 5; i++ {
		r := GetRecord()
		r["key"] = "value"
		// 不调用 PutRecord(r)，让finalizer自动释放
	}
	
	// 强制垃圾回收，触发finalizer
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	
	// 获取全局统计信息
	globalStats := GlobalPoolStats()
	t.Logf("操作后全局统计信息:")
	t.Logf("Record - 创建: %d, 获取: %d, 放回: %d", globalStats["globalRecordCreated"], globalStats["globalRecordGet"], globalStats["globalRecordPut"])
	t.Logf("Records - 创建: %d, 获取: %d, 放回: %d", globalStats["globalRecordsCreated"], globalStats["globalRecordsGet"], globalStats["globalRecordsPut"])
	
	// 验证统计信息是否合理
	if globalStats["globalRecordGet"] == 0 {
		t.Fatalf("全局Record获取数量应为正数，实际为0")
	}
	if globalStats["globalRecordsGet"] == 0 {
		t.Fatalf("全局Records获取数量应为正数，实际为0")
	}
	if globalStats["globalRecordPut"] == 0 {
		t.Fatalf("全局Record放回数量应为正数，实际为0")
	}
	if globalStats["globalRecordsPut"] == 0 {
		t.Fatalf("全局Records放回数量应为正数，实际为0")
	}
	
	// 验证获取和放回的数量是否合理
	totalRecordGet := globalStats["globalRecordGet"]
	totalRecordPut := globalStats["globalRecordPut"]
	if totalRecordGet < totalRecordPut {
		t.Logf("警告: Record获取数量小于放回数量 - 获取: %d, 放回: %d", totalRecordGet, totalRecordPut)
	}
	
	totalRecordsGet := globalStats["globalRecordsGet"]
	totalRecordsPut := globalStats["globalRecordsPut"]
	if totalRecordsGet < totalRecordsPut {
		t.Logf("警告: Records获取数量小于放回数量 - 获取: %d, 放回: %d", totalRecordsGet, totalRecordsPut)
	}
	
	t.Logf("全局跟踪机制测试通过")
}

// TestGlobalPoolStatsReset 测试重置全局统计信息
func TestGlobalPoolStatsReset(t *testing.T) {
	// 执行一些对象池操作
	for i := 0; i < 5; i++ {
		r := GetRecord()
		r["key"] = "value"
		PutRecord(r)
	}
	
	// 获取重置前的统计信息
	beforeReset := GlobalPoolStats()
	t.Logf("重置前全局统计信息:")
	t.Logf("Record - 创建: %d, 获取: %d, 放回: %d", beforeReset["globalRecordCreated"], beforeReset["globalRecordGet"], beforeReset["globalRecordPut"])
	
	// 重置全局统计信息
	ResetGlobalPoolStats()
	
	// 获取重置后的统计信息
	afterReset := GlobalPoolStats()
	t.Logf("重置后全局统计信息:")
	t.Logf("Record - 创建: %d, 获取: %d, 放回: %d", afterReset["globalRecordCreated"], afterReset["globalRecordGet"], afterReset["globalRecordPut"])
	
	// 验证重置是否成功
	if afterReset["globalRecordCreated"] != 0 {
		t.Fatalf("重置后globalRecordCreated应为0，实际为%d", afterReset["globalRecordCreated"])
	}
	if afterReset["globalRecordGet"] != 0 {
		t.Fatalf("重置后globalRecordGet应为0，实际为%d", afterReset["globalRecordGet"])
	}
	if afterReset["globalRecordPut"] != 0 {
		t.Fatalf("重置后globalRecordPut应为0，实际为%d", afterReset["globalRecordPut"])
	}
	if afterReset["globalRecordsCreated"] != 0 {
		t.Fatalf("重置后globalRecordsCreated应为0，实际为%d", afterReset["globalRecordsCreated"])
	}
	if afterReset["globalRecordsGet"] != 0 {
		t.Fatalf("重置后globalRecordsGet应为0，实际为%d", afterReset["globalRecordsGet"])
	}
	if afterReset["globalRecordsPut"] != 0 {
		t.Fatalf("重置后globalRecordsPut应为0，实际为%d", afterReset["globalRecordsPut"])
	}
	
	t.Logf("全局统计信息重置测试通过")
}
