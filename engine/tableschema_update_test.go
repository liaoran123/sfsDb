package engine

import (
	"fmt"
	"testing"
)

// TestTableSchemaUpdate 测试更新后的 TableSchema 功能
func TestTableSchemaUpdate(t *testing.T) {
	// 创建表
	table, err := NewTable("test_schema_update")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置字段
	fields := map[string]any{
		"id":    0,
		"name":  "",
		"email": "",
		"age":   0,
	}
	if err := table.SetFields(fields); err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键索引
	primaryKey, err := NewDefaultPrimaryKey("id")
	if err != nil {
		t.Fatalf("Failed to create primary key: %v", err)
	}
	primaryKey.AddFields("id")
	if err := table.CreateIndex(primaryKey); err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// 创建普通索引
	normalIndex, err := NewDefaultNormalIndex("name_idx")
	if err != nil {
		t.Fatalf("Failed to create normal index: %v", err)
	}
	normalIndex.AddFields("name")
	if err := table.CreateIndex(normalIndex); err != nil {
		t.Fatalf("Failed to create normal index: %v", err)
	}

	// 测试1: 转换为 Schema
	schema := table.ToSchema()
	if schema == nil {
		t.Fatalf("Failed to convert table to schema")
	}

	// 验证 Schema 字段
	if schema.ID != table.id {
		t.Errorf("Schema ID mismatch: expected %d, got %d", table.id, schema.ID)
	}
	if schema.Name != table.name {
		t.Errorf("Schema Name mismatch: expected %s, got %s", table.name, schema.Name)
	}
	if len(schema.Fields) != len(table.fields) {
		t.Errorf("Schema Fields length mismatch: expected %d, got %d", len(table.fields), len(schema.Fields))
	}
	if len(schema.FieldsID) != len(table.fieldsid) {
		t.Errorf("Schema FieldsID length mismatch: expected %d, got %d", len(table.fieldsid), len(schema.FieldsID))
	}
	if len(schema.Indexes) != len(table.indexs.GetAllIndexes()) {
		t.Errorf("Schema Indexes length mismatch: expected %d, got %d", len(table.indexs.GetAllIndexes()), len(schema.Indexes))
	}
	if len(schema.PrimaryFields) != len(table.GetPrimaryFields()) {
		t.Errorf("Schema PrimaryFields length mismatch: expected %d, got %d", len(table.GetPrimaryFields()), len(schema.PrimaryFields))
	}

	// 测试2: 从 Schema 恢复表
	restoredTable, err := FromSchema(schema)
	if err != nil {
		t.Fatalf("Failed to restore table from schema: %v", err)
	}

	// 验证恢复的表字段
	if restoredTable.id != table.id {
		t.Errorf("Restored table ID mismatch: expected %d, got %d", table.id, restoredTable.id)
	}
	if restoredTable.name != table.name {
		t.Errorf("Restored table Name mismatch: expected %s, got %s", table.name, restoredTable.name)
	}
	if len(restoredTable.fields) != len(table.fields) {
		t.Errorf("Restored table Fields length mismatch: expected %d, got %d", len(table.fields), len(restoredTable.fields))
	}
	if len(restoredTable.fieldsid) != len(table.fieldsid) {
		t.Errorf("Restored table FieldsID length mismatch: expected %d, got %d", len(table.fieldsid), len(restoredTable.fieldsid))
	}
	if len(restoredTable.indexs.GetAllIndexes()) != len(table.indexs.GetAllIndexes()) {
		t.Errorf("Restored table Indexes length mismatch: expected %d, got %d", len(table.indexs.GetAllIndexes()), len(restoredTable.indexs.GetAllIndexes()))
	}

	// 测试3: 序列化和反序列化
	jsonStr, err := TableToJSON(table)
	if err != nil {
		t.Fatalf("Failed to convert table to JSON: %v", err)
	}

	parsedTable, err := TableFromJSON(jsonStr)
	if err != nil {
		t.Fatalf("Failed to parse table from JSON: %v", err)
	}

	// 验证解析的表字段
	if parsedTable.id != table.id {
		t.Errorf("Parsed table ID mismatch: expected %d, got %d", table.id, parsedTable.id)
	}
	if parsedTable.name != table.name {
		t.Errorf("Parsed table Name mismatch: expected %s, got %s", table.name, parsedTable.name)
	}
	if len(parsedTable.fields) != len(table.fields) {
		t.Errorf("Parsed table Fields length mismatch: expected %d, got %d", len(table.fields), len(parsedTable.fields))
	}

	fmt.Println("All TableSchema update tests passed!")
}
