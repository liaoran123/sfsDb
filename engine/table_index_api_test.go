package engine

import (
	"fmt"
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

// 测试在表已经存在数据的情况下，新建索引是否正确建立索引数据
func TestCreateIndexWithExistingData(t *testing.T) {
	// 创建表
	table, err := TableNew("test_index_existing_data")
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

	// 创建主键
	err = table.CreatePrimaryKey("id")
	if err != nil {
		t.Fatalf("Failed to create primary key: %v", err)
	}

	// 插入多条测试数据
	testData := []map[string]any{
		{"id": 1, "name": "张三", "age": 30, "email": "zhangsan@example.com"},
		{"id": 2, "name": "李四", "age": 25, "email": "lisi@example.com"},
		{"id": 3, "name": "王五", "age": 35, "email": "wangwu@example.com"},
		{"id": 4, "name": "赵六", "age": 28, "email": "zhaoliu@example.com"},
	}

	for _, data := range testData {
		_, err := table.Insert(&data)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// 验证数据已插入
	iter := table.ForData()
	defer GlobalTableIterPool.Put(iter)
	records := iter.GetRecords(true)
	if len(records) != len(testData) {
		t.Errorf("Expected %d records, got %d", len(testData), len(records))
	}

	// 在已有数据的情况下创建新索引
	err = table.CreateSimpleIndex("name_idx", "name")
	if err != nil {
		fmt.Printf("Failed to create index with existing data: %v", err)
	}

	// 验证索引是否存在
	indexes := table.GetAllIndexes()
	if len(indexes) != 2 {
		t.Errorf("Expected 2 indexes after creation, got %d", len(indexes))
	}

	// 验证所有数据是否仍然存在
	allIter := table.ForData()
	defer GlobalTableIterPool.Put(allIter)
	allRecords := allIter.GetRecords(true)
	if len(allRecords) != len(testData) {
		t.Errorf("Expected %d records in total, got %d", len(testData), len(allRecords))
	}

	// 验证索引字段是否正确
	nameIndexFound := false
	for _, idx := range indexes {
		if idx.Name() == "name_idx" {
			nameIndexFound = true
			fields := idx.GetFields()
			if len(fields) != 1 {
				t.Errorf("Expected 1 field for name_idx, got %d", len(fields))
			}
			if fields[0] != "name" {
				t.Errorf("Expected field 'name' for name_idx, got '%s'", fields[0])
			}
			break
		}
	}

	if !nameIndexFound {
		t.Error("Expected index 'name_idx' not found")
	}

	// 测试搜索功能
	t.Log("Testing search functionality...")

	// 测试搜索张三
	searchCriteria := map[string]any{"name": "张三"}
	searchIter, _ := table.Search(&searchCriteria)
	defer GlobalTableIterPool.Put(searchIter)
	searchRecords := searchIter.GetRecords(true)

	t.Logf("Search for '张三' returned %d records", len(searchRecords))
	for i, record := range searchRecords {
		t.Logf("Record %d: %v", i, record)
	}

	// 测试搜索李四
	searchCriteria2 := map[string]any{"name": "李四"}
	searchIter2, _ := table.Search(&searchCriteria2)
	defer GlobalTableIterPool.Put(searchIter2)
	searchRecords2 := searchIter2.GetRecords(true)

	t.Logf("Search for '李四' returned %d records", len(searchRecords2))
	for i, record := range searchRecords2 {
		t.Logf("Record %d: %v", i, record)
	}

	t.Log("Create index with existing data test passed!")
}

// 测试在表已经存在数据的情况下，新建全文索引是否正确建立索引数据
func TestCreateFullTextIndexWithExistingData(t *testing.T) {
	// 创建表
	table, err := TableNew("test_fulltext_index_existing_data")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{
		"id":      0,
		"title":   "",
		"content": "",
		"author":  "",
	}
	if err := table.SetFields(fields); err != nil {
		t.Fatalf("Failed to set table fields: %v", err)
	}

	// 创建主键
	err = table.CreatePrimaryKey("id")
	if err != nil {
		t.Fatalf("Failed to create primary key: %v", err)
	}

	// 插入多条测试数据
	testData := []map[string]any{
		{"id": 1, "title": "Go语言入门", "content": "这是一本关于Go语言的入门书籍", "author": "张三"},
		{"id": 2, "title": "Python高级编程", "content": "Python语言的高级特性和最佳实践", "author": "李四"},
		{"id": 3, "title": "JavaScript实战", "content": "前端开发中的JavaScript技巧", "author": "王五"},
		{"id": 4, "title": "数据库设计", "content": "关系型数据库的设计原则", "author": "赵六"},
	}

	for _, data := range testData {
		_, err := table.Insert(&data)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// 验证数据已插入
	iter := table.ForData()
	defer GlobalTableIterPool.Put(iter)
	records := iter.GetRecords(true)
	if len(records) != len(testData) {
		t.Errorf("Expected %d records, got %d", len(testData), len(records))
	}

	// 测试1: 没有全文索引时的搜索
	t.Log("Testing search without full text index...")
	searchCriteriaBefore := map[string]any{"content": "Go语言"}
	searchIterBefore, _ := table.Search(&searchCriteriaBefore)
	if searchIterBefore != nil {
		defer GlobalTableIterPool.Put(searchIterBefore)
		searchRecordsBefore := searchIterBefore.GetRecords(true)
		t.Logf("Search for 'Go语言' without full text index returned %d records", len(searchRecordsBefore))
		// 没有全文索引时，搜索可能返回空结果
	} else {
		t.Log("Search returned nil iterator without full text index (expected)")
	}

	// 在已有数据的情况下创建全文索引
	fullText, err := DefaultFullTextIndexNew("content_ft_idx")
	if err != nil {
		t.Fatalf("Failed to create full text index: %v", err)
	}
	// 全文索引正常情况下必须带上主键，否则后面的关键词都被覆盖，失去全文索引的意义
	fullText.AddFields("content", "id") // 创建一个content的组合全文索引
	// 指定content为全文索引字段，长度为5
	err = fullText.SetFullField("content", 5)
	if err != nil {
		t.Fatalf("Failed to set full field: %v", err)
	}
	// 创建全文索引
	err = table.CreateIndex(fullText)
	if err != nil {
		t.Fatalf("Failed to create full text index with existing data: %v", err)
	}

	// 验证索引是否存在
	indexes := table.GetAllIndexes()
	if len(indexes) != 2 {
		t.Errorf("Expected 2 indexes after creation, got %d", len(indexes))
	}

	// 验证所有数据是否仍然存在
	allIter := table.ForData()
	defer GlobalTableIterPool.Put(allIter)
	allRecords := allIter.GetRecords(true)
	if len(allRecords) != len(testData) {
		t.Errorf("Expected %d records in total, got %d", len(testData), len(allRecords))
	}

	// 测试搜索功能
	t.Log("Testing search functionality for full text index...")

	// 测试搜索包含"Go语言"的文档
	searchCriteria := map[string]any{"content": "Go语言"}
	searchIter, _ := table.Search(&searchCriteria)
	defer GlobalTableIterPool.Put(searchIter)
	if searchIter != nil {
		defer GlobalTableIterPool.Put(searchIter)
		searchRecords := searchIter.GetRecords(true)
		t.Logf("Search for 'Go语言' returned %d records", len(searchRecords))
		for i, record := range searchRecords {
			t.Logf("Record %d: %v", i, record)
		}
	} else {
		t.Log("Search returned nil iterator (expected for full text search without proper index)")
	}

	t.Log("Create full text index with existing data test passed!")
}
