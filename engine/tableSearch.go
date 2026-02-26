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
	record, err := t.ReadByBytes(key)
	if err != nil {
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
func (t *Table) ReadByBytes(key []byte) ([]byte, error) {
	v, err := t.kvStore.Get(key)
	if err != nil {
		return nil, err
	}
	return v, nil
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
	var tbiter *TableIter
	field := GetStringSlice()
	defer PutStringSlice(field)
	for k := range *fields {
		//判断字段是否在表中
		if _, ok := t.fields[k]; !ok {
			return nil, fmt.Errorf("字段 '%s' 不存在于表 '%s'", k, t.name)
		}
		field = append(field, k)
	}
	//匹配索引
	idx := t.MatchIndexCached(field)
	var key []byte
	var fieldsBytes *map[string][]byte
	if idx != nil {
		fieldsBytes = t.FieldsToBytesNil(fields)
		key = idx.JoinValue(fieldsBytes, t.id)
		// 使用完后将 fieldsBytes 放回对象池
		defer GlobalFieldsBytesPool.Put(*fieldsBytes)
	} else {
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
	if funIter == nil {
		funIter = t.kvStore.Iterator
	}
	var iter storage.Iterator
	if op != util.NotEqual {
		slice := rangeHelper.FromComparison(op, key)
		iter = funIter(slice.Start, slice.Limit)
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

// SearchRange 范围搜索，区间搜索
// fieldname 对应索引字段名，单主键或组合主键，单索引或组合索引字段。
// 组合主键或索引时，Start, Limit any为最后一个字段的范围值。第一个字段必须全量匹配，也就是后缀匹配。
//  5. 当Limit的最后一个字段值为nil时，表示搜索到该前缀的最大值
//  6. 搜索基于索引进行，必须存在匹配的索引
//
// 该功能实现了比sql的between查询更高效强大的范围搜索
func (t *Table) SearchRange(funIter storage.FunIter, Start, Limit *map[string]any) (*TableIter, error) {
	/*
		if Start == nil || Limit == nil {
			return nil, fmt.Errorf("Start, Limit 不能为空")
		}
		if funIter == nil {
			funIter = t.kvStore.Iterator
		}
		//判断Start, Limit 的 value值是否相同
		for k := range *Start {
			if _, ok := (*Limit)[k]; !ok {
				return nil, fmt.Errorf("Start, Limit 的 key值不相同")
			}
		}
		fieldname := GetStringSlice()
		defer PutStringSlice(fieldname)
		for k := range *Start {
			fieldname = append(fieldname, k)
		}
		idx := t.MatchIndexCached(fieldname)
		if idx == nil {
			return nil, fmt.Errorf("字段 '%s' 不存在于表 '%s'", fieldname, t.name)
		}
		var Startkey []byte
		var Limitkey []byte
		var fieldsBytes *map[string][]byte

		fieldsBytes = t.FieldsToBytesNil(Start)
		Startkey = idx.JoinValue(fieldsBytes, t.id)
		fieldsBytes = t.FieldsToBytesNil(Limit)
		Limitkey = idx.JoinValue(fieldsBytes, t.id)
		// 使用完后将 fieldsBytes 放回对象池
		defer func() {
			if fieldsBytes != nil && *fieldsBytes != nil {
				GlobalFieldsBytesPool.Put(*fieldsBytes)
			}
		}()

		pfx := idx.Prefix(t.id)
		pfx = append(pfx, SPLIT[0])
		rangeHelper := util.NewRangeHelper(pfx)

		Startslice := rangeHelper.FromComparison(util.Equal, Startkey)
		LimitSlice := rangeHelper.FromComparison(util.Equal, Limitkey)
		slice := &util.Range{
			Start: Startslice.Start,
			Limit: LimitSlice.Limit,
		}
		if (*Limit)[fieldname[len(fieldname)-1]] == nil { //// 当Limit最后一个字段值为nil时，使用前缀的下一个字节作为上限，表示到无穷大
			slice.Limit = util.BytesPrefix(pfx).Limit
		}
		iter := funIter(slice.Start, slice.Limit)
	*/
	iter, idx, err := t.RangeForAny(funIter, Start, Limit)
	if err != nil {
		return nil, err
	}
	tbiter := GlobalTableIterPool.Get(t, iter, idx)
	return tbiter, nil
}

// 区间迭代器,用于范围搜索和*TableIter的跳跃区间
func (t *Table) RangeForAny(funIter storage.FunIter, Start, Limit *map[string]any) (storage.Iterator, Index, error) {
	if Start == nil || Limit == nil {
		return nil, nil, fmt.Errorf("Start, Limit 不能为空")
	}
	if funIter == nil {
		funIter = t.kvStore.Iterator
	}
	//判断Start, Limit 的 value值是否相同
	for k := range *Start {
		if _, ok := (*Limit)[k]; !ok {
			return nil, nil, fmt.Errorf("Start, Limit 的 key值不相同")
		}
	}
	fieldname := GetStringSlice()
	defer PutStringSlice(fieldname)
	for k := range *Start {
		fieldname = append(fieldname, k)
	}
	idx := t.MatchIndexCached(fieldname)
	if idx == nil {
		return nil, nil, fmt.Errorf("字段 '%s' 不存在于表 '%s'", fieldname, t.name)
	}
	var Startkey []byte
	var Limitkey []byte
	var fieldsBytes *map[string][]byte

	fieldsBytes = t.FieldsToBytesNil(Start)
	Startkey = idx.JoinValue(fieldsBytes, t.id)
	fieldsBytes = t.FieldsToBytesNil(Limit)
	Limitkey = idx.JoinValue(fieldsBytes, t.id)
	// 使用完后将 fieldsBytes 放回对象池
	defer func() {
		if fieldsBytes != nil && *fieldsBytes != nil {
			GlobalFieldsBytesPool.Put(*fieldsBytes)
		}
	}()

	pfx := idx.Prefix(t.id)
	pfx = append(pfx, SPLIT[0])
	rangeHelper := util.NewRangeHelper(pfx)

	Startslice := rangeHelper.FromComparison(util.Equal, Startkey)
	LimitSlice := rangeHelper.FromComparison(util.Equal, Limitkey)
	slice := &util.Range{
		Start: Startslice.Start,
		Limit: LimitSlice.Limit,
	}
	if (*Limit)[fieldname[len(fieldname)-1]] == nil { //// 当Limit最后一个字段值为nil时，使用前缀的下一个字节作为上限，表示到无穷大
		slice.Limit = util.BytesPrefix(pfx).Limit
	}
	iter := funIter(slice.Start, slice.Limit)
	return iter, idx, nil
}
