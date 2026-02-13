package engine

import (
	"fmt"
	"testing"

	"github.com/liaoran123/sfsDb/storage"
)

// TestTransactionManager 测试事务管理器功能
func TestTransactionManager(t *testing.T) {
	// 清理测试环境
	defer storage.CloseDb()

	// 初始化数据库
	dbPath := "./test_transaction_manager"
	_, err := storage.OpenDefaultDb(dbPath)
	if err != nil {
		t.Fatalf("OpenDefaultDb failed: %v", err)
	}

	// 创建表1
	table1, err := TableNew("users")
	if err != nil {
		t.Fatalf("TableNew users failed: %v", err)
	}

	// 设置表1字段
	fields1 := map[string]any{
		"id":   0,
		"name": "",
	}
	err = table1.SetFields(fields1)
	if err != nil {
		t.Fatalf("SetFields users failed: %v", err)
	}

	// 创建表1主键索引
	pk1, _ := DefaultPrimaryKeyNew("pk_id")
	pk1.AddFields("id")
	table1.CreateIndex(pk1)

	// 创建表2
	table2, err := TableNew("orders")
	if err != nil {
		t.Fatalf("TableNew orders failed: %v", err)
	}

	// 设置表2字段
	fields2 := map[string]any{
		"id":      0,
		"user_id": 0,
	}
	err = table2.SetFields(fields2)
	if err != nil {
		t.Fatalf("SetFields orders failed: %v", err)
	}

	// 创建表2主键索引
	pk2, _ := DefaultPrimaryKeyNew("pk_id")
	pk2.AddFields("id")
	table2.CreateIndex(pk2)

	// 创建共享的batch
	batch := storage.KVDb.GetBatch()
	if batch == nil {
		t.Fatalf("GetBatch failed")
	}

	// 创建事务管理器
	tm := NewTransactionManager(batch)

	// 添加表到事务管理器
	tx1, err := tm.AddTable(table1)
	if err != nil {
		t.Fatalf("AddTable users failed: %v", err)
	}

	tx2, err := tm.AddTable(table2)
	if err != nil {
		t.Fatalf("AddTable orders failed: %v", err)
	}

	// 在事务中操作表
	user := map[string]any{"name": "张三"}
	userID, err := tx1.Insert(&user)
	if err != nil {
		t.Fatalf("Insert user failed: %v", err)
	}
	fmt.Printf("Inserted user with ID: %d\n", userID)

	order := map[string]any{"user_id": userID}
	orderID, err := tx2.Insert(&order)
	if err != nil {
		t.Fatalf("Insert order failed: %v", err)
	}
	fmt.Printf("Inserted order with ID: %d\n", orderID)

	// 提交事务
	err = tm.Commit()
	if err != nil {
		t.Fatalf("Commit failed: %v", err)
	}
	fmt.Println("Transaction committed successfully")

	// 验证数据
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

	fmt.Println("Transaction manager test passed successfully!")
}

// TestTransactionManagerRollback 测试事务管理器回滚功能
func TestTransactionManagerRollback(t *testing.T) {
	// 清理测试环境
	defer storage.CloseDb()

	// 初始化数据库
	dbPath := "./test_transaction_manager_rollback"
	_, err := storage.OpenDefaultDb(dbPath)
	if err != nil {
		t.Fatalf("OpenDefaultDb failed: %v", err)
	}

	// 创建表
	table1, err := TableNew("users")
	if err != nil {
		t.Fatalf("TableNew users failed: %v", err)
	}

	// 设置表字段
	fields1 := map[string]any{
		"id":   0,
		"name": "",
	}
	err = table1.SetFields(fields1)
	if err != nil {
		t.Fatalf("SetFields users failed: %v", err)
	}

	// 创建主键索引
	pk1, _ := DefaultPrimaryKeyNew("pk_id")
	pk1.AddFields("id")
	table1.CreateIndex(pk1)

	// 创建共享的batch
	batch := storage.KVDb.GetBatch()
	if batch == nil {
		t.Fatalf("GetBatch failed")
	}

	// 创建事务管理器
	tm := NewTransactionManager(batch)

	// 添加表到事务管理器
	tx1, err := tm.AddTable(table1)
	if err != nil {
		t.Fatalf("AddTable users failed: %v", err)
	}

	// 在事务中操作表
	user := map[string]any{"name": "李四"}
	tx1.Insert(&user)

	// 回滚事务
	err = tm.Rollback()
	if err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	// 验证数据是否被回滚
	userIter := table1.ForData()
	if userIter == nil {
		t.Fatalf("ForData failed")
	}
	defer GlobalTableIterPool.Put(userIter)

	userRecords := userIter.GetRecords(true)
	if len(userRecords) != 0 {
		t.Fatalf("Expected 0 user records after rollback, got %d", len(userRecords))
	}

	fmt.Println("Transaction manager rollback test passed successfully!")
}

// TestWithTransaction 测试WithTransaction便捷函数
func TestWithTransaction(t *testing.T) {
	// 清理测试环境
	defer storage.CloseDb()

	// 初始化数据库
	dbPath := "./test_with_transaction"
	_, err := storage.OpenDefaultDb(dbPath)
	if err != nil {
		t.Fatalf("OpenDefaultDb failed: %v", err)
	}

	// 创建表1
	table1, err := TableNew("users")
	if err != nil {
		t.Fatalf("TableNew users failed: %v", err)
	}

	// 设置表1字段
	fields1 := map[string]any{
		"id":   0,
		"name": "",
	}
	err = table1.SetFields(fields1)
	if err != nil {
		t.Fatalf("SetFields users failed: %v", err)
	}

	// 创建表1主键索引
	pk1, _ := DefaultPrimaryKeyNew("pk_id")
	pk1.AddFields("id")
	table1.CreateIndex(pk1)

	// 创建表2
	table2, err := TableNew("orders")
	if err != nil {
		t.Fatalf("TableNew orders failed: %v", err)
	}

	// 设置表2字段
	fields2 := map[string]any{
		"id":      0,
		"user_id": 0,
	}
	err = table2.SetFields(fields2)
	if err != nil {
		t.Fatalf("SetFields orders failed: %v", err)
	}

	// 创建表2主键索引
	pk2, _ := DefaultPrimaryKeyNew("pk_id")
	pk2.AddFields("id")
	table2.CreateIndex(pk2)

	// 创建共享的batch
	batch := storage.KVDb.GetBatch()
	if batch == nil {
		t.Fatalf("GetBatch failed")
	}

	// 使用WithTransaction便捷函数
	var userID, orderID int
	err = WithTransaction(batch, []*Table{table1, table2}, func(transactions map[*Table]Transaction) error {
		// 获取表1的事务
		tx1, ok := transactions[table1]
		if !ok {
			return fmt.Errorf("table1 transaction not found")
		}

		// 获取表2的事务
		tx2, ok := transactions[table2]
		if !ok {
			return fmt.Errorf("table2 transaction not found")
		}

		// 在事务中操作表1
		user := map[string]any{"name": "王五"}
		var err error
		userID, err = tx1.Insert(&user)
		if err != nil {
			return err
		}
		fmt.Printf("Inserted user with ID: %d\n", userID)

		// 在事务中操作表2
		order := map[string]any{"user_id": userID}
		orderID, err = tx2.Insert(&order)
		if err != nil {
			return err
		}
		fmt.Printf("Inserted order with ID: %d\n", orderID)

		return nil
	})

	if err != nil {
		t.Fatalf("WithTransaction failed: %v", err)
	}

	fmt.Println("WithTransaction test passed successfully!")

	// 验证数据
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
}

// TestWithTransactionError 测试WithTransaction错误处理
func TestWithTransactionError(t *testing.T) {
	// 清理测试环境
	defer storage.CloseDb()

	// 初始化数据库
	dbPath := "./test_with_transaction_error"
	_, err := storage.OpenDefaultDb(dbPath)
	if err != nil {
		t.Fatalf("OpenDefaultDb failed: %v", err)
	}

	// 创建表
	table1, err := TableNew("users")
	if err != nil {
		t.Fatalf("TableNew users failed: %v", err)
	}

	// 设置表字段
	fields1 := map[string]any{
		"id":   0,
		"name": "",
	}
	err = table1.SetFields(fields1)
	if err != nil {
		t.Fatalf("SetFields users failed: %v", err)
	}

	// 创建主键索引
	pk1, _ := DefaultPrimaryKeyNew("pk_id")
	pk1.AddFields("id")
	table1.CreateIndex(pk1)

	// 创建共享的batch
	batch := storage.KVDb.GetBatch()
	if batch == nil {
		t.Fatalf("GetBatch failed")
	}

	// 模拟错误
	expectedErr := fmt.Errorf("simulated error")
	err = WithTransaction(batch, []*Table{table1}, func(transactions map[*Table]Transaction) error {
		// 获取表1的事务
		tx1, ok := transactions[table1]
		if !ok {
			return fmt.Errorf("table1 transaction not found")
		}

		// 在事务中操作表
		user := map[string]any{"name": "赵六"}
		tx1.Insert(&user)

		// 模拟错误
		return expectedErr
	})

	// 验证错误是否正确返回
	if err == nil || err != expectedErr {
		t.Fatalf("Expected error %v, got %v", expectedErr, err)
	}

	// 验证数据是否被回滚
	userIter := table1.ForData()
	if userIter == nil {
		t.Fatalf("ForData failed")
	}
	defer GlobalTableIterPool.Put(userIter)

	userRecords := userIter.GetRecords(true)
	if len(userRecords) != 0 {
		t.Fatalf("Expected 0 user records after error, got %d", len(userRecords))
	}

	fmt.Println("WithTransaction error handling test passed successfully!")
}
