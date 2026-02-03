# 修改记录

## 1. 修改记录概述

sfsDb 支持修改现有记录的功能，通过 `Update` 方法可以更新记录中的字段值。修改记录操作需要提供主键字段来定位要修改的记录，然后指定要更新的字段及其新值。

## 2. 基本修改操作

### 2.1 简单修改

使用 `Update` 方法修改记录的基本语法如下：

```go
// 修改记录
updateData := map[string]any{
    "id":   1,        // 主键字段，用于定位记录
    "name": "李四",    // 要更新的字段
    "age":  35,        // 要更新的字段
}

err = table.Update(&updateData)
if err != nil {
    panic(err)
}
```

### 2.2 批量修改

在批量操作中，也可以使用 `Update` 方法：

```go
// 获取批量操作对象
batch := storage.KVDb.GetBatch()
if batch == nil {
    panic("无法获取批量操作对象")
}

// 批量修改多个记录
for i := 1; i <= 5; i++ {
    updateData := map[string]any{
        "id":   i,
        "age":  20 + i, // 更新为新值
        "email": fmt.Sprintf("user%d@example.com", i),
    }
    
    err = table.Update(&updateData, batch)
    if err != nil {
        panic(err)
    }
}

// 提交批量操作
err = storage.KVDb.WriteBatch(batch)
if err != nil {
    panic(fmt.Sprintf("批量提交失败: %v", err))
}
```

### 2.3 迭代器批量修改

`sfsDb` 还提供了通过迭代器进行批量修改的功能，使用 `TableIter.Update` 方法可以批量更新查询结果中的记录：

```go
// 创建查询条件
searchFields := map[string]any{"age": map[string]any{"$gte": 25}}

// 获取迭代器
iter, err := table.Search(&searchFields)
if err != nil {
    panic(err)
}
defer engine.GlobalTableIterPool.Put(iter)

// 准备更新数据
updateData := map[string]any{
    "status": "active", // 要更新的字段
}

// 执行批量更新，限制最多更新100条记录
err = iter.Update(&updateData, 100)
if err != nil {
    panic(err)
}

fmt.Println("批量更新完成")
```

#### 迭代器批量修改的工作原理

`TableIter.Update` 方法的内部实现如下：

1. **批量大小控制**：默认每1000条记录执行一次批量写入
2. **自动分批**：当达到批量大小限制时，自动执行写入并创建新的批量操作
3. **主键检查**：确保每条记录都有主键字段
4. **错误处理**：在任何步骤失败时都会停止并返回错误
5. **资源管理**：确保在函数结束时执行所有待处理的批量操作

#### 迭代器批量修改的优势

- **简化代码**：不需要手动遍历记录和管理批量操作
- **自动优化**：内置批量大小控制，平衡内存使用和性能
- **灵活限制**：可以指定最大更新记录数
- **错误处理**：统一的错误处理机制
- **性能优异**：减少网络往返和磁盘操作，提高批量更新性能

## 3. 修改记录的工作原理

### 3.1 内部实现

`Update` 方法的内部实现流程如下：

1. **主键检查**：验证是否提供了所有主键字段
2. **类型检查**：检查字段类型是否匹配
3. **记录读取**：根据主键读取要修改的记录
4. **字段处理**：确定要更新的字段（排除主键字段）
5. **记录更新**：更新记录中的字段值
6. **索引更新**：更新相关索引

### 3.2 关键代码

```go
func (t *Table) Update(fields *map[string]any, batchs ...storage.Batch) error {
    // 检查是否提供了所有主键字段
    for _, field := range t.GetPrimaryFields() {
        if _, ok := (*fields)[field]; !ok {
            return fmt.Errorf("必须提供主键字段 '%s'", field)
        }
    }
    
    // 检查字段类型是否匹配
    if err := t.CheckType(fields); err != nil {
        return err
    }
    
    // 读取记录
    record, err := t.Read(fields)
    if err != nil {
        return err
    }
    if record == nil {
        return fmt.Errorf("主键值 '%v' 的记录不存在", fields)
    }
    
    // 处理更新字段
    updateFields := GetStringSlice()
    defer PutStringSlice(updateFields)
    for field := range *fields {
        // 排除主键字段
        if t.GetPrimaryKey().MatchFields(field) {
            continue
        }
        updateFields = append(updateFields, field)
    }
    
    // 检查是否有字段需要更新
    if len(updateFields) == 0 {
        return nil
    }
    
    // 解析记录并更新
    pk := t.GetPrimaryKey()
    fieldsBytes, err := pk.Parse(t.fieldsid, record)
    if err != nil {
        return err
    }
    
    // 执行更新操作
    BatchContainer := NewBatchContainer(batch, t.indexs, t.id, t.kvStore)
    BatchContainer.Operation(fieldsBytes, updateFields...)
    
    // ... 后续代码
}
```

## 4. 修改记录的最佳实践

### 4.1 注意事项

1. **必须提供主键**：修改记录时必须提供所有主键字段，否则会返回错误
2. **字段类型匹配**：更新的字段值类型必须与表定义的类型匹配
3. **记录存在性**：如果根据主键找不到记录，会返回错误
4. **索引更新**：修改被索引的字段时，系统会自动更新相关索引
5. **批量操作**：对于多个记录的修改，建议使用批量操作以提高性能

### 4.2 推荐做法

**推荐流程**：

1. **验证记录存在**：在修改前可以先检查记录是否存在
2. **准备更新数据**：构建包含主键和要更新字段的映射
3. **执行修改操作**：调用 `Update` 方法执行修改
4. **处理错误**：妥善处理可能的错误
5. **验证结果**：修改后可以查询记录验证修改是否成功

## 5. 示例：完整的修改记录流程

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/record"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // 初始化数据库
    _, err := storage.OpenDefaultDb("./update_demo_db")
    if err != nil {
        panic(err)
    }
    defer storage.CloseDb()

    // 创建表
    table, err := engine.TableNew("users")
    if err != nil {
        panic(err)
    }

    // 设置字段
    fields := map[string]any{
        "id":    0,
        "name":  "",
        "age":   0,
        "email": "",
    }
    err = table.SetFields(fields)
    if err != nil {
        panic(err)
    }

    // 插入测试数据
    user := map[string]any{
        "name":  "张三",
        "age":   30,
        "email": "zhangsan@example.com",
    }
    id, err := table.Insert(&user)
    if err != nil {
        panic(err)
    }
    fmt.Printf("插入成功，ID: %d\n", id)

    // 修改记录
    fmt.Println("\n=== 修改记录 ===")
    
    updateData := map[string]any{
        "id":    id,        // 主键字段
        "name":  "张三改",    // 更新姓名
        "age":   31,        // 更新年龄
        "email": "zhangsan_updated@example.com", // 更新邮箱
    }
    
    err = table.Update(&updateData)
    if err != nil {
        panic(err)
    }
    fmt.Println("记录修改成功")

    // 验证修改结果
    fmt.Println("\n=== 验证修改结果 ===")
    
    searchData := map[string]any{"id": id}
iter, err := table.Search(&searchData)
if err != nil {
    panic(err)
}
defer engine.GlobalTableIterPool.Put(iter)
    
    records := iter.GetRecords(true)
    defer record.PutRecords(records)
    
    for _, r := range records {
        fmt.Printf("修改后的记录: %v\n", r)
        
        // 检查字段是否已更新
        if r["name"] == "张三改" {
            fmt.Println("✓ 姓名更新成功")
        }
        if r["age"] == 31 {
            fmt.Println("✓ 年龄更新成功")
        }
        if r["email"] == "zhangsan_updated@example.com" {
            fmt.Println("✓ 邮箱更新成功")
        }
    }

    fmt.Println("\n修改记录流程完成")
}
```

## 6. 常见问题与解决方案

### 6.1 错误处理

**常见错误**：

1. **缺少主键字段**：修改时未提供所有主键字段
2. **记录不存在**：根据主键找不到对应的记录
3. **字段类型不匹配**：更新的字段值类型与表定义不匹配
4. **批量操作失败**：批量修改时部分操作失败

**解决方案**：

1. **确保提供主键**：修改前确认包含所有主键字段
2. **检查记录存在**：可以先查询记录是否存在
3. **验证字段类型**：确保更新的值类型正确
4. **使用事务**：对于重要操作，使用事务确保原子性

### 6.2 性能优化

- **批量操作**：对于多个记录的修改，使用批量操作
- **只更新必要字段**：只包含需要更新的字段，减少数据传输
- **避免频繁修改**：频繁修改会影响性能，应合理设计数据结构
- **索引考虑**：修改被索引的字段会更新索引，可能影响性能

## 7. 总结

修改记录是 sfsDb 的重要功能，通过 `Update` 方法可以方便地更新现有记录的字段值。使用时需要注意提供主键字段、确保字段类型匹配，并合理使用批量操作以提高性能。

**核心要点**：

1. **必须提供主键**：修改记录时必须包含所有主键字段
2. **字段类型匹配**：更新的值类型必须与表定义一致
3. **批量操作优化**：对于多个记录的修改，使用批量操作
4. **错误处理**：妥善处理可能的错误情况
5. **性能考虑**：合理设计修改操作，避免频繁修改被索引的字段

通过正确使用修改记录功能，可以有效地维护数据的准确性和一致性，确保系统的数据状态与业务需求保持同步。