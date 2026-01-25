package test

import (
	"fmt"
	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/storage"
	"testing"
)

// TestTransactionConsistency 测试事务的读一致性
func TestTransactionConsistency(t *testing.T) {
	// 打开默认数据库
	_, err := storage.OpenDefaultDb("./test_transaction_consistency_db")
	if err != nil {
		t.Fatalf("打开数据库失败: %v", err)
	}
	defer storage.CloseDb()

	// 创建表
	table, err := engine.TableNew("test_transaction_consistency")
	if err != nil {
		t.Fatalf("创建表失败: %v", err)
	}

	// 设置表字段
	fields := map[string]any{
		"id":    0,
		"name":  "",
		"age":   0,
		"salary": 0.0,
	}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("设置表字段失败: %v", err)
	}

	// 创建主键索引
	primaryKey, err := engine.DefaultPrimaryKeyNew("pk_id")
	if err != nil {
		t.Fatalf("创建主键失败: %v", err)
	}
	primaryKey.AddFields("id")
	err = table.CreateIndex(primaryKey)
	if err != nil {
		t.Fatalf("创建主键索引失败: %v", err)
	}

	fmt.Println("表创建成功")

	// 测试1: 事务内读取自己的写操作
	fmt.Println("\n=== 测试1: 事务内读取自己的写操作 ===")
	testReadOwnWrite(table, t)

	// 测试2: 事务的读一致性（快照功能）
	fmt.Println("\n=== 测试2: 事务的读一致性（快照功能） ===")
	testSnapshotConsistency(table, t)

	// 测试3: 事务提交和回滚
	fmt.Println("\n=== 测试3: 事务提交和回滚 ===")
	testCommitRollback(table, t)

	fmt.Println("\n所有测试通过！")
}

// 测试事务内读取自己的写操作
func testReadOwnWrite(table *engine.Table, t *testing.T) {
	// 开始事务
	tx, err := table.Begin()
	if err != nil {
		t.Fatalf("开始事务失败: %v", err)
	}

	// 插入一条记录
	record1 := map[string]any{"name": "张三", "age": 30, "salary": 5000.0}
	id1, err := tx.Insert(&record1)
	if err != nil {
		t.Fatalf("插入记录失败: %v", err)
	}
	fmt.Printf("插入记录成功，ID: %d\n", id1)

	// 读取刚插入的记录
	readFields := map[string]any{"id": id1}
	recordBytes, err := tx.Read(&readFields)
	if err != nil {
		t.Fatalf("读取记录失败: %v", err)
	}

	if recordBytes == nil {
		t.Fatal("事务内无法读取自己刚插入的记录")
	}
	fmt.Println("✓ 事务内成功读取到自己刚插入的记录")

	// 提交事务
	err = tx.Commit()
	if err != nil {
		t.Fatalf("提交事务失败: %v", err)
	}
}

// 测试事务的读一致性（快照功能）
func testSnapshotConsistency(table *engine.Table, t *testing.T) {
	// 先插入一条初始记录
	initRecord := map[string]any{"name": "初始记录", "age": 25, "salary": 4000.0}
	initId, err := table.Insert(&initRecord)
	if err != nil {
		t.Fatalf("插入初始记录失败: %v", err)
	}

	// 开始事务，获取快照
	tx, err := table.Begin()
	if err != nil {
		t.Fatalf("开始事务失败: %v", err)
	}

	// 在事务外修改这条记录
	updateRecord := map[string]any{"id": initId, "name": "外部修改后的记录", "age": 26, "salary": 4500.0}
	err = table.Update(&updateRecord)
	if err != nil {
		t.Fatalf("外部修改记录失败: %v", err)
	}
	fmt.Println("事务外修改记录成功")

	// 在事务内读取这条记录，应该读取到的是快照中的初始状态
	readFields := map[string]any{"id": initId}
	iter := tx.Search(&readFields)
	if iter == nil {
		t.Fatal("搜索记录失败")
	}
	defer iter.Release()

	records := iter.GetRecords(true)
	if len(records) != 1 {
		t.Fatalf("读取记录失败，期望1条记录，实际%d条", len(records))
	}

	record := records[0]
	if record["name"] != "初始记录" {
		t.Fatalf("事务内读取到的记录与快照不一致，期望'初始记录'，实际'%s'", record["name"])
	}
	fmt.Printf("✓ 事务内读取到的记录: %s (年龄: %d, 薪资: %.2f)\n", record["name"], record["age"], record["salary"])
	fmt.Println("✓ 事务读一致性验证通过: 读取到的是事务开始时的快照数据，不受外部修改影响")

	// 提交事务
	err = tx.Commit()
	if err != nil {
		t.Fatalf("提交事务失败: %v", err)
	}
}

// 测试事务提交和回滚
func testCommitRollback(table *engine.Table, t *testing.T) {
	// 测试事务提交
	fmt.Println("\n--- 测试事务提交 ---")
	tx1, err := table.Begin()
	if err != nil {
		t.Fatalf("开始事务失败: %v", err)
	}

	record1 := map[string]any{"name": "提交测试", "age": 35, "salary": 6000.0}
	id1, err := tx1.Insert(&record1)
	if err != nil {
		t.Fatalf("插入记录失败: %v", err)
	}

	err = tx1.Commit()
	if err != nil {
		t.Fatalf("提交事务失败: %v", err)
	}

	// 验证记录已提交
	readFields := map[string]any{"id": id1}
	iter := table.Search(&readFields)
	if iter == nil {
		t.Fatal("搜索记录失败")
	}
	defer iter.Release()

	records := iter.GetRecords(true)
	if len(records) != 1 {
		t.Fatalf("事务提交后，记录未找到")
	}
	fmt.Println("✓ 事务提交成功，记录已持久化")

	// 测试事务回滚
	fmt.Println("\n--- 测试事务回滚 ---")
	tx2, err := table.Begin()
	if err != nil {
		t.Fatalf("开始事务失败: %v", err)
	}

	record2 := map[string]any{"name": "回滚测试", "age": 40, "salary": 7000.0}
	id2, err := tx2.Insert(&record2)
	if err != nil {
		t.Fatalf("插入记录失败: %v", err)
	}

	err = tx2.Rollback()
	if err != nil {
		t.Fatalf("回滚事务失败: %v", err)
	}

	// 验证记录已回滚，未持久化
	readFields2 := map[string]any{"id": id2}
	iter2 := table.Search(&readFields2)
	if iter2 == nil {
		t.Fatal("搜索记录失败")
	}
	defer iter2.Release()

	records2 := iter2.GetRecords(true)
	if len(records2) != 0 {
		t.Fatalf("事务回滚后，记录仍存在，回滚失败")
	}
	fmt.Println("✓ 事务回滚成功，记录未持久化")
}
