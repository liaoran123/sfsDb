package main

import (
	"fmt"

	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/storage"
)

func main() {
	// 打开默认数据库
	_, err := storage.OpenDefaultDb("./test_transaction_db")
	if err != nil {
		panic(err)
	}
	defer storage.CloseDb()

	// 创建表
	table, err := engine.TableNew("transaction_test")
	if err != nil {
		panic(err)
	}

	// 设置表字段
	fields := map[string]any{
		"id":     0,
		"name":   "",
		"age":    0,
		"salary": 0.0,
	}
	err = table.SetFields(fields)
	if err != nil {
		panic(err)
	}

	// 创建主键索引
	primaryKey, err := engine.DefaultPrimaryKeyNew("pk_id")
	if err != nil {
		panic(err)
	}
	primaryKey.AddFields("id")
	err = table.CreateIndex(primaryKey)
	if err != nil {
		panic(err)
	}

	fmt.Println("表创建成功")

	// 演示1：使用封装的事务API（推荐）
	fmt.Println("\n=== 演示1：使用封装的事务API ===")
	testEncapsulatedTransaction(table)

	// 演示2：使用手动事务控制（底层API）
	fmt.Println("\n=== 演示2：使用手动事务控制 ===")
	testManualTransaction(table)
}

// 演示使用封装的事务API
func testEncapsulatedTransaction(table *engine.Table) {
	// 开始事务
	tx, err := table.Begin()
	if err != nil {
		panic(fmt.Sprintf("开始事务失败: %v", err))
	}

	// 插入第一条记录（直接设置正确的salary值，避免后续更新）
	record1 := map[string]any{"name": "张三", "age": 30, "salary": 5500.0}
	id1, err := tx.Insert(&record1)
	if err != nil {
		tx.Rollback()
		panic(fmt.Sprintf("插入记录失败: %v", err))
	}
	fmt.Printf("插入记录1成功，ID: %d\n", id1)

	// 插入第二条记录
	record2 := map[string]any{"name": "李四", "age": 25, "salary": 4000.0}
	id2, err := tx.Insert(&record2)
	if err != nil {
		tx.Rollback()
		panic(fmt.Sprintf("插入记录失败: %v", err))
	}
	fmt.Printf("插入记录2成功，ID: %d\n", id2)

	// 提交事务
	err = tx.Commit()
	if err != nil {
		panic(fmt.Sprintf("提交事务失败: %v", err))
	}
	fmt.Println("事务提交成功")

	// 验证结果
	printAllRecords(table)
}

// 演示使用手动事务控制
func testManualTransaction(table *engine.Table) {
	// 获取批量操作对象
	batch := storage.KVDb.GetBatch()
	if batch == nil {
		panic("获取batch失败")
	}

	// 插入第三条记录
	record3 := map[string]any{"name": "王五", "age": 35, "salary": 6000.0}
	id3, err := table.Insert(&record3, batch)
	if err != nil {
		panic(fmt.Sprintf("插入记录失败: %v", err))
	}
	fmt.Printf("插入记录3成功，ID: %d\n", id3)

	// 插入第四条记录（直接设置正确的salary值，避免后续更新）
	record4 := map[string]any{"name": "赵六", "age": 28, "salary": 4800.0}
	id4, err := table.Insert(&record4, batch)
	if err != nil {
		panic(fmt.Sprintf("插入记录失败: %v", err))
	}
	fmt.Printf("插入记录4成功，ID: %d\n", id4)

	// 提交批量操作（事务）
	err = storage.KVDb.WriteBatch(batch)
	if err != nil {
		panic(fmt.Sprintf("提交事务失败: %v", err))
	}
	fmt.Println("手动事务提交成功")

	// 验证结果
	printAllRecords(table)
}

// 打印所有记录
func printAllRecords(table *engine.Table) {
	iter := table.ForData()
	defer iter.Release()

	records := iter.GetRecords(true)
	fmt.Printf("\n表中共有 %d 条记录:\n", len(records))
	for _, record := range records {
		fmt.Printf("ID: %d, Name: %s, Age: %d, Salary: %.2f\n",
			record["id"], record["name"], record["age"], record["salary"])
	}
}
