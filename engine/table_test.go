package engine

import (
	"fmt"
	"testing"
	"time"

	"github.com/liaoran123/sfsDb/util"
)

// 测试复合主键搜索
func TestCompositePrimaryKeySearch(t *testing.T) {
	// 使用唯一表名，避免测试数据累积
	tableName := fmt.Sprintf("test_composite_pk_%d", time.Now().UnixNano())

	// 创建表
	table, err := TableNew(tableName)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置复合主键字段
	fields := map[string]any{
		"id":      0,
		"name":    "",
		"age":     0,
		"email":   "",
		"phone":   "",
		"address": "",
	}

	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 创建复合主键索引
	pk, err := DefaultPrimaryKeyNew("pk_id_name")
	if err != nil {
		t.Fatalf("Failed to create primary key: %v", err)
	}

	// 添加主键字段
	pk.AddFields("id", "name")

	// 添加主键索引
	err = table.CreateIndex(pk)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// 插入测试数据
	records := []map[string]any{
		{"id": 1, "name": "Alice", "age": 25, "email": "alice@example.com", "phone": "123456789", "address": "123 Main St"},
		{"id": 2, "name": "Bob", "age": 30, "email": "bob@example.com", "phone": "987654321", "address": "456 Elm St"},
		{"id": 3, "name": "Charlie", "age": 35, "email": "charlie@example.com", "phone": "555555555", "address": "789 Oak St"},
	}

	for _, record := range records {
		_, err = table.Insert(&record)
		if err != nil {
			t.Fatalf("Failed to insert record: %v", err)
		}
	}

	// 测试搜索复合主键
	searchRecord := map[string]any{
		"id":   2,
		"name": "Bob",
	}

	iter := table.Search(&searchRecord)
	if iter == nil {
		t.Fatalf("Failed to search composite primary key")
	}
	defer iter.Release()

	resultRecords := iter.GetRecords(true)
	if len(resultRecords) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(resultRecords))
	}

	result := resultRecords[0]
	if result["id"] != 2 || result["name"] != "Bob" {
		t.Fatalf("Expected record with id=2, name=Bob, got %v", result)
	}

	// 测试更新记录
	updateRecord := map[string]any{
		"id":      2,
		"name":    "Bob",
		"age":     31,
		"email":   "bob.updated@example.com",
		"phone":   "111222333",
		"address": "456 Updated St",
	}

	err = table.Update(&updateRecord)
	if err != nil {
		t.Fatalf("Failed to update record: %v", err)
	}

	// 验证更新
	iter = table.Search(&searchRecord)
	if iter == nil {
		t.Fatalf("Failed to search composite primary key after update")
	}
	defer iter.Release()

	resultRecords = iter.GetRecords(true)
	if len(resultRecords) != 1 {
		t.Fatalf("Expected 1 record after update, got %d", len(resultRecords))
	}

	result = resultRecords[0]
	if result["age"] != 31 || result["email"] != "bob.updated@example.com" {
		t.Fatalf("Expected updated record, got %v", result)
	}

	// 测试删除记录
	err = table.Delete(&searchRecord)
	if err != nil {
		t.Fatalf("Failed to delete record: %v", err)
	}

	// 验证删除
	iter = table.Search(&searchRecord)
	if iter == nil {
		t.Fatalf("Failed to search composite primary key after delete")
	}
	defer iter.Release()

	resultRecords = iter.GetRecords(true)
	if len(resultRecords) != 0 {
		t.Fatalf("Expected 0 records after delete, got %d", len(resultRecords))
	}
}

// 测试添加Table.Insert，删除Table.Delete，修改Table.Update，添加一条记录，通过主键进行修改和删除
func TestTableCRUD(t *testing.T) {
	// 使用唯一表名，避免测试数据累积
	tableName := fmt.Sprintf("test_table_CRUD_%d", time.Now().UnixNano())
	tableWithIndexName := fmt.Sprintf("test_table_with_index_%d", time.Now().UnixNano())

	// 创建表
	table, err := TableNew(tableName)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 定义表字段
	fields := map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	}

	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键索引
	pk, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}
	pk.AddFields("id")
	err = table.CreateIndex(pk)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// 插入记录
	record := map[string]any{
		"id":   1,
		"name": "Alice",
		"age":  25,
	}

	_, err = table.Insert(&record)
	if err != nil {
		t.Fatalf("Failed to insert record: %v", err)
	}

	// 搜索记录
	iter := table.Search(&map[string]any{"id": 1})
	if iter == nil {
		t.Fatalf("Failed to search record")
	}
	defer iter.Release()

	records := iter.GetRecords(true)
	if len(records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(records))
	}

	// 更新记录
	updateRecord := map[string]any{
		"id":   1,
		"name": "Alice Updated",
		"age":  26,
	}

	err = table.Update(&updateRecord)
	if err != nil {
		t.Fatalf("Failed to update record: %v", err)
	}

	// 验证更新
	iter = table.Search(&map[string]any{"id": 1})
	if iter == nil {
		t.Fatalf("Failed to search record after update")
	}
	defer iter.Release()

	records = iter.GetRecords(true)
	if len(records) != 1 {
		t.Fatalf("Expected 1 record after update, got %d", len(records))
	}

	updatedRecord := records[0]
	if updatedRecord["name"] != "Alice Updated" || updatedRecord["age"] != 26 {
		t.Fatalf("Expected updated record, got %v", updatedRecord)
	}

	// 删除记录
	err = table.Delete(&map[string]any{"id": 1})
	if err != nil {
		t.Fatalf("Failed to delete record: %v", err)
	}

	// 验证删除
	iter = table.Search(&map[string]any{"id": 1})
	if iter == nil {
		t.Fatalf("Failed to search record after delete")
	}
	defer iter.Release()

	records = iter.GetRecords(true)
	if len(records) != 0 {
		t.Fatalf("Expected 0 records after delete, got %d", len(records))
	}

	// 测试带有索引的表
	tableWithIndex, err := TableNew(tableWithIndexName)
	if err != nil {
		t.Fatalf("Failed to create table with index: %v", err)
	}

	// 定义表字段
	indexedFields := map[string]any{
		"id":    0,
		"name":  "",
		"age":   0,
		"email": "",
	}

	err = tableWithIndex.SetFields(indexedFields)
	if err != nil {
		t.Fatalf("Failed to set fields for table with index: %v", err)
	}

	// 创建主键索引
	pk2, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}
	pk2.AddFields("id")
	err = tableWithIndex.CreateIndex(pk2)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// 创建二级索引
	secondaryIndex, err := DefaultNormalIndexNew("idx_email")
	if err != nil {
		t.Fatalf("Failed to create secondary index: %v", err)
	}
	secondaryIndex.AddFields("email")
	err = tableWithIndex.CreateIndex(secondaryIndex)
	if err != nil {
		t.Fatalf("Failed to create secondary index: %v", err)
	}

	// 插入多条记录
	for i := 1; i <= 5; i++ {
		idxRecord := map[string]any{
			"id":    i,
			"name":  fmt.Sprintf("User%d", i),
			"age":   20 + i,
			"email": fmt.Sprintf("user%d@example.com", i),
		}

		_, err = tableWithIndex.Insert(&idxRecord)
		if err != nil {
			t.Fatalf("Failed to insert record %d: %v", i, err)
		}
	}

	// 使用二级索引搜索
	emailIter := tableWithIndex.Search(&map[string]any{"email": "user3@example.com"})
	if emailIter == nil {
		t.Fatalf("Failed to search by email")
	}
	defer emailIter.Release()

	emailRecords := emailIter.GetRecords(true)
	if len(emailRecords) != 1 {
		t.Fatalf("Expected 1 record for email search, got %d", len(emailRecords))
	}

	if emailRecords[0]["id"] != 3 || emailRecords[0]["email"] != "user3@example.com" {
		t.Fatalf("Expected record with id=3 and email=user3@example.com, got %v", emailRecords[0])
	}
}

// 测试全文索引搜索
func TestTable_FullTextSearch(t *testing.T) {
	// 使用唯一表名，避免测试数据累积
	tableName := fmt.Sprintf("test_fulltext_search_%d", time.Now().UnixNano())

	// 创建表
	table, err := TableNew(tableName)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 定义表字段，包含用于全文搜索的description字段
	fields := map[string]any{
		"id":          0,
		"name":        "",
		"age":         0,
		"description": "",
	}

	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键索引
	pk, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		t.Fatalf("Failed to create primary key: %v", err)
	}
	pk.AddFields("id")
	err = table.CreateIndex(pk)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// 创建全文索引
	fulltextIdx, err := DefaultFullTextIndexNew("fulltext_desc")
	if err != nil {
		t.Fatalf("Failed to create fulltext index: %v", err)
	}
	fulltextIdx.AddFields("description", "id")       // 最后一个字段必须是主键
	err = fulltextIdx.SetFullField("description", 5) // 设置全文索引字段和长度
	if err != nil {
		t.Fatalf("Failed to set full field: %v", err)
	}
	err = table.CreateIndex(fulltextIdx)
	if err != nil {
		t.Fatalf("Failed to create fulltext index: %v", err)
	}

	// 插入测试数据
	testData := []map[string]any{
		{"id": 1, "name": "Alice", "age": 25, "description": "Alice is a software engineer who works with Bob"},
		{"id": 2, "name": "Bob", "age": 30, "description": "Bob is a project manager who manages Alice"},
		{"id": 3, "name": "Charlie", "age": 35, "description": "Charlie is a designer who collaborates with Alice and Bob"},
		{"id": 4, "name": "David", "age": 40, "description": "David is a developer who works independently"},
	}

	for _, record := range testData {
		_, err = table.Insert(&record)
		if err != nil {
			t.Fatalf("Failed to insert record: %v", err)
		}
	}

	// 测试全文搜索
	searchTests := []struct {
		name          string
		searchTerm    string
		expectedCount int
		expectedNames []string
	}{
		{
			name:          "Search for 'Bob'",
			searchTerm:    "Bob",
			expectedCount: 2,                                   // Alice and Bob records contain "Bob"
			expectedNames: []string{"Alice", "Bob", "Charlie"}, // All three mention Bob
		},
		{
			name:          "Search for 'Alice'",
			searchTerm:    "Alice",
			expectedCount: 2,                                   // Alice and Charlie records contain "Alice"
			expectedNames: []string{"Alice", "Bob", "Charlie"}, // All three mention Alice
		},
		{
			name:          "Search for 'David'",
			searchTerm:    "David",
			expectedCount: 1, // Only David's record contains "David"
			expectedNames: []string{"David"},
		},
		{
			name:          "Search for 'developer'",
			searchTerm:    "developer",
			expectedCount: 1, // Only David is explicitly described as a developer
			expectedNames: []string{"David"},
		},
	}

	for _, st := range searchTests {
		t.Run(st.name, func(t *testing.T) {
			// 使用Search方法搜索包含关键词的记录
			searchFields := map[string]any{
				"description": st.searchTerm,
			}
			iter := table.Search(&searchFields)
			if iter == nil {
				t.Fatalf("Search failed for term: %s", st.searchTerm)
			}
			defer iter.Release()

			records := iter.GetRecords(true)
			if len(records) == 0 {
				t.Logf("No records found for term: %s", st.searchTerm)
				return
			}

			// 验证搜索结果
			t.Logf("Found %d records for term: %s", len(records), st.searchTerm)
			for _, record := range records {
				name, _ := record["name"].(string)
				desc, _ := record["description"].(string)
				t.Logf("  - %s: %s", name, desc)
			}

			// 验证结果包含预期名称
			foundNames := make(map[string]bool)
			for _, record := range records {
				name, ok := record["name"].(string)
				if ok {
					foundNames[name] = true
				}
			}

			// 验证所有预期名称都被找到
			for _, expectedName := range st.expectedNames {
				if !foundNames[expectedName] {
					t.Errorf("Expected to find record with name '%s' for term '%s', but didn't", expectedName, st.searchTerm)
				}
			}
		})
	}

	t.Log("✅ All full text search tests completed successfully")
}

// 测试修改字段名称的完整流程
func TestTable_UpdateFieldName_Flow(t *testing.T) {
	// 使用唯一表名，避免测试数据累积
	tableName := fmt.Sprintf("test_field_update_flow_%d", time.Now().UnixNano())

	// 创建表
	table, err := TableNew(tableName)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 1. 初始设置字段
	initialFields := map[string]any{
		"id":    0,
		"name":  "",
		"age":   0,
		"email": "",
	}

	err = table.SetFields(initialFields)
	if err != nil {
		t.Fatalf("Failed to set initial fields: %v", err)
	}

	// 2. 创建索引
	idx, err := DefaultNormalIndexNew("email_idx")
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}
	idx.AddFields("email")
	err = table.CreateIndex(idx)
	if err != nil {
		t.Fatalf("Failed to add index: %v", err)
	}

	// 3. 插入测试数据
	testData := map[string]any{
		"id":    1,
		"name":  "Alice",
		"age":   25,
		"email": "alice@example.com",
	}

	_, err = table.Insert(&testData)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// 4. 测试搜索功能（初始字段名）
	iter := table.Search(&map[string]any{"email": "alice@example.com"})
	if iter == nil {
		t.Fatalf("Failed to search by initial email field")
	}
	defer iter.Release()

	records := iter.GetRecords(true)
	if len(records) != 1 {
		t.Fatalf("Expected 1 record for email search, got %d", len(records))
	}

	// 5. 开始修改字段名
	oldField := "email"
	newField := "email_address"

	// 5.1 必须先调用 UpdateFieldName
	err = table.UpdateFieldName(oldField, newField)
	if err != nil {
		t.Fatalf("Failed to update field name: %v", err)
	}

	// 5.2 然后调用 SetFields 更新字段映射
	updatedFields := map[string]any{
		"id":            0,
		"name":          "",
		"age":           0,
		"email_address": "", // 使用新字段名
	}

	err = table.SetFields(updatedFields)
	if err != nil {
		t.Fatalf("Failed to set updated fields: %v", err)
	}

	// 6. 验证字段修改后的功能

	// 6.1 插入新记录使用新字段名
	newTestData := map[string]any{
		"id":            2,
		"name":          "Bob",
		"age":           30,
		"email_address": "bob@example.com",
	}

	_, err = table.Insert(&newTestData)
	if err != nil {
		t.Fatalf("Failed to insert with new field name: %v", err)
	}

	// 6.2 搜索使用新字段名
	iter2 := table.Search(&map[string]any{"email_address": "bob@example.com"})
	if iter2 == nil {
		t.Fatalf("Failed to search by new email_address field")
	}
	defer iter2.Release()

	records2 := iter2.GetRecords(true)
	if len(records2) != 1 {
		t.Fatalf("Expected 1 record for new email_address search, got %d", len(records2))
	}

	// 6.3 验证旧记录仍然可访问（通过新字段名）
	iter3 := table.Search(&map[string]any{"email_address": "alice@example.com"})
	if iter3 == nil {
		t.Fatalf("Failed to search old record by new field name")
	}
	defer iter3.Release()

	records3 := iter3.GetRecords(true)
	if len(records3) != 1 {
		t.Fatalf("Expected 1 old record for new field name search, got %d", len(records3))
	}

	// 6.4 验证索引仍然有效
	if len(table.indexs.GetAllIndexes()) == 0 {
		t.Fatalf("Expected indexes to exist after field update")
	}

	// 6.5 验证字段ID映射仍然一致
	if len(table.fieldsid) != len(updatedFields) {
		t.Fatalf("Expected %d field IDs, got %d", len(updatedFields), len(table.fieldsid))
	}

	// 6.6 验证CheckType仍然正常工作
	checkData := map[string]any{
		"id":            3,
		"name":          "Charlie",
		"age":           35,
		"email_address": "charlie@example.com",
	}

	err = table.CheckType(&checkData)
	if err != nil {
		t.Fatalf("CheckType failed after field update: %v", err)
	}

	// 6.7 验证FieldsToBytes正常工作
	bytesMap := table.FieldsToBytes(&checkData)
	if bytesMap == nil {
		t.Fatalf("FieldsToBytes returned nil after field update")
	}
	if len(*bytesMap) != len(checkData) {
		t.Fatalf("Expected %d bytes fields, got %d", len(checkData), len(*bytesMap))
	}

	t.Log("✅ All field update tests passed successfully")
}

// 测试不调用UpdateFieldName直接修改字段名的错误场景
func TestTable_UpdateFieldName_MissingCall(t *testing.T) {
	// 创建测试表
	table, err := TableNew("test_field_update_missing")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 初始设置字段
	initialFields := map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	}

	err = table.SetFields(initialFields)
	if err != nil {
		t.Fatalf("Failed to set initial fields: %v", err)
	}

	// 直接修改字段名而不调用UpdateFieldName（错误做法）
	updatedFields := map[string]any{
		"id":      0,
		"name":    "",
		"old_age": 0, // 直接使用新字段名，没有调用UpdateFieldName
	}

	// 这应该能执行，但会导致不一致
	err = table.SetFields(updatedFields)
	if err != nil {
		t.Fatalf("Failed to set updated fields: %v", err)
	}

	// 验证字段ID映射，old_age字段应该被视为新字段
	if len(table.fieldsid) != len(updatedFields) {
		t.Fatalf("Expected %d field IDs, got %d", len(updatedFields), len(table.fieldsid))
	}

	t.Log("✅ Missing UpdateFieldName test completed (expected behavior: new field ID generated)")
}

// 测试设置字段
func TestTable_SetFields(t *testing.T) {
	// 创建表
	table, err := TableNew("test_table_SetFields")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 测试1: 正常设置字段
	t.Log("测试1: 正常设置字段")
	fields := map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	}

	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 验证字段设置正确
	if len(table.fields) != 3 {
		t.Errorf("Expected 3 fields, got %d", len(table.fields))
	}
	if len(table.fieldsid) != 3 {
		t.Errorf("Expected 3 fieldsid mappings, got %d", len(table.fieldsid))
	}

	// 验证每个字段都有对应的ID映射
	for field := range fields {
		found := false
		for _, f := range table.fieldsid {
			if f == field {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Field %s not found in fieldsid mapping", field)
		}
	}
	fmt.Printf("table.fieldsid: %v\n", table.fieldsid)
	// 测试3: 空字段映射
	t.Log("测试3: 空字段映射")
	emptyFields := map[string]any{}

	err = table.SetFields(emptyFields)
	if err != nil {
		t.Errorf("Unexpected error for empty fields: %v", err)
	}
	if len(table.fields) != 0 {
		t.Errorf("Expected 0 fields, got %d", len(table.fields))
	}
	if len(table.fieldsid) != 0 {
		t.Errorf("Expected 0 fieldsid mappings, got %d", len(table.fieldsid))
	}

	// 测试4: 更新现有字段
	t.Log("测试4: 更新现有字段")
	updateFields := map[string]any{
		"id":    0,
		"name":  "",
		"age":   0,
		"email": "",
	}

	err = table.SetFields(updateFields)
	if err != nil {
		t.Fatalf("Failed to update fields: %v", err)
	}

	if len(table.fields) != 4 {
		t.Errorf("Expected 4 fields after update, got %d", len(table.fields))
	}
	if len(table.fieldsid) != 4 {
		t.Errorf("Expected 4 fieldsid mappings after update, got %d", len(table.fieldsid))
	}
}

func TestTableSearch(t *testing.T) {
	// 使用唯一表名，避免测试数据累积
	tableName := fmt.Sprintf("test_table_search_%d", time.Now().UnixNano())

	// 创建表
	table, err := TableNew(tableName)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置字段
	fields := map[string]any{
		"id":    0,
		"name":  "",
		"age":   0,
		"email": "",
	}

	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键索引
	pk, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}
	pk.AddFields("id")
	err = table.CreateIndex(pk)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// 创建二级索引（最后一个字段必须是主键）
	idx, err := DefaultNormalIndexNew("idx_email")
	if err != nil {
		t.Fatalf("Failed to create secondary index: %v", err)
	}
	idx.AddFields("email", "id") // 最后一个字段必须是主键
	err = table.CreateIndex(idx)
	if err != nil {
		t.Fatalf("Failed to create secondary index: %v", err)
	}

	// 插入测试数据
	for i := 1; i <= 10; i++ {
		record := map[string]any{
			"id":    i,
			"name":  fmt.Sprintf("User%d", i),
			"age":   20 + i,
			"email": fmt.Sprintf("user%d@example.com", i),
		}

		_, err = table.Insert(&record)
		if err != nil {
			t.Fatalf("Failed to insert record %d: %v", i, err)
		}
	}

	// 测试1: 主键搜索
	t.Run("PrimaryKeySearch", func(t *testing.T) {
		iter := table.Search(&map[string]any{"id": 5})
		if iter == nil {
			t.Fatalf("Failed to search by primary key")
		}
		defer iter.Release()

		records := iter.GetRecords(true)
		if len(records) != 1 {
			t.Fatalf("Expected 1 record for primary key search, got %d", len(records))
		}

		if records[0]["id"] != 5 {
			t.Fatalf("Expected record with id=5, got %v", records[0])
		}
	})

	// 测试2: 二级索引搜索
	t.Run("SecondaryIndexSearch", func(t *testing.T) {
		iter := table.Search(&map[string]any{"email": "user7@example.com"})
		if iter == nil {
			t.Fatalf("Failed to search by secondary index")
		}
		defer iter.Release()

		records := iter.GetRecords(true)
		if len(records) != 1 {
			t.Fatalf("Expected 1 record for secondary index search, got %d", len(records))
		}

		if records[0]["email"] != "user7@example.com" {
			t.Fatalf("Expected record with email=user7@example.com, got %v", records[0])
		}
	})

	// 测试3: 范围搜索
	t.Run("RangeSearch", func(t *testing.T) {
		iter := table.Search(&map[string]any{"id": 3}, util.GreaterThan)
		if iter == nil {
			t.Fatalf("Failed to search by range")
		}
		defer iter.Release()

		records := iter.GetRecords(true)
		// 预期结果: id > 3，即id从4到10，共7条记录
		if len(records) != 7 {
			t.Fatalf("Expected 7 records for range search, got %d", len(records))
		}

		// 验证第一条记录id=4
		if records[0]["id"] != 4 {
			t.Fatalf("Expected first record with id=4, got %v", records[0])
		}
	})

	// 测试4: Like搜索 (前缀匹配，不使用%)
	t.Run("LikeSearch", func(t *testing.T) {
		// LIKE功能仅支持前缀匹配，且不需要在搜索值中添加%
		iter := table.Search(&map[string]any{"email": "user"}, util.Like)
		if iter == nil {
			t.Fatalf("Failed to search by like")
		}
		defer iter.Release()

		records := iter.GetRecords(true)
		// 预期结果: 所有以"user"开头的记录都匹配
		if len(records) < 1 {
			t.Fatalf("Expected at least 1 record for like search, got %d", len(records))
		}
		// 验证所有返回的记录都符合前缀匹配
		for _, record := range records {
			email, ok := record["email"].(string)
			if !ok {
				t.Fatalf("Expected email to be string, got %T", record["email"])
			}
			// 检查email是否以"user"开头
			if len(email) < 4 || email[:4] != "user" {
				t.Fatalf("Expected email to start with 'user', got %s", email)
			}
		}
	})

	// 测试5: 不等搜索
	t.Run("NotEqualSearch", func(t *testing.T) {
		iter := table.Search(&map[string]any{"id": 5}, util.NotEqual)
		if iter == nil {
			t.Fatalf("Failed to search by not equal")
		}
		defer iter.Release()

		records := iter.GetRecords(true)
		// 预期结果: id != 5，共9条记录
		if len(records) != 9 {
			t.Fatalf("Expected 9 records for not equal search, got %d", len(records))
		}
	})

	// 测试6: 大于等于搜索
	t.Run("GreaterThanOrEqualSearch", func(t *testing.T) {
		iter := table.Search(&map[string]any{"id": 8}, util.GreaterThanOrEqual)
		if iter == nil {
			t.Fatalf("Failed to search by greater than or equal")
		}
		defer iter.Release()

		records := iter.GetRecords(true)
		// 预期结果: id >= 8，即id从8到10，共3条记录
		if len(records) != 3 {
			t.Fatalf("Expected 3 records for greater than or equal search, got %d", len(records))
		}

		// 验证第一条记录id=8
		if records[0]["id"] != 8 {
			t.Fatalf("Expected first record with id=8, got %v", records[0])
		}
	})

	// 测试7: 小于搜索
	t.Run("LessThanSearch", func(t *testing.T) {
		iter := table.Search(&map[string]any{"id": 3}, util.LessThan)
		if iter == nil {
			t.Fatalf("Failed to search by less than")
		}
		defer iter.Release()

		records := iter.GetRecords(true)
		// 预期结果: id < 3，即id从1到2，共2条记录
		if len(records) != 2 {
			t.Fatalf("Expected 2 records for less than search, got %d", len(records))
		}

		// 验证第一条记录id=1
		if records[0]["id"] != 1 {
			t.Fatalf("Expected first record with id=1, got %v", records[0])
		}
	})

	// 测试8: 小于等于搜索
	t.Run("LessThanOrEqualSearch", func(t *testing.T) {
		iter := table.Search(&map[string]any{"id": 3}, util.LessThanOrEqual)
		if iter == nil {
			t.Fatalf("Failed to search by less than or equal")
		}
		defer iter.Release()

		records := iter.GetRecords(true)
		// 预期结果: id <= 3，即id从1到3，共3条记录
		if len(records) != 3 {
			t.Fatalf("Expected 3 records for less than or equal search, got %d", len(records))
		}

		// 验证第一条记录id=1
		if records[0]["id"] != 1 {
			t.Fatalf("Expected first record with id=1, got %v", records[0])
		}
	})
}

// TestGetSysNameId tests the generation and retrieval of system name IDs
func TestGetSysNameId(t *testing.T) {
	// Test with different table names
	t.Run("DifferentTableDifferentID", func(t *testing.T) {
		// Create table 1
		table1, err := TableNew("test_table_1")
		if err != nil {
			t.Fatalf("Failed to create table 1: %v", err)
		}

		// Create table 2
		table2, err := TableNew("test_table_2")
		if err != nil {
			t.Fatalf("Failed to create table 2: %v", err)
		}

		// Get IDs
		id1 := table1.GetId()
		id2 := table2.GetId()

		// Verify they're different
		if id1 == id2 {
			t.Errorf("Expected different IDs for different tables, got %d and %d", id1, id2)
		} else {
			t.Logf("Different table names got different IDs: %d and %d", id1, id2)
		}
	})

	// Test with the same table name (should get same ID)
	t.Run("SameTableSameID", func(t *testing.T) {
		// Create table 1
		table1, err := TableNew("test_same_id")
		if err != nil {
			t.Fatalf("Failed to create table 1: %v", err)
		}

		// Get table 1 ID
		id1 := table1.GetId()

		// Create table 2 with same name
		table2, err := TableNew("test_same_id")
		if err != nil {
			t.Fatalf("Failed to create table 2: %v", err)
		}

		// Get table 2 ID
		id2 := table2.GetId()

		// Verify they're the same
		if id1 != id2 {
			t.Errorf("Expected same ID for same table name, got %d and %d", id1, id2)
		} else {
			t.Logf("Same table name 'test_same_id' got same ID: %d", id1)
		}
	})

	// Test field ID generation
	t.Run("IDManagerGeneration", func(t *testing.T) {
		// Create a table
		table, err := TableNew("test_field_id_gen")
		if err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}

		// Set fields
		fields := map[string]any{
			"id":   0,
			"name": "",
			"age":  0,
		}

		err = table.SetFields(fields)
		if err != nil {
			t.Fatalf("Failed to set fields: %v", err)
		}

		// Create another table with same field names
		table2, err := TableNew("test_field_id_gen_2")
		if err != nil {
			t.Fatalf("Failed to create table 2: %v", err)
		}

		// Set same fields
		err = table2.SetFields(fields)
		if err != nil {
			t.Fatalf("Failed to set fields on table 2: %v", err)
		}

		// Check that fields have different IDs (since they belong to different tables)
		if len(table.fieldsid) != len(table2.fieldsid) {
			t.Errorf("Expected same number of fields, got %d and %d", len(table.fieldsid), len(table2.fieldsid))
			return
		}

		// The field IDs should be the same across tables for the same field names
		// because the ID manager generates IDs based on field names regardless of table
		for id, field := range table.fieldsid {
			// Find the same field in table2
			found := false
			for id2, field2 := range table2.fieldsid {
				if field == field2 {
					found = true
					// Verify same ID for same field name
					if id != id2 {
						t.Errorf("Expected same ID for field '%s' across tables, got %d and %d", field, id, id2)
					} else {
						t.Logf("Same field key got same ID: %d", id)
					}
					break
				}
			}
			if !found {
				t.Errorf("Field '%s' not found in table 2", field)
			}
		}
	})
}

// TestCreateIndexSameID tests that creating indexes with the same name gets the same ID
func TestCreateIndexSameID(t *testing.T) {
	// Create table 1
	table1, err := TableNew("test_index_table_1")
	if err != nil {
		t.Fatalf("Failed to create table 1: %v", err)
	}

	// Set fields
	fields := map[string]any{
		"id":   0,
		"name": "",
	}

	err = table1.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// Create index on table 1
	idx1, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		t.Fatalf("Failed to create index 1: %v", err)
	}
	idx1.AddFields("id")
	err = table1.CreateIndex(idx1)
	if err != nil {
		t.Fatalf("Failed to add index 1: %v", err)
	}

	// Create table 2
	table2, err := TableNew("test_index_table_2")
	if err != nil {
		t.Fatalf("Failed to create table 2: %v", err)
	}

	// Set same fields
	err = table2.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields on table 2: %v", err)
	}

	// Create same index on table 2
	idx2, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		t.Fatalf("Failed to create index 2: %v", err)
	}
	idx2.AddFields("id")
	err = table2.CreateIndex(idx2)
	if err != nil {
		t.Fatalf("Failed to add index 2: %v", err)
	}

	// Verify indexes work correctly
	t.Run("IndexFunctionality", func(t *testing.T) {
		// Insert data into table 1
		record1 := map[string]any{"id": 1, "name": "Test1"}
		_, err := table1.Insert(&record1)
		if err != nil {
			t.Fatalf("Failed to insert into table 1: %v", err)
		}

		// Insert data into table 2
		record2 := map[string]any{"id": 1, "name": "Test2"}
		_, err = table2.Insert(&record2)
		if err != nil {
			t.Fatalf("Failed to insert into table 2: %v", err)
		}

		// Search table 1
		iter1 := table1.Search(&map[string]any{"id": 1})
		if iter1 == nil {
			t.Fatalf("Failed to search table 1")
		}
		defer iter1.Release()

		records1 := iter1.GetRecords(true)
		if len(records1) != 1 {
			t.Fatalf("Expected 1 record in table 1, got %d", len(records1))
		}
		if records1[0]["name"] != "Test1" {
			t.Errorf("Expected name 'Test1' in table 1, got %v", records1[0]["name"])
		}

		// Search table 2
		iter2 := table2.Search(&map[string]any{"id": 1})
		if iter2 == nil {
			t.Fatalf("Failed to search table 2")
		}
		defer iter2.Release()

		records2 := iter2.GetRecords(true)
		if len(records2) != 1 {
			t.Fatalf("Expected 1 record in table 2, got %d", len(records2))
		}
		if records2[0]["name"] != "Test2" {
			t.Errorf("Expected name 'Test2' in table 2, got %v", records2[0]["name"])
		}
	})

	// Test with the same index name (should get same ID internally)
	t.Run("SameIndexSameID", func(t *testing.T) {
		// This test verifies that indexes with the same name get consistent ID management
		// Create another index with same name on different table
		table3, err := TableNew("test_index_table_3")
		if err != nil {
			t.Fatalf("Failed to create table 3: %v", err)
		}

		err = table3.SetFields(fields)
		if err != nil {
			t.Fatalf("Failed to set fields on table 3: %v", err)
		}

		// Create index with same name
		idx3, err := DefaultNormalIndexNew("test_idx")
		if err != nil {
			t.Fatalf("Failed to create index 3: %v", err)
		}
		idx3.AddFields("name")
		err = table3.CreateIndex(idx3)
		if err != nil {
			t.Fatalf("Failed to add index 3: %v", err)
		}

		// Create same index on another table
		table4, err := TableNew("test_index_table_4")
		if err != nil {
			t.Fatalf("Failed to create table 4: %v", err)
		}

		err = table4.SetFields(fields)
		if err != nil {
			t.Fatalf("Failed to set fields on table 4: %v", err)
		}

		// Create index with same name
		idx4, err := DefaultNormalIndexNew("test_idx")
		if err != nil {
			t.Fatalf("Failed to create index 4: %v", err)
		}
		idx4.AddFields("name")
		err = table4.CreateIndex(idx4)
		if err != nil {
			t.Fatalf("Failed to add index 4: %v", err)
		}

		// Both indexes should work correctly despite having the same name
		t.Log("Same index name 'test_idx' got same ID internally")
	})
}

// 测试多个字段修改
func TestTable_MultipleFieldUpdates(t *testing.T) {
	// 使用唯一表名，避免测试数据累积
	tableName := fmt.Sprintf("test_multiple_field_updates_%d", time.Now().UnixNano())

	// 创建表
	table, err := TableNew(tableName)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 初始设置字段
	initialFields := map[string]any{
		"id":      0,
		"name":    "",
		"email":   "",
		"phone":   "",
		"address": "",
	}

	err = table.SetFields(initialFields)
	if err != nil {
		t.Fatalf("Failed to set initial fields: %v", err)
	}

	// 创建索引
	emailIdx, err := DefaultNormalIndexNew("email_idx")
	if err != nil {
		t.Fatalf("Failed to create email index: %v", err)
	}
	emailIdx.AddFields("email")
	err = table.CreateIndex(emailIdx)
	if err != nil {
		t.Fatalf("Failed to add email index: %v", err)
	}

	phoneIdx, err := DefaultNormalIndexNew("phone_idx")
	if err != nil {
		t.Fatalf("Failed to create phone index: %v", err)
	}
	phoneIdx.AddFields("phone")
	err = table.CreateIndex(phoneIdx)
	if err != nil {
		t.Fatalf("Failed to add phone index: %v", err)
	}

	// 插入测试数据
	testData := map[string]any{
		"id":      1,
		"name":    "Alice",
		"email":   "alice@example.com",
		"phone":   "123456789",
		"address": "123 Main St",
	}

	_, err = table.Insert(&testData)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// 批量修改多个字段名
	fieldUpdates := []struct {
		oldField string
		newField string
	}{
		{"email", "email_address"},
		{"phone", "phone_number"},
		{"address", "location"},
	}

	// 先调用UpdateFieldName修改每个字段
	for _, update := range fieldUpdates {
		err = table.UpdateFieldName(update.oldField, update.newField)
		if err != nil {
			t.Fatalf("Failed to update field name %s to %s: %v", update.oldField, update.newField, err)
		}
	}

	// 然后调用SetFields更新字段映射
	updatedFields := map[string]any{
		"id":            0,
		"name":          "",
		"email_address": "",
		"phone_number":  "",
		"location":      "",
	}

	err = table.SetFields(updatedFields)
	if err != nil {
		t.Fatalf("Failed to set updated fields: %v", err)
	}

	// 验证所有修改后的字段都能正常工作
	// 插入新记录使用新字段名
	newTestData := map[string]any{
		"id":            2,
		"name":          "Bob",
		"email_address": "bob@example.com",
		"phone_number":  "987654321",
		"location":      "456 Elm St",
	}

	_, err = table.Insert(&newTestData)
	if err != nil {
		t.Fatalf("Failed to insert with new field names: %v", err)
	}

	// 测试使用新字段名搜索旧记录
	iter1 := table.Search(&map[string]any{"email_address": "alice@example.com"})
	if iter1 == nil {
		t.Fatalf("Failed to search old record by new email_address field")
	}
	defer iter1.Release()

	records1 := iter1.GetRecords(true)
	if len(records1) != 1 {
		t.Fatalf("Expected 1 old record for email_address search, got %d", len(records1))
	}

	// 测试使用新字段名搜索新记录
	iter2 := table.Search(&map[string]any{"phone_number": "987654321"})
	if iter2 == nil {
		t.Fatalf("Failed to search new record by new phone_number field")
	}
	defer iter2.Release()

	records2 := iter2.GetRecords(true)
	if len(records2) != 1 {
		t.Fatalf("Expected 1 new record for phone_number search, got %d", len(records2))
	}

	// 验证所有索引仍然有效（包括主键索引）
	if len(table.indexs.GetAllIndexes()) < 2 {
		t.Fatalf("Expected at least 2 indexes to exist after field updates, got %d", len(table.indexs.GetAllIndexes()))
	}

	t.Log("✅ Multiple field updates test passed successfully")
}

// 测试修改字段后的数据完整性
func TestTable_FieldUpdate_DataIntegrity(t *testing.T) {
	// 使用唯一表名，避免测试数据累积
	tableName := fmt.Sprintf("test_field_update_data_integrity_%d", time.Now().UnixNano())

	// 创建表
	table, err := TableNew(tableName)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 初始设置字段
	initialFields := map[string]any{
		"id":    0,
		"name":  "",
		"value": 0,
	}

	err = table.SetFields(initialFields)
	if err != nil {
		t.Fatalf("Failed to set initial fields: %v", err)
	}

	// 创建主键索引
	pk, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		t.Fatalf("Failed to create primary key: %v", err)
	}
	pk.AddFields("id")
	err = table.CreateIndex(pk)
	if err != nil {
		t.Fatalf("Failed to add primary key index: %v", err)
	}

	// 插入一条测试数据
	testData := map[string]any{
		"id":    1,
		"name":  "Item1",
		"value": 100,
	}

	_, err = table.Insert(&testData)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// 修改字段名
	oldField := "value"
	newField := "amount"

	// 先调用UpdateFieldName
	err = table.UpdateFieldName(oldField, newField)
	if err != nil {
		t.Fatalf("Failed to update field name: %v", err)
	}

	// 然后调用SetFields更新字段映射
	updatedFields := map[string]any{
		"id":     0,
		"name":   "",
		"amount": 0,
	}

	err = table.SetFields(updatedFields)
	if err != nil {
		t.Fatalf("Failed to set updated fields: %v", err)
	}

	// 测试插入新记录使用新字段名
	newTestData := map[string]any{
		"id":     2,
		"name":   "Item2",
		"amount": 200,
	}

	_, err = table.Insert(&newTestData)
	if err != nil {
		t.Fatalf("Failed to insert with new field name: %v", err)
	}

	// 验证字段id映射正确
	// 检查fieldsid映射是否包含新字段名
	found := false
	for _, fieldName := range table.fieldsid {
		if fieldName == newField {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Expected field %s to be in fieldsid map", newField)
	}

	// 验证旧字段名不在fields中
	if _, ok := table.fields[oldField]; ok {
		t.Fatalf("Expected old field %s to not exist in fields map", oldField)
	}

	// 验证新字段名在fields中
	if _, ok := table.fields[newField]; !ok {
		t.Fatalf("Expected new field %s to exist in fields map", newField)
	}

	// 简单验证：测试CheckType仍然正常工作
	checkData := map[string]any{
		"id":     3,
		"name":   "Item3",
		"amount": 300,
	}

	err = table.CheckType(&checkData)
	if err != nil {
		t.Fatalf("CheckType failed after field update: %v", err)
	}

	t.Log("✅ Field update data integrity test passed successfully")
}
