# 第 6 章：时序数据处理

时序数据是工业物联网和边缘计算场景中最常见的数据类型。本章将学习 sfsDb 的 time 包，掌握时序数据的高效处理。

## 6.1 time 包介绍

### 6.1.1 时序数据特点

时序数据是按时间顺序记录的数据序列，具有以下特点：

| 特点 | 说明 | 工业物联网场景 |
|------|------|----------------|
| **时间有序** | 数据按时间戳递增顺序排列 | 传感器每分钟采集一次数据 |
| **写入密集** | 写入频率远高于读取频率 | 设备状态实时上报 |
| **批量读取** | 通常读取一段时间范围内的数据 | 查询过去24小时的温度曲线 |
| **数据压缩** | 相邻数据相关性高，适合压缩 | 温度变化缓慢，可大幅压缩 |
| **保留策略** | 旧数据可按策略聚合或删除 | 保留原始数据7天，聚合数据保留1年 |

**sfsDb 时序数据优势：**
- ✅ 原生支持多种时间粒度
- ✅ 高效的时间窗口计算
- ✅ 内置数据聚合函数
- ✅ 自动数据压缩
- ✅ 智能预测功能

### 6.1.2 time 包架构

sfsDb 的 time 包采用模块化设计，主要包含以下组件：

```
time/
├── granularity.go        # 时间粒度管理
├── window.go            # 时间窗口计算
├── aggregationResult.go # 聚合结果处理
├── compression.go       # 数据压缩
├── time_query.go        # 时序查询
├── prediction.go        # 数据预测
└── timestamp_conversion_test.go # 时间戳转换
```

**核心设计理念：**
1. **粒度分层**：不同时间粒度的数据分开存储
2. **预聚合**：数据写入时自动进行多粒度聚合
3. **延迟计算**：复杂查询按需计算，避免预计算开销
4. **压缩优先**：自动检测数据模式并应用压缩算法

### 6.1.3 核心数据结构

让我们查看 time 包中的核心数据结构：

```go
// TimeGranularity 时间粒度定义
type TimeGranularity int

const (
    GranularitySecond TimeGranularity = iota  // 秒级
    GranularityMinute                          // 分钟级
    GranularityHour                            // 小时级
    GranularityDay                             // 天级
    GranularityWeek                            // 周级
    GranularityMonth                           // 月级
    GranularityYear                            // 年级
)

// TimeWindow 时间窗口
type TimeWindow struct {
    StartTime int64           // 开始时间戳
    EndTime   int64           // 结束时间戳
    Granularity TimeGranularity // 时间粒度
}

// AggregationResult 聚合结果
type AggregationResult struct {
    Timestamp int64       // 时间戳
    Value     float64     // 聚合值
    Count     int         // 数据点数量
    Min       float64     // 最小值
    Max       float64     // 最大值
    Sum       float64     // 总和
    Avg       float64     // 平均值
}
```

## 6.2 时间粒度处理

### 6.2.1 支持的时间粒度

sfsDb 支持7种时间粒度，满足不同场景需求：

| 粒度 | 常量名 | 典型场景 | 数据量估算 |
|------|--------|----------|-----------|
| 秒级 | `GranularitySecond` | 高频交易、实时监控 | 86,400 条/天 |
| 分钟级 | `GranularityMinute` | 设备监控、环境监测 | 1,440 条/天 |
| 小时级 | `GranularityHour` | 能耗统计、流量分析 | 24 条/天 |
| 天级 | `GranularityDay` | 日报表、销售统计 | 1 条/天 |
| 周级 | `GranularityWeek` | 周报表、趋势分析 | ~52 条/年 |
| 月级 | `GranularityMonth` | 月报表、财务统计 | 12 条/年 |
| 年级 | `GranularityYear` | 年度总结、长期趋势 | 1 条/年 |

**使用示例：**

```go
package main

import (
    "fmt"
    "time"

    "github.com/liaoran123/sfsDb/time"
)

func main() {
    // 获取当前时间戳
    now := time.Now().Unix()

    // 秒级粒度
    secondTimestamp := time.RoundToGranularity(now, time.GranularitySecond)
    fmt.Printf("秒级时间戳: %d\n", secondTimestamp)

    // 分钟级粒度
    minuteTimestamp := time.RoundToGranularity(now, time.GranularityMinute)
    fmt.Printf("分钟级时间戳: %d\n", minuteTimestamp)

    // 小时级粒度
    hourTimestamp := time.RoundToGranularity(now, time.GranularityHour)
    fmt.Printf("小时级时间戳: %d\n", hourTimestamp)

    // 天级粒度
    dayTimestamp := time.RoundToGranularity(now, time.GranularityDay)
    fmt.Printf("天级时间戳: %d\n", dayTimestamp)
}
```

### 6.2.2 时间粒度转换

sfsDb 提供了便捷的时间粒度转换功能：

```go
// 在不同粒度之间转换
func convertGranularityExample() {
    now := time.Now().Unix()

    // 从秒级转换到分钟级
    minuteTS := time.ConvertGranularity(now, time.GranularitySecond, time.GranularityMinute)

    // 从分钟级转换到小时级
    hourTS := time.ConvertGranularity(minuteTS, time.GranularityMinute, time.GranularityHour)

    fmt.Printf("秒级: %d → 分钟级: %d → 小时级: %d\n", now, minuteTS, hourTS)
}

// 获取时间范围的所有粒度点
func getGranularityRangeExample() {
    start := time.Now().Add(-24 * time.Hour).Unix()
    end := time.Now().Unix()

    // 获取小时级粒度的所有时间点
    hourlyTimestamps := time.GetGranularityRange(start, end, time.GranularityHour)

    fmt.Printf("过去24小时的小时级时间点数量: %d\n", len(hourlyTimestamps))
    for i, ts := range hourlyTimestamps {
        fmt.Printf("  %d: %s\n", i+1, time.Unix(ts, 0).Format("2006-01-02 15:04:05"))
    }
}
```

### 6.2.3 粒度选择策略

选择合适的时间粒度是时序数据处理的关键：

**策略1：根据业务需求选择**
- 实时监控 → 秒级或分钟级
- 日报表 → 天级
- 趋势分析 → 周级或月级

**策略2：多级粒度并存**
```go
// 工业传感器数据存储策略
type SensorDataStorage struct {
    // 原始数据：保留7天，秒级
    rawData     *engine.Table
    // 分钟级聚合：保留30天
    minuteData  *engine.Table
    // 小时级聚合：保留90天
    hourData    *engine.Table
    // 天级聚合：永久保留
    dayData     *engine.Table
}

func (s *SensorDataStorage) WriteSensorData(timestamp int64, value float64) {
    // 1. 写入原始数据
    s.writeRawData(timestamp, value)

    // 2. 异步更新聚合数据
    go s.updateAggregations(timestamp, value)
}
```

**策略3：动态调整粒度**
```go
// 根据数据量自动选择粒度
func chooseOptimalGranularity(dataPoints int) time.TimeGranularity {
    switch {
    case dataPoints < 1000:
        return time.GranularitySecond
    case dataPoints < 10000:
        return time.GranularityMinute
    case dataPoints < 100000:
        return time.GranularityHour
    default:
        return time.GranularityDay
    }
}
```

## 6.3 时间窗口计算

### 6.3.1 滑动窗口

滑动窗口（Sliding Window）是指窗口大小固定，按固定步长向前滑动的窗口。

**适用场景：**
- 实时监控告警
- 移动平均线计算
- 实时流量统计

**示例：5分钟滑动窗口计算平均温度**

```go
package main

import (
    "fmt"
    "time"

    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/time"
)

func slidingWindowExample() {
    // 创建温度表
    tempTable, _ := engine.TableNew("temperature")
    tempTable.SetFields(map[string]any{
        "timestamp": int64(0),
        "value":     0.0,
        "sensor_id": "",
    })

    // 创建滑动窗口计算器
    window := time.NewSlidingWindow(
        5*time.Minute,      // 窗口大小：5分钟
        1*time.Minute,      // 滑动步长：1分钟
    )

    // 模拟温度数据
    now := time.Now()
    for i := 0; i < 60; i++ {
        timestamp := now.Add(-time.Duration(i) * time.Minute).Unix()
        value := 25.0 + float64(i%10) // 模拟温度波动

        // 添加数据到窗口
        window.Add(timestamp, value)

        // 写入数据库
        tempTable.Insert(&map[string]interface{}{
            "timestamp": timestamp,
            "value":     value,
            "sensor_id": "sensor_001",
        })
    }

    // 计算滑动窗口统计
    results := window.Calculate()
    fmt.Printf("滑动窗口计算结果:\n")
    for _, result := range results {
        fmt.Printf("  时间: %s, 平均温度: %.2f°C, 数据点: %d\n",
            time.Unix(result.Timestamp, 0).Format("15:04"),
            result.Avg,
            result.Count)
    }
}
```

### 6.3.2 滚动窗口

滚动窗口（Tumbling Window）是指窗口不重叠，按固定大小分割的窗口。

**适用场景：**
- 日报表、周报表
- 按时间段统计
- 批量数据处理

**示例：按小时计算能耗统计**

```go
func tumblingWindowExample() {
    // 创建能耗表
    energyTable, _ := engine.TableNew("energy_consumption")
    energyTable.SetFields(map[string]any{
        "timestamp": int64(0),
        "kwh":       0.0,
        "device_id": "",
    })

    // 查询过去24小时的数据
    startTime := time.Now().Add(-24 * time.Hour).Unix()
    endTime := time.Now().Unix()

    // 创建滚动窗口（每小时一个窗口）
    windows := time.CreateTumblingWindows(
        startTime,
        endTime,
        time.GranularityHour,
    )

    fmt.Printf("过去24小时的滚动窗口统计:\n")
    for _, window := range windows {
        // 查询该窗口的数据
        iter, _ := energyTable.Search(&map[string]any{
            "timestamp": map[string]any{
                "$gte": window.StartTime,
                "$lt":  window.EndTime,
            },
        })
        defer engine.GlobalTableIterPool.Put(iter)

        records := iter.GetRecords(true)
        defer record.PutRecords(records)

        // 计算统计
        totalKwh := 0.0
        for _, r := range records {
            totalKwh += r["kwh"].(float64)
        }

        fmt.Printf("  %s - %s: %.2f kWh\n",
            time.Unix(window.StartTime, 0).Format("15:04"),
            time.Unix(window.EndTime, 0).Format("15:04"),
            totalKwh)
    }
}
```

### 6.3.3 会话窗口

会话窗口（Session Window）是根据数据的活跃度来划分的窗口，当数据间隔超过阈值时，窗口结束。

**适用场景：**
- 用户行为分析
- 设备活跃时段分析
- 会话追踪

**示例：分析设备活跃会话**

```go
func sessionWindowExample() {
    // 设备心跳数据表
    heartbeatTable, _ := engine.TableNew("device_heartbeat")
    heartbeatTable.SetFields(map[string]any{
        "timestamp": int64(0),
        "device_id": "",
        "status":    "",
    })

    // 会话超时时间：5分钟
    sessionTimeout := 5 * time.Minute

    // 查询某设备的心跳数据
    iter, _ := heartbeatTable.Search(&map[string]any{
        "device_id": "device_001",
    })
    defer engine.GlobalTableIterPool.Put(iter)

    records := iter.GetRecords(true)
    defer record.PutRecords(records)

    // 识别会话
    sessions := time.DetectSessions(records, "timestamp", sessionTimeout)

    fmt.Printf("设备活跃会话分析:\n")
    for i, session := range sessions {
        duration := time.Duration(session.EndTime - session.StartTime) * time.Second
        fmt.Printf("  会话 %d: %s - %s, 持续时间: %v, 心跳次数: %d\n",
            i+1,
            time.Unix(session.StartTime, 0).Format("2006-01-02 15:04:05"),
            time.Unix(session.EndTime, 0).Format("2006-01-02 15:04:05"),
            duration,
            session.Count)
    }
}
```

## 6.4 数据聚合功能

### 6.4.1 常见聚合函数

sfsDb 内置了丰富的聚合函数：

| 函数名 | 说明 | 示例 |
|--------|------|------|
| `Sum` | 求和 | 计算总能耗 |
| `Avg` | 平均值 | 计算平均温度 |
| `Min` | 最小值 | 最低温度 |
| `Max` | 最大值 | 最高温度 |
| `Count` | 计数 | 数据点数量 |
| `StdDev` | 标准差 | 数据波动程度 |
| `Variance` | 方差 | 数据离散程度 |
| `Percentile` | 百分位数 | P50、P95、P99 |

**使用示例：**

```go
package main

import (
    "fmt"

    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/record"
    "github.com/liaoran123/sfsDb/time"
)

func aggregationExample() {
    // 创建传感器数据表
    sensorTable, _ := engine.TableNew("sensor_data")
    sensorTable.SetFields(map[string]any{
        "timestamp": int64(0),
        "temperature": 0.0,
        "humidity":    0.0,
        "pressure":    0.0,
    })

    // 查询所有数据
    iter, _ := sensorTable.Search(&map[string]any{})
    defer engine.GlobalTableIterPool.Put(iter)

    records := iter.GetRecords(true)
    defer record.PutRecords(records)

    // 提取温度数据
    temperatures := make([]float64, len(records))
    for i, r := range records {
        temperatures[i] = r["temperature"].(float64)
    }

    // 计算聚合
    agg := time.Aggregate(temperatures)

    fmt.Printf("温度数据聚合结果:\n")
    fmt.Printf("  数据点数量: %d\n", agg.Count)
    fmt.Printf("  总和: %.2f\n", agg.Sum)
    fmt.Printf("  平均值: %.2f°C\n", agg.Avg)
    fmt.Printf("  最小值: %.2f°C\n", agg.Min)
    fmt.Printf("  最大值: %.2f°C\n", agg.Max)
    fmt.Printf("  标准差: %.2f\n", agg.StdDev)
    fmt.Printf("  方差: %.2f\n", agg.Variance)

    // 计算百分位数
    p50 := time.Percentile(temperatures, 50)
    p95 := time.Percentile(temperatures, 95)
    p99 := time.Percentile(temperatures, 99)

    fmt.Printf("\n百分位数:\n")
    fmt.Printf("  P50: %.2f°C\n", p50)
    fmt.Printf("  P95: %.2f°C\n", p95)
    fmt.Printf("  P99: %.2f°C\n", p99)
}
```

### 6.4.2 自定义聚合

你可以实现自定义的聚合函数：

```go
// 自定义聚合：计算加权平均
type WeightedAverageAggregator struct {
    weights []float64
}

func NewWeightedAverageAggregator(weights []float64) *WeightedAverageAggregator {
    return &WeightedAverageAggregator{weights: weights}
}

func (wa *WeightedAverageAggregator) Aggregate(values []float64) float64 {
    if len(values) == 0 {
        return 0
    }

    // 使用最近N个数据的权重
    n := len(wa.weights)
    if len(values) < n {
        n = len(values)
    }

    sum := 0.0
    weightSum := 0.0
    for i := 0; i < n; i++ {
        idx := len(values) - n + i
        sum += values[idx] * wa.weights[i]
        weightSum += wa.weights[i]
    }

    return sum / weightSum
}

// 使用自定义聚合
func customAggregationExample() {
    // 创建指数衰减权重（最近的数据权重更大）
    weights := []float64{0.1, 0.2, 0.3, 0.4}
    aggregator := NewWeightedAverageAggregator(weights)

    // 模拟温度数据
    temperatures := []float64{24.5, 24.8, 25.1, 25.3, 25.5, 25.2, 25.0}

    weightedAvg := aggregator.Aggregate(temperatures)
    fmt.Printf("加权平均温度: %.2f°C\n", weightedAvg)
}
```

### 6.4.3 聚合性能优化

**优化策略1：预聚合**

```go
// 数据写入时自动聚合
type PreAggregator struct {
    table     *engine.Table
    granularity time.TimeGranularity
    cache     map[int64]*AggregationResult
}

func (pa *PreAggregator) Write(timestamp int64, value float64) {
    // 写入原始数据
    pa.table.Insert(&map[string]interface{}{
        "timestamp": timestamp,
        "value":     value,
    })

    // 更新缓存中的聚合
    roundedTs := time.RoundToGranularity(timestamp, pa.granularity)
    if agg, exists := pa.cache[roundedTs]; exists {
        agg.Count++
        agg.Sum += value
        agg.Avg = agg.Sum / float64(agg.Count)
        if value < agg.Min {
            agg.Min = value
        }
        if value > agg.Max {
            agg.Max = value
        }
    } else {
        pa.cache[roundedTs] = &AggregationResult{
            Timestamp: roundedTs,
            Count:     1,
            Sum:       value,
            Avg:       value,
            Min:       value,
            Max:       value,
        }
    }
}
```

**优化策略2：批量聚合**

```go
// 批量聚合查询
func batchAggregationExample() {
    sensorTable, _ := engine.TableNew("sensor_data")

    // 按小时批量查询并聚合
    startTime := time.Now().Add(-7 * 24 * time.Hour).Unix()
    endTime := time.Now().Unix()

    // 使用并行聚合
    results := time.ParallelAggregate(
        sensorTable,
        "timestamp",
        "value",
        startTime,
        endTime,
        time.GranularityHour,
        4, // 4个 goroutine 并行处理
    )

    fmt.Printf("7天按小时聚合结果数量: %d\n", len(results))
}
```

## 6.5 时间戳转换

### 6.5.1 时间格式转换

sfsDb 提供了便捷的时间格式转换功能：

```go
package main

import (
    "fmt"
    "time"

    "github.com/liaoran123/sfsDb/time"
)

func timestampConversionExample() {
    // Unix 时间戳
    timestamp := time.Now().Unix()
    fmt.Printf("Unix 时间戳: %d\n", timestamp)

    // 转换为时间字符串
    timeStr := time.FormatTimestamp(timestamp, "2006-01-02 15:04:05")
    fmt.Printf("格式化时间: %s\n", timeStr)

    // 从字符串解析
    parsedTs, err := time.ParseTimestamp("2024-01-01 12:00:00", "2006-01-02 15:04:05")
    if err != nil {
        fmt.Printf("解析失败: %v\n", err)
    } else {
        fmt.Printf("解析得到的时间戳: %d\n", parsedTs)
    }

    // RFC3339 格式
    rfc3339Str := time.ToRFC3339(timestamp)
    fmt.Printf("RFC3339 格式: %s\n", rfc3339Str)

    // 相对时间
    relative := time.RelativeTime(timestamp)
    fmt.Printf("相对时间: %s\n", relative) // 例如："2小时前"
}
```

### 6.5.2 时区处理

```go
func timezoneExample() {
    timestamp := time.Now().Unix()

    // 转换到不同时区
    locations := []string{
        "Asia/Shanghai",   // 中国标准时间
        "America/New_York", // 美国东部时间
        "Europe/London",   // 伦敦时间
        "UTC",             // UTC时间
    }

    for _, loc := range locations {
        localTime, err := time.ToTimezone(timestamp, loc)
        if err != nil {
            fmt.Printf("时区 %s 转换失败: %v\n", loc, err)
            continue
        }
        fmt.Printf("%s: %s\n", loc, localTime.Format("2006-01-02 15:04:05 MST"))
    }
}
```

### 6.5.3 精度控制

```go
func precisionExample() {
    now := time.Now()

    // 不同精度的时间戳
    millis := now.UnixMilli()      // 毫秒级
    micros := now.UnixMicro()      // 微秒级
    nanos := now.UnixNano()        // 纳秒级

    fmt.Printf("毫秒: %d\n", millis)
    fmt.Printf("微秒: %d\n", micros)
    fmt.Printf("纳秒: %d\n", nanos)

    // 截断到指定精度
    secondTs := time.TruncateToPrecision(nanos, time.PrecisionSecond)
    minuteTs := time.TruncateToPrecision(nanos, time.PrecisionMinute)

    fmt.Printf("截断到秒: %d\n", secondTs)
    fmt.Printf("截断到分钟: %d\n", minuteTs)
}
```

## 6.6 实战示例

### 6.6.1 工业传感器数据处理

让我们创建一个完整的工业传感器数据处理系统：

```go
package main

import (
    "fmt"
    "log"
    "math/rand"
    "time"

    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/storage"
    "github.com/liaoran123/sfsDb/time"
)

// SensorDataManager 传感器数据管理器
type SensorDataManager struct {
    rawTable    *engine.Table // 原始数据表
    minuteTable *engine.Table // 分钟级聚合表
    hourTable   *engine.Table // 小时级聚合表
    dayTable    *engine.Table // 天级聚合表
}

func NewSensorDataManager() (*SensorDataManager, error) {
    // 初始化数据库
    dbManager := storage.GetDBManager()
    _, err := dbManager.OpenDB("./sensor_db")
    if err != nil {
        return nil, err
    }

    // 创建原始数据表
    rawTable, err := engine.TableNew("sensor_raw")
    if err != nil {
        return nil, err
    }
    rawTable.SetFields(map[string]any{
        "timestamp": int64(0),
        "sensor_id": "",
        "temperature": 0.0,
        "humidity":    0.0,
        "pressure":    0.0,
    })
    rawPk, _ := engine.DefaultPrimaryKeyNew("timestamp,sensor_id")
    rawPk.AddFields("timestamp")
    rawPk.AddFields("sensor_id")
    rawTable.CreateIndex(rawPk)

    // 创建分钟级聚合表
    minuteTable, _ := engine.TableNew("sensor_minute")
    minuteTable.SetFields(map[string]any{
        "timestamp": int64(0),
        "sensor_id": "",
        "temp_avg":   0.0,
        "temp_min":   0.0,
        "temp_max":   0.0,
        "humidity_avg": 0.0,
        "count":      0,
    })

    // 创建小时级聚合表
    hourTable, _ := engine.TableNew("sensor_hour")
    hourTable.SetFields(map[string]any{
        "timestamp": int64(0),
        "sensor_id": "",
        "temp_avg":   0.0,
        "temp_min":   0.0,
        "temp_max":   0.0,
        "humidity_avg": 0.0,
        "count":      0,
    })

    // 创建天级聚合表
    dayTable, _ := engine.TableNew("sensor_day")
    dayTable.SetFields(map[string]any{
        "timestamp": int64(0),
        "sensor_id": "",
        "temp_avg":   0.0,
        "temp_min":   0.0,
        "temp_max":   0.0,
        "humidity_avg": 0.0,
        "count":      0,
    })

    return &SensorDataManager{
        rawTable:    rawTable,
        minuteTable: minuteTable,
        hourTable:   hourTable,
        dayTable:    dayTable,
    }, nil
}

func (sdm *SensorDataManager) WriteSensorData(sensorID string, timestamp int64, temp, humidity, pressure float64) error {
    // 1. 写入原始数据
    _, err := sdm.rawTable.Insert(&map[string]interface{}{
        "timestamp":   timestamp,
        "sensor_id":   sensorID,
        "temperature": temp,
        "humidity":    humidity,
        "pressure":    pressure,
    })
    if err != nil {
        return err
    }

    // 2. 异步更新聚合数据
    go sdm.updateAggregations(sensorID, timestamp, temp, humidity)

    return nil
}

func (sdm *SensorDataManager) updateAggregations(sensorID string, timestamp int64, temp, humidity float64) {
    // 更新分钟级聚合
    minuteTs := time.RoundToGranularity(timestamp, time.GranularityMinute)
    sdm.updateAggregationTable(sdm.minuteTable, sensorID, minuteTs, temp, humidity)

    // 更新小时级聚合
    hourTs := time.RoundToGranularity(timestamp, time.GranularityHour)
    sdm.updateAggregationTable(sdm.hourTable, sensorID, hourTs, temp, humidity)

    // 更新天级聚合
    dayTs := time.RoundToGranularity(timestamp, time.GranularityDay)
    sdm.updateAggregationTable(sdm.dayTable, sensorID, dayTs, temp, humidity)
}

func (sdm *SensorDataManager) updateAggregationTable(table *engine.Table, sensorID string, timestamp int64, temp, humidity float64) {
    // 查询是否已有该时间段的聚合数据
    iter, _ := table.Search(&map[string]any{
        "timestamp": timestamp,
        "sensor_id": sensorID,
    })
    defer engine.GlobalTableIterPool.Put(iter)

    records := iter.GetRecords(true)
    defer record.PutRecords(records)

    if len(records) > 0 {
        // 更新现有聚合
        existing := records[0]
        count := existing["count"].(int) + 1
        tempAvg := (existing["temp_avg"].(float64)*float64(existing["count"].(int)) + temp) / float64(count)
        tempMin := min(existing["temp_min"].(float64), temp)
        tempMax := max(existing["temp_max"].(float64), temp)
        humidityAvg := (existing["humidity_avg"].(float64)*float64(existing["count"].(int)) + humidity) / float64(count)

        table.Update(&map[string]interface{}{
            "timestamp":    timestamp,
            "sensor_id":    sensorID,
            "temp_avg":     tempAvg,
            "temp_min":     tempMin,
            "temp_max":     tempMax,
            "humidity_avg": humidityAvg,
            "count":        count,
        })
    } else {
        // 创建新聚合
        table.Insert(&map[string]interface{}{
            "timestamp":    timestamp,
            "sensor_id":    sensorID,
            "temp_avg":     temp,
            "temp_min":     temp,
            "temp_max":     temp,
            "humidity_avg": humidity,
            "count":        1,
        })
    }
}

func main() {
    sdm, err := NewSensorDataManager()
    if err != nil {
        log.Fatalf("初始化失败: %v", err)
    }
    defer storage.GetDBManager().CloseDB()

    fmt.Println("=== 工业传感器数据处理系统 ===\n")

    // 模拟生成传感器数据
    now := time.Now()
    sensorIDs := []string{"sensor_001", "sensor_002", "sensor_003"}

    fmt.Println("正在生成传感器数据...")
    for i := 0; i < 1000; i++ {
        timestamp := now.Add(-time.Duration(1000-i) * time.Second).Unix()
        sensorID := sensorIDs[i%3]

        // 模拟传感器数据
        temp := 25.0 + rand.Float64()*5.0
        humidity := 50.0 + rand.Float64()*20.0
        pressure := 1013.0 + rand.Float64()*10.0

        err := sdm.WriteSensorData(sensorID, timestamp, temp, humidity, pressure)
        if err != nil {
            log.Printf("写入数据失败: %v", err)
        }

        if (i+1)%100 == 0 {
            fmt.Printf("已写入 %d 条数据\n", i+1)
        }
    }

    fmt.Println("\n数据生成完成！")
}

func min(a, b float64) float64 {
    if a < b {
        return a
    }
    return b
}

func max(a, b float64) float64 {
    if a > b {
        return a
    }
    return b
}
```

### 6.6.2 设备监控仪表盘

```go
// Dashboard 监控仪表盘
type Dashboard struct {
    dataManager *SensorDataManager
}

func NewDashboard(dataManager *SensorDataManager) *Dashboard {
    return &Dashboard{dataManager: dataManager}
}

func (d *Dashboard) ShowRealtimeStatus(sensorID string) {
    // 查询最新数据
    iter, _ := d.dataManager.rawTable.Search(&map[string]any{
        "sensor_id": sensorID,
    })
    defer engine.GlobalTableIterPool.Put(iter)

    records := iter.GetRecords(true)
    defer record.PutRecords(records)

    if len(records) > 0 {
        latest := records[len(records)-1]
        fmt.Printf("\n=== %s 实时状态 ===\n", sensorID)
        fmt.Printf("时间: %s\n", time.Unix(latest["timestamp"].(int64), 0).Format("2006-01-02 15:04:05"))
        fmt.Printf("温度: %.2f°C\n", latest["temperature"])
        fmt.Printf("湿度: %.2f%%\n", latest["humidity"])
        fmt.Printf("气压: %.2f hPa\n", latest["pressure"])
    }
}

func (d *Dashboard) ShowHourlyReport(sensorID string) {
    // 查询过去24小时的小时级数据
    startTime := time.Now().Add(-24 * time.Hour).Unix()
    endTime := time.Now().Unix()

    iter, _ := d.dataManager.hourTable.Search(&map[string]any{
        "sensor_id": sensorID,
        "timestamp": map[string]any{
            "$gte": startTime,
            "$lte": endTime,
        },
    })
    defer engine.GlobalTableIterPool.Put(iter)

    records := iter.GetRecords(true)
    defer record.PutRecords(records)

    fmt.Printf("\n=== %s 过去24小时报告 ===\n", sensorID)
    for _, r := range records {
        fmt.Printf("%s: 平均温度 %.2f°C (%.2f°C ~ %.2f°C), 平均湿度 %.2f%%, 数据点 %d\n",
            time.Unix(r["timestamp"].(int64), 0).Format("15:04"),
            r["temp_avg"],
            r["temp_min"],
            r["temp_max"],
            r["humidity_avg"],
            r["count"])
    }
}
```

### 6.6.3 异常检测系统

```go
// AnomalyDetector 异常检测器
type AnomalyDetector struct {
    thresholds map[string]Threshold
}

type Threshold struct {
    MinTemp float64
    MaxTemp float64
    MinHumidity float64
    MaxHumidity float64
}

func NewAnomalyDetector() *AnomalyDetector {
    return &AnomalyDetector{
        thresholds: make(map[string]Threshold),
    }
}

func (ad *AnomalyDetector) SetThreshold(sensorID string, threshold Threshold) {
    ad.thresholds[sensorID] = threshold
}

func (ad *AnomalyDetector) Detect(sensorID string, temp, humidity float64) []string {
    threshold, exists := ad.thresholds[sensorID]
    if !exists {
        return nil
    }

    var anomalies []string

    if temp < threshold.MinTemp {
        anomalies = append(anomalies, fmt.Sprintf("温度过低: %.2f°C < %.2f°C", temp, threshold.MinTemp))
    }
    if temp > threshold.MaxTemp {
        anomalies = append(anomalies, fmt.Sprintf("温度过高: %.2f°C > %.2f°C", temp, threshold.MaxTemp))
    }
    if humidity < threshold.MinHumidity {
        anomalies = append(anomalies, fmt.Sprintf("湿度过低: %.2f%% < %.2f%%", humidity, threshold.MinHumidity))
    }
    if humidity > threshold.MaxHumidity {
        anomalies = append(anomalies, fmt.Sprintf("湿度过高: %.2f%% > %.2f%%", humidity, threshold.MaxHumidity))
    }

    return anomalies
}

// 使用异常检测
func anomalyDetectionExample() {
    detector := NewAnomalyDetector()
    detector.SetThreshold("sensor_001", Threshold{
        MinTemp:     20.0,
        MaxTemp:     30.0,
        MinHumidity: 40.0,
        MaxHumidity: 70.0,
    })

    // 检测异常
    anomalies := detector.Detect("sensor_001", 32.5, 75.0)
    if len(anomalies) > 0 {
        fmt.Printf("检测到 %d 个异常:\n", len(anomalies))
        for _, a := range anomalies {
            fmt.Printf("  - %s\n", a)
        }
    }
}
```

## 6.7 本章小结

本章深入学习了 sfsDb 的时序数据处理能力，主要内容包括：

✅ **time 包介绍**：时序数据特点、架构设计、核心数据结构
✅ **时间粒度处理**：7种粒度支持、粒度转换、选择策略
✅ **时间窗口计算**：滑动窗口、滚动窗口、会话窗口
✅ **数据聚合功能**：内置聚合函数、自定义聚合、性能优化
✅ **时间戳转换**：格式转换、时区处理、精度控制
✅ **实战示例**：传感器数据处理、监控仪表盘、异常检测

**关键要点：**
1. 选择合适的时间粒度是时序数据处理的关键
2. 多级粒度并存可以平衡查询性能和存储成本
3. 预聚合和批量处理可以大幅提升性能
4. 滑动窗口适合实时监控，滚动窗口适合报表统计
5. 异常检测是工业物联网的重要功能

下一章我们将学习 sfsDb 的加密存储功能，这是数据安全的核心保障。
