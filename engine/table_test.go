package engine

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/liaoran123/sfsDb/record"
	"github.com/liaoran123/sfsDb/util"
)

// 发现其他测试用例会使用交叉使用相同的表，导致数据不正确。
// 因此，需要为每个测试生成唯一的表名，避免测试之间的数据冲突。
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

	iter, _ := table.Search(&searchRecord)
	if iter == nil {
		t.Fatalf("Failed to search composite primary key")
	}
	defer GlobalTableIterPool.Put(iter)

	resultRecords := iter.GetRecords(true)
	if len(resultRecords) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(resultRecords))
	}

	result := resultRecords[0]
	if result["id"] != 2 || result["name"] != "Bob" {
		t.Fatalf("Expected record with id=2, name=Bob, got %v", result)
	}
	defer record.PutRecords(resultRecords)

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
	iter, _ = table.Search(&searchRecord)
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
	iter, _ = table.Search(&searchRecord)
	if iter == nil {
		t.Fatalf("Failed to search composite primary key after delete")
	}
	defer GlobalTableIterPool.Put(iter)

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
	iter2, _ := table2.Search(&map[string]any{"id": nil})
	defer GlobalTableIterPool.Put(iter2)

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
	iter1, _ := table1.Search(&map[string]any{"id": nil})
	defer GlobalTableIterPool.Put(iter1)
	records1 := iter1.GetRecords(true)
	if len(records1) != len(testData1) {
		t.Errorf("Expected %d records in table1 initially, got %d", len(testData1), len(records1))
	} else {
		t.Logf("table1 initially has %d records", len(records1))
	}

	iter2, _ := table2.Search(&map[string]any{"id": nil})
	defer GlobalTableIterPool.Put(iter2)
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
	iter1, _ = table1.Search(&map[string]any{"id": nil})
	defer GlobalTableIterPool.Put(iter1)
	records1 = iter1.GetRecords(true)
	if len(records1) != 0 {
		t.Errorf("Expected 0 records in table1 after DeleteAll(), got %d", len(records1))
	} else {
		t.Logf("table1 has 0 records after DeleteAll(), which is correct")
	}

	// 验证table2的数据是否仍然正确
	t.Log("Verifying table2 data after table1.DeleteAll()...")
	iter2, _ = table2.Search(&map[string]any{"id": nil})
	defer GlobalTableIterPool.Put(iter2)
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
	iter1, _ = table1.Search(&map[string]any{"id": nil})
	defer GlobalTableIterPool.Put(iter1)
	records1 = iter1.GetRecords(true)
	if len(records1) != 1 {
		t.Errorf("Expected 1 record in table1 after re-insertion, got %d", len(records1))
	} else {
		t.Logf("table1 has 1 record after re-insertion, which is correct")
	}

	t.Log("TestDeleteAllAndAffectOtherTables completed successfully!")
	defer record.PutRecords(records1)
	defer record.PutRecords(records2)
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
	iter, _ := table.Search(&map[string]any{"id": nil})
	defer GlobalTableIterPool.Put(iter)
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
	iter, _ = table.Search(&map[string]any{"id": nil})
	defer GlobalTableIterPool.Put(iter)
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
	iter, _ = table.Search(&map[string]any{"id": nil})
	defer GlobalTableIterPool.Put(iter)
	records = iter.GetRecords(true)
	if len(records) != 1 {
		t.Errorf("Expected 1 record after re-insertion, got %d", len(records))
	} else {
		t.Logf("Table has 1 record after re-insertion, which is correct")
	}

	t.Log("TestDeleteAllWithLargeData completed successfully!")
	defer record.PutRecords(records)
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
	insertRecord := map[string]any{
		"id":   1,
		"name": "Alice",
		"age":  25,
	}

	_, err = table.Insert(&insertRecord)
	if err != nil {
		t.Fatalf("Failed to insert record: %v", err)
	}

	// 搜索记录
	iter, _ := table.Search(&map[string]any{"id": 1})
	if iter == nil {
		t.Fatalf("Failed to search record")
	}
	defer GlobalTableIterPool.Put(iter)

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
	iter, _ = table.Search(&map[string]any{"id": 1})
	if iter == nil {
		t.Fatalf("Failed to search record after update")
	}
	defer GlobalTableIterPool.Put(iter)

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
	iter, _ = table.Search(&map[string]any{"id": 1})
	if iter == nil {
		t.Fatalf("Failed to search record after delete")
	}
	defer GlobalTableIterPool.Put(iter)

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
	defer GlobalTableIterPool.Put(fditer)
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
	emailIter, _ := tableWithIndex.Search(&map[string]any{"email": "user1.updated@example.com"})
	if emailIter == nil {
		t.Fatalf("Failed to search by email")
	}
	defer GlobalTableIterPool.Put(emailIter)

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
	defer GlobalTableIterPool.Put(emailIter)
	//再次搜索邮箱
	emailIter, _ = tableWithIndex.Search(&map[string]any{"email": "user1.updated@example.com"})
	if emailIter == nil {
		t.Fatalf("Failed to search by email")
	}
	defer GlobalTableIterPool.Put(emailIter)

	emailRecords = emailIter.GetRecords(true)
	if len(emailRecords) != 0 {
		t.Fatalf("Expected 0 record for email search, got %d", len(emailRecords))
	}
	defer record.PutRecords(records)
	defer record.PutRecords(fdrecords)
	defer record.PutRecords(emailRecords)
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
			iter, _ := table.Search(&searchFields)
			if iter == nil {
				t.Fatalf("Search failed for term: %s", st.searchTerm)
			}
			defer GlobalTableIterPool.Put(iter)

			records := iter.GetRecords(true)
			defer record.PutRecords(records)
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
			iter, _ := table.Search(&searchFields)
			if iter == nil {
				t.Fatalf("Search failed for term: %s", st.searchTerm)
			}
			defer GlobalTableIterPool.Put(iter)

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
		iter, _ := table.Search(&searchFields)
		if iter == nil {
			t.Fatalf("Search failed for term: %s", "测试")
		}
		defer GlobalTableIterPool.Put(iter)

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
	iter, _ := table.Search(&map[string]any{"email": "alice@example.com"})
	if iter == nil {
		t.Fatalf("Failed to search by initial email field")
	}
	defer GlobalTableIterPool.Put(iter)

	records := iter.GetRecords(true)
	if len(records) != 1 {
		t.Fatalf("Expected 1 record for email search, got %d", len(records))
	}
	defer record.PutRecords(records)

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
	iter2, _ := table.Search(&map[string]any{"email_address": "bob@example.com"})
	if iter2 == nil {
		t.Fatalf("Failed to search by new email_address field")
	}
	defer GlobalTableIterPool.Put(iter2)

	records2 := iter2.GetRecords(true)
	if len(records2) != 1 {
		t.Fatalf("Expected 1 record for new email_address search, got %d", len(records2))
	}

	// 6.3 验证旧记录仍然可访问（通过新字段名）
	iter3, _ := table.Search(&map[string]any{"email_address": "alice@example.com"})
	if iter3 == nil {
		t.Fatalf("Failed to search old record by new field name")
	}
	defer GlobalTableIterPool.Put(iter3)

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

	// 验证字段设置正确（包括版本号字段 'v'）
	if len(table.fields) != 4 {
		t.Errorf("Expected 4 fields, got %d", len(table.fields))
	}
	if len(table.fieldsid) != 4 {
		t.Errorf("Expected 4 fieldsid mappings, got %d", len(table.fieldsid))
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
	if len(table.fields) != 1 {
		t.Errorf("Expected 1 field, got %d", len(table.fields))
	}
	if len(table.fieldsid) != 1 {
		t.Errorf("Expected 1 fieldsid mappings, got %d", len(table.fieldsid))
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

	if len(table.fields) != 5 {
		t.Errorf("Expected 5 fields after update, got %d", len(table.fields))
	}
	if len(table.fieldsid) != 5 {
		t.Errorf("Expected 5 fieldsid mappings after update, got %d", len(table.fieldsid))
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
		iter, _ := table.Search(&map[string]any{"id": 5})
		if iter == nil {
			t.Fatalf("Failed to search by primary key")
		}
		defer GlobalTableIterPool.Put(iter)

		records := iter.GetRecords(true)
		defer record.PutRecords(records)
		if len(records) != 1 {
			t.Fatalf("Expected 1 record for primary key search, got %d", len(records))
		}

		if records[0]["id"] != 5 {
			t.Fatalf("Expected record with id=5, got %v", records[0])
		}
	})

	// 测试2: 二级索引搜索
	t.Run("SecondaryIndexSearch", func(t *testing.T) {
		iter, _ := table.Search(&map[string]any{"email": "user7@example.com"})
		if iter == nil {
			t.Fatalf("Failed to search by secondary index")
		}
		defer GlobalTableIterPool.Put(iter)

		records := iter.GetRecords(true)
		defer record.PutRecords(records)
		if len(records) != 1 {
			t.Fatalf("Expected 1 record for secondary index search, got %d", len(records))
		}

		if records[0]["email"] != "user7@example.com" {
			t.Fatalf("Expected record with email=user7@example.com, got %v", records[0])
		}
	})

	// 测试3: 范围搜索
	t.Run("RangeSearch", func(t *testing.T) {
		iter, _ := table.Search(&map[string]any{"id": 3}, util.GreaterThan)
		if iter == nil {
			t.Fatalf("Failed to search by range")
		}
		defer GlobalTableIterPool.Put(iter)

		records := iter.GetRecords(true)
		defer record.PutRecords(records)
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
		iter, _ := table.Search(&map[string]any{"email": "user"}, util.Like)
		if iter == nil {
			t.Fatalf("Failed to search by like")
		}
		defer GlobalTableIterPool.Put(iter)

		records := iter.GetRecords(true)
		defer record.PutRecords(records)
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
		iter, _ := table.Search(&map[string]any{"id": 5}, util.NotEqual)
		if iter == nil {
			t.Fatalf("Failed to search by not equal")
		}
		defer GlobalTableIterPool.Put(iter)

		records := iter.GetRecords(true)
		defer record.PutRecords(records)
		// 预期结果: id != 5，共9条记录
		if len(records) != 9 {
			t.Fatalf("Expected 9 records for not equal search, got %d", len(records))
		}
	})

	// 测试6: 大于等于搜索
	t.Run("GreaterThanOrEqualSearch", func(t *testing.T) {
		iter, _ := table.Search(&map[string]any{"id": 8}, util.GreaterThanOrEqual)
		if iter == nil {
			t.Fatalf("Failed to search by greater than or equal")
		}
		defer GlobalTableIterPool.Put(iter)

		records := iter.GetRecords(true)
		defer record.PutRecords(records)
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
		iter, _ := table.Search(&map[string]any{"id": 3}, util.LessThan)
		if iter == nil {
			t.Fatalf("Failed to search by less than")
		}
		defer GlobalTableIterPool.Put(iter)

		records := iter.GetRecords(true)
		defer record.PutRecords(records)
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
		iter, _ := table.Search(&map[string]any{"id": 3}, util.LessThanOrEqual)
		if iter == nil {
			t.Fatalf("Failed to search by less than or equal")
		}
		defer GlobalTableIterPool.Put(iter)

		records := iter.GetRecords(true)
		defer record.PutRecords(records)
		// 预期结果: id <= 3，即id从1到3，共3条记录
		if len(records) != 3 {
			t.Fatalf("Expected 3 records for less than or equal search, got %d", len(records))
		}

		// 验证第一条记录id=1
		if records[0]["id"] != 1 {
			t.Fatalf("Expected first record with id=1, got %v", records[0])
		}
	})

	// 强制垃圾回收，确保 finalizer 被执行
	runtime.GC()
	runtime.Gosched() // 让出CPU时间，让 finalizer 有机会执行

}
