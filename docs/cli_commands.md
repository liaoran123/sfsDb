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

#### 1. 直接集成 sfsdb 的 CLI 命令

其他项目可以直接使用 sfsdb 的 `cmd` 包，快速实现相同的命令行功能：

```go
package main

import (
    "fmt"
    "os"

    "github.com/liaoran123/sfsDb/cmd/sfsdb/cmd"
)

func main() {
    // 获取 sfsdb 的根命令
    rootCmd := cmd.NewRootCmd()
    
    // 可以添加自己的自定义命令
    // rootCmd.AddCommand(...)
    
    // 执行命令
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

#### 2. 自定义命令集成

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
