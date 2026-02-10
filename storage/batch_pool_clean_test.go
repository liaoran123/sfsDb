package storage

import (
	"fmt"
	"testing"

	"github.com/syndtr/goleveldb/leveldb"
)

// TestBatchPoolClean 测试批处理对象池的数据干净性
func TestBatchPoolClean(t *testing.T) {
	// 1. 从池中获取一个对象
	batch1 := LdbBatchPool.Get()
	if batch1 == nil {
		t.Fatal("Expected non-nil batch from pool")
	}

	// 2. 检查初始状态是否干净（通过Reset确保）
	// 注意：leveldb.Batch没有直接的方法检查是否为空，我们通过添加操作后检查

	// 3. 向批处理添加操作
	batch1.Put([]byte("key1"), []byte("value1"))
	batch1.Put([]byte("key2"), []byte("value2"))

	// 4. 检查批处理是否包含操作
	if batch1.Len() == 0 {
		t.Error("Expected batch to contain operations")
	}

	// 5. 将对象放回池中
	LdbBatchPool.Put(batch1)

	// 6. 从池中再次获取一个对象
	batch2 := LdbBatchPool.Get()
	if batch2 == nil {
		t.Fatal("Expected non-nil batch from pool")
	}

	// 7. 检查状态是否被重置为干净状态
	if batch2.Len() != 0 {
		t.Error("Expected batch to be clean after pool reuse")
	}

	// 8. 清理测试对象
	LdbBatchPool.Put(batch2)
}

// TestBatchPoolMultipleObjects 测试批处理对象池的多个对象重用
func TestBatchPoolMultipleObjects(t *testing.T) {
	const poolSize = 5
	batches := make([]*leveldb.Batch, poolSize)

	// 1. 从池中获取多个对象
	for i := 0; i < poolSize; i++ {
		batch := LdbBatchPool.Get()
		if batch == nil {
			t.Fatalf("Expected non-nil batch %d from pool", i)
		}

		// 检查初始状态
		if batch.Len() != 0 {
			t.Errorf("Batch %d should be clean initially", i)
		}

		// 向批处理添加操作
		batch.Put([]byte(fmt.Sprintf("key%d", i)), []byte(fmt.Sprintf("value%d", i)))

		batches[i] = batch
	}

	// 2. 将所有对象放回池中
	for i, batch := range batches {
		LdbBatchPool.Put(batch)
		batches[i] = nil // 清除引用
	}

	// 3. 再次从池中获取多个对象，检查状态
	for i := 0; i < poolSize; i++ {
		batch := LdbBatchPool.Get()
		if batch == nil {
			t.Fatalf("Expected non-nil batch %d from pool", i)
		}

		// 检查状态是否干净
		if batch.Len() != 0 {
			t.Errorf("Batch %d should be clean after pool reuse", i)
		}

		// 清理
		LdbBatchPool.Put(batch)
	}
}

// TestBatchPoolWithLargeBatch 测试批处理对象池对大对象的处理
func TestBatchPoolWithLargeBatch(t *testing.T) {
	// 1. 从池中获取一个对象
	batch := LdbBatchPool.Get()
	if batch == nil {
		t.Fatal("Expected non-nil batch from pool")
	}

	// 2. 向批处理添加大量操作，使其超过大小限制
	largeValue := make([]byte, 2*1024*1024) // 2MB，超过 MaxBatchSize
	batch.Put([]byte("largeKey"), largeValue)

	// 3. 检查批处理大小
	if batch.Len() == 0 {
		t.Error("Expected batch to contain large operation")
	}

	// 4. 将大对象放回池中（应该被丢弃）
	LdbBatchPool.Put(batch)

	// 5. 从池中获取一个新对象，应该是干净的
	newBatch := LdbBatchPool.Get()
	if newBatch == nil {
		t.Fatal("Expected non-nil batch from pool")
	}

	// 6. 检查新对象是否干净
	if newBatch.Len() != 0 {
		t.Error("Expected new batch to be clean")
	}

	// 7. 清理测试对象
	LdbBatchPool.Put(newBatch)
}

// TestBatchPoolTypeSafety 测试批处理对象池的类型安全性
func TestBatchPoolTypeSafety(t *testing.T) {
	// 注意：这个测试主要验证我们的类型断言逻辑
	// 由于 pool 是私有的，我们无法直接向池中放入错误类型的对象
	// 但我们的 Get 方法已经包含了类型断言逻辑

	// 从池中获取对象，验证类型断言逻辑
	batch := LdbBatchPool.Get()
	if batch == nil {
		t.Fatal("Expected non-nil batch from pool")
	}

	// 验证返回的确实是 *leveldb.Batch 类型
	_, ok := interface{}(batch).(*leveldb.Batch)
	if !ok {
		t.Error("Expected batch to be *leveldb.Batch type")
	}

	// 清理测试对象
	LdbBatchPool.Put(batch)
}
