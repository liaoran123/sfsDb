package util

import (
	"testing"
)

// TestCurrentBytesAll 测试当前Bytes.go文件的所有方法
func TestCurrentBytesAll(t *testing.T) {
	t.Run("Bool", func(t *testing.T) {
		tests := []struct {
			name     string
			input    []byte
		}{
			{"true", []byte{1}},
			{"false", []byte{0}},
			{"non-zero", []byte{2}},
			{"empty", []byte{}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := Bytes(tt.input).Bool()
				t.Logf("Bytes(%v).Bool() = %v", tt.input, result)
			})
		}
	})

	t.Run("Int8", func(t *testing.T) {
		tests := []struct {
			name     string
			input    []byte
		}{
			{"zero", []byte{0}},
			{"positive", []byte{42}},
			{"negative", []byte{255}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := Bytes(tt.input).Int8()
				t.Logf("Bytes(%v).Int8() = %v", tt.input, result)
			})
		}
	})

	t.Run("Uint8", func(t *testing.T) {
		tests := []struct {
			name     string
			input    []byte
		}{
			{"zero", []byte{0}},
			{"positive", []byte{42}},
			{"max", []byte{255}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := Bytes(tt.input).Uint8()
				t.Logf("Bytes(%v).Uint8() = %v", tt.input, result)
			})
		}
	})

	t.Run("Int16", func(t *testing.T) {
		tests := []struct {
			name     string
			input    []byte
		}{
			{"zero", []byte{0, 0}},
			{"positive", []byte{0, 42}},
			{"max", []byte{127, 255}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := Bytes(tt.input).Int16()
				t.Logf("Bytes(%v).Int16() = %v", tt.input, result)
			})
		}
	})

	t.Run("Uint16", func(t *testing.T) {
		tests := []struct {
			name     string
			input    []byte
		}{
			{"zero", []byte{0, 0}},
			{"positive", []byte{0, 42}},
			{"max", []byte{255, 255}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := Bytes(tt.input).Uint16()
				t.Logf("Bytes(%v).Uint16() = %v", tt.input, result)
			})
		}
	})

	t.Run("Int32", func(t *testing.T) {
		tests := []struct {
			name     string
			input    []byte
		}{
			{"zero", []byte{0, 0, 0, 0}},
			{"positive", []byte{0, 0, 0, 42}},
			{"max", []byte{127, 255, 255, 255}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := Bytes(tt.input).Int32()
				t.Logf("Bytes(%v).Int32() = %v", tt.input, result)
			})
		}
	})

	t.Run("Uint32", func(t *testing.T) {
		tests := []struct {
			name     string
			input    []byte
		}{
			{"zero", []byte{0, 0, 0, 0}},
			{"positive", []byte{0, 0, 0, 42}},
			{"max", []byte{255, 255, 255, 255}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := Bytes(tt.input).Uint32()
				t.Logf("Bytes(%v).Uint32() = %v", tt.input, result)
			})
		}
	})

	t.Run("Int64", func(t *testing.T) {
		tests := []struct {
			name     string
			input    []byte
		}{
			{"zero", []byte{0, 0, 0, 0, 0, 0, 0, 0}},
			{"positive", []byte{0, 0, 0, 0, 0, 0, 0, 42}},
			{"max", []byte{127, 255, 255, 255, 255, 255, 255, 255}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := Bytes(tt.input).Int64()
				t.Logf("Bytes(%v).Int64() = %v", tt.input, result)
			})
		}
	})

	t.Run("Uint64", func(t *testing.T) {
		tests := []struct {
			name     string
			input    []byte
		}{
			{"zero", []byte{0, 0, 0, 0, 0, 0, 0, 0}},
			{"positive", []byte{0, 0, 0, 0, 0, 0, 0, 42}},
			{"max", []byte{255, 255, 255, 255, 255, 255, 255, 255}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := Bytes(tt.input).Uint64()
				t.Logf("Bytes(%v).Uint64() = %v", tt.input, result)
			})
		}
	})

	t.Run("String", func(t *testing.T) {
		tests := []struct {
			name     string
			input    []byte
		}{
			{"empty", []byte{}},
			{"hello", []byte("hello")},
			{"chinese", []byte("你好")},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := Bytes(tt.input).String()
				t.Logf("Bytes(%v).String() = %q", tt.input, result)
			})
		}
	})

	t.Run("ToAny", func(t *testing.T) {
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
	})
}