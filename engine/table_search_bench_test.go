package engine

import (
	"testing"
)

// 基准测试 Table.Search 函数的性能（主键搜索）
func BenchmarkTableSearchPrimaryKey(b *testing.B) {
	// 创建测试表
	table, err := TableNew("benchmark_search_pk")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{"id": 0, "name": "", "age": 0, "email": ""}
	err = table.SetFields(fields)
	if err != nil {
		b.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键索引
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("id")
	table.CreateIndex(pk)

	// 插入测试数据
	for i := 0; i < 1000; i++ {
		testData := map[string]any{
			"id":    i + 1,
			"name":  "User" + string(rune('A'+i%26)),
			"age":   20 + i%30,
			"email": "user" + string(rune('A'+i%26)) + "@example.com",
		}
		_, err := table.Insert(&testData)
		if err != nil {
			b.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// 重置计时器
	b.ResetTimer()

	// 运行基准测试
	for i := 0; i < b.N; i++ {
		// 搜索主键
		searchData := map[string]any{"id": (i % 1000) + 1}
		iter := table.Search(&searchData)
		if iter != nil {
			defer GlobalTableIterPool.Put(iter)
		}
	}
}

// 基准测试 Table.Search 函数的性能（普通索引搜索）
func BenchmarkTableSearchIndex(b *testing.B) {
	// 创建测试表
	table, err := TableNew("benchmark_search_index")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{"id": 0, "name": "", "age": 0, "email": ""}
	err = table.SetFields(fields)
	if err != nil {
		b.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键索引
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("id")
	table.CreateIndex(pk)

	// 创建普通索引
	idx, _ := DefaultNormalIndexNew("name_index")
	idx.AddFields("name")
	table.CreateIndex(idx)

	// 插入测试数据
	for i := 0; i < 1000; i++ {
		testData := map[string]any{
			"id":    i + 1,
			"name":  "User" + string(rune('A'+i%26)),
			"age":   20 + i%30,
			"email": "user" + string(rune('A'+i%26)) + "@example.com",
		}
		_, err := table.Insert(&testData)
		if err != nil {
			b.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// 重置计时器
	b.ResetTimer()

	// 运行基准测试
	for i := 0; i < b.N; i++ {
		// 搜索普通索引
		searchData := map[string]any{"name": "User" + string(rune('A'+(i%26)))}
		iter := table.Search(&searchData)
		if iter != nil {
			defer GlobalTableIterPool.Put(iter)
		}
	}
}

// 基准测试 Table.Search 函数的性能（全表扫描）
func BenchmarkTableSearchFullScan(b *testing.B) {
	// 创建测试表
	table, err := TableNew("benchmark_search_fullscan")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{"id": 0, "name": "", "age": 0, "email": ""}
	err = table.SetFields(fields)
	if err != nil {
		b.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键索引
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("id")
	table.CreateIndex(pk)

	// 插入测试数据
	for i := range 1000 {
		testData := map[string]any{
			"id":    i + 1,
			"name":  "User" + string(rune('A'+i%26)),
			"age":   20 + i%30,
			"email": "user" + string(rune('A'+i%26)) + "@example.com",
		}
		_, err := table.Insert(&testData)
		if err != nil {
			b.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// 重置计时器
	b.ResetTimer()

	// 运行基准测试
	for i := 0; i < b.N; i++ {
		// 全表扫描（搜索不存在的字段，触发全表扫描）
		searchData := map[string]any{"non_existent_field": "value"}
		iter := table.Search(&searchData)
		if iter != nil {
			defer GlobalTableIterPool.Put(iter)
		}
	}
}
