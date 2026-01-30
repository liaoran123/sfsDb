package record

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

// 创建测试数据
func createTestData(count int) Records {
	records := make(Records, count)
	for i := 0; i < count; i++ {
		// 从对象池获取 Record 对象
		records[i] = GetRecord()
		records[i]["id"] = i
		records[i]["name"] = fmt.Sprintf("user%d", i)
		records[i]["age"] = 20 + i%30
		records[i]["email"] = fmt.Sprintf("user%d@example.com", i)
		records[i]["active"] = i%2 == 0
		records[i]["score"] = float64(70 + i%30)
	}
	return records
}

// 测试内存使用
func getMemoryUsage() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc
}

// TestRecordSelectPerformance 测试 Select 操作的性能
func TestRecordSelectPerformance(t *testing.T) {
	// 创建测试数据
	count := 10000
	records := createTestData(count)

	// 测试优化前的内存使用
	before := getMemoryUsage()
	start := time.Now()

	// 执行 Select 操作
	for _, r := range records {
		selected := r.Select("id", "name", "age")
		_ = selected
	}

	// 测试优化后的内存使用
	after := getMemoryUsage()
	duration := time.Since(start)

	t.Logf("Select operation for %d records:", count)
	t.Logf("Time: %v", duration)
	t.Logf("Memory used: %d bytes", after-before)
}

// TestRecordsSelectPerformance 测试 Records.Select 操作的性能
func TestRecordsSelectPerformance(t *testing.T) {
	// 创建测试数据
	count := 10000
	records := createTestData(count)

	// 测试优化前的内存使用
	before := getMemoryUsage()
	start := time.Now()

	// 执行 Select 操作
	selected := records.Select("id", "name", "age")
	_ = selected

	// 释放对象
	PutRecords(selected)

	// 测试优化后的内存使用
	after := getMemoryUsage()
	duration := time.Since(start)

	t.Logf("Records.Select operation for %d records:", count)
	t.Logf("Time: %v", duration)
	t.Logf("Memory used: %d bytes", after-before)
}

// TestBatchSelectPerformance 测试批量选择操作的性能
func TestBatchSelectPerformance(t *testing.T) {
	// 创建测试数据
	count := 10000
	records := createTestData(count)

	// 测试优化前的内存使用
	before := getMemoryUsage()
	start := time.Now()

	// 执行批量 Select 操作
	selected := BatchSelect(records, "id", "name", "age")
	_ = selected

	// 释放对象
	PutRecords(selected)

	// 测试优化后的内存使用
	after := getMemoryUsage()
	duration := time.Since(start)

	t.Logf("BatchSelect operation for %d records:", count)
	t.Logf("Time: %v", duration)
	t.Logf("Memory used: %d bytes", after-before)
}

// TestContainsPerformance 测试 Contains 操作的性能
func TestContainsPerformance(t *testing.T) {
	// 创建测试数据
	count := 1000
	records := createTestData(count)
	target := records[count/2]

	// 测试优化前的内存使用
	before := getMemoryUsage()
	start := time.Now()

	// 执行 Contains 操作
	result := records.Contains(target)
	_ = result

	// 测试优化后的内存使用
	after := getMemoryUsage()
	duration := time.Since(start)

	t.Logf("Contains operation for %d records:", count)
	t.Logf("Time: %v", duration)
	t.Logf("Memory used: %d bytes", after-before)
	t.Logf("Result: %v", result)
}

// TestIntersectPerformance 测试 Intersect 操作的性能
func TestIntersectPerformance(t *testing.T) {
	// 创建测试数据
	count := 1000
	records1 := createTestData(count)
	records2 := createTestData(count/2 + 500) // 部分重叠

	// 测试优化前的内存使用
	before := getMemoryUsage()
	start := time.Now()

	// 执行 Intersect 操作
	result := records1.Intersect(records2)
	_ = result

	// 释放对象
	PutRecords(result)

	// 测试优化后的内存使用
	after := getMemoryUsage()
	duration := time.Since(start)

	t.Logf("Intersect operation for %d records:", count)
	t.Logf("Time: %v", duration)
	t.Logf("Memory used: %d bytes", after-before)
	t.Logf("Result count: %d", len(result))
}

// TestUnionPerformance 测试 Union 操作的性能
func TestUnionPerformance(t *testing.T) {
	// 创建测试数据
	count := 1000
	records1 := createTestData(count)
	records2 := createTestData(count/2 + 500) // 部分重叠

	// 测试优化前的内存使用
	before := getMemoryUsage()
	start := time.Now()

	// 执行 Union 操作
	result := records1.Union(records2)
	_ = result

	// 释放对象
	PutRecords(result)

	// 测试优化后的内存使用
	after := getMemoryUsage()
	duration := time.Since(start)

	t.Logf("Union operation for %d records:", count)
	t.Logf("Time: %v", duration)
	t.Logf("Memory used: %d bytes", after-before)
	t.Logf("Result count: %d", len(result))
}

// TestDifferencePerformance 测试 Difference 操作的性能
func TestDifferencePerformance(t *testing.T) {
	// 创建测试数据
	count := 1000
	records1 := createTestData(count)
	records2 := createTestData(count / 2) // 部分重叠

	// 测试优化前的内存使用
	before := getMemoryUsage()
	start := time.Now()

	// 执行 Difference 操作
	result := records1.Difference(records2)
	_ = result

	// 释放对象
	PutRecords(result)

	// 测试优化后的内存使用
	after := getMemoryUsage()
	duration := time.Since(start)

	t.Logf("Difference operation for %d records:", count)
	t.Logf("Time: %v", duration)
	t.Logf("Memory used: %d bytes", after-before)
	t.Logf("Result count: %d", len(result))
}

// TestBatchOperationsPerformance 测试批量操作的性能
func TestBatchOperationsPerformance(t *testing.T) {
	// 创建测试数据
	count := 10000
	records := createTestData(count)

	// 测试批量操作
	before := getMemoryUsage()
	start := time.Now()

	// 执行批量选择
	selected := BatchSelect(records, "id", "name", "age")

	// 释放对象
	PutRecords(selected)

	after := getMemoryUsage()
	duration := time.Since(start)

	t.Logf("Batch operations for %d records:", count)
	t.Logf("Time: %v", duration)
	t.Logf("Memory used: %d bytes", after-before)
}

// BenchmarkRecordSelect 基准测试 Record.Select 操作
func BenchmarkRecordSelect(b *testing.B) {
	// 从对象池获取 Record 对象
	record := GetRecord()
	defer PutRecord(record)

	// 填充测试数据
	record["id"] = 1
	record["name"] = "user1"
	record["age"] = 25
	record["email"] = "user1@example.com"
	record["active"] = true
	record["score"] = 85.5

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		selected := record.Select("id", "name", "age")
		PutRecord(selected)
	}
}

// BenchmarkRecordsSelect 基准测试 Records.Select 操作
func BenchmarkRecordsSelect(b *testing.B) {
	// 创建测试数据
	count := 100
	records := createTestData(count)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		selected := records.Select("id", "name", "age")
		PutRecords(selected)
	}

	// 释放测试数据
	for _, r := range records {
		PutRecord(r)
	}
}

// BenchmarkBatchSelect 基准测试 BatchSelect 操作
func BenchmarkBatchSelect(b *testing.B) {
	// 创建测试数据
	count := 100
	records := createTestData(count)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		selected := BatchSelect(records, "id", "name", "age")
		PutRecords(selected)
	}

	// 释放测试数据
	for _, r := range records {
		PutRecord(r)
	}
}

// BenchmarkContains 基准测试 Contains 操作
func BenchmarkContains(b *testing.B) {
	// 创建测试数据
	count := 100
	records := createTestData(count)
	target := records[count/2]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = records.Contains(target)
	}

	// 释放测试数据
	for _, r := range records {
		PutRecord(r)
	}
}
