package main

import (
	"fmt"
	"time"

	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/management"
	"github.com/liaoran123/sfsDb/storage"
)

// ExampleManager 使用示例
// 展示如何使用管理工具库的各个功能

func ExampleManager() {
	// 打开数据库
	store, err := storage.OpenDefaultDb("./kvdb")
	if err != nil {
		fmt.Printf("打开数据库失败: %v\n", err)
		return
	}
	defer storage.CloseDb()

	// 创建管理器
	manager := management.NewManager(store)

	// 1. 获取数据库状态
	statusInfo, err := manager.GetStatus()
	if err != nil {
		fmt.Printf("获取状态失败: %v\n", err)
	} else {
		fmt.Println("=== 数据库状态 ===")
		fmt.Printf("内存使用: %.2f MB\n", float64(statusInfo.Memory.Alloc)/1024/1024)
		fmt.Printf("累计分配: %.2f MB\n", float64(statusInfo.Memory.TotalAlloc)/1024/1024)
		fmt.Printf("系统内存: %.2f MB\n", float64(statusInfo.Memory.Sys)/1024/1024)
		fmt.Printf("GC 次数: %d\n", statusInfo.Memory.NumGC)
		fmt.Printf("存储类型: %s\n", statusInfo.Storage.StoreType)
	}

	// 2. 索引管理
	indexMgr := manager.IndexManager()
	indexes, err := indexMgr.ListIndexes("test_table")
	if err != nil {
		fmt.Printf("获取索引失败: %v\n", err)
	} else {
		fmt.Println("\n=== 索引列表 ===")
		for _, idx := range indexes {
			fmt.Printf("索引名称: %s, 类型: %s, 字段: %v\n", idx.Name, idx.Type, idx.Fields)
		}
	}

	// 3. 性能统计
	statsMgr := manager.StatsManager()
	queryStats, err := statsMgr.GetQueryStats()
	if err != nil {
		fmt.Printf("获取查询统计失败: %v\n", err)
	} else {
		fmt.Println("\n=== 查询性能统计 ===")
		fmt.Printf("查询次数: %d\n", queryStats.Count)
		fmt.Printf("总耗时: %v\n", queryStats.TotalTime)
		fmt.Printf("平均耗时: %v\n", queryStats.AvgTime)
		fmt.Printf("最大耗时: %v\n", queryStats.MaxTime)
		fmt.Printf("最小耗时: %v\n", queryStats.MinTime)
	}

	// 4. 备份数据库
	backupMgr := manager.BackupManager()
	backupFile, err := backupMgr.Backup("./backup")
	if err != nil {
		fmt.Printf("备份失败: %v\n", err)
	} else {
		fmt.Printf("\n备份成功，备份文件: %s\n", backupFile)
	}

	// 5. 验证备份
	if backupFile != "" {
		valid, err := backupMgr.ValidateBackup(backupFile)
		if err != nil {
			fmt.Printf("验证备份失败: %v\n", err)
		} else if valid {
			fmt.Println("备份文件验证成功")
		} else {
			fmt.Println("备份文件验证失败")
		}
	}

	// 6. 配置管理
	configMgr := manager.ConfigManager()
	configInfo, err := configMgr.GetConfig()
	if err != nil {
		fmt.Printf("获取配置失败: %v\n", err)
	} else {
		fmt.Println("\n=== 配置信息 ===")
		fmt.Printf("存储类型: %s\n", configInfo.StoreType)
		fmt.Println("配置选项:")
		for key, value := range configInfo.Options {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	// 获取优化建议
	suggestions, err := configMgr.GetOptimizationSuggestions()
	if err != nil {
		fmt.Printf("获取优化建议失败: %v\n", err)
	} else {
		fmt.Println("\n=== 优化建议 ===")
		for _, suggestion := range suggestions {
			fmt.Printf("  - %s\n", suggestion)
		}
	}

	fmt.Println("\n基本功能示例完成！")
}

// ExampleManagerWithTable 使用带表实例的管理器
// 展示如何使用深度集成功能

func ExampleManagerWithTable() {
	// 打开数据库
	store, err := storage.OpenDefaultDb("./kvdb")
	if err != nil {
		fmt.Printf("打开数据库失败: %v\n", err)
		return
	}
	defer storage.CloseDb()

	// 创建表
	table, err := engine.TableNew("test_table")
	if err != nil {
		fmt.Printf("创建表失败: %v\n", err)
		return
	}

	// 设置表字段
	fields := map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	}

	if err := table.SetFields(fields); err != nil {
		fmt.Printf("设置字段失败: %v\n", err)
		return
	}

	// 创建带表实例的管理器
	manager := management.NewManagerWithTable(store, table)

	// 使用深度集成功能
	indexMgr := manager.IndexManager()
	indexes, err := indexMgr.ListIndexes("test_table")
	if err != nil {
		fmt.Printf("获取索引失败: %v\n", err)
	} else {
		fmt.Println("=== 深度集成 - 索引列表 ===")
		for _, idx := range indexes {
			fmt.Printf("索引名称: %s, 类型: %s, 字段: %v\n", idx.Name, idx.Type, idx.Fields)
		}
	}

	fmt.Println("\n深度集成功能示例完成！")
}

// ExampleMonitor 使用监控功能
// 展示如何使用监控告警功能

func ExampleMonitor() {
	// 打开数据库
	store, err := storage.OpenDefaultDb("./kvdb")
	if err != nil {
		fmt.Printf("打开数据库失败: %v\n", err)
		return
	}
	defer storage.CloseDb()

	// 创建管理器
	manager := management.NewManager(store)

	// 定义监控阈值
	thresholds := management.Thresholds{
		MemoryUsage: 100.0, // 100MB
		GCCount:     10,    // 10次
	}

	// 创建监控器
	monitor := manager.Monitor(5*time.Second, thresholds)

	// 启动监控
	if err := monitor.Start(); err != nil {
		fmt.Printf("启动监控失败: %v\n", err)
		return
	}

	// 运行一段时间
	fmt.Println("监控已启动，将运行10秒...")
	time.Sleep(10 * time.Second)

	// 停止监控
	monitor.Stop()

	fmt.Println("\n监控功能示例完成！")
}

func main() {
	fmt.Println("=== 管理工具库使用示例 ===")
	ExampleManager()
	ExampleManagerWithTable()
	ExampleMonitor()
}
