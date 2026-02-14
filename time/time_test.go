package time

import (
	"fmt"
	"testing"
	"time"

	"github.com/liaoran123/sfsDb/record"
)

func TestFormatTimeByGranularity(t *testing.T) {
	// 创建一个测试时间
	testTime := time.Date(2024, 12, 25, 14, 30, 45, 123456789, time.UTC)

	// 测试不同时间粒度的格式化
	tests := []struct {
		granularity TimeGranularity
		expected    string
	}{
		{TimeGranularitySecond, "2024-12-25 14:30:45"},
		{TimeGranularityMinute, "2024-12-25 14:30:00"},
		{TimeGranularityHour, "2024-12-25 14:00:00"},
		{TimeGranularityDay, "2024-12-25"},
		{TimeGranularityMonth, "2024-12"},
		{TimeGranularityYear, "2024"},
	}

	for _, tt := range tests {
		result := FormatTimeByGranularity(testTime, tt.granularity)
		if result != tt.expected {
			t.Errorf("FormatTimeByGranularity(%v) = %v; want %v", tt.granularity, result, tt.expected)
		}
	}
}

func TestTimeBucket(t *testing.T) {
	// 创建一个测试时间
	testTime := time.Date(2024, 12, 25, 14, 30, 45, 123456789, time.UTC)

	// 测试不同时间桶
	tests := []struct {
		duration time.Duration
		expectedHour int
		expectedMinute int
		expectedSecond int
	}{
		{time.Second, 14, 30, 45},
		{time.Minute, 14, 30, 0},
		{time.Hour, 14, 0, 0},
	}

	for _, tt := range tests {
		result := TimeBucket(testTime, tt.duration)
		if result.Hour() != tt.expectedHour || result.Minute() != tt.expectedMinute || result.Second() != tt.expectedSecond {
			t.Errorf("TimeBucket(%v) = %v; want hour=%d, minute=%d, second=%d", 
				tt.duration, result, tt.expectedHour, tt.expectedMinute, tt.expectedSecond)
		}
	}
}

func TestTimeRange(t *testing.T) {
	// 创建一个测试时间
	testTime := time.Date(2024, 12, 25, 14, 30, 45, 0, time.UTC)

	// 测试不同时间粒度的时间范围
	tests := []struct {
		granularity TimeGranularity
		startExpected string
		endExpected   string
	}{
		{
			TimeGranularityHour,
			"2024-12-25 14:30:45 +0000 UTC",
			"2024-12-25 15:30:45 +0000 UTC",
		},
		{
			TimeGranularityDay,
			"2024-12-25 14:30:45 +0000 UTC",
			"2024-12-26 14:30:45 +0000 UTC",
		},
		{
			TimeGranularityMonth,
			"2024-12-25 14:30:45 +0000 UTC",
			"2025-01-24 14:30:45 +0000 UTC", // 30天后
		},
	}

	for _, tt := range tests {
		start, end := TimeRange(testTime, tt.granularity)
		if start.String() != tt.startExpected {
			t.Errorf("TimeRange(%v) start = %v; want %v", tt.granularity, start, tt.startExpected)
		}
		if end.String() != tt.endExpected {
			t.Errorf("TimeRange(%v) end = %v; want %v", tt.granularity, end, tt.endExpected)
		}
	}
}

func TestAggregateByTimeGranularity(t *testing.T) {
	// 创建测试数据 - 使用固定时间确保可重复性
	baseTime := time.Date(2024, 12, 25, 14, 0, 0, 0, time.UTC)
	records := make(record.Records, 0, 4)

	// 创建4条测试记录，分布在2个小时内
	timestamps := []time.Time{
		baseTime.Add(15 * time.Minute),  // 14:15
		baseTime.Add(45 * time.Minute),  // 14:45
		baseTime.Add(75 * time.Minute),  // 15:15
		baseTime.Add(105 * time.Minute), // 15:45
	}

	for i, timestamp := range timestamps {
		record := record.GetRecord()
		record["timestamp"] = timestamp
		record["value"] = float64(i + 1)
		records = append(records, record)
	}

	// 测试按小时聚合
	results, err := AggregateByTimeGranularity(records, "timestamp", "value", TimeGranularityHour, "sum")
	if err != nil {
		t.Fatalf("AggregateByTimeGranularity failed: %v", err)
	}

	// 验证结果
	if len(results) != 2 {
		t.Errorf("Expected 2 hourly results, got %d", len(results))
	}

	// 打印结果
	fmt.Println("Hourly aggregation results:")
	for _, result := range results {
		fmt.Printf("Time: %s, Sum: %.2f\n", result.TimeKey, result.Value)
	}

	// 释放记录
	for _, r := range records {
		r.Release()
	}
}
