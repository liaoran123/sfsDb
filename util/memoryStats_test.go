package util

import (
	"fmt"
	"testing"
)

// TestTrackMemoryUsage 测试无返回值函数的内存跟踪
func TestTrackMemoryUsage(t *testing.T) {
	fmt.Println("=== 测试无返回值函数的内存跟踪 ===")
	
	// 测试内存分配
	diff := TrackMemoryUsage("内存分配测试", func() {
		// 分配10MB内存
		data := make([]byte, 1024*1024*10)
		_ = data
	})
	
	// 验证内存变化
	if diff.AllocDiff < 0 {
		t.Error("内存分配测试失败: 内存使用应该增加")
	}
	
	fmt.Println("内存分配测试通过")
}

// TestTrackMemoryUsageWithResult 测试有返回值函数的内存跟踪
func TestTrackMemoryUsageWithResult(t *testing.T) {
	fmt.Println("=== 测试有返回值函数的内存跟踪 ===")
	
	// 测试生成大切片
	result, diff := TrackMemoryUsageWithResult("大切片生成测试", func() []int {
		// 生成一个包含100万个元素的切片
		data := make([]int, 1000000)
		for i := range data {
			data[i] = i
		}
		return data
	})
	
	// 验证结果
	if len(result) != 1000000 {
		t.Errorf("大切片生成测试失败: 期望长度为1000000, 实际长度为%d", len(result))
	}
	
	// 验证内存变化
	if diff.AllocDiff < 0 {
		t.Error("大切片生成测试失败: 内存使用应该增加")
	}
	
	fmt.Println("大切片生成测试通过")
}

// TestMemoryStats 获取内存状态测试
func TestMemoryStats(t *testing.T) {
	fmt.Println("=== 测试获取内存状态 ===")
	
	// 获取内存状态
	stats := GetMemoryStats()
	
	// 验证内存状态有效
	if stats.Sys == 0 {
		t.Error("获取内存状态测试失败: 系统内存应该大于0")
	}
	
	fmt.Printf("当前内存状态: Alloc=%.2fMB, Sys=%.2fMB, NumGC=%d\n", 
		float64(stats.Alloc)/1024/1024, 
		float64(stats.Sys)/1024/1024, 
		stats.NumGC)
	fmt.Println("获取内存状态测试通过")
}
