package metrics

import (
	"testing"
	"time"
)

func TestMetrics(t *testing.T) {
	// 创建指标收集器
	m := NewDefaultMetrics()

	// 测试事务指标
	t.Run("TransactionMetrics", func(t *testing.T) {
		// 记录事务
		m.RecordTransaction(100*time.Millisecond, true)
		m.RecordTransaction(200*time.Millisecond, false)

		// 验证事务计数
		if m.GetTransactionCount() != 2 {
			t.Errorf("Expected transaction count 2, got %d", m.GetTransactionCount())
		}

		// 验证事务成功率
		successRate := m.GetTransactionSuccessRate()
		if successRate != 0.5 {
			t.Errorf("Expected success rate 0.5, got %f", successRate)
		}

		// 验证平均事务持续时间
		averageDuration := m.GetAverageTransactionDuration()
		expectedDuration := 150 * time.Millisecond
		if averageDuration != expectedDuration {
			t.Errorf("Expected average duration %v, got %v", expectedDuration, averageDuration)
		}
	})

	// 测试查询指标
	t.Run("QueryMetrics", func(t *testing.T) {
		// 记录查询
		m.RecordQuery(50*time.Millisecond, "SELECT")
		m.RecordQuery(75*time.Millisecond, "INSERT")

		// 验证查询计数
		if m.GetQueryCount() != 2 {
			t.Errorf("Expected query count 2, got %d", m.GetQueryCount())
		}

		// 验证平均查询持续时间
		averageDuration := m.GetAverageQueryDuration()
		expectedDuration := 62 * time.Millisecond // (50 + 75) / 2 = 62.5, but time.Duration uses nanoseconds
		if averageDuration != expectedDuration {
			t.Errorf("Expected average duration %v, got %v", expectedDuration, averageDuration)
		}
	})

	// 测试存储操作指标
	t.Run("StorageMetrics", func(t *testing.T) {
		// 记录存储操作
		m.RecordStorageOperation("PUT", 10*time.Millisecond)
		m.RecordStorageOperation("GET", 5*time.Millisecond)

		// 验证存储操作计数
		if m.GetStorageOperationCount() != 2 {
			t.Errorf("Expected storage operation count 2, got %d", m.GetStorageOperationCount())
		}
	})

	// 测试系统状态指标
	t.Run("SystemMetrics", func(t *testing.T) {
		// 记录系统状态
		m.RecordSystemState(10, 500, 100)

		// 验证系统状态
		concurrency, memory, _ := m.GetCurrentSystemState()
		if concurrency != 10 {
			t.Errorf("Expected concurrency 10, got %d", concurrency)
		}
		if memory != 500 {
			t.Errorf("Expected memory usage 500, got %d", memory)
		}
		// 请求数会被时间窗口计数器处理，这里不做严格验证
	})

	// 测试重置功能
	t.Run("Reset", func(t *testing.T) {
		// 重置指标
		m.Reset()

		// 验证重置后的值
		if m.GetTransactionCount() != 0 {
			t.Errorf("Expected transaction count 0 after reset, got %d", m.GetTransactionCount())
		}
		if m.GetQueryCount() != 0 {
			t.Errorf("Expected query count 0 after reset, got %d", m.GetQueryCount())
		}
		if m.GetStorageOperationCount() != 0 {
			t.Errorf("Expected storage operation count 0 after reset, got %d", m.GetStorageOperationCount())
		}
	})
}

func TestTimeWindowCounter(t *testing.T) {
	// 创建时间窗口计数器
	tc := NewTimeWindowCounter(5) // 5秒窗口

	// 记录事件
	tc.Record()
	tc.Record()

	// 验证计数
	if tc.GetTotal() != 2 {
		t.Errorf("Expected count 2, got %d", tc.GetTotal())
	}

	// 等待窗口过期
	time.Sleep(6 * time.Second)

	// 验证计数已重置
	if tc.GetTotal() != 0 {
		t.Errorf("Expected count 0 after window expiry, got %d", tc.GetTotal())
	}
}
