package engine

import (
	"fmt"

	"github.com/liaoran123/sfsDb/storage"
)

// 删除记录
// fields *map[string]any 主键值，可能是组合主键
// 之前Delete的缺省参数为batchs ...storage.Batch ，支持乐观锁需要增加一个参数，故而为兼容之前的函数，
// 使用使用 batchs ...storage.Batch 。batch和timeout合并为一个参数组数
func (t *Table) Delete(fields *map[string]any, batchs ...storage.Batch) error {
	// 使用 DeleteImpl
	deleteImpl := NewDeleteImpl(t, fields)
	if err := deleteImpl.HasPrimaryKey(); err != nil {
		return err
	}
	deleteImpl.PrepareBatch(batchs...)
	if err := deleteImpl.ReadRecord(); err != nil {
		return err
	}
	if err := deleteImpl.DeleteRecord(); err != nil {
		return err
	}
	// 提交事务
	if err := deleteImpl.Commit(); err != nil {
		return err
	}
	return nil
}

// 删除表的所有数据
func (t *Table) DeleteAll() error {
	// 获取表的所有kv键值对迭代器
	iter := t.For()
	defer iter.Release()

	// 创建批量操作
	batch := t.kvStore.GetBatch()
	if batch == nil {
		return fmt.Errorf("failed to get batch")
	}

	// 定义批量操作的大小限制
	const batchSizeLimit = 1000

	// 遍历并删除所有键值对
	count := 0
	for iter.Next() {
		key := iter.Key()
		batch.Delete(key)
		count++

		// 当批量操作的大小达到限制时，执行批量操作并重置批量操作对象
		if count >= batchSizeLimit {
			// 提交批量操作
			if err := t.kvStore.WriteBatch(batch); err != nil {
				return err
			}

			// 重置计数器和批量操作对象
			count = 0
			batch = t.kvStore.GetBatch()
			if batch == nil {
				return fmt.Errorf("failed to get batch")
			}
		}
	}

	// 执行剩余的批量操作
	if count > 0 {
		if err := t.kvStore.WriteBatch(batch); err != nil {
			return err
		}
	}

	return nil
}
