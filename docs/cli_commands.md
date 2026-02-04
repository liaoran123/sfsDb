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

## 总结

sfsDb 命令行工具提供了全面的数据库管理功能，通过简单的命令即可完成复杂的数据库管理任务。结合 Web 界面，用户可以根据自己的需求选择合适的管理方式，提高数据库管理效率。
