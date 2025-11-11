package db

type Table struct {
	id   uint8
	name string             //表名
	Col  map[string]*Column //列词典
}

func TableNew(id uint8, name string) *Table {
	return &Table{id: id, name: name}
}

// 通过转换为暴露接口的TableInfo结构，给使用者查看表信息
func (t *Table) ToInfo() *TableInfo {
	return nil
}
