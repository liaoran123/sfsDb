package time

import "time"

// TimeGranularity 时间粒度类型
type TimeGranularity string

const (
	TimeGranularitySecond TimeGranularity = "second"
	TimeGranularityMinute TimeGranularity = "minute"
	TimeGranularityHour   TimeGranularity = "hour"
	TimeGranularityDay    TimeGranularity = "day"
	TimeGranularityMonth  TimeGranularity = "month"
	TimeGranularityYear   TimeGranularity = "year"
)

//时间粒度转换工具函数
// FormatTimeByGranularity 根据时间粒度格式化时间
func FormatTimeByGranularity(t time.Time, granularity TimeGranularity) string {
	switch granularity {
	case TimeGranularitySecond:
		return t.Format("2006-01-02 15:04:05")
	case TimeGranularityMinute:
		return t.Format("2006-01-02 15:04:00")
	case TimeGranularityHour:
		return t.Format("2006-01-02 15:00:00")
	case TimeGranularityDay:
		return t.Format("2006-01-02")
	case TimeGranularityMonth:
		return t.Format("2006-01")
	case TimeGranularityYear:
		return t.Format("2006")
	default:
		return t.Format("2006-01-02 15:04:05")
	}
}

// TimeToUnixTimestamp 将 time.Time 转换为秒级 Unix 时间戳
func TimeToUnixTimestamp(t time.Time) int64 {
	return t.Unix()
}

// TimeToUnixTimestampMs 将 time.Time 转换为毫秒级 Unix 时间戳
func TimeToUnixTimestampMs(t time.Time) int64 {
	return t.UnixMilli()
}

// TimeToUnixTimestampNs 将 time.Time 转换为纳秒级 Unix 时间戳
func TimeToUnixTimestampNs(t time.Time) int64 {
	return t.UnixNano()
}

// UnixTimestampToTime 将秒级 Unix 时间戳转换为 time.Time
func UnixTimestampToTime(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}

// UnixTimestampMsToTime 将毫秒级 Unix 时间戳转换为 time.Time
func UnixTimestampMsToTime(timestampMs int64) time.Time {
	return time.Unix(0, timestampMs*int64(time.Millisecond))
}

// UnixTimestampNsToTime 将纳秒级 Unix 时间戳转换为 time.Time
func UnixTimestampNsToTime(timestampNs int64) time.Time {
	return time.Unix(0, timestampNs)
}
