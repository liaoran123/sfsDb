package cmd

import (
	"fmt"
	"time"

	"github.com/liaoran123/sfsDb/management"
	"github.com/spf13/cobra"
)

func NewMonitorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "Monitor database",
		Long:  "Monitor database status and key changes",
		Run: func(cmd *cobra.Command, args []string) {
			if err := initStore(); err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			defer closeStore()

			// 获取键值变化统计
			fmt.Println("=== Database Monitoring ===")
			fmt.Println("\n1. Key Change Statistics:")
			monitorMgr := manager.MonitorManager()
			keyChangeStats := monitorMgr.GetKeyChangeStats()

			fmt.Println("  Put Counters:")
			if len(keyChangeStats.PutCounters) > 0 {
				for key, count := range keyChangeStats.PutCounters {
					fmt.Printf("    %s: %d\n", key, count)
				}
			} else {
				fmt.Println("    No put operations recorded")
			}

			fmt.Println("  Delete Counters:")
			if len(keyChangeStats.DeleteCounters) > 0 {
				for key, count := range keyChangeStats.DeleteCounters {
					fmt.Printf("    %s: %d\n", key, count)
				}
			} else {
				fmt.Println("    No delete operations recorded")
			}

			// 启动监控器
			fmt.Println("\n2. Starting Monitor:")
			fmt.Println("  Monitoring database status with thresholds...")
			fmt.Println("  Press Ctrl+C to stop...")

			monitor := manager.Monitor(time.Second*5, management.Thresholds{
				MemoryUsage: 1024, // 1GB
				GCCount:     100,
			})

			if err := monitor.Start(); err != nil {
				fmt.Printf("Error starting monitor: %v\n", err)
				return
			}

			// 等待一段时间
			time.Sleep(time.Second * 15)

			monitor.Stop()
			fmt.Println("\n3. Monitor stopped")
		},
	}
	return cmd
}
