package main

import (
	"fmt"
	"time"

	ttime "github.com/liaoran123/sfsDb/time"
)

func main() {
	fmt.Println("=== sfsDb 时序数据处理示例 ===")

	example1TimeGranularity()
	example2SlidingWindow()
	example3TumblingWindow()
	example4Aggregation()
}

func example1TimeGranularity() {
	fmt.Println("\n--- 示例1: 时间粒度处理 ---")

	now := time.Now()
	fmt.Printf("当前时间: %s\n", now.Format(time.RFC3339))

	// 不同时间粒度的格式化
	formattedSecond := ttime.FormatTimeByGranularity(now, ttime.TimeGranularitySecond)
	formattedMinute := ttime.FormatTimeByGranularity(now, ttime.TimeGranularityMinute)
	formattedHour := ttime.FormatTimeByGranularity(now, ttime.TimeGranularityHour)
	formattedDay := ttime.FormatTimeByGranularity(now, ttime.TimeGranularityDay)

	fmt.Printf("按秒格式化:   %s\n", formattedSecond)
	fmt.Printf("按分钟格式化: %s\n", formattedMinute)
	fmt.Printf("按小时格式化: %s\n", formattedHour)
	fmt.Printf("按天格式化:   %s\n", formattedDay)

	// 检查是否是工作日/周末
	fmt.Printf("是否是工作日: %v\n", ttime.IsWeekday(now))
	fmt.Printf("是否是周末:   %v\n", ttime.IsWeekend(now))

	// 获取季度和周数
	fmt.Printf("当前季度: Q%d\n", ttime.GetQuarter(now))
	fmt.Printf("当前周数: 第%d周\n", ttime.GetWeekNumber(now))

	// 获取时间戳
	timestampSec := ttime.TimeToUnixTimestamp(now)
	timestampMs := ttime.TimeToUnixTimestampMs(now)
	fmt.Printf("秒时间戳:   %d\n", timestampSec)
	fmt.Printf("毫秒时间戳: %d\n", timestampMs)

	// 从时间戳还原
	restoredFromSec := ttime.UnixTimestampToTime(timestampSec)
	fmt.Printf("从秒时间戳还原: %s\n", restoredFromSec.Format(time.RFC3339))
}

func example2SlidingWindow() {
	fmt.Println("\n--- 示例2: 滑动窗口 ---")

	// 创建滑动窗口
	startTime := time.Now().Add(-1 * time.Hour)
	endTime := time.Now()
	windowSize := 15 * time.Minute
	stepSize := 5 * time.Minute

	window := ttime.NewSlidingWindow(startTime, endTime, windowSize, stepSize)

	fmt.Printf("滑动窗口配置:\n")
	fmt.Printf("  起始时间: %s\n", startTime.Format("15:04:05"))
	fmt.Printf("  结束时间: %s\n", endTime.Format("15:04:05"))
	fmt.Printf("  窗口大小: %v\n", windowSize)
	fmt.Printf("  滑动步长: %v\n", stepSize)

	fmt.Println("\n窗口序列:")
	i := 1
	window.Reset()
	for window.Next() {
		fmt.Printf("  窗口%d: %s - %s\n",
			i,
			window.Start().Format("15:04:05"),
			window.End().Format("15:04:05"))
		i++
	}
}

func example3TumblingWindow() {
	fmt.Println("\n--- 示例3: 滚动窗口 ---")

	// 创建滚动窗口
	startTime := time.Now().Add(-2 * time.Hour)
	endTime := time.Now()
	windowSize := 30 * time.Minute

	window := ttime.NewTumblingWindow(startTime, endTime, windowSize)

	fmt.Printf("滚动窗口配置:\n")
	fmt.Printf("  起始时间: %s\n", startTime.Format("15:04:05"))
	fmt.Printf("  结束时间: %s\n", endTime.Format("15:04:05"))
	fmt.Printf("  窗口大小: %v (不重叠)\n", windowSize)

	fmt.Println("\n窗口序列:")
	i := 1
	window.Reset()
	for window.Next() {
		fmt.Printf("  窗口%d: %s - %s\n",
			i,
			window.Start().Format("15:04:05"),
			window.End().Format("15:04:05"))
		i++
	}
}

func example4Aggregation() {
	fmt.Println("\n--- 示例4: 窗口聚合 ---")

	// 创建测试数据
	records := []map[string]any{}
	baseTime := time.Now().Add(-1 * time.Hour)

	for i := 0; i < 20; i++ {
		timestamp := baseTime.Add(time.Duration(i*3) * time.Minute)
		value := 10.0 + float64(i*2)
		records = append(records, map[string]any{
			"time":  timestamp,
			"value": value,
		})
	}

	fmt.Printf("测试数据: %d 条记录\n", len(records))
	fmt.Println("前5条记录:")
	for i := 0; i < 5 && i < len(records); i++ {
		fmt.Printf("  %d. 时间=%s, 值=%.1f\n",
			i+1,
			records[i]["time"].(time.Time).Format("15:04:05"),
			records[i]["value"].(float64))
	}

	// 创建滚动窗口
	startTime := baseTime
	endTime := time.Now()
	windowSize := 20 * time.Minute
	window := ttime.NewTumblingWindow(startTime, endTime, windowSize)

	// 进行聚合
	aggTypes := []string{"sum", "avg", "max", "min", "count"}
	for _, aggType := range aggTypes {
		results, err := ttime.AggregateByWindow(records, "time", "value", window, aggType)
		if err != nil {
			fmt.Printf("聚合失败 (%s): %v\n", aggType, err)
			continue
		}

		fmt.Printf("\n%s 聚合结果:\n", aggType)
		for _, result := range results {
			fmt.Printf("  窗口 %s - %s: %.2f\n",
				result.WindowStart.Format("15:04"),
				result.WindowEnd.Format("15:04"),
				result.Value)
		}
	}
}
