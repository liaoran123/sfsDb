# 第 5 章：事务处理

事务保证了数据操作的原子性、一致性、隔离性和持久性。本章将深入学习 sfsDb 的无锁事务系统。

## 5.1 事务基础

### 5.1.1 ACID 特性

ACID 是数据库事务的四个核心特性，sfsDb 通过无锁事务系统完整支持：

| 特性 | 定义 | sfsDb 实现方式 |
|------|------|----------------|
| **原子性 (Atomicity)** | 事务中的所有操作要么全部成功，要么全部失败 | 基于 LevelDB 的批量操作机制 |
| **一致性 (Consistency)** | 事务执行前后，数据库从一个一致状态转换到另一个一致状态 | 事务内操作验证和约束检查 |
| **隔离性 (Isolation)** | 并发事务之间互不干扰，每个事务感觉自己是唯一执行的 | 乐观并发控制 + 版本管理 |
| **持久性 (Durability)** | 事务提交后，数据永久保存，即使系统故障也不会丢失 | LevelDB 的 Write-Ahead Log (WAL) 机制 |

### 5.1.2 乐观并发控制（OCC）

sfsDb 采用**乐观并发控制 (Optimistic Concurrency Control, OCC)** 而非传统的悲观锁机制：

**核心思想：**
- 假设并发冲突很少发生
- 事务执行时不加锁
- 提交时检查是否有冲突
- 如有冲突则重试或回滚

**实现原理：**
```go
// LockFreeVersionManager 版本管理器
type LockFreeVersionManager struct {
    versions map[string]uint64 // 键 -> 版本号
    mu       sync.RWMutex      // 读写锁
    initOnce sync.Once         // 仅用于初始化
}
```

**版本管理流程：**
1. 事务开始时记录读取的版本号
2. 事务执行时在本地缓存修改
3. 提交时检查版本号是否变化
4. 如无变化则递增版本号并提交
5. 如有变化则触发冲突处理

### 5.1.3 无锁 vs 有锁

| 对比项 | 无锁事务 (sfsDb) | 传统有锁事务 |
|--------|-----------------|-------------|
| **并发性能** | 高，无锁等待 | 低，锁竞争严重 |
| **死锁风险** | 无 | 有 |
| **实现复杂度** | 较高 | 较低 |
| **适用场景** | 读多写少，冲突概率低 | 写多读少，冲突概率高 |
| **资源消耗** | 低，无锁开销 | 高，锁管理开销 |

**性能对比（并发事务 TPS）：**
- sfsDb 无锁事务：~16,000 TPS
- MySQL InnoDB：~2,000-5,000 TPS
- PostgreSQL：~3,000-6,000 TPS

## 5.2 事务基本操作

### 5.2.1 开始事务

**事务选项配置：**
```go
// TransactionOptions 事务选项结构体
type TransactionOptions struct {
    IsolationLevel     string        // 隔离级别
    AllowNested        bool          // 是否启用嵌套事务
    Timeout            time.Duration // 事务超时时间
    MaxRetries         int           // 最大重试次数
    InitialRetryDelay  time.Duration // 初始重试延迟
    RetryBackoffFactor float64       // 重试退避因子
}
```

**隔离级别：**
- `ReadUncommitted`：读未提交（性能最高，隔离性最低）
- `ReadCommitted`：读已提交（默认）
- `RepeatableRead`：可重复读
- `Serializable`：可序列化（隔离性最高，性能最低）

**创建基本事务：**
```go
import (
    "github.com/liaoran123/sfsDb/transactionLockFree"
)

// 创建事务
tx, err := transactionLockFree.NewTransaction(store)
if err != nil {
    log.Fatalf("创建事务失败: %v", err)
}

// 使用自定义选项
options := transactionLockFree.DefaultTransactionOptions()
options.IsolationLevel = transactionLockFree.RepeatableRead
options.MaxRetries = 5

tx, err := transactionLockFree.NewTransactionWithOptions(store, options)
```

**表事务操作：**
```go
// 为指定表创建事务
tableTx, err := transactionLockFree.NewTableTransaction(table)
if err != nil {
    log.Fatalf("创建表事务失败: %v", err)
}
```

### 5.2.2 提交事务

**基本键值事务提交：**
```go
// 执行操作
tx.Put([]byte("user:1"), []byte(`{"name":"张三","age":25}`))
tx.Put([]byte("user:2"), []byte(`{"name":"李四","age":30}`))

// 提交事务
err = tx.Commit()
if err != nil {
    log.Fatalf("提交事务失败: %v", err)
}

fmt.Println("事务提交成功！")
```

**表事务提交：**
```go
// 插入记录
fields := map[string]interface{}{
    "id":   1,
    "name": "张三",
    "age":  25,
}
id, err := tableTx.Insert(&fields)

// 提交事务
err = tableTx.Commit()
if err != nil {
    log.Fatalf("提交表事务失败: %v", err)
}
```

### 5.2.3 回滚事务

**基本回滚：**
```go
// 执行操作
tx.Put([]byte("key1"), []byte("value1"))

// 发现问题，回滚事务
err = tx.Rollback()
if err != nil {
    log.Fatalf("回滚事务失败: %v", err)
}

fmt.Println("事务已回滚，修改未生效")
```

**带错误处理的完整事务：**
```go
func transferMoney(fromID, toID int, amount float64) error {
    tx, err := transactionLockFree.NewTransaction(store)
    if err != nil {
        return err
    }
    
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            panic(r)
        }
    }()
    
    // 检查余额
    fromBalance, err := getBalance(tx, fromID)
    if err != nil {
        tx.Rollback()
        return err
    }
    
    if fromBalance < amount {
        tx.Rollback()
        return fmt.Errorf("余额不足")
    }
    
    // 执行转账
    err = updateBalance(tx, fromID, fromBalance-amount)
    if err != nil {
        tx.Rollback()
        return err
    }
    
    toBalance, err := getBalance(tx, toID)
    if err != nil {
        tx.Rollback()
        return err
    }
    
    err = updateBalance(tx, toID, toBalance+amount)
    if err != nil {
        tx.Rollback()
        return err
    }
    
    // 提交事务
    return tx.Commit()
}
```

## 5.3 批量操作

### 5.3.1 批量插入

**使用事务批量插入：**
```go
tx, err := transactionLockFree.NewTableTransaction(userTable)
if err != nil {
    log.Fatal(err)
}
defer tx.Rollback()

users := []map[string]interface{}{
    {"id": 1, "name": "张三", "age": 25},
    {"id": 2, "name": "李四", "age": 30},
    {"id": 3, "name": "王五", "age": 28},
    {"id": 4, "name": "赵六", "age": 35},
    {"id": 5, "name": "钱七", "age": 22},
}

for _, user := range users {
    _, err := tx.Insert(&user)
    if err != nil {
        log.Fatalf("插入用户失败: %v", err)
    }
}

err = tx.Commit()
if err != nil {
    log.Fatalf("批量插入失败: %v", err)
}

fmt.Printf("成功插入 %d 条记录\n", len(users))
```

### 5.3.2 批量更新

**使用 BatchOperations 方法：**
```go
tx, err := transactionLockFree.NewTableTransaction(accountTable)
if err != nil {
    log.Fatal(err)
}
defer tx.Rollback()

// 准备批量操作
operations := []func() error{
    func() error {
        fields := map[string]interface{}{"id": 1, "balance": 1000.0}
        return tx.Update(&fields)
    },
    func() error {
        fields := map[string]interface{}{"id": 2, "balance": 2000.0}
        return tx.Update(&fields)
    },
    func() error {
        fields := map[string]interface{}{"id": 3, "balance": 3000.0}
        return tx.Update(&fields)
    },
}

// 执行批量操作
err = tx.BatchOperations(operations)
if err != nil {
    log.Fatalf("批量更新失败: %v", err)
}

err = tx.Commit()
if err != nil {
    log.Fatalf("提交失败: %v", err)
}

fmt.Println("批量更新成功！")
```

### 5.3.3 批量删除

**批量删除记录：**
```go
tx, err := transactionLockFree.NewTableTransaction(orderTable)
if err != nil {
    log.Fatal(err)
}
defer tx.Rollback()

// 删除过期订单
expiredOrderIDs := []int{1001, 1002, 1003, 1004, 1005}

for _, orderID := range expiredOrderIDs {
    fields := map[string]interface{}{"id": orderID}
    err := tx.Delete(&fields)
    if err != nil {
        log.Printf("删除订单 %d 失败: %v", orderID, err)
    }
}

err = tx.Commit()
if err != nil {
    log.Fatalf("批量删除失败: %v", err)
}

fmt.Printf("成功删除 %d 条过期订单\n", len(expiredOrderIDs))
```

## 5.4 嵌套事务

### 5.4.1 保存点（Savepoint）

**创建和使用保存点：**
```go
// 创建事务管理器
batch := store.GetBatch()
manager := transactionLockFree.NewTransactionManager(batch)

// 执行初始操作
// ...

// 创建保存点
savepoint, err := manager.CreateSavepoint("before_risky_operation")
if err != nil {
    log.Fatalf("创建保存点失败: %v", err)
}

// 执行风险操作
err = performRiskyOperation(manager)
if err != nil {
    // 操作失败，回滚到保存点
    log.Printf("风险操作失败: %v，回滚到保存点", err)
    err = manager.RollbackToSavepoint(savepoint)
    if err != nil {
        log.Fatalf("回滚到保存点失败: %v", err)
    }
} else {
    // 操作成功，提交事务
    err = manager.Commit()
    if err != nil {
        log.Fatalf("提交失败: %v", err)
    }
}
```

### 5.4.2 嵌套事务使用场景

**嵌套事务示例：**
```go
// 创建父事务
parentTx, err := transactionLockFree.NewTransaction(store)
if err != nil {
    log.Fatal(err)
}
defer parentTx.Rollback()

// 父事务操作
parentTx.Put([]byte("parent:key"), []byte("parent:value"))

// 创建嵌套事务
nestedTx, err := parentTx.BeginNested()
if err != nil {
    log.Fatal(err)
}

// 嵌套事务操作
nestedTx.Put([]byte("nested:key1"), []byte("nested:value1"))
nestedTx.Put([]byte("nested:key2"), []byte("nested:value2"))

// 提交嵌套事务
err = nestedTx.Commit()
if err != nil {
    log.Fatalf("嵌套事务提交失败: %v", err)
}

// 嵌套事务的操作会累积到父事务中
// 提交父事务
err = parentTx.Commit()
if err != nil {
    log.Fatalf("父事务提交失败: %v", err)
}

fmt.Println("嵌套事务提交成功！")
```

**多层嵌套事务：**
```go
func complexWorkflow() error {
    // 第一层事务
    tx1, err := transactionLockFree.NewTransaction(store)
    if err != nil {
        return err
    }
    defer tx1.Rollback()
    
    tx1.Put([]byte("level1:key"), []byte("level1:value"))
    
    // 第二层事务
    tx2, err := tx1.BeginNested()
    if err != nil {
        return err
    }
    
    tx2.Put([]byte("level2:key"), []byte("level2:value"))
    
    // 第三层事务
    tx3, err := tx2.BeginNested()
    if err != nil {
        return err
    }
    
    tx3.Put([]byte("level3:key"), []byte("level3:value"))
    
    // 提交第三层
    err = tx3.Commit()
    if err != nil {
        return err
    }
    
    // 提交第二层
    err = tx2.Commit()
    if err != nil {
        return err
    }
    
    // 提交第一层
    return tx1.Commit()
}
```

## 5.5 性能优化

### 5.5.1 事务大小优化

**最佳实践：**
1. **保持事务简短**：减少事务持有时间，降低冲突概率
2. **避免长事务**：长事务会增加冲突概率，影响并发性能
3. **合理批量**：将相关操作合并，但不要过度批量
4. **避免热点**：避免多个事务同时更新同一行数据

**反模式示例：**
```go
// ❌ 不好的做法：一个事务处理所有数据
tx, err := transactionLockFree.NewTransaction(store)
if err != nil {
    log.Fatal(err)
}
defer tx.Rollback()

// 处理 10000 条记录，事务持续时间很长
for i := 0; i < 10000; i++ {
    tx.Put([]byte(fmt.Sprintf("key:%d", i)), []byte("value"))
}

err = tx.Commit() // 冲突概率很高
```

**优化后：**
```go
// ✅ 好的做法：分批处理
batchSize := 100
for batch := 0; batch < 100; batch++ {
    tx, err := transactionLockFree.NewTransaction(store)
    if err != nil {
        log.Fatal(err)
    }
    
    for i := 0; i < batchSize; i++ {
        key := fmt.Sprintf("key:%d", batch*batchSize+i)
        tx.Put([]byte(key), []byte("value"))
    }
    
    err = tx.Commit()
    if err != nil {
        log.Printf("批次 %d 提交失败，重试: %v", batch, err)
        batch-- // 重试该批次
        continue
    }
}
```

### 5.5.2 并发事务处理

**配置重试策略：**
```go
options := transactionLockFree.DefaultTransactionOptions()
options.MaxRetries = 5
options.InitialRetryDelay = 10 * time.Millisecond
options.RetryBackoffFactor = 2.0 // 指数退避

// 使用乐观更新
fields := map[string]interface{}{"id": 1, "balance": 100}
err := tx.OptimisticUpdate(&fields, 3) // 最多重试3次
if err != nil {
    log.Fatalf("乐观更新失败: %v", err)
}
```

**多表事务：**
```go
// 创建共享批量操作
batch := store.GetBatch()

// 使用便捷函数执行多表事务
err := transactionLockFree.WithTransaction(
    batch,
    []*engine.Table{orderTable, inventoryTable},
    func(txs map[*engine.Table]transactionLockFree.TableTransactionInterface) error {
        // 在 orderTable 上执行操作
        orderTx := txs[orderTable]
        orderFields := map[string]interface{}{
            "id":         1001,
            "product_id": 1,
            "quantity":   5,
            "status":     "pending",
        }
        _, err := orderTx.Insert(&orderFields)
        if err != nil {
            return err
        }
        
        // 在 inventoryTable 上执行操作
        inventoryTx := txs[inventoryTable]
        inventoryFields := map[string]interface{}{
            "product_id": 1,
            "stock":      95,
        }
        err = inventoryTx.Update(&inventoryFields)
        if err != nil {
            return err
        }
        
        return nil
    },
)

if err != nil {
    log.Fatalf("多表事务失败: %v", err)
}
```

### 5.5.3 性能基准测试

**并发性能测试结果：**

| 操作类型 | 每秒操作数 | 平均延迟 |
|---------|-----------|----------|
| 转账操作 | ~16,000 | < 1ms |
| 订单创建 | ~3,700 | ~0.3ms |
| 批量更新 | ~5,000 | ~0.2ms |

**与其他数据库比较：**

| 数据库 | 并发事务性能 (TPS) | 事务延迟 (ms) | ACID 完整性 | 隔离级别支持 |
|--------|-------------------|--------------|-------------|-------------|
| sfsDb 无锁事务 | ~16,000 | < 1 | 完整 | 4 级 |
| MySQL InnoDB | ~2,000-5,000 | 2-5 | 完整 | 4 级 |
| PostgreSQL | ~3,000-6,000 | 1-4 | 完整 | 4 级 |
| MongoDB 4.0+ | ~8,000-12,000 | 1-3 | 部分 | 2 级 |
| Redis | ~100,000+ | < 0.1 | 部分 | 有限 |

**性能优化策略总结：**
1. ✅ 使用批量操作减少磁盘 I/O
2. ✅ 合理设置隔离级别（不要过度隔离）
3. ✅ 使用对象池减少内存分配
4. ✅ 优化重试策略
5. ✅ 避免长事务和热点数据

## 5.6 实战示例

### 5.6.1 银行转账系统

**完整的转账事务示例：**
```go
package main

import (
    "fmt"
    "log"
    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/storage"
    "github.com/liaoran123/sfsDb/transactionLockFree"
)

type Account struct {
    ID      int
    Name    string
    Balance float64
}

func transfer(dbManager *storage.DBManager, fromID, toID int, amount float64) error {
    // 获取账户表
    accountTable := dbManager.GetTable("accounts")
    if accountTable == nil {
        return fmt.Errorf("账户表不存在")
    }
    
    // 创建共享批量操作
    batch := dbManager.GetDB("bank_db").GetBatch()
    
    // 执行转账事务
    return transactionLockFree.WithTransaction(
        batch,
        []*engine.Table{accountTable},
        func(txs map[*engine.Table]transactionLockFree.TableTransactionInterface) error {
            tx := txs[accountTable]
            
            // 1. 查询转出账户
            fromFields := map[string]any{"id": fromID}
            fromData, err := tx.Read(&fromFields)
            if err != nil {
                return fmt.Errorf("查询转出账户失败: %v", err)
            }
            
            // 解析账户数据
            var fromAccount Account
            // 解析 fromData 到 fromAccount...
            
            // 2. 检查余额
            if fromAccount.Balance < amount {
                return fmt.Errorf("账户 %d 余额不足: %.2f < %.2f", 
                    fromID, fromAccount.Balance, amount)
            }
            
            // 3. 查询转入账户
            toFields := map[string]any{"id": toID}
            toData, err := tx.Read(&toFields)
            if err != nil {
                return fmt.Errorf("查询转入账户失败: %v", err)
            }
            
            var toAccount Account
            // 解析 toData 到 toAccount...
            
            // 4. 扣减转出账户余额
            fromUpdate := map[string]interface{}{
                "id":      fromID,
                "balance": fromAccount.Balance - amount,
            }
            err = tx.Update(&fromUpdate)
            if err != nil {
                return fmt.Errorf("扣减转出账户余额失败: %v", err)
            }
            
            // 5. 增加转入账户余额
            toUpdate := map[string]interface{}{
                "id":      toID,
                "balance": toAccount.Balance + amount,
            }
            err = tx.Update(&toUpdate)
            if err != nil {
                return fmt.Errorf("增加转入账户余额失败: %v", err)
            }
            
            fmt.Printf("转账成功: %d -> %d, 金额: %.2f\n", 
                fromID, toID, amount)
            
            return nil
        },
    )
}

func main() {
    // 初始化数据库
    dbManager := storage.GetDBManager()
    _, err := dbManager.OpenDB("./bank_db")
    if err != nil {
        log.Fatalf("打开数据库失败: %v", err)
    }
    defer dbManager.CloseAllDB()
    
    // 执行转账
    err = transfer(dbManager, 1, 2, 100.0)
    if err != nil {
        log.Fatalf("转账失败: %v", err)
    }
}
```

### 5.6.2 订单处理流程

**电商订单处理事务：**
```go
package main

import (
    "fmt"
    "log"
    "time"
    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/storage"
    "github.com/liaoran123/sfsDb/transactionLockFree"
)

func createOrder(
    dbManager *storage.DBManager,
    orderID int,
    productID int,
    userID int,
    quantity int,
) error {
    // 获取表
    orderTable := dbManager.GetTable("orders")
    inventoryTable := dbManager.GetTable("inventory")
    
    if orderTable == nil || inventoryTable == nil {
        return fmt.Errorf("所需表不存在")
    }
    
    // 创建共享批量操作
    batch := dbManager.GetDB("ecommerce_db").GetBatch()
    
    // 执行订单事务
    return transactionLockFree.WithTransaction(
        batch,
        []*engine.Table{orderTable, inventoryTable},
        func(txs map[*engine.Table]transactionLockFree.TableTransactionInterface) error {
            orderTx := txs[orderTable]
            inventoryTx := txs[inventoryTable]
            
            // 1. 检查库存
            stockFields := map[string]any{"product_id": productID}
            stockData, err := inventoryTx.Read(&stockFields)
            if err != nil {
                return fmt.Errorf("查询库存失败: %v", err)
            }
            
            var currentStock int
            // 解析库存数据...
            
            if currentStock < quantity {
                return fmt.Errorf("库存不足: 当前 %d, 需要 %d", 
                    currentStock, quantity)
            }
            
            // 2. 创建订单
            orderFields := map[string]interface{}{
                "id":         orderID,
                "user_id":    userID,
                "product_id": productID,
                "quantity":   quantity,
                "status":     "pending",
                "created_at": time.Now().Unix(),
            }
            _, err = orderTx.Insert(&orderFields)
            if err != nil {
                return fmt.Errorf("创建订单失败: %v", err)
            }
            
            // 3. 扣减库存
            inventoryUpdate := map[string]interface{}{
                "product_id": productID,
                "stock":      currentStock - quantity,
            }
            err = inventoryTx.Update(&inventoryUpdate)
            if err != nil {
                return fmt.Errorf("扣减库存失败: %v", err)
            }
            
            fmt.Printf("订单创建成功: 订单ID=%d, 商品ID=%d, 数量=%d\n", 
                orderID, productID, quantity)
            
            return nil
        },
    )
}

func main() {
    dbManager := storage.GetDBManager()
    _, err := dbManager.OpenDB("./ecommerce_db")
    if err != nil {
        log.Fatal(err)
    }
    defer dbManager.CloseAllDB()
    
    err = createOrder(dbManager, 1001, 1, 123, 2)
    if err != nil {
        log.Fatalf("创建订单失败: %v", err)
    }
}
```

## 5.7 架构师视角

### 5.7.1 无锁事务实现原理

**核心数据结构：**
```go
// TransactionManager 事务管理器
type TransactionManager struct {
    batch      storage.Batch
    tableTxMap map[*engine.Table]*TableTransaction
    savepoints map[string]*Savepoint
    parent     *TransactionManager
    children   []*TransactionManager
    committed  bool
    rolledBack bool
    options    *TransactionOptions
    txID       uint64
    startTime  time.Time
}
```

**事务生命周期：**
1. **开始**：创建 TransactionManager，记录开始时间和事务 ID
2. **执行**：操作缓存在本地，不立即写入存储
3. **验证**：提交前检查版本号，检测冲突
4. **提交**：如无冲突，批量写入存储，递增版本号
5. **回滚**：如有冲突或显式回滚，丢弃本地修改

### 5.7.2 冲突检测机制

**版本号检查流程：**
```
事务开始
    ↓
读取数据，记录版本号 V1
    ↓
本地修改数据
    ↓
提交时检查
    ↓
当前版本号 == V1 ?
    ├─ 是 → 递增版本号 → 提交成功
    └─ 否 → 冲突 → 重试或回滚
```

**冲突处理策略：**
1. **自动重试**：配置 MaxRetries 和退避策略
2. **指数退避**：InitialRetryDelay * (RetryBackoffFactor ^ retryCount)
3. **失败回滚**：超过重试次数后返回错误

### 5.7.3 扩展性设计

**当前限制：**
1. 单节点事务，不支持分布式事务
2. 保存点为简化实现
3. 复杂约束支持有限

**未来扩展方向：**
1. **分布式事务**：支持两阶段提交（2PC）或三阶段提交（3PC）
2. **事务协调器**：引入专门的事务协调服务
3. **增强保存点**：支持完全回滚和部分提交
4. **事务监控**：提供事务执行指标和告警
5. **异构数据源**：支持跨不同存储引擎的事务

## 5.8 本章小结

本章深入探讨了 sfsDb 的无锁事务系统，主要内容包括：

✅ **ACID 特性**：原子性、一致性、隔离性、持久性的完整实现
✅ **乐观并发控制**：无锁设计，通过版本管理实现冲突检测
✅ **基本操作**：开始、提交、回滚事务的标准流程
✅ **批量操作**：批量插入、更新、删除的优化方法
✅ **嵌套事务**：保存点和多层嵌套事务的使用
✅ **性能优化**：事务大小、并发处理、基准测试
✅ **实战示例**：银行转账和订单处理的完整实现
✅ **架构师视角**：实现原理、冲突检测、扩展性设计

**关键要点：**
1. sfsDb 的无锁事务在高并发场景下性能显著优于传统有锁事务
2. 合理使用批量操作和重试策略可以大幅提升性能
3. 保持事务简短、避免热点数据是最佳实践
4. 根据业务需求选择合适的隔离级别，不要过度隔离

下一章我们将学习 sfsDb 的时序数据处理能力，这是工业物联网场景中的关键特性。
