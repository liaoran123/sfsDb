package engine

import (
	"testing"

	"github.com/liaoran123/sfsDb/record"
	"github.com/liaoran123/sfsDb/storage"
)

// BenchmarkSimpleOperations 简单基准测试，确保资源正确释放
func BenchmarkSimpleOperations(b *testing.B) {
	// 初始化数据库
	_, err := storage.GetDBManager().OpenDB("./benchmark_simple_db")
	if err != nil {
		b.Fatalf("Failed to open database: %v", err)
	}
	defer storage.CloseDb()

	// 创建表
	table, err := TableNew("benchmark_simple_table")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// 设置字段
	fields := map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	}
	err = table.SetFields(fields)
	if err != nil {
		b.Fatalf("Failed to set fields: %v", err)
	}

	// 基准测试插入和搜索操作，确保资源正确释放
	b.Run("InsertAndSearch", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			// 插入操作
			user := map[string]any{
				"name": "Test",
				"age":  30,
			}
			_, err := table.Insert(&user)
			if err != nil {
				b.Fatalf("Failed to insert: %v", err)
			}

			// 搜索操作，使用后立即释放资源
			searchData := map[string]any{"id": 1}
			iter, err := table.Search(&searchData)
			if err != nil {
				b.Fatalf("Search 失败: %v", err)
			}
			if iter == nil {
				b.Fatalf("Failed to get iterator: %v", err)
			}
			if iter != nil {
				// 立即使用并释放，而不是使用defer
				records := iter.GetRecords(true)
				if records != nil {
					record.PutRecords(records)
				}
				GlobalTableIterPool.Put(iter)
			}
		}
	})
}
