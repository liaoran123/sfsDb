package engine

import (
	"fmt"
	"sync"

	"github.com/liaoran123/sfsDb/storage"
)

// DeleteImplPool 是 DeleteImpl 的对象池
type DeleteImplPool struct {
	pool sync.Pool
}

// 全局 DeleteImpl 对象池
var GlobalDeleteImplPool = &DeleteImplPool{
	pool: sync.Pool{
		New: func() interface{} {
			return &DeleteImpl{}
		},
	},
}

// Get 从对象池中获取一个 DeleteImpl 实例
func (p *DeleteImplPool) Get() *DeleteImpl {
	return p.pool.Get().(*DeleteImpl)
}

// Put 将 DeleteImpl 实例放回对象池
func (p *DeleteImplPool) Put(impl *DeleteImpl) {
	// 重置实例状态
	impl.Reset()
	p.pool.Put(impl)
}

// 提供了一个统一的删除流程接口
type Delete interface {
	// 是否提供了主键字段
	HasPrimaryKey() error
	//准备删除操作的batch
	PrepareBatch(batchs ...storage.Batch)
	// 读取要删除的记录
	ReadRecord() error
	// 删除记录
	DeleteRecord() error
	// 提交事务
	Commit() error
}

type DeleteImpl struct {
	table             *Table
	batch             storage.Batch
	userProvidedBatch bool
	fields            *map[string]any
	fieldsBytes       *map[string][]byte
	record            []byte
	pkValue           any
	// 批量操作相关字段
	records []*map[string]any
}

// Reset 重置 DeleteImpl 实例的状态
func (d *DeleteImpl) Reset() {
	d.table = nil
	d.batch = nil
	d.userProvidedBatch = false
	d.fields = nil
	d.fieldsBytes = nil
	d.record = nil
	d.pkValue = nil
	d.records = nil
}

// NewDeleteImpl 创建一个新的 DeleteImpl 实例
func NewDeleteImpl(table *Table, fields *map[string]any) *DeleteImpl {
	impl := GlobalDeleteImplPool.Get()
	impl.table = table
	impl.fields = fields
	return impl
}
func (d *DeleteImpl) HasPrimaryKey() error {
	// 检查是否提供了所有主键字段
	for _, field := range d.table.GetPrimaryFields() {
		if _, ok := (*d.fields)[field]; !ok {
			return fmt.Errorf("必须提供主键字段 '%s'", field)
		}
	}
	return nil
}
func (d *DeleteImpl) PrepareBatch(batchs ...storage.Batch) {
	// 准备batch
	d.batch, d.userProvidedBatch = d.table.prepareBatch(batchs...)
}
func (d *DeleteImpl) ReadRecord() error {
	record, err := d.table.Read(d.fields)
	if err != nil {
		return err
	}
	if record == nil {
		return fmt.Errorf("主键值 '%v' 的记录不存在", d.fields)
	}
	d.record = record
	return nil
}
func (d *DeleteImpl) DeleteRecord() error {
	pk := d.table.GetPrimaryKey()
	fieldsBytes, err := pk.Parse(d.table.fieldsid, d.record)
	defer GlobalFieldsBytesPool.Put(*fieldsBytes)
	if err != nil {
		//释放batch资源
		return err
	}
	BatchContainer := GetBatchContainer(d.batch, d.table.indexs, d.table.id, d.table.kvStore)
	defer PutBatchContainer(BatchContainer)
	BatchContainer.Operation(fieldsBytes)
	return nil
}

// Commit 提交事务
func (d *DeleteImpl) Commit() error {
	// 提交事务
	if !d.userProvidedBatch {
		if err := d.table.kvStore.WriteBatch(d.batch); err != nil {
			return err
		}
	}
	// 归还对象池
	defer GlobalDeleteImplPool.Put(d)

	return nil
}

/*
// -------以下的函数功能可能是无用的，以防万一，保留一下。--------------------------------------------------------
// 批量删除接口
type BatchDelete interface {
	Delete
	// 批量删除多条记录
	BatchDelete(records []*map[string]any, params ...any) error
	// 带批量大小控制的批量删除
	BatchDeleteWithSize(records []*map[string]any, batchSize int, params ...any) error
	// 批量提交事务
	BatchCommit() error
}

// NewBatchDeleteImpl 创建一个新的用于批量删除的 DeleteImpl 实例
func NewBatchDeleteImpl(table *Table, batch storage.Batch, userProvidedBatch bool, records []*map[string]any) *DeleteImpl {
	impl := GlobalDeleteImplPool.Get()
	impl.table = table
	impl.batch = batch
	impl.userProvidedBatch = userProvidedBatch
	impl.records = records
	return impl
}

// BatchDelete 批量删除多条记录
func (d *DeleteImpl) BatchDelete(records []*map[string]any, batchs ...storage.Batch) error {
	// 检查参数
	if len(records) == 0 {
		return nil
	}
	if records == nil {
		return fmt.Errorf("records cannot be nil")
	}

	// 保存记录
	d.records = records

	// 准备batch
	d.batch, d.userProvidedBatch = d.table.prepareBatch(batchs...)

	// 处理记录并批量删除
	for _, fields := range records {
		// 创建临时 DeleteImpl 实例处理单条记录
		deleteImpl := NewDeleteImpl(d.table, fields)
		// 设置与主实例相同的batch
		deleteImpl.batch = d.batch
		deleteImpl.userProvidedBatch = d.userProvidedBatch

		// 检查主键
		if err := deleteImpl.HasPrimaryKey(); err != nil {
			GlobalDeleteImplPool.Put(deleteImpl)
			return err
		}

		// 读取记录
		if err := deleteImpl.ReadRecord(); err != nil {
			GlobalDeleteImplPool.Put(deleteImpl)
			return err
		}

		// 执行删除操作
		if err := deleteImpl.DeleteRecord(); err != nil {
			GlobalDeleteImplPool.Put(deleteImpl)
			return err
		}
		// 释放临时实例回对象池
		GlobalDeleteImplPool.Put(deleteImpl)
	}

	// 提交批量操作
	if err := d.BatchCommit(); err != nil {
		return err
	}

	return nil
}

// BatchDeleteWithSize 带批量大小控制的批量删除
func (d *DeleteImpl) BatchDeleteWithSize(records []*map[string]any, batchSize int, batchs ...storage.Batch) error {
	// 检查参数
	if batchSize <= 0 {
		batchSize = 100 // 默认批量大小
	}

	// 计算总批次
	totalRecords := len(records)
	if totalRecords == 0 {
		return nil
	}

	// 分批处理
	for start := 0; start < totalRecords; start += batchSize {
		end := start + batchSize
		if end > totalRecords {
			end = totalRecords
		}

		// 处理当前批次
		batchRecords := records[start:end]
		if err := d.BatchDelete(batchRecords, batchs...); err != nil {
			return err
		}
	}

	return nil
}

// BatchCommit 批量提交事务
func (d *DeleteImpl) BatchCommit() error {
	// 提交事务
	if !d.userProvidedBatch {
		if err := d.table.kvStore.WriteBatch(d.batch); err != nil {
			return err
		}
	}

	// 归还对象池
	defer GlobalDeleteImplPool.Put(d)

	return nil
}
*/
