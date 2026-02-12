# 查询数据

## 6.1 基本搜索

```go
// 基本搜索：精确匹配
searchFields := map[string]any{
    "name": "张三",
}
iter, err := table.Search(&searchFields)
if err != nil {
    panic(err)
}
defer iter.Release()

// 获取所有匹配记录
records := iter.GetRecords(true)
defer records.Release()   
for _, record := range records {
    fmt.Printf("找到记录: %v\n", record)
}
```

## 6.2 使用比较操作符

sfsDb 支持多种比较操作符，位于 `util` 包中：

| 操作符 | 描述 | 示例 |
|-------|------|------|
| `Equal` | 等于 | `util.Equal` |
| `NotEqual` | 不等于 | `util.NotEqual` |
| `GreaterThan` | 大于 | `util.GreaterThan` |
| `GreaterThanOrEqual` | 大于等于 | `util.GreaterThanOrEqual` |
| `LessThan` | 小于 | `util.LessThan` |
| `LessThanOrEqual` | 小于等于 | `util.LessThanOrEqual` |
| `Like` | 前缀匹配（类似 SQL LIKE） | `util.Like` |

## 6.3 比较操作符使用示例

```go
import (
    "github.com/liaoran123/sfsDb/util"
)

// 示例：使用比较操作符搜索

// 1. 搜索年龄大于30的用户
fmt.Println("\n年龄大于30的用户:")
ageGt30 := map[string]any{
    "age": 30,
}
iterGt30, err := table.Search(&ageGt30, util.GreaterThan) // 传递比较操作符作为第二个参数
if err != nil {
    panic(err)
}
defer iterGt30.Release()
recordsGt30 := iterGt30.GetRecords(true)
defer recordsGt30.Release()   
for _, record := range recordsGt30 {
    fmt.Printf("   - %s: %d岁\n", record["name"], record["age"])
}

// 2. 前缀搜索（Like操作符）- 默认就是Like操作
fmt.Println("\n邮箱以'user'开头的用户:")
emailPrefix := map[string]any{
    "email": "user",
}
iterPrefix, err := table.Search(&emailPrefix) // 默认使用util.Like操作符，这里的like实则是前缀匹配
if err != nil {
    panic(err)
}
defer iterPrefix.Release()
recordsPrefix := iterPrefix.GetRecords(true)
defer recordsPrefix.Release()      
for _, record := range recordsPrefix {
    fmt.Printf("   - %s: %s\n", record["name"], record["email"])
}

// 3. 显式使用Like操作符
fmt.Println("\n名字以'张'开头的用户:")
namePrefix := map[string]any{
    "name": "张",
}
iterName, err := table.Search(&namePrefix, util.Like) // 显式指定util.Like操作符 
if err != nil {
    panic(err)
}
defer iterName.Release()
recordsName := iterName.GetRecords(true)
defer recordsName.Release()      
for _, record := range recordsName {
    fmt.Printf("   - %s: %s\n", record["name"], record["email"])
}

// 4. 精确匹配搜索
fmt.Println("\n精确搜索名字为'张三'的用户:")
exactSearch := map[string]any{
    "name": "张三",
}
iterExact, err := table.Search(&exactSearch, util.Equal) // 显式指定util.Equal操作符
if err != nil {
    panic(err)
}
defer iterExact.Release()
recordsExact := iterExact.GetRecords(true)
defer recordsExact.Release()   
for _, record := range recordsExact {
    fmt.Printf("   - %s: %s\n", record["name"], record["email"])
}

// 5. 不等于搜索
fmt.Println("\nid不等于1的用户:")
notEqualSearch := map[string]any{
    "id": 1,
}
iterNotEqual, err := table.Search(&notEqualSearch, util.NotEqual) // 使用util.NotEqual操作符
if err != nil {
    panic(err)
}
defer iterNotEqual.Release()
defer GlobalTableIterPool.Put(iterNotEqual)
recordsNotEqual := iterNotEqual.GetRecords(true)
defer record.PutRecords(recordsNotEqual)   
for _, record := range recordsNotEqual {
    fmt.Printf("   - %s: ID=%d\n", record["name"], record["id"])
}
```

## 6.4 匹配器接口（Match Interface）

sfsDb 支持自定义匹配器接口，用于实现复杂的查询逻辑：

```go
// Match 接口定义
type Match interface {
    Match(record map[string]any) bool
}

// 自定义匹配器示例：年龄大于指定值
 type AgeGreaterThanMatcher struct {
    MinAge int
}

func (m *AgeGreaterThanMatcher) Match(record map[string]any) bool {
    if age, ok := record["age"].(int); ok {
        return age > m.MinAge
    }
    return false
}

// 使用自定义匹配器
matcher := &AgeGreaterThanMatcher{MinAge: 30}
records := table.MatchRecords(matcher)
for _, record := range records {
    fmt.Printf("匹配记录: %v\n", record)
}
```

## 6.5 AND 匹配器

sfsDb 提供了内置的 `AND` 匹配器，用于实现类似 SQL 中的 IN、NOT IN、AND、OR 等操作，特别适合多表连接查询。

### 6.5.1 AND 匹配器概述

`AND` 匹配器用于判断记录的字段值是否在指定的数据集合中，支持正向匹配（IN）和反向匹配（NOT IN）。

**核心功能**：
- 实现类似 SQL 的 IN 和 NOT IN 操作
- 支持多字段组合匹配
- 适合多表连接查询
- 支持与其他匹配器组合使用

### 6.5.2 AND 结构体定义

```go
type AND struct {
    fields []string // 需要匹配的字段名
    data   map[any]bool // 匹配数据集合
    rule   bool // 匹配规则：true=IN/AND，false=NOT IN/OR
}
```

**字段说明**：
- `fields`：需要匹配的字段名列表，与记录中的 key 对应
- `data`：匹配数据集合，由 `TableIter.Map()` 方法生成或自定义
- `rule`：匹配规则，`true` 表示 IN/AND，`false` 表示 NOT IN/OR

### 6.5.3 构造函数

```go
func NewAND(fields []string, data map[any]bool, rule ...bool) *AND
```

**参数说明**：
- `fields`：需要匹配的字段名列表
- `data`：匹配数据集合
- `rule`：可选参数，匹配规则，默认为 `true`

### 6.5.4 使用示例

**示例 1：基本 IN 操作**

```go
// 假设我们有一个用户表，需要查询 ID 在指定集合中的用户

// 1. 获取 ID 集合
idMap := map[any]bool{1: true, 3: true, 5: true}

// 2. 创建 AND 匹配器
andMatcher := match.NewAND([]string{"id"}, idMap)

// 3. 使用匹配器
iter, err := table.Search(&map[string]any{"id": nil})
if err != nil {
    panic(err)
}
defer iter.Release()

iter.SetMatch(andMatcher)
records := iter.GetRecords(true)
defer records.Release()   

// 结果：返回 ID 为 1、3、5 的用户
```

**示例 2：NOT IN 操作**

```go
// 查询 ID 不在指定集合中的用户

// 1. 获取 ID 集合
idMap := map[any]bool{1: true, 3: true, 5: true}

// 2. 创建 AND 匹配器，设置 rule=false 表示 NOT IN
andMatcher := match.NewAND([]string{"id"}, idMap, false)

// 3. 使用匹配器
iter, err := table.Search(&map[string]any{"id": nil})
if err != nil {
    panic(err)
}
defer iter.Release()

iter.SetMatch(andMatcher)
records := iter.GetRecords(true)
defer records.Release()   

// 结果：返回 ID 不为 1、3、5 的用户
```

**示例 3：多表连接查询（重点示例）**

```go
// 实现类似 SQL 的连接查询：SELECT table1.* FROM table1, table2 WHERE table1.id = table2.id

// 1. 获取两个表的迭代器
iter1, err := table1.Search(&map[string]any{"id": nil})
if err != nil {
    panic(err)
}
defer GlobalTableIterPool.Put(iter1)

iter2, err := table2.Search(&map[string]any{"id": nil})
if err != nil {
    panic(err)
}
defer iter2.Release()

// 2. 获取 table2 的 ID 映射
// Map() 方法生成 map[any]bool，键为指定字段的值
idMap := iter2.Map()
defer iter2.ReleaseMap(map2)
	/*
		// 如果map2生命周期大于iter2，则使用
		// defer PutMap(map2)
	*/
defer PutMap(map2)

// 3. 创建 AND 匹配器
// 匹配 table1 的 id 字段是否在 table2 的 id 集合中
andMatcher := match.NewAND([]string{"id"}, idMap)

// 4. 设置匹配器并获取结果
iter1.SetMatch(andMatcher)
records := iter1.GetRecords(true)
defer records.Release()   

// 结果：返回 table1 中 ID 与 table2 中 ID 匹配的记录
```

**示例 4：多表连接的 NOT 操作**

```go
// 实现类似 SQL 的连接查询：SELECT table1.* FROM table1, table2 WHERE table1.id != table2.id

// 1. 获取两个表的迭代器
iter1, err := table1.Search(&map[string]any{"id": nil})
if err != nil {
    panic(err)
}
defer iter1.Release()

iter2, err := table2.Search(&map[string]any{"id": nil})
if err != nil {
    panic(err)
}
defer iter2.Release()

// 2. 获取 table2 的 ID 映射
idMap := iter2.Map()
defer iter2.ReleaseMap(map2)
	/*
		// 如果map2生命周期大于iter2，则使用
		// defer PutMap(map2)
	*/
//defer PutMap(map2)

// 3. 创建 AND 匹配器，设置 rule=false 表示 NOT IN
andMatcher := match.NewAND([]string{"id"}, idMap, false)

// 4. 设置匹配器并获取结果
iter1.SetMatch(andMatcher)
records := iter1.GetRecords(true)
defer records.Release()   

// 结果：返回 table1 中 ID 与 table2 中 ID 不匹配的记录
```

### 6.5.5 组合主键匹配

`AND` 匹配器主要用于主键匹配，特别是组合主键的情况。当 `fields` 参数包含多个字段名时，它会通过 `util.MergeFields()` 将这些字段值合并为一个值进行匹配，这正是处理组合主键的典型方式：

```go
// 示例：组合主键匹配，假设表有组合主键 (user_id, product_id)

// 1. 获取另一个表的组合主键映射
// 假设iter2是包含组合主键的表的迭代器
// Map("user_id", "product_id") 生成组合主键值的映射
combinedKeyMap := iter2.Map("user_id", "product_id")

// 2. 创建 AND 匹配器
// 匹配当前表的组合主键 (user_id, product_id) 是否在另一个表的组合主键集合中
andMatcher := match.NewAND([]string{"user_id", "product_id"}, combinedKeyMap)

// 3. 使用匹配器
iter1.SetMatch(andMatcher)
records := iter1.GetRecords(true)
defer record.PutRecords(records)   

// 结果：返回组合主键与另一个表匹配的记录
```

**注意事项**：
- AND 结构体主要设计用于主键值的匹配，特别是多个迭代器之间的主键匹配
- 当 `fields` 参数包含多个字段时，它会将这些字段视为组合主键进行处理
- 合并后的字段值格式取决于 `util.MergeFields()` 函数的实现
- 对于非主键字段的组合匹配，建议使用其他匹配器（如 FieldComparison）或自定义匹配器

**AND 匹配器的设计初衷**：
```go
/*
//该结构作用是多个迭代器的主键值进行相同或不相同的匹配
//rule为true时，多个迭代器的主键值必须相同，才匹配成功
//rule为false时，多个迭代器的主键值必须不同，才匹配成功
//rule为false的作用主要是用在跳跃查询中，比如sql语句中 field not in (1,2,3)，则data=map[any]bool{1:true,2:true,3:true}
*/
```

### 6.5.6 与其他匹配器组合使用

`AND` 匹配器可以与其他匹配器组合使用，实现更复杂的查询逻辑：

```go
// 实现：SELECT * FROM table WHERE id IN (1,3,5) AND age > 25

// 1. 创建 AND 匹配器（ID IN (1,3,5)）
idMap := map[any]bool{1: true, 3: true, 5: true}
idMatcher := match.NewAND([]string{"id"}, idMap)

// 2. 创建 AgeGreaterThanMatcher（age > 25）
ageMatcher := &AgeGreaterThanMatcher{MinAge: 25}

// 3. 使用匹配器
iter1, err := table.Search(&map[string]any{"id": nil})
if err != nil {
    panic(err)
}
defer iter1.Release()

// 设置多个匹配器，它们之间是 AND 关系
iter1.SetMatch(idMatcher, ageMatcher)
records := iter1.GetRecords(true)
defer records.Release()   

// 结果：返回 ID 为 1、3、5 且年龄大于 25 的用户
```

### 6.5.7 AND 匹配器的优势

1. **高效的多表连接**：通过预计算的映射表，避免了嵌套循环，提高了连接查询的效率
2. **灵活的匹配规则**：支持 IN、NOT IN、AND、OR 等多种匹配方式
3. **支持多字段组合**：可以根据多个字段的组合值进行匹配
4. **易于与其他匹配器组合**：可以与其他自定义或内置匹配器组合使用
5. **适合处理复杂查询场景**：特别适合需要关联多个表或集合的查询场景

通过 `AND` 匹配器，sfsDb 实现了高效灵活的多表连接查询功能，为用户提供了强大的数据查询能力。

## 6.6 FieldComparison 匹配器

sfsDb 提供了内置的 `FieldComparison` 匹配器，用于实现各种比较操作，支持将匹配器应用于迭代器，特别适合主键迭代器对无索引字段的匹配：

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/match"
)

func main() {
    // 创建测试表
    table, err := engine.TableNew("test_field_comparison")
    if err != nil {
        panic(err)
    }

    // 设置表字段
    fields := map[string]any{
        "id":     0,
        "name":   "",
        "age":    0,
        "score":  0.0,
        "active": false,
    }

    err = table.SetFields(fields)
    if err != nil {
        panic(err)
    }

    // 创建主键索引
    pkIndex, err := engine.DefaultPrimaryKeyNew("pk")
    if err != nil {
        panic(err)
    }
    pkIndex.AddFields("id")
    err = table.CreateIndex(pkIndex)
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

    for _, record := range testData {
        _, err = table.Insert(&record)
        if err != nil {
            panic(err)
        }
    }

    // FieldComparison 匹配器示例
    fmt.Println("=== FieldComparison 匹配器示例 ===")

    // 1. 使用 FieldComparison 进行比较
    // 获取迭代器
iter, err := table.Search(&map[string]any{"id": nil})
if err != nil {
    panic(err)
}
defer iter.Release()

// 创建 FieldComparison 匹配器
matcher := match.NewFieldComparison("age", match.GreaterThan, 25)

// 将匹配器设置到迭代器上
iter.SetMatch(matcher)

// 获取过滤后的记录
records := iter.GetRecords(true)
defer records.Release()   
    fmt.Printf("年龄大于25的记录 (%d 条):\n", len(records))
    for _, record := range records {
        fmt.Printf("   - %v\n", record)
    }

    // 2. 使用便捷函数创建匹配器
    iter2, err := table.Search(&map[string]any{"id": nil})
if err != nil {
    panic(err)
}
defer iter2.Release()

// 使用 GreaterThanMatch 便捷函数
highScoreMatcher := match.NewGreaterThanMatch("score", 90.0)
iter2.SetMatch(highScoreMatcher)

highScoreRecords := iter2.GetRecords(true)
defer records.Release()   
    fmt.Printf("\n分数大于90的记录 (%d 条):\n", len(highScoreRecords))
    for _, record := range highScoreRecords {
        fmt.Printf("   - %v\n", record)
    }

    // 3. 使用 EqualMatch 便捷函数
    iter3, err := table.Search(&map[string]any{"id": nil})
if err != nil {
    panic(err)
}
defer iter3.Release()

inactiveMatcher := match.NewEqualMatch("active", false)
iter3.SetMatch(inactiveMatcher)

inactiveRecords := iter3.GetRecords(true)
defer records.Release()   
    fmt.Printf("\n非活跃用户 (%d 条):\n", len(inactiveRecords))
    for _, record := range inactiveRecords {
        fmt.Printf("   - %v\n", record)
    }
}
```

## 6.7 FieldComparison 支持的比较操作

| 比较操作 | 描述 | 便捷函数 |
|---------|------|---------|
| `Equal` | 等于 | `NewEqualMatch` |
| `NotEqual` | 不等于 | `NewNotEqualMatch` |
| `GreaterThan` | 大于 | `NewGreaterThanMatch` |
| `GreaterThanOrEqual` | 大于等于 | `NewGreaterThanOrEqualMatch` |
| `LessThan` | 小于 | `NewLessThanMatch` |
| `LessThanOrEqual` | 小于等于 | `NewLessThanOrEqualMatch` |
| `Like` | 前缀匹配 | `NewLikeMatch` |
| `Prefix` | 前缀匹配 | `NewPrefixMatch` |
| `Suffix` | 后缀匹配 | `NewSuffixMatch` |
| `Contains` | 包含匹配 | `NewContainsMatch` |

## 6.8 时序数据区间搜索（SearchRange）

sfsDb 提供了 `SearchRange` 方法，专门用于时序数据的区间搜索，特别适合 IoT 设备产生的时间序列数据查询。

### 6.8.1 方法签名

```go
func (t *Table) SearchRange(funIter storage.FunIter, fieldname string, Start, Limit any) (*TableIter, error)
```

### 6.8.2 参数说明

- `funIter`：迭代器函数，用于创建范围迭代器
- `fieldname`：要搜索的字段名，通常是时间戳字段
- `Start`：范围的起始值
- `Limit`：范围的结束值

### 6.8.3 使用示例

```go
// 时序数据区间搜索示例：IoT设备数据查询
func querySensorDataByTimeRange(startTime, endTime int) error {
    // 创建表
    sensorTable, err := engine.TableNew("sensor_data")
    if err != nil {
        return err
    }
    
    // 设置字段，timestamp作为主键
    fields := map[string]any{
        "timestamp": 0,   // 时间戳，int类型
        "value":     0.0, // 传感器值
        "sensor_id": "",  // 传感器ID
    }
    err = sensorTable.SetFields(fields)
    if err != nil {
        return err
    }
    
    // 创建主键索引
    pk, err := engine.DefaultPrimaryKeyNew("pk")
    if err != nil {
        return err
    }
    pk.AddFields("timestamp")
    err = sensorTable.CreateIndex(pk)
    if err != nil {
        return err
    }
    
    // 插入测试数据（模拟IoT设备产生的时序数据）
    now := int(time.Now().Unix())
    for i := 0; i < 100; i++ {
        data := map[string]any{
            "timestamp": now + i,
            "value":     float64(i * 10),
            "sensor_id": fmt.Sprintf("sensor_%d", i%10),
        }
        _, err := sensorTable.Insert(&data)
        if err != nil {
            return err
        }
    }
    
    // 使用SearchRange进行时间范围查询
    fmt.Println("=== 时序数据区间搜索示例 ===")
    
    // 定义迭代器函数
    funIter := storage.FunIter(func(start, limit []byte) storage.Iterator {
        return sensorTable.kvStore.Iterator(start, limit)
    })
    
    // 执行区间搜索
    iter, err := sensorTable.SearchRange(funIter, "timestamp", startTime, endTime)
    if err != nil {
        return err
    }
    defer iter.Release()
    
    // 获取结果
    records := iter.GetRecords(true)
    defer records.Release()
    
    // 打印结果
    fmt.Printf("时间范围 [%d, %d] 内的传感器数据：\n", startTime, endTime)
    for _, record := range records {
        fmt.Printf("时间戳: %d, 传感器ID: %s, 值: %f\n", 
            record["timestamp"], record["sensor_id"], record["value"])
    }
    
    return nil
}

// 使用示例
func main() {
    // 查询最近10秒的传感器数据
    now := int(time.Now().Unix())
    err := querySensorDataByTimeRange(now-10, now)
    if err != nil {
        panic(err)
    }
}
```

### 6.8.4 时序数据区间搜索的优势

1. **高效的范围查询**：专门针对时序数据的特性优化，提供高效的时间范围查询
2. **灵活的迭代器函数**：允许自定义迭代器函数，适应不同的存储引擎和查询场景
3. **支持事务**：可以在事务中使用，确保数据一致性
4. **适合 IoT 场景**：特别适合处理 IoT 设备产生的海量时序数据
5. **易于集成**：简洁的 API 设计，易于集成到各种应用场景

通过 `SearchRange` 方法，sfsDb 为 IoT 设备提供了高效、可靠的时序数据查询能力，满足了设备对时间序列数据快速检索的需求。