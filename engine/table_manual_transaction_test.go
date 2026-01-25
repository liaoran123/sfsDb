package engine

import (
	"fmt"
	"testing"
	"time"
)

// TestTableManualTransaction 测试手动控制事务，在同一个事务中执行多个操作
func TestTableManualTransaction(t *testing.T) {
	// 使用唯一表名，避免测试数据累积
	tableName := fmt.Sprintf("test_manual_transaction_%d", time.Now().UnixNano())

	// 创建表
	table, err := TableNew(tableName)
	if err != nil {
		t.Fatalf("创建表格失败: %v", err)
	}

	// 设置字段
	fields := map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	}

	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("设置字段失败: %v", err)
	}

	// 创建主键索引
	pk, err := DefaultPrimaryKeyNew("pk_id")
	if err != nil {
		t.Fatalf("创建主键失败: %v", err)
	}

	// 添加主键字段
	pk.AddFields("id")

	// 添加主键索引
	err = table.CreateIndex(pk)
	if err != nil {
		t.Fatalf("创建主键索引失败: %v", err)
	}

	// 创建batch，手动控制事务
	batch := table.kvStore.GetBatch()
	if batch == nil {
		t.Fatalf("获取batch失败")
	}

	// 测试1：在同一个事务中插入两条记录
	fmt.Println("=== 测试1：在同一个事务中插入两条记录 ===")
	record1 := map[string]any{"id": 1, "name": "张三", "age": 20}
	record2 := map[string]any{"id": 2, "name": "李四", "age": 25}

	// 插入第一条记录
	id1, err := table.Insert(&record1, batch)
	if err != nil {
		t.Fatalf("插入第一条记录失败: %v", err)
	}
	fmt.Printf("插入第一条记录成功，ID: %v\n", id1)

	// 插入第二条记录
	id2, err := table.Insert(&record2, batch)
	if err != nil {
		t.Fatalf("插入第二条记录失败: %v", err)
	}
	fmt.Printf("插入第二条记录成功，ID: %v\n", id2)

	// 提交事务
	err = table.kvStore.WriteBatch(batch)
	if err != nil {
		t.Fatalf("提交事务失败: %v", err)
	}
	fmt.Println("事务提交成功")

	// 验证记录是否插入成功
	readRecord1 := map[string]any{"id": 1}
	result1, err := table.Read(&readRecord1)
	if err != nil {
		t.Fatalf("读取第一条记录失败: %v", err)
	}
	if result1 == nil {
		t.Fatalf("第一条记录不存在")
	}
	fmt.Printf("读取第一条记录成功: %v\n", result1)

	readRecord2 := map[string]any{"id": 2}
	result2, err := table.Read(&readRecord2)
	if err != nil {
		t.Fatalf("读取第二条记录失败: %v", err)
	}
	if result2 == nil {
		t.Fatalf("第二条记录不存在")
	}
	fmt.Printf("读取第二条记录成功: %v\n", result2)

	// 测试2：在同一个事务中执行更新和删除操作
	fmt.Println("\n=== 测试2：在同一个事务中执行更新和删除操作 ===")
	batch2 := table.kvStore.GetBatch()

	// 更新记录1
	updateRecord1 := map[string]any{"id": 1, "name": "张三更新", "age": 30}
	err = table.Update(&updateRecord1, batch2)
	if err != nil {
		t.Fatalf("更新记录失败: %v", err)
	}
	fmt.Println("更新记录成功")

	// 删除记录2
	deleteRecord2 := map[string]any{"id": 2}
	err = table.Delete(&deleteRecord2, batch2)
	if err != nil {
		t.Fatalf("删除记录失败: %v", err)
	}
	fmt.Println("删除记录成功")

	// 提交事务
	err = table.kvStore.WriteBatch(batch2)
	if err != nil {
		t.Fatalf("提交事务失败: %v", err)
	}
	fmt.Println("事务提交成功")

	// 验证更新是否成功
	readUpdateRecord1 := map[string]any{"id": 1}
	resultUpdate1, err := table.Read(&readUpdateRecord1)
	if err != nil {
		t.Fatalf("读取更新后的记录失败: %v", err)
	}
	if resultUpdate1 == nil {
		t.Fatalf("更新后的记录不存在")
	}
	fmt.Printf("读取更新后的记录成功: %v\n", resultUpdate1)

	// 验证删除是否成功
	readDeleteRecord2 := map[string]any{"id": 2}
	resultDelete2, err := table.Read(&readDeleteRecord2)
	if err != nil {
		t.Fatalf("读取删除后的记录失败: %v", err)
	}
	if resultDelete2 != nil {
		t.Fatalf("删除后的记录仍然存在")
	}
	fmt.Println("验证删除成功，记录不存在")

	// 测试3：测试回滚（不提交事务）
	fmt.Println("\n=== 测试3：测试回滚（不提交事务） ===")
	batch3 := table.kvStore.GetBatch()

	// 插入一条新记录
	record3 := map[string]any{"id": 3, "name": "王五", "age": 35}
	id3, err := table.Insert(&record3, batch3)
	if err != nil {
		t.Fatalf("插入记录失败: %v", err)
	}
	fmt.Printf("插入记录成功，ID: %v\n", id3)

	// 不提交事务，模拟回滚
	fmt.Println("不提交事务，模拟回滚")
	// 不调用 WriteBatch，batch 会被垃圾回收，操作不会生效

	// 验证记录是否未插入
	readRecord3 := map[string]any{"id": 3}
	result3, err := table.Read(&readRecord3)
	if err != nil {
		t.Fatalf("读取记录失败: %v", err)
	}
	if result3 != nil {
		t.Fatalf("记录不应该存在，回滚失败")
	}
	fmt.Println("验证回滚成功，记录不存在")

	fmt.Println("\n=== 手动控制事务测试完成 ===")
}