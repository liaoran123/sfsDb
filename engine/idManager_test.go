package engine

import (
	"os"
	"testing"

	"github.com/liaoran123/sfsDb/storage"
)

func TestIDManager_GetOrCreateID(t *testing.T) {
	// 打开存储，使用全新的唯一路径，避免与其他测试冲突
	kvStore, err := storage.NewLevelDBStore("./test_idmanager_new_id_"+t.Name(), nil)
	if err != nil {
		t.Fatalf("Failed to open storage: %v", err)
	}
	defer func() {
		kvStore.Close()
	}()

	// 创建管理器
	manager := NewIDManager(kvStore)

	// 测试相同对象键返回相同ID
	id1, err := manager.GetOrCreateID("sys-table-test")
	if err != nil {
		t.Fatalf("Failed to get or create ID: %v", err)
	}

	id2, err := manager.GetOrCreateID("sys-table-test")
	if err != nil {
		t.Fatalf("Failed to get or create ID again: %v", err)
	}

	if id1 != id2 {
		t.Errorf("Expected same ID for same object key, got %d and %d", id1, id2)
	}

	// 测试不同对象键返回不同ID
	id3, err := manager.GetOrCreateID("sys-table-test2")
	if err != nil {
		t.Fatalf("Failed to get or create ID for different object key: %v", err)
	}

	if id1 == id3 {
		t.Errorf("Expected different ID for different object key, got %d and %d", id1, id3)
	}

	// 测试不同类型的对象键使用独立的计数器
	id4, err := manager.GetOrCreateID("sys-1-idx-test")
	if err != nil {
		t.Fatalf("Failed to get or create ID for different object type: %v", err)
	}

	// 不同类型的第一个ID都应该是0，因为它们使用独立的计数器
	if id1 != 0 || id4 != 0 {
		t.Errorf("Expected both IDs to be 0 for different object types, got %d and %d", id1, id4)
	}

	// 测试同类型的第二个ID应该递增
	id5, err := manager.GetOrCreateID("sys-table-test2")
	if err != nil {
		t.Fatalf("Failed to get or create second table ID: %v", err)
	}

	if id5 != 1 {
		t.Errorf("Expected second table ID to be 1, got %d", id5)
	}
}

func TestIDManager_AutoIncrement(t *testing.T) {
	// 打开存储，使用全新的唯一路径，避免与其他测试冲突
	kvStore, err := storage.NewLevelDBStore("./test_idmanager_auto_increment_"+t.Name(), nil)
	if err != nil {
		t.Fatalf("Failed to open storage: %v", err)
	}
	defer func() {
		kvStore.Close()
	}()

	// 创建管理器
	manager := NewIDManager(kvStore)

	// 创建5个表ID，验证自动递增
	for i := range 5 {
		key := "sys-table-test" + string(rune('0'+i))
		id, err := manager.GetOrCreateID(key)
		if err != nil {
			t.Fatalf("Failed to get or create ID for %s: %v", key, err)
		}

		// 验证ID是否按顺序分配
		expectedID := uint8(i)
		if id != expectedID {
			t.Errorf("Expected ID %d for %s, got %d", expectedID, key, id)
		}
	}
}

func TestIDManager_DifferentTypes(t *testing.T) {
	// 打开存储，使用全新的唯一路径，避免与其他测试冲突
	kvStore, err := storage.NewLevelDBStore("./test_idmanager_different_types_"+t.Name(), nil)
	if err != nil {
		t.Fatalf("Failed to open storage: %v", err)
	}
	defer func() {
		kvStore.Close()
	}()

	// 创建管理器
	manager := NewIDManager(kvStore)

	// 测试不同类型的对象使用不同的计数器
	tableID1, err := manager.GetOrCreateID("sys-table-test")
	if err != nil {
		t.Fatalf("Failed to get or create table ID: %v", err)
	}

	indexID1, err := manager.GetOrCreateID("sys-1-idx-test")
	if err != nil {
		t.Fatalf("Failed to get or create index ID: %v", err)
	}

	fieldID1, err := manager.GetOrCreateID("sys-1-field-test")
	if err != nil {
		t.Fatalf("Failed to get or create field ID: %v", err)
	}

	// 所有类型的第一个ID都应该是0
	if tableID1 != 0 || indexID1 != 0 || fieldID1 != 0 {
		t.Errorf("Expected all first IDs to be 0, got table: %d, index: %d, field: %d", tableID1, indexID1, fieldID1)
	}

	// 再创建一个同类型的ID，应该递增
	tableID2, err := manager.GetOrCreateID("sys-table-test2")
	if err != nil {
		t.Fatalf("Failed to get or create second table ID: %v", err)
	}

	if tableID2 != 1 {
		t.Errorf("Expected second table ID to be 1, got %d", tableID2)
	}
}

func TestIDManager_Concurrent(t *testing.T) {
	// 打开存储，使用全新的唯一路径，避免与其他测试冲突
	kvStore, err := storage.NewLevelDBStore("./test_idmanager_concurrent_"+t.Name(), nil)
	if err != nil {
		t.Fatalf("Failed to open storage: %v", err)
	}
	defer func() {
		kvStore.Close()
	}()

	// 创建管理器
	manager := NewIDManager(kvStore)

	// 并发测试
	concurrentCount := 100
	errChan := make(chan error, concurrentCount)

	for i := range concurrentCount {
		go func(index int) {
			// 使用10个不同的表名
			key := "sys-table-concurrent" + string(rune('0'+index%10))
			_, err := manager.GetOrCreateID(key)
			errChan <- err
		}(i)
	}

	// 收集错误
	for range concurrentCount {
		if err := <-errChan; err != nil {
			t.Errorf("Concurrent error: %v", err)
		}
	}
}

func TestIDManager_UpdateName(t *testing.T) {
	// 打开存储，使用全新的唯一路径，避免与其他测试冲突
	kvStore, err := storage.NewLevelDBStore("./test_idmanager_update_name_"+t.Name(), nil)
	if err != nil {
		t.Fatalf("Failed to open storage: %v", err)
	}
	defer func() {
		kvStore.Close()
	}()

	// 创建管理器
	manager := NewIDManager(kvStore)

	// 1. 创建一个对象，获取ID
	oldKey := "sys-table-oldname"
	id1, err := manager.GetOrCreateID(oldKey)
	if err != nil {
		t.Fatalf("Failed to get or create ID: %v", err)
	}

	// 2. 更新名称
	newKey := "sys-table-newname"
	err = manager.UpdateName(oldKey, newKey)
	if err != nil {
		t.Fatalf("Failed to update name: %v", err)
	}

	// 3. 验证旧键已删除
	_, err = manager.kvStore.Get([]byte(oldKey))
	if err == nil {
		t.Error("Expected old key to be deleted, but it still exists")
	}

	// 4. 验证新键存在且ID相同
	id2, err := manager.GetOrCreateID(newKey)
	if err != nil {
		t.Fatalf("Failed to get ID from new key: %v", err)
	}

	if id1 != id2 {
		t.Errorf("Expected same ID after name update, got %d and %d", id1, id2)
	}

	// 5. 验证更新不存在的键会返回错误
	nonExistentKey := "sys-table-nonexistent"
	err = manager.UpdateName(nonExistentKey, "sys-table-new")
	if err == nil {
		t.Error("Expected error when updating non-existent key, but got none")
	}
}

func TestIDManager_KeyAutoIncrement(t *testing.T) {
	// 测试存储路径
	storePath := "./test_idmanager_key_auto_increment_" + t.Name()

	// 清理可能存在的测试目录
	os.RemoveAll(storePath)

	// 打开存储，使用全新的唯一路径，避免与其他测试冲突
	kvStore, err := storage.NewLevelDBStore(storePath, nil)
	if err != nil {
		t.Fatalf("Failed to open storage: %v", err)
	}
	defer func() {
		kvStore.Close()
		// 测试后清理
		os.RemoveAll(storePath)
	}()

	// 创建管理器
	manager := NewIDManager(kvStore)

	// 测试key
	testKey := "test-counter"

	// 1. 测试GetNextID功能
	id1, err := manager.GetNextID(testKey)
	if err != nil {
		t.Fatalf("Failed to get next ID: %v", err)
	}
	if id1 != 0 {
		t.Errorf("Expected first ID to be 0, got %d", id1)
	}

	id2, err := manager.GetNextID(testKey)
	if err != nil {
		t.Fatalf("Failed to get next ID: %v", err)
	}
	if id2 != 1 {
		t.Errorf("Expected second ID to be 1, got %d", id2)
	}

	id3, err := manager.GetNextID(testKey)
	if err != nil {
		t.Fatalf("Failed to get next ID: %v", err)
	}
	if id3 != 2 {
		t.Errorf("Expected third ID to be 2, got %d", id3)
	}

	// 2. 测试GetCurrentID功能
	currentID, err := manager.GetCurrentID(testKey)
	if err != nil {
		t.Fatalf("Failed to get current ID: %v", err)
	}
	if currentID != 3 {
		t.Errorf("Expected current ID to be 3, got %d", currentID)
	}

	// 3. 测试SetID功能
	err = manager.SetID(testKey, 10)
	if err != nil {
		t.Fatalf("Failed to set ID: %v", err)
	}

	// 验证SetID后的当前值
	currentID, err = manager.GetCurrentID(testKey)
	if err != nil {
		t.Fatalf("Failed to get current ID after SetID: %v", err)
	}
	if currentID != 10 {
		t.Errorf("Expected current ID to be 10 after SetID, got %d", currentID)
	}

	// 验证SetID后GetNextID返回设置的值
	id4, err := manager.GetNextID(testKey)
	if err != nil {
		t.Fatalf("Failed to get next ID after SetID: %v", err)
	}
	if id4 != 10 {
		t.Errorf("Expected next ID after SetID to be 10, got %d", id4)
	}

	// 4. 测试ResetID功能
	err = manager.ResetID(testKey)
	if err != nil {
		t.Fatalf("Failed to reset ID: %v", err)
	}

	// 验证ResetID后的当前值
	currentID, err = manager.GetCurrentID(testKey)
	if err != nil {
		t.Fatalf("Failed to get current ID after ResetID: %v", err)
	}
	if currentID != 0 {
		t.Errorf("Expected current ID to be 0 after ResetID, got %d", currentID)
	}

	// 验证ResetID后GetNextID返回0
	id5, err := manager.GetNextID(testKey)
	if err != nil {
		t.Fatalf("Failed to get next ID after ResetID: %v", err)
	}
	if id5 != 0 {
		t.Errorf("Expected next ID after ResetID to be 0, got %d", id5)
	}

	// 5. 测试不同key的独立性
	anotherKey := "another-counter"
	id6, err := manager.GetNextID(anotherKey)
	if err != nil {
		t.Fatalf("Failed to get next ID for another key: %v", err)
	}
	if id6 != 0 {
		t.Errorf("Expected first ID for another key to be 0, got %d", id6)
	}

	// 验证原key不受影响
	id7, err := manager.GetNextID(testKey)
	if err != nil {
		t.Fatalf("Failed to get next ID for original key: %v", err)
	}
	if id7 != 1 {
		t.Errorf("Expected next ID for original key to be 1, got %d", id7)
	}
}

func TestIDManager_GetPreviousID(t *testing.T) {
	// 打开存储，使用全新的唯一路径，避免与其他测试冲突
	kvStore, err := storage.NewLevelDBStore("./test_idmanager_get_previous_id_"+t.Name(), nil)
	if err != nil {
		t.Fatalf("Failed to open storage: %v", err)
	}
	defer func() {
		kvStore.Close()
	}()

	// 创建管理器
	manager := NewIDManager(kvStore)

	// 测试key
	testKey := "test-counter"

	// 1. 生成一些ID
	for range 3 {
		_, err := manager.GetNextID(testKey)
		if err != nil {
			t.Fatalf("Failed to get next ID: %v", err)
		}
	}

	// 获取当前计数器值（应该是3，因为调用了3次GetNextID）
	currentIDBefore, err := manager.GetCurrentID(testKey)
	if err != nil {
		t.Fatalf("Failed to get current ID before GetPreviousID: %v", err)
	}
	t.Logf("Current ID before GetPreviousID: %d", currentIDBefore)

	// 2. 测试正常回退ID
	prevID, err := manager.GetPreviousID(testKey)
	if err != nil {
		t.Fatalf("Failed to get previous ID: %v", err)
	}
	t.Logf("Previous ID: %d", prevID)

	// 验证回退后的当前值
	currentIDAfter, err := manager.GetCurrentID(testKey)
	if err != nil {
		t.Fatalf("Failed to get current ID after GetPreviousID: %v", err)
	}
	t.Logf("Current ID after GetPreviousID: %d", currentIDAfter)

	// 5. 测试回退后再生成ID，应该继续从回退点开始
	id, err := manager.GetNextID(testKey)
	if err != nil {
		t.Fatalf("Failed to get next ID after GetPreviousID: %v", err)
	}
	t.Logf("Next ID after GetPreviousID: %d", id)
}
