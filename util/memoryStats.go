package util

import (
	"fmt"
	"runtime"
	"time"
)

// MemoryStats 内存使用统计信息结构体
type MemoryStats struct {
	Alloc      uint64 // 当前分配的内存大小（字节）
	TotalAlloc uint64 // 累计分配的内存大小（字节）
	Sys        uint64 // 从系统获取的内存大小（字节）
	NumGC      uint32 // GC 次数
}

// MemoryStatsDiff 内存使用差异
type MemoryStatsDiff struct {
	Before    MemoryStats   // 执行前内存状态
	After     MemoryStats   // 执行后内存状态
	AllocDiff int64         // Alloc 变化量（字节）
	SysDiff   int64         // Sys 变化量（字节）
	NumGCDiff int           // GC 次数变化量
	ExecTime  time.Duration // 执行时间
}

// GetMemoryStats 获取当前内存使用情况
func GetMemoryStats() MemoryStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return MemoryStats{
		Alloc:      m.Alloc,
		TotalAlloc: m.TotalAlloc,
		Sys:        m.Sys,
		NumGC:      m.NumGC,
	}
}

// TrackMemoryUsage 跟踪函数执行的内存使用情况
func TrackMemoryUsage(name string, f func()) MemoryStatsDiff {
	// 执行前内存状态
	before := GetMemoryStats()

	// 开始时间
	start := time.Now()

	// 执行函数
	f()

	// 执行后内存状态
	after := GetMemoryStats()

	// 计算差异
	diff := MemoryStatsDiff{
		Before:    before,
		After:     after,
		AllocDiff: int64(after.Alloc - before.Alloc),
		SysDiff:   int64(after.Sys - before.Sys),
		NumGCDiff: int(after.NumGC - before.NumGC),
		ExecTime:  time.Since(start),
	}

	// 打印分析结果
	fmt.Printf("=== %s 内存使用分析 ===\n", name)
	fmt.Printf("执行时间: %v\n", diff.ExecTime)
	fmt.Printf("内存分配变化: %+.2f MB\n", float64(diff.AllocDiff)/1024/1024)
	fmt.Printf("系统内存变化: %+.2f MB\n", float64(diff.SysDiff)/1024/1024)
	fmt.Printf("GC 次数变化: %+d\n", diff.NumGCDiff)
	fmt.Printf("执行前分配内存: %.2f MB\n", float64(before.Alloc)/1024/1024)
	fmt.Printf("执行后分配内存: %.2f MB\n", float64(after.Alloc)/1024/1024)
	fmt.Println("========================")
	fmt.Println()

	return diff
}

// TrackMemoryUsageWithResult 跟踪函数执行的内存使用情况（带返回值）
func TrackMemoryUsageWithResult[T any](name string, f func() T) (T, MemoryStatsDiff) {
	// 执行前内存状态
	before := GetMemoryStats()

	// 开始时间
	start := time.Now()

	// 执行函数
	result := f()

	// 执行后内存状态
	after := GetMemoryStats()

	// 计算差异
	diff := MemoryStatsDiff{
		Before:    before,
		After:     after,
		AllocDiff: int64(after.Alloc - before.Alloc),
		SysDiff:   int64(after.Sys - before.Sys),
		NumGCDiff: int(after.NumGC - before.NumGC),
		ExecTime:  time.Since(start),
	}

	// 打印分析结果
	fmt.Printf("=== %s 内存使用分析 ===\n", name)
	fmt.Printf("执行时间: %v\n", diff.ExecTime)
	fmt.Printf("内存分配变化: %+.2f MB\n", float64(diff.AllocDiff)/1024/1024)
	fmt.Printf("系统内存变化: %+.2f MB\n", float64(diff.SysDiff)/1024/1024)
	fmt.Printf("GC 次数变化: %+d\n", diff.NumGCDiff)
	fmt.Printf("执行前分配内存: %.2f MB\n", float64(before.Alloc)/1024/1024)
	fmt.Printf("执行后分配内存: %.2f MB\n", float64(after.Alloc)/1024/1024)
	fmt.Println("========================")
	fmt.Println()

	return result, diff
}
