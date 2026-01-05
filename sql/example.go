package sql

import (
	"fmt"
)

// Example 展示SQL解析器的完整功能
func Example() {
	// 创建解析器实例
	parser := NewParser()
	// 创建访问器实例
	visitor := NewVisitor()

	fmt.Println("=== SQL解析器功能示例 ===")

	// 1. 测试SELECT语句解析
	fmt.Println("1. SELECT语句解析")
	fmt.Println("================")
	selectSQL := `SELECT u.id, u.name, o.order_id, o.order_date, p.product_name 
FROM users u 
JOIN orders o ON u.id = o.user_id 
JOIN products p ON o.product_id = p.id 
WHERE u.age > 18 AND o.status = 'completed' 
ORDER BY o.order_date DESC`
	stmt, _ := parser.Parse(selectSQL)
	fmt.Printf("语句类型: %s\n", stmt.Type())
	fmt.Printf("原始SQL: %s\n", stmt.Raw())

	// 提取表名
	tables, _ := visitor.ExtractTableNames(stmt)
	fmt.Printf("\n提取的表名: %v\n", tables)

	// 提取表名及其别名
	tableAliases, _ := visitor.ExtractTablesWithAliases(stmt)
	fmt.Println("\n提取的表名及其别名:")
	for i, table := range tableAliases {
		fmt.Printf("  表 %d: 名称='%s', 别名='%s'\n", i+1, table["table"], table["alias"])
	}

	// 提取列名
	columns, _ := visitor.ExtractColumns(stmt)
	fmt.Printf("\n提取的列名: %v\n", columns)

	// 提取列名到表名的映射
	columnMappings, _ := visitor.ExtractColumnTableMapping(stmt)
	fmt.Println("\n列名到表名的映射关系:")
	for i, mapping := range columnMappings {
		fmt.Printf("  映射 %d: 列名='%s', 表名='%s', 是否限定='%t'\n",
			i+1, mapping.ColumnName, mapping.TableName, mapping.IsQualified)
	}

	// 2. 测试INSERT语句解析
	fmt.Println("\n\n2. INSERT语句解析")
	fmt.Println("================")
	insertSQL := "INSERT INTO users (id, name, age, email) VALUES (1, '张三', 25, 'zhangsan@example.com')"
	stmt, _ = parser.Parse(insertSQL)
	fmt.Printf("语句类型: %s\n", stmt.Type())
	fmt.Printf("原始SQL: %s\n", stmt.Raw())

	// 提取表名
	tables, _ = visitor.ExtractTableNames(stmt)
	fmt.Printf("\n提取的表名: %v\n", tables)

	// 提取列名
	columns, _ = visitor.ExtractColumns(stmt)
	fmt.Printf("提取的列名: %v\n", columns)

	// 3. 测试UPDATE语句解析
	fmt.Println("\n\n3. UPDATE语句解析")
	fmt.Println("================")
	updateSQL := "UPDATE users SET name = '李四', age = 26 WHERE id = 1 AND email = 'zhangsan@example.com'"
	stmt, _ = parser.Parse(updateSQL)
	fmt.Printf("语句类型: %s\n", stmt.Type())
	fmt.Printf("原始SQL: %s\n", stmt.Raw())

	// 提取表名
	tables, _ = visitor.ExtractTableNames(stmt)
	fmt.Printf("\n提取的表名: %v\n", tables)

	// 提取列名
	columns, _ = visitor.ExtractColumns(stmt)
	fmt.Printf("提取的列名: %v\n", columns)

	// 4. 测试DELETE语句解析
	fmt.Println("\n\n4. DELETE语句解析")
	fmt.Println("================")
	deleteSQL := "DELETE FROM users WHERE id = 1 AND age > 30"
	stmt, _ = parser.Parse(deleteSQL)
	fmt.Printf("语句类型: %s\n", stmt.Type())
	fmt.Printf("原始SQL: %s\n", stmt.Raw())

	// 提取表名
	tables, _ = visitor.ExtractTableNames(stmt)
	fmt.Printf("\n提取的表名: %v\n", tables)

	// 5. 测试CREATE TABLE语句解析
	fmt.Println("\n\n5. CREATE TABLE语句解析")
	fmt.Println("====================")
	createTableSQL := `CREATE TABLE users (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    age INT DEFAULT 0,
    email VARCHAR(255) UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
)`
	stmt, _ = parser.Parse(createTableSQL)
	fmt.Printf("语句类型: %s\n", stmt.Type())

	// 提取表名
	tables, _ = visitor.ExtractTableNames(stmt)
	fmt.Printf("\n提取的表名: %v\n", tables)

	// 提取列名
	columns, _ = visitor.ExtractColumns(stmt)
	fmt.Printf("提取的列名: %v\n", columns)

	// 6. 测试CREATE INDEX语句解析
	fmt.Println("\n\n6. CREATE INDEX语句解析")
	fmt.Println("====================")
	createIndexSQL := "CREATE UNIQUE INDEX idx_email ON users (email)"
	stmt, _ = parser.Parse(createIndexSQL)
	fmt.Printf("语句类型: %s\n", stmt.Type())
	fmt.Printf("原始SQL: %s\n", stmt.Raw())

	// 提取表名
	tables, _ = visitor.ExtractTableNames(stmt)
	fmt.Printf("\n提取的表名: %v\n", tables)

	// 提取索引信息
	indexInfo, _ := visitor.ExtractIndexInfo(stmt)
	fmt.Println("\n提取的索引信息:")
	for key, value := range indexInfo {
		fmt.Printf("  %s: %v\n", key, value)
	}

	// 7. 测试CREATE DATABASE语句解析
	fmt.Println("\n\n7. CREATE DATABASE语句解析")
	fmt.Println("======================")
	createDatabaseSQL := "CREATE DATABASE test_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"
	stmt, _ = parser.Parse(createDatabaseSQL)
	fmt.Printf("语句类型: %s\n", stmt.Type())
	fmt.Printf("原始SQL: %s\n", stmt.Raw())

	fmt.Println("\n\n=== 示例结束 ===")
}
