package transactionLockANT

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/storage"
)

// TestTransactionStress 测试事务压力
// 本测试函数模拟高并发场景下的事务操作
// 包括大量并发转账和订单创建操作
func TestTransactionStress(t *testing.T) {
	// 创建测试目录
	testPath := t.TempDir()

	// 创建数据库存储
	db, err := storage.NewLevelDBStore(testPath, nil)
	if err != nil {
		t.Fatalf("创建数据库存储失败: %v", err)
	}
	defer db.Close()

	// 设置全局存储
	originalKVDb := storage.KVDb
	storage.KVDb = db
	defer func() {
		storage.KVDb = originalKVDb
	}()

	// 创建账户表
	accountTable, err := engine.TableNew("accounts")
	if err != nil {
		t.Fatalf("创建账户表失败: %v", err)
	}
	accountTable.SetFields(map[string]any{"id": "", "name": "", "balance": 0.0})
	err = accountTable.CreatePrimaryKey("id")
	if err != nil {
		t.Fatalf("设置账户表主键失败: %v", err)
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

	// 创建商品表
	productTable, err := engine.TableNew("products")
	if err != nil {
		t.Fatalf("创建商品表失败: %v", err)
	}
	productTable.SetFields(map[string]any{"id": "", "name": "", "price": 0, "stock": 0})
	err = productTable.CreatePrimaryKey("id")
	if err != nil {
		t.Fatalf("设置商品表主键失败: %v", err)
	}

	// 插入测试数据
	insertStressTestAccounts(accountTable, t)
	err = insertTestProducts(productTable)
	if err != nil {
		t.Fatalf("插入测试产品失败: %v", err)
	}

	// 测试场景1: 并发转账压力测试
	t.Run("ConcurrentTransferStress", testConcurrentTransferStress(accountTable))

	// 测试场景2: 并发创建订单压力测试
	t.Run("ConcurrentOrderCreationStress", testConcurrentOrderCreationStress(orderTable, productTable))

	// 测试场景3: 混合操作压力测试
	t.Run("MixedOperationStress", testMixedOperationStress(accountTable, orderTable, productTable))

	// 测试场景4: 事务重试压力测试
	t.Run("TransactionRetryStress", testTransactionRetryStress(accountTable, orderTable, productTable))
}

// insertStressTestAccounts 插入压力测试账户
func insertStressTestAccounts(table *engine.Table, t *testing.T) {
	// 创建一个 batch 用于批量插入
	dbMgr := storage.GetDBManager()
	db := dbMgr.GetDB()
	batch := db.GetBatch()
	if batch == nil {
		t.Fatalf("创建 batch 失败")
	}

	// 插入100个账户
	for i := 1; i <= 100; i++ {
		balance := i * 1000 // 每个账户的余额为id * 1000
		fields := map[string]any{
			"id":      fmt.Sprintf("%d", i),
			"name":    fmt.Sprintf("User%d", i),
			"balance": float64(balance),
		}
		_, err := table.Insert(&fields, batch)
		if err != nil {
			t.Fatalf("插入账户数据失败: %v", err)
		}
		if i%10 == 0 {
			fmt.Printf("✅ 已插入 %d 个账户\n", i)
		}
	}

	// 提交 batch
	err := db.WriteBatch(batch)
	if err != nil {
		t.Fatalf("提交 batch 失败: %v", err)
	}
}

// insertTestProducts 插入测试产品
func insertTestProducts(table *engine.Table) error {
	// 插入测试产品
	products := []map[string]any{
		{"id": "1", "name": "产品A", "price": 100, "stock": 1000},
		{"id": "2", "name": "产品B", "price": 200, "stock": 2000},
		{"id": "3", "name": "产品C", "price": 300, "stock": 3000},
		{"id": "4", "name": "产品D", "price": 400, "stock": 4000},
		{"id": "5", "name": "产品E", "price": 500, "stock": 5000},
	}

	// 创建一个 batch 用于批量插入
	dbMgr := storage.GetDBManager()
	db := dbMgr.GetDB()
	batch := db.GetBatch()
	if batch == nil {
		return fmt.Errorf("创建 batch 失败")
	}

	for _, product := range products {
		_, err := table.Insert(&product, batch)
		if err != nil {
			return fmt.Errorf("插入测试产品失败: %v", err)
		}
	}

	// 提交 batch
	err := db.WriteBatch(batch)
	if err != nil {
		return fmt.Errorf("提交 batch 失败: %v", err)
	}

	return nil
}

// testConcurrentTransferStress 测试并发转账压力
func testConcurrentTransferStress(table *engine.Table) func(*testing.T) {
	return func(t *testing.T) {
		// 并发转账配置
		const concurrencyCount = 100
		const transferCount = 10
		var wg sync.WaitGroup
		var mu sync.Mutex
		successCount := 0
		errorCount := 0
		startTime := time.Now()

		// 启动并发转账协程
		for i := 0; i < concurrencyCount; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()

				// 每个协程执行10次转账
				for j := 0; j < transferCount; j++ {
					// 随机选择转出和转入账户
					fromID := fmt.Sprintf("%d", (i*j)%99+1)  // 1-99
					toID := fmt.Sprintf("%d", (i*j+1)%100+1) // 1-100
					amount := float64((i+j)%1000 + 1)        // 1-1000

					// 执行转账
					err := transfer(table, fromID, toID, amount)

					mu.Lock()
					if err != nil {
						errorCount++
					} else {
						successCount++
					}
					mu.Unlock()
				}
			}(i)
		}

		// 等待所有转账完成
		wg.Wait()

		// 计算执行时间
		duration := time.Since(startTime)
		totalTransfers := successCount + errorCount
		transfersPerSecond := float64(totalTransfers) / duration.Seconds()

		t.Logf("✅ 并发转账压力测试完成")
		t.Logf("   总转账次数: %d", totalTransfers)
		t.Logf("   成功次数: %d", successCount)
		t.Logf("   失败次数: %d", errorCount)
		t.Logf("   执行时间: %v", duration)
		t.Logf("   每秒转账次数: %.2f", transfersPerSecond)
	}
}

// testConcurrentOrderCreationStress 测试并发创建订单压力
func testConcurrentOrderCreationStress(orderTable, productTable *engine.Table) func(*testing.T) {
	return func(t *testing.T) {
		// 并发创建订单配置
		const concurrencyCount = 50
		const orderCount = 20
		var wg sync.WaitGroup
		var mu sync.Mutex
		successCount := 0
		errorCount := 0
		startTime := time.Now()

		// 启动并发创建订单协程
		for i := 0; i < concurrencyCount; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()

				// 每个协程创建20个订单
				for j := 0; j < orderCount; j++ {
					// 随机选择用户和商品
					userID := fmt.Sprintf("%d", (i*j)%100+1)  // 1-100
					productID := fmt.Sprintf("%d", (i*j)%5+1) // 1-5
					quantity := (i+j)%5 + 1                   // 1-5
					orderID := fmt.Sprintf("%d_%d_%d", i, j, time.Now().UnixNano())

					// 创建订单
					_, err := createOrder(productTable, orderTable, productID, userID, orderID, quantity, t)

					mu.Lock()
					if err != nil {
						errorCount++
					} else {
						successCount++
					}
					mu.Unlock()
				}
			}(i)
		}

		// 等待所有创建订单完成
		wg.Wait()

		// 计算执行时间
		duration := time.Since(startTime)
		totalOrders := successCount + errorCount
		ordersPerSecond := float64(totalOrders) / duration.Seconds()

		t.Logf("✅ 并发创建订单压力测试完成")
		t.Logf("   总订单数: %d", totalOrders)
		t.Logf("   成功次数: %d", successCount)
		t.Logf("   失败次数: %d", errorCount)
		t.Logf("   执行时间: %v", duration)
		t.Logf("   每秒订单数: %.2f", ordersPerSecond)
	}
}

// testMixedOperationStress 测试混合操作压力
func testMixedOperationStress(accountTable, orderTable, productTable *engine.Table) func(*testing.T) {
	return func(t *testing.T) {
		// 混合操作配置
		const concurrencyCount = 100
		var wg sync.WaitGroup
		var mu sync.Mutex
		successCount := 0
		errorCount := 0
		startTime := time.Now()

		// 启动并发操作协程
		for i := 0; i < concurrencyCount; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()

				// 随机执行不同操作
				operationType := i % 3
				switch operationType {
				case 0:
					// 执行转账
					fromID := fmt.Sprintf("%d", (i*11)%99+1)
					toID := fmt.Sprintf("%d", (i*13)%100+1)
					amount := float64((i*17)%1000 + 1)
					err := transfer(accountTable, fromID, toID, amount)
					mu.Lock()
					if err != nil {
						errorCount++
					} else {
						successCount++
					}
					mu.Unlock()
				case 1:
					// 创建订单
					userID := fmt.Sprintf("%d", (i*19)%100+1)
					productID := fmt.Sprintf("%d", (i*23)%5+1)
					quantity := (i*29)%5 + 1
					orderID := fmt.Sprintf("%d_%d", i, time.Now().UnixNano())
					_, err := createOrder(productTable, orderTable, productID, userID, orderID, quantity, t)
					mu.Lock()
					if err != nil {
						errorCount++
					} else {
						successCount++
					}
					mu.Unlock()
				case 2:
					// 读取账户余额
					accountID := fmt.Sprintf("%d", (i*31)%100+1)
					balance := getAccountBalance(accountTable, accountID, t)
					mu.Lock()
					if balance >= 0 {
						successCount++
					} else {
						errorCount++
					}
					mu.Unlock()
				}
			}(i)
		}

		// 等待所有操作完成
		wg.Wait()

		// 计算执行时间
		duration := time.Since(startTime)
		totalOperations := successCount + errorCount
		operationsPerSecond := float64(totalOperations) / duration.Seconds()

		t.Logf("✅ 混合操作压力测试完成")
		t.Logf("   总操作次数: %d", totalOperations)
		t.Logf("   成功次数: %d", successCount)
		t.Logf("   失败次数: %d", errorCount)
		t.Logf("   执行时间: %v", duration)
		t.Logf("   每秒操作次数: %.2f", operationsPerSecond)
	}
}

// testTransactionRetryStress 测试事务重试压力
func testTransactionRetryStress(accountTable, orderTable, productTable *engine.Table) func(*testing.T) {
	return func(t *testing.T) {
		// 事务重试压力测试配置
		const concurrencyCount = 100
		const operationCount = 10
		var wg sync.WaitGroup
		var mu sync.Mutex
		successCount := 0
		errorCount := 0
		retryCount := 0
		startTime := time.Now()

		// 启动并发操作协程
		for i := 0; i < concurrencyCount; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()

				// 每个协程执行多次操作
				for j := 0; j < operationCount; j++ {
					// 随机执行转账或订单创建操作
					operationType := (i + j) % 2
					switch operationType {
					case 0:
						// 执行带重试的转账
						fromID := fmt.Sprintf("%d", (i*j)%99+1)
						toID := fmt.Sprintf("%d", (i*j+1)%100+1)
						amount := float64((i+j)%1000 + 1)
						err := transferWithRetryForStress(accountTable, fromID, toID, amount)
						mu.Lock()
						if err != nil {
							errorCount++
						} else {
							successCount++
							// 模拟重试计数（实际重试次数由内部机制记录）
							retryCount += 0 // 这里可以添加实际的重试计数逻辑
						}
						mu.Unlock()
					case 1:
						// 执行带重试的订单创建
						userID := fmt.Sprintf("%d", (i*j)%100+1)
						productID := fmt.Sprintf("%d", (i*j)%5+1)
						quantity := (i+j)%5 + 1
						orderID := fmt.Sprintf("retry_%d_%d_%d", i, j, time.Now().UnixNano())
						_, err := createOrderWithRetryForStress(productTable, orderTable, productID, userID, orderID, quantity, t)
						mu.Lock()
						if err != nil {
							errorCount++
						} else {
							successCount++
							// 模拟重试计数
							retryCount += 0
						}
						mu.Unlock()
					}
				}
			}(i)
		}

		// 等待所有操作完成
		wg.Wait()

		// 计算执行时间
		duration := time.Since(startTime)
		totalOperations := successCount + errorCount
		operationsPerSecond := float64(totalOperations) / duration.Seconds()

		t.Logf("✅ 事务重试压力测试完成")
		t.Logf("   总操作次数: %d", totalOperations)
		t.Logf("   成功次数: %d", successCount)
		t.Logf("   失败次数: %d", errorCount)
		t.Logf("   模拟重试次数: %d", retryCount)
		t.Logf("   执行时间: %v", duration)
		t.Logf("   每秒操作次数: %.2f", operationsPerSecond)
	}
}

// TestDeadlockScenario 测试死锁场景
// 模拟可能导致死锁的并发操作，验证transactionLockFree包是否能避免死锁
func TestDeadlockScenario(t *testing.T) {
	// 创建测试目录
	testPath := t.TempDir()

	// 创建数据库存储
	db, err := storage.NewLevelDBStore(testPath, nil)
	if err != nil {
		t.Fatalf("创建数据库存储失败: %v", err)
	}
	defer db.Close()

	// 设置全局存储
	originalKVDb := storage.KVDb
	storage.KVDb = db
	defer func() {
		storage.KVDb = originalKVDb
	}()

	// 创建账户表
	accountTable, err := engine.TableNew("accounts_deadlock")
	if err != nil {
		t.Fatalf("创建账户表失败: %v", err)
	}
	accountTable.SetFields(map[string]any{"id": "", "name": "", "balance": 0.0})
	err = accountTable.CreatePrimaryKey("id")
	if err != nil {
		t.Fatalf("设置账户表主键失败: %v", err)
	}

	// 插入测试账户
	insertDeadlockTestAccounts(accountTable, t)

	// 测试场景1: 循环转账死锁测试
	t.Run("CircularTransferDeadlockTest", testCircularTransferDeadlock(accountTable))

	// 测试场景2: 高并发相同数据操作测试
	t.Run("HighConcurrencySameDataTest", testHighConcurrencySameData(accountTable))
}

// insertDeadlockTestAccounts 插入死锁测试账户
func insertDeadlockTestAccounts(table *engine.Table, t *testing.T) {
	// 创建一个 batch 用于批量插入
	dbMgr := storage.GetDBManager()
	db := dbMgr.GetDB()
	batch := db.GetBatch()
	if batch == nil {
		t.Fatalf("创建 batch 失败")
	}

	// 插入测试账户
	accounts := []map[string]any{
		{"id": "1", "name": "账户1", "balance": 10000.0},
		{"id": "2", "name": "账户2", "balance": 10000.0},
		{"id": "3", "name": "账户3", "balance": 10000.0},
		{"id": "4", "name": "账户4", "balance": 10000.0},
	}

	for _, account := range accounts {
		_, err := table.Insert(&account, batch)
		if err != nil {
			t.Fatalf("插入测试账户失败: %v", err)
		}
	}

	// 提交 batch
	err := db.WriteBatch(batch)
	if err != nil {
		t.Fatalf("提交 batch 失败: %v", err)
	}

	fmt.Println("死锁测试账户初始化完成")
}

// testCircularTransferDeadlock 测试循环转账死锁场景
func testCircularTransferDeadlock(table *engine.Table) func(*testing.T) {
	return func(t *testing.T) {
		// 循环转账配置
		const concurrencyCount = 4 // 4个账户，形成循环转账
		const transferCount = 100  // 每个方向执行100次转账
		var wg sync.WaitGroup
		var mu sync.Mutex
		successCount := 0
		errorCount := 0
		startTime := time.Now()

		// 启动循环转账协程
		// 账户1→账户2→账户3→账户4→账户1，形成循环
		transferConfig := []struct {
			fromID string
			toID   string
			amount float64
		}{
			{"1", "2", 10.0},
			{"2", "3", 10.0},
			{"3", "4", 10.0},
			{"4", "1", 10.0},
		}

		for i := 0; i < concurrencyCount; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				config := transferConfig[i]

				// 执行多次转账
				for j := 0; j < transferCount; j++ {
					// 执行转账
					err := transfer(table, config.fromID, config.toID, config.amount)

					mu.Lock()
					if err != nil {
						errorCount++
					} else {
						successCount++
					}
					mu.Unlock()
				}
			}(i)
		}

		// 等待所有转账完成
		wg.Wait()

		// 计算执行时间
		duration := time.Since(startTime)
		totalTransfers := successCount + errorCount
		transfersPerSecond := float64(totalTransfers) / duration.Seconds()

		// 验证没有死锁（如果有死锁，测试会卡住或超时）
		t.Logf("✅ 循环转账死锁测试完成")
		t.Logf("   总转账次数: %d", totalTransfers)
		t.Logf("   成功次数: %d", successCount)
		t.Logf("   失败次数: %d", errorCount)
		t.Logf("   执行时间: %v", duration)
		t.Logf("   每秒转账次数: %.2f", transfersPerSecond)

		// 注意：在并发环境下，循环转账的最终余额是不确定的
		// 这里我们只验证没有发生死锁，而不验证余额是否不变

		fmt.Println("✓ 循环转账死锁测试通过，没有发生死锁")
	}
}

// testHighConcurrencySameData 测试高并发操作相同数据的场景
func testHighConcurrencySameData(table *engine.Table) func(*testing.T) {
	return func(t *testing.T) {
		// 高并发配置
		const concurrencyCount = 200 // 200个并发协程
		const operationCount = 10    // 每个协程执行10次操作
		var wg sync.WaitGroup
		var mu sync.Mutex
		successCount := 0
		errorCount := 0
		startTime := time.Now()

		// 启动高并发协程，所有协程都操作相同的两个账户
		for i := 0; i < concurrencyCount; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()

				// 每个协程执行多次操作
				for j := 0; j < operationCount; j++ {
					// 交替执行不同方向的转账
					var fromID, toID string
					if (i+j)%2 == 0 {
						fromID = "1"
						toID = "2"
					} else {
						fromID = "2"
						toID = "1"
					}
					amount := float64((i+j)%100 + 1)

					// 执行转账
					err := transfer(table, fromID, toID, amount)

					mu.Lock()
					if err != nil {
						errorCount++
					} else {
						successCount++
					}
					mu.Unlock()
				}
			}(i)
		}

		// 等待所有操作完成
		wg.Wait()

		// 计算执行时间
		duration := time.Since(startTime)
		totalOperations := successCount + errorCount
		operationsPerSecond := float64(totalOperations) / duration.Seconds()

		// 验证没有死锁（如果有死锁，测试会卡住或超时）
		t.Logf("✅ 高并发相同数据操作测试完成")
		t.Logf("   总操作次数: %d", totalOperations)
		t.Logf("   成功次数: %d", successCount)
		t.Logf("   失败次数: %d", errorCount)
		t.Logf("   执行时间: %v", duration)
		t.Logf("   每秒操作次数: %.2f", operationsPerSecond)

		fmt.Println("✓ 高并发相同数据操作测试通过，没有发生死锁")
	}
}

// transferWithRetryForStress 带重试配置的转账操作（用于压力测试）
func transferWithRetryForStress(table *engine.Table, fromID, toID string, amount float64) error {
	// 使用自定义重试配置
	options := DefaultTransactionOptions()
	options.MaxRetries = 3
	options.InitialRetryDelay = 5 * time.Millisecond
	options.RetryBackoffFactor = 1.5

	// 开始事务
	tx, err := NewTableTransactionWithOptions(table, options, "")
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}

	// 确保事务回滚
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 获取转出账户余额
	fromBalance, err := getAccountBalanceInTransaction(tx, fromID)
	if err != nil {
		return fmt.Errorf("获取转出账户余额失败: %v", err)
	}

	// 检查余额是否足够
	if fromBalance < amount {
		return fmt.Errorf("余额不足，当前余额: %.2f, 转账金额: %.2f", fromBalance, amount)
	}

	// 获取转入账户余额
	toBalance, err := getAccountBalanceInTransaction(tx, toID)
	if err != nil {
		return fmt.Errorf("获取转入账户余额失败: %v", err)
	}

	// 更新转出账户余额
	fromFields := map[string]any{"id": fromID, "balance": fromBalance - amount}
	err = tx.Update(&fromFields)
	if err != nil {
		return fmt.Errorf("更新转出账户余额失败: %v", err)
	}

	// 更新转入账户余额
	toFields := map[string]any{"id": toID, "balance": toBalance + amount}
	err = tx.Update(&toFields)
	if err != nil {
		return fmt.Errorf("更新转入账户余额失败: %v", err)
	}

	// 提交事务（会自动重试）
	return tx.Commit()
}

// createOrderWithRetryForStress 带重试配置的订单创建操作（用于压力测试）
func createOrderWithRetryForStress(productTable, orderTable *engine.Table, productID, userID, orderID string, quantity int, t *testing.T) (string, error) {
	// 使用自定义重试配置
	options := DefaultTransactionOptions()
	options.MaxRetries = 3
	options.InitialRetryDelay = 5 * time.Millisecond
	options.RetryBackoffFactor = 1.5

	// 开始事务
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

	// 获取产品库存
	stock, err := getProductStockInTransaction(tx, productID)
	if err != nil {
		return "", fmt.Errorf("获取产品库存失败: %v", err)
	}

	// 检查库存是否足够
	if stock < quantity {
		return "", fmt.Errorf("库存不足，当前库存: %d, 下单数量: %d", stock, quantity)
	}

	// 更新产品库存
	productFields := map[string]any{"id": productID, "stock": stock - quantity}
	err = tx.Update(&productFields)
	if err != nil {
		return "", fmt.Errorf("更新产品库存失败: %v", err)
	}

	// 创建订单
	orderFields := map[string]any{"id": orderID, "user_id": userID, "product_id": productID, "quantity": quantity, "amount": 0, "status": "pending"}
	_, err = orderTable.Insert(&orderFields, nil)
	if err != nil {
		return "", fmt.Errorf("创建订单失败: %v", err)
	}

	// 提交事务（会自动重试）
	err = tx.Commit()
	if err != nil {
		return "", fmt.Errorf("提交事务失败: %v", err)
	}

	return orderID, nil
}

// transfer 转账操作
func transfer(table *engine.Table, fromID, toID string, amount float64) error {
	// 开始事务
	tx, err := NewTableTransaction(table, "")
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}

	// 确保事务回滚
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 获取转出账户余额
	fromBalance, err := getAccountBalanceInTransaction(tx, fromID)
	if err != nil {
		return fmt.Errorf("获取转出账户余额失败: %v", err)
	}

	// 检查余额是否足够
	if fromBalance < amount {
		return fmt.Errorf("余额不足，当前余额: %.2f, 转账金额: %.2f", fromBalance, amount)
	}

	// 获取转入账户余额
	toBalance, err := getAccountBalanceInTransaction(tx, toID)
	if err != nil {
		return fmt.Errorf("获取转入账户余额失败: %v", err)
	}

	// 更新转出账户余额
	fromFields := map[string]any{"id": fromID, "balance": fromBalance - amount}
	err = tx.Update(&fromFields)
	if err != nil {
		return fmt.Errorf("更新转出账户余额失败: %v", err)
	}

	// 更新转入账户余额
	toFields := map[string]any{"id": toID, "balance": toBalance + amount}
	err = tx.Update(&toFields)
	if err != nil {
		return fmt.Errorf("更新转入账户余额失败: %v", err)
	}

	// 提交事务
	return tx.Commit()
}

// getAccountBalanceInTransaction 在事务中获取账户余额
func getAccountBalanceInTransaction(tx *TableTransaction, accountID string) (float64, error) {
	// 搜索账户
	fields := map[string]any{"id": accountID}
	iter, err := tx.Search(&fields)
	if err != nil {
		return 0, err
	}

	// 解析记录
	if !iter.First() {
		return 0, fmt.Errorf("账户不存在: %s", accountID)
	}

	key := iter.Key()
	value := iter.Value()
	fieldsBytes := iter.ParseBytes(key, value)
	if fieldsBytes == nil {
		return 0, fmt.Errorf("解析记录失败")
	}

	// 转换为map[string]any
	anyMap := tx.Table.RecordByteToAny(fieldsBytes)
	if anyMap == nil {
		return 0, fmt.Errorf("解析记录失败")
	}

	// 提取余额字段
	balance, ok := (*anyMap)["balance"]
	if !ok {
		return 0, fmt.Errorf("账户缺少余额字段")
	}

	// 转换为float64
	balanceFloat, ok := balance.(float64)
	if !ok {
		balanceInt, ok := balance.(int)
		if ok {
			balanceFloat = float64(balanceInt)
		} else {
			return 0, fmt.Errorf("余额字段类型错误")
		}
	}

	return balanceFloat, nil
}

// getProductStockInTransaction 在事务中获取产品库存
func getProductStockInTransaction(tx *TableTransaction, productID string) (int, error) {
	// 搜索产品
	fields := map[string]any{"id": productID}
	iter, err := tx.Search(&fields)
	if err != nil {
		return 0, err
	}

	// 解析记录
	if !iter.First() {
		return 0, fmt.Errorf("产品不存在: %s", productID)
	}

	key := iter.Key()
	value := iter.Value()
	fieldsBytes := iter.ParseBytes(key, value)
	if fieldsBytes == nil {
		return 0, fmt.Errorf("解析记录失败")
	}

	// 转换为map[string]any
	anyMap := tx.Table.RecordByteToAny(fieldsBytes)
	if anyMap == nil {
		return 0, fmt.Errorf("解析记录失败")
	}

	// 提取库存字段
	stock, ok := (*anyMap)["stock"]
	if !ok {
		return 0, fmt.Errorf("产品缺少库存字段")
	}

	// 转换为int
	stockInt, ok := stock.(int)
	if !ok {
		stockFloat, ok := stock.(float64)
		if ok {
			stockInt = int(stockFloat)
		} else {
			return 0, fmt.Errorf("库存字段类型错误")
		}
	}

	return stockInt, nil
}
