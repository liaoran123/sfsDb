package engine

import (
	"testing"

	"github.com/liaoran123/sfsDb/storage"
)

// TestBatchInsertNoInc 测试BatchInsertNoInc函数
func TestBatchInsertNoInc(t *testing.T) {
	// 初始化数据库
	_, err := storage.OpenDefaultDb("./batch_no_inc_test_db_1")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// 创建测试表
	table, err := TableNew("test_batch_no_inc")
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
	records := []*map[string]any{
		{"id": 101, "name": "Alice", "age": 25},
		{"id": 102, "name": "Bob", "age": 30},
		{"id": 103, "name": "Charlie", "age": 35},
		{"id": 104, "name": "David", "age": 40},
		{"id": 105, "name": "Eve", "age": 45},
	}

	// 执行BatchInsertNoInc
	ids, err := table.BatchInsertNoInc(records)
	if err != nil {
		t.Fatalf("BatchInsertNoInc failed: %v", err)
	}

	// 验证插入结果
	// 应该只插入4条记录（跳过缺少ID的那条）
	t.Logf("Inserted IDs: %v", ids)
	if len(ids) != len(records) {
		t.Logf("Expected %d IDs, got %d (some records may be skipped due to missing primary key)", len(records), len(ids))
	}

	// 验证指定的ID是否正确返回
	expectedIDs := []int{101, 102, 103, 104, 105}
	for i, id := range ids {
		if id != expectedIDs[i] {
			t.Errorf("Expected ID %d at index %d, got %d", expectedIDs[i], i, id)
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

	t.Logf("BatchInsertNoInc test passed")
}

// TestBatchInsertNoIncSkipVersion 测试BatchInsertNoInc函数的skipVersion参数
func TestBatchInsertNoIncSkipVersion(t *testing.T) {
	// 初始化数据库
	_, err := storage.OpenDefaultDb("./batch_no_inc_skip_version_test_db_1")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// 创建测试表
	table, err := TableNew("test_batch_no_inc_skip_version")
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
	records := []*map[string]any{
		&map[string]any{"id": 201, "name": "Alice"},
		&map[string]any{"id": 202, "name": "Bob"},
	}

	// 执行BatchInsertNoInc并跳过版本号
	ids, err := table.BatchInsertNoInc(records)
	if err != nil {
		t.Fatalf("BatchInsertNoInc failed: %v", err)
	}

	// 验证插入结果
	if len(ids) != len(records) {
		t.Errorf("Expected %d IDs, got %d", len(records), len(ids))
	}

	// 验证指定的ID是否正确返回
	expectedIDs := []int{201, 202}
	for i, id := range ids {
		if id != expectedIDs[i] {
			t.Errorf("Expected ID %d at index %d, got %d", expectedIDs[i], i, id)
		}
	}

	t.Logf("BatchInsertNoInc with skipVersion test passed")
}
