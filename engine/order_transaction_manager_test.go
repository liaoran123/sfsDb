package engine

import (
	"testing"
	"github.com/liaoran123/sfsDb/storage"
)

// TestOrderTransactionManager_CreateOrderWithInventory 测试创建订单并管理库存
func TestOrderTransactionManager_CreateOrderWithInventory(t *testing.T) {
	// 打开默认存储
	_, err := storage.OpenDefaultDb("./test/kvdb_order")
	if err != nil {
		t.Fatalf("Failed to open store: %v", err)
	}

	// 创建产品表
	productTable, err := TableNew("products")
	if err != nil {
		t.Fatalf("Failed to create product table: %v", err)
	}

	// 设置产品表字段
	err = productTable.SetFields(map[string]any{"id": "", "name": "", "price": 0.0, "stock": 0})
	if err != nil {
		t.Fatalf("Failed to set product fields: %v", err)
	}

	// 创建订单表
	orderTable, err := TableNew("orders")
	if err != nil {
		t.Fatalf("Failed to create order table: %v", err)
	}

	// 设置订单表字段
	err = orderTable.SetFields(map[string]any{"id": "", "product_id": "", "quantity": 0, "status": ""})
	if err != nil {
		t.Fatalf("Failed to set order fields: %v", err)
	}

	// 初始化测试数据
	productData := map[string]any{"id": "1", "name": "Laptop", "price": 5000.0, "stock": 10}
	_, err = productTable.Insert(&productData)
	if err != nil {
		t.Fatalf("Failed to insert product: %v", err)
	}

	// 创建事务管理器
	batch := productTable.kvStore.GetBatch()
	manager := NewTransactionManager(batch)

	// 添加产品表到事务管理器
	productTx, err := manager.AddTable(productTable)
	if err != nil {
		t.Fatalf("Failed to add product table to transaction: %v", err)
	}

	// 添加订单表到事务管理器
	orderTx, err := manager.AddTable(orderTable)
	if err != nil {
		t.Fatalf("Failed to add order table to transaction: %v", err)
	}

	// 读取产品
	productReadFields := map[string]any{"id": "1"}
	_, err = productTx.Read(&productReadFields)
	if err != nil {
		t.Fatalf("Failed to read product: %v", err)
	}

	// 创建订单
	orderData := map[string]any{
		"id": "1",
		"product_id": "1",
		"quantity": 2,
		"status": "pending",
	}
	_, err = orderTx.Insert(&orderData)
	if err != nil {
		t.Fatalf("Failed to create order: %v", err)
	}

	// 提交事务
	if err := manager.Commit(); err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// 验证结果
	// 注意：这里简化测试，只验证事务提交成功
	t.Log("Order creation transaction committed successfully")
}

// TestOrderTransactionManager_OrderPayment 测试订单支付
func TestOrderTransactionManager_OrderPayment(t *testing.T) {
	// 打开默认存储
	_, err := storage.OpenDefaultDb("./test/kvdb_order_payment")
	if err != nil {
		t.Fatalf("Failed to open store: %v", err)
	}

	// 创建账户表
	accountTable, err := TableNew("accounts")
	if err != nil {
		t.Fatalf("Failed to create account table: %v", err)
	}

	// 设置账户表字段
	err = accountTable.SetFields(map[string]any{"id": "", "name": "", "balance": 0.0})
	if err != nil {
		t.Fatalf("Failed to set account fields: %v", err)
	}

	// 创建订单表
	orderTable, err := TableNew("orders")
	if err != nil {
		t.Fatalf("Failed to create order table: %v", err)
	}

	// 设置订单表字段
	err = orderTable.SetFields(map[string]any{"id": "", "user_id": "", "amount": 0.0, "status": ""})
	if err != nil {
		t.Fatalf("Failed to set order fields: %v", err)
	}

	// 初始化测试数据
	accountData := map[string]any{"id": "1", "name": "Alice", "balance": 10000.0}
	_, err = accountTable.Insert(&accountData)
	if err != nil {
		t.Fatalf("Failed to insert account: %v", err)
	}

	orderData := map[string]any{"id": "1", "user_id": "1", "amount": 5000.0, "status": "pending"}
	_, err = orderTable.Insert(&orderData)
	if err != nil {
		t.Fatalf("Failed to insert order: %v", err)
	}

	// 创建事务管理器
	batch := accountTable.kvStore.GetBatch()
	manager := NewTransactionManager(batch)

	// 添加账户表到事务管理器
	accountTx, err := manager.AddTable(accountTable)
	if err != nil {
		t.Fatalf("Failed to add account table to transaction: %v", err)
	}

	// 添加订单表到事务管理器
	orderTx, err := manager.AddTable(orderTable)
	if err != nil {
		t.Fatalf("Failed to add order table to transaction: %v", err)
	}

	// 读取用户账户
	accountReadFields := map[string]any{"id": "1"}
	_, err = accountTx.Read(&accountReadFields)
	if err != nil {
		t.Fatalf("Failed to read account: %v", err)
	}

	// 读取订单
	orderReadFields := map[string]any{"id": "1"}
	_, err = orderTx.Read(&orderReadFields)
	if err != nil {
		t.Fatalf("Failed to read order: %v", err)
	}

	// 提交事务
	if err := manager.Commit(); err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// 验证结果
	// 注意：这里简化测试，只验证事务提交成功
	t.Log("Order payment transaction committed successfully")
}

// TestOrderTransactionManager_OrderCancellation 测试订单取消
func TestOrderTransactionManager_OrderCancellation(t *testing.T) {
	// 打开默认存储
	_, err := storage.OpenDefaultDb("./test/kvdb_order_cancellation")
	if err != nil {
		t.Fatalf("Failed to open store: %v", err)
	}

	// 创建产品表
	productTable, err := TableNew("products")
	if err != nil {
		t.Fatalf("Failed to create product table: %v", err)
	}

	// 设置产品表字段
	err = productTable.SetFields(map[string]any{"id": "", "name": "", "stock": 0})
	if err != nil {
		t.Fatalf("Failed to set product fields: %v", err)
	}

	// 创建订单表
	orderTable, err := TableNew("orders")
	if err != nil {
		t.Fatalf("Failed to create order table: %v", err)
	}

	// 设置订单表字段
	err = orderTable.SetFields(map[string]any{"id": "", "product_id": "", "quantity": 0, "status": ""})
	if err != nil {
		t.Fatalf("Failed to set order fields: %v", err)
	}

	// 初始化测试数据
	productData := map[string]any{"id": "1", "name": "Laptop", "stock": 8}
	_, err = productTable.Insert(&productData)
	if err != nil {
		t.Fatalf("Failed to insert product: %v", err)
	}

	orderData := map[string]any{"id": "1", "product_id": "1", "quantity": 2, "status": "pending"}
	_, err = orderTable.Insert(&orderData)
	if err != nil {
		t.Fatalf("Failed to insert order: %v", err)
	}

	// 创建事务管理器
	batch := productTable.kvStore.GetBatch()
	manager := NewTransactionManager(batch)

	// 添加产品表到事务管理器
	productTx, err := manager.AddTable(productTable)
	if err != nil {
		t.Fatalf("Failed to add product table to transaction: %v", err)
	}

	// 添加订单表到事务管理器
	orderTx, err := manager.AddTable(orderTable)
	if err != nil {
		t.Fatalf("Failed to add order table to transaction: %v", err)
	}

	// 读取订单
	orderReadFields := map[string]any{"id": "1"}
	_, err = orderTx.Read(&orderReadFields)
	if err != nil {
		t.Fatalf("Failed to read order: %v", err)
	}

	// 读取产品
	productReadFields := map[string]any{"id": "1"}
	_, err = productTx.Read(&productReadFields)
	if err != nil {
		t.Fatalf("Failed to read product: %v", err)
	}

	// 提交事务
	if err := manager.Commit(); err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// 验证结果
	// 注意：这里简化测试，只验证事务提交成功
	t.Log("Order cancellation transaction committed successfully")
}
