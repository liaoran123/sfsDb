# Management Tool Library

The management tool library is one of the core components of sfsDb, providing comprehensive database management, monitoring, and optimization functions. This document will detail the usage methods and functional features of the management tool library.

## Overview

The management tool library is located in the `management` package and includes the following core modules:

- **System Information Management**: Get system information such as tables, fields, and indexes
- **Index Management**: Analyze index usage and provide optimization suggestions
- **Performance Statistics**: Record query performance and identify hotspots
- **Configuration Management**: Manage database configuration and provide optimization suggestions
- **Backup and Restore**: Backup and restore the database without shutting it down
- **Monitoring and Alerting**: Real-time monitoring of database status with threshold-based alerts
- **Status Management**: Get database memory usage and storage status

## Core Features

### 1. Deep Integration

The management tool library supports deep integration with `engine.Table` instances, allowing for more detailed table and index information.

```go
// Create a manager with table instance
manager := management.NewManagerWithTable(store, table)

// Use deep integration to get index information
indexMgr := manager.IndexManager()
indexes, err := indexMgr.ListIndexes("test_table")
```

### 2. Monitoring and Alerting System

The monitoring and alerting system can real-time monitor database status and trigger alerts when thresholds are reached.

```go
// Create monitor
monitor := manager.Monitor(time.Second*5, management.Thresholds{
    MemoryUsage: 1024, // 1GB
    GCCount:     100,
})

// Start monitoring
if err := monitor.Start(); err != nil {
    fmt.Printf("Failed to start monitoring: %v\n", err)
}

// Stop monitoring
monitor.Stop()
```

### 3. Configuration Management

The configuration management module allows for runtime configuration adjustments and provides optimization suggestions.

```go
// Get configuration manager
configMgr := manager.ConfigManager()

// Get current configuration
config, err := configMgr.GetConfig()

// Set configuration
if err := configMgr.SetConfig("write_buffer", "128MB"); err != nil {
    fmt.Printf("Failed to set configuration: %v\n", err)
}

// Get optimization suggestions
suggestions, err := configMgr.GetOptimizationSuggestions()
```

### 4. Backup and Restore

The backup and restore module supports backup and restore operations without shutting down the database.

```go
// Get backup manager
backupMgr := manager.BackupManager()

// Create backup
backupPath, err := backupMgr.Backup("./backups")

// Restore database
if err := backupMgr.Restore(backupPath); err != nil {
    fmt.Printf("Failed to restore database: %v\n", err)
}
```

### 5. Performance Statistics

The performance statistics module records query performance and identifies hotspot data.

```go
// Get performance statistics manager
statsMgr := manager.StatsManager()

// Record query performance
statsMgr.RecordQuery(time.Millisecond*10, "select")

// Record data access
statsMgr.RecordAccess("key_1")

// Get query statistics
queryStats, err := statsMgr.GetQueryStats()

// Get hotspots
hotspots, err := statsMgr.GetHotspots(5)

// Start profiling
if err := statsMgr.StartProfiling(); err != nil {
    fmt.Printf("Failed to start profiling: %v\n", err)
}

// Stop profiling and get results
profilingResult, err := statsMgr.StopProfiling()
```

### 6. System Information Management

The system information management module provides system information such as tables, fields, and indexes.

```go
// Get system information manager
systemMgr := manager.SystemManager()

// Get all tables
tables, err := systemMgr.GetAllTables()

// Get table fields
fields, err := systemMgr.GetTableFields(tableID)

// Get table indexes
indexes, err := systemMgr.GetTableIndexes(tableID)

// Get all system information
systemInfo, err := systemMgr.GetAllSystemInfo()
```

### 7. Status Management

The status management module provides database memory usage and storage status information.

```go
// Get database status
statusInfo, err := manager.GetStatus()
fmt.Printf("Memory usage: %.2f MB\n", float64(statusInfo.Memory.Alloc)/1024/1024)
fmt.Printf("GC count: %d\n", statusInfo.Memory.NumGC)
fmt.Printf("Storage type: %s\n", statusInfo.Storage.StoreType)
```

### 8. Web Interface Management

The Web interface management module allows users to enable/disable the web interface and configure its port. Web interface settings can be easily managed through CLI commands.

#### Enable Web Interface and Configure Port

```bash
# Enable web interface and set port to 8083
sfsdb web --enable --port :8083

# Enable web interface and set port to 8084
sfsdb web --enable --port :8084
```

#### Disable Web Interface

```bash
# Disable web interface
sfsdb web --disable
```

#### Start Web Server

```bash
# Start web server (using configured port)
sfsdb web --start

# Start web server and specify port (temporarily override configuration)
sfsdb web --start --port :8085
```

#### View Current Web Configuration

```bash
# View current web configuration
sfsdb web
```

Output example:

```
Web Interface Configuration:
Enabled: true
Port: :8084
Address: http://localhost:8084
```

#### Web Interface Features

The Web interface provides the following features:

- **System Information**: View database and system status
- **Configuration Management**: View and modify database configuration
- **Backup and Restore**: Create and restore database backups
- **Performance Statistics**: View query performance and hotspot data
- **Monitoring and Alerting**: Real-time monitoring of database status
- **Data Operations**: Basic data CRUD operations

## Usage Examples

### Complete Example

Here is a complete example of using the management tool library:

```go
package main

import (
    "fmt"
    "time"

    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/management"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // Initialize database
    _, err := storage.OpenDefaultDb("./kvdb")
    if err != nil {
        fmt.Printf("Failed to open database: %v\n", err)
        return
    }
    defer storage.CloseDb()

    // Create test table
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

    // Create indexes
    primaryKey, err := engine.DefaultPrimaryKeyNew("primary_key")
    if err != nil {
        fmt.Printf("Failed to create primary key: %v\n", err)
        return
    }
    primaryKey.AddFields("id")
    if err := table.CreateIndex(primaryKey); err != nil {
        fmt.Printf("Failed to add primary key: %v\n", err)
        return
    }

    normalIndex, err := engine.DefaultNormalIndexNew("name_index")
    if err != nil {
        fmt.Printf("Failed to create normal index: %v\n", err)
        return
    }
    normalIndex.AddFields("name")
    if err := table.CreateIndex(normalIndex); err != nil {
        fmt.Printf("Failed to add normal index: %v\n", err)
        return
    }

    // Create manager
    manager := management.NewManagerWithTable(storage.KVDb, table)

    // Test system information manager
    fmt.Println("=== Testing System Information Manager ===")
    systemMgr := manager.SystemManager()
    tables, err := systemMgr.GetAllTables()
    if err != nil {
        fmt.Printf("Failed to get table information: %v\n", err)
        return
    }
    fmt.Printf("All tables: %+v\n", tables)

    // Test index manager
    fmt.Println("\n=== Testing Index Manager ===")
    indexMgr := manager.IndexManager()
    indexes, err := indexMgr.ListIndexes("test_table")
    if err != nil {
        fmt.Printf("Failed to list indexes: %v\n", err)
        return
    }
    fmt.Printf("All indexes: %+v\n", indexes)

    // Test performance statistics manager
    fmt.Println("\n=== Testing Performance Statistics Manager ===")
    statsMgr := manager.StatsManager()

    // Simulate queries
    for i := 0; i < 10; i++ {
        statsMgr.RecordQuery(time.Millisecond*10, "select")
        statsMgr.RecordAccess(fmt.Sprintf("key_%d", i))
    }

    queryStats, err := statsMgr.GetQueryStats()
    if err != nil {
        fmt.Printf("Failed to get query statistics: %v\n", err)
        return
    }
    fmt.Printf("Query statistics: %+v\n", queryStats)

    // Test configuration manager
    fmt.Println("\n=== Testing Configuration Manager ===")
    configMgr := manager.ConfigManager()
    config, err := configMgr.GetConfig()
    if err != nil {
        fmt.Printf("Failed to get configuration: %v\n", err)
        return
    }
    fmt.Printf("Current configuration: %+v\n", config)

    // Test backup manager
    fmt.Println("\n=== Testing Backup Manager ===")
    backupMgr := manager.BackupManager()

    // Create backup
    backupPath, err := backupMgr.Backup("./backups")
    if err != nil {
        fmt.Printf("Failed to create backup: %v\n", err)
        return
    }
    fmt.Printf("Backup created successfully: %s\n", backupPath)

    // Test monitor manager
    fmt.Println("\n=== Testing Monitor Manager ===")
    monitorMgr := manager.MonitorManager()
    keyChangeStats := monitorMgr.GetKeyChangeStats()
    fmt.Printf("Key change statistics: %+v\n", keyChangeStats)

    // Test status manager
    fmt.Println("\n=== Testing Status Manager ===")
    statusInfo, err := manager.GetStatus()
    if err != nil {
        fmt.Printf("Failed to get status: %v\n", err)
        return
    }
    fmt.Printf("Database status: %+v\n", statusInfo)

    // Test monitor
    fmt.Println("\n=== Testing Monitor ===")
    monitor := manager.Monitor(time.Second*5, management.Thresholds{
        MemoryUsage: 1024, // 1GB
        GCCount:     100,
    })

    err = monitor.Start()
    if err != nil {
        fmt.Printf("Failed to start monitor: %v\n", err)
        return
    }
    fmt.Println("Monitor started")

    // Wait for a while
    time.Sleep(time.Second * 10)

    // Stop monitor
    monitor.Stop()
    fmt.Println("Monitor stopped")

    fmt.Println("\n=== All management features tested ===")
}
```

## Performance Optimization Suggestions

When using the management tool library, here are some performance optimization suggestions:

1. **Regularly analyze index usage**: Use `IndexManager.AnalyzeIndexes` to regularly analyze index usage and remove unused indexes.

2. **Monitor database status**: Use the monitor to regularly monitor database status and detect performance issues early.

3. **Optimize configuration parameters**: Adjust configuration parameters such as `write_buffer` and `max_open_files` based on your application scenario.

4. **Backup strategy**: Develop a reasonable backup strategy, regularly backup the database to ensure data security.

5. **Hotspot data handling**: Identify hotspot data and consider using caching or other optimization strategies.

## Summary

The management tool library provides comprehensive management, monitoring, and optimization functions for sfsDb, making it an important tool for database operations. By using the management tool library appropriately, you can improve database performance, ensure data security, and simplify database management tasks.
