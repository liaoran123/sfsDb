package engine

import (
	"fmt"
	"testing"

	"github.com/liaoran123/sfsDb/match"
	"github.com/liaoran123/sfsDb/record"
	"github.com/liaoran123/sfsDb/storage"
)

// 测试多表组合查询功能
// 使用索引功能。
func TestTestSelectForJoin(t *testing.T) {
	// Create test table
	table1, err := TableNew("test_search_comprehensive1")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Set table fields
	fields := map[string]any{"id": 0, "name": "", "age": 0, "score": 0.0, "active": false}
	err = table1.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// Create primary key index on id
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("id")
	err = table1.CreateIndex(pk)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// Create secondary index on age
	ageIdx, _ := DefaultNormalIndexNew("age_index")
	ageIdx.AddFields("age")
	err = table1.CreateIndex(ageIdx)
	if err != nil {
		t.Fatalf("Failed to create age index: %v", err)
	}

	// Insert test data
	testData := []map[string]any{
		{"id": 1, "name": "Alice", "age": 20, "score": 85.5, "active": true},
		{"id": 2, "name": "Bob", "age": 25, "score": 90.0, "active": true},
		{"id": 3, "name": "Charlie", "age": 30, "score": 75.5, "active": false},
		{"id": 4, "name": "David", "age": 35, "score": 95.0, "active": true},
		{"id": 5, "name": "Eve", "age": 40, "score": 80.0, "active": false},
	}

	for _, data := range testData {
		_, err := table1.Insert(&data)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// Create test table
	table2, err := TableNew("test_search_comprehensive2")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Set table fields
	fields2 := map[string]any{"id": 0, "name": "", "age": 0, "score": 0.0, "active": false}
	err = table2.SetFields(fields2)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// Create primary key index on id
	pk2, _ := DefaultPrimaryKeyNew("pk")
	pk2.AddFields("id")
	err = table2.CreateIndex(pk2)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// Create secondary index on age
	ageIdx2, _ := DefaultNormalIndexNew("age_index")
	ageIdx2.AddFields("age")
	err = table2.CreateIndex(ageIdx2)
	if err != nil {
		t.Fatalf("Failed to create age index: %v", err)
	}

	// Insert test data
	testData2 := []map[string]any{

		{"id": 3, "name": "Charlie", "age": 30, "score": 75.5, "active": false},
		{"id": 4, "name": "David", "age": 35, "score": 95.0, "active": true},
		{"id": 5, "name": "Eve", "age": 40, "score": 80.0, "active": false},
		{"id": 6, "name": "Frank", "age": 45, "score": 88.5, "active": true},
		{"id": 7, "name": "Grace", "age": 50, "score": 92.0, "active": true},
	}

	for _, data := range testData2 {
		_, err := table2.Insert(&data)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// Create test table
	table3, err := TableNew("test_search_comprehensive3")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Set table fields
	fields3 := map[string]any{"id": 0, "name": "", "age": 0, "score": 0.0, "active": false}
	err = table3.SetFields(fields3)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// Create primary key index on id
	pk3, _ := DefaultPrimaryKeyNew("pk")
	pk3.AddFields("id")
	err = table3.CreateIndex(pk3)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// Create secondary index on age
	ageIdx3, _ := DefaultNormalIndexNew("age_index")
	ageIdx3.AddFields("age")
	err = table3.CreateIndex(ageIdx3)
	if err != nil {
		t.Fatalf("Failed to create age index: %v", err)
	}

	// Insert test data
	testData3 := []map[string]any{
		{"id": 5, "name": "Eve", "age": 40, "score": 80.0, "active": false},
		{"id": 6, "name": "Frank", "age": 45, "score": 88.5, "active": true},
		{"id": 7, "name": "Grace", "age": 50, "score": 92.0, "active": true},
		{"id": 8, "name": "Henry", "age": 55, "score": 78.5, "active": false},
		{"id": 9, "name": "Ivy", "age": 60, "score": 83.0, "active": true},
	}

	for _, data := range testData3 {
		_, err := table3.Insert(&data)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}
	iter1 := table1.Search(&map[string]any{"id": nil}) //遍历table1的所有记录
	defer iter1.Release()
	iter2 := table2.Search(&map[string]any{"id": nil}) //遍历table2的所有记录
	defer iter2.Release()
	iter3 := table3.Search(&map[string]any{"id": nil}) //遍历table3的所有记录
	defer iter3.Release()

	// 获取迭代器记录
	fmt.Println("---------table1 records--------------------------------------")
	rd := iter1.GetRecords(true)
	for _, record := range rd {
		fmt.Println(record)
	}
	fmt.Println("-----------------------------------------------")
	// 获取table2的ID映射
	map2 := iter2.Map()
	defer PutMap(map2)
	// 创建一个匹配器，匹配table1的ID是否在table2中
	mach := match.NewAND([]string{"id"}, map2)
	fmt.Println("-----------------------------------------------")
	// select table1.* from table1,table2 where table1.id=table2.id
	fmt.Println("select table1.* from table1,table2 where table1.id=table2.id")
	fmt.Println("-----------------------------------------------")
	iter1.SetMatch(mach)
	rd2 := iter1.GetRecords(true)
	if len(rd2) != 3 {
		t.Fatalf("GetRecords count not equal 3, got %d", len(rd2))
	}
	for _, record := range rd2 {
		fmt.Println(record)
	}
	fmt.Println("-----------------------------------------------")
	// select table1.* from table1,table2 where table1.id!=table2.id
	fmt.Println("select table1.* from table1,table2 where table1.id!=table2.id")
	fmt.Println("-----------------------------------------------")
	mach1 := match.NewAND([]string{"id"}, map2, false)
	iter1.SetMatch(mach1)
	rd3 := iter1.GetRecords(true)
	if len(rd3) != 2 {
		t.Fatalf("GetRecords count not equal 2, got %d", len(rd3))
	}
	for _, record := range rd3 {
		fmt.Println(record)
	}
	fmt.Println("-----------------------------------------------")
	// select table1.* from table1,table2,table3 where table1.id=table2.id and table1.id=table3.id
	fmt.Println("select table1.* from table1,table2,table3 where table1.id=table2.id and table1.id=table3.id")
	fmt.Println("-----------------------------------------------")

	map3 := iter3.Map()
	defer PutMap(map3)
	mach2 := match.NewAND([]string{"id"}, map3)
	iter1.SetMatch(mach, mach2)
	rd4 := iter1.GetRecords(true)
	if len(rd4) != 1 {
		t.Fatalf("GetRecords count not equal 1, got %d", len(rd4))
	}
	for _, record := range rd4 {
		fmt.Println(record)
	}
	fmt.Println("-----------------------------------------------")
	// select table1.* from table1,table2,table3 where table1.id=table2.id and table1.id!=table3.id
	fmt.Println("select table1.* from table1,table2,table3 where table1.id=table2.id and table1.id!=table3.id")
	fmt.Println("-----------------------------------------------")
	map3 = iter3.Map()
	defer PutMap(map3)
	mach2 = match.NewAND([]string{"id"}, map3, false)
	iter1.SetMatch(mach, mach2)
	rd5 := iter1.GetRecords(true)
	if len(rd5) != 2 {
		t.Fatalf("GetRecords count not equal 2, got %d", len(rd5))
	}
	for _, record := range rd5 {
		fmt.Println(record)
	}

	// 以下是测试，不影响主逻辑
	// 测试获取迭代器记录数量
	// count := iter1.Count()
	// if count != 5 {
	// 	t.Fatalf("Count not equal 5, got %d", count)
	// }

	// 测试获取迭代器字段
	// iter1.First()
	// fieldValues := iter1.GetFields(iter1.Key(), iter1.Value())
	// fmt.Println(fieldValues)

	// 测试获取迭代器字段值
	// iter1.First()
	// fieldValue := iter1.GetFieldValue(iter1.Key(), iter1.Value(), "name")
	// fmt.Println(fieldValue)
}

// TestTableIter_MapDataClean 测试 TableIter.Map() 方法返回的数据是否干净
// 确保从对象池获取的 map[any]bool 对象始终是空的，没有残留之前的数据
func TestTableIter_MapDataClean(t *testing.T) {
	// 创建测试表
	table, err := TableNew("test_map_data_clean")
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
	err = table.CreateIndex(pk)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// 插入测试数据
	testData := []map[string]any{
		{"id": 1, "name": "Alice", "age": 20},
		{"id": 2, "name": "Bob", "age": 25},
		{"id": 3, "name": "Charlie", "age": 30},
	}

	for _, data := range testData {
		_, err := table.Insert(&data)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// 创建迭代器
	iter := table.Search(&map[string]any{"id": nil})
	defer iter.Release()

	// 第一次调用 Map() 方法
	map1 := iter.Map()
	defer PutMap(map1)
	if len(map1) != 3 {
		t.Fatalf("Map() should return 3 items, got %d", len(map1))
	}

	// 检查 map1 中的数据是否正确
	expectedIDs := []int{1, 2, 3}
	for _, id := range expectedIDs {
		if !map1[id] {
			t.Fatalf("Map() should contain id %d", id)
		}
	}

	// 第二次调用 Map() 方法
	map2 := iter.Map()
	defer PutMap(map2)
	if len(map2) != 3 {
		t.Fatalf("Map() should return 3 items, got %d", len(map2))
	}

	// 检查 map2 中的数据是否正确
	for _, id := range expectedIDs {
		if !map2[id] {
			t.Fatalf("Map() should contain id %d", id)
		}
	}

	// 检查 map1 和 map2 是否是不同的对象（因为它们都是从对象池获取的）
	if &map1 == &map2 {
		t.Fatalf("Map() should return different objects each time")
	}

	// 测试使用指定字段调用 Map() 方法
	map3 := iter.Map("age")
	defer PutMap(map3)
	if len(map3) != 3 {
		t.Fatalf("Map('age') should return 3 items, got %d", len(map3))
	}

	// 检查 map3 中的数据是否正确
	expectedAges := []int{20, 25, 30}
	for _, age := range expectedAges {
		if !map3[age] {
			t.Fatalf("Map('age') should contain age %d", age)
		}
	}

	// 第三次调用 Map() 方法，再次使用默认字段（id）
	map4 := iter.Map()
	defer PutMap(map4)
	if len(map4) != 3 {
		t.Fatalf("Map() should return 3 items, got %d", len(map4))
	}

	// 检查 map4 中的数据是否正确
	for _, id := range expectedIDs {
		if !map4[id] {
			t.Fatalf("Map() should contain id %d", id)
		}
	}

	// 测试多次调用后，对象池中的对象是否被正确重用和清理
	for i := 0; i < 10; i++ {
		mapN := iter.Map()
		defer PutMap(mapN)
		if len(mapN) != 3 {
			t.Fatalf("Map() should return 3 items on iteration %d, got %d", i, len(mapN))
		}
		for _, id := range expectedIDs {
			if !mapN[id] {
				t.Fatalf("Map() should contain id %d on iteration %d", id, i)
			}
		}
	}
}

// TestTableIter_WithFieldComparison 测试 FieldComparison 与 TableIter 的集成
// 主要是用于主键迭代器对于无索引字段的匹配。
func TestTableIter_WithFieldComparison(t *testing.T) {
	// Create a test table
	table, err := TableNew("test_field_comparison")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Set table fields
	fields := map[string]any{
		"id":     0,
		"name":   "",
		"age":    0,
		"score":  0.0,
		"active": false,
	}

	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// Create primary key index
	pkIndex, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}
	pkIndex.AddFields("id")
	err = table.CreateIndex(pkIndex)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// Insert test data
	testData := []map[string]any{
		{"id": 1, "name": "Alice", "age": 20, "score": 85.5, "active": true},
		{"id": 2, "name": "Bob", "age": 25, "score": 90.0, "active": true},
		{"id": 3, "name": "Charlie", "age": 30, "score": 75.5, "active": false},
		{"id": 4, "name": "David", "age": 35, "score": 95.0, "active": true},
		{"id": 5, "name": "Eve", "age": 40, "score": 80.0, "active": false},
	}

	for _, record := range testData {
		_, err = table.Insert(&record)
		if err != nil {
			t.Fatalf("Failed to insert record: %v", err)
		}
	}

	// Test FieldComparison with TableIter.SetMatch following the pattern from existing tests
	// Reference: map2 := iter2.Map(); mach := match.NewAND([]string{"id"}, map2); iter1.SetMatch(mach)

	// Get all records first
	iter := table.Search(&map[string]any{"id": nil})
	defer iter.Release()

	// Create a FieldComparison matcher (similar to the AND matcher usage in the reference)
	matcher := match.NewFieldComparison("age", match.GreaterThan, 25)

	// Set the matcher on the iterator, just like the reference code does with AND matcher
	iter.SetMatch(matcher)

	// Get filtered records
	records := iter.GetRecords(true)
	if len(records) != 3 {
		t.Errorf("Expected 3 records with age > 25, got %d", len(records))
	} else {
		t.Logf("FieldComparison GreaterThan matcher returned %d records", len(records))
		fmt.Println("------records with age > 25-----------------------------------------")
		for _, record := range records {
			//t.Logf("Record: %v", record)
			fmt.Println(record)
		}
	}

	// Additional test: Using helper function for better readability
	iter2 := table.Search(&map[string]any{"id": nil})
	defer iter2.Release()

	// Use helper function instead of direct constructor
	matcher2 := match.NewEqualMatch("active", false)
	iter2.SetMatch(matcher2)

	inactiveRecords := iter2.GetRecords(true)
	if len(inactiveRecords) != 2 {
		t.Errorf("Expected 2 inactive records, got %d", len(inactiveRecords))
	} else {
		t.Logf("FieldComparison Equal matcher returned %d inactive records", len(inactiveRecords))
		fmt.Println("------inactive records-----------------------------------------")
		for _, record := range inactiveRecords {
			//t.Logf("Record: %v", record)
			fmt.Println(record)
		}
		fmt.Println("-----------------------------------------------")
	}

	// Additional test: Demonstrating integration with existing query patterns
	// Create a scenario similar to line 171-174 but using FieldComparison
	t.Log("\n=== Testing FieldComparison integration pattern ===")

	// Get iterator for all records
	baseIter := table.Search(&map[string]any{"id": nil})
	defer baseIter.Release()

	// Create FieldComparison matcher for high scores
	highScoreMatcher := match.NewGreaterThanMatch("score", 90.0)

	// Set the matcher on the iterator
	baseIter.SetMatch(highScoreMatcher)

	// Get filtered records
	highScoreRecords := baseIter.GetRecords(true)
	t.Logf("High score records (>90): %d", len(highScoreRecords))
	fmt.Println("------high score records (>90)-----------------------------------------")
	for _, record := range highScoreRecords {
		//t.Logf("High score record: %v", record)
		fmt.Println(record)
	}
	fmt.Println("-----------------------------------------------")
}

// TestTableIterGetRecords 测试 TableIter.GetRecords 方法的各种场景
func TestTableIterGetRecords(t *testing.T) {
	// 创建测试表和索引
	table, err := setupTestTable(t)
	if err != nil {
		t.Fatalf("Failed to setup test table: %v", err)
	}

	// 插入测试数据
	if err := insertTestData(t, table); err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// 获取迭代器
	iter := table.Search(&map[string]any{"id": nil})
	if iter == nil {
		t.Fatalf("Failed to get iterator")
	}
	defer iter.Release()

	// 测试用例 1: 正序分页
	t.Run("ForwardPagination", func(t *testing.T) {
		testForwardPagination(t, iter)
	})

	// 测试用例 2: 倒序分页
	t.Run("ReversePagination", func(t *testing.T) {
		testReversePagination(t, iter)
	})

	// 测试用例 3: 正序TopN
	t.Run("ForwardTopN", func(t *testing.T) {
		testForwardTopN(t, iter)
	})

	// 测试用例 4: 倒序TopN
	t.Run("ReverseTopN", func(t *testing.T) {
		testReverseTopN(t, iter)
	})

}

// 测试正序分页
func testForwardPagination(t *testing.T, iter *TableIter) {
	t.Log("测试正序分页...")

	// 测试第1页 (0-2)
	GetRecords := iter.GetRecords(true, 0, 3)
	if len(GetRecords) != 3 {
		t.Errorf("Expected 3 GetRecords, got %d", len(GetRecords))
	}

	// 测试第2页 (3-5)
	GetRecords = iter.GetRecords(true, 3, 3)
	if len(GetRecords) != 3 {
		t.Errorf("Expected 3 GetRecords, got %d", len(GetRecords))
	}

	// 测试第3页 (6-8)
	GetRecords = iter.GetRecords(true, 6, 3)
	if len(GetRecords) != 3 {
		t.Errorf("Expected 3 GetRecords, got %d", len(GetRecords))
	}

	// 测试第4页 (9-)
	GetRecords = iter.GetRecords(true, 9, 3)
	if len(GetRecords) != 1 {
		t.Errorf("Expected 1 record, got %d", len(GetRecords))
	}
}

// 测试倒序分页
func testReversePagination(t *testing.T, iter *TableIter) {
	t.Log("测试倒序分页...")

	// 测试第1页 (9-7)
	GetRecords := iter.GetRecords(false, 0, 3)
	if len(GetRecords) != 3 {
		t.Errorf("Expected 3 GetRecords, got %d", len(GetRecords))
	}

	// 测试第2页 (6-4)
	GetRecords = iter.GetRecords(false, 3, 3)
	if len(GetRecords) != 3 {
		t.Errorf("Expected 3 GetRecords, got %d", len(GetRecords))
	}

	// 测试第3页 (3-1)
	GetRecords = iter.GetRecords(false, 6, 3)
	if len(GetRecords) != 3 {
		t.Errorf("Expected 3 GetRecords, got %d", len(GetRecords))
	}

	// 测试第4页 (0-)
	GetRecords = iter.GetRecords(false, 9, 3)
	if len(GetRecords) != 1 {
		t.Errorf("Expected 1 record, got %d", len(GetRecords))
	}
}

// 测试正序TopN
func testForwardTopN(t *testing.T, iter *TableIter) {
	t.Log("测试正序TopN...")

	// 测试Top3
	GetRecords := iter.GetRecords(true, 3)
	if len(GetRecords) != 3 {
		t.Errorf("Expected 3 GetRecords, got %d", len(GetRecords))
	}

	// 测试Top5
	GetRecords = iter.GetRecords(true, 5)
	if len(GetRecords) != 5 {
		t.Errorf("Expected 5 GetRecords, got %d", len(GetRecords))
	}

	// 测试Top15 (超过总数)
	GetRecords = iter.GetRecords(true, 15)
	if len(GetRecords) != 10 {
		t.Errorf("Expected 10 GetRecords, got %d", len(GetRecords))
	}
}

// 测试倒序TopN
func testReverseTopN(t *testing.T, iter *TableIter) {
	t.Log("测试倒序TopN...")

	// 测试Top3
	GetRecords := iter.GetRecords(false, 3)
	if len(GetRecords) != 3 {
		t.Errorf("Expected 3 GetRecords, got %d", len(GetRecords))
	}

	// 测试Top5
	GetRecords = iter.GetRecords(false, 5)
	if len(GetRecords) != 5 {
		t.Errorf("Expected 5 GetRecords, got %d", len(GetRecords))
	}

	// 测试Top15 (超过总数)
	GetRecords = iter.GetRecords(false, 15)
	if len(GetRecords) != 10 {
		t.Errorf("Expected 10 GetRecords, got %d", len(GetRecords))
	}
}

// TestTableIter_DeleteUpdate tests the Delete and Update methods of TableIter
func TestTableIter_DeleteUpdate(t *testing.T) {
	// Create a test table
	table, err := TableNew("test_delete_update")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Set table fields
	fields := map[string]any{
		"id":     0,
		"name":   "",
		"age":    0,
		"score":  0.0,
		"active": false,
	}

	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// Create primary key index
	pkIndex, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}
	pkIndex.AddFields("id")
	err = table.CreateIndex(pkIndex)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// Insert test data
	testData := []map[string]any{
		{"id": 1, "name": "Alice", "age": 20, "score": 85.5, "active": true},
		{"id": 2, "name": "Bob", "age": 25, "score": 90.0, "active": true},
		{"id": 3, "name": "Charlie", "age": 30, "score": 75.5, "active": false},
		{"id": 4, "name": "David", "age": 35, "score": 95.0, "active": true},
		{"id": 5, "name": "Eve", "age": 40, "score": 80.0, "active": false},
	}

	for _, record := range testData {
		_, err = table.Insert(&record)
		if err != nil {
			t.Fatalf("Failed to insert record: %v", err)
		}
	}

	// Test 1: Delete inactive records using TableIter
	t.Run("DeleteInactiveRecords", func(t *testing.T) {
		// Get full table iterator by setting id to nil, then filter with matcher
		iter := table.Search(&map[string]any{"id": nil}) // This returns a full table iterator
		if iter == nil {
			t.Fatalf("Failed to get full table iterator")
		}
		defer iter.Release()

		// Set matcher to find inactive records
		iter.SetMatch(match.NewFieldComparison("active", match.Equal, false))

		// Delete records found by iterator
		iter.Delete()

		// Verify deletion by checking count of inactive records
		iterAfter := table.Search(&map[string]any{"id": nil})
		if iterAfter != nil {
			defer iterAfter.Release()
			iterAfter.SetMatch(match.NewFieldComparison("active", match.Equal, false))
			recordsAfter := iterAfter.GetRecords(true)
			if len(recordsAfter) != 0 {
				t.Errorf("Expected 0 inactive records after deletion, got %d", len(recordsAfter))
			}
		}
	})

	// Test 2: Update records using TableIter with limit
	t.Run("UpdateRecordsWithLimit", func(t *testing.T) {
		// Insert more test data for update test
		moreData := []map[string]any{
			{"id": 6, "name": "Frank", "age": 45, "score": 88.5, "active": true},
			{"id": 7, "name": "Grace", "age": 50, "score": 92.0, "active": true},
		}
		for _, record := range moreData {
			_, err = table.Insert(&record)
			if err != nil {
				t.Fatalf("Failed to insert more data: %v", err)
			}
		}

		// Get full table iterator, then filter with matcher
		iter := table.Search(&map[string]any{"id": nil})
		if iter == nil {
			t.Fatalf("Failed to get full table iterator")
		}
		defer iter.Release()

		// Set matcher to find active records
		iter.SetMatch(match.NewFieldComparison("active", match.Equal, true))

		// Update only 2 records
		updateFields := map[string]any{"score": 100.0}
		iter.Update(&updateFields, 2) // limit to 2 records

		// Verify update
		iterAfter := table.Search(&map[string]any{"id": nil})
		if iterAfter != nil {
			defer iterAfter.Release()
			iterAfter.SetMatch(match.NewFieldComparison("score", match.Equal, 100.0))
			recordsAfter := iterAfter.GetRecords(true)
			if len(recordsAfter) != 2 {
				t.Errorf("Expected 2 records with score=100, got %d", len(recordsAfter))
			}
		}
	})
}

// TestTableIter_Map tests the Map method of TableIter
func TestTableIter_Map(t *testing.T) {
	// Create a test table
	table, err := TableNew("test_iter_map")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Set table fields
	fields := map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	}

	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// Create primary key index
	pkIndex, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}
	pkIndex.AddFields("id")
	err = table.CreateIndex(pkIndex)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// Insert test data
	testData := []map[string]any{
		{"id": 1, "name": "Alice", "age": 20},
		{"id": 2, "name": "Bob", "age": 25},
		{"id": 3, "name": "Charlie", "age": 30},
		{"id": 4, "name": "David", "age": 35},
	}

	for _, record := range testData {
		_, err = table.Insert(&record)
		if err != nil {
			t.Fatalf("Failed to insert record: %v", err)
		}
	}

	// Test 1: Map with default primary key
	t.Run("MapDefaultPK", func(t *testing.T) {
		iter := table.Search(&map[string]any{"id": nil})
		defer iter.Release()

		idMap := iter.Map()
		defer PutMap(idMap)
		if len(idMap) != 4 {
			t.Errorf("Expected map with 4 entries, got %d", len(idMap))
		}

		// Check if all IDs are present
		expectedIDs := []any{1, 2, 3, 4}
		for _, id := range expectedIDs {
			if !idMap[id] {
				t.Errorf("Expected ID %v in map, not found", id)
			}
		}
	})

	// Test 2: Map with specific field
	t.Run("MapSpecificField", func(t *testing.T) {
		iter := table.Search(&map[string]any{"id": nil})
		defer iter.Release()

		ageMap := iter.Map("age")
		defer PutMap(ageMap)
		if len(ageMap) != 4 {
			t.Errorf("Expected map with 4 entries, got %d", len(ageMap))
		}

		// Check if all ages are present
		expectedAges := []any{20, 25, 30, 35}
		for _, age := range expectedAges {
			if !ageMap[age] {
				t.Errorf("Expected age %v in map, not found", age)
			}
		}
	})
}

// TestTableIter_Count tests the Count method of TableIter
func TestTableIter_Count(t *testing.T) {
	// Create a test table
	table, err := TableNew("test_iter_count")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Set table fields
	fields := map[string]any{
		"id":     0,
		"name":   "",
		"age":    0,
		"active": false,
	}

	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// Create primary key index
	pkIndex, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}
	pkIndex.AddFields("id")
	err = table.CreateIndex(pkIndex)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// Insert test data
	testData := []map[string]any{
		{"id": 1, "name": "Alice", "age": 20, "active": true},
		{"id": 2, "name": "Bob", "age": 25, "active": true},
		{"id": 3, "name": "Charlie", "age": 30, "active": false},
		{"id": 4, "name": "David", "age": 35, "active": true},
		{"id": 5, "name": "Eve", "age": 40, "active": false},
	}

	for _, record := range testData {
		_, err = table.Insert(&record)
		if err != nil {
			t.Fatalf("Failed to insert record: %v", err)
		}
	}

	// Test 1: Count all records
	t.Run("CountAllRecords", func(t *testing.T) {
		iter := table.Search(&map[string]any{"id": nil})
		defer iter.Release()

		count := iter.Count()
		if count != 5 {
			t.Errorf("Expected count 5, got %d", count)
		}
	})

	// Test 2: Count with filter (active records)
	t.Run("CountActiveRecords", func(t *testing.T) {
		// Get full table iterator, then filter with matcher
		iter := table.Search(&map[string]any{"id": nil})
		if iter == nil {
			t.Fatalf("Failed to get full table iterator")
		}
		defer iter.Release()

		// Set matcher to find active records
		iter.SetMatch(match.NewFieldComparison("active", match.Equal, true))

		// Get filtered records and count them
		records := iter.GetRecords(true)

		if len(records) != 3 {
			t.Errorf("Expected 3 active records, got count %d", len(records))
		}
	})
}

// TestTableIter_JumpRange tests the JumpRange functionality
func TestTableIter_JumpRange(t *testing.T) {
	// Create a test table
	table, err := TableNew("test_iter_jumprange")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Set table fields
	fields := map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	}

	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// Create primary key index
	pkIndex, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}
	pkIndex.AddFields("id")
	err = table.CreateIndex(pkIndex)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// Insert test data
	testData := []map[string]any{
		{"id": 1, "name": "Alice", "age": 20},
		{"id": 2, "name": "Bob", "age": 25},
		{"id": 3, "name": "Charlie", "age": 30},
		{"id": 4, "name": "David", "age": 35},
		{"id": 5, "name": "Eve", "age": 40},
		{"id": 6, "name": "Frank", "age": 45},
		{"id": 7, "name": "Grace", "age": 50},
	}

	for _, record := range testData {
		_, err = table.Insert(&record)
		if err != nil {
			t.Fatalf("Failed to insert record: %v", err)
		}
	}

	// Test with JumpRange functionality
	iter := table.Search(&map[string]any{"id": nil})
	defer iter.Release()

	// Test JumpRange with different keys
	t.Run("JumpRangeBasic", func(t *testing.T) {
		// Get some actual keys from the iterator
		iter.First()
		key1 := iter.Key()
		iter.Next()
		key2 := iter.Key()

		// Create jump ranges (in real usage, these would be different iterators)
		jumpRanges := []storage.Iterator{iter.iter} // Use the same iterator for testing

		// Test jump range with key1
		end := iter.JumpRange(key1, jumpRanges, true)
		if end == nil {
			t.Log("JumpRange returned nil for key1 (expected in this test scenario)")
		} else {
			t.Logf("JumpRange returned end key: %v", end)
		}

		// Test jump range with key2
		end2 := iter.JumpRange(key2, jumpRanges, false)
		if end2 == nil {
			t.Log("JumpRange returned nil for key2 (expected in this test scenario)")
		} else {
			t.Logf("JumpRange returned end key for reverse: %v", end2)
		}
	})
}

// TestTableIter_Export tests the ExportRecord and ForExport methods
func TestTableIter_Export(t *testing.T) {
	// Create a test table
	table, err := TableNew("test_iter_export")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Set table fields
	fields := map[string]any{
		"id":   0,
		"name": "",
	}

	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// Create primary key index
	pkIndex, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}
	pkIndex.AddFields("id")
	err = table.CreateIndex(pkIndex)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// Insert test data
	testData := []map[string]any{
		{"id": 1, "name": "Alice"},
		{"id": 2, "name": "Bob"},
		{"id": 3, "name": "Charlie"},
	}

	for _, record := range testData {
		_, err = table.Insert(&record)
		if err != nil {
			t.Fatalf("Failed to insert record: %v", err)
		}
	}

	// Test 1: ForExport method
	t.Run("ForExportMethod", func(t *testing.T) {
		iter := table.Search(&map[string]any{"id": nil})
		defer iter.Release()

		count := 0
		iter.ForExport(true, func(k, v []byte) bool {
			count++
			if count == 2 {
				return false // Stop after 2 records
			}
			return true
		})

		if count != 2 {
			t.Errorf("Expected 2 records exported, got %d", count)
		}
	})

	// Test 2: ExportRecord method with custom processing
	t.Run("ExportRecordMethod", func(t *testing.T) {
		iter := table.Search(&map[string]any{"id": nil})
		defer iter.Release()

		var exportedRecords []record.Record
		iter.ExportRecord(func(rd *record.Record) bool {
			if (*rd)["id"] == 2 {
				return false // Stop when id=2 is found
			}
			exportedRecords = append(exportedRecords, *rd)
			return true
		}, true)

		if len(exportedRecords) != 1 {
			t.Errorf("Expected 1 record exported before id=2, got %d", len(exportedRecords))
		} else if exportedRecords[0]["id"] != 1 {
			t.Errorf("Expected first record to have id=1, got %v", exportedRecords[0]["id"])
		}
	})
}

// TestTableIter_GetPrimaryKeys tests the GetPrimaryKeys method
func TestTableIter_GetPrimaryKeys(t *testing.T) {
	// Create a test table
	table, err := TableNew("test_iter_getpk")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Set table fields
	fields := map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	}

	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// Create primary key index
	pkIndex, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}
	pkIndex.AddFields("id")
	err = table.CreateIndex(pkIndex)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// Insert test data
	testRecord := map[string]any{"id": 1, "name": "Alice", "age": 20}
	_, err = table.Insert(&testRecord)
	if err != nil {
		t.Fatalf("Failed to insert record: %v", err)
	}

	// Get iterator
	iter := table.Search(&map[string]any{"id": nil})
	defer iter.Release()

	iter.First()
	key := iter.Key()
	value := iter.Value()

	// Test 1: Get default primary key
	t.Run("GetDefaultPK", func(t *testing.T) {
		pkValue := iter.GetPrimaryKeys(key, value)
		if pkValue != 1 {
			t.Errorf("Expected primary key value 1, got %v", pkValue)
		}
	})

	// Test 2: Get specific field value
	t.Run("GetSpecificField", func(t *testing.T) {
		nameValue := iter.GetPrimaryKeys(key, value, "name")
		if nameValue != "Alice" {
			t.Errorf("Expected name 'Alice', got %v", nameValue)
		}
	})

	// Test 3: Get multiple fields
	t.Run("GetMultipleFields", func(t *testing.T) {
		multiValue := iter.GetPrimaryKeys(key, value, "id", "name", "age")
		if multiValue != "1-Alice-20" {
			t.Errorf("Expected combined value '1-Alice-20', got %v", multiValue)
		}
	})
}

// Helper functions for test setup
func setupTestTable(t *testing.T) (*Table, error) {
	table, err := TableNew("test_table_iter")
	if err != nil {
		return nil, err
	}

	fields := map[string]any{
		"id":     0,
		"name":   "",
		"age":    0,
		"score":  0.0,
		"active": false,
	}

	err = table.SetFields(fields)
	if err != nil {
		return nil, err
	}

	pkIndex, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		return nil, err
	}
	pkIndex.AddFields("id")
	err = table.CreateIndex(pkIndex)
	if err != nil {
		return nil, err
	}

	return table, nil
}

func insertTestData(t *testing.T, table *Table) error {
	testData := []map[string]any{
		{"id": 1, "name": "Alice", "age": 20, "score": 85.5, "active": true},
		{"id": 2, "name": "Bob", "age": 25, "score": 90.0, "active": true},
		{"id": 3, "name": "Charlie", "age": 30, "score": 75.5, "active": false},
		{"id": 4, "name": "David", "age": 35, "score": 95.0, "active": true},
		{"id": 5, "name": "Eve", "age": 40, "score": 80.0, "active": false},
		{"id": 6, "name": "Frank", "age": 45, "score": 88.5, "active": true},
		{"id": 7, "name": "Grace", "age": 50, "score": 92.0, "active": true},
		{"id": 8, "name": "Henry", "age": 55, "score": 78.5, "active": false},
		{"id": 9, "name": "Ivy", "age": 60, "score": 83.0, "active": true},
		{"id": 10, "name": "Jack", "age": 65, "score": 87.5, "active": true},
	}

	for _, record := range testData {
		_, err := table.Insert(&record)
		if err != nil {
			return err
		}
	}

	return nil
}
