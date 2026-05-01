package engine

import (
	"fmt"
	"sync"

	"github.com/liaoran123/sfsDb/storage"
	"github.com/liaoran123/sfsDb/util"
)

// InsertImplPool 是 InsertImpl 的对象池
type InsertImplPool struct {
	pool sync.Pool
}

// 全局 InsertImpl 对象池
var GlobalInsertImplPool = &InsertImplPool{
	pool: sync.Pool{
		New: func() interface{} {
			return &InsertImpl{}
		},
	},
}

// BatchInsertImplPool 是 BatchInsertImpl 的对象池
type BatchInsertImplPool struct {
	pool sync.Pool
}

// 全局 BatchInsertImpl 对象池
var GlobalBatchInsertImplPool = &BatchInsertImplPool{
	pool: sync.Pool{
		New: func() interface{} {
			return &BatchInsertImpl{}
		},
	},
}

// Get 从对象池中获取一个 BatchInsertImpl 实例
func (p *BatchInsertImplPool) Get() *BatchInsertImpl {
	return p.pool.Get().(*BatchInsertImpl)
}

// Put 将 BatchInsertImpl 实例放回对象池
func (p *BatchInsertImplPool) Put(impl *BatchInsertImpl) {
	// 重置实例状态
	impl.Reset()
	p.pool.Put(impl)
}

// Get 从对象池中获取一个 InsertImpl 实例
func (p *InsertImplPool) Get() *InsertImpl {
	return p.pool.Get().(*InsertImpl)
}

// Put 将 InsertImpl 实例放回对象池
func (p *InsertImplPool) Put(impl *InsertImpl) {
	// 重置实例状态
	impl.Reset()
	p.pool.Put(impl)
}

// 提供了一个统一的插入流程接口
// 添加记录流程接口
type Insert interface {
	//检查参数
	CheckParams() error
	//准备插入操作的batch
	PrepareBatch(batchs ...storage.Batch)
	//自动增值主键
	AutoIncrement() (int, error)
	// 转换字段为字节数组
	FieldsToBytes(fields *map[string]any) *map[string][]byte
	// 格式化记录
	FormatRecord(fieldsBytes *map[string][]byte) []byte
	// 添加数据到batch
	AddRecord()
	// 提交事务
	Commit() error
}

type InsertImpl struct {
	table             *Table
	batch             storage.Batch
	userProvidedBatch bool
	fields            *map[string]any
	fieldsBytes       *map[string][]byte
	record            []byte
}

// Reset 重置 InsertImpl 实例的状态
func (i *InsertImpl) Reset() {
	i.table = nil
	i.batch = nil
	i.userProvidedBatch = false
	i.fields = nil
	i.fieldsBytes = nil
	i.record = nil
	//i.records = nil
	//i.fieldsBytesList = nil
	//i.recordsList = nil
	//i.ids = nil
}

// Reset 重置 BatchInsertImpl 实例的状态
func (i *BatchInsertImpl) Reset() {
	// 重置嵌入的 InsertImpl 实例
	i.InsertImpl.Reset()
	// 重置 BatchInsertImpl 特有的字段
	i.records = nil
	i.fieldsBytesList = nil
	i.recordsList = nil
	i.ids = nil
}

// NewInsertImpl 创建一个新的 InsertImpl 实例
func NewInsertImpl(table *Table, fields *map[string]any) *InsertImpl {
	impl := GlobalInsertImplPool.Get()
	impl.table = table
	//impl.batch = batch
	//impl.userProvidedBatch = userProvidedBatch
	impl.fields = fields
	return impl
}

func (i *InsertImpl) CheckParams() error {
	// 检查参数
	if i.table.fields == nil {
		return fmt.Errorf("表 '%s' 未设置字段和类型", i.table.name)
	}
	if i.fields == nil {
		return fmt.Errorf("不能添加空记录")
	}
	return nil
}

// AutoIncrement 实现自动增值主键
func (i *InsertImpl) AutoIncrement() (int, error) {
	// 获取主键字段
	primaryFields := i.table.GetPrimaryFields()
	pklen := len(primaryFields)
	pkfield := primaryFields[0]
	if pklen == 0 {
		return -1, fmt.Errorf("表 '%s' 没有设置主键", i.table.name)
	}

	// 是否支持默认自动增值主键，单主键并且主键字段名为"id"
	supportDefault := pklen == 1 && pkfield == "id"
	currentID := -1

	if supportDefault {
		// 检查是否提供了主键字段
		//使用默认自动增值主键时，不需要提供主键字段，系统自动生成，强制使用"id"字段和自动增值主键
		_, ok := (*i.fields)[pkfield]
		if !ok { //未提供主键字段，自动生成主键值
			currentID = i.table.GetAutoInc()
			(*i.fields)[pkfield] = currentID
		} else { //提供了主键字段id，但是值为nil，自动生成主键值
			if (*i.fields)[pkfield] == nil {
				currentID = i.table.GetAutoInc()
				(*i.fields)[pkfield] = currentID
			}
		}
	}
	currentID = util.AnyToInt((*i.fields)[pkfield])
	return currentID, nil
}

// PrepareBatch 准备插入操作的batch
func (i *InsertImpl) PrepareBatch(batchs ...storage.Batch) {
	//是否用户手动控制事务
	i.userProvidedBatch = len(batchs) > 0
	if i.userProvidedBatch { //用户手动控制事务
		i.batch = batchs[0]
	} else {
		i.batch = i.table.kvStore.GetBatch()
	}
}
func (i *InsertImpl) AddRecord() {
	fieldsBytes := i.table.FieldsToBytes(i.fields)
	defer GlobalFieldsBytesPool.PutMapPointer(fieldsBytes)
	//格式化记录
	record := i.table.FormatRecord(fieldsBytes)
	//BatchContainer := NewBatchContainer(i.batch, i.table.indexs, i.table.id, i.table.kvStore)
	BatchContainer := GetBatchContainer(i.batch, i.table.indexs, i.table.id, i.table.kvStore)
	defer PutBatchContainer(BatchContainer)
	BatchContainer.SetValue(0, record)                                     //添加主键value=record
	BatchContainer.SetValue(1, i.table.GetPrimaryKey().GetID(fieldsBytes)) //添加普通索引value=GetPrimaryKey().GetID()
	//添加全文索引key=joinValue,value=nil
	BatchContainer.Operation(fieldsBytes)
}

// Commit 提交事务
func (i *InsertImpl) Commit() error {
	// 提交事务
	if !i.userProvidedBatch {
		if err := i.table.kvStore.WriteBatch(i.batch); err != nil {
			return err
		}
	}
	// 归还对象池
	defer GlobalInsertImplPool.Put(i)
	return nil
}

// 批量插入接口
type BatchInsert interface {
	Insert
	// 批量插入多条记录
	BatchInsert(records []*map[string]any, batchs ...storage.Batch) ([]int, error)
	// 批量插入多条记录，不自动生成主键
	BatchInsertNoInc(batchs ...storage.Batch) ([]int, error)
	// 批量提交事务
	BatchCommit() error
}

type BatchInsertImpl struct {
	InsertImpl
	// 批量操作相关字段
	records         []*map[string]any
	fieldsBytesList []*map[string][]byte
	recordsList     [][]byte
	ids             []int
}

// NewBatchInsertImpl 创建一个新的用于批量插入的 BatchInsertImpl 实例
func NewBatchInsertImpl(table *Table, records []*map[string]any) *BatchInsertImpl {
	impl := GlobalBatchInsertImplPool.Get()
	impl.table = table
	impl.records = records
	return impl
}

// BatchInsertInc 批量插入多条记录，自动生成主键
func (i *BatchInsertImpl) BatchInsertInc(records []*map[string]any, batchs ...storage.Batch) ([]int, error) {
	// 检查参数
	if i.table.fields == nil {
		return nil, fmt.Errorf("表 '%s' 未设置字段和类型", i.table.name)
	}
	if len(records) == 0 {
		return []int{}, nil
	}
	if records == nil {
		return nil, fmt.Errorf("records cannot be nil")
	}
	// 保存记录
	i.records = records
	// 获取主键字段
	primaryFields := i.table.GetPrimaryFields()
	if len(primaryFields) == 0 {
		return nil, fmt.Errorf("表 '%s' 没有设置主键", i.table.name)
	}
	// 是否支持默认自动增值主键，单主键并且主键字段名为"id"
	pklen := len(primaryFields)
	pkfield := primaryFields[0]
	supportDefault := pklen == 1 && pkfield == "id"
	if !supportDefault {
		return nil, fmt.Errorf("表 '%s' 没有设置默认自动增值主键", i.table.name)
	}
	// 处理批量操作
	i.batch, i.userProvidedBatch = i.table.prepareBatch(batchs...)
	// 预分配ID列表容量
	rdlen := len(records)
	ids := make([]int, rdlen)
	i.ids = ids

	// 计算需要自动生成的ID数量
	autoIncCount := rdlen
	/*
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
	*/
	// 批量获取自动增值ID，确保并发安全
	var autoIncStart int
	if autoIncCount > 0 {
		autoIncStart = i.table.GetAutoIncBatch(autoIncCount)
	}

	// 处理记录并批量插入
	autoIncIdx := 0
	// 从对象池中获取一个 batchContainer
	BatchContainer := GetBatchContainer(i.batch, i.table.indexs, i.table.id, i.table.kvStore)
	defer PutBatchContainer(BatchContainer)

	// 预分配资源列表
	i.fieldsBytesList = make([]*map[string][]byte, len(records))
	i.recordsList = make([][]byte, len(records))

	for j, fields := range records {
		// 检查字段类型
		if err := i.table.CheckType(fields); err != nil {
			return nil, err
		}

		// 处理自动增值主键
		// 使用预分配的自动增值ID
		ids[j] = autoIncStart + autoIncIdx
		(*fields)[pkfield] = ids[j]
		autoIncIdx++
		/*
			if supportDefault {
				if _, ok := (*fields)[pkfield]; !ok || (*fields)[pkfield] == nil {
					// 使用预分配的自动增值ID
					ids[j] = autoIncStart + autoIncIdx
					(*fields)[pkfield] = ids[j]
					autoIncIdx++
				} else {
					// 使用提供的主键值
					ids[j] = util.AnyToInt((*fields)[pkfield])
				}
			}  else {
				// 非默认自动增值主键，使用提供的主键值
				ids[j] = util.AnyToInt((*fields)[pkfield])
			}
		*/
		// 转换字段为字节数组
		fieldsBytes := i.table.FieldsToBytes(fields)
		// 检查fieldsBytes是否为nil
		if fieldsBytes == nil {
			return nil, fmt.Errorf("failed to convert fields to bytes")
		}
		i.fieldsBytesList[j] = fieldsBytes

		// 格式化记录
		record := i.table.FormatRecord(fieldsBytes)
		i.recordsList[j] = record

		// 批量添加记录
		BatchContainer.SetValue(0, record)
		BatchContainer.SetValue(1, i.table.GetPrimaryKey().GetID(fieldsBytes))
		BatchContainer.Operation(fieldsBytes)
	}

	// 提交批量操作
	if err := i.BatchCommit(); err != nil {
		return nil, err
	}

	return ids, nil
}

// BatchCommit 批量提交事务
func (i *BatchInsertImpl) BatchCommit() error {
	// 使用 defer 确保实例在方法结束后被放回对象池
	defer GlobalBatchInsertImplPool.Put(i)

	// 提交事务
	if !i.userProvidedBatch {
		if err := i.table.kvStore.WriteBatch(i.batch); err != nil {
			return err
		}
	}

	// 释放资源
	if i.fieldsBytesList != nil {
		for _, fieldsBytes := range i.fieldsBytesList {
			GlobalFieldsBytesPool.PutMapPointer(fieldsBytes)
		}
		i.fieldsBytesList = nil
	}

	// 重置字段
	i.records = nil
	i.recordsList = nil
	i.ids = nil
	return nil
}

// 批量添加不需要自动增值的记录，并且全部记录规则相同。
// 可用于批量插入时序数据，当表主键为时间戳时，建议使用此方法
func (i *BatchInsertImpl) BatchInsertNoInc(batchs ...storage.Batch) ([]int, error) {
	// 检查参数
	if i.table.fields == nil {
		return nil, fmt.Errorf("表 '%s' 未设置字段和类型", i.table.name)
	}
	if len(i.records) == 0 {
		return []int{}, nil
	}

	// 获取主键字段
	primaryFields := i.table.GetPrimaryFields()
	if len(primaryFields) == 0 {
		return nil, fmt.Errorf("表 '%s' 没有设置主键", i.table.name)
	}

	pkfield := primaryFields[0]

	// 处理批量操作
	i.batch, i.userProvidedBatch = i.table.prepareBatch(batchs...)
	// 预分配ID列表容量
	ids := make([]int, len(i.records))
	i.ids = ids

	// 预分配资源列表
	i.fieldsBytesList = make([]*map[string][]byte, len(i.records))
	i.recordsList = make([][]byte, len(i.records))

	// 从对象池中获取一个 batchContainer
	BatchContainer := GetBatchContainer(i.batch, i.table.indexs, i.table.id, i.table.kvStore)
	defer PutBatchContainer(BatchContainer)

	for j, fields := range i.records {
		// 检查字段类型
		if err := i.table.CheckType(fields); err != nil {
			return nil, err
		}
		if _, ok := (*fields)[pkfield]; !ok || (*fields)[pkfield] == nil {
			return nil, fmt.Errorf("primary key field '%s' not found in record", pkfield)
		} else {
			ids[j] = util.AnyToInt((*fields)[pkfield])
		}

		// 转换字段为字节数组
		fieldsBytes := i.table.FieldsToBytes(fields)
		// 检查fieldsBytes是否为nil
		if fieldsBytes == nil {
			return nil, fmt.Errorf("failed to convert fields to bytes")
		}
		i.fieldsBytesList[j] = fieldsBytes

		// 格式化记录
		record := i.table.FormatRecord(fieldsBytes)
		i.recordsList[j] = record

		// 批量添加记录
		BatchContainer.SetValue(0, record)
		BatchContainer.SetValue(1, i.table.GetPrimaryKey().GetID(fieldsBytes))
		BatchContainer.Operation(fieldsBytes)
	}

	// 提交批量操作
	if err := i.BatchCommit(); err != nil {
		return nil, err
	}
	return ids, nil
}

/*

// BatchInsertWithSizeNoInc 带批量大小控制的批量插入（不需要自动增值）
func (i *BatchInsertImpl) BatchInsertWithSizeNoInc(records []*map[string]any, batchSize int, batchs ...storage.Batch) ([]int, error) {

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

		// 创建新的BatchInsertImpl实例处理每一批
		batchImpl := NewBatchInsertImpl(i.table, batchRecords)
		batchIds, err := batchImpl.BatchInsertNoInc(batchs...)
		if err != nil {
			return nil, err
		}

		// 复制ID到结果列表
		copy(allIds[start:end], batchIds)
	}

	return allIds, nil
}


*/
