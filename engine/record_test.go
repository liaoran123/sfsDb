package engine

import (
	"testing"
)

// 测试 Record.Select 方法
func TestRecordSelect(t *testing.T) {
	// 测试用例1: 空记录
	var r Record
	result := r.Select("name", "age")
	if result != nil {
		t.Errorf("Expected nil for nil Record, got %v", result)
	}

	// 测试用例2: 选择所有字段
	r = Record{"id": 1, "name": "张三", "age": 25}
	result = r.Select()
	if len(result) != 3 {
		t.Errorf("Expected 3 fields, got %d", len(result))
	}

	// 测试用例3: 选择指定字段
	result = r.Select("name", "age")
	if len(result) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(result))
	}
	if result["name"] != "张三" {
		t.Errorf("Expected name=张三, got %v", result["name"])
	}
	if result["age"] != 25 {
		t.Errorf("Expected age=25, got %v", result["age"])
	}
}

// 测试 Records 基本方法
func TestRecordsBasic(t *testing.T) {
	// 创建测试表
	table, err := TableNew("test_table")
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
	table.indexs.CreateIndex(pk)

	// 创建 Records 对象
	records := NewRecords(table)

	// 测试 Append 方法
	err = records.Append(Record{"id": 1, "name": "张三", "age": 25})
	if err != nil {
		t.Errorf("Failed to append record: %v", err)
	}
	err = records.Append(Record{"id": 2, "name": "李四", "age": 30})
	if err != nil {
		t.Errorf("Failed to append record: %v", err)
	}

	// 测试 Len 方法
	if records.Len() != 2 {
		t.Errorf("Expected length 2, got %d", records.Len())
	}

	// 测试 Get 方法
	record := records.Get(0)
	if record["name"] != "张三" {
		t.Errorf("Expected name=张三, got %v", record["name"])
	}

	// 测试 GetAll 方法
	allRecords := records.GetAll()
	if len(allRecords) != 2 {
		t.Errorf("Expected 2 records, got %d", len(allRecords))
	}

	// 测试 Contains 方法
	if !records.Contains(Record{"id": 1, "name": "张三", "age": 25}) {
		t.Errorf("Expected record to be contained")
	}
	if records.Contains(Record{"id": 3, "name": "王五", "age": 35}) {
		t.Errorf("Expected record not to be contained")
	}
}

// 测试 Records.Select 方法
func TestRecordsSelect(t *testing.T) {
	// 创建测试表
	table, err := TableNew("test_table")
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
	table.indexs.CreateIndex(pk)

	// 创建 Records 对象并添加记录
	records := NewRecords(table)
	err = records.Append(Record{"id": 1, "name": "张三", "age": 25})
	if err != nil {
		t.Errorf("Failed to append record: %v", err)
	}
	err = records.Append(Record{"id": 2, "name": "李四", "age": 30})
	if err != nil {
		t.Errorf("Failed to append record: %v", err)
	}

	// 测试用例1: 空字段列表
	result := records.Select()
	if len(result) != 0 {
		t.Errorf("Expected empty result for empty fields, got %d", len(result))
	}

	// 测试用例2: 选择指定字段
	result = records.Select("name", "age")
	if len(result) != 2 {
		t.Errorf("Expected 2 records, got %d", len(result))
	}
	if result[0]["name"] != "张三" {
		t.Errorf("Expected name=张三, got %v", result[0]["name"])
	}
	if result[1]["age"] != 30 {
		t.Errorf("Expected age=30, got %v", result[1]["age"])
	}
}

// 测试 Records 集合操作方法
func TestRecordsSetOperations(t *testing.T) {
	// 创建测试表
	table, err := TableNew("test_table")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}
	// 设置表字段
	fields := map[string]any{"id": 0, "name": ""}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}
	// 创建主键索引
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("id")
	table.indexs.CreateIndex(pk)

	// 创建第一个 Records 对象
	records1 := NewRecords(table)
	err = records1.Append(Record{"id": 1, "name": "张三"})
	if err != nil {
		t.Errorf("Failed to append record: %v", err)
	}
	err = records1.Append(Record{"id": 2, "name": "李四"})
	if err != nil {
		t.Errorf("Failed to append record: %v", err)
	}

	// 创建第二个 Records 对象
	records2 := NewRecords(table)
	err = records2.Append(Record{"id": 2, "name": "李四"})
	if err != nil {
		t.Errorf("Failed to append record: %v", err)
	}
	err = records2.Append(Record{"id": 3, "name": "王五"})
	if err != nil {
		t.Errorf("Failed to append record: %v", err)
	}

	// 测试 Intersect 方法
	intersect := records1.Intersect(records2)
	if intersect.Len() != 1 {
		t.Errorf("Expected 1 record in intersection, got %d", intersect.Len())
	}
	if intersect.Get(0)["name"] != "李四" {
		t.Errorf("Expected name=李四 in intersection, got %v", intersect.Get(0)["name"])
	}

	// 测试 Union 方法
	union := records1.Union(records2)
	if union.Len() != 3 {
		t.Errorf("Expected 3 records in union, got %d", union.Len())
	}

	// 测试 Difference 方法
	difference := records1.Difference(records2)
	if difference.Len() != 1 {
		t.Errorf("Expected 1 record in difference, got %d", difference.Len())
	}
	if difference.Get(0)["name"] != "张三" {
		t.Errorf("Expected name=张三 in difference, got %v", difference.Get(0)["name"])
	}
}

// 测试 Records.hasPrimaryKey 方法
func TestRecordsHasPrimaryKey(t *testing.T) {
	// 创建测试表
	table, err := TableNew("test_table")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}
	// 设置表字段
	fields := map[string]any{"id": 0, "name": ""}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}
	// 创建主键索引
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("id")
	table.indexs.CreateIndex(pk)

	// 创建 Records 对象
	records := NewRecords(table)

	// 测试用例1: 有主键的记录
	recordWithPK := Record{"id": 1, "name": "张三"}
	if !records.hasPrimaryKey(recordWithPK) {
		t.Errorf("Expected record with primary key to return true")
	}

	// 测试用例2: 无主键的记录
	recordWithoutPK := Record{"name": "张三"}
	if records.hasPrimaryKey(recordWithoutPK) {
		t.Errorf("Expected record without primary key to return false")
	}
}
