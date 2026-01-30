package engine

import (
	"testing"

	"github.com/liaoran123/sfsDb/util"
)

// 模拟 FieldsToBytes 方法（不使用对象池）
func (t *Table) FieldsToBytesNoPool(fields *map[string]any) map[string][]byte {
	// 每次创建新的 map
	result := make(map[string][]byte)
	for k, v := range *fields {
		result[k] = util.AnyToBytes(v)
	}
	return result
}

// 基准测试：不使用对象池
func BenchmarkFieldsToBytesNoPool(b *testing.B) {
	// 创建测试表
	table, err := TableNew("test_bench")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// 设置字段
	fields := map[string]any{
		"id":    0,
		"name":  "",
		"age":   0,
		"email": "",
	}
	err = table.SetFields(fields)
	if err != nil {
		b.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("id")
	err = table.CreateIndex(pk)
	if err != nil {
		b.Fatalf("Failed to create primary key: %v", err)
	}

	// 测试数据
	testFields := map[string]any{
		"id":    1,
		"name":  "test",
		"age":   25,
		"email": "test@example.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = table.FieldsToBytesNoPool(&testFields)
	}
}

// 基准测试：使用对象池
func BenchmarkFieldsToBytesWithPool(b *testing.B) {
	// 创建测试表
	table, err := TableNew("test_bench")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// 设置字段
	fields := map[string]any{
		"id":    0,
		"name":  "",
		"age":   0,
		"email": "",
	}
	err = table.SetFields(fields)
	if err != nil {
		b.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("id")
	err = table.CreateIndex(pk)
	if err != nil {
		b.Fatalf("Failed to create primary key: %v", err)
	}

	// 测试数据
	testFields := map[string]any{
		"id":    1,
		"name":  "test",
		"age":   25,
		"email": "test@example.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = table.FieldsToBytes(&testFields)
	}
}

// 基准测试：批量操作（不使用对象池）
func BenchmarkBatchFieldsToBytesNoPool(b *testing.B) {
	// 创建测试表
	table, err := TableNew("test_bench")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// 设置字段
	fields := map[string]any{
		"id":    0,
		"name":  "",
		"age":   0,
		"email": "",
	}
	err = table.SetFields(fields)
	if err != nil {
		b.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("id")
	err = table.CreateIndex(pk)
	if err != nil {
		b.Fatalf("Failed to create primary key: %v", err)
	}

	// 测试数据
	testFieldsList := make([]map[string]any, 100)
	for i := range testFieldsList {
		testFieldsList[i] = map[string]any{
			"id":    i,
			"name":  "test",
			"age":   25,
			"email": "test@example.com",
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, fields := range testFieldsList {
			_ = table.FieldsToBytesNoPool(&fields)
		}
	}
}

// 基准测试：批量操作（使用对象池）
func BenchmarkBatchFieldsToBytesWithPool(b *testing.B) {
	// 创建测试表
	table, err := TableNew("test_bench")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// 设置字段
	fields := map[string]any{
		"id":    0,
		"name":  "",
		"age":   0,
		"email": "",
	}
	err = table.SetFields(fields)
	if err != nil {
		b.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("id")
	err = table.CreateIndex(pk)
	if err != nil {
		b.Fatalf("Failed to create primary key: %v", err)
	}

	// 测试数据
	testFieldsList := make([]map[string]any, 100)
	for i := range testFieldsList {
		testFieldsList[i] = map[string]any{
			"id":    i,
			"name":  "test",
			"age":   25,
			"email": "test@example.com",
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, fields := range testFieldsList {
			_ = table.FieldsToBytes(&fields)
		}
	}
}
