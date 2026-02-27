package main

import (
	"testing"

	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/storage"
	"github.com/liaoran123/sfsDb/transactionLockFree"
)

// 基准测试：直接使用 engine 包的批量操作
func BenchmarkEngineDirectBatchOperations(b *testing.B) {
	// 初始化数据库
	dbMgr := storage.GetDBManager()
	_, err := dbMgr.OpenDB("./benchmark_db")
	if err != nil {
		b.Fatalf("Failed to open database: %v", err)
	}
	defer dbMgr.CloseDB()

	// 创建测试表
	table, err := engine.TableNew("test_table")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]interface{}{
		"id":   0,
		"name": "",
		"age":  0,
	}
	if err := table.SetFields(fields); err != nil {
		b.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键索引
	primaryKey, err := engine.DefaultPrimaryKeyNew("id")
	if err != nil {
		b.Fatalf("Failed to create primary key: %v", err)
	}
	primaryKey.AddFields("id")
	if err := table.CreateIndex(primaryKey); err != nil {
		b.Fatalf("Failed to create index: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 直接使用 engine 包的批量操作
		batch := dbMgr.GetDB().GetBatch()
		if batch == nil {
			b.Fatalf("Failed to get batch")
		}

		// 插入多条记录
		for j := 0; j < 10; j++ {
			insertData := map[string]interface{}{
				"id":   i*10 + j,
				"name": "test_name",
				"age":  30,
			}
			_, err := table.Insert(&insertData, batch)
			if err != nil {
				b.Fatalf("Failed to insert data: %v", err)
			}
		}

		// 提交批量操作
		if err := dbMgr.GetDB().WriteBatch(batch); err != nil {
			b.Fatalf("Failed to write batch: %v", err)
		}
	}
}

// 基准测试：使用 transactionLockFree 包的批量操作
func BenchmarkTransactionLockFreeBatchOperations(b *testing.B) {
	// 初始化数据库
	dbMgr := storage.GetDBManager()
	_, err := dbMgr.OpenDB("./benchmark_db")
	if err != nil {
		b.Fatalf("Failed to open database: %v", err)
	}
	defer dbMgr.CloseDB()

	// 创建测试表
	table, err := engine.TableNew("test_table")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]interface{}{
		"id":   0,
		"name": "",
		"age":  0,
	}
	if err := table.SetFields(fields); err != nil {
		b.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键索引
	primaryKey, err := engine.DefaultPrimaryKeyNew("id")
	if err != nil {
		b.Fatalf("Failed to create primary key: %v", err)
	}
	primaryKey.AddFields("id")
	if err := table.CreateIndex(primaryKey); err != nil {
		b.Fatalf("Failed to create index: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 使用 transactionLockFree 包的事务
		tx, err := transactionLockFree.NewTableTransaction(table)
		if err != nil {
			b.Fatalf("Failed to create table transaction: %v", err)
		}

		// 插入多条记录
		for j := 0; j < 10; j++ {
			insertData := map[string]interface{}{
				"id":   i*10 + j,
				"name": "test_name",
				"age":  30,
			}
			_, err := tx.Insert(&insertData)
			if err != nil {
				b.Fatalf("Failed to insert data: %v", err)
			}
		}

		// 提交事务
		if err := tx.Commit(); err != nil {
			b.Fatalf("Failed to commit transaction: %v", err)
		}
	}
}

// 基准测试：直接使用 engine 包的单条操作
func BenchmarkEngineDirectSingleOperations(b *testing.B) {
	// 初始化数据库
	dbMgr := storage.GetDBManager()
	_, err := dbMgr.OpenDB("./benchmark_db")
	if err != nil {
		b.Fatalf("Failed to open database: %v", err)
	}
	defer dbMgr.CloseDB()

	// 创建测试表
	table, err := engine.TableNew("test_table")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]interface{}{
		"id":   0,
		"name": "",
		"age":  0,
	}
	if err := table.SetFields(fields); err != nil {
		b.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键索引
	primaryKey, err := engine.DefaultPrimaryKeyNew("id")
	if err != nil {
		b.Fatalf("Failed to create primary key: %v", err)
	}
	primaryKey.AddFields("id")
	if err := table.CreateIndex(primaryKey); err != nil {
		b.Fatalf("Failed to create index: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 直接使用 engine 包的单条插入
		insertData := map[string]interface{}{
			"id":   i,
			"name": "test_name",
			"age":  30,
		}
		_, err := table.Insert(&insertData)
		if err != nil {
			b.Fatalf("Failed to insert data: %v", err)
		}
	}
}

// 基准测试：使用 transactionLockFree 包的单条操作
func BenchmarkTransactionLockFreeSingleOperations(b *testing.B) {
	// 初始化数据库
	dbMgr := storage.GetDBManager()
	_, err := dbMgr.OpenDB("./benchmark_db")
	if err != nil {
		b.Fatalf("Failed to open database: %v", err)
	}
	defer dbMgr.CloseDB()

	// 创建测试表
	table, err := engine.TableNew("test_table")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]interface{}{
		"id":   0,
		"name": "",
		"age":  0,
	}
	if err := table.SetFields(fields); err != nil {
		b.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键索引
	primaryKey, err := engine.DefaultPrimaryKeyNew("id")
	if err != nil {
		b.Fatalf("Failed to create primary key: %v", err)
	}
	primaryKey.AddFields("id")
	if err := table.CreateIndex(primaryKey); err != nil {
		b.Fatalf("Failed to create index: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 使用 transactionLockFree 包的事务
		tx, err := transactionLockFree.NewTableTransaction(table)
		if err != nil {
			b.Fatalf("Failed to create table transaction: %v", err)
		}

		// 插入单条记录
		insertData := map[string]interface{}{
			"id":   i,
			"name": "test_name",
			"age":  30,
		}
		_, err = tx.Insert(&insertData)
		if err != nil {
			b.Fatalf("Failed to insert data: %v", err)
		}

		// 提交事务
		if err := tx.Commit(); err != nil {
			b.Fatalf("Failed to commit transaction: %v", err)
		}
	}
}
