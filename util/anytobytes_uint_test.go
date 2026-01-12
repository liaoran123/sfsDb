package util

import (
	"fmt"
	"testing"
)

func TestAnyToBytesUint(t *testing.T) {
	// 测试 uint 类型
	testValue := uint(123456789)
	bytes := AnyToBytes(testValue)
	expectedLength := TypeSize(testValue)
	actualLength := len(bytes)

	fmt.Printf("Testing uint type:\n")
	fmt.Printf("Value: %d\n", testValue)
	fmt.Printf("Expected length: %d bytes\n", expectedLength)
	fmt.Printf("Actual length: %d bytes\n", actualLength)
	fmt.Printf("Bytes: %v\n", bytes)

	if actualLength != expectedLength {
		t.Errorf("Length mismatch: expected %d bytes, got %d bytes", expectedLength, actualLength)
	}

	// 测试 StrToAny 函数的错误处理
	testCases := []struct {
		name        string
		input       string
		value       any
		shouldError bool
	}{
		{"valid int", "123", int(0), false},
		{"invalid int", "abc", int(0), true},
		{"valid float", "123.456", float64(0), false},
		{"invalid float", "abc.123", float64(0), true},
		{"valid bool", "true", bool(false), false},
		{"invalid bool", "maybe", bool(false), true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := StrToAny(tc.input, tc.value)
			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error for input '%s', but got none", tc.input)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input '%s': %v", tc.input, err)
				}
				fmt.Printf("Input '%s' converted to: %v (type: %T)\n", tc.input, result, result)
			}
		})
	}
}
