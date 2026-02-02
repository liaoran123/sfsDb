# 事务优化

## 1. 事务概述

sfsDb 支持事务操作，用于确保数据一致性和可靠性。事务是一组操作的集合，这些操作要么全部成功执行，要么全部失败回滚，保证了数据的完整性。

## 2. 事务类型

sfsDb 支持两种主要的事务操作方式：

### 2.1 自动事务

默认情况下，sfsDb 的每个操作（如插入、更新、删除）都会自动作为一个单独的事务执行：

```go
// 自动事务：每个操作单独作为一个事务
_, err := table.Insert(&user) // 自动事务
if err != nil {
    panic(err)
}

err = table.Delete(&deleteUser) // 自动事务
if err != nil {
    panic(err)
}
```

### 2.2 手动事务（批量操作）

对于需要原子性执行的多个操作，sfsDb 提供了手动事务机制，通过批处理（batch）实现：

```go
// 1. 获取批量操作对象
batch := storage.KVDb.GetBatch()
if batch == nil {
    panic("无法获取批量操作对象")
}

// 2. 添加多个操作到批处理
_, err := table.Insert(&user1, batch)
if err != nil {
    panic(err)
}

_, err = table.Insert(&user2, batch)
if err != nil {
    panic(err)
}

// 3. 手动提交批处理（所有操作一次性执行）
err = storage.KVDb.WriteBatch(batch)
if err != nil {
    panic(fmt.Sprintf("批量提交失败: %v", err))
}
```

## 3. 手动事务最佳实践

### 3.1 基本流程

**推荐做法**：

```go
// 1. 获取批处理对象
batch := storage.KVDb.GetBatch()
if batch == nil {
    panic("无法获取批量操作对象")
}

// 2. 添加操作到批处理
// 添加插入操作
_, err := table.Insert(&user1, batch)
if err != nil {
    panic(err)
}

// 添加更新操作
_, err = table.Update(&updateData, batch)
if err != nil {
    panic(err)
}

// 添加删除操作
err = table.Delete(&deleteData, batch)
if err != nil {
    panic(err)
}

// 3. 提交批处理
err = storage.KVDb.WriteBatch(batch)
if err != nil {
    panic(fmt.Sprintf("批量提交失败: %v", err))
}
```

### 3.2 错误处理

在手动事务中，应妥善处理错误，确保在任何步骤失败时能够及时发现：

```go
// 错误处理示例
batch := storage.KVDb.GetBatch()
if batch == nil {
    return fmt.Errorf("无法获取批量操作对象")
}

// 添加操作
if _, err := table.Insert(&user1, batch); err != nil {
    return fmt.Errorf("添加插入操作失败: %w", err)
}

if _, err := table.Insert(&user2, batch); err != nil {
    return fmt.Errorf("添加插入操作失败: %w", err)
}

// 提交操作
if err := storage.KVDb.WriteBatch(batch); err != nil {
    return fmt.Errorf("批量提交失败: %w", err)
}

return nil
```

### 3.3 组合操作

手动事务支持多种操作的组合，如插入+删除、更新+插入等：

```go
// 组合操作示例
batch := storage.KVDb.GetBatch()
if batch == nil {
    panic("无法获取批量操作对象")
}

// 添加一条新记录
newUser := map[string]any{
    "name": "测试用户",
    "age":  25,
    "email": "test@example.com",
}
_, err = table.Insert(&newUser, batch)
if err != nil {
    panic(err)
}

// 删除一条现有记录
deleteUser := map[string]any{
    "id": 1, // 删除ID为1的记录
}
err = table.Delete(&deleteUser, batch)
if err != nil {
    panic(err)
}

// 提交组合操作
err = storage.KVDb.WriteBatch(batch)
if err != nil {
    panic(fmt.Sprintf("组合操作提交失败: %v", err))
}
```

## 4. 事务性能优化

### 4.1 批量操作的性能优势

使用手动事务（批量操作）可以显著提高性能，特别是在处理大量数据时：

| 操作类型 | 自动事务 | 手动事务 | 性能提升 |
|---------|---------|---------|---------|
| 插入100条记录 | 100次磁盘操作 | 1次磁盘操作 | 90%+ |
| 更新50条记录 | 50次磁盘操作 | 1次磁盘操作 | 80%+ |
| 混合操作（20插入+20删除） | 40次磁盘操作 | 1次磁盘操作 | 85%+ |

### 4.2 批量大小优化

批量操作的大小应根据实际情况进行调整：

- **小批量**：适合内存有限的系统，每次处理 100-500 条记录
- **中批量**：适合一般系统，每次处理 500-2000 条记录
- **大批量**：适合内存充足的系统，每次处理 2000-5000 条记录

### 4.3 并发事务

在高并发场景下，应合理控制事务的粒度和并发度：

- **减小事务粒度**：将大事务拆分为多个小事务
- **避免长事务**：尽量缩短事务持有时间
- **合理使用锁**：只对必要的数据加锁

## 5. 示例：批量数据导入

```go
package main

import (
    "fmt"
    "time"
    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // 初始化数据库
    _, err := storage.OpenDefaultDb("./batch_import_db")
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
        "id":   0,
        "name": "",
        "age":  0,
        "email": "",
    }
    err = table.SetFields(fields)
    if err != nil {
        panic(err)
    }

    // 批量导入数据
    fmt.Println("=== 批量数据导入 ===")
    
    batchSize := 1000 // 每批处理1000条记录
    totalRecords := 10000 // 总记录数
    
    for i := 0; i < totalRecords; i += batchSize {
        // 获取批处理对象
        batch := storage.KVDb.GetBatch()
        if batch == nil {
            panic("无法获取批量操作对象")
        }
        
        // 计算当前批次的记录范围
        end := i + batchSize
        if end > totalRecords {
            end = totalRecords
        }
        
        fmt.Printf("处理批次: %d-%d\n", i+1, end)
        
        // 添加记录到批处理
        for j := i; j < end; j++ {
            user := map[string]any{
                "name": fmt.Sprintf("用户%d", j+1),
                "age":  20 + (j%30),
                "email": fmt.Sprintf("user%d@example.com", j+1),
            }
            
            _, err := table.Insert(&user, batch)
            if err != nil {
                panic(fmt.Sprintf("添加记录失败 (第%d条): %v", j+1, err))
            }
        }
        
        // 提交批处理
        start := time.Now()
        err = storage.KVDb.WriteBatch(batch)
        if err != nil {
            panic(fmt.Sprintf("批量提交失败: %v", err))
        }
        elapsed := time.Since(start)
        
        fmt.Printf("批次提交成功，耗时: %v\n", elapsed)
    }
    
    fmt.Printf("\n批量导入完成，共导入 %d 条记录\n", totalRecords)
}
```

## 6. 事务实现原理

### 6.1 批处理机制

sfsDb 的事务实现基于底层 KV 存储的批处理机制：

1. **操作收集**：将多个操作收集到一个批处理对象中
2. **原子执行**：通过底层 KV 存储的原子操作，将批处理中的所有操作一次性执行
3. **故障处理**：如果任何操作失败，整个批处理都会失败，确保数据一致性

### 6.2 并发控制

sfsDb 的事务处理考虑了并发场景：

- **乐观并发控制**：假设冲突很少发生，通过检测冲突来确保一致性
- **批量操作的原子性**：批处理中的所有操作要么全部成功，要么全部失败
- **避免长事务**：鼓励使用短事务，减少并发冲突的可能性

## 7. 性能测试

### 7.1 测试场景

| 操作类型 | 记录数 | 自动事务耗时 | 手动事务耗时 | 性能提升 |
|---------|-------|------------|------------|---------|
| 插入 | 1000 | 2.5s | 0.3s | 88% |
| 插入 | 5000 | 12.3s | 1.2s | 90% |
| 插入 | 10000 | 25.1s | 2.1s | 91% |
| 混合操作 | 1000 | 3.2s | 0.4s | 87.5% |

### 7.2 测试结论

1. **显著性能提升**：手动事务比自动事务快 85-90%
2. **批量优势明显**：处理的记录数越多，手动事务的优势越明显
3. **内存使用合理**：批量操作的内存使用在可控范围内

## 9. 事务管理器

### 9.1 概述

sfsDb 提供了 `TransactionManager` 用于管理跨多个表的事务，确保跨表操作的原子性。

### 9.2 创建事务管理器

```go
// 创建事务管理器，使用共享批处理对象
batch := storage.KVDb.GetBatch()
tm := engine.NewTransactionManager(batch)
```

### 9.3 事务管理器方法

#### AddTable

将表添加到事务管理器并返回对应的事务：

```go
// 添加表到事务管理器
tx, err := tm.AddTable(table)
if err != nil {
    panic(err)
}
```

#### Commit

提交事务管理器管理的所有事务：

```go
// 提交所有事务
err = tm.Commit()
if err != nil {
    panic(err)
}
```

#### Rollback

回滚事务管理器管理的所有事务：

```go
// 回滚所有事务
err = tm.Rollback()
if err != nil {
    panic(err)
}
```

### 9.4 WithTransaction 辅助函数

`WithTransaction` 函数提供了执行多表事务的便捷方式：

```go
// 执行多表事务
batch := storage.KVDb.GetBatch()
tables := []*engine.Table{table1, table2}

err := engine.WithTransaction(batch, tables, func(transactions map[*engine.Table]engine.Transaction) error {
    // 获取每个表的事务
    tx1 := transactions[table1]
    tx2 := transactions[table2]
    
    // 在 table1 上执行操作
    user := map[string]any{"name": "张三", "age": 30}
    _, err := tx1.Insert(&user)
    if err != nil {
        return err
    }
    
    // 在 table2 上执行操作
    order := map[string]any{"user_id": 1, "product": "商品A"}
    _, err = tx2.Insert(&order)
    if err != nil {
        return err
    }
    
    return nil
})

if err != nil {
    panic(err)
}
```

### 9.5 多表事务示例

```go
// 多表事务示例
func multiTableTransactionExample() error {
    // 获取批处理对象
    batch := storage.KVDb.GetBatch()
    if batch == nil {
        return fmt.Errorf("无法获取批处理对象")
    }
    
    // 创建表（假设它们已经存在）
    // table1 := ... // 用户表
    // table2 := ... // 订单表
    
    // 执行多表事务
    return engine.WithTransaction(batch, []*engine.Table{table1, table2}, func(txs map[*engine.Table]engine.Transaction) error {
        // 插入用户
        user := map[string]any{
            "name": "Alice",
            "email": "alice@example.com",
        }
        userID, err := txs[table1].Insert(&user)
        if err != nil {
            return fmt.Errorf("插入用户失败: %w", err)
        }
        
        // 为用户插入订单
        order := map[string]any{
            "user_id": userID,
            "product": "Premium Plan",
            "amount":  99.99,
        }
        _, err = txs[table2].Insert(&order)
        if err != nil {
            return fmt.Errorf("插入订单失败: %w", err)
        }
        
        fmt.Println("多表事务执行成功")
        return nil
    })
}
```

## 10. 总结

正确使用事务操作可以显著提高 sfsDb 的性能和可靠性：

1. **自动事务**：适合单个操作，简单易用
2. **手动事务**：适合多个操作的原子执行，性能优异
3. **事务管理器**：适合多表事务，确保跨表操作的原子性
4. **批量大小**：根据系统资源和数据量选择合适的批量大小
5. **错误处理**：妥善处理事务中的错误，确保数据一致性
6. **并发控制**：合理控制事务粒度，避免长事务

通过合理使用事务机制，可以在保证数据一致性的同时，充分发挥 sfsDb 的性能潜力，特别是在处理大量数据或执行跨表操作时。