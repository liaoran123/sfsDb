package time

import (
	"testing"
	"time"
)

// TestSlidingWindow 测试滑动窗口功能
func TestSlidingWindow(t *testing.T) {
	startTime := time.Now().Add(-10 * time.Minute)
	endTime := time.Now()
	windowSize := 2 * time.Minute
	stepSize := 1 * time.Minute

	window := NewSlidingWindow(startTime, endTime, windowSize, stepSize)

	count := 0
	for {
		count++
		start := window.Start()
		end := window.End()
		if start.After(end) {
			t.Errorf("Window start time should be before end time")
		}
		if !window.Next() {
			break
		}
	}

	// 验证窗口数量
	expectedCount := 9 // 10分钟的时间范围，1分钟步长，2分钟窗口大小，应该有9个窗口
	if count != expectedCount {
		t.Errorf("Expected %d windows, got %d", expectedCount, count)
	}

	// 测试重置功能
	window.Reset()
	resetCount := 0
	for {
		resetCount++
		if !window.Next() {
			break
		}
	}
	if resetCount != expectedCount {
		t.Errorf("Expected %d windows after reset, got %d", expectedCount, resetCount)
	}
}

// TestTumblingWindow 测试滚动窗口功能
func TestTumblingWindow(t *testing.T) {
	startTime := time.Now().Add(-10 * time.Minute)
	endTime := time.Now()
	windowSize := 2 * time.Minute

	window := NewTumblingWindow(startTime, endTime, windowSize)

	count := 0
	for {
		count++
		start := window.Start()
		end := window.End()
		if start.After(end) {
			t.Errorf("Window start time should be before end time")
		}
		if !window.Next() {
			break
		}
	}

	// Next 契约为"先移后查":首窗 [0,2m] 不经过 Next 可见;
	// 恰在 endTime 结束的退化空窗 [10m,10m] 已按新语义剔除,故为 4 个
	expectedCount := 5
	if count != expectedCount {
		t.Errorf("Expected %d windows, got %d", expectedCount, count)
	}

	// 测试重置功能
	window.Reset()
	resetCount := 0
	for {
		resetCount++
		if !window.Next() {
			break
		}
	}
	if resetCount != expectedCount {
		t.Errorf("Expected %d windows after reset, got %d", expectedCount, resetCount)
	}
}

// TestMovingAveragePrediction 测试移动平均预测功能
func TestMovingAveragePrediction(t *testing.T) {
	// 创建测试数据点
	var points []TimeSeriesPoint
	now := time.Now()
	for i := 0; i < 10; i++ {
		points = append(points, TimeSeriesPoint{
			Time:  now.Add(time.Duration(i) * time.Minute),
			Value: float64(i),
		})
	}

	// 测试移动平均预测
	prediction := NewMovingAveragePrediction(points, 3, 5, time.Minute)

	if len(prediction.PredictedPoints) != 5 {
		t.Errorf("Expected 5 predicted points, got %d", len(prediction.PredictedPoints))
	}

	if prediction.WindowSize != 3 {
		t.Errorf("Expected window size 3, got %d", prediction.WindowSize)
	}
}

// TestLinearRegressionPrediction 测试线性回归预测功能
func TestLinearRegressionPrediction(t *testing.T) {
	// 创建测试数据点
	var points []TimeSeriesPoint
	now := time.Now()
	for i := 0; i < 10; i++ {
		points = append(points, TimeSeriesPoint{
			Time:  now.Add(time.Duration(i) * time.Minute),
			Value: float64(i),
		})
	}

	// 测试线性回归预测
	prediction := NewLinearRegressionPrediction(points, 5, time.Minute)

	if len(prediction.PredictedPoints) != 5 {
		t.Errorf("Expected 5 predicted points, got %d", len(prediction.PredictedPoints))
	}

	// 验证预测值是否合理（应该接近线性增长）
	if len(prediction.PredictedPoints) > 0 {
		firstPredictedValue := prediction.PredictedPoints[0].Value
		if firstPredictedValue < 9 || firstPredictedValue > 11 {
			t.Errorf("Expected first predicted value around 10, got %f", firstPredictedValue)
		}
	}
}

// TestCompression 测试时间序列数据压缩功能
func TestCompression(t *testing.T) {
	// 创建测试数据点
	var points []TimeSeriesPoint
	now := time.Now()
	for i := 0; i < 10; i++ {
		points = append(points, TimeSeriesPoint{
			Time:  now.Add(time.Duration(i) * time.Minute),
			Value: float64(i),
		})
	}

	// 测试Delta编码压缩
	deltaCompressed, err := CompressTimeSeries(points, "delta", time.Minute)
	if err != nil {
		t.Errorf("Delta compression failed: %v", err)
	}

	// 测试Delta编码解压缩
	deltaDecompressed, err := DecompressTimeSeries(deltaCompressed, 10)
	if err != nil {
		t.Errorf("Delta decompression failed: %v", err)
	}

	if len(deltaDecompressed) != 10 {
		t.Errorf("Expected 10 decompressed points, got %d", len(deltaDecompressed))
	}

	// 测试RLE编码压缩
	rleCompressed, err := CompressTimeSeries(points, "rle", time.Minute)
	if err != nil {
		t.Errorf("RLE compression failed: %v", err)
	}

	// 测试RLE编码解压缩
	rleDecompressed, err := DecompressTimeSeries(rleCompressed, 10)
	if err != nil {
		t.Errorf("RLE decompression failed: %v", err)
	}

	if len(rleDecompressed) != 10 {
		t.Errorf("Expected 10 decompressed points, got %d", len(rleDecompressed))
	}
}

// TestEnhancedGranularity 测试增强的时间粒度支持
func TestEnhancedGranularity(t *testing.T) {
	now := time.Now()

	// 测试毫秒粒度
	millisecondStr := FormatTimeByGranularity(now, TimeGranularityMillisecond)
	if len(millisecondStr) == 0 {
		t.Errorf("Millisecond granularity format should not be empty")
	}

	// 测试微秒粒度
	microsecondStr := FormatTimeByGranularity(now, TimeGranularityMicrosecond)
	if len(microsecondStr) == 0 {
		t.Errorf("Microsecond granularity format should not be empty")
	}

	// 测试周粒度
	weekStr := FormatTimeByGranularity(now, TimeGranularityWeek)
	if len(weekStr) == 0 {
		t.Errorf("Week granularity format should not be empty")
	}

	// 测试季度粒度
	quarterStr := FormatTimeByGranularity(now, TimeGranularityQuarter)
	if len(quarterStr) == 0 {
		t.Errorf("Quarter granularity format should not be empty")
	}

	// 测试工作日判断
	isWeekday := IsWeekday(now)
	isWeekend := IsWeekend(now)
	if isWeekday == isWeekend {
		t.Errorf("A day cannot be both a weekday and a weekend")
	}

	// 测试季度获取
	quarter := GetQuarter(now)
	if quarter < 1 || quarter > 4 {
		t.Errorf("Quarter should be between 1 and 4, got %d", quarter)
	}

	// 测试周数获取
	week := GetWeekNumber(now)
	if week < 1 || week > 53 {
		t.Errorf("Week number should be between 1 and 53, got %d", week)
	}
}

// TestTimeRangeAdjustment 测试时间范围调整功能
func TestTimeRangeAdjustment(t *testing.T) {
	now := time.Now()

	// 测试按小时调整
	hourAdjusted := AdjustTimeToGranularity(now, TimeGranularityHour)
	if hourAdjusted.Minute() != 0 || hourAdjusted.Second() != 0 {
		t.Errorf("Hour adjusted time should have minute and second set to 0")
	}

	// 测试按天调整
	dayAdjusted := AdjustTimeToGranularity(now, TimeGranularityDay)
	if dayAdjusted.Hour() != 0 || dayAdjusted.Minute() != 0 {
		t.Errorf("Day adjusted time should have hour and minute set to 0")
	}

	// 测试按周调整
	weekAdjusted := AdjustTimeToGranularity(now, TimeGranularityWeek)
	weekday := weekAdjusted.Weekday()
	if weekday != time.Monday {
		t.Errorf("Week adjusted time should be Monday, got %v", weekday)
	}

	// 测试按季度调整
	quarterAdjusted := AdjustTimeToGranularity(now, TimeGranularityQuarter)
	month := int(quarterAdjusted.Month())
	if (month-1)%3 != 0 {
		t.Errorf("Quarter adjusted time should be the first month of the quarter, got %d", month)
	}
}

// TestWindowAggregation 测试窗口聚合功能
func TestWindowAggregation(t *testing.T) {
	// 创建测试数据
	var records []map[string]any
	now := time.Now()
	for i := 0; i < 10; i++ {
		records = append(records, map[string]any{
			"timestamp": now.Add(time.Duration(i) * time.Minute),
			"value":     float64(i),
		})
	}

	// 创建滑动窗口
	startTime := now.Add(-1 * time.Minute)
	endTime := now.Add(10 * time.Minute)
	windowSize := 3 * time.Minute
	stepSize := 2 * time.Minute
	window := NewSlidingWindow(startTime, endTime, windowSize, stepSize)

	// 执行窗口聚合
	results, err := AggregateByWindow(records, "timestamp", "value", window, "sum")
	if err != nil {
		t.Errorf("Window aggregation failed: %v", err)
	}

	if len(results) == 0 {
		t.Errorf("Window aggregation should return at least one result")
	}
}

// TestAggregateByWindowTermination 回归测试:空窗口与 endTime 恰好落在窗口边界时不得死循环
func TestAggregateByWindowTermination(t *testing.T) {
	startTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := startTime.Add(10 * time.Hour)

	// 仅首窗有数据,中间与边界窗口为空;endTime 恰好等于窗口边界
	window := NewTumblingWindow(startTime, endTime, 5*time.Hour)
	records := []map[string]any{
		{"timestamp": startTime.Add(time.Minute), "value": float64(1)},
	}
	results, err := AggregateByWindow(records, "timestamp", "value", window, "sum")
	if err != nil {
		t.Fatalf("AggregateByWindow returned error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 aggregated window, got %d", len(results))
	}
	if !results[0].WindowStart.Equal(startTime) {
		t.Errorf("first window should start at %v, got %v", startTime, results[0].WindowStart)
	}
	if results[0].Value != 1 {
		t.Errorf("expected sum 1, got %v", results[0].Value)
	}

	// 无任何数据,所有窗口为空,循环必须正常终止
	emptyWindow := NewTumblingWindow(startTime, endTime, 5*time.Hour)
	results, err = AggregateByWindow(nil, "timestamp", "value", emptyWindow, "count")
	if err != nil {
		t.Fatalf("AggregateByWindow returned error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results for empty records, got %d", len(results))
	}

	// max 全负值:math.Inf(-1) 初值不得泄漏
	maxWindow := NewTumblingWindow(startTime, endTime, 5*time.Hour)
	negRecords := []map[string]any{
		{"timestamp": startTime.Add(time.Minute), "value": float64(-5)},
	}
	negResults, err := AggregateByWindow(negRecords, "timestamp", "value", maxWindow, "max")
	if err != nil {
		t.Fatalf("AggregateByWindow returned error: %v", err)
	}
	if len(negResults) != 1 || negResults[0].Value != -5 {
		t.Errorf("max of all-negative values should be -5, got %v", negResults)
	}
}

// TestTumblingWindowNextBoundary 回归测试:当前窗口结束时间恰等于 endTime 时 Next() 应直接停止
func TestTumblingWindowNextBoundary(t *testing.T) {
	startTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := startTime.Add(10 * time.Hour)
	window := NewTumblingWindow(startTime, endTime, 5*time.Hour)

	// 首窗 [0,5h],Next 后移到 [5h,10h]
	if !window.Next() {
		t.Fatal("expected Next() to advance to second window")
	}
	if !window.Start().Equal(startTime.Add(5 * time.Hour)) {
		t.Fatalf("expected window start at 5h, got %v", window.Start())
	}

	// 当前窗 [5h,10h] 恰在 endTime 结束,下一窗必为空,Next 应停止
	if window.Next() {
		t.Fatal("expected Next() to stop when current window ends exactly at endTime")
	}

	// endTime 非整窗倍数:不完整尾窗仍应被生成
	partial := NewTumblingWindow(startTime, startTime.Add(12*time.Hour), 5*time.Hour)
	// Next 契约为"先移后查",首窗起点由调用方给
	starts := []time.Duration{0}
	for partial.Next() {
		starts = append(starts, partial.Start().Sub(startTime))
		if len(starts) > 10 {
			t.Fatal("TumblingWindow did not terminate")
		}
	}
	expected := []time.Duration{0, 5 * time.Hour, 10 * time.Hour}
	if len(starts) != len(expected) {
		t.Fatalf("expected window starts %v, got %v", expected, starts)
	}
}
