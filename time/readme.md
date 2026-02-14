
## 二、time 包功能分析

### 1. 核心功能

- **时间粒度格式化** (`FormatTimeByGranularity`): 将时间戳格式化为指定粒度的时间字符串，支持秒、分钟、小时、天、月、年等粒度。

- **时间桶分配** (`TimeBucket`): 将时间戳分配到指定粒度的时间桶，保持原始时区信息。

- **时间范围计算** (`TimeRange`): 根据时间粒度计算时间范围，返回起始时间和结束时间。

- **时间粒度聚合** (`AggregateByTimeGranularity`): 按时间粒度聚合数据，支持求和、平均值、计数、最大值、最小值等聚合类型。

### 2. 实现特点

- **时区处理**：正确处理时区信息，确保时间操作的一致性。

- **灵活的时间粒度**：支持多种时间粒度，满足不同场景的需求。

- **集成聚合功能**：与 record 包的聚合功能无缝集成，提供强大的数据分析能力。

- **易用性**：提供简洁明了的 API，便于用户使用。

## 三、应用场景

### 1. 时序数据处理

- **IoT 传感器数据**：按小时、天聚合传感器数据，分析趋势。

- **系统监控**：按分钟、小时聚合系统指标，检测异常。

- **业务分析**：按天、月聚合业务数据，生成报表。

### 2. 时间范围查询优化

- **时间窗口查询**：根据时间粒度快速定位数据范围。

- **数据分区**：基于时间桶进行数据分区，提高查询效率。

- **预计算**：按时间粒度预计算聚合结果，加速查询。

## 四、使用示例

### 1. 基本使用

```go
import (
    "time"
    sfsTime "github.com/liaoran123/sfsDb/time"
)

// 格式化时间
t := time.Now()
hourStr := sfsTime.FormatTimeByGranularity(t, sfsTime.TimeGranularityHour)
fmt.Println("Hour:", hourStr) // 输出: Hour: 2024-12-25 14:00:00

// 时间桶分配
bucket := sfsTime.TimeBucket(t, time.Hour)
fmt.Println("Bucket:", bucket) // 输出: Bucket: 2024-12-25 14:00:00 +0000 UTC

// 时间范围计算
start, end := sfsTime.TimeRange(t, sfsTime.TimeGranularityDay)
fmt.Println("Start:", start, "End:", end) // 输出: Start: 2024-12-25 14:30:45 +0000 UTC End: 2024-12-26 14:30:45 +0000 UTC
```

### 2. 数据聚合

```go
// 假设我们有一组记录
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

## 五、结论

sfsDb 的 `time` 包提供了强大的时序数据处理功能，包括时间粒度格式化、时间桶分配、时间范围计算和时间粒度聚合等。这些功能可以帮助用户更有效地处理和分析时序数据，特别是在 IoT、监控和业务分析等场景中。

测试结果表明，`time` 包的功能正常工作，所有测试都已通过。这意味着用户可以放心使用这些功能来构建时序数据应用。

通过外部开发的方式实现这些功能，sfsDb 保持了核心库的轻量化，同时为用户提供了强大的时序数据处理能力，这符合 sfsDb 作为轻量级嵌入式数据库的定位。