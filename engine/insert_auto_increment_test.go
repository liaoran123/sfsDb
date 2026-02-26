package engine

import (
	"testing"
)

// TestInsertImpl_AutoIncrement 测试AutoIncrement方法的各种情况
func TestInsertImpl_AutoIncrement(t *testing.T) {
	// 测试情况1: 主键不是"id"
	t.Run("NonIdPrimaryKey", func(t *testing.T) {
		// 创建一个测试表
		table, err := TableNew("test_table2")
		if err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}
		// 设置主键为非"id"字段
		table.DropPrimaryKey()
		err = table.CreatePrimaryKey("custom_id")
		if err != nil {
			t.Fatalf("Failed to create primary key: %v", err)
		}
		
		fields := map[string]any{"custom_id": 1, "name": "test"}
		insertImpl := NewInsertImpl(table, &fields)
		currentID, err := insertImpl.AutoIncrement()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if currentID != 1 {
			t.Errorf("Expected currentID to be 1, got %d", currentID)
		}
	})
	
	// 测试情况2: 未提供主键值
	t.Run("NoPrimaryKeyValue", func(t *testing.T) {
		// 创建一个测试表
		table, err := TableNew("test_table3")
		if err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}
		// 确保主键是"id"
		table.DropPrimaryKey()
		err = table.CreatePrimaryKey("id")
		if err != nil {
			t.Fatalf("Failed to create primary key: %v", err)
		}
		
		fields := map[string]any{"name": "test"}
		insertImpl := NewInsertImpl(table, &fields)
		currentID, err := insertImpl.AutoIncrement()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if currentID != 1 {
			t.Errorf("Expected currentID to be 1, got %d", currentID)
		}
		if (*insertImpl.fields)["id"] != 1 {
			t.Errorf("Expected fields['id'] to be 1, got %v", (*insertImpl.fields)["id"])
		}
	})
	
	// 测试情况3: 提供了nil主键值
	t.Run("NilPrimaryKeyValue", func(t *testing.T) {
		// 创建一个测试表
		table, err := TableNew("test_table4")
		if err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}
		// 确保主键是"id"
		table.DropPrimaryKey()
		err = table.CreatePrimaryKey("id")
		if err != nil {
			t.Fatalf("Failed to create primary key: %v", err)
		}
		
		fields := map[string]any{"id": nil, "name": "test"}
		insertImpl := NewInsertImpl(table, &fields)
		currentID, err := insertImpl.AutoIncrement()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if currentID != 1 {
			t.Errorf("Expected currentID to be 1, got %d", currentID)
		}
		if (*insertImpl.fields)["id"] != 1 {
			t.Errorf("Expected fields['id'] to be 1, got %v", (*insertImpl.fields)["id"])
		}
	})
	
	// 测试情况4: 提供了非nil主键值
	t.Run("NonNilPrimaryKeyValue", func(t *testing.T) {
		// 创建一个测试表
		table, err := TableNew("test_table5")
		if err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}
		// 确保主键是"id"
		table.DropPrimaryKey()
		err = table.CreatePrimaryKey("id")
		if err != nil {
			t.Fatalf("Failed to create primary key: %v", err)
		}
		
		fields := map[string]any{"id": 10, "name": "test"}
		insertImpl := NewInsertImpl(table, &fields)
		currentID, err := insertImpl.AutoIncrement()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if currentID != 10 {
			t.Errorf("Expected currentID to be 10, got %d", currentID)
		}
		if (*insertImpl.fields)["id"] != 10 {
			t.Errorf("Expected fields['id'] to be 10, got %v", (*insertImpl.fields)["id"])
		}
	})
}
