package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sfsdb",
	Short: "sfsDb database management tool",
	Long: `sfsDb is a lightweight, high-performance embedded database library for Go.
This tool provides management capabilities for sfsDb databases.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("sfsDb management tool")
		fmt.Println("Use 'sfsdb --help' for more information about available commands.")
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "./kvdb", "Database path")
	rootCmd.AddCommand(
		NewStatusCmd(),
		NewSystemCmd(),
		NewIndexCmd(),
		NewConfigCmd(),
		NewBackupCmd(),
		NewStatsCmd(),
		NewMonitorCmd(),
	)
}

func NewRootCmd() *cobra.Command {
	return rootCmd
}
