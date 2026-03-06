package storage

import (
	"bytes"
	"testing"
)

// TestSwitchToSnapshotAndDB 测试SwitchToSnapshot和SwitchToDB功能，验证读一致
func TestSwitchToSnapshotAndDB(t *testing.T) {
	// 创建临时目录用于测试
	tmpDir := t.TempDir()

	// 创建LevelDBStore实例
	store, err := dbManager.NewLevelDBStore(tmpDir, nil)
	if err != nil {
		t.Fatalf("Failed to create LevelDBStore: %v", err)
	}
	defer store.Close()

	// 转换为*LevelDBStore以便调用测试函数
	leveldbStore := store.(*LevelDBStore)

	// 插入初始数据
	key1 := []byte("key1")
	value1 := []byte("value1")
	key2 := []byte("key2")
	value2 := []byte("value2")

	if err := store.Put(key1, value1); err != nil {
		t.Fatalf("Failed to put key1: %v", err)
	}
	if err := store.Put(key2, value2); err != nil {
		t.Fatalf("Failed to put key2: %v", err)
	}

	// 切换到快照模式
	if err := leveldbStore.SwitchToSnapshot(); err != nil {
		t.Fatalf("Failed to switch to snapshot: %v", err)
	}

	// 验证当前是快照模式
	if !leveldbStore.isSnapshot {
		t.Error("Expected store to be in snapshot mode, but it's not")
	}

	// 在快照模式下读取数据
	snapshotValue1, err := store.Get(key1)
	if err != nil {
		t.Fatalf("Failed to get key1 in snapshot mode: %v", err)
	}
	if !bytes.Equal(snapshotValue1, value1) {
		t.Errorf("Expected value1 %s, got %s", value1, snapshotValue1)
	}

	// 切换回数据库模式
	if err := leveldbStore.SwitchToDB(); err != nil {
		t.Fatalf("Failed to switch to DB: %v", err)
	}

	// 修改数据
	newValue1 := []byte("new_value1")
	if err := store.Put(key1, newValue1); err != nil {
		t.Fatalf("Failed to update key1: %v", err)
	}

	// 切换到新的快照模式
	if err := leveldbStore.SwitchToSnapshot(); err != nil {
		t.Fatalf("Failed to switch to snapshot again: %v", err)
	}

	// 读取数据，应该看到最新值
	updatedSnapshotValue1, err := store.Get(key1)
	if err != nil {
		t.Fatalf("Failed to get key1 in updated snapshot: %v", err)
	}
	if !bytes.Equal(updatedSnapshotValue1, newValue1) {
		t.Errorf("Expected updated value1 %s, got %s", newValue1, updatedSnapshotValue1)
	}

	// 切换回数据库模式
	if err := leveldbStore.SwitchToDB(); err != nil {
		t.Fatalf("Failed to switch to DB again: %v", err)
	}

	// 验证当前不是快照模式
	if leveldbStore.isSnapshot {
		t.Error("Expected store to be in DB mode, but it's in snapshot mode")
	}
}

// TestSnapshotReadConsistency 测试快照读一致功能
func TestSnapshotReadConsistency(t *testing.T) {
	// 创建临时目录用于测试
	tmpDir := t.TempDir()

	// 创建LevelDBStore实例
	store, err := dbManager.NewLevelDBStore(tmpDir, nil)
	if err != nil {
		t.Fatalf("Failed to create LevelDBStore: %v", err)
	}
	defer store.Close()

	// 转换为*LevelDBStore以便调用测试函数
	leveldbStore := store.(*LevelDBStore)

	// 插入初始数据
	key := []byte("consistent_key")
	initialValue := []byte("initial_value")

	if err := store.Put(key, initialValue); err != nil {
		t.Fatalf("Failed to put initial key: %v", err)
	}

	// 切换到快照模式
	if err := leveldbStore.SwitchToSnapshot(); err != nil {
		t.Fatalf("Failed to switch to snapshot: %v", err)
	}

	// 在快照模式下读取初始值
	snapshotValue, err := store.Get(key)
	if err != nil {
		t.Fatalf("Failed to get key in snapshot: %v", err)
	}
	if !bytes.Equal(snapshotValue, initialValue) {
		t.Errorf("Expected initial value %s, got %s", initialValue, snapshotValue)
	}

	// 切换回数据库模式
	if err := leveldbStore.SwitchToDB(); err != nil {
		t.Fatalf("Failed to switch to DB: %v", err)
	}

	// 修改数据
	updatedValue := []byte("updated_value")
	if err := store.Put(key, updatedValue); err != nil {
		t.Fatalf("Failed to update key: %v", err)
	}

	// 切换到快照模式（创建新快照）
	if err := leveldbStore.SwitchToSnapshot(); err != nil {
		t.Fatalf("Failed to switch to new snapshot: %v", err)
	}

	// 在新快照中读取，应该看到更新后的值
	newSnapshotValue, err := store.Get(key)
	if err != nil {
		t.Fatalf("Failed to get key in new snapshot: %v", err)
	}
	if !bytes.Equal(newSnapshotValue, updatedValue) {
		t.Errorf("Expected updated value %s, got %s", updatedValue, newSnapshotValue)
	}

	// 切换回数据库模式
	if err := leveldbStore.SwitchToDB(); err != nil {
		t.Fatalf("Failed to switch to DB again: %v", err)
	}

	t.Logf("Snapshot read consistency test completed successfully")
}
