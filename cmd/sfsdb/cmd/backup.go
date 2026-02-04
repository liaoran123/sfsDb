package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewBackupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Manage backups",
		Long:  "Manage database backups including creating and restoring",
		Run: func(cmd *cobra.Command, args []string) {
			if err := initStore(); err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			defer closeStore()

			backupMgr := manager.BackupManager()
			
			// 创建备份
			fmt.Println("=== Backup Management ===")
			fmt.Println("\n1. Creating Backup:")
			backupPath, err := backupMgr.Backup("./backups")
			if err != nil {
				fmt.Printf("Error creating backup: %v\n", err)
				return
			}
			fmt.Printf("  Backup created successfully: %s\n", backupPath)

			// 验证备份
			fmt.Println("\n2. Validating Backup:")
			isValid, err := backupMgr.ValidateBackup(backupPath)
			if err != nil {
				fmt.Printf("Error validating backup: %v\n", err)
				return
			}
			if isValid {
				fmt.Println("  Backup validation: SUCCESS")
			} else {
				fmt.Println("  Backup validation: FAILED")
			}

			fmt.Println("\n3. To restore from backup:")
			fmt.Printf("  Run: sfsdb backup restore %s\n", backupPath)
		},
	}
	return cmd
}
