# sfsDb Database Detailed User Guide

## 1. Database Initialization

### 1.1 Custom Database Path

By default, sfsDb uses the `kvdb` folder in the current directory as the database storage path. If you need to customize the database path, you can call the `OpenDefaultDb` function when starting the program:

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // Custom database path
    dbPath := "./my_custom_db"
    _, err := storage.OpenDefaultDb(dbPath)
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Database initialization successful")
}
```

### 1.2 Automatic Use of Default Path

If you don't call the `OpenDefaultDb` function, the system will automatically use the default path `./kvdb` when creating the table for the first time:

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
)

func main() {
    // When TableNew is called for the first time, the database will be automatically initialized with the default path "./kvdb"
    table, err := engine.TableNew("users")
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Table created successfully, database automatically initialized")
}
```

### 1.3 Database Closure

When the program ends, you can call the `CloseDb` function to close the database and release resources:

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
    
    // Close database when program ends
    defer storage.CloseDb()
    
    fmt.Println("Table created successfully")
}
```

## 2. Creating Tables and Setting Fields

### 2.1 Basic Table Creation/Opening

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

### 2.2 Supported Field Types

sfsDb uses the `any` type (i.e., `interface{}`) to support all Go language types. The following table lists common type examples:

| Type Example | Description |
|-------------|-------------|
| `0`         | Integer types (int, int32, int64, uint, uint32, uint64, etc.) |
| `0.0`       | Floating-point types (float32, float64) |
| `""`        | String type |
| `false`     | Boolean type |
| `time.Now()` | Time type |
| `complex(1, 2)` | Complex number types (complex64, complex128) |

#### 2.2.1 Custom Type Handling

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

#### 2.2.2 Type Conversion Mechanism

sfsDb internally uses type conversion functions in the `util` package to handle different types:

- `AnyToBytes()`: Convert any type to byte array for storage
- `AnyToStr()`: Convert any type to string for display or comparison
- `StrToAny()`: Convert string to specified type for reading

For types not explicitly supported, these functions will use JSON serialization/deserialization to ensure all Go language types can be correctly stored and read.

#### 2.2.3 Notes

1. **Custom types must be serializable**: Custom types used for fields must be JSON serializable, otherwise storage will fail
2. **Performance considerations**: JSON serialization and deserialization of complex custom types will bring some performance overhead
3. **Type safety**: When reading custom types, ensure type matching to avoid runtime errors
4. **Field extension**: New fields can be added to the table at any time without modifying the table structure
5. **Default values**: Field default values are used for type inference, actual storage will use the inserted data

Through this design, sfsDb implements true semi-structured data storage, supporting both common basic types and flexible handling of complex custom types.

### 2.3 Example: Creating a Table with Various Field Types

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

## 3. Inserting Data

### 3.1 Basic Insertion

```go
// Insert record
user := map[string]any{
    "name": "Zhang San",
    "age":  30,
    "email": "zhangsan@example.com",
}
id, err := table.Insert(&user)
if err != nil {
    panic(err)
}
fmt.Printf("Insert successful, ID: %d\n", id)
```

### 3.2 Auto-increment

For the primary key field named id, if omitted, sfsDb will automatically generate an incrementing value:

```go
// Do not specify ID field, system automatically generates
user2 := map[string]any{
    "name": "Li Si",
    "age":  25,
    "email": "lisi@example.com",
}
id2, err := table.Insert(&user2)
if err != nil {
    panic(err)
}
fmt.Printf("Automatically generated ID: %d\n", id2) // Output: 2

// Can also explicitly set ID to nil, same as automatic generation
user3 := map[string]any{
    "id":   nil, // Explicitly set to nil, same as automatic generation
    "name": "Wang Wu",
    "age":  30,
    "email": "wangwu@example.com",
}
id3, err := table.Insert(&user3)
if err != nil {
    panic(err)
}
fmt.Printf("Explicitly set to 0, automatically generated ID: %d\n", id3) // Output: 3
```

### 3.3 Batch Insertion

```go
// Batch insert multiple records
users := []map[string]any{
    {"name": "Wang Wu", "age": 35, "email": "wangwu@example.com"},
    {"name": "Zhao Liu", "age": 28, "email": "zhaoliu@example.com"},
    {"name": "Sun Qi", "age": 40, "email": "sunqi@example.com"},
}

for _, user := range users {
    _, err := table.Insert(&user)
    if err != nil {
        panic(err)
    }
}
fmt.Println("Normal batch insertion completed")

// Manual transaction batch insertion (more efficient batch operation)
fmt.Println("\nManual transaction batch insertion:")

// 1. Get batch operation object (via global storage.KVDb)
batch := storage.KVDb.GetBatch()
if batch == nil {
    panic("Cannot get batch operation object")
}

// 2. Prepare data to be inserted
batchUsers := []map[string]any{
    {"name": "Zhou Ba", "age": 26, "email": "zhouba@example.com"},
    {"name": "Wu Jiu", "age": 32, "email": "wujiu@example.com"},
    {"name": "Zheng Shi", "age": 38, "email": "zhengshi@example.com"},
}

// 3. Add multiple insert operations to the same batch
for i, user := range batchUsers {
    _, err := table.Insert(&user, batch) // Pass batch parameter, manually control transaction
    if err != nil {
        panic(fmt.Sprintf("Failed to insert record %d: %v", i+1, err))
    }
    fmt.Printf("Added record %d to batch\n", i+1)
}

// 4. Manually commit batch (all operations executed at once)
err = storage.KVDb.WriteBatch(batch)
if err != nil {
    panic(fmt.Sprintf("Batch commit failed: %v", err))
}
fmt.Println("Manual transaction batch insertion completed")

// 5. Verify insertion results
fmt.Println("\nVerify insertion results:")
allIter := table.ForData()
defer allIter.Release()
records := allIter.GetRecords(true)
fmt.Printf("Total %d records in table\n", len(records))

// Manual transaction batch operations support multiple combinations
fmt.Println("\nCombined operation example (insert + delete):")
batch2 := storage.KVDb.GetBatch()
if batch2 == nil {
    panic("Cannot get batch operation object")
}

// Add a new record
newUser := map[string]any{
    "name": "Test User",
    "age":  25,
    "email": "test@example.com",
}
_, err = table.Insert(&newUser, batch2)
if err != nil {
    panic(err)
}

// Delete an existing record
deleteUser := map[string]any{
    "id": 1, // Delete record with ID 1
}
err = table.Delete(&deleteUser, batch2)
if err != nil {
    panic(err)
}

// Commit combined operation
err = storage.KVDb.WriteBatch(batch2)
if err != nil {
    panic(fmt.Sprintf("Combined operation commit failed: %v", err))
}
fmt.Println("Combined operation completed")
```

## 4. Primary Key Management

### 4.1 Single Primary Key

```go
// Create single primary key table
userTable, err := engine.TableNew("user_table")
if err != nil {
    panic(err)
}

// Set fields
userFields := map[string]any{
    "id":   0,     // Primary key field
    "name": "",    // Regular field
    "age":  0,     // Regular field
}
err = userTable.SetFields(userFields)
if err != nil {
    panic(err)
}

// Create single primary key index
pk, err := engine.DefaultPrimaryKeyNew("pk_id")
err = pk.AddFields("id")
err = userTable.CreateIndex(pk)
if err != nil {
    panic(err)
}
```

### 4.2 Composite Primary Key

```go
// Create composite primary key table
orderTable, err := engine.TableNew("order_table")
if err != nil {
    panic(err)
}

// Set composite primary key fields
orderFields := map[string]any{
    "order_id": 0,     // Composite primary key field 1
    "user_id":  0,     // Composite primary key field 2
    "product":  "",    // Regular field
    "quantity": 0,     // Regular field
}
err = orderTable.SetFields(orderFields)
if err != nil {
    panic(err)
}

// Create composite primary key index
compositePk, err := engine.DefaultPrimaryKeyNew("pk_order_user")
if err != nil {
    panic(err)
}
// Add multiple primary key fields
compositePk.AddFields("order_id", "user_id")
err = orderTable.CreateIndex(compositePk)
if err != nil {
    panic(err)
}

// Insert composite primary key record
order := map[string]any{
    "order_id": 1,
    "user_id":  100,
    "product":  "Product A",
    "quantity": 2,
}
_, err = orderTable.Insert(&order)
if err != nil {
    panic(err)
}
```

## 5. Index Management

### 5.1 Index Interface Overview

sfsDb defines a complete index interface system, providing unified operation methods for different types of indexes.

#### 5.1.1 Basic Index Interface

All index types implement the basic `Index` interface, which defines the core methods of indexes:

```go
// Basic index interface, defines common methods for all index types
type Index interface {
    // Add index fields
    AddFields(field ...string)
    // Get index field list
    GetFields() []string
    Len() int
    setId(id uint8)
    GetId() uint8
    Name() string
    SetName(name string) error
    // Modify index field name
    UpdateFields(oldfields string, newfields string)
    // Delete index field
    DeleteFields(field ...string)
    // Concatenate prefix
    Prefix(tbid uint8) []byte
    // Concatenate values, no prefix needed
    Join(fieldsBytes *map[string][]byte) []byte
    // Concatenate prefix + value
    JoinPrefix(tbid uint8, val []byte) []byte
    // Concatenate index prefix + index value, calls JoinPrefix, Join methods
    JoinValue(fieldsBytes *map[string][]byte, tbid uint8, existFields ...string) []byte
    // Match index fields
    MatchFields(fields ...string) bool
}
```

#### 5.1.2 Index Type Hierarchy

sfsDb supports three main index types, each with corresponding interfaces and implementations:

1. **PrimaryKey** - Used to uniquely identify records
2. **NormalIndex** - Used to speed up queries
3. **FullTextIndex** - Used for text search

### 5.2 Primary Key Index

#### 5.2.1 Primary Key Index Interface

```go
// Primary key interface, embeds basic index interface
type PrimaryKey interface {
    Index
    // Set primary key ID
    GetID(fieldsBytes *map[string][]byte, existFields ...string) []byte
    GetfieldTypeLen(tablefields *map[string]any) *map[string]uint8
    Parse(fieldsid map[uint8]string, value []byte) (*map[string][]byte, error)
}
```

#### 5.2.2 Default Primary Key Index Implementation

```go
// Default primary key index
// When using composite primary keys, only fixed-length type combinations are supported.
// For string types, length must be specified. Otherwise, parsing may fail or escape issues may cause bugs.
type DefaultPrimaryKey struct {
    BaseIndex // Embeds basic index
}

func DefaultPrimaryKeyNew(name string) (*DefaultPrimaryKey, error)
```

#### 5.2.3 Usage Example

```go
// Create single primary key table
userTable, err := engine.TableNew("user_table")
if err != nil {
    panic(err)
}

// Set fields
userFields := map[string]any{
    "id":   0,     // Primary key field
    "name": "",    // Regular field
    "age":  0,     // Regular field
}
err = userTable.SetFields(userFields)
if err != nil {
    panic(err)
}

// Create single primary key index
pk, err := engine.DefaultPrimaryKeyNew("pk_id")
err = pk.AddFields("id")
err = userTable.CreateIndex(pk)
if err != nil {
    panic(err)
}

// Create composite primary key index
compositePk, err := engine.DefaultPrimaryKeyNew("pk_order_user")
if err != nil {
    panic(err)
}
// Add multiple primary key fields
compositePk.AddFields("order_id", "user_id")
err = orderTable.CreateIndex(compositePk)
if err != nil {
    panic(err)
}
```

### 5.3 Normal Index

#### 5.3.1 Normal Index Interface

```go
// Normal index interface, embeds basic index interface
type NormalIndex interface {
    Index
    // Convert index value to primary key map value

    // Since NormalIndex completely matches the index interface, a Tag method is needed to distinguish if it's a secondary index.
    Tag() bool
    Parse(primaryFields []string, pkfieldTypeLen *map[string]uint8, value []byte) (*map[string][]byte, error)
}
```

#### 5.3.2 Default Normal Index Implementation

```go
// Default normal index, secondary index
type DefaultNormalIndex struct {
    BaseIndex // Embeds basic index
}

func DefaultNormalIndexNew(name string) (*DefaultNormalIndex, error)

// Tag method returns true, indicating it's a secondary index.
func (dni *DefaultNormalIndex) Tag() bool
```

#### 5.3.3 Usage Example

```go
// Create normal index
normalIndex, err := engine.DefaultNormalIndexNew("idx_name")
if err != nil {
    panic(err)
}
// Add index field
normalIndex.AddFields("name")
err = table.CreateIndex(normalIndex)
if err != nil {
    panic(err)
}

// Create composite index
compositeIndex, err := engine.DefaultNormalIndexNew("idx_name_age")
if err != nil {
    panic(err)
}
// Add multiple index fields
compositeIndex.AddFields("name", "age")
err = table.CreateIndex(compositeIndex)
if err != nil {
    panic(err)
}
```

### 5.4 Full-text Index

#### 5.4.1 Full-text Index Interface

```go
// Full-text index interface, embeds basic index interface
type FullTextIndex interface {
    Index
    SetFullField(field string, len int) error
    //GetFtlen() int
    // Concatenate full-text index values
    JoinFullValues(fieldsBytes *map[string][]byte, tbid uint8, existFields ...string) [][]byte
    // Tokenization method
    Tokenize(nr string, ftlen int) (tokens []string)
    Parse(primaryFields []string, pkfieldTypeLen *map[string]uint8, value []byte) (*map[string][]byte, error)
}
```

#### 5.4.2 Default Full-text Index Implementation

```go
// Default full-text index
type DefaultFullTextIndex struct {
    BaseIndex // Embeds basic index
    // Full-text index split field
    ftsplit string
    // Split length
    ftlen int
}

func DefaultFullTextIndexNew(name string) (*DefaultFullTextIndex, error)
```

#### 5.4.3 Usage Example

```go
// Create full-text index
fullTextIndex, err := engine.DefaultFullTextIndexNew("fulltext_desc")
if err != nil {
    panic(err)
}
// Add fields, last field must be primary key
// Full-text index must include primary key (single or composite) in full, generally at the end. Otherwise, index uniqueness will be lost.
fullTextIndex.AddFields("description", "id")
or
fullTextIndex.AddFields("description", "id", "did") // "id", "did" are composite primary keys
// Set full-text index field and length
fullTextIndex.SetFullField("description", 5) // 5 represents full-text index length. Generally choose 5 or 7 for Chinese. Mainly supports ideographic writing. This full-text index is not very suitable for non-ideographic writing.
err = table.CreateIndex(fullTextIndex)
if err != nil {
    panic(err)
}
```

### 5.5 Example: Creating a Table with Multiple Index Types

```go
// Create table
indexDemoTable, err := engine.TableNew("index_demo")
if err != nil {
    panic(err)
}

// Set fields
indexFields := map[string]any{
    "id":    0,
    "name":  "",
    "age":   0,
    "email": "",
    "city":  "",
}
err = indexDemoTable.SetFields(indexFields)
if err != nil {
    panic(err)
}

// 1. Create primary key index
primaryKey, err := engine.DefaultPrimaryKeyNew("pk_id")
primaryKey.AddFields("id")
err = indexDemoTable.CreateIndex(primaryKey)
if err != nil {
    panic(err)
}

// 2. Create normal index
emailIndex, err := engine.DefaultNormalIndexNew("idx_email")
emailIndex.AddFields("email")
err = indexDemoTable.CreateIndex(emailIndex)
if err != nil {
    panic(err)
}

// 3. Create composite index
cityAgeIndex, err := engine.DefaultNormalIndexNew("idx_city_age")
cityAgeIndex.AddFields("city", "age")
err = indexDemoTable.CreateIndex(cityAgeIndex)
if err != nil {
    panic(err)
}
```

### 5.6 Index Implementation Details

#### 5.6.1 Basic Index Structure

All index types embed the `BaseIndex` struct, which provides common index functionality:

```go
// Basic index struct, contains common fields and methods for all index types
type BaseIndex struct {
    fields []string
    id     uint8
    name   string
}
```

#### 5.6.2 Index Field Handling

- **Add fields**: `AddFields(field ...string)` - Add one or more fields to the index
- **Get fields**: `GetFields() []string` - Get all fields of the index
- **Update field name**: `UpdateFields(oldfields string, newfields string)` - Modify index field name
- **Delete field**: `DeleteFields(field ...string)` - Delete fields from the index

#### 5.6.3 Index Value Concatenation

Indexes internally use byte arrays to store index values, providing multiple concatenation methods:

- **Prefix**: Concatenate index prefix
- **Join**: Concatenate field values
- **JoinPrefix**: Concatenate prefix and value
- **JoinValue**: Concatenate complete index value

#### 5.6.4 Composite Index Notes

1. **Composite primary key limitation**: Composite primary keys only support combinations of fixed-length types; string types must specify length
2. **Field order**: The order of index fields affects query performance; place the most commonly used fields first
3. **Index size**: Indexes increase storage overhead; only create necessary indexes

### 5.7 Index Management System

#### 5.7.1 Indexs Struct

`Indexs` is the core struct for sfsDb's index management, responsible for managing all indexes of a table:

```go
// Index management structure
type Indexs struct {
    id     uint8
    indexs []Index
    fields *map[string]any // Table fields
}
```

**Main functions**:
- Create indexes (check field existence, index name uniqueness, etc.)
- Delete indexes
- Match indexes (priority: primary key index, then normal index, finally full-text index)
- Manage index lifecycle

#### 5.7.2 Table-level Index Management Methods

`sfsDb` provides rich table-level index management methods to facilitate users in creating and managing indexes:

##### Creating Indexes

```go
// Create custom index
func (t *Table) CreateIndex(index Index) error

// Create normal composite index
func (t *Table) CreateCompositeIndex(name string, fields ...string) error

// Create composite primary key index
func (t *Table) CreateCompositePrimaryKey(name string, fields ...string) error

// Create primary key index (supports single or multiple fields)
func (t *Table) CreatePrimaryKey(fields ...string) error

// Create normal index (simplified version, directly specify name and fields)
func (t *Table) CreateSimpleIndex(name string, fields ...string) error
```

##### Getting Indexes

```go
// Get primary key index
func (t *Table) GetPrimaryKey() PrimaryKey

// Get all indexes
func (t *Table) GetAllIndexes() []Index

// Get index by name
func (t *Table) GetIndexByName(name string) Index

// Get all indexes containing the field by field name
func (t *Table) GetIndexesByField(field string) []Index

// Match index
func (t *Table) MatchIndex(fields ...string) Index
```

##### Deleting Indexes

```go
// Delete index by name
func (t *Table) DropIndex(name string) error

// Delete primary key index
func (t *Table) DropPrimaryKey() error
```

#### 5.7.3 Index Management Example

```go
// Example: Table-level index management

// 1. Create various types of indexes
// Create primary key index
table.CreatePrimaryKey("id")

// Create composite primary key index
table.CreateCompositePrimaryKey("pk_user_id", "user_id", "product_id")

// Create normal index
table.CreateSimpleIndex("idx_name", "name")

// Create composite index
table.CreateCompositeIndex("idx_name_age", "name", "age")

// 2. Get index information
// Get primary key index
pk := table.GetPrimaryKey()

// Get all indexes
allIndexes := table.GetAllIndexes()

// Get index by name
nameIndex := table.GetIndexByName("idx_name")

// Get indexes by field
nameIndexes := table.GetIndexesByField("name")

// 3. Delete indexes
// Delete normal index
table.DropIndex("idx_name")

// Delete primary key index
table.DropPrimaryKey()
```

### 5.8 Index Optimization Suggestions

1. **Choose appropriate index types**: Choose appropriate index types based on query needs
   - Unique identification of records: Use primary key index
   - Speed up normal queries: Use normal index
   - Text search: Use full-text index

2. **Reasonable index field design**:
   - Primary key fields should be unique and stable
   - Normal indexes should be fields frequently used in query conditions
   - Full-text indexes should be fields that need text search

3. **Control index quantity**:
   - Too many indexes will affect write performance
   - Only create indexes for frequently queried fields

4. **Use composite indexes**:
   - For multi-field queries, composite indexes are more efficient than multiple single-column indexes
   - Follow the leftmost prefix principle when designing composite indexes

5. **Regular index maintenance**:
   - For frequently updated tables, rebuild indexes regularly
   - Delete unused indexes

By rationally using and optimizing indexes, you can significantly improve the query performance of sfsDb, especially when processing large amounts of data.

## 6. Full-text Search

### 6.1 Creating Full-text Index

```go
// Create full-text index
fullTextIndex, err := engine.DefaultFullTextIndexNew("fulltext_desc")
if err != nil {
    panic(err)
}
// Add fields, last field must be primary key
fullTextIndex.AddFields("description", "id")
// Set full-text index field and length
fullTextIndex.SetFullField("description", 5) // 5 represents full-text index length
err = table.CreateIndex(fullTextIndex)
if err != nil {
    panic(err)
}
```

### 6.2 Inserting Records with Full-text Index

```go
// Insert records with full-text index
contentRecords := []map[string]any{
    {"id": 1, "name": "Product 1", "description": "This is a high-performance laptop suitable for programming and gaming"},
    {"id": 2, "name": "Product 2", "description": "Smartphone with powerful camera and long-lasting battery"},
    {"id": 3, "name": "Product 3", "description": "Wireless earphones providing immersive audio experience"},
    {"id": 4, "name": "Product 4", "description": "Smartwatch that can monitor health data and receive notifications"},
}

for _, record := range contentRecords {
    _, err = table.Insert(&record)
    if err != nil {
        panic(err)
    }
}
```

### 6.3 Executing Full-text Search

```go
// Full-text search example
fmt.Println("=== Full-text Search Example ===")

// Search for records containing "laptop"
fmt.Println("\n1. Search for 'laptop':")
search1 := map[string]any{"description": "laptop"}
iter1 := table.Search(&search1)
defer iter1.Release()
records1 := iter1.GetRecords(true)
for _, record := range records1 {
    fmt.Printf("   - %s: %s\n", record["name"], record["description"])
}

// Search for records containing "smart"
fmt.Println("\n2. Search for 'smart':")
search2 := map[string]any{"description": "smart"}
iter2 := table.Search(&search2)
defer iter2.Release()
records2 := iter2.GetRecords(true)
for _, record := range records2 {
    fmt.Printf("   - %s: %s\n", record["name"], record["description"])
}

// 3. Search result field selection example
fmt.Println("\n3. Search result field selection:")
search3 := map[string]any{"description": "smart"}
iter3 := table.Search(&search3)
defer iter3.Release()
records3 := iter3.GetRecords(true)

// Use Select method to only select name field
selectedNames := records3.Select("name")
fmt.Println("Only show names of matching records:")
for _, record := range selectedNames {
    fmt.Printf("   - %s\n", record["name"])
}
```

## 7. Querying Data

### 7.1 Basic Search

```go
// Basic search: exact match
searchFields := map[string]any{
    "name": "Zhang San",
}
iter := table.Search(&searchFields)
defer iter.Release()

// Get all matching records
records := iter.GetRecords(true)
for _, record := range records {
    fmt.Printf("Found record: %v\n", record)
}
```

### 7.2 Using Comparison Operators

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

### 7.3 Comparison Operator Usage Examples

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
defer iterGt30.Release()
recordsGt30 := iterGt30.GetRecords(true)
for _, record := range recordsGt30 {
    fmt.Printf("   - %s: %d years old\n", record["name"], record["age"])
}

// 2. Prefix search (Like operator) - default is Like operation
fmt.Println("\nUsers with email starting with 'user':")
emailPrefix := map[string]any{
    "email": "user",
}
iterPrefix := table.Search(&emailPrefix) // Default uses util.Like operator, here like is actually prefix match
defer iterPrefix.Release()
recordsPrefix := iterPrefix.GetRecords(true)
for _, record := range recordsPrefix {
    fmt.Printf("   - %s: %s\n", record["name"], record["email"])
}

// 3. Explicitly use Like operator
fmt.Println("\nUsers with name starting with 'Zhang':")
namePrefix := map[string]any{
    "name": "Zhang",
}
iterName := table.Search(&namePrefix, util.Like) // Explicitly specify util.Like operator
defer iterName.Release()
recordsName := iterName.GetRecords(true)
for _, record := range recordsName {
    fmt.Printf("   - %s: %s\n", record["name"], record["email"])
}

// 4. Exact match search
fmt.Println("\nExact search for user named 'Zhang San':")
exactSearch := map[string]any{
    "name": "Zhang San",
}
iterExact := table.Search(&exactSearch, util.Equal) // Explicitly specify util.Equal operator
defer iterExact.Release()
recordsExact := iterExact.GetRecords(true)
for _, record := range recordsExact {
    fmt.Printf("   - %s: %s\n", record["name"], record["email"])
}

// 5. Not equal search
fmt.Println("\nUsers with id not equal to 1:")
notEqualSearch := map[string]any{
    "id": 1,
}
iterNotEqual := table.Search(&notEqualSearch, util.NotEqual) // Use util.NotEqual operator
defer iterNotEqual.Release()
recordsNotEqual := iterNotEqual.GetRecords(true)
for _, record := range recordsNotEqual {
    fmt.Printf("   - %s: ID=%d\n", record["name"], record["id"])
}
```

### 7.4 Matcher Interface

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

### 7.5 AND Matcher

#### 7.5.1 AND Matcher Overview

The `AND` matcher is used to determine if a record's field values are in a specified data set, supporting positive matching (IN) and negative matching (NOT IN).

**Core functions**:
- Implement SQL-like IN and NOT IN operations
- Support multi-field combination matching
- Suitable for multi-table join queries
- Support combination with other matchers

#### 7.5.2 AND Struct Definition

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

#### 7.5.3 Constructor

```go
func NewAND(fields []string, data map[any]bool, rule ...bool) *AND
```

**Parameter descriptions**:
- `fields`: List of field names to match
- `data`: Matching data set
- `rule`: Optional parameter, matching rule, default is `true`

#### 7.5.4 Usage Examples

**Example 1: Basic IN Operation**

```go
// Suppose we have a user table and need to query users with ID in the specified set

// 1. Get ID set
idMap := map[any]bool{1: true, 3: true, 5: true}

// 2. Create AND matcher
andMatcher := match.NewAND([]string{"id"}, idMap)

// 3. Use matcher
iter := table.Search(&map[string]any{"id": nil})
defer iter.Release()

iter.SetMatch(andMatcher)
records := iter.GetRecords(true)

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
defer iter.Release()

iter.SetMatch(andMatcher)
records := iter.GetRecords(true)

// Result: Returns users with ID not 1, 3, 5
```

**Example 3: Multi-table Join Query (Key Example)**

```go
// Implement SQL-like join query: SELECT table1.* FROM table1, table2 WHERE table1.id = table2.id

// 1. Get iterators for both tables
iter1 := table1.Search(&map[string]any{"id": nil})
defer iter1.Release()

iter2 := table2.Search(&map[string]any{"id": nil})
defer iter2.Release()

// 2. Get ID mapping from table2
// Map() method generates map[any]bool, keys are values of specified fields
idMap := iter2.Map()

// 3. Create AND matcher
// Match if table1's id field is in table2's id set
andMatcher := match.NewAND([]string{"id"}, idMap)

// 4. Set matcher and get results
iter1.SetMatch(andMatcher)
records := iter1.GetRecords(true)

// Result: Returns records from table1 where ID matches ID in table2
```

**Example 4: Multi-table Join NOT Operation**

```go
// Implement SQL-like join query: SELECT table1.* FROM table1, table2 WHERE table1.id != table2.id

// 1. Get iterators for both tables
iter1 := table1.Search(&map[string]any{"id": nil})
defer iter1.Release()

iter2 := table2.Search(&map[string]any{"id": nil})
defer iter2.Release()

// 2. Get ID mapping from table2
idMap := iter2.Map()

// 3. Create AND matcher, set rule=false for NOT IN
andMatcher := match.NewAND([]string{"id"}, idMap, false)

// 4. Set matcher and get results
iter1.SetMatch(andMatcher)
records := iter1.GetRecords(true)

// Result: Returns records from table1 where ID does not match ID in table2
```

#### 7.5.5 Composite Primary Key Matching

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
// The purpose of this structure is to match primary key values of multiple iterators for same or different matching
// When rule is true, primary key values of multiple iterators must be the same to match successfully
// When rule is false, primary key values of multiple iterators must be different to match successfully
// The purpose of rule=false is mainly used in jump queries, such as field not in (1,2,3) in SQL statements, then data=map[any]bool{1:true,2:true,3:true}
*/
```

#### 7.5.6 Combination with Other Matchers

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
defer iter.Release()

// Set multiple matchers, they have an AND relationship
iter.SetMatch(idMatcher, ageMatcher)
records := iter.GetRecords(true)

// Result: Returns users with ID 1, 3, 5 and age greater than 25
```

#### 7.5.7 Advantages of AND Matcher

1. **Efficient multi-table joins**: Avoid nested loops through precomputed mapping tables, improving join query efficiency
2. **Flexible matching rules**: Support IN, NOT IN, AND, OR and other matching methods
3. **Support multi-field combinations**: Can match based on combined values of multiple fields
4. **Easy to combine with other matchers**: Can be used in combination with other custom or built-in matchers
5. **Suitable for complex query scenarios**: Especially suitable for query scenarios that need to relate multiple tables or sets

Through the `AND` matcher, sfsDb implements efficient and flexible multi-table join query functionality, providing users with powerful data query capabilities.

### 7.5 FieldComparison Matcher

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
    defer iter.Release()

    // Create FieldComparison matcher
    matcher := match.NewFieldComparison("age", match.GreaterThan, 25)

    // Set matcher to iterator
    iter.SetMatch(matcher)

    // Get filtered records
    records := iter.GetRecords(true)
    fmt.Printf("Records with age greater than 25 (%d records):\n", len(records))
    for _, record := range records {
        fmt.Printf("   - %v\n", record)
    }

    // 2. Using convenience functions to create matchers
    iter2 := table.Search(&map[string]any{"id": nil})
    defer iter2.Release()

    // Using GreaterThanMatch convenience function
    highScoreMatcher := match.NewGreaterThanMatch("score", 90.0)
    iter2.SetMatch(highScoreMatcher)

    highScoreRecords := iter2.GetRecords(true)
    fmt.Printf("\nRecords with score greater than 90 (%d records):\n", len(highScoreRecords))
    for _, record := range highScoreRecords {
        fmt.Printf("   - %v\n", record)
    }

    // 3. Using EqualMatch convenience function
    iter3 := table.Search(&map[string]any{"id": nil})
    defer iter3.Release()

    inactiveMatcher := match.NewEqualMatch("active", false)
    iter3.SetMatch(inactiveMatcher)

    inactiveRecords := iter3.GetRecords(true)
    fmt.Printf("\nInactive users (%d records):\n", len(inactiveRecords))
    for _, record := range inactiveRecords {
        fmt.Printf("   - %v\n", record)
    }
}
```

### 7.6 Comparison Operations Supported by FieldComparison

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

## 8. Field Modification

### 8.1 Single Field Modification

```go
// Field modification workflow: Rename "position" field to "job_title"

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

### 8.2 Batch Field Modification

```go
// Batch modify multiple fields

// 1. Successively call UpdateFieldName to update multiple field name mappings
err = table.UpdateFieldName("old_field1", "new_field1")
if err != nil {
    panic(err)
}
err = table.UpdateFieldName("old_field2", "new_field2")
if err != nil {
    panic(err)
}
err = table.UpdateFieldName("old_field3", "new_field3")
if err != nil {
    panic(err)
}

// 2. Call SetFields to update complete field mapping
newFields := map[string]any{
    "id":         0,
    "name":       "",
    "new_field1": "",
    "new_field2": 0,
    "new_field3": 0.0,
}
err = table.SetFields(newFields)
if err != nil {
    panic(err)
}
```

## 9. Deleting Records

### 9.1 Single Record Deletion

```go
// Delete single record
recordToDelete := map[string]any{
    "id": 1, // Must include primary key
}
err = table.Delete(&recordToDelete)
if err != nil {
    panic(err)
}
fmt.Println("Record deleted successfully")
```

### 9.2 Batch Deletion

```go
// Batch delete records: Delete users younger than 25

// First search for records matching the criteria
searchCriteria := map[string]any{
    "age": map[util.ComparisonOperator]any{util.LessThan: 25},
}
iter := table.Search(&searchCriteria)
defer iter.Release()

// Use iterator's Delete method to batch delete matching records
// Directly delete all matching records (unlimited)
iter.Delete()

// Or limit deletion quantity: Delete first 2 matching records
// iter.Delete(2)
```

## 10. Other Functions

### 10.1 Record Set Operations

#### 10.1.1 Dynamic Operation Functions

```go
// RecordsOperation defines record operation function type
type RecordsOperation func(rs Records, other ...Records) Records

// Apply method for dynamically applying operation functions to record sets
func (rs Records) Apply(op RecordsOperation, other ...Records) Records {
    return op(rs, other...)
}
```

**Core functions**:
- Allow users to customize set operations based on needs
- Support combining multiple operations
- Provide more flexible set operation methods
- Can be used in combination with existing set operations (Intersect, Union, Difference)

**Usage scenarios**:
- When custom set operation logic is needed
- When multiple set operations need to be combined
- When set operations need to be dynamically selected

**Basic usage example**:
```go
// Define custom operation function
customOp := record.RecordsOperation(func(rs record.Records, other ...record.Records) record.Records {
    // Custom operation logic
    result := make(record.Records, 0)
    for _, r := range rs {
        // Implement custom filtering logic
        if age, ok := r["age"].(int); ok && age > 30 {
            result = append(result, r)
        }
    }
    return result
})

// Apply operation function
result := records1.Apply(customOp)
```

### 10.2 Iterator Usage

**TableIter structure creation methods**:
- TableIter structure is generated by the table's `ForData()` method or `Search()` method
- `Search()` method is located at `d:\MyGo\src\sfsDb\engine\tableCRUD.go#L283`, used to search records based on conditions and return an iterator
- `ForData()` method is used to traverse all records in the table and return an iterator

```go
// Traverse all records
iter := table.ForData()
defer iter.Release()

// Recommended method: Use GetRecords() // Get all records (ascending)
records := iter.GetRecords(true)
for _, record := range records {
    fmt.Printf("Record: %v\n", record)
}

// Get all records (descending)
recordsReverse := iter.GetRecords(false)
for _, record := range recordsReverse {
    fmt.Printf("Record (reverse): %v\n", record)
}

// Pagination to get records
// Get page 2, 10 records per page (ascending)
page2Records := iter.GetRecords(true, 10, 10) // skip=10, limit=10
for _, record := range page2Records {
    fmt.Printf("Page 2 record: %v\n", record)
}

// Get page 2, 10 records per page (descending)
page2Reverse := iter.GetRecords(false, 10, 10) // skip=10, limit=10
for _, record := range page2Reverse {
    fmt.Printf("Page 2 record (reverse): %v\n", record)
}

// TopN to get records
// Test Top3: Get first 3 records (ascending)
fmt.Println("\nTop3 records (ascending):")
top3Records := iter.GetRecords(true, 3) // true means ascending, only get first 3 records
for _, record := range top3Records {
    fmt.Printf("   - %v\n", record)
}

// Test Top3: Get last 3 records (descending)
fmt.Println("\nTop3 records (descending):")
top3Reverse := iter.GetRecords(false, 3) // false means descending, only get last 3 records
for _, record := range top3Reverse {
    fmt.Printf("   - %v\n", record)
}

// Test Top5: Get first 5 records (ascending)
fmt.Println("\nTop5 records (ascending):")
top5Records := iter.GetRecords(true, 5) // true means ascending, only get first 5 records
for _, record := range top5Records {
    fmt.Printf("   - %v\n", record)
}

// Field selection function example
// 1. Use Records.Select() to select specific fields
fmt.Println("\nField selection example:")
allRecords := iter.GetRecords(true)

// Select only name and age fields
selectedFields := allRecords.Select("name", "age")
fmt.Println("Only show name and age fields:")
for _, record := range selectedFields {
    fmt.Printf("   - %v\n", record)
}

// 2. Use single Record.Select() to select fields
fmt.Println("\nSingle record field selection:")
if len(allRecords) > 0 {
    firstRecord := allRecords[0]
    partialRecord := firstRecord.Select("name", "score")
    fmt.Printf("First record's name and score fields: %v\n", partialRecord)
}

// 3. Use TableIter.SetSelects() to set selection fields at iterator level
fmt.Println("\nIterator-level field selection:")
iter3 := table.ForData()
defer iter3.Release()

// Set iterator to only return specified fields
iter3.SetSelects("name", "active")

// When getting records, only specified fields will be included
selectedRecords := iter3.GetRecords(true)
for _, record := range selectedRecords {
    fmt.Printf("   - %v\n", record)
}

// Test Top5: Get last 5 records (descending)
fmt.Println("\nTop5 records (descending):")
top5Reverse := iter.GetRecords(false, 5) // false means descending, only get last 5 records
for _, record := range top5Reverse {
    fmt.Printf("   - %v\n", record)
}

// Test Top15: Get all records (ascending)
fmt.Println("\nAll records (ascending):")
allRecords := iter.GetRecords(true, 15) // When limit exceeds total records, return all records
fmt.Printf("Total %d records obtained\n", len(allRecords))
for _, record := range allRecords {
    fmt.Printf("   - %v\n", record)
}

// Test Top15: Get all records (descending)
fmt.Println("\nAll records (descending):")
allReverse := iter.GetRecords(false, 15) // When limit exceeds total records, return all records
fmt.Printf("Total %d records obtained\n", len(allReverse))
for _, record := range allReverse {
    fmt.Printf("   - %v\n", record)
}

// Traditional method: Use First(), Valid(), Next() to traverse (not recommended, for reference only)
// for iter.First(); iter.Valid(); iter.Next() {
//     record := table.ParseRecordValue(iter.Value())
//     fmt.Printf("Record: %v\n", record)
// }
```

### 10.2 Record Update

#### 10.2.1 Basic Update

```go
// Update single record
updateRecord := map[string]any{
    "id":   1,      // Must include primary key
    "name": "Updated Zhang San", // Fields to update
    "age":  31,     // Fields to update
}
err = table.Update(&updateRecord)
if err != nil {
    panic(err)
}
fmt.Println("Record updated successfully")
```

#### 10.2.2 Optimistic Lock Mechanism

sfsDb has a built-in optimistic lock mechanism for handling concurrent update conflicts. Each table automatically adds a `v` field (int type) as a version number, which is automatically incremented each time the record is updated.

**Optimistic lock working principle**:
1. When reading a record, the current version number is obtained
2. When updating a record, you can choose to pass the expected version number
3. If the passed version number does not match the current version number, an optimistic lock conflict error is returned
4. If no version number is passed, the record is directly updated and the version number is incremented

**Version field description**:
- Field name: `v`
- Type: int
- Default value: 1
- Auto-increment: Automatically +1 each time the record is updated

#### 10.2.3 Specifying Version Number for Update

```go
// 1. First read the record to get current version number
readFields := map[string]any{"id": 1}
recordData, err := table.Read(&readFields)
if err != nil {
    panic(err)
}

// Assume the read record contains version number v=1
fmt.Println("Current record version number:", recordData["v"])

// 2. Use the obtained version number for update
updateRecord := map[string]any{
    "id":   1,      // Must include primary key
    "name": "Updated Zhang San", // Fields to update
    "v":    1,      // Specify expected version number
}

// 3. Execute update
// If another transaction updates the record during this period, the version number will change and this update will fail
err = table.Update(&updateRecord)
if err != nil {
    // Check if it's an optimistic lock conflict
    if strings.Contains(err.Error(), "optimistic lock conflict") {
        fmt.Println("Optimistic lock conflict, record has been updated by another transaction")
        // You can choose a retry strategy: re-read the record, get the latest version number and try updating again
    } else {
        panic(err)
    }
}

// After successful update, version number will be automatically incremented to 2
fmt.Println("Record updated successfully, new version number:", 2)
```

#### 10.2.4 Optimistic Lock Conflict Handling Example

```go
// Typical optimistic lock conflict handling process
func updateWithRetry(table *engine.Table, id int, updateFunc func(map[string]any)) error {
    maxRetries := 3
    
    for i := 0; i < maxRetries; i++ {
        // 1. Read record to get current data and version number
        readFields := map[string]any{"id": id}
        recordData, err := table.Read(&readFields)
        if err != nil {
            return err
        }
        
        // 2. Prepare update data, including current version number
        updateData := map[string]any{"id": id}
        
        // Copy existing fields to update data
        for k, v := range recordData {
            updateData[k] = v
        }
        
        // 3. Apply update function
        updateFunc(updateData)
        
        // 4. Execute update
        err = table.Update(&updateData)
        if err == nil {
            // Update successful
            return nil
        } else if strings.Contains(err.Error(), "optimistic lock conflict") {
            // Optimistic lock conflict, retry
            fmt.Printf("Optimistic lock conflict, retry %d...\n", i+1)
            continue
        } else {
            // Other errors
            return err
        }
    }
    
    return fmt.Errorf("Update failed, maximum retry attempts reached")
}

// Usage example
updateFunc := func(data map[string]any) {
    data["name"] = "Final Updated Zhang San"
    data["age"] = data["age"].(int) + 1
}

err = updateWithRetry(table, 1, updateFunc)
if err != nil {
    panic(err)
}
fmt.Println("Record updated successfully")
```

#### 10.2.5 Update Without Specifying Version Number

If no version number is specified, the system will directly update the record and increment the version number without optimistic lock checking:

```go
// Update without specifying version number, direct update
updateRecord := map[string]any{
    "id":   1,      // Must include primary key
    "name": "Updated Zhang San", // Fields to update
    // No v field specified
}

// Will directly update the record and automatically increment the version number
err = table.Update(&updateRecord)
if err != nil {
    panic(err)
}
fmt.Println("Record updated successfully")
```

**Usage suggestions**:
- In high concurrency environments, it is recommended to use the update method with specified version numbers to avoid lost updates
- In low concurrency environments, you can use the update method without specifying version numbers to simplify code
- When encountering optimistic lock conflicts, it is recommended to implement a retry mechanism or prompt the user to re-operate
- The optimistic lock mechanism is suitable for read-heavy scenarios

### 10.3 Batch Update Records

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/util"
)

func main() {
    // Create or get table
    table, err := engine.TableNew("users")
    if err != nil {
        panic(err)
    }

    // Search criteria: users younger than 25
    searchCriteria := map[string]any{
        "age": map[util.ComparisonOperator]any{util.LessThan: 25},
    }
    iter := table.Search(&searchCriteria)
    defer iter.Release()

    // Prepare fields to update
    updateFields := map[string]any{
        "active": false, // Set young users to inactive status
    }

    // Use iterator's Update method to batch update all matching records
    iter.Update(&updateFields)

    // Or limit update quantity: Only update first 2 matching records
    // iter.Update(&updateFields, 2)

    fmt.Println("Batch update completed")
}
```

### 10.4 Database Backup

sfsDb provides database backup functionality, which can back up all records of the current database to a specified path:

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // 1. Open source database (if not already open)
    sourceDb, err := storage.OpenDefaultDb("./source_db")
    if err != nil {
        panic(err)
    }
    defer storage.CloseDb()

    // 2. Execute database backup
    backupPath := "./backup_db"
    err = storage.BackupDb(backupPath)
    if err != nil {
        panic(fmt.Sprintf("Database backup failed: %v", err))
    }

    fmt.Printf("Database backup successful, backup path: %s\n", backupPath)
}
```

**Function description**:

```go
// Backup database
func BackupDb(Path string) error
// Parameters:
//   Path: Backup database save path
// Return value:
//   error: Error that occurred during backup, returns nil if successful
```

**Usage notes**:

1. The backup operation will copy all records from the source database to the target path
2. The backup process will not modify the state of the source database
3. After backup is completed, the target path will contain a complete database copy
4. You can use `storage.OpenDefaultDb(backupPath)` to open the backed-up database for recovery or query
5. The backup database is a complete database instance that can be used independently

### 10.5 Data Import and Export

sfsDb supports exporting table data to CSV, JSON, and SQL formats, and importing data from CSV and JSON formats.

#### 10.5.1 Export Data Example

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // Open database
    _, err := storage.OpenDefaultDb("./test_db")
    if err != nil {
        panic(err)
    }
    defer storage.CloseDb()

    // Get table
    table, err := engine.TableNew("test_table")
    if err != nil {
        panic(err)
    }

    // Set table fields
    fields := map[string]any{
        "id":     0,
        "name":   "",
        "age":    0,
        "active": false,
        "score":  0.0,
    }
    err = table.SetFields(fields)
    if err != nil {
        panic(err)
    }

    // Insert test data
    testData := []map[string]any{
        {"id": 1, "name": "Alice", "age": 25, "active": true, "score": 85.5},
        {"id": 2, "name": "Bob", "age": 30, "active": false, "score": 90.0},
        {"id": 3, "name": "Charlie", "age": 35, "active": true, "score": 75.5},
    }

    for _, data := range testData {
        _, err := table.Insert(&data)
        if err != nil {
            panic(err)
        }
    }

    // Export to CSV
    err = table.ExportToCSV("./test.csv")
    if err != nil {
        panic(err)
    }
    fmt.Println("CSV export successful")

    // Export to JSON
    err = table.ExportToJSON("./test.json")
    if err != nil {
        panic(err)
    }
    fmt.Println("JSON export successful")

    // Export to SQL
    err = table.ExportToSQL("./test.sql")
    if err != nil {
        panic(err)
    }
    fmt.Println("SQL export successful")
}
```

#### 10.5.2 Import Data Example

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // Open database
    _, err := storage.OpenDefaultDb("./test_db")
    if err != nil {
        panic(err)
    }
    defer storage.CloseDb()

    // Create new table for importing data
    table2, err := engine.TableNew("test_table2")
    if err != nil {
        panic(err)
    }

    // Set table fields (same as export table)
    fields := map[string]any{
        "id":     0,
        "name":   "",
        "age":    0,
        "active": false,
        "score":  0.0,
    }
    err = table2.SetFields(fields)
    if err != nil {
        panic(err)
    }

    // Import data from CSV (batch size 100)
    err = table2.ImportFromCSV("./test.csv", 100)
    if err != nil {
        panic(err)
    }
    fmt.Println("Import from CSV successful")

    // Create another table for importing JSON data
    table3, err := engine.TableNew("test_table3")
    if err != nil {
        panic(err)
    }

    // Set table fields (same as export table)
    err = table3.SetFields(fields)
    if err != nil {
        panic(err)
    }

    // Import data from JSON (batch size 100)
    err = table3.ImportFromJSON("./test.json", 100)
    if err != nil {
        panic(err)
    }
    fmt.Println("Import from JSON successful")
}
```

#### 10.5.3 Function Description

**Export functions**:

```go
// Export table data to CSV format
func (t *Table) ExportToCSV(filePath string) error
// Parameters:
//   filePath: Export file path
// Return value:
//   error: Error that occurred during export, returns nil if successful

// Export table data to JSON format
func (t *Table) ExportToJSON(filePath string) error
// Parameters:
//   filePath: Export file path
// Return value:
//   error: Error that occurred during export, returns nil if successful

// Export table data to SQL format
func (t *Table) ExportToSQL(filePath string) error
// Parameters:
//   filePath: Export file path
// Return value:
//   error: Error that occurred during export, returns nil if successful
```

**Import functions**:

```go
// Import data from CSV file to table
func (t *Table) ImportFromCSV(filePath string, batchSize int) error
// Parameters:
//   filePath: Import file path
//   batchSize: Batch import size, used for performance optimization
// Return value:
//   error: Error that occurred during import, returns nil if successful

// Import data from JSON file to table
func (t *Table) ImportFromJSON(filePath string, batchSize int) error
// Parameters:
//   filePath: Import file path
//   batchSize: Batch import size, used for performance optimization
// Return value:
//   error: Error that occurred during import, returns nil if successful
```

#### 10.5.4 Usage Notes

1. **Data format requirements**:
   - CSV files must contain headers, and header field names must match table field names
   - JSON files must be arrays containing objects, and object keys must match table field names
   - SQL files contain create table statements and insert statements, which can be used to initialize new tables

2. **Type conversion**:
   - Type conversion is automatically performed during import to ensure data types match table definitions
   - Numbers in JSON are automatically converted to int, float32, etc. types according to table definitions
   - Empty values in CSV are converted to table field default values

3. **Performance optimization**:
   - When importing large amounts of data, it is recommended to adjust the batchSize parameter to optimize performance
   - Larger batchSize can improve import speed but will increase memory usage
   - Default batchSize is 100

4. **Data consistency**:
   - Type checking is performed during import to ensure data conforms to table definitions
   - Detailed error information is returned when import fails, including record index and error reason
   - It is recommended to back up original data before importing

### 10.6 Query Performance Tracking

sfsDb has built-in query performance tracking functionality, which can help developers understand the execution situation of queries, including query times, total time consumption, and other indicators, facilitating performance optimization and problem diagnosis.

### 10.7 Data Encryption

sfsDb provides powerful data encryption functionality, which can protect the security of data stored on disk. The encryption functionality uses the AES-256-GCM algorithm, providing flexible key management methods and good performance.

#### 10.7.1 Encryption Overview

**Core features**:
- **Strong encryption algorithm**: Uses AES-256-GCM algorithm, providing 256-bit key protection
- **Transparent encryption**: Transparent to upper-layer applications, no need to modify existing code
- **Flexible key management**: Supports direct keys, password-derived keys, and key rotation
- **High performance**: Optimizes performance through decryption cache and batch operations
- **Optional encryption**: Can choose whether to enable encryption, or encrypt specific tables

**Encryption scope**:
- All data stored on disk
- Including table data, indexes, and metadata
- Database structure information (table names, field names) is not encrypted by default

#### 10.7.2 Encryption Configuration

```go
// EncryptionConfig encryption configuration
type EncryptionConfig struct {
    // Whether to enable encryption
    Enabled bool `json:"enabled"`
    
    // Encryption algorithm, default AES-256-GCM
    Algorithm string `json:"algorithm"`
    
    // Master key, directly provided 256-bit key
    MasterKey []byte `json:"master_key,omitempty"`
    
    // Password, used for key derivation
    Password string `json:"password,omitempty"`
    
    // Salt value, used for password derivation
    Salt []byte `json:"salt,omitempty"`
    
    // Iteration count, used for password derivation, default 100000
    Iterations int `json:"iterations,omitempty"`
}
```

**Configuration description**:
- `Enabled`: Whether to enable encryption, default false
- `Algorithm`: Encryption algorithm, currently only supports "AES-256-GCM", default value is empty string (automatically use AES-256-GCM)
- `MasterKey`: Directly provided 256-bit key, choose one with Password
- `Password`: Password used for key derivation, choose one with MasterKey
- `Salt`: Salt value used during password derivation, it is recommended to use randomly generated 16-byte salt value
- `Iterations`: PBKDF2 algorithm iteration count, default 100,000

#### 10.7.3 Usage Examples

##### 10.7.3.1 Using Direct Key Encryption

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // Generate or get 256-bit master key
    masterKey := make([]byte, 32) // 256-bit key
    // In actual applications, it is recommended to use cryptographically secure random number generator to generate keys
    for i := range masterKey {
        masterKey[i] = byte(i)
    }
    
    // Create encryption configuration
    encryptConfig := &storage.EncryptionConfig{
        Enabled:   true,
        Algorithm: "AES-256-GCM",
        MasterKey: masterKey,
    }
    
    // Open encrypted database
    db, err := storage.OpenDefaultDbWithEncryption("./encrypted_db", encryptConfig)
    if err != nil {
        panic(err)
    }
    defer storage.CloseDb()
    
    fmt.Println("Encrypted database opened successfully")
    
    // Subsequent operations are the same as regular databases
    // ...
}
```

##### 10.7.3.2 Using Password-derived Key

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // Create encryption configuration, using password-derived key
    encryptConfig := &storage.EncryptionConfig{
        Enabled:    true,
        Algorithm:  "AES-256-GCM",
        Password:   "my_secure_password",
        Salt:       []byte("my_random_salt_123"), // It is recommended to use randomly generated salt value
        Iterations: 100000,
    }
    
    // Open encrypted database
    db, err := storage.OpenDefaultDbWithEncryption("./encrypted_db", encryptConfig)
    if err != nil {
        panic(err)
    }
    defer storage.CloseDb()
    
    fmt.Println("Encrypted database with password-derived key opened successfully")
    
    // Subsequent operations are the same as regular databases
    // ...
}
```

##### 10.7.3.3 Encryption in Multi-instance Mode

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // Generate test key
    masterKey := make([]byte, 32)
    for i := range masterKey {
        masterKey[i] = byte(i)
    }
    
    // Create encryption configuration
    encryptConfig := &storage.EncryptionConfig{
        Enabled:   true,
        MasterKey: masterKey,
    }
    
    // Use NewLevelDBStoreWithEncryption to create independent encrypted database instance
    // Does not use global KVDb variable
    db1, err := storage.NewLevelDBStoreWithEncryption("./encrypted_db_1", encryptConfig)
    if err != nil {
        panic(err)
    }
    defer db1.Close()
    
    // Can create multiple independent encrypted database instances
    db2, err := storage.NewLevelDBStoreWithEncryption("./encrypted_db_2", encryptConfig)
    if err != nil {
        panic(err)
    }
    defer db2.Close()
    
    fmt.Println("Multiple encrypted database instances created successfully")
    
    // Subsequent operations are the same as regular databases
    // ...
}
```

#### 10.7.4 API Reference

**Core APIs**:

```go
// OpenDefaultDbWithEncryption opens default database with encryption
// Uses global KVDb variable to ensure global unique instance
func OpenDefaultDbWithEncryption(Path string, config *EncryptionConfig) (Store, error)

// NewLevelDBStoreWithEncryption creates LevelDB storage instance with encryption
// Creates new instance each time it is called, does not use global variable
func NewLevelDBStoreWithEncryption(Path string, config *EncryptionConfig) (Store, error)

// ReEncrypt re-encrypts all data (key rotation)
// Only EncryptedStoreWrapper implements this method
func (es *EncryptedStoreWrapper) ReEncrypt(newKey []byte) error
```

#### 10.7.5 Key Rotation

Key rotation is an important security practice that can regularly replace encryption keys to reduce the risk of key leakage.

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/storage"
)

func main() {
    // Initial key
    oldKey := make([]byte, 32)
    for i := range oldKey {
        oldKey[i] = byte(i)
    }
    
    // Open encrypted database
    encryptConfig := &storage.EncryptionConfig{
        Enabled:   true,
        MasterKey: oldKey,
    }
    db, err := storage.OpenDefaultDbWithEncryption("./encrypted_db", encryptConfig)
    if err != nil {
        panic(err)
    }
    defer storage.CloseDb()
    
    // Store some data
    // ...
    
    // Generate new key
    newKey := make([]byte, 32)
    for i := range newKey {
        newKey[i] = byte(255 - i)
    }
    
    // Execute key rotation
    encryptedWrapper, ok := db.(*storage.EncryptedStoreWrapper)
    if !ok {
        panic("Failed to convert to EncryptedStoreWrapper")
    }
    
    err = encryptedWrapper.ReEncrypt(newKey)
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Key rotation successful")
    
    // After key rotation, all data will be encrypted with the new key
    // Subsequent operations are the same as regular databases
    // ...
}
```

#### 10.7.6 Performance Considerations

- **Encryption overhead**: AES-256-GCM encryption and decryption are very fast,预计增加约 5-15% 的 CPU 开销
- **Storage overhead**: Each record increases by about 16 bytes (authentication tag)
- **Cache optimization**: Implements decryption cache to reduce repeated decryption operations
- **Batch operations**: Supports batch encryption/decryption to optimize batch operation performance
- **Configuration suggestions**:
  - For high concurrency scenarios, it is recommended to use direct keys instead of password derivation
  -适当调整解密缓存大小（默认 1000 条记录）
  - 对批量操作进行优化，减少加密解密次数

#### 10.7.7 Security Best Practices

1. **Key management**:
   - Do not hardcode keys in code
   - Consider using key management services (KMS) to store keys
   - Rotate keys regularly
   - Ensure the security and integrity of keys

2. **Password security**:
   - Use strong passwords (at least 16 characters, including uppercase and lowercase letters, numbers, and special characters)
   - Use randomly generated salt values
   -适当增加迭代次数（默认 100,000）

3. **Deployment security**:
   - Ensure database files have correct access permissions set
   - Consider using TLS to encrypt network transmission
   - Back up encrypted databases regularly
   - Implement access control mechanisms

By properly configuring and using sfsDb's encryption functionality, you can effectively protect data security and meet various security compliance requirements.

### 10.7 Key-Value Change Tracking

sfsDb has built-in key-value change tracking functionality, which can help developers monitor the increase and decrease of key-value pairs in the database, including the number of record insertion (put) and deletion (delete) operations, facilitating monitoring of database operation status and performance analysis.

#### 10.7.1 Tracking Data Structure

Key-value changes are tracked through two global variables:

```go
// Increment variable: Records the number of put operation key-value pairs
// key format: table_id,index_id (string composed of table ID and index ID)
var AtomicInt map[string]*atomic.Int64
 
// Decrement variable: Records the number of delete operation key-value pairs
// key format: table_id,index_id (string composed of table ID and index ID)
var AtomicDec map[string]*atomic.Int64
```

#### 10.7.2 Tracking Data Access

You can safely access and manipulate tracking data through methods of the `AtomicMap` type:

```go
// AtomicMap is a type for atomic operation key-value counters
// When using, you need to convert AtomicInt or AtomicDec to AtomicMap type

type AtomicMap map[string]*atomic.Int64

// Inc increments the counter for the specified key
func (m AtomicMap) Inc(tableID, indexID byte) {}

// Dec decrements the counter for the specified key
func (m AtomicMap) Dec(tableID, indexID byte) {}

// Get gets the current count value for the specified key
func (m AtomicMap) Get(tableID, indexID byte) int64 {}
```

#### 10.7.3 Usage Examples

**Example 1: Get table insertion and deletion operation counts**

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/monitor"
)

func main() {
    // Assume table ID is 0, index ID is 0
    tableID := byte(0)
    indexID := byte(0)
    
    // Get insertion operation count
    putCount := monitor.AtomicMap(monitor.AtomicInt).Get(tableID, indexID)
    fmt.Printf("Table %d index %d insertion operation count: %d\n", tableID, indexID, putCount)
    
    // Get deletion operation count
    deleteCount := monitor.AtomicMap(monitor.AtomicDec).Get(tableID, indexID)
    fmt.Printf("Table %d index %d deletion operation count: %d\n", tableID, indexID, deleteCount)
    
    // Calculate net change
    netChange := putCount - deleteCount
    fmt.Printf("Table %d index %d net change: %d\n", tableID, indexID, netChange)
}
```

**Example 2: Monitor multiple tables and indexes**

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/monitor"
)

func main() {
    // Define tables and indexes to monitor
    monitorConfigs := []struct {
        tableID  byte
        indexID  byte
        name     string
    }{
        {0, 0, "User Table - Primary Key Index"},
        {0, 1, "User Table - Name Index"},
        {1, 0, "Order Table - Primary Key Index"},
    }
    
    fmt.Println("Database key-value change monitoring report:")
    fmt.Println("==================================")
    
    for _, config := range monitorConfigs {
        putCount := monitor.AtomicMap(monitor.AtomicInt).Get(config.tableID, config.indexID)
        deleteCount := monitor.AtomicMap(monitor.AtomicDec).Get(config.tableID, config.indexID)
        netChange := putCount - deleteCount
        
        fmt.Printf("%s:\n", config.name)
        fmt.Printf("  - Insertion operations: %d\n", putCount)
        fmt.Printf("  - Deletion operations: %d\n", deleteCount)
        fmt.Printf("  - Net change: %d\n", netChange)
        fmt.Println()
    }
}
```

#### 10.7.4 Automatically Tracked Operations

sfsDb automatically updates key-value change tracking data in the following operations:

- **Insert records**: When using `table.Insert()` or `table.BatchInsert()`, automatically increment the insertion counter for the corresponding table and index
- **Delete records**: When using `table.Delete()` or `table.BatchDelete()`, automatically increment the deletion counter for the corresponding table and index
- **Update records**: When updating records, old records are first deleted and then new records are inserted, so both insertion and deletion counters are updated

#### 10.7.5 Usage Scenarios

1. **Database monitoring**: Real-time monitoring of database write and delete operations to understand database activity level
2. **Performance analysis**: Analyze operation frequency of different tables and indexes to identify hot data
3. **Capacity planning**: Predict database capacity growth based on key-value change trends
4. **Problem diagnosis**: When database performance is abnormal, check key-value change situation to locate problems
5. **Audit**: Record database operation counts for audit and compliance requirements

#### 10.7.6 Notes

1. **Concurrent security**: `AtomicInt` and `AtomicDec` internally use `atomic.Int64` to ensure concurrent security, but direct access to the map itself is not thread-safe. It is recommended to use methods of the `AtomicMap` type for operations
2. **Data persistence**: Key-value change tracking data only exists in memory and will be lost after application restart
3. **Performance overhead**: Tracking operations use atomic operations, with small performance overhead, but may have an impact in extremely high concurrency scenarios
4. **Key format**: Key format is `table_id,index_id`, where table_id and index_id are character representations of byte types
5. **Initialization**: Tracking data is automatically initialized when the application starts, no manual operation required
6. **Applicable scope**: Only tracks operations performed through sfsDb API, direct modifications to underlying storage will not be tracked

### 10.8 Record Operations

sfsDb has built-in record operation functionality, which can perform various operations on fields in record sets and add operation results as new fields to records.

#### 10.8.1 Records.Operation Method

**Method signature**:
```go
func (rs Records) Operation(op ...Operation) (rs2 Records) {
    if len(rs) == 0 {
        return nil
    }
    if len(op) == 0 {
        return rs
    }
    // Directly create result slice to avoid copying and then modifying
    result := make(Records, 0, len(rs))
    for _, r := range rs {
        // Create new copy for each record to avoid modifying original records
        newRecord := make(Record, len(r)+1)
        // Directly copy fields to avoid using maps.Copy
        for k, v := range r {
            newRecord[k] = v
        }
        // Detect if new field name already exists
        newField := op[0].NewField()
        if _, exists := newRecord[newField]; exists {
            panic(fmt.Sprintf("field '%s' already exists in record, cannot add duplicate field", newField))
        }
        // Add new field
        newRecord[newField] = op[0].Evaluate(r)
        // Add new record to result slice
        result = append(result, newRecord)
    }
    return result
}
```

**Function description**:
- Execute specified operation operations on each record in the record set
- Return new record set containing operation results, do not modify original records
- Support multiple operation operations (current version only processes the first one)
- Automatically detect if new field names already exist to avoid duplicates

**Parameter description**:
- `op`: Operation operation object, implements the `Operation` interface
- Through the `Operation` interface, custom operation operations can be defined, including field lists, new field names, and additional parameters
- Custom operation operations must implement `NewField` and `Evaluate` methods, which return new field names and operation results respectively
- Custom operation operations can use `args` parameter to pass additional parameters for calculation during operation

**Return value**:
- `rs2`: New record set containing operation results

#### 10.8.2 Built-in Operation Types

sfsDb has built-in multiple common operation types, implemented through the `CommonOperation` struct:

```go
type CommonOperation struct {
    opType   string // Operation type: add, sub, mul, div, avg, sum, max, min, concat
    fields   []string // Field list participating in operation
    newField string // New field name generated after operation
    args     map[string]any // Additional parameters
}
```

**Supported operation types**:

| Operation Type | Description | Example |
|---------------|-------------|---------|
| `add`/`sum` | Addition operation | `NewAddOperation([]string{"salary", "bonus"}, "total_income")` |
| `sub` | Subtraction operation | `NewSubOperation([]string{"salary", "bonus"}, "net_salary")` |
| `mul` | Multiplication operation | `NewMulOperation([]string{"price", "quantity"}, "total_price")` |
| `div` | Division operation | `NewDivOperation([]string{"salary", "days"}, "daily_salary", 1.0)` |
| `avg` | Average operation | `NewAvgOperation([]string{"score1", "score2"}, "avg_score")` |
| `max` | Maximum operation | `NewMaxOperation([]string{"field1", "field2"}, "max_value")` |
| `min` | Minimum operation | `NewMinOperation([]string{"field1", "field2"}, "min_value")` |
| `concat` | String concatenation | `NewConcatOperation([]string{"str1", "str2"}, "combined_str", " ")` |

#### 10.8.3 Convenience Creation Functions

sfsDb provides convenient functions to create various operation operations:

```go
// Create addition operation instance
func NewAddOperation(fields []string, newField string) *CommonOperation

// Create subtraction operation instance
func NewSubOperation(fields []string, newField string) *CommonOperation

// Create multiplication operation instance
func NewMulOperation(fields []string, newField string) *CommonOperation

// Create division operation instance
func NewDivOperation(fields []string, newField string, defaultDivisor float64) *CommonOperation

// Create average operation instance
func NewAvgOperation(fields []string, newField string) *CommonOperation

// Create sum operation instance
func NewSumOperation(fields []string, newField string) *CommonOperation

// Create maximum operation instance
func NewMaxOperation(fields []string, newField string) *CommonOperation

// Create minimum operation instance
func NewMinOperation(fields []string, newField string) *CommonOperation

// Create string concatenation operation instance
func NewConcatOperation(fields []string, newField string, separator string) *CommonOperation
```

#### 10.8.4 Usage Examples

**Example 1: Calculate employee total income**

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/record"
)

func main() {
    // Create table
    table, _ := engine.TableNew("employees")
    
    // Add some records
    record1 := map[string]any{"name": "Zhang San", "salary": 5000.0, "bonus": 1000.0}
    record2 := map[string]any{"name": "Li Si", "salary": 4000.0, "bonus": 800.0}
    table.Insert(&record1)
    table.Insert(&record2)
    
    // Query all records
    iter := table.Search(nil)
    records, _ := iter.GetRecords()
    
    // Create addition operation: salary + bonus
    op := record.NewAddOperation([]string{"salary", "bonus"}, "total_income")
    
    // Apply operation
    resultRecords := records.Operation(op)
    
    // Print results
    for _, r := range resultRecords {
        fmt.Printf("Name: %s, Total Income: %.2f\n", r["name"], r["total_income"])
    }
}
```

**Example 2: String concatenation**

```go
// Create string concatenation operation: str1 + str2
op := record.NewConcatOperation([]string{"first_name", "last_name"}, "full_name", " ")
resultRecords := records.Operation(op)
```

**Example 3: Multiple operations combined use**

```go
// First calculate total income
incomeOp := record.NewAddOperation([]string{"salary", "bonus"}, "total_income")
incomeRecords := records.Operation(incomeOp)

// Then calculate average salary
avgOp := record.NewAvgOperation([]string{"salary"}, "avg_salary")
avgRecords := incomeRecords.Operation(avgOp)
```

#### 10.8.6 Vertical Operation (OperationVertical) Method

**Method signature**:
```go
func (rs Records) OperationVertical(op ...VerticalOperation) (rs2 Records) {
    if len(rs) == 0 {
        return nil
    }
    if len(op) == 0 {
        return rs
    }
    // Implementation omitted
}
```

**Function description**:
- Execute vertical operations on record sets, i.e., operations on the same field of multiple records
- Return new record set containing operation results, do not modify original records
- Support multiple vertical operation operations
- Automatically detect if new field names already exist to avoid duplicates

**Parameter description**:
- `op`: Vertical operation operation object, implements the `VerticalOperation` interface
- Through the `VerticalOperation` interface, custom vertical operation operations can be defined
- Custom vertical operation operations must implement `NewField` and `Evaluate` methods

**Return value**:
- `rs2`: New record set containing vertical operation results

#### 10.8.7 Built-in Vertical Operation Types

sfsDb has built-in multiple common vertical operation types, implemented through the `CommonVerticalOperation` struct:

| Operation Type | Description | Example |
|---------------|-------------|---------|
| `sum` | Sum operation | `NewSumVerticalOperation("salary", "total_salary")` |
| `avg` | Average operation | `NewAvgVerticalOperation("salary", "avg_salary")` |
| `count` | Count operation | `NewCountVerticalOperation("salary", "record_count")` |
| `max` | Maximum operation | `NewMaxVerticalOperation("salary", "max_salary")` |
| `min` | Minimum operation | `NewMinVerticalOperation("salary", "min_salary")` |
| `group` | Group operation | `NewGroupVerticalOperation("department", "group_result")` |

#### 10.8.8 Vertical Operation Convenience Creation Functions

sfsDb provides convenient functions to create various vertical operation operations:

```go
// Create sum vertical operation instance
func NewSumVerticalOperation(field string, newField string) *CommonVerticalOperation

// Create average vertical operation instance
func NewAvgVerticalOperation(field string, newField string) *CommonVerticalOperation

// Create count vertical operation instance
func NewCountVerticalOperation(field string, newField string) *CommonVerticalOperation

// Create maximum vertical operation instance
func NewMaxVerticalOperation(field string, newField string) *CommonVerticalOperation

// Create minimum vertical operation instance
func NewMinVerticalOperation(field string, newField string) *CommonVerticalOperation

// Create group vertical operation instance
func NewGroupVerticalOperation(field string, newField string) *CommonVerticalOperation
```

#### 10.8.9 Vertical Operation Usage Examples

**Example 1: Sum operation**

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
    "github.com/liaoran123/sfsDb/record"
)

func main() {
    // Create table
    table, err := engine.TableNew("employees")
    if err != nil {
        panic(err)
    }
    
    // Set fields
    fields := map[string]any{
        "id":         0,
        "name":       "",
        "salary":     0.0,
        "department": "",
    }
    err = table.SetFields(fields)
    if err != nil {
        panic(err)
    }
    
    // Insert test data
    employees := []map[string]any{
        {"name": "Zhang San", "salary": 5000.0, "department": "Technical Department"},
        {"name": "Li Si", "salary": 4000.0, "department": "Technical Department"},
        {"name": "Wang Wu", "salary": 6000.0, "department": "Sales Department"},
        {"name": "Zhao Liu", "salary": 5500.0, "department": "Sales Department"},
    }
    
    for _, emp := range employees {
        _, err = table.Insert(&emp)
        if err != nil {
            panic(err)
        }
    }
    
    // Query all records
    iter := table.ForData()
    defer iter.Release()
    records := iter.GetRecords(true)
    
    // Create sum vertical operation: calculate total salary
    sumOp := record.NewSumVerticalOperation("salary", "total_salary")
    
    // Apply vertical operation
    result := records.OperationVertical(sumOp)
    
    // Print results
    fmt.Println("Employee salary information:")
    for _, r := range result {
        fmt.Printf("Name: %s, Department: %s, Salary: %.2f, Total Salary: %.2f\n", 
            r["name"], r["department"], r["salary"], r["total_salary"])
    }
}
```

**Example 2: Multiple vertical operations combined use**

```go
// Create multiple vertical operations
countOp := record.NewCountVerticalOperation("salary", "employee_count")
avgOp := record.NewAvgVerticalOperation("salary", "avg_salary")
maxOp := record.NewMaxVerticalOperation("salary", "max_salary")

// Apply multiple vertical operations
result := records.OperationVertical(countOp, avgOp, maxOp)
```

**Example 3: Group operation**

```go
// Create group vertical operation: group by department
groupOp := record.NewGroupVerticalOperation("department", "group_result")
result := records.OperationVertical(groupOp)

// Print group results
for _, r := range result {
    fmt.Printf("Name: %s\n", r["name"])
    fmt.Printf("Group result: %v\n", r["group_result"])
    fmt.Println("---")
}
```

#### 10.8.10 Vertical Operation Notes

1. **Cannot modify original records**: `OperationVertical` method does not modify original records, always returns new record sets
2. **Duplicate field detection**: Automatically detect if new field names already exist to avoid duplicates
3. **Vertical operation characteristics**: Vertical operations are operations on the same field of multiple records, results will be added to each record
4. **Performance considerations**: For large record sets, vertical operations may consume more memory and CPU
5. **Data types**: Vertical operations will automatically handle conversions between different numeric types
6. **Group operation results**: Group operation returns map[any]Records type, containing grouped record sets
7. **Empty record handling**: Executing vertical operations on empty record sets will return nil

#### 10.8.11 Record Operation Notes

1. **Cannot modify original records**: `Operation` and `OperationVertical` methods do not modify original records, always return new record sets
2. **Duplicate field detection**: Automatically detect if new field names already exist to avoid duplicates
3. **Operation order**: `Operation` method currently only processes the first operation operation, `OperationVertical` method supports multiple operation operations
4. **Performance considerations**: For large record sets, operation operations may consume more memory and CPU
5. **Data types**: Operation operations will automatically handle conversions between different numeric types
6. **Error handling**: When duplicate fields are detected, panic will be triggered, it is recommended to test during development
7. **Vertical vs horizontal operations**: `Operation` is horizontal operation (operations on multiple fields of a single record), `OperationVertical` is vertical operation (operations on the same field of multiple records)

### 10.9 Snapshot Functionality

### 10.10 Transaction Management

sfsDb provides complete transaction support, through the `Transaction` interface you can conveniently perform transaction operations. Transactions support ACID characteristics to ensure data consistency and reliability.

#### 10.10.1 Transaction Interface Introduction

```go
// Transaction defines transaction interface
type Transaction interface {
    // Insert inserts record in transaction
    Insert(fields *map[string]any) (int, error)
    // Update updates record in transaction
    Update(fields *map[string]any) error
    // Delete deletes record in transaction
    Delete(fields *map[string]any) error
    // Search searches records in transaction (supports read consistency)
    Search(fields *map[string]any, ops ...util.ComparisonOperator) *TableIter
    // Read reads single record in transaction (supports read consistency)
    Read(fields *map[string]any) ([]byte, error)
    // Commit commits transaction
    Commit() error
    // Rollback rolls back transaction
    Rollback() error
}
```

#### 10.10.2 Transaction Basic Usage Process

```go
// 1. Start transaction
tx, err := table.Begin()
if err != nil {
    panic(err)
}

try {
    // 2. Execute transaction operations (insert, update, delete, query)
    _, err = tx.Insert(&insertRecord)
    if err != nil {
        panic(err)
    }
    
    err = tx.Update(&updateRecord)
    if err != nil {
        panic(err)
    }
    
    // 3. Commit transaction
    err = tx.Commit()
    if err != nil {
        panic(err)
    }
    fmt.Println("Transaction committed successfully")
} catch {
    // 4. Rollback transaction when error occurs
    err = tx.Rollback()
    if err != nil {
        panic(err)
    }
    fmt.Println("Transaction rolled back successfully")
}
```

#### 10.10.3 Transaction Method Detailed Description

1. **`Insert(fields *map[string]any) (int, error)`**
   - Insert a record in transaction
   - Parameters: Mapping containing field names and values, must contain all required fields except auto-increment primary key
   - Return value: ID of inserted record and error information
   - Records inserted in transaction can be immediately read by other operations in transaction (read your own writes)

2. **`Update(fields *map[string]any) error`**
   - Update a record in transaction
   - Parameters: Mapping containing primary key and fields to update, must contain primary key
   - Return value: Error information
   - Updates made in transaction will immediately reflect in subsequent operations within transaction

3. **`Delete(fields *map[string]any) error`**
   - Delete a record in transaction
   - Parameters: Mapping containing primary key, must contain primary key
   - Return value: Error information
   - Records deleted in transaction can still be read by operations within transaction before transaction commit

4. **`Search(fields *map[string]any, ops ...util.ComparisonOperator) *TableIter`**
   - Search records in transaction
   - Parameters: Search criteria and comparison operators
   - Return value: Result iterator
   - Supports read consistency, reads data based on snapshot, not affected by external modifications

5. **`Read(fields *map[string]any) ([]byte, error)`**
   - Read single record in transaction
   - Parameters: Mapping containing primary key, must contain primary key
   - Return value: Byte array of record and error information
   - Priority reads from transaction cache, supports read your own writes

6. **`Commit() error`**
   - Commit transaction, persist all operations to database
   - Return value: Error information
   - After commit, all operations in transaction will take effect
   - After commit, transaction object can no longer be used

7. **`Rollback() error`**
   - Rollback transaction, cancel all operations
   - Return value: Error information
   - After rollback, all operations in transaction will not take effect
   - After rollback, transaction object can no longer be used

#### 10.10.4 Transaction Usage Examples

**Example 1: Basic transaction operation**

```go
package main

import (
    "fmt"
    "github.com/liaoran123/sfsDb/engine"
)

func main() {
    // Create or get table
    table, err := engine.TableNew("users")
    if err != nil {
        panic(err)
    }
    
    // Set table fields
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
    
    // Start transaction
    tx, err := table.Begin()
    if err != nil {
        panic(err)
    }
    
    // Insert record in transaction
    insertRecord := map[string]any{"name": "Zhang San", "age": 25, "email": "zhangsan@example.com"}
    id, err := tx.Insert(&insertRecord)
    if err != nil {
        tx.Rollback()
        panic(err)
    }
    fmt.Printf("Transaction insert record successful, ID: %d\n", id)
    
    // Read just inserted record in transaction
    readRecord := map[string]any{"id": id}
    recordData, err := tx.Read(&readRecord)
    if err != nil {
        tx.Rollback()
        panic(err)
    }
    fmt.Printf("Transaction read just inserted record: %v\n", recordData)
    
    // Update record in transaction
    updateRecord := map[string]any{"id": id, "age": 26}
    err = tx.Update(&updateRecord)
    if err != nil {
        tx.Rollback()
        panic(err)
    }
    fmt.Println("Transaction update record successful")
    
    // Commit transaction
    err = tx.Commit()
    if err != nil {
        panic(err)
    }
    fmt.Println("Transaction committed successfully")
}
```

**Example 2: Transaction rollback**

```go
// Start transaction
tx, err := table.Begin()
if err != nil {
    panic(err)
}

// Insert record in transaction
insertRecord := map[string]any{"name": "Test Rollback", "age": 30, "email": "test@example.com"}
_, err = tx.Insert(&insertRecord)
if err != nil {
    tx.Rollback()
    panic(err)
}

// Read just inserted record in transaction
readRecord := map[string]any{"name": "Test Rollback"}
iter := tx.Search(&readRecord)
defer iter.Release()
records := iter.GetRecords(true)
fmt.Printf("Transaction read %d records\n", len(records))

// Rollback transaction
err = tx.Rollback()
if err != nil {
    panic(err)
}
fmt.Println("Transaction rolled back successfully")

// Verify if records were rolled back
readRecordAfterRollback := map[string]any{"name": "Test Rollback"}
iterAfterRollback := table.Search(&readRecordAfterRollback)
defer iterAfterRollback.Release()
recordsAfterRollback := iterAfterRollback.GetRecords(true)
fmt.Printf("Read %d records after rollback\n", len(recordsAfterRollback))
```

**Example 3: Transaction read consistency**

```go
// Start transaction
tx, err := table.Begin()
if err != nil {
    panic(err)
}

// Read initial data in transaction
readRecord := map[string]any{"id": 1}
recordBeforeUpdate := readSingleRecord(tx, &readRecord)
fmt.Printf("Data read at transaction start: %v\n", recordBeforeUpdate)

// Update data outside transaction
updateRecordOutside := map[string]any{"id": 1, "age": 100}
err = table.Update(&updateRecordOutside)
if err != nil {
    tx.Rollback()
    panic(err)
}
fmt.Println("Data updated outside transaction successfully")

// Read data again in transaction, should still be initial data (read consistency)
recordInsideTx := readSingleRecord(tx, &readRecord)
fmt.Printf("Data read again in transaction: %v\n", recordInsideTx)

// Commit transaction
err = tx.Commit()
if err != nil {
    panic(err)
}

// Read data after transaction commit, should be updated data
recordAfterCommit := readSingleRecord(table, &readRecord)
fmt.Printf("Data read after transaction commit: %v\n", recordAfterCommit)
```

#### 10.10.4 Usage Suggestions

1. **Single table transaction**: Directly use transaction created by `table.Begin()`, operation is simple and convenient
2. **Cross-table transaction**: Currently need to manually manage batch, directly use underlying API:
   ```go
   // Get global batch
   batch := storage.KVDb.GetBatch()
   
   // Cross-table operations
   _, err := table1.Insert(&record1, batch)
   _, err := table2.Insert(&record2, batch)
   
   // Commit batch
   err = storage.KVDb.WriteBatch(batch)
   ```

#### 10.10.5 Transaction Notes

1. **Transaction lifecycle**: Transaction objects cannot be used after `Commit()` or `Rollback()`
2. **Read consistency**: Read operations within transaction are based on snapshots, not affected by external modifications
3. **Read your own writes**: Transactions can immediately read data you just wrote
4. **Concurrency control**:
   - Directly using underlying `batch` API supports concurrent write operations
   - Transactions created through `table.Begin()` method support concurrency, each transaction has its own independent snapshot and cache, transactions are completely isolated
   - Transactions created by `table.Begin()` can be safely used in multiple threads, they will not affect each other
   - Write operations of each transaction will eventually be written to database through atomic commit (`Commit()`), ensuring data consistency
5. **Performance considerations**: Long-held transactions will affect database performance, it is recommended to complete transaction operations as soon as possible
6. **Error handling**: Must handle errors that may occur in transaction operations and rollback transaction when errors occur
7. **Resource release**: Transaction objects will automatically release resources after use, no manual handling needed

### 10.9 Snapshot Functionality

sfsDb supports LevelDB snapshot functionality, which can create consistent read views of the database for reading consistent data or implementing read-write separation.

#### 10.9.1 Snapshot Functionality Introduction

LevelDB snapshots provide a consistent read view of the database at a certain point in time. Even if the database is modified after snapshot creation, the data in the snapshot will not change.

**Main features**:
- Consistent reading: After snapshot creation, regardless of how the database is modified, the data in the snapshot always remains consistent
- Read-write separation: Can perform read operations in snapshot mode without affecting write operations
- Efficient creation: Snapshot creation operation is very efficient, only requiring very little time and memory
- Automatic resource management: Snapshot resources will be automatically released when switching back to database mode

#### 10.9.2 Snapshot Related Methods

```go
// SwitchToSnapshot switches current instance to snapshot mode
// If currently already in snapshot mode, returns error
func (s *LevelDBStore) SwitchToSnapshot() error

// SwitchToDB switches current instance back to database mode
// If currently already in database mode, returns error
// If currently in snapshot mode, snapshot resources will be released first
func (s *LevelDBStore) SwitchToDB() error
```

#### 10.9.3 Usage Examples

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
    table, err := engine.TableNew("test_snapshot")
    if err != nil {
        panic(err)
    }

    // Set table fields
    fields := map[string]any{
        "id":   0,
        "name": "",
        "age":  0,
    }
    err = table.SetFields(fields)
    if err != nil {
        panic(err)
    }

    // Initial insert a record
    record1 := map[string]any{"id": 1, "name": "Initial Data", "age": 20}
    _, err = table.Insert(&record1)
    if err != nil {
        panic(err)
    }

    // Check if current kvStore is LevelDBStore type
    levelDBStore, ok := table.kvStore.(*storage.LevelDBStore)
    if !ok {
        fmt.Println("Current storage engine does not support snapshot functionality")
        return
    }

    // Create snapshot
    err = levelDBStore.SwitchToSnapshot()
    if err != nil {
        panic(err)
    }
    fmt.Println("Snapshot created successfully")

    // Read data in snapshot mode
    readRecord1 := map[string]any{"id": 1}
    snapshotResult, err := table.Read(&readRecord1)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Record read in snapshot mode: %v\n", snapshotResult)

    // Switch back to DB mode
    err = levelDBStore.SwitchToDB()
    if err != nil {
        panic(err)
    }
    fmt.Println("Switched back to DB mode successfully")

    // Update data
    updateRecord := map[string]any{"id": 1, "name": "Updated Data", "age": 25}
    err = table.Update(&updateRecord)
    if err != nil {
        panic(err)
    }
    fmt.Println("Record updated successfully")

    // Create snapshot again, this snapshot will contain updated data
    err = levelDBStore.SwitchToSnapshot()
    if err != nil {
        panic(err)
    }
    fmt.Println("Snapshot created again successfully")

    // Read data in new snapshot mode, should be updated data
    newSnapshotResult, err := table.Read(&readRecord1)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Record read in new snapshot mode: %v\n", newSnapshotResult)

    // Switch back to DB mode
    err = levelDBStore.SwitchToDB()
    if err != nil {
        panic(err)
    }
    fmt.Println("Switched back to DB mode successfully")
}
```

#### 10.9.4 Notes

1. **Snapshot lifecycle**: Snapshots will automatically release resources when switching back to database mode
2. **Nested snapshots**: Nested snapshots are not supported, i.e., snapshots cannot be created in snapshot mode
3. **Performance impact**: Snapshot creation operations are very efficient, but holding snapshots will affect LevelDB's garbage collection, it is recommended to release them in time
4. **Read consistency**: Snapshots provide a consistent view at creation time, subsequent modifications will not affect snapshots