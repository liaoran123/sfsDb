package transactionANT

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/storage"
)

// TestOrderTransaction 测试订单事务的各种场景
func TestOrderTransaction(t *testing.T) {
	// 创建测试目录
	testPath := t.TempDir()

	// 创建数据库存储
	db, err := storage.GetDBManager().NewLevelDBStore(testPath, nil)
	if err != nil {
		t.Fatalf("创建数据库存储失败: %v", err)
	}
	defer db.Close()

	// 设置全局存储
	originalDB := storage.GetDBManager().GetDB()
	storage.GetDBManager().SetDB(db)
	defer func() {
		storage.GetDBManager().SetDB(originalDB)
	}()

	// 创建产品表
	productTable, err := engine.TableNew("products")
	if err != nil {
		t.Fatalf("创建产品表失败: %v", err)
	}
	productTable.SetFields(map[string]any{"id": "", "name": "", "price": 0, "stock": 0})
	err = productTable.CreatePrimaryKey("id")
	if err != nil {
		t.Fatalf("设置产品表主键失败: %v", err)
	}

	// 创建订单表
	orderTable, err := engine.TableNew("orders")
	if err != nil {
		t.Fatalf("创建订单表失败: %v", err)
	}
	orderTable.SetFields(map[string]any{"id": "", "user_id": "", "product_id": "", "quantity": 0, "amount": 0, "status": ""})
	err = orderTable.CreatePrimaryKey("id")
	if err != nil {
		t.Fatalf("设置订单表主键失败: %v", err)
	}

	// 初始化测试数据
	initTestProducts(productTable, t)

	fmt.Println("=== 测试1: 订单创建与库存管理 ===")
	// 测试订单创建与库存管理
	if err := testOrderCreationWithInventory(productTable, orderTable, t); err != nil {
		t.Errorf("订单创建与库存管理测试失败: %v", err)
	}

	fmt.Println("\n=== 测试2: 订单支付与余额扣减 ===")
	// 测试订单支付与余额扣减
	if err := testOrderPayment(productTable, orderTable, t); err != nil {
		t.Errorf("订单支付与余额扣减测试失败: %v", err)
	}

	fmt.Println("\n=== 测试3: 订单取消与库存恢复 ===")
	// 测试订单取消与库存恢复
	if err := testOrderCancellation(productTable, orderTable, t); err != nil {
		t.Errorf("订单取消与库存恢复测试失败: %v", err)
	}

	fmt.Println("\n=== 测试4: 并发订单创建 ===")
	// 测试并发订单创建
	if err := testConcurrentOrderCreation(productTable, orderTable, t); err != nil {
		t.Errorf("并发订单创建测试失败: %v", err)
	}

	fmt.Println("\n=== 测试5: 库存不足场景 ===")
	// 测试库存不足场景
	if err := testInsufficientInventory(productTable, orderTable, t); err != nil {
		t.Errorf("库存不足场景测试失败: %v", err)
	}

	fmt.Println("\n=== 测试6: 订单事务重试机制测试 ===")
	// 测试订单事务重试机制
	if err := testOrderTransactionRetry(productTable, orderTable, t); err != nil {
		t.Errorf("订单事务重试测试失败: %v", err)
	}

	fmt.Println("\n=== 所有订单事务测试通过 ===")
}

// initTestProducts 初始化测试产品
func initTestProducts(table *engine.Table, t *testing.T) {
	// 创建测试产品
	products := []map[string]any{
		{"id": "1", "name": "产品A", "price": 100, "stock": 100},
		{"id": "2", "name": "产品B", "price": 200, "stock": 200},
		{"id": "3", "name": "产品C", "price": 300, "stock": 300},
	}

	// 创建一个临时 batch 用于初始化数据
	dbMgr := storage.GetDBManager()
	db := dbMgr.GetDB()
	batch := db.GetBatch()
	if batch == nil {
		t.Fatalf("创建 batch 失败")
	}

	for _, product := range products {
		_, err := table.Insert(&product, batch)
		if err != nil {
			t.Fatalf("插入测试产品失败: %v", err)
		}
	}

	// 提交 batch
	err := db.WriteBatch(batch)
	if err != nil {
		t.Fatalf("提交 batch 失败: %v", err)
	}

	// 验证初始化结果
	for _, product := range products {
		stock := getProductStock(table, product["id"].(string), t)
		if stock != product["stock"].(int) {
			t.Errorf("产品初始化失败，产品 %s 库存应为 %d，实际为 %d", product["id"], product["stock"], stock)
		}
	}

	fmt.Println("测试产品初始化完成:")
	for _, product := range products {
		fmt.Printf("产品 %s: 库存=%d\n", product["id"], getProductStock(table, product["id"].(string), t))
	}
}

// testOrderCreationWithInventory 测试订单创建与库存管理
func testOrderCreationWithInventory(productTable, orderTable *engine.Table, t *testing.T) error {
	// 获取初始库存
	initialStock := getProductStock(productTable, "1", t)
	quantity := 10

	fmt.Printf("下单前: 产品1库存=%d\n", initialStock)

	// 创建订单
	_, err := createOrder(productTable, orderTable, "1", "1", "1", quantity, t)
	if err != nil {
		return fmt.Errorf("创建订单失败: %v", err)
	}

	// 验证库存变化
	finalStock := getProductStock(productTable, "1", t)

	fmt.Printf("下单后: 产品1库存=%d\n", finalStock)

	if finalStock != initialStock-quantity {
		return fmt.Errorf("库存变化不正确，期望 %d，实际 %d", initialStock-quantity, finalStock)
	}

	fmt.Println("✓ 订单创建与库存管理测试通过")
	return nil
}

// testOrderPayment 测试订单支付与余额扣减
func testOrderPayment(productTable, orderTable *engine.Table, t *testing.T) error {
	// 创建测试账户表
	accountTable, err := engine.TableNew("accounts")
	if err != nil {
		return fmt.Errorf("创建账户表失败: %v", err)
	}
	accountTable.SetFields(map[string]any{"id": "", "balance": 0.0})
	err = accountTable.CreatePrimaryKey("id")
	if err != nil {
		return fmt.Errorf("设置账户表主键失败: %v", err)
	}

	// 初始化测试账户
	dbMgr := storage.GetDBManager()
	db := dbMgr.GetDB()
	batch := db.GetBatch()
	if batch == nil {
		return fmt.Errorf("创建 batch 失败")
	}

	account := map[string]any{"id": "1", "balance": 1000.0}
	_, err = accountTable.Insert(&account, batch)
	if err != nil {
		return fmt.Errorf("插入测试账户失败: %v", err)
	}

	// 提交 batch
	err = db.WriteBatch(batch)
	if err != nil {
		return fmt.Errorf("提交 batch 失败: %v", err)
	}

	// 创建订单
	orderID, err := createOrder(productTable, orderTable, "2", "1", "2", 2, t)
	if err != nil {
		return fmt.Errorf("创建订单失败: %v", err)
	}

	// 获取初始余额
	initialBalance := getAccountBalance(accountTable, "1", t)

	// 支付订单
	err = payOrder(orderTable, accountTable, orderID, "1", t)
	if err != nil {
		return fmt.Errorf("支付订单失败: %v", err)
	}

	// 验证余额变化
	finalBalance := getAccountBalance(accountTable, "1", t)

	fmt.Printf("支付前: 账户1余额=%.2f\n", initialBalance)
	fmt.Printf("支付后: 账户1余额=%.2f\n", finalBalance)

	if finalBalance >= initialBalance {
		return fmt.Errorf("余额没有减少")
	}

	fmt.Println("✓ 订单支付与余额扣减测试通过")
	return nil
}

// testOrderCancellation 测试订单取消与库存恢复
func testOrderCancellation(productTable, orderTable *engine.Table, t *testing.T) error {
	// 创建订单
	orderID, err := createOrder(productTable, orderTable, "3", "1", "3", 5, t)
	if err != nil {
		return fmt.Errorf("创建订单失败: %v", err)
	}

	// 获取订单后的库存
	stockAfterOrder := getProductStock(productTable, "3", t)

	// 取消订单
	err = cancelOrder(productTable, orderTable, orderID, t)
	if err != nil {
		return fmt.Errorf("取消订单失败: %v", err)
	}

	// 验证库存恢复
	stockAfterCancellation := getProductStock(productTable, "3", t)

	fmt.Printf("下单后: 产品3库存=%d\n", stockAfterOrder)
	fmt.Printf("取消后: 产品3库存=%d\n", stockAfterCancellation)

	if stockAfterCancellation <= stockAfterOrder {
		return fmt.Errorf("库存没有恢复")
	}

	fmt.Println("✓ 订单取消与库存恢复测试通过")
	return nil
}

// testConcurrentOrderCreation 测试并发订单创建
func testConcurrentOrderCreation(productTable, orderTable *engine.Table, t *testing.T) error {
	// 重置库存
	resetProductStock(productTable, "2", 100, t)

	// 获取初始库存
	initialStock := getProductStock(productTable, "2", t)
	concurrentCount := 20 // 增加并发数量，测试锁机制的有效性
	quantityPerOrder := 1

	fmt.Printf("并发下单前: 产品2库存=%d\n", initialStock)

	var wg sync.WaitGroup
	var mu sync.Mutex
	var orderErrors []error
	var orderIDs []string

	// 启动并发下单
	startTime := time.Now()
	for i := 0; i < concurrentCount; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			userID := fmt.Sprintf("%d", i+1)
			orderID := fmt.Sprintf("concurrent_%d_%d", i+1, time.Now().UnixNano()) // 增加时间戳，确保订单ID唯一

			// 尝试创建订单
			createdOrderID, err := createOrder(productTable, orderTable, "2", userID, orderID, quantityPerOrder, t)
			if err != nil {
				mu.Lock()
				orderErrors = append(orderErrors, err)
				mu.Unlock()
				fmt.Printf("创建订单失败: %v\n", err)
			} else {
				// 验证订单是否真正创建成功
				order := getOrder(orderTable, createdOrderID, t)
				if order != nil {
					mu.Lock()
					orderIDs = append(orderIDs, createdOrderID)
					mu.Unlock()
				} else {
					mu.Lock()
					orderErrors = append(orderErrors, fmt.Errorf("订单创建失败: 订单不存在"))
					mu.Unlock()
				}
			}
		}(i)
	}

	// 等待所有下单完成
	wg.Wait()
	executionTime := time.Since(startTime)

	// 验证库存变化
	finalStock := getProductStock(productTable, "2", t)
	actualReduction := initialStock - finalStock
	expectedReduction := len(orderIDs) * quantityPerOrder

	fmt.Printf("并发下单后: 产品2库存=%d (减少了: %d, 期望减少: %d)\n", finalStock, actualReduction, expectedReduction)
	fmt.Printf("并发下单次数: %d, 成功次数: %d, 失败次数: %d\n", concurrentCount, len(orderIDs), len(orderErrors))
	fmt.Printf("执行时间: %v, 每秒下单数: %.2f\n", executionTime, float64(concurrentCount)/executionTime.Seconds())

	if actualReduction != expectedReduction {
		return fmt.Errorf("库存变化不正确，期望减少 %d，实际减少 %d", expectedReduction, actualReduction)
	}

	if len(orderErrors) > concurrentCount/3 { // 允许最多33%的失败率
		return fmt.Errorf("并发下单失败率过高: %d/%d", len(orderErrors), concurrentCount)
	}

	fmt.Println("✓ 并发订单创建测试通过")
	return nil
}

// testInsufficientInventory 测试库存不足场景
func testInsufficientInventory(productTable, orderTable *engine.Table, t *testing.T) error {
	// 设置低库存
	resetProductStock(productTable, "3", 5, t)

	// 获取初始库存
	initialStock := getProductStock(productTable, "3", t)
	quantity := 10 // 超过库存

	fmt.Printf("下单前: 产品3库存=%d\n", initialStock)

	// 尝试创建订单
	_, err := createOrder(productTable, orderTable, "3", "1", "3", quantity, t)
	if err == nil {
		return fmt.Errorf("创建订单应该失败，但成功了")
	}

	// 验证库存未变化
	finalStock := getProductStock(productTable, "3", t)

	fmt.Printf("下单后: 产品3库存=%d\n", finalStock)

	if finalStock != initialStock {
		return fmt.Errorf("库存应该不变，期望 %d，实际 %d", initialStock, finalStock)
	}

	fmt.Println("✓ 库存不足场景测试通过")
	return nil
}

// 全局库存锁映射
var inventoryLocks sync.Map

// createOrder 创建订单
func createOrder(productTable, orderTable *engine.Table, productID, userID, orderID string, quantity int, t *testing.T) (string, error) {
	// 尝试多次执行，处理并发冲突
	maxRetries := 10 // 减少重试次数，增加延迟
	for i := 0; i < maxRetries; i++ {
		// 获取产品库存锁
		lockKey := fmt.Sprintf("inventory_%s", productID)
		lock, _ := inventoryLocks.LoadOrStore(lockKey, &sync.Mutex{})
		inventoryLock := lock.(*sync.Mutex)

		// 尝试获取锁
		locked := false
		for j := 0; j < 5; j++ {
			if inventoryLock.TryLock() {
				locked = true
				break
			}
			time.Sleep(50 * time.Millisecond)
		}

		if !locked {
			time.Sleep(time.Millisecond * time.Duration(i+1) * 100)
			continue
		}

		// 确保锁被释放
		defer inventoryLock.Unlock()

		// 为每个订单创建独立的batch
		dbMgr := storage.GetDBManager()
		db := dbMgr.GetDB()
		batch := db.GetBatch()
		if batch == nil {
			return "", fmt.Errorf("创建batch失败")
		}

		// 创建产品表事务（使用Serializable隔离级别）
		options := DefaultTransactionOptions()
		options.IsolationLevel = Serializable
		productTx, err := NewTableTransactionWithBatchAndOptions(productTable, batch, options, "")
		if err != nil {
			return "", fmt.Errorf("创建产品表事务失败: %v", err)
		}

		// 创建订单表事务（使用Serializable隔离级别）
		orderTx, err := NewTableTransactionWithBatchAndOptions(orderTable, batch, options, "")
		if err != nil {
			return "", fmt.Errorf("创建订单表事务失败: %v", err)
		}

		// 在事务中获取产品信息
		productSearchFields := map[string]any{"id": productID}
		txIter, err := productTx.Search(&productSearchFields)
		if err != nil {
			productTx.Rollback()
			orderTx.Rollback()
			time.Sleep(time.Millisecond * time.Duration(i+1) * 100) // 增加延迟时间
			continue
		}

		if !txIter.First() {
			productTx.Rollback()
			orderTx.Rollback()
			return "", fmt.Errorf("产品不存在: %s", productID)
		}

		// 解析产品记录
		txKey := txIter.Key()
		txValue := txIter.Value()
		txFieldsBytes := txIter.ParseBytes(txKey, txValue)
		if txFieldsBytes == nil {
			productTx.Rollback()
			orderTx.Rollback()
			time.Sleep(time.Millisecond * time.Duration(i+1) * 100) // 增加延迟时间
			continue
		}
		defer func() {
			if txFieldsBytes != nil && *txFieldsBytes != nil {
				// 注意：这里我们不能直接使用engine.GlobalFieldsBytesPool.Put，因为它是未导出的
				// 我们需要使用其他方式管理内存
			}
		}()

		// 转换为map[string]any
		txProductMap := productTable.RecordByteToAny(txFieldsBytes)
		if txProductMap == nil {
			productTx.Rollback()
			orderTx.Rollback()
			time.Sleep(time.Millisecond * time.Duration(i+1) * 100) // 增加延迟时间
			continue
		}

		// 获取库存
		currentStock, ok := (*txProductMap)["stock"]
		if !ok {
			productTx.Rollback()
			orderTx.Rollback()
			return "", fmt.Errorf("产品缺少库存字段")
		}

		stockInt, ok := currentStock.(int)
		if !ok {
			productTx.Rollback()
			orderTx.Rollback()
			return "", fmt.Errorf("库存字段类型错误")
		}

		// 检查库存是否足够
		if stockInt < quantity {
			productTx.Rollback()
			orderTx.Rollback()
			return "", fmt.Errorf("库存不足，当前库存: %d, 下单数量: %d", stockInt, quantity)
		}

		// 获取价格
		price, ok := (*txProductMap)["price"]
		if !ok {
			productTx.Rollback()
			orderTx.Rollback()
			return "", fmt.Errorf("产品缺少价格字段")
		}

		priceInt, ok := price.(int)
		if !ok {
			productTx.Rollback()
			orderTx.Rollback()
			return "", fmt.Errorf("价格字段类型错误")
		}

		// 计算订单金额
		amount := priceInt * quantity

		// 计算新库存
		newStock := stockInt - quantity

		// 更新库存
		updateFields := map[string]any{
			"id":    productID,
			"stock": newStock,
			"name":  (*txProductMap)["name"],
			"price": (*txProductMap)["price"],
		}
		err = productTx.Update(&updateFields)
		if err != nil {
			productTx.Rollback()
			orderTx.Rollback()
			time.Sleep(time.Millisecond * time.Duration(i+1) * 100) // 增加延迟时间
			continue
		}

		// 创建订单
		orderFields := map[string]any{"id": orderID, "user_id": userID, "product_id": productID, "quantity": quantity, "amount": amount, "status": "pending"}
		_, err = orderTx.Insert(&orderFields)
		if err != nil {
			productTx.Rollback()
			orderTx.Rollback()
			return "", fmt.Errorf("创建订单失败: %v", err)
		}

		// 提交所有事务
		err = productTx.Commit()
		if err != nil {
			orderTx.Rollback()
			time.Sleep(time.Millisecond * time.Duration(i+1) * 100) // 增加延迟时间
			continue
		}

		// 订单事务也需要提交
		err = orderTx.Commit()
		if err != nil {
			time.Sleep(time.Millisecond * time.Duration(i+1) * 100) // 增加延迟时间
			continue
		}

		return orderID, nil
	}

	return "", fmt.Errorf("创建订单失败: 多次尝试后仍无法完成库存更新")
}

// payOrder 支付订单
func payOrder(orderTable, accountTable *engine.Table, orderID, userID string, t *testing.T) error {
	// 获取订单信息
	orderAmount := float64(getOrderAmount(orderTable, orderID, t))

	// 开始账户表事务
	accountTx, err := NewTableTransaction(accountTable, "")
	if err != nil {
		return fmt.Errorf("开始账户事务失败: %v", err)
	}

	// 确保事务回滚
	defer func() {
		if err != nil {
			accountTx.Rollback()
		}
	}()

	// 搜索账户
	accountSearchFields := map[string]any{"id": userID}
	iter, err := accountTx.Search(&accountSearchFields)
	if err != nil {
		return fmt.Errorf("搜索账户失败: %v", err)
	}

	if !iter.First() {
		return fmt.Errorf("账户不存在: %s", userID)
	}

	// 获取账户余额
	balance := getAccountBalance(accountTable, userID, t)

	// 检查余额是否足够
	if balance < orderAmount {
		return fmt.Errorf("余额不足，当前余额: %.2f, 订单金额: %.2f", balance, orderAmount)
	}

	// 更新账户余额
	accountFields := map[string]any{"id": userID, "balance": balance - orderAmount}
	err = accountTx.Update(&accountFields)
	if err != nil {
		return fmt.Errorf("更新账户余额失败: %v", err)
	}

	// 提交账户事务
	err = accountTx.Commit()
	if err != nil {
		return fmt.Errorf("提交账户事务失败: %v", err)
	}

	// 开始订单表事务
	orderTx, err := NewTableTransaction(orderTable, "")
	if err != nil {
		return fmt.Errorf("开始订单事务失败: %v", err)
	}

	// 确保事务回滚
	defer func() {
		if err != nil {
			orderTx.Rollback()
		}
	}()

	// 更新订单状态
	orderFields := map[string]any{"id": orderID, "status": "paid"}
	err = orderTx.Update(&orderFields)
	if err != nil {
		return fmt.Errorf("更新订单状态失败: %v", err)
	}

	// 提交订单事务
	err = orderTx.Commit()
	if err != nil {
		return fmt.Errorf("提交订单事务失败: %v", err)
	}

	return nil
}

// cancelOrder 取消订单
func cancelOrder(productTable, orderTable *engine.Table, orderID string, t *testing.T) error {
	// 获取订单信息
	order := getOrder(orderTable, orderID, t)
	productID := order["product_id"].(string)
	quantity := order["quantity"].(int)

	// 开始产品表事务
	productTx, err := NewTableTransaction(productTable, "")
	if err != nil {
		return fmt.Errorf("开始产品事务失败: %v", err)
	}

	// 确保事务回滚
	defer func() {
		if err != nil {
			productTx.Rollback()
		}
	}()

	// 获取产品库存
	stock := getProductStock(productTable, productID, t)

	// 更新产品库存
	productFields := map[string]any{"id": productID, "stock": stock + quantity}
	err = productTx.Update(&productFields)
	if err != nil {
		return fmt.Errorf("更新产品库存失败: %v", err)
	}

	// 提交产品事务
	err = productTx.Commit()
	if err != nil {
		return fmt.Errorf("提交产品事务失败: %v", err)
	}

	// 开始订单表事务
	orderTx, err := NewTableTransaction(orderTable, "")
	if err != nil {
		return fmt.Errorf("开始订单事务失败: %v", err)
	}

	// 确保事务回滚
	defer func() {
		if err != nil {
			orderTx.Rollback()
		}
	}()

	// 更新订单状态
	orderFields := map[string]any{"id": orderID, "status": "cancelled"}
	err = orderTx.Update(&orderFields)
	if err != nil {
		return fmt.Errorf("更新订单状态失败: %v", err)
	}

	// 提交订单事务
	err = orderTx.Commit()
	if err != nil {
		return fmt.Errorf("提交订单事务失败: %v", err)
	}

	return nil
}

// testOrderTransactionRetry 测试订单事务重试机制
func testOrderTransactionRetry(productTable, orderTable *engine.Table, t *testing.T) error {
	// 获取初始库存
	initialStock := getProductStock(productTable, "1", t)
	quantity := 5

	fmt.Printf("重试测试前: 产品1库存=%d\n", initialStock)

	// 创建带重试配置的订单
	orderID, err := createOrderWithRetry(productTable, orderTable, "1", "1", "retry_1", quantity, t)
	if err != nil {
		return fmt.Errorf("带重试的订单创建失败: %v", err)
	}

	// 验证库存变化
	finalStock := getProductStock(productTable, "1", t)

	fmt.Printf("重试测试后: 产品1库存=%d\n", finalStock)

	if finalStock != initialStock-quantity {
		return fmt.Errorf("库存变化不正确，期望 %d，实际 %d", initialStock-quantity, finalStock)
	}

	// 创建测试账户表用于支付测试
	accountTable, err := engine.TableNew("accounts_retry")
	if err != nil {
		return fmt.Errorf("创建账户表失败: %v", err)
	}
	accountTable.SetFields(map[string]any{"id": "", "balance": 0.0})
	err = accountTable.CreatePrimaryKey("id")
	if err != nil {
		return fmt.Errorf("设置账户表主键失败: %v", err)
	}

	// 初始化测试账户
	dbMgr := storage.GetDBManager()
	db := dbMgr.GetDB()
	batch := db.GetBatch()
	if batch == nil {
		return fmt.Errorf("创建 batch 失败")
	}

	account := map[string]any{"id": "1", "balance": 1000.0}
	_, err = accountTable.Insert(&account, batch)
	if err != nil {
		return fmt.Errorf("插入测试账户失败: %v", err)
	}

	// 提交 batch
	err = db.WriteBatch(batch)
	if err != nil {
		return fmt.Errorf("提交 batch 失败: %v", err)
	}

	// 测试带重试的订单支付
	err = payOrderWithRetry(orderTable, accountTable, orderID, "1", t)
	if err != nil {
		return fmt.Errorf("带重试的订单支付失败: %v", err)
	}

	fmt.Println("✓ 订单事务重试测试通过")
	return nil
}

// createOrderWithRetry 带重试配置的订单创建操作
func createOrderWithRetry(productTable, orderTable *engine.Table, productID, userID, orderID string, quantity int, t *testing.T) (string, error) {
	// 使用自定义重试配置
	options := DefaultTransactionOptions()
	options.MaxRetries = 5
	options.InitialRetryDelay = 5 * time.Millisecond
	options.RetryBackoffFactor = 1.5

	// 开始产品表事务
	tx, err := NewTableTransactionWithOptions(productTable, options, "")
	if err != nil {
		return "", fmt.Errorf("开始事务失败: %v", err)
	}

	// 确保事务回滚
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 获取产品库存和价格（使用事务内的搜索，确保一致性）
	productSearchFields := map[string]any{"id": productID}
	iter, err := tx.Search(&productSearchFields)
	if err != nil {
		return "", fmt.Errorf("搜索产品失败: %v", err)
	}

	if !iter.First() {
		return "", fmt.Errorf("产品不存在: %s", productID)
	}

	// 解析产品记录
	key := iter.Key()
	value := iter.Value()
	fieldsBytes := iter.ParseBytes(key, value)
	if fieldsBytes == nil {
		return "", fmt.Errorf("解析产品记录失败")
	}
	defer func() {
		if fieldsBytes != nil && *fieldsBytes != nil {
			// 注意：我们无法直接访问 GlobalFieldsBytesPool，因为它是未导出的
			// 但我们可以确保在使用后不再引用它
		}
	}()

	// 转换为map[string]any
	productMap := productTable.RecordByteToAny(fieldsBytes)
	if productMap == nil {
		return "", fmt.Errorf("解析产品记录失败")
	}

	// 获取库存
	stock, ok := (*productMap)["stock"]
	if !ok {
		return "", fmt.Errorf("产品缺少库存字段")
	}

	stockInt, ok := stock.(int)
	if !ok {
		return "", fmt.Errorf("库存字段类型错误")
	}

	// 获取价格
	price, ok := (*productMap)["price"]
	if !ok {
		return "", fmt.Errorf("产品缺少价格字段")
	}

	priceInt, ok := price.(int)
	if !ok {
		return "", fmt.Errorf("价格字段类型错误")
	}

	// 检查库存是否足够
	if stockInt < quantity {
		return "", fmt.Errorf("库存不足，当前库存: %d, 下单数量: %d", stockInt, quantity)
	}

	// 计算订单金额
	amount := priceInt * quantity

	// 更新产品库存
	(*productMap)["stock"] = stockInt - quantity
	err = tx.Update(productMap)
	if err != nil {
		return "", fmt.Errorf("更新产品库存失败: %v", err)
	}

	// 提交产品事务（会自动重试）
	err = tx.Commit()
	if err != nil {
		return "", fmt.Errorf("提交产品事务失败: %v", err)
	}

	// 创建订单（使用单独的事务，避免batch nil问题）
	orderTx, err := NewTableTransaction(orderTable, "")
	if err != nil {
		return "", fmt.Errorf("开始订单事务失败: %v", err)
	}

	orderFields := map[string]any{"id": orderID, "user_id": userID, "product_id": productID, "quantity": quantity, "amount": amount, "status": "pending"}
	_, err = orderTx.Insert(&orderFields)
	if err != nil {
		orderTx.Rollback()
		return "", fmt.Errorf("创建订单失败: %v", err)
	}

	if err := orderTx.Commit(); err != nil {
		return "", fmt.Errorf("提交订单事务失败: %v", err)
	}

	return orderID, nil
}

// payOrderWithRetry 带重试配置的订单支付操作
func payOrderWithRetry(orderTable, accountTable *engine.Table, orderID, userID string, t *testing.T) error {
	// 使用自定义重试配置
	options := DefaultTransactionOptions()
	options.MaxRetries = 5
	options.InitialRetryDelay = 5 * time.Millisecond
	options.RetryBackoffFactor = 1.5

	// 开始订单表事务
	orderTx, err := NewTableTransactionWithOptions(orderTable, options, "")
	if err != nil {
		return fmt.Errorf("开始订单事务失败: %v", err)
	}

	// 确保事务回滚
	defer func() {
		if err != nil {
			orderTx.Rollback()
		}
	}()

	// 获取订单信息
	orderAmount := float64(getOrderAmount(orderTable, orderID, t))

	// 更新订单状态
	orderFields := map[string]any{"id": orderID, "status": "paid"}
	err = orderTx.Update(&orderFields)
	if err != nil {
		return fmt.Errorf("更新订单状态失败: %v", err)
	}

	// 提交订单事务（会自动重试）
	err = orderTx.Commit()
	if err != nil {
		return fmt.Errorf("提交订单事务失败: %v", err)
	}

	// 开始账户表事务
	accountTx, err := NewTableTransactionWithOptions(accountTable, options, "")
	if err != nil {
		return fmt.Errorf("开始账户事务失败: %v", err)
	}

	// 确保事务回滚
	defer func() {
		if err != nil {
			accountTx.Rollback()
		}
	}()

	// 获取账户余额
	balance := getAccountBalance(accountTable, userID, t)

	// 检查余额是否足够
	if balance < orderAmount {
		return fmt.Errorf("余额不足，当前余额: %.2f, 订单金额: %.2f", balance, orderAmount)
	}

	// 更新账户余额
	accountFields := map[string]any{"id": userID, "balance": balance - orderAmount}
	err = accountTx.Update(&accountFields)
	if err != nil {
		return fmt.Errorf("更新账户余额失败: %v", err)
	}

	// 提交账户事务（会自动重试）
	err = accountTx.Commit()
	if err != nil {
		return fmt.Errorf("提交账户事务失败: %v", err)
	}

	return nil
}

// getProductStock 获取产品库存
func getProductStock(table *engine.Table, productID string, t *testing.T) int {
	// 搜索产品
	fields := map[string]any{"id": productID}
	iter, err := table.Search(&fields)
	if err != nil {
		t.Fatalf("搜索产品失败: %v", err)
	}

	// 解析记录
	if !iter.First() {
		t.Fatalf("产品不存在: %s", productID)
	}

	key := iter.Key()
	value := iter.Value()
	fieldsBytes := iter.ParseBytes(key, value)
	if fieldsBytes == nil {
		t.Fatalf("解析记录失败: %s", productID)
	}

	// 转换为map[string]any
	anyMap := table.RecordByteToAny(fieldsBytes)
	if anyMap == nil {
		t.Fatalf("解析记录失败: %s", productID)
	}

	// 提取库存字段
	stock, ok := (*anyMap)["stock"]
	if !ok {
		t.Fatalf("产品缺少库存字段: %s", productID)
	}

	// 转换为int
	stockInt, ok := stock.(int)
	if !ok {
		stockFloat, ok := stock.(float64)
		if ok {
			stockInt = int(stockFloat)
		} else {
			t.Fatalf("库存字段类型错误: %s", productID)
		}
	}

	return stockInt
}

// getOrderAmount 获取订单金额
func getOrderAmount(table *engine.Table, orderID string, t *testing.T) int {
	// 搜索订单
	fields := map[string]any{"id": orderID}
	iter, err := table.Search(&fields)
	if err != nil {
		t.Fatalf("搜索订单失败: %v", err)
	}

	// 解析记录
	if !iter.First() {
		t.Fatalf("订单不存在: %s", orderID)
	}

	key := iter.Key()
	value := iter.Value()
	fieldsBytes := iter.ParseBytes(key, value)
	if fieldsBytes == nil {
		t.Fatalf("解析记录失败: %s", orderID)
	}

	// 转换为map[string]any
	anyMap := table.RecordByteToAny(fieldsBytes)
	if anyMap == nil {
		t.Fatalf("解析记录失败: %s", orderID)
	}

	// 提取金额字段
	amount, ok := (*anyMap)["amount"]
	if !ok {
		t.Fatalf("订单缺少金额字段: %s", orderID)
	}

	// 转换为int
	amountInt, ok := amount.(int)
	if !ok {
		amountFloat, ok := amount.(float64)
		if ok {
			amountInt = int(amountFloat)
		} else {
			t.Fatalf("金额字段类型错误: %s", orderID)
		}
	}

	return amountInt
}

// getOrder 获取订单信息
func getOrder(table *engine.Table, orderID string, t *testing.T) map[string]any {
	// 搜索订单
	fields := map[string]any{"id": orderID}
	iter, err := table.Search(&fields)
	if err != nil {
		t.Fatalf("搜索订单失败: %v", err)
	}

	// 解析记录
	if !iter.First() {
		t.Fatalf("订单不存在: %s", orderID)
	}

	key := iter.Key()
	value := iter.Value()
	fieldsBytes := iter.ParseBytes(key, value)
	if fieldsBytes == nil {
		t.Fatalf("解析记录失败: %s", orderID)
	}

	// 转换为map[string]any
	anyMap := table.RecordByteToAny(fieldsBytes)
	if anyMap == nil {
		t.Fatalf("解析记录失败: %s", orderID)
	}

	return *anyMap
}

// resetProductStock 重置产品库存
func resetProductStock(table *engine.Table, productID string, stock int, t *testing.T) {
	// 开始事务
	tx, err := NewTableTransaction(table, "")
	if err != nil {
		t.Fatalf("开始事务失败: %v", err)
	}

	// 更新库存
	fields := map[string]any{"id": productID, "stock": stock}
	err = tx.Update(&fields)
	if err != nil {
		t.Fatalf("更新产品库存失败: %v", err)
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		t.Fatalf("提交事务失败: %v", err)
	}
}

// getAccountBalance 获取账户余额
func getAccountBalance(table *engine.Table, accountID string, t *testing.T) float64 {
	// 搜索账户
	fields := map[string]any{"id": accountID}
	iter, err := table.Search(&fields)
	if err != nil {
		t.Fatalf("搜索账户失败: %v", err)
	}

	// 解析记录
	if !iter.First() {
		t.Fatalf("账户不存在: %s", accountID)
	}

	key := iter.Key()
	value := iter.Value()
	fieldsBytes := iter.ParseBytes(key, value)
	if fieldsBytes == nil {
		t.Fatalf("解析记录失败: %s", accountID)
	}

	// 转换为map[string]any
	anyMap := table.RecordByteToAny(fieldsBytes)
	if anyMap == nil {
		t.Fatalf("解析记录失败: %s", accountID)
	}

	// 提取余额字段
	balance, ok := (*anyMap)["balance"]
	if !ok {
		t.Fatalf("账户缺少余额字段: %s", accountID)
	}

	// 转换为float64
	balanceFloat, ok := balance.(float64)
	if !ok {
		balanceInt, ok := balance.(int)
		if ok {
			balanceFloat = float64(balanceInt)
		} else {
			t.Fatalf("余额字段类型错误: %s", accountID)
		}
	}

	return balanceFloat
}
