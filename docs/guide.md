# 使用指南

## 快速入门

### 基本使用

#### 1. 初始化表

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
        "name": "",    // 字符串
        "age":  0,     // 整数
        "email": "",   // 字符串
    }
    err = table.SetFields(fields)
    if err != nil {
        panic(err)
    }
    
    // 后续操作...
}
```

#### 2. 插入数据

```go
// 插入记录
user := map[string]any{
    "name": "张三",
    "age":  30,
    "email": "zhangsan@example.com",
}
id, err := table.Insert(&user)
if err != nil {
    panic(err)
}

fmt.Printf("插入成功，ID: %d\n", id)
```

#### 3. 查询数据

```go
// 搜索记录
searchFields := map[string]any{
    "name": "张三",
}
iter := table.Search(&searchFields)
defer iter.Release()

// 遍历搜索结果
for iter.First(); iter.Valid(); iter.Next() {
    record := table.ParseRecordValue(iter.Value())
    fmt.Printf("找到记录: %v\n", record)
}

// 遍历表所有数据
allIter := table.ForData()
defer allIter.Release()

for allIter.First(); allIter.Valid(); allIter.Next() {
    record := table.ParseRecordValue(allIter.Value())
    fmt.Printf("记录: %v\n", record)
}
```

#### 4. 更新数据

```go
// 更新记录
updateFields := map[string]any{
    "id": id,      // 必须包含主键
    "age": 31,     // 要更新的字段
}
err = table.Update(&updateFields)
if err != nil {
    panic(err)
}
```

#### 5. 删除数据

```go
// 删除记录
deleteFields := map[string]any{
    "id": id,      // 必须包含主键
}
err = table.Delete(&deleteFields)
if err != nil {
    panic(err)
}
```

## 高级功能

### 全文索引

#### 创建全文索引

```go
// 创建全文索引
fullTextIndex, err := engine.DefaultFullTextIndexNew("fulltext_idx")
if err != nil {
    panic(err)
}
fullTextIndex.AddFields("content", "id") // 添加字段，最后一个字段必须是主键
fullTextIndex.SetFullField("content", 5) // 设置全文索引字段和长度
err = table.CreateIndex(fullTextIndex)
if err != nil {
    panic(err)
}
```

#### 使用全文搜索

```go
// 插入带全文索引的记录
contentRecord := map[string]any{
    "id": 1,
    "content": "这是一段测试文本，用于全文索引测试",
}
_, err = table.Insert(&contentRecord)
if err != nil {
    panic(err)
}

// 全文搜索
searchContent := map[string]any{
    "content": "测试",
}
iter := table.Search(&searchContent)
defer iter.Release()

for iter.First(); iter.Valid(); iter.Next() {
    record := table.ParseRecordValue(iter.Value())
    fmt.Printf("全文搜索找到: %v\n", record)
}
```

### 普通索引

#### 创建普通索引

```go
// 创建普通索引
normalIndex, err := engine.DefaultNormalIndexNew("name_idx")
if err != nil {
    panic(err)
}
normalIndex.AddFields("name", "id") // 添加字段，最后一个字段必须是主键
err = table.CreateIndex(normalIndex)
if err != nil {
    panic(err)
}
```

#### 使用普通索引搜索

```go
// 使用普通索引搜索
searchFields := map[string]any{
    "name": "张三",
}
iter := table.Search(&searchFields)
defer iter.Release()

for iter.First(); iter.Valid(); iter.Next() {
    record := table.ParseRecordValue(iter.Value())
    fmt.Printf("搜索找到: %v\n", record)
}
```

### 批量操作

#### 批量插入

```go
// 批量插入
for i := 0; i < 100; i++ {
    user := map[string]any{
        "name": fmt.Sprintf("用户%d", i),
        "age":  20 + i%30,
        "email": fmt.Sprintf("user%d@example.com", i),
    }
    _, err := table.Insert(&user)
    if err != nil {
        panic(err)
    }
}
```

### 自定义匹配策略

#### 实现自定义匹配器

```go
// 创建匹配条件
match := engine.NewMatchFunc(func(record *engine.Record) bool {
    // 自定义匹配逻辑，例如：年龄大于25且名字包含"张"
    age, ok1 := (*record)["age"].(int)
    name, ok2 := (*record)["name"].(string)
    return ok1 && ok2 && age > 25 && strings.Contains(name, "张")
})

// 设置匹配条件
iter := table.ForData()
defer iter.Release()
iter.SetMatch(match)

// 遍历，只会返回符合条件的记录
for iter.First(); iter.Valid(); iter.Next() {
    record := table.ParseRecordValue(iter.Value())
    fmt.Printf("匹配记录: %v\n", record)
}
```

## 性能优化

### 索引优化

1. **只为必要字段创建索引**：索引会增加存储和写入开销
2. **合理选择索引类型**：
   - 对于精确匹配，使用普通索引
   - 对于文本搜索，使用全文索引
3. **组合索引**：对于经常一起查询的字段，使用组合索引

### 查询优化

1. **使用索引字段查询**：避免全表扫描
2. **分页获取数据**：对于大量数据，使用迭代器的分页功能
3. **选择必要字段**：使用 `SetSelects` 方法只选择必要的字段
4. **使用跳跃区间**：对于需要跳过特定范围数据的场景，使用跳跃区间提高遍历效率

### 存储优化

1. **定期清理**：删除不再需要的数据
2. **合理设计表结构**：根据实际需求设计字段和数据类型
3. **监控数据库大小**：及时备份和清理数据

## 最佳实践

1. **错误处理**：始终检查并适当处理错误
2. **资源管理**：使用 defer 确保迭代器资源被释放
3. **事务管理**：对于需要原子性的操作，使用批量操作
4. **日志记录**：记录关键操作和错误
5. **备份策略**：定期备份数据库文件

## 常见问题

### 问题：插入数据失败

**可能原因**：
- 字段名包含分隔符
- 字段不存在于表中
- 字段类型不匹配
- 索引创建失败

**解决方案**：
- 确保字段名不包含分隔符"-"
- 确保所有字段都已在 `SetFields` 中定义
- 确保字段类型与 `SetFields` 中定义的类型一致
- 查看详细错误信息

### 问题：查询性能慢

**可能原因**：
- 未使用索引
- 数据量过大
- 查询条件过于复杂

**解决方案**：
- 为查询字段创建索引
- 使用分页功能减少返回数据量
- 优化查询条件，使用索引字段进行查询
- 考虑使用全文索引进行文本搜索

### 问题：数据库文件过大

**可能原因**：
- 数据量增长
- 频繁的插入/删除操作
- 未优化的存储

**解决方案**：
- 定期清理不必要的数据
- 考虑数据分片策略
- 监控并优化存储使用
- 合理设计表结构和索引

## 高级示例

### 分页查询

```go
// 获取迭代器
iter := table.ForData()
defer iter.Release()

// 获取第2页，每页10条记录
records := iter.GetRecords(true, 10, 10)
for _, record := range records {
    fmt.Printf("记录: %v\n", record)
}
```

### 批量更新

```go
// 搜索条件
searchFields := map[string]any{
    "status": "inactive",
}
iter := table.Search(&searchFields)
defer iter.Release()

// 批量更新
updateData := map[string]any{
    "status": "active",
}
iter.Update(&updateData)
```

### 批量删除

```go
// 搜索条件
searchFields := map[string]any{
    "age": 0,
}
iter := table.Search(&searchFields)
defer iter.Release()

// 批量删除
iter.Delete()
```

### 使用记录集合操作

```go
// 获取两个查询的结果
iter1 := table.Search(&map[string]any{"age": 30})
defer iter1.Release()
records1 := iter1.GetRecords(true)

iter2 := table.Search(&map[string]any{"name": "张三"})
defer iter2.Release()
records2 := iter2.GetRecords(true)

// 计算交集
intersection := records1.Intersect(records2)
fmt.Printf("交集: %v\n", intersection)

// 计算并集
union := records1.Union(records2)
fmt.Printf("并集: %v\n", union)

// 计算差集
difference := records1.Difference(records2)
fmt.Printf("差集: %v\n", difference)
```

### 使用跳跃区间

```go
// 创建跳跃区间迭代器（示例）
// 实际使用中需要根据具体情况创建跳跃区间
// jumpRange := createJumpRangeIterator("status", "deleted")

// 设置跳跃区间
// iter.SetJumpRanges(jumpRange)

// 遍历，会自动跳过指定区间的记录
for iter.First(); iter.Valid(); iter.Next() {
    record := table.ParseRecordValue(iter.Value())
    fmt.Printf("记录: %v\n", record)
}
```