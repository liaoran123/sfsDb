package engine

import (
	"fmt"

	"github.com/liaoran123/sfsDb/storage"
	"github.com/liaoran123/sfsDb/util"
)

// 插入记录
func (t *Table) Insert(fields *map[string]any, batchs ...storage.Batch) (currentID int, err error) {
	// 检查参数
	if t.fields == nil {
		return 0, fmt.Errorf("表 '%s' 未设置字段和类型", t.name)
	}
	if fields == nil {
		return 0, fmt.Errorf("fields cannot be nil")
	}

	//当前自动增值的值
	currentID = -1

	//获取主键字段
	primaryFields := t.GetPrimaryFields() //t.GetPrimaryKey().GetFields()
	if len(primaryFields) == 0 {
		return 0, fmt.Errorf("表 '%s' 没有设置主键", t.name)
	}

	//是否支持默认自动增值主键，单主键并且主键字段名为"id"
	pklen := len(primaryFields)
	pkfield := primaryFields[0]
	supportDefault := pklen == 1 && pkfield == "id"
	if supportDefault {
		// 检查是否提供了主键字段
		//使用默认自动增值主键时，不需要提供主键字段，系统自动生成，强制使用"id"字段和自动增值主键
		_, ok := (*fields)[pkfield]
		if !ok { //未提供主键字段，自动生成主键值
			currentID = t.GetAutoInc()
			(*fields)[pkfield] = currentID
		} else { //提供了主键字段id，但是值为nil，自动生成主键值
			if (*fields)[pkfield] == nil {
				currentID = t.GetAutoInc()
				(*fields)[pkfield] = currentID
			}
		}
	}

	// 检查字段类型是否匹配
	if err = t.CheckType(fields); err != nil {
		return -1, err
	}
	currentID = (*fields)[pkfield].(int)
	//currentID = util.AnyToInt((*fields)[pkfield])
	// 添加初始版本号
	if _, hasVersion := (*fields)["v"]; !hasVersion {
		(*fields)["v"] = 1 // 初始版本号为1
	}

	// 转换字段为字节数组
	fieldsBytes := t.FieldsToBytes(fields)
	var batch storage.Batch
	//是否用户手动控制事务
	if len(batchs) > 0 { //用户手动控制事务
		batch = batchs[0]
		if batch == nil {
			return -1, fmt.Errorf("batch cannot be nil")
		}
	} else {
		batch = t.kvStore.GetBatch()
		if batch == nil {
			return -1, fmt.Errorf("failed to get batch")
		}
	}
	//格式化记录
	record := t.FormatRecord(fieldsBytes)
	BatchContainer := NewBatchContainer(batch, t.indexs, t.id, t.kvStore)
	BatchContainer.SetValue(0, record)                               //添加主键value=record
	BatchContainer.SetValue(1, t.GetPrimaryKey().GetID(fieldsBytes)) //添加普通索引value=GetPrimaryKey().GetID()
	//添加全文索引key=joinValue,value=nil
	BatchContainer.Operation(fieldsBytes)
	//t.Operation(fieldsBytes, batch, BatchContainer)

	if len(batchs) == 0 {
		// 提交批量操作
		if err = t.kvStore.WriteBatch(batch); err != nil {
			return -1, err
		}
	}

	//fmt.Printf("Insert BatchContainer.Len(): %v\n", BatchContainer.Len())
	return currentID, nil
	/*
		数据流动流程
		1,外部传入 Insert(fields *map[string]any
		2,// 转换字段为字节数组
		fieldsBytes := t.FieldsToBytes(fields)
		3,//格式化记录
		record := t.FormatRecord(fieldsBytes)
		4,添加更新记录
		tableiter 查询功能则与上面添加的流程相反。一正一逆。
	*/
}

// BatchInsert 批量插入多条记录
// records []*map[string]any 要插入的记录列表
// batchs ...storage.Batch 可选的批量操作容器
// 返回值：插入记录的ID列表和错误信息
func (t *Table) BatchInsert(records []*map[string]any, batchs ...storage.Batch) ([]int, error) {
	// 检查参数
	if t.fields == nil {
		return nil, fmt.Errorf("表 '%s' 未设置字段和类型", t.name)
	}
	if len(records) == 0 {
		return []int{}, nil
	}
	if records == nil {
		return nil, fmt.Errorf("records cannot be nil")
	}

	// 获取主键字段
	primaryFields := t.GetPrimaryFields() //t.GetPrimaryKey().GetFields()
	if len(primaryFields) == 0 {
		return nil, fmt.Errorf("表 '%s' 没有设置主键", t.name)
	}

	// 是否支持默认自动增值主键，单主键并且主键字段名为"id"
	pklen := len(primaryFields)
	pkfield := primaryFields[0]
	supportDefault := pklen == 1 && pkfield == "id"

	// 处理批量操作
	var batch storage.Batch
	if len(batchs) > 0 { // 用户手动控制事务
		batch = batchs[0]
		if batch == nil {
			return nil, fmt.Errorf("batch cannot be nil")
		}
	} else {
		batch = t.kvStore.GetBatch()
		if batch == nil {
			return nil, fmt.Errorf("failed to get batch")
		}
	}

	// 预分配ID列表容量
	ids := make([]int, len(records))

	// 计算需要自动生成的ID数量
	autoIncCount := 0
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

	// 批量获取自动增值ID，确保并发安全
	var autoIncStart int
	if supportDefault && autoIncCount > 0 {
		autoIncStart = t.GetAutoIncBatch(autoIncCount)
		// 后续ID可以直接计算，不需要重复调用GetAutoInc()
	}

	// 处理记录并批量插入
	autoIncIdx := 0
	BatchContainer := NewBatchContainer(batch, t.indexs, t.id, t.kvStore) // 重用BatchContainer

	for i, fields := range records {
		// 检查字段类型
		if err := t.CheckType(fields); err != nil {
			return nil, err
		}

		// 处理自动增值主键
		if supportDefault {
			if _, ok := (*fields)[pkfield]; !ok || (*fields)[pkfield] == nil {
				// 使用预分配的自动增值ID
				ids[i] = autoIncStart + autoIncIdx
				(*fields)[pkfield] = ids[i]
				autoIncIdx++
			} else {
				// 使用提供的主键值
				ids[i] = util.AnyToInt((*fields)[pkfield])
			}
		} else {
			// 非默认自动增值主键，使用提供的主键值
			ids[i] = util.AnyToInt((*fields)[pkfield])
		}

		// 添加初始版本号
		if _, hasVersion := (*fields)["v"]; !hasVersion {
			(*fields)["v"] = 1 // 初始版本号为1
		}

		// 转换字段为字节数组
		fieldsBytes := t.FieldsToBytes(fields)

		// 格式化记录
		record := t.FormatRecord(fieldsBytes)

		// 批量添加记录
		BatchContainer.SetValue(0, record)                               // 添加主键value=record
		BatchContainer.SetValue(1, t.GetPrimaryKey().GetID(fieldsBytes)) // 添加普通索引value=GetPrimaryKey().GetID()
		// 添加全文索引key=joinValue,value=nil
		BatchContainer.Operation(fieldsBytes)
	}

	// 提交批量操作
	if len(batchs) == 0 {
		if err := t.kvStore.WriteBatch(batch); err != nil {
			return nil, err
		}
	}

	return ids, nil
}

// BatchInsertWithSize 带批量大小控制的批量插入
// records []*map[string]any 要插入的记录列表
// batchSize int 每批处理的记录数量
// batchs ...storage.Batch 可选的批量操作容器
// 返回值：插入记录的ID列表和错误信息
func (t *Table) BatchInsertWithSize(records []*map[string]any, batchSize int, batchs ...storage.Batch) ([]int, error) {
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
		batchIds, err := t.BatchInsert(batchRecords, batchs...)
		if err != nil {
			return nil, err
		}

		// 复制ID到结果列表
		copy(allIds[start:end], batchIds)
	}

	return allIds, nil
}

// 删除记录
// fields *map[string]any 主键值，可能是组合主键
func (t *Table) Delete(fields *map[string]any, batchs ...storage.Batch) error {
	//检查是否提供了所有主键字段
	for _, field := range t.GetPrimaryFields() {
		if _, ok := (*fields)[field]; !ok {
			return fmt.Errorf("必须提供主键字段 '%s'", field)
		}
	}
	//读取记录
	record, err := t.Read(fields)
	if err != nil {
		return err
	}
	if record == nil {
		return fmt.Errorf("主键值 '%v' 的记录不存在", fields)
	}

	var batch storage.Batch
	//是否用户手动控制事务
	useBatch := len(batchs) > 0
	if useBatch { //用户手动控制事务
		batch = batchs[0]
	} else {
		batch = t.kvStore.GetBatch()
	}

	//反序列化记录，并且将字段值转换为对应的类型
	//fieldsBytes := t.ParseRecord(record)
	pk := t.GetPrimaryKey()
	fieldsBytes, err := pk.Parse(t.fieldsid, record)
	if err != nil {
		//释放batch资源
		return err
	}
	BatchContainer := NewBatchContainer(batch, t.indexs, t.id, t.kvStore)
	BatchContainer.Operation(fieldsBytes)
	if len(batchs) == 0 { //用户未手动控制事务，自动提交
		t.kvStore.WriteBatch(batch)
	}
	//fmt.Printf("Delete BatchContainer.Len(): %v\n", BatchContainer.Len())
	return nil
}

// 从按主键数据库读取记录
func (t *Table) Read(fields *map[string]any) ([]byte, error) {
	fieldsBytes := t.FieldsToBytes(fields)
	key := t.GetPrimaryKey().JoinValue(fieldsBytes, t.id)
	return t.ReadByBytes(key), nil
}

// 更新记录，不支持修改主键字段
// fields *map[string]any 主键值，可能是组合主键
// 乐观锁并发控制，允许多个事务同时读取记录，但只有一个事务能成功更新记录，避免了并发更新冲突。
func (t *Table) Update(fields *map[string]any, batchs ...storage.Batch) error {
	//检查是否提供了所有主键字段
	for _, field := range t.GetPrimaryFields() {
		if _, ok := (*fields)[field]; !ok {
			return fmt.Errorf("必须提供主键字段 '%s'", field)
		}
	}
	// 检查字段类型是否匹配
	if err := t.CheckType(fields); err != nil {
		return err
	}
	var batch storage.Batch
	//是否用户手动控制事务
	useBatch := len(batchs) > 0
	if useBatch { //用户手动控制事务
		batch = batchs[0]
	} else {
		batch = t.kvStore.GetBatch()
	}
	//读取记录
	record, err := t.Read(fields)
	if err != nil {
		return err
	}
	if record == nil {
		return fmt.Errorf("主键值 '%v' 的记录不存在", fields)
	}
	updateFields := GetStringSlice()
	defer PutStringSlice(updateFields)
	for field := range *fields {
		//排除主键字段
		if t.GetPrimaryKey().MatchFields(field) {
			continue
		}
		updateFields = append(updateFields, field)
	}
	//检查更新字段个数是否为0
	if len(updateFields) == 0 {
		return nil
	}
	//反序列化记录，并且将字段值转换为对应的类型
	//fieldsBytes := t.ParseRecord(record)
	pk := t.GetPrimaryKey()
	fieldsBytes, err := pk.Parse(t.fieldsid, record)
	if err != nil {
		return err
	}
	//---------删除-----------------------
	BatchContainer := NewBatchContainer(batch, t.indexs, t.id, t.kvStore)
	BatchContainer.Operation(fieldsBytes, updateFields...)

	//---------更新-----------------------
	// 获取当前版本号
	currentVersionbyte, exists := (*fieldsBytes)["v"]
	if !exists {
		currentVersionbyte = []byte{1} // 默认版本号
	}
	currentVersion := int(util.Bytes(currentVersionbyte).Uint64())

	// 检查版本号是否匹配
	if updateVersion, hasVersion := (*fields)["v"].(int); hasVersion {
		if updateVersion != currentVersion {
			return fmt.Errorf("optimistic lock conflict: version mismatch, expected %d, got %d", currentVersion, updateVersion)
		}
	}
	(*fields)["v"] = currentVersion + 1 // 更新版本号

	//更新字段值
	for field, val := range *fields {
		//排除主键字段，主键字段不能更新
		if t.GetPrimaryKey().MatchFields(field) {
			continue
		}
		if _, ok := t.fields[field]; ok {
			(*fieldsBytes)[field] = util.AnyToBytes(val)
		}
	}
	//设置新值添加
	record = t.FormatRecord(fieldsBytes)
	BatchContainer.SetValue(0, record)                               //添加主键value=record
	BatchContainer.SetValue(1, t.GetPrimaryKey().GetID(fieldsBytes)) //添加普通索引value=GetPrimaryKey().GetID()
	//添加全文索引key=joinValue,value=nil
	BatchContainer.Operation(fieldsBytes, updateFields...)
	//提交事务
	if len(batchs) == 0 { //用户未手动控制事务，自动提交
		if err := t.kvStore.WriteBatch(batch); err != nil {
			return err
		}
	}
	//fmt.Printf("Update BatchContainer.Len(): %v\n", BatchContainer.Len())
	return nil

	/*
		数据流动流程
		1,外部传入 Update(fields *map[string]any
		2,//读取旧记录
		record, err := t.Read(fields)
		3,反格式化旧记录：
		fieldsBytes, err := pk.Parse(t.fieldsid, record)
		4，更新字段值fieldsBytes
		5,//格式化新记录
		record = t.FormatRecord(fieldsBytes) 回转到2.
		6，添加更新记录
	*/
}

// 从按主键数据库读取记录
func (t *Table) ReadByBytes(key []byte) []byte {
	v, err := t.kvStore.Get(key)
	if err != nil {
		// ErrNotFound 是正常的未找到错误，不需要打印
		if err != storage.ErrNotFound {
			fmt.Printf("读取记录失败: %v\n", err)
		}
		return nil
	}
	return v
}

// 遍历表所有kv键值对，用于快速复制表用或删除表数据
func (t *Table) For() storage.Iterator {
	pfx := []byte{byte(t.id), SPLIT[0]}
	rangeHelper := util.NewRangeHelper(pfx)
	slice := rangeHelper.FromComparison(util.Like, pfx)
	return t.kvStore.Iterator(slice.Start, slice.Limit)
}

// 删除表的所有数据
func (t *Table) DeleteAll() error {
	// 获取表的所有kv键值对迭代器
	iter := t.For()
	defer iter.Release()

	// 创建批量操作
	batch := t.kvStore.GetBatch()
	if batch == nil {
		return fmt.Errorf("failed to get batch")
	}

	// 定义批量操作的大小限制
	const batchSizeLimit = 1000

	// 遍历并删除所有键值对
	count := 0
	for iter.Next() {
		key := iter.Key()
		batch.Delete(key)
		count++

		// 当批量操作的大小达到限制时，执行批量操作并重置批量操作对象
		if count >= batchSizeLimit {
			// 提交批量操作
			if err := t.kvStore.WriteBatch(batch); err != nil {
				return err
			}

			// 重置计数器和批量操作对象
			count = 0
			batch = t.kvStore.GetBatch()
			if batch == nil {
				return fmt.Errorf("failed to get batch")
			}
		}
	}

	// 执行剩余的批量操作
	if count > 0 {
		if err := t.kvStore.WriteBatch(batch); err != nil {
			return err
		}
	}

	return nil
}

// 遍历表所有数据
func (t *Table) ForData() *TableIter {
	pfx := t.GetPrimaryKey().Prefix(t.id)
	pfx = append(pfx, SPLIT[0])
	rangeHelper := util.NewRangeHelper(pfx)
	slice := rangeHelper.FromComparison(util.Like, []byte(pfx))
	return TableIterNew(t, t.kvStore.Iterator(slice.Start, slice.Limit), t.GetPrimaryKey())
}

// 将数据转换为字节数组，该合适搜索时用。搜索时nil值不能更改,否则导致结果错误
func (t *Table) FieldsToBytesNil(fields *map[string]any) *map[string][]byte {
	result := make(map[string][]byte, len(*fields))
	for k, v := range *fields {
		result[k] = util.AnyToBytes(v)
	}
	return &result
}

// 默认ComparisonOperator是like，前缀匹配功能
func (t *Table) Search(fields *map[string]any, ops ...util.ComparisonOperator) (*TableIter, error) {
	return t.Searchs(t.kvStore.Iterator, fields, ops...)
}

/*
	func (t *Table) Searchs(funIter storage.FunIter, fields *map[string]any, ops ...util.ComparisonOperator) *TableIter {
		var tbiter *TableIter
		field := GetStringSlice()
		defer PutStringSlice(field)
		for k := range *fields {
			//判断字段是否在表中
			if _, ok := t.fields[k]; !ok {
				//写错误日志
				//log.Printf("字段 '%s' 不存在于表 '%s'", k, t.name)
				return nil
			}
			field = append(field, k)
		}
		//匹配索引
		idx := t.MatchIndexCached(field) //t.MatchIndex(field...) //
		var key []byte
		var fieldsBytes *map[string][]byte
		if idx != nil {
			fieldsBytes = t.FieldsToBytesNilCached(fields) //t.FieldsToBytesNil(fields) // //t.FieldsToBytesNil(fields)
			key = idx.JoinValue(fieldsBytes, t.id)
		} else {

			return nil
		}
		var op util.ComparisonOperator
		if len(ops) == 0 { //默认是Like操作
			op = util.Like
		} else {
			op = ops[0]
		}
		pfx := idx.Prefix(t.id)
		pfx = append(pfx, SPLIT[0])
		rangeHelper := util.NewRangeHelper(pfx)
		var iter storage.Iterator
		if op != util.NotEqual {
			slice := rangeHelper.FromComparison(op, key)
			iter = funIter(slice.Start, slice.Limit)
			tbiter = TableIterNew(t, iter, idx)
		} else { //不等于将会通过主键或索引进行全表扫描，并且设置跳跃区间
			slice := rangeHelper.FromComparison(util.Like, pfx) //遍历前缀，即通过主键或索引全表扫描
			iter = funIter(slice.Start, slice.Limit)
			tbiter = TableIterNew(t, iter, idx)
			//设置跳跃区间
			neslice := rangeHelper.FromComparison(util.Like, key) //跳跃区间key=0-1-100==>0-1-101
			tbiter.SetJumpRanges(funIter(neslice.Start, neslice.Limit))
		}
		return tbiter
	}
*/
func (t *Table) Searchs(funIter storage.FunIter, fields *map[string]any, ops ...util.ComparisonOperator) (*TableIter, error) {
	var tbiter *TableIter
	field := GetStringSlice()
	defer PutStringSlice(field)
	for k := range *fields {
		//判断字段是否在表中
		if _, ok := t.fields[k]; !ok {
			//写错误日志
			//log.Printf("字段 '%s' 不存在于表 '%s'", k, t.name)
			return nil, fmt.Errorf("字段 '%s' 不存在于表 '%s'", k, t.name)
		}
		field = append(field, k)
	}
	//匹配索引
	idx := t.MatchIndexCached(field) //t.MatchIndex(field...) //
	var key []byte
	var fieldsBytes *map[string][]byte
	if idx != nil {
		fieldsBytes = t.FieldsToBytesNil(fields) // 业务有需要可以开启缓存 FieldsToBytesNilLRU(fields *map[string]any) *map[string][]byte
		key = idx.JoinValue(fieldsBytes, t.id)
	} else {
		/*
			该函数不支持无索引的搜索。
			如果需要支持，可以使用ForData()方法或当前函数设置主键值为nil，则得到遍历全表迭代器，然后配合mach接口自定义匹配规则。
			mach接口自定义匹配规则，理论上可以支持任意查询匹配。
		*/
		return nil, fmt.Errorf("表 '%s' 没有设置索引", t.name)
	}
	var op util.ComparisonOperator
	if len(ops) == 0 { //默认是Like操作
		op = util.Like
	} else {
		op = ops[0]
	}
	pfx := idx.Prefix(t.id)
	pfx = append(pfx, SPLIT[0])
	rangeHelper := util.NewRangeHelper(pfx)
	var iter storage.Iterator
	if op != util.NotEqual {
		slice := rangeHelper.FromComparison(op, key)
		iter = funIter(slice.Start, slice.Limit)
		//tbiter = TableIterNew(t, iter, idx)
		tbiter = GlobalTableIterPool.Get(t, iter, idx)
	} else { //不等于将会通过主键或索引进行全表扫描，并且设置跳跃区间
		slice := rangeHelper.FromComparison(util.Like, pfx) //遍历前缀，即通过主键或索引全表扫描
		iter = funIter(slice.Start, slice.Limit)
		tbiter = GlobalTableIterPool.Get(t, iter, idx)
		//设置跳跃区间
		neslice := rangeHelper.FromComparison(util.Like, key) //跳跃区间key=0-1-100==>0-1-101
		tbiter.SetJumpRanges(funIter(neslice.Start, neslice.Limit))
	}
	return tbiter, nil
}
