package util

import (
	"testing"
)

// TestCurrentBytesInt 测试当前的Int方法
func TestCurrentBytesInt(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
	}{
		{"zero", []byte{0, 0, 0, 0}},
		{"positive", []byte{0, 0, 0, 42}},
		{"max int32", []byte{127, 255, 255, 255}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Bytes(tt.input).Int()
			t.Logf("Bytes(%v).Int() = %v", tt.input, result)
		})
	}
}

// TestCurrentBytesInt64 测试当前的Int64方法
func TestCurrentBytesInt64(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
	}{
		{"zero", []byte{0, 0, 0, 0, 0, 0, 0, 0}},
		{"positive", []byte{0, 0, 0, 0, 0, 0, 0, 42}},
		{"max int64", []byte{127, 255, 255, 255, 255, 255, 255, 255}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Bytes(tt.input).Int64()
			t.Logf("Bytes(%v).Int64() = %v", tt.input, result)
		})
	}
}

// TestCurrentBytesUint32 测试当前的Uint32方法
func TestCurrentBytesUint32(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
	}{
		{"zero", []byte{0, 0, 0, 0}},
		{"positive", []byte{0, 0, 0, 42}},
		{"max uint32", []byte{255, 255, 255, 255}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Bytes(tt.input).Uint32()
			t.Logf("Bytes(%v).Uint32() = %v", tt.input, result)
		})
	}
}

// TestCurrentBytesUint64 测试当前的Uint64方法
func TestCurrentBytesUint64(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
	}{
		{"zero", []byte{0, 0, 0, 0, 0, 0, 0, 0}},
		{"positive", []byte{0, 0, 0, 0, 0, 0, 0, 42}},
		{"max uint64", []byte{255, 255, 255, 255, 255, 255, 255, 255}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Bytes(tt.input).Uint64()
			t.Logf("Bytes(%v).Uint64() = %v", tt.input, result)
		})
	}
}

// TestCurrentBytesToAny 测试当前的ToAny方法
func TestCurrentBytesToAny(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		target   any
	}{
		{"string", []byte("hello"), ""},
		{"int", []byte{0, 0, 0, 42}, 0},
		{"bool", []byte{1}, false},
		{"float64", []byte("3.14159"), float64(0)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Bytes(tt.input).ToAny(tt.target)
			t.Logf("Bytes(%v).ToAny(%T) = %v", tt.input, tt.target, result)
		})
	}
}