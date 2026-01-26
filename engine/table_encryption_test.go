package engine

import (
	"os"
	"testing"

	"github.com/liaoran123/sfsDb/storage"
)

// TestTableEncryption 测试结合加密存储的表操作
func TestTableEncryption(t *testing.T) {
	// 清理旧的测试数据库
	removeDir("./test_encrypted_table_db")
	defer removeDir("./test_encrypted_table_db")

	// 生成测试密钥
	masterKey := make([]byte, 32)
	for i := range masterKey {
		masterKey[i] = byte(i)
	}

	// 创建加密配置
	encryptConfig := &storage.EncryptionConfig{
		Enabled:   true,
		Algorithm: "AES-256-GCM",
		MasterKey: masterKey,
	}

	// 初始化加密的全局KVDb
	_, err := storage.OpenDefaultDbWithEncryption("./test_encrypted_table_db", encryptConfig)
	if err != nil {
		t.Fatalf("Failed to open encrypted database: %v", err)
	}
	defer storage.CloseDb()

	// 1. 创建表
	table, err := TableNew("test_encrypted_table")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 2. 设置字段
	fields := map[string]any{
		"id":    0,
		"name":  "",
		"age":   0,
		"email": "",
	}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 3. 插入记录
	record := map[string]any{
		"name":  "张三",
		"age":   30,
		"email": "zhangsan@example.com",
	}
	id, err := table.Insert(&record)
	if err != nil {
		t.Fatalf("Failed to insert record: %v", err)
	}
	t.Logf("Inserted record with ID: %d", id)

	// 4. 读取记录
	// 使用迭代器获取记录
	iter := table.ForData()
	defer iter.Release()

	// 获取所有记录
	allRecords := iter.GetRecords(true)
	if len(allRecords) == 0 {
		t.Fatal("Failed to read records: no records found")
	}

	// 查找我们插入的记录
	var readResult map[string]any
	for _, record := range allRecords {
		if record["id"] == id {
			readResult = record
			break
		}
	}

	if readResult == nil {
		t.Fatalf("Failed to find record with ID %d", id)
	}

	if readResult["name"] != "张三" {
		t.Fatalf("Failed to read correct name: got %v, want 张三", readResult["name"])
	}
	t.Logf("Read record: %v", readResult)

	// 5. 更新记录
	updateRecord := map[string]any{
		"id":   id,
		"name": "张三更新",
		"age":  31,
	}
	err = table.Update(&updateRecord)
	if err != nil {
		t.Fatalf("Failed to update record: %v", err)
	}
	t.Logf("Updated record with ID: %d", id)

	// 验证更新
	// 重新获取所有记录
	iter2 := table.ForData()
	defer iter2.Release()

	updatedRecords := iter2.GetRecords(true)
	var updatedResult map[string]any
	for _, record := range updatedRecords {
		if record["id"] == id {
			updatedResult = record
			break
		}
	}

	if updatedResult == nil {
		t.Fatalf("Failed to find updated record with ID %d", id)
	}

	if updatedResult["name"] != "张三更新" {
		t.Fatalf("Failed to read correct updated name: got %v, want 张三更新", updatedResult["name"])
	}
	if updatedResult["age"] != 31 {
		t.Fatalf("Failed to read correct updated age: got %v, want 31", updatedResult["age"])
	}
	t.Logf("Read updated record: %v", updatedResult)

	// 6. 删除记录
	// 重新定义readFields变量
	readFields := map[string]any{"id": id}
	err = table.Delete(&readFields)
	if err != nil {
		t.Fatalf("Failed to delete record: %v", err)
	}
	t.Logf("Deleted record with ID: %d", id)

	// 验证删除
	// 重新获取所有记录
	iter3 := table.ForData()
	defer iter3.Release()

	remainingRecords := iter3.GetRecords(true)
	recordFound := false
	for _, record := range remainingRecords {
		if record["id"] == id {
			recordFound = true
			break
		}
	}

	if recordFound {
		t.Fatalf("Failed to delete record: record with ID %d still exists", id)
	}
	t.Logf("Verified record deletion: record with ID %d no longer exists", id)
}

// removeDir 删除目录（用于测试清理）
func removeDir(path string) {
	os.RemoveAll(path)
}
