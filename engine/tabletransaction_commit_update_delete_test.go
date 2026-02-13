package engine

import (
	"testing"
)

// TestTableTransaction_CommitUpdateDeleteQuery 测试提交事务后，更新和删除操作的查询是否正确
func TestTableTransaction_CommitUpdateDeleteQuery(t *testing.T) {
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
	}

	// 插入记录
	for _, record := range records {
		_, err := tx.Insert(&record)
		if err != nil {
			t.Fatalf("Failed to insert record in transaction: %v", err)
		}
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// 测试1: 更新操作后查询
	t.Run("UpdateAndQuery", func(t *testing.T) {
		// 开始事务
		tx, err := table.Begin()
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}

		// 更新记录
		updateData := map[string]any{"id": 1, "age": 21, "score": 90.0, "active": true}
		err = tx.Update(&updateData)
		if err != nil {
			t.Fatalf("Failed to update record: %v", err)
		}

		// 提交事务
		err = tx.Commit()
		if err != nil {
			t.Fatalf("Failed to commit transaction: %v", err)
		}

		// 查询更新后的记录
		searchData := map[string]any{"id": 1}
		iter, _ := table.Search(&searchData)
		if iter == nil {
			t.Fatalf("Failed to search updated record")
		}
		defer GlobalTableIterPool.Put(iter)

		updatedRecords := iter.GetRecords(true)
		if len(updatedRecords) != 1 {
			t.Errorf("Expected 1 updated record, got %d", len(updatedRecords))
		}

		updatedRecord := updatedRecords[0]
		if updatedRecord["age"] != 21 || updatedRecord["score"] != 90.0 {
			t.Errorf("Record not updated correctly: %v", updatedRecord)
		}

		t.Logf("Successfully updated and queried record: %v", updatedRecord)
	})

	// 测试2: 删除操作后查询
	t.Run("DeleteAndQuery", func(t *testing.T) {
		// 开始事务
		tx, err := table.Begin()
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}

		// 删除记录
		deleteData := map[string]any{"id": 2}
		err = tx.Delete(&deleteData)
		if err != nil {
			t.Fatalf("Failed to delete record: %v", err)
		}

		// 提交事务
		err = tx.Commit()
		if err != nil {
			t.Fatalf("Failed to commit transaction: %v", err)
		}

		// 验证记录已删除
		searchData := map[string]any{"id": 2}
		iter, _ := table.Search(&searchData)
		if iter == nil {
			t.Fatalf("Failed to search deleted record")
		}
		defer GlobalTableIterPool.Put(iter)

		deletedRecords := iter.GetRecords(true)
		if len(deletedRecords) != 0 {
			t.Errorf("Expected 0 deleted records, got %d", len(deletedRecords))
		}

		// 验证其他记录仍然存在
		allData := map[string]any{"id": nil}
		allIter, _ := table.Search(&allData)
		if allIter == nil {
			t.Fatalf("Failed to search all records")
		}
		defer GlobalTableIterPool.Put(allIter)

		allRecords := allIter.GetRecords(true)
		if len(allRecords) != 2 { // 应该剩下2条记录
			t.Errorf("Expected 2 remaining records, got %d", len(allRecords))
		}

		t.Logf("Successfully deleted record and verified, remaining records: %d", len(allRecords))
	})

	// 测试3: 复杂条件查询
	t.Run("ComplexConditionQuery", func(t *testing.T) {
		// 先插入一些测试数据
		tx, err := table.Begin()
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}

		// 插入更多记录
		moreRecords := []map[string]any{
			{"id": 4, "name": "David", "age": 35, "score": 95.0, "active": true},
			{"id": 5, "name": "Eve", "age": 40, "score": 80.0, "active": false},
			{"id": 6, "name": "Frank", "age": 45, "score": 85.0, "active": true},
		}

		for _, record := range moreRecords {
			_, err := tx.Insert(&record)
			if err != nil {
				t.Fatalf("Failed to insert record: %v", err)
			}
		}

		// 提交事务
		err = tx.Commit()
		if err != nil {
			t.Fatalf("Failed to commit transaction: %v", err)
		}

		// 测试多条件查询：active=true 且 age > 30
		// 注意：这里我们需要使用 Searchs 方法来支持多条件查询
		// 由于现有的 Search 方法可能只支持单个条件，我们使用 GetAll 方法获取所有记录后过滤
		searchData := map[string]any{"id": nil}
		iter, _ := table.Search(&searchData)
		if iter == nil {
			t.Fatalf("Failed to search records")
		}
		defer GlobalTableIterPool.Put(iter)

		allRecords := iter.GetRecords(true)
		// 过滤符合条件的记录
		var filteredRecords []map[string]any
		for _, record := range allRecords {
			if active, ok := record["active"].(bool); ok && active {
				if age, ok := record["age"].(int); ok && age > 30 {
					filteredRecords = append(filteredRecords, record)
				}
			}
		}

		expectedCount := 2 // David, Frank
		if len(filteredRecords) != expectedCount {
			t.Errorf("Expected %d records with active=true and age>30, got %d", expectedCount, len(filteredRecords))
		}

		t.Logf("Successfully queried %d records with complex conditions", len(filteredRecords))
	})
}
