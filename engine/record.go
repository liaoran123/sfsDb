package engine

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
func (r *Records) Set(index int, rd Record) {
	r.records[index] = rd
}
func (r *Records) Append(rd Record) error {
	r.records = append(r.records, rd)
	return nil
}

// 选择显示的key字段，如sql语句中的select f0,f1,... from table
// 如果keys为空，则返回所有字段
func (r *Records) Select(fields ...string) (rs []Record) {
	if len(fields) == 0 {
		return rs
	}
	rs = make([]Record, len(r.records))
	for i, r := range r.records {
		rs[i] = r.Select(fields...)
	}
	return rs
}

// 删除记录
func (r *Records) Delete() error {
	return nil
}

// 更新记录
func (r *Records) Update(fields *map[string]any) error {
	return nil
}
