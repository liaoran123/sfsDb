# Multi-table Join Query

This document introduces the multi-table join query functionality in sfsDb, which allows you to perform SQL JOIN-like operations between multiple tables to implement more complex data query requirements.

## Implementation Principle

The multi-table join query functionality in sfsDb is based on the following core components:

1. **TableIter Iterator**: Used to traverse records in a table
2. **Map() Method**: Extracts values of a specified field from a table as a map
3. **match Package**: Provides creation and management of matching conditions
4. **SetMatch() Method**: Sets matching conditions on an iterator
5. **GetRecords() Method**: Retrieves records that meet the matching conditions

## Basic Usage

### 1. Preparation

First, you need to create multiple tables and insert test data:

```go
// Create table 1
table1, err := TableNew("table1")
if err != nil {
    log.Fatalf("Failed to create table: %v", err)
}

// Set fields
fields := map[string]any{"id": 0, "name": "", "age": 0}
err = table1.SetFields(fields)
if err != nil {
    log.Fatalf("Failed to set fields: %v", err)
}

// Create primary key index
pk, _ := DefaultPrimaryKeyNew("pk")
pk.AddFields("id")
err = table1.CreateIndex(pk)
if err != nil {
    log.Fatalf("Failed to create primary key index: %v", err)
}

// Insert test data
testData := []map[string]any{
    {"id": 1, "name": "Alice", "age": 20},
    {"id": 2, "name": "Bob", "age": 25},
    {"id": 3, "name": "Charlie", "age": 30},
}

for _, data := range testData {
    _, err := table1.Insert(&data)
    if err != nil {
        log.Fatalf("Failed to insert test data: %v", err)
    }
}

// Similarly create table 2 and insert data...
```

### 2. Perform Inner Join Query

Inner join returns matching records from both tables:

```go
// Create iterators for table1 and table2
iter1, err := table1.Search(&map[string]any{"id": nil}) // Traverse all records
defer iter1.Release()

iter2, err := table2.Search(&map[string]any{"id": nil}) // Traverse all records
defer iter2.Release()

// Extract id field values from table2 as a map
map2 := iter2.Map()
defer iter2.ReleaseMap(map2)

// Create matching condition: id field value of table1 must be in id map of table2
mach := match.NewAND([]string{"id"}, map2)

// Set matching condition on table1 iterator
iter1.SetMatch(mach)

// Get matching records (join result)
rd4 := iter1.GetRecords(true)
defer rd4.Release()

// Print join results
for _, record := range rd4 {
    fmt.Println(record)
}
```

### 3. Perform Outer Join Query

Outer join returns records from one table that do not match the other table:

```go
// Create non-matching condition: id field value of table1 not in id map of table2
mach1 := match.NewAND([]string{"id"}, map2, false)

// Set non-matching condition on table1 iterator
iter1.SetMatch(mach1)

// Get non-matching records (outer join result)
rd5 := iter1.GetRecords(true)
defer rd5.Release()

// Print outer join results
for _, record := range rd5 {
    fmt.Println(record)
}
```

### 4. Perform Multi-table Join Query

You can perform join queries between multiple tables:

```go
// Create iterator for table3
iter3, err := table3.Search(&map[string]any{"id": nil}) // Traverse all records
defer iter3.Release()

// Extract id field values from table3 as a map
map3 := iter3.Map()
defer iter3.ReleaseMap(map3)

// Create matching condition: id field value of table1 must be in both id maps of table2 and table3
mach2 := match.NewAND([]string{"id"}, map3)

// Set multiple matching conditions on table1 iterator
iter1.SetMatch(mach, mach2)

// Get matching records (multi-table join result)
rd6 := iter1.GetRecords(true)
defer rd6.Release()

// Print multi-table join results
for _, record := range rd6 {
    fmt.Println(record)
}
```

## Advanced Usage

### 1. Combine with Condition Query

You can combine multi-table join with condition query:

```go
// Create condition query for age greater than 25
iterAgeGreater, _ := table1.Search(&map[string]any{"age": 25}, util.GreaterThan)
defer iterAgeGreater.Release()

// Set multi-table join condition
iterAgeGreater.SetMatch(mach) // Use previously created table2 join condition

// Get results
rdAgeGreater := iterAgeGreater.GetRecords(true)
defer rdAgeGreater.Release()

// Print results
for _, record := range rdAgeGreater {
    fmt.Println(record)
}
```

### 2. Resource Management

When using multi-table join queries, be sure to pay attention to resource management:

```go
// Properly release iterator
iter1, err := table1.Search(&map[string]any{"id": nil})
defer iter1.Release() // Or use object pool: defer GlobalTableIterPool.Put(iter1)

// Properly release map
map2 := iter2.Map()
defer iter2.ReleaseMap(map2) // Or use: defer PutMap(map2)

// Properly release record set
rd4 := iter1.GetRecords(true)
defer rd4.Release() // Or use: defer record.PutRecords(rd4)
```

## Example: Complete Multi-table Query

The following is a complete multi-table query example based on the `TestTestSelectForJoin1` function:

```go
// Code for creating tables and inserting data is omitted...

// Create iterators
iter1, err := table1.Search(&map[string]any{"id": nil})
defer iter1.Release()

iter2, err := table2.Search(&map[string]any{"id": nil})
defer iter2.Release()

iter3, err := table3.Search(&map[string]any{"id": nil})
defer iter3.Release()

// Test 1: Inner join query - select table1.* from table1,table2 where table1.id=table2.id
fmt.Println("=== Inner Join Query ===")
map2 := iter2.Map()
defer iter2.ReleaseMap(map2)
mach := match.NewAND([]string{"id"}, map2)
iter1.SetMatch(mach)
rd4 := iter1.GetRecords(true)
defer rd4.Release()
for _, record := range rd4 {
    fmt.Println(record)
}

// Test 2: Outer join query - select table1.* from table1,table2 where table1.id!=table2.id
fmt.Println("\n=== Outer Join Query ===")
mach1 := match.NewAND([]string{"id"}, map2, false)
iter1.SetMatch(mach1)
rd5 := iter1.GetRecords(true)
defer rd5.Release()
for _, record := range rd5 {
    fmt.Println(record)
}

// Test 3: Multi-table join query - select table1.* from table1,table2,table3 where table1.id=table2.id and table1.id=table3.id
fmt.Println("\n=== Multi-table Join Query ===")
map3 := iter3.Map()
defer iter3.ReleaseMap(map3)
mach2 := match.NewAND([]string{"id"}, map3)
iter1.SetMatch(mach, mach2)
rd6 := iter1.GetRecords(true)
defer rd6.Release()
for _, record := range rd6 {
    fmt.Println(record)
}
```

## Performance Optimization

1. **Use Indexes**: Ensure that join fields have indexes to improve query performance
2. **Reasonable Use of Object Pools**: For frequently executed queries, using object pools can reduce memory allocation
3. **Limit Result Set Size**: Use pagination or TopN functionality to limit the number of returned records
4. **Avoid Unnecessary Joins**: Only join tables that are truly needed

## Common Issues

### 1. Join Query Returns Empty Results

- Check if there are matching field values in both tables
- Ensure correct field names are used
- Verify that tables have data

### 2. High Memory Usage

- For large tables, consider using object pools
- Timely release maps and record sets that are no longer used
- Consider using pagination queries to reduce the amount of data loaded at once

### 3. Slow Query Performance

- Ensure join fields have indexes
- Reduce the number of joined tables
- Consider using more specific search conditions to reduce the number of records that need to be traversed

## Summary

sfsDb's multi-table join query functionality provides a flexible and efficient way to perform SQL JOIN-like operations. By properly using iterators, maps, and matching conditions, you can implement complex data query requirements while maintaining code simplicity and readability.

Please refer to the `TestTestSelectForJoin1` function in the `engine/tableiter_test.go` file for more example code.