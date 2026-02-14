package time

import "time"

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

/*
// 1. 执行时间范围查询
startTime := time.Now().Add(-24 * time.Hour)
endTime := time.Now()
iter, err := deviceTable.SearchRange("timestamp", startTime, endTime)
defer engine.GlobalTableIterPool.Put(iter)

// 2. 获取记录集
records := iter.GetRecordSet(true)
defer record.PutRecords(records)

// 3. 按小时聚合
hourlyResults, err := AggregateByTimeGranularity(
    records,
    "timestamp",    // 时间字段
    "value",        // 值字段
    TimeGranularityHour,  // 时间粒度
    "sum",          // 聚合类型
)

// 4. 输出结果
for _, result := range hourlyResults {
    fmt.Printf("Hour: %s, Sum: %.2f\n", result.TimeKey, result.Value)
}
*/
