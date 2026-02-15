package engine

import (
	"fmt"
	"testing"

	"github.com/liaoran123/sfsDb/storage"
)

// BenchmarkInsertPerformanceComparison 比较时序数据插入和普通批量插入的性能
func BenchmarkInsertPerformanceComparison(b *testing.B) {
	// 1. 初始化数据库
	dbManager := storage.GetDBManager()
	_, err := dbManager.OpenDB("./benchmark_comparison_db")
	if err != nil {
		b.Fatalf("打开数据库失败: %v", err)
	}
	defer dbManager.CloseDB()

	// 2. 创建测试表
	table, err := TableNew("performance_test")
	if err != nil {
		b.Fatalf("创建表失败: %v", err)
	}

	// 3. 设置字段
	fields := map[string]any{
		"id": 0,         // 普通ID作为主键
		"timestamp": 0,  // 时间戳
		"value": 0.0,    // 传感器值
	}
	err = table.SetFields(fields)
	if err != nil {
		b.Fatalf("设置字段失败: %v", err)
	}

	// 4. 创建主键索引
	primaryKey, err := DefaultPrimaryKeyNew("id")
	if err != nil {
		b.Fatalf("创建主键索引失败: %v", err)
	}
	primaryKey.AddFields("id")
	err = table.CreateIndex(primaryKey)
	if err != nil {
		b.Fatalf("创建索引失败: %v", err)
	}

	// 5. 准备测试数据
	testDataSize := 1000
	testData := make([]*map[string]any, testDataSize)
	for i := 0; i < testDataSize; i++ {
		testData[i] = &map[string]any{
			"id": i + 1,
			"timestamp": 1620000000 + int64(i*60),
			"value": float64(i % 100) / 10.0,
		}
	}

	// 6. 测试普通批量插入性能
	b.Run("BatchInsert", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// 创建InsertImpl实例
			batch := table.kvStore.GetBatch()
			insertImpl := NewInsertImpl(table, batch, false, &fields)
			
			_, err := insertImpl.BatchInsert(testData)
			if err != nil {
				b.Fatalf("批量插入失败: %v", err)
			}
		}
		b.StopTimer()
	})

	// 7. 修改表结构，使用timestamp作为主键
	// 先删除旧表
	table, err = TableNew("performance_test_time_series")
	if err != nil {
		b.Fatalf("创建表失败: %v", err)
	}

	// 设置新字段
	timeSeriesFields := map[string]any{
		"timestamp": 0,  // 时间戳作为主键
		"value": 0.0,    // 传感器值
	}
	err = table.SetFields(timeSeriesFields)
	if err != nil {
		b.Fatalf("设置字段失败: %v", err)
	}

	// 创建时间戳主键索引
	timePrimaryKey, err := DefaultPrimaryKeyNew("timestamp")
	if err != nil {
		b.Fatalf("创建主键索引失败: %v", err)
	}
	timePrimaryKey.AddFields("timestamp")
	err = table.CreateIndex(timePrimaryKey)
	if err != nil {
		b.Fatalf("创建索引失败: %v", err)
	}

	// 8. 准备时序数据测试数据
	timeSeriesTestData := make([]*map[string]any, testDataSize)
	for i := 0; i < testDataSize; i++ {
		timeSeriesTestData[i] = &map[string]any{
			"timestamp": 1620000000 + int64(i*60),
			"value": float64(i % 100) / 10.0,
		}
	}

	// 9. 测试时序数据批量插入性能
	b.Run("BatchInsertTimeSeries", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// 创建InsertImpl实例
			batch := table.kvStore.GetBatch()
			insertImpl := NewInsertImpl(table, batch, false, &timeSeriesFields)
			
			_, err := insertImpl.BatchInsertTimeSeries("timestamp", true, timeSeriesTestData)
			if err != nil {
				b.Fatalf("批量插入失败: %v", err)
			}
		}
		b.StopTimer()
	})
}

// TestInsertPerformanceComparison 测试并比较两种插入方式的性能
func TestInsertPerformanceComparison(t *testing.T) {
	// 1. 初始化数据库
	dbManager := storage.GetDBManager()
	_, err := dbManager.OpenDB("./test_comparison_db")
	if err != nil {
		t.Fatalf("打开数据库失败: %v", err)
	}
	defer dbManager.CloseDB()

	// 2. 创建普通表
	regularTable, err := TableNew("regular_test")
	if err != nil {
		t.Fatalf("创建表失败: %v", err)
	}

	// 3. 设置普通表字段
	regularFields := map[string]any{
		"id": 0,         // 普通ID作为主键
		"timestamp": 0,  // 时间戳
		"value": 0.0,    // 传感器值
	}
	err = regularTable.SetFields(regularFields)
	if err != nil {
		t.Fatalf("设置字段失败: %v", err)
	}

	// 4. 创建普通表主键索引
	regularPrimaryKey, err := DefaultPrimaryKeyNew("id")
	if err != nil {
		t.Fatalf("创建主键索引失败: %v", err)
	}
	regularPrimaryKey.AddFields("id")
	err = regularTable.CreateIndex(regularPrimaryKey)
	if err != nil {
		t.Fatalf("创建索引失败: %v", err)
	}

	// 5. 创建时序数据表
	timeSeriesTable, err := TableNew("time_series_test")
	if err != nil {
		t.Fatalf("创建表失败: %v", err)
	}

	// 6. 设置时序表字段
	timeSeriesFields := map[string]any{
		"timestamp": 0,  // 时间戳作为主键
		"value": 0.0,    // 传感器值
	}
	err = timeSeriesTable.SetFields(timeSeriesFields)
	if err != nil {
		t.Fatalf("设置字段失败: %v", err)
	}

	// 7. 创建时序表主键索引
	timePrimaryKey, err := DefaultPrimaryKeyNew("timestamp")
	if err != nil {
		t.Fatalf("创建主键索引失败: %v", err)
	}
	timePrimaryKey.AddFields("timestamp")
	err = timeSeriesTable.CreateIndex(timePrimaryKey)
	if err != nil {
		t.Fatalf("创建索引失败: %v", err)
	}

	// 8. 准备测试数据
	testDataSize := 1000
	regularTestData := make([]*map[string]any, testDataSize)
	timeSeriesTestData := make([]*map[string]any, testDataSize)

	for i := 0; i < testDataSize; i++ {
		regularTestData[i] = &map[string]any{
			"id": i + 1,
			"timestamp": 1620000000 + int64(i*60),
			"value": float64(i % 100) / 10.0,
		}

		timeSeriesTestData[i] = &map[string]any{
			"timestamp": 1620000000 + int64(i*60),
			"value": float64(i % 100) / 10.0,
		}
	}

	// 9. 测试普通批量插入
	fmt.Println("测试普通批量插入...")
	regularBatch := regularTable.kvStore.GetBatch()
	regularInsertImpl := NewInsertImpl(regularTable, regularBatch, false, &regularFields)

	regularIds, err := regularInsertImpl.BatchInsert(regularTestData)
	if err != nil {
		t.Fatalf("普通批量插入失败: %v", err)
	}
	fmt.Printf("普通批量插入成功，插入 %d 条记录\n", len(regularIds))

	// 10. 测试时序数据批量插入
	fmt.Println("测试时序数据批量插入...")
	timeSeriesBatch := timeSeriesTable.kvStore.GetBatch()
	timeSeriesInsertImpl := NewInsertImpl(timeSeriesTable, timeSeriesBatch, false, &timeSeriesFields)

	timeSeriesIds, err := timeSeriesInsertImpl.BatchInsertTimeSeries("timestamp", true, timeSeriesTestData)
	if err != nil {
		t.Fatalf("时序数据批量插入失败: %v", err)
	}
	fmt.Printf("时序数据批量插入成功，插入 %d 条记录\n", len(timeSeriesIds))

	// 11. 验证数据插入
	fmt.Println("验证数据插入...")

	// 验证普通表
	regularIter, err := regularTable.Search(&map[string]any{})
	if err != nil {
		t.Fatalf("查询普通表失败: %v", err)
	}
	regularRecords := regularIter.GetRecordSet(true)
	fmt.Printf("普通表中共有 %d 条记录\n", len(regularRecords))
	regularIter.Release()
	regularRecords.Release()

	// 验证时序表
	timeSeriesIter, err := timeSeriesTable.Search(&map[string]any{})
	if err != nil {
		t.Fatalf("查询时序表失败: %v", err)
	}
	timeSeriesRecords := timeSeriesIter.GetRecordSet(true)
	fmt.Printf("时序表中共有 %d 条记录\n", len(timeSeriesRecords))
	timeSeriesIter.Release()
	timeSeriesRecords.Release()

	fmt.Println("性能比较测试完成！")
}
