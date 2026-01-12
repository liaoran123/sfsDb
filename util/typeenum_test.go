package util

import (
	"fmt"
	"testing"
	"time"
)

func TestTypeEnum(t *testing.T) {
	// 测试 GetTypeInstance 函数
	testCases := []struct {
		name     string
		enum     TypeEnum
		expected string
	}{
		{"TypeInt", TypeInt, "int"},
		{"TypeInt8", TypeInt8, "int8"},
		{"TypeInt16", TypeInt16, "int16"},
		{"TypeInt32", TypeInt32, "int32"},
		{"TypeInt64", TypeInt64, "int64"},
		{"TypeUint", TypeUint, "uint"},
		{"TypeUint8", TypeUint8, "uint8"},
		{"TypeUint16", TypeUint16, "uint16"},
		{"TypeUint32", TypeUint32, "uint32"},
		{"TypeUint64", TypeUint64, "uint64"},
		{"TypeFloat32", TypeFloat32, "float32"},
		{"TypeFloat64", TypeFloat64, "float64"},
		{"TypeBool", TypeBool, "bool"},
		{"TypeString", TypeString, "string"},
		{"TypeTime", TypeTime, "time.Time"},
		{"TypeComplex64", TypeComplex64, "complex64"},
		{"TypeComplex128", TypeComplex128, "complex128"},
	}

	fmt.Println("Testing GetTypeInstance:")
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			instance := GetTypeInstance(tc.enum)
			fmt.Printf("%s: %T\n", tc.name, instance)
		})
	}

	// 测试 GetTypeEnum 函数
	fmt.Println("\nTesting GetTypeEnum:")
	testValues := []any{
		int(123),
		int8(123),
		int16(123),
		int32(123),
		int64(123),
		uint(123),
		uint8(123),
		uint16(123),
		uint32(123),
		uint64(123),
		float32(123.456),
		float64(123.456),
		bool(true),
		string("test"),
		time.Now(),
		complex64(1+2i),
		complex128(1+2i),
	}

	for _, v := range testValues {
		enum := GetTypeEnum(v)
		fmt.Printf("%T: %d\n", v, enum)
	}

	// 测试 ParseByType 函数
	fmt.Println("\nTesting ParseByType:")
	parseTestCases := []struct {
		name  string
		enum  TypeEnum
		input string
	}{
		{"int", TypeInt, "123"},
		{"int8", TypeInt8, "123"},
		{"int16", TypeInt16, "12345"},
		{"int32", TypeInt32, "123456789"},
		{"int64", TypeInt64, "123456789012345"},
		{"uint", TypeUint, "123456"},
		{"uint8", TypeUint8, "200"},
		{"uint16", TypeUint16, "50000"},
		{"uint32", TypeUint32, "2000000000"},
		{"uint64", TypeUint64, "18000000000000000000"},
		{"float32", TypeFloat32, "123.456"},
		{"float64", TypeFloat64, "123.456789"},
		{"bool", TypeBool, "true"},
		{"string", TypeString, "hello world"},
		{"time", TypeTime, "2024-01-01 12:00:00"},
	}

	for _, tc := range parseTestCases {
		result, err := ParseByType(tc.enum, tc.input)
		if err != nil {
			fmt.Printf("%s: Error: %v\n", tc.name, err)
		} else {
			fmt.Printf("%s: %v (%T)\n", tc.name, result, result)
		}
	}

	// 测试 ConvertToString 函数
	fmt.Println("\nTesting ConvertToString:")
	for _, v := range testValues {
		str := ConvertToString(v)
		fmt.Printf("%T: %s\n", v, str)
	}

	// 测试 ConvertToBytes 函数
	fmt.Println("\nTesting ConvertToBytes:")
	for _, v := range testValues {
		bytes := ConvertToBytes(v)
		fmt.Printf("%T: %v\n", v, bytes)
	}
}