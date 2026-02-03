# 字段修改

## 1. 字段修改概述

sfsDb 支持动态修改表字段，包括添加新字段、修改字段名称和更新字段类型。字段修改操作需要遵循特定的工作流程，以确保数据一致性和系统稳定性。

## 2. 单个字段修改

### 2.1 字段重命名

要修改字段名称，需要按照以下步骤操作：

```go
// 字段修改工作流：将 "position" 字段重命名为 "job_title"

// 1. 首先调用 UpdateFieldName 更新字段名称映射
err = table.UpdateFieldName("position", "job_title")
if err != nil {
    panic(err)
}

// 2. 然后调用 SetFields 更新字段映射
updatedFields := map[string]any{
    "id":        0,
    "name":      "",
    "age":       0,
    "email":     "",
    "job_title": "", // 使用新的字段名
}
err = table.SetFields(updatedFields)
if err != nil {
    panic(err)
}
```

### 2.2 添加新字段

要添加新字段，只需调用 `SetFields` 方法并包含新字段：

```go
// 添加新字段 "department"
updatedFields := map[string]any{
    "id":         0,
    "name":       "",
    "age":        0,
    "email":      "",
    "job_title":  "",
    "department": "", // 新字段
}
err = table.SetFields(updatedFields)
if err != nil {
    panic(err)
}
```

## 3. 批量字段修改

对于需要同时修改多个字段的情况，可以按照以下步骤操作：

```go
// 批量修改多个字段

// 1. 依次调用 UpdateFieldName 更新多个字段名称映射
err = table.UpdateFieldName("old_field1", "new_field1")
if err != nil {
    panic(err)
}

// 2. 再次更新另一个字段名称
err = table.UpdateFieldName("old_field2", "new_field2")
if err != nil {
    panic(err)
}

// 3. 最后更新字段映射
newFields := map[string]any{
    "id":         0,
    "name":       "",
    "new_field1": "", // 使用新的字段名
    "new_field2": 0,   // 使用新的字段名
    "new_field3": 0.0, // 新添加的字段
}
err = table.SetFields(newFields)
if err != nil {
    panic(err)
}
```

## 4. 字段修改最佳实践

### 4.1 注意事项

1. **数据兼容性**：修改字段类型时，需要确保现有数据能够兼容新类型
2. **索引影响**：修改被索引的字段可能需要重新创建索引
3. **批量操作**：批量修改字段时，应确保所有操作都成功执行
4. **备份数据**：在进行大量字段修改前，建议备份数据

### 4.2 推荐做法

**推荐流程**：

1. **规划修改**：明确需要修改的字段和修改方式
2. **备份数据**：确保数据安全
3. **执行修改**：按照正确的顺序执行字段修改操作
4. **验证结果**：确认修改后的表结构和数据正确
5. **更新索引**：如果需要，重新创建相关索引

## 5. 示例：完整的字段修改流程

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
    _, err := storage.OpenDefaultDb("./field_modification_db")
    if err != nil {
        panic(err)
    }
    defer storage.CloseDb()

    // 创建表
    table, err := engine.TableNew("employees")
    if err != nil {
        panic(err)
    }

    // 设置初始字段
    initialFields := map[string]any{
        "id":       0,
        "name":     "",
        "age":      0,
        "position": "", // 初始字段名
        "salary":   0.0,
    }
    err = table.SetFields(initialFields)
    if err != nil {
        panic(err)
    }

    // 插入测试数据
    employee := map[string]any{
        "name":     "张三",
        "age":      30,
        "position": "工程师",
        "salary":   10000.0,
    }
    _, err = table.Insert(&employee)
    if err != nil {
        panic(err)
    }

    fmt.Println("初始表结构和数据创建完成")

    // 字段修改流程
    fmt.Println("\n=== 字段修改流程 ===")

    // 1. 更新字段名称
    fmt.Println("1. 将 'position' 字段重命名为 'job_title'")
    err = table.UpdateFieldName("position", "job_title")
    if err != nil {
        panic(err)
    }
    fmt.Println("字段名称更新成功")

    // 2. 更新字段映射，添加新字段
    fmt.Println("2. 更新字段映射，添加 'department' 字段")
    updatedFields := map[string]any{
        "id":         0,
        "name":       "",
        "age":        0,
        "job_title":  "", // 使用新的字段名
        "salary":     0.0,
        "department": "", // 新添加的字段
    }
    err = table.SetFields(updatedFields)
    if err != nil {
        panic(err)
    }
    fmt.Println("字段映射更新成功")

    // 3. 验证修改结果
    fmt.Println("\n3. 验证修改结果")
    
    // 查询修改后的数据
    searchFields := map[string]any{"id": 1}
    iter, _ := table.Search(&searchFields)
    defer engine.GlobalTableIterPool.Put(iter)
    
    records := iter.GetRecords(true)
    defer record.PutRecords(records)
    
    for _, record := range records {
        fmt.Printf("修改后的记录: %v\n", record)
        
        // 检查新字段是否存在
        if _, exists := record["job_title"]; exists {
            fmt.Println("✓ 字段重命名成功: position → job_title")
        }
        
        if _, exists := record["department"]; exists {
            fmt.Println("✓ 新字段添加成功: department")
        }
    }

    fmt.Println("\n字段修改流程完成")
}
```

## 6. 字段修改常见问题

### 6.1 错误处理

**常见错误**：

1. **字段不存在**：尝试修改不存在的字段
2. **索引冲突**：修改被索引的字段时没有重新创建索引
3. **数据类型不兼容**：修改字段类型时，现有数据无法转换为新类型

**解决方案**：

1. **验证字段存在性**：在修改前检查字段是否存在
2. **重新创建索引**：修改被索引的字段后，重新创建相关索引
3. **数据迁移**：对于类型不兼容的情况，考虑先导出数据，修改字段后重新导入

### 6.2 性能考虑

- **批量修改**：对于多个字段的修改，建议使用批量操作
- **避免频繁修改**：频繁修改字段会影响性能，应在设计阶段尽量确定字段结构
- **索引重建**：修改被索引的字段后，及时重建索引以保持查询性能

## 7. 总结

字段修改是 sfsDb 的重要功能，支持系统的动态演进和需求变更。通过遵循正确的工作流程和最佳实践，可以安全、高效地完成字段修改操作，确保系统的稳定性和数据的一致性。

**核心要点**：

1. **遵循工作流程**：先更新字段名称映射，再更新字段映射
2. **注意数据兼容**：确保修改不会导致数据丢失或类型错误
3. **更新相关索引**：修改被索引的字段后，重新创建相关索引
4. **备份数据**：在进行重大修改前，备份数据以防止意外情况

通过合理使用字段修改功能，可以使 sfsDb 表结构更好地适应业务需求的变化，提高系统的灵活性和可维护性。