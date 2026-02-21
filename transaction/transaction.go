package transaction

import (
	"fmt"
	"sync"
	"time"

	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/monitor"
	"github.com/liaoran123/sfsDb/storage"
	"github.com/liaoran123/sfsDb/util"

	"github.com/liaoran123/sfsDb/transaction/batch"
	"github.com/liaoran123/sfsDb/transaction/lock"
)

// 全局事务对象池
var GlobalTransactionPool = &TransactionPool{
	pool: sync.Pool{
		New: func() interface{} {
			return &SfsTransaction{}
		},
	},
}

// 全局表事务对象池
var GlobalTableTransactionPool = &TableTransactionPool{
	Pool: sync.Pool{
		New: func() interface{} {
			return &TableTransaction{}
		},
	},
}

// 全局锁管理器
var GlobalLockManager = lock.NewShardedManager(16)

// 全局批量操作优化器
var GlobalBatchOptimizer = batch.NewOptimizer()

// TransactionPool 事务对象池
type TransactionPool struct {
	pool sync.Pool
}

// Get 从池中获取事务对象
func (p *TransactionPool) Get() *SfsTransaction {
	return p.pool.Get().(*SfsTransaction)
}

// Put 将事务对象归还到池中
func (p *TransactionPool) Put(tx *SfsTransaction) {
	// 重置事务对象
	tx.committed = false
	tx.cache = make(map[string][]byte)
	tx.parent = nil
	tx.children = nil
	// 其他字段在使用时会被覆盖，不需要重置

	p.pool.Put(tx)
}

// TableTransactionPool 表事务对象池
type TableTransactionPool struct {
	Pool sync.Pool
}

// Get 从池中获取表事务对象
func (p *TableTransactionPool) Get() *TableTransaction {
	return p.Pool.Get().(*TableTransaction)
}

// Put 将表事务对象归还到池中
func (p *TableTransactionPool) Put(tx *TableTransaction) {
	// 重置事务对象
	tx.Committed = false
	tx.Cache = make(map[string][]byte)
	tx.LockKeyCache = make(map[string]string)
	tx.Parent = nil
	tx.Children = nil
	tx.OCCSupport = NewOCCSupport() // 重置乐观并发控制支持
	// 其他字段在使用时会被覆盖，不需要重置

	p.Pool.Put(tx)
}

// Transaction 定义事务接口
type Transaction interface {
	// Get 从事务中获取值
	Get(key []byte) ([]byte, error)
	// Put 向事务中设置值
	Put(key []byte, value []byte)
	// Delete 从事务中删除值
	Delete(key []byte)
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

// TableTransactionInterface 定义表事务接口
type TableTransactionInterface interface {
	// Insert 在事务中插入记录
	Insert(fields *map[string]interface{}) (int, error)
	// Update 在事务中更新记录
	Update(fields *map[string]interface{}) error
	// Delete 在事务中删除记录
	Delete(fields *map[string]interface{}) error
	// Read 在事务中读取单条记录（支持读一致性）
	Read(fields *map[string]any) ([]byte, error)
	// Search 在事务中搜索记录（支持读一致性）
	Search(fields *map[string]any, ops ...util.ComparisonOperator) (*engine.TableIter, error)
	// Searchs 在事务中搜索记录（通过funIter支持原数据库或快照查询）
	Searchs(funIter storage.FunIter, fields *map[string]any, ops ...util.ComparisonOperator) (*engine.TableIter, error)
	// SearchRange 在事务中进行区间搜索（支持读一致性）
	SearchRange(funIter storage.FunIter, Start, Limit *map[string]any) (*engine.TableIter, error)
	// Commit 提交事务
	Commit() error
	// Rollback 回滚事务
	Rollback() error
	// BeginNested 创建一个嵌套事务
	BeginNested() (TableTransactionInterface, error)
	// GetOptions 获取事务选项
	GetOptions() *TransactionOptions
	// GetTxID 获取事务ID
	GetTxID() uint64
}

// SfsTransaction 基于sfsdb的事务实现
type SfsTransaction struct {
	store     storage.Store
	batch     storage.Batch
	snapshot  storage.Snapshot
	committed bool
	cache     map[string][]byte // 事务内修改缓存
	options   *TransactionOptions
	parent    *SfsTransaction
	children  []*SfsTransaction
	txID      uint64
	startTime time.Time
}

// TableTransaction 基于sfsdb的表事务实现
type TableTransaction struct {
	Table          *engine.Table        // 关联的表
	Batch          storage.Batch        // 事务使用的batch，原子性
	Committed      bool                 // 是否已提交，提交成功后则是持久性。
	Snapshot       storage.Snapshot     // 事务使用的快照，一致性。
	OriginalStore  storage.Store        // 原始存储，用于写操作
	Cache          map[string][]byte    // 事务内修改缓存，用于读取自己的写操作
	LockKeyCache   map[string]string    // 锁键缓存，避免重复生成锁键
	Options        *TransactionOptions  // 事务选项
	Parent         *TableTransaction    // 父事务（用于嵌套事务）
	Children       []*TableTransaction  // 子事务列表
	TxID           uint64               // 事务ID
	StartTime      time.Time            // 事务开始时间
	LockManager    *lock.ShardedManager // 锁管理器
	BatchOptimizer *batch.Optimizer     // 批量操作优化器
	OCCSupport     *OCCSupport          // 乐观并发控制支持
}

// NewTransaction 创建新的事务
func NewTransaction(store storage.Store) (*SfsTransaction, error) {
	return NewTransactionWithOptions(store, DefaultTransactionOptions())
}

// NewTransactionWithOptions 使用指定选项创建新的事务
func NewTransactionWithOptions(store storage.Store, options *TransactionOptions) (*SfsTransaction, error) {
	// 创建批量操作对象
	batch := store.GetBatch()
	if batch == nil {
		return nil, fmt.Errorf("failed to create batch")
	}

	// 为每个事务创建自己的快照实例
	var snapshot storage.Snapshot
	var err error

	// 根据隔离级别决定是否创建快照
	if options.IsolationLevel == RepeatableRead || options.IsolationLevel == Serializable {
		// 创建快照，用于读取一致性
		snapshot, err = store.Snapshot()
		if err != nil {
			return nil, fmt.Errorf("failed to create snapshot: %v", err)
		}
	}

	// 生成事务ID
	txID := uint64(time.Now().UnixNano())

	// 从对象池获取事务对象
	tx := GlobalTransactionPool.Get()

	// 初始化事务对象
	tx.store = store
	tx.batch = batch
	tx.snapshot = snapshot
	tx.committed = false
	tx.cache = make(map[string][]byte) // 初始化事务内缓存
	tx.options = options
	tx.txID = txID
	tx.startTime = time.Now()

	return tx, nil
}

// NewTableTransaction 为指定表创建事务
func NewTableTransaction(table *engine.Table) (*TableTransaction, error) {
	options := DefaultTransactionOptions()
	return NewTableTransactionWithOptions(table, options)
}

// NewTableTransactionWithOptions 为指定表创建带选项的事务
func NewTableTransactionWithOptions(table *engine.Table, options *TransactionOptions) (*TableTransaction, error) {
	// 创建批量操作对象
	// 由于table.kvStore是未导出的，我们需要通过其他方式获取batch
	// 这里我们使用一个临时的batch来获取存储实例
	tempBatch := storage.GetDBManager().GetDB().GetBatch()
	if tempBatch == nil {
		return nil, fmt.Errorf("failed to create batch")
	}

	return NewTableTransactionWithBatchAndOptions(table, tempBatch, options)
}

// NewTableTransactionWithBatch 创建一个使用外部传入batch的表事务
// 用于实现多表事务，多个表共享同一个batch
func NewTableTransactionWithBatch(table *engine.Table, batch storage.Batch) (*TableTransaction, error) {
	options := DefaultTransactionOptions()
	return NewTableTransactionWithBatchAndOptions(table, batch, options)
}

// NewTableTransactionWithBatchAndOptions 使用外部传入batch和指定选项创建一个表事务
func NewTableTransactionWithBatchAndOptions(table *engine.Table, batch storage.Batch, options *TransactionOptions) (*TableTransaction, error) {
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
	// 隔离级别实现说明：
	// 1. ReadUncommitted：不创建快照，直接读取原始存储，可能导致脏读、不可重复读、幻读
	// 2. ReadCommitted：每次读取都创建新的快照，避免脏读，但可能导致不可重复读、幻读
	// 3. RepeatableRead：事务开始时创建快照，整个事务使用同一个快照，避免脏读、不可重复读，但可能导致幻读
	// 4. Serializable：与RepeatableRead类似，但使用更严格的快照管理，尝试避免所有并发问题
	switch options.IsolationLevel {
	case RepeatableRead, Serializable:
		// 对于RepeatableRead和Serializable，事务开始时创建快照
		dbMgr := storage.GetDBManager()
		db := dbMgr.GetDB()
		if levelDBStore, ok := db.(*storage.LevelDBStore); ok {
			// 创建一个新的快照实例
			snapshot, err = levelDBStore.Snapshot()
			if err != nil {
				return nil, fmt.Errorf("failed to create snapshot: %v", err)
			}
		}
	case ReadCommitted:
		// 对于ReadCommitted，暂时不创建快照，每次读取时会直接使用原始存储
		// 注意：严格来说，ReadCommitted应该每次读取都创建新的快照
		// 但为了简化实现，当前版本使用直接读取原始存储的方式
	case ReadUncommitted:
		// 对于ReadUncommitted，不创建快照，直接读取原始存储
		// 允许读取未提交的数据
	}

	// 生成事务ID
	txID := uint64(time.Now().UnixNano())

	// 从对象池获取事务对象
	tx := GlobalTableTransactionPool.Get()

	// 初始化事务对象
	tx.Table = table
	tx.Batch = batch
	tx.Committed = false
	tx.Snapshot = snapshot
	tx.OriginalStore = storage.GetDBManager().GetDB()
	tx.Cache = make(map[string][]byte)        // 初始化事务内缓存
	tx.LockKeyCache = make(map[string]string) // 初始化锁键缓存
	tx.Options = options
	tx.TxID = txID
	tx.StartTime = time.Now()
	tx.LockManager = GlobalLockManager       // 使用全局锁管理器
	tx.BatchOptimizer = GlobalBatchOptimizer // 使用全局批量操作优化器
	tx.OCCSupport = NewOCCSupport()          // 初始化乐观并发控制支持

	return tx, nil
}

// Get 从事务中获取值
// 优先从缓存中查找，然后从批量操作中查找，最后从快照或存储中查找
func (tx *SfsTransaction) Get(key []byte) ([]byte, error) {
	if tx.committed {
		return nil, fmt.Errorf("transaction already committed")
	}

	// 1. 优先从缓存中读取，支持读取自己的写操作
	cacheKey := string(key)
	if value, exists := tx.cache[cacheKey]; exists {
		return value, nil
	}

	// 2. 缓存中没有，使用事务自己的快照或原始存储读取
	if tx.snapshot != nil {
		return tx.snapshot.Get(key)
	}
	return tx.store.Get(key)
}

// Put 向事务中设置值
func (tx *SfsTransaction) Put(key []byte, value []byte) {
	if !tx.committed {
		tx.batch.Put(key, value)
		// 将值存入缓存，用于读取自己的写操作
		tx.cache[string(key)] = value
	}
}

// Delete 从事务中删除值
func (tx *SfsTransaction) Delete(key []byte) {
	if !tx.committed {
		tx.batch.Delete(key)
		// 从缓存中删除记录，确保读一致性
		delete(tx.cache, string(key))
	}
}

// BeginNested 创建一个嵌套事务
func (tx *SfsTransaction) BeginNested() (Transaction, error) {
	if !tx.options.AllowNested {
		return nil, fmt.Errorf("nested transactions are not allowed")
	}

	if tx.committed {
		return nil, fmt.Errorf("cannot create nested transaction on committed transaction")
	}

	// 为嵌套事务创建新的缓存，但共享同一个batch
	nestedTx := &SfsTransaction{
		store:     tx.store,
		batch:     tx.batch,    // 共享父事务的batch
		snapshot:  tx.snapshot, // 共享父事务的快照
		committed: false,
		cache:     make(map[string][]byte), // 新的缓存
		options:   tx.options,
		parent:    tx,
		txID:      uint64(time.Now().UnixNano()),
		startTime: time.Now(),
	}

	// 将嵌套事务添加到父事务的子事务列表
	tx.children = append(tx.children, nestedTx)

	return nestedTx, nil
}

// GetOptions 获取事务选项
func (tx *SfsTransaction) GetOptions() *TransactionOptions {
	return tx.options
}

// GetTxID 获取事务ID
func (tx *SfsTransaction) GetTxID() uint64 {
	return tx.txID
}

// checkCommitted 检查事务是否已提交
func (tx *SfsTransaction) checkCommitted() error {
	if tx.committed {
		return fmt.Errorf("transaction already committed")
	}
	return nil
}

// Commit 提交事务
func (tx *SfsTransaction) Commit() error {
	if err := tx.checkCommitted(); err != nil {
		return err
	}

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
	tx.children = nil

	// 4. 归还事务对象到池中
	if tx.parent == nil {
		// 只有根事务才归还到池，子事务由父事务管理
		GlobalTransactionPool.Put(tx)
	}

	return nil
}

// executeWithRetry 执行批量操作，支持重试
func (tx *SfsTransaction) executeWithRetry() error {
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
		err := tx.store.WriteBatch(tx.batch)
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
func (tx *SfsTransaction) isRetryableError(err error) bool {
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
func (tx *SfsTransaction) Rollback() error {
	if err := tx.checkCommitted(); err != nil {
		return err
	}

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
	tx.children = nil

	// 3. 归还事务对象到池中
	if tx.parent == nil {
		// 只有根事务才归还到池，子事务由父事务管理
		GlobalTransactionPool.Put(tx)
	}

	return nil
}

// CheckCommitted 检查表事务是否已提交
func (tx *TableTransaction) CheckCommitted() error {
	if tx.Committed {
		return fmt.Errorf("transaction already committed")
	}
	return nil
}

// Insert 在事务中插入记录
func (tx *TableTransaction) Insert(fields *map[string]interface{}) (int, error) {
	if err := tx.CheckCommitted(); err != nil {
		return 0, err
	}

	// 生成缓存键
	cacheKey := tx.GetCacheKey(fields)

	// 使用乐观并发控制检查冲突
	currentVersion := tx.OCCSupport.GetVersion(cacheKey)
	if tx.OCCSupport.CheckConflict(cacheKey, currentVersion) {
		return 0, fmt.Errorf("并发冲突，记录已被修改，请重试")
	}

	// 尝试获取写锁（非阻塞模式）
	err := tx.LockManager.AcquireLock(lock.LockRequest{
		Key:       cacheKey,
		Type:      lock.WriteLock,
		Mode:      lock.NonBlocking,
		Timeout:   0,
		Requestor: fmt.Sprintf("tx_%d", tx.TxID),
	})
	if err != nil {
		// 锁获取失败，使用阻塞模式重试
		err = tx.LockManager.AcquireLock(lock.LockRequest{
			Key:       cacheKey,
			Type:      lock.WriteLock,
			Mode:      lock.Blocking,
			Timeout:   5 * time.Second,
			Requestor: fmt.Sprintf("tx_%d", tx.TxID),
		})
		if err != nil {
			return 0, fmt.Errorf("获取锁失败: %v", err)
		}
	}

	// 执行插入操作
	id, err := tx.Table.Insert(fields, tx.Batch)
	if err != nil {
		// 操作失败，释放锁
		tx.LockManager.ReleaseLock(cacheKey, fmt.Sprintf("tx_%d", tx.TxID))
		return 0, err
	}

	// 将fields转换为fieldsBytes，用于生成主键和记录
	fieldsBytes := tx.Table.FieldsToBytes(fields)

	// 生成记录字节数组
	record := tx.Table.FormatRecord(fieldsBytes)

	// 将记录存入缓存，用于读取自己的写操作
	tx.Cache[cacheKey] = record

	// 缓存锁键（使用简单的缓存键作为锁键）
	tx.LockKeyCache[cacheKey] = cacheKey

	// 递增版本号
	tx.OCCSupport.IncrementVersion(cacheKey)

	return id, nil
}

// Update 在事务中更新记录
func (tx *TableTransaction) Update(fields *map[string]interface{}) error {
	if err := tx.CheckCommitted(); err != nil {
		return err
	}

	// 生成缓存键
	cacheKey := tx.GetCacheKey(fields)

	// 使用乐观并发控制检查冲突
	currentVersion := tx.OCCSupport.GetVersion(cacheKey)
	if tx.OCCSupport.CheckConflict(cacheKey, currentVersion) {
		return fmt.Errorf("并发冲突，记录已被修改，请重试")
	}

	// 尝试获取写锁（非阻塞模式）
	err := tx.LockManager.AcquireLock(lock.LockRequest{
		Key:       cacheKey,
		Type:      lock.WriteLock,
		Mode:      lock.NonBlocking,
		Timeout:   0,
		Requestor: fmt.Sprintf("tx_%d", tx.TxID),
	})
	if err != nil {
		// 锁获取失败，使用阻塞模式重试
		err = tx.LockManager.AcquireLock(lock.LockRequest{
			Key:       cacheKey,
			Type:      lock.WriteLock,
			Mode:      lock.Blocking,
			Timeout:   5 * time.Second,
			Requestor: fmt.Sprintf("tx_%d", tx.TxID),
		})
		if err != nil {
			return fmt.Errorf("获取锁失败: %v", err)
		}
	}

	// 执行更新操作
	err = tx.Table.Update(fields, tx.Batch)
	if err != nil {
		// 操作失败，释放锁
		tx.LockManager.ReleaseLock(cacheKey, fmt.Sprintf("tx_%d", tx.TxID))
		return err
	}

	// 从缓存中删除旧记录，强制后续读取从数据库获取最新值
	delete(tx.Cache, cacheKey)

	// 缓存锁键（使用简单的缓存键作为锁键）
	tx.LockKeyCache[cacheKey] = cacheKey

	// 递增版本号
	tx.OCCSupport.IncrementVersion(cacheKey)

	return nil
}

// Delete 在事务中删除记录
func (tx *TableTransaction) Delete(fields *map[string]interface{}) error {
	if err := tx.CheckCommitted(); err != nil {
		return err
	}

	// 生成缓存键
	cacheKey := tx.GetCacheKey(fields)

	// 使用乐观并发控制检查冲突
	currentVersion := tx.OCCSupport.GetVersion(cacheKey)
	if tx.OCCSupport.CheckConflict(cacheKey, currentVersion) {
		return fmt.Errorf("并发冲突，记录已被修改，请重试")
	}

	// 尝试获取写锁（非阻塞模式）
	err := tx.LockManager.AcquireLock(lock.LockRequest{
		Key:       cacheKey,
		Type:      lock.WriteLock,
		Mode:      lock.NonBlocking,
		Timeout:   0,
		Requestor: fmt.Sprintf("tx_%d", tx.TxID),
	})
	if err != nil {
		// 锁获取失败，使用阻塞模式重试
		err = tx.LockManager.AcquireLock(lock.LockRequest{
			Key:       cacheKey,
			Type:      lock.WriteLock,
			Mode:      lock.Blocking,
			Timeout:   5 * time.Second,
			Requestor: fmt.Sprintf("tx_%d", tx.TxID),
		})
		if err != nil {
			return fmt.Errorf("获取锁失败: %v", err)
		}
	}

	// 执行删除操作
	err = tx.Table.Delete(fields, tx.Batch)
	if err != nil {
		// 操作失败，释放锁
		tx.LockManager.ReleaseLock(cacheKey, fmt.Sprintf("tx_%d", tx.TxID))
		return err
	}

	// 从缓存中删除记录，确保读一致性
	delete(tx.Cache, cacheKey)

	// 缓存锁键（使用简单的缓存键作为锁键）
	tx.LockKeyCache[cacheKey] = cacheKey

	// 递增版本号
	tx.OCCSupport.IncrementVersion(cacheKey)

	return nil
}

// GetCacheKey 生成缓存键
func (tx *TableTransaction) GetCacheKey(fields *map[string]interface{}) string {
	// 将fields转换为fieldsBytes，用于生成主键
	fieldsBytes := tx.Table.FieldsToBytes(fields)
	defer func() {
		if fieldsBytes != nil && *fieldsBytes != nil {
			// 注意：这里我们不能直接使用engine.GlobalFieldsBytesPool.Put，因为它是未导出的
			// 我们需要使用其他方式管理内存
		}
	}()
	// 生成主键键值，用于缓存
	pkKey := tx.Table.GetPrimaryKey().JoinValue(fieldsBytes, tx.Table.GetId())
	return string(pkKey)
}

// Read 在事务中读取单条记录（支持读一致性）
func (tx *TableTransaction) Read(fields *map[string]interface{}) ([]byte, error) {
	if err := tx.CheckCommitted(); err != nil {
		return nil, err
	}

	// 生成缓存键
	cacheKey := tx.GetCacheKey(fields)

	// 1. 优先从缓存中读取，支持读取自己的写操作
	if record, exists := tx.Cache[cacheKey]; exists {
		return record, nil
	}

	// 2. 缓存中没有，尝试获取读锁（非阻塞模式）
	err := tx.LockManager.AcquireLock(lock.LockRequest{
		Key:       cacheKey,
		Type:      lock.ReadLock,
		Mode:      lock.NonBlocking,
		Timeout:   0,
		Requestor: fmt.Sprintf("tx_%d", tx.TxID),
	})
	if err != nil {
		// 锁获取失败，使用阻塞模式重试
		err = tx.LockManager.AcquireLock(lock.LockRequest{
			Key:       cacheKey,
			Type:      lock.ReadLock,
			Mode:      lock.Blocking,
			Timeout:   5 * time.Second,
			Requestor: fmt.Sprintf("tx_%d", tx.TxID),
		})
		if err != nil {
			return nil, fmt.Errorf("获取锁失败: %v", err)
		}
	}

	// 3. 根据隔离级别选择读取方式
	// 生成主键键值
	fieldsBytes := tx.Table.FieldsToBytes(fields)
	defer func() {
		if fieldsBytes != nil && *fieldsBytes != nil {
			// 注意：这里我们不能直接使用engine.GlobalFieldsBytesPool.Put，因为它是未导出的
			// 我们需要使用其他方式管理内存
		}
	}()
	pkKey := tx.Table.GetPrimaryKey().JoinValue(fieldsBytes, tx.Table.GetId())

	// 根据隔离级别选择读取方式
	switch tx.Options.IsolationLevel {
	case ReadUncommitted:
		// 对于ReadUncommitted，直接使用原始存储读取，允许读取未提交的数据
		return tx.OriginalStore.Get(pkKey)
	case ReadCommitted:
		// 对于ReadCommitted，每次读取都使用原始存储，确保只能读取已提交的数据
		// 注意：严格来说，ReadCommitted应该每次读取都创建新的快照
		// 但为了简化实现，当前版本使用直接读取原始存储的方式
		return tx.OriginalStore.Get(pkKey)
	case RepeatableRead, Serializable:
		// 对于RepeatableRead和Serializable，使用事务开始时创建的快照
		if tx.Snapshot != nil {
			return tx.Snapshot.Get(pkKey)
		}
		return tx.OriginalStore.Get(pkKey)
	default:
		// 默认使用原始存储读取
		return tx.OriginalStore.Get(pkKey)
	}
}

// Search 在事务中搜索记录（支持读一致性）
func (tx *TableTransaction) Search(fields *map[string]any, ops ...util.ComparisonOperator) (*engine.TableIter, error) {
	return tx.Searchs(nil, fields, ops...)
}

// Searchs 在事务中搜索记录（通过funIter支持原数据库或快照查询）
func (tx *TableTransaction) Searchs(funIter storage.FunIter, fields *map[string]any, ops ...util.ComparisonOperator) (*engine.TableIter, error) {
	if err := tx.CheckCommitted(); err != nil {
		return nil, err
	}

	// 创建一个函数，根据隔离级别和传入的funIter选择不同的存储获取迭代器
	transactionFunIter := func(start, limit []byte) storage.Iterator {
		// 如果传入了funIter，优先使用传入的funIter
		if funIter != nil {
			return funIter(start, limit)
		}

		// 根据隔离级别选择读取方式
		switch tx.Options.IsolationLevel {
		case ReadUncommitted:
			// 对于ReadUncommitted，直接使用原始存储
			return tx.OriginalStore.Iterator(start, limit)
		case ReadCommitted:
			// 对于ReadCommitted，每次读取都使用原始存储
			return tx.OriginalStore.Iterator(start, limit)
		case RepeatableRead, Serializable:
			// 对于RepeatableRead和Serializable，使用事务开始时创建的快照
			if tx.Snapshot != nil {
				return tx.Snapshot.Iterator(start, limit)
			}
			return tx.OriginalStore.Iterator(start, limit)
		default:
			// 默认使用原始存储
			return tx.OriginalStore.Iterator(start, limit)
		}
	}

	// 调用table.Searchs方法，传入事务的funIter函数
	return tx.Table.Searchs(transactionFunIter, fields, ops...)
}

// SearchRange 在事务中进行区间搜索（支持读一致性）
func (tx *TableTransaction) SearchRange(funIter storage.FunIter, Start, Limit *map[string]any) (*engine.TableIter, error) {
	if err := tx.CheckCommitted(); err != nil {
		return nil, err
	}

	// 创建一个函数，根据隔离级别和传入的funIter选择不同的存储获取迭代器
	transactionFunIter := func(start, limit []byte) storage.Iterator {
		// 如果传入了funIter，优先使用传入的funIter
		if funIter != nil {
			return funIter(start, limit)
		}

		// 根据隔离级别选择读取方式
		switch tx.Options.IsolationLevel {
		case ReadUncommitted:
			// 对于ReadUncommitted，直接使用原始存储
			return tx.OriginalStore.Iterator(start, limit)
		case ReadCommitted:
			// 对于ReadCommitted，每次读取都使用原始存储
			return tx.OriginalStore.Iterator(start, limit)
		case RepeatableRead, Serializable:
			// 对于RepeatableRead和Serializable，使用事务开始时创建的快照
			if tx.Snapshot != nil {
				return tx.Snapshot.Iterator(start, limit)
			}
			return tx.OriginalStore.Iterator(start, limit)
		default:
			// 默认使用原始存储
			return tx.OriginalStore.Iterator(start, limit)
		}
	}

	// 调用table.SearchRange方法，传入事务的funIter函数
	return tx.Table.SearchRange(transactionFunIter, Start, Limit)
}

// GetOptions 获取事务选项
func (tx *TableTransaction) GetOptions() *TransactionOptions {
	return tx.Options
}

// GetTxID 获取事务ID
func (tx *TableTransaction) GetTxID() uint64 {
	return tx.TxID
}

// BeginNested 创建一个嵌套表事务
func (tx *TableTransaction) BeginNested() (TableTransactionInterface, error) {
	if !tx.Options.AllowNested {
		return nil, fmt.Errorf("nested transactions are not allowed")
	}

	if tx.Committed {
		return nil, fmt.Errorf("cannot create nested transaction on committed transaction")
	}

	// 为嵌套事务创建新的缓存，但共享同一个batch
	nestedTx := &TableTransaction{
		Table:         tx.Table,
		Batch:         tx.Batch, // 共享父事务的batch
		Committed:     false,
		Snapshot:      tx.Snapshot, // 共享父事务的快照
		OriginalStore: tx.OriginalStore,
		Cache:         make(map[string][]byte), // 新的缓存
		LockKeyCache:  make(map[string]string), // 新的锁键缓存
		Options:       tx.Options,
		Parent:        tx,
		TxID:          uint64(time.Now().UnixNano()),
		StartTime:     time.Now(),
	}

	// 将嵌套事务添加到父事务的子事务列表
	tx.Children = append(tx.Children, nestedTx)

	return nestedTx, nil
}

// Commit 提交表事务
func (tx *TableTransaction) Commit() error {
	if err := tx.CheckCommitted(); err != nil {
		return err
	}

	// 记录事务结束时间并计算用时
	endTime := time.Now()                 // 记录事务结束时间
	duration := endTime.Sub(tx.StartTime) // 计算事务用时

	// 提交所有子事务
	for _, child := range tx.Children {
		if !child.Committed {
			if err := child.Commit(); err != nil {
				return err
			}
		}
	}

	// 只有根事务才真正提交batch
	if tx.Parent == nil {
		// 1. 提交批量操作
		// 使用原始存储执行写操作，支持重试
		startTime := time.Now()
		err := tx.ExecuteWithRetry()
		batchDuration := time.Since(startTime)

		// 记录批量操作统计信息
		tx.BatchOptimizer.RecordOperation(len(tx.LockKeyCache), batchDuration)

		if err != nil {
			// 提交失败，释放快照资源
			if tx.Snapshot != nil {
				tx.Snapshot.Release()
				// 注意：这里我们不能直接使用sfsDb的LdbSnapshotPool，因为它管理的是LevelDBStore类型
				// 而我们使用的是snapshot接口，类型不匹配
				tx.Snapshot = nil
			}
			// 释放所有锁
			tx.LockManager.ReleaseAllLocks(fmt.Sprintf("tx_%d", tx.TxID))
			// 记录事务失败的统计信息
			if monitor.GTransactionStatsMap != nil {
				tableName := ""
				if tx.Table != nil {
					tableName = tx.Table.GetName()
				}
				isolationLevel := ""
				if tx.Options != nil {
					isolationLevel = tx.Options.IsolationLevel
				}
				monitor.GTransactionStatsMap.SetTimeAsync(tx.TxID, duration, tableName, isolationLevel, false)
			}
			return err
		}

		// 2. 释放快照资源
		if tx.Snapshot != nil {
			tx.Snapshot.Release()
			// 注意：这里我们不能直接使用sfsDb的LdbSnapshotPool，因为它管理的是LevelDBStore类型
			// 而我们使用的是snapshot接口，类型不匹配
			tx.Snapshot = nil
		}
	}

	// 3. 释放所有锁
	tx.LockManager.ReleaseAllLocks(fmt.Sprintf("tx_%d", tx.TxID))

	// 4. 标记事务已结束，清空缓存
	tx.Committed = true
	tx.Cache = nil
	tx.LockKeyCache = nil
	tx.Children = nil

	// 5. 归还事务对象到池中
	if tx.Parent == nil {
		// 只有根事务才归还到池，子事务由父事务管理
		GlobalTableTransactionPool.Put(tx)

		// 记录事务成功的统计信息
		if monitor.GTransactionStatsMap != nil {
			tableName := ""
			if tx.Table != nil {
				tableName = tx.Table.GetName()
			}
			isolationLevel := ""
			if tx.Options != nil {
				isolationLevel = tx.Options.IsolationLevel
			}
			monitor.GTransactionStatsMap.SetTimeAsync(tx.TxID, duration, tableName, isolationLevel, true)
		}
	}

	return nil
}

// ExecuteWithRetry 执行批量操作，支持重试
func (tx *TableTransaction) ExecuteWithRetry() error {
	maxRetries := tx.Options.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}

	initialDelay := tx.Options.InitialRetryDelay
	if initialDelay <= 0 {
		initialDelay = 10 * time.Millisecond
	}

	backoffFactor := tx.Options.RetryBackoffFactor
	if backoffFactor < 1.0 {
		backoffFactor = 2.0
	}

	// 尝试执行，最多重试maxRetries次
	for attempt := 0; attempt <= maxRetries; attempt++ {
		// 执行批量操作
		err := tx.OriginalStore.WriteBatch(tx.Batch)
		if err == nil {
			// 执行成功
			return nil
		}

		// 检查是否是可重试的错误
		if !tx.IsRetryableError(err) {
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

// IsRetryableError 判断错误是否可重试
func (tx *TableTransaction) IsRetryableError(err error) bool {
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

// Rollback 回滚表事务
// 注意：LevelDB的WriteBatch不支持真正的回滚
// 此方法仅标记事务已结束，防止重复提交，并释放快照资源
func (tx *TableTransaction) Rollback() error {
	if err := tx.CheckCommitted(); err != nil {
		return err
	}

	// 记录事务结束时间并计算用时
	endTime := time.Now()                 // 记录事务结束时间
	duration := endTime.Sub(tx.StartTime) // 计算事务用时

	// 回滚所有子事务
	for _, child := range tx.Children {
		if !child.Committed {
			if err := child.Rollback(); err != nil {
				return err
			}
		}
	}

	// 只有根事务才释放快照资源
	if tx.Parent == nil {
		// 1. 释放快照资源
		if tx.Snapshot != nil {
			tx.Snapshot.Release()
			// 注意：这里我们不能直接使用sfsDb的LdbSnapshotPool，因为它管理的是LevelDBStore类型
			// 而我们使用的是snapshot接口，类型不匹配
			tx.Snapshot = nil
		}
	}

	// 2. 释放所有锁
	tx.LockManager.ReleaseAllLocks(fmt.Sprintf("tx_%d", tx.TxID))

	// 3. 标记事务已结束，清空缓存
	tx.Committed = true
	tx.Cache = nil
	tx.LockKeyCache = nil
	tx.Children = nil

	// 4. 归还事务对象到池中
	if tx.Parent == nil {
		// 只有根事务才归还到池，子事务由父事务管理
		GlobalTableTransactionPool.Put(tx)

		// 记录事务回滚的统计信息
		if monitor.GTransactionStatsMap != nil {
			tableName := ""
			if tx.Table != nil {
				tableName = tx.Table.GetName()
			}
			isolationLevel := ""
			if tx.Options != nil {
				isolationLevel = tx.Options.IsolationLevel
			}
			monitor.GTransactionStatsMap.SetTimeAsync(tx.TxID, duration, tableName, isolationLevel, false)
		}
	}

	return nil
}
