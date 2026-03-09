# 第 3 章：基础 CRUD 操作

本章将深入学习 sfsDb 的基础 CRUD（创建、读取、更新、删除）操作。这是使用 sfsDb 最核心的技能。

## 3.1 表的创建与管理

### 3.1.1 创建表

**基本表创建：**

```go
package main

import (
    "log"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    dbManager := storage.GetDBManager()
    db, err := dbManager.OpenDB("./my_database")
    if err != nil {
        log.Fatalf("打开数据库失败: %v", err)
    }
    defer dbManager.CloseAllDB()

    // 创建用户表
    userTable, err := db.CreateTable("users", map[string]string{
        "id":       "int",
        "username": "string",
        "email":    "string",
        "age":      "int",
        "created_at": "int64",
    })
    if err != nil {
        log.Fatalf("创建用户表失败: %v", err)
    }

    log.Println("用户表创建成功")
}
```

**字段类型支持：**

| 类型 | 说明 | 示例 |
|------|------|------|
| `int` | 整数类型 | 1, 100, -50 |
| `float64` | 浮点数类型 | 3.14, 100.5 |
| `string` | 字符串类型 | "hello", "张三" |
| `int64` | 64位整数（常用于时间戳） | 1620000000 |

### 3.1.2 设置字段

**添加主键和索引：**

```go
// 添加主键索引（唯一）
userTable.AddIndex("id", true)

// 添加普通索引
userTable.AddIndex("username", false)

// 添加复合索引
userTable.AddIndex("email,age", false)
```

**AddIndex 方法参数：**
- 第一个参数：索引字段名（多个字段用逗号分隔）
- 第二个参数：是否为唯一索引（true=唯一，false=普通）

### 3.1.3 表的删除

```go
// 删除表
err := db.DropTable("users")
if err != nil {
    log.Fatalf("删除表失败: %v", err)
}
```

**获取表列表：

```go
// 获取所有表
tables := db.GetTables()
for _, table := range tables {
    log.Printf("表名: %s", table.GetName())
}
```

## 3.2 数据插入

### 3.2.1 单条插入

```go
// 插入单条记录
fields := map[string]interface{}{
    "id":         1,
    "username":   "zhangsan",
    "email":      "zhangsan@example.com",
    "age":        25,
    "created_at": time.Now().Unix(),
}

id, err := userTable.Insert(&fields)
if err != nil {
    log.Fatalf("插入数据失败: %v", err)
}

log.Printf("插入成功，记录ID: %d", id)
```

### 3.2.2 批量插入

```go
// 批量插入用户
users := []map[string]interface{}{
    {
        "id":         2,
        "username":   "lisi",
        "email":      "lisi@example.com",
        "age":        30,
        "created_at": time.Now().Unix(),
    },
    {
        "id":         3,
        "username":   "wangwu",
        "email":      "wangwu@example.com",
        "age":        28,
        "created_at": time.Now().Unix(),
    },
}

for _, user := range users {
    _, err := userTable.Insert(&user)
    if err != nil {
        log.Printf("插入用户失败: %v", err)
    }
}
```

**使用事务批量插入（推荐）：

```go
import "github.com/liaoran123/sfsDb/transactionLockFree"

tx, err := transactionLockFree.NewTableTransaction(userTable)
if err != nil {
    log.Fatal(err)
}
defer tx.Rollback()

for _, user := range users {
    _, err := tx.Insert(&user)
    if err != nil {
        log.Fatal(err)
    }
}

err = tx.Commit()
if err != nil {
    log.Fatal(err)
}
```

### 3.2.3 自动增值 ID

```go
// sfsDb 支持自增ID
// 可以手动管理ID
nextID := 1

// 或者使用表的自动递增功能
// （具体实现根据你的业务逻辑
```

## 3.3 数据查询

### 3.3.1 主键查询

```go
// 主键查询
searchFields := map[string]any{
    "id": 1,
}

data, err := userTable.Read(&searchFields)
if err != nil {
    log.Fatalf("查询失败: %v", err)
}

log.Printf("查询结果: %s", string(data))
```

### 3.3.2 条件查询

```go
import "github.com/liaoran123/sfsDb/util"

// 条件查询 - 大于
iter, err := userTable.Search(
    &map[string]any{"age": 25},
    util.GreaterThan,
)
if err != nil {
    log.Fatalf("查询失败: %v", err)
}
defer iter.Close()

// 遍历结果
for iter.Next() {
    key, value := iter.Key(), iter.Value()
    log.Printf("Key: %s, Value: %s", string(key), string(value))
}
```

**比较操作符：**

| 操作符 | 说明 |
|--------|------|
| `Equal` | 等于 |
| `NotEqual` | 不等于 |
| `GreaterThan` | 大于 |
| `GreaterOrEqual` | 大于等于 |
| `LessThan` | 小于 |
| `LessOrEqual` | 小于等于 |

### 3.3.3 查询所有数据

```go
// 查询所有数据
iter, err := userTable.SearchAll()
if err != nil {
    log.Fatalf("查询失败: %v", err)
}
defer iter.Close()

// 遍历所有记录
count := 0
for iter.Next() {
    key, value := iter.Key(), iter.Value()
    log.Printf("记录 %d: Key=%s, Value=%s", count, string(key), string(value))
    count++
}

log.Printf("共查询到 %d 条记录", count)
```

## 3.4 数据更新

### 3.4.1 更新单条记录

```go
// 更新记录
updateFields := map[string]interface{}{
    "id":   1,
    "age":  26,
    "email": "zhangsan_new@example.com",
}

err := userTable.Update(&updateFields)
if err != nil {
    log.Fatalf("更新失败: %v", err)
}

log.Println("记录更新成功")
```

### 3.4.2 批量更新

```go
// 使用事务批量更新
tx, err := transactionLockFree.NewTableTransaction(userTable)
if err != nil {
    log.Fatal(err)
}
defer tx.Rollback()

updates := []map[string]interface{}{
    {"id": 1, "age": 26},
    {"id": 2, "age": 31},
    {"id": 3, "age": 29},
}

for _, update := range updates {
    err := tx.Update(&update)
    if err != nil {
        log.Fatal(err)
    }
}

err = tx.Commit()
if err != nil {
    log.Fatal(err)
}
```

## 3.5 数据删除

### 3.5.1 删除单条记录

```go
// 删除记录
deleteFields := map[string]interface{}{
    "id": 1,
}

err := userTable.Delete(&deleteFields)
if err != nil {
    log.Fatalf("删除失败: %v", err)
}

log.Println("记录删除成功")
```

### 3.5.2 批量删除

```go
// 使用事务批量删除
tx, err := transactionLockFree.NewTableTransaction(userTable)
if err != nil {
    log.Fatal(err)
}
defer tx.Rollback()

deleteIDs := []int{2, 3, 4, 5}

for _, id := range deleteIDs {
    deleteFields := map[string]interface{}{"id": id}
    err := tx.Delete(&deleteFields)
    if err != nil {
        log.Printf("删除ID %d 失败: %v", id, err)
    }
}

err = tx.Commit()
if err != nil {
    log.Fatal(err)
}
```

## 3.6 实战示例

### 3.6.1 完整的用户管理系统

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/liaoran123/sfsDb/storage"
    "github.com/liaoran123/sfsDb/transactionLockFree"
)

type User struct {
    ID        int
    Username  string
    Email     string
    Age       int
    CreatedAt int64
}

type UserManager struct {
    dbManager *storage.DBManager
    userTable *engine.Table
}

func NewUserManager(dbPath string) (*UserManager, error) {
    dbManager := storage.GetDBManager()
    db, err := dbManager.OpenDB(dbPath)
    if err != nil {
        return nil, err
    }

    userTable, err := db.CreateTable("users", map[string]string{
        "id":         "int",
        "username":   "string",
        "email":      "string",
        "age":        "int",
        "created_at": "int64",
    })
    if err != nil {
        return nil, err
    }

    userTable.AddIndex("id", true)
    userTable.AddIndex("username", false)

    return &UserManager{
        dbManager: dbManager,
        userTable: userTable,
    }, nil
}

func (um *UserManager) AddUser(user *User) (int, error) {
    fields := map[string]interface{}{
        "id":         user.ID,
        "username":   user.Username,
        "email":      user.Email,
        "age":        user.Age,
        "created_at": user.CreatedAt,
    }
    return um.userTable.Insert(&fields)
}

func (um *UserManager) GetUser(id int) (*User, error) {
    searchFields := map[string]any{"id": id}
    data, err := um.userTable.Read(&searchFields)
    if err != nil {
        return nil, err
    }

    // 解析数据...
    return &User{}, nil
}

func (um *UserManager) UpdateUser(user *User) error {
    fields := map[string]interface{}{
        "id":    user.ID,
        "email": user.Email,
        "age":   user.Age,
    }
    return um.userTable.Update(&fields)
}

func (um *UserManager) DeleteUser(id int) error {
    fields := map[string]interface{}{"id": id}
    return um.userTable.Delete(&fields)
}

func (um *UserManager) Close() {
    um.dbManager.CloseAllDB()
}

func main() {
    um, err := NewUserManager("./user_manager_db")
    if err != nil {
        log.Fatal(err)
    }
    defer um.Close()

    // 添加用户
    user := &User{
        ID:        1,
        Username:  "admin",
        Email:     "admin@example.com",
        Age:       30,
        CreatedAt: time.Now().Unix(),
    }
    id, err := um.AddUser(user)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("添加用户成功，ID: %d\n", id)

    // 查询用户
    retrievedUser, err := um.GetUser(1)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("查询到用户: %+v\n", retrievedUser)

    // 更新用户
    user.Age = 31
    err = um.UpdateUser(user)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("用户更新成功")

    // 删除用户
    err = um.DeleteUser(1)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("用户删除成功")
}
```

### 3.6.2 错误处理最佳实践

```go
// 总是检查错误
user, err := um.GetUser(1)
if err != nil {
    log.Printf("查询用户失败: %v", err)
    // 根据错误类型进行处理
    return
}

// 使用 defer 确保资源释放
dbManager := storage.GetDBManager()
db, err := dbManager.OpenDB("./mydb")
if err != nil {
    log.Fatal(err)
}
defer dbManager.CloseAllDB()

// 事务错误处理
tx, err := transactionLockFree.NewTableTransaction(table)
if err != nil {
    log.Fatal(err)
}
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
        panic(r)
    }
}()

// 执行操作...

err = tx.Commit()
if err != nil {
    tx.Rollback()
    log.Fatalf("提交失败: %v", err)
}
```

## 3.7 本章小结

本章学习了 sfsDb 的基础 CRUD 操作，包括：

✅ **表管理**：创建表、设置字段、删除表
✅ **数据插入**：单条插入、批量插入
✅ **数据查询**：主键查询、条件查询、查询所有
✅ **数据更新**：单条更新、批量更新
✅ **数据删除**：单条删除、批量删除
✅ **实战示例**：完整的用户管理系统
✅ **最佳实践**：错误处理和资源管理

**关键要点：**
1. 总是检查错误，妥善处理异常情况
2. 使用 defer 确保资源正确释放
3. 批量操作使用事务可以显著提升性能
4. 合理使用索引可以大幅提高查询效率
5. 封装数据访问层可以让代码更清晰易维护

下一章我们将学习索引与查询优化，这是提升数据库性能的关键。
