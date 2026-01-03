package sql

// Example 展示如何使用SQL解析器
// 这个示例文件用于演示SQL解析器的基本用法

// ExampleUsage 展示基本的SQL解析用法
func ExampleUsage() {
	_ = "unimplemented"
	// 创建解析器实例
	// parser := NewParser()
	//
	// 待解析的SQL语句
	// sql := "SELECT id, name, age FROM users WHERE age > 18 ORDER BY id DESC"
	//
	// 解析SQL语句
	// stmt, err := parser.Parse(sql)
	// if err != nil {
	//  panic(err)
	// }
	//
	// 创建访问器实例
	// visitor := NewVisitor()
	//
	// 提取表名
	// tables, err := visitor.ExtractTableNames(stmt)
	// if err != nil {
	//  panic(err)
	// }
	//
	// 提取列名
	// columns, err := visitor.ExtractColumns(stmt)
	// if err != nil {
	//  panic(err)
	// }
	//
	// 输出结果
	/*
		fmt.Printf("Statement Type: %s\n", stmt.Type())
		fmt.Printf("Raw SQL: %s\n", stmt.Raw())
		fmt.Printf("Tables: %v\n", tables)
		fmt.Printf("Columns: %v\n", columns)
	*/
}

// ExampleInsert 展示INSERT语句的解析
func ExampleInsert() {
	_ = "unimplemented"
	// 创建解析器实例
	// parser := NewParser()
	//
	// INSERT语句示例
	// sql := "INSERT INTO users (name, age, email) VALUES ('张三', 25, 'zhangsan@example.com')"
	//
	// 解析SQL语句
	// stmt, err := parser.Parse(sql)
	// if err != nil {
	//  panic(err)
	// }
	//
	// 创建访问器实例
	// visitor := NewVisitor()
	//
	// 提取表名
	// tables, err := visitor.ExtractTableNames(stmt)
	// if err != nil {
	//  panic(err)
	// }
	//
	// 提取列名
	// columns, err := visitor.ExtractColumns(stmt)
	// if err != nil {
	//  panic(err)
	// }
	//
	// 输出结果
	/*
		fmt.Printf("Statement Type: %s\n", stmt.Type())
		fmt.Printf("Raw SQL: %s\n", stmt.Raw())
		fmt.Printf("Tables: %v\n", tables)
		fmt.Printf("Columns: %v\n", columns)
	*/
}
