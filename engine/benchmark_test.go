package engine

import (
	"testing"

	"github.com/liaoran123/sfsDb/record"
	"github.com/liaoran123/sfsDb/storage"
)

// BenchmarkTableOperations 基准测试表操作性能
func BenchmarkTableOperations(b *testing.B) {
	// 初始化数据库
	_, err := storage.GetDBManager().OpenDB("./benchmark_db")
	if err != nil {
		b.Fatalf("Failed to open database: %v", err)
	}
	defer storage.CloseDb()

	// 创建表
	table, err := TableNew("benchmark_table")
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

	// 创建主键索引
	pk, err := DefaultPrimaryKeyNew("pk_id")
	if err != nil {
		b.Fatalf("Failed to create primary key: %v", err)
	}
	pk.AddFields("id")
	err = table.CreateIndex(pk)
	if err != nil {
		b.Fatalf("Failed to create index: %v", err)
	}

	// 创建普通索引
	normalIndex, err := DefaultNormalIndexNew("idx_name")
	if err != nil {
		b.Fatalf("Failed to create normal index: %v", err)
	}
	normalIndex.AddFields("name")
	err = table.CreateIndex(normalIndex)
	if err != nil {
		b.Fatalf("Failed to create normal index: %v", err)
	}

	// 基准测试插入操作
	b.Run("Insert", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			user := map[string]any{
				"name": "Test",
				"age":  30,
			}
			_, err := table.Insert(&user)
			if err != nil {
				b.Fatalf("Failed to insert: %v", err)
			}
		}
	})

	// 基准测试搜索操作
	b.Run("Search", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			searchData := map[string]any{"id": 1}
			iter, err := table.Search(&searchData)
			if err != nil {
				b.Fatalf("Search 失败: %v", err)
			}
			if iter == nil {
				b.Fatalf("Failed to get iterator: %v", err)
			}
			defer GlobalTableIterPool.Put(iter)
			if iter != nil {
				// 立即使用并释放资源，而不是使用defer
				records := iter.GetRecords(true)
				if records != nil {
					record.PutRecords(records)
				}
				GlobalTableIterPool.Put(iter)
			}
		}
	})

	// 基准测试更新操作
	b.Run("Update", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			updateData := map[string]any{
				"id":   1,
				"name": "Updated",
				"age":  31,
			}
			err := table.Update(&updateData)
			if err != nil {
				b.Fatalf("Failed to update: %v", err)
			}
		}
	})

	// 基准测试批量插入操作
	b.Run("BatchInsert", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			batch := storage.GetDBManager().GetDB().GetBatch()
			if batch == nil {
				b.Fatal("Failed to get batch")
			}

			for j := 0; j < 10; j++ {
				user := map[string]any{
					"name": "Batch",
					"age":  25,
				}
				_, err := table.Insert(&user, batch)
				if err != nil {
					b.Fatalf("Failed to batch insert: %v", err)
				}
			}

			err := storage.GetDBManager().GetDB().WriteBatch(batch)
			if err != nil {
				b.Fatalf("Failed to write batch: %v", err)
			}
		}
	})
}

// BenchmarkObjectPoolUsage 基准测试对象池使用性能
func BenchmarkObjectPoolUsage(b *testing.B) {
	// 初始化数据库
	_, err := storage.GetDBManager().OpenDB("./benchmark_pool_db")
	if err != nil {
		b.Fatalf("Failed to open database: %v", err)
	}
	defer storage.CloseDb()

	// 创建表
	table, err := TableNew("pool_test_table")
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

	// 创建主键索引
	pk, err := DefaultPrimaryKeyNew("pk_id")
	if err != nil {
		b.Fatalf("Failed to create primary key: %v", err)
	}
	pk.AddFields("id")
	err = table.CreateIndex(pk)
	if err != nil {
		b.Fatalf("Failed to create index: %v", err)
	}

	// 创建普通索引
	normalIndex, err := DefaultNormalIndexNew("idx_name")
	if err != nil {
		b.Fatalf("Failed to create normal index: %v", err)
	}
	normalIndex.AddFields("name")
	err = table.CreateIndex(normalIndex)
	if err != nil {
		b.Fatalf("Failed to create normal index: %v", err)
	}

	// 插入测试数据
	for i := 0; i < 10; i++ {
		user := map[string]any{
			"name": "Test",
			"age":  30,
		}
		_, err := table.Insert(&user)
		if err != nil {
			b.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// 基准测试使用对象池
	b.Run("WithObjectPool", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			// 使用对象池
			searchData := map[string]any{"name": "Test"}
			iter, err := table.Search(&searchData)
			if err != nil {
				b.Fatalf("Search 失败: %v", err)
			}
			if iter == nil {
				b.Fatalf("Failed to get iterator: %v", err)
			}
			defer GlobalTableIterPool.Put(iter)
			if iter == nil {
				continue
			}
			if iter != nil {
				// 立即使用并释放资源，而不是使用defer
				records := iter.GetRecords(true)
				if records != nil {
					record.PutRecords(records)
				}
				GlobalTableIterPool.Put(iter)
			}
		}
	})
}

// BenchmarkIteratorBatchUpdate 基准测试迭代器批量更新性能
func BenchmarkIteratorBatchUpdate(b *testing.B) {
	// 初始化数据库
	_, err := storage.GetDBManager().OpenDB("./benchmark_batch_update_db")
	if err != nil {
		b.Fatalf("Failed to open database: %v", err)
	}
	defer storage.CloseDb()

	// 创建表
	table, err := TableNew("batch_update_table")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// 设置字段
	fields := map[string]any{
		"id":     0,
		"name":   "",
		"age":    0,
		"status": "",
	}
	err = table.SetFields(fields)
	if err != nil {
		b.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键索引
	pk, err := DefaultPrimaryKeyNew("pk_id")
	if err != nil {
		b.Fatalf("Failed to create primary key: %v", err)
	}
	pk.AddFields("id")
	err = table.CreateIndex(pk)
	if err != nil {
		b.Fatalf("Failed to create index: %v", err)
	}

	// 创建普通索引
	normalIndex, err := DefaultNormalIndexNew("idx_name")
	if err != nil {
		b.Fatalf("Failed to create normal index: %v", err)
	}
	normalIndex.AddFields("name")
	err = table.CreateIndex(normalIndex)
	if err != nil {
		b.Fatalf("Failed to create normal index: %v", err)
	}

	// 插入测试数据
	for i := 0; i < 100; i++ {
		user := map[string]any{
			"name":   "Test",
			"age":    30,
			"status": "inactive",
		}
		_, err := table.Insert(&user)
		if err != nil {
			b.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// 基准测试迭代器批量更新
	b.Run("IteratorBatchUpdate", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			// 创建查询条件
			searchFields := map[string]any{"name": "Test"}

			// 获取迭代器
			iter, err := table.Search(&searchFields)
			if err != nil {
				b.Fatalf("Search 失败: %v", err)
			}
			if iter == nil {
				b.Fatalf("Failed to get iterator: %v", err)
			}
			defer GlobalTableIterPool.Put(iter)
			if iter == nil {
				continue
			}
			if iter != nil {
				// 立即使用并释放资源，而不是使用defer
				// 准备更新数据
				updateData := map[string]any{
					"status": "active",
				}

				// 执行批量更新
				err := iter.Update(&updateData, 50)
				if err != nil {
					b.Fatalf("Failed to batch update: %v", err)
				}

				// 释放迭代器
				GlobalTableIterPool.Put(iter)
			}
		}
	})
}
