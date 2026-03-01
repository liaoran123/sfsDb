package transactionANT

import (
	"testing"

	"github.com/liaoran123/sfsDb/engine"
)

// TestNewTransaction 测试创建新事务
func TestNewTransaction(t *testing.T) {
	// 创建测试表
	table, err := engine.TableNew("test_transaction")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{"id": 0, "name": "", "age": 0}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 创建事务
	tx, err := NewTableTransaction(table, "")
	if err != nil {
		t.Fatalf("Failed to create transaction: %v", err)
	}

	// 验证事务创建成功
	if tx == nil {
		t.Fatal("Transaction is nil")
	}

	if tx.GetTxID() == 0 {
		t.Error("Transaction ID should not be zero")
	}

	// 回滚事务
	err = tx.Rollback()
	if err != nil {
		t.Errorf("Failed to rollback transaction: %v", err)
	}
}

// TestTransactionInsert 测试事务插入操作
func TestTransactionInsert(t *testing.T) {
	// 创建测试表
	table, err := engine.TableNew("test_transaction_insert")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{"id": 0, "name": "", "age": 0}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 创建事务
	tx, err := NewTableTransaction(table, "")
	if err != nil {
		t.Fatalf("Failed to create transaction: %v", err)
	}

	// 插入测试数据
	insertData := map[string]interface{}{"id": 1, "name": "Test", "age": 25}
	id, err := tx.Insert(&insertData)
	if err != nil {
		t.Fatalf("Failed to insert data: %v", err)
	}

	if id == 0 {
		t.Error("Insert should return a valid ID")
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		t.Errorf("Failed to commit transaction: %v", err)
	}
}

// TestTransactionUpdate 测试事务更新操作
func TestTransactionUpdate(t *testing.T) {
	// 创建测试表
	table, err := engine.TableNew("test_transaction_update")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{"id": 0, "name": "", "age": 0}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 先插入测试数据
	insertData := map[string]interface{}{"id": 1, "name": "Test", "age": 25}
	_, err = table.Insert(&insertData)
	if err != nil {
		t.Fatalf("Failed to insert initial data: %v", err)
	}

	// 创建事务
	tx, err := NewTableTransaction(table, "")
	if err != nil {
		t.Fatalf("Failed to create transaction: %v", err)
	}

	// 更新测试数据
	updateData := map[string]interface{}{"id": 1, "name": "Updated Test", "age": 30}
	err = tx.Update(&updateData)
	if err != nil {
		t.Fatalf("Failed to update data: %v", err)
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		t.Errorf("Failed to commit transaction: %v", err)
	}
}

// TestTransactionDelete 测试事务删除操作
func TestTransactionDelete(t *testing.T) {
	// 创建测试表
	table, err := engine.TableNew("test_transaction_delete")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{"id": 0, "name": "", "age": 0}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 先插入测试数据
	insertData := map[string]interface{}{"id": 1, "name": "Test", "age": 25}
	_, err = table.Insert(&insertData)
	if err != nil {
		t.Fatalf("Failed to insert initial data: %v", err)
	}

	// 创建事务
	tx, err := NewTableTransaction(table, "")
	if err != nil {
		t.Fatalf("Failed to create transaction: %v", err)
	}

	// 删除测试数据
	deleteData := map[string]interface{}{"id": 1}
	err = tx.Delete(&deleteData)
	if err != nil {
		t.Fatalf("Failed to delete data: %v", err)
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		t.Errorf("Failed to commit transaction: %v", err)
	}
}
