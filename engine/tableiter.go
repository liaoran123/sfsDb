package engine

import (
	"bytes"
	"sync"

	"github.com/liaoran123/sfsDb/engine/record"
	"github.com/liaoran123/sfsDb/storage"
	"github.com/liaoran123/sfsDb/util"
)

// 实现sql语句中的select f0,f1,... from table 要返回的字段
// 如果keys为空，则返回所有字段
// 在最底层转换，最大化减少内存占用
type TableIter struct {
	//baseIter
	iter         storage.Iterator
	table        *Table
	selectFields []string
	index        Index //搜索时使用的索引
	move         map[bool]func() bool
	top          map[bool]func() bool
	mu           sync.Mutex
}

// 解析函数
type Parser func(k, v []byte) record.Record

// 导出函数
type Export func(k, v []byte) bool

func TableIterNew(table *Table, iter storage.Iterator, index Index, selectFields ...string) *TableIter {
	return &TableIter{
		table:        table,
		selectFields: selectFields,
		index:        index,
		move: map[bool]func() bool{
			true:  iter.Next,
			false: iter.Prev,
		},
		top: map[bool]func() bool{
			true:  iter.First,
			false: iter.Last,
		},

		iter: iter,
	}
}
func (t *TableIter) Release() {
	t.iter.Release()
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
// 这是为了提取组合主键中的某个字段值作为匹配索引的功能，必须转义才能正确根据转义符分割提取。
// 所以在用索引的主键值回表记录时，需要对索引值进行反转义。
func (t *TableIter) ByIndexGetRecord(v []byte) record.Record {
	value := v
	if len(t.table.indexs.GetPrimaryKey().GetFields()) > 1 { //只有组合主键才需要转义
		value = util.Bytes(v).UnEscape()
	}
	//获取并且拼接前缀得到key值
	pfx := t.table.indexs.GetPrimaryKey().Prefix(t.table.name)
	key := bytes.Join([][]byte{pfx, value}, []byte(SPLIT))
	// 读取完整记录
	byrecord := t.table.ReadByBytes(key)
	if byrecord == nil {
		return nil
	}
	// 解析记录并提取指定字段
	fields := t.table.ParseRecordValue(byrecord)
	if fields == nil {
		return nil
	}
	rd := record.Record(*fields)
	return rd
}

// 通过主键获取记录
// ByPrimaryGet=ByPrimary.Get 两种方式都可
func (t *TableIter) ByPrimaryGetRecord(v []byte) record.Record {
	fields := t.table.ParseRecordValue(v)
	if fields == nil {
		return nil
	}
	rd := record.Record(*fields)
	return rd
}

// 还原全文索引的value原始值，专用于全文索引。
// 全文索引设计索引时，规定必须在末尾将主键全量追加到key的键值中。为了避免重复储存字段值，设置value键值为空。
func (t *TableIter) RestoreValByIndex(k, v []byte) (Value []byte) {
	//从key中提取主键
	fieldsBytes := t.ParseBytes(k, v)
	Value = t.table.indexs.GetPrimaryKey().GetID(fieldsBytes)
	return Value
}
func (t *TableIter) ParseBytes(k, v []byte) *map[string][]byte {
	var fieldsBytes *map[string][]byte
	switch t.index.(type) {
	case FullTextIndex:
		//全文索引时，value值为空，需要从key中提取主键
		fieldsBytes = t.index.Parse(t.table.indexs.GetPrimaryKey().GetFields(), k)
	case PrimaryKey:
		fieldsBytes = t.index.Parse(nil, v)
	default:
		fieldsBytes = t.index.Parse(t.table.indexs.GetPrimaryKey().GetFields(), v)
	}
	return fieldsBytes
}

// 遍历迭代器返回解析后的记录
// 单个简单查询直接使用
func (t *TableIter) GerRecords(esc bool, limit ...int) (r record.Records) {
	//添加锁，防止并发访问
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.top[esc]() {
		return nil
	}
	var rd record.Record

	page := PageNew(limit...)

	if page.Count > 0 {
		r = make(record.Records, 0, page.Count)
	}

	var Value []byte
	loop := 0
	// 处理当前位置的元素
	for {
		if loop < page.Start {
			loop++
		} else {
			if len(t.iter.Value()) == 0 { //这种情况是全文索引，value值为空，需要从key中提取主键，末尾就是对应主键的值
				Value = t.RestoreValByIndex(t.iter.Key(), t.iter.Value())
			} else {
				Value = t.iter.Value()
			}
			switch t.index.(type) { //需要区别主键和索引，根据handler命名规则
			case PrimaryKey:
				rd = t.ByPrimaryGetRecord(Value)
			default:
				rd = t.ByIndexGetRecord(Value)
			}
			rd = rd.GetKeys(t.selectFields...)
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
// 复杂组合查询函数，根据Match条件匹配记录
func (t *TableIter) GerMultiRecords(esc bool, page Page, match ...Match) (r record.Records) {
	//添加锁，防止并发访问
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.top[esc]() {
		return nil
	}
	var rd record.Record
	if page.Count > 0 {
		r = make(record.Records, 0, page.Count)
	}
	loop := 0
	// 处理当前位置的元素
	for {
		if loop < page.Start {
			loop++
		} else {
			for _, m := range match {
				fieldsBytes := t.ParseBytes(t.iter.Key(), t.iter.Value())
				idxrd := t.table.RecordByteToAny(fieldsBytes)
				if m.Match(idxrd) {
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

// 提取主键值
// if len(fields) == 0 ，默认是提取主键值，主键也可以是组合主键
// 单主键则返回原始值，组合主键则返回拼接的字符串
func (t *TableIter) GetPrimaryKeys(k, v []byte, fields ...string) (r any) {
	fbs := t.ParseBytes(k, v)
	if len(fields) == 0 {
		fields = t.table.indexs.GetPrimaryKey().GetFields()
	}
	//提取对应的值
	val := make(map[string][]byte, len(fields))
	for _, f := range fields {
		if v, ok := (*fbs)[f]; ok {
			val[f] = v
		}
	}
	isComposite := len(fields) > 1
	if isComposite {
		r = ""
		for _, f := range val {
			r = r.(string) + string(f) + SPLIT
		}
	} else {
		r = util.Bytes((*fbs)[fields[0]]).ToAny(t.table.fields[fields[0]])
	}
	return
}

// 将某字段的所有值转换为map[any]bool
// if len(fields) == 0 ，默认是提取主键值，否则提取指定字段值
// 用于与其他迭代器进行匹配。
func (t *TableIter) Map(fields ...string) (data map[any]bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	data = make(map[any]bool)
	for t.iter.Next() {
		data[t.GetPrimaryKeys(t.iter.Key(), t.iter.Value(), fields...)] = true
	}
	return
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
