# Primary Key Management

## 3.1 Single Primary Key

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

## 3.2 Composite Primary Key

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

## 3.3 Primary Key Auto-Creation Mechanism

When no primary key is explicitly created for a table, sfsDb will automatically create a default primary key when a primary key is first needed. This mechanism is implemented by the `GetPrimaryKey()` method:

```go
// Get primary key index
func (t *Table) GetPrimaryKey() PrimaryKey {
    pk := t.indexs.getPrimaryKey()
    if pk == nil {
        // No primary key, need to create a default primary key.
        // Therefore, primary key must be created before table has data.
        pk, _ = DefaultPrimaryKeyNew("id")
        pk.AddFields("id")
        t.CreateIndex(pk)
    }
    return pk
}
```

**Working Principle**:
1. When the `GetPrimaryKey()` method is called, the system first checks if the table already has a primary key
2. If there is no primary key, the system automatically creates a default primary key named "id"
3. The default primary key adds the "id" field as the primary key field
4. This default primary key index is then added to the table
5. Finally, the created primary key index is returned

**Notes**:
- The system comment clearly states: "Therefore, primary key must be created before table has data."
- This means if the table already has data, calling operations that require a primary key may lead to data inconsistency
- Therefore, it is recommended to explicitly set the primary key after table creation and before data insertion

**Usage Scenarios**:
- Rapid prototyping, temporarily not needing a custom primary key
- Simple applications where using the default auto-increment ID as the primary key suffices

**Importance of Index Creation Order**:
When adding indexes, the system executes the `createIndexData` function to create index data for existing data, which calls the `GetPrimaryKey()` method:

```go
// Create corresponding new index data for existing table data
func (t *Table) createIndexData(index Index) error {
    // Exclude primary key
    if index == t.GetPrimaryKey() {
        return nil
    }
    // Iterate through all data based on primary key prefix to create index data
    // Get primary key prefix
    pk := t.GetPrimaryKey() // This triggers auto-creation of primary key
    pkPrefix := pk.Prefix(t.id)
    slice := util.NewRangeHelper(pkPrefix).FromComparison(util.Like, pkPrefix)
    // Create iterator with pkPrefix
    iter := t.kvStore.Iterator(slice.Start, slice.Limit)
    defer GlobalTableIterPool.Put( iter)
    // Iterate through all data
    var value []byte
    for iter.Next() {
        _, value = iter.Key(), iter.Value()
        rval, err := pk.Parse(t.fieldsid, value)
        if err != nil {
            return err
        }
        if rval == nil {
            continue
        }
        indexValue := pk.GetID(rval)
        switch index := index.(type) {
        case NormalIndex:
            idxvalueofkey := index.Join(rval)
            idxkey := index.JoinPrefix(t.id, idxvalueofkey)
            t.kvStore.Put(idxkey, indexValue)
        case FullTextIndex:
            // Similar to func (c *batchContainer) Operation
            fieldsBytes, err := pk.Parse(t.fieldsid, value)
            if err != nil {
                return err
            }
            if fieldsBytes == nil {
                continue
            }
            joinValues := index.JoinFullValues(fieldsBytes, t.id)
            defer util.PutBytesArray(joinValues)
            for _, joinValue := range joinValues {
                if joinValue == nil {
                    continue
                }
                // Create index data for full-text index
                t.kvStore.Put(joinValue, indexValue)
            }
        }
    }
    return nil
}
```

**Important Recommendations**:
1. **Add primary key first, then add other indexes**: If no primary key is explicitly created before adding other indexes, the system will automatically create a default primary key when executing `createIndexData`
2. **Avoid auto-creation of primary key**: Auto-created primary keys may not meet your business requirements; it is recommended to always explicitly create primary keys
3. **Set primary key before data insertion**: Ensure the primary key is set before inserting any data to avoid data inconsistency