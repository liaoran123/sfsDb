package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Long:  "Manage database configuration including getting, setting, and optimizing",
		Run: func(cmd *cobra.Command, args []string) {
			if err := initStore(); err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			defer closeStore()

			configMgr := manager.ConfigManager()
			
			// 获取当前配置
			fmt.Println("=== Configuration Management ===")
			fmt.Println("\n1. Current Configuration:")
			config, err := configMgr.GetConfig()
			if err != nil {
				fmt.Printf("Error getting config: %v\n", err)
				return
			}
			
			fmt.Printf("  Store Type: %s\n", config.StoreType)
			fmt.Println("  Options:")
			for key, value := range config.Options {
				fmt.Printf("    %s: %s\n", key, value)
			}

			// 获取优化建议
			fmt.Println("\n2. Optimization Suggestions:")
			suggestions, err := configMgr.GetOptimizationSuggestions()
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
