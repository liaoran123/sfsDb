package util

import (
	"testing"
)

// 测试 AnyToBytes 函数的 JSON 序列化功能
func TestAnyToBytesJSON(t *testing.T) {
	// 定义一个测试结构体
	type TestStruct struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email,omitempty"`
	}

	// 测试用例1: 结构体序列化
	t.Run("struct", func(t *testing.T) {
		testData := TestStruct{
			Name: "张三",
			Age:  25,
		}

		result := AnyToBytes(testData)
		expected := []byte(`{"name":"张三","age":25}`)

		if string(result) != string(expected) {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	// 测试用例2: 结构体带空字段
	t.Run("struct_with_empty_field", func(t *testing.T) {
		testData := TestStruct{
			Name:  "李四",
			Age:   30,
			Email: "", // 应该被忽略
		}

		result := AnyToBytes(testData)
		expected := []byte(`{"name":"李四","age":30}`)

		if string(result) != string(expected) {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	// 测试用例3: map 类型
	t.Run("map", func(t *testing.T) {
		testData := map[string]any{
			"key1": "value1",
			"key2": 123,
		}

		result := AnyToBytes(testData)
		expected := []byte(`{"key1":"value1","key2":123}`)

		if string(result) != string(expected) {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	// 测试用例4: 切片类型
	t.Run("slice", func(t *testing.T) {
		testData := []string{"a", "b", "c"}

		result := AnyToBytes(testData)
		expected := []byte(`["a","b","c"]`)

		if string(result) != string(expected) {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})
}