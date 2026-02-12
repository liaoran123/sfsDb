package main

import (
	"fmt"
	"log"
	"os"

	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/match"
	"github.com/liaoran123/sfsDb/storage"
)

// 测试多表查询示例
func testMultiTableQuery() error {
	fmt.Println("\n=== 测试多表查询示例 ===")

	// 清理旧数据
	_ = os.RemoveAll("./test_multi_table_db")
	defer os.RemoveAll("./test_multi_table_db")

	// 初始化数据库
	_, err := storage.OpenDefaultDb("./test_multi_table_db")
	if err != nil {
		return fmt.Errorf("打开数据库失败: %v", err)
	}
	defer storage.CloseDb()

	// 创建表1
	table1, err := engine.TableNew("table1")
	if err != nil {
		return fmt.Errorf("创建表1失败: %v", err)
	}

	// 设置字段
	fields := map[string]any{"id": 0, "name": "", "age": 0}
	err = table1.SetFields(fields)
	if err != nil {
		return fmt.Errorf("设置表1字段失败: %v", err)
	}

	// 创建主键索引
	pk, _ := engine.DefaultPrimaryKeyNew("pk")
	pk.AddFields("id")
	err = table1.CreateIndex(pk)
	if err != nil {
		return fmt.Errorf("创建表1主键索引失败: %v", err)
	}

	// 插入测试数据
	testData := []map[string]any{
		{"id": 1, "name": "Alice", "age": 20},
		{"id": 2, "name": "Bob", "age": 25},
		{"id": 3, "name": "Charlie", "age": 30},
	}

	for _, data := range testData {
		_, err := table1.Insert(&data)
		if err != nil {
			return fmt.Errorf("插入表1数据失败: %v", err)
		}
	}

	// 创建表2
	table2, err := engine.TableNew("table2")
	if err != nil {
		return fmt.Errorf("创建表2失败: %v", err)
	}

	// 设置字段
	fields2 := map[string]any{"id": 0, "name": "", "age": 0}
	err = table2.SetFields(fields2)
	if err != nil {
		return fmt.Errorf("设置表2字段失败: %v", err)
	}

	// 创建主键索引
	pk2, _ := engine.DefaultPrimaryKeyNew("pk")
	pk2.AddFields("id")
	err = table2.CreateIndex(pk2)
	if err != nil {
		return fmt.Errorf("创建表2主键索引失败: %v", err)
	}

	// 插入测试数据
	testData2 := []map[string]any{
		{"id": 2, "name": "Bob", "age": 25},
		{"id": 3, "name": "Charlie", "age": 30},
		{"id": 4, "name": "David", "age": 35},
	}

	for _, data := range testData2 {
		_, err := table2.Insert(&data)
		if err != nil {
			return fmt.Errorf("插入表2数据失败: %v", err)
		}
	}

	// 测试内连接查询
	fmt.Println("\n1. 测试内连接查询:")
	iter1, err := table1.Search(&map[string]any{"id": nil})
	if err != nil {
		return fmt.Errorf("创建表1迭代器失败: %v", err)
	}
	defer iter1.Release()

	iter2, err := table2.Search(&map[string]any{"id": nil})
	if err != nil {
		return fmt.Errorf("创建表2迭代器失败: %v", err)
	}
	defer iter2.Release()

	// 从表2中提取id字段的值作为映射
	map2 := iter2.Map()
	defer iter2.ReleaseMap(map2)

	// 创建匹配条件：表1的id字段值必须在表2的id映射中
	mach := match.NewAND([]string{"id"}, map2)

	// 在表1的迭代器上设置匹配条件
	iter1.SetMatch(mach)

	// 获取匹配的记录（连接结果）
	rd4 := iter1.GetRecords(true)
	defer rd4.Release()

	// 打印连接结果
	fmt.Println("内连接结果:")
	for _, record := range rd4 {
		fmt.Println(record)
	}

	// 测试外连接查询
	fmt.Println("\n2. 测试外连接查询:")
	// 创建不匹配条件：表1的id字段值不在表2的id映射中
	mach1 := match.NewAND([]string{"id"}, map2, false)

	// 在表1的迭代器上设置不匹配条件
	iter1.SetMatch(mach1)

	// 获取不匹配的记录（外连接结果）
	rd5 := iter1.GetRecords(true)
	defer rd5.Release()

	// 打印外连接结果
	fmt.Println("外连接结果:")
	for _, record := range rd5 {
		fmt.Println(record)
	}

	fmt.Println("多表查询示例测试成功!")
	return nil
}

// 测试全文搜索示例
func testFullTextSearch() error {
	fmt.Println("\n=== 测试全文搜索示例 ===")

	// 清理旧数据
	_ = os.RemoveAll("./test_full_text_db")
	defer os.RemoveAll("./test_full_text_db")

	// 初始化数据库
	_, err := storage.OpenDefaultDb("./test_full_text_db")
	if err != nil {
		return fmt.Errorf("打开数据库失败: %v", err)
	}
	defer storage.CloseDb()

	// 创建表
	table, err := engine.TableNew("products")
	if err != nil {
		return fmt.Errorf("创建表失败: %v", err)
	}

	// 设置字段
	fields := map[string]any{"id": 0, "name": "", "description": ""}
	err = table.SetFields(fields)
	if err != nil {
		return fmt.Errorf("设置字段失败: %v", err)
	}

	// 创建主键索引
	pk, _ := engine.DefaultPrimaryKeyNew("pk")
	pk.AddFields("id")
	err = table.CreateIndex(pk)
	if err != nil {
		return fmt.Errorf("创建主键索引失败: %v", err)
	}

	// 创建全文索引
	fullTextIndex, err := engine.DefaultFullTextIndexNew("fulltext_desc")
	if err != nil {
		return fmt.Errorf("创建全文索引失败: %v", err)
	}
	fullTextIndex.AddFields("description", "id")
	fullTextIndex.SetFullField("description", 5)
	err = table.CreateIndex(fullTextIndex)
	if err != nil {
		return fmt.Errorf("创建全文索引失败: %v", err)
	}

	// 插入测试数据
	contentRecords := []map[string]any{
		{"id": 1, "name": "商品1", "description": "这是一款高性能的笔记本电脑，适合编程和游戏"},
		{"id": 2, "name": "商品2", "description": "智能手机，拥有强大的摄像头和长续航电池"},
		{"id": 3, "name": "商品3", "description": "无线耳机，提供沉浸式音频体验"},
		{"id": 4, "name": "商品4", "description": "智能手表，可监测健康数据和接收通知"},
	}

	for _, record := range contentRecords {
		_, err = table.Insert(&record)
		if err != nil {
			return fmt.Errorf("插入数据失败: %v", err)
		}
	}

	// 测试全文搜索
	fmt.Println("\n1. 搜索 '笔记本':")
	search1 := map[string]any{"description": "笔记本"}
	iter1, _ := table.Search(&search1)
	defer iter1.Release()
	records1 := iter1.GetRecords(true)
	defer records1.Release()
	for _, record := range records1 {
		fmt.Printf("   - %s: %s\n", record["name"], record["description"])
	}

	// 搜索包含"智能"的记录
	fmt.Println("\n2. 搜索 '智能':")
	search2 := map[string]any{"description": "智能"}
	iter2, _ := table.Search(&search2)
	defer iter2.Release()
	records2 := iter2.GetRecords(true)
	defer records2.Release()
	for _, record := range records2 {
		fmt.Printf("   - %s: %s\n", record["name"], record["description"])
	}

	fmt.Println("全文搜索示例测试成功!")
	return nil
}

// 测试基本使用示例（来自README）
func testBasicUsage() error {
	fmt.Println("\n=== 测试基本使用示例 ===")

	// 清理旧数据
	_ = os.RemoveAll("./test_basic_usage_db")
	defer os.RemoveAll("./test_basic_usage_db")

	// 初始化数据库
	_, err := storage.OpenDefaultDb("./test_basic_usage_db")
	if err != nil {
		return fmt.Errorf("打开数据库失败: %v", err)
	}
	defer storage.CloseDb()

	// 创建/打开用户表
	userTable, err := engine.TableNew("users")
	if err != nil {
		return fmt.Errorf("创建表失败: %v", err)
	}

	// 设置字段
	userFields := map[string]any{
		"id":      0,  // 用户ID
		"name":    "", // 用户名
		"age":     0,  // 年龄
		"email":   "", // 邮箱
		"address": "", // 地址
	}
	err = userTable.SetFields(userFields)
	if err != nil {
		return fmt.Errorf("设置字段失败: %v", err)
	}

	// 创建主键索引
	primaryKey, err := engine.DefaultPrimaryKeyNew("id")
	if err != nil {
		return fmt.Errorf("创建主键索引失败: %v", err)
	}
	primaryKey.AddFields("id")
	err = userTable.CreateIndex(primaryKey)
	if err != nil {
		return fmt.Errorf("创建索引失败: %v", err)
	}

	// 创建普通索引
	nameIndex, err := engine.DefaultNormalIndexNew("name_index")
	if err != nil {
		return fmt.Errorf("创建普通索引失败: %v", err)
	}
	nameIndex.AddFields("name")
	err = userTable.CreateIndex(nameIndex)
	if err != nil {
		return fmt.Errorf("创建索引失败: %v", err)
	}

	// 插入数据
	users := []map[string]any{
		{"id": 1, "name": "张三", "age": 25, "email": "zhangsan@example.com", "address": "北京市"},
		{"id": 2, "name": "李四", "age": 30, "email": "lisi@example.com", "address": "上海市"},
		{"id": 3, "name": "王五", "age": 35, "email": "wangwu@example.com", "address": "广州市"},
	}

	for _, user := range users {
		currentID, err := userTable.Insert(&user)
		if err != nil {
			return fmt.Errorf("插入数据失败: %v", err)
		}
		fmt.Printf("插入用户成功, ID: %d\n", currentID)
	}

	// 主键查询
	fmt.Println("\n主键查询:")
	iter, err := userTable.Search(&map[string]any{"id": 1})
	if err != nil {
		return fmt.Errorf("搜索失败: %v", err)
	}
	records := iter.GetRecords(true)
	defer records.Release()

	if len(records) > 0 {
		fmt.Printf("查询结果: %v\n", records[0])
	}

	// 普通索引查询
	fmt.Println("\n普通索引查询:")
	nameIter, err := userTable.Search(&map[string]any{"name": "李四"})
	if err != nil {
		return fmt.Errorf("搜索失败: %v", err)
	}
	nameRecords := nameIter.GetRecords(true)
	defer nameRecords.Release()

	if len(nameRecords) > 0 {
		fmt.Printf("按姓名查询结果: %v\n", nameRecords[0])
	}

	// 更新数据
	fmt.Println("\n更新数据:")
	updateData := map[string]any{
		"id":      1,                          // 用于定位记录
		"email":   "zhangsan_new@example.com", // 更新邮箱
		"address": "深圳市",                      // 更新地址
	}
	err = userTable.Update(&updateData)
	if err != nil {
		return fmt.Errorf("更新数据失败: %v", err)
	}
	fmt.Println("更新数据成功")

	// 验证更新
	iter, err = userTable.Search(&map[string]any{"id": 1})
	if err != nil {
		return fmt.Errorf("搜索失败: %v", err)
	}
	records = iter.GetRecords(true)
	if len(records) > 0 {
		fmt.Printf("更新后的数据: %v\n", records[0])
	}

	// 删除数据
	fmt.Println("\n删除数据:")
	deleteData := map[string]any{
		"id": 3, // 用于定位要删除的记录
	}
	err = userTable.Delete(&deleteData)
	if err != nil {
		return fmt.Errorf("删除数据失败: %v", err)
	}
	fmt.Println("删除数据成功")

	// 验证删除
	iter, err = userTable.Search(&map[string]any{"id": 3})
	if err != nil {
		return fmt.Errorf("搜索失败: %v", err)
	}
	records = iter.GetRecords(true)
	fmt.Printf("删除后查询结果数: %d\n", len(records))

	// 查询所有数据
	fmt.Println("\n查询所有数据:")
	allIter, err := userTable.Search(&map[string]any{})
	if err != nil {
		return fmt.Errorf("搜索失败: %v", err)
	}
	allRecords := allIter.GetRecords(true)
	fmt.Printf("当前表中共有 %d 条记录\n", len(allRecords))
	for i, r := range allRecords {
		fmt.Printf("记录 %d: %v\n", i+1, r)
	}

	fmt.Println("基本使用示例测试成功!")
	return nil
}

func main11() {
	fmt.Println("=== 测试docs包所有示例 ===")

	// 测试基本使用示例
	if err := testBasicUsage(); err != nil {
		log.Printf("基本使用示例测试失败: %v", err)
	} else {
		fmt.Println("基本使用示例测试通过!")
	}

	// 测试多表查询示例
	if err := testMultiTableQuery(); err != nil {
		log.Printf("多表查询示例测试失败: %v", err)
	} else {
		fmt.Println("多表查询示例测试通过!")
	}

	// 测试全文搜索示例
	if err := testFullTextSearch(); err != nil {
		log.Printf("全文搜索示例测试失败: %v", err)
	} else {
		fmt.Println("全文搜索示例测试通过!")
	}

	fmt.Println("\n=== 所有示例测试完成 ===")
}
