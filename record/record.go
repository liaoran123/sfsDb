package record

import (
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

type Records []Record

// 选择显示的key字段，如sql语句中的select f0,f1,... from table
// 如果keys为空，则返回所有字段
func (rs Records) Select(fields ...string) (rs2 Records) {
	if len(fields) == 0 {
		return rs
	}
	rs2 = make([]Record, len(rs))
	for i, r := range rs {
		rs2[i] = r.Select(fields...)
	}
	return rs2
}

// 添加交集，并集，差集等等
func (rs Records) Intersect(other ...Records) (rs2 Records) {
	if len(rs) == 0 || len(other) == 0 {
		return nil
	}
	rs2 = make([]Record, 0, len(rs))
	for _, r := range rs {
		found := false
		for _, o := range other {
			if o.Contains(r) {
				found = true
				break
			}
		}
		if found {
			rs2 = append(rs2, r)
		}
	}
	return rs2
}

// 判断是否包含指定记录
func (rs Records) Contains(r Record) bool {
	for _, record := range rs {
		if reflect.DeepEqual(record, r) {
			return true
		}
	}
	return false
}
func (rs Records) Union(other ...Records) (rs2 Records) {
	if len(rs) == 0 {
		if len(other) == 0 {
			return nil
		}
		// 当 rs 为空时，返回所有 other 记录的并集
		rs2 = make([]Record, 0)
		for _, o := range other {
			for _, r := range o {
				if !rs2.Contains(r) {
					rs2 = append(rs2, r)
				}
			}
		}
		return rs2
	}
	if len(other) == 0 {
		return rs
	}
	rs2 = make([]Record, 0, len(rs))
	// 添加 rs 中的所有记录
	for _, r := range rs {
		rs2 = append(rs2, r)
	}
	// 添加 other 中不存在于 rs 的记录
	for _, o := range other {
		for _, r := range o {
			if !rs2.Contains(r) {
				rs2 = append(rs2, r)
			}
		}
	}
	return rs2
}
func (rs Records) Difference(other ...Records) (rs2 Records) {
	if len(rs) == 0 {
		return nil
	}
	if len(other) == 0 {
		return rs
	}
	rs2 = make([]Record, 0, len(rs))
	for _, r := range rs {
		found := false
		for _, o := range other {
			if o.Contains(r) {
				found = true
				break
			}
		}
		if !found {
			rs2 = append(rs2, r)
		}
	}
	return rs2
}
