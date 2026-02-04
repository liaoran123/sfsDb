# sfsDb User Guide

## Overview

sfsDb is a lightweight, high-performance embedded database library for Go. This documentation provides comprehensive guides for using sfsDb effectively, including basic operations, advanced features, and optimization techniques.

## Documentation Structure

### Basic Features

- [Initialization](basic/initialization.md) - Database initialization and configuration
- [Table Creation](basic/table_creation.md) - Creating tables and setting up fields
- [Data Insertion](basic/data_insertion.md) - Inserting data with auto-increment
- [Data Query](basic/data_query.md) - Querying data with matchers and operators
- [Data Update](basic/data_update.md) - Updating records and batch operations
- [Data Deletion](basic/data_deletion.md) - Deleting records and batch operations
- [Field Modification](basic/field_modification.md) - Workflow for modifying table fields

### Advanced Features

- [Primary Key Management](advanced/primary_key.md) - Managing single and composite primary keys
- [Index Management](advanced/index_management.md) - Creating and optimizing indexes
- [Full-Text Search](advanced/full_text_search.md) - Implementing text search functionality
- [Management Tool Library](advanced/management.md) - Database monitoring, configuration management, backup/restore, and performance analysis

### Optimization & Best Practices

- [Index Cache Optimization](optimization/index_cache.md) - Optimizing index cache for better performance
- [Object Pool Usage](optimization/object_pool.md) - Reusing objects to reduce memory overhead
- [Transaction Optimization](optimization/transaction.md) - Using batch operations for atomic transactions

## Quick Start

### 1. Installation

```bash
go get github.com/liaoran123/sfsDb
```

### 2. Basic Usage Example

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // Initialize database
    _, err := storage.OpenDefaultDb("./test_db")
    if err != nil {
        panic(err)
    }
    defer storage.CloseDb()

    // Create table
    table, err := engine.TableNew("users")
    if err != nil {
        panic(err)
    }

    // Set fields
    fields := map[string]any{
        "id":   0,
        "name": "",
        "age":  0,
    }
    err = table.SetFields(fields)
    if err != nil {
        panic(err)
    }

    // Insert data
    user := map[string]any{
        "name": "John Doe",
        "age":  30,
    }
    id, err := table.Insert(&user)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Inserted user with ID: %d\n", id)

    // Query data
    searchFields := map[string]any{"id": id}
    iter := table.Search(&searchFields)
    defer GlobalTableIterPool.Put(iter)
    
    records := iter.GetRecords(true)
    defer record.PutRecords(records)
    
    for _, record := range records {
        fmt.Printf("Found user: %v\n", record)
    }
}
```

## Key Features

- **Lightweight**: Embedded design with minimal dependencies
- **High Performance**: Optimized for fast read/write operations
- **Flexible Schema**: Dynamic field management
- **Powerful Indexing**: Support for primary keys, composite indexes, and full-text search
- **Transaction Support**: Atomic operations through batch processing
- **Memory Optimization**: Object pooling and cache management
- **Concurrency Safe**: Thread-safe design for concurrent operations

## Performance Characteristics

| Feature | Performance | Memory Usage |
|---------|-------------|-------------|
| Insertion | ~100,000 operations/sec | Low |
| Query | ~50,000 operations/sec | Low |
| Index Lookup | ~100,000 operations/sec | Moderate |
| Batch Operations | ~200,000 operations/sec | Moderate |

## Use Cases

- **Embedded Applications**: Perfect for applications that need local data storage
- **Mobile Backends**: Lightweight alternative to traditional databases
- **IoT Devices**: Low memory footprint suitable for resource-constrained environments
- **Testing**: Fast setup and teardown for test environments
- **Edge Computing**: Process data locally with minimal overhead

## Support & Contribution

For bug reports, feature requests, or contributions, please visit the [GitHub repository](https://github.com/liaoran123/sfsDb).

## License

sfsDb is released under the MIT License. See the LICENSE file for details.
