package engine

import (
	"fmt"
	"time"

	"github.com/liaoran123/sfsDb/storage"
)

// parseDeleteParams 解析删除操作的参数
func (t *Table) parseDeleteParams(params ...any) (storage.Batch, time.Duration) {
	return t.parseParams(params...)
}

// prepareDeleteBatch 准备删除操作的batch
func (t *Table) prepareDeleteBatch(batch storage.Batch) (storage.Batch, bool, error) {
	return t.prepareBatch(batch)
}

// validateDeleteFields 验证删除操作的字段
func (t *Table) validateDeleteFields(fields *map[string]any) (any, error) {
	// 检查是否提供了所有主键字段
	for _, field := range t.GetPrimaryFields() {
		if _, ok := (*fields)[field]; !ok {
			return nil, fmt.Errorf("必须提供主键字段 '%s'", field)
		}
	}

	// 获取主键值用于行级锁
	pkField := t.GetPrimaryFields()[0]
	pkValue := (*fields)[pkField]

	return pkValue, nil
}

// readRecordForDelete 读取要删除的记录
func (t *Table) readRecordForDelete(fields *map[string]any) ([]byte, error) {
	//读取记录 - 直接使用 ReadByBytes 避免死锁
	fieldsBytes := t.FieldsToBytes(fields)
	defer func() {
		if fieldsBytes != nil && *fieldsBytes != nil {
			GlobalFieldsBytesPool.Put(*fieldsBytes)
		}
	}()
	key := t.GetPrimaryKey().JoinValue(fieldsBytes, t.id)
	record := t.ReadByBytes(key)
	if record == nil {
		return nil, fmt.Errorf("主键值 '%v' 的记录不存在", fields)
	}

	return key, nil
}

// commitDeleteTransaction 提交删除事务
func (t *Table) commitDeleteTransaction(batch storage.Batch, userProvidedBatch bool) error {
	return t.commitTransaction(batch, userProvidedBatch)
}

// 删除记录
// fields *map[string]any 主键值，可能是组合主键
// 之前Delete的缺省参数为batchs ...storage.Batch ，支持乐观锁需要增加一个参数，故而为兼容之前的函数，
// 使用使用 params ...any 。batch和timeout合并为一个参数组数
func (t *Table) Delete(fields *map[string]any, params ...any) error {
	// 解析参数
	batch, timeout := t.parseDeleteParams(params...)

	// 验证字段
	pkValue, err := t.validateDeleteFields(fields)
	if err != nil {
		return err
	}

	// 获取行级排他锁
	lockKey := fmt.Sprintf("%v", pkValue)
	if err := t.acquireRowWriteLock(pkValue, 0, timeout); err != nil {
		return err
	}

	// 释放行级锁
	defer func() {
		if rowLock, ok := t.rowLocks.Load(lockKey); ok {
			rl := rowLock.(*RowLock)
			rl.rwLock.Unlock()
		}
	}()

	// 读取记录
	key, err := t.readRecordForDelete(fields)
	if err != nil {
		return err
	}

	// 读取原始记录
	record := t.ReadByBytes(key)
	if record == nil {
		return fmt.Errorf("主键值 '%v' 的记录不存在", fields)
	}

	// 准备batch
	batch, userProvidedBatch, err := t.prepareDeleteBatch(batch)
	if err != nil {
		return err
	}

	// 反序列化记录
	pk := t.GetPrimaryKey()
	fieldsBytes, err := pk.Parse(t.fieldsid, record)
	if err != nil {
		return err
	}

	// 从对象池中获取一个 batchContainer
	BatchContainer := GetBatchContainer(batch, t.indexs, t.id, t.kvStore)
	defer PutBatchContainer(BatchContainer)

	// 对于删除操作，需要将values[0]设置为nil，这样Add方法才会执行删除操作
	BatchContainer.SetValue(0, nil)
	BatchContainer.Operation(fieldsBytes)

	// 提交事务
	if err := t.commitDeleteTransaction(batch, userProvidedBatch); err != nil {
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
