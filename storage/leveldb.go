package storage

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// LevelDBGetter 定义LevelDB读取操作的公共接口，适用于DB和Snapshot
type LevelDBGetter interface {
	// Get 获取指定key的值
	Get(key []byte, ro *opt.ReadOptions) (value []byte, err error)

	// NewIterator 创建迭代器
	NewIterator(slice *util.Range, ro *opt.ReadOptions) iterator.Iterator
}

// LevelDBStore LevelDB存储实现，支持快照功能
type LevelDBStore struct {
	ldb        LevelDBGetter // 可以是*leveldb.DB或*leveldb.Snapshot
	originalDB *leveldb.DB   // 保存原始数据库实例，用于快照切换回数据库模式
	isSnapshot bool          // 标记是否为快照
	opts       *opt.Options
}

// Get 获取指定key的值
func (s *LevelDBStore) Get(key []byte) ([]byte, error) {
	if s.ldb == nil {
		return nil, NewError("ldb is nil, cannot perform read operations")
	}
	value, err := s.ldb.Get(key, nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return value, nil
}

// Put 设置key-value对
func (s *LevelDBStore) Put(key []byte, value []byte) error {

	// 总是使用originalDB执行写操作，无论当前是否为快照模式
	db := s.originalDB
	if db == nil {
		return NewError("originalDB is nil, cannot perform write operations")
	}

	return db.Put(key, value, nil)
}

// Delete 删除指定key
func (s *LevelDBStore) Delete(key []byte) error {

	// 总是使用originalDB执行写操作，无论当前是否为快照模式
	db := s.originalDB
	if db == nil {
		return NewError("originalDB is nil, cannot perform write operations")
	}
	return db.Delete(key, nil)
}

// GetBatch 获取批处理操作对象
//
// 返回值：
//
//	Batch - 批处理操作对象，可用于执行多个写操作
//
// 说明：
//   - 从批处理对象池中获取一个批处理对象，实现了对象复用
//   - 即使在快照模式下也返回有效的批处理对象
//   - 批处理对象的写操作会通过WriteBatch方法使用originalDB执行
//   - 如果对象池获取失败，会创建一个新的批处理对象
func (s *LevelDBStore) GetBatch() Batch {
	// 从batchPool中获取一个Batch对象
	// 即使在快照模式下也返回有效的Batch，WriteBatch会使用originalDB执行写操作
	return LdbBatchPool.Get()
}

// WriteBatch 执行批量写入操作
// batch: 批量操作对象
// put: 是否将batch放回对象池，默认是true
func (s *LevelDBStore) WriteBatch(batch Batch, put ...bool) error {
	// 检查batch是否为nil
	if batch == nil {
		return NewError("batch cannot be nil")
	}

	// 总是使用originalDB执行写操作，无论当前是否为快照模式
	db := s.originalDB
	if db == nil {
		return NewError("originalDB is nil, cannot perform write operations")
	}

	// 确保batch是*leveldb.Batch类型
	ldbBatch, ok := batch.(*leveldb.Batch)
	if !ok {
		return NewError("invalid batch type")
	}

	// 检查batch是否为空
	if ldbBatch.Len() == 0 {
		// 空batch直接返回，无需写入
		if len(put) == 0 || put[0] {
			LdbBatchPool.Put(ldbBatch)
		} else {
			// 即使不放回对象池，也需要重置batch，以便后续使用
			ldbBatch.Reset()
		}
		return nil
	}

	// 执行批量写入操作
	err := db.Write(ldbBatch, nil)

	// 处理写入结果
	if err != nil {
		// 写入失败，也可以将batch放回池中
		if len(put) == 0 || put[0] {
			LdbBatchPool.Put(ldbBatch)
		} else {
			// 即使不放回对象池，也需要重置batch，以便后续使用
			ldbBatch.Reset()
		}
		return err
	}

	// 写入成功，处理batch
	if len(put) == 0 || put[0] {
		LdbBatchPool.Put(ldbBatch)
	} else {
		// 即使不放回对象池，也需要重置batch，以便后续使用
		ldbBatch.Reset()
	}

	return nil
}

/*
// Iterator 创建迭代器

	func (s *LevelDBStore) Iterator1(slice *util.Range) Iterator {
		return s.ldb.NewIterator(slice, nil)
	}
*/
type FunIter func(start, limit []byte) Iterator

// Iterator 创建迭代器
func (s *LevelDBStore) Iterator(start, limit []byte) Iterator {
	if s.ldb == nil {
		return nil
	}
	return s.ldb.NewIterator(&util.Range{Start: start, Limit: limit}, nil)
}

// Snapshot 创建快照
func (s *LevelDBStore) Snapshot() (Snapshot, error) {

	// 快照不支持创建快照
	if s.isSnapshot {
		return nil, NewError("cannot create snapshot from snapshot")
	}

	// 确保ldb是*leveldb.DB类型
	db, ok := s.ldb.(*leveldb.DB)
	if !ok {
		return nil, NewError("invalid database type for creating snapshot")
	}

	// 创建LevelDB快照
	levelDBSnapshot, err := db.GetSnapshot()
	if err != nil {
		return nil, err
	}
	if levelDBSnapshot == nil {
		return nil, NewError("failed to create snapshot")
	}

	// 从对象池中获取快照实例
	snapshotStore := LdbSnapshotPool.Get()
	// 设置快照状态
	snapshotStore.ldb = levelDBSnapshot
	snapshotStore.originalDB = db
	snapshotStore.isSnapshot = true
	snapshotStore.opts = s.opts

	return snapshotStore, nil
}

// Close 关闭存储
func (s *LevelDBStore) Close() error {

	// 如果是快照，释放快照资源并将实例放回对象池
	if s.isSnapshot {
		if snapshot, ok := s.ldb.(*leveldb.Snapshot); ok {
			snapshot.Release()
		}
		// 将实例放回对象池
		LdbSnapshotPool.Put(s)
		return nil
	}

	// 如果是数据库实例，关闭数据库
	if db, ok := s.ldb.(*leveldb.DB); ok {
		err := db.Close()
		if err != nil {
			return err
		}
		s.ldb = nil
		return nil
	}

	return NewError("invalid database type")
}

// Release 释放快照资源
// 对于普通数据库实例，Release方法不做任何事情
// 对于快照实例，Release方法会释放快照资源
func (s *LevelDBStore) Release() error {
	// 如果是快照，调用Close方法释放资源
	if s.isSnapshot {
		return s.Close()
	}
	// 对于普通数据库实例，Release方法不做任何事情
	return nil
}

// SwitchToSnapshot 将当前实例切换到快照模式
// 如果当前已是快照模式，则返回错误
func (s *LevelDBStore) SwitchToSnapshot() error {
	// 如果当前已是快照模式，不能再切换到快照
	if s.isSnapshot {
		return NewError("cannot switch to snapshot from snapshot")
	}

	// 确保当前ldb是*leveldb.DB类型
	db, ok := s.ldb.(*leveldb.DB)
	if !ok {
		return NewError("invalid database type for creating snapshot")
	}

	// 创建新的快照
	snapshot, err := db.GetSnapshot()
	if err != nil {
		return err
	}
	if snapshot == nil {
		return NewError("failed to create snapshot")
	}

	// 切换ldb字段为新快照
	s.ldb = snapshot
	s.isSnapshot = true

	return nil
}

// SwitchToDB 将当前实例切换回数据库模式
// 如果当前已是数据库模式，则返回错误
// 如果当前是快照模式，会先释放快照资源
func (s *LevelDBStore) SwitchToDB() error {
	// 如果当前已是数据库模式，不需要切换
	if !s.isSnapshot {
		return NewError("already in database mode")
	}

	// 先释放当前快照资源
	if snapshot, ok := s.ldb.(*leveldb.Snapshot); ok {
		snapshot.Release()
	}

	// 确保有原始数据库实例可以切换回
	if s.originalDB == nil {
		return NewError("no original database instance available for switching back")
	}

	// 切换ldb字段为原始数据库实例
	s.ldb = s.originalDB
	s.isSnapshot = false

	return nil
}
