package db

//暴露一个表结构信息，用于创建和查看表
type TableInfo struct {
	Id   uint8
	Name string       //表名
	Col  []ColumnInfo //列词典
}

//转为Table
func (t *TableInfo) ToTable() (*Table, error) {
	NewTable := TableNew(t.Id, t.Name)
	var e error
	for _, v := range t.Col {
		col := Column{id: v.Id, primary: v.Primary, index: v.Index}
		//e = col.AddTypes(v.Types)
		if e != nil {
			return nil, e
		}
		e = col.Name(v.Name)
		if e != nil {
			return nil, e
		}
		//v.DefultValue根据Types转为对应的[]byte
		NewTable.Col[v.Name] = &col
	}
	return NewTable, e
}
