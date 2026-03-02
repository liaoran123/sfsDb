package main

import (
	"fmt"

	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/storage"
)

func main() {
	fmt.Println("sfsDb README示例代码测试")
	fmt.Println("====================")

	// 1. 初始化数据库
	fmt.Println("\n1. 初始化数据库")
	dbManager := storage.GetDBManager()
	_, err := dbManager.OpenDB("./readme_example_db")
	if err != nil {
		fmt.Printf("打开数据库失败: %v\n", err)
		return
	}
	defer dbManager.CloseDB()

	// 2. 创建/打开用户表
	fmt.Println("\n2. 创建用户表")
	userTable, err := engine.TableNew("users")
	if err != nil {
		fmt.Printf("创建表失败: %v\n", err)
		return
	}

	// 3. 设置字段
	fmt.Println("\n3. 设置字段")
	userFields := map[string]any{
		"id":      0,  // 用户ID
		"name":    "", // 用户名
		"age":     0,  // 年龄
		"email":   "", // 邮箱
		"address": "", // 地址
	}
	err = userTable.SetFields(userFields)
	if err != nil {
		fmt.Printf("设置字段失败: %v\n", err)
		return
	}

	// 4. 创建主键索引
	fmt.Println("\n4. 创建主键索引")
	primaryKey, err := engine.DefaultPrimaryKeyNew("id")
	if err != nil {
		fmt.Printf("创建主键索引失败: %v\n", err)
		return
	}
	primaryKey.AddFields("id")
	err = userTable.CreateIndex(primaryKey)
	if err != nil {
		fmt.Printf("创建索引失败: %v\n", err)
		return
	}

	// 5. 创建普通索引
	fmt.Println("\n5. 创建普通索引")
	nameIndex, err := engine.DefaultNormalIndexNew("name_index")
	if err != nil {
		fmt.Printf("创建普通索引失败: %v\n", err)
		return
	}
	nameIndex.AddFields("name")
	err = userTable.CreateIndex(nameIndex)
	if err != nil {
		fmt.Printf("创建索引失败: %v\n", err)
		return
	}

	// 6. 插入数据
	fmt.Println("\n6. 插入数据")
	users := []map[string]any{
		{"id": 1, "name": "张三", "age": 25, "email": "zhangsan@example.com", "address": "北京市"},
		{"id": 2, "name": "李四", "age": 30, "email": "lisi@example.com", "address": "上海市"},
		{"id": 3, "name": "王五", "age": 35, "email": "wangwu@example.com", "address": "广州市"},
	}

	for _, user := range users {
		currentID, err := userTable.Insert(&user)
		if err != nil {
			fmt.Printf("插入数据失败: %v\n", err)
			return
		}
		fmt.Printf("插入用户成功, ID: %d\n", currentID)
	}

	// 7. 主键查询
	fmt.Println("\n7. 主键查询")
	{
		iter, err := userTable.Search(&map[string]any{"id": 1})
		defer iter.Release()
		if err != nil {
			fmt.Printf("搜索失败: %v\n", err)
			return
		}
		records := iter.GetRecords(true)
		defer records.Release()

		if len(records) > 0 {
			fmt.Printf("查询结果: %v\n", records[0])
		}
	}

	// 8. 普通索引查询
	fmt.Println("\n8. 普通索引查询")
	{
		nameIter, err := userTable.Search(&map[string]any{"name": "李四"})
		defer nameIter.Release()
		if err != nil {
			fmt.Printf("搜索失败: %v\n", err)
			return
		}
		nameRecords := nameIter.GetRecords(true)
		defer nameRecords.Release()

		if len(nameRecords) > 0 {
			fmt.Printf("按姓名查询结果: %v\n", nameRecords[0])
		}
	}

	// 9. 更新数据
	fmt.Println("\n9. 更新数据")
	updateData := map[string]any{
		"id":      1,                          // 用于定位记录
		"email":   "zhangsan_new@example.com", // 更新邮箱
		"address": "深圳市",                      // 更新地址
	}
	err = userTable.Update(&updateData)
	if err != nil {
		fmt.Printf("更新数据失败: %v\n", err)
		return
	}
	fmt.Println("更新数据成功")

	// 验证更新
	{
		iter, err := userTable.Search(&map[string]any{"id": 1})
		defer iter.Release()
		if err != nil {
			fmt.Printf("搜索失败: %v\n", err)
			return
		}
		records := iter.GetRecords(true)
		defer records.Release()
		if len(records) > 0 {
			fmt.Printf("更新后的数据: %v\n", records[0])
		}
	}

	// 10. 删除数据
	fmt.Println("\n10. 删除数据")
	deleteData := map[string]any{
		"id": 3, // 用于定位要删除的记录
	}
	err = userTable.Delete(&deleteData)
	if err != nil {
		fmt.Printf("删除数据失败: %v\n", err)
		return
	}
	fmt.Println("删除数据成功")

	// 验证删除
	{
		iter, err := userTable.Search(&map[string]any{"id": 3})
		defer iter.Release()
		if err != nil {
			fmt.Printf("搜索失败: %v\n", err)
			return
		}
		records := iter.GetRecords(true)
		defer records.Release()
		fmt.Printf("删除后查询结果数: %d\n", len(records))
	}

	// 11. 查询所有数据
	fmt.Println("\n11. 查询所有数据")
	{
		allIter, err := userTable.Search(&map[string]any{})
		defer allIter.Release()
		if err != nil {
			fmt.Printf("搜索失败: %v\n", err)
			return
		}
		allRecords := allIter.GetRecords(true)
		defer allRecords.Release()
		fmt.Printf("当前表中共有 %d 条记录\n", len(allRecords))
		for i, r := range allRecords {
			fmt.Printf("记录 %d: %v\n", i+1, r)
		}
	}

	fmt.Println("\n测试完成，所有操作均成功执行！")
}
