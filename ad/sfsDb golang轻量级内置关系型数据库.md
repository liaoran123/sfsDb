# sfsDb golang轻量级内置关系型数据库

## 0. sfsDb 简介

sfsDb 是一款轻量级、高性能的嵌入式数据库，专为 Go 语言应用设计。它采用了灵活的数据模型和高效的索引机制，支持复杂的查询操作，并可以通过 `TableIter` 和 `Match` 等mach接口自定义强大和灵活的查询功能，特别适合需要快速开发、低部署成本的应用场景。

### 核心特性

- **轻量级嵌入式设计**：无需独立的数据库服务器，直接嵌入到 Go 应用中
- **灵活的数据模型**：支持半结构化数据，无需预定义表结构
- **高效的索引机制**：支持多种索引类型，包括主键索引、普通索引和全文索引
- **强大的查询能力**：支持多表组合查询、范围查询、模糊查询等
- **良好的扩展性**：提供了插件机制，允许自定义索引和查询逻辑
- **易于使用的 API**：简洁直观的 API 设计，降低学习曲线

### 适用场景

- 桌面应用和嵌入式设备
- 小型 Web 应用
- 边缘计算和 IoT 设备
- 需要快速开发的原型项目
- 需要轻量级存储解决方案的场景

## 1. 功能介绍

sfsDb 支持多表组合查询功能，允许开发者在多个表之间进行数据关联和查询。本示例将演示如何使用 sfsDb 创建多个表，插入测试数据，并执行多表组合查询。

## 2. 准备工作

### 2.1 导入依赖

```go
import (
    "github.com/liaoran123/sfsDb/engine"
)
```

## 3. 示例代码

### 3.1 创建测试表

```go
// 创建第一个测试表
table1, err := engine.TableNew("test_search_comprehensive1")
if err != nil {
    panic(err)
}

// 设置表字段
fields := map[string]any{"id": 0, "name": "", "age": 0, "score": 0.0, "active": false}
err = table1.SetFields(fields)
if err != nil {
    panic(err)
}

// 为id字段创建主键索引
pk, _ := engine.DefaultPrimaryKeyNew("pk")
pk.AddFields("id")
err = table1.CreateIndex(pk)
if err != nil {
    panic(err)
}

// 为age字段创建二级索引
ageIdx, _ := engine.DefaultNormalIndexNew("age_index")
ageIdx.AddFields("age")
err = table1.CreateIndex(ageIdx)
if err != nil {
    panic(err)
}
```

### 3.2 向第一个表插入测试数据

```go
// 插入测试数据
testData := []map[string]any{
    {"id": 1, "name": "Alice", "age": 20, "score": 85.5, "active": true},
    {"id": 2, "name": "Bob", "age": 25, "score": 90.0, "active": true},
    {"id": 3, "name": "Charlie", "age": 30, "score": 75.5, "active": false},
    {"id": 4, "name": "David", "age": 35, "score": 95.0, "active": true},
    {"id": 5, "name": "Eve", "age": 40, "score": 80.0, "active": false},
}

for _, data := range testData {
    _, err := table1.Insert(&data)
    if err != nil {
        panic(err)
    }
}
```

### 3.3 创建第二个测试表

```go
// 创建第二个测试表
table2, err := engine.TableNew("test_search_comprehensive2")
if err != nil {
    panic(err)
}

// 设置表字段
fields2 := map[string]any{"id": 0, "name": "", "age": 0, "score": 0.0, "active": false}
err = table2.SetFields(fields2)
if err != nil {
    panic(err)
}

// 为id字段创建主键索引
pk2, _ := engine.DefaultPrimaryKeyNew("pk")
pk2.AddFields("id")
err = table2.CreateIndex(pk2)
if err != nil {
    panic(err)
}

// 为age字段创建二级索引
ageIdx2, _ := engine.DefaultNormalIndexNew("age_index")
ageIdx2.AddFields("age")
err = table2.CreateIndex(ageIdx2)
if err != nil {
    panic(err)
}
```

### 3.4 向第二个表插入测试数据

```go
// 插入测试数据
testData2 := []map[string]any{
    {"id": 3, "name": "Charlie", "age": 30, "score": 75.5, "active": false},
    {"id": 4, "name": "David", "age": 35, "score": 95.0, "active": true},
    {"id": 5, "name": "Eve", "age": 40, "score": 80.0, "active": false},
    {"id": 6, "name": "Frank", "age": 45, "score": 88.5, "active": true},
    {"id": 7, "name": "Grace", "age": 50, "score": 92.0, "active": true},
}

for _, data := range testData2 {
    _, err := table2.Insert(&data)
    if err != nil {
        panic(err)
    }
}
```

### 3.5 创建第三个测试表

```go
// 创建第三个测试表
table3, err := engine.TableNew("test_search_comprehensive3")
if err != nil {
    panic(err)
}

// 设置表字段
fields3 := map[string]any{"id": 0, "name": "", "age": 0, "score": 0.0, "active": false}
err = table3.SetFields(fields3)
if err != nil {
    panic(err)
}

// 为id字段创建主键索引
pk3, _ := engine.DefaultPrimaryKeyNew("pk")
pk3.AddFields("id")
err = table3.CreateIndex(pk3)
if err != nil {
    panic(err)
}

// 为age字段创建二级索引
ageIdx3, _ := engine.DefaultNormalIndexNew("age_index")
ageIdx3.AddFields("age")
err = table3.CreateIndex(ageIdx3)
if err != nil {
    panic(err)
}
```

## 4. 执行多表组合查询

### 4.1 基本查询示例

```go
// 搜索第一个表中id=1的数据
iter1 := table1.Search(&map[string]any{"id": 1})
defer iter1.Release()
records1 := iter1.GetRecords(true)
fmt.Printf("表1查询结果: %v\n", records1)

// 搜索第二个表中age>30的数据
iter2 := table2.Search(&map[string]any{"age": 30}, util.GreaterThan)
defer iter2.Release()
records2 := iter2.GetRecords(true)
fmt.Printf("表2查询结果: %v\n", records2)
```

### 4.2 多表组合查询示例

```go
// 示例：查询所有表中age>30的活跃用户
fmt.Println("\n=== 多表组合查询结果 ===")

// 查询第一个表
iter1 = table1.Search(&map[string]any{"age": 30, "active": true}, util.GreaterThan)
defer iter1.Release()
records1 = iter1.GetRecords(true)

// 查询第二个表
iter2 = table2.Search(&map[string]any{"age": 30, "active": true}, util.GreaterThan)
defer iter2.Release()
records2 = iter2.GetRecords(true)

// 合并查询结果
allRecords := append(records1, records2...)
fmt.Printf("合并后的查询结果数: %d\n", len(allRecords))
for i, record := range allRecords {
    fmt.Printf("结果 %d: %v\n", i+1, record)
}
```

## 5. 完整示例代码

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/util"
)

func main() {
    fmt.Println("sfsDb 多表组合查询示例")
    fmt.Println("========================")

    // 创建第一个测试表
    table1, err := engine.TableNew("test_search_comprehensive1")
    if err != nil {
        panic(err)
    }

    // 设置表字段
    fields := map[string]any{"id": 0, "name": "", "age": 0, "score": 0.0, "active": false}
    err = table1.SetFields(fields)
    if err != nil {
        panic(err)
    }

    // 为id字段创建主键索引
    pk, _ := engine.DefaultPrimaryKeyNew("pk")
    pk.AddFields("id")
    err = table1.CreateIndex(pk)
    if err != nil {
        panic(err)
    }

    // 为age字段创建二级索引
    ageIdx, _ := engine.DefaultNormalIndexNew("age_index")
    ageIdx.AddFields("age")
    err = table1.CreateIndex(ageIdx)
    if err != nil {
        panic(err)
    }

    // 插入测试数据
    testData := []map[string]any{
        {"id": 1, "name": "Alice", "age": 20, "score": 85.5, "active": true},
        {"id": 2, "name": "Bob", "age": 25, "score": 90.0, "active": true},
        {"id": 3, "name": "Charlie", "age": 30, "score": 75.5, "active": false},
        {"id": 4, "name": "David", "age": 35, "score": 95.0, "active": true},
        {"id": 5, "name": "Eve", "age": 40, "score": 80.0, "active": false},
    }

    for _, data := range testData {
        _, err := table1.Insert(&data)
        if err != nil {
            panic(err)
        }
    }

    // 创建第二个测试表
    table2, err := engine.TableNew("test_search_comprehensive2")
    if err != nil {
        panic(err)
    }

    // 设置表字段
    fields2 := map[string]any{"id": 0, "name": "", "age": 0, "score": 0.0, "active": false}
    err = table2.SetFields(fields2)
    if err != nil {
        panic(err)
    }

    // 为id字段创建主键索引
    pk2, _ := engine.DefaultPrimaryKeyNew("pk")
    pk2.AddFields("id")
    err = table2.CreateIndex(pk2)
    if err != nil {
        panic(err)
    }

    // 为age字段创建二级索引
    ageIdx2, _ := engine.DefaultNormalIndexNew("age_index")
    ageIdx2.AddFields("age")
    err = table2.CreateIndex(ageIdx2)
    if err != nil {
        panic(err)
    }

    // 插入测试数据
    testData2 := []map[string]any{
        {"id": 3, "name": "Charlie", "age": 30, "score": 75.5, "active": false},
        {"id": 4, "name": "David", "age": 35, "score": 95.0, "active": true},
        {"id": 5, "name": "Eve", "age": 40, "score": 80.0, "active": false},
        {"id": 6, "name": "Frank", "age": 45, "score": 88.5, "active": true},
        {"id": 7, "name": "Grace", "age": 50, "score": 92.0, "active": true},
    }

    for _, data := range testData2 {
        _, err := table2.Insert(&data)
        if err != nil {
            panic(err)
        }
    }

    // 执行多表组合查询
    fmt.Println("\n=== 执行多表组合查询 ===")

    // 查询第一个表中age>30的活跃用户
    iter1 := table1.Search(&map[string]any{"age": 30, "active": true}, util.GreaterThan)
    defer iter1.Release()
    records1 := iter1.GetRecords(true)
    fmt.Printf("表1查询结果数: %d\n", len(records1))

    // 查询第二个表中age>30的活跃用户
    iter2 := table2.Search(&map[string]any{"age": 30, "active": true}, util.GreaterThan)
    defer iter2.Release()
    records2 := iter2.GetRecords(true)
    fmt.Printf("表2查询结果数: %d\n", len(records2))

    // 合并查询结果
    allRecords := append(records1, records2...)
    fmt.Printf("\n合并后的查询结果数: %d\n", len(allRecords))
    for i, record := range allRecords {
        fmt.Printf("结果 %d: ID=%d, Name=%s, Age=%d, Score=%.1f, Active=%t\n", 
            i+1, record["id"], record["name"], record["age"], record["score"], record["active"])
    }

    fmt.Println("\n多表组合查询示例完成！")
}
```
## 6. 测试结果展示

以下是 `TestTestSelectForJoin` 函数的测试输出结果，展示了不同 SQL 查询条件下的多表查询结果：

### 6.1 测试执行信息

```
=== RUN   TestTestSelectForJoin
```

### 6.2 等值连接查询

```sql
select table1.* from table1,table2 where table1.id=table2.id
```

**查询结果**：
```
map[active:false age:30 id:3 name:Charlie score:75.5]
map[active:true age:35 id:4 name:David score:95]
map[active:false age:40 id:5 name:Eve score:80]
```

**解释**：查询两个表中 `id` 字段相等的记录，返回了 3 条匹配记录。

### 6.3 不等值连接查询

```sql
select table1.* from table1,table2 where table1.id!=table2.id
```

**查询结果**：
```
map[active:true age:20 id:1 name:Alice score:85.5]
map[active:true age:25 id:2 name:Bob score:90]
```

**解释**：查询两个表中 `id` 字段不相等的记录，返回了 2 条匹配记录。

### 6.4 三表等值连接查询

```sql
select table1.* from table1,table2,table3 where table1.id=table2.id and table1.id=table3.id
```

**查询结果**：
```
map[active:false age:40 id:5 name:Eve score:80]
```

**解释**：查询三个表中 `id` 字段都相等的记录，返回了 1 条匹配记录。

### 6.5 三表不等值连接查询

```sql
select table1.* from table1,table2,table3 where table1.id!=table2.id and table1.id!=table3.id
```

**查询结果**：
```
map[active:true age:20 id:1 name:Alice score:85.5]
map[active:true age:25 id:2 name:Bob score:90]
```

**解释**：查询三个表中 `id` 字段都不相等的记录，返回了 2 条匹配记录。

### 6.6 特定ID查询

```sql
select table1.* from table1,table2 where table1.id=4 and table1.id=table2.id
```

**查询结果**：
```
map[active:true age:35 id:4 name:David score:95]
```

**解释**：查询 `id=4` 且在两个表中都存在的记录，返回了 1 条匹配记录。

### 6.7 排除特定ID查询

```sql
select table1.* from table1,table2 where table1.id!=4 and table1.id=table2.id
```

**查询结果**：
```
map[active:false age:30 id:3 name:Charlie score:75.5]
map[active:false age:40 id:5 name:Eve score:80]
```

**解释**：查询 `id!=4` 且在两个表中都存在的记录，返回了 2 条匹配记录。

### 6.8 小于特定值查询

```sql
select table1.* from table1,table2 where table1.id<4 and table1.id=table2.id
```

**查询结果**：
```
map[active:false age:30 id:3 name:Charlie score:75.5]
```

**解释**：查询 `id<4` 且在两个表中都存在的记录，返回了 1 条匹配记录。

### 6.9 小于等于特定值查询

```sql
select table1.* from table1,table2 where table1.id<=4 and table1.id=table2.id
```

**查询结果**：
```
map[active:false age:30 id:3 name:Charlie score:75.5]
map[active:true age:35 id:4 name:David score:95]
```

**解释**：查询 `id<=4` 且在两个表中都存在的记录，返回了 2 条匹配记录。

### 6.10 大于特定值查询

```sql
select table1.* from table1,table2 where table1.id>4 and table1.id=table2.id
```

**查询结果**：
```
map[active:false age:40 id:5 name:Eve score:80]
```

**解释**：查询 `id>4` 且在两个表中都存在的记录，返回了 1 条匹配记录。

### 6.11 大于等于特定值查询

```sql
select table1.* from table1,table2 where table1.id>=4 and table1.id=table2.id
```

**查询结果**：
```
map[active:true age:35 id:4 name:David score:95]
map[active:false age:40 id:5 name:Eve score:80]
```

**解释**：查询 `id>=4` 且在两个表中都存在的记录，返回了 2 条匹配记录。

### 6.12 按年龄查询

```sql
select table1.* from table1,table2 where table1.age=40 and table1.id=table2.id
```

**查询结果**：
```
map[active:false age:40 id:5 name:Eve score:80]
```

**解释**：查询 `age=40` 且在两个表中都存在的记录，返回了 1 条匹配记录。

### 6.13 测试完成信息

```
----Search函数不支持无索引的搜索，如需要支持无索引或自己的匹配策略，可以自定义mach接口实现-----
--- PASS: TestTestSelectForJoin (0.04s)
```

**解释**：测试执行成功，耗时 0.04 秒。同时提示 Search 函数不支持无索引的搜索，如需支持可自定义 match 接口实现。

## 7. 使用场景

### 7.1 数据关联分析

多表组合查询适用于需要在多个表之间进行数据关联和分析的场景，例如：

- 分析用户在不同表中的行为数据
- 关联多个业务系统的数据进行综合分析
- 实现复杂的数据报表和统计功能

### 7.2 查询功能优势

sfsDb 的查询功能具有以下显著优势：

- **灵活性高**：支持多种查询条件和比较操作，包括等值查询、不等值查询、范围查询等
- **性能优秀**：通过索引优化查询性能，支持快速定位和检索数据
- **易用性好**：提供简洁的API接口，降低开发者学习和使用成本
- **扩展性强**：支持任意数量的表组合查询，满足复杂业务需求
- **SQL友好**：支持类SQL语法的查询，降低从传统数据库迁移的成本
- **支持全文搜索**：内置全文索引，支持高效的文本搜索功能
- **支持自定义匹配策略**：允许开发者自定义匹配逻辑，满足特殊查询需求

### 7.3 多表查询优势

sfsDb 的多表组合查询功能具有以下特点：

- **简单直观**：API设计简洁，易于理解和使用
- **高效执行**：通过索引优化，多表查询性能优秀
- **支持多种连接类型**：支持等值连接、不等值连接等多种连接方式
- **支持多表关联**：支持任意数量的表进行关联查询
- **结果易于处理**：查询结果格式统一，易于后续处理和分析

## 8. 总结

sfsDb 的多表组合查询功能为开发者提供了强大的数据查询能力，允许在多个表之间进行灵活的数据关联和分析。通过合理设计表结构和索引，可以实现高效的多表查询操作。

本示例演示了如何创建多个表，插入测试数据，并执行多表组合查询。开发者可以根据实际业务需求，调整查询条件和表结构，实现更复杂的多表查询功能。

## 9. 注意事项

1. 确保在创建表时为常用查询字段添加索引，以提高查询性能
2. 合理设计表结构，避免不必要的数据冗余
3. 在执行大量数据查询时，注意及时释放资源
4. 根据实际业务需求选择合适的查询条件和比较操作

通过以上示例和说明，相信开发者已经掌握了 sfsDb 多表组合查询的基本使用方法，可以在实际项目中灵活应用。