package engine

import (
	"github.com/liaoran123/sfsDb/util"
)

// 初始化自动增值计数器（如果需要）可能存在并发问题，需要使用锁保护
func (t *Table) initAutoCounterIfNeeded() {
	if t.counter.Get() == 0 {
		// 使用表级别的锁，避免全局锁导致的并发瓶颈
		t.initMutex.Lock()
		// 再次检查计数器是否为 0，避免重复初始化
		if t.counter.Get() == 0 {
			t.InitAuto()
		}
		t.initMutex.Unlock()
	}
}

// 获取自动增值的值
func (t *Table) GetAutoInc() int {
	t.initAutoCounterIfNeeded()
	return int(t.counter.Increment())
}

// GetAutoIncBatch 批量获取自动增值的值
// count 需要获取的ID数量
// 返回值：第一个ID的值
func (t *Table) GetAutoIncBatch(count int) int {
	if count <= 0 {
		return 0
	}
	t.initAutoCounterIfNeeded()
	// 使用原子操作获取当前值并增加指定的数量
	current := t.counter.GetAndIncrementBy(count)
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
	tableIter, err := t.Search(&fields)
	if err != nil {
		return 0
	}
	defer tableIter.Release()
	//defer GlobalTableIterPool.Put(tableIter)
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
