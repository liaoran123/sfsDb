package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewIndexCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "index",
		Short: "Manage indexes",
		Long:  "Manage indexes including listing, analyzing, and optimizing",
		Run: func(cmd *cobra.Command, args []string) {
			if err := initStore(); err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			defer closeStore()

			indexMgr := manager.IndexManager()
			
			// 列出所有索引
			fmt.Println("=== Index Management ===")
			fmt.Println("\n1. Listing Indexes:")
			indexes, err := indexMgr.ListIndexes("")
			if err != nil {
				fmt.Printf("Error listing indexes: %v\n", err)
				return
			}
			
			if len(indexes) == 0 {
				fmt.Println("  No indexes found")
			} else {
				for _, index := range indexes {
					fmt.Printf("  - %s (Type: %s)\n", index.Name, index.Type)
					fmt.Printf("    Fields: %v\n", index.Fields)
				}
			}

			// 分析索引使用情况
			fmt.Println("\n2. Analyzing Indexes:")
			analysis, err := indexMgr.AnalyzeIndexes("")
			if err != nil {
				fmt.Printf("Error analyzing indexes: %v\n", err)
				return
			}
			
			if len(analysis.UnusedIndexes) > 0 {
				fmt.Println("  Unused Indexes:")
				for _, indexName := range analysis.UnusedIndexes {
					fmt.Printf("    - %s\n", indexName)
				}
			} else {
				fmt.Println("  No unused indexes found")
			}

			// 提供优化建议
			fmt.Println("\n3. Optimization Suggestions:")
			suggestions, err := indexMgr.OptimizeIndexes("")
			if err != nil {
				fmt.Printf("Error getting optimization suggestions: %v\n", err)
				return
			}
			
			if len(suggestions) > 0 {
				for _, suggestion := range suggestions {
					fmt.Printf("  - %s\n", suggestion)
				}
			} else {
				fmt.Println("  No optimization suggestions")
			}
		},
	}
	return cmd
}
