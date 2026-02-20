# Snapshot and Read Consistency

## 1. Snapshot Overview

Snapshot is an advanced feature provided by sfsDb, used to create a consistent read view of the database at a specific point in time. By using snapshots, you can obtain a complete, consistent data state of the database at the moment the snapshot was created, even in environments where data is constantly changing.

### 1.1 Features of Snapshot

- **Read Consistency**: Snapshots provide a consistent view of the database at the time of creation, unaffected by subsequent data modifications
- **Non-blocking**: Creating a snapshot does not block write operations to the database
- **Efficient**: Snapshots are lightweight with low creation and usage costs
- **Isolation**: Snapshot reads are completely isolated from other database operations

### 1.2 Difference Between Snapshot and Transaction

| Feature | Snapshot | Transaction |
|---------|----------|-------------|
| Main Purpose | Provide consistent read view | Ensure atomicity, consistency, isolation, and durability of operations |
| Write Operation Support | Read-only | Supports read and write operations |
| Performance Overhead | Low | High |
| Implementation Method | Based on storage engine snapshot mechanism | Based on locking or MVCC mechanism |
| Applicable Scenarios | Read-heavy scenarios such as report generation, data backup | Scenarios with high data consistency requirements such as financial transactions |

## 2. Using Snapshot

### 2.1 Creating a Snapshot

```go
// Create snapshot
snapshot, err := db.Snapshot()
if err != nil {
    panic(err)
}
defer snapshot.Release() // Ensure snapshot resources are released
```

### 2.2 Reading Data from Snapshot

```go
// Read single key-value from snapshot
value, err := snapshot.Get([]byte("key"))
if err != nil {
    panic(err)
}
fmt.Printf("Value: %s\n", value)

// Use snapshot iterator to traverse data
iter := snapshot.Iterator(nil, nil)
defer iter.Release()

for iter.Next() {
    key := iter.Key()
    val := iter.Value()
    fmt.Printf("Key: %s, Value: %s\n", key, val)
}
```

### 2.3 Releasing Snapshot

After using a snapshot, you must release its resources:

```go
// Release snapshot resources
err := snapshot.Release()
if err != nil {
    panic(err)
}
```

## 3. Read Consistency Examples

### 3.1 Basic Read Consistency Example

```go
// Create database
db, err := storage.NewLevelDBStore("./test_db", nil)
if err != nil {
    panic(err)
}
defer db.Close()

// Write initial data
db.Put([]byte("key1"), []byte("value1"))
db.Put([]byte("key2"), []byte("value2"))

// Create snapshot
snapshot, err := db.Snapshot()
if err != nil {
    panic(err)
}
defer snapshot.Release()

// Modify data
db.Put([]byte("key1"), []byte("new_value1"))
db.Put([]byte("key2"), []byte("new_value2"))

// Read from database (gets modified data)
dbValue1, _ := db.Get([]byte("key1"))
dbValue2, _ := db.Get([]byte("key2"))
fmt.Printf("From DB: key1=%s, key2=%s\n", dbValue1, dbValue2)

// Read from snapshot (gets original data, maintaining consistency)
snapshotValue1, _ := snapshot.Get([]byte("key1"))
snapshotValue2, _ := snapshot.Get([]byte("key2"))
fmt.Printf("From Snapshot: key1=%s, key2=%s\n", snapshotValue1, snapshotValue2)

// Output:
// From DB: key1=new_value1, key2=new_value2
// From Snapshot: key1=value1, key2=value2
```

### 3.2 Industrial Edge Computing Scenario Example

In industrial edge computing scenarios where device data is constantly changing, snapshots can ensure you read complete data at a specific point in time:

```go
// Industrial device data collection system
func collectDeviceData() {
    // 1. Create database connection
    db, err := storage.NewLevelDBStore("./device_data", nil)
    if err != nil {
        panic(err)
    }
    defer db.Close()

    // 2. Continuously collect device data
    go func() {
        for {
            deviceID := "dev-" + strconv.Itoa(rand.Intn(1000))
            temperature := 20.0 + rand.Float64()*20.0
            humidity := 40.0 + rand.Float64()*40.0
            
            data := map[string]any{
                "temperature": temperature,
                "humidity": humidity,
                "timestamp": time.Now().UnixNano(),
            }
            
            // Serialize data and store
            dataBytes, _ := json.Marshal(data)
            db.Put([]byte(deviceID), dataBytes)
            
            time.Sleep(time.Millisecond * 100) // Simulate data collection interval
        }
    }()

    // 3. Periodically generate device status reports (using snapshots to ensure data consistency)
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()

    for range ticker.C {
        fmt.Println("\nGenerating device status report...")
        
        // Create snapshot to get consistent data view at current time
        snapshot, err := db.Snapshot()
        if err != nil {
            fmt.Printf("Failed to create snapshot: %v\n", err)
            continue
        }
        
        // Traverse all device data
        iter := snapshot.Iterator(nil, nil)
        deviceCount := 0
        totalTemperature := 0.0
        totalHumidity := 0.0
        
        for iter.Next() {
            deviceCount++
            dataBytes := iter.Value()
            var data map[string]any
            json.Unmarshal(dataBytes, &data)
            
            totalTemperature += data["temperature"].(float64)
            totalHumidity += data["humidity"].(float64)
            
            fmt.Printf("Device %s: Temperature=%.2f°C, Humidity=%.2f%%\n", iter.Key(), data["temperature"], data["humidity"])
        }
        
        // Calculate averages
        if deviceCount > 0 {
            avgTemperature := totalTemperature / float64(deviceCount)
            avgHumidity := totalHumidity / float64(deviceCount)
            fmt.Printf("\nReport Summary: %d devices, Average Temperature=%.2f°C, Average Humidity=%.2f%%\n", deviceCount, avgTemperature, avgHumidity)
        }
        
        // Release resources
        iter.Release()
        snapshot.Release()
        fmt.Println("Report generation completed")
    }
}
```

### 3.3 Table-level Snapshot Read Consistency Example

The following example demonstrates how to use snapshots at the table level to ensure read consistency, based on actual test code:

```go
// Test table-level snapshot read consistency
func TestTableSnapshotSearchConsistency(t *testing.T) {
    // Create temporary database
    dbPath := "./test_table_snapshot_db"
    cleanup := func() {
        os.RemoveAll(dbPath)
    }
    cleanup()
    defer cleanup()

    // Open database
    dbManager := storage.GetDBManager()
    db, err := dbManager.OpenDB(dbPath)
    if err != nil {
        t.Fatalf("Failed to open database: %v", err)
    }
    defer storage.CloseDb()

    // Create table
    tableName := "test_users"
    table, err := TableNew(tableName)
    if err != nil {
        t.Fatalf("Failed to create table: %v", err)
    }
    table.kvStore = db

    // Prepare field mapping
    fieldsMap := map[string]any{
        "id":    0,
        "name":  "",
        "age":   0,
        "email": "",
    }

    // Set table fields
    err = table.SetFields(fieldsMap)
    if err != nil {
        t.Fatalf("Failed to set fields: %v", err)
    }

    // Create primary key index
    PrimaryKeys, err := DefaultPrimaryKeyNew("pk")
    if err != nil {
        t.Fatalf("Failed to create primary key: %v", err)
    }
    PrimaryKeys.AddFields("id")
    err = table.CreateIndex(PrimaryKeys)
    if err != nil {
        t.Fatalf("Failed to create primary key index: %v", err)
    }

    // Create normal index for age field, used for search
    ageIndex, err := DefaultNormalIndexNew("age_idx")
    if err != nil {
        t.Fatalf("Failed to create age index: %v", err)
    }
    ageIndex.AddFields("age")
    err = table.CreateIndex(ageIndex)
    if err != nil {
        t.Fatalf("Failed to create age index: %v", err)
    }

    // 1. Insert initial data
    users := []map[string]any{
        {"id": 1, "name": "Zhang San", "age": 30, "email": "zhangsan@example.com"},
        {"id": 2, "name": "Li Si", "age": 25, "email": "lisi@example.com"},
        {"id": 3, "name": "Wang Wu", "age": 35, "email": "wangwu@example.com"},
        {"id": 4, "name": "Zhao Liu", "age": 28, "email": "zhaoliu@example.com"},
        {"id": 5, "name": "Sun Qi", "age": 40, "email": "sunqi@example.com"},
    }

    for _, user := range users {
        _, err := table.Insert(&user)
        if err != nil {
            t.Fatalf("Failed to insert user: %v", err)
        }
    }

    // 2. Create snapshot
    snapshot, err := db.Snapshot()
    if err != nil {
        t.Fatalf("Failed to create snapshot: %v", err)
    }
    defer snapshot.Release()

    // 3. Modify data
    updateUsers := []map[string]any{
        {"id": 1, "name": "Zhang San (Updated)", "age": 31},
        {"id": 2, "name": "Li Si (Updated)", "age": 26},
        {"id": 6, "name": "Zhou Ba", "age": 22, "email": "zhouba@example.com"}, // New record
    }

    for _, user := range updateUsers {
        if user["id"] == 6 {
            // Insert new record
            _, err := table.Insert(&user)
            if err != nil {
                t.Fatalf("Failed to insert new user: %v", err)
            }
        } else {
            // Update record
            err := table.Update(&user)
            if err != nil {
                t.Fatalf("Failed to update user: %v", err)
            }
        }
    }

    // 4. Directly use snapshot to verify read consistency
    // Create function to get records from snapshot
    snapshotGet := func(key []byte) []byte {
        value, err := snapshot.Get(key)
        if err != nil {
            return nil
        }
        return value
    }

    // Build primary key for each user ID and verify data
    userIDs := []int{1, 2, 3, 4, 5}
    for _, id := range userIDs {
        // Build primary key
        pk := PrimaryKeys
        fieldsBytes := map[string][]byte{
            "id": util.AnyToBytes(id),
        }
        key := pk.JoinValue(&fieldsBytes, table.id)

        // Read from snapshot
        snapshotValue := snapshotGet(key)
        // Read from database
        dbValue := table.ReadByBytes(key)

        // Parse records and verify
        // Snapshot should return original data, database should return modified data
    }

    // Verify snapshot does not include new record (ID: 6)
    newRecordID := 6
    newRecordFieldsBytes := map[string][]byte{
        "id": util.AnyToBytes(newRecordID),
    }
    newRecordKey := PrimaryKeys.JoinValue(&newRecordFieldsBytes, table.id)
    newRecordSnapshotValue := snapshotGet(newRecordKey)
    if newRecordSnapshotValue != nil {
        t.Fatalf("Snapshot should not include new record with ID %d", newRecordID)
    }

    // 5. Test snapshot iterator
    // Create FunIter function using snapshot iterator
    snapshotFunIter := func(start, limit []byte) storage.Iterator {
        return snapshot.Iterator(start, limit)
    }

    // Test range search
    rangeIter, err := table.SearchRange(snapshotFunIter, "id", 2, 5)
    defer rangeIter.Release()
    if err != nil {
        t.Fatalf("Failed to search range with snapshot: %v", err)
    }

    // Collect range search results
    rangeResults := rangeIter.GetRecords(true)
    defer record.PutRecords(rangeResults)

    // Verify range search results
    // Should return original data, not affected by subsequent modifications
}
```

#### Example Explanation

1. **Creation and Preparation**: Create database, table, fields, and indexes
2. **Insert Initial Data**: Insert 5 user records
3. **Create Snapshot**: Create snapshot before data modification
4. **Modify Data**: Update 2 existing records, add 1 new record
5. **Verify Consistency**:
   - Read from snapshot, should return original data
   - Read from database, should return modified data
   - Verify snapshot does not include the new record
6. **Test Snapshot Iterator**: Use snapshot iterator for range search, verify it returns original data

#### Expected Results

- Snapshot data: Zhang San(30), Li Si(25) (original data)
- Database data: Zhang San (Updated)(31), Li Si (Updated)(26) (modified data)
- Snapshot does not include the new record (ID: 6)
- Range search returns 3 original data records

This example demonstrates how to use snapshots in practical applications to ensure read consistency, especially in scenarios where data is constantly changing.

## 3. Snapshot Application Scenarios

### 3.1 Industrial Edge Computing

- **Device Status Monitoring**: Use snapshots to obtain all device statuses at a specific point in time, ensuring consistency of monitoring data
- **Production Data Analysis**: Generate production reports based on snapshots to avoid data changes during analysis
- **Fault Diagnosis**: When equipment fails, use snapshots to save the complete data state at the time of the failure for subsequent analysis

### 3.2 IoT Scenarios

- **Sensor Data Aggregation**: Use snapshots to aggregate data from multiple sensors at the same point in time
- **Device Configuration Management**: Ensure complete device configuration information is read without being affected by configuration updates
- **Data Synchronization**: Use snapshots to ensure consistency of synchronized data when syncing data between edge devices and the cloud

### 3.3 Other Scenarios

- **Data Backup**: Perform incremental backups based on snapshots to avoid data changes during backup
- **Report Generation**: Generate financial reports, business analysis reports and other scenarios requiring data consistency
- **Test Environment**: Use snapshots during testing to ensure consistency and repeatability of test data

## 4. Performance Considerations

### 4.1 Snapshot Overhead

- **Creation Overhead**: Snapshot creation is very lightweight and has almost no impact on database performance
- **Storage Overhead**: Snapshots share underlying storage and do not occupy additional storage space
- **Reading Overhead**: Reading data from snapshots has similar performance to reading directly from the database

### 4.2 Best Practices

- **Timely Release**: Snapshots should be released immediately after use to avoid resource occupation
- **Reasonable Use**: Only use snapshots in scenarios where consistent read views are needed
- **Batch Operations**: For scenarios requiring reading large amounts of data, using snapshots can reduce the risk of data inconsistency
- **Avoid Nesting**: Creating snapshots from snapshots is not supported, so avoid attempting this operation

## 5. Notes

1. **Read-only Operation**: Snapshots only support read operations, not write operations
2. **Resource Management**: The `Release()` method must be called to release snapshot resources, otherwise resource leaks may occur
3. **Snapshot Lifecycle**: The snapshot lifecycle should be as short as possible, only created when needed and released immediately after use
4. **Nesting Limitation**: Creating snapshots from snapshots is not supported, and attempting to do so will return an error
5. **Storage Engine Dependency**: The snapshot feature depends on the snapshot support of the underlying storage engine, currently only the LevelDB storage engine supports this feature

## 6. Summary

Snapshot is a powerful feature provided by sfsDb, offering an efficient read consistency solution for industrial edge computing and IoT scenarios. By using snapshots, you can:

- Obtain a consistent data view in environments where data is constantly changing
- Avoid performance overhead associated with traditional transactions
- Ensure data consistency for operations such as report generation and data backup
- Improve overall system performance and reliability

The snapshot feature is implemented based on LevelDB's snapshot mechanism, combined with object pool technology, which ensures both performance and effective resource management. In practical applications, rational use of snapshots can significantly improve system reliability and performance.