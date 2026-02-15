package engine

import (
	"fmt"
	"sync"
	"time"

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
	// 验证删除字段
	ValidateDeleteFields() (any, error)
	// 读取要删除的记录
	ReadRecordForDelete() ([]byte, error)
	// 执行删除操作
	ExecuteDeleteOperation() error
	// 提交事务
	Commit() error
}

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

type DeleteImpl struct {
	table             *Table
	batch             storage.Batch
	userProvidedBatch bool
	fields            *map[string]any
	fieldsBytes       *map[string][]byte
	key               []byte
	pkValue           any
	timeout           time.Duration
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
	d.key = nil
	d.pkValue = nil
	d.timeout = 0
	d.records = nil
}

// NewDeleteImpl 创建一个新的 DeleteImpl 实例
func NewDeleteImpl(table *Table, batch storage.Batch, userProvidedBatch bool, fields *map[string]any, timeout time.Duration) *DeleteImpl {
	impl := GlobalDeleteImplPool.Get()
	impl.table = table
	impl.batch = batch
	impl.userProvidedBatch = userProvidedBatch
	impl.fields = fields
	impl.timeout = timeout
	return impl
}

// NewBatchDeleteImpl 创建一个新的用于批量删除的 DeleteImpl 实例
func NewBatchDeleteImpl(table *Table, batch storage.Batch, userProvidedBatch bool, records []*map[string]any, timeout time.Duration) *DeleteImpl {
	impl := GlobalDeleteImplPool.Get()
	impl.table = table
	impl.batch = batch
	impl.userProvidedBatch = userProvidedBatch
	impl.records = records
	impl.timeout = timeout
	return impl
}

// ValidateDeleteFields 验证删除操作的字段
func (d *DeleteImpl) ValidateDeleteFields() (any, error) {
	// 检查是否提供了所有主键字段
	for _, field := range d.table.GetPrimaryFields() {
		if _, ok := (*d.fields)[field]; !ok {
			return nil, fmt.Errorf("必须提供主键字段 '%s'", field)
		}
	}

	// 获取主键值用于行级锁
	pkField := d.table.GetPrimaryFields()[0]
	pkValue := (*d.fields)[pkField]
	d.pkValue = pkValue

	return pkValue, nil
}

// ReadRecordForDelete 读取要删除的记录
func (d *DeleteImpl) ReadRecordForDelete() ([]byte, error) {
	//读取记录 - 直接使用 ReadByBytes 避免死锁
	fieldsBytes := d.table.FieldsToBytes(d.fields)
	defer GlobalFieldsBytesPool.Put(*fieldsBytes)

	key := d.table.GetPrimaryKey().JoinValue(fieldsBytes, d.table.id)
	record := d.table.ReadByBytes(key)
	if record == nil {
		return nil, fmt.Errorf("主键值 '%v' 的记录不存在", d.fields)
	}

	// 反序列化记录
	pk := d.table.GetPrimaryKey()
	parsedFieldsBytes, err := pk.Parse(d.table.fieldsid, record)
	if err != nil {
		return nil, err
	}

	d.key = key
	d.fieldsBytes = parsedFieldsBytes

	return key, nil
}

// ExecuteDeleteOperation 执行删除操作
func (d *DeleteImpl) ExecuteDeleteOperation() error {
	// 从对象池中获取一个 batchContainer
	BatchContainer := GetBatchContainer(d.batch, d.table.indexs, d.table.id, d.table.kvStore)
	defer PutBatchContainer(BatchContainer)

	// 对于删除操作，需要将values[0]设置为nil，这样Add方法才会执行删除操作
	BatchContainer.SetValue(0, nil)
	BatchContainer.Operation(d.fieldsBytes)

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

// BatchDelete 批量删除多条记录
func (d *DeleteImpl) BatchDelete(records []*map[string]any, params ...any) error {
	// 检查参数
	if len(records) == 0 {
		return nil
	}
	if records == nil {
		return fmt.Errorf("records cannot be nil")
	}

	// 保存记录
	d.records = records

	// 解析参数
	batch, timeout := d.table.parseDeleteParams(params...)

	// 准备batch
	batch, userProvidedBatch, err := d.table.prepareDeleteBatch(batch)
	if err != nil {
		return err
	}
	d.batch = batch
	d.userProvidedBatch = userProvidedBatch
	d.timeout = timeout

	// 处理记录并批量删除
	for _, fields := range records {
		// 创建临时 DeleteImpl 实例处理单条记录
		deleteImpl := NewDeleteImpl(d.table, batch, userProvidedBatch, fields, timeout)

		// 验证字段
		pkValue, err := deleteImpl.ValidateDeleteFields()
		if err != nil {
			return err
		}

		// 获取行级排他锁
		lockKey := fmt.Sprintf("%v", pkValue)
		if err := d.table.acquireRowWriteLock(pkValue, 0, timeout); err != nil {
			return err
		}

		// 释放行级锁
		defer func() {
			if rowLock, ok := d.table.rowLocks.Load(lockKey); ok {
				rl := rowLock.(*RowLock)
				rl.rwLock.Unlock()
			}
		}()

		// 读取记录
		_, err = deleteImpl.ReadRecordForDelete()
		if err != nil {
			return err
		}

		// 执行删除操作
		if err := deleteImpl.ExecuteDeleteOperation(); err != nil {
			return err
		}
	}

	// 提交批量操作
	if err := d.BatchCommit(); err != nil {
		return err
	}

	return nil
}

// BatchDeleteWithSize 带批量大小控制的批量删除
func (d *DeleteImpl) BatchDeleteWithSize(records []*map[string]any, batchSize int, params ...any) error {
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
		if err := d.BatchDelete(batchRecords, params...); err != nil {
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
