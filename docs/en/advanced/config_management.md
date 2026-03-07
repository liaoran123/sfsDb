# Storage Configuration Management

This document details the configuration management features of the sfsDb storage package, including scenario configurations, custom configurations, and configuration management APIs.

## Overview

The sfsDb storage package provides a flexible configuration management system that supports:

- **Predefined Scenario Configurations**: Built-in optimized configurations for various application scenarios
- **Custom Configurations**: Create and use custom configurations based on specific needs
- **Runtime Configuration**: Dynamically adjust configurations while the database is running

## Scenario Configurations

sfsDb provides several predefined scenario configurations optimized for different application scenarios.

### Built-in Scenarios

| Scenario Name | Constant | Use Case | Characteristics |
|---------------|----------|----------|-----------------|
| Default Scenario | `ScenarioDefault` | General purpose | Balanced performance and resource usage |
| Embedded Scenario | `ScenarioEmbedded` | Resource-constrained embedded devices | Low memory footprint, low power |
| IoT Scenario | `ScenarioIoT` | IoT devices | Low power, suitable for time-series data |
| Edge Scenario | `ScenarioEdge` | Edge computing nodes | Balanced performance and resources |
| Game Scenario | `ScenarioGame` | Gaming applications | Low latency, high concurrency |

### Getting Scenario Configurations

Use the `GetScenarioConfig()` function to get scenario configurations:

```go
package main

import (
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // Get Embedded scenario configuration
    embeddedConfig := storage.GetScenarioConfig(storage.ScenarioEmbedded)
    
    // Get IoT scenario configuration
    iotConfig := storage.GetScenarioConfig(storage.ScenarioIoT)
    
    // Get Edge scenario configuration
    edgeConfig := storage.GetScenarioConfig(storage.ScenarioEdge)
    
    // Get Game scenario configuration
    gameConfig := storage.GetScenarioConfig(storage.ScenarioGame)
    
    // Get Default scenario configuration
    defaultConfig := storage.GetScenarioConfig(storage.ScenarioDefault)
}
```

### Getting LevelDB Options for Scenarios

Use the `GetScenarioOptions()` function to directly get LevelDB options for scenario configurations:

```go
package main

import (
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // Directly get LevelDB options for a scenario
    opts := storage.GetScenarioOptions(storage.ScenarioGame)
    
    // Use opts to create LevelDB
    // ...
}
```

## Config Struct

The `Config` struct contains the following configuration items:

```go
type Config struct {
    WriteBuffer            int           // Write buffer size (bytes)
    OpenFilesCacheCapacity int           // Open files cache capacity
    BlockCacheCapacity     int           // Block cache capacity (bytes)
    Compression            opt.Compression // Compression type
}
```

### Configuration Item Description

| Configuration Item | Description | Default Value |
|-------------------|-------------|---------------|
| `WriteBuffer` | Write buffer size, affects write performance | 64MB |
| `OpenFilesCacheCapacity` | Open files cache capacity, affects concurrent reads | 200 |
| `BlockCacheCapacity` | Block cache capacity, affects read performance | 128MB |
| `Compression` | Compression type, affects storage space and performance | `opt.DefaultCompression` |

## Creating Configurations

### Creating Default Configuration

Use `GetScenarioConfig(ScenarioDefault)` to create a configuration with default values:

```go
package main

import (
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // Create default configuration
    config := storage.GetScenarioConfig(storage.ScenarioDefault)
}
```

### Creating Fully Custom Configuration

Directly construct the `Config` struct to create a fully custom configuration:

```go
package main

import (
    "github.com/liaoran123/sfsDb/storage"
    "github.com/syndtr/goleveldb/leveldb/opt"
)

func main() {
    // Create fully custom configuration
    customConfig := storage.Config{
        WriteBuffer:            128 * 1024 * 1024, // WriteBuffer: 128MB
        OpenFilesCacheCapacity: 500,                  // OpenFilesCacheCapacity: 500
        BlockCacheCapacity:     256 * 1024 * 1024, // BlockCacheCapacity: 256MB
        Compression:            opt.NoCompression,    // Compression: No compression
    }
}
```

### Partially Custom Configuration

Get the default configuration first, then modify only what you need:

```go
package main

import (
    "github.com/liaoran123/sfsDb/storage"
    "github.com/syndtr/goleveldb/leveldb/opt"
)

func main() {
    // Get default configuration
    config := storage.GetScenarioConfig(storage.ScenarioDefault)
    
    // Modify only what you need
    config.WriteBuffer = 32 * 1024 * 1024  // Only modify write buffer
    config.Compression = opt.SnappyCompression // Only modify compression type
}
```

## Configuration Manager

### Getting Configuration Manager

```go
package main

import (
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // Get configuration manager instance
    cm := storage.GetConfigManager()
}
```

### Setting Current Configuration

```go
package main

import (
    "github.com/liaoran123/sfsDb/storage"
    "github.com/syndtr/goleveldb/leveldb/opt"
)

func main() {
    // Create custom configuration
    customConfig := storage.Config{
        WriteBuffer:            128 * 1024 * 1024,
        OpenFilesCacheCapacity: 500,
        BlockCacheCapacity:     256 * 1024 * 1024,
        Compression:            opt.NoCompression,
    }
    
    // Set as current configuration
    storage.SetConfig(customConfig)
}
```

**Note**: When setting a configuration, if a configuration item's value is 0 or negative, the original value is retained.

### Getting Current Configuration

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // Get current configuration
    config := storage.GetConfig()
    
    fmt.Printf("WriteBuffer: %d\n", config.WriteBuffer)
    fmt.Printf("OpenFilesCacheCapacity: %d\n", config.OpenFilesCacheCapacity)
    fmt.Printf("BlockCacheCapacity: %d\n", config.BlockCacheCapacity)
    fmt.Printf("Compression: %v\n", config.Compression)
}
```

### Getting LevelDB Options

```go
package main

import (
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // Get configuration manager
    cm := storage.GetConfigManager()
    
    // Get LevelDB options
    opts := cm.GetOptions()
    
    // Use opts to create LevelDB
    // ...
}
```

## Scenario Configuration Details

### Default Scenario (ScenarioDefault)

Suitable for most general-purpose scenarios, balancing performance and resource usage.

| Configuration Item | Value |
|-------------------|-------|
| WriteBuffer | 64 MB |
| OpenFilesCacheCapacity | 200 |
| BlockCacheCapacity | 128 MB |
| Compression | DefaultCompression |

### Embedded Scenario (ScenarioEmbedded)

Suitable for resource-constrained embedded devices, such as smartphones, embedded systems, etc.

| Configuration Item | Value |
|-------------------|-------|
| WriteBuffer | 2 MB |
| OpenFilesCacheCapacity | 5 |
| BlockCacheCapacity | 4 MB |
| Compression | DefaultCompression |

### IoT Scenario (ScenarioIoT)

Suitable for IoT devices, such as sensors, smart gateways, etc., requiring low power and high concurrency writes.

| Configuration Item | Value |
|-------------------|-------|
| WriteBuffer | 4 MB |
| OpenFilesCacheCapacity | 10 |
| BlockCacheCapacity | 8 MB |
| Compression | DefaultCompression |

### Edge Scenario (ScenarioEdge)

Suitable for edge computing nodes, such as edge servers, edge gateways, etc., requiring balanced performance and resources.

| Configuration Item | Value |
|-------------------|-------|
| WriteBuffer | 16 MB |
| OpenFilesCacheCapacity | 50 |
| BlockCacheCapacity | 32 MB |
| Compression | DefaultCompression |

### Game Scenario (ScenarioGame)

Suitable for gaming applications, such as game servers, game clients, etc., requiring low latency and high concurrency.

| Configuration Item | Value |
|-------------------|-------|
| WriteBuffer | 64 MB |
| OpenFilesCacheCapacity | 200 |
| BlockCacheCapacity | 128 MB |
| Compression | NoCompression |

## Complete Example

Here's a complete configuration management usage example:

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
    "github.com/syndtr/goleveldb/leveldb/opt"
)

func main() {
    // 1. Use scenario configuration
    fmt.Println("=== Using Scenario Configuration ===")
    gameConfig := storage.GetScenarioConfig(storage.ScenarioGame)
    fmt.Printf("Game scenario config: %+v\n", gameConfig)
    
    // 2. Create custom configuration
    fmt.Println("\n=== Creating Custom Configuration ===")
    customConfig := storage.Config{
        WriteBuffer:            128 * 1024 * 1024,
        OpenFilesCacheCapacity: 500,
        BlockCacheCapacity:     256 * 1024 * 1024,
        Compression:            opt.NoCompression,
    }
    fmt.Printf("Custom config: %+v\n", customConfig)
    
    // 3. Set current configuration
    fmt.Println("\n=== Setting Current Configuration ===")
    storage.SetConfig(customConfig)
    
    // 4. Get current configuration
    fmt.Println("\n=== Getting Current Configuration ===")
    currentConfig := storage.GetConfig()
    fmt.Printf("Current config: %+v\n", currentConfig)
    
    // 5. Get LevelDB options
    fmt.Println("\n=== Getting LevelDB Options ===")
    cm := storage.GetConfigManager()
    opts := cm.GetOptions()
    fmt.Printf("LevelDB options: %+v\n", opts)
    
    // 6. Partially modify configuration
    fmt.Println("\n=== Partially Modifying Configuration ===")
    partialConfig := storage.GetScenarioConfig(storage.ScenarioDefault)
    partialConfig.WriteBuffer = 32 * 1024 * 1024 // Only modify write buffer
    storage.SetConfig(partialConfig)
    
    updatedConfig := storage.GetConfig()
    fmt.Printf("Updated config: %+v\n", updatedConfig)
}
```

## Best Practices

### 1. Choose the Right Scenario

Select the appropriate predefined configuration based on your application scenario:

- **Embedded devices**: Use `ScenarioEmbedded`
- **IoT devices**: Use `ScenarioIoT`
- **Edge computing**: Use `ScenarioEdge`
- **Gaming applications**: Use `ScenarioGame`
- **General applications**: Use `ScenarioDefault`

### 2. Progressive Optimization

Start with a predefined scenario configuration, then fine-tune based on actual performance test results:

```go
// Start with scenario configuration
config := storage.GetScenarioConfig(storage.ScenarioGame)

// Fine-tune based on test results
config.WriteBuffer = 128 * 1024 * 1024 // Increase write buffer

storage.SetConfig(config)
```

### 3. Monitor Performance

Monitor performance metrics based on actual usage and adjust configuration parameters:

- High write latency → Increase `WriteBuffer`
- High read latency → Increase `BlockCacheCapacity`
- High memory usage → Decrease cache configurations
- Tight storage space → Enable compression

## Summary

The sfsDb storage package provides a flexible and powerful configuration management system. Through the combination of predefined scenario configurations and custom configurations, it can meet the needs of various application scenarios. Choosing the right configuration can significantly improve database performance and resource usage efficiency.
