package engine

import (
	"fmt"
	"testing"

	"github.com/liaoran123/sfsDb/storage"
)

// TestMultiTableTransaction 测试多表事务功能
func TestMultiTableTransaction(t *testing.T) {
	// 清理测试环境
	defer storage.CloseDb()

	// 初始化数据库
	dbPath := "./test_multitable"
	_, err := storage.OpenDefaultDb(dbPath)
	if err != nil {
		t.Fatalf("OpenDefaultDb failed: %v", err)
	}

	// 创建第一个表
	table1, err := TableNew("users")
	if err != nil {
		t.Fatalf("TableNew users failed: %v", err)
	}

	// 设置表1字段
	fields1 := map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	}
	err = table1.SetFields(fields1)
	if err != nil {
		t.Fatalf("SetFields users failed: %v", err)
	}

	// 创建表1主键索引
	pk1, err := DefaultPrimaryKeyNew("pk_id")
	if err != nil {
		t.Fatalf("DefaultPrimaryKeyNew failed: %v", err)
	}
	pk1.AddFields("id")
	err = table1.CreateIndex(pk1)
	if err != nil {
		t.Fatalf("CreateIndex users failed: %v", err)
	}

	// 创建第二个表
	table2, err := TableNew("orders")
	if err != nil {
		t.Fatalf("TableNew orders failed: %v", err)
	}

	// 设置表2字段
	fields2 := map[string]any{
		"id":      0,
		"user_id": 0,
		"amount":  0.0,
	}
	err = table2.SetFields(fields2)
	if err != nil {
		t.Fatalf("SetFields orders failed: %v", err)
	}

	// 创建表2主键索引
	pk2, err := DefaultPrimaryKeyNew("pk_id")
	if err != nil {
		t.Fatalf("DefaultPrimaryKeyNew failed: %v", err)
	}
	pk2.AddFields("id")
	err = table2.CreateIndex(pk2)
	if err != nil {
		t.Fatalf("CreateIndex orders failed: %v", err)
	}

	// 创建共享的batch
	batch := storage.KVDb.GetBatch()
	if batch == nil {
		t.Fatalf("GetBatch failed")
	}

	// 为两个表创建使用同一个batch的事务
	tx1, err := table1.BeginWithBatch(batch)
	if err != nil {
		t.Fatalf("BeginWithBatch users failed: %v", err)
	}

	tx2, err := table2.BeginWithBatch(batch)
	if err != nil {
		t.Fatalf("BeginWithBatch orders failed: %v", err)
	}

	// 在事务中操作第一个表
	user := map[string]any{
		"name": "张三",
		"age":  30,
	}
	userID, err := tx1.Insert(&user)
	if err != nil {
		t.Fatalf("Insert user failed: %v", err)
	}
	fmt.Printf("Inserted user with ID: %d\n", userID)

	// 在事务中操作第二个表
	order := map[string]any{
		"user_id": userID,
		"amount":  100.50,
	}
	orderID, err := tx2.Insert(&order)
	if err != nil {
		t.Fatalf("Insert order failed: %v", err)
	}
	fmt.Printf("Inserted order with ID: %d\n", orderID)

	// 提交事务（只需提交一次，因为所有操作都在同一个batch中）
	err = tx1.Commit()
	if err != nil {
		t.Fatalf("Commit tx1 failed: %v", err)
	}
	fmt.Println("Transaction committed successfully")

	// 第二个事务不需要再提交，因为batch已经被提交
	tx2.Rollback() // 只是释放资源

	// 验证第一个表中的数据
	userSearch := map[string]any{"id": userID}
	userIter, _ := table1.Search(&userSearch)
	if userIter == nil {
		t.Fatalf("Search user failed")
	}
	defer GlobalTableIterPool.Put(userIter)

	userRecords := userIter.GetRecords(true)
	if len(userRecords) != 1 {
		t.Fatalf("Expected 1 user record, got %d", len(userRecords))
	}
	fmt.Printf("Found user: %v\n", userRecords[0])

	// 验证第二个表中的数据
	orderSearch := map[string]any{"id": orderID}
	orderIter, _ := table2.Search(&orderSearch)
	if orderIter == nil {
		t.Fatalf("Search order failed")
	}
	defer GlobalTableIterPool.Put(orderIter)

	orderRecords := orderIter.GetRecords(true)
	if len(orderRecords) != 1 {
		t.Fatalf("Expected 1 order record, got %d", len(orderRecords))
	}
	fmt.Printf("Found order: %v\n", orderRecords[0])

	// 验证外键关联
	if orderRecords[0]["user_id"] != userID {
		t.Fatalf("Expected order.user_id = %d, got %v", userID, orderRecords[0]["user_id"])
	}

	fmt.Println("Multi-table transaction test passed successfully!")
}

// TestMultiTableTransactionRollback 测试多表事务回滚功能
func TestMultiTableTransactionRollback(t *testing.T) {
	// 清理测试环境
	defer storage.CloseDb()

	// 初始化数据库
	dbPath := "./test_multitable_rollback"
	_, err := storage.OpenDefaultDb(dbPath)
	if err != nil {
		t.Fatalf("OpenDefaultDb failed: %v", err)
	}

	// 创建表
	table1, err := TableNew("users")
	if err != nil {
		t.Fatalf("TableNew users failed: %v", err)
	}

	table2, err := TableNew("orders")
	if err != nil {
		t.Fatalf("TableNew orders failed: %v", err)
	}

	// 设置字段
	fields1 := map[string]any{"id": 0, "name": ""}
	err = table1.SetFields(fields1)
	if err != nil {
		t.Fatalf("SetFields users failed: %v", err)
	}

	fields2 := map[string]any{"id": 0, "user_id": 0}
	err = table2.SetFields(fields2)
	if err != nil {
		t.Fatalf("SetFields orders failed: %v", err)
	}

	// 创建主键索引
	pk1, _ := DefaultPrimaryKeyNew("pk_id")
	pk1.AddFields("id")
	table1.CreateIndex(pk1)

	pk2, _ := DefaultPrimaryKeyNew("pk_id")
	pk2.AddFields("id")
	table2.CreateIndex(pk2)

	// 创建共享的batch
	batch := storage.KVDb.GetBatch()
	if batch == nil {
		t.Fatalf("GetBatch failed")
	}

	// 为两个表创建事务
	tx1, _ := table1.BeginWithBatch(batch)
	tx2, _ := table2.BeginWithBatch(batch)

	// 在事务中操作表
	user := map[string]any{"name": "李四"}
	tx1.Insert(&user)

	order := map[string]any{"user_id": 1}
	tx2.Insert(&order)

	// 回滚事务
	tx1.Rollback()
	tx2.Rollback()

	// 验证数据是否被回滚
	userIter := table1.ForData()
	defer GlobalTableIterPool.Put(userIter)
	userRecords := userIter.GetRecords(true)
	if len(userRecords) != 0 {
		t.Fatalf("Expected 0 user records after rollback, got %d", len(userRecords))
	}

	orderIter := table2.ForData()
	defer GlobalTableIterPool.Put(orderIter)
	orderRecords := orderIter.GetRecords(true)
	if len(orderRecords) != 0 {
		t.Fatalf("Expected 0 order records after rollback, got %d", len(orderRecords))
	}

	fmt.Println("Multi-table transaction rollback test passed successfully!")
}
