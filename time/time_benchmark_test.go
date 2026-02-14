package time

import (
	"testing"
	"time"
)

// BenchmarkTimeToUnixTimestamp 测试时间转时间戳性能
func BenchmarkTimeToUnixTimestamp(b *testing.B) {
	now := time.Now()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TimeToUnixTimestamp(now)
	}
}

// BenchmarkUnixTimestampToTime 测试时间戳转时间性能
func BenchmarkUnixTimestampToTime(b *testing.B) {
	timestamp := time.Now().Unix()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		UnixTimestampToTime(timestamp)
	}
}

// BenchmarkFormatTimeByGranularity 测试时间粒度格式化性能
func BenchmarkFormatTimeByGranularity(b *testing.B) {
	now := time.Now()
	granularity := TimeGranularitySecond
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FormatTimeByGranularity(now, granularity)
	}
}

// BenchmarkTimeBucket 测试时间桶分配性能
func BenchmarkTimeBucket(b *testing.B) {
	now := time.Now()
	duration := time.Hour
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TimeBucket(now, duration)
	}
}

// BenchmarkTimeRange 测试时间范围计算性能
func BenchmarkTimeRange(b *testing.B) {
	now := time.Now()
	granularity := TimeGranularityHour
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TimeRange(now, granularity)
	}
}

// BenchmarkIsWeekday 测试工作日检查性能
func BenchmarkIsWeekday(b *testing.B) {
	now := time.Now()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsWeekday(now)
	}
}

// BenchmarkGetQuarter 测试季度获取性能
func BenchmarkGetQuarter(b *testing.B) {
	now := time.Now()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetQuarter(now)
	}
}

// BenchmarkGetWeekNumber 测试周数获取性能
func BenchmarkGetWeekNumber(b *testing.B) {
	now := time.Now()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetWeekNumber(now)
	}
}

// BenchmarkNewMovingAveragePrediction 测试移动平均预测创建性能
func BenchmarkNewMovingAveragePrediction(b *testing.B) {
	// 创建测试数据点
	var points []TimeSeriesPoint
	now := time.Now()
	for i := 0; i < 100; i++ {
		points = append(points, TimeSeriesPoint{
			Time:  now.Add(time.Duration(i) * time.Minute),
			Value: float64(i),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewMovingAveragePrediction(points, 10, 5, time.Minute)
	}
}

// BenchmarkNewLinearRegressionPrediction 测试线性回归预测创建性能
func BenchmarkNewLinearRegressionPrediction(b *testing.B) {
	// 创建测试数据点
	var points []TimeSeriesPoint
	now := time.Now()
	for i := 0; i < 100; i++ {
		points = append(points, TimeSeriesPoint{
			Time:  now.Add(time.Duration(i) * time.Minute),
			Value: float64(i),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewLinearRegressionPrediction(points, 5, time.Minute)
	}
}

// BenchmarkCompressTimeSeries 测试时间序列压缩性能
func BenchmarkCompressTimeSeries(b *testing.B) {
	// 创建测试数据点
	var points []TimeSeriesPoint
	now := time.Now()
	for i := 0; i < 100; i++ {
		points = append(points, TimeSeriesPoint{
			Time:  now.Add(time.Duration(i) * time.Minute),
			Value: float64(i),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CompressTimeSeries(points, "delta", time.Minute)
	}
}

// BenchmarkDecompressTimeSeries 测试时间序列解压性能
func BenchmarkDecompressTimeSeries(b *testing.B) {
	// 创建测试数据点并压缩
	var points []TimeSeriesPoint
	now := time.Now()
	for i := 0; i < 100; i++ {
		points = append(points, TimeSeriesPoint{
			Time:  now.Add(time.Duration(i) * time.Minute),
			Value: float64(i),
		})
	}

	compressed, _ := CompressTimeSeries(points, "delta", time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DecompressTimeSeries(compressed, 100)
	}
}

// BenchmarkAggregateByTimeGranularity 测试时间粒度聚合性能
func BenchmarkAggregateByTimeGranularity(b *testing.B) {
	// 创建测试数据
	var records []map[string]any
	now := time.Now()
	for i := 0; i < 1000; i++ {
		records = append(records, map[string]any{
			"timestamp": now.Add(time.Duration(i) * time.Minute),
			"value":     float64(i),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AggregateByTimeGranularity(records, "timestamp", "value", TimeGranularityHour, "sum")
	}
}
