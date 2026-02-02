package engine

import "github.com/liaoran123/sfsDb/util"

// 获取自动增值的值
func (t *Table) GetAutoInc() int {
	if t.counter.Get() == 0 {
		t.InitAuto()
	}
	return int(t.counter.Increment())
}

// GetAutoIncBatch 批量获取自动增值的值
// count 需要获取的ID数量
// 返回值：第一个ID的值
func (t *Table) GetAutoIncBatch(count int) int {
	if count <= 0 {
		return 0
	}
	if t.counter.Get() == 0 {
		t.InitAuto()
	}
	// 先获取当前值，然后增加count
	current := t.counter.Get()
	t.counter.IncrementBy(count)
	return current + 1 // 返回第一个可用的ID
}

// 初始化自动增值的值
func (t *Table) InitAuto() {
	maxValue := t.MaxAutoValue()
	// MaxAutoValue现在直接返回int64类型
	t.counter.Set(int(maxValue))
}

// 获取当前最大自动增值记录的主键值
func (t *Table) MaxAutoValue() int {
	fields := map[string]any{"id": nil} //id为nil时，全表扫描。
	tableIter := t.Search(&fields)
	defer GlobalTableIterPool.Put(tableIter)
	if tableIter == nil {
		return 0
	}
	if !tableIter.Last() {
		return 0
	}
	key := tableIter.Key() //取最后一个key值
	rkey := key[len(t.GetPrimaryKey().Prefix(t.id))+1:]
	var target any
	// 如果主键字段为空，默认使用"id"
	primaryFields := t.GetPrimaryFields()
	if len(primaryFields) == 0 || primaryFields[0] == "" {
		target = 0
	} else {
		target = t.fields[primaryFields[0]]
	}
	r := util.Bytes(rkey).ToAny(target)
	// 使用 reflect 包进行类型转换，更灵活地处理各种数值类型
	return util.AnyToInt(r)

}
