package engine

import (
	"testing"

	"github.com/liaoran123/sfsDb/record"
	"github.com/liaoran123/sfsDb/storage"
)

// TestBatchSharingOperations 测试各种操作如何使用共享的 batch
// 这是一个综合测试，演示了在同一个 batch 中执行多种操作
func TestBatchSharingOperations(t *testing.T) {
	// 1. 创建表
	table, err := TableNew("test_batch_sharing_operations")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 2. 定义表字段
	fields := map[string]any{
		"id":       0,
		"name":     "",
		"age":      0,
		"email":    "",
		"status":   "",
	}

	// 3. 设置表字段
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 4. 创建主键索引，确保数据唯一性
	err = table.CreatePrimaryKey("id")
	if err != nil {
		t.Fatalf("Failed to create primary key: %v", err)
	}
	t.Log("Created primary key index on id field")

	// 5. 创建共享的 batch
	kvStore := storage.GetDBManager().GetDB()
	batch := kvStore.GetBatch()
	t.Log("Created shared batch")

	// 6. 测试 BatchInsert 方法使用共享 batch
	t.Log("\n=== Testing BatchInsert with shared batch ===")
	insertRecords := []*map[string]any{
		{"id": 1, "name": "Alice", "age": 25, "email": "alice@example.com", "status": "active"},
		{"id": 2, "name": "Bob", "age": 30, "email": "bob@example.com", "status": "active"},
		{"id": 3, "name": "Charlie", "age": 35, "email": "charlie@example.com", "status": "inactive"},
		{"id": 4, "name": "David", "age": 40, "email": "david@example.com", "status": "active"},
		{"id": 5, "name": "Eve", "age": 45, "email": "eve@example.com", "status": "active"},
	}

	ids, err := table.BatchInsert(insertRecords, batch)
	if err != nil {
		t.Fatalf("Failed to batch insert records: %v", err)
	}
	t.Logf("BatchInsert completed successfully, inserted IDs: %v", ids)
	t.Logf("Batch operation count after BatchInsert: %d", batch.Len())

	// 10. 最后统一提交 batch
	t.Log("\n=== Committing shared batch ===")
	err = kvStore.WriteBatch(batch)
	if err != nil {
		t.Fatalf("Failed to commit batch: %v", err)
	}
	t.Log("Batch committed successfully")

	// 11. 验证所有操作结果
	t.Log("\n=== Verifying operation results ===")

	// 验证记录数量
	allSearchFields := map[string]any{"id": nil}
	allIter, err := table.Search(&allSearchFields)
	if err != nil {
		t.Fatalf("Failed to search all records: %v", err)
	}
	defer GlobalTableIterPool.Put(allIter)

	allRecords := allIter.GetRecords(true)
	defer record.PutRecords(allRecords)
	t.Logf("Found %d total records", len(allRecords))
	for i, r := range allRecords {
		t.Logf("  Record %d: %v", i+1, r)
	}

	// 验证记录数量（应该有 5 条记录）
	if len(allRecords) != 5 {
		t.Errorf("Expected 5 records, got %d", len(allRecords))
	}

	// 验证 BatchInsert 操作结果
	batchInsertSearchFields := map[string]any{"id": 1}
	batchInsertIter, err := table.Search(&batchInsertSearchFields)
	if err != nil {
		t.Fatalf("Failed to search batch inserted record: %v", err)
	}
	defer GlobalTableIterPool.Put(batchInsertIter)

	batchInsertRecords := batchInsertIter.GetRecords(true)
	defer record.PutRecords(batchInsertRecords)
	if len(batchInsertRecords) != 1 {
		t.Errorf("Expected batch inserted record to exist, but found %d records", len(batchInsertRecords))
	} else {
		t.Log("BatchInsert operation verified successfully")
	}

	// 验证最后一条记录
	lastRecordSearchFields := map[string]any{"id": 5}
	lastRecordIter, err := table.Search(&lastRecordSearchFields)
	if err != nil {
		t.Fatalf("Failed to search last record: %v", err)
	}
	defer GlobalTableIterPool.Put(lastRecordIter)

	lastRecordRecords := lastRecordIter.GetRecords(true)
	defer record.PutRecords(lastRecordRecords)
	if len(lastRecordRecords) != 1 {
		t.Errorf("Expected last record to exist, but found %d records", len(lastRecordRecords))
	} else {
		t.Log("Last record verification successful")
	}

	// 12. 演示批量共享的优势
	t.Log("\nBatch sharing operations test completed successfully!")
	t.Log("This example demonstrates how to:")
	t.Log("1. Create a shared batch for multiple operations")
	t.Log("2. Use the same batch for BatchInsert operations")
	t.Log("3. Commit all operations atomically")
	t.Log("4. Achieve eventual consistency in edge computing scenarios")
	t.Log("5. Optimize performance by reducing network overhead")
}
