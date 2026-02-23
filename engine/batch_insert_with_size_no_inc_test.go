package engine

import (
	"fmt"
	"testing"

	"github.com/liaoran123/sfsDb/storage"
)

// TestBatchInsertWithSizeNoInc 测试BatchInsertWithSizeNoInc函数
func TestBatchInsertWithSizeNoInc(t *testing.T) {
	// 初始化数据库
	_, err := storage.OpenDefaultDb("./batch_with_size_no_inc_test_db_1")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// 创建测试表
	table, err := TableNew("test_batch_with_size_no_inc")
	if err != nil {
		t.Fatalf("TableNew failed: %v", err)
	}

	// 设置字段
	fields := map[string]any{"id": 0, "name": "", "age": 0}
	table.SetFields(fields)

	// 创建主键
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("id")
	table.CreateIndex(pk)

	// 创建测试记录（包含手动指定的ID）
	recordCount := 20
	records := make([]*map[string]any, recordCount)
	for i := 0; i < recordCount; i++ {
		records[i] = &map[string]any{
			"id":   200 + i,
			"name": fmt.Sprintf("Record%d", i),
			"age":  20 + i%30,
		}
	}

}

// TestBatchInsertWithSizeNoIncSkipVersion 测试BatchInsertWithSizeNoInc函数的skipVersion参数
func TestBatchInsertWithSizeNoIncSkipVersion(t *testing.T) {
	// 初始化数据库
	_, err := storage.OpenDefaultDb("./batch_with_size_no_inc_skip_version_test_db_1")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// 创建测试表
	table, err := TableNew("test_batch_with_size_no_inc_skip_version")
	if err != nil {
		t.Fatalf("TableNew failed: %v", err)
	}

	// 设置字段
	fields := map[string]any{"id": 0, "name": ""}
	table.SetFields(fields)

	// 创建主键
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("id")
	table.CreateIndex(pk)

	// 创建测试记录
	recordCount := 10
	records := make([]*map[string]any, recordCount)
	for i := 0; i < recordCount; i++ {
		records[i] = &map[string]any{
			"id":   300 + i,
			"name": fmt.Sprintf("Record%d", i),
		}
	}

}
