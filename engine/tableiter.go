package engine

import (
	"sync"

	"github.com/liaoran123/sfsDb/storage"
	"github.com/liaoran123/sfsDb/util"
)

// 基础迭代器结构体，提取公共字段和方法
type baseIter struct {
	table      *Table
	showfields []string
	index      Index //搜索时使用的索引
	move       map[bool]func() bool
	top        map[bool]func() bool
	mu         sync.Mutex
}

// 实现sql语句中的select f0,f1,... from table 要返回的字段
// 如果keys为空，则返回所有字段
// 在最底层转换，最大化减少内存占用
type TableIter struct {
	baseIter
	iter storage.Iterator
}

// 解析函数
type Parser func(k, v []byte) Record

// 导出函数
type Export func(k, v []byte) bool

func TableIterNew(table *Table, iter storage.Iterator, index Index, showfields ...string) *TableIter {
	return &TableIter{
		baseIter: baseIter{
			table:      table,
			showfields: showfields,
			index:      index,
			move: map[bool]func() bool{
				true:  iter.Next,
				false: iter.Prev,
			},
			top: map[bool]func() bool{
				true:  iter.First,
				false: iter.Last,
			},
		},
		iter: iter,
	}
}
func (t *TableIter) Release() {
	t.iter.Release()
}

// 将主键转为字符串 ，主要是用于组合主键需要进行map比较时使用
func (t *TableIter) PrimaryToStr(k []byte) string {
	/*
		isComposite := len(t.table.primary) > 1
		if !isComposite {
			return ""
		}*/
	var key any
	rstr := ""
	for _, p := range t.index.GetFields() {
		key = util.Bytes(k).ToAny(t.table.fields[p])
		rstr += string(util.AnyToBytes(key)) + util.SPLIT
	}
	return rstr[:len(rstr)-1]
}

// 判断是否存在指定的主键记录
func (t *TableIter) Exist() bool {
	return t.iter.First()
}

// 统计索引记录数
func (t *TableIter) Count() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	i := 0
	for t.iter.Next() {
		i++
	}
	return i
}

// 分页变量
type Page struct {
	Start int
	Count int
}

func PageNew(No ...int) Page {
	start := 0
	count := -1
	switch len(No) {
	case 0:
	case 1:
		count = No[0]
	default:
		start = No[0]
		count = No[1]
	}
	return Page{
		Start: start,
		Count: count,
	}
}

// 通过二级索引解析索引记录
// 组合主键需要反转义
// 组合主键时，所有索引对应的主键值都是经过转义的。
// 这是为了提取主键中的某个字段值作为匹配索引，必须转义才能正确根据转义符分割提取。
// 所以在用索引的主键值回表记录时，需要对索引值进行反转义。
func (t *TableIter) ByIndexGetRecord(v []byte) Record {
	value := v
	if len(t.table.primaryKey.GetFields()) > 1 { //只有组合主键才需要转义
		value = util.Bytes(v).UnEscape()
	}
	// 读取完整记录
	byrecord := t.table.ReadByBytes(value)
	if byrecord == nil {
		return nil
	}
	// 解析记录并提取指定字段
	rd := Record(t.table.ParseRecordValue(byrecord))
	return rd
}

// 通过主键获取记录
// ByPrimaryGet=ByPrimary.Get 两种方式都可
func (t *TableIter) ByPrimaryGetRecord(v []byte) Record {
	rd := Record(t.table.ParseRecordValue(v))
	return rd
}

// 还原全文索引的value原始值，专用于全文索引。
// 全文索引设计索引时，规定必须在末尾将主键全量追加到key的键值中。为了避免重复储存字段值，设置value键值为空。
func (t *TableIter) RestoreValByIndex(k []byte) (Value []byte) {
	ks := util.Bytes(t.iter.Key()).Split() //分解并反转义
	fieldlen := len(t.table.primaryKey.GetFields())
	ks = ks[len(ks)-fieldlen:]
	if len(ks) > 1 { //组合主键,分解反转义后再重新拼接
		for _, k := range ks { //重新拼接
			Value = append(Value, k...)
			Value = append(Value, []byte(util.SPLIT)...)
		}
		Value = Value[:len(Value)-1]
	} else { //单主键
		Value = ks[0]
	}
	return
}

// 遍历迭代器返回解析后的记录
// 单个简单查询直接使用
// handler可以对value进行各种处理
func (t *TableIter) GerRecords(esc bool, limit ...int) (r Records) {
	//添加锁，防止并发访问
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.top[esc]() {
		return nil
	}
	var rd Record

	page := PageNew(limit...)

	if page.Count > 0 {
		r = make(Records, 0, page.Count)
	}

	var Value []byte
	loop := 0
	// 处理当前位置的元素
	for {
		if loop < page.Start {
			loop++
		} else {
			if len(t.iter.Value()) == 0 { //这种情况是全文索引，value值为空，需要从key中提取主键，末尾就是对应主键的值
				Value = t.RestoreValByIndex(t.iter.Key())
			} else {
				Value = t.iter.Value()
			}
			switch t.index.(type) { //需要区别主键和索引，根据handler命名规则
			case PrimaryKey:
				rd = t.ByPrimaryGetRecord(Value)
			default:
				rd = t.ByIndexGetRecord(Value)
			}
			rd = rd.GetKeys(t.showfields...)
			r = append(r, rd)
			if page.Count > 0 && len(r) >= page.Count {
				break
			}
		}
		// 移动到下一个/前一个元素
		if !t.move[esc]() {
			break
		}
	}
	return r
}

// 遍历迭代器返回解析后的记录
// 复杂组合查询使用
func (t *TableIter) ForMatchRecord(esc bool, page Page, match ...MatchKeyValue) (r Records) {
	//添加锁，防止并发访问
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.top[esc]() {
		return nil
	}
	var rd Record
	if page.Count > 0 {
		r = make(Records, 0, page.Count)
	}
	loop := 0
	// 处理当前位置的元素
	for {
		if loop < page.Start {
			loop++
		} else {
			for _, m := range match {
				// 当前的key,value所有能得到的field的值组织成map[string]any，传入MatchKeyValue进行所需匹配
				if m.Match(nil) {
					r = append(r, rd)
					if page.Count > 0 && len(r) >= page.Count {
						break
					}
				}
			}
		}
		// 移动到下一个/前一个元素
		if !t.move[esc]() {
			break
		}
	}
	return r
}

// 遍历迭代器导出数据
// export导出函数，返回false则停止导出
func (t *TableIter) ForExport(esc bool, export Export) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.top[esc]() {
		return
	}
	for {
		if !export(t.iter.Key(), t.iter.Value()) {
			return
		}
		if !t.move[esc]() {
			return
		}
	}
}

// 提取转换主键值
// 单主键则返回转换的值，组合主键则返回字符串
// field用于对组合主键提取其中的部分值
func (t *TableIter) GetPrimaryKeys(field ...string) (r any) {
	//是否组合主键
	isComposite := len(t.table.primaryKey.GetFields()) > 1
	switch isComposite {
	case true:
		// 组合主键，返回字符串
		keys := util.Bytes(t.iter.Key()).Split()
		if len(field) > 0 {
			newkeys := ""
			for _, f := range field {
				for i, p := range t.table.primaryKey.GetFields() {
					if f == p {
						newkeys += string(keys[i]) + util.SPLIT
					}
				}
			}
			r = newkeys[:len(newkeys)-1]
		}
	default:
		// 单主键，返回转换的值
		r = util.Bytes(t.iter.Value()).ToAny(t.table.fields[t.table.primaryKey.GetFields()[0]])
	}
	return r
}
func (t *TableIter) First() bool {
	return t.iter.First()
}
func (t *TableIter) Last() bool {
	return t.iter.Last()
}
func (t *TableIter) Next() bool {
	return t.iter.Next()
}
func (t *TableIter) Prev() bool {
	if t.iter.Prev() {
		return true
	}
	return false
}
func (t *TableIter) Key() []byte {
	return t.iter.Key()
}
func (t *TableIter) Value() []byte {
	return t.iter.Value()
}
