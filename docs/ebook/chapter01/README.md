# 第 1 章：概述与快速开始

欢迎来到 sfsDb 的世界！在这一章中，我们将了解 sfsDb 是什么、它的核心特性，以及如何快速搭建开发环境并运行第一个示例程序。

## 1.1 sfsDb 是什么

sfsDb 是一款专为**工业物联网（IIoT）和边缘计算**场景设计的通用嵌入式数据库。它具有以下核心特点：

### 核心定位

- **纯 Go 语言开发**：无 CGO 依赖，零配置，无需外部依赖库
- **极致轻量**：启动占用内存仅数 MB，单次操作内存分配约 1KB
- **单文件部署**：编译后为单一可执行文件，可轻松集成到 Docker 容器
- **工业级可靠**：针对工业物联网场景优化，具备断电保护能力
- **广泛硬件兼容**：支持 ARM（ARM64/ARMv7）和 x86 架构

### 技术优势

1. **轻量级设计与多模能力融合**
   - 创新性地将 NoSQL 的高并发写入能力与 SQL 的复杂查询能力融合
   - 避免了传统解决方案中分别部署 NoSQL 和 SQL 数据库的问题

2. **原生支持考据级全文索引**
   - 内置高性能全文索引引擎
   - 提供精准的文本搜索能力

3. **基于 LevelDB 封装实现**
   - 使用 goleveldb 作为存储引擎基础
   - 充分利用 LSM-Tree 架构优势

4. **无锁事务系统**
   - 采用乐观并发控制（OCC）机制
   - 在 10 并发下达到 26,315 ops/s

## 1.2 为什么选择 sfsDb

### 边缘计算的数据困境

在 5G 和工业 4.0 的推动下，数据正从中心云向边缘侧转移。然而，边缘环境具有独特的约束：

- **资源受限**：ARM 架构的边缘网关通常仅有数百 MB 内存
- **弱网环境**：网络带宽低且不稳定
- **异构硬件**：需要同时支持 ARM 和 x86 等多种芯片架构

### 与其他方案对比

| 特性 | sfsDb | SQLite | BoltDB | BadgerDB |
|------|-------|--------|--------|----------|
| 语言 | Go | C | Go | Go |
| CGO 依赖 | 无 | 有 | 无 | 无 |
| 内存占用 | 极低 | 低 | 中 | 中 |
| 并发写入 | 优秀 | 一般 | 良好 | 良好 |
| 事务支持 | ✅ | ✅ | ✅ | ✅ |
| 全文索引 | ✅ | ❌ | ❌ | ❌ |
| 时序优化 | ✅ | ❌ | ❌ | ❌ |
| 加密存储 | ✅ | 需扩展 | ❌ | ❌ |

## 1.3 环境搭建

### 系统要求

- Go 1.25 或更高版本
- 支持的操作系统：
  - Linux（推荐）
  - macOS
  - Windows
- 支持的架构：
  - x86_64
  - ARM64
  - ARMv7

### 安装 sfsDb

使用 Go 模块安装 sfsDb：

```bash
go get github.com/liaoran123/sfsDb
```

### 验证安装

创建一个简单的测试文件来验证安装：

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    fmt.Println("sfsDb 安装成功！")
    fmt.Println("版本信息：v1.0")
}
```

运行测试：

```bash
go run test_install.go
```

## 1.4 Hello World 示例

让我们创建第一个完整的 sfsDb 应用程序！

### 完整示例代码

创建 `hello_world.go`：

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    fmt.Println("=== sfsDb Hello World 示例 ===")

    // 1. 初始化数据库
    fmt.Println("\n1. 初始化数据库...")
    dbManager := storage.GetDBManager()
    _, err := dbManager.OpenDB("./hello_world_db")
    if err != nil {
        panic(fmt.Sprintf("打开数据库失败: %v", err))
    }
    defer dbManager.CloseDB()
    defer storage.RemoveDir("./hello_world_db")
    fmt.Println("✓ 数据库初始化成功")

    // 2. 创建表
    fmt.Println("\n2. 创建用户表...")
    userTable, err := engine.TableNew("users")
    if err != nil {
        panic(fmt.Sprintf("创建表失败: %v", err))
    }
    fmt.Println("✓ 表创建成功")

    // 3. 设置字段
    fmt.Println("\n3. 设置表字段...")
    userFields := map[string]any{
        "id":   0,  // 自动增值主键
        "name": "", // 用户名
        "age":  0,  // 年龄
    }
    err = userTable.SetFields(userFields)
    if err != nil {
        panic(fmt.Sprintf("设置字段失败: %v", err))
    }
    fmt.Println("✓ 字段设置成功")

    // 4. 创建主键索引
    fmt.Println("\n4. 创建主键索引...")
    primaryKey, err := engine.DefaultPrimaryKeyNew("id")
    if err != nil {
        panic(fmt.Sprintf("创建主键索引失败: %v", err))
    }
    primaryKey.AddFields("id")
    err = userTable.CreateIndex(primaryKey)
    if err != nil {
        panic(fmt.Sprintf("创建索引失败: %v", err))
    }
    fmt.Println("✓ 主键索引创建成功")

    // 5. 插入数据
    fmt.Println("\n5. 插入测试数据...")
    users := []map[string]any{
        {"id": 1, "name": "张三", "age": 25},
        {"id": 2, "name": "李四", "age": 30},
        {"id": 3, "name": "王五", "age": 35},
    }

    for _, user := range users {
        currentID, err := userTable.Insert(&user)
        if err != nil {
            panic(fmt.Sprintf("插入数据失败: %v", err))
        }
        fmt.Printf("  ✓ 插入用户成功, ID: %d\n", currentID)
    }

    // 6. 查询数据
    fmt.Println("\n6. 查询数据...")
    {
        iter, err := userTable.Search(&map[string]any{"id": 1})
        defer iter.Release()
        if err != nil {
            panic(fmt.Sprintf("搜索失败: %v", err))
        }
        records := iter.GetRecords(true)
        defer records.Release()

        if len(records) > 0 {
            fmt.Printf("  ✓ 查询结果: %v\n", records[0])
        }
    }

    // 7. 查询所有数据
    fmt.Println("\n7. 查询所有数据...")
    {
        allIter, err := userTable.Search(&map[string]any{})
        defer allIter.Release()
        if err != nil {
            panic(fmt.Sprintf("搜索失败: %v", err))
        }
        allRecords := allIter.GetRecords(true)
        defer allRecords.Release()

        fmt.Printf("  ✓ 当前表中共有 %d 条记录\n", len(allRecords))
        for i, r := range allRecords {
            fmt.Printf("  记录 %d: %v\n", i+1, r)
        }
    }

    fmt.Println("\n=== 示例运行完成！===")
}
```

### 运行示例

```bash
go run hello_world.go
```

### 预期输出

```
=== sfsDb Hello World 示例 ===

1. 初始化数据库...
✓ 数据库初始化成功

2. 创建用户表...
✓ 表创建成功

3. 设置表字段...
✓ 字段设置成功

4. 创建主键索引...
✓ 主键索引创建成功

5. 插入测试数据...
  ✓ 插入用户成功, ID: 1
  ✓ 插入用户成功, ID: 2
  ✓ 插入用户成功, ID: 3

6. 查询数据...
  ✓ 查询结果: map[age:25 id:1 name:张三]

7. 查询所有数据...
  ✓ 当前表中共有 3 条记录
  记录 1: map[age:25 id:1 name:张三]
  记录 2: map[age:30 id:2 name:李四]
  记录 3: map[age:35 id:3 name:王五]

=== 示例运行完成！===
```

## 1.5 代码解析

让我们逐段解析这个示例：

### 1.5.1 导入依赖

```go
import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"    // 核心引擎
    "github.com/liaoran123/sfsDb/storage"   // 存储管理
)
```

- `engine` 包：提供表操作、索引、查询等核心功能
- `storage` 包：提供数据库管理、存储配置等功能

### 1.5.2 初始化数据库

```go
dbManager := storage.GetDBManager()
_, err := dbManager.OpenDB("./hello_world_db")
defer dbManager.CloseDB()
defer storage.RemoveDir("./hello_world_db")
```

- `GetDBManager()`：获取数据库管理器单例
- `OpenDB()`：打开或创建数据库
- `defer`：确保程序退出时清理资源

### 1.5.3 创建表

```go
userTable, err := engine.TableNew("users")
```

- `TableNew()`：创建一个新表
- 表名必须唯一

### 1.5.4 设置字段

```go
userFields := map[string]any{
    "id":   0,  // 整数类型
    "name": "", // 字符串类型
    "age":  0,  // 整数类型
}
err = userTable.SetFields(userFields)
```

- 使用 `map[string]any` 定义字段
- 字段类型通过默认值推断
- 支持多种数据类型

### 1.5.5 插入数据

```go
user := map[string]any{
    "id": 1, "name": "张三", "age": 25,
}
currentID, err := userTable.Insert(&user)
```

- `Insert()`：插入一条记录
- 返回自动生成的 ID（如果有）

### 1.5.6 查询数据

```go
iter, err := userTable.Search(&map[string]any{"id": 1})
defer iter.Release()
records := iter.GetRecords(true)
defer records.Release()
```

- `Search()`：执行查询
- `GetRecords()`：获取查询结果
- 记得使用 `defer` 释放资源

## 1.6 常见问题

### Q: 如何选择数据库存储路径？

A: 建议使用相对路径或绝对路径，确保应用有读写权限。对于生产环境，建议使用独立的数据目录。

### Q: sfsDb 支持并发访问吗？

A: 是的！sfsDb 设计为并发安全的，可以在多个 goroutine 中同时使用。

### Q: 数据持久化如何保证？

A: sfsDb 基于 LevelDB，所有写入操作都会持久化到磁盘。

## 1.7 本章小结

在这一章中，我们：
- ✅ 了解了 sfsDb 的核心特性和优势
- ✅ 完成了环境搭建
- ✅ 运行了第一个 Hello World 示例
- ✅ 学习了基础的 CRUD 操作

在下一章中，我们将深入了解 sfsDb 的核心概念和架构设计！

---

**💡 练习题**：
1. 修改示例代码，添加更多字段（如 email、address）
2. 尝试更新和删除数据
3. 探索使用不同的查询条件

**📝 参考资料**：
- [sfsDb README](https://github.com/liaoran123/sfsDb)
- [Go 语言官方文档](https://golang.org/doc/)
