package storage

import (
	"testing"
)

// TestSnapshotPoolClean 测试快照对象池的数据干净性
func TestSnapshotPoolClean(t *testing.T) {
	// 1. 从池中获取一个对象
	snapshot1 := LdbSnapshotPool.Get()
	if snapshot1 == nil {
		t.Fatal("Expected non-nil snapshot from pool")
	}

	// 2. 检查初始状态是否干净
	if snapshot1.ldb != nil {
		t.Error("Expected ldb to be nil for new snapshot")
	}
	if snapshot1.originalDB != nil {
		t.Error("Expected originalDB to be nil for new snapshot")
	}
	if snapshot1.isSnapshot != false {
		t.Error("Expected isSnapshot to be false for new snapshot")
	}
	if snapshot1.opts != nil {
		t.Error("Expected opts to be nil for new snapshot")
	}

	// 3. 修改对象状态
	snapshot1.isSnapshot = true

	// 4. 将对象放回池中
	LdbSnapshotPool.Put(snapshot1)

	// 5. 从池中再次获取一个对象
	snapshot2 := LdbSnapshotPool.Get()
	if snapshot2 == nil {
		t.Fatal("Expected non-nil snapshot from pool")
	}

	// 6. 检查状态是否被重置为干净状态
	if snapshot2.ldb != nil {
		t.Error("Expected ldb to be nil after pool reuse")
	}
	if snapshot2.originalDB != nil {
		t.Error("Expected originalDB to be nil after pool reuse")
	}
	if snapshot2.isSnapshot != false {
		t.Error("Expected isSnapshot to be false after pool reuse")
	}
	if snapshot2.opts != nil {
		t.Error("Expected opts to be nil after pool reuse")
	}

	// 7. 清理测试对象
	LdbSnapshotPool.Put(snapshot2)
}

// TestSnapshotPoolMultipleObjects 测试快照对象池的多个对象重用
func TestSnapshotPoolMultipleObjects(t *testing.T) {
	const poolSize = 5
	snapshots := make([]*LevelDBStore, poolSize)

	// 1. 从池中获取多个对象
	for i := 0; i < poolSize; i++ {
		snapshot := LdbSnapshotPool.Get()
		if snapshot == nil {
			t.Fatalf("Expected non-nil snapshot %d from pool", i)
		}

		// 检查初始状态
		if snapshot.ldb != nil || snapshot.originalDB != nil || snapshot.isSnapshot != false || snapshot.opts != nil {
			t.Errorf("Snapshot %d should be clean initially", i)
		}

		// 修改状态
		snapshot.isSnapshot = true

		snapshots[i] = snapshot
	}

	// 2. 将所有对象放回池中
	for i, snapshot := range snapshots {
		LdbSnapshotPool.Put(snapshot)
		snapshots[i] = nil // 清除引用
	}

	// 3. 再次从池中获取多个对象，检查状态
	for i := 0; i < poolSize; i++ {
		snapshot := LdbSnapshotPool.Get()
		if snapshot == nil {
			t.Fatalf("Expected non-nil snapshot %d from pool", i)
		}

		// 检查状态是否干净
		if snapshot.ldb != nil {
			t.Errorf("Snapshot %d ldb should be nil after pool reuse", i)
		}
		if snapshot.originalDB != nil {
			t.Errorf("Snapshot %d originalDB should be nil after pool reuse", i)
		}
		if snapshot.isSnapshot != false {
			t.Errorf("Snapshot %d isSnapshot should be false after pool reuse", i)
		}
		if snapshot.opts != nil {
			t.Errorf("Snapshot %d opts should be nil after pool reuse", i)
		}

		// 清理
		LdbSnapshotPool.Put(snapshot)
	}
}

// TestSnapshotPoolWithTransactions 测试快照对象池在事务中的使用
func TestSnapshotPoolWithTransactions(t *testing.T) {
	// 这个测试主要验证快照对象池在实际事务场景中的表现
	// 由于事务测试需要完整的表结构，这里只做基本验证

	// 1. 模拟事务中获取和使用快照
	for i := 0; i < 10; i++ {
		// 获取快照对象
		snapshot := LdbSnapshotPool.Get()
		if snapshot == nil {
			t.Fatalf("Expected non-nil snapshot %d from pool", i)
		}

		// 检查状态
		if snapshot.ldb != nil || snapshot.originalDB != nil || snapshot.isSnapshot != false || snapshot.opts != nil {
			t.Errorf("Snapshot %d should be clean for transaction use", i)
		}

		// 模拟事务操作
		// 这里只是简单模拟，实际事务会有更复杂的操作
		snapshot.isSnapshot = true

		// 完成后放回池
		LdbSnapshotPool.Put(snapshot)
	}
}
