# sfsDb

<div align="center">
  <h1>📊 sfsDb：面向工业物联网与边缘计算的嵌入式数据库</h1>
  
  <p>🚀 专为 工业物联网(IIoT) 和 边缘计算 场景打造的轻量级 嵌入式关系型数据库</p>
  
  <p>基于 纯 Go (Golang) 语言开发，无 CGO 依赖，编译后仅为单个静态二进制文件，极致轻量，启动内存仅数 MB。完美适配各类 ARM 及 x86 架构的边缘网关与工业设备，解决资源受限环境下的数据持久化存储难题。</p>
  
  <div>
    <a href="https://github.com/liaoran123/sfsDb"><img src="https://img.shields.io/github/stars/liaoran123/sfsDb?style=social" alt="GitHub Stars"></a>
    <a href="https://github.com/liaoran123/sfsDb"><img src="https://img.shields.io/github/forks/liaoran123/sfsDb?style=social" alt="GitHub Forks"></a>
  </div>
</div>

## 核心特性 (Key Features)

### 🎯 纯 Go 开发 (Pure Go)
- 无 CGO 依赖，零配置，无需外部依赖库
- 支持交叉编译，部署极其简单
- 适合工业物联网和边缘计算场景的技术栈

### 📦 极致轻量 (Lightweight)
- 启动占用内存极低（仅数 MB）
- 对 CPU 和磁盘 I/O 消耗极小
- 专为资源受限设备优化，适合边缘网关数据存储

### 🚀 单文件部署 (Single Binary)
- 编译后为单一可执行文件
- 可轻松集成到 Docker 容器或直接运行在边缘网关上
- 简化工业设备的部署和维护流程

### 🔧 工业级可靠 (Industrial Ready)
- 针对 工业物联网(IIoT) 场景优化
- 支持高并发时序数据写入
- 具备断电保护与数据持久化能力
- 确保工业环境下的数据可靠性

### 🌍 广泛硬件兼容
- 支持 ARM (ARM64/ARMv7) 和 x86 架构
- 通用于主流工业网关、PLC 和边缘服务器
- 适应各种边缘侧存储环境

## 技术优势

### ✅ 轻量级设计，复杂查询场景支持
- 采用创新设计，在保持轻量级的同时，能够支持复杂的查询场景
- 超越了传统嵌入式数据库的能力边界，为应用提供更强大的数据处理能力

### ✅ 原生支持考据级全文索引
- 内置高性能全文索引引擎，提供精准的文本搜索能力
- 支持复杂的文本匹配和检索需求，满足考据级应用场景

### ✅ 基于 LevelDB 封装实现
- 使用 `github.com/syndtr/goleveldb/leveldb` 库作为存储引擎基础
- 充分利用 LevelDB 的 LSM-Tree 架构优势，提供高性能的读写操作

## 生产应用示例

### 考据级文档搜索引擎

sfsDb 已在实际生产环境中得到应用，其中最典型的案例是 **ReSearchCMS** - 一个专业的考据级文档搜索引擎。

- **项目地址**: [https://github.com/liaoran123/ReSearchCMS](https://github.com/liaoran123/ReSearchCMS)
- **应用场景**: 提供高精度、高性能的文档搜索功能，支持复杂的文本匹配和检索需求
- **技术亮点**: 充分利用 sfsDb 的原生全文索引和高性能查询能力，实现了考据级的文档搜索体验

### 示例项目
- [sfsDbIIoT](https://github.com/liaoran123/sfsDbIIoT) - 使用sfsDb实现的智能工厂设备监控系统，展示了sfsDb在工业IoT领域的应用
- [sfsDbGateway](https://github.com/liaoran123/sfsDbGateway) - 基于sfsDb的工业网关可靠性测试示例，验证sfsDb在网络波动、电源中断等恶劣工业环境下的可靠性和稳定性

## 主要目标用户群体

### 1. 🌐 边缘智能与 IoT 场景
- **边缘计算节点**：为资源受限的边缘设备提供本地数据存储能力，支持离线运行和边缘分析
- **IoT 网关设备**：高效处理和存储设备产生的时序数据，减少云端依赖，降低网络带宽消耗
- **智能终端设备**：在本地提供数据持久化能力，确保设备在网络不稳定时仍能正常运行

### 2. 🚀 微服务与容器化环境
- **微服务本地状态存储**：为无状态微服务提供轻量级的本地数据持久化方案，简化服务架构
- **容器化应用**：适合作为容器化应用的嵌入式存储组件，提供快速启动和低资源占用的特性
- **Serverless 函数**：为需要状态管理的 Serverless 函数提供临时数据存储能力

### 3. ⚡ 实时数据处理场景
- **实时分析系统**：支持高并发的写入操作，适合实时数据采集和分析场景
- **在线游戏服务器**：为游戏服务器提供高性能的本地存储，支持快速的玩家数据读写
- **实时监控系统**：高效存储和查询监控数据，确保系统状态的实时可见性

### 4. 🔒 安全敏感场景
- **隐私保护应用**：数据存储在本地，减少数据传输过程中的安全风险
- **金融科技应用**：满足金融场景对数据安全性和一致性的要求
- **医疗健康系统**：符合数据隐私法规要求，确保敏感健康数据的本地安全存储

### 5. 🎯 其他场景
- **嵌入式系统开发者**：轻量级、易部署的特性适合各类嵌入式设备
- **原型开发与快速迭代**：为项目原型阶段提供快速的数据存储解决方案
- **教育与研究**：适合作为数据库原理学习和研究的实验平台


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
	fmt.Println("sfsDb README示例代码测试")
	fmt.Println("====================")

	// 1. 初始化数据库
	fmt.Println("\n1. 初始化数据库")
	dbManager := storage.GetDBManager()
	_, err := dbManager.OpenDB("./readme_example_db")
	if err != nil {
		fmt.Printf("打开数据库失败: %v\n", err)
		return
	}
	defer dbManager.CloseDB()

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
		"id":      0,  // 用户ID
		"name":    "", // 用户名
		"age":     0,  // 年龄
		"email":   "", // 邮箱
		"address": "", // 地址
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
	{
		iter, err := userTable.Search(&map[string]any{"id": 1})
		defer iter.Release()
		if err != nil {
			fmt.Printf("搜索失败: %v\n", err)
			return
		}
		records := iter.GetRecordSet(true)
		defer records.Release()

		if len(records) > 0 {
			fmt.Printf("查询结果: %v\n", records[0])
		}
	}

	// 8. 普通索引查询
	fmt.Println("\n8. 普通索引查询")
	{
		nameIter, err := userTable.Search(&map[string]any{"name": "李四"})
		defer nameIter.Release()
		if err != nil {
			fmt.Printf("搜索失败: %v\n", err)
			return
		}
		nameRecords := nameIter.GetRecordSet(true)
		defer nameRecords.Release()

		if len(nameRecords) > 0 {
			fmt.Printf("按姓名查询结果: %v\n", nameRecords[0])
		}
	}

	// 9. 更新数据
	fmt.Println("\n9. 更新数据")
	updateData := map[string]any{
		"id":      1,                          // 用于定位记录
		"email":   "zhangsan_new@example.com", // 更新邮箱
		"address": "深圳市",                      // 更新地址
	}
	err = userTable.Update(&updateData)
	if err != nil {
		fmt.Printf("更新数据失败: %v\n", err)
		return
	}
	fmt.Println("更新数据成功")

	// 验证更新
	{
		iter, err := userTable.Search(&map[string]any{"id": 1})
		defer iter.Release()
		if err != nil {
			fmt.Printf("搜索失败: %v\n", err)
			return
		}
		records := iter.GetRecordSet(true)
		defer records.Release()
		if len(records) > 0 {
			fmt.Printf("更新后的数据: %v\n", records[0])
		}
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
	{
		iter, err := userTable.Search(&map[string]any{"id": 3})
		defer iter.Release()
		if err != nil {
			fmt.Printf("搜索失败: %v\n", err)
			return
		}
		records := iter.GetRecordSet(true)
		defer records.Release()
		fmt.Printf("删除后查询结果数: %d\n", len(records))
	}

	// 11. 查询所有数据
	fmt.Println("\n11. 查询所有数据")
	{
		allIter, err := userTable.Search(&map[string]any{})
		defer allIter.Release()
		if err != nil {
			fmt.Printf("搜索失败: %v\n", err)
			return
		}
		allRecords := allIter.GetRecordSet(true)
		defer allRecords.Release()
		fmt.Printf("当前表中共有 %d 条记录\n", len(allRecords))
		for i, r := range allRecords {
			fmt.Printf("记录 %d: %v\n", i+1, r)
		}
	}

	fmt.Println("\n测试完成，所有操作均成功执行！")
}

```
## 文档

### 使用指南
- [中文文档首页](./docs/zh/README.md) - 中文文档总览
- [English Documentation Home](./docs/en/README.md) - English documentation overview
- [基础操作指南](./docs/zh/basic/) - 包含表创建、字段修改等基础操作
- [Basic Operations Guide](./docs/en/basic/) - Includes table creation, field modification and other basic operations

### 核心功能文档
- [全文索引使用指南](./docs/zh/advanced/full_text_search.md) - 全文索引的高级使用方法（中文）
- [Full Text Search Guide](./docs/en/advanced/full_text_search.md) - Advanced full text search usage (English)
- [索引管理指南](./docs/zh/advanced/index_management.md) - 索引的创建、管理和优化（中文）
- [Index Management Guide](./docs/en/advanced/index_management.md) - Index creation, management and optimization (English)

### 性能测试报告

#### 性能对比图表

<div align="center">
  <img src="./docs/performance/database_comparison.png" alt="数据库性能比较" width="600">
  <p>sfsDb 与其他嵌入式数据库性能比较</p>
</div>

<div align="center">
  <img src="./docs/performance/acid_comparison.png" alt="ACID vs 非ACID性能比较" width="600">
  <p>ACID vs 非ACID模式性能比较</p>
</div>

<div align="center">
  <img src="./docs/performance/data_volume_impact.png" alt="数据量对性能的影响" width="600">
  <p>数据量增长对性能的影响</p>
</div>

<div align="center">
  <img src="./docs/performance/concurrency_impact.png" alt="并发对性能的影响" width="600">
  <p>并发增长对性能的影响</p>
</div>

- [综合性能基准测试报告](./engine/comprehensive_benchmark_report.md) - 全面的性能测试和分析，包括读写性能、并发性能、不同数据量下的表现，以及与其他数据库的性能比较
- [事务基准测试报告](./engine/transaction_benchmark_report.md) - 详细的事务性能测试和分析，包括单事务和多事务场景下的性能表现，以及事务优化效果
- [ACID vs Non-ACID 性能比较报告](./engine/acid_vs_nonacid_benchmark_report.md) - ACID特性对性能的影响分析，比较不同事务模式下的性能差异和适用场景
- [时序数据库性能比较报告](./docs/performance/time_series_benchmark.md) - time包基准测试与其他时序数据库性能比较，包括单线程和并发性能测试，以及时间序列数据处理的效率分析

### 性能优势分析
- [sfsDb性能优势分析文章](./docs/marketing/performance_advantage_article.md) - 详细分析sfsDb如何通过无SQL设计和嵌入式架构实现性能突破


### 资源管理
- [对象池使用指南](./docs/zh/optimization/object_pool.md) - 高效的对象复用机制（中文）
- [Object Pool Guide](./docs/en/optimization/object_pool.md) - Efficient object reuse mechanism (English)
- [索引缓存指南](./docs/zh/optimization/index_cache.md) - 提升查询性能的索引缓存机制（中文）
- [Index Cache Guide](./docs/en/optimization/index_cache.md) - Index caching mechanism for improved query performance (English)

### 高级功能
- [主键管理指南](./docs/zh/advanced/primary_key.md) - 主键的设计和使用（中文）
- [Primary Key Guide](./docs/en/advanced/primary_key.md) - Primary key design and usage (English)
- [表管理指南](./docs/zh/advanced/management.md) - 表的创建、修改和管理（中文）
- [Table Management Guide](./docs/en/advanced/management.md) - Table creation, modification and management (English)

### API 参考
- [API 参考文档](./docs/api.md) - 详细的 API 文档

## 联系我们

如果您有任何问题或建议，欢迎通过以下方式联系我们：

- **邮箱**: sfsweb@qq.com

## 贡献

欢迎提交 Issue 和 Pull Request 来帮助改进 sfsDb！

## 许可证

sfsDb 采用双许可证模式：

### 核心引擎 - MIT 许可证
- **开源免费**: 核心数据库引擎使用 MIT 开源许可证
- **商业友好**: 允许自由使用、修改和商业分发
- **简洁灵活**: 许可证文本简洁，限制少，易于理解和使用
- **许可证文本**: 详见 [LICENSE](./LICENSE) 文件


### 依赖库许可证
- **LevelDB**: 使用 BSD 2-clause 开源许可证，与 MIT 许可证兼容
- **其他依赖**: 详见 `go.mod` 文件中的依赖声明

### 许可证兼容性
MIT 许可证是一种广泛使用的开源许可证，具有以下优势：
- **简洁明了**: 许可证文本简洁，易于理解和使用
- **高度兼容**: 与几乎所有其他开源许可证兼容
- **商业友好**: 允许在商业项目中自由使用和修改
- **社区广泛采用**: 被众多开源项目采用，是最受欢迎的开源许可证之一

这意味着 sfsDb 可以自由地基于 LevelDB 进行封装和商业开发，用户可以放心在商业项目中使用 sfsDb。

