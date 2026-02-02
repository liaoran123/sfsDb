package engine

import (
	"fmt"
	"testing"

	"github.com/liaoran123/sfsDb/storage"
)

// TestBatchContainerMaxBatchSize 测试 batchContainer 的批量大小检查逻辑
func TestBatchContainerMaxBatchSize(t *testing.T) {
	// 初始化数据库
	_, err := storage.OpenDefaultDb("./batchcontainer_test_db1")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// 创建测试表
	table, err := TableNew("test_batchcontainer")
	if err != nil {
		t.Fatalf("TableNew failed: %v", err)
	}

	// 设置字段
	fields := map[string]any{"id": 0, "name": ""}
	table.SetFields(fields)

	// 创建主键
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("id")
	table.CreateIndex(pk)

	// 获取 kvStore
	kvStore := table.kvStore
	if kvStore == nil {
		t.Fatalf("Failed to get kvStore")
	}

	// 创建 batch
	batch := kvStore.GetBatch()
	if batch == nil {
		t.Fatalf("Failed to create batch")
	}

	// 创建 batchContainer，设置较小的 maxBatchSize 以便测试
	container := NewBatchContainer(batch, table.indexs, table.id, kvStore)
	// 设置 maxBatchSize 为 5，以便快速触发批量写入
	container.SetMaxBatchSize(5)

	// 创建测试数据
	for i := 1; i <= 15; i++ {
		// 创建字段字节映射
		fieldsBytes := map[string][]byte{
			"id":   []byte{byte(i)},
			"name": []byte{byte('a' + i - 1)},
		}

		// 执行操作
		container.Operation(&fieldsBytes)

		// 每执行 5 次操作后，检查 batch 长度是否被重置
		if i%5 == 0 {
			// 执行完第 5、10、15 次操作后，batch 应该被重置
			if container.Len() >= 5 {
				t.Errorf("After %d operations, batch length should be reset, but got %d", i, container.Len())
			}
			t.Logf("After %d operations, batch length: %d (should be reset)", i, container.Len())
		} else {
			// 其他情况下，batch 长度应该增加
			if container.Len() == 0 {
				t.Errorf("After %d operations, batch length should not be 0", i)
			}
			t.Logf("After %d operations, batch length: %d", i, container.Len())
		}
	}

	t.Logf("TestBatchContainerMaxBatchSize passed")
}

// TestBatchContainerDifferentMaxBatchSizes 测试不同的 maxBatchSize 值
func TestBatchContainerDifferentMaxBatchSizes(t *testing.T) {
	// 初始化数据库
	_, err := storage.OpenDefaultDb("./batchcontainer_test_db2")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// 创建测试表
	table, err := TableNew("test_batchcontainer_sizes")
	if err != nil {
		t.Fatalf("TableNew failed: %v", err)
	}

	// 设置字段
	fields := map[string]any{"id": 0, "name": ""}
	table.SetFields(fields)

	// 创建主键
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("id")
	table.CreateIndex(pk)

	// 获取 kvStore
	kvStore := table.kvStore
	if kvStore == nil {
		t.Fatalf("Failed to get kvStore")
	}

	// 测试不同的 maxBatchSize 值
	testSizes := []int{3, 7, 10}

	for _, size := range testSizes {
		t.Run(fmt.Sprintf("MaxBatchSize=%d", size), func(t *testing.T) {
			// 创建 batch
			batch := kvStore.GetBatch()
			if batch == nil {
				t.Fatalf("Failed to create batch")
			}

			// 创建 batchContainer
			container := NewBatchContainer(batch, table.indexs, table.id, kvStore)
			container.SetMaxBatchSize(size)

			// 执行操作
			writeCount := 0
			for i := 1; i <= 20; i++ {
				// 创建字段字节映射
				fieldsBytes := map[string][]byte{
					"id":   []byte{byte(i)},
					"name": []byte{byte('a' + i - 1)},
				}

				// 记录操作前的 batch 长度
				beforeLen := container.Len()

				// 执行操作
				container.Operation(&fieldsBytes)

				// 检查是否触发了写入
				afterLen := container.Len()
				// 每个Operation调用会添加1个操作到batch中（只有主键索引）
				// 所以如果操作前的batch长度加上1大于等于maxBatchSize，那么操作后batch长度应该小于操作前的batch长度
				if beforeLen+1 >= size && afterLen < beforeLen {
					writeCount++
					t.Logf("Write batch triggered at operation %d", i)
				}
			}

			// 验证写入次数
			// 计算期望的写入次数：每次写入处理size个操作
			expectedWrites := 20 / size

			if writeCount != expectedWrites {
				t.Errorf("Expected %d writes for maxBatchSize=%d, got %d", expectedWrites, size, writeCount)
			} else {
				t.Logf("Expected %d writes for maxBatchSize=%d, got %d (correct)", expectedWrites, size, writeCount)
			}
		})
	}
}

// TestBatchContainerNoMaxBatchSize 测试当 maxBatchSize <= 0 时的行为
func TestBatchContainerNoMaxBatchSize(t *testing.T) {
	// 初始化数据库
	_, err := storage.OpenDefaultDb("./batchcontainer_test_db3")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// 创建测试表
	table, err := TableNew("test_batchcontainer_no_limit")
	if err != nil {
		t.Fatalf("TableNew failed: %v", err)
	}

	// 设置字段
	fields := map[string]any{"id": 0, "name": ""}
	table.SetFields(fields)

	// 创建主键
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("id")
	table.CreateIndex(pk)

	// 获取 kvStore
	kvStore := table.kvStore
	if kvStore == nil {
		t.Fatalf("Failed to get kvStore")
	}

	// 创建 batch
	batch := kvStore.GetBatch()
	if batch == nil {
		t.Fatalf("Failed to create batch")
	}

	// 创建 batchContainer，设置 maxBatchSize <= 0
	container := NewBatchContainer(batch, table.indexs, table.id, kvStore)
	// 设置 maxBatchSize 为 0，表示无限制
	container.SetMaxBatchSize(0)

	// 执行多次操作
	for i := 1; i <= 20; i++ {
		// 创建字段字节映射
		fieldsBytes := map[string][]byte{
			"id":   []byte{byte(i)},
			"name": []byte{byte('a' + i - 1)},
		}

		// 执行操作
		container.Operation(&fieldsBytes)

		// 检查 batch 长度是否在增加（不应该被重置）
		currentLen := container.Len()
		if currentLen == 0 {
			t.Errorf("After %d operations, batch length should not be 0 when maxBatchSize <= 0", i)
		}
	}

	// 验证最终 batch 长度
	finalLen := container.Len()
	if finalLen == 0 {
		t.Errorf("Final batch length should not be 0 when maxBatchSize <= 0")
	} else {
		t.Logf("Final batch length: %d (correct, should not be reset)", finalLen)
	}

	t.Logf("TestBatchContainerNoMaxBatchSize passed")
}
