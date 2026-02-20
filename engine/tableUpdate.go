package engine

import (
	"fmt"

	"github.com/liaoran123/sfsDb/storage"
	"github.com/liaoran123/sfsDb/util"
)

// 检查是否提供了所有主键字段
func (t *Table) checkPrimaryFields(fields *map[string]any) error {
	for _, field := range t.GetPrimaryFields() {
		if _, ok := (*fields)[field]; !ok {
			return fmt.Errorf("必须提供主键字段 '%s'", field)
		}
	}
	return nil
}

// 准备更新字段列表
func (t *Table) prepareUpdateFields(fields *map[string]any) ([]string, error) {
	updateFields := GetStringSlice()
	for field := range *fields {
		// 排除主键字段
		if t.GetPrimaryKey().MatchFields(field) {
			continue
		}
		updateFields = append(updateFields, field)
	}
	return updateFields, nil
}

// 更新字段值
func (t *Table) updateFieldsValue(fields *map[string]any, fieldsBytes *map[string][]byte) {
	for field, val := range *fields {
		// 排除主键字段，主键字段不能更新
		if t.GetPrimaryKey().MatchFields(field) {
			continue
		}
		if _, ok := t.fields[field]; ok {
			(*fieldsBytes)[field] = util.AnyToBytes(val)
		}
	}
}

// 执行更新操作
func (t *Table) executeUpdateOperation(batch storage.Batch, fields *map[string]any, fieldsBytes *map[string][]byte, updateFields []string) error {
	// 从对象池中获取一个 batchContainer
	batchContainer := GetBatchContainer(batch, t.indexs, t.id, t.kvStore)
	defer PutBatchContainer(batchContainer)

	// 删除旧记录
	batchContainer.Operation(fieldsBytes, updateFields...)

	// 更新字段值
	t.updateFieldsValue(fields, fieldsBytes)

	// 格式化记录
	record := t.FormatRecord(fieldsBytes)

	// 添加新记录
	batchContainer.SetValue(0, record)                               // 添加主键value=record
	batchContainer.SetValue(1, t.GetPrimaryKey().GetID(fieldsBytes)) // 添加普通索引value=GetPrimaryKey().GetID()
	batchContainer.Operation(fieldsBytes, updateFields...)

	return nil
}

/*
// parseUpdateBatch 解析更新操作的参数，只返回 batch
func (t *Table) parseUpdateBatch(params ...any) storage.Batch {
	batch, _ := t.parseParams(params...)
	return batch
}

// parseUpdateParams 解析更新操作的参数
func (t *Table) parseUpdateParams(params ...any) (storage.Batch, time.Duration) {
	return t.parseParams(params...)
}

// prepareUpdateBatch 准备更新操作的batch
func (t *Table) prepareUpdateBatch(batch storage.Batch) (storage.Batch, bool, error) {
	return t.prepareBatch(batch)
}*/

// validateUpdateFields 验证更新操作的字段
func (t *Table) validateUpdateFields(fields *map[string]any) (string, any, error) {
	// 检查是否提供了所有主键字段
	if err := t.checkPrimaryFields(fields); err != nil {
		return "", nil, err
	}

	// 检查字段类型是否匹配
	if err := t.CheckType(fields); err != nil {
		return "", nil, err
	}

	// 获取主键值用于行级锁
	pkField := t.GetPrimaryFields()[0]
	pkValue := (*fields)[pkField]

	return pkField, pkValue, nil
}

// readRecordForUpdate 读取要更新的记录
func (t *Table) readRecordForUpdate(fields *map[string]any) ([]byte, error) {
	// 读取记录 - 直接使用 ReadByBytes 避免死锁
	fieldsBytes := GlobalFieldsBytesPool.Get()
	defer func() {
		if fieldsBytes != nil {
			GlobalFieldsBytesPool.Put(fieldsBytes)
		}
	}()
	// 只使用主键字段来构建键值
	for _, field := range t.GetPrimaryFields() {
		if val, ok := (*fields)[field]; ok {
			fieldsBytes[field] = util.AnyToBytes(val)
		}
	}
	key := t.GetPrimaryKey().JoinValue(&fieldsBytes, t.id)
	record := t.ReadByBytes(key)
	if record == nil {
		return nil, fmt.Errorf("主键值 '%v' 的记录不存在", fields)
	}

	return key, nil
}

// prepareUpdateFieldsList 准备更新字段列表
func (t *Table) prepareUpdateFieldsList(fields *map[string]any) ([]string, error) {
	// 准备更新字段列表
	updateFields, err := t.prepareUpdateFields(fields)
	if err != nil {
		return nil, err
	}

	// 检查更新字段个数是否为0
	if len(updateFields) == 0 {
		return nil, nil
	}

	return updateFields, nil
}

// deserializeRecord 反序列化记录
func (t *Table) deserializeRecord(record []byte) (*map[string][]byte, error) {
	// 反序列化记录
	pk := t.GetPrimaryKey()
	fieldsBytes, err := pk.Parse(t.fieldsid, record)
	if err != nil {
		return nil, err
	}

	return fieldsBytes, nil
}

/*
// commitUpdateTransaction 提交更新事务
func (t *Table) commitUpdateTransaction(batch storage.Batch, userProvidedBatch bool) error {
	return t.commitTransaction(batch, userProvidedBatch)
}
*/
// 更新记录，不支持修改主键字段
// fields *map[string]any 主键值，可能是组合主键
// 乐观锁并发控制，允许多个事务同时读取记录，但只有一个事务能成功更新记录，避免了并发更新冲突。
// 之前Update的缺省参数为batchs ...storage.Batch ，支持乐观锁需要增加一个参数，故而为兼容之前的函数，
// 使用使用 batchs ...storage.Batch  。batch和timeout合并为一个参数组数
func (t *Table) Update(fields *map[string]any, batchs ...storage.Batch) error {
	// 准备batch
	batch, userProvidedBatch, err := t.prepareBatch(batchs...)
	if err != nil {
		return err
	}

	// 验证字段
	_, _, err = t.validateUpdateFields(fields)
	if err != nil {
		return err
	}

	// 读取记录
	key, err := t.readRecordForUpdate(fields)
	if err != nil {
		return err
	}

	// 读取原始记录
	record := t.ReadByBytes(key)
	if record == nil {
		return fmt.Errorf("主键值 '%v' 的记录不存在", fields)
	}

	// 准备更新字段列表
	updateFields, err := t.prepareUpdateFieldsList(fields)
	if err != nil {
		return err
	}
	if updateFields == nil {
		return nil
	}

	// 反序列化记录
	fieldsBytes, err := t.deserializeRecord(record)
	if err != nil {
		return err
	}

	// 使用 UpdateImpl
	updateImpl := NewUpdateImpl(t, batch, userProvidedBatch, fields)
	updateImpl.fieldsBytes = fieldsBytes
	updateImpl.updateFields = updateFields
	updateImpl.key = key

	// 执行更新操作
	if err := updateImpl.ExecuteUpdateOperation(); err != nil {
		return err
	}

	// 提交事务
	if err := updateImpl.Commit(); err != nil {
		return err
	}

	return nil
}
