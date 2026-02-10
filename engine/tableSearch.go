package engine

import (
	"fmt"
	"time"

	"github.com/liaoran123/sfsDb/storage"
	"github.com/liaoran123/sfsDb/util"
)

// 从按主键数据库读取记录
func (t *Table) Read(fields *map[string]any, timeout ...time.Duration) ([]byte, error) {
	// 获取主键值用于行级锁
	pkField := t.GetPrimaryFields()[0]
	pkValue := (*fields)[pkField]

	// 获取行级共享锁（使用默认事务ID）
	if err := t.acquireRowReadLock(pkValue, 0, timeout...); err != nil {
		return nil, err
	}
	// 直接使用 RUnlock 释放读锁
	lockKey := fmt.Sprintf("%v", pkValue)
	defer func() {
		if rowLock, ok := t.rowLocks.Load(lockKey); ok {
			rl := rowLock.(*RowLock)
			rl.rwLock.RUnlock()
		}
	}()

	fieldsBytes := t.FieldsToBytes(fields)
	defer func() {
		if fieldsBytes != nil && *fieldsBytes != nil {
			GlobalFieldsBytesPool.Put(*fieldsBytes)
		}
	}()
	key := t.GetPrimaryKey().JoinValue(fieldsBytes, t.id)
	return t.ReadByBytes(key), nil
}

// ReadWithTimeout 带超时的读取方法
func (t *Table) ReadWithTimeout(fields *map[string]any, timeout time.Duration) ([]byte, error) {
	return t.Read(fields, timeout)
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
	result := GlobalFieldsBytesPool.Get()
	// 直接使用从对象池获取的 map，Go 会自动处理 map 的扩容
	for k, v := range *fields {
		result[k] = util.AnyToBytes(v)
	}
	return &result
}

// 默认ComparisonOperator是like，前缀匹配功能
func (t *Table) Search(fields *map[string]any, ops ...util.ComparisonOperator) (*TableIter, error) {
	return t.Searchs(t.kvStore.Iterator, fields, ops...)
}

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
		// 使用完后将 fieldsBytes 放回对象池
		defer func() {
			if fieldsBytes != nil && *fieldsBytes != nil {
				GlobalFieldsBytesPool.Put(*fieldsBytes)
			}
		}()
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
