# 第 8 章：工业物联网实战

本章将通过一个完整的工业物联网项目，综合运用前面各章所学的知识，构建一个真实可用的边缘计算应用。我们将基于 **sfsEdgeStore** 的设计理念，实现一个边缘数据存储适配器。

## 8.1 项目需求分析

### 8.1.1 场景描述

**工业物联网边缘网关场景：**

在现代工业生产环境中，大量的传感器、PLC、工业设备产生海量的实时数据。这些数据需要：
- **本地存储**：在网络不稳定或断网时，数据不能丢失
- **边缘计算**：在本地进行初步的数据处理和分析
- **断网续传**：网络恢复后，自动同步数据到云端
- **低延迟**：对实时监控数据需要毫秒级响应
- **高可靠**：工业环境下电源中断、网络波动是常态

**sfsEdgeStore 的角色：**
- 作为 **EdgeX Foundry** 等边缘计算框架与 sfsDb 之间的桥梁
- 提供高效的本地数据读写和缓存能力
- 实现数据持久化、批量处理、断网续传等功能

### 8.1.2 功能需求

| 功能模块 | 需求描述 | 优先级 |
|---------|---------|--------|
| 数据采集 | 从多种工业协议采集数据（Modbus、OPC UA、MQTT） | P0 |
| 数据存储 | 使用 sfsDb 本地持久化存储时序数据 | P0 |
| 数据查询 | 支持实时查询、历史查询、范围查询 | P0 |
| 数据缓存 | LRU 缓存热点数据，提升查询性能 | P1 |
| 数据同步 | 断网续传，网络恢复后同步到云端 | P1 |
| 数据聚合 | 本地数据聚合计算（平均值、最大值、最小值） | P1 |
| 告警检测 | 本地异常检测和告警触发 | P2 |
| 数据加密 | 敏感数据加密存储 | P2 |

### 8.1.3 非功能需求

| 指标 | 要求 |
|------|------|
| **启动时间** | < 5 秒 |
| **内存占用** | < 100 MB |
| **数据写入** | > 10,000 TPS |
| **查询延迟** | < 10 ms（缓存命中） |
| **数据可靠性** | 99.999%（断电不丢失） |
| **并发连接** | 支持 100+ 设备同时连接 |

### 8.1.4 技术选型

| 组件 | 选型 | 说明 |
|------|------|------|
| **数据库** | sfsDb | 嵌入式数据库，纯 Go，无 CGO |
| **边缘框架** | EdgeX Foundry | 工业物联网边缘计算框架 |
| **通信协议** | MQTT | 设备数据采集协议 |
| **Web 框架** | Gin | 轻量级 Go Web 框架 |
| **序列化** | JSON/Protobuf | 数据序列化格式 |
| **配置管理** | Viper | 配置文件管理 |

## 8.2 架构设计

### 8.2.1 整体架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        工业物联网边缘网关                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐   │
│  │  工业设备    │    │  PLC 控制器  │    │   传感器     │   │
│  │ (Modbus)     │    │ (OPC UA)    │    │ (MQTT)      │   │
│  └──────┬───────┘    └──────┬───────┘    └──────┬───────┘   │
│         │                   │                   │            │
│         └───────────────────┴───────────────────┘            │
│                             │                                │
│                    ┌────────▼─────────┐                      │
│                    │  数据采集层       │                      │
│                    │  (Protocol Adapters)│                     │
│                    └────────┬─────────┘                      │
│                             │                                │
│                    ┌────────▼─────────┐                      │
│                    │  sfsEdgeStore      │                      │
│                    │  (边缘数据存储)    │                      │
│  ┌─────────────────┴─────────────────┴─────────────────┐    │
│  │                                                     │    │
│  │  ┌─────────────┐  ┌─────────────┐  ┌───────────┐  │    │
│  │  │  数据缓存   │  │  数据处理   │  │  数据同步  │  │    │
│  │  │  (LRU)     │  │  (聚合/告警) │  │  (断网续传)│  │    │
│  │  └──────┬──────┘  └──────┬──────┘  └─────┬─────┘  │    │
│  └─────────┼──────────────────┼────────────────┼──────────┘    │
│            │                  │                │               │
│            └──────────────────┴────────────────┘               │
│                               │                                │
│                    ┌──────────▼──────────┐                    │
│                    │      sfsDb          │                    │
│                    │  (嵌入式数据库)      │                    │
│                    │  - 时序数据存储      │                    │
│                    │  - 无锁事务          │                    │
│                    │  - 加密存储          │                    │
│                    └──────────┬──────────┘                    │
│                               │                                │
└───────────────────────────────┼────────────────────────────────┘
                                │
                        ┌───────▼────────┐
                        │     云端平台    │
                        │  (数据同步)     │
                        └─────────────────┘
```

### 8.2.2 模块划分

**核心模块：**

1. **采集模块 (Collector)**
   - 协议适配：Modbus、OPC UA、MQTT
   - 数据解码和标准化
   - 设备状态管理

2. **存储模块 (Storage)**
   - sfsDb 封装
   - 时序数据写入
   - 数据查询接口

3. **缓存模块 (Cache)**
   - LRU 缓存实现
   - 热点数据管理
   - 缓存过期策略

4. **处理模块 (Processor)**
   - 数据聚合计算
   - 异常检测
   - 告警触发

5. **同步模块 (Sync)**
   - 断网检测
   - 数据队列管理
   - 云端同步

6. **API 模块 (API)**
   - RESTful API
   - WebSocket 实时推送
   - 认证授权

### 8.2.3 数据流设计

**数据写入流程：**

```
工业设备 → 数据采集 → 数据标准化 → 缓存检查
                                    ↓
                           缓存命中？←───┐
                            ↓ 否        │ 是
                         sfsDb 存储    │
                            ↓           │
                         更新缓存 ←─────┘
                            ↓
                         （可选）数据聚合
                            ↓
                         （可选）异常检测
                            ↓
                         （可选）加入同步队列
```

**数据查询流程：**

```
查询请求 → 检查缓存
              ↓ 命中？
         ┌────┴────┐
         是        否
         ↓         ↓
     返回缓存    sfsDb 查询
         ↓         ↓
      （可选）更新缓存
         ↓
      返回结果
```

### 8.2.4 数据库设计

**数据表设计：**

1. **设备信息表 (devices)**
   ```go
   type Device struct {
       ID          int    // 设备ID
       Name        string // 设备名称
       Type        string // 设备类型
       Protocol    string // 通信协议
       Status      string // 设备状态
       CreatedAt   int64  // 创建时间
       UpdatedAt   int64  // 更新时间
   }
   ```

2. **传感器数据表 (sensor_data)**
   ```go
   type SensorData struct {
       ID          int       // 记录ID
       DeviceID    int       // 设备ID
       SensorID    string    // 传感器ID
       Value       float64   // 数值
       Unit        string    // 单位
       Timestamp   int64     // 时间戳
       Quality     string    // 数据质量
   }
   ```

3. **告警信息表 (alerts)**
   ```go
   type Alert struct {
       ID          int       // 告警ID
       DeviceID    int       // 设备ID
       Type        string    // 告警类型
       Level       string    // 告警级别
       Message     string    // 告警消息
       Timestamp   int64     // 告警时间
       Status      string    // 告警状态
   }
   ```

4. **同步队列表 (sync_queue)**
   ```go
   type SyncQueue struct {
       ID          int       // 队列ID
       DataType    string    // 数据类型
       Data        string    // 数据内容（JSON）
       Status      string    // 同步状态
       CreatedAt   int64     // 创建时间
       SyncAt      int64     // 同步时间
   }
   ```

**索引设计：**

```go
// 设备表索引
- 主键：id
- 索引：status, created_at

// 传感器数据表索引
- 主键：id
- 复合索引：device_id, timestamp
- 索引：sensor_id, timestamp
- 索引：timestamp

// 告警表索引
- 主键：id
- 索引：device_id, status
- 索引：timestamp, level

// 同步队列表索引
- 主键：id
- 索引：status, created_at
```

## 8.3 数据模型设计

### 8.3.1 设备信息表

**Go 结构体定义：**

```go
package models

import "time"

// Device 设备信息
type Device struct {
    ID          int       `json:"id"`
    Name        string    `json:"name"`
    Type        string    `json:"type"`        // "sensor", "plc", "actuator"
    Protocol    string    `json:"protocol"`    // "modbus", "opcua", "mqtt"
    Status      string    `json:"status"`      // "online", "offline", "error"
    Config      string    `json:"config"`      // JSON 配置
    CreatedAt   int64     `json:"created_at"`
    UpdatedAt   int64     `json:"updated_at"`
}

// NewDevice 创建新设备
func NewDevice(name, deviceType, protocol string) *Device {
    now := time.Now().Unix()
    return &Device{
        Name:      name,
        Type:      deviceType,
        Protocol:  protocol,
        Status:    "offline",
        CreatedAt: now,
        UpdatedAt: now,
    }
}
```

**数据库操作：**

```go
package storage

import (
    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/storage"
)

type DeviceStore struct {
    table *engine.Table
}

func NewDeviceStore(db *storage.DB) (*DeviceStore, error) {
    table, err := engine.TableNew("devices")
    if err != nil {
        return nil, err
    }

    fields := map[string]any{
        "id":         0,
        "name":       "",
        "type":       "",
        "protocol":   "",
        "status":     "",
        "config":     "",
        "created_at": int64(0),
        "updated_at": int64(0),
    }

    if err := table.SetFields(fields); err != nil {
        return nil, err
    }

    // 主键索引
    pk, _ := engine.DefaultPrimaryKeyNew("pk")
    pk.AddFields("id")
    table.CreateIndex(pk)

    // 状态索引
    statusIdx, _ := engine.DefaultNormalIndexNew("status_idx")
    statusIdx.AddFields("status")
    table.CreateIndex(statusIdx)

    return &DeviceStore{table: table}, nil
}

func (s *DeviceStore) Insert(device *Device) (int, error) {
    fields := map[string]any{
        "id":         device.ID,
        "name":       device.Name,
        "type":       device.Type,
        "protocol":   device.Protocol,
        "status":     device.Status,
        "config":     device.Config,
        "created_at": device.CreatedAt,
        "updated_at": device.UpdatedAt,
    }
    return s.table.Insert(&fields)
}

func (s *DeviceStore) GetByID(id int) (*Device, error) {
    iter, err := s.table.Search(&map[string]any{"id": id})
    if err != nil {
        return nil, err
    }
    defer iter.Release()

    records := iter.GetRecords(true)
    defer records.Release()

    if len(records) == 0 {
        return nil, nil
    }

    // 解析记录到 Device 结构体
    return &Device{
        ID:        records[0]["id"].(int),
        Name:      records[0]["name"].(string),
        Type:      records[0]["type"].(string),
        Protocol:  records[0]["protocol"].(string),
        Status:    records[0]["status"].(string),
        Config:    records[0]["config"].(string),
        CreatedAt: records[0]["created_at"].(int64),
        UpdatedAt: records[0]["updated_at"].(int64),
    }, nil
}

func (s *DeviceStore) Update(device *Device) error {
    device.UpdatedAt = time.Now().Unix()
    fields := map[string]any{
        "id":         device.ID,
        "name":       device.Name,
        "status":     device.Status,
        "config":     device.Config,
        "updated_at": device.UpdatedAt,
    }
    return s.table.Update(&fields)
}
```

### 8.3.2 传感器数据表

**Go 结构体定义：**

```go
package models

import "time"

// SensorData 传感器数据
type SensorData struct {
    ID          int       `json:"id"`
    DeviceID    int       `json:"device_id"`
    SensorID    string    `json:"sensor_id"`
    Value       float64   `json:"value"`
    Unit        string    `json:"unit"`
    Timestamp   int64     `json:"timestamp"`
    Quality     string    `json:"quality"` // "good", "bad", "uncertain"
}

// NewSensorData 创建传感器数据
func NewSensorData(deviceID int, sensorID string, value float64, unit string) *SensorData {
    return &SensorData{
        DeviceID:  deviceID,
        SensorID:  sensorID,
        Value:     value,
        Unit:      unit,
        Timestamp: time.Now().UnixNano() / int64(time.Millisecond),
        Quality:   "good",
    }
}
```

**时序数据存储：**

```go
package storage

import (
    "time"
    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/storage"
    "github.com/liaoran123/sfsDb/time"
)

type SensorDataStore struct {
    table     *engine.Table
    timeStore *time.Store
}

func NewSensorDataStore(db *storage.DB) (*SensorDataStore, error) {
    table, err := engine.TableNew("sensor_data")
    if err != nil {
        return nil, err
    }

    fields := map[string]any{
        "id":         0,
        "device_id":  0,
        "sensor_id":  "",
        "value":      0.0,
        "unit":       "",
        "timestamp":  int64(0),
        "quality":    "",
    }

    if err := table.SetFields(fields); err != nil {
        return nil, err
    }

    // 主键索引
    pk, _ := engine.DefaultPrimaryKeyNew("pk")
    pk.AddFields("id")
    table.CreateIndex(pk)

    // 复合索引：设备 + 时间
    deviceTimeIdx, _ := engine.DefaultNormalIndexNew("device_time_idx")
    deviceTimeIdx.AddFields("device_id", "timestamp")
    table.CreateIndex(deviceTimeIdx)

    // 时间索引（用于时序查询）
    timeIdx, _ := engine.DefaultNormalIndexNew("time_idx")
    timeIdx.AddFields("timestamp")
    table.CreateIndex(timeIdx)

    return &SensorDataStore{table: table}, nil
}

func (s *SensorDataStore) Insert(data *SensorData) (int, error) {
    fields := map[string]any{
        "id":         data.ID,
        "device_id":  data.DeviceID,
        "sensor_id":  data.SensorID,
        "value":      data.Value,
        "unit":       data.Unit,
        "timestamp":  data.Timestamp,
        "quality":    data.Quality,
    }
    return s.table.Insert(&fields)
}

func (s *SensorDataStore) BatchInsert(dataList []*SensorData) error {
    for _, data := range dataList {
        _, err := s.Insert(data)
        if err != nil {
            return err
        }
    }
    return nil
}

// QueryByTimeRange 按时间范围查询
func (s *SensorDataStore) QueryByTimeRange(
    deviceID int,
    startTime, endTime int64,
) ([]*SensorData, error) {
    
    iter, err := s.table.Search(&map[string]any{
        "device_id": deviceID,
        "timestamp": startTime,
    })
    if err != nil {
        return nil, err
    }
    defer iter.Release()

    records := iter.GetRecords(true)
    defer records.Release()

    var result []*SensorData
    for _, r := range records {
        ts := r["timestamp"].(int64)
        if ts > endTime {
            continue
        }
        result = append(result, &SensorData{
            ID:        r["id"].(int),
            DeviceID:  r["device_id"].(int),
            SensorID:  r["sensor_id"].(string),
            Value:     r["value"].(float64),
            Unit:      r["unit"].(string),
            Timestamp: ts,
            Quality:   r["quality"].(string),
        })
    }

    return result, nil
}
```

### 8.3.3 告警信息表

**Go 结构体定义：**

```go
package models

import "time"

// Alert 告警信息
type Alert struct {
    ID          int       `json:"id"`
    DeviceID    int       `json:"device_id"`
    Type        string    `json:"type"`        // "threshold", "offline", "error"
    Level       string    `json:"level"`       // "info", "warning", "error", "critical"
    Message     string    `json:"message"`
    Timestamp   int64     `json:"timestamp"`
    Status      string    `json:"status"`      // "active", "acknowledged", "resolved"
    AckAt       int64     `json:"ack_at"`
    ResolvedAt  int64     `json:"resolved_at"`
}

// NewAlert 创建告警
func NewAlert(deviceID int, alertType, level, message string) *Alert {
    return &Alert{
        DeviceID:  deviceID,
        Type:      alertType,
        Level:     level,
        Message:   message,
        Timestamp: time.Now().Unix(),
        Status:    "active",
    }
}
```

### 8.3.4 索引设计

（已在 8.2.4 中详细描述）

## 8.4 核心功能实现

### 8.4.1 设备数据采集

**MQTT 数据采集器：**

```go
package collector

import (
    "encoding/json"
    "log"
    "time"

    "github.com/eclipse/paho.mqtt.golang"
    "your-project/models"
    "your-project/storage"
)

type MqttCollector struct {
    client      mqtt.Client
    dataStore   *storage.SensorDataStore
    deviceStore *storage.DeviceStore
    topicPrefix string
}

func NewMqttCollector(
    broker string,
    topicPrefix string,
    dataStore *storage.SensorDataStore,
    deviceStore *storage.DeviceStore,
) (*MqttCollector, error) {
    
    opts := mqtt.NewClientOptions()
    opts.AddBroker(broker)
    opts.SetClientID("sfsEdgeStore_" + time.Now().String())
    opts.SetAutoReconnect(true)

    client := mqtt.NewClient(opts)
    if token := client.Connect(); token.Wait() && token.Error() != nil {
        return nil, token.Error()
    }

    return &MqttCollector{
        client:      client,
        dataStore:   dataStore,
        deviceStore: deviceStore,
        topicPrefix: topicPrefix,
    }, nil
}

func (c *MqttCollector) Start() error {
    topic := c.topicPrefix + "/device/+/data"
    
    token := c.client.Subscribe(topic, 0, c.messageHandler)
    if token.Wait() && token.Error() != nil {
        return token.Error()
    }

    log.Printf("MQTT 采集器已启动，订阅主题: %s", topic)
    return nil
}

func (c *MqttCollector) messageHandler(client mqtt.Client, msg mqtt.Message) {
    var data struct {
        DeviceID  int     `json:"device_id"`
        SensorID  string  `json:"sensor_id"`
        Value     float64 `json:"value"`
        Unit      string  `json:"unit"`
    }

    if err := json.Unmarshal(msg.Payload(), &data); err != nil {
        log.Printf("解析 MQTT 消息失败: %v", err)
        return
    }

    sensorData := models.NewSensorData(
        data.DeviceID,
        data.SensorID,
        data.Value,
        data.Unit,
    )

    if _, err := c.dataStore.Insert(sensorData); err != nil {
        log.Printf("存储传感器数据失败: %v", err)
    }
}

func (c *MqttCollector) Stop() {
    c.client.Disconnect(250)
}
```

### 8.4.2 数据实时存储

**带缓存的数据存储：**

```go
package storage

import (
    "container/list"
    "sync"
    "your-project/models"
)

type CacheItem struct {
    key   string
    value *models.SensorData
}

type LRUCache struct {
    capacity int
    cache    map[string]*list.Element
    list     *list.List
    mu       sync.RWMutex
}

func NewLRUCache(capacity int) *LRUCache {
    return &LRUCache{
        capacity: capacity,
        cache:    make(map[string]*list.Element),
        list:     list.New(),
    }
}

func (c *LRUCache) Get(key string) (*models.SensorData, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    if elem, ok := c.cache[key]; ok {
        c.list.MoveToFront(elem)
        return elem.Value.(*CacheItem).value, true
    }
    return nil, false
}

func (c *LRUCache) Put(key string, value *models.SensorData) {
    c.mu.Lock()
    defer c.mu.Unlock()

    if elem, ok := c.cache[key]; ok {
        c.list.MoveToFront(elem)
        elem.Value.(*CacheItem).value = value
        return
    }

    if c.list.Len() >= c.capacity {
        back := c.list.Back()
        if back != nil {
            delete(c.cache, back.Value.(*CacheItem).key)
            c.list.Remove(back)
        }
    }

    elem := c.list.PushFront(&CacheItem{key: key, value: value})
    c.cache[key] = elem
}

type CachedSensorDataStore struct {
    store *SensorDataStore
    cache *LRUCache
}

func NewCachedSensorDataStore(store *SensorDataStore, cacheCapacity int) *CachedSensorDataStore {
    return &CachedSensorDataStore{
        store: store,
        cache: NewLRUCache(cacheCapacity),
    }
}

func (s *CachedSensorDataStore) Insert(data *models.SensorData) (int, error) {
    id, err := s.store.Insert(data)
    if err == nil {
        key := dataCacheKey(data.DeviceID, data.SensorID, data.Timestamp)
        s.cache.Put(key, data)
    }
    return id, err
}

func (s *CachedSensorDataStore) GetLatest(deviceID int, sensorID string) (*models.SensorData, error) {
    key := latestDataCacheKey(deviceID, sensorID)
    
    if data, ok := s.cache.Get(key); ok {
        return data, nil
    }

    // 缓存未命中，从数据库查询
    // ... 实现查询最新数据的逻辑
    
    return nil, nil
}

func dataCacheKey(deviceID int, sensorID string, timestamp int64) string {
    return fmt.Sprintf("%d:%s:%d", deviceID, sensorID, timestamp)
}

func latestDataCacheKey(deviceID int, sensorID string) string {
    return fmt.Sprintf("latest:%d:%s", deviceID, sensorID)
}
```

### 8.4.3 本地数据分析

**数据聚合计算：**

```go
package processor

import (
    "math"
    "your-project/models"
    "your-project/storage"
)

type AggregationResult struct {
    Min   float64
    Max   float64
    Avg   float64
    Sum   float64
    Count int
}

type DataProcessor struct {
    dataStore *storage.SensorDataStore
}

func NewDataProcessor(dataStore *storage.SensorDataStore) *DataProcessor {
    return &DataProcessor{dataStore: dataStore}
}

// Aggregate 数据聚合
func (p *DataProcessor) Aggregate(
    deviceID int,
    sensorID string,
    startTime, endTime int64,
) (*AggregationResult, error) {
    
    dataList, err := p.dataStore.QueryByTimeRange(deviceID, startTime, endTime)
    if err != nil {
        return nil, err
    }

    if len(dataList) == 0 {
        return &AggregationResult{}, nil
    }

    result := &AggregationResult{
        Min: math.MaxFloat64,
        Max: -math.MaxFloat64,
    }

    for _, data := range dataList {
        if data.SensorID != sensorID {
            continue
        }

        value := data.Value
        result.Sum += value
        result.Count++
        
        if value < result.Min {
            result.Min = value
        }
        if value > result.Max {
            result.Max = value
        }
    }

    if result.Count > 0 {
        result.Avg = result.Sum / float64(result.Count)
    }

    return result, nil
}
```

### 8.4.4 异常检测与告警

**阈值告警检测：**

```go
package processor

import (
    "log"
    "your-project/models"
    "your-project/storage"
)

type AlertRule struct {
    ID          int
    DeviceID    int
    SensorID    string
    Type        string  // "above_threshold", "below_threshold"
    Threshold   float64
    Level       string  // "warning", "error", "critical"
    Enabled     bool
}

type AlertProcessor struct {
    alertStore  *storage.AlertStore
    rules       []*AlertRule
}

func NewAlertProcessor(alertStore *storage.AlertStore) *AlertProcessor {
    return &AlertProcessor{
        alertStore: alertStore,
        rules:      make([]*AlertRule, 0),
    }
}

func (p *AlertProcessor) AddRule(rule *AlertRule) {
    p.rules = append(p.rules, rule)
}

func (p *AlertProcessor) Process(data *models.SensorData) {
    for _, rule := range p.rules {
        if !rule.Enabled {
            continue
        }
        if rule.DeviceID != data.DeviceID {
            continue
        }
        if rule.SensorID != data.SensorID {
            continue
        }

        triggered := false
        switch rule.Type {
        case "above_threshold":
            if data.Value > rule.Threshold {
                triggered = true
            }
        case "below_threshold":
            if data.Value < rule.Threshold {
                triggered = true
            }
        }

        if triggered {
            p.triggerAlert(rule, data)
        }
    }
}

func (p *AlertProcessor) triggerAlert(rule *AlertRule, data *models.SensorData) {
    message := ""
    switch rule.Type {
    case "above_threshold":
        message = "传感器 %s 数值 %.2f 超过阈值 %.2f"
    case "below_threshold":
        message = "传感器 %s 数值 %.2f 低于阈值 %.2f"
    }

    alert := models.NewAlert(
        rule.DeviceID,
        "threshold",
        rule.Level,
        message,
    )

    if _, err := p.alertStore.Insert(alert); err != nil {
        log.Printf("创建告警失败: %v", err)
    } else {
        log.Printf("告警已触发: %s", alert.Message)
    }
}
```

## 8.5 边缘-云端协同

### 8.5.1 断网续传

**同步队列管理：**

```go
package sync

import (
    "encoding/json"
    "log"
    "sync"
    "time"

    "your-project/models"
    "your-project/storage"
)

type SyncManager struct {
    queueStore   *storage.SyncQueueStore
    cloudClient  *CloudClient
    isOnline     bool
    mu           sync.RWMutex
    stopChan     chan struct{}
}

func NewSyncManager(
    queueStore *storage.SyncQueueStore,
    cloudClient *CloudClient,
) *SyncManager {
    return &SyncManager{
        queueStore:  queueStore,
        cloudClient: cloudClient,
        isOnline:    true,
        stopChan:    make(chan struct{}),
    }
}

func (m *SyncManager) Start() {
    go m.syncLoop()
    go m.monitorNetwork()
}

func (m *SyncManager) Stop() {
    close(m.stopChan)
}

func (m *SyncManager) Enqueue(dataType string, data interface{}) error {
    jsonData, err := json.Marshal(data)
    if err != nil {
        return err
    }

    item := &models.SyncQueue{
        DataType:  dataType,
        Data:      string(jsonData),
        Status:    "pending",
        CreatedAt: time.Now().Unix(),
    }

    return m.queueStore.Insert(item)
}

func (m *SyncManager) syncLoop() {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            m.trySync()
        case <-m.stopChan:
            return
        }
    }
}

func (m *SyncManager) trySync() {
    m.mu.RLock()
    if !m.isOnline {
        m.mu.RUnlock()
        return
    }
    m.mu.RUnlock()

    pendingItems, err := m.queueStore.GetPending(100)
    if err != nil {
        log.Printf("获取待同步数据失败: %v", err)
        return
    }

    for _, item := range pendingItems {
        if err := m.cloudClient.Upload(item.DataType, item.Data); err != nil {
            log.Printf("同步数据失败: %v", err)
            m.setOnline(false)
            return
        }

        if err := m.queueStore.MarkSynced(item.ID); err != nil {
            log.Printf("标记同步状态失败: %v", err)
        }
    }
}

func (m *SyncManager) monitorNetwork() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            if m.cloudClient.Ping() {
                m.setOnline(true)
            } else {
                m.setOnline(false)
            }
        case <-m.stopChan:
            return
        }
    }
}

func (m *SyncManager) setOnline(online bool) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    if m.isOnline != online {
        m.isOnline = online
        log.Printf("网络状态变化: %v", online)
    }
}
```

### 8.5.2 数据同步策略

**批量同步策略：**

```go
package sync

type SyncStrategy interface {
    ShouldSync(item *models.SyncQueue) bool
    GetBatchSize() int
}

type ImmediateSyncStrategy struct{}

func (s *ImmediateSyncStrategy) ShouldSync(item *models.SyncQueue) bool {
    return true // 立即同步
}

func (s *ImmediateSyncStrategy) GetBatchSize() int {
    return 10
}

type BatchSyncStrategy struct {
    batchSize int
    maxWait   time.Duration
}

func (s *BatchSyncStrategy) ShouldSync(item *models.SyncQueue) bool {
    // 根据时间或数量决定是否同步
    return true
}

func (s *BatchSyncStrategy) GetBatchSize() int {
    return s.batchSize
}
```

### 8.5.3 带宽优化

**数据压缩和批量上传：**

```go
package sync

import (
    "bytes"
    "compress/gzip"
    "encoding/json"
)

func compressData(data []byte) ([]byte, error) {
    var buf bytes.Buffer
    gz := gzip.NewWriter(&buf)
    
    if _, err := gz.Write(data); err != nil {
        return nil, err
    }
    if err := gz.Close(); err != nil {
        return nil, err
    }
    
    return buf.Bytes(), nil
}

func batchUpload(items []*models.SyncQueue) error {
    // 将多个数据项合并
    batch := make([]map[string]interface{}, 0, len(items))
    for _, item := range items {
        var data map[string]interface{}
        json.Unmarshal([]byte(item.Data), &data)
        batch = append(batch, data)
    }
    
    // 批量上传
    jsonData, _ := json.Marshal(batch)
    compressed, _ := compressData(jsonData)
    
    // 上传 compressed...
    return nil
}
```

## 8.6 加密与安全

### 8.6.1 数据加密存储

**使用 sfsDb 加密功能：**

```go
package security

import (
    "github.com/liaoran123/sfsDb/storage"
)

func setupEncryptedDB(dbPath string, masterKey []byte) (*storage.DB, error) {
    encryptConfig := &storage.EncryptionConfig{
        Enabled:   true,
        Algorithm: "AES-256-GCM",
        MasterKey: masterKey,
    }

    dbManager := storage.GetDBManager()
    return dbManager.OpenDB(dbPath, encryptConfig)
}

// 密钥派生示例
import (
    "crypto/sha256"
    "golang.org/x/crypto/pbkdf2"
)

func deriveKey(password string, salt []byte) []byte {
    return pbkdf2.Key(
        []byte(password),
        salt,
        10000, // 迭代次数
        32,    // 密钥长度（256位）
        sha256.New,
    )
}
```

### 8.6.2 密钥管理

**安全密钥存储：**

```go
package security

import (
    "os"
    "sync"
)

type KeyManager struct {
    masterKey []byte
    mu        sync.RWMutex
}

func NewKeyManager() *KeyManager {
    return &KeyManager{}
}

func (m *KeyManager) LoadKeyFromEnv() error {
    keyHex := os.Getenv("SFSDB_MASTER_KEY")
    if keyHex == "" {
        return nil
    }

    // 解析 hex key...
    return nil
}

func (m *KeyManager) LoadKeyFromFile(path string) error {
    key, err := os.ReadFile(path)
    if err != nil {
        return err
    }

    m.mu.Lock()
    m.masterKey = key
    m.mu.Unlock()

    return nil
}

func (m *KeyManager) GetKey() []byte {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.masterKey
}
```

### 8.6.3 安全审计

**操作审计日志：**

```go
package security

import (
    "encoding/json"
    "log"
    "os"
    "sync"
    "time"
)

type AuditLog struct {
    Timestamp int64  `json:"timestamp"`
    User      string `json:"user"`
    Action    string `json:"action"`
    Resource  string `json:"resource"`
    Details   string `json:"details"`
    IP        string `json:"ip"`
}

type AuditLogger struct {
    file *os.File
    mu   sync.Mutex
}

func NewAuditLogger(path string) (*AuditLogger, error) {
    file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
    if err != nil {
        return nil, err
    }

    return &AuditLogger{file: file}, nil
}

func (l *AuditLogger) Log(action, resource, details, user, ip string) {
    auditLog := &AuditLog{
        Timestamp: time.Now().Unix(),
        User:      user,
        Action:    action,
        Resource:  resource,
        Details:   details,
        IP:        ip,
    }

    l.mu.Lock()
    defer l.mu.Unlock()

    jsonData, _ := json.Marshal(auditLog)
    l.file.Write(jsonData)
    l.file.WriteString("\n")
}

func (l *AuditLogger) Close() error {
    return l.file.Close()
}
```

## 8.7 性能优化

### 8.7.1 批量操作优化

**批量写入优化：**

```go
package storage

import (
    "github.com/liaoran123/sfsDb/transactionLockFree"
)

func (s *SensorDataStore) BatchInsertOptimized(dataList []*SensorData) error {
    tx, err := transactionLockFree.NewTableTransaction(s.table)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    for _, data := range dataList {
        fields := map[string]interface{}{
            "id":         data.ID,
            "device_id":  data.DeviceID,
            "sensor_id":  data.SensorID,
            "value":      data.Value,
            "unit":       data.Unit,
            "timestamp":  data.Timestamp,
            "quality":    data.Quality,
        }
        if _, err := tx.Insert(&fields); err != nil {
            return err
        }
    }

    return tx.Commit()
}
```

### 8.7.2 索引优化

**索引选择策略：**

```go
// 已经在 8.2.4 和 8.3.4 中详细描述
```

### 8.7.3 缓存策略

**多级缓存架构：**

```go
package cache

import (
    "sync"
    "time"
)

type MultiLevelCache struct {
    l1 *LRUCache      // 内存缓存（热数据）
    l2 *RocksDBCache // 本地缓存（温数据）
    mu sync.RWMutex
}

func (c *MultiLevelCache) Get(key string) (interface{}, bool) {
    // L1 缓存查询
    if val, ok := c.l1.Get(key); ok {
        return val, true
    }

    // L2 缓存查询
    if val, ok := c.l2.Get(key); ok {
        // 回填到 L1
        c.l1.Put(key, val)
        return val, true
    }

    return nil, false
}

func (c *MultiLevelCache) Put(key string, val interface{}) {
    c.l1.Put(key, val)
    c.l2.Put(key, val)
}
```

### 8.7.4 内存优化

**对象池使用：**

```go
package pool

import (
    "sync"
    "your-project/models"
)

var sensorDataPool = sync.Pool{
    New: func() interface{} {
        return &models.SensorData{}
    },
}

func GetSensorData() *models.SensorData {
    return sensorDataPool.Get().(*models.SensorData)
}

func PutSensorData(data *models.SensorData) {
    // 重置字段
    data.ID = 0
    data.DeviceID = 0
    data.SensorID = ""
    data.Value = 0
    data.Unit = ""
    data.Timestamp = 0
    data.Quality = ""
    
    sensorDataPool.Put(data)
}
```

## 8.8 部署与测试

### 8.8.1 环境配置

**配置文件 (config.yaml)：**

```yaml
database:
  path: "/var/lib/sfsEdgeStore/db"
  encryption:
    enabled: true
    master_key_env: "SFSDB_MASTER_KEY"

cache:
  lru_capacity: 10000

mqtt:
  broker: "tcp://localhost:1883"
  topic_prefix: "industrial"

cloud:
  endpoint: "https://api.example.com"
  sync_interval: 5s

logging:
  level: "info"
  file: "/var/log/sfsEdgeStore/app.log"
```

### 8.8.2 Docker 容器化

**Dockerfile：**

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o sfsEdgeStore ./cmd/sfsEdgeStore

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/sfsEdgeStore .
COPY --from=builder /app/config.yaml .

EXPOSE 8080

CMD ["./sfsEdgeStore"]
```

**docker-compose.yml：**

```yaml
version: '3.8'

services:
  sfsedgestore:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - db_data:/var/lib/sfsEdgeStore/db
      - ./config.yaml:/root/config.yaml
    environment:
      - SFSDB_MASTER_KEY=${SFSDB_MASTER_KEY}
    restart: unless-stopped

  mqtt:
    image: eclipse-mosquitto:2.0
    ports:
      - "1883:1883"
    volumes:
      - mosquitto_data:/mosquitto/data

volumes:
  db_data:
  mosquitto_data:
```

### 8.8.3 系统测试

**端到端测试：**

```go
package e2e

import (
    "testing"
    "your-project/collector"
    "your-project/models"
    "your-project/storage"
)

func TestEndToEnd(t *testing.T) {
    // 1. 初始化存储
    db, err := storage.OpenDB("./test_db")
    if err != nil {
        t.Fatal(err)
    }
    defer db.Close()

    // 2. 创建数据存储
    sensorStore, err := storage.NewSensorDataStore(db)
    if err != nil {
        t.Fatal(err)
    }

    // 3. 创建设备
    device := models.NewDevice("测试设备", "sensor", "mqtt")
    deviceStore, _ := storage.NewDeviceStore(db)
    deviceStore.Insert(device)

    // 4. 模拟传感器数据
    data := models.NewSensorData(1, "temperature", 25.5, "°C")
    id, err := sensorStore.Insert(data)
    if err != nil {
        t.Fatal(err)
    }

    // 5. 验证数据存储
    retrieved, err := sensorStore.GetByID(id)
    if err != nil {
        t.Fatal(err)
    }
    if retrieved.Value != 25.5 {
        t.Errorf("期望 25.5，实际 %f", retrieved.Value)
    }
}
```

### 8.8.4 性能测试

**基准测试：**

```go
package benchmarks

import (
    "testing"
    "your-project/models"
    "your-project/storage"
)

func BenchmarkInsert(b *testing.B) {
    db, _ := storage.OpenDB("./bench_db")
    store, _ := storage.NewSensorDataStore(db)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        data := models.NewSensorData(1, "test", float64(i), "unit")
        store.Insert(data)
    }
}

func BenchmarkBatchInsert(b *testing.B) {
    db, _ := storage.OpenDB("./bench_db")
    store, _ := storage.NewSensorDataStore(db)

    batchSize := 100
    dataList := make([]*models.SensorData, batchSize)
    for i := 0; i < batchSize; i++ {
        dataList[i] = models.NewSensorData(1, "test", float64(i), "unit")
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        store.BatchInsert(dataList)
    }
}
```

## 8.9 架构师视角

### 8.9.1 可扩展性设计

**水平扩展策略：**

```
┌─────────────────────────────────────────────────────────┐
│                    负载均衡层                          │
└─────────────────────┬───────────────────────────────────┘
                      │
        ┌─────────────┼─────────────┐
        │             │             │
   ┌────▼───┐   ┌───▼────┐   ┌──▼─────┐
   │ 网关实例1│   │ 网关实例2│   │ 网关实例3│
   │ (分区A) │   │ (分区B) │   │ (分区C) │
   └────┬───┘   └───┬────┘   └──┬─────┘
        │             │             │
   ┌────▼───┐   ┌───▼────┐   ┌──▼─────┐
   │sfsDb    │   │sfsDb    │   │sfsDb    │
   │(分区A)  │   │(分区B)  │   │(分区C)  │
   └────────┘   └────────┘   └────────┘
```

**数据分区策略：**
- 按设备 ID 范围分区
- 按地理位置分区
- 按时间范围分区（冷热数据分离）

### 8.9.2 高可用方案

**主从复制架构：**

```
        ┌─────────────┐
        │   主节点    │
        │ (读写)      │
        └──────┬──────┘
               │
        ┌──────┴──────┐
        │             │
   ┌────▼───┐   ┌───▼────┐
   │ 从节点1 │   │ 从节点2 │
   │ (只读)  │   │ (只读)  │
   └────────┘   └────────┘
```

**故障切换：**
- 使用 Raft 共识算法
- 自动故障检测和选举
- 数据同步保证一致性

### 8.9.3 监控与运维

**监控指标：**

```go
package metrics

import (
    "sync/atomic"
)

type Metrics struct {
    InsertTotal    uint64
    QueryTotal     uint64
    CacheHits      uint64
    CacheMisses    uint64
    SyncSuccess    uint64
    SyncFailures   uint64
    ActiveAlerts   int64
}

var globalMetrics Metrics

func RecordInsert() {
    atomic.AddUint64(&globalMetrics.InsertTotal, 1)
}

func RecordQuery() {
    atomic.AddUint64(&globalMetrics.QueryTotal, 1)
}

func GetMetrics() *Metrics {
    return &globalMetrics
}
```

**Prometheus 集成：**

```go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    insertCounter = promauto.NewCounter(prometheus.CounterOpts{
        Name: "sfsedgestore_inserts_total",
        Help: "Total number of inserts",
    })
    
    queryDuration = promauto.NewHistogram(prometheus.HistogramOpts{
        Name:    "sfsedgestore_query_duration_seconds",
        Help:    "Query duration distribution",
        Buckets: prometheus.DefBuckets,
    })
)
```

## 8.10 完整代码

### 8.10.1 项目结构

```
sfsEdgeStore/
├── cmd/
│   └── sfsEdgeStore/
│       └── main.go           # 主程序入口
├── internal/
│   ├── collector/            # 数据采集模块
│   │   ├── mqtt.go
│   │   ├── modbus.go
│   │   └── opcua.go
│   ├── storage/              # 数据存储模块
│   │   ├── device.go
│   │   ├── sensor_data.go
│   │   ├── alert.go
│   │   └── cache.go
│   ├── processor/            # 数据处理模块
│   │   ├── aggregation.go
│   │   └── alert.go
│   ├── sync/                 # 数据同步模块
│   │   ├── manager.go
│   │   └── cloud.go
│   ├── api/                  # API 模块
│   │   ├── router.go
│   │   ├── device.go
│   │   ├── data.go
│   │   └── alert.go
│   ├── models/               # 数据模型
│   │   ├── device.go
│   │   ├── sensor_data.go
│   │   └── alert.go
│   └── config/               # 配置管理
│       └── config.go
├── pkg/                      # 公共库
│   └── ...
├── config/
│   └── config.yaml
├── deployments/
│   ├── docker/
│   │   ├── Dockerfile
│   │   └── docker-compose.yml
│   └── k8s/
│       └── deployment.yaml
├── scripts/
│   └── init.sh
├── test/
│   ├── e2e/
│   └── benchmarks/
├── go.mod
├── go.sum
└── README.md
```

### 8.10.2 核心代码

**主程序入口 (main.go)：**

```go
package main

import (
    "log"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/spf13/viper"

    "your-project/internal/api"
    "your-project/internal/collector"
    "your-project/internal/config"
    "your-project/internal/processor"
    "your-project/internal/storage"
    "your-project/internal/sync"
)

func main() {
    log.Println("sfsEdgeStore 启动中...")

    // 1. 加载配置
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("加载配置失败: %v", err)
    }

    // 2. 初始化数据库
    dbManager := storage.GetDBManager()
    db, err := dbManager.OpenDB(cfg.Database.Path)
    if err != nil {
        log.Fatalf("打开数据库失败: %v", err)
    }
    defer dbManager.CloseAllDB()

    // 3. 初始化存储
    deviceStore, err := storage.NewDeviceStore(db)
    if err != nil {
        log.Fatal(err)
    }
    sensorStore, err := storage.NewSensorDataStore(db)
    if err != nil {
        log.Fatal(err)
    }
    alertStore, err := storage.NewAlertStore(db)
    if err != nil {
        log.Fatal(err)
    }

    // 4. 初始化缓存
    cachedSensorStore := storage.NewCachedSensorDataStore(sensorStore, 10000)

    // 5. 初始化处理器
    dataProcessor := processor.NewDataProcessor(sensorStore)
    alertProcessor := processor.NewAlertProcessor(alertStore)

    // 6. 初始化云同步
    cloudClient := sync.NewCloudClient(cfg.Cloud.Endpoint)
    syncManager := sync.NewSyncManager(alertStore, cloudClient)
    syncManager.Start()
    defer syncManager.Stop()

    // 7. 初始化采集器
    mqttCollector, err := collector.NewMqttCollector(
        cfg.MQTT.Broker,
        cfg.MQTT.TopicPrefix,
        cachedSensorStore,
        deviceStore,
    )
    if err != nil {
        log.Fatal(err)
    }
    mqttCollector.Start()
    defer mqttCollector.Stop()

    // 8. 初始化 API
    router := api.SetupRouter(
        deviceStore,
        cachedSensorStore,
        alertStore,
        dataProcessor,
    )

    // 9. 启动服务
    log.Printf("API 服务启动，监听端口: %d", cfg.Server.Port)
    if err := router.Run(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil {
        log.Fatalf("启动服务失败: %v", err)
    }
}
```

**API 路由 (router.go)：**

```go
package api

import (
    "github.com/gin-gonic/gin"
)

func SetupRouter(
    deviceStore *storage.DeviceStore,
    sensorStore *storage.CachedSensorDataStore,
    alertStore *storage.AlertStore,
    dataProcessor *processor.DataProcessor,
) *gin.Engine {
    
    r := gin.Default()

    // 健康检查
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    // 设备 API
    devices := r.Group("/api/v1/devices")
    {
        devices.GET("", listDevices(deviceStore))
        devices.GET("/:id", getDevice(deviceStore))
        devices.POST("", createDevice(deviceStore))
        devices.PUT("/:id", updateDevice(deviceStore))
        devices.DELETE("/:id", deleteDevice(deviceStore))
    }

    // 数据 API
    data := r.Group("/api/v1/data")
    {
        data.GET("/latest", getLatestData(sensorStore))
        data.GET("/history", queryHistoryData(sensorStore))
        data.GET("/aggregate", aggregateData(dataProcessor))
    }

    // 告警 API
    alerts := r.Group("/api/v1/alerts")
    {
        alerts.GET("", listAlerts(alertStore))
        alerts.GET("/:id", getAlert(alertStore))
        alerts.PUT("/:id/ack", acknowledgeAlert(alertStore))
        alerts.PUT("/:id/resolve", resolveAlert(alertStore))
    }

    return r
}
```

### 8.10.3 配置文件

**完整配置示例 (config.full.yaml)：**

```yaml
server:
  port: 8080
  mode: release

database:
  path: "/var/lib/sfsEdgeStore/db"
  encryption:
    enabled: true
    algorithm: "AES-256-GCM"
    master_key_env: "SFSDB_MASTER_KEY"

cache:
  lru_capacity: 10000
  ttl: 300

mqtt:
  broker: "tcp://mqtt-broker:1883"
  client_id: "sfsEdgeStore"
  topic_prefix: "industrial"
  qos: 1

cloud:
  endpoint: "https://api.example.com/v1"
  api_key: "${CLOUD_API_KEY}"
  sync_interval: 5s
  batch_size: 100

alert:
  enabled: true
  rules_file: "/etc/sfsEdgeStore/alert_rules.yaml"

logging:
  level: "info"
  format: "json"
  file: "/var/log/sfsEdgeStore/app.log"
  max_size: 100
  max_backups: 10
  max_age: 30

metrics:
  enabled: true
  prometheus:
    enabled: true
    path: "/metrics"
```

## 8.11 本章小结

本章通过一个完整的 **sfsEdgeStore** 工业物联网项目，综合运用了前面各章的知识：

✅ **项目需求分析**：场景描述、功能需求、非功能需求、技术选型
✅ **架构设计**：整体架构、模块划分、数据流、数据库设计
✅ **数据模型设计**：设备信息、传感器数据、告警信息、同步队列
✅ **核心功能实现**：数据采集、实时存储、本地分析、异常检测
✅ **边缘-云端协同**：断网续传、数据同步、带宽优化
✅ **加密与安全**：数据加密、密钥管理、安全审计
✅ **性能优化**：批量操作、索引优化、缓存策略、内存优化
✅ **部署与测试**：环境配置、Docker 容器化、系统测试、性能测试
✅ **架构师视角**：可扩展性、高可用、监控与运维
✅ **完整代码**：项目结构、核心代码、配置文件

**关键要点：**
1. sfsDb 非常适合边缘计算场景，轻量、高效、可靠
2. 合理使用缓存可以显著提升查询性能
3. 断网续传是工业物联网的必备功能
4. 数据安全和加密在边缘场景中尤为重要
5. 监控和运维是生产环境的关键保障

**sfsEdgeStore 的核心价值：**
- 作为 EdgeX Foundry 与 sfsDb 之间的桥梁
- 提供高效的本地数据读写和缓存能力
- 实现数据持久化、批量处理、断网续传
- 为工业物联网应用提供可靠的数据基础设施

---

**全书完**

感谢您阅读《sfsDb 深度解析：构建高性能工业物联网边缘计算嵌入式数据库应用》！
