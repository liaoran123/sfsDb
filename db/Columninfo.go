package db

// 列，字段结构信息，用于创建表
type ColumnInfo struct {
	Id          uint8    //不可修改
	Name        string   //字段名
	Types       string   //数据类型
	Primary     *Primary //主键，默认：nil，即非主键
	Index       []*Index //索引、全文索引、组合索引的Column.Id数组
	DefultValue string   //默认值
}
