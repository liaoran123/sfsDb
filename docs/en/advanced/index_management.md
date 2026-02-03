# Index Management

## 4.1 Index Interface Overview

sfsDb defines a complete index interface system, providing unified operation methods for different types of indexes.

### 4.1.1 Basic Index Interface

All index types implement the basic `Index` interface, which defines the core methods of indexes:

```go
// Basic index interface, defines methods common to all index types
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
    // Delete index fields
    DeleteFields(field ...string)
    // Concatenate prefix
    Prefix(tbid uint8) []byte
    // Concatenate values, no prefix needed
    Join(fieldsBytes *map[string][]byte) []byte
    // Concatenate prefix + value
    JoinPrefix(tbid uint8, val []byte) []byte
    // Concatenate index prefix + index value, call JoinPrefix, Join methods
    JoinValue(fieldsBytes *map[string][]byte, tbid uint8, existFields ...string) []byte
    // Match index fields
    MatchFields(fields ...string) bool
}
```

### 4.1.2 Index Type Hierarchy

sfsDb supports three main index types, each with corresponding interfaces and implementations:

1. **PrimaryKey Index** - Used to uniquely identify records
2. **NormalIndex** - Used to speed up queries
3. **FullTextIndex** - Used for text search

## 4.2 Primary Key Index

### 4.2.1 Primary Key Index Interface

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

### 4.2.2 Default Primary Key Index Implementation

```go
// Default primary key index
// Composite primary keys only support combinations of fixed-length types.
// String types must specify length. Otherwise, parsing issues or escape problems may cause bugs.
type DefaultPrimaryKey struct {
    BaseIndex // Embed basic index
}

func DefaultPrimaryKeyNew(name string) (*DefaultPrimaryKey, error)
```

### 4.2.3 Usage Example

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

## 4.3 Normal Index

### 4.3.1 Normal Index Interface

```go
// Normal index interface, embeds basic index interface
type NormalIndex interface {
    Index
    // Convert index value to primary key map value

    // Since NormalIndex completely matches the index interface, a Tag method is needed to distinguish whether it is a secondary index.
    Tag() bool
    Parse(primaryFields []string, pkfieldTypeLen *map[string]uint8, value []byte) (*map[string][]byte, error)
}
```

### 4.3.2 Default Normal Index Implementation

```go
// Default normal index, secondary index
type DefaultNormalIndex struct {
    BaseIndex // Embed basic index
}

func DefaultNormalIndexNew(name string) (*DefaultNormalIndex, error)

// Tag method returns true, indicating it is a secondary index.
func (dni *DefaultNormalIndex) Tag() bool
```

### 4.3.3 Usage Example

```go
// Create normal index
normalIndex, err := engine.DefaultNormalIndexNew("idx_name")
if err != nil {
    panic(err)
}
// Add index fields
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

## 4.4 Full-Text Index

### 4.4.1 Full-Text Index Interface

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

### 4.4.2 Default Full-Text Index Implementation

```go
// Default full-text index
type DefaultFullTextIndex struct {
    BaseIndex // Embed basic index
    // Full-text index split field
    ftsplit string
    // Split length
    ftlen int
}

func DefaultFullTextIndexNew(name string) (*DefaultFullTextIndex, error)
```

### 4.4.3 Usage Example

```go
// Create full-text index
fullTextIndex, err := engine.DefaultFullTextIndexNew("fulltext_desc")
if err != nil {
    panic(err)
}
// Add fields, the last field must be the primary key
// Full-text index must include the primary key (single or composite) in full, generally at the end. Otherwise, it will cause index non-uniqueness.
fullTextIndex.AddFields("description", "id")
or
fullTextIndex.AddFields("description", "id", "did") // "id", "did" are composite primary keys
// Set full-text index field and length
fullTextIndex.SetFullField("description", 5) // 5 represents the length of the full-text index. Generally choose 5 or 7 for Chinese. Mainly supports pictographic characters. This full-text index is not very suitable for non-pictographic characters.
err = table.CreateIndex(fullTextIndex)
if err != nil {
    panic(err)
}
```

## 4.5 Example: Creating a Table with Multiple Index Types

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

## 4.6 Index Implementation Details

### 4.6.1 Basic Index Structure

All index types embed the `BaseIndex` struct, which provides common index functionality:

```go
// Basic index struct, contains fields and methods common to all index types
type BaseIndex struct {
    fields []string
    id     uint8
    name   string
}
```

### 4.6.2 Index Field Processing

- **Add Fields**: `AddFields(field ...string)` - Add one or more fields to the index
- **Get Fields**: `GetFields() []string` - Get all fields of the index
- **Update Field Names**: `UpdateFields(oldfields string, newfields string)` - Modify index field names
- **Delete Fields**: `DeleteFields(field ...string)` - Remove fields from the index

### 4.6.3 Index Value Concatenation

Indexes internally use byte arrays to store index values, providing multiple concatenation methods:

- **Prefix**: Concatenate index prefix
- **Join**: Concatenate field values
- **JoinPrefix**: Concatenate prefix and values
- **JoinValue**: Concatenate complete index values

### 4.6.4 Composite Index Notes

1. **Composite Primary Key Limitations**: Composite primary keys only support combinations of fixed-length types; string types must specify length
2. **Field Order**: The order of index fields affects query performance; put the most frequently used fields first
3. **Index Size**: Indexes increase storage overhead; only create necessary indexes

## 4.7 Index Management System

### 4.7.1 Indexs Struct

`Indexs` is the core struct for index management in sfsDb, responsible for managing all indexes of a table:

```go
// Index management structure
type Indexs struct {
    id     uint8
    indexs []Index
    fields *map[string]any // Table fields
}
```

**Main Functions**:
- Create indexes (check field existence, index name uniqueness, etc.)
- Delete indexes
- Match indexes (prioritize primary key indexes, then normal indexes, finally full-text indexes)
- Manage index lifecycle

### 4.7.2 Index Type Limitations

#### 4.7.2.1 Boolean Type Not Supported as Index

sfsDb **does not support using boolean types (bool) as index fields** for the following reasons:

- **Index Key Conflicts**: Boolean types only have two possible values (true/false), which can cause multiple records to generate the same index key
- **Record Overwriting**: In KV storage, the same index key will be overwritten, only retaining the last inserted record
- **Inaccurate Queries**: Due to index key conflicts, queries can only find the last inserted record, not all matching records
- **Limited Performance Improvement**: Boolean fields usually have low selectivity, so the performance improvement from indexing is limited

#### 4.7.2.2 Alternative Solutions

**Solution 1: Use Primary Key Index Filtering**

```go
// 1. Use primary key index to get all records
searchData := map[string]any{"id": nil}
iter, _ := table.Search(&searchData)
if iter != nil {
    defer engine.GlobalTableIterPool.Put(iter)
}

// 2. Get all records and filter
allResults := iter.GetRecords(true)
defer record.PutRecords(allResults)
var activeRecords []map[string]any
for _, record := range allResults {
    if active, ok := record["active"].(bool); ok && active {
        activeRecords = append(activeRecords, record)
    }
}
```

**Solution 2: Use Integer Type Instead of Boolean Type**

- Define the `active` field as `int` type
- Use `0` to represent false, `1` to represent true
- This allows using existing integer index implementation, avoiding index key conflicts

```go
// Define fields using integer type
fields := map[string]any{
    "id":     0,
    "name":   "",
    "active": 0, // Use integer type, 0=false, 1=true
}

// Create index
activeIndex, err := engine.DefaultNormalIndexNew("idx_active")
activeIndex.AddFields("active")
err = table.CreateIndex(activeIndex)
```

### 4.7.3 Table-Level Index Management Methods

sfsDb provides rich table-level index management methods to facilitate users in creating and managing indexes:

#### Creating Indexes

```go
// Create custom index
func (t *Table) CreateIndex(index Index) error {
	if t.indexIDManager == nil {
		t.indexIDManager = NewIDManager(t.kvStore)
	}
	fkey := t.indexIDManager.GenerateIndexKey(t.id, index.Name())
	idxID, isNew, err := t.indexIDManager.GetOrCreateID(fkey)
	if err != nil {
		return err
	}
	err = t.indexs.createIndex(index, idxID)
	if err != nil {
		// Roll back ID
		_, err = t.indexIDManager.GetPreviousID(fkey)
		if err != nil {
			return err
		}
		return err
	}
	// Create corresponding new index data for existing table data
	if isNew {
		t.createIndexData(index)
	}
	return nil
}

// Create normal composite index
func (t *Table) CreateCompositeIndex(name string, fields ...string) error

// Create composite primary key index
func (t *Table) CreateCompositePrimaryKey(name string, fields ...string) error

// Create primary key index (supports single or multiple fields)
func (t *Table) CreatePrimaryKey(fields ...string) error

// Create normal index (simplified version, directly specify name and fields)
func (t *Table) CreateSimpleIndex(name string, fields ...string) error
```

#### Getting Indexes

```go
// Get primary key index
func (t *Table) GetPrimaryKey() PrimaryKey

// Get all indexes
func (t *Table) GetAllIndexes() []Index

// Get index by name
func (t *Table) GetIndexByName(name string) Index

// Get indexes by field name
func (t *Table) GetIndexesByField(field string) []Index

// Match index
func (t *Table) MatchIndex(fields ...string) Index
```

#### Deleting Indexes

**Important Note**: Deleting an index does not delete existing data, does not affect data, only creates redundancy.

```go
// Delete specified name index
// Deleting an index does not delete existing data, does not affect data, only creates redundancy.
func (t *Table) DropIndex(name string) error

// Delete primary key index
func (t *Table) DropPrimaryKey() error
```

### 4.7.4 Index Management Example

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

## 4.8 Index Optimization Recommendations

1. **Choose Appropriate Index Types**: Select appropriate index types based on query requirements
   - Unique record identification: Use primary key index
   - Speed up normal queries: Use normal index
   - Text search: Use full-text index

2. **Design Index Fields Reasonably**:
   - Primary key fields should be unique and stable
   - Normal indexes should be on fields frequently used in query conditions
   - Full-text indexes should be on fields requiring text search

3. **Control Index Quantity**:
   - Too many indexes will affect write performance
   - Only create indexes for frequently queried fields

4. **Use Composite Indexes**:
   - For multi-field queries, composite indexes are more efficient than multiple single-column indexes
   - Design composite indexes following the leftmost prefix principle

5. **Regularly Maintain Indexes**:
   - For frequently updated tables, rebuild indexes regularly
   - Delete indexes that are no longer used

## 4.9 Index Cache Settings

The `SetIndexCacheSizeLimit` function is used to set the size limit of the index cache, controlling memory usage and optimizing index operation performance.

```go
// Set index cache size limit
func (t *Table) SetIndexCacheSizeLimit(limit int) {
    // Control a reasonable value to prevent excessive cache size
    if limit <= 0 {
        limit = 1000
    }
    indexCacheSizeLimit = limit
}
```

**Parameter Description**:
- `limit`: Maximum number of index cache entries

**Usage Example**:

```go
// Set index cache size to 2000
table.SetIndexCacheSizeLimit(2000)

// Set index cache size to default value (1000)
table.SetIndexCacheSizeLimit(0) // When value ≤ 0, default value 1000 is used
```

**Notes**:
- Default value is 1000, suitable for most scenarios
- Larger cache size can improve performance for frequent index operations, but increases memory usage
- Smaller cache size reduces memory usage but may decrease performance for frequent index operations