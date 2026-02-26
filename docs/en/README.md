# sfsDb User Guide

## Overview

sfsDb is a lightweight, high-performance embedded database library for Go. This documentation provides comprehensive guides for using sfsDb effectively, including basic operations, advanced features, and optimization techniques. Based on a **service-oriented business model**, we focus on deepening our presence in the domestic market by providing professional services and value-added content to realize commercial value.

## Service System

### Service Tiers
- **Community Service**: Basic support through GitHub Issues, WeChat groups, etc. (Free)
- **Professional Service**: Technical support, problem troubleshooting, performance optimization, etc. (Pay-per-use or monthly subscription)
- **Enterprise Service**: Custom development, on-site deployment, training and certification, etc. (Project-based pricing)

### Annual Support Service Packages

- **Basic Operation Package**: $700/year
  - **Applicable to**: Individual developers, startup projects, non-critical business test environments
  - **Service Content**: Remote self-diagnostic tools + weekday remote assistance, 5 lightweight technical consultations (≤30 minutes each), 24-hour fault response (email/ticket), basic documentation and knowledge base access
  - **Node Limit**: Supports 1-5 edge nodes
  - **Value Point**: Low-cost entry, solving basic adaptation and configuration issues

- **Advanced Protection Package**: $2,100/year
  - **Applicable to**: Medium-sized enterprises, production environments, high availability requirement scenarios
  - **Service Content**: 7×24-hour remote monitoring and assistance, unlimited technical consultations (including complex problem troubleshooting), 4-hour emergency response (P1-level faults), monthly remote health check (including database fragmentation, lock status analysis), 2 on-site supports (travel expenses reimbursed, ≤2 person-days each)
  - **Node Limit**: Supports 1-50 edge nodes
  - **Value Point**: Production environment guarantee, preventing edge data corruption risks through health checks

- **Enterprise Exclusive Package**: $7,100/year起
  - **Applicable to**: Large enterprises, critical business systems, multi-node distributed edge clusters
  - **Service Content**: 7×24-hour dedicated engineer on-site (remote), unlimited on-site support (travel expenses reimbursed), customized remote management solutions (including data synchronization strategy, offline transmission logic design), SLA agreement (99.9% availability commitment), priority access to new version testing
  - **Node Limit**: Supports 50+ edge nodes
  - **Value Point**: Expert-level personal service, solving complex edge-cloud data collaboration problems

### Special Services and Add-ons

- **On-site Support Travel Expenses**: Reimbursed at actual cost, standards as follows: Transportation (high-speed rail second class/airplane economy class), Accommodation (first-tier cities ≤$70/day, second-tier cities ≤$60/day), Meal allowance ($14/day)
- **Emergency Fault Expedited Fee**: Non-working hours (night 22:00-8:00 next day) emergency response, 50% service fee surcharge
- **Hardware Adaptation Fee**: Driver development for non-standard edge devices (such as special industrial gateways), quoted separately at $210/person-day
- **Health Check**: $700/time, providing system health check report and optimization suggestions
- **Disaster Recovery Plan**: Starting from $1,400, customized according to customer scale
- **Performance Optimization**: $1,100/time, performance tuning for specific scenarios
- **Custom Development**: Starting from $2,800/person-month, high-difficulty tasks (such as kernel modification, cross-platform transplantation) $3,600/person-month or more

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
- [Jump Ranges](advanced/jump_ranges.md) - Skipping specific data ranges during iteration
- [Management Tool Library](advanced/management.md) - Database monitoring, configuration management, backup/restore, and performance analysis
- [Time Series Processing](advanced/time_series.md) - Time granularity handling, time window calculations, data aggregation, and timestamp conversion
- [Service Support](advanced/service_support.md) - Service levels, support packages and service processes
- [Range Search](advanced/search_range.md) - Efficient range queries with SearchRange method

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
    defer iter.Release()
    
    records := iter.GetRecords(true)
    defer records.Release()
    
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
- **Time Series Processing**: Built-in time package with support for time granularity handling, time window calculations, data aggregation, and timestamp conversion

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

## Performance Reports

- [Time Series Database Performance Comparison](../performance/time_series_benchmark.md) - Benchmark tests for time package and performance comparison with other time series databases

## Support & Contribution

For bug reports, feature requests, or contributions, please visit the [GitHub repository](https://github.com/liaoran123/sfsDb).

## Contact Us

If you have any questions or suggestions, please contact us through:

- **Email**: sfsweb@qq.com

## License

sfsDb adopts a dual-license model:

### Core Engine - MIT License
- **Open Source Free**: Core database engine uses MIT open source license
- **Business Friendly**: Allows free use, modification, and commercial distribution
- **Concise and Flexible**: License text is concise, with few restrictions, easy to understand and use

### Enterprise Edition Plugins - Commercial License
- **Advanced Features**: Enterprise edition plugins (such as advanced security features, multi-node coordination, etc.) use commercial license
- **Support Services**: Includes professional technical support, regular updates, and enterprise-level features

## Privacy Policy

### Data Collection and Usage

1. **Data Collection**:
   - As an embedded database, all data is stored locally in the user's environment. sfsDb does not automatically collect or transmit any user data to external servers.
   - When users choose to use our service support, they may need to provide necessary contact information and problem descriptions, which are only used to solve user issues.

2. **Data Usage**:
   - Locally stored data is completely controlled by the user, and sfsDb does not access or use this data.
   - Information collected during service support is only used to provide technical support and improve service quality.

3. **Data Protection**:
   - sfsDb provides encryption storage functionality, and users can choose to encrypt sensitive data.
   - We take strict security measures to protect information collected during service support.

### User Rights

- **Data Access**: Users have the right to access all data stored locally.
- **Data Modification**: Users have the right to modify or delete locally stored data.
- **Data Export**: Users can export stored data at any time.

## GDPR Compliance Statement

### Scope of Application

This GDPR compliance statement applies to users using sfsDb in the European Union region.

### Data Processing Principles

1. **Lawfulness, Fairness, and Transparency**: sfsDb processes data only with the user's explicit consent, and the processing process is transparent.

2. **Data Minimization**: sfsDb only processes necessary data, and locally stored data is completely controlled by the user.

3. **Purpose Limitation**: sfsDb processes data only for purposes explicitly authorized by the user.

4. **Accuracy**: sfsDb ensures that users can update and correct their data at any time.

5. **Storage Limitation**: Users can control the storage period of data, and sfsDb does not permanently store user data.

6. **Integrity and Confidentiality**: sfsDb takes appropriate technical and organizational measures to protect user data.

### User Rights (GDPR)

Under GDPR, EU users have the following rights:

- **Right to be Informed**: The right to understand how sfsDb processes their data.
- **Right of Access**: The right to obtain their personal data stored by sfsDb.
- **Right to be Forgotten**: The right to request deletion of their personal data.
- **Right to Data Portability**: The right to receive their personal data in a structured, commonly used format.
- **Right to Restrict Processing**: The right to restrict sfsDb's processing of their personal data.
- **Right to Object**: The right to object to sfsDb's processing of their personal data.
- **Rights Related to Automated Decision-Making and Profiling**: The right not to be subject to decisions based solely on automated processing.

### Data Protection Contact

If you have any questions about data protection, please contact us through:

- Email: sfsweb@qq.com
