package time

import (
	"sync"
	"testing"
	"time"
)

const concurrentWorkers = 100

// BenchmarkConcurrentTimeToUnixTimestamp 测试并发时间转时间戳性能
func BenchmarkConcurrentTimeToUnixTimestamp(b *testing.B) {
	now := time.Now()
	var wg sync.WaitGroup

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			TimeToUnixTimestamp(now)
		}()
	}
	wg.Wait()
}

// BenchmarkConcurrentUnixTimestampToTime 测试并发时间戳转时间性能
func BenchmarkConcurrentUnixTimestampToTime(b *testing.B) {
	timestamp := time.Now().Unix()
	var wg sync.WaitGroup

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			UnixTimestampToTime(timestamp)
		}()
	}
	wg.Wait()
}

// BenchmarkConcurrentFormatTimeByGranularity 测试并发时间粒度格式化性能
func BenchmarkConcurrentFormatTimeByGranularity(b *testing.B) {
	now := time.Now()
	granularity := TimeGranularitySecond
	var wg sync.WaitGroup

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			FormatTimeByGranularity(now, granularity)
		}()
	}
	wg.Wait()
}

// BenchmarkConcurrentTimeBucket 测试并发时间桶分配性能
func BenchmarkConcurrentTimeBucket(b *testing.B) {
	now := time.Now()
	duration := time.Hour
	var wg sync.WaitGroup

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			TimeBucket(now, duration)
		}()
	}
	wg.Wait()
}

// BenchmarkConcurrentAggregateByTimeGranularity 测试并发时间粒度聚合性能
func BenchmarkConcurrentAggregateByTimeGranularity(b *testing.B) {
	// 创建测试数据
	createTestData := func() []map[string]any {
		var records []map[string]any
		now := time.Now()
		for i := 0; i < 100; i++ {
			records = append(records, map[string]any{
				"timestamp": now.Add(time.Duration(i) * time.Minute),
				"value":     float64(i),
			})
		}
		return records
	}

	var wg sync.WaitGroup

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// 为每个goroutine创建自己的数据副本
			records := createTestData()
			AggregateByTimeGranularity(records, "timestamp", "value", TimeGranularityHour, "sum")
		}()
	}
	wg.Wait()
}

// BenchmarkConcurrentCompressTimeSeries 测试并发时间序列压缩性能
func BenchmarkConcurrentCompressTimeSeries(b *testing.B) {
	// 创建测试数据点
	createTestPoints := func() []TimeSeriesPoint {
		var points []TimeSeriesPoint
		now := time.Now()
		for i := 0; i < 50; i++ {
			points = append(points, TimeSeriesPoint{
				Time:  now.Add(time.Duration(i) * time.Minute),
				Value: float64(i),
			})
		}
		return points
	}

	var wg sync.WaitGroup

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// 为每个goroutine创建自己的数据副本
			points := createTestPoints()
			CompressTimeSeries(points, "delta", time.Minute)
		}()
	}
	wg.Wait()
}

// BenchmarkConcurrentNewMovingAveragePrediction 测试并发移动平均预测性能
func BenchmarkConcurrentNewMovingAveragePrediction(b *testing.B) {
	// 创建测试数据点
	createTestPoints := func() []TimeSeriesPoint {
		var points []TimeSeriesPoint
		now := time.Now()
		for i := 0; i < 50; i++ {
			points = append(points, TimeSeriesPoint{
				Time:  now.Add(time.Duration(i) * time.Minute),
				Value: float64(i),
			})
		}
		return points
	}

	var wg sync.WaitGroup

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// 为每个goroutine创建自己的数据副本
			points := createTestPoints()
			NewMovingAveragePrediction(points, 5, 3, time.Minute)
		}()
	}
	wg.Wait()
}
