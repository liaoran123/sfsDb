package engine

import (
	"testing"

	"github.com/liaoran123/sfsDb/storage"
)

// TestBatchSharingExample 展示如何使用共享的 batch 操作来实现最终一致性
// 这是一个用户示例，演示了在边缘计算场景中如何优化数据同步
func TestBatchSharingExample(t *testing.T) {
	// 1. 创建表
	table, err := TableNew("test_batch_sharing")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 2. 定义表字段
	fields := map[string]any{
		"id":          0,
		"device_id":   "",
		"temperature": 0.0,
		"humidity":    0,
		"timestamp":   "",
	}

	// 3. 设置表字段
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 4. 创建主键索引，确保数据唯一性
	err = table.CreatePrimaryKey("id")
	if err != nil {
		t.Fatalf("Failed to create primary key: %v", err)
	}
	t.Log("Created primary key index on id field")

	// 4. 创建共享的 batch（用于实现最终一致性）
	// 在实际应用中，这个 batch 可以在多个操作之间共享
	// 例如，在边缘设备收集数据后，批量同步到云端
	kvStore := storage.GetDBManager().GetDB()
	batch := kvStore.GetBatch()

	// 5. 模拟边缘设备收集的多条数据
	// 这些数据可以来自不同的传感器或时间点
	sensorData1 := map[string]any{
		"id":          1,
		"device_id":   "sensor-001",
		"temperature": 25.5,
		"humidity":    60,
		"timestamp":   "2026-02-20T10:00:00Z",
	}

	sensorData2 := map[string]any{
		"id":          2,
		"device_id":   "sensor-001",
		"temperature": 25.6,
		"humidity":    59,
		"timestamp":   "2026-02-20T10:01:00Z",
	}

	sensorData3 := map[string]any{
		"id":          3,
		"device_id":   "sensor-002",
		"temperature": 24.8,
		"humidity":    65,
		"timestamp":   "2026-02-20T10:00:00Z",
	}

	sensorData4 := map[string]any{
		"id":          4,
		"device_id":   "sensor-002",
		"temperature": 24.9,
		"humidity":    64,
		"timestamp":   "2026-02-20T10:01:00Z",
	}

	sensorData5 := map[string]any{
		"id":          5,
		"device_id":   "sensor-003",
		"temperature": 26.0,
		"humidity":    58,
		"timestamp":   "2026-02-20T10:00:00Z",
	}

	// 6. 使用共享 batch 执行多个插入操作
	// 所有操作都使用同一个 batch，确保原子性
	// 这样可以减少网络开销，提高性能
	t.Log("使用共享 batch 执行多个插入操作...")

	// 插入第一条设备数据
	_, err = table.Insert(&sensorData1, batch)
	if err != nil {
		t.Fatalf("Failed to insert sensor-001 data: %v", err)
	}
	t.Log("Added sensor-001 data to batch")

	// 插入第二条设备数据（使用同一个 batch）
	_, err = table.Insert(&sensorData2, batch)
	if err != nil {
		t.Fatalf("Failed to insert sensor-001 next data: %v", err)
	}
	t.Log("Added sensor-001 next data to batch")

	// 插入第三条设备数据（使用同一个 batch）
	_, err = table.Insert(&sensorData3, batch)
	if err != nil {
		t.Fatalf("Failed to insert sensor-002 data: %v", err)
	}
	t.Log("Added sensor-002 data to batch")

	// 插入第四条设备数据（使用同一个 batch）
	_, err = table.Insert(&sensorData4, batch)
	if err != nil {
		t.Fatalf("Failed to insert sensor-002 next data: %v", err)
	}
	t.Log("Added sensor-002 next data to batch")

	// 插入第五条设备数据（使用同一个 batch）
	_, err = table.Insert(&sensorData5, batch)
	if err != nil {
		t.Fatalf("Failed to insert sensor-003 data: %v", err)
	}
	t.Log("Added sensor-003 data to batch")

	// 7. 检查 batch 中的操作数量
	t.Logf("Batch operation count: %d", batch.Len())

	// 8. 最后统一提交 batch
	// 在实际应用中，这一步可以在所有数据收集完成后执行
	// 例如，每 5 分钟或每收集 100 条数据后提交一次
	// 这样可以保证所有操作要么全部成功，要么全部失败
	// 实现了批量操作的原子性，同时避免了频繁提交的性能开销
	t.Log("提交 batch...")
	err = kvStore.WriteBatch(batch)
	if err != nil {
		t.Fatalf("Failed to commit batch: %v", err)
	}
	t.Log("Batch committed successfully")

	// 9. 验证所有记录都已成功插入
	t.Log("验证所有记录都已成功插入...")

	// 验证所有记录
	allSearchFields := map[string]any{"id": nil}
	allIter, err := table.Search(&allSearchFields)
	if err != nil {
		t.Fatalf("Failed to search all records: %v", err)
	}
	defer allIter.Release()

	allRecords := allIter.GetRecords(true)
	defer allRecords.Release()
	t.Logf("Found %d total records", len(allRecords))
	for i, r := range allRecords {
		t.Logf("  Record %d: %v", i+1, r)
	}

	// 验证记录数量
	if len(allRecords) != 5 {
		t.Errorf("Expected 5 records, got %d", len(allRecords))
	}

	// 验证每条记录的设备ID
	deviceCounts := make(map[string]int)
	for _, r := range allRecords {
		deviceID, ok := r["device_id"].(string)
		if ok {
			deviceCounts[deviceID]++
		}
	}

	// 验证每个设备的记录数量
	expectedCounts := map[string]int{
		"sensor-001": 2,
		"sensor-002": 2,
		"sensor-003": 1,
	}

	for deviceID, expectedCount := range expectedCounts {
		actualCount := deviceCounts[deviceID]
		t.Logf("Device %s: Expected %d records, got %d", deviceID, expectedCount, actualCount)
		if actualCount != expectedCount {
			t.Errorf("Device %s: Expected %d records, got %d", deviceID, expectedCount, actualCount)
		}
	}

	// 10. 演示批量共享的优势
	t.Log("\nBatch sharing example completed successfully!")
	t.Log("This example demonstrates how to:")
	t.Log("1. Create a shared batch for multiple operations")
	t.Log("2. Use the same batch for multiple insert/update/delete operations")
	t.Log("3. Commit all operations atomically")
	t.Log("4. Achieve eventual consistency in edge computing scenarios")
	t.Log("5. Optimize performance by reducing network overhead")

	// 11. 实际应用示例
	t.Log("\n实际应用示例:")
	t.Log(`// 创建共享 batch
kvStore := storage.GetDBManager().GetDB()
batch := kvStore.GetBatch()

// 收集数据过程中，多次使用同一个 batch
collectedData := []map[string]any{
	{"id": 1, "device_id": "sensor-001", "temperature": 25.5, "humidity": 60, "timestamp": "2026-02-20T10:00:00Z"},
	{"id": 2, "device_id": "sensor-001", "temperature": 25.6, "humidity": 59, "timestamp": "2026-02-20T10:01:00Z"},
	// 更多数据...
}

for _, data := range collectedData {
	// 多次调用 Insert，使用同一个 batch
	table.Insert(&data, batch)
}

// 所有数据收集完成后，统一提交
kvStore.WriteBatch(batch)
batch.Reset()
`)
}
