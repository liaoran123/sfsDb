# Updating Records

## 1. Overview of Record Updates

sfsDb supports updating existing records through the `Update` method. This allows you to modify field values in existing records. Update operations require primary key fields to locate the record to be modified, followed by the fields and their new values to update.

## 2. Basic Update Operations

### 2.1 Simple Update

The basic syntax for updating a record using the `Update` method is as follows:

```go
// Update record
updateData := map[string]any{
    "id":   1,        // Primary key field to locate the record
    "name": "John",   // Field to update
    "age":  35,        // Field to update
}

err = table.Update(&updateData)
if err != nil {
    panic(err)
}
```

### 2.2 Batch Update

You can also use the `Update` method in batch operations:

```go
// Get batch operation object
batch := storage.KVDb.GetBatch()
if batch == nil {
    panic("Failed to get batch operation object")
}

// Batch update multiple records
for i := 1; i <= 5; i++ {
    updateData := map[string]any{
        "id":   i,
        "age":  20 + i, // Update to new value
        "email": fmt.Sprintf("user%d@example.com", i),
    }
    
    err = table.Update(&updateData, batch)
    if err != nil {
        panic(err)
    }
}

// Commit batch operation
err = storage.KVDb.WriteBatch(batch)
if err != nil {
    panic(fmt.Sprintf("Batch commit failed: %v", err))
}
```

### 2.3 Iterator Batch Update

`sfsDb` also provides batch update functionality through iterators. The `TableIter.Update` method can batch update records in query results:

```go
// Create search condition
searchFields := map[string]any{"age": map[string]any{"$gte": 25}}

// Get iterator
iter, err := table.Search(&searchFields)
if err != nil {
    panic(err)
}
defer iter.Release()

// Prepare update data
updateData := map[string]any{
    "status": "active", // Field to update
}

// Execute batch update, limit to maximum 100 records
err = iter.Update(&updateData, 100)
if err != nil {
    panic(err)
}

fmt.Println("Batch update completed")
```

#### How Iterator Batch Update Works

The internal implementation of `TableIter.Update` method is as follows:

1. **Batch Size Control**: By default, executes a batch write every 1000 records
2. **Automatic Batching**: When the batch size limit is reached, automatically executes the write and creates a new batch operation
3. **Primary Key Check**: Ensures each record has primary key fields
4. **Error Handling**: Stops and returns an error if any step fails
5. **Resource Management**: Ensures all pending batch operations are executed at the end of the function

#### Advantages of Iterator Batch Update

- **Simplified Code**: No need to manually iterate records and manage batch operations
- **Automatic Optimization**: Built-in batch size control to balance memory usage and performance
- **Flexible Limiting**: Can specify maximum number of records to update
- **Unified Error Handling**: Consistent error handling mechanism
- **Excellent Performance**: Reduces network round-trips and disk operations, improving batch update performance

## 3. Working Principle of Record Updates

### 3.1 Internal Implementation

The internal implementation flow of the `Update` method is as follows:

1. **Primary Key Check**: Verify that all primary key fields are provided
2. **Type Check**: Check that field types match
3. **Record Reading**: Read the record to be modified based on primary key
4. **Field Processing**: Determine which fields to update (excluding primary key fields)
5. **Record Update**: Update field values in the record
6. **Index Update**: Update related indexes

### 3.2 Key Code

```go
func (t *Table) Update(fields *map[string]any, batchs ...storage.Batch) error {
    // Check if all primary key fields are provided
    for _, field := range t.GetPrimaryFields() {
        if _, ok := (*fields)[field]; !ok {
            return fmt.Errorf("Must provide primary key field '%s'", field)
        }
    }
    
    // Check if field types match
    if err := t.CheckType(fields); err != nil {
        return err
    }
    
    // Read record
    record, err := t.Read(fields)
    if err != nil {
        return err
    }
    if record == nil {
        return fmt.Errorf("Record with primary key value '%v' does not exist", fields)
    }
    
    // Process update fields
    updateFields := GetStringSlice()
    defer PutStringSlice(updateFields)
    for field := range *fields {
        // Exclude primary key fields
        if t.GetPrimaryKey().MatchFields(field) {
            continue
        }
        updateFields = append(updateFields, field)
    }
    
    // Check if there are fields to update
    if len(updateFields) == 0 {
        return nil
    }
    
    // Parse record and update
    pk := t.GetPrimaryKey()
    fieldsBytes, err := pk.Parse(t.fieldsid, record)
    if err != nil {
        return err
    }
    
    // Execute update operation
    BatchContainer := NewBatchContainer(batch, t.indexs, t.id, t.kvStore)
    BatchContainer.Operation(fieldsBytes, updateFields...)
    
    // ... remaining code
}
```

## 4. Best Practices for Updating Records

### 4.1 Notes

1. **Must Provide Primary Key**: You must provide all primary key fields when updating a record, otherwise an error will be returned
2. **Field Type Matching**: The type of updated field values must match the type defined in the table
3. **Record Existence**: If no record is found based on the primary key, an error will be returned
4. **Index Updates**: When modifying indexed fields, the system automatically updates related indexes
5. **Batch Operations**: For updating multiple records, batch operations are recommended for better performance

### 4.2 Recommended Practices

**Recommended Process**:

1. **Verify Record Existence**: You can check if the record exists before updating
2. **Prepare Update Data**: Build a map containing primary key and fields to update
3. **Execute Update Operation**: Call the `Update` method to perform the update
4. **Handle Errors**: Properly handle possible errors
5. **Verify Results**: After updating, you can query the record to verify the update was successful

## 5. Example: Complete Record Update Flow

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/record"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // Initialize database
    _, err := storage.OpenDefaultDb("./update_demo_db")
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
        "id":    0,
        "name":  "",
        "age":   0,
        "email": "",
    }
    err = table.SetFields(fields)
    if err != nil {
        panic(err)
    }

    // Insert test data
    user := map[string]any{
        "name":  "John",
        "age":   30,
        "email": "john@example.com",
    }
    id, err := table.Insert(&user)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Insert successful, ID: %d\n", id)

    // Update record
    fmt.Println("\n=== Updating Record ===")
    
    updateData := map[string]any{
        "id":    id,        // Primary key field
        "name":  "John Updated",    // Update name
        "age":   31,        // Update age
        "email": "john_updated@example.com", // Update email
    }
    
    err = table.Update(&updateData)
    if err != nil {
        panic(err)
    }
    fmt.Println("Record updated successfully")

    // Verify update results
    fmt.Println("\n=== Verifying Update Results ===")
    
    searchData := map[string]any{"id": id}
iter, err := table.Search(&searchData)
if err != nil {
    panic(err)
}
defer iter.Release()
    
    records := iter.GetRecords(true)
    defer records.Release()
    
    for _, r := range records {
        fmt.Printf("Updated record: %v\n", r)
        
        // Check if fields were updated
        if r["name"] == "John Updated" {
            fmt.Println("✓ Name updated successfully")
        }
        if r["age"] == 31 {
            fmt.Println("✓ Age updated successfully")
        }
        if r["email"] == "john_updated@example.com" {
            fmt.Println("✓ Email updated successfully")
        }
    }

    fmt.Println("\nRecord update process completed")
}
```

## 6. Common Issues and Solutions

### 6.1 Error Handling

**Common Errors**:

1. **Missing Primary Key Fields**: Not providing all primary key fields when updating
2. **Record Not Found**: No record found based on the primary key
3. **Field Type Mismatch**: Updated field value type does not match the table definition
4. **Batch Operation Failure**: Partial failure during batch updates

**Solutions**:

1. **Ensure Primary Key Provided**: Confirm all primary key fields are included before updating
2. **Check Record Existence**: You can query the record first to ensure it exists
3. **Verify Field Types**: Ensure the updated value types are correct
4. **Use Transactions**: For important operations, use transactions to ensure atomicity

### 6.2 Performance Optimization

- **Batch Operations**: Use batch operations for updating multiple records
- **Only Update Necessary Fields**: Include only the fields that need to be updated to reduce data transfer
- **Avoid Frequent Updates**: Frequent updates can affect performance; design data structures合理
- **Index Considerations**: Modifying indexed fields will update indexes, which may affect performance

## 7. Summary

Record updates are an important feature of sfsDb, allowing you to conveniently update field values in existing records through the `Update` method. When using this feature, you need to provide primary key fields, ensure field type matching, and合理 use batch operations for better performance.

**Key Points**:

1. **Must Provide Primary Key**: All primary key fields must be included when updating records
2. **Field Type Matching**: Updated values must match the types defined in the table
3. **Batch Operation Optimization**: Use batch operations for updating multiple records
4. **Error Handling**: Properly handle possible error conditions
5. **Performance Considerations**: Design update operations合理 to avoid frequent modifications to indexed fields

By correctly using the record update feature, you can effectively maintain data accuracy and consistency, ensuring that your system's data state remains synchronized with business requirements.