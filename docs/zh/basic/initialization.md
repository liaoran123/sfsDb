# 数据库初始化

## 1.1 自定义数据库路径

默认情况下，sfsDb会使用当前目录下的`kvdb`文件夹作为数据库存储路径。如果需要自定义数据库路径，可以在程序启动时调用`OpenDefaultDb`函数：

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // 自定义数据库路径
    dbPath := "./my_custom_db"
    _, err := storage.OpenDefaultDb(dbPath)
    if err != nil {
        panic(err)
    }
    
    fmt.Println("数据库初始化成功")
}
```

## 1.2 自动使用默认路径

如果不调用`OpenDefaultDb`函数，系统会在首次创建表时自动使用默认路径`./kvdb`：

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
)

func main() {
    // 首次调用TableNew时，会自动初始化数据库，使用默认路径"./kvdb"
    table, err := engine.TableNew("users")
    if err != nil {
        panic(err)
    }
    
    fmt.Println("表创建成功，数据库已自动初始化")
}
```

## 1.3 数据库关闭

程序结束时，可以调用`CloseDb`函数关闭数据库，释放资源：

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // 创建表
    table, err := engine.TableNew("users")
    if err != nil {
        panic(err)
    }
    
    // 程序结束时关闭数据库
    defer storage.CloseDb()
    
    fmt.Println("表创建成功")
}
```

## 1.4 使用外部存储实例

除了使用内置的存储引擎外，sfsDb 还支持使用外部实现的存储实例。只需通过`SetStore`函数设置即可：

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

// 自定义存储实现
type CustomStore struct {
    // 实现存储逻辑
}

// 实现 Store 接口的方法
func (s *CustomStore) Get(key []byte) ([]byte, error) {
    // 实现获取逻辑
    return nil, nil
}

func (s *CustomStore) Put(key []byte, value []byte) error {
    // 实现存储逻辑
    return nil
}

func (s *CustomStore) Delete(key []byte) error {
    // 实现删除逻辑
    return nil
}

func (s *CustomStore) Batch() storage.Batch {
    // 实现批量操作逻辑
    return nil
}

func (s *CustomStore) Iterator(para ...[]byte) storage.Iterator {
    // 实现迭代器逻辑
    return nil
}

func (s *CustomStore) Snapshot() (storage.Snapshot, error) {
    // 实现快照逻辑
    return nil, nil
}

func (s *CustomStore) Close() error {
    // 实现关闭逻辑
    return nil
}

func main() {
    // 创建自定义存储实例
    customStore := &CustomStore{}
    
    // 设置外部存储实例
    storage.SetStore(customStore)
    
    fmt.Println("外部存储实例设置成功")
    
    // 现在整个 sfsdb 项目都会使用这个自定义存储实例
    // 例如，创建表、插入数据等操作都会通过这个实例执行
}
```

**使用场景**：
- 集成第三方存储实现
- 为特定场景定制存储逻辑
- 在测试中使用内存存储或模拟存储
- 实现特殊的存储功能，如加密、压缩等

**注意事项**：
- 使用外部存储实例时，需要自行管理其生命周期
- 确保在不再使用时正确关闭存储实例
- 外部存储实现必须完整实现`Store`接口的所有方法

## 1.6 使用场景配置

sfsDb 提供了预定义的场景配置，可以根据不同的使用场景选择合适的内存和性能配置，特别适合边缘计算和 IoT 设备。

### 1.6.1 可用场景

| 场景常量 | 说明 | 总内存 | 适用场景 |
|---------|------|--------|----------|
| `ScenarioEmbedded` | 嵌入式设备 | ~6MB | 智能终端设备 |
| `ScenarioIoT` | IoT 设备 | ~12MB | IoT 网关设备 |
| `ScenarioEdge` | 边缘计算节点 | ~48MB | 边缘计算节点 ⭐ |
| `ScenarioGame` | 游戏服务器 | ~192MB | 高性能场景 |

### 1.6.2 使用场景配置打开数据库

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // 方式1：使用 DBManager 并指定场景
    dbManager := storage.GetDBManager()
    db, err := dbManager.OpenDBWithScenario("./edge_db", storage.ScenarioEdge)
    if err != nil {
        panic(err)
    }
    defer dbManager.CloseDB()
    
    fmt.Println("数据库初始化成功，使用边缘计算场景配置")
}
```

### 1.6.3 使用向后兼容的全局函数

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // 使用全局函数并指定场景
    _, err := storage.OpenDefaultDbWithScenario("./edge_db", storage.ScenarioEdge)
    if err != nil {
        panic(err)
    }
    defer storage.CloseDb()
    
    fmt.Println("数据库初始化成功，使用边缘计算场景配置")
}
```

### 1.6.4 场景配置与加密结合使用

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // 创建加密配置
    encryptConfig := &storage.EncryptionConfig{
        Enabled: true,
        Key:     []byte("your-secure-encryption-key-32bytes"),
    }
    
    // 方式1：使用 DBManager 同时指定场景和加密
    dbManager := storage.GetDBManager()
    db, err := dbManager.OpenDBWithScenarioAndEncryption(
        "./edge_db", 
        storage.ScenarioEdge, 
        encryptConfig
    )
    if err != nil {
        panic(err)
    }
    defer dbManager.CloseDB()
    
    fmt.Println("数据库初始化成功，使用边缘计算场景配置并启用加密")
}
```

### 1.6.5 直接使用场景配置创建 LevelDB 存储

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // 方式1：仅使用场景配置
    db, err := storage.NewLevelDBStoreWithScenario("./edge_db", storage.ScenarioEdge)
    if err != nil {
        panic(err)
    }
    defer db.Close()
    
    fmt.Println("LevelDB 存储初始化成功，使用边缘计算场景配置")
    
    // 方式2：同时使用场景配置和加密
    encryptConfig := &storage.EncryptionConfig{
        Enabled: true,
        Key:     []byte("your-secure-encryption-key-32bytes"),
    }
    
    db, err = storage.NewLevelDBStoreWithScenarioAndEncryption(
        "./edge_db_encrypted", 
        storage.ScenarioEdge, 
        encryptConfig
    )
    if err != nil {
        panic(err)
    }
    defer db.Close()
    
    fmt.Println("LevelDB 存储初始化成功，使用边缘计算场景配置并启用加密")
}
```

### 1.6.6 场景选择建议

- **嵌入式设备**：使用 `ScenarioEmbedded`，内存占用最低，适合资源极度受限的设备
- **IoT 网关设备**：使用 `ScenarioIoT`，适合处理中等规模的时序数据
- **边缘计算节点**：使用 `ScenarioEdge`，平衡性能和资源占用，适合大多数边缘场景 ⭐
- **高性能服务器**：使用 `ScenarioGame`，性能最优，适合需要高吞吐量的场景

## 1.5 使用DBManager管理数据库

sfsDb 提供了 `DBManager` 结构体，用于更结构化、模块化地管理数据库实例。`DBManager` 保持了与原有 `KVDb` 方式的兼容性，同时提供了更清晰的 API 接口。

### 1.5.1 基本使用

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // 获取 DBManager 实例
    dbMgr := storage.GetDBManager()

    // 打开数据库
    dbPath := "./my_custom_db"
    db, err := dbMgr.OpenDB(dbPath)
    if err != nil {
        panic(err)
    }
    defer dbMgr.CloseDB()
    
    fmt.Println("数据库初始化成功")
}
```

### 1.5.2 使用外部存储实例

通过 `DBManager` 也可以设置和使用外部存储实例：

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

// 自定义存储实现
type CustomStore struct {
    // 实现存储逻辑
}

// 实现 Store 接口的方法
func (s *CustomStore) Get(key []byte) ([]byte, error) {
    // 实现获取逻辑
    return nil, nil
}

func (s *CustomStore) Put(key []byte, value []byte) error {
    // 实现存储逻辑
    return nil
}

func (s *CustomStore) Delete(key []byte) error {
    // 实现删除逻辑
    return nil
}

func (s *CustomStore) Batch() storage.Batch {
    // 实现批量操作逻辑
    return nil
}

func (s *CustomStore) Iterator(para ...[]byte) storage.Iterator {
    // 实现迭代器逻辑
    return nil
}

func (s *CustomStore) Snapshot() (storage.Snapshot, error) {
    // 实现快照逻辑
    return nil, nil
}

func (s *CustomStore) Close() error {
    // 实现关闭逻辑
    return nil
}

func main() {
    // 创建自定义存储实例
    customStore := &CustomStore{}
    
    // 获取 DBManager 实例
    dbMgr := storage.GetDBManager()
    
    // 设置外部存储实例
    dbMgr.SetDB(customStore)
    
    fmt.Println("外部存储实例设置成功")
    
    // 获取并使用外部存储实例
    externalStore := dbMgr.GetDB()
    // 现在可以使用 externalStore 进行操作
}
```

### 1.5.3 与原有KVDb方式的兼容性

`DBManager` 保持了与原有 `KVDb` 方式的完全兼容性，您可以在两种方式之间自由切换：

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // 方式1：使用 DBManager
    dbMgr := storage.GetDBManager()
    db, err := dbMgr.OpenDB("./data")
    if err != nil {
        panic(err)
    }
    defer dbMgr.CloseDB()

    // 方式2：使用 DBManager 设置
    // 使用 DBManager 管理存储实例
    dbMgr := storage.GetDBManager()
    dbMgr.SetDB(db)

    fmt.Println("两种方式都可以正常使用")
}
```

### 1.5.4 使用管理工具库

sfsDb 提供了管理工具库，用于监控和管理数据库。通过这个库，您可以获取数据库状态、管理索引、分析性能和执行备份操作。

#### 1.5.4.1 基本使用

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/management"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // 获取 DBManager 实例
    dbMgr := storage.GetDBManager()
    
    // 打开数据库
    store, err := dbMgr.OpenDB("./kvdb")
    if err != nil {
        panic(err)
    }
    defer dbMgr.CloseDB()

    // 创建管理器
    manager := management.NewManager(store)

    // 获取数据库状态
    statusInfo, err := manager.GetStatus()
    if err != nil {
        fmt.Printf("获取状态失败: %v\n", err)
    } else {
        fmt.Println("=== 数据库状态 ===")
        fmt.Printf("内存使用: %.2f MB\n", float64(statusInfo.Memory.Alloc)/1024/1024)
        fmt.Printf("GC 次数: %d\n", statusInfo.Memory.NumGC)
        fmt.Printf("存储类型: %s\n", statusInfo.Storage.StoreType)
    }

    fmt.Println("管理工具库使用成功")
}
```

#### 1.5.4.2 功能模块

管理工具库包含以下功能模块：

##### 1. 状态管理
- **功能**：查询内存使用、GC 次数、存储类型等信息
- **使用示例**：
  ```go
  statusInfo, err := manager.GetStatus()
  if err != nil {
      fmt.Printf("获取状态失败: %v\n", err)
  }
  ```

##### 2. 索引管理
- **功能**：列出索引、分析索引使用情况、提供优化建议
- **使用示例**：
  ```go
  indexMgr := manager.IndexManager()
  indexes, err := indexMgr.ListIndexes("test_table")
  if err != nil {
      fmt.Printf("获取索引失败: %v\n", err)
  }
  ```

##### 3. 性能统计
- **功能**：查询性能统计、识别热点数据、性能分析
- **使用示例**：
  ```go
  statsMgr := manager.StatsManager()
  queryStats, err := statsMgr.GetQueryStats()
  if err != nil {
      fmt.Printf("获取查询统计失败: %v\n", err)
  }
  ```

##### 4. 备份恢复
- **功能**：数据库备份、带选项的备份、备份验证
- **使用示例**：
  ```go
  backupMgr := manager.BackupManager()
  backupFile, err := backupMgr.Backup("./backup")
  if err != nil {
      fmt.Printf("备份失败: %v\n", err)
  } else {
      fmt.Printf("备份成功，备份文件: %s\n", backupFile)
  }
  ```

##### 5. 配置管理
- **功能**：获取配置、设置配置、获取优化建议
- **使用示例**：
  ```go
  configMgr := manager.ConfigManager()
  configInfo, err := configMgr.GetConfig()
  if err != nil {
      fmt.Printf("获取配置失败: %v\n", err)
  } else {
      fmt.Println("=== 配置信息 ===")
      fmt.Printf("存储类型: %s\n", configInfo.StoreType)
      fmt.Println("配置选项:")
      for key, value := range configInfo.Options {
          fmt.Printf("  %s: %s\n", key, value)
      }
  }

  // 获取优化建议
  suggestions, err := configMgr.GetOptimizationSuggestions()
  if err != nil {
      fmt.Printf("获取优化建议失败: %v\n", err)
  } else {
      fmt.Println("=== 优化建议 ===")
      for _, suggestion := range suggestions {
          fmt.Printf("  - %s\n", suggestion)
      }
  }
  ```

##### 6. 监控告警
- **功能**：实时监控、阈值告警、自定义通知器
- **使用示例**：
  ```go
  import (
      "time"
  )

  // 定义监控阈值
  thresholds := management.Thresholds{
      MemoryUsage: 100.0, // 100MB
      GCCount:     10,    // 10次
  }

  // 创建监控器
  monitor := manager.Monitor(5*time.Second, thresholds)

  // 启动监控
  if err := monitor.Start(); err != nil {
      fmt.Printf("启动监控失败: %v\n", err)
  } else {
      fmt.Println("监控已启动")
  }

  // 运行一段时间后停止监控
  time.Sleep(10 * time.Second)
  monitor.Stop()
  ```

##### 7. 深度集成
- **功能**：使用表实例获取更详细的表和索引信息
- **使用示例**：
  ```go
  import (
      "github.com/liaoran123/sfsDb/engine"
  )

  // 创建表
  table, err := engine.TableNew("test_table")
  if err != nil {
      fmt.Printf("创建表失败: %v\n", err)
      return
  }

  // 设置表字段
  fields := map[string]any{
      "id":   0,
      "name": "",
      "age":  0,
  }

  if err := table.SetFields(fields); err != nil {
      fmt.Printf("设置字段失败: %v\n", err)
      return
  }

  // 创建带表实例的管理器
  manager := management.NewManagerWithTable(store, table)

  // 使用深度集成功能
  indexMgr := manager.IndexManager()
  indexes, err := indexMgr.ListIndexes("test_table")
  if err != nil {
      fmt.Printf("获取索引失败: %v\n", err)
  } else {
      fmt.Println("=== 深度集成 - 索引列表 ===")
      for _, idx := range indexes {
          fmt.Printf("索引名称: %s, 类型: %s, 字段: %v\n", idx.Name, idx.Type, idx.Fields)
      }
  }
  ```

#### 1.5.4.3 注意事项

- **导入路径**：管理工具库位于 `github.com/liaoran123/sfsDb/management` 包
- **依赖关系**：需要先打开数据库，获取存储实例，才能创建管理器
- **资源管理**：使用完毕后，需要调用 `dbMgr.CloseDB()` 关闭数据库
- **性能影响**：部分管理操作可能会影响数据库性能，建议在适当的时机执行