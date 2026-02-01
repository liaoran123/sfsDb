# Monitoring Package

监控包提供了sfsDb的监控系统，包括指标收集、告警管理和外部系统集成。

## 功能特性

- **指标收集**：收集事务、查询、存储和系统指标
- **告警管理**：基于规则的告警系统，支持多级告警
- **Prometheus集成**：导出指标到Prometheus
- **自定义扩展**：支持自定义指标和告警规则

## 核心模块

### 1. Metrics

指标收集模块，负责收集和管理各种指标数据。

#### 主要接口

```go
// Metrics 指标收集器接口
type Metrics interface {
	// 事务指标
	RecordTransaction(duration time.Duration, success bool)
	GetTransactionCount() int64
	GetTransactionSuccessRate() float64
	GetAverageTransactionDuration() time.Duration

	// 查询指标
	RecordQuery(duration time.Duration, queryType string)
	GetQueryCount() int64
	GetAverageQueryDuration() time.Duration

	// 存储指标
	RecordStorageOperation(opType string, duration time.Duration)
	GetStorageOperationCount() int64

	// 系统指标
	RecordSystemState(concurrency int32, memoryUsage uint64, requests int)
	GetCurrentSystemState() (int32, uint64, int)

	// 重置指标
	Reset()
}
```

### 2. Exporters

指标导出模块，负责将指标导出到外部系统。

#### Prometheus导出器

```go
// PrometheusExporter Prometheus指标导出器
type PrometheusExporter struct {
	// ...
}

// NewPrometheusExporter 创建Prometheus指标导出器
func NewPrometheusExporter(metrics metrics.Metrics) *PrometheusExporter

// StartHTTP 启动HTTP服务器，暴露Prometheus指标
func (e *PrometheusExporter) StartHTTP(addr string) error
```

### 3. Alerts

告警管理模块，负责管理告警规则和处理告警事件。

#### 主要接口

```go
// AlertManager 告警管理器
type AlertManager struct {
	// ...
}

// NewAlertManager 创建告警管理器
func NewAlertManager() *AlertManager

// AddRule 添加告警规则
func (am *AlertManager) AddRule(rule AlertRule)

// CheckAlerts 检查告警规则
func (am *AlertManager) CheckAlerts()
```

## 使用示例

### 1. 初始化监控系统

```go
import (
	"github.com/liaoran123/sfsDb/monitoring"
)

// 创建监控系统
monitor := monitoring.NewMonitor()

// 启动监控系统
monitor.Start()
```

### 2. 记录指标

```go
// 记录事务指标
start := time.Now()
success := true
// 执行事务...
duration := time.Since(start)
monitor.Metrics().RecordTransaction(duration, success)

// 记录查询指标
start = time.Now()
// 执行查询...
duration = time.Since(start)
monitor.Metrics().RecordQuery(duration, "SELECT")
```

### 3. 添加告警规则

```go
// 添加告警规则
alertManager := monitor.AlertManager()
alertManager.AddRule(alerts.AlertRule{
	Name:        "high_transaction_failure",
	Description: "事务失败率过高",
	Level:       alerts.AlertLevelWarning,
	Condition: func() bool {
		successRate := monitor.Metrics().GetTransactionSuccessRate()
		return successRate < 0.9
	},
	Labels: map[string]string{
		"service": "sfsdb",
		"severity": "warning",
	},
})
```

### 4. 配置Prometheus

在Prometheus配置文件中添加以下内容：

```yaml
scrape_configs:
  - job_name: 'sfsdb'
    static_configs:
      - targets: ['localhost:9090']
```

## 配置选项

| 配置项 | 类型 | 默认值 | 描述 |
|--------|------|--------|------|
| metrics.enabled | bool | true | 是否启用指标收集 |
| metrics.interval | time.Duration | 10s | 指标收集间隔 |
| alerts.enabled | bool | true | 是否启用告警 |
| alerts.checkInterval | time.Duration | 30s | 告警检查间隔 |
| prometheus.enabled | bool | true | 是否启用Prometheus导出 |
| prometheus.port | int | 9090 | Prometheus导出端口 |

## 最佳实践

1. **合理设置告警阈值**：根据实际系统负载设置合理的告警阈值
2. **定期清理指标数据**：避免指标数据过大导致内存占用过高
3. **集成监控系统**：将sfsDb监控与现有监控系统集成
4. **设置合理的采样率**：对于高频操作，考虑降低采样率以减少开销

## 故障排查

### 常见问题

1. **指标导出失败**
   - 检查端口是否被占用
   - 检查Prometheus配置是否正确

2. **告警不触发**
   - 检查告警规则条件是否正确
   - 检查告警检查间隔是否合理

3. **内存占用过高**
   - 检查指标收集频率是否过高
   - 考虑启用指标数据压缩

## 版本历史

- **v1.0.0**：初始版本，支持基本指标收集和Prometheus导出
- **v1.1.0**：添加告警系统和自定义指标支持
- **v1.2.0**：优化性能，减少内存占用
