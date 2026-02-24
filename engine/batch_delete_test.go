package engine

import (
	"fmt"
	"testing"
	"time"
)

func TestBatchDelete(t *testing.T) {
	// 创建测试表
	tableName := fmt.Sprintf("test_batch_delete_%d", time.Now().UnixNano())
	table, err := TableNew(tableName)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置字段
	fields := map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	}

	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 创建主键索引
	pk, err := DefaultPrimaryKeyNew("pk_id")
	if err != nil {
		t.Fatalf("Failed to create primary key: %v", err)
	}

	// 添加主键字段
	pk.AddFields("id")

	// 添加主键索引
	err = table.CreateIndex(pk)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// 插入测试数据
	records := []map[string]any{
		{"id": 1, "name": "Alice", "age": 25},
		{"id": 2, "name": "Bob", "age": 30},
		{"id": 3, "name": "Charlie", "age": 35},
		{"id": 4, "name": "David", "age": 40},
		{"id": 5, "name": "Eve", "age": 45},
	}

	for _, record := range records {
		_, err := table.Insert(&record)
		if err != nil {
			t.Fatalf("Failed to insert record: %v", err)
		}
	}

	// 准备批量删除的记录
	deleteRecords := []*map[string]any{
		&map[string]any{"id": 1},
		&map[string]any{"id": 3},
		&map[string]any{"id": 5},
	}

	// 执行批量删除
	deleteImpl := NewDeleteImpl(table, nil)
	if err := deleteImpl.BatchDelete(deleteRecords); err != nil {
		t.Fatalf("Failed to batch delete: %v", err)
	}

	// 验证删除结果 - 尝试读取已删除的记录应该失败
	for _, deleteRec := range deleteRecords {
		deleteImpl := NewDeleteImpl(table, deleteRec)
		err := deleteImpl.HasPrimaryKey()
		if err != nil {
			t.Fatalf("Failed to check primary key: %v", err)
		}
		err = deleteImpl.ReadRecord()
		if err == nil {
			t.Fatalf("Expected record %v to be deleted, but it still exists", deleteRec)
		}
		GlobalDeleteImplPool.Put(deleteImpl)
	}

	// 验证剩余记录仍然存在
	remainingRecords := []*map[string]any{
		&map[string]any{"id": 2},
		&map[string]any{"id": 4},
	}

	for _, remainingRec := range remainingRecords {
		deleteImpl := NewDeleteImpl(table, remainingRec)
		err := deleteImpl.HasPrimaryKey()
		if err != nil {
			t.Fatalf("Failed to check primary key: %v", err)
		}
		err = deleteImpl.ReadRecord()
		if err != nil {
			t.Fatalf("Expected record %v to exist, but it was not found: %v", remainingRec, err)
		}
		GlobalDeleteImplPool.Put(deleteImpl)
	}

	t.Log("BatchDelete test passed successfully!")
}
