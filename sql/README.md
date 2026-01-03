# SQL Parser for sfsDb

基于 vitess.io/vitess/go/vt/sqlparser 的SQL解析架构，为sfsDb数据库引擎提供SQL解析功能。

## 目录结构

```
sql/
├── parser.go       # 核心解析器，负责解析SQL语句
├── visitor.go      # AST访问器，用于提取表名、列名等信息
├── parser_test.go  # 测试用例
├── example.go      # 使用示例
└── README.md       # 本文件
```

## 依赖

- **vitess.io/vitess/go/vt/sqlparser**: 用于SQL语句的解析和AST生成

## 安装

```bash
# 安装vitess/sqlparser依赖
go get vitess.io/vitess/go/vt/sqlparser
```

## 核心组件

### Parser

解析SQL语句并返回抽象语法树(AST)。

```go
// 创建解析器实例
parser := sql.NewParser()

// 解析SQL语句
stmt, err := parser.Parse("SELECT id, name FROM users")
if err != nil {
    panic(err)
}
```

### Statement

表示解析后的SQL语句，提供统一的接口访问不同类型的SQL语句。

```go
// 获取语句类型
stmtType := stmt.Type() // 返回 "SELECT", "INSERT", "UPDATE", "DELETE" 或 "UNKNOWN"

// 获取原始SQL
rawSQL := stmt.Raw()

// 根据类型处理不同语句
switch s := stmt.(type) {
case *sql.SelectStatement:
    // 处理SELECT语句
case *sql.InsertStatement:
    // 处理INSERT语句
// 其他类型...
}
```

### Visitor

用于遍历AST并提取信息，如表名、列名等。

```go
// 创建访问器
visitor := sql.NewVisitor()

// 提取表名
tables, err := visitor.ExtractTableNames(stmt)
if err != nil {
    panic(err)
}

// 提取列名
columns, err := visitor.ExtractColumns(stmt)
if err != nil {
    panic(err)
}
```

## 使用示例

### 基本用法

```go
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
```

### 支持的SQL语句类型

#### SELECT语句
```go
// 简单查询
stmt, err := parser.Parse("SELECT id, name FROM users")

// 带WHERE条件的查询
stmt, err := parser.Parse("SELECT * FROM users WHERE age > 18")

// 带ORDER BY和LIMIT的查询
stmt, err := parser.Parse("SELECT id, name FROM users ORDER BY id DESC LIMIT 10")
```

#### INSERT语句
```go
stmt, err := parser.Parse("INSERT INTO users (name, age) VALUES ('张三', 25)")
```

#### UPDATE语句
```go
stmt, err := parser.Parse("UPDATE users SET age = 26 WHERE id = 1")
```

#### DELETE语句
```go
stmt, err := parser.Parse("DELETE FROM users WHERE id = 1")
```

## 功能特性

### 1. 语句类型识别

自动识别SQL语句类型，支持SELECT、INSERT、UPDATE、DELETE等常见语句。

### 2. 表名提取

从SQL语句中提取涉及的表名，支持多种语句类型。

### 3. 列名提取

从SELECT语句中提取查询的列名，忽略别名和通配符。

### 4. SQL格式化

提供基本的SQL格式化功能。

```go
formattedSQL, err := sql.FormatSQL("select id,name from users where age>18")
// 返回: "SELECT id, name FROM users WHERE age > 18"
```

## 高级用法

### 自定义访问器

可以通过实现自己的访问器来提取更复杂的信息：

```go
// 自定义访问器类型
type MyVisitor struct {
    // 自定义字段
}

// 实现Visit方法
func (v *MyVisitor) Visit(node sqlparser.SQLNode) (kontinue bool, err error) {
    // 自定义处理逻辑
    // ...
    return true, nil
}
```

## 测试

运行测试用例：

```bash
go test ./sql -v
```

## 注意事项

1. **大小写敏感性**：SQL解析器会保留原始SQL的大小写，表名和列名的提取也会保持原始大小写。

2. **支持的SQL语法**：当前支持基本的SQL语法，复杂的查询如子查询、连接查询等可能需要进一步扩展。

3. **错误处理**：解析错误会返回详细的错误信息，便于调试。

4. **性能考虑**：对于大量SQL的解析，建议复用解析器实例，避免频繁创建和销毁。

## 扩展建议

1. **支持更多SQL语法**：如子查询、连接查询、CTE等。

2. **查询优化**：基于AST实现查询优化，如谓词下推、索引选择等。

3. **执行计划生成**：根据AST生成执行计划，用于后续的查询执行。

4. **SQL验证**：验证SQL语句的合法性，如表和列是否存在等。

5. **参数化查询支持**：支持预编译语句和参数绑定。

## 许可证

MIT

## 贡献

欢迎提交Issue和Pull Request，共同改进SQL解析架构。

## 联系信息

如有问题，请联系项目维护者。
