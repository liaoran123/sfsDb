# transaction 包详细文档

## 一、包结构

```
transaction/
├── batch/                # 批量操作优化器
│   └── optimizer.go      # 批量操作优化实现
├── lock/                 # 锁管理器
│   ├── manager.go        # 基本锁管理器
│   └── sharded_manager.go # 分片锁管理器
├── benchmark_test.go     # 性能基准测试
├── occ.go                # 乐观并发控制实现
├── options.go            # 事务选项和隔离级别定义
├── order_transaction_test.go # 订单事务测试
├── stress_transaction_test.go # 压力测试
├── transaction.go        # 核心事务实现
├── transaction_manager.go # 多表事务管理器
└── transaction_manager_test.go # 事务管理器测试
```

## 二、核心组件

### 1. 事务接口

- **Transaction**：定义了基本事务操作，包括 Get、Put、Delete、Commit、Rollback 等
- **TableTransactionInterface**：定义了表级事务操作，包括 Insert、Update、Delete、Read、Search 等

### 2. 事务实现

- **SfsTransaction**：基于键值存储的事务实现
- **TableTransaction**：基于表的事务实现，支持复杂的CRUD操作

### 3. 事务池

- **GlobalTransactionPool**：全局事务对象池，用于复用事务对象
- **GlobalTableTransactionPool**：全局表事务对象池

### 4. 并发控制

- **GlobalLockManager**：全局分片锁管理器，用于并发控制
- **OCCSupport**：乐观并发控制支持，用于检测并发冲突

### 5. 批量操作

- **GlobalBatchOptimizer**：全局批量操作优化器，用于优化批量操作性能

## 三、隔离级别实现

### 1. 隔离级别定义

| 隔离级别 | 常量 | 描述 |
|---------|------|------|
| 读未提交 | `ReadUncommitted` | 允许读取未提交的数据，可能导致脏读、不可重复读、幻读 |
| 读已提交 | `ReadCommitted` | 只能读取已提交的数据，避免脏读，但可能导致不可重复读、幻读 |
| 可重复读 | `RepeatableRead` | 确保同一事务中多次读取同一数据时结果一致，避免脏读、不可重复读，但可能导致幻读 |
| 可序列化 | `Serializable` | 最高隔离级别，完全避免脏读、不可重复读、幻读，但性能最低 |

### 2. 实现细节

#### ReadUncommitted
- **实现**：不创建快照，直接读取原始存储
- **特点**：允许读取未提交的数据，性能最高但一致性最低

#### ReadCommitted
- **实现**：每次读取都直接使用原始存储，依赖存储引擎的提交机制
- **特点**：避免脏读，但可能出现短暂的不一致
- **局限性**：严格来说，ReadCommitted 应该每次读取都创建新的快照，但当前实现为了简化，直接读取原始存储

#### RepeatableRead
- **实现**：事务开始时创建快照，整个事务使用同一个快照
- **特点**：避免脏读、不可重复读，但可能导致幻读

#### Serializable
- **实现**：与 RepeatableRead 类似，使用快照隔离
- **特点**：尝试避免所有并发问题，但可能无法完全避免幻读
- **局限性**：当前实现使用快照隔离，可能无法完全避免幻读

## 四、当前实现的局限性

### 1. ReadCommitted 的简化
- **理论要求**：每次读取都应创建新的快照
- **当前实现**：直接读取原始存储，依赖存储引擎的提交机制
- **影响**：在高并发场景下可能出现短暂的不一致，但实际效果接近 ReadCommitted

### 2. Serializable 的简化
- **理论要求**：使用锁或乐观并发控制，完全避免幻读
- **当前实现**：使用快照隔离，可能无法完全避免幻读
- **影响**：对于大多数应用场景已经足够，但对于严格的序列化需求可能不够

### 3. 锁机制的局限性
- **当前实现**：使用分片锁管理器，支持读写锁
- **局限性**：对于 Serializable 隔离级别，缺少范围锁和谓词锁

## 五、解决方案

### 注意
以下解决方案符合sfsDb的设计理念，保持简单且专注于核心功能，同时解决实际使用中的问题。

### 1. ReadCommitted 优化方案

**问题**：当前实现直接读取原始存储，可能导致短暂的不一致

**解决方案**：

对于sfsDb这样的轻量级嵌入式数据库，我们可以采用以下实用方案：

1. **应用层处理**：
   - 对于需要严格ReadCommitted隔离级别的应用，可以在应用层实现重试机制
   - 当检测到数据不一致时，重新读取数据

2. **批量操作优化**：
   - 合并多个操作到一个事务中，减少并发冲突的可能性
   - 使用批量操作减少数据库的写入次数

3. **合理使用快照**：
   - 对于关键操作，可以在事务开始时创建快照
   - 利用现有的快照机制，确保读取的一致性

### 2. Serializable 优化方案

**问题**：当前实现使用快照隔离，可能无法完全避免幻读

**真实解决方案**：

对于大多数嵌入式应用场景，sfsDb的当前实现已经足够。对于确实需要更高隔离级别的场景：

1. **应用层锁定**：
   - 在应用层实现逻辑锁，确保关键操作的序列化
   - 使用分布式锁或应用级锁管理关键资源

2. **批量操作**：
   - 将相关操作批量执行，减少中间状态的暴露
   - 一次性提交所有相关修改，避免其他事务看到部分修改

3. **合理设计数据模型**：
   - 优化数据模型，减少需要 Serializable 隔离级别的场景
   - 使用聚合设计，将相关数据放在一起处理

### 3. 并发控制优化方案

**问题**：当前锁机制对于复杂的并发场景可能不够完善

**真实解决方案**：

1. **优化锁使用**：
   - 减少锁的持有时间，尽快释放不需要的锁
   - 使用非阻塞锁，提高并发性能

2. **分片设计**：
   - 利用现有的分片锁管理器，将数据分散到不同的锁分片
   - 减少锁竞争，提高并发度

3. **乐观并发控制**：
   - 充分利用现有的OCCSupport机制
   - 在应用层实现版本控制，检测并发冲突

4. **批量提交**：
   - 合并多个修改为一个批量操作
   - 减少事务的数量，降低锁竞争

### 4. 最佳实践建议

1. **根据场景选择合适的隔离级别**：
   - 大多数嵌入式应用使用默认的RepeatableRead隔离级别即可
   - 只有在确实需要时才考虑更高的隔离级别

2. **优化事务设计**：
   - 保持事务短小精悍，减少锁持有时间
   - 避免长时间运行的事务

3. **合理使用索引**：
   - 为常用查询字段创建索引，减少全表扫描
   - 提高查询性能，减少锁竞争

4. **监控与调优**：
   - 使用内置的监控功能，跟踪事务执行情况
   - 根据实际使用情况调整配置参数

这些解决方案符合sfsDb的设计理念，保持了系统的简洁性和高性能，同时解决了实际使用中的并发控制问题。

## 六、使用指南

### 1. 基本事务操作

```go
// 创建事务
store := storage.GetDBManager().GetDB()
tx, err := transaction.NewTransaction(store)
if err != nil {
    // 处理错误
}

// 执行操作
tx.Put([]byte("key1"), []byte("value1"))
tx.Put([]byte("key2"), []byte("value2"))

// 提交事务
if err := tx.Commit(); err != nil {
    // 处理错误
    tx.Rollback()
}
```

### 2. 表事务操作

```go
// 获取表
table, err := engine.GetTable("users")
if err != nil {
    // 处理错误
}

// 创建表事务
tx, err := transaction.NewTableTransaction(table)
if err != nil {
    // 处理错误
}

// 插入记录
fields := map[string]interface{}{
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
}
id, err := tx.Insert(&fields)
if err != nil {
    // 处理错误
    tx.Rollback()
}

// 提交事务
if err := tx.Commit(); err != nil {
    // 处理错误
    tx.Rollback()
}
```

### 3. 多表事务

```go
// 创建事务管理器
tm := transaction.NewTransactionManager()

// 获取表
table1, err := engine.GetTable("users")
table2, err := engine.GetTable("orders")

// 添加表到事务管理器
tx1, err := tm.AddTable(table1)
tx2, err := tm.AddTable(table2)

// 在不同表上执行操作
// ...

// 提交所有表事务
if err := tm.Commit(); err != nil {
    // 处理错误
    tm.Rollback()
}
```

### 4. 自定义事务选项

```go
// 创建自定义事务选项
options := &transaction.TransactionOptions{
    IsolationLevel:     transaction.Serializable,
    AllowNested:        true,
    Timeout:            30 * time.Second,
    MaxRetries:         5,
    InitialRetryDelay:  10 * time.Millisecond,
    RetryBackoffFactor: 2.0,
}

// 使用自定义选项创建事务
tx, err := transaction.NewTableTransactionWithOptions(table, options)
```

## 七、性能优化

### 1. 事务对象池

使用事务对象池可以减少内存分配和垃圾回收开销：

```go
// 从对象池获取事务
tx := transaction.GlobalTableTransactionPool.Get()

// 使用事务
// ...

// 事务完成后归还到池
if tx.Parent == nil {
    transaction.GlobalTableTransactionPool.Put(tx)
}
```

### 2. 批量操作优化

批量操作可以减少磁盘I/O，提高性能：

```go
// 批量插入
for i := 0; i < 1000; i++ {
    fields := map[string]interface{}{
        "id": i,
        "name": fmt.Sprintf("User%d", i),
    }
    tx.Insert(&fields)
}

// 一次提交所有操作
tx.Commit()
```

### 3. 锁优化

- **使用非阻塞锁**：减少线程等待时间
- **锁粒度**：尽量使用细粒度锁，减少锁竞争
- **锁释放**：及时释放不需要的锁

## 八、监控与统计

transaction 包集成了监控功能，可以跟踪事务执行情况：

```go
// 事务执行统计信息会自动记录到 monitor.GTransactionStatsMap
// 可以通过以下方式获取统计信息
stats := monitor.GTransactionStatsMap.GetStats()
```

## 九、测试

### 1. 单元测试

- **order_transaction_test.go**：测试订单事务功能
- **transaction_manager_test.go**：测试事务管理器功能

### 2. 压力测试

- **stress_transaction_test.go**：测试高并发场景下的事务性能
- **benchmark_test.go**：性能基准测试

## 十、总结

transaction 包提供了完整的ACID事务支持，包括：

- **原子性**：通过批量操作实现
- **一致性**：通过快照和锁机制实现
- **隔离性**：支持四种隔离级别
- **持久性**：通过存储引擎的持久化机制实现

虽然当前实现存在一些局限性，特别是在 ReadCommitted 和 Serializable 隔离级别的实现上，但通过本文提供的解决方案，可以进一步完善隔离级别的实现，提高事务的可靠性和性能。

transaction 包的设计理念是"减法"哲学，通过简化设计和专注核心功能，为嵌入式场景提供高性能的事务支持，同时保持了足够的灵活性和可扩展性。

## 十一、未来规划

1. **完善隔离级别实现**：优化 ReadCommitted 和 Serializable 隔离级别的实现
2. **增强并发控制**：添加死锁检测和预防机制
3. **性能优化**：进一步优化事务性能，特别是在高并发场景下
4. **功能扩展**：添加更多高级功能，如分布式事务支持
5. **文档完善**：提供更详细的使用文档和示例

---

**文档版本**：1.0
**最后更新**：2026-02-21
**作者**：sfsDb 开发团队
