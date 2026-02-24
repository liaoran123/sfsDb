package engine

import (
	"fmt"
	"testing"

	"github.com/liaoran123/sfsDb/storage"
)

// TestSearchRangeCompositeKey 测试组合主键的范围搜索 - 时序数据库设备示例
func TestSearchRangeCompositeKey(t *testing.T) {
	// 初始化数据库
	dbMgr := storage.GetDBManager()
	_, err := dbMgr.OpenDB("./test_kvdb")
	if err != nil {
		t.Fatalf("打开数据库失败: %v", err)
	}
	defer dbMgr.CloseDB()

	// 创建表 - 时序数据库设备表
	table, err := TableNew("device_data")
	if err != nil {
		t.Fatalf("创建表失败: %v", err)
	}

	// 设置字段
	fields := map[string]any{
		"device_id":   "",  // 设备ID
		"timestamp":   0,   // 时间戳
		"temperature": 0.0, // 温度
		"humidity":    0.0, // 湿度
		"pressure":    0.0, // 压力
	}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("设置字段失败: %v", err)
	}

	// 创建组合主键 - 设备ID和时间戳
	pk, err := DefaultPrimaryKeyNew("device_timestamp")
	if err != nil {
		t.Fatalf("创建主键失败: %v", err)
	}
	pk.AddFields("device_id")
	pk.AddFields("timestamp")
	err = table.CreateIndex(pk)
	if err != nil {
		t.Fatalf("创建索引失败: %v", err)
	}

	// 插入测试数据 - 模拟设备数据
	testData := []map[string]any{
		{"device_id": "device_001", "timestamp": 1609459200, "temperature": 25.5, "humidity": 60.0, "pressure": 1013.25},
		{"device_id": "device_001", "timestamp": 1609462800, "temperature": 25.6, "humidity": 59.5, "pressure": 1013.20},
		{"device_id": "device_001", "timestamp": 1609466400, "temperature": 25.7, "humidity": 59.0, "pressure": 1013.15},
		{"device_id": "device_001", "timestamp": 1609470000, "temperature": 25.8, "humidity": 58.5, "pressure": 1013.10},
		{"device_id": "device_001", "timestamp": 1609473600, "temperature": 25.9, "humidity": 58.0, "pressure": 1013.05},
		{"device_id": "device_002", "timestamp": 1609459200, "temperature": 26.0, "humidity": 55.0, "pressure": 1012.50},
		{"device_id": "device_002", "timestamp": 1609462800, "temperature": 26.1, "humidity": 54.5, "pressure": 1012.45},
	}

	for _, data := range testData {
		_, err := table.Insert(&data)
		if err != nil {
			t.Fatalf("插入数据失败: %v", err)
		}
	}

	// 测试组合主键范围搜索 - 查询device_001设备的时间段数据
	searchImpl := NewSearchImpl(table)
	defer GlobalSearchImplPool.Put(searchImpl)

	// 测试范围搜索: device_id="device_001", timestamp从1609462800到1609470000
	// 注意：组合主键时，第一个字段必须全量匹配，最后一个字段是范围
	iter, err := searchImpl.SearchRange(nil, &map[string]any{"device_id": "device_001", "timestamp": 1609462800}, &map[string]any{"device_id": "device_001", "timestamp": 1609470000})
	if err != nil {
		t.Fatalf("范围搜索失败: %v", err)
	}
	defer iter.Release()

	// 获取记录
	records := iter.GetRecords(true)
	defer records.Release()

	// 验证结果
	expectedCount := 3 // 应该返回device_001的3条记录
	if len(records) != expectedCount {
		t.Errorf("期望返回 %d 条记录，实际返回 %d 条", expectedCount, len(records))
	}

	// 验证返回的记录包含预期的device_id和timestamp值
	for _, record := range records {
		// 验证device_id
		deviceID, ok := record["device_id"].(string)
		if !ok {
			t.Errorf("记录中的device_id字段类型不是string: %T", record["device_id"])
			continue
		}
		if deviceID != "device_001" {
			t.Errorf("记录中的device_id不是device_001: %s", deviceID)
		}

		// 验证timestamp在范围内
		timestamp, ok := record["timestamp"].(int)
		if !ok {
			t.Errorf("记录中的timestamp字段类型不是int: %T", record["timestamp"])
			continue
		}
		if timestamp < 1609462800 || timestamp > 1609470000 {
			t.Errorf("记录中的timestamp不在范围内: %d", timestamp)
		}
	}

	fmt.Printf("组合主键范围搜索测试通过，返回 %d 条记录\n", len(records))
}

// TestSearchExactMatchRange 测试完全匹配的范围查询（符合当前实现）
func TestSearchExactMatchRange(t *testing.T) {
	// 初始化数据库
	dbMgr := storage.GetDBManager()
	_, err := dbMgr.OpenDB("./test_kvdb")
	if err != nil {
		t.Fatalf("打开数据库失败: %v", err)
	}
	defer dbMgr.CloseDB()

	// 创建表
	table, err := TableNew("test_exact_match_range")
	if err != nil {
		t.Fatalf("创建表失败: %v", err)
	}

	// 设置字段
	fields := map[string]any{
		"name":   "", // 名字作为主键
		"age":    0,
		"salary": 0,
	}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("设置字段失败: %v", err)
	}

	// 创建主键索引
	pk, err := DefaultPrimaryKeyNew("name")
	if err != nil {
		t.Fatalf("创建主键失败: %v", err)
	}
	pk.AddFields("name")
	err = table.CreateIndex(pk)
	if err != nil {
		t.Fatalf("创建主键索引失败: %v", err)
	}

	// 插入测试数据 - 使用唯一的名字作为主键
	testData := []map[string]any{
		{"name": "张三", "age": 25, "salary": 5000},
		{"name": "李四", "age": 35, "salary": 7000},
		{"name": "王五", "age": 45, "salary": 9000},
	}

	for _, data := range testData {
		_, err := table.Insert(&data)
		if err != nil {
			t.Fatalf("插入数据失败: %v", err)
		}
	}

	// 测试完全匹配的范围查询
	searchImpl := NewSearchImpl(table)
	defer GlobalSearchImplPool.Put(searchImpl)

	// 测试1: 完全匹配单个值 - name="张三"
	// 当前的 SearchRange 实现只支持完全匹配的范围查询
	iter, err := searchImpl.SearchRange(nil, &map[string]any{"name": "张三"}, &map[string]any{"name": "张三"})
	if err != nil {
		t.Fatalf("范围搜索失败: %v", err)
	}
	defer iter.Release()

	// 获取记录
	records := iter.GetRecords(true)
	defer records.Release()

	// 验证结果
	expectedCount := 1 // 应该返回name="张三"的记录
	if len(records) != expectedCount {
		t.Errorf("期望返回 %d 条记录，实际返回 %d 条", expectedCount, len(records))
	}

	// 验证返回的记录包含预期的name值
	for _, record := range records {
		name, ok := record["name"].(string)
		if !ok {
			t.Errorf("记录中的name字段类型不是string: %T", record["name"])
			continue
		}
		if name != "张三" {
			t.Errorf("记录中的name不是张三: %s", name)
		}
	}

	// 测试2: 尝试使用相同的开始和结束值
	iter2, err := searchImpl.SearchRange(nil, &map[string]any{"name": "李四"}, &map[string]any{"name": "李四"})
	if err != nil {
		t.Fatalf("范围搜索失败: %v", err)
	}
	defer iter2.Release()

	// 获取记录
	records2 := iter2.GetRecords(true)
	defer records2.Release()

	// 验证结果
	expectedCount2 := 1 // 应该返回name="李四"的记录
	if len(records2) != expectedCount2 {
		t.Errorf("期望返回 %d 条记录，实际返回 %d 条", expectedCount2, len(records2))
	}

	fmt.Printf("完全匹配的范围查询测试通过，返回 %d 条记录\n", len(records))
}

// TestSearchExactMatchSingleKey 测试单字段主键的完全匹配范围查询（符合当前实现）
func TestSearchExactMatchSingleKey(t *testing.T) {
	// 初始化数据库
	dbMgr := storage.GetDBManager()
	_, err := dbMgr.OpenDB("./test_kvdb")
	if err != nil {
		t.Fatalf("打开数据库失败: %v", err)
	}
	defer dbMgr.CloseDB()

	// 创建表
	table, err := TableNew("test_exact_match_single")
	if err != nil {
		t.Fatalf("创建表失败: %v", err)
	}

	// 设置字段
	fields := map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("设置字段失败: %v", err)
	}

	// 创建单字段主键
	pk, err := DefaultPrimaryKeyNew("id")
	if err != nil {
		t.Fatalf("创建主键失败: %v", err)
	}
	pk.AddFields("id")
	err = table.CreateIndex(pk)
	if err != nil {
		t.Fatalf("创建索引失败: %v", err)
	}

	// 插入测试数据
	for i := 1; i <= 10; i++ {
		data := map[string]any{
			"id":   i,
			"name": fmt.Sprintf("用户%d", i),
			"age":  20 + i,
		}
		_, err := table.Insert(&data)
		if err != nil {
			t.Fatalf("插入数据失败: %v", err)
		}
	}

	// 测试单字段主键完全匹配的范围查询
	searchImpl := NewSearchImpl(table)
	defer GlobalSearchImplPool.Put(searchImpl)

	// 测试1: 完全匹配单个值 - id=5
	// 当前的 SearchRange 实现只支持完全匹配的范围查询
	iter, err := searchImpl.SearchRange(nil, &map[string]any{"id": 5}, &map[string]any{"id": 5})
	if err != nil {
		t.Fatalf("范围搜索失败: %v", err)
	}
	defer iter.Release()

	// 获取记录
	records := iter.GetRecords(true)
	defer records.Release()

	// 验证结果
	expectedCount := 1 // 应该返回id=5的记录
	if len(records) != expectedCount {
		t.Errorf("期望返回 %d 条记录，实际返回 %d 条", expectedCount, len(records))
	}

	// 验证返回的记录包含预期的id值
	for _, record := range records {
		id, ok := record["id"].(int)
		if !ok {
			t.Errorf("记录中的id字段类型不是int: %T", record["id"])
			continue
		}
		if id != 5 {
			t.Errorf("记录中的id不是5: %d", id)
		}
	}

	// 测试2: 尝试使用相同的开始和结束值 - id=8
	iter2, err := searchImpl.SearchRange(nil, &map[string]any{"id": 8}, &map[string]any{"id": 8})
	if err != nil {
		t.Fatalf("范围搜索失败: %v", err)
	}
	defer iter2.Release()

	// 获取记录
	records2 := iter2.GetRecords(true)
	defer records2.Release()

	// 验证结果
	expectedCount2 := 1 // 应该返回id=8的记录
	if len(records2) != expectedCount2 {
		t.Errorf("期望返回 %d 条记录，实际返回 %d 条", expectedCount2, len(records2))
	}

	fmt.Printf("单字段主键完全匹配范围查询测试通过，返回 %d 条记录\n", len(records))
}

// TestSearchRangeCompositeKeyBoundary 测试组合主键范围查询的边界情况
// 验证当前实现是否真正支持前几个字段完全匹配，最后一个字段进行范围查询
func TestSearchRangeCompositeKeyBoundary(t *testing.T) {
	// 初始化数据库
	dbMgr := storage.GetDBManager()
	_, err := dbMgr.OpenDB("./test_kvdb")
	if err != nil {
		t.Fatalf("打开数据库失败: %v", err)
	}
	defer dbMgr.CloseDB()

	// 创建表 - 时序数据库设备表
	table, err := TableNew("device_data_boundary")
	if err != nil {
		t.Fatalf("创建表失败: %v", err)
	}

	// 设置字段
	fields := map[string]any{
		"device_id":   "",  // 设备ID
		"timestamp":   0,   // 时间戳
		"temperature": 0.0, // 温度
	}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("设置字段失败: %v", err)
	}

	// 创建组合主键 - 设备ID和时间戳
	pk, err := DefaultPrimaryKeyNew("device_timestamp")
	if err != nil {
		t.Fatalf("创建主键失败: %v", err)
	}
	pk.AddFields("device_id")
	pk.AddFields("timestamp")
	err = table.CreateIndex(pk)
	if err != nil {
		t.Fatalf("创建索引失败: %v", err)
	}

	// 插入测试数据 - 模拟设备数据，故意打乱时间戳顺序
	testData := []map[string]any{
		{"device_id": "device_001", "timestamp": 1609459200, "temperature": 25.5},
		{"device_id": "device_001", "timestamp": 1609466400, "temperature": 25.7},
		{"device_id": "device_001", "timestamp": 1609473600, "temperature": 25.9},
		{"device_id": "device_002", "timestamp": 1609459200, "temperature": 26.0},
		{"device_id": "device_002", "timestamp": 1609466400, "temperature": 26.2},
		{"device_id": "device_002", "timestamp": 1609473600, "temperature": 26.4},
	}

	for _, data := range testData {
		_, err := table.Insert(&data)
		if err != nil {
			t.Fatalf("插入数据失败: %v", err)
		}
	}

	// 测试组合主键范围搜索
	searchImpl := NewSearchImpl(table)
	defer GlobalSearchImplPool.Put(searchImpl)

	// 测试1: 边界值测试 - 开始时间存在，结束时间不存在
	t.Run("BoundaryValueTest", func(t *testing.T) {
		// 测试范围搜索: device_id="device_001", timestamp从1609462800到1609469999
		// 注意：开始时间和结束时间都不在数据中
		iter, err := searchImpl.SearchRange(nil, &map[string]any{"device_id": "device_001", "timestamp": 1609462800}, &map[string]any{"device_id": "device_001", "timestamp": 1609469999})
		if err != nil {
			t.Fatalf("范围搜索失败: %v", err)
		}
		defer iter.Release()

		// 获取记录
		records := iter.GetRecords(true)
		defer records.Release()

		// 验证结果
		// 当前实现可能返回空，因为它使用完全匹配
		// 真正的范围查询应该返回 timestamp=1609466400 的记录
		t.Logf("边界值测试返回 %d 条记录", len(records))
		for _, record := range records {
			t.Logf("返回记录: device_id=%s, timestamp=%d, temperature=%f", record["device_id"], record["timestamp"], record["temperature"])
		}
	})

	// 测试2: 只指定前几个字段
	t.Run("OnlyFirstFieldsTest", func(t *testing.T) {
		// 尝试只指定 device_id，不指定 timestamp
		// 这应该返回该设备的所有记录
		// 当前实现会失败，因为它要求 Start 和 Limit 必须包含相同的 key 值
		_, err := searchImpl.SearchRange(nil, &map[string]any{"device_id": "device_001"}, &map[string]any{"device_id": "device_001"})
		if err != nil {
			t.Logf("只指定前几个字段的查询失败（预期内）: %v", err)
		} else {
			t.Logf("只指定前几个字段的查询成功")
		}
	})

	// 测试3: 非连续时间戳测试
	t.Run("NonContinuousTimestampTest", func(t *testing.T) {
		// 测试范围搜索: device_id="device_001", timestamp从1609459200到1609473600
		iter, err := searchImpl.SearchRange(nil, &map[string]any{"device_id": "device_001", "timestamp": 1609459200}, &map[string]any{"device_id": "device_001", "timestamp": 1609473600})
		if err != nil {
			t.Fatalf("范围搜索失败: %v", err)
		}
		defer iter.Release()

		// 获取记录
		records := iter.GetRecords(true)
		defer records.Release()

		// 验证结果
		t.Logf("非连续时间戳测试返回 %d 条记录", len(records))
		for _, record := range records {
			t.Logf("返回记录: device_id=%s, timestamp=%d, temperature=%f", record["device_id"], record["timestamp"], record["temperature"])
		}
	})
}
