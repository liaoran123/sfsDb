package engine

import (
	"testing"
)

// BenchmarkTableInsertWithACID 测试ACID事务模式下的插入性能
func BenchmarkTableInsertWithACID(b *testing.B) {
	// 创建测试表
	table, err := TableNew("benchmark_table_acid")
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

	// 重置计时器
	b.ResetTimer()

	// 运行基准测试
	for i := 0; i < b.N; i++ {
		// 开始事务
		tx, err := table.Begin()
		if err != nil {
			b.Fatalf("Failed to begin transaction: %v", err)
		}

		// 创建测试数据
		testData := map[string]any{
			"id":    i + 1,
			"name":  "User" + string(rune('A'+i%26)),
			"age":   20 + i%30,
			"email": "user" + string(rune('A'+i%26)) + "@example.com",
		}

		// 在事务中插入数据
		_, err = tx.Insert(&testData)
		if err != nil {
			b.Fatalf("Failed to insert data: %v", err)
		}

		// 提交事务
		err = tx.Commit()
		if err != nil {
			b.Fatalf("Failed to commit transaction: %v", err)
		}
	}
}

// BenchmarkTableSearchWithACID 测试ACID事务模式下的搜索性能
func BenchmarkTableSearchWithACID(b *testing.B) {
	// 创建测试表
	table, err := TableNew("benchmark_table_acid_search")
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
		// 开始事务
		tx, err := table.Begin()
		if err != nil {
			b.Fatalf("Failed to begin transaction: %v", err)
		}

		// 搜索条件
		searchData := map[string]any{"id": (i % 1000) + 1}

		// 在事务中搜索数据
		iter := tx.Search(&searchData)
		if iter == nil {
			b.Fatalf("Failed to create search iterator")
		}

		// 使用GetRecords方法遍历结果
		records := iter.GetRecords(true, 1) // 只获取1条记录
		_ = records
		iter.Release()

		// 提交事务
		err = tx.Commit()
		if err != nil {
			b.Fatalf("Failed to commit transaction: %v", err)
		}
	}
}
