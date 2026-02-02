package engine

import (
	"fmt"
	"sync"
	"testing"

	"github.com/liaoran123/sfsDb/storage"
)

// TestBatchInsert 测试批量插入功能
func TestBatchInsert(t *testing.T) {
	// 初始化数据库
	_, err := storage.OpenDefaultDb("./batch_test_db")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// 创建测试表
	table, err := TableNew("test_batch")
	if err != nil {
		t.Fatalf("TableNew failed: %v", err)
	}

	// 设置字段
	fields := map[string]any{"id": 0, "name": "", "age": 0, "description": ""}
	table.SetFields(fields)

	// 创建主键
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("id")
	table.CreateIndex(pk)

	// 创建测试记录
	records := []*map[string]any{
		&map[string]any{"name": "Alice", "age": 25, "description": "Software Engineer"},
		&map[string]any{"name": "Bob", "age": 30, "description": "Product Manager"},
		&map[string]any{"name": "Charlie", "age": 35, "description": "Designer"},
		&map[string]any{"name": "David", "age": 40, "description": "Developer"},
		&map[string]any{"name": "Eve", "age": 45, "description": "Manager"},
	}

	// 批量插入
	ids, err := table.BatchInsert(records)
	if err != nil {
		t.Fatalf("BatchInsert failed: %v", err)
	}

	// 验证插入结果
	if len(ids) != len(records) {
		t.Errorf("Expected %d IDs, got %d", len(records), len(ids))
	}

	// 验证ID是否连续
	for i := 1; i < len(ids); i++ {
		if ids[i] != ids[i-1]+1 {
			t.Errorf("IDs should be consecutive, got %d and %d", ids[i-1], ids[i])
		}
	}

	// 验证数据是否正确插入
	for i, id := range ids {
		searchFields := map[string]any{"id": id}
		iter := table.Search(&searchFields)
		defer GlobalTableIterPool.Put(iter)
		if !iter.Next() {
			t.Errorf("Record with ID %d not found", id)
			continue
		}
		foundRecords := iter.GetRecords(true)
		if len(foundRecords) == 0 {
			t.Errorf("No records found for ID %d", id)
			continue
		}
		foundRecord := foundRecords[0]
		expectedRecord := *records[i]
		if foundRecord["name"] != expectedRecord["name"] {
			t.Errorf("Expected name %s, got %s", expectedRecord["name"], foundRecord["name"])
		}
		if foundRecord["age"] != expectedRecord["age"] {
			t.Errorf("Expected age %d, got %d", expectedRecord["age"], foundRecord["age"])
		}
		if foundRecord["description"] != expectedRecord["description"] {
			t.Errorf("Expected description %s, got %s", expectedRecord["description"], foundRecord["description"])
		}
	}

	t.Logf("BatchInsert test passed, inserted %d records with IDs: %v", len(ids), ids)
}

// TestBatchInsertWithSize 测试带批量大小控制的批量插入
func TestBatchInsertWithSize(t *testing.T) {
	// 初始化数据库
	_, err := storage.OpenDefaultDb("./batch_test_db")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// 创建测试表
	table, err := TableNew("test_batch_size")
	if err != nil {
		t.Fatalf("TableNew failed: %v", err)
	}

	// 设置字段
	fields := map[string]any{"id": 0, "name": "", "age": 0}
	table.SetFields(fields)

	// 创建主键
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("id")
	table.CreateIndex(pk)

	// 创建大量测试记录
	recordCount := 250
	records := make([]*map[string]any, recordCount)
	for i := 0; i < recordCount; i++ {
		records[i] = &map[string]any{
			"name": fmt.Sprintf("Record%d", i),
			"age":  20 + i%50,
		}
	}

	// 带批量大小控制的批量插入
	batchSize := 50
	ids, err := table.BatchInsertWithSize(records, batchSize)
	if err != nil {
		t.Fatalf("BatchInsertWithSize failed: %v", err)
	}

	// 验证插入结果
	if len(ids) != recordCount {
		t.Errorf("Expected %d IDs, got %d", recordCount, len(ids))
	}

	// 验证ID是否连续
	for i := 1; i < len(ids); i++ {
		if ids[i] != ids[i-1]+1 {
			t.Errorf("IDs should be consecutive, got %d and %d", ids[i-1], ids[i])
		}
	}

	t.Logf("BatchInsertWithSize test passed, inserted %d records with batch size %d", recordCount, batchSize)
}

// TestBatchInsertConcurrent 测试并发批量插入
func TestBatchInsertConcurrent(t *testing.T) {
	// 初始化数据库
	_, err := storage.OpenDefaultDb("./batch_test_db")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// 创建测试表
	table, err := TableNew("test_batch_concurrent")
	if err != nil {
		t.Fatalf("TableNew failed: %v", err)
	}

	// 设置字段
	fields := map[string]any{"id": 0, "name": "", "age": 0}
	table.SetFields(fields)

	// 创建主键
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("id")
	table.CreateIndex(pk)

	// 并发批量插入
	var wg sync.WaitGroup
	batchCount := 5
	batchSize := 10
	var mu sync.Mutex
	allIDs := make(map[int]bool)

	for i := 0; i < batchCount; i++ {
		wg.Add(1)
		go func(batchID int) {
			defer wg.Done()
			records := make([]*map[string]any, batchSize)
			for j := 0; j < batchSize; j++ {
				records[j] = &map[string]any{
					"name": fmt.Sprintf("Batch%d_Record%d", batchID, j),
					"age":  20 + j,
				}
			}
			ids, err := table.BatchInsert(records)
			if err != nil {
				t.Errorf("BatchInsert failed: %v", err)
				return
			}
			// 收集ID，检查重复
			mu.Lock()
			for _, id := range ids {
				if allIDs[id] {
					t.Errorf("Duplicate ID found: %d", id)
				}
				allIDs[id] = true
			}
			mu.Unlock()
			t.Logf("Batch %d inserted IDs: %v", batchID, ids)
		}(i)
	}

	wg.Wait()

	// 验证总记录数
	totalRecords := batchCount * batchSize
	if len(allIDs) != totalRecords {
		t.Errorf("Expected %d records, got %d", totalRecords, len(allIDs))
	}

	t.Logf("Concurrent batch insert test passed, inserted %d records", totalRecords)
}
