# Deleting Records

## 8.1 Single Record Deletion

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

## 8.2 Batch Deletion

```go
// Batch delete records: Delete users younger than 25

// First search for records matching the criteria
searchCriteria := map[string]any{
    "age": map[util.ComparisonOperator]any{util.LessThan: 25},
}
iter := table.Search(&searchCriteria)
defer GlobalTableIterPool.Put(iter)

// Use iterator's Delete method to batch delete matching records
// Directly delete all matching records (unlimited)
iter.Delete()

// Or limit deletion quantity: Delete first 2 matching records
// iter.Delete(2)
```

## 8.3 Delete All Data in Table

sfsDb provides a `DeleteAll()` method for deleting all data in a table. This method traverses all key-value pairs in the table and uses batch operations to delete all data.

```go
// Delete all data in the table
func (t *Table) DeleteAll() error {
	// Get iterator for all kv key-value pairs in the table
	iter := t.For()
	defer GlobalTableIterPool.Put(iter)
	
	// Create batch operation
	batch := t.kvStore.GetBatch()
	if batch == nil {
		return fmt.Errorf("failed to get batch")
	}
	
	// Define batch operation size limit
	const batchSizeLimit = 1000
	
	// Traverse and delete all key-value pairs
	count := 0
	for iter.Next() {
		key := iter.Key()
		batch.Delete(key)
		count++
		
		// When batch operation size reaches limit, execute batch operation and reset batch operation object
		if count >= batchSizeLimit {
			// Commit batch operation
			if err := t.kvStore.WriteBatch(batch); err != nil {
				return err
			}
			
			// Reset counter and batch operation object
			count = 0
			batch = t.kvStore.GetBatch()
			if batch == nil {
				return fmt.Errorf("failed to get batch")
			}
		}
	}
	
	// Execute remaining batch operations
	if count > 0 {
		if err := t.kvStore.WriteBatch(batch); err != nil {
			return err
		}
	}
	
	return nil
}
```

**Usage example**:

```go
// Delete all data in the table
err = table.DeleteAll()
if err != nil {
    panic(err)
}
fmt.Println("Table data deleted successfully")
```

**Working principle**:
1. The `DeleteAll()` method first calls the `For()` method to get an iterator for all key-value pairs in the table
2. Then creates a batch operation object for batch deletion of key-value pairs
3. Defines a batch operation size limit (default 1000), and when the limit is reached, executes the batch operation and resets
4. Traverses the iterator, adding each key to the batch operation
5. When the batch operation reaches the limit, executes the batch operation and resets
6. Finally executes the remaining batch operations

**Notes**:
- The `DeleteAll()` method will delete all data in the table, including index data
- This method uses a batch processing mechanism to ensure it doesn't exceed batch operation size limits when processing large amounts of data
- Deletion operations do not affect data in other tables
- After deletion, the table still exists, only the data is cleared
- You can insert data into the table again after deletion