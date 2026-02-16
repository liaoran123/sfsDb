package engine

import (
	"fmt"

	"github.com/liaoran123/sfsDb/storage"
)

/*
// prepareInsertBatch 准备插入操作的batch
func (t *Table) prepareInsertBatch1(batchs ...storage.Batch) (storage.Batch, bool, error) {
	var batch storage.Batch
	userProvidedBatch := len(batchs) > 0

	//是否用户手动控制事务
	if userProvidedBatch { //用户手动控制事务
		batch = batchs[0]
		if batch == nil {
			return nil, false, fmt.Errorf("batch cannot be nil")
		}
	} else {
		batch = t.kvStore.GetBatch()
		if batch == nil {
			return nil, false, fmt.Errorf("failed to get batch")
		}
	}

	return batch, userProvidedBatch, nil
}
*/
// 插入记录
func (t *Table) Insert(fields *map[string]any, batchs ...storage.Batch) (currentID int, err error) {
	// 检查参数
	if t.fields == nil {
		return 0, fmt.Errorf("表 '%s' 未设置字段和类型", t.name)
	}
	if fields == nil {
		return 0, fmt.Errorf("fields cannot be nil")
	}

	// 准备批量操作
	batch, userProvidedBatch, err := t.prepareBatch(batchs...)
	if err != nil {
		return -1, err
	}

	// 使用 InsertImpl
	insertImpl := NewInsertImpl(t, batch, userProvidedBatch, fields)

	// 处理自动增值主键
	currentID, err = insertImpl.AutoIncrement()
	if err != nil {
		return -1, err
	}
	// 检查字段类型是否匹配
	if err := insertImpl.CheckType(); err != nil {
		return -1, err
	}
	// 添加初始版本号
	insertImpl.AddVersion()
	// 提交事务
	if err := insertImpl.Commit(); err != nil {
		return -1, err
	}

	return currentID, nil
}

/*
	数据流动流程
	1,外部传入 Insert(fields *map[string]any
	2,// 转换字段为字节数组
	fieldsBytes := t.FieldsToBytes(fields)
	3,//格式化记录
	record := t.FormatRecord(fieldsBytes)
	4,添加更新记录
	tableiter 查询功能则与上面添加的流程相反。一正一逆。
*/
// BatchInsert 批量插入多条记录
// records []*map[string]any 要插入的记录列表
// batchs ...storage.Batch 可选的批量操作容器
// 返回值：插入记录的ID列表和错误信息
func (t *Table) BatchInsert(records []*map[string]any, batchs ...storage.Batch) ([]int, error) {
	// 使用 InsertImpl
	insertImpl := NewBatchInsertImpl(t, nil, false, records)

	// 执行批量插入
	return insertImpl.BatchInsert(records, batchs...)
}

// BatchInsertWithSize 带批量大小控制的批量插入
// records []*map[string]any 要插入的记录列表
// batchSize int 每批处理的记录数量
// batchs ...storage.Batch 可选的批量操作容器
// 返回值：插入记录的ID列表和错误信息
func (t *Table) BatchInsertWithSize(records []*map[string]any, batchSize int, batchs ...storage.Batch) ([]int, error) {
	// 准备批量操作
	_, userProvidedBatch, err := t.prepareBatch(batchs...)
	if err != nil {
		return nil, err
	}

	// 使用 InsertImpl
	insertImpl := NewBatchInsertImpl(t, nil, userProvidedBatch, records)

	// 执行带批量大小控制的批量插入
	return insertImpl.BatchInsertWithSize(records, batchSize, batchs...)
}

// BatchInsertWithSizeNoInc 带批量大小控制的批量插入（不需要自动增值）
// records []*map[string]any 要插入的记录列表
// batchSize int 每批处理的记录数量
// skipVersion bool 是否跳过版本号
// batchs ...storage.Batch 可选的批量操作容器
// 返回值：插入记录的ID列表和错误信息
func (t *Table) BatchInsertWithSizeNoInc(records []*map[string]any, batchSize int, skipVersion bool, batchs ...storage.Batch) ([]int, error) {
	// 使用 InsertImpl
	insertImpl := NewBatchInsertImpl(t, nil, false, records)
	// 执行带批量大小控制的批量插入
	return insertImpl.BatchInsertWithSizeNoInc(records, batchSize, skipVersion, batchs...)
}

// BatchInsertNoInc 批量插入不需要自动增值的记录
// records []*map[string]any 要插入的记录列表
// skipVersion bool 是否跳过版本号
// batchs ...storage.Batch 可选的批量操作容器
// 返回值：插入记录的ID列表和错误信息
func (t *Table) BatchInsertNoInc(records []*map[string]any, skipVersion bool, batchs ...storage.Batch) ([]int, error) {
	// 使用 InsertImpl
	insertImpl := NewBatchInsertImpl(t, nil, false, records)
	// 执行批量插入
	return insertImpl.BatchInsertNoInc(records, skipVersion, batchs...)
}
