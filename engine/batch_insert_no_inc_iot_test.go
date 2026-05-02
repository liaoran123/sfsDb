package engine

import (
	"testing"
)

// TestBatchInsertNoIncIoT 测试BatchInsertNoIncIoT函数
func TestBatchInsertNoIncIoT(t *testing.T) {

	// 创建测试表
	table, err := NewTable("test_batch_no_inc_iot")
	if err != nil {
		t.Fatalf("NewTable failed: %v", err)
	}

	// 设置字段
	fields := map[string]any{"id": 0, "name": "", "age": 0}
	table.SetFields(fields)

	// 创建主键
	pk, _ := NewDefaultPrimaryKey("pk")
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

	// 执行BatchInsertNoIncIoT
	ids, err := table.BatchInsertNoIncIoT(records)
	if err != nil {
		t.Fatalf("BatchInsertNoIncIoT failed: %v", err)
	}

	// 验证插入结果
	t.Logf("Inserted IDs: %v", ids)
	if len(ids) != len(records) {
		t.Logf("Expected %d IDs, got %d", len(records), len(ids))
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

	t.Logf("BatchInsertNoIncIoT test passed")
}
