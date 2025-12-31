package util

import (
	"bytes"
	"testing"
)

func TestBytesBool(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected bool
	}{
		{"true", []byte{1}, true},
		{"false", []byte{0}, false},
		{"non-zero", []byte{2}, true},
		{"empty", []byte{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Bytes(tt.input).Bool()
			if result != tt.expected {
				t.Errorf("Bytes(%v).Bool() = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBytesInt(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected int
	}{
		{"zero", []byte{0, 0, 0, 0}, 0},
		{"positive", []byte{0, 0, 0, 42}, 42},
		{"max", []byte{127, 255, 255, 255}, 2147483647},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Bytes(tt.input).Int()
			if result != tt.expected {
				t.Errorf("Bytes(%v).Int() = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBytesInt64(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected int64
	}{
		{"zero", []byte{0, 0, 0, 0, 0, 0, 0, 0}, 0},
		{"positive", []byte{0, 0, 0, 0, 0, 0, 0, 42}, 42},
		{"max", []byte{127, 255, 255, 255, 255, 255, 255, 255}, 9223372036854775807},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Bytes(tt.input).Int64()
			if result != tt.expected {
				t.Errorf("Bytes(%v).Int64() = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBytesUint32(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected uint32
	}{
		{"zero", []byte{0, 0, 0, 0}, 0},
		{"positive", []byte{0, 0, 0, 42}, 42},
		{"max", []byte{255, 255, 255, 255}, 4294967295},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Bytes(tt.input).Uint32()
			if result != tt.expected {
				t.Errorf("Bytes(%v).Uint32() = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBytesUint64(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected uint64
	}{
		{"zero", []byte{0, 0, 0, 0, 0, 0, 0, 0}, 0},
		{"positive", []byte{0, 0, 0, 0, 0, 0, 0, 42}, 42},
		{"max", []byte{255, 255, 255, 255, 255, 255, 255, 255}, 18446744073709551615},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Bytes(tt.input).Uint64()
			if result != tt.expected {
				t.Errorf("Bytes(%v).Uint64() = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBytesToAny(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		target   any
		expected any
	}{
		{"string", []byte("hello"), "", "hello"},
		{"int", []byte{0, 0, 0, 42}, 0, 42},
		{"bool", []byte{1}, false, true},
		{"float64", []byte("3.14159"), float64(0), 3.14159},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Bytes(tt.input).ToAny(tt.target)
			if result != tt.expected {
				t.Errorf("Bytes(%v).ToAny(%T) = %v, want %v", tt.input, tt.target, result, tt.expected)
			}
		})
	}
}

func TestBytesString(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{"empty", []byte{}, ""},
		{"hello", []byte("hello"), "hello"},
		{"chinese", []byte("你好"), "你好"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Bytes(tt.input).String()
			if result != tt.expected {
				t.Errorf("Bytes(%v).String() = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBytesEscape(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{"empty", []byte{}, []byte{}},
		{"no-split", []byte("hello"), []byte("hello")},
		{"with-split", []byte("hello-world"), []byte("hello--world")},
		{"multiple-split", []byte("a-b-c"), []byte("a--b--c")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Bytes(tt.input).Escape()
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("Bytes(%q).Escape() = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBytesUnEscape(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{"empty", []byte{}, []byte{}},
		{"no-escape", []byte("hello"), []byte("hello")},
		{"with-escape", []byte("hello--world"), []byte("hello-world")},
		{"multiple-escape", []byte("a--b--c"), []byte("a-b-c")},
		{"single-split", []byte("hello-world"), []byte("hello-world")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Bytes(tt.input).UnEscape()
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("Bytes(%q).UnEscape() = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBytesJion(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		joins    [][]byte
		expected []byte
	}{
		{"empty", []byte{}, [][]byte{}, []byte{}},
		{"single-join", []byte("hello"), [][]byte{[]byte("world")}, []byte("helloworld")},
		{"multiple-join", []byte("a"), [][]byte{[]byte("b"), []byte("c")}, []byte("abc")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Bytes(tt.input).Jion(tt.joins...)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("Bytes(%q).Jion(%v) = %q, want %q", tt.input, tt.joins, result, tt.expected)
			}
		})
	}
}

func TestBytesSplit(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected [][]byte
	}{
		{"empty", []byte{}, [][]byte{}},
		{"no-split", []byte("hello"), [][]byte{[]byte("hello")}},
		{"with-split", []byte("hello-world"), [][]byte{[]byte("hello"), []byte("world")}},
		{"multiple-split", []byte("a-b-c"), [][]byte{[]byte("a"), []byte("b"), []byte("c")}},
		{"start-split", []byte("-hello"), [][]byte{[]byte("hello")}},
		{"end-split", []byte("hello-"), [][]byte{[]byte("hello")}},
		{"double-split", []byte("hello--world"), [][]byte{[]byte("hello-world")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Bytes(tt.input).Split()
			if len(result) != len(tt.expected) {
				t.Errorf("Bytes(%q).Split() length = %d, want %d", tt.input, len(result), len(tt.expected))
				return
			}
			for i := range result {
				if !bytes.Equal(result[i], tt.expected[i]) {
					t.Errorf("Bytes(%q).Split()[%d] = %q, want %q", tt.input, i, result[i], tt.expected[i])
				}
			}
		})
	}
}
