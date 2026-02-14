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

- **Millisecond**
- **Microsecond**
- **Second**
- **Minute**
- **Hour**
- **Day**
- **Week**
- **Month**
- **Quarter**
- **Year**

Additionally, the time package provides utilities for working with time patterns:

- **IsWeekday**: Check if a given time is a weekday
- **IsWeekend**: Check if a given time is a weekend
- **GetQuarter**: Get the quarter of the year for a given time
- **GetWeekNumber**: Get the week number of the year for a given time

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

Time window calculation computes time ranges based on the specified time granularity, returning start and end times. The time package now supports two types of time windows:

1. **Sliding Window**: A fixed-size window that slides over time with a specified step size
2. **Tumbling Window**: Non-overlapping fixed-size windows that tumble over time

#### Sliding Window Example

```go
// Create sliding window
startTime := time.Now().Add(-10 * time.Minute)
endTime := time.Now()
windowSize := 2 * time.Minute
stepSize := 1 * time.Minute

window := sfsTime.NewSlidingWindow(startTime, endTime, windowSize, stepSize)

// Iterate through windows
for window.Next() {
    start := window.Start()
    end := window.End()
    fmt.Printf("Window: %s to %s\n", start, end)
}

// Reset window
window.Reset()
```

#### Tumbling Window Example

```go
// Create tumbling window
startTime := time.Now().Add(-10 * time.Minute)
endTime := time.Now()
windowSize := 2 * time.Minute

window := sfsTime.NewTumblingWindow(startTime, endTime, windowSize)

// Iterate through windows
for window.Next() {
    start := window.Start()
    end := window.End()
    fmt.Printf("Window: %s to %s\n", start, end)
}
```

#### Window Aggregation

You can aggregate data within each time window:

```go
// Create test data
var records []map[string]any
now := time.Now()
for i := 0; i < 10; i++ {
    records = append(records, map[string]any{
        "timestamp": now.Add(time.Duration(i) * time.Minute),
        "value":     float64(i),
    })
}

// Create window
window := sfsTime.NewSlidingWindow(startTime, endTime, windowSize, stepSize)

// Aggregate by window
results, err := sfsTime.AggregateByWindow(
    records,
    "timestamp",
    "value",
    window,
    "sum"
)

// Output results
for _, result := range results {
    fmt.Printf("Window: %s to %s, Sum: %.2f\n", result.WindowStart, result.WindowEnd, result.Value)
}
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

### Time Series Prediction

The time package now includes time series prediction capabilities, supporting two prediction methods:

1. **Moving Average Prediction**: Uses the moving average of historical data to predict future values
2. **Linear Regression Prediction**: Uses linear regression to predict future values based on historical trends

#### Moving Average Prediction Example

```go
// Create test data points
var points []sfsTime.TimeSeriesPoint
now := time.Now()
for i := 0; i < 10; i++ {
    points = append(points, sfsTime.TimeSeriesPoint{
        Time:  now.Add(time.Duration(i) * time.Minute),
        Value: float64(i),
    })
}

// Create moving average prediction
maPrediction := sfsTime.NewMovingAveragePrediction(points, 3, 5, time.Minute)

// Output predicted points
fmt.Println("Moving average predictions:")
for _, point := range maPrediction.PredictedPoints {
    fmt.Printf("Predicted: %s, Value: %.2f\n", point.Time, point.Value)
}
```

#### Linear Regression Prediction Example

```go
// Create linear regression prediction
lrPrediction := sfsTime.NewLinearRegressionPrediction(points, 5, time.Minute)

// Output predicted points
fmt.Println("Linear regression predictions:")
for _, point := range lrPrediction.PredictedPoints {
    fmt.Printf("Predicted: %s, Value: %.2f\n", point.Time, point.Value)
}

// Output regression parameters
fmt.Printf("Regression parameters: Slope=%.2f, Intercept=%.2f, R²=%.2f\n", 
    lrPrediction.Slope, lrPrediction.Intercept, lrPrediction.R2)
```

### Time Series Data Compression

The time package now supports time series data compression to reduce storage space:

1. **Delta Encoding**: Compresses data by storing the difference between consecutive values
2. **Run-Length Encoding (RLE)**: Compresses data by storing repeated values and their counts

#### Compression Example

```go
// Create test data points
var points []sfsTime.TimeSeriesPoint
now := time.Now()
for i := 0; i < 10; i++ {
    points = append(points, sfsTime.TimeSeriesPoint{
        Time:  now.Add(time.Duration(i) * time.Minute),
        Value: float64(i),
    })
}

// Compress time series data
compressed, err := sfsTime.CompressTimeSeries(points, "delta", time.Minute)
if err != nil {
    panic(err)
}

// Decompress time series data
decompressed, err := sfsTime.DecompressTimeSeries(compressed, 10)
if err != nil {
    panic(err)
}

// Calculate compression ratio
originalSize := len(points) * 16 // Assuming 16 bytes per point
compressedSize := len(compressed.CompressedValues)
compressionRatio := sfsTime.GetCompressionRatio(originalSize, compressedSize)
fmt.Printf("Compression ratio: %.2f\n", compressionRatio)
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

### Time Range Query Optimization

The time package now provides optimized time range query capabilities that integrate seamlessly with sfsDb:

#### Enhanced Time Range Query Example

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

// Create time range query options
options := sfsTime.NewTimeRangeQueryOptions(
    "timestamp",
    time.Now().Add(-1*time.Hour),
    time.Now(),
    sfsTime.TimeGranularityHour
)

// Execute time range query with optimized options
iter, err := sfsTime.SearchTimeRange(sensorTable, options)
if err != nil {
    return err
}
defer iter.Release()

// Get results
records := iter.GetRecords(true)
defer records.Release()

// Aggregate by hour with optimized time range
aggregationResults, err := sfsTime.TimeRangeQueryWithAggregation(
    sensorTable,
    options,
    "value",
    "sum"
)
if err != nil {
    return err
}

// Output aggregation results
fmt.Println("Hourly sum of sensor values:")
for _, result := range aggregationResults {
    fmt.Printf("Time: %s, Sum: %.2f\n", result.TimeKey, result.Value)
}
```

#### Time Range Query with Granularity

```go
// Execute time range query with specified granularity
iter, err := sfsTime.SearchTimeRangeWithGranularity(
    sensorTable,
    "timestamp",
    time.Now().Add(-24*time.Hour),
    time.Now(),
    sfsTime.TimeGranularityHour
)
if err != nil {
    return err
}
defer iter.Release()

// Process results...
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