package storage

import (
	"fmt"
	"os"
	"testing"
)

// TestSnapshotConsistency 测试快照的读一致性
func TestSnapshotConsistency(t *testing.T) {
	// 创建临时数据库
	dbPath := "./test_snapshot_db"
	cleanup := func() {
		os.RemoveAll(dbPath)
	}
	cleanup()
	defer cleanup()

	// 打开数据库
	db, err := NewLevelDBStore(dbPath, nil)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// 1. 写入初始数据
	key1 := []byte("key1")
	value1 := []byte("value1")
	key2 := []byte("key2")
	value2 := []byte("value2")

	err = db.Put(key1, value1)
	if err != nil {
		t.Fatalf("Failed to put key1: %v", err)
	}

	err = db.Put(key2, value2)
	if err != nil {
		t.Fatalf("Failed to put key2: %v", err)
	}

	// 验证初始数据
	v1, err := db.Get(key1)
	if err != nil {
		t.Fatalf("Failed to get key1: %v", err)
	}
	if string(v1) != string(value1) {
		t.Fatalf("Expected key1 value: %s, got: %s", value1, v1)
	}

	v2, err := db.Get(key2)
	if err != nil {
		t.Fatalf("Failed to get key2: %v", err)
	}
	if string(v2) != string(value2) {
		t.Fatalf("Expected key2 value: %s, got: %s", value2, v2)
	}

	fmt.Println("Step 1: Initial data written successfully")
	fmt.Printf("key1: %s, key2: %s\n", v1, v2)

	// 2. 创建快照
	snapshot, err := db.Snapshot()
	if err != nil {
		t.Fatalf("Failed to create snapshot: %v", err)
	}
	defer snapshot.Release()

	fmt.Println("Step 2: Snapshot created successfully")

	// 3. 在创建快照后修改数据
	newValue1 := []byte("new_value1")
	newValue2 := []byte("new_value2")

	err = db.Put(key1, newValue1)
	if err != nil {
		t.Fatalf("Failed to update key1: %v", err)
	}

	err = db.Put(key2, newValue2)
	if err != nil {
		t.Fatalf("Failed to update key2: %v", err)
	}

	// 验证数据已修改
	updatedV1, err := db.Get(key1)
	if err != nil {
		t.Fatalf("Failed to get updated key1: %v", err)
	}
	if string(updatedV1) != string(newValue1) {
		t.Fatalf("Expected updated key1 value: %s, got: %s", newValue1, updatedV1)
	}

	updatedV2, err := db.Get(key2)
	if err != nil {
		t.Fatalf("Failed to get updated key2: %v", err)
	}
	if string(updatedV2) != string(newValue2) {
		t.Fatalf("Expected updated key2 value: %s, got: %s", newValue2, updatedV2)
	}

	fmt.Println("Step 3: Data updated successfully after snapshot creation")
	fmt.Printf("Updated key1: %s, updated key2: %s\n", updatedV1, updatedV2)

	// 4. 使用快照读取数据，验证读取到的是快照创建时的数据
	snapshotV1, err := snapshot.Get(key1)
	if err != nil {
		t.Fatalf("Failed to get key1 from snapshot: %v", err)
	}
	if string(snapshotV1) != string(value1) {
		t.Fatalf("Expected snapshot key1 value: %s, got: %s", value1, snapshotV1)
	}

	snapshotV2, err := snapshot.Get(key2)
	if err != nil {
		t.Fatalf("Failed to get key2 from snapshot: %v", err)
	}
	if string(snapshotV2) != string(value2) {
		t.Fatalf("Expected snapshot key2 value: %s, got: %s", value2, snapshotV2)
	}

	fmt.Println("Step 4: Snapshot read consistency verified")
	fmt.Printf("Snapshot key1: %s, snapshot key2: %s\n", snapshotV1, snapshotV2)

	// 5. 验证快照的迭代器功能
	iter := snapshot.Iterator(nil, nil)
	defer iter.Release()

	fmt.Println("Step 5: Testing snapshot iterator")
	for iter.Next() {
		key := iter.Key()
		val := iter.Value()
		fmt.Printf("Iterator - Key: %s, Value: %s\n", key, val)
		// 验证迭代器读取到的数据也是快照创建时的数据
		if string(key) == string(key1) && string(val) != string(value1) {
			t.Fatalf("Iterator - Expected key1 value: %s, got: %s", value1, val)
		}
		if string(key) == string(key2) && string(val) != string(value2) {
			t.Fatalf("Iterator - Expected key2 value: %s, got: %s", value2, val)
		}
	}

	// 6. 总结
	fmt.Println("\n=== Test Summary ===")
	fmt.Println("✓ Initial data written successfully")
	fmt.Println("✓ Snapshot created successfully")
	fmt.Println("✓ Data updated successfully after snapshot creation")
	fmt.Println("✓ Snapshot read consistency verified")
	fmt.Println("✓ Snapshot iterator works correctly")
	fmt.Println("\nConclusion: Snapshot provides consistent read view of the database at the time of creation")
	fmt.Println("Even when data is modified after snapshot creation, the snapshot still returns the original data")
}

// TestSnapshotFromSnapshot 测试从快照创建快照的错误处理
func TestSnapshotFromSnapshot(t *testing.T) {
	// 创建临时数据库
	dbPath := "./test_snapshot_from_snapshot_db"
	cleanup := func() {
		os.RemoveAll(dbPath)
	}
	cleanup()
	defer cleanup()

	// 打开数据库
	db, err := NewLevelDBStore(dbPath, nil)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// 写入一些数据
	err = db.Put([]byte("key"), []byte("value"))
	if err != nil {
		t.Fatalf("Failed to put key: %v", err)
	}

	// 创建第一个快照
	snapshot1, err := db.Snapshot()
	if err != nil {
		t.Fatalf("Failed to create first snapshot: %v", err)
	}
	defer snapshot1.Release()

	// 尝试从快照创建快照，应该失败
	snapshot2, err := snapshot1.(*LevelDBStore).Snapshot()
	if err == nil {
		t.Fatalf("Expected error when creating snapshot from snapshot, but got none")
	}
	if snapshot2 != nil {
		t.Fatalf("Expected nil snapshot when creating snapshot from snapshot, but got non-nil")
	}

	fmt.Printf("✓ Correctly rejected creating snapshot from snapshot: %v\n", err)
}
