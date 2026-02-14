package time

import (
	"fmt"
	"time"
)

// TimeGranularity 时间粒度类型
type TimeGranularity string

const (
	TimeGranularityMillisecond TimeGranularity = "millisecond"
	TimeGranularityMicrosecond TimeGranularity = "microsecond"
	TimeGranularitySecond      TimeGranularity = "second"
	TimeGranularityMinute      TimeGranularity = "minute"
	TimeGranularityHour        TimeGranularity = "hour"
	TimeGranularityDay         TimeGranularity = "day"
	TimeGranularityWeek        TimeGranularity = "week"
	TimeGranularityMonth       TimeGranularity = "month"
	TimeGranularityQuarter     TimeGranularity = "quarter"
	TimeGranularityYear        TimeGranularity = "year"
)

// IsWeekday 检查给定时间是否为工作日
func IsWeekday(t time.Time) bool {
	weekday := t.Weekday()
	return weekday != time.Saturday && weekday != time.Sunday
}

// IsWeekend 检查给定时间是否为周末
func IsWeekend(t time.Time) bool {
	weekday := t.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// GetQuarter 获取给定时间所在的季度
func GetQuarter(t time.Time) int {
	month := int(t.Month())
	return (month-1)/3 + 1
}

// GetWeekNumber 获取给定时间所在的周数（一年中的第几周）
func GetWeekNumber(t time.Time) int {
	_, week := t.ISOWeek()
	return week
}

// 时间粒度转换工具函数
// FormatTimeByGranularity 根据时间粒度格式化时间
func FormatTimeByGranularity(t time.Time, granularity TimeGranularity) string {
	switch granularity {
	case TimeGranularityMillisecond:
		return t.Format("2006-01-02 15:04:05.000")
	case TimeGranularityMicrosecond:
		return t.Format("2006-01-02 15:04:05.000000")
	case TimeGranularitySecond:
		return t.Format("2006-01-02 15:04:05")
	case TimeGranularityMinute:
		return t.Format("2006-01-02 15:04:00")
	case TimeGranularityHour:
		return t.Format("2006-01-02 15:00:00")
	case TimeGranularityDay:
		return t.Format("2006-01-02")
	case TimeGranularityWeek:
		_, week := t.ISOWeek()
		return t.Format("2006-") + fmt.Sprintf("W%02d", week)
	case TimeGranularityMonth:
		return t.Format("2006-01")
	case TimeGranularityQuarter:
		quarter := GetQuarter(t)
		return t.Format("2006-") + fmt.Sprintf("Q%d", quarter)
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
