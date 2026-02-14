# sfsDb 数据库使用文档（中文版）

## 项目简介

sfsDb 是一个灵活、高效的嵌入式数据库，支持多种数据类型、索引类型和查询方式。它提供了简洁的 API 设计，易于集成到各种应用场景中。

## 目录

### 基础功能
- [数据库初始化](./basic/initialization.md)
- [创建表与设置字段](./basic/table_creation.md)
- [插入数据](./basic/data_insertion.md)
- [查询数据](./basic/data_query.md)
- [修改记录](./basic/data_update.md)
- [删除记录](./basic/data_deletion.md)

### 高级功能
- [主键管理](./advanced/primary_key.md)
- [索引管理](./advanced/index_management.md)
- [全文搜索](./advanced/full_text_search.md)
- [字段修改](./advanced/field_modification.md)
- [事务管理](./advanced/transaction.md)
- [管理工具库](./advanced/management.md)
- [跳跃区间](./advanced/jump_ranges.md)
- [时序数据处理](./advanced/time_series.md)
- [其他功能](./advanced/other_features.md)

### 优化与最佳实践
- [对象池与内存管理](./optimization/object_pool.md)
- [半结构化数据支持](./optimization/semi_structured.md)
- [最佳实践](./optimization/best_practices.md)
- [常见问题](./optimization/common_issues.md)
- [总结](./optimization/summary.md)

## 主要特性

1. **灵活的数据模型**：支持多种字段类型和动态字段
2. **强大的索引系统**：支持单主键、复合主键、普通索引和全文索引
3. **丰富的查询功能**：支持比较操作符、自定义匹配器和全文搜索
4. **易于使用的 API**：简洁的 API 设计，易于集成到各种应用场景
5. **高效的性能**：优化的存储结构和查询算法
6. **事务支持**：确保数据操作的原子性和一致性
7. **对象池机制**：优化内存使用和性能
8. **半结构化数据支持**：灵活处理复杂数据结构
9. **管理工具库**：提供数据库监控、配置管理、备份恢复和性能分析功能
10. **时序数据处理**：内置 time 包，支持时间粒度处理、时间窗口计算、数据聚合和时间戳转换等时序数据相关功能

## 快速开始

### 1. 安装

```bash
go get github.com/liaoran123/sfsDb
```

### 2. 基本使用

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
)

func main() {
    // 创建表
    table, err := engine.TableNew("users")
    if err != nil {
        panic(err)
    }
    
    // 设置字段
    fields := map[string]any{
        "id":   0,     // 自动增值主键
        "name": "",    // 字符串类型
        "age":  0,     // 整数类型
    }
    err = table.SetFields(fields)
    if err != nil {
        panic(err)
    }
    
    // 插入数据
    user := map[string]any{
        "name": "Alice",
        "age":  30,
    }
    id, err := table.Insert(&user)
    if err != nil {
        panic(err)
    }
    fmt.Printf("插入成功，ID: %v\n", id)
    
    // 查询数据
    searchData := map[string]any{"name": "Alice"}
    iter := table.Search(&searchData)
    if iter != nil {
        defer engine.GlobalTableIterPool.Put(iter)
    }
    
    records := iter.GetRecords(true)
    if records != nil {
        defer engine.PutRecords(records)
    }
    
    for _, record := range records {
        fmt.Printf("记录: %v\n", record)
    }
}
```

## 性能测试报告

- [时序数据库性能比较报告](../performance/time_series_benchmark.md) - time包基准测试与其他时序数据库性能比较

## 文档维护

本文档采用模块化结构组织，便于维护和更新。如有任何问题或建议，请参考 [文档维护指南](../DOCUMENTATION_MAINTENANCE.md)。

## 贡献指南

欢迎贡献代码和文档！请参考 [CONTRIBUTING.md](https://github.com/liaoran123/sfsDb/blob/main/CONTRIBUTING.md) 了解如何参与项目。

## 许可证

sfsDb 使用 MIT 许可证。详见 [LICENSE](https://github.com/liaoran123/sfsDb/blob/main/LICENSE) 文件。