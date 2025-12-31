package engine

type Record map[string]any

// 选择显示的key字段，如sql语句中的select f0,f1,... from table
// 如果keys为空，则返回所有字段
func (r Record) GetKeys(keys ...string) (rs Record) {
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

type Records []Record

// 选择显示的key字段，如sql语句中的select f0,f1,... from table
// 如果keys为空，则返回所有字段
func (rs Records) GetKeys(keys ...string) (rs2 Records) {
	if len(keys) == 0 {
		return rs
	}
	rs2 = make([]Record, len(rs))
	for i, r := range rs {
		rs2[i] = r.GetKeys(keys...)
	}
	return rs2
}
