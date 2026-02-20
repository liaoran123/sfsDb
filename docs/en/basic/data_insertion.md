# Inserting Data

## 3.1 Basic Insertion

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

## 3.2 Auto-increment

sfsDb supports primary key auto-increment functionality. **The field name must be "id"** and it must be a single primary key. The system will automatically generate incrementing primary key values in the following two cases:

### 3.2.1 Not providing id field

When inserting a record without including the "id" field, the system will automatically generate and add a primary key value:

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
```

### 3.2.2 Providing id field with nil value

When inserting a record that includes the "id" field but with a nil value, the system will also automatically generate a primary key value:

```go
// Explicitly set ID to nil, system also automatically generates
user3 := map[string]any{
    "id":   nil, // Explicitly set to nil, system also automatically generates
    "name": "Wang Wu",
    "age":  30,
    "email": "wangwu@example.com",
}
id3, err := table.Insert(&user3)
if err != nil {
    panic(err)
}
fmt.Printf("Automatically generated ID: %d\n", id3) // Output: 3
```

### 3.2.3 Auto-increment implementation principle

The auto-increment functionality is implemented based on the following conditions:

1. **Field name must be "id"**: The system only provides auto-increment functionality for fields named "id"
2. **Must be a single primary key**: Auto-increment only applies to single primary key tables
3. **Increment mechanism**: Uses an internal auto-increment counter to ensure unique and incrementing primary key values
4. **Version number automatically added**: The system automatically adds a version number field "v" to each record, with an initial value of 1

### 3.2.4 Notes

- Auto-increment functionality only takes effect during insertion operations
- Manually specifying a non-nil id value will override the automatic generation behavior
- The primary key value type is int, ensure it is consistent with the table structure definition
- The system automatically handles the version number field, no manual management is needed

## 3.3 Batch Insertion

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
defer GlobalTableIterPool.Put(allIter)
records := allIter.GetRecords(true)
defer record.PutRecords(records)
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

## 3.4 BatchInsert Functions

### 3.4.1 BatchInsert
Batch insert multiple records

```go
// Batch insert multiple records
users := []map[string]any{
    {"name": "Wang Wu", "age": 35, "email": "wangwu@example.com"},
    {"name": "Zhao Liu", "age": 28, "email": "zhaoliu@example.com"},
    {"name": "Sun Qi", "age": 40, "email": "sunqi@example.com"},
}

// Normal batch insertion
ids, err := table.BatchInsert(users)
if err != nil {
    panic(err)
}
fmt.Printf("Batch insertion successful, IDs: %v\n", ids)

// Batch insertion with shared batch (supports eventual consistency)
batch := storage.KVDb.GetBatch()
ids2, err := table.BatchInsert(users, batch)
if err != nil {
    panic(err)
}

// Commit batch
err = storage.KVDb.WriteBatch(batch)
if err != nil {
    panic(err)
}
fmt.Printf("Batch insertion with shared batch successful, IDs: %v\n", ids2)
```

### 3.4.2 BatchInsertWithSize
Batch insertion with batch size control

```go
// Batch insertion with batch size control
largeUsers := []map[string]any{}
for i := 0; i < 1000; i++ {
    largeUsers = append(largeUsers, map[string]any{
        "name": "User" + strconv.Itoa(i),
        "age":  20 + i%50,
        "email": "user" + strconv.Itoa(i) + "@example.com",
    })
}

// Process 100 records per batch
ids, err := table.BatchInsertWithSize(largeUsers, 100)
if err != nil {
    panic(err)
}
fmt.Printf("Batch insertion with size control successful, inserted %d records\n", len(ids))
```

### 3.4.3 BatchInsertNoInc
Batch insert records without auto-increment

```go
// Batch insert records without auto-increment (suitable for time-series data)
timeSeriesData := []map[string]any{
    {"id": time.Now().UnixNano(), "device_id": "dev001", "temperature": 25.5, "humidity": 60.0},
    {"id": time.Now().UnixNano(), "device_id": "dev001", "temperature": 25.6, "humidity": 59.8},
    {"id": time.Now().UnixNano(), "device_id": "dev002", "temperature": 26.0, "humidity": 58.5},
}

ids, err := table.BatchInsertNoInc(timeSeriesData)
if err != nil {
    panic(err)
}
fmt.Printf("Time-series data batch insertion successful, IDs: %v\n", ids)
```

### 3.4.4 BatchInsertWithSizeNoInc
Batch insertion with batch size control (without auto-increment)

```go
// Batch insertion with batch size control (without auto-increment)
largeTimeSeriesData := []map[string]any{}
for i := 0; i < 1000; i++ {
    largeTimeSeriesData = append(largeTimeSeriesData, map[string]any{
        "id": time.Now().UnixNano() + int64(i),
        "device_id": "dev" + strconv.Itoa(i%10),
        "temperature": 20.0 + float64(i%20),
        "humidity": 50.0 + float64(i%30),
    })
}

// Process 200 records per batch
ids, err := table.BatchInsertWithSizeNoInc(largeTimeSeriesData, 200)
if err != nil {
    panic(err)
}
fmt.Printf("Time-series data insertion with size control successful, inserted %d records\n", len(ids))
```

## 3.5 Eventual Consistency Support

### 3.5.1 What is Eventual Consistency

Eventual consistency is a data consistency model in distributed systems that guarantees the system will eventually reach a consistent state if no new update operations are performed. In industrial edge computing scenarios, eventual consistency is often preferred over strong consistency because it provides better performance and availability.

### 3.5.2 How to Implement Eventual Consistency with Shared Batch

In sfsDb, you can implement eventual consistency by using shared batch objects:

```go
// 1. Create a shared batch object
batch := storage.KVDb.GetBatch()
if batch == nil {
    panic("Cannot get batch operation object")
}

// 2. Add multiple operations to the same batch

// Operation 1: Insert record
user := map[string]any{
    "name": "Zhang San",
    "age":  30,
    "email": "zhangsan@example.com",
}
table1.Insert(&user, batch)

// Operation 2: Update record
updateData := map[string]any{
    "id": 1,
    "age": 31,
}
table1.Update(&updateData, batch)

// Operation 3: Delete record
deleteData := map[string]any{
    "id": 2,
}
table1.Delete(&deleteData, batch)

// Operation 4: Insert record in another table
deviceData := map[string]any{
    "id": "dev001",
    "name": "Temperature Sensor",
    "status": "online",
}
table2.Insert(&deviceData, batch)

// 3. Manually commit batch (all operations executed at once)
err = storage.KVDb.WriteBatch(batch)
if err != nil {
    panic(fmt.Sprintf("Batch commit failed: %v", err))
}
fmt.Println("All operations committed, system will eventually reach consistent state")
```

### 3.5.3 Application Scenarios for Eventual Consistency

1. **Industrial Edge Computing**: Data synchronization between edge devices and cloud, prioritizing performance and availability
2. **IoT Scenarios**: Collection and processing of large amounts of sensor data, tolerating short-term data inconsistency
3. **Log Processing**: Batch collection and processing of logs, focusing on throughput rather than real-time consistency
4. **Cache Updates**: Asynchronous updates between cache and database, improving system response speed

By using BatchInsert functions and shared batch objects, sfsDb provides efficient data operation methods for edge computing and IoT scenarios, while supporting the implementation of eventual consistency.