package engine

import (
	"fmt"
	"sync"
	"time"

	"github.com/liaoran123/sfsDb/monitor"
	"github.com/liaoran123/sfsDb/storage"
	"github.com/liaoran123/sfsDb/util"
)

// 全局事务对象池
var GlobalTableTransactionPool = &TableTransactionPool{
	pool: sync.Pool{
		New: func() interface{} {
			return &TableTransaction{}
		},
	},
}

// TableTransactionPool 事务对象池
type TableTransactionPool struct {
	pool sync.Pool
}

// Get 从池中获取事务对象
func (p *TableTransactionPool) Get() *TableTransaction {
	return p.pool.Get().(*TableTransaction)
}

// Put 将事务对象归还到池中
func (p *TableTransactionPool) Put(tx *TableTransaction) {
	// 重置事务对象
	tx.committed = false
	tx.cache = make(map[string][]byte)
	tx.lockKeyCache = make(map[string]string)
	tx.parent = nil
	tx.children = nil
	// 其他字段在使用时会被覆盖，不需要重置

	p.pool.Put(tx)
}

// 事务隔离级别常量
const (
	// ReadUncommitted 读未提交：允许读取未提交的数据，可能导致脏读、不可重复读、幻读
	ReadUncommitted = "READ_UNCOMMITTED"
	// ReadCommitted 读已提交：只能读取已提交的数据，避免脏读，但可能导致不可重复读、幻读
	ReadCommitted = "READ_COMMITTED"
	// RepeatableRead 可重复读：确保同一事务中多次读取同一数据时结果一致，避免脏读、不可重复读，但可能导致幻读
	RepeatableRead = "REPEATABLE_READ"
	// Serializable 可序列化：最高隔离级别，完全避免脏读、不可重复读、幻读，但性能最低
	Serializable = "SERIALIZABLE"
)

// 事务选项结构体
type TransactionOptions struct {
	// 隔离级别
	IsolationLevel string `json:"isolationLevel"`
	// 是否启用嵌套事务
	AllowNested bool `json:"allowNested"`
	// 事务超时时间
	Timeout time.Duration `json:"timeout"`
	// 最大重试次数
	MaxRetries int `json:"maxRetries"`
	// 初始重试延迟
	InitialRetryDelay time.Duration `json:"initialRetryDelay"`
	// 重试退避因子（sleep时间乘法因子）
	RetryBackoffFactor float64 `json:"retryBackoffFactor"`
}

// 默认事务选项
func DefaultTransactionOptions() *TransactionOptions {
	return &TransactionOptions{
		IsolationLevel:     RepeatableRead,
		AllowNested:        false,
		Timeout:            0,
		MaxRetries:         3,
		InitialRetryDelay:  10 * time.Millisecond,
		RetryBackoffFactor: 2.0,
	}
}

// Transaction 定义事务接口
type Transaction interface {
	// Insert 在事务中插入记录
	Insert(fields *map[string]any) (int, error)
	// Update 在事务中更新记录
	Update(fields *map[string]any) error
	// Delete 在事务中删除记录
	Delete(fields *map[string]any) error
	// Search 在事务中搜索记录（支持读一致性）
	Search(fields *map[string]any, ops ...util.ComparisonOperator) (*TableIter, error)
	// SearchRange 在事务中进行区间搜索（支持读一致性）
	SearchRange(funIter storage.FunIter, fieldname string, Start, Limit any) (*TableIter, error)
	// Read 在事务中读取单条记录（支持读一致性）
	Read(fields *map[string]any) ([]byte, error)
	// Commit 提交事务
	Commit() error
	// Rollback 回滚事务
	Rollback() error
	// BeginNested 创建一个嵌套事务
	BeginNested() (Transaction, error)
	// GetOptions 获取事务选项
	GetOptions() *TransactionOptions
	// GetTxID 获取事务ID
	GetTxID() uint64
}

// TableTransaction 实现Transaction接口的具体结构体
type TableTransaction struct {
	table         *Table           // 关联的表
	batch         storage.Batch    // 事务使用的batch，原子性
	committed     bool             // 是否已提交，提交成功后则是持久性。
	snapshot      storage.Snapshot // 事务使用的快照，一致性。
	originalStore storage.Store    // 原始存储，用于写操作
	// 事务内修改缓存，用于读取自己的写操作
	// key: 主键值的字符串表示，value: 记录的字节数组
	cache map[string][]byte // 事务内修改缓存 //隔离性
	// 锁键缓存，避免重复生成锁键
	lockKeyCache map[string]string // 锁键缓存
	// 事务选项
	options *TransactionOptions
	// 父事务（用于嵌套事务）
	parent *TableTransaction
	// 子事务列表
	children []*TableTransaction
	// 事务ID
	txID uint64
	// 事务开始时间
	startTime time.Time
}

// Begin 创建一个新的事务
func (t *Table) Begin() (Transaction, error) {
	return t.BeginWithOptions(DefaultTransactionOptions())
}

// BeginWithOptions 使用指定选项创建一个新的事务
func (t *Table) BeginWithOptions(options *TransactionOptions) (Transaction, error) {
	batch := t.kvStore.GetBatch()
	if batch == nil {
		return nil, fmt.Errorf("failed to create batch for transaction")
	}

	return t.BeginWithBatchAndOptions(batch, options)
}

// BeginWithBatch 创建一个使用外部传入batch的事务
// 用于实现多表事务，多个表共享同一个batch
func (t *Table) BeginWithBatch(batch storage.Batch) (Transaction, error) {
	return t.BeginWithBatchAndOptions(batch, DefaultTransactionOptions())
}

// BeginWithBatchAndOptions 使用外部传入batch和指定选项创建一个事务
func (t *Table) BeginWithBatchAndOptions(batch storage.Batch, options *TransactionOptions) (Transaction, error) {
	if batch == nil {
		return nil, fmt.Errorf("batch cannot be nil")
	}

	if options == nil {
		options = DefaultTransactionOptions()
	}

	// 为每个事务创建自己的快照实例，而不是共享表级别的快照
	var snapshot storage.Snapshot
	var err error

	// 根据隔离级别决定是否创建快照
	if options.IsolationLevel == RepeatableRead || options.IsolationLevel == Serializable {
		// 检查是否是LevelDBStore，如果是则创建快照
		if levelDBStore, ok := t.kvStore.(*storage.LevelDBStore); ok {
			// 创建一个新的快照实例
			snapshot, err = levelDBStore.Snapshot()
			if err != nil {
				return nil, fmt.Errorf("failed to create snapshot: %v", err)
			}
		}
	}

	// 生成事务ID
	txID := uint64(time.Now().UnixNano())

	// 从对象池获取事务对象
	tx := GlobalTableTransactionPool.Get()

	// 初始化事务对象
	tx.table = t
	tx.batch = batch
	tx.committed = false
	tx.snapshot = snapshot
	tx.originalStore = t.kvStore
	tx.cache = make(map[string][]byte)        // 初始化事务内缓存
	tx.lockKeyCache = make(map[string]string) // 初始化锁键缓存
	tx.options = options
	tx.txID = txID
	tx.startTime = time.Now()

	return tx, nil
}

// BeginNested 创建一个嵌套事务
func (tx *TableTransaction) BeginNested() (Transaction, error) {
	if !tx.options.AllowNested {
		return nil, fmt.Errorf("nested transactions are not allowed")
	}

	if tx.committed {
		return nil, fmt.Errorf("cannot create nested transaction on committed transaction")
	}

	// 为嵌套事务创建新的缓存，但共享同一个batch
	nestedTx := &TableTransaction{
		table:         tx.table,
		batch:         tx.batch, // 共享父事务的batch
		committed:     false,
		snapshot:      tx.snapshot, // 共享父事务的快照
		originalStore: tx.originalStore,
		cache:         make(map[string][]byte), // 新的缓存
		lockKeyCache:  make(map[string]string), // 新的锁键缓存
		options:       tx.options,
		parent:        tx,
		txID:          uint64(time.Now().UnixNano()),
		startTime:     time.Now(),
	}

	// 将嵌套事务添加到父事务的子事务列表
	tx.children = append(tx.children, nestedTx)

	return nestedTx, nil
}

// checkCommitted 检查事务是否已提交
func (tx *TableTransaction) checkCommitted() error {
	if tx.committed {
		return fmt.Errorf("transaction already committed")
	}
	return nil
}

// Insert 在事务中插入记录
func (tx *TableTransaction) Insert(fields *map[string]any) (int, error) {
	if err := tx.checkCommitted(); err != nil {
		return 0, err
	}

	// 执行插入操作
	id, err := tx.table.Insert(fields, tx.batch)
	if err != nil {
		return 0, err
	}

	// 将fields转换为fieldsBytes，用于生成主键和记录
	fieldsBytes := tx.table.FieldsToBytes(fields)

	// 生成缓存键
	cacheKey := tx.getCacheKey(fields)

	// 生成记录字节数组
	record := tx.table.FormatRecord(fieldsBytes)

	// 将记录存入缓存，用于读取自己的写操作
	tx.cache[cacheKey] = record

	// 缓存锁键
	tx.lockKeyCache[cacheKey] = tx.table.generateLockKey(fields)

	return id, nil
}

// Update 在事务中更新记录
func (tx *TableTransaction) Update(fields *map[string]any) error {
	if err := tx.checkCommitted(); err != nil {
		return err
	}

	// 生成缓存键
	cacheKey := tx.getCacheKey(fields)

	// 检查缓存中是否有要更新的记录
	if record, exists := tx.cache[cacheKey]; exists {
		// 缓存中有记录，使用提取的方法处理
		if err := tx.updateFromCache(fields, cacheKey, record); err != nil {
			return err
		}
	} else {
		// 缓存中没有记录，直接执行更新操作
		err := tx.table.Update(fields, tx.batch)
		if err != nil {
			return err
		}

		// 从缓存中删除旧记录，强制后续读取从数据库获取最新值
		delete(tx.cache, cacheKey)
	}

	// 缓存锁键
	tx.lockKeyCache[cacheKey] = tx.table.generateLockKey(fields)

	return nil
}

// updateFromCache 从缓存中更新记录
func (tx *TableTransaction) updateFromCache(fields *map[string]any, cacheKey string, record []byte) error {
	// 1. 解析记录
	pk := tx.table.GetPrimaryKey()
	fieldsBytes, err := pk.Parse(tx.table.fieldsid, record)
	if err != nil {
		return err
	}

	// 2. 准备更新字段列表
	updateFields, err := tx.table.prepareUpdateFieldsList(fields)
	if err != nil {
		return err
	}
	if updateFields == nil {
		return nil
	}
	defer PutStringSlice(updateFields)

	// 3. 执行更新操作
	if err := tx.table.executeUpdateOperation(tx.batch, fields, fieldsBytes, updateFields); err != nil {
		return err
	}

	// 4. 格式化记录
	updatedRecord := tx.table.FormatRecord(fieldsBytes)

	// 5. 更新缓存
	tx.cache[cacheKey] = updatedRecord

	return nil
}

// Delete 在事务中删除记录
func (tx *TableTransaction) Delete(fields *map[string]any) error {
	if err := tx.checkCommitted(); err != nil {
		return err
	}

	// 执行删除操作
	err := tx.table.Delete(fields, tx.batch)
	if err != nil {
		return err
	}

	// 生成缓存键
	cacheKey := tx.getCacheKey(fields)

	// 从缓存中删除记录，确保读一致性
	delete(tx.cache, cacheKey)

	// 缓存锁键
	tx.lockKeyCache[cacheKey] = tx.table.generateLockKey(fields)

	return nil
}

// getCacheKey 生成缓存键
func (tx *TableTransaction) getCacheKey(fields *map[string]any) string {
	// 将fields转换为fieldsBytes，用于生成主键
	fieldsBytes := tx.table.FieldsToBytes(fields)
	defer func() {
		if fieldsBytes != nil && *fieldsBytes != nil {
			GlobalFieldsBytesPool.Put(*fieldsBytes)
		}
	}()
	// 生成主键键值，用于缓存
	pkKey := tx.table.GetPrimaryKey().JoinValue(fieldsBytes, tx.table.id)
	return string(pkKey)
}

// Read 在事务中读取单条记录（支持读一致性）
func (tx *TableTransaction) Read(fields *map[string]any) ([]byte, error) {
	if err := tx.checkCommitted(); err != nil {
		return nil, err
	}

	// 生成缓存键
	cacheKey := tx.getCacheKey(fields)

	// 1. 优先从缓存中读取，支持读取自己的写操作
	if record, exists := tx.cache[cacheKey]; exists {
		return record, nil
	}

	// 2. 缓存中没有，使用事务自己的快照或原始存储读取
	// 生成主键键值
	fieldsBytes := tx.table.FieldsToBytes(fields)
	defer func() {
		if fieldsBytes != nil && *fieldsBytes != nil {
			GlobalFieldsBytesPool.Put(*fieldsBytes)
		}
	}()
	pkKey := tx.table.GetPrimaryKey().JoinValue(fieldsBytes, tx.table.id)

	// 如果有快照，使用快照读取；否则使用原始存储
	if tx.snapshot != nil {
		return tx.snapshot.Get(pkKey)
	}
	return tx.originalStore.Get(pkKey)
}

// Search 在事务中搜索记录（支持读一致性，即使用快照）
func (tx *TableTransaction) Search(fields *map[string]any, ops ...util.ComparisonOperator) (*TableIter, error) {
	// 检查事务是否已提交
	if err := tx.checkCommitted(); err != nil {
		return nil, err
	}

	// 检查fields参数是否为nil
	if fields == nil {
		return nil, fmt.Errorf("fields cannot be nil")
	}

	// 创建一个函数，根据是否有快照选择不同的存储获取迭代器
	funIter := func(start, limit []byte) storage.Iterator {
		if tx.snapshot != nil {
			return tx.snapshot.Iterator(start, limit)
		} else {
			return tx.originalStore.Iterator(start, limit)
		}
	}

	// 调用table.Searchs方法，传入funIter函数
	tbiter, err := tx.table.Searchs(funIter, fields, ops...)
	return tbiter, err
}

// SearchRange 在事务中进行区间搜索（支持读一致性，即使用快照）
func (tx *TableTransaction) SearchRange(funIter storage.FunIter, fieldname string, Start, Limit any) (*TableIter, error) {
	// 检查事务是否已提交
	if err := tx.checkCommitted(); err != nil {
		return nil, err
	}

	// 创建一个函数，根据是否有快照选择不同的存储获取迭代器
	transactionFunIter := func(start, limit []byte) storage.Iterator {
		if tx.snapshot != nil {
			return tx.snapshot.Iterator(start, limit)
		} else {
			return tx.originalStore.Iterator(start, limit)
		}
	}

	// 调用table.SearchRange方法，传入事务的funIter函数
	return tx.table.SearchRange(transactionFunIter, fieldname, Start, Limit)
}

// GetOptions 获取事务选项
func (tx *TableTransaction) GetOptions() *TransactionOptions {
	return tx.options
}

// GetTxID 获取事务ID
func (tx *TableTransaction) GetTxID() uint64 {
	return tx.txID
}

// Commit 提交事务
func (tx *TableTransaction) Commit() error {
	if err := tx.checkCommitted(); err != nil {
		return err
	}

	endTime := time.Now()                 // 记录事务结束时间
	duration := endTime.Sub(tx.startTime) // 计算事务用时
	monitor.GTransactionStatsMap.SetTimeAsync(tx.txID, duration, tx.table.name, tx.options.IsolationLevel, true)

	// 提交所有子事务
	for _, child := range tx.children {
		if !child.committed {
			if err := child.Commit(); err != nil {
				return err
			}
		}
	}

	// 只有根事务才真正提交batch
	if tx.parent == nil {
		// 1. 提交批量操作
		// 使用原始存储执行写操作，支持重试
		err := tx.executeWithRetry()
		if err != nil {
			// 提交失败，释放快照资源
			if tx.snapshot != nil {
				tx.snapshot.Release()
			}
			return err
		}

		// 2. 释放快照资源
		if tx.snapshot != nil {
			tx.snapshot.Release()
		}
	}

	// 3. 标记事务已结束，清空缓存
	tx.committed = true
	tx.cache = nil
	tx.lockKeyCache = nil
	tx.children = nil

	// 4. 归还事务对象到池中
	if tx.parent == nil {
		// 只有根事务才归还到池，子事务由父事务管理
		GlobalTableTransactionPool.Put(tx)
	}

	return nil
}

// executeWithRetry 执行批量操作，支持重试
func (tx *TableTransaction) executeWithRetry() error {
	maxRetries := tx.options.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}

	initialDelay := tx.options.InitialRetryDelay
	if initialDelay <= 0 {
		initialDelay = 10 * time.Millisecond
	}

	backoffFactor := tx.options.RetryBackoffFactor
	if backoffFactor < 1.0 {
		backoffFactor = 2.0
	}

	// 尝试执行，最多重试maxRetries次
	for attempt := 0; attempt <= maxRetries; attempt++ {
		// 执行批量操作
		err := tx.originalStore.WriteBatch(tx.batch)
		if err == nil {
			// 执行成功
			return nil
		}

		// 检查是否是可重试的错误
		if !tx.isRetryableError(err) {
			// 不可重试的错误，直接返回
			return err
		}

		// 检查是否达到最大重试次数
		if attempt >= maxRetries {
			// 达到最大重试次数，返回最后一次错误
			return fmt.Errorf("failed after %d retries: %w", maxRetries, err)
		}

		// 计算重试延迟（指数退避）
		delay := initialDelay
		for i := 0; i < attempt; i++ {
			delay = time.Duration(float64(delay) * backoffFactor)
		}

		// 等待后重试
		time.Sleep(delay)
	}

	return nil
}

// isRetryableError 判断错误是否可重试
func (tx *TableTransaction) isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	// 这里可以根据具体的错误类型判断是否可重试
	// 例如：锁冲突、临时网络问题等
	// 对于LevelDB，常见的可重试错误包括：
	// - 锁冲突
	// - 临时的I/O错误

	// 暂时默认所有错误都可重试，实际应用中需要根据具体错误类型判断
	// 后续可以根据storage包中的错误类型进行更精确的判断
	return true
}

// Rollback 回滚事务
// 注意：LevelDB的WriteBatch不支持真正的回滚
// 此方法仅标记事务已结束，防止重复提交，并释放快照资源
func (tx *TableTransaction) Rollback() error {
	if err := tx.checkCommitted(); err != nil {
		return err
	}

	endTime := time.Now()                 // 记录事务结束时间
	duration := endTime.Sub(tx.startTime) // 计算事务用时
	monitor.GTransactionStatsMap.SetTimeAsync(tx.txID, duration, tx.table.name, tx.options.IsolationLevel, false)

	// 回滚所有子事务
	for _, child := range tx.children {
		if !child.committed {
			if err := child.Rollback(); err != nil {
				return err
			}
		}
	}

	// 只有根事务才释放快照资源
	if tx.parent == nil {
		// 1. 释放快照资源
		if tx.snapshot != nil {
			tx.snapshot.Release()
		}
	}

	// 2. 标记事务已结束，清空缓存
	tx.committed = true
	tx.cache = nil
	tx.lockKeyCache = nil
	tx.children = nil

	// 3. 归还事务对象到池中
	if tx.parent == nil {
		// 只有根事务才归还到池，子事务由父事务管理
		GlobalTableTransactionPool.Put(tx)
	}

	return nil
}
