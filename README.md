# sfsDb

sfsDb 是一个轻量级嵌入式数据库，专注于提供高性能、灵活的存储解决方案，同时保持代码简洁和资源占用低。

## 项目特点

### 核心亮点

1. **轻量级设计，复杂场景支持**
   - 采用创新设计，在保持轻量级的同时，能够支持相当复杂的查询场景
   - 超越了传统嵌入式数据库的能力边界，为应用提供更强大的数据处理能力

2. **原生支持考据级全文索引**
   - 内置高性能全文索引引擎，提供精准的文本搜索能力
   - 支持复杂的文本匹配和检索需求，满足考据级应用场景

### 其他特性

- **灵活的存储模型**：支持多种数据结构和存储格式
- **高性能查询**：优化的查询引擎，提供快速的数据检索
- **简单易用的 API**：简洁直观的接口设计，降低开发成本
- **跨平台兼容**：支持多种操作系统和环境

## 快速开始

### 安装

```bash
# 通过 Go 模块安装
go get github.com/liaoran123/sfsDb
```

### 基本使用

```go
package main

import (
    "fmt"

    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    fmt.Println("SFSDB Example")
    fmt.Println("=====================")

    // 1. 初始化数据库
    fmt.Println("\n1. 初始化数据库")
    storage.OpenDefaultDb("./example_db")
    defer storage.CloseDb()

    // 2. 创建文章表
    fmt.Println("\n2. 创建文章表")
    articleTable, err := engine.TableNew("articles")
    if err != nil {
        fmt.Printf("创建表失败: %v\n", err)
        return
    }

    // 3. 设置字段
    fmt.Println("\n3. 设置字段")
    articleFields := map[string]any{
        "id":      0,  // 自动增值主键
        "title":   "", // 文章标题
        "content": "", // 文章内容
        "author":  "", // 作者
    }
    err = articleTable.SetFields(articleFields)
    if err != nil {
        fmt.Printf("设置字段失败: %v\n", err)
        return
    }

    // 4. 创建全文索引
    fmt.Println("\n4. 创建全文索引")
    fullTextIdx, err := engine.DefaultFullTextIndexNew("content_ft")
    if err != nil {
        fmt.Printf("创建全文索引失败: %v\n", err)
        return
    }
    fullTextIdx.AddFields("content", "id")
    err = fullTextIdx.SetFullField("content", 5) // 设置content为全文索引字段
    if err != nil {
        fmt.Printf("设置全文索引字段失败: %v\n", err)
        return
    }
    err = articleTable.CreateIndex(fullTextIdx)
    if err != nil {
        fmt.Printf("创建全文索引失败: %v\n", err)
        return
    }

    // 5. 插入文章数据
    fmt.Println("\n5. 插入文章数据")
    articles := []map[string]any{
        {"title": "Go语言入门", "content": "Go语言是一种开源的编程语言，它能让构造简单、可靠且高效的软件变得容易。", "author": "张三"},
        {"title": "数据库基础", "content": "数据库是按照数据结构来组织、存储和管理数据的仓库。", "author": "李四"},
        {"title": "网络编程", "content": "网络编程是指编写运行在多个设备（计算机）之间的程序。", "author": "王五"},
    }

    for _, article := range articles {
        id, err := articleTable.Insert(&article)
        if err != nil {
            fmt.Printf("插入数据失败: %v\n", err)
            return
        }
        fmt.Printf("插入文章成功, ID: %d, Title: %s\n", id, article["title"])
    }

    // 6. 全文搜索
    fmt.Println("\n6. 全文搜索")
    searchTerm := "编程"
    searchFields := map[string]any{"content": searchTerm}
    iter := articleTable.Search(&searchFields)
    defer iter.Release()

    fmt.Printf("搜索包含 '%s' 的文章:\n", searchTerm)
    for iter.First(); iter.Valid(); iter.Next() {
        record := articleTable.ParseRecordValue(iter.Value())
        fmt.Printf("- ID: %d, Title: %s\n", record["id"], record["title"])
    }

    fmt.Println("\n示例完成!")
}
```

## 文档

- [完整 API 文档](./docs/api.md)
- [使用指南](./docs/guide.md)
- [性能基准测试报告](./engine/comprehensive_benchmark_report.md)

## 贡献

欢迎提交 Issue 和 Pull Request 来帮助改进 sfsDb！

## 许可证

sfsDb 使用 MIT 许可证，详见 [LICENSE](./LICENSE) 文件。