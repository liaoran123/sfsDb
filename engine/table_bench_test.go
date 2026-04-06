package engine

import (
	"testing"
)

// 基准测试 Table.Insert 函数的性能
func BenchmarkTableInsert(b *testing.B) {
	// 创建测试表
	table, err := NewTable("benchmark_table")
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
	pk, _ := NewDefaultPrimaryKey("pk")
	pk.AddFields("id")
	table.CreateIndex(pk)

	// 重置计时器
	b.ResetTimer()

	// 运行基准测试
	for i := 0; i < b.N; i++ {
		// 创建测试数据
		testData := map[string]any{
			"id":    i + 1,
			"name":  "User" + string(rune('A'+i%26)),
			"age":   20 + i%30,
			"email": "user" + string(rune('A'+i%26)) + "@example.com",
		}

		// 插入数据
		_, err := table.Insert(&testData)
		if err != nil {
			b.Fatalf("Failed to insert data: %v", err)
		}
	}
}

// 基准测试 Table.Insert 函数的性能（使用批量操作）
func BenchmarkTableInsertBatch(b *testing.B) {
	// 创建测试表
	table, err := TableNew("benchmark_table_batch")
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
	pk, _ := NewDefaultPrimaryKey("pk")
	pk.AddFields("id")
	table.CreateIndex(pk)

	// 重置计时器
	b.ResetTimer()

	// 运行基准测试
	for i := 0; i < b.N; i++ {
		// 创建批量操作
		batch := table.kvStore.GetBatch()

		// 插入多条数据
		for j := 0; j < 10; j++ {
			// 创建测试数据
			testData := map[string]any{
				"id":    i*10 + j + 1,
				"name":  "User" + string(rune('A'+(i*10+j)%26)),
				"age":   20 + (i*10+j)%30,
				"email": "user" + string(rune('A'+(i*10+j)%26)) + "@example.com",
			}

			// 插入数据（使用批量操作）
			_, err := table.Insert(&testData, batch)
			if err != nil {
				b.Fatalf("Failed to insert data: %v", err)
			}
		}

		// 提交批量操作
		err = table.kvStore.WriteBatch(batch)
		if err != nil {
			b.Fatalf("Failed to write batch: %v", err)
		}
	}
}
