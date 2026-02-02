# Querying Data

## 6.1 Basic Search

```go
// Basic search: exact match
searchFields := map[string]any{
    "name": "Zhang San",
}
iter := table.Search(&searchFields)
defer GlobalTableIterPool.Put(iter)

// Get all matching records
records := iter.GetRecords(true)
defer record.PutRecords(records)   
for _, record := range records {
    fmt.Printf("Found record: %v\n", record)
}
```

## 6.2 Using Comparison Operators

sfsDb supports multiple comparison operators, located in the `util` package:

| Operator | Description | Example |
|---------|-------------|---------|
| `Equal` | Equal to | `util.Equal` |
| `NotEqual` | Not equal to | `util.NotEqual` |
| `GreaterThan` | Greater than | `util.GreaterThan` |
| `GreaterThanOrEqual` | Greater than or equal to | `util.GreaterThanOrEqual` |
| `LessThan` | Less than | `util.LessThan` |
| `LessThanOrEqual` | Less than or equal to | `util.LessThanOrEqual` |
| `Like` | Prefix match (similar to SQL LIKE) | `util.Like` |

## 6.3 Comparison Operator Usage Examples

```go
import (
    "github.com/liaoran123/sfsDb/util"
)

// Example: Using comparison operators for search

// 1. Search for users older than 30
fmt.Println("\nUsers older than 30:")
ageGt30 := map[string]any{
    "age": 30,
}
iterGt30 := table.Search(&ageGt30, util.GreaterThan) // Pass comparison operator as second parameter
defer GlobalTableIterPool.Put(iterGt30)
recordsGt30 := iterGt30.GetRecords(true)
defer record.PutRecords(recordsGt30)   
for _, record := range recordsGt30 {
    fmt.Printf("   - %s: %d years old\n", record["name"], record["age"])
}

// 2. Prefix search (Like operator) - default is Like operation
fmt.Println("\nUsers with email starting with 'user':")
emailPrefix := map[string]any{
    "email": "user",
}
iterPrefix := table.Search(&emailPrefix) // Default uses util.Like operator, here like is actually prefix match
defer GlobalTableIterPool.Put(iterPrefix)
recordsPrefix := iterPrefix.GetRecords(true)
defer record.PutRecords(recordsPrefix)      
for _, record := range recordsPrefix {
    fmt.Printf("   - %s: %s\n", record["name"], record["email"])
}

// 3. Explicitly use Like operator
fmt.Println("\nUsers with name starting with 'Zhang':")
namePrefix := map[string]any{
    "name": "Zhang",
}
iterName := table.Search(&namePrefix, util.Like) // Explicitly specify util.Like operator
defer GlobalTableIterPool.Put(iterName)
recordsName := iterName.GetRecords(true)
defer record.PutRecords(recordsName)      
for _, record := range recordsName {
    fmt.Printf("   - %s: %s\n", record["name"], record["email"])
}

// 4. Exact match search
fmt.Println("\nExact search for user named 'Zhang San':")
exactSearch := map[string]any{
    "name": "Zhang San",
}
iterExact := table.Search(&exactSearch, util.Equal) // Explicitly specify util.Equal operator
defer GlobalTableIterPool.Put(iterExact)
recordsExact := iterExact.GetRecords(true)
defer record.PutRecords(recordsExact)   
for _, record := range recordsExact {
    fmt.Printf("   - %s: %s\n", record["name"], record["email"])
}

// 5. Not equal search
fmt.Println("\nUsers with id not equal to 1:")
notEqualSearch := map[string]any{
    "id": 1,
}
iterNotEqual := table.Search(&notEqualSearch, util.NotEqual) // Use util.NotEqual operator
defer GlobalTableIterPool.Put(iterNotEqual)
defer record.PutRecords(recordsNotEqual)   
recordsNotEqual := iterNotEqual.GetRecords(true)
for _, record := range recordsNotEqual {
    fmt.Printf("   - %s: ID=%d\n", record["name"], record["id"])
}
```

## 6.4 Matcher Interface

sfsDb supports a custom matcher interface for implementing complex query logic:

```go
// Match interface definition
type Match interface {
    Match(record map[string]any) bool
}

// Custom matcher example: Age greater than specified value
 type AgeGreaterThanMatcher struct {
    MinAge int
}

func (m *AgeGreaterThanMatcher) Match(record map[string]any) bool {
    if age, ok := record["age"].(int); ok {
        return age > m.MinAge
    }
    return false
}

// Using custom matcher
matcher := &AgeGreaterThanMatcher{MinAge: 30}
records := table.MatchRecords(matcher)
for _, record := range records {
    fmt.Printf("Matching record: %v\n", record)
}
```

## 6.5 AND Matcher

sfsDb provides a built-in `AND` matcher for implementing SQL-like IN, NOT IN, AND, OR operations, especially suitable for multi-table join queries.

### 6.5.1 AND Matcher Overview

The `AND` matcher is used to determine if a record's field values are in a specified data set, supporting positive matching (IN) and negative matching (NOT IN).

**Core functions**:
- Implement SQL-like IN and NOT IN operations
- Support multi-field combination matching
- Suitable for multi-table join queries
- Support combination with other matchers

### 6.5.2 AND Struct Definition

```go
type AND struct {
    fields []string // Field names to match
    data   map[any]bool // Matching data set
    rule   bool // Matching rule: true=IN/AND, false=NOT IN/OR
}
```

**Field descriptions**:
- `fields`: List of field names to match, corresponding to keys in records
- `data`: Matching data set, generated by `TableIter.Map()` method or custom
- `rule`: Matching rule, `true` means IN/AND, `false` means NOT IN/OR

### 6.5.3 Constructor

```go
func NewAND(fields []string, data map[any]bool, rule ...bool) *AND
```

**Parameter descriptions**:
- `fields`: List of field names to match
- `data`: Matching data set
- `rule`: Optional parameter, matching rule, default is `true`

### 6.5.4 Usage Examples

**Example 1: Basic IN Operation**

```go
// Suppose we have a user table and need to query users with ID in the specified set

// 1. Get ID set
idMap := map[any]bool{1: true, 3: true, 5: true}

// 2. Create AND matcher
andMatcher := match.NewAND([]string{"id"}, idMap)

// 3. Use matcher
iter := table.Search(&map[string]any{"id": nil})
defer GlobalTableIterPool.Put(iter)

iter.SetMatch(andMatcher)
records := iter.GetRecords(true)
defer record.PutRecords(records)   

// Result: Returns users with ID 1, 3, 5
```

**Example 2: NOT IN Operation**

```go
// Query users with ID not in the specified set

// 1. Get ID set
idMap := map[any]bool{1: true, 3: true, 5: true}

// 2. Create AND matcher, set rule=false for NOT IN
andMatcher := match.NewAND([]string{"id"}, idMap, false)

// 3. Use matcher
iter := table.Search(&map[string]any{"id": nil})
defer GlobalTableIterPool.Put(iter)

iter.SetMatch(andMatcher)
records := iter.GetRecords(true)
defer record.PutRecords(records)   

// Result: Returns users with ID not 1, 3, 5
```

**Example 3: Multi-table Join Query (Key Example)**

```go
// Implement SQL-like join query: SELECT table1.* FROM table1, table2 WHERE table1.id = table2.id

// 1. Get iterators for both tables
iter1 := table1.Search(&map[string]any{"id": nil})
defer GlobalTableIterPool.Put(iter1)

iter2 := table2.Search(&map[string]any{"id": nil})
defer GlobalTableIterPool.Put(iter2)

// 2. Get ID mapping from table2
// Map() method generates map[any]bool, keys are values of specified fields
idMap := iter2.Map()
defer PutMap(map2)

// 3. Create AND matcher
// Match if table1's id field is in table2's id set
andMatcher := match.NewAND([]string{"id"}, idMap)

// 4. Set matcher and get results
iter1.SetMatch(andMatcher)
records := iter1.GetRecords(true)
defer record.PutRecords(records)   

// Result: Returns records from table1 where ID matches ID in table2
```

**Example 4: Multi-table Join NOT Operation**

```go
// Implement SQL-like join query: SELECT table1.* FROM table1, table2 WHERE table1.id != table2.id

// 1. Get iterators for both tables
iter1 := table1.Search(&map[string]any{"id": nil})
defer GlobalTableIterPool.Put(iter1)

iter2 := table2.Search(&map[string]any{"id": nil})
defer GlobalTableIterPool.Put(iter2)

// 2. Get ID mapping from table2
idMap := iter2.Map()
defer PutMap(map2)

// 3. Create AND matcher, set rule=false for NOT IN
andMatcher := match.NewAND([]string{"id"}, idMap, false)

// 4. Set matcher and get results
iter1.SetMatch(andMatcher)
records := iter1.GetRecords(true)
defer record.PutRecords(records)   

// Result: Returns records from table1 where ID does not match ID in table2
```

### 6.5.5 Composite Primary Key Matching

The `AND` matcher is mainly used for primary key matching, especially for composite primary keys. When the `fields` parameter contains multiple field names, it will merge these field values into a single value for matching through `util.MergeFields()`, which is the typical way to handle composite primary keys:

```go
// Example: Composite primary key matching, assuming table has composite primary key (user_id, product_id)

// 1. Get composite primary key mapping from another table
// Assuming iter2 is an iterator for a table with composite primary keys
// Map("user_id", "product_id") generates mapping of composite primary key values
combinedKeyMap := iter2.Map("user_id", "product_id")

// 2. Create AND matcher
// Match if current table's composite primary key (user_id, product_id) is in another table's composite primary key set
andMatcher := match.NewAND([]string{"user_id", "product_id"}, combinedKeyMap)

// 3. Use matcher
iter1.SetMatch(andMatcher)
records := iter1.GetRecords(true)
defer record.PutRecords(records)   

// Result: Returns records with composite primary keys matching those in another table
```

**Notes**:
- The AND struct is mainly designed for primary key value matching, especially between multiple iterators
- When the `fields` parameter contains multiple fields, it treats these fields as a composite primary key
- The format of merged field values depends on the implementation of `util.MergeFields()`
- For non-primary key field combination matching, it is recommended to use other matchers (such as FieldComparison) or custom matchers

**Design intent of the AND matcher**:
```go
/*
//The purpose of this structure is to match primary key values of multiple iterators for same or different matching
//rule=true: primary key values of multiple iterators must be the same to match successfully
//rule=false: primary key values of multiple iterators must be different to match successfully
//rule=false is mainly used in jump queries, such as field not in (1,2,3) in SQL statements, then data=map[any]bool{1:true,2:true,3:true}
*/
```

### 6.5.6 Combination with Other Matchers

The `AND` matcher can be used in combination with other matchers to implement more complex query logic:

```go
// Implement: SELECT * FROM table WHERE id IN (1,3,5) AND age > 25

// 1. Create AND matcher (ID IN (1,3,5))
idMap := map[any]bool{1: true, 3: true, 5: true}
idMatcher := match.NewAND([]string{"id"}, idMap)

// 2. Create AgeGreaterThanMatcher (age > 25)
ageMatcher := &AgeGreaterThanMatcher{MinAge: 25}

// 3. Use combination matcher
iter := table.Search(&map[string]any{"id": nil})
defer GlobalTableIterPool.Put(iter)

// Set multiple matchers, they have an AND relationship
iter.SetMatch(idMatcher, ageMatcher)
records := iter.GetRecords(true)
defer record.PutRecords(records)   

// Result: Returns users with ID 1, 3, 5 and age greater than 25
```

### 6.5.7 Advantages of AND Matcher

1. **Efficient multi-table joins**: Avoid nested loops through precomputed mapping tables, improving join query efficiency
2. **Flexible matching rules**: Support IN, NOT IN, AND, OR and other matching methods
3. **Support multi-field combinations**: Can match based on combined values of multiple fields
4. **Easy to combine with other matchers**: Can be used in combination with other custom or built-in matchers
5. **Suitable for complex query scenarios**: Especially suitable for query scenarios that need to relate multiple tables or sets

Through the `AND` matcher, sfsDb implements efficient and flexible multi-table join query functionality, providing users with powerful data query capabilities.

## 6.6 FieldComparison Matcher

sfsDb provides a built-in `FieldComparison` matcher for implementing various comparison operations, supporting application of matchers to iterators, especially suitable for primary key iterator matching of non-indexed fields:

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/match"
)

func main() {
    // Create test table
    table, err := engine.TableNew("test_field_comparison")
    if err != nil {
        panic(err)
    }

    // Set table fields
    fields := map[string]any{
        "id":     0,
        "name":   "",
        "age":    0,
        "score":  0.0,
        "active": false,
    }

    err = table.SetFields(fields)
    if err != nil {
        panic(err)
    }

    // Create primary key index
    pkIndex, err := engine.DefaultPrimaryKeyNew("pk")
    if err != nil {
        panic(err)
    }
    pkIndex.AddFields("id")
    err = table.CreateIndex(pkIndex)
    if err != nil {
        panic(err)
    }

    // Insert test data
    testData := []map[string]any{
        {"id": 1, "name": "Alice", "age": 20, "score": 85.5, "active": true},
        {"id": 2, "name": "Bob", "age": 25, "score": 90.0, "active": true},
        {"id": 3, "name": "Charlie", "age": 30, "score": 75.5, "active": false},
        {"id": 4, "name": "David", "age": 35, "score": 95.0, "active": true},
        {"id": 5, "name": "Eve", "age": 40, "score": 80.0, "active": false},
    }

    for _, record := range testData {
        _, err = table.Insert(&record)
        if err != nil {
            panic(err)
        }
    }

    // FieldComparison matcher example
    fmt.Println("=== FieldComparison Matcher Example ===")

    // 1. Using FieldComparison for comparison
    // Get iterator
    iter := table.Search(&map[string]any{"id": nil})
    defer GlobalTableIterPool.Put(iter)

    // Create FieldComparison matcher
    matcher := match.NewFieldComparison("age", match.GreaterThan, 25)

    // Set matcher to iterator
    iter.SetMatch(matcher)

    // Get filtered records
    records := iter.GetRecords(true)
    defer record.PutRecords(records)   
    fmt.Printf("Records with age greater than 25 (%d records):\n", len(records))
    for _, record := range records {
        fmt.Printf("   - %v\n", record)
    }

    // 2. Using convenience functions to create matchers
    iter2 := table.Search(&map[string]any{"id": nil})
    defer GlobalTableIterPool.Put(iter2)

    // Using GreaterThanMatch convenience function
    highScoreMatcher := match.NewGreaterThanMatch("score", 90.0)
    iter2.SetMatch(highScoreMatcher)

    highScoreRecords := iter2.GetRecords(true)
    defer record.PutRecords(highScoreRecords)   
    fmt.Printf("\nRecords with score greater than 90 (%d records):\n", len(highScoreRecords))
    for _, record := range highScoreRecords {
        fmt.Printf("   - %v\n", record)
    }

    // 3. Using EqualMatch convenience function
    iter3 := table.Search(&map[string]any{"id": nil})
    defer GlobalTableIterPool.Put(iter3)

    inactiveMatcher := match.NewEqualMatch("active", false)
    iter3.SetMatch(inactiveMatcher)

    inactiveRecords := iter3.GetRecords(true)
    defer record.PutRecords(inactiveRecords)   
    fmt.Printf("\nInactive users (%d records):\n", len(inactiveRecords))
    for _, record := range inactiveRecords {
        fmt.Printf("   - %v\n", record)
    }
}
```

## 6.7 Comparison Operations Supported by FieldComparison

| Comparison Operation | Description | Convenience Function |
|---------------------|-------------|----------------------|
| `Equal` | Equal to | `NewEqualMatch` |
| `NotEqual` | Not equal to | `NewNotEqualMatch` |
| `GreaterThan` | Greater than | `NewGreaterThanMatch` |
| `GreaterThanOrEqual` | Greater than or equal to | `NewGreaterThanOrEqualMatch` |
| `LessThan` | Less than | `NewLessThanMatch` |
| `LessThanOrEqual` | Less than or equal to | `NewLessThanOrEqualMatch` |
| `Like` | Prefix match | `NewLikeMatch` |
| `Prefix` | Prefix match | `NewPrefixMatch` |
| `Suffix` | Suffix match | `NewSuffixMatch` |
| `Contains` | Contains match | `NewContainsMatch` |