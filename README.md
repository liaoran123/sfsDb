# sfsDb

[![Go Version](https://img.shields.io/badge/Go-1.25.3-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

sfsDb是一个轻量级、高性能的关系型数据库，基于KV存储（默认使用LevelDB）实现，提供SQL支持和事务处理能力。

## 项目特点

- **分层架构**：清晰的分层设计，便于扩展和维护
- **模块化设计**：各功能模块独立，便于测试和扩展
- **高性能**：基于LevelDB实现，提供高效的存储性能
- **易用性**：简洁的API设计，便于用户使用
- **可扩展性**：支持多种存储后端和索引类型
- **SQL支持**：预留SQL解析和执行框架
- **事务处理**：支持事务的ACID特性

## 安装

```bash
go get github.com/liaoran123/sfsDb
```

## 快速开始

### 基本使用示例

```go
package main

import (
	"fmt"
	"github.com/liaoran123/sfsDb/api"
	"github.com/liaoran123/sfsDb/engine/types"
)

func main() {
	// 配置数据库
	config := api.Config{
		Path: "./test_db",
	}

	// 打开数据库
	db, err := api.Open(config)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// 定义表结构
	userSchema := &types.TableSchema{
		Name: "users",
		Fields: []*types.Field{
			{Name: "id", Type: types.TypeInt64, Nullable: false, Comment: "用户ID"},
			{Name: "name", Type: types.TypeString, Nullable: false, Comment: "用户名"},
			{Name: "email", Type: types.TypeString, Nullable: false, Comment: "邮箱"},
			{Name: "age", Type: types.TypeInt32, Nullable: true, Comment: "年龄"},
		},
		PrimaryKey: []string{"id"},
		Indexes: []*types.IndexDef{
			{
				Name:   "idx_email",
				Type:   types.IndexTypeBTree,
				Fields: []string{"email"},
			},
		},
	}

	// 创建表
	_, err = db.CreateTable("users", userSchema)
	if err != nil {
		panic(err)
	}

	// 列出所有表
	tableNames, err := db.ListTables()
	if err != nil {
		panic(err)
	}

	fmt.Println("Tables:", tableNames)
}
```

## 项目结构

```
sfsDb/
├── api/            # 数据库对外API接口
├── engine/         # 数据库核心引擎
│   ├── table/      # 表和索引实现
│   └── types/      # 数据类型定义
├── examples/       # 使用示例代码
├── storage/        # 存储层
│   └── kv/         # KV存储接口和LevelDB实现
├── go.mod          # Go模块定义
├── go.sum          # 依赖校验和
└── README.md       # 项目文档
```

### 核心模块说明

#### api/
- 提供数据库对外的API接口
- 定义DB、Table、Tx、Result等核心接口
- 提供数据库打开、关闭、表创建、查询等操作

#### engine/
- 数据库核心引擎实现
- **table/**：表结构管理、索引管理、数据操作
- **types/**：数据类型定义、表结构定义、索引定义

#### storage/
- 存储层抽象和实现
- **kv/**：KV存储接口和LevelDB实现
- 支持多种KV存储后端扩展

## 核心功能

### 1. 表管理
- 创建表
- 删除表
- 获取表
- 列出所有表

### 2. 数据操作
- 插入数据
- 更新数据
- 删除数据
- 查询数据

### 3. 索引管理
- 主键索引
- 二级索引
- 支持多种索引类型（B-tree等）

### 4. 事务支持
- 事务开始、提交、回滚
- ACID特性支持

### 5. SQL支持
- 预留SQL解析和执行框架
- 支持SQL语句的执行和查询

## API文档

### 数据库操作

#### 打开数据库
```go
func Open(config Config) (DB, error)
```

#### 数据库接口
```go
type DB interface {
    Open(config Config) error
    Close() error
    CreateTable(name string, schema *types.TableSchema) (Table, error)
    DropTable(name string) error
    GetTable(name string) (Table, error)
    ListTables() ([]string, error)
    Begin() (Tx, error)
    Exec(sql string, args ...interface{}) (Result, error)
    Query(sql string, args ...interface{}) (Result, error)
}
```

### 表操作

#### 表接口
```go
type Table interface {
    GetName() string
    GetSchema() *types.TableSchema
    Insert(values map[string]interface{}) (Result, error)
    Update(where Expr, values map[string]interface{}) (Result, error)
    Delete(where Expr) (Result, error)
    Select(fields []string, where Expr) (Result, error)
    Get(primaryKey interface{}) (map[string]interface{}, error)
}
```

## 后续扩展方向

1. 完善SQL解析和执行功能
2. 实现完整的事务处理机制
3. 支持分布式部署
4. 提供更多索引类型支持（哈希索引、全文索引等）
5. 完善监控和调试工具
6. 提供可视化管理界面
7. 支持更多数据类型
8. 实现数据备份和恢复功能

## 贡献指南

欢迎对sfsDb项目做出贡献！如果您有任何建议或问题，请提交Issue或Pull Request。

### 贡献流程

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 许可证

sfsDb项目采用MIT许可证，详情请查看[LICENSE](LICENSE)文件。

## 联系方式

如有任何问题或建议，请通过以下方式联系我们：

- GitHub: [https://github.com/liaoran123/sfsDb](https://github.com/liaoran123/sfsDb)
- 邮箱: [liaoran123@example.com](mailto:liaoran123@example.com)
