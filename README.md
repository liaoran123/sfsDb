# sfsDb

sfsDb 是一个轻量级嵌入式关系型数据库，专注于提供高性能、灵活的存储解决方案，同时保持代码简洁和资源占用低。

## 项目特点

### 核心亮点

1. **轻量级设计，复杂查询场景支持**
   - 采用创新设计，在保持轻量级的同时，能够支持相当复杂的查询场景
   - 超越了传统嵌入式数据库的能力边界，为应用提供更强大的数据处理能力

2. **原生支持考据级全文索引**
   - 内置高性能全文索引引擎，提供精准的文本搜索能力
   - 支持复杂的文本匹配和检索需求，满足考据级应用场景

3. **基于 LevelDB 封装实现**
   - 使用 `github.com/syndtr/goleveldb/leveldb` 库作为存储引擎基础
   - 充分利用 LevelDB 的 LSM-Tree 架构优势，提供高性能的读写操作

### 其他特性

- **灵活的存储模型**：支持多种数据结构和存储格式
- **高性能查询**：优化的查询引擎，提供快速的数据检索
- **简单易用的 API**：简洁直观的接口设计，降低开发成本
- **跨平台兼容**：支持多种操作系统和环境

## 生产应用示例

### 考据级文档搜索引擎

sfsDb 已在实际生产环境中得到应用，其中最典型的案例是 **ReSearchCMS** - 一个专业的考据级文档搜索引擎。

- **项目地址**: [https://github.com/liaoran123/ReSearchCMS](https://github.com/liaoran123/ReSearchCMS)
- **应用场景**: 提供高精度、高性能的文档搜索功能，支持复杂的文本匹配和检索需求
- **技术亮点**: 充分利用 sfsDb 的原生全文索引和高性能查询能力，实现了考据级的文档搜索体验

## 主要目标用户群体

- **嵌入式系统开发者**：轻量级、易部署的特性适合嵌入式设备
- **边缘计算场景**：边缘节点需要本地数据存储，sfsDb 可以提供轻量级解决方案
- **IoT 应用开发者**：针对 IoT 设备产生的时序数据，优化存储和查询
- **游戏开发者**：游戏服务器需要高性能的本地存储，适合使用轻量级数据库
- **其他**：......


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
	"github.com/liaoran123/sfsDb/util"
)

func main() {
	fmt.Println("sfsDb 基本使用示例")
	fmt.Println("==================")

	// 1. 初始化数据库
	fmt.Println("\n1. 初始化数据库")
	_, err := storage.OpenDefaultDb("./basic_example_db")
	if err != nil {
		fmt.Printf("打开数据库失败: %v\n", err)
		return
	}
	defer storage.CloseDb()

	// 2. 创建/打开用户表
	fmt.Println("\n2. 创建用户表")
	userTable, err := engine.TableNew("users")
	if err != nil {
		fmt.Printf("创建表失败: %v\n", err)
		return
	}

	// 3. 设置字段
	fmt.Println("\n3. 设置字段")
	userFields := map[string]any{
		"id":       0,   // 用户ID
		"name":     "",  // 用户名
		"age":      0,   // 年龄
		"email":    "",  // 邮箱
		"address":  "",  // 地址
	}
	err = userTable.SetFields(userFields)
	if err != nil {
		fmt.Printf("设置字段失败: %v\n", err)
		return
	}

	// 4. 创建主键索引
	fmt.Println("\n4. 创建主键索引")
	primaryKey, err := engine.DefaultPrimaryKeyNew("id")
	if err != nil {
		fmt.Printf("创建主键索引失败: %v\n", err)
		return
	}
	primaryKey.AddFields("id")
	err = userTable.CreateIndex(primaryKey)
	if err != nil {
		fmt.Printf("创建索引失败: %v\n", err)
		return
	}

	// 5. 创建普通索引
	fmt.Println("\n5. 创建普通索引")
	nameIndex, err := engine.DefaultNormalIndexNew("name_index")
	if err != nil {
		fmt.Printf("创建普通索引失败: %v\n", err)
		return
	}
	nameIndex.AddFields("name")
	err = userTable.CreateIndex(nameIndex)
	if err != nil {
		fmt.Printf("创建索引失败: %v\n", err)
		return
	}

	// 6. 插入数据
	fmt.Println("\n6. 插入数据")
	users := []map[string]any{
		{"id": 1, "name": "张三", "age": 25, "email": "zhangsan@example.com", "address": "北京市"},
		{"id": 2, "name": "李四", "age": 30, "email": "lisi@example.com", "address": "上海市"},
		{"id": 3, "name": "王五", "age": 35, "email": "wangwu@example.com", "address": "广州市"},
	}

	for _, user := range users {
		currentID, err := userTable.Insert(&user)
		if err != nil {
			fmt.Printf("插入数据失败: %v\n", err)
			return
		}
		fmt.Printf("插入用户成功, ID: %d\n", currentID)
	}

	// 7. 主键查询
	fmt.Println("\n7. 主键查询")
	iter := userTable.Search(&map[string]any{"id": 1})
	defer iter.Release()
	records := iter.GetRecordSet(true)
	defer PutRecords(records)

	if len(records) > 0 {
		fmt.Printf("查询结果: %v\n", records[0])
	}

	// 8. 普通索引查询
	fmt.Println("\n8. 普通索引查询")
	nameIter := userTable.Search(&map[string]any{"name": "李四"})
	defer nameIter.Release()
	nameRecords := nameIter.GetRecordSet(true)
	defer PutRecordSet(nameRecords)

	if len(nameRecords) > 0 {
		fmt.Printf("按姓名查询结果: %v\n", nameRecords[0])
	}

	// 9. 更新数据
	fmt.Println("\n9. 更新数据")
	updateData := map[string]any{
		"id":      1,                // 用于定位记录
		"email":   "zhangsan_new@example.com", // 更新邮箱
		"address": "深圳市",          // 更新地址
	}
	err = userTable.Update(&updateData)
	if err != nil {
		fmt.Printf("更新数据失败: %v\n", err)
		return
	}
	fmt.Println("更新数据成功")

	// 验证更新
	iter = userTable.Search(&map[string]any{"id": 1})
	defer iter.Release()
	records = iter.GetRecords(true)
	if len(records) > 0 {
		fmt.Printf("更新后的数据: %v\n", records[0])
	}

	// 10. 删除数据
	fmt.Println("\n10. 删除数据")
	deleteData := map[string]any{
		"id": 3, // 用于定位要删除的记录
	}
	err = userTable.Delete(&deleteData)
	if err != nil {
		fmt.Printf("删除数据失败: %v\n", err)
		return
	}
	fmt.Println("删除数据成功")

	// 验证删除
	iter = userTable.Search(&map[string]any{"id": 3})
	defer iter.Release()
	records = iter.GetRecordSet(true)
	defer PutRecords(records)
	fmt.Printf("删除后查询结果数: %d\n", len(records))

	// 11. 查询所有数据
	fmt.Println("\n11. 查询所有数据")
	allIter := userTable.Search(&map[string]any{})
	defer allIter.Release()
	allRecords := allIter.GetRecordSet(true)
	defer PutRecordSet(allRecords)
	fmt.Printf("当前表中共有 %d 条记录\n", len(allRecords))
	for i, record := range allRecords {
		fmt.Printf("记录 %d: %v\n", i+1, record)
	}
}

```

## 文档

### 使用指南
- [中文详细使用指南](./docs/sfsDbUserGuide.md)
- [English User Guide](./docs/sfsDbUserGuide.en.md)

### 功能包文档
- [监控包文档](./monitoring/README.md) - 提供指标收集、告警管理和Prometheus集成
- [安全包文档](./security/README.md) - 提供认证、授权、加密和审计功能
- [扩展包文档](./extensions/README.md) - 提供插件管理、索引扩展、存储引擎扩展和数据类型扩展

### 其他文档
- [API 文档](./docs/guide.md)
- [性能基准测试报告](./engine/comprehensive_benchmark_report.md)

## 贡献

欢迎提交 Issue 和 Pull Request 来帮助改进 sfsDb！

## 许可证

sfsDb 采用双许可证模式：

### 核心引擎 - BSD 2-clause 许可证
- **开源免费**: 核心数据库引擎使用 BSD 2-clause 开源许可证
- **商业友好**: 允许自由使用、修改和商业分发
- **许可证文本**: 详见 [LICENSE](./LICENSE) 文件

### 企业版插件 - 商业许可证
- **高级功能**: 企业版插件（如高级安全特性、多节点协调等）使用商业许可证
- **支持服务**: 包含专业技术支持、定期更新和企业级功能
- **许可证文本**: 详见 [LICENSE.COMMERCIAL](./LICENSE.COMMERCIAL) 文件

### 依赖库许可证
- **LevelDB**: 使用 BSD 2-clause 开源许可证，与 sfsDb 核心引擎许可证一致
- **其他依赖**: 详见 `go.mod` 文件中的依赖声明

### 许可证兼容性
BSD 2-clause 许可证是最自由的开源许可证之一，与所有主要商业许可证完全兼容：
- 允许在闭源商业产品中使用
- 允许修改后闭源分发
- 无需在衍生作品中包含原始许可证文本
- 与 LevelDB 的许可证保持一致，确保技术栈的许可证兼容性

这意味着 sfsDb 可以自由地基于 LevelDB 进行封装和商业开发，用户可以放心在商业项目中使用 sfsDb。

