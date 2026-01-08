package engine

import (
	"fmt"
	"testing"

	"github.com/liaoran123/sfsDb/util"
)

// ######所有测试以这个文件基准############

// 测试批量修改，删除
func TestTableIterCRUDS(t *testing.T) {
	table, err := TableNew("test_table_iter")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Set table fields
	fields := map[string]any{"id": 0, "name": "", "age": 0, "score": 0.0, "active": false}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// Create primary key index on id
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("id")
	err = table.CreateIndex(pk)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// Create secondary index on age
	ageIdx, _ := DefaultNormalIndexNew("age_index")
	ageIdx.AddFields("age")
	err = table.CreateIndex(ageIdx)
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
		{"id": 6, "name": "Frank", "age": 45, "score": 88.5, "active": false},
		{"id": 7, "name": "Grace", "age": 50, "score": 92.0, "active": true},
		{"id": 8, "name": "Henry", "age": 55, "score": 78.5, "active": false},
		{"id": 9, "name": "Ivy", "age": 60, "score": 83.0, "active": true},
		{"id": 10, "name": "Jack", "age": 65, "score": 87.5, "active": true},
	}

	for _, data := range testData {
		_, err := table.Insert(&data)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	iter := table.Search(&map[string]any{"age": 60}, util.NotEqual)
	defer iter.Release()
	iter.Update(&map[string]any{"score": 100.0})
	rd := iter.Records(true)
	if len(rd) != 9 {
		t.Fatalf("Records count not equal 9, got %d", len(rd))
	}
	for _, record := range rd {
		fmt.Printf("更新记录：%v\n", record)
	}
	iter.Delete()
	rd = iter.Records(true)
	for _, record := range rd {
		fmt.Printf("更新记录：%v\n", record)
	}

	iter = table.Search(&map[string]any{"id": nil})
	defer iter.Release()
	rd = iter.Records(true)

	for _, record := range rd {
		fmt.Printf("删除后的记录：%v\n", record)
	}
}

// TestTableIterRecordsLimit
func TestTableIterRecordsLimit(t *testing.T) {
	table, err := TableNew("test_search_comprehensive")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Set table fields
	fields := map[string]any{"id": 0, "name": "", "age": 0, "score": 0.0, "active": false}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// Create primary key index on id
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("id")
	err = table.CreateIndex(pk)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// Create secondary index on age
	ageIdx, _ := DefaultNormalIndexNew("age_index")
	ageIdx.AddFields("age")
	err = table.CreateIndex(ageIdx)
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
		{"id": 6, "name": "Frank", "age": 45, "score": 88.5, "active": true},
		{"id": 7, "name": "Grace", "age": 50, "score": 92.0, "active": true},
		{"id": 8, "name": "Henry", "age": 55, "score": 78.5, "active": false},
		{"id": 9, "name": "Ivy", "age": 60, "score": 83.0, "active": true},
		{"id": 10, "name": "Jack", "age": 65, "score": 87.5, "active": true},
	}

	for _, data := range testData {
		_, err := table.Insert(&data)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	//测试分页
	iter := table.Search(&map[string]any{"id": nil})
	defer iter.Release()
	fmt.Println("测试分页：----顺序------------------------")
	for i := range 3 {
		records := iter.Records(true, i*3, 3)
		for _, record := range records {
			fmt.Printf("分页：%v, %v\n", i*3, record)
		}
	}
	fmt.Println("测试分页：----倒序------------------------")
	for i := range 3 {
		records := iter.Records(false, i*3, 3)
		for _, record := range records {
			fmt.Printf("分页：%v, %v\n", i*3, record)
		}
	}
	fmt.Println("Top：----3------------------------")

	records := iter.Records(true, 3)
	for _, record := range records {
		fmt.Printf("Top：%v, %v\n", 3, record)
	}

	fmt.Println("Bot：----3------------------------")

	records = iter.Records(false, 3)
	for _, record := range records {
		fmt.Printf("Bot：%v, %v\n", 3, record)
	}

}

// TestTableIterRecords 测试 TableIter.Records 方法的各种场景
func TestTableIterRecords(t *testing.T) {
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
	records := iter.Records(true, 0, 3)
	if len(records) != 3 {
		t.Errorf("Expected 3 records, got %d", len(records))
	}

	// 测试第2页 (3-5)
	records = iter.Records(true, 3, 3)
	if len(records) != 3 {
		t.Errorf("Expected 3 records, got %d", len(records))
	}

	// 测试第3页 (6-8)
	records = iter.Records(true, 6, 3)
	if len(records) != 3 {
		t.Errorf("Expected 3 records, got %d", len(records))
	}

	// 测试第4页 (9-)
	records = iter.Records(true, 9, 3)
	if len(records) != 1 {
		t.Errorf("Expected 1 record, got %d", len(records))
	}
}

// 测试倒序分页
func testReversePagination(t *testing.T, iter *TableIter) {
	t.Log("测试倒序分页...")

	// 测试第1页 (9-7)
	records := iter.Records(false, 0, 3)
	if len(records) != 3 {
		t.Errorf("Expected 3 records, got %d", len(records))
	}

	// 测试第2页 (6-4)
	records = iter.Records(false, 3, 3)
	if len(records) != 3 {
		t.Errorf("Expected 3 records, got %d", len(records))
	}

	// 测试第3页 (3-1)
	records = iter.Records(false, 6, 3)
	if len(records) != 3 {
		t.Errorf("Expected 3 records, got %d", len(records))
	}

	// 测试第4页 (0-)
	records = iter.Records(false, 9, 3)
	if len(records) != 1 {
		t.Errorf("Expected 1 record, got %d", len(records))
	}
}

// 测试正序TopN
func testForwardTopN(t *testing.T, iter *TableIter) {
	t.Log("测试正序TopN...")

	// 测试Top3
	records := iter.Records(true, 3)
	if len(records) != 3 {
		t.Errorf("Expected 3 records, got %d", len(records))
	}

	// 测试Top5
	records = iter.Records(true, 5)
	if len(records) != 5 {
		t.Errorf("Expected 5 records, got %d", len(records))
	}

	// 测试Top15 (超过总数)
	records = iter.Records(true, 15)
	if len(records) != 10 {
		t.Errorf("Expected 10 records, got %d", len(records))
	}
}

// 测试倒序TopN
func testReverseTopN(t *testing.T, iter *TableIter) {
	t.Log("测试倒序TopN...")

	// 测试Top3
	records := iter.Records(false, 3)
	if len(records) != 3 {
		t.Errorf("Expected 3 records, got %d", len(records))
	}

	// 测试Top5
	records = iter.Records(false, 5)
	if len(records) != 5 {
		t.Errorf("Expected 5 records, got %d", len(records))
	}

	// 测试Top15 (超过总数)
	records = iter.Records(false, 15)
	if len(records) != 10 {
		t.Errorf("Expected 10 records, got %d", len(records))
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
