package util

import (
	"fmt"
	"strconv"
	"testing"
)

func TestSizeOfType(t *testing.T) {
	// 测试各种整型类型
	testCases := []struct {
		name  string
		value any
		want  int
	}{
		{"int", int(0), strconv.IntSize / 8},
		{"int8", int8(0), 1},
		{"uint8", uint8(0), 1},
		{"int16", int16(0), 2},
		{"uint16", uint16(0), 2},
		{"int32", int32(0), 4},
		{"uint32", uint32(0), 4},
		{"int64", int64(0), 8},
		{"uint64", uint64(0), 8},
		{"non-integer", "string", 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := TypeSize(tc.value)
			if got != tc.want {
				t.Errorf("TypeSize(%v) = %d; want %d", tc.name, got, tc.want)
			}
			fmt.Printf("%s: %d bytes\n", tc.name, got)
		})
	}
}
