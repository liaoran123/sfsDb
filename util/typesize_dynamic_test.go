package util

import (
	"fmt"
	"testing"
	"time"
)

// 自定义类型示例
type Person struct {
	Name string
	Age  int
}

// 计算 Person 类型大小的函数
func personSize(value any) int {
	// 简单实现：返回固定大小
	return 32 // 假设 Person 类型占用 32 字节
}

func TestTypeSizeDynamic(t *testing.T) {
	// 测试基本类型
	testCases := []struct {
		name  string
		value any
	}{
		{"bool", true},
		{"int", 123},
		{"int8", int8(123)},
		{"uint8", uint8(123)},
		{"int16", int16(12345)},
		{"uint16", uint16(12345)},
		{"int32", int32(123456789)},
		{"uint32", uint32(123456789)},
		{"int64", int64(123456789012345)},
		{"uint64", uint64(123456789012345)},
		{"float32", float32(123.456)},
		{"float64", float64(123.456)},
		{"time.Time", time.Now()},
	}

	fmt.Println("Testing basic types:")
	for _, tc := range testCases {
		size := TypeSize(tc.value)
		fmt.Printf("%s: %d bytes\n", tc.name, size)
	}

	// 测试自定义类型（注册前）
	person := Person{Name: "John", Age: 30}
	sizeBefore := TypeSize(person)
	fmt.Printf("\nPerson size before registration: %d bytes\n", sizeBefore)

	// 注册自定义类型
	RegisterTypeSize("util.Person", personSize)
	fmt.Println("Registered Person type")

	// 测试自定义类型（注册后）
	sizeAfter := TypeSize(person)
	fmt.Printf("Person size after registration: %d bytes\n", sizeAfter)

	if sizeAfter != personSize(person) {
		t.Errorf("Person size mismatch: expected %d, got %d", personSize(person), sizeAfter)
	}
}