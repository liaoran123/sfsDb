# Field Modification

## 1. Field Modification Overview

sfsDb supports dynamic table field modification, including adding new fields, modifying field names, and updating field types. Field modification operations need to follow a specific workflow to ensure data consistency and system stability.

## 2. Single Field Modification

### 2.1 Field Renaming

To modify a field name, follow these steps:

```go
// Field modification workflow: rename "position" field to "job_title"

// 1. First call UpdateFieldName to update field name mapping
err = table.UpdateFieldName("position", "job_title")
if err != nil {
    panic(err)
}

// 2. Then call SetFields to update field mapping
updatedFields := map[string]any{
    "id":        0,
    "name":      "",
    "age":       0,
    "email":     "",
    "job_title": "", // Use new field name
}
err = table.SetFields(updatedFields)
if err != nil {
    panic(err)
}
```

### 2.2 Adding New Fields

To add a new field, simply call the `SetFields` method and include the new field:

```go
// Add new field "department"
updatedFields := map[string]any{
    "id":         0,
    "name":       "",
    "age":        0,
    "email":      "",
    "job_title":  "",
    "department": "", // New field
}
err = table.SetFields(updatedFields)
if err != nil {
    panic(err)
}
```

## 3. Batch Field Modification

For cases where multiple fields need to be modified simultaneously, follow these steps:

```go
// Batch modify multiple fields

// 1. Call UpdateFieldName sequentially to update multiple field name mappings
err = table.UpdateFieldName("old_field1", "new_field1")
if err != nil {
    panic(err)
}

// 2. Update another field name
err = table.UpdateFieldName("old_field2", "new_field2")
if err != nil {
    panic(err)
}

// 3. Finally update field mapping
newFields := map[string]any{
    "id":         0,
    "name":       "",
    "new_field1": "", // Use new field name
    "new_field2": 0,   // Use new field name
    "new_field3": 0.0, // Newly added field
}
err = table.SetFields(newFields)
if err != nil {
    panic(err)
}
```

## 4. Field Modification Best Practices

### 4.1 Notes

1. **Data Compatibility**: When modifying field types, ensure existing data is compatible with the new type
2. **Index Impact**: Modifying indexed fields may require recreating indexes
3. **Batch Operations**: When modifying multiple fields, ensure all operations execute successfully
4. **Data Backup**: It is recommended to back up data before making extensive field modifications

### 4.2 Recommended Practices

**Recommended Process**:

1. **Plan Modifications**: Clearly define the fields to be modified and the modification method
2. **Back Up Data**: Ensure data security
3. **Execute Modifications**: Perform field modification operations in the correct order
4. **Verify Results**: Confirm that the modified table structure and data are correct
5. **Update Indexes**: If needed, recreate relevant indexes

## 5. Example: Complete Field Modification Process

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
    _, err := storage.OpenDefaultDb("./field_modification_db")
    if err != nil {
        panic(err)
    }
    defer storage.CloseDb()

    // Create table
    table, err := engine.TableNew("employees")
    if err != nil {
        panic(err)
    }

    // Set initial fields
    initialFields := map[string]any{
        "id":       0,
        "name":     "",
        "age":      0,
        "position": "", // Initial field name
        "salary":   0.0,
    }
    err = table.SetFields(initialFields)
    if err != nil {
        panic(err)
    }

    // Insert test data
    employee := map[string]any{
        "name":     "Zhang San",
        "age":      30,
        "position": "Engineer",
        "salary":   10000.0,
    }
    _, err = table.Insert(&employee)
    if err != nil {
        panic(err)
    }

    fmt.Println("Initial table structure and data creation completed")

    // Field modification process
    fmt.Println("\n=== Field Modification Process ===")

    // 1. Update field name
    fmt.Println("1. Rename 'position' field to 'job_title'")
    err = table.UpdateFieldName("position", "job_title")
    if err != nil {
        panic(err)
    }
    fmt.Println("Field name update successful")

    // 2. Update field mapping, add new field
    fmt.Println("2. Update field mapping, add 'department' field")
    updatedFields := map[string]any{
        "id":         0,
        "name":       "",
        "age":        0,
        "job_title":  "", // Use new field name
        "salary":     0.0,
        "department": "", // Newly added field
    }
    err = table.SetFields(updatedFields)
    if err != nil {
        panic(err)
    }
    fmt.Println("Field mapping update successful")

    // 3. Verify modification results
    fmt.Println("\n3. Verify modification results")
    
    // Query modified data
    searchFields := map[string]any{"id": 1}
    iter, _ := table.Search(&searchFields)   
    defer iter.Release()
    
    records := iter.GetRecords(true)
    defer records.Release()
    
    for _, record := range records {
        fmt.Printf("Modified record: %v\n", record)
        
        // Check if new fields exist
        if _, exists := record["job_title"]; exists {
            fmt.Println("✓ Field rename successful: position → job_title")
        }
        
        if _, exists := record["department"]; exists {
            fmt.Println("✓ New field added successfully: department")
        }
    }

    fmt.Println("\nField modification process completed")
}
```

## 6. Common Field Modification Issues

### 6.1 Error Handling

**Common Errors**:

1. **Field Not Exist**: Attempting to modify a non-existent field
2. **Index Conflict**: Modifying indexed fields without recreating indexes
3. **Data Type Incompatibility**: When modifying field types, existing data cannot be converted to the new type

**Solutions**:

1. **Verify Field Existence**: Check if the field exists before modification
2. **Recreate Indexes**: After modifying indexed fields, recreate relevant indexes
3. **Data Migration**: For type incompatibility cases, consider exporting data first, modifying fields, then reimporting

### 6.2 Performance Considerations

- **Batch Modifications**: For modifying multiple fields, batch operations are recommended
- **Avoid Frequent Modifications**: Frequent field modifications can affect performance; field structure should be determined as much as possible during the design phase
- **Index Reconstruction**: After modifying indexed fields, rebuild indexes promptly to maintain query performance

## 7. Summary

Field modification is an important feature of sfsDb, supporting dynamic system evolution and requirement changes. By following the correct workflow and best practices, field modification operations can be completed safely and efficiently, ensuring system stability and data consistency.

**Key Points**:

1. **Follow Workflow**: Update field name mapping first, then update field mapping
2. **Note Data Compatibility**: Ensure modifications do not cause data loss or type errors
3. **Update Related Indexes**: After modifying indexed fields, recreate relevant indexes
4. **Back Up Data**: Back up data before making major modifications to prevent unexpected situations

By properly using the field modification feature, sfsDb table structures can better adapt to changes in business requirements, improving system flexibility and maintainability.