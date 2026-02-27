package transactionLockANT

import (
	"fmt"
	"sync"
	"time"

	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/storage"
	"github.com/liaoran123/sfsDb/transactionLockANT/recovery"
	"github.com/liaoran123/sfsDb/transactionLockANT/wal"
	"github.com/liaoran123/sfsDb/util"
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

// 全局无锁版本管理器
var GlobalVersionManager = NewLockFreeVersionManager()

// 全局WAL实例
var GlobalWAL *wal.WAL

// 全局恢复管理器
var GlobalRecoveryManager *recovery.RecoveryManager

// 全局加密管理器
var GlobalEncryptionManager *EncryptionManager

// InitWAL 初始化WAL
func InitWAL(db storage.Store, logDir string) error {
	if levelDBStore, ok := db.(*storage.LevelDBStore); ok {
		// 获取底层LevelDB实例
		ldb := levelDBStore.GetDB()
		if ldb == nil {
			return fmt.Errorf("failed to get leveldb.DB instance")
		}

		walInstance, err := wal.NewWAL(ldb, logDir)
		if err != nil {
			return err
		}
		GlobalWAL = walInstance

		// 初始化恢复管理器
		GlobalRecoveryManager = recovery.NewRecoveryManager(ldb, walInstance, 10*time.Minute)

		// 启动检查点管理器
		GlobalRecoveryManager.StartCheckpointManager()

		// 执行恢复
		return GlobalRecoveryManager.Recover()
	}
	return fmt.Errorf("unsupported store type")
}

// GetWAL 获取WAL实例
func GetWAL() *wal.WAL {
	return GlobalWAL
}

// GetRecoveryManager 获取恢复管理器
func GetRecoveryManager() *recovery.RecoveryManager {
	return GlobalRecoveryManager
}

// InitEncryption 初始化加密管理器
func InitEncryption(config *TransactionEncryptionConfig) error {
	manager, err := NewEncryptionManager(config)
	if err != nil {
		return err
	}
	GlobalEncryptionManager = manager
	return nil
}

// GetEncryptionManager 获取加密管理器
func GetEncryptionManager() *EncryptionManager {
	return GlobalEncryptionManager
}

// IsAccessControlEnabled 检查访问控制是否启用
func IsAccessControlEnabled() bool {
	return accessControlEnabled
}

// 全局访问控制启用标志
var accessControlEnabled bool

// setAccessControlEnabled 设置访问控制启用状态
func setAccessControlEnabled(enabled bool) {
	accessControlEnabled = enabled
}

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
	tx.readSet = make(map[string]bool)
	tx.writeSet = make(map[string]bool)
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
	tx.ReadSet = make(map[string]bool)
	tx.WriteSet = make(map[string]bool)
	tx.savepoints = nil
	tx.Parent = nil
	tx.Children = nil
	tx.VersionManager = NewLockFreeVersionManager() // 重置版本管理器
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
	// CreateSavepoint 创建保存点
	CreateSavepoint(name string) error
	// RollbackToSavepoint 回滚到保存点
	RollbackToSavepoint(name string) error
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
	// CreateSavepoint 创建保存点
	CreateSavepoint(name string) error
	// RollbackToSavepoint 回滚到保存点
	RollbackToSavepoint(name string) error
	// GetOptions 获取事务选项
	GetOptions() *TransactionOptions
	// GetTxID 获取事务ID
	GetTxID() uint64
}

// Savepoint 保存点结构
type Savepoint struct {
	name     string
	cache    map[string][]byte // 保存缓存状态
	writeSet map[string]bool   // 保存写集状态
}

// SfsTransaction 基于sfsdb的无锁事务实现
type SfsTransaction struct {
	store      storage.Store
	batch      storage.Batch
	snapshot   storage.Snapshot
	committed  bool
	cache      map[string][]byte     // 事务内修改缓存
	readSet    map[string]bool       // 读集，用于Serializable隔离级别
	writeSet   map[string]bool       // 写集，用于冲突检测
	savepoints map[string]*Savepoint // 保存点
	options    *TransactionOptions
	parent     *SfsTransaction
	children   []*SfsTransaction
	txID       uint64
	startTime  time.Time
	wal        *wal.WAL           // 事务日志
	encryption *EncryptionManager // 加密管理器
	userID     string             // 用户ID，用于权限检查
}

// TableTransaction 基于sfsdb的无锁表事务实现
type TableTransaction struct {
	Table          *engine.Table           // 关联的表
	Batch          storage.Batch           // 事务使用的batch，原子性
	Committed      bool                    // 是否已提交，提交成功后则是持久性。
	Snapshot       storage.Snapshot        // 事务使用的快照，一致性。
	OriginalStore  storage.Store           // 原始存储，用于写操作
	Cache          map[string][]byte       // 事务内修改缓存，用于读取自己的写操作
	ReadSet        map[string]bool         // 读集，用于Serializable隔离级别
	WriteSet       map[string]bool         // 写集，用于冲突检测
	savepoints     map[string]*Savepoint   // 保存点
	Options        *TransactionOptions     // 事务选项
	Parent         *TableTransaction       // 父事务（用于嵌套事务）
	Children       []*TableTransaction     // 子事务列表
	TxID           uint64                  // 事务ID
	StartTime      time.Time               // 事务开始时间
	VersionManager *LockFreeVersionManager // 无锁版本管理器
	WAL            *wal.WAL                // 事务日志
	Encryption     *EncryptionManager      // 加密管理器
	UserID         string                  // 用户ID，用于权限检查
}

// NewTransaction 创建新的事务
func NewTransaction(store storage.Store, userID string) (*SfsTransaction, error) {
	return NewTransactionWithOptions(store, DefaultTransactionOptions(), userID)
}

// NewTransactionWithOptions 使用指定选项创建新的事务
func NewTransactionWithOptions(store storage.Store, options *TransactionOptions, userID string) (*SfsTransaction, error) {
	// 验证用户存在
	if userID != "" {
		acm := GetAccessControlManager()
		if acm == nil {
			return nil, fmt.Errorf("access control manager not initialized")
		}
		_, err := acm.GetUser(userID)
		if err != nil {
			return nil, fmt.Errorf("invalid user: %v", err)
		}
	}

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
	tx.cache = make(map[string][]byte)  // 初始化事务内缓存
	tx.readSet = make(map[string]bool)  // 初始化读集
	tx.writeSet = make(map[string]bool) // 初始化写集
	tx.options = options
	tx.txID = txID
	tx.startTime = time.Now()
	tx.wal = GlobalWAL
	tx.encryption = GlobalEncryptionManager
	tx.userID = userID

	// 写入事务开始日志
	if tx.wal != nil {
		beginRecord := wal.LogRecord{
			Type:      wal.LogTypeBegin,
			TxID:      txID,
			Timestamp: time.Now().UnixNano(),
		}
		tx.wal.WriteLog(beginRecord)
	}

	return tx, nil
}

// NewTableTransaction 为指定表创建事务
func NewTableTransaction(table *engine.Table, userID string) (*TableTransaction, error) {
	options := DefaultTransactionOptions()
	return NewTableTransactionWithOptions(table, options, userID)
}

// NewTableTransactionWithOptions 为指定表创建带选项的事务
func NewTableTransactionWithOptions(table *engine.Table, options *TransactionOptions, userID string) (*TableTransaction, error) {
	// 创建批量操作对象
	tempBatch := storage.GetDBManager().GetDB().GetBatch()
	if tempBatch == nil {
		return nil, fmt.Errorf("failed to create batch")
	}

	return NewTableTransactionWithBatchAndOptions(table, tempBatch, options, userID)
}

// NewTableTransactionWithBatch 创建一个使用外部传入batch的表事务
func NewTableTransactionWithBatch(table *engine.Table, batch storage.Batch, userID string) (*TableTransaction, error) {
	options := DefaultTransactionOptions()
	return NewTableTransactionWithBatchAndOptions(table, batch, options, userID)
}

// NewTableTransactionWithBatchAndOptions 使用外部传入batch和指定选项创建一个表事务
func NewTableTransactionWithBatchAndOptions(table *engine.Table, batch storage.Batch, options *TransactionOptions, userID string) (*TableTransaction, error) {
	// 验证用户存在
	if userID != "" {
		acm := GetAccessControlManager()
		if acm == nil {
			return nil, fmt.Errorf("access control manager not initialized")
		}
		_, err := acm.GetUser(userID)
		if err != nil {
			return nil, fmt.Errorf("invalid user: %v", err)
		}
	}

	if batch == nil {
		return nil, fmt.Errorf("batch cannot be nil")
	}

	if options == nil {
		options = DefaultTransactionOptions()
	}

	// 为每个事务创建自己的快照实例
	var snapshot storage.Snapshot
	var err error

	// 根据隔离级别决定是否创建快照
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
		// 对于ReadCommitted，暂时不创建快照
	case ReadUncommitted:
		// 对于ReadUncommitted，不创建快照
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
	tx.Cache = make(map[string][]byte)  // 初始化事务内缓存
	tx.ReadSet = make(map[string]bool)  // 初始化读集
	tx.WriteSet = make(map[string]bool) // 初始化写集
	tx.Options = options
	tx.TxID = txID
	tx.StartTime = time.Now()
	tx.VersionManager = NewLockFreeVersionManager() // 初始化无锁版本管理器
	tx.WAL = GlobalWAL
	tx.Encryption = GlobalEncryptionManager
	tx.UserID = userID

	// 写入事务开始日志
	if tx.WAL != nil {
		beginRecord := wal.LogRecord{
			Type:      wal.LogTypeBegin,
			TxID:      txID,
			Timestamp: time.Now().UnixNano(),
		}
		tx.WAL.WriteLog(beginRecord)
	}

	return tx, nil
}

// Get 从事务中获取值
func (tx *SfsTransaction) Get(key []byte) ([]byte, error) {
	if tx.committed {
		return nil, fmt.Errorf("transaction already committed")
	}

	// 权限检查：读取操作（非入侵式）
	if tx.userID != "" && IsAccessControlEnabled() {
		acm := GetAccessControlManager()
		if acm != nil {
			hasPermission, _ := acm.CheckPermission(tx.userID, ResourceTypeTable, "*", PermissionRead)
			if !hasPermission {
				// 权限检查失败，返回错误（读取操作需要返回错误）
				// 记录审计日志
				LogAudit(AuditActionRead, AuditResourceTable, string(key), tx.userID, "Permission denied", AuditStatusFailed)
				return nil, fmt.Errorf("permission denied: read operation not allowed")
			}
		}
	}

	// 1. 优先从缓存中读取，支持读取自己的写操作
	cacheKey := string(key)
	if value, exists := tx.cache[cacheKey]; exists {
		// 对于Serializable隔离级别，即使从缓存读取也要记录到读集
		if tx.options.IsolationLevel == Serializable {
			tx.readSet[cacheKey] = true
		}
		// 记录审计日志
		LogAudit(AuditActionRead, AuditResourceTable, string(key), tx.userID, "Read from cache", AuditStatusSuccess)
		return value, nil
	}

	// 2. 缓存中没有，根据隔离级别选择读取方式
	var value []byte
	var err error

	switch tx.options.IsolationLevel {
	case ReadUncommitted:
		// 对于ReadUncommitted，尝试读取其他事务的未提交数据
		// 这里简化实现，直接使用原始存储，实际生产环境中需要更复杂的实现
		value, err = tx.store.Get(key)
	case ReadCommitted:
		// 对于ReadCommitted，每次读取都使用原始存储
		value, err = tx.store.Get(key)
	case RepeatableRead, Serializable:
		// 对于RepeatableRead和Serializable，使用事务开始时创建的快照
		if tx.snapshot != nil {
			value, err = tx.snapshot.Get(key)
		} else {
			value, err = tx.store.Get(key)
		}
	default:
		// 默认使用原始存储
		value, err = tx.store.Get(key)
	}

	// 解密数据
	if err == nil && tx.encryption != nil {
		decryptedValue, decryptErr := tx.encryption.Decrypt(value)
		if decryptErr == nil {
			value = decryptedValue
		}
	}

	// 对于Serializable隔离级别，记录读集
	if tx.options.IsolationLevel == Serializable && err == nil {
		tx.readSet[cacheKey] = true
	}

	// 记录审计日志
	if err == nil {
		LogAudit(AuditActionRead, AuditResourceTable, string(key), tx.userID, "Read from storage", AuditStatusSuccess)
	} else {
		LogAudit(AuditActionRead, AuditResourceTable, string(key), tx.userID, err.Error(), AuditStatusFailed)
	}

	return value, err
}

// Put 向事务中设置值
func (tx *SfsTransaction) Put(key []byte, value []byte) {
	if tx.committed {
		return
	}

	// 权限检查：写入操作
	if tx.userID != "" && IsAccessControlEnabled() {
		acm := GetAccessControlManager()
		if acm != nil {
			hasPermission, _ := acm.CheckPermission(tx.userID, ResourceTypeTable, "*", PermissionWrite)
			if !hasPermission {
				// 权限检查失败，记录错误但继续执行（非入侵式）
				if tx.wal != nil {
					errorRecord := wal.LogRecord{
						Type:      wal.LogTypeRollback,
						TxID:      tx.txID,
						Key:       key,
						Value:     []byte("permission denied: write operation not allowed"),
						Timestamp: time.Now().UnixNano(),
					}
					tx.wal.WriteLog(errorRecord)
				}
				// 记录审计日志
				LogAudit(AuditActionUpdate, AuditResourceTable, string(key), tx.userID, "Permission denied", AuditStatusFailed)
				return
			}
		}
	}

	cacheKey := string(key)
	// 对于Serializable隔离级别，检查写集冲突
	if tx.options.IsolationLevel == Serializable {
		if _, exists := tx.writeSet[cacheKey]; exists {
			// 已经写入过，不需要重复记录
		} else {
			// 记录写集
			tx.writeSet[cacheKey] = true
		}
	}

	// 加密数据
	encryptedValue := value
	if tx.encryption != nil {
		var err error
		encryptedValue, err = tx.encryption.Encrypt(value)
		if err != nil {
			// 加密失败，使用原始值
			encryptedValue = value
		}
	}

	tx.batch.Put(key, encryptedValue)
	// 将原始值存入缓存，用于读取自己的写操作
	tx.cache[cacheKey] = value

	// 写入事务日志
	if tx.wal != nil {
		putRecord := wal.LogRecord{
			Type:      wal.LogTypePut,
			TxID:      tx.txID,
			Key:       key,
			Value:     encryptedValue,
			Timestamp: time.Now().UnixNano(),
		}
		tx.wal.WriteLog(putRecord)
	}

	// 记录审计日志
	LogAudit(AuditActionUpdate, AuditResourceTable, string(key), tx.userID, "Write operation", AuditStatusSuccess)
}

// Delete 从事务中删除值
func (tx *SfsTransaction) Delete(key []byte) {
	if tx.committed {
		return
	}

	// 权限检查：删除操作
	if tx.userID != "" && IsAccessControlEnabled() {
		acm := GetAccessControlManager()
		if acm != nil {
			hasPermission, _ := acm.CheckPermission(tx.userID, ResourceTypeTable, "*", PermissionDelete)
			if !hasPermission {
				// 权限检查失败，记录错误但继续执行（非入侵式）
				if tx.wal != nil {
					errorRecord := wal.LogRecord{
						Type:      wal.LogTypeRollback,
						TxID:      tx.txID,
						Key:       key,
						Value:     []byte("permission denied: delete operation not allowed"),
						Timestamp: time.Now().UnixNano(),
					}
					tx.wal.WriteLog(errorRecord)
				}
				// 记录审计日志
				LogAudit(AuditActionDelete, AuditResourceTable, string(key), tx.userID, "Permission denied", AuditStatusFailed)
				return
			}
		}
	}

	cacheKey := string(key)
	// 对于Serializable隔离级别，检查写集冲突
	if tx.options.IsolationLevel == Serializable {
		if _, exists := tx.writeSet[cacheKey]; exists {
			// 已经写入过，不需要重复记录
		} else {
			// 记录写集
			tx.writeSet[cacheKey] = true
		}
	}
	tx.batch.Delete(key)
	// 从缓存中删除记录，确保读一致性
	delete(tx.cache, cacheKey)

	// 写入事务日志
	if tx.wal != nil {
		deleteRecord := wal.LogRecord{
			Type:      wal.LogTypeDelete,
			TxID:      tx.txID,
			Key:       key,
			Timestamp: time.Now().UnixNano(),
		}
		tx.wal.WriteLog(deleteRecord)
	}

	// 记录审计日志
	LogAudit(AuditActionDelete, AuditResourceTable, string(key), tx.userID, "Delete operation", AuditStatusSuccess)
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

// CreateSavepoint 创建保存点
func (tx *SfsTransaction) CreateSavepoint(name string) error {
	if tx.committed {
		return fmt.Errorf("transaction already committed")
	}

	// 初始化保存点映射
	if tx.savepoints == nil {
		tx.savepoints = make(map[string]*Savepoint)
	}

	// 创建保存点，复制当前缓存和写集状态
	cacheCopy := make(map[string][]byte)
	for k, v := range tx.cache {
		cacheCopy[k] = v
	}

	writeSetCopy := make(map[string]bool)
	for k, v := range tx.writeSet {
		writeSetCopy[k] = v
	}

	// 创建并存储保存点
	savepoint := &Savepoint{
		name:     name,
		cache:    cacheCopy,
		writeSet: writeSetCopy,
	}
	tx.savepoints[name] = savepoint

	return nil
}

// RollbackToSavepoint 回滚到保存点
func (tx *SfsTransaction) RollbackToSavepoint(name string) error {
	if tx.committed {
		return fmt.Errorf("transaction already committed")
	}

	// 检查保存点是否存在
	savepoint, exists := tx.savepoints[name]
	if !exists {
		return fmt.Errorf("savepoint not found")
	}

	// 恢复缓存和写集状态
	tx.cache = savepoint.cache
	tx.writeSet = savepoint.writeSet

	// 重置Batch，确保回滚后不会提交之前的操作
	tx.batch.Reset()

	// 移除保存点之后创建的保存点
	for spName := range tx.savepoints {
		if spName > name {
			delete(tx.savepoints, spName)
		}
	}

	return nil
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
		// 对于Serializable隔离级别，执行冲突检测
		if tx.options.IsolationLevel == Serializable {
			// 检查读集与其他事务的写集是否冲突
			for key := range tx.readSet {
				// 获取当前版本号
				currentVersion := GlobalVersionManager.GetVersion(key)
				// 检查是否有冲突
				if GlobalVersionManager.CheckConflict(key, currentVersion) {
					// 提交失败，释放快照资源
					if tx.snapshot != nil {
						tx.snapshot.Release()
					}
					// 记录审计日志
					LogAudit(AuditActionCommit, AuditResourceTransaction, fmt.Sprintf("%d", tx.txID), tx.userID, "Serialization conflict", AuditStatusFailed)
					return fmt.Errorf("可串行化冲突，读取的数据已被其他事务修改，请重试")
				}
			}
		}

		// 1. 提交批量操作
		// 使用原始存储执行写操作，支持重试
		err := tx.executeWithRetry()
		if err != nil {
			// 提交失败，释放快照资源
			if tx.snapshot != nil {
				tx.snapshot.Release()
			}
			// 记录审计日志
			LogAudit(AuditActionCommit, AuditResourceTransaction, fmt.Sprintf("%d", tx.txID), tx.userID, err.Error(), AuditStatusFailed)
			return err
		}

		// 2. 对于Serializable隔离级别，更新版本号
		if tx.options.IsolationLevel == Serializable {
			for key := range tx.writeSet {
				GlobalVersionManager.IncrementVersion(key)
			}
		}

		// 3. 写入事务提交日志
		if tx.wal != nil {
			commitRecord := wal.LogRecord{
				Type:      wal.LogTypeCommit,
				TxID:      tx.txID,
				Timestamp: time.Now().UnixNano(),
			}
			tx.wal.WriteLog(commitRecord)
		}

		// 4. 释放快照资源
		if tx.snapshot != nil {
			tx.snapshot.Release()
		}

		// 记录审计日志
		LogAudit(AuditActionCommit, AuditResourceTransaction, fmt.Sprintf("%d", tx.txID), tx.userID, "Transaction committed", AuditStatusSuccess)
	}

	// 5. 标记事务已结束，清空缓存
	tx.committed = true
	tx.cache = nil
	tx.readSet = nil
	tx.writeSet = nil
	tx.children = nil
	tx.savepoints = nil

	// 6. 归还事务对象到池中
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
	// 暂时默认所有错误都可重试
	return true
}

// Rollback 回滚事务
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
		// 1. 写入事务回滚日志
		if tx.wal != nil {
			rollbackRecord := wal.LogRecord{
				Type:      wal.LogTypeRollback,
				TxID:      tx.txID,
				Timestamp: time.Now().UnixNano(),
			}
			tx.wal.WriteLog(rollbackRecord)
		}

		// 2. 释放快照资源
		if tx.snapshot != nil {
			tx.snapshot.Release()
		}

		// 记录审计日志
		LogAudit(AuditActionRollback, AuditResourceTransaction, fmt.Sprintf("%d", tx.txID), tx.userID, "Transaction rolled back", AuditStatusSuccess)
	}

	// 3. 标记事务已结束，清空缓存
	tx.committed = true
	tx.cache = nil
	tx.readSet = nil
	tx.writeSet = nil
	tx.children = nil
	tx.savepoints = nil

	// 4. 归还事务对象到池中
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

	// 权限检查：创建操作（非入侵式）
	if tx.UserID != "" && IsAccessControlEnabled() {
		acm := GetAccessControlManager()
		if acm != nil {
			tableName := tx.Table.GetName()
			hasPermission, _ := acm.CheckTablePermission(tx.UserID, tableName, PermissionCreate)
			if !hasPermission {
				// 记录审计日志
				LogAudit(AuditActionCreate, AuditResourceTable, tableName, tx.UserID, "Permission denied", AuditStatusFailed)
				return 0, fmt.Errorf("permission denied: create operation not allowed")
			}
		}
	}

	// 生成缓存键
	cacheKey := tx.GetCacheKey(fields)

	// 对于Serializable隔离级别，检查写集冲突
	if tx.Options.IsolationLevel == Serializable {
		if _, exists := tx.WriteSet[cacheKey]; exists {
			// 记录审计日志
			LogAudit(AuditActionCreate, AuditResourceTable, tx.Table.GetName(), tx.UserID, "Duplicate write in transaction", AuditStatusFailed)
			return 0, fmt.Errorf("事务内重复写入同一记录")
		}
		// 记录写集
		tx.WriteSet[cacheKey] = true
	}

	// 使用无锁版本管理器检查冲突
	currentVersion := tx.VersionManager.GetVersion(cacheKey)
	if tx.VersionManager.CheckConflict(cacheKey, currentVersion) {
		// 记录审计日志
		LogAudit(AuditActionCreate, AuditResourceTable, tx.Table.GetName(), tx.UserID, "Concurrency conflict", AuditStatusFailed)
		return 0, fmt.Errorf("并发冲突，记录已被修改，请重试")
	}

	// 执行插入操作
	id, err := tx.Table.Insert(fields, tx.Batch)
	if err != nil {
		// 记录审计日志
		LogAudit(AuditActionCreate, AuditResourceTable, tx.Table.GetName(), tx.UserID, err.Error(), AuditStatusFailed)
		return 0, err
	}

	// 将fields转换为fieldsBytes，用于生成主键和记录
	fieldsBytes := tx.Table.FieldsToBytes(fields)

	// 生成记录字节数组
	record := tx.Table.FormatRecord(fieldsBytes)

	// 加密记录
	encryptedRecord := record
	if tx.Encryption != nil {
		var encryptErr error
		encryptedRecord, encryptErr = tx.Encryption.Encrypt(record)
		if encryptErr != nil {
			// 加密失败，使用原始记录
			encryptedRecord = record
		}
	}

	// 将原始记录存入缓存，用于读取自己的写操作
	tx.Cache[cacheKey] = record

	// 无锁递增版本号
	tx.VersionManager.IncrementVersion(cacheKey)

	// 写入事务日志
	if tx.WAL != nil {
		// 生成主键键值
		pkKey := tx.Table.GetPrimaryKey().JoinValue(fieldsBytes, tx.Table.GetId())
		putRecord := wal.LogRecord{
			Type:      wal.LogTypePut,
			TxID:      tx.TxID,
			Key:       pkKey,
			Value:     encryptedRecord,
			Timestamp: time.Now().UnixNano(),
		}
		tx.WAL.WriteLog(putRecord)
	}

	// 记录审计日志
	LogAudit(AuditActionCreate, AuditResourceTable, tx.Table.GetName(), tx.UserID, fmt.Sprintf("Inserted record with ID: %d", id), AuditStatusSuccess)

	return id, nil
}

// Update 在事务中更新记录
func (tx *TableTransaction) Update(fields *map[string]interface{}) error {
	if err := tx.CheckCommitted(); err != nil {
		return err
	}

	// 权限检查：写入操作（非入侵式）
	if tx.UserID != "" && IsAccessControlEnabled() {
		acm := GetAccessControlManager()
		if acm != nil {
			tableName := tx.Table.GetName()
			hasPermission, _ := acm.CheckTablePermission(tx.UserID, tableName, PermissionWrite)
			if !hasPermission {
				// 记录审计日志
				LogAudit(AuditActionUpdate, AuditResourceTable, tableName, tx.UserID, "Permission denied", AuditStatusFailed)
				return fmt.Errorf("permission denied: write operation not allowed")
			}
		}
	}

	// 生成缓存键
	cacheKey := tx.GetCacheKey(fields)

	// 对于Serializable隔离级别，检查写集冲突
	if tx.Options.IsolationLevel == Serializable {
		if _, exists := tx.WriteSet[cacheKey]; exists {
			// 记录审计日志
			LogAudit(AuditActionUpdate, AuditResourceTable, tx.Table.GetName(), tx.UserID, "Duplicate write in transaction", AuditStatusFailed)
			return fmt.Errorf("事务内重复写入同一记录")
		}
		// 记录写集
		tx.WriteSet[cacheKey] = true
	}

	// 使用无锁版本管理器检查冲突
	currentVersion := tx.VersionManager.GetVersion(cacheKey)
	if tx.VersionManager.CheckConflict(cacheKey, currentVersion) {
		// 记录审计日志
		LogAudit(AuditActionUpdate, AuditResourceTable, tx.Table.GetName(), tx.UserID, "Concurrency conflict", AuditStatusFailed)
		return fmt.Errorf("并发冲突，记录已被修改，请重试")
	}

	// 执行更新操作
	err := tx.Table.Update(fields, tx.Batch)
	if err != nil {
		// 记录审计日志
		LogAudit(AuditActionUpdate, AuditResourceTable, tx.Table.GetName(), tx.UserID, err.Error(), AuditStatusFailed)
		return err
	}

	// 从缓存中删除旧记录，强制后续读取从数据库获取最新值
	delete(tx.Cache, cacheKey)

	// 无锁递增版本号
	tx.VersionManager.IncrementVersion(cacheKey)

	// 写入事务日志
	if tx.WAL != nil {
		// 将fields转换为fieldsBytes，用于生成主键
		fieldsBytes := tx.Table.FieldsToBytes(fields)
		// 生成主键键值
		pkKey := tx.Table.GetPrimaryKey().JoinValue(fieldsBytes, tx.Table.GetId())
		// 读取最新记录值
		var record []byte
		if tx.Snapshot != nil {
			record, _ = tx.Snapshot.Get(pkKey)
		} else {
			record, _ = tx.OriginalStore.Get(pkKey)
		}
		putRecord := wal.LogRecord{
			Type:      wal.LogTypePut,
			TxID:      tx.TxID,
			Key:       pkKey,
			Value:     record,
			Timestamp: time.Now().UnixNano(),
		}
		tx.WAL.WriteLog(putRecord)
	}

	// 记录审计日志
	LogAudit(AuditActionUpdate, AuditResourceTable, tx.Table.GetName(), tx.UserID, "Updated record", AuditStatusSuccess)

	return nil
}

// Delete 在事务中删除记录
func (tx *TableTransaction) Delete(fields *map[string]interface{}) error {
	if err := tx.CheckCommitted(); err != nil {
		return err
	}

	// 权限检查：删除操作（非入侵式）
	if tx.UserID != "" && IsAccessControlEnabled() {
		acm := GetAccessControlManager()
		if acm != nil {
			tableName := tx.Table.GetName()
			hasPermission, _ := acm.CheckTablePermission(tx.UserID, tableName, PermissionDelete)
			if !hasPermission {
				// 记录审计日志
				LogAudit(AuditActionDelete, AuditResourceTable, tableName, tx.UserID, "Permission denied", AuditStatusFailed)
				return fmt.Errorf("permission denied: delete operation not allowed")
			}
		}
	}

	// 生成缓存键
	cacheKey := tx.GetCacheKey(fields)

	// 对于Serializable隔离级别，检查写集冲突
	if tx.Options.IsolationLevel == Serializable {
		if _, exists := tx.WriteSet[cacheKey]; exists {
			// 记录审计日志
			LogAudit(AuditActionDelete, AuditResourceTable, tx.Table.GetName(), tx.UserID, "Duplicate write in transaction", AuditStatusFailed)
			return fmt.Errorf("事务内重复写入同一记录")
		}
		// 记录写集
		tx.WriteSet[cacheKey] = true
	}

	// 使用无锁版本管理器检查冲突
	currentVersion := tx.VersionManager.GetVersion(cacheKey)
	if tx.VersionManager.CheckConflict(cacheKey, currentVersion) {
		// 记录审计日志
		LogAudit(AuditActionDelete, AuditResourceTable, tx.Table.GetName(), tx.UserID, "Concurrency conflict", AuditStatusFailed)
		return fmt.Errorf("并发冲突，记录已被修改，请重试")
	}

	// 执行删除操作
	err := tx.Table.Delete(fields, tx.Batch)
	if err != nil {
		// 记录审计日志
		LogAudit(AuditActionDelete, AuditResourceTable, tx.Table.GetName(), tx.UserID, err.Error(), AuditStatusFailed)
		return err
	}

	// 从缓存中删除记录，确保读一致性
	delete(tx.Cache, cacheKey)

	// 无锁递增版本号
	tx.VersionManager.IncrementVersion(cacheKey)

	// 写入事务日志
	if tx.WAL != nil {
		// 将fields转换为fieldsBytes，用于生成主键
		fieldsBytes := tx.Table.FieldsToBytes(fields)
		// 生成主键键值
		pkKey := tx.Table.GetPrimaryKey().JoinValue(fieldsBytes, tx.Table.GetId())
		deleteRecord := wal.LogRecord{
			Type:      wal.LogTypeDelete,
			TxID:      tx.TxID,
			Key:       pkKey,
			Timestamp: time.Now().UnixNano(),
		}
		tx.WAL.WriteLog(deleteRecord)
	}

	// 记录审计日志
	LogAudit(AuditActionDelete, AuditResourceTable, tx.Table.GetName(), tx.UserID, "Deleted record", AuditStatusSuccess)

	return nil
}

// OptimisticUpdate 乐观更新操作，适用于复杂事务
// 提供更细粒度的冲突检测和自动重试机制
func (tx *TableTransaction) OptimisticUpdate(fields *map[string]interface{}, maxRetries int) error {
	if err := tx.CheckCommitted(); err != nil {
		return err
	}

	// 生成缓存键
	cacheKey := tx.GetCacheKey(fields)

	// 尝试执行更新，支持重试
	for attempt := 0; attempt <= maxRetries; attempt++ {
		// 使用无锁版本管理器检查冲突
		currentVersion := tx.VersionManager.GetVersion(cacheKey)
		if tx.VersionManager.CheckConflict(cacheKey, currentVersion) {
			// 检测到冲突，等待后重试
			delay := time.Duration(attempt*10) * time.Millisecond
			time.Sleep(delay)
			continue
		}

		// 执行更新操作
		err := tx.Table.Update(fields, tx.Batch)
		if err != nil {
			return err
		}

		// 从缓存中删除旧记录，强制后续读取从数据库获取最新值
		delete(tx.Cache, cacheKey)

		// 无锁递增版本号
		tx.VersionManager.IncrementVersion(cacheKey)

		return nil
	}

	return fmt.Errorf("更新失败，多次尝试后仍无法解决并发冲突")
}

// BatchOperations 批量操作，适用于复杂事务中的多个操作
// 通过一次Batch提交多个操作，减少磁盘I/O，提高性能
func (tx *TableTransaction) BatchOperations(operations []func() error) error {
	if err := tx.CheckCommitted(); err != nil {
		return err
	}

	// 执行所有操作
	for _, operation := range operations {
		if err := operation(); err != nil {
			return err
		}
	}

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
		// 对于Serializable隔离级别，即使从缓存读取也要记录到读集
		if tx.Options.IsolationLevel == Serializable {
			tx.ReadSet[cacheKey] = true
		}
		return record, nil
	}

	// 2. 缓存中没有，根据隔离级别选择读取方式
	// 生成主键键值
	fieldsBytes := tx.Table.FieldsToBytes(fields)
	defer func() {
		if fieldsBytes != nil && *fieldsBytes != nil {
			// 注意：这里我们不能直接使用engine.GlobalFieldsBytesPool.Put，因为它是未导出的
			// 我们需要使用其他方式管理内存
		}
	}()
	pkKey := tx.Table.GetPrimaryKey().JoinValue(fieldsBytes, tx.Table.GetId())
	pkKeyStr := string(pkKey)

	// 对于Serializable隔离级别，记录读集
	if tx.Options.IsolationLevel == Serializable {
		tx.ReadSet[pkKeyStr] = true
		tx.ReadSet[cacheKey] = true
	}

	// 根据隔离级别选择读取方式
	var value []byte
	var err error
	switch tx.Options.IsolationLevel {
	case ReadUncommitted:
		// 对于ReadUncommitted，尝试读取其他事务的未提交数据
		// 这里简化实现，直接使用原始存储，实际生产环境中需要更复杂的实现
		value, err = tx.OriginalStore.Get(pkKey)
	case ReadCommitted:
		// 对于ReadCommitted，每次读取都使用原始存储
		value, err = tx.OriginalStore.Get(pkKey)
	case RepeatableRead, Serializable:
		// 对于RepeatableRead和Serializable，使用事务开始时创建的快照
		if tx.Snapshot != nil {
			value, err = tx.Snapshot.Get(pkKey)
		} else {
			value, err = tx.OriginalStore.Get(pkKey)
		}
	default:
		// 默认使用原始存储读取
		value, err = tx.OriginalStore.Get(pkKey)
	}

	// 解密数据
	if err == nil && tx.Encryption != nil {
		decryptedValue, decryptErr := tx.Encryption.Decrypt(value)
		if decryptErr == nil {
			value = decryptedValue
		}
	}

	return value, err
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

// CreateSavepoint 创建保存点
func (tx *TableTransaction) CreateSavepoint(name string) error {
	if tx.Committed {
		return fmt.Errorf("transaction already committed")
	}

	// 初始化保存点映射（如果不存在）
	if tx.savepoints == nil {
		tx.savepoints = make(map[string]*Savepoint)
	}

	// 创建保存点，复制当前缓存和写集状态
	cacheCopy := make(map[string][]byte)
	for k, v := range tx.Cache {
		cacheCopy[k] = v
	}

	writeSetCopy := make(map[string]bool)
	for k, v := range tx.WriteSet {
		writeSetCopy[k] = v
	}

	// 创建并存储保存点
	savepoint := &Savepoint{
		name:     name,
		cache:    cacheCopy,
		writeSet: writeSetCopy,
	}
	tx.savepoints[name] = savepoint

	return nil
}

// RollbackToSavepoint 回滚到保存点
func (tx *TableTransaction) RollbackToSavepoint(name string) error {
	if tx.Committed {
		return fmt.Errorf("transaction already committed")
	}

	// 检查保存点是否存在
	savepoint, exists := tx.savepoints[name]
	if !exists {
		return fmt.Errorf("savepoint not found")
	}

	// 恢复缓存和写集状态
	tx.Cache = savepoint.cache
	tx.WriteSet = savepoint.writeSet

	// 重置Batch，确保回滚后不会提交之前的操作
	tx.Batch.Reset()

	// 移除保存点之后创建的保存点
	for spName := range tx.savepoints {
		if spName > name {
			delete(tx.savepoints, spName)
		}
	}

	return nil
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
		Table:          tx.Table,
		Batch:          tx.Batch, // 共享父事务的batch
		Committed:      false,
		Snapshot:       tx.Snapshot, // 共享父事务的快照
		OriginalStore:  tx.OriginalStore,
		Cache:          make(map[string][]byte), // 新的缓存
		Options:        tx.Options,
		Parent:         tx,
		TxID:           uint64(time.Now().UnixNano()),
		StartTime:      time.Now(),
		VersionManager: NewLockFreeVersionManager(), // 新的版本管理器
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
		// 对于Serializable隔离级别，执行冲突检测
		if tx.Options.IsolationLevel == Serializable {
			// 检查读集与其他事务的写集是否冲突
			for key := range tx.ReadSet {
				// 获取当前版本号
				currentVersion := tx.VersionManager.GetVersion(key)
				// 检查是否有冲突
				if tx.VersionManager.CheckConflict(key, currentVersion) {
					// 释放快照资源
					if tx.Snapshot != nil {
						tx.Snapshot.Release()
						tx.Snapshot = nil
					}
					return fmt.Errorf("可串行化冲突，读取的数据已被其他事务修改，请重试")
				}
			}
		}

		// 1. 提交批量操作
		// 使用原始存储执行写操作，支持重试
		err := tx.ExecuteWithRetry()

		if err != nil {
			// 提交失败，释放快照资源
			if tx.Snapshot != nil {
				tx.Snapshot.Release()
				tx.Snapshot = nil
			}
			return err
		}

		// 2. 对于Serializable隔离级别，更新版本号
		if tx.Options.IsolationLevel == Serializable {
			for key := range tx.WriteSet {
				tx.VersionManager.IncrementVersion(key)
			}
		}

		// 3. 写入事务提交日志
		if tx.WAL != nil {
			commitRecord := wal.LogRecord{
				Type:      wal.LogTypeCommit,
				TxID:      tx.TxID,
				Timestamp: time.Now().UnixNano(),
			}
			tx.WAL.WriteLog(commitRecord)
		}

		// 4. 释放快照资源
		if tx.Snapshot != nil {
			tx.Snapshot.Release()
			tx.Snapshot = nil
		}
	}

	// 5. 标记事务已结束，清空缓存
	tx.Committed = true
	tx.Cache = nil
	tx.ReadSet = nil
	tx.WriteSet = nil
	tx.Children = nil
	tx.savepoints = nil

	// 6. 归还事务对象到池中
	if tx.Parent == nil {
		// 只有根事务才归还到池，子事务由父事务管理
		GlobalTableTransactionPool.Put(tx)
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
	// 暂时默认所有错误都可重试
	return true
}

// Rollback 回滚表事务
func (tx *TableTransaction) Rollback() error {
	if err := tx.CheckCommitted(); err != nil {
		return err
	}

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
		// 1. 写入事务回滚日志
		if tx.WAL != nil {
			rollbackRecord := wal.LogRecord{
				Type:      wal.LogTypeRollback,
				TxID:      tx.TxID,
				Timestamp: time.Now().UnixNano(),
			}
			tx.WAL.WriteLog(rollbackRecord)
		}

		// 2. 释放快照资源
		if tx.Snapshot != nil {
			tx.Snapshot.Release()
			tx.Snapshot = nil
		}
	}

	// 3. 标记事务已结束，清空缓存
	tx.Committed = true
	tx.Cache = nil
	tx.ReadSet = nil
	tx.WriteSet = nil
	tx.Children = nil
	tx.savepoints = nil

	// 4. 归还事务对象到池中
	if tx.Parent == nil {
		// 只有根事务才归还到池，子事务由父事务管理
		GlobalTableTransactionPool.Put(tx)
	}

	return nil
}

// TransactionManager 事务管理器
// 用于管理多个表的事务，确保它们在同一个batch中执行，保证原子性
type TransactionManager struct {
	batch      storage.Batch
	tableTxMap map[*engine.Table]*TableTransaction
	savepoints map[string]*Savepoint
	parent     *TransactionManager
	children   []*TransactionManager
	committed  bool
	rolledBack bool
	options    *TransactionOptions
	txID       uint64
	startTime  time.Time
	userID     string // 用户ID，用于权限检查
}

// NewTransactionManager 创建事务管理器
func NewTransactionManager(batch storage.Batch, userID string) *TransactionManager {
	return NewTransactionManagerWithOptions(batch, DefaultTransactionOptions(), userID)
}

// NewTransactionManagerWithOptions 使用指定选项创建事务管理器
func NewTransactionManagerWithOptions(batch storage.Batch, options *TransactionOptions, userID string) *TransactionManager {
	if options == nil {
		options = DefaultTransactionOptions()
	}
	return &TransactionManager{
		batch:      batch,
		tableTxMap: make(map[*engine.Table]*TableTransaction),
		savepoints: make(map[string]*Savepoint),
		children:   make([]*TransactionManager, 0),
		committed:  false,
		rolledBack: false,
		options:    options,
		txID:       uint64(time.Now().UnixNano()),
		startTime:  time.Now(),
		userID:     userID,
	}
}

// AddTable 添加表到事务管理器
func (m *TransactionManager) AddTable(table *engine.Table) (*TableTransaction, error) {
	if m.committed || m.rolledBack {
		return nil, fmt.Errorf("transaction manager already completed")
	}

	// 检查表是否已经添加
	if _, exists := m.tableTxMap[table]; exists {
		return nil, fmt.Errorf("table already added to transaction manager")
	}

	// 创建表事务
	tx, err := NewTableTransactionWithBatchAndOptions(table, m.batch, m.options, m.userID)
	if err != nil {
		return nil, err
	}

	// 添加到映射
	m.tableTxMap[table] = tx

	return tx, nil
}

// GetTableTransaction 获取指定表的事务
func (m *TransactionManager) GetTableTransaction(table *engine.Table) (*TableTransaction, error) {
	if m.committed || m.rolledBack {
		return nil, fmt.Errorf("transaction manager already completed")
	}

	tx, exists := m.tableTxMap[table]
	if !exists {
		return nil, fmt.Errorf("table not found in transaction manager")
	}
	return tx, nil
}

// BeginNested 创建嵌套事务
func (m *TransactionManager) BeginNested() (*TransactionManager, error) {
	if m.committed || m.rolledBack {
		return nil, fmt.Errorf("transaction manager already completed")
	}

	if !m.options.AllowNested {
		return nil, fmt.Errorf("nested transactions are not allowed")
	}

	// 创建嵌套事务管理器
	nested := &TransactionManager{
		batch:      m.batch, // 共享父事务的batch
		tableTxMap: make(map[*engine.Table]*TableTransaction),
		savepoints: make(map[string]*Savepoint),
		parent:     m,
		children:   make([]*TransactionManager, 0),
		committed:  false,
		rolledBack: false,
		options:    m.options,
		txID:       uint64(time.Now().UnixNano()),
		startTime:  time.Now(),
		userID:     m.userID, // 继承父事务的用户ID
	}

	// 添加到父事务的子事务列表
	m.children = append(m.children, nested)

	return nested, nil
}

// CreateSavepoint 创建事务保存点
func (m *TransactionManager) CreateSavepoint(name string) error {
	if m.committed || m.rolledBack {
		return fmt.Errorf("transaction manager already completed")
	}

	// 为每个表事务创建保存点
	for _, tx := range m.tableTxMap {
		if err := tx.CreateSavepoint(name); err != nil {
			return err
		}
	}

	// 记录保存点
	m.savepoints[name] = &Savepoint{
		name:     name,
		cache:    nil,
		writeSet: nil,
	}

	return nil
}

// RollbackToSavepoint 回滚到保存点
func (m *TransactionManager) RollbackToSavepoint(name string) error {
	if m.committed || m.rolledBack {
		return fmt.Errorf("transaction manager already completed")
	}

	// 检查保存点是否存在
	if _, exists := m.savepoints[name]; !exists {
		return fmt.Errorf("savepoint not found")
	}

	// 为每个表事务回滚到保存点
	for _, tx := range m.tableTxMap {
		if err := tx.RollbackToSavepoint(name); err != nil {
			return err
		}
	}

	// 移除保存点之后创建的保存点
	for spName := range m.savepoints {
		if spName > name {
			delete(m.savepoints, spName)
		}
	}

	return nil
}

// Commit 提交所有事务
func (m *TransactionManager) Commit() error {
	if m.committed || m.rolledBack {
		return fmt.Errorf("transaction manager already completed")
	}

	// 提交所有子事务
	for _, child := range m.children {
		if !child.committed && !child.rolledBack {
			if err := child.Commit(); err != nil {
				return err
			}
		}
	}

	// 只有根事务才真正提交batch
	if m.parent == nil {
		// 提交所有表事务
		for _, tableTx := range m.tableTxMap {
			if err := tableTx.Commit(); err != nil {
				return err
			}
		}
	}

	// 标记为已提交
	m.committed = true

	return nil
}

// Rollback 回滚所有事务
func (m *TransactionManager) Rollback() error {
	if m.committed || m.rolledBack {
		return fmt.Errorf("transaction manager already completed")
	}

	// 回滚所有子事务
	for _, child := range m.children {
		if !child.committed && !child.rolledBack {
			if err := child.Rollback(); err != nil {
				return err
			}
		}
	}

	// 只有根事务才释放资源
	if m.parent == nil {
		// 回滚所有表事务
		for _, tx := range m.tableTxMap {
			if err := tx.Rollback(); err != nil {
				return err
			}
		}
	}

	// 标记为已回滚
	m.rolledBack = true

	return nil
}

// GetTxID 获取事务ID
func (m *TransactionManager) GetTxID() uint64 {
	return m.txID
}

// GetOptions 获取事务选项
func (m *TransactionManager) GetOptions() *TransactionOptions {
	return m.options
}

// IsActive 检查事务是否活跃
func (m *TransactionManager) IsActive() bool {
	return !m.committed && !m.rolledBack
}

// GetElapsedTime 获取事务执行时间
func (m *TransactionManager) GetElapsedTime() time.Duration {
	return time.Since(m.startTime)
}

// WithTransaction 便捷函数，用于执行多表事务
func WithTransaction(batch storage.Batch, tables []*engine.Table, userID string, fn func(txs map[*engine.Table]TableTransactionInterface) error) error {
	// 创建事务管理器
	manager := NewTransactionManager(batch, userID)

	// 添加所有表
	txs := make(map[*engine.Table]TableTransactionInterface)
	for _, table := range tables {
		tx, err := manager.AddTable(table)
		if err != nil {
			return err
		}
		txs[table] = tx
	}

	// 执行函数
	err := fn(txs)
	if err != nil {
		// 执行失败，回滚事务
		manager.Rollback()
		return err
	}

	// 执行成功，提交事务
	return manager.Commit()
}
