package time

import (
	"time"
)

// TimeWindow 时间窗口接口
type TimeWindow interface {
	// Next 移动到下一个窗口，返回是否还有下一个窗口
	Next() bool
	// Start 获取当前窗口的开始时间
	Start() time.Time
	// End 获取当前窗口的结束时间
	End() time.Time
	// Reset 重置窗口到初始状态
	Reset()
}

// SlidingWindow 滑动窗口实现
type SlidingWindow struct {
	startTime    time.Time
	endTime      time.Time
	windowSize   time.Duration
	stepSize     time.Duration
	currentStart time.Time
}

// NewSlidingWindow 创建一个新的滑动窗口
// startTime: 窗口起始时间
// windowSize: 窗口大小
// stepSize: 滑动步长
// endTime: 窗口结束时间
func NewSlidingWindow(startTime, endTime time.Time, windowSize, stepSize time.Duration) *SlidingWindow {
	return &SlidingWindow{
		startTime:    startTime,
		endTime:      endTime,
		windowSize:   windowSize,
		stepSize:     stepSize,
		currentStart: startTime,
	}
}

// Next 移动到下一个窗口，返回是否还有下一个窗口
func (w *SlidingWindow) Next() bool {
	if w.currentStart.Add(w.windowSize).After(w.endTime) {
		return false
	}
	w.currentStart = w.currentStart.Add(w.stepSize)
	return true
}

// Start 获取当前窗口的开始时间
func (w *SlidingWindow) Start() time.Time {
	return w.currentStart
}

// End 获取当前窗口的结束时间
func (w *SlidingWindow) End() time.Time {
	end := w.currentStart.Add(w.windowSize)
	if end.After(w.endTime) {
		return w.endTime
	}
	return end
}

// Reset 重置窗口到初始状态
func (w *SlidingWindow) Reset() {
	w.currentStart = w.startTime
}

// TumblingWindow 滚动窗口实现
type TumblingWindow struct {
	startTime    time.Time
	endTime      time.Time
	windowSize   time.Duration
	currentStart time.Time
}

// NewTumblingWindow 创建一个新的滚动窗口
// startTime: 窗口起始时间
// windowSize: 窗口大小
// endTime: 窗口结束时间
func NewTumblingWindow(startTime, endTime time.Time, windowSize time.Duration) *TumblingWindow {
	return &TumblingWindow{
		startTime:    startTime,
		endTime:      endTime,
		windowSize:   windowSize,
		currentStart: startTime,
	}
}

// Next 移动到下一个窗口，返回是否还有下一个窗口
func (w *TumblingWindow) Next() bool {
	if w.currentStart.Add(w.windowSize).After(w.endTime) {
		return false
	}
	w.currentStart = w.currentStart.Add(w.windowSize)
	return true
}

// Start 获取当前窗口的开始时间
func (w *TumblingWindow) Start() time.Time {
	return w.currentStart
}

// End 获取当前窗口的结束时间
func (w *TumblingWindow) End() time.Time {
	end := w.currentStart.Add(w.windowSize)
	if end.After(w.endTime) {
		return w.endTime
	}
	return end
}

// Reset 重置窗口到初始状态
func (w *TumblingWindow) Reset() {
	w.currentStart = w.startTime
}

// WindowAggregation 窗口聚合结果
type WindowAggregation struct {
	WindowStart time.Time  `json:"window_start"`
	WindowEnd   time.Time  `json:"window_end"`
	Value       float64    `json:"value"`
}

// AggregateByWindow 按时间窗口聚合数据
func AggregateByWindow(records []map[string]any, timeField string, valueField string, window TimeWindow, aggregationType string) ([]WindowAggregation, error) {
	var results []WindowAggregation

	// 重置窗口
	window.Reset()

	// 对每个窗口进行聚合
	for window.Next() {
		windowStart := window.Start()
		windowEnd := window.End()

		// 筛选窗口内的记录
		var windowRecords []map[string]any
		for _, record := range records {
			if record == nil {
				continue
			}

			timestamp, ok := record[timeField].(time.Time)
			if !ok {
				continue
			}

			// 检查时间戳是否在窗口内
			if (timestamp.Equal(windowStart) || timestamp.After(windowStart)) && timestamp.Before(windowEnd) {
				windowRecords = append(windowRecords, record)
			}
		}

		// 如果窗口内没有记录，跳过
		if len(windowRecords) == 0 {
			continue
		}

		// 计算聚合值
		var aggregatedValue float64
		switch aggregationType {
		case "sum":
			for _, record := range windowRecords {
				if val, ok := record[valueField].(float64); ok {
					aggregatedValue += val
				}
			}
		case "avg":
			var sum float64
			var count int
			for _, record := range windowRecords {
				if val, ok := record[valueField].(float64); ok {
					sum += val
					count++
				}
			}
			if count > 0 {
				aggregatedValue = sum / float64(count)
			}
		case "max":
			aggregatedValue = -1
			for _, record := range windowRecords {
				if val, ok := record[valueField].(float64); ok && val > aggregatedValue {
					aggregatedValue = val
				}
			}
		case "min":
			aggregatedValue = 1<<63 - 1
			for _, record := range windowRecords {
				if val, ok := record[valueField].(float64); ok && val < aggregatedValue {
					aggregatedValue = val
				}
			}
		case "count":
			aggregatedValue = float64(len(windowRecords))
		default:
			aggregatedValue = 0
		}

		// 添加聚合结果
		results = append(results, WindowAggregation{
			WindowStart: windowStart,
			WindowEnd:   windowEnd,
			Value:       aggregatedValue,
		})
	}

	return results, nil
}
