package engine

import (
	"fmt"
	"time"

	"github.com/liaoran123/sfsDb/storage"
	"github.com/liaoran123/sfsDb/util"
)

// 从按主键数据库读取记录
func (t *Table) Read(fields *map[string]any, timeout ...time.Duration) ([]byte, error) {
	// 解析超时参数
	var duration time.Duration
	if len(timeout) > 0 {
		duration = timeout[0]
	}

	// 使用 SearchImpl
	searchImpl := NewSearchImpl(t, fields, nil, duration, nil)

	// 读取记录
	record, err := searchImpl.Read()
	if err != nil {
		return nil, err
	}

	// 归还对象池
	GlobalSearchImplPool.Put(searchImpl)

	return record, nil
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
	// 使用 SearchImpl
	searchImpl := NewSearchImpl(t, fields, ops, 0, funIter)

	// 搜索记录
	tbiter, err := searchImpl.Search()
	if err != nil {
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
func (t *Table) SearchRange(funIter storage.FunIter, fieldname string, Start, Limit any) (*TableIter, error) {
	// 使用 SearchImpl
	searchImpl := NewSearchImpl(t, nil, nil, 0, funIter)

	// 范围搜索
	tbiter, err := searchImpl.SearchRange(fieldname, Start, Limit)
	if err != nil {
		return nil, err
	}

	// 归还对象池
	GlobalSearchImplPool.Put(searchImpl)

	return tbiter, nil
}

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
