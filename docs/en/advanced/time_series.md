# Time Series Processing

## Overview

The time package is a time-series data processing tool provided by sfsDb, used for handling time-related data operations including time granularity handling, time window calculations, data aggregation, and timestamp conversion.

This package is particularly suitable for:
- Processing and analyzing IoT device sensor data
- Aggregating and trend analysis of system monitoring data
- Time series analysis of financial transaction data
- Any application scenario that requires processing and analyzing time-series data

## Core Features

### Time Granularity Handling

Time granularity handling allows you to process time data at different precision levels, supporting the following time granularities:

- **Second**
- **Minute**
- **Hour**
- **Day**
- **Month**
- **Year**

Use the `FormatTimeByGranularity` function to format timestamps into time strings with the specified granularity:

```go
import (
    "time"
    sfsTime "github.com/liaoran123/sfsDb/time"
)

// Format time
now := time.Now()
hourStr := sfsTime.FormatTimeByGranularity(now, sfsTime.TimeGranularityHour)
fmt.Println("Hour:", hourStr) // Output: Hour: 2024-12-25 14:00:00
```

### Time Window Calculation

Time window calculation computes time ranges based on the specified time granularity, returning start and end times:

```go
// Calculate time range
now := time.Now()
start, end := sfsTime.TimeRange(now, sfsTime.TimeGranularityDay)
fmt.Println("Day range:", start, "to", end)
```

### Data Aggregation

Data aggregation aggregates data by time granularity, supporting multiple aggregation types:

- **Sum**
- **Average** (avg)
- **Count**
- **Maximum** (max)
- **Minimum** (min)

Use the `AggregateByTimeGranularity` function for data aggregation:

```go
// Assuming we have a set of sensor data records
records := record.Records{...}

// Aggregate by hour
results, err := sfsTime.AggregateByTimeGranularity(
    records,
    "timestamp",    // Time field
    "value",        // Value field
    sfsTime.TimeGranularityHour,  // Time granularity
    "sum",          // Aggregation type
)

// Output results
for _, result := range results {
    fmt.Printf("Time: %s, Sum: %.2f\n", result.TimeKey, result.Value)
}
```

### Timestamp Conversion

Timestamp conversion functionality allows conversion between `time.Time` objects and integer timestamps, supporting different timestamp precisions:

- **Second** precision timestamps
- **Millisecond** precision timestamps
- **Nanosecond** precision timestamps

```go
// Convert time object to integer timestamp
now := time.Now()
unixTime := sfsTime.TimeToUnixTimestamp(now)         // Second precision
unixTimeMs := sfsTime.TimeToUnixTimestampMs(now)     // Millisecond precision
unixTimeNs := sfsTime.TimeToUnixTimestampNs(now)     // Nanosecond precision

// Convert integer timestamp to time object
timeObj := sfsTime.UnixTimestampToTime(unixTime)     // Second precision
timeObjMs := sfsTime.UnixTimestampMsToTime(unixTimeMs) // Millisecond precision
timeObjNs := sfsTime.UnixTimestampNsToTime(unixTimeNs) // Nanosecond precision
```

### Time Bucket Allocation

Time bucket allocation assigns timestamps to time buckets of the specified granularity, returning the start time of the time bucket:

```go
// Allocate time bucket
now := time.Now()
bucket := sfsTime.TimeBucket(now, time.Hour)
fmt.Println("Bucket:", bucket) // Output: start time of the time bucket
```

## Database Integration

### Basic Integration Example

Here's a basic example of integrating the time package with sfsDb:

```go
// Create table
sensorTable, err := engine.TableNew("sensor_data")
if err != nil {
    return err
}

// Set fields
fields := map[string]any{
    "timestamp": time.Now(), // Timestamp, time.Time type
    "value":     0.0,        // Sensor value
    "sensor_id": "",         // Sensor ID
}
err = sensorTable.SetFields(fields)
if err != nil {
    return err
}

// Create primary key index
primaryKey, err := engine.DefaultPrimaryKeyNew("pk")
if err != nil {
    return err
}
primaryKey.AddFields("timestamp")
err = sensorTable.CreateIndex(primaryKey)
if err != nil {
    return err
}

// Insert test data
now := time.Now()
for i := 0; i < 10; i++ {
    data := map[string]any{
        "timestamp": now.Add(time.Duration(i) * time.Minute),
        "value":     float64(i * 10),
        "sensor_id": fmt.Sprintf("sensor_%d", i%3),
    }
    _, err := sensorTable.Insert(&data)
    if err != nil {
        return err
    }
}

// Query data from the last 1 hour
startTime := time.Now().Add(-1 * time.Hour)
endTime := time.Now()
iter, err := sensorTable.SearchRange(sensorTable.kvStore.Iterator, "timestamp", startTime, endTime)
if err != nil {
    return err
}
defer engine.GlobalTableIterPool.Put(iter)

// Get results
records := iter.GetRecordSet(true)
defer record.PutRecords(records)

// Aggregate by hour
aggregationResults, err := sfsTime.AggregateByTimeGranularity(
    records,
    "timestamp",
    "value",
    sfsTime.TimeGranularityHour,
    "avg",
)
if err != nil {
    return err
}

// Output aggregation results
fmt.Println("Hourly average sensor values:")
for _, result := range aggregationResults {
    fmt.Printf("Time: %s, Average: %.2f\n", result.TimeKey, result.Value)
}
```

### Composite Primary Key Example

For IoT scenarios, it's common to use composite primary keys (like `sensor_id` + `timestamp`) to uniquely identify records:

```go
// Create table
deviceTable, err := engine.TableNew("device_data")
if err != nil {
    return err
}

// Set fields
fields := map[string]any{
    "sensor_id": "",         // Sensor ID
    "timestamp": time.Now(), // Timestamp
    "value":     0.0,        // Sensor value
}
err = deviceTable.SetFields(fields)
if err != nil {
    return err
}

// Create composite primary key index
primaryKey, err := engine.DefaultPrimaryKeyNew("device_time_idx")
if err != nil {
    return err
}
primaryKey.AddFields("sensor_id")
primaryKey.AddFields("timestamp")
err = deviceTable.CreateIndex(primaryKey)
if err != nil {
    return err
}

// Insert data and query logic similar to the example above
```

## Best Practices

### 1. Time Field Type Selection

- **Recommended: Use `time.Time` type** - Seamless integration with the time package, supporting full time functionality
- **Use integer timestamps** - If you have specific storage efficiency requirements or need integration with other systems

### 2. Index Design

- **Timestamp as primary key** - Suitable for single-device time-series data
- **Composite primary key** - For multi-device scenarios, use `(device_id, timestamp)` composite primary key
- **Secondary indexes** - Create secondary indexes for frequently queried fields

### 3. Performance Optimization

- **Batch insertion** - Use batch operations to reduce database interactions
- **Data downsampling** - For long-term stored data, consider downsampling to larger time granularities
- **Time range queries** - Use the `SearchRange` method for efficient time range queries
- **Memory management** - Release iterator and record set resources promptly

### 4. Data Management

- **Data expiration strategy** - Set reasonable expiration times for time-series data
- **Data archiving** - Archive historical data to low-cost storage
- **Backup strategy** - Regularly back up important time-series data

## Summary

The time package provides sfsDb with powerful time-series data processing capabilities, enabling you to:

- Process and analyze data at different time granularities
- Perform time window calculations and data aggregation
- Convert between different time representation formats
- Seamlessly integrate with the database to build complete time-series data solutions

By properly using the features of the time package, you can build high-performance, reliable time-series data processing systems that meet the needs of various application scenarios.