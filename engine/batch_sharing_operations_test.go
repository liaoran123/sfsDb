package engine

import (
	"testing"

	"github.com/liaoran123/sfsDb/record"
	"github.com/liaoran123/sfsDb/storage"
)

// TestBatchSharingOperations 测试各种操作如何使用共享的 batch
// 这是一个综合测试，演示了在同一个 batch 中执行多种操作
func TestBatchSharingOperations(t *testing.T) {
	// 1. 创建多个表
	table1, err := TableNew("test_batch_table1")
	if err != nil {
		t.Fatalf("Failed to create table1: %v", err)
	}

	table2, err := TableNew("test_batch_table2")
	if err != nil {
		t.Fatalf("Failed to create table2: %v", err)
	}

	// 2. 定义表1字段
	fields1 := map[string]any{
		"id":     0,
		"name":   "",
		"age":    0,
		"email":  "",
		"status": "",
	}

	// 3. 定义表2字段
	fields2 := map[string]any{
		"id":       0,
		"product":  "",
		"price":    0.0,
		"quantity": 0,
		"category": "",
	}

	// 4. 设置表1字段
	err = table1.SetFields(fields1)
	if err != nil {
		t.Fatalf("Failed to set fields for table1: %v", err)
	}

	// 5. 设置表2字段
	err = table2.SetFields(fields2)
	if err != nil {
		t.Fatalf("Failed to set fields for table2: %v", err)
	}

	// 6. 创建主键索引，确保数据唯一性
	err = table1.CreatePrimaryKey("id")
	if err != nil {
		t.Fatalf("Failed to create primary key for table1: %v", err)
	}
	t.Log("Created primary key index on id field for table1")

	err = table2.CreatePrimaryKey("id")
	if err != nil {
		t.Fatalf("Failed to create primary key for table2: %v", err)
	}
	t.Log("Created primary key index on id field for table2")

	// 7. 创建共享的 batch
	kvStore := storage.GetDBManager().GetDB()
	batch := kvStore.GetBatch()
	t.Log("Created shared batch")

	// 8. 在表1上执行批量插入操作
	t.Log("\n=== Testing BatchInsert on table1 with shared batch ===")
	insertRecords1 := []*map[string]any{
		{"id": 1, "name": "Alice", "age": 25, "email": "alice@example.com", "status": "active"},
		{"id": 2, "name": "Bob", "age": 30, "email": "bob@example.com", "status": "active"},
		{"id": 3, "name": "Charlie", "age": 35, "email": "charlie@example.com", "status": "inactive"},
	}

	ids1, err := table1.BatchInsertInc(insertRecords1, batch)
	if err != nil {
		t.Fatalf("Failed to batch insert records into table1: %v", err)
	}
	t.Logf("BatchInsertInc on table1 completed successfully, inserted IDs: %v", ids1)
	t.Logf("Batch operation count after table1 insert: %d", batch.Len())

	// 9. 在表2上执行批量插入操作
	t.Log("\n=== Testing BatchInsert on table2 with shared batch ===")
	insertRecords2 := []*map[string]any{
		{"id": 1, "product": "Laptop", "price": 999.99, "quantity": 10, "category": "Electronics"},
		{"id": 2, "product": "Smartphone", "price": 699.99, "quantity": 20, "category": "Electronics"},
		{"id": 3, "product": "Desk Chair", "price": 199.99, "quantity": 15, "category": "Furniture"},
	}

	ids2, err := table2.BatchInsertInc(insertRecords2, batch)
	if err != nil {
		t.Fatalf("Failed to batch insert records into table2: %v", err)
	}
	t.Logf("BatchInsertInc on table2 completed successfully, inserted IDs: %v", ids2)
	t.Logf("Batch operation count after table2 insert: %d", batch.Len())

	// 10. 最后统一提交 batch
	t.Log("\n=== Committing shared batch ===")
	err = kvStore.WriteBatch(batch)
	if err != nil {
		t.Fatalf("Failed to commit batch: %v", err)
	}
	t.Log("Batch committed successfully")

	// 11. 验证所有操作结果
	t.Log("\n=== Verifying operation results ===")

	// 验证表1记录
	t.Log("\n--- Verifying table1 records ---")
	allSearchFields1 := map[string]any{"id": nil}
	allIter1, err := table1.Search(&allSearchFields1)
	if err != nil {
		t.Fatalf("Failed to search all records in table1: %v", err)
	}
	defer GlobalTableIterPool.Put(allIter1)

	allRecords1 := allIter1.GetRecords(true)
	defer record.PutRecords(allRecords1)
	t.Logf("Found %d total records in table1", len(allRecords1))
	for i, r := range allRecords1 {
		t.Logf("  Table1 Record %d: %v", i+1, r)
	}

	// 验证表1记录数量
	if len(allRecords1) != 3 {
		t.Errorf("Expected 3 records in table1, got %d", len(allRecords1))
	}

	// 验证表2记录
	t.Log("\n--- Verifying table2 records ---")
	allSearchFields2 := map[string]any{"id": nil}
	allIter2, err := table2.Search(&allSearchFields2)
	if err != nil {
		t.Fatalf("Failed to search all records in table2: %v", err)
	}
	defer GlobalTableIterPool.Put(allIter2)

	allRecords2 := allIter2.GetRecords(true)
	defer record.PutRecords(allRecords2)
	t.Logf("Found %d total records in table2", len(allRecords2))
	for i, r := range allRecords2 {
		t.Logf("  Table2 Record %d: %v", i+1, r)
	}

	// 验证表2记录数量
	if len(allRecords2) != 3 {
		t.Errorf("Expected 3 records in table2, got %d", len(allRecords2))
	}

	// 12. 验证表1特定记录
	t.Log("\n--- Verifying specific records ---")
	batchInsertSearchFields1 := map[string]any{"id": 1}
	batchInsertIter1, err := table1.Search(&batchInsertSearchFields1)
	if err != nil {
		t.Fatalf("Failed to search batch inserted record in table1: %v", err)
	}
	defer GlobalTableIterPool.Put(batchInsertIter1)

	batchInsertRecords1 := batchInsertIter1.GetRecords(true)
	defer record.PutRecords(batchInsertRecords1)
	if len(batchInsertRecords1) != 1 {
		t.Errorf("Expected batch inserted record to exist in table1, but found %d records", len(batchInsertRecords1))
	} else {
		t.Log("BatchInsert operation on table1 verified successfully")
	}

	// 验证表2特定记录
	batchInsertSearchFields2 := map[string]any{"id": 2}
	batchInsertIter2, err := table2.Search(&batchInsertSearchFields2)
	if err != nil {
		t.Fatalf("Failed to search batch inserted record in table2: %v", err)
	}
	defer GlobalTableIterPool.Put(batchInsertIter2)

	batchInsertRecords2 := batchInsertIter2.GetRecords(true)
	defer record.PutRecords(batchInsertRecords2)
	if len(batchInsertRecords2) != 1 {
		t.Errorf("Expected batch inserted record to exist in table2, but found %d records", len(batchInsertRecords2))
	} else {
		t.Log("BatchInsert operation on table2 verified successfully")
	}

	// 13. 演示批量共享的优势
	t.Log("\nBatch sharing operations test completed successfully!")
	t.Log("This example demonstrates how to:")
	t.Log("1. Create multiple tables")
	t.Log("2. Create a shared batch for multiple tables")
	t.Log("3. Use the same batch for operations on different tables")
	t.Log("4. Commit all operations atomically across multiple tables")
	t.Log("5. Verify results from all tables")
	t.Log("6. Achieve atomicity when modifying multiple tables")
	t.Log("7. Optimize performance by reducing network overhead")
}
