package engine

import (
	"bytes"
	"maps"
	"sync"

	"github.com/liaoran123/sfsDb/match"
	"github.com/liaoran123/sfsDb/record"
	"github.com/liaoran123/sfsDb/storage"
	"github.com/liaoran123/sfsDb/util"
)

type TableIter struct {
	iter       storage.Iterator
	jumpRanges []storage.Iterator
	table      *Table
	match      []match.Match
	selects    []string
	index      Index //搜索时使用的索引
	move       map[bool]func() bool
	top        map[bool]func() bool
	mu         sync.Mutex
}

// 导出记录 ，用于流式处理删除修改记录操作等。
type ExportRecord func(rd *record.Record) bool

// 导出函数
type Export func(k, v []byte) bool

func TableIterNew(table *Table, iter storage.Iterator, index Index, selects ...string) *TableIter {
	return &TableIter{
		table:   table,
		selects: selects,
		index:   index,
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

// 分页变量
type Page struct {
	Start int
	Count int
}

func PageNew(No ...int) Page {
	// 分页参数：start,count
	// start：分页开始位置，默认0
	// count：分页数量，默认-1表示返回所有记录
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
func (t *TableIter) SetJumpRanges(jumpRanges ...storage.Iterator) {
	t.jumpRanges = jumpRanges
}
func (t *TableIter) SetMatch(match ...match.Match) {
	t.match = match
}

// sql语句中的select f0,f1,... from table 要返回的字段
func (t *TableIter) SetSelects(fields ...string) {
	t.selects = fields
}

// 解析k，v里所有存在的字段byte值
// 主键得到整个记录值,其他索引得到主键ID值
func (t *TableIter) ParseBytes(k, v []byte) *map[string][]byte {
	var fieldsBytes *map[string][]byte
	var err error

	pk := t.table.GetPrimaryKey()
	pktylen := pk.GetfieldTypeLen(&t.table.fields)
	switch t.index.(type) {
	case FullTextIndex:
		//全文索引时，value值为空，需要从key中提取主键
		fieldsBytes, err = t.index.(FullTextIndex).Parse(pk.GetFields(), pktylen, k)
	case PrimaryKey:
		fieldsBytes, err = t.index.(PrimaryKey).Parse(t.table.fieldsid, v)
	default:
		fieldsBytes, err = t.index.(NormalIndex).Parse(pk.GetFields(), pktylen, v)

	}
	if err != nil {
		return nil
	}
	return fieldsBytes
}

func (t *TableIter) ParseRecord(fieldsBytes *map[string][]byte) (rd record.Record) {
	if fieldsBytes == nil {
		return nil
	}
	switch t.index.(type) {
	case PrimaryKey: //主键通过Parse直接得到的就是记录
		rd = record.Record(*t.table.RecordByteToAny(fieldsBytes))
	default: //其他二级索引通过Parse得到的是主键ID值，需要回表才能得到记录。
		// 拼接主键前缀和索引值，得到主键key
		pk := t.table.GetPrimaryKey()
		//pktylen := pk.GetfieldTypeLen(&t.table.fields)
		pfx := pk.JoinValue(fieldsBytes, t.table.id)
		// 回表读取完整记录
		byrecord := t.table.ReadByBytes(pfx)
		if byrecord == nil {
			return nil
		}
		//通过主键解析记录
		trd, err := pk.Parse(t.table.fieldsid, byrecord)
		if err != nil || trd == nil {
			return nil
		}
		//转换为记录 ： *map[string][]byte ==> *map[string]any
		rd = record.Record(*t.table.RecordByteToAny(trd))
	}
	return rd
}

/*
// 检测跳跃区间内是否包含key
// 如果包含，返回跳跃区间的结束位置
// 如果不包含，返回nil
//esc  true 表示顺序，false 表示倒序
场景1: 电商系统
// 商品格式：product_类别_12345
// 跳跃区间：SkipStart = []byte("product_electronics_"), SkipLimit = append([]byte("product_electronics_"), 0)
// 查询除电子产品外的所有商品
场景2: 金融系统
// 交易格式：transaction_20231201_12345
// 跳跃区间：SkipStart = []byte("transaction_20231201"), SkipLimit = []byte("transaction_20231202")
// 查询除2023年12月1日外的所有交易
*/
func (t *TableIter) JumpRange(key []byte, jumpRanges []storage.Iterator, esc bool) []byte {
	if len(jumpRanges) == 0 {
		return nil
	}
	for _, jumpRange := range jumpRanges {
		if esc {
			jumpRange.First()
		} else {
			jumpRange.Last()
		}
		if bytes.Equal(jumpRange.Key(), key) {
			if esc {
				// 顺序时，需要判断是否是最后一个元素
				if jumpRange.Last() {
					return jumpRange.Key()
				}
			} else {
				// 倒序时，需要判断是否是第一个元素
				if jumpRange.First() {
					return jumpRange.Key()
				}
			}
		}
	}
	return nil
}

/*
// 检测记录是否符合Match条件
// 如果符合，返回true
// 如果不符合，返回false
常见sql场景，f in (1,2,3) 或  and 等操作
*/
func (t *TableIter) Match(rd *map[string]any, match []match.Match) bool {
	if len(match) == 0 {
		return true //不需要匹配
	}
	for _, m := range match {
		if !m.Match(rd) {
			return false
		}
	}
	return true
}

// 遍历迭代器返回解析后的记录
func (t *TableIter) GetRecords(esc bool, limit ...int) (r record.Records) {
	t.ExportRecord(func(rd *record.Record) bool {
		if len(*rd) == 0 { //删除记录后，数据为空，但是迭代器依然存在，只是返回空。
			return true
		}
		ird := rd.Select(t.selects...)
		r = append(r, ird)
		return true
	}, esc, limit...)
	return r
}

// 判断主键是否存在
func (t *TableIter) hasPrimaryKey(rd record.Record) bool {
	pkfs := t.table.GetPrimaryKey().GetFields()
	for _, f := range pkfs {
		if _, ok := rd[f]; !ok {
			return false
		}
	}
	return true
}

// 删除迭代器中的记录
func (t *TableIter) Delete(limit ...int) {
	existpk := false
	var rdMap map[string]any
	var err error
	t.ExportRecord(func(rd *record.Record) bool {
		//判断是否存在主键字段
		if !existpk { //只需要判断一次，因为所有的记录字段是一样。
			existpk = t.hasPrimaryKey(*rd)
			if !existpk {
				return false
			}
		}
		//删除记录
		rdMap = map[string]any(*rd)
		err = t.table.Delete(&rdMap)
		if err != nil {
			return false
		}
		return true
	}, true, limit...)
}

// 更新迭代器中的记录
func (t *TableIter) Update(fields *map[string]any, limit ...int) {
	existpk := false
	pkfs := t.table.GetPrimaryKey().GetFields()
	var err error
	var pkValues map[string]any
	t.ExportRecord(func(rd *record.Record) bool {
		//判断是否存在主键字段
		if !existpk { //只需要判断一次，因为所有的记录字段是一样。
			existpk = t.hasPrimaryKey(*rd)
			if !existpk {
				return false
			}
		}
		//从记录中提取主键值
		pkValues = make(map[string]any, len(pkfs))
		for _, f := range pkfs {
			pkValues[f] = (*rd)[f]
		}
		//合并pkValues和fields
		maps.Copy(pkValues, *fields)
		//更新记录
		err = t.table.Update(&pkValues)
		if err != nil {
			return false
		}
		return true
	}, true, limit...)
}

// 遍历迭代器返回解析后的记录
// 复杂组合查询函数，根据Match条件匹配记录
// jumpRanges 跳跃区间，用于跳过某些数据
// export 导出函数，用于处理匹配的记录

func (t *TableIter) ExportRecord(export ExportRecord, esc bool, limit ...int) {
	//添加锁，防止并发访问
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.top[esc]() {
		return
	}
	var rd record.Record
	var rdany *map[string]any
	var fieldsBytes *map[string][]byte
	page := PageNew(limit...)
	count := 0
	loop := 0
	var end []byte
	var key, value []byte
	for {
		key = t.iter.Key()
		value = t.iter.Value()
		end = t.JumpRange(key, t.jumpRanges, esc)
		if end != nil {
			// 跳跃到区间的结束位置
			t.iter.Seek(end)
			if !t.move[esc]() {
				break
			}
		}
		fieldsBytes = t.ParseBytes(key, value)
		rdany = t.table.RecordByteToAny(fieldsBytes)
		if t.Match(rdany, t.match) {
			if loop < page.Start {
				loop++
				if !t.move[esc]() {
					break
				}
				continue
			}
			rd = t.ParseRecord(fieldsBytes)
			if rd == nil { //删除记录后，数据为空，但是迭代器依然存在，只是返回nil。
				loop++
				if !t.move[esc]() {
					break
				}
				continue
			}
			if !export(&rd) {
				return
			}
			count++
			loop++
			if page.Count > 0 && count >= page.Count {
				break
			}
		}
		// 移动到下一个/前一个元素
		if !t.move[esc]() {
			break
		}
	}

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

/*
func MergeFields(fns []string, fields *map[string]any) (r any) {
	if len(fns) > 1 {
		r = ""
		for _, f := range fns {
			//用分隔符SPLIT将fields中的值合并为一个字符串
			r = r.(string) + fmt.Sprintf("%v", (*fields)[f]) + SPLIT
		}
		// 去掉最后一个分隔符
		r = r.(string)[:len(r.(string))-1]
	} else {
		r = (*fields)[fns[0]]
	}
	return r
}
*/
// 提取主键值
// if len(fields) == 0 ，默认是提取主键值，主键也可以是组合主键
// 单主键则返回原始值，组合主键则返回拼接的字符串
func (t *TableIter) GetPrimaryKeys(k, v []byte, fields ...string) (r any) {
	fbs := t.ParseBytes(k, v)
	fany := t.table.RecordByteToAny(fbs)
	if len(fields) == 0 {
		fields = t.table.GetPrimaryKey().GetFields()
	}
	r = util.MergeFields(fields, fany)
	return
}

// 将某字段的所有值转换为map[any]bool
// if len(fields) == 0 ，默认是提取主键值，否则提取指定字段值
// 用于与其他迭代器进行匹配。
func (t *TableIter) Map(fields ...string) (data map[any]bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	data = make(map[any]bool)
	if t.iter.First() {
		data[t.GetPrimaryKeys(t.iter.Key(), t.iter.Value(), fields...)] = true
	}
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

// Seek 移动到大于等于指定key的位置
func (t *TableIter) Seek(key []byte) bool {
	return t.iter.Seek(key)
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
func (t *TableIter) Release() {
	t.iter.Release()
	for _, jumpRange := range t.jumpRanges {
		jumpRange.Release()
	}
}
