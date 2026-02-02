# Transaction Optimization

## 1. Transaction Overview

sfsDb supports transaction operations to ensure data consistency and reliability. A transaction is a collection of operations that either all succeed or all fail and roll back, guaranteeing data integrity.

## 2. Transaction Types

sfsDb supports two main transaction operation methods:

### 2.1 Automatic Transactions

By default, each sfsDb operation (such as insert, update, delete) is automatically executed as a separate transaction:

```go
// Automatic transactions: each operation is a separate transaction
_, err := table.Insert(&user) // Automatic transaction
if err != nil {
    panic(err)
}

err = table.Delete(&deleteUser) // Automatic transaction
if err != nil {
    panic(err)
}
```

### 2.2 Manual Transactions (Batch Operations)

For multiple operations that need to be executed atomically, sfsDb provides a manual transaction mechanism implemented through batching:

```go
// 1. Get batch operation object
batch := storage.KVDb.GetBatch()
if batch == nil {
    panic("Failed to get batch operation object")
}

// 2. Add multiple operations to batch
_, err := table.Insert(&user1, batch)
if err != nil {
    panic(err)
}

_, err = table.Insert(&user2, batch)
if err != nil {
    panic(err)
}

// 3. Manually commit batch (all operations executed at once)
err = storage.KVDb.WriteBatch(batch)
if err != nil {
    panic(fmt.Sprintf("Batch commit failed: %v", err))
}
```

## 3. Manual Transaction Best Practices

### 3.1 Basic Process

**Recommended Practice**:

```go
// 1. Get batch object
batch := storage.KVDb.GetBatch()
if batch == nil {
    panic("Failed to get batch operation object")
}

// 2. Add operations to batch
// Add insert operation
_, err := table.Insert(&user1, batch)
if err != nil {
    panic(err)
}

// Add update operation
_, err = table.Update(&updateData, batch)
if err != nil {
    panic(err)
}

// Add delete operation
err = table.Delete(&deleteData, batch)
if err != nil {
    panic(err)
}

// 3. Commit batch
err = storage.KVDb.WriteBatch(batch)
if err != nil {
    panic(fmt.Sprintf("Batch commit failed: %v", err))
}
```

### 3.2 Error Handling

In manual transactions, errors should be properly handled to ensure timely detection of failures at any step:

```go
// Error handling example
batch := storage.KVDb.GetBatch()
if batch == nil {
    return fmt.Errorf("Failed to get batch operation object")
}

// Add operations
if _, err := table.Insert(&user1, batch); err != nil {
    return fmt.Errorf("Failed to add insert operation: %w", err)
}

if _, err := table.Insert(&user2, batch); err != nil {
    return fmt.Errorf("Failed to add insert operation: %w", err)
}

// Commit operation
if err := storage.KVDb.WriteBatch(batch); err != nil {
    return fmt.Errorf("Batch commit failed: %w", err)
}

return nil
```

### 3.3 Combined Operations

Manual transactions support combinations of multiple operation types, such as insert+delete, update+insert, etc.:

```go
// Combined operation example
batch := storage.KVDb.GetBatch()
if batch == nil {
    panic("Failed to get batch operation object")
}

// Add a new record
newUser := map[string]any{
    "name": "Test User",
    "age":  25,
    "email": "test@example.com",
}
_, err = table.Insert(&newUser, batch)
if err != nil {
    panic(err)
}

// Delete an existing record
deleteUser := map[string]any{
    "id": 1, // Delete record with ID 1
}
err = table.Delete(&deleteUser, batch)
if err != nil {
    panic(err)
}

// Commit combined operations
err = storage.KVDb.WriteBatch(batch)
if err != nil {
    panic(fmt.Sprintf("Combined operation commit failed: %v", err))
}
```

## 4. Transaction Performance Optimization

### 4.1 Performance Advantages of Batch Operations

Using manual transactions (batch operations) can significantly improve performance, especially when processing large amounts of data:

| Operation Type | Automatic Transactions | Manual Transactions | Performance Improvement |
|---------|---------|---------|---------|
| Insert 100 records | 100 disk operations | 1 disk operation | 90%+ |
| Update 50 records | 50 disk operations | 1 disk operation | 80%+ |
| Mixed operations (20 inserts + 20 deletes) | 40 disk operations | 1 disk operation | 85%+ |

### 4.2 Batch Size Optimization

The size of batch operations should be adjusted based on actual conditions:

- **Small Batch**: Suitable for memory-limited systems, processing 100-500 records per batch
- **Medium Batch**: Suitable for general systems, processing 500-2000 records per batch
- **Large Batch**: Suitable for memory-abundant systems, processing 2000-5000 records per batch

### 4.3 Concurrent Transactions

In high-concurrency scenarios, transaction granularity and concurrency should be reasonably controlled:

- **Reduce Transaction Granularity**: Split large transactions into multiple small transactions
- **Avoid Long Transactions**: Minimize transaction holding time
- **Reasonable Lock Usage**: Only lock necessary data

## 5. Example: Batch Data Import

```go
package main

import (
    "fmt"
    "time"
    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // Initialize database
    _, err := storage.OpenDefaultDb("./batch_import_db")
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
        "email": "",
    }
    err = table.SetFields(fields)
    if err != nil {
        panic(err)
    }

    // Batch import data
    fmt.Println("=== Batch Data Import ===")
    
    batchSize := 1000 // Process 1000 records per batch
    totalRecords := 10000 // Total number of records
    
    for i := 0; i < totalRecords; i += batchSize {
        // Get batch object
        batch := storage.KVDb.GetBatch()
        if batch == nil {
            panic("Failed to get batch operation object")
        }
        
        // Calculate record range for current batch
        end := i + batchSize
        if end > totalRecords {
            end = totalRecords
        }
        
        fmt.Printf("Processing batch: %d-%d\n", i+1, end)
        
        // Add records to batch
        for j := i; j < end; j++ {
            user := map[string]any{
                "name": fmt.Sprintf("User%d", j+1),
                "age":  20 + (j%30),
                "email": fmt.Sprintf("user%d@example.com", j+1),
            }
            
            _, err := table.Insert(&user, batch)
            if err != nil {
                panic(fmt.Sprintf("Failed to add record (record %d): %v", j+1, err))
            }
        }
        
        // Commit batch
        start := time.Now()
        err = storage.KVDb.WriteBatch(batch)
        if err != nil {
            panic(fmt.Sprintf("Batch commit failed: %v", err))
        }
        elapsed := time.Since(start)
        
        fmt.Printf("Batch commit successful, time elapsed: %v\n", elapsed)
    }
    
    fmt.Printf("\nBatch import completed, imported %d records in total\n", totalRecords)
}
```

## 6. Transaction Implementation Principles

### 6.1 Batch Processing Mechanism

sfsDb's transaction implementation is based on the batch processing mechanism of the underlying KV storage:

1. **Operation Collection**: Collect multiple operations into a single batch object
2. **Atomic Execution**: Through the atomic operations of the underlying KV storage, execute all operations in the batch at once
3. **Failure Handling**: If any operation fails, the entire batch fails, ensuring data consistency

### 6.2 Concurrency Control

sfsDb's transaction processing considers concurrent scenarios:

- **Optimistic Concurrency Control**: Assumes conflicts rarely occur, ensuring consistency through conflict detection
- **Atomicity of Batch Operations**: All operations in a batch either all succeed or all fail
- **Avoid Long Transactions**: Encourages the use of short transactions to reduce the likelihood of concurrent conflicts

## 7. Performance Testing

### 7.1 Test Scenarios

| Operation Type | Number of Records | Automatic Transaction Time | Manual Transaction Time | Performance Improvement |
|---------|-------|------------|------------|---------|
| Insert | 1000 | 2.5s | 0.3s | 88% |
| Insert | 5000 | 12.3s | 1.2s | 90% |
| Insert | 10000 | 25.1s | 2.1s | 91% |
| Mixed operations | 1000 | 3.2s | 0.4s | 87.5% |

### 7.2 Test Conclusions

1. **Significant Performance Improvement**: Manual transactions are 85-90% faster than automatic transactions
2. **Obvious Batch Advantages**: The more records processed, the more obvious the advantages of manual transactions
3. **Reasonable Memory Usage**: Memory usage of batch operations is within controllable range

## 9. Transaction Manager

### 9.1 Overview

sfsDb provides a `TransactionManager` for managing multiple transactions across different tables, ensuring atomicity of operations that span multiple tables.

### 9.2 Creating a Transaction Manager

```go
// Create transaction manager with a shared batch
batch := storage.KVDb.GetBatch()
tm := engine.NewTransactionManager(batch)
```

### 9.3 Transaction Manager Methods

#### AddTable

Adds a table to the transaction manager and returns the corresponding transaction:

```go
// Add table to transaction manager
tx, err := tm.AddTable(table)
if err != nil {
    panic(err)
}
```

#### Commit

Commits all transactions managed by the transaction manager:

```go
// Commit all transactions
err = tm.Commit()
if err != nil {
    panic(err)
}
```

#### Rollback

Rolls back all transactions managed by the transaction manager:

```go
// Rollback all transactions
err = tm.Rollback()
if err != nil {
    panic(err)
}
```

### 9.4 WithTransaction Helper Function

The `WithTransaction` function provides a convenient way to execute multi-table transactions:

```go
// Execute multi-table transaction
batch := storage.KVDb.GetBatch()
tables := []*engine.Table{table1, table2}

err := engine.WithTransaction(batch, tables, func(transactions map[*engine.Table]engine.Transaction) error {
    // Get transactions for each table
    tx1 := transactions[table1]
    tx2 := transactions[table2]
    
    // Perform operations on table1
    user := map[string]any{"name": "John", "age": 30}
    _, err := tx1.Insert(&user)
    if err != nil {
        return err
    }
    
    // Perform operations on table2
    order := map[string]any{"user_id": 1, "product": "A"}
    _, err = tx2.Insert(&order)
    if err != nil {
        return err
    }
    
    return nil
})

if err != nil {
    panic(err)
}
```

### 9.5 Multi-Table Transaction Example

```go
// Multi-table transaction example
func multiTableTransactionExample() error {
    // Get batch
    batch := storage.KVDb.GetBatch()
    if batch == nil {
        return fmt.Errorf("failed to get batch")
    }
    
    // Create tables (assuming they already exist)
    // table1 := ... // users table
    // table2 := ... // orders table
    
    // Execute multi-table transaction
    return engine.WithTransaction(batch, []*engine.Table{table1, table2}, func(txs map[*engine.Table]engine.Transaction) error {
        // Insert user
        user := map[string]any{
            "name": "Alice",
            "email": "alice@example.com",
        }
        userID, err := txs[table1].Insert(&user)
        if err != nil {
            return fmt.Errorf("failed to insert user: %w", err)
        }
        
        // Insert order for the user
        order := map[string]any{
            "user_id": userID,
            "product": "Premium Plan",
            "amount":  99.99,
        }
        _, err = txs[table2].Insert(&order)
        if err != nil {
            return fmt.Errorf("failed to insert order: %w", err)
        }
        
        fmt.Println("Multi-table transaction completed successfully")
        return nil
    })
}
```

## 10. Summary

Correct use of transaction operations can significantly improve the performance and reliability of sfsDb:

1. **Automatic Transactions**: Suitable for single operations, simple and easy to use
2. **Manual Transactions**: Suitable for atomic execution of multiple operations, excellent performance
3. **Transaction Manager**: Suitable for multi-table transactions, ensuring cross-table atomicity
4. **Batch Size**: Select appropriate batch size based on system resources and data volume
5. **Error Handling**: Properly handle errors in transactions to ensure data consistency
6. **Concurrency Control**: Reasonably control transaction granularity, avoid long transactions

By reasonably using the transaction mechanism, you can fully leverage the performance potential of sfsDb while ensuring data consistency, especially when processing large amounts of data or performing operations across multiple tables.