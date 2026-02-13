package engine

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/liaoran123/sfsDb/storage"
)

// TestTable_SearchRange 测试时序数据的区间搜索功能
func TestTable_SearchRange(t *testing.T) {
	// 清理旧数据
	tableName := "test_search_range_" + t.Name()
	defer os.RemoveAll("./test_search_range_db")

	// 初始化数据库
	_, err := storage.OpenDefaultDb("./test_search_range_db")
	if err != nil {
		t.Fatalf("打开数据库失败: %v", err)
	}
	defer storage.CloseDb()

	// 创建表，使用时间戳作为主键
	table, err := TableNew(tableName)
	if err != nil {
		t.Fatalf("创建表失败: %v", err)
	}

	// 设置字段，timestamp作为主键（int）
	fields := map[string]any{
		"timestamp": 0,   // 时间戳主键
		"value":     0.0, // 传感器值
		"sensor_id": "",  // 传感器ID
	}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("设置字段失败: %v", err)
	}

	// 创建主键索引
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("timestamp")
	err = table.CreateIndex(pk)
	if err != nil {
		t.Fatalf("创建主键索引失败: %v", err)
	}

	// 生成测试数据（时序数据）
	now := int(time.Now().Unix())
	testData := make([]map[string]any, 100)

	for i := 0; i < 100; i++ {
		testData[i] = map[string]any{
			"timestamp": now + i,
			"value":     float64(i * 10),
			"sensor_id": fmt.Sprintf("sensor_%d", i%10),
		}
	}

	// 插入测试数据
	for _, data := range testData {
		_, err := table.Insert(&data)
		if err != nil {
			t.Fatalf("插入数据失败: %v", err)
		}
	}

	// 测试1: 基本区间搜索
	t.Run("BasicRangeSearch", func(t *testing.T) {
		// 创建范围
		rangeStart := now + 10
		rangeEnd := now + 20

		// 定义迭代器函数
		funIter := storage.FunIter(func(start, limit []byte) storage.Iterator {
			// 使用存储引擎的Iterator方法创建范围迭代器
			return table.kvStore.Iterator(start, limit)
		})

		// 执行区间搜索
		iter, err := table.SearchRange(funIter, "timestamp", rangeStart, rangeEnd)
		if err != nil {
			t.Fatalf("SearchRange 失败: %v", err)
		}
		defer iter.Release()

		// 获取结果
		records := iter.GetRecords(true)
		defer records.Release()

		// 验证结果数量
		expectedCount := 10 // 10-19 共10条数据
		if len(records) != expectedCount {
			t.Errorf("期望 %d 条记录，实际得到 %d 条", expectedCount, len(records))
		}

		// 验证结果范围
		for _, record := range records {
			timestamp := record["timestamp"].(int)
			fmt.Printf("10-19 共10条数据 timestamp: %v\n", timestamp)
			if timestamp < rangeStart || timestamp >= rangeEnd {
				t.Errorf("记录时间戳 %d 超出范围 [%d, %d]", timestamp, rangeStart, rangeEnd)
			}
		}
	})

	// 测试2: 空范围搜索
	t.Run("EmptyRangeSearch", func(t *testing.T) {
		// 创建范围
		rangeStart := now + 200
		rangeEnd := now + 210

		// 定义迭代器函数
		funIter := storage.FunIter(func(start, limit []byte) storage.Iterator {
			// 使用存储引擎的Iterator方法创建范围迭代器
			return table.kvStore.Iterator(start, limit)
		})

		// 执行区间搜索
		iter, err := table.SearchRange(funIter, "timestamp", rangeStart, rangeEnd)
		if err != nil {
			t.Fatalf("SearchRange 失败: %v", err)
		}
		defer iter.Release()

		// 获取结果
		records := iter.GetRecords(true)
		defer records.Release()

		// 验证结果为空
		if len(records) != 0 {
			t.Errorf("期望空结果，实际得到 %d 条记录", len(records))
		}
	})

	// 测试3: 边界值测试
	t.Run("BoundaryValues", func(t *testing.T) {
		// 创建范围
		rangeStart := now
		rangeEnd := now + 1

		// 定义迭代器函数
		funIter := storage.FunIter(func(start, limit []byte) storage.Iterator {
			// 使用存储引擎的Iterator方法创建范围迭代器
			return table.kvStore.Iterator(start, limit)
		})

		// 执行区间搜索
		iter, err := table.SearchRange(funIter, "timestamp", rangeStart, rangeEnd)
		if err != nil {
			t.Fatalf("SearchRange 失败: %v", err)
		}
		defer iter.Release()

		// 获取结果
		records := iter.GetRecords(true)
		defer records.Release()

		// 验证结果
		if len(records) != 0 {
			t.Logf("边界值测试得到 %d 条记录", len(records))
		}
	})
}

// TestTable_SearchRange_PrimaryKey 测试当索引为主键时，SearchRange 函数的行为是否符合预期
// 验证：Start=nil表示从表的主键值最小值开始，Limit=nil表示到表的主键值最大值结束，同时为nil即表示遍历全表
func TestTable_SearchRange_PrimaryKey(t *testing.T) {
	// 清理旧数据
	tableName := "test_search_range_pk_" + t.Name()
	defer os.RemoveAll("./test_search_range_pk_db")

	// 初始化数据库
	_, err := storage.OpenDefaultDb("./test_search_range_pk_db")
	if err != nil {
		t.Fatalf("打开数据库失败: %v", err)
	}
	defer storage.CloseDb()

	// 创建表
	table, err := TableNew(tableName)
	if err != nil {
		t.Fatalf("创建表失败: %v", err)
	}

	// 设置字段
	fields := map[string]any{
		"id":    0,   // 主键
		"name":  "",  // 名称
		"value": 0.0, // 值
	}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("设置字段失败: %v", err)
	}

	// 创建主键索引
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("id")
	err = table.CreateIndex(pk)
	if err != nil {
		t.Fatalf("创建主键索引失败: %v", err)
	}

	// 生成测试数据
	testData := make([]map[string]any, 100)

	for i := 0; i < 100; i++ {
		testData[i] = map[string]any{
			"id":    i,
			"name":  fmt.Sprintf("item_%d", i),
			"value": float64(i * 10),
		}
	}

	// 插入测试数据
	for _, data := range testData {
		_, err := table.Insert(&data)
		if err != nil {
			t.Fatalf("插入数据失败: %v", err)
		}
	}

	// 定义迭代器函数
	funIter := storage.FunIter(func(start, limit []byte) storage.Iterator {
		return table.kvStore.Iterator(start, limit)
	})

	// 测试1: Start=nil 表示从表的主键值最小值开始
	t.Run("StartNilFromMinPrimaryKey", func(t *testing.T) {
		// 设置搜索范围：Start=nil, Limit=50
		rangeEnd := 50

		// 执行区间搜索
		iter, err := table.SearchRange(funIter, "id", nil, rangeEnd)
		if err != nil {
			t.Fatalf("SearchRange 失败: %v", err)
		}
		defer iter.Release()

		// 获取结果
		records := iter.GetRecords(true)
		defer records.Release()

		// 验证结果数量
		expectedCount := 50 // 0-49 共50条数据
		if len(records) != expectedCount {
			t.Errorf("期望 %d 条记录，实际得到 %d 条", expectedCount, len(records))
		}

		// 验证结果从最小值开始
		if len(records) > 0 {
			firstID := records[0]["id"].(int)
			if firstID != 0 {
				t.Errorf("期望从最小值 0 开始，实际从 %d 开始", firstID)
			}
		}
	})

	// 测试2: Limit=nil 表示到表的主键值最大值结束
	t.Run("LimitNilToMaxPrimaryKey", func(t *testing.T) {
		// 设置搜索范围：Start=50, Limit=nil
		rangeStart := 50

		// 执行区间搜索
		iter, err := table.SearchRange(funIter, "id", rangeStart, nil)
		if err != nil {
			t.Fatalf("SearchRange 失败: %v", err)
		}
		defer iter.Release()

		// 获取结果
		records := iter.GetRecords(true)
		defer records.Release()

		// 验证结果数量
		expectedCount := 50 // 50-99 共50条数据
		if len(records) != expectedCount {
			t.Errorf("期望 %d 条记录，实际得到 %d 条", expectedCount, len(records))
		}

		// 验证结果到最大值结束
		if len(records) > 0 {
			lastID := records[len(records)-1]["id"].(int)
			if lastID != 99 {
				t.Errorf("期望到最大值 99 结束，实际到 %d 结束", lastID)
			}
		}
	})

	// 测试3: 同时为nil 即表示遍历全表
	t.Run("BothNilFullTableScan", func(t *testing.T) {
		// 执行区间搜索
		iter, err := table.SearchRange(funIter, "id", nil, nil)
		if err != nil {
			t.Fatalf("SearchRange 失败: %v", err)
		}
		defer iter.Release()

		// 获取结果
		records := iter.GetRecords(true)
		defer records.Release()

		// 验证结果数量
		expectedCount := 100 // 0-99 共100条数据
		if len(records) != expectedCount {
			t.Errorf("期望 %d 条记录，实际得到 %d 条", expectedCount, len(records))
		}

		// 验证结果包含所有数据
		if len(records) > 0 {
			firstID := records[0]["id"].(int)
			lastID := records[len(records)-1]["id"].(int)
			if firstID != 0 || lastID != 99 {
				t.Errorf("期望全表扫描包含所有数据 [0, 99]，实际包含 [%d, %d]",
					firstID, lastID)
			}
		}
	})
}

// TestTable_SearchRange_CommentValidation 测试 SearchRange 函数的行为是否符合注释中描述的预期
// 验证：Start=nil表示从索引最小值开始，Limit=nil表示到索引最大值结束，同时为nil即表示遍历索引
func TestTable_SearchRange_CommentValidation(t *testing.T) {
	// 清理旧数据
	tableName := "test_search_range_comment_" + t.Name()
	defer os.RemoveAll("./test_search_range_comment_db")

	// 初始化数据库
	_, err := storage.OpenDefaultDb("./test_search_range_comment_db")
	if err != nil {
		t.Fatalf("打开数据库失败: %v", err)
	}
	defer storage.CloseDb()

	// 创建表
	table, err := TableNew(tableName)
	if err != nil {
		t.Fatalf("创建表失败: %v", err)
	}

	// 设置字段
	fields := map[string]any{
		"id":    0,   // 主键
		"name":  "",  // 名称
		"value": 0.0, // 值
	}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("设置字段失败: %v", err)
	}

	// 创建主键索引
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("id")
	err = table.CreateIndex(pk)
	if err != nil {
		t.Fatalf("创建主键索引失败: %v", err)
	}

	// 生成测试数据
	testData := make([]map[string]any, 100)

	for i := 0; i < 100; i++ {
		testData[i] = map[string]any{
			"id":    i,
			"name":  fmt.Sprintf("item_%d", i),
			"value": float64(i * 10),
		}
	}

	// 插入测试数据
	for _, data := range testData {
		_, err := table.Insert(&data)
		if err != nil {
			t.Fatalf("插入数据失败: %v", err)
		}
	}

	// 定义迭代器函数
	funIter := storage.FunIter(func(start, limit []byte) storage.Iterator {
		return table.kvStore.Iterator(start, limit)
	})

	// 测试1: Start=nil 表示从索引最小值开始
	t.Run("StartNilFromMin", func(t *testing.T) {
		// 设置搜索范围：Start=nil, Limit=50
		rangeEnd := 50

		// 执行区间搜索
		iter, err := table.SearchRange(funIter, "id", nil, rangeEnd)
		if err != nil {
			t.Fatalf("SearchRange 失败: %v", err)
		}
		defer iter.Release()

		// 获取结果
		records := iter.GetRecords(true)
		defer records.Release()

		// 验证结果从最小值开始
		if len(records) > 0 {
			firstID := records[0]["id"].(int)
			if firstID != 0 {
				t.Errorf("期望从最小值 0 开始，实际从 %d 开始", firstID)
			}
		}
	})

	// 测试2: Limit=nil 表示到索引最大值结束
	t.Run("LimitNilToMax", func(t *testing.T) {
		// 设置搜索范围：Start=50, Limit=nil
		rangeStart := 50

		// 执行区间搜索
		iter, err := table.SearchRange(funIter, "id", rangeStart, nil)
		if err != nil {
			t.Fatalf("SearchRange 失败: %v", err)
		}
		defer iter.Release()

		// 获取结果
		records := iter.GetRecords(true)
		defer records.Release()

		// 验证结果到最大值结束
		if len(records) > 0 {
			lastID := records[len(records)-1]["id"].(int)
			if lastID != 99 {
				t.Errorf("期望到最大值 99 结束，实际到 %d 结束", lastID)
			}
		}
	})

	// 测试3: 同时为nil 即表示遍历索引
	t.Run("BothNilScanIndex", func(t *testing.T) {
		// 执行区间搜索
		iter, err := table.SearchRange(funIter, "id", nil, nil)
		if err != nil {
			t.Fatalf("SearchRange 失败: %v", err)
		}
		defer iter.Release()

		// 获取结果
		records := iter.GetRecords(true)
		defer records.Release()

		// 验证结果包含所有数据
		if len(records) > 0 {
			firstID := records[0]["id"].(int)
			lastID := records[len(records)-1]["id"].(int)
			if firstID != 0 || lastID != 99 {
				t.Errorf("期望遍历索引包含所有数据 [0, 99]，实际包含 [%d, %d]",
					firstID, lastID)
			}
		}
	})
}

// TestTable_SearchRange_Performance 测试时序数据区间搜索的性能
func TestTable_SearchRange_Performance(t *testing.T) {
	// 跳过性能测试，除非明确指定
	if testing.Short() {
		t.Skip("跳过性能测试")
	}

	// 清理旧数据
	tableName := "test_search_range_perf_" + t.Name()
	defer os.RemoveAll("./test_search_range_perf_db")

	// 初始化数据库
	_, err := storage.OpenDefaultDb("./test_search_range_perf_db")
	if err != nil {
		t.Fatalf("打开数据库失败: %v", err)
	}
	defer storage.CloseDb()

	// 创建表
	table, err := TableNew(tableName)
	if err != nil {
		t.Fatalf("创建表失败: %v", err)
	}

	// 设置字段
	fields := map[string]any{
		"timestamp": 0,
		"value":     0.0,
		"sensor_id": "",
	}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("设置字段失败: %v", err)
	}

	// 创建主键索引
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("timestamp")
	err = table.CreateIndex(pk)
	if err != nil {
		t.Fatalf("创建主键索引失败: %v", err)
	}

	// 生成大量测试数据
	now := int(time.Now().Unix())
	dataCount := 10000
	testData := make([]map[string]any, dataCount)

	for i := 0; i < dataCount; i++ {
		testData[i] = map[string]any{
			"timestamp": now + i,
			"value":     float64(i * 10),
			"sensor_id": fmt.Sprintf("sensor_%d", i%100),
		}
	}

	// 插入测试数据
	startTime := time.Now()
	for _, data := range testData {
		_, err := table.Insert(&data)
		if err != nil {
			t.Fatalf("插入数据失败: %v", err)
		}
	}
	insertTime := time.Since(startTime)
	t.Logf("插入 %d 条数据耗时: %v", dataCount, insertTime)

	// 测试性能
	t.Run("PerformanceTest", func(t *testing.T) {
		// 创建范围
		rangeStart := now + 1000
		rangeEnd := now + 2000

		// 定义迭代器函数
		funIter := storage.FunIter(func(start, limit []byte) storage.Iterator {
			// 使用存储引擎的Iterator方法创建范围迭代器
			return table.kvStore.Iterator(start, limit)
		})

		// 执行多次搜索，测试性能
		iterations := 10
		totalTime := 0

		for i := 0; i < iterations; i++ {
			startTime := time.Now()

			// 执行区间搜索
			iter, err := table.SearchRange(funIter, "timestamp", rangeStart, rangeEnd)
			if err != nil {
				t.Fatalf("SearchRange 失败: %v", err)
			}

			// 获取结果
			records := iter.GetRecords(true)

			// 释放资源
			records.Release()
			iter.Release()

			searchTime := time.Since(startTime).Milliseconds()
			totalTime += int(searchTime)
			t.Logf("第 %d 次搜索耗时: %d ms", i+1, searchTime)
		}

		averageTime := totalTime / iterations
		t.Logf("平均搜索耗时: %d ms", averageTime)

		// 性能阈值检查
		if averageTime > 100 {
			t.Logf("搜索性能可能需要优化，平均耗时 %d ms", averageTime)
		}
	})
}
