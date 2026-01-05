package sql

import (
	"testing"
)

// TestParser 测试SQL解析器
func TestParser(t *testing.T) {
	parser := NewParser()

	// 测试用例：各种SQL语句类型
	testCases := []struct {
		name     string
		sql      string
		expected string
	}{
		{
			name:     "SELECT Statement",
			sql:      "SELECT id, name FROM users WHERE age > 18",
			expected: "SELECT",
		},
		{
			name:     "INSERT Statement",
			sql:      "INSERT INTO users (id, name) VALUES (1, 'test')",
			expected: "INSERT",
		},
		{
			name:     "UPDATE Statement",
			sql:      "UPDATE users SET name = 'new' WHERE id = 1",
			expected: "UPDATE",
		},
		{
			name:     "DELETE Statement",
			sql:      "DELETE FROM users WHERE id = 1",
			expected: "DELETE",
		},
		{
			name:     "CREATE TABLE Statement",
			sql:      "CREATE TABLE users (id INT PRIMARY KEY, name VARCHAR(255))",
			expected: "CREATE TABLE",
		},
		{
			name:     "CREATE INDEX Statement",
			sql:      "CREATE INDEX idx_name ON users (name)",
			expected: "CREATE INDEX",
		},
		{
			name:     "CREATE DATABASE Statement",
			sql:      "CREATE DATABASE test_db",
			expected: "CREATE DATABASE",
		},
	}

	// 运行测试用例
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stmt, err := parser.Parse(tc.sql)
			if err != nil {
				t.Fatalf("解析SQL语句失败: %v", err)
			}

			if stmt.Type() != tc.expected {
				t.Errorf("SQL语句类型不匹配，期望: %s, 实际: %s", tc.expected, stmt.Type())
			}

			if stmt.Raw() != tc.sql {
				t.Errorf("原始SQL语句不匹配，期望: %s, 实际: %s", tc.sql, stmt.Raw())
			}
		})
	}
}
