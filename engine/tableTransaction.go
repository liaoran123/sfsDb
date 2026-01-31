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
	table         *Table           // 关联的表
	batch         storage.Batch    // 事务使用的batch，原子性
	committed     bool             // 是否已提交，提交成功后则是持久性。
	snapshot      storage.Snapshot // 事务使用的快照，一致性。
	originalStore storage.Store    // 原始存储，用于写操作
	// 事务内修改缓存，用于读取自己的写操作
	// key: 主键值的字符串表示，value: 记录的字节数组
	cache map[string][]byte // 事务内修改缓存 //隔离性
}

// Begin 创建一个新的事务
func (t *Table) Begin() (Transaction, error) {
	batch := t.kvStore.GetBatch()
	if batch == nil {
		return nil, fmt.Errorf("failed to create batch for transaction")
	}

	// 为每个事务创建自己的快照实例，而不是共享表级别的快照
	var snapshot storage.Snapshot
	var err error

	// 检查是否是LevelDBStore，如果是则创建快照
	if levelDBStore, ok := t.kvStore.(*storage.LevelDBStore); ok {
		// 创建一个新的快照实例
		snapshot, err = levelDBStore.Snapshot()
		if err != nil {
			return nil, fmt.Errorf("failed to create snapshot: %v", err)
		}
	}

	return &TableTransaction{
		table:         t,
		batch:         batch,
		committed:     false,
		snapshot:      snapshot,
		originalStore: t.kvStore,
		cache:         make(map[string][]byte), // 初始化事务内缓存
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

	// 生成缓存键
	cacheKey := tx.getCacheKey(fields)

	// 检查缓存中是否有要更新的记录
	if record, exists := tx.cache[cacheKey]; exists {
		// 缓存中有记录，模拟 Table.Update 的行为
		// 1. 解析记录
		pk := tx.table.GetPrimaryKey()
		fieldsBytes, err := pk.Parse(tx.table.fieldsid, record)
		if err != nil {
			return err
		}

		// 2. 准备更新字段列表
		updateFields := GetStringSliceWithStrategy()
		defer PutStringSliceWithStrategy(updateFields)
		for field := range *fields {
			// 排除主键字段
			if pk.MatchFields(field) {
				continue
			}
			updateFields = append(updateFields, field)
		}

		// 3. 检查更新字段个数是否为0
		if len(updateFields) == 0 {
			return nil
		}

		// 4. 执行更新操作（删除旧记录）
		batchContainer := NewBatchContainer(tx.batch, tx.table.indexs, tx.table.id, tx.table.kvStore)
		batchContainer.Operation(fieldsBytes, updateFields...)

		// 5. 更新字段值
		for field, val := range *fields {
			// 排除主键字段
			if pk.MatchFields(field) {
				continue
			}
			if _, ok := tx.table.fields[field]; ok {
				(*fieldsBytes)[field] = util.AnyToBytes(val)
			}
		}

		// 6. 更新版本号
		currentVersionbyte, exists := (*fieldsBytes)["v"]
		if !exists {
			currentVersionbyte = []byte{1} // 默认版本号
		}
		currentVersion := int(util.Bytes(currentVersionbyte).Uint64())
		(*fields)["v"] = currentVersion + 1
		(*fieldsBytes)["v"] = util.AnyToBytes(currentVersion + 1)

		// 7. 格式化记录
		updatedRecord := tx.table.FormatRecord(fieldsBytes)

		// 8. 执行更新操作（添加新记录）
		batchContainer.SetValue(0, updatedRecord)                               // 添加主键value=record
		batchContainer.SetValue(1, tx.table.GetPrimaryKey().GetID(fieldsBytes)) // 添加普通索引value=GetPrimaryKey().GetID()
		batchContainer.Operation(fieldsBytes, updateFields...)

		// 9. 更新缓存
		tx.cache[cacheKey] = updatedRecord
	} else {
		// 缓存中没有记录，直接执行更新操作
		err := tx.table.Update(fields, tx.batch)
		if err != nil {
			return err
		}

		// 从缓存中删除旧记录，强制后续读取从数据库获取最新值
		delete(tx.cache, cacheKey)
	}

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

	// 2. 缓存中没有，使用事务自己的快照或原始存储读取
	// 生成主键键值
	fieldsBytes := tx.table.FieldsToBytes(fields)
	pkKey := tx.table.GetPrimaryKey().JoinValue(fieldsBytes, tx.table.id)

	// 如果有快照，使用快照读取；否则使用原始存储
	if tx.snapshot != nil {
		return tx.snapshot.Get(pkKey)
	}
	return tx.originalStore.Get(pkKey)
}

// Search 在事务中搜索记录（支持读一致性，即使用快照）
func (tx *TableTransaction) Search(fields *map[string]any, ops ...util.ComparisonOperator) *TableIter {
	if tx.committed {
		return nil
	}

	// 复制table.Search方法的核心逻辑，但使用事务的快照或原始存储
	var field []string
	for k := range *fields {
		//判断字段是否在表中
		if _, ok := tx.table.fields[k]; !ok {
			return nil
		}
		field = append(field, k)
	}

	//匹配索引
	idx := tx.table.MatchIndex(field...)
	if idx == nil {
		return nil
	}

	//生成搜索键
	fieldsBytes := tx.table.FieldsToBytesNil(fields)
	key := idx.JoinValue(fieldsBytes, tx.table.id)

	//确定比较操作符
	var op util.ComparisonOperator
	if len(ops) == 0 {
		op = util.Like
	} else {
		op = ops[0]
	}

	//生成迭代器的前缀范围
	pfx := idx.Prefix(tx.table.id)
	pfx = append(pfx, SPLIT[0])
	rangeHelper := util.NewRangeHelper(pfx)

	//根据操作符生成范围并获取迭代器
	var iter storage.Iterator
	var slice *util.Range
	var tbiter *TableIter

	if op != util.NotEqual {
		slice = rangeHelper.FromComparison(op, key)
	} else {
		slice = rangeHelper.FromComparison(util.Like, pfx)
	}

	//获取迭代器
	if tx.snapshot != nil {
		iter = tx.snapshot.Iterator(slice.Start, slice.Limit)
	} else {
		iter = tx.originalStore.Iterator(slice.Start, slice.Limit)
	}

	//创建TableIter
	tbiter = TableIterNew(tx.table, iter, idx)

	//如果是NotEqual操作，设置跳跃区间
	if op == util.NotEqual {
		neslice := rangeHelper.FromComparison(util.Like, key)
		var jumpIter storage.Iterator
		if tx.snapshot != nil {
			jumpIter = tx.snapshot.Iterator(neslice.Start, neslice.Limit)
		} else {
			jumpIter = tx.originalStore.Iterator(neslice.Start, neslice.Limit)
		}
		tbiter.SetJumpRanges(jumpIter)
	}

	return tbiter
}

// Commit 提交事务
func (tx *TableTransaction) Commit() error {
	if err := tx.checkCommitted(); err != nil {
		return err
	}

	// 1. 提交批量操作
	// 使用原始存储执行写操作
	err := tx.originalStore.WriteBatch(tx.batch)
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

	// 3. 标记事务已结束，清空缓存
	tx.committed = true
	tx.cache = nil

	return nil
}

// Rollback 回滚事务
// 注意：LevelDB的WriteBatch不支持真正的回滚
// 此方法仅标记事务已结束，防止重复提交，并释放快照资源
func (tx *TableTransaction) Rollback() error {
	if err := tx.checkCommitted(); err != nil {
		return err
	}

	// 1. 释放快照资源
	if tx.snapshot != nil {
		tx.snapshot.Release()
	}

	// 2. 标记事务已结束，清空缓存
	tx.committed = true
	tx.cache = nil

	return nil
}
