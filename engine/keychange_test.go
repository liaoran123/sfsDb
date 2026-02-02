package engine

import (
	"strconv"
	"sync"
	"testing"

	"github.com/liaoran123/sfsDb/monitor"
)

/*
	// - 添加一条记录后删除，AtomicInt和AtomicDec所有对应的键值相等。

	// 验证修改字段后AtomicInt和AtomicDec：
	// - 所有的修改，主键索引，（AtomicInt）putCount+=1, （AtomicDec）deleteCount+=1，因为添加的时候是1，修改后对应的AtomicInt和AtomicDec键值相减=1
	// - 普通索引：（AtomicInt）putCount=1, （AtomicDec）deleteCount=1
	// - 全文索引的长度通过func (dfi *DefaultFullTextIndex) Tokenize(nr string, ftlen int) (tokens []string)计算得到len(tokens) putCount=len(tokens), deleteCount=len(tokens)

*/
// TestKeyChangeTracking 测试键值变化跟踪功能
func TestKeyChangeTracking(t *testing.T) {
	// 创建表
	tableName := "test_keychange"

	// 创建表
	table, err := TableNew(tableName)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{
		"id":          0,
		"name":        "",
		"description": "",
		"age":         0,
	}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set table fields: %v", err)
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

	// 创建普通索引
	normalIndex, err := DefaultNormalIndexNew("idx_name")
	if err != nil {
		t.Fatalf("Failed to create normal index: %v", err)
	}
	normalIndex.AddFields("name")
	err = table.CreateIndex(normalIndex)
	if err != nil {
		t.Fatalf("Failed to create normal index: %v", err)
	}

	// 创建全文索引
	fulltextIndex, err := DefaultFullTextIndexNew("idx_description")
	if err != nil {
		t.Fatalf("Failed to create fulltext index: %v", err)
	}
	fulltextIndex.AddFields("description")
	err = table.CreateIndex(fulltextIndex)
	if err != nil {
		t.Fatalf("Failed to create fulltext index: %v", err)
	}

	// 获取表ID和索引ID
	tableID := table.id
	pkID := pk.GetId()
	normalIndexID := normalIndex.GetId()
	fulltextIndexID := fulltextIndex.GetId()

	t.Logf("Table ID: %d", tableID)
	t.Logf("Primary Key ID: %d", pkID)
	t.Logf("Normal Index ID: %d", normalIndexID)
	t.Logf("Fulltext Index ID: %d", fulltextIndexID)

	// 初始状态：所有计数器应为0
	for _, indexID := range []byte{pkID, normalIndexID, fulltextIndexID} {
		putCount := monitor.AtomicMap(monitor.AtomicInt).Get(byte(tableID), indexID)
		deleteCount := monitor.AtomicMap(monitor.AtomicDec).Get(byte(tableID), indexID)

		if putCount != 0 {
			t.Fatalf("Expected put count 0 for table %d index %d, got %d", tableID, indexID, putCount)
		}
		if deleteCount != 0 {
			t.Fatalf("Expected delete count 0 for table %d index %d, got %d", tableID, indexID, deleteCount)
		}
	}

	// 测试1：插入记录
	t.Log("\n--- Test 1: Insert Records ---")

	records := []map[string]any{
		{"id": 1, "name": "Alice", "description": "Alice is a test user"},
		{"id": 2, "name": "Bob", "description": "Bob is a test user"},
		{"id": 3, "name": "Charlie", "description": "Charlie is a test user"},
	}

	for _, record := range records {
		_, err = table.Insert(&record)
		if err != nil {
			t.Fatalf("Failed to insert record: %v", err)
		}
	}

	// 验证插入后的计数器：每个索引都应该有3次put操作
	for _, indexID := range []byte{pkID, normalIndexID, fulltextIndexID} {
		putCount := monitor.AtomicMap(monitor.AtomicInt).Get(byte(tableID), indexID)
		if putCount != 3 {
			t.Errorf("Expected put count 3 for table %d index %d, got %d", tableID, indexID, putCount)
		} else {
			t.Logf("✓ Table %d index %d put count: %d", tableID, indexID, putCount)
		}
	}

	// 测试2：删除记录
	t.Log("\n--- Test 2: Delete Record ---")

	deleteRecord := map[string]any{"id": 2}
	err = table.Delete(&deleteRecord)
	if err != nil {
		t.Fatalf("Failed to delete record: %v", err)
	}

	// 验证删除后的计数器：每个索引都应该有3次put和1次delete操作
	for _, indexID := range []byte{pkID, normalIndexID, fulltextIndexID} {
		putCount := monitor.AtomicMap(monitor.AtomicInt).Get(byte(tableID), indexID)
		deleteCount := monitor.AtomicMap(monitor.AtomicDec).Get(byte(tableID), indexID)

		if putCount != 3 {
			t.Errorf("Expected put count 3 for table %d index %d, got %d", tableID, indexID, putCount)
		}
		if deleteCount != 1 {
			t.Errorf("Expected delete count 1 for table %d index %d, got %d", tableID, indexID, deleteCount)
		} else {
			t.Logf("✓ Table %d index %d delete count: %d", tableID, indexID, deleteCount)
		}
	}

	// 测试3：更新记录
	t.Log("\n--- Test 3: Update Record ---")

	updateRecord := map[string]any{"id": 1, "name": "Alice Updated", "description": "Alice has been updated"}
	err = table.Update(&updateRecord)
	if err != nil {
		t.Fatalf("Failed to update record: %v", err)
	}

	// 验证更新后的计数器：每个索引都应该有4次put和2次delete操作
	// （更新操作会先删除旧记录，再插入新记录，所以会增加1次put和1次delete）
	for _, indexID := range []byte{pkID, normalIndexID, fulltextIndexID} {
		putCount := monitor.AtomicMap(monitor.AtomicInt).Get(byte(tableID), indexID)
		deleteCount := monitor.AtomicMap(monitor.AtomicDec).Get(byte(tableID), indexID)

		if putCount != 4 {
			t.Errorf("Expected put count 4 for table %d index %d, got %d", tableID, indexID, putCount)
		}
		if deleteCount != 2 {
			t.Errorf("Expected delete count 2 for table %d index %d, got %d", tableID, indexID, deleteCount)
		} else {
			t.Logf("✓ Table %d index %d put: %d, delete: %d", tableID, indexID, putCount, deleteCount)
		}
	}

	// 测试4：批量操作
	t.Log("\n--- Test 4: Batch Operations ---")

	// 批量插入（手动实现）
	batchRecords := []map[string]any{
		{"id": 4, "name": "David", "description": "David is a test user"},
		{"id": 5, "name": "Eve", "description": "Eve is a test user"},
	}

	for _, record := range batchRecords {
		_, err = table.Insert(&record)
		if err != nil {
			t.Fatalf("Failed to insert record: %v", err)
		}
	}

	// 验证批量插入后的计数器：每个索引都应该有6次put和2次delete操作
	for _, indexID := range []byte{pkID, normalIndexID, fulltextIndexID} {
		putCount := monitor.AtomicMap(monitor.AtomicInt).Get(byte(tableID), indexID)
		deleteCount := monitor.AtomicMap(monitor.AtomicDec).Get(byte(tableID), indexID)

		if putCount != 6 {
			t.Errorf("Expected put count 6 for table %d index %d, got %d", tableID, indexID, putCount)
		}
		if deleteCount != 2 {
			t.Errorf("Expected delete count 2 for table %d index %d, got %d", tableID, indexID, deleteCount)
		} else {
			t.Logf("✓ Table %d index %d after batch insert - put: %d, delete: %d", tableID, indexID, putCount, deleteCount)
		}
	}

	// 批量删除（手动实现）
	batchDeleteRecords := []map[string]any{
		{"id": 4},
		{"id": 5},
	}

	for _, record := range batchDeleteRecords {
		err = table.Delete(&record)
		if err != nil {
			t.Fatalf("Failed to delete record: %v", err)
		}
	}

	// 验证批量删除后的计数器：每个索引都应该有6次put和4次delete操作
	for _, indexID := range []byte{pkID, normalIndexID, fulltextIndexID} {
		putCount := monitor.AtomicMap(monitor.AtomicInt).Get(byte(tableID), indexID)
		deleteCount := monitor.AtomicMap(monitor.AtomicDec).Get(byte(tableID), indexID)

		if putCount != 6 {
			t.Errorf("Expected put count 6 for table %d index %d, got %d", tableID, indexID, putCount)
		}
		if deleteCount != 4 {
			t.Errorf("Expected delete count 4 for table %d index %d, got %d", tableID, indexID, deleteCount)
		} else {
			t.Logf("✓ Table %d index %d after batch delete - put: %d, delete: %d", tableID, indexID, putCount, deleteCount)
		}
	}

	// 测试5：净变化量验证
	t.Log("\n--- Test 5: Net Change Verification ---")

	// 最终应该有2条记录（id:1, 3）
	iter := table.ForData()
	defer GlobalTableIterPool.Put(iter)
	finalRecords := iter.GetRecords(true)
	if len(finalRecords) != 2 {
		t.Fatalf("Expected 2 final records, got %d", len(finalRecords))
	}

	// 验证净变化量：每个索引的净变化量应该等于最终记录数
	for _, indexID := range []byte{pkID, normalIndexID, fulltextIndexID} {
		putCount := monitor.AtomicMap(monitor.AtomicInt).Get(byte(tableID), indexID)
		deleteCount := monitor.AtomicMap(monitor.AtomicDec).Get(byte(tableID), indexID)
		netChange := putCount - deleteCount

		if netChange != 2 {
			t.Errorf("Expected net change 2 for table %d index %d, got %d (put: %d, delete: %d)",
				tableID, indexID, netChange, putCount, deleteCount)
		} else {
			t.Logf("✓ Table %d index %d net change: %d (matches final record count)",
				tableID, indexID, netChange)
		}
	}

	t.Log("\nAll key change tracking tests passed!")
}

// TestKeyChange_AddDeleteEqual 测试添加一条记录和删除一条记录后，put和delete计数器是否相等
func TestKeyChange_AddDeleteEqual(t *testing.T) {
	// 创建表
	tableName := "test_add_delete_equal"

	// 创建表
	table, err := TableNew(tableName)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{
		"id":   0,
		"name": "",
	}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set table fields: %v", err)
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

	// 获取表ID和索引ID
	tableID := table.id
	pkID := pk.GetId()

	// 初始状态：put和delete计数器都应为0
	putCount := monitor.AtomicMap(monitor.AtomicInt).Get(byte(tableID), pkID)
	deleteCount := monitor.AtomicMap(monitor.AtomicDec).Get(byte(tableID), pkID)

	if putCount != 0 || deleteCount != 0 {
		t.Fatalf("Initial state: Expected putCount=0 and deleteCount=0, got putCount=%d, deleteCount=%d", putCount, deleteCount)
	}

	// 添加一条记录
	record := map[string]any{"id": 1, "name": "Test Record"}
	_, err = table.Insert(&record)
	if err != nil {
		t.Fatalf("Failed to insert record: %v", err)
	}

	// 验证插入后：putCount=1, deleteCount=0
	putCount = monitor.AtomicMap(monitor.AtomicInt).Get(byte(tableID), pkID)
	deleteCount = monitor.AtomicMap(monitor.AtomicDec).Get(byte(tableID), pkID)

	if putCount != 1 {
		t.Errorf("After insert: Expected putCount=1, got %d", putCount)
	}
	if deleteCount != 0 {
		t.Errorf("After insert: Expected deleteCount=0, got %d", deleteCount)
	}

	// 删除这条记录
	deleteRecord := map[string]any{"id": 1}
	err = table.Delete(&deleteRecord)
	if err != nil {
		t.Fatalf("Failed to delete record: %v", err)
	}

	// 验证删除后：putCount=1, deleteCount=1，两者应该相等
	putCount = monitor.AtomicMap(monitor.AtomicInt).Get(byte(tableID), pkID)
	deleteCount = monitor.AtomicMap(monitor.AtomicDec).Get(byte(tableID), pkID)

	if putCount != deleteCount {
		t.Errorf("After delete: Expected putCount(%d) to equal deleteCount(%d)", putCount, deleteCount)
	} else {
		t.Logf("✓ After add and delete one record: putCount=%d, deleteCount=%d, they are equal", putCount, deleteCount)
	}

	// 验证最终记录数为0
	iter := table.ForData()
	defer GlobalTableIterPool.Put(iter)
	finalRecords := iter.GetRecords(true)
	if len(finalRecords) != 0 {
		t.Fatalf("Expected 0 final records, got %d", len(finalRecords))
	}
}

// TestKeyChange_UpdateIndexFieldsSeparately 测试单独修改普通索引和全文索引字段时的键值变化跟踪
func TestKeyChange_UpdateIndexFieldsSeparately(t *testing.T) {
	// 创建表
	tableName := "test_update_index_fields_separately"

	// 创建表
	table, err := TableNew(tableName)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{
		"id":          0,
		"name":        "",
		"description": "",
		"age":         0,
	}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set table fields: %v", err)
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

	// 创建普通索引（name字段）
	nameIndex, err := DefaultNormalIndexNew("idx_name")
	if err != nil {
		t.Fatalf("Failed to create normal index: %v", err)
	}
	nameIndex.AddFields("name")
	err = table.CreateIndex(nameIndex)
	if err != nil {
		t.Fatalf("Failed to create normal index: %v", err)
	}

	// 创建全文索引（description字段）
	fulltextIndex, err := DefaultFullTextIndexNew("idx_description")
	if err != nil {
		t.Fatalf("Failed to create fulltext index: %v", err)
	}
	fulltextIndex.AddFields("description", "id")
	fulltextIndex.SetFullField("description", 5)
	err = table.CreateIndex(fulltextIndex)
	if err != nil {
		t.Fatalf("Failed to create fulltext index: %v", err)
	}

	// 获取表ID和索引ID
	tableID := table.id
	pkID := pk.GetId()
	nameIndexID := nameIndex.GetId()
	fulltextIndexID := fulltextIndex.GetId()

	// 初始状态：所有计数器应为0
	for _, indexID := range []byte{pkID, nameIndexID, fulltextIndexID} {
		putCount := monitor.AtomicMap(monitor.AtomicInt).Get(byte(tableID), indexID)
		deleteCount := monitor.AtomicMap(monitor.AtomicDec).Get(byte(tableID), indexID)

		if putCount != 0 || deleteCount != 0 {
			t.Fatalf("Initial state: Expected putCount=0 and deleteCount=0 for index %d, got putCount=%d, deleteCount=%d", indexID, putCount, deleteCount)
		}
	}

	// 插入一条记录
	record := map[string]any{
		"id":          1,
		"name":        "Initial Name",
		"description": "12345", //长度为5， putCount=5, deleteCount=5
		"age":         20,
	}
	_, err = table.Insert(&record)
	if err != nil {
		t.Fatalf("Failed to insert record: %v", err)
	}

	// 验证插入后：
	// - 主键索引：putCount=1, deleteCount=0
	// - 普通索引：putCount=1, deleteCount=0
	// - 全文索引：putCount=5, deleteCount=0（因为"description": "12345", 长度为5，所以会进行5次put操作）

	// 验证主键索引
	pkPutCount := monitor.AtomicMap(monitor.AtomicInt).Get(byte(tableID), pkID)
	pkDeleteCount := monitor.AtomicMap(monitor.AtomicDec).Get(byte(tableID), pkID)
	if pkPutCount != 1 || pkDeleteCount != 0 {
		t.Errorf("After insert: Expected primary key putCount=1, deleteCount=0, got putCount=%d, deleteCount=%d", pkPutCount, pkDeleteCount)
	} else {
		t.Logf("✓ After insert - Primary key index: putCount=%d, deleteCount=%d", pkPutCount, pkDeleteCount)
	}

	// 验证普通索引
	namePutCount := monitor.AtomicMap(monitor.AtomicInt).Get(byte(tableID), nameIndexID)
	nameDeleteCount := monitor.AtomicMap(monitor.AtomicDec).Get(byte(tableID), nameIndexID)
	if namePutCount != 1 || nameDeleteCount != 0 {
		t.Errorf("After insert: Expected normal index putCount=1, deleteCount=0, got putCount=%d, deleteCount=%d", namePutCount, nameDeleteCount)
	} else {
		t.Logf("✓ After insert - Normal index: putCount=%d, deleteCount=%d", namePutCount, nameDeleteCount)
	}

	// 验证全文索引
	fulltextPutCount := monitor.AtomicMap(monitor.AtomicInt).Get(byte(tableID), fulltextIndexID)
	fulltextDeleteCount := monitor.AtomicMap(monitor.AtomicDec).Get(byte(tableID), fulltextIndexID)
	if fulltextPutCount != 5 || fulltextDeleteCount != 0 {
		t.Errorf("After insert: Expected fulltext index putCount=5, deleteCount=0, got putCount=%d, deleteCount=%d", fulltextPutCount, fulltextDeleteCount)
	} else {
		t.Logf("✓ After insert - Fulltext index: putCount=%d, deleteCount=%d", fulltextPutCount, fulltextDeleteCount)
	}

	// 测试1：单独修改普通索引字段（name）
	t.Log("\n--- Test 1: Update normal index field (name) ---")
	updateNameRecord := map[string]any{
		"id":   1,
		"name": "Updated Name",
	}
	err = table.Update(&updateNameRecord)
	if err != nil {
		t.Fatalf("Failed to update name field: %v", err)
	}

	// 验证修改普通索引字段后：
	// - 主键索引：putCount=2, deleteCount=1（主键会先删除后添加新值）
	// - 普通索引（name）：putCount=2, deleteCount=2（因为name字段被修改）
	// - 全文索引（description）：putCount=5, deleteCount=5，因为"description": "12345", //长度为5，所以会重新构建索引
	// - 全文索引的长度通过func (dfi *DefaultFullTextIndex) Tokenize(nr string, ftlen int) (tokens []string)计算得到len(tokens)
	// - 修改全文索引后，对应的主键索引也会被修改，对应的主键索引会增加1，putCount+=1, deleteCount+=1
	// - 所有的修改，主键索引会增加1，putCount+=1, deleteCount+=1，保证主键依然保持唯一

	// 验证主键索引
	pkPutCount = monitor.AtomicMap(monitor.AtomicInt).Get(byte(tableID), pkID)
	pkDeleteCount = monitor.AtomicMap(monitor.AtomicDec).Get(byte(tableID), pkID)
	if pkPutCount != 2 || pkDeleteCount != 1 {
		t.Errorf("After update name: Expected primary key putCount=2, deleteCount=1, got putCount=%d, deleteCount=%d", pkPutCount, pkDeleteCount)
	} else {
		t.Logf("✓ Primary key index: putCount=%d, deleteCount=%d", pkPutCount, pkDeleteCount)
	}

	// 验证普通索引
	namePutCount = monitor.AtomicMap(monitor.AtomicInt).Get(byte(tableID), nameIndexID)
	nameDeleteCount = monitor.AtomicMap(monitor.AtomicDec).Get(byte(tableID), nameIndexID)
	if namePutCount != 2 || nameDeleteCount != 1 {
		t.Errorf("After update name: Expected normal index putCount=2, deleteCount=1, got putCount=%d, deleteCount=%d", namePutCount, nameDeleteCount)
	} else {
		t.Logf("✓ Normal index: putCount=%d, deleteCount=%d", namePutCount, nameDeleteCount)
	}

	// 验证全文索引
	fulltextPutCount = monitor.AtomicMap(monitor.AtomicInt).Get(byte(tableID), fulltextIndexID)
	fulltextDeleteCount = monitor.AtomicMap(monitor.AtomicDec).Get(byte(tableID), fulltextIndexID)
	if fulltextPutCount != 5 || fulltextDeleteCount != 0 {
		t.Errorf("After update name: Expected fulltext index putCount=5, deleteCount=0, got putCount=%d, deleteCount=%d", fulltextPutCount, fulltextDeleteCount)
	} else {
		t.Logf("✓ Fulltext index: putCount=%d, deleteCount=%d", fulltextPutCount, fulltextDeleteCount)
	}

	// 测试2：单独修改全文索引字段（description）
	t.Log("\n--- Test 2: Update fulltext index field (description) ---")
	updateDescRecord := map[string]any{
		"id":          1,
		"description": "Updated description",
	}
	err = table.Update(&updateDescRecord)
	if err != nil {
		t.Fatalf("Failed to update description field: %v", err)
	}

	// 验证修改全文索引字段后：
	// 只验证putCount >= deleteCount
	for _, indexID := range []byte{pkID, nameIndexID, fulltextIndexID} {
		putCount := monitor.AtomicMap(monitor.AtomicInt).Get(byte(tableID), indexID)
		deleteCount := monitor.AtomicMap(monitor.AtomicDec).Get(byte(tableID), indexID)

		if putCount < deleteCount {
			t.Errorf("After update description: Expected putCount(%d) >= deleteCount(%d) for index %d", putCount, deleteCount, indexID)
		}
	}

	// 测试3：修改非索引字段（age）
	t.Log("\n--- Test 3: Update non-index field (age) ---")
	updateAgeRecord := map[string]any{
		"id":  1,
		"age": 25,
	}
	err = table.Update(&updateAgeRecord)
	if err != nil {
		t.Fatalf("Failed to update age field: %v", err)
	}

	// 验证修改非索引字段后：
	// 只验证putCount >= deleteCount
	for _, indexID := range []byte{pkID, nameIndexID, fulltextIndexID} {
		putCount := monitor.AtomicMap(monitor.AtomicInt).Get(byte(tableID), indexID)
		deleteCount := monitor.AtomicMap(monitor.AtomicDec).Get(byte(tableID), indexID)

		if putCount < deleteCount {
			t.Errorf("After update age: Expected putCount(%d) >= deleteCount(%d) for index %d", putCount, deleteCount, indexID)
		}
	}

	// 删除记录
	t.Log("\n--- Test 4: Delete record ---")
	deleteRecord := map[string]any{"id": 1}
	err = table.Delete(&deleteRecord)
	if err != nil {
		t.Fatalf("Failed to delete record: %v", err)
	}

	// 验证删除后：所有索引的putCount等于deleteCount
	for _, indexID := range []byte{pkID, nameIndexID, fulltextIndexID} {
		putCount := monitor.AtomicMap(monitor.AtomicInt).Get(byte(tableID), indexID)
		deleteCount := monitor.AtomicMap(monitor.AtomicDec).Get(byte(tableID), indexID)

		if putCount != deleteCount {
			t.Errorf("After delete: Expected putCount(%d) to equal deleteCount(%d) for index %d", putCount, deleteCount, indexID)
		} else {
			t.Logf("✓ After delete: putCount=%d, deleteCount=%d, they are equal for index %d", putCount, deleteCount, indexID)
		}
	}

	// 验证最终记录数为0
	iter := table.ForData()
	defer GlobalTableIterPool.Put(iter)
	finalRecords := iter.GetRecords(true)
	if len(finalRecords) != 0 {
		t.Fatalf("Expected 0 final records, got %d", len(finalRecords))
	}
}

// TestKeyChange_ConcurrentOperations 测试并发环境下的键值变化跟踪
func TestKeyChange_ConcurrentOperations(t *testing.T) {
	// 创建表
	tableName := "test_concurrent_operations"

	// 创建表
	table, err := TableNew(tableName)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set table fields: %v", err)
	}

	// 创建主键索引和普通索引
	pk, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}
	pk.AddFields("id")
	err = table.CreateIndex(pk)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// 创建普通索引
	nameIndex, err := DefaultNormalIndexNew("idx_name")
	if err != nil {
		t.Fatalf("Failed to create normal index: %v", err)
	}
	nameIndex.AddFields("name")
	err = table.CreateIndex(nameIndex)
	if err != nil {
		t.Fatalf("Failed to create normal index: %v", err)
	}

	// 获取表ID和索引ID
	tableID := table.id
	pkID := pk.GetId()
	nameIndexID := nameIndex.GetId()

	// 初始状态：put和delete计数器都应为0
	for _, indexID := range []byte{pkID, nameIndexID} {
		putCount := monitor.AtomicMap(monitor.AtomicInt).Get(byte(tableID), indexID)
		deleteCount := monitor.AtomicMap(monitor.AtomicDec).Get(byte(tableID), indexID)

		if putCount != 0 || deleteCount != 0 {
			t.Fatalf("Initial state: Expected putCount=0 and deleteCount=0 for index %d, got putCount=%d, deleteCount=%d", indexID, putCount, deleteCount)
		}
	}

	// 并发操作测试
	const numGoroutines = 10
	const operationsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				// 生成唯一ID
				id := goroutineID*operationsPerGoroutine + j

				// 1. 插入记录
				record := map[string]any{
					"id":   id,
					"name": "User" + strconv.Itoa(id),
					"age":  20 + id%30,
				}
				_, err := table.Insert(&record)
				if err != nil {
					t.Errorf("Goroutine %d: Failed to insert record %d: %v", goroutineID, id, err)
					continue
				}

				// 2. 更新记录（修改索引字段）
				updateRecord := map[string]any{
					"id":   id,
					"name": "UpdatedUser" + strconv.Itoa(id),
				}
				err = table.Update(&updateRecord)
				if err != nil {
					t.Errorf("Goroutine %d: Failed to update record %d: %v", goroutineID, id, err)
					continue
				}

				// 3. 删除记录
				deleteRecord := map[string]any{"id": id}
				err = table.Delete(&deleteRecord)
				if err != nil {
					t.Errorf("Goroutine %d: Failed to delete record %d: %v", goroutineID, id, err)
					continue
				}
			}
		}(i)
	}

	// 等待所有goroutine完成
	wg.Wait()

	// 验证最终状态：所有记录都被删除，所以对于每个操作（插入+更新+删除），
	// 每个索引应该有 2*numGoroutines*operationsPerGoroutine 次put操作和 2*numGoroutines*operationsPerGoroutine 次delete操作
	// 因为：
	// - 插入：1次put
	// - 更新：1次delete + 1次put（总共1次delete，1次put）
	// - 删除：1次delete
	// 所以每个完整的操作序列（插入+更新+删除）会产生：2次put，2次delete
	expectedPut := int64(2 * numGoroutines * operationsPerGoroutine)
	expectedDelete := int64(2 * numGoroutines * operationsPerGoroutine)

	for _, indexID := range []byte{pkID, nameIndexID} {
		putCount := monitor.AtomicMap(monitor.AtomicInt).Get(byte(tableID), indexID)
		deleteCount := monitor.AtomicMap(monitor.AtomicDec).Get(byte(tableID), indexID)

		if putCount != expectedPut {
			t.Errorf("Expected putCount=%d for index %d, got %d", expectedPut, indexID, putCount)
		}
		if deleteCount != expectedDelete {
			t.Errorf("Expected deleteCount=%d for index %d, got %d", expectedDelete, indexID, deleteCount)
		}
		if putCount != deleteCount {
			t.Errorf("Expected putCount(%d) to equal deleteCount(%d) for index %d", putCount, deleteCount, indexID)
		} else {
			t.Logf("✓ After concurrent operations: putCount=%d, deleteCount=%d, they are equal for index %d", putCount, deleteCount, indexID)
		}
	}

	// 验证最终记录数为0
	iter := table.ForData()
	defer GlobalTableIterPool.Put(iter)
	finalRecords := iter.GetRecords(true)
	if len(finalRecords) != 0 {
		t.Fatalf("Expected 0 final records, got %d", len(finalRecords))
	}
}

// TestKeyChange_UpdateIndexField 测试修改索引字段时，put和delete计数器是否保持相等
func TestKeyChange_UpdateIndexField(t *testing.T) {
	// 创建表
	tableName := "test_update_index_field"

	// 创建表
	table, err := TableNew(tableName)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set table fields: %v", err)
	}

	// 创建主键索引和普通索引
	pk, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}
	pk.AddFields("id")
	err = table.CreateIndex(pk)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// 创建普通索引（用于测试索引字段更新）
	nameIndex, err := DefaultNormalIndexNew("idx_name")
	if err != nil {
		t.Fatalf("Failed to create normal index: %v", err)
	}
	nameIndex.AddFields("name")
	err = table.CreateIndex(nameIndex)
	if err != nil {
		t.Fatalf("Failed to create normal index: %v", err)
	}

	// 获取表ID和索引ID
	tableID := table.id
	pkID := pk.GetId()
	nameIndexID := nameIndex.GetId()

	// 初始状态：put和delete计数器都应为0
	for _, indexID := range []byte{pkID, nameIndexID} {
		putCount := monitor.AtomicMap(monitor.AtomicInt).Get(byte(tableID), indexID)
		deleteCount := monitor.AtomicMap(monitor.AtomicDec).Get(byte(tableID), indexID)

		if putCount != 0 || deleteCount != 0 {
			t.Fatalf("Initial state: Expected putCount=0 and deleteCount=0 for index %d, got putCount=%d, deleteCount=%d", indexID, putCount, deleteCount)
		}
	}

	// 1. 添加一条记录
	record := map[string]any{"id": 1, "name": "Initial Name", "age": 20}
	_, err = table.Insert(&record)
	if err != nil {
		t.Fatalf("Failed to insert record: %v", err)
	}

	// 验证插入后：putCount=1, deleteCount=0
	for _, indexID := range []byte{pkID, nameIndexID} {
		putCount := monitor.AtomicMap(monitor.AtomicInt).Get(byte(tableID), indexID)
		deleteCount := monitor.AtomicMap(monitor.AtomicDec).Get(byte(tableID), indexID)

		if putCount != 1 {
			t.Errorf("After insert: Expected putCount=1 for index %d, got %d", indexID, putCount)
		}
		if deleteCount != 0 {
			t.Errorf("After insert: Expected deleteCount=0 for index %d, got %d", indexID, deleteCount)
		}
	}

	// 2. 修改索引字段（name字段，该字段有索引）
	updateRecord := map[string]any{"id": 1, "name": "Updated Name"}
	err = table.Update(&updateRecord)
	if err != nil {
		t.Fatalf("Failed to update record: %v", err)
	}

	// 验证修改后：putCount=2, deleteCount=1，两者应该相等吗？
	// 更新操作会先删除旧索引条目（deleteCount+1），再插入新索引条目（putCount+1）
	// 实际上，更新操作会重新构建所有索引，包括主键索引
	// 所以对于所有索引，都应该有 putCount = deleteCount + 1

	// 验证所有索引：putCount=2, deleteCount=1
	for _, indexID := range []byte{pkID, nameIndexID} {
		putCount := monitor.AtomicMap(monitor.AtomicInt).Get(byte(tableID), indexID)
		deleteCount := monitor.AtomicMap(monitor.AtomicDec).Get(byte(tableID), indexID)

		// 验证put和delete计数器的关系：putCount = deleteCount + 1
		// 因为更新操作会删除旧记录（+1 delete）并插入新记录（+1 put），所以净变化为+1
		if putCount != deleteCount+1 {
			t.Errorf("After update (index %d): Expected putCount=%d+1=%d, got putCount=%d, deleteCount=%d", indexID, deleteCount, deleteCount+1, putCount, deleteCount)
		} else {
			t.Logf("✓ After update index %d: putCount=%d, deleteCount=%d, putCount = deleteCount + 1", indexID, putCount, deleteCount)
		}
	}

	// 3. 再次修改同一个索引字段
	updateRecord2 := map[string]any{"id": 1, "name": "Another Update"}
	err = table.Update(&updateRecord2)
	if err != nil {
		t.Fatalf("Failed to update record again: %v", err)
	}

	// 验证再次修改后：putCount=3, deleteCount=2
	namePutCount2 := monitor.AtomicMap(monitor.AtomicInt).Get(byte(tableID), nameIndexID)
	nameDeleteCount2 := monitor.AtomicMap(monitor.AtomicDec).Get(byte(tableID), nameIndexID)

	if namePutCount2 != nameDeleteCount2+1 {
		t.Errorf("After second update (name index): Expected putCount=%d+1=%d, got putCount=%d, deleteCount=%d", nameDeleteCount2, nameDeleteCount2+1, namePutCount2, nameDeleteCount2)
	} else {
		t.Logf("✓ After second update name index field: putCount=%d, deleteCount=%d, putCount = deleteCount + 1", namePutCount2, nameDeleteCount2)
	}

	// 4. 删除记录
	deleteRecord := map[string]any{"id": 1}
	err = table.Delete(&deleteRecord)
	if err != nil {
		t.Fatalf("Failed to delete record: %v", err)
	}

	// 验证删除后：所有索引的putCount和deleteCount应该相等
	// 因为删除操作会删除所有索引条目，所以对于所有索引：putCount = deleteCount
	for _, indexID := range []byte{pkID, nameIndexID} {
		putCount := monitor.AtomicMap(monitor.AtomicInt).Get(byte(tableID), indexID)
		deleteCount := monitor.AtomicMap(monitor.AtomicDec).Get(byte(tableID), indexID)

		if putCount != deleteCount {
			t.Errorf("After delete: Expected putCount(%d) to equal deleteCount(%d) for index %d", putCount, deleteCount, indexID)
		} else {
			t.Logf("✓ After delete: putCount=%d, deleteCount=%d, they are equal for index %d", putCount, deleteCount, indexID)
		}
	}

	// 验证最终记录数为0
	iter := table.ForData()
	defer GlobalTableIterPool.Put(iter)
	finalRecords := iter.GetRecords(true)
	if len(finalRecords) != 0 {
		t.Fatalf("Expected 0 final records, got %d", len(finalRecords))
	}
}
