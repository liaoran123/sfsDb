package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Get database status",
		Long:  "Get current database status including memory usage and storage information",
		Run: func(cmd *cobra.Command, args []string) {
			if err := initStore(); err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			defer closeStore()

			statusInfo, err := manager.GetStatus()
			if err != nil {
				fmt.Printf("Error getting status: %v\n", err)
				return
			}

			fmt.Println("=== Database Status ===")
			fmt.Printf("Memory Usage: %.2f MB\n", float64(statusInfo.Memory.Alloc)/1024/1024)
			fmt.Printf("Total Allocated: %.2f MB\n", float64(statusInfo.Memory.TotalAlloc)/1024/1024)
			fmt.Printf("System Memory: %.2f MB\n", float64(statusInfo.Memory.Sys)/1024/1024)
			fmt.Printf("GC Count: %d\n", statusInfo.Memory.NumGC)
			fmt.Printf("Storage Type: %s\n", statusInfo.Storage.StoreType)
		},
	}
	return cmd
}
