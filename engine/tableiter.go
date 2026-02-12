package engine

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/liaoran123/sfsDb/match"
	"github.com/liaoran123/sfsDb/monitor"
	"github.com/liaoran123/sfsDb/record"
	"github.com/liaoran123/sfsDb/storage"
	"github.com/liaoran123/sfsDb/util"
)

type TableIter struct {
	iter       storage.Iterator
	jumpRanges []storage.Iterator //跳跃区间
	table      *Table
	match      []match.Match //匹配规则
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
	if iter == nil {
		return nil
	}
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

	// 获取一个干净的 Record 对象
	rd = record.GetRecord()

	switch t.index.(type) {
	case PrimaryKey: //主键通过Parse直接得到的就是记录
		anyMap := t.table.RecordByteToAny(fieldsBytes)
		if anyMap == nil {
			return nil
		} else {
			defer PutAnyMap(*anyMap)
		}
		// 填充数据到已获取的 Record 对象
		for k, v := range *anyMap {
			rd[k] = v
		}
	default: //其他二级索引通过Parse得到的是主键ID值，需要回表才能得到记录。
		// 拼接主键前缀和索引值，得到主键key
		pk := t.table.GetPrimaryKey()
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
		} else {
			defer GlobalFieldsBytesPool.Put(*trd)
		}
		//转换为记录 ： *map[string][]byte ==> *map[string]any
		anyMap := t.table.RecordByteToAny(trd)
		if anyMap == nil {
			return nil
		} else {
			defer PutAnyMap(*anyMap)
		}
		// 填充数据到已获取的 Record 对象
		for k, v := range *anyMap {
			rd[k] = v
		}
	}

	//版本号字段是乐观锁内部机制，不应该返回给用户。
	// 不返回会导致事务等操作失败。
	//delete(rd, "v")  //测试的时候需要屏蔽该句，否则相关测试会错误。

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
//jumpRange.First()=nil或jumpRange.Last()=nil的情况，需要特殊处理todo...(也可能不存在这样的情况)
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

// 遍历迭代器返回解析后的记录，不包含版本号字段。
// 用于外部调用，不包含版本号字段。
func (t *TableIter) GetRecordSet(esc bool, limit ...int) (r record.Records) {
	startTime := time.Now()
	Count := PageNew(limit...).Count
	if Count > 0 {
		r = record.GetRecordsWithCapacity(Count)
	} else {
		r = record.GetRecords()
	}
	r = r[:0]
	t.ExportRecord(func(rd *record.Record) bool {
		if len(*rd) == 0 { //删除记录后，数据为空，但是迭代器依然存在，只是返回空。
			return true
		}
		ird := rd.Select(t.selects...)
		//删除版本号字段
		delete(ird, "v")
		r = append(r, ird)
		return true
	}, esc, limit...)

	// 记录查询耗时
	t.RecordIndexTime(startTime, "GetRecordSet")
	return r
}

// 跟踪索引耗时
func (t *TableIter) RecordIndexTime(startTime time.Time, searchType string) {
	endTime := time.Now()
	//计算耗时
	duration := endTime.Sub(startTime)
	indexKey := monitor.GetIndexKey(t.table.id, t.index.GetId())
	monitor.GIndexStatsMap.SettimeAsync(indexKey, duration, t.table.name, t.index.Name(), searchType)
}

// 遍历迭代器返回解析后的记录，包含版本号字段。
// 用于系统内部调用，包含版本号字段。
func (t *TableIter) GetRecords(esc bool, limit ...int) (r record.Records) {
	startTime := time.Now()
	Count := PageNew(limit...).Count
	if Count > 0 {
		r = record.GetRecordsWithCapacity(Count)
	} else {
		r = record.GetRecords()
	}
	r = r[:0]
	t.ExportRecord(func(rd *record.Record) bool {
		if len(*rd) == 0 { //删除记录后，数据为空，但是迭代器依然存在，只是返回空。
			return true
		}
		ird := rd.Select(t.selects...)
		r = append(r, ird)
		return true
	}, esc, limit...)

	// 记录查询耗时
	t.RecordIndexTime(startTime, "GetRecords")
	return r
}

// 判断主键是否存在
func (t *TableIter) hasPrimaryKey(rd record.Record) bool {
	pkfs := t.table.GetPrimaryFields() // t.table.GetPrimaryKey().GetFields()
	for _, f := range pkfs {
		if _, ok := rd[f]; !ok {
			return false
		}
	}
	return true
}

// 删除迭代器中的记录
func (t *TableIter) Delete(limit ...int) error {
	var err error
	var batch storage.Batch
	deleteCount := 0
	maxLimit := -1
	if len(limit) > 0 {
		maxLimit = limit[0]
	}

	// 批量大小限制，当达到此大小时，执行一次写入
	const batchSizeLimit = 1000

	// 获取批量操作对象
	batch = t.table.kvStore.GetBatch()
	if batch == nil {
		return fmt.Errorf("创建批量操作对象失败")
	}

	// 批量删除处理
	t.ExportRecord(func(rd *record.Record) bool {
		// 判断是否达到删除限制
		if maxLimit > 0 && deleteCount >= maxLimit {
			return false
		}

		// 检查批量操作大小，如果达到限制，执行写入并创建新的批量操作
		if batch != nil && batch.Len() >= batchSizeLimit {
			// 执行当前批量操作
			if writeErr := t.table.kvStore.WriteBatch(batch); writeErr != nil {
				err = fmt.Errorf("执行批量操作失败: %v", writeErr)
				return false
			}

			// 创建新的批量操作对象
			batch = t.table.kvStore.GetBatch()
			if batch == nil {
				err = fmt.Errorf("创建新的批量操作对象失败")
				return false
			}
		}

		// 转换记录为map[string]any并调用Table.Delete
		rdMap := map[string]any(*rd)
		if deleteErr := t.table.Delete(&rdMap, batch); deleteErr != nil {
			// 记录不存在的错误可以忽略，继续删除其他记录
			if !strings.Contains(deleteErr.Error(), "记录不存在") {
				err = deleteErr
				return false
			}
			// 记录不存在，不增加删除计数
		} else {
			// 删除成功，增加删除计数
			deleteCount++
		}
		return true
	}, true, limit...)

	// 如果有删除操作，执行批量写入
	if deleteCount > 0 && batch != nil {
		if writeErr := t.table.kvStore.WriteBatch(batch); writeErr != nil && err == nil {
			err = fmt.Errorf("执行批量操作失败: %v", writeErr)
		}
	}

	return err
}

// 更新迭代器中的记录
func (t *TableIter) Update(fields *map[string]any, limit ...int) error {
	existpk := false
	var err error
	var batch storage.Batch
	updateCount := 0
	maxLimit := -1
	if len(limit) > 0 {
		maxLimit = limit[0]
	}
	// 批量大小限制，当达到此大小时，执行一次写入
	const batchSizeLimit = 1000
	// 获取批量操作对象
	batch = t.table.kvStore.GetBatch()
	defer func() {
		// 确保在发生错误时也能释放batch资源
		if batch != nil && updateCount > 0 {
			// 只有在有更新操作时才写入batch
			t.table.kvStore.WriteBatch(batch)
		}
	}()
	// 批量更新处理
	t.ExportRecord(func(rd *record.Record) bool {
		// 判断是否达到更新限制
		if maxLimit > 0 && updateCount >= maxLimit {
			return false
		}

		// 判断是否存在主键字段
		if !existpk { //只需要判断一次，因为所有的记录字段是一样。
			existpk = t.hasPrimaryKey(*rd)
			if !existpk {
				err = fmt.Errorf("记录不存在主键字段，无法更新")
				return false
			}
		}
		// 检查批量操作大小，如果达到限制，执行写入并创建新的批量操作
		if batch != nil && batch.Len() >= batchSizeLimit {
			// 执行当前批量操作
			if err = t.table.kvStore.WriteBatch(batch); err != nil {
				err = fmt.Errorf("执行批量操作失败: %v", err)
				return false
			}
			// 创建新的批量操作对象
			batch = t.table.kvStore.GetBatch()
			if batch == nil {
				err = fmt.Errorf("创建新的批量操作对象失败")
				return false
			}
		}
		// 从记录中提取主键值并合并更新字段
		rdMap := map[string]any(*rd)
		// 更新记录
		for field, value := range *fields {
			rdMap[field] = value
		}
		// 执行更新操作，使用批量操作
		err = t.table.Update(&rdMap, batch)
		if err != nil {
			err = fmt.Errorf("更新记录失败: %v", err)
			return false
		}
		updateCount++
		return true
	}, true, limit...)
	// 如果没有错误且有更新操作，执行批量写入
	if err == nil && updateCount > 0 && batch != nil {
		t.table.kvStore.WriteBatch(batch)
	}
	return err
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
	//var rdany *map[string]any
	//var fieldsBytes *map[string][]byte
	isMatch := true
	matchlen := 0
	page := PageNew(limit...)
	count := 0
	loop := 0
	var end []byte
	var key, value []byte
	for {
		key = t.iter.Key()
		value = t.iter.Value()
		if len(t.jumpRanges) > 0 { //如果有跳跃区间，需要跳过某些数据
			end = t.JumpRange(key, t.jumpRanges, esc)
			if end != nil {
				// 跳跃到区间的结束位置
				t.iter.Seek(end)
				if !t.move[esc]() {
					break
				}
			}
		}
		fieldsBytes := t.ParseBytes(key, value)
		defer func() {
			if fieldsBytes != nil && *fieldsBytes != nil {
				GlobalFieldsBytesPool.Put(*fieldsBytes)
			}
		}()
		isMatch = true
		matchlen = len(t.match)
		if matchlen > 0 { //是否需要匹配
			rdany := t.table.RecordByteToAny(fieldsBytes)
			defer func() {
				if rdany != nil {
					PutAnyMap(*rdany)
				}
			}()
			isMatch = t.Match(rdany, t.match)
		}
		if isMatch {
			if loop < page.Start {
				loop++
				if !t.move[esc]() {
					break
				}
				continue
			}
			rd = t.ParseRecord(fieldsBytes)
			if rd == nil { //删除记录后，数据为空，但是迭代器依然存在，只是返回nil。
				//loop++
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

// 提取主键值
// if len(fields) == 0 ，默认是提取主键值，主键也可以是组合主键
// 单主键则返回原始值，组合主键则返回拼接的字符串
func (t *TableIter) GetPrimaryKeys(k, v []byte, fields ...string) (r any) {
	fbs := t.ParseBytes(k, v)
	fany := t.table.RecordByteToAny(fbs)
	if len(fields) == 0 {
		fields = t.table.GetPrimaryFields() // t.table.GetPrimaryPrimaryKey().GetFields()
	}
	r = util.MergeFields(fields, fany)
	return
}

// 将某字段的所有值转换为map[any]bool
// if len(fields) == 0 ，默认是提取主键值，否则提取指定字段值
// 用于与其他迭代器进行匹配。
func (t *TableIter) Map(fields ...string) (data map[any]bool) {
	startTime := time.Now()
	data = GetMap()
	traverseCount := 0
	t.mu.Lock()
	for t.iter.First(); t.iter.Valid(); t.iter.Next() {
		data[t.GetPrimaryKeys(t.iter.Key(), t.iter.Value(), fields...)] = true
		traverseCount++
	}
	t.mu.Unlock()
	// 记录查询耗时
	t.RecordIndexTime(startTime, "Map")
	return
}
func (t *TableIter) ReleaseMap(data map[any]bool) {
	if data != nil {
		PutMap(data)
	}
}

// 获取搜索时使用的索引的名称
func (t *TableIter) GetIndexName() string {
	return t.index.Name()
}

// 获取搜索时使用的索引的id
func (t *TableIter) GetIndexId() uint8 {
	return t.index.GetId()
}

// 获取搜索时使用的索引的字段
func (t *TableIter) GetIndexFields() []string {
	return t.index.GetFields()
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

// Valid 检查迭代器是否有效
func (t *TableIter) Valid() bool {
	return t.iter.Valid()
}

// Seek 移动到大于等于指定key的位置
func (t *TableIter) Seek(key []byte) bool {
	startTime := time.Now()
	t.mu.Lock()
	exist := t.iter.Seek(key)
	t.mu.Unlock()
	t.RecordIndexTime(startTime, "Seek")
	return exist
}

// 判断是否存在指定的主键记录
func (t *TableIter) Exist() bool {
	startTime := time.Now()
	t.mu.Lock()
	exist := t.iter.First()
	t.mu.Unlock()
	t.RecordIndexTime(startTime, "Exist")
	return exist
}

// 统计索引记录数
func (t *TableIter) Count() int {
	startTime := time.Now()
	i := 0
	t.mu.Lock()
	// 重置迭代器到开头
	t.iter.First()
	// 遍历计数
	for t.iter.Valid() {
		i++
		t.iter.Next()
	}
	t.mu.Unlock()
	// 记录查询耗时
	t.RecordIndexTime(startTime, "Count")
	return i
}

func (t *TableIter) Release() {
	if t == nil {
		return
	}
	GlobalTableIterPool.Put(t)
}
