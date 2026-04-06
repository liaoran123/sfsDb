package engine

import (
	"fmt"
	"os"
	"testing"

	"github.com/liaoran123/sfsDb/record"
	"github.com/liaoran123/sfsDb/storage"
	"github.com/liaoran123/sfsDb/util"
)

// TestTableSnapshotSearchConsistency 测试使用 Searchs 函数的快照读一致性功能
func TestTableSnapshotSearchConsistency(t *testing.T) {
	// 创建临时数据库
	dbPath := "./test_table_snapshot_db"
	cleanup := func() {
		removeAll(dbPath)
	}
	cleanup()
	defer cleanup()

	// 打开数据库
	dbManager := storage.GetDBManager()
	db, err := dbManager.OpenDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer dbManager.CloseDB()

	// 创建表
	tableName := "test_users"

	table, err := NewTable(tableName)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 准备字段映射
	fieldsMap := map[string]any{
		"id":    0,
		"name":  "",
		"age":   0,
		"email": "",
	}

	// 设置表的字段
	err = table.SetFields(fieldsMap)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键索引
	PrimaryKeys, err := NewDefaultPrimaryKey("pk")
	if err != nil {
		t.Fatalf("Failed to create primary key: %v", err)
	}
	PrimaryKeys.AddFields("id")
	err = table.CreateIndex(PrimaryKeys)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// 为 age 字段创建普通索引，用于搜索
	ageIndex, err := NewDefaultNormalIndex("age_idx")
	if err != nil {
		t.Fatalf("Failed to create age index: %v", err)
	}
	ageIndex.AddFields("age")
	err = table.CreateIndex(ageIndex)
	if err != nil {
		t.Fatalf("Failed to create age index: %v", err)
	}

	// 1. 插入初始数据
	users := []map[string]any{
		{"id": 1, "name": "张三", "age": 30, "email": "zhangsan@example.com"},
		{"id": 2, "name": "李四", "age": 25, "email": "lisi@example.com"},
		{"id": 3, "name": "王五", "age": 35, "email": "wangwu@example.com"},
		{"id": 4, "name": "赵六", "age": 28, "email": "zhaoliu@example.com"},
		{"id": 5, "name": "孙七", "age": 40, "email": "sunqi@example.com"},
	}

	for _, user := range users {
		_, err := table.Insert(&user)
		if err != nil {
			t.Fatalf("Failed to insert user: %v", err)
		}
	}

	// 验证初始数据
	fmt.Println("Step 1: Initial data inserted successfully")
	printTableData(t, table, "Initial data")

	// 2. 创建快照
	snapshot, err := db.Snapshot()
	if err != nil {
		t.Fatalf("Failed to create snapshot: %v", err)
	}
	defer snapshot.Release()

	fmt.Println("Step 2: Snapshot created successfully")

	// 3. 修改数据
	updateUsers := []map[string]any{
		{"id": 1, "name": "张三(已更新)", "age": 31},
		{"id": 2, "name": "李四(已更新)", "age": 26},
		{"id": 6, "name": "周八", "age": 22, "email": "zhouba@example.com"}, // 新增记录
	}

	for _, user := range updateUsers {
		if user["id"] == 6 {
			// 新增记录
			_, err := table.Insert(&user)
			if err != nil {
				t.Fatalf("Failed to insert new user: %v", err)
			}
		} else {
			// 更新记录
			err := table.Update(&user)
			if err != nil {
				t.Fatalf("Failed to update user: %v", err)
			}
		}
	}

	// 验证数据已修改
	fmt.Println("Step 3: Data updated successfully after snapshot creation")
	printTableData(t, table, "Updated data")

	// 4. 直接使用快照验证读一致性
	fmt.Println("Step 4: Testing snapshot read consistency directly")

	// 创建从快照中获取记录的函数
	snapshotGet := func(key []byte) []byte {
		value, err := snapshot.Get(key)
		if err != nil {
			return nil
		}
		return value
	}

	// 为每个用户ID构建主键并验证数据
	userIDs := []int{1, 2, 3, 4, 5}
	for _, id := range userIDs {
		// 构建主键
		pk := PrimaryKeys
		fieldsBytes := map[string][]byte{
			"id": util.AnyToBytes(id),
		}
		key := pk.JoinValue(&fieldsBytes, table.id)

		// 从快照中读取
		snapshotValue := snapshotGet(key)
		// 从数据库中读取
		dbValue, err := table.ReadByBytes(key)
		if err != nil {
			t.Fatalf("Failed to read record from db: %v", err)
		}

		// 解析记录
		var snapshotRecord, dbRecord *map[string]any
		if snapshotValue != nil {
			parsed, err := pk.Parse(table.fieldsid, snapshotValue)
			if err == nil && parsed != nil {
				snapshotRecord = table.RecordByteToAny(parsed)
				defer GlobalFieldsBytesPool.Put(*parsed)
			}
		}
		if dbValue != nil {
			parsed, err := pk.Parse(table.fieldsid, dbValue)
			if err == nil && parsed != nil {
				dbRecord = table.RecordByteToAny(parsed)
				defer GlobalFieldsBytesPool.Put(*parsed)
			}
		}

		// 验证快照记录是原始数据
		if snapshotRecord != nil {
			fmt.Printf("Snapshot data for ID %d: Name=%s, Age=%d\n", id, (*snapshotRecord)["name"], (*snapshotRecord)["age"])
			// 验证原始数据
			switch id {
			case 1:
				if (*snapshotRecord)["name"] != "张三" || (*snapshotRecord)["age"] != 30 {
					t.Fatalf("Snapshot data for ID 1 is not original: Name=%s, Age=%d", (*snapshotRecord)["name"], (*snapshotRecord)["age"])
				}
			case 2:
				if (*snapshotRecord)["name"] != "李四" || (*snapshotRecord)["age"] != 25 {
					t.Fatalf("Snapshot data for ID 2 is not original: Name=%s, Age=%d", (*snapshotRecord)["name"], (*snapshotRecord)["age"])
				}
			case 3:
				if (*snapshotRecord)["name"] != "王五" || (*snapshotRecord)["age"] != 35 {
					t.Fatalf("Snapshot data for ID 3 is not original: Name=%s, Age=%d", (*snapshotRecord)["name"], (*snapshotRecord)["age"])
				}
			case 4:
				if (*snapshotRecord)["name"] != "赵六" || (*snapshotRecord)["age"] != 28 {
					t.Fatalf("Snapshot data for ID 4 is not original: Name=%s, Age=%d", (*snapshotRecord)["name"], (*snapshotRecord)["age"])
				}
			case 5:
				if (*snapshotRecord)["name"] != "孙七" || (*snapshotRecord)["age"] != 40 {
					t.Fatalf("Snapshot data for ID 5 is not original: Name=%s, Age=%d", (*snapshotRecord)["name"], (*snapshotRecord)["age"])
				}
			}
		}

		// 验证数据库记录是修改后的数据
		if dbRecord != nil {
			fmt.Printf("Database data for ID %d: Name=%s, Age=%d\n", id, (*dbRecord)["name"], (*dbRecord)["age"])
			// 验证修改后的数据
			switch id {
			case 1:
				if (*dbRecord)["name"] != "张三(已更新)" || (*dbRecord)["age"] != 31 {
					t.Fatalf("Database data for ID 1 is not updated: Name=%s, Age=%d", (*dbRecord)["name"], (*dbRecord)["age"])
				}
			case 2:
				if (*dbRecord)["name"] != "李四(已更新)" || (*dbRecord)["age"] != 26 {
					t.Fatalf("Database data for ID 2 is not updated: Name=%s, Age=%d", (*dbRecord)["name"], (*dbRecord)["age"])
				}
			}
		}
	}

	// 验证快照不包含新增记录（ID: 6）
	fmt.Println("\nStep 4.1: Testing snapshot does not include new record")
	newRecordID := 6
	newRecordFieldsBytes := map[string][]byte{
		"id": util.AnyToBytes(newRecordID),
	}
	newRecordKey := PrimaryKeys.JoinValue(&newRecordFieldsBytes, table.id)
	newRecordSnapshotValue := snapshotGet(newRecordKey)
	if newRecordSnapshotValue != nil {
		t.Fatalf("Snapshot should not include new record with ID %d", newRecordID)
	}
	fmt.Println("✓ Snapshot does not include new record with ID", newRecordID)

	// 5. 测试快照迭代器
	fmt.Println("\nStep 5: Testing snapshot iterator")

	// 创建使用快照迭代器的 FunIter 函数
	snapshotFunIter := func(start, limit []byte) storage.Iterator {
		return snapshot.Iterator(start, limit)
	}

	// 测试范围搜索
	rangeIter, err := table.SearchRange(snapshotFunIter, &map[string]any{"id": 2}, &map[string]any{"id": 5})
	defer rangeIter.Release()
	if err != nil {
		t.Fatalf("Failed to search range with snapshot: %v", err)
	}

	// 收集范围搜索结果
	rangeResults := rangeIter.GetRecords(true)
	defer rangeResults.Release()

	fmt.Println("Range search results (id between 2 and 4):")
	for _, result := range rangeResults {
		fmt.Printf("  ID: %d, Name: %s, Age: %d\n", result["id"], result["name"], result["age"])
	}

	// 验证范围搜索结果数量
	if len(rangeResults) != 3 {
		t.Fatalf("Expected 3 results from range search, got %d", len(rangeResults))
	}

	// 6. 总结
	fmt.Println("\n=== Test Summary ===")
	fmt.Println("✓ Initial data inserted successfully")
	fmt.Println("✓ Snapshot created successfully")
	fmt.Println("✓ Data updated successfully after snapshot creation")
	fmt.Println("✓ Snapshot read consistency verified directly")
	fmt.Println("✓ Range search with snapshot works correctly")
	fmt.Println("\nConclusion: Snapshot provides consistent read view")
	fmt.Println("Even when data is modified after snapshot creation, the snapshot still returns the original data")
}

// printTableData 打印表中的所有数据
func printTableData(t *testing.T, table *Table, title string) {
	iter := table.ForData()
	defer GlobalTableIterPool.Put(iter)

	fmt.Printf("\n%s:\n", title)

	// 使用 GetRecords 方法获取所有记录
	records := iter.GetRecords(true)
	defer record.PutRecords(records)

	for _, record := range records {
		fmt.Printf("  ID: %d, Name: %s, Age: %d, Email: %s\n", record["id"], record["name"], record["age"], record["email"])
	}
}

// removeAll 删除目录（用于测试清理）
func removeAll(path string) {
	os.RemoveAll(path)
}
