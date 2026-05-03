package monitor

import (
	"testing"
	"time"
)

// TestTransactionStatsMapInitialization 测试 GTransactionStatsMap 是否被正确初始化
func TestTransactionStatsMapInitialization(t *testing.T) {
	// 检查 GTransactionStatsMap 是否为 nil
	if GTransactionStatsMap == nil {
		t.Fatalf("GTransactionStatsMap 未被正确初始化")
	}

	// 由于 GTransactionStatsMap 已经被声明为 *TransactionStatsMap 类型
	// 这里不需要再进行类型断言，只需要检查它是否为 nil 即可

	t.Logf("GTransactionStatsMap 初始化测试通过")
}

// TestTransactionStatsMapSetTime 测试 SetTime 方法是否正确记录事务统计信息
func TestTransactionStatsMapSetTime(t *testing.T) {
	// 创建一个新的 TransactionStatsMap 实例用于测试
	testMap := NewTransactionStatsMap()

	// 测试数据
	txID := uint64(time.Now().UnixNano())
	duration := 100 * time.Millisecond
	tableName := "test_table"
	isolationLevel := "REPEATABLE_READ"
	isCommitted := true

	// 调用 SetTime 方法
	testMap.SetTime(txID, duration, tableName, isolationLevel, isCommitted)

	// 验证统计信息是否被正确记录
	value, ok := testMap.Data.Load(txID)
	if !ok {
		t.Fatalf("事务统计信息未被记录")
	}

	// 检查统计信息是否正确
	stats, ok := value.(*TransactionStats)
	if !ok {
		t.Fatalf("记录的不是 *TransactionStats 类型")
	}

	if stats.TxID != txID {
		t.Errorf("TxID 不匹配，期望: %d, 实际: %d", txID, stats.TxID)
	}

	if stats.TableName != tableName {
		t.Errorf("TableName 不匹配，期望: %s, 实际: %s", tableName, stats.TableName)
	}

	if stats.IsolationLevel != isolationLevel {
		t.Errorf("IsolationLevel 不匹配，期望: %s, 实际: %s", isolationLevel, stats.IsolationLevel)
	}

	if stats.IsCommitted != isCommitted {
		t.Errorf("IsCommitted 不匹配，期望: %v, 实际: %v", isCommitted, stats.IsCommitted)
	}

	if stats.GetTotalCount() != 1 {
		t.Errorf("TotalCount 不匹配，期望: 1, 实际: %d", stats.GetTotalCount())
	}

	if stats.GetDuration() != duration {
		t.Errorf("Duration 不匹配，期望: %v, 实际: %v", duration, stats.GetDuration())
	}

	if stats.GetAvgDuration() != duration {
		t.Errorf("AvgDuration 不匹配，期望: %v, 实际: %v", duration, stats.GetAvgDuration())
	}

	if stats.GetMaxDuration() != duration {
		t.Errorf("MaxDuration 不匹配，期望: %v, 实际: %v", duration, stats.GetMaxDuration())
	}

	if stats.GetMinDuration() != duration {
		t.Errorf("MinDuration 不匹配，期望: %v, 实际: %v", duration, stats.GetMinDuration())
	}

	t.Logf("SetTime 方法测试通过")
}

// TestTransactionStatsMapGetAll 测试 GetAll 方法是否正确返回所有统计信息
func TestTransactionStatsMapGetAll(t *testing.T) {
	// 创建一个新的 TransactionStatsMap 实例用于测试
	testMap := NewTransactionStatsMap()

	// 添加测试数据
	txID1 := uint64(time.Now().UnixNano())
	txID2 := txID1 + 1
	duration := 100 * time.Millisecond
	tableName := "test_table"
	isolationLevel := "REPEATABLE_READ"
	isCommitted := true

	testMap.SetTime(txID1, duration, tableName, isolationLevel, isCommitted)
	testMap.SetTime(txID2, duration, tableName, isolationLevel, isCommitted)

	// 调用 GetAll 方法
	allStats := testMap.GetAll()

	// 验证返回值是否正确
	if len(allStats) != 2 {
		t.Errorf("返回的统计信息数量不匹配，期望: 2, 实际: %d", len(allStats))
	}

	// 验证每个统计信息是否正确
	for txID, stats := range allStats {
		if stats == nil {
			t.Errorf("事务 %d 的统计信息为 nil", txID)
			continue
		}

		if stats.TxID != txID {
			t.Errorf("事务 %d 的 TxID 不匹配，期望: %d, 实际: %d", txID, txID, stats.TxID)
		}

		if stats.GetTotalCount() != 1 {
			t.Errorf("事务 %d 的 TotalCount 不匹配，期望: 1, 实际: %d", txID, stats.GetTotalCount())
		}
	}

	t.Logf("GetAll 方法测试通过")
}

// TestTransactionStatsAddDuration 测试 AddDuration 方法是否正确更新统计信息
func TestTransactionStatsAddDuration(t *testing.T) {
	// 创建一个新的 TransactionStats 实例
	txID := uint64(time.Now().UnixNano())
	tableName := "test_table"
	isolationLevel := "REPEATABLE_READ"
	stats := NewTransactionStats(txID, tableName, isolationLevel)

	// 添加测试数据
	duration1 := 100 * time.Millisecond
	duration2 := 200 * time.Millisecond
	duration3 := 50 * time.Millisecond
	totalDuration := duration1 + duration2 + duration3
	avgDuration := totalDuration / 3

	stats.AddDuration(duration1)
	stats.AddDuration(duration2)
	stats.AddDuration(duration3)

	// 验证统计信息是否正确
	if stats.GetTotalCount() != 3 {
		t.Errorf("TotalCount 不匹配，期望: 3, 实际: %d", stats.GetTotalCount())
	}

	if stats.GetDuration() != totalDuration {
		t.Errorf("Duration 不匹配，期望: %v, 实际: %v", totalDuration, stats.GetDuration())
	}

	if stats.GetAvgDuration() != avgDuration {
		t.Errorf("AvgDuration 不匹配，期望: %v, 实际: %v", avgDuration, stats.GetAvgDuration())
	}

	if stats.GetMaxDuration() != duration2 {
		t.Errorf("MaxDuration 不匹配，期望: %v, 实际: %v", duration2, stats.GetMaxDuration())
	}

	if stats.GetMinDuration() != duration3 {
		t.Errorf("MinDuration 不匹配，期望: %v, 实际: %v", duration3, stats.GetMinDuration())
	}

	t.Logf("AddDuration 方法测试通过")
}

// TestTransactionStatsEdgeCases 测试边界情况
func TestTransactionStatsEdgeCases(t *testing.T) {
	// 创建一个新的 TransactionStats 实例
	txID := uint64(time.Now().UnixNano())
	tableName := "test_table"
	isolationLevel := "REPEATABLE_READ"
	stats := NewTransactionStats(txID, tableName, isolationLevel)

	// 测试 Getter 方法在没有添加任何时长时的行为
	if stats.GetTotalCount() != 0 {
		t.Errorf("TotalCount 不匹配，期望: 0, 实际: %d", stats.GetTotalCount())
	}

	if stats.GetDuration() != 0 {
		t.Errorf("Duration 不匹配，期望: 0, 实际: %v", stats.GetDuration())
	}

	if stats.GetAvgDuration() != 0 {
		t.Errorf("AvgDuration 不匹配，期望: 0, 实际: %v", stats.GetAvgDuration())
	}

	if stats.GetMaxDuration() != 0 {
		t.Errorf("MaxDuration 不匹配，期望: 0, 实际: %v", stats.GetMaxDuration())
	}

	if stats.GetMinDuration() != 0 {
		t.Errorf("MinDuration 不匹配，期望: 0, 实际: %v", stats.GetMinDuration())
	}

	t.Logf("边界情况测试通过")
}
