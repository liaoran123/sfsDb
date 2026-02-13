package engine

import (
	"fmt"
	"testing"
	"time"

	"github.com/liaoran123/sfsDb/util"
)

// createTestTableWithBoolOnly 创建只包含bool类型字段的测试表并设置索引
func createTestTableWithBoolOnly(t *testing.T) *Table {
	// 创建测试表，使用唯一的表名，避免测试之间的相互影响
	tableName := fmt.Sprintf("test_bool_index_%d", time.Now().UnixNano())
	table, err := TableNew(tableName)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{
		"id":     0,     // 整数类型
		"active": false, // 布尔类型
	}
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

	// 创建active索引（布尔类型）
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

// TestBoolIndex_CreationAndQuery 测试布尔类型索引的创建和查询过程
func TestBoolIndex_CreationAndQuery(t *testing.T) {
	// 创建测试表
	table := createTestTableWithBoolOnly(t)

	// 开始事务
	tx, err := table.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// 插入测试数据
	records := []map[string]any{
		{"id": 1, "active": true},
		{"id": 2, "active": true},
		{"id": 3, "active": false},
		{"id": 4, "active": true},
		{"id": 5, "active": false},
	}

	// 插入记录并分析索引创建
	t.Logf("Inserting records...")
	for i, record := range records {
		id, err := tx.Insert(&record)
		if err != nil {
			t.Fatalf("Failed to insert record %d: %v", i, err)
		}
		t.Logf("Inserted record %d with id %d, active=%v", i, id, record["active"])

		// 分析索引键的生成
		fieldsBytes := table.FieldsToBytes(&record)
		activeIdx, _ := DefaultNormalIndexNew("active_idx")
		activeIdx.AddFields("active")
		idxKey := activeIdx.JoinValue(fieldsBytes, table.id)
		t.Logf("Index key for record %d (active=%v): %v", i, record["active"], idxKey)
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// 测试1: 分析布尔类型索引的范围查询
	t.Run("BoolIndexRangeAnalysis", func(t *testing.T) {
		// 获取active索引
		var activeIdx Index
		for _, idx := range table.indexs.GetAllIndexes() {
			if idx.Name() == "active_idx" {
				activeIdx = idx
				break
			}
		}
		if activeIdx == nil {
			t.Fatalf("Active index not found")
		}

		// 测试true值的范围
		trueFields := map[string]any{"active": true}
		trueFieldsBytes := table.FieldsToBytes(&trueFields)
		trueKey := activeIdx.JoinValue(trueFieldsBytes, table.id)
		pfx := activeIdx.Prefix(table.id)
		pfx = append(pfx, SPLIT[0])
		rh := util.NewRangeHelper(pfx)

		// 测试Equal操作符
		equalRange := rh.FromComparison(util.Equal, trueKey)
		t.Logf("Equal(true) range: Start=%v, Limit=%v", equalRange.Start, equalRange.Limit)

		// 测试GreaterThan操作符
		gtRange := rh.FromComparison(util.GreaterThan, trueKey)
		t.Logf("GreaterThan(true) range: Start=%v, Limit=%v", gtRange.Start, gtRange.Limit)

		// 测试LessThan操作符
		ltRange := rh.FromComparison(util.LessThan, trueKey)
		t.Logf("LessThan(true) range: Start=%v, Limit=%v", ltRange.Start, ltRange.Limit)

		// 测试false值的范围
		falseFields := map[string]any{"active": false}
		falseFieldsBytes := table.FieldsToBytes(&falseFields)
		falseKey := activeIdx.JoinValue(falseFieldsBytes, table.id)

		// 测试Equal操作符
		equalFalseRange := rh.FromComparison(util.Equal, falseKey)
		t.Logf("Equal(false) range: Start=%v, Limit=%v", equalFalseRange.Start, equalFalseRange.Limit)

		// 测试GreaterThan操作符
		gtFalseRange := rh.FromComparison(util.GreaterThan, falseKey)
		t.Logf("GreaterThan(false) range: Start=%v, Limit=%v", gtFalseRange.Start, gtFalseRange.Limit)

		// 测试LessThan操作符
		ltFalseRange := rh.FromComparison(util.LessThan, falseKey)
		t.Logf("LessThan(false) range: Start=%v, Limit=%v", ltFalseRange.Start, ltFalseRange.Limit)
	})

	// 测试2: 直接使用布尔类型索引查询
	t.Run("DirectBoolIndexQuery", func(t *testing.T) {
		// 测试active=true
		t.Run("ActiveTrue", func(t *testing.T) {
			searchData := map[string]any{"active": true}
			iter, err := table.Search(&searchData, util.Equal)
			if err != nil {
				t.Fatalf("Search 失败: %v", err)
			}
			if err != nil && iter == nil {
				t.Fatalf("Failed to search active=true: %v", err)
			}
			defer GlobalTableIterPool.Put(iter)

			// 遍历迭代器并打印所有记录
			t.Logf("Records for active=true:")
			count := 0
			if iter.First() {
				count++
				key, value := iter.Key(), iter.Value()
				t.Logf("Record %d: Key=%v, Value=%v", count, key, value)
				for iter.Next() {
					count++
					key, value := iter.Key(), iter.Value()
					t.Logf("Record %d: Key=%v, Value=%v", count, key, value)
				}
			}
			t.Logf("Total records for active=true: %d", count)

			// 使用GetRecords获取记录
			iter.First() // 重置迭代器
			results := iter.GetRecords(true)
			t.Logf("GetRecords results for active=true: %v", results)
			t.Logf("GetRecords returned %d records", len(results))
		})

		// 测试active=false
		t.Run("ActiveFalse", func(t *testing.T) {
			searchData := map[string]any{"active": false}
			iter, err := table.Search(&searchData, util.Equal)
			if err != nil {
				t.Fatalf("Search 失败: %v", err)
			}
			if err != nil && iter == nil {
				t.Fatalf("Failed to search active=false: %v", err)
			}
			defer GlobalTableIterPool.Put(iter)

			// 遍历迭代器并打印所有记录
			t.Logf("Records for active=false:")
			count := 0
			if iter.First() {
				count++
				key, value := iter.Key(), iter.Value()
				t.Logf("Record %d: Key=%v, Value=%v", count, key, value)
				for iter.Next() {
					count++
					key, value := iter.Key(), iter.Value()
					t.Logf("Record %d: Key=%v, Value=%v", count, key, value)
				}
			}
			t.Logf("Total records for active=false: %d", count)

			// 使用GetRecords获取记录
			iter.First() // 重置迭代器
			results := iter.GetRecords(true)
			t.Logf("GetRecords results for active=false: %v", results)
			t.Logf("GetRecords returned %d records", len(results))
		})
	})

	// 测试3: 使用主键索引获取所有记录并过滤
	t.Run("PrimaryKeyWithFilter", func(t *testing.T) {
		// 使用主键索引获取所有记录
		searchData := map[string]any{"id": nil}
		iter, err := table.Search(&searchData)
		if err != nil {
			t.Fatalf("Search 失败: %v", err)
		}
		if iter == nil {
			t.Fatalf("Failed to search all records: %v", err)
		}
		defer GlobalTableIterPool.Put(iter)

		// 获取所有记录
		allResults := iter.GetRecords(true)
		t.Logf("All records: %v", allResults)

		// 过滤active=true的记录
		var activeTrueResults []map[string]any
		var activeFalseResults []map[string]any
		for _, record := range allResults {
			if active, ok := record["active"].(bool); ok {
				if active {
					activeTrueResults = append(activeTrueResults, record)
				} else {
					activeFalseResults = append(activeFalseResults, record)
				}
			}
		}

		t.Logf("Filtered active=true records: %v", activeTrueResults)
		t.Logf("Filtered active=false records: %v", activeFalseResults)

		// 验证结果数量
		if len(activeTrueResults) != 3 {
			t.Errorf("Expected 3 active=true records, got %d", len(activeTrueResults))
		}
		if len(activeFalseResults) != 2 {
			t.Errorf("Expected 2 active=false records, got %d", len(activeFalseResults))
		}
	})
}
