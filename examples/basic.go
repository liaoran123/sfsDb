package main

import (
	"fmt"

	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/storage"
	"github.com/liaoran123/sfsDb/util"
)

func RunExample() {
	fmt.Println("SFSDB Basic Example")
	fmt.Println("==================")

	// 1. 初始化数据库
	fmt.Println("\n1. 初始化数据库")
	storage.OpenDefaultDb("./example_db")
	defer storage.CloseDb()

	// 2. 创建表
	fmt.Println("\n2. 创建表")
	table, err := engine.TableNew("users")
	if err != nil {
		fmt.Printf("创建表失败: %v\n", err)
		return
	}

	// 3. 设置字段
	fmt.Println("\n3. 设置字段")
	fields := map[string]any{
		"id":     0,     // 整数类型
		"name":   "",    // 字符串类型
		"age":    0,     // 整数类型
		"email":  "",    // 字符串类型
		"active": false, // 布尔类型
		"score":  0.0,   // 浮点数类型
	}
	err = table.SetFields(fields)
	if err != nil {
		fmt.Printf("设置字段失败: %v\n", err)
		return
	}

	// 4. 创建主键索引
	fmt.Println("\n4. 创建主键索引")
	pk, err := engine.DefaultPrimaryKeyNew("pk")
	if err != nil {
		fmt.Printf("创建主键索引失败: %v\n", err)
		return
	}
	pk.AddFields("id")
	err = table.CreateIndex(pk)
	if err != nil {
		fmt.Printf("创建索引失败: %v\n", err)
		return
	}

	// 5. 创建二级索引
	fmt.Println("\n5. 创建二级索引")
	ageIdx, err := engine.DefaultNormalIndexNew("age_idx")
	if err != nil {
		fmt.Printf("创建二级索引失败: %v\n", err)
		return
	}
	ageIdx.AddFields("age")
	err = table.CreateIndex(ageIdx)
	if err != nil {
		fmt.Printf("创建索引失败: %v\n", err)
		return
	}

	// 6. 插入数据
	fmt.Println("\n6. 插入数据")
	users := []map[string]any{
		{"id": 1, "name": "Alice", "age": 25, "email": "alice@example.com", "active": true, "score": 95.5},
		{"id": 2, "name": "Bob", "age": 30, "email": "bob@example.com", "active": false, "score": 85.0},
		{"id": 3, "name": "Charlie", "age": 35, "email": "charlie@example.com", "active": true, "score": 90.5},
		{"id": 4, "name": "David", "age": 28, "email": "david@example.com", "active": true, "score": 88.0},
		{"id": 5, "name": "Eve", "age": 22, "email": "eve@example.com", "active": false, "score": 92.0},
	}

	for _, user := range users {
		currentID, err := table.Insert(&user)
		if err != nil {
			fmt.Printf("插入数据失败: %v\n", err)
			return
		}
		fmt.Printf("插入用户 %s 成功, ID: %d\n", user["name"], currentID)
	}

	// 7. 查询所有数据
	fmt.Println("\n7. 查询所有数据")
	iter := table.Search(&map[string]any{"id": nil})
	defer iter.Release()
	records := iter.GetRecords(true)
	fmt.Printf("总记录数: %d\n", len(records))
	for i, record := range records {
		fmt.Printf("记录 %d: %v\n", i+1, record)
	}

	// 8. 条件查询
	fmt.Println("\n8. 条件查询")
	// 查询年龄大于25的用户
	iter = table.Search(&map[string]any{"age": 25}, util.GreaterThan)
	defer iter.Release()
	records = iter.GetRecords(true)
	fmt.Printf("年龄大于25的用户数: %d\n", len(records))
	for i, record := range records {
		fmt.Printf("用户 %d: %s, 年龄: %d\n", i+1, record["name"], record["age"])
	}

	// 9. 更新数据
	fmt.Println("\n9. 更新数据")
	updateData := map[string]any{
		"id":     1,             // 主键，用于定位记录
		"name":   "Alice Smith", // 更新姓名
		"age":    26,            // 更新年龄
		"active": true,          // 更新状态
		"score":  97.5,          // 更新分数
	}
	err = table.Update(&updateData)
	if err != nil {
		fmt.Printf("更新数据失败: %v\n", err)
		return
	}
	fmt.Println("更新用户 Alice 成功")

	// 验证更新
	iter = table.Search(&map[string]any{"id": 1})
	defer iter.Release()
	records = iter.GetRecords(true)
	if len(records) > 0 {
		fmt.Printf("更新后的数据: %v\n", records[0])
	}

	// 10. 删除数据
	fmt.Println("\n10. 删除数据")
	deleteData := map[string]any{"id": 5} // 删除ID为5的用户
	err = table.Delete(&deleteData)
	if err != nil {
		fmt.Printf("删除数据失败: %v\n", err)
		return
	}
	fmt.Println("删除用户 Eve 成功")

	// 验证删除
	iter = table.Search(&map[string]any{"id": nil})
	defer iter.Release()
	records = iter.GetRecords(true)
	fmt.Printf("删除后总记录数: %d\n", len(records))

	// 11. 使用二级索引查询
	fmt.Println("\n11. 使用二级索引查询")
	iter = table.Search(&map[string]any{"age": 30})
	defer iter.Release()
	records = iter.GetRecords(true)
	fmt.Printf("年龄等于30的用户数: %d\n", len(records))
	for i, record := range records {
		fmt.Printf("用户 %d: %s, 年龄: %d\n", i+1, record["name"], record["age"])
	}

	fmt.Println("\n示例完成!")
}
