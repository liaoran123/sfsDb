package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewSystemCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "system",
		Short: "Get system information",
		Long:  "Get system information including tables, fields, and indexes",
		Run: func(cmd *cobra.Command, args []string) {
			if err := initStore(); err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			defer closeStore()

			systemMgr := manager.SystemManager()
			systemInfo, err := systemMgr.GetAllSystemInfo()
			if err != nil {
				fmt.Printf("Error getting system information: %v\n", err)
				return
			}

			fmt.Println("=== System Information ===")
			
			// 打印表信息
			if tables, ok := systemInfo["tables"].([]interface{}); ok {
				fmt.Println("\nTables:")
				for _, table := range tables {
					if tableMap, ok := table.(map[string]interface{}); ok {
						name := tableMap["Name"]
						id := tableMap["ID"]
						fmt.Printf("  - %v (ID: %v)\n", name, id)
					}
				}
			}

			// 打印表详情
			if tableDetails, ok := systemInfo["tableDetails"].(map[interface{}]interface{}); ok {
				for tableID, details := range tableDetails {
					if detailsMap, ok := details.(map[string]interface{}); ok {
						name := detailsMap["name"]
						fmt.Printf("\nTable: %v (ID: %v)\n", name, tableID)
						
						// 打印字段信息
						if fields, ok := detailsMap["fields"].([]interface{}); ok {
							fmt.Println("  Fields:")
							for _, field := range fields {
								if fieldMap, ok := field.(map[string]interface{}); ok {
									fieldName := fieldMap["Name"]
									fieldID := fieldMap["ID"]
									fmt.Printf("    - %v (ID: %v)\n", fieldName, fieldID)
								}
							}
						}
						
						// 打印索引信息
						if indexes, ok := detailsMap["indexes"].([]interface{}); ok {
							fmt.Println("  Indexes:")
							for _, index := range indexes {
								if indexMap, ok := index.(map[string]interface{}); ok {
									indexName := indexMap["Name"]
									indexID := indexMap["ID"]
									fmt.Printf("    - %v (ID: %v)\n", indexName, indexID)
								}
							}
						}
					}
				}
			}
		},
	}
	return cmd
}
