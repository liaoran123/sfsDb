# 插入数据

## 3.1 基本插入

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

## 3.2 自动增值

sfsDb 支持主键自动增值功能，**字段名称必须是"id"**，且必须是单主键。系统会在以下两种情况下自动生成递增的主键值：

### 3.2.1 不提供id字段
当插入记录时不包含"id"字段，系统会自动生成并添加主键值：

```go
// 不指定 ID 字段，系统自动生成
user2 := map[string]any{
    "name": "李四",
    "age":  25,
    "email": "lisi@example.com",
}
id2, err := table.Insert(&user2)
if err != nil {
    panic(err)
}
fmt.Printf("自动生成 ID: %d\n", id2) // 输出: 2
```

### 3.2.2 提供id字段但值为nil
当插入记录时包含"id"字段但值为nil，系统同样会自动生成主键值：

```go
// 显式设置 ID 为 nil，同样会自动生成
user3 := map[string]any{
    "id":   nil, // 显式设置为 nil，同样会自动生成
    "name": "王五",
    "age":  30,
    "email": "wangwu@example.com",
}
id3, err := table.Insert(&user3)
if err != nil {
    panic(err)
}
fmt.Printf("自动生成 ID: %d\n", id3) // 输出: 3
```

### 3.2.3 自动增值的实现原理

自动增值功能的实现基于以下条件：

1. **字段名称必须是"id"**：系统只对名为"id"的字段提供自动增值功能
2. **必须是单主键**：自动增值仅适用于单主键表
3. **递增机制**：使用内部自增计数器确保主键值唯一且递增
4. **版本号自动添加**：系统会自动为每条记录添加版本号字段"v"，初始值为1

### 3.2.4 注意事项

- 自动增值功能仅在插入操作时生效
- 手动指定非nil的id值会覆盖自动生成行为
- 主键值类型为int，确保与表结构定义一致
- 系统会自动处理版本号字段，无需手动管理

## 3.3 批量插入

```go
// 批量插入多条记录
users := []map[string]any{
    {"name": "王五", "age": 35, "email": "wangwu@example.com"},
    {"name": "赵六", "age": 28, "email": "zhaoliu@example.com"},
    {"name": "孙七", "age": 40, "email": "sunqi@example.com"},
}

for _, user := range users {
    _, err := table.Insert(&user)
    if err != nil {
        panic(err)
    }
}
fmt.Println("普通批量插入完成")

// 手动事务批量插入（更高效的批量操作）
fmt.Println("\n手动事务批量插入:")

// 1. 获取批量操作对象（通过全局storage.KVDb获取）
batch := storage.KVDb.GetBatch()
if batch == nil {
    panic("无法获取批量操作对象")
}

// 2. 准备要插入的数据
batchUsers := []map[string]any{
    {"name": "周八", "age": 26, "email": "zhouba@example.com"},
    {"name": "吴九", "age": 32, "email": "wujiu@example.com"},
    {"name": "郑十", "age": 38, "email": "zhengshi@example.com"},
}

// 3. 将多个插入操作添加到同一个批处理中
for i, user := range batchUsers {
    _, err := table.Insert(&user, batch) // 传递batch参数，手动控制事务
    if err != nil {
        panic(fmt.Sprintf("插入第%d条记录失败: %v", i+1, err))
    }
    fmt.Printf("已添加第%d条记录到批处理\n", i+1)
}

// 4. 手动提交批处理（所有操作一次性执行）
err = storage.KVDb.WriteBatch(batch)
if err != nil {
    panic(fmt.Sprintf("批量提交失败: %v", err))
}
fmt.Println("手动事务批量插入完成")

// 5. 验证插入结果
fmt.Println("\n验证插入结果:")
allIter := table.ForData()
defer GlobalTableIterPool.Put(allIter)
records := allIter.GetRecords(true)
defer record.PutRecords(records)
fmt.Printf("表中共有%d条记录\n", len(records))

// 手动事务批量操作支持多种组合
fmt.Println("\n组合操作示例（插入+删除）:")
batch2 := storage.KVDb.GetBatch()
if batch2 == nil {
    panic("无法获取批量操作对象")
}

// 添加一条新记录
newUser := map[string]any{
    "name": "测试用户",
    "age":  25,
    "email": "test@example.com",
}
_, err = table.Insert(&newUser, batch2)
if err != nil {
    panic(err)
}

// 删除一条现有记录
deleteUser := map[string]any{
    "id": 1, // 删除ID为1的记录
}
err = table.Delete(&deleteUser, batch2)
if err != nil {
    panic(err)
}

// 提交组合操作
err = storage.KVDb.WriteBatch(batch2)
if err != nil {
    panic(fmt.Sprintf("组合操作提交失败: %v", err))
}
fmt.Println("组合操作完成")
```