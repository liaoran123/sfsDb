package engine

import (
	"testing"
)

// 测试新添加的索引创建API
func TestIndexCreationAPI(t *testing.T) {
	// 创建表
	table, err := TableNew("test_index_api")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{
		"id":    0,
		"name":  "",
		"age":   0,
		"email": "",
	}
	if err := table.SetFields(fields); err != nil {
		t.Fatalf("Failed to set table fields: %v", err)
	}

	// 测试1: 创建复合主键
	err = table.CreateCompositePrimaryKey("composite_pk", "id", "name")
	if err != nil {
		t.Fatalf("Failed to create composite primary key: %v", err)
	}

	// 验证主键已创建
	pk := table.GetPrimaryKey()
	if pk == nil {
		t.Fatal("Failed to get primary key after creation")
	}
	if len(pk.GetFields()) != 2 {
		t.Errorf("Expected 2 primary key fields, got %d", len(pk.GetFields()))
	}

	// 测试2: 创建普通复合索引
	err = table.CreateCompositeIndex("name_age_idx", "name", "age")
	if err != nil {
		t.Fatalf("Failed to create composite index: %v", err)
	}

	// 测试3: 使用简化API创建普通索引
	err = table.CreateSimpleIndex("email_idx", "email")
	if err != nil {
		t.Fatalf("Failed to create simple index: %v", err)
	}

	// 测试4: 验证索引数量
	indexes := table.indexs.GetAllIndexes()
	if len(indexes) != 3 {
		t.Errorf("Expected 3 indexes, got %d", len(indexes))
	}

	// 测试5: 验证索引字段
	for _, idx := range indexes {
		switch idx.Name() {
		case "composite_pk":
			if len(idx.GetFields()) != 2 {
				t.Errorf("Expected composite_pk to have 2 fields, got %d", len(idx.GetFields()))
			}
		case "name_age_idx":
			if len(idx.GetFields()) != 2 {
				t.Errorf("Expected name_age_idx to have 2 fields, got %d", len(idx.GetFields()))
			}
		case "email_idx":
			if len(idx.GetFields()) != 1 {
				t.Errorf("Expected email_idx to have 1 field, got %d", len(idx.GetFields()))
			}
		}
	}

	t.Log("All index creation API tests passed!")
}

// 测试单字段主键创建
func TestSingleFieldPrimaryKey(t *testing.T) {
	// 创建表
	table, err := TableNew("test_single_pk")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{
		"user_id": 0,
		"name":    "",
		"score":   0.0,
	}
	if err := table.SetFields(fields); err != nil {
		t.Fatalf("Failed to set table fields: %v", err)
	}

	// 使用简化API创建单字段主键
	err = table.CreatePrimaryKey("user_id")
	if err != nil {
		t.Fatalf("Failed to create single field primary key: %v", err)
	}

	// 验证主键
	pk := table.GetPrimaryKey()
	if pk == nil {
		t.Fatal("Failed to get primary key")
	}
	if len(pk.GetFields()) != 1 {
		t.Errorf("Expected 1 primary key field, got %d", len(pk.GetFields()))
	}
	if pk.GetFields()[0] != "user_id" {
		t.Errorf("Expected primary key field to be 'user_id', got '%s'", pk.GetFields()[0])
	}

	t.Log("Single field primary key test passed!")
}
