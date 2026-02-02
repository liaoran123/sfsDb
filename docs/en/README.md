# sfsDb Database User Guide (English)

## Project Introduction

sfsDb is a flexible, efficient embedded database that supports multiple data types, index types, and query methods. It provides a clean API design that is easy to integrate into various application scenarios.

## Table of Contents

### Basic Features
- [Database Initialization](./basic/initialization.md)
- [Creating Tables and Setting Fields](./basic/table_creation.md)
- [Inserting Data](./basic/data_insertion.md)
- [Querying Data](./basic/data_query.md)
- [Deleting Records](./basic/data_deletion.md)

### Advanced Features
- [Primary Key Management](./advanced/primary_key.md)
- [Index Management](./advanced/index_management.md)
- [Full-Text Search](./advanced/full_text_search.md)
- [Field Modification](./advanced/field_modification.md)
- [Transaction Management](./advanced/transaction.md)
- [Other Features](./advanced/other_features.md)

### Optimization and Best Practices
- [Object Pool and Memory Management](./optimization/object_pool.md)
- [Semi-structured Data Support](./optimization/semi_structured.md)
- [Best Practices](./optimization/best_practices.md)
- [Common Issues](./optimization/common_issues.md)
- [Summary](./optimization/summary.md)

## Key Features

1. **Flexible Data Model**: Supports multiple field types and dynamic fields
2. **Powerful Index System**: Supports single primary key, composite primary key, normal index, and full-text index
3. **Rich Query Functions**: Supports comparison operators, custom matchers, and full-text search
4. **Easy-to-Use API**: Clean API design, easy to integrate into various application scenarios
5. **Efficient Performance**: Optimized storage structure and query algorithms
6. **Transaction Support**: Ensures atomicity and consistency of data operations
7. **Object Pool Mechanism**: Optimizes memory usage and performance
8. **Semi-structured Data Support**: Flexible handling of complex data structures

## Quick Start

### 1. Installation

```bash
go get github.com/liaoran123/sfsDb
```

### 2. Basic Usage

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
)

func main() {
    // Create table
    table, err := engine.TableNew("users")
    if err != nil {
        panic(err)
    }
    
    // Set fields
    fields := map[string]any{
        "id":   0,     // Auto-increment primary key
        "name": "",    // String type
        "age":  0,     // Integer type
    }
    err = table.SetFields(fields)
    if err != nil {
        panic(err)
    }
    
    // Insert data
    user := map[string]any{
        "name": "Alice",
        "age":  30,
    }
    id, err := table.Insert(&user)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Insert successful, ID: %v\n", id)
    
    // Query data
    searchData := map[string]any{"name": "Alice"}
    iter := table.Search(&searchData)
    if iter != nil {
        defer engine.GlobalTableIterPool.Put(iter)
    }
    
    records := iter.GetRecords(true)
    if records != nil {
        defer engine.PutRecords(records)
    }
    
    for _, record := range records {
        fmt.Printf("Record: %v\n", record)
    }
}
```

## Documentation Maintenance

This documentation is organized in a modular structure for easy maintenance and updates. For any questions or suggestions, please refer to the [Documentation Maintenance Guide](../DOCUMENTATION_MAINTENANCE.md).

## Contribution Guide

Contributions to code and documentation are welcome! Please refer to [CONTRIBUTING.md](https://github.com/liaoran123/sfsDb/blob/main/CONTRIBUTING.md) to learn how to participate in the project.

## License

sfsDb is licensed under the MIT License. See the [LICENSE](https://github.com/liaoran123/sfsDb/blob/main/LICENSE) file for details.