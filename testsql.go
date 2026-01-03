package main

import (
	"fmt"
	"github.com/liaoran123/sfsDb/sql"
)

func main() {
	// 创建解析器实例
	parser := sql.NewParser()
	
	// 测试SELECT语句
	sqlStr := "SELECT id, name, age FROM users WHERE age > 18 ORDER BY id DESC"
	fmt.Printf("原始SQL: %s\n", sqlStr)
	
	// 解析SQL
	stmt, err := parser.Parse(sqlStr)
	if err != nil {
		fmt.Printf("解析错误: %v\n", err)
		return
	}
	
	fmt.Printf("语句类型: %s\n", stmt.Type())
	
	// 创建访问器
	visitor := sql.NewVisitor()
	
	// 提取表名
	tables, err := visitor.ExtractTableNames(stmt)
	if err != nil {
		fmt.Printf("提取表名错误: %v\n", err)
		return
	}
	fmt.Printf("表名: %v\n", tables)
	
	// 提取列名
	columns, err := visitor.ExtractColumns(stmt)
	if err != nil {
		fmt.Printf("提取列名错误: %v\n", err)
		return
	}
	fmt.Printf("列名: %v\n", columns)
	
	fmt.Println("测试完成!")
}