package transactionANT

import (
	"testing"

	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/storage"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

// TestSavepoint 测试保存点功能
func TestSavepoint(t *testing.T) {
	// 创建一个内存存储实例用于测试
	db, err := storage.NewLevelDBStore("./testdb", &opt.Options{})
	if err != nil {
		t.Fatalf("Failed to create LevelDB store: %v", err)
	}
	defer db.Close()

	// 创建一个表
	table, err := engine.TableNew("test_table")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 创建事务管理器
	batch := db.GetBatch()
	txMgr := NewTransactionManager(batch, "")

	// 添加表到事务管理器
	tableTx, err := txMgr.AddTable(table)
	if err != nil {
		t.Fatalf("Failed to add table to transaction manager: %v", err)
	}

	// 测试场景1: 创建保存点并回滚
	t.Run("CreateSavepointAndRollback", func(t *testing.T) {
		// 执行一些操作
		fields1 := map[string]any{"id": 1, "name": "test1"}
		_, err := tableTx.Insert(&fields1)
		if err != nil {
			t.Fatalf("Failed to insert record: %v", err)
		}

		// 创建保存点
		err = txMgr.CreateSavepoint("savepoint1")
		if err != nil {
			t.Fatalf("Failed to create savepoint: %v", err)
		}

		// 执行更多操作
		fields2 := map[string]any{"id": 2, "name": "test2"}
		_, err = tableTx.Insert(&fields2)
		if err != nil {
			t.Fatalf("Failed to insert record: %v", err)
		}

		// 回滚到保存点
		err = txMgr.RollbackToSavepoint("savepoint1")
		if err != nil {
			t.Fatalf("Failed to rollback to savepoint: %v", err)
		}

		// 验证回滚效果
		// 注意：由于我们只重置了Batch和恢复了保存点时的缓存状态，
		// 所以缓存中应该包含保存点之前的操作
		// 这里我们验证缓存中只有保存点之前的操作
		if len(tableTx.Cache) != 1 {
			t.Errorf("Expected 1 entry in cache after rollback, got %d entries", len(tableTx.Cache))
		}
	})

	// 测试场景2: 嵌套保存点
	t.Run("NestedSavepoints", func(t *testing.T) {
		// 执行一些操作
		fields1 := map[string]any{"id": 1, "name": "test1"}
		_, err := tableTx.Insert(&fields1)
		if err != nil {
			t.Fatalf("Failed to insert record: %v", err)
		}

		// 创建第一个保存点
		err = txMgr.CreateSavepoint("savepoint1")
		if err != nil {
			t.Fatalf("Failed to create savepoint1: %v", err)
		}

		// 执行更多操作
		fields2 := map[string]any{"id": 2, "name": "test2"}
		_, err = tableTx.Insert(&fields2)
		if err != nil {
			t.Fatalf("Failed to insert record: %v", err)
		}

		// 创建第二个保存点
		err = txMgr.CreateSavepoint("savepoint2")
		if err != nil {
			t.Fatalf("Failed to create savepoint2: %v", err)
		}

		// 执行更多操作
		fields3 := map[string]any{"id": 3, "name": "test3"}
		_, err = tableTx.Insert(&fields3)
		if err != nil {
			t.Fatalf("Failed to insert record: %v", err)
		}

		// 回滚到第一个保存点
		err = txMgr.RollbackToSavepoint("savepoint1")
		if err != nil {
			t.Fatalf("Failed to rollback to savepoint1: %v", err)
		}

		// 验证第二个保存点已被移除
		if _, exists := txMgr.savepoints["savepoint2"]; exists {
			t.Error("Expected savepoint2 to be removed after rollback to savepoint1")
		}

		// 验证回滚效果
		// 注意：由于我们只重置了Batch和恢复了保存点时的缓存状态，
		// 所以缓存中应该包含保存点之前的操作
		// 这里我们验证缓存中只有保存点之前的操作
		if len(tableTx.Cache) != 1 {
			t.Errorf("Expected 1 entry in cache after rollback, got %d entries", len(tableTx.Cache))
		}
	})

	// 测试场景3: 提交事务
	t.Run("CommitAfterSavepoint", func(t *testing.T) {
		// 执行一些操作
		fields1 := map[string]any{"id": 1, "name": "test1"}
		_, err := tableTx.Insert(&fields1)
		if err != nil {
			t.Fatalf("Failed to insert record: %v", err)
		}

		// 创建保存点
		err = txMgr.CreateSavepoint("savepoint1")
		if err != nil {
			t.Fatalf("Failed to create savepoint: %v", err)
		}

		// 执行更多操作
		fields2 := map[string]any{"id": 2, "name": "test2"}
		_, err = tableTx.Insert(&fields2)
		if err != nil {
			t.Fatalf("Failed to insert record: %v", err)
		}

		// 提交事务
		err = txMgr.Commit()
		if err != nil {
			t.Fatalf("Failed to commit transaction: %v", err)
		}

		// 验证事务已提交
		if !txMgr.committed {
			t.Error("Expected transaction to be committed")
		}
	})
}
