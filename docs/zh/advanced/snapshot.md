# 快照与读一致性

## 1. 快照概述

快照是 sfsDb 提供的一种高级功能，用于创建数据库在某个时间点的一致性读视图。通过使用快照，您可以在数据持续变化的环境中，获取到数据库在快照创建时刻的完整、一致的数据状态。

### 1.1 快照的特点

- **读一致性**：快照提供数据库在创建时刻的一致性视图，不受后续数据修改的影响
- **无阻塞**：创建快照不会阻塞数据库的写操作
- **高效**：快照是轻量级的，创建和使用成本低
- **隔离性**：快照读取与数据库的其他操作完全隔离

### 1.2 快照与事务的区别

| 特性 | 快照 | 事务 |
|------|------|------|
| 主要用途 | 提供一致性读视图 | 确保操作的原子性、一致性、隔离性和持久性 |
| 写操作支持 | 仅支持读操作 | 支持读写操作 |
| 性能开销 | 低 | 高 |
| 实现方式 | 基于存储引擎的快照机制 | 基于锁或MVCC机制 |
| 适用场景 | 读多写少的场景，如报表生成、数据备份 | 对数据一致性要求高的场景，如金融交易 |

## 2. 使用快照

### 2.1 创建快照

```go
// 创建快照
snapshot, err := db.Snapshot()
if err != nil {
    panic(err)
}
defer snapshot.Release() // 确保释放快照资源
```

### 2.2 从快照读取数据

```go
// 从快照读取单个键值
value, err := snapshot.Get([]byte("key"))
if err != nil {
    panic(err)
}
fmt.Printf("Value: %s\n", value)

// 使用快照迭代器遍历数据
iter := snapshot.Iterator(nil, nil)
defer iter.Release()

for iter.Next() {
    key := iter.Key()
    val := iter.Value()
    fmt.Printf("Key: %s, Value: %s\n", key, val)
}
```

### 2.3 释放快照

快照使用完毕后，必须释放其资源：

```go
// 释放快照资源
err := snapshot.Release()
if err != nil {
    panic(err)
}
```

## 3. 读一致性示例

### 3.1 基本读一致性示例

```go
// 创建数据库
db, err := storage.NewLevelDBStore("./test_db", nil)
if err != nil {
    panic(err)
}
defer db.Close()

// 写入初始数据
db.Put([]byte("key1"), []byte("value1"))
db.Put([]byte("key2"), []byte("value2"))

// 创建快照
snapshot, err := db.Snapshot()
if err != nil {
    panic(err)
}
defer snapshot.Release()

// 修改数据
db.Put([]byte("key1"), []byte("new_value1"))
db.Put([]byte("key2"), []byte("new_value2"))

// 从数据库读取（获取到修改后的数据）
dbValue1, _ := db.Get([]byte("key1"))
dbValue2, _ := db.Get([]byte("key2"))
fmt.Printf("From DB: key1=%s, key2=%s\n", dbValue1, dbValue2)

// 从快照读取（获取到修改前的数据，保持一致性）
snapshotValue1, _ := snapshot.Get([]byte("key1"))
snapshotValue2, _ := snapshot.Get([]byte("key2"))
fmt.Printf("From Snapshot: key1=%s, key2=%s\n", snapshotValue1, snapshotValue2)

// 输出结果：
// From DB: key1=new_value1, key2=new_value2
// From Snapshot: key1=value1, key2=value2
```

### 3.2 工业边缘计算场景示例

在工业边缘计算场景中，设备数据持续变化，使用快照可以确保读取到某个时间点的完整数据：

```go
// 工业设备数据采集系统
func collectDeviceData() {
    // 1. 创建数据库连接
    db, err := storage.NewLevelDBStore("./device_data", nil)
    if err != nil {
        panic(err)
    }
    defer db.Close()

    // 2. 持续采集设备数据
    go func() {
        for {
            deviceID := "dev-" + strconv.Itoa(rand.Intn(1000))
            temperature := 20.0 + rand.Float64()*20.0
            humidity := 40.0 + rand.Float64()*40.0
            
            data := map[string]any{
                "temperature": temperature,
                "humidity": humidity,
                "timestamp": time.Now().UnixNano(),
            }
            
            // 序列化数据并存储
            dataBytes, _ := json.Marshal(data)
            db.Put([]byte(deviceID), dataBytes)
            
            time.Sleep(time.Millisecond * 100) // 模拟数据采集间隔
        }
    }()

    // 3. 定期生成设备状态报告（使用快照确保数据一致性）
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()

    for range ticker.C {
        fmt.Println("\n生成设备状态报告...")
        
        // 创建快照，获取当前时间点的一致数据视图
        snapshot, err := db.Snapshot()
        if err != nil {
            fmt.Printf("创建快照失败: %v\n", err)
            continue
        }
        
        // 遍历所有设备数据
        iter := snapshot.Iterator(nil, nil)
        deviceCount := 0
        totalTemperature := 0.0
        totalHumidity := 0.0
        
        for iter.Next() {
            deviceCount++
            dataBytes := iter.Value()
            var data map[string]any
            json.Unmarshal(dataBytes, &data)
            
            totalTemperature += data["temperature"].(float64)
            totalHumidity += data["humidity"].(float64)
            
            fmt.Printf("设备 %s: 温度=%.2f°C, 湿度=%.2f%%\n", iter.Key(), data["temperature"], data["humidity"])
        }
        
        // 计算平均值
        if deviceCount > 0 {
            avgTemperature := totalTemperature / float64(deviceCount)
            avgHumidity := totalHumidity / float64(deviceCount)
            fmt.Printf("\n报告摘要: 共%d台设备, 平均温度=%.2f°C, 平均湿度=%.2f%%\n", deviceCount, avgTemperature, avgHumidity)
        }
        
        // 释放资源
        iter.Release()
        snapshot.Release()
        fmt.Println("报告生成完成")
    }
}
```

### 3.3 表级快照读一致性示例

以下示例展示了如何在表级操作中使用快照确保读一致性，基于实际测试代码：

```go
// 测试表级快照读一致性
func TestTableSnapshotSearchConsistency(t *testing.T) {
    // 创建临时数据库
    dbPath := "./test_table_snapshot_db"
    cleanup := func() {
        os.RemoveAll(dbPath)
    }
    cleanup()
    defer cleanup()

    // 打开数据库
    dbManager := storage.GetDBManager()
    db, err := dbManager.OpenDB(dbPath)
    if err != nil {
        t.Fatalf("Failed to open database: %v", err)
    }
    defer storage.CloseDb()

    // 创建表
    tableName := "test_users"
    table, err := TableNew(tableName)
    if err != nil {
        t.Fatalf("Failed to create table: %v", err)
    }
    table.kvStore = db

    // 准备字段映射
    fieldsMap := map[string]any{
        "id":    0,
        "name":  "",
        "age":   0,
        "email": "",
    }

    // 设置表的字段
    err = table.SetFields(fieldsMap)
    if err != nil {
        t.Fatalf("Failed to set fields: %v", err)
    }

    // 创建主键索引
    PrimaryKeys, err := DefaultPrimaryKeyNew("pk")
    if err != nil {
        t.Fatalf("Failed to create primary key: %v", err)
    }
    PrimaryKeys.AddFields("id")
    err = table.CreateIndex(PrimaryKeys)
    if err != nil {
        t.Fatalf("Failed to create primary key index: %v", err)
    }

    // 为 age 字段创建普通索引，用于搜索
    ageIndex, err := DefaultNormalIndexNew("age_idx")
    if err != nil {
        t.Fatalf("Failed to create age index: %v", err)
    }
    ageIndex.AddFields("age")
    err = table.CreateIndex(ageIndex)
    if err != nil {
        t.Fatalf("Failed to create age index: %v", err)
    }

    // 1. 插入初始数据
    users := []map[string]any{
        {"id": 1, "name": "张三", "age": 30, "email": "zhangsan@example.com"},
        {"id": 2, "name": "李四", "age": 25, "email": "lisi@example.com"},
        {"id": 3, "name": "王五", "age": 35, "email": "wangwu@example.com"},
        {"id": 4, "name": "赵六", "age": 28, "email": "zhaoliu@example.com"},
        {"id": 5, "name": "孙七", "age": 40, "email": "sunqi@example.com"},
    }

    for _, user := range users {
        _, err := table.Insert(&user)
        if err != nil {
            t.Fatalf("Failed to insert user: %v", err)
        }
    }

    // 2. 创建快照
    snapshot, err := db.Snapshot()
    if err != nil {
        t.Fatalf("Failed to create snapshot: %v", err)
    }
    defer snapshot.Release()

    // 3. 修改数据
    updateUsers := []map[string]any{
        {"id": 1, "name": "张三(已更新)", "age": 31},
        {"id": 2, "name": "李四(已更新)", "age": 26},
        {"id": 6, "name": "周八", "age": 22, "email": "zhouba@example.com"}, // 新增记录
    }

    for _, user := range updateUsers {
        if user["id"] == 6 {
            // 新增记录
            _, err := table.Insert(&user)
            if err != nil {
                t.Fatalf("Failed to insert new user: %v", err)
            }
        } else {
            // 更新记录
            err := table.Update(&user)
            if err != nil {
                t.Fatalf("Failed to update user: %v", err)
            }
        }
    }

    // 4. 直接使用快照验证读一致性
    // 创建从快照中获取记录的函数
    snapshotGet := func(key []byte) []byte {
        value, err := snapshot.Get(key)
        if err != nil {
            return nil
        }
        return value
    }

    // 为每个用户ID构建主键并验证数据
    userIDs := []int{1, 2, 3, 4, 5}
    for _, id := range userIDs {
        // 构建主键
        pk := PrimaryKeys
        fieldsBytes := map[string][]byte{
            "id": util.AnyToBytes(id),
        }
        key := pk.JoinValue(&fieldsBytes, table.id)

        // 从快照中读取
        snapshotValue := snapshotGet(key)
        // 从数据库中读取
        dbValue := table.ReadByBytes(key)

        // 解析记录并验证
        // 快照应返回原始数据，数据库应返回修改后的数据
    }

    // 验证快照不包含新增记录（ID: 6）
    newRecordID := 6
    newRecordFieldsBytes := map[string][]byte{
        "id": util.AnyToBytes(newRecordID),
    }
    newRecordKey := PrimaryKeys.JoinValue(&newRecordFieldsBytes, table.id)
    newRecordSnapshotValue := snapshotGet(newRecordKey)
    if newRecordSnapshotValue != nil {
        t.Fatalf("Snapshot should not include new record with ID %d", newRecordID)
    }

    // 5. 测试快照迭代器
    // 创建使用快照迭代器的 FunIter 函数
    snapshotFunIter := func(start, limit []byte) storage.Iterator {
        return snapshot.Iterator(start, limit)
    }

    // 测试范围搜索
    rangeIter, err := table.SearchRange(snapshotFunIter, "id", 2, 5)
    defer rangeIter.Release()
    if err != nil {
        t.Fatalf("Failed to search range with snapshot: %v", err)
    }

    // 收集范围搜索结果
    rangeResults := rangeIter.GetRecords(true)
    defer record.PutRecords(rangeResults)

    // 验证范围搜索结果
    // 应返回原始数据，不受后续修改的影响
}
```

#### 示例说明

1. **创建和准备**：创建数据库、表、字段和索引
2. **插入初始数据**：插入5条用户记录
3. **创建快照**：在数据修改前创建快照
4. **修改数据**：更新2条现有记录，新增1条记录
5. **验证一致性**：
   - 从快照读取数据，应返回原始数据
   - 从数据库读取数据，应返回修改后的数据
   - 验证快照不包含新增的记录
6. **测试快照迭代器**：使用快照迭代器进行范围搜索，验证返回原始数据

#### 预期结果

- 快照数据：张三(30)、李四(25)（原始数据）
- 数据库数据：张三(已更新)(31)、李四(已更新)(26)（修改后数据）
- 快照中不包含新增记录（ID: 6）
- 范围搜索返回3条原始数据记录

这个示例展示了如何在实际应用中使用快照确保读一致性，特别是在数据持续变化的场景中。

## 3. 快照的应用场景

### 3.1 工业边缘计算

- **设备状态监控**：使用快照获取某个时间点的所有设备状态，确保监控数据的一致性
- **生产数据分析**：基于快照生成生产报表，避免数据在分析过程中发生变化
- **故障诊断**：在设备出现故障时，使用快照保存故障发生时的完整数据状态，便于后续分析

### 3.2 物联网场景

- **传感器数据聚合**：使用快照聚合多个传感器在同一时间点的数据
- **设备配置管理**：确保读取到设备的完整配置信息，不受配置更新的影响
- **数据同步**：在边缘设备与云端同步数据时，使用快照确保同步数据的一致性

### 3.3 其他场景

- **数据备份**：基于快照进行增量备份，避免备份过程中数据发生变化
- **报表生成**：生成财务报表、业务分析报表等需要数据一致性的场景
- **测试环境**：在测试过程中使用快照，确保测试数据的一致性和可重复性

## 4. 性能考量

### 4.1 快照的开销

- **创建开销**：快照创建操作非常轻量，几乎不影响数据库性能
- **存储开销**：快照共享底层存储，不会额外占用存储空间
- **读取开销**：从快照读取数据的性能与从数据库直接读取相近

### 4.2 最佳实践

- **及时释放**：快照使用完毕后应立即释放，避免占用资源
- **合理使用**：仅在需要一致性读视图的场景中使用快照
- **批量操作**：对于需要读取大量数据的场景，使用快照可以减少数据不一致的风险
- **避免嵌套**：不支持从快照创建快照，应避免尝试这种操作

## 5. 注意事项

1. **只读操作**：快照仅支持读操作，不支持写操作
2. **资源管理**：必须调用 `Release()` 方法释放快照资源，否则可能导致资源泄漏
3. **快照生命周期**：快照的生命周期应尽可能短，仅在需要时创建，使用完毕后立即释放
4. **嵌套限制**：不支持从快照创建快照，尝试这样做会返回错误
5. **存储引擎依赖**：快照功能依赖于底层存储引擎的快照支持，目前仅 LevelDB 存储引擎支持此功能

## 6. 总结

快照是 sfsDb 提供的一种强大功能，为工业边缘计算和物联网场景提供了高效的读一致性解决方案。通过使用快照，您可以：

- 在数据持续变化的环境中获取一致的数据视图
- 避免使用传统事务带来的性能开销
- 确保报表生成、数据备份等操作的数据一致性
- 提高系统的整体性能和可靠性

快照功能的实现基于 LevelDB 的快照机制，结合了对象池技术，既保证了性能，又确保了资源的有效管理。在实际应用中，合理使用快照可以显著提升系统的可靠性和性能。