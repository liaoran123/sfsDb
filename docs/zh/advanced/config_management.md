# 存储配置管理

本文档详细介绍 sfsDb storage 包的配置管理功能，包括场景配置、自定义配置和配置管理 API。

## 概述

sfsDb storage 包提供了灵活的配置管理系统，支持：

- **预定义场景配置**：内置多种适用场景的优化配置
- **自定义配置**：根据具体需求创建和使用自定义配置
- **运行时配置**：在数据库运行时动态调整配置

## 场景配置

sfsDb 提供了多种预定义的场景配置，针对不同的应用场景进行了优化。

### 内置场景

| 场景名称 | 常量 | 适用场景 | 特点 |
|---------|------|---------|------|
| 默认场景 | `ScenarioDefault` | 通用场景 | 平衡的性能和资源使用 |
| 嵌入式场景 | `ScenarioEmbedded` | 资源受限的嵌入式设备 | 低内存占用，低功耗 |
| IoT 场景 | `ScenarioIoT` | 物联网设备 | 低功耗，适合时序数据 |
| 边缘场景 | `ScenarioEdge` | 边缘计算节点 | 平衡性能和资源 |
| 游戏场景 | `ScenarioGame` | 游戏应用 | 低延迟，高并发 |

### 获取场景配置

使用 `GetScenarioConfig()` 函数获取场景配置：

```go
package main

import (
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // 获取 Embedded 场景配置
    embeddedConfig := storage.GetScenarioConfig(storage.ScenarioEmbedded)
    
    // 获取 IoT 场景配置
    iotConfig := storage.GetScenarioConfig(storage.ScenarioIoT)
    
    // 获取 Edge 场景配置
    edgeConfig := storage.GetScenarioConfig(storage.ScenarioEdge)
    
    // 获取 Game 场景配置
    gameConfig := storage.GetScenarioConfig(storage.ScenarioGame)
    
    // 获取 Default 场景配置
    defaultConfig := storage.GetScenarioConfig(storage.ScenarioDefault)
}
```

### 获取场景配置的 LevelDB 选项

使用 `GetScenarioOptions()` 函数直接获取场景配置的 LevelDB 选项：

```go
package main

import (
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // 直接获取场景的 LevelDB 选项
    opts := storage.GetScenarioOptions(storage.ScenarioGame)
    
    // 使用 opts 创建 LevelDB
    // ...
}
```

## Config 结构体

`Config` 结构体包含以下配置项：

```go
type Config struct {
    WriteBuffer            int           // 写缓冲区大小（字节）
    OpenFilesCacheCapacity int           // 打开文件缓存容量
    BlockCacheCapacity     int           // 块缓存容量（字节）
    Compression            opt.Compression // 压缩类型
}
```

### 配置项说明

| 配置项 | 说明 | 默认值 |
|-------|------|--------|
| `WriteBuffer` | 写缓冲区大小，影响写入性能 | 64MB |
| `OpenFilesCacheCapacity` | 打开文件缓存容量，影响并发读取 | 200 |
| `BlockCacheCapacity` | 块缓存容量，影响读取性能 | 128MB |
| `Compression` | 压缩类型，影响存储空间和性能 | `opt.DefaultCompression` |

## 创建配置

### 创建默认配置

使用 `GetScenarioConfig(ScenarioDefault)` 创建使用默认值的配置：

```go
package main

import (
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // 创建默认配置
    config := storage.GetScenarioConfig(storage.ScenarioDefault)
}
```

### 创建完全自定义配置

直接构造 `Config` 结构体创建完全自定义的配置：

```go
package main

import (
    "github.com/liaoran123/sfsDb/storage"
    "github.com/syndtr/goleveldb/leveldb/opt"
)

func main() {
    // 创建完全自定义的配置
    customConfig := storage.Config{
        WriteBuffer:            128 * 1024 * 1024, // WriteBuffer: 128MB
        OpenFilesCacheCapacity: 500,                  // OpenFilesCacheCapacity: 500
        BlockCacheCapacity:     256 * 1024 * 1024, // BlockCacheCapacity: 256MB
        Compression:            opt.NoCompression,    // Compression: 不压缩
    }
}
```

### 部分自定义配置

先获取默认配置，然后只修改需要的部分：

```go
package main

import (
    "github.com/liaoran123/sfsDb/storage"
    "github.com/syndtr/goleveldb/leveldb/opt"
)

func main() {
    // 获取默认配置
    config := storage.GetScenarioConfig(storage.ScenarioDefault)
    
    // 只修改需要的部分
    config.WriteBuffer = 32 * 1024 * 1024  // 只修改写缓冲区
    config.Compression = opt.SnappyCompression // 只修改压缩类型
}
```

## 配置管理器

### 获取配置管理器

```go
package main

import (
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // 获取配置管理器实例
    cm := storage.GetConfigManager()
}
```

### 设置当前配置

```go
package main

import (
    "github.com/liaoran123/sfsDb/storage"
    "github.com/syndtr/goleveldb/leveldb/opt"
)

func main() {
    // 创建自定义配置
    customConfig := storage.Config{
        WriteBuffer:            128 * 1024 * 1024,
        OpenFilesCacheCapacity: 500,
        BlockCacheCapacity:     256 * 1024 * 1024,
        Compression:            opt.NoCompression,
    }
    
    // 设置为当前配置
    storage.SetConfig(customConfig)
}
```

**注意**：设置配置时，如果某个配置项的值为 0 或负数，会保留原来的值。

### 获取当前配置

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // 获取当前配置
    config := storage.GetConfig()
    
    fmt.Printf("WriteBuffer: %d\n", config.WriteBuffer)
    fmt.Printf("OpenFilesCacheCapacity: %d\n", config.OpenFilesCacheCapacity)
    fmt.Printf("BlockCacheCapacity: %d\n", config.BlockCacheCapacity)
    fmt.Printf("Compression: %v\n", config.Compression)
}
```

### 获取 LevelDB 选项

```go
package main

import (
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // 获取配置管理器
    cm := storage.GetConfigManager()
    
    // 获取 LevelDB 选项
    opts := cm.GetOptions()
    
    // 使用 opts 创建 LevelDB
    // ...
}
```

## 场景配置详情

### 默认场景 (ScenarioDefault)

适用于大多数通用场景，平衡了性能和资源使用。

| 配置项 | 值 |
|-------|-----|
| WriteBuffer | 64 MB |
| OpenFilesCacheCapacity | 200 |
| BlockCacheCapacity | 128 MB |
| Compression | DefaultCompression |

### 嵌入式场景 (ScenarioEmbedded)

适用于资源受限的嵌入式设备，如智能手机、嵌入式系统等。

| 配置项 | 值 |
|-------|-----|
| WriteBuffer | 2 MB |
| OpenFilesCacheCapacity | 5 |
| BlockCacheCapacity | 4 MB |
| Compression | DefaultCompression |

### IoT 场景 (ScenarioIoT)

适用于物联网设备，如传感器、智能网关等，需要低功耗和高并发写入。

| 配置项 | 值 |
|-------|-----|
| WriteBuffer | 4 MB |
| OpenFilesCacheCapacity | 10 |
| BlockCacheCapacity | 8 MB |
| Compression | DefaultCompression |

### 边缘场景 (ScenarioEdge)

适用于边缘计算节点，如边缘服务器、边缘网关等，需要平衡性能和资源。

| 配置项 | 值 |
|-------|-----|
| WriteBuffer | 16 MB |
| OpenFilesCacheCapacity | 50 |
| BlockCacheCapacity | 32 MB |
| Compression | DefaultCompression |

### 游戏场景 (ScenarioGame)

适用于游戏应用，如游戏服务器、游戏客户端等，需要低延迟和高并发。

| 配置项 | 值 |
|-------|-----|
| WriteBuffer | 64 MB |
| OpenFilesCacheCapacity | 200 |
| BlockCacheCapacity | 128 MB |
| Compression | NoCompression |

## 完整示例

以下是一个完整的配置管理使用示例：

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
    "github.com/syndtr/goleveldb/leveldb/opt"
)

func main() {
    // 1. 使用场景配置
    fmt.Println("=== 使用场景配置 ===")
    gameConfig := storage.GetScenarioConfig(storage.ScenarioGame)
    fmt.Printf("Game 场景配置: %+v\n", gameConfig)
    
    // 2. 创建自定义配置
    fmt.Println("\n=== 创建自定义配置 ===")
    customConfig := storage.Config{
        WriteBuffer:            128 * 1024 * 1024,
        OpenFilesCacheCapacity: 500,
        BlockCacheCapacity:     256 * 1024 * 1024,
        Compression:            opt.NoCompression,
    }
    fmt.Printf("自定义配置: %+v\n", customConfig)
    
    // 3. 设置当前配置
    fmt.Println("\n=== 设置当前配置 ===")
    storage.SetConfig(customConfig)
    
    // 4. 获取当前配置
    fmt.Println("\n=== 获取当前配置 ===")
    currentConfig := storage.GetConfig()
    fmt.Printf("当前配置: %+v\n", currentConfig)
    
    // 5. 获取 LevelDB 选项
    fmt.Println("\n=== 获取 LevelDB 选项 ===")
    cm := storage.GetConfigManager()
    opts := cm.GetOptions()
    fmt.Printf("LevelDB 选项: %+v\n", opts)
    
    // 6. 部分修改配置
    fmt.Println("\n=== 部分修改配置 ===")
    partialConfig := storage.GetScenarioConfig(storage.ScenarioDefault)
    partialConfig.WriteBuffer = 32 * 1024 * 1024 // 只修改写缓冲区
    storage.SetConfig(partialConfig)
    
    updatedConfig := storage.GetConfig()
    fmt.Printf("更新后的配置: %+v\n", updatedConfig)
}
```

## 最佳实践

### 1. 选择合适的场景

根据您的应用场景选择合适的预定义配置：

- **嵌入式设备**：使用 `ScenarioEmbedded`
- **物联网设备**：使用 `ScenarioIoT`
- **边缘计算**：使用 `ScenarioEdge`
- **游戏应用**：使用 `ScenarioGame`
- **通用应用**：使用 `ScenarioDefault`

### 2. 渐进式优化

先使用预定义场景配置，然后根据实际性能测试结果进行微调：

```go
// 先使用场景配置
config := storage.GetScenarioConfig(storage.ScenarioGame)

// 根据测试结果微调
config.WriteBuffer = 128 * 1024 * 1024 // 增大写缓冲区

storage.SetConfig(config)
```

### 3. 监控性能

根据实际使用情况监控性能指标，调整配置参数：

- 写入延迟高 → 增大 `WriteBuffer`
- 读取延迟高 → 增大 `BlockCacheCapacity`
- 内存占用高 → 减小缓存配置
- 存储空间紧张 → 启用压缩

## 总结

sfsDb storage 包提供了灵活且强大的配置管理系统，通过预定义场景配置和自定义配置的组合，可以满足各种应用场景的需求。选择合适的配置可以显著提高数据库性能和资源使用效率。
