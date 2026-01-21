package engine

import (
	"testing"
)

func TestValueJoiner(t *testing.T) {
	// 创建值拼接器
	joiner := NewValueJoiner()

	// 测试基本拼接功能
	testJoin := joiner.Join("sys", "table", "test")
	if testJoin != "sys-table-test" {
		t.Errorf("Expected 'sys-table-test', got '%s'", testJoin)
	}

	// 测试表对象键生成
	tableKey := joiner.GenerateTableKey("users")
	if tableKey != "sys-table-users" {
		t.Errorf("Expected 'sys-table-users', got '%s'", tableKey)
	}

	// 测试索引对象键生成
	indexKey := joiner.GenerateIndexKey(1, "name_index")
	if indexKey != "sys-1-idx-name_index" {
		t.Errorf("Expected 'sys-1-idx-name_index', got '%s'", indexKey)
	}

	// 测试字段对象键生成
	fieldKey := joiner.GenerateFieldKey(1, "name")
	if fieldKey != "sys-1-field-name" {
		t.Errorf("Expected 'sys-1-field-name', got '%s'", fieldKey)
	}

	// 测试表计数器键生成
	tableCounter := joiner.GenerateTableCounterKey()
	if tableCounter != "sys-table" {
		t.Errorf("Expected 'sys-table', got '%s'", tableCounter)
	}

	// 测试索引计数器键生成
	indexCounter := joiner.GenerateIndexCounterKey(1)
	if indexCounter != "sys-1-idx" {
		t.Errorf("Expected 'sys-1-idx', got '%s'", indexCounter)
	}

	// 测试字段计数器键生成
	fieldCounter := joiner.GenerateFieldCounterKey(1)
	if fieldCounter != "sys-1-field" {
		t.Errorf("Expected 'sys-1-field', got '%s'", fieldCounter)
	}
}
