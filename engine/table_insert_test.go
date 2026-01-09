package engine

import (
	"testing"
	"time"
)

// TestTableInsertWithAllTypes 测试插入各种数据类型
func TestTableInsertWithAllTypes(t *testing.T) {
	table, err := TableNew("test_insert_all_types")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置包含各种数据类型的字段
	fields := map[string]any{
		"id":        0,          // 整数
		"name":      "",         // 字符串
		"age":       0,          // 整数
		"score":     0.0,        // 浮点数
		"active":    false,      // 布尔值
		"createdAt": time.Now(), // 时间
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

	// 测试1: 插入完整数据
	t.Run("CompleteData", func(t *testing.T) {
		testData := map[string]any{
			"id":        1,
			"name":      "Alice",
			"age":       20,
			"score":     85.5,
			"active":    true,
			"createdAt": time.Now(),
		}
		currentID, err := table.Insert(&testData)
		if err != nil {
			t.Fatalf("Failed to insert complete data: %v", err)
		}
		if currentID != 1 {
			t.Errorf("Expected currentID 1, got %d", currentID)
		}
	})

	// 测试2: 测试自动生成主键
	t.Run("AutoIncrementPK", func(t *testing.T) {
		testData := map[string]any{
			"name":      "Bob",
			"age":       25,
			"score":     90.0,
			"active":    false,
			"createdAt": time.Now(),
		}
		currentID, err := table.Insert(&testData)
		if err != nil {
			t.Fatalf("Failed to insert with auto increment: %v", err)
		}
		if currentID != 2 {
			t.Errorf("Expected currentID 2, got %d", currentID)
		}
	})

	// 测试3: 测试nil值的默认处理
	t.Run("NilValues", func(t *testing.T) {
		testData := map[string]any{
			"id":        3,
			"name":      nil, // 应该使用空字符串
			"age":       nil, // 应该使用0
			"score":     nil, // 应该使用0.0
			"active":    nil, // 应该使用false
			"createdAt": nil, // 应该使用当前时间
		}
		currentID, err := table.Insert(&testData)
		if err != nil {
			t.Fatalf("Failed to insert with nil values: %v", err)
		}
		if currentID != 3 {
			t.Errorf("Expected currentID 3, got %d", currentID)
		}
	})

	// 验证插入的数据
	iter := table.Search(&map[string]any{"id": nil})
	defer iter.Release()
	records := iter.GetRecords(true)
	if len(records) != 3 {
		t.Fatalf("Expected 3 records, got %d", len(records))
	}

	// 验证第一条记录（完整数据）
	record1 := records[0]
	if record1["id"] != 1 || record1["name"] != "Alice" || record1["age"] != 20 || record1["score"] != 85.5 || record1["active"] != true {
		t.Errorf("Record 1 data mismatch: %v", record1)
	}

	// 验证第二条记录（自动生成主键）
	record2 := records[1]
	if record2["id"] != 2 || record2["name"] != "Bob" || record2["age"] != 25 || record2["score"] != 90.0 || record2["active"] != false {
		t.Errorf("Record 2 data mismatch: %v", record2)
	}

	// 验证第三条记录（nil值默认处理）
	record3 := records[2]
	if record3["id"] != 3 || record3["name"] != "" || record3["age"] != 0 || record3["score"] != 0.0 || record3["active"] != false {
		t.Errorf("Record 3 data mismatch: %v", record3)
	}
	if record3["createdAt"] == nil {
		t.Error("Record 3 createdAt should not be nil")
	}
}

// TestTableFieldsToBytesWithNil 测试FieldsToBytes函数的nil值处理
func TestTableFieldsToBytesWithNil(t *testing.T) {
	table, err := TableNew("test_field_to_bytes")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置字段
	fields := map[string]any{
		"id":        0,
		"name":      "",
		"age":       0,
		"score":     0.0,
		"active":    false,
		"createdAt": time.Now(),
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

	// 测试nil值处理
	testData := map[string]any{
		"id":        1,
		"name":      nil,
		"age":       nil,
		"score":     nil,
		"active":    nil,
		"createdAt": nil,
	}

	// 调用FieldsToBytes
	result := table.FieldsToBytes(&testData)
	if result == nil {
		t.Fatal("FieldsToBytes returned nil")
	}

	// 验证结果
	if len(*result) != 6 {
		t.Errorf("Expected 6 fields, got %d", len(*result))
	}

	// 验证各个字段的处理
	if _, ok := (*result)["id"]; !ok {
		t.Error("id field not found in result")
	}
	if _, ok := (*result)["name"]; !ok {
		t.Error("name field not found in result")
	}
	if _, ok := (*result)["age"]; !ok {
		t.Error("age field not found in result")
	}
	if _, ok := (*result)["score"]; !ok {
		t.Error("score field not found in result")
	}
	if _, ok := (*result)["active"]; !ok {
		t.Error("active field not found in result")
	}
	if _, ok := (*result)["createdAt"]; !ok {
		t.Error("createdAt field not found in result")
	}
}

// TestTableUpdateWithAllTypes 测试更新各种数据类型
func TestTableUpdateWithAllTypes(t *testing.T) {
	table, err := TableNew("test_update_all_types")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置包含各种数据类型的字段
	fields := map[string]any{
		"id":        0,          // 整数
		"name":      "",         // 字符串
		"age":       0,          // 整数
		"score":     0.0,        // 浮点数
		"active":    false,      // 布尔值
		"createdAt": time.Now(), // 时间
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

	// 插入测试数据
	testData := map[string]any{
		"id":        1,
		"name":      "Alice",
		"age":       20,
		"score":     85.5,
		"active":    true,
		"createdAt": time.Now(),
	}
	_, err = table.Insert(&testData)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// 测试1: 更新所有类型的字段
	t.Run("UpdateAllTypes", func(t *testing.T) {
		updateData := map[string]any{
			"id":        1,               // 主键，用于定位记录
			"name":      "Alice Updated", // 字符串
			"age":       21,              // 整数
			"score":     90.5,            // 浮点数
			"active":    false,           // 布尔值
			"createdAt": time.Now(),      // 时间
		}
		err := table.Update(&updateData)
		if err != nil {
			t.Fatalf("Failed to update all types: %v", err)
		}
	})

	// 测试2: 只更新部分字段
	t.Run("UpdatePartialFields", func(t *testing.T) {
		updateData := map[string]any{
			"id":     1,                         // 主键，用于定位记录
			"name":   "Alice Partially Updated", // 只更新字符串字段
			"active": true,                      // 只更新布尔字段
		}
		err := table.Update(&updateData)
		if err != nil {
			t.Fatalf("Failed to update partial fields: %v", err)
		}
	})

	// 测试3: 更新字段为nil值
	t.Run("UpdateWithNilValues", func(t *testing.T) {
		updateData := map[string]any{
			"id":        1,   // 主键，用于定位记录
			"name":      nil, // 字符串设为nil
			"age":       nil, // 整数设为nil
			"score":     nil, // 浮点数设为nil
			"active":    nil, // 布尔值设为nil
			"createdAt": nil, // 时间设为nil
		}
		err := table.Update(&updateData)
		if err != nil {
			t.Fatalf("Failed to update with nil values: %v", err)
		}
	})

	// 验证更新后的数据
	iter := table.Search(&map[string]any{"id": 1})
	defer iter.Release()
	records := iter.GetRecords(true)
	if len(records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(records))
	}

	record := records[0]
	t.Logf("Updated record: %v", record)

	// 验证字段类型是否正确
	if _, ok := record["id"].(int); !ok {
		t.Error("id should be int")
	}
	if _, ok := record["name"].(string); !ok {
		t.Error("name should be string")
	}
	if _, ok := record["age"].(int); !ok {
		t.Error("age should be int")
	}
	if _, ok := record["score"].(float64); !ok {
		t.Error("score should be float64")
	}
	if _, ok := record["active"].(bool); !ok {
		t.Error("active should be bool")
	}
	if _, ok := record["createdAt"].(time.Time); !ok {
		t.Error("createdAt should be time.Time")
	}
}
