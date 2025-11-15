package db

/*
ACID 特性	LevelDB 实现方式
原子性	WriteBatch 批量提交
一致性	应用层逻辑约束 + 原子写
隔离性	单线程写入，无并发修改
持久性	写前日志（WAL）+ SSTable 持久化

通过ApproximateSize()方法监控批次大小，防止内存溢出：sDB
WriteBatch.SetMaxBatchSize(size int)

在合并两个 WriteBatch 的时候，也会累计两部分的 count 的值，如下 WriteBatchInternal::Append 方法：

void WriteBatchInternal::Append(WriteBatch* dst, const WriteBatch* src) {
  SetCount(dst, Count(dst) + Count(src));
  assert(src->rep_.size() >= kHeader);
  dst->rep_.append(src->rep_.data() + kHeader, src->rep_.size() - kHeader);
}

*/

import (
	"fmt"
	"log"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// 示例：打开数据库（可传入 opt），并演示常用操作
func ExampleLevelDB(path string) error {
	// 配置并打开
	opts := &opt.Options{
		// 示例：最大打开文件数、写入缓存等可在此配置
	}
	db, err := leveldb.OpenFile(path, opts)
	if err != nil {
		// 如果损坏可以尝试 RecoverFile
		if db, err = leveldb.RecoverFile(path, nil); err != nil {
			return fmt.Errorf("repair failed: %v", err)
		}
		// 如果损坏可以尝试 Repair
		/*
			   // 如果损坏可以尝试 Repair


			if _, ok := err.(*leveldb.ErrCorrupted); ok {
				if db, err= leveldb.RecoverFile(path, nil); err != nil {
					return fmt.Errorf("repair failed: %v", err)
				}
				// 重新打开
				db, err = leveldb.OpenFile(path, opts)
				if err != nil {
					return err
				}
			} else {
				return err
			}*/
	}
	defer db.Close()

	// Put / Get / Delete
	if err := db.Put([]byte("k1"), []byte("v1"), &opt.WriteOptions{Sync: true}); err != nil {
		return err
	}
	val, err := db.Get([]byte("k1"), nil)
	if err != nil {
		return err
	}
	log.Printf("k1=%s\n", string(val))

	if err := db.Delete([]byte("k1"), nil); err != nil {
		return err
	}

	// 原子批量写（WriteBatch）
	batch := new(leveldb.Batch)
	batch.Put([]byte("a"), []byte("1"))
	batch.Put([]byte("b"), []byte("2"))
	batch.Delete([]byte("c"))
	if err := db.Write(batch, &opt.WriteOptions{}); err != nil { //&opt.WriteOptions{Sync: true}
		return err
	}
	//batch.Append() 合并其他 Batch 示例
	//源代码是batch.append()不开放该功能。需要修改源码才能使用
	otherBatch := new(leveldb.Batch)
	otherBatch.Put([]byte("d"), []byte("4"))
	batch.Append(otherBatch)
	if err := db.Write(batch, &opt.WriteOptions{}); err != nil { //&opt.WriteOptions{Sync: true}
		return err
	}
	// 迭代器 - 全库遍历
	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()
		log.Printf("iter: %s = %s\n", string(key), string(value))
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return err
	}

	// 前缀扫描（使用 util.Range）
	prefix := []byte("user:")
	r := util.BytesPrefix(prefix) // start=prefix, limit=prefix+0x00.. style
	iter = db.NewIterator(r, nil)
	for iter.Next() {
		log.Printf("prefix: %s = %s\n", string(iter.Key()), string(iter.Value()))
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return err
	}
	// 范围扫描
	iter = db.NewIterator(&util.Range{Start: []byte("foo"), Limit: []byte("xoo")}, nil)
	for iter.Next() {
		log.Printf("prefix: %s = %s\n", string(iter.Key()), string(iter.Value()))
	}
	iter.Release()
	err = iter.Error()
	if err := iter.Error(); err != nil {
		return err
	}
	// 使用 Snapshot 进行一致读
	//通过channel控制一个全局的读快照，保证在同一时间点读取数据的一致性
	//存在Snapshot时，用Snapshot读，释放Snapshot后，则用db读。
	snap, err := db.GetSnapshot()
	if err != nil {
		return err
	}
	iter = snap.NewIterator(&util.Range{Start: []byte("foo"), Limit: []byte("xoo")}, nil)
	for iter.Next() {
		log.Printf("prefix: %s = %s\n", string(iter.Key()), string(iter.Value()))
	}
	defer snap.Release()
	v, err := snap.Get([]byte("a"), nil)
	if err != nil && err != leveldb.ErrNotFound {
		return err
	}
	log.Printf("snapshot a=%s\n", string(v))

	// SizeOf 示例（range 大小估计）
	ranges := []util.Range{{Start: []byte("a"), Limit: []byte("z")}}
	sz, err := db.SizeOf(ranges)
	if err == nil {
		log.Printf("approx size: %d\n", sz)
	}

	// CompactRange 会对给定键范围的底层数据库进行压缩。特别是，已删除和被覆盖的版本会被丢弃，数据会被重新排列，以减少访问数据所需操作的成本。
	if err := db.CompactRange(util.Range{Start: []byte("a"), Limit: []byte("z")}); err != nil {
		log.Printf("compact error: %v\n", err)
	}

	// 获取内部属性（示例）
	if prop, ok := db.GetProperty("leveldb.stats"); ok != nil {
		log.Println("leveldb.stats:", prop)
	}
	// OpenTransaction 打开一个原子数据库事务。一次只能打开一个事务。
	// 后续对 Write 和 OpenTransaction 的调用将被阻塞，直到当前事务被提交或丢弃。
	// 返回的事务句柄可安全并发使用。

	// 事务开销大，尤其是在事务规模较小时，可能会压垮压缩操作。请谨慎使用。

	// 事务完成后必须关闭，要么通过提交，要么通过丢弃事务。
	// 关闭数据库将丢弃未完成的事务。
	//func (db *DB) Write(batch *Batch, wo *opt.WriteOptions)内部就是调用OpenTransaction()
	//tr所有操作都是在db的操作上加上一个锁而已，并且可以回滚。
	tr, err := db.OpenTransaction()
	tr.Delete([]byte("a"), nil)
	tr.Put([]byte("b"), []byte("20"), nil)
	if err := tr.Commit(); err != nil {
		return err
	}
	//数据是否存在
	_, err = db.Has([]byte("b"), nil)
	return nil
}
