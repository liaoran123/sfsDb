package engine

import (
	"fmt"
	"sync"

	"github.com/liaoran123/sfsDb/storage"
	"github.com/liaoran123/sfsDb/util"
)

// UpdateImplPool 是 UpdateImpl 的对象池
type UpdateImplPool struct {
	pool sync.Pool
}

// 全局 UpdateImpl 对象池
var GlobalUpdateImplPool = &UpdateImplPool{
	pool: sync.Pool{
		New: func() interface{} {
			return &UpdateImpl{}
		},
	},
}

// Get 从对象池中获取一个 UpdateImpl 实例
func (p *UpdateImplPool) Get() *UpdateImpl {
	return p.pool.Get().(*UpdateImpl)
}

// Put 将 UpdateImpl 实例放回对象池
func (p *UpdateImplPool) Put(impl *UpdateImpl) {
	// 重置实例状态
	impl.Reset()
	p.pool.Put(impl)
}

// 提供了一个统一的更新流程接口
type Update interface {
	//检查参数
	CheckParams() error
	//准备更新操作的batch
	PrepareBatch(batchs ...storage.Batch)
	//读取记录
	ReadRecord() error
	// 准备更新字段列表
	PrepareUpdateFields() error
	// 获取更新字段的字节表示
	GetFieldsBytes() error
	//删除旧记录
	DeleteOldRecord() error
	//添加新记录
	AddNewRecord() error
	// 提交事务
	Commit() error
}

type UpdateImpl struct {
	table             *Table
	batch             storage.Batch
	userProvidedBatch bool
	fields            *map[string]any
	fieldsBytes       *map[string][]byte
	updateFields      []string
	record            []byte
}

// Reset 重置 UpdateImpl 实例的状态
func (u *UpdateImpl) Reset() {
	u.table = nil
	u.batch = nil
	u.userProvidedBatch = false
	u.fields = nil
	u.fieldsBytes = nil
	u.updateFields = nil
	u.record = nil
	//u.records = nil
}

// NewUpdateImpl 创建一个新的 UpdateImpl 实例
func NewUpdateImpl(table *Table, fields *map[string]any) *UpdateImpl {
	impl := GlobalUpdateImplPool.Get()
	impl.table = table
	//impl.batch = batch
	//impl.userProvidedBatch = userProvidedBatch
	impl.fields = fields
	return impl
}

// CheckParams 检查参数是否正确
func (u *UpdateImpl) CheckParams() error {
	// 检查是否提供了所有主键字段
	for _, field := range u.table.GetPrimaryFields() {
		if _, ok := (*u.fields)[field]; !ok {
			return fmt.Errorf("必须提供主键字段 '%s'", u.fields)
		}
	}
	// 检查字段类型是否匹配
	if err := u.table.CheckType(u.fields); err != nil {
		return err
	}
	return nil
}

// PrepareBatch 准备更新操作的batch
func (u *UpdateImpl) PrepareBatch(batchs ...storage.Batch) {
	//是否用户手动控制事务
	u.batch, u.userProvidedBatch = u.table.prepareBatch(batchs...)
}

// ReadRecord 读取记录
func (u *UpdateImpl) ReadRecord() error {
	var err error
	u.record, err = u.table.Read(u.fields)
	if err != nil {
		return err
	}
	if u.record == nil {
		return fmt.Errorf("主键值 '%v' 的记录不存在", u.fields)
	}
	return nil
}

// PrepareUpdateFields 准备更新字段列表
func (u *UpdateImpl) PrepareUpdateFields() error {
	u.updateFields = GetStringSlice()
	for field := range *u.fields {
		//排除主键字段
		if u.table.GetPrimaryKey().MatchFields(field) {
			continue
		}
		u.updateFields = append(u.updateFields, field)
	}
	//检查更新字段个数是否为0
	if len(u.updateFields) == 0 {
		return fmt.Errorf("更新字段不能为空")
	}
	return nil
}

// GetFieldsBytes 获取更新字段的字节表示
func (u *UpdateImpl) GetFieldsBytes() error {
	pk := u.table.GetPrimaryKey()
	var err error
	u.fieldsBytes, err = pk.Parse(u.table.fieldsid, u.record)
	if err != nil {
		return err
	}
	return nil
}

// DeleteOldRecord 删除旧记录
func (u *UpdateImpl) DeleteOldRecord() {
	batchContainer := GetBatchContainer(u.batch, u.table.indexs, u.table.id, u.table.kvStore)
	defer PutBatchContainer(batchContainer)
	batchContainer.Operation(u.fieldsBytes, u.updateFields...)
}
func (u *UpdateImpl) AddNewRecord() {
	//更新字段值
	for field, val := range *u.fields {
		//排除主键字段，主键字段不能更新
		if u.table.GetPrimaryKey().MatchFields(field) {
			continue
		}
		if _, ok := u.table.fields[field]; ok {
			(*u.fieldsBytes)[field] = util.AnyToBytes(val)
		}
	}
	batchContainer := GetBatchContainer(u.batch, u.table.indexs, u.table.id, u.table.kvStore)
	defer PutBatchContainer(batchContainer)
	//设置新值添加
	record := u.table.FormatRecord(u.fieldsBytes)
	batchContainer.SetValue(0, record)                                       //添加主键value=record
	batchContainer.SetValue(1, u.table.GetPrimaryKey().GetID(u.fieldsBytes)) //添加普通索引value=GetPrimaryKey().GetID()
	//添加全文索引key=joinValue,value=nil
	batchContainer.Operation(u.fieldsBytes, u.updateFields...)
}

// Commit 提交事务
func (u *UpdateImpl) Commit() error {
	// 提交事务
	if !u.userProvidedBatch {
		if err := u.table.kvStore.WriteBatch(u.batch); err != nil {
			return err
		}
	}
	// 释放资源
	PutStringSlice(u.updateFields)
	// 释放 fieldsBytes 到对象池
	GlobalFieldsBytesPool.Put(*u.fieldsBytes)
	// 归还对象池
	defer GlobalUpdateImplPool.Put(u)
	return nil
}

/*
// -----以下功能函数，并无实质用处，以防万一，保留一下。-------------------------------------------------

// 批量更新接口
type BatchUpdate interface {
	Update
	// 批量更新多条记录
	BatchUpdate(records []*map[string]any, params ...any) error
	// 带批量大小控制的批量更新
	BatchUpdateWithSize(records []*map[string]any, batchSize int, params ...any) error
	// 批量提交事务
	BatchCommit() error
}

type BatchUpdateImpl struct {
	UpdateImpl
	records []*map[string]any
}

// NewBatchUpdateImpl 创建一个新的用于批量更新的 BatchUpdateImpl 实例
func NewBatchUpdateImpl(table *Table, records []*map[string]any) *BatchUpdateImpl {
	impl := GlobalUpdateImplPool.Get()
	impl.table = table
	return &BatchUpdateImpl{
		UpdateImpl: *impl,
		records:    records,
	}
}

// BatchUpdate 批量更新多条记录
func (u *BatchUpdateImpl) BatchUpdate(batchs ...storage.Batch) error {
	// 检查参数
	if len(u.records) == 0 {
		return nil
	}

	// 准备batch
	batch, userProvidedBatch := u.table.prepareBatch(batchs...)

	u.batch = batch
	u.userProvidedBatch = userProvidedBatch

	// 处理记录并批量更新
	for _, fields := range u.records {
		// 创建临时 UpdateImpl 实例处理单条记录
		updateImpl := NewUpdateImpl(u.table, fields)
		defer GlobalUpdateImplPool.Put(updateImpl)

		// 检查参数
		if err := updateImpl.CheckParams(); err != nil {
			return err
		}

		// 准备batch（使用同一个batch）
		updateImpl.batch = batch
		updateImpl.userProvidedBatch = userProvidedBatch

		// 读取记录
		if err := updateImpl.ReadRecord(); err != nil {
			return err
		}

		// 准备更新字段列表
		if err := updateImpl.PrepareUpdateFields(); err != nil {
			return err
		}

		// 获取字段字节
		if err := updateImpl.GetFieldsBytes(); err != nil {
			return err
		}

		// 删除旧记录
		updateImpl.DeleteOldRecord()

		// 添加新记录
		updateImpl.AddNewRecord()

		// 释放资源
		if updateImpl.fieldsBytes != nil {
			GlobalFieldsBytesPool.Put(*updateImpl.fieldsBytes)
		}
		if updateImpl.updateFields != nil {
			PutStringSlice(updateImpl.updateFields)
		}

		// 注意：这里不调用 updateImpl.Commit()，因为我们要批量提交
	}

	// 提交批量操作
	if err := u.BatchCommit(); err != nil {
		return err
	}

	return nil
}

// BatchUpdateWithSize 带批量大小控制的批量更新
func (u *BatchUpdateImpl) BatchUpdateWithSize(batchSize int, batchs ...storage.Batch) error {
	// 检查参数
	if batchSize <= 0 {
		batchSize = 100 // 默认批量大小
	}

	// 计算总批次
	totalRecords := len(u.records)
	if totalRecords == 0 {
		return nil
	}

	// 分批处理
	for start := 0; start < totalRecords; start += batchSize {
		end := start + batchSize
		if end > totalRecords {
			end = totalRecords
		}

		// 创建临时 BatchUpdateImpl 处理当前批次
		batchRecords := u.records[start:end]
		batchUpdateImpl := NewBatchUpdateImpl(u.table, batchRecords)
		if err := batchUpdateImpl.BatchUpdate(batchs...); err != nil {
			return err
		}
	}

	return nil
}

// BatchCommit 批量提交事务
func (u *BatchUpdateImpl) BatchCommit() error {
	// 提交事务
	if !u.userProvidedBatch {
		if err := u.table.kvStore.WriteBatch(u.batch); err != nil {
			return err
		}
	}

	// 归还对象池
	defer GlobalUpdateImplPool.Put(&u.UpdateImpl)

	return nil
}
*/
