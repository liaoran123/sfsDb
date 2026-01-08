package engine

import (
	"maps"
	"reflect"
)

type Record map[string]any

// 选择显示的key字段，如sql语句中的select f0,f1,... from table
// 如果keys为空，则返回所有字段
func (r Record) Select(keys ...string) (rs Record) {
	if r == nil {
		return nil
	}
	if len(keys) == 0 {
		return r
	}
	rs = make(map[string]any, len(keys))
	for _, key := range keys {
		rs[key] = r[key]
	}
	return rs
}

// ----------------------------------
type Records struct {
	table   *Table
	records []Record
}

func NewRecords(table *Table) *Records {
	return &Records{
		table:   table,
		records: make([]Record, 0),
	}
}

func (r *Records) Len() int {
	return len(r.records)
}

func (r *Records) Get(index int) Record {
	return r.records[index]
}
func (r *Records) GetAll() []Record {
	return r.records
}

// 检查记录是否存在
func (r *Records) Contains(rd Record) bool {
	for _, r := range r.records {
		if reflect.DeepEqual(r, rd) {
			return true
		}
	}
	return false
}

/*
不提供该函数，避免被注入风险

	func (r *Records) Set(index int, rd Record) {
		r.records[index] = rd
	}
*/
func (r *Records) Append(rd Record) error {
	r.records = append(r.records, rd)
	return nil
}

// 选择显示的key字段，如sql语句中的select f0,f1,... from table
// 如果keys为空，则返回所有字段
func (r *Records) Select(fields ...string) (rs []Record) {
	if len(fields) == 0 {
		return r.records
	}
	rs = make([]Record, len(r.records))
	for i, r := range r.records {
		rs[i] = r.Select(fields...)
	}
	return rs
}

// 判断主键是否存在
func (r *Records) hasPrimaryKey(rd Record) bool {
	pkfs := r.table.GetPrimaryKey().GetFields()
	for _, f := range pkfs {
		if _, ok := rd[f]; !ok {
			return false
		}
	}
	return true
}

// 删除记录
func (r *Records) Delete() error {
	var rdMap map[string]any
	var err error
	for _, rd := range r.records {
		//判断主键是否存在，因为通过select语句返回的记录是没有主键的
		if !r.hasPrimaryKey(rd) {
			continue
		}
		//删除记录
		rdMap = map[string]any(rd)
		err = r.table.Delete(&rdMap)
		if err != nil {
			return err
		}
	}
	return nil
}

// 更新记录
// fields *map[string]any传入的仅仅是修改的字段和值，并没有带主键，主键值从记录中提取
func (r *Records) Update(fields *map[string]any) error {
	if len(r.records) == 0 {
		return ErrNoUpdateFields
	}
	pkfs := r.table.GetPrimaryKey().GetFields()
	//判断主键是否存在，因为通过select语句返回的记录是没有主键的
	//同一个表，只需判断第一条记录是否有主键即可
	if !r.hasPrimaryKey(r.records[0]) {
		return ErrNoPrimaryKey
	}
	var pkValues map[string]any
	var err error
	for _, rd := range r.records {
		//从记录中提取主键值
		pkValues = make(map[string]any, len(pkfs))
		for _, f := range pkfs {
			pkValues[f] = rd[f]
		}
		//合并pkValues和fields
		maps.Copy(pkValues, *fields)
		//更新记录
		err = r.table.Update(&pkValues)
		if err != nil {
			return err
		}
	}
	return nil
}

// 交集
func (r *Records) Intersect(other ...*Records) *Records {
	rs := NewRecords(r.table)
	for _, rd := range r.records {
		found := true
		for _, o := range other {
			if !o.Contains(rd) {
				found = false
				break
			}
		}
		if found {
			rs.Append(rd)
		}
	}
	return rs
}

// 并集
func (r *Records) Union(other ...*Records) *Records {
	rs := NewRecords(r.table)
	for _, rd := range r.records {
		if !rs.Contains(rd) {
			rs.Append(rd)
		}
	}
	for _, o := range other {
		for _, rd := range o.records {
			if !rs.Contains(rd) {
				rs.Append(rd)
			}
		}
	}
	return rs
}

// 差集
func (r *Records) Difference(other ...*Records) *Records {
	rs := NewRecords(r.table)
	for _, rd := range r.records {
		found := false
		for _, o := range other {
			if o.Contains(rd) {
				found = true
				break
			}
		}
		if !found {
			rs.Append(rd)
		}
	}
	return rs
}
