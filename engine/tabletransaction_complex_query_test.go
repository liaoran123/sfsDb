package engine

import (
	"fmt"
	"testing"

	"github.com/liaoran123/sfsDb/match"
	"github.com/liaoran123/sfsDb/storage"
	"github.com/liaoran123/sfsDb/util"
)

// TestTransactionComplexQuery 测试事务对复杂查询功能的支持
func TestTransactionComplexQuery(t *testing.T) {
	// 清理测试环境
	defer storage.CloseDb()

	// 初始化数据库
	dbPath := "./test_transaction_complex_query"
	_, err := storage.OpenDefaultDb(dbPath)
	if err != nil {
		t.Fatalf("OpenDefaultDb failed: %v", err)
	}

	// 创建表
	table, err := TableNew("users")
	if err != nil {
		t.Fatalf("TableNew failed: %v", err)
	}

	// 设置表字段
	fields := map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
		"city": "",
	}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("SetFields failed: %v", err)
	}

	// 创建主键索引
	pk, err := DefaultPrimaryKeyNew("pk_id")
	if err != nil {
		t.Fatalf("DefaultPrimaryKeyNew failed: %v", err)
	}
	pk.AddFields("id")
	err = table.CreateIndex(pk)
	if err != nil {
		t.Fatalf("CreateIndex failed: %v", err)
	}

	// 创建普通索引
	idx, err := DefaultNormalIndexNew("idx_age")
	if err != nil {
		t.Fatalf("DefaultNormalIndexNew failed: %v", err)
	}
	idx.AddFields("age")
	err = table.CreateIndex(idx)
	if err != nil {
		t.Fatalf("CreateIndex failed: %v", err)
	}

	// 插入初始数据
	initialData := []map[string]any{
		{"name": "张三", "age": 25, "city": "北京"},
		{"name": "李四", "age": 30, "city": "上海"},
		{"name": "王五", "age": 35, "city": "广州"},
		{"name": "赵六", "age": 20, "city": "北京"},
		{"name": "孙七", "age": 40, "city": "上海"},
	}

	for _, data := range initialData {
		table.Insert(&data)
	}

	// 开始事务
	tx, err := table.Begin()
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	// 在事务中插入新数据
	newUser := map[string]any{"name": "周八", "age": 28, "city": "北京"}
	userID, err := tx.Insert(&newUser)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}
	fmt.Printf("Inserted new user with ID: %d\n", userID)

	// 测试 1: 基本搜索 - 年龄大于 25
	fmt.Println("\nTest 1: Search age > 25")
	ageSearch := map[string]any{"age": 25}
	ageIter, _ := tx.Search(&ageSearch, util.GreaterThan)
	if ageIter == nil {
		t.Fatalf("Search failed")
	}
	ageRecords := ageIter.GetRecords(true)
	fmt.Printf("Found %d records for age > 25\n", len(ageRecords))
	for _, record := range ageRecords {
		fmt.Printf("  - %v\n", record)
	}

	// 测试 2: 多条件搜索 - 年龄大于 25 且城市为北京
	fmt.Println("\nTest 2: Multi-condition search")
	// 先获取所有记录
	allIter, _ := tx.Search(&map[string]any{"id": nil})
	if allIter == nil {
		t.Fatalf("Search failed")
	}

	// 使用 FieldComparison 匹配器
	ageMatcher := match.NewGreaterThanMatch("age", 25)
	cityMatcher := match.NewEqualMatch("city", "北京")

	// 设置匹配器（多个匹配器之间是 AND 关系）
	allIter.SetMatch(ageMatcher, cityMatcher)
	matchedRecords := allIter.GetRecords(true)
	fmt.Printf("Found %d records for age > 25 and city = '北京'\n", len(matchedRecords))
	for _, record := range matchedRecords {
		fmt.Printf("  - %v\n", record)
	}

	// 测试 3: 使用 FieldComparison 匹配器
	fmt.Println("\nTest 3: FieldComparison matcher")
	fcIter, _ := tx.Search(&map[string]any{"id": nil})
	if fcIter == nil {
		t.Fatalf("Search failed")
	}

	fcMatcher := match.NewFieldComparison("age", match.GreaterThan, 25)
	fcIter.SetMatch(fcMatcher)
	fcRecords := fcIter.GetRecords(true)
	fmt.Printf("Found %d records using FieldComparison (age > 25)\n", len(fcRecords))

	// 测试 4: 测试事务内的 Read 方法（可以读取未提交的修改）
	fmt.Println("\nTest 4: Read method in transaction")
	readFields := map[string]any{"id": userID}
	readData, err := tx.Read(&readFields)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	fmt.Printf("Read data for new user: %v\n", string(readData))

	// 提交事务
	err = tx.Commit()
	if err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	// 验证事务提交后数据是否正确
	fmt.Println("\nTest 5: Verify data after commit")
	verifyIter, _ := table.Search(&map[string]any{"id": userID})
	if verifyIter == nil {
		t.Fatalf("Search failed")
	}
	verifyRecords := verifyIter.GetRecords(true)
	if len(verifyRecords) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(verifyRecords))
	}
	fmt.Printf("Verified new user: %v\n", verifyRecords[0])

	fmt.Println("\nAll tests passed successfully!")
}

// TestTransactionMultiTableQuery 测试事务对多表连接查询的支持
func TestTransactionMultiTableQuery(t *testing.T) {
	// 清理测试环境
	defer storage.CloseDb()

	// 初始化数据库
	dbPath := "./test_transaction_multitable_query"
	_, err := storage.OpenDefaultDb(dbPath)
	if err != nil {
		t.Fatalf("OpenDefaultDb failed: %v", err)
	}

	// 创建用户表
	userTable, err := TableNew("users")
	if err != nil {
		t.Fatalf("TableNew users failed: %v", err)
	}

	// 设置用户表字段
	userFields := map[string]any{
		"id":   0,
		"name": "",
	}
	err = userTable.SetFields(userFields)
	if err != nil {
		t.Fatalf("SetFields users failed: %v", err)
	}

	// 创建用户表主键索引
	userPK, err := DefaultPrimaryKeyNew("pk_id")
	if err != nil {
		t.Fatalf("DefaultPrimaryKeyNew failed: %v", err)
	}
	userPK.AddFields("id")
	userTable.CreateIndex(userPK)

	// 创建订单表
	orderTable, err := TableNew("orders")
	if err != nil {
		t.Fatalf("TableNew orders failed: %v", err)
	}

	// 设置订单表字段
	orderFields := map[string]any{
		"id":      0,
		"user_id": 0,
		"amount":  0.0,
	}
	err = orderTable.SetFields(orderFields)
	if err != nil {
		t.Fatalf("SetFields orders failed: %v", err)
	}

	// 创建订单表主键索引
	orderPK, err := DefaultPrimaryKeyNew("pk_id")
	if err != nil {
		t.Fatalf("DefaultPrimaryKeyNew failed: %v", err)
	}
	orderPK.AddFields("id")
	orderTable.CreateIndex(orderPK)

	// 创建共享的 batch
	batch := storage.KVDb.GetBatch()
	if batch == nil {
		t.Fatalf("GetBatch failed")
	}

	// 使用 WithTransaction 执行多表事务
	err = WithTransaction(batch, []*Table{userTable, orderTable}, func(transactions map[*Table]Transaction) error {
		// 获取表的事务
		userTx := transactions[userTable]
		orderTx := transactions[orderTable]

		// 插入用户
		user := map[string]any{"name": "张三"}
		userID, err := userTx.Insert(&user)
		if err != nil {
			return err
		}
		fmt.Printf("Inserted user with ID: %d\n", userID)

		// 插入订单
		order := map[string]any{"user_id": userID, "amount": 100.50}
		orderID, err := orderTx.Insert(&order)
		if err != nil {
			return err
		}
		fmt.Printf("Inserted order with ID: %d\n", orderID)

		// 测试多表连接查询
		fmt.Println("\nTesting multi-table join in transaction")

		// 获取订单表的 user_id 映射
		orderIter, _ := orderTx.Search(&map[string]any{"id": nil})
		if orderIter == nil {
			return fmt.Errorf("Search orders failed")
		}
		userIDMap := orderIter.Map("user_id")
		fmt.Printf("User ID map from orders: %v\n", userIDMap)

		// 获取用户表的迭代器
		userIter, _ := userTx.Search(&map[string]any{"id": nil})
		if userIter == nil {
			return fmt.Errorf("Search users failed")
		}

		// 创建 AND 匹配器
		andMatcher := match.NewAND([]string{"id"}, userIDMap)

		// 设置匹配器
		userIter.SetMatch(andMatcher)

		// 获取结果
		matchedUsers := userIter.GetRecords(true)
		fmt.Printf("Found %d users with orders\n", len(matchedUsers))
		for _, u := range matchedUsers {
			fmt.Printf("  - User: %v\n", u)
		}

		return nil
	})

	if err != nil {
		t.Fatalf("WithTransaction failed: %v", err)
	}

	fmt.Println("\nMulti-table transaction test passed successfully!")
}
