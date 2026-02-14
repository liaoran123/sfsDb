package time

import (
	"fmt"
	"reflect"
	"time"

	"github.com/liaoran123/sfsDb/record"
)

// TimeAggregationResult 时间聚合结果
type TimeAggregationResult struct {
	TimeKey string  `json:"time_key"`
	Value   float64 `json:"value"`
}

// AggregateByTimeGranularity 按时间粒度聚合数据
func AggregateByTimeGranularity(records interface{}, timeField string, valueField string,
	granularity TimeGranularity, aggregationType string) ([]TimeAggregationResult, error) {

	// 处理不同类型的输入
	var recordList []map[string]any

	switch r := records.(type) {
	case []map[string]any:
		recordList = r
	case record.Records:
		// 处理record.Records类型
		for _, rec := range r {
			if rec == nil {
				continue
			}
			// 将record.Record转换为map[string]any
			recordMap := make(map[string]any)
			for k, v := range rec {
				recordMap[k] = v
			}
			recordList = append(recordList, recordMap)
		}
	default:
		// 尝试处理其他切片类型
		// 检查是否是切片类型
		if slice, ok := r.(interface{ Len() int }); ok {
			// 尝试遍历切片
			for i := 0; i < slice.Len(); i++ {
				// 尝试获取第i个元素
				if elem, ok := getSliceElement(r, i); ok {
					if recordMap, ok := elem.(map[string]any); ok {
						recordList = append(recordList, recordMap)
					}
				}
			}
			if len(recordList) > 0 {
				// 成功转换
			} else {
				return nil, fmt.Errorf("unsupported records type")
			}
		} else {
			return nil, fmt.Errorf("unsupported records type")
		}
	}

	// 1. 为每个记录添加时间粒度键
	for i := range recordList {
		if recordList[i] == nil {
			continue
		}

		timestamp, ok := recordList[i][timeField].(time.Time)
		if !ok {
			return nil, fmt.Errorf("field %s is not a time.Time", timeField)
		}

		// 添加时间粒度键
		timeKey := FormatTimeByGranularity(timestamp, granularity)
		recordList[i]["time_granularity_key"] = timeKey
	}

	// 2. 按时间粒度键分组
	// 这里使用简单的map分组，而不是依赖record包的OperationVertical
	groupedData := make(map[string][]map[string]any)
	for _, record := range recordList {
		if record == nil {
			continue
		}

		timeKey, ok := record["time_granularity_key"].(string)
		if !ok {
			continue
		}

		groupedData[timeKey] = append(groupedData[timeKey], record)
	}

	// 3. 对每个分组应用聚合函数
	var results []TimeAggregationResult
	for timeKey, groupRecords := range groupedData {
		// 计算聚合值
		var aggregatedValue float64
		var sum float64
		var count int
		var min float64 = 1<<63 - 1
		var max float64 = -1

		for _, record := range groupRecords {
			if record == nil {
				continue
			}

			val, ok := record[valueField].(float64)
			if !ok {
				continue
			}

			sum += val
			count++
			if val < min {
				min = val
			}
			if val > max {
				max = val
			}
		}

		// 根据聚合类型计算结果
		switch aggregationType {
		case "sum":
			aggregatedValue = sum
		case "avg":
			if count > 0 {
				aggregatedValue = sum / float64(count)
			}
		case "count":
			aggregatedValue = float64(count)
		case "max":
			aggregatedValue = max
		case "min":
			aggregatedValue = min
		default:
			return nil, fmt.Errorf("unsupported aggregation type: %s", aggregationType)
		}

		// 添加聚合结果
		results = append(results, TimeAggregationResult{
			TimeKey: timeKey,
			Value:   aggregatedValue,
		})
	}

	return results, nil
}

// getSliceElement 使用反射获取切片中的元素
func getSliceElement(slice interface{}, index int) (interface{}, bool) {
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice {
		return nil, false
	}
	if index < 0 || index >= v.Len() {
		return nil, false
	}
	return v.Index(index).Interface(), true
}

// TimeBucket 将时间戳分配到指定粒度的时间桶
func TimeBucket(t time.Time, duration time.Duration) time.Time {
	// 计算时间桶的起始时间，保持原始时区
	unixNano := t.UnixNano()
	bucketNano := unixNano / int64(duration) * int64(duration)

	// 使用 time.Date 重建时间，保持原始时区信息
	loc := t.Location()
	t = time.Unix(0, bucketNano).In(loc)

	return t
}

// TimeRange 根据时间粒度计算时间范围
func TimeRange(start time.Time, granularity TimeGranularity) (time.Time, time.Time) {
	var duration time.Duration
	switch granularity {
	case TimeGranularityHour:
		duration = time.Hour
	case TimeGranularityDay:
		duration = 24 * time.Hour
	case TimeGranularityMonth:
		// 简化处理，实际需要更复杂的逻辑
		duration = 30 * 24 * time.Hour
	default:
		duration = time.Hour
	}

	end := start.Add(duration)
	return start, end
}
