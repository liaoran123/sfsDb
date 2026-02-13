package engine

import (
	"testing"

	"github.com/liaoran123/sfsDb/util"
)

// TestTableTransaction_CommitQuery 测试提交事务后，正常查询是否正确
func TestTableTransaction_CommitQuery(t *testing.T) {
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

	// 提交事务后，使用表的Search方法进行正常查询

	// 测试1: 查询所有记录
	t.Run("QueryAllRecords", func(t *testing.T) {
		searchData := map[string]any{"id": nil}
		iter, _ := table.Search(&searchData)
		if iter == nil {
			t.Fatalf("Failed to search all records")
		}
		defer GlobalTableIterPool.Put(iter)

		allRecords := iter.GetRecords(true)
		if len(allRecords) != len(records) {
			t.Errorf("Expected %d records, got %d", len(records), len(allRecords))
		}

		t.Logf("Successfully queried all %d records", len(allRecords))
	})

	// 测试2: 按条件查询（active=true）
	t.Run("QueryActiveRecords", func(t *testing.T) {
		// 首先使用主键索引获取所有记录，验证数据是否正确插入
		searchData := map[string]any{"id": nil}
		iter, _ := table.Search(&searchData)
		if iter == nil {
			t.Fatalf("Failed to search all records")
		}
		defer GlobalTableIterPool.Put(iter)

		allRecords := iter.GetRecords(true)
		// 打印所有记录，以便调试
		for i, record := range allRecords {
			t.Logf("All Record %d: %v", i, record)
		}

		// 过滤active=true的记录
		var activeRecords []map[string]any
		for _, record := range allRecords {
			if active, ok := record["active"].(bool); ok && active {
				activeRecords = append(activeRecords, record)
			}
		}

		expectedActiveCount := 3 // Alice, Bob, David
		if len(activeRecords) != expectedActiveCount {
			t.Errorf("Expected %d active records, got %d", expectedActiveCount, len(activeRecords))
		} else {
			for i, record := range activeRecords {
				t.Logf("Active Record %d: %v", i, record)
			}
		}

		t.Logf("Successfully queried %d active records", len(activeRecords))
	})

	// 测试3: 按条件查询（age > 30）
	t.Run("QueryAgeGreaterThan30", func(t *testing.T) {
		searchData := map[string]any{"age": 30}
		iter, _ := table.Search(&searchData, util.GreaterThan)
		if iter == nil {
			t.Fatalf("Failed to search records with age > 30")
		}
		defer GlobalTableIterPool.Put(iter)

		ageRecords := iter.GetRecords(true)
		expectedAgeCount := 2 // David, Eve
		if len(ageRecords) != expectedAgeCount {
			t.Errorf("Expected %d records with age > 30, got %d", expectedAgeCount, len(ageRecords))
		}

		t.Logf("Successfully queried %d records with age > 30", len(ageRecords))
	})

	// 测试4: 按条件查询（score >= 90）
	t.Run("QueryScoreGreaterThanOrEqual90", func(t *testing.T) {
		searchData := map[string]any{"score": 90.0}
		iter, _ := table.Search(&searchData, util.GreaterThanOrEqual)
		if iter == nil {
			t.Fatalf("Failed to search records with score >= 90")
		}
		defer GlobalTableIterPool.Put(iter)

		scoreRecords := iter.GetRecords(true)
		expectedScoreCount := 2 // Bob, David
		if len(scoreRecords) != expectedScoreCount {
			t.Errorf("Expected %d records with score >= 90, got %d", expectedScoreCount, len(scoreRecords))
		}

		t.Logf("Successfully queried %d records with score >= 90", len(scoreRecords))
	})

	// 测试5: 使用事务查询（提交后）
	t.Run("QueryWithTransaction", func(t *testing.T) {
		// 开始新事务
		tx2, err := table.Begin()
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}

		// 在事务中查询所有记录
		searchData := map[string]any{"id": nil}
		iter, _ := tx2.Search(&searchData)
		if iter == nil {
			t.Fatalf("Failed to search all records in transaction")
		}
		defer GlobalTableIterPool.Put(iter)

		allRecords := iter.GetRecords(true)
		if len(allRecords) != len(records) {
			t.Errorf("Expected %d records in transaction, got %d", len(records), len(allRecords))
		}

		t.Logf("Successfully queried all %d records in transaction", len(allRecords))

		// 提交事务
		err = tx2.Commit()
		if err != nil {
			t.Fatalf("Failed to commit transaction: %v", err)
		}
	})
}
