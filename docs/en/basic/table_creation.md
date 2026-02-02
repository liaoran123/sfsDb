# Creating Tables and Setting Fields

## 2.1 Basic Table Creation/Opening

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // Create table
    table, err := engine.TableNew("users")
    if err != nil {
        panic(err)
    }
    
    // Set table fields
    fields := map[string]any{
        "id":   0,     // Auto-increment primary key (using numeric type)
        "name": "",    // String type
        "age":  0,     // Integer type
        "email": "",   // String type
    }
    err = table.SetFields(fields)
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Table created successfully")
}
```

## 2.2 Supported Field Types

sfsDb uses the `any` type (i.e., `interface{}`) to support all Go language types. The following table lists common type examples:

| Type Example | Description |
|-------------|-------------|
| `0`         | Integer types (int, int32, int64, uint, uint32, uint64, etc.) |
| `0.0`       | Floating-point types (float32, float64) |
| `""`        | String type |
| `false`     | Boolean type |
| `time.Now()` | Time type |
| `complex(1, 2)` | Complex number types (complex64, complex128) |

### 2.2.1 Custom Type Handling

For types not explicitly listed in the table (such as custom structs, slices, maps, etc.), sfsDb will use JSON serialization by default:

- **When storing**: The custom type is serialized into a JSON string through `json.Marshal()`, then stored as a byte array
- **When reading**: The JSON string is deserialized into the corresponding type through `json.Unmarshal()`

**Example: Using Custom Struct as Field Type**

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/storage"
)

// Custom struct type
type Address struct {
    City    string `json:"city"`
    Street  string `json:"street"`
    ZipCode string `json:"zip_code"`
}

func main() {
    // Initialize database
    _, err := storage.OpenDefaultDb("./custom_type_db")
    if err != nil {
        panic(err)
    }
    defer storage.CloseDb()

    // Create table
    table, err := engine.TableNew("users")
    if err != nil {
        panic(err)
    }

    // Set fields, including custom struct type
    fields := map[string]any{
        "id":      0,
        "name":    "",
        "age":     0,
        "address": Address{}, // Use custom struct type
    }
    err = table.SetFields(fields)
    if err != nil {
        panic(err)
    }

    // Insert record with custom struct
    user := map[string]any{
        "name": "Zhang San",
        "age":  30,
        "address": Address{
            City:    "Beijing",
            Street:  "Jianguo Road, Chaoyang District",
            ZipCode: "100022",
        },
    }
    id, err := table.Insert(&user)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Insert successful, ID: %d\n", id)

    // Read record
    readFields := map[string]any{"id": id}
    recordData, err := table.Read(&readFields)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Read record: %v\n", recordData)
}
```

### 2.2.2 Type Conversion Mechanism

sfsDb internally uses type conversion functions in the `util` package to handle different types:

- `AnyToBytes()`: Convert any type to byte array for storage
- `AnyToStr()`: Convert any type to string for display or comparison
- `StrToAny()`: Convert string to specified type for reading

For types not explicitly supported, these functions will use JSON serialization/deserialization to ensure all Go language types can be correctly stored and read.

### 2.2.3 Notes

1. **Custom types must be serializable**: Custom types used for fields must be JSON serializable, otherwise storage will fail
2. **Performance considerations**: JSON serialization and deserialization of complex custom types will bring some performance overhead
3. **Type safety**: When reading custom types, ensure type matching to avoid runtime errors
4. **Field extension**: New fields can be added to the table at any time without modifying the table structure
5. **Default values**: Field default values are used for type inference, actual storage will use the inserted data

Through this design, sfsDb implements true semi-structured data storage, supporting both common basic types and flexible handling of complex custom types.

## 2.3 Example: Creating a Table with Various Field Types

```go
// Create table with multiple field types
complexTable, err := engine.TableNew("complex_table")
if err != nil {
    panic(err)
}

complexFields := map[string]any{
    "id":        0,          // Integer type (primary key)
    "name":      "",         // String type
    "age":       0,          // Integer type
    "salary":    0.0,        // Floating-point type
    "active":    false,      // Boolean type
    "created_at": time.Now(), // Time type
    "description": "",       // String type (for full-text search)
    "tags":      "",         // String type (for tags)
}
err = complexTable.SetFields(complexFields)
if err != nil {
    panic(err)
}
```