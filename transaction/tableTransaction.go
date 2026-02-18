package transaction

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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
	tx.versionCache = make(map[string]uint64)
	tx.parent = nil
	tx.children = nil
	tx.operationCount = 0
	tx.conflictCount = 0
	tx.auditEventCount = 0
	tx.readOperationCount = 0
	tx.writeOperationCount = 0
	tx.currentBatchSize = 100
	tx.maxBatchSize = 1000
	tx.minBatchSize = 10
	tx.batchThreshold = 80
	tx.timeoutDuration = 0
	tx.timeoutTimer = nil
	tx.timedOut = false
	// 其他字段在使用时会被覆盖，不需要重置

	p.pool.Put(tx)
}

// TableTransaction 实现Transaction接口的具体结构体
type TableTransaction struct {
	table         TableAccessor    // 关联的表访问器
	batch         storage.Batch    // 事务使用的batch，原子性
	committed     bool             // 是否已提交，提交成功后则是持久性。
	snapshot      storage.Snapshot // 事务使用的快照，一致性。
	originalStore storage.Store    // 原始存储，用于写操作
	// 事务内修改缓存，用于读取自己的写操作
	// key: 主键值的字符串表示，value: 记录的字节数组
	cache map[string][]byte // 事务内修改缓存 //隔离性
	// 锁键缓存，避免重复生成锁键
	lockKeyCache map[string]string // 锁键缓存
	// 版本缓存，用于乐观并发控制
	versionCache map[string]uint64 // 版本缓存
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
	// 对于Serializable隔离级别，使用的锁集合
	readLocks  map[string]bool // 读锁
	writeLocks map[string]bool // 写锁
	// 权限映射，用于快速权限检查
	permissionMap map[string]bool
	// 加密上下文
	encryptionContext *EncryptionContext
	// 事务日志相关
	transactionLog     *os.File
	transactionLogPath string
	auditLogPath       string
	lastBackupTime     time.Time
	lastAuditTime      time.Time
	// 事务状态监控
	operationCount      int // 操作计数
	conflictCount       int // 冲突计数
	auditEventCount     int // 审计事件计数
	readOperationCount  int // 读操作计数
	writeOperationCount int // 写操作计数
	// 批量大小控制
	currentBatchSize int // 当前批量大小
	maxBatchSize     int // 最大批量大小
	minBatchSize     int // 最小批量大小
	batchThreshold   int // 批量提交阈值
	// 超时控制
	timeoutDuration time.Duration // 事务超时时间
	timeoutTimer    *time.Timer   // 超时定时器
	timedOut        bool          // 是否已超时
}

// EncryptionContext 加密上下文
type EncryptionContext struct {
	key    []byte
	cipher cipher.Block
}

// TransactionLogEntry 事务日志条目
type TransactionLogEntry struct {
	TxID         uint64    `json:"txID"`
	Timestamp    time.Time `json:"timestamp"`
	Operation    string    `json:"operation"`
	Table        string    `json:"table"`
	Fields       any       `json:"fields"`
	Status       string    `json:"status"`
	ErrorMessage string    `json:"errorMessage,omitempty"`
}

// AuditLogEntry 审计日志条目
type AuditLogEntry struct {
	TxID        uint64    `json:"txID"`
	Timestamp   time.Time `json:"timestamp"`
	Operation   string    `json:"operation"`
	Table       string    `json:"table"`
	Fields      any       `json:"fields"`
	Status      string    `json:"status"`
	User        string    `json:"user,omitempty"`
	IPAddress   string    `json:"ipAddress,omitempty"`
	ActionType  string    `json:"actionType"` // READ, WRITE, ADMIN
	Resource    string    `json:"resource"`
	Permissions []string  `json:"permissions"`
	Error       string    `json:"error,omitempty"`
}

// NewTransactionError 创建一个新的事务错误
func NewTransactionError(errorType, message, operation, table string, txID uint64, isRetryable bool, cause error) *TransactionError {
	return &TransactionError{
		Type:        errorType,
		Message:     message,
		Operation:   operation,
		Table:       table,
		TxID:        txID,
		IsRetryable: isRetryable,
		Cause:       cause,
	}
}

// 全局锁管理器，用于Serializable隔离级别
var globalLockManager = &LockManager{
	locks: make(map[string]int),
	mutex: &sync.Mutex{},
}

// LockManager 锁管理器
type LockManager struct {
	locks map[string]int // 锁计数器，正数表示读锁数量，-1表示写锁
	mutex *sync.Mutex
}

// AcquireReadLock 获取读锁
func (lm *LockManager) AcquireReadLock(key string) bool {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	// 如果有写锁，无法获取读锁
	if lm.locks[key] == -1 {
		return false
	}

	// 增加读锁计数
	lm.locks[key]++
	return true
}

// ReleaseReadLock 释放读锁
func (lm *LockManager) ReleaseReadLock(key string) {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	// 减少读锁计数
	if lm.locks[key] > 0 {
		lm.locks[key]--
		// 如果读锁计数为0，删除锁条目
		if lm.locks[key] == 0 {
			delete(lm.locks, key)
		}
	}
}

// AcquireWriteLock 获取写锁
func (lm *LockManager) AcquireWriteLock(key string) bool {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	// 如果有任何锁（读锁或写锁），无法获取写锁
	if lm.locks[key] != 0 {
		return false
	}

	// 设置写锁
	lm.locks[key] = -1
	return true
}

// ReleaseWriteLock 释放写锁
func (lm *LockManager) ReleaseWriteLock(key string) {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	// 释放写锁
	if lm.locks[key] == -1 {
		delete(lm.locks, key)
	}
}

// NewTableTransaction 创建一个新的表事务
func NewTableTransaction(table TableAccessor, batch storage.Batch, options *TransactionOptions) (*TableTransaction, error) {
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
		if levelDBStore, ok := table.GetKVStore().(*storage.LevelDBStore); ok {
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
	tx.table = table
	tx.batch = batch
	tx.committed = false
	tx.snapshot = snapshot
	tx.originalStore = table.GetKVStore()
	tx.cache = make(map[string][]byte)        // 初始化事务内缓存
	tx.lockKeyCache = make(map[string]string) // 初始化锁键缓存
	tx.versionCache = make(map[string]uint64) // 初始化版本缓存

	// 为Serializable隔离级别初始化锁映射
	if options.IsolationLevel == Serializable {
		tx.readLocks = make(map[string]bool)
		tx.writeLocks = make(map[string]bool)
	}

	// 初始化权限映射
	tx.permissionMap = make(map[string]bool)
	for _, perm := range options.Permissions {
		tx.permissionMap[perm] = true
	}

	// 初始化监控字段
	tx.operationCount = 0
	tx.conflictCount = 0
	tx.auditEventCount = 0
	tx.readOperationCount = 0
	tx.writeOperationCount = 0
	tx.lastAuditTime = time.Now()

	// 设置事务选项
	tx.options = options
	tx.txID = txID
	tx.startTime = time.Now()
	tx.lastBackupTime = time.Now()

	// 初始化批量大小控制
	tx.currentBatchSize = 100 // 默认批量大小
	tx.maxBatchSize = 1000    // 最大批量大小
	tx.minBatchSize = 10      // 最小批量大小
	tx.batchThreshold = 80    // 批量提交阈值

	// 初始化超时控制
	tx.timeoutDuration = tx.options.Timeout
	tx.timedOut = false

	// 如果设置了超时时间，启动超时定时器
	if tx.timeoutDuration > 0 {
		tx.timeoutTimer = time.AfterFunc(tx.timeoutDuration, func() {
			tx.timedOut = true
			// 这里可以添加超时处理逻辑，例如自动回滚事务
		})
	}

	// 初始化加密上下文
	if options.EnableEncryption && options.EncryptionKeyID != "" {
		// 这里使用示例密钥，实际应用中应该从密钥管理服务获取
		key := []byte(options.EncryptionKeyID[:32]) // 截取前32字节作为密钥
		cipher, err := aes.NewCipher(key)
		if err == nil {
			tx.encryptionContext = &EncryptionContext{
				key:    key,
				cipher: cipher,
			}
		}
	}

	// 初始化事务日志
	if options.EnableTransactionLog {
		// 确保事务日志目录存在
		logPath := options.TransactionLogPath
		if logPath == "" {
			logPath = "./transaction_logs"
		}
		err := os.MkdirAll(logPath, 0755)
		if err == nil {
			// 创建或打开事务日志文件
			logFile := filepath.Join(logPath, fmt.Sprintf("tx_%d.log", txID))
			tx.transactionLogPath = logPath
			file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err == nil {
				tx.transactionLog = file
				// 写入事务开始日志
				entry := TransactionLogEntry{
					TxID:      txID,
					Timestamp: time.Now(),
					Operation: "BEGIN",
					Table:     table.GetName(),
					Status:    "STARTED",
				}
				tx.writeTransactionLog(entry)

				// 初始化审计日志路径
				tx.auditLogPath = filepath.Join(logPath, "audit")
				err = os.MkdirAll(tx.auditLogPath, 0755)
				if err != nil {
					tx.auditLogPath = ""
				}
			}
		}
	}

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
		versionCache:  make(map[string]uint64), // 新的版本缓存
		options:       tx.options,
		parent:        tx,
		txID:          uint64(time.Now().UnixNano()),
		startTime:     time.Now(),
		// 初始化监控字段
		operationCount:      0,
		conflictCount:       0,
		auditEventCount:     0,
		readOperationCount:  0,
		writeOperationCount: 0,
		lastAuditTime:       time.Now(),
		// 初始化批量大小控制
		currentBatchSize: 100,
		maxBatchSize:     1000,
		minBatchSize:     10,
		batchThreshold:   80,
		// 初始化超时控制
		timeoutDuration: tx.timeoutDuration,
		timedOut:        false,
	}

	// 初始化权限映射，继承父事务的权限
	nestedTx.permissionMap = make(map[string]bool)
	for perm := range tx.permissionMap {
		nestedTx.permissionMap[perm] = true
	}

	// 为Serializable隔离级别初始化锁映射
	if tx.options.IsolationLevel == Serializable {
		nestedTx.readLocks = make(map[string]bool)
		nestedTx.writeLocks = make(map[string]bool)
	}

	// 继承加密上下文
	nestedTx.encryptionContext = tx.encryptionContext

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

// checkTimeout 检查事务是否已超时
func (tx *TableTransaction) checkTimeout() error {
	if tx.timedOut {
		err := NewTransactionError(
			ErrorTypeTimeout,
			"transaction timeout",
			"OPERATION",
			tx.table.GetName(),
			tx.txID,
			false,
			nil,
		)
		return err
	}
	return nil
}

// checkTransactionState 检查事务状态（是否已提交或超时）
func (tx *TableTransaction) checkTransactionState() error {
	if err := tx.checkCommitted(); err != nil {
		return err
	}
	if err := tx.checkTimeout(); err != nil {
		return err
	}
	return nil
}

// Insert 在事务中插入记录
func (tx *TableTransaction) Insert(fields *map[string]any) (int, error) {
	if err := tx.checkTransactionState(); err != nil {
		return 0, err
	}

	// 检查插入权限
	if !tx.CheckPermission("INSERT") {
		err := NewTransactionError(
			ErrorTypePermission,
			"permission denied",
			"INSERT",
			tx.table.GetName(),
			tx.txID,
			false,
			nil,
		)
		return 0, err
	}

	// 生成缓存键
	cacheKey := tx.getCacheKey(fields)

	// 对于Serializable隔离级别，获取写锁
	if tx.options.IsolationLevel == Serializable {
		// 检查是否已经持有写锁
		if !tx.writeLocks[cacheKey] {
			// 尝试获取写锁
			if !globalLockManager.AcquireWriteLock(cacheKey) {
				err := NewTransactionError(
					ErrorTypeLock,
					fmt.Sprintf("failed to acquire write lock for key: %s", cacheKey),
					"INSERT",
					tx.table.GetName(),
					tx.txID,
					true,
					nil,
				)
				return 0, err
			}
			// 记录已获取的写锁
			tx.writeLocks[cacheKey] = true
		}
	}

	// 执行插入操作
	id, err := tx.table.Insert(fields, tx.batch)
	if err != nil {
		// 写入失败日志
		entry := TransactionLogEntry{
			TxID:         tx.txID,
			Timestamp:    time.Now(),
			Operation:    "INSERT",
			Table:        tx.table.GetName(),
			Fields:       fields,
			Status:       "FAILED",
			ErrorMessage: err.Error(),
		}
		tx.writeTransactionLog(entry)

		// 写入审计日志
		auditEntry := AuditLogEntry{
			TxID:        tx.txID,
			Timestamp:   time.Now(),
			Operation:   "INSERT",
			Table:       tx.table.GetName(),
			Fields:      fields,
			Status:      "FAILED",
			ActionType:  "WRITE",
			Resource:    fmt.Sprintf("%s:%s", tx.table.GetName(), tx.getCacheKey(fields)),
			Permissions: tx.options.Permissions,
			Error:       err.Error(),
		}
		tx.writeAuditLog(auditEntry)

		// 创建事务错误
		txErr := NewTransactionError(
			ErrorTypeStorage,
			fmt.Sprintf("failed to insert record: %v", err),
			"INSERT",
			tx.table.GetName(),
			tx.txID,
			true,
			err,
		)
		return 0, txErr
	}

	// 更新记录版本（乐观并发控制）
	if err := tx.updateRecordVersion(cacheKey); err != nil {
		// 写入失败日志
		entry := TransactionLogEntry{
			TxID:         tx.txID,
			Timestamp:    time.Now(),
			Operation:    "INSERT",
			Table:        tx.table.GetName(),
			Fields:       fields,
			Status:       "FAILED",
			ErrorMessage: fmt.Sprintf("failed to update version: %v", err),
		}
		tx.writeTransactionLog(entry)

		// 创建事务错误
		txErr := NewTransactionError(
			ErrorTypeStorage,
			fmt.Sprintf("failed to update record version: %v", err),
			"INSERT",
			tx.table.GetName(),
			tx.txID,
			true,
			err,
		)
		return 0, txErr
	}

	// 写入成功日志
	entry := TransactionLogEntry{
		TxID:      tx.txID,
		Timestamp: time.Now(),
		Operation: "INSERT",
		Table:     tx.table.GetName(),
		Fields:    fields,
		Status:    "SUCCESS",
	}
	tx.writeTransactionLog(entry)

	// 写入审计日志
	auditEntry := AuditLogEntry{
		TxID:        tx.txID,
		Timestamp:   time.Now(),
		Operation:   "INSERT",
		Table:       tx.table.GetName(),
		Fields:      fields,
		Status:      "SUCCESS",
		ActionType:  "WRITE",
		Resource:    fmt.Sprintf("%s:%s", tx.table.GetName(), cacheKey),
		Permissions: tx.options.Permissions,
	}
	tx.writeAuditLog(auditEntry)

	// 将fields转换为fieldsBytes，用于生成主键和记录
	fieldsBytes := tx.table.FieldsToBytes(fields)

	// 生成记录字节数组
	record := tx.table.FormatRecord(fieldsBytes)

	// 如果启用了加密，加密记录
	if tx.options.EnableEncryption {
		encryptedRecord, err := tx.EncryptData(record)
		if err != nil {
			txErr := NewTransactionError(
				ErrorTypeEncryption,
				fmt.Sprintf("failed to encrypt data: %v", err),
				"INSERT",
				tx.table.GetName(),
				tx.txID,
				false,
				err,
			)
			return 0, txErr
		}
		record = encryptedRecord
	}

	// 将记录存入缓存，用于读取自己的写操作
	tx.cache[cacheKey] = record

	// 缓存锁键
	tx.lockKeyCache[cacheKey] = tx.table.GenerateLockKey(fields)

	// 增加操作计数
	tx.operationCount++
	tx.writeOperationCount++

	// 调整批量大小
	tx.adjustBatchSize()

	return id, nil
}

// Update 在事务中更新记录
func (tx *TableTransaction) Update(fields *map[string]any) error {
	if err := tx.checkTransactionState(); err != nil {
		return err
	}

	// 检查更新权限
	if !tx.CheckPermission("UPDATE") {
		err := NewTransactionError(
			ErrorTypePermission,
			"permission denied",
			"UPDATE",
			tx.table.GetName(),
			tx.txID,
			false,
			nil,
		)
		return err
	}

	// 生成缓存键
	cacheKey := tx.getCacheKey(fields)

	// 对于Serializable隔离级别，获取写锁
	if tx.options.IsolationLevel == Serializable {
		// 检查是否已经持有写锁
		if !tx.writeLocks[cacheKey] {
			// 尝试获取写锁
			if !globalLockManager.AcquireWriteLock(cacheKey) {
				err := NewTransactionError(
					ErrorTypeLock,
					fmt.Sprintf("failed to acquire write lock for key: %s", cacheKey),
					"UPDATE",
					tx.table.GetName(),
					tx.txID,
					true,
					nil,
				)
				return err
			}
			// 记录已获取的写锁
			tx.writeLocks[cacheKey] = true
		}
	}

	// 乐观并发控制：检查记录版本
	_, err := tx.getRecordVersion(cacheKey)
	if err != nil {
		// 写入失败日志
		entry := TransactionLogEntry{
			TxID:         tx.txID,
			Timestamp:    time.Now(),
			Operation:    "UPDATE",
			Table:        tx.table.GetName(),
			Fields:       fields,
			Status:       "FAILED",
			ErrorMessage: fmt.Sprintf("failed to get record version: %v", err),
		}
		tx.writeTransactionLog(entry)

		// 创建事务错误
		txErr := NewTransactionError(
			ErrorTypeStorage,
			fmt.Sprintf("failed to get record version: %v", err),
			"UPDATE",
			tx.table.GetName(),
			tx.txID,
			true,
			err,
		)
		return txErr
	}

	// 检查缓存中是否有要更新的记录
	if record, exists := tx.cache[cacheKey]; exists {
		// 缓存中有记录，使用提取的方法处理
		if err := tx.updateFromCache(fields, cacheKey, record); err != nil {
			// 写入失败日志
			entry := TransactionLogEntry{
				TxID:         tx.txID,
				Timestamp:    time.Now(),
				Operation:    "UPDATE",
				Table:        tx.table.GetName(),
				Fields:       fields,
				Status:       "FAILED",
				ErrorMessage: err.Error(),
			}
			tx.writeTransactionLog(entry)

			// 创建事务错误
			txErr := NewTransactionError(
				ErrorTypeStorage,
				fmt.Sprintf("failed to update record from cache: %v", err),
				"UPDATE",
				tx.table.GetName(),
				tx.txID,
				true,
				err,
			)
			return txErr
		}
	} else {
		// 缓存中没有记录，直接执行更新操作
		err := tx.table.Update(fields, tx.batch)
		if err != nil {
			// 写入失败日志
			entry := TransactionLogEntry{
				TxID:         tx.txID,
				Timestamp:    time.Now(),
				Operation:    "UPDATE",
				Table:        tx.table.GetName(),
				Fields:       fields,
				Status:       "FAILED",
				ErrorMessage: err.Error(),
			}
			tx.writeTransactionLog(entry)

			// 写入审计日志
			auditEntry := AuditLogEntry{
				TxID:        tx.txID,
				Timestamp:   time.Now(),
				Operation:   "UPDATE",
				Table:       tx.table.GetName(),
				Fields:      fields,
				Status:      "FAILED",
				ActionType:  "WRITE",
				Resource:    fmt.Sprintf("%s:%s", tx.table.GetName(), cacheKey),
				Permissions: tx.options.Permissions,
				Error:       err.Error(),
			}
			tx.writeAuditLog(auditEntry)

			// 创建事务错误
			txErr := NewTransactionError(
				ErrorTypeStorage,
				fmt.Sprintf("failed to update record: %v", err),
				"UPDATE",
				tx.table.GetName(),
				tx.txID,
				true,
				err,
			)
			return txErr
		}

		// 从缓存中删除旧记录，强制后续读取从数据库获取最新值
		delete(tx.cache, cacheKey)
	}

	// 更新记录版本（乐观并发控制）
	if err := tx.updateRecordVersion(cacheKey); err != nil {
		// 写入失败日志
		entry := TransactionLogEntry{
			TxID:         tx.txID,
			Timestamp:    time.Now(),
			Operation:    "UPDATE",
			Table:        tx.table.GetName(),
			Fields:       fields,
			Status:       "FAILED",
			ErrorMessage: fmt.Sprintf("failed to update version: %v", err),
		}
		tx.writeTransactionLog(entry)

		// 创建事务错误
		txErr := NewTransactionError(
			ErrorTypeStorage,
			fmt.Sprintf("failed to update record version: %v", err),
			"UPDATE",
			tx.table.GetName(),
			tx.txID,
			true,
			err,
		)
		return txErr
	}

	// 写入成功日志
	entry := TransactionLogEntry{
		TxID:      tx.txID,
		Timestamp: time.Now(),
		Operation: "UPDATE",
		Table:     tx.table.GetName(),
		Fields:    fields,
		Status:    "SUCCESS",
	}
	tx.writeTransactionLog(entry)

	// 写入审计日志
	auditEntry := AuditLogEntry{
		TxID:        tx.txID,
		Timestamp:   time.Now(),
		Operation:   "UPDATE",
		Table:       tx.table.GetName(),
		Fields:      fields,
		Status:      "SUCCESS",
		ActionType:  "WRITE",
		Resource:    fmt.Sprintf("%s:%s", tx.table.GetName(), cacheKey),
		Permissions: tx.options.Permissions,
	}
	tx.writeAuditLog(auditEntry)

	// 缓存锁键
	tx.lockKeyCache[cacheKey] = tx.table.GenerateLockKey(fields)

	// 增加操作计数
	tx.operationCount++
	tx.writeOperationCount++

	// 调整批量大小
	tx.adjustBatchSize()

	return nil
}

// updateFromCache 从缓存中更新记录
func (tx *TableTransaction) updateFromCache(fields *map[string]any, cacheKey string, record []byte) error {
	// 1. 解析记录
	// 注意：这里需要根据实际的主键实现来解析记录
	// 由于我们使用了接口，这里简化处理

	// 2. 直接执行更新操作
	err := tx.table.Update(fields, tx.batch)
	if err != nil {
		return err
	}

	// 3. 生成新的记录
	fieldsBytes := tx.table.FieldsToBytes(fields)
	updatedRecord := tx.table.FormatRecord(fieldsBytes)

	// 4. 如果启用了加密，加密记录
	if tx.options.EnableEncryption {
		encryptedRecord, err := tx.EncryptData(updatedRecord)
		if err != nil {
			return fmt.Errorf("failed to encrypt data: %v", err)
		}
		updatedRecord = encryptedRecord
	}

	// 5. 更新缓存
	tx.cache[cacheKey] = updatedRecord

	return nil
}

// Delete 在事务中删除记录
func (tx *TableTransaction) Delete(fields *map[string]any) error {
	if err := tx.checkTransactionState(); err != nil {
		return err
	}

	// 检查删除权限
	if !tx.CheckPermission("DELETE") {
		err := NewTransactionError(
			ErrorTypePermission,
			"permission denied",
			"DELETE",
			tx.table.GetName(),
			tx.txID,
			false,
			nil,
		)
		return err
	}

	// 生成缓存键
	cacheKey := tx.getCacheKey(fields)

	// 对于Serializable隔离级别，获取写锁
	if tx.options.IsolationLevel == Serializable {
		// 检查是否已经持有写锁
		if !tx.writeLocks[cacheKey] {
			// 尝试获取写锁
			if !globalLockManager.AcquireWriteLock(cacheKey) {
				err := NewTransactionError(
					ErrorTypeLock,
					fmt.Sprintf("failed to acquire write lock for key: %s", cacheKey),
					"DELETE",
					tx.table.GetName(),
					tx.txID,
					true,
					nil,
				)
				return err
			}
			// 记录已获取的写锁
			tx.writeLocks[cacheKey] = true
		}
	}

	// 乐观并发控制：检查记录版本
	_, err := tx.getRecordVersion(cacheKey)
	if err != nil {
		// 写入失败日志
		entry := TransactionLogEntry{
			TxID:         tx.txID,
			Timestamp:    time.Now(),
			Operation:    "DELETE",
			Table:        tx.table.GetName(),
			Fields:       fields,
			Status:       "FAILED",
			ErrorMessage: fmt.Sprintf("failed to get record version: %v", err),
		}
		tx.writeTransactionLog(entry)

		// 创建事务错误
		txErr := NewTransactionError(
			ErrorTypeStorage,
			fmt.Sprintf("failed to get record version: %v", err),
			"DELETE",
			tx.table.GetName(),
			tx.txID,
			true,
			err,
		)
		return txErr
	}

	// 执行删除操作
	err = tx.table.Delete(fields, tx.batch)
	if err != nil {
		// 写入失败日志
		entry := TransactionLogEntry{
			TxID:         tx.txID,
			Timestamp:    time.Now(),
			Operation:    "DELETE",
			Table:        tx.table.GetName(),
			Fields:       fields,
			Status:       "FAILED",
			ErrorMessage: err.Error(),
		}
		tx.writeTransactionLog(entry)

		// 写入审计日志
		auditEntry := AuditLogEntry{
			TxID:        tx.txID,
			Timestamp:   time.Now(),
			Operation:   "DELETE",
			Table:       tx.table.GetName(),
			Fields:      fields,
			Status:      "FAILED",
			ActionType:  "WRITE",
			Resource:    fmt.Sprintf("%s:%s", tx.table.GetName(), cacheKey),
			Permissions: tx.options.Permissions,
			Error:       err.Error(),
		}
		tx.writeAuditLog(auditEntry)

		// 创建事务错误
		txErr := NewTransactionError(
			ErrorTypeStorage,
			fmt.Sprintf("failed to delete record: %v", err),
			"DELETE",
			tx.table.GetName(),
			tx.txID,
			true,
			err,
		)
		return txErr
	}

	// 删除版本记录
	versionKey := tx.getVersionKey(cacheKey)
	tx.batch.Delete([]byte(versionKey))

	// 写入成功日志
	entry := TransactionLogEntry{
		TxID:      tx.txID,
		Timestamp: time.Now(),
		Operation: "DELETE",
		Table:     tx.table.GetName(),
		Fields:    fields,
		Status:    "SUCCESS",
	}
	tx.writeTransactionLog(entry)

	// 写入审计日志
	auditEntry := AuditLogEntry{
		TxID:        tx.txID,
		Timestamp:   time.Now(),
		Operation:   "DELETE",
		Table:       tx.table.GetName(),
		Fields:      fields,
		Status:      "SUCCESS",
		ActionType:  "WRITE",
		Resource:    fmt.Sprintf("%s:%s", tx.table.GetName(), cacheKey),
		Permissions: tx.options.Permissions,
	}
	tx.writeAuditLog(auditEntry)

	// 从缓存中删除记录，确保读一致性
	delete(tx.cache, cacheKey)
	delete(tx.versionCache, cacheKey)

	// 缓存锁键
	tx.lockKeyCache[cacheKey] = tx.table.GenerateLockKey(fields)

	// 增加操作计数
	tx.operationCount++
	tx.writeOperationCount++

	// 调整批量大小
	tx.adjustBatchSize()

	return nil
}

// getCacheKey 生成缓存键
func (tx *TableTransaction) getCacheKey(fields *map[string]any) string {
	// 生成主键键值，用于缓存
	// 注意：这里需要根据实际的主键实现来生成键
	// 由于我们使用了接口，这里简化处理
	// 实际实现中，应该调用主键的JoinValue方法
	return fmt.Sprintf("%v", fields)
}

// getVersionKey 生成版本键
func (tx *TableTransaction) getVersionKey(cacheKey string) string {
	return fmt.Sprintf("version:%s", cacheKey)
}

// getRecordVersion 获取记录的版本
func (tx *TableTransaction) getRecordVersion(cacheKey string) (uint64, error) {
	// 先从版本缓存中获取
	if version, exists := tx.versionCache[cacheKey]; exists {
		return version, nil
	}

	// 从存储中获取版本
	versionKey := tx.getVersionKey(cacheKey)
	versionBytes, err := tx.originalStore.Get([]byte(versionKey))
	if err != nil {
		// 版本不存在，返回0
		return 0, nil
	}

	// 解析版本号
	var version uint64
	if _, err := fmt.Sscanf(string(versionBytes), "%d", &version); err != nil {
		return 0, nil
	}

	// 缓存版本号
	tx.versionCache[cacheKey] = version
	return version, nil
}

// updateRecordVersion 更新记录的版本
func (tx *TableTransaction) updateRecordVersion(cacheKey string) error {
	// 获取当前版本
	currentVersion, err := tx.getRecordVersion(cacheKey)
	if err != nil {
		return err
	}

	// 增加版本号
	newVersion := currentVersion + 1

	// 缓存新版本
	tx.versionCache[cacheKey] = newVersion

	// 写入存储
	versionKey := tx.getVersionKey(cacheKey)
	tx.batch.Put([]byte(versionKey), []byte(fmt.Sprintf("%d", newVersion)))

	return nil
}

// Read 在事务中读取单条记录（支持读一致性）
func (tx *TableTransaction) Read(fields *map[string]any) ([]byte, error) {
	if err := tx.checkTransactionState(); err != nil {
		return nil, err
	}

	// 检查读取权限
	if !tx.CheckPermission("READ") {
		err := NewTransactionError(
			ErrorTypePermission,
			"permission denied",
			"READ",
			tx.table.GetName(),
			tx.txID,
			false,
			nil,
		)
		return nil, err
	}

	// 生成缓存键
	cacheKey := tx.getCacheKey(fields)

	// 1. 优先从缓存中读取，支持读取自己的写操作
	if record, exists := tx.cache[cacheKey]; exists {
		// 如果启用了加密，解密记录
		if tx.options.EnableEncryption {
			decryptedRecord, err := tx.DecryptData(record)
			if err != nil {
				txErr := NewTransactionError(
					ErrorTypeEncryption,
					fmt.Sprintf("failed to decrypt data: %v", err),
					"READ",
					tx.table.GetName(),
					tx.txID,
					false,
					err,
				)
				return nil, txErr
			}
			// 增加操作计数
			tx.operationCount++
			tx.readOperationCount++

			return decryptedRecord, nil
		}

		// 增加操作计数
		tx.operationCount++
		tx.readOperationCount++

		return record, nil
	}

	// 2. 对于Serializable隔离级别，获取读锁
	if tx.options.IsolationLevel == Serializable {
		// 检查是否已经持有读锁
		if !tx.readLocks[cacheKey] {
			// 尝试获取读锁
			if !globalLockManager.AcquireReadLock(cacheKey) {
				err := NewTransactionError(
					ErrorTypeLock,
					fmt.Sprintf("failed to acquire read lock for key: %s", cacheKey),
					"READ",
					tx.table.GetName(),
					tx.txID,
					true,
					nil,
				)
				return nil, err
			}
			// 记录已获取的读锁
			tx.readLocks[cacheKey] = true
		}
	}

	// 3. 根据隔离级别选择不同的读取方式
	var record []byte
	var err error
	switch tx.options.IsolationLevel {
	case ReadUncommitted:
		// 读未提交：直接从原始存储读取，可能读取到未提交的数据
		record, err = tx.originalStore.Get([]byte(cacheKey))
	case ReadCommitted:
		// 读已提交：每次读取都创建新的快照
		if levelDBStore, ok := tx.originalStore.(*storage.LevelDBStore); ok {
			snapshot, err := levelDBStore.Snapshot()
			if err != nil {
				txErr := NewTransactionError(
					ErrorTypeStorage,
					fmt.Sprintf("failed to create snapshot: %v", err),
					"READ",
					tx.table.GetName(),
					tx.txID,
					true,
					err,
				)
				return nil, txErr
			}
			defer snapshot.Release()
			record, err = snapshot.Get([]byte(cacheKey))
		} else {
			record, err = tx.originalStore.Get([]byte(cacheKey))
		}
	case RepeatableRead, Serializable:
		// 可重复读和可序列化：使用事务开始时创建的快照
		if tx.snapshot != nil {
			record, err = tx.snapshot.Get([]byte(cacheKey))
		} else {
			record, err = tx.originalStore.Get([]byte(cacheKey))
		}
	default:
		// 默认使用读已提交
		record, err = tx.originalStore.Get([]byte(cacheKey))
	}

	if err != nil {
		// 写入审计日志
		auditEntry := AuditLogEntry{
			TxID:        tx.txID,
			Timestamp:   time.Now(),
			Operation:   "READ",
			Table:       tx.table.GetName(),
			Fields:      fields,
			Status:      "FAILED",
			ActionType:  "READ",
			Resource:    fmt.Sprintf("%s:%s", tx.table.GetName(), cacheKey),
			Permissions: tx.options.Permissions,
			Error:       err.Error(),
		}
		tx.writeAuditLog(auditEntry)

		txErr := NewTransactionError(
			ErrorTypeStorage,
			fmt.Sprintf("failed to read record: %v", err),
			"READ",
			tx.table.GetName(),
			tx.txID,
			true,
			err,
		)
		return nil, txErr
	}

	// 写入审计日志
	auditEntry := AuditLogEntry{
		TxID:        tx.txID,
		Timestamp:   time.Now(),
		Operation:   "READ",
		Table:       tx.table.GetName(),
		Fields:      fields,
		Status:      "SUCCESS",
		ActionType:  "READ",
		Resource:    fmt.Sprintf("%s:%s", tx.table.GetName(), cacheKey),
		Permissions: tx.options.Permissions,
	}
	tx.writeAuditLog(auditEntry)

	// 如果启用了加密，解密记录
	if tx.options.EnableEncryption {
		decryptedRecord, err := tx.DecryptData(record)
		if err != nil {
			txErr := NewTransactionError(
				ErrorTypeEncryption,
				fmt.Sprintf("failed to decrypt data: %v", err),
				"READ",
				tx.table.GetName(),
				tx.txID,
				false,
				err,
			)
			return nil, txErr
		}
		// 增加操作计数
		tx.operationCount++
		tx.readOperationCount++

		return decryptedRecord, nil
	}

	// 增加操作计数
	tx.operationCount++
	tx.readOperationCount++

	return record, nil
}

// Search 在事务中搜索记录（支持读一致性，即使用快照）
func (tx *TableTransaction) Search(fields *map[string]any, ops ...util.ComparisonOperator) (interface{}, error) {
	// 检查事务状态
	if err := tx.checkTransactionState(); err != nil {
		return nil, err
	}

	// 检查读取权限
	if !tx.CheckPermission("READ") {
		err := NewTransactionError(
			ErrorTypePermission,
			"permission denied",
			"SEARCH",
			tx.table.GetName(),
			tx.txID,
			false,
			nil,
		)
		return nil, err
	}

	// 检查fields参数是否为nil
	if fields == nil {
		err := NewTransactionError(
			ErrorTypeValidation,
			"fields cannot be nil",
			"SEARCH",
			tx.table.GetName(),
			tx.txID,
			false,
			nil,
		)
		return nil, err
	}

	// 创建一个函数，根据隔离级别选择不同的存储获取迭代器
	funIter := func(start, limit []byte) storage.Iterator {
		switch tx.options.IsolationLevel {
		case ReadUncommitted:
			// 读未提交：直接从原始存储读取
			return tx.originalStore.Iterator(start, limit)
		case ReadCommitted:
			// 读已提交：每次读取都创建新的快照
			if levelDBStore, ok := tx.originalStore.(*storage.LevelDBStore); ok {
				snapshot, err := levelDBStore.Snapshot()
				if err == nil {
					return snapshot.Iterator(start, limit)
				}
			}
			return tx.originalStore.Iterator(start, limit)
		case RepeatableRead, Serializable:
			// 可重复读和可序列化：使用事务开始时创建的快照
			if tx.snapshot != nil {
				return tx.snapshot.Iterator(start, limit)
			}
			return tx.originalStore.Iterator(start, limit)
		default:
			// 默认使用读已提交
			return tx.originalStore.Iterator(start, limit)
		}
	}

	// 调用table.Searchs方法，传入funIter函数
	result, err := tx.table.Searchs(funIter, fields, ops...)
	if err != nil {
		// 写入审计日志
		auditEntry := AuditLogEntry{
			TxID:        tx.txID,
			Timestamp:   time.Now(),
			Operation:   "SEARCH",
			Table:       tx.table.GetName(),
			Fields:      fields,
			Status:      "FAILED",
			ActionType:  "READ",
			Resource:    tx.table.GetName(),
			Permissions: tx.options.Permissions,
			Error:       err.Error(),
		}
		tx.writeAuditLog(auditEntry)

		txErr := NewTransactionError(
			ErrorTypeStorage,
			fmt.Sprintf("failed to search records: %v", err),
			"SEARCH",
			tx.table.GetName(),
			tx.txID,
			true,
			err,
		)
		return nil, txErr
	}

	// 写入审计日志
	auditEntry := AuditLogEntry{
		TxID:        tx.txID,
		Timestamp:   time.Now(),
		Operation:   "SEARCH",
		Table:       tx.table.GetName(),
		Fields:      fields,
		Status:      "SUCCESS",
		ActionType:  "READ",
		Resource:    tx.table.GetName(),
		Permissions: tx.options.Permissions,
	}
	tx.writeAuditLog(auditEntry)

	// 增加操作计数
	tx.operationCount++
	tx.readOperationCount++

	return result, nil
}

// SearchRange 在事务中进行区间搜索（支持读一致性，即使用快照）
func (tx *TableTransaction) SearchRange(funIter storage.FunIter, fieldname string, Start, Limit any) (interface{}, error) {
	// 检查事务状态
	if err := tx.checkTransactionState(); err != nil {
		return nil, err
	}

	// 检查读取权限
	if !tx.CheckPermission("READ") {
		err := NewTransactionError(
			ErrorTypePermission,
			"permission denied",
			"SEARCH_RANGE",
			tx.table.GetName(),
			tx.txID,
			false,
			nil,
		)
		return nil, err
	}

	// 检查字段名是否为空
	if fieldname == "" {
		err := NewTransactionError(
			ErrorTypeValidation,
			"fieldname cannot be empty",
			"SEARCH_RANGE",
			tx.table.GetName(),
			tx.txID,
			false,
			nil,
		)
		return nil, err
	}

	// 创建一个函数，根据隔离级别选择不同的存储获取迭代器
	transactionFunIter := func(start, limit []byte) storage.Iterator {
		switch tx.options.IsolationLevel {
		case ReadUncommitted:
			// 读未提交：直接从原始存储读取
			return tx.originalStore.Iterator(start, limit)
		case ReadCommitted:
			// 读已提交：每次读取都创建新的快照
			if levelDBStore, ok := tx.originalStore.(*storage.LevelDBStore); ok {
				snapshot, err := levelDBStore.Snapshot()
				if err == nil {
					return snapshot.Iterator(start, limit)
				}
			}
			return tx.originalStore.Iterator(start, limit)
		case RepeatableRead, Serializable:
			// 可重复读和可序列化：使用事务开始时创建的快照
			if tx.snapshot != nil {
				return tx.snapshot.Iterator(start, limit)
			}
			return tx.originalStore.Iterator(start, limit)
		default:
			// 默认使用读已提交
			return tx.originalStore.Iterator(start, limit)
		}
	}

	// 调用table.SearchRange方法，传入事务的funIter函数
	result, err := tx.table.SearchRange(transactionFunIter, fieldname, Start, Limit)
	if err != nil {
		// 写入审计日志
		auditEntry := AuditLogEntry{
			TxID:        tx.txID,
			Timestamp:   time.Now(),
			Operation:   "SEARCH_RANGE",
			Table:       tx.table.GetName(),
			Fields:      map[string]any{"fieldname": fieldname, "start": Start, "limit": Limit},
			Status:      "FAILED",
			ActionType:  "READ",
			Resource:    tx.table.GetName(),
			Permissions: tx.options.Permissions,
			Error:       err.Error(),
		}
		tx.writeAuditLog(auditEntry)

		txErr := NewTransactionError(
			ErrorTypeStorage,
			fmt.Sprintf("failed to search range: %v", err),
			"SEARCH_RANGE",
			tx.table.GetName(),
			tx.txID,
			true,
			err,
		)
		return nil, txErr
	}

	// 写入审计日志
	auditEntry := AuditLogEntry{
		TxID:        tx.txID,
		Timestamp:   time.Now(),
		Operation:   "SEARCH_RANGE",
		Table:       tx.table.GetName(),
		Fields:      map[string]any{"fieldname": fieldname, "start": Start, "limit": Limit},
		Status:      "SUCCESS",
		ActionType:  "READ",
		Resource:    tx.table.GetName(),
		Permissions: tx.options.Permissions,
	}
	tx.writeAuditLog(auditEntry)

	// 增加操作计数
	tx.operationCount++
	tx.readOperationCount++

	return result, nil
}

// GetOptions 获取事务选项
func (tx *TableTransaction) GetOptions() *TransactionOptions {
	return tx.options
}

// GetTxID 获取事务ID
func (tx *TableTransaction) GetTxID() uint64 {
	return tx.txID
}

// CheckPermission 检查事务是否有指定权限
func (tx *TableTransaction) CheckPermission(permission string) bool {
	return tx.permissionMap[permission]
}

// EncryptData 加密数据
func (tx *TableTransaction) EncryptData(data []byte) ([]byte, error) {
	if tx.encryptionContext == nil || tx.encryptionContext.cipher == nil {
		return data, nil // 未启用加密，直接返回原始数据
	}

	// 创建一个随机的初始化向量
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("failed to generate IV: %v", err)
	}

	// 创建一个CFB模式的加密器
	stream := cipher.NewCFBEncrypter(tx.encryptionContext.cipher, iv)

	// 加密数据
	ciphertext := make([]byte, len(data))
	stream.XORKeyStream(ciphertext, data)

	// 将IV和密文组合返回
	return append(iv, ciphertext...), nil
}

// DecryptData 解密数据
func (tx *TableTransaction) DecryptData(data []byte) ([]byte, error) {
	if tx.encryptionContext == nil || tx.encryptionContext.cipher == nil {
		return data, nil // 未启用加密，直接返回原始数据
	}

	// 检查数据长度是否足够
	if len(data) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	// 分离IV和密文
	iv := data[:aes.BlockSize]
	ciphertext := data[aes.BlockSize:]

	// 创建一个CFB模式的解密器
	stream := cipher.NewCFBDecrypter(tx.encryptionContext.cipher, iv)

	// 解密数据
	plaintext := make([]byte, len(ciphertext))
	stream.XORKeyStream(plaintext, ciphertext)

	return plaintext, nil
}

// writeTransactionLog 写入事务日志
func (tx *TableTransaction) writeTransactionLog(entry TransactionLogEntry) error {
	if tx.transactionLog == nil {
		return nil
	}

	// 序列化日志条目
	jsonData, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	// 写入日志文件
	_, err = tx.transactionLog.Write(append(jsonData, '\n'))
	if err != nil {
		return err
	}

	// 刷新缓冲区，确保日志持久化
	return tx.transactionLog.Sync()
}

// writeAuditLog 写入审计日志
func (tx *TableTransaction) writeAuditLog(entry AuditLogEntry) error {
	if tx.auditLogPath == "" {
		return nil
	}

	// 创建审计日志文件
	auditFile := filepath.Join(tx.auditLogPath, fmt.Sprintf("audit_%s.log", time.Now().Format("20060102")))
	file, err := os.OpenFile(auditFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// 序列化审计日志条目
	jsonData, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	// 写入审计日志文件
	_, err = file.Write(append(jsonData, '\n'))
	if err != nil {
		return err
	}

	// 刷新缓冲区，确保日志持久化
	err = file.Sync()
	if err != nil {
		return err
	}

	// 更新审计事件计数和时间
	tx.auditEventCount++
	tx.lastAuditTime = time.Now()

	return nil
}

// BackupTransactionLog 备份事务日志
func (tx *TableTransaction) BackupTransactionLog() error {
	if !tx.options.EnableTransactionLog || tx.transactionLogPath == "" {
		return nil
	}

	// 创建备份目录
	backupDir := filepath.Join(tx.transactionLogPath, "backups")
	err := os.MkdirAll(backupDir, 0755)
	if err != nil {
		return err
	}

	// 创建备份文件
	backupFile := filepath.Join(backupDir, fmt.Sprintf("backup_%s.log", time.Now().Format("20060102150405")))

	// 复制所有日志文件到备份
	logFiles, err := filepath.Glob(filepath.Join(tx.transactionLogPath, "tx_*.log"))
	if err != nil {
		return err
	}

	// 打开备份文件
	backup, err := os.Create(backupFile)
	if err != nil {
		return err
	}
	defer backup.Close()

	// 复制每个日志文件的内容
	for _, logFile := range logFiles {
		content, err := os.ReadFile(logFile)
		if err != nil {
			continue
		}
		_, err = backup.Write(content)
		if err != nil {
			continue
		}
	}

	// 刷新并关闭备份文件
	err = backup.Sync()
	if err != nil {
		return err
	}

	// 更新最后备份时间
	tx.lastBackupTime = time.Now()

	return nil
}

// RecoverFromTransactionLog 从事务日志恢复
func (tx *TableTransaction) RecoverFromTransactionLog(logPath string) error {
	// 检查日志路径是否存在
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		return fmt.Errorf("transaction log path does not exist: %s", logPath)
	}

	// 读取日志文件
	logFiles, err := filepath.Glob(filepath.Join(logPath, "tx_*.log"))
	if err != nil {
		return err
	}

	// 处理每个日志文件
	for _, logFile := range logFiles {
		content, err := os.ReadFile(logFile)
		if err != nil {
			continue
		}

		// 解析日志条目
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}

			var entry TransactionLogEntry
			err := json.Unmarshal([]byte(line), &entry)
			if err != nil {
				continue
			}

			// 根据日志条目执行恢复操作
			// 这里简化处理，实际应用中需要根据具体操作类型执行相应的恢复逻辑
			// 例如：重新执行未提交的事务，回滚已提交但未持久化的事务等
		}
	}

	return nil
}

// GetTransactionLogPath 获取事务日志路径
func (tx *TableTransaction) GetTransactionLogPath() string {
	return tx.transactionLogPath
}

// Commit 提交事务
func (tx *TableTransaction) Commit() error {
	if err := tx.checkCommitted(); err != nil {
		return err
	}

	endTime := time.Now()                 // 记录事务结束时间
	duration := endTime.Sub(tx.startTime) // 计算事务用时
	monitor.GTransactionStatsMap.SetTimeAsync(tx.txID, duration, tx.table.GetName(), tx.options.IsolationLevel, true)

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
			// 写入提交失败日志
			entry := TransactionLogEntry{
				TxID:         tx.txID,
				Timestamp:    time.Now(),
				Operation:    "COMMIT",
				Table:        tx.table.GetName(),
				Status:       "FAILED",
				ErrorMessage: err.Error(),
			}
			tx.writeTransactionLog(entry)

			// 提交失败，释放快照资源
			if tx.snapshot != nil {
				tx.snapshot.Release()
			}
			// 释放锁
			tx.releaseLocks()
			// 关闭事务日志
			if tx.transactionLog != nil {
				tx.transactionLog.Close()
				tx.transactionLog = nil
			}
			return err
		}

		// 写入提交成功日志
		entry := TransactionLogEntry{
			TxID:      tx.txID,
			Timestamp: time.Now(),
			Operation: "COMMIT",
			Table:     tx.table.GetName(),
			Status:    "SUCCESS",
		}
		tx.writeTransactionLog(entry)

		// 2. 释放快照资源
		if tx.snapshot != nil {
			tx.snapshot.Release()
		}

		// 3. 检查是否需要备份事务日志
		if tx.options.EnableTransactionLog && time.Since(tx.lastBackupTime) >= tx.options.TransactionLogBackupInterval {
			tx.BackupTransactionLog()
		}
	}

	// 4. 标记事务已结束，清空缓存
	tx.committed = true
	tx.cache = nil
	tx.lockKeyCache = nil
	tx.versionCache = nil
	tx.children = nil

	// 5. 停止超时定时器
	if tx.timeoutTimer != nil {
		tx.timeoutTimer.Stop()
		tx.timeoutTimer = nil
	}

	// 6. 释放锁
	tx.releaseLocks()

	// 6. 记录事务完成监控信息
	monitor.GTransactionStatsMap.SetCountAsync(tx.txID, tx.operationCount, tx.conflictCount, tx.table.GetName(), tx.options.IsolationLevel, true)

	// 6. 关闭事务日志
	if tx.transactionLog != nil {
		tx.transactionLog.Close()
		tx.transactionLog = nil
	}

	// 7. 归还事务对象到池中
	if tx.parent == nil {
		// 只有根事务才归还到池，子事务由父事务管理
		GlobalTableTransactionPool.Put(tx)
	}

	return nil
}

// releaseLocks 释放事务获取的所有锁
func (tx *TableTransaction) releaseLocks() {
	// 对于Serializable隔离级别，释放所有获取的锁
	if tx.options.IsolationLevel == Serializable {
		// 释放读锁
		for key := range tx.readLocks {
			if tx.readLocks[key] {
				globalLockManager.ReleaseReadLock(key)
			}
		}

		// 释放写锁
		for key := range tx.writeLocks {
			if tx.writeLocks[key] {
				globalLockManager.ReleaseWriteLock(key)
			}
		}

		// 清空锁映射
		tx.readLocks = nil
		tx.writeLocks = nil
	}
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

	// 检查是否是TransactionError类型的错误
	if txErr, ok := err.(*TransactionError); ok {
		// 根据IsRetryable字段判断
		return txErr.IsRetryable
	}

	// 检查是否是包装了TransactionError的错误
	var wrappedTxErr *TransactionError
	if As(err, &wrappedTxErr) {
		return wrappedTxErr.IsRetryable
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

// As 实现errors.As接口
func As(err error, target interface{}) bool {
	if target == nil {
		return false
	}

	// 检查直接类型
	if e, ok := err.(*TransactionError); ok {
		if t, ok := target.(**TransactionError); ok {
			*t = e
			return true
		}
	}

	// 检查包装的错误
	if wrapper, ok := err.(interface{ Unwrap() error }); ok {
		return As(wrapper.Unwrap(), target)
	}

	return false
}

// adjustBatchSize 根据操作频率和冲突率调整批量大小
func (tx *TableTransaction) adjustBatchSize() {
	// 计算操作频率（操作数/事务时间）
	elapsed := time.Since(tx.startTime)
	if elapsed < time.Millisecond {
		elapsed = time.Millisecond
	}
	operationRate := float64(tx.operationCount) / elapsed.Seconds()

	// 计算冲突率
	var conflictRate float64
	if tx.operationCount > 0 {
		conflictRate = float64(tx.conflictCount) / float64(tx.operationCount)
	}

	// 根据操作频率和冲突率调整批量大小
	if conflictRate > 0.1 { // 冲突率高于10%
		// 减少批量大小以降低冲突
		newBatchSize := tx.currentBatchSize / 2
		if newBatchSize < tx.minBatchSize {
			newBatchSize = tx.minBatchSize
		}
		tx.currentBatchSize = newBatchSize
	} else if operationRate > 100 { // 操作频率高于100次/秒
		// 增加批量大小以提高吞吐量
		newBatchSize := tx.currentBatchSize * 2
		if newBatchSize > tx.maxBatchSize {
			newBatchSize = tx.maxBatchSize
		}
		tx.currentBatchSize = newBatchSize
	}

	// 检查是否需要提交部分批量
	if tx.writeOperationCount > 0 && tx.writeOperationCount%tx.currentBatchSize == 0 {
		// 这里可以添加部分批量提交逻辑
		// 例如：创建一个新的batch，提交当前batch，然后继续使用新batch
	}
}

// shouldCommitBatch 检查是否应该提交批量操作
func (tx *TableTransaction) shouldCommitBatch() bool {
	// 如果批量大小达到阈值，应该提交
	if tx.writeOperationCount > 0 && tx.writeOperationCount%tx.currentBatchSize == 0 {
		return true
	}
	// 如果事务时间过长，应该提交
	if time.Since(tx.startTime) > 30*time.Second {
		return true
	}
	return false
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
	monitor.GTransactionStatsMap.SetTimeAsync(tx.txID, duration, tx.table.GetName(), tx.options.IsolationLevel, false)

	// 写入回滚日志
	entry := TransactionLogEntry{
		TxID:      tx.txID,
		Timestamp: time.Now(),
		Operation: "ROLLBACK",
		Table:     tx.table.GetName(),
		Status:    "SUCCESS",
	}
	tx.writeTransactionLog(entry)

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
	tx.versionCache = nil
	tx.children = nil

	// 3. 停止超时定时器
	if tx.timeoutTimer != nil {
		tx.timeoutTimer.Stop()
		tx.timeoutTimer = nil
	}

	// 4. 释放锁
	tx.releaseLocks()

	// 4. 记录事务回滚监控信息
	monitor.GTransactionStatsMap.SetCountAsync(tx.txID, tx.operationCount, tx.conflictCount, tx.table.GetName(), tx.options.IsolationLevel, false)

	// 4. 关闭事务日志
	if tx.transactionLog != nil {
		tx.transactionLog.Close()
		tx.transactionLog = nil
	}

	// 5. 归还事务对象到池中
	if tx.parent == nil {
		// 只有根事务才归还到池，子事务由父事务管理
		GlobalTableTransactionPool.Put(tx)
	}

	return nil
}
