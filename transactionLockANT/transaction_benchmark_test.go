package transactionLockANT

import (
	"testing"
	"time"

	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/storage"
)

// 基准测试：事务创建和提交
func BenchmarkTransactionCreationAndCommit(b *testing.B) {
	// 初始化数据库
	dbMgr := storage.GetDBManager()
	db, err := dbMgr.OpenDB("./benchmark_db")
	if err != nil {
		b.Fatalf("Failed to open database: %v", err)
	}
	defer dbMgr.CloseDB()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 创建事务
		tx, err := NewTransaction(db, "")
		if err != nil {
			b.Fatalf("Failed to create transaction: %v", err)
		}

		// 提交事务
		if err := tx.Commit(); err != nil {
			b.Fatalf("Failed to commit transaction: %v", err)
		}
	}
}

// 基准测试：事务内的读写操作
func BenchmarkTransactionReadWrite(b *testing.B) {
	// 初始化数据库
	dbMgr := storage.GetDBManager()
	db, err := dbMgr.OpenDB("./benchmark_db")
	if err != nil {
		b.Fatalf("Failed to open database: %v", err)
	}
	defer dbMgr.CloseDB()

	// 准备测试数据
	key := []byte("test_key")
	value := []byte("test_value")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 创建事务
		tx, err := NewTransaction(db, "")
		if err != nil {
			b.Fatalf("Failed to create transaction: %v", err)
		}

		// 写入数据
		tx.Put(key, value)

		// 读取数据
		_, err = tx.Get(key)
		if err != nil {
			b.Fatalf("Failed to get value: %v", err)
		}

		// 提交事务
		if err := tx.Commit(); err != nil {
			b.Fatalf("Failed to commit transaction: %v", err)
		}
	}
}

// 基准测试：批量操作性能
func BenchmarkTransactionBatchOperations(b *testing.B) {
	// 初始化数据库
	dbMgr := storage.GetDBManager()
	db, err := dbMgr.OpenDB("./benchmark_db")
	if err != nil {
		b.Fatalf("Failed to open database: %v", err)
	}
	defer dbMgr.CloseDB()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 创建事务
		tx, err := NewTransaction(db, "")
		if err != nil {
			b.Fatalf("Failed to create transaction: %v", err)
		}

		// 批量写入
		for j := 0; j < 10; j++ {
			key := []byte("batch_key_" + string(rune(j)))
			value := []byte("batch_value_" + string(rune(j)))
			tx.Put(key, value)
		}

		// 提交事务
		if err := tx.Commit(); err != nil {
			b.Fatalf("Failed to commit transaction: %v", err)
		}
	}
}

// 基准测试：嵌套事务性能
func BenchmarkNestedTransaction(b *testing.B) {
	// 初始化数据库
	dbMgr := storage.GetDBManager()
	db, err := dbMgr.OpenDB("./benchmark_db")
	if err != nil {
		b.Fatalf("Failed to open database: %v", err)
	}
	defer dbMgr.CloseDB()

	// 创建带嵌套事务选项的事务
	options := DefaultTransactionOptions()
	options.AllowNested = true

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 创建父事务
		parentTx, err := NewTransactionWithOptions(db, options, "")
		if err != nil {
			b.Fatalf("Failed to create parent transaction: %v", err)
		}

		// 创建嵌套事务
		nestedTx, err := parentTx.BeginNested()
		if err != nil {
			b.Fatalf("Failed to create nested transaction: %v", err)
		}

		// 在嵌套事务中写入数据
		nestedTx.Put([]byte("nested_key"), []byte("nested_value"))

		// 提交嵌套事务
		if err := nestedTx.Commit(); err != nil {
			b.Fatalf("Failed to commit nested transaction: %v", err)
		}

		// 提交父事务
		if err := parentTx.Commit(); err != nil {
			b.Fatalf("Failed to commit parent transaction: %v", err)
		}
	}
}

// 基准测试：表事务性能
func BenchmarkTableTransaction(b *testing.B) {
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
		// 创建表事务
		tx, err := NewTableTransaction(table, "")
		if err != nil {
			b.Fatalf("Failed to create table transaction: %v", err)
		}

		// 插入数据
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

// 基准测试：并发事务性能
func BenchmarkConcurrentTransactions(b *testing.B) {
	// 初始化数据库
	dbMgr := storage.GetDBManager()
	db, err := dbMgr.OpenDB("./benchmark_db")
	if err != nil {
		b.Fatalf("Failed to open database: %v", err)
	}
	defer dbMgr.CloseDB()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// 创建事务
			tx, err := NewTransaction(db, "")
			if err != nil {
				b.Fatalf("Failed to create transaction: %v", err)
			}

			// 写入数据
			key := []byte("concurrent_key_" + time.Now().String())
			value := []byte("concurrent_value")
			tx.Put(key, value)

			// 提交事务
			if err := tx.Commit(); err != nil {
				b.Fatalf("Failed to commit transaction: %v", err)
			}
		}
	})
}

// 基准测试：事务回滚性能
func BenchmarkTransactionRollback(b *testing.B) {
	// 初始化数据库
	dbMgr := storage.GetDBManager()
	db, err := dbMgr.OpenDB("./benchmark_db")
	if err != nil {
		b.Fatalf("Failed to open database: %v", err)
	}
	defer dbMgr.CloseDB()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 创建事务
		tx, err := NewTransaction(db, "")
		if err != nil {
			b.Fatalf("Failed to create transaction: %v", err)
		}

		// 写入数据
		tx.Put([]byte("rollback_key"), []byte("rollback_value"))

		// 回滚事务
		if err := tx.Rollback(); err != nil {
			b.Fatalf("Failed to rollback transaction: %v", err)
		}
	}
}
