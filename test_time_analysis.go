package main

import (
	"fmt"
	"math"
	"time"

	"github.com/liaoran123/sfsDb/record"
)

// TestTimeRangeAnalysis 测试基于时间范围的数据分析
func TestTimeRangeAnalysis() {
	// 这里应该是实际的数据库初始化代码
	// 为了测试，我们模拟一些数据
	fmt.Println("=== 测试基于时间范围的数据分析 ===")

	// 创建时间范围
	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

	fmt.Printf("时间范围: %s 至 %s\n", startTime, endTime)

	// 模拟数据
	mockResults := []record.Record{
		{"timestamp": time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), "value": 10.5},
		{"timestamp": time.Date(2024, 1, 2, 11, 0, 0, 0, time.UTC), "value": 15.2},
		{"timestamp": time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC), "value": 8.7},
		{"timestamp": time.Date(2024, 1, 4, 13, 0, 0, 0, time.UTC), "value": 20.1},
		{"timestamp": time.Date(2024, 1, 5, 14, 0, 0, 0, time.UTC), "value": 12.3},
	}

	// 计算基本统计信息
	var sum, count float64
	var min, max float64 = math.MaxFloat64, -math.MaxFloat64

	for _, result := range mockResults {
		value := result["value"].(float64)
		sum += value
		count++
		if value < min {
			min = value
		}
		if value > max {
			max = value
		}
	}

	// 输出分析结果
	fmt.Printf("数据量: %d\n", int(count))
	fmt.Printf("平均值: %.2f\n", sum/count)
	fmt.Printf("最小值: %.2f\n", min)
	fmt.Printf("最大值: %.2f\n", max)
	fmt.Println()
}

// TestTimeWindowAnalysis 测试时间窗口分析
func TestTimeWindowAnalysis() {
	fmt.Println("=== 测试时间窗口分析 ===")

	// 定义时间窗口大小（1小时）
	windowSize := time.Hour

	// 初始化窗口分析
	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	currentWindowStart := startTime
	windowResults := make(map[string]map[string]float64)

	// 模拟数据
	mockResults := []record.Record{
		{"timestamp": time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), "value": 10.5},
		{"timestamp": time.Date(2024, 1, 1, 10, 30, 0, 0, time.UTC), "value": 12.3},
		{"timestamp": time.Date(2024, 1, 1, 11, 15, 0, 0, time.UTC), "value": 15.7},
		{"timestamp": time.Date(2024, 1, 1, 12, 45, 0, 0, time.UTC), "value": 8.9},
	}

	// 遍历数据并按窗口分组
	for _, result := range mockResults {
		timestamp := result["timestamp"].(time.Time)
		value := result["value"].(float64)

		// 确定所属窗口
		windowKey := currentWindowStart.Format("2006-01-02 15:04:05")

		// 初始化窗口统计
		if _, exists := windowResults[windowKey]; !exists {
			windowResults[windowKey] = map[string]float64{
				"sum":   0,
				"count": 0,
				"min":   math.MaxFloat64,
				"max":   -math.MaxFloat64,
			}
		}

		// 更新窗口统计
		stats := windowResults[windowKey]
		stats["sum"] += value
		stats["count"]++
		if value < stats["min"] {
			stats["min"] = value
		}
		if value > stats["max"] {
			stats["max"] = value
		}

		// 移动到下一个窗口
		if timestamp.Sub(currentWindowStart) >= windowSize {
			currentWindowStart = currentWindowStart.Add(windowSize)
		}
	}

	// 输出窗口分析结果
	for window, stats := range windowResults {
		fmt.Printf("窗口: %s\n", window)
		fmt.Printf("  平均值: %.2f\n", stats["sum"]/stats["count"])
		fmt.Printf("  最小值: %.2f\n", stats["min"])
		fmt.Printf("  最大值: %.2f\n", stats["max"])
		fmt.Printf("  数据量: %.0f\n", stats["count"])
	}
	fmt.Println()
}

func mainRangeAnalysis() {
	// 测试基于时间范围的数据分析
	TestTimeRangeAnalysis()

	// 测试时间窗口分析
	TestTimeWindowAnalysis()

	fmt.Println("=== 测试完成 ===")
}
