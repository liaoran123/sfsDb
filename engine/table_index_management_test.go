package engine

import (
	"testing"
)

// 测试索引管理功能
func TestIndexManagement(t *testing.T) {
	// 创建表
	table, err := TableNew("test_index_management")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}
	
	// 设置表字段
	fields := map[string]any{
		"id":    0,
		"name":  "",
		"age":   0,
		"email": "",
		"city":  "",
	}
	if err := table.SetFields(fields); err != nil {
		t.Fatalf("Failed to set table fields: %v", err)
	}
	
	// 创建各种索引
	err = table.CreatePrimaryKey("id")
	if err != nil {
		t.Fatalf("Failed to create primary key: %v", err)
	}
	
	err = table.CreateCompositeIndex("name_age_idx", "name", "age")
	if err != nil {
		t.Fatalf("Failed to create composite index: %v", err)
	}
	
	err = table.CreateSimpleIndex("email_idx", "email")
	if err != nil {
		t.Fatalf("Failed to create simple index: %v", err)
	}
	
	err = table.CreateCompositeIndex("city_email_idx", "city", "email")
	if err != nil {
		t.Fatalf("Failed to create city_email_idx: %v", err)
	}
	
	// 测试1: 获取所有索引
	allIndexes := table.GetAllIndexes()
	if len(allIndexes) != 4 {
		t.Errorf("Expected 4 indexes, got %d", len(allIndexes))
	}
	
	// 测试2: 根据名称获取索引
	emailIdx := table.GetIndexByName("email_idx")
	if emailIdx == nil {
		t.Fatal("Failed to get index by name 'email_idx'")
	}
	if emailIdx.Name() != "email_idx" {
		t.Errorf("Expected index name 'email_idx', got '%s'", emailIdx.Name())
	}
	
	// 测试3: 根据字段名获取索引
	emailIndexes := table.GetIndexesByField("email")
	if len(emailIndexes) != 2 {
		t.Errorf("Expected 2 indexes containing field 'email', got %d", len(emailIndexes))
	}
	
	cityIndexes := table.GetIndexesByField("city")
	if len(cityIndexes) != 1 {
		t.Errorf("Expected 1 index containing field 'city', got %d", len(cityIndexes))
	}
	
	// 测试4: 获取最佳匹配索引
	bestIdx := table.GetBestMatchIndex("id")
	if bestIdx == nil {
		t.Fatal("Failed to get best match index for 'id'")
	}
	if !bestIdx.MatchFields("id") {
		t.Error("Best match index should match field 'id'")
	}
	
	bestIdx = table.GetBestMatchIndex("name", "age")
	if bestIdx == nil {
		t.Fatal("Failed to get best match index for 'name', 'age'")
	}
	if bestIdx.Name() != "name_age_idx" {
		t.Errorf("Expected best match index 'name_age_idx', got '%s'", bestIdx.Name())
	}
	
	// 测试5: 获取匹配索引结果
	matchIdx, found := table.GetMatchIndexResult("email")
	if !found {
		t.Fatal("Expected to find match index for 'email'")
	}
	if matchIdx == nil {
		t.Fatal("Expected non-nil match index for 'email'")
	}
	
	// 测试6: 删除索引
	err = table.DropIndex("email_idx")
	if err != nil {
		t.Fatalf("Failed to drop index: %v", err)
	}
	
	// 验证索引已删除
	emailIdx = table.GetIndexByName("email_idx")
	if emailIdx != nil {
		t.Error("Expected index 'email_idx' to be deleted")
	}
	
	allIndexes = table.GetAllIndexes()
	if len(allIndexes) != 3 {
		t.Errorf("Expected 3 indexes after deletion, got %d", len(allIndexes))
	}
	
	// 测试7: 删除主键索引
	err = table.DropPrimaryKey()
	if err != nil {
		t.Fatalf("Failed to drop primary key: %v", err)
	}
	
	allIndexes = table.GetAllIndexes()
	if len(allIndexes) != 2 {
		t.Errorf("Expected 2 indexes after dropping primary key, got %d", len(allIndexes))
	}
	
	t.Log("All index management tests passed!")
}
