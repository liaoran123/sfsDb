package engine

import (
	"fmt"
	"log"

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
	primaryFields := t.GetPrimaryKey().GetFields()
	if len(primaryFields) == 0 {
		return 0, fmt.Errorf("表 '%s' 没有设置主键", t.name)
	}

	//是否支持默认自动增值主键，单主键并且主键字段名为"id"
	pklen := len(primaryFields)
	pkfield := primaryFields[0]
	supportDefault := pklen == 1 && pkfield == "id"

	if supportDefault {
		// 检查是否提供了主键字段
		//使用默认自动增值主键时，不需要提供主键字段，系统自动生成
		_, ok := (*fields)[pkfield]
		if !ok { //未提供主键字段，自动生成主键值
			currentID = t.GetAutoInc()
			(*fields)[pkfield] = currentID
		} else { //提供了主键字段，但是值为nil，自动生成主键值
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
}

// 删除记录
// fields *map[string]any 主键值，可能是组合主键
func (t *Table) Delete(fields *map[string]any, batchs ...storage.Batch) error {
	if t.fields == nil {
		return fmt.Errorf("表 '%s' 未设置字段和类型", t.name)
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
	//反序列化记录，并且将字段值转换为对应的类型
	//fieldsBytes := t.ParseRecord(record)
	pk := t.GetPrimaryKey()
	fieldsBytes, err := pk.Parse(t.fieldsid, record)
	if err != nil {
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
	//检查是否提供了所有主键字段
	for _, field := range t.GetPrimaryKey().GetFields() {
		if _, ok := (*fields)[field]; !ok {
			return nil, fmt.Errorf("删除操作必须提供主键字段 '%s'", field)
		}
	}
	fieldsBytes := t.FieldsToBytes(fields)
	key := t.GetPrimaryKey().JoinValue(fieldsBytes, t.id)
	return t.ReadByBytes(key), nil
}

// 更新记录，不支持修改主键字段
// fields *map[string]any 主键值，可能是组合主键
func (t *Table) Update(fields *map[string]any, batchs ...storage.Batch) error {
	if t.fields == nil {
		return fmt.Errorf("表 '%s' 未设置字段和类型", t.name)
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
	var updateFields []string
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
	BatchContainer := NewBatchContainer(batch, t.indexs, t.id, t.kvStore)
	//删除
	BatchContainer.Operation(fieldsBytes, updateFields...)

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

// 遍历表所有kv，复制表用
func (t *Table) For() storage.Iterator {
	pfx := []byte{byte(t.id), SPLIT[0]}
	rangeHelper := util.NewRangeHelper(pfx)
	slice := rangeHelper.FromComparison(util.Like, pfx)
	return t.kvStore.Iterator(slice.Start, slice.Limit)
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
func (t *Table) Search(fields *map[string]any, ops ...util.ComparisonOperator) *TableIter {
	var field []string
	for k := range *fields {
		//判断字段是否在表中
		if _, ok := t.fields[k]; !ok {
			//写错误日志
			log.Printf("字段 '%s' 不存在于表 '%s'", k, t.name)
			return nil
		}
		field = append(field, k)
	}
	//匹配索引
	idx := t.MatchIndex(field...)
	var key []byte
	var fieldsBytes *map[string][]byte
	if idx != nil {
		fieldsBytes = t.FieldsToBytesNil(fields)
		key = idx.JoinValue(fieldsBytes, t.id)
	} else {
		/*
			该函数不支持无索引的搜索。
			如果需要支持，可以使用ForData()方法或当前函数设置主键值为nil，则得到遍历全表迭代器，然后配合mach接口自定义匹配规则。
			mach接口自定义匹配规则，理论上可以支持任意查询匹配。
		*/
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
	var tbiter *TableIter
	if op != util.NotEqual {
		slice := rangeHelper.FromComparison(op, key)
		iter = t.kvStore.Iterator(slice.Start, slice.Limit)
		tbiter = TableIterNew(t, iter, idx)
	} else { //不等于将会通过主键或索引进行全表扫描，并且设置跳跃区间
		slice := rangeHelper.FromComparison(util.Like, pfx) //遍历前缀，即通过主键或索引全表扫描
		iter = t.kvStore.Iterator(slice.Start, slice.Limit)
		tbiter = TableIterNew(t, iter, idx)
		//设置跳跃区间
		neslice := rangeHelper.FromComparison(util.Like, key) //跳跃区间key=0-1-100==>0-1-101
		tbiter.SetJumpRanges(t.kvStore.Iterator(neslice.Start, neslice.Limit))
	}
	return tbiter
}
