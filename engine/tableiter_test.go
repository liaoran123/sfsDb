package engine

import (
	"fmt"
	"testing"

	"github.com/liaoran123/sfsDb/match"
	"github.com/liaoran123/sfsDb/record"
	"github.com/liaoran123/sfsDb/storage"
	"github.com/liaoran123/sfsDb/util"
)

// 测试多表组合查询功能
// 使用索引功能。
// 发现其他测试用例会使用到这个测试函数，例如TestTableTransaction_CommitQuery、TestTableTransaction_CommitQuery_3等，导致数据不正确。
// 因此，需要为每个测试生成唯一的表名，避免测试之间的数据冲突。
func TestTestSelectForJoin(t *testing.T) {
	// 生成唯一的表名，避免测试之间的数据冲突
	table1Name := "test_search_comprehensive1_" + t.Name()
	table2Name := "test_search_comprehensive2_" + t.Name()
	table3Name := "test_search_comprehensive3_" + t.Name()

	// Create test table
	table1, err := TableNew(table1Name)
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
	table2, err := TableNew(table2Name)
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
	table3, err := TableNew(table3Name)
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
	iter1, err := table1.Search(&map[string]any{"id": nil}) //遍历table1的所有记录
	if err != nil {
		t.Fatalf("Search 失败: %v", err)
	}
	if iter1 == nil {
		t.Fatalf("Failed to search records after concurrent transactions")
	}
	if !iter1.Last() {
		t.Errorf("Expected last record to be the last record inserted: %v", err)
	}
	defer GlobalTableIterPool.Put(iter1)
	iter2, err := table2.Search(&map[string]any{"id": nil}) //遍历table2的所有记录
	if err != nil {
		t.Fatalf("Search 失败: %v", err)
	}
	if iter2 == nil {
		t.Fatalf("Failed to search records after concurrent transactions")
	}
	if !iter2.Last() {
		t.Errorf("Expected last record to be the last record inserted: %v", err)
	}
	defer GlobalTableIterPool.Put(iter2)
	iter3, err := table3.Search(&map[string]any{"id": nil}) //遍历table3的所有记录
	if err != nil {
		t.Fatalf("Search 失败: %v", err)
	}
	if iter3 == nil {
		t.Fatalf("Failed to search records after concurrent transactions")
	}
	if !iter3.Last() {
		t.Errorf("Expected last record to be the last record inserted: %v", err)
	}
	defer GlobalTableIterPool.Put(iter3)

	// 获取迭代器记录
	fmt.Println("---------table1 records--------------------------------------")
	rd := iter1.GetRecords(true)
	defer record.PutRecords(rd)
	for _, record := range rd {
		fmt.Println(record)
	}
	fmt.Println("-----------------------------------------------")
	fmt.Println("---------table2 records--------------------------------------")
	rd = iter2.GetRecords(true)
	defer record.PutRecords(rd)

	for _, record := range rd {
		fmt.Println(record)
	}
	fmt.Println("-----------------------------------------------")
	fmt.Println("---------table3 records--------------------------------------")
	rd = iter3.GetRecords(true)
	defer record.PutRecords(rd)

	for _, record := range rd {
		fmt.Println(record)
	}
	fmt.Println("-----------------------------------------------")
	// 获取table2的ID映射
	map2 := iter2.Map()
	defer iter2.ReleaseMap(map2)
	//defer PutMap(map2)
	/*
		如果map2的生命周期小于iter2，则使用iter2.ReleaseMap(map2)
		否则，使用PutMap(map2)
	*/

	// 创建一个匹配器，匹配table1的ID是否在table2中
	mach := match.NewAND([]string{"id"}, map2)
	fmt.Println("-----------------------------------------------")
	// select table1.* from table1,table2 where table1.id=table2.id
	fmt.Println("select table1.* from table1,table2 where table1.id=table2.id")
	fmt.Println("-----------------------------------------------")
	iter1.SetMatch(mach)
	rd2 := iter1.GetRecords(true)
	defer record.PutRecords(rd2)
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
	defer record.PutRecords(rd3)
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
	defer iter3.ReleaseMap(map3)
	//defer PutMap(map3)
	mach2 := match.NewAND([]string{"id"}, map3)
	iter1.SetMatch(mach, mach2)
	rd4 := iter1.GetRecords(true)
	defer record.PutRecords(rd4)
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
	defer iter3.ReleaseMap(map3)
	//defer PutMap(map3)
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

// 测试多表组合查询功能
func TestTestSelectForJoin1(t *testing.T) {
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
	iter1, err := table1.Search(&map[string]any{"id": nil}) //遍历table1的所有记录
	defer iter1.Release()
	rd := iter1.GetRecords(true)
	defer rd.Release()
	if len(rd) != 5 {
		t.Fatalf("GetRecords count not equal 5, got %d", len(rd))
	}
	iter2, err := table2.Search(&map[string]any{"id": nil}) //遍历table2的所有记录
	defer iter2.Release()
	rd2 := iter2.GetRecords(true)
	defer rd2.Release()
	if len(rd2) != 5 {
		t.Fatalf("GetRecords count not equal 5, got %d", len(rd2))
	}
	iter3, err := table3.Search(&map[string]any{"id": nil}) //遍历table3的所有记录
	defer iter3.Release()
	rd3 := iter3.GetRecords(true)
	defer rd3.Release()
	if len(rd3) != 5 {
		t.Fatalf("GetRecords count not equal 5, got %d", len(rd3))
	}
	fmt.Println("---------------------------------------------")
	// 测试join查询
	// select table1.* from table1,table2 where table1.id=table2.id
	fmt.Println("select table1.* from table1,table2 where table1.id=table2.id")
	fmt.Println("---------------------------------------------")
	map2 := iter2.Map()
	defer iter2.ReleaseMap(map2)
	/*
		// 如果map2生命周期大于iter2，则使用
		// defer PutMap(map2)
	*/
	mach := match.NewAND([]string{"id"}, map2)
	iter1.SetMatch(mach)
	rd4 := iter1.GetRecords(true)
	defer rd4.Release()
	if len(rd4) != 3 {
		t.Fatalf("GetRecords count not equal 3, got %d", len(rd4))
	}
	for _, record := range rd4 {
		if record["id"] != 5 && record["id"] != 4 && record["id"] != 3 {
			t.Errorf("Record with id=%v should have been joined, got %v", record["id"], record)
		}
		fmt.Println(record)
	}
	fmt.Println("---------------------------------------------")
	// select table1.* from table1,table2 where table1.id!=table2.id
	fmt.Println("select table1.* from table1,table2 where table1.id!=table2.id")
	fmt.Println("---------------------------------------------")
	mach1 := match.NewAND([]string{"id"}, map2, false)
	iter1.SetMatch(mach1)
	rd5 := iter1.GetRecords(true)
	defer rd5.Release()
	if len(rd5) != 2 {
		t.Fatalf("GetRecords count not equal 2, got %d", len(rd5))
	}
	for _, record := range rd5 {
		if record["id"] != 1 && record["id"] != 2 {
			t.Errorf("Record with id=%v should have been joined, got %v", record["id"], record)
		}
		fmt.Println(record)
	}
	fmt.Println("---------------------------------------------")
	// select table1.* from table1,table2,table3 where table1.id=table2.id and table1.id=table3.id
	fmt.Println("select table1.* from table1,table2,table3 where table1.id=table2.id and table1.id=table3.id")
	fmt.Println("---------------------------------------------")
	map3 := iter3.Map()
	mach2 := match.NewAND([]string{"id"}, map3)
	iter1.SetMatch(mach, mach2)
	rd6 := iter1.GetRecords(true)
	defer rd6.Release()
	if len(rd6) != 1 {
		t.Fatalf("GetRecords count not equal 2, got %d", len(rd6))
	}
	for _, record := range rd6 {
		if record["id"] != 5 {
			t.Errorf("Record with id=%v should have been joined, got %v", record["id"], record)
		}
		fmt.Println(record)
	}
	fmt.Println("---------------------------------------------")
	// select table1.* from table1,table2,table3 where table1.id!=table2.id and table1.id!=table3.id
	fmt.Println("select table1.* from table1,table2,table3 where table1.id!=table2.id and table1.id!=table3.id")
	fmt.Println("---------------------------------------------")
	mach3 := match.NewAND([]string{"id"}, map3, false)
	iter1.SetMatch(mach1, mach3)
	rd7 := iter1.GetRecords(true)
	defer rd7.Release()
	if len(rd7) != 2 {
		t.Fatalf("GetRecords count not equal 4, got %d", len(rd7))
	}
	for _, record := range rd7 {
		if record["id"] != 1 && record["id"] != 2 {
			t.Errorf("Record with id=%v should have been joined, got %v", record["id"], record)
		}
		fmt.Println(record)
	}
	// select table1.* from table1,table2 where table1.id=4 and table1.id=table2.id
	fmt.Println("select table1.* from table1,table2 where table1.id=4 and table1.id=table2.id")
	fmt.Println("---------------------------------------------")
	iterid4, _ := table1.Search(&map[string]any{"id": 4}, util.Equal)
	defer iterid4.Release()
	iterid4.SetMatch(mach) //table2的mach
	rd8 := iterid4.GetRecords(true)
	defer rd8.Release()
	if len(rd8) != 1 {
		t.Fatalf("GetRecords count not equal 1, got %d", len(rd8))
	}
	for _, record := range rd8 {
		if record["id"] != 4 {
			t.Errorf("Record with id=%v should have been joined, got %v", record["id"], record)
		}
		fmt.Println(record)
	}

	// select table1.* from table1,table2 where table1.id!=4 and table1.id=table2.id
	fmt.Println("select table1.* from table1,table2 where table1.id!=4 and table1.id=table2.id")
	fmt.Println("---------------------------------------------")
	iterid4NotEqual, _ := table1.Search(&map[string]any{"id": 4}, util.NotEqual)
	defer iterid4NotEqual.Release()
	iterid4NotEqual.SetMatch(mach) //table2的mach
	rd9 := iterid4NotEqual.GetRecords(true)
	defer rd9.Release()
	if len(rd9) != 2 {
		t.Fatalf("GetRecords count not equal 4, got %d", len(rd9))
	}
	for _, record := range rd9 {
		if record["id"] != 3 && record["id"] != 5 {
			t.Errorf("Record with id=%v should have been joined, got %v", record["id"], record)
		}
		fmt.Println(record)
	}
	// select table1.* from table1,table2 where table1.id<4 and table1.id=table2.id
	fmt.Println("select table1.* from table1,table2 where table1.id<4 and table1.id=table2.id")
	fmt.Println("---------------------------------------------")
	iterid4Less, _ := table1.Search(&map[string]any{"id": 4}, util.LessThan)
	defer iterid4Less.Release()
	iterid4Less.SetMatch(mach) //table2的mach
	rd10 := iterid4Less.GetRecords(true)
	defer rd10.Release()
	if len(rd10) != 1 {
		t.Fatalf("GetRecords count not equal 3, got %d", len(rd10))
	}
	for _, record := range rd10 {
		if record["id"] != 3 {
			t.Errorf("Record with id=%v should have been joined, got %v", record["id"], record)
		}
		fmt.Println(record)
	}
	// select table1.* from table1,table2 where table1.id<=4 and table1.id=table2.id
	fmt.Println("select table1.* from table1,table2 where table1.id<=4 and table1.id=table2.id")
	fmt.Println("---------------------------------------------")
	iterid4LessOrEqual, _ := table1.Search(&map[string]any{"id": 4}, util.LessThanOrEqual)
	defer iterid4LessOrEqual.Release()
	iterid4LessOrEqual.SetMatch(mach) //table2的mach
	rd11 := iterid4LessOrEqual.GetRecords(true)
	defer rd11.Release()
	if len(rd11) != 2 {
		t.Fatalf("GetRecords count not equal 2, got %d", len(rd11))
	}
	for _, record := range rd11 {
		if record["id"] != 3 && record["id"] != 4 {
			t.Errorf("Record with id=%v should have been joined, got %v", record["id"], record)
		}
		fmt.Println(record)
	}
	// select table1.* from table1,table2 where table1.id>4 and table1.id=table2.id
	fmt.Println("select table1.* from table1,table2 where table1.id>4 and table1.id=table2.id")
	fmt.Println("---------------------------------------------")
	iterid4Greater, _ := table1.Search(&map[string]any{"id": 4}, util.GreaterThan)
	defer iterid4Greater.Release()
	iterid4Greater.SetMatch(mach) //table2的mach
	rd12 := iterid4Greater.GetRecords(true)
	defer rd12.Release()
	if len(rd12) != 1 {
		t.Fatalf("GetRecords count not equal 1, got %d", len(rd12))
	}
	for _, record := range rd12 {
		if record["id"] != 5 {
			t.Errorf("Record with id=%v should have been joined, got %v", record["id"], record)
		}
		fmt.Println(record)
	}
	// select table1.* from table1,table2 where table1.id>=4 and table1.id=table2.id
	fmt.Println("select table1.* from table1,table2 where table1.id>=4 and table1.id=table2.id")
	fmt.Println("---------------------------------------------")
	iterid4GreaterOrEqual, _ := table1.Search(&map[string]any{"id": 4}, util.GreaterThanOrEqual)
	defer iterid4GreaterOrEqual.Release()
	iterid4GreaterOrEqual.SetMatch(mach) //table2的mach
	rd13 := iterid4GreaterOrEqual.GetRecords(true)
	defer rd13.Release()
	if len(rd13) != 2 {
		t.Fatalf("GetRecords count not equal 2, got %d", len(rd13))
	}
	for _, record := range rd13 {
		if record["id"] != 4 && record["id"] != 5 {
			t.Errorf("Record with id=%v should have been joined, got %v", record["id"], record)
		}
		fmt.Println(record)
	}
	// select table1.* from table1,table2 where table1.id>4 and table1.id=table2.id
	fmt.Println("select table1.* from table1,table2 where table1.id>4 and table1.id=table2.id")
	fmt.Println("---------------------------------------------")
	iterid4GreaterThan, _ := table1.Search(&map[string]any{"id": 4}, util.GreaterThan)
	defer iterid4GreaterThan.Release()
	iterid4GreaterThan.SetMatch(mach) //table2的mach
	rd14 := iterid4GreaterThan.GetRecords(true)
	defer rd14.Release()
	if len(rd14) != 1 {
		t.Fatalf("GetRecords count not equal 1, got %d", len(rd14))
	}
	for _, record := range rd14 {
		if record["id"] != 5 {
			t.Errorf("Record with id=%v should have been joined, got %v", record["id"], record)
		}
		fmt.Println(record)
	}

	// select table1.* from table1,table2 where table1.id>4 and table1.id=table2.id
	fmt.Println("select table1.* from table1,table2 where table1.age=40 and table1.id=table2.id")
	fmt.Println("---------------------------------------------")
	iterAge40, _ := table1.Search(&map[string]any{"age": 40}, util.Equal)
	defer iterAge40.Release()
	iterAge40.SetMatch(mach) //table2的mach
	rd15 := iterAge40.GetRecords(true)
	defer rd15.Release()
	if len(rd15) != 1 {
		t.Fatalf("GetRecords count not equal 1, got %d", len(rd15))
	}
	for _, record := range rd15 {
		if record["id"] != 5 {
			t.Errorf("Record with id=%v should have been joined, got %v", record["id"], record)
		}
		fmt.Println(record)
	}
	fmt.Println("---------------------------------------------")
	fmt.Println("----Search函数不支持无索引的搜索，如需要支持无索引或自己的匹配策略，可以自定义mach接口实现-----")
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
	iter, err := table.Search(&map[string]any{"id": nil})
	if err != nil {
		t.Fatalf("Search 失败: %v", err)
	}
	defer GlobalTableIterPool.Put(iter)

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
	iter, err := table.Search(&map[string]any{"id": nil})
	if err != nil {
		t.Fatalf("Search 失败: %v", err)
	}
	defer GlobalTableIterPool.Put(iter)

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
	iter2, err := table.Search(&map[string]any{"id": nil})
	if err != nil {
		t.Fatalf("Search 失败: %v", err)
	}
	defer GlobalTableIterPool.Put(iter2)

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
	baseIter, err := table.Search(&map[string]any{"id": nil})
	if err != nil {
		t.Fatalf("Search 失败: %v", err)
	}
	defer GlobalTableIterPool.Put(baseIter)

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
	iter, err := table.Search(&map[string]any{"id": nil})
	if err != nil {
		t.Fatalf("Search 失败: %v", err)
	}
	if iter == nil {
		t.Fatalf("Failed to get iterator: %v", err)
	}
	defer GlobalTableIterPool.Put(iter)

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
		iter, err := table.Search(&map[string]any{"id": nil})
		if err != nil {
			t.Fatalf("Failed to get full table iterator: %v", err)
		}
		defer GlobalTableIterPool.Put(iter)

		// Set matcher to find inactive records
		iter.SetMatch(match.NewFieldComparison("active", match.Equal, false))

		// Delete records found by iterator
		iter.Delete()

		// Verify deletion by checking count of inactive records
		iterAfter, err := table.Search(&map[string]any{"id": nil})
		if err != nil {
			defer GlobalTableIterPool.Put(iterAfter)
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
		iter, err := table.Search(&map[string]any{"id": nil})
		if err != nil {
			t.Fatalf("Failed to get full table iterator: %v", err)
		}
		defer GlobalTableIterPool.Put(iter)

		iter.SetMatch(match.NewFieldComparison("active", match.Equal, true))

		// Update only 2 records
		updateFields := map[string]any{"score": 100.0}
		iter.Update(&updateFields, 2) // limit to 2 records

		// Verify update
		iterAfter, err := table.Search(&map[string]any{"id": nil})
		if err != nil {
			t.Fatalf("Search 失败: %v", err)
		}
		defer GlobalTableIterPool.Put(iterAfter)
		iterAfter.SetMatch(match.NewFieldComparison("score", match.Equal, 100.0))
		recordsAfter := iterAfter.GetRecords(true)
		if len(recordsAfter) != 2 {
			t.Errorf("Expected 2 records with score=100, got %d", len(recordsAfter))
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
		iter, err := table.Search(&map[string]any{"id": nil})
		if err != nil {
			t.Fatalf("Search 失败: %v", err)
		}
		if iter == nil {
			t.Fatalf("Failed to get iterator: %v", err)
		}
		defer GlobalTableIterPool.Put(iter)

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
		iter, err := table.Search(&map[string]any{"id": nil})
		if err != nil {
			t.Fatalf("Search 失败: %v", err)
		}
		if iter == nil {
			t.Fatalf("Failed to get iterator: %v", err)
		}
		defer GlobalTableIterPool.Put(iter)

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
		iter, err := table.Search(&map[string]any{"id": nil})
		if err != nil {
			t.Fatalf("Search 失败: %v", err)
		}
		defer GlobalTableIterPool.Put(iter)

		count := iter.Count()
		if count != 5 {
			t.Errorf("Expected count 5, got %d", count)
		}
	})

	// Test 2: Count with filter (active records)
	t.Run("CountActiveRecords", func(t *testing.T) {
		// Get full table iterator, then filter with matcher
		iter, err := table.Search(&map[string]any{"id": nil})
		if err != nil {
			t.Fatalf("Failed to get full table iterator")
		}
		defer GlobalTableIterPool.Put(iter)

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
	iter, err := table.Search(&map[string]any{"id": nil})
	if err != nil {
		t.Fatalf("Search 失败: %v", err)
	}
	defer GlobalTableIterPool.Put(iter)

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
		iter, err := table.Search(&map[string]any{"id": nil})
		if err != nil {
			t.Fatalf("Search 失败: %v", err)
		}
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
		iter, _ := table.Search(&map[string]any{"id": nil})
		defer GlobalTableIterPool.Put(iter)

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
	iter, _ := table.Search(&map[string]any{"id": nil})
	defer GlobalTableIterPool.Put(iter)

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

// TestTableIter_UpdatePagination 测试 TableIter.Update 方法的分页功能
func TestTableIter_UpdatePagination(t *testing.T) {
	// Create a test table
	table, err := TableNew("test_update_pagination")
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
		{"id": 6, "name": "Frank", "age": 45, "score": 88.5, "active": true},
		{"id": 7, "name": "Grace", "age": 50, "score": 92.0, "active": true},
		{"id": 8, "name": "Henry", "age": 55, "score": 78.5, "active": false},
		{"id": 9, "name": "Ivy", "age": 60, "score": 83.0, "active": true},
		{"id": 10, "name": "Jack", "age": 65, "score": 87.5, "active": true},
	}

	for _, record := range testData {
		_, err = table.Insert(&record)
		if err != nil {
			t.Fatalf("Failed to insert record: %v", err)
		}
	}

	// Test 1: Update with limit (pagination)
	t.Run("UpdateWithLimit", func(t *testing.T) {
		// Get full table iterator
		iter, err := table.Search(&map[string]any{"id": nil})
		if err != nil {
			t.Fatalf("Failed to get full table iterator: %v", err)
		}
		defer GlobalTableIterPool.Put(iter)

		// Set matcher to find active records
		iter.SetMatch(match.NewFieldComparison("active", match.Equal, true))

		// Update only 3 records
		updateFields := map[string]any{"score": 100.0}
		err = iter.Update(&updateFields, 3) // limit to 3 records
		if err != nil {
			t.Fatalf("Failed to update records: %v", err)
		}

		// Verify update
		iterAfter, err := table.Search(&map[string]any{"id": nil})
		if err != nil {
			t.Fatalf("Search 失败: %v", err)
		}
		defer GlobalTableIterPool.Put(iterAfter)
		iterAfter.SetMatch(match.NewFieldComparison("score", match.Equal, 100.0))
		recordsAfter := iterAfter.GetRecords(true)
		if len(recordsAfter) != 3 {
			t.Errorf("Expected 3 records with score=100, got %d", len(recordsAfter))
		}
	})

	// Test 2: Update with no limit (update all)
	t.Run("UpdateAll", func(t *testing.T) {
		// Get full table iterator
		iter, err := table.Search(&map[string]any{"id": nil})
		if err != nil {
			t.Fatalf("Failed to get full table iterator: %v", err)
		}
		defer GlobalTableIterPool.Put(iter)

		// Set matcher to find inactive records
		iter.SetMatch(match.NewFieldComparison("active", match.Equal, false))

		// Update all inactive records (no limit)
		updateFields := map[string]any{"score": 70.0}
		err = iter.Update(&updateFields) // no limit
		if err != nil {
			t.Fatalf("Failed to update records: %v", err)
		}

		// Verify update
		iterAfter, err := table.Search(&map[string]any{"id": nil})
		if err != nil {
			t.Fatalf("Search 失败: %v", err)
		}
		defer GlobalTableIterPool.Put(iterAfter)
		iterAfter.SetMatch(match.NewFieldComparison("score", match.Equal, 70.0))
		recordsAfter := iterAfter.GetRecords(true)
		if len(recordsAfter) != 3 {
			t.Errorf("Expected 3 records with score=70, got %d", len(recordsAfter))
		}
	})

	// Test 3: Update with limit greater than total records
	t.Run("UpdateWithLargeLimit", func(t *testing.T) {
		// Get full table iterator
		iter, err := table.Search(&map[string]any{"id": nil})
		if err != nil {
			t.Fatalf("Failed to get full table iterator: %v", err)
		}
		defer GlobalTableIterPool.Put(iter)

		// Update with limit greater than total records
		updateFields := map[string]any{"age": 99}
		err = iter.Update(&updateFields, 20) // limit larger than total records
		if err != nil {
			t.Fatalf("Failed to update records: %v", err)
		}

		// Verify all records were updated
		iterAfter, err := table.Search(&map[string]any{"id": nil})
		if err != nil {
			t.Fatalf("Search 失败: %v", err)
		}
		defer GlobalTableIterPool.Put(iterAfter)
		recordsAfter := iterAfter.GetRecords(true)
		for _, record := range recordsAfter {
			if record["age"] != 99 {
				t.Errorf("Expected age=99 for record %v, got %v", record["id"], record["age"])
			}
		}
	})

	// Test 4: Update with pagination (offset and limit)
	t.Run("UpdateWithPagination", func(t *testing.T) {
		// Get full table iterator
		iter, err := table.Search(&map[string]any{"id": nil})
		if err != nil {
			t.Fatalf("Failed to get full table iterator: %v", err)
		}
		defer GlobalTableIterPool.Put(iter)

		// Update with pagination: skip 2 records, update 3 records
		updateFields := map[string]any{"score": 110.0}
		err = iter.Update(&updateFields, 2, 3) // offset=2, limit=3
		if err != nil {
			t.Fatalf("Failed to update records with pagination: %v", err)
		}

		// Verify update
		iterAfter, err := table.Search(&map[string]any{"id": nil})
		if err != nil {
			t.Fatalf("Search 失败: %v", err)
		}
		defer GlobalTableIterPool.Put(iterAfter)
		recordsAfter := iterAfter.GetRecords(true)
		if len(recordsAfter) != 10 {
			t.Errorf("Expected 10 records, got %d", len(recordsAfter))
		}

		// Check that only 3 records were updated
		updatedCount := 0
		for _, record := range recordsAfter {
			if record["score"] == 110.0 {
				updatedCount++
			}
		}
		if updatedCount != 3 {
			t.Errorf("Expected 3 records with score=110, got %d", updatedCount)
		}
	})

	// Test 5: Update with pagination - second page
	t.Run("UpdateWithPaginationSecondPage", func(t *testing.T) {
		// Get full table iterator
		iter, err := table.Search(&map[string]any{"id": nil})
		if err != nil {
			t.Fatalf("Failed to get full table iterator: %v", err)
		}
		defer GlobalTableIterPool.Put(iter)

		// Update with pagination: skip 5 records, update 3 records
		updateFields := map[string]any{"score": 120.0}
		err = iter.Update(&updateFields, 5, 3) // offset=5, limit=3
		if err != nil {
			t.Fatalf("Failed to update records with pagination: %v", err)
		}

		// Verify update
		iterAfter, err := table.Search(&map[string]any{"id": nil})
		if err != nil {
			t.Fatalf("Search 失败: %v", err)
		}
		defer GlobalTableIterPool.Put(iterAfter)
		recordsAfter := iterAfter.GetRecords(true)

		// Check that only 3 records were updated
		updatedCount := 0
		for _, record := range recordsAfter {
			if record["score"] == 120.0 {
				updatedCount++
			}
		}
		if updatedCount != 3 {
			t.Errorf("Expected 3 records with score=120, got %d", updatedCount)
		}
	})
}

// TestTableIter_DeletePagination 测试 TableIter.Delete 方法的分页功能
func TestTableIter_DeletePagination(t *testing.T) {
	// Create a test table
	table, err := TableNew("test_delete_pagination")
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
		{"id": 6, "name": "Frank", "age": 45, "score": 88.5, "active": true},
		{"id": 7, "name": "Grace", "age": 50, "score": 92.0, "active": true},
		{"id": 8, "name": "Henry", "age": 55, "score": 78.5, "active": false},
		{"id": 9, "name": "Ivy", "age": 60, "score": 83.0, "active": true},
		{"id": 10, "name": "Jack", "age": 65, "score": 87.5, "active": true},
	}

	for _, record := range testData {
		_, err = table.Insert(&record)
		if err != nil {
			t.Fatalf("Failed to insert record: %v", err)
		}
	}

	// Test 1: Delete with limit (pagination)
	t.Run("DeleteWithLimit", func(t *testing.T) {
		// Get full table iterator
		iter, err := table.Search(&map[string]any{"id": nil})
		if err != nil {
			t.Fatalf("Failed to get full table iterator: %v", err)
		}
		defer GlobalTableIterPool.Put(iter)

		// Set matcher to find active records
		iter.SetMatch(match.NewFieldComparison("active", match.Equal, true))

		// Delete only 3 records
		err = iter.Delete(3) // limit to 3 records
		if err != nil {
			t.Fatalf("Failed to delete records: %v", err)
		}

		// Verify deletion
		iterAfter, err := table.Search(&map[string]any{"id": nil})
		if err != nil {
			t.Fatalf("Search 失败: %v", err)
		}
		defer GlobalTableIterPool.Put(iterAfter)
		// Count all remaining records
		recordsAfter := iterAfter.GetRecords(true)
		if len(recordsAfter) != 7 {
			t.Errorf("Expected 7 records after deletion, got %d", len(recordsAfter))
		}
	})

	// Test 2: Delete with no limit (delete all)
	t.Run("DeleteAll", func(t *testing.T) {
		// Recreate test table for fresh test
		table2, err := TableNew("test_delete_all")
		if err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}

		err = table2.SetFields(fields)
		if err != nil {
			t.Fatalf("Failed to set fields: %v", err)
		}

		pkIndex2, _ := DefaultPrimaryKeyNew("pk")
		pkIndex2.AddFields("id")
		err = table2.CreateIndex(pkIndex2)
		if err != nil {
			t.Fatalf("Failed to create primary key index: %v", err)
		}

		// Insert test data
		for _, record := range testData {
			_, err = table2.Insert(&record)
			if err != nil {
				t.Fatalf("Failed to insert record: %v", err)
			}
		}

		// Get full table iterator
		iter, err := table2.Search(&map[string]any{"id": nil})
		if err != nil {
			t.Fatalf("Failed to get full table iterator: %v", err)
		}
		defer GlobalTableIterPool.Put(iter)

		// Set matcher to find inactive records
		iter.SetMatch(match.NewFieldComparison("active", match.Equal, false))

		// Delete all inactive records (no limit)
		err = iter.Delete() // no limit
		if err != nil {
			t.Fatalf("Failed to delete records: %v", err)
		}

		// Verify deletion
		iterAfter, err := table2.Search(&map[string]any{"id": nil})
		if err != nil {
			t.Fatalf("Search 失败: %v", err)
		}
		defer GlobalTableIterPool.Put(iterAfter)
		// Count all remaining records
		recordsAfter := iterAfter.GetRecords(true)
		if len(recordsAfter) != 7 {
			t.Errorf("Expected 7 records after deletion, got %d", len(recordsAfter))
		}
	})

	// Test 3: Delete with limit greater than total records
	t.Run("DeleteWithLargeLimit", func(t *testing.T) {
		// Recreate test table for fresh test
		table3, err := TableNew("test_delete_large_limit")
		if err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}

		err = table3.SetFields(fields)
		if err != nil {
			t.Fatalf("Failed to set fields: %v", err)
		}

		pkIndex3, _ := DefaultPrimaryKeyNew("pk")
		pkIndex3.AddFields("id")
		err = table3.CreateIndex(pkIndex3)
		if err != nil {
			t.Fatalf("Failed to create primary key index: %v", err)
		}

		// Insert test data
		for _, record := range testData {
			_, err = table3.Insert(&record)
			if err != nil {
				t.Fatalf("Failed to insert record: %v", err)
			}
		}

		// Get full table iterator
		iter, err := table3.Search(&map[string]any{"id": nil})
		if err != nil {
			t.Fatalf("Failed to get full table iterator: %v", err)
		}
		defer GlobalTableIterPool.Put(iter)

		// Delete with limit greater than total records
		err = iter.Delete(20) // limit larger than total records
		if err != nil {
			t.Fatalf("Failed to delete records: %v", err)
		}

		// Verify all records were deleted
		iterAfter, err := table3.Search(&map[string]any{"id": nil})
		if err != nil {
			t.Fatalf("Search 失败: %v", err)
		}
		defer GlobalTableIterPool.Put(iterAfter)
		recordsAfter := iterAfter.GetRecords(true)
		if len(recordsAfter) != 0 {
			t.Errorf("Expected 0 records after deletion, got %d", len(recordsAfter))
		}
	})

	// Test 4: Delete with pagination (offset and limit)
	t.Run("DeleteWithPagination", func(t *testing.T) {
		// Recreate test table for fresh test
		table4, err := TableNew("test_delete_pagination")
		if err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}

		err = table4.SetFields(fields)
		if err != nil {
			t.Fatalf("Failed to set fields: %v", err)
		}

		pkIndex4, _ := DefaultPrimaryKeyNew("pk")
		pkIndex4.AddFields("id")
		err = table4.CreateIndex(pkIndex4)
		if err != nil {
			t.Fatalf("Failed to create primary key index: %v", err)
		}

		// Insert test data
		for _, record := range testData {
			_, err = table4.Insert(&record)
			if err != nil {
				t.Fatalf("Failed to insert record: %v", err)
			}
		}

		// Get full table iterator
		iter, err := table4.Search(&map[string]any{"id": nil})
		if err != nil {
			t.Fatalf("Failed to get full table iterator: %v", err)
		}
		defer GlobalTableIterPool.Put(iter)

		// Delete with pagination: skip 2 records, delete 3 records
		err = iter.Delete(2, 3) // offset=2, limit=3
		if err != nil {
			t.Fatalf("Failed to delete records with pagination: %v", err)
		}

		// Verify deletion
		iterAfter, err := table4.Search(&map[string]any{"id": nil})
		if err != nil {
			t.Fatalf("Search 失败: %v", err)
		}
		defer GlobalTableIterPool.Put(iterAfter)
		recordsAfter := iterAfter.GetRecords(true)
		if len(recordsAfter) != 7 {
			t.Errorf("Expected 7 records after deletion, got %d", len(recordsAfter))
		}
	})

	// Test 5: Delete with pagination - second page
	t.Run("DeleteWithPaginationSecondPage", func(t *testing.T) {
		// Recreate test table for fresh test
		table5, err := TableNew("test_delete_pagination_second")
		if err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}

		err = table5.SetFields(fields)
		if err != nil {
			t.Fatalf("Failed to set fields: %v", err)
		}

		pkIndex5, _ := DefaultPrimaryKeyNew("pk")
		pkIndex5.AddFields("id")
		err = table5.CreateIndex(pkIndex5)
		if err != nil {
			t.Fatalf("Failed to create primary key index: %v", err)
		}

		// Insert test data
		for _, record := range testData {
			_, err = table5.Insert(&record)
			if err != nil {
				t.Fatalf("Failed to insert record: %v", err)
			}
		}

		// Get full table iterator
		iter, err := table5.Search(&map[string]any{"id": nil})
		if err != nil {
			t.Fatalf("Failed to get full table iterator: %v", err)
		}
		defer GlobalTableIterPool.Put(iter)

		// Delete with pagination: skip 5 records, delete 3 records
		err = iter.Delete(5, 3) // offset=5, limit=3
		if err != nil {
			t.Fatalf("Failed to delete records with pagination: %v", err)
		}

		// Verify deletion
		iterAfter, err := table5.Search(&map[string]any{"id": nil})
		if err != nil {
			t.Fatalf("Search 失败: %v", err)
		}
		defer GlobalTableIterPool.Put(iterAfter)
		recordsAfter := iterAfter.GetRecords(true)
		if len(recordsAfter) != 7 {
			t.Errorf("Expected 7 records after deletion, got %d", len(recordsAfter))
		}
	})
}
