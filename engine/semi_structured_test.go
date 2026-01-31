package engine

import (
	"fmt"
	"testing"

	"github.com/liaoran123/sfsDb/storage"
)

// 测试半结构化数据支持
func TestSemiStructuredData(t *testing.T) {
	// 打开默认数据库
	_, err := storage.OpenDefaultDb("./test_kvdb")
	if err != nil {
		t.Fatalf("打开数据库失败: %v", err)
	}

	// 创建一个表
	tableName := "test_semi_structured"
	table, err := TableNew(tableName)
	if err != nil {
		t.Fatalf("创建表失败: %v", err)
	}

	// 定义表结构，包含半结构化字段
	fields := map[string]any{
		"id":   0,                // 主键
		"name": "",               // 字符串
		"age":  0,                // 整型
		"data": map[string]any{}, // 半结构化数据（map）
		"tags": []string{},       // 半结构化数据（数组）
		"nested": map[string]any{
			"level1": map[string]any{
				"level2": "value",
			},
		},
	}

	// 设置表字段
	if err := table.SetFields(fields); err != nil {
		t.Fatalf("设置表字段失败: %v", err)
	}

	// 设置主键
	if err := table.CreatePrimaryKey("id"); err != nil {
		t.Fatalf("设置主键失败: %v", err)
	}

	// 插入包含半结构化数据的记录
	record1 := map[string]any{
		"id":   1,
		"name": "张三",
		"age":  25,
		"data": map[string]any{
			"address": "北京市朝阳区",
			"phone":   "13800138000",
			"email":   "zhangsan@example.com",
		},
		"tags": []string{"user", "active", "vip"},
		"nested": map[string]any{
			"level1": map[string]any{
				"level2": "value1",
				"level3": map[string]any{
					"value":  123,
					"active": true,
				},
			},
		},
	}

	id, err := table.Insert(&record1)
	if err != nil {
		t.Fatalf("插入记录失败: %v", err)
	}
	if id != 1 {
		t.Errorf("期望id为1，实际为%d", id)
	}

	// 读取记录
	readFields := map[string]any{"id": 1}
	recordBytes, err := table.Read(&readFields)
	if err != nil {
		t.Fatalf("读取记录失败: %v", err)
	}
	if recordBytes == nil {
		t.Fatalf("记录不存在")
	}

	// 打印记录
	fmt.Printf("记录1: %v\n", recordBytes)

	// 插入另一条包含不同半结构化数据的记录
	record2 := map[string]any{
		"id":   2,
		"name": "李四",
		"age":  30,
		"data": map[string]any{
			"address": "上海市浦东新区",
			"phone":   "13900139000",
			"email":   "lisi@example.com",
			"social": map[string]any{
				"wechat": "lisi_wechat",
				"weibo":  "lisi_weibo",
			},
		},
		"tags": []string{"user", "inactive"},
		"nested": map[string]any{
			"level1": map[string]any{
				"level2": "value2",
			},
		},
	}

	id2, err := table.Insert(&record2)
	if err != nil {
		t.Fatalf("插入记录2失败: %v", err)
	}
	if id2 != 2 {
		t.Errorf("期望id为2，实际为%d", id2)
	}

	// 打印记录2
	readFields2 := map[string]any{"id": 2}
	recordBytes2, err := table.Read(&readFields2)
	if err != nil {
		t.Fatalf("读取记录2失败: %v", err)
	}
	fmt.Printf("记录2: %v\n", recordBytes2)

	// 测试更新半结构化数据
	updateRecord := map[string]any{
		"id":   1,
		"name": "张三（更新）",
		"data": map[string]any{
			"address": "北京市海淀区",
			"phone":   "13800138001",
			"email":   "zhangsan@example.com",
			"social": map[string]any{
				"wechat": "zhangsan_wechat",
			},
		},
		"tags": []string{"user", "active", "vip", "updated"},
	}

	if err := table.Update(&updateRecord); err != nil {
		t.Fatalf("更新记录失败: %v", err)
	}

	// 读取更新后的记录
	recordBytesUpdated, err := table.Read(&readFields)
	if err != nil {
		t.Fatalf("读取更新后的记录失败: %v", err)
	}
	fmt.Printf("更新后的记录1: %v\n", recordBytesUpdated)

	// 测试删除记录
	if err := table.Delete(&readFields); err != nil {
		t.Fatalf("删除记录失败: %v", err)
	}

	// 验证记录已删除
	recordBytesDeleted, err := table.Read(&readFields)
	if err != nil {
		t.Fatalf("读取删除后的记录失败: %v", err)
	}
	if recordBytesDeleted != nil {
		t.Errorf("记录应该已删除，但仍然存在")
	}

	fmt.Println("半结构化数据测试成功！")
}
