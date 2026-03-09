# 代码示例

本目录包含本书各章节的完整可运行代码示例。

## 目录结构

```
code/
├── hello_world.go          # 第 1 章：Hello World 示例
├── transaction_examples.go # 第 5 章：事务处理综合示例
├── chapter03/              # 第 3 章：基础 CRUD 操作示例
│   ├── crud_basic.go
│   └── user_management.go
├── chapter04/              # 第 4 章：索引与查询优化示例
│   ├── index_basic.go
│   └── query_optimization.go
├── chapter05/              # 第 5 章：事务处理示例
│   ├── transaction_basic.go
│   └── bank_transfer.go
├── chapter06/              # 第 6 章：时序数据处理示例
│   ├── time_basic.go
│   └── sensor_data.go
├── chapter07/              # 第 7 章：加密存储示例
│   ├── encryption_basic.go
│   └── key_rotation.go
└── chapter08/              # 第 8 章：工业物联网实战项目
    └── iot_gateway/
        ├── main.go
        ├── device.go
        ├── sensor.go
        └── README.md
```

## 运行示例

### Hello World 示例

```bash
cd docs/ebook/code
go run hello_world.go
```

### 事务处理综合示例

```bash
cd docs/ebook/code
go run transaction_examples.go
```

该示例包含：
- 基本事务操作
- 批量操作
- 银行转账
- 订单处理

### 其他章节示例

每个章节的示例代码都有独立的 README，请查看对应目录。

## 代码约定

- 所有示例都包含完整的错误处理
- 资源清理使用 `defer`
- 数据库路径使用相对路径
- 示例运行后会自动清理测试数据

## 注意事项

- 运行示例前请确保已安装 sfsDb
- 部分示例可能需要较长时间运行
- 请确保有足够的磁盘空间
