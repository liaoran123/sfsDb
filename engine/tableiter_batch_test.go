package engine

import (
	"fmt"
	"testing"
	"time"

	"github.com/liaoran123/sfsDb/match"
)

// TestTableIter_BatchOperations 测试 TableIter 的批量操作功能
// 包括批量删除和批量更新
func TestTableIter_BatchOperations(t *testing.T) {
	// 创建测试表
	table, err := TableNew("test_batch_operations")
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}

	// 设置表字段
	fields := map[string]any{
		"id":     0,
		"name":   "",
		"age":    0,
		"score":  0.0,
		"active": false,
	}

	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("设置表字段失败: %v", err)
	}

	// 创建主键索引
	pkIndex, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		t.Fatalf("创建主键索引失败: %v", err)
	}
	pkIndex.AddFields("id")
	err = table.CreateIndex(pkIndex)
	if err != nil {
		t.Fatalf("创建主键索引失败: %v", err)
	}

	// 插入测试数据
	testData := []map[string]any{
		{"id": 1, "name": "Alice", "age": 20, "score": 85.5, "active": true},
		{"id": 2, "name": "Bob", "age": 25, "score": 90.0, "active": true},
		{"id": 3, "name": "Charlie", "age": 30, "score": 75.5, "active": false},
		{"id": 4, "name": "David", "age": 35, "score": 95.0, "active": true},
		{"id": 5, "name": "Eve", "age": 40, "score": 80.0, "active": false},
		{"id": 6, "name": "Frank", "age": 45, "score": 88.5, "active": true},
		{"id": 7, "name": "Grace", "age": 50, "score": 92.0, "active": true},
		{"id": 8, "name": "Henry", "age": 55, "score": 78.5, "active": false},
		{"id": 9, "name": "Ivy", "age": 60, "score": 83.0, "active": true},
		{"id": 10, "name": "Jack", "age": 65, "score": 87.5, "active": true},
	}

	for _, record := range testData {
		_, err = table.Insert(&record)
		if err != nil {
			t.Fatalf("插入测试数据失败: %v", err)
		}
	}

	// 测试批量删除功能
	t.Run("BatchDelete", func(t *testing.T) {
		// 获取迭代器并设置匹配条件
		iter, _ := table.Search(&map[string]any{"id": nil})
		if iter == nil {
			t.Fatalf("获取迭代器失败")
		}
		defer GlobalTableIterPool.Put(iter)

		// 设置匹配条件：删除 inactive 记录
		iter.SetMatch(match.NewFieldComparison("active", match.Equal, false))

		// 执行批量删除
		err = iter.Delete()
		if err != nil {
			t.Fatalf("批量删除失败: %v", err)
		}

		// 验证删除结果
		iterAfter, _ := table.Search(&map[string]any{"id": nil})
		if iterAfter != nil {
			defer GlobalTableIterPool.Put(iterAfter)
			iterAfter.SetMatch(match.NewFieldComparison("active", match.Equal, false))
			recordsAfter := iterAfter.GetRecords(true)
			if len(recordsAfter) != 0 {
				t.Errorf("删除后应该没有 inactive 记录，实际还有 %d 条", len(recordsAfter))
			} else {
				t.Logf("批量删除测试通过，成功删除所有 inactive 记录")
			}
		}
	})

	// 测试批量更新功能
	t.Run("BatchUpdate", func(t *testing.T) {
		// 插入更多测试数据
		moreData := []map[string]any{
			{"id": 11, "name": "Ken", "age": 22, "score": 75.0, "active": false},
			{"id": 12, "name": "Lily", "age": 28, "score": 82.0, "active": false},
			{"id": 13, "name": "Mike", "age": 33, "score": 91.0, "active": false},
			{"id": 14, "name": "Nancy", "age": 38, "score": 86.0, "active": false},
			{"id": 15, "name": "Oscar", "age": 43, "score": 79.0, "active": false},
		}

		for _, record := range moreData {
			_, err = table.Insert(&record)
			if err != nil {
				t.Fatalf("插入更多测试数据失败: %v", err)
			}
		}

		// 获取迭代器并设置匹配条件
		iter, _ := table.Search(&map[string]any{"id": nil})
		if iter == nil {
			t.Fatalf("获取迭代器失败")
		}
		defer GlobalTableIterPool.Put(iter)

		// 设置匹配条件：更新 age < 30 的记录
		iter.SetMatch(match.NewFieldComparison("age", match.LessThan, 30))

		// 准备更新字段
		updateFields := map[string]any{
			"active": true,
			"score":  95.0,
		}

		// 执行批量更新
		err = iter.Update(&updateFields)
		if err != nil {
			t.Fatalf("批量更新失败: %v", err)
		}

		// 验证更新结果
		iterAfter, _ := table.Search(&map[string]any{"id": nil})
		if iterAfter != nil {
			defer GlobalTableIterPool.Put(iterAfter)
			// 查找 age < 30 且 active = true 的记录
			iterAfter.SetMatch(match.NewAND(
				[]string{"age"},
				map[any]bool{22: true, 28: true}, // Ken 和 Lily
			))
			recordsAfter := iterAfter.GetRecords(true)
			if len(recordsAfter) != 2 {
				t.Errorf("更新后应该有 2 条 age < 30 的记录，实际有 %d 条", len(recordsAfter))
			} else {
				// 检查更新字段
				for _, record := range recordsAfter {
					if record["active"] != true {
						t.Errorf("记录 %v 的 active 字段应该为 true", record["id"])
					}
					if record["score"] != 95.0 {
						t.Errorf("记录 %v 的 score 字段应该为 95.0", record["id"])
					}
				}
				t.Logf("批量更新测试通过，成功更新所有 age < 30 的记录")
			}
		}
	})

	// 测试批量操作的限制功能
	t.Run("BatchOperationsWithLimit", func(t *testing.T) {
		// 获取迭代器
		iter, _ := table.Search(&map[string]any{"id": nil})
		if iter == nil {
			t.Fatalf("获取迭代器失败")
		}
		defer GlobalTableIterPool.Put(iter)

		// 准备更新字段
		updateFields := map[string]any{
			"score": 100.0,
		}

		// 执行批量更新，限制更新 3 条记录
		err = iter.Update(&updateFields, 3)
		if err != nil {
			t.Fatalf("带限制的批量更新失败: %v", err)
		}

		// 验证更新结果
		iterAfter, _ := table.Search(&map[string]any{"id": nil})
		if iterAfter != nil {
			defer GlobalTableIterPool.Put(iterAfter)
			// 查找 score = 100.0 的记录
			iterAfter.SetMatch(match.NewFieldComparison("score", match.Equal, 100.0))
			recordsAfter := iterAfter.GetRecords(true)
			if len(recordsAfter) != 3 {
				t.Errorf("更新后应该有 3 条 score = 100.0 的记录，实际有 %d 条", len(recordsAfter))
			} else {
				t.Logf("带限制的批量更新测试通过，成功更新 3 条记录")
			}
		}
	})
}

// TestTableIter_BatchPerformance 测试 TableIter 批量操作的性能
func TestTableIter_BatchPerformance(t *testing.T) {
	// 跳过性能测试，除非明确指定
	if testing.Short() {
		t.Skip("跳过性能测试")
	}

	// 创建测试表
	table, err := TableNew("test_batch_performance")
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}

	// 设置表字段
	fields := map[string]any{
		"id":     0,
		"name":   "",
		"age":    0,
		"active": false,
	}

	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("设置表字段失败: %v", err)
	}

	// 创建主键索引
	pkIndex, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		t.Fatalf("创建主键索引失败: %v", err)
	}
	pkIndex.AddFields("id")
	err = table.CreateIndex(pkIndex)
	if err != nil {
		t.Fatalf("创建主键索引失败: %v", err)
	}

	// 插入大量测试数据
	const recordCount = 1000
	t.Logf("插入 %d 条测试数据...", recordCount)
	for i := 1; i <= recordCount; i++ {
		record := map[string]any{
			"id":     i,
			"name":   fmt.Sprintf("User%d", i),
			"age":    i % 100,
			"active": i%2 == 0, // 一半记录为 active
		}
		_, err = table.Insert(&record)
		if err != nil {
			t.Fatalf("插入测试数据失败: %v", err)
		}
	}
	t.Logf("插入测试数据完成")

	// 测试批量删除性能
	t.Run("BatchDeletePerformance", func(t *testing.T) {
		// 获取迭代器并设置匹配条件
		iter, _ := table.Search(&map[string]any{"id": nil})
		if iter == nil {
			t.Fatalf("获取迭代器失败")
		}
		defer GlobalTableIterPool.Put(iter)

		// 设置匹配条件：删除 inactive 记录
		iter.SetMatch(match.NewFieldComparison("active", match.Equal, false))

		// 执行批量删除
		defer func() {
			start := time.Now()
			err = iter.Delete()
			if err != nil {
				t.Fatalf("批量删除失败: %v", err)
			}
			duration := time.Since(start)
			t.Logf("批量删除 %d 条记录耗时: %v", recordCount/2, duration)
		}()
	})

	// 测试批量更新性能
	t.Run("BatchUpdatePerformance", func(t *testing.T) {
		// 获取迭代器并设置匹配条件
		iter, _ := table.Search(&map[string]any{"id": nil})
		if iter == nil {
			t.Fatalf("获取迭代器失败")
		}
		defer GlobalTableIterPool.Put(iter)

		// 准备更新字段
		updateFields := map[string]any{
			"active": true,
		}

		// 执行批量更新
		defer func() {
			start := time.Now()
			err = iter.Update(&updateFields)
			if err != nil {
				t.Fatalf("批量更新失败: %v", err)
			}
			duration := time.Since(start)
			t.Logf("批量更新 %d 条记录耗时: %v", recordCount/2, duration)
		}()
	})
}
