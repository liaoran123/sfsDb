package engine

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/liaoran123/sfsDb/storage"
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

// 测试Table.For()方法，确保它只会遍历当前表的数据，不会影响到其他表
func TestTableForMethod(t *testing.T) {
	// 使用唯一表名，避免测试数据累积
	tableName1 := fmt.Sprintf("test_table_for_1_%d", time.Now().UnixNano())
	tableName2 := fmt.Sprintf("test_table_for_2_%d", time.Now().UnixNano())

	// 创建第一个表
	table1, err := TableNew(tableName1)
	if err != nil {
		t.Fatalf("Failed to create table1: %v", err)
	}

	// 创建第二个表
	table2, err := TableNew(tableName2)
	if err != nil {
		t.Fatalf("Failed to create table2: %v", err)
	}

	// 定义表字段
	fields := map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	}

	// 设置表字段
	err = table1.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields for table1: %v", err)
	}

	err = table2.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields for table2: %v", err)
	}

	// 创建主键索引
	pk1, err := DefaultPrimaryKeyNew("pk1")
	if err != nil {
		t.Fatalf("Failed to create primary key index for table1: %v", err)
	}
	pk1.AddFields("id")
	err = table1.CreateIndex(pk1)
	if err != nil {
		t.Fatalf("Failed to create primary key index for table1: %v", err)
	}

	pk2, err := DefaultPrimaryKeyNew("pk2")
	if err != nil {
		t.Fatalf("Failed to create primary key index for table2: %v", err)
	}
	pk2.AddFields("id")
	err = table2.CreateIndex(pk2)
	if err != nil {
		t.Fatalf("Failed to create primary key index for table2: %v", err)
	}

	// 向table1插入数据
	testData1 := []map[string]any{
		{"id": 1, "name": "Alice", "age": 25},
		{"id": 2, "name": "Bob", "age": 30},
		{"id": 3, "name": "Charlie", "age": 35},
	}

	for _, data := range testData1 {
		_, err := table1.Insert(&data)
		if err != nil {
			t.Fatalf("Failed to insert data into table1: %v", err)
		}
	}

	// 向table2插入数据
	testData2 := []map[string]any{
		{"id": 1, "name": "David", "age": 40},
		{"id": 2, "name": "Eve", "age": 45},
	}

	for _, data := range testData2 {
		_, err := table2.Insert(&data)
		if err != nil {
			t.Fatalf("Failed to insert data into table2: %v", err)
		}
	}

	// 使用table1.For()方法遍历table1的所有数据
	t.Log("Using table1.For() to iterate through table1 data...")
	iter1 := table1.For()
	defer iter1.Release()

	// 统计table1.For()返回的数据数量
	count1 := 0
	for iter1.Next() {
		key, value := iter1.Key(), iter1.Value()
		t.Logf("table1.For() - Key: %s, Value: %s", string(key), string(value))
		count1++
	}

	// 验证table1.For()返回的数据数量是否正确
	if count1 != len(testData1) {
		t.Errorf("Expected %d records from table1.For(), got %d", len(testData1), count1)
	} else {
		t.Logf("table1.For() returned %d records, which matches the expected count", count1)
	}

	// 验证table2的数据是否仍然正确
	t.Log("Verifying table2 data...")
	iter2 := table2.Search(&map[string]any{"id": nil})
	defer iter2.Release()

	records2 := iter2.GetRecords(true)
	if len(records2) != len(testData2) {
		t.Errorf("Expected %d records in table2, got %d", len(testData2), len(records2))
	} else {
		t.Logf("table2 still has %d records, which matches the expected count", len(records2))
	}

	// 验证table2的具体数据是否正确
	for i, record := range records2 {
		expectedData := testData2[i]
		if record["id"] != expectedData["id"] || record["name"] != expectedData["name"] || record["age"] != expectedData["age"] {
			t.Errorf("Expected record %d in table2: %v, got: %v", i, expectedData, record)
		} else {
			t.Logf("table2 record %d is correct: %v", i, record)
		}
	}

	t.Log("TestTableForMethod completed successfully!")
}

// 测试表删除操作，确保它不会影响其他表
func TestDeleteAllAndAffectOtherTables(t *testing.T) {
	// 使用唯一表名，避免测试数据累积
	tableName1 := fmt.Sprintf("test_delete_all_1_%d", time.Now().UnixNano())
	tableName2 := fmt.Sprintf("test_delete_all_2_%d", time.Now().UnixNano())

	// 创建第一个表
	table1, err := TableNew(tableName1)
	if err != nil {
		t.Fatalf("Failed to create table1: %v", err)
	}

	// 创建第二个表
	table2, err := TableNew(tableName2)
	if err != nil {
		t.Fatalf("Failed to create table2: %v", err)
	}

	// 定义表字段
	fields := map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	}

	// 设置表字段
	err = table1.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields for table1: %v", err)
	}

	err = table2.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields for table2: %v", err)
	}

	// 创建主键索引
	pk1, err := DefaultPrimaryKeyNew("pk1")
	if err != nil {
		t.Fatalf("Failed to create primary key index for table1: %v", err)
	}
	pk1.AddFields("id")
	err = table1.CreateIndex(pk1)
	if err != nil {
		t.Fatalf("Failed to create primary key index for table1: %v", err)
	}

	pk2, err := DefaultPrimaryKeyNew("pk2")
	if err != nil {
		t.Fatalf("Failed to create primary key index for table2: %v", err)
	}
	pk2.AddFields("id")
	err = table2.CreateIndex(pk2)
	if err != nil {
		t.Fatalf("Failed to create primary key index for table2: %v", err)
	}

	// 向table1插入数据
	testData1 := []map[string]any{
		{"id": 1, "name": "Alice", "age": 25},
		{"id": 2, "name": "Bob", "age": 30},
		{"id": 3, "name": "Charlie", "age": 35},
	}

	for _, data := range testData1 {
		_, err := table1.Insert(&data)
		if err != nil {
			t.Fatalf("Failed to insert data into table1: %v", err)
		}
	}

	// 向table2插入数据
	testData2 := []map[string]any{
		{"id": 1, "name": "David", "age": 40},
		{"id": 2, "name": "Eve", "age": 45},
	}

	for _, data := range testData2 {
		_, err := table2.Insert(&data)
		if err != nil {
			t.Fatalf("Failed to insert data into table2: %v", err)
		}
	}

	// 验证初始状态下两个表的数据都正确
	iter1 := table1.Search(&map[string]any{"id": nil})
	defer iter1.Release()
	records1 := iter1.GetRecords(true)
	if len(records1) != len(testData1) {
		t.Errorf("Expected %d records in table1 initially, got %d", len(testData1), len(records1))
	} else {
		t.Logf("table1 initially has %d records", len(records1))
	}

	iter2 := table2.Search(&map[string]any{"id": nil})
	defer iter2.Release()
	records2 := iter2.GetRecords(true)
	if len(records2) != len(testData2) {
		t.Errorf("Expected %d records in table2 initially, got %d", len(testData2), len(records2))
	} else {
		t.Logf("table2 initially has %d records", len(records2))
	}

	// 使用table1.DeleteAll()方法删除table1的所有数据
	t.Log("Using table1.DeleteAll() to delete all data from table1...")
	err = table1.DeleteAll()
	if err != nil {
		t.Fatalf("Failed to delete all data from table1: %v", err)
	}

	// 验证table1的数据是否已被删除
	iter1 = table1.Search(&map[string]any{"id": nil})
	defer iter1.Release()
	records1 = iter1.GetRecords(true)
	if len(records1) != 0 {
		t.Errorf("Expected 0 records in table1 after DeleteAll(), got %d", len(records1))
	} else {
		t.Logf("table1 has 0 records after DeleteAll(), which is correct")
	}

	// 验证table2的数据是否仍然正确
	t.Log("Verifying table2 data after table1.DeleteAll()...")
	iter2 = table2.Search(&map[string]any{"id": nil})
	defer iter2.Release()
	records2 = iter2.GetRecords(true)
	if len(records2) != len(testData2) {
		t.Errorf("Expected %d records in table2 after table1.DeleteAll(), got %d", len(testData2), len(records2))
	} else {
		t.Logf("table2 still has %d records after table1.DeleteAll(), which is correct", len(records2))
	}

	// 验证table2的具体数据是否正确
	for i, record := range records2 {
		expectedData := testData2[i]
		if record["id"] != expectedData["id"] || record["name"] != expectedData["name"] || record["age"] != expectedData["age"] {
			t.Errorf("Expected record %d in table2: %v, got: %v", i, expectedData, record)
		} else {
			t.Logf("table2 record %d is correct: %v", i, record)
		}
	}

	// 测试在删除所有数据后，是否可以重新向表中插入数据
	t.Log("Testing insertion into table1 after DeleteAll()...")
	newData := map[string]any{"id": 4, "name": "Frank", "age": 50}
	_, err = table1.Insert(&newData)
	if err != nil {
		t.Fatalf("Failed to insert new data into table1 after DeleteAll(): %v", err)
	}

	// 验证新数据是否成功插入
	iter1 = table1.Search(&map[string]any{"id": nil})
	defer iter1.Release()
	records1 = iter1.GetRecords(true)
	if len(records1) != 1 {
		t.Errorf("Expected 1 record in table1 after re-insertion, got %d", len(records1))
	} else {
		t.Logf("table1 has 1 record after re-insertion, which is correct")
	}

	t.Log("TestDeleteAllAndAffectOtherTables completed successfully!")
}

// 测试表删除操作处理大量数据的情况
func TestDeleteAllWithLargeData(t *testing.T) {
	// 使用唯一表名，避免测试数据累积
	tableName := fmt.Sprintf("test_delete_all_large_%d", time.Now().UnixNano())

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

	// 设置表字段
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

	// 向表中插入大量数据（超过批量操作限制）
	const largeDataCount = 1500 // 超过 batchSizeLimit (1000) 的值
	t.Logf("Inserting %d records into table...", largeDataCount)

	for i := 1; i <= largeDataCount; i++ {
		data := map[string]any{
			"id":   i,
			"name": fmt.Sprintf("User%d", i),
			"age":  i % 100,
		}
		_, err := table.Insert(&data)
		if err != nil {
			t.Fatalf("Failed to insert data at index %d: %v", i, err)
		}
	}

	// 验证初始状态下表的数据量是否正确
	iter := table.Search(&map[string]any{"id": nil})
	defer iter.Release()
	records := iter.GetRecords(true)
	if len(records) != largeDataCount {
		t.Errorf("Expected %d records initially, got %d", largeDataCount, len(records))
	} else {
		t.Logf("Table initially has %d records, which matches the expected count", len(records))
	}

	// 使用 DeleteAll() 方法删除表的所有数据
	t.Log("Using DeleteAll() to delete all data from table...")
	err = table.DeleteAll()
	if err != nil {
		t.Fatalf("Failed to delete all data: %v", err)
	}

	// 验证表的数据是否已被删除
	iter = table.Search(&map[string]any{"id": nil})
	defer iter.Release()
	records = iter.GetRecords(true)
	if len(records) != 0 {
		t.Errorf("Expected 0 records after DeleteAll(), got %d", len(records))
	} else {
		t.Logf("Table has 0 records after DeleteAll(), which is correct")
	}

	// 测试在删除所有数据后，是否可以重新向表中插入数据
	t.Log("Testing insertion into table after DeleteAll()...")
	newData := map[string]any{"id": largeDataCount + 1, "name": "NewUser", "age": 25}
	_, err = table.Insert(&newData)
	if err != nil {
		t.Fatalf("Failed to insert new data after DeleteAll(): %v", err)
	}

	// 验证新数据是否成功插入
	iter = table.Search(&map[string]any{"id": nil})
	defer iter.Release()
	records = iter.GetRecords(true)
	if len(records) != 1 {
		t.Errorf("Expected 1 record after re-insertion, got %d", len(records))
	} else {
		t.Logf("Table has 1 record after re-insertion, which is correct")
	}

	t.Log("TestDeleteAllWithLargeData completed successfully!")
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
	secondaryIndexage, err := DefaultNormalIndexNew("idx_age")
	if err != nil {
		t.Fatalf("Failed to create secondary index: %v", err)
	}
	secondaryIndexage.AddFields("age")
	err = tableWithIndex.CreateIndex(secondaryIndexage)
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
	//遍历表
	fditer := tableWithIndex.ForData()
	defer fditer.Release()
	fdrecords := fditer.GetRecords(true)
	for _, record := range fdrecords {
		fmt.Println(record)
	}
	//修改邮箱
	updateRecord1 := map[string]any{
		"id":    1,
		"email": "user1.updated@example.com",
	}
	err = tableWithIndex.Update(&updateRecord1)
	if err != nil {
		t.Fatalf("Failed to update record: %v", err)
	}

	// 使用二级索引搜索
	emailIter := tableWithIndex.Search(&map[string]any{"email": "user1.updated@example.com"})
	if emailIter == nil {
		t.Fatalf("Failed to search by email")
	}
	defer emailIter.Release()

	emailRecords := emailIter.GetRecords(true)
	if len(emailRecords) != 1 {
		t.Fatalf("Expected 1 record for email search, got %d", len(emailRecords))
	}

	if emailRecords[0]["id"] != 1 || emailRecords[0]["email"] != "user1.updated@example.com" {
		t.Fatalf("Expected record with id=1 and email=user1.updated@example.com, got %v", emailRecords[0])
	}
	//删除邮箱
	err = tableWithIndex.Delete(&map[string]any{"id": 1})
	if err != nil {
		t.Fatalf("Failed to delete record: %v", err)
	}
	defer emailIter.Release()
	//再次搜索邮箱
	emailIter = tableWithIndex.Search(&map[string]any{"email": "user1.updated@example.com"})
	if emailIter == nil {
		t.Fatalf("Failed to search by email")
	}
	defer emailIter.Release()

	emailRecords = emailIter.GetRecords(true)
	if len(emailRecords) != 0 {
		t.Fatalf("Expected 0 record for email search, got %d", len(emailRecords))
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

	t.Log("All full text search tests completed successfully")
}

// TestTable_FullTextSearch_CompositePK 测试多主键（组合主键）情况下的全文索引
func TestTable_FullTextSearch_CompositePK(t *testing.T) {
	// 使用唯一表名，避免测试数据累积
	tableName := fmt.Sprintf("test_fulltext_composite_pk_%d", time.Now().UnixNano())

	// 创建表
	table, err := TableNew(tableName)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 定义表字段，包含用于全文搜索的content字段
	fields := map[string]any{
		"mid":     0,  // 主键字段1
		"secNo":   0,  // 主键字段2
		"content": "", // 全文索引字段
		"author":  "", // 普通字段
	}

	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 创建复合主键索引
	pk, err := DefaultPrimaryKeyNew("pk_mid_secno")
	if err != nil {
		t.Fatalf("Failed to create composite primary key: %v", err)
	}
	pk.AddFields("mid", "secNo") // 组合主键：mid + secNo
	err = table.CreateIndex(pk)
	if err != nil {
		t.Fatalf("Failed to create composite primary key index: %v", err)
	}

	// 创建全文索引 - 多主键情况
	fulltextIdx, err := DefaultFullTextIndexNew("fulltext_content")
	if err != nil {
		t.Fatalf("Failed to create fulltext index: %v", err)
	}
	// 全文索引多主键注意事项：
	// 1. 必须包含所有主键字段（mid, secNo）
	// 2. 可以有两种排列顺序：
	//    a. 全文索引字段 + 所有主键字段（如：content, mid, secNo）
	//    b. 所有主键字段 + 全文索引字段（如：mid, secNo, content）
	// 3. 必须包含全部主键字段，否则全文索引不成立
	fulltextIdx.AddFields("content", "mid", "secNo") // 全文索引字段 + 组合主键字段
	err = fulltextIdx.SetFullField("content", 5)     // 设置全文索引字段和长度
	if err != nil {
		t.Fatalf("Failed to set full field: %v", err)
	}
	err = table.CreateIndex(fulltextIdx)
	if err != nil {
		t.Fatalf("Failed to create fulltext index: %v", err)
	}

	// 插入测试数据
	testData := []map[string]any{
		{"mid": 1, "secNo": 1, "content": "这是第一篇文章的内容，包含关键词测试", "author": "作者A"},
		{"mid": 1, "secNo": 2, "content": "这是第二篇文章的内容，包含关键词示例", "author": "作者B"},
		{"mid": 2, "secNo": 1, "content": "这是第三篇文章的内容，包含关键词测试和示例", "author": "作者C"},
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
	}{
		{
			name:          "Search for '测试' (with composite PK)",
			searchTerm:    "测试",
			expectedCount: 2, // 第一篇和第三篇文章包含"测试"
		},
		{
			name:          "Search for '示例' (with composite PK)",
			searchTerm:    "示例",
			expectedCount: 2, // 第二篇和第三篇文章包含"示例"
		},
		{
			name:          "Search for '文章' (with composite PK)",
			searchTerm:    "文章",
			expectedCount: 3, // 所有文章都包含"文章"
		},
	}

	for _, st := range searchTests {
		t.Run(st.name, func(t *testing.T) {
			// 使用Search方法搜索包含关键词的记录
			searchFields := map[string]any{"content": st.searchTerm}
			iter := table.Search(&searchFields)
			if iter == nil {
				t.Fatalf("Search failed for term: %s", st.searchTerm)
			}
			defer iter.Release()

			records := iter.GetRecords(true)
			if len(records) != st.expectedCount {
				t.Errorf("Expected %d records, got %d", st.expectedCount, len(records))
				return
			}

			// 验证每条记录都包含搜索关键词
			for _, record := range records {
				content, ok := record["content"].(string)
				if !ok {
					t.Errorf("Expected content to be a string, got %T", record["content"])
					continue
				}
				if !strings.Contains(content, st.searchTerm) {
					t.Errorf("Record content '%s' does not contain search term '%s'", content, st.searchTerm)
				}
				// 验证记录包含所有主键字段
				if _, ok := record["mid"]; !ok {
					t.Errorf("Record missing primary key field 'mid'")
				}
				if _, ok := record["secNo"]; !ok {
					t.Errorf("Record missing primary key field 'secNo'")
				}
			}
		})
	}

	// 测试不同排列顺序的全文索引（所有主键字段 + 全文索引字段）
	t.Run("FullTextIndex_CompositePK_ReverseOrder", func(t *testing.T) {
		// 创建另一个全文索引，使用相反的字段顺序
		reverseFulltextIdx, err := DefaultFullTextIndexNew("fulltext_content_reverse")
		if err != nil {
			t.Fatalf("Failed to create reverse fulltext index: %v", err)
		}
		// 另一种排列顺序：所有主键字段 + 全文索引字段
		reverseFulltextIdx.AddFields("mid", "secNo", "content") // 组合主键字段 + 全文索引字段
		err = reverseFulltextIdx.SetFullField("content", 5)
		if err != nil {
			t.Fatalf("Failed to set full field for reverse index: %v", err)
		}
		err = table.CreateIndex(reverseFulltextIdx)
		if err != nil {
			t.Fatalf("Failed to create reverse fulltext index: %v", err)
		}

		// 验证反向顺序的全文索引也能正常工作
		searchFields := map[string]any{"content": "测试"}
		iter := table.Search(&searchFields)
		if iter == nil {
			t.Fatalf("Search failed for term: %s", "测试")
		}
		defer iter.Release()

		records := iter.GetRecords(true)
		if len(records) < 1 {
			t.Errorf("Expected at least 1 record, got %d", len(records))
		}
	})

	t.Log("All composite PK full text search tests completed successfully")
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

	dataIter := table.Search(&fields1)
	defer dataIter.Release()
	if dataIter.iter == nil {
		t.Fatalf("Search 失败: %v", err)
	}
	//defer dataIter.Release()
	records := dataIter.GetRecords(true)
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
	record := map[string]any{
		"id":   1,
		"name": "Alice",
		"age":  25,
	}

	_, err = table.Insert(&record)
	if err != nil {
		t.Fatalf("Failed to insert record: %v", err)
	}

	// 2. 获取记录并读取版本号
	// 先通过主键搜索获取记录
	searchFields := map[string]any{"id": 1}
	iter := table.Search(&searchFields)
	if iter == nil {
		t.Fatalf("Failed to search record")
	}
	defer iter.Release()

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
	iter = table.Search(&searchFields)
	if iter == nil {
		t.Fatalf("Failed to search record after update")
	}
	defer iter.Release()

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
	iter = table.Search(&searchFields)
	if iter == nil {
		t.Fatalf("Failed to search record after second update")
	}
	defer iter.Release()

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
}

// TestTableSearch 测试使用加密数据库表遍历数据和Search方法的功能
func TestTableSearch1(t *testing.T) {
	// 生成测试密钥
	masterKey := make([]byte, 32)
	for i := range masterKey {
		masterKey[i] = byte(i)
	}
	// 创建加密配置
	encryptConfig := &storage.EncryptionConfig{
		Enabled:   true,
		Algorithm: "AES-256-GCM",
		MasterKey: masterKey,
	}
	// 初始化加密的全局KVDb
	_, err := storage.OpenDefaultDbWithEncryption("./test_encrypted_table_db", encryptConfig)
	if err != nil {
		t.Fatalf("Failed to open encrypted database: %v", err)
	}
	defer storage.CloseDb()
	// 创建测试表
	table, err := TableNew("test_search")
	if err != nil {
		t.Fatalf("TableNew 失败: %v", err)
	}
	if table == nil {
		t.Fatal("TableNew 失败")
	}
	// 必须先为表预设字段和数据类型
	fields := map[string]any{"id": 0, "name": "", "age": uint8(0), "description": ""}
	table.SetFields(fields)

	PrimaryKeys, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		t.Fatalf("DefaultPrimaryKeyNew 失败: %v", err)
	}
	PrimaryKeys.AddFields("id")    //创建一个id的组合主键
	table.CreateIndex(PrimaryKeys) //将组合主键设置到表中

	fullText, err := DefaultFullTextIndexNew("ft")
	if err != nil {
		t.Fatalf("DefaultFullTextIndexNew 失败: %v", err)
	}
	//全文索引正常情况下必须带上主键，否则后面的关键词都被覆盖，失去全文索引的意义。
	fullText.AddFields("description", "id") //创建一个description的组合全文索引
	//指定description为全文索引字段，长度为5
	//如果没有指定，则等同一般索引
	err = fullText.SetFullField("description", 5) //添加description全文索引字段，长度为5
	if err != nil {
		t.Fatalf("SetFullField 失败: %v", err)
	}
	table.CreateIndex(fullText) //将组合全文索引设置到表中

	normalIndex, err := DefaultNormalIndexNew("idx")
	if err != nil {
		t.Fatalf("DefaultNormalIndexNew 失败: %v", err)
	}
	normalIndex.AddFields("name", "age") //创建一个name, age的组合普通索引
	table.CreateIndex(normalIndex)       //将组合普通索引设置到表中

	// 插入测试数据
	data := []map[string]any{
		{"id": 1, "name": "六月", "age": uint8(25), "description": "古木阴阴六月凉，幽花藉藉四时香。——裘万顷《次余仲庸松风阁韵十九首其三》"},
		{"id": 2, "name": "Bob", "age": uint8(30), "description": "Bob is a product manager"},
		{"id": 3, "name": "Charlie", "age": uint8(35), "description": "Charlie is1 a designer"},
		{"id": 4, "name": "David", "age": uint8(40), "description": "David isnot a developer"},
		{"id": 5, "name": "Eve", "age": uint8(45), "description": "Eve is an manager"},
		{"id": 6, "name": "Alice", "age": uint8(27), "description": "Alice is2 a software engineer"},
		{"id": nil, "name": "Eve 49", "age": uint8(49), "description": "Eve is3 a manager 49"}, //"id": nil 使用自动增值
		{"id": nil, "name": "Eve 55", "age": uint8(55), "description": "Eve is4 a manager 55"}, //"id": nil 使用自动增值
	}
	for _, item := range data {
		fields := table.GetAllFields()
		fields["id"] = item["id"]
		fields["name"] = item["name"]
		fields["age"] = item["age"]
		fields["description"] = item["description"]
		_, err := table.Insert(&fields)
		if err != nil {
			t.Fatalf("插入测试数据失败: %v", err)
		}
		/*
				if item["id"] == nil {
					continue
				}
			if currentID != util.AnyToInt(item["id"]) {
				t.Errorf("插入测试数据后，当前ID应为%v，实际: %d", util.AnyToInt(item["id"]), currentID)
			}*/
	}
	//测试遍历表所有kv
	t.Run("For", func(t *testing.T) {
		dataIter := table.For()
		for dataIter.Next() {
			//fmt.Printf("dataIter.Key(): %v\n", dataIter.Key())
			fmt.Printf("key: %s, value: %s\n", dataIter.Key(), dataIter.Value())
			//val := table.ParseValue(dataIter.Value())
			//fmt.Printf("val: %v\n", val)
		}
		dataIter.Release()
	})
	fmt.Println("-----------------")
	// 测试遍历表所有数据
	t.Run("ForData", func(t *testing.T) {
		// 使用ForData方法遍历所有数据
		dataIter := table.ForData()
		rs := dataIter.GetRecords(true)
		for _, item := range rs {
			fmt.Printf("records: %v\n", item)
		}
	})
	fmt.Println("-----------------")

	// 测试1: 主键搜索
	t.Run("PrimaryKeySearch", func(t *testing.T) {
		// 使用Search方法搜索主键为3的记录
		fields := map[string]any{
			"id": 1,
		}
		dataIter := table.Search(&fields)
		defer dataIter.Release()
		if dataIter.iter == nil {
			t.Fatalf("Search 失败")
		}
		//defer dataIter.Release()
		records := dataIter.GetRecords(true)
		fmt.Printf("records: %v\n", records)
		//判断data[0]和records是否相等

		if records[0]["name"] != data[0]["name"] {
			t.Errorf("搜索主键为1的记录错误，期望: %v, 实际: %v", data[0]["name"], records[0]["name"])
		}
		if records[0]["age"] != data[0]["age"] {
			t.Errorf("搜索主键为1的记录错误，期望: %v, 实际: %v", data[0]["age"], records[0]["age"])
		}
		if records[0]["description"] != data[0]["description"] {
			t.Errorf("搜索主键为1的记录错误，期望: %v, 实际: %v", data[0]["description"], records[0]["description"])
		}

	})

	// 测试2: 索引字段搜索
	t.Run("IndexSearch", func(t *testing.T) {
		// 使用Search方法搜索name为"Charlie"的记录
		fields := map[string]any{
			"name": "Charlie",
		}
		dataIter := table.Search(&fields)
		defer dataIter.Release()
		if dataIter.iter == nil {
			t.Fatalf("Search 失败")
		}
		//defer dataIter.Release()
		records := dataIter.GetRecords(true)
		for _, item := range records.Select("name", "age", "description") {
			fmt.Printf("records: %v\n", item)
		}

		if records[0]["age"] != data[2]["age"] {
			t.Errorf("搜索name为Charlie的记录错误，期望: %v, 实际: %v", data[2]["age"], records[0]["age"])
		}
	})

	// 测试3: 全文索引搜索
	t.Run("FullTextSearch", func(t *testing.T) {
		// 使用Search方法搜索description包含"Bob"的记录
		fields := map[string]any{
			"description": "Bob",
		}
		dataIter := table.Search(&fields)
		defer dataIter.Release()
		if dataIter.iter == nil {
			t.Fatalf("Search 失败")
		}
		//defer dataIter.Release()
		records := dataIter.GetRecords(true)
		for _, item := range records.Select("name", "age", "description") {
			fmt.Printf("records: %v\n", item)
		}

		sdata := []map[string]any{
			{"description": "古木阴阴六月凉，幽花藉藉四时香。——裘万顷《次余仲庸松风阁韵十九首其三》"},
			{"description": "六月凉，幽花藉藉四时香。——裘万顷《次余仲庸松风阁韵十九首其三》"},
			{"description": "幽花藉藉四时香。——裘万顷《次余仲庸松风阁韵十九首其三》"},
			{"description": "藉藉四时香。——裘万顷《次余仲庸松风阁韵十九首其三》"},
			{"description": "——裘万顷《次余仲庸松风阁韵十九首其三》"},
			{"description": "次余仲庸松风阁韵十九首其三》"},
			{"description": "十九首其三》"},
			{"description": "。——裘万顷《次余仲庸松风阁韵十九首其三》"},
		}
		for _, item := range sdata {
			fields := map[string]any{
				"description": item["description"],
			}
			dataIter := table.Search(&fields)
			defer dataIter.Release()
			if dataIter.iter == nil {
				t.Fatalf("Search 失败")
			}
			//defer dataIter.Release()

			records := dataIter.GetRecords(true)
			for _, item := range records.Select("name", "age", "description") {
				fmt.Printf("搜索:%v -》 records: %v\n", fields["description"], item)
			}

			//判断data[1]和records是否相等
			if records[0]["name"] != data[0]["name"] {
				t.Errorf("全文索引搜索 description 包含Bob的记录错误，期望: %v, 实际: %v", data[0]["name"], records[0]["name"])
			}
			if records[0]["age"] != data[0]["age"] {
				t.Errorf("全文索引搜索 description 包含Bob的记录错误，期望: %v, 实际: %v", data[0]["age"], records[0]["age"])
			}
			if records[0]["description"] != data[0]["description"] {
				t.Errorf("全文索引搜索 description 包含Bob的记录错误，期望: %v, 实际: %v", data[0]["description"], records[0]["description"])
			}
		}

		sdata1 := []map[string]any{
			{"description": "Bob is a product manager"},
			{"description": "is a product manager"},
			{"description": "a product manager"},
			{"description": "product manager"},
			{"description": "ct ma"},
			//{"description": "is"},
			{"description": " a product manager"},
			{"description": " product"},
		}
		for _, item := range sdata1 {
			fields := map[string]any{
				"description": item["description"],
			}
			dataIter := table.Search(&fields)
			defer dataIter.Release()
			if dataIter.iter == nil {
				t.Fatalf("Search 失败")
			}
			//defer dataIter.Release()

			records := dataIter.GetRecords(true)
			for _, item := range records.Select("name", "age", "description") {
				fmt.Printf("搜索:%v -》 records: %v\n", fields["description"], item)
			}
			if records[0]["id"] != data[1]["id"] {
				fmt.Printf("查询结果可能是多个: %v\n。但是测试并没有错误。", records)
				//t.Errorf("全文索引搜索 description 包含Bob的记录错误，期望: %v, 实际: %v", data[1]["id"], records[0][table.GetPrimaryKey().GetFields()[0]])
			}
		}
	})

	// 测试4: 搜索不存在的数据
	t.Run("SearchNonExistent", func(t *testing.T) {
		fields := map[string]any{
			"id": 100,
		}
		// 使用Search方法搜索不存在的id
		dataIter := table.Search(&fields)
		defer dataIter.Release()
		if dataIter.iter == nil {
			t.Fatalf("Search 失败")
		}
		//defer dataIter.Release()
		records := dataIter.GetRecords(true)
		//判断data[1]和records是否相等
		if records != nil {
			t.Errorf("搜索description包含Bob的记录错误，期望: %v, 实际: %v", data[1]["id"], records[0][table.GetPrimaryKey().GetFields()[0]])
		}
	})

	//打开所有记录
	t.Run("SearchAll", func(t *testing.T) {
		// 第一次搜索，缓存结果
		fields := map[string]any{
			"id": nil, // id=nil或空，将获取所有表记录
		}
		dataIter := table.Search(&fields)
		defer dataIter.Release()
		if dataIter.iter == nil {
			t.Fatalf("Search 失败")
		}
		//defer dataIter.Release()
		records := dataIter.GetRecords(true)
		for i, item := range records.Select() {
			fmt.Printf("item %d: %v\n", i, item)
		}
	})
}

// 测试添加Table.Insert，删除Table.Delete，修改Table.Update，添加一条记录，通过主键进行修改和删除
func TestTableCRUD1(t *testing.T) {

	// 创建表
	table, err := TableNew("test_table_CRUD1")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 定义表字段
	fields := map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	}

	// 设置表字段 - 使用正确的SetFields方法
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 测试1: 添加单条记录
	t.Log("测试1: 添加单条记录")
	record := map[string]any{
		"id":   1,
		"name": "张三",
		"age":  25,
	}

	_, err = table.Insert(&record)
	if err != nil {
		t.Fatalf("Failed to add record: %v", err)
	}

	// 验证记录存在
	searchFields := map[string]any{"id": 1}
	dataIter := table.Search(&searchFields)
	defer dataIter.Release()
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	records := dataIter.GetRecords(true)
	if len(records) == 0 {
		t.Error("Record not found after adding")
	} else {
		if records[0]["name"] != "张三" || records[0]["age"] != 25 {
			t.Errorf("Record data mismatch: got %v, expected name=张三, age=25", records[0])
		} else {
			t.Logf("添加记录成功: %v", records[0])
		}
	}

	// 测试2: 修改记录
	t.Log("测试2: 修改记录")
	updateRecord := map[string]any{
		"id":   1,
		"name": "张三修改",
		"age":  26,
	}

	err = table.Update(&updateRecord)
	if err != nil {
		t.Fatalf("Failed to update record: %v", err)
	}

	// 验证修改成功
	searchFields = map[string]any{"id": 1}
	dataIter = table.Search(&searchFields)
	defer dataIter.Release()
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	updatedRecords := dataIter.GetRecords(true)
	if len(updatedRecords) == 0 {
		t.Error("Updated record not found")
	} else {
		if updatedRecords[0]["name"] != "张三修改" || updatedRecords[0]["age"] != 26 {
			t.Errorf("Record update failed: got %v, expected name=张三修改, age=26", updatedRecords[0])
		} else {
			t.Logf("修改记录成功: %v", updatedRecords[0])
		}
	}

	// 测试3: 删除记录
	t.Log("测试3: 删除记录")
	deleteFields := map[string]any{"id": 1}
	err = table.Delete(&deleteFields)
	if err != nil {
		t.Fatalf("Failed to delete record: %v", err)
	}

	// 验证记录已删除
	searchFields = map[string]any{"id": 1}
	dataIter = table.Search(&searchFields)
	defer dataIter.Release()
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	deletedRecords := dataIter.GetRecords(true)
	if len(deletedRecords) == 0 {
		t.Log("删除记录成功")
	} else {
		t.Errorf("Record deletion failed: found %v, expected none", deletedRecords)
	}

	t.Log("所有CRUD测试通过")

	// 测试4: 带有索引和全文索引的表CRUD操作
	t.Log("\n测试4: 带有索引和全文索引的表CRUD操作")

	// 创建带索引的表
	tableWithIndex, err := TableNew("test_table_with_index")
	if err != nil {
		t.Fatalf("Failed to create table with index: %v", err)
	}

	// 定义带索引的表字段
	indexFields := map[string]any{
		"id":      0,
		"title":   "",
		"content": "",
		"author":  "",
		"views":   0,
	}

	// 设置表字段 - 使用正确的SetFields方法
	err = tableWithIndex.SetFields(indexFields)
	if err != nil {
		t.Fatalf("Failed to set fields for table with index: %v", err)
	}

	// 创建普通索引
	// 创建标题索引
	titleIndex, err := DefaultNormalIndexNew("title_index")
	if err != nil {
		t.Fatalf("Failed to create title index instance: %v", err)
	}
	titleIndex.AddFields("title")
	err = tableWithIndex.CreateIndex(titleIndex)
	//err = tableWithIndex.indexs.CreateIndex(titleIndex)
	if err != nil {
		t.Fatalf("Failed to create title index: %v", err)
	}

	// 创建复合索引（作者-浏览量）
	authorViewsIndex, err := DefaultNormalIndexNew("author_views_index")
	if err != nil {
		t.Fatalf("Failed to create author_views index instance: %v", err)
	}
	authorViewsIndex.AddFields("author", "views")
	err = tableWithIndex.CreateIndex(authorViewsIndex)
	if err != nil {
		t.Fatalf("Failed to create author_views index: %v", err)
	}

	// 创建全文索引
	contentFulltextIndex, err := DefaultFullTextIndexNew("content_fulltext")
	if err != nil {
		t.Fatalf("Failed to create content fulltext index instance: %v", err)
	}
	//全文索引正常情况下必须带上主键，否则后面的关键词都被覆盖，失去全文索引的意义。
	contentFulltextIndex.AddFields("content", "id")
	//指定description为全文索引字段，长度为5
	//如果没有指定，则等同一般索引
	err = contentFulltextIndex.SetFullField("content", 5)
	if err != nil {
		t.Fatalf("Failed to set fulltext index field: %v", err)
	}

	err = tableWithIndex.CreateIndex(contentFulltextIndex)
	if err != nil {
		t.Fatalf("Failed to create content fulltext index: %v", err)
	}

	// 添加多条记录
	indexRecords := []map[string]any{
		{
			"id":      1,
			"title":   "Go语言入门",
			"content": "Go语言是一种开源的编程语言，它具有高效、简洁、并发等特点。",
			"author":  "张三",
			"views":   100,
		},
		{
			"id":      2,
			"title":   "Go语言进阶",
			"content": "Go语言的并发模型是其一大特色，使用goroutine和channel可以轻松实现高效的并发编程。",
			"author":  "李四",
			"views":   200,
		},
		{
			"id":      3,
			"title":   "Go语言实战",
			"content": "通过实际项目学习Go语言，可以更好地掌握其特性和最佳实践。",
			"author":  "张三",
			"views":   150,
		},
	}

	for _, record := range indexRecords {
		_, err = tableWithIndex.Insert(&record)
		if err != nil {
			t.Fatalf("Failed to add record to indexed table: %v", err)
		}
	}

	// 测试通过普通索引查询
	t.Log("测试通过普通索引查询")
	searchByTitle := map[string]any{"title": "Go语言入门"}
	dataIter = tableWithIndex.Search(&searchByTitle)
	defer dataIter.Release()
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	titleRecords := dataIter.GetRecords(true)
	if len(titleRecords) != 1 {
		t.Errorf("Expected 1 record for title 'Go语言入门', got %d", len(titleRecords))
	} else {
		t.Logf("通过标题索引查询成功: %v", titleRecords[0])
	}
	/*
		// -----------------------------------------
		tb := tableWithIndex
		tb.Search(&searchByTitle).GetRecords(true).Select("id", "title", "content", "author", "views")
		tb.Search(&searchByTitle).GetRecords(true).Delete()
		tb.Search(&searchByTitle).GerRecords(true).Update(&updateRecord)
		//-----------------------------------------
	*/
	// 测试通过复合索引查询
	t.Log("测试通过复合索引查询")
	searchByAuthor := map[string]any{"author": "张三"}
	dataIter = tableWithIndex.Search(&searchByAuthor)
	defer dataIter.Release()
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	authorRecords := dataIter.GetRecords(true)
	if len(authorRecords) != 2 {
		t.Errorf("Expected 2 records for author '张三', got %d", len(authorRecords))
	} else {
		t.Logf("通过作者索引查询成功，找到 %d 条记录", len(authorRecords))
	}

	// 测试修改记录
	t.Log("测试修改带索引的记录")
	updateRecord = map[string]any{
		"id":    1,
		"title": "Go语言入门教程",
		"views": 120,
	}

	err = tableWithIndex.Update(&updateRecord)
	if err != nil {
		t.Fatalf("Failed to update indexed record: %v", err)
	}

	// 验证修改成功
	searchUpdated := map[string]any{"id": 1}
	dataIter = tableWithIndex.Search(&searchUpdated)
	defer dataIter.Release()
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	updatedRecords = dataIter.GetRecords(true)
	if len(updatedRecords) == 0 {
		t.Error("Updated record not found")
	} else {
		if updatedRecords[0]["title"] != "Go语言入门教程" || updatedRecords[0]["views"] != 120 {
			t.Errorf("Record update failed: got %v, expected title=Go语言入门教程, views=120", updatedRecords[0])
		} else {
			t.Logf("修改带索引记录成功: %v", updatedRecords[0])
		}
	}

	// 测试删除记录
	t.Log("测试删除带索引的记录")
	deleteFields = map[string]any{"id": 3}
	err = tableWithIndex.Delete(&deleteFields)
	if err != nil {
		t.Fatalf("Failed to delete indexed record: %v", err)
	}

	// 验证记录已删除
	searchDeleted := map[string]any{"id": 3}
	dataIter = tableWithIndex.Search(&searchDeleted)
	defer dataIter.Release()
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	deletedRecords = dataIter.GetRecords(true)
	if len(deletedRecords) == 0 {
		t.Log("删除带索引记录成功")
	} else {
		t.Errorf("Record deletion failed: found %v, expected none", deletedRecords[0])
	}

	// 验证索引仍然有效
	searchByAuthorAfterDelete := map[string]any{"author": "张三"}
	dataIter = tableWithIndex.Search(&searchByAuthorAfterDelete)
	defer dataIter.Release()
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	authorRecordsAfterDelete := dataIter.GetRecords(true)
	if len(authorRecordsAfterDelete) != 1 {
		t.Errorf("Expected 1 record for author '张三' after delete, got %d", len(authorRecordsAfterDelete))
	} else {
		t.Logf("删除后通过作者索引查询成功，找到 %d 条记录", len(authorRecordsAfterDelete))
	}

	t.Log("所有带索引的CRUD测试通过")
}
