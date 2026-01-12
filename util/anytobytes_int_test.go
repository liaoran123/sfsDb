package util

import (
	"fmt"
	"testing"
)

func TestAnyToBytesIntLength(t *testing.T) {
	// 测试 int 类型
	testValue := 42
	bytes := AnyToBytes(testValue)
	expectedLength := TypeSize(testValue) // 已经是字节长度
	actualLength := len(bytes)

	fmt.Printf("Testing int type:\n")
	fmt.Printf("Value: %d\n", testValue)
	fmt.Printf("Expected length: %d bytes\n", expectedLength)
	fmt.Printf("Actual length: %d bytes\n", actualLength)
	fmt.Printf("Bytes: %v\n", bytes)

	if actualLength != expectedLength {
		t.Errorf("Length mismatch: expected %d bytes, got %d bytes", expectedLength, actualLength)
	}

	// 测试其他整型类型
	testCases := []struct {
		name  string
		value any
	}{
		{"int8", int8(100)},
		{"uint8", uint8(200)},
		{"int16", int16(30000)},
		{"uint16", uint16(50000)},
		{"int32", int32(1000000)},
		{"uint32", uint32(2000000000)},
		{"int64", int64(9000000000000000000)},
		{"uint64", uint64(18000000000000000000)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bytes := AnyToBytes(tc.value)
			expectedLength := TypeSize(tc.value)
			actualLength := len(bytes)

			fmt.Printf("\nTesting %s type:\n", tc.name)
			fmt.Printf("Value: %v\n", tc.value)
			fmt.Printf("Expected length: %d bytes\n", expectedLength)
			fmt.Printf("Actual length: %d bytes\n", actualLength)
			fmt.Printf("Bytes: %v\n", bytes)

			if actualLength != expectedLength {
				t.Errorf("Length mismatch for %s: expected %d bytes, got %d bytes", tc.name, expectedLength, actualLength)
			}
		})
	}
}
