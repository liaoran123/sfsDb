# 多表组合查询功能

本文档介绍 sfsDb 中的多表组合查询功能，该功能允许您在多个表之间执行类似 SQL JOIN 的操作，以实现更复杂的数据查询需求。

## 实现原理

sfsDb 的多表组合查询功能基于以下核心组件：

1. **TableIter 迭代器**：用于遍历表中的记录
2. **Map() 方法**：从一个表中提取指定字段的值作为映射
3. **match 包**：提供匹配条件的创建和管理
4. **SetMatch() 方法**：在迭代器上设置匹配条件
5. **GetRecords() 方法**：获取满足匹配条件的记录

## 基本使用方法

### 1. 准备工作

首先，您需要创建多个表并插入测试数据：

```go
// 创建表1
table1, err := TableNew("table1")
if err != nil {
    log.Fatalf("Failed to create table: %v", err)
}

// 设置字段
fields := map[string]any{"id": 0, "name": "", "age": 0}
err = table1.SetFields(fields)
if err != nil {
    log.Fatalf("Failed to set fields: %v", err)
}

// 创建主键索引
pk, _ := DefaultPrimaryKeyNew("pk")
pk.AddFields("id")
err = table1.CreateIndex(pk)
if err != nil {
    log.Fatalf("Failed to create primary key index: %v", err)
}

// 插入测试数据
testData := []map[string]any{
    {"id": 1, "name": "Alice", "age": 20},
    {"id": 2, "name": "Bob", "age": 25},
    {"id": 3, "name": "Charlie", "age": 30},
}

for _, data := range testData {
    _, err := table1.Insert(&data)
    if err != nil {
        log.Fatalf("Failed to insert test data: %v", err)
    }
}

// 类似地创建表2并插入数据...
```

### 2. 执行内连接查询

内连接（INNER JOIN）返回两个表中匹配的记录：

```go
// 创建表1和表2的迭代器
iter1, err := table1.Search(&map[string]any{"id": nil}) // 遍历所有记录
defer iter1.Release()

iter2, err := table2.Search(&map[string]any{"id": nil}) // 遍历所有记录
defer iter2.Release()

// 从表2中提取id字段的值作为映射
map2 := iter2.Map()
defer iter2.ReleaseMap(map2)

// 创建匹配条件：表1的id字段值必须在表2的id映射中
mach := match.NewAND([]string{"id"}, map2)

// 在表1的迭代器上设置匹配条件
iter1.SetMatch(mach)

// 获取匹配的记录（连接结果）
rd4 := iter1.GetRecords(true)
defer rd4.Release()

// 打印连接结果
for _, record := range rd4 {
    fmt.Println(record)
}
```

### 3. 执行外连接查询

外连接（OUTER JOIN）返回一个表中不匹配另一个表的记录：

```go
// 创建不匹配条件：表1的id字段值不在表2的id映射中
mach1 := match.NewAND([]string{"id"}, map2, false)

// 在表1的迭代器上设置不匹配条件
iter1.SetMatch(mach1)

// 获取不匹配的记录（外连接结果）
rd5 := iter1.GetRecords(true)
defer rd5.Release()

// 打印外连接结果
for _, record := range rd5 {
    fmt.Println(record)
}
```

### 4. 执行多表连接查询

您可以执行多个表之间的连接查询：

```go
// 创建表3的迭代器
iter3, err := table3.Search(&map[string]any{"id": nil}) // 遍历所有记录
defer iter3.Release()

// 从表3中提取id字段的值作为映射
map3 := iter3.Map()
defer iter3.ReleaseMap(map3)

// 创建匹配条件：表1的id字段值必须同时在表2和表3的id映射中
mach2 := match.NewAND([]string{"id"}, map3)

// 在表1的迭代器上设置多个匹配条件
iter1.SetMatch(mach, mach2)

// 获取匹配的记录（多表连接结果）
rd6 := iter1.GetRecords(true)
defer rd6.Release()

// 打印多表连接结果
for _, record := range rd6 {
    fmt.Println(record)
}
```

## 高级用法

### 1. 结合条件查询

您可以将多表连接与条件查询结合使用：

```go
// 创建年龄大于25的条件查询
iterAgeGreater, _ := table1.Search(&map[string]any{"age": 25}, util.GreaterThan)
defer iterAgeGreater.Release()

// 设置多表连接条件
iterAgeGreater.SetMatch(mach) // 使用之前创建的表2连接条件

// 获取结果
rdAgeGreater := iterAgeGreater.GetRecords(true)
defer rdAgeGreater.Release()

// 打印结果
for _, record := range rdAgeGreater {
    fmt.Println(record)
}
```

### 2. 资源管理

使用多表组合查询时，务必注意资源管理：

```go
// 正确释放迭代器
iter1, err := table1.Search(&map[string]any{"id": nil})
defer iter1.Release() // 或使用对象池：defer GlobalTableIterPool.Put(iter1)

// 正确释放映射
map2 := iter2.Map()
defer iter2.ReleaseMap(map2) // 或使用：defer PutMap(map2)

// 正确释放记录集
rd4 := iter1.GetRecords(true)
defer rd4.Release() // 或使用：defer record.PutRecords(rd4)
```

## 示例：完整的多表查询

以下是一个完整的多表查询示例，基于 `TestTestSelectForJoin1` 函数：

```go
// 创建表和插入数据的代码省略...

// 创建迭代器
iter1, err := table1.Search(&map[string]any{"id": nil})
defer iter1.Release()

iter2, err := table2.Search(&map[string]any{"id": nil})
defer iter2.Release()

iter3, err := table3.Search(&map[string]any{"id": nil})
defer iter3.Release()

// 测试1：内连接查询 - select table1.* from table1,table2 where table1.id=table2.id
fmt.Println("=== 内连接查询 ===")
map2 := iter2.Map()
defer iter2.ReleaseMap(map2)
mach := match.NewAND([]string{"id"}, map2)
iter1.SetMatch(mach)
rd4 := iter1.GetRecords(true)
defer rd4.Release()
for _, record := range rd4 {
    fmt.Println(record)
}

// 测试2：外连接查询 - select table1.* from table1,table2 where table1.id!=table2.id
fmt.Println("\n=== 外连接查询 ===")
mach1 := match.NewAND([]string{"id"}, map2, false)
iter1.SetMatch(mach1)
rd5 := iter1.GetRecords(true)
defer rd5.Release()
for _, record := range rd5 {
    fmt.Println(record)
}

// 测试3：多表连接查询 - select table1.* from table1,table2,table3 where table1.id=table2.id and table1.id=table3.id
fmt.Println("\n=== 多表连接查询 ===")
map3 := iter3.Map()
defer iter3.ReleaseMap(map3)
mach2 := match.NewAND([]string{"id"}, map3)
iter1.SetMatch(mach, mach2)
rd6 := iter1.GetRecords(true)
defer rd6.Release()
for _, record := range rd6 {
    fmt.Println(record)
}
```

## 性能优化

1. **使用索引**：确保连接字段上有索引，以提高查询性能
2. **合理使用对象池**：对于频繁执行的查询，使用对象池可以减少内存分配
3. **限制结果集大小**：使用分页或TopN功能限制返回的记录数量
4. **避免不必要的连接**：只连接真正需要的表

## 常见问题

### 1. 连接查询返回空结果

- 检查两个表中是否有匹配的字段值
- 确保使用了正确的字段名
- 验证表中是否有数据

### 2. 内存使用过高

- 对于大表，考虑使用对象池
- 及时释放不再使用的映射和记录集
- 考虑使用分页查询减少一次性加载的数据量

### 3. 查询性能慢

- 确保连接字段上有索引
- 减少连接的表数量
- 考虑使用更具体的搜索条件减少需要遍历的记录数

## 总结

sfsDb 的多表组合查询功能提供了一种灵活、高效的方式来执行类似 SQL JOIN 的操作。通过合理使用迭代器、映射和匹配条件，您可以实现复杂的数据查询需求，同时保持代码的简洁性和可读性。

请参考 `engine/tableiter_test.go` 文件中的 `TestTestSelectForJoin1` 函数获取更多示例代码。