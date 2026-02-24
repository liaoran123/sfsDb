package engine

import (
	"fmt"

	"github.com/liaoran123/sfsDb/storage"
	"github.com/liaoran123/sfsDb/util"
)

// 从按主键数据库读取记录
func (t *Table) Read(fields *map[string]any) ([]byte, error) {
	fieldsBytes := t.FieldsToBytes(fields)
	defer GlobalFieldsBytesPool.Put(*fieldsBytes)
	key := t.GetPrimaryKey().JoinValue(fieldsBytes, t.id)
	record := t.ReadByBytes(key)
	if record == nil {
		return nil, fmt.Errorf("主键值 '%v' 的记录不存在", fields)
	}
	return record, nil
}

/*
// 从按主键数据库读取记录
func (t *Table) Read(fields *map[string]any) ([]byte, error) {

	// 使用 SearchImpl
	searchImpl := NewSearchImpl(t)
	// 读取记录
	record := searchImpl.Read(fields)
	if record == nil {
		GlobalSearchImplPool.Put(searchImpl)
		return nil, fmt.Errorf("主键值 '%v' 的记录不存在", fields)
	}
	// 归还对象池
	GlobalSearchImplPool.Put(searchImpl)
	return record, nil
}
*/
// 从按主键数据库读取记录
func (t *Table) ReadByBytes(key []byte) []byte {
	v, err := t.kvStore.Get(key)
	if err != nil {
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
	return t.Searchs(nil, fields, ops...)
}

func (t *Table) Searchs(funIter storage.FunIter, fields *map[string]any, ops ...util.ComparisonOperator) (*TableIter, error) {
	// 使用 SearchImpl
	searchImpl := NewSearchImpl(t)

	// 搜索记录
	tbiter, err := searchImpl.Searchs(funIter, fields, ops...)
	if err != nil {
		GlobalSearchImplPool.Put(searchImpl)
		return nil, err
	}

	// 归还对象池
	GlobalSearchImplPool.Put(searchImpl)

	return tbiter, nil
}

// 区间搜索
// fieldname 字段名
// Start, Limit 区间开始值和结束值。Start=nil表示从索引最小值开始，Limit=nil表示到索引最大值结束。同时为nil即表示遍历索引。
// 索引为主键时，Start=nil表示从表的主键值最小值开始，Limit=nil表示到表的主键值最大值结束。同时为nil即表示遍历全表。
// funIter 区间迭代器
func (t *Table) SearchRange(funIter storage.FunIter, Start, Limit *map[string]any) (*TableIter, error) {
	// 使用 SearchImpl
	searchImpl := NewSearchImpl(t)

	if funIter == nil { //默认是原始数据库的迭代器
		funIter = t.kvStore.Iterator
	}
	// 范围搜索
	tbiter, err := searchImpl.SearchRange(funIter, Start, Limit)
	if err != nil {
		GlobalSearchImplPool.Put(searchImpl)
		return nil, err
	}

	// 归还对象池
	GlobalSearchImplPool.Put(searchImpl)

	return tbiter, nil
}

/*
// 区间迭代器,用于范围搜索和跳跃区间
func (t *Table) RangeForAny(funIter storage.FunIter, fieldname string, Start, Limit any) (storage.Iterator, Index, error) {
	if funIter == nil {
		funIter = t.kvStore.Iterator
	}
	idx := t.MatchIndexCached([]string{fieldname})
	if idx == nil {
		return nil, nil, fmt.Errorf("字段 '%s' 不存在于表 '%s'", fieldname, t.name)
	}
	pfx := idx.Prefix(t.id)
	pfx = append(pfx, SPLIT[0])
	// 处理Start参数
	var startBytes []byte
	if Start != nil {
		startBytes = util.AnyToBytes(Start)
	}
	// 处理Limit参数
	var limitBytes []byte
	if Limit != nil {
		limitBytes = util.AnyToBytes(Limit)
	}
	// 创建范围对象
	slice := &util.Range{
		Start: startBytes,
		Limit: limitBytes,
	}
	// 构建完整的搜索范围
	slice.Start = append(pfx, slice.Start...)
	if Limit == nil {
		// 当Limit为nil时，使用前缀的下一个字节作为上限，表示到无穷大
		slice.Limit = util.BytesPrefix(pfx).Limit
	} else {
		// 当Limit不为nil时，构建完整的上限字节
		slice.Limit = append(pfx, slice.Limit...)
	}
	iter := funIter(slice.Start, slice.Limit)
	if iter == nil {
		return nil, nil, fmt.Errorf("区间迭代器不能为空")
	}
	return iter, idx, nil
}
*/
