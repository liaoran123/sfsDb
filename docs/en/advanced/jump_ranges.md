# Jump Ranges

Jump Ranges is an advanced feature in sfsDb that allows skipping specific ranges of data during iteration, improving query efficiency. This document provides a detailed explanation of the implementation principles, usage methods, and application scenarios of Jump Ranges.

## 1. Overview

Jump Ranges allow skipping predefined ranges of data during iteration, particularly suitable for the following scenarios:

- **Excluding specific categories** - Such as excluding electronics products in e-commerce systems
- **Skipping invalid data** - Such as skipping abnormal sensor data in IoT systems
- **Time range exclusion** - Such as skipping transactions in specific time periods in financial systems
- **Any query scenario requiring range exclusion**

## 2. Implementation Principles

### 2.1 Core Data Structure

The core implementation of Jump Ranges is located in the `TableIter` struct:

```go
type TableIter struct {
    iter       storage.Iterator
    jumpRanges []storage.Iterator // Jump ranges
    table      *Table
    match      []match.Match // Matching rules
    selects    []string
    index      Index // Index used for search
    move       map[bool]func() bool
    top        map[bool]func() bool
    mu         sync.Mutex
}
```

### 2.2 Core Methods

#### 2.2.1 Setting Jump Ranges

```go
func (t *TableIter) SetJumpRanges(jumpRanges ...storage.Iterator) {
    t.jumpRanges = jumpRanges
}
```

#### 2.2.2 Jump Range Detection

```go
func (t *TableIter) JumpRange(key []byte, jumpRanges []storage.Iterator, esc bool) []byte {
    if len(jumpRanges) == 0 {
        return nil
    }
    for _, jumpRange := range jumpRanges {
        if esc {
            jumpRange.First()
        } else {
            jumpRange.Last()
        }
        if bytes.Equal(jumpRange.Key(), key) {
            if esc {
                // For forward iteration, check if it's the last element
                if jumpRange.Last() {
                    return jumpRange.Key()
                }
            } else {
                // For reverse iteration, check if it's the first element
                if jumpRange.First() {
                    return jumpRange.Key()
                }
            }
        }
    }
    return nil
}
```

#### 2.2.3 Jumping During Iteration

In the `ExportRecord` method, when iterating through each key-value pair, it checks if there are jump ranges:

```go
func (t *TableIter) ExportRecord(export ExportRecord, esc bool, limit ...int) {
    // Add lock to prevent concurrent access
    t.mu.Lock()
    defer t.mu.Unlock()
    if !t.top[esc]() {
        return
    }
    var rd record.Record
    isMatch := true
    matchlen := 0
    page := PageNew(limit...)
    count := 0
    loop := 0
    var end []byte
    var key, value []byte
    for {
        key = t.iter.Key()
        value = t.iter.Value()
        if len(t.jumpRanges) > 0 { // If there are jump ranges, skip certain data
            end = t.JumpRange(key, t.jumpRanges, esc)
            if end != nil {
                // Jump to the end of the range
                t.iter.Seek(end)
                if !t.move[esc]() {
                    break
                }
            }
        }
        // Subsequent processing...
    }
}
```

## 3. Relationship with RangeForAny

The `RangeForAny` function is the foundation for implementing Jump Ranges, used to create iterators representing ranges that need to be skipped:

```go
func (t *Table) RangeForAny(funIter storage.FunIter, fieldname string, Start, Limit any) (storage.Iterator, Index, error) {
    if funIter == nil {
        funIter = t.kvStore.Iterator
    }
    idx := t.MatchIndexCached([]string{fieldname})
    if idx == nil {
        return nil, nil, fmt.Errorf("field '%s' does not exist in table '%s'", fieldname, t.name)
    }
    pfx := idx.Prefix(t.id)
    pfx = append(pfx, SPLIT[0])
    // Process Start parameter
    var startBytes []byte
    if Start != nil {
        startBytes = util.AnyToBytes(Start)
    }
    // Process Limit parameter
    var limitBytes []byte
    if Limit != nil {
        limitBytes = util.AnyToBytes(Limit)
    }
    // Create range object
    slice := &util.Range{
        Start: startBytes,
        Limit: limitBytes,
    }
    // Build complete search range
    slice.Start = append(pfx, slice.Start...)
    if Limit == nil {
        // When Limit is nil, use the next byte of the prefix as the upper limit, representing infinity
        slice.Limit = util.BytesPrefix(pfx).Limit
    } else {
        // When Limit is not nil, build complete upper limit bytes
        slice.Limit = append(pfx, slice.Limit...)
    }
    iter := funIter(slice.Start, slice.Limit)
    if iter == nil {
        return nil, nil, fmt.Errorf("range iterator cannot be nil")
    }
    return iter, idx, nil
}
```

## 4. Usage Methods

### 4.1 Basic Usage Steps

1. **Create jump range iterators**: Use the `RangeForAny` function to create iterators representing ranges that need to be skipped
2. **Set jump ranges**: Call the `SetJumpRanges` method of `TableIter` to set jump ranges
3. **Execute iteration**: Call `GetRecordSet` or other iteration methods, which will automatically skip data within jump ranges during iteration

### 4.2 Code Examples

#### 4.2.1 E-commerce System Example: Excluding Electronics

```go
// Create table
table, err := engine.TableNew("products")
if err != nil {
    panic(err)
}

// Set fields
fields := map[string]any{
    "id":       0,
    "name":     "",
    "category": "",
    "price":    0.0,
}
err = table.SetFields(fields)
if err != nil {
    panic(err)
}

// Create indexes
pk, err := engine.DefaultPrimaryKeyNew("pk")
pk.AddFields("id")
err = table.CreateIndex(pk)

// Create category index
categoryIndex, err := engine.DefaultNormalIndexNew("idx_category")
categoryIndex.AddFields("category")
err = table.CreateIndex(categoryIndex)

// Insert test data
// ... Insert data code ...

// Create jump range: Electronics range
electronicsIter, _, _ := table.RangeForAny(nil, "category", "electronics", "electronics")

// Create main query iterator: All products
allProductsIter, _, _ := table.SearchRange(nil, "category", nil, nil)

// Set jump ranges
allProductsIter.SetJumpRanges(electronicsIter)

// Execute query
result := allProductsIter.GetRecordSet(true)

// Process results
for _, record := range result {
    fmt.Printf("Product: %s, Category: %s, Price: %.2f\n", 
        record["name"], record["category"], record["price"])
}
```

#### 4.2.2 IoT System Example: Skipping Abnormal Data

```go
// Create table
sensorTable, err := engine.TableNew("sensor_data")
if err != nil {
    panic(err)
}

// Set fields
fields := map[string]any{
    "timestamp": 0,
    "sensor_id": "",
    "value":     0.0,
    "status":    "",
}
err = sensorTable.SetFields(fields)
if err != nil {
    panic(err)
}

// Create indexes
// ... Create index code ...

// Insert test data
// ... Insert data code ...

// Create jump range: Error data range
errorIter, _, _ := sensorTable.RangeForAny(nil, "status", "error", "error")

// Create main query iterator: All sensor data
allDataIter, _, _ := sensorTable.SearchRange(nil, "timestamp", nil, nil)

// Set jump ranges
allDataIter.SetJumpRanges(errorIter)

// Execute query
result := allDataIter.GetRecordSet(true)

// Process results
for _, record := range result {
    fmt.Printf("Timestamp: %d, Sensor: %s, Value: %.2f, Status: %s\n", 
        record["timestamp"], record["sensor_id"], record["value"], record["status"])
}
```

#### 4.2.3 Financial System Example: Skipping Specific Time Period Transactions

```go
// Create table
transactionTable, err := engine.TableNew("transactions")
if err != nil {
    panic(err)
}

// Set fields
fields := map[string]any{
    "id":        0,
    "timestamp": 0,
    "amount":    0.0,
    "type":      "",
}
err = transactionTable.SetFields(fields)
if err != nil {
    panic(err)
}

// Create indexes
// ... Create index code ...

// Insert test data
// ... Insert data code ...

// Create jump range: December 1, 2023 transactions
startTime := 1672531200 // 2023-12-01 00:00:00
endTime := 1672617599   // 2023-12-01 23:59:59
december1Iter, _, _ := transactionTable.RangeForAny(nil, "timestamp", startTime, endTime)

// Create main query iterator: All transactions
allTransactionsIter, _, _ := transactionTable.SearchRange(nil, "timestamp", nil, nil)

// Set jump ranges
allTransactionsIter.SetJumpRanges(december1Iter)

// Execute query
result := allTransactionsIter.GetRecordSet(true)

// Process results
for _, record := range result {
    fmt.Printf("Transaction ID: %d, Timestamp: %d, Amount: %.2f, Type: %s\n", 
        record["id"], record["timestamp"], record["amount"], record["type"])
}
```

## 5. Performance Optimization

### 5.1 Best Practices

1. **Reasonable jump range setting**: Only skip ranges that truly need to be excluded, avoid overusing jump ranges
2. **Use appropriate indexes**: Create indexes for jump range fields to improve range creation speed
3. **Batch operations**: For scenarios requiring multiple jump ranges, set all jump ranges at once
4. **Timely resource release**: Call `GlobalTableIterPool.Put(iter)` to release iterator resources after use

### 5.2 Performance Comparison

Using jump ranges can significantly improve query efficiency, especially in the following scenarios:

- **Large amount of data to exclude**: When excluding more than 80% of data, performance improvement is significant
- **Multiple repeated queries**: For repeatedly executed queries, jump ranges can avoid repeated filtering
- **Complex filtering conditions**: When filtering conditions are complex, jump ranges are more efficient than traditional `Match` conditions

## 6. Notes

1. **Correctness of jump ranges**: Ensure jump range iterators correctly represent ranges that need to be skipped
2. **Index dependency**: Jump ranges depend on field indexes, ensure appropriate indexes are created for jump range fields
3. **Memory usage**: Multiple jump ranges increase memory usage, control the number of jump ranges
4. **Iterator release**: Release iterator resources in time after use to avoid memory leaks

## 7. Summary

Jump Ranges is a powerful advanced feature in sfsDb. Through range iterators created by `RangeForAny`, it can skip unnecessary data during queries, greatly improving query efficiency. It is particularly suitable for scenarios requiring exclusion of specific ranges of data, such as e-commerce systems, IoT systems, and financial systems.

Reasonable use of jump ranges can significantly improve query performance, reduce unnecessary data processing, and make sfsDb more efficient when handling complex query scenarios.