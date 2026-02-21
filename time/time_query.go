package time

import (
	"time"

	"github.com/liaoran123/sfsDb/engine"
)

// TimeRangeQueryOptions 时间范围查询选项
type TimeRangeQueryOptions struct {
	FieldName       string
	StartTime       time.Time
	EndTime         time.Time
	TimeGranularity TimeGranularity
	Inclusive       bool // 是否包含边界值
}

// NewTimeRangeQueryOptions 创建时间范围查询选项
func NewTimeRangeQueryOptions(fieldName string, startTime, endTime time.Time, granularity TimeGranularity) *TimeRangeQueryOptions {
	return &TimeRangeQueryOptions{
		FieldName:       fieldName,
		StartTime:       startTime,
		EndTime:         endTime,
		TimeGranularity: granularity,
		Inclusive:       true,
	}
}

// SearchTimeRange 执行时间范围查询
// table: sfsDb表实例
// options: 时间范围查询选项
// 返回值: 表迭代器和错误
func SearchTimeRange(table *engine.Table, options *TimeRangeQueryOptions) (*engine.TableIter, error) {
	// 执行范围查询，使用nil作为迭代器函数，让table.SearchRange使用默认迭代器
	iter, err := table.SearchRange(nil, &map[string]any{options.FieldName: options.StartTime}, &map[string]any{options.FieldName: options.EndTime})
	if err != nil {
		return nil, err
	}

	return iter, nil
}

// SearchTimeRangeWithGranularity 按时间粒度执行时间范围查询
// table: sfsDb表实例
// fieldName: 时间字段名
// startTime: 开始时间
// endTime: 结束时间
// granularity: 时间粒度
// 返回值: 表迭代器和错误
func SearchTimeRangeWithGranularity(table *engine.Table, fieldName string, startTime, endTime time.Time, granularity TimeGranularity) (*engine.TableIter, error) {
	// 根据时间粒度调整时间范围
	adjustedStart, adjustedEnd := AdjustTimeRangeByGranularity(startTime, endTime, granularity)

	// 创建查询选项
	options := NewTimeRangeQueryOptions(fieldName, adjustedStart, adjustedEnd, granularity)

	// 执行查询
	return SearchTimeRange(table, options)
}

// AdjustTimeRangeByGranularity 根据时间粒度调整时间范围
func AdjustTimeRangeByGranularity(startTime, endTime time.Time, granularity TimeGranularity) (time.Time, time.Time) {
	// 调整开始时间到粒度边界
	adjustedStart := AdjustTimeToGranularity(startTime, granularity)

	// 调整结束时间到粒度边界的下一个点
	adjustedEnd := AdjustTimeToGranularity(endTime, granularity)
	switch granularity {
	case TimeGranularityMillisecond:
		adjustedEnd = adjustedEnd.Add(time.Millisecond)
	case TimeGranularityMicrosecond:
		adjustedEnd = adjustedEnd.Add(time.Microsecond)
	case TimeGranularitySecond:
		adjustedEnd = adjustedEnd.Add(time.Second)
	case TimeGranularityMinute:
		adjustedEnd = adjustedEnd.Add(time.Minute)
	case TimeGranularityHour:
		adjustedEnd = adjustedEnd.Add(time.Hour)
	case TimeGranularityDay:
		adjustedEnd = adjustedEnd.Add(24 * time.Hour)
	case TimeGranularityWeek:
		adjustedEnd = adjustedEnd.Add(7 * 24 * time.Hour)
	case TimeGranularityMonth:
		// 简化处理，实际需要更复杂的逻辑
		adjustedEnd = adjustedEnd.AddDate(0, 1, 0)
	case TimeGranularityQuarter:
		// 简化处理，实际需要更复杂的逻辑
		adjustedEnd = adjustedEnd.AddDate(0, 3, 0)
	case TimeGranularityYear:
		adjustedEnd = adjustedEnd.AddDate(1, 0, 0)
	}

	return adjustedStart, adjustedEnd
}

// AdjustTimeToGranularity 将时间调整到指定粒度的边界
func AdjustTimeToGranularity(t time.Time, granularity TimeGranularity) time.Time {
	switch granularity {
	case TimeGranularityMillisecond:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond()/1000000*1000000, t.Location())
	case TimeGranularityMicrosecond:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond()/1000*1000, t.Location())
	case TimeGranularitySecond:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, t.Location())
	case TimeGranularityMinute:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, t.Location())
	case TimeGranularityHour:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
	case TimeGranularityDay:
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	case TimeGranularityWeek:
		// 调整到本周的第一天（周一）
		weekday := int(t.Weekday())
		if weekday == 0 { // 周日
			weekday = 7
		}
		return time.Date(t.Year(), t.Month(), t.Day()-weekday+1, 0, 0, 0, 0, t.Location())
	case TimeGranularityMonth:
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	case TimeGranularityQuarter:
		// 调整到本季度的第一天
		quarter := (int(t.Month())-1)/3 + 1
		return time.Date(t.Year(), time.Month((quarter-1)*3+1), 1, 0, 0, 0, 0, t.Location())
	case TimeGranularityYear:
		return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
	default:
		return t
	}
}

// TimeRangeQueryWithAggregation 带聚合的时间范围查询
// table: sfsDb表实例
// options: 时间范围查询选项
// valueField: 聚合值字段名
// aggregationType: 聚合类型 (sum, avg, count, max, min)
// 返回值: 聚合结果和错误
func TimeRangeQueryWithAggregation(table *engine.Table, options *TimeRangeQueryOptions, valueField string, aggregationType string) ([]TimeAggregationResult, error) {
	// 执行时间范围查询
	iter, err := SearchTimeRange(table, options)
	if err != nil {
		return nil, err
	}
	defer iter.Release()

	// 获取记录集
	records := iter.GetRecords(true)
	defer records.Release()

	// 转换为map切片
	var recordMaps []map[string]any
	for _, record := range records {
		if record != nil {
			recordMaps = append(recordMaps, map[string]any(record))
		}
	}

	// 按时间粒度聚合
	results, err := AggregateByTimeGranularity(recordMaps, options.FieldName, valueField, options.TimeGranularity, aggregationType)
	if err != nil {
		return nil, err
	}

	return results, nil
}
