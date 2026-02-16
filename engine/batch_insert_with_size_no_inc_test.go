package engine

import (
	"fmt"
	"testing"

	"github.com/liaoran123/sfsDb/storage"
)

// TestBatchInsertWithSizeNoInc 测试BatchInsertWithSizeNoInc函数
func TestBatchInsertWithSizeNoInc(t *testing.T) {
	// 初始化数据库
	_, err := storage.OpenDefaultDb("./batch_with_size_no_inc_test_db_1")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// 创建测试表
	table, err := TableNew("test_batch_with_size_no_inc")
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

	// 创建测试记录（包含手动指定的ID）
	recordCount := 20
	records := make([]*map[string]any, recordCount)
	for i := 0; i < recordCount; i++ {
		records[i] = &map[string]any{
			"id":   200 + i,
			"name": fmt.Sprintf("Record%d", i),
			"age":  20 + i%30,
		}
	}

	// 执行BatchInsertWithSizeNoInc（批次大小为5）
	batchSize := 5
	ids, err := table.BatchInsertWithSizeNoInc(records, batchSize, false)
	if err != nil {
		t.Fatalf("BatchInsertWithSizeNoInc failed: %v", err)
	}

	// 验证插入结果
	t.Logf("Inserted IDs: %v", ids)
	if len(ids) != recordCount {
		t.Errorf("Expected %d IDs, got %d", recordCount, len(ids))
	}

	// 验证指定的ID是否正确返回
	for i, id := range ids {
		expectedID := 200 + i
		if id != expectedID {
			t.Errorf("Expected ID %d at index %d, got %d", expectedID, i, id)
		}
	}

	// 验证数据是否正确插入
	for _, id := range ids {
		searchFields := map[string]any{"id": id}
		iter, err := table.Search(&searchFields)
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}
		if iter == nil {
			t.Fatalf("Failed to get iterator: %v", err)
		}
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
	}

	t.Logf("BatchInsertWithSizeNoInc test passed")
}

// TestBatchInsertWithSizeNoIncSkipVersion 测试BatchInsertWithSizeNoInc函数的skipVersion参数
func TestBatchInsertWithSizeNoIncSkipVersion(t *testing.T) {
	// 初始化数据库
	_, err := storage.OpenDefaultDb("./batch_with_size_no_inc_skip_version_test_db_1")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// 创建测试表
	table, err := TableNew("test_batch_with_size_no_inc_skip_version")
	if err != nil {
		t.Fatalf("TableNew failed: %v", err)
	}

	// 设置字段
	fields := map[string]any{"id": 0, "name": ""}
	table.SetFields(fields)

	// 创建主键
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("id")
	table.CreateIndex(pk)

	// 创建测试记录
	recordCount := 10
	records := make([]*map[string]any, recordCount)
	for i := 0; i < recordCount; i++ {
		records[i] = &map[string]any{
			"id":   300 + i,
			"name": fmt.Sprintf("Record%d", i),
		}
	}

	// 执行BatchInsertWithSizeNoInc并跳过版本号
	batchSize := 3
	ids, err := table.BatchInsertWithSizeNoInc(records, batchSize, true)
	if err != nil {
		t.Fatalf("BatchInsertWithSizeNoInc with skipVersion failed: %v", err)
	}

	// 验证插入结果
	if len(ids) != recordCount {
		t.Errorf("Expected %d IDs, got %d", recordCount, len(ids))
	}

	// 验证指定的ID是否正确返回
	for i, id := range ids {
		expectedID := 300 + i
		if id != expectedID {
			t.Errorf("Expected ID %d at index %d, got %d", expectedID, i, id)
		}
	}

	t.Logf("BatchInsertWithSizeNoInc with skipVersion test passed")
}
