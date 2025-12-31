package util

import (
	"testing"
	"time"
)

func TestStrToAny(t *testing.T) {
	// 准备测试数据
	intVal := 123
	int64Val := int64(1234567890)
	int32Val := int32(123456789)
	int16Val := int16(32767)
	int8Val := int8(127)
	uintVal := uint(456)
	uint64Val := uint64(1234567890123456789)
	uint32Val := uint32(123456789)
	uint16Val := uint16(65535)
	uint8Val := uint8(255)
	timeVal := time.Date(2023, 10, 25, 14, 30, 45, 0, time.Local)
	boolVal := true
	float64Val := 3.1415926535
	float32Val := float32(2.71828)
	complex64Val := complex64(1 + 2i)
	complex128Val := 3 + 4i
	stringVal := "hello world"

	// 测试用例
	tests := []struct {
		name     string
		input    string
		target   any
		expected any
	}{
		{"int", "123", int(0), intVal},
		{"int64", "1234567890", int64(0), int64Val},
		{"int32", "123456789", int32(0), int32Val},
		{"int16", "32767", int16(0), int16Val},
		{"int8", "127", int8(0), int8Val},
		{"uint", "456", uint(0), uintVal},
		{"uint64", "1234567890123456789", uint64(0), uint64Val},
		{"uint32", "123456789", uint32(0), uint32Val},
		{"uint16", "65535", uint16(0), uint16Val},
		{"uint8", "255", uint8(0), uint8Val},
		{"time.Time", "2023-10-25 14:30:45", time.Time{}, timeVal},
		{"bool", "true", bool(false), boolVal},
		{"float64", "3.1415926535", float64(0), float64Val},
		{"float32", "2.71828", float32(0), float32Val},
		{"complex64", "(1+2i)", complex64(0), complex64Val},
		{"complex128", "(3+4i)", complex128(0), complex128Val},
		{"string", "hello world", "", stringVal},
	}

	// 执行测试
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StrToAny(tt.input, tt.target)
			if result != tt.expected {
				t.Errorf("StrToAny(%q, %T) = %v (%T), want %v (%T)", tt.input, tt.target, result, result, tt.expected, tt.expected)
			}
		})
	}
}

func TestAnyToStr(t *testing.T) {
	// 准备测试数据
	intVal := 123
	int64Val := int64(1234567890)
	int32Val := int32(123456789)
	int16Val := int16(32767)
	int8Val := int8(127)
	uintVal := uint(456)
	uint64Val := uint64(1234567890123456789)
	uint32Val := uint32(123456789)
	uint16Val := uint16(65535)
	uint8Val := uint8(255)
	timeVal := time.Date(2023, 10, 25, 14, 30, 45, 0, time.Local)
	boolVal := true
	float64Val := 3.1415926535
	float32Val := float32(2.71828)
	complex64Val := complex64(1 + 2i)
	complex128Val := 3 + 4i
	stringVal := "hello world"

	// 测试用例
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"int", intVal, "123"},
		{"int64", int64Val, "1234567890"},
		{"int32", int32Val, "123456789"},
		{"int16", int16Val, "32767"},
		{"int8", int8Val, "127"},
		{"uint", uintVal, "456"},
		{"uint64", uint64Val, "1234567890123456789"},
		{"uint32", uint32Val, "123456789"},
		{"uint16", uint16Val, "65535"},
		{"uint8", uint8Val, "255"},
		{"time.Time", timeVal, "2023-10-25 14:30:45"},
		{"bool", boolVal, "true"},
		{"float64", float64Val, "3.141593"}, // 注意：格式化会四舍五入到6位小数
		{"float32", float32Val, "2.718280"}, // 注意：格式化会补零到6位小数
		{"complex64", complex64Val, "(1.000000+2.000000i)"},
		{"complex128", complex128Val, "(3.000000+4.000000i)"},
		{"string", stringVal, "hello world"},
	}

	// 执行测试
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AnyToStr(tt.input)
			if result != tt.expected {
				t.Errorf("AnyToStr(%v (%T)) = %q, want %q", tt.input, tt.input, result, tt.expected)
			}
		})
	}
}
