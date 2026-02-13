package engine

import (
	"fmt"

	"github.com/liaoran123/sfsDb/storage"
	"github.com/liaoran123/sfsDb/util"
)

// prepareInsertBatch 准备插入操作的batch
func (t *Table) prepareInsertBatch(batchs ...storage.Batch) (storage.Batch, bool, error) {
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

// commitInsertTransaction 提交插入事务
func (t *Table) commitInsertTransaction(batch storage.Batch, userProvidedBatch bool) error {
	//提交事务
	if !userProvidedBatch {
		if err := t.kvStore.WriteBatch(batch); err != nil {
			return err
		}
	}

	return nil
}

// handleAutoIncrement 处理自动增值主键
func (t *Table) handleAutoIncrement(fields *map[string]any) (int, error) {
	//获取主键字段
	primaryFields := t.GetPrimaryFields()
	if len(primaryFields) == 0 {
		return -1, fmt.Errorf("表 '%s' 没有设置主键", t.name)
	}

	//是否支持默认自动增值主键，单主键并且主键字段名为"id"
	pklen := len(primaryFields)
	pkfield := primaryFields[0]
	supportDefault := pklen == 1 && pkfield == "id"
	currentID := -1

	if supportDefault {
		// 检查是否提供了主键字段
		//使用默认自动增值主键时，不需要提供主键字段，系统自动生成，强制使用"id"字段和自动增值主键
		_, ok := (*fields)[pkfield]
		if !ok { //未提供主键字段，自动生成主键值
			currentID = t.GetAutoInc()
			(*fields)[pkfield] = currentID
		} else { //提供了主键字段id，但是值为nil，自动生成主键值
			if (*fields)[pkfield] == nil {
				currentID = t.GetAutoInc()
				(*fields)[pkfield] = currentID
			}
		}
	}

	// 检查字段类型是否匹配
	if err := t.CheckType(fields); err != nil {
		return -1, err
	}

	currentID = util.AnyToInt((*fields)[pkfield])
	// 添加初始版本号
	if _, hasVersion := (*fields)["v"]; !hasVersion || (*fields)["v"] == "" {
		(*fields)["v"] = generateEnhancedVersion() // 使用增强版版本号
	}

	return currentID, nil
}

// 插入记录
func (t *Table) Insert(fields *map[string]any, batchs ...storage.Batch) (currentID int, err error) {
	// 检查参数
	if t.fields == nil {
		return 0, fmt.Errorf("表 '%s' 未设置字段和类型", t.name)
	}
	if fields == nil {
		return 0, fmt.Errorf("fields cannot be nil")
	}

	// 处理自动增值主键
	currentID, err = t.handleAutoIncrement(fields)
	if err != nil {
		return -1, err
	}

	// 准备批量操作
	batch, userProvidedBatch, err := t.prepareInsertBatch(batchs...)
	if err != nil {
		return -1, err
	}

	// 转换字段为字节数组
	fieldsBytes := t.FieldsToBytes(fields)
	defer func() {
		if fieldsBytes != nil && *fieldsBytes != nil {
			GlobalFieldsBytesPool.Put(*fieldsBytes)
		}
	}()

	// 格式化记录
	record := t.FormatRecord(fieldsBytes)

	// 从对象池中获取一个 batchContainer
	BatchContainer := GetBatchContainer(batch, t.indexs, t.id, t.kvStore)
	defer PutBatchContainer(BatchContainer)

	// 设置值并执行操作
	BatchContainer.SetValue(0, record)                               //添加主键value=record
	BatchContainer.SetValue(1, t.GetPrimaryKey().GetID(fieldsBytes)) //添加普通索引value=GetPrimaryKey().GetID()
	BatchContainer.Operation(fieldsBytes)                            //添加全文索引key=joinValue,value=nil

	// 提交事务
	if err := t.commitInsertTransaction(batch, userProvidedBatch); err != nil {
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
	// 检查参数
	if t.fields == nil {
		return nil, fmt.Errorf("表 '%s' 未设置字段和类型", t.name)
	}
	if len(records) == 0 {
		return []int{}, nil
	}
	if records == nil {
		return nil, fmt.Errorf("records cannot be nil")
	}

	// 获取主键字段
	primaryFields := t.GetPrimaryFields() //t.GetPrimaryKey().GetFields()
	if len(primaryFields) == 0 {
		return nil, fmt.Errorf("表 '%s' 没有设置主键", t.name)
	}

	// 是否支持默认自动增值主键，单主键并且主键字段名为"id"
	pklen := len(primaryFields)
	pkfield := primaryFields[0]
	supportDefault := pklen == 1 && pkfield == "id"

	// 处理批量操作
	var batch storage.Batch
	if len(batchs) > 0 { // 用户手动控制事务
		batch = batchs[0]
		if batch == nil {
			return nil, fmt.Errorf("batch cannot be nil")
		}
	} else {
		batch = t.kvStore.GetBatch()
		if batch == nil {
			return nil, fmt.Errorf("failed to get batch")
		}
	}

	// 预分配ID列表容量
	ids := make([]int, len(records))

	// 计算需要自动生成的ID数量
	autoIncCount := 0
	for _, fields := range records {
		if fields == nil {
			return nil, fmt.Errorf("record cannot be nil")
		}
		if supportDefault {
			if _, ok := (*fields)[pkfield]; !ok || (*fields)[pkfield] == nil {
				autoIncCount++
			}
		}
	}

	// 批量获取自动增值ID，确保并发安全
	var autoIncStart int
	if supportDefault && autoIncCount > 0 {
		autoIncStart = t.GetAutoIncBatch(autoIncCount)
		// 后续ID可以直接计算，不需要重复调用GetAutoInc()
	}

	// 处理记录并批量插入
	autoIncIdx := 0
	// 从对象池中获取一个 batchContainer
	BatchContainer := GetBatchContainer(batch, t.indexs, t.id, t.kvStore)
	defer PutBatchContainer(BatchContainer)

	for i, fields := range records {
		// 检查字段类型
		if err := t.CheckType(fields); err != nil {
			return nil, err
		}

		// 处理自动增值主键
		if supportDefault {
			if _, ok := (*fields)[pkfield]; !ok || (*fields)[pkfield] == nil {
				// 使用预分配的自动增值ID
				ids[i] = autoIncStart + autoIncIdx
				(*fields)[pkfield] = ids[i]
				autoIncIdx++
			} else {
				// 使用提供的主键值
				ids[i] = util.AnyToInt((*fields)[pkfield])
			}
		} else {
			// 非默认自动增值主键，使用提供的主键值
			ids[i] = util.AnyToInt((*fields)[pkfield])
		}

		// 添加初始版本号
		if _, hasVersion := (*fields)["v"]; !hasVersion || (*fields)["v"] == "" {
			(*fields)["v"] = generateEnhancedVersion() // 使用增强版版本号
		}

		// 转换字段为字节数组
		fieldsBytes := t.FieldsToBytes(fields)

		// 格式化记录
		record := t.FormatRecord(fieldsBytes)

		// 批量添加记录
		BatchContainer.SetValue(0, record)                               // 添加主键value=record
		BatchContainer.SetValue(1, t.GetPrimaryKey().GetID(fieldsBytes)) // 添加普通索引value=GetPrimaryKey().GetID()
		// 添加全文索引key=joinValue,value=nil
		BatchContainer.Operation(fieldsBytes)
	}

	// 提交批量操作
	if len(batchs) == 0 {
		if err := t.kvStore.WriteBatch(batch); err != nil {
			return nil, err
		}
	}

	return ids, nil
}

// BatchInsertWithSize 带批量大小控制的批量插入
// records []*map[string]any 要插入的记录列表
// batchSize int 每批处理的记录数量
// batchs ...storage.Batch 可选的批量操作容器
// 返回值：插入记录的ID列表和错误信息
func (t *Table) BatchInsertWithSize(records []*map[string]any, batchSize int, batchs ...storage.Batch) ([]int, error) {
	// 检查参数
	if batchSize <= 0 {
		batchSize = 100 // 默认批量大小
	}

	// 计算总批次
	totalRecords := len(records)
	if totalRecords == 0 {
		return []int{}, nil
	}

	// 预分配ID列表
	allIds := make([]int, totalRecords)

	// 分批处理
	for start := 0; start < totalRecords; start += batchSize {
		end := start + batchSize
		if end > totalRecords {
			end = totalRecords
		}

		// 处理当前批次
		batchRecords := records[start:end]
		batchIds, err := t.BatchInsert(batchRecords, batchs...)
		if err != nil {
			return nil, err
		}

		// 复制ID到结果列表
		copy(allIds[start:end], batchIds)
	}

	return allIds, nil
}
