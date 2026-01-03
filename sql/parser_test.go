package sql

import (
	"testing"
)

// TestParser 测试SQL解析器的基本功能
func TestParser(t *testing.T) {
	parser := NewParser()

	// 测试用例：SELECT语句
	t.Run("SelectStatement", func(t *testing.T) {
		sql := "SELECT id, name, age FROM users WHERE age > 18 ORDER BY id DESC"
		stmt, err := parser.Parse(sql)
		if err != nil {
			t.Fatalf("Failed to parse SELECT statement: %v", err)
		}

		if stmt.Type() != "SELECT" {
			t.Errorf("Expected statement type 'SELECT', got '%s'", stmt.Type())
		}

		if stmt.Raw() != sql {
			t.Errorf("Expected raw SQL '%s', got '%s'", sql, stmt.Raw())
		}
	})

	// 测试用例：INSERT语句
	t.Run("InsertStatement", func(t *testing.T) {
		sql := "INSERT INTO users (name, age) VALUES ('张三', 25)"
		stmt, err := parser.Parse(sql)
		if err != nil {
			t.Fatalf("Failed to parse INSERT statement: %v", err)
		}

		if stmt.Type() != "INSERT" {
			t.Errorf("Expected statement type 'INSERT', got '%s'", stmt.Type())
		}
	})

	// 测试用例：UPDATE语句
	t.Run("UpdateStatement", func(t *testing.T) {
		sql := "UPDATE users SET age = 26 WHERE id = 1"
		stmt, err := parser.Parse(sql)
		if err != nil {
			t.Fatalf("Failed to parse UPDATE statement: %v", err)
		}

		if stmt.Type() != "UPDATE" {
			t.Errorf("Expected statement type 'UPDATE', got '%s'", stmt.Type())
		}
	})

	// 测试用例：DELETE语句
	t.Run("DeleteStatement", func(t *testing.T) {
		sql := "DELETE FROM users WHERE id = 1"
		stmt, err := parser.Parse(sql)
		if err != nil {
			t.Fatalf("Failed to parse DELETE statement: %v", err)
		}

		if stmt.Type() != "DELETE" {
			t.Errorf("Expected statement type 'DELETE', got '%s'", stmt.Type())
		}
	})
}

// TestVisitor 测试AST访问器功能
func TestVisitor(t *testing.T) {
	parser := NewParser()
	visitor := NewVisitor()

	// 测试用例：提取表名
	t.Run("ExtractTableNames", func(t *testing.T) {
		// SELECT语句
		sql := "SELECT id, name FROM users WHERE age > 18"
		stmt, err := parser.Parse(sql)
		if err != nil {
			t.Fatalf("Failed to parse SQL: %v", err)
		}

		tables, err := visitor.ExtractTableNames(stmt)
		if err != nil {
			t.Fatalf("Failed to extract table names: %v", err)
		}

		expected := []string{"users"}
		if len(tables) != len(expected) {
			t.Fatalf("Expected %d tables, got %d", len(expected), len(tables))
		}

		for i, table := range tables {
			if table != expected[i] {
				t.Errorf("Expected table '%s', got '%s'", expected[i], table)
			}
		}

		// INSERT语句
		sql = "INSERT INTO users (name, age) VALUES ('张三', 25)"
		stmt, err = parser.Parse(sql)
		if err != nil {
			t.Fatalf("Failed to parse SQL: %v", err)
		}

		tables, err = visitor.ExtractTableNames(stmt)
		if err != nil {
			t.Fatalf("Failed to extract table names: %v", err)
		}

		expected = []string{"users"}
		if len(tables) != len(expected) {
			t.Fatalf("Expected %d tables, got %d", len(expected), len(tables))
		}
	})

	// 测试用例：提取列名
	t.Run("ExtractColumns", func(t *testing.T) {
		sql := "SELECT id, name, age FROM users WHERE age > 18"
		stmt, err := parser.Parse(sql)
		if err != nil {
			t.Fatalf("Failed to parse SQL: %v", err)
		}

		columns, err := visitor.ExtractColumns(stmt)
		if err != nil {
			t.Fatalf("Failed to extract columns: %v", err)
		}

		// 预期的列名集合
		expectedMap := map[string]bool{
			"id":  true,
			"name": true,
			"age":  true,
		}

		if len(columns) != len(expectedMap) {
			t.Fatalf("Expected %d columns, got %d", len(expectedMap), len(columns))
		}

		// 检查所有预期列名是否存在
		columnMap := make(map[string]bool)
		for _, column := range columns {
			columnMap[column] = true
		}

		for expected := range expectedMap {
			if !columnMap[expected] {
				t.Errorf("Expected column '%s' not found", expected)
			}
		}
	})
}

// TestFormatSQL 测试SQL格式化功能
func TestFormatSQL(t *testing.T) {
	// 测试用例：简单SELECT语句
	t.Run("SimpleSelect", func(t *testing.T) {
		original := "SELECT id, name FROM users"
		formatted, err := FormatSQL(original)
		if err != nil {
			t.Fatalf("Failed to format SQL: %v", err)
		}

		// 简单的格式化验证，实际格式化结果可能因库而异
		if formatted == "" {
			t.Error("Expected formatted SQL, got empty string")
		}
	})

	// 测试用例：带有WHERE条件的SELECT语句
	t.Run("SelectWithWhere", func(t *testing.T) {
		original := "SELECT * FROM users WHERE age > 18 AND name LIKE '%张%'"
		formatted, err := FormatSQL(original)
		if err != nil {
			t.Fatalf("Failed to format SQL: %v", err)
		}

		if formatted == "" {
			t.Error("Expected formatted SQL, got empty string")
		}
	})
}
