package time

import (
	"testing"
	"time"
)

func TestTimeToUnixTimestamp(t *testing.T) {
	// 创建测试时间
	testTime := time.Date(2024, 12, 25, 14, 30, 45, 123456789, time.UTC)

	// 测试秒级时间戳转换 - 与标准库方法对比
	expected := testTime.Unix()
	result := TimeToUnixTimestamp(testTime)
	if result != expected {
		t.Errorf("TimeToUnixTimestamp() = %v; want %v", result, expected)
	}
}

func TestTimeToUnixTimestampMs(t *testing.T) {
	// 创建测试时间
	testTime := time.Date(2024, 12, 25, 14, 30, 45, 123456789, time.UTC)

	// 测试毫秒级时间戳转换 - 与标准库方法对比
	expected := testTime.UnixMilli()
	result := TimeToUnixTimestampMs(testTime)
	if result != expected {
		t.Errorf("TimeToUnixTimestampMs() = %v; want %v", result, expected)
	}
}

func TestTimeToUnixTimestampNs(t *testing.T) {
	// 创建测试时间
	testTime := time.Date(2024, 12, 25, 14, 30, 45, 123456789, time.UTC)

	// 测试纳秒级时间戳转换 - 与标准库方法对比
	expected := testTime.UnixNano()
	result := TimeToUnixTimestampNs(testTime)
	if result != expected {
		t.Errorf("TimeToUnixTimestampNs() = %v; want %v", result, expected)
	}
}

func TestUnixTimestampToTime(t *testing.T) {
	// 创建测试时间
	testTime := time.Date(2024, 12, 25, 14, 30, 45, 0, time.UTC)
	timestamp := testTime.Unix()

	// 测试秒级时间戳转换为时间 - 与标准库方法对比
	expected := time.Unix(timestamp, 0)
	result := UnixTimestampToTime(timestamp)
	
	// 比较时间是否相等
	if !result.Equal(expected) {
		t.Errorf("UnixTimestampToTime() = %v; want %v", result, expected)
	}
}

func TestUnixTimestampMsToTime(t *testing.T) {
	// 创建测试时间
	testTime := time.Date(2024, 12, 25, 14, 30, 45, 123000000, time.UTC)
	timestampMs := testTime.UnixMilli()

	// 测试毫秒级时间戳转换为时间 - 与标准库方法对比
	expected := time.Unix(0, timestampMs*int64(time.Millisecond))
	result := UnixTimestampMsToTime(timestampMs)
	
	// 比较时间是否相等
	if !result.Equal(expected) {
		t.Errorf("UnixTimestampMsToTime() = %v; want %v", result, expected)
	}
}

func TestUnixTimestampNsToTime(t *testing.T) {
	// 创建测试时间
	testTime := time.Date(2024, 12, 25, 14, 30, 45, 123456789, time.UTC)
	timestampNs := testTime.UnixNano()

	// 测试纳秒级时间戳转换为时间 - 与标准库方法对比
	expected := time.Unix(0, timestampNs)
	result := UnixTimestampNsToTime(timestampNs)
	
	// 比较时间是否相等
	if !result.Equal(expected) {
		t.Errorf("UnixTimestampNsToTime() = %v; want %v", result, expected)
	}
}

func TestTimestampConversionRoundTrip(t *testing.T) {
	// 测试时间戳转换的往返一致性
	originalTime := time.Date(2024, 12, 25, 14, 30, 45, 123456789, time.UTC)

	// 秒级时间戳往返
	unixTime := TimeToUnixTimestamp(originalTime)
	convertedTime := UnixTimestampToTime(unixTime)
	if !convertedTime.Equal(time.Date(2024, 12, 25, 14, 30, 45, 0, time.UTC)) {
		t.Errorf("Second timestamp round trip failed: got %v", convertedTime)
	}

	// 毫秒级时间戳往返
	unixTimeMs := TimeToUnixTimestampMs(originalTime)
	convertedTimeMs := UnixTimestampMsToTime(unixTimeMs)
	if !convertedTimeMs.Equal(time.Date(2024, 12, 25, 14, 30, 45, 123000000, time.UTC)) {
		t.Errorf("Millisecond timestamp round trip failed: got %v", convertedTimeMs)
	}

	// 纳秒级时间戳往返
	unixTimeNs := TimeToUnixTimestampNs(originalTime)
	convertedTimeNs := UnixTimestampNsToTime(unixTimeNs)
	if !convertedTimeNs.Equal(originalTime) {
		t.Errorf("Nanosecond timestamp round trip failed: got %v", convertedTimeNs)
	}
}
