package storage

import (
	"fmt"
	"testing"

	"github.com/syndtr/goleveldb/leveldb"
)

// TestBatchPool_Cleanliness 测试批处理对象池中的对象是否干净
func TestBatchPool_Cleanliness(t *testing.T) {
	// 测试场景1：从池中获取新对象，应该是干净的
	t.Run("GetCleanObject", func(t *testing.T) {
		batch := LdbBatchPool.Get()
		defer LdbBatchPool.Put(batch)

		// 验证批处理对象是否干净
		if batch.Len() != 0 {
			t.Errorf("Expected clean batch with 0 operations, got %d operations", batch.Len())
		}
	})

	// 测试场景2：向批处理对象添加操作后放回池中，再次获取应该是干净的
	t.Run("PutAndGetCleanObject", func(t *testing.T) {
		// 从池中获取批处理对象
		batch := LdbBatchPool.Get()

		// 向批处理对象添加一些操作
		batch.Put([]byte("key1"), []byte("value1"))
		batch.Put([]byte("key2"), []byte("value2"))
		batch.Delete([]byte("key3"))

		// 验证批处理对象现在有操作
		expectedOperations := 3
		if batch.Len() != expectedOperations {
			t.Errorf("Expected %d operations, got %d operations", expectedOperations, batch.Len())
		}

		// 将批处理对象放回池中
		LdbBatchPool.Put(batch)

		// 再次从池中获取批处理对象
		newBatch := LdbBatchPool.Get()
		defer LdbBatchPool.Put(newBatch)

		// 验证批处理对象是否被重置干净
		if newBatch.Len() != 0 {
			t.Errorf("Expected clean batch with 0 operations after reset, got %d operations", newBatch.Len())
		}
	})

	// 测试场景3：验证 Reset 方法的有效性
	t.Run("ResetEffectiveness", func(t *testing.T) {
		batch := LdbBatchPool.Get()
		defer LdbBatchPool.Put(batch)

		// 向批处理对象添加操作
		batch.Put([]byte("test_key"), []byte("test_value"))
		if batch.Len() != 1 {
			t.Errorf("Expected 1 operation, got %d operations", batch.Len())
		}

		// 手动调用 Reset
		batch.Reset()

		// 验证批处理对象是否干净
		if batch.Len() != 0 {
			t.Errorf("Expected clean batch with 0 operations after Reset, got %d operations", batch.Len())
		}
	})

	// 测试场景4：测试大型批处理对象的处理
	t.Run("LargeBatchCleanliness", func(t *testing.T) {
		batch := LdbBatchPool.Get()
		defer LdbBatchPool.Put(batch)

		// 向批处理对象添加大量操作
		for i := 0; i < 1000; i++ {
			key := []byte(fmt.Sprintf("large_key_%d", i))
			value := []byte(fmt.Sprintf("large_value_%d", i))
			batch.Put(key, value)
		}

		// 验证批处理对象有操作
		if batch.Len() != 1000 {
			t.Errorf("Expected 1000 operations, got %d operations", batch.Len())
		}

		// 将批处理对象放回池中
		LdbBatchPool.Put(batch)

		// 再次获取批处理对象
		newBatch := LdbBatchPool.Get()
		defer LdbBatchPool.Put(newBatch)

		// 验证批处理对象是否干净
		if newBatch.Len() != 0 {
			t.Errorf("Expected clean batch with 0 operations for large batch, got %d operations", newBatch.Len())
		}
	})
}

// TestBatchPool_NilSafety 测试批处理对象池对 nil 对象的处理
func TestBatchPool_NilSafety(t *testing.T) {
	// 测试场景：向池中放入 nil 对象，应该不会导致错误
	t.Run("PutNilObject", func(t *testing.T) {
		// 尝试向池中放入 nil 对象，应该不会 panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Putting nil batch should not panic, got: %v", r)
			}
		}()

		var nilBatch *leveldb.Batch
		LdbBatchPool.Put(nilBatch)
	})
}
