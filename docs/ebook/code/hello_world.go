package main

import (
	"fmt"

	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/record"
	"github.com/liaoran123/sfsDb/storage"
)

func main() {
	fmt.Println("=== sfsDb Hello World 示例 ===")

	// 1. 初始化数据库
	fmt.Println("\n1. 初始化数据库...")
	dbManager := storage.GetDBManager()
	_, err := dbManager.OpenDB("./hello_world_db")
	if err != nil {
		panic(fmt.Sprintf("打开数据库失败: %v", err))
	}
	defer dbManager.CloseDB()
	fmt.Println("✓ 数据库初始化成功")

	// 2. 创建表
	fmt.Println("\n2. 创建用户表...")
	userTable, err := engine.TableNew("users")
	if err != nil {
		panic(fmt.Sprintf("创建表失败: %v", err))
	}
	fmt.Println("✓ 表创建成功")

	// 3. 设置字段
	fmt.Println("\n3. 设置表字段...")
	userFields := map[string]any{
		"id":   0,  // 自动增值主键
		"name": "", // 用户名
		"age":  0,  // 年龄
	}
	err = userTable.SetFields(userFields)
	if err != nil {
		panic(fmt.Sprintf("设置字段失败: %v", err))
	}
	fmt.Println("✓ 字段设置成功")

	// 4. 创建主键索引
	fmt.Println("\n4. 创建主键索引...")
	primaryKey, err := engine.DefaultPrimaryKeyNew("id")
	if err != nil {
		panic(fmt.Sprintf("创建主键索引失败: %v", err))
	}
	primaryKey.AddFields("id")
	err = userTable.CreateIndex(primaryKey)
	if err != nil {
		panic(fmt.Sprintf("创建索引失败: %v", err))
	}
	fmt.Println("✓ 主键索引创建成功")

	// 5. 插入数据
	fmt.Println("\n5. 插入测试数据...")
	users := []map[string]any{
		{"id": 1, "name": "张三", "age": 25},
		{"id": 2, "name": "李四", "age": 30},
		{"id": 3, "name": "王五", "age": 35},
	}

	for _, user := range users {
		currentID, err := userTable.Insert(&user)
		if err != nil {
			panic(fmt.Sprintf("插入数据失败: %v", err))
		}
		fmt.Printf("  ✓ 插入用户成功, ID: %d\n", currentID)
	}

	// 6. 查询数据
	fmt.Println("\n6. 查询数据...")
	{
		iter, err := userTable.Search(&map[string]any{"id": 1})
		defer engine.GlobalTableIterPool.Put(iter)
		if err != nil {
			panic(fmt.Sprintf("搜索失败: %v", err))
		}
		records := iter.GetRecords(true)
		defer record.PutRecords(records)

		if len(records) > 0 {
			fmt.Printf("  ✓ 查询结果: %v\n", records[0])
		}
	}

	// 7. 查询所有数据
	fmt.Println("\n7. 查询所有数据...")
	{
		allIter, err := userTable.Search(&map[string]any{})
		defer engine.GlobalTableIterPool.Put(allIter)
		if err != nil {
			panic(fmt.Sprintf("搜索失败: %v", err))
		}
		allRecords := allIter.GetRecords(true)
		defer record.PutRecords(allRecords)

		fmt.Printf("  ✓ 当前表中共有 %d 条记录\n", len(allRecords))
		for i, r := range allRecords {
			fmt.Printf("  记录 %d: %v\n", i+1, r)
		}
	}

	fmt.Println("\n=== 示例运行完成！===")
}
