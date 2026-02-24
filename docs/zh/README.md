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
- [服务支持](./advanced/service_support.md)
- [范围搜索](./advanced/search_range.md)
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

## 隐私政策

### 数据收集与使用

1. **数据收集**：
   - sfsDb 作为嵌入式数据库，所有数据均存储在用户本地环境中，不会自动收集或传输任何用户数据到外部服务器。
   - 当用户选择使用我们的服务支持时，可能需要提供必要的联系信息和问题描述，这些信息仅用于解决用户问题。

2. **数据使用**：
   - 本地存储的数据完全由用户控制，sfsDb 不会访问或使用这些数据。
   - 服务支持过程中收集的信息仅用于提供技术支持和改进服务质量。

3. **数据保护**：
   - sfsDb 提供加密存储功能，用户可以选择对敏感数据进行加密。
   - 我们采取严格的安全措施保护服务支持过程中收集的信息。

### 用户权利

- **数据访问**：用户有权访问存储在本地的所有数据。
- **数据修改**：用户有权修改或删除本地存储的数据。
- **数据导出**：用户可以随时导出存储的数据。

## GDPR 合规声明

### 适用范围

本 GDPR 合规声明适用于在欧盟地区使用 sfsDb 的用户。

### 数据处理原则

1. **合法性、公正性和透明性**：sfsDb 仅在用户明确同意的情况下处理数据，且处理过程透明。

2. **数据最小化**：sfsDb 仅处理必要的数据，且本地存储的数据完全由用户控制。

3. **目的限制**：sfsDb 处理数据的目的仅限于用户明确授权的范围。

4. **数据准确性**：sfsDb 确保用户可以随时更新和修正其数据。

5. **存储限制**：用户可以控制数据的存储期限，sfsDb 不会永久存储用户数据。

6. **完整性和保密性**：sfsDb 采取适当的技术和组织措施保护用户数据。

### 用户权利（GDPR）

根据 GDPR，欧盟用户享有以下权利：

- **知情权**：了解 sfsDb 如何处理其数据的权利。
- **访问权**：获取 sfsDb 存储的其个人数据的权利。
- **被遗忘权**：要求删除其个人数据的权利。
- **数据可携带权**：以结构化、常用格式接收其个人数据的权利。
- **限制处理权**：限制 sfsDb 处理其个人数据的权利。
- **反对权**：反对 sfsDb 处理其个人数据的权利。
- **自动化决策和 profiling 相关权利**：免受仅基于自动化处理的决策影响的权利。

### 数据保护联系人

如有关于数据保护的问题，请通过以下方式联系我们：

- 电子邮件：sfsweb@qq.com