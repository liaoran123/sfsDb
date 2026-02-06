# CLI Command Reference

This document provides a detailed guide for using the sfsDb command-line tool, including all available commands and their functions.

## Overview

The sfsDb command-line tool provides the following core functions:

- **Database Status Management**: View current database status
- **System Information Management**: Get table, field, index and other system information
- **Index Management**: Analyze and manage database indexes
- **Configuration Management**: View and modify database configuration
- **Backup and Restore**: Create and restore database backups
- **Performance Statistics**: View query performance and hotspot data
- **Monitoring and Alerting**: Real-time database status monitoring
- **Web Interface Management**: Enable/disable and configure the web interface

## Basic Usage

### Command Format

```bash
sfsdb [command] [parameters]
```

### Global Parameters

| Parameter | Description | Default Value |
|-----------|-------------|---------------|
| `--db` | Database path | `./kvdb` |
| `--help` | Show help information | - |

## Command Details

### 1. status - Get Database Status

**Function**: Get current database status, including memory usage and storage information.

**Usage Examples**:

```bash
# Get status of default path database
sfsdb status

# Get status of specified path database
sfsdb status --db ./mydb
```

**Output Example**:

```
=== Database Status ===
Memory Usage: 1.23 MB
Total Allocated: 2.45 MB
System Memory: 5.67 MB
GC Count: 10
Storage Type: LevelDB
```

### 2. system - System Information Management

**Function**: Get table, field, index and other system information.

**Usage Example**:

```bash
# Get system information
sfsdb system
```

### 3. index - Index Management

**Function**: Analyze and manage database indexes.

**Usage Example**:

```bash
# Manage indexes
sfsdb index
```

### 4. config - Configuration Management

**Function**: View and modify database configuration.

**Usage Example**:

```bash
# Manage configuration
sfsdb config
```

### 5. backup - Backup and Restore

**Function**: Create and restore database backups.

**Usage Example**:

```bash
# Manage backups
sfsdb backup
```

### 6. stats - Performance Statistics

**Function**: View query performance and hotspot data.

**Usage Example**:

```bash
# View performance statistics
sfsdb stats
```

### 7. monitor - Monitoring and Alerting

**Function**: Real-time database status monitoring.

**Usage Example**:

```bash
# Start monitoring
sfsdb monitor
```

### 8. web - Web Interface Management

**Function**: Enable/disable and configure the web interface.

**Subcommand Parameters**:

| Parameter | Description | Default Value |
|-----------|-------------|---------------|
| `--enable` `-e` | Enable web interface | false |
| `--disable` `-d` | Disable web interface | false |
| `--start` `-s` | Start web server | false |
| `--port` `-p` | Web server port | `:8083` |

**Usage Examples**:

```bash
# Enable web interface and set port to 8083
sfsdb web --enable --port :8083

# Disable web interface
sfsdb web --disable

# Start web server (using configured port)
sfsdb web --start

# Start web server and specify port (temporarily override configuration)
sfsdb web --start --port :8085

# View current web configuration
sfsdb web
```

**Output Example**:

```
Web Interface Configuration:
Enabled: true
Port: :8084
Address: http://localhost:8084
```

## Example Workflows

### 1. Start Web Interface

```bash
# Enable web interface and set port
sfsdb web --enable --port :8083

# Start web server
sfsdb web --start

# Access web interface
# http://localhost:8083
```

### 2. Backup Database

```bash
# Create database backup
sfsdb backup

# View database status
sfsdb status
```

### 3. Monitor Database

```bash
# Start monitoring
sfsdb monitor

# View performance statistics
sfsdb stats
```

## Other Project Integration

### Overview

Other projects can integrate sfsdb's CLI command functionality to generate their own executable files (such as abc.exe) and use the same command structure to manage the database.

### Simplest Implementation

If users don't want to use the command-line tool, they can also directly enable and configure the web interface through code. Here's the complete simplest implementation:

```go
package main

import (
    "github.com/liaoran123/sfsDb/storage"
    "github.com/liaoran123/sfsDb/web"
    "github.com/liaoran123/sfsDb/management"
)

func main() {
    // Initialize database
    store, _ := storage.OpenDefaultDb("./kvdb")
    defer storage.CloseDb()
    
    // Create manager
    manager := management.NewManager(store)
    
    // Core configuration (only these two lines are needed)
    configMgr := manager.ConfigManager()
    configMgr.SetConfig("web_enable", "true")
    configMgr.SetConfig("web_port", ":8083")
    
    // Start web server
    server := web.NewServer(":8083", manager)
    server.Start()
}
```

### Core Principle

Enabling and configuring the web interface only requires setting two core parameters:

1. **`web_enable`**：Controls whether the web interface is enabled (values: `"true"` or `"false"`)
2. **`web_port`**：Configures the web server port (value: port string like `":8083"`)

These parameters are persistently saved to the database and automatically take effect on the next startup.

### Implementation Methods



#### 2. Using Subcommands Integration (Recommended)

To avoid command conflicts and program restart issues, it is recommended to integrate sfsdb's management functionality using subcommands:

```go
package main

import (
    "fmt"
    "os"

    "github.com/liaoran123/sfsDb/management"
    "github.com/liaoran123/sfsDb/storage"
    "github.com/spf13/cobra"
)

var (
    dbPath string
    manager *management.Manager
)

func main() {
    // Create root command
    rootCmd := &cobra.Command{
        Use:   "myapp",
        Short: "My Application",
        Long:  "My Application with integrated database management",
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Println("My Application")
            fmt.Println("Use 'myapp --help' for more information about available commands.")
        },
    }

    // Add global database path parameter
    rootCmd.PersistentFlags().StringVar(&dbPath, "db", "./kvdb", "Database path")

    // Create database management subcommand
    dbCmd := &cobra.Command{
        Use:   "db",
        Short: "Database management commands",
        Long:  "Commands for managing the embedded database",
        PersistentPreRun: func(cmd *cobra.Command, args []string) {
            // Initialize database and manager (executed only once)
            store, err := storage.NewLevelDBStore(dbPath, nil)
            if err != nil {
                fmt.Fprintf(os.Stderr, "Failed to open database: %v\n", err)
                os.Exit(1)
            }
            manager = management.NewManager(store)
        },
    }

    // Add specific database management subcommands
    dbCmd.AddCommand(
        newStatusCmd(),
        newSystemCmd(),
        newIndexCmd(),
        newConfigCmd(),
        newBackupCmd(),
        newStatsCmd(),
        newMonitorCmd(),
    )

    // Add database management subcommand to root command
    rootCmd.AddCommand(dbCmd)

    // Execute command
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}

// Status command
func newStatusCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "status",
        Short: "Get database status",
        Run: func(cmd *cobra.Command, args []string) {
            status, err := manager.GetStatus()
            if err != nil {
                fmt.Fprintf(os.Stderr, "Failed to get status: %v\n", err)
                return
            }
            fmt.Println("=== Database Status ===")
            fmt.Printf("Memory Usage: %.2f MB\n", float64(status.Memory.Alloc)/1024/1024)
            fmt.Printf("Total Allocated: %.2f MB\n", float64(status.Memory.TotalAlloc)/1024/1024)
            fmt.Printf("System Memory: %.2f MB\n", float64(status.Memory.Sys)/1024/1024)
            fmt.Printf("GC Count: %d\n", status.Memory.NumGC)
            fmt.Printf("Storage Type: %s\n", status.Storage.StoreType)
        },
    }
}

// System information command
func newSystemCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "system",
        Short: "Get system information",
        Run: func(cmd *cobra.Command, args []string) {
            systemInfo, err := manager.SystemManager().GetAllSystemInfo()
            if err != nil {
                fmt.Fprintf(os.Stderr, "Failed to get system info: %v\n", err)
                return
            }
            fmt.Println("=== System Information ===")
            if tables, ok := systemInfo["tables"].([]interface{}); ok {
                fmt.Printf("Tables: %d\n", len(tables))
                for _, table := range tables {
                    if tableInfo, ok := table.(map[string]interface{}); ok {
                        fmt.Printf("  - %s (ID: %v)\n", tableInfo["Name"], tableInfo["ID"])
                    }
                }
            }
        },
    }
}

// Index management command
func newIndexCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "index",
        Short: "Manage indexes",
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Println("=== Index Management ===")
            systemInfo, err := manager.SystemManager().GetAllSystemInfo()
            if err != nil {
                fmt.Fprintf(os.Stderr, "Failed to get system info: %v\n", err)
                return
            }
            if tableDetails, ok := systemInfo["tableDetails"].(map[uint8]map[string]interface{}); ok {
                for tableID, details := range tableDetails {
                    tableName := ""
                    if name, ok := details["name"].(string); ok {
                        tableName = name
                    }
                    fmt.Printf("Table: %s (ID: %d)\n", tableName, tableID)
                    if indexes, ok := details["indexes"].([]interface{}); ok {
                        for _, index := range indexes {
                            if indexInfo, ok := index.(map[string]interface{}); ok {
                                indexName := ""
                                if name, ok := indexInfo["Name"].(string); ok {
                                    indexName = name
                                }
                                fmt.Printf("  - Index: %s (ID: %v)\n", indexName, indexInfo["ID"])
                            }
                        }
                    }
                }
            }
        },
    }
}

// Configuration management command
func newConfigCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "config",
        Short: "Manage configuration",
        Run: func(cmd *cobra.Command, args []string) {
            configMgr := manager.ConfigManager()
            config, err := configMgr.GetConfig()
            if err != nil {
                fmt.Fprintf(os.Stderr, "Failed to get config: %v\n", err)
                return
            }
            fmt.Println("=== Configuration ===")
            fmt.Printf("Storage Type: %s\n", config.StoreType)
            fmt.Println("Options:")
            for key, value := range config.Options {
                fmt.Printf("  - %s: %s\n", key, value)
            }
        },
    }
}

// Backup management command
func newBackupCmd() *cobra.Command {
    var backupPath string
    var restoreFile string
    var operation string

    cmd := &cobra.Command{
        Use:   "backup",
        Short: "Manage backups",
        Run: func(cmd *cobra.Command, args []string) {
            backupMgr := manager.BackupManager()
            switch operation {
            case "create":
                if backupPath == "" {
                    backupPath = "./backups"
                }
                path, err := backupMgr.Backup(backupPath)
                if err != nil {
                    fmt.Fprintf(os.Stderr, "Failed to create backup: %v\n", err)
                    return
                }
                fmt.Printf("Backup created successfully at: %s\n", path)
            case "restore":
                if restoreFile == "" {
                    fmt.Fprintf(os.Stderr, "Restore file is required\n")
                    return
                }
                err := backupMgr.Restore(restoreFile)
                if err != nil {
                    fmt.Fprintf(os.Stderr, "Failed to restore backup: %v\n", err)
                    return
                }
                fmt.Println("Backup restored successfully")
            default:
                fmt.Println("=== Backup Management ===")
                fmt.Println("Usage: myapp db backup --op=create [--path=./backups]")
                fmt.Println("       myapp db backup --op=restore --file=<backup file>")
            }
        },
    }

    cmd.Flags().StringVar(&operation, "op", "", "Operation: create or restore")
    cmd.Flags().StringVar(&backupPath, "path", "./backups", "Backup path")
    cmd.Flags().StringVar(&restoreFile, "file", "", "Restore file")

    return cmd
}

// Performance statistics command
func newStatsCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "stats",
        Short: "Get performance statistics",
        Run: func(cmd *cobra.Command, args []string) {
            statsMgr := manager.StatsManager()
            queryStats, err := statsMgr.GetQueryStats()
            if err != nil {
                fmt.Fprintf(os.Stderr, "Failed to get stats: %v\n", err)
                return
            }
            fmt.Println("=== Performance Statistics ===")
            fmt.Printf("Query Count: %d\n", queryStats.Count)
            fmt.Printf("Total Time: %d ms\n", queryStats.TotalTime)
            if queryStats.Count > 0 {
                fmt.Printf("Average Time: %.2f ms\n", float64(queryStats.TotalTime)/float64(queryStats.Count))
            }
        },
    }
}

// Monitor command
func newMonitorCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "monitor",
        Short: "Monitor database status",
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Println("=== Database Monitor ===")
            fmt.Println("Starting monitoring... (Press Ctrl+C to stop)")
            
            // Real-time monitoring logic can be implemented here
            // For example:
            // monitor := manager.Monitor(time.Second*5, management.Thresholds{...})
            // for {
            //     status := monitor.Status()
            //     fmt.Printf("Memory: %.2f MB, Queries: %d\n", status.MemoryUsage, status.QueryCount)
            //     time.Sleep(time.Second)
            // }
        },
    }
}
```

**Usage**:

```bash
# Get database status
myapp db status

# Get system information
myapp db system

# Manage indexes
myapp db index

# Get configuration
myapp db config

# Create backup
myapp db backup --op=create --path=./backups

# Restore backup
myapp db backup --op=restore --file=./backups/backup_20260205_123456

# Get performance statistics
myapp db stats

# Monitor database
myapp db monitor
```

**Advantages**:

1. **No conflicts**: Completely avoids conflicts with the system `sfsdb` command
2. **High integration**: Seamlessly integrates with other project commands
3. **Flexible**: Can customize command behavior and output format according to project needs
4. **Good performance**: Directly calls management interfaces, avoiding command-line parsing overhead
5. **Maintainable**: Clear code structure, easy to understand and extend

#### 3. One-click Integration (Recommended)

To simplify the integration process, sfsDb provides a one-click integration function that allows you to integrate all management functions with a single line of code:

```go
package main

import (
    "fmt"
    "os"

    "github.com/liaoran123/sfsDb/management"
    "github.com/spf13/cobra"
)

func main() {
    // Create root command
    rootCmd := &cobra.Command{
        Use:   "yourapp",
        Short: "Your Application",
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Println("Your Application")
        },
    }

    // One-click integration of management commands
    management.AddManagementCommands(rootCmd, func() storage.Store {
        // Return your project's storage instance
        return getDBStore() // Replace with your storage instance获取 method
    })

    // Execute command
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}

// Replace with your project's storage instance获取 method
func getDBStore() storage.Store {
    // Here return your project's storage instance
    return yourProject.GetDatabaseStore()
}
```

**Integration Points**:

1. **One line integration**: Integrate all management functions with a single line via `management.AddManagementCommands`
2. **Storage instance获取**: Provide a function to获取 your project's storage instance
3. **Command customization**: Can specify the name of the management command (default is "db")
4. **Complete functionality**: Automatically integrates all core management commands

**Advantages**:

1. **Simplest**: Complete all integration work with one line of code
2. **Non-intrusive**: Does not modify the project's existing database management logic
3. **Complete functionality**: Provides all core management functions
4. **Highly customizable**: Can customize command name and storage instance获取 method
5. **Easy maintenance**: Automatically get new features when sfsDb is updated

#### 4. Integration Based on Existing Storage Instance (For projects with existing database)

If your project already has a custom database opening method, you can directly use the existing storage instance to integrate management functions:

```go
package main

import (
    "fmt"
    "os"

    "github.com/liaoran123/sfsDb/management"
    "github.com/spf13/cobra"
)

var (
    manager *management.Manager
)

func main() {
    // Create root command
    rootCmd := &cobra.Command{
        Use:   "yourapp",
        Short: "Your Application",
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Println("Your Application")
        },
    }

    // Create database management subcommand
    dbCmd := &cobra.Command{
        Use:   "db",
        Short: "Database management commands",
        PersistentPreRun: func(cmd *cobra.Command, args []string) {
            // Use your project's existing storage instance
            // Here it's assumed your project has a function named getDBStore()
            // This function returns an instance that implements the storage.Store interface
            store := getDBStore() // Replace with your project's database获取 method
            
            // Create manager
            manager = management.NewManager(store)
        },
    }

    // Add specific management commands
    dbCmd.AddCommand(
        newStatusCmd(),
        newStatsCmd(),
        // Other management commands...
    )

    rootCmd.AddCommand(dbCmd)
    rootCmd.Execute()
}

// Status command
func newStatusCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "status",
        Short: "Get database status",
        Run: func(cmd *cobra.Command, args []string) {
            status, err := manager.GetStatus()
            if err != nil {
                fmt.Fprintf(os.Stderr, "Failed to get status: %v\n", err)
                return
            }
            fmt.Println("=== Database Status ===")
            fmt.Printf("Memory Usage: %.2f MB\n", float64(status.Memory.Alloc)/1024/1024)
            fmt.Printf("Storage Type: %s\n", status.Storage.StoreType)
        },
    }
}

// Performance statistics command
func newStatsCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "stats",
        Short: "Get performance statistics",
        Run: func(cmd *cobra.Command, args []string) {
            statsMgr := manager.StatsManager()
            queryStats, err := statsMgr.GetQueryStats()
            if err != nil {
                fmt.Fprintf(os.Stderr, "Failed to get stats: %v\n", err)
                return
            }
            fmt.Println("=== Performance Statistics ===")
            fmt.Printf("Query Count: %d\n", queryStats.Count)
            fmt.Printf("Total Time: %d ms\n", queryStats.TotalTime)
        },
    }
}

// Replace with your project's database获取 method
func getDBStore() storage.Store {
    // Here should return your project's existing storage instance
    // For example:
    // return yourProject.GetDatabaseStore()
    return nil // Example code, need to replace in actual use
}
```

**Integration Points**:

1. **Use existing storage instance**: Directly use the existing storage instance in the project, no need to recreate
2. **Adapter pattern**: If the storage instance interface is not fully compatible, you can create an adapter
3. **Command integration**: Integrate management commands into the project's existing command structure
4. **Feature selection**: Select the management features to integrate according to needs, no need to integrate all

**Advantages**:

1. **Non-intrusive**: Does not modify the project's existing database management logic
2. **High flexibility**: Adapts to various custom database opening methods
3. **High integration**: Seamlessly integrates with the project's existing command structure
4. **Complete functionality**: Provides all sfsDb management features

#### 4. Custom Command Integration

Other projects can also create their own command system based on sfsdb's command structure:

```go
package main

import (
    "fmt"
    "os"

    "github.com/liaoran123/sfsDb/management"
    "github.com/liaoran123/sfsDb/web"
    "github.com/spf13/cobra"
)

var (
    dbPath string
    manager *management.Manager
)

func main() {
    rootCmd := &cobra.Command{
        Use:   "abc",
        Short: "My Application with sfsdb",
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Println("My Application")
            fmt.Println("Use 'abc --help' for more information about available commands.")
        },
    }

    rootCmd.PersistentFlags().StringVar(&dbPath, "db", "./kvdb", "Database path")
    rootCmd.AddCommand(
        newWebCmd(),
        // Add other commands...
    )

    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}

// Custom web command
func newWebCmd() *cobra.Command {
    var webPort string
    
    cmd := &cobra.Command{
        Use:   "web",
        Short: "Manage web interface",
        Run: func(cmd *cobra.Command, args []string) {
            // Initialize storage and manager...
            // Implement web command functionality...
        },
    }
    
    cmd.Flags().BoolP("enable", "e", false, "Enable web interface")
    cmd.Flags().BoolP("disable", "d", false, "Disable web interface")
    cmd.Flags().BoolP("start", "s", false, "Start web server")
    cmd.Flags().StringVarP(&webPort, "port", "p", ":8083", "Web server port")
    
    return cmd
}
```

### Usage

If other projects implement according to the above methods, users can manage sfsdb through the following commands:

```bash
# Enable web interface and configure port
abc.exe web --enable --port :8083

# Disable web interface
abc.exe web --disable

# Start web server
abc.exe web --start

# View current web configuration
abc.exe web
```

## Summary

The sfsDb command-line tool provides comprehensive database management functions, allowing you to complete complex database management tasks with simple commands. Combined with the web interface, users can choose the appropriate management method according to their needs, improving database management efficiency.

Other projects can easily integrate sfsdb's CLI command functionality, generate their own executable files, and maintain consistent command structure and user experience with sfsdb.
