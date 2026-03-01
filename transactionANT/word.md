# transactionLockANT 包详细文档

## 1. 包概述

transactionLockANT 是一个高性能、可靠的事务处理包，实现了企业级应用所需的完整事务功能。该包采用无锁设计，结合乐观并发控制，提供了高性能的事务处理能力，同时支持丰富的安全特性。

### 1.1 主要特性

- **完整的事务生命周期管理**：支持开始、提交、回滚事务
- **多种隔离级别**：支持 ReadUncommitted、ReadCommitted、RepeatableRead、Serializable
- **嵌套事务**：支持创建嵌套事务，实现更复杂的业务逻辑
- **保存点**：支持事务保存点，实现部分回滚
- **访问控制**：基于角色的访问控制（RBAC）系统
- **会话管理**：支持会话创建、验证和过期管理
- **审计日志**：详细的操作审计记录
- **数据加密**：支持数据加密，保护敏感信息
- **事务日志**：基于 WAL（Write-Ahead Logging）的事务日志
- **故障恢复**：系统崩溃后的事务恢复机制
- **高性能设计**：对象池、缓存机制、批量操作优化

### 1.2 适用场景

- **企业级应用**：需要可靠事务处理的企业系统
- **金融系统**：需要高安全性和可靠性的金融交易
- **电商系统**：订单处理、库存管理等场景
- **ERP系统**：复杂业务流程的事务处理
- **任何需要ACID特性的应用**：需要原子性、一致性、隔离性、持久性的场景

## 2. 快速开始

### 2.1 安装

```bash
go get github.com/liaoran123/sfsDb/transactionLockANT
```

### 2.2 基本使用示例

#### 2.2.1 初始化

```go
import (
    "github.com/liaoran123/sfsDb/transactionLockANT"
    "github.com/liaoran123/sfsDb/storage"
)

// 初始化存储
store, err := storage.NewLevelDBStore("testdb")
if err != nil {
    panic(err)
}

// 初始化WAL
if err := transactionLockANT.InitWAL(store, "testdb"); err != nil {
    panic(err)
}

// 初始化访问控制
if err := transactionLockANT.InitAccessControl(); err != nil {
    panic(err)
}

// 初始化会话管理
transactionLockANT.InitSessionManager(30 * 60 * time.Second)

// 初始化审计日志
transactionLockANT.InitAudit(store)

// 初始化加密管理器
encryptionConfig := &transactionLockANT.TransactionEncryptionConfig{
    Key: []byte("your-encryption-key"),
}
if err := transactionLockANT.InitEncryption(encryptionConfig); err != nil {
    panic(err)
}
```

#### 2.2.2 创建事务

```go
// 创建事务
userID := "user:admin" // 使用默认管理员用户
tx, err := transactionLockANT.NewTransaction(store, userID)
if err != nil {
    panic(err)
}

// 执行事务操作
tx.Put([]byte("key1"), []byte("value1"))
tx.Put([]byte("key2"), []byte("value2"))

// 提交事务
if err := tx.Commit(); err != nil {
    // 回滚事务
    tx.Rollback()
    panic(err)
}
```

#### 2.2.3 表事务

```go
import (
    "github.com/liaoran123/sfsDb/engine"
)

// 创建表
table, err := engine.NewTable("users", []engine.Field{
    {Name: "id", Type: engine.FieldTypeInt, PrimaryKey: true, AutoIncrement: true},
    {Name: "name", Type: engine.FieldTypeString},
    {Name: "email", Type: engine.FieldTypeString},
})
if err != nil {
    panic(err)
}

// 创建表事务
tx, err := transactionLockANT.NewTableTransaction(table, "user:admin")
if err != nil {
    panic(err)
}

// 插入记录
fields := map[string]interface{}{
    "name":  "John Doe",
    "email": "john@example.com",
}
id, err := tx.Insert(&fields)
if err != nil {
    tx.Rollback()
    panic(err)
}

// 提交事务
if err := tx.Commit(); err != nil {
    tx.Rollback()
    panic(err)
}

fmt.Printf("Inserted record with ID: %d\n", id)
```

## 3. 核心组件

### 3.1 事务管理

#### 3.1.1 事务接口

```go
type Transaction interface {
    // Get 从事务中获取值
    Get(key []byte) ([]byte, error)
    // Put 向事务中设置值
    Put(key []byte, value []byte)
    // Delete 从事务中删除值
    Delete(key []byte)
    // Commit 提交事务
    Commit() error
    // Rollback 回滚事务
    Rollback() error
    // BeginNested 创建一个嵌套事务
    BeginNested() (Transaction, error)
    // CreateSavepoint 创建保存点
    CreateSavepoint(name string) error
    // RollbackToSavepoint 回滚到保存点
    RollbackToSavepoint(name string) error
    // GetOptions 获取事务选项
    GetOptions() *TransactionOptions
    // GetTxID 获取事务ID
    GetTxID() uint64
}
```

#### 3.1.2 表事务接口

```go
type TableTransactionInterface interface {
    // Insert 在事务中插入记录
    Insert(fields *map[string]interface{}) (int, error)
    // Update 在事务中更新记录
    Update(fields *map[string]interface{}) error
    // Delete 在事务中删除记录
    Delete(fields *map[string]interface{}) error
    // Read 在事务中读取单条记录（支持读一致性）
    Read(fields *map[string]any) ([]byte, error)
    // Search 在事务中搜索记录（支持读一致性）
    Search(fields *map[string]any, ops ...util.ComparisonOperator) (*engine.TableIter, error)
    // Searchs 在事务中搜索记录（通过funIter支持原数据库或快照查询）
    Searchs(funIter storage.FunIter, fields *map[string]any, ops ...util.ComparisonOperator) (*engine.TableIter, error)
    // SearchRange 在事务中进行区间搜索（支持读一致性）
    SearchRange(funIter storage.FunIter, Start, Limit *map[string]any) (*engine.TableIter, error)
    // Commit 提交事务
    Commit() error
    // Rollback 回滚事务
    Rollback() error
    // BeginNested 创建一个嵌套事务
    BeginNested() (TableTransactionInterface, error)
    // CreateSavepoint 创建保存点
    CreateSavepoint(name string) error
    // RollbackToSavepoint 回滚到保存点
    RollbackToSavepoint(name string) error
    // GetOptions 获取事务选项
    GetOptions() *TransactionOptions
    // GetTxID 获取事务ID
    GetTxID() uint64
}
```

### 3.2 访问控制

#### 3.2.1 核心结构

```go
// Permission 权限结构体
type Permission struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    ResourceType string    `json:"resourceType"`
    ResourceID  string    `json:"resourceId"`
    Action      string    `json:"action"`
    CreatedAt   time.Time `json:"createdAt"`
}

// Role 角色结构体
type Role struct {
    ID          string       `json:"id"`
    Name        string       `json:"name"`
    Description string       `json:"description"`
    Permissions []*Permission `json:"permissions"`
    CreatedAt   time.Time    `json:"createdAt"`
}

// User 用户结构体
type User struct {
    ID        string    `json:"id"`
    Username  string    `json:"username"`
    Password  string    `json:"password"` // 实际应用中应该存储哈希值
    Roles     []*Role   `json:"roles"`
    CreatedAt time.Time `json:"createdAt"`
}
```

#### 3.2.2 权限检查

```go
// 检查用户是否具有指定权限
hasPermission, err := acm.CheckPermission(userID, ResourceTypeTable, "users", PermissionWrite)
if err != nil {
    panic(err)
}

if !hasPermission {
    // 权限不足
    return fmt.Errorf("permission denied")
}
```

### 3.3 会话管理

```go
// 创建会话
session, err := sm.CreateSession("user:admin")
if err != nil {
    panic(err)
}

// 获取会话
session, err := sm.GetSession(sessionID)
if err != nil {
    // 会话不存在或已过期
    return fmt.Errorf("invalid session")
}

// 验证会话
valid, err := sm.ValidateSession(sessionID)
if !valid {
    // 会话无效
    return fmt.Errorf("invalid session")
}

// 从会话获取用户ID
userID, err := sm.GetUserIDFromSession(sessionID)
```

### 3.4 审计日志

```go
// 记录审计日志
LogAudit(AuditActionCreate, AuditResourceTable, "users", userID, "Inserted new user", AuditStatusSuccess)

// 查询审计日志
logs, err := am.QueryLogs(startTime, endTime, userID, AuditActionCreate, AuditResourceTable)
```

## 4. 高级功能

### 4.1 事务管理器

事务管理器用于管理多个表的事务，确保它们在同一个batch中执行，保证原子性。

```go
// 创建事务管理器
batch := store.GetBatch()
manager := transactionLockANT.NewTransactionManager(batch, "user:admin")

// 添加表事务
tx1, err := manager.AddTable(table1)
tx2, err := manager.AddTable(table2)

// 执行操作
tx1.Insert(&fields1)
tx2.Update(&fields2)

// 提交所有事务
if err := manager.Commit(); err != nil {
    manager.Rollback()
    panic(err)
}
```

### 4.2 乐观更新

乐观更新操作，适用于复杂事务，提供更细粒度的冲突检测和自动重试机制。

```go
// 乐观更新，最多重试3次
err := tx.OptimisticUpdate(&fields, 3)
if err != nil {
    // 更新失败
    panic(err)
}
```

### 4.3 批量操作

批量操作，适用于复杂事务中的多个操作，通过一次Batch提交多个操作，减少磁盘I/O，提高性能。

```go
// 批量操作
operations := []func() error{
    func() error {
        return tx1.Insert(&fields1)
    },
    func() error {
        return tx2.Update(&fields2)
    },
    func() error {
        return tx3.Delete(&fields3)
    },
}

if err := tx.BatchOperations(operations); err != nil {
    panic(err)
}
```

## 5. 配置选项

### 5.1 事务选项

```go
// 默认事务选项
options := transactionLockANT.DefaultTransactionOptions()

// 自定义事务选项
options := &transactionLockANT.TransactionOptions{
    IsolationLevel:      transactionLockANT.Serializable, // 隔离级别
    AllowNested:         true,                           // 允许嵌套事务
    MaxRetries:          3,                              // 最大重试次数
    InitialRetryDelay:   10 * time.Millisecond,          // 初始重试延迟
    RetryBackoffFactor:  2.0,                            // 重试退避因子
}

// 使用自定义选项创建事务
tx, err := transactionLockANT.NewTransactionWithOptions(store, options, userID)
```

### 5.2 加密配置

```go
// 加密配置
config := &transactionLockANT.TransactionEncryptionConfig{
    Key: []byte("your-encryption-key"),
}

// 初始化加密管理器
if err := transactionLockANT.InitEncryption(config); err != nil {
    panic(err)
}
```

## 6. 性能优化

### 6.1 最佳实践

1. **使用对象池**：利用事务对象池减少内存分配
2. **批量操作**：使用批量操作减少磁盘I/O
3. **合理设置隔离级别**：根据业务需求选择合适的隔离级别
4. **使用保存点**：对于复杂事务，使用保存点实现部分回滚
5. **优化事务大小**：将大事务拆分为多个小事务
6. **使用乐观并发控制**：对于高并发场景，使用乐观更新机制

### 6.2 性能指标

| 测试场景 | 性能指标 |
|---------|---------|
| 并发转账 | 约19,629次/秒 |
| 并发订单创建 | 约24.66次/秒 |
| 执行时间 | 约50.94ms |
| 成功率 | 约98.3% |

## 7. 安全最佳实践

1. **使用访问控制**：为每个用户分配适当的角色和权限
2. **启用审计日志**：记录所有关键操作
3. **使用加密**：对敏感数据进行加密
4. **安全的会话管理**：设置合理的会话超时时间
5. **输入验证**：验证所有用户输入
6. **定期备份**：定期备份数据库和事务日志

## 8. 故障恢复

### 8.1 事务日志

transactionLockANT使用WAL（Write-Ahead Logging）机制，确保事务的持久性和可恢复性。当系统崩溃时，可以通过重放事务日志来恢复未提交的事务。

### 8.2 恢复过程

1. **系统启动时**：自动检查和恢复未完成的事务
2. **重放日志**：按照日志顺序重放事务操作
3. **提交或回滚**：根据事务状态决定提交或回滚
4. **清理日志**：定期清理已完成的事务日志

## 9. API参考

### 9.1 事务创建

- `NewTransaction(store storage.Store, userID string) (*SfsTransaction, error)`：创建新的事务
- `NewTransactionWithOptions(store storage.Store, options *TransactionOptions, userID string) (*SfsTransaction, error)`：使用指定选项创建新的事务
- `NewTableTransaction(table *engine.Table, userID string) (*TableTransaction, error)`：为指定表创建事务
- `NewTableTransactionWithOptions(table *engine.Table, options *TransactionOptions, userID string) (*TableTransaction, error)`：为指定表创建带选项的事务

### 9.2 访问控制

- `InitAccessControl() error`：初始化访问控制管理器
- `GetAccessControlManager() *AccessControlManager`：获取访问控制管理器
- `(acm *AccessControlManager) AddUser(user *User) error`：添加用户
- `(acm *AccessControlManager) GetUser(userID string) (*User, error)`：获取用户
- `(acm *AccessControlManager) AddRole(role *Role) error`：添加角色
- `(acm *AccessControlManager) GetRole(roleID string) (*Role, error)`：获取角色
- `(acm *AccessControlManager) AddPermission(permission *Permission) error`：添加权限
- `(acm *AccessControlManager) GetPermission(permissionID string) (*Permission, error)`：获取权限
- `(acm *AccessControlManager) AssignRoleToUser(userID, roleID string) error`：为用户分配角色
- `(acm *AccessControlManager) RevokeRoleFromUser(userID, roleID string) error`：从用户撤销角色
- `(acm *AccessControlManager) AssignPermissionToRole(roleID, permissionID string) error`：为角色分配权限
- `(acm *AccessControlManager) RevokePermissionFromRole(roleID, permissionID string) error`：从角色撤销权限
- `(acm *AccessControlManager) CheckPermission(userID, resourceType, resourceID, action string) (bool, error)`：检查用户是否具有指定权限

### 9.3 会话管理

- `InitSessionManager(timeout time.Duration) error`：初始化会话管理器
- `GetSessionManager() *SessionManager`：获取会话管理器
- `(sm *SessionManager) CreateSession(userID string) (*Session, error)`：创建会话
- `(sm *SessionManager) GetSession(sessionID string) (*Session, error)`：获取会话
- `(sm *SessionManager) RefreshSession(sessionID string) error`：刷新会话
- `(sm *SessionManager) DeleteSession(sessionID string) error`：删除会话
- `(sm *SessionManager) ValidateSession(sessionID string) (bool, error)`：验证会话
- `(sm *SessionManager) GetUserIDFromSession(sessionID string) (string, error)`：从会话获取用户ID

### 9.4 审计日志

- `InitAudit(store storage.Store) error`：初始化审计日志管理器
- `GetAuditManager() *AuditManager`：获取审计日志管理器
- `IsAuditEnabled() bool`：检查审计日志是否启用
- `EnableAudit()`：启用审计日志
- `DisableAudit()`：禁用审计日志
- `(am *AuditManager) Log(log *AuditLog) error`：记录审计日志
- `LogAudit(action, resource, resourceID, userID, details, status string) error`：便捷函数：记录审计日志
- `(am *AuditManager) QueryLogs(startTime, endTime time.Time, userID, action, resource string) ([]*AuditLog, error)`：查询审计日志
- `(am *AuditManager) GetAuditLog(logID string) (*AuditLog, error)`：获取单个审计日志

### 9.5 事务管理

- `NewTransactionManager(batch storage.Batch, userID string) *TransactionManager`：创建事务管理器
- `NewTransactionManagerWithOptions(batch storage.Batch, options *TransactionOptions, userID string) *TransactionManager`：使用指定选项创建事务管理器
- `(m *TransactionManager) AddTable(table *engine.Table) (*TableTransaction, error)`：添加表到事务管理器
- `(m *TransactionManager) GetTableTransaction(table *engine.Table) (*TableTransaction, error)`：获取指定表的事务
- `(m *TransactionManager) BeginNested() (*TransactionManager, error)`：创建嵌套事务
- `(m *TransactionManager) CreateSavepoint(name string) error`：创建事务保存点
- `(m *TransactionManager) RollbackToSavepoint(name string) error`：回滚到保存点
- `(m *TransactionManager) Commit() error`：提交所有事务
- `(m *TransactionManager) Rollback() error`：回滚所有事务

## 10. 故障排除

### 10.1 常见错误

| 错误信息 | 可能原因 | 解决方案 |
|---------|---------|--------|
| `permission denied` | 权限不足 | 检查用户权限和角色分配 |
| `session expired` | 会话已过期 | 重新登录获取新会话 |
| `concurrency conflict` | 并发冲突 | 使用乐观更新或增加重试次数 |
| `serialization conflict` | 可串行化冲突 | 重试事务或使用更低的隔离级别 |
| `transaction already committed` | 事务已提交 | 确保只提交一次事务 |

### 10.2 调试技巧

1. **启用审计日志**：查看详细的操作记录
2. **检查事务日志**：分析WAL日志文件
3. **使用压力测试**：模拟高并发场景
4. **监控系统资源**：检查CPU、内存和磁盘使用情况
5. **使用pprof**：分析性能瓶颈

## 11. 版本兼容性

### 11.1 向后兼容性

transactionLockANT包设计为向后兼容，支持以下场景：

- **空用户ID**：允许使用空用户ID创建事务，此时不进行权限检查
- **未初始化组件**：如果未初始化访问控制、会话管理或审计日志，系统会自动降级，不影响核心事务功能

### 11.2 版本历史

| 版本 | 主要变更 |
|------|---------|
| 1.0.0 | 初始版本，实现基本事务功能 |
| 1.1.0 | 添加访问控制和会话管理 |
| 1.2.0 | 添加审计日志和加密支持 |
| 1.3.0 | 性能优化和故障恢复增强 |

## 12. 总结

transactionLockANT包是一个功能完整、性能优异的事务处理解决方案，适合企业级应用的各种场景。它提供了完整的事务生命周期管理、丰富的安全特性和高性能的实现，同时保持了良好的向后兼容性。

通过合理的配置和使用最佳实践，您可以充分利用transactionLockANT包的优势，构建可靠、安全、高性能的企业级应用。

## 13. 示例应用

### 13.1 金融交易系统

```go
// 初始化
store, _ := storage.NewLevelDBStore("financial_db")
transactionLockANT.InitWAL(store, "financial_db")
transactionLockANT.InitAccessControl()
transactionLockANT.InitSessionManager(30 * time.Minute)
transactionLockANT.InitAudit(store)

// 处理转账
transferFunds := func(fromAccount, toAccount string, amount float64, userID string) error {
    // 创建事务
    tx, err := transactionLockANT.NewTransaction(store, userID)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // 读取账户余额
    fromBalance, err := getAccountBalance(tx, fromAccount)
    if err != nil {
        return err
    }

    if fromBalance < amount {
        return fmt.Errorf("insufficient funds")
    }

    // 扣减转出账户余额
    if err := updateAccountBalance(tx, fromAccount, fromBalance-amount); err != nil {
        return err
    }

    // 增加转入账户余额
    toBalance, err := getAccountBalance(tx, toAccount)
    if err != nil {
        return err
    }

    if err := updateAccountBalance(tx, toAccount, toBalance+amount); err != nil {
        return err
    }

    // 记录交易
    if err := recordTransaction(tx, fromAccount, toAccount, amount); err != nil {
        return err
    }

    // 提交事务
    return tx.Commit()
}

// 调用转账函数
err := transferFunds("account1", "account2", 100.0, "user:admin")
if err != nil {
    fmt.Printf("Transfer failed: %v\n", err)
} else {
    fmt.Println("Transfer successful")
}
```

### 13.2 电商订单系统

```go
// 初始化
store, _ := storage.NewLevelDBStore("ecommerce_db")
transactionLockANT.InitWAL(store, "ecommerce_db")
transactionLockANT.InitAccessControl()
transactionLockANT.InitSessionManager(30 * time.Minute)
transactionLockANT.InitAudit(store)

// 创建订单
createOrder := func(userID string, items []OrderItem) (int, error) {
    // 创建事务管理器
    batch := store.GetBatch()
    manager := transactionLockANT.NewTransactionManager(batch, userID)
    defer manager.Rollback()

    // 添加表事务
    orderTable, _ := engine.NewTable("orders", []engine.Field{
        {Name: "id", Type: engine.FieldTypeInt, PrimaryKey: true, AutoIncrement: true},
        {Name: "user_id", Type: engine.FieldTypeString},
        {Name: "total_amount", Type: engine.FieldTypeFloat},
        {Name: "status", Type: engine.FieldTypeString},
        {Name: "created_at", Type: engine.FieldTypeString},
    })

    orderItemTable, _ := engine.NewTable("order_items", []engine.Field{
        {Name: "id", Type: engine.FieldTypeInt, PrimaryKey: true, AutoIncrement: true},
        {Name: "order_id", Type: engine.FieldTypeInt},
        {Name: "product_id", Type: engine.FieldTypeInt},
        {Name: "quantity", Type: engine.FieldTypeInt},
        {Name: "price", Type: engine.FieldTypeFloat},
    })

    productTable, _ := engine.NewTable("products", []engine.Field{
        {Name: "id", Type: engine.FieldTypeInt, PrimaryKey: true},
        {Name: "name", Type: engine.FieldTypeString},
        {Name: "stock", Type: engine.FieldTypeInt},
        {Name: "price", Type: engine.FieldTypeFloat},
    })

    orderTx, _ := manager.AddTable(orderTable)
    orderItemTx, _ := manager.AddTable(orderItemTable)
    productTx, _ := manager.AddTable(productTable)

    // 计算总金额
    var totalAmount float64
    for _, item := range items {
        // 检查库存
        productFields := map[string]interface{}{"id": item.ProductID}
        productData, _ := productTx.Read(&productFields)
        product := parseProductData(productData)

        if product.Stock < item.Quantity {
            return 0, fmt.Errorf("insufficient stock for product %d", item.ProductID)
        }

        // 扣减库存
        product.Stock -= item.Quantity
        updateFields := map[string]interface{}{"id": item.ProductID, "stock": product.Stock}
        productTx.Update(&updateFields)

        totalAmount += product.Price * float64(item.Quantity)
    }

    // 创建订单
    orderFields := map[string]interface{}{
        "user_id":     userID,
        "total_amount": totalAmount,
        "status":      "pending",
        "created_at":  time.Now().Format(time.RFC3339),
    }
    orderID, _ := orderTx.Insert(&orderFields)

    // 创建订单项
    for _, item := range items {
        orderItemFields := map[string]interface{}{
            "order_id":   orderID,
            "product_id": item.ProductID,
            "quantity":   item.Quantity,
            "price":      item.Price,
        }
        orderItemTx.Insert(&orderItemFields)
    }

    // 提交事务
    manager.Commit()
    return orderID, nil
}

// 调用创建订单函数
items := []OrderItem{
    {ProductID: 1, Quantity: 2, Price: 99.99},
    {ProductID: 2, Quantity: 1, Price: 199.99},
}

orderID, err := createOrder("user:123", items)
if err != nil {
    fmt.Printf("Order creation failed: %v\n", err)
} else {
    fmt.Printf("Order created successfully with ID: %d\n", orderID)
}
```

## 14. 贡献指南

### 14.1 代码风格

- 遵循Go语言标准代码风格
- 使用`go fmt`格式化代码
- 编写清晰的注释
- 保持函数简洁，每个函数只做一件事

### 14.2 测试

- 编写单元测试和集成测试
- 运行所有测试确保代码质量
- 使用基准测试验证性能

### 14.3 提交规范

- 提交消息清晰明了
- 描述变更的目的和影响
- 包含相关的issue编号

## 15. 许可证

transactionLockANT包采用MIT许可证，详见LICENSE文件。

## 16. 联系方式

如有问题或建议，请联系：

- 项目地址：https://github.com/liaoran123/sfsDb
- 邮箱：liaoran123@example.com

---

本文档详细介绍了transactionLockANT包的功能、使用方法和最佳实践，希望能帮助您充分利用这个强大的事务处理工具。