# 跳跃区间（Jump Ranges）

跳跃区间是 sfsDb 中的一项高级功能，用于在迭代过程中跳过特定范围的数据，提高查询效率。本文档将详细介绍跳跃区间的实现原理、使用方法和应用场景。

## 1. 概述

跳跃区间允许在遍历数据时跳过预定义的范围，特别适用于以下场景：

- **排除特定类别数据**：如电商系统中排除特定类别的商品
- **跳过无效数据**：如 IoT 系统中跳过异常传感器数据
- **时间范围排除**：如金融系统中跳过特定时间段的交易
- **任何需要范围排除的查询场景**

## 2. 实现原理

### 2.1 核心数据结构

跳跃区间的核心实现位于 `TableIter` 结构体中：

```go
type TableIter struct {
    iter       storage.Iterator
    jumpRanges []storage.Iterator // 跳跃区间
    table      *Table
    match      []match.Match // 匹配规则
    selects    []string
    index      Index // 搜索时使用的索引
    move       map[bool]func() bool
    top        map[bool]func() bool
    mu         sync.Mutex
}
```

### 2.2 核心方法

#### 2.2.1 设置跳跃区间

```go
func (t *TableIter) SetJumpRanges(jumpRanges ...storage.Iterator) {
    t.jumpRanges = jumpRanges
}
```

#### 2.2.2 跳跃区间检测

```go
func (t *TableIter) JumpRange(key []byte, jumpRanges []storage.Iterator, esc bool) []byte {
    if len(jumpRanges) == 0 {
        return nil
    }
    for _, jumpRange := range jumpRanges {
        if esc {
            jumpRange.First()
        } else {
            jumpRange.Last()
        }
        if bytes.Equal(jumpRange.Key(), key) {
            if esc {
                // 顺序时，需要判断是否是最后一个元素
                if jumpRange.Last() {
                    return jumpRange.Key()
                }
            } else {
                // 倒序时，需要判断是否是第一个元素
                if jumpRange.First() {
                    return jumpRange.Key()
                }
            }
        }
    }
    return nil
}
```

#### 2.2.3 迭代过程中的跳跃

在 `ExportRecord` 方法中，当遍历到每个键值对时，会检查是否有跳跃区间：

```go
func (t *TableIter) ExportRecord(export ExportRecord, esc bool, limit ...int) {
    // 添加锁，防止并发访问
    t.mu.Lock()
    defer t.mu.Unlock()
    if !t.top[esc]() {
        return
    }
    var rd record.Record
    isMatch := true
    matchlen := 0
    page := PageNew(limit...)
    count := 0
    loop := 0
    var end []byte
    var key, value []byte
    for {
        key = t.iter.Key()
        value = t.iter.Value()
        if len(t.jumpRanges) > 0 { // 如果有跳跃区间，需要跳过某些数据
            end = t.JumpRange(key, t.jumpRanges, esc)
            if end != nil {
                // 跳跃到区间的结束位置
                t.iter.Seek(end)
                if !t.move[esc]() {
                    break
                }
            }
        }
        // 后续处理...
    }
}
```

## 3. 与 RangeForAny 的关联

`RangeForAny` 函数是实现跳跃区间的基础，它用于创建表示需要跳过的区间的迭代器：

```go
func (t *Table) RangeForAny(funIter storage.FunIter, fieldname string, Start, Limit any) (storage.Iterator, Index, error) {
    if funIter == nil {
        funIter = t.kvStore.Iterator
    }
    idx := t.MatchIndexCached([]string{fieldname})
    if idx == nil {
        return nil, nil, fmt.Errorf("字段 '%s' 不存在于表 '%s'", fieldname, t.name)
    }
    pfx := idx.Prefix(t.id)
    pfx = append(pfx, SPLIT[0])
    // 处理 Start 参数
    var startBytes []byte
    if Start != nil {
        startBytes = util.AnyToBytes(Start)
    }
    // 处理 Limit 参数
    var limitBytes []byte
    if Limit != nil {
        limitBytes = util.AnyToBytes(Limit)
    }
    // 创建范围对象
    slice := &util.Range{
        Start: startBytes,
        Limit: limitBytes,
    }
    // 构建完整的搜索范围
    slice.Start = append(pfx, slice.Start...)
    if Limit == nil {
        // 当 Limit 为 nil 时，使用前缀的下一个字节作为上限，表示到无穷大
        slice.Limit = util.BytesPrefix(pfx).Limit
    } else {
        // 当 Limit 不为 nil 时，构建完整的上限字节
        slice.Limit = append(pfx, slice.Limit...)
    }
    iter := funIter(slice.Start, slice.Limit)
    if iter == nil {
        return nil, nil, fmt.Errorf("区间迭代器不能为空")
    }
    return iter, idx, nil
}
```

## 4. 使用方法

### 4.1 基本使用步骤

1. **创建跳跃区间迭代器**：使用 `RangeForAny` 函数创建表示需要跳过的区间的迭代器
2. **设置跳跃区间**：调用 `TableIter` 的 `SetJumpRanges` 方法设置跳跃区间
3. **执行迭代**：调用 `GetRecordSet` 或其他迭代方法，迭代过程中会自动跳过跳跃区间内的数据

### 4.2 代码示例

#### 4.2.1 电商系统示例：排除电子产品

```go
// 创建表
table, err := engine.TableNew("products")
if err != nil {
    panic(err)
}

// 设置字段
fields := map[string]any{
    "id":       0,
    "name":     "",
    "category": "",
    "price":    0.0,
}
err = table.SetFields(fields)
if err != nil {
    panic(err)
}

// 创建索引
pk, err := engine.DefaultPrimaryKeyNew("pk")
pk.AddFields("id")
err = table.CreateIndex(pk)

// 创建类别索引
categoryIndex, err := engine.DefaultNormalIndexNew("idx_category")
categoryIndex.AddFields("category")
err = table.CreateIndex(categoryIndex)

// 插入测试数据
// ... 插入数据代码 ...

// 创建跳跃区间：电子产品范围
electronicsIter, _, _ := table.RangeForAny(nil, "category", "electronics", "electronics")

// 创建主查询迭代器：所有商品
allProductsIter, _, _ := table.SearchRange(nil, "category", nil, nil)

// 设置跳跃区间
allProductsIter.SetJumpRanges(electronicsIter)

// 执行查询
result := allProductsIter.GetRecordSet(true)

// 处理结果
for _, record := range result {
    fmt.Printf("商品: %s, 类别: %s, 价格: %.2f\n", 
        record["name"], record["category"], record["price"])
}
```

#### 4.2.2 IoT 系统示例：跳过异常数据

```go
// 创建表
sensorTable, err := engine.TableNew("sensor_data")
if err != nil {
    panic(err)
}

// 设置字段
fields := map[string]any{
    "timestamp": 0,
    "sensor_id": "",
    "value":     0.0,
    "status":    "",
}
err = sensorTable.SetFields(fields)
if err != nil {
    panic(err)
}

// 创建索引
// ... 创建索引代码 ...

// 插入测试数据
// ... 插入数据代码 ...

// 创建跳跃区间：异常数据范围
errorIter, _, _ := sensorTable.RangeForAny(nil, "status", "error", "error")

// 创建主查询迭代器：所有传感器数据
allDataIter, _, _ := sensorTable.SearchRange(nil, "timestamp", nil, nil)

// 设置跳跃区间
allDataIter.SetJumpRanges(errorIter)

// 执行查询
result := allDataIter.GetRecordSet(true)

// 处理结果
for _, record := range result {
    fmt.Printf("时间戳: %d, 传感器: %s, 值: %.2f, 状态: %s\n", 
        record["timestamp"], record["sensor_id"], record["value"], record["status"])
}
```

#### 4.2.3 金融系统示例：跳过特定时间段交易

```go
// 创建表
transactionTable, err := engine.TableNew("transactions")
if err != nil {
    panic(err)
}

// 设置字段
fields := map[string]any{
    "id":        0,
    "timestamp": 0,
    "amount":    0.0,
    "type":      "",
}
err = transactionTable.SetFields(fields)
if err != nil {
    panic(err)
}

// 创建索引
// ... 创建索引代码 ...

// 插入测试数据
// ... 插入数据代码 ...

// 创建跳跃区间：2023年12月1日的交易
startTime := 1672531200 // 2023-12-01 00:00:00
endTime := 1672617599   // 2023-12-01 23:59:59
december1Iter, _, _ := transactionTable.RangeForAny(nil, "timestamp", startTime, endTime)

// 创建主查询迭代器：所有交易
allTransactionsIter, _, _ := transactionTable.SearchRange(nil, "timestamp", nil, nil)

// 设置跳跃区间
allTransactionsIter.SetJumpRanges(december1Iter)

// 执行查询
result := allTransactionsIter.GetRecordSet(true)

// 处理结果
for _, record := range result {
    fmt.Printf("交易ID: %d, 时间戳: %d, 金额: %.2f, 类型: %s\n", 
        record["id"], record["timestamp"], record["amount"], record["type"])
}
```

## 5. 性能优化

### 5.1 最佳实践

1. **合理设置跳跃区间**：只跳过确实需要排除的数据范围，避免过度使用跳跃区间
2. **使用合适的索引**：为跳跃区间字段创建索引，提高区间创建速度
3. **批量操作**：对于需要跳过多个区间的场景，一次性设置所有跳跃区间
4. **及时释放资源**：使用完毕后调用 `GlobalTableIterPool.Put(iter)` 释放迭代器资源

### 5.2 性能对比

使用跳跃区间可以显著提高查询效率，特别是在以下场景：

- **大量数据需要排除**：如排除 80% 以上的数据时，性能提升明显
- **多次重复查询**：对于重复执行的查询，跳跃区间可以避免重复过滤
- **复杂过滤条件**：当过滤条件复杂时，跳跃区间比传统的 `Match` 条件更高效

## 6. 注意事项

1. **跳跃区间的正确性**：确保跳跃区间迭代器正确表示需要跳过的范围
2. **索引依赖**：跳跃区间依赖于字段的索引，确保为跳跃区间字段创建了合适的索引
3. **内存使用**：多个跳跃区间会增加内存使用，注意控制跳跃区间的数量
4. **迭代器释放**：使用完毕后及时释放迭代器资源，避免内存泄漏

## 7. 总结

跳跃区间是 sfsDb 中一项强大的高级功能，通过 `RangeForAny` 创建的区间迭代器，可以在查询过程中跳过不需要的数据，大大提高了查询效率。它特别适用于需要排除特定范围数据的场景，如电商系统、IoT 系统和金融系统等。

合理使用跳跃区间，可以显著提升查询性能，减少不必要的数据处理，使 sfsDb 在处理复杂查询场景时更加高效。