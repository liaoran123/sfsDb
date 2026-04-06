package engine

import (
	"testing"
)

// TestTableDelete 测试 Table.Delete 方法的功能
func TestTableDelete(t *testing.T) {
	// 创建测试表
	table, err := NewTable("test_delete")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
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
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键索引
	pkIndex, err := NewDefaultPrimaryKey("pk")
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}
	pkIndex.AddFields("id")
	err = table.CreateIndex(pkIndex)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// 插入测试数据
	testData := []map[string]any{
		{"id": 1, "name": "Alice", "age": 20, "active": true},
		{"id": 2, "name": "Bob", "age": 25, "active": true},
		{"id": 3, "name": "Charlie", "age": 30, "active": false},
	}

	for _, record := range testData {
		_, err = table.Insert(&record)
		if err != nil {
			t.Fatalf("Failed to insert record: %v", err)
		}
	}

	// 测试 1: 基本删除操作
	t.Run("BasicDelete", func(t *testing.T) {
		// 准备删除数据
		deleteFields := map[string]any{"id": 2}

		// 执行删除操作
		err = table.Delete(&deleteFields)
		if err != nil {
			t.Fatalf("Failed to delete record: %v", err)
		}

		// 验证删除是否成功
		// 尝试读取已删除的记录，应该返回错误
		readFields := map[string]any{"id": 2}
		record, _ := table.Read(&readFields)
		//fmt.Printf("record: %v\n", record)
		if len(record) != 0 {
			t.Errorf("Expected error when reading deleted record, but got %v", record)
		}

		// 验证其他记录是否仍然存在
		readFields = map[string]any{"id": 1}
		_, err = table.Read(&readFields)
		if err != nil {
			t.Errorf("Expected to read record with id=1, but got error: %v", err)
		}

		readFields = map[string]any{"id": 3}
		_, err = table.Read(&readFields)
		if err != nil {
			t.Errorf("Expected to read record with id=3, but got error: %v", err)
		}
	})

	// 测试 2: 使用批量操作删除
	t.Run("DeleteWithBatch", func(t *testing.T) {
		// 准备删除数据
		deleteFields := map[string]any{"id": 3}

		// 执行删除操作
		err = table.Delete(&deleteFields)
		if err != nil {
			t.Fatalf("Failed to delete record with batch: %v", err)
		}

		// 验证删除是否成功
		readFields := map[string]any{"id": 3}
		record, _ := table.Read(&readFields)
		//fmt.Printf("record: %v\n", record)
		if len(record) != 0 {
			t.Errorf("Expected error when reading deleted record, but got %v", record)
		}
	})

	// 测试 3: 删除不存在的记录
	t.Run("DeleteNonExistentRecord", func(t *testing.T) {
		// 准备删除数据
		deleteFields := map[string]any{"id": 999}

		// 执行删除操作，应该返回错误
		err = table.Delete(&deleteFields)
		if err == nil {
			t.Errorf("Expected error when deleting non-existent record, but got nil")
		}
	})

	// 测试 4: 删除所有记录
	t.Run("DeleteAll", func(t *testing.T) {
		// 执行删除所有操作
		err = table.DeleteAll()
		if err != nil {
			t.Fatalf("Failed to delete all records: %v", err)
		}

		// 验证所有记录是否都被删除
		readFields := map[string]any{"id": 1}
		record, _ := table.Read(&readFields)
		if len(record) != 0 {
			t.Errorf("Expected error when reading record after DeleteAll, but got %v", record)
		}
	})
}
