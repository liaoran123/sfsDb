package engine

import (
	"fmt"
	"testing"

	"github.com/liaoran123/sfsDb/util"
)

// 测试多表组合查询
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
	rd := iter1.GetRecords(true)
	if len(rd) != 5 {
		t.Fatalf("GetRecords count not equal 5, got %d", len(rd))
	}
	iter2 := table2.Search(&map[string]any{"id": nil}) //遍历table2的所有记录
	defer iter2.Release()
	rd2 := iter2.GetRecords(true)
	if len(rd2) != 5 {
		t.Fatalf("GetRecords count not equal 5, got %d", len(rd2))
	}
	iter3 := table3.Search(&map[string]any{"id": nil}) //遍历table3的所有记录
	defer iter3.Release()
	rd3 := iter3.GetRecords(true)
	if len(rd3) != 5 {
		t.Fatalf("GetRecords count not equal 5, got %d", len(rd3))
	}
	fmt.Println("---------------------------------------------")
	// 测试join查询
	// select table1.* from table1,table2 where table1.id=table2.id
	fmt.Println("select table1.* from table1,table2 where table1.id=table2.id")
	fmt.Println("---------------------------------------------")
	map2 := iter2.Map()
	mach := NewAND([]string{"id"}, map2)
	iter1.SetMatch(mach)
	rd4 := iter1.GetRecords(true)
	if len(rd4) != 3 {
		t.Fatalf("GetRecords count not equal 5, got %d", len(rd4))
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
	mach1 := NewAND([]string{"id"}, map2, false)
	iter1.SetMatch(mach1)
	rd5 := iter1.GetRecords(true)
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
	mach2 := NewAND([]string{"id"}, map3)
	iter1.SetMatch(mach, mach2)
	rd6 := iter1.GetRecords(true)
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
	mach3 := NewAND([]string{"id"}, map3, false)
	iter1.SetMatch(mach1, mach3)
	rd7 := iter1.GetRecords(true)
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
	iterid4 := table1.Search(&map[string]any{"id": 4}, util.Equal)
	defer iterid4.Release()
	iterid4.SetMatch(mach) //table2的mach
	rd8 := iterid4.GetRecords(true)
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
	iterid4NotEqual := table1.Search(&map[string]any{"id": 4}, util.NotEqual)
	defer iterid4NotEqual.Release()
	iterid4NotEqual.SetMatch(mach) //table2的mach
	rd9 := iterid4NotEqual.GetRecords(true)
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
	iterid4Less := table1.Search(&map[string]any{"id": 4}, util.LessThan)
	defer iterid4Less.Release()
	iterid4Less.SetMatch(mach) //table2的mach
	rd10 := iterid4Less.GetRecords(true)
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
	iterid4LessOrEqual := table1.Search(&map[string]any{"id": 4}, util.LessThanOrEqual)
	defer iterid4LessOrEqual.Release()
	iterid4LessOrEqual.SetMatch(mach) //table2的mach
	rd11 := iterid4LessOrEqual.GetRecords(true)
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
	iterid4Greater := table1.Search(&map[string]any{"id": 4}, util.GreaterThan)
	defer iterid4Greater.Release()
	iterid4Greater.SetMatch(mach) //table2的mach
	rd12 := iterid4Greater.GetRecords(true)
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
	iterid4GreaterOrEqual := table1.Search(&map[string]any{"id": 4}, util.GreaterThanOrEqual)
	defer iterid4GreaterOrEqual.Release()
	iterid4GreaterOrEqual.SetMatch(mach) //table2的mach
	rd13 := iterid4GreaterOrEqual.GetRecords(true)
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
	iterid4GreaterThan := table1.Search(&map[string]any{"id": 4}, util.GreaterThan)
	defer iterid4GreaterThan.Release()
	iterid4GreaterThan.SetMatch(mach) //table2的mach
	rd14 := iterid4GreaterThan.GetRecords(true)
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
	iterAge40 := table1.Search(&map[string]any{"age": 40}, util.Equal)
	defer iterAge40.Release()
	iterAge40.SetMatch(mach) //table2的mach
	rd15 := iterAge40.GetRecords(true)
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

// 测试批量修改，删除
func TestTableIterCRUDS(t *testing.T) {
	table, err := setupTestTable(t)
	if err != nil {
		t.Fatalf("Failed to setup test table: %v", err)
	}

	// 插入测试数据
	if err := insertTestData(t, table); err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// 测试更新操作
	iter := table.Search(&map[string]any{"age": 60}, util.NotEqual)
	defer iter.Release()
	iter.Update(&map[string]any{"score": 100.0})
	rd := iter.GetRecords(true)
	if len(rd) != 9 {
		t.Fatalf("GetRecords count after update not equal 9, got %d", len(rd))
	}
	// 验证更新是否成功
	for _, record := range rd {
		if record["id"] == 9 {
			// id=9 年龄是60，不应该被更新
			if record["score"] == 100.0 {
				t.Error("Record with id=9 should not have been updated")
			}
		} else {
			// 其他记录应该被更新
			if record["score"] != 100.0 {
				t.Errorf("Record with id=%v should have score 100.0, got %v", record["id"], record["score"])
			}
		}
	}
	// 添加测试修改active字段
	iter = table.Search(&map[string]any{"age": 30}, util.GreaterThanOrEqual)
	defer iter.Release()
	iter.Update(&map[string]any{"active": true})
	rd = iter.GetRecords(true)
	// 验证active字段是否被正确更新
	for _, record := range rd {
		if record["active"] != true {
			t.Errorf("Record with id=%v should have active=true, got %v", record["id"], record["active"])
		}
		fmt.Println(record)
	}
	// 测试删除操作
	deleteIter := table.Search(&map[string]any{"age": 60}, util.NotEqual)
	defer deleteIter.Release()
	deleteIter.Delete()
	rd = deleteIter.GetRecords(true)
	if len(rd) != 0 {
		t.Fatalf("GetRecords count after delete should be 0, got %d", len(rd))
	}

	// 验证删除后的记录
	verifyIter := table.Search(&map[string]any{"id": nil})
	defer verifyIter.Release()
	rd = verifyIter.GetRecords(true)
	// 应该只剩下id=9的记录
	if len(rd) != 1 {
		t.Fatalf("Expected 1 record after delete, got %d", len(rd))
	}
	if rd[0]["id"] != 9 {
		t.Errorf("Expected record with id=9, got id=%v", rd[0]["id"])
	}
}

// TestTableIterGetRecordsLimit 测试分页和TopN功能
func TestTableIterGetRecordsLimit(t *testing.T) {
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
	defer iter.Release()

	// 测试分页功能
	t.Run("Pagination", func(t *testing.T) {
		// 测试正序分页
		t.Run("Forward", func(t *testing.T) {
			for i := 0; i < 3; i++ {
				GetRecords := iter.GetRecords(true, i*3, 3)
				if len(GetRecords) != 3 {
					t.Errorf("Forward pagination page %d: expected 3 GetRecords, got %d", i+1, len(GetRecords))
				}
				// 验证记录ID是否按预期顺序
				for j, record := range GetRecords {
					expectedID := i*3 + j + 1
					if record["id"] != expectedID {
						t.Errorf("Forward pagination page %d, record %d: expected id %d, got %v", i+1, j+1, expectedID, record["id"])
					}
				}
			}
		})

		// 测试倒序分页
		t.Run("Reverse", func(t *testing.T) {
			for i := 0; i < 3; i++ {
				GetRecords := iter.GetRecords(false, i*3, 3)
				if len(GetRecords) != 3 {
					t.Errorf("Reverse pagination page %d: expected 3 GetRecords, got %d", i+1, len(GetRecords))
				}
				// 验证记录ID是否按预期顺序
				for j, record := range GetRecords {
					expectedID := 10 - (i*3 + j)
					if record["id"] != expectedID {
						t.Errorf("Reverse pagination page %d, record %d: expected id %d, got %v", i+1, j+1, expectedID, record["id"])
					}
				}
			}
		})
	})

	// 测试TopN功能
	t.Run("TopN", func(t *testing.T) {
		// 测试正序Top3
		GetRecords := iter.GetRecords(true, 3)
		if len(GetRecords) != 3 {
			t.Errorf("Forward Top3: expected 3 GetRecords, got %d", len(GetRecords))
		}
		// 验证是否获取到前3条记录
		for i, record := range GetRecords {
			if record["id"] != i+1 {
				t.Errorf("Forward Top3, record %d: expected id %d, got %v", i+1, i+1, record["id"])
			}
		}

		// 测试倒序Top3
		GetRecords = iter.GetRecords(false, 3)
		if len(GetRecords) != 3 {
			t.Errorf("Reverse Top3: expected 3 GetRecords, got %d", len(GetRecords))
		}
		// 验证是否获取到最后3条记录
		for i, record := range GetRecords {
			expectedID := 10 - i
			if record["id"] != expectedID {
				t.Errorf("Reverse Top3, record %d: expected id %d, got %v", i+1, expectedID, record["id"])
			}
		}
	})
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

// 设置测试表
func setupTestTable(t *testing.T) (*Table, error) {
	table, err := TableNew("test_table_iter")
	if err != nil {
		return nil, err
	}

	// 设置表字段
	fields := map[string]any{"id": 0, "name": "", "age": 0, "score": 0.0, "active": false}
	if err := table.SetFields(fields); err != nil {
		return nil, err
	}

	// 创建主键索引
	pk, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		return nil, err
	}
	pk.AddFields("id")
	if err := table.CreateIndex(pk); err != nil {
		return nil, err
	}

	// 创建年龄索引
	ageIdx, err := DefaultNormalIndexNew("age_index")
	if err != nil {
		return nil, err
	}
	ageIdx.AddFields("age")
	if err := table.CreateIndex(ageIdx); err != nil {
		return nil, err
	}

	return table, nil
}

// 插入测试数据
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

	for _, data := range testData {
		_, err := table.Insert(&data)
		if err != nil {
			return err
		}
	}

	return nil
}
