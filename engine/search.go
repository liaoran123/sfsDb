package engine

import (
	"fmt"
	"sync"

	"github.com/liaoran123/sfsDb/storage"
	"github.com/liaoran123/sfsDb/util"
)

// SearchImplPool 是 SearchImpl 的对象池
type SearchImplPool struct {
	pool sync.Pool
}

// 全局 SearchImpl 对象池
var GlobalSearchImplPool = &SearchImplPool{
	pool: sync.Pool{
		New: func() interface{} {
			return &SearchImpl{}
		},
	},
}

// Get 从对象池中获取一个 SearchImpl 实例
func (p *SearchImplPool) Get() *SearchImpl {
	return p.pool.Get().(*SearchImpl)
}

// Put 将 SearchImpl 实例放回对象池
func (p *SearchImplPool) Put(impl *SearchImpl) {
	// 重置实例状态
	impl.Reset()
	p.pool.Put(impl)
}

// 提供了一个统一的搜索流程接口
type Search interface {
	// 读取记录
	Read() ([]byte, error)
	// 搜索记录
	Search() (*TableIter, error)
	// 范围搜索
	SearchRange(fieldname string, Start, Limit any) (*TableIter, error)
}

// 批量搜索接口
type BatchSearch interface {
	Search
	// 批量读取记录
	BatchRead(records []*map[string]any) (map[any][]byte, error)
	// 批量搜索记录
	BatchSearch(records []*map[string]any) (map[any]*TableIter, error)
}

type SearchImpl struct {
	table       *Table
	fields      *map[string]any
	fieldsBytes *map[string][]byte
	ops         []util.ComparisonOperator
	funIter     storage.FunIter
	// 批量操作相关字段
	records []*map[string]any
}

// Reset 重置 SearchImpl 实例的状态
func (s *SearchImpl) Reset() {
	s.table = nil
	s.fields = nil
	s.fieldsBytes = nil
	s.ops = nil
	s.funIter = nil
	s.records = nil
}

// NewSearchImpl 创建一个新的 SearchImpl 实例
func NewSearchImpl(table *Table, fields *map[string]any, ops []util.ComparisonOperator, funIter storage.FunIter) *SearchImpl {
	impl := GlobalSearchImplPool.Get()
	impl.table = table
	impl.fields = fields
	impl.ops = ops
	impl.funIter = funIter
	if impl.funIter == nil {
		impl.funIter = table.kvStore.Iterator
	}
	return impl
}

// NewBatchSearchImpl 创建一个新的用于批量搜索的 SearchImpl 实例
func NewBatchSearchImpl(table *Table, records []*map[string]any) *SearchImpl {
	impl := GlobalSearchImplPool.Get()
	impl.table = table
	impl.records = records
	impl.funIter = table.kvStore.Iterator
	return impl
}

// Read 读取记录
func (s *SearchImpl) Read() ([]byte, error) {
	// 获取主键值用于行级锁
	pkField := s.table.GetPrimaryFields()[0]
	_ = (*s.fields)[pkField]

	// 获取行级共享锁（使用默认事务ID）

	fieldsBytes := s.table.FieldsToBytes(s.fields)
	defer func() {
		if fieldsBytes != nil && *fieldsBytes != nil {
			GlobalFieldsBytesPool.Put(*fieldsBytes)
		}
	}()
	key := s.table.GetPrimaryKey().JoinValue(fieldsBytes, s.table.id)
	return s.table.ReadByBytes(key), nil
}

// Search 搜索记录
func (s *SearchImpl) Search() (*TableIter, error) {
	var tbiter *TableIter
	field := GetStringSlice()
	defer PutStringSlice(field)
	for k := range *s.fields {
		//判断字段是否在表中
		if _, ok := s.table.fields[k]; !ok {
			return nil, fmt.Errorf("字段 '%s' 不存在于表 '%s'", k, s.table.name)
		}
		field = append(field, k)
	}
	//匹配索引
	idx := s.table.MatchIndexCached(field)
	var key []byte
	var fieldsBytes *map[string][]byte
	if idx != nil {
		fieldsBytes = s.table.FieldsToBytesNil(s.fields)
		key = idx.JoinValue(fieldsBytes, s.table.id)
		// 使用完后将 fieldsBytes 放回对象池
		defer func() {
			if fieldsBytes != nil && *fieldsBytes != nil {
				GlobalFieldsBytesPool.Put(*fieldsBytes)
			}
		}()
	} else {
		return nil, fmt.Errorf("表 '%s' 没有设置索引", s.table.name)
	}
	var op util.ComparisonOperator
	if len(s.ops) == 0 { //默认是Like操作
		op = util.Like
	} else {
		op = s.ops[0]
	}
	pfx := idx.Prefix(s.table.id)
	pfx = append(pfx, SPLIT[0])
	rangeHelper := util.NewRangeHelper(pfx)
	var iter storage.Iterator
	if op != util.NotEqual {
		slice := rangeHelper.FromComparison(op, key)
		iter = s.funIter(slice.Start, slice.Limit)
		tbiter = GlobalTableIterPool.Get(s.table, iter, idx)
	} else { //不等于将会通过主键或索引进行全表扫描，并且设置跳跃区间
		slice := rangeHelper.FromComparison(util.Like, pfx) //遍历前缀，即通过主键或索引全表扫描
		iter = s.funIter(slice.Start, slice.Limit)
		tbiter = GlobalTableIterPool.Get(s.table, iter, idx)
		//设置跳跃区间
		neslice := rangeHelper.FromComparison(util.Like, key) //跳跃区间key=0-1-100==>0-1-101
		tbiter.SetJumpRanges(s.funIter(neslice.Start, neslice.Limit))
	}
	return tbiter, nil
}

// SearchRange 范围搜索
func (s *SearchImpl) SearchRange(fieldname string, Start, Limit any) (*TableIter, error) {
	iter, idx, err := s.table.RangeForAny(s.funIter, fieldname, Start, Limit)
	if err != nil {
		return nil, err
	}
	tbiter := GlobalTableIterPool.Get(s.table, iter, idx)
	if tbiter == nil {
		return nil, fmt.Errorf("TableIter为nil")
	}
	return tbiter, nil
}

// BatchRead 批量读取记录
func (s *SearchImpl) BatchRead(records []*map[string]any) (map[any][]byte, error) {
	// 检查参数
	if len(records) == 0 {
		return map[any][]byte{}, nil
	}
	if records == nil {
		return nil, fmt.Errorf("records cannot be nil")
	}

	// 保存记录
	s.records = records

	// 预分配结果映射
	results := make(map[any][]byte, len(records))

	// 处理记录并批量读取
	for _, fields := range records {
		// 创建临时 SearchImpl 实例处理单条记录
		searchImpl := NewSearchImpl(s.table, fields, nil, s.funIter)

		// 读取记录
		record, err := searchImpl.Read()
		if err != nil {
			return nil, err
		}

		// 获取主键值作为结果映射的键
		pkField := s.table.GetPrimaryFields()[0]
		pkValue := (*fields)[pkField]
		results[pkValue] = record
	}

	// 归还对象池
	GlobalSearchImplPool.Put(s)

	return results, nil
}

// BatchSearch 批量搜索记录
func (s *SearchImpl) BatchSearch(records []*map[string]any) (map[any]*TableIter, error) {
	// 检查参数
	if len(records) == 0 {
		return map[any]*TableIter{}, nil
	}
	if records == nil {
		return nil, fmt.Errorf("records cannot be nil")
	}

	// 保存记录
	s.records = records

	// 预分配结果映射
	results := make(map[any]*TableIter, len(records))

	// 处理记录并批量搜索
	for _, fields := range records {
		// 创建临时 SearchImpl 实例处理单条记录
		searchImpl := NewSearchImpl(s.table, fields, nil, s.funIter)

		// 搜索记录
		tbiter, err := searchImpl.Search()
		if err != nil {
			return nil, err
		}

		// 获取主键值作为结果映射的键
		pkField := s.table.GetPrimaryFields()[0]
		pkValue := (*fields)[pkField]
		results[pkValue] = tbiter
	}

	// 归还对象池
	GlobalSearchImplPool.Put(s)

	return results, nil
}
