# Database Initialization

## 1.1 Custom Database Path

By default, sfsDb uses the `kvdb` folder in the current directory as the database storage path. If you need to customize the database path, you can call the `OpenDefaultDb` function when starting the program:

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // Custom database path
    dbPath := "./my_custom_db"
    _, err := storage.OpenDefaultDb(dbPath)
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Database initialization successful")
}
```

## 1.2 Automatic Use of Default Path

If you don't call the `OpenDefaultDb` function, the system will automatically use the default path `./kvdb` when creating the table for the first time:

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
)

func main() {
    // When TableNew is called for the first time, the database will be automatically initialized with the default path "./kvdb"
    table, err := engine.TableNew("users")
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Table created successfully, database automatically initialized")
}
```

## 1.3 Database Closure

When the program ends, you can call the `CloseDb` function to close the database and release resources:

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // Create table
    table, err := engine.TableNew("users")
    if err != nil {
        panic(err)
    }
    
    // Close database when program ends
    defer storage.CloseDb()
    
    fmt.Println("Table created successfully")
}
```

## 1.4 Using External Storage Instances

In addition to using built-in storage engines, sfsDb also supports using externally implemented storage instances. Simply set it through the `SetStore` function:

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

// Custom storage implementation
type CustomStore struct {
    // Implement storage logic
}

// Implement Store interface methods
func (s *CustomStore) Get(key []byte) ([]byte, error) {
    // Implement get logic
    return nil, nil
}

func (s *CustomStore) Put(key []byte, value []byte) error {
    // Implement put logic
    return nil
}

func (s *CustomStore) Delete(key []byte) error {
    // Implement delete logic
    return nil
}

func (s *CustomStore) Batch() storage.Batch {
    // Implement batch operation logic
    return nil
}

func (s *CustomStore) Iterator(para ...[]byte) storage.Iterator {
    // Implement iterator logic
    return nil
}

func (s *CustomStore) Snapshot() (storage.Snapshot, error) {
    // Implement snapshot logic
    return nil, nil
}

func (s *CustomStore) Close() error {
    // Implement close logic
    return nil
}

func main() {
    // Create custom storage instance
    customStore := &CustomStore{}
    
    // Set external storage instance
    storage.SetStore(customStore)
    
    fmt.Println("External storage instance set successfully")
    
    // Now the entire sfsdb project will use this custom storage instance
    // For example, creating tables, inserting data, etc. will all be executed through this instance
}
```

**Usage scenarios**:
- Integrating third-party storage implementations
- Customizing storage logic for specific scenarios
- Using in-memory storage or mock storage in tests
- Implementing special storage features such as encryption, compression, etc.

**Notes**:
- When using external storage instances, you need to manage their lifecycle yourself
- Ensure to properly close the storage instance when it's no longer needed
- The external storage implementation must fully implement all methods of the `Store` interface

## 1.6 Using Scenario Configurations

sfsDb provides predefined scenario configurations to choose appropriate memory and performance settings for different use cases, especially suitable for edge computing and IoT devices.

### 1.6.1 Available Scenarios

| Scenario Constant | Description | Total Memory | Use Case |
|-------------------|-------------|--------------|----------|
| `ScenarioEmbedded` | Embedded device | ~6MB | Smart terminal devices |
| `ScenarioIoT` | IoT device | ~12MB | IoT gateway devices |
| `ScenarioEdge` | Edge computing node | ~48MB | Edge computing nodes ⭐ |
| `ScenarioGame` | Game server | ~192MB | High-performance scenarios |

### 1.6.2 Opening Database with Scenario Configuration

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // Method 1: Using DBManager with scenario
    dbManager := storage.GetDBManager()
    db, err := dbManager.OpenDBWithScenario("./edge_db", storage.ScenarioEdge)
    if err != nil {
        panic(err)
    }
    defer dbManager.CloseDB()
    
    fmt.Println("Database initialized successfully with edge computing scenario configuration")
}
```

### 1.6.3 Using Backward-Compatible Global Function

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // Using global function with scenario
    _, err := storage.OpenDefaultDbWithScenario("./edge_db", storage.ScenarioEdge)
    if err != nil {
        panic(err)
    }
    defer storage.CloseDb()
    
    fmt.Println("Database initialized successfully with edge computing scenario configuration")
}
```

### 1.6.4 Using Scenario Configuration with Encryption

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // Create encryption configuration
    encryptConfig := &storage.EncryptionConfig{
        Enabled: true,
        Key:     []byte("your-secure-encryption-key-32bytes"),
    }
    
    // Method 1: Using DBManager with both scenario and encryption
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
    
    fmt.Println("Database initialized successfully with edge computing scenario configuration and encryption enabled")
}
```

### 1.6.5 Creating LevelDB Store Directly with Scenario Configuration

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // Method 1: Using only scenario configuration
    db, err := storage.NewLevelDBStoreWithScenario("./edge_db", storage.ScenarioEdge)
    if err != nil {
        panic(err)
    }
    defer db.Close()
    
    fmt.Println("LevelDB store initialized successfully with edge computing scenario configuration")
    
    // Method 2: Using both scenario configuration and encryption
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
    
    fmt.Println("LevelDB store initialized successfully with edge computing scenario configuration and encryption enabled")
}
```

### 1.6.6 Scenario Selection Recommendations

- **Embedded devices**: Use `ScenarioEmbedded`, lowest memory footprint, suitable for extremely resource-constrained devices
- **IoT gateway devices**: Use `ScenarioIoT`, suitable for processing medium-scale time-series data
- **Edge computing nodes**: Use `ScenarioEdge`, balance between performance and resource usage, suitable for most edge scenarios ⭐
- **High-performance servers**: Use `ScenarioGame`, optimal performance, suitable for scenarios requiring high throughput

## 1.5 Using DBManager to Manage Database

sfsDb provides a `DBManager` struct for more structured and modular management of database instances. `DBManager` maintains compatibility with the original `KVDb` approach while providing a clearer API interface.

### 1.5.1 Basic Usage

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // Get DBManager instance
    dbMgr := storage.GetDBManager()

    // Open database
    dbPath := "./my_custom_db"
    db, err := dbMgr.OpenDB(dbPath)
    if err != nil {
        panic(err)
    }
    defer dbMgr.CloseDB()
    
    fmt.Println("Database initialization successful")
}
```

### 1.5.2 Using External Storage Instances

You can also set and use external storage instances through `DBManager`:

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

// Custom storage implementation
type CustomStore struct {
    // Implement storage logic
}

// Implement Store interface methods
func (s *CustomStore) Get(key []byte) ([]byte, error) {
    // Implement get logic
    return nil, nil
}

func (s *CustomStore) Put(key []byte, value []byte) error {
    // Implement put logic
    return nil
}

func (s *CustomStore) Delete(key []byte) error {
    // Implement delete logic
    return nil
}

func (s *CustomStore) Batch() storage.Batch {
    // Implement batch operation logic
    return nil
}

func (s *CustomStore) Iterator(para ...[]byte) storage.Iterator {
    // Implement iterator logic
    return nil
}

func (s *CustomStore) Snapshot() (storage.Snapshot, error) {
    // Implement snapshot logic
    return nil, nil
}

func (s *CustomStore) Close() error {
    // Implement close logic
    return nil
}

func main() {
    // Create custom storage instance
    customStore := &CustomStore{}
    
    // Get DBManager instance
    dbMgr := storage.GetDBManager()
    
    // Set external storage instance
    dbMgr.SetDB(customStore)
    
    fmt.Println("External storage instance set successfully")
    
    // Get and use external storage instance
    externalStore := dbMgr.GetDB()
    // Now you can use externalStore for operations
}
```

### 1.5.3 Compatibility with Original KVDb Approach

`DBManager` maintains full compatibility with the original `KVDb` approach, so you can freely switch between the two methods:

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // Method 1: Using DBManager
    dbMgr := storage.GetDBManager()
    db, err := dbMgr.OpenDB("./data")
    if err != nil {
        panic(err)
    }
    defer dbMgr.CloseDB()

    // Method 2: Using DBManager
    // Use DBManager to manage storage instance
    dbMgr := storage.GetDBManager()
    dbMgr.SetDB(db)

    fmt.Println("Both methods can be used normally")
}
```

### 1.5.4 Using Management Tool Library

sfsDb provides a management tool library for monitoring and managing the database. Through this library, you can get database status, manage indexes, analyze performance, and perform backup operations.

#### 1.5.4.1 Basic Usage

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/management"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // Get DBManager instance
    dbMgr := storage.GetDBManager()
    
    // Open database
    store, err := dbMgr.OpenDB("./kvdb")
    if err != nil {
        panic(err)
    }
    defer dbMgr.CloseDB()

    // Create manager
    manager := management.NewManager(store)

    // Get database status
    statusInfo, err := manager.GetStatus()
    if err != nil {
        fmt.Printf("Failed to get status: %v\n", err)
    } else {
        fmt.Println("=== Database Status ===")
        fmt.Printf("Memory usage: %.2f MB\n", float64(statusInfo.Memory.Alloc)/1024/1024)
        fmt.Printf("GC count: %d\n", statusInfo.Memory.NumGC)
        fmt.Printf("Storage type: %s\n", statusInfo.Storage.StoreType)
    }

    fmt.Println("Management tool library used successfully")
}
```

#### 1.5.4.2 Functional Modules

The management tool library includes the following functional modules:

##### 1. Status Management
- **Function**: Query memory usage, GC count, storage type, etc.
- **Usage example**:
  ```go
  statusInfo, err := manager.GetStatus()
  if err != nil {
      fmt.Printf("Failed to get status: %v\n", err)
  }
  ```

##### 2. Index Management
- **Function**: List indexes, analyze index usage, provide optimization suggestions
- **Usage example**:
  ```go
  indexMgr := manager.IndexManager()
  indexes, err := indexMgr.ListIndexes("test_table")
  if err != nil {
      fmt.Printf("Failed to get indexes: %v\n", err)
  }
  ```

##### 3. Performance Statistics
- **Function**: Query performance statistics, identify hotspots, analyze performance
- **Usage example**:
  ```go
  statsMgr := manager.StatsManager()
  queryStats, err := statsMgr.GetQueryStats()
  if err != nil {
      fmt.Printf("Failed to get query stats: %v\n", err)
  }
  ```

##### 4. Backup and Restore
- **Function**: Database backup, backup with options, backup validation
- **Usage example**:
  ```go
  backupMgr := manager.BackupManager()
  backupFile, err := backupMgr.Backup("./backup")
  if err != nil {
      fmt.Printf("Backup failed: %v\n", err)
  } else {
      fmt.Printf("Backup successful, backup file: %s\n", backupFile)
  }
  ```

##### 5. Configuration Management
- **Function**: Get configuration, set configuration, get optimization suggestions
- **Usage example**:
  ```go
  configMgr := manager.ConfigManager()
  configInfo, err := configMgr.GetConfig()
  if err != nil {
      fmt.Printf("Failed to get config: %v\n", err)
  } else {
      fmt.Println("=== Configuration Info ===")
      fmt.Printf("Storage type: %s\n", configInfo.StoreType)
      fmt.Println("Configuration options:")
      for key, value := range configInfo.Options {
          fmt.Printf("  %s: %s\n", key, value)
      }
  }

  // Get optimization suggestions
  suggestions, err := configMgr.GetOptimizationSuggestions()
  if err != nil {
      fmt.Printf("Failed to get optimization suggestions: %v\n", err)
  } else {
      fmt.Println("=== Optimization Suggestions ===")
      for _, suggestion := range suggestions {
          fmt.Printf("  - %s\n", suggestion)
      }
  }
  ```

##### 6. Monitoring and Alerting
- **Function**: Real-time monitoring, threshold alerts, custom notifiers
- **Usage example**:
  ```go
  import (
      "time"
  )

  // Define monitoring thresholds
  thresholds := management.Thresholds{
      MemoryUsage: 100.0, // 100MB
      GCCount:     10,    // 10 times
  }

  // Create monitor
  monitor := manager.Monitor(5*time.Second, thresholds)

  // Start monitoring
  if err := monitor.Start(); err != nil {
      fmt.Printf("Failed to start monitor: %v\n", err)
  } else {
      fmt.Println("Monitor started")
  }

  // Run for a while then stop monitoring
  time.Sleep(10 * time.Second)
  monitor.Stop()
  ```

##### 7. Deep Integration
- **Function**: Use table instance to get more detailed table and index information
- **Usage example**:
  ```go
  import (
      "github.com/liaoran123/sfsDb/engine"
  )

  // Create table
  table, err := engine.TableNew("test_table")
  if err != nil {
      fmt.Printf("Failed to create table: %v\n", err)
      return
  }

  // Set table fields
  fields := map[string]any{
      "id":   0,
      "name": "",
      "age":  0,
  }

  if err := table.SetFields(fields); err != nil {
      fmt.Printf("Failed to set fields: %v\n", err)
      return
  }

  // Create manager with table instance
  manager := management.NewManagerWithTable(store, table)

  // Use deep integration feature
  indexMgr := manager.IndexManager()
  indexes, err := indexMgr.ListIndexes("test_table")
  if err != nil {
      fmt.Printf("Failed to get indexes: %v\n", err)
  } else {
      fmt.Println("=== Deep Integration - Index List ===")
      for _, idx := range indexes {
          fmt.Printf("Index name: %s, Type: %s, Fields: %v\n", idx.Name, idx.Type, idx.Fields)
      }
  }
  ```

#### 1.5.4.2 Notes

- **Import path**: The management tool library is located in the `github.com/liaoran123/sfsDb/management` package
- **Dependency**: You need to open the database and get the storage instance before creating the manager
- **Resource management**: After use, you need to call `dbMgr.CloseDB()` to close the database
- **Performance impact**: Some management operations may affect database performance, so it's recommended to execute them at appropriate times