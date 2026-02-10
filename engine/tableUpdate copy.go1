package engine

import (
	"fmt"
	"time"

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

// 检查并更新版本号
func (t *Table) checkAndUpdateVersion(fields *map[string]any, fieldsBytes *map[string][]byte) (string, error) {
	// 获取当前版本号
	currentVersionBytes, exists := (*fieldsBytes)["v"]
	if !exists {
		currentVersionBytes = []byte{1} // 默认版本号
	}
	currentVersion := string(currentVersionBytes)

	// 检查版本号是否匹配
	if updateVersion, hasVersion := (*fields)["v"].(string); hasVersion {
		if updateVersion != currentVersion {
			return "", fmt.Errorf("optimistic lock conflict: version mismatch, expected %s, got %s", currentVersion, updateVersion)
		}
	}

	// 更新版本号为增强版版本号
	enhancedVersion := generateEnhancedVersion()
	(*fields)["v"] = enhancedVersion
	(*fieldsBytes)["v"] = []byte(enhancedVersion)

	return enhancedVersion, nil
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

	// 检查并更新版本号
	_, err := t.checkAndUpdateVersion(fields, fieldsBytes)
	if err != nil {
		return err
	}

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

// 更新记录，不支持修改主键字段
// fields *map[string]any 主键值，可能是组合主键
// 乐观锁并发控制，允许多个事务同时读取记录，但只有一个事务能成功更新记录，避免了并发更新冲突。
// 之前Update的缺省参数为batchs ...storage.Batch ，支持乐观锁需要增加一个参数，故而为兼容之前的函数，
// 使用使用 params ...any 。batch和timeout合并为一个参数组数
func (t *Table) Update(fields *map[string]any, params ...any) error {
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

	//是否用户手动控制事务
	userProvidedBatch := batch != nil

	// 如果没有提供batch，使用默认batch
	if batch == nil {
		batch = t.kvStore.GetBatch()
		if batch == nil {
			return fmt.Errorf("failed to get batch")
		}
	}

	// 检查是否提供了所有主键字段
	if err := t.checkPrimaryFields(fields); err != nil {
		return err
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

	// 检查字段类型是否匹配
	if err := t.CheckType(fields); err != nil {
		return err
	}

	// 读取记录 - 直接使用 ReadByBytes 避免死锁
	fieldsBytes := t.FieldsToBytes(fields)
	key := t.GetPrimaryKey().JoinValue(fieldsBytes, t.id)
	record := t.ReadByBytes(key)
	if record == nil {
		return fmt.Errorf("主键值 '%v' 的记录不存在", fields)
	}

	// 准备更新字段列表
	updateFields, err := t.prepareUpdateFields(fields)
	if err != nil {
		return err
	}
	defer PutStringSlice(updateFields)

	// 检查更新字段个数是否为0
	if len(updateFields) == 0 {
		return nil
	}

	// 反序列化记录
	pk := t.GetPrimaryKey()
	fieldsBytes, err = pk.Parse(t.fieldsid, record)
	if err != nil {
		return err
	}

	// 执行更新操作
	if err := t.executeUpdateOperation(batch, fields, fieldsBytes, updateFields); err != nil {
		return err
	}

	// 提交事务
	if !userProvidedBatch {
		if err := t.kvStore.WriteBatch(batch); err != nil {
			return err
		}
	}

	return nil
}
