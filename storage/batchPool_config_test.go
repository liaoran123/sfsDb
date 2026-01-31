package storage

import (
	"testing"

	"github.com/syndtr/goleveldb/leveldb"
)

// TestBatchPoolCustomSize 测试自定义批处理对象池大小
func TestBatchPoolCustomSize(t *testing.T) {
	// 保存原始值，以便测试后恢复
	originalMaxBatchPoolSize := MaxBatchPoolSize
	originalMaxBatchSize := MaxBatchSize

	// 测试结束后恢复原始值
	defer func() {
		MaxBatchPoolSize = originalMaxBatchPoolSize
		MaxBatchSize = originalMaxBatchSize
	}()

	// 测试场景1：设置自定义的最大批处理对象池大小
	t.Run("SetCustomMaxBatchPoolSize", func(t *testing.T) {
		// 设置自定义大小
		customSize := 500
		SetMaxBatchPoolSize(customSize)

		// 验证设置是否成功
		if GetMaxBatchPoolSize() != customSize {
			t.Errorf("Expected MaxBatchPoolSize to be %d, got %d", customSize, GetMaxBatchPoolSize())
		}

		// 测试边界情况：设置负值
		SetMaxBatchPoolSize(-100)
		if GetMaxBatchPoolSize() != customSize {
			t.Errorf("Expected MaxBatchPoolSize to remain %d, got %d", customSize, GetMaxBatchPoolSize())
		}

		// 测试边界情况：设置零值
		SetMaxBatchPoolSize(0)
		if GetMaxBatchPoolSize() != customSize {
			t.Errorf("Expected MaxBatchPoolSize to remain %d, got %d", customSize, GetMaxBatchPoolSize())
		}
	})

	// 测试场景2：设置自定义的批处理对象大小阈值
	t.Run("SetCustomMaxBatchSize", func(t *testing.T) {
		// 设置自定义大小阈值
		customSize := 512 * 1024 // 512KB
		SetMaxBatchSize(customSize)

		// 验证设置是否成功
		if GetMaxBatchSize() != customSize {
			t.Errorf("Expected MaxBatchSize to be %d, got %d", customSize, GetMaxBatchSize())
		}

		// 测试边界情况：设置负值
		SetMaxBatchSize(-1024)
		if GetMaxBatchSize() != customSize {
			t.Errorf("Expected MaxBatchSize to remain %d, got %d", customSize, GetMaxBatchSize())
		}

		// 测试边界情况：设置零值
		SetMaxBatchSize(0)
		if GetMaxBatchSize() != customSize {
			t.Errorf("Expected MaxBatchSize to remain %d, got %d", customSize, GetMaxBatchSize())
		}
	})

	// 测试场景3：验证自定义大小是否生效
	t.Run("CustomSizeEffectiveness", func(t *testing.T) {
		// 设置较小的自定义大小
		smallSize := 10
		SetMaxBatchPoolSize(smallSize)

		// 获取多个批处理对象，超过自定义大小
		var batches []interface{}
		for i := 0; i < smallSize*2; i++ {
			batch := LdbBatchPool.Get()
			batches = append(batches, batch)
		}

		// 将所有批处理对象放回池中
		for _, batch := range batches {
			LdbBatchPool.Put(batch.(*leveldb.Batch))
		}

		// 验证池大小是否在限制范围内
		currentSize := GetLdbBatchPoolSize()
		if currentSize > uint64(smallSize) {
			t.Errorf("Expected pool size to be <= %d, got %d", smallSize, currentSize)
		}
	})
}
