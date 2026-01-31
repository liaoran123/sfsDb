package engine

import (
	"fmt"
	"testing"
	"time"

	"github.com/liaoran123/sfsDb/storage"
)

// TestTableTransactionWithSnapshot 测试事务与快照的交互
func TestTableTransactionWithSnapshot(t *testing.T) {
	// 使用唯一表名，避免测试数据累积
	tableName := fmt.Sprintf("test_transaction_snapshot_%d", time.Now().UnixNano())

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

	// 初始插入一条记录
	record1 := map[string]any{"id": 1, "name": "初始数据", "age": 20}
	_, err = table.Insert(&record1)
	if err != nil {
		t.Fatalf("初始插入记录失败: %v", err)
	}

	fmt.Println("=== 测试1：初始数据插入 ===")
	// 验证初始数据
	readRecord1 := map[string]any{"id": 1}
	result1, err := table.Read(&readRecord1)
	if err != nil {
		t.Fatalf("读取初始记录失败: %v", err)
	}
	if result1 == nil {
		t.Fatalf("初始记录不存在")
	}
	fmt.Printf("初始记录: %v\n", result1)

	// 检查当前kvStore是否为LevelDBStore类型
	levelDBStore, ok := table.kvStore.(*storage.LevelDBStore)
	if !ok {
		fmt.Println("当前存储引擎不支持快照功能，跳过快照测试")
		return
	}

	fmt.Println("\n=== 测试2：创建快照 ===")
	// 创建快照
	err = levelDBStore.SwitchToSnapshot()
	if err != nil {
		t.Fatalf("创建快照失败: %v", err)
	}
	fmt.Println("成功创建快照")

	fmt.Println("\n=== 测试3：在快照模式下读取数据 ===")
	// 在快照模式下读取数据
	snapshotResult1, err := table.Read(&readRecord1)
	if err != nil {
		t.Fatalf("快照模式读取记录失败: %v", err)
	}
	if snapshotResult1 == nil {
		t.Fatalf("快照模式下记录不存在")
	}
	fmt.Printf("快照模式读取到的记录: %v\n", snapshotResult1)

	fmt.Println("\n=== 测试4：切换回DB模式并更新数据 ===")
	// 切换回DB模式
	err = levelDBStore.SwitchToDB()
	if err != nil {
		t.Fatalf("切换回DB模式失败: %v", err)
	}
	fmt.Println("成功切换回DB模式")

	// 更新记录
	updateRecord1 := map[string]any{"id": 1, "name": "更新后数据", "age": 25}
	err = table.Update(&updateRecord1)
	if err != nil {
		t.Fatalf("更新记录失败: %v", err)
	}
	fmt.Println("成功更新记录")

	// 在DB模式下读取更新后的数据
	dbResult1, err := table.Read(&readRecord1)
	if err != nil {
		t.Fatalf("DB模式读取更新后记录失败: %v", err)
	}
	if dbResult1 == nil {
		t.Fatalf("DB模式下更新后记录不存在")
	}
	fmt.Printf("DB模式读取到的更新后记录: %v\n", dbResult1)

	fmt.Println("\n=== 测试5：切换到快照模式验证数据一致性 ===")
	// 再次切换到快照模式 - 注意：每次调用SwitchToSnapshot()都会创建一个新的快照
	// 新快照会包含更新后的数据，而不是之前的旧数据
	err = levelDBStore.SwitchToSnapshot()
	if err != nil {
		t.Fatalf("再次创建快照失败: %v", err)
	}
	fmt.Println("成功再次创建快照")

	// 在新快照模式下读取数据，应该是更新后的数据
	newSnapshotResult1, err := table.Read(&readRecord1)
	if err != nil {
		t.Fatalf("新快照模式读取记录失败: %v", err)
	}
	if newSnapshotResult1 == nil {
		t.Fatalf("新快照模式下记录不存在")
	}
	fmt.Printf("新快照模式读取到的记录: %v\n", newSnapshotResult1)

	// 验证新快照读取到的是新数据，与DB模式下读取到的数据一致
	if string(newSnapshotResult1) != string(dbResult1) {
		t.Fatalf("新快照应该包含更新后的数据，与DB模式数据不一致")
	}
	fmt.Println("验证成功：新快照包含更新后的数据，与DB模式数据一致")

	// 切换回DB模式，准备下一个测试
	err = levelDBStore.SwitchToDB()
	if err != nil {
		t.Fatalf("切换回DB模式失败: %v", err)
	}
	fmt.Println("切换回DB模式")

	fmt.Println("\n=== 测试6：手动事务与快照的交互 ===")
	// 当前已经在DB模式下，直接创建batch
	batch := table.kvStore.GetBatch()
	if batch == nil {
		t.Fatalf("获取batch失败")
	}

	// 插入一条新记录
	record2 := map[string]any{"id": 2, "name": "事务插入数据", "age": 30}
	_, err = table.Insert(&record2, batch)
	if err != nil {
		t.Fatalf("事务插入记录失败: %v", err)
	}
	fmt.Println("事务插入记录成功")

	// 切换到快照模式，此时事务未提交
	err = levelDBStore.SwitchToSnapshot()
	if err != nil {
		t.Fatalf("创建快照失败: %v", err)
	}
	fmt.Println("成功创建快照")

	// 在快照模式下读取新插入的记录，应该不存在（事务未提交）
	readRecord2 := map[string]any{"id": 2}
	snapshotResult2, err := table.Read(&readRecord2)
	if err != nil {
		t.Fatalf("快照模式读取新记录失败: %v", err)
	}
	if snapshotResult2 != nil {
		t.Fatalf("事务未提交，快照模式下不应该读取到新记录")
	}
	fmt.Println("快照模式下未读取到未提交事务的记录")

	// 切换回DB模式，提交事务
	err = levelDBStore.SwitchToDB()
	if err != nil {
		t.Fatalf("切换回DB模式失败: %v", err)
	}
	fmt.Println("成功切换回DB模式")

	// 提交事务
	err = table.kvStore.WriteBatch(batch)
	if err != nil {
		t.Fatalf("提交事务失败: %v", err)
	}
	fmt.Println("成功提交事务")

	// 在DB模式下读取新插入的记录，应该存在
	dbResult2, err := table.Read(&readRecord2)
	if err != nil {
		t.Fatalf("DB模式读取新记录失败: %v", err)
	}
	if dbResult2 == nil {
		t.Fatalf("事务提交后，DB模式下应该读取到新记录")
	}
	fmt.Printf("DB模式读取到事务提交后的记录: %v\n", dbResult2)

	// 再次创建快照，此时应该能读取到新记录
	err = levelDBStore.SwitchToSnapshot()
	if err != nil {
		t.Fatalf("再次创建快照失败: %v", err)
	}
	fmt.Println("成功再次创建快照")

	// 在新快照模式下读取新记录，应该存在
	newSnapshotResult2, err := table.Read(&readRecord2)
	if err != nil {
		t.Fatalf("新快照模式读取新记录失败: %v", err)
	}
	if newSnapshotResult2 == nil {
		t.Fatalf("事务提交后，新快照模式下应该读取到新记录")
	}
	fmt.Printf("新快照模式读取到事务提交后的记录: %v\n", newSnapshotResult2)

	fmt.Println("\n=== 测试7：切换回DB模式 ===")
	// 切换回DB模式
	err = levelDBStore.SwitchToDB()
	if err != nil {
		t.Fatalf("切换回DB模式失败: %v", err)
	}
	fmt.Println("成功切换回DB模式")

	fmt.Println("\n=== 事务与快照交互测试完成 ===")
}
