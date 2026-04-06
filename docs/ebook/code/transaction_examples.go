package main

import (
	"fmt"
	"log"
	"time"

	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/record"
	"github.com/liaoran123/sfsDb/storage"
	"github.com/liaoran123/sfsDb/transactionLockFree"
)

func main1() {
	fmt.Println("=== sfsDb 事务处理示例 ===\n")

	// 初始化数据库
	dbManager := storage.GetDBManager()
	_, err := dbManager.OpenDB("./transaction_example_db")
	if err != nil {
		log.Fatalf("打开数据库失败: %v", err)
	}
	defer dbManager.CloseDB()

	// 示例1: 基本事务操作
	fmt.Println("--- 示例1: 基本事务操作 ---")
	basicTransactionExample()

	// 示例2: 批量操作
	fmt.Println("\n--- 示例2: 批量操作 ---")
	batchOperationExample()

	// 示例3: 银行转账
	fmt.Println("\n--- 示例3: 银行转账 ---")
	bankTransferExample()

	// 示例4: 订单处理
	fmt.Println("\n--- 示例4: 订单处理 ---")
	orderProcessingExample()

	fmt.Println("\n=== 所有示例执行完成 ===")
}

func basicTransactionExample() {
	// 创建用户表
	userTable, err := engine.NewTable("users")
	if err != nil {
		log.Printf("创建用户表失败: %v", err)
		return
	}

	userFields := map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	}
	err = userTable.SetFields(userFields)
	if err != nil {
		log.Printf("设置字段失败: %v", err)
		return
	}

	primaryKey, _ := engine.NewDefaultPrimaryKey("id")
	primaryKey.AddFields("id")
	userTable.CreateIndex(primaryKey)

	// 创建表事务
	tx, err := transactionLockFree.NewTableTransaction(userTable)
	if err != nil {
		log.Fatalf("创建事务失败: %v", err)
	}
	defer tx.Rollback()

	// 插入用户
	users := []map[string]interface{}{
		{"id": 1, "name": "张三", "age": 25},
		{"id": 2, "name": "李四", "age": 30},
		{"id": 3, "name": "王五", "age": 28},
	}

	for _, user := range users {
		_, err := tx.Insert(&user)
		if err != nil {
			log.Fatalf("插入用户失败: %v", err)
		}
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		log.Fatalf("提交事务失败: %v", err)
	}

	fmt.Printf("成功插入 %d 个用户\n", len(users))
}

func batchOperationExample() {
	// 创建账户表
	accountTable, _ := engine.TableNew("accounts")
	accountFields := map[string]any{
		"id":      0,
		"name":    "",
		"balance": 0.0,
	}
	accountTable.SetFields(accountFields)
	primaryKey, _ := engine.NewDefaultPrimaryKey("id")
	primaryKey.AddFields("id")
	accountTable.CreateIndex(primaryKey)

	// 先初始化一些账户
	initTx, _ := transactionLockFree.NewTableTransaction(accountTable)
	defer initTx.Rollback()

	accounts := []map[string]interface{}{
		{"id": 1, "name": "账户A", "balance": 1000.0},
		{"id": 2, "name": "账户B", "balance": 2000.0},
		{"id": 3, "name": "账户C", "balance": 3000.0},
	}

	for _, acc := range accounts {
		_, err := initTx.Insert(&acc)
		if err != nil {
			log.Fatal(err)
		}
	}

	err := initTx.Commit()
	if err != nil {
		log.Fatal(err)
	}

	// 批量更新
	tx, err := transactionLockFree.NewTableTransaction(accountTable)
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	updates := []map[string]interface{}{
		{"id": 1, "balance": 1500.0},
		{"id": 2, "balance": 2500.0},
		{"id": 3, "balance": 3500.0},
	}

	for _, update := range updates {
		err = tx.Update(&update)
		if err != nil {
			log.Fatalf("更新失败: %v", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("批量更新账户余额成功")
}

func bankTransferExample() {
	// 创建账户表
	accountTable, _ := engine.TableNew("bank_accounts")
	accountFields := map[string]any{
		"id":      0,
		"name":    "",
		"balance": 0.0,
	}
	accountTable.SetFields(accountFields)
	primaryKey, _ := engine.NewDefaultPrimaryKey("id")
	primaryKey.AddFields("id")
	accountTable.CreateIndex(primaryKey)

	// 创建测试账户
	initTx, _ := transactionLockFree.NewTableTransaction(accountTable)
	testAccounts := []map[string]interface{}{
		{"id": 101, "name": "张三账户", "balance": 5000.0},
		{"id": 102, "name": "李四账户", "balance": 3000.0},
	}
	for _, acc := range testAccounts {
		initTx.Insert(&acc)
	}
	initTx.Commit()

	// 执行转账
	tx, err := transactionLockFree.NewTableTransaction(accountTable)
	if err != nil {
		log.Fatalf("创建转账事务失败: %v", err)
	}
	defer tx.Rollback()

	// 查询转出账户
	fromIter, err := accountTable.Search(&map[string]any{"id": 101})
	if err != nil {
		log.Fatalf("查询转出账户失败: %v", err)
	}
	defer engine.GlobalTableIterPool.Put(fromIter)
	fromRecords := fromIter.GetRecords(true)
	defer record.PutRecords(fromRecords)
	fmt.Printf("转出账户数据: %v\n", fromRecords[0])

	// 查询转入账户
	toIter, err := accountTable.Search(&map[string]any{"id": 102})
	if err != nil {
		log.Fatalf("查询转入账户失败: %v", err)
	}
	defer engine.GlobalTableIterPool.Put(toIter)
	toRecords := toIter.GetRecords(true)
	defer record.PutRecords(toRecords)
	fmt.Printf("转入账户数据: %v\n", toRecords[0])

	// 简化演示：直接更新余额
	fromUpdate := map[string]interface{}{"id": 101, "balance": 4500.0}
	err = tx.Update(&fromUpdate)
	if err != nil {
		log.Fatalf("扣减余额失败: %v", err)
	}

	toUpdate := map[string]interface{}{"id": 102, "balance": 3500.0}
	err = tx.Update(&toUpdate)
	if err != nil {
		log.Fatalf("增加余额失败: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		log.Fatalf("提交转账事务失败: %v", err)
	}

	fmt.Println("转账成功: 101 -> 102, 金额: 500.0")
}

func orderProcessingExample() {
	// 创建订单表
	orderTable, _ := engine.TableNew("orders")
	orderFields := map[string]any{
		"id":         0,
		"user_id":    0,
		"product_id": 0,
		"quantity":   0,
		"status":     "",
		"created_at": int64(0),
	}
	orderTable.SetFields(orderFields)
	orderKey, _ := engine.NewDefaultPrimaryKey("id")
	orderKey.AddFields("id")
	orderTable.CreateIndex(orderKey)

	// 创建库存表
	inventoryTable, _ := engine.TableNew("inventory")
	inventoryFields := map[string]any{
		"product_id":   0,
		"product_name": "",
		"stock":        0,
	}
	inventoryTable.SetFields(inventoryFields)
	inventoryKey, _ := engine.NewDefaultPrimaryKey("product_id")
	inventoryKey.AddFields("product_id")
	inventoryTable.CreateIndex(inventoryKey)

	// 初始化库存
	initInventoryTx, _ := transactionLockFree.NewTableTransaction(inventoryTable)
	initInventoryTx.Insert(&map[string]interface{}{
		"product_id":   1,
		"product_name": "手机",
		"stock":        100,
	})
	initInventoryTx.Commit()

	// 执行订单处理 - 第一部分：处理订单
	{
		orderTx, err := transactionLockFree.NewTableTransaction(orderTable)
		if err != nil {
			log.Fatalf("创建订单事务失败: %v", err)
		}
		defer orderTx.Rollback()

		// 查询库存
		stockIter, err := inventoryTable.Search(&map[string]any{"product_id": 1})
		if err != nil {
			log.Fatalf("查询库存失败: %v", err)
		}
		defer engine.GlobalTableIterPool.Put(stockIter)
		stockRecords := stockIter.GetRecords(true)
		defer record.PutRecords(stockRecords)
		fmt.Printf("库存数据: %v\n", stockRecords[0])

		// 创建订单
		newOrderFields := map[string]interface{}{
			"id":         1001,
			"user_id":    1,
			"product_id": 1,
			"quantity":   2,
			"status":     "pending",
			"created_at": time.Now().Unix(),
		}
		_, err = orderTx.Insert(&newOrderFields)
		if err != nil {
			log.Fatalf("创建订单失败: %v", err)
		}

		err = orderTx.Commit()
		if err != nil {
			log.Fatalf("提交订单事务失败: %v", err)
		}
	}

	// 执行订单处理 - 第二部分：更新库存
	{
		inventoryTx, err := transactionLockFree.NewTableTransaction(inventoryTable)
		if err != nil {
			log.Fatalf("创建库存事务失败: %v", err)
		}
		defer inventoryTx.Rollback()

		// 扣减库存
		inventoryUpdate := map[string]interface{}{
			"product_id": 1,
			"stock":      98,
		}
		err = inventoryTx.Update(&inventoryUpdate)
		if err != nil {
			log.Fatalf("扣减库存失败: %v", err)
		}

		err = inventoryTx.Commit()
		if err != nil {
			log.Fatalf("提交库存事务失败: %v", err)
		}
	}

	fmt.Println("订单创建成功: 订单ID=1001, 商品=手机, 数量=2")
}
