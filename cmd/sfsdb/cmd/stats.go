package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewStatsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Get performance statistics",
		Long:  "Get performance statistics including query stats and hotspots",
		Run: func(cmd *cobra.Command, args []string) {
			if err := initStore(); err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			defer closeStore()

			statsMgr := manager.StatsManager()
			
			// 获取查询统计
			fmt.Println("=== Performance Statistics ===")
			fmt.Println("\n1. Query Statistics:")
			queryStats, err := statsMgr.GetQueryStats()
			if err != nil {
				fmt.Printf("Error getting query stats: %v\n", err)
				return
			}
			
			fmt.Printf("  Total Queries: %d\n", queryStats.Count)
			fmt.Printf("  Total Time: %v\n", queryStats.TotalTime)
			fmt.Printf("  Average Time: %v\n", queryStats.AvgTime)
			fmt.Printf("  Max Time: %v\n", queryStats.MaxTime)
			fmt.Printf("  Min Time: %v\n", queryStats.MinTime)

			// 获取按类型查询统计
			fmt.Println("\n2. Query Stats by Type:")
			queryStatsByType, err := statsMgr.GetQueryStatsByType()
			if err != nil {
				fmt.Printf("Error getting query stats by type: %v\n", err)
				return
			}
			
			if len(queryStatsByType) > 0 {
				for queryType, stats := range queryStatsByType {
					fmt.Printf("  %s:\n", queryType)
					fmt.Printf("    Count: %d\n", stats.Count)
					fmt.Printf("    Avg Time: %v\n", stats.AvgTime)
				}
			} else {
				fmt.Println("  No query stats by type available")
			}

			// 获取热点数据
			fmt.Println("\n3. Hotspot Data:")
			hotspots, err := statsMgr.GetHotspots(10)
			if err != nil {
				fmt.Printf("Error getting hotspots: %v\n", err)
				return
			}
			
			if len(hotspots) > 0 {
				for _, hotspot := range hotspots {
					fmt.Printf("  - %s (Access Count: %d)\n", hotspot.Key, hotspot.Count)
				}
			} else {
				fmt.Println("  No hotspots found")
			}
		},
	}
	return cmd
}
