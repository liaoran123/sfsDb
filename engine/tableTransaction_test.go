package engine

import (
	"fmt"
	"testing"
	"time"

	"github.com/liaoran123/sfsDb/util"
)

// createTestTable 创建测试表并设置索引
func createTestTable(t *testing.T) *Table {
	// 创建测试表，使用唯一的表名，避免测试之间的相互影响
	tableName := fmt.Sprintf("test_transaction_basic_%d", time.Now().UnixNano())
	table, err := TableNew(tableName)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{"id": 0, "name": "", "age": 0, "score": 0.0, "active": false}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键索引
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("id")
	err = table.CreateIndex(pk)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// 创建age索引
	ageIdx, err := DefaultNormalIndexNew("age_idx")
	if err != nil {
		t.Fatalf("Failed to create age index: %v", err)
	}
	ageIdx.AddFields("age")
	err = table.CreateIndex(ageIdx)
	if err != nil {
		t.Fatalf("Failed to create age index: %v", err)
	}

	// 创建score索引
	scoreIdx, err := DefaultNormalIndexNew("score_idx")
	if err != nil {
		t.Fatalf("Failed to create score index: %v", err)
	}
	scoreIdx.AddFields("score")
	err = table.CreateIndex(scoreIdx)
	if err != nil {
		t.Fatalf("Failed to create score index: %v", err)
	}

	// 创建普通索引
	activeIdx, err := DefaultNormalIndexNew("active_idx")
	if err != nil {
		t.Fatalf("Failed to create active index: %v", err)
	}
	activeIdx.AddFields("active")
	err = table.CreateIndex(activeIdx)
	if err != nil {
		t.Fatalf("Failed to create active index: %v", err)
	}

	return table
}

// TestTableTransaction_BasicOperations 测试事务的基本操作
func TestTableTransaction_BasicOperations(t *testing.T) {
	// 测试场景1: 基本的事务操作
	t.Run("BasicTransaction", func(t *testing.T) {
		// 创建测试表
		table := createTestTable(t)

		// 开始事务
		tx, err := table.Begin()
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}

		// 插入记录
		insertData := map[string]any{"id": 1, "name": "Alice", "age": 20, "score": 85.5, "active": true}
		id, err := tx.Insert(&insertData)
		if err != nil {
			t.Fatalf("Failed to insert record in transaction: %v", err)
		}
		if id != 1 {
			t.Errorf("Expected id 1, got %d", id)
		}

		// 读取记录（使用tx.Read可以看到事务中插入的记录）
		readData := map[string]any{"id": 1}
		record, err := tx.Read(&readData)
		if err != nil {
			t.Fatalf("Failed to read record in transaction: %v", err)
		}
		if len(record) == 0 {
			t.Errorf("Expected non-empty record, got empty")
		}

		// 更新记录
		updateData := map[string]any{"id": 1, "age": 21, "score": 90.0}
		err = tx.Update(&updateData)
		if err != nil {
			t.Fatalf("Failed to update record in transaction: %v", err)
		}

		// 提交事务
		err = tx.Commit()
		if err != nil {
			t.Fatalf("Failed to commit transaction: %v", err)
		}

		// 验证提交后的结果
		searchData := map[string]any{"id": nil}
		iterAfter := table.Search(&searchData)
		if iterAfter == nil {
			t.Fatalf("Failed to search records after transaction")
		}
		defer iterAfter.Release()

		recordsAfter := iterAfter.GetRecords(true)
		if len(recordsAfter) != 1 {
			t.Errorf("Expected 1 record after commit, got %d", len(recordsAfter))
		}
	})

	// 测试场景2: 事务回滚
	t.Run("TransactionRollback", func(t *testing.T) {
		// 创建测试表
		table := createTestTable(t)

		// 开始事务
		tx, err := table.Begin()
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}

		// 插入记录
		insertData := map[string]any{"id": 1, "name": "Alice", "age": 20, "score": 85.5, "active": true}
		_, err = tx.Insert(&insertData)
		if err != nil {
			t.Fatalf("Failed to insert record in transaction: %v", err)
		}

		// 提交事务
		err = tx.Commit()
		if err != nil {
			t.Fatalf("Failed to commit transaction: %v", err)
		}

		// 开始新事务
		tx2, err := table.Begin()
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}

		// 插入记录
		insertData2 := map[string]any{"id": 2, "name": "Bob", "age": 25, "score": 80.0, "active": true}
		_, err = tx2.Insert(&insertData2)
		if err != nil {
			t.Fatalf("Failed to insert record in transaction: %v", err)
		}

		// 回滚事务
		err = tx2.Rollback()
		if err != nil {
			t.Fatalf("Failed to rollback transaction: %v", err)
		}

		// 验证回滚后的结果
		searchData := map[string]any{"id": nil}
		iter := table.Search(&searchData)
		if iter == nil {
			t.Fatalf("Failed to search records after rollback")
		}
		defer iter.Release()

		records := iter.GetRecords(true)
		if len(records) != 1 {
			t.Errorf("Expected 1 record after rollback, got %d", len(records))
		}
	})

	// 测试场景3: 事务中的搜索操作
	t.Run("TransactionSearch", func(t *testing.T) {
		// 创建测试表
		table := createTestTable(t)

		// 开始事务
		tx, err := table.Begin()
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}

		// 插入记录
		insertData := map[string]any{"id": 1, "name": "Alice", "age": 20, "score": 85.5, "active": true}
		_, err = tx.Insert(&insertData)
		if err != nil {
			t.Fatalf("Failed to insert record in transaction: %v", err)
		}

		// 提交事务
		err = tx.Commit()
		if err != nil {
			t.Fatalf("Failed to commit transaction: %v", err)
		}

		// 开始新事务
		tx2, err := table.Begin()
		if err != nil {
			t.Fatalf("Failed to begin transaction 2: %v", err)
		}

		// 测试搜索所有记录
		allData := map[string]any{"id": nil}
		iterAll := tx2.Search(&allData)
		if iterAll == nil {
			t.Fatalf("Failed to search all records in transaction")
		}
		defer iterAll.Release()

		allRecords := iterAll.GetRecords(true)
		if len(allRecords) != 1 { // 只有Alice一条记录
			t.Errorf("Expected 1 record, got %d", len(allRecords))
		}

		// 测试条件搜索
		activeData := map[string]any{"active": true}
		iterActive := tx2.Search(&activeData)
		if iterActive == nil {
			t.Fatalf("Failed to search active records in transaction")
		}
		defer iterActive.Release()

		activeRecords := iterActive.GetRecords(true)
		if len(activeRecords) != 1 { // 只有Alice一条active记录
			t.Errorf("Expected 1 active record, got %d", len(activeRecords))
		}

		// 提交事务
		err = tx2.Commit()
		if err != nil {
			t.Fatalf("Failed to commit transaction: %v", err)
		}
	})

	// 测试场景4: 事务中的读一致性
	t.Run("TransactionReadConsistency", func(t *testing.T) {
		// 创建测试表
		table := createTestTable(t)

		// 开始事务
		tx, err := table.Begin()
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}

		// 插入记录
		insertData := map[string]any{"id": 1, "name": "Alice", "age": 20, "score": 85.5, "active": true}
		_, err = tx.Insert(&insertData)
		if err != nil {
			t.Fatalf("Failed to insert record in transaction: %v", err)
		}

		// 提交事务
		err = tx.Commit()
		if err != nil {
			t.Fatalf("Failed to commit transaction: %v", err)
		}

		// 开始第一个事务
		tx1, err := table.Begin()
		if err != nil {
			t.Fatalf("Failed to begin transaction 1: %v", err)
		}

		// 读取记录
		readData := map[string]any{"id": 1}
		record1, err := tx1.Read(&readData)
		if err != nil {
			t.Fatalf("Failed to read record in transaction 1: %v", err)
		}

		// 开始第二个事务并更新记录
		tx2, err := table.Begin()
		if err != nil {
			t.Fatalf("Failed to begin transaction 2: %v", err)
		}

		updateData := map[string]any{"id": 1, "name": "Alice Updated"}
		err = tx2.Update(&updateData)
		if err != nil {
			t.Fatalf("Failed to update record in transaction 2: %v", err)
		}

		// 提交第二个事务
		err = tx2.Commit()
		if err != nil {
			t.Fatalf("Failed to commit transaction 2: %v", err)
		}

		// 在第一个事务中再次读取记录，应该看到原始值（读一致性）
		record1Again, err := tx1.Read(&readData)
		if err != nil {
			t.Fatalf("Failed to read record again in transaction 1: %v", err)
		}

		// 验证两次读取的结果相同（读一致性）
		if string(record1) != string(record1Again) {
			t.Errorf("Read consistency violated: first read %s, second read %s", string(record1), string(record1Again))
		}

		// 提交第一个事务
		err = tx1.Commit()
		if err != nil {
			t.Fatalf("Failed to commit transaction 1: %v", err)
		}

		// 验证最终值
		tx3, err := table.Begin()
		if err != nil {
			t.Fatalf("Failed to begin transaction 3: %v", err)
		}

		recordFinal, err := tx3.Read(&readData)
		if err != nil {
			t.Fatalf("Failed to read record in transaction 3: %v", err)
		}

		if string(recordFinal) == string(record1) {
			t.Errorf("Expected updated record, got original: %s", string(recordFinal))
		}

		// 提交第三个事务
		err = tx3.Commit()
		if err != nil {
			t.Fatalf("Failed to commit transaction 3: %v", err)
		}
	})
}

// TestTableTransaction_ErrorHandling 测试事务的错误处理
func TestTableTransaction_ErrorHandling(t *testing.T) {
	// 创建测试表，使用唯一的表名，避免测试之间的相互影响
	tableName := fmt.Sprintf("test_transaction_error_%d", time.Now().UnixNano())
	table, err := TableNew(tableName)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{"id": 0, "name": "", "age": 0}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键索引
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("id")
	err = table.CreateIndex(pk)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// 测试场景1: 重复提交
	t.Run("DuplicateCommit", func(t *testing.T) {
		tx, err := table.Begin()
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}

		// 提交事务
		err = tx.Commit()
		if err != nil {
			t.Fatalf("Failed to commit transaction: %v", err)
		}

		// 再次提交，应该返回错误
		err = tx.Commit()
		if err == nil {
			t.Errorf("Expected error on duplicate commit, got nil")
		}
	})

	// 测试场景2: 重复回滚
	t.Run("DuplicateRollback", func(t *testing.T) {
		tx, err := table.Begin()
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}

		// 回滚事务
		err = tx.Rollback()
		if err != nil {
			t.Fatalf("Failed to rollback transaction: %v", err)
		}

		// 再次回滚，应该返回错误
		err = tx.Rollback()
		if err == nil {
			t.Errorf("Expected error on duplicate rollback, got nil")
		}
	})

	// 测试场景3: 在已提交的事务中操作
	t.Run("OperationAfterCommit", func(t *testing.T) {
		tx, err := table.Begin()
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}

		// 提交事务
		err = tx.Commit()
		if err != nil {
			t.Fatalf("Failed to commit transaction: %v", err)
		}

		// 尝试在已提交的事务中插入记录
		insertData := map[string]any{"id": 1, "name": "Alice", "age": 20}
		_, err = tx.Insert(&insertData)
		if err == nil {
			t.Errorf("Expected error on insert after commit, got nil")
		}

		// 尝试在已提交的事务中搜索
		searchData := map[string]any{"id": nil}
		iter := tx.Search(&searchData)
		if iter != nil {
			t.Errorf("Expected nil iterator on search after commit, got non-nil")
		}
	})
}

// TestTableTransaction_SearchWithOperators 测试事务中的搜索操作（带操作符）
func TestTableTransaction_SearchWithOperators(t *testing.T) {
	// 创建测试表，使用唯一的表名，避免测试之间的相互影响
	tableName := fmt.Sprintf("test_transaction_search_op_%d", time.Now().UnixNano())
	table, err := TableNew(tableName)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{"id": 0, "name": "", "age": 0, "score": 0.0}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键索引
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("id")
	err = table.CreateIndex(pk)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// 创建age索引
	ageIdx, _ := DefaultNormalIndexNew("age_idx")
	ageIdx.AddFields("age")
	err = table.CreateIndex(ageIdx)
	if err != nil {
		t.Fatalf("Failed to create age index: %v", err)
	}

	// 创建score索引
	scoreIdx, _ := DefaultNormalIndexNew("score_idx")
	scoreIdx.AddFields("score")
	err = table.CreateIndex(scoreIdx)
	if err != nil {
		t.Fatalf("Failed to create score index: %v", err)
	}

	// 插入测试数据
	testData := []map[string]any{
		{"id": 1, "name": "Alice", "age": 20, "score": 85.5},
		{"id": 2, "name": "Bob", "age": 25, "score": 90.0},
		{"id": 3, "name": "Charlie", "age": 30, "score": 75.5},
		{"id": 4, "name": "David", "age": 35, "score": 95.0},
		{"id": 5, "name": "Eve", "age": 40, "score": 80.0},
	}

	for _, data := range testData {
		_, err := table.Insert(&data)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// 开始事务
	tx, err := table.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// 测试场景1: 大于操作
	t.Run("GreaterThanSearch", func(t *testing.T) {
		searchData := map[string]any{"age": 25}
		iter := tx.Search(&searchData, util.GreaterThan)
		if iter == nil {
			t.Fatalf("Failed to search records with age > 25")
		}
		defer iter.Release()

		records := iter.GetRecords(true)
		if len(records) != 3 { // Charlie, David, Eve
			t.Errorf("Expected 3 records with age > 25, got %d", len(records))
		}
	})

	// 测试场景2: 小于等于操作
	t.Run("LessThanOrEqualSearch", func(t *testing.T) {
		searchData := map[string]any{"score": 85.5}
		iter := tx.Search(&searchData, util.LessThanOrEqual)
		if iter == nil {
			t.Fatalf("Failed to search records with score <= 85.5")
		}
		defer iter.Release()

		records := iter.GetRecords(true)
		if len(records) != 3 { // Alice, Charlie, Eve
			t.Errorf("Expected 3 records with score <= 85.5, got %d", len(records))
		}
	})

	// 提交事务
	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}
}
