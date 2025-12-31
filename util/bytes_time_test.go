package util

import (
	"testing"
	"time"
)

func TestBytesTime(t *testing.T) {
	// 测试用例：[]byte字符串和预期的time.Time值
	tests := []struct {
		name     string
		input    []byte
		expected time.Time
	}{
		{
			name:     "默认格式",
			input:    []byte("2023-10-25 14:30:45"),
			expected: time.Date(2023, 10, 25, 14, 30, 45, 0, time.UTC),
		},
		{
			name:     "ISO格式",
			input:    []byte("2023-10-25T14:30:45"),
			expected: time.Date(2023, 10, 25, 14, 30, 45, 0, time.UTC),
		},
		{
			name:     "仅日期",
			input:    []byte("2023-10-25"),
			expected: time.Date(2023, 10, 25, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "仅时间",
			input:    []byte("14:30:45"),
			expected: time.Date(0, 1, 1, 14, 30, 45, 0, time.UTC),
		},
		{
			name:     "带毫秒的默认格式",
			input:    []byte("2023-10-25 14:30:45.123"),
			expected: time.Date(2023, 10, 25, 14, 30, 45, 123000000, time.UTC),
		},
		{
			name:     "带毫秒的ISO格式",
			input:    []byte("2023-10-25T14:30:45.123"),
			expected: time.Date(2023, 10, 25, 14, 30, 45, 123000000, time.UTC),
		},
		{
			name:     "Unix时间戳(秒)",
			input:    []byte("1698244245"),
			expected: time.Unix(1698244245, 0),
		},
		{
			name:     "Unix时间戳(毫秒)",
			input:    []byte("1698244245123"),
			expected: time.UnixMilli(1698244245123),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Bytes(tt.input).Time()
			// 比较年、月、日、时、分、秒、毫秒
			if result.Year() != tt.expected.Year() ||
				result.Month() != tt.expected.Month() ||
				result.Day() != tt.expected.Day() ||
				result.Hour() != tt.expected.Hour() ||
				result.Minute() != tt.expected.Minute() ||
				result.Second() != tt.expected.Second() ||
				result.Nanosecond()/1000000 != tt.expected.Nanosecond()/1000000 {
				t.Errorf("Bytes(%q).Time() = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}

	// 测试解析失败的情况
	t.Run("解析失败", func(t *testing.T) {
		result := Bytes([]byte("invalid time format")).Time()
		// 解析失败应该返回零时间
		if !result.IsZero() {
			t.Errorf("Bytes(%q).Time() should return zero time, got %v", "invalid time format", result)
		}
	})
}
