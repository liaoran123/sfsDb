package db

import "fmt"

//sfs数据库
type SfsDb struct {
	path  string            //数据库路径
	Table map[string]*Table //所有表词典
}

func SfsDbNew(path string) *SfsDb {
	return &SfsDb{path: path}
}

//加载所有表信息
func (s *SfsDb) LoadTables() error {
	//s.Table=
	return nil
}

//添加表
func (s *SfsDb) AddTable(TableInfo *TableInfo) error {
	tb, err := TableInfo.ToTable()
	if err != nil {
		return nil
	}
	fmt.Printf("tb: %v\n", tb)
	return nil
}

//删除表
func (s *SfsDb) DelTable(tableid uint8) error {
	return nil
}
