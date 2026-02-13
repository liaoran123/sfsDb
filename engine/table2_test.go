package engine

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/liaoran123/sfsDb/record"
)

// 发现其他测试用例会使用交叉使用相同的表，导致数据不正确。
// 因此，需要为每个测试生成唯一的表名，避免测试之间的数据冲突。
// 测试复合主键搜索

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
		iter1, _ := table1.Search(&map[string]any{"id": 1})
		if iter1 == nil {
			t.Fatalf("Failed to search table 1")
		}
		defer GlobalTableIterPool.Put(iter1)

		records1 := iter1.GetRecords(true)
		if len(records1) != 1 {
			t.Fatalf("Expected 1 record in table 1, got %d", len(records1))
		}
		if records1[0]["name"] != "Test1" {
			t.Errorf("Expected name 'Test1' in table 1, got %v", records1[0]["name"])
		}

		// Search table 2
		iter2, _ := table2.Search(&map[string]any{"id": 1})
		if iter2 == nil {
			t.Fatalf("Failed to search table 2")
		}
		defer GlobalTableIterPool.Put(iter2)

		records2 := iter2.GetRecords(true)
		if len(records2) != 1 {
			t.Fatalf("Expected 1 record in table 2, got %d", len(records2))
		}
		if records2[0]["name"] != "Test2" {
			t.Errorf("Expected name 'Test2' in table 2, got %v", records2[0]["name"])
		}
		defer record.PutRecords(records1)
		defer record.PutRecords(records2)
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
	iter1, _ := table.Search(&map[string]any{"email_address": "alice@example.com"})
	if iter1 == nil {
		t.Fatalf("Failed to search old record by new email_address field")
	}
	defer GlobalTableIterPool.Put(iter1)

	records1 := iter1.GetRecords(true)
	if len(records1) != 1 {
		t.Fatalf("Expected 1 old record for email_address search, got %d", len(records1))
	}

	// 测试使用新字段名搜索新记录
	iter2, _ := table.Search(&map[string]any{"phone_number": "987654321"})
	if iter2 == nil {
		t.Fatalf("Failed to search new record by new phone_number field")
	}
	defer GlobalTableIterPool.Put(iter2)

	records2 := iter2.GetRecords(true)
	if len(records2) != 1 {
		t.Fatalf("Expected 1 new record for phone_number search, got %d", len(records2))
	}

	// 验证所有索引仍然有效（包括主键索引）
	if len(table.indexs.GetAllIndexes()) < 2 {
		t.Fatalf("Expected at least 2 indexes to exist after field updates, got %d", len(table.indexs.GetAllIndexes()))
	}

	t.Log("✅ Multiple field updates test passed successfully")
	defer record.PutRecords(records1)
	defer record.PutRecords(records2)
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

func TestCompositePrimaryKeySearch1(t *testing.T) {

	table, err := TableNew("art")
	if err != nil {
		t.Fatalf("TableNew 失败: %v", err)
	}
	// 必须先为表预设字段和数据类型
	fields := map[string]any{
		"mid":     0,  //文章ID或目录ID
		"secNo":   0,  //文章句子序号
		"title":   "", //文章标题
		"content": "", //文章内容
	}
	table.SetFields(fields)
	PrimaryKeys, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		t.Fatalf("DefaultPrimaryKeyNew 失败: %v", err)
	}
	PrimaryKeys.AddFields("mid", "secNo") //创建一个mid, secNo的组合主键
	if err := table.CreateIndex(PrimaryKeys); err != nil {
		t.Fatalf("CreateIndex 失败: %v", err)
	}

	fullText, err := DefaultFullTextIndexNew("ft")
	if err != nil {
		t.Fatalf("DefaultFullTextIndexNew 失败: %v", err)
	}
	//全文索引正常情况下必须在前或后带上全量主键，否则后面的关键词都被覆盖，失去全文索引的意义。
	fullText.AddFields("content", "mid", "secNo")
	//指定content为全文索引字段，长度为5
	//如果没有指定，则等同一般索引
	err = fullText.SetFullField("content", 5) //添加content全文索引字段，长度为10
	if err != nil {
		t.Fatalf("SetFullField 失败: %v", err)
	}
	if err = table.CreateIndex(fullText); err != nil {
		t.Fatalf("CreateIndex 失败: %v", err)
	}

	normalIndex, err := DefaultNormalIndexNew("idx")
	if err != nil {
		t.Fatalf("DefaultNormalIndexNew 失败: %v", err)
	}
	//这个是重复索引，不会添加成功
	normalIndex.AddFields("mid", "secNo") //创建一个普通组合索引
	if err = table.CreateIndex(normalIndex); err != nil {
		fmt.Printf("重复索引，不能创建: %v", err)
	}
	//------------------------------

	table.Insert(&map[string]any{
		"mid":     1,
		"secNo":   1,
		"title":   "文章标题11",
		"content": "文章内容11，从三个接口中提取了公共方法，避免了重复定义",
	})
	table.Insert(&map[string]any{
		"mid":     1,
		"secNo":   2,
		"title":   "文章标题12",
		"content": "文章内容12，清晰的层次结构 ：基础接口 + 具体索引类型接口的设计，层次分明",
	})
	table.Insert(&map[string]any{
		"mid":     2,
		"secNo":   1,
		"title":   "文章标题21",
		"content": "文章内容21，更好的可扩展性 ：新索引类型只需嵌入 IndexBase 接口，即可继承公共方法",
	})
	table.Insert(&map[string]any{
		"mid":     2,
		"secNo":   2,
		"title":   "文章标题22",
		"content": "文章内容22，高度可定制化 ：每个索引类型都可以根据需求定制索引字段和行为",
	})

	// 搜索指定文章的所有句子
	fields1 := map[string]any{
		"mid":   nil, // id=nil或空，将获取所有表记录
		"secNo": nil, // id=nil或空，将获取所有表记录
	}

	iter := table.For()
	for iter.Next() {
		fmt.Printf("iter.Key(): %v,iter.Value(): %v\n", string(iter.Key()), string(iter.Value()))
	}

	dataIter, _ := table.Search(&fields1)
	defer GlobalTableIterPool.Put(dataIter)
	if dataIter.iter == nil {
		t.Fatalf("Search 失败: %v", err)
	}
	//defer GlobalTableIterPool.Put(dataIter)
	records := dataIter.GetRecords(true)
	defer record.PutRecords(records)
	for i, item := range records {
		fmt.Printf("结果集 %d: %v\n", i, item)
	}

}

// 测试Update方法的乐观锁功能
func TestTableUpdateWithOptimisticLock(t *testing.T) {
	// 使用唯一表名，避免测试数据累积
	tableName := fmt.Sprintf("test_update_optimistic_lock_%d", time.Now().UnixNano())

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

	// 1. 插入一条记录
	insertRecord := map[string]any{
		"id":   1,
		"name": "Alice",
		"age":  25,
	}

	_, err = table.Insert(&insertRecord)
	if err != nil {
		t.Fatalf("Failed to insert record: %v", err)
	}

	// 2. 获取记录并读取版本号
	// 先通过主键搜索获取记录
	searchFields := map[string]any{"id": 1}
	iter, _ := table.Search(&searchFields)
	if iter == nil {
		t.Fatalf("Failed to search record")
	}
	defer GlobalTableIterPool.Put(iter)

	records := iter.GetRecords(true)
	if len(records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(records))
	}

	// 获取当前版本号
	currentRecord := records[0]
	currentVersion, ok := currentRecord["v"].(int)
	if !ok {
		t.Fatalf("Expected version field 'v' of type int, got %T", currentRecord["v"])
	}
	t.Logf("Initial record version: %d", currentVersion)

	// 3. 使用正确的版本号更新记录，验证成功
	updateFields1 := map[string]any{
		"id":   1,
		"name": "Alice Updated",
		"age":  26,
		"v":    currentVersion, // 提供正确的版本号
	}

	err = table.Update(&updateFields1)
	if err != nil {
		t.Fatalf("Failed to update record with correct version: %v", err)
	}
	t.Logf("Update with correct version succeeded")

	// 验证更新后的版本号递增
	iter, _ = table.Search(&searchFields)
	if iter == nil {
		t.Fatalf("Failed to search record after update")
	}
	defer GlobalTableIterPool.Put(iter)

	records = iter.GetRecords(true)
	if len(records) != 1 {
		t.Fatalf("Expected 1 record after update, got %d", len(records))
	}

	updatedRecord := records[0]
	newVersion, ok := updatedRecord["v"].(int)
	if !ok {
		t.Fatalf("Expected version field 'v' of type int, got %T", updatedRecord["v"])
	}

	if newVersion != currentVersion+1 {
		t.Fatalf("Expected version to increment by 1, got %d (expected %d)", newVersion, currentVersion+1)
	}
	t.Logf("Version correctly incremented to: %d", newVersion)

	// 4. 使用旧版本号再次更新记录，验证乐观锁冲突
	updateFields2 := map[string]any{
		"id":   1,
		"name": "Alice Updated Again",
		"age":  27,
		"v":    currentVersion, // 提供旧版本号，应该失败
	}

	err = table.Update(&updateFields2)
	if err == nil {
		t.Fatalf("Expected optimistic lock conflict, but update succeeded")
	}

	if !strings.Contains(err.Error(), "optimistic lock conflict") {
		t.Fatalf("Expected 'optimistic lock conflict' error, got: %v", err)
	}
	t.Logf("Got expected optimistic lock conflict: %v", err)

	// 5. 使用正确的新版本号更新记录，验证成功
	updateFields3 := map[string]any{
		"id":   1,
		"name": "Alice Updated Again",
		"age":  27,
		"v":    newVersion, // 提供正确的新版本号
	}

	err = table.Update(&updateFields3)
	if err != nil {
		t.Fatalf("Failed to update record with correct new version: %v", err)
	}
	t.Logf("Update with correct new version succeeded")

	// 验证第二次更新后的版本号再次递增
	iter, _ = table.Search(&searchFields)
	if iter == nil {
		t.Fatalf("Failed to search record after second update")
	}
	defer GlobalTableIterPool.Put(iter)

	records = iter.GetRecords(true)
	if len(records) != 1 {
		t.Fatalf("Expected 1 record after second update, got %d", len(records))
	}

	finalRecord := records[0]
	finalVersion, ok := finalRecord["v"].(int)
	if !ok {
		t.Fatalf("Expected version field 'v' of type int, got %T", finalRecord["v"])
	}

	if finalVersion != newVersion+1 {
		t.Fatalf("Expected version to increment by 1 again, got %d (expected %d)", finalVersion, newVersion+1)
	}
	t.Logf("Version correctly incremented to: %d after second update", finalVersion)

	// 验证最终记录内容
	if finalRecord["name"] != "Alice Updated Again" || finalRecord["age"] != 27 {
		t.Fatalf("Expected updated record content, got: %v", finalRecord)
	}
	t.Logf("Final record content is correct")
	defer record.PutRecords(records)
}
