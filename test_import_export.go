package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/storage"
)

func main() {
	// 创建测试目录
	testDir := "test_import_export"
	os.RemoveAll(testDir)
	defer os.RemoveAll(testDir)

	// 打开或创建数据库
	_, err := storage.OpenDefaultDb(testDir)
	if err != nil {
		fmt.Printf("打开数据库失败: %v\n", err)
		return
	}
	defer storage.CloseDb()

	// 创建表
	table, err := engine.TableNew("test_table")
	if err != nil {
		fmt.Printf("创建表失败: %v\n", err)
		return
	}

	// 设置表字段
	tableFields := map[string]any{
		"id":     0,
		"name":   "",
		"age":    0,
		"active": false,
		"score":  0.0,
	}

	if err := table.SetFields(tableFields); err != nil {
		fmt.Printf("设置表字段失败: %v\n", err)
		return
	}

	// 创建主键索引
	pkIndex, err := engine.DefaultPrimaryKeyNew("id")
	if err != nil {
		fmt.Printf("创建主键索引失败: %v\n", err)
		return
	}
	pkIndex.AddFields("id")
	if err := table.CreateIndex(pkIndex); err != nil {
		fmt.Printf("创建主键索引失败: %v\n", err)
		return
	}

	// 插入测试数据
	testData := []map[string]any{
		{"id": 1, "name": "Alice", "age": 25, "active": true, "score": 85.5},
		{"id": 2, "name": "Bob", "age": 30, "active": false, "score": 90.0},
		{"id": 3, "name": "Charlie", "age": 35, "active": true, "score": 75.5},
	}

	for _, data := range testData {
		_, err := table.Insert(&data)
		if err != nil {
			fmt.Printf("插入数据失败: %v\n", err)
			return
		}
	}

	fmt.Println("测试数据插入成功!")

	// 测试导出为CSV
	csvPath := filepath.Join(testDir, "test.csv")
	if err := table.ExportToCSV(csvPath); err != nil {
		fmt.Printf("导出CSV失败: %v\n", err)
		return
	}
	fmt.Printf("CSV导出成功: %s\n", csvPath)

	// 测试导出为JSON
	jsonPath := filepath.Join(testDir, "test.json")
	if err := table.ExportToJSON(jsonPath); err != nil {
		fmt.Printf("导出JSON失败: %v\n", err)
		return
	}
	fmt.Printf("JSON导出成功: %s\n", jsonPath)

	// 测试导出为SQL
	sqlPath := filepath.Join(testDir, "test.sql")
	if err := table.ExportToSQL(sqlPath); err != nil {
		fmt.Printf("导出SQL失败: %v\n", err)
		return
	}
	fmt.Printf("SQL导出成功: %s\n", sqlPath)

	// 查询所有记录，验证导出数据的正确性
	iter := table.ForData()
	defer iter.Release()

	fmt.Println("\n导出前的数据:")
	for iter.Next() {
		value := iter.Value()
		fieldsBytes := iter.ParseBytes(iter.Key(), value)
		rec := table.RecordByteToAny(fieldsBytes)
		fmt.Printf("%v\n", rec)
	}

	// 为了测试导入功能，我们重新创建一个表
	table2, err := engine.TableNew("test_table2")
	if err != nil {
		fmt.Printf("创建第二个表失败: %v\n", err)
		return
	}

	if err := table2.SetFields(tableFields); err != nil {
		fmt.Printf("设置第二个表字段失败: %v\n", err)
		return
	}

	// 创建主键索引
	pkIndex2, err := engine.DefaultPrimaryKeyNew("id")
	if err != nil {
		fmt.Printf("创建第二个表主键索引失败: %v\n", err)
		return
	}
	pkIndex2.AddFields("id")
	if err := table2.CreateIndex(pkIndex2); err != nil {
		fmt.Printf("创建第二个表主键索引失败: %v\n", err)
		return
	}

	// 测试从CSV导入
	if err := table2.ImportFromCSV(csvPath, 100); err != nil {
		fmt.Printf("从CSV导入失败: %v\n", err)
		return
	}
	fmt.Println("从CSV导入成功!")

	// 验证导入的数据
	iter2 := table2.ForData()
	defer iter2.Release()

	fmt.Println("\n从CSV导入后的数据:")
	for iter2.Next() {
		value := iter2.Value()
		fieldsBytes := iter2.ParseBytes(iter2.Key(), value)
		rec := table2.RecordByteToAny(fieldsBytes)
		fmt.Printf("%v\n", rec)
	}

	// 为了测试JSON导入，我们重新创建一个表
	table3, err := engine.TableNew("test_table3")
	if err != nil {
		fmt.Printf("创建第三个表失败: %v\n", err)
		return
	}

	if err := table3.SetFields(tableFields); err != nil {
		fmt.Printf("设置第三个表字段失败: %v\n", err)
		return
	}

	// 创建主键索引
	pkIndex3, err := engine.DefaultPrimaryKeyNew("id")
	if err != nil {
		fmt.Printf("创建第三个表主键索引失败: %v\n", err)
		return
	}
	pkIndex3.AddFields("id")
	if err := table3.CreateIndex(pkIndex3); err != nil {
		fmt.Printf("创建第三个表主键索引失败: %v\n", err)
		return
	}

	// 测试从JSON导入
	if err := table3.ImportFromJSON(jsonPath, 100); err != nil {
		fmt.Printf("从JSON导入失败: %v\n", err)
		return
	}
	fmt.Println("从JSON导入成功!")

	// 验证导入的数据
	iter3 := table3.ForData()
	defer iter3.Release()

	fmt.Println("\n从JSON导入后的数据:")
	for iter3.Next() {
		value := iter3.Value()
		fieldsBytes := iter3.ParseBytes(iter3.Key(), value)
		rec := table3.RecordByteToAny(fieldsBytes)
		fmt.Printf("%v\n", rec)
	}

	fmt.Println("\n所有测试都通过了!")
}
