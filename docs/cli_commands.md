# CLI 命令参考

本文档提供了 sfsDb 命令行工具的详细使用指南，包括所有可用命令及其功能。

## 概述

sfsDb 命令行工具提供了以下核心功能：

- **数据库状态管理**：查看数据库当前状态
- **系统信息管理**：获取表、字段、索引等系统信息
- **索引管理**：分析和管理数据库索引
- **配置管理**：查看和修改数据库配置
- **备份恢复**：创建和恢复数据库备份
- **性能统计**：查看查询性能和热点数据
- **监控告警**：实时监控数据库状态
- **Web 界面管理**：启用/禁用和配置 Web 界面

## 基本使用

### 命令格式

```bash
sfsdb [命令] [参数]
```

### 全局参数

| 参数 | 描述 | 默认值 |
|------|------|--------|
| `--db` | 数据库路径 | `./kvdb` |
| `--help` | 显示帮助信息 | - |

## 命令详情

### 1. status - 获取数据库状态

**功能**：获取数据库当前状态，包括内存使用和存储信息。

**使用示例**：

```bash
# 获取默认路径数据库状态
sfsdb status

# 获取指定路径数据库状态
sfsdb status --db ./mydb
```

**输出示例**：

```
=== Database Status ===
Memory Usage: 1.23 MB
Total Allocated: 2.45 MB
System Memory: 5.67 MB
GC Count: 10
Storage Type: LevelDB
```

### 2. system - 系统信息管理

**功能**：获取表、字段、索引等系统信息。

**使用示例**：

```bash
# 获取系统信息
sfsdb system
```

### 3. index - 索引管理

**功能**：分析和管理数据库索引。

**使用示例**：

```bash
# 管理索引
sfsdb index
```

### 4. config - 配置管理

**功能**：查看和修改数据库配置。

**使用示例**：

```bash
# 管理配置
sfsdb config
```

### 5. backup - 备份恢复

**功能**：创建和恢复数据库备份。

**使用示例**：

```bash
# 管理备份
sfsdb backup
```

### 6. stats - 性能统计

**功能**：查看查询性能和热点数据。

**使用示例**：

```bash
# 查看性能统计
sfsdb stats
```

### 7. monitor - 监控告警

**功能**：实时监控数据库状态。

**使用示例**：

```bash
# 启动监控
sfsdb monitor
```

### 8. web - Web 界面管理

**功能**：启用/禁用和配置 Web 界面。

**子命令参数**：

| 参数 | 描述 | 默认值 |
|------|------|--------|
| `--enable` `-e` | 启用 Web 界面 | false |
| `--disable` `-d` | 禁用 Web 界面 | false |
| `--start` `-s` | 启动 Web 服务器 | false |
| `--port` `-p` | Web 服务器端口 | `:8083` |

**使用示例**：

```bash
# 启用 Web 界面并设置端口为 8083
sfsdb web --enable --port :8083

# 禁用 Web 界面
sfsdb web --disable

# 启动 Web 服务器（使用配置的端口）
sfsdb web --start

# 启动 Web 服务器并指定端口（临时覆盖配置）
sfsdb web --start --port :8085

# 查看当前 Web 配置
sfsdb web
```

**输出示例**：

```
Web Interface Configuration:
Enabled: true
Port: :8084
Address: http://localhost:8084
```

## 示例工作流

### 1. 启动 Web 界面

```bash
# 启用 Web 界面并设置端口
sfsdb web --enable --port :8083

# 启动 Web 服务器
sfsdb web --start

# 访问 Web 界面
# http://localhost:8083
```

### 2. 备份数据库

```bash
# 创建数据库备份
sfsdb backup

# 查看数据库状态
sfsdb status
```

### 3. 监控数据库

```bash
# 启动监控
sfsdb monitor

# 查看性能统计
sfsdb stats
```

## 其他项目集成

### 概述

其他项目可以集成 sfsdb 的 CLI 命令功能，生成自己的可执行文件（如 abc.exe），并使用相同的命令结构来管理数据库。

### 最简实现

如果用户不希望使用命令行工具，也可以通过代码直接启用和配置 Web 界面。以下是完整的最简实现：

```go
package main

import (
    "github.com/liaoran123/sfsDb/storage"
    "github.com/liaoran123/sfsDb/web"
    "github.com/liaoran123/sfsDb/management"
)

func main() {
    // 初始化数据库
    store, _ := storage.OpenDefaultDb("./kvdb")
    defer storage.CloseDb()
    
    // 创建管理器
    manager := management.NewManager(store)
    
    // 核心配置（只需这两行）
    configMgr := manager.ConfigManager()
    configMgr.SetConfig("web_enable", "true")
    configMgr.SetConfig("web_port", ":8083")
    
    // 启动Web服务器
    server := web.NewServer(":8083", manager)
    server.Start()
}
```

### 核心原理

启用和配置 Web 界面的核心只需设置两个参数：

1. **`web_enable`**：控制 Web 界面是否启用（值为 `"true"` 或 `"false"`）
2. **`web_port`**：配置 Web 服务器的端口（值为端口字符串，如 `":8083"`）

这两个参数会被持久化保存到数据库中，下次启动时自动生效。

### 实现方式



#### 2. 使用子命令集成（推荐）

为了避免命令冲突和程序重启的问题，推荐使用子命令的方式集成 sfsdb 的管理功能：

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
    // 创建根命令
    rootCmd := &cobra.Command{
        Use:   "myapp",
        Short: "My Application",
        Long:  "My Application with integrated database management",
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Println("My Application")
            fmt.Println("Use 'myapp --help' for more information about available commands.")
        },
    }

    // 添加全局数据库路径参数
    rootCmd.PersistentFlags().StringVar(&dbPath, "db", "./kvdb", "Database path")

    // 创建数据库管理子命令
    dbCmd := &cobra.Command{
        Use:   "db",
        Short: "Database management commands",
        Long:  "Commands for managing the embedded database",
        PersistentPreRun: func(cmd *cobra.Command, args []string) {
            // 初始化数据库和管理器（只执行一次）
            store, err := storage.NewLevelDBStore(dbPath, nil)
            if err != nil {
                fmt.Fprintf(os.Stderr, "Failed to open database: %v\n", err)
                os.Exit(1)
            }
            manager = management.NewManager(store)
        },
    }

    // 添加具体的数据库管理子命令
    dbCmd.AddCommand(
        newStatusCmd(),
        newSystemCmd(),
        newIndexCmd(),
        newConfigCmd(),
        newBackupCmd(),
        newStatsCmd(),
        newMonitorCmd(),
    )

    // 将数据库管理子命令添加到根命令
    rootCmd.AddCommand(dbCmd)

    // 执行命令
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}

// 状态命令
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

// 系统信息命令
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

// 索引管理命令
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

// 配置管理命令
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

// 备份管理命令
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

// 性能统计命令
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

// 监控命令
func newMonitorCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "monitor",
        Short: "Monitor database status",
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Println("=== Database Monitor ===")
            fmt.Println("Starting monitoring... (Press Ctrl+C to stop)")
            
            // 这里可以实现实时监控逻辑
            // 例如：
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

**使用方法**：

```bash
# 查看数据库状态
myapp db status

# 查看系统信息
myapp db system

# 管理索引
myapp db index

# 查看配置
myapp db config

# 创建备份
myapp db backup --op=create --path=./backups

# 恢复备份
myapp db backup --op=restore --file=./backups/backup_20260205_123456

# 查看性能统计
myapp db stats

# 监控数据库
myapp db monitor
```

**优势**：

1. **无冲突**：完全避免与系统 `sfsdb` 命令的冲突
2. **集成度高**：与项目的其他命令无缝集成
3. **灵活性强**：可以根据项目需求定制命令行为和输出格式
4. **性能好**：直接调用管理接口，避免命令行解析的开销
5. **可维护性强**：代码结构清晰，易于理解和扩展

#### 3. 一键集成（推荐）

为了简化集成过程，sfsDb 提供了一键集成函数，允许您通过一行代码集成所有管理功能：

```go
package main

import (
    "fmt"
    "os"

    "github.com/liaoran123/sfsDb/management"
    "github.com/spf13/cobra"
)

func main() {
    // 创建根命令
    rootCmd := &cobra.Command{
        Use:   "yourapp",
        Short: "Your Application",
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Println("Your Application")
        },
    }

    // 一键集成管理命令
    management.AddManagementCommands(rootCmd, func() storage.Store {
        // 返回您项目的存储实例
        return getDBStore() // 替换为您的存储实例获取方法
    })

    // 执行命令
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}

// 替换为您项目的存储实例获取方法
func getDBStore() storage.Store {
    // 这里返回您项目的存储实例
    return yourProject.GetDatabaseStore()
}
```

**集成要点**：

1. **一行代码集成**：通过 `management.AddManagementCommands` 函数一键集成所有管理功能
2. **存储实例获取**：提供一个函数来获取您项目的存储实例
3. **命令自定义**：可以指定管理命令的名称（默认为 "db"）
4. **功能完整**：自动集成所有核心管理命令

**优势**：

1. **最简单**：一行代码完成所有集成工作
2. **无侵入性**：不修改项目现有的数据库管理逻辑
3. **功能完整**：提供所有核心管理功能
4. **高度定制**：可以自定义命令名称和存储实例获取方法
5. **维护便捷**：后续 sfsDb 更新时自动获得新功能

#### 4. 基于现有存储实例的集成（适合已有数据库的项目）

如果您的项目已经有了自定义的数据库打开方式，可以直接使用现有的存储实例集成管理功能：

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
    // 创建根命令
    rootCmd := &cobra.Command{
        Use:   "yourapp",
        Short: "Your Application",
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Println("Your Application")
        },
    }

    // 创建数据库管理子命令
    dbCmd := &cobra.Command{
        Use:   "db",
        Short: "Database management commands",
        PersistentPreRun: func(cmd *cobra.Command, args []string) {
            // 使用您项目现有的存储实例
            // 这里假设您的项目已经有了一个名为 getDBStore() 的函数
            // 该函数返回一个实现了 storage.Store 接口的实例
            store := getDBStore() // 替换为您项目的数据库获取方法
            
            // 创建管理器
            manager = management.NewManager(store)
        },
    }

    // 添加具体的管理命令
    dbCmd.AddCommand(
        newStatusCmd(),
        newStatsCmd(),
        // 其他管理命令...
    )

    rootCmd.AddCommand(dbCmd)
    rootCmd.Execute()
}

// 状态命令
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

// 性能统计命令
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

// 替换为您项目的数据库获取方法
func getDBStore() storage.Store {
    // 这里应该返回您项目现有的存储实例
    // 例如：
    // return yourProject.GetDatabaseStore()
    return nil // 示例代码，实际使用时需要替换
}
```

**集成要点**：

1. **使用现有存储实例**：直接使用项目中已有的存储实例，不需要重新创建
2. **适配器模式**：如果存储实例接口不完全兼容，可以创建适配器
3. **命令集成**：将管理命令集成到项目现有的命令结构中
4. **功能选择**：根据需要选择集成的管理功能，不需要全部集成

**优势**：

1. **无侵入性**：不修改项目现有的数据库管理逻辑
2. **灵活性高**：适应各种自定义的数据库打开方式
3. **集成度高**：与项目的现有命令结构无缝集成
4. **功能完整**：提供所有 sfsDb 的管理功能

#### 4. 自定义命令集成

其他项目也可以基于 sfsdb 的命令结构，创建自己的命令体系：

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
        // 添加其他命令...
    )

    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}

// 自定义 web 命令
func newWebCmd() *cobra.Command {
    var webPort string
    
    cmd := &cobra.Command{
        Use:   "web",
        Short: "Manage web interface",
        Run: func(cmd *cobra.Command, args []string) {
            // 初始化存储和管理器...
            // 实现 web 命令功能...
        },
    }
    
    cmd.Flags().BoolP("enable", "e", false, "Enable web interface")
    cmd.Flags().BoolP("disable", "d", false, "Disable web interface")
    cmd.Flags().BoolP("start", "s", false, "Start web server")
    cmd.Flags().StringVarP(&webPort, "port", "p", ":8083", "Web server port")
    
    return cmd
}
```

### 使用方式

如果其他项目按照上述方式实现，用户就可以通过以下命令来管理 sfsdb：

```bash
# 启用 web 界面并配置端口
abc.exe web --enable --port :8083

# 禁用 web 界面
abc.exe web --disable

# 启动 web 服务器
abc.exe web --start

# 查看当前 web 配置
abc.exe web
```

## 总结

sfsDb 命令行工具提供了全面的数据库管理功能，通过简单的命令即可完成复杂的数据库管理任务。结合 Web 界面，用户可以根据自己的需求选择合适的管理方式，提高数据库管理效率。

其他项目可以轻松集成 sfsdb 的 CLI 命令功能，生成自己的可执行文件，并保持与 sfsdb 一致的命令结构和使用体验。
