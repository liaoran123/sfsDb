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

		records := iter.GetRecordSet(true)
		for _, record := range records {
			fmt.Printf("ID: %v, Name: %v, Age: %v\n", record["id"], record["name"], record["age"])
		}
	}

	// 8. 普通索引查询
	fmt.Println("\n8. 普通索引查询")
	{
		iter, err := userTable.Search(&map[string]any{"name": "李四"})
		defer iter.Release()
		if err != nil {
			fmt.Printf("搜索失败: %v\n", err)
			return
		}

		records := iter.GetRecordSet(true)
		for _, record := range records {
			fmt.Printf("ID: %v, Name: %v, Age: %v\n", record["id"], record["name"], record["age"])
		}
	}

	// 9. 更新数据
	fmt.Println("\n9. 更新数据")
	updateUser := map[string]any{"id": 1, "name": "张三(已更新)", "age": 26}
	err = userTable.Update(&updateUser)
	if err != nil {
		fmt.Printf("更新失败: %v\n", err)
		return
	}
	fmt.Println("更新用户成功")

	// 10. 删除数据
	fmt.Println("\n10. 删除数据")
	deleteUser := map[string]any{"id": 3}
	err = userTable.Delete(&deleteUser)
	if err != nil {
		fmt.Printf("删除失败: %v\n", err)
		return
	}
	fmt.Println("删除用户成功")

	// 11. 事务示例
	fmt.Println("\n11. 事务示例")
	db := dbManager.GetDB()
	batch := db.GetBatch()
	defer batch.Reset()

	// 插入新用户（使用批量操作）
	newUser := map[string]any{"id": 4, "name": "赵六", "age": 40, "email": "zhaoliu@example.com", "address": "深圳市"}
	currentID, err := userTable.Insert(&newUser, batch)
	if err != nil {
		fmt.Printf("事务插入失败: %v\n", err)
		return
	}
	fmt.Printf("事务插入用户成功, ID: %d\n", currentID)

	// 提交事务
	err = db.WriteBatch(batch)
	if err != nil {
		fmt.Printf("事务提交失败: %v\n", err)
		return
	}
	fmt.Println("事务提交成功")

	fmt.Println("\n测试完成!")
}