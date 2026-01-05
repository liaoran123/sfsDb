package sql

// Statement SQL语句接口
type Statement interface {
	// Type 返回SQL语句类型
	Type() string
	// Raw 返回原始SQL语句
	Raw() string
}

// BaseStatement 基础SQL语句结构体
type BaseStatement struct {
	sqlType string // SQL语句类型
	rawSQL  string // 原始SQL语句
}

// Type 返回SQL语句类型
func (s *BaseStatement) Type() string {
	return s.sqlType
}

// Raw 返回原始SQL语句
func (s *BaseStatement) Raw() string {
	return s.rawSQL
}

// SelectStatement SELECT语句
type SelectStatement struct {
	BaseStatement
	Fields     []string        // 查询字段
	Tables     []string        // 查询表
	Conditions []Condition     // 查询条件
	OrderBy    []OrderByClause // 排序条件
	GroupBy    []string        // 分组字段
	Limit      int             // 限制条数
	Offset     int             // 偏移量
}

// InsertStatement INSERT语句
type InsertStatement struct {
	BaseStatement
	Table  string   // 插入表
	Fields []string // 插入字段
	Values [][]any  // 插入值
}

// UpdateStatement UPDATE语句
type UpdateStatement struct {
	BaseStatement
	Table      string         // 更新表
	SetValues  map[string]any // 更新值
	Conditions []Condition    // 更新条件
}

// DeleteStatement DELETE语句
type DeleteStatement struct {
	BaseStatement
	Table      string      // 删除表
	Conditions []Condition // 删除条件
}

// CreateTableStatement CREATE TABLE语句
type CreateTableStatement struct {
	BaseStatement
	TableName string      // 表名
	Columns   []ColumnDef // 列定义
	Indexes   []IndexDef  // 索引定义
}

// CreateIndexStatement CREATE INDEX语句
type CreateIndexStatement struct {
	BaseStatement
	IndexName string   // 索引名
	TableName string   // 表名
	IndexType string   // 索引类型
	Fields    []string // 索引字段
	Unique    bool     // 是否唯一索引
}

// CreateDatabaseStatement CREATE DATABASE语句
type CreateDatabaseStatement struct {
	BaseStatement
	DatabaseName string // 数据库名
}

// UnknownStatement 未知类型SQL语句
type UnknownStatement struct {
	BaseStatement
}

// Condition 查询条件
type Condition struct {
	Field    string // 字段名
	Operator string // 操作符
	Value    any    // 值
}

// OrderByClause 排序条件
type OrderByClause struct {
	Field string // 排序字段
	Desc  bool   // 是否降序
}

// ColumnDef 列定义
type ColumnDef struct {
	Name       string // 列名
	Type       string // 列类型
	Length     int    // 长度
	NotNull    bool   // 是否非空
	PrimaryKey bool   // 是否主键
	Unique     bool   // 是否唯一
	Default    any    // 默认值
}

// IndexDef 索引定义
type IndexDef struct {
	Name    string   // 索引名
	Type    string   // 索引类型
	Fields  []string // 索引字段
	Unique  bool     // 是否唯一索引
	Primary bool     // 是否主键索引
}
