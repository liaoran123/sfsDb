package engine

import (
	"fmt"
	"testing"

	"github.com/liaoran123/sfsDb/storage"
)

// TestBatchInsertTimeSeries 测试时序数据批量插入功能
func TestBatchInsertTimeSeries(t *testing.T) {
	// 1. 初始化数据库
	dbManager := storage.GetDBManager()
	_, err := dbManager.OpenDB("./test_time_series_db")
	if err != nil {
		t.Fatalf("打开数据库失败: %v", err)
	}
	defer dbManager.CloseDB()

	// 2. 创建时序数据表
	table, err := TableNew("sensor_data")
	if err != nil {
		t.Fatalf("创建表失败: %v", err)
	}

	// 3. 设置字段
	fields := map[string]any{
		"timestamp": 0,  // 时间戳作为主键
		"temperature": 0.0, // 温度
		"humidity": 0.0,    // 湿度
		"pressure": 0.0,    // 气压
	}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("设置字段失败: %v", err)
	}

	// 4. 创建主键索引（使用timestamp作为主键）
	primaryKey, err := DefaultPrimaryKeyNew("timestamp")
	if err != nil {
		t.Fatalf("创建主键索引失败: %v", err)
	}
	primaryKey.AddFields("timestamp")
	err = table.CreateIndex(primaryKey)
	if err != nil {
		t.Fatalf("创建索引失败: %v", err)
	}

	// 5. 准备测试数据
	testData := []*map[string]any{
		{"timestamp": 1620000000, "temperature": 25.5, "humidity": 60.0, "pressure": 1013.25},
		{"timestamp": 1620000060, "temperature": 25.6, "humidity": 59.5, "pressure": 1013.30},
		{"timestamp": 1620000120, "temperature": 25.7, "humidity": 59.0, "pressure": 1013.35},
		{"timestamp": 1620000180, "temperature": 25.8, "humidity": 58.5, "pressure": 1013.40},
		{"timestamp": 1620000240, "temperature": 25.9, "humidity": 58.0, "pressure": 1013.45},
	}

	// 6. 创建InsertImpl实例
	batch := table.kvStore.GetBatch()
	insertImpl := NewInsertImpl(table, batch, false, &fields)

	// 7. 测试批量插入时序数据（跳过版本号）
	fmt.Println("测试批量插入时序数据（跳过版本号）...")
	ids, err := insertImpl.BatchInsertTimeSeries("timestamp", true, testData)
	if err != nil {
		t.Fatalf("批量插入失败: %v", err)
	}

	// 8. 验证插入结果
	if len(ids) != len(testData) {
		t.Fatalf("插入记录数不匹配，期望: %d, 实际: %d", len(testData), len(ids))
	}

	// 9. 验证返回的ID是否正确
	for i, id := range ids {
		expectedID := (*testData[i])["timestamp"].(int)
		if id != expectedID {
			t.Errorf("返回的ID不匹配，期望: %d, 实际: %d", expectedID, id)
		}
	}

	// 10. 测试批量插入时序数据（带版本号）
	fmt.Println("测试批量插入时序数据（带版本号）...")
	testDataWithVersion := []*map[string]any{
		{"timestamp": 1620000300, "temperature": 26.0, "humidity": 57.5, "pressure": 1013.50},
		{"timestamp": 1620000360, "temperature": 26.1, "humidity": 57.0, "pressure": 1013.55},
	}

	// 重新创建InsertImpl实例
	batch2 := table.kvStore.GetBatch()
	insertImpl2 := NewInsertImpl(table, batch2, false, &fields)
	
	idsWithVersion, err := insertImpl2.BatchInsertTimeSeries("timestamp", false, testDataWithVersion)
	if err != nil {
		t.Fatalf("批量插入失败: %v", err)
	}

	// 11. 验证带版本号的插入结果
	if len(idsWithVersion) != len(testDataWithVersion) {
		t.Fatalf("插入记录数不匹配，期望: %d, 实际: %d", len(testDataWithVersion), len(idsWithVersion))
	}

	// 12. 查询数据验证
	fmt.Println("验证插入的数据...")
	iter, err := table.Search(&map[string]any{}) 
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	defer iter.Release()

	records := iter.GetRecordSet(true)
	defer records.Release()

	if len(records) != len(testData)+len(testDataWithVersion) {
		t.Fatalf("查询到的记录数不匹配，期望: %d, 实际: %d", len(testData)+len(testDataWithVersion), len(records))
	}

	fmt.Printf("测试成功！共插入 %d 条时序数据\n", len(testData)+len(testDataWithVersion))
	for i, record := range records {
		fmt.Printf("记录 %d: %v\n", i+1, record)
	}
}

// BenchmarkBatchInsertTimeSeries 基准测试时序数据批量插入性能
func BenchmarkBatchInsertTimeSeries(b *testing.B) {
	// 1. 初始化数据库
	dbManager := storage.GetDBManager()
	_, err := dbManager.OpenDB("./benchmark_time_series_db")
	if err != nil {
		b.Fatalf("打开数据库失败: %v", err)
	}
	defer dbManager.CloseDB()

	// 2. 创建时序数据表
	table, err := TableNew("benchmark_sensor_data")
	if err != nil {
		b.Fatalf("创建表失败: %v", err)
	}

	// 3. 设置字段
	fields := map[string]any{
		"timestamp": 0,  // 时间戳作为主键
		"value": 0.0,    // 传感器值
	}
	err = table.SetFields(fields)
	if err != nil {
		b.Fatalf("设置字段失败: %v", err)
	}

	// 4. 创建主键索引
	primaryKey, err := DefaultPrimaryKeyNew("timestamp")
	if err != nil {
		b.Fatalf("创建主键索引失败: %v", err)
	}
	primaryKey.AddFields("timestamp")
	err = table.CreateIndex(primaryKey)
	if err != nil {
		b.Fatalf("创建索引失败: %v", err)
	}

	// 5. 准备测试数据
	testData := make([]*map[string]any, 1000)
	for i := 0; i < 1000; i++ {
		timestamp := 1620000000 + int64(i*60) // 每分钟一条数据
		value := float64(i % 100) / 10.0     // 模拟传感器值
		testData[i] = &map[string]any{
			"timestamp": timestamp,
			"value":     value,
		}
	}

	// 6. 基准测试
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 创建InsertImpl实例
		batch := table.kvStore.GetBatch()
		insertImpl := NewInsertImpl(table, batch, false, &fields)
		
		_, err := insertImpl.BatchInsertTimeSeries("timestamp", true, testData)
		if err != nil {
			b.Fatalf("批量插入失败: %v", err)
		}
	}
	b.StopTimer()

	// 7. 计算性能
	ops := float64(b.N * len(testData)) / b.Elapsed().Seconds()
	fmt.Printf("时序数据批量插入性能: %.2f ops/sec\n", ops)
}
