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

- **毫秒** (millisecond)
- **微秒** (microsecond)
- **秒** (second)
- **分钟** (minute)
- **小时** (hour)
- **天** (day)
- **周** (week)
- **月** (month)
- **季度** (quarter)
- **年** (year)

此外，time 包还提供了处理时间模式的工具函数：

- **IsWeekday**：检查给定时间是否为工作日
- **IsWeekend**：检查给定时间是否为周末
- **GetQuarter**：获取给定时间所在的季度
- **GetWeekNumber**：获取给定时间所在的周数（一年中的第几周）

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

时间窗口计算根据指定的时间粒度计算时间范围，返回起始时间和结束时间。time 包现在支持两种类型的时间窗口：

1. **滑动窗口**：固定大小的窗口，以指定的步长在时间上滑动
2. **滚动窗口**：非重叠的固定大小窗口，在时间上滚动

#### 滑动窗口示例

```go
// 创建滑动窗口
startTime := time.Now().Add(-10 * time.Minute)
endTime := time.Now()
windowSize := 2 * time.Minute
stepSize := 1 * time.Minute

window := sfsTime.NewSlidingWindow(startTime, endTime, windowSize, stepSize)

// 遍历窗口
for window.Next() {
    start := window.Start()
    end := window.End()
    fmt.Printf("窗口: %s 到 %s\n", start, end)
}

// 重置窗口
window.Reset()
```

#### 滚动窗口示例

```go
// 创建滚动窗口
startTime := time.Now().Add(-10 * time.Minute)
endTime := time.Now()
windowSize := 2 * time.Minute

window := sfsTime.NewTumblingWindow(startTime, endTime, windowSize)

// 遍历窗口
for window.Next() {
    start := window.Start()
    end := window.End()
    fmt.Printf("窗口: %s 到 %s\n", start, end)
}
```

#### 窗口聚合

您可以在每个时间窗口内聚合数据：

```go
// 创建测试数据
var records []map[string]any
now := time.Now()
for i := 0; i < 10; i++ {
    records = append(records, map[string]any{
        "timestamp": now.Add(time.Duration(i) * time.Minute),
        "value":     float64(i),
    })
}

// 创建窗口
window := sfsTime.NewSlidingWindow(startTime, endTime, windowSize, stepSize)

// 按窗口聚合
results, err := sfsTime.AggregateByWindow(
    records,
    "timestamp",
    "value",
    window,
    "sum"
)

// 输出结果
for _, result := range results {
    fmt.Printf("窗口: %s 到 %s, 总和: %.2f\n", result.WindowStart, result.WindowEnd, result.Value)
}
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

### 时间序列预测

time 包现在包含时间序列预测功能，支持两种预测方法：

1. **移动平均预测**：使用历史数据的移动平均值来预测未来值
2. **线性回归预测**：使用线性回归基于历史趋势预测未来值

#### 移动平均预测示例

```go
// 创建测试数据点
var points []sfsTime.TimeSeriesPoint
now := time.Now()
for i := 0; i < 10; i++ {
    points = append(points, sfsTime.TimeSeriesPoint{
        Time:  now.Add(time.Duration(i) * time.Minute),
        Value: float64(i),
    })
}

// 创建移动平均预测
maPrediction := sfsTime.NewMovingAveragePrediction(points, 3, 5, time.Minute)

// 输出预测点
fmt.Println("移动平均预测结果：")
for _, point := range maPrediction.PredictedPoints {
    fmt.Printf("预测: %s, 值: %.2f\n", point.Time, point.Value)
}
```

#### 线性回归预测示例

```go
// 创建线性回归预测
lrPrediction := sfsTime.NewLinearRegressionPrediction(points, 5, time.Minute)

// 输出预测点
fmt.Println("线性回归预测结果：")
for _, point := range lrPrediction.PredictedPoints {
    fmt.Printf("预测: %s, 值: %.2f\n", point.Time, point.Value)
}

// 输出回归参数
fmt.Printf("回归参数: 斜率=%.2f, 截距=%.2f, R²=%.2f\n", 
    lrPrediction.Slope, lrPrediction.Intercept, lrPrediction.R2)
```

### 时间序列数据压缩

time 包现在支持时间序列数据压缩，以减少存储空间：

1. **增量编码**：通过存储连续值之间的差异来压缩数据
2. **游程编码 (RLE)**：通过存储重复值及其计数来压缩数据

#### 压缩示例

```go
// 创建测试数据点
var points []sfsTime.TimeSeriesPoint
now := time.Now()
for i := 0; i < 10; i++ {
    points = append(points, sfsTime.TimeSeriesPoint{
        Time:  now.Add(time.Duration(i) * time.Minute),
        Value: float64(i),
    })
}

// 压缩时间序列数据
compressed, err := sfsTime.CompressTimeSeries(points, "delta", time.Minute)
if err != nil {
    panic(err)
}

// 解压缩时间序列数据
decompressed, err := sfsTime.DecompressTimeSeries(compressed, 10)
if err != nil {
    panic(err)
}

// 计算压缩率
originalSize := len(points) * 16 // 假设每个点16字节
compressedSize := len(compressed.CompressedValues)
compressionRatio := sfsTime.GetCompressionRatio(originalSize, compressedSize)
fmt.Printf("压缩率: %.2f\n", compressionRatio)
```

## 与数据库集成

### 时间范围查询优化

time 包现在提供了优化的时间范围查询功能，可以与 sfsDb 无缝集成：

#### 增强的时间范围查询示例

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

// 创建时间范围查询选项
options := sfsTime.NewTimeRangeQueryOptions(
    "timestamp",
    time.Now().Add(-1*time.Hour),
    time.Now(),
    sfsTime.TimeGranularityHour
)

// 执行带优化选项的时间范围查询
iter, err := sfsTime.SearchTimeRange(sensorTable, options)
if err != nil {
    return err
}
defer iter.Release()

// 获取结果
records := iter.GetRecords(true)
defer records.Release()

// 按小时聚合，使用优化的时间范围
aggregationResults, err := sfsTime.TimeRangeQueryWithAggregation(
    sensorTable,
    options,
    "value",
    "sum"
)
if err != nil {
    return err
}

// 输出聚合结果
fmt.Println("传感器值的每小时总和：")
for _, result := range aggregationResults {
    fmt.Printf("时间: %s, 总和: %.2f\n", result.TimeKey, result.Value)
}
```

#### 带粒度的时间范围查询

```go
// 执行带指定粒度的时间范围查询
iter, err := sfsTime.SearchTimeRangeWithGranularity(
    sensorTable,
    "timestamp",
    time.Now().Add(-24*time.Hour),
    time.Now(),
    sfsTime.TimeGranularityHour
)
if err != nil {
    return err
}
defer iter.Release()

// 处理结果...
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