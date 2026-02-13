package engine

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/liaoran123/sfsDb/storage"
)

// TestTableTransactionPool 测试事务对象池的基本功能
func TestTableTransactionPool(t *testing.T) {
	// 创建存储
	store, err := storage.OpenDefaultDb("./test_transaction_pool")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer func() {
		store.Close()
		_ = os.RemoveAll("./test_transaction_pool")
	}()

	// 创建表
	table, err := TableNew("test_table")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置字段
	fields := map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	}
	if err := table.SetFields(fields); err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键索引
	pk, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		t.Fatalf("Failed to create primary key: %v", err)
	}
	pk.AddFields("id")
	if err := table.CreateIndex(pk); err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	// 测试事务对象池
	for i := 0; i < 100; i++ {
		// 开始事务
		tx, err := table.Begin()
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}

		// 插入数据
		record := map[string]any{
			"name": fmt.Sprintf("test_%d", i),
			"age":  i % 100,
		}
		id, err := tx.Insert(&record)
		if err != nil {
			t.Fatalf("Failed to insert record: %v", err)
		}

		// 提交事务
		if err := tx.Commit(); err != nil {
			t.Fatalf("Failed to commit transaction: %v", err)
		}

		// 验证数据
		checkRecord := map[string]any{"id": id}
		result, err := table.Read(&checkRecord)
		if err != nil {
			t.Fatalf("Failed to read record: %v", err)
		}
		if len(result) == 0 {
			t.Fatalf("Record not found after insert: %d", id)
		}
	}

	t.Log("TableTransactionPool basic test passed")
}

// BenchmarkTableTransactionCreation 基准测试事务创建性能
func BenchmarkTableTransactionCreation(b *testing.B) {
	// 创建存储
	store, err := storage.OpenDefaultDb("./bench_transaction_creation")
	if err != nil {
		b.Fatalf("Failed to create store: %v", err)
	}
	defer func() {
		store.Close()
		_ = os.RemoveAll("./bench_transaction_creation")
	}()

	// 创建表
	table, err := TableNew("bench_table")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// 设置字段
	fields := map[string]any{
		"id":   0,
		"name": "",
	}
	if err := table.SetFields(fields); err != nil {
		b.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键索引
	pk, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		b.Fatalf("Failed to create primary key: %v", err)
	}
	pk.AddFields("id")
	if err := table.CreateIndex(pk); err != nil {
		b.Fatalf("Failed to create index: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 开始事务
		tx, err := table.Begin()
		if err != nil {
			b.Fatalf("Failed to begin transaction: %v", err)
		}

		// 插入数据
		record := map[string]any{
			"name": fmt.Sprintf("bench_%d", i),
		}
		_, err = tx.Insert(&record)
		if err != nil {
			b.Fatalf("Failed to insert record: %v", err)
		}

		// 提交事务
		if err := tx.Commit(); err != nil {
			b.Fatalf("Failed to commit transaction: %v", err)
		}
	}
}

// BenchmarkTableTransactionConcurrent 基准测试并发事务性能
func BenchmarkTableTransactionConcurrent(b *testing.B) {
	// 创建存储
	store, err := storage.OpenDefaultDb("./bench_transaction_concurrent")
	if err != nil {
		b.Fatalf("Failed to create store: %v", err)
	}
	defer func() {
		store.Close()
		_ = os.RemoveAll("./bench_transaction_concurrent")
	}()

	// 创建表
	table, err := TableNew("bench_concurrent")
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// 设置字段
	fields := map[string]any{
		"id":    0,
		"name":  "",
		"value": 0,
	}
	if err := table.SetFields(fields); err != nil {
		b.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键索引
	pk, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		b.Fatalf("Failed to create primary key: %v", err)
	}
	pk.AddFields("id")
	if err := table.CreateIndex(pk); err != nil {
		b.Fatalf("Failed to create index: %v", err)
	}

	// 并发测试
	var wg sync.WaitGroup
	concurrency := 10
	tasksPerGoroutine := b.N / concurrency
	if tasksPerGoroutine < 1 {
		tasksPerGoroutine = 1
	}

	b.ResetTimer()

	for g := 0; g < concurrency; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for i := 0; i < tasksPerGoroutine; i++ {
				// 开始事务
				tx, err := table.Begin()
				if err != nil {
					b.Errorf("Failed to begin transaction: %v", err)
					return
				}

				// 插入数据
				record := map[string]any{
					"name":  fmt.Sprintf("concurrent_%d_%d", goroutineID, i),
					"value": goroutineID*1000 + i,
				}
				_, err = tx.Insert(&record)
				if err != nil {
					b.Errorf("Failed to insert record: %v", err)
					return
				}

				// 提交事务
				if err := tx.Commit(); err != nil {
					b.Errorf("Failed to commit transaction: %v", err)
					return
				}
			}
		}(g)
	}

	wg.Wait()
}

// TestTableTransactionPoolReset 测试事务对象池的重置功能
func TestTableTransactionPoolReset(t *testing.T) {
	// 创建存储
	store, err := storage.OpenDefaultDb("./test_transaction_reset")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer func() {
		store.Close()
		_ = os.RemoveAll("./test_transaction_reset")
	}()

	// 创建表
	table, err := TableNew("test_reset")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置字段
	fields := map[string]any{
		"id":   0,
		"name": "",
	}
	if err := table.SetFields(fields); err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键索引
	pk, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		t.Fatalf("Failed to create primary key: %v", err)
	}
	pk.AddFields("id")
	if err := table.CreateIndex(pk); err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	// 测试对象重用
	var tx1, tx2 Transaction
	var err1, err2 error

	// 第一次使用
	tx1, err1 = table.Begin()
	if err1 != nil {
		t.Fatalf("Failed to begin transaction 1: %v", err1)
	}

	record1 := map[string]any{"name": "test1"}
	_, err1 = tx1.Insert(&record1)
	if err1 != nil {
		t.Fatalf("Failed to insert record 1: %v", err1)
	}

	if err1 := tx1.Commit(); err1 != nil {
		t.Fatalf("Failed to commit transaction 1: %v", err1)
	}

	// 第二次使用
	tx2, err2 = table.Begin()
	if err2 != nil {
		t.Fatalf("Failed to begin transaction 2: %v", err2)
	}

	record2 := map[string]any{"name": "test2"}
	_, err2 = tx2.Insert(&record2)
	if err2 != nil {
		t.Fatalf("Failed to insert record 2: %v", err2)
	}

	if err2 := tx2.Commit(); err2 != nil {
		t.Fatalf("Failed to commit transaction 2: %v", err2)
	}

	// 验证两次事务都成功执行
	checkRecord1 := map[string]any{"id": 1}
	result1, err := table.Read(&checkRecord1)
	if err != nil || len(result1) == 0 {
		t.Fatalf("Record 1 not found: %v", err)
	}

	checkRecord2 := map[string]any{"id": 2}
	result2, err := table.Read(&checkRecord2)
	if err != nil || len(result2) == 0 {
		t.Fatalf("Record 2 not found: %v", err)
	}

	t.Log("TableTransactionPool reset test passed")
}

// TestTableTransactionNested 测试嵌套事务
func TestTableTransactionNested(t *testing.T) {
	// 创建存储
	store, err := storage.OpenDefaultDb("./test_transaction_nested")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer func() {
		store.Close()
		_ = os.RemoveAll("./test_transaction_nested")
	}()

	// 创建表
	table, err := TableNew("test_nested")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置字段
	fields := map[string]any{
		"id":    0,
		"name":  "",
		"value": 0,
	}
	if err := table.SetFields(fields); err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键索引
	pk, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		t.Fatalf("Failed to create primary key: %v", err)
	}
	pk.AddFields("id")
	if err := table.CreateIndex(pk); err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	// 创建支持嵌套事务的选项
	options := DefaultTransactionOptions()
	options.AllowNested = true

	// 开始父事务
	parentTx, err := table.BeginWithOptions(options)
	if err != nil {
		t.Fatalf("Failed to begin parent transaction: %v", err)
	}

	// 插入父事务数据
	parentRecord := map[string]any{"name": "parent", "value": 100}
	parentID, err := parentTx.Insert(&parentRecord)
	if err != nil {
		t.Fatalf("Failed to insert parent record: %v", err)
	}

	// 开始嵌套事务
	nestedTx, err := parentTx.BeginNested()
	if err != nil {
		t.Fatalf("Failed to begin nested transaction: %v", err)
	}

	// 插入嵌套事务数据
	nestedRecord := map[string]any{"name": "nested", "value": 200}
	nestedID, err := nestedTx.Insert(&nestedRecord)
	if err != nil {
		t.Fatalf("Failed to insert nested record: %v", err)
	}

	// 提交嵌套事务
	if err := nestedTx.Commit(); err != nil {
		t.Fatalf("Failed to commit nested transaction: %v", err)
	}

	// 提交父事务
	if err := parentTx.Commit(); err != nil {
		t.Fatalf("Failed to commit parent transaction: %v", err)
	}

	// 验证数据
	checkParent := map[string]any{"id": parentID}
	resultParent, err := table.Read(&checkParent)
	if err != nil || len(resultParent) == 0 {
		t.Fatalf("Parent record not found: %v", err)
	}

	checkNested := map[string]any{"id": nestedID}
	resultNested, err := table.Read(&checkNested)
	if err != nil || len(resultNested) == 0 {
		t.Fatalf("Nested record not found: %v", err)
	}

	t.Log("TableTransaction nested test passed")
}

// TestTableTransactionPoolPerformance 测试事务对象池的性能改进
func TestTableTransactionPoolPerformance(t *testing.T) {
	// 创建存储
	store, err := storage.OpenDefaultDb("./test_transaction_performance")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer func() {
		store.Close()
		_ = os.RemoveAll("./test_transaction_performance")
	}()

	// 创建表
	table, err := TableNew("test_performance")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置字段
	fields := map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	}
	if err := table.SetFields(fields); err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键索引
	pk, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		t.Fatalf("Failed to create primary key: %v", err)
	}
	pk.AddFields("id")
	if err := table.CreateIndex(pk); err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	// 测试参数
	iterations := 1000
	concurrency := 5

	// 测量并发性能
	startTime := time.Now()
	var wg sync.WaitGroup

	for g := 0; g < concurrency; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for i := 0; i < iterations/concurrency; i++ {
				// 开始事务
				tx, err := table.Begin()
				if err != nil {
					t.Errorf("Failed to begin transaction: %v", err)
					return
				}

				// 插入数据
				record := map[string]any{
					"name": fmt.Sprintf("performance_%d_%d", goroutineID, i),
					"age":  (goroutineID*1000 + i) % 100,
				}
				_, err = tx.Insert(&record)
				if err != nil {
					t.Errorf("Failed to insert record: %v", err)
					return
				}

				// 提交事务
				if err := tx.Commit(); err != nil {
					t.Errorf("Failed to commit transaction: %v", err)
					return
				}
			}
		}(g)
	}

	wg.Wait()
	duration := time.Since(startTime)

	t.Logf("Performance test completed: %d transactions in %v (%f transactions/sec)",
		iterations, duration, float64(iterations)/duration.Seconds())
}
