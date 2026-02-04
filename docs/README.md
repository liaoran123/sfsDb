# sfsDb 数据库使用文档

## 项目简介

sfsDb 是一个灵活、高效的嵌入式数据库，支持多种数据类型、索引类型和查询方式。它提供了简洁的 API 设计，易于集成到各种应用场景中。

## 语言选择

- [中文文档](./zh/README.md)
- [English Documentation](./en/README.md)

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
        defer record.PutRecords(records)
    }
    
    for _, record := range records {
        fmt.Printf("记录: %v\n", record)
    }
}
```

## 文档结构

### 基础功能
- 数据库初始化
- 创建表与设置字段
- 插入数据
- 查询数据
- 删除记录
- 管理工具库使用

### 高级功能
- 主键管理
- 索引管理
- 全文搜索
- 字段修改
- 事务管理
- 其他功能

### 优化与最佳实践
- 对象池与内存管理
- 半结构化数据支持
- 最佳实践
- 常见问题
- 总结

## 贡献指南

欢迎贡献代码和文档！请参考 [CONTRIBUTING.md](https://github.com/liaoran123/sfsDb/blob/main/CONTRIBUTING.md) 了解如何参与项目。

## 许可证

sfsDb 使用 MIT 许可证。详见 [LICENSE](https://github.com/liaoran123/sfsDb/blob/main/LICENSE) 文件。