package engine

import (
	"testing"
)

// BenchmarkFormatRecord 测试 FormatRecord 方法的性能
func BenchmarkFormatRecord(b *testing.B) {
	// 创建测试表
	table, err := TableNew("test_table")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// 设置字段
	err = table.SetFields(map[string]any{
		"id":    0,
		"name":  "",
		"age":   0,
		"email": "",
	})
	if err != nil {
		b.Fatalf("Failed to set fields: %v", err)
	}

	// 准备测试数据
	testData := map[string][]byte{
		"id":    []byte("1"),
		"name":  []byte("John Doe"),
		"age":   []byte("30"),
		"email": []byte("john@example.com"),
	}

	// 重置计时器
	b.ResetTimer()

	// 运行基准测试
	for i := 0; i < b.N; i++ {
		table.FormatRecord(&testData)
	}
}

// BenchmarkFormatRecordBatch 测试批量格式化的性能
func BenchmarkFormatRecordBatch(b *testing.B) {
	// 创建测试表
	table, err := TableNew("test_table")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// 设置字段
	err = table.SetFields(map[string]any{
		"id":    0,
		"name":  "",
		"age":   0,
		"email": "",
	})
	if err != nil {
		b.Fatalf("Failed to set fields: %v", err)
	}

	// 准备测试数据
	batchSize := 100
	testDataBatch := make([]*map[string][]byte, batchSize)
	for i := 0; i < batchSize; i++ {
		testData := map[string][]byte{
			"id":    []byte(string(rune('0' + i%10))),
			"name":  []byte("John Doe " + string(rune('A'+i%26))),
			"age":   []byte(string(rune('0'+i%10)) + string(rune('0'+i%10))),
			"email": []byte("john" + string(rune('a'+i%26)) + "@example.com"),
		}
		testDataBatch[i] = &testData
	}

	// 重置计时器
	b.ResetTimer()

	// 运行基准测试
	for i := 0; i < b.N; i++ {
		for _, data := range testDataBatch {
			table.FormatRecord(data)
		}
	}
}
