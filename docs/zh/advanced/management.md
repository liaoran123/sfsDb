# 管理工具库

管理工具库是sfsDb的核心组件之一，提供了全面的数据库管理、监控和优化功能。本文档将详细介绍管理工具库的使用方法和功能特性。

## 概述

管理工具库位于`management`包中，包含以下核心模块：

- **系统信息管理**：获取表、字段、索引等系统信息
- **索引管理**：分析索引使用情况，提供优化建议
- **性能统计**：记录查询性能，识别热点数据
- **配置管理**：管理数据库配置，提供优化建议
- **备份恢复**：备份和恢复数据库，无需关闭数据库
- **监控告警**：实时监控数据库状态，设置阈值告警
- **状态管理**：获取数据库内存使用和存储状态

## 核心功能

### 1. 深度集成

管理工具库支持与`engine.Table`实例的深度集成，可以获取更详细的表和索引信息。

```go
// 创建带表实例的管理器
manager := management.NewManagerWithTable(store, table)

// 使用深度集成功能获取索引信息
indexMgr := manager.IndexManager()
indexes, err := indexMgr.ListIndexes("test_table")
```

### 2. 监控告警系统

监控告警系统可以实时监控数据库状态，并在达到阈值时触发告警。

```go
// 创建监控器
monitor := manager.Monitor(time.Second*5, management.Thresholds{
    MemoryUsage: 1024, // 1GB
    GCCount:     100,
})

// 启动监控
if err := monitor.Start(); err != nil {
    fmt.Printf("启动监控失败: %v\n", err)
}

// 停止监控
monitor.Stop()
```

### 3. 配置管理

配置管理模块允许运行时调整配置，并提供优化建议。

```go
// 获取配置管理器
configMgr := manager.ConfigManager()

// 获取当前配置
config, err := configMgr.GetConfig()

// 设置配置
if err := configMgr.SetConfig("write_buffer", "128MB"); err != nil {
    fmt.Printf("设置配置失败: %v\n", err)
}

// 获取优化建议
suggestions, err := configMgr.GetOptimizationSuggestions()
```

### 4. 备份恢复

备份恢复模块支持在不关闭数据库的情况下进行备份和恢复操作。

```go
// 获取备份管理器
backupMgr := manager.BackupManager()

// 创建备份
backupPath, err := backupMgr.Backup("./backups")

// 恢复数据库
if err := backupMgr.Restore(backupPath); err != nil {
    fmt.Printf("恢复数据库失败: %v\n", err)
}
```

### 5. 性能统计

性能统计模块记录查询性能，并识别热点数据。

```go
// 获取性能统计管理器
statsMgr := manager.StatsManager()

// 记录查询性能
statsMgr.RecordQuery(time.Millisecond*10, "select")

// 记录数据访问
statsMgr.RecordAccess("key_1")

// 获取查询统计
queryStats, err := statsMgr.GetQueryStats()

// 获取热点数据
hotspots, err := statsMgr.GetHotspots(5)

// 启动性能分析
if err := statsMgr.StartProfiling(); err != nil {
    fmt.Printf("启动性能分析失败: %v\n", err)
}

// 停止性能分析并获取结果
profilingResult, err := statsMgr.StopProfiling()
```

### 6. 系统信息管理

系统信息管理模块提供表、字段、索引等系统信息。

```go
// 获取系统信息管理器
systemMgr := manager.SystemManager()

// 获取所有表信息
tables, err := systemMgr.GetAllTables()

// 获取指定表的字段信息
fields, err := systemMgr.GetTableFields(tableID)

// 获取指定表的索引信息
indexes, err := systemMgr.GetTableIndexes(tableID)

// 获取所有系统信息
systemInfo, err := systemMgr.GetAllSystemInfo()
```

### 7. 状态管理

状态管理模块提供数据库内存使用和存储状态信息。

```go
// 获取数据库状态
statusInfo, err := manager.GetStatus()
fmt.Printf("内存使用: %.2f MB\n", float64(statusInfo.Memory.Alloc)/1024/1024)
fmt.Printf("GC 次数: %d\n", statusInfo.Memory.NumGC)
fmt.Printf("存储类型: %s\n", statusInfo.Storage.StoreType)
```

### 8. Web界面管理

Web界面管理模块允许用户启用/禁用web界面并配置其端口。通过CLI命令可以方便地管理web界面设置。

#### 启用Web界面并配置端口

```bash
# 启用web界面并设置端口为8083
sfsdb web --enable --port :8083

# 启用web界面并设置端口为8084
sfsdb web --enable --port :8084
```

#### 禁用Web界面

```bash
# 禁用web界面
sfsdb web --disable
```

#### 启动Web服务器

```bash
# 启动web服务器（使用配置的端口）
sfsdb web --start

# 启动web服务器并指定端口（临时覆盖配置）
sfsdb web --start --port :8085
```

#### 查看当前Web配置

```bash
# 查看当前web配置
sfsdb web
```

输出示例：

```
Web Interface Configuration:
Enabled: true
Port: :8084
Address: http://localhost:8084
```

#### Web界面功能

Web界面提供以下功能：

- **系统信息**：查看数据库和系统状态
- **配置管理**：查看和修改数据库配置
- **备份恢复**：创建和恢复数据库备份
- **性能统计**：查看查询性能和热点数据
- **监控告警**：实时监控数据库状态
- **数据操作**：基本的数据CRUD操作

## 使用示例

### 完整示例

以下是一个完整的管理工具库使用示例：

```go
package main

import (
    "fmt"
    "time"

    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/management"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // 初始化数据库
    _, err := storage.OpenDefaultDb("./kvdb")
    if err != nil {
        fmt.Printf("打开数据库失败: %v\n", err)
        return
    }
    defer storage.CloseDb()

    // 创建测试表
    table, err := engine.TableNew("test_table")
    if err != nil {
        fmt.Printf("创建表失败: %v\n", err)
        return
    }

    // 设置表字段
    fields := map[string]any{
        "id":   0,
        "name": "",
        "age":  0,
    }
    if err := table.SetFields(fields); err != nil {
        fmt.Printf("设置字段失败: %v\n", err)
        return
    }

    // 创建索引
    primaryKey, err := engine.DefaultPrimaryKeyNew("primary_key")
    if err != nil {
        fmt.Printf("创建主键失败: %v\n", err)
        return
    }
    primaryKey.AddFields("id")
    if err := table.CreateIndex(primaryKey); err != nil {
        fmt.Printf("添加主键失败: %v\n", err)
        return
    }

    normalIndex, err := engine.DefaultNormalIndexNew("name_index")
    if err != nil {
        fmt.Printf("创建普通索引失败: %v\n", err)
        return
    }
    normalIndex.AddFields("name")
    if err := table.CreateIndex(normalIndex); err != nil {
        fmt.Printf("添加普通索引失败: %v\n", err)
        return
    }

    // 创建管理器
    manager := management.NewManagerWithTable(storage.KVDb, table)

    // 测试系统信息管理器
    fmt.Println("=== 测试系统信息管理器 ===")
    systemMgr := manager.SystemManager()
    tables, err := systemMgr.GetAllTables()
    if err != nil {
        fmt.Printf("获取表信息失败: %v\n", err)
        return
    }
    fmt.Printf("所有表: %+v\n", tables)

    // 测试索引管理器
    fmt.Println("\n=== 测试索引管理器 ===")
    indexMgr := manager.IndexManager()
    indexes, err := indexMgr.ListIndexes("test_table")
    if err != nil {
        fmt.Printf("列出索引失败: %v\n", err)
        return
    }
    fmt.Printf("所有索引: %+v\n", indexes)

    // 测试性能统计管理器
    fmt.Println("\n=== 测试性能统计管理器 ===")
    statsMgr := manager.StatsManager()

    // 模拟查询
    for i := 0; i < 10; i++ {
        statsMgr.RecordQuery(time.Millisecond*10, "select")
        statsMgr.RecordAccess(fmt.Sprintf("key_%d", i))
    }

    queryStats, err := statsMgr.GetQueryStats()
    if err != nil {
        fmt.Printf("获取查询统计失败: %v\n", err)
        return
    }
    fmt.Printf("查询统计: %+v\n", queryStats)

    // 测试配置管理器
    fmt.Println("\n=== 测试配置管理器 ===")
    configMgr := manager.ConfigManager()
    config, err := configMgr.GetConfig()
    if err != nil {
        fmt.Printf("获取配置失败: %v\n", err)
        return
    }
    fmt.Printf("当前配置: %+v\n", config)

    // 测试备份管理器
    fmt.Println("\n=== 测试备份管理器 ===")
    backupMgr := manager.BackupManager()

    // 创建备份
    backupPath, err := backupMgr.Backup("./backups")
    if err != nil {
        fmt.Printf("创建备份失败: %v\n", err)
        return
    }
    fmt.Printf("创建备份成功: %s\n", backupPath)

    // 测试监控管理器
    fmt.Println("\n=== 测试监控管理器 ===")
    monitorMgr := manager.MonitorManager()
    keyChangeStats := monitorMgr.GetKeyChangeStats()
    fmt.Printf("键值变化统计: %+v\n", keyChangeStats)

    // 测试状态管理器
    fmt.Println("\n=== 测试状态管理器 ===")
    statusInfo, err := manager.GetStatus()
    if err != nil {
        fmt.Printf("获取状态失败: %v\n", err)
        return
    }
    fmt.Printf("数据库状态: %+v\n", statusInfo)

    // 测试监控器
    fmt.Println("\n=== 测试监控器 ===")
    monitor := manager.Monitor(time.Second*5, management.Thresholds{
        MemoryUsage: 1024, // 1GB
        GCCount:     100,
    })

    err = monitor.Start()
    if err != nil {
        fmt.Printf("启动监控失败: %v\n", err)
        return
    }
    fmt.Println("监控已启动")

    // 等待一段时间
    time.Sleep(time.Second * 10)

    // 停止监控
    monitor.Stop()
    fmt.Println("监控已停止")

    fmt.Println("\n=== 所有管理功能测试完成 ===")
}
```

## 性能优化建议

使用管理工具库时，以下是一些性能优化建议：

1. **定期分析索引使用情况**：使用`IndexManager.AnalyzeIndexes`定期分析索引使用情况，删除未使用的索引。

2. **监控数据库状态**：使用监控器定期监控数据库状态，及时发现性能问题。

3. **优化配置参数**：根据应用场景调整配置参数，如`write_buffer`、`max_open_files`等。

4. **备份策略**：制定合理的备份策略，定期备份数据库，确保数据安全。

5. **热点数据处理**：识别热点数据，考虑使用缓存或其他优化策略。

6. **使用异步操作进行监控**：监控系统现在支持异步操作，以获得更好的性能，特别是在高并发场景下。

### 监控中的异步操作

监控系统已优化为使用异步操作进行计数器更新和索引时间记录。这减少了主协程阻塞，提高了系统吞吐量。

#### 主要优势

- **减少主协程阻塞**：主协程可以立即返回，无需等待监控操作
- **提高并发能力**：多个监控任务可以在线程池中并发处理
- **更好的资源利用**：线程池更有效地管理资源
- **更平滑的峰值处理**：异步操作可以更优雅地处理请求峰值

#### 实现细节

监控系统使用全局线程池处理异步任务：

```go
// 全局线程池初始化
var globalPool *Pool

func init() {
    // 创建线程池，大小为 CPU 核心数 * 4
    globalPool = NewPoolWithSize(runtime.NumCPU() * 4)
}

// 异步计数器更新
func (m *KeysMap) IncAsync(key int, tbId uint8, indxName string) {
    globalPool.Submit(func() {
        m.Inc(key, tbId, indxName)
    })
}

// 异步索引时间记录
func (i *IndexStatsMap) SettimeAsync(indexKey int, duration time.Duration, tblName string, indxName string) {
    globalPool.Submit(func() {
        i.Settime(indexKey, duration, tblName, indxName)
    })
}
```

#### 使用方法

系统会自动使用异步操作，因此用户不需要进行任何代码更改。以下操作现在使用异步处理：

1. **键值计数器更新**（通过`monitor.KeyInc`和`monitor.KeyDec`）
2. **索引时间记录**（通过`monitor.GIndexStatsMap.SettimeAsync`）

## 总结

管理工具库为sfsDb提供了全面的管理、监控和优化功能，是数据库运维的重要工具。通过合理使用管理工具库，可以提高数据库性能，确保数据安全，简化数据库管理工作。
