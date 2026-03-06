package transactionANT

import (
	"fmt"
	"testing"

	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/storage"
)

// TestTransactionManager 测试 TransactionManager 的基本功能
func TestTransactionManager(t *testing.T) {
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
	initTestData(productTable, t)

	// 测试场景1: 基本多表事务
	t.Run("BasicMultiTableTransaction", func(t *testing.T) {
		testBasicMultiTableTransaction(productTable, orderTable, t)
	})

	// 测试场景2: WithTransaction 便捷函数
	t.Run("WithTransactionConvenience", func(t *testing.T) {
		testWithTransactionConvenience(productTable, orderTable, t)
	})

	// 测试场景3: 事务回滚
	t.Run("TransactionRollback", func(t *testing.T) {
		testTransactionRollback(productTable, orderTable, t)
	})

	// 测试场景4: 空事务
	t.Run("EmptyTransaction", func(t *testing.T) {
		testEmptyTransaction(t)
	})
}

// initTestData 初始化测试数据
func initTestData(productTable *engine.Table, t *testing.T) {
	// 创建一个临时 batch 用于初始化数据
	dbMgr := storage.GetDBManager()
	db := dbMgr.GetDB()
	batch := db.GetBatch()
	if batch == nil {
		t.Fatalf("创建 batch 失败")
	}

	// 插入测试产品
	products := []map[string]any{
		{"id": "1", "name": "产品A", "price": 100, "stock": 100},
		{"id": "2", "name": "产品B", "price": 200, "stock": 200},
	}

	for _, product := range products {
		_, err := productTable.Insert(&product, batch)
		if err != nil {
			t.Fatalf("插入测试产品失败: %v", err)
		}
	}

	// 提交 batch
	err := db.WriteBatch(batch)
	if err != nil {
		t.Fatalf("提交 batch 失败: %v", err)
	}

	fmt.Println("测试数据初始化完成")
}

// testBasicMultiTableTransaction 测试基本的多表事务
func testBasicMultiTableTransaction(productTable, orderTable *engine.Table, t *testing.T) {
	// 创建共享 batch
	dbMgr := storage.GetDBManager()
	db := dbMgr.GetDB()
	batch := db.GetBatch()
	if batch == nil {
		t.Fatalf("创建 batch 失败")
	}

	// 创建事务管理器
	manager := NewTransactionManager(batch, "")

	// 添加产品表到事务管理器
	productTx, err := manager.AddTable(productTable)
	if err != nil {
		t.Fatalf("添加产品表到事务管理器失败: %v", err)
	}

	// 添加订单表到事务管理器
	orderTx, err := manager.AddTable(orderTable)
	if err != nil {
		t.Fatalf("添加订单表到事务管理器失败: %v", err)
	}

	// 获取初始库存
	initialStock := getProductStock(productTable, "1", t)
	quantity := 5

	// 在产品表上更新库存
	productFields := map[string]any{"id": "1", "stock": initialStock - quantity}
	err = productTx.Update(&productFields)
	if err != nil {
		t.Fatalf("更新产品库存失败: %v", err)
	}

	// 在订单表上创建订单
	orderFields := map[string]any{
		"id":         "1",
		"user_id":    "1",
		"product_id": "1",
		"quantity":   quantity,
		"amount":     100 * quantity,
		"status":     "pending",
	}
	_, err = orderTx.Insert(&orderFields)
	if err != nil {
		t.Fatalf("创建订单失败: %v", err)
	}

	// 提交事务
	err = manager.Commit()
	if err != nil {
		t.Fatalf("提交事务失败: %v", err)
	}

	// 验证结果
	finalStock := getProductStock(productTable, "1", t)
	if finalStock != initialStock-quantity {
		t.Fatalf("库存变化不正确，期望 %d，实际 %d", initialStock-quantity, finalStock)
	}

	// 验证订单是否创建成功
	orderAmount := getOrderAmount(orderTable, "1", t)
	if orderAmount != 100*quantity {
		t.Fatalf("订单金额不正确，期望 %d，实际 %d", 100*quantity, orderAmount)
	}

	fmt.Println("✓ 基本多表事务测试通过")
}

// testWithTransactionConvenience 测试 WithTransaction 便捷函数
func testWithTransactionConvenience(productTable, orderTable *engine.Table, t *testing.T) {
	// 创建共享 batch
	dbMgr := storage.GetDBManager()
	db := dbMgr.GetDB()
	batch := db.GetBatch()
	if batch == nil {
		t.Fatalf("创建 batch 失败")
	}

	// 获取初始库存
	initialStock := getProductStock(productTable, "2", t)
	quantity := 3

	// 使用 WithTransaction 便捷函数
	err := WithTransaction(batch, []*engine.Table{productTable, orderTable}, "", func(txs map[*engine.Table]TableTransactionInterface) error {
		// 在产品表上更新库存
		productTx := txs[productTable]
		productFields := map[string]any{"id": "2", "stock": initialStock - quantity}
		err := productTx.Update(&productFields)
		if err != nil {
			return fmt.Errorf("更新产品库存失败: %v", err)
		}

		// 在订单表上创建订单
		orderTx := txs[orderTable]
		orderFields := map[string]any{
			"id":         "2",
			"user_id":    "1",
			"product_id": "2",
			"quantity":   quantity,
			"amount":     200 * quantity,
			"status":     "pending",
		}
		_, err = orderTx.Insert(&orderFields)
		if err != nil {
			return fmt.Errorf("创建订单失败: %v", err)
		}

		return nil
	})

	if err != nil {
		t.Fatalf("执行多表事务失败: %v", err)
	}

	// 验证结果
	finalStock := getProductStock(productTable, "2", t)
	if finalStock != initialStock-quantity {
		t.Fatalf("库存变化不正确，期望 %d，实际 %d", initialStock-quantity, finalStock)
	}

	// 验证订单是否创建成功
	orderAmount := getOrderAmount(orderTable, "2", t)
	if orderAmount != 200*quantity {
		t.Fatalf("订单金额不正确，期望 %d，实际 %d", 200*quantity, orderAmount)
	}

	fmt.Println("✓ WithTransaction 便捷函数测试通过")
}

// testTransactionRollback 测试事务回滚
func testTransactionRollback(productTable, orderTable *engine.Table, t *testing.T) {
	// 创建共享 batch
	dbMgr := storage.GetDBManager()
	db := dbMgr.GetDB()
	batch := db.GetBatch()
	if batch == nil {
		t.Fatalf("创建 batch 失败")
	}

	// 获取初始库存
	initialStock := getProductStock(productTable, "1", t)

	// 创建事务管理器
	manager := NewTransactionManager(batch, "")

	// 添加产品表到事务管理器
	productTx, err := manager.AddTable(productTable)
	if err != nil {
		t.Fatalf("添加产品表到事务管理器失败: %v", err)
	}

	// 在产品表上更新库存
	productFields := map[string]any{"id": "1", "stock": initialStock - 10}
	err = productTx.Update(&productFields)
	if err != nil {
		t.Fatalf("更新产品库存失败: %v", err)
	}

	// 回滚事务
	err = manager.Rollback()
	if err != nil {
		t.Fatalf("回滚事务失败: %v", err)
	}

	// 验证库存没有变化
	finalStock := getProductStock(productTable, "1", t)
	if finalStock != initialStock {
		t.Fatalf("库存应该没有变化，期望 %d，实际 %d", initialStock, finalStock)
	}

	fmt.Println("✓ 事务回滚测试通过")
}

// testEmptyTransaction 测试空事务
func testEmptyTransaction(t *testing.T) {
	// 创建共享 batch
	dbMgr := storage.GetDBManager()
	db := dbMgr.GetDB()
	batch := db.GetBatch()
	if batch == nil {
		t.Fatalf("创建 batch 失败")
	}

	// 创建事务管理器
	manager := NewTransactionManager(batch, "")

	// 提交空事务
	err := manager.Commit()
	if err != nil {
		t.Fatalf("提交空事务失败: %v", err)
	}

	fmt.Println("✓ 空事务测试通过")
}
