package engine

import (
	"fmt"

	"github.com/liaoran123/sfsDb/storage"
	"github.com/liaoran123/sfsDb/util"
)

// Transaction 定义事务接口
type Transaction interface {
	// Insert 在事务中插入记录
	Insert(fields *map[string]any) (int, error)
	// Update 在事务中更新记录
	Update(fields *map[string]any) error
	// Delete 在事务中删除记录
	Delete(fields *map[string]any) error
	// Search 在事务中搜索记录（支持读一致性）
	Search(fields *map[string]any, ops ...util.ComparisonOperator) *TableIter
	// Read 在事务中读取单条记录（支持读一致性）
	Read(fields *map[string]any) ([]byte, error)
	// Commit 提交事务
	Commit() error
	// Rollback 回滚事务
	Rollback() error
}

// TableTransaction 实现Transaction接口的具体结构体
type TableTransaction struct {
	table          *Table        // 关联的表
	batch          storage.Batch // 事务使用的batch
	committed      bool          // 是否已提交
	useSnapshot    bool          // 是否使用快照模式
	isLevelDBStore bool          // 当前KVStore是否是LevelDBStore
	// 事务内修改缓存，用于读取自己的写操作
	// key: 主键值的字符串表示，value: 记录的字节数组
	cache map[string][]byte // 事务内修改缓存
}

// Begin 创建一个新的事务
func (t *Table) Begin() (Transaction, error) {
	batch := t.kvStore.GetBatch()
	if batch == nil {
		return nil, fmt.Errorf("failed to create batch for transaction")
	}

	useSnapshot := false
	isLevelDBStore := false

	// 检查是否是LevelDBStore，如果是则切换到快照模式
	if levelDBStore, ok := t.kvStore.(*storage.LevelDBStore); ok {
		isLevelDBStore = true
		if err := levelDBStore.SwitchToSnapshot(); err != nil {
			return nil, fmt.Errorf("failed to switch to snapshot: %v", err)
		}
		useSnapshot = true
	}

	return &TableTransaction{
		table:          t,
		batch:          batch,
		committed:      false,
		useSnapshot:    useSnapshot,
		isLevelDBStore: isLevelDBStore,
		cache:          make(map[string][]byte), // 初始化事务内缓存
	}, nil
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

	return id, nil
}

// Update 在事务中更新记录
func (tx *TableTransaction) Update(fields *map[string]any) error {
	if err := tx.checkCommitted(); err != nil {
		return err
	}

	// 执行更新操作
	err := tx.table.Update(fields, tx.batch)
	if err != nil {
		return err
	}

	// 生成缓存键
	cacheKey := tx.getCacheKey(fields)

	// 从缓存中删除旧记录，强制后续读取从数据库获取最新值
	delete(tx.cache, cacheKey)

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

	return nil
}

// getCacheKey 生成缓存键
func (tx *TableTransaction) getCacheKey(fields *map[string]any) string {
	// 将fields转换为fieldsBytes，用于生成主键
	fieldsBytes := tx.table.FieldsToBytes(fields)
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

	// 2. 缓存中没有，直接从表中读取（此时表处于快照模式）
	return tx.table.Read(fields)
}

// Search 在事务中搜索记录（支持读一致性）
func (tx *TableTransaction) Search(fields *map[string]any, ops ...util.ComparisonOperator) *TableIter {
	if tx.committed {
		return nil
	}

	// 直接调用表的Search方法，此时表处于快照模式，会从快照中读取数据
	// 这样返回的迭代器会使用快照来迭代，保证读一致性
	return tx.table.Search(fields, ops...)
}

// Commit 提交事务
func (tx *TableTransaction) Commit() error {
	if err := tx.checkCommitted(); err != nil {
		return err
	}

	// 1. 提交批量操作
	err := tx.table.kvStore.WriteBatch(tx.batch)
	if err != nil {
		return err
	}

	// 2. 如果使用了快照模式，切换回DB模式
	if tx.useSnapshot && tx.isLevelDBStore {
		if levelDBStore, ok := tx.table.kvStore.(*storage.LevelDBStore); ok {
			if err := levelDBStore.SwitchToDB(); err != nil {
				return fmt.Errorf("failed to switch back to DB mode: %v", err)
			}
		}
	}

	// 3. 标记事务已提交，清空缓存
	tx.committed = true
	tx.cache = nil

	return nil
}

// Rollback 回滚事务
// 注意：LevelDB的WriteBatch不支持真正的回滚
// 此方法仅标记事务已结束，防止重复提交，并切换回DB模式
func (tx *TableTransaction) Rollback() error {
	if err := tx.checkCommitted(); err != nil {
		return err
	}

	// 1. 如果使用了快照模式，切换回DB模式
	if tx.useSnapshot && tx.isLevelDBStore {
		if levelDBStore, ok := tx.table.kvStore.(*storage.LevelDBStore); ok {
			if err := levelDBStore.SwitchToDB(); err != nil {
				return fmt.Errorf("failed to switch back to DB mode: %v", err)
			}
		}
	}
	// 2. 标记事务已提交，清空缓存
	tx.committed = true
	tx.cache = nil

	return nil
}
