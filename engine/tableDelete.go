package engine

import (
	"fmt"
	"time"

	"github.com/liaoran123/sfsDb/storage"
)

// 删除记录
// fields *map[string]any 主键值，可能是组合主键
// 之前Delete的缺省参数为batchs ...storage.Batch ，支持乐观锁需要增加一个参数，故而为兼容之前的函数，
// 使用使用 params ...any 。batch和timeout合并为一个参数组数
func (t *Table) Delete(fields *map[string]any, params ...any) error {
	// 解析参数，支持超时参数和batch参数
	var batch storage.Batch
	var timeout time.Duration

	// 处理可变参数
	for _, param := range params {
		switch v := param.(type) {
		case storage.Batch:
			batch = v
		case time.Duration:
			timeout = v
		}
	}

	// 如果没有提供batch，使用默认batch
	if batch == nil {
		batch = t.kvStore.GetBatch()
		if batch == nil {
			return fmt.Errorf("failed to get batch")
		}
	}
	//检查是否提供了所有主键字段
	for _, field := range t.GetPrimaryFields() {
		if _, ok := (*fields)[field]; !ok {
			return fmt.Errorf("必须提供主键字段 '%s'", field)
		}
	}

	// 获取主键值用于行级锁
	pkField := t.GetPrimaryFields()[0]
	pkValue := (*fields)[pkField]

	// 获取行级排他锁（使用默认事务ID）
	if err := t.acquireRowWriteLock(pkValue, 0, timeout); err != nil {
		return err
	}
	// 直接使用 Unlock 释放写锁
	lockKey := fmt.Sprintf("%v", pkValue)
	defer func() {
		if rowLock, ok := t.rowLocks.Load(lockKey); ok {
			rl := rowLock.(*RowLock)
			rl.rwLock.Unlock()
		}
	}()

	//读取记录 - 直接使用 ReadByBytes 避免死锁
	fieldsBytes := t.FieldsToBytes(fields)
	key := t.GetPrimaryKey().JoinValue(fieldsBytes, t.id)
	record := t.ReadByBytes(key)
	if record == nil {
		return fmt.Errorf("主键值 '%v' 的记录不存在", fields)
	}

	//是否用户手动控制事务
	useBatch := batch != nil
	if !useBatch { //用户未手动控制事务，创建新batch
		batch = t.kvStore.GetBatch()
	}

	//反序列化记录，并且将字段值转换为对应的类型
	//fieldsBytes := t.ParseRecord(record)
	pk := t.GetPrimaryKey()
	fieldsBytes, err := pk.Parse(t.fieldsid, record)
	if err != nil {
		//释放batch资源
		return err
	}
	// 从对象池中获取一个 batchContainer
	BatchContainer := GetBatchContainer(batch, t.indexs, t.id, t.kvStore)
	defer PutBatchContainer(BatchContainer)
	BatchContainer.Operation(fieldsBytes)
	if !useBatch { //用户未手动控制事务，自动提交
		t.kvStore.WriteBatch(batch)
	}
	//fmt.Printf("Delete BatchContainer.Len(): %v\n", BatchContainer.Len())
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
