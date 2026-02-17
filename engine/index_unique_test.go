package engine

import (
	"testing"

	"github.com/liaoran123/sfsDb/storage"
)

// TestDefaultNormalIndex_IsUnique 测试DefaultNormalIndex的IsUnique方法
func TestDefaultNormalIndex_IsUnique(t *testing.T) {
	// 创建一个测试表
	table, err := TableNew("test_index_unique")
	if err != nil {
		t.Fatalf("创建表失败: %v", err)
	}

	// 设置字段
	fields := map[string]any{
		"id":   1,
		"name": "test",
	}
	if err := table.SetFields(fields); err != nil {
		t.Fatalf("设置字段失败: %v", err)
	}

	// 创建一个普通索引
	index, err := DefaultNormalIndexNew("test_index")
	if err != nil {
		t.Fatalf("创建索引失败: %v", err)
	}
	index.AddFields("name")

	// 测试1: 空表时，任何值都应该返回true（唯一）
	key := []byte("test_key_1")
	isUnique := index.IsUnique(key)
	if !isUnique {
		t.Errorf("空表时，IsUnique应该返回true，但返回了false")
	}

	// 测试2: 插入一条数据后，相同的值应该返回false（不唯一）
	// 手动向存储中添加一个key来模拟已存在的情况
	testKey := []byte("test_key_exists")
	testValue := []byte("test_value")

	// 获取存储实例并添加key
	dbManager := storage.GetDBManager()
	db, err := dbManager.OpenDB("./test_index_db")
	if err != nil {
		t.Fatalf("打开数据库失败: %v", err)
	}
	defer dbManager.CloseDB()

	// 向存储中添加key
	err = db.Put(testKey, testValue)
	if err != nil {
		t.Fatalf("添加key失败: %v", err)
	}

	// 测试添加后的key是否返回false（不唯一）
	isUniqueAfterPut := index.IsUnique(testKey)
	if isUniqueAfterPut {
		t.Errorf("添加key后，IsUnique应该返回false，但返回了true")
	}

	// 清理测试数据
	defer db.Delete(testKey)

	t.Log("IsUnique方法测试完成")
}

// TestBaseIndex_IsUnique 测试BaseIndex的IsUnique方法
func TestBaseIndex_IsUnique(t *testing.T) {
	// 创建一个BaseIndex实例
	baseIndex := &BaseIndex{
		name: "test_base_index",
		id:   1,
	}

	// 测试1: 空表时，任何值都应该返回true（唯一）
	key := []byte("test_key_1")
	isUnique := baseIndex.IsUnique(key)
	if !isUnique {
		t.Errorf("空表时，IsUnique应该返回true，但返回了false")
	}

	// 测试2: 不同的key都应该返回true（唯一）
	key2 := []byte("test_key_2")
	isUnique2 := baseIndex.IsUnique(key2)
	if !isUnique2 {
		t.Errorf("新key时，IsUnique应该返回true，但返回了false")
	}

	t.Log("BaseIndex IsUnique方法测试完成")
}

// TestDefaultPrimaryKey_IsUnique 测试DefaultPrimaryKey的IsUnique方法
func TestDefaultPrimaryKey_IsUnique(t *testing.T) {
	// 创建一个测试表
	table, err := TableNew("test_primary_unique")
	if err != nil {
		t.Fatalf("创建表失败: %v", err)
	}

	// 设置字段
	fields := map[string]any{
		"id":   1,
		"name": "test",
	}
	if err := table.SetFields(fields); err != nil {
		t.Fatalf("设置字段失败: %v", err)
	}

	// 创建一个主键索引
	primaryKey, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		t.Fatalf("创建主键索引失败: %v", err)
	}
	primaryKey.AddFields("id")

	// 测试1: 空表时，任何值都应该返回true（唯一）
	key := []byte("test_key_1")
	isUnique := primaryKey.IsUnique(key)
	if !isUnique {
		t.Errorf("空表时，IsUnique应该返回true，但返回了false")
	}

	t.Log("主键索引IsUnique方法测试完成")
}
