package engine

import (
	"fmt"
	"testing"
	"time"
)

// TestTableCounterContinuity 测试表重新打开后计数器的连续性
func TestTableCounterContinuity(t *testing.T) {
	// 创建唯一表名，避免测试冲突
	tableName := fmt.Sprintf("test_counter_continuity_%d", time.Now().UnixNano())

	// 步骤1: 创建表并插入记录
	table, err := TableNew(tableName)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段和主键
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
	pkIndex, err := DefaultPrimaryKeyNew("pk_id")
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}
	pkIndex.AddFields("id")
	err = table.CreateIndex(pkIndex)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// 插入测试数据
	insertedIDs := make([]int, 0, 3)
	for i := 0; i < 3; i++ {
		record := map[string]any{
			"name": fmt.Sprintf("User%d", i),
			"age":  20 + i,
		}

		id, err := table.Insert(&record)
		if err != nil {
			t.Fatalf("Failed to insert record %d: %v", i, err)
		}
		insertedIDs = append(insertedIDs, id)
		t.Logf("Inserted record with ID: %d", id)
	}

	// 验证插入的ID是否连续
	expectedID := 1
	for _, id := range insertedIDs {
		if id != expectedID {
			t.Errorf("Expected ID %d, got %d", expectedID, id)
		}
		expectedID++
	}

	// 步骤2: 重新打开表
	// 这里我们通过创建一个新的Table实例来模拟重新打开
	// 在实际应用中，这相当于关闭连接后重新连接
	table2, err := TableNew(tableName)
	if err != nil {
		t.Fatalf("Failed to reopen table: %v", err)
	}

	// 重新设置字段和索引信息，因为TableNew不会自动加载
	err = table2.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields after reopen: %v", err)
	}

	// 重新创建主键索引
	pkIndex2, err := DefaultPrimaryKeyNew("pk_id")
	if err != nil {
		t.Fatalf("Failed to create primary key index after reopen: %v", err)
	}
	pkIndex2.AddFields("id")
	err = table2.CreateIndex(pkIndex2)
	if err != nil {
		t.Fatalf("Failed to create primary key index after reopen: %v", err)
	}

	// 重新初始化自动增值计数器，现在表结构已加载，可以正确获取最大ID值
	table2.InitAuto()

	// 步骤3: 再次插入记录，检查计数器是否连续
	for i := 0; i < 2; i++ {
		record := map[string]any{
			"name": fmt.Sprintf("User%d", 3+i),
			"age":  23 + i,
		}

		id, err := table2.Insert(&record)
		if err != nil {
			t.Fatalf("Failed to insert record after reopen %d: %v", i, err)
		}

		expectedID = 4 + i // 应该从4开始继续
		if id != expectedID {
			t.Errorf("Expected ID %d after reopen, got %d", expectedID, id)
		}
		t.Logf("Inserted record after reopen with ID: %d", id)
	}

	// 步骤4: 验证所有记录都存在
	allRecords := make(map[int]bool)
	searchFields := map[string]any{"id": nil} // 全表扫描
	tableIter, err := table2.Search(&searchFields)
	if err != nil {
		t.Fatalf("Failed to search all records: %v", err)
	}
	defer tableIter.Release()

	for tableIter.Next() {
		key := tableIter.Key()
		value := tableIter.Value()
		fieldsBytes := tableIter.ParseBytes(key, value)
		if fieldsBytes != nil {
			record := tableIter.ParseRecord(fieldsBytes)
			if record != nil {
				id := record["id"].(int)
				allRecords[id] = true
				t.Logf("Found record with ID: %d", id)
			}
		}
	}

	// 验证所有5个记录（3个初始 + 2个重新打开后）都存在
	expectedCount := 5
	if len(allRecords) != expectedCount {
		t.Errorf("Expected %d records, got %d", expectedCount, len(allRecords))
	}

	// 验证ID范围是否正确
	for i := 1; i <= expectedCount; i++ {
		if !allRecords[i] {
			t.Errorf("Missing record with ID: %d", i)
		}
	}

	t.Logf("TestTableCounterContinuity completed successfully! Total records: %d", len(allRecords))
}

// TestTableCounterInit 测试表初始化时计数器的设置
func TestTableCounterInit(t *testing.T) {
	// 创建唯一表名
	tableName := fmt.Sprintf("test_counter_init_%d", time.Now().UnixNano())

	// 步骤1: 创建表并手动插入记录（使用自定义ID）
	table, err := TableNew(tableName)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段和主键
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
	pkIndex, err := DefaultPrimaryKeyNew("pk_id")
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}
	pkIndex.AddFields("id")
	err = table.CreateIndex(pkIndex)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// 手动插入几条记录，使用自定义ID
	customIDs := []int{5, 10, 15}
	for _, id := range customIDs {
		record := map[string]any{
			"id":   id,
			"name": fmt.Sprintf("User%d", id),
			"age":  30,
		}

		_, err := table.Insert(&record)
		if err != nil {
			t.Fatalf("Failed to insert record with custom ID %d: %v", id, err)
		}
		t.Logf("Inserted record with custom ID: %d", id)
	}

	// 步骤2: 重新打开表
	table2, err := TableNew(tableName)
	if err != nil {
		t.Fatalf("Failed to reopen table: %v", err)
	}

	// 重新设置字段和索引信息，因为TableNew不会自动加载
	err = table2.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields after reopen: %v", err)
	}

	// 重新创建主键索引
	pkIndex2, err := DefaultPrimaryKeyNew("pk_id")
	if err != nil {
		t.Fatalf("Failed to create primary key index after reopen: %v", err)
	}
	pkIndex2.AddFields("id")
	err = table2.CreateIndex(pkIndex2)
	if err != nil {
		t.Fatalf("Failed to create primary key index after reopen: %v", err)
	}

	// 重新初始化自动增值计数器，现在表结构已加载，可以正确获取最大ID值
	table2.InitAuto()

	// 步骤3: 插入新记录，检查是否从最大ID+1开始
	record := map[string]any{
		"name": "NewUser",
		"age":  25,
	}

	id, err := table2.Insert(&record)
	if err != nil {
		t.Fatalf("Failed to insert record after reopen: %v", err)
	}

	// 应该从16开始（最大自定义ID是15）
	expectedID := 16
	if id != expectedID {
		t.Errorf("Expected ID %d after custom IDs, got %d", expectedID, id)
	}
	t.Logf("Inserted record after custom IDs with ID: %d", id)

	// 步骤4: 验证所有记录都存在
	allRecords := make(map[int]bool)
	searchFields := map[string]any{"id": nil} // 全表扫描
	tableIter, err := table2.Search(&searchFields)
	if err != nil {
		t.Fatalf("Failed to search all records: %v", err)
	}
	defer tableIter.Release()

	for tableIter.Next() {
		key := tableIter.Key()
		value := tableIter.Value()
		fieldsBytes := tableIter.ParseBytes(key, value)
		if fieldsBytes != nil {
			record := tableIter.ParseRecord(fieldsBytes)
			if record != nil {
				id := record["id"].(int)
				allRecords[id] = true
			}
		}
	}

	// 验证所有4个记录（3个自定义 + 1个自动增长）都存在
	expectedCount := 4
	if len(allRecords) != expectedCount {
		t.Errorf("Expected %d records, got %d", expectedCount, len(allRecords))
	}

	// 验证自定义ID和新的自动增长ID都存在
	for _, id := range customIDs {
		if !allRecords[id] {
			t.Errorf("Missing custom ID record: %d", id)
		}
	}
	if !allRecords[expectedID] {
		t.Errorf("Missing auto-increment ID record: %d", expectedID)
	}

	t.Logf("TestTableCounterInit completed successfully! Total records: %d", len(allRecords))
}
