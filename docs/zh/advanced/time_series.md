# 时序数据处理

## 概述

time 包是 sfsDb 提供的时序数据处理工具，用于处理时间相关的数据操作，包括时间粒度处理、时间窗口计算、数据聚合和时间戳转换等功能。

这个包特别适合以下场景：
- IoT 设备传感器数据的处理和分析
- 系统监控数据的聚合和趋势分析
- 金融交易数据的时间序列分析
- 任何需要处理和分析时序数据的应用场景

## 核心功能

### 时间粒度处理

时间粒度处理允许您以不同的精度处理时间数据，支持以下时间粒度：

- **秒** (second)
- **分钟** (minute)
- **小时** (hour)
- **天** (day)
- **月** (month)
- **年** (year)

使用 `FormatTimeByGranularity` 函数可以将时间戳格式化为指定粒度的时间字符串：

```go
import (
    "time"
    sfsTime "github.com/liaoran123/sfsDb/time"
)

// 格式化时间
now := time.Now()
hourStr := sfsTime.FormatTimeByGranularity(now, sfsTime.TimeGranularityHour)
fmt.Println("Hour:", hourStr) // 输出: Hour: 2024-12-25 14:00:00
```

### 时间窗口计算

时间窗口计算根据指定的时间粒度计算时间范围，返回起始时间和结束时间：

```go
// 计算时间范围
now := time.Now()
start, end := sfsTime.TimeRange(now, sfsTime.TimeGranularityDay)
fmt.Println("Day range:", start, "to", end)
```

### 数据聚合

数据聚合按时间粒度对数据进行聚合，支持多种聚合类型：

- **求和** (sum)
- **平均值** (avg)
- **计数** (count)
- **最大值** (max)
- **最小值** (min)

使用 `AggregateByTimeGranularity` 函数进行数据聚合：

```go
// 假设我们有一组传感器数据记录
records := record.Records{...}

// 按小时聚合
results, err := sfsTime.AggregateByTimeGranularity(
    records,
    "timestamp",    // 时间字段
    "value",        // 值字段
    sfsTime.TimeGranularityHour,  // 时间粒度
    "sum",          // 聚合类型
)

// 输出结果
for _, result := range results {
    fmt.Printf("Time: %s, Sum: %.2f\n", result.TimeKey, result.Value)
}
```

### 时间戳转换

时间戳转换功能允许在 `time.Time` 对象和整数时间戳之间进行转换，支持不同精度的时间戳：

- **秒级**时间戳
- **毫秒级**时间戳
- **纳秒级**时间戳

```go
// 时间对象转整数时间戳
now := time.Now()
unixTime := sfsTime.TimeToUnixTimestamp(now)         // 秒级
unixTimeMs := sfsTime.TimeToUnixTimestampMs(now)     // 毫秒级
unixTimeNs := sfsTime.TimeToUnixTimestampNs(now)     // 纳秒级

// 整数时间戳转时间对象
timeObj := sfsTime.UnixTimestampToTime(unixTime)     // 秒级
timeObjMs := sfsTime.UnixTimestampMsToTime(unixTimeMs) // 毫秒级
timeObjNs := sfsTime.UnixTimestampNsToTime(unixTimeNs) // 纳秒级
```

### 时间桶分配

时间桶分配将时间戳分配到指定粒度的时间桶，返回时间桶的起始时间：

```go
// 时间桶分配
now := time.Now()
bucket := sfsTime.TimeBucket(now, time.Hour)
fmt.Println("Bucket:", bucket) // 输出时间桶的起始时间
```

## 与数据库集成

### 基本集成示例

以下是 time 包与 sfsDb 数据库集成的基本示例：

```go
// 创建表
sensorTable, err := engine.TableNew("sensor_data")
if err != nil {
    return err
}

// 设置字段
fields := map[string]any{
    "timestamp": time.Now(), // 时间戳，time.Time 类型
    "value":     0.0,        // 传感器值
    "sensor_id": "",         // 传感器ID
}
err = sensorTable.SetFields(fields)
if err != nil {
    return err
}

// 创建主键索引
primaryKey, err := engine.DefaultPrimaryKeyNew("pk")
if err != nil {
    return err
}
primaryKey.AddFields("timestamp")
err = sensorTable.CreateIndex(primaryKey)
if err != nil {
    return err
}

// 插入测试数据
now := time.Now()
for i := 0; i < 10; i++ {
    data := map[string]any{
        "timestamp": now.Add(time.Duration(i) * time.Minute),
        "value":     float64(i * 10),
        "sensor_id": fmt.Sprintf("sensor_%d", i%3),
    }
    _, err := sensorTable.Insert(&data)
    if err != nil {
        return err
    }
}

// 查询最近1小时的数据
startTime := time.Now().Add(-1 * time.Hour)
endTime := time.Now()
iter, err := sensorTable.SearchRange(sensorTable.kvStore.Iterator, "timestamp", startTime, endTime)
if err != nil {
    return err
}
defer engine.GlobalTableIterPool.Put(iter)

// 获取结果
records := iter.GetRecordSet(true)
defer record.PutRecords(records)

// 按小时聚合
aggregationResults, err := sfsTime.AggregateByTimeGranularity(
    records,
    "timestamp",
    "value",
    sfsTime.TimeGranularityHour,
    "avg",
)
if err != nil {
    return err
}

// 输出聚合结果
fmt.Println("按小时聚合的平均传感器值：")
for _, result := range aggregationResults {
    fmt.Printf("时间: %s, 平均值: %.2f\n", result.TimeKey, result.Value)
}
```

### 复合主键示例

对于 IoT 场景，通常需要使用复合主键（如 `sensor_id` + `timestamp`）来唯一标识记录：

```go
// 创建表
deviceTable, err := engine.TableNew("device_data")
if err != nil {
    return err
}

// 设置字段
fields := map[string]any{
    "sensor_id": "",         // 传感器ID
    "timestamp": time.Now(), // 时间戳
    "value":     0.0,        // 传感器值
}
err = deviceTable.SetFields(fields)
if err != nil {
    return err
}

// 创建复合主键索引
primaryKey, err := engine.DefaultPrimaryKeyNew("device_time_idx")
if err != nil {
    return err
}
primaryKey.AddFields("sensor_id")
primaryKey.AddFields("timestamp")
err = deviceTable.CreateIndex(primaryKey)
if err != nil {
    return err
}

// 插入数据和查询逻辑类似上面的示例
```

## 最佳实践

### 1. 时间字段类型选择

- **推荐使用 `time.Time` 类型**：与 time 包无缝集成，支持完整的时间功能
- **使用整数时间戳**：如果对存储效率有特殊要求，或需要与其他系统集成

### 2. 索引设计

- **时间戳作为主键**：适合单设备的时序数据
- **复合主键**：对于多设备场景，使用 `(device_id, timestamp)` 复合主键
- **辅助索引**：为常用查询字段创建辅助索引

### 3. 性能优化

- **批量插入**：使用批量操作减少数据库交互
- **数据降采样**：对于长期存储的数据，考虑按更大的时间粒度进行降采样
- **时间范围查询**：使用 `SearchRange` 方法进行高效的时间范围查询
- **内存管理**：及时释放迭代器和记录集合资源

### 4. 数据管理

- **数据过期策略**：为时序数据设置合理的过期时间
- **数据归档**：将历史数据归档到低成本存储
- **备份策略**：定期备份重要的时序数据

## 总结

time 包为 sfsDb 提供了强大的时序数据处理能力，使您能够：

- 以不同时间粒度处理和分析数据
- 执行时间窗口计算和数据聚合
- 在不同时间表示形式之间进行转换
- 与数据库无缝集成，构建完整的时序数据解决方案

通过合理使用 time 包的功能，您可以构建高性能、可靠的时序数据处理系统，满足各种应用场景的需求。