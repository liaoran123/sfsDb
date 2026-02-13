package engine

import (
	"testing"
)

// TestTableTransaction_MultipleInsert 使用tx.Read()方法验证事务中插入的多条记录
func TestTableTransaction_MultipleInsert(t *testing.T) {
	// 创建测试表
	table := createTestTable(t)

	// 开始事务
	tx, err := table.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// 插入多条记录
	records := []map[string]any{
		{"id": 1, "name": "Alice", "age": 20, "score": 85.5, "active": true},
		{"id": 2, "name": "Bob", "age": 25, "score": 90.0, "active": true},
		{"id": 3, "name": "Charlie", "age": 30, "score": 75.5, "active": false},
		{"id": 4, "name": "David", "age": 35, "score": 95.0, "active": true},
		{"id": 5, "name": "Eve", "age": 40, "score": 80.0, "active": false},
	}

	// 插入记录并使用tx.Read()验证
	for _, record := range records {
		// 插入记录
		id, err := tx.Insert(&record)
		if err != nil {
			t.Fatalf("Failed to insert record in transaction: %v", err)
		}

		// 验证返回的ID正确
		expectedID := record["id"].(int)
		if id != expectedID {
			t.Errorf("Expected id %d, got %d", expectedID, id)
		}

		// 使用tx.Read()验证记录被正确插入
		readData := map[string]any{"id": id}
		readRecord, err := tx.Read(&readData)
		if err != nil {
			t.Fatalf("Failed to read record in transaction: %v", err)
		}
		if len(readRecord) == 0 {
			t.Errorf("Expected non-empty record for id %d, got empty", id)
		}

		t.Logf("Successfully inserted and read record with id %d", id)
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// 提交后使用表的Search方法验证所有记录都被持久化
	searchData := map[string]any{"id": nil}
	iter, _ := table.Search(&searchData)
	if iter == nil {
		t.Fatalf("Failed to search records after commit")
	}
	defer GlobalTableIterPool.Put(iter)

	allRecords := iter.GetRecords(true)
	if len(allRecords) != len(records) {
		t.Errorf("Expected %d records after commit, got %d", len(records), len(allRecords))
	}

	t.Logf("Successfully committed and verified %d records", len(allRecords))
}
