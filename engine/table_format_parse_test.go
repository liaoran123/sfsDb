package engine

import (
	"testing"
)

// 测试格式化和反格式化函数是否匹配
func TestFormatParseMatch(t *testing.T) {
	// 创建表
	table, err := TableNew("test_format_parse")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	}
	if err := table.SetFields(fields); err != nil {
		t.Fatalf("Failed to set table fields: %v", err)
	}

	// 创建测试数据
	testData := map[string][]byte{
		"id":   []byte{0, 0, 0, 0, 0, 0, 0, 1},
		"name": []byte("testuser"),
		"age":  []byte{0, 0, 0, 25},
	}

	// 测试格式化
	formatted := table.FormatRecord(&testData)
	if len(formatted) == 0 {
		t.Fatal("Failed to format record")
	}

	// 测试反格式化
	pk := table.GetPrimaryKey()
	parsed, err := pk.Parse(table.fieldsid, formatted)
	if err != nil {
		t.Fatalf("Failed to parse record: %v", err)
	}

	// 验证解析结果
	if len(*parsed) == 0 {
		t.Fatal("Failed to parse any fields")
	}

	// 检查字段值是否匹配
	for field, expected := range testData {
		actual, ok := (*parsed)[field]
		if !ok {
			t.Errorf("Field '%s' not found in parsed data", field)
			continue
		}

		if string(actual) != string(expected) {
			t.Errorf("Field '%s': expected '%s', got '%s'", field, expected, actual)
		}
	}

	t.Log("Format and parse functions match correctly!")
}

// 测试包含需要转义值的格式化和反格式化
func TestFormatParseWithEscape(t *testing.T) {
	// 创建表
	table, err := TableNew("test_format_parse_escape")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{
		"id":   0,
		"name": "",
		"data": "",
	}
	if err := table.SetFields(fields); err != nil {
		t.Fatalf("Failed to set table fields: %v", err)
	}

	// 创建包含需要转义的值（包含-fieldname-格式的数据）
	testData := map[string][]byte{
		"id":   []byte{0, 0, 0, 0, 0, 0, 0, 1},
		"name": []byte("testuser"),
		"data": []byte("test-data-with-field-name"),
	}

	// 测试格式化
	formatted := table.FormatRecord(&testData)
	if len(formatted) == 0 {
		t.Fatal("Failed to format record")
	}

	// 测试反格式化
	pk := table.GetPrimaryKey()
	parsed, err := pk.Parse(table.fieldsid, formatted)
	if err != nil {
		t.Fatalf("Failed to parse record: %v", err)
	}

	// 验证解析结果
	if len(*parsed) == 0 {
		t.Fatal("Failed to parse any fields")
	}

	// 检查字段值是否匹配
	for field, expected := range testData {
		actual, ok := (*parsed)[field]
		if !ok {
			t.Errorf("Field '%s' not found in parsed data", field)
			continue
		}

		if string(actual) != string(expected) {
			t.Errorf("Field '%s': expected '%s', got '%s'", field, expected, actual)
		}
	}

	t.Log("Format and parse functions work correctly with escaped values!")
}
